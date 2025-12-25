#!/bin/bash
# Wrapper script to run VM commands from Windows via PowerShell
# This allows running VM tests from WSL while using Windows-native Vagrant

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Convert WSL path to Windows path
wsl_to_windows_path() {
    local wsl_path="$1"
    # Convert /home/user/... to \\wsl.localhost\Ubuntu\home\user\...
    # Then to Windows format
    echo "$wsl_path" | sed 's|^/|\\\\wsl.localhost\\Ubuntu\\|' | sed 's|/|\\|g'
}

# Get the Windows path to the PowerShell script
PS_SCRIPT="$SCRIPT_DIR/test-vm.ps1"
WIN_PS_SCRIPT=$(wslpath -w "$PS_SCRIPT" 2>/dev/null || wsl_to_windows_path "$PS_SCRIPT")

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }

# Check if we're in WSL
if [ ! -d "/mnt/c" ]; then
    echo "This script is designed to run from WSL."
    echo "It calls PowerShell on Windows to manage the VM."
    exit 1
fi

# Check if PowerShell is available
if ! command -v powershell.exe &> /dev/null; then
    echo "PowerShell not found. Make sure you're running from WSL2."
    exit 1
fi

log_info "Running VM command via Windows PowerShell..."
log_info "Command: $1 ${2:-}"

# Run the PowerShell script with arguments
# -ExecutionPolicy Bypass allows running the script
powershell.exe -ExecutionPolicy Bypass -File "$WIN_PS_SCRIPT" "$@"
