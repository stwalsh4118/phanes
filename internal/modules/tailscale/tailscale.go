package tailscale

import (
	"fmt"
	"strings"

	"github.com/stwalsh4118/phanes/internal/config"
	"github.com/stwalsh4118/phanes/internal/exec"
	"github.com/stwalsh4118/phanes/internal/log"
	"github.com/stwalsh4118/phanes/internal/module"
)

const (
	tailscaleInstallScript = "https://tailscale.com/install.sh"
	tailscaleServiceName    = "tailscaled"
)

// TailscaleModule implements the Module interface for Tailscale VPN installation.
type TailscaleModule struct{}

// Name returns the unique name identifier for this module.
func (m *TailscaleModule) Name() string {
	return "tailscale"
}

// Description returns a human-readable description of what this module does.
func (m *TailscaleModule) Description() string {
	return "Installs and configures Tailscale VPN"
}

// tailscaleInstalled checks if Tailscale is installed by checking if the tailscale command exists.
func tailscaleInstalled() (bool, error) {
	return exec.CommandExists("tailscale"), nil
}

// tailscaleConnected checks if Tailscale is authenticated and connected.
// Runs tailscale status and checks for successful exit (non-error means connected).
func tailscaleConnected() (bool, error) {
	err := exec.Run("tailscale", "status")
	if err != nil {
		return false, nil
	}
	return true, nil
}

// tailscaleServiceEnabled checks if the tailscaled systemd service is enabled.
func tailscaleServiceEnabled() (bool, error) {
	output, err := exec.RunWithOutput("systemctl", "is-enabled", tailscaleServiceName)
	if err != nil {
		return false, nil
	}
	return strings.TrimSpace(output) == "enabled", nil
}

// IsInstalled checks if Tailscale is already installed, authenticated, and the service is enabled.
func (m *TailscaleModule) IsInstalled() (bool, error) {
	// Check if tailscale command exists
	installed, err := tailscaleInstalled()
	if err != nil {
		return false, fmt.Errorf("failed to check Tailscale installation: %w", err)
	}
	if !installed {
		return false, nil
	}

	// Check if Tailscale is connected/authenticated
	connected, err := tailscaleConnected()
	if err != nil {
		return false, fmt.Errorf("failed to check Tailscale connection status: %w", err)
	}
	if !connected {
		return false, nil
	}

	// Check if tailscaled service is enabled
	enabled, err := tailscaleServiceEnabled()
	if err != nil {
		return false, fmt.Errorf("failed to check Tailscale service status: %w", err)
	}
	if !enabled {
		return false, nil
	}

	return true, nil
}

// Install installs and configures Tailscale using the official install script.
func (m *TailscaleModule) Install(cfg *config.Config) error {
	dryRun := log.IsDryRun()

	// Check if Tailscale is enabled
	if !cfg.Tailscale.Enabled {
		log.Skip("Tailscale installation is disabled")
		return nil
	}

	// Validate auth key is provided (unless manual auth is enabled)
	if !cfg.Tailscale.SkipAuth && cfg.Tailscale.AuthKey == "" {
		return fmt.Errorf("tailscale.auth_key is required when tailscale is enabled and skip_auth is false")
	}

	// Check if Tailscale is already installed
	installed, err := m.IsInstalled()
	if err != nil {
		return fmt.Errorf("failed to check Tailscale installation status: %w", err)
	}

	if installed {
		log.Skip("Tailscale is already installed and configured")
		return nil
	}

	if dryRun {
		log.Info("Would install Tailscale using official install script")
		log.Info("Would run: curl -fsSL %s | sh", tailscaleInstallScript)
		if cfg.Tailscale.SkipAuth {
			log.Info("Would skip authentication (manual login enabled)")
			log.Info("Would enable and start tailscaled systemd service")
			log.Info("You would need to manually run 'tailscale up' to authenticate")
		} else {
			log.Info("Would authenticate Tailscale with provided auth key")
			log.Info("Would enable and start tailscaled systemd service")
			log.Info("Would display Tailscale status after installation")
		}
		return nil
	}

	// Install Tailscale using official install script
	log.Info("Installing Tailscale using official install script")
	installCmd := fmt.Sprintf("curl -fsSL %s | sh", tailscaleInstallScript)
	if err := exec.Run("sh", "-c", installCmd); err != nil {
		return fmt.Errorf("failed to install Tailscale: %w", err)
	}

	// Authenticate Tailscale (unless manual auth is enabled)
	if cfg.Tailscale.SkipAuth {
		log.Info("Skipping automatic authentication (skip_auth is enabled)")
		log.Info("To authenticate manually, run: tailscale up")
		log.Info("This will open a browser window for authentication")
	} else {
		log.Info("Authenticating Tailscale with provided auth key")
		if err := exec.Run("tailscale", "up", "--authkey", cfg.Tailscale.AuthKey); err != nil {
			return fmt.Errorf("failed to authenticate Tailscale: %w", err)
		}
	}

	// Enable and start tailscaled service
	log.Info("Enabling and starting tailscaled service")
	if err := exec.Run("systemctl", "enable", "--now", tailscaleServiceName); err != nil {
		return fmt.Errorf("failed to enable Tailscale service: %w", err)
	}

	// Display Tailscale status (only if authenticated)
	if !cfg.Tailscale.SkipAuth {
		log.Info("Tailscale status:")
		statusOutput, err := exec.RunWithOutput("tailscale", "status")
		if err != nil {
			log.Warn("Failed to get Tailscale status: %v", err)
		} else {
			// Extract and display the Tailscale IP address if available
			lines := strings.Split(strings.TrimSpace(statusOutput), "\n")
			for _, line := range lines {
				if strings.Contains(line, "100.") {
					log.Info("%s", line)
				}
			}
			// Also show the full status output
			log.Info("Full status output:")
			log.Info("%s", statusOutput)
		}
	}

	if cfg.Tailscale.SkipAuth {
		log.Success("Tailscale installed successfully")
		log.Info("Run 'tailscale up' to authenticate manually")
	} else {
		log.Success("Tailscale installed and configured successfully")
	}

	return nil
}

// Ensure TailscaleModule implements the Module interface
var _ module.Module = (*TailscaleModule)(nil)

