package postgres

import (
	"bufio"
	"fmt"
	"os"
	osexec "os/exec"
	"strings"

	"github.com/stwalsh4118/phanes/internal/config"
	"github.com/stwalsh4118/phanes/internal/exec"
	"github.com/stwalsh4118/phanes/internal/log"
	"github.com/stwalsh4118/phanes/internal/module"
)

const (
	postgresGPGKeyURL      = "https://www.postgresql.org/media/keys/ACCC4CF8.asc"
	postgresGPGKeyringPath = "/usr/share/keyrings/postgresql-archive-keyring.gpg"
	postgresAptSourcesPath = "/etc/apt/sources.list.d/pgdg.list"
	postgresRepoBaseURL    = "http://apt.postgresql.org/pub/repos/apt/"
	postgresServiceName    = "postgresql"
	postgresDefaultPort    = 5432
	defaultVersion         = "16"
	defaultDatabase        = "phanes"
	defaultUser            = "phanes"
)

// PostgresModule implements the Module interface for PostgreSQL installation.
type PostgresModule struct{}

// Name returns the unique name identifier for this module.
func (m *PostgresModule) Name() string {
	return "postgres"
}

// Description returns a human-readable description of what this module does.
func (m *PostgresModule) Description() string {
	return "Installs and configures PostgreSQL database server"
}

// getDistributionCodename gets the distribution codename (e.g., "jammy", "focal").
// Tries lsb_release first, then falls back to reading /etc/os-release.
func getDistributionCodename() (string, error) {
	// Try lsb_release first
	output, err := exec.RunWithOutput("lsb_release", "-cs")
	if err == nil {
		codename := strings.TrimSpace(output)
		if codename != "" {
			return codename, nil
		}
	}

	// Fallback to /etc/os-release
	if !exec.FileExists("/etc/os-release") {
		return "", fmt.Errorf("cannot determine distribution codename: lsb_release failed and /etc/os-release not found")
	}

	file, err := os.Open("/etc/os-release")
	if err != nil {
		return "", fmt.Errorf("failed to open /etc/os-release: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Look for VERSION_CODENAME or VERSION_ID
		if strings.HasPrefix(line, "VERSION_CODENAME=") {
			codename := strings.TrimPrefix(line, "VERSION_CODENAME=")
			codename = strings.Trim(codename, `"`)
			if codename != "" {
				return codename, nil
			}
		}
		// Fallback to VERSION_ID for Ubuntu versions
		if strings.HasPrefix(line, "VERSION_ID=") {
			versionID := strings.TrimPrefix(line, "VERSION_ID=")
			versionID = strings.Trim(versionID, `"`)
			// Map common Ubuntu version IDs to codenames
			versionMap := map[string]string{
				"22.04": "jammy",
				"20.04": "focal",
				"18.04": "bionic",
				"16.04": "xenial",
			}
			if codename, ok := versionMap[versionID]; ok {
				return codename, nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("failed to read /etc/os-release: %w", err)
	}

	return "", fmt.Errorf("cannot determine distribution codename from /etc/os-release")
}

// postgresInstalled checks if PostgreSQL is installed by running psql --version.
func postgresInstalled() (bool, error) {
	err := exec.Run("psql", "--version")
	if err != nil {
		return false, nil
	}
	return true, nil
}

// postgresServiceRunning checks if PostgreSQL service is running.
func postgresServiceRunning() (bool, error) {
	output, err := exec.RunWithOutput("systemctl", "is-active", postgresServiceName)
	if err != nil {
		return false, nil
	}
	return strings.TrimSpace(output) == "active", nil
}

// postgresServiceEnabled checks if PostgreSQL service is enabled.
func postgresServiceEnabled() (bool, error) {
	output, err := exec.RunWithOutput("systemctl", "is-enabled", postgresServiceName)
	if err != nil {
		return false, nil
	}
	status := strings.TrimSpace(output)
	return status == "enabled" || status == "enabled-runtime", nil
}

// postgresPortAccessible checks if PostgreSQL is listening on port 5432.
func postgresPortAccessible() (bool, error) {
	// Try ss first (more modern), fallback to netstat
	output, err := exec.RunWithOutput("ss", "-tlnp")
	if err != nil {
		// Fallback to netstat
		output, err = exec.RunWithOutput("netstat", "-tlnp")
		if err != nil {
			return false, fmt.Errorf("failed to check port accessibility: %w", err)
		}
	}

	// Check if port 5432 is in the output
	portStr := fmt.Sprintf(":%d", postgresDefaultPort)
	if strings.Contains(output, portStr) {
		return true, nil
	}

	return false, nil
}

// runPsqlCommand runs a psql command with optional PGPASSWORD environment variable.
func runPsqlCommand(password string, args ...string) error {
	cmd := osexec.Command("psql", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if password != "" {
		cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", password))
	}
	return cmd.Run()
}

// runPsqlWithOutput runs a psql command with output capture and optional PGPASSWORD.
func runPsqlWithOutput(password string, args ...string) (string, error) {
	cmd := osexec.Command("psql", args...)
	if password != "" {
		cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", password))
	}
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// databaseExists checks if a database exists.
func databaseExists(databaseName string) (bool, error) {
	output, err := runPsqlWithOutput("", "-U", "postgres", "-lqt")
	if err != nil {
		return false, fmt.Errorf("failed to list databases: %w", err)
	}

	// Parse output to check if database exists
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) > 0 && fields[0] == databaseName {
			return true, nil
		}
	}

	return false, nil
}

// userExists checks if a PostgreSQL user exists.
func userExists(userName string) (bool, error) {
	query := fmt.Sprintf("SELECT 1 FROM pg_roles WHERE rolname='%s'", userName)
	output, err := runPsqlWithOutput("", "-U", "postgres", "-tAc", query)
	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}

	output = strings.TrimSpace(output)
	return output == "1", nil
}

// createDatabase creates a PostgreSQL database.
func createDatabase(databaseName string) error {
	return runPsqlCommand("", "-U", "postgres", "-c", fmt.Sprintf("CREATE DATABASE %s;", databaseName))
}

// createUser creates a PostgreSQL user with a password.
func createUser(userName, password string) error {
	query := fmt.Sprintf("CREATE USER %s WITH PASSWORD '%s';", userName, password)
	return runPsqlCommand("", "-U", "postgres", "-c", query)
}

// grantPrivileges grants all privileges on a database to a user.
func grantPrivileges(databaseName, userName string) error {
	query := fmt.Sprintf("GRANT ALL PRIVILEGES ON DATABASE %s TO %s;", databaseName, userName)
	return runPsqlCommand("", "-U", "postgres", "-c", query)
}

// getPostgresConfigDir returns the PostgreSQL configuration directory for a version.
func getPostgresConfigDir(version string) string {
	return fmt.Sprintf("/etc/postgresql/%s/main/", version)
}

// IsInstalled checks if PostgreSQL is already installed and configured.
// Since IsInstalled() doesn't receive config, it performs generic checks.
// Install() performs specific checks with config and is fully idempotent.
func (m *PostgresModule) IsInstalled() (bool, error) {
	// Check if PostgreSQL is installed
	installed, err := postgresInstalled()
	if err != nil {
		return false, fmt.Errorf("failed to check PostgreSQL installation: %w", err)
	}
	if !installed {
		return false, nil
	}

	// Check if PostgreSQL service is running
	running, err := postgresServiceRunning()
	if err != nil {
		return false, fmt.Errorf("failed to check PostgreSQL service status: %w", err)
	}
	if !running {
		return false, nil
	}

	// Check if PostgreSQL port is accessible
	accessible, err := postgresPortAccessible()
	if err != nil {
		return false, fmt.Errorf("failed to check PostgreSQL port accessibility: %w", err)
	}
	if !accessible {
		return false, nil
	}

	return true, nil
}

// Install installs and configures PostgreSQL database server.
func (m *PostgresModule) Install(cfg *config.Config) error {
	dryRun := log.IsDryRun()

	// Check if PostgreSQL is enabled in config
	if !cfg.Postgres.Enabled {
		log.Skip("PostgreSQL module is disabled in configuration")
		return nil
	}

	// Validate password is not empty
	if cfg.Postgres.Password == "" {
		return fmt.Errorf("postgres.password is required but is empty")
	}

	// Apply defaults
	version := cfg.Postgres.Version
	if version == "" {
		version = defaultVersion
	}

	databaseName := cfg.Postgres.Database
	if databaseName == "" {
		databaseName = defaultDatabase
	}

	userName := cfg.Postgres.User
	if userName == "" {
		userName = defaultUser
	}

	// Check if PostgreSQL is already installed
	installed, err := postgresInstalled()
	if err != nil {
		return fmt.Errorf("failed to check PostgreSQL installation: %w", err)
	}

	if !installed {
		if dryRun {
			log.Info("Would install PostgreSQL %s", version)
		} else {
			log.Info("Installing PostgreSQL %s", version)

			// Install prerequisites
			log.Info("Installing prerequisites")
			if err := exec.Run("apt-get", "update"); err != nil {
				return fmt.Errorf("failed to update apt: %w", err)
			}
			if err := exec.Run("apt-get", "install", "-y", "wget", "ca-certificates"); err != nil {
				return fmt.Errorf("failed to install prerequisites: %w", err)
			}

			// Download and add PostgreSQL GPG key
			log.Info("Adding PostgreSQL GPG key")
			cmd := fmt.Sprintf("wget --quiet -O - %s | gpg --dearmor -o %s", postgresGPGKeyURL, postgresGPGKeyringPath)
			if err := exec.Run("bash", "-c", cmd); err != nil {
				return fmt.Errorf("failed to add PostgreSQL GPG key: %w", err)
			}

			// Get distribution codename
			codename, err := getDistributionCodename()
			if err != nil {
				return fmt.Errorf("failed to get distribution codename: %w", err)
			}

			// Add PostgreSQL repository
			log.Info("Adding PostgreSQL repository")
			repoLine := fmt.Sprintf("deb [signed-by=%s] %s %s-pgdg main\n", postgresGPGKeyringPath, postgresRepoBaseURL, codename)
			if err := exec.WriteFile(postgresAptSourcesPath, []byte(repoLine), 0644); err != nil {
				return fmt.Errorf("failed to add PostgreSQL repository: %w", err)
			}

			// Update apt package list
			log.Info("Updating apt package list")
			if err := exec.Run("apt-get", "update"); err != nil {
				return fmt.Errorf("failed to update apt after adding PostgreSQL repository: %w", err)
			}

			// Install PostgreSQL
			log.Info("Installing PostgreSQL %s", version)
			packageName := fmt.Sprintf("postgresql-%s", version)
			if err := exec.Run("apt-get", "install", "-y", packageName); err != nil {
				return fmt.Errorf("failed to install PostgreSQL: %w", err)
			}

			// Verify PostgreSQL installation
			log.Info("Verifying PostgreSQL installation")
			if err := exec.Run("psql", "--version"); err != nil {
				return fmt.Errorf("PostgreSQL installation verification failed: %w", err)
			}

			log.Success("PostgreSQL %s installed successfully", version)
		}
	} else {
		log.Skip("PostgreSQL is already installed")
	}

	// Configure service to start on boot
	enabled, err := postgresServiceEnabled()
	if err != nil {
		return fmt.Errorf("failed to check if PostgreSQL service is enabled: %w", err)
	}

	if !enabled {
		if dryRun {
			log.Info("Would enable PostgreSQL service to start on boot")
		} else {
			log.Info("Enabling PostgreSQL service to start on boot")
			if err := exec.Run("systemctl", "enable", postgresServiceName); err != nil {
				return fmt.Errorf("failed to enable PostgreSQL service: %w", err)
			}
			log.Success("PostgreSQL service enabled")
		}
	} else {
		log.Skip("PostgreSQL service is already enabled")
	}

	// Start service if not running
	running, err := postgresServiceRunning()
	if err != nil {
		return fmt.Errorf("failed to check PostgreSQL service status: %w", err)
	}

	if !running {
		if dryRun {
			log.Info("Would start PostgreSQL service")
		} else {
			log.Info("Starting PostgreSQL service")
			if err := exec.Run("systemctl", "start", postgresServiceName); err != nil {
				return fmt.Errorf("failed to start PostgreSQL service: %w", err)
			}

			// Verify service is running
			running, err := postgresServiceRunning()
			if err != nil {
				return fmt.Errorf("failed to verify PostgreSQL service status: %w", err)
			}
			if !running {
				return fmt.Errorf("PostgreSQL service is not running after start")
			}

			log.Success("PostgreSQL service started")
		}
	} else {
		log.Skip("PostgreSQL service is already running")
	}

	// Create database if it doesn't exist
	dbExists, err := databaseExists(databaseName)
	if err != nil {
		return fmt.Errorf("failed to check if database exists: %w", err)
	}

	if !dbExists {
		if dryRun {
			log.Info("Would create database %s", databaseName)
		} else {
			log.Info("Creating database %s", databaseName)
			if err := createDatabase(databaseName); err != nil {
				return fmt.Errorf("failed to create database: %w", err)
			}
			log.Success("Database %s created", databaseName)
		}
	} else {
		log.Skip("Database %s already exists", databaseName)
	}

	// Create user if it doesn't exist
	usrExists, err := userExists(userName)
	if err != nil {
		return fmt.Errorf("failed to check if user exists: %w", err)
	}

	if !usrExists {
		if dryRun {
			log.Info("Would create user %s", userName)
		} else {
			log.Info("Creating user %s", userName)
			if err := createUser(userName, cfg.Postgres.Password); err != nil {
				return fmt.Errorf("failed to create user: %w", err)
			}
			log.Success("User %s created", userName)
		}
	} else {
		log.Skip("User %s already exists", userName)
	}

	// Grant privileges (idempotent - safe to run multiple times)
	if dryRun {
		log.Info("Would grant privileges on database %s to user %s", databaseName, userName)
	} else {
		log.Info("Granting privileges on database %s to user %s", databaseName, userName)
		if err := grantPrivileges(databaseName, userName); err != nil {
			return fmt.Errorf("failed to grant privileges: %w", err)
		}
		log.Success("Privileges granted")
	}

	// Verify PostgreSQL is accessible
	accessible, err := postgresPortAccessible()
	if err != nil {
		return fmt.Errorf("failed to verify PostgreSQL port accessibility: %w", err)
	}

	if !accessible {
		if dryRun {
			log.Info("Would verify PostgreSQL is accessible on port %d", postgresDefaultPort)
		} else {
			log.Warn("PostgreSQL port %d is not yet accessible. The service may still be starting.", postgresDefaultPort)
		}
	} else {
		if !dryRun {
			log.Success("PostgreSQL is accessible on port %d", postgresDefaultPort)
			log.Info("PostgreSQL connection details:")
			log.Info("  Host: localhost")
			log.Info("  Port: %d", postgresDefaultPort)
			log.Info("  Database: %s", databaseName)
			log.Info("  User: %s", userName)
			log.Info("  Password: [configured]")
		}
	}

	if !dryRun {
		log.Success("PostgreSQL module installation completed successfully")
	}

	return nil
}

// Ensure PostgresModule implements the Module interface
var _ module.Module = (*PostgresModule)(nil)

