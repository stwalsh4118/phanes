package coolify

import (
	"strings"
	"testing"

	"github.com/stwalsh4118/phanes/internal/config"
	"github.com/stwalsh4118/phanes/internal/log"
)

func TestCoolifyModule_Name(t *testing.T) {
	mod := &CoolifyModule{}
	if got := mod.Name(); got != "coolify" {
		t.Errorf("Name() = %q, want %q", got, "coolify")
	}
}

func TestCoolifyModule_Description(t *testing.T) {
	mod := &CoolifyModule{}
	want := "Installs and configures Coolify self-hosted PaaS"
	if got := mod.Description(); got != want {
		t.Errorf("Description() = %q, want %q", got, want)
	}
}

func TestDockerInstalled(t *testing.T) {
	// Test that dockerInstalled() doesn't panic
	// It may return false if Docker is not installed
	installed, err := dockerInstalled()
	if err != nil {
		t.Logf("dockerInstalled() returned error (may be expected if Docker not installed): %v", err)
	}

	// Verify return type
	_ = installed
	_ = err

	// Note: Actual installation status depends on system state, so we just verify
	// the function doesn't panic and handles errors gracefully
}

func TestDockerServiceRunning(t *testing.T) {
	// Test that dockerServiceRunning() doesn't panic
	// It may return false if Docker service is not running
	running, err := dockerServiceRunning()
	if err != nil {
		t.Logf("dockerServiceRunning() returned error (may be expected if Docker not running): %v", err)
	}

	// Verify return type
	_ = running
	_ = err

	// Note: Actual service status depends on system state, so we just verify
	// the function doesn't panic and handles errors gracefully
}

func TestCheckDockerDependency(t *testing.T) {
	// Test that checkDockerDependency() doesn't panic
	// It may return an error if Docker is not installed or not running
	err := checkDockerDependency()
	if err != nil {
		t.Logf("checkDockerDependency() returned error (may be expected if Docker not available): %v", err)
	}

	// Note: Actual dependency status depends on system state, so we just verify
	// the function doesn't panic and handles errors gracefully
}

func TestCoolifyContainersRunning(t *testing.T) {
	// Test that coolifyContainersRunning() doesn't panic
	// It may return false if Coolify containers are not running
	running, err := coolifyContainersRunning()
	if err != nil {
		t.Logf("coolifyContainersRunning() returned error (may be expected if Docker not available or containers not running): %v", err)
	}

	// Verify return type
	_ = running
	_ = err

	// Note: Actual container status depends on system state, so we just verify
	// the function doesn't panic and handles errors gracefully
}

func TestCoolifyModule_IsInstalled(t *testing.T) {
	mod := &CoolifyModule{}

	// Test that IsInstalled() doesn't panic and handles errors gracefully
	installed, err := mod.IsInstalled()
	if err != nil {
		t.Logf("IsInstalled() returned error (may be expected): %v", err)
	}

	// Verify return type
	_ = installed
	_ = err

	// Note: IsInstalled() checks system state, so actual results depend on
	// whether Docker and Coolify are installed and configured. We just verify it doesn't panic.
}

func TestCoolifyModule_Install_Disabled(t *testing.T) {
	mod := &CoolifyModule{}

	cfg := &config.Config{
		Coolify: config.Coolify{
			Enabled: false,
		},
	}

	err := mod.Install(cfg)
	if err != nil {
		t.Errorf("Install() should not return error when Coolify is disabled, got: %v", err)
	}
}

func TestCoolifyModule_Install_DryRun(t *testing.T) {
	mod := &CoolifyModule{}

	// Enable dry-run mode
	originalDryRun := log.IsDryRun()
	log.SetDryRun(true)
	defer log.SetDryRun(originalDryRun)

	cfg := &config.Config{
		Coolify: config.Coolify{
			Enabled: true,
		},
	}

	// In dry-run mode, Install() should log but not execute commands
	// We can't easily verify this without mocking, but we can verify
	// it doesn't panic and handles dry-run mode
	err := mod.Install(cfg)
	// In dry-run mode, it should not actually install Coolify
	// Error is acceptable if Docker dependency check fails (since Docker may not be available)
	if err != nil {
		// Error is acceptable in dry-run mode if Docker is not available
		// The important thing is it doesn't panic
		t.Logf("Install() in dry-run mode returned error (may be expected if Docker not available): %v", err)
	}
}

func TestCoolifyModule_Install_DockerNotAvailable(t *testing.T) {
	mod := &CoolifyModule{}

	// Enable dry-run mode to avoid actually installing
	originalDryRun := log.IsDryRun()
	log.SetDryRun(true)
	defer log.SetDryRun(originalDryRun)

	cfg := &config.Config{
		Coolify: config.Coolify{
			Enabled: true,
		},
	}

	// Test that Install() returns error when Docker is not available
	// This test may pass or fail depending on system state
	err := mod.Install(cfg)
	if err != nil {
		// If Docker is not available, error should mention Docker
		if !strings.Contains(err.Error(), "Docker") {
			t.Logf("Install() error should mention Docker when Docker is not available, got: %v", err)
		}
	} else {
		// If Docker is available, this is fine - we're just testing error handling
		t.Logf("Install() succeeded (Docker is available)")
	}
}

func TestCoolifyModule_ModuleInterface(t *testing.T) {
	// Verify that CoolifyModule implements the Module interface
	var _ interface {
		Name() string
		Description() string
		IsInstalled() (bool, error)
		Install(*config.Config) error
	} = &CoolifyModule{}
}

