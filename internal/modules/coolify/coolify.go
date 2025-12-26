package coolify

import (
	"fmt"
	"strings"

	"github.com/stwalsh4118/phanes/internal/config"
	"github.com/stwalsh4118/phanes/internal/exec"
	"github.com/stwalsh4118/phanes/internal/log"
	"github.com/stwalsh4118/phanes/internal/module"
)

const (
	coolifyInstallScript = "https://cdn.coollabs.io/coolify/install.sh"
)

// CoolifyModule implements the Module interface for Coolify installation.
type CoolifyModule struct{}

// Name returns the unique name identifier for this module.
func (m *CoolifyModule) Name() string {
	return "coolify"
}

// Description returns a human-readable description of what this module does.
func (m *CoolifyModule) Description() string {
	return "Installs and configures Coolify self-hosted PaaS"
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

// checkDockerDependency checks if Docker is installed and running.
// Returns an error if Docker is not available.
func checkDockerDependency() error {
	installed, err := dockerInstalled()
	if err != nil {
		return fmt.Errorf("failed to check Docker installation: %w", err)
	}
	if !installed {
		return fmt.Errorf("Docker is not installed. Please install Docker before installing Coolify")
	}

	running, err := dockerServiceRunning()
	if err != nil {
		return fmt.Errorf("failed to check Docker service status: %w", err)
	}
	if !running {
		return fmt.Errorf("Docker service is not running. Please start Docker before installing Coolify")
	}

	return nil
}

// coolifyContainersRunning checks if Coolify containers are running.
// Uses docker ps to list containers and checks for "coolify" in container names.
func coolifyContainersRunning() (bool, error) {
	output, err := exec.RunWithOutput("docker", "ps", "--format", "{{.Names}}")
	if err != nil {
		return false, fmt.Errorf("failed to list Docker containers: %w", err)
	}

	// Check if any container name contains "coolify"
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), "coolify") {
			return true, nil
		}
	}

	return false, nil
}

// IsInstalled checks if Coolify is already installed and running.
func (m *CoolifyModule) IsInstalled() (bool, error) {
	// First check Docker dependency
	if err := checkDockerDependency(); err != nil {
		return false, nil
	}

	// Check if Coolify containers are running
	running, err := coolifyContainersRunning()
	if err != nil {
		return false, fmt.Errorf("failed to check Coolify installation: %w", err)
	}

	return running, nil
}

// Install installs Coolify using the official install script.
func (m *CoolifyModule) Install(cfg *config.Config) error {
	dryRun := log.IsDryRun()

	// Check if Coolify is enabled
	if !cfg.Coolify.Enabled {
		log.Skip("Coolify installation is disabled")
		return nil
	}

	// Check Docker dependency before proceeding
	if err := checkDockerDependency(); err != nil {
		log.Warn("Docker dependency check failed: %v", err)
		return fmt.Errorf("Docker dependency check failed: %w", err)
	}

	// Check if Coolify is already installed
	installed, err := m.IsInstalled()
	if err != nil {
		return fmt.Errorf("failed to check Coolify installation status: %w", err)
	}

	if installed {
		log.Skip("Coolify is already installed and running")
		return nil
	}

	if dryRun {
		log.Info("Would install Coolify using official install script")
		log.Info("Would run: curl -fsSL %s | bash", coolifyInstallScript)
		log.Info("Would verify Coolify containers are running after installation")
		return nil
	}

	// Install Coolify using official install script
	log.Info("Installing Coolify using official install script")
	installCmd := fmt.Sprintf("curl -fsSL %s | bash", coolifyInstallScript)
	if err := exec.Run("sh", "-c", installCmd); err != nil {
		return fmt.Errorf("failed to install Coolify: %w", err)
	}

	// Verify installation by checking if containers are running
	log.Info("Verifying Coolify installation")
	running, err := coolifyContainersRunning()
	if err != nil {
		return fmt.Errorf("failed to verify Coolify installation: %w", err)
	}

	if !running {
		return fmt.Errorf("Coolify installation completed but containers are not running. Please check Docker logs")
	}

	log.Success("Coolify installed successfully")
	log.Info("Coolify is now running!")
	log.Info("Access the dashboard at: http://localhost:8000")
	log.Info("On first visit, you'll create an admin account")

	return nil
}

// Ensure CoolifyModule implements the Module interface
var _ module.Module = (*CoolifyModule)(nil)

