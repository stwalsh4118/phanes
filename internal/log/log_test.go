package log

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

// setStdoutForTesting sets the stdout writer for testing purposes.
// This is a test helper that should only be used in tests.
func setStdoutForTesting(w io.Writer) {
	stdoutMu.Lock()
	defer stdoutMu.Unlock()
	stdout = w
}

// setStderrForTesting sets the stderr writer for testing purposes.
// This is a test helper that should only be used in tests.
func setStderrForTesting(w io.Writer) {
	stderrMu.Lock()
	defer stderrMu.Unlock()
	stderr = w
}

// resetWritersForTesting resets the writers to their original values.
func resetWritersForTesting() {
	stdoutMu.Lock()
	defer stdoutMu.Unlock()
	stderrMu.Lock()
	defer stderrMu.Unlock()
	stdout = os.Stdout
	stderr = os.Stderr
}

func TestSetDryRun(t *testing.T) {
	SetDryRun(true)
	if !isDryRun() {
		t.Error("Expected dry-run to be enabled")
	}

	SetDryRun(false)
	if isDryRun() {
		t.Error("Expected dry-run to be disabled")
	}
}

// TestSetDryRunNoDeadlock verifies that SetDryRun doesn't deadlock when called.
// This test would hang indefinitely if the deadlock bug exists.
func TestSetDryRunNoDeadlock(t *testing.T) {
	// This test verifies the fix for the deadlock where SetDryRun() holds a write lock
	// and calls updateLogger() which tries to acquire a read lock on the same mutex.
	// If the deadlock exists, this test will hang.
	done := make(chan bool)
	go func() {
		// Call SetDryRun multiple times to ensure no deadlock
		SetDryRun(true)
		SetDryRun(false)
		SetDryRun(true)
		SetDryRun(false)
		done <- true
	}()

	select {
	case <-done:
		// Success - no deadlock
	case <-time.After(2 * time.Second):
		t.Fatal("Deadlock detected: SetDryRun() hung indefinitely")
	}
}

func TestInfo(t *testing.T) {
	SetDryRun(false)
	var buf bytes.Buffer
	setStdoutForTesting(&buf)
	defer resetWritersForTesting()

	Info("test message")
	output := buf.String()

	if !strings.Contains(output, "test message") {
		t.Error("Expected output to contain 'test message'")
	}
	if strings.Contains(output, dryRunField) {
		t.Error("Expected output to not contain dry-run field")
	}
}

func TestSuccess(t *testing.T) {
	SetDryRun(false)
	var buf bytes.Buffer
	setStdoutForTesting(&buf)
	defer resetWritersForTesting()

	Success("operation completed")
	output := buf.String()

	if !strings.Contains(output, "operation completed") {
		t.Error("Expected output to contain 'operation completed'")
	}
	if !strings.Contains(output, "success") {
		t.Error("Expected output to contain success field")
	}
}

func TestError(t *testing.T) {
	SetDryRun(false)
	var buf bytes.Buffer
	setStderrForTesting(&buf)
	defer resetWritersForTesting()

	Error("something went wrong")
	output := buf.String()

	if !strings.Contains(output, "something went wrong") {
		t.Error("Expected output to contain 'something went wrong'")
	}
	// Zerolog console writer outputs "ERR" for error level, which becomes "err" when lowercased
	if !strings.Contains(strings.ToLower(output), "err") {
		t.Error("Expected output to contain error level (ERR)")
	}
}

func TestSkip(t *testing.T) {
	SetDryRun(false)
	var buf bytes.Buffer
	setStdoutForTesting(&buf)
	defer resetWritersForTesting()

	Skip("already installed")
	output := buf.String()

	if !strings.Contains(output, "already installed") {
		t.Error("Expected output to contain 'already installed'")
	}
	if !strings.Contains(output, "skip") {
		t.Error("Expected output to contain skip field")
	}
}

func TestWarn(t *testing.T) {
	SetDryRun(false)
	var buf bytes.Buffer
	setStdoutForTesting(&buf)
	defer resetWritersForTesting()

	Warn("this is a warning")
	output := buf.String()

	if !strings.Contains(output, "this is a warning") {
		t.Error("Expected output to contain 'this is a warning'")
	}
	if !strings.Contains(strings.ToLower(output), "warn") {
		t.Error("Expected output to contain warn level")
	}
}

func TestDryRunField(t *testing.T) {
	SetDryRun(true)
	var buf bytes.Buffer
	setStdoutForTesting(&buf)
	defer resetWritersForTesting()

	Info("test message")
	output := buf.String()

	if !strings.Contains(output, "test message") {
		t.Error("Expected output to contain message")
	}
	// Dry-run should be included as a field in zerolog output
	if !strings.Contains(strings.ToLower(output), "dry") {
		t.Error("Expected output to contain dry-run indication")
	}
}

func TestFormatString(t *testing.T) {
	SetDryRun(false)
	var buf bytes.Buffer
	setStdoutForTesting(&buf)
	defer resetWritersForTesting()

	Info("Installing %s version %s", "docker", "24.0")
	output := buf.String()

	if !strings.Contains(output, "docker") {
		t.Error("Expected output to contain 'docker'")
	}
	if !strings.Contains(output, "24.0") {
		t.Error("Expected output to contain '24.0'")
	}
}
