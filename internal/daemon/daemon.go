package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/yoosuf/Slash/internal/adapters"
	"github.com/yoosuf/Slash/internal/compress"
	"github.com/yoosuf/Slash/internal/store"
	"github.com/yoosuf/Slash/internal/track"
)

type Daemon struct {
	socket      string
	logLevel    string
	listener    net.Listener
	cache       *store.CCRCache
	readTracker *track.ReadStateTracker
	compressor  *compress.Compressor
	sessions    map[string]*SessionState
	sessionsMu  sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	logger      *log.Logger
	metrics     *SessionMetrics
}

type SessionState struct {
	SessionID string
	Workspace string
	ReadState *track.ReadState
	Metrics   *SessionMetrics
	LastSeen  time.Time
}

type SessionMetrics struct {
	mu                 sync.RWMutex
	TotalCalls         int64
	CallsCompressed    int64
	ReadsDiffed        int64
	LatencyMS          []int64
	MethodBreakdown    map[string]int64
}

type Config struct {
	Socket   string
	LogLevel string
	CacheDir string
	CacheTTL time.Duration
	CacheMax int64
}

func NewDaemon(cfg Config) (*Daemon, error) {
	ctx, cancel := context.WithCancel(context.Background())

	cache, err := store.NewCCRCache(cfg.CacheDir, cfg.CacheTTL, cfg.CacheMax)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize CCR cache: %w", err)
	}

	logger := log.New(os.Stderr, "[slash] ", log.LstdFlags)

	d := &Daemon{
		socket:      cfg.Socket,
		logLevel:    cfg.LogLevel,
		cache:       cache,
		readTracker: track.NewReadStateTracker(),
		compressor:  compress.NewCompressor(cache),
		sessions:    make(map[string]*SessionState),
		ctx:         ctx,
		cancel:      cancel,
		logger:      logger,
		metrics:     &SessionMetrics{MethodBreakdown: make(map[string]int64)},
	}

	return d, nil
}

func (d *Daemon) Start() error {
	os.Remove(d.socket)

	listener, err := net.Listen("unix", d.socket)
	if err != nil {
		return fmt.Errorf("failed to listen on socket %s: %w", d.socket, err)
	}

	d.listener = listener
	d.logger.Printf("daemon started on socket: %s", d.socket)

	go d.acceptLoop()
	go d.cleanupLoop()

	return nil
}

func (d *Daemon) acceptLoop() {
	for {
		select {
		case <-d.ctx.Done():
			return
		default:
		}

		conn, err := d.listener.Accept()
		if err != nil {
			continue
		}

		go d.handleConnection(conn)
	}
}

func (d *Daemon) handleConnection(conn net.Conn) {
	defer conn.Close()

	if err := conn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return
	}
	if err := conn.SetWriteDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return
	}

	decoder := json.NewDecoder(conn)
	var rawEvent map[string]interface{}

	if err := decoder.Decode(&rawEvent); err != nil {
		d.writeError(conn, "decode error")
		return
	}

	if msgType, ok := rawEvent["type"].(string); ok && msgType == "stats" {
		stats := d.Stats()
		statsBytes, _ := json.Marshal(stats)
		_, _ = conn.Write(statsBytes)
		return
	}

	eventBytes, err := json.Marshal(rawEvent)
	if err != nil {
		d.writeError(conn, "marshal error")
		return
	}

	hostType := ""
	if h, ok := rawEvent["host_type"].(string); ok {
		hostType = h
	} else {
		detected, err := adapters.DetectHostType(eventBytes)
		if err != nil {
			d.writeError(conn, "host detection failed")
			return
		}
		hostType = detected
	}

	adapter, err := adapters.GetAdapter(hostType)
	if err != nil {
		d.writeError(conn, fmt.Sprintf("unknown host: %s", hostType))
		return
	}

	event, err := adapter.DecodeHookEvent(eventBytes)
	if err != nil {
		d.writeError(conn, "decode failed")
		return
	}

	startTime := time.Now()
	session := d.getOrCreateSession(event.SessionID, event.Workspace)
	result := d.compress(event, session)
	latencyMS := int64(time.Since(startTime).Milliseconds())
	d.recordMetrics(session, event, result, latencyMS)

	resultBytes, err := adapter.EncodeHookResult(result)
	if err != nil {
		d.writeError(conn, "encode failed")
		return
	}

	if _, err := conn.Write(resultBytes); err != nil {
		return
	}
}

func (d *Daemon) compress(event *adapters.HookEvent, session *SessionState) *adapters.HookResult {
	result := &adapters.HookResult{
		PermissionDecision: adapters.PermissionAllow,
	}

	if event.EventKind != adapters.EventPostToolUse || event.ToolOutput == nil {
		return result
	}

	if event.Tool == "read" {
		path := ""
		if p, ok := event.ToolInput["path"].(string); ok {
			path = p
		}

		outputStr, ok := event.ToolOutput.(string)
		if !ok {
			return result
		}

		if path != "" && session.ReadState.WasRead(path) {
			previousContent, hasPrev := session.ReadState.GetPreviousContent(path)
			if hasPrev {
				compressed, meta := d.compressor.CompressReRead(outputStr, previousContent)
				result.UpdatedToolOutput = compressed
				result.CompressionMeta = meta

				session.ReadState.RecordRead(path, outputStr)
				return result
			}
		}

		compressed, meta := d.compressor.Compress(outputStr)
		if compressed != outputStr {
			result.UpdatedToolOutput = compressed
			result.CompressionMeta = meta
		}

		if path != "" {
			session.ReadState.RecordRead(path, outputStr)
		}
		return result
	}

	if event.Tool == "apply_patch" || event.Tool == "bash" {
		outputStr, ok := event.ToolOutput.(string)
		if !ok {
			return result
		}

		compressed, meta := d.compressor.Compress(outputStr)
		if compressed != outputStr {
			result.UpdatedToolOutput = compressed
			result.CompressionMeta = meta
		}

		if path, ok := event.ToolInput["path"].(string); ok && path != "" {
			session.ReadState.MarkEdited(path)
		}
		if path, ok := event.ToolInput["file"].(string); ok && path != "" {
			session.ReadState.MarkEdited(path)
		}
		return result
	}

	outputStr, ok := event.ToolOutput.(string)
	if ok {
		compressed, meta := d.compressor.Compress(outputStr)
		if compressed != outputStr {
			result.UpdatedToolOutput = compressed
			result.CompressionMeta = meta
		}
	}

	return result
}

func (d *Daemon) getOrCreateSession(sessionID, workspace string) *SessionState {
	d.sessionsMu.Lock()
	defer d.sessionsMu.Unlock()

	if session, ok := d.sessions[sessionID]; ok {
		session.LastSeen = time.Now()
		return session
	}

	session := &SessionState{
		SessionID: sessionID,
		Workspace: workspace,
		ReadState: track.NewReadState(),
		Metrics:   &SessionMetrics{MethodBreakdown: make(map[string]int64)},
		LastSeen:  time.Now(),
	}

	d.sessions[sessionID] = session
	return session
}

func (d *Daemon) recordMetrics(session *SessionState, event *adapters.HookEvent, result *adapters.HookResult, latencyMS int64) {
	if session.Metrics == nil {
		session.Metrics = &SessionMetrics{MethodBreakdown: make(map[string]int64)}
	}

	session.Metrics.mu.Lock()
	defer session.Metrics.mu.Unlock()

	d.metrics.mu.Lock()
	defer d.metrics.mu.Unlock()

	session.Metrics.TotalCalls++
	d.metrics.TotalCalls++

	session.Metrics.LatencyMS = append(session.Metrics.LatencyMS, latencyMS)
	d.metrics.LatencyMS = append(d.metrics.LatencyMS, latencyMS)

	if result.UpdatedToolOutput != nil || result.UpdatedInput != nil {
		session.Metrics.CallsCompressed++
		d.metrics.CallsCompressed++

		method := extractMethod(result.CompressionMeta)
		session.Metrics.MethodBreakdown[method]++
		d.metrics.MethodBreakdown[method]++
	}
}

func extractMethod(meta string) string {
	if meta == "" {
		return "none"
	}
	if contains(meta, "re-read") || contains(meta, "diff") {
		return "re_read_diff"
	}
	if contains(meta, "skeleton") {
		return "skeleton"
	}
	if contains(meta, "dedup") {
		return "dedup"
	}
	if contains(meta, "truncat") {
		return "truncate"
	}
	return "other"
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s[len(s)-len(substr):] == substr || s[:len(substr)] == substr || (len(s) > len(substr) && stringsContains(s, substr)))
}

func stringsContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func (d *Daemon) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-d.ctx.Done():
			return
		case <-ticker.C:
			d.cleanup()
		}
	}
}

func (d *Daemon) cleanup() {
	d.sessionsMu.Lock()
	defer d.sessionsMu.Unlock()

	now := time.Now()
	for sessionID, session := range d.sessions {
		if now.Sub(session.LastSeen) > 30*time.Minute {
			delete(d.sessions, sessionID)
		}
	}

	if err := d.cache.Cleanup(); err != nil {
		d.logger.Printf("ERROR: cache cleanup failed: %v", err)
	}
}

func (d *Daemon) Stats() map[string]interface{} {
	d.metrics.mu.RLock()
	defer d.metrics.mu.RUnlock()

	d.sessionsMu.RLock()
	activeSessions := len(d.sessions)
	d.sessionsMu.RUnlock()

	latencyP50, latencyP95 := percentiles(d.metrics.LatencyMS)

	return map[string]interface{}{
		"total_calls":        d.metrics.TotalCalls,
		"calls_compressed":   d.metrics.CallsCompressed,
		"active_sessions":    activeSessions,
		"latency_p50_ms":     latencyP50,
		"latency_p95_ms":     latencyP95,
		"method_breakdown":   d.metrics.MethodBreakdown,
		"cache_size_mb":      d.cache.SizeMB(),
		"cache_entries":      d.cache.EntryCount(),
	}
}

func percentiles(latencies []int64) (int64, int64) {
	if len(latencies) == 0 {
		return 0, 0
	}
	var p50, p95 int64
	if len(latencies) > 0 {
		p50 = latencies[len(latencies)/2]
	}
	if len(latencies) > 1 {
		p95 = latencies[len(latencies)*95/100]
	}
	return p50, p95
}

func (d *Daemon) Stop() error {
	d.cancel()
	if d.listener != nil {
		d.listener.Close()
	}
	if d.cache != nil {
		d.cache.Close()
	}
	os.Remove(d.socket)
	d.logger.Printf("daemon stopped")
	return nil
}

func (d *Daemon) writeError(conn net.Conn, errMsg string) {
	resp := map[string]interface{}{
		"permission_decision": "allow",
		"error":               errMsg,
	}
	bytes, _ := json.Marshal(resp)
	_, _ = conn.Write(bytes)
}
