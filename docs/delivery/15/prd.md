# PBI-15: Remote Execution and State Tracking

[View in Backlog](../backlog.md#user-content-15)

## Overview

Add support for running Phanes remotely via SSH and tracking installation state. This enables provisioning servers from a local machine and understanding what's already installed on remote servers.

## Problem Statement

Currently, Phanes must be run directly on the target server, which requires:
- Copying the binary to the server
- SSH'ing into the server
- Running Phanes locally
- Manually tracking what's installed

This is inconvenient for managing multiple servers. Supporting remote execution and state tracking enables:
- Provisioning servers from a local machine
- Understanding installation state without manual inspection
- Better automation and orchestration workflows

## User Stories

- As a developer, I want to run Phanes remotely via SSH so that I can provision servers from my local machine
- As a developer, I want state tracking so that I know what modules are installed on each server
- As a developer, I want to query installation state so that I can see what's already configured
- As a developer, I want state persistence so that idempotency works across runs

## Technical Approach

### Remote Execution

Add `--remote user@host` flag to CLI:

```go
type RemoteConfig struct {
    User string
    Host string
    Port int    // Default: 22
    Key  string // SSH key path (optional)
}
```

### SSH Implementation

Two approaches:

**Option 1: Upload and Execute**
- Upload Phanes binary to remote server via SCP
- Execute binary on remote server
- Stream output back to local terminal

**Option 2: Command Execution**
- Execute commands over SSH directly
- Use `ssh user@host 'phanes ...'` pattern
- Requires Phanes already on remote server

**Recommendation**: Option 1 (upload and execute) for better user experience.

### State Tracking

Create state file at `/var/lib/phanes/state.json`:

```go
type InstallationState struct {
    Modules []ModuleState `json:"modules"`
    Version string        `json:"phanes_version"`
    Updated time.Time     `json:"updated"`
}

type ModuleState struct {
    Name        string    `json:"name"`
    InstalledAt time.Time `json:"installed_at"`
    Version     string    `json:"version"`  // Optional: module version
    ConfigHash  string    `json:"config_hash"` // Optional: config snapshot
}
```

### State Management

- **Write State**: After successful module installation, update state file
- **Read State**: Before installation, check state to determine if already installed
- **State Location**: `/var/lib/phanes/state.json` (requires root/sudo)

### CLI Commands

Add new commands:

```bash
# Remote execution
phanes --remote user@host --profile dev --config config.yaml

# State query
phanes status                    # Show local state
phanes status --remote user@host # Show remote state
```

### State Display

```
Installation State:
┌─────────────┬─────────────────────┬──────────┐
│ Module      │ Installed At        │ Version  │
├─────────────┼─────────────────────┼──────────┤
│ baseline    │ 2025-01-27 10:30:00 │ 1.0.0    │
│ user        │ 2025-01-27 10:30:15 │ 1.0.0    │
│ docker      │ 2025-01-27 10:32:00 │ 1.0.0    │
└─────────────┴─────────────────────┴──────────┘
```

### Integration with Runner

- Runner checks state file before calling `IsInstalled()`
- Runner updates state file after successful installation
- State can supplement or replace `IsInstalled()` checks

## UX/UI Considerations

- Clear indication when executing remotely
- Show SSH connection status
- Display remote execution progress
- State query should be fast and informative
- Remote execution should handle SSH errors gracefully

## Acceptance Criteria

1. CLI supports `--remote user@host` to execute on remote server via SSH
2. State file tracks what modules have been installed and when
3. State can be queried with `phanes status` command
4. Remote execution uploads binary or runs commands over SSH
5. State file is created/updated after successful installations
6. State file location is configurable (default: `/var/lib/phanes/state.json`)
7. Remote execution handles SSH authentication (key-based or password)
8. State query works for both local and remote servers

## Dependencies

- PBI-1 (Foundation) must be complete - exec helpers needed
- PBI-2 (Module Framework) must be complete - runner system needed
- SSH access to remote servers required
- Go SSH library (e.g., `golang.org/x/crypto/ssh`) for implementation

## Open Questions

- Should we support SSH password authentication? (Key-based preferred, password optional)
- Should state file be JSON or YAML? (JSON - simpler parsing)
- Should we support state file encryption? (Future enhancement)
- Should state include config snapshots? (Useful for debugging)
- Should we support state migration/upgrades? (Future enhancement)

## Related Tasks

See [Tasks](./tasks.md)

