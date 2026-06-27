# Slash Project Roadmap

This roadmap outlines the planned development phases, features, and target timelines for **Slash**.

---

## ✅ V1.0 — Released

**Core Engine**
- Warm daemon over Unix domain socket with session tracking
- Structural compression for 50+ languages (braced, Python, Ruby, Shell, SQL, HTML/XML, CSS, structured data, Markdown, Makefile)
- LCS-based diff-only re-reads — unchanged files return `[content unchanged]`
- Regex-based repo map symbol extraction (Go, TS, JS, Python, Java, Rust, C#, Kotlin, Swift, PHP, Dart, Scala, Ruby, Lua, R, Haskell, Elixir, Erlang, Groovy, Clojure, F#, OCaml, Fortran, Ada, Protocol Buffers, SQL, Shell, Crystal)
- SQLite-backed CCR cache with TTL and LRU eviction
- Read-state tracking per session

**Host Support**
- 8 fully implemented adapters: Claude Code, Cursor, Windsurf, Codex, Antigravity, Copilot CLI, Aider, Zed (MCP)
- MCP server with `retrieve`, `repomap`, `stats` tools

**Packaging & Distribution**
- Homebrew tap: `brew install yoosuf/tap/slash`
- Scoop bucket: `scoop install yoosuf/slash`
- Release pipeline producing tarballs, .deb, .rpm, .zip + SHA256SUMS

---

## ✅ V1.1 — Polish & Plugin System (Released)

- [x] **`slash plugin install`** — Auto-wire into Claude Code, Cursor, Codex, Antigravity, Copilot CLI, Aider, Zed
- [x] **`slash plugin ls / uninstall`** — List and remove plugin integrations
- [x] **`slash audit`** — Compression breakdown by file, token savings report (sort by savings/file/type, JSON output)
- [x] **`slash bench`** — Built-in benchmark command (JSON, code, logs, text compression benchmarks)
- [x] **Interactive stats** — `slash stats` connects to daemon socket for live session metrics
- [x] **Config file support** — `~/.slash/config.json` wired into daemon startup

---

## 🚀 V1.2 — Deeper Integrations & Distribution (Target: Q1 2027)

- [ ] **Winget (Microsoft Store)** — Submit Winget manifests to `microsoft/winget-pkgs`
- [x] **VS Code extension** — Native extension with status bar indicator, toggle, stats
- [ ] **JetBrains plugin** — Support for IntelliJ, PyCharm, GoLand
- [ ] **NPM package** — Install via `npx slash` for Node.js ecosystem
- [ ] **Network proxy fallback** — HTTP proxy for editors without native hook APIs
- [ ] **Interactive CLI** — `slash ui` for cache inspection, compression ratios, ignore patterns
- [ ] **Config UI** — `slash config` wizard for setting up compression preferences

---

## 🧠 V2.0 — Intelligence & Teams (Target: 2027+)

- [ ] **Learned compression policies** — Local lightweight models to predict critical vs. compressible content per session, targeting 80%+ compression without quality loss
- [ ] **Team shared cache (opt-in)** — Secure remote cache server for team-wide skeleton and retrieval sharing
- [ ] **Multi-repo map** — Index across multiple repositories for cross-project agent context
- [ ] **Plugin SDK** — Public API for third-party compression strategies and host adapters
