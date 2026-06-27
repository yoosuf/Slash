# Architecture Deep Dive

## Design Pillars

1. **Fail open, always.** If the daemon is down, slow, or crashes, the hook returns input unchanged. A broken compressor never breaks a session.
2. **Warm daemon, thin hooks.** Compression runs on the hot path; the hook process must be a tiny client talking to an already-running daemon over a unix socket, returning in <50ms.
3. **Lossy is fine if reversible.** Every compression stashes the original and hands the model a `retrieve(handle)` tool. Aggressive compression is only safe because detail is fetchable on demand.
4. **Go native, not blind-proxy.** The reason to build per-host instead of just a wire proxy is that the host tells you *what* a tool call is (a Read, a re-read, a Bash dump). That structure enables diff-only re-reads and AST skeletonization a proxy can't do cleanly.
5. **Measure both axes.** Token reduction is meaningless without task pass-rate. Build the eval harness early.
6. **Local by default, no surprises.** Zero network calls in the core. The cache holds the user's own source under their own file permissions; nothing is sent anywhere without an explicit, documented opt-in.

## Data Flow

```
Host Tool (Claude Code, Codex, etc.)
  ↓
  Hook fires (PreToolUse or PostToolUse event)
  ↓
  Hook event → stdout (JSON) or file
  ↓
  Hook adapter (host-specific client)
  ├─ Parse JSON → HookEvent
  ├─ Call daemon over socket
  │ └─→ [Daemon process]
  │     ├─ Router (content-type detection)
  │     ├─ Read-state tracker (for diff-only)
  │     ├─ Compressor (JSON/code/logs/text)
  │     ├─ CCR cache (write handle → original)
  │     └─ Repo index (symbol map)
  │
  ├─ Receive HookResult
  ├─ Parse → HostSpecific fields
  └─ Return host's native JSON
  ↓
  Host tool receives response
  ├─ updatedInput (if pre-call, modified)
  ├─ updatedToolOutput (if post-call, compressed)
  └─ Tool executes / response enters context
```

## Key Components

### 1. Hook Adapter (`internal/adapters/`)

**Responsibility:** Translate a host's native hook JSON to/from the canonical schema.

**Files:**
- `schema.go` — HookEvent, HookResult, HostAdapter interface (the contract).
- `claudecode.go`, `codex.go`, `antigravity.go`, `copilot.go` — per-host implementations.
- `registry.go` — factory to get the right adapter.
- `fixtures/` — real hook JSON examples for testing.

**Example (Claude Code):**
```
Input (Claude Code pre-call hook):
  {
    "eventId": "evt_123",
    "eventKind": "preToolUse",
    "tool": "read",
    "toolInput": {"path": "/file.txt"},
    "workspaceDir": "/workspace"
  }
  
Adapter.DecodeHookEvent():
  → HookEvent{
      HostType: "claudecode",
      EventID: "evt_123",
      EventKind: "preToolUse",
      Tool: "read",
      ToolInput: {"path": "/file.txt"},
      Workspace: "/workspace",
      ...
    }

[Compressor processes this]

Adapter.EncodeHookResult(result):
  ← HookResult{
      PermissionDecision: "allow",
      UpdatedInput: {"path": "/file.txt", "range": [0, 100]}, // optimized!
    }
  
Output (Claude Code response):
  {
    "permissionDecision": "allow",
    "hookSpecificOutput": {
      "updatedInput": {"path": "/file.txt", "range": [0, 100]}
    }
  }
```

### 2. Daemon (`internal/daemon/` — PHASE 1)

**Responsibility:** Long-running process that compresses tool calls. Listens on a unix socket, routes to compressors, manages lifecycle.

**Pseudocode:**
```go
func (d *Daemon) Run() {
  listener := net.Listen("unix", d.Socket)
  for {
    conn := listener.Accept()
    go d.handleConnection(conn)
  }
}

func (d *Daemon) handleConnection(conn net.Conn) {
  event := parseJSON(conn)  // HookEvent
  result := d.compress(event)  // → HookResult
  writeJSON(conn, result)
}
```

**Manages:**
- Read-state tracking (per-session file content hashes).
- CCR cache lifecycle (TTL, eviction, invalidation).
- Repo index (symbol map, file-watch refresh).
- Session telemetry (token counts, latency).

### 3. Router (`internal/router/` — PHASE 2)

**Responsibility:** Detect content type and route to the right compressor.

**Types:**
- **JSON** — tree structure, skeletonize non-leaf nodes.
- **Code** — AST-based (tree-sitter), skeleton + optional full via retrieve.
- **Logs/stderr** — deduplicate repeated lines, truncate + handle.
- **Text** — simple heuristics (truncate + retrieve).
- **Diff/patch** — structured; compress applied regions only.

**Example:**
```go
func (r *Router) Detect(output interface{}) ContentType {
  if s, ok := output.(string); ok {
    if isJSON(s) { return TypeJSON }
    if isCode(s) { return TypeCode }  // via MIME or file ext
    if isLogs(s) { return TypeLogs }
    return TypeText
  }
  return TypeBinary  // binary output, pass through
}
```

### 4. Compressors (`internal/compress/`)

#### JSON Compressor
**Input:** JSON tree (parsed or string).
**Strategy:** Skeleton (keep structure, drop leaf values), optionally include counts/types.
**Output:** Compressed tree + handle to full body.
**Example:**
```
Input:
  {
    "user": {
      "id": 12345,
      "email": "user@example.com",
      "preferences": {
        "theme": "dark",
        "notifications": {"email": true, "sms": false}
      }
    }
  }

Output:
  {
    "user": {
      "id": "<number>",
      "email": "<string:20 chars>",
      "preferences": {
        "theme": "<string>",
        "notifications": {"email": "<boolean>", "sms": "<boolean>"}
      },
      "__slash_note": "[retrieve(h_abc123) for full]"
    }
  }
```

#### Code Compressor (tree-sitter)
**Input:** Source code (Go, Python, TS, etc.).
**Strategy:** Parse to AST, emit skeleton (function sigs, class defs, imports, comments) + optional full body.
**Output:** Skeleton + handle.
**Example:**
```go
// Input
func (r *Router) Compress(output interface{}) {
  if s, ok := output.(string); ok {
    detectType(s)
    // 1000 lines...
  }
}

// Output (skeleton)
func (r *Router) Compress(output interface{}) {
  // [skipped: 1000 lines of implementation]
  // [retrieve(h_def456) for full function body]
}
```

#### Log Compressor
**Input:** Bash stdout/stderr (often multi-line).
**Strategy:** Deduplicate repeated lines, truncate long runs, keep first + last + unique lines.
**Output:** Deduplicated log + handle.
**Example:**
```
Input (100 lines of repeated "Building..." ):
  Building package A...
  Building package A...
  ... [repeated 97 times]
  Build failed.

Output:
  Building package A... [repeated 100 times]
  Build failed.
  [retrieve(h_ghi789) for full log]
```

#### Text Compressor
**Input:** Plain text (no structure).
**Strategy:** Simple heuristics — truncate at token limit, keep first + last, add handle.
**Output:** Truncated text + handle.

### 5. CCR Cache (`internal/store/` — PHASE 2)

**Responsibility:** Store compressed content's original, retrieve on demand.

**Schema (SQLite):**
```sql
CREATE TABLE ccr_cache (
  handle TEXT PRIMARY KEY,  -- h_<sha8 of content>
  original BLOB,
  tool TEXT,                -- "read", "bash", "apply_patch", etc.
  path TEXT,                -- file path (or NULL)
  compression_type TEXT,    -- "json_skeleton", "code_ast", "log_dedup"
  created_at TIMESTAMP,
  accessed_at TIMESTAMP,
  size_original INT,
  size_compressed INT
);

CREATE INDEX idx_path_tool ON ccr_cache(path, tool);
```

**Lifecycle:**
- On compression: insert `(handle, original, metadata)`.
- On tool call `retrieve(handle, [range])`: fetch original, return (optionally sliced).
- On file edit: invalidate handles tied to that path.
- On TTL expiry or LRU eviction: delete oldest/largest rows.

**Example:**
```
compress(read("/foo.txt")) → result: "[skeleton]\n[retrieve(h_abc123)]"
→ cache.Insert(h_abc123, "full content of foo.txt", {tool: "read", path: "/foo.txt"})

model calls retrieve(h_abc123) → cache.Get(h_abc123) → returns "full content"
```

### 6. Read-State Tracker (`internal/track/` — PHASE 1)

**Responsibility:** Track which files were read and edited, enable diff-only re-reads.

**State per session:**
```go
type ReadState struct {
  FileHashes map[string]string  // path → content hash
  FileEdits  map[string]time.Time  // path → last edit time
}
```

**Logic:**
- On `read(path)` pre-call: record hash.
- On `apply_patch(path)` or `bash("... > path")`: mark as edited.
- On next `read(path)` pre-call: if hash changed, rewrite to `read(path, range=[first_diff, last_diff])`.
- After tool returns: update hash for next iteration.

**Benefit:** 80–95% savings on re-reads (model only gets the delta, not the whole file).

### 7. Repo Index (`internal/repomap/` — PHASE 3)

**Responsibility:** Build a symbol index at session start, inject into context so agent orients without exploratory reads.

**Computed from:**
- Tree-sitter parsing (symbols, dependencies).
- File-watch events (add/delete/rename).
- Gitignore patterns (exclude build artifacts).

**Output:**
- Compact JSON tree of `{module → exports, types, functions}`.
- Injected as context at session start (not a tool output).

**Benefit:** ~10–15% faster (model knows the codebase shape without reading 20 files).

### 8. MCP Server (`internal/mcp/` — PHASE 2)

**Responsibility:** Expose the compressor's tools to the model (and to Zed).

**Tools:**
```
retrieve(handle: string, range?: [start, end]) → original content (or slice)
repomap(path?: string) → symbol index JSON
stats() → {"tokens_saved": 123456, "calls": 45, "latency_p95": 35}
```

**Used by:**
- Claude Code / Codex / Antigravity / Copilot: when model needs detail.
- Zed (MCP client): for read/repomap/stats without full compression (MCP is tools-only).

### 9. Wire Proxy (`internal/proxy/` — PHASE 7)

**Responsibility:** Fallback for tools that don't have native hook support (or use base-URL overrides).

**Approach:**
- Proxy HTTPS traffic.
- Intercept `POST https://api.anthropic.com/v1/messages`.
- Compress request body (system prompt, history, tools) and response (tokens, usage).
- Return compressed response to client.

**Limitation:** Can't use host metadata (doesn't know if a tool call is a re-read), so less aggressive than native hooks.

## Session Lifecycle

```
1. User starts editor (Claude Code, Codex, etc.)
2. Editor loads plugins (hooks + MCP servers)
3. First hook call fires
   → Hook client tries to connect to daemon
   → If daemon not running, start it (or fail open)
4. Daemon starts (warm)
   → Initializes CCR cache
   → Starts repo indexing (async)
   → Listens on socket
5. Hook client sends HookEvent
6. Daemon:
   → Router detects type (JSON/code/logs/text)
   → Compressor runs (using read-state for diffs)
   → CCR cache stores original + handle
   → Repo index updated if file changed
7. Daemon returns HookResult
8. Hook client parses, converts back to host JSON
9. Host tool receives response (compressed output + retrieve tool available)
10. Model processes context, may call retrieve(handle)
11. Hook client → daemon → cache.Get(handle) → returns original
12. Loop until task complete
13. Session ends
    → Daemon stays warm (or exits after idle timeout)
    → CCR cache persists (TTL: 24h)
```

## Failure Modes & Recovery

| Failure | Behavior | Recovery |
|---|---|---|
| **Daemon crashes** | Hook client fails open (returns input unchanged). | Daemon auto-restarts on next hook call. |
| **Socket timeout** | Hook client returns input unchanged if >50ms. | User may experience slower model response (no compression). |
| **Cache full** | LRU eviction starts; oldest entries deleted. | Compression continues; just fewer retrievable items. |
| **File edited mid-session** | Handle invalidated; retrieve fails gracefully. | Model gets a "content expired" error, can re-read fresh. |
| **Compressed file > original** | Don't compress; return input as-is. | Safe, but signals a tuning opportunity. |

## Performance Targets

| Metric | Target | Rationale |
|---|---|---|
| Hook latency (p50) | <5ms | On the critical path; must be imperceptible. |
| Hook latency (p95) | <40ms | Aggressive; anything >50ms and we fail open. |
| Compression ratio | 40–60% blended | 2–3x on some surfaces, near-identity on others. |
| Pass-rate delta | <2% | Quality loss must be negligible. |
| Cache TTL | 24h | Balances privacy (purged daily) and re-read frequency. |
| Daemon startup | <1s | Auto-starts on first hook; must not stall editor. |

## Extensibility

The schema (`HookEvent`, `HookResult`, `HostAdapter`) is the public contract. To add a new host:

1. Create `internal/adapters/newtool.go` implementing `HostAdapter`.
2. Add fixtures in `internal/adapters/fixtures/newtool_*.json`.
3. Add tests to `schema_test.go`.
4. Update `registry.go` to return the adapter.
5. Submit a PR.

No changes to daemon, router, compressors, or cache. The core stays pure.

---

**Design rationale:** The schema is minimal and stable; adapters are thin and disposable. This allows anyone to add a new host without understanding the entire compressor. The daemon, compressors, and cache are shared, battle-tested, and improved once for all hosts.
