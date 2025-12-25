package devtools

import (
	"fmt"

	"github.com/stwalsh4118/phanes/internal/config"
	"github.com/stwalsh4118/phanes/internal/log"
	"github.com/stwalsh4118/phanes/internal/module"
)

// DevToolsModule implements the Module interface for development tools installation.
// It orchestrates the installation of core tools, Node.js via nvm, Python/uv, and Go.
type DevToolsModule struct{}

// Name returns the unique name identifier for this module.
func (m *DevToolsModule) Name() string {
	return "devtools"
}

// Description returns a human-readable description of what this module does.
func (m *DevToolsModule) Description() string {
	return "Installs development tools (Git, build-essential, Node.js, Python, Go)"
}

// IsInstalled checks if development tools are already installed.
// Returns true if all enabled components are installed.
// Note: Since IsInstalled() doesn't receive config, it checks if the core tools
// are installed as a basic check. Install() will do specific checks with config.
func (m *DevToolsModule) IsInstalled() (bool, error) {
	// Check if core tools are installed as a basic check
	coreOk, err := coreToolsInstalled()
	if err != nil {
		return false, fmt.Errorf("failed to check core tools: %w", err)
	}

	// If core tools aren't installed, module is not installed
	if !coreOk {
		return false, nil
	}

	// Return false to ensure Install() is called for specific checks
	// Install() is idempotent and will skip components that are already installed
	return false, nil
}

// Install orchestrates the installation of all development tools.
// It installs components in order: core tools, Node.js, Python, Go.
// Respects cfg.DevTools.Enabled flag - if false, skips installation.
func (m *DevToolsModule) Install(cfg *config.Config) error {
	// Check if DevTools is enabled
	if !cfg.DevTools.Enabled {
		log.Skip("DevTools module is disabled in configuration")
		return nil
	}

	log.Info("Installing development tools")

	// Install core tools (git, build-essential, curl, wget, ca-certificates)
	log.Info("Installing core development tools...")
	if err := installCoreTools(cfg); err != nil {
		return fmt.Errorf("failed to install core tools: %w", err)
	}

	// Install Node.js via nvm
	log.Info("Installing Node.js via nvm...")
	if err := installNodeJS(cfg); err != nil {
		return fmt.Errorf("failed to install Node.js: %w", err)
	}

	// TODO: Install Python and uv (task 7-4)
	// log.Info("Installing Python and uv...")
	// if err := installPython(cfg); err != nil {
	// 	return fmt.Errorf("failed to install Python: %w", err)
	// }

	// TODO: Install Go (task 7-5)
	// log.Info("Installing Go...")
	// if err := installGo(cfg); err != nil {
	// 	return fmt.Errorf("failed to install Go: %w", err)
	// }

	log.Success("Development tools installation completed")
	return nil
}

// Ensure DevToolsModule implements the Module interface
var _ module.Module = (*DevToolsModule)(nil)

