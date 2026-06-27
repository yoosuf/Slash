#!/bin/bash
# Deploy Slash to multiple machines via SSH
# Usage: bash deploy-cluster.sh .machines "claude-code cursor"

set -e

MACHINES_FILE="${1:-.machines}"
VERSION="${VERSION:-1.0.0}"
EDITOR_PLUGINS="${2:-claude-code}"
SSH_USER="${SSH_USER:-}"
SSH_KEY="${SSH_KEY:-}"
SSH_OPTS=""

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

log_error() {
    echo -e "${RED}[✗]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[!]${NC} $1"
}

log_step() {
    echo -e "${BLUE}==>${NC} $1"
}

show_usage() {
    cat << EOF
Slash Cluster Deployment Script

Usage:
  bash deploy-cluster.sh <machines-file> [plugins] [options]

Arguments:
  <machines-file>    File with one hostname per line
  [plugins]          Space-separated editor plugins (default: claude-code)

Environment Variables:
  SSH_USER          SSH username (default: current user)
  SSH_KEY           SSH key file (default: ~/.ssh/id_rsa)
  VERSION           Slash version to install (default: 1.0.0)

Examples:
  bash deploy-cluster.sh .machines
  bash deploy-cluster.sh .machines "claude-code cursor"
  SSH_USER=ubuntu SSH_KEY=~/.ssh/aws.pem bash deploy-cluster.sh machines.txt

Machine File Format:
  # Comments are ignored
  dev1.example.com
  dev2.example.com
  192.168.1.100

  # With custom SSH port
  dev3.example.com:2222
EOF
}

if [[ "$1" == "-h" ]] || [[ "$1" == "--help" ]]; then
    show_usage
    exit 0
fi

if [[ ! -f "$MACHINES_FILE" ]]; then
    log_error "Machines file not found: $MACHINES_FILE"
    echo ""
    show_usage
    exit 1
fi

# Setup SSH options
if [[ -n "$SSH_KEY" ]]; then
    SSH_OPTS="-i $SSH_KEY"
fi

if [[ -n "$SSH_USER" ]]; then
    SSH_OPTS="$SSH_OPTS -l $SSH_USER"
fi

# Read machines
MACHINES=($(grep -v '^#' "$MACHINES_FILE" | grep -v '^$'))
TOTAL=${#MACHINES[@]}

if [[ $TOTAL -eq 0 ]]; then
    log_error "No machines found in $MACHINES_FILE"
    exit 1
fi

# Confirmation
clear
echo "========================================="
echo "Slash Cluster Deployment"
echo "========================================="
echo "Machines file: $MACHINES_FILE"
echo "Total machines: $TOTAL"
echo "Version: $VERSION"
echo "Plugins: $EDITOR_PLUGINS"
echo "SSH options: $SSH_OPTS"
echo "========================================="
echo ""
echo "Machines to deploy:"
for machine in "${MACHINES[@]}"; do
    echo "  - $machine"
done
echo ""
read -p "Proceed with deployment? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    log_warn "Deployment cancelled"
    exit 0
fi

# Deployment
SUCCEEDED=0
FAILED=0
FAILED_MACHINES=()

log_step "Starting deployment to $TOTAL machines..."
echo ""

for machine in "${MACHINES[@]}"; do
    echo -ne "${BLUE}==>${NC} Deploying to $machine... "

    if ssh -o ConnectTimeout=10 -o BatchMode=yes $SSH_OPTS "$machine" << 'DEPLOY_CMD' >/dev/null 2>&1
#!/bin/bash
set -e
curl -sSL https://raw.githubusercontent.com/slash/slash/main/scripts/install.sh | bash >/dev/null 2>&1
DEPLOY_CMD
    then
        echo -e "${GREEN}✓${NC}"
        log_info "$machine deployed successfully"

        # Install plugins
        for plugin in $EDITOR_PLUGINS; do
            ssh -o BatchMode=yes $SSH_OPTS "$machine" "slash plugin install $plugin" >/dev/null 2>&1
        done

        ((SUCCEEDED++))
    else
        echo -e "${RED}✗${NC}"
        log_error "$machine deployment failed"
        FAILED_MACHINES+=("$machine")
        ((FAILED++))
    fi
done

# Summary
echo ""
echo "========================================="
echo "Deployment Summary"
echo "========================================="
echo -e "Total: $TOTAL machines"
echo -e "Succeeded: ${GREEN}$SUCCEEDED${NC}"
echo -e "Failed: ${RED}$FAILED${NC}"
echo "========================================="

if [[ $FAILED -gt 0 ]]; then
    echo ""
    log_warn "Failed machines:"
    for machine in "${FAILED_MACHINES[@]}"; do
        echo "  - $machine"
    done
    echo ""
    log_warn "Troubleshooting tips:"
    echo "  1. Check SSH connectivity: ssh -i KEY $machine"
    echo "  2. Verify SSH key permissions: chmod 600 KEY"
    echo "  3. Check firewall/security groups"
    echo "  4. Try manual install:"
    echo "     ssh -i KEY $machine 'curl -sSL https://raw.githubusercontent.com/slash/slash/main/scripts/install.sh | bash'"
fi

echo ""
if [[ $FAILED -eq 0 ]]; then
    log_info "All machines deployed successfully!"
    exit 0
else
    log_error "Some machines failed"
    exit 1
fi
