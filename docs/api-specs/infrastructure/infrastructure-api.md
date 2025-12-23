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

## Exec Package

Package: `github.com/stwalsh4118/phanes/internal/exec`

Provides safe command execution helpers and file system utilities. All functions handle errors gracefully and follow Go best practices.

### Public Functions

```go
// Run executes a command with the given name and arguments.
// The command's stdout and stderr are connected to os.Stdout and os.Stderr respectively.
// Returns an error if the command fails to execute or exits with a non-zero status.
func Run(name string, args ...string) error

// RunWithOutput executes a command with the given name and arguments and captures its stdout.
// Returns the command's stdout as a string and an error if the command fails to execute
// or exits with a non-zero status.
func RunWithOutput(name string, args ...string) (string, error)

// CommandExists checks if a command exists in the system PATH.
// Returns true if the command is found, false otherwise.
func CommandExists(cmd string) bool

// FileExists checks if a file or directory exists at the given path.
// Returns true if the path exists, false otherwise.
func FileExists(path string) bool

// WriteFile writes content to a file at the given path with the specified permissions.
// If the file already exists, it will be overwritten.
// Returns an error if the file cannot be written.
func WriteFile(path string, content []byte, perm os.FileMode) error
```

### Usage Examples

```go
import (
    "os"
    "github.com/stwalsh4118/phanes/internal/exec"
)

// Execute a command
if err := exec.Run("docker", "version"); err != nil {
    log.Error("Failed to run docker: %v", err)
}

// Capture command output
output, err := exec.RunWithOutput("uname", "-a")
if err != nil {
    log.Error("Failed to get system info: %v", err)
} else {
    log.Info("System: %s", output)
}

// Check if command exists
if exec.CommandExists("docker") {
    log.Info("Docker is available")
} else {
    log.Warn("Docker not found in PATH")
}

// Check if file exists
if exec.FileExists("/etc/docker/daemon.json") {
    log.Skip("Docker config already exists")
} else {
    log.Info("Creating Docker config")
}

// Write a file
content := []byte("Hello, World!")
if err := exec.WriteFile("/tmp/test.txt", content, 0644); err != nil {
    log.Error("Failed to write file: %v", err)
}
```

### Behavior

- **Error Handling**: All functions return errors appropriately; no panics
- **Command Execution**: Uses `os/exec` package for safe command execution
- **File Operations**: Uses `os` package for file system operations
- **Idempotency**: FileExists and CommandExists are safe to call multiple times
- **Permissions**: WriteFile respects the provided file mode (Unix-like systems)

### Dependencies

- Standard library only (`os`, `os/exec`)

