package module

import "github.com/stwalsh4118/phanes/internal/config"

// Module defines the interface that all provisioning modules must implement.
// This interface ensures consistency across all modules and enables the runner
// to execute modules in a uniform way.
//
// All modules should be idempotent - safe to run multiple times. The IsInstalled()
// method allows modules to check if they're already configured before performing
// installation steps.
//
// Example usage:
//
//	type BaselineModule struct{}
//
//	func (m *BaselineModule) Name() string {
//		return "baseline"
//	}
//
//	func (m *BaselineModule) Description() string {
//		return "Sets timezone, locale, and runs apt update"
//	}
//
//	func (m *BaselineModule) IsInstalled() (bool, error) {
//		// Check if baseline configuration is already applied
//		// Return true if already installed, false if not
//		return false, nil
//	}
//
//	func (m *BaselineModule) Install(cfg *config.Config) error {
//		// Perform installation steps using cfg for configuration
//		// This should only be called if IsInstalled() returns false
//		return nil
//	}
type Module interface {
	// Name returns the unique name identifier for this module.
	// This name is used to register the module and reference it in profiles.
	// Example: "baseline", "docker", "postgres"
	Name() string

	// Description returns a human-readable description of what this module does.
	// This is used for help text and module listings.
	// Example: "Sets timezone, locale, and runs apt update"
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
