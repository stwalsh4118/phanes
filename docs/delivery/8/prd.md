# PBI-8: CLI and Profiles

[View in Backlog](../backlog.md#user-content-8)

## Overview

Create the main CLI entry point with profile support, module selection, dry-run mode, and configuration file handling. This ties everything together into a usable tool.

## Problem Statement

Users need:
- A simple CLI to run provisioning
- Predefined profiles for common server types
- Ability to select specific modules
- Dry-run mode to preview changes
- Configuration file support

This is the user-facing interface of the entire system.

## User Stories

- As a user, I want to run a profile so that I can quickly set up a server type
- As a user, I want to select specific modules so that I can customize my setup
- As a user, I want dry-run mode so that I can see what would happen
- As a user, I want configuration files so that I can customize settings

## Technical Approach

### CLI (main.go)
- Use `flag` package for command-line arguments
- Support `--profile` flag for profile selection
- Support `--modules` flag for module selection (comma-separated)
- Support `--config` flag for config file path
- Support `--dry-run` flag for preview mode
- Support `--list` flag to show available modules/profiles

### Profile Integration
- Load profiles from profile package
- Validate profile exists
- Resolve module names to Module instances
- Execute via runner

### Config Integration
- Load config from file (default `config.yaml`)
- Validate config structure
- Merge with defaults
- Pass to modules

### Error Handling
- Clear error messages
- Exit codes (0 success, 1 error)
- Help text for usage

## UX/UI Considerations

- Clear help text
- Progress indicators for long operations
- Summary at end showing what was installed
- Color-coded output (success/error/skip)

## Acceptance Criteria

1. CLI accepts profile, modules, config, and dry-run flags
2. CLI can list available modules and profiles
3. Profile execution works correctly
4. Module selection works correctly
5. Dry-run mode shows actions without executing
6. Config file loading and validation works
7. Help text is clear and comprehensive
8. Error handling provides actionable messages

## Dependencies

- All previous PBIs must be complete (modules, framework, etc.)

## Open Questions

- Should we support interactive mode for config creation?

## Related Tasks

See [Tasks](./tasks.md)

