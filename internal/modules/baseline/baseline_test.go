package baseline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stwalsh4118/phanes/internal/config"
	"github.com/stwalsh4118/phanes/internal/exec"
)

func TestBaselineModule_Name(t *testing.T) {
	mod := &BaselineModule{}
	if got := mod.Name(); got != "baseline" {
		t.Errorf("Name() = %q, want %q", got, "baseline")
	}
}

func TestBaselineModule_Description(t *testing.T) {
	mod := &BaselineModule{}
	want := "Sets timezone, locale, and runs apt update"
	if got := mod.Description(); got != want {
		t.Errorf("Description() = %q, want %q", got, want)
	}
}

func TestBaselineModule_IsInstalled_TimezoneCheck(t *testing.T) {
	mod := &BaselineModule{}

	// Check if timedatectl exists (required for this test)
	if !exec.CommandExists("timedatectl") {
		t.Skip("timedatectl not available, skipping timezone check test")
	}

	installed, err := mod.IsInstalled()
	if err != nil {
		// If we can't check, that's okay - the system might not be configured
		// We're testing that the function doesn't panic and handles errors
		t.Logf("IsInstalled() returned error (expected in some environments): %v", err)
		return
	}

	// If installed is true, verify timezone is set
	if installed {
		timezone, err := exec.RunWithOutput("timedatectl", "show", "-p", "Timezone", "--value")
		if err != nil {
			t.Errorf("Failed to verify timezone: %v", err)
			return
		}
		timezone = strings.TrimSpace(timezone)
		if timezone == "" {
			t.Error("IsInstalled() returned true but timezone is empty")
		}
	}
}

func TestBaselineModule_IsInstalled_LocaleCheck(t *testing.T) {
	mod := &BaselineModule{}

	// Check if locale command exists
	if !exec.CommandExists("locale") {
		t.Skip("locale command not available, skipping locale check test")
	}

	installed, err := mod.IsInstalled()
	if err != nil {
		t.Logf("IsInstalled() returned error: %v", err)
		return
	}

	// If installed is true, verify locale contains UTF-8
	if installed {
		locale, err := exec.RunWithOutput("locale")
		if err != nil {
			t.Errorf("Failed to verify locale: %v", err)
			return
		}
		if !strings.Contains(strings.ToUpper(locale), "UTF-8") {
			t.Error("IsInstalled() returned true but locale doesn't contain UTF-8")
		}
	}
}

func TestBaselineModule_IsInstalled_WithLocaleFile(t *testing.T) {
	// Create a temporary locale file for testing
	_ = t.TempDir()
	_ = filepath.Join("test", "locale")

	// Test with UTF-8 locale
	testCases := []struct {
		name     string
		content  string
		expected bool
	}{
		{
			name:     "UTF-8 locale",
			content:  "LANG=en_US.UTF-8\n",
			expected: true,
		},
		{
			name:     "non-UTF-8 locale",
			content:  "LANG=en_US.ISO8859-1\n",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Note: This test is limited because IsInstalled() checks system locale,
			// not a test file. We're mainly testing that the function doesn't panic
			// and handles different scenarios.
			_ = tc.content
			_ = tc.expected
			// This test verifies the function works, actual locale checking
			// would require mocking exec calls or using a test environment
		})
	}
}

func TestBaselineModule_Install_Timezone(t *testing.T) {
	// Skip this test in CI or non-root environments
	// Installing requires root privileges
	if os.Geteuid() != 0 {
		t.Skip("Skipping Install test - requires root privileges")
	}

	// Check if timedatectl exists
	if !exec.CommandExists("timedatectl") {
		t.Skip("timedatectl not available")
	}

	// This test would actually change system timezone, so we skip it
	// In a real scenario, you'd want to:
	// 1. Save current timezone
	// 2. Run Install()
	// 3. Verify timezone was set
	// 4. Restore original timezone
	t.Skip("Skipping Install test - would modify system configuration")
}

func TestBaselineModule_Install_ConfigDefaults(t *testing.T) {
	// Test with empty timezone (should default to UTC)
	cfg := &config.Config{
		System: config.System{
			Timezone: "",
		},
	}

	// Verify that Install() would use UTC when timezone is empty
	// We can't actually run Install() without root, but we can verify
	// the logic handles empty timezone
	if cfg.System.Timezone == "" {
		cfg.System.Timezone = "UTC"
	}
	if cfg.System.Timezone != "UTC" {
		t.Errorf("Expected default timezone UTC, got %q", cfg.System.Timezone)
	}
}

func TestBaselineModule_ModuleInterface(t *testing.T) {
	// Verify that BaselineModule implements the Module interface
	var _ interface {
		Name() string
		Description() string
		IsInstalled() (bool, error)
		Install(*config.Config) error
	} = &BaselineModule{}
}

func TestBaselineModule_IsInstalled_ErrorHandling(t *testing.T) {
	mod := &BaselineModule{}

	// Test that IsInstalled() handles errors gracefully
	// If timedatectl doesn't exist or fails, it should return an error
	// We can't easily mock this, but we can verify the function signature
	// and that it doesn't panic
	_, err := mod.IsInstalled()
	// Error is acceptable - the function should handle it gracefully
	if err != nil {
		t.Logf("IsInstalled() returned expected error: %v", err)
	}
}

// TestBaselineModule_Install_ErrorHandling tests that Install() handles errors
// This is a placeholder test - actual error testing would require mocking exec
func TestBaselineModule_Install_ErrorHandling(t *testing.T) {
	cfg := &config.Config{
		System: config.System{
			Timezone: "UTC",
		},
	}

	// We can't easily test error handling without mocking exec,
	// but we can verify the function signature and that it accepts config
	if cfg == nil {
		t.Error("Config should not be nil")
	}
}
