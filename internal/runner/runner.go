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
// Returns a slice of ModuleResult for each module processed and an error if any module fails.
func (r *Runner) RunModules(names []string, cfg *config.Config, dryRun bool) ([]ModuleResult, error) {
	if len(names) == 0 {
		return nil, fmt.Errorf("no modules specified")
	}

	var errors []error
	results := make([]ModuleResult, 0, len(names))

	for _, name := range names {
		mod, exists := r.modules[name]
		if !exists {
			err := fmt.Errorf("module %s not found in registry", name)
			log.Error("Failed to find module: %s", name)
			results = append(results, ModuleResult{
				Name:   name,
				Status: StatusError,
				Error:  err,
			})
			errors = append(errors, err)
			continue
		}

		log.Info("Processing module: %s", name)

		if dryRun {
			// In dry-run mode, check IsInstalled but don't call Install
			installed, err := mod.IsInstalled()
			if err != nil {
				log.Error("Failed to check if module %s is installed: %v", name, err)
				result := ModuleResult{
					Name:   name,
					Status: StatusError,
					Error:  fmt.Errorf("module %s: %w", name, err),
				}
				results = append(results, result)
				errors = append(errors, result.Error)
				continue
			}

			if installed {
				log.Skip("Module %s is already installed (dry-run)", name)
				results = append(results, ModuleResult{
					Name:   name,
					Status: StatusSkipped,
				})
			} else {
				log.Info("Would install module %s (dry-run)", name)
				results = append(results, ModuleResult{
					Name:   name,
					Status: StatusWouldInstall,
				})
			}
			continue
		}

		// Check if module is already installed
		installed, err := mod.IsInstalled()
		if err != nil {
			log.Error("Failed to check if module %s is installed: %v", name, err)
			result := ModuleResult{
				Name:   name,
				Status: StatusError,
				Error:  fmt.Errorf("module %s: %w", name, err),
			}
			results = append(results, result)
			errors = append(errors, result.Error)
			continue
		}

		if installed {
			log.Skip("Module %s is already installed, skipping", name)
			results = append(results, ModuleResult{
				Name:   name,
				Status: StatusSkipped,
			})
			continue
		}

		// Install the module
		log.Info("Installing module: %s", name)
		if err := mod.Install(cfg); err != nil {
			log.Error("Failed to install module %s: %v", name, err)
			result := ModuleResult{
				Name:   name,
				Status: StatusFailed,
				Error:  fmt.Errorf("module %s: %w", name, err),
			}
			results = append(results, result)
			errors = append(errors, result.Error)
			continue
		}

		log.Success("Successfully installed module: %s", name)
		results = append(results, ModuleResult{
			Name:   name,
			Status: StatusInstalled,
		})
	}

	if len(errors) > 0 {
		return results, fmt.Errorf("failed to execute %d module(s): %v", len(errors), errors)
	}

	return results, nil
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
