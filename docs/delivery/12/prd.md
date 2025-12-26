# PBI-12: Backup Configuration Module

[View in Backlog](../backlog.md#user-content-12)

## Overview

Create a backup module that installs and configures restic backup tool with automated scheduling. This enables automated backups of server data to local or remote repositories during provisioning.

## Problem Statement

Setting up automated backups manually requires:
- Installing backup tools (restic)
- Configuring backup repositories (local or remote)
- Setting up authentication credentials
- Creating backup scripts
- Configuring cron jobs or systemd timers
- Testing backup functionality

This is complex and error-prone. Having backups configured automatically during provisioning ensures data protection is in place from the start.

## User Stories

- As a developer, I want restic installed automatically so that I can back up server data
- As a developer, I want to configure backup repositories so that backups are stored securely
- As a developer, I want automated backup scheduling so that backups happen regularly
- As a developer, I want to specify what paths to backup so that I control what's protected
- As a developer, I want idempotent configuration so that I can update backup settings safely

## Technical Approach

### Module Implementation

Create `internal/modules/backup/backup.go` implementing the `Module` interface:

```go
type BackupModule struct{}

func (m *BackupModule) Name() string { return "backup" }
func (m *BackupModule) Description() string { return "Installs and configures restic backup" }
func (m *BackupModule) IsInstalled() (bool, error) { ... }
func (m *BackupModule) Install(cfg *config.Config) error { ... }
```

### Installation Process

1. Check if restic is already installed (`restic version`)
2. Install restic via apt or download binary
3. Initialize backup repository (if not exists)
4. Create backup script with configured paths
5. Set up systemd timer or cron job for scheduled backups
6. Create password file for repository (secure permissions)

### Repository Types

Support multiple repository backends:
- **Local**: File system path (e.g., `/backups/repo`)
- **S3**: S3-compatible storage (requires access key/secret)
- **B2**: Backblaze B2 (requires key ID/secret)
- **SFTP**: Remote server via SFTP (requires SSH key)

### Configuration

Add to `internal/config/config.go`:

```go
type Config struct {
    // ... existing fields
    Backup Backup `yaml:"backup"`
}

type Backup struct {
    Enabled    bool     `yaml:"enabled"`
    Repository string   `yaml:"repository"`  // Repository URL or path
    Password   string   `yaml:"password"`    // Repository password
    Paths      []string `yaml:"paths"`       // Paths to backup
    Schedule   string   `yaml:"schedule"`   // Cron expression or systemd timer spec
    Retention  string   `yaml:"retention"`   // Retention policy (e.g., "30d", "12m")
}
```

### Backup Script

Create `/usr/local/bin/phanes-backup.sh`:
- Uses restic to backup configured paths
- Handles repository authentication
- Logs backup results
- Sends notifications on failure (optional)

### Scheduling

- **Systemd Timer** (preferred): Create timer unit for scheduled backups
- **Cron** (fallback): Create cron job if systemd not available
- Default schedule: Daily at 2 AM

### IsInstalled Check

- Check if restic is installed
- Check if repository is initialized
- Check if backup script exists
- Check if timer/cron job is configured

## UX/UI Considerations

- Clear logging during backup setup
- Show repository location and type after configuration
- Display backup schedule information
- Warn if password is weak or missing
- Show example restore command after setup

## Acceptance Criteria

1. Module installs restic backup tool
2. Module configures a backup repository (local or remote)
3. Module sets up cron/systemd timer for automated backups
4. Config struct includes backup section with all required fields
5. Module is idempotent (safe to run multiple times)
6. Backup script handles errors gracefully
7. Repository password is stored securely
8. Module validates repository connectivity

## Dependencies

- PBI-1 (Foundation) must be complete - exec helpers needed
- PBI-2 (Module Framework) must be complete - module interface needed
- Repository credentials required from user (if using remote storage)

## Open Questions

- Should we support multiple backup repositories? (Future enhancement)
- Should we include backup verification/restore testing? (Future enhancement)
- Should we support backup notifications (email/Slack)? (Future enhancement)
- What default retention policy should we use? (Suggest 30 days)

## Related Tasks

See [Tasks](./tasks.md)

