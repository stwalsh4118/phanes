package caddy

import (
	"fmt"
	"os"
	"strings"

	"github.com/stwalsh4118/phanes/internal/config"
	"github.com/stwalsh4118/phanes/internal/exec"
	"github.com/stwalsh4118/phanes/internal/log"
	"github.com/stwalsh4118/phanes/internal/module"
)

const (
	caddyServiceName        = "caddy"
	caddyDefaultPort        = 80
	caddyBinaryPath         = "/usr/bin/caddy"
	caddyConfigDir          = "/etc/caddy"
	caddyfilePath           = "/etc/caddy/Caddyfile"
	caddyGPGKeyURL          = "https://dl.cloudsmith.io/public/caddy/stable/gpg.key"
	caddyRepositoryURL       = "https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt"
	caddyGPGKeyringPath     = "/usr/share/keyrings/caddy-stable-archive-keyring.gpg"
	caddyAptSourcesPath     = "/etc/apt/sources.list.d/caddy-stable.list"
	defaultCaddyfileContent = `localhost {
	respond "Caddy is running!"
}`
)

// CaddyModule implements the Module interface for Caddy web server installation.
type CaddyModule struct{}

// Name returns the unique name identifier for this module.
func (m *CaddyModule) Name() string {
	return "caddy"
}

// Description returns a human-readable description of what this module does.
func (m *CaddyModule) Description() string {
	return "Installs and configures Caddy web server with automatic HTTPS"
}

// caddyInstalled checks if Caddy is installed by checking if the binary exists.
func caddyInstalled() (bool, error) {
	if exec.FileExists(caddyBinaryPath) {
		return true, nil
	}
	// Also check if service exists as fallback
	output, err := exec.RunWithOutput("systemctl", "list-unit-files", "--type=service", "--no-pager")
	if err == nil {
		if strings.Contains(output, caddyServiceName+".service") {
			return true, nil
		}
	}
	// Try caddy version as another fallback
	if exec.CommandExists("caddy") {
		_, err := exec.RunWithOutput("caddy", "version")
		if err == nil {
			return true, nil
		}
	}
	return false, nil
}

// caddyServiceRunning checks if Caddy service is running.
func caddyServiceRunning() (bool, error) {
	output, err := exec.RunWithOutput("systemctl", "is-active", caddyServiceName)
	if err != nil {
		return false, nil
	}
	return strings.TrimSpace(output) == "active", nil
}

// caddyServiceEnabled checks if Caddy service is enabled.
func caddyServiceEnabled() (bool, error) {
	output, err := exec.RunWithOutput("systemctl", "is-enabled", caddyServiceName)
	if err != nil {
		return false, nil
	}
	status := strings.TrimSpace(output)
	return status == "enabled" || status == "enabled-runtime", nil
}

// caddyPortAccessible checks if Caddy is listening on port 80.
func caddyPortAccessible() (bool, error) {
	// Try ss first (more modern), fallback to netstat
	output, err := exec.RunWithOutput("ss", "-tlnp")
	if err != nil {
		// Fallback to netstat
		output, err = exec.RunWithOutput("netstat", "-tlnp")
		if err != nil {
			return false, fmt.Errorf("failed to check port accessibility: %w", err)
		}
	}

	// Check if port 80 is in the output
	portStr := fmt.Sprintf(":%d", caddyDefaultPort)
	if strings.Contains(output, portStr) {
		return true, nil
	}

	return false, nil
}

// port80InUse checks if port 80 is already in use by another service (not Caddy).
func port80InUse() (bool, error) {
	// Try ss first (more modern), fallback to netstat
	output, err := exec.RunWithOutput("ss", "-tlnp")
	if err != nil {
		// Fallback to netstat
		output, err = exec.RunWithOutput("netstat", "-tlnp")
		if err != nil {
			return false, fmt.Errorf("failed to check port 80 usage: %w", err)
		}
	}

	// Check if port 80 is in the output
	portStr := fmt.Sprintf(":%d", caddyDefaultPort)
	if !strings.Contains(output, portStr) {
		return false, nil
	}

	// Port 80 is in use, check if it's Caddy
	// Check if caddy is mentioned in the output
	if strings.Contains(strings.ToLower(output), caddyServiceName) {
		// It's Caddy, so port is not in use by another service
		return false, nil
	}

	// Port 80 is in use by another service
	return true, nil
}

// caddyfileExists checks if the Caddyfile exists.
func caddyfileExists() (bool, error) {
	return exec.FileExists(caddyfilePath), nil
}

// createDefaultCaddyfile creates the default Caddyfile if it doesn't exist.
func createDefaultCaddyfile() error {
	// Create config directory if it doesn't exist
	if err := os.MkdirAll(caddyConfigDir, 0755); err != nil {
		return fmt.Errorf("failed to create Caddy config directory: %w", err)
	}

	// Write default Caddyfile content
	content := []byte(defaultCaddyfileContent)
	if err := exec.WriteFile(caddyfilePath, content, 0644); err != nil {
		return fmt.Errorf("failed to create default Caddyfile: %w", err)
	}

	return nil
}

// IsInstalled checks if Caddy is already installed and configured.
// Since IsInstalled() doesn't receive config, it performs generic checks.
// Install() performs specific checks with config and is fully idempotent.
func (m *CaddyModule) IsInstalled() (bool, error) {
	// Check if Caddy is installed
	installed, err := caddyInstalled()
	if err != nil {
		return false, fmt.Errorf("failed to check Caddy installation: %w", err)
	}
	if !installed {
		return false, nil
	}

	// Check if Caddy service is running
	running, err := caddyServiceRunning()
	if err != nil {
		return false, fmt.Errorf("failed to check Caddy service status: %w", err)
	}
	if !running {
		return false, nil
	}

	// Check if Caddy port is accessible
	accessible, err := caddyPortAccessible()
	if err != nil {
		return false, fmt.Errorf("failed to check Caddy port accessibility: %w", err)
	}
	if !accessible {
		return false, nil
	}

	return true, nil
}

// Install installs and configures Caddy web server.
func (m *CaddyModule) Install(cfg *config.Config) error {
	dryRun := log.IsDryRun()

	// Check if Caddy is enabled in config
	if !cfg.Caddy.Enabled {
		log.Skip("Caddy module is disabled in configuration")
		return nil
	}

	// Check if port 80 is in use by another service
	inUse, err := port80InUse()
	if err != nil {
		return fmt.Errorf("failed to check port 80 usage: %w", err)
	}
	if inUse {
		log.Warn("Port 80 is already in use. Caddy may not be able to bind to this port.")
	}

	// Check if Caddy is already installed
	installed, err := caddyInstalled()
	if err != nil {
		return fmt.Errorf("failed to check Caddy installation: %w", err)
	}

	if !installed {
		if dryRun {
			log.Info("Would install Caddy web server")
		} else {
			log.Info("Installing Caddy web server")

			// Install prerequisites
			log.Info("Installing prerequisites")
			if err := exec.Run("apt-get", "update"); err != nil {
				return fmt.Errorf("failed to update package list: %w", err)
			}
			if err := exec.Run("apt-get", "install", "-y", "debian-keyring", "debian-archive-keyring", "apt-transport-https", "curl"); err != nil {
				return fmt.Errorf("failed to install prerequisites: %w", err)
			}

			// Download and install GPG key
			log.Info("Adding Caddy GPG key")
			// Download GPG key and pipe to gpg --dearmor
			// Using curl to download and pipe to gpg
			cmd := fmt.Sprintf("curl -1sLf '%s' | gpg --dearmor -o %s", caddyGPGKeyURL, caddyGPGKeyringPath)
			if err := exec.Run("bash", "-c", cmd); err != nil {
				return fmt.Errorf("failed to add Caddy GPG key: %w", err)
			}

			// Add Caddy repository
			log.Info("Adding Caddy repository")
			cmd = fmt.Sprintf("curl -1sLf '%s' | tee %s", caddyRepositoryURL, caddyAptSourcesPath)
			if err := exec.Run("bash", "-c", cmd); err != nil {
				return fmt.Errorf("failed to add Caddy repository: %w", err)
			}

			// Update apt package list
			log.Info("Updating package list")
			if err := exec.Run("apt-get", "update"); err != nil {
				return fmt.Errorf("failed to update package list: %w", err)
			}

			// Install caddy
			log.Info("Installing caddy package")
			if err := exec.Run("apt-get", "install", "-y", "caddy"); err != nil {
				return fmt.Errorf("failed to install caddy: %w", err)
			}

			// Verify installation
			if err := exec.Run("caddy", "version"); err != nil {
				return fmt.Errorf("failed to verify caddy installation: %w", err)
			}

			log.Success("Caddy installed successfully")
		}
	} else {
		log.Skip("Caddy is already installed")
	}

	// Create default Caddyfile if it doesn't exist
	exists, err := caddyfileExists()
	if err != nil {
		return fmt.Errorf("failed to check if Caddyfile exists: %w", err)
	}

	if !exists {
		if dryRun {
			log.Info("Would create default Caddyfile")
		} else {
			log.Info("Creating default Caddyfile")
			if err := createDefaultCaddyfile(); err != nil {
				return fmt.Errorf("failed to create default Caddyfile: %w", err)
			}
			log.Success("Default Caddyfile created")
		}
	} else {
		log.Skip("Caddyfile already exists")
	}

	// Configure service to start on boot
	enabled, err := caddyServiceEnabled()
	if err != nil {
		return fmt.Errorf("failed to check if Caddy service is enabled: %w", err)
	}

	if !enabled {
		if dryRun {
			log.Info("Would enable Caddy service to start on boot")
		} else {
			log.Info("Enabling Caddy service to start on boot")
			if err := exec.Run("systemctl", "enable", caddyServiceName); err != nil {
				return fmt.Errorf("failed to enable Caddy service: %w", err)
			}
			log.Success("Caddy service enabled")
		}
	} else {
		log.Skip("Caddy service is already enabled")
	}

	// Start service if not running
	running, err := caddyServiceRunning()
	if err != nil {
		return fmt.Errorf("failed to check Caddy service status: %w", err)
	}

	if !running {
		if dryRun {
			log.Info("Would start Caddy service")
		} else {
			log.Info("Starting Caddy service")
			if err := exec.Run("systemctl", "start", caddyServiceName); err != nil {
				return fmt.Errorf("failed to start Caddy service: %w", err)
			}

			// Verify service is running
			running, err := caddyServiceRunning()
			if err != nil {
				return fmt.Errorf("failed to verify Caddy service status: %w", err)
			}
			if !running {
				return fmt.Errorf("Caddy service is not running after start")
			}

			log.Success("Caddy service started")
		}
	} else {
		log.Skip("Caddy service is already running")
	}

	// Verify Caddy is accessible
	accessible, err := caddyPortAccessible()
	if err != nil {
		return fmt.Errorf("failed to verify Caddy port accessibility: %w", err)
	}

	if !accessible {
		if dryRun {
			log.Info("Would verify Caddy is accessible on port %d", caddyDefaultPort)
		} else {
			// Give service a moment to fully start
			log.Warn("Caddy port %d is not yet accessible. The service may still be starting.", caddyDefaultPort)
		}
	} else {
		if !dryRun {
			log.Success("Caddy is accessible at http://localhost")
			log.Info("Caddy provides automatic HTTPS certificates via Let's Encrypt")
		}
	}

	if !dryRun {
		log.Success("Caddy web server module installation completed successfully")
	}

	return nil
}

// Ensure CaddyModule implements the Module interface
var _ module.Module = (*CaddyModule)(nil)


