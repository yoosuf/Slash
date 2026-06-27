# Project Renaming: Slash → Slash

This document tracks the renaming of the Slash project to Slash.

## Scope

The project has been renamed from **Slash** to **Slash** across all:
- Binary names
- Package names
- Configuration paths
- Documentation
- Comments and strings
- Plugin references
- GitHub repository references

## Changes Made

### 1. Code Changes

#### Go Module
- `go.mod`: `github.com/yoosuf/Slash` → `github.com/yoosuf/Slash`

#### Package Imports
All internal imports updated to reflect new module path:
```go
// Before
import "github.com/yoosuf/Slash/internal/daemon"

// After
import "github.com/yoosuf/Slash/internal/daemon"
```

#### Binary Names
- `slash` → `slash`
- All CLI help text updated
- All example commands updated (e.g., `slash version` → `slash version`)

#### Configuration Paths
- `~/.slash/` → `~/.slash/`
- `$XDG_CACHE_HOME/slash/` → `$XDG_CACHE_HOME/slash/`

#### Log Prefixes
- `[slash]` → `[slash]`

#### Compression Metadata
- `[slash: JSON skeleton, 45% reduction]` → `[slash: JSON skeleton, 45% reduction]`
- All compression method messages updated

#### MCP Server Name
- `slash` → `slash`

### 2. Documentation Changes

#### README.md
- Project title: `Slash` → `Slash`
- Description updated
- All CLI commands updated
- GitHub URL updated
- Installation commands updated

#### All Documentation Files
Major files updated:
- `BENCHMARKING.md`
- `BENCHMARKING_SUMMARY.md`
- `FINAL_DELIVERY.md`
- `HOSTS.md`
- `INSTALLATION.md`
- `PROJECT_OVERVIEW.md`
- And more...

### 3. Configuration & Plugin References

#### Plugin Manifests
All plugin manifests in `plugin/*/` updated:
- `plugin/claude-code/manifest.json`
- `plugin/cursor/manifest.json`
- `plugin/windsurf/manifest.json`
- `plugin/codex/manifest.json`
- `plugin/antigravity/plugin.json`
- `plugin/copilot/plugin.json`
- `plugin/aider/plugin.json`

#### Hook Scripts
- JavaScript hooks in Claude Code, Cursor, Windsurf updated
- Python hook in Aider updated
- All socket/daemon references changed

## Migration Guide for Users

If you installed Slash previously, here's what to do:

### 1. Clear Old Installation
```bash
rm -rf ~/.slash/
rm -rf $XDG_CACHE_HOME/slash/
which slash && rm /usr/local/bin/slash
```

### 2. Install Slash
```bash
curl -sSL https://github.com/yoosuf/Slash/releases/download/v1.0.0/slash_install.sh | bash
```

### 3. Re-Install for Your Editor
```bash
slash plugin install claude-code  # or: cursor, windsurf, codex, antigravity, copilot, aider
```

### 4. Verify
```bash
slash version
slash stats
```

## Migration Guide for Developers

If you cloned the Slash repo, here's how to update:

### 1. Update Local Repo
```bash
git remote set-url origin https://github.com/yoosuf/Slash.git
git pull origin main
```

### 2. Update Imports
Run a search-replace in your IDE:
- Find: `github.com/yoosuf/Slash`
- Replace: `github.com/yoosuf/Slash`

### 3. Rebuild
```bash
go mod tidy
make build
```

### 4. Update Documentation
If you reference Slash, update to Slash.

## Files Changed Summary

### Core Code
- `go.mod` (module path)
- `cmd/slash/main.go` (binary name, help text, config paths)
- `internal/daemon/daemon.go` (log prefix)
- `internal/compress/compressor.go` (metadata strings)
- `internal/client/hookclient.go` (socket path helper)
- `internal/mcp/server.go` (MCP name)

### Plugin Bundles
- `plugin/claude-code/manifest.json`
- `plugin/cursor/manifest.json`
- `plugin/windsurf/manifest.json`
- `plugin/codex/manifest.json`
- `plugin/antigravity/plugin.json`
- `plugin/copilot/plugin.json`
- `plugin/aider/plugin.json`

### Documentation
- README.md
- INSTALLATION.md
- HOSTS.md
- ARCHITECTURE.md
- BENCHMARKING.md
- BENCHMARKING_SUMMARY.md
- FINAL_DELIVERY.md
- PROJECT_OVERVIEW.md
- And many more...

## Testing the Rename

### 1. Build
```bash
cd /path/to/slash
make clean
make build
```

### 2. Verify Binary
```bash
./slash version
# Output: Slash v1.0.0
```

### 3. Test Commands
```bash
./slash bench --help
./slash daemon --help
./slash plugin install --help
```

### 4. Check Configuration Paths
```bash
./slash daemon &
sleep 1
ls -la ~/.slash/daemon.sock
# Should exist
pkill slash
```

## Breaking Changes for End Users

- **Binary name**: `slash` → `slash`
- **Config directory**: `~/.slash/` → `~/.slash/`
- **Cache directory**: `$XDG_CACHE_HOME/slash/` → `$XDG_CACHE_HOME/slash/`
- **Plugin names**: All plugin names change (must reinstall)
- **Commands**: All commands now use `slash` instead of `slash`

Example: `slash plugin install claude-code` → `slash plugin install claude-code`

## No Breaking Changes for Functionality

The compression algorithm, pass-rates, latency, and all functionality remain identical. This is purely a naming change.

## Next Steps

1. Update GitHub repository name
2. Update marketplace listings (Claude Code, Copilot, etc.)
3. Update GitHub documentation/wiki
4. Announce to users
5. Provide migration guide

---

**Rename completed successfully.** Slash is ready to ship. 🚀
