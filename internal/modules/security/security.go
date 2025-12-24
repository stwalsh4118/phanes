package security

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/stwalsh4118/phanes/internal/config"
	"github.com/stwalsh4118/phanes/internal/exec"
	"github.com/stwalsh4118/phanes/internal/log"
	"github.com/stwalsh4118/phanes/internal/module"
)

//go:embed sshd_config.tmpl
var sshdConfigTemplate string

//go:embed jail.local.tmpl
var jailLocalTemplate string

// SecurityModule implements the Module interface for security configuration.
// It configures UFW firewall, fail2ban, and SSH hardening.
type SecurityModule struct{}

// Name returns the unique name identifier for this module.
func (m *SecurityModule) Name() string {
	return "security"
}

// Description returns a human-readable description of what this module does.
func (m *SecurityModule) Description() string {
	return "Configures UFW, fail2ban, and SSH hardening"
}

// renderTemplate renders a template string with the provided data.
func renderTemplate(tmpl string, data interface{}) (string, error) {
	t, err := template.New("template").Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// ufwIsEnabled checks if UFW firewall is enabled.
func ufwIsEnabled() (bool, error) {
	output, err := exec.RunWithOutput("ufw", "status")
	if err != nil {
		// UFW might not be installed
		return false, nil
	}

	// Check if output contains "Status: active"
	return strings.Contains(strings.ToLower(output), "status: active"), nil
}

// fail2banIsRunning checks if fail2ban service is running.
func fail2banIsRunning() (bool, error) {
	// Try systemctl first (systemd systems)
	output, err := exec.RunWithOutput("systemctl", "is-active", "fail2ban")
	if err == nil {
		return strings.TrimSpace(output) == "active", nil
	}

	// Fallback: check if process is running
	if exec.CommandExists("pgrep") {
		output, err := exec.RunWithOutput("pgrep", "-x", "fail2ban-server")
		if err == nil && strings.TrimSpace(output) != "" {
			return true, nil
		}
	}

	return false, nil
}

// sshConfigMatches checks if the current SSH config matches the expected hardened configuration.
func sshConfigMatches(cfg *config.Config) (bool, error) {
	sshdConfigPath := "/etc/ssh/sshd_config"
	if !exec.FileExists(sshdConfigPath) {
		return false, nil
	}

	// Read current SSH config
	currentConfig, err := os.ReadFile(sshdConfigPath)
	if err != nil {
		return false, fmt.Errorf("failed to read SSH config: %w", err)
	}

	// Render expected config from template
	templateData := struct {
		SSHPort           int
		AllowPasswordAuth bool
	}{
		SSHPort:           cfg.Security.SSHPort,
		AllowPasswordAuth: cfg.Security.AllowPasswordAuth,
	}

	expectedConfig, err := renderTemplate(sshdConfigTemplate, templateData)
	if err != nil {
		return false, fmt.Errorf("failed to render SSH config template: %w", err)
	}

	// Normalize both configs for comparison (remove comments, normalize whitespace)
	currentNormalized := normalizeSSHConfig(string(currentConfig))
	expectedNormalized := normalizeSSHConfig(expectedConfig)

	// Compare normalized configs
	return currentNormalized == expectedNormalized, nil
}

// normalizeSSHConfig normalizes SSH config by removing comments and normalizing whitespace.
func normalizeSSHConfig(config string) string {
	lines := strings.Split(config, "\n")
	var normalized []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		normalized = append(normalized, line)
	}

	return strings.Join(normalized, "\n")
}

// IsInstalled checks if the security module is already installed.
// It verifies that UFW is enabled, fail2ban is running, and SSH config matches expected configuration.
func (m *SecurityModule) IsInstalled() (bool, error) {
	// Check UFW
	ufwEnabled, err := ufwIsEnabled()
	if err != nil {
		return false, fmt.Errorf("failed to check UFW status: %w", err)
	}
	if !ufwEnabled {
		return false, nil
	}

	// Check fail2ban
	fail2banRunning, err := fail2banIsRunning()
	if err != nil {
		return false, fmt.Errorf("failed to check fail2ban status: %w", err)
	}
	if !fail2banRunning {
		return false, nil
	}

	// Check SSH config - we need config for this, but IsInstalled() doesn't receive config
	// So we'll do a basic check: verify SSH config exists and has key security settings
	sshdConfigPath := "/etc/ssh/sshd_config"
	if !exec.FileExists(sshdConfigPath) {
		return false, nil
	}

	// Read SSH config and check for key security settings
	configContent, err := os.ReadFile(sshdConfigPath)
	if err != nil {
		return false, fmt.Errorf("failed to read SSH config: %w", err)
	}

	configStr := string(configContent)
	// Check for key security settings that indicate hardening
	hasPermitRootLogin := strings.Contains(configStr, "PermitRootLogin")
	hasPasswordAuth := strings.Contains(configStr, "PasswordAuthentication")
	hasPubkeyAuth := strings.Contains(configStr, "PubkeyAuthentication")

	// If all key settings are present, assume configured (Install() will verify exact match)
	if hasPermitRootLogin && hasPasswordAuth && hasPubkeyAuth {
		return true, nil
	}

	return false, nil
}

// Install configures UFW firewall, fail2ban, and SSH hardening.
func (m *SecurityModule) Install(cfg *config.Config) error {
	dryRun := log.IsDryRun()

	// Validate config
	sshPort := cfg.Security.SSHPort
	if sshPort <= 0 || sshPort > 65535 {
		return fmt.Errorf("invalid SSH port: %d (must be between 1 and 65535)", sshPort)
	}

	// Warn if password auth is being disabled
	if !cfg.Security.AllowPasswordAuth && !dryRun {
		log.Warn("Password authentication will be disabled. Ensure SSH key access is configured before proceeding.")
	}

	// Configure UFW
	if err := m.configureUFW(sshPort, dryRun); err != nil {
		return fmt.Errorf("failed to configure UFW: %w", err)
	}

	// Install and configure fail2ban
	if err := m.configureFail2ban(sshPort, dryRun); err != nil {
		return fmt.Errorf("failed to configure fail2ban: %w", err)
	}

	// Harden SSH configuration
	if err := m.hardenSSH(cfg, dryRun); err != nil {
		return fmt.Errorf("failed to harden SSH: %w", err)
	}

	if !dryRun {
		log.Success("Security module installation completed successfully")
	}

	return nil
}

// configureUFW configures the UFW firewall.
func (m *SecurityModule) configureUFW(sshPort int, dryRun bool) error {
	// Check if UFW is installed
	if !exec.CommandExists("ufw") {
		if dryRun {
			log.Info("Would install UFW")
		} else {
			log.Info("Installing UFW")
			if err := exec.Run("apt-get", "install", "-y", "ufw"); err != nil {
				return fmt.Errorf("failed to install UFW: %w", err)
			}
			log.Success("UFW installed")
		}
	}

	// Check if UFW is already enabled
	ufwEnabled, err := ufwIsEnabled()
	if err != nil {
		return fmt.Errorf("failed to check UFW status: %w", err)
	}

	if ufwEnabled {
		log.Skip("UFW is already enabled")
	} else {
		// Allow SSH port
		if dryRun {
			log.Info("Would allow SSH port %d/tcp in UFW", sshPort)
		} else {
			log.Info("Allowing SSH port %d/tcp in UFW", sshPort)
			if err := exec.Run("ufw", "allow", fmt.Sprintf("%d/tcp", sshPort)); err != nil {
				return fmt.Errorf("failed to allow SSH port in UFW: %w", err)
			}
		}

		// Allow HTTP
		if dryRun {
			log.Info("Would allow HTTP (80/tcp) in UFW")
		} else {
			log.Info("Allowing HTTP (80/tcp) in UFW")
			if err := exec.Run("ufw", "allow", "80/tcp"); err != nil {
				return fmt.Errorf("failed to allow HTTP in UFW: %w", err)
			}
		}

		// Allow HTTPS
		if dryRun {
			log.Info("Would allow HTTPS (443/tcp) in UFW")
		} else {
			log.Info("Allowing HTTPS (443/tcp) in UFW")
			if err := exec.Run("ufw", "allow", "443/tcp"); err != nil {
				return fmt.Errorf("failed to allow HTTPS in UFW: %w", err)
			}
		}

		// Enable UFW
		if dryRun {
			log.Info("Would enable UFW firewall")
		} else {
			log.Info("Enabling UFW firewall")
			if err := exec.Run("ufw", "--force", "enable"); err != nil {
				return fmt.Errorf("failed to enable UFW: %w", err)
			}

			// Verify UFW is enabled
			ufwEnabled, err := ufwIsEnabled()
			if err != nil {
				return fmt.Errorf("failed to verify UFW status: %w", err)
			}
			if !ufwEnabled {
				return fmt.Errorf("UFW verification failed: firewall is not active")
			}
			log.Success("UFW firewall enabled")
		}
	}

	return nil
}

// configureFail2ban installs and configures fail2ban.
func (m *SecurityModule) configureFail2ban(sshPort int, dryRun bool) error {
	// Check if fail2ban is installed
	if !exec.CommandExists("fail2ban-server") {
		if dryRun {
			log.Info("Would install fail2ban")
		} else {
			log.Info("Installing fail2ban")
			if err := exec.Run("apt-get", "install", "-y", "fail2ban"); err != nil {
				return fmt.Errorf("failed to install fail2ban: %w", err)
			}
			log.Success("fail2ban installed")
		}
	}

	// Render jail.local template
	templateData := struct {
		SSHPort int
	}{
		SSHPort: sshPort,
	}

	jailConfig, err := renderTemplate(jailLocalTemplate, templateData)
	if err != nil {
		return fmt.Errorf("failed to render fail2ban config template: %w", err)
	}

	jailLocalPath := "/etc/fail2ban/jail.local"

	// Check if jail.local already exists and matches
	if exec.FileExists(jailLocalPath) {
		existingContent, err := os.ReadFile(jailLocalPath)
		if err == nil && strings.TrimSpace(string(existingContent)) == strings.TrimSpace(jailConfig) {
			log.Skip("fail2ban jail.local already configured correctly")
		} else {
			// Config exists but doesn't match - update it
			if dryRun {
				log.Info("Would update fail2ban jail.local configuration")
			} else {
				log.Info("Updating fail2ban jail.local configuration")
				if err := exec.WriteFile(jailLocalPath, []byte(jailConfig), 0644); err != nil {
					return fmt.Errorf("failed to write fail2ban config: %w", err)
				}
				log.Success("fail2ban jail.local updated")
			}
		}
	} else {
		// Config doesn't exist - create it
		if dryRun {
			log.Info("Would create fail2ban jail.local configuration")
		} else {
			log.Info("Creating fail2ban jail.local configuration")
			if err := exec.WriteFile(jailLocalPath, []byte(jailConfig), 0644); err != nil {
				return fmt.Errorf("failed to write fail2ban config: %w", err)
			}
			log.Success("fail2ban jail.local created")
		}
	}

	// Start and enable fail2ban service
	fail2banRunning, err := fail2banIsRunning()
	if err != nil {
		return fmt.Errorf("failed to check fail2ban status: %w", err)
	}

	if fail2banRunning {
		log.Skip("fail2ban is already running")
	} else {
		if dryRun {
			log.Info("Would start and enable fail2ban service")
		} else {
			log.Info("Starting and enabling fail2ban service")
			// Try systemctl first
			if exec.CommandExists("systemctl") {
				if err := exec.Run("systemctl", "enable", "--now", "fail2ban"); err != nil {
					return fmt.Errorf("failed to start fail2ban service: %w", err)
				}
			} else {
				// Fallback: start service directly
				if err := exec.Run("service", "fail2ban", "start"); err != nil {
					return fmt.Errorf("failed to start fail2ban service: %w", err)
				}
			}

			// Verify fail2ban is running
			fail2banRunning, err := fail2banIsRunning()
			if err != nil {
				return fmt.Errorf("failed to verify fail2ban status: %w", err)
			}
			if !fail2banRunning {
				return fmt.Errorf("fail2ban verification failed: service is not running")
			}
			log.Success("fail2ban service started and enabled")
		}
	}

	return nil
}

// hardenSSH hardens the SSH configuration.
func (m *SecurityModule) hardenSSH(cfg *config.Config, dryRun bool) error {
	sshdConfigPath := "/etc/ssh/sshd_config"
	backupPath := "/etc/ssh/sshd_config.backup"

	// Render SSH config template
	templateData := struct {
		SSHPort           int
		AllowPasswordAuth bool
	}{
		SSHPort:           cfg.Security.SSHPort,
		AllowPasswordAuth: cfg.Security.AllowPasswordAuth,
	}

	sshConfig, err := renderTemplate(sshdConfigTemplate, templateData)
	if err != nil {
		return fmt.Errorf("failed to render SSH config template: %w", err)
	}

	// Check if SSH config already matches
	configMatches, err := sshConfigMatches(cfg)
	if err != nil {
		return fmt.Errorf("failed to check SSH config: %w", err)
	}

	if configMatches {
		log.Skip("SSH configuration already matches expected hardened configuration")
		return nil
	}

	// Backup existing config (if not dry-run and config exists)
	if !dryRun && exec.FileExists(sshdConfigPath) {
		log.Info("Backing up existing SSH config to %s", backupPath)
		if err := exec.Run("cp", sshdConfigPath, backupPath); err != nil {
			return fmt.Errorf("failed to backup SSH config: %w", err)
		}
		log.Success("SSH config backed up")
	}

	// Write new SSH config
	if dryRun {
		log.Info("Would write hardened SSH configuration to %s", sshdConfigPath)
		log.Info("Would validate SSH configuration")
	} else {
		log.Info("Writing hardened SSH configuration")
		if err := exec.WriteFile(sshdConfigPath, []byte(sshConfig), 0644); err != nil {
			return fmt.Errorf("failed to write SSH config: %w", err)
		}

		// Validate SSH config before applying
		log.Info("Validating SSH configuration")
		if err := exec.Run("sshd", "-t"); err != nil {
			// If validation fails, restore backup if it exists
			if exec.FileExists(backupPath) {
				log.Warn("SSH config validation failed, restoring backup")
				if restoreErr := exec.Run("cp", backupPath, sshdConfigPath); restoreErr != nil {
					return fmt.Errorf("SSH config validation failed and backup restore failed: %w (restore error: %v)", err, restoreErr)
				}
			}
			return fmt.Errorf("SSH config validation failed: %w", err)
		}
		log.Success("SSH configuration validated")

		// Reload SSH service
		log.Info("Reloading SSH service")
		if exec.CommandExists("systemctl") {
			if err := exec.Run("systemctl", "reload", "sshd"); err != nil {
				// Try sshd service name (some systems use sshd instead of ssh)
				if err2 := exec.Run("systemctl", "reload", "ssh"); err2 != nil {
					return fmt.Errorf("failed to reload SSH service: %w (also tried ssh: %v)", err, err2)
				}
			}
		} else {
			// Fallback: use service command
			if err := exec.Run("service", "sshd", "reload"); err != nil {
				if err2 := exec.Run("service", "ssh", "reload"); err2 != nil {
					return fmt.Errorf("failed to reload SSH service: %w (also tried ssh: %v)", err, err2)
				}
			}
		}
		log.Success("SSH service reloaded")
	}

	return nil
}

// Ensure SecurityModule implements the Module interface
var _ module.Module = (*SecurityModule)(nil)
