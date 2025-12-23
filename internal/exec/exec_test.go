package exec

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name    string
		command string
		args    []string
		wantErr bool
	}{
		{
			name:    "successful command",
			command: "echo",
			args:    []string{"test"},
			wantErr: false,
		},
		{
			name:    "command that exits with error",
			command: "false",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "non-existent command",
			command: "nonexistentcommand12345",
			args:    []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Run(tt.command, tt.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunWithOutput(t *testing.T) {
	tests := []struct {
		name      string
		command   string
		args      []string
		wantErr   bool
		checkFunc func(string) bool
	}{
		{
			name:    "successful command with output",
			command: "echo",
			args:    []string{"hello"},
			wantErr: false,
			checkFunc: func(output string) bool {
				return output == "hello\n" || output == "hello\r\n"
			},
		},
		{
			name:    "command that exits with error",
			command: "false",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "non-existent command",
			command: "nonexistentcommand12345",
			args:    []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := RunWithOutput(tt.command, tt.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunWithOutput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.checkFunc != nil {
				if !tt.checkFunc(output) {
					t.Errorf("RunWithOutput() output = %q, did not match expected pattern", output)
				}
			}
		})
	}
}

func TestCommandExists(t *testing.T) {
	tests := []struct {
		name    string
		command string
		want    bool
	}{
		{
			name:    "existing command",
			command: "echo",
			want:    true,
		},
		{
			name:    "non-existent command",
			command: "nonexistentcommand12345",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CommandExists(tt.command)
			if got != tt.want {
				t.Errorf("CommandExists(%q) = %v, want %v", tt.command, got, tt.want)
			}
		})
	}
}

func TestFileExists(t *testing.T) {
	// Create a temporary file for testing
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "testfile.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "existing file",
			path: tmpFile,
			want: true,
		},
		{
			name: "existing directory",
			path: tmpDir,
			want: true,
		},
		{
			name: "non-existent file",
			path: filepath.Join(tmpDir, "nonexistent.txt"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FileExists(tt.path)
			if got != tt.want {
				t.Errorf("FileExists(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestWriteFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "write_test.txt")
	testContent := []byte("test content")
	testPerm := os.FileMode(0644)

	tests := []struct {
		name    string
		path    string
		content []byte
		perm    os.FileMode
		wantErr bool
	}{
		{
			name:    "write new file",
			path:    testFile,
			content: testContent,
			perm:    testPerm,
			wantErr: false,
		},
		{
			name:    "write to non-existent directory",
			path:    filepath.Join(tmpDir, "nonexistent", "file.txt"),
			content: testContent,
			perm:    testPerm,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := WriteFile(tt.path, tt.content, tt.perm)
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				// Verify file was created with correct content
				readContent, err := os.ReadFile(tt.path)
				if err != nil {
					t.Errorf("Failed to read written file: %v", err)
					return
				}
				if string(readContent) != string(tt.content) {
					t.Errorf("WriteFile() content = %q, want %q", string(readContent), string(tt.content))
				}
				// Verify permissions (on Unix-like systems)
				if runtime.GOOS != "windows" {
					info, err := os.Stat(tt.path)
					if err != nil {
						t.Errorf("Failed to stat file: %v", err)
						return
					}
					gotPerm := info.Mode().Perm()
					if gotPerm != tt.perm {
						t.Errorf("WriteFile() permissions = %o, want %o", gotPerm, tt.perm)
					}
				}
			}
		})
	}
}

func TestWriteFileOverwritesExisting(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "overwrite_test.txt")
	originalContent := []byte("original")
	newContent := []byte("new content")
	testPerm := os.FileMode(0644)

	// Create original file
	if err := os.WriteFile(testFile, originalContent, testPerm); err != nil {
		t.Fatalf("Failed to create original file: %v", err)
	}

	// Overwrite with new content
	if err := WriteFile(testFile, newContent, testPerm); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	// Verify content was overwritten
	readContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read overwritten file: %v", err)
	}
	if string(readContent) != string(newContent) {
		t.Errorf("WriteFile() did not overwrite correctly: got %q, want %q", string(readContent), string(newContent))
	}
}
