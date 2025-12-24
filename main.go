package main

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stwalsh4118/phanes/internal/config"
	"github.com/stwalsh4118/phanes/internal/exec"
	"github.com/stwalsh4118/phanes/internal/log"
	"github.com/stwalsh4118/phanes/internal/modules/baseline"
	"github.com/stwalsh4118/phanes/internal/modules/user"
	"github.com/stwalsh4118/phanes/internal/profile"
	"github.com/stwalsh4118/phanes/internal/runner"
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
		listProfilesAndModules()
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

	// Handle profile selection if --profile flag is set
	var profileModules []string
	if profileFlag != "" {
		modules, err := getProfileModules(profileFlag)
		if err != nil {
			return fmt.Errorf("profile selection failed: %w", err)
		}
		profileModules = modules
	}

	// Handle module selection if --modules flag is set
	var selectedModules []string
	if modulesFlag != "" {
		modules, err := parseModuleList(modulesFlag)
		if err != nil {
			return fmt.Errorf("module parsing failed: %w", err)
		}
		selectedModules = modules
	}

	log.Info("Config file: %s", configFlag)

	// Combine profile modules and selected modules
	modulesToExecute := combineModules(profileModules, selectedModules)
	if len(modulesToExecute) == 0 {
		return fmt.Errorf("no modules to execute")
	}

	log.Info("Modules to execute: %s", strings.Join(modulesToExecute, ", "))

	// Execute modules using runner
	if err := executeModules(modulesToExecute, cfg, dryRunFlag); err != nil {
		return fmt.Errorf("module execution failed: %w", err)
	}

	log.Success("All modules executed successfully")
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
		if errors.Is(err, os.ErrNotExist) {
			// This shouldn't happen since we checked FileExists, but handle it anyway
			return nil, fmt.Errorf("config file not found: %s", path)
		}
		// config.Load() wraps errors, so we can return them directly
		return nil, fmt.Errorf("failed to load config from %s: %w", path, err)
	}

	log.Info("Configuration loaded successfully from %s", path)
	return cfg, nil
}

// getProfileModules validates that a profile exists and returns its module list.
// If the profile doesn't exist, it returns an error with a list of available profiles.
func getProfileModules(profileName string) ([]string, error) {
	// Validate profile exists
	if !profile.ProfileExists(profileName) {
		// Get available profiles for error message
		availableProfiles := profile.ListProfiles()
		log.Error("Profile '%s' not found", profileName)
		log.Error("Available profiles: %s", strings.Join(availableProfiles, ", "))
		return nil, fmt.Errorf("profile '%s' not found. Available profiles: %s", profileName, strings.Join(availableProfiles, ", "))
	}

	// Get profile modules
	modules, err := profile.GetProfile(profileName)
	if err != nil {
		// This shouldn't happen since we checked ProfileExists, but handle it anyway
		return nil, fmt.Errorf("failed to get profile '%s': %w", profileName, err)
	}

	// Log profile selection and modules
	log.Info("Profile selected: %s", profileName)
	log.Info("Profile modules: %s", strings.Join(modules, ", "))

	return modules, nil
}

// parseModuleList parses a comma-separated module list string.
// It splits by comma, trims whitespace, filters empty strings, and deduplicates module names.
// Returns the parsed and validated module list.
// Note: Actual module validation against the registry will be done by the runner in task 8-6.
func parseModuleList(moduleStr string) ([]string, error) {
	if moduleStr == "" {
		return nil, fmt.Errorf("empty module list")
	}

	// Split by comma
	parts := strings.Split(moduleStr, ",")

	// Trim whitespace and filter out empty strings
	modules := make([]string, 0, len(parts))
	seen := make(map[string]bool)

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		// Filter out empty strings
		if trimmed == "" {
			continue
		}
		// Deduplicate
		if !seen[trimmed] {
			modules = append(modules, trimmed)
			seen[trimmed] = true
		}
	}

	// Check if we have any modules after filtering
	if len(modules) == 0 {
		return nil, fmt.Errorf("no valid modules found in module list")
	}

	// Log selected modules
	log.Info("Modules selected: %s", strings.Join(modules, ", "))

	return modules, nil
}

// registerAllModules creates a runner instance and registers all available modules.
// This function is used for listing modules and can be reused for module execution.
func registerAllModules() *runner.Runner {
	r := runner.NewRunner()

	// Register all available modules
	// Note: As more modules are implemented, they should be added here
	r.RegisterModule(&baseline.BaselineModule{})
	r.RegisterModule(&user.UserModule{})

	return r
}

// combineModules merges profile modules and selected modules, deduplicating module names.
// Profile modules come first, followed by selected modules (which may override duplicates).
// Returns the combined and deduplicated module list.
func combineModules(profileModules []string, selectedModules []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0)

	// Add profile modules first
	for _, module := range profileModules {
		if !seen[module] {
			result = append(result, module)
			seen[module] = true
		}
	}

	// Add selected modules (may include duplicates from profile, but we deduplicate)
	for _, module := range selectedModules {
		if !seen[module] {
			result = append(result, module)
			seen[module] = true
		}
	}

	return result
}

// executeModules creates a runner instance, registers all available modules, and executes
// the specified modules with the given configuration and dry-run flag.
// Returns an error if module execution fails.
func executeModules(moduleNames []string, cfg *config.Config, dryRun bool) error {
	if len(moduleNames) == 0 {
		return fmt.Errorf("no modules specified")
	}

	// Create runner and register all modules
	r := registerAllModules()

	// Execute modules
	log.Info("Starting module execution...")
	if err := r.RunModules(moduleNames, cfg, dryRun); err != nil {
		return fmt.Errorf("failed to execute modules: %w", err)
	}

	return nil
}

// listProfilesAndModules displays all available profiles and modules in a user-friendly format.
// It lists profiles with their module lists, and all registered modules with their descriptions.
func listProfilesAndModules() {
	// Register all modules to get access to the module registry
	r := registerAllModules()

	// List profiles
	log.Info("Available Profiles:")
	profiles := profile.ListProfiles()
	for _, profileName := range profiles {
		modules, err := profile.GetProfile(profileName)
		if err != nil {
			// This shouldn't happen since ListProfiles() only returns valid profiles
			log.Error("Failed to get profile %s: %v", profileName, err)
			continue
		}
		log.Info("  - %s: %s", profileName, strings.Join(modules, ", "))
	}

	// Add blank line between sections
	fmt.Println()

	// List modules
	log.Info("Available Modules:")
	moduleNames := r.ListModules()
	// Sort module names for consistent output
	sort.Strings(moduleNames)
	for _, moduleName := range moduleNames {
		mod := r.GetModule(moduleName)
		if mod != nil {
			log.Info("  - %s: %s", moduleName, mod.Description())
		}
	}
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
