# Product Backlog

This document contains all Product Backlog Items (PBIs) for the Phanes VPS Provisioning System, ordered by priority.

**Master PRD**: [Product Requirements Document](../prd.md)

| ID | Actor | User Story | Status | Conditions of Satisfaction (CoS) |
|:---|:------|:-----------|:-------|:--------------------------------|
| 1 | Developer | As a developer, I want core infrastructure (logging, exec helpers, config parsing) so that modules can be built on a solid foundation | Proposed | Logging package with colored output, exec helpers for command execution, YAML config parsing with struct validation |
| 2 | Developer | As a developer, I want a module interface and runner system so that provisioning components can be composed and executed idempotently | Proposed | Module interface defined, runner executes modules in order, idempotency checks work correctly |
| 3 | Developer | As a developer, I want baseline server setup modules (timezone, user, security, swap, updates) so that every server has a secure foundation | Proposed | Baseline, user, security, swap, and updates modules implemented and tested |
| 4 | Developer | As a developer, I want Docker and monitoring modules so that servers can run containers and be monitored | Proposed | Docker CE + Compose module, Netdata monitoring module, both idempotent |
| 5 | Developer | As a developer, I want web server modules (Nginx and Caddy) so that I can serve web applications | Proposed | Nginx module, Caddy module, both configurable and idempotent |
| 6 | Developer | As a developer, I want database modules (PostgreSQL and Redis) so that applications can store and cache data | Proposed | PostgreSQL module, Redis module, both configurable and idempotent |
| 7 | Developer | As a developer, I want development tools and Coolify modules so that I can set up dev servers and hosting platforms | Proposed | Dev tools module (Git, build tools, language runtimes), Coolify installation module |
| 8 | Developer | As a developer, I want a CLI with profiles so that I can easily provision different server types | Proposed | CLI with profile selection, module selection, dry-run mode, config file support |
| 9 | Developer | As a developer, I want documentation and examples so that I can understand and use the provisioning system | Proposed | README with usage examples, config.yaml.example with all options documented, inline code documentation |
| 10 | Developer | As a developer, I want a clear summary table at the end of provisioning runs so that I can immediately see what was installed, skipped, or failed | Proposed | Runner tracks module execution results (installed/skipped/failed/error); Summary table displayed after all modules complete; Color-coded status indicators; Shows total counts (X installed, Y skipped, Z failed); Works in both normal and dry-run modes |
| 11 | Developer | As a developer, I want a Tailscale module so that I can automatically join my servers to my tailnet during provisioning | Proposed | Module installs Tailscale via official install script; Module authenticates using provided auth key from config; IsInstalled() correctly detects if Tailscale is already configured; Module is idempotent; Config struct includes tailscale section with auth_key field |
| 12 | Developer | As a developer, I want a backup module so that I can set up automated backups on my servers during provisioning | Proposed | Module installs restic backup tool; Module configures a backup repository (local or remote); Module sets up cron/systemd timer for automated backups; Config struct includes backup paths and schedule; Module is idempotent |
| 13 | Developer | As a developer, I want a K3s module so that I can quickly set up a lightweight Kubernetes cluster on my servers | Proposed | Module installs K3s using official install script; Module configures as server (single-node) or agent; kubectl is available and configured for the user; Config struct includes k3s options; Module is idempotent |
| 14 | Developer | As a developer, I want post-install health checks so that I can verify all installed services are running correctly after provisioning | Proposed | Each module optionally implements a HealthCheck() method; Runner can optionally run health checks after installation; Health check results included in execution summary; Failed health checks are clearly reported with troubleshooting hints |
| 15 | Developer | As a developer, I want to run Phanes remotely via SSH and track installation state so that I can provision servers from my local machine and understand what's already installed | Proposed | CLI supports --remote user@host to execute on remote server via SSH; State file tracks what modules have been installed and when; State can be queried with phanes status; Remote execution uploads binary or runs commands over SSH |
| 16 | Developer | As a developer, I want a TUI version of Phanes so that I can interactively configure and run provisioning without editing config files | Proposed | TUI allows interactive module selection with checkboxes; TUI allows configuration editing for all config options; TUI shows real-time progress during module execution; TUI displays execution summary table; TUI works alongside existing CLI (accessible via phanes tui or --tui flag); TUI can load and save config files |

## PBI History

| Timestamp | PBI_ID | Event_Type | Details | User |
|:----------|:-------|:-----------|:--------|:----|
| 2025-01-27-000000 | 1-9 | create_pbi | Initial backlog created from plan | sean |
| 2025-01-27-000000 | 10 | create_pbi | PBI 10: Execution Summary Report created | sean |
| 2025-01-27-000000 | 11 | create_pbi | PBI 11: Tailscale VPN Module created | sean |
| 2025-01-27-000000 | 12 | create_pbi | PBI 12: Backup Configuration Module created | sean |
| 2025-01-27-000000 | 13 | create_pbi | PBI 13: K3s Lightweight Kubernetes Module created | sean |
| 2025-01-27-000000 | 14 | create_pbi | PBI 14: Post-Install Health Check created | sean |
| 2025-01-27-000000 | 15 | create_pbi | PBI 15: Remote Execution and State Tracking created | sean |
| 2025-01-27-000000 | 16 | create_pbi | PBI 16: TUI Version created | sean |

