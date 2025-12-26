package runner

import (
	"errors"
	"testing"

	"github.com/stwalsh4118/phanes/internal/config"
)

// mockModule is a test implementation of the Module interface.
type mockModule struct {
	name        string
	description string
	installed   bool
	installErr  error
	checkErr    error
}

func (m *mockModule) Name() string {
	return m.name
}

func (m *mockModule) Description() string {
	return m.description
}

func (m *mockModule) IsInstalled() (bool, error) {
	return m.installed, m.checkErr
}

func (m *mockModule) Install(cfg *config.Config) error {
	return m.installErr
}

func TestNewRunner(t *testing.T) {
	r := NewRunner()
	if r == nil {
		t.Fatal("NewRunner() returned nil")
	}
	if r.modules == nil {
		t.Fatal("Runner.modules is nil")
	}
	if len(r.modules) != 0 {
		t.Fatalf("Expected empty registry, got %d modules", len(r.modules))
	}
}

func TestRegisterModule(t *testing.T) {
	r := NewRunner()
	mod := &mockModule{
		name:        "test",
		description: "Test module",
	}

	r.RegisterModule(mod)

	if len(r.modules) != 1 {
		t.Fatalf("Expected 1 module in registry, got %d", len(r.modules))
	}

	registered := r.modules["test"]
	if registered == nil {
		t.Fatal("Module not found in registry")
	}

	if registered.Name() != "test" {
		t.Fatalf("Expected module name 'test', got '%s'", registered.Name())
	}
}

func TestRegisterModule_Duplicate(t *testing.T) {
	r := NewRunner()
	mod1 := &mockModule{
		name:        "test",
		description: "First module",
	}
	mod2 := &mockModule{
		name:        "test",
		description: "Second module",
	}

	r.RegisterModule(mod1)
	r.RegisterModule(mod2)

	if len(r.modules) != 1 {
		t.Fatalf("Expected 1 module in registry after duplicate registration, got %d", len(r.modules))
	}

	registered := r.modules["test"]
	if registered.Description() != "Second module" {
		t.Fatalf("Expected second module to overwrite first, got description '%s'", registered.Description())
	}
}

func TestRunModules_EmptyList(t *testing.T) {
	r := NewRunner()
	cfg := config.DefaultConfig()

	results, err := r.RunModules([]string{}, cfg, false)
	if err == nil {
		t.Fatal("Expected error for empty module list")
	}
	if results != nil {
		t.Fatal("Expected nil results for empty module list")
	}
}

func TestRunModules_UnknownModule(t *testing.T) {
	r := NewRunner()
	cfg := config.DefaultConfig()

	results, err := r.RunModules([]string{"unknown"}, cfg, false)
	if err == nil {
		t.Fatal("Expected error for unknown module")
	}
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	if results[0].Status != StatusError {
		t.Fatalf("Expected StatusError, got %s", results[0].Status)
	}
}

func TestRunModules_Success(t *testing.T) {
	r := NewRunner()
	mod := &mockModule{
		name:        "test",
		description: "Test module",
		installed:   false,
		installErr:  nil,
	}
	r.RegisterModule(mod)

	cfg := config.DefaultConfig()
	results, err := r.RunModules([]string{"test"}, cfg, false)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	if results[0].Status != StatusInstalled {
		t.Fatalf("Expected StatusInstalled, got %s", results[0].Status)
	}
	if results[0].Name != "test" {
		t.Fatalf("Expected name 'test', got %s", results[0].Name)
	}
}

func TestRunModules_SkipInstalled(t *testing.T) {
	r := NewRunner()
	mod := &mockModule{
		name:        "test",
		description: "Test module",
		installed:   true,
		installErr:  nil,
	}
	r.RegisterModule(mod)

	cfg := config.DefaultConfig()
	results, err := r.RunModules([]string{"test"}, cfg, false)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	if results[0].Status != StatusSkipped {
		t.Fatalf("Expected StatusSkipped, got %s", results[0].Status)
	}
}

func TestRunModules_InstallError(t *testing.T) {
	r := NewRunner()
	installErr := errors.New("installation failed")
	mod := &mockModule{
		name:        "test",
		description: "Test module",
		installed:   false,
		installErr:  installErr,
	}
	r.RegisterModule(mod)

	cfg := config.DefaultConfig()
	results, err := r.RunModules([]string{"test"}, cfg, false)
	if err == nil {
		t.Fatal("Expected error for failed installation")
	}
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	if results[0].Status != StatusFailed {
		t.Fatalf("Expected StatusFailed, got %s", results[0].Status)
	}
	if results[0].Error == nil {
		t.Fatal("Expected error to be set")
	}
}

func TestRunModules_IsInstalledError(t *testing.T) {
	r := NewRunner()
	checkErr := errors.New("check failed")
	mod := &mockModule{
		name:        "test",
		description: "Test module",
		installed:   false,
		checkErr:    checkErr,
	}
	r.RegisterModule(mod)

	cfg := config.DefaultConfig()
	results, err := r.RunModules([]string{"test"}, cfg, false)
	if err == nil {
		t.Fatal("Expected error for failed IsInstalled check")
	}
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	if results[0].Status != StatusError {
		t.Fatalf("Expected StatusError, got %s", results[0].Status)
	}
	if results[0].Error == nil {
		t.Fatal("Expected error to be set")
	}
}

func TestRunModules_DryRun(t *testing.T) {
	r := NewRunner()
	mod := &mockModule{
		name:        "test",
		description: "Test module",
		installed:   false,
		installErr:  nil,
	}
	r.RegisterModule(mod)

	cfg := config.DefaultConfig()
	results, err := r.RunModules([]string{"test"}, cfg, true)
	if err != nil {
		t.Fatalf("Expected no error in dry-run mode, got: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	if results[0].Status != StatusWouldInstall {
		t.Fatalf("Expected StatusWouldInstall in dry-run, got %s", results[0].Status)
	}

	// Verify Install was not called (installErr would be set if it was)
	// Since we can't easily verify this without more complex mocking,
	// we just verify that dry-run completes without error
}

func TestRunModules_DryRunInstalled(t *testing.T) {
	r := NewRunner()
	mod := &mockModule{
		name:        "test",
		description: "Test module",
		installed:   true,
		installErr:  nil,
	}
	r.RegisterModule(mod)

	cfg := config.DefaultConfig()
	results, err := r.RunModules([]string{"test"}, cfg, true)
	if err != nil {
		t.Fatalf("Expected no error in dry-run mode, got: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	if results[0].Status != StatusSkipped {
		t.Fatalf("Expected StatusSkipped in dry-run, got %s", results[0].Status)
	}
}

func TestRunModules_MultipleModules(t *testing.T) {
	r := NewRunner()
	mod1 := &mockModule{
		name:        "module1",
		description: "First module",
		installed:   false,
		installErr:  nil,
	}
	mod2 := &mockModule{
		name:        "module2",
		description: "Second module",
		installed:   false,
		installErr:  nil,
	}
	r.RegisterModule(mod1)
	r.RegisterModule(mod2)

	cfg := config.DefaultConfig()
	results, err := r.RunModules([]string{"module1", "module2"}, cfg, false)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}
	if results[0].Status != StatusInstalled || results[1].Status != StatusInstalled {
		t.Fatal("Expected both modules to be installed")
	}
}

func TestRunModules_MultipleModulesWithError(t *testing.T) {
	r := NewRunner()
	mod1 := &mockModule{
		name:        "module1",
		description: "First module",
		installed:   false,
		installErr:  nil,
	}
	mod2 := &mockModule{
		name:        "module2",
		description: "Second module",
		installed:   false,
		installErr:  errors.New("install failed"),
	}
	r.RegisterModule(mod1)
	r.RegisterModule(mod2)

	cfg := config.DefaultConfig()
	results, err := r.RunModules([]string{"module1", "module2"}, cfg, false)
	if err == nil {
		t.Fatal("Expected error when one module fails")
	}
	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}
	if results[0].Status != StatusInstalled {
		t.Fatalf("Expected module1 to be installed, got %s", results[0].Status)
	}
	if results[1].Status != StatusFailed {
		t.Fatalf("Expected module2 to be failed, got %s", results[1].Status)
	}
}

func TestRunModules_MixedInstalledAndNotInstalled(t *testing.T) {
	r := NewRunner()
	mod1 := &mockModule{
		name:        "installed",
		description: "Installed module",
		installed:   true,
		installErr:  nil,
	}
	mod2 := &mockModule{
		name:        "notinstalled",
		description: "Not installed module",
		installed:   false,
		installErr:  nil,
	}
	r.RegisterModule(mod1)
	r.RegisterModule(mod2)

	cfg := config.DefaultConfig()
	results, err := r.RunModules([]string{"installed", "notinstalled"}, cfg, false)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}
	// Order may vary, so check both
	skippedFound := false
	installedFound := false
	for _, result := range results {
		if result.Name == "installed" && result.Status == StatusSkipped {
			skippedFound = true
		}
		if result.Name == "notinstalled" && result.Status == StatusInstalled {
			installedFound = true
		}
	}
	if !skippedFound || !installedFound {
		t.Fatal("Expected one skipped and one installed module")
	}
}

func TestGetModule(t *testing.T) {
	r := NewRunner()
	mod := &mockModule{
		name:        "test",
		description: "Test module",
	}
	r.RegisterModule(mod)

	retrieved := r.GetModule("test")
	if retrieved == nil {
		t.Fatal("GetModule returned nil for registered module")
	}
	if retrieved.Name() != "test" {
		t.Fatalf("Expected module name 'test', got '%s'", retrieved.Name())
	}

	notFound := r.GetModule("nonexistent")
	if notFound != nil {
		t.Fatal("GetModule returned non-nil for unregistered module")
	}
}

func TestListModules(t *testing.T) {
	r := NewRunner()
	mod1 := &mockModule{name: "module1", description: "First"}
	mod2 := &mockModule{name: "module2", description: "Second"}
	mod3 := &mockModule{name: "module3", description: "Third"}

	r.RegisterModule(mod1)
	r.RegisterModule(mod2)
	r.RegisterModule(mod3)

	modules := r.ListModules()
	if len(modules) != 3 {
		t.Fatalf("Expected 3 modules, got %d", len(modules))
	}

	// Check that all modules are present (order may vary)
	moduleMap := make(map[string]bool)
	for _, name := range modules {
		moduleMap[name] = true
	}

	if !moduleMap["module1"] || !moduleMap["module2"] || !moduleMap["module3"] {
		t.Fatal("ListModules did not return all registered modules")
	}
}
