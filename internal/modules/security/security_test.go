package security

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stwalsh4118/phanes/internal/config"
	"github.com/stwalsh4118/phanes/internal/exec"
	"github.com/stwalsh4118/phanes/internal/log"
)

func TestSecurityModule_Name(t *testing.T) {
	mod := &SecurityModule{}
	if got := mod.Name(); got != "security" {
		t.Errorf("Name() = %q, want %q", got, "security")
	}
}

func TestSecurityModule_Description(t *testing.T) {
	mod := &SecurityModule{}
	want := "Configures UFW, fail2ban, and SSH hardening"
	if got := mod.Description(); got != want {
		t.Errorf("Description() = %q, want %q", got, want)
	}
}

func TestRenderTemplate(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     interface{}
		want     string
		wantErr  bool
	}{
		{
			name:     "simple template",
			template: "Hello {{.Name}}",
			data:     struct{ Name string }{Name: "World"},
			want:     "Hello World",
			wantErr:  false,
		},
		{
			name:     "SSH port template",
			template: "Port {{.SSHPort}}",
			data: struct {
				SSHPort int
			}{SSHPort: 2222},
			want:    "Port 2222",
			wantErr: false,
		},
		{
			name:     "boolean template",
			template: "PasswordAuth {{if .AllowPasswordAuth}}yes{{else}}no{{end}}",
			data: struct {
				AllowPasswordAuth bool
			}{AllowPasswordAuth: false},
			want:    "PasswordAuth no",
			wantErr: false,
		},
		{
			name:     "invalid template",
			template: "{{.InvalidField",
			data:     struct{}{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := renderTemplate(tt.template, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("renderTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("renderTemplate() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestUFWIsEnabled(t *testing.T) {
	// Test that ufwIsEnabled() doesn't panic
	// It may return false if UFW is not installed or not enabled
	enabled, err := ufwIsEnabled()
	if err != nil {
		t.Logf("ufwIsEnabled() returned error (may be expected if UFW not installed): %v", err)
	}

	// Verify return type
	_ = enabled
	_ = err

	// Note: Actual UFW status depends on system state, so we just verify
	// the function doesn't panic and handles errors gracefully
}

func TestFail2banIsRunning(t *testing.T) {
	// Test that fail2banIsRunning() doesn't panic
	// It may return false if fail2ban is not installed or not running
	running, err := fail2banIsRunning()
	if err != nil {
		t.Logf("fail2banIsRunning() returned error (may be expected if fail2ban not installed): %v", err)
	}

	// Verify return type
	_ = running
	_ = err

	// Note: Actual fail2ban status depends on system state, so we just verify
	// the function doesn't panic and handles errors gracefully
}

func TestSSHConfigMatches(t *testing.T) {
	tempDir := t.TempDir()
	sshdConfigPath := filepath.Join(tempDir, "sshd_config")

	cfg := &config.Config{
		Security: config.Security{
			SSHPort:           22,
			AllowPasswordAuth: false,
		},
	}

	// Test with non-existent file (uses hardcoded /etc/ssh/sshd_config path)
	// If the system file doesn't exist, it should return false, nil
	matches, err := sshConfigMatches(cfg)
	if err != nil {
		t.Logf("sshConfigMatches() returned error (may be expected if template rendering fails): %v", err)
	}
	// If file doesn't exist, matches should be false
	_ = matches

	// Create test SSH config file
	testConfig := `Port 22
PermitRootLogin no
PasswordAuthentication no
PubkeyAuthentication yes`
	if err := os.WriteFile(sshdConfigPath, []byte(testConfig), 0644); err != nil {
		t.Fatalf("Failed to create test SSH config: %v", err)
	}

	// Temporarily override the SSH config path for testing
	// Note: This is a limitation - sshConfigMatches uses hardcoded path
	// In a real scenario, we'd want to make the path configurable or use dependency injection
	// For now, we'll test the function with the actual system path if it exists
	originalPath := "/etc/ssh/sshd_config"
	if exec.FileExists(originalPath) {
		matches, err := sshConfigMatches(cfg)
		if err != nil {
			t.Logf("sshConfigMatches() returned error (may be expected): %v", err)
		}
		_ = matches
	}
}

func TestNormalizeSSHConfig(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "config with comments",
			input: "# Comment\nPort 22\n# Another comment\nPermitRootLogin no",
			want:  "Port 22\nPermitRootLogin no",
		},
		{
			name:  "config with empty lines",
			input: "Port 22\n\nPermitRootLogin no\n\n",
			want:  "Port 22\nPermitRootLogin no",
		},
		{
			name:  "config with whitespace",
			input: "  Port 22  \n  PermitRootLogin no  ",
			want:  "Port 22\nPermitRootLogin no",
		},
		{
			name:  "empty config",
			input: "",
			want:  "",
		},
		{
			name:  "only comments",
			input: "# Comment 1\n# Comment 2",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeSSHConfig(tt.input)
			if got != tt.want {
				t.Errorf("normalizeSSHConfig() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSecurityModule_IsInstalled(t *testing.T) {
	mod := &SecurityModule{}

	// Test that IsInstalled() doesn't panic and handles errors gracefully
	installed, err := mod.IsInstalled()
	if err != nil {
		t.Logf("IsInstalled() returned error (may be expected): %v", err)
	}

	// Verify return type
	_ = installed
	_ = err

	// Note: IsInstalled() checks system state, so actual results depend on
	// whether UFW, fail2ban, and SSH are configured. We just verify it doesn't panic.
}

func TestSecurityModule_Install_Validation(t *testing.T) {
	mod := &SecurityModule{}

	// Test with invalid SSH port (too low)
	cfg := &config.Config{
		Security: config.Security{
			SSHPort:           0,
			AllowPasswordAuth: false,
		},
	}

	err := mod.Install(cfg)
	if err == nil {
		t.Error("Install() should return error for invalid SSH port (0)")
	}
	if !strings.Contains(err.Error(), "invalid SSH port") {
		t.Errorf("Install() error should mention invalid SSH port, got: %v", err)
	}

	// Test with invalid SSH port (too high)
	cfg.Security.SSHPort = 65536
	err = mod.Install(cfg)
	if err == nil {
		t.Error("Install() should return error for invalid SSH port (65536)")
	}
	if !strings.Contains(err.Error(), "invalid SSH port") {
		t.Errorf("Install() error should mention invalid SSH port, got: %v", err)
	}

	// Test with valid SSH port
	cfg.Security.SSHPort = 2222
	err = mod.Install(cfg)
	// Error is expected since we're not running as root and can't actually configure
	// The important thing is it doesn't fail validation
	if err != nil && !strings.Contains(err.Error(), "invalid SSH port") {
		t.Logf("Install() returned error (expected for non-root): %v", err)
	}
}

func TestSecurityModule_Install_DryRun(t *testing.T) {
	mod := &SecurityModule{}

	// Enable dry-run mode
	originalDryRun := log.IsDryRun()
	log.SetDryRun(true)
	defer log.SetDryRun(originalDryRun)

	cfg := &config.Config{
		Security: config.Security{
			SSHPort:           22,
			AllowPasswordAuth: false,
		},
	}

	// In dry-run mode, Install() should log but not execute commands
	// We can't easily verify this without mocking, but we can verify
	// it doesn't panic and handles dry-run mode
	err := mod.Install(cfg)
	// In dry-run mode, it should not actually configure UFW/fail2ban/SSH
	// Error is acceptable if commands fail (since we're not running as root)
	if err != nil {
		// Error is acceptable in dry-run mode if system tools are not available
		// The important thing is it doesn't panic
		t.Logf("Install() in dry-run mode returned error (may be expected): %v", err)
	}
}

func TestSecurityModule_ModuleInterface(t *testing.T) {
	// Verify that SecurityModule implements the Module interface
	var _ interface {
		Name() string
		Description() string
		IsInstalled() (bool, error)
		Install(*config.Config) error
	} = &SecurityModule{}
}

func TestSecurityModule_Install_RequiresRoot(t *testing.T) {
	// Skip this test in non-root environments
	// Installing requires root privileges
	if os.Geteuid() != 0 {
		t.Skip("Skipping Install test - requires root privileges")
	}

	// Check if required commands exist
	if !exec.CommandExists("ufw") {
		t.Skip("ufw not available")
	}
	if !exec.CommandExists("fail2ban-server") {
		t.Skip("fail2ban-server not available")
	}
	if !exec.CommandExists("sshd") {
		t.Skip("sshd not available")
	}

	// This test would actually configure UFW, fail2ban, and SSH, so we skip it
	// In a real scenario, you'd want to:
	// 1. Backup current configurations
	// 2. Run Install()
	// 3. Verify UFW, fail2ban, and SSH are configured correctly
	// 4. Restore backups
	t.Skip("Skipping Install test - would modify system configuration")
}

func TestSecurityModule_Install_Idempotency(t *testing.T) {
	// Test that Install() is idempotent
	// This would require running Install() twice and verifying
	// the second run doesn't create duplicates or errors
	// We skip this test as it requires root and would modify system
	if os.Geteuid() != 0 {
		t.Skip("Skipping idempotency test - requires root privileges")
	}
	t.Skip("Skipping idempotency test - would modify system configuration")
}

func TestSecurityModule_TemplateRendering(t *testing.T) {
	// Test that SSH config template renders correctly
	templateData := struct {
		SSHPort           int
		AllowPasswordAuth bool
	}{
		SSHPort:           2222,
		AllowPasswordAuth: false,
	}

	rendered, err := renderTemplate(sshdConfigTemplate, templateData)
	if err != nil {
		t.Fatalf("Failed to render SSH config template: %v", err)
	}

	// Verify key settings are in rendered config
	if !strings.Contains(rendered, "Port 2222") {
		t.Error("Rendered SSH config should contain Port 2222")
	}
	if !strings.Contains(rendered, "PermitRootLogin no") {
		t.Error("Rendered SSH config should contain PermitRootLogin no")
	}
	if !strings.Contains(rendered, "PasswordAuthentication no") {
		t.Error("Rendered SSH config should contain PasswordAuthentication no")
	}
	if !strings.Contains(rendered, "PubkeyAuthentication yes") {
		t.Error("Rendered SSH config should contain PubkeyAuthentication yes")
	}
	if !strings.Contains(rendered, "UsePAM yes") {
		t.Error("Rendered SSH config should contain UsePAM yes (required for locked accounts)")
	}

	// Test with password auth enabled
	templateData.AllowPasswordAuth = true
	rendered, err = renderTemplate(sshdConfigTemplate, templateData)
	if err != nil {
		t.Fatalf("Failed to render SSH config template with password auth: %v", err)
	}
	if !strings.Contains(rendered, "PasswordAuthentication yes") {
		t.Error("Rendered SSH config should contain PasswordAuthentication yes when enabled")
	}

	// Test fail2ban template
	fail2banData := struct {
		SSHPort int
	}{
		SSHPort: 2222,
	}

	rendered, err = renderTemplate(jailLocalTemplate, fail2banData)
	if err != nil {
		t.Fatalf("Failed to render fail2ban template: %v", err)
	}

	// Verify key settings are in rendered config
	if !strings.Contains(rendered, "port = 2222") {
		t.Error("Rendered fail2ban config should contain port = 2222")
	}
	if !strings.Contains(rendered, "[sshd]") {
		t.Error("Rendered fail2ban config should contain [sshd] section")
	}
	if !strings.Contains(rendered, "enabled = true") {
		t.Error("Rendered fail2ban config should contain enabled = true")
	}
}
