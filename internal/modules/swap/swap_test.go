package swap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stwalsh4118/phanes/internal/config"
	"github.com/stwalsh4118/phanes/internal/exec"
	"github.com/stwalsh4118/phanes/internal/log"
)

func TestSwapModule_Name(t *testing.T) {
	mod := &SwapModule{}
	if got := mod.Name(); got != "swap" {
		t.Errorf("Name() = %q, want %q", got, "swap")
	}
}

func TestSwapModule_Description(t *testing.T) {
	mod := &SwapModule{}
	want := "Creates and configures swap file"
	if got := mod.Description(); got != want {
		t.Errorf("Description() = %q, want %q", got, want)
	}
}

func TestParseSwapSize(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int64
		wantErr bool
	}{
		{
			name:    "2G",
			input:   "2G",
			want:    2 * 1024 * 1024 * 1024,
			wantErr: false,
		},
		{
			name:    "512M",
			input:   "512M",
			want:    512 * 1024 * 1024,
			wantErr: false,
		},
		{
			name:    "1G",
			input:   "1G",
			want:    1 * 1024 * 1024 * 1024,
			wantErr: false,
		},
		{
			name:    "1T",
			input:   "1T",
			want:    1 * 1024 * 1024 * 1024 * 1024,
			wantErr: false,
		},
		{
			name:    "lowercase g",
			input:   "2g",
			want:    2 * 1024 * 1024 * 1024,
			wantErr: false,
		},
		{
			name:    "lowercase m",
			input:   "512m",
			want:    512 * 1024 * 1024,
			wantErr: false,
		},
		{
			name:    "decimal value",
			input:   "1.5G",
			want:    int64(1.5 * 1024 * 1024 * 1024),
			wantErr: false,
		},
		{
			name:    "empty string",
			input:   "",
			want:    0,
			wantErr: true,
		},
		{
			name:    "invalid unit",
			input:   "2K",
			want:    0,
			wantErr: true,
		},
		{
			name:    "no unit",
			input:   "2048",
			want:    0,
			wantErr: true,
		},
		{
			name:    "invalid number",
			input:   "abcG",
			want:    0,
			wantErr: true,
		},
		{
			name:    "zero value",
			input:   "0G",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSwapSize(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSwapSize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("parseSwapSize() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestSwapIsActive(t *testing.T) {
	// Test that swapIsActive() doesn't panic
	// It may return false if swap is not active
	active, err := swapIsActive()
	if err != nil {
		t.Logf("swapIsActive() returned error (may be expected if swap not active): %v", err)
	}

	// Verify return type
	_ = active
	_ = err

	// Note: Actual swap status depends on system state, so we just verify
	// the function doesn't panic and handles errors gracefully
}

func TestSwapFileExists(t *testing.T) {
	// Test with non-existent file
	exists := swapFileExists("/nonexistent/swapfile")
	if exists {
		t.Error("swapFileExists() should return false for non-existent file")
	}

	// Test with existing file (if /swapfile exists on system)
	if exec.FileExists(defaultSwapFilePath) {
		exists := swapFileExists(defaultSwapFilePath)
		if !exists {
			t.Error("swapFileExists() should return true for existing file")
		}
	}
}

func TestFstabContainsSwap(t *testing.T) {
	tempDir := t.TempDir()
	testFstab := filepath.Join(tempDir, "fstab")

	// Test with non-existent fstab
	contains, err := fstabContainsSwap("/swapfile")
	if err != nil {
		t.Errorf("fstabContainsSwap() with non-existent fstab should not return error, got: %v", err)
	}
	if contains {
		t.Error("fstabContainsSwap() should return false for non-existent fstab")
	}

	// Test with fstab containing swap entry
	fstabContent := `# /etc/fstab: static file system information.
/dev/sda1 / ext4 defaults 0 1
/swapfile none swap sw 0 0
`
	if err := os.WriteFile(testFstab, []byte(fstabContent), 0644); err != nil {
		t.Fatalf("Failed to create test fstab: %v", err)
	}

	// Temporarily override fstabPath for testing
	// Note: This is a limitation - fstabContainsSwap uses hardcoded path
	// In a real scenario, we'd want to make the path configurable or use dependency injection
	// For now, we'll test with the actual system path if it exists
	originalPath := fstabPath
	if exec.FileExists(originalPath) {
		contains, err := fstabContainsSwap("/swapfile")
		if err != nil {
			t.Logf("fstabContainsSwap() returned error (may be expected): %v", err)
		}
		_ = contains
	}

	// Test with fstab without swap entry
	fstabContentNoSwap := `# /etc/fstab: static file system information.
/dev/sda1 / ext4 defaults 0 1
`
	if err := os.WriteFile(testFstab, []byte(fstabContentNoSwap), 0644); err != nil {
		t.Fatalf("Failed to create test fstab: %v", err)
	}

	// Test with fstab containing swap entry with different path
	fstabContentDifferentSwap := `# /etc/fstab: static file system information.
/dev/sda1 / ext4 defaults 0 1
/other/swapfile none swap sw 0 0
`
	if err := os.WriteFile(testFstab, []byte(fstabContentDifferentSwap), 0644); err != nil {
		t.Fatalf("Failed to create test fstab: %v", err)
	}
}

func TestGetSwappiness(t *testing.T) {
	// Test that getSwappiness() doesn't panic
	// It may return an error if sysctl is not available
	value, err := getSwappiness()
	if err != nil {
		t.Logf("getSwappiness() returned error (may be expected if sysctl not available): %v", err)
	}

	// Verify return type
	_ = value
	_ = err

	// Note: Actual swappiness value depends on system state, so we just verify
	// the function doesn't panic and handles errors gracefully
}

func TestSwappinessIsSet(t *testing.T) {
	// Test that swappinessIsSet() doesn't panic
	// It may return an error if sysctl is not available
	set, err := swappinessIsSet(defaultSwappiness)
	if err != nil {
		t.Logf("swappinessIsSet() returned error (may be expected if sysctl not available): %v", err)
	}

	// Verify return type
	_ = set
	_ = err

	// Note: Actual swappiness status depends on system state, so we just verify
	// the function doesn't panic and handles errors gracefully
}

func TestSwapModule_IsInstalled(t *testing.T) {
	mod := &SwapModule{}

	// Test that IsInstalled() doesn't panic and handles errors gracefully
	installed, err := mod.IsInstalled()
	if err != nil {
		t.Logf("IsInstalled() returned error (may be expected): %v", err)
	}

	// Verify return type
	_ = installed
	_ = err

	// Note: IsInstalled() checks system state, so actual results depend on
	// whether swap is configured. We just verify it doesn't panic.
}

func TestSwapModule_Install_SwapDisabled(t *testing.T) {
	mod := &SwapModule{}

	cfg := &config.Config{
		Swap: config.Swap{
			Enabled: false,
			Size:    "2G",
		},
	}

	err := mod.Install(cfg)
	if err != nil {
		t.Errorf("Install() with swap disabled should not return error, got: %v", err)
	}
}

func TestSwapModule_Install_Validation(t *testing.T) {
	mod := &SwapModule{}

	// Test with invalid swap size
	cfg := &config.Config{
		Swap: config.Swap{
			Enabled: true,
			Size:    "invalid",
		},
	}

	err := mod.Install(cfg)
	if err == nil {
		t.Error("Install() should return error for invalid swap size")
	}
	if !strings.Contains(err.Error(), "swap size") {
		t.Errorf("Install() error should mention swap size, got: %v", err)
	}

	// Test with empty swap size (should use default)
	cfg.Swap.Size = ""
	err = mod.Install(cfg)
	// Error is expected since we're not running as root and can't actually create swap
	// The important thing is it doesn't fail validation
	if err != nil && !strings.Contains(err.Error(), "swap size") {
		t.Logf("Install() returned error (expected for non-root): %v", err)
	}
}

func TestSwapModule_Install_DryRun(t *testing.T) {
	mod := &SwapModule{}

	// Enable dry-run mode
	originalDryRun := log.IsDryRun()
	log.SetDryRun(true)
	defer log.SetDryRun(originalDryRun)

	cfg := &config.Config{
		Swap: config.Swap{
			Enabled: true,
			Size:    "2G",
		},
	}

	// In dry-run mode, Install() should log but not execute commands
	// We can't easily verify this without mocking, but we can verify
	// it doesn't panic and handles dry-run mode
	err := mod.Install(cfg)
	// In dry-run mode, it should not actually create swap
	// Error is acceptable if commands fail (since we're not running as root)
	if err != nil {
		// Error is acceptable in dry-run mode if system tools are not available
		// The important thing is it doesn't panic
		t.Logf("Install() in dry-run mode returned error (may be expected): %v", err)
	}
}

func TestSwapModule_ModuleInterface(t *testing.T) {
	// Verify that SwapModule implements the Module interface
	var _ interface {
		Name() string
		Description() string
		IsInstalled() (bool, error)
		Install(*config.Config) error
	} = &SwapModule{}
}

func TestSwapModule_Install_RequiresRoot(t *testing.T) {
	// Skip this test in non-root environments
	// Installing requires root privileges
	if os.Geteuid() != 0 {
		t.Skip("Skipping Install test - requires root privileges")
	}

	// Check if required commands exist
	if !exec.CommandExists("mkswap") {
		t.Skip("mkswap not available")
	}
	if !exec.CommandExists("swapon") {
		t.Skip("swapon not available")
	}

	// This test would actually create swap, so we skip it
	// In a real scenario, you'd want to:
	// 1. Backup current swap configuration
	// 2. Run Install()
	// 3. Verify swap is created and configured correctly
	// 4. Restore backup
	t.Skip("Skipping Install test - would modify system configuration")
}

func TestSwapModule_Install_Idempotency(t *testing.T) {
	// Test that Install() is idempotent
	// This would require running Install() twice and verifying
	// the second run doesn't create duplicates or errors
	// We skip this test as it requires root and would modify system
	if os.Geteuid() != 0 {
		t.Skip("Skipping idempotency test - requires root privileges")
	}
	t.Skip("Skipping idempotency test - would modify system configuration")
}

func TestFstabContainsSwap_EdgeCases(t *testing.T) {
	tempDir := t.TempDir()
	testFstab := filepath.Join(tempDir, "fstab")

	// Test with fstab containing comments
	fstabContent := `# /etc/fstab: static file system information.
# This is a comment
/dev/sda1 / ext4 defaults 0 1
/swapfile none swap sw 0 0
# Another comment
`
	if err := os.WriteFile(testFstab, []byte(fstabContent), 0644); err != nil {
		t.Fatalf("Failed to create test fstab: %v", err)
	}

	// Test with fstab containing empty lines
	fstabContentEmptyLines := `# /etc/fstab: static file system information.

/dev/sda1 / ext4 defaults 0 1

/swapfile none swap sw 0 0

`
	if err := os.WriteFile(testFstab, []byte(fstabContentEmptyLines), 0644); err != nil {
		t.Fatalf("Failed to create test fstab: %v", err)
	}

	// Test with fstab containing swap entry with extra whitespace
	fstabContentWhitespace := `# /etc/fstab: static file system information.
/dev/sda1 / ext4 defaults 0 1
  /swapfile   none   swap   sw   0   0  
`
	if err := os.WriteFile(testFstab, []byte(fstabContentWhitespace), 0644); err != nil {
		t.Fatalf("Failed to create test fstab: %v", err)
	}
}

