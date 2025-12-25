package docker

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/stwalsh4118/phanes/internal/config"
	"github.com/stwalsh4118/phanes/internal/exec"
	"github.com/stwalsh4118/phanes/internal/log"
	"github.com/stwalsh4118/phanes/internal/module"
)

const (
	dockerGPGKeyURL      = "https://download.docker.com/linux/ubuntu/gpg"
	dockerGPGKeyringPath = "/usr/share/keyrings/docker-archive-keyring.gpg"
	dockerAptSourcesPath = "/etc/apt/sources.list.d/docker.list"
	dockerRepoURL        = "https://download.docker.com/linux/ubuntu"
)

// DockerModule implements the Module interface for Docker CE and Docker Compose installation.
type DockerModule struct{}

// Name returns the unique name identifier for this module.
func (m *DockerModule) Name() string {
	return "docker"
}

// Description returns a human-readable description of what this module does.
func (m *DockerModule) Description() string {
	return "Installs Docker CE and Docker Compose"
}

// dockerInstalled checks if Docker is installed by running docker --version.
func dockerInstalled() (bool, error) {
	err := exec.Run("docker", "--version")
	if err != nil {
		return false, nil
	}
	return true, nil
}

// dockerServiceRunning checks if Docker service is running.
func dockerServiceRunning() (bool, error) {
	output, err := exec.RunWithOutput("systemctl", "is-active", "docker")
	if err != nil {
		return false, nil
	}
	return strings.TrimSpace(output) == "active", nil
}

// dockerComposeInstalled checks if Docker Compose v2 is installed.
func dockerComposeInstalled() (bool, error) {
	err := exec.Run("docker", "compose", "version")
	if err != nil {
		return false, nil
	}
	return true, nil
}

// userExists checks if a user exists on the system.
func userExists(username string) bool {
	if username == "" {
		return false
	}
	_, err := exec.RunWithOutput("id", username)
	return err == nil
}

// userInDockerGroup checks if a user is in the docker group.
// Returns (false, nil) if user doesn't exist (caller should check userExists first).
func userInDockerGroup(username string) (bool, error) {
	if username == "" {
		return false, nil
	}

	output, err := exec.RunWithOutput("id", "-nG", username)
	if err != nil {
		// User doesn't exist - return false without error (caller should use userExists to check)
		return false, nil
	}

	groups := strings.Fields(output)
	for _, group := range groups {
		if group == "docker" {
			return true, nil
		}
	}

	return false, nil
}

// getDistributionCodename gets the distribution codename (e.g., "jammy", "focal").
// Tries lsb_release first, then falls back to reading /etc/os-release.
func getDistributionCodename() (string, error) {
	// Try lsb_release first
	output, err := exec.RunWithOutput("lsb_release", "-cs")
	if err == nil {
		codename := strings.TrimSpace(output)
		if codename != "" {
			return codename, nil
		}
	}

	// Fallback to /etc/os-release
	if !exec.FileExists("/etc/os-release") {
		return "", fmt.Errorf("cannot determine distribution codename: lsb_release failed and /etc/os-release not found")
	}

	file, err := os.Open("/etc/os-release")
	if err != nil {
		return "", fmt.Errorf("failed to open /etc/os-release: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Look for VERSION_CODENAME or VERSION_ID
		if strings.HasPrefix(line, "VERSION_CODENAME=") {
			codename := strings.TrimPrefix(line, "VERSION_CODENAME=")
			codename = strings.Trim(codename, `"`)
			if codename != "" {
				return codename, nil
			}
		}
		// Fallback to VERSION_ID for Ubuntu versions
		if strings.HasPrefix(line, "VERSION_ID=") {
			versionID := strings.TrimPrefix(line, "VERSION_ID=")
			versionID = strings.Trim(versionID, `"`)
			// Map common Ubuntu version IDs to codenames
			// This is a fallback, lsb_release should work in most cases
			versionMap := map[string]string{
				"22.04": "jammy",
				"20.04": "focal",
				"18.04": "bionic",
				"16.04": "xenial",
			}
			if codename, ok := versionMap[versionID]; ok {
				return codename, nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("failed to read /etc/os-release: %w", err)
	}

	return "", fmt.Errorf("cannot determine distribution codename from /etc/os-release")
}

// IsInstalled checks if Docker is already installed and configured.
// Since IsInstalled() doesn't receive config, it performs generic checks.
// Install() performs specific checks with config and is fully idempotent.
func (m *DockerModule) IsInstalled() (bool, error) {
	// Check if Docker is installed
	installed, err := dockerInstalled()
	if err != nil {
		return false, fmt.Errorf("failed to check Docker installation: %w", err)
	}
	if !installed {
		return false, nil
	}

	// Check if Docker service is running
	running, err := dockerServiceRunning()
	if err != nil {
		return false, fmt.Errorf("failed to check Docker service status: %w", err)
	}
	if !running {
		return false, nil
	}

	// Check if Docker Compose is installed (always check since we can't access config here)
	composeInstalled, err := dockerComposeInstalled()
	if err != nil {
		return false, fmt.Errorf("failed to check Docker Compose installation: %w", err)
	}
	if !composeInstalled {
		return false, nil
	}

	return true, nil
}

// Install installs Docker CE and Docker Compose v2, and adds the user to the docker group.
func (m *DockerModule) Install(cfg *config.Config) error {
	dryRun := log.IsDryRun()

	// Validate config
	if cfg.User.Username == "" {
		return fmt.Errorf("username is required for Docker module installation")
	}

	// Check if Docker is already installed
	dockerInstalled, err := dockerInstalled()
	if err != nil {
		return fmt.Errorf("failed to check Docker installation: %w", err)
	}

	if !dockerInstalled {
		if dryRun {
			log.Info("Would install Docker CE and Docker Compose")
		} else {
			log.Info("Installing Docker CE and Docker Compose")

			// Install prerequisites
			log.Info("Installing prerequisites")
			if err := exec.Run("apt-get", "update"); err != nil {
				return fmt.Errorf("failed to update apt: %w", err)
			}
			if err := exec.Run("apt-get", "install", "-y", "ca-certificates", "curl"); err != nil {
				return fmt.Errorf("failed to install prerequisites: %w", err)
			}

			// Download and add Docker GPG key
			log.Info("Adding Docker GPG key")
			// Download GPG key and pipe through gpg --dearmor to create keyring
			// Use curl to download and pipe to gpg --dearmor
			if err := exec.Run("sh", "-c", fmt.Sprintf("curl -fsSL %s | gpg --dearmor -o %s", dockerGPGKeyURL, dockerGPGKeyringPath)); err != nil {
				return fmt.Errorf("failed to add Docker GPG key: %w", err)
			}

			// Get distribution codename
			codename, err := getDistributionCodename()
			if err != nil {
				return fmt.Errorf("failed to get distribution codename: %w", err)
			}

			// Get architecture
			arch, err := exec.RunWithOutput("dpkg", "--print-architecture")
			if err != nil {
				return fmt.Errorf("failed to get architecture: %w", err)
			}
			arch = strings.TrimSpace(arch)

			// Add Docker repository
			log.Info("Adding Docker repository")
			repoLine := fmt.Sprintf("deb [arch=%s signed-by=%s] %s %s stable\n", arch, dockerGPGKeyringPath, dockerRepoURL, codename)
			if err := exec.WriteFile(dockerAptSourcesPath, []byte(repoLine), 0644); err != nil {
				return fmt.Errorf("failed to add Docker repository: %w", err)
			}

			// Update apt package list
			log.Info("Updating apt package list")
			if err := exec.Run("apt-get", "update"); err != nil {
				return fmt.Errorf("failed to update apt after adding Docker repository: %w", err)
			}

			// Install Docker CE packages
			log.Info("Installing Docker CE packages")
			if err := exec.Run("apt-get", "install", "-y", "docker-ce", "docker-ce-cli", "containerd.io", "docker-buildx-plugin", "docker-compose-plugin"); err != nil {
				return fmt.Errorf("failed to install Docker packages: %w", err)
			}

			// Verify Docker installation
			log.Info("Verifying Docker installation")
			if err := exec.Run("docker", "--version"); err != nil {
				return fmt.Errorf("Docker installation verification failed: %w", err)
			}

			// Start and enable Docker service
			log.Info("Starting Docker service")
			if err := exec.Run("systemctl", "enable", "--now", "docker"); err != nil {
				return fmt.Errorf("failed to start Docker service: %w", err)
			}

			// Verify Docker service is running
			running, err := dockerServiceRunning()
			if err != nil {
				return fmt.Errorf("failed to verify Docker service status: %w", err)
			}
			if !running {
				return fmt.Errorf("Docker service is not running after start")
			}

			log.Success("Docker CE installed and started")
		}
	} else {
		log.Skip("Docker is already installed")
	}

	// Verify Docker Compose if enabled
	if cfg.Docker.InstallCompose {
		composeInstalled, err := dockerComposeInstalled()
		if err != nil {
			return fmt.Errorf("failed to check Docker Compose installation: %w", err)
		}

		if !composeInstalled {
			return fmt.Errorf("Docker Compose is not installed but InstallCompose is enabled")
		}

		if dryRun {
			log.Info("Would verify Docker Compose installation")
		} else {
			log.Info("Verifying Docker Compose installation")
			if err := exec.Run("docker", "compose", "version"); err != nil {
				return fmt.Errorf("Docker Compose verification failed: %w", err)
			}
			log.Success("Docker Compose is installed")
		}
	} else {
		log.Skip("Docker Compose installation is disabled")
	}

	// Add user to docker group (only if user exists)
	if cfg.User.Username == "" {
		log.Skip("No username configured, skipping docker group membership")
	} else if !userExists(cfg.User.Username) {
		log.Warn("User %s does not exist on the system, skipping docker group membership. Ensure user module runs before docker module.", cfg.User.Username)
	} else {
		inGroup, err := userInDockerGroup(cfg.User.Username)
		if err != nil {
			return fmt.Errorf("failed to check docker group membership: %w", err)
		}

		if !inGroup {
			if dryRun {
				log.Info("Would add user %s to docker group", cfg.User.Username)
			} else {
				log.Info("Adding user %s to docker group", cfg.User.Username)
				if err := exec.Run("usermod", "-aG", "docker", cfg.User.Username); err != nil {
					return fmt.Errorf("failed to add user to docker group: %w", err)
				}
				log.Success("User %s added to docker group", cfg.User.Username)
			}
			log.Warn("User %s has been added to the docker group. Please logout and login again for the changes to take effect.", cfg.User.Username)
		} else {
			log.Skip("User %s is already in docker group", cfg.User.Username)
		}
	}

	if !dryRun {
		log.Success("Docker module installation completed successfully")
	}

	return nil
}

// Ensure DockerModule implements the Module interface
var _ module.Module = (*DockerModule)(nil)

