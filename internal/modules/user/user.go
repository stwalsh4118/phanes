package user

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/stwalsh4118/phanes/internal/config"
	"github.com/stwalsh4118/phanes/internal/exec"
	"github.com/stwalsh4118/phanes/internal/log"
	"github.com/stwalsh4118/phanes/internal/module"
)

const (
	sshDirPerm     = 0700
	authorizedKeysPerm = 0600
	sudoersPerm   = 0440
)

// UserModule implements the Module interface for user creation and SSH key setup.
// It creates a non-root user, sets up SSH key access, and configures passwordless sudo.
type UserModule struct{}

// Name returns the unique name identifier for this module.
func (m *UserModule) Name() string {
	return "user"
}

// Description returns a human-readable description of what this module does.
func (m *UserModule) Description() string {
	return "Creates user and sets up SSH keys"
}

// validateSSHKey checks if the SSH public key has a valid format.
// Valid formats include: ssh-rsa, ssh-ed25519, ecdsa-sha2-*, ssh-dss
func validateSSHKey(key string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return fmt.Errorf("SSH public key is empty")
	}

	validPrefixes := []string{
		"ssh-rsa",
		"ssh-ed25519",
		"ecdsa-sha2-",
		"ssh-dss",
	}

	for _, prefix := range validPrefixes {
		if strings.HasPrefix(key, prefix) {
			return nil
		}
	}

	return fmt.Errorf("invalid SSH key format: must start with ssh-rsa, ssh-ed25519, ecdsa-sha2-, or ssh-dss")
}

// userExists checks if a user exists on the system.
func userExists(username string) (bool, error) {
	_, err := user.Lookup(username)
	if err != nil {
		if _, ok := err.(user.UnknownUserError); ok {
			return false, nil
		}
		return false, fmt.Errorf("failed to check if user exists: %w", err)
	}
	return true, nil
}

// sshKeyExists checks if the SSH key already exists in the authorized_keys file.
func sshKeyExists(authorizedKeysPath string, key string) (bool, error) {
	if !exec.FileExists(authorizedKeysPath) {
		return false, nil
	}

	content, err := os.ReadFile(authorizedKeysPath)
	if err != nil {
		return false, fmt.Errorf("failed to read authorized_keys file: %w", err)
	}

	key = strings.TrimSpace(key)
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Check if the key matches exactly (handles whitespace differences)
		if line == key {
			return true, nil
		}
		// Also check if it's the same key by comparing the key type and full key data
		// The key data (second field) is what makes the key unique
		// This handles cases where whitespace differs or comments differ
		keyParts := strings.Fields(key)
		lineParts := strings.Fields(line)
		if len(keyParts) >= 2 && len(lineParts) >= 2 {
			// Match on key type and full key data (second field)
			// The key data is unique per key, so if both match, it's the same key
			if keyParts[0] == lineParts[0] && keyParts[1] == lineParts[1] {
				return true, nil
			}
		}
	}

	return false, nil
}

// IsInstalled checks if the user module is already installed.
// Note: Since IsInstalled() doesn't receive config, it can't check against
// specific username or SSH key values. It performs a generic check to see
// if the system appears to have been set up for user management.
// The Install() method performs the specific checks with config and is fully idempotent.
func (m *UserModule) IsInstalled() (bool, error) {
	// Check if sudoers.d directory exists and has files
	// This is a generic check that suggests the module has been run
	sudoersDir := "/etc/sudoers.d"
	if !exec.FileExists(sudoersDir) {
		return false, nil
	}

	// Try to list files in sudoers.d to see if any exist
	// We can't check for a specific username without config, so we just
	// check if the directory exists and appears to be set up
	entries, err := os.ReadDir(sudoersDir)
	if err != nil {
		return false, fmt.Errorf("failed to read sudoers.d directory: %w", err)
	}

	// If sudoers.d exists and has files, assume the module might be installed
	// Install() will do the specific checks with config
	if len(entries) > 0 {
		// Return false to ensure Install() is called for specific checks
		// Install() is idempotent and will skip if already configured
		return false, nil
	}

	return false, nil
}

// Install creates the user, sets up SSH keys, and configures passwordless sudo.
// This method is idempotent - it checks if each step is already done before doing it.
func (m *UserModule) Install(cfg *config.Config) error {
	// Validate required fields
	if cfg.User.Username == "" {
		return fmt.Errorf("username is required")
	}

	if err := validateSSHKey(cfg.User.SSHPublicKey); err != nil {
		return fmt.Errorf("invalid SSH public key: %w", err)
	}

	dryRun := log.IsDryRun()
	username := cfg.User.Username
	sshKey := strings.TrimSpace(cfg.User.SSHPublicKey)

	// Get user's home directory path
	homeDir := filepath.Join("/home", username)
	sshDir := filepath.Join(homeDir, ".ssh")
	authorizedKeysPath := filepath.Join(sshDir, "authorized_keys")
	sudoersPath := filepath.Join("/etc/sudoers.d", username)

	// Check if user exists
	userExists, err := userExists(username)
	if err != nil {
		return fmt.Errorf("failed to check if user exists: %w", err)
	}

	if !userExists {
		if dryRun {
			log.Info("Would create user: %s", username)
		} else {
			log.Info("Creating user: %s", username)
			if err := exec.Run("useradd", "-m", "-s", "/bin/bash", username); err != nil {
				// Check if error is because user already exists (race condition)
				// useradd returns exit code 9 if user already exists
				if strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "exit status 9") {
					log.Info("User %s already exists, continuing", username)
				} else {
					return fmt.Errorf("failed to create user: %w", err)
				}
			} else {
				log.Success("Created user: %s", username)
			}
		}
	} else {
		log.Skip("User %s already exists", username)
	}

	// Create .ssh directory
	if dryRun {
		log.Info("Would create SSH directory: %s", sshDir)
	} else {
		if !exec.FileExists(sshDir) {
			log.Info("Creating SSH directory: %s", sshDir)
			if err := os.MkdirAll(sshDir, sshDirPerm); err != nil {
				return fmt.Errorf("failed to create SSH directory: %w", err)
			}
			if err := os.Chmod(sshDir, sshDirPerm); err != nil {
				return fmt.Errorf("failed to set SSH directory permissions: %w", err)
			}
			log.Success("Created SSH directory: %s", sshDir)
		} else {
			log.Skip("SSH directory already exists: %s", sshDir)
		}
	}

	// Add SSH key to authorized_keys
	keyExists, err := sshKeyExists(authorizedKeysPath, sshKey)
	if err != nil && !dryRun {
		return fmt.Errorf("failed to check if SSH key exists: %w", err)
	}

	if !keyExists {
		if dryRun {
			log.Info("Would add SSH key to authorized_keys: %s", authorizedKeysPath)
		} else {
			log.Info("Adding SSH key to authorized_keys")
			var content []byte
			if exec.FileExists(authorizedKeysPath) {
				existingContent, err := os.ReadFile(authorizedKeysPath)
				if err != nil {
					return fmt.Errorf("failed to read authorized_keys file: %w", err)
				}
				content = existingContent
				// Add newline if file doesn't end with one
				if len(content) > 0 && content[len(content)-1] != '\n' {
					content = append(content, '\n')
				}
			}
			content = append(content, []byte(sshKey)...)
			content = append(content, '\n')

			if err := exec.WriteFile(authorizedKeysPath, content, authorizedKeysPerm); err != nil {
				return fmt.Errorf("failed to write authorized_keys file: %w", err)
			}
			if err := os.Chmod(authorizedKeysPath, authorizedKeysPerm); err != nil {
				return fmt.Errorf("failed to set authorized_keys permissions: %w", err)
			}
			log.Success("Added SSH key to authorized_keys")
		}
	} else {
		log.Skip("SSH key already exists in authorized_keys")
	}

	// Configure passwordless sudo
	sudoersContent := fmt.Sprintf("%s ALL=(ALL) NOPASSWD:ALL\n", username)
	if dryRun {
		log.Info("Would create sudoers file: %s", sudoersPath)
		log.Info("Would configure passwordless sudo for user: %s", username)
	} else {
		// Check if sudoers file already exists and is correct
		needsUpdate := true
		if exec.FileExists(sudoersPath) {
			existingContent, err := os.ReadFile(sudoersPath)
			if err == nil {
				if strings.TrimSpace(string(existingContent)) == strings.TrimSpace(sudoersContent) {
					needsUpdate = false
				}
			}
		}

		if needsUpdate {
			log.Info("Configuring passwordless sudo for user: %s", username)
			if err := exec.WriteFile(sudoersPath, []byte(sudoersContent), sudoersPerm); err != nil {
				return fmt.Errorf("failed to write sudoers file: %w", err)
			}
			if err := os.Chmod(sudoersPath, sudoersPerm); err != nil {
				return fmt.Errorf("failed to set sudoers file permissions: %w", err)
			}

			// Validate sudoers file
			log.Info("Validating sudoers file")
			if err := exec.Run("visudo", "-c", "-f", sudoersPath); err != nil {
				// If validation fails, remove the file we just created
				os.Remove(sudoersPath)
				return fmt.Errorf("sudoers file validation failed: %w", err)
			}
			log.Success("Configured passwordless sudo for user: %s", username)
		} else {
			log.Skip("Sudoers file already configured correctly")
		}
	}

	if !dryRun {
		log.Success("User module installation completed successfully")
	}

	return nil
}

// Ensure UserModule implements the Module interface
var _ module.Module = (*UserModule)(nil)

