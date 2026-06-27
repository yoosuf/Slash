package mcp

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/yoosuf/Slash/internal/repomap"
	"github.com/yoosuf/Slash/internal/store"
)

type MCPServer struct {
	cache  *store.CCRCache
	logger *log.Logger
	port   int
}

func NewMCPServer(cache *store.CCRCache, port int) *MCPServer {
	return &MCPServer{
		cache:  cache,
		logger: log.New(os.Stderr, "[slash-mcp] ", log.LstdFlags),
		port:   port,
	}
}

func (s *MCPServer) Start() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/mcp/retrieve", s.handleRetrieve)
	mux.HandleFunc("/mcp/repomap", s.handleRepomap)
	mux.HandleFunc("/mcp/stats", s.handleStats)
	mux.HandleFunc("/health", s.handleHealth)

	addr := fmt.Sprintf(":%d", s.port)
	s.logger.Printf("MCP server starting on %s", addr)

	return http.ListenAndServe(addr, mux)
}

func (s *MCPServer) handleRetrieve(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Handle string `json:"handle"`
		Start  int64  `json:"start,omitempty"`
		End    int64  `json:"end,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, 400, "invalid request: "+err.Error())
		return
	}

	if req.Handle == "" {
		s.writeError(w, 400, "handle required")
		return
	}

	var content []byte
	var err error

	if req.Start > 0 || req.End > 0 {
		content, err = s.cache.GetRange(req.Handle, req.Start, req.End)
	} else {
		content, err = s.cache.Get(req.Handle)
	}

	if err != nil {
		s.writeError(w, 404, "not found: "+err.Error())
		return
	}

	s.writeJSON(w, 200, map[string]interface{}{
		"handle":  req.Handle,
		"content": string(content),
		"size":    len(content),
	})
}

func (s *MCPServer) handleRepomap(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Path string `json:"path,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, 400, "invalid request")
		return
	}

	root := req.Path
	if root == "" {
		cwd, err := os.Getwd()
		if err != nil {
			s.writeError(w, 500, "failed to get working directory")
			return
		}
		root = cwd
	}

	builder := repomap.NewBuilder()
	rm, err := builder.Build(root)
	if err != nil {
		s.writeError(w, 500, "failed to build repo map: "+err.Error())
		return
	}

	jsonData, err := rm.ToJSON()
	if err != nil {
		s.writeError(w, 500, "failed to serialize repo map")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	_, _ = w.Write([]byte(jsonData))
}

func (s *MCPServer) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}

	stats := map[string]interface{}{
		"cache_size_mb":  s.cache.SizeMB(),
		"cache_entries":  s.cache.EntryCount(),
		"mcp_version":    "1.0.0",
	}

	s.writeJSON(w, 200, stats)
}

func (s *MCPServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.writeJSON(w, 200, map[string]string{
		"status":  "healthy",
		"service": "slash-mcp",
	})
}

func (s *MCPServer) writeJSON(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		s.logger.Printf("error encoding JSON response: %v", err)
	}
}

func (s *MCPServer) writeError(w http.ResponseWriter, code int, message string) {
	s.writeJSON(w, code, map[string]string{
		"error": message,
	})
}

type MCPToolDefinition struct {
	Name        string
	Description string
	InputSchema map[string]interface{}
}

func GetToolDefinitions() []MCPToolDefinition {
	return []MCPToolDefinition{
		{
			Name:        "retrieve",
			Description: "Retrieve the full original content by handle (for compressed items)",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"handle": map[string]interface{}{
						"type":        "string",
						"description": "Content handle (e.g., 'h_abc123')",
					},
					"start": map[string]interface{}{
						"type":        "integer",
						"description": "Optional start byte offset",
					},
					"end": map[string]interface{}{
						"type":        "integer",
						"description": "Optional end byte offset",
					},
				},
				"required": []string{"handle"},
			},
		},
		{
			Name:        "repomap",
			Description: "Get a symbol index of the repository (functions, classes, structs, interfaces, types, imports)",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Optional file or directory path to scope the index (defaults to current working directory)",
					},
				},
			},
		},
		{
			Name:        "stats",
			Description: "Get current session compression statistics",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}
}

type MCPManifest struct {
	Name        string
	Version     string
	Description string
	Tools       []MCPToolDefinition
}

func GetManifest() MCPManifest {
	return MCPManifest{
		Name:        "slash",
		Version:     "1.0.0",
		Description: "Token reduction and compression for agentic coding tools",
		Tools:       GetToolDefinitions(),
	}
}
