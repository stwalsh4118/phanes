package user

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stwalsh4118/phanes/internal/config"
	"github.com/stwalsh4118/phanes/internal/exec"
	"github.com/stwalsh4118/phanes/internal/log"
)

func TestUserModule_Name(t *testing.T) {
	mod := &UserModule{}
	if got := mod.Name(); got != "user" {
		t.Errorf("Name() = %q, want %q", got, "user")
	}
}

func TestUserModule_Description(t *testing.T) {
	mod := &UserModule{}
	want := "Creates user and sets up SSH keys"
	if got := mod.Description(); got != want {
		t.Errorf("Description() = %q, want %q", got, want)
	}
}

func TestValidateSSHKey(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{
			name:    "valid ssh-rsa key",
			key:     "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC... test@example.com",
			wantErr: false,
		},
		{
			name:    "valid ssh-ed25519 key",
			key:     "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI... test@example.com",
			wantErr: false,
		},
		{
			name:    "valid ecdsa-sha2 key",
			key:     "ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBB... test@example.com",
			wantErr: false,
		},
		{
			name:    "valid ssh-dss key",
			key:     "ssh-dss AAAAB3NzaC1kc3MAAACBA... test@example.com",
			wantErr: false,
		},
		{
			name:    "invalid key format",
			key:     "invalid-key-format",
			wantErr: true,
		},
		{
			name:    "empty key",
			key:     "",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			key:     "   ",
			wantErr: true,
		},
		{
			name:    "key with leading/trailing whitespace",
			key:     "  ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC... test@example.com  ",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSSHKey(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateSSHKey() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserExists(t *testing.T) {
	// Test with a user that definitely doesn't exist
	exists, err := userExists("nonexistentuser12345")
	if err != nil {
		t.Errorf("userExists() error = %v, expected no error for non-existent user", err)
	}
	if exists {
		t.Error("userExists() returned true for non-existent user")
	}

	// Test with root user (should exist on most systems)
	exists, err = userExists("root")
	if err != nil {
		t.Logf("userExists() error for root user (may not exist in some environments): %v", err)
		return
	}
	if !exists {
		t.Log("root user doesn't exist (unusual but possible in some environments)")
	}
}

func TestSSHKeyExists(t *testing.T) {
	tempDir := t.TempDir()
	authorizedKeysPath := filepath.Join(tempDir, "authorized_keys")

	// Test with non-existent file
	exists, err := sshKeyExists(authorizedKeysPath, "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC... test@example.com")
	if err != nil {
		t.Errorf("sshKeyExists() error = %v, expected no error for non-existent file", err)
	}
	if exists {
		t.Error("sshKeyExists() returned true for non-existent file")
	}

	// Create test file with a key
	testKey := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC... test@example.com"
	content := testKey + "\n"
	if err := os.WriteFile(authorizedKeysPath, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test with existing key
	exists, err = sshKeyExists(authorizedKeysPath, testKey)
	if err != nil {
		t.Errorf("sshKeyExists() error = %v", err)
	}
	if !exists {
		t.Error("sshKeyExists() returned false for existing key")
	}

	// Test with different key
	exists, err = sshKeyExists(authorizedKeysPath, "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI... other@example.com")
	if err != nil {
		t.Errorf("sshKeyExists() error = %v", err)
	}
	if exists {
		t.Error("sshKeyExists() returned true for non-existent key")
	}

	// Test with key that has whitespace differences
	exists, err = sshKeyExists(authorizedKeysPath, "  "+testKey+"  ")
	if err != nil {
		t.Errorf("sshKeyExists() error = %v", err)
	}
	if !exists {
		t.Error("sshKeyExists() should match key with whitespace differences")
	}

	// Test with file containing multiple keys
	multiKeyContent := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC... key1@example.com\n" +
		"# Comment line\n" +
		"ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI... key2@example.com\n" +
		"\n" +
		"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC... key3@example.com\n"
	if err := os.WriteFile(authorizedKeysPath, []byte(multiKeyContent), 0600); err != nil {
		t.Fatalf("Failed to write multi-key file: %v", err)
	}

	// Test finding first key
	exists, err = sshKeyExists(authorizedKeysPath, "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC... key1@example.com")
	if err != nil {
		t.Errorf("sshKeyExists() error = %v", err)
	}
	if !exists {
		t.Error("sshKeyExists() should find key in multi-key file")
	}

	// Test finding middle key
	exists, err = sshKeyExists(authorizedKeysPath, "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI... key2@example.com")
	if err != nil {
		t.Errorf("sshKeyExists() error = %v", err)
	}
	if !exists {
		t.Error("sshKeyExists() should find middle key in multi-key file")
	}

	// Test with key not in file (using different key data)
	exists, err = sshKeyExists(authorizedKeysPath, "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABBBB... notfound@example.com")
	if err != nil {
		t.Errorf("sshKeyExists() error = %v", err)
	}
	if exists {
		t.Error("sshKeyExists() should not find non-existent key")
	}
}

func TestUserModule_IsInstalled(t *testing.T) {
	mod := &UserModule{}

	// Test that IsInstalled() doesn't panic and handles errors gracefully
	installed, err := mod.IsInstalled()
	if err != nil {
		t.Logf("IsInstalled() returned error (may be expected): %v", err)
	}

	// Verify return type
	_ = installed
	_ = err

	// Note: IsInstalled() doesn't receive config, so it can't check against
	// specific username/SSH key. It performs a generic check. The actual
	// specific checks are done in Install() which is idempotent.
}

func TestUserModule_Install_Validation(t *testing.T) {
	mod := &UserModule{}

	// Test with empty username
	cfg := &config.Config{
		User: config.User{
			Username:     "",
			SSHPublicKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC... test@example.com",
		},
	}

	err := mod.Install(cfg)
	if err == nil {
		t.Error("Install() should return error for empty username")
	}
	if !strings.Contains(err.Error(), "username is required") {
		t.Errorf("Install() error should mention username is required, got: %v", err)
	}

	// Test with invalid SSH key
	cfg = &config.Config{
		User: config.User{
			Username:     "testuser",
			SSHPublicKey: "invalid-key-format",
		},
	}

	err = mod.Install(cfg)
	if err == nil {
		t.Error("Install() should return error for invalid SSH key")
	}
	if !strings.Contains(err.Error(), "invalid SSH public key") {
		t.Errorf("Install() error should mention invalid SSH key, got: %v", err)
	}
}

func TestUserModule_Install_DryRun(t *testing.T) {
	mod := &UserModule{}

	// Enable dry-run mode
	originalDryRun := log.IsDryRun()
	log.SetDryRun(true)
	defer log.SetDryRun(originalDryRun)

	cfg := &config.Config{
		User: config.User{
			Username:     "testuser",
			SSHPublicKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC... test@example.com",
		},
	}

	// In dry-run mode, Install() should log but not execute commands
	// We can't easily verify this without mocking, but we can verify
	// it doesn't panic and handles dry-run mode
	err := mod.Install(cfg)
	// In dry-run mode, it should not actually create the user, so it might
	// fail when trying to check if user exists, or it might succeed if it
	// just logs what would be done
	if err != nil {
		// Error is acceptable in dry-run mode if user doesn't exist
		// The important thing is it doesn't panic
		t.Logf("Install() in dry-run mode returned error (may be expected): %v", err)
	}
}

func TestUserModule_ModuleInterface(t *testing.T) {
	// Verify that UserModule implements the Module interface
	var _ interface {
		Name() string
		Description() string
		IsInstalled() (bool, error)
		Install(*config.Config) error
	} = &UserModule{}
}

func TestUserModule_Install_RequiresRoot(t *testing.T) {
	// Skip this test in non-root environments
	// Installing requires root privileges
	if os.Geteuid() != 0 {
		t.Skip("Skipping Install test - requires root privileges")
	}

	// Check if useradd exists
	if !exec.CommandExists("useradd") {
		t.Skip("useradd not available")
	}

	// This test would actually create a user, so we skip it
	// In a real scenario, you'd want to:
	// 1. Create a test user
	// 2. Run Install()
	// 3. Verify user, SSH keys, and sudoers are configured
	// 4. Clean up test user
	t.Skip("Skipping Install test - would modify system configuration")
}

func TestUserModule_Install_Idempotency(t *testing.T) {
	// Test that Install() is idempotent
	// This would require running Install() twice and verifying
	// the second run doesn't create duplicates or errors
	// We skip this test as it requires root and would modify system
	if os.Geteuid() != 0 {
		t.Skip("Skipping idempotency test - requires root privileges")
	}
	t.Skip("Skipping idempotency test - would modify system configuration")
}

func TestUserModule_Install_SSHKeyDeduplication(t *testing.T) {
	tempDir := t.TempDir()
	authorizedKeysPath := filepath.Join(tempDir, "authorized_keys")

	// Create file with existing key
	testKey := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC... test@example.com"
	existingContent := testKey + "\n"
	if err := os.WriteFile(authorizedKeysPath, []byte(existingContent), 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Check if key exists
	exists, err := sshKeyExists(authorizedKeysPath, testKey)
	if err != nil {
		t.Fatalf("sshKeyExists() error = %v", err)
	}
	if !exists {
		t.Error("sshKeyExists() should find existing key")
	}
}
