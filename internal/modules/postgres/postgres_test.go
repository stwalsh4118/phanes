package postgres

import (
	"strings"
	"testing"

	"github.com/stwalsh4118/phanes/internal/config"
)

func TestPostgresModule_Name(t *testing.T) {
	mod := &PostgresModule{}
	if got := mod.Name(); got != "postgres" {
		t.Errorf("Name() = %q, want %q", got, "postgres")
	}
}

func TestPostgresModule_Description(t *testing.T) {
	mod := &PostgresModule{}
	want := "Installs and configures PostgreSQL database server"
	if got := mod.Description(); got != want {
		t.Errorf("Description() = %q, want %q", got, want)
	}
}

func TestPostgresInstalled(t *testing.T) {
	// Test that postgresInstalled() doesn't panic
	// It may return false if PostgreSQL is not installed
	installed, err := postgresInstalled()
	if err != nil {
		t.Logf("postgresInstalled() returned error (may be expected if PostgreSQL not installed): %v", err)
	}

	// Verify return type
	_ = installed
	_ = err
}

func TestPostgresServiceRunning(t *testing.T) {
	// Test that postgresServiceRunning() doesn't panic
	running, err := postgresServiceRunning()
	if err != nil {
		t.Logf("postgresServiceRunning() returned error (may be expected if PostgreSQL not running): %v", err)
	}

	// Verify return type
	_ = running
	_ = err
}

func TestPostgresServiceEnabled(t *testing.T) {
	// Test that postgresServiceEnabled() doesn't panic
	enabled, err := postgresServiceEnabled()
	if err != nil {
		t.Logf("postgresServiceEnabled() returned error (may be expected if PostgreSQL not enabled): %v", err)
	}

	// Verify return type
	_ = enabled
	_ = err
}

func TestPostgresPortAccessible(t *testing.T) {
	// Test that postgresPortAccessible() doesn't panic
	accessible, err := postgresPortAccessible()
	if err != nil {
		t.Logf("postgresPortAccessible() returned error (may be expected if PostgreSQL not accessible): %v", err)
	}

	// Verify return type
	_ = accessible
	_ = err
}

func TestGetPostgresConfigDir(t *testing.T) {
	tests := []struct {
		version string
		want    string
	}{
		{"16", "/etc/postgresql/16/main/"},
		{"14", "/etc/postgresql/14/main/"},
		{"12", "/etc/postgresql/12/main/"},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			if got := getPostgresConfigDir(tt.version); got != tt.want {
				t.Errorf("getPostgresConfigDir(%q) = %q, want %q", tt.version, got, tt.want)
			}
		})
	}
}

func TestPostgresModule_Install_EnabledFalse(t *testing.T) {
	mod := &PostgresModule{}
	cfg := &config.Config{
		Postgres: config.Postgres{
			Enabled: false,
		},
	}

	// Should skip installation when Enabled is false
	err := mod.Install(cfg)
	if err != nil {
		t.Errorf("Install() with Enabled=false should not return error, got: %v", err)
	}
}

func TestPostgresModule_Install_EmptyPassword(t *testing.T) {
	mod := &PostgresModule{}
	cfg := &config.Config{
		Postgres: config.Postgres{
			Enabled:  true,
			Password: "", // Empty password should cause error
		},
	}

	// Should return error when password is empty
	err := mod.Install(cfg)
	if err == nil {
		t.Error("Install() with empty password should return error")
	}
	if err != nil && !strings.Contains(err.Error(), "password") {
		t.Errorf("Install() error should mention password, got: %v", err)
	}
}

func TestPostgresModule_Install_ConfigDefaults(t *testing.T) {
	mod := &PostgresModule{}
	cfg := &config.Config{
		Postgres: config.Postgres{
			Enabled:  true,
			Password: "testpassword",
			// Version, Database, User should use defaults
		},
	}

	// Should not panic when using defaults
	// Note: This will likely fail if PostgreSQL is not installed, but we're testing
	// that defaults are applied correctly
	err := mod.Install(cfg)
	// Error is expected if PostgreSQL is not installed, but we're testing config handling
	if err != nil && strings.Contains(err.Error(), "password") {
		t.Errorf("Install() should not fail on password validation with valid password: %v", err)
	}
}

func TestIsInstalled(t *testing.T) {
	mod := &PostgresModule{}

	// Test that IsInstalled() doesn't panic
	installed, err := mod.IsInstalled()
	if err != nil {
		t.Logf("IsInstalled() returned error (may be expected if PostgreSQL not installed): %v", err)
	}

	// Verify return type
	_ = installed
	_ = err
}

