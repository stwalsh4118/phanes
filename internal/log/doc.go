// Package log provides structured logging for Phanes using zerolog.
// It supports colored console output, dry-run mode, and different log levels.
//
// Features:
//   - Colored console output for better readability
//   - Dry-run mode support (adds dry_run field to all logs)
//   - Separate stdout/stderr handling
//   - Thread-safe logging operations
//   - Multiple log levels: Info, Success, Warn, Error, Skip
//
// Log Levels:
//   - Info: General informational messages
//   - Success: Successful operations (includes success=true field)
//   - Warn: Warning messages
//   - Error: Error messages (written to stderr)
//   - Skip: Skipped operations (includes skip=true field)
//
// Usage:
//
//	// Enable dry-run mode
//	log.SetDryRun(true)
//
//	// Log messages
//	log.Info("Starting module execution")
//	log.Success("Module installed successfully")
//	log.Warn("Docker not found, skipping")
//	log.Error("Failed to install module: %v", err)
//	log.Skip("Module already installed")
package log
