package log

import (
	"io"
	"os"
	"sync"

	"github.com/rs/zerolog"
)

const (
	dryRunField = "dry_run"
)

var (
	dryRun   bool
	mu       sync.RWMutex
	stdout   io.Writer = os.Stdout
	stderr   io.Writer = os.Stderr
	stdoutMu sync.RWMutex
	stderrMu sync.RWMutex
)

// SetDryRun sets the dry-run mode flag. When enabled, all log messages
// will include a dry_run field set to true.
func SetDryRun(enabled bool) {
	mu.Lock()
	defer mu.Unlock()
	dryRun = enabled
}

// IsDryRun returns the current dry-run mode state in a thread-safe manner.
func IsDryRun() bool {
	mu.RLock()
	defer mu.RUnlock()
	return dryRun
}

// isDryRun returns the current dry-run mode state in a thread-safe manner.
func isDryRun() bool {
	mu.RLock()
	defer mu.RUnlock()
	return dryRun
}

// getStdout returns the current stdout writer in a thread-safe manner.
func getStdout() io.Writer {
	stdoutMu.RLock()
	defer stdoutMu.RUnlock()
	return stdout
}

// getStderr returns the current stderr writer in a thread-safe manner.
func getStderr() io.Writer {
	stderrMu.RLock()
	defer stderrMu.RUnlock()
	return stderr
}

// Info logs an informational message to stdout.
func Info(format string, args ...interface{}) {
	// Create logger with current settings (thread-safe)
	output := zerolog.ConsoleWriter{
		Out:        getStdout(),
		NoColor:    false,
		TimeFormat: "15:04:05",
	}
	log := zerolog.New(output).With().Timestamp().Logger()

	mu.RLock()
	if dryRun {
		log = log.With().Bool(dryRunField, true).Logger()
	}
	mu.RUnlock()

	log.Info().Msgf(format, args...)
}

// Success logs a success message to stdout at info level with a success field.
func Success(format string, args ...interface{}) {
	// Create logger with current settings (thread-safe)
	output := zerolog.ConsoleWriter{
		Out:        getStdout(),
		NoColor:    false,
		TimeFormat: "15:04:05",
	}
	log := zerolog.New(output).With().Timestamp().Logger()

	mu.RLock()
	if dryRun {
		log = log.With().Bool(dryRunField, true).Logger()
	}
	mu.RUnlock()

	log.Info().Bool("success", true).Msgf(format, args...)
}

// Error logs an error message to stderr.
func Error(format string, args ...interface{}) {
	// Create separate logger for errors that writes to stderr
	output := zerolog.ConsoleWriter{
		Out:        getStderr(),
		NoColor:    false,
		TimeFormat: "15:04:05",
	}
	errLogger := zerolog.New(output).With().Timestamp().Logger()

	mu.RLock()
	if dryRun {
		errLogger = errLogger.With().Bool(dryRunField, true).Logger()
	}
	mu.RUnlock()

	errLogger.Error().Msgf(format, args...)
}

// Skip logs a skip message to stdout at info level with a skip field.
func Skip(format string, args ...interface{}) {
	// Create logger with current settings (thread-safe)
	output := zerolog.ConsoleWriter{
		Out:        getStdout(),
		NoColor:    false,
		TimeFormat: "15:04:05",
	}
	log := zerolog.New(output).With().Timestamp().Logger()

	mu.RLock()
	if dryRun {
		log = log.With().Bool(dryRunField, true).Logger()
	}
	mu.RUnlock()

	log.Info().Bool("skip", true).Msgf(format, args...)
}

// Warn logs a warning message to stdout.
func Warn(format string, args ...interface{}) {
	// Create logger with current settings (thread-safe)
	output := zerolog.ConsoleWriter{
		Out:        getStdout(),
		NoColor:    false,
		TimeFormat: "15:04:05",
	}
	log := zerolog.New(output).With().Timestamp().Logger()

	mu.RLock()
	if dryRun {
		log = log.With().Bool(dryRunField, true).Logger()
	}
	mu.RUnlock()

	log.Warn().Msgf(format, args...)
}
