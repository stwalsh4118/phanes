# PBI-7: Development Tools and Coolify Modules

[View in Backlog](../backlog.md#user-content-7)

## Overview

Implement development tools module (Git, build tools, language runtimes) and Coolify installation module for self-hosted PaaS capabilities.

## Problem Statement

Development servers need:
- Git for version control
- Build tools (gcc, make, etc.)
- Language runtimes (Node.js, Python, Go, etc.)
- Optional: Coolify for self-hosted application deployment

These enable development workflows and application hosting.

## User Stories

- As a developer, I want development tools installed so that I can build and run applications
- As a developer, I want Coolify installed so that I can deploy applications easily

## Technical Approach

### DevTools Module
- Install Git
- Install build-essential (gcc, make, etc.)
- Install Node.js (via NodeSource or nvm)
- Install Python 3 and pip
- Install Go (latest stable)
- Optionally install other tools (Docker already handled separately)

### Coolify Module
- Check Docker is installed (dependency)
- Clone Coolify repository or use installation script
- Set up Coolify with Docker Compose
- Configure Coolify service
- Provide access URL and initial setup instructions

## UX/UI Considerations

- DevTools installation can take time - show progress
- Coolify requires Docker - check dependency first
- Show versions of installed tools

## Acceptance Criteria

1. DevTools module installs Git, build tools, and language runtimes
2. Coolify module installs and configures Coolify
3. Both modules are idempotent
4. Both modules verify installations
5. Coolify module checks Docker dependency
6. Proper error handling

## Dependencies

- PBI-1 (Foundation) must be complete
- PBI-2 (Module Framework) must be complete
- PBI-4 (Docker module) should be complete for Coolify

## Open Questions

- Which Node.js version to install? (LTS recommended)
- Should we support multiple Python versions?

## Related Tasks

See [Tasks](./tasks.md)

