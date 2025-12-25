package devtools

import (
	"fmt"
	"strings"

	"github.com/stwalsh4118/phanes/internal/config"
	"github.com/stwalsh4118/phanes/internal/exec"
	"github.com/stwalsh4118/phanes/internal/log"
)

const (
	packageGit            = "git"
	packageBuildEssential = "build-essential"
	packageCurl           = "curl"
	packageWget           = "wget"
	packageCaCertificates = "ca-certificates"
)

// gitInstalled checks if Git is installed.
func gitInstalled() (bool, error) {
	return exec.CommandExists("git"), nil
}

// buildEssentialInstalled checks if build-essential package is installed.
// This checks for the package itself and verifies gcc and make are available.
func buildEssentialInstalled() (bool, error) {
	// Check if build-essential package is installed via dpkg
	if exec.CommandExists("dpkg") {
		output, err := exec.RunWithOutput("dpkg", "-l", packageBuildEssential)
		if err == nil {
			// dpkg -l returns 0 even if package is not installed, but output will indicate status
			// Look for "ii" which means installed and configured
			if strings.Contains(output, packageBuildEssential) && strings.Contains(output, "ii") {
				// Also verify gcc and make are available (they come with build-essential)
				if exec.CommandExists("gcc") && exec.CommandExists("make") {
					return true, nil
				}
			}
		}
	}

	// Fallback: check if gcc and make are available (they might be installed separately)
	if exec.CommandExists("gcc") && exec.CommandExists("make") {
		return true, nil
	}

	return false, nil
}

// curlInstalled checks if curl is installed.
func curlInstalled() (bool, error) {
	return exec.CommandExists("curl"), nil
}

// wgetInstalled checks if wget is installed.
func wgetInstalled() (bool, error) {
	return exec.CommandExists("wget"), nil
}

// caCertificatesInstalled checks if ca-certificates package is installed.
func caCertificatesInstalled() (bool, error) {
	// Check if ca-certificates package is installed via dpkg
	if exec.CommandExists("dpkg") {
		output, err := exec.RunWithOutput("dpkg", "-l", packageCaCertificates)
		if err == nil {
			// dpkg -l returns 0 even if package is not installed, but output will indicate status
			// Look for "ii" which means installed and configured
			if strings.Contains(output, packageCaCertificates) && strings.Contains(output, "ii") {
				return true, nil
			}
		}
	}

	// Fallback: check if update-ca-certificates command exists (comes with ca-certificates)
	if exec.CommandExists("update-ca-certificates") {
		return true, nil
	}

	return false, nil
}

// coreToolsInstalled checks if all core tools are installed.
// Returns true only if ALL tools are installed.
func coreToolsInstalled() (bool, error) {
	gitOk, err := gitInstalled()
	if err != nil {
		return false, fmt.Errorf("failed to check Git installation: %w", err)
	}
	if !gitOk {
		return false, nil
	}

	buildEssentialOk, err := buildEssentialInstalled()
	if err != nil {
		return false, fmt.Errorf("failed to check build-essential installation: %w", err)
	}
	if !buildEssentialOk {
		return false, nil
	}

	curlOk, err := curlInstalled()
	if err != nil {
		return false, fmt.Errorf("failed to check curl installation: %w", err)
	}
	if !curlOk {
		return false, nil
	}

	wgetOk, err := wgetInstalled()
	if err != nil {
		return false, fmt.Errorf("failed to check wget installation: %w", err)
	}
	if !wgetOk {
		return false, nil
	}

	caCertOk, err := caCertificatesInstalled()
	if err != nil {
		return false, fmt.Errorf("failed to check ca-certificates installation: %w", err)
	}
	if !caCertOk {
		return false, nil
	}

	return true, nil
}

// installCoreTools installs core development tools via apt.
// Installs: git, build-essential, curl, wget, ca-certificates
func installCoreTools(cfg *config.Config) error {
	dryRun := log.IsDryRun()

	// Check if all tools are already installed
	installed, err := coreToolsInstalled()
	if err != nil {
		return fmt.Errorf("failed to check core tools installation: %w", err)
	}

	if installed {
		log.Skip("Core development tools are already installed")
		return nil
	}

	if dryRun {
		log.Info("Would install core development tools: git, build-essential, curl, wget, ca-certificates")
		return nil
	}

	log.Info("Installing core development tools")

	// Update apt package list
	log.Info("Updating apt package list")
	if err := exec.Run("apt-get", "update"); err != nil {
		return fmt.Errorf("failed to update apt: %w", err)
	}

	// Install all packages in one command
	log.Info("Installing packages: git, build-essential, curl, wget, ca-certificates")
	if err := exec.Run("apt-get", "install", "-y", packageGit, packageBuildEssential, packageCurl, packageWget, packageCaCertificates); err != nil {
		return fmt.Errorf("failed to install core development tools: %w", err)
	}

	// Verify installation
	installed, err = coreToolsInstalled()
	if err != nil {
		return fmt.Errorf("failed to verify core tools installation: %w", err)
	}
	if !installed {
		return fmt.Errorf("core tools installation verification failed: not all tools are installed")
	}

	log.Success("Core development tools installed successfully")
	return nil
}

