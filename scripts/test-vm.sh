#!/bin/bash
# Script to manage Vagrant VM for testing Phanes modules
# Simplified for WSL compatibility - uses manual file copy instead of rsync

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
VM_DIR="$PROJECT_ROOT/test/vm"
SNAPSHOT_NAME="clean-state"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Detect WSL and set correct commands
VAGRANT_CMD="vagrant"
if [ -d "/mnt/c" ] && [ ! -f "/.dockerenv" ]; then
    log_info "Running in WSL - using .exe commands"
    if command -v vagrant.exe &> /dev/null; then
        VAGRANT_CMD="vagrant.exe"
    fi
fi

cd "$VM_DIR"

# Check if Vagrant is available
check_vagrant() {
    if ! command -v "$VAGRANT_CMD" &> /dev/null; then
        log_error "Vagrant not found. Please install Vagrant."
        exit 1
    fi
}

# Check if VM exists
vm_exists() {
    "$VAGRANT_CMD" status --machine-readable 2>/dev/null | grep -q "state,running\|state,poweroff\|state,saved" || false
}

# Check if snapshot exists
snapshot_exists() {
    if ! vm_exists; then
        return 1
    fi
    "$VAGRANT_CMD" snapshot list 2>/dev/null | grep -q "$SNAPSHOT_NAME" || false
}

# Copy project files to VM using scp
sync_files() {
    log_info "Copying project files to VM..."
    
    # Get SSH config
    local ssh_port=2222
    local ssh_host="127.0.0.1"
    local ssh_user="vagrant"
    local ssh_pass="vagrant"
    
    # Use sshpass if available, otherwise prompt for password
    if command -v sshpass &> /dev/null; then
        # Create tarball excluding unnecessary files
        cd "$PROJECT_ROOT"
        tar czf /tmp/phanes.tar.gz \
            --exclude='.git' \
            --exclude='test/vm/.vagrant' \
            --exclude='.vagrant' \
            --exclude='*.log' \
            --exclude='phanes' \
            .
        
        # Copy and extract
        sshpass -p "$ssh_pass" scp -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -P "$ssh_port" \
            /tmp/phanes.tar.gz "$ssh_user@$ssh_host:/tmp/"
        
        sshpass -p "$ssh_pass" ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -p "$ssh_port" \
            "$ssh_user@$ssh_host" "cd /workspace && sudo tar xzf /tmp/phanes.tar.gz && sudo chown -R vagrant:vagrant /workspace"
        
        rm /tmp/phanes.tar.gz
        cd "$VM_DIR"
        log_success "Files synced to VM"
    else
        log_warn "sshpass not found. Installing it for automated file sync..."
        sudo apt-get install -y sshpass 2>/dev/null || {
            log_error "Could not install sshpass. Please install it manually: sudo apt-get install sshpass"
            exit 1
        }
        sync_files  # Retry after installing
    fi
}

# Run command in VM
vm_run() {
    local cmd="$1"
    if command -v sshpass &> /dev/null; then
        sshpass -p "vagrant" ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -p 2222 \
            vagrant@127.0.0.1 "$cmd"
    else
        "$VAGRANT_CMD" ssh -c "$cmd"
    fi
}

# Setup VM and create initial snapshot
setup() {
    log_info "Setting up Vagrant VM..."
    check_vagrant
    
    # Destroy existing VM if present
    if vm_exists; then
        log_warn "VM already exists. Destroying it first..."
        "$VAGRANT_CMD" destroy -f 2>/dev/null || true
    fi
    
    log_info "Creating VM (this may take 2-3 minutes)..."
    "$VAGRANT_CMD" up
    
    log_info "Waiting for VM to be fully ready..."
    sleep 5
    
    # Sync files to VM
    sync_files
    
    log_info "Creating clean snapshot '$SNAPSHOT_NAME'..."
    "$VAGRANT_CMD" snapshot save "$SNAPSHOT_NAME"
    
    log_success "Setup complete! VM is ready with snapshot."
    log_info "You can now use: ./scripts/test-vm.sh test [modules]"
}

# Restore snapshot and run tests
test_modules() {
    local modules="${1:-baseline,user,security}"
    
    log_info "Testing modules: $modules"
    
    if ! vm_exists; then
        log_error "VM does not exist. Run './scripts/test-vm.sh setup' first."
        exit 1
    fi
    
    # Restore snapshot if it exists
    if snapshot_exists; then
        log_info "Restoring snapshot '$SNAPSHOT_NAME'..."
        "$VAGRANT_CMD" snapshot restore "$SNAPSHOT_NAME"
        log_info "Snapshot restored"
    else
        log_warn "Snapshot '$SNAPSHOT_NAME' not found. VM may not be in clean state."
        if ! "$VAGRANT_CMD" status | grep -q "running"; then
            log_info "Starting VM..."
            "$VAGRANT_CMD" up
        fi
    fi
    
    # Always sync latest files
    sync_files
    
    # Build phanes in VM
    log_info "Building phanes in VM..."
    vm_run "cd /workspace && export PATH=\$PATH:/usr/local/go/bin && go build -o phanes ."
    
    # Run phanes with specified modules
    log_info "Running phanes with modules: $modules"
    vm_run "cd /workspace && sudo ./phanes --modules $modules --config test-config.yaml"
    
    log_success "Test complete!"
    log_info "Run './scripts/test-vm.sh test' again to restore clean state and test again."
}

# Open shell in VM
shell() {
    if ! vm_exists; then
        log_error "VM does not exist. Run './scripts/test-vm.sh setup' first."
        exit 1
    fi
    
    if ! "$VAGRANT_CMD" status | grep -q "running"; then
        log_info "Starting VM..."
        "$VAGRANT_CMD" up
    fi
    
    log_info "Opening SSH shell in VM..."
    log_info "Project is at /workspace"
    log_info "Password: vagrant"
    
    if command -v sshpass &> /dev/null; then
        sshpass -p "vagrant" ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -p 2222 vagrant@127.0.0.1
    else
        "$VAGRANT_CMD" ssh
    fi
}

# Destroy VM
destroy() {
    if ! vm_exists; then
        log_warn "VM does not exist."
        return 0
    fi
    
    log_warn "This will destroy the VM and all snapshots!"
    read -p "Are you sure? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "Cancelled."
        return 0
    fi
    
    "$VAGRANT_CMD" destroy -f
    log_info "VM destroyed."
}

# Show VM status
status() {
    check_vagrant
    
    if ! vm_exists; then
        log_info "VM does not exist."
        return 0
    fi
    
    log_info "VM Status:"
    "$VAGRANT_CMD" status
    
    log_info ""
    log_info "Snapshots:"
    "$VAGRANT_CMD" snapshot list 2>/dev/null || echo "  No snapshots"
}

# Show usage
usage() {
    echo "Usage: $0 <command> [options]"
    echo ""
    echo "Commands:"
    echo "  setup              Create VM and initial clean snapshot"
    echo "  test [modules]     Restore snapshot and test modules (default: baseline,user,security)"
    echo "  shell              Open SSH shell in VM"
    echo "  sync               Sync project files to VM"
    echo "  status             Show VM status"
    echo "  destroy            Destroy VM and all snapshots"
    echo ""
    echo "Examples:"
    echo "  $0 setup                    # Initial setup"
    echo "  $0 test                     # Test default modules"
    echo "  $0 test baseline            # Test specific module"
    echo "  $0 test baseline,user       # Test multiple modules"
    echo "  $0 shell                    # Open shell for manual testing"
}

# Main
case "${1:-}" in
    setup)
        setup
        ;;
    test)
        test_modules "${2:-}"
        ;;
    shell)
        shell
        ;;
    sync)
        sync_files
        ;;
    status)
        status
        ;;
    destroy)
        destroy
        ;;
    *)
        usage
        exit 1
        ;;
esac
