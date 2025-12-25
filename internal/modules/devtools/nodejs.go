package devtools

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/stwalsh4118/phanes/internal/config"
	"github.com/stwalsh4118/phanes/internal/exec"
	"github.com/stwalsh4118/phanes/internal/log"
)

const (
	nvmVersion   = "v0.40.0"
	nvmInstallURL = "https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.0/install.sh"
	nvmDirName   = ".nvm"
)

// nvmInitScript returns the nvm initialization script that should be added to shell profiles.
func nvmInitScript() string {
	return `export NVM_DIR="$HOME/.nvm"
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"
[ -s "$NVM_DIR/bash_completion" ] && \. "$NVM_DIR/bash_completion"
`
}

// nvmInstalled checks if nvm is installed for a specific user.
// Checks if ~/.nvm directory and nvm.sh script exist.
func nvmInstalled(username string) (bool, error) {
	homeDir := filepath.Join("/home", username)
	nvmDir := filepath.Join(homeDir, nvmDirName)
	nvmScript := filepath.Join(nvmDir, "nvm.sh")

	if !exec.FileExists(nvmDir) {
		return false, nil
	}

	if !exec.FileExists(nvmScript) {
		return false, nil
	}

	return true, nil
}

// nodeInstalled checks if a Node.js version is installed via nvm for a user.
// Uses nvm which to check if the version is installed.
func nodeInstalled(username, version string) (bool, error) {
	// First check if nvm is installed
	nvmOk, err := nvmInstalled(username)
	if err != nil {
		return false, fmt.Errorf("failed to check nvm installation: %w", err)
	}
	if !nvmOk {
		return false, nil
	}

	// Check if Node.js version is installed via nvm
	// Use su to run as the user with nvm sourced
	cmd := fmt.Sprintf("source ~/.nvm/nvm.sh && nvm which %s >/dev/null 2>&1", version)
	output, err := exec.RunWithOutput("su", "-", username, "-c", cmd)
	if err != nil {
		// If command fails, version is not installed
		return false, nil
	}

	// If output contains a path, version is installed
	if strings.TrimSpace(output) != "" {
		return true, nil
	}

	return false, nil
}

// shellProfileHasNvm checks if a shell profile already contains nvm initialization.
func shellProfileHasNvm(profilePath string) (bool, error) {
	if !exec.FileExists(profilePath) {
		return false, nil
	}

	content, err := os.ReadFile(profilePath)
	if err != nil {
		return false, fmt.Errorf("failed to read shell profile: %w", err)
	}

	// Check if nvm initialization is already present
	// Look for NVM_DIR export or nvm.sh sourcing
	contentStr := string(content)
	if strings.Contains(contentStr, "NVM_DIR") && strings.Contains(contentStr, "nvm.sh") {
		return true, nil
	}

	return false, nil
}

// configureShellProfile appends nvm initialization to a shell profile if not already present.
func configureShellProfile(profilePath string, userUID, userGID int) error {
	hasNvm, err := shellProfileHasNvm(profilePath)
	if err != nil {
		return fmt.Errorf("failed to check shell profile: %w", err)
	}

	if hasNvm {
		log.Skip("Shell profile %s already has nvm initialization", profilePath)
		return nil
	}

	// Read existing content
	var content []byte
	if exec.FileExists(profilePath) {
		existingContent, err := os.ReadFile(profilePath)
		if err != nil {
			return fmt.Errorf("failed to read shell profile: %w", err)
		}
		content = existingContent
		// Add newline if file doesn't end with one
		if len(content) > 0 && content[len(content)-1] != '\n' {
			content = append(content, '\n')
		}
	}

	// Append nvm initialization
	content = append(content, []byte("\n# nvm initialization\n")...)
	content = append(content, []byte(nvmInitScript())...)

	// Write file
	if err := exec.WriteFile(profilePath, content, 0644); err != nil {
		return fmt.Errorf("failed to write shell profile: %w", err)
	}

	// Set ownership to the user
	if err := os.Chown(profilePath, userUID, userGID); err != nil {
		return fmt.Errorf("failed to set shell profile ownership: %w", err)
	}

	log.Success("Configured shell profile: %s", profilePath)
	return nil
}

// installNodeJS installs nvm and Node.js for the configured user.
func installNodeJS(cfg *config.Config) error {
	dryRun := log.IsDryRun()

	// Validate username is set
	if cfg.User.Username == "" {
		log.Warn("Username is not set, skipping Node.js installation")
		return nil
	}

	username := cfg.User.Username
	nodeVersion := cfg.DevTools.NodeVersion
	if nodeVersion == "" {
		nodeVersion = "22" // Default to Node.js 22
	}

	// Get user info for file ownership
	var userUID, userGID int
	if !dryRun {
		userInfo, err := user.Lookup(username)
		if err != nil {
			return fmt.Errorf("failed to look up user %s: %w", username, err)
		}
		userUID, err = strconv.Atoi(userInfo.Uid)
		if err != nil {
			return fmt.Errorf("failed to parse user UID: %w", err)
		}
		userGID, err = strconv.Atoi(userInfo.Gid)
		if err != nil {
			return fmt.Errorf("failed to parse user GID: %w", err)
		}
	}

	homeDir := filepath.Join("/home", username)
	bashrcPath := filepath.Join(homeDir, ".bashrc")
	zshrcPath := filepath.Join(homeDir, ".zshrc")

	// Check if nvm is already installed
	nvmOk, err := nvmInstalled(username)
	if err != nil {
		return fmt.Errorf("failed to check nvm installation: %w", err)
	}

	if !nvmOk {
		if dryRun {
			log.Info("Would install nvm for user: %s", username)
		} else {
			log.Info("Installing nvm for user: %s", username)
			// Install nvm using the official install script
			// Run as the user to ensure it's installed in their home directory
			installCmd := fmt.Sprintf("curl -o- %s | bash", nvmInstallURL)
			if err := exec.Run("su", "-", username, "-c", installCmd); err != nil {
				return fmt.Errorf("failed to install nvm: %w", err)
			}

			// Verify nvm was installed
			nvmOk, err = nvmInstalled(username)
			if err != nil {
				return fmt.Errorf("failed to verify nvm installation: %w", err)
			}
			if !nvmOk {
				return fmt.Errorf("nvm installation verification failed: nvm directory not found")
			}

			log.Success("nvm installed successfully")
		}
	} else {
		log.Skip("nvm is already installed for user: %s", username)
	}

	// Configure shell profiles
	if !dryRun {
		// Configure .bashrc
		if exec.FileExists(bashrcPath) {
			if err := configureShellProfile(bashrcPath, userUID, userGID); err != nil {
				return fmt.Errorf("failed to configure .bashrc: %w", err)
			}
		} else {
			// Create .bashrc if it doesn't exist
			if err := configureShellProfile(bashrcPath, userUID, userGID); err != nil {
				return fmt.Errorf("failed to create .bashrc: %w", err)
			}
		}

		// Configure .zshrc
		if exec.FileExists(zshrcPath) {
			if err := configureShellProfile(zshrcPath, userUID, userGID); err != nil {
				return fmt.Errorf("failed to configure .zshrc: %w", err)
			}
		} else {
			// Create .zshrc if it doesn't exist
			if err := configureShellProfile(zshrcPath, userUID, userGID); err != nil {
				return fmt.Errorf("failed to create .zshrc: %w", err)
			}
		}
	} else {
		log.Info("Would configure shell profiles (.bashrc and .zshrc) for nvm")
	}

	// Check if Node.js version is already installed
	nodeOk, err := nodeInstalled(username, nodeVersion)
	if err != nil {
		return fmt.Errorf("failed to check Node.js installation: %w", err)
	}

	if !nodeOk {
		if dryRun {
			log.Info("Would install Node.js version %s for user: %s", nodeVersion, username)
		} else {
			log.Info("Installing Node.js version %s for user: %s", nodeVersion, username)

			// Install Node.js via nvm
			installCmd := fmt.Sprintf("source ~/.nvm/nvm.sh && nvm install %s", nodeVersion)
			if err := exec.Run("su", "-", username, "-c", installCmd); err != nil {
				return fmt.Errorf("failed to install Node.js: %w", err)
			}

			// Set as default version
			aliasCmd := fmt.Sprintf("source ~/.nvm/nvm.sh && nvm alias default %s", nodeVersion)
			if err := exec.Run("su", "-", username, "-c", aliasCmd); err != nil {
				return fmt.Errorf("failed to set Node.js default version: %w", err)
			}

			// Verify installation
			verifyCmd := "source ~/.nvm/nvm.sh && node --version && npm --version"
			output, err := exec.RunWithOutput("su", "-", username, "-c", verifyCmd)
			if err != nil {
				return fmt.Errorf("failed to verify Node.js installation: %w", err)
			}
			if strings.TrimSpace(output) == "" {
				return fmt.Errorf("Node.js installation verification failed: no version output")
			}

			log.Success("Node.js version %s installed successfully", nodeVersion)
		}
	} else {
		log.Skip("Node.js version %s is already installed for user: %s", nodeVersion, username)
	}

	if !dryRun {
		log.Success("Node.js installation completed successfully")
	}

	return nil
}

