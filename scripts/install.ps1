# Slash Installation Script for Windows
# Usage: powershell -ExecutionPolicy Bypass -Command "Invoke-WebRequest -Uri 'https://raw.githubusercontent.com/slash/slash/main/scripts/install.ps1' -OutFile 'install.ps1'; .\install.ps1"

param(
    [string]$InstallDir = "C:\Program Files\slash",
    [string]$Version = "1.0.0",
    [switch]$Help
)

$ErrorActionPreference = "Stop"

# Functions
function Write-Header {
    Clear-Host
    Write-Host "================================" -ForegroundColor Cyan
    Write-Host "  Slash Installation Script" -ForegroundColor Cyan
    Write-Host "================================" -ForegroundColor Cyan
    Write-Host ""
}

function Write-Info {
    param([string]$Message)
    Write-Host "[✓] $Message" -ForegroundColor Green
}

function Write-Warn {
    param([string]$Message)
    Write-Host "[!] $Message" -ForegroundColor Yellow
}

function Write-Step {
    param([string]$Message)
    Write-Host "==> $Message" -ForegroundColor Cyan
}

function Write-Error-Custom {
    param([string]$Message)
    Write-Host "[✗] $Message" -ForegroundColor Red
}

function Show-Help {
    Write-Host @"
Slash Installation Script for Windows

Usage:
  .\install.ps1 [Options]

Options:
  -InstallDir <path>    Installation directory (default: C:\Program Files\slash)
  -Version <version>    Version to install (default: 1.0.0)
  -Help                 Show this help message

Examples:
  .\install.ps1                              # Install with defaults
  .\install.ps1 -InstallDir "D:\slash"       # Custom install directory
  .\install.ps1 -Version "1.1.0"             # Specific version

Requirements:
  - Windows 10 or later
  - .NET Framework 4.7.2 or later
  - Administrator privileges (for installation)
"@
}

# Main installation
function Install-Slash {
    Write-Header

    if ($Help) {
        Show-Help
        return
    }

    Write-Step "Detecting Windows version and architecture..."

    # Get Windows version
    $osVersion = [System.Environment]::OSVersion.VersionString
    $arch = $env:PROCESSOR_ARCHITECTURE

    Write-Info "OS: $osVersion"
    Write-Info "Architecture: $arch"

    # Map architecture
    $archName = if ($arch -eq "AMD64") { "amd64" } else { "arm64" }

    # URLs
    $downloadUrl = "https://github.com/yoosuf/Slash/releases/download/v$Version/slash_${Version}_windows_${archName}.zip"
    $checksumUrl = "https://github.com/yoosuf/Slash/releases/download/v$Version/SHA256SUMS"

    # Temp directory
    $tempDir = [System.IO.Path]::GetTempPath() + "slash-install-" + (Get-Random)
    New-Item -ItemType Directory -Path $tempDir -Force | Out-Null

    try {
        # Download
        Write-Step "Downloading Slash v$Version..."
        Write-Info "Download URL: $downloadUrl"

        [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
        Invoke-WebRequest -Uri $downloadUrl -OutFile "$tempDir\slash.zip" -UseBasicParsing

        Write-Info "Downloaded successfully"

        # Extract
        Write-Step "Extracting binary..."
        Expand-Archive -Path "$tempDir\slash.zip" -DestinationPath $tempDir -Force
        Write-Info "Extracted"

        # Create installation directory
        Write-Step "Creating installation directory: $InstallDir"
        if (-not (Test-Path $InstallDir)) {
            New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
        }

        # Copy binary
        Write-Step "Installing binary..."
        $binaryPath = Get-ChildItem -Path $tempDir -Filter "slash.exe" -Recurse | Select-Object -First 1
        if (-not $binaryPath) {
            Write-Error-Custom "slash.exe not found in downloaded archive"
            exit 1
        }

        Copy-Item $binaryPath.FullName "$InstallDir\slash.exe" -Force
        Write-Info "Binary installed: $InstallDir\slash.exe"

        # Add to PATH
        Write-Step "Adding to PATH..."
        $userPath = [Environment]::GetEnvironmentVariable("PATH", "User")
        if ($userPath -notlike "*$InstallDir*") {
            $newPath = "$userPath;$InstallDir"
            [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
            $env:PATH = "$env:PATH;$InstallDir"
            Write-Info "Added to PATH"
        } else {
            Write-Warn "Already in PATH"
        }

        # Setup config directory
        Write-Step "Setting up configuration..."
        $configDir = "$env:LOCALAPPDATA\slash"
        $cacheDir = "$env:LOCALAPPDATA\slash\cache"

        if (-not (Test-Path $configDir)) {
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
            Write-Info "Config created: $configFile"
        } else {
            Write-Warn "Config already exists"
        }

        # Verify installation
        Write-Step "Verifying installation..."
        $slashVersion = & "$InstallDir\slash.exe" version 2>$null
        if ($LASTEXITCODE -eq 0) {
            Write-Info "Successfully installed: $slashVersion"
        } else {
            Write-Error-Custom "Failed to verify installation"
            exit 1
        }

        # Success
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
        Write-Host "3. Start using Slash:"
        Write-Host "   slash stats"
        Write-Host ""
        Write-Host "Documentation: https://github.com/yoosuf/Slash"
        Write-Host ""

    } catch {
        Write-Error-Custom "Installation failed: $_"
        exit 1
    } finally {
        # Cleanup
        if (Test-Path $tempDir) {
            Remove-Item -Path $tempDir -Recurse -Force -ErrorAction SilentlyContinue
        }
    }
}

# Run installation
Install-Slash
