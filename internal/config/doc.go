// Package config provides configuration management for Phanes.
// It handles loading, parsing, and validating YAML configuration files
// that define settings for all provisioning modules.
//
// Configuration Structure:
//
// The Config struct contains nested configuration sections for each module:
//   - User: User creation and SSH key configuration
//   - System: System-level settings like timezone
//   - Swap: Swap file configuration
//   - Security: SSH and firewall settings
//   - Docker: Docker installation options
//   - Postgres: PostgreSQL database configuration
//   - Redis: Redis cache configuration
//   - Nginx: Nginx web server configuration
//   - Caddy: Caddy web server configuration
//   - DevTools: Development tools configuration
//   - Coolify: Coolify PaaS platform configuration
//
// Usage:
//
//	cfg, err := config.Load("config.yaml")
//	if err != nil {
//	    log.Fatal("Failed to load config: %v", err)
//	}
//
//	// Access configuration values
//	username := cfg.User.Username
//	timezone := cfg.System.Timezone
package config

