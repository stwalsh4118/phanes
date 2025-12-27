package nginx

import (
	"fmt"
	"strings"

	"github.com/stwalsh4118/phanes/internal/config"
	"github.com/stwalsh4118/phanes/internal/exec"
	"github.com/stwalsh4118/phanes/internal/log"
	"github.com/stwalsh4118/phanes/internal/module"
)

const (
	nginxServiceName = "nginx"
	nginxDefaultPort = 80
	nginxBinaryPath  = "/usr/sbin/nginx"
)

// NginxModule implements the Module interface for Nginx web server installation.
type NginxModule struct{}

// Name returns the unique name identifier for this module.
func (m *NginxModule) Name() string {
	return "nginx"
}

// Description returns a human-readable description of what this module does.
func (m *NginxModule) Description() string {
	return "Installs and configures Nginx web server"
}

// nginxInstalled checks if Nginx is installed by checking if the binary exists.
func nginxInstalled() (bool, error) {
	if exec.FileExists(nginxBinaryPath) {
		return true, nil
	}
	// Also check if service exists as fallback
	output, err := exec.RunWithOutput("systemctl", "list-unit-files", "--type=service", "--no-pager")
	if err == nil {
		if strings.Contains(output, nginxServiceName+".service") {
			return true, nil
		}
	}
	// Try nginx -v as another fallback
	if exec.CommandExists("nginx") {
		_, err := exec.RunWithOutput("nginx", "-v")
		if err == nil {
			return true, nil
		}
	}
	return false, nil
}

// nginxServiceRunning checks if Nginx service is running.
func nginxServiceRunning() (bool, error) {
	output, err := exec.RunWithOutput("systemctl", "is-active", nginxServiceName)
	if err != nil {
		return false, nil
	}
	return strings.TrimSpace(output) == "active", nil
}

// nginxServiceEnabled checks if Nginx service is enabled.
func nginxServiceEnabled() (bool, error) {
	output, err := exec.RunWithOutput("systemctl", "is-enabled", nginxServiceName)
	if err != nil {
		return false, nil
	}
	status := strings.TrimSpace(output)
	return status == "enabled" || status == "enabled-runtime", nil
}

// nginxPortAccessible checks if Nginx is listening on port 80.
func nginxPortAccessible() (bool, error) {
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
	portStr := fmt.Sprintf(":%d", nginxDefaultPort)
	if strings.Contains(output, portStr) {
		return true, nil
	}

	return false, nil
}

// port80InUse checks if port 80 is already in use by another service (not Nginx).
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
	portStr := fmt.Sprintf(":%d", nginxDefaultPort)
	if !strings.Contains(output, portStr) {
		return false, nil
	}

	// Port 80 is in use, check if it's Nginx
	// Check if nginx is mentioned in the output
	if strings.Contains(strings.ToLower(output), nginxServiceName) {
		// It's Nginx, so port is not in use by another service
		return false, nil
	}

	// Port 80 is in use by another service
	return true, nil
}

// IsInstalled checks if Nginx is already installed and configured.
// Since IsInstalled() doesn't receive config, it performs generic checks.
// Install() performs specific checks with config and is fully idempotent.
func (m *NginxModule) IsInstalled() (bool, error) {
	// Check if Nginx is installed
	installed, err := nginxInstalled()
	if err != nil {
		return false, fmt.Errorf("failed to check Nginx installation: %w", err)
	}
	if !installed {
		return false, nil
	}

	// Check if Nginx service is running
	running, err := nginxServiceRunning()
	if err != nil {
		return false, fmt.Errorf("failed to check Nginx service status: %w", err)
	}
	if !running {
		return false, nil
	}

	// Check if Nginx port is accessible
	accessible, err := nginxPortAccessible()
	if err != nil {
		return false, fmt.Errorf("failed to check Nginx port accessibility: %w", err)
	}
	if !accessible {
		return false, nil
	}

	return true, nil
}

// Install installs and configures Nginx web server.
func (m *NginxModule) Install(cfg *config.Config) error {
	dryRun := log.IsDryRun()

	// Check if Nginx is enabled in config
	if !cfg.Nginx.Enabled {
		log.Skip("Nginx module is disabled in configuration")
		return nil
	}

	// Check if port 80 is in use by another service
	inUse, err := port80InUse()
	if err != nil {
		return fmt.Errorf("failed to check port 80 usage: %w", err)
	}
	if inUse {
		log.Warn("Port 80 is already in use. Nginx may not be able to bind to this port.")
	}

	// Check if Nginx is already installed
	installed, err := nginxInstalled()
	if err != nil {
		return fmt.Errorf("failed to check Nginx installation: %w", err)
	}

	if !installed {
		if dryRun {
			log.Info("Would install Nginx web server")
		} else {
			log.Info("Installing Nginx web server")

			// Update apt package list
			log.Info("Updating package list")
			if err := exec.Run("apt-get", "update"); err != nil {
				return fmt.Errorf("failed to update package list: %w", err)
			}

			// Install nginx
			log.Info("Installing nginx package")
			if err := exec.Run("apt-get", "install", "-y", "nginx"); err != nil {
				return fmt.Errorf("failed to install nginx: %w", err)
			}

			// Verify installation
			if err := exec.Run("nginx", "-v"); err != nil {
				return fmt.Errorf("failed to verify nginx installation: %w", err)
			}

			log.Success("Nginx installed successfully")
		}
	} else {
		log.Skip("Nginx is already installed")
	}

	// Configure service to start on boot
	enabled, err := nginxServiceEnabled()
	if err != nil {
		return fmt.Errorf("failed to check if Nginx service is enabled: %w", err)
	}

	if !enabled {
		if dryRun {
			log.Info("Would enable Nginx service to start on boot")
		} else {
			log.Info("Enabling Nginx service to start on boot")
			if err := exec.Run("systemctl", "enable", nginxServiceName); err != nil {
				return fmt.Errorf("failed to enable Nginx service: %w", err)
			}
			log.Success("Nginx service enabled")
		}
	} else {
		log.Skip("Nginx service is already enabled")
	}

	// Start service if not running
	running, err := nginxServiceRunning()
	if err != nil {
		return fmt.Errorf("failed to check Nginx service status: %w", err)
	}

	if !running {
		if dryRun {
			log.Info("Would start Nginx service")
		} else {
			log.Info("Starting Nginx service")
			if err := exec.Run("systemctl", "start", nginxServiceName); err != nil {
				return fmt.Errorf("failed to start Nginx service: %w", err)
			}

			// Verify service is running
			running, err := nginxServiceRunning()
			if err != nil {
				return fmt.Errorf("failed to verify Nginx service status: %w", err)
			}
			if !running {
				return fmt.Errorf("Nginx service is not running after start")
			}

			log.Success("Nginx service started")
		}
	} else {
		log.Skip("Nginx service is already running")
	}

	// Verify Nginx is accessible
	accessible, err := nginxPortAccessible()
	if err != nil {
		return fmt.Errorf("failed to verify Nginx port accessibility: %w", err)
	}

	if !accessible {
		if dryRun {
			log.Info("Would verify Nginx is accessible on port %d", nginxDefaultPort)
		} else {
			// Give service a moment to fully start
			log.Warn("Nginx port %d is not yet accessible. The service may still be starting.", nginxDefaultPort)
		}
	} else {
		if !dryRun {
			log.Success("Nginx is accessible at http://localhost")
		}
	}

	if !dryRun {
		log.Success("Nginx web server module installation completed successfully")
	}

	return nil
}

// Ensure NginxModule implements the Module interface
var _ module.Module = (*NginxModule)(nil)



