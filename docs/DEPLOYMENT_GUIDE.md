# Slash Deployment & Operations Guide

This guide describes how to build, install, run, and maintain the **Slash** daemon and its associated adapters in development, staging, or production environments.

---

## 1. Prerequisites

- **Go**: Version 1.21 or higher.
- **Operating Systems**: macOS 10.15+, Linux (glibc 2.28+), or Windows 10+.
- **Build Tools**: `make` (optional, for convenience) and `gcc` (for tree-sitter C-binding support).

---

## 2. Building from Source

To build a fresh production-ready binary of Slash, run:

```bash
# Clone the repository
git clone https://github.com/yoosuf/Slash.git
cd slash

# Fetch dependencies and build
go mod tidy
make build
```

This compiles the binary to the root directory as `slash` (or `slash.exe` on Windows).

To install the binary globally:

```bash
sudo mv slash /usr/local/bin/
```

---

## 3. Daemon Deployment

Slash runs as a user-level background daemon to ensure fast response times (<50ms per hook). It should **never** be run as the root user; run it under the same user account as your IDE/agent host.

### Starting Manually

```bash
slash daemon &
```

### Auto-Starting at Login (Recommended)

#### macOS (Launchd)

Create a LaunchAgent file at `~/Library/LaunchAgents/com.slash.daemon.plist`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.slash.daemon</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/slash</string>
        <string>daemon</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>/Users/your-username/.slash/daemon.stdout.log</string>
    <key>StandardErrorPath</key>
    <string>/Users/your-username/.slash/daemon.stderr.log</string>
</dict>
</plist>
```

Load the daemon:
```bash
launchctl bootstrap gui/$(id -u) ~/Library/LaunchAgents/com.slash.daemon.plist
```

#### Linux (Systemd User Unit)

Create a systemd unit file at `~/.config/systemd/user/slash.service`:

```ini
[Unit]
Description=Slash context compression daemon
After=network.target

[Service]
ExecStart=/usr/local/bin/slash daemon
Restart=on-failure
RestartSec=5
StandardOutput=append:%h/.slash/daemon.stdout.log
StandardError=append:%h/.slash/daemon.stderr.log

[Install]
WantedBy=default.target
```

Enable and start the service:
```bash
systemctl --user daemon-reload
systemctl --user enable slash.service
systemctl --user start slash.service
```

---

## 4. Configuration

All configuration is managed via a JSON file placed at `~/.slash/config.json`.

```json
{
  "daemon": {
    "socket": "$HOME/.slash/daemon.sock",
    "port": 8080,
    "log_level": "warn"
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
  }
}
```

---

## 5. Host Adapter Installation

Once the binary is on your `PATH` and the daemon is running, install the integration plugins for your editor of choice:

```bash
# Claude Code integration
slash plugin install claude-code

# Cursor integration
slash plugin install cursor

# Windsurf integration
slash plugin install windsurf

# Copilot CLI integration
slash plugin install copilot
```

---

## 6. Maintenance & Troubleshooting

### Log Locations
Standard log locations:
- **Daemon Logs**: `~/.slash/daemon.stderr.log`
- **Hook Logs**: In your host tool's respective plugin log folder.

### Monitoring Health
Check daemon status and cache size:
```bash
slash stats
```

### Cache Management
Slash automatically prunes files older than the configured TTL (default 24h) and runs an LRU eviction cycle if the cache directory exceeds the maximum size (default 1GB).

To manually purge the cache at any time:
```bash
slash purge
```
To set up a cron job for weekly cache purges:
```bash
0 0 * * 0 /usr/local/bin/slash purge > /dev/null 2>&1
```
