# PBI-4: Service Modules - Docker and Monitoring

[View in Backlog](../backlog.md#user-content-4)

## Overview

Implement Docker CE and Docker Compose installation, plus Netdata monitoring module. These are common services needed for containerized applications and server monitoring.

## Problem Statement

Many server setups require:
- Docker for containerization
- Docker Compose for multi-container applications
- Monitoring to track server health and performance

These modules enable modern application deployment patterns.

## User Stories

- As a developer, I want Docker installed so that I can run containerized applications
- As a developer, I want Docker Compose so that I can manage multi-container setups
- As a developer, I want monitoring so that I can track server performance

## Technical Approach

### Docker Module
- Add Docker's official GPG key and repository
- Install Docker CE
- Add user to docker group
- Install Docker Compose v2 (as plugin)
- Verify installation with `docker --version` and `docker compose version`

### Monitoring Module (Netdata)
- Download and run Netdata's kickstart script
- Configure Netdata to start on boot
- Optionally configure basic settings
- Verify it's accessible on default port (19999)

## UX/UI Considerations

- Docker installation can take time - show progress
- Netdata kickstart is interactive - handle non-interactive mode
- Warn if user is not in docker group (needs logout/login)

## Acceptance Criteria

1. Docker module installs Docker CE and Compose v2
2. Docker module adds user to docker group
3. Monitoring module installs Netdata
4. Both modules are idempotent
5. Both modules verify installation success
6. Proper error handling for installation failures

## Dependencies

- PBI-1 (Foundation) must be complete
- PBI-2 (Module Framework) must be complete
- PBI-3 (Baseline Modules) should be complete (user module)

## Open Questions

- Should we support other monitoring tools (Prometheus, Grafana) or just Netdata for now?

## Related Tasks

See [Tasks](./tasks.md)

