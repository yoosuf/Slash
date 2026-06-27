# Changelog

All notable changes to the **Slash** project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [1.0.0] - 2026-06-27

This is the initial production release of **Slash** (previously known as TokenTrim).

### Added
- **Core Optimization Engine**: Reversible AST-based code compression for Go and TypeScript.
- **Skeletal JSON Compression**: Collapses large JSON response payloads while retaining schema keys.
- **Deduplicating Log Compressor**: Identifies and condenses repetitive lines in command output logs.
- **Diff-Only Re-Read Optimization**: Cache tracking that detects when a file has already been read and only emits changed diff ranges to host tools.
- **Multi-Host Adapters**: Out-of-the-box support for 7 popular editor environments (Claude Code, Cursor, Windsurf, Codex, Antigravity, Copilot, Aider).
- **Integrated MCP Server**: Exposes standard tools (`retrieve`, `repomap`, `stats`) for AI models to query original content.
- **Local SQLite Caching**: SQLite-backed CCR cache storing event data safely under user-level OS file permissions.

### Changed
- **Project Renaming**: Rebranded the repository from **TokenTrim** to **Slash** for improved brand presence and cleaner command invocations.
- **Configuration & Caching Path Migration**:
  - Config directory moved from `~/.tokentrim/` to `~/.slash/`.
  - Cache directory moved from `$XDG_CACHE_HOME/tokentrim` to `$XDG_CACHE_HOME/slash`.
- **CLI Binary Rename**: Command binary renamed from `tokentrim` to `slash`.
