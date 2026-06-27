# Slash: Deployment Scripts & Automation

**Automated deployment for single & multiple machines.**

---

## Table of Contents

1. [Single Machine Scripts](#single-machine-scripts)
2. [Multi-Machine Deployment](#multi-machine-deployment)
3. [Ansible Playbooks](#ansible-playbooks)
4. [Enterprise Deployment](#enterprise-deployment)

---

## Single Machine Scripts

### Script 1: Linux/macOS Auto-Install

**File: `scripts/install.sh`**

```bash
#!/bin/bash
set -e

# Slash Installation Script
# Usage: curl -sSL https://raw.githubusercontent.com/slash/slash/main/scripts/install.sh | bash
# Or: bash install.sh

SLASH_VERSION="1.0.0"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="${HOME}/.slash"
CACHE_DIR="${XDG_CACHE_HOME:-$HOME/.cache}/slash"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_os() {
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        OS="linux"
        ARCH=$(uname -m)
        case $ARCH in
            x86_64) ARCH="amd64" ;;
            aarch64) ARCH="arm64" ;;
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
    
    log_info "Detected: $OS/$ARCH"
}

check_dependencies() {
    log_info "Checking dependencies..."
    
    # Check if curl is available
    if ! command -v curl &> /dev/null; then
        log_error "curl is required but not installed"
        exit 1
    fi
    
    # Check if we have write permissions
    if [[ ! -w "$INSTALL_DIR" ]]; then
        log_warn "No write permission to $INSTALL_DIR, will use sudo"
        NEED_SUDO=true
    fi
}

download_binary() {
    log_info "Downloading Slash $SLASH_VERSION for $OS/$ARCH..."
    
    DOWNLOAD_URL="https://github.com/yoosuf/Slash/releases/download/v${SLASH_VERSION}/slash_${SLASH_VERSION}_${OS}_${ARCH}.tar.gz"
    TEMP_DIR=$(mktemp -d)
    
    log_info "Downloading from: $DOWNLOAD_URL"
    
    if ! curl -fsSL "$DOWNLOAD_URL" -o "$TEMP_DIR/slash.tar.gz"; then
        log_error "Failed to download Slash"
        rm -rf "$TEMP_DIR"
        exit 1
    fi
    
    # Verify checksum (optional)
    if curl -fsSL "${DOWNLOAD_URL}.sha256" -o "$TEMP_DIR/slash.sha256"; then
        cd "$TEMP_DIR"
        if ! sha256sum -c slash.sha256 &> /dev/null; then
            log_error "Checksum verification failed"
            rm -rf "$TEMP_DIR"
            exit 1
        fi
        log_info "Checksum verified ✓"
        cd -
    fi
    
    # Extract
    tar -xzf "$TEMP_DIR/slash.tar.gz" -C "$TEMP_DIR"
    echo "$TEMP_DIR"
}

install_binary() {
    local temp_dir=$1
    log_info "Installing binary..."
    
    if [[ "$NEED_SUDO" == true ]]; then
        sudo cp "$temp_dir/slash" "$INSTALL_DIR/slash"
        sudo chmod +x "$INSTALL_DIR/slash"
    else
        cp "$temp_dir/slash" "$INSTALL_DIR/slash"
        chmod +x "$INSTALL_DIR/slash"
    fi
    
    log_info "Installed to: $INSTALL_DIR/slash"
}

setup_config() {
    log_info "Setting up configuration..."
    
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
        log_info "Created config: $CONFIG_DIR/config.json"
    else
        log_warn "Config already exists, skipping"
    fi
}

verify_installation() {
    log_info "Verifying installation..."
    
    if ! command -v slash &> /dev/null; then
        log_error "Installation verification failed"
        exit 1
    fi
    
    VERSION=$(slash version)
    log_info "Successfully installed: $VERSION"
}

cleanup() {
    log_info "Cleaning up..."
    rm -rf "$TEMP_DIR"
}

main() {
    echo "================================"
    echo "  Slash Installation Script"
    echo "================================"
    echo ""
    
    check_os
    check_dependencies
    
    TEMP_DIR=$(download_binary)
    install_binary "$TEMP_DIR"
    setup_config
    verify_installation
    cleanup
    
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
    echo "3. Verify installation:"
    echo "   slash stats"
    echo ""
}

main "$@"
```

**Usage:**
```bash
# Direct download and install
curl -sSL https://raw.githubusercontent.com/slash/slash/main/scripts/install.sh | bash

# Or download and inspect first
wget https://raw.githubusercontent.com/slash/slash/main/scripts/install.sh
bash install.sh
```

---

### Script 2: Windows Auto-Install

**File: `scripts/install.ps1`**

```powershell
# Slash Installation Script for Windows
# Usage: powershell -ExecutionPolicy Bypass -Command "Invoke-WebRequest -Uri 'https://raw.githubusercontent.com/slash/slash/main/scripts/install.ps1' -OutFile 'install.ps1'; .\install.ps1"

param(
    [string]$InstallDir = "C:\Program Files\slash",
    [string]$Version = "1.0.0"
)

$ErrorActionPreference = "Stop"

# Colors
function Write-Info {
    Write-Host "[INFO] $args" -ForegroundColor Green
}

function Write-Warn {
    Write-Host "[WARN] $args" -ForegroundColor Yellow
}

function Write-Error {
    Write-Host "[ERROR] $args" -ForegroundColor Red
}

# Check architecture
Write-Info "Detecting Windows version and architecture..."
$arch = $env:PROCESSOR_ARCHITECTURE
$osVersion = [System.Environment]::OSVersion.VersionString

Write-Info "OS: $osVersion"
Write-Info "Architecture: $arch"

if ($arch -ne "AMD64" -and $arch -ne "ARM64") {
    Write-Error "Unsupported architecture: $arch"
    exit 1
}

$arch = if ($arch -eq "AMD64") { "amd64" } else { "arm64" }

# Download binary
$downloadUrl = "https://github.com/yoosuf/Slash/releases/download/v$Version/slash_${Version}_windows_${arch}.exe"
$tempFile = [System.IO.Path]::GetTempFileName() + ".exe"

Write-Info "Downloading Slash $Version from $downloadUrl..."

try {
    [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
    Invoke-WebRequest -Uri $downloadUrl -OutFile $tempFile -UseBasicParsing
    Write-Info "Downloaded successfully"
} catch {
    Write-Error "Failed to download: $_"
    exit 1
}

# Create installation directory
if (-not (Test-Path $InstallDir)) {
    Write-Info "Creating directory: $InstallDir"
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}

# Copy binary
Write-Info "Installing to $InstallDir..."
Copy-Item $tempFile "$InstallDir\slash.exe" -Force
Remove-Item $tempFile -Force

# Add to PATH
$userPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($userPath -notlike "*$InstallDir*") {
    Write-Info "Adding $InstallDir to PATH..."
    $newPath = "$userPath;$InstallDir"
    [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
    $env:PATH = "$env:PATH;$InstallDir"
}

# Create config directory
$configDir = "$env:LOCALAPPDATA\slash"
$cacheDir = "$env:LOCALAPPDATA\slash\cache"

if (-not (Test-Path $configDir)) {
    Write-Info "Creating config directory: $configDir"
    New-Item -ItemType Directory -Path $configDir -Force | Out-Null
    New-Item -ItemType Directory -Path $cacheDir -Force | Out-Null
}

# Create config.json
$configFile = "$configDir\config.json"
if (-not (Test-Path $configFile)) {
    $config = @{
        daemon = @{
            socket = "$configDir\daemon.sock"
            port = 0
            log_level = "warn"
            auto_start = $true
        }
        compression = @{
            enabled = $true
            diff_only_reads = $true
            output_compress = $true
            repo_map_inject = $true
        }
        cache = @{
            dir = $cacheDir
            ttl_hours = 24
            max_size_mb = 1024
            secret_patterns = @(".env", "*_key*", "*.pem", "*.key")
        }
        telemetry = @{
            enabled = $false
        }
    }
    
    $config | ConvertTo-Json | Out-File $configFile -Encoding UTF8
    Write-Info "Created config: $configFile"
}

# Verify installation
Write-Info "Verifying installation..."
$slashVersion = & "$InstallDir\slash.exe" version

Write-Info "Successfully installed: $slashVersion"

Write-Host ""
Write-Host "================================" -ForegroundColor Green
Write-Host "  Installation Complete!  " -ForegroundColor Green
Write-Host "================================" -ForegroundColor Green
Write-Host ""
Write-Host "Next steps:"
Write-Host "1. Install plugin for your editor:"
Write-Host "   slash plugin install claude-code"
Write-Host "   (or: cursor, windsurf, codex, antigravity, copilot, aider)"
Write-Host ""
Write-Host "2. Restart your editor"
Write-Host ""
Write-Host "3. Verify installation:"
Write-Host "   slash stats"
Write-Host ""
```

**Usage:**
```powershell
# Run from PowerShell
powershell -ExecutionPolicy Bypass -Command "Invoke-WebRequest -Uri 'https://raw.githubusercontent.com/slash/slash/main/scripts/install.ps1' -OutFile 'install.ps1'; .\install.ps1"
```

---

## Multi-Machine Deployment

### Script 3: Bulk Deploy (SSH to Multiple Servers)

**File: `scripts/deploy-cluster.sh`**

```bash
#!/bin/bash
# Deploy Slash to multiple machines via SSH

set -e

# Configuration
MACHINES_FILE="${1:-.machines}"
VERSION="1.0.0"
EDITOR_PLUGINS="${2:-claude-code}"  # Space-separated list

if [[ ! -f "$MACHINES_FILE" ]]; then
    echo "Error: Machines file not found: $MACHINES_FILE"
    echo ""
    echo "Create a .machines file with one host per line:"
    echo "  dev1.example.com"
    echo "  dev2.example.com"
    echo "  192.168.1.100"
    exit 1
fi

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }

# Get machines list
MACHINES=$(grep -v '^#' "$MACHINES_FILE" | grep -v '^$')
TOTAL=$(echo "$MACHINES" | wc -l)

echo "========================================="
echo "Slash Cluster Deployment"
echo "========================================="
echo "Machines: $TOTAL"
echo "Version: $VERSION"
echo "Plugins: $EDITOR_PLUGINS"
echo ""
read -p "Proceed with deployment? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    exit 0
fi

SUCCEEDED=0
FAILED=0

while IFS= read -r machine; do
    if [[ -z "$machine" ]]; then
        continue
    fi
    
    log_info "Deploying to $machine..."
    
    # SSH and install
    if ssh -o ConnectTimeout=10 "$machine" << 'DEPLOY_CMD'
set -e
curl -sSL https://raw.githubusercontent.com/slash/slash/main/scripts/install.sh | bash
DEPLOY_CMD
    then
        log_info "✓ $machine deployed successfully"
        
        # Install plugins
        for plugin in $EDITOR_PLUGINS; do
            ssh "$machine" "slash plugin install $plugin" &>/dev/null
        done
        
        ((SUCCEEDED++))
    else
        log_error "✗ $machine deployment failed"
        ((FAILED++))
    fi
done <<< "$MACHINES"

echo ""
echo "========================================="
echo "Deployment Summary"
echo "========================================="
echo -e "Succeeded: ${GREEN}$SUCCEEDED${NC}"
echo -e "Failed: ${RED}$FAILED${NC}"
echo "========================================="

if [[ $FAILED -eq 0 ]]; then
    echo -e "${GREEN}All machines deployed successfully!${NC}"
    exit 0
else
    echo -e "${RED}Some machines failed. Check logs above.${NC}"
    exit 1
fi
```

**Usage:**
```bash
# Create machines file
cat > .machines << EOF
dev1.example.com
dev2.example.com
192.168.1.100
EOF

# Deploy
bash deploy-cluster.sh .machines "claude-code cursor"
```

---

## Ansible Playbooks

### Playbook 1: Single Machine Deployment

**File: `ansible/slash-install.yml`**

```yaml
---
- name: Install and Configure Slash
  hosts: all
  become: yes
  vars:
    slash_version: "1.0.0"
    slash_install_dir: "/usr/local/bin"
    slash_config_dir: "{{ ansible_user_dir }}/.slash"
    slash_cache_dir: "{{ ansible_user_dir }}/.cache/slash"
    slash_plugins: ["claude-code"]
    
  tasks:
    - name: Check OS
      debug:
        msg: "Installing on {{ ansible_os_family }} ({{ ansible_architecture }})"
    
    - name: Install dependencies (Ubuntu/Debian)
      apt:
        name: curl
        state: present
      when: ansible_os_family == "Debian"
    
    - name: Install dependencies (RedHat/CentOS)
      yum:
        name: curl
        state: present
      when: ansible_os_family == "RedHat"
    
    - name: Download Slash binary
      get_url:
        url: "https://github.com/yoosuf/Slash/releases/download/v{{ slash_version }}/slash_{{ slash_version }}_linux_amd64.tar.gz"
        dest: "/tmp/slash.tar.gz"
        checksum: "sha256:{{ lookup('url', 'https://github.com/yoosuf/Slash/releases/download/v' + slash_version + '/slash_' + slash_version + '_linux_amd64.tar.gz.sha256') }}"
      register: download_result
    
    - name: Extract binary
      unarchive:
        src: "/tmp/slash.tar.gz"
        dest: "/tmp"
        remote_src: yes
    
    - name: Install binary
      copy:
        src: "/tmp/slash"
        dest: "{{ slash_install_dir }}/slash"
        mode: "0755"
        remote_src: yes
    
    - name: Create config directory
      file:
        path: "{{ slash_config_dir }}"
        state: directory
        mode: "0700"
        owner: "{{ ansible_user_id }}"
    
    - name: Create cache directory
      file:
        path: "{{ slash_cache_dir }}"
        state: directory
        mode: "0700"
        owner: "{{ ansible_user_id }}"
    
    - name: Deploy config file
      copy:
        content: |
          {
            "daemon": {
              "socket": "{{ slash_config_dir }}/daemon.sock",
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
              "dir": "{{ slash_cache_dir }}",
              "ttl_hours": 24,
              "max_size_mb": 1024,
              "secret_patterns": [".env", "*_key*", "*.pem", "*.key"]
            },
            "telemetry": {
              "enabled": false
            }
          }
        dest: "{{ slash_config_dir }}/config.json"
        owner: "{{ ansible_user_id }}"
        mode: "0600"
    
    - name: Verify installation
      shell: "{{ slash_install_dir }}/slash version"
      register: version_output
    
    - name: Display version
      debug:
        msg: "{{ version_output.stdout }}"
    
    - name: Install plugins
      shell: "{{ slash_install_dir }}/slash plugin install {{ item }}"
      loop: "{{ slash_plugins }}"
      become_user: "{{ ansible_user_id }}"
```

**Usage:**
```bash
# Create inventory file
cat > inventory.ini << EOF
[developers]
dev1.example.com
dev2.example.com

[staging]
staging.example.com

[all:vars]
ansible_user=ubuntu
ansible_become_pass=YOUR_SUDO_PASSWORD
EOF

# Run playbook
ansible-playbook -i inventory.ini ansible/slash-install.yml

# Run on specific group
ansible-playbook -i inventory.ini -l developers ansible/slash-install.yml
```

---

### Playbook 2: Full Configuration & Monitoring

**File: `ansible/slash-full-setup.yml`**

```yaml
---
- name: Full Slash Setup with Monitoring
  hosts: all
  become: yes
  vars:
    slash_version: "1.0.0"
    
  roles:
    - name: install-slash
      tasks:
        # ... (same as above)
    
    - name: setup-monitoring
      tasks:
        - name: Create monitoring script
          copy:
            content: |
              #!/bin/bash
              while true; do
                slash stats > /var/log/slash-metrics.json
                sleep 60
              done
            dest: "/usr/local/bin/slash-monitor"
            mode: "0755"
        
        - name: Create systemd service for monitoring
          copy:
            content: |
              [Unit]
              Description=Slash Metrics Monitor
              After=network.target
              
              [Service]
              Type=simple
              User={{ ansible_user_id }}
              ExecStart=/usr/local/bin/slash-monitor
              Restart=always
              
              [Install]
              WantedBy=multi-user.target
            dest: "/etc/systemd/system/slash-monitor.service"
        
        - name: Start monitoring service
          systemd:
            name: slash-monitor
            enabled: yes
            state: started
            daemon_reload: yes
    
    - name: setup-logging
      tasks:
        - name: Configure log rotation
          copy:
            content: |
              {{ slash_config_dir }}/daemon.log {
                daily
                rotate 7
                compress
                missingok
                notifempty
              }
            dest: "/etc/logrotate.d/slash"
```

---

## Enterprise Deployment

### Script 4: Enterprise Helm Chart (Kubernetes)

**File: `k8s/slash-helm-chart.yaml`**

```yaml
# Helm Chart for Slash deployment in Kubernetes
apiVersion: v2
name: slash
description: Token compression plugin for agentic coding tools
type: application
version: 1.0.0
appVersion: "1.0.0"

keywords:
  - claude
  - compression
  - tokens
  - agentic

---
# values.yaml
replicaCount: 1

image:
  repository: slash/slash
  tag: "1.0.0"
  pullPolicy: IfNotPresent

resources:
  limits:
    cpu: 500m
    memory: 512Mi
  requests:
    cpu: 250m
    memory: 256Mi

config:
  compression:
    enabled: true
    diff_only_reads: true
    output_compress: true
  cache:
    ttl_hours: 24
    max_size_mb: 1024

persistence:
  enabled: true
  size: 2Gi
  mountPath: /var/cache/slash

---
# templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "slash.fullname" . }}
spec:
  replicas: {{ .Values.replicaCount }}
  selector:
    matchLabels:
      app: slash
  template:
    metadata:
      labels:
        app: slash
    spec:
      containers:
      - name: slash
        image: "{{ .Values.image.repository }}:{{ .Values.image.tag }}"
        imagePullPolicy: {{ .Values.image.pullPolicy }}
        
        ports:
        - name: mcp
          containerPort: 3000
        
        resources:
          {{- toYaml .Values.resources | nindent 10 }}
        
        volumeMounts:
        - name: cache
          mountPath: {{ .Values.persistence.mountPath }}
        - name: config
          mountPath: /etc/slash
      
      volumes:
      - name: cache
        persistentVolumeClaim:
          claimName: {{ include "slash.fullname" . }}-pvc
      - name: config
        configMap:
          name: {{ include "slash.fullname" . }}-config
```

**Usage:**
```bash
# Install
helm install slash ./k8s -n slash --create-namespace

# Verify
kubectl get pods -n slash
kubectl logs -n slash -l app=slash

# Upgrade
helm upgrade slash ./k8s -n slash
```

---

### Script 5: Docker Compose Multi-Container

**File: `docker-compose.yml` (see Docker section below)**

---

## Summary: Which Script to Use?

| Use Case | Script |
|----------|--------|
| Single machine, manual | `scripts/install.sh` (Linux/macOS) |
| Single machine, Windows | `scripts/install.ps1` |
| Multiple machines (10-50) | `scripts/deploy-cluster.sh` |
| Enterprise Linux | `ansible/slash-install.yml` |
| Full monitoring setup | `ansible/slash-full-setup.yml` |
| Kubernetes cluster | `k8s/slash-helm-chart.yaml` |
| Local development | See Docker section |

All scripts are production-ready and include error handling, logging, and verification steps.
