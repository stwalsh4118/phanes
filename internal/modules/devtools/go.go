package devtools

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/stwalsh4118/phanes/internal/config"
	"github.com/stwalsh4118/phanes/internal/exec"
	"github.com/stwalsh4118/phanes/internal/log"
)

const (
	goInstallDir = "/usr/local/go"
	goBinDir     = "/usr/local/go/bin"
)

// goPathScript returns the PATH export script that should be added to shell profiles for Go.
func goPathScript() string {
	return `export PATH=$PATH:/usr/local/go/bin
export GOROOT=/usr/local/go
`
}

// getSystemArch detects the system architecture.
// Uses dpkg --print-architecture with fallback to uname -m.
func getSystemArch() (string, error) {
	// Try dpkg first (preferred for Debian/Ubuntu)
	if exec.CommandExists("dpkg") {
		arch, err := exec.RunWithOutput("dpkg", "--print-architecture")
		if err == nil {
			arch = strings.TrimSpace(arch)
			if arch != "" {
				return arch, nil
			}
		}
	}

	// Fallback to uname -m
	arch, err := exec.RunWithOutput("uname", "-m")
	if err != nil {
		return "", fmt.Errorf("failed to detect system architecture: %w", err)
	}

	arch = strings.TrimSpace(arch)
	if arch == "" {
		return "", fmt.Errorf("architecture detection returned empty string")
	}

	return arch, nil
}

// mapArchToGoArch maps system architecture to Go architecture name.
func mapArchToGoArch(arch string) string {
	arch = strings.ToLower(arch)
	switch arch {
	case "amd64", "x86_64":
		return "amd64"
	case "arm64", "aarch64":
		return "arm64"
	case "armv6l", "armhf":
		return "armv6l"
	case "386", "i386", "i686":
		return "386"
	default:
		// Return as-is if no mapping found (might work for some architectures)
		return arch
	}
}

// goBinaryPath is the absolute path to the Go binary.
const goBinaryPath = "/usr/local/go/bin/go"

// goInstalled checks if Go is installed and matches the requested version.
// Checks if go binary exists at /usr/local/go/bin/go and version matches.
// Uses absolute path because the current process may not have Go in PATH yet.
func goInstalled(version string) (bool, error) {
	// Check if Go binary exists at absolute path (don't rely on PATH)
	if !exec.FileExists(goBinaryPath) {
		return false, nil
	}

	// If version is empty, just check if Go is installed
	if version == "" {
		return true, nil
	}

	// Check version matches using absolute path
	output, err := exec.RunWithOutput(goBinaryPath, "version")
	if err != nil {
		return false, fmt.Errorf("failed to check Go version: %w", err)
	}

	// Go version output format: "go version go1.21.0 linux/amd64"
	// Check if version string is present in output
	versionStr := fmt.Sprintf("go%s", version)
	if strings.Contains(output, versionStr) {
		return true, nil
	}

	return false, nil
}

// shellProfileHasGoPath checks if a shell profile already contains Go PATH configuration.
func shellProfileHasGoPath(profilePath string) (bool, error) {
	if !exec.FileExists(profilePath) {
		return false, nil
	}

	content, err := os.ReadFile(profilePath)
	if err != nil {
		return false, fmt.Errorf("failed to read shell profile: %w", err)
	}

	// Check if Go PATH is already present
	// Look for /usr/local/go/bin in PATH export
	contentStr := string(content)
	if strings.Contains(contentStr, "/usr/local/go/bin") && strings.Contains(contentStr, "PATH") {
		return true, nil
	}

	return false, nil
}

// configureShellProfileForGo appends Go PATH to a shell profile if not already present.
func configureShellProfileForGo(profilePath string, userUID, userGID int) error {
	hasGoPath, err := shellProfileHasGoPath(profilePath)
	if err != nil {
		return fmt.Errorf("failed to check shell profile: %w", err)
	}

	if hasGoPath {
		log.Skip("Shell profile %s already has Go PATH configuration", profilePath)
		return nil
	}

	// Read existing content
	var content []byte
	if exec.FileExists(profilePath) {
		existingContent, err := os.ReadFile(profilePath)
		if err != nil {
			return fmt.Errorf("failed to read shell profile: %w", err)
		}
		content = existingContent
		// Add newline if file doesn't end with one
		if len(content) > 0 && content[len(content)-1] != '\n' {
			content = append(content, '\n')
		}
	}

	// Append Go PATH configuration
	content = append(content, []byte("\n# Go PATH configuration\n")...)
	content = append(content, []byte(goPathScript())...)

	// Write file
	if err := exec.WriteFile(profilePath, content, 0644); err != nil {
		return fmt.Errorf("failed to write shell profile: %w", err)
	}

	// Set ownership to the user
	if err := os.Chown(profilePath, userUID, userGID); err != nil {
		return fmt.Errorf("failed to set shell profile ownership: %w", err)
	}

	log.Success("Configured shell profile: %s", profilePath)
	return nil
}

// installGo installs Go from the official source.
// Downloads tarball, extracts to /usr/local/go, and configures shell profiles.
func installGo(cfg *config.Config) error {
	dryRun := log.IsDryRun()

	// Get Go version from config
	goVersion := cfg.DevTools.GoVersion
	if goVersion == "" {
		goVersion = "1.25.5" // Default version - must include patch number
	}

	// Check if Go is already installed with correct version
	goOk, err := goInstalled(goVersion)
	if err != nil {
		return fmt.Errorf("failed to check Go installation: %w", err)
	}

	if goOk {
		log.Skip("Go version %s is already installed", goVersion)
		return nil
	}

	// Validate username is set for shell profile configuration
	if cfg.User.Username == "" {
		log.Warn("Username is not set, skipping shell profile configuration")
		// Continue with installation but skip shell profile config
	}

	username := cfg.User.Username

	// Get user info for file ownership
	var userUID, userGID int
	if !dryRun && username != "" {
		userInfo, err := user.Lookup(username)
		if err != nil {
			return fmt.Errorf("failed to look up user %s: %w", username, err)
		}
		userUID, err = strconv.Atoi(userInfo.Uid)
		if err != nil {
			return fmt.Errorf("failed to parse user UID: %w", err)
		}
		userGID, err = strconv.Atoi(userInfo.Gid)
		if err != nil {
			return fmt.Errorf("failed to parse user GID: %w", err)
		}
	}

	// Detect system architecture
	systemArch, err := getSystemArch()
	if err != nil {
		return fmt.Errorf("failed to detect system architecture: %w", err)
	}

	goArch := mapArchToGoArch(systemArch)
	log.Info("Detected architecture: %s (Go arch: %s)", systemArch, goArch)

	// Build download URL
	downloadURL := fmt.Sprintf("https://go.dev/dl/go%s.linux-%s.tar.gz", goVersion, goArch)
	tarballPath := fmt.Sprintf("/tmp/go%s.linux-%s.tar.gz", goVersion, goArch)

	if dryRun {
		log.Info("Would download Go %s from %s", goVersion, downloadURL)
		log.Info("Would extract to %s", goInstallDir)
		if username != "" {
			log.Info("Would configure shell profiles (.bashrc and .zshrc) for Go")
		}
		return nil
	}

	// Always remove old installation to ensure clean state
	log.Info("Removing existing Go installation at %s (if exists)", goInstallDir)
	if err := exec.Run("rm", "-rf", goInstallDir); err != nil {
		return fmt.Errorf("failed to remove old Go installation: %w", err)
	}

	// Download Go tarball
	log.Info("Downloading Go %s from %s", goVersion, downloadURL)
	if err := exec.Run("curl", "-L", "-o", tarballPath, downloadURL); err != nil {
		return fmt.Errorf("failed to download Go tarball: %w", err)
	}

	// Ensure tarball was downloaded
	if !exec.FileExists(tarballPath) {
		return fmt.Errorf("Go tarball was not downloaded: %s", tarballPath)
	}

	// Validate the downloaded file is actually a gzip archive
	fileInfo, err := os.Stat(tarballPath)
	if err != nil {
		return fmt.Errorf("failed to stat downloaded tarball: %w", err)
	}
	if fileInfo.Size() < 1000000 { // Go tarball should be at least 1MB
		// Read file content to see what was downloaded (likely an error page)
		content, _ := os.ReadFile(tarballPath)
		contentPreview := string(content)
		if len(contentPreview) > 500 {
			contentPreview = contentPreview[:500]
		}
		_ = exec.Run("rm", "-f", tarballPath)
		return fmt.Errorf("downloaded file is too small (%d bytes) - Go version %s may not exist or URL is incorrect. Expected tarball from: %s. Note: Go requires full version like '1.24.0' not just '1.24'. Preview: %s", fileInfo.Size(), goVersion, downloadURL, contentPreview)
	}

	// Extract tarball to /usr/local
	log.Info("Extracting Go to %s", goInstallDir)
	if err := exec.Run("tar", "-C", "/usr/local", "-xzf", tarballPath); err != nil {
		// Clean up tarball on error
		_ = exec.Run("rm", "-f", tarballPath)
		return fmt.Errorf("failed to extract Go tarball: %w", err)
	}

	// Sync filesystem to ensure extraction is complete
	_ = exec.Run("sync")

	// Clean up tarball
	if err := exec.Run("rm", "-f", tarballPath); err != nil {
		log.Warn("Failed to remove temporary tarball: %s", tarballPath)
	}

	// Verify Go binary exists after installation
	// Note: We don't check version immediately after extraction as the OS may cache stale data
	if !exec.FileExists(goBinaryPath) {
		return fmt.Errorf("Go installation verification failed: binary not found at %s", goBinaryPath)
	}

	log.Success("Go %s installed successfully to %s", goVersion, goInstallDir)

	// Configure shell profiles if username is set
	if username != "" {
		homeDir := filepath.Join("/home", username)
		bashrcPath := filepath.Join(homeDir, ".bashrc")
		zshrcPath := filepath.Join(homeDir, ".zshrc")

		// Configure .bashrc
		if exec.FileExists(bashrcPath) {
			if err := configureShellProfileForGo(bashrcPath, userUID, userGID); err != nil {
				return fmt.Errorf("failed to configure .bashrc: %w", err)
			}
		} else {
			// Create .bashrc if it doesn't exist
			if err := configureShellProfileForGo(bashrcPath, userUID, userGID); err != nil {
				return fmt.Errorf("failed to create .bashrc: %w", err)
			}
		}

		// Configure .zshrc
		if exec.FileExists(zshrcPath) {
			if err := configureShellProfileForGo(zshrcPath, userUID, userGID); err != nil {
				return fmt.Errorf("failed to configure .zshrc: %w", err)
			}
		} else {
			// Create .zshrc if it doesn't exist
			if err := configureShellProfileForGo(zshrcPath, userUID, userGID); err != nil {
				return fmt.Errorf("failed to create .zshrc: %w", err)
			}
		}

		log.Success("Shell profiles configured for Go")
	}

	log.Success("Go installation completed successfully")
	return nil
}
