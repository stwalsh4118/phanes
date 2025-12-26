# PBI-14: Post-Install Health Check

[View in Backlog](../backlog.md#user-content-14)

## Overview

Add an optional post-install health check system that verifies all installed services are running correctly after provisioning. This provides confidence that provisioning was successful and services are operational.

## Problem Statement

After provisioning completes, users need to manually verify:
- Services are running (systemctl status)
- Ports are listening (netstat/ss)
- Services respond to requests
- Configuration is correct

This is time-consuming and error-prone. Having automated health checks provides immediate feedback on provisioning success and identifies issues early.

## User Stories

- As a developer, I want health checks after installation so that I know services are working
- As a developer, I want to see which services passed or failed so that I can identify issues
- As a developer, I want troubleshooting hints for failures so that I can fix problems quickly
- As a developer, I want optional health checks so that I can skip them if needed

## Technical Approach

### Health Checker Interface

Extend the `Module` interface with an optional `HealthChecker` interface:

```go
type HealthChecker interface {
    HealthCheck() error
}
```

Modules can optionally implement this interface to provide health verification.

### Runner Enhancement

Modify `internal/runner/runner.go` to support health checks:

```go
func (r *Runner) RunHealthChecks(moduleNames []string) ([]HealthCheckResult, error) {
    // Run health checks for specified modules
    // Return results with pass/fail status
}
```

### Health Check Results

```go
type HealthCheckResult struct {
    ModuleName string
    Status     HealthStatus  // Pass, Fail, Skipped
    Error      error
    Hints      []string     // Troubleshooting hints
}

type HealthStatus string

const (
    HealthPass   HealthStatus = "pass"
    HealthFail   HealthStatus = "fail"
    HealthSkipped HealthStatus = "skipped"
)
```

### Module Health Checks

Each module can implement health checks specific to its service:

- **Docker**: Check `docker ps` works, docker daemon is running
- **PostgreSQL**: Check service status, can connect, can query
- **Redis**: Check service status, can connect, can ping
- **Nginx/Caddy**: Check service status, port listening, HTTP response
- **Tailscale**: Check service status, connected state, IP assigned
- **K3s**: Check service status, kubectl access, nodes ready

### CLI Integration

Add `--health-check` flag to CLI:
- When set, run health checks after module installation
- Display health check results in summary
- Include health check results in execution summary table (PBI-10)

### Health Check Display

```
Health Check Results:
┌─────────────┬────────┬─────────────────────────────────┐
│ Module      │ Status │ Details                         │
├─────────────┼────────┼─────────────────────────────────┤
│ docker      │ ✓ Pass │ Docker daemon running           │
│ postgres    │ ✓ Pass │ Service active, connection OK   │
│ redis       │ ✗ Fail │ Service not running             │
│             │        │ Hint: Check logs with journalctl │
└─────────────┴────────┴─────────────────────────────────┘
```

## UX/UI Considerations

- Health checks should be optional (opt-in via flag)
- Clear visual indicators for pass/fail
- Troubleshooting hints should be actionable
- Health check results integrated into execution summary
- Should not block provisioning if health checks fail (informational)

## Acceptance Criteria

1. Each module optionally implements a `HealthCheck()` method
2. Runner can optionally run health checks after installation
3. Health check results are included in execution summary
4. Failed health checks are clearly reported with troubleshooting hints
5. Health checks are optional (controlled by CLI flag)
6. Health checks don't block provisioning completion
7. Common services have meaningful health checks implemented

## Dependencies

- PBI-2 (Module Framework) must be complete - module interface needed
- PBI-10 (Execution Summary) recommended - for displaying results
- Individual modules must be implemented to have health checks

## Open Questions

- Should health checks be run automatically or always require flag? (Require flag - opt-in)
- Should failed health checks cause non-zero exit code? (No - informational only)
- Should we support custom health check scripts? (Future enhancement)
- Should health checks be retryable? (Future enhancement)

## Related Tasks

See [Tasks](./tasks.md)

