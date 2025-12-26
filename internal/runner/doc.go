// Package runner provides the module execution engine for Phanes.
// It manages a registry of modules and executes them in order with
// idempotency checks and dry-run support.
//
// Key Features:
//   - Module registry management
//   - Idempotent execution (checks IsInstalled before Install)
//   - Dry-run mode support
//   - Error handling and aggregation
//   - Module discovery and listing
//
// Usage:
//
//	// Create a new runner
//	r := runner.NewRunner()
//
//	// Register modules
//	r.RegisterModule(&baseline.BaselineModule{})
//	r.RegisterModule(&docker.DockerModule{})
//
//	// Execute modules
//	cfg := config.DefaultConfig()
//	err := r.RunModules([]string{"baseline", "docker"}, cfg, false)
//
//	// List available modules
//	modules := r.ListModules()
package runner

