package updates

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
	unattendedUpgradesConfigPath = "/etc/apt/apt.conf.d/50unattended-upgrades"
	autoUpgradesConfigPath       = "/etc/apt/apt.conf.d/20auto-upgrades"
)

// UpdatesModule implements the Module interface for automatic security updates configuration.
type UpdatesModule struct{}

// Name returns the unique name identifier for this module.
func (m *UpdatesModule) Name() string {
	return "updates"
}

// Description returns a human-readable description of what this module does.
func (m *UpdatesModule) Description() string {
	return "Configures automatic security updates"
}

// unattendedUpgradesInstalled checks if the unattended-upgrades package is installed.
func unattendedUpgradesInstalled() (bool, error) {
	// Try dpkg -l first (most reliable for Debian/Ubuntu)
	if exec.CommandExists("dpkg") {
		output, err := exec.RunWithOutput("dpkg", "-l", "unattended-upgrades")
		if err == nil {
			// dpkg -l returns 0 even if package is not installed, but output will indicate status
			// Look for "ii" which means installed and configured
			return strings.Contains(output, "unattended-upgrades") && strings.Contains(output, "ii"), nil
		}
	}

	// Fallback: check if command exists in PATH
	if exec.CommandExists("unattended-upgrades") {
		return true, nil
	}

	// Fallback: check if binary exists
	if exec.FileExists("/usr/bin/unattended-upgrades") || exec.FileExists("/usr/sbin/unattended-upgrades") {
		return true, nil
	}

	return false, nil
}

// generate50UnattendedUpgrades generates the content for 50unattended-upgrades configuration file.
func generate50UnattendedUpgrades() string {
	return `// Automatically upgrade packages from these (origin:archive) pairs
Unattended-Upgrade::Allowed-Origins {
	"${distro_id}:${distro_codename}-security";
	"${distro_id}ESMApps:${distro_codename}-apps-security";
	"${distro_id}ESM:${distro_codename}-infra-security";
};

// List of packages to not update (regexp are supported)
Unattended-Upgrade::Package-Blacklist {
};

// This option allows you to control if on a unclean dpkg exit
// unattended-upgrades will automatically run dpkg --configure -a
Unattended-Upgrade::AutoFixInterruptedDpkg "true";

// Split the upgrade into the smallest possible chunks so that
// they can be interrupted with SIGUSR1. This makes the upgrade
// a bit slower but it has the benefit that shutdown while a upgrade
// is running is possible (with a small delay)
Unattended-Upgrade::MinimalSteps "true";

// Install all unattended-upgrades when the machine is shutting down
// instead of doing it in the background while the machine is running
// This will (obviously) make shutdown slower
// Unattended-Upgrade::InstallOnShutdown "false";

// Send email to this address for problems or packages upgrades
// If empty or unset then no email is sent, make sure that you
// have a working mail setup on your system. A package that provides
// 'mailx' must be installed. E.g. "user@example.com"
// Unattended-Upgrade::Mail "";

// Set this value to "true" to get emails only on errors. Default
// is to always send a mail if Unattended-Upgrade::Mail is set
// Unattended-Upgrade::MailOnlyOnError "true";

// Do automatic removal of new unused dependencies after the upgrade
// (equivalent to apt-get autoremove)
Unattended-Upgrade::Remove-Unused-Kernel-Packages "true";
Unattended-Upgrade::Remove-New-Unused-Dependencies "true";
Unattended-Upgrade::Remove-Unused-Dependencies "true";

// Automatically reboot *WITHOUT CONFIRMATION* if
//  the file /var/run/reboot-required exists after the upgrade
Unattended-Upgrade::Automatic-Reboot "false";

// Automatically reboot even if there are users currently logged in
// when Unattended-Upgrade::Automatic-Reboot is set to true
Unattended-Upgrade::Automatic-Reboot-WithUsers "false";

// If automatic reboot is enabled and needed, reboot at the specific
// time instead of immediately
//  Default: "now"
Unattended-Upgrade::Automatic-Reboot-Time "02:00";

// Use 'aptitude safe-upgrade' to minimize the risk of breaking changes
// Unattended-Upgrade::UseAptitude "false";

// Enable logging to syslog. Default is False
Unattended-Upgrade::SyslogEnable "false";

// Specify syslog facility. Default is daemon. Give a string here, no syslog.conf
// Unattended-Upgrade::SyslogFacility "daemon";

// Download and install upgrades only on AC power
// (i.e. skip or gracefully stop updates on battery)
Unattended-Upgrade::OnlyOnACPower "false";

// Download and install upgrades only on unmetered connection
// (i.e. skip or gracefully stop updates on a metered connection)
Unattended-Upgrade::SkipUpdatesOnMeteredConnection "false";

// Verbose logging
Unattended-Upgrade::Verbose "false";

// Print debugging information through syslog
Unattended-Upgrade::Debug "false";

// Allow package downgrade if Pin-Priority exceeds 1000
Unattended-Upgrade::Allow-downgrade "false";
`
}

// generate20AutoUpgrades generates the content for 20auto-upgrades configuration file.
func generate20AutoUpgrades() string {
	return `// Enable automatic updates
APT::Periodic::Update-Package-Lists "1";
APT::Periodic::Unattended-Upgrade "1";
APT::Periodic::Download-Upgradeable-Packages "1";
APT::Periodic::AutocleanInterval "7";
`
}

// configFileMatches checks if a configuration file exists and matches the expected content.
func configFileMatches(filePath string, expectedContent string) (bool, error) {
	if !exec.FileExists(filePath) {
		return false, nil
	}

	// Read current config
	currentContent, err := os.ReadFile(filePath)
	if err != nil {
		return false, fmt.Errorf("failed to read %s: %w", filePath, err)
	}

	// Normalize both contents for comparison (remove comments and normalize whitespace)
	currentNormalized := normalizeConfig(string(currentContent))
	expectedNormalized := normalizeConfig(expectedContent)

	// Compare normalized configs
	return currentNormalized == expectedNormalized, nil
}

// normalizeConfig normalizes configuration by removing comments and normalizing whitespace.
func normalizeConfig(config string) string {
	lines := strings.Split(config, "\n")
	var normalized []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		// Remove inline comments
		if idx := strings.Index(line, "//"); idx != -1 {
			line = strings.TrimSpace(line[:idx])
		}
		if line != "" {
			// Normalize whitespace: collapse multiple spaces into single space
			fields := strings.Fields(line)
			normalizedLine := strings.Join(fields, " ")
			// Remove spaces before semicolons
			normalizedLine = strings.ReplaceAll(normalizedLine, " ;", ";")
			normalized = append(normalized, normalizedLine)
		}
	}

	return strings.Join(normalized, "\n")
}

// unattendedUpgradesConfigured checks if unattended-upgrades is properly configured.
func unattendedUpgradesConfigured() (bool, error) {
	// Check 50unattended-upgrades config
	expected50Config := generate50UnattendedUpgrades()
	matches50, err := configFileMatches(unattendedUpgradesConfigPath, expected50Config)
	if err != nil {
		return false, fmt.Errorf("failed to check 50unattended-upgrades config: %w", err)
	}
	if !matches50 {
		return false, nil
	}

	// Check 20auto-upgrades config
	expected20Config := generate20AutoUpgrades()
	matches20, err := configFileMatches(autoUpgradesConfigPath, expected20Config)
	if err != nil {
		return false, fmt.Errorf("failed to check 20auto-upgrades config: %w", err)
	}
	if !matches20 {
		return false, nil
	}

	return true, nil
}

// IsInstalled checks if the updates module is already installed.
// It verifies that unattended-upgrades is installed and properly configured.
func (m *UpdatesModule) IsInstalled() (bool, error) {
	// Check if package is installed
	installed, err := unattendedUpgradesInstalled()
	if err != nil {
		return false, fmt.Errorf("failed to check if unattended-upgrades is installed: %w", err)
	}
	if !installed {
		return false, nil
	}

	// Check if configuration files are correct
	configured, err := unattendedUpgradesConfigured()
	if err != nil {
		return false, fmt.Errorf("failed to check unattended-upgrades configuration: %w", err)
	}
	if !configured {
		return false, nil
	}

	return true, nil
}

// Install installs and configures unattended-upgrades for automatic security updates.
func (m *UpdatesModule) Install(cfg *config.Config) error {
	dryRun := log.IsDryRun()

	// Install unattended-upgrades package
	installed, err := unattendedUpgradesInstalled()
	if err != nil {
		return fmt.Errorf("failed to check if unattended-upgrades is installed: %w", err)
	}

	if !installed {
		if dryRun {
			log.Info("Would install unattended-upgrades package")
		} else {
			log.Info("Installing unattended-upgrades package")
			if err := exec.Run("apt-get", "install", "-y", "unattended-upgrades"); err != nil {
				return fmt.Errorf("failed to install unattended-upgrades: %w", err)
			}
			log.Success("unattended-upgrades package installed")
		}
	} else {
		log.Skip("unattended-upgrades package already installed")
	}

	// Configure 50unattended-upgrades
	config50 := generate50UnattendedUpgrades()
	matches50, err := configFileMatches(unattendedUpgradesConfigPath, config50)
	if err != nil {
		return fmt.Errorf("failed to check 50unattended-upgrades config: %w", err)
	}

	if !matches50 {
		if dryRun {
			log.Info("Would configure %s", unattendedUpgradesConfigPath)
		} else {
			log.Info("Configuring %s", unattendedUpgradesConfigPath)
			if err := exec.WriteFile(unattendedUpgradesConfigPath, []byte(config50), 0644); err != nil {
				return fmt.Errorf("failed to write %s: %w", unattendedUpgradesConfigPath, err)
			}
			log.Success("Configured %s", unattendedUpgradesConfigPath)
		}
	} else {
		log.Skip("%s already configured", unattendedUpgradesConfigPath)
	}

	// Configure 20auto-upgrades
	config20 := generate20AutoUpgrades()
	matches20, err := configFileMatches(autoUpgradesConfigPath, config20)
	if err != nil {
		return fmt.Errorf("failed to check 20auto-upgrades config: %w", err)
	}

	if !matches20 {
		if dryRun {
			log.Info("Would configure %s", autoUpgradesConfigPath)
		} else {
			log.Info("Configuring %s", autoUpgradesConfigPath)
			if err := exec.WriteFile(autoUpgradesConfigPath, []byte(config20), 0644); err != nil {
				return fmt.Errorf("failed to write %s: %w", autoUpgradesConfigPath, err)
			}
			log.Success("Configured %s", autoUpgradesConfigPath)
		}
	} else {
		log.Skip("%s already configured", autoUpgradesConfigPath)
	}

	// Verify configuration (optional, but helpful)
	if !dryRun && exec.CommandExists("unattended-upgrades") {
		log.Info("Verifying unattended-upgrades configuration")
		// Run dry-run to test configuration
		// Note: This may produce output, but it's informational
		if err := exec.Run("unattended-upgrades", "--dry-run", "--debug"); err != nil {
			// Don't fail on verification errors - config might be valid but command might fail for other reasons
			log.Warn("unattended-upgrades verification returned an error (this may be expected)")
		}
	}

	if !dryRun {
		log.Success("Updates module installation completed successfully")
		log.Info("Automatic security updates are now enabled. Automatic reboot is disabled by default.")
	}

	return nil
}

// Ensure UpdatesModule implements the Module interface
var _ module.Module = (*UpdatesModule)(nil)

