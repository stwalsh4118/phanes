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
	uvInstallURL = "https://astral.sh/uv/install.sh"
	uvBinDir     = ".local/bin"
	uvBinName    = "uv"
)

const (
	packagePython3     = "python3"
	packagePython3Venv = "python3-venv"
	packagePython3Pip = "python3-pip"
)

// uvPathScript returns the PATH export script that should be added to shell profiles for uv.
func uvPathScript() string {
	return `export PATH="$HOME/.local/bin:$PATH"
`
}

// pythonInstalled checks if Python 3 is installed.
// Checks if python3 command exists in PATH.
func pythonInstalled() (bool, error) {
	return exec.CommandExists("python3"), nil
}

// uvInstalled checks if uv is installed for a specific user.
// Checks if ~/.local/bin/uv exists.
func uvInstalled(username string) (bool, error) {
	homeDir := filepath.Join("/home", username)
	uvBinPath := filepath.Join(homeDir, uvBinDir, uvBinName)

	if !exec.FileExists(uvBinPath) {
		return false, nil
	}

	return true, nil
}

// shellProfileHasUvPath checks if a shell profile already contains uv PATH configuration.
func shellProfileHasUvPath(profilePath string) (bool, error) {
	if !exec.FileExists(profilePath) {
		return false, nil
	}

	content, err := os.ReadFile(profilePath)
	if err != nil {
		return false, fmt.Errorf("failed to read shell profile: %w", err)
	}

	// Check if uv PATH is already present
	// Look for .local/bin in PATH export
	contentStr := string(content)
	if strings.Contains(contentStr, ".local/bin") && strings.Contains(contentStr, "PATH") {
		return true, nil
	}

	return false, nil
}

// configureShellProfileForUv appends uv PATH to a shell profile if not already present.
func configureShellProfileForUv(profilePath string, userUID, userGID int) error {
	hasUvPath, err := shellProfileHasUvPath(profilePath)
	if err != nil {
		return fmt.Errorf("failed to check shell profile: %w", err)
	}

	if hasUvPath {
		log.Skip("Shell profile %s already has uv PATH configuration", profilePath)
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

	// Append uv PATH configuration
	content = append(content, []byte("\n# uv PATH configuration\n")...)
	content = append(content, []byte(uvPathScript())...)

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

// installPython installs Python 3 and optionally uv for the configured user.
func installPython(cfg *config.Config) error {
	dryRun := log.IsDryRun()

	// Check if Python 3 is already installed
	pythonOk, err := pythonInstalled()
	if err != nil {
		return fmt.Errorf("failed to check Python installation: %w", err)
	}

	if !pythonOk {
		if dryRun {
			log.Info("Would install Python 3")
		} else {
			log.Info("Installing Python 3")

			// Update apt package list
			log.Info("Updating apt package list")
			if err := exec.Run("apt-get", "update"); err != nil {
				return fmt.Errorf("failed to update apt: %w", err)
			}

			// Install Python 3 and related packages
			log.Info("Installing packages: python3, python3-venv, python3-pip")
			if err := exec.Run("apt-get", "install", "-y", packagePython3, packagePython3Venv, packagePython3Pip); err != nil {
				return fmt.Errorf("failed to install Python 3: %w", err)
			}

			// Verify installation
			output, err := exec.RunWithOutput("python3", "--version")
			if err != nil {
				return fmt.Errorf("failed to verify Python 3 installation: %w", err)
			}
			if strings.TrimSpace(output) == "" {
				return fmt.Errorf("Python 3 installation verification failed: no version output")
			}

			log.Success("Python 3 installed successfully: %s", strings.TrimSpace(output))
		}
	} else {
		log.Skip("Python 3 is already installed")
	}

	// Install uv if enabled
	if cfg.DevTools.InstallUv {
		// Validate username is set
		if cfg.User.Username == "" {
			log.Warn("Username is not set, skipping uv installation")
			return nil
		}

		username := cfg.User.Username

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

		// Check if uv is already installed
		uvOk, err := uvInstalled(username)
		if err != nil {
			return fmt.Errorf("failed to check uv installation: %w", err)
		}

		if !uvOk {
			if dryRun {
				log.Info("Would install uv for user: %s", username)
			} else {
				log.Info("Installing uv for user: %s", username)
				// Install uv using the official install script
				// Run as the user to ensure it's installed in their home directory
				installCmd := fmt.Sprintf("curl -LsSf %s | sh", uvInstallURL)
				if err := exec.Run("su", "-", username, "-c", installCmd); err != nil {
					return fmt.Errorf("failed to install uv: %w", err)
				}

				// Verify uv was installed
				uvOk, err = uvInstalled(username)
				if err != nil {
					return fmt.Errorf("failed to verify uv installation: %w", err)
				}
				if !uvOk {
					return fmt.Errorf("uv installation verification failed: uv binary not found")
				}

				log.Success("uv installed successfully")
			}
		} else {
			log.Skip("uv is already installed for user: %s", username)
		}

		// Configure shell profiles
		if !dryRun {
			// Configure .bashrc
			if exec.FileExists(bashrcPath) {
				if err := configureShellProfileForUv(bashrcPath, userUID, userGID); err != nil {
					return fmt.Errorf("failed to configure .bashrc: %w", err)
				}
			} else {
				// Create .bashrc if it doesn't exist
				if err := configureShellProfileForUv(bashrcPath, userUID, userGID); err != nil {
					return fmt.Errorf("failed to create .bashrc: %w", err)
				}
			}

			// Configure .zshrc
			if exec.FileExists(zshrcPath) {
				if err := configureShellProfileForUv(zshrcPath, userUID, userGID); err != nil {
					return fmt.Errorf("failed to configure .zshrc: %w", err)
				}
			} else {
				// Create .zshrc if it doesn't exist
				if err := configureShellProfileForUv(zshrcPath, userUID, userGID); err != nil {
					return fmt.Errorf("failed to create .zshrc: %w", err)
				}
			}

			// Verify uv installation
			verifyCmd := "~/.local/bin/uv --version"
			output, err := exec.RunWithOutput("su", "-", username, "-c", verifyCmd)
			if err != nil {
				return fmt.Errorf("failed to verify uv installation: %w", err)
			}
			if strings.TrimSpace(output) == "" {
				return fmt.Errorf("uv installation verification failed: no version output")
			}

			log.Success("uv version: %s", strings.TrimSpace(output))
		} else {
			log.Info("Would configure shell profiles (.bashrc and .zshrc) for uv")
		}
	}

	if !dryRun {
		log.Success("Python installation completed successfully")
	}

	return nil
}

