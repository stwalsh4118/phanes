package runner

import "time"

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
