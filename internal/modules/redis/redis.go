package redis

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/stwalsh4118/phanes/internal/config"
	"github.com/stwalsh4118/phanes/internal/exec"
	"github.com/stwalsh4118/phanes/internal/log"
	"github.com/stwalsh4118/phanes/internal/module"
)

const (
	redisServiceName   = "redis-server"
	redisDefaultPort    = 6379
	redisConfigPath    = "/etc/redis/redis.conf"
	redisPackageName   = "redis-server"
	defaultBindAddress  = "127.0.0.1"
)

// RedisModule implements the Module interface for Redis installation.
type RedisModule struct{}

// Name returns the unique name identifier for this module.
func (m *RedisModule) Name() string {
	return "redis"
}

// Description returns a human-readable description of what this module does.
func (m *RedisModule) Description() string {
	return "Installs and configures Redis in-memory data store"
}

// redisInstalled checks if Redis is installed by running redis-cli --version.
func redisInstalled() (bool, error) {
	err := exec.Run("redis-cli", "--version")
	if err != nil {
		return false, nil
	}
	return true, nil
}

// redisServiceRunning checks if Redis service is running.
func redisServiceRunning() (bool, error) {
	output, err := exec.RunWithOutput("systemctl", "is-active", redisServiceName)
	if err != nil {
		return false, nil
	}
	return strings.TrimSpace(output) == "active", nil
}

// redisServiceEnabled checks if Redis service is enabled.
func redisServiceEnabled() (bool, error) {
	output, err := exec.RunWithOutput("systemctl", "is-enabled", redisServiceName)
	if err != nil {
		return false, nil
	}
	status := strings.TrimSpace(output)
	return status == "enabled" || status == "enabled-runtime", nil
}

// redisPortAccessible checks if Redis is listening on port 6379.
func redisPortAccessible() (bool, error) {
	// Try ss first (more modern), fallback to netstat
	output, err := exec.RunWithOutput("ss", "-tlnp")
	if err != nil {
		// Fallback to netstat
		output, err = exec.RunWithOutput("netstat", "-tlnp")
		if err != nil {
			return false, fmt.Errorf("failed to check port accessibility: %w", err)
		}
	}

	// Check if port 6379 is in the output
	portStr := fmt.Sprintf(":%d", redisDefaultPort)
	if strings.Contains(output, portStr) {
		return true, nil
	}

	return false, nil
}

// redisRespondsToPing checks if Redis responds to ping command.
func redisRespondsToPing(password string) (bool, error) {
	var args []string
	if password != "" {
		args = []string{"-a", password, "ping"}
	} else {
		args = []string{"ping"}
	}

	output, err := exec.RunWithOutput("redis-cli", args...)
	if err != nil {
		return false, nil
	}

	// Check if output contains "PONG"
	return strings.Contains(strings.ToUpper(output), "PONG"), nil
}

// isBindingToAllInterfaces checks if bind address is 0.0.0.0 or ::.
func isBindingToAllInterfaces(bindAddress string) bool {
	return bindAddress == "0.0.0.0" || bindAddress == "::"
}

// getRedisConfigValue reads a value from Redis config file.
func getRedisConfigValue(key string) (string, bool, error) {
	if !exec.FileExists(redisConfigPath) {
		return "", false, nil
	}

	file, err := os.Open(redisConfigPath)
	if err != nil {
		return "", false, fmt.Errorf("failed to open Redis config file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip comments and empty lines
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		// Check if line starts with the key
		if strings.HasPrefix(line, key) {
			// Extract value (key value or key "value")
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				value := parts[1]
				// Remove quotes if present
				value = strings.Trim(value, `"`)
				return value, true, nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", false, fmt.Errorf("failed to read Redis config file: %w", err)
	}

	return "", false, nil
}

// updateRedisConfig updates or adds a key-value pair in Redis config file.
// If value is empty, the line is commented out instead of being removed.
func updateRedisConfig(key, value string) error {
	if !exec.FileExists(redisConfigPath) {
		return fmt.Errorf("Redis config file does not exist: %s", redisConfigPath)
	}

	// Read current config
	file, err := os.Open(redisConfigPath)
	if err != nil {
		return fmt.Errorf("failed to open Redis config file: %w", err)
	}

	var lines []string
	keyFound := false
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Check if this line contains the key (commented or not)
		if strings.HasPrefix(trimmed, key) || strings.HasPrefix(trimmed, "#"+key) {
			keyFound = true
			if value == "" {
				// Comment out the line if value is empty
				if !strings.HasPrefix(trimmed, "#") {
					lines = append(lines, "# "+line)
				} else {
					// Already commented, keep as is
					lines = append(lines, line)
				}
			} else {
				// Update the line with new value
				lines = append(lines, fmt.Sprintf("%s %s", key, value))
			}
		} else {
			lines = append(lines, line)
		}
	}
	file.Close()

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read Redis config file: %w", err)
	}

	// If key not found and value is not empty, add it at the end
	if !keyFound && value != "" {
		lines = append(lines, fmt.Sprintf("%s %s", key, value))
	}

	// Write updated config back
	content := strings.Join(lines, "\n") + "\n"
	if err := exec.WriteFile(redisConfigPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write Redis config file: %w", err)
	}

	return nil
}

// configureRedisBind updates the bind address in Redis config.
func configureRedisBind(bindAddress string) error {
	return updateRedisConfig("bind", bindAddress)
}

// configureRedisPassword updates or removes the password in Redis config.
// If password is empty, the requirepass line is commented out.
func configureRedisPassword(password string) error {
	return updateRedisConfig("requirepass", password)
}

// reloadRedisConfig reloads or restarts Redis to apply config changes.
func reloadRedisConfig() error {
	// Try reload first
	err := exec.Run("systemctl", "reload", redisServiceName)
	if err != nil {
		// If reload fails, restart
		log.Info("Reload failed, restarting Redis service")
		if err := exec.Run("systemctl", "restart", redisServiceName); err != nil {
			return fmt.Errorf("failed to restart Redis service: %w", err)
		}
	}

	// Verify service is still running
	running, err := redisServiceRunning()
	if err != nil {
		return fmt.Errorf("failed to verify Redis service status: %w", err)
	}
	if !running {
		return fmt.Errorf("Redis service is not running after reload/restart")
	}

	return nil
}

// IsInstalled checks if Redis is already installed and configured.
// Since IsInstalled() doesn't receive config, it performs generic checks.
// Install() performs specific checks with config and is fully idempotent.
func (m *RedisModule) IsInstalled() (bool, error) {
	// Check if Redis is installed
	installed, err := redisInstalled()
	if err != nil {
		return false, fmt.Errorf("failed to check Redis installation: %w", err)
	}
	if !installed {
		return false, nil
	}

	// Check if Redis service is running
	running, err := redisServiceRunning()
	if err != nil {
		return false, fmt.Errorf("failed to check Redis service status: %w", err)
	}
	if !running {
		return false, nil
	}

	// Check if Redis port is accessible
	accessible, err := redisPortAccessible()
	if err != nil {
		return false, fmt.Errorf("failed to check Redis port accessibility: %w", err)
	}
	if !accessible {
		return false, nil
	}

	// Check if Redis responds to ping (without password for IsInstalled check)
	responds, err := redisRespondsToPing("")
	if err != nil {
		return false, fmt.Errorf("failed to check Redis ping response: %w", err)
	}
	if !responds {
		return false, nil
	}

	return true, nil
}

// Install installs and configures Redis in-memory data store.
func (m *RedisModule) Install(cfg *config.Config) error {
	dryRun := log.IsDryRun()

	// Check if Redis is enabled in config
	if !cfg.Redis.Enabled {
		log.Skip("Redis module is disabled in configuration")
		return nil
	}

	// Apply defaults
	bindAddress := cfg.Redis.BindAddress
	if bindAddress == "" {
		bindAddress = defaultBindAddress
	}

	password := cfg.Redis.Password

	// Warn if binding to all interfaces without password
	if isBindingToAllInterfaces(bindAddress) && password == "" {
		log.Warn("Warning: Redis is configured to bind to all interfaces (0.0.0.0 or ::) without a password. This is insecure.")
	}

	// Check if Redis is already installed
	installed, err := redisInstalled()
	if err != nil {
		return fmt.Errorf("failed to check Redis installation: %w", err)
	}

	if !installed {
		if dryRun {
			log.Info("Would install Redis")
		} else {
			log.Info("Installing Redis")

			// Update apt package list
			log.Info("Updating apt package list")
			if err := exec.Run("apt-get", "update"); err != nil {
				return fmt.Errorf("failed to update apt: %w", err)
			}

			// Install Redis
			log.Info("Installing Redis package")
			if err := exec.Run("apt-get", "install", "-y", redisPackageName); err != nil {
				return fmt.Errorf("failed to install Redis: %w", err)
			}

			// Verify Redis installation
			log.Info("Verifying Redis installation")
			if err := exec.Run("redis-cli", "--version"); err != nil {
				return fmt.Errorf("Redis installation verification failed: %w", err)
			}

			log.Success("Redis installed successfully")
		}
	} else {
		log.Skip("Redis is already installed")
	}

	// Configure Redis bind address
	currentBind, found, err := getRedisConfigValue("bind")
	if err != nil {
		return fmt.Errorf("failed to read Redis bind address: %w", err)
	}

	if !found || currentBind != bindAddress {
		if dryRun {
			log.Info("Would configure Redis bind address to %s", bindAddress)
		} else {
			log.Info("Configuring Redis bind address to %s", bindAddress)
			if err := configureRedisBind(bindAddress); err != nil {
				return fmt.Errorf("failed to configure Redis bind address: %w", err)
			}
			log.Success("Redis bind address configured")
		}
	} else {
		log.Skip("Redis bind address already configured")
	}

	// Configure Redis password
	currentPassword, found, err := getRedisConfigValue("requirepass")
	if err != nil {
		return fmt.Errorf("failed to read Redis password configuration: %w", err)
	}

	// Check if password needs to be updated
	passwordNeedsUpdate := false
	if password == "" {
		// If password is empty, check if requirepass is set
		if found && currentPassword != "" {
			passwordNeedsUpdate = true
		}
	} else {
		// If password is set, check if it's different
		if !found || currentPassword != password {
			passwordNeedsUpdate = true
		}
	}

	if passwordNeedsUpdate {
		if dryRun {
			if password == "" {
				log.Info("Would remove Redis password")
			} else {
				log.Info("Would configure Redis password")
			}
		} else {
			if password == "" {
				log.Info("Removing Redis password")
			} else {
				log.Info("Configuring Redis password")
			}
			if err := configureRedisPassword(password); err != nil {
				return fmt.Errorf("failed to configure Redis password: %w", err)
			}
			log.Success("Redis password configured")
		}
	} else {
		log.Skip("Redis password already configured")
	}

	// Reload Redis configuration if we made changes
	if !dryRun && (passwordNeedsUpdate || (!found || currentBind != bindAddress)) {
		log.Info("Reloading Redis configuration")
		if err := reloadRedisConfig(); err != nil {
			return fmt.Errorf("failed to reload Redis configuration: %w", err)
		}
		log.Success("Redis configuration reloaded")
	}

	// Configure service to start on boot
	enabled, err := redisServiceEnabled()
	if err != nil {
		return fmt.Errorf("failed to check if Redis service is enabled: %w", err)
	}

	if !enabled {
		if dryRun {
			log.Info("Would enable Redis service to start on boot")
		} else {
			log.Info("Enabling Redis service to start on boot")
			if err := exec.Run("systemctl", "enable", redisServiceName); err != nil {
				return fmt.Errorf("failed to enable Redis service: %w", err)
			}
			log.Success("Redis service enabled")
		}
	} else {
		log.Skip("Redis service is already enabled")
	}

	// Start service if not running
	running, err := redisServiceRunning()
	if err != nil {
		return fmt.Errorf("failed to check Redis service status: %w", err)
	}

	if !running {
		if dryRun {
			log.Info("Would start Redis service")
		} else {
			log.Info("Starting Redis service")
			if err := exec.Run("systemctl", "start", redisServiceName); err != nil {
				return fmt.Errorf("failed to start Redis service: %w", err)
			}

			// Verify service is running
			running, err := redisServiceRunning()
			if err != nil {
				return fmt.Errorf("failed to verify Redis service status: %w", err)
			}
			if !running {
				return fmt.Errorf("Redis service is not running after start")
			}

			log.Success("Redis service started")
		}
	} else {
		log.Skip("Redis service is already running")
	}

	// Verify Redis is accessible
	accessible, err := redisPortAccessible()
	if err != nil {
		return fmt.Errorf("failed to verify Redis port accessibility: %w", err)
	}

	if !accessible {
		if dryRun {
			log.Info("Would verify Redis is accessible on port %d", redisDefaultPort)
		} else {
			log.Warn("Redis port %d is not yet accessible. The service may still be starting.", redisDefaultPort)
		}
	} else {
		// Test ping with password if configured
		responds, err := redisRespondsToPing(password)
		if err != nil {
			return fmt.Errorf("failed to test Redis ping: %w", err)
		}

		if !dryRun {
			if responds {
				log.Success("Redis is accessible on port %d", redisDefaultPort)
				log.Info("Redis connection details:")
				log.Info("  Host: %s", bindAddress)
				log.Info("  Port: %d", redisDefaultPort)
				if password != "" {
					log.Info("  Password: [configured]")
				} else {
					log.Info("  Password: [not set]")
				}
			} else {
				log.Warn("Redis port is accessible but ping test failed. Check password configuration.")
			}
		}
	}

	if !dryRun {
		log.Success("Redis module installation completed successfully")
	}

	return nil
}

// Ensure RedisModule implements the Module interface
var _ module.Module = (*RedisModule)(nil)

