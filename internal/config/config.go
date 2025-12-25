package config

import (
	"fmt"
	"os"

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
	DevTools DevTools `yaml:"devtools"`
	Coolify  Coolify  `yaml:"coolify"`
}

// User contains user-related configuration.
type User struct {
	Username     string `yaml:"username"`
	SSHPublicKey string `yaml:"ssh_public_key"`
}

// System contains system-level configuration.
type System struct {
	Timezone string `yaml:"timezone"`
}

// Swap contains swap file configuration.
type Swap struct {
	Enabled bool   `yaml:"enabled"`
	Size    string `yaml:"size"`
}

// Security contains security-related configuration.
type Security struct {
	SSHPort           int  `yaml:"ssh_port"`
	AllowPasswordAuth bool `yaml:"allow_password_auth"`
}

// Docker contains Docker-related configuration.
type Docker struct {
	InstallCompose bool `yaml:"install_compose"`
}

// Postgres contains PostgreSQL configuration.
type Postgres struct {
	Enabled  bool   `yaml:"enabled"`
	Version  string `yaml:"version"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	User     string `yaml:"user"`
}

// Redis contains Redis configuration.
type Redis struct {
	Password string `yaml:"password"`
}

// Nginx contains Nginx configuration.
type Nginx struct {
	Enabled bool `yaml:"enabled"`
}

// Caddy contains Caddy configuration.
type Caddy struct {
	Enabled bool `yaml:"enabled"`
}

// DevTools contains development tools configuration.
type DevTools struct {
	NodeVersion   string `yaml:"node_version"`
	PythonVersion string `yaml:"python_version"`
	GoVersion     string `yaml:"go_version"`
}

// Coolify contains Coolify configuration.
type Coolify struct {
	Enabled bool `yaml:"enabled"`
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
			Password: "",
		},
		Nginx: Nginx{
			Enabled: true,
		},
		Caddy: Caddy{
			Enabled: true,
		},
		DevTools: DevTools{
			NodeVersion:   "20",
			PythonVersion: "3.12",
			GoVersion:     "1.25",
		},
		Coolify: Coolify{
			Enabled: true,
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

	return nil
}
