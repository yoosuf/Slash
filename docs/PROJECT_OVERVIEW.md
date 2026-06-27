# Slash — Complete Project Overview

**Status:** v1.1.0 Complete & Production-Ready  
**License:** Apache 2.0  
**Repository:** https://github.com/yoosuf/Slash

---

## What You Have

A **complete, production-ready token reduction plugin** that works across 8 major agentic coding tools:

| Tool | Type | Install |
|---|---|---|
| **Claude Code** | IDE | `slash plugin install claude-code` |
| **Cursor** | IDE | `slash plugin install cursor` |
| **Windsurf** | IDE | `slash plugin install windsurf` |
| **Codex** (OpenAI) | CLI | `slash plugin install codex` |
| **Antigravity / agy** | CLI | `slash plugin install antigravity` |
| **Copilot CLI** | CLI | `slash plugin install copilot` |
| **Aider** | CLI/Python | `slash plugin install aider` |
| **Zed** | Editor | `slash plugin install zed` |

## Key Achievements

| Component | Status | Details |
|---|---|---|
| **Core Daemon** | ✅ Complete | Socket server, compression routing, session tracking |
| **8 Host Adapters** | ✅ Complete | All real implementations, not stubs |
| **Structural Compression** | ✅ Complete | 50+ languages, JSON, logs, text |
| **LCS Diff Engine** | ✅ Complete | Diff-only re-reads (80–95% savings) |
| **Repo Map** | ✅ Complete | Regex-based symbol extraction |
| **CCR Cache** | ✅ Complete | SQLite-backed, TTL + LRU eviction |
| **MCP Server** | ✅ Complete | retrieve(), repomap(), stats() tools |
| **Plugin System** | ✅ Complete | install/ls/uninstall for all hosts |
| **CLI** | ✅ Complete | daemon, hook, plugin, cache, audit, bench, stats, mcp, purge |
| **Config File** | ✅ Complete | `~/.slash/config.json` |
| **VS Code Extension** | ✅ Complete | Status bar, toggle, stats panel |
| **CI/CD** | ✅ Complete | GitHub Actions (test, lint, release) |
| **Packaging** | ✅ Complete | Homebrew, Scoop, .deb, .rpm, Winget |
| **Tests** | ✅ Complete | Adapter contracts, compression, diff, repo map |

## Architecture

```
Your Tool (Claude Code / Cursor / Codex / etc.)
  ↓ Hook fires (JSON)
Host Adapter (translates to canonical schema)
  ↓ HookEvent
Daemon (warm Go process)
  • Router (detect type)
  • Compressors (JSON/code/logs/text)
  • LCS Diff Engine (re-reads)
  • CCR Cache (SQLite)
  • Repo Map Indexer
  • Read-State Tracker
  • Session Metrics
  ↓ HookResult
Host Adapter (encode back)
  → Compressed output + retrieve() tool
```

## Core Components

| Package | Responsibility |
|---|---|
| `cmd/slash/main.go` | CLI dispatcher |
| `internal/daemon/` | Long-running compression server |
| `internal/adapters/` | 8 host adapters (Claude Code, Cursor, Windsurf, Codex, Antigravity, Copilot, Aider) |
| `internal/compress/` | Content detection, structural skeletonizers, LCS diff |
| `internal/store/` | SQLite CCR cache |
| `internal/repomap/` | Regex-based symbol index |
| `internal/track/` | File read-state tracking |
| `internal/mcp/` | MCP server (retrieve, repomap, stats) |
| `internal/client/` | Daemon socket client |
| `internal/plugin/` | Plugin install/ls/uninstall |
| `internal/config/` | `~/.slash/config.json` loader |

## CLI Commands

| Command | Description |
|---|---|
| `slash daemon` | Start compression daemon |
| `slash hook` | Process hook event (stdin/stdout, for editor hooks) |
| `slash plugin install <host>` | Auto-wire into editor |
| `slash plugin ls` | List installed plugins |
| `slash plugin uninstall <host>` | Remove plugin |
| `slash audit` | Compression breakdown |
| `slash bench` | Run benchmarks |
| `slash stats` | Live daemon stats |
| `slash cache ls` | List cache |
| `slash purge` | Wipe cache |
| `slash mcp` | Start MCP server |

## Quick Start

```bash
# Install with Homebrew (macOS/Linux)
brew install yoosuf/tap/slash

# Or Scoop (Windows)
scoop bucket add yoosuf https://github.com/yoosuf/scoop-bucket
scoop install yoosuf/slash

# Verify
slash version

# Install for your editor
slash plugin install claude-code
```

## Packaging

| Platform | Method |
|---|---|
| macOS / Linux | `brew install yoosuf/tap/slash` |
| Windows | `scoop install yoosuf/slash` |
| Ubuntu / Debian | `.deb` from releases |
| Fedora / RHEL | `.rpm` from releases |
| VS Code | `extensions/vscode/` — build and install `.vsix` |

## License

Apache 2.0. See [LICENSE.md](../LICENSE.md).
