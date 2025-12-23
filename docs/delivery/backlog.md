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

## PBI History

| Timestamp | PBI_ID | Event_Type | Details | User |
|:----------|:-------|:-----------|:--------|:----|
| 2025-01-27-000000 | 1-9 | create_pbi | Initial backlog created from plan | sean |

