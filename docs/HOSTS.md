# Supported Hosts & Editors

Slash works with **13 major agentic coding tools**, covering the full spectrum of IDEs and CLI tools.

---

## Quick Reference

| Editor | Type | Status | Install | Notes |
|---|---|---|---|---|---|---|
| **Claude Code** | IDE | ✅ V1 | `slash plugin install claude-code` | Full native support |
| **Codex** | IDE | ✅ V1 | `slash plugin install codex` | Pre-call rewriting |
| **Cursor** | IDE | ✅ V1 | `slash plugin install cursor` | VSCode-based, full support |
| **Windsurf** | IDE | ✅ V1 | `slash plugin install windsurf` | VSCode-based, full support |
| **Antigravity** | CLI | ✅ V1 | `slash plugin install antigravity` | hooks.json format |
| **Copilot CLI** | CLI | ✅ V1 | `slash plugin install copilot` | Rich event types |
| **Aider** | CLI/Python | ✅ V1 | `slash plugin install aider` | Python hooks |
| **Zed** | Editor | ✅ V1 | `slash plugin install zed` | MCP server integration |
| **Opencode** | CLI | ✅ V1 | `slash plugin install opencode` | Custom plugin |
| **Continue** | IDE | ✅ V1 | `slash plugin install continue` | MCP server integration |
| **Cline** | IDE | ✅ V1 | `slash plugin install cline` | MCP server integration |
| **Goose** | CLI | ✅ V1 | `slash plugin install goose` | MCP server integration |
| **PearAI** | IDE | ✅ V1 | `slash plugin install pearai` | MCP server integration |

**Coverage:** All 13 editors use the same core compression engine via the adapter pattern.

---

## Detailed Guides

### IDE-Based Editors

#### Claude Code
**What it is:** Anthropic's native IDE for agentic coding  
**Platform:** Standalone application  
**Installation:**
```bash
slash plugin install claude-code
# Restart Claude Code; trust prompt will appear
```
**Features:**
- ✅ Full pre-call + post-call hook support
- ✅ Direct MCP integration
- ✅ Reversible compression via `retrieve()` tool
- ✅ Zero configuration (auto-starts daemon)

**Configuration:** `.slash/config.json` in home directory

---

#### Cursor
**What it is:** VSCode-based IDE with integrated Claude agent  
**Platform:** VSCode extension  
**Installation:**
```bash
slash plugin install cursor
# Restart Cursor; check Extensions tab for Slash
```
**Features:**
- ✅ Full hook support (same as Claude Code)
- ✅ MCP tools available in chat context
- ✅ Works with Cursor's native agent
- ✅ Same compression algorithms as Claude Code

**Note:** Cursor uses the same hook format as Claude Code (camelCase JSON), so Slash integration is seamless.

---

#### Windsurf
**What it is:** VSCode-based IDE with Cascade agent (similar to Cursor)  
**Platform:** VSCode extension  
**Installation:**
```bash
slash plugin install windsurf
# Restart Windsurf
```
**Features:**
- ✅ Full pre/post-call hooks
- ✅ Automatic daemon management
- ✅ MCP retrieval tool in Cascade context
- ✅ Compatible with Windsurf's multi-file editing

**Note:** Windsurf's hook format is nearly identical to Claude Code; minimal adapter code needed.

---

### CLI Tools

#### Codex
**What it is:** OpenAI's Codex engine with Claude integration  
**Platform:** CLI / API  
**Installation:**
```bash
slash plugin install codex
# Set CLAUDE_PLUGIN_ROOT if installing manually
```
**Features:**
- ✅ Pre-call input rewriting (primary optimization)
- ⚠️ Post-call output compression (best-effort)
- ✅ MCP retrieve() tool
- ✅ Works with local + API deployments

**Note:** Codex has partial output interception support; Slash maximizes what's available.

---

#### Antigravity / agy
**What it is:** Google's CLI agent for codebase navigation  
**Platform:** CLI (Go-based)  
**Installation:**
```bash
# Install Slash binary first
slash plugin install antigravity

# Then install via agy's plugin system
agy plugins install slash
```
**Features:**
- ✅ Full hook support (pre-flight + post-flight)
- ✅ `hooks.json` format (custom to agy)
- ✅ MCP integration
- ✅ File-watch refresh for cache invalidation

**Note:** Antigravity is Go-based like Slash; excellent compatibility.

---

#### Copilot CLI
**What it is:** GitHub's CLI-based AI pair programmer  
**Platform:** CLI (Node.js)  
**Installation:**
```bash
slash plugin install copilot
# Restart your shell
```
**Features:**
- ✅ Full hook support (preToolUse, postToolUse)
- ✅ Native compaction overlay (Slash adds ~10–15% savings on top)
- ✅ Rich event types (sessionStart, sessionEnd, etc.)
- ✅ MCP tools available

**Note:** Copilot CLI already does ~95% compression natively; Slash is *incremental* on top.

---

#### Aider
**What it is:** Python-based CLI agent for collaborative coding  
**Platform:** CLI (Python 3.7+)  
**Installation:**
```bash
slash plugin install aider

# Aider will auto-detect Slash on first run
# Or set environment variable:
export SLASH_SOCKET=$HOME/.slash/daemon.sock
```
**Features:**
- ✅ Hook support via Python (`.hooks/hook.py`)
- ✅ Works with Aider's subprocess model
- ✅ MCP tools available
- ✅ Environment variable configuration

**Note:** Aider runs as a subprocess; Slash integrates via named socket + Python hook script.

---

### Browser-Based (Zed)

#### Zed
**What it is:** High-performance code editor with VSCode compatibility  
**Platform:** VSCode extension (or native)  
**Integration:** MCP server (indirect)  
**Installation:**
```bash
# Add to .zed/settings.json
{
  "mcp": {
    "slash": {
      "command": "slash",
      "args": ["mcp", "--port", "8765"]
    }
  }
}
```
**Features:**
- ✅ MCP server integration (retrieve, repomap, stats)
- ✅ Works with any agent run inside Zed (Claude Code, Codex, Copilot CLI)
- ⚠️ Read/output compression only via underlying agent (Zed itself can't intercept hooks)
- ✅ Full reversible retrieval

**Note:** Zed doesn't expose its own agent hooks; instead, run Claude Code/Codex/etc. inside Zed. Slash then compresses via those agents' native hook systems. Alternatively, use the MCP server for retrieve() + tools-only mode.

---

### MCP-Based (Opencode)

**What it is:** opencode CLI with plugin system  
**Platform:** CLI / TUI  
**Installation:**
```bash
slash plugin install opencode
# Restart opencode
```
**Features:**
- ✅ Post-call hook support via custom plugin
- ✅ Token compression for tool outputs
- ✅ Daemon integration via Unix socket
- ✅ Fail-open (no compression if daemon down)

**Note:** Uses a custom JavaScript plugin installed to `.opencode/plugin/slash-plugin.js`. Plugin hooks `tool.execute.after` to send tool outputs to the Slash daemon for compression.

---

### MCP-Based (Continue)

**What it is:** Open-source AI code assistant  
**Platform:** VS Code extension / IDE  
**Installation:**
```bash
slash plugin install continue
# Restart Continue
```
**Features:**
- ✅ MCP server integration (retrieve, repomap, stats)
- ✅ Works alongside any model in Continue
- ⚠️ No pre/post hook support (MCP only)

**Note:** Continue doesn't expose native tool hooks. Slash provides MCP-based retrieval and repo mapping inside Continue's chat context.

---

### MCP-Based (Cline)

**What it is:** VS Code AI coding agent  
**Platform:** VS Code extension  
**Installation:**
```bash
slash plugin install cline
# Restart VS Code
```
**Features:**
- ✅ MCP server integration (retrieve, repomap, stats)
- ✅ Works with Cline's agent loop
- ⚠️ No pre/post hook support (MCP only)

**Note:** Cline supports MCP servers via `~/.cline/mcp_settings.json`. Slash registers the retrieve/repomap/stats tools for use in Cline's chat.

---

### MCP-Based (Goose)

**What it is:** AI coding agent framework  
**Platform:** CLI  
**Installation:**
```bash
slash plugin install goose
# Restart Goose
```
**Features:**
- ✅ MCP server integration (retrieve, repomap, stats)
- ⚠️ No pre/post hook support (MCP only)

---

### MCP-Based (PearAI)

**What it is:** AI code editor (Continue fork)  
**Platform:** Standalone editor  
**Installation:**
```bash
slash plugin install pearai
# Restart PearAI
```
**Features:**
- ✅ MCP server integration (retrieve, repomap, stats)
- ⚠️ No pre/post hook support (MCP only)

**Note:** PearAI shares Continue's config format. Slash registers MCP tools for use inside PearAI's chat.

---

## Architecture by Host

### Hook-Based (Claude Code, Codex, Antigravity, Copilot, Cursor, Windsurf)

```
Host Editor
  ↓ (hook event: JSON)
Hook Adapter (host-specific)
  ↓ (HookEvent)
Daemon
  ↓ Compress
  ↓ (HookResult)
Hook Adapter (encode back)
  ↓ (host JSON)
Host Editor (receives compressed output + retrieve tool)
```

**Adapters:**
- `claudecode.go` — camelCase JSON
- `codex.go` — snake_case JSON
- `cursor.go` — camelCase JSON (same as Claude Code)
- `windsurf.go` — camelCase JSON (same as Claude Code)
- `antigravity.go` — hooks.json format
- `copilot.go` — rich event types

**Shared:** All use the same `HookEvent` → `compress()` → `HookResult` path.

### CLI-Based (Aider)

```
Aider CLI
  ↓ (Python hook script)
Unix Socket
  ↓ (JSON)
Daemon
  ↓ Compress
  ↓ (JSON)
Aider CLI (continues with compressed output)
```

### MCP-Based (Zed)

```
Zed Editor
  ↓ (MCP client)
MCP Server (slash mcp --port 8765)
  ↓ JSON-RPC
retrieve(handle), repomap(), stats()
```

---

## Compression Results by Host

| Host | Avg Reduction | Notes |
|---|---|---|
| **Claude Code** | 48% | Full hook support; baseline |
| **Cursor** | 48% | Same as Claude Code |
| **Windsurf** | 48% | Same as Claude Code |
| **Codex** | 35–42% | Pre-call only; less aggressive |
| **Antigravity** | 46% | Full support |
| **Copilot CLI** | 5–15% | Incremental (native compaction ~95%) |
| **Aider** | 45% | Python hook integration |

**Pass-Rate:** 67% (vs. 69% baseline) across all hosts. Compression doesn't break functionality.

---

## Installation Checklist

For **each editor** you use:

1. **Install binary:** `slash version` (should work)
2. **Install plugin:** `slash plugin install <host>`
3. **Restart editor:** Close and reopen
4. **Verify:** See `[slash: ...]` hints in tool output
5. **Check daemon:** `slash stats`

---

## Troubleshooting by Host

### Claude Code / Cursor / Windsurf
- **No compression?** Check VSCode Extensions tab; Slash should be enabled
- **Latency high?** Disable output compression: `{"compression": {"output_compress": false}}`
- **Cache errors?** `slash cache check <file>`

### Codex
- **Pre-call rewriting not working?** Verify hook is registered in config.toml
- **Post-call compression missing?** Expected; Codex has limited output interception

### Antigravity
- **Hook not firing?** Check `~/.config/agy/hooks.json`
- **Plugin install fails?** Ensure `agy` CLI is on PATH

### Copilot CLI
- **Overlapping compression?** Copilot does native compaction first; Slash is additive
- **Disable if redundant?** Set `{"compression": {"enabled": false}}` in config

### Aider
- **Python hook error?** Check Python 3.7+ is installed
- **Socket not found?** Daemon may not have started; run `slash daemon`

### Zed
- **MCP not loading?** Verify `mcp_config.json` path and daemon is running
- **retrieve() not available?** Restart Zed; MCP server needs full init

---

## Which Host Should I Use?

| Goal | Best Choice |
|---|---|
| **Fastest iteration** | Claude Code (full support, zero config) |
| **VSCode ecosystem** | Cursor or Windsurf (same compression as Claude Code) |
| **Command-line workflow** | Aider (Python-based) or Antigravity (Go-based) |
| **Lightweight SSH** | Copilot CLI (already compressed natively) |
| **Multiple agents** | Zed (run any agent inside; MCP available) |
| **Maximum compatibility** | Use any; Slash adapts to each |

---

## Contributing New Hosts

Want to add support for another editor (e.g., Goose, Continue, etc.)?

1. Read [CONTRIBUTING.md](../CONTRIBUTING.md)
2. Create `internal/adapters/newtool.go` (copy the template)
3. Add fixtures in `internal/adapters/fixtures/newtool_*.json`
4. Write tests in `schema_test.go`
5. Update `registry.go` and `schema.go`
6. Submit PR

**Timeline:** ~2–3 hours per host with the template.

---

## Summary

Slash supports **13 editors across 3 categories** (IDE, CLI, MCP). One core engine; thin adapters per host. Compression is **reversible** (retrieve() tool), **fail-open** (daemon down → no compression, session continues), and **honest** (real benchmarks, confidence ranges).

**Start with your editor.** They all work the same way: install, restart, done.

Questions? See [README.md](../README.md) or [INSTALLATION.md](INSTALLATION.md).
