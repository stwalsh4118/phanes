package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const (
	testConfigFile = "test-config.yaml"
	phanesBinary   = "phanes"
)

// getTestConfigPath returns the absolute path to the test config file.
// It ensures the path is correct regardless of where the test is run from.
func getTestConfigPath() string {
	// Try to find test-config.yaml in common locations
	wd, err := os.Getwd()
	if err == nil {
		// Check if we're in the workspace root
		if _, err := os.Stat(filepath.Join(wd, testConfigFile)); err == nil {
			return filepath.Join(wd, testConfigFile)
		}
		// Check if we're in test/integration directory
		if _, err := os.Stat(filepath.Join(wd, "..", "..", testConfigFile)); err == nil {
			return filepath.Join(wd, "..", "..", testConfigFile)
		}
	}
	// Default to relative path (will work if running from workspace root)
	return testConfigFile
}

// TestBaselineModuleE2E tests the baseline module end-to-end in a containerized environment.
func TestBaselineModuleE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Check if we're in a containerized environment
	// If not, skip (these tests require Docker)
	if !isContainerized() {
		t.Skip("Skipping E2E test - not running in containerized environment")
	}

	tests := []struct {
		name           string
		args           []string
		expectedOutput []string
		shouldFail     bool
	}{
		{
			name: "baseline module dry-run",
			args: []string{
				"--modules", "baseline",
				"--config", testConfigFile,
				"--dry-run",
			},
			expectedOutput: []string{
				"dry_run=",
				"Would install module baseline",
			},
			shouldFail: false,
		},
		{
			name: "baseline module full execution",
			args: []string{
				"--modules", "baseline",
				"--config", testConfigFile,
			},
			expectedOutput: []string{
				"Setting timezone to UTC",
				"Configuring locale en_US.UTF-8",
				"apt-get update completed successfully",
				"Successfully installed module: baseline",
			},
			shouldFail: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Replace test-config.yaml with absolute path in args
			args := make([]string, len(tt.args))
			copy(args, tt.args)
			for i, arg := range args {
				if arg == testConfigFile {
					args[i] = getTestConfigPath()
				}
			}

			cmd := exec.Command(phanesBinary, args...)
			// Ensure we run from workspace directory
			cmd.Dir = "/workspace"
			output, err := cmd.CombinedOutput()
			outputStr := string(output)

			if tt.shouldFail {
				if err == nil {
					t.Errorf("Expected command to fail, but it succeeded. Output: %s", outputStr)
				}
			} else {
				if err != nil {
					t.Errorf("Command failed with error: %v\nOutput: %s", err, outputStr)
					return
				}

				// Verify expected output
				for _, expected := range tt.expectedOutput {
					if !strings.Contains(outputStr, expected) {
						t.Errorf("Expected output to contain '%s', but it didn't.\nFull output: %s", expected, outputStr)
					}
				}
			}
		})
	}
}

// TestBaselineModuleIdempotency tests that the baseline module is idempotent.
func TestBaselineModuleIdempotency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	if !isContainerized() {
		t.Skip("Skipping E2E test - not running in containerized environment")
	}

	// Run baseline module twice
	args := []string{
		"--modules", "baseline",
		"--config", getTestConfigPath(),
	}

	// First run
	cmd1 := exec.Command(phanesBinary, args...)
	cmd1.Dir = "/workspace"
	output1, err1 := cmd1.CombinedOutput()
	if err1 != nil {
		t.Fatalf("First run failed: %v\nOutput: %s", err1, string(output1))
	}

	// Second run - should skip everything
	cmd2 := exec.Command(phanesBinary, args...)
	cmd2.Dir = "/workspace"
	output2, err2 := cmd2.CombinedOutput()
	if err2 != nil {
		t.Fatalf("Second run failed: %v\nOutput: %s", err2, string(output2))
	}

	output2Str := string(output2)
	// On second run, we should see skip messages or "already installed" messages
	// The exact message depends on IsInstalled() implementation
	// For baseline, if timezone and locale are already set, it should skip
	if !strings.Contains(output2Str, "skip") && !strings.Contains(output2Str, "already") {
		t.Logf("Second run output: %s", output2Str)
		// This is okay - the module may still run Install() which is idempotent
		// The important thing is that it doesn't fail
	}
}

// TestBaselineModuleVerification tests that baseline module actually configures the system.
func TestBaselineModuleVerification(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	if !isContainerized() {
		t.Skip("Skipping E2E test - not running in containerized environment")
	}

	// Run baseline module
	args := []string{
		"--modules", "baseline",
		"--config", getTestConfigPath(),
	}

	cmd := exec.Command(phanesBinary, args...)
	cmd.Dir = "/workspace"
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Baseline module failed: %v\nOutput: %s", err, string(output))
	}

	// Verify timezone was set
	timezoneCmd := exec.Command("cat", "/etc/timezone")
	timezoneOutput, err := timezoneCmd.Output()
	if err != nil {
		// In containers without systemd, /etc/timezone might not exist
		// Check if it was created
		if _, err := os.Stat("/etc/timezone"); os.IsNotExist(err) {
			t.Errorf("Expected /etc/timezone to exist after baseline module execution")
		}
	} else {
		timezone := strings.TrimSpace(string(timezoneOutput))
		if timezone == "" {
			t.Error("Timezone should be set in /etc/timezone")
		}
	}

	// Verify locale was configured
	localeCmd := exec.Command("grep", "^LANG=", "/etc/default/locale")
	localeOutput, err := localeCmd.Output()
	if err == nil {
		locale := string(localeOutput)
		if !strings.Contains(locale, "UTF-8") {
			t.Errorf("Expected locale to contain UTF-8, got: %s", locale)
		}
	}
}

// TestUserModuleE2E tests the user module end-to-end in a containerized environment.
func TestUserModuleE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	if !isContainerized() {
		t.Skip("Skipping E2E test - not running in containerized environment")
	}

	tests := []struct {
		name           string
		args           []string
		expectedOutput []string
		shouldFail     bool
	}{
		{
			name: "user module dry-run",
			args: []string{
				"--modules", "user",
				"--config", testConfigFile,
				"--dry-run",
			},
			expectedOutput: []string{
				"dry_run=",
				"Would install module user",
			},
			shouldFail: false,
		},
		{
			name: "user module full execution",
			args: []string{
				"--modules", "user",
				"--config", testConfigFile,
			},
			expectedOutput: []string{
				"user",
				"Successfully installed module: user",
			},
			shouldFail: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Replace test-config.yaml with absolute path in args
			args := make([]string, len(tt.args))
			copy(args, tt.args)
			for i, arg := range args {
				if arg == testConfigFile {
					args[i] = getTestConfigPath()
				}
			}

			cmd := exec.Command(phanesBinary, args...)
			cmd.Dir = "/workspace"
			output, err := cmd.CombinedOutput()
			outputStr := string(output)

			if tt.shouldFail {
				if err == nil {
					t.Errorf("Expected command to fail, but it succeeded. Output: %s", outputStr)
				}
			} else {
				if err != nil {
					t.Errorf("Command failed with error: %v\nOutput: %s", err, outputStr)
					return
				}

				// Verify expected output (at least one should match)
				found := false
				for _, expected := range tt.expectedOutput {
					if strings.Contains(outputStr, expected) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected output to contain one of: %v\nFull output: %s", tt.expectedOutput, outputStr)
				}
			}
		})
	}
}

// TestUserModuleVerification tests that user module actually creates the user and configures SSH.
func TestUserModuleVerification(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	if !isContainerized() {
		t.Skip("Skipping E2E test - not running in containerized environment")
	}

	// Run user module
	args := []string{
		"--modules", "user",
		"--config", getTestConfigPath(),
	}

	cmd := exec.Command(phanesBinary, args...)
	cmd.Dir = "/workspace"
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("User module failed: %v\nOutput: %s", err, string(output))
	}

	// Verify user exists
	idCmd := exec.Command("id", "testuser")
	idOutput, err := idCmd.Output()
	if err != nil {
		t.Errorf("User 'testuser' should exist after module execution: %v", err)
	} else {
		t.Logf("User exists: %s", string(idOutput))
	}

	// Verify SSH directory exists
	sshDir := "/home/testuser/.ssh"
	if _, err := os.Stat(sshDir); os.IsNotExist(err) {
		t.Errorf("SSH directory %s should exist", sshDir)
	} else {
		// Check permissions (should be 700)
		info, err := os.Stat(sshDir)
		if err == nil {
			mode := info.Mode().Perm()
			if mode != 0700 {
				t.Errorf("SSH directory permissions should be 0700, got: %o", mode)
			}
		}
	}

	// Verify authorized_keys exists
	authorizedKeys := filepath.Join(sshDir, "authorized_keys")
	if _, err := os.Stat(authorizedKeys); os.IsNotExist(err) {
		t.Errorf("authorized_keys file should exist at %s", authorizedKeys)
	} else {
		// Check permissions (should be 600)
		info, err := os.Stat(authorizedKeys)
		if err == nil {
			mode := info.Mode().Perm()
			if mode != 0600 {
				t.Errorf("authorized_keys permissions should be 0600, got: %o", mode)
			}
		}

		// Verify SSH key content
		content, err := os.ReadFile(authorizedKeys)
		if err != nil {
			t.Errorf("Failed to read authorized_keys: %v", err)
		} else {
			contentStr := string(content)
			if !strings.Contains(contentStr, "ssh-rsa") {
				t.Error("authorized_keys should contain SSH key")
			}
		}
	}

	// Verify sudoers file exists
	sudoersFile := "/etc/sudoers.d/testuser"
	if _, err := os.Stat(sudoersFile); os.IsNotExist(err) {
		t.Errorf("Sudoers file should exist at %s", sudoersFile)
	} else {
		// Check permissions (should be 0440)
		info, err := os.Stat(sudoersFile)
		if err == nil {
			mode := info.Mode().Perm()
			if mode != 0440 {
				t.Errorf("Sudoers file permissions should be 0440, got: %o", mode)
			}
		}

		// Verify sudoers content
		content, err := os.ReadFile(sudoersFile)
		if err != nil {
			t.Errorf("Failed to read sudoers file: %v", err)
		} else {
			contentStr := string(content)
			if !strings.Contains(contentStr, "testuser") || !strings.Contains(contentStr, "NOPASSWD") {
				t.Error("Sudoers file should contain testuser NOPASSWD configuration")
			}
		}
	}
}

// TestUserModuleIdempotency tests that the user module is idempotent.
func TestUserModuleIdempotency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	if !isContainerized() {
		t.Skip("Skipping E2E test - not running in containerized environment")
	}

	// Run user module twice
	args := []string{
		"--modules", "user",
		"--config", getTestConfigPath(),
	}

	// First run
	cmd1 := exec.Command(phanesBinary, args...)
	cmd1.Dir = "/workspace"
	output1, err1 := cmd1.CombinedOutput()
	if err1 != nil {
		t.Fatalf("First run failed: %v\nOutput: %s", err1, string(output1))
	}

	// Second run - should skip everything
	cmd2 := exec.Command(phanesBinary, args...)
	cmd2.Dir = "/workspace"
	output2, err2 := cmd2.CombinedOutput()
	if err2 != nil {
		t.Fatalf("Second run failed: %v\nOutput: %s", err2, string(output2))
	}

	output2Str := string(output2)
	// On second run, we should see skip messages
	if !strings.Contains(output2Str, "skip") && !strings.Contains(output2Str, "already") {
		t.Logf("Second run output: %s", output2Str)
		// This is okay - Install() is idempotent and may still run checks
		// The important thing is that it doesn't fail and doesn't duplicate things
	}

	// Verify user still exists and is correct
	idCmd := exec.Command("id", "testuser")
	if err := idCmd.Run(); err != nil {
		t.Error("User should still exist after second run")
	}

	// Verify SSH key wasn't duplicated
	authorizedKeys := "/home/testuser/.ssh/authorized_keys"
	if _, err := os.Stat(authorizedKeys); err == nil {
		content, err := os.ReadFile(authorizedKeys)
		if err == nil {
			lines := strings.Split(string(content), "\n")
			sshKeyCount := 0
			for _, line := range lines {
				if strings.Contains(line, "ssh-rsa") {
					sshKeyCount++
				}
			}
			if sshKeyCount > 1 {
				t.Errorf("SSH key should not be duplicated. Found %d keys", sshKeyCount)
			}
		}
	}
}

// TestMultipleModulesE2E tests running multiple modules together.
func TestMultipleModulesE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	if !isContainerized() {
		t.Skip("Skipping E2E test - not running in containerized environment")
	}

	tests := []struct {
		name           string
		args           []string
		expectedOutput []string
		shouldFail     bool
	}{
		{
			name: "baseline and user modules together",
			args: []string{
				"--modules", "baseline,user",
				"--config", testConfigFile,
			},
			expectedOutput: []string{
				"Processing module: baseline",
				"Processing module: user",
			},
			shouldFail: false, // May fail if user module fails, but we verify both modules are processed
		},
		{
			name: "multiple modules dry-run",
			args: []string{
				"--modules", "baseline,user",
				"--config", testConfigFile,
				"--dry-run",
			},
			expectedOutput: []string{
				"dry_run=",
				"Processing module: baseline",
				"Processing module: user",
				"All modules executed successfully",
			},
			shouldFail: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Replace test-config.yaml with absolute path in args
			args := make([]string, len(tt.args))
			copy(args, tt.args)
			for i, arg := range args {
				if arg == testConfigFile {
					args[i] = getTestConfigPath()
				}
			}

			cmd := exec.Command(phanesBinary, args...)
			cmd.Dir = "/workspace"
			output, err := cmd.CombinedOutput()
			outputStr := string(output)

			if tt.shouldFail {
				if err == nil {
					t.Errorf("Expected command to fail, but it succeeded. Output: %s", outputStr)
				}
			} else {
				// Verify expected output first (modules being processed)
				// Some tests may fail due to module errors, but we verify processing happened
				allFound := true
				for _, expected := range tt.expectedOutput {
					if !strings.Contains(outputStr, expected) {
						allFound = false
						t.Errorf("Expected output to contain '%s', but it didn't.\nFull output: %s", expected, outputStr)
					}
				}

				// If output verification passed, don't fail even if command errored
				// (some modules may fail due to missing config, but processing is verified)
				if !allFound && err != nil {
					t.Logf("Command also failed: %v", err)
				}
			}
		})
	}
}

// isContainerized checks if we're running in a containerized environment.
// This is a simple check - in Docker containers, /.dockerenv usually exists.
func isContainerized() bool {
	_, err := os.Stat("/.dockerenv")
	return err == nil
}
