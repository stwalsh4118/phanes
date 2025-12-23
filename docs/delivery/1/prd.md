# PBI-1: Foundation - Core Infrastructure

[View in Backlog](../backlog.md#user-content-1)

## Overview

Establish the foundational infrastructure for the Phanes VPS provisioning system, including logging, command execution helpers, and configuration management. This provides the base layer that all modules will depend on.

## Problem Statement

Before building provisioning modules, we need:
- Consistent, colored logging output for user feedback
- Safe command execution helpers that handle errors properly
- YAML configuration parsing with validation
- File system utilities for idempotency checks

Without these foundations, modules would duplicate code and have inconsistent behavior.

## User Stories

- As a developer, I want colored log output so that I can easily see what the tool is doing
- As a developer, I want exec helpers so that I don't have to write command execution boilerplate
- As a developer, I want YAML config parsing so that server configuration is easy to manage
- As a developer, I want file utilities so that modules can check if things are already installed

## Technical Approach

### Package Structure
```
internal/
├── log/
│   └── log.go          # Colored logging (info, success, error, skip, warn)
├── exec/
│   └── exec.go         # Command execution helpers
└── config/
    └── config.go       # YAML config struct and parsing
```

### Logging Package
- Use `fmt` with ANSI color codes for terminal output
- Functions: `Info()`, `Success()`, `Error()`, `Skip()`, `Warn()`
- Support for dry-run mode (prefix with [DRY-RUN])

### Exec Package
- `Run(name string, args ...string) error` - Execute command, return error
- `RunWithOutput(name string, args ...string) (string, error)` - Execute and capture output
- `CommandExists(cmd string) bool` - Check if command is in PATH
- `FileExists(path string) bool` - Check if file exists
- `WriteFile(path string, content []byte, perm os.FileMode) error` - Write file with permissions

### Config Package
- Define `Config` struct with YAML tags
- Load from `config.yaml` file
- Validate required fields
- Provide sensible defaults

## UX/UI Considerations

- Logging should be clear and informative
- Errors should be actionable
- Config validation errors should point to the exact field

## Acceptance Criteria

1. Logging package provides colored output for all log levels
2. Exec helpers can run commands, check existence, and manage files
3. Config package loads YAML and validates structure
4. All packages have basic unit tests
5. Code follows Go best practices (error handling, no panics)

## Dependencies

- Go standard library
- `gopkg.in/yaml.v3` for YAML parsing

## Open Questions

None

## Related Tasks

See [Tasks](./tasks.md)

