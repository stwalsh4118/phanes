// Package exec provides utilities for executing system commands and
// performing file operations in a cross-platform manner.
//
// It wraps the standard os/exec package with convenience functions for
// common operations like running commands, checking if commands exist,
// and file operations.
//
// Key Features:
//   - Execute commands with automatic stdout/stderr handling
//   - Capture command output
//   - Check if commands exist in PATH
//   - File existence checks
//   - File writing utilities
//
// Usage:
//
//	// Run a command
//	err := exec.Run("apt", "update")
//
//	// Capture output
//	output, err := exec.RunWithOutput("docker", "--version")
//
//	// Check if command exists
//	if exec.CommandExists("docker") {
//	    log.Info("Docker is available")
//	}
//
//	// Check if file exists
//	if exec.FileExists("/etc/nginx/nginx.conf") {
//	    log.Info("Nginx config exists")
//	}
package exec
