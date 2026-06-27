# Slash: Local Installation & Testing Guide

**Complete step-by-step guide to install Slash on your machine for testing.**

---

## Quick Start (TL;DR)

```bash
# 1. Clone the repository
git clone https://github.com/yoosuf/Slash.git
cd slash

# 2. Build the binary
make build

# 3. Install it
sudo mv ./slash /usr/local/bin/

# 4. Verify
slash version

# 5. Install for your editor
slash plugin install claude-code
# or: cursor, windsurf, codex, antigravity, copilot, aider

# 6. Restart your editor and test!
```

**That's it.** Daemon starts automatically on first use.

---

## Detailed Installation Guide

### Step 1: Prerequisites

#### macOS

```bash
# Check if Go is installed (need 1.21+)
go version

# If not installed:
brew install go

# Verify
go version
→ go version go1.21.0 darwin/arm64
```

#### Linux

```bash
# Debian/Ubuntu
sudo apt-get update
sudo apt-get install golang-1.21 git make

# Or download from golang.org
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

#### Windows

```powershell
# Use Chocolatey
choco install golang

# Or download from golang.org
# https://go.dev/dl/go1.21.0.windows-amd64.msi
```

---

### Step 2: Clone Repository

```bash
# Clone Slash
git clone https://github.com/yoosuf/Slash.git
cd slash

# Verify you're in the right directory
ls -la
→ cmd/
  internal/
  plugin/
  go.mod
  Makefile
  README.md
```

---

### Step 3: Build the Binary

#### Option A: Using Make (Recommended)

```bash
# Build for your current OS
make build

# Binary created at: ./slash
ls -lh ./slash
→ -rwxr-xr-x  slash (5.2MB)

# Test it works
./slash version
→ Slash v1.0.0 (darwin/arm64)
```

#### Option B: Manual Build

```bash
# Build manually
go build -o slash ./cmd/slash

# Or with version info
VERSION=$(git describe --tags --always)
go build -ldflags="-X main.Version=$VERSION" -o slash ./cmd/slash
```

#### Option C: Cross-Compile (Advanced)

```bash
# Build for multiple platforms
make build-all

# Creates:
# ./bin/slash_linux_amd64
# ./bin/slash_darwin_arm64
# ./bin/slash_darwin_amd64
# ./bin/slash_windows_amd64.exe
```

---

### Step 4: Install Binary

#### macOS/Linux

```bash
# Copy to system path
sudo cp ./slash /usr/local/bin/

# Verify it's in PATH
which slash
→ /usr/local/bin/slash

# Make sure it's executable
chmod +x /usr/local/bin/slash

# Test
slash version
→ Slash v1.0.0 darwin/arm64
```

#### Windows

```powershell
# Copy to Program Files
Copy-Item .\slash.exe 'C:\Program Files\slash\slash.exe'

# Add to PATH (via System Properties)
# Or use environment variable:
$env:PATH += ";C:\Program Files\slash"

# Test
slash version
```

---

### Step 5: Verify Installation

```bash
# Check version
slash version
→ Slash v1.0.0

# Check daemon can start
slash daemon --help
→ [Shows help text]

# Check default config paths
echo $HOME/.slash
→ /Users/yourname/.slash

# List available commands
slash --help
```

---

### Step 6: Configure (Optional)

Slash works out-of-the-box with defaults, but you can customize:

```bash
# Create config directory
mkdir -p ~/.slash

# Create config file
cat > ~/.slash/config.json << 'EOF'
{
  "daemon": {
    "socket": "$HOME/.slash/daemon.sock",
    "port": 8765,
    "log_level": "info",
    "auto_start": true
  },
  "compression": {
    "enabled": true,
    "diff_only_reads": true,
    "output_compress": true,
    "repo_map_inject": true
  },
  "cache": {
    "dir": "$XDG_CACHE_HOME/slash",
    "ttl_hours": 24,
    "max_size_mb": 1024,
    "secret_patterns": [".env", "*_key*", "*.pem", "*.key"]
  },
  "telemetry": {
    "enabled": false
  }
}
EOF
```

**All fields are optional.** If missing, defaults are used.

---

### Step 7: Install for Your Editor

#### Claude Code

```bash
slash plugin install claude-code
```

**Setup:**
1. Restart Claude Code
2. Slash daemon starts automatically
3. Compression begins on first file read

**Verify:**
```bash
slash stats
→ {
    "total_calls": 5,
    "compression_rate": 0.58,
    "cache_entries": 3
  }
```

#### Cursor

```bash
slash plugin install cursor
```

**Setup:**
1. Restart Cursor
2. Check: Settings → Extensions → Slash
3. Should show "Active"

#### Windsurf

```bash
slash plugin install windsurf
```

**Setup:**
1. Restart Windsurf
2. Plugin auto-activates

#### Codex

```bash
slash plugin install codex
```

**Setup:**
1. Restart terminal
2. Use `codex` normally; Slash runs in background

#### Antigravity (agy)

```bash
slash plugin install antigravity
```

**Setup:**
1. Use `agy` as usual
2. Slash compression transparent

#### Copilot CLI

```bash
slash plugin install copilot
```

**Setup:**
1. Restart `copilot` CLI
2. Use normally

#### Aider

```bash
slash plugin install aider
```

**Setup:**
```bash
# Run aider with Slash enabled
aider --with-slash .

# Or set default
export SLASH_ENABLED=1
aider .
```

---

## Testing Slash

### Test 1: Verify Daemon Started

```bash
# Check if daemon is running
ps aux | grep "slash daemon"
→ username  12345  ... slash daemon

# Or check socket exists
ls -la ~/.slash/daemon.sock
→ srwxr-xr-x daemon.sock

# Get stats
slash stats
→ Compression stats (should show entries if active)
```

### Test 2: Manual Compression Test

```bash
# Create a test file
cat > /tmp/test_code.go << 'EOF'
package main

import "fmt"

func main() {
    fmt.Println("Hello, Slash!")
}

// ... imagine 100 more lines ...
EOF

# Audit compression
slash audit /tmp/test_code.go
→ File: test_code.go
  Size: 5.2 KB
  Detected: Go code
  Method: AST skeleton
  Reduction: 67%
  Compressed: 1.7 KB
```

### Test 3: Cache Verification

```bash
# List cache entries
slash cache ls
→ h_abc123 test_code.go (5.2 KB)
  h_def456 auth.go (8.3 KB)
  ...

# Check specific file
slash cache check auth.go
→ Found: h_def456
  Original: 8.3 KB
  Compressed: 2.1 KB
  Reduction: 75%
  TTL: 23h 45m remaining
```

### Test 4: Retrieve Test

```bash
# In your editor, ask Claude:
# "Show me the auth.go file"

# Claude gets: [skeleton + retrieve handle]
# 
# Then Claude can call:
# retrieve(h_def456)
#
# Should return: Full original file

# Or test manually:
curl -X POST http://localhost:8765/mcp/retrieve \
  -H "Content-Type: application/json" \
  -d '{"handle": "h_def456"}'

→ {
    "handle": "h_def456",
    "content": "[full file content]",
    "size": 8300
  }
```

### Test 5: End-to-End Test

```bash
# 1. Start fresh
slash purge
slash daemon &

# 2. Have Claude read a real file
# (in your IDE, ask Claude Code:
#  "Read handlers/auth.go and explain it")

# 3. Check compression happened
slash stats
→ total_calls: 1
  compression_rate: 0.65
  cache_entries: 1

# 4. Ask Claude to read same file again
# (you ask: "Can you show me the Login function?")

# 5. Verify diff-only re-read
slash stats
→ total_calls: 2
  compression_rate: 0.92  ← Much higher! (diff-only)

# 6. Have Claude retrieve full file
# ("I need the complete implementation")

# 7. Verify retrieve worked
curl -X POST http://localhost:8765/mcp/retrieve \
  -H "Content-Type: application/json" \
  -d '{"handle": "h_abc123"}'
→ [Full content returned]
```

---

## Testing Workflow: Debugging Session

### Scenario: Debug a Bug

```bash
# 1. Start with a real codebase
cd ~/my-project

# 2. Start Slash daemon
slash daemon &
→ [slash] Starting daemon on ~/.slash/daemon.sock

# 3. Open Claude Code
# Ask: "Why is auth failing on line 45?"

# Watch compression happen:
watch -n 1 'slash stats | head -20'

# 4. You should see:
#    - total_calls increases
#    - compression_rate grows (50-70%)
#    - cache_entries increases

# 5. Claude digs deeper, calls retrieve()
#    - compression_rate spikes (99%+)
#    - More entries added

# 6. Close Slash
pkill slash

# 7. Check final stats
slash stats
→ total_calls: 15
  compression_rate: 0.62 (average)
  cache_size_mb: 2.3
  latency_p95: 42ms
```

---

## Troubleshooting Local Install

### Issue: "bash: slash: command not found"

**Problem:** Binary not in PATH

**Fix:**
```bash
# Verify installation
which slash

# If not found, reinstall:
sudo cp ./slash /usr/local/bin/
chmod +x /usr/local/bin/slash

# Refresh shell
source ~/.bashrc  # or ~/.zshrc

# Test
slash version
```

---

### Issue: "Permission denied" when running slash

**Problem:** Binary not executable

**Fix:**
```bash
chmod +x /usr/local/bin/slash
slash version
```

---

### Issue: "Error: address already in use"

**Problem:** Daemon already running or port conflict

**Fix:**
```bash
# Kill existing daemon
pkill slash

# Check if socket exists
rm -f ~/.slash/daemon.sock

# Start fresh
slash daemon
```

---

### Issue: "Plugin install fails"

**Problem:** Editor not found or permissions issue

**Fix:**
```bash
# Verify editor is installed
which claude-code  # or cursor, windsurf, etc

# Try reinstalling plugin
slash plugin uninstall claude-code
slash plugin install claude-code

# Check plugin directory
ls -la ~/.slash/plugins/
```

---

### Issue: "Cache disk usage is high"

**Problem:** Cache growing too large

**Fix:**
```bash
# Clean cache
slash purge
→ Purged 234 entries, freed 512MB

# Check cache size
du -sh ~/.cache/slash
→ 1.2GB

# Reduce cache TTL
cat > ~/.slash/config.json << EOF
{
  "cache": {
    "ttl_hours": 12,
    "max_size_mb": 512
  }
}
EOF

# Restart daemon
pkill slash
slash daemon &
```

---

### Issue: "Compression not working"

**Problem:** No compression being applied

**Debug:**
```bash
# 1. Check daemon is running
ps aux | grep slash
→ Should see "slash daemon"

# 2. Check config
cat ~/.slash/config.json | grep enabled
→ Should see "enabled": true

# 3. Check stats
slash stats
→ Should show calls and compression_rate

# 4. Enable debug logging
cat > ~/.slash/config.json << EOF
{
  "daemon": {
    "log_level": "debug"
  }
}
EOF

# 5. Watch logs
tail -f ~/.slash/daemon.log
→ Should show compression events

# 6. Try manual audit
slash audit <filename>
→ Should show compression details
```

---

## Performance Testing

### Benchmark Your Installation

```bash
# Run built-in benchmarks
slash bench

→ Running benchmark suite...
  
  JSON Compression:
  ├─ Small (100B):   PASS  (98% reduction)
  ├─ Medium (1KB):   PASS  (75% reduction)
  ├─ Large (10KB):   PASS  (72% reduction)
  └─ Nested (5KB):   PASS  (80% reduction)
  
  Code Compression:
  ├─ Go (50 lines):   PASS  (65% reduction, <5ms)
  ├─ Python (200 l):  PASS  (68% reduction, <8ms)
  └─ TypeScript (1K): PASS  (71% reduction, <12ms)
  
  Log Compression:
  ├─ Errors (50):      PASS  (45% reduction)
  ├─ Repeated (2K):    PASS  (99% reduction)
  └─ Mixed (100):      PASS  (72% reduction)
  
  SUMMARY:
  ├─ Pass rate: 100% (12/12)
  ├─ Avg reduction: 74%
  ├─ Avg latency: <10ms (p95: <40ms)
  └─ ✓ All tests passed
```

---

## Development/Source Code Testing

If you want to test changes locally:

```bash
# 1. Make changes to source
vim internal/compress/compressor.go

# 2. Rebuild
make build

# 3. Stop running daemon
pkill slash

# 4. Test new binary
./slash daemon &

# 5. Run benchmarks against changes
./slash bench

# 6. Or test with your editor
# Restart Claude Code, test manually
```

---

## Uninstall

### Remove Slash Completely

```bash
# 1. Stop daemon
pkill slash

# 2. Remove binary
sudo rm /usr/local/bin/slash

# 3. Remove config/cache
rm -rf ~/.slash
rm -rf ~/.cache/slash  # or $XDG_CACHE_HOME/slash

# 4. Uninstall plugins (for each editor)
slash plugin uninstall claude-code  # etc

# 5. Verify removed
which slash
→ slash not found ✓
```

---

## Next Steps After Installation

### 1. Basic Testing

```bash
# Test with Claude Code
1. Open Claude Code
2. Ask it to read a file: "Show me main.go"
3. Check compression: slash stats
4. Ask to read again: compression should spike
5. Ask Claude to retrieve(): should get full file
```

### 2. Real Project Testing

```bash
# Try on your actual codebase
cd ~/my-project
slash audit .  # See what would compress
```

### 3. Integration Testing

```bash
# Test each editor you use
- Claude Code: slash plugin install claude-code
- Cursor: slash plugin install cursor
- etc.

# Verify each works independently
```

### 4. Performance Baseline

```bash
# Run benchmarks to establish baseline
slash bench --output baseline.json

# Use this to compare future changes
slash bench --output comparison.json
diff baseline.json comparison.json
```

---

## Advanced Local Testing

### Test Custom Compression Settings

```json
// Create ~/.slash/config.json with different settings
{
  "compression": {
    "code_skeleton_depth": "minimal",
    "json_skeleton_keep_values": false,
    "log_dedup_threshold": 5
  }
}
```

Then re-run benchmarks:
```bash
slash bench
# Should show different reduction percentages
```

### Test with Large Monorepo

```bash
# Clone a large open-source project
git clone https://github.com/kubernetes/kubernetes.git

# Test compression
slash audit kubernetes/
→ Should handle large codebase gracefully

# Monitor cache
watch -n 5 'du -sh ~/.cache/slash'
```

### Test Concurrent Access

```bash
# Start multiple Claude Code instances (if possible)
# Or simulate concurrent access:

# Terminal 1:
while true; do
  slash stats
  sleep 1
done

# Terminal 2:
# Use Claude Code to read files
# Stats should update in real-time without errors
```

---

## Getting Help

### Check Logs

```bash
# View daemon logs
tail -f ~/.slash/daemon.log

# Or check last error
grep ERROR ~/.slash/daemon.log | tail -20
```

### Enable Verbose Logging

```bash
cat > ~/.slash/config.json << EOF
{
  "daemon": {
    "log_level": "debug"
  }
}
EOF

pkill slash
slash daemon &
tail -f ~/.slash/daemon.log
```

### Run Diagnostics

```bash
# Check system info
slash --version
uname -a
go version

# Check installation
which slash
file /usr/local/bin/slash

# Check config
cat ~/.slash/config.json

# Check cache
du -sh ~/.cache/slash
ls -la ~/.cache/slash/

# Check daemon
ps aux | grep slash
netstat -ln | grep slash
```

---

## Verification Checklist

After installation, verify:

```bash
✓ slash binary in /usr/local/bin/
  → which slash

✓ slash version shows correct version
  → slash version

✓ config directory created
  → ls -la ~/.slash/

✓ daemon can start
  → slash daemon &
  → ps aux | grep slash

✓ socket file created
  → ls -la ~/.slash/daemon.sock

✓ plugin installed for your editor
  → slash plugin ls

✓ compression working
  → slash stats shows entries

✓ cache working
  → ls -la ~/.cache/slash/

✓ retrieve tool works
  → curl http://localhost:8765/mcp/retrieve ...

✓ benchmark passes
  → slash bench
```

---

## Quick Reference: Common Commands

```bash
# Management
slash daemon              # Start daemon
slash version             # Show version
slash --help             # Help text

# Plugins
slash plugin ls          # List installed plugins
slash plugin install <host>  # Install plugin
slash plugin uninstall <host> # Remove plugin

# Cache
slash cache ls           # List cache entries
slash cache check <file> # Check if file is cached
slash purge              # Clear all cache

# Monitoring
slash stats              # Compression statistics
slash audit <file>       # Analyze single file
slash bench              # Run performance benchmarks

# Configuration
~/.slash/config.json     # Config file location
```

---

**Ready to test locally!** 🚀

Once installed, start with the [SLASH_COMPREHENSIVE_GUIDE.md](SLASH_COMPREHENSIVE_GUIDE.md) for detailed usage examples.
