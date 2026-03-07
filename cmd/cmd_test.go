package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// executeCommand runs the root command with the given args and captures stdout output.
func executeCommand(t *testing.T, args ...string) (string, error) {
	t.Helper()

	// Reset config to defaults before each test
	cfg = Config{}

	// Capture stdout since commands use fmt.Println
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	rootCmd.SetArgs(args)
	execErr := rootCmd.Execute()

	// Restore stdout and read captured output
	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatal(err)
	}

	return buf.String(), execErr
}

// createTestINI creates a temporary INI file with given content and returns its path.
func createTestINI(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.ini")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestGetCommand(t *testing.T) {
	iniContent := "[database]\nhost = localhost\nport = 5432\n\n[app]\nname = myapp\n"

	tests := []struct {
		name       string
		args       []string
		wantOutput string
		wantErr    bool
	}{
		{
			name:       "get existing key",
			args:       []string{"get", "-f", "PLACEHOLDER", "-s", "database", "-k", "host"},
			wantOutput: "localhost\n",
			wantErr:    false,
		},
		{
			name:       "get non-existent key (non-strict)",
			args:       []string{"get", "-f", "PLACEHOLDER", "-s", "database", "-k", "missing"},
			wantOutput: "",
			wantErr:    false,
		},
		{
			name:       "get non-existent key (strict)",
			args:       []string{"get", "--strict", "-f", "PLACEHOLDER", "-s", "database", "-k", "missing"},
			wantOutput: "",
			wantErr:    true,
		},
		{
			name:       "get key from default section",
			args:       []string{"get", "-f", "PLACEHOLDER", "-s", "", "-k", "host"},
			wantOutput: "",
			wantErr:    false,
		},
		{
			name:    "get with non-existent file",
			args:    []string{"get", "-f", "/nonexistent/file.ini", "-s", "db", "-k", "host"},
			wantErr: true,
		},
		{
			name:    "get with invalid key",
			args:    []string{"get", "-f", "PLACEHOLDER", "-s", "db", "-k", "key=bad"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iniPath := createTestINI(t, iniContent)
			// Replace PLACEHOLDER with actual path
			args := make([]string, len(tt.args))
			copy(args, tt.args)
			for i, a := range args {
				if a == "PLACEHOLDER" {
					args[i] = iniPath
				}
			}

			output, err := executeCommand(t, args...)
			if tt.wantErr && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.wantOutput != "" && output != tt.wantOutput {
				t.Errorf("output = %q, want %q", output, tt.wantOutput)
			}
		})
	}
}

func TestSetCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		setup   string // initial INI content
		wantKey string // key to check after set
		wantVal string // expected value
		wantErr bool
	}{
		{
			name:    "set new key",
			args:    []string{"set", "-f", "PLACEHOLDER", "-s", "database", "-k", "host", "-v", "localhost"},
			setup:   "[database]\nport = 5432\n",
			wantKey: "host",
			wantVal: "localhost",
			wantErr: false,
		},
		{
			name:    "update existing key",
			args:    []string{"set", "-f", "PLACEHOLDER", "-s", "database", "-k", "port", "-v", "3306"},
			setup:   "[database]\nport = 5432\n",
			wantKey: "port",
			wantVal: "3306",
			wantErr: false,
		},
		{
			name:    "set with create flag on new file",
			args:    []string{"set", "-f", "NEWFILE", "-s", "app", "-k", "name", "-v", "test", "-c"},
			setup:   "",
			wantKey: "name",
			wantVal: "test",
			wantErr: false,
		},
		{
			name:    "set on non-existent file without create flag",
			args:    []string{"set", "-f", "/nonexistent/file.ini", "-s", "db", "-k", "host", "-v", "x"},
			setup:   "",
			wantErr: true,
		},
		{
			name:    "set with invalid key",
			args:    []string{"set", "-f", "PLACEHOLDER", "-s", "db", "-k", "k=v", "-v", "x"},
			setup:   "[db]\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			var iniPath string

			args := make([]string, len(tt.args))
			copy(args, tt.args)

			for i, a := range args {
				if a == "PLACEHOLDER" {
					iniPath = filepath.Join(tmpDir, "test.ini")
					if err := os.WriteFile(iniPath, []byte(tt.setup), 0o644); err != nil {
						t.Fatal(err)
					}
					args[i] = iniPath
				}
				if a == "NEWFILE" {
					iniPath = filepath.Join(tmpDir, "new.ini")
					args[i] = iniPath
				}
			}

			_, err := executeCommand(t, args...)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify the value was set
			if tt.wantKey != "" {
				content, err := os.ReadFile(iniPath)
				if err != nil {
					t.Fatalf("failed to read file: %v", err)
				}
				if !bytes.Contains(content, []byte(tt.wantVal)) {
					t.Errorf("file content does not contain %q:\n%s", tt.wantVal, content)
				}
			}
		})
	}
}

func TestDelCommand(t *testing.T) {
	iniContent := "[database]\nhost = localhost\nport = 5432\n"

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "delete existing key",
			args:    []string{"del", "-f", "PLACEHOLDER", "-s", "database", "-k", "host"},
			wantErr: false,
		},
		{
			name:    "delete non-existent key (non-strict)",
			args:    []string{"del", "-f", "PLACEHOLDER", "-s", "database", "-k", "missing"},
			wantErr: false,
		},
		{
			name:    "delete non-existent key (strict)",
			args:    []string{"del", "--strict", "-f", "PLACEHOLDER", "-s", "database", "-k", "missing"},
			wantErr: true,
		},
		{
			name:    "delete from non-existent file",
			args:    []string{"del", "-f", "/nonexistent/file.ini", "-s", "db", "-k", "host"},
			wantErr: true,
		},
		{
			name:    "delete with invalid key",
			args:    []string{"del", "-f", "PLACEHOLDER", "-s", "db", "-k", "k[0]"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iniPath := createTestINI(t, iniContent)
			args := make([]string, len(tt.args))
			copy(args, tt.args)
			for i, a := range args {
				if a == "PLACEHOLDER" {
					args[i] = iniPath
				}
			}

			_, err := executeCommand(t, args...)
			if tt.wantErr && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}

	t.Run("verify key deleted from file", func(t *testing.T) {
		iniPath := createTestINI(t, iniContent)
		_, err := executeCommand(t, "del", "-f", iniPath, "-s", "database", "-k", "host")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		content, err := os.ReadFile(iniPath)
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}
		if bytes.Contains(content, []byte("host")) {
			t.Errorf("file still contains deleted key 'host':\n%s", content)
		}
	})
}

func TestDelSectionCommand(t *testing.T) {
	iniContent := "[database]\nhost = localhost\n\n[cache]\nttl = 60\n"

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "delete existing section",
			args:    []string{"delsection", "-f", "PLACEHOLDER", "-s", "cache"},
			wantErr: false,
		},
		{
			name:    "delete non-existent section (non-strict)",
			args:    []string{"delsection", "-f", "PLACEHOLDER", "-s", "missing"},
			wantErr: false,
		},
		{
			name:    "delete non-existent section (strict)",
			args:    []string{"delsection", "--strict", "-f", "PLACEHOLDER", "-s", "missing"},
			wantErr: true,
		},
		{
			name:    "delete from non-existent file",
			args:    []string{"delsection", "-f", "/nonexistent/file.ini", "-s", "db"},
			wantErr: true,
		},
		{
			name:    "delete with invalid section",
			args:    []string{"delsection", "-f", "PLACEHOLDER", "-s", "sect[ion]"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iniPath := createTestINI(t, iniContent)
			args := make([]string, len(tt.args))
			copy(args, tt.args)
			for i, a := range args {
				if a == "PLACEHOLDER" {
					args[i] = iniPath
				}
			}

			_, err := executeCommand(t, args...)
			if tt.wantErr && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}

	t.Run("verify section deleted from file", func(t *testing.T) {
		iniPath := createTestINI(t, iniContent)
		_, err := executeCommand(t, "delsection", "-f", iniPath, "-s", "cache")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		content, err := os.ReadFile(iniPath)
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}
		if bytes.Contains(content, []byte("[cache]")) {
			t.Errorf("file still contains deleted section [cache]:\n%s", content)
		}
	})
}

func TestVersionCommand(t *testing.T) {
	output, err := executeCommand(t, "version")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output == "" {
		t.Error("version command returned empty output")
	}
}

func TestGetCommand_DebugFlag(t *testing.T) {
	iniPath := createTestINI(t, "[db]\nhost = localhost\n")
	output, err := executeCommand(t, "get", "--debug", "-f", iniPath, "-s", "db", "-k", "host")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output != "localhost\n" {
		t.Errorf("output = %q, want %q", output, "localhost\n")
	}
}

func TestGetCommand_QuietFlag(t *testing.T) {
	iniPath := createTestINI(t, "[db]\nhost = localhost\n")
	output, err := executeCommand(t, "get", "--quiet", "-f", iniPath, "-s", "db", "-k", "host")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output != "localhost\n" {
		t.Errorf("output = %q, want %q", output, "localhost\n")
	}
}

func TestGetCommand_NoFileFlag(t *testing.T) {
	_, err := executeCommand(t, "get", "-s", "db", "-k", "host")
	if err == nil {
		t.Error("expected error when --file not provided")
	}
}

func TestSetCommand_NoFileFlag(t *testing.T) {
	_, err := executeCommand(t, "set", "-s", "db", "-k", "host", "-v", "x")
	if err == nil {
		t.Error("expected error when --file not provided")
	}
}

func TestSetCommand_InvalidSection(t *testing.T) {
	iniPath := createTestINI(t, "[db]\n")
	_, err := executeCommand(t, "set", "-f", iniPath, "-s", "db[x]", "-k", "host", "-v", "x")
	if err == nil {
		t.Error("expected error for invalid section")
	}
}

func TestSetCommand_CreateExistingFile(t *testing.T) {
	iniPath := createTestINI(t, "[db]\nport = 5432\n")
	_, err := executeCommand(t, "set", "-f", iniPath, "-s", "db", "-k", "host", "-v", "localhost", "-c")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	content, err := os.ReadFile(iniPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(content, []byte("localhost")) {
		t.Errorf("file does not contain expected value:\n%s", content)
	}
}

func TestDelCommand_InvalidSection(t *testing.T) {
	iniPath := createTestINI(t, "[db]\nhost = localhost\n")
	_, err := executeCommand(t, "del", "-f", iniPath, "-s", "db[x]", "-k", "host")
	if err == nil {
		t.Error("expected error for invalid section")
	}
}

func TestDelSectionCommand_InvalidSection(t *testing.T) {
	iniPath := createTestINI(t, "[db]\nhost = localhost\n")
	_, err := executeCommand(t, "delsection", "-f", iniPath, "-s", "db=x")
	if err == nil {
		t.Error("expected error for invalid section")
	}
}

func TestGetCommand_InvalidSection(t *testing.T) {
	iniPath := createTestINI(t, "[db]\nhost = localhost\n")
	_, err := executeCommand(t, "get", "-f", iniPath, "-s", "db;x", "-k", "host")
	if err == nil {
		t.Error("expected error for invalid section")
	}
}
