package nginx

import (
	"os"
	"strings"
	"testing"

	"github.com/stwalsh4118/phanes/internal/config"
	"github.com/stwalsh4118/phanes/internal/exec"
	"github.com/stwalsh4118/phanes/internal/log"
)

func TestNginxModule_Name(t *testing.T) {
	mod := &NginxModule{}
	if got := mod.Name(); got != "nginx" {
		t.Errorf("Name() = %q, want %q", got, "nginx")
	}
}

func TestNginxModule_Description(t *testing.T) {
	mod := &NginxModule{}
	want := "Installs and configures Nginx web server"
	if got := mod.Description(); got != want {
		t.Errorf("Description() = %q, want %q", got, want)
	}
}

func TestNginxInstalled(t *testing.T) {
	// Test that nginxInstalled() doesn't panic
	// It may return false if Nginx is not installed
	installed, err := nginxInstalled()
	if err != nil {
		t.Logf("nginxInstalled() returned error (may be expected if Nginx not installed): %v", err)
	}

	// Verify return type
	_ = installed
	_ = err

	// Note: Actual installation status depends on system state, so we just verify
	// the function doesn't panic and handles errors gracefully
}

func TestNginxServiceRunning(t *testing.T) {
	// Test that nginxServiceRunning() doesn't panic
	// It may return false if Nginx service is not running
	running, err := nginxServiceRunning()
	if err != nil {
		t.Logf("nginxServiceRunning() returned error (may be expected if Nginx not running): %v", err)
	}

	// Verify return type
	_ = running
	_ = err

	// Note: Actual service status depends on system state, so we just verify
	// the function doesn't panic and handles errors gracefully
}

func TestNginxServiceEnabled(t *testing.T) {
	// Test that nginxServiceEnabled() doesn't panic
	// It may return false if Nginx service is not enabled
	enabled, err := nginxServiceEnabled()
	if err != nil {
		t.Logf("nginxServiceEnabled() returned error (may be expected if Nginx not enabled): %v", err)
	}

	// Verify return type
	_ = enabled
	_ = err

	// Note: Actual service status depends on system state, so we just verify
	// the function doesn't panic and handles errors gracefully
}

func TestNginxPortAccessible(t *testing.T) {
	// Test that nginxPortAccessible() doesn't panic
	// It may return false if Nginx port is not accessible
	accessible, err := nginxPortAccessible()
	if err != nil {
		t.Logf("nginxPortAccessible() returned error (may be expected if Nginx not running or ss/netstat not available): %v", err)
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

func TestNginxModule_IsInstalled(t *testing.T) {
	mod := &NginxModule{}

	// Test that IsInstalled() doesn't panic and handles errors gracefully
	installed, err := mod.IsInstalled()
	if err != nil {
		t.Logf("IsInstalled() returned error (may be expected): %v", err)
	}

	// Verify return type
	_ = installed
	_ = err

	// Note: IsInstalled() checks system state, so actual results depend on
	// whether Nginx is installed and configured. We just verify it doesn't panic.
}

func TestNginxModule_Install_Disabled(t *testing.T) {
	mod := &NginxModule{}

	cfg := &config.Config{
		Nginx: config.Nginx{
			Enabled: false,
		},
	}

	// Install should skip when disabled
	err := mod.Install(cfg)
	if err != nil {
		t.Errorf("Install() should not return error when disabled, got: %v", err)
	}
}

func TestNginxModule_Install_DryRun(t *testing.T) {
	mod := &NginxModule{}

	// Enable dry-run mode
	originalDryRun := log.IsDryRun()
	log.SetDryRun(true)
	defer log.SetDryRun(originalDryRun)

	cfg := &config.Config{
		Nginx: config.Nginx{
			Enabled: true,
		},
	}

	// In dry-run mode, Install() should log but not execute commands
	// We can't easily verify this without mocking, but we can verify
	// it doesn't panic and handles dry-run mode
	err := mod.Install(cfg)
	// In dry-run mode, it should not actually install Nginx
	// Error is acceptable if commands fail (since we're not running as root)
	if err != nil {
		// Error is acceptable in dry-run mode if system tools are not available
		// The important thing is it doesn't panic
		t.Logf("Install() in dry-run mode returned error (may be expected): %v", err)
	}
}

func TestNginxModule_ModuleInterface(t *testing.T) {
	// Verify that NginxModule implements the Module interface
	var _ interface {
		Name() string
		Description() string
		IsInstalled() (bool, error)
		Install(*config.Config) error
	} = &NginxModule{}
}

func TestNginxModule_Install_RequiresRoot(t *testing.T) {
	// Skip this test in non-root environments
	// Installing requires root privileges
	if os.Geteuid() != 0 {
		t.Skip("Skipping Install test - requires root privileges")
	}

	// Check if required commands exist
	if !exec.CommandExists("apt-get") {
		t.Skip("apt-get not available")
	}

	// This test would actually install Nginx, so we skip it
	// In a real scenario, you'd want to:
	// 1. Backup current Nginx configuration
	// 2. Run Install()
	// 3. Verify Nginx is installed and configured correctly
	// 4. Restore backup
	t.Skip("Skipping Install test - would modify system configuration")
}

func TestNginxModule_Install_Idempotency(t *testing.T) {
	// Test that Install() is idempotent
	// This would require running Install() twice and verifying
	// the second run doesn't create duplicates or errors
	// We skip this test as it requires root and would modify system
	if os.Geteuid() != 0 {
		t.Skip("Skipping idempotency test - requires root privileges")
	}
	t.Skip("Skipping idempotency test - would modify system configuration")
}

func TestNginxPortAccessible_EdgeCases(t *testing.T) {
	// Test that nginxPortAccessible() handles missing ss/netstat gracefully
	// This is tested implicitly in TestNginxPortAccessible, but we can verify
	// the function doesn't panic even if both commands fail
	accessible, err := nginxPortAccessible()
	if err != nil {
		// Error is acceptable if ss and netstat are both unavailable
		if !strings.Contains(err.Error(), "failed to check port accessibility") {
			t.Errorf("nginxPortAccessible() should return descriptive error, got: %v", err)
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


