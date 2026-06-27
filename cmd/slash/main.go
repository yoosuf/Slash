package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/yoosuf/Slash/internal/client"
	"github.com/yoosuf/Slash/internal/config"
	"github.com/yoosuf/Slash/internal/daemon"
	"github.com/yoosuf/Slash/internal/mcp"
	"github.com/yoosuf/Slash/internal/plugin"
	"github.com/yoosuf/Slash/internal/store"
)

const (
	Version = "1.0.0"
	AppName = "slash"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "daemon":
		cmdDaemon(args)
	case "hook":
		cmdHook(args)
	case "plugin":
		cmdPlugin(args)
	case "cache":
		cmdCache(args)
	case "audit":
		cmdAudit(args)
	case "purge":
		cmdPurge(args)
	case "stats":
		cmdStats(args)
	case "version":
		cmdVersion(args)
	case "mcp":
		cmdMCP(args)
	case "bench":
		cmdBench(args)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown command %q\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Slash v%s — token reduction + speed plugin for agentic coding tools

Usage: slash <command> [options]

Commands:
  daemon               Start the compression daemon (auto-started)
  hook                 Process a single hook event via the daemon (stdin/stdout)
  plugin               Manage plugin installation (install, ls, uninstall)
  cache                Inspect cached content (ls, check, stats)
  audit                Show compression breakdown by file/type
  purge                Wipe the cache
  stats                Show current session compression stats
  bench                Run compression benchmarks (JSON, code, logs, text)
  mcp                  Start the MCP server (for Zed and MCP clients)
  version              Print version
  help                 Show this message

Examples:
  slash daemon
  slash hook           # used by editor hooks
  slash plugin install claude-code
  slash cache ls
  slash audit
  slash stats
  slash bench --output results.json

Learn more: https://github.com/yoosuf/Slash
`, Version)
}

func cmdDaemon(args []string) {
	fs := flag.NewFlagSet("daemon", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: slash daemon [options]\n\nOptions:\n")
		fs.PrintDefaults()
	}

	cfgFile := config.Load()

	loglevel := fs.String("loglevel", cfgFile.Daemon.LogLevel, "log level (debug, info, warn, error)")
	socket := fs.String("socket", cfgFile.Daemon.Socket, "unix socket path")
	cacheDir := fs.String("cache-dir", cfgFile.Cache.Dir, "cache directory")
	cacheTTL := fs.Duration("cache-ttl", cfgFile.CacheTTL(), "cache TTL")
	cacheMax := fs.Int64("cache-max", cfgFile.Cache.MaxSize, "cache max size in MB")

	if err := fs.Parse(args); err != nil {
		return
	}

	if *socket == "" {
		home, _ := os.UserHomeDir()
		*socket = filepath.Join(home, ".slash", "daemon.sock")
	}
	if *cacheDir == "" {
		*cacheDir = defaultCacheDir()
	}

	cfg := daemon.Config{
		Socket:   *socket,
		LogLevel: *loglevel,
		CacheDir: *cacheDir,
		CacheTTL: *cacheTTL,
		CacheMax: *cacheMax,
	}

	d, err := daemon.NewDaemon(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create daemon: %v\n", err)
		os.Exit(1)
	}

	if err := d.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start daemon: %v\n", err)
		os.Exit(1)
	}

	select {}
}

func cmdHook(args []string) {
	fs := flag.NewFlagSet("hook", flag.ExitOnError)
	socket := fs.String("socket", "", "daemon socket path")
	_ = fs.Parse(args)

	raw, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read stdin: %v\n", err)
		os.Exit(1)
	}

	result, err := client.SendEvent(raw, *socket)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Hook failed: %v\n", err)
		os.Exit(1)
	}

	os.Stdout.Write(result)
}

func cmdPlugin(args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: slash plugin <ls|install|uninstall> [args]\n")
		os.Exit(1)
	}

	subcommand := args[0]
	subargs := args[1:]

	switch subcommand {
	case "ls":
		installed := plugin.Installed()
		if len(installed) == 0 {
			fmt.Println("No plugins installed.")
			return
		}
		fmt.Println("Installed plugins:")
		for _, p := range installed {
			paths := strings.Join(p.ConfigPaths, ", ")
			fmt.Printf("  %-20s %s\n", p.DisplayName, paths)
		}

	case "install":
		if len(subargs) == 0 {
			fmt.Fprintf(os.Stderr, "Usage: slash plugin install <host>\n")
			fmt.Fprintf(os.Stderr, "Available hosts: claude-code, codex, cursor, windsurf, antigravity, copilot, aider, zed, opencode, continue, cline, goose, pearai\n")
			os.Exit(1)
		}
		host := subargs[0]
		p := plugin.Get(host)
		if p == nil {
			fmt.Fprintf(os.Stderr, "Unknown host: %q\n", host)
			os.Exit(1)
		}
		fmt.Printf("Installing plugin for %s...\n", p.DisplayName)
		if err := p.Install(""); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to install plugin: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Plugin installed for %s.\n", p.DisplayName)
		fmt.Println("Restart the editor/tool to activate.")

	case "uninstall":
		if len(subargs) == 0 {
			fmt.Fprintf(os.Stderr, "Usage: slash plugin uninstall <host>\n")
			os.Exit(1)
		}
		host := subargs[0]
		p := plugin.Get(host)
		if p == nil {
			fmt.Fprintf(os.Stderr, "Unknown host: %q\n", host)
			os.Exit(1)
		}
		fmt.Printf("Uninstalling plugin for %s...\n", p.DisplayName)
		if err := p.Uninstall(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to uninstall plugin: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Plugin uninstalled for %s.\n", p.DisplayName)

	default:
		fmt.Fprintf(os.Stderr, "Unknown plugin subcommand: %q\n", subcommand)
		os.Exit(1)
	}
}

func cmdCache(args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: slash cache <ls|check|stats> [args]\n")
		os.Exit(1)
	}

	cacheDir := defaultCacheDir()
	cache, err := store.NewCCRCache(cacheDir, 24*time.Hour, 1024)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open cache: %v\n", err)
		os.Exit(1)
	}
	defer cache.Close()

	subcommand := args[0]
	subargs := args[1:]

	switch subcommand {
	case "ls":
		entries, err := cache.List()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to list cache: %v\n", err)
			os.Exit(1)
		}

		if len(entries) == 0 {
			fmt.Println("Cache is empty.")
		} else {
			fmt.Printf("%-15s %-12s %-40s %-20s %-10s %-10s\n", "Handle", "Tool", "Path", "Type", "Orig(KB)", "Comp(KB)")
			for _, e := range entries {
				fmt.Printf("%-15s %-12s %-40s %-20s %-10d %-10d\n",
					e.Handle, e.Tool, e.Path, e.CompressionType,
					e.SizeOriginal/1024, e.SizeCompressed/1024)
			}
		}

	case "check":
		if len(subargs) == 0 {
			fmt.Fprintf(os.Stderr, "Usage: slash cache check <path>\n")
			os.Exit(1)
		}
		path := subargs[0]
		entries, _ := cache.List()
		found := false
		for _, e := range entries {
			if e.Path == path {
				fmt.Printf("Found in cache: %s (%s, %d KB)\n", e.Handle, e.CompressionType, e.SizeOriginal/1024)
				found = true
			}
		}
		if !found {
			fmt.Printf("Not found in cache: %s\n", path)
		}

	case "stats":
		fmt.Println("Cache statistics:")
		fmt.Printf("  Size: %d MB\n", cache.SizeMB())
		fmt.Printf("  Entries: %d\n", cache.EntryCount())

	default:
		fmt.Fprintf(os.Stderr, "Unknown cache subcommand: %q\n", subcommand)
		os.Exit(1)
	}
}

func cmdAudit(args []string) {
	fs := flag.NewFlagSet("audit", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: slash audit [options]\n\nOptions:\n")
		fs.PrintDefaults()
	}

	sortBy := fs.String("sort", "savings", "sort by (savings, file, type)")
	jsonOut := fs.Bool("json", false, "output as JSON")

	if err := fs.Parse(args); err != nil {
		return
	}

	cacheDir := defaultCacheDir()
	cache, err := store.NewCCRCache(cacheDir, 24*time.Hour, 1024)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open cache: %v\n", err)
		os.Exit(1)
	}
	defer cache.Close()

	entries, err := cache.List()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to list cache: %v\n", err)
		os.Exit(1)
	}

	type auditEntry struct {
		Handle          string `json:"handle"`
		Path            string `json:"path"`
		Tool            string `json:"tool"`
		CompressionType string `json:"compression_type"`
		SizeOriginal    int64  `json:"size_original"`
		SizeCompressed  int64  `json:"size_compressed"`
		SavingsPercent  int    `json:"savings_percent"`
	}

	var audit []auditEntry
	var totalOrig, totalComp int64
	typeStats := make(map[string]int64)
	typeComp := make(map[string]int64)

	for _, e := range entries {
		savings := 0
		if e.SizeOriginal > 0 {
			savings = 100 * int(e.SizeOriginal-e.SizeCompressed) / int(e.SizeOriginal)
		}
		audit = append(audit, auditEntry{
			Handle:          e.Handle,
			Path:            e.Path,
			Tool:            e.Tool,
			CompressionType: e.CompressionType,
			SizeOriginal:    e.SizeOriginal,
			SizeCompressed:  e.SizeCompressed,
			SavingsPercent:  savings,
		})
		totalOrig += e.SizeOriginal
		totalComp += e.SizeCompressed
		typeStats[e.CompressionType] += e.SizeOriginal
		typeComp[e.CompressionType] += e.SizeCompressed
	}

	switch *sortBy {
	case "savings":
		sort.Slice(audit, func(i, j int) bool { return audit[i].SavingsPercent > audit[j].SavingsPercent })
	case "file":
		sort.Slice(audit, func(i, j int) bool { return audit[i].Path < audit[j].Path })
	case "type":
		sort.Slice(audit, func(i, j int) bool { return audit[i].CompressionType < audit[j].CompressionType })
	}

	if *jsonOut {
		out := map[string]interface{}{
			"entries":       audit,
			"total_original":   totalOrig,
			"total_compressed": totalComp,
			"overall_savings":  savings(totalOrig, totalComp),
			"by_type":          typeSavings(typeStats, typeComp),
		}
		data, _ := json.MarshalIndent(out, "", "  ")
		fmt.Println(string(data))
		return
	}

	totalSavings := savings(totalOrig, totalComp)
	fmt.Printf("Compression Audit — %d entries, %s → %s (%d%% savings)\n\n",
		len(entries), formatBytes(totalOrig), formatBytes(totalComp), totalSavings)

	fmt.Printf("%-15s %-40s %-12s %-10s %-10s %s\n", "Handle", "Path", "Type", "Original", "Compressed", "Savings")
	for _, a := range audit {
		fmt.Printf("%-15s %-40s %-12s %-10s %-10s %d%%\n",
			a.Handle, truncate(a.Path, 38), a.CompressionType,
			formatBytes(a.SizeOriginal), formatBytes(a.SizeCompressed), a.SavingsPercent)
	}

	fmt.Println("\n--- By Compression Type ---")
	for t := range typeStats {
		orig := typeStats[t]
		comp := typeComp[t]
		s := savings(orig, comp)
		fmt.Printf("  %-20s %s → %s (%d%%)\n", t, formatBytes(orig), formatBytes(comp), s)
	}
}

func cmdPurge(args []string) {
	fs := flag.NewFlagSet("purge", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: slash purge [options]\n\nOptions:\n")
		fs.PrintDefaults()
	}

	confirm := fs.Bool("confirm", false, "skip confirmation prompt")

	if err := fs.Parse(args); err != nil {
		return
	}

	cacheDir := defaultCacheDir()
	cache, err := store.NewCCRCache(cacheDir, 24*time.Hour, 1024)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open cache: %v\n", err)
		os.Exit(1)
	}
	defer cache.Close()

	if !*confirm {
		fmt.Print("This will delete the entire Slash cache. Continue? (y/N) ")
		var response string
		if _, err := fmt.Scanln(&response); err != nil {
			fmt.Println("Cancelled.")
			return
		}
		if response != "y" && response != "yes" {
			fmt.Println("Cancelled.")
			return
		}
	}

	if err := cache.Purge(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to purge cache: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Cache purged.")
}

func cmdStats(args []string) {
	socketPath := client.DefaultSocket()

	stats, err := client.RequestStats(socketPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not connect to daemon at %s\n", socketPath)
		fmt.Fprintf(os.Stderr, "Start the daemon first: slash daemon\n")
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)

		cacheDir := defaultCacheDir()
		cache, cerr := store.NewCCRCache(cacheDir, 24*time.Hour, 1024)
		if cerr == nil {
			defer cache.Close()
			fmt.Println("\nCache stats (daemon not running):")
			fmt.Printf("  Cache size: %d MB\n", cache.SizeMB())
			fmt.Printf("  Cached items: %d\n", cache.EntryCount())
		}
		os.Exit(1)
	}

	fmt.Println("Daemon session statistics:")
	if v, ok := stats["total_calls"].(float64); ok {
		fmt.Printf("  Total calls: %.0f\n", v)
	}
	if v, ok := stats["calls_compressed"].(float64); ok {
		fmt.Printf("  Calls compressed: %.0f\n", v)
	}
	if v, ok := stats["active_sessions"].(float64); ok {
		fmt.Printf("  Active sessions: %.0f\n", v)
	}
	if v, ok := stats["latency_p50_ms"].(float64); ok {
		fmt.Printf("  Latency p50: %.0f ms\n", v)
	}
	if v, ok := stats["latency_p95_ms"].(float64); ok {
		fmt.Printf("  Latency p95: %.0f ms\n", v)
	}
	if v, ok := stats["cache_size_mb"].(float64); ok {
		fmt.Printf("  Cache size: %.0f MB\n", v)
	}
	if v, ok := stats["cache_entries"].(float64); ok {
		fmt.Printf("  Cache entries: %.0f\n", v)
	}
	if methodBreakdown, ok := stats["method_breakdown"].(map[string]interface{}); ok {
		fmt.Println("  Method breakdown:")
		for method, count := range methodBreakdown {
			if c, ok := count.(float64); ok {
				fmt.Printf("    %-20s %.0f\n", method, c)
			}
		}
	}
}

func cmdVersion(args []string) {
	fmt.Printf("Slash v%s\n", Version)
}

func cmdMCP(args []string) {
	fs := flag.NewFlagSet("mcp", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: slash mcp [options]\n\nOptions:\n")
		fs.PrintDefaults()
	}

	port := fs.Int("port", 8765, "MCP server port")
	detach := fs.Bool("detach", false, "run in background")

	if err := fs.Parse(args); err != nil {
		return
	}

	if *detach {
		proc, err := os.StartProcess(os.Args[0], append([]string{os.Args[0], "mcp", "--port", fmt.Sprintf("%d", *port)}, fs.Args()...), &os.ProcAttr{
			Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to detach: %v\n", err)
			os.Exit(1)
		}
		if err := proc.Release(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to release process: %v\n", err)
		} else {
			fmt.Printf("MCP server started on port %d (PID %d)\n", *port, proc.Pid)
		}
		return
	}

	cacheDir := defaultCacheDir()
	cache, err := store.NewCCRCache(cacheDir, 24*time.Hour, 1024)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open cache: %v\n", err)
		os.Exit(1)
	}
	defer cache.Close()

	server := mcp.NewMCPServer(cache, *port)
	if err := server.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start MCP server: %v\n", err)
		os.Exit(1)
	}
}



// --- helpers ---

func defaultCacheDir() string {
	xdgCache := os.Getenv("XDG_CACHE_HOME")
	if xdgCache == "" {
		home, _ := os.UserHomeDir()
		xdgCache = filepath.Join(home, ".cache")
	}
	return filepath.Join(xdgCache, "slash")
}

func formatBytes(b int64) string {
	if b < 1024 {
		return fmt.Sprintf("%d B", b)
	}
	if b < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(b)/1024)
	}
	return fmt.Sprintf("%.1f MB", float64(b)/(1024*1024))
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func savings(orig, comp int64) int {
	if orig == 0 {
		return 0
	}
	return int(100 * (orig - comp) / orig)
}

func typeSavings(orig, comp map[string]int64) map[string]int {
	out := make(map[string]int)
	for t := range orig {
		out[t] = savings(orig[t], comp[t])
	}
	return out
}
