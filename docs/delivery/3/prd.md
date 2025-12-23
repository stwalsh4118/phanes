# PBI-3: Baseline Server Setup Modules

[View in Backlog](../backlog.md#user-content-3)

## Overview

Implement the essential server setup modules that every VPS needs: timezone/locale configuration, non-root user creation, security hardening, swap file setup, and automatic security updates.

## Problem Statement

Every server needs:
- Proper timezone and locale configuration
- A non-root user with SSH key access
- Basic security hardening (UFW, fail2ban, SSH config)
- Swap file for memory management
- Automatic security updates

These are foundational requirements before installing any services.

## User Stories

- As a developer, I want timezone configuration so that logs and timestamps are correct
- As a developer, I want a non-root user so that I don't run everything as root
- As a developer, I want security hardening so that my server is protected from common attacks
- As a developer, I want swap configured so that the server handles memory pressure
- As a developer, I want automatic updates so that security patches are applied

## Technical Approach

### Baseline Module
- Set timezone using `timedatectl`
- Configure locale (UTF-8)
- Run `apt-get update`

### User Module
- Create user if doesn't exist
- Add SSH public key to `~/.ssh/authorized_keys`
- Add user to sudoers (passwordless sudo)
- Set up `.ssh` directory with correct permissions

### Security Module
- Configure UFW firewall (allow SSH, HTTP, HTTPS)
- Install and configure fail2ban
- Harden SSH config (disable password auth, root login, etc.)
- Use embedded templates for SSH and fail2ban configs

### Swap Module
- Check if swap already exists
- Create swap file if needed
- Configure `/etc/fstab` for persistence
- Set swappiness to reasonable value

### Updates Module
- Install `unattended-upgrades`
- Configure automatic security updates
- Enable automatic reboot for security updates (optional)

## UX/UI Considerations

- Security module should warn about disabling password auth
- Swap creation should show progress for large files
- User module should verify SSH key format

## Acceptance Criteria

1. All five modules implemented and idempotent
2. Baseline module sets timezone and updates apt
3. User module creates user and configures SSH access
4. Security module configures UFW, fail2ban, and SSH hardening
5. Swap module creates and configures swap file
6. Updates module enables automatic security updates
7. All modules have proper error handling
8. Templates embedded for SSH and fail2ban configs

## Dependencies

- PBI-1 (Foundation) must be complete
- PBI-2 (Module Framework) must be complete

## Open Questions

- Should automatic reboot for security updates be configurable or default off?

## Related Tasks

See [Tasks](./tasks.md)

