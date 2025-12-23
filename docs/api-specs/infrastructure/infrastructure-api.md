# Infrastructure API

Last Updated: 2025-01-27

## Logging Package

Package: `github.com/stwalsh4118/phanes/internal/log`

Provides colored, structured logging using zerolog with console output. All log functions support format strings and arguments.

### Public Functions

```go
// SetDryRun enables or disables dry-run mode. When enabled, all log messages
// include a dry_run: true field.
func SetDryRun(enabled bool)

// Info logs an informational message to stdout (cyan).
func Info(format string, args ...interface{})

// Success logs a success message to stdout at info level with success field (cyan).
func Success(format string, args ...interface{})

// Error logs an error message to stderr (red).
func Error(format string, args ...interface{})

// Skip logs a skip message to stdout at info level with skip field (cyan).
func Skip(format string, args ...interface{})

// Warn logs a warning message to stdout (yellow).
func Warn(format string, args ...interface{})
```

### Usage Examples

```go
import "github.com/stwalsh4118/phanes/internal/log"

// Basic logging
log.Info("Starting installation")
log.Success("Docker installed successfully")
log.Warn("Configuration file not found, using defaults")
log.Error("Failed to connect to database: %v", err)
log.Skip("Package already installed")

// With format strings
log.Info("Installing %s version %s", "docker", "24.0")

// Dry-run mode
log.SetDryRun(true)
log.Info("Would install docker") // Output includes dry_run: true field
log.SetDryRun(false)
```

### Behavior

- **Output**: Info, Success, Skip, Warn → stdout; Error → stderr
- **Colors**: Automatic via zerolog ConsoleWriter (info=cyan, warn=yellow, error=red)
- **Dry-run**: Adds `dry_run: true` field to all log entries when enabled
- **Thread-safe**: All functions are safe for concurrent use
- **Format**: Uses zerolog's console format with timestamps (HH:MM:SS)

### Dependencies

- `github.com/rs/zerolog` v1.33.0

