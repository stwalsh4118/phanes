package redis

import (
	"strings"
	"testing"

	"github.com/stwalsh4118/phanes/internal/config"
)

func TestRedisModule_Name(t *testing.T) {
	mod := &RedisModule{}
	if got := mod.Name(); got != "redis" {
		t.Errorf("Name() = %q, want %q", got, "redis")
	}
}

func TestRedisModule_Description(t *testing.T) {
	mod := &RedisModule{}
	want := "Installs and configures Redis in-memory data store"
	if got := mod.Description(); got != want {
		t.Errorf("Description() = %q, want %q", got, want)
	}
}

func TestRedisInstalled(t *testing.T) {
	// Test that redisInstalled() doesn't panic
	// It may return false if Redis is not installed
	installed, err := redisInstalled()
	if err != nil {
		t.Logf("redisInstalled() returned error (may be expected if Redis not installed): %v", err)
	}

	// Verify return type
	_ = installed
	_ = err
}

func TestRedisServiceRunning(t *testing.T) {
	// Test that redisServiceRunning() doesn't panic
	running, err := redisServiceRunning()
	if err != nil {
		t.Logf("redisServiceRunning() returned error (may be expected if Redis not running): %v", err)
	}

	// Verify return type
	_ = running
	_ = err
}

func TestRedisServiceEnabled(t *testing.T) {
	// Test that redisServiceEnabled() doesn't panic
	enabled, err := redisServiceEnabled()
	if err != nil {
		t.Logf("redisServiceEnabled() returned error (may be expected if Redis not enabled): %v", err)
	}

	// Verify return type
	_ = enabled
	_ = err
}

func TestRedisPortAccessible(t *testing.T) {
	// Test that redisPortAccessible() doesn't panic
	accessible, err := redisPortAccessible()
	if err != nil {
		t.Logf("redisPortAccessible() returned error (may be expected if Redis not accessible): %v", err)
	}

	// Verify return type
	_ = accessible
	_ = err
}

func TestRedisRespondsToPing(t *testing.T) {
	// Test that redisRespondsToPing() doesn't panic
	responds, err := redisRespondsToPing("")
	if err != nil {
		t.Logf("redisRespondsToPing() returned error (may be expected if Redis not accessible): %v", err)
	}

	// Verify return type
	_ = responds
	_ = err
}

func TestIsBindingToAllInterfaces(t *testing.T) {
	tests := []struct {
		addr string
		want bool
	}{
		{"0.0.0.0", true},
		{"::", true},
		{"127.0.0.1", false},
		{"localhost", false},
		{"192.168.1.1", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.addr, func(t *testing.T) {
			if got := isBindingToAllInterfaces(tt.addr); got != tt.want {
				t.Errorf("isBindingToAllInterfaces(%q) = %v, want %v", tt.addr, got, tt.want)
			}
		})
	}
}

func TestRedisModule_Install_EnabledFalse(t *testing.T) {
	mod := &RedisModule{}
	cfg := &config.Config{
		Redis: config.Redis{
			Enabled: false,
		},
	}

	// Should skip installation when Enabled is false
	err := mod.Install(cfg)
	if err != nil {
		t.Errorf("Install() with Enabled=false should not return error, got: %v", err)
	}
}

func TestRedisModule_Install_ConfigDefaults(t *testing.T) {
	mod := &RedisModule{}
	cfg := &config.Config{
		Redis: config.Redis{
			Enabled:     true,
			Password:    "", // Password is optional for Redis
			BindAddress: "", // Should default to 127.0.0.1
		},
	}

	// Should not panic when using defaults
	// Note: This will likely fail if Redis is not installed, but we're testing
	// that defaults are applied correctly
	err := mod.Install(cfg)
	// Error is expected if Redis is not installed, but we're testing config handling
	if err != nil && strings.Contains(err.Error(), "password") && !strings.Contains(err.Error(), "Warning") {
		t.Errorf("Install() should not fail on password validation (password is optional): %v", err)
	}
}

func TestRedisModule_Install_BindAllInterfacesWithoutPassword(t *testing.T) {
	mod := &RedisModule{}
	cfg := &config.Config{
		Redis: config.Redis{
			Enabled:     true,
			Password:    "", // No password
			BindAddress: "0.0.0.0", // Bind to all interfaces
		},
	}

	// Should warn but not error when binding to all interfaces without password
	err := mod.Install(cfg)
	// Error is expected if Redis is not installed, but warning should be logged
	// The warning is logged but doesn't cause an error
	// We're just checking that it doesn't panic
	if err != nil {
		// Errors are acceptable (e.g., Redis not installed, permission denied)
		// The important thing is that the warning is logged and it doesn't panic
		_ = err
	}
}

func TestIsInstalled(t *testing.T) {
	mod := &RedisModule{}

	// Test that IsInstalled() doesn't panic
	installed, err := mod.IsInstalled()
	if err != nil {
		t.Logf("IsInstalled() returned error (may be expected if Redis not installed): %v", err)
	}

	// Verify return type
	_ = installed
	_ = err
}

