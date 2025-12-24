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
	timezone, err := exec.RunWithOutput("timedatectl", "show", "-p", "Timezone", "--value")
	if err != nil {
		return false, fmt.Errorf("failed to check timezone: %w", err)
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
	if err := exec.Run("timedatectl", "set-timezone", timezone); err != nil {
		return fmt.Errorf("failed to set timezone: %w", err)
	}

	// Verify timezone was set correctly
	actualTimezone, err := exec.RunWithOutput("timedatectl", "show", "-p", "Timezone", "--value")
	if err != nil {
		return fmt.Errorf("failed to verify timezone: %w", err)
	}
	actualTimezone = strings.TrimSpace(actualTimezone)
	if actualTimezone != timezone {
		return fmt.Errorf("timezone verification failed: expected %s, got %s", timezone, actualTimezone)
	}
	log.Success("Timezone set to %s", timezone)

	// Configure locale
	log.Info("Configuring locale %s", defaultLocale)
	if err := exec.Run("locale-gen", defaultLocale); err != nil {
		return fmt.Errorf("failed to generate locale: %w", err)
	}

	if err := exec.Run("update-locale", fmt.Sprintf("LANG=%s", defaultLocale)); err != nil {
		return fmt.Errorf("failed to update locale: %w", err)
	}

	// Verify locale is configured
	locale, err := exec.RunWithOutput("locale")
	if err != nil {
		return fmt.Errorf("failed to verify locale: %w", err)
	}
	if !strings.Contains(strings.ToUpper(locale), "UTF-8") {
		return fmt.Errorf("locale verification failed: UTF-8 not found in locale output")
	}
	log.Success("Locale configured to %s", defaultLocale)

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
