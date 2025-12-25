package swap

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/stwalsh4118/phanes/internal/config"
	"github.com/stwalsh4118/phanes/internal/exec"
	"github.com/stwalsh4118/phanes/internal/log"
	"github.com/stwalsh4118/phanes/internal/module"
)

const (
	defaultSwapFilePath  = "/swapfile"
	defaultSwappiness    = 10
	swappinessConfigPath = "/etc/sysctl.d/99-swappiness.conf"
	fstabPath            = "/etc/fstab"
)

// SwapModule implements the Module interface for swap file creation and configuration.
type SwapModule struct{}

// Name returns the unique name identifier for this module.
func (m *SwapModule) Name() string {
	return "swap"
}

// Description returns a human-readable description of what this module does.
func (m *SwapModule) Description() string {
	return "Creates and configures swap file"
}

// parseSwapSize parses a size string (e.g., "2G", "512M", "1T") and returns the size in bytes.
// Supports formats: G/g (gigabytes), M/m (megabytes), T/t (terabytes).
// Returns an error if the format is invalid.
func parseSwapSize(sizeStr string) (int64, error) {
	if sizeStr == "" {
		return 0, fmt.Errorf("swap size cannot be empty")
	}

	sizeStr = strings.TrimSpace(sizeStr)
	originalSizeStr := sizeStr
	sizeStr = strings.ToUpper(sizeStr)

	// Find where the numeric part ends and unit begins
	// Look for the first non-digit, non-decimal point character
	unitIndex := -1
	for i, r := range sizeStr {
		if r < '0' || r > '9' {
			if r == '.' {
				continue // Allow decimal point
			}
			unitIndex = i
			break
		}
	}

	if unitIndex == -1 {
		return 0, fmt.Errorf("invalid swap size format: %s (no unit found)", originalSizeStr)
	}

	// Extract numeric part and unit
	numericPart := sizeStr[:unitIndex]
	unit := sizeStr[unitIndex:]

	if numericPart == "" {
		return 0, fmt.Errorf("invalid swap size format: %s (no numeric value)", originalSizeStr)
	}

	// Parse numeric value
	value, err := strconv.ParseFloat(numericPart, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid swap size format: %s (invalid number: %w)", originalSizeStr, err)
	}

	// Convert to bytes based on unit
	var bytes int64
	switch unit {
	case "M":
		bytes = int64(value * 1024 * 1024)
	case "G":
		bytes = int64(value * 1024 * 1024 * 1024)
	case "T":
		bytes = int64(value * 1024 * 1024 * 1024 * 1024)
	default:
		return 0, fmt.Errorf("invalid swap size unit: %s (supported: M, G, T)", unit)
	}

	if bytes <= 0 {
		return 0, fmt.Errorf("swap size must be greater than 0")
	}

	return bytes, nil
}

// swapIsActive checks if swap is currently active on the system.
func swapIsActive() (bool, error) {
	// Try swapon --show first (preferred method)
	output, err := exec.RunWithOutput("swapon", "--show")
	if err == nil {
		// If command succeeds, check if output contains any swap entries
		// Empty output means no swap is active
		output = strings.TrimSpace(output)
		if output == "" {
			return false, nil
		}
		// Check if output contains swap file path or device name
		return strings.Contains(output, defaultSwapFilePath) || len(output) > 0, nil
	}

	// Fallback: read /proc/swaps
	if exec.FileExists("/proc/swaps") {
		content, err := os.ReadFile("/proc/swaps")
		if err != nil {
			return false, fmt.Errorf("failed to read /proc/swaps: %w", err)
		}
		// /proc/swaps has a header line, so if there's more than one line, swap is active
		lines := strings.Split(strings.TrimSpace(string(content)), "\n")
		// First line is header, so check if there are more lines
		return len(lines) > 1, nil
	}

	// If neither method works, assume no swap
	return false, nil
}

// swapFileExists checks if the swap file exists at the given path.
func swapFileExists(path string) bool {
	return exec.FileExists(path)
}

// fstabContainsSwap checks if /etc/fstab contains an entry for the swap file.
func fstabContainsSwap(swapPath string) (bool, error) {
	if !exec.FileExists(fstabPath) {
		return false, nil
	}

	file, err := os.Open(fstabPath)
	if err != nil {
		return false, fmt.Errorf("failed to open %s: %w", fstabPath, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split by whitespace
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		// Check if first field matches swap path and third field is "swap"
		if fields[0] == swapPath && len(fields) >= 3 && fields[2] == "swap" {
			return true, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("failed to read %s: %w", fstabPath, err)
	}

	return false, nil
}

// getSwappiness reads the current swappiness value from sysctl.
func getSwappiness() (int, error) {
	// Try reading from /proc/sys/vm/swappiness first (most reliable)
	if exec.FileExists("/proc/sys/vm/swappiness") {
		content, err := os.ReadFile("/proc/sys/vm/swappiness")
		if err != nil {
			return 0, fmt.Errorf("failed to read swappiness: %w", err)
		}
		value, err := strconv.Atoi(strings.TrimSpace(string(content)))
		if err != nil {
			return 0, fmt.Errorf("failed to parse swappiness: %w", err)
		}
		return value, nil
	}

	// Fallback: use sysctl command
	output, err := exec.RunWithOutput("sysctl", "-n", "vm.swappiness")
	if err != nil {
		return 0, fmt.Errorf("failed to get swappiness: %w", err)
	}

	value, err := strconv.Atoi(strings.TrimSpace(output))
	if err != nil {
		return 0, fmt.Errorf("failed to parse swappiness: %w", err)
	}

	return value, nil
}

// swappinessIsSet checks if swappiness is set to the expected value.
func swappinessIsSet(expectedValue int) (bool, error) {
	currentValue, err := getSwappiness()
	if err != nil {
		return false, err
	}
	return currentValue == expectedValue, nil
}

// IsInstalled checks if the swap module is already installed.
// Since IsInstalled() doesn't receive config, it performs generic checks.
// Install() performs specific checks with config and is fully idempotent.
func (m *SwapModule) IsInstalled() (bool, error) {
	// Check if swap is active
	active, err := swapIsActive()
	if err != nil {
		return false, fmt.Errorf("failed to check swap status: %w", err)
	}
	if !active {
		return false, nil
	}

	// Check if swap file exists
	if !swapFileExists(defaultSwapFilePath) {
		return false, nil
	}

	// Check if fstab contains swap entry
	fstabHasSwap, err := fstabContainsSwap(defaultSwapFilePath)
	if err != nil {
		return false, fmt.Errorf("failed to check fstab: %w", err)
	}
	if !fstabHasSwap {
		return false, nil
	}

	// Check if swappiness is set (check for default value)
	swappinessSet, err := swappinessIsSet(defaultSwappiness)
	if err != nil {
		return false, fmt.Errorf("failed to check swappiness: %w", err)
	}
	if !swappinessSet {
		return false, nil
	}

	return true, nil
}

// Install creates and configures the swap file.
func (m *SwapModule) Install(cfg *config.Config) error {
	dryRun := log.IsDryRun()

	// Check if swap is enabled
	if !cfg.Swap.Enabled {
		log.Skip("Swap is disabled in configuration")
		return nil
	}

	// Parse swap size
	swapSize := cfg.Swap.Size
	if swapSize == "" {
		swapSize = "2G" // Use default
	}

	sizeBytes, err := parseSwapSize(swapSize)
	if err != nil {
		return fmt.Errorf("failed to parse swap size: %w", err)
	}

	// Check if swap already exists
	swapActive, err := swapIsActive()
	if err != nil {
		return fmt.Errorf("failed to check swap status: %w", err)
	}

	// Create swap file if it doesn't exist
	if !swapActive || !swapFileExists(defaultSwapFilePath) {
		if dryRun {
			log.Info("Would create swap file of size %s (%d bytes)", swapSize, sizeBytes)
		} else {
			log.Info("Creating swap file of size %s", swapSize)

			// Try fallocate first (faster and more efficient)
			if exec.CommandExists("fallocate") {
				if err := exec.Run("fallocate", "-l", fmt.Sprintf("%d", sizeBytes), defaultSwapFilePath); err != nil {
					// Fallback to dd if fallocate fails
					log.Info("fallocate failed, using dd as fallback")
					// Calculate size in MB for dd
					sizeMB := sizeBytes / (1024 * 1024)
					if sizeMB == 0 {
						sizeMB = 1 // At least 1MB
					}
					if err := exec.Run("dd", "if=/dev/zero", fmt.Sprintf("of=%s", defaultSwapFilePath), "bs=1M", fmt.Sprintf("count=%d", sizeMB)); err != nil {
						return fmt.Errorf("failed to create swap file: %w", err)
					}
				}
			} else {
				// Use dd if fallocate is not available
				sizeMB := sizeBytes / (1024 * 1024)
				if sizeMB == 0 {
					sizeMB = 1
				}
				if err := exec.Run("dd", "if=/dev/zero", fmt.Sprintf("of=%s", defaultSwapFilePath), "bs=1M", fmt.Sprintf("count=%d", sizeMB)); err != nil {
					return fmt.Errorf("failed to create swap file: %w", err)
				}
			}

			// Set permissions
			log.Info("Setting swap file permissions")
			if err := exec.Run("chmod", "600", defaultSwapFilePath); err != nil {
				return fmt.Errorf("failed to set swap file permissions: %w", err)
			}

			// Format as swap
			log.Info("Formatting swap file")
			if err := exec.Run("mkswap", defaultSwapFilePath); err != nil {
				return fmt.Errorf("failed to format swap file: %w", err)
			}

			// Enable swap
			log.Info("Enabling swap")
			if err := exec.Run("swapon", defaultSwapFilePath); err != nil {
				return fmt.Errorf("failed to enable swap: %w", err)
			}

			log.Success("Swap file created and enabled")
		}
	} else {
		log.Skip("Swap file already exists and is active")
	}

	// Configure fstab
	fstabHasSwap, err := fstabContainsSwap(defaultSwapFilePath)
	if err != nil {
		return fmt.Errorf("failed to check fstab: %w", err)
	}

	if !fstabHasSwap {
		if dryRun {
			log.Info("Would add swap entry to %s", fstabPath)
		} else {
			log.Info("Adding swap entry to %s", fstabPath)

			// Read existing fstab
			var existingContent []byte
			if exec.FileExists(fstabPath) {
				existingContent, err = os.ReadFile(fstabPath)
				if err != nil {
					return fmt.Errorf("failed to read %s: %w", fstabPath, err)
				}
			}

			// Append swap entry
			swapEntry := fmt.Sprintf("%s none swap sw 0 0\n", defaultSwapFilePath)
			newContent := string(existingContent)
			if !strings.HasSuffix(newContent, "\n") && newContent != "" {
				newContent += "\n"
			}
			newContent += swapEntry

			// Write back to fstab
			if err := exec.WriteFile(fstabPath, []byte(newContent), 0644); err != nil {
				return fmt.Errorf("failed to write %s: %w", fstabPath, err)
			}

			log.Success("Swap entry added to %s", fstabPath)
		}
	} else {
		log.Skip("Swap entry already exists in %s", fstabPath)
	}

	// Set swappiness
	currentSwappiness, err := getSwappiness()
	if err != nil {
		return fmt.Errorf("failed to get current swappiness: %w", err)
	}

	if currentSwappiness != defaultSwappiness {
		if dryRun {
			log.Info("Would set swappiness to %d", defaultSwappiness)
		} else {
			log.Info("Setting swappiness to %d", defaultSwappiness)

			// Set runtime value
			if err := exec.Run("sysctl", fmt.Sprintf("vm.swappiness=%d", defaultSwappiness)); err != nil {
				return fmt.Errorf("failed to set swappiness: %w", err)
			}

			// Make persistent
			swappinessConfig := fmt.Sprintf("vm.swappiness=%d\n", defaultSwappiness)
			if err := exec.WriteFile(swappinessConfigPath, []byte(swappinessConfig), 0644); err != nil {
				return fmt.Errorf("failed to write swappiness config: %w", err)
			}

			log.Success("Swappiness set to %d", defaultSwappiness)
		}
	} else {
		log.Skip("Swappiness is already set to %d", defaultSwappiness)
	}

	if !dryRun {
		log.Success("Swap module installation completed successfully")
	}

	return nil
}

// Ensure SwapModule implements the Module interface
var _ module.Module = (*SwapModule)(nil)
