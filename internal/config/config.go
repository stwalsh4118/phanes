package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the complete configuration structure for Phanes.
type Config struct {
	User     User     `yaml:"user"`
	System   System   `yaml:"system"`
	Swap     Swap     `yaml:"swap"`
	Security Security `yaml:"security"`
	Docker   Docker   `yaml:"docker"`
	Postgres Postgres `yaml:"postgres"`
	Redis    Redis    `yaml:"redis"`
	Nginx    Nginx    `yaml:"nginx"`
	Caddy    Caddy    `yaml:"caddy"`
	DevTools  DevTools  `yaml:"devtools"`
	Coolify   Coolify   `yaml:"coolify"`
	Tailscale Tailscale `yaml:"tailscale"`
}

// User contains user-related configuration.
// Both fields are required for the user module to function.
type User struct {
	// Username is the Linux username to create on the server.
	Username string `yaml:"username"`
	// SSHPublicKey is the SSH public key to add to the user's authorized_keys file.
	SSHPublicKey string `yaml:"ssh_public_key"`
}

// System contains system-level configuration.
type System struct {
	// Timezone is the system timezone (e.g., "UTC", "America/New_York").
	Timezone string `yaml:"timezone"`
}

// Swap contains swap file configuration.
type Swap struct {
	// Enabled determines whether to create a swap file.
	Enabled bool `yaml:"enabled"`
	// Size is the swap file size (e.g., "1G", "2G", "4G").
	Size string `yaml:"size"`
}

// Security contains security-related configuration.
type Security struct {
	// SSHPort is the port number for SSH (1-65535).
	SSHPort int `yaml:"ssh_port"`
	// AllowPasswordAuth enables password authentication for SSH (not recommended).
	AllowPasswordAuth bool `yaml:"allow_password_auth"`
}

// Docker contains Docker-related configuration.
type Docker struct {
	// InstallCompose determines whether to install Docker Compose alongside Docker CE.
	InstallCompose bool `yaml:"install_compose"`
}

// Postgres contains PostgreSQL configuration.
type Postgres struct {
	// Enabled determines whether to install PostgreSQL.
	Enabled bool `yaml:"enabled"`
	// Version is the PostgreSQL version to install (e.g., "16", "15").
	Version string `yaml:"version"`
	// Password is the PostgreSQL superuser password (empty for no password).
	Password string `yaml:"password"`
	// Database is the initial database name to create.
	Database string `yaml:"database"`
	// User is the PostgreSQL user name.
	User string `yaml:"user"`
}

// Redis contains Redis configuration.
type Redis struct {
	// Enabled determines whether to install Redis.
	Enabled bool `yaml:"enabled"`
	// Password is the Redis password (empty for no password).
	Password string `yaml:"password"`
	// BindAddress is the IP address Redis should bind to (e.g., "127.0.0.1", "0.0.0.0").
	BindAddress string `yaml:"bind_address"`
}

// Nginx contains Nginx configuration.
type Nginx struct {
	// Enabled determines whether to install Nginx.
	Enabled bool `yaml:"enabled"`
}

// Caddy contains Caddy configuration.
type Caddy struct {
	// Enabled determines whether to install Caddy.
	Enabled bool `yaml:"enabled"`
}

// DevTools contains development tools configuration.
type DevTools struct {
	// Enabled determines whether to install development tools.
	Enabled bool `yaml:"enabled"`
	// NodeVersion is the Node.js version to install via nvm (e.g., "22", "20").
	NodeVersion string `yaml:"node_version"`
	// PythonVersion is the Python version to install (e.g., "3", "3.11").
	PythonVersion string `yaml:"python_version"`
	// GoVersion is the Go version to install (e.g., "1.25.5").
	GoVersion string `yaml:"go_version"`
	// InstallUv determines whether to install the uv package manager for Python.
	InstallUv bool `yaml:"install_uv"`
}

// Coolify contains Coolify configuration.
type Coolify struct {
	// Enabled determines whether to install Coolify (requires Docker).
	Enabled bool `yaml:"enabled"`
}

// Tailscale contains Tailscale VPN configuration.
type Tailscale struct {
	// Enabled determines whether to install and configure Tailscale.
	Enabled bool `yaml:"enabled"`
	// AuthKey is the Tailscale auth key for authentication (must start with "tskey-").
	AuthKey string `yaml:"auth_key"`
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		User: User{
			Username:     "",
			SSHPublicKey: "",
		},
		System: System{
			Timezone: "UTC",
		},
		Swap: Swap{
			Enabled: true,
			Size:    "2G",
		},
		Security: Security{
			SSHPort:           22,
			AllowPasswordAuth: false,
		},
		Docker: Docker{
			InstallCompose: true,
		},
		Postgres: Postgres{
			Enabled:  true,
			Version:  "16",
			Password: "",
			Database: "phanes",
			User:     "phanes",
		},
		Redis: Redis{
			Enabled:     true,
			Password:    "",
			BindAddress: "127.0.0.1",
		},
		Nginx: Nginx{
			Enabled: true,
		},
		Caddy: Caddy{
			Enabled: true,
		},
		DevTools: DevTools{
			Enabled:       true,
			NodeVersion:   "22",
			PythonVersion: "3",
			GoVersion:     "1.25.5",
			InstallUv:     true,
		},
		Coolify: Coolify{
			Enabled: true,
		},
		Tailscale: Tailscale{
			Enabled: false,
			AuthKey: "",
		},
	}
}

// Load reads and parses a YAML configuration file, applies defaults, and validates it.
// Returns the parsed Config and an error if loading, parsing, or validation fails.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	if err := Validate(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// Validate checks that all required fields in the Config are set.
// Returns an error if any required field is missing or empty.
func Validate(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}

	if cfg.User.Username == "" {
		return fmt.Errorf("user.username is required")
	}

	if cfg.User.SSHPublicKey == "" {
		return fmt.Errorf("user.ssh_public_key is required")
	}

	if cfg.Tailscale.Enabled {
		if cfg.Tailscale.AuthKey == "" {
			return fmt.Errorf("tailscale.auth_key is required when tailscale is enabled")
		}
		if !strings.HasPrefix(cfg.Tailscale.AuthKey, "tskey-") {
			return fmt.Errorf("tailscale.auth_key must start with 'tskey-'")
		}
	}

	return nil
}
