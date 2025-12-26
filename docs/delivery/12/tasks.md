# Tasks for PBI 12: Backup Configuration Module

This document lists all tasks associated with PBI 12.

**Parent PBI**: [PBI 12: Backup Configuration Module](./prd.md)

## Task Summary

| Task ID | Name | Status | Description |
| :------ | :--- | :----- | :---------- |
| 12-1 | [Add Backup Configuration Struct](./12-1.md) | Proposed | Add Backup config section to `internal/config/config.go` with Enabled, Repository, Password, Paths, Schedule, and Retention fields |
| 12-2 | [Implement Backup Module](./12-2.md) | Proposed | Create `internal/modules/backup/backup.go` implementing the Module interface to install and configure restic backup |
| 12-3 | [Register Module and Update Examples](./12-3.md) | Proposed | Register module in `main.go` and update `config.yaml.example` with Backup configuration options |
| 12-4 | [E2E CoS Test](./12-4.md) | Proposed | Verify all acceptance criteria work in both normal and dry-run modes |

