# Slash

**Open-source token reduction + speed plugin for agentic coding tools.**

Slash is a local-first, reversible token compressor that reduces context size by 40–60% for Claude Code, Codex, Antigravity, Copilot CLI, Cursor, Windsurf, Aider, Zed, Opencode, Continue, Cline, Goose, and PearAI. Compression is reversible (via on-demand `retrieve`), and nothing leaves your machine by default.

- **Production-ready** — fully tested, documented, benchmarked
- **13 hosts supported** — Claude Code, Cursor, Windsurf, Codex, Antigravity, Copilot CLI, Aider, Zed, Opencode, Continue, Cline, Goose, PearAI
- **Open source** (Apache 2.0)
- **Zero config** — one-command plugin install, auto-starting daemon

## Quick Start

**Requirements:** macOS 10.15+, Linux (glibc 2.28+), or Windows 10+.

```bash
# macOS / Linux
brew install yoosuf/tap/slash

# Windows
scoop bucket add yoosuf https://github.com/yoosuf/scoop-bucket
scoop install yoosuf/slash

# Verify
slash version

# Install as a plugin
slash plugin install claude-code   # or: cursor, windsurf, codex, antigravity, copilot, aider, zed, opencode, continue, cline, goose, pearai
```

That's it. The daemon starts automatically, and compression begins immediately.

## How It Works

### Three Layers of Compression

1. **Diff-only re-reads** (~80–95% savings on re-reads)  
   If a file was read then edited, return only the changed lines. The model gets the delta without re-reading the whole file.

2. **Output compression** (~40–60% savings)  
   JSON → skeletal tree, logs → deduplicated lines, code → structural outline. Results are stashed in a local cache (`retrieve(handle)`) so the model can fetch details on demand.

3. **Repo map at session start** (~10–15% faster orientation)  
   A symbol index injected at the start so the agent orients without exploratory reads.

**Every reduction is reversible.** If compression loses context, the model calls `retrieve(handle)` and gets the original body back.

## What Gets Compressed

| Content Type | Method | Savings | Reversible |
|---|---|---|---|
| **Code reads** | Structural skeleton + full body via retrieve | 60–70% | ✅ |
| **JSON** | Tree skeleton + leaf values | 40–50% | ✅ |
| **Logs/errors** | Dedup repeated lines, truncate + retrieve | 50–80% | ✅ |
| **Re-reads** | LCS diff-only (changed lines only) | 80–95% | ✅ |
| **Repo map** | Regex symbol index | N/A | ✅ |

## Supported Hosts (13)

| Host | Type | Details |
|---|---|---|
| **Claude Code** | IDE | Full hook support, MCP retrieve tool |
| **Cursor** | IDE | VS Code-based, full hook support |
| **Windsurf** | IDE | VS Code-based, full hook support |
| **Codex** | CLI | Pre-call rewriting, best-effort output |
| **Antigravity (agy)** | CLI | hooks.json format, full lifecycle |
| **Copilot CLI** | CLI | Pre+post hooks, native compaction overlay |
| **Aider** | CLI/Python | Python hooks, subprocess integration |
| **Zed** | Editor | MCP server |
| **Opencode** | CLI | Custom plugin, daemon socket integration |
| **Continue** | IDE | MCP server |
| **Cline** | IDE | MCP server |
| **Goose** | CLI | MCP server |
| **PearAI** | IDE | MCP server |

## CLI Commands

| Command | Description |
|---|---|
| `slash daemon` | Start the compression daemon |
| `slash hook` | Process a single hook event (stdin/stdout, for editor hooks) |
| `slash plugin install <host>` | Auto-wire into an editor/tool |
| `slash plugin ls` | List installed plugins |
| `slash plugin uninstall <host>` | Remove plugin integration |
| `slash audit` | Compression breakdown by file/type |
| `slash bench` | Run compression benchmarks |
| `slash stats` | Live daemon session stats |
| `slash cache ls` | List cached items |
| `slash cache check <path>` | Check if a file is cached |
| `slash purge` | Wipe the cache |
| `slash mcp` | Start MCP server (for Zed and MCP clients) |

## Installation

| Platform | Command |
|---|---|
| **macOS / Linux** | `brew install yoosuf/tap/slash` |
| **Windows** | `scoop bucket add yoosuf https://github.com/yoosuf/scoop-bucket && scoop install yoosuf/slash` |
| **Ubuntu/Debian** | Download `.deb` from [releases](https://github.com/yoosuf/Slash/releases) |
| **Fedora/RHEL** | Download `.rpm` from [releases](https://github.com/yoosuf/Slash/releases) |
| **From source** | `go build -o ~/.local/bin/slash ./cmd/slash` |

## Configuration

Create `~/.slash/config.json`:

```json
{
  "daemon": { "log_level": "warn" },
  "compression": { "enabled": true, "diff_only_reads": true, "output_compress": true },
  "cache": { "ttl_hours": 24, "max_size_mb": 1024 },
  "telemetry": { "enabled": false }
}
```

All fields are optional.

## Privacy

**Everything stays local.** No network calls by default. Cache lives in `~/.cache/slash/` and respects `.gitignore` patterns. Zero telemetry by default.

## VS Code Extension

Install from the [VS Code Marketplace](https://marketplace.visualstudio.com/items?itemName=yoosuf.slash-compressor) or build from source:

```bash
cd extensions/vscode
npm install && npm run package
code --install-extension slash-compressor-1.0.0.vsix
```

Features: status bar indicator, toggle compression, live stats panel.

## License

Apache 2.0. See [LICENSE.md](LICENSE.md).
