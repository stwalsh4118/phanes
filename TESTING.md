# Testing Phanes in Docker Containers

This guide explains how to test Phanes safely in Docker containers without affecting your host system.

## Quick Start

### Prerequisites
- Docker installed and running
- Docker Compose (optional, but recommended)

### Using the Test Script

The easiest way to test is using the provided script:

```bash
# List available modules and profiles
./scripts/test-in-container.sh list

# Test with dry-run (safe, no changes)
./scripts/test-in-container.sh dry-run

# Execute with custom arguments (FULL EXECUTION - makes changes!)
./scripts/test-in-container.sh exec --modules baseline --config test-config.yaml

# Execute multiple modules
./scripts/test-in-container.sh exec --modules baseline,user --config test-config.yaml

# Open interactive shell for manual testing
./scripts/test-in-container.sh shell
```

### Using Docker Compose Directly

```bash
# Build the test image
docker-compose -f docker-compose.test.yml build

# List modules and profiles
docker-compose -f docker-compose.test.yml run --rm phanes-test phanes --list

# Test with dry-run (safe, no changes)
docker-compose -f docker-compose.test.yml run --rm phanes-test phanes --modules baseline,user --config test-config.yaml --dry-run

# FULL EXECUTION - Run baseline module (makes real changes!)
docker-compose -f docker-compose.test.yml run --rm phanes-test phanes --modules baseline --config test-config.yaml

# FULL EXECUTION - Run user module (creates user, SSH keys, sudoers!)
docker-compose -f docker-compose.test.yml run --rm phanes-test phanes --modules user --config test-config.yaml

# Open interactive shell
docker-compose -f docker-compose.test.yml run --rm phanes-test /bin/bash
```

### Using Docker Directly (without Compose)

```bash
# Build the image
docker build -f Dockerfile.test -t phanes-test .

# Run tests
docker run --rm -it -v "$(pwd):/workspace" -w /workspace phanes-test phanes --list

# Interactive shell
docker run --rm -it -v "$(pwd):/workspace" -w /workspace phanes-test /bin/bash
```

## Test Config File

Make sure you have a `test-config.yaml` file in the project root:

```yaml
user:
  username: "testuser"
  ssh_public_key: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC... test@example.com"
system:
  timezone: "UTC"
```

## Testing Scenarios

### 1. List Available Modules
```bash
./scripts/test-in-container.sh list
```

### 2. Dry-Run Test (Safe)
```bash
./scripts/test-in-container.sh dry-run
```

### 3. Test Specific Module (Dry-Run)
```bash
./scripts/test-in-container.sh exec --modules baseline --config test-config.yaml --dry-run
```

### 4. FULL EXECUTION - Baseline Module
```bash
# WARNING: This will actually modify the container's filesystem
./scripts/test-in-container.sh exec --modules baseline --config test-config.yaml
```

This will:
- Set timezone to UTC (writes to `/etc/timezone`)
- Configure locale `en_US.UTF-8`
- Run `apt-get update`

### 5. FULL EXECUTION - User Module
```bash
# WARNING: This will create a user, SSH keys, and sudoers file!
./scripts/test-in-container.sh exec --modules user --config test-config.yaml
```

This will:
- Create user `testuser` with home directory
- Create `~/.ssh` directory with permissions 700
- Add SSH public key to `~/.ssh/authorized_keys` with permissions 600
- Create `/etc/sudoers.d/testuser` with passwordless sudo

### 6. Verify Changes in Container
```bash
# Open interactive shell
./scripts/test-in-container.sh shell

# Then inside the container:
id testuser                    # Check if user exists
ls -la /home/testuser/.ssh/   # Check SSH directory
cat /etc/sudoers.d/testuser   # Check sudoers file
cat /etc/timezone             # Check timezone
cat /etc/default/locale       # Check locale
```

### 7. Test with Profile
```bash
# Dry-run with profile
./scripts/test-in-container.sh exec --profile minimal --config test-config.yaml --dry-run

# Note: Profiles include modules not yet implemented, so full execution will show errors
```

## Notes

- The container runs as **root** to allow system modifications
- Changes made inside the container are **isolated** and won't affect your host
- The container is **ephemeral** - changes are lost when the container stops (unless you use volumes)
- Use `--dry-run` flag for safe testing without making changes
- The test config file is mounted read-only to prevent accidental modifications
- **Full execution makes real changes** - test modules will actually be installed!

## E2E Tests

End-to-end tests verify that modules work correctly in a containerized environment. These tests:

- Run modules in Docker containers
- Verify modules actually configure the system correctly
- Test idempotency (running modules multiple times)
- Test dry-run mode
- Verify file permissions and content

### Running E2E Tests

**Using Make:**
```bash
make test-e2e
```

**Using Docker Compose directly:**
```bash
docker-compose -f docker-compose.test.yml build
docker-compose -f docker-compose.test.yml run --rm phanes-test go test -v ./test/integration/... -run "E2E"
```

**Using the test script:**
```bash
./scripts/test-in-container.sh e2e
```

### E2E Test Coverage

The E2E tests cover:

1. **Baseline Module**:
   - Dry-run mode
   - Full execution
   - Idempotency
   - System verification (timezone, locale, apt update)

2. **User Module**:
   - Dry-run mode
   - Full execution
   - Idempotency
   - System verification (user creation, SSH keys, sudoers)

3. **Multiple Modules**:
   - Running multiple modules together
   - Dry-run with multiple modules

### Test Structure

E2E tests are located in `test/integration/modules_e2e_test.go`. They:

- Automatically detect if running in a containerized environment
- Skip if not in a container (to avoid modifying host system)
- Can be run with `-short` flag to skip E2E tests
- Verify actual system state after module execution

### Writing New E2E Tests

When adding new modules, add corresponding E2E tests:

1. Add test cases to `test/integration/modules_e2e_test.go`
2. Test both dry-run and full execution
3. Test idempotency
4. Verify system state after execution
5. Use the `isContainerized()` helper to ensure tests only run in containers

## Troubleshooting

### Container won't start
- Make sure Docker is running: `docker ps`
- Check Docker Compose version: `docker-compose --version`

### Build fails
- Ensure you're in the project root directory
- Check that `go.mod` exists

### Permission errors
- The container runs as root, so this shouldn't happen
- If you see permission errors, check Docker daemon permissions

### Module execution fails
- Some modules require specific system tools (e.g., `useradd`, `visudo`)
- The test container includes these tools, but if a module fails, check its dependencies
- Baseline module handles Docker containers (no systemd) gracefully
- User module requires `useradd` and `visudo` (both included in container)

### Locale not active in shell
- In Docker containers, locale changes may not be active in the current shell session
- The locale is still configured correctly in `/etc/default/locale`
- New shells will use the configured locale

## Testing in Vagrant VM (Recommended for System Modules)

For testing modules that require full system capabilities (UFW, fail2ban, systemd), use a Vagrant VM. This provides a real Ubuntu environment with instant rollback via snapshots.

### Prerequisites

- **VirtualBox** installed ([Download](https://www.virtualbox.org/wiki/Downloads))
- **Vagrant** installed ([Download](https://www.vagrantup.com/downloads))
- ~4GB disk space for VM

### Quick Start

```bash
# One-time setup (creates VM + snapshot, ~3 minutes)
./scripts/test-vm.sh setup

# Run security module test (restore + test, ~10 seconds)
./scripts/test-vm.sh test security

# Run all modules test
./scripts/test-vm.sh test baseline,user,security

# Interactive debugging
./scripts/test-vm.sh shell

# Check VM status
./scripts/test-vm.sh status

# Clean up when done
./scripts/test-vm.sh destroy
```

### How It Works

The VM testing uses **snapshots** for fast iteration:

1. **Initial Setup**: Creates Ubuntu 22.04 VM and takes a "clean" snapshot
2. **Each Test Run**: 
   - Restores snapshot (~2-5 seconds)
   - Builds phanes binary
   - Runs specified modules
   - VM is now "dirty" but you don't care
3. **Next Test**: Restore snapshot again (instant clean state)

### Commands

| Command | Description |
|---------|-------------|
| `setup` | Create VM and take initial snapshot (one-time, ~3 min) |
| `test [modules]` | Restore snapshot and test modules (default: baseline,user,security) |
| `shell` | Open SSH shell in VM for manual testing |
| `status` | Show VM and snapshot status |
| `destroy` | Destroy VM (with confirmation) |

### Testing Workflow

```bash
# 1. One-time setup
./scripts/test-vm.sh setup

# 2. Test security module (restores clean state each time)
./scripts/test-vm.sh test security

# 3. Test multiple modules
./scripts/test-vm.sh test baseline,user,security

# 4. Test with dry-run
./scripts/test-vm.sh shell
# Inside VM:
cd /workspace
sudo ./phanes --modules security --config test-config.yaml --dry-run

# 5. Test with profile
./scripts/test-vm.sh shell
# Inside VM:
cd /workspace
sudo ./phanes --profile minimal --config test-config.yaml
```

### VM Configuration

- **OS**: Ubuntu 22.04 LTS (jammy)
- **Resources**: 2GB RAM, 2 CPUs
- **Synced Folder**: Project root → `/workspace` in VM
- **SSH Port**: Forwarded to host port 2222
- **Pre-installed**: Go 1.21, UFW, fail2ban, SSH server, locales

### Advantages Over Docker

- **Full system capabilities**: UFW, fail2ban, systemd all work properly
- **Fast iteration**: Snapshot restore takes 2-5 seconds vs 3+ minutes for VM recreation
- **Real environment**: Matches production VPS setup exactly
- **Persistent**: VM stays running between tests (faster than Docker container recreation)

### Time Estimates

| Operation | Time |
|-----------|------|
| Initial VM creation | ~3 minutes (one-time) |
| Snapshot restore | 2-5 seconds |
| Build phanes in VM | ~5 seconds |
| Run single module | ~10-30 seconds |
| **Total per test cycle** | **~20-40 seconds** |

### Troubleshooting

#### VM won't start
- Check VirtualBox is running: `VBoxManage --version`
- Check Vagrant is installed: `vagrant --version`
- Try: `cd test/vm && vagrant up`

#### Snapshot restore fails
- Check snapshot exists: `cd test/vm && vagrant snapshot list`
- If missing, recreate: `./scripts/test-vm.sh setup`

#### Synced folder issues
- Ensure VirtualBox Guest Additions are installed (should be automatic)
- Try: `cd test/vm && vagrant reload`

#### Port conflicts
- If port 2222 is in use, edit `test/vm/Vagrantfile` and change the forwarded port

#### Out of disk space
- Destroy old VMs: `./scripts/test-vm.sh destroy`
- Clean Vagrant cache: `vagrant box prune`

### Manual VM Management

If you need to manage the VM manually:

```bash
cd test/vm

# Start VM
vagrant up

# Stop VM
vagrant halt

# Suspend VM
vagrant suspend

# SSH into VM
vagrant ssh

# List snapshots
vagrant snapshot list

# Create snapshot
vagrant snapshot save my-snapshot

# Restore snapshot
vagrant snapshot restore my-snapshot

# Delete snapshot
vagrant snapshot delete my-snapshot

# Destroy VM
vagrant destroy
```

### When to Use VM vs Docker

**Use VM for:**
- Testing modules that require systemd (fail2ban, SSH service management)
- Testing UFW firewall configuration
- Testing modules that modify system services
- Full end-to-end testing before deployment

**Use Docker for:**
- Quick unit/integration tests that don't need full system
- CI/CD pipelines (faster startup)
- Testing modules that work in containers (baseline, user)
