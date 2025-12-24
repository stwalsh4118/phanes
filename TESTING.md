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
