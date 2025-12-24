package runner

import (
	"fmt"

	"github.com/stwalsh4118/phanes/internal/config"
	"github.com/stwalsh4118/phanes/internal/log"
	"github.com/stwalsh4118/phanes/internal/module"
)

// Runner manages a registry of modules and executes them in order.
// It ensures idempotency by checking IsInstalled() before calling Install(),
// and supports dry-run mode for previewing actions without executing them.
type Runner struct {
	modules map[string]module.Module
}

// NewRunner creates a new Runner instance with an empty module registry.
func NewRunner() *Runner {
	return &Runner{
		modules: make(map[string]module.Module),
	}
}

// RegisterModule adds a module to the registry.
// If a module with the same name is already registered, it will be overwritten
// and a warning will be logged.
func (r *Runner) RegisterModule(mod module.Module) {
	name := mod.Name()
	if _, exists := r.modules[name]; exists {
		log.Warn("Module %s is already registered, overwriting", name)
	}
	r.modules[name] = mod
	log.Info("Registered module: %s - %s", name, mod.Description())
}

// RunModules executes the specified modules in order.
// It checks IsInstalled() before calling Install() to ensure idempotency.
// If dryRun is true, it logs what would happen without actually executing Install().
// Returns an error if any module fails to execute or if a module name is not found.
func (r *Runner) RunModules(names []string, cfg *config.Config, dryRun bool) error {
	if len(names) == 0 {
		return fmt.Errorf("no modules specified")
	}

	var errors []error

	for _, name := range names {
		mod, exists := r.modules[name]
		if !exists {
			err := fmt.Errorf("module %s not found in registry", name)
			log.Error("Failed to find module: %s", name)
			errors = append(errors, err)
			continue
		}

		log.Info("Processing module: %s", name)

		if dryRun {
			// In dry-run mode, check IsInstalled but don't call Install
			installed, err := mod.IsInstalled()
			if err != nil {
				log.Error("Failed to check if module %s is installed: %v", name, err)
				errors = append(errors, fmt.Errorf("module %s: %w", name, err))
				continue
			}

			if installed {
				log.Skip("Module %s is already installed (dry-run)", name)
			} else {
				log.Info("Would install module %s (dry-run)", name)
			}
			continue
		}

		// Check if module is already installed
		installed, err := mod.IsInstalled()
		if err != nil {
			log.Error("Failed to check if module %s is installed: %v", name, err)
			errors = append(errors, fmt.Errorf("module %s: %w", name, err))
			continue
		}

		if installed {
			log.Skip("Module %s is already installed, skipping", name)
			continue
		}

		// Install the module
		log.Info("Installing module: %s", name)
		if err := mod.Install(cfg); err != nil {
			log.Error("Failed to install module %s: %v", name, err)
			errors = append(errors, fmt.Errorf("module %s: %w", name, err))
			continue
		}

		log.Success("Successfully installed module: %s", name)
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to execute %d module(s): %v", len(errors), errors)
	}

	return nil
}

// GetModule returns a module from the registry by name.
// Returns nil if the module is not found.
func (r *Runner) GetModule(name string) module.Module {
	return r.modules[name]
}

// ListModules returns a list of all registered module names.
func (r *Runner) ListModules() []string {
	names := make([]string, 0, len(r.modules))
	for name := range r.modules {
		names = append(names, name)
	}
	return names
}
