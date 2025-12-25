package updates

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stwalsh4118/phanes/internal/config"
	"github.com/stwalsh4118/phanes/internal/exec"
	"github.com/stwalsh4118/phanes/internal/log"
)

func TestUpdatesModule_Name(t *testing.T) {
	mod := &UpdatesModule{}
	if got := mod.Name(); got != "updates" {
		t.Errorf("Name() = %q, want %q", got, "updates")
	}
}

func TestUpdatesModule_Description(t *testing.T) {
	mod := &UpdatesModule{}
	want := "Configures automatic security updates"
	if got := mod.Description(); got != want {
		t.Errorf("Description() = %q, want %q", got, want)
	}
}

func TestGenerate50UnattendedUpgrades(t *testing.T) {
	config := generate50UnattendedUpgrades()

	// Check that config contains key settings
	if !strings.Contains(config, "Unattended-Upgrade::Automatic-Reboot \"false\"") {
		t.Error("Config should contain Automatic-Reboot set to false")
	}

	if !strings.Contains(config, "Unattended-Upgrade::Remove-Unused-Dependencies \"true\"") {
		t.Error("Config should contain Remove-Unused-Dependencies set to true")
	}

	if !strings.Contains(config, "${distro_id}:${distro_codename}-security") {
		t.Error("Config should contain security update origins")
	}
}

func TestGenerate20AutoUpgrades(t *testing.T) {
	config := generate20AutoUpgrades()

	// Check that config contains key settings
	if !strings.Contains(config, "APT::Periodic::Update-Package-Lists \"1\"") {
		t.Error("Config should contain Update-Package-Lists set to 1")
	}

	if !strings.Contains(config, "APT::Periodic::Unattended-Upgrade \"1\"") {
		t.Error("Config should contain Unattended-Upgrade set to 1")
	}
}

func TestNormalizeConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "remove comments",
			input:    "Setting \"value\";\n// This is a comment\nAnother \"setting\";",
			expected: "Setting \"value\";\nAnother \"setting\";",
		},
		{
			name:     "remove empty lines",
			input:    "Setting \"value\";\n\nAnother \"setting\";",
			expected: "Setting \"value\";\nAnother \"setting\";",
		},
		{
			name:     "remove inline comments",
			input:    "Setting \"value\"; // inline comment",
			expected: "Setting \"value\";",
		},
		{
			name:     "normalize whitespace",
			input:    "  Setting   \"value\"  ;  ",
			expected: "Setting \"value\";",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeConfig(tt.input)
			if got != tt.expected {
				t.Errorf("normalizeConfig() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestConfigFileMatches(t *testing.T) {
	tempDir := t.TempDir()
	testConfig := filepath.Join(tempDir, "test.conf")

	expectedConfig := `Setting "value";
Another "setting";`

	// Test with non-existent file
	matches, err := configFileMatches(testConfig, expectedConfig)
	if err != nil {
		t.Errorf("configFileMatches() with non-existent file should not return error, got: %v", err)
	}
	if matches {
		t.Error("configFileMatches() should return false for non-existent file")
	}

	// Test with matching config
	if err := os.WriteFile(testConfig, []byte(expectedConfig), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	matches, err = configFileMatches(testConfig, expectedConfig)
	if err != nil {
		t.Errorf("configFileMatches() returned error: %v", err)
	}
	if !matches {
		t.Error("configFileMatches() should return true for matching config")
	}

	// Test with config containing comments (should still match)
	configWithComments := `Setting "value";
// This is a comment
Another "setting";`
	if err := os.WriteFile(testConfig, []byte(configWithComments), 0644); err != nil {
		t.Fatalf("Failed to update test config: %v", err)
	}

	matches, err = configFileMatches(testConfig, expectedConfig)
	if err != nil {
		t.Errorf("configFileMatches() returned error: %v", err)
	}
	if !matches {
		t.Error("configFileMatches() should return true for config with comments")
	}

	// Test with non-matching config
	nonMatchingConfig := `Different "setting";
Another "value";`
	if err := os.WriteFile(testConfig, []byte(nonMatchingConfig), 0644); err != nil {
		t.Fatalf("Failed to update test config: %v", err)
	}

	matches, err = configFileMatches(testConfig, expectedConfig)
	if err != nil {
		t.Errorf("configFileMatches() returned error: %v", err)
	}
	if matches {
		t.Error("configFileMatches() should return false for non-matching config")
	}
}

func TestUnattendedUpgradesInstalled(t *testing.T) {
	// Test that unattendedUpgradesInstalled() doesn't panic
	// It may return false if package is not installed
	installed, err := unattendedUpgradesInstalled()
	if err != nil {
		t.Logf("unattendedUpgradesInstalled() returned error (may be expected if package not installed): %v", err)
	}

	// Verify return type
	_ = installed
	_ = err

	// Note: Actual installation status depends on system state, so we just verify
	// the function doesn't panic and handles errors gracefully
}

func TestUnattendedUpgradesConfigured(t *testing.T) {
	// Test that unattendedUpgradesConfigured() doesn't panic
	// It may return false if configs don't exist or don't match
	configured, err := unattendedUpgradesConfigured()
	if err != nil {
		t.Logf("unattendedUpgradesConfigured() returned error (may be expected): %v", err)
	}

	// Verify return type
	_ = configured
	_ = err

	// Note: Actual configuration status depends on system state, so we just verify
	// the function doesn't panic and handles errors gracefully
}

func TestUpdatesModule_IsInstalled(t *testing.T) {
	mod := &UpdatesModule{}

	// Test that IsInstalled() doesn't panic and handles errors gracefully
	installed, err := mod.IsInstalled()
	if err != nil {
		t.Logf("IsInstalled() returned error (may be expected): %v", err)
	}

	// Verify return type
	_ = installed
	_ = err

	// Note: IsInstalled() checks system state, so actual results depend on
	// whether unattended-upgrades is installed and configured. We just verify it doesn't panic.
}

func TestUpdatesModule_Install_DryRun(t *testing.T) {
	mod := &UpdatesModule{}

	// Enable dry-run mode
	originalDryRun := log.IsDryRun()
	log.SetDryRun(true)
	defer log.SetDryRun(originalDryRun)

	cfg := &config.Config{}

	// In dry-run mode, Install() should log but not execute commands
	// We can't easily verify this without mocking, but we can verify
	// it doesn't panic and handles dry-run mode
	err := mod.Install(cfg)
	// In dry-run mode, it should not actually install packages or write files
	// Error is acceptable if commands fail (since we're not running as root)
	if err != nil {
		// Error is acceptable in dry-run mode if system tools are not available
		// The important thing is it doesn't panic
		t.Logf("Install() in dry-run mode returned error (may be expected): %v", err)
	}
}

func TestUpdatesModule_ModuleInterface(t *testing.T) {
	// Verify that UpdatesModule implements the Module interface
	var _ interface {
		Name() string
		Description() string
		IsInstalled() (bool, error)
		Install(*config.Config) error
	} = &UpdatesModule{}
}

func TestUpdatesModule_Install_RequiresRoot(t *testing.T) {
	// Skip this test in non-root environments
	// Installing requires root privileges
	if os.Geteuid() != 0 {
		t.Skip("Skipping Install test - requires root privileges")
	}

	// Check if required commands exist
	if !exec.CommandExists("apt-get") {
		t.Skip("apt-get not available")
	}

	// This test would actually install unattended-upgrades, so we skip it
	// In a real scenario, you'd want to:
	// 1. Backup current configuration
	// 2. Run Install()
	// 3. Verify unattended-upgrades is installed and configured correctly
	// 4. Restore backup
	t.Skip("Skipping Install test - would modify system configuration")
}

func TestUpdatesModule_Install_Idempotency(t *testing.T) {
	// Test that Install() is idempotent
	// This would require running Install() twice and verifying
	// the second run doesn't create duplicates or errors
	// We skip this test as it requires root and would modify system
	if os.Geteuid() != 0 {
		t.Skip("Skipping idempotency test - requires root privileges")
	}
	t.Skip("Skipping idempotency test - would modify system configuration")
}

func TestConfigFileMatches_EdgeCases(t *testing.T) {
	tempDir := t.TempDir()
	testConfig := filepath.Join(tempDir, "test.conf")

	expectedConfig := `Setting "value";
Another "setting";`

	// Test with config containing only comments
	configOnlyComments := `// This is a comment
// Another comment`
	if err := os.WriteFile(testConfig, []byte(configOnlyComments), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	matches, err := configFileMatches(testConfig, expectedConfig)
	if err != nil {
		t.Errorf("configFileMatches() returned error: %v", err)
	}
	if matches {
		t.Error("configFileMatches() should return false for config with only comments")
	}

	// Test with empty config file
	if err := os.WriteFile(testConfig, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	matches, err = configFileMatches(testConfig, expectedConfig)
	if err != nil {
		t.Errorf("configFileMatches() returned error: %v", err)
	}
	if matches {
		t.Error("configFileMatches() should return false for empty config")
	}

	// Test with config containing extra whitespace (should still match after normalization)
	configWithWhitespace := `  Setting   "value"  ;  
	
	Another   "setting"  ;  `
	if err := os.WriteFile(testConfig, []byte(configWithWhitespace), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	matches, err = configFileMatches(testConfig, expectedConfig)
	if err != nil {
		t.Errorf("configFileMatches() returned error: %v", err)
	}
	if !matches {
		t.Error("configFileMatches() should return true for config with extra whitespace (after normalization)")
	}
}

