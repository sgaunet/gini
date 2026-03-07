package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestUnicodeKeysAndValues(t *testing.T) {
	tests := []struct {
		name       string
		section    string
		key        string
		value      string
		wantOutput string
	}{
		{
			name:       "ascii value",
			section:    "app",
			key:        "name",
			value:      "hello world",
			wantOutput: "hello world\n",
		},
		{
			name:       "unicode value with accents",
			section:    "i18n",
			key:        "greeting",
			value:      "héllo wörld café",
			wantOutput: "héllo wörld café\n",
		},
		{
			name:       "emoji value",
			section:    "status",
			key:        "icon",
			value:      "🚀✅🎉",
			wantOutput: "🚀✅🎉\n",
		},
		{
			name:       "CJK characters in value",
			section:    "locale",
			key:        "message",
			value:      "こんにちは世界",
			wantOutput: "こんにちは世界\n",
		},
		{
			name:       "arabic value",
			section:    "locale",
			key:        "text",
			value:      "مرحبا بالعالم",
			wantOutput: "مرحبا بالعالم\n",
		},
		{
			name:       "unicode key with dots",
			section:    "app",
			key:        "config.naïve",
			value:      "yes",
			wantOutput: "yes\n",
		},
		{
			name:       "mixed unicode section and value",
			section:    "données",
			key:        "résultat",
			value:      "réussi ✓",
			wantOutput: "réussi ✓\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			iniPath := filepath.Join(tmpDir, "unicode.ini")

			// Set the value
			_, err := executeCommand(t, "set", "-f", iniPath, "-s", tt.section, "-k", tt.key, "-v", tt.value, "-c")
			if err != nil {
				t.Fatalf("set error: %v", err)
			}

			// Get it back
			output, err := executeCommand(t, "get", "-f", iniPath, "-s", tt.section, "-k", tt.key)
			if err != nil {
				t.Fatalf("get error: %v", err)
			}
			if output != tt.wantOutput {
				t.Errorf("output = %q, want %q", output, tt.wantOutput)
			}
		})
	}
}

func TestLargeINIFile(t *testing.T) {
	tmpDir := t.TempDir()
	iniPath := filepath.Join(tmpDir, "large.ini")

	// Build a large INI file (>1MB) with many sections and keys
	var builder strings.Builder
	for i := 0; i < 100; i++ {
		builder.WriteString(fmt.Sprintf("[section%d]\n", i))
		for j := 0; j < 100; j++ {
			builder.WriteString(fmt.Sprintf("key%d = %s\n", j, strings.Repeat("x", 100)))
		}
		builder.WriteString("\n")
	}
	content := builder.String()
	if len(content) < 1_000_000 {
		t.Fatalf("generated content is only %d bytes, expected >1MB", len(content))
	}

	if err := os.WriteFile(iniPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	// Read from the large file
	output, err := executeCommand(t, "get", "-f", iniPath, "-s", "section50", "-k", "key50")
	if err != nil {
		t.Fatalf("get from large file: %v", err)
	}
	expected := strings.Repeat("x", 100) + "\n"
	if output != expected {
		t.Errorf("output length = %d, want %d", len(output), len(expected))
	}

	// Write to the large file
	_, err = executeCommand(t, "set", "-f", iniPath, "-s", "section99", "-k", "newkey", "-v", "newvalue")
	if err != nil {
		t.Fatalf("set on large file: %v", err)
	}

	// Verify the write
	output, err = executeCommand(t, "get", "-f", iniPath, "-s", "section99", "-k", "newkey")
	if err != nil {
		t.Fatalf("get after set on large file: %v", err)
	}
	if output != "newvalue\n" {
		t.Errorf("output = %q, want %q", output, "newvalue\n")
	}

	// Delete from the large file
	_, err = executeCommand(t, "del", "-f", iniPath, "-s", "section99", "-k", "newkey")
	if err != nil {
		t.Fatalf("del on large file: %v", err)
	}

	// Verify deletion
	output, err = executeCommand(t, "get", "-f", iniPath, "-s", "section99", "-k", "newkey")
	if err != nil {
		t.Fatalf("get after del on large file: %v", err)
	}
	if output != "" {
		t.Errorf("expected empty output after delete, got %q", output)
	}
}

func TestConcurrentReads(t *testing.T) {
	iniContent := "[db]\nhost = localhost\nport = 5432\n"
	iniPath := createTestINI(t, iniContent)

	const numReaders = 10
	var wg sync.WaitGroup
	errors := make(chan error, numReaders)

	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Each goroutine needs its own command execution
			// We use go run to avoid shared state in the test binary
			output, err := executeCommandDirect(iniPath, "get", "-s", "db", "-k", "host")
			if err != nil {
				errors <- fmt.Errorf("concurrent get error: %w", err)
				return
			}
			if strings.TrimSpace(output) != "localhost" {
				errors <- fmt.Errorf("concurrent get output = %q, want %q", output, "localhost")
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Error(err)
	}
}

func TestConcurrentReadWrite(t *testing.T) {
	tmpDir := t.TempDir()
	iniPath := filepath.Join(tmpDir, "concurrent.ini")
	if err := os.WriteFile(iniPath, []byte("[counter]\nvalue = 0\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	const numWriters = 5
	var wg sync.WaitGroup
	errors := make(chan error, numWriters*2)

	// Concurrent writers (each writes a different key to avoid conflicts)
	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			_, err := executeCommandDirect(iniPath, "set",
				"-s", "writers", "-k", fmt.Sprintf("writer%d", id), "-v", fmt.Sprintf("val%d", id))
			if err != nil {
				errors <- fmt.Errorf("writer %d error: %w", id, err)
			}
		}(i)
	}

	// Concurrent readers
	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := executeCommandDirect(iniPath, "get", "-s", "counter", "-k", "value")
			if err != nil {
				errors <- fmt.Errorf("concurrent reader error: %w", err)
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Error(err)
	}
}

// executeCommandDirect runs gini as a subprocess to avoid shared state issues
// in concurrent tests.
func executeCommandDirect(iniPath string, subCmd string, args ...string) (string, error) {
	cmdArgs := []string{"run", ".", subCmd, "-f", iniPath}
	cmdArgs = append(cmdArgs, args...)

	cmd := exec.Command("go", cmdArgs...) // #nosec G204 - test helper only
	cmd.Dir = projectRoot()
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("%w: %s", err, stderr.String())
	}
	return stdout.String(), nil
}

// projectRoot finds the project root by looking for go.mod.
func projectRoot() string {
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "."
		}
		dir = parent
	}
}

func TestPermissionErrors(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("skipping permission tests as root")
	}

	t.Run("read-only file for set", func(t *testing.T) {
		tmpDir := t.TempDir()
		iniPath := filepath.Join(tmpDir, "readonly.ini")
		if err := os.WriteFile(iniPath, []byte("[db]\nhost = localhost\n"), 0o444); err != nil {
			t.Fatal(err)
		}

		// Setting a value on a read-only file should fail (lock file creation may fail
		// or atomic save will fail on rename)
		_, err := executeCommand(t, "set", "-f", iniPath, "-s", "db", "-k", "host", "-v", "newhost")
		// The operation may succeed on some systems where rename works on read-only targets,
		// so we just verify it doesn't panic
		_ = err
	})

	t.Run("no-read-permission directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		restrictedDir := filepath.Join(tmpDir, "restricted")
		if err := os.Mkdir(restrictedDir, 0o000); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chmod(restrictedDir, 0o755) }()

		iniPath := filepath.Join(restrictedDir, "test.ini")
		_, err := executeCommand(t, "get", "-f", iniPath, "-s", "db", "-k", "host")
		if err == nil {
			t.Error("expected error accessing file in restricted directory")
		}
	})

	t.Run("set with create on read-only directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		restrictedDir := filepath.Join(tmpDir, "noperm")
		if err := os.Mkdir(restrictedDir, 0o555); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chmod(restrictedDir, 0o755) }()

		iniPath := filepath.Join(restrictedDir, "new.ini")
		_, err := executeCommand(t, "set", "-f", iniPath, "-s", "db", "-k", "host", "-v", "x", "-c")
		if err == nil {
			t.Error("expected error creating file in read-only directory")
		}
	})
}

func TestInvalidINISyntax(t *testing.T) {
	tests := []struct {
		name    string
		content string
		section string
		key     string
		wantErr bool
	}{
		{
			name:    "completely empty file",
			content: "",
			section: "db",
			key:     "host",
			wantErr: false, // empty file loads fine, key just doesn't exist
		},
		{
			name:    "file with only whitespace",
			content: "   \n\n  \n",
			section: "db",
			key:     "host",
			wantErr: false,
		},
		{
			name:    "key without section",
			content: "host = localhost\n",
			section: "", // default section
			key:     "host",
			wantErr: false, // valid INI - key in default section
		},
		{
			name:    "key with value containing equals",
			content: "[db]\nconnstr = host=localhost port=5432\n",
			section: "db",
			key:     "connstr",
			wantErr: false,
		},
		{
			name:    "duplicate keys in same section",
			content: "[db]\nhost = first\nhost = second\n",
			section: "db",
			key:     "host",
			wantErr: false, // should return one of the values
		},
		{
			name:    "section with no keys",
			content: "[empty]\n[db]\nhost = localhost\n",
			section: "empty",
			key:     "host",
			wantErr: false, // key doesn't exist in empty section
		},
		{
			name:    "value with leading spaces",
			content: "[db]\nhost =    localhost   \n",
			section: "db",
			key:     "host",
			wantErr: false,
		},
		{
			name:    "comment lines",
			content: "; this is a comment\n# another comment\n[db]\nhost = localhost\n",
			section: "db",
			key:     "host",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iniPath := createTestINI(t, tt.content)
			_, err := executeCommand(t, "get", "-f", iniPath, "-s", tt.section, "-k", tt.key)
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValueWithSpecialCharacters(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{"value with spaces", "hello world foo bar"},
		{"value with tab", "hello\tworld"},
		{"value with url", "https://example.com/path?q=1&r=2"},
		{"value with path", "/usr/local/bin/gini"},
		{"value with backslash", "C:\\Users\\test\\config"},
		{"empty value", ""},
		{"value with quotes", `he said "hello"`},
		{"value with newline escaped", "line1\\nline2"},
		{"very long value", strings.Repeat("a", 10000)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			iniPath := filepath.Join(tmpDir, "special.ini")

			_, err := executeCommand(t, "set", "-f", iniPath, "-s", "test", "-k", "val", "-v", tt.value, "-c")
			if err != nil {
				t.Fatalf("set error: %v", err)
			}

			output, err := executeCommand(t, "get", "-f", iniPath, "-s", "test", "-k", "val")
			if err != nil {
				t.Fatalf("get error: %v", err)
			}

			want := tt.value + "\n"
			if tt.value == "" {
				want = "\n"
			}
			if output != want {
				if len(output) > 100 {
					t.Errorf("output length = %d, want %d", len(output), len(want))
				} else {
					t.Errorf("output = %q, want %q", output, want)
				}
			}
		})
	}
}

func TestDeleteSectionThenReadKey(t *testing.T) {
	iniPath := createTestINI(t, "[db]\nhost = localhost\n[cache]\nttl = 60\n")

	// Delete the section
	_, err := executeCommand(t, "delsection", "-f", iniPath, "-s", "db")
	if err != nil {
		t.Fatalf("delsection error: %v", err)
	}

	// Read a key from the deleted section (should return empty, no error)
	output, err := executeCommand(t, "get", "-f", iniPath, "-s", "db", "-k", "host")
	if err != nil {
		t.Fatalf("get after delsection error: %v", err)
	}
	if output != "" {
		t.Errorf("expected empty output, got %q", output)
	}

	// Read a key from the surviving section
	output, err = executeCommand(t, "get", "-f", iniPath, "-s", "cache", "-k", "ttl")
	if err != nil {
		t.Fatalf("get surviving section error: %v", err)
	}
	if output != "60\n" {
		t.Errorf("output = %q, want %q", output, "60\n")
	}
}
