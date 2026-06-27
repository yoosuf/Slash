# Installation Guide

## Quick Install

### macOS / Linux (Homebrew)
```bash
brew install yoosuf/tap/slash
```

### Windows (Scoop)
```bash
scoop bucket add yoosuf https://github.com/yoosuf/scoop-bucket
scoop install yoosuf/slash
```

### Ubuntu / Debian
```bash
curl -sSL https://github.com/yoosuf/Slash/releases/download/v1.0.0/slash-linux-amd64.deb -o slash.deb
sudo dpkg -i slash.deb
```

### Fedora / RHEL
```bash
curl -sSL https://github.com/yoosuf/Slash/releases/download/v1.0.0/slash-linux-amd64.rpm -o slash.rpm
sudo rpm -i slash.rpm
```

### Direct Download
```bash
# macOS ARM64
curl -sSL https://github.com/yoosuf/Slash/releases/download/v1.0.0/slash-darwin-arm64.tar.gz | tar xz
sudo mv slash-darwin-arm64 /usr/local/bin/slash

# macOS x86_64
curl -sSL https://github.com/yoosuf/Slash/releases/download/v1.0.0/slash-darwin-amd64.tar.gz | tar xz
sudo mv slash-darwin-amd64 /usr/local/bin/slash

# Linux x86_64
curl -sSL https://github.com/yoosuf/Slash/releases/download/v1.0.0/slash-linux-amd64.tar.gz | tar xz
sudo mv slash-linux-amd64 /usr/local/bin/slash

# Windows
# Download and unzip: https://github.com/yoosuf/Slash/releases/download/v1.0.0/slash-windows-amd64.zip
```

### From Source
```bash
git clone https://github.com/yoosuf/Slash.git
cd slash
CGO_ENABLED=1 go build -o ~/.local/bin/slash ./cmd/slash
```

## Verify
```bash
slash version
```

## Plugin Install
```bash
# Pick your editor/tool
slash plugin install claude-code   # Claude Code
slash plugin install cursor        # Cursor
slash plugin install windsurf      # Windsurf
slash plugin install codex         # Codex (OpenAI)
slash plugin install antigravity   # Antigravity (agy)
slash plugin install copilot       # Copilot CLI
slash plugin install aider         # Aider
slash plugin install zed           # Zed
```

Restart your editor/tool after installing.

## VS Code Extension
```bash
cd extensions/vscode
npm install && npm run package
code --install-extension slash-compressor-1.0.0.vsix
```

## Uninstall
```bash
slash plugin uninstall <host>  # remove plugin from editor
slash purge --confirm          # clear cache
sudo rm /usr/local/bin/slash   # remove binary
```
