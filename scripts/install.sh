#!/bin/bash
set -e

# Slash Installation Script
# Usage: curl -sSL https://raw.githubusercontent.com/slash/slash/main/scripts/install.sh | bash
# Or: bash install.sh

SLASH_VERSION="${SLASH_VERSION:-1.0.0}"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
CONFIG_DIR="${HOME}/.slash"
CACHE_DIR="${XDG_CACHE_HOME:-$HOME/.cache}/slash"
REPO_URL="https://github.com/yoosuf/Slash"
RELEASE_URL="${REPO_URL}/releases/download/v${SLASH_VERSION}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Functions
log_info() {
    echo -e "${GREEN}[✓]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[!]${NC} $1"
}

log_error() {
    echo -e "${RED}[✗]${NC} $1"
}

log_step() {
    echo -e "${BLUE}==>${NC} $1"
}

# Detect OS and architecture
detect_os() {
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        OS="linux"
        ARCH=$(uname -m)
        case $ARCH in
            x86_64) ARCH="amd64" ;;
            aarch64) ARCH="arm64" ;;
            armv7l) ARCH="armv7" ;;
            *) log_error "Unsupported architecture: $ARCH"; exit 1 ;;
        esac
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        OS="darwin"
        ARCH=$(uname -m)
        case $ARCH in
            arm64) ARCH="arm64" ;;
            x86_64) ARCH="amd64" ;;
            *) log_error "Unsupported architecture: $ARCH"; exit 1 ;;
        esac
    else
        log_error "Unsupported OS: $OSTYPE"
        exit 1
    fi
}

# Check dependencies
check_deps() {
    log_step "Checking dependencies..."

    if ! command -v curl &> /dev/null; then
        log_error "curl is required but not installed"
        exit 1
    fi

    log_info "curl found"
}

# Download binary
download_binary() {
    log_step "Downloading Slash v${SLASH_VERSION}..."

    BINARY_URL="${RELEASE_URL}/slash_${SLASH_VERSION}_${OS}_${ARCH}.tar.gz"
    CHECKSUM_URL="${RELEASE_URL}/SHA256SUMS"

    TEMP_DIR=$(mktemp -d)
    trap "rm -rf $TEMP_DIR" EXIT

    log_info "OS: $OS, Architecture: $ARCH"
    log_info "Download URL: $BINARY_URL"

    if ! curl -fsSL "$BINARY_URL" -o "$TEMP_DIR/slash.tar.gz"; then
        log_error "Failed to download Slash"
        exit 1
    fi

    log_info "Downloaded successfully"

    # Verify with checksum
    if curl -fsSL "$CHECKSUM_URL" -o "$TEMP_DIR/SHA256SUMS" 2>/dev/null; then
        cd "$TEMP_DIR"
        if ! sha256sum -c SHA256SUMS 2>/dev/null | grep -q "slash_.*OK"; then
            log_warn "Could not verify checksum (continuing anyway)"
        else
            log_info "Checksum verified"
        fi
        cd -
    fi

    # Extract
    tar -xzf "$TEMP_DIR/slash.tar.gz" -C "$TEMP_DIR"
    echo "$TEMP_DIR"
}

# Install binary
install_binary() {
    local temp_dir=$1
    log_step "Installing binary to ${INSTALL_DIR}..."

    if [[ ! -w "$INSTALL_DIR" ]]; then
        log_warn "No write permission to ${INSTALL_DIR}, will use sudo"
        sudo cp "$temp_dir/slash" "${INSTALL_DIR}/slash"
        sudo chmod +x "${INSTALL_DIR}/slash"
    else
        cp "$temp_dir/slash" "${INSTALL_DIR}/slash"
        chmod +x "${INSTALL_DIR}/slash"
    fi

    log_info "Binary installed: ${INSTALL_DIR}/slash"
}

# Setup configuration
setup_config() {
    log_step "Setting up configuration..."

    mkdir -p "$CONFIG_DIR"
    mkdir -p "$CACHE_DIR"

    if [[ ! -f "$CONFIG_DIR/config.json" ]]; then
        cat > "$CONFIG_DIR/config.json" << 'EOF'
{
  "daemon": {
    "socket": "$HOME/.slash/daemon.sock",
    "port": 0,
    "log_level": "warn",
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
        log_info "Config created: ${CONFIG_DIR}/config.json"
    else
        log_warn "Config already exists, skipping"
    fi
}

# Verify installation
verify_install() {
    log_step "Verifying installation..."

    if ! command -v slash &> /dev/null; then
        log_error "Slash not found in PATH"
        log_error "Try adding ${INSTALL_DIR} to your PATH:"
        log_error "  export PATH=\"${INSTALL_DIR}:\$PATH\""
        exit 1
    fi

    VERSION=$(slash version 2>/dev/null || echo "unknown")
    log_info "Successfully installed: $VERSION"
}

# Main
main() {
    clear
    echo "================================"
    echo "  Slash Installation Script"
    echo "================================"
    echo ""

    detect_os
    check_deps

    TEMP_DIR=$(download_binary)
    install_binary "$TEMP_DIR"
    setup_config
    verify_install

    echo ""
    echo "================================"
    echo -e "${GREEN}Installation Complete!${NC}"
    echo "================================"
    echo ""
    echo "Next steps:"
    echo "1. Install plugin for your editor:"
    echo "   slash plugin install claude-code"
    echo "   (or: cursor, windsurf, codex, antigravity, copilot, aider)"
    echo ""
    echo "2. Restart your editor"
    echo ""
    echo "3. Start using Slash:"
    echo "   slash stats"
    echo ""
    echo "Documentation: https://github.com/yoosuf/Slash"
    echo ""
}

main "$@"
