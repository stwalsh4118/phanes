# PBI-5: Web Server Modules

[View in Backlog](../backlog.md#user-content-5)

## Overview

Implement Nginx and Caddy web server modules. These provide HTTP/HTTPS serving capabilities for web applications.

## Problem Statement

Web servers are needed to:
- Serve static files
- Reverse proxy to applications
- Handle SSL/TLS termination
- Route traffic to different services

Nginx is traditional and powerful, Caddy provides automatic HTTPS.

## User Stories

- As a developer, I want Nginx installed so that I can serve web applications
- As a developer, I want Caddy installed so that I can get automatic HTTPS certificates

## Technical Approach

### Nginx Module
- Install nginx from apt
- Enable and start nginx service
- Create basic configuration directory structure
- Verify nginx is running and accessible

### Caddy Module
- Download Caddy binary
- Install to `/usr/local/bin/caddy`
- Create systemd service file
- Create Caddyfile in `/etc/caddy/`
- Enable and start Caddy service
- Verify Caddy is running

## UX/UI Considerations

- Both servers use port 80/443 - warn if already in use
- Caddy's automatic HTTPS is a selling point - mention in output
- Provide example configurations in comments

## Acceptance Criteria

1. Nginx module installs and starts nginx
2. Caddy module installs Caddy binary and creates service
3. Both modules are idempotent
4. Both modules verify service is running
5. Proper error handling for port conflicts

## Dependencies

- PBI-1 (Foundation) must be complete
- PBI-2 (Module Framework) must be complete

## Open Questions

- Should we create default configuration files or leave them empty?

## Related Tasks

See [Tasks](./tasks.md)

