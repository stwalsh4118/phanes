# PBI-2: Module Framework

[View in Backlog](../backlog.md#user-content-2)

## Overview

Create the module interface and runner system that allows provisioning components to be composed and executed idempotently. This is the core architecture that enables modularity.

## Problem Statement

We need a way to:
- Define modules that can check if they're installed
- Execute modules in a specific order
- Compose modules into profiles
- Ensure modules are idempotent (safe to run multiple times)

Without this framework, we can't build reusable provisioning components.

## User Stories

- As a developer, I want a module interface so that all provisioning components follow the same pattern
- As a developer, I want a runner so that modules execute in order with proper error handling
- As a developer, I want profiles so that I can combine modules for different server types

## Technical Approach

### Module Interface
```go
type Module interface {
    Name() string
    Description() string
    IsInstalled() (bool, error)
    Install(cfg *config.Config) error
}
```

### Runner Package
- `Runner` struct that holds module registry
- `RegisterModule(module Module)` - Add module to registry
- `RunModules(names []string, cfg *config.Config, dryRun bool) error` - Execute modules by name
- Handle errors gracefully, continue or stop based on severity

### Profile Package
- Define profiles as maps of module name lists
- `GetProfile(name string) ([]string, error)` - Get module list for profile
- `ListProfiles() []string` - List available profiles

### Module Registry
- Central registry of all available modules
- Modules register themselves
- Runner looks up modules by name

## UX/UI Considerations

- Clear output showing which module is running
- Skip messages when modules are already installed
- Error messages should indicate which module failed

## Acceptance Criteria

1. Module interface is defined and documented
2. Runner can execute modules in order
3. Runner handles idempotency checks correctly
4. Profiles can be defined and retrieved
5. Dry-run mode shows what would happen without executing
6. Error handling prevents partial installations

## Dependencies

- PBI-1 (Foundation) must be complete

## Open Questions

None

## Related Tasks

See [Tasks](./tasks.md)

