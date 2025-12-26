# Infrastructure API

Last Updated: 2025-01-27

**Note**: Updated with execution summary functionality (PBI 10). Runner now returns execution results and provides summary display.

**Note**: All packages now include comprehensive package-level documentation (doc.go files) and enhanced field-level documentation. See individual package documentation with `go doc` for complete details.

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
    DevTools  DevTools  `yaml:"devtools"`
    Coolify   Coolify   `yaml:"coolify"`
    Tailscale Tailscale `yaml:"tailscale"`
}

// User contains user-related configuration.
// Both fields are required for the user module to function.
type User struct {
    // Username is the Linux username to create on the server.
    Username string `yaml:"username"`
    // SSHPublicKey is the SSH public key to add to the user's authorized_keys file.
    SSHPublicKey string `yaml:"ssh_public_key"`
}

// System contains system-level configuration.
type System struct {
    // Timezone is the system timezone (e.g., "UTC", "America/New_York").
    Timezone string `yaml:"timezone"`
}

// Swap contains swap file configuration.
type Swap struct {
    // Enabled determines whether to create a swap file.
    Enabled bool `yaml:"enabled"`
    // Size is the swap file size (e.g., "1G", "2G", "4G").
    Size string `yaml:"size"`
}

// Security contains security-related configuration.
type Security struct {
    // SSHPort is the port number for SSH (1-65535).
    SSHPort int `yaml:"ssh_port"`
    // AllowPasswordAuth enables password authentication for SSH (not recommended).
    AllowPasswordAuth bool `yaml:"allow_password_auth"`
}

// Docker contains Docker-related configuration.
type Docker struct {
    // InstallCompose determines whether to install Docker Compose alongside Docker CE.
    InstallCompose bool `yaml:"install_compose"`
}

// Postgres contains PostgreSQL configuration.
type Postgres struct {
    // Enabled determines whether to install PostgreSQL.
    Enabled bool `yaml:"enabled"`
    // Version is the PostgreSQL version to install (e.g., "16", "15").
    Version string `yaml:"version"`
    // Password is the PostgreSQL superuser password (empty for no password).
    Password string `yaml:"password"`
    // Database is the initial database name to create.
    Database string `yaml:"database"`
    // User is the PostgreSQL user name.
    User string `yaml:"user"`
}

// Redis contains Redis configuration.
type Redis struct {
    // Enabled determines whether to install Redis.
    Enabled bool `yaml:"enabled"`
    // Password is the Redis password (empty for no password).
    Password string `yaml:"password"`
    // BindAddress is the IP address Redis should bind to (e.g., "127.0.0.1", "0.0.0.0").
    BindAddress string `yaml:"bind_address"`
}

// Nginx contains Nginx configuration.
type Nginx struct {
    // Enabled determines whether to install Nginx.
    Enabled bool `yaml:"enabled"`
}

// Caddy contains Caddy configuration.
type Caddy struct {
    // Enabled determines whether to install Caddy.
    Enabled bool `yaml:"enabled"`
}

// DevTools contains development tools configuration.
type DevTools struct {
    // Enabled determines whether to install development tools.
    Enabled bool `yaml:"enabled"`
    // NodeVersion is the Node.js version to install via nvm (e.g., "22", "20").
    NodeVersion string `yaml:"node_version"`
    // PythonVersion is the Python version to install (e.g., "3", "3.11").
    PythonVersion string `yaml:"python_version"`
    // GoVersion is the Go version to install (e.g., "1.25.5").
    GoVersion string `yaml:"go_version"`
    // InstallUv determines whether to install the uv package manager for Python.
    InstallUv bool `yaml:"install_uv"`
}

// Coolify contains Coolify configuration.
type Coolify struct {
    // Enabled determines whether to install Coolify (requires Docker).
    Enabled bool `yaml:"enabled"`
}

// Tailscale contains Tailscale VPN configuration.
type Tailscale struct {
    // Enabled determines whether to install and configure Tailscale.
    Enabled bool `yaml:"enabled"`
    // AuthKey is the Tailscale auth key for authentication (must start with "tskey-").
    // Required unless SkipAuth is true.
    AuthKey string `yaml:"auth_key"`
    // SkipAuth allows manual authentication after installation.
    // When true, the module will install Tailscale but skip automatic authentication,
    // allowing you to manually run "tailscale up" to authenticate via browser.
    SkipAuth bool `yaml:"skip_auth"`
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
- **Postgres.Enabled**: `true`
- **Postgres.Version**: `"16"`
- **Postgres.Database**: `"phanes"`
- **Postgres.User**: `"phanes"`
- **Redis.Enabled**: `true`
- **Redis.BindAddress**: `"127.0.0.1"`
- **DevTools.Enabled**: `true`
- **DevTools.NodeVersion**: `"22"`
- **DevTools.PythonVersion**: `"3"`
- **DevTools.GoVersion**: `"1.25.5"`
- **DevTools.InstallUv**: `true`
- **Nginx.Enabled**: `true`
- **Caddy.Enabled**: `true`
- **Coolify.Enabled**: `true`
- **Tailscale.Enabled**: `false`
- **Tailscale.AuthKey**: `""`
- **Tailscale.SkipAuth**: `false`

**Note on Module Enabled Defaults**: Modules default to `enabled: true` because when a user explicitly runs a module via `--modules`, they intend to install it. The `enabled` flag is primarily useful for disabling modules when included in profiles (e.g., `caddy: enabled: false` in a profile config to skip Caddy installation). Config files should focus on actual configuration values rather than acting as gates for module execution.

### Required Fields

- `user.username` - Username for the deployment user
- `user.ssh_public_key` - SSH public key for the deployment user

### Conditional Validation

- `tailscale.auth_key` - Required when `tailscale.enabled` is `true` and `tailscale.skip_auth` is `false`. Must start with `"tskey-"`

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

- **0**: Success (including when `--list` flag is used)
- **1**: General error (config errors, profile errors, module errors, execution failures, panics)
- **2**: Invalid usage (no profile/modules specified, usage errors)

### Error Handling

The CLI provides comprehensive error handling with clear, actionable error messages:

#### Error Types

```go
// usageError distinguishes usage errors (exit code 2) from general errors (exit code 1)
type usageError struct {
    message string
}
```

#### Error Handling Behavior

- **Panic Recovery**: Panic recovery is implemented in both `runCommand()` and `main()` using `defer` and `recover()`. Panics are logged and program exits with code 1.
- **Config Errors**:
  - **File not found**: Shows file path and suggests checking if file exists
  - **Invalid YAML**: Shows file path, error details, and suggests checking YAML syntax
  - **Validation errors**: Shows which field is missing and provides specific guidance (e.g., "Please add 'username' field under 'user' section")
- **Profile Errors**: Shows profile name that wasn't found, lists all available profiles, suggests using `--list` flag
- **Module Errors**:
  - **Unknown module**: Shows error details, lists available modules, suggests using `--list` flag
  - **Execution failures**: Shows which module failed with clear error messages
- **All Errors**: Use `log.Error()` for consistent error logging and include actionable suggestions

#### Error Message Format

All error messages:
- Use `log.Error()` for consistent error logging
- Include actionable suggestions (e.g., "Use --list to see all available modules")
- Show available options when applicable (profiles, modules)
- Provide specific guidance for fixing issues (e.g., which config field is missing)

### Behavior

- **Flag Parsing**: Uses Cobra's flag system (built on `spf13/pflag`)
- **Help Text**: Automatically generated by Cobra with examples and flag descriptions
- **Validation**: Requires either `--profile` or `--modules` to be specified (unless `--list` is used)
- **Dry-Run**: When `--dry-run` is set:
  - Calls `log.SetDryRun(true)` to enable dry-run mode in logging (all log entries include `dry_run=true` field)
  - Logs "Dry-run mode enabled. No changes will be made." message
  - Passes `dryRunFlag` to `executeModules()` which propagates to `runner.RunModules()`
  - Modules check `IsInstalled()` but don't execute `Install()` - only preview what would happen
  - Logs "Would install module {name} (dry-run)" for uninstalled modules
  - Logs "Module {name} is already installed (dry-run)" for installed modules
- **List Mode**: When `--list` is set, displays all available profiles and modules, then exits without executing
- **Error Handling**: Returns errors from `runCommand()` which are checked in `main()` for error type (usageError vs general error) and appropriate exit code is set
- **Version**: Built-in version support via Cobra's `Version` field

### Implementation Details

- **Command Execution**: `runCommand()` function handles all command logic with panic recovery
- **Flag Access**: Flags are accessed via package-level variables set by Cobra
- **Error Propagation**: Errors returned from `runCommand()` are checked in `main()` using `errors.As()` to determine error type (usageError vs general error) and appropriate exit code is set
- **Panic Recovery**: Panic recovery is implemented in both `runCommand()` (using `defer` and `recover()`) and `main()` as a safety net
- **Help Integration**: Cobra automatically shows help when `--help` is used or when validation fails
- **List Functionality**: `listProfilesAndModules()` function displays profiles and modules. Uses `registerAllModules()` helper to create runner and register all available modules for listing
- **Module Registration**: `registerAllModules()` creates a runner instance and registers all available modules (currently baseline and user). This function is reusable for both listing and execution
- **Error Handling Functions**:
  - `loadConfig()`: Handles config file loading with detailed error messages for file not found, invalid YAML, and validation errors
  - `getProfileModules()`: Handles profile selection with error messages showing available profiles
  - `executeModules()`: Handles module execution with error messages showing available modules when unknown modules are specified
- **Dry-Run Integration**: 
  - `runCommand()` sets dry-run mode via `log.SetDryRun(true)` when `--dry-run` flag is set
  - `executeModules()` receives `dryRun` parameter and passes it to `runner.RunModules()`
  - Dry-run mode is respected throughout the execution chain: CLI → executeModules() → runner.RunModules() → modules

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
    // Returns true if the module is already installed, false if it needs to be installed.
    // Returns an error if the check fails (e.g., permission issues, system errors).
    //
    // This method should perform a quick check to determine if the module's
    // configuration is already in place. For example:
    //   - Check if a package is installed
    //   - Check if a configuration file exists
    //   - Check if a service is configured
    //
    // The runner will call this before Install() to ensure idempotency.
    IsInstalled() (bool, error)

    // Install performs the installation and configuration of this module.
    // The cfg parameter provides access to all configuration values needed
    // for the installation process.
    //
    // This method should only be called when IsInstalled() returns false.
    // It should perform all necessary steps to install and configure the module,
    // such as:
    //   - Installing packages
    //   - Creating configuration files
    //   - Setting up services
    //   - Applying security settings
    //
    // Returns an error if installation fails. The runner will handle the error
    // and may stop execution or continue depending on error severity.
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

- **Idempotency**: All modules must be idempotent - safe to run multiple times. The runner calls `IsInstalled()` before `Install()`. If `IsInstalled()` returns `true`, `Install()` will not be called.
- **Error Handling**: Both `IsInstalled()` and `Install()` return errors that the runner handles appropriately. Errors from `IsInstalled()` cause the module to be skipped. Errors from `Install()` are collected and returned together.
- **Configuration**: Modules receive a `*config.Config` parameter in `Install()` to access all configuration values needed for installation.
- **Naming**: Module names must be unique and are used for registration and profile references. Names are case-sensitive.
- **IsInstalled() Implementation**: Should perform quick checks (e.g., file existence, package installation status, service configuration) without making changes to the system.
- **Install() Implementation**: Should only be called when `IsInstalled()` returns `false`. Must perform all necessary installation and configuration steps atomically where possible.

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

// ModuleStatus represents the execution status of a module.
type ModuleStatus string

const (
    // StatusInstalled indicates the module was successfully installed.
    StatusInstalled ModuleStatus = "installed"
    
    // StatusSkipped indicates the module was skipped because it was already installed.
    StatusSkipped ModuleStatus = "skipped"
    
    // StatusFailed indicates the module installation failed.
    StatusFailed ModuleStatus = "failed"
    
    // StatusError indicates an error occurred during module check or execution.
    StatusError ModuleStatus = "error"
    
    // StatusWouldInstall indicates the module would be installed in dry-run mode.
    // This status is only used when dry-run is enabled and the module is not currently installed.
    StatusWouldInstall ModuleStatus = "would_install"
)

// ModuleResult represents the execution result of a single module.
type ModuleResult struct {
    // Name is the unique name identifier of the module.
    Name string
    
    // Status indicates the execution outcome of the module.
    Status ModuleStatus
    
    // Error contains error details if Status is StatusFailed or StatusError.
    // This field is nil for successful or skipped modules.
    Error error
    
    // Duration is the time taken to execute the module (optional).
    // This field may be zero if duration tracking is not implemented.
    Duration time.Duration
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
// Returns a slice of ModuleResult for each module processed and an error if any module fails.
func (r *Runner) RunModules(names []string, cfg *config.Config, dryRun bool) ([]ModuleResult, error)

// PrintSummary displays a formatted summary table of module execution results.
// The table shows each module's name, status, and error details (if any).
// Status indicators are color-coded: green for installed/would install, yellow for skipped, red for failed/error.
// A summary line shows total counts for each status.
// If dryRun is true, a dry-run indicator is displayed.
func PrintSummary(results []ModuleResult, dryRun bool)

// GetModule returns a module from the registry by name.
// Returns nil if the module is not found.
func (r *Runner) GetModule(name string) module.Module

// ListModules returns a list of all registered module names.
// The order is not guaranteed (map iteration order). Use sort.Strings() if sorted order is needed.
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
results, err := r.RunModules(moduleNames, cfg, false)
if err != nil {
    log.Error("Failed to execute modules: %v", err)
}
// Print execution summary
runner.PrintSummary(results, false)

// Dry-run mode
results, err = r.RunModules(moduleNames, cfg, true)
if err != nil {
    log.Error("Dry-run failed: %v", err)
}
// Print dry-run summary
runner.PrintSummary(results, true)

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

- **Idempotency**: The runner checks `IsInstalled()` before calling `Install()` for each module. If a module is already installed, it is skipped with a log message using `log.Skip()`.
- **Error Handling**: Errors from `IsInstalled()` or `Install()` are logged using `log.Error()` and collected. The runner continues processing remaining modules even if one fails, but returns an error at the end if any modules failed.
- **Dry-Run Mode**: When `dryRun` is true, the runner checks `IsInstalled()` but does not call `Install()`. It logs what would happen without making changes. Uses `log.Skip()` for already-installed modules and `log.Info()` for modules that would be installed. Modules that would be installed are marked with `StatusWouldInstall` (not `StatusInstalled`) to distinguish them from actually installed modules.
- **Result Collection**: `RunModules()` returns a slice of `ModuleResult` for each module processed, allowing callers to display execution summaries. Results include status, error details (if any), and optional duration tracking.
- **Summary Display**: `PrintSummary()` provides a formatted table showing module execution results with color-coded status indicators and totals. Works in both normal and dry-run modes.
- **Module Registry**: Modules are registered by their name (from `Module.Name()`). Duplicate registrations overwrite the previous module with a warning logged using `log.Warn()`.
- **Order**: Modules are executed in the order specified in the `names` slice.
- **Unknown Modules**: If a module name is not found in the registry, an error is logged using `log.Error()` and the runner continues with the next module.
- **Logging During Execution**: 
  - `RegisterModule()` logs module registration using `log.Info()` with module name and description
  - `RunModules()` logs "Processing module: {name}" before each module
  - `RunModules()` logs "Installing module: {name}" before calling `Install()`
  - `RunModules()` logs "Successfully installed module: {name}" using `log.Success()` after successful installation
- **Module List Ordering**: `ListModules()` returns module names in an unsorted order (map iteration order). Use `sort.Strings()` if sorted order is needed.

### Error Handling

- **Empty Module List**: Returns an error if no modules are specified.
- **Unknown Module**: Returns an error if a module name is not found in the registry.
- **IsInstalled Error**: If `IsInstalled()` returns an error, it is logged and the module is skipped.
- **Install Error**: If `Install()` returns an error, it is logged and collected. The runner continues with remaining modules.
- **Multiple Errors**: If multiple modules fail, all errors are collected and returned together. Results are still returned even when errors occur, allowing summary display.

### Summary Display

The `PrintSummary()` function displays execution results in a formatted table:

- **Table Format**: Box-drawing characters with columns for Module, Status, and Details
- **Status Indicators**: 
  - `✓ Installed` (green) - Module successfully installed
  - `→ Would Install` (green) - Module would be installed in dry-run mode
  - `⊘ Skipped` (yellow) - Module already installed, skipped
  - `✗ Failed` (red) - Module installation failed
  - `✗ Error` (red) - Error during module check or execution
- **Error Details**: Failed/error modules show error messages in the Details column (truncated to 50 chars)
- **Totals**: Summary line shows counts: "X installed, Y would install, Z skipped, W failed"
- **Dry-Run Indicator**: Summary includes "(dry-run)" suffix when in dry-run mode

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

## Baseline Module

Package: `github.com/stwalsh4118/phanes/internal/modules/baseline`

Implements the baseline server configuration module that sets timezone, configures locale (UTF-8), and runs apt update. This is typically the first module executed in any profile.

### Public Types

```go
// BaselineModule implements the Module interface for baseline server configuration.
type BaselineModule struct{}
```

### Module Interface Implementation

```go
// Name returns "baseline"
func (m *BaselineModule) Name() string

// Description returns "Sets timezone, locale, and runs apt update"
func (m *BaselineModule) Description() string

// IsInstalled checks if baseline configuration is already applied.
// Verifies that a timezone is set (not empty) and locale contains UTF-8.
// Note: Since IsInstalled() doesn't receive config, it checks if
// timezone/locale are configured, not if they match a specific configured value.
func (m *BaselineModule) IsInstalled() (bool, error)

// Install configures timezone, locale, and runs apt update.
// Uses cfg.System.Timezone (defaults to "UTC" if empty).
// Sets locale to en_US.UTF-8 and runs apt-get update.
func (m *BaselineModule) Install(cfg *config.Config) error
```

### Usage Examples

```go
import (
    "github.com/stwalsh4118/phanes/internal/modules/baseline"
    "github.com/stwalsh4118/phanes/internal/config"
    "github.com/stwalsh4118/phanes/internal/runner"
)

// Create and register baseline module
mod := &baseline.BaselineModule{}
r := runner.NewRunner()
r.RegisterModule(mod)

// Load configuration
cfg, err := config.Load("config.yaml")
if err != nil {
    log.Error("Failed to load config: %v", err)
    return
}

// Check if already installed
installed, err := mod.IsInstalled()
if err != nil {
    log.Error("Failed to check installation status: %v", err)
    return
}

if !installed {
    // Install baseline configuration
    if err := mod.Install(cfg); err != nil {
        log.Error("Failed to install baseline: %v", err)
        return
    }
    log.Success("Baseline configuration installed")
} else {
    log.Skip("Baseline already configured")
}
```

### Configuration

The module uses the following configuration fields:

- `config.System.Timezone` - Timezone to set (e.g., "UTC", "America/New_York"). Defaults to "UTC" if empty.

### Behavior

- **Timezone**: Uses `timedatectl set-timezone` to set the system timezone. Verifies the timezone was set correctly after setting it.
- **Locale**: Generates and configures `en_US.UTF-8` locale using `locale-gen` and `update-locale`. Verifies UTF-8 is configured.
- **Apt Update**: Runs `apt-get update` to refresh package lists.
- **Idempotency**: `IsInstalled()` checks if timezone is set and locale contains UTF-8. If both are configured, the module is considered installed.
- **Error Handling**: All commands use the `exec` package and return descriptive errors if any step fails.
- **Logging**: Uses `log.Info()` for progress, `log.Success()` for completion, and `log.Error()` for errors.

### Commands Used

- `timedatectl show -p Timezone --value` - Check current timezone
- `timedatectl set-timezone <timezone>` - Set timezone
- `locale` - Check current locale settings
- `locale-gen en_US.UTF-8` - Generate UTF-8 locale
- `update-locale LANG=en_US.UTF-8` - Update locale configuration
- `apt-get update` - Update package lists

### Dependencies

- `github.com/stwalsh4118/phanes/internal/module` - Module interface
- `github.com/stwalsh4118/phanes/internal/config` - Configuration structure
- `github.com/stwalsh4118/phanes/internal/exec` - Command execution
- `github.com/stwalsh4118/phanes/internal/log` - Logging functions

## User Module

Package: `github.com/stwalsh4118/phanes/internal/modules/user`

Implements the user creation and SSH key setup module that creates a non-root user, sets up SSH key access, and configures passwordless sudo. This module ensures secure user access without root privileges.

### Public Types

```go
// UserModule implements the Module interface for user creation and SSH key setup.
type UserModule struct{}
```

### Module Interface Implementation

```go
// Name returns "user"
func (m *UserModule) Name() string

// Description returns "Creates user and sets up SSH keys"
func (m *UserModule) Description() string

// IsInstalled checks if the user module is already installed.
// Note: Since IsInstalled() doesn't receive config, it performs a generic check
// to see if the system appears to have been set up for user management.
// The Install() method performs the specific checks with config and is fully idempotent.
func (m *UserModule) IsInstalled() (bool, error)

// Install creates the user, sets up SSH keys, and configures passwordless sudo.
// This method is idempotent - it checks if each step is already done before doing it.
// Validates username and SSH key format before proceeding.
func (m *UserModule) Install(cfg *config.Config) error
```

### Usage Examples

```go
import (
    "github.com/stwalsh4118/phanes/internal/modules/user"
    "github.com/stwalsh4118/phanes/internal/config"
    "github.com/stwalsh4118/phanes/internal/runner"
)

// Create and register user module
mod := &user.UserModule{}
r := runner.NewRunner()
r.RegisterModule(mod)

// Load configuration
cfg, err := config.Load("config.yaml")
if err != nil {
    log.Error("Failed to load config: %v", err)
    return
}

// Check if already installed
installed, err := mod.IsInstalled()
if err != nil {
    log.Error("Failed to check installation status: %v", err)
    return
}

if !installed {
    // Install user configuration
    if err := mod.Install(cfg); err != nil {
        log.Error("Failed to install user: %v", err)
        return
    }
    log.Success("User module installation completed")
} else {
    log.Skip("User module already configured")
}
```

### Configuration

The module uses the following configuration fields:

- `config.User.Username` - Username to create (required)
- `config.User.SSHPublicKey` - SSH public key to add to authorized_keys (required)

### Behavior

- **User Creation**: Creates user with home directory using `useradd -m -s /bin/bash <username>`. Handles case where user already exists gracefully.
- **SSH Directory**: Creates `~/.ssh` directory with permissions 700 if it doesn't exist. Sets ownership to the user (required by OpenSSH StrictModes).
- **SSH Key**: Adds SSH public key to `~/.ssh/authorized_keys` with permissions 600. Sets ownership to the user (required by OpenSSH StrictModes). Validates SSH key format (ssh-rsa, ssh-ed25519, ecdsa-sha2-*, ssh-dss). Skips if key already exists.
- **Sudoers**: Creates `/etc/sudoers.d/<username>` file with passwordless sudo configuration. Validates sudoers file with `visudo -c` before applying. Sets permissions 0440.
- **Idempotency**: `Install()` is fully idempotent - checks if user exists, if SSH key is already in authorized_keys, and if sudoers file is correctly configured before making changes.
- **Error Handling**: Validates username and SSH key format before proceeding. Returns descriptive errors if any step fails.
- **Dry-Run Support**: Checks dry-run mode using `log.IsDryRun()` and logs what would be done without executing commands or creating files.
- **Logging**: Uses `log.Info()` for progress, `log.Success()` for completion, `log.Skip()` for already-configured items, and `log.Error()` for errors.

### SSH Key Validation

The module validates SSH public key format. Valid formats include:
- `ssh-rsa` - RSA keys
- `ssh-ed25519` - Ed25519 keys
- `ecdsa-sha2-*` - ECDSA keys (any variant)
- `ssh-dss` - DSA keys

### Commands Used

- `useradd -m -s /bin/bash <username>` - Create user with home directory
- `id <username>` - Check if user exists (via user.Lookup)
- `visudo -c -f <file>` - Validate sudoers file

### File Operations

- Creates `~/.ssh` directory with permissions 700 and ownership set to the user
- Creates/updates `~/.ssh/authorized_keys` with permissions 600 and ownership set to the user
- Creates `/etc/sudoers.d/<username>` with permissions 0440

### Ownership Requirements

**Critical**: When running as root, the module sets correct ownership on `.ssh` directory and `authorized_keys` file using `os.Chown()` with the user's UID/GID. This is required because:
- OpenSSH's StrictModes (enabled by default) requires these files to be owned by the user
- Files created by root will have `root:root` ownership by default
- SSH key authentication will silently fail if ownership is incorrect
- The module looks up the user's UID/GID after user creation and applies ownership to all SSH-related files

### Dependencies

- `github.com/stwalsh4118/phanes/internal/module` - Module interface
- `github.com/stwalsh4118/phanes/internal/config` - Configuration structure
- `github.com/stwalsh4118/phanes/internal/exec` - Command execution and file operations
- `github.com/stwalsh4118/phanes/internal/log` - Logging functions
- `os` - File operations
- `os/user` - User lookup
- `path/filepath` - Path operations

## Security Module

Package: `github.com/stwalsh4118/phanes/internal/modules/security`

Implements the security hardening module that configures UFW firewall, installs and configures fail2ban, and hardens SSH configuration. Uses embedded templates for SSH and fail2ban configs.

### Public Types

```go
// SecurityModule implements the Module interface for security configuration.
type SecurityModule struct{}
```

### Module Interface Implementation

```go
// Name returns "security"
func (m *SecurityModule) Name() string

// Description returns "Configures UFW, fail2ban, and SSH hardening"
func (m *SecurityModule) Description() string

// IsInstalled checks if security configuration is already applied.
// Verifies that UFW is enabled, fail2ban is running, and SSH config has key security settings.
// Note: Since IsInstalled() doesn't receive config, it performs generic checks.
// Install() performs specific checks with config and is fully idempotent.
func (m *SecurityModule) IsInstalled() (bool, error)

// Install configures UFW firewall, fail2ban, and SSH hardening.
// Uses cfg.Security.SSHPort (defaults to 22) and cfg.Security.AllowPasswordAuth (defaults to false).
func (m *SecurityModule) Install(cfg *config.Config) error
```

### Usage Examples

```go
import (
    "github.com/stwalsh4118/phanes/internal/modules/security"
    "github.com/stwalsh4118/phanes/internal/config"
    "github.com/stwalsh4118/phanes/internal/runner"
)

// Create and register security module
mod := &security.SecurityModule{}
r := runner.NewRunner()
r.RegisterModule(mod)

// Load configuration
cfg, err := config.Load("config.yaml")
if err != nil {
    log.Error("Failed to load config: %v", err)
    return
}

// Check if already installed
installed, err := mod.IsInstalled()
if err != nil {
    log.Error("Failed to check installation status: %v", err)
    return
}

if !installed {
    // Install security configuration
    if err := mod.Install(cfg); err != nil {
        log.Error("Failed to install security: %v", err)
        return
    }
    log.Success("Security configuration installed")
} else {
    log.Skip("Security already configured")
}
```

### Configuration

The module uses the following configuration fields:

- `config.Security.SSHPort` - SSH port to configure (defaults to 22)
- `config.Security.AllowPasswordAuth` - Whether to allow password authentication (defaults to false)

### Behavior

- **UFW Firewall**: Installs UFW if not installed, allows SSH port (from config), HTTP (80), and HTTPS (443), then enables UFW. Verifies UFW is active after enabling.
- **Fail2ban**: Installs fail2ban if not installed, creates `/etc/fail2ban/jail.local` from embedded template with SSH jail configuration, starts and enables fail2ban service. Verifies fail2ban is running after starting.
- **SSH Hardening**: Backs up existing `/etc/ssh/sshd_config`, creates new config from embedded template with security settings (disables root login, configures password auth based on config, enables pubkey auth, etc.), validates config with `sshd -t`, and reloads SSH service. Warns user if password auth is being disabled.
- **Idempotency**: `IsInstalled()` checks if UFW is enabled, fail2ban is running, and SSH config has key security settings. `Install()` is fully idempotent - checks if each component is already configured before making changes.
- **Error Handling**: Validates SSH port (must be between 1 and 65535). Returns descriptive errors if any step fails. SSH config validation failures prevent invalid config from being applied (restores backup if validation fails).
- **Dry-Run Support**: Checks dry-run mode using `log.IsDryRun()` and logs what would be done without executing commands or writing files. Still validates SSH config in dry-run mode.
- **Logging**: Uses `log.Info()` for progress, `log.Success()` for completion, `log.Skip()` for already-configured items, `log.Warn()` when disabling password authentication, and `log.Error()` for errors.

### Embedded Templates

The module uses embedded templates for configuration files. Templates are located in `internal/modules/security/` (same directory as the module) because `go:embed` doesn't support `..` paths.

- `internal/modules/security/sshd_config.tmpl` - SSH server configuration template
  - Variables: `{{.SSHPort}}`, `{{.AllowPasswordAuth}}`
  - Includes security hardening settings (no root login, password auth configurable, pubkey auth enabled, etc.)
- `internal/modules/security/jail.local.tmpl` - Fail2ban jail configuration template
  - Variables: `{{.SSHPort}}`
  - Configures SSH jail with ban time, find time, and max retries

### Commands Used

- `apt-get install -y ufw` - Install UFW firewall
- `ufw allow <port>/tcp` - Allow port in UFW
- `ufw --force enable` - Enable UFW firewall
- `ufw status` - Check UFW status
- `apt-get install -y fail2ban` - Install fail2ban
- `systemctl enable --now fail2ban` - Start and enable fail2ban service
- `systemctl is-active fail2ban` - Check fail2ban status
- `sshd -t` - Validate SSH configuration
- `systemctl reload sshd` or `systemctl reload ssh` - Reload SSH service

### File Operations

- Creates `/etc/fail2ban/jail.local` with permissions 0644
- Creates `/etc/ssh/sshd_config` with permissions 0644
- Backs up existing `/etc/ssh/sshd_config` to `/etc/ssh/sshd_config.backup` before modification

### Security Considerations

- **Password Authentication**: When `AllowPasswordAuth` is false, the module warns the user before disabling password authentication. Users must ensure SSH key access is configured before disabling password auth.
- **SSH Config Validation**: The module validates SSH configuration with `sshd -t` before applying changes. If validation fails, the backup is restored to prevent locking users out.
- **UFW Rules**: The module allows SSH port (from config), HTTP (80), and HTTPS (443). Additional ports must be configured manually or via other modules.
- **Fail2ban Configuration**: The module configures fail2ban with reasonable defaults (1 hour ban time, 10 minute find time, 5 max retries). These can be customized by editing `/etc/fail2ban/jail.local` after installation.

### Dependencies

- `github.com/stwalsh4118/phanes/internal/module` - Module interface
- `github.com/stwalsh4118/phanes/internal/config` - Configuration structure
- `github.com/stwalsh4118/phanes/internal/exec` - Command execution and file operations
- `github.com/stwalsh4118/phanes/internal/log` - Logging functions
- `embed` - Template embedding
- `text/template` - Template rendering
- `os` - File operations
- `bytes` - Template output buffering

## Swap Module

Package: `github.com/stwalsh4118/phanes/internal/modules/swap`

Implements the swap file creation and configuration module that creates a swap file, configures it in `/etc/fstab` for persistence, and sets swappiness. This helps servers handle memory pressure gracefully.

### Public Types

```go
// SwapModule implements the Module interface for swap file creation and configuration.
type SwapModule struct{}
```

### Module Interface Implementation

```go
// Name returns "swap"
func (m *SwapModule) Name() string

// Description returns "Creates and configures swap file"
func (m *SwapModule) Description() string

// IsInstalled checks if swap configuration is already applied.
// Verifies that swap is active, swap file exists, fstab contains swap entry, and swappiness is set.
// Note: Since IsInstalled() doesn't receive config, it performs generic checks.
// Install() performs specific checks with config and is fully idempotent.
func (m *SwapModule) IsInstalled() (bool, error)

// Install creates and configures the swap file.
// Uses cfg.Swap.Enabled (defaults to true) and cfg.Swap.Size (defaults to "2G").
func (m *SwapModule) Install(cfg *config.Config) error
```

### Usage Examples

```go
import (
    "github.com/stwalsh4118/phanes/internal/modules/swap"
    "github.com/stwalsh4118/phanes/internal/config"
    "github.com/stwalsh4118/phanes/internal/runner"
)

// Create and register swap module
mod := &swap.SwapModule{}
r := runner.NewRunner()
r.RegisterModule(mod)

// Load configuration
cfg, err := config.Load("config.yaml")
if err != nil {
    log.Error("Failed to load config: %v", err)
    return
}

// Check if already installed
installed, err := mod.IsInstalled()
if err != nil {
    log.Error("Failed to check installation status: %v", err)
    return
}

if !installed {
    // Install swap configuration
    if err := mod.Install(cfg); err != nil {
        log.Error("Failed to install swap: %v", err)
        return
    }
    log.Success("Swap configuration installed")
} else {
    log.Skip("Swap already configured")
}
```

### Configuration

The module uses the following configuration fields:

- `config.Swap.Enabled` - Whether to create swap (defaults to `true`)
- `config.Swap.Size` - Size of swap file in format "2G", "512M", "1T" (defaults to `"2G"`)

### Behavior

- **Swap File Creation**: Creates swap file at `/swapfile` using `fallocate` (preferred) or `dd` as fallback. Sets permissions to 600, formats with `mkswap`, and enables with `swapon`. Shows progress logging for large swap files.
- **Fstab Configuration**: Adds swap entry to `/etc/fstab` for persistence: `/swapfile none swap sw 0 0`. Checks if entry already exists before adding.
- **Swappiness**: Sets swappiness to 10 (server-optimized value) both at runtime (`sysctl vm.swappiness=10`) and persistently via `/etc/sysctl.d/99-swappiness.conf`.
- **Idempotency**: `IsInstalled()` checks if swap is active, swap file exists, fstab contains swap entry, and swappiness is set. `Install()` is fully idempotent - checks if each component is already configured before making changes.
- **Error Handling**: Validates swap size format before proceeding. Returns descriptive errors if swap file creation fails (disk space, permissions), fstab write fails, or swappiness setting fails.
- **Dry-Run Support**: Checks dry-run mode using `log.IsDryRun()` and logs what would be done without executing commands or writing files. Still performs checks (swap exists, etc.).
- **Swap Disabled**: If `cfg.Swap.Enabled` is false, logs skip message and returns without creating swap. Still checks/configures if swap already exists.
- **Logging**: Uses `log.Info()` for progress messages (especially for large swap file creation), `log.Success()` for completion, `log.Skip()` for already-configured items, and `log.Error()` for errors.

### Size Format

Swap size supports the following formats (case-insensitive):
- `M` or `m` - Megabytes (multiply by 1024^2)
- `G` or `g` - Gigabytes (multiply by 1024^3)
- `T` or `t` - Terabytes (multiply by 1024^4)

Examples: `"2G"`, `"512M"`, `"1T"`, `"1.5G"`

### Commands Used

- `fallocate -l <size> <file>` - Create swap file (preferred method)
- `dd if=/dev/zero of=<file> bs=1M count=<size>` - Create swap file (fallback)
- `chmod 600 <file>` - Set swap file permissions
- `mkswap <file>` - Format file as swap
- `swapon <file>` - Enable swap
- `swapon --show` - Check active swap
- `sysctl vm.swappiness=<value>` - Set swappiness runtime value

### File Operations

- Creates `/swapfile` with permissions 600
- Appends to `/etc/fstab` with swap entry: `/swapfile none swap sw 0 0`
- Creates `/etc/sysctl.d/99-swappiness.conf` with: `vm.swappiness=10`

### Dependencies

- `github.com/stwalsh4118/phanes/internal/module` - Module interface
- `github.com/stwalsh4118/phanes/internal/config` - Configuration structure
- `github.com/stwalsh4118/phanes/internal/exec` - Command execution and file operations
- `github.com/stwalsh4118/phanes/internal/log` - Logging functions
- `os` - File operations
- `bufio` - Reading fstab
- `strconv` - Parsing size strings
- `strings` - String operations

## Updates Module

Package: `github.com/stwalsh4118/phanes/internal/modules/updates`

Implements the automatic security updates configuration module that installs and configures `unattended-upgrades` for automatic security updates. This ensures servers stay patched with security updates automatically.

### Public Types

```go
// UpdatesModule implements the Module interface for automatic security updates configuration.
type UpdatesModule struct{}
```

### Module Interface Implementation

```go
// Name returns "updates"
func (m *UpdatesModule) Name() string

// Description returns "Configures automatic security updates"
func (m *UpdatesModule) Description() string

// IsInstalled checks if updates configuration is already applied.
// Verifies that unattended-upgrades package is installed and configuration files are correctly set.
// Note: Since IsInstalled() doesn't receive config, it performs generic checks.
// Install() performs specific checks with config and is fully idempotent.
func (m *UpdatesModule) IsInstalled() (bool, error)

// Install installs and configures unattended-upgrades for automatic security updates.
// No config fields are required - uses sensible defaults.
func (m *UpdatesModule) Install(cfg *config.Config) error
```

### Usage Examples

```go
import (
    "github.com/stwalsh4118/phanes/internal/modules/updates"
    "github.com/stwalsh4118/phanes/internal/config"
    "github.com/stwalsh4118/phanes/internal/runner"
)

// Create and register updates module
mod := &updates.UpdatesModule{}
r := runner.NewRunner()
r.RegisterModule(mod)

// Load configuration
cfg, err := config.Load("config.yaml")
if err != nil {
    log.Error("Failed to load config: %v", err)
    return
}

// Check if already installed
installed, err := mod.IsInstalled()
if err != nil {
    log.Error("Failed to check installation status: %v", err)
    return
}

if !installed {
    // Install updates configuration
    if err := mod.Install(cfg); err != nil {
        log.Error("Failed to install updates: %v", err)
        return
    }
    log.Success("Updates configuration installed")
} else {
    log.Skip("Updates already configured")
}
```

### Configuration

The module does not require any configuration fields. It uses sensible defaults:
- Automatic security updates enabled
- Automatic reboot disabled (per PRD)
- Auto-remove unused dependencies enabled
- Daily update checks

### Behavior

- **Package Installation**: Installs `unattended-upgrades` package using `apt-get install -y unattended-upgrades`. Checks if package is already installed before attempting installation.
- **50unattended-upgrades Configuration**: Creates `/etc/apt/apt.conf.d/50unattended-upgrades` with security update origins enabled, auto-remove unused dependencies enabled, and automatic reboot disabled. Configures security update origins for the distribution.
- **20auto-upgrades Configuration**: Creates `/etc/apt/apt.conf.d/20auto-upgrades` with daily automatic updates enabled (`APT::Periodic::Update-Package-Lists "1"` and `APT::Periodic::Unattended-Upgrade "1"`).
- **Configuration Verification**: Optionally runs `unattended-upgrades --dry-run --debug` to verify configuration after installation.
- **Idempotency**: `IsInstalled()` checks if package is installed and config files match expected content. `Install()` is fully idempotent - checks if each component is already configured before making changes.
- **Error Handling**: Returns descriptive errors if package installation fails, config file write fails, or verification fails.
- **Dry-Run Support**: Checks dry-run mode using `log.IsDryRun()` and logs what would be done without executing commands or writing files.
- **Logging**: Uses `log.Info()` for progress messages, `log.Success()` for completion, `log.Skip()` for already-configured items, and `log.Error()` for errors. Informs user that automatic reboot is disabled by default.

### Configuration Files

- **`/etc/apt/apt.conf.d/50unattended-upgrades`**: Configures what gets updated and how
  - Security update origins enabled
  - Auto-remove unused dependencies enabled
  - Automatic reboot disabled (`Unattended-Upgrade::Automatic-Reboot "false"`)
- **`/etc/apt/apt.conf.d/20auto-upgrades`**: Enables automatic updates
  - `APT::Periodic::Update-Package-Lists "1"` - Daily package list updates
  - `APT::Periodic::Unattended-Upgrade "1"` - Daily automatic upgrades
  - `APT::Periodic::Download-Upgradeable-Packages "1"` - Download upgradeable packages
  - `APT::Periodic::AutocleanInterval "7"` - Weekly autoclean

### Commands Used

- `apt-get install -y unattended-upgrades` - Install unattended-upgrades package
- `dpkg -l unattended-upgrades` - Check if package is installed
- `unattended-upgrades --dry-run --debug` - Verify configuration (optional)

### File Operations

- Creates `/etc/apt/apt.conf.d/50unattended-upgrades` with permissions 0644
- Creates `/etc/apt/apt.conf.d/20auto-upgrades` with permissions 0644

### Dependencies

- `github.com/stwalsh4118/phanes/internal/module` - Module interface
- `github.com/stwalsh4118/phanes/internal/config` - Configuration structure
- `github.com/stwalsh4118/phanes/internal/exec` - Command execution and file operations
- `github.com/stwalsh4118/phanes/internal/log` - Logging functions
- `os` - File operations
- `bufio` - Reading config files
- `strings` - String operations

## Docker Module

Package: `github.com/stwalsh4118/phanes/internal/modules/docker`

Implements the Docker CE and Docker Compose v2 installation module that installs Docker from the official Docker repository, configures the Docker service, and adds the configured user to the docker group. This enables containerized application deployment.

### Public Types

```go
// DockerModule implements the Module interface for Docker CE and Docker Compose installation.
type DockerModule struct{}
```

### Module Interface Implementation

```go
// Name returns "docker"
func (m *DockerModule) Name() string

// Description returns "Installs Docker CE and Docker Compose"
func (m *DockerModule) Description() string

// IsInstalled checks if Docker is already installed and configured.
// Verifies that Docker is installed, Docker service is running, and Docker Compose is installed.
// Note: Since IsInstalled() doesn't receive config, it performs generic checks.
// Install() performs specific checks with config and is fully idempotent.
func (m *DockerModule) IsInstalled() (bool, error)

// Install installs Docker CE and Docker Compose v2, and adds the user to the docker group.
// Uses cfg.User.Username (required) and cfg.Docker.InstallCompose (defaults to true).
func (m *DockerModule) Install(cfg *config.Config) error
```

### Usage Examples

```go
import (
    "github.com/stwalsh4118/phanes/internal/modules/docker"
    "github.com/stwalsh4118/phanes/internal/config"
    "github.com/stwalsh4118/phanes/internal/runner"
)

// Create and register docker module
mod := &docker.DockerModule{}
r := runner.NewRunner()
r.RegisterModule(mod)

// Load configuration
cfg, err := config.Load("config.yaml")
if err != nil {
    log.Error("Failed to load config: %v", err)
    return
}

// Check if already installed
installed, err := mod.IsInstalled()
if err != nil {
    log.Error("Failed to check installation status: %v", err)
    return
}

if !installed {
    // Install Docker
    if err := mod.Install(cfg); err != nil {
        log.Error("Failed to install Docker: %v", err)
        return
    }
    log.Success("Docker installation completed")
} else {
    log.Skip("Docker already installed")
}
```

### Configuration

The module uses the following configuration fields:

- `config.User.Username` - Username to add to docker group (required)
- `config.Docker.InstallCompose` - Whether to verify Docker Compose installation (defaults to `true`)

### Behavior

- **Prerequisites**: Installs `ca-certificates` and `curl` packages before adding Docker repository.
- **GPG Key**: Downloads Docker's official GPG key from `https://download.docker.com/linux/ubuntu/gpg` and adds it to `/usr/share/keyrings/docker-archive-keyring.gpg`.
- **Repository Setup**: Detects distribution codename using `lsb_release -cs` or `/etc/os-release`, gets system architecture, and adds Docker repository to `/etc/apt/sources.list.d/docker.list`.
- **Package Installation**: Installs Docker CE packages: `docker-ce`, `docker-ce-cli`, `containerd.io`, `docker-buildx-plugin`, `docker-compose-plugin`.
- **Service Configuration**: Enables and starts Docker service using `systemctl enable --now docker`. Verifies service is running after start.
- **Docker Compose**: Verifies Docker Compose v2 installation if `cfg.Docker.InstallCompose` is true. Docker Compose v2 is installed as part of `docker-compose-plugin` package.
- **User Group**: Checks if user exists before attempting to add to docker group. If user doesn't exist, logs a warning and skips group membership (allows Docker module to run independently). If user exists, adds user to docker group using `usermod -aG docker <username>`. Warns user that logout/login is required for group changes to take effect.
- **Idempotency**: `IsInstalled()` checks if Docker is installed, service is running, and Docker Compose is installed. `Install()` is fully idempotent - checks if each component is already configured before making changes.
- **Error Handling**: Validates username is set before proceeding. Returns descriptive errors if GPG key download fails, repository addition fails, package installation fails, service start fails, or user group addition fails.
- **Dry-Run Support**: Checks dry-run mode using `log.IsDryRun()` and logs what would be done without executing commands or writing files. Still performs checks (Docker installed, etc.).
- **Logging**: Uses `log.Info()` for progress messages (especially during installation), `log.Success()` for completion, `log.Skip()` for already-configured items, `log.Warn()` when user needs to logout/login for docker group changes, and `log.Error()` for errors.

### Distribution Codename Detection

The module detects the distribution codename using:
1. `lsb_release -cs` (preferred method)
2. `/etc/os-release` file (fallback) - reads `VERSION_CODENAME` or maps `VERSION_ID` to codename

Supported Ubuntu versions: 22.04 (jammy), 20.04 (focal), 18.04 (bionic), 16.04 (xenial)

### Commands Used

- `apt-get update` - Update package lists
- `apt-get install -y ca-certificates curl` - Install prerequisites
- `curl -fsSL <url>` - Download GPG key
- `gpg --dearmor` - Convert GPG key to keyring format
- `lsb_release -cs` - Get distribution codename
- `dpkg --print-architecture` - Get system architecture
- `apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin` - Install Docker packages
- `docker --version` - Verify Docker installation
- `systemctl enable --now docker` - Enable and start Docker service
- `systemctl is-active docker` - Check Docker service status
- `docker compose version` - Verify Docker Compose installation
- `id <username>` - Check if user exists
- `id -nG <username>` - Check user groups
- `usermod -aG docker <username>` - Add user to docker group

### File Operations

- Creates `/usr/share/keyrings/docker-archive-keyring.gpg` with GPG keyring
- Creates `/etc/apt/sources.list.d/docker.list` with Docker repository entry

### Dependencies

- `github.com/stwalsh4118/phanes/internal/module` - Module interface
- `github.com/stwalsh4118/phanes/internal/config` - Configuration structure
- `github.com/stwalsh4118/phanes/internal/exec` - Command execution and file operations
- `github.com/stwalsh4118/phanes/internal/log` - Logging functions
- `os` - File operations
- `bufio` - Reading /etc/os-release
- `strings` - String operations

## Monitoring Module

Package: `github.com/stwalsh4118/phanes/internal/modules/monitoring`

Implements the Netdata monitoring installation module that installs and configures Netdata using the official kickstart script. Netdata provides real-time server monitoring and performance metrics accessible via web UI on port 19999.

### Public Types

```go
// MonitoringModule implements the Module interface for Netdata monitoring installation.
type MonitoringModule struct{}
```

### Module Interface Implementation

```go
// Name returns "monitoring"
func (m *MonitoringModule) Name() string

// Description returns "Installs and configures Netdata monitoring"
func (m *MonitoringModule) Description() string

// IsInstalled checks if Netdata is already installed and configured.
// Verifies that Netdata is installed, service is running, and port 19999 is accessible.
// Note: Since IsInstalled() doesn't receive config, it performs generic checks.
// Install() performs specific checks with config and is fully idempotent.
func (m *MonitoringModule) IsInstalled() (bool, error)

// Install installs and configures Netdata using the official kickstart script.
// No config fields are required - uses Netdata defaults.
func (m *MonitoringModule) Install(cfg *config.Config) error
```

### Usage Examples

```go
import (
    "github.com/stwalsh4118/phanes/internal/modules/monitoring"
    "github.com/stwalsh4118/phanes/internal/config"
    "github.com/stwalsh4118/phanes/internal/runner"
)

// Create and register monitoring module
mod := &monitoring.MonitoringModule{}
r := runner.NewRunner()
r.RegisterModule(mod)

// Load configuration
cfg, err := config.Load("config.yaml")
if err != nil {
    log.Error("Failed to load config: %v", err)
    return
}

// Check if already installed
installed, err := mod.IsInstalled()
if err != nil {
    log.Error("Failed to check installation status: %v", err)
    return
}

if !installed {
    // Install Netdata
    if err := mod.Install(cfg); err != nil {
        log.Error("Failed to install Netdata: %v", err)
        return
    }
    log.Success("Netdata installation completed")
} else {
    log.Skip("Netdata already installed")
}
```

### Configuration

The module does not require any configuration fields. It uses Netdata defaults:
- Default port: 19999
- Default installation via kickstart script
- Service auto-starts on boot

### Behavior

- **Kickstart Script**: Downloads the official Netdata kickstart script from `https://get.netdata.cloud/kickstart.sh` and runs it with `--non-interactive` flag. The script handles all installation, configuration, and service setup automatically.
- **Service Configuration**: Enables Netdata service to start on boot using `systemctl enable netdata`. Starts the service if not running using `systemctl start netdata`. Verifies service is running after start.
- **Port Verification**: Checks if Netdata is listening on port 19999 using `ss -tlnp` (or `netstat -tlnp` as fallback). Logs access URL: `http://localhost:19999`.
- **Idempotency**: `IsInstalled()` checks if Netdata is installed, service is running, and port is accessible. `Install()` is fully idempotent - checks if each component is already configured before making changes.
- **Error Handling**: Returns descriptive errors if kickstart script download fails, script execution fails, service start/enable fails, or port check fails.
- **Dry-Run Support**: Checks dry-run mode using `log.IsDryRun()` and logs what would be done without executing commands or downloading files. Still performs checks (Netdata installed, etc.).
- **Logging**: Uses `log.Info()` for progress messages (especially during kickstart script execution), `log.Success()` for completion, `log.Skip()` for already-configured items, `log.Warn()` if port is not yet accessible (service may still be starting), and `log.Error()` for errors.

### Commands Used

- `curl -fsSL https://get.netdata.cloud/kickstart.sh -o /tmp/netdata-kickstart.sh` - Download kickstart script
- `chmod +x /tmp/netdata-kickstart.sh` - Make script executable
- `bash /tmp/netdata-kickstart.sh --non-interactive` - Run kickstart script
- `systemctl enable netdata` - Enable Netdata service
- `systemctl start netdata` - Start Netdata service
- `systemctl is-active netdata` - Check Netdata service status
- `systemctl is-enabled netdata` - Check if Netdata service is enabled
- `ss -tlnp` or `netstat -tlnp` - Check if port 19999 is listening

### File Operations

- Downloads `/tmp/netdata-kickstart.sh` (temporary file, cleaned up after installation)
- Checks for `/usr/sbin/netdata` binary to verify installation

### Dependencies

- `github.com/stwalsh4118/phanes/internal/module` - Module interface
- `github.com/stwalsh4118/phanes/internal/config` - Configuration structure
- `github.com/stwalsh4118/phanes/internal/exec` - Command execution and file operations
- `github.com/stwalsh4118/phanes/internal/log` - Logging functions
- `os` - File operations
- `strings` - String operations

## Nginx Module

Package: `github.com/stwalsh4118/phanes/internal/modules/nginx`

Implements the Nginx web server installation module that installs Nginx via apt, configures the service, and verifies it is running and accessible on port 80. This provides HTTP/HTTPS serving capabilities for web applications.

### Public Types

```go
// NginxModule implements the Module interface for Nginx web server installation.
type NginxModule struct{}
```

### Module Interface Implementation

```go
// Name returns "nginx"
func (m *NginxModule) Name() string

// Description returns "Installs and configures Nginx web server"
func (m *NginxModule) Description() string

// IsInstalled checks if Nginx is already installed and configured.
// Verifies that Nginx is installed, service is running, and port 80 is accessible.
// Note: Since IsInstalled() doesn't receive config, it performs generic checks.
// Install() performs specific checks with config and is fully idempotent.
func (m *NginxModule) IsInstalled() (bool, error)

// Install installs and configures Nginx web server.
// Uses cfg.Nginx.Enabled (defaults to true) - skips installation if disabled.
func (m *NginxModule) Install(cfg *config.Config) error
```

### Usage Examples

```go
import (
    "github.com/stwalsh4118/phanes/internal/modules/nginx"
    "github.com/stwalsh4118/phanes/internal/config"
    "github.com/stwalsh4118/phanes/internal/runner"
)

// Create and register nginx module
mod := &nginx.NginxModule{}
r := runner.NewRunner()
r.RegisterModule(mod)

// Load configuration
cfg, err := config.Load("config.yaml")
if err != nil {
    log.Error("Failed to load config: %v", err)
    return
}

// Check if already installed
installed, err := mod.IsInstalled()
if err != nil {
    log.Error("Failed to check installation status: %v", err)
    return
}

if !installed {
    // Install Nginx
    if err := mod.Install(cfg); err != nil {
        log.Error("Failed to install Nginx: %v", err)
        return
    }
    log.Success("Nginx installation completed")
} else {
    log.Skip("Nginx already installed")
}
```

### Configuration

The module uses the following configuration fields:

- `config.Nginx.Enabled` - Whether to install Nginx (defaults to `true`)

### Behavior

- **Package Installation**: Installs Nginx via apt (`apt-get install -y nginx`). Updates package list first, then installs nginx package, and verifies installation with `nginx -v`.
- **Service Configuration**: Enables Nginx service to start on boot using `systemctl enable nginx`. Starts the service if not running using `systemctl start nginx`. Verifies service is running after start.
- **Port Verification**: Checks if Nginx is listening on port 80 using `ss -tlnp` (or `netstat -tlnp` as fallback). Logs access URL: `http://localhost`.
- **Port Conflict Detection**: Before installation, checks if port 80 is already in use by another service. If so, logs a warning but proceeds with installation (user may have configured it intentionally).
- **Idempotency**: `IsInstalled()` checks if Nginx is installed, service is running, and port is accessible. `Install()` is fully idempotent - checks if each component is already configured before making changes.
- **Error Handling**: Returns descriptive errors if apt update fails, nginx installation fails, service start/enable fails, or port check fails.
- **Dry-Run Support**: Checks dry-run mode using `log.IsDryRun()` and logs what would be done without executing commands. Still performs checks (Nginx installed, etc.).
- **Configuration Flag**: Respects `cfg.Nginx.Enabled` flag. If `Enabled` is `false`, skips installation and logs skip message. If `Enabled` is `true` (default), proceeds with installation. Set to `false` in config to disable the module when included in a profile.
- **Logging**: Uses `log.Info()` for progress messages (especially during installation), `log.Success()` for completion, `log.Skip()` for already-configured items, `log.Warn()` when port conflicts are detected, and `log.Error()` for errors.

### Commands Used

- `apt-get update` - Update package list
- `apt-get install -y nginx` - Install nginx package
- `nginx -v` - Verify nginx installation
- `systemctl enable nginx` - Enable Nginx service
- `systemctl start nginx` - Start Nginx service
- `systemctl is-active nginx` - Check Nginx service status
- `systemctl is-enabled nginx` - Check if Nginx service is enabled
- `ss -tlnp` or `netstat -tlnp` - Check if port 80 is listening

### File Operations

- Checks for `/usr/sbin/nginx` binary to verify installation

### Dependencies

- `github.com/stwalsh4118/phanes/internal/module` - Module interface
- `github.com/stwalsh4118/phanes/internal/config` - Configuration structure
- `github.com/stwalsh4118/phanes/internal/exec` - Command execution and file operations
- `github.com/stwalsh4118/phanes/internal/log` - Logging functions
- `strings` - String operations

## Caddy Module

Package: `github.com/stwalsh4118/phanes/internal/modules/caddy`

Implements the Caddy web server installation module that installs Caddy via the official apt repository, creates a default Caddyfile, configures the service, and verifies it is running and accessible on port 80. Caddy provides automatic HTTPS certificate management via Let's Encrypt.

### Public Types

```go
// CaddyModule implements the Module interface for Caddy web server installation.
type CaddyModule struct{}
```

### Module Interface Implementation

```go
// Name returns "caddy"
func (m *CaddyModule) Name() string

// Description returns "Installs and configures Caddy web server with automatic HTTPS"
func (m *CaddyModule) Description() string

// IsInstalled checks if Caddy is already installed and configured.
// Verifies that Caddy is installed, service is running, and port 80 is accessible.
// Note: Since IsInstalled() doesn't receive config, it performs generic checks.
// Install() performs specific checks with config and is fully idempotent.
func (m *CaddyModule) IsInstalled() (bool, error)

// Install installs and configures Caddy web server.
// Uses cfg.Caddy.Enabled (defaults to true) - skips installation if disabled.
func (m *CaddyModule) Install(cfg *config.Config) error
```

### Usage Examples

```go
import (
    "github.com/stwalsh4118/phanes/internal/modules/caddy"
    "github.com/stwalsh4118/phanes/internal/config"
    "github.com/stwalsh4118/phanes/internal/runner"
)

// Create and register caddy module
mod := &caddy.CaddyModule{}
r := runner.NewRunner()
r.RegisterModule(mod)

// Load configuration
cfg, err := config.Load("config.yaml")
if err != nil {
    log.Error("Failed to load config: %v", err)
    return
}

// Check if already installed
installed, err := mod.IsInstalled()
if err != nil {
    log.Error("Failed to check installation status: %v", err)
    return
}

if !installed {
    // Install Caddy
    if err := mod.Install(cfg); err != nil {
        log.Error("Failed to install Caddy: %v", err)
        return
    }
    log.Success("Caddy installation completed")
} else {
    log.Skip("Caddy already installed")
}
```

### Configuration

The module uses the following configuration fields:

- `config.Caddy.Enabled` - Whether to install Caddy (defaults to `true`)

### Behavior

- **Repository Setup**: Adds Caddy's official apt repository by installing prerequisites (`debian-keyring`, `debian-archive-keyring`, `apt-transport-https`, `curl`), downloading and installing GPG key from `https://dl.cloudsmith.io/public/caddy/stable/gpg.key`, adding repository source from `https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt`, and updating package list.
- **Package Installation**: Installs Caddy via apt (`apt-get install -y caddy`) from the official repository and verifies installation with `caddy version`.
- **Caddyfile Creation**: Creates default Caddyfile at `/etc/caddy/Caddyfile` with content `localhost { respond "Caddy is running!" }` if it doesn't exist. Creates `/etc/caddy/` directory if needed.
- **Service Configuration**: Enables Caddy service to start on boot using `systemctl enable caddy`. Starts the service if not running using `systemctl start caddy`. Verifies service is running after start.
- **Port Verification**: Checks if Caddy is listening on port 80 using `ss -tlnp` (or `netstat -tlnp` as fallback). Logs access URL: `http://localhost` and mentions automatic HTTPS capability.
- **Port Conflict Detection**: Before installation, checks if port 80 is already in use by another service. If so, logs a warning but proceeds with installation (user may have configured it intentionally).
- **Idempotency**: `IsInstalled()` checks if Caddy is installed, service is running, and port is accessible. `Install()` is fully idempotent - checks if each component is already configured before making changes.
- **Error Handling**: Returns descriptive errors if GPG key download fails, repository addition fails, apt update fails, caddy installation fails, Caddyfile creation fails, service start/enable fails, or port check fails.
- **Dry-Run Support**: Checks dry-run mode using `log.IsDryRun()` and logs what would be done without executing commands. Still performs checks (Caddy installed, etc.).
- **Configuration Flag**: Respects `cfg.Caddy.Enabled` flag. If `Enabled` is `false`, skips installation and logs skip message. If `Enabled` is `true` (default), proceeds with installation. Set to `false` in config to disable the module when included in a profile.
- **Logging**: Uses `log.Info()` for progress messages (especially during installation), `log.Success()` for completion, `log.Skip()` for already-configured items, `log.Warn()` when port conflicts are detected, and `log.Error()` for errors. Mentions automatic HTTPS capability after successful installation.

### Commands Used

- `apt-get update` - Update package list
- `apt-get install -y debian-keyring debian-archive-keyring apt-transport-https curl` - Install prerequisites
- `curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg` - Add GPG key
- `curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | tee /etc/apt/sources.list.d/caddy-stable.list` - Add repository
- `apt-get install -y caddy` - Install caddy package
- `caddy version` - Verify caddy installation
- `systemctl enable caddy` - Enable Caddy service
- `systemctl start caddy` - Start Caddy service
- `systemctl is-active caddy` - Check Caddy service status
- `systemctl is-enabled caddy` - Check if Caddy service is enabled
- `ss -tlnp` or `netstat -tlnp` - Check if port 80 is listening

### File Operations

- Creates `/etc/caddy/` directory with permissions 0755 if it doesn't exist
- Creates `/etc/caddy/Caddyfile` with default content if it doesn't exist (permissions 0644)
- Checks for `/usr/bin/caddy` binary to verify installation
- Creates `/usr/share/keyrings/caddy-stable-archive-keyring.gpg` with GPG keyring
- Creates `/etc/apt/sources.list.d/caddy-stable.list` with repository entry

### Dependencies

- `github.com/stwalsh4118/phanes/internal/module` - Module interface
- `github.com/stwalsh4118/phanes/internal/config` - Configuration structure
- `github.com/stwalsh4118/phanes/internal/exec` - Command execution and file operations
- `github.com/stwalsh4118/phanes/internal/log` - Logging functions
- `os` - File operations
- `strings` - String operations

## Coolify Module

Package: `github.com/stwalsh4118/phanes/internal/modules/coolify`

Implements the Coolify module that installs Coolify, a self-hosted PaaS platform. Coolify requires Docker to be installed and running before installation can proceed.

### Public Types

```go
// CoolifyModule implements the Module interface for Coolify installation.
type CoolifyModule struct{}
```

### Module Interface Implementation

```go
// Name returns "coolify"
func (m *CoolifyModule) Name() string

// Description returns "Installs and configures Coolify self-hosted PaaS"
func (m *CoolifyModule) Description() string

// IsInstalled checks if Coolify is already installed and running.
// Checks Docker dependency and Coolify containers.
func (m *CoolifyModule) IsInstalled() (bool, error)

// Install installs Coolify using the official install script.
// Requires Docker to be installed and running.
func (m *CoolifyModule) Install(cfg *config.Config) error
```

### Usage Examples

```go
import (
    "github.com/stwalsh4118/phanes/internal/modules/coolify"
    "github.com/stwalsh4118/phanes/internal/config"
)

// Check if Coolify is installed
mod := &coolify.CoolifyModule{}
installed, err := mod.IsInstalled()
if err != nil {
    log.Error("Failed to check Coolify installation: %v", err)
    return
}

if !installed {
    // Install Coolify
    cfg := config.DefaultConfig()
    cfg.Coolify.Enabled = true
    if err := mod.Install(cfg); err != nil {
        log.Error("Failed to install Coolify: %v", err)
        return
    }
    log.Success("Coolify installed")
} else {
    log.Skip("Coolify already installed")
}
```

### Behavior

- **Docker Dependency**: Checks that Docker is installed (`docker --version`) and running (`systemctl is-active docker`) before proceeding. Returns error if Docker is not available.
- **Installation**: Uses official Coolify install script (`curl -fsSL https://cdn.coollabs.io/coolify/install.sh | bash`). Script handles Docker Compose setup and configuration automatically.
- **Verification**: After installation, verifies Coolify containers are running by checking `docker ps` output for containers with "coolify" in their names.
- **Enabled Flag**: Respects `cfg.Coolify.Enabled` - if false, skips installation with `log.Skip()`.
- **Idempotency**: `IsInstalled()` checks Docker dependency and Coolify containers. `Install()` checks if already installed before installing.
- **Error Handling**: Returns descriptive errors if Docker dependency check fails, installation script fails, or verification fails.
- **Dry-Run Support**: Checks dry-run mode using `log.IsDryRun()` and logs what would be done without executing commands.
- **Logging**: Uses `log.Info()` for progress messages, `log.Success()` for completion, `log.Skip()` when disabled or already installed, `log.Warn()` if Docker is not available.

### Commands Used

- `docker --version` - Verify Docker is installed
- `systemctl is-active docker` - Verify Docker service is running
- `docker ps --format '{{.Names}}'` - List container names to check for Coolify containers
- `curl -fsSL https://cdn.coollabs.io/coolify/install.sh | bash` - Install Coolify using official script

### Configuration

The module uses the following configuration fields:

- `config.Coolify.Enabled` - Whether to install Coolify (defaults to `true`)

### Web Interface

Coolify provides a web-based dashboard accessible on **port 8000** by default. After installation:

1. **Access URL**: `http://localhost:8000` (or `http://<server-ip>:8000` for remote servers)
2. **First Visit**: Create an admin account on first access
3. **Vagrant Setup**: For Vagrant VMs, port forwarding must be configured:
   ```ruby
   config.vm.network "forwarded_port", guest: 8000, host: 8000, id: "coolify"
   ```
   Then reload the VM: `vagrant reload`

### Files Created/Modified

Coolify installation script creates:
- Docker Compose configuration files (location depends on Coolify installation)
- Coolify containers and volumes

### Dependencies

**Runtime Dependencies:**
- **Docker**: Requires Docker to be installed and running. The Docker module should be run before the Coolify module.

**Code Dependencies:**
- `github.com/stwalsh4118/phanes/internal/module` - Module interface
- `github.com/stwalsh4118/phanes/internal/config` - Configuration structure
- `github.com/stwalsh4118/phanes/internal/exec` - Command execution
- `github.com/stwalsh4118/phanes/internal/log` - Logging functions
- `strings` - String operations

## PostgreSQL Module

Package: `github.com/stwalsh4118/phanes/internal/modules/postgres`

Implements the PostgreSQL database server installation module that installs PostgreSQL via the official APT repository, creates initial database and user, configures the service, and verifies it is running and accessible on port 5432. This provides relational database capabilities for applications.

### Public Types

```go
// PostgresModule implements the Module interface for PostgreSQL installation.
type PostgresModule struct{}
```

### Module Interface Implementation

```go
// Name returns "postgres"
func (m *PostgresModule) Name() string

// Description returns "Installs and configures PostgreSQL database server"
func (m *PostgresModule) Description() string

// IsInstalled checks if PostgreSQL is already installed and configured.
// Verifies that PostgreSQL is installed, service is running, and port 5432 is accessible.
// Note: Since IsInstalled() doesn't receive config, it performs generic checks.
// Install() performs specific checks with config and is fully idempotent.
func (m *PostgresModule) IsInstalled() (bool, error)

// Install installs and configures PostgreSQL database server.
// Uses cfg.Postgres.Enabled (defaults to true), cfg.Postgres.Version (defaults to "16"),
// cfg.Postgres.Password (required), cfg.Postgres.Database (defaults to "phanes"),
// and cfg.Postgres.User (defaults to "phanes").
func (m *PostgresModule) Install(cfg *config.Config) error
```

### Usage Examples

```go
import (
    "github.com/stwalsh4118/phanes/internal/modules/postgres"
    "github.com/stwalsh4118/phanes/internal/config"
    "github.com/stwalsh4118/phanes/internal/runner"
)

// Create and register postgres module
mod := &postgres.PostgresModule{}
r := runner.NewRunner()
r.RegisterModule(mod)

// Load configuration
cfg, err := config.Load("config.yaml")
if err != nil {
    log.Error("Failed to load config: %v", err)
    return
}

// Check if already installed
installed, err := mod.IsInstalled()
if err != nil {
    log.Error("Failed to check installation status: %v", err)
    return
}

if !installed {
    // Install PostgreSQL
    if err := mod.Install(cfg); err != nil {
        log.Error("Failed to install PostgreSQL: %v", err)
        return
    }
    log.Success("PostgreSQL installation completed")
} else {
    log.Skip("PostgreSQL already installed")
}
```

### Configuration

The module uses the following configuration fields:

- `config.Postgres.Enabled` - Whether to install PostgreSQL (defaults to `true`)
- `config.Postgres.Version` - PostgreSQL version to install (defaults to `"16"`)
- `config.Postgres.Password` - Password for the database user (required)
- `config.Postgres.Database` - Database name to create (defaults to `"phanes"`)
- `config.Postgres.User` - Database user name to create (defaults to `"phanes"`)

### Behavior

- **Repository Setup**: Adds PostgreSQL's official APT repository by installing prerequisites (`wget`, `ca-certificates`), downloading GPG key from `https://www.postgresql.org/media/keys/ACCC4CF8.asc`, adding repository source, and updating package list.
- **Package Installation**: Installs PostgreSQL via apt (`apt-get install -y postgresql-<version>`) from the official repository and verifies installation with `psql --version`.
- **Service Configuration**: Enables PostgreSQL service to start on boot using `systemctl enable postgresql`. Starts the service if not running using `systemctl start postgresql`. Verifies service is running after start.
- **Local Connection Configuration**: PostgreSQL's default `pg_hba.conf` configuration allows local connections via peer authentication for the postgres user, which is sufficient for local database operations. The module relies on these defaults and does not modify `pg_hba.conf`.
- **Database Creation**: Creates initial database (from config, default "phanes") if it doesn't exist using `psql -U postgres -c "CREATE DATABASE <database>;"`.
- **User Creation**: Creates database user (from config, default "phanes") with password if it doesn't exist using `psql -U postgres -c "CREATE USER <user> WITH PASSWORD '<password>';"`. Password is passed securely via command (not logged).
- **Privilege Granting**: Grants all privileges on the database to the user using `psql -U postgres -c "GRANT ALL PRIVILEGES ON DATABASE <database> TO <user>;"`.
- **Port Verification**: Checks if PostgreSQL is listening on port 5432 using `ss -tlnp` (or `netstat -tlnp` as fallback). Logs connection details after successful installation (without password).
- **Idempotency**: `IsInstalled()` checks if PostgreSQL is installed, service is running, and port is accessible. `Install()` is fully idempotent - checks if each component is already configured before making changes.
- **Error Handling**: Validates password is not empty before proceeding. Returns descriptive errors if GPG key download fails, repository addition fails, apt update fails, PostgreSQL installation fails, database/user creation fails, service start/enable fails, or port check fails.
- **Dry-Run Support**: Checks dry-run mode using `log.IsDryRun()` and logs what would be done without executing commands. Still performs checks (PostgreSQL installed, etc.).
- **Configuration Flag**: Respects `cfg.Postgres.Enabled` flag. If `Enabled` is `false`, skips installation and logs skip message. If `Enabled` is `true` (default), proceeds with installation. Set to `false` in config to disable the module when included in a profile.
- **Password Security**: Never logs passwords in any form. Uses secure command execution for password handling. Logs connection details without password after installation.
- **Logging**: Uses `log.Info()` for progress messages (especially during installation), `log.Success()` for completion, `log.Skip()` for already-configured items, `log.Warn()` when port is not yet accessible, and `log.Error()` for errors. Shows connection details after successful installation.

### Commands Used

- `apt-get update` - Update package list
- `apt-get install -y wget ca-certificates` - Install prerequisites
- `wget --quiet -O - <gpg-url> | gpg --dearmor -o <keyring>` - Add GPG key
- `lsb_release -cs` or `/etc/os-release` - Get distribution codename
- `apt-get install -y postgresql-<version>` - Install PostgreSQL package
- `psql --version` - Verify PostgreSQL installation
- `systemctl enable postgresql` - Enable PostgreSQL service
- `systemctl start postgresql` - Start PostgreSQL service
- `systemctl is-active postgresql` - Check PostgreSQL service status
- `systemctl is-enabled postgresql` - Check if PostgreSQL service is enabled
- `psql -U postgres -lqt` - List databases
- `psql -U postgres -c "CREATE DATABASE <database>;"` - Create database
- `psql -U postgres -tAc "SELECT 1 FROM pg_roles WHERE rolname='<user>'"` - Check if user exists
- `psql -U postgres -c "CREATE USER <user> WITH PASSWORD '<password>';"` - Create user
- `psql -U postgres -c "GRANT ALL PRIVILEGES ON DATABASE <database> TO <user>;"` - Grant privileges
- `ss -tlnp` or `netstat -tlnp` - Check if port 5432 is listening

### File Operations

- Creates `/usr/share/keyrings/postgresql-archive-keyring.gpg` with GPG keyring
- Creates `/etc/apt/sources.list.d/pgdg.list` with repository entry
- Checks for `/usr/bin/psql` binary to verify installation

### Dependencies

- `github.com/stwalsh4118/phanes/internal/module` - Module interface
- `github.com/stwalsh4118/phanes/internal/config` - Configuration structure
- `github.com/stwalsh4118/phanes/internal/exec` - Command execution and file operations
- `github.com/stwalsh4118/phanes/internal/log` - Logging functions
- `os` - File operations and environment variables
- `os/exec` - Command execution with environment variables
- `bufio` - Reading /etc/os-release
- `strings` - String operations

## Redis Module

Package: `github.com/stwalsh4118/phanes/internal/modules/redis`

Implements the Redis in-memory data store installation module that installs Redis via apt, configures bind address and password, enables and starts the service, and verifies it is running and accessible on port 6379. This provides caching and session storage capabilities for applications.

### Public Types

```go
// RedisModule implements the Module interface for Redis installation.
type RedisModule struct{}
```

### Module Interface Implementation

```go
// Name returns "redis"
func (m *RedisModule) Name() string

// Description returns "Installs and configures Redis in-memory data store"
func (m *RedisModule) Description() string

// IsInstalled checks if Redis is already installed and configured.
// Verifies that Redis is installed, service is running, port 6379 is accessible, and responds to ping.
// Note: Since IsInstalled() doesn't receive config, it performs generic checks.
// Install() performs specific checks with config and is fully idempotent.
func (m *RedisModule) IsInstalled() (bool, error)

// Install installs and configures Redis in-memory data store.
// Uses cfg.Redis.Enabled (defaults to true), cfg.Redis.BindAddress (defaults to "127.0.0.1"),
// and cfg.Redis.Password (optional, empty means no password).
func (m *RedisModule) Install(cfg *config.Config) error
```

### Usage Examples

```go
import (
    "github.com/stwalsh4118/phanes/internal/modules/redis"
    "github.com/stwalsh4118/phanes/internal/config"
    "github.com/stwalsh4118/phanes/internal/runner"
)

// Create and register redis module
mod := &redis.RedisModule{}
r := runner.NewRunner()
r.RegisterModule(mod)

// Load configuration
cfg, err := config.Load("config.yaml")
if err != nil {
    log.Error("Failed to load config: %v", err)
    return
}

// Check if already installed
installed, err := mod.IsInstalled()
if err != nil {
    log.Error("Failed to check installation status: %v", err)
    return
}

if !installed {
    // Install Redis
    if err := mod.Install(cfg); err != nil {
        log.Error("Failed to install Redis: %v", err)
        return
    }
    log.Success("Redis installation completed")
} else {
    log.Skip("Redis already installed")
}
```

### Configuration

The module uses the following configuration fields:

- `config.Redis.Enabled` - Whether to install Redis (defaults to `true`)
- `config.Redis.BindAddress` - Bind address for Redis (defaults to `"127.0.0.1"`)
- `config.Redis.Password` - Password for Redis authentication (optional, empty means no password)

### Behavior

- **Package Installation**: Installs Redis via apt (`apt-get install -y redis-server`) from default Ubuntu repositories. Verifies installation with `redis-cli --version`.
- **Config File Modification**: Modifies `/etc/redis/redis.conf` to configure bind address and password. Updates `bind` directive for network binding and `requirepass` directive for password authentication. When password is removed, the `requirepass` line is commented out.
- **Bind Address Configuration**: Configures Redis to bind to specified address (defaults to `127.0.0.1` for localhost-only access). Warns if binding to all interfaces (`0.0.0.0` or `::`) without password configured.
- **Password Configuration**: Sets password via `requirepass` directive if provided. Password is optional - if empty, the `requirepass` directive is commented out. Never logs passwords in any form.
- **Service Configuration**: Enables Redis service to start on boot using `systemctl enable redis-server`. Starts the service if not running using `systemctl start redis-server`. Verifies service is running after start.
- **Configuration Reload**: Reloads Redis configuration using `systemctl reload redis-server` after config changes. Falls back to restart if reload fails. Verifies service is still running after reload/restart.
- **Port Verification**: Checks if Redis is listening on port 6379 using `ss -tlnp` (or `netstat -tlnp` as fallback). Logs connection details after successful installation (without password).
- **Ping Verification**: Tests Redis connectivity using `redis-cli ping` (or `redis-cli -a <password> ping` if password is configured). Verifies Redis responds with "PONG".
- **Idempotency**: `IsInstalled()` checks if Redis is installed, service is running, port is accessible, and ping responds. `Install()` is fully idempotent - checks if each component is already configured before making changes.
- **Error Handling**: Returns descriptive errors if apt update fails, Redis installation fails, config file modification fails, service start/enable fails, config reload fails, port check fails, or ping test fails.
- **Dry-Run Support**: Checks dry-run mode using `log.IsDryRun()` and logs what would be done without executing commands or modifying files. Still performs checks (Redis installed, etc.).
- **Configuration Flag**: Respects `cfg.Redis.Enabled` flag. If `Enabled` is `false`, skips installation and logs skip message. If `Enabled` is `true` (default), proceeds with installation. Set to `false` in config to disable the module when included in a profile.
- **Security Warning**: Warns if `BindAddress` is set to `0.0.0.0` or `::` (all interfaces) without password configured. This is insecure and exposes Redis to the network without authentication.
- **Password Security**: Never logs passwords in any form. Uses secure command execution for password handling. Logs connection details without password after installation.
- **Logging**: Uses `log.Info()` for progress messages (especially during installation), `log.Success()` for completion, `log.Skip()` for already-configured items, `log.Warn()` when binding to all interfaces without password or when port is not yet accessible, and `log.Error()` for errors. Shows connection details after successful installation.

### Commands Used

- `apt-get update` - Update package list
- `apt-get install -y redis-server` - Install Redis package
- `redis-cli --version` - Verify Redis installation
- `systemctl enable redis-server` - Enable Redis service
- `systemctl start redis-server` - Start Redis service
- `systemctl reload redis-server` - Reload Redis configuration
- `systemctl restart redis-server` - Restart Redis service (fallback if reload fails)
- `systemctl is-active redis-server` - Check Redis service status
- `systemctl is-enabled redis-server` - Check if Redis service is enabled
- `redis-cli ping` - Test Redis connectivity (without password)
- `redis-cli -a <password> ping` - Test Redis connectivity (with password)
- `ss -tlnp` or `netstat -tlnp` - Check if port 6379 is listening

### File Operations

- Modifies `/etc/redis/redis.conf` to configure bind address and password
- Checks for `/usr/bin/redis-cli` binary to verify installation

### Dependencies

- `github.com/stwalsh4118/phanes/internal/module` - Module interface
- `github.com/stwalsh4118/phanes/internal/config` - Configuration structure
- `github.com/stwalsh4118/phanes/internal/exec` - Command execution and file operations
- `github.com/stwalsh4118/phanes/internal/log` - Logging functions
- `os` - File operations
- `os/exec` - Command execution
- `bufio` - Reading Redis config file
- `strings` - String operations

## DevTools Module - Core Tools

Package: `github.com/stwalsh4118/phanes/internal/modules/devtools`

Implements helper functions for installing core development tools (Git, build-essential, curl, wget, ca-certificates) via apt. These functions are used by the main DevTools module orchestrator.

### Public Functions

```go
// installCoreTools installs core development tools via apt.
// Installs: git, build-essential, curl, wget, ca-certificates
// Returns an error if installation fails.
func installCoreTools(cfg *config.Config) error

// coreToolsInstalled checks if all core tools are installed.
// Returns true only if ALL tools are installed.
func coreToolsInstalled() (bool, error)

// gitInstalled checks if Git is installed.
func gitInstalled() (bool, error)

// buildEssentialInstalled checks if build-essential package is installed.
// Verifies the package and that gcc and make are available.
func buildEssentialInstalled() (bool, error)

// curlInstalled checks if curl is installed.
func curlInstalled() (bool, error)

// wgetInstalled checks if wget is installed.
func wgetInstalled() (bool, error)

// caCertificatesInstalled checks if ca-certificates package is installed.
func caCertificatesInstalled() (bool, error)
```

### Usage Examples

```go
import (
    "github.com/stwalsh4118/phanes/internal/modules/devtools"
    "github.com/stwalsh4118/phanes/internal/config"
)

// Check if core tools are installed
installed, err := devtools.coreToolsInstalled()
if err != nil {
    log.Error("Failed to check core tools: %v", err)
    return
}

if !installed {
    // Install core tools
    cfg := config.DefaultConfig()
    if err := devtools.installCoreTools(cfg); err != nil {
        log.Error("Failed to install core tools: %v", err)
        return
    }
    log.Success("Core tools installed")
} else {
    log.Skip("Core tools already installed")
}
```

### Behavior

- **Package Installation**: Installs packages via apt (`apt-get install -y git build-essential curl wget ca-certificates`). Updates package list first, then installs all packages in a single command.
- **Installation Checks**: 
  - Uses `exec.CommandExists()` for binary checks (git, curl, wget)
  - Uses `dpkg -l` for package checks (build-essential, ca-certificates)
  - Verifies `gcc` and `make` are available for build-essential
- **Idempotency**: `coreToolsInstalled()` checks if all tools are installed. `installCoreTools()` returns early with `log.Skip()` if all tools are already installed.
- **Error Handling**: Returns descriptive errors if apt update fails, package installation fails, or verification fails.
- **Dry-Run Support**: Checks dry-run mode using `log.IsDryRun()` and logs what would be done without executing commands.
- **Logging**: Uses `log.Info()` for progress messages, `log.Success()` for completion, and `log.Skip()` when tools are already installed.

### Commands Used

- `apt-get update` - Update package list
- `apt-get install -y git build-essential curl wget ca-certificates` - Install core development tools
- `dpkg -l <package>` - Check if package is installed
- `which <command>` - Check if command exists (via exec.CommandExists)

### Packages Installed

- `git` - Version control system
- `build-essential` - GCC, make, and other build tools
- `curl` - HTTP client
- `wget` - File downloader
- `ca-certificates` - SSL certificates

### Dependencies

- `github.com/stwalsh4118/phanes/internal/config` - Configuration structure
- `github.com/stwalsh4118/phanes/internal/exec` - Command execution
- `github.com/stwalsh4118/phanes/internal/log` - Logging functions
- `strings` - String operations

## DevTools Module - Node.js via nvm

Package: `github.com/stwalsh4118/phanes/internal/modules/devtools`

Implements helper functions for installing nvm (Node Version Manager) and Node.js per-user. nvm is installed in the user's home directory (`~/.nvm`) and requires shell profile configuration.

### Public Functions

```go
// installNodeJS installs nvm and Node.js for the configured user.
// Installs nvm to ~/.nvm and configures shell profiles (.bashrc, .zshrc).
// Uses cfg.User.Username (required) and cfg.DevTools.NodeVersion (defaults to "22").
// Returns an error if installation fails.
func installNodeJS(cfg *config.Config) error

// nvmInstalled checks if nvm is installed for a specific user.
// Checks if ~/.nvm directory and nvm.sh script exist.
func nvmInstalled(username string) (bool, error)

// nodeInstalled checks if a Node.js version is installed via nvm for a user.
// Uses nvm which to check if the version is installed.
func nodeInstalled(username, version string) (bool, error)
```

### Usage Examples

```go
import (
    "github.com/stwalsh4118/phanes/internal/modules/devtools"
    "github.com/stwalsh4118/phanes/internal/config"
)

// Check if nvm is installed
installed, err := devtools.nvmInstalled("myuser")
if err != nil {
    log.Error("Failed to check nvm: %v", err)
    return
}

if !installed {
    // Install nvm and Node.js
    cfg := config.DefaultConfig()
    cfg.User.Username = "myuser"
    cfg.DevTools.NodeVersion = "22"
    if err := devtools.installNodeJS(cfg); err != nil {
        log.Error("Failed to install Node.js: %v", err)
        return
    }
    log.Success("Node.js installed")
} else {
    log.Skip("nvm already installed")
}
```

### Behavior

- **nvm Installation**: Downloads and runs the official nvm install script (`curl -o- <url> | bash`) as the target user. Installs to `~/.nvm` directory. Uses nvm version v0.40.0.
- **Shell Profile Configuration**: Appends nvm initialization to `.bashrc` and `.zshrc` if not already present. Initialization includes:
  - `export NVM_DIR="$HOME/.nvm"`
  - `[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"`
  - `[ -s "$NVM_DIR/bash_completion" ] && \. "$NVM_DIR/bash_completion"`
- **Node.js Installation**: Uses nvm to install the specified Node.js version (defaults to "22"). Sets the version as default using `nvm alias default <version>`. Verifies installation with `node --version` and `npm --version`.
- **Installation Checks**: 
  - `nvmInstalled()` checks if `~/.nvm` directory and `nvm.sh` script exist
  - `nodeInstalled()` uses `nvm which <version>` to check if a specific version is installed
- **Idempotency**: `installNodeJS()` checks if nvm and Node.js are already installed before installing. Returns early with `log.Skip()` if already configured.
- **Error Handling**: Validates username is set (warns and skips if not). Returns descriptive errors if nvm installation fails, shell profile configuration fails, Node.js installation fails, or verification fails.
- **Dry-Run Support**: Checks dry-run mode using `log.IsDryRun()` and logs what would be done without executing commands or writing files.
- **User Context**: All commands run as the target user using `su - <username> -c "<command>"` to ensure proper environment and file ownership.
- **File Ownership**: Sets correct ownership on shell profile files using `os.Chown()` with the user's UID/GID.
- **Logging**: Uses `log.Info()` for progress messages, `log.Success()` for completion, `log.Skip()` when already installed, and `log.Warn()` if username is not set.

### Commands Used

- `su - <username> -c "curl -o- <nvm-url> | bash"` - Install nvm as the user
- `su - <username> -c "source ~/.nvm/nvm.sh && nvm install <version>"` - Install Node.js via nvm
- `su - <username> -c "source ~/.nvm/nvm.sh && nvm alias default <version>"` - Set default Node.js version
- `su - <username> -c "source ~/.nvm/nvm.sh && nvm which <version>"` - Check if Node.js version is installed
- `su - <username> -c "source ~/.nvm/nvm.sh && node --version && npm --version"` - Verify Node.js installation

### Configuration

The module uses the following configuration fields:

- `config.User.Username` - Username for installing nvm and Node.js (required)
- `config.DevTools.NodeVersion` - Node.js version to install (defaults to "22" if empty)

### Files Created/Modified

- `~/.nvm/` - nvm installation directory (created by nvm install script)
- `~/.bashrc` - Shell profile with nvm initialization appended
- `~/.zshrc` - Shell profile with nvm initialization appended

### Dependencies

- `github.com/stwalsh4118/phanes/internal/config` - Configuration structure
- `github.com/stwalsh4118/phanes/internal/exec` - Command execution
- `github.com/stwalsh4118/phanes/internal/log` - Logging functions
- `os` - File operations and ownership
- `os/user` - User lookup for UID/GID
- `path/filepath` - Path operations
- `strconv` - String to integer conversion
- `strings` - String operations

## DevTools Module - Python and uv

Package: `github.com/stwalsh4118/phanes/internal/modules/devtools`

Implements helper functions for installing system Python 3 via apt and optionally uv (Python package manager) per-user. Python is installed system-wide, while uv is installed in the user's home directory (`~/.local/bin/uv`) and requires shell profile configuration.

### Public Functions

```go
// installPython installs Python 3 and optionally uv for the configured user.
// Installs Python 3 via apt and uv to ~/.local/bin/uv (if enabled).
// Configures shell profiles (.bashrc, .zshrc) for uv PATH.
// Uses cfg.User.Username (required for uv) and cfg.DevTools.InstallUv (defaults to true).
// Returns an error if installation fails.
func installPython(cfg *config.Config) error

// pythonInstalled checks if Python 3 is installed.
// Checks if python3 command exists in PATH.
func pythonInstalled() (bool, error)

// uvInstalled checks if uv is installed for a specific user.
// Checks if ~/.local/bin/uv exists.
func uvInstalled(username string) (bool, error)
```

### Usage Examples

```go
import (
    "github.com/stwalsh4118/phanes/internal/modules/devtools"
    "github.com/stwalsh4118/phanes/internal/config"
)

// Check if Python 3 is installed
installed, err := devtools.pythonInstalled()
if err != nil {
    log.Error("Failed to check Python: %v", err)
    return
}

if !installed {
    // Install Python 3 and uv
    cfg := config.DefaultConfig()
    cfg.User.Username = "myuser"
    cfg.DevTools.InstallUv = true
    if err := devtools.installPython(cfg); err != nil {
        log.Error("Failed to install Python: %v", err)
        return
    }
    log.Success("Python installed")
} else {
    log.Skip("Python 3 already installed")
}
```

### Behavior

- **Python 3 Installation**: Installs system Python 3 via apt (`apt-get install -y python3 python3-venv python3-pip`). Updates apt package list before installation. Verifies installation with `python3 --version`.
- **uv Installation**: If `cfg.DevTools.InstallUv` is true, downloads and runs the official uv install script (`curl -LsSf https://astral.sh/uv/install.sh | sh`) as the target user. Installs to `~/.local/bin/uv`. Verifies installation with `uv --version`.
- **Shell Profile Configuration**: Appends uv PATH export to `.bashrc` and `.zshrc` if not already present. PATH export: `export PATH="$HOME/.local/bin:$PATH"`.
- **Installation Checks**: 
  - `pythonInstalled()` checks if `python3` command exists in PATH
  - `uvInstalled()` checks if `~/.local/bin/uv` file exists
- **Idempotency**: `installPython()` checks if Python 3 and uv are already installed before installing. Returns early with `log.Skip()` if already configured.
- **Error Handling**: Validates username is set for uv installation (warns and skips if not). Returns descriptive errors if apt update fails, Python installation fails, uv installation fails, shell profile configuration fails, or verification fails.
- **Dry-Run Support**: Checks dry-run mode using `log.IsDryRun()` and logs what would be done without executing commands or writing files.
- **User Context**: uv installation commands run as the target user using `su - <username> -c "<command>"` to ensure proper environment and file ownership.
- **File Ownership**: Sets correct ownership on shell profile files using `os.Chown()` with the user's UID/GID.
- **Logging**: Uses `log.Info()` for progress messages, `log.Success()` for completion, `log.Skip()` when already installed, and `log.Warn()` if username is not set for uv.

### Commands Used

- `apt-get update` - Update apt package list
- `apt-get install -y python3 python3-venv python3-pip` - Install Python 3 and related packages
- `python3 --version` - Verify Python 3 installation
- `su - <username> -c "curl -LsSf https://astral.sh/uv/install.sh | sh"` - Install uv as the user
- `su - <username> -c "~/.local/bin/uv --version"` - Verify uv installation

### Configuration

The module uses the following configuration fields:

- `config.User.Username` - Username for installing uv (required if InstallUv is true)
- `config.DevTools.InstallUv` - Whether to install uv package manager (defaults to `true`)

### Files Created/Modified

- `~/.bashrc` - Shell profile with uv PATH export appended
- `~/.zshrc` - Shell profile with uv PATH export appended
- `~/.local/bin/uv` - uv binary (created by uv install script)

### Dependencies

- `github.com/stwalsh4118/phanes/internal/config` - Configuration structure
- `github.com/stwalsh4118/phanes/internal/exec` - Command execution
- `github.com/stwalsh4118/phanes/internal/log` - Logging functions
- `os` - File operations and ownership
- `os/user` - User lookup for UID/GID
- `path/filepath` - Path operations
- `strconv` - String to integer conversion
- `strings` - String operations

## DevTools Module - Go

Package: `github.com/stwalsh4118/phanes/internal/modules/devtools`

Implements helper functions for installing Go from the official source. Go is downloaded as a tarball, extracted to `/usr/local/go`, and PATH is configured in user shell profiles.

### Public Functions

```go
// installGo installs Go from the official source.
// Downloads tarball, extracts to /usr/local/go, and configures shell profiles.
// Uses cfg.DevTools.GoVersion (defaults to "1.25.5") and cfg.User.Username (required for shell profile config).
// Returns an error if installation fails.
func installGo(cfg *config.Config) error

// goInstalled checks if Go is installed and matches the requested version.
// Checks if go command exists and version matches.
func goInstalled(version string) (bool, error)

// getSystemArch detects the system architecture.
// Uses dpkg --print-architecture with fallback to uname -m.
func getSystemArch() (string, error)

// mapArchToGoArch maps system architecture to Go architecture name.
func mapArchToGoArch(arch string) string
```

### Usage Examples

```go
import (
    "github.com/stwalsh4118/phanes/internal/modules/devtools"
    "github.com/stwalsh4118/phanes/internal/config"
)

// Check if Go is installed
installed, err := devtools.goInstalled("1.25.5")
if err != nil {
    log.Error("Failed to check Go: %v", err)
    return
}

if !installed {
    // Install Go
    cfg := config.DefaultConfig()
    cfg.User.Username = "myuser"
    cfg.DevTools.GoVersion = "1.25.5"
    if err := devtools.installGo(cfg); err != nil {
        log.Error("Failed to install Go: %v", err)
        return
    }
    log.Success("Go installed")
} else {
    log.Skip("Go already installed")
}
```

### Behavior

- **Architecture Detection**: Detects system architecture using `dpkg --print-architecture` (preferred) or `uname -m` (fallback). Maps system architecture to Go architecture (amd64, arm64, armv6l, 386).
- **Go Installation**: Downloads Go tarball from `https://go.dev/dl/go<version>.linux-<arch>.tar.gz`. Removes existing installation at `/usr/local/go` if present. Extracts tarball to `/usr/local`. Verifies installation with `go version`.
- **Shell Profile Configuration**: Appends Go PATH export to `.bashrc` and `.zshrc` if not already present. PATH export: `export PATH=$PATH:/usr/local/go/bin` and `export GOROOT=/usr/local/go`.
- **Installation Checks**: 
  - `goInstalled()` checks if `go` command exists in PATH and version matches
  - Checks if `/usr/local/go/bin` directory exists
- **Idempotency**: `installGo()` checks if Go is already installed with the correct version before installing. Returns early with `log.Skip()` if already configured.
- **Error Handling**: Validates username is set for shell profile configuration (warns and skips if not). Returns descriptive errors if architecture detection fails, download fails, extraction fails, shell profile configuration fails, or verification fails.
- **Dry-Run Support**: Checks dry-run mode using `log.IsDryRun()` and logs what would be done without executing commands or writing files.
- **File Ownership**: Sets correct ownership on shell profile files using `os.Chown()` with the user's UID/GID.
- **Logging**: Uses `log.Info()` for progress messages, `log.Success()` for completion, `log.Skip()` when already installed, and `log.Warn()` if username is not set for shell profile configuration.

### Commands Used

- `dpkg --print-architecture` - Get system architecture (preferred)
- `uname -m` - Get system architecture (fallback)
- `curl -L -o <path> <url>` - Download Go tarball
- `rm -rf /usr/local/go` - Remove old Go installation
- `tar -C /usr/local -xzf <tarball>` - Extract Go tarball
- `go version` - Verify Go installation

### Configuration

The module uses the following configuration fields:

- `config.DevTools.GoVersion` - Go version to install (defaults to `"1.25.5"`) - must include patch number
- `config.User.Username` - Username for shell profile configuration (required for PATH setup)

### Files Created/Modified

- `/usr/local/go` - Go installation directory (extracted from tarball)
- `~/.bashrc` - Shell profile with Go PATH export appended
- `~/.zshrc` - Shell profile with Go PATH export appended
- `/tmp/go<version>.linux-<arch>.tar.gz` - Temporary tarball (downloaded and removed)

### Architecture Mapping

| System Arch | Go Arch |
|-------------|---------|
| amd64, x86_64 | amd64 |
| arm64, aarch64 | arm64 |
| armv6l, armhf | armv6l |
| 386, i386, i686 | 386 |

### Dependencies

- `github.com/stwalsh4118/phanes/internal/config` - Configuration structure
- `github.com/stwalsh4118/phanes/internal/exec` - Command execution
- `github.com/stwalsh4118/phanes/internal/log` - Logging functions
- `os` - File operations and ownership
- `os/user` - User lookup for UID/GID
- `path/filepath` - Path operations
- `strconv` - String to integer conversion
- `strings` - String operations

## DevTools Module - Main Orchestrator

Package: `github.com/stwalsh4118/phanes/internal/modules/devtools`

Implements the main DevTools module that orchestrates installation of all development tools. This module implements the `module.Module` interface and can be run via the CLI.

### Public Types

```go
// DevToolsModule implements the Module interface for development tools installation.
// It orchestrates the installation of core tools, Node.js via nvm, Python/uv, and Go.
type DevToolsModule struct{}
```

### Module Interface Implementation

```go
// Name returns "devtools"
func (m *DevToolsModule) Name() string

// Description returns "Installs development tools (Git, build-essential, Node.js, Python, Go)"
func (m *DevToolsModule) Description() string

// IsInstalled checks if development tools are already installed.
// Returns true if all enabled components are installed.
func (m *DevToolsModule) IsInstalled() (bool, error)

// Install orchestrates the installation of all development tools.
// Respects cfg.DevTools.Enabled flag - if false, skips installation.
func (m *DevToolsModule) Install(cfg *config.Config) error
```

### Usage Examples

```go
import (
    "github.com/stwalsh4118/phanes/internal/modules/devtools"
    "github.com/stwalsh4118/phanes/internal/config"
    "github.com/stwalsh4118/phanes/internal/runner"
)

// Register module with runner
mod := &devtools.DevToolsModule{}
r := runner.NewRunner()
r.RegisterModule(mod)

// Load configuration
cfg, err := config.Load("config.yaml")
if err != nil {
    log.Error("Failed to load config: %v", err)
    return
}

// Execute via runner
err = r.RunModules([]string{"devtools"}, cfg, false)
if err != nil {
    log.Error("Failed to run devtools: %v", err)
}
```

### CLI Usage

```bash
# Run devtools module directly
./phanes --modules devtools --config config.yaml

# Run as part of dev profile
./phanes --profile dev --config config.yaml

# Dry-run to preview
./phanes --modules devtools --config config.yaml --dry-run
```

### Configuration

The module uses the following configuration fields:

- `config.DevTools.Enabled` - Whether to install development tools (defaults to `true`)
- `config.DevTools.NodeVersion` - Node.js version to install (defaults to "22")
- `config.DevTools.PythonVersion` - Python version to install (defaults to "3")
- `config.DevTools.GoVersion` - Go version to install (defaults to "1.25.5")
- `config.DevTools.InstallUv` - Whether to install uv package manager (defaults to `true`)
- `config.User.Username` - Required for per-user installations (nvm, etc.)

### Behavior

- **Orchestration**: Installs components in order: core tools → Node.js → Python → Go. Stops on first error.
- **Core Tools**: Installs git, build-essential, curl, wget, ca-certificates via apt.
- **Node.js**: Installs nvm per-user in `~/.nvm`, then installs specified Node.js version.
- **Python**: Installs system Python 3 via apt and optionally uv package manager per-user.
- **Go**: Installs Go from official source by downloading tarball and extracting to `/usr/local/go`.
- **Enabled Flag**: Respects `cfg.DevTools.Enabled` - if false, skips installation with `log.Skip()`.
- **Idempotency**: Each sub-component checks if already installed before installing.
- **Error Handling**: Returns descriptive errors indicating which component failed.
- **Dry-Run Support**: Propagates dry-run mode to all sub-installations.
- **Logging**: Uses `log.Info()` for progress, `log.Success()` for completion, `log.Skip()` when disabled or already installed.

### Components Installed

| Component | Status | Description |
|-----------|--------|-------------|
| Core Tools | ✅ Implemented | git, build-essential, curl, wget, ca-certificates |
| Node.js/nvm | ✅ Implemented | nvm + Node.js LTS per-user |
| Python/uv | ✅ Implemented | System Python + uv package manager |
| Go | ✅ Implemented | Go from official source |

### Dependencies

- `github.com/stwalsh4118/phanes/internal/config` - Configuration structure
- `github.com/stwalsh4118/phanes/internal/log` - Logging functions
- `github.com/stwalsh4118/phanes/internal/module` - Module interface

