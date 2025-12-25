package monitoring

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
	netdataKickstartURL = "https://get.netdata.cloud/kickstart.sh"
	netdataDefaultPort  = 19999
	netdataServiceName  = "netdata"
	netdataBinaryPath   = "/usr/sbin/netdata"
	kickstartScriptPath = "/tmp/netdata-kickstart.sh"
)

// MonitoringModule implements the Module interface for Netdata monitoring installation.
type MonitoringModule struct{}

// Name returns the unique name identifier for this module.
func (m *MonitoringModule) Name() string {
	return "monitoring"
}

// Description returns a human-readable description of what this module does.
func (m *MonitoringModule) Description() string {
	return "Installs and configures Netdata monitoring"
}

// netdataInstalled checks if Netdata is installed by checking if the binary exists.
func netdataInstalled() (bool, error) {
	if exec.FileExists(netdataBinaryPath) {
		return true, nil
	}
	// Also check if service exists as fallback
	output, err := exec.RunWithOutput("systemctl", "list-unit-files", "--type=service", "--no-pager")
	if err == nil {
		if strings.Contains(output, netdataServiceName+".service") {
			return true, nil
		}
	}
	return false, nil
}

// netdataServiceRunning checks if Netdata service is running.
func netdataServiceRunning() (bool, error) {
	output, err := exec.RunWithOutput("systemctl", "is-active", netdataServiceName)
	if err != nil {
		return false, nil
	}
	return strings.TrimSpace(output) == "active", nil
}

// netdataServiceEnabled checks if Netdata service is enabled.
func netdataServiceEnabled() (bool, error) {
	output, err := exec.RunWithOutput("systemctl", "is-enabled", netdataServiceName)
	if err != nil {
		return false, nil
	}
	status := strings.TrimSpace(output)
	return status == "enabled" || status == "enabled-runtime", nil
}

// netdataPortAccessible checks if Netdata is listening on port 19999.
func netdataPortAccessible() (bool, error) {
	// Try ss first (more modern), fallback to netstat
	output, err := exec.RunWithOutput("ss", "-tlnp")
	if err != nil {
		// Fallback to netstat
		output, err = exec.RunWithOutput("netstat", "-tlnp")
		if err != nil {
			return false, fmt.Errorf("failed to check port accessibility: %w", err)
		}
	}

	// Check if port 19999 is in the output
	portStr := fmt.Sprintf(":%d", netdataDefaultPort)
	if strings.Contains(output, portStr) {
		return true, nil
	}

	return false, nil
}

// IsInstalled checks if Netdata is already installed and configured.
// Since IsInstalled() doesn't receive config, it performs generic checks.
// Install() performs specific checks with config and is fully idempotent.
func (m *MonitoringModule) IsInstalled() (bool, error) {
	// Check if Netdata is installed
	installed, err := netdataInstalled()
	if err != nil {
		return false, fmt.Errorf("failed to check Netdata installation: %w", err)
	}
	if !installed {
		return false, nil
	}

	// Check if Netdata service is running
	running, err := netdataServiceRunning()
	if err != nil {
		return false, fmt.Errorf("failed to check Netdata service status: %w", err)
	}
	if !running {
		return false, nil
	}

	// Check if Netdata port is accessible
	accessible, err := netdataPortAccessible()
	if err != nil {
		return false, fmt.Errorf("failed to check Netdata port accessibility: %w", err)
	}
	if !accessible {
		return false, nil
	}

	return true, nil
}

// Install installs and configures Netdata using the official kickstart script.
func (m *MonitoringModule) Install(cfg *config.Config) error {
	dryRun := log.IsDryRun()

	// Check if Netdata is already installed
	installed, err := netdataInstalled()
	if err != nil {
		return fmt.Errorf("failed to check Netdata installation: %w", err)
	}

	if !installed {
		if dryRun {
			log.Info("Would install Netdata monitoring")
		} else {
			log.Info("Installing Netdata monitoring")

			// Download kickstart script
			log.Info("Downloading Netdata kickstart script")
			if err := exec.Run("curl", "-fsSL", netdataKickstartURL, "-o", kickstartScriptPath); err != nil {
				return fmt.Errorf("failed to download Netdata kickstart script: %w", err)
			}

			// Make script executable
			if err := exec.Run("chmod", "+x", kickstartScriptPath); err != nil {
				return fmt.Errorf("failed to make kickstart script executable: %w", err)
			}

			// Run kickstart script in non-interactive mode
			log.Info("Running Netdata kickstart script (this may take a few minutes)")
			// The kickstart script provides its own progress output, so we let it stream to stdout/stderr
			if err := exec.Run("bash", kickstartScriptPath, "--non-interactive"); err != nil {
				// Clean up script even on error
				os.Remove(kickstartScriptPath)
				return fmt.Errorf("failed to run Netdata kickstart script: %w", err)
			}

			// Clean up kickstart script after successful installation
			if err := os.Remove(kickstartScriptPath); err != nil {
				log.Warn("Failed to remove kickstart script: %v", err)
			}

			log.Success("Netdata installed successfully")
		}
	} else {
		log.Skip("Netdata is already installed")
	}

	// Configure service to start on boot
	enabled, err := netdataServiceEnabled()
	if err != nil {
		return fmt.Errorf("failed to check if Netdata service is enabled: %w", err)
	}

	if !enabled {
		if dryRun {
			log.Info("Would enable Netdata service to start on boot")
		} else {
			log.Info("Enabling Netdata service to start on boot")
			if err := exec.Run("systemctl", "enable", netdataServiceName); err != nil {
				return fmt.Errorf("failed to enable Netdata service: %w", err)
			}
			log.Success("Netdata service enabled")
		}
	} else {
		log.Skip("Netdata service is already enabled")
	}

	// Start service if not running
	running, err := netdataServiceRunning()
	if err != nil {
		return fmt.Errorf("failed to check Netdata service status: %w", err)
	}

	if !running {
		if dryRun {
			log.Info("Would start Netdata service")
		} else {
			log.Info("Starting Netdata service")
			if err := exec.Run("systemctl", "start", netdataServiceName); err != nil {
				return fmt.Errorf("failed to start Netdata service: %w", err)
			}

			// Verify service is running
			running, err := netdataServiceRunning()
			if err != nil {
				return fmt.Errorf("failed to verify Netdata service status: %w", err)
			}
			if !running {
				return fmt.Errorf("Netdata service is not running after start")
			}

			log.Success("Netdata service started")
		}
	} else {
		log.Skip("Netdata service is already running")
	}

	// Verify Netdata is accessible
	accessible, err := netdataPortAccessible()
	if err != nil {
		return fmt.Errorf("failed to verify Netdata port accessibility: %w", err)
	}

	if !accessible {
		if dryRun {
			log.Info("Would verify Netdata is accessible on port %d", netdataDefaultPort)
		} else {
			// Give service a moment to fully start
			log.Warn("Netdata port %d is not yet accessible. The service may still be starting.", netdataDefaultPort)
		}
	} else {
		if !dryRun {
			log.Success("Netdata is accessible at http://localhost:%d", netdataDefaultPort)
		}
	}

	if !dryRun {
		log.Success("Netdata monitoring module installation completed successfully")
	}

	return nil
}

// Ensure MonitoringModule implements the Module interface
var _ module.Module = (*MonitoringModule)(nil)

