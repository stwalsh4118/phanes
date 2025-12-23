package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.User.Username != "" {
		t.Errorf("Expected empty username, got %q", cfg.User.Username)
	}
	if cfg.User.SSHPublicKey != "" {
		t.Errorf("Expected empty SSH public key, got %q", cfg.User.SSHPublicKey)
	}
	if cfg.System.Timezone != "UTC" {
		t.Errorf("Expected timezone UTC, got %q", cfg.System.Timezone)
	}
	if !cfg.Swap.Enabled {
		t.Error("Expected swap enabled by default")
	}
	if cfg.Swap.Size != "2G" {
		t.Errorf("Expected swap size 2G, got %q", cfg.Swap.Size)
	}
	if cfg.Security.SSHPort != 22 {
		t.Errorf("Expected SSH port 22, got %d", cfg.Security.SSHPort)
	}
	if cfg.Security.AllowPasswordAuth {
		t.Error("Expected password auth disabled by default")
	}
	if !cfg.Docker.InstallCompose {
		t.Error("Expected Docker Compose installation enabled by default")
	}
	if cfg.Postgres.Version != "16" {
		t.Errorf("Expected Postgres version 16, got %q", cfg.Postgres.Version)
	}
	if cfg.DevTools.NodeVersion != "20" {
		t.Errorf("Expected Node version 20, got %q", cfg.DevTools.NodeVersion)
	}
	if cfg.DevTools.PythonVersion != "3.12" {
		t.Errorf("Expected Python version 3.12, got %q", cfg.DevTools.PythonVersion)
	}
	if cfg.DevTools.GoVersion != "1.25" {
		t.Errorf("Expected Go version 1.25, got %q", cfg.DevTools.GoVersion)
	}
	if cfg.Nginx.Enabled {
		t.Error("Expected Nginx disabled by default")
	}
	if cfg.Caddy.Enabled {
		t.Error("Expected Caddy disabled by default")
	}
	if cfg.Coolify.Enabled {
		t.Error("Expected Coolify disabled by default")
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil config",
			cfg:     nil,
			wantErr: true,
			errMsg:  "config is nil",
		},
		{
			name: "missing username",
			cfg: &Config{
				User: User{
					Username:     "",
					SSHPublicKey: "ssh-ed25519 AAAA...",
				},
			},
			wantErr: true,
			errMsg:  "user.username is required",
		},
		{
			name: "missing SSH public key",
			cfg: &Config{
				User: User{
					Username:     "deploy",
					SSHPublicKey: "",
				},
			},
			wantErr: true,
			errMsg:  "user.ssh_public_key is required",
		},
		{
			name: "valid config",
			cfg: &Config{
				User: User{
					Username:     "deploy",
					SSHPublicKey: "ssh-ed25519 AAAA...",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && err.Error() != tt.errMsg {
				t.Errorf("Validate() error = %q, want %q", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestLoad(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		yaml      string
		wantErr   bool
		checkFunc func(*Config) error
	}{
		{
			name: "valid minimal config",
			yaml: `user:
  username: deploy
  ssh_public_key: "ssh-ed25519 AAAA... test@host"
`,
			wantErr: false,
			checkFunc: func(cfg *Config) error {
				if cfg.User.Username != "deploy" {
					return fmt.Errorf("expected username 'deploy', got %q", cfg.User.Username)
				}
				if cfg.User.SSHPublicKey != "ssh-ed25519 AAAA... test@host" {
					return fmt.Errorf("expected SSH key, got %q", cfg.User.SSHPublicKey)
				}
				// Check defaults are applied
				if cfg.System.Timezone != "UTC" {
					return fmt.Errorf("expected default timezone UTC, got %q", cfg.System.Timezone)
				}
				return nil
			},
		},
		{
			name: "valid full config",
			yaml: `user:
  username: deploy
  ssh_public_key: "ssh-ed25519 AAAA... test@host"
system:
  timezone: America/New_York
swap:
  enabled: true
  size: 4G
security:
  ssh_port: 2222
  allow_password_auth: false
docker:
  install_compose: true
postgres:
  version: "15"
  password: "secret123"
redis:
  password: "redis-secret"
nginx:
  enabled: true
caddy:
  enabled: false
devtools:
  node_version: "18"
  python_version: "3.11"
  go_version: "1.24"
coolify:
  enabled: true
`,
			wantErr: false,
			checkFunc: func(cfg *Config) error {
				if cfg.User.Username != "deploy" {
					return fmt.Errorf("expected username 'deploy', got %q", cfg.User.Username)
				}
				if cfg.System.Timezone != "America/New_York" {
					return fmt.Errorf("expected timezone 'America/New_York', got %q", cfg.System.Timezone)
				}
				if cfg.Swap.Size != "4G" {
					return fmt.Errorf("expected swap size '4G', got %q", cfg.Swap.Size)
				}
				if cfg.Security.SSHPort != 2222 {
					return fmt.Errorf("expected SSH port 2222, got %d", cfg.Security.SSHPort)
				}
				if cfg.Postgres.Version != "15" {
					return fmt.Errorf("expected Postgres version '15', got %q", cfg.Postgres.Version)
				}
				if cfg.Postgres.Password != "secret123" {
					return fmt.Errorf("expected Postgres password 'secret123', got %q", cfg.Postgres.Password)
				}
				if cfg.Redis.Password != "redis-secret" {
					return fmt.Errorf("expected Redis password 'redis-secret', got %q", cfg.Redis.Password)
				}
				if !cfg.Nginx.Enabled {
					return fmt.Errorf("expected Nginx enabled")
				}
				if cfg.Caddy.Enabled {
					return fmt.Errorf("expected Caddy disabled")
				}
				if cfg.DevTools.NodeVersion != "18" {
					return fmt.Errorf("expected Node version '18', got %q", cfg.DevTools.NodeVersion)
				}
				if !cfg.Coolify.Enabled {
					return fmt.Errorf("expected Coolify enabled")
				}
				return nil
			},
		},
		{
			name: "config with defaults applied",
			yaml: `user:
  username: deploy
  ssh_public_key: "ssh-ed25519 AAAA... test@host"
`,
			wantErr: false,
			checkFunc: func(cfg *Config) error {
				// Verify defaults are applied for fields not in YAML
				if cfg.Swap.Enabled != true {
					return fmt.Errorf("expected default swap enabled")
				}
				if cfg.Swap.Size != "2G" {
					return fmt.Errorf("expected default swap size '2G', got %q", cfg.Swap.Size)
				}
				if cfg.Security.SSHPort != 22 {
					return fmt.Errorf("expected default SSH port 22, got %d", cfg.Security.SSHPort)
				}
				if cfg.Docker.InstallCompose != true {
					return fmt.Errorf("expected default Docker Compose enabled")
				}
				return nil
			},
		},
		{
			name: "missing required username",
			yaml: `user:
  ssh_public_key: "ssh-ed25519 AAAA... test@host"
`,
			wantErr: true,
		},
		{
			name: "missing required SSH public key",
			yaml: `user:
  username: deploy
`,
			wantErr: true,
		},
		{
			name:    "invalid YAML",
			yaml:    `user: [invalid`,
			wantErr: true,
		},
		{
			name:    "empty file",
			yaml:    ``,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configFile := filepath.Join(tmpDir, "config.yaml")
			if err := os.WriteFile(configFile, []byte(tt.yaml), 0644); err != nil {
				t.Fatalf("Failed to write test config: %v", err)
			}

			cfg, err := Load(configFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if cfg == nil {
					t.Fatal("Load() returned nil config without error")
				}
				if tt.checkFunc != nil {
					if err := tt.checkFunc(cfg); err != nil {
						t.Errorf("Config check failed: %v", err)
					}
				}
			}
		})
	}
}

func TestLoadNonExistentFile(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("Load() expected error for non-existent file, got nil")
	}
}

func TestLoadPartialOverride(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	yaml := `user:
  username: deploy
  ssh_public_key: "ssh-ed25519 AAAA... test@host"
swap:
  enabled: false
devtools:
  node_version: "22"
`

	if err := os.WriteFile(configFile, []byte(yaml), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := Load(configFile)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Check overridden values
	if cfg.Swap.Enabled {
		t.Error("Expected swap disabled, got enabled")
	}
	if cfg.DevTools.NodeVersion != "22" {
		t.Errorf("Expected Node version '22', got %q", cfg.DevTools.NodeVersion)
	}

	// Check defaults are still applied for other fields
	if cfg.Swap.Size != "2G" {
		t.Errorf("Expected default swap size '2G', got %q", cfg.Swap.Size)
	}
	if cfg.DevTools.PythonVersion != "3.12" {
		t.Errorf("Expected default Python version '3.12', got %q", cfg.DevTools.PythonVersion)
	}
}
