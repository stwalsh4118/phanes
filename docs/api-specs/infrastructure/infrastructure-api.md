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
- **Output Capture**: `RunWithOutput()` captures stdout only; stderr is not captured and will be written to the process's stderr
- **File Operations**: Uses `os` package for file system operations
- **Idempotency**: FileExists and CommandExists are safe to call multiple times
- **Permissions**: WriteFile respects the provided file mode (Unix-like systems)

### Dependencies

- Standard library only (`os`, `os/exec`)

## Config Package

Package: `github.com/stwalsh4118/phanes/internal/config`

Provides YAML configuration loading, parsing, validation, and sensible defaults for all Phanes module settings.

### Public Types

```go
// Config represents the complete configuration structure for Phanes.
type Config struct {
    User     User     `yaml:"user"`
    System   System   `yaml:"system"`
    Swap     Swap     `yaml:"swap"`
    Security Security `yaml:"security"`
    Docker   Docker   `yaml:"docker"`
    Postgres Postgres `yaml:"postgres"`
    Redis    Redis    `yaml:"redis"`
    Nginx    Nginx    `yaml:"nginx"`
    Caddy    Caddy    `yaml:"caddy"`
    DevTools DevTools `yaml:"devtools"`
    Coolify  Coolify  `yaml:"coolify"`
}

// User contains user-related configuration.
type User struct {
    Username     string `yaml:"username"`
    SSHPublicKey string `yaml:"ssh_public_key"`
}

// System contains system-level configuration.
type System struct {
    Timezone string `yaml:"timezone"`
}

// Swap contains swap file configuration.
type Swap struct {
    Enabled bool   `yaml:"enabled"`
    Size    string `yaml:"size"`
}

// Security contains security-related configuration.
type Security struct {
    SSHPort           int  `yaml:"ssh_port"`
    AllowPasswordAuth bool `yaml:"allow_password_auth"`
}

// Docker contains Docker-related configuration.
type Docker struct {
    InstallCompose bool `yaml:"install_compose"`
}

// Postgres contains PostgreSQL configuration.
type Postgres struct {
    Version  string `yaml:"version"`
    Password string `yaml:"password"`
}

// Redis contains Redis configuration.
type Redis struct {
    Password string `yaml:"password"`
}

// Nginx contains Nginx configuration.
type Nginx struct {
    Enabled bool `yaml:"enabled"`
}

// Caddy contains Caddy configuration.
type Caddy struct {
    Enabled bool `yaml:"enabled"`
}

// DevTools contains development tools configuration.
type DevTools struct {
    NodeVersion   string `yaml:"node_version"`
    PythonVersion string `yaml:"python_version"`
    GoVersion     string `yaml:"go_version"`
}

// Coolify contains Coolify configuration.
type Coolify struct {
    Enabled bool `yaml:"enabled"`
}
```

### Public Functions

```go
// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config

// Load reads and parses a YAML configuration file, applies defaults, and validates it.
// Returns the parsed Config and an error if loading, parsing, or validation fails.
func Load(path string) (*Config, error)

// Validate checks that all required fields in the Config are set.
// Returns an error if any required field is missing or empty.
func Validate(cfg *Config) error
```

### Usage Examples

```go
import "github.com/stwalsh4118/phanes/internal/config"

// Load configuration from file
cfg, err := config.Load("config.yaml")
if err != nil {
    log.Error("Failed to load config: %v", err)
    return
}

// Access configuration values
log.Info("Username: %s", cfg.User.Username)
log.Info("Timezone: %s", cfg.System.Timezone)

// Validate existing config
if err := config.Validate(cfg); err != nil {
    log.Error("Invalid config: %v", err)
}

// Get default configuration
defaultCfg := config.DefaultConfig()
```

### Default Values

- **System.Timezone**: `"UTC"`
- **Swap.Enabled**: `true`
- **Swap.Size**: `"2G"`
- **Security.SSHPort**: `22`
- **Security.AllowPasswordAuth**: `false`
- **Docker.InstallCompose**: `true`
- **Postgres.Version**: `"16"`
- **DevTools.NodeVersion**: `"20"`
- **DevTools.PythonVersion**: `"3.12"`
- **DevTools.GoVersion**: `"1.25"`
- **Nginx.Enabled**: `false`
- **Caddy.Enabled**: `false`
- **Coolify.Enabled**: `false`

### Required Fields

- `user.username` - Username for the deployment user
- `user.ssh_public_key` - SSH public key for the deployment user

### Behavior

- **YAML Parsing**: Uses `gopkg.in/yaml.v3` for YAML parsing
- **Defaults**: Missing optional fields are automatically filled with sensible defaults
- **Validation**: Required fields are validated after loading and applying defaults
- **Error Messages**: Validation errors clearly indicate which field is missing
- **File Reading**: Uses `os.ReadFile` to read configuration files

### Dependencies

- `gopkg.in/yaml.v3` v3.0.1

## CLI (Command-Line Interface)

Location: `main.go` (project root)

The CLI is built using Cobra framework for structured command handling, automatic help generation, and consistent flag parsing.

### Command Structure

```go
// Root command using Cobra
var rootCmd = &cobra.Command{
    Use:     "phanes",
    Short:   "VPS Provisioning System",
    Long:    "phanes is a tool for provisioning Linux VPS servers...",
    Version: "0.1.0",
    RunE:    runCommand,
}

// Command execution handler
func runCommand(cmd *cobra.Command, args []string) error
```

### Flags

All flags are defined on the root command:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--profile` | string | `""` | Profile name to execute (e.g., 'dev', 'web', 'database') |
| `--modules` | string | `""` | Comma-separated list of module names to execute |
| `--config` | string | `"config.yaml"` | Path to configuration file |
| `--dry-run` | bool | `false` | Enable dry-run mode (preview changes without executing) |
| `--list` | bool | `false` | List available modules and profiles |
| `--help` / `-h` | bool | - | Show help text (automatic via Cobra) |
| `--version` / `-v` | bool | - | Show version (automatic via Cobra) |

### Usage Examples

```bash
# Run a profile
./phanes --profile dev --config config.yaml

# Run specific modules
./phanes --modules baseline,user,docker --config config.yaml

# Preview changes without executing
./phanes --profile dev --config config.yaml --dry-run

# List available modules and profiles
./phanes --list

# Show help
./phanes --help

# Show version
./phanes --version
```

### Exit Codes

- **0**: Success
- **1**: General error
- **2**: Invalid usage (wrong flags or arguments)

### Behavior

- **Flag Parsing**: Uses Cobra's flag system (built on `spf13/pflag`)
- **Help Text**: Automatically generated by Cobra with examples and flag descriptions
- **Validation**: Requires either `--profile` or `--modules` to be specified (unless `--list` is used)
- **Dry-Run**: When `--dry-run` is set, calls `log.SetDryRun(true)` to enable dry-run mode in logging
- **Error Handling**: Returns errors from `runCommand()` which Cobra handles and displays with help text
- **Version**: Built-in version support via Cobra's `Version` field

### Implementation Details

- **Command Execution**: `runCommand()` function handles all command logic
- **Flag Access**: Flags are accessed via package-level variables set by Cobra
- **Error Propagation**: Errors returned from `runCommand()` are handled in `main()` with appropriate exit codes
- **Help Integration**: Cobra automatically shows help when `--help` is used or when validation fails

### Dependencies

- `github.com/spf13/cobra` v1.10.2
- `github.com/spf13/pflag` v1.0.10 (indirect, used by Cobra)

## Module Interface

Package: `github.com/stwalsh4118/phanes/internal/module`

Defines the core interface that all provisioning modules must implement. This interface ensures consistency across all modules and enables the runner to execute modules in a uniform way.

### Public Interface

```go
// Module defines the interface that all provisioning modules must implement.
type Module interface {
    // Name returns the unique name identifier for this module.
    Name() string

    // Description returns a human-readable description of what this module does.
    Description() string

    // IsInstalled checks if this module is already installed/configured.
    // Returns true if installed, false if needs installation, error on check failure.
    IsInstalled() (bool, error)

    // Install performs the installation and configuration of this module.
    // The cfg parameter provides access to all configuration values.
    // Returns an error if installation fails.
    Install(cfg *config.Config) error
}
```

### Usage Examples

```go
import (
    "github.com/stwalsh4118/phanes/internal/module"
    "github.com/stwalsh4118/phanes/internal/config"
)

// Example module implementation
type BaselineModule struct{}

func (m *BaselineModule) Name() string {
    return "baseline"
}

func (m *BaselineModule) Description() string {
    return "Sets timezone, locale, and runs apt update"
}

func (m *BaselineModule) IsInstalled() (bool, error) {
    // Check if baseline configuration is already applied
    return false, nil
}

func (m *BaselineModule) Install(cfg *config.Config) error {
    // Perform installation steps using cfg for configuration
    return nil
}

// Register module with runner
runner.RegisterModule(&BaselineModule{})
```

### Behavior

- **Idempotency**: All modules must be idempotent - safe to run multiple times. The runner calls `IsInstalled()` before `Install()`.
- **Error Handling**: Both `IsInstalled()` and `Install()` return errors that the runner handles appropriately.
- **Configuration**: Modules receive a `*config.Config` parameter in `Install()` to access all configuration values.
- **Naming**: Module names must be unique and are used for registration and profile references.

### Dependencies

- `github.com/stwalsh4118/phanes/internal/config` - For configuration access

## Runner Package

Package: `github.com/stwalsh4118/phanes/internal/runner`

Manages a registry of modules and executes them in order with proper error handling and idempotency checks. The runner ensures modules are only installed if they're not already installed, and supports dry-run mode for previewing actions.

### Public Types

```go
// Runner manages a registry of modules and executes them in order.
type Runner struct {
    // modules is a map of module names to Module instances
    modules map[string]module.Module
}
```

### Public Functions

```go
// NewRunner creates a new Runner instance with an empty module registry.
func NewRunner() *Runner

// RegisterModule adds a module to the registry.
// If a module with the same name is already registered, it will be overwritten
// and a warning will be logged.
func (r *Runner) RegisterModule(mod module.Module)

// RunModules executes the specified modules in order.
// It checks IsInstalled() before calling Install() to ensure idempotency.
// If dryRun is true, it logs what would happen without actually executing Install().
// Returns an error if any module fails to execute or if a module name is not found.
func (r *Runner) RunModules(names []string, cfg *config.Config, dryRun bool) error

// GetModule returns a module from the registry by name.
// Returns nil if the module is not found.
func (r *Runner) GetModule(name string) module.Module

// ListModules returns a list of all registered module names.
func (r *Runner) ListModules() []string
```

### Usage Examples

```go
import (
    "github.com/stwalsh4118/phanes/internal/runner"
    "github.com/stwalsh4118/phanes/internal/config"
    "github.com/stwalsh4118/phanes/internal/module"
)

// Create a new runner
r := runner.NewRunner()

// Register modules
baselineMod := &BaselineModule{}
dockerMod := &DockerModule{}
r.RegisterModule(baselineMod)
r.RegisterModule(dockerMod)

// Load configuration
cfg, err := config.Load("config.yaml")
if err != nil {
    log.Error("Failed to load config: %v", err)
    return
}

// Execute modules
moduleNames := []string{"baseline", "docker"}
err = r.RunModules(moduleNames, cfg, false)
if err != nil {
    log.Error("Failed to execute modules: %v", err)
}

// Dry-run mode
err = r.RunModules(moduleNames, cfg, true)
if err != nil {
    log.Error("Dry-run failed: %v", err)
}

// List all registered modules
modules := r.ListModules()
log.Info("Available modules: %v", modules)

// Get a specific module
mod := r.GetModule("baseline")
if mod != nil {
    log.Info("Found module: %s", mod.Description())
}
```

### Behavior

- **Idempotency**: The runner checks `IsInstalled()` before calling `Install()` for each module. If a module is already installed, it is skipped with a log message.
- **Error Handling**: Errors from `IsInstalled()` or `Install()` are logged and collected. The runner continues processing remaining modules even if one fails, but returns an error at the end if any modules failed.
- **Dry-Run Mode**: When `dryRun` is true, the runner checks `IsInstalled()` but does not call `Install()`. It logs what would happen without making changes.
- **Module Registry**: Modules are registered by their name (from `Module.Name()`). Duplicate registrations overwrite the previous module with a warning.
- **Order**: Modules are executed in the order specified in the `names` slice.
- **Unknown Modules**: If a module name is not found in the registry, an error is logged and the runner continues with the next module.

### Error Handling

- **Empty Module List**: Returns an error if no modules are specified.
- **Unknown Module**: Returns an error if a module name is not found in the registry.
- **IsInstalled Error**: If `IsInstalled()` returns an error, it is logged and the module is skipped.
- **Install Error**: If `Install()` returns an error, it is logged and collected. The runner continues with remaining modules.
- **Multiple Errors**: If multiple modules fail, all errors are collected and returned together.

### Dependencies

- `github.com/stwalsh4118/phanes/internal/module` - Module interface
- `github.com/stwalsh4118/phanes/internal/config` - Configuration structure
- `github.com/stwalsh4118/phanes/internal/log` - Logging functions

## Profile Package

Package: `github.com/stwalsh4118/phanes/internal/profile`

Defines server profiles as lists of module names and provides lookup functionality. Profiles represent common server configurations that combine multiple modules to achieve specific server setups.

### Public Functions

```go
// GetProfile returns the list of module names for the specified profile.
// Returns an error if the profile does not exist.
func GetProfile(name string) ([]string, error)

// ListProfiles returns a sorted list of all available profile names.
func ListProfiles() []string

// ProfileExists checks if a profile with the given name exists.
// Returns true if the profile exists, false otherwise.
func ProfileExists(name string) bool
```

### Available Profiles

The following profiles are predefined:

- **minimal**: `baseline`, `user`, `security`, `swap`, `updates`
- **dev**: `baseline`, `user`, `security`, `swap`, `updates`, `docker`, `monitoring`, `devtools`
- **web**: `baseline`, `user`, `security`, `swap`, `updates`, `docker`, `monitoring`, `caddy`
- **database**: `baseline`, `user`, `security`, `swap`, `updates`, `docker`, `monitoring`, `postgres`, `redis`
- **coolify**: `baseline`, `user`, `security`, `swap`, `updates`, `docker`, `coolify`

### Usage Examples

```go
import "github.com/stwalsh4118/phanes/internal/profile"

// Get modules for a profile
modules, err := profile.GetProfile("dev")
if err != nil {
    log.Error("Profile not found: %v", err)
    return
}
// modules = ["baseline", "user", "security", "swap", "updates", "docker", "monitoring", "devtools"]

// List all available profiles
profiles := profile.ListProfiles()
// profiles = ["coolify", "database", "dev", "minimal", "web"]

// Check if a profile exists
if profile.ProfileExists("dev") {
    log.Info("Dev profile is available")
}
```

### Behavior

- **Profile Lookup**: `GetProfile()` returns a copy of the module list to prevent external modification
- **Sorted Results**: `ListProfiles()` returns profile names in alphabetical order
- **Case Sensitive**: Profile names are case-sensitive (e.g., "dev" != "DEV")
- **Error Handling**: `GetProfile()` returns an error with a clear message if the profile doesn't exist

### Dependencies

- Standard library only (`fmt`, `sort`)

