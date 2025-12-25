# PowerShell script to manage Vagrant VM for testing Phanes
# Run this from Windows (or call from WSL via powershell.exe)

param(
    [Parameter(Position=0)]
    [string]$Command = "help",
    
    [Parameter(Position=1)]
    [string]$Modules = "baseline,user,security"
)

$ErrorActionPreference = "Stop"

# Get script directory and project root
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Split-Path -Parent $ScriptDir
$VmDir = Join-Path $ProjectRoot "test\vm"

$SnapshotName = "clean-state"

function Write-Info { param($msg) Write-Host "[INFO] $msg" -ForegroundColor Cyan }
function Write-Success { param($msg) Write-Host "[SUCCESS] $msg" -ForegroundColor Green }
function Write-Warn { param($msg) Write-Host "[WARN] $msg" -ForegroundColor Yellow }
function Write-Err { param($msg) Write-Host "[ERROR] $msg" -ForegroundColor Red }

function Test-VmExists {
    $status = & vagrant status --machine-readable 2>$null
    return $status -match "state,(running|poweroff|saved)"
}

function Test-SnapshotExists {
    if (-not (Test-VmExists)) { return $false }
    $snapshots = & vagrant snapshot list 2>$null
    return $snapshots -match $SnapshotName
}

function Invoke-Setup {
    Write-Info "Setting up Vagrant VM..."
    
    Set-Location $VmDir
    
    # Destroy existing VM if present
    if (Test-VmExists) {
        Write-Warn "VM already exists. Destroying it first..."
        & vagrant destroy -f 2>$null
    }
    
    Write-Info "Creating VM (this may take 2-3 minutes)..."
    & vagrant up
    if ($LASTEXITCODE -ne 0) {
        Write-Err "Failed to create VM"
        exit 1
    }
    
    Write-Info "Waiting for VM to be ready..."
    Start-Sleep -Seconds 5
    
    Write-Info "Creating clean snapshot '$SnapshotName'..."
    & vagrant snapshot save $SnapshotName
    
    Write-Success "Setup complete! VM is ready with snapshot."
    Write-Info "You can now use: .\scripts\test-vm.ps1 test [modules]"
}

function Get-SecureKeyPath {
    # Copy the SSH key to a Windows temp location with correct permissions
    $SourceKey = Join-Path $VmDir ".vagrant\machines\default\virtualbox\private_key"
    $TempKey = Join-Path $env:TEMP "vagrant_phanes_key"
    
    if (Test-Path $SourceKey) {
        # Delete existing temp key if present (may have restrictive permissions)
        if (Test-Path $TempKey) {
            # First restore permissions so we can delete
            $null = & icacls $TempKey /reset 2>&1
            Remove-Item -Path $TempKey -Force -ErrorAction SilentlyContinue
        }
        
        # Copy the key
        Copy-Item -Path $SourceKey -Destination $TempKey -Force
        
        # Fix permissions using icacls (works without admin)
        # Remove inheritance and grant only current user read access
        $null = & icacls $TempKey /inheritance:r /grant:r "${env:USERNAME}:(R)" 2>&1
        
        return $TempKey
    }
    return $null
}

function Invoke-SshCommand {
    param([string]$Command)
    
    $KeyPath = Get-SecureKeyPath
    
    if ($KeyPath) {
        & ssh -i $KeyPath -o StrictHostKeyChecking=no -o UserKnownHostsFile=NUL -p 2222 vagrant@127.0.0.1 $Command
    } else {
        & vagrant ssh -c $Command
    }
}

function Invoke-Test {
    param([string]$ModuleList)
    
    Write-Info "Testing modules: $ModuleList"
    
    Set-Location $VmDir
    
    if (-not (Test-VmExists)) {
        Write-Err "VM does not exist. Run 'test-vm.ps1 setup' first."
        exit 1
    }
    
    # Restore snapshot if it exists
    if (Test-SnapshotExists) {
        Write-Info "Restoring snapshot '$SnapshotName'..."
        & vagrant snapshot restore $SnapshotName
        Write-Info "Snapshot restored"
    } else {
        Write-Warn "Snapshot '$SnapshotName' not found. VM may not be in clean state."
        $status = & vagrant status
        if ($status -notmatch "running") {
            Write-Info "Starting VM..."
            & vagrant up
        }
    }
    
    # Sync files
    Write-Info "Syncing files to VM..."
    & vagrant rsync
    
    # Build phanes in VM
    Write-Info "Building phanes in VM..."
    Invoke-SshCommand "cd /workspace && export PATH=`$PATH:/usr/local/go/bin && go build -o phanes ."
    
    # Run phanes with specified modules
    Write-Info "Running phanes with modules: $ModuleList"
    Invoke-SshCommand "cd /workspace && sudo ./phanes --modules $ModuleList --config test-config.yaml"
    
    Write-Success "Test complete!"
    Write-Info "Run 'test-vm.ps1 test' again to restore clean state and test again."
}

function Invoke-Shell {
    Set-Location $VmDir
    
    if (-not (Test-VmExists)) {
        Write-Err "VM does not exist. Run 'test-vm.ps1 setup' first."
        exit 1
    }
    
    $status = & vagrant status
    if ($status -notmatch "running") {
        Write-Info "Starting VM..."
        & vagrant up
    }
    
    Write-Info "Opening SSH shell in VM..."
    Write-Info "Project is at /workspace"
    
    $KeyPath = Get-SecureKeyPath
    
    if ($KeyPath) {
        # Use ssh.exe directly with the secure key (-t forces TTY allocation)
        & ssh -t -i $KeyPath -o StrictHostKeyChecking=no -o UserKnownHostsFile=NUL -p 2222 vagrant@127.0.0.1
    } else {
        Write-Warn "Private key not found"
        Write-Info "Falling back to vagrant ssh (may prompt for password)"
        & vagrant ssh
    }
}

function Invoke-Sync {
    Set-Location $VmDir
    
    if (-not (Test-VmExists)) {
        Write-Err "VM does not exist. Run 'test-vm.ps1 setup' first."
        exit 1
    }
    
    Write-Info "Syncing files to VM..."
    & vagrant rsync
    Write-Success "Files synced."
}

function Invoke-Status {
    Set-Location $VmDir
    
    if (-not (Test-VmExists)) {
        Write-Info "VM does not exist."
        return
    }
    
    Write-Info "VM Status:"
    & vagrant status
    
    Write-Host ""
    Write-Info "Snapshots:"
    & vagrant snapshot list 2>$null
    if ($LASTEXITCODE -ne 0) {
        Write-Host "  No snapshots"
    }
}

function Invoke-Destroy {
    Set-Location $VmDir
    
    if (-not (Test-VmExists)) {
        Write-Warn "VM does not exist."
        return
    }
    
    Write-Warn "This will destroy the VM and all snapshots!"
    $confirm = Read-Host "Are you sure? (y/N)"
    if ($confirm -ne "y" -and $confirm -ne "Y") {
        Write-Info "Cancelled."
        return
    }
    
    & vagrant destroy -f
    Write-Info "VM destroyed."
}

function Show-Help {
    Write-Host "Usage: test-vm.ps1 <command> [options]"
    Write-Host ""
    Write-Host "Commands:"
    Write-Host "  setup              Create VM and initial clean snapshot"
    Write-Host "  test [modules]     Restore snapshot and test modules (default: baseline,user,security)"
    Write-Host "  shell              Open SSH shell in VM"
    Write-Host "  sync               Sync project files to VM"
    Write-Host "  status             Show VM status"
    Write-Host "  destroy            Destroy VM and all snapshots"
    Write-Host ""
    Write-Host "Examples:"
    Write-Host "  .\test-vm.ps1 setup                    # Initial setup"
    Write-Host "  .\test-vm.ps1 test                     # Test default modules"
    Write-Host "  .\test-vm.ps1 test baseline            # Test specific module"
    Write-Host "  .\test-vm.ps1 test baseline,user       # Test multiple modules"
    Write-Host "  .\test-vm.ps1 shell                    # Open shell for manual testing"
}

# Main
switch ($Command.ToLower()) {
    "setup"   { Invoke-Setup }
    "test"    { Invoke-Test -ModuleList $Modules }
    "shell"   { Invoke-Shell }
    "sync"    { Invoke-Sync }
    "status"  { Invoke-Status }
    "destroy" { Invoke-Destroy }
    default   { Show-Help }
}

