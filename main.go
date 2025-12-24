package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/stwalsh4118/phanes/internal/config"
	"github.com/stwalsh4118/phanes/internal/exec"
	"github.com/stwalsh4118/phanes/internal/log"
)

const (
	programName = "phanes"
	version     = "0.1.0"
)

var (
	profileFlag string
	modulesFlag string
	configFlag  string
	dryRunFlag  bool
	listFlag    bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   programName,
	Short: "VPS Provisioning System",
	Long: `phanes is a tool for provisioning Linux VPS servers with predefined modules
and profiles. It supports idempotent execution, dry-run mode, and
configuration-driven setup.`,
	Version: version,
	RunE:    runCommand,
}

func init() {
	// Define flags
	rootCmd.Flags().StringVar(&profileFlag, "profile", "", "Profile name to execute (e.g., 'dev', 'web', 'database')")
	rootCmd.Flags().StringVar(&modulesFlag, "modules", "", "Comma-separated list of module names to execute")
	rootCmd.Flags().StringVar(&configFlag, "config", "config.yaml", "Path to configuration file")
	rootCmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "Enable dry-run mode (preview changes without executing)")
	rootCmd.Flags().BoolVar(&listFlag, "list", false, "List available modules and profiles")

	// Add example usage
	rootCmd.Example = `  # Run a profile
  phanes --profile dev --config config.yaml

  # Run specific modules
  phanes --modules baseline,user,docker --config config.yaml

  # Preview changes without executing
  phanes --profile dev --config config.yaml --dry-run

  # List available modules and profiles
  phanes --list`
}

// runCommand executes the main command logic
func runCommand(cmd *cobra.Command, args []string) error {
	// Set up dry-run mode for logging if flag is set
	if dryRunFlag {
		log.SetDryRun(true)
		log.Info("Dry-run mode enabled. No changes will be made.")
	}

	// Handle --list flag (show available modules and profiles, then exit)
	if listFlag {
		// TODO: Implement in task 8-5
		log.Info("Listing available modules and profiles...")
		log.Warn("List functionality will be implemented in a future task")
		return nil
	}

	// Validate that either profile or modules is specified
	if profileFlag == "" && modulesFlag == "" {
		log.Error("Error: Either --profile or --modules must be specified")
		fmt.Fprintf(os.Stderr, "\n")
		// Return error - Cobra will show help automatically
		return fmt.Errorf("invalid usage: either --profile or --modules must be specified")
	}

	// Basic validation: if both profile and modules are specified, that's okay
	// (we'll handle combining them in later tasks)
	if profileFlag != "" && modulesFlag != "" {
		log.Warn("Both --profile and --modules specified. Profile modules will be combined with specified modules.")
	}

	// Load configuration file
	cfg, err := loadConfig(configFlag)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Log what we're about to do (for now, just log the flags)
	if profileFlag != "" {
		log.Info("Profile selected: %s", profileFlag)
	}
	if modulesFlag != "" {
		log.Info("Modules selected: %s", modulesFlag)
	}
	log.Info("Config file: %s", configFlag)

	// Store config for later use (will be used in subsequent tasks)
	_ = cfg

	// For now, just exit successfully - actual execution will be in later tasks
	log.Info("Flag parsing complete. Execution will be implemented in subsequent tasks.")
	return nil
}

// loadConfig loads a configuration file from the given path.
// If the file doesn't exist, it returns a default config with a warning.
// If the file exists but is invalid, it returns an error with a clear message.
func loadConfig(path string) (*config.Config, error) {
	// Check if config file exists
	if !exec.FileExists(path) {
		log.Warn("Config file not found at %s, using default configuration", path)
		log.Info("Note: Default config has empty username and SSH public key. These must be set in config file for module execution.")
		return config.DefaultConfig(), nil
	}

	// Load config from file
	cfg, err := config.Load(path)
	if err != nil {
		// Provide clear error messages based on error type
		if os.IsNotExist(err) {
			// This shouldn't happen since we checked FileExists, but handle it anyway
			return nil, fmt.Errorf("config file not found: %s", path)
		}
		// config.Load() wraps errors, so we can return them directly
		return nil, fmt.Errorf("failed to load config from %s: %w", path, err)
	}

	log.Info("Configuration loaded successfully from %s", path)
	return cfg, nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		// Check if it's an invalid usage error (exit code 2)
		if err.Error() == "invalid usage: either --profile or --modules must be specified" {
			os.Exit(2)
		}
		// Other errors exit with code 1
		os.Exit(1)
	}
}
