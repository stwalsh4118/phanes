package monitoring

import (
	"os"
	"strings"
	"testing"

	"github.com/stwalsh4118/phanes/internal/config"
	"github.com/stwalsh4118/phanes/internal/exec"
	"github.com/stwalsh4118/phanes/internal/log"
)

func TestMonitoringModule_Name(t *testing.T) {
	mod := &MonitoringModule{}
	if got := mod.Name(); got != "monitoring" {
		t.Errorf("Name() = %q, want %q", got, "monitoring")
	}
}

func TestMonitoringModule_Description(t *testing.T) {
	mod := &MonitoringModule{}
	want := "Installs and configures Netdata monitoring"
	if got := mod.Description(); got != want {
		t.Errorf("Description() = %q, want %q", got, want)
	}
}

func TestNetdataInstalled(t *testing.T) {
	// Test that netdataInstalled() doesn't panic
	// It may return false if Netdata is not installed
	installed, err := netdataInstalled()
	if err != nil {
		t.Logf("netdataInstalled() returned error (may be expected if Netdata not installed): %v", err)
	}

	// Verify return type
	_ = installed
	_ = err

	// Note: Actual installation status depends on system state, so we just verify
	// the function doesn't panic and handles errors gracefully
}

func TestNetdataServiceRunning(t *testing.T) {
	// Test that netdataServiceRunning() doesn't panic
	// It may return false if Netdata service is not running
	running, err := netdataServiceRunning()
	if err != nil {
		t.Logf("netdataServiceRunning() returned error (may be expected if Netdata not running): %v", err)
	}

	// Verify return type
	_ = running
	_ = err

	// Note: Actual service status depends on system state, so we just verify
	// the function doesn't panic and handles errors gracefully
}

func TestNetdataServiceEnabled(t *testing.T) {
	// Test that netdataServiceEnabled() doesn't panic
	// It may return false if Netdata service is not enabled
	enabled, err := netdataServiceEnabled()
	if err != nil {
		t.Logf("netdataServiceEnabled() returned error (may be expected if Netdata not enabled): %v", err)
	}

	// Verify return type
	_ = enabled
	_ = err

	// Note: Actual service status depends on system state, so we just verify
	// the function doesn't panic and handles errors gracefully
}

func TestNetdataPortAccessible(t *testing.T) {
	// Test that netdataPortAccessible() doesn't panic
	// It may return false if Netdata port is not accessible
	accessible, err := netdataPortAccessible()
	if err != nil {
		t.Logf("netdataPortAccessible() returned error (may be expected if Netdata not running or ss/netstat not available): %v", err)
	}

	// Verify return type
	_ = accessible
	_ = err

	// Note: Actual port accessibility depends on system state, so we just verify
	// the function doesn't panic and handles errors gracefully
}

func TestMonitoringModule_IsInstalled(t *testing.T) {
	mod := &MonitoringModule{}

	// Test that IsInstalled() doesn't panic and handles errors gracefully
	installed, err := mod.IsInstalled()
	if err != nil {
		t.Logf("IsInstalled() returned error (may be expected): %v", err)
	}

	// Verify return type
	_ = installed
	_ = err

	// Note: IsInstalled() checks system state, so actual results depend on
	// whether Netdata is installed and configured. We just verify it doesn't panic.
}

func TestMonitoringModule_Install_DryRun(t *testing.T) {
	mod := &MonitoringModule{}

	// Enable dry-run mode
	originalDryRun := log.IsDryRun()
	log.SetDryRun(true)
	defer log.SetDryRun(originalDryRun)

	cfg := &config.Config{}

	// In dry-run mode, Install() should log but not execute commands
	// We can't easily verify this without mocking, but we can verify
	// it doesn't panic and handles dry-run mode
	err := mod.Install(cfg)
	// In dry-run mode, it should not actually install Netdata
	// Error is acceptable if commands fail (since we're not running as root)
	if err != nil {
		// Error is acceptable in dry-run mode if system tools are not available
		// The important thing is it doesn't panic
		t.Logf("Install() in dry-run mode returned error (may be expected): %v", err)
	}
}

func TestMonitoringModule_ModuleInterface(t *testing.T) {
	// Verify that MonitoringModule implements the Module interface
	var _ interface {
		Name() string
		Description() string
		IsInstalled() (bool, error)
		Install(*config.Config) error
	} = &MonitoringModule{}
}

func TestMonitoringModule_Install_RequiresRoot(t *testing.T) {
	// Skip this test in non-root environments
	// Installing requires root privileges
	if os.Geteuid() != 0 {
		t.Skip("Skipping Install test - requires root privileges")
	}

	// Check if required commands exist
	if !exec.CommandExists("curl") {
		t.Skip("curl not available")
	}
	if !exec.CommandExists("bash") {
		t.Skip("bash not available")
	}

	// This test would actually install Netdata, so we skip it
	// In a real scenario, you'd want to:
	// 1. Backup current Netdata configuration
	// 2. Run Install()
	// 3. Verify Netdata is installed and configured correctly
	// 4. Restore backup
	t.Skip("Skipping Install test - would modify system configuration")
}

func TestMonitoringModule_Install_Idempotency(t *testing.T) {
	// Test that Install() is idempotent
	// This would require running Install() twice and verifying
	// the second run doesn't create duplicates or errors
	// We skip this test as it requires root and would modify system
	if os.Geteuid() != 0 {
		t.Skip("Skipping idempotency test - requires root privileges")
	}
	t.Skip("Skipping idempotency test - would modify system configuration")
}

func TestNetdataPortAccessible_EdgeCases(t *testing.T) {
	// Test that netdataPortAccessible() handles missing ss/netstat gracefully
	// This is tested implicitly in TestNetdataPortAccessible, but we can verify
	// the function doesn't panic even if both commands fail
	accessible, err := netdataPortAccessible()
	if err != nil {
		// Error is acceptable if ss and netstat are both unavailable
		if !strings.Contains(err.Error(), "failed to check port accessibility") {
			t.Errorf("netdataPortAccessible() should return descriptive error, got: %v", err)
		}
	}
	_ = accessible
}

