// Package profile provides server profile management for Phanes.
// Profiles are pre-configured sets of modules that represent common
// server configurations (e.g., dev, web, database).
//
// Available Profiles:
//   - minimal: Basic secure server setup (baseline, user, security, swap, updates)
//   - dev: Development environment (includes Docker, monitoring, devtools)
//   - web: Web server configuration (includes Docker, monitoring, Caddy)
//   - database: Database server (includes Docker, monitoring, PostgreSQL, Redis)
//   - coolify: Self-hosted PaaS platform (includes Docker, Coolify)
//
// Usage:
//
//	// Get modules for a profile
//	modules, err := profile.GetProfile("dev")
//	if err != nil {
//	    log.Fatal("Profile not found: %v", err)
//	}
//
//	// List all available profiles
//	profiles := profile.ListProfiles()
//
//	// Check if profile exists
//	if profile.ProfileExists("web") {
//	    log.Info("Web profile is available")
//	}
package profile

