package baseline

import (
	"fmt"
	"strings"

	"github.com/stwalsh4118/phanes/internal/config"
	"github.com/stwalsh4118/phanes/internal/exec"
	"github.com/stwalsh4118/phanes/internal/log"
	"github.com/stwalsh4118/phanes/internal/module"
)

const (
	defaultLocale = "en_US.UTF-8"
)

// BaselineModule implements the Module interface for baseline server configuration.
// It configures timezone, locale, and runs apt update.
type BaselineModule struct{}

// Name returns the unique name identifier for this module.
func (m *BaselineModule) Name() string {
	return "baseline"
}

// Description returns a human-readable description of what this module does.
func (m *BaselineModule) Description() string {
	return "Sets timezone, locale, and runs apt update"
}

// IsInstalled checks if the baseline configuration is already applied.
// It verifies that a timezone is set (not empty) and that the locale
// is configured with UTF-8. Note: Since IsInstalled() doesn't receive
// config, it checks if timezone/locale are configured, not if they match
// a specific configured value.
func (m *BaselineModule) IsInstalled() (bool, error) {
	// Check current timezone
	// Try timedatectl first (requires systemd), fallback to /etc/timezone if not available
	var timezone string
	var err error
	
	timezone, err = exec.RunWithOutput("timedatectl", "show", "-p", "Timezone", "--value")
	if err != nil {
		// timedatectl not available (e.g., in Docker containers without systemd)
		// Fallback to reading /etc/timezone
		if exec.FileExists("/etc/timezone") {
			timezone, err = exec.RunWithOutput("cat", "/etc/timezone")
			if err != nil {
				return false, fmt.Errorf("failed to check timezone: %w", err)
			}
		} else {
			// If neither method works, assume not installed
			return false, nil
		}
	}
	
	timezone = strings.TrimSpace(timezone)
	if timezone == "" {
		return false, nil
	}

	// Check locale - read LANG from locale output or /etc/default/locale
	var lang string
	var err2 error

	// Try to get LANG from locale command output
	localeOutput, err2 := exec.RunWithOutput("locale")
	if err2 == nil {
		// Parse LANG from locale output (format: LANG=en_US.UTF-8)
		lines := strings.Split(localeOutput, "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "LANG=") {
				lang = strings.TrimPrefix(line, "LANG=")
				lang = strings.Trim(lang, "\"")
				break
			}
		}
	}

	// Fallback: check /etc/default/locale if locale command failed or LANG not found
	if lang == "" && exec.FileExists("/etc/default/locale") {
		localeContent, err3 := exec.RunWithOutput("grep", "^LANG=", "/etc/default/locale")
		if err3 == nil {
			lang = strings.TrimSpace(localeContent)
			if strings.Contains(lang, "=") {
				lang = strings.SplitN(lang, "=", 2)[1]
				lang = strings.Trim(lang, "\"")
			}
		}
	}

	if lang == "" {
		return false, fmt.Errorf("failed to determine locale: %w", err2)
	}

	// Check if LANG contains UTF-8
	if !strings.Contains(strings.ToUpper(lang), "UTF-8") {
		return false, nil
	}

	return true, nil
}

// Install configures the timezone, locale, and runs apt update.
// It uses the provided config to get the timezone setting.
func (m *BaselineModule) Install(cfg *config.Config) error {
	timezone := cfg.System.Timezone
	if timezone == "" {
		timezone = "UTC"
	}

	// Set timezone
	log.Info("Setting timezone to %s", timezone)
	// Try timedatectl first (requires systemd), fallback to /etc/timezone if not available
	err := exec.Run("timedatectl", "set-timezone", timezone)
	if err != nil {
		// timedatectl not available (e.g., in Docker containers without systemd)
		// Fallback to writing /etc/timezone and creating symlink
		log.Info("timedatectl not available, using /etc/timezone method")
		if err := exec.WriteFile("/etc/timezone", []byte(timezone+"\n"), 0644); err != nil {
			return fmt.Errorf("failed to set timezone: %w", err)
		}
		// Note: Creating /etc/localtime symlink requires the timezone data files
		// For containers, we'll just set /etc/timezone and skip the symlink
		log.Info("Timezone set to %s (via /etc/timezone)", timezone)
	} else {
		// Verify timezone was set correctly using timedatectl
		actualTimezone, err := exec.RunWithOutput("timedatectl", "show", "-p", "Timezone", "--value")
		if err != nil {
			return fmt.Errorf("failed to verify timezone: %w", err)
		}
		actualTimezone = strings.TrimSpace(actualTimezone)
		if actualTimezone != timezone {
			return fmt.Errorf("timezone verification failed: expected %s, got %s", timezone, actualTimezone)
		}
		log.Success("Timezone set to %s", timezone)
	}

	// Configure locale
	log.Info("Configuring locale %s", defaultLocale)
	if err := exec.Run("locale-gen", defaultLocale); err != nil {
		return fmt.Errorf("failed to generate locale: %w", err)
	}

	if err := exec.Run("update-locale", fmt.Sprintf("LANG=%s", defaultLocale)); err != nil {
		return fmt.Errorf("failed to update locale: %w", err)
	}

	// Verify locale is configured
	// Note: In containers, the locale may not be active in the current shell session
	// but it's been generated and configured. Check /etc/default/locale as verification.
	locale, err := exec.RunWithOutput("locale")
	if err != nil {
		// If locale command fails, check /etc/default/locale as fallback
		if exec.FileExists("/etc/default/locale") {
			localeContent, err2 := exec.RunWithOutput("grep", "^LANG=", "/etc/default/locale")
			if err2 == nil && strings.Contains(strings.ToUpper(localeContent), "UTF-8") {
				log.Success("Locale configured to %s (verified via /etc/default/locale)", defaultLocale)
				// Continue - locale is configured even if not active in current shell
			} else {
				return fmt.Errorf("failed to verify locale: %w", err)
			}
		} else {
			return fmt.Errorf("failed to verify locale: %w", err)
		}
	} else {
		// Check if UTF-8 is in locale output or in /etc/default/locale
		if strings.Contains(strings.ToUpper(locale), "UTF-8") {
			log.Success("Locale configured to %s", defaultLocale)
		} else if exec.FileExists("/etc/default/locale") {
			// Fallback: check /etc/default/locale
			localeContent, err2 := exec.RunWithOutput("grep", "^LANG=", "/etc/default/locale")
			if err2 == nil && strings.Contains(strings.ToUpper(localeContent), "UTF-8") {
				log.Success("Locale configured to %s (verified via /etc/default/locale)", defaultLocale)
			} else {
				return fmt.Errorf("locale verification failed: UTF-8 not found in locale output or /etc/default/locale")
			}
		} else {
			return fmt.Errorf("locale verification failed: UTF-8 not found in locale output")
		}
	}

	// Run apt update
	log.Info("Running apt-get update")
	if err := exec.Run("apt-get", "update"); err != nil {
		return fmt.Errorf("failed to run apt-get update: %w", err)
	}
	log.Success("apt-get update completed successfully")

	return nil
}

// Ensure BaselineModule implements the Module interface
var _ module.Module = (*BaselineModule)(nil)
