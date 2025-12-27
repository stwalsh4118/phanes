package caddy

import (
	"os"
	"strings"
	"testing"

	"github.com/stwalsh4118/phanes/internal/config"
	"github.com/stwalsh4118/phanes/internal/exec"
	"github.com/stwalsh4118/phanes/internal/log"
)

func TestCaddyModule_Name(t *testing.T) {
	mod := &CaddyModule{}
	if got := mod.Name(); got != "caddy" {
		t.Errorf("Name() = %q, want %q", got, "caddy")
	}
}

func TestCaddyModule_Description(t *testing.T) {
	mod := &CaddyModule{}
	want := "Installs and configures Caddy web server with automatic HTTPS"
	if got := mod.Description(); got != want {
		t.Errorf("Description() = %q, want %q", got, want)
	}
}

func TestCaddyInstalled(t *testing.T) {
	// Test that caddyInstalled() doesn't panic
	// It may return false if Caddy is not installed
	installed, err := caddyInstalled()
	if err != nil {
		t.Logf("caddyInstalled() returned error (may be expected if Caddy not installed): %v", err)
	}

	// Verify return type
	_ = installed
	_ = err

	// Note: Actual installation status depends on system state, so we just verify
	// the function doesn't panic and handles errors gracefully
}

func TestCaddyServiceRunning(t *testing.T) {
	// Test that caddyServiceRunning() doesn't panic
	// It may return false if Caddy service is not running
	running, err := caddyServiceRunning()
	if err != nil {
		t.Logf("caddyServiceRunning() returned error (may be expected if Caddy not running): %v", err)
	}

	// Verify return type
	_ = running
	_ = err

	// Note: Actual service status depends on system state, so we just verify
	// the function doesn't panic and handles errors gracefully
}

func TestCaddyServiceEnabled(t *testing.T) {
	// Test that caddyServiceEnabled() doesn't panic
	// It may return false if Caddy service is not enabled
	enabled, err := caddyServiceEnabled()
	if err != nil {
		t.Logf("caddyServiceEnabled() returned error (may be expected if Caddy not enabled): %v", err)
	}

	// Verify return type
	_ = enabled
	_ = err

	// Note: Actual service status depends on system state, so we just verify
	// the function doesn't panic and handles errors gracefully
}

func TestCaddyPortAccessible(t *testing.T) {
	// Test that caddyPortAccessible() doesn't panic
	// It may return false if Caddy port is not accessible
	accessible, err := caddyPortAccessible()
	if err != nil {
		t.Logf("caddyPortAccessible() returned error (may be expected if Caddy not running or ss/netstat not available): %v", err)
	}

	// Verify return type
	_ = accessible
	_ = err

	// Note: Actual port accessibility depends on system state, so we just verify
	// the function doesn't panic and handles errors gracefully
}

func TestPort80InUse(t *testing.T) {
	// Test that port80InUse() doesn't panic
	// It may return false if port 80 is not in use
	inUse, err := port80InUse()
	if err != nil {
		t.Logf("port80InUse() returned error (may be expected if ss/netstat not available): %v", err)
	}

	// Verify return type
	_ = inUse
	_ = err

	// Note: Actual port usage depends on system state, so we just verify
	// the function doesn't panic and handles errors gracefully
}

func TestCaddyfileExists(t *testing.T) {
	// Test that caddyfileExists() doesn't panic
	// It may return false if Caddyfile doesn't exist
	exists, err := caddyfileExists()
	if err != nil {
		t.Logf("caddyfileExists() returned error (may be expected): %v", err)
	}

	// Verify return type
	_ = exists
	_ = err

	// Note: Actual file existence depends on system state, so we just verify
	// the function doesn't panic and handles errors gracefully
}

func TestCaddyModule_IsInstalled(t *testing.T) {
	mod := &CaddyModule{}

	// Test that IsInstalled() doesn't panic and handles errors gracefully
	installed, err := mod.IsInstalled()
	if err != nil {
		t.Logf("IsInstalled() returned error (may be expected): %v", err)
	}

	// Verify return type
	_ = installed
	_ = err

	// Note: IsInstalled() checks system state, so actual results depend on
	// whether Caddy is installed and configured. We just verify it doesn't panic.
}

func TestCaddyModule_Install_Disabled(t *testing.T) {
	mod := &CaddyModule{}

	cfg := &config.Config{
		Caddy: config.Caddy{
			Enabled: false,
		},
	}

	// Install should skip when disabled
	err := mod.Install(cfg)
	if err != nil {
		t.Errorf("Install() should not return error when disabled, got: %v", err)
	}
}

func TestCaddyModule_Install_DryRun(t *testing.T) {
	mod := &CaddyModule{}

	// Enable dry-run mode
	originalDryRun := log.IsDryRun()
	log.SetDryRun(true)
	defer log.SetDryRun(originalDryRun)

	cfg := &config.Config{
		Caddy: config.Caddy{
			Enabled: true,
		},
	}

	// In dry-run mode, Install() should log but not execute commands
	// We can't easily verify this without mocking, but we can verify
	// it doesn't panic and handles dry-run mode
	err := mod.Install(cfg)
	// In dry-run mode, it should not actually install Caddy
	// Error is acceptable if commands fail (since we're not running as root)
	if err != nil {
		// Error is acceptable in dry-run mode if system tools are not available
		// The important thing is it doesn't panic
		t.Logf("Install() in dry-run mode returned error (may be expected): %v", err)
	}
}

func TestCaddyModule_ModuleInterface(t *testing.T) {
	// Verify that CaddyModule implements the Module interface
	var _ interface {
		Name() string
		Description() string
		IsInstalled() (bool, error)
		Install(*config.Config) error
	} = &CaddyModule{}
}

func TestCaddyModule_Install_RequiresRoot(t *testing.T) {
	// Skip this test in non-root environments
	// Installing requires root privileges
	if os.Geteuid() != 0 {
		t.Skip("Skipping Install test - requires root privileges")
	}

	// Check if required commands exist
	if !exec.CommandExists("apt-get") {
		t.Skip("apt-get not available")
	}

	// This test would actually install Caddy, so we skip it
	// In a real scenario, you'd want to:
	// 1. Backup current Caddy configuration
	// 2. Run Install()
	// 3. Verify Caddy is installed and configured correctly
	// 4. Restore backup
	t.Skip("Skipping Install test - would modify system configuration")
}

func TestCaddyModule_Install_Idempotency(t *testing.T) {
	// Test that Install() is idempotent
	// This would require running Install() twice and verifying
	// the second run doesn't create duplicates or errors
	// We skip this test as it requires root and would modify system
	if os.Geteuid() != 0 {
		t.Skip("Skipping idempotency test - requires root privileges")
	}
	t.Skip("Skipping idempotency test - would modify system configuration")
}

func TestCaddyPortAccessible_EdgeCases(t *testing.T) {
	// Test that caddyPortAccessible() handles missing ss/netstat gracefully
	// This is tested implicitly in TestCaddyPortAccessible, but we can verify
	// the function doesn't panic even if both commands fail
	accessible, err := caddyPortAccessible()
	if err != nil {
		// Error is acceptable if ss and netstat are both unavailable
		if !strings.Contains(err.Error(), "failed to check port accessibility") {
			t.Errorf("caddyPortAccessible() should return descriptive error, got: %v", err)
		}
	}
	_ = accessible
}

func TestPort80InUse_EdgeCases(t *testing.T) {
	// Test that port80InUse() handles missing ss/netstat gracefully
	inUse, err := port80InUse()
	if err != nil {
		// Error is acceptable if ss and netstat are both unavailable
		if !strings.Contains(err.Error(), "failed to check port 80 usage") {
			t.Errorf("port80InUse() should return descriptive error, got: %v", err)
		}
	}
	_ = inUse
}


