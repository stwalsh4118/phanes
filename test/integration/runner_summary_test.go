package integration

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/stwalsh4118/phanes/internal/config"
	"github.com/stwalsh4118/phanes/internal/runner"
)

// mockModuleForSummary is a test implementation of the Module interface for summary testing.
type mockModuleForSummary struct {
	name        string
	description string
	installed   bool
	installErr  error
	checkErr    error
}

func (m *mockModuleForSummary) Name() string {
	return m.name
}

func (m *mockModuleForSummary) Description() string {
	return m.description
}

func (m *mockModuleForSummary) IsInstalled() (bool, error) {
	return m.installed, m.checkErr
}

func (m *mockModuleForSummary) Install(cfg *config.Config) error {
	return m.installErr
}

// TestSummary_AllStatuses verifies that PrintSummary displays all status types correctly.
func TestSummary_AllStatuses(t *testing.T) {
	results := []runner.ModuleResult{
		{Name: "baseline", Status: runner.StatusInstalled},
		{Name: "user", Status: runner.StatusSkipped},
		{Name: "docker", Status: runner.StatusFailed, Error: errors.New("Permission denied")},
		{Name: "nginx", Status: runner.StatusError, Error: errors.New("Module check failed")},
	}

	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	runner.PrintSummary(results, false)

	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	// Verify table structure
	if !strings.Contains(output, "Module") {
		t.Error("Summary table should contain 'Module' header")
	}
	if !strings.Contains(output, "Status") {
		t.Error("Summary table should contain 'Status' header")
	}
	if !strings.Contains(output, "Details") {
		t.Error("Summary table should contain 'Details' header")
	}

	// Verify module names appear
	if !strings.Contains(output, "baseline") {
		t.Error("Summary should contain 'baseline' module")
	}
	if !strings.Contains(output, "user") {
		t.Error("Summary should contain 'user' module")
	}
	if !strings.Contains(output, "docker") {
		t.Error("Summary should contain 'docker' module")
	}
	if !strings.Contains(output, "nginx") {
		t.Error("Summary should contain 'nginx' module")
	}

	// Verify status strings appear
	if !strings.Contains(output, "Installed") {
		t.Error("Summary should contain 'Installed' status")
	}
	if !strings.Contains(output, "Skipped") {
		t.Error("Summary should contain 'Skipped' status")
	}
	if !strings.Contains(output, "Failed") {
		t.Error("Summary should contain 'Failed' status")
	}
	if !strings.Contains(output, "Error") {
		t.Error("Summary should contain 'Error' status")
	}

	// Verify error messages appear
	if !strings.Contains(output, "Permission denied") {
		t.Error("Summary should contain error message for failed module")
	}
	if !strings.Contains(output, "Module check failed") {
		t.Error("Summary should contain error message for error module")
	}

	// Verify color codes are present (ANSI escape sequences)
	if !strings.Contains(output, "\033[32m") {
		t.Error("Summary should contain green color code for installed")
	}
	if !strings.Contains(output, "\033[33m") {
		t.Error("Summary should contain yellow color code for skipped")
	}
	if !strings.Contains(output, "\033[31m") {
		t.Error("Summary should contain red color code for failed/error")
	}

	// Verify totals
	if !strings.Contains(output, "1 installed") {
		t.Error("Summary should show 1 installed")
	}
	if !strings.Contains(output, "1 skipped") {
		t.Error("Summary should show 1 skipped")
	}
	if !strings.Contains(output, "2 failed") {
		t.Error("Summary should show 2 failed (1 failed + 1 error)")
	}
}

// TestSummary_Totals verifies that summary totals are calculated correctly.
func TestSummary_Totals(t *testing.T) {
	tests := []struct {
		name           string
		results        []runner.ModuleResult
		expectedTotals []string
	}{
		{
			name: "all installed",
			results: []runner.ModuleResult{
				{Name: "module1", Status: runner.StatusInstalled},
				{Name: "module2", Status: runner.StatusInstalled},
				{Name: "module3", Status: runner.StatusInstalled},
			},
			expectedTotals: []string{"3 installed"},
		},
		{
			name: "all skipped",
			results: []runner.ModuleResult{
				{Name: "module1", Status: runner.StatusSkipped},
				{Name: "module2", Status: runner.StatusSkipped},
			},
			expectedTotals: []string{"2 skipped"},
		},
		{
			name: "all failed",
			results: []runner.ModuleResult{
				{Name: "module1", Status: runner.StatusFailed},
				{Name: "module2", Status: runner.StatusError},
			},
			expectedTotals: []string{"2 failed"},
		},
		{
			name: "mixed results",
			results: []runner.ModuleResult{
				{Name: "module1", Status: runner.StatusInstalled},
				{Name: "module2", Status: runner.StatusInstalled},
				{Name: "module3", Status: runner.StatusSkipped},
				{Name: "module4", Status: runner.StatusFailed},
			},
			expectedTotals: []string{"2 installed", "1 skipped", "1 failed"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture output
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			runner.PrintSummary(tt.results, false)

			w.Close()
			os.Stdout = oldStdout

			// Read captured output
			buf := make([]byte, 4096)
			n, _ := r.Read(buf)
			output := string(buf[:n])

			// Verify all expected totals appear
			for _, expected := range tt.expectedTotals {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected summary to contain '%s', got output:\n%s", expected, output)
				}
			}
		})
	}
}

// TestSummary_DryRun verifies that dry-run mode is indicated in the summary.
func TestSummary_DryRun(t *testing.T) {
	results := []runner.ModuleResult{
		{Name: "baseline", Status: runner.StatusWouldInstall},
		{Name: "user", Status: runner.StatusSkipped},
	}

	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	runner.PrintSummary(results, true)

	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	// Verify dry-run indicator
	if !strings.Contains(output, "(dry-run)") {
		t.Error("Summary should indicate dry-run mode")
	}
	// Verify "would install" status appears
	if !strings.Contains(output, "Would Install") {
		t.Error("Summary should show 'Would Install' status for dry-run modules")
	}
	// Verify "would install" count appears
	if !strings.Contains(output, "would install") {
		t.Error("Summary should show 'would install' count")
	}
}

// TestSummary_EmptyResults verifies that empty results don't cause issues.
func TestSummary_EmptyResults(t *testing.T) {
	results := []runner.ModuleResult{}

	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	runner.PrintSummary(results, false)

	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	// Should not output anything for empty results
	if len(output) > 0 && !strings.Contains(output, "No modules processed") {
		// If there's output, it should be minimal
		if strings.Contains(output, "Module") || strings.Contains(output, "Status") {
			t.Error("Empty results should not print table headers")
		}
	}
}

// TestSummary_FullFlow verifies the complete flow from RunModules to PrintSummary.
// This tests acceptance criteria: "Runner tracks module execution results for all modules"
func TestSummary_FullFlow(t *testing.T) {
	r := runner.NewRunner()

	// Register modules with different outcomes
	r.RegisterModule(&mockModuleForSummary{
		name:        "installed",
		description: "Will be installed",
		installed:   false,
		installErr:  nil,
	})
	r.RegisterModule(&mockModuleForSummary{
		name:        "skipped",
		description: "Will be skipped",
		installed:   true,
		installErr:  nil,
	})
	r.RegisterModule(&mockModuleForSummary{
		name:        "failed",
		description: "Will fail",
		installed:   false,
		installErr:  errors.New("installation failed"),
	})

	cfg := config.DefaultConfig()
	results, err := r.RunModules([]string{"installed", "skipped", "failed"}, cfg, false)

	// Should have error due to failed module
	if err == nil {
		t.Error("Expected error when module fails")
	}

	// Should have 3 results
	if len(results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(results))
	}

	// Verify results are correct
	foundInstalled := false
	foundSkipped := false
	foundFailed := false

	for _, result := range results {
		switch result.Name {
		case "installed":
			if result.Status != runner.StatusInstalled {
				t.Errorf("Expected 'installed' module to have StatusInstalled, got %s", result.Status)
			}
			foundInstalled = true
		case "skipped":
			if result.Status != runner.StatusSkipped {
				t.Errorf("Expected 'skipped' module to have StatusSkipped, got %s", result.Status)
			}
			foundSkipped = true
		case "failed":
			if result.Status != runner.StatusFailed {
				t.Errorf("Expected 'failed' module to have StatusFailed, got %s", result.Status)
			}
			if result.Error == nil {
				t.Error("Expected 'failed' module to have error set")
			}
			foundFailed = true
		}
	}

	if !foundInstalled || !foundSkipped || !foundFailed {
		t.Error("Not all expected results were found")
	}

	// Verify summary can be printed with these results
	oldStdout := os.Stdout
	outputR, outputW, _ := os.Pipe()
	os.Stdout = outputW

	runner.PrintSummary(results, false)

	outputW.Close()
	os.Stdout = oldStdout

	buf := make([]byte, 4096)
	n, _ := outputR.Read(buf)
	output := string(buf[:n])

	// Verify summary contains all modules
	if !strings.Contains(output, "installed") {
		t.Error("Summary should contain 'installed' module")
	}
	if !strings.Contains(output, "skipped") {
		t.Error("Summary should contain 'skipped' module")
	}
	if !strings.Contains(output, "failed") {
		t.Error("Summary should contain 'failed' module")
	}
}

// TestSummary_DryRunFlow verifies the complete flow in dry-run mode.
func TestSummary_DryRunFlow(t *testing.T) {
	r := runner.NewRunner()

	// Register modules
	r.RegisterModule(&mockModuleForSummary{
		name:        "installed",
		description: "Will be installed",
		installed:   false,
		installErr:  nil,
	})
	r.RegisterModule(&mockModuleForSummary{
		name:        "skipped",
		description: "Will be skipped",
		installed:   true,
		installErr:  nil,
	})

	cfg := config.DefaultConfig()
	results, err := r.RunModules([]string{"installed", "skipped"}, cfg, true)

	// Should not have error in dry-run
	if err != nil {
		t.Fatalf("Unexpected error in dry-run mode: %v", err)
	}

	// Should have 2 results
	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}

	// Verify results are correct for dry-run
	for _, result := range results {
		if result.Name == "installed" {
			if result.Status != runner.StatusWouldInstall {
				t.Errorf("Expected 'installed' module to have StatusWouldInstall in dry-run, got %s", result.Status)
			}
		}
		if result.Name == "skipped" {
			if result.Status != runner.StatusSkipped {
				t.Errorf("Expected 'skipped' module to have StatusSkipped in dry-run, got %s", result.Status)
			}
		}
	}

	// Verify summary shows dry-run indicator
	oldStdout := os.Stdout
	outputR, outputW, _ := os.Pipe()
	os.Stdout = outputW

	runner.PrintSummary(results, true)

	outputW.Close()
	os.Stdout = oldStdout

	buf := make([]byte, 4096)
	n, _ := outputR.Read(buf)
	output := string(buf[:n])

	if !strings.Contains(output, "(dry-run)") {
		t.Error("Summary should indicate dry-run mode")
	}
}
