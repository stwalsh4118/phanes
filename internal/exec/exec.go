package exec

import (
	"os"
	"os/exec"
)

// Run executes a command with the given name and arguments.
// The command's stdout and stderr are connected to os.Stdout and os.Stderr respectively.
// Returns an error if the command fails to execute or exits with a non-zero status.
func Run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// RunWithOutput executes a command with the given name and arguments and captures its stdout.
// Returns the command's stdout as a string and an error if the command fails to execute
// or exits with a non-zero status.
func RunWithOutput(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// CommandExists checks if a command exists in the system PATH.
// Returns true if the command is found, false otherwise.
func CommandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// FileExists checks if a file or directory exists at the given path.
// Returns true if the path exists, false otherwise.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// WriteFile writes content to a file at the given path with the specified permissions.
// If the file already exists, it will be overwritten.
// Returns an error if the file cannot be written.
func WriteFile(path string, content []byte, perm os.FileMode) error {
	return os.WriteFile(path, content, perm)
}
