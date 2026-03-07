package tools

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/ini.v1"
)

func TestIsFileExists(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	existingFile := filepath.Join(tmpDir, "exists.ini")
	if err := os.WriteFile(existingFile, []byte("[section]\nkey=value\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		path string
		want bool
	}{
		{"existing file", existingFile, true},
		{"non-existent file", filepath.Join(tmpDir, "nope.ini"), false},
		{"directory", tmpDir, false},
		{"empty path", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := IsFileExists(tt.path)
			if got != tt.want {
				t.Errorf("IsFileExists(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestTouchFile(t *testing.T) {
	t.Parallel()

	t.Run("creates new file", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "newfile.ini")

		err := TouchFile(path)
		if err != nil {
			t.Fatalf("TouchFile() error = %v", err)
		}

		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("file not created: %v", err)
		}
		if info.Size() != 0 {
			t.Errorf("file size = %d, want 0", info.Size())
		}
	})

	t.Run("error on invalid path", func(t *testing.T) {
		t.Parallel()
		err := TouchFile("/nonexistent/dir/file.ini")
		if err == nil {
			t.Error("TouchFile() expected error for invalid path")
		}
	})
}

func TestValidateKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		key     string
		wantErr error
	}{
		{"valid key", "host", nil},
		{"valid key with underscore", "db_host", nil},
		{"valid key with dot", "app.name", nil},
		{"empty key", "", ErrEmptyKey},
		{"leading whitespace", " key", ErrKeyWhitespace},
		{"trailing whitespace", "key ", ErrKeyWhitespace},
		{"contains equals", "key=val", ErrKeyInvalidChar},
		{"contains open bracket", "key[0]", ErrKeyInvalidChar},
		{"contains semicolon", "key;comment", ErrKeyInvalidChar},
		{"contains hash", "key#tag", ErrKeyInvalidChar},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateKey(tt.key)
			if tt.wantErr == nil {
				if err != nil {
					t.Errorf("ValidateKey(%q) unexpected error: %v", tt.key, err)
				}
				return
			}
			if err == nil {
				t.Errorf("ValidateKey(%q) expected error %v, got nil", tt.key, tt.wantErr)
				return
			}
			// Check that the error wraps the expected sentinel
			if !containsError(err, tt.wantErr) {
				t.Errorf("ValidateKey(%q) error = %v, want %v", tt.key, err, tt.wantErr)
			}
		})
	}
}

func TestValidateSection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		section string
		wantErr error
	}{
		{"valid section", "database", nil},
		{"empty section (default)", "", nil},
		{"valid with underscore", "my_section", nil},
		{"leading whitespace", " section", ErrSectionWhitespace},
		{"trailing whitespace", "section ", ErrSectionWhitespace},
		{"contains bracket", "sect[ion]", ErrSectionInvalidChar},
		{"contains equals", "sect=ion", ErrSectionInvalidChar},
		{"contains semicolon", "sect;ion", ErrSectionInvalidChar},
		{"contains hash", "sect#ion", ErrSectionInvalidChar},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateSection(tt.section)
			if tt.wantErr == nil {
				if err != nil {
					t.Errorf("ValidateSection(%q) unexpected error: %v", tt.section, err)
				}
				return
			}
			if err == nil {
				t.Errorf("ValidateSection(%q) expected error %v, got nil", tt.section, tt.wantErr)
				return
			}
			if !containsError(err, tt.wantErr) {
				t.Errorf("ValidateSection(%q) error = %v, want %v", tt.section, err, tt.wantErr)
			}
		})
	}
}

func TestAtomicSave(t *testing.T) {
	t.Parallel()

	t.Run("saves file atomically", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		target := filepath.Join(tmpDir, "test.ini")

		cfg := ini.Empty()
		cfg.Section("database").Key("host").SetValue("localhost")
		cfg.Section("database").Key("port").SetValue("5432")

		err := AtomicSave(cfg, target)
		if err != nil {
			t.Fatalf("AtomicSave() error = %v", err)
		}

		// Verify file was written correctly
		loaded, err := ini.Load(target)
		if err != nil {
			t.Fatalf("failed to reload saved file: %v", err)
		}
		if got := loaded.Section("database").Key("host").String(); got != "localhost" {
			t.Errorf("host = %q, want %q", got, "localhost")
		}
		if got := loaded.Section("database").Key("port").String(); got != "5432" {
			t.Errorf("port = %q, want %q", got, "5432")
		}
	})

	t.Run("overwrites existing file", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		target := filepath.Join(tmpDir, "overwrite.ini")

		// Write initial content
		if err := os.WriteFile(target, []byte("[old]\nkey=old\n"), 0o644); err != nil {
			t.Fatal(err)
		}

		cfg := ini.Empty()
		cfg.Section("new").Key("key").SetValue("new")

		err := AtomicSave(cfg, target)
		if err != nil {
			t.Fatalf("AtomicSave() error = %v", err)
		}

		loaded, err := ini.Load(target)
		if err != nil {
			t.Fatalf("failed to reload: %v", err)
		}
		if got := loaded.Section("new").Key("key").String(); got != "new" {
			t.Errorf("key = %q, want %q", got, "new")
		}
	})

	t.Run("error on invalid directory", func(t *testing.T) {
		t.Parallel()
		err := AtomicSave(ini.Empty(), "/nonexistent/dir/test.ini")
		if err == nil {
			t.Error("AtomicSave() expected error for invalid directory")
		}
	})
}

func TestTouchFile_OverwriteExisting(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "existing.ini")

	// Create file with content
	if err := os.WriteFile(path, []byte("[section]\nkey=value\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// TouchFile should overwrite to empty
	err := TouchFile(path)
	if err != nil {
		t.Fatalf("TouchFile() error = %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("file not found: %v", err)
	}
	if info.Size() != 0 {
		t.Errorf("file size = %d, want 0", info.Size())
	}
}

func TestAtomicSave_PreservesContentOnSuccess(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "multi.ini")

	cfg := ini.Empty()
	cfg.Section("s1").Key("a").SetValue("1")
	cfg.Section("s2").Key("b").SetValue("2")
	cfg.Section("s2").Key("c").SetValue("3")

	err := AtomicSave(cfg, target)
	if err != nil {
		t.Fatalf("AtomicSave() error = %v", err)
	}

	loaded, err := ini.Load(target)
	if err != nil {
		t.Fatalf("failed to reload: %v", err)
	}
	if got := loaded.Section("s1").Key("a").String(); got != "1" {
		t.Errorf("s1.a = %q, want %q", got, "1")
	}
	if got := loaded.Section("s2").Key("b").String(); got != "2" {
		t.Errorf("s2.b = %q, want %q", got, "2")
	}
	if got := loaded.Section("s2").Key("c").String(); got != "3" {
		t.Errorf("s2.c = %q, want %q", got, "3")
	}

	// Verify no temp files left behind
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".tmp" {
			t.Errorf("temp file left behind: %s", e.Name())
		}
	}
}

func TestAtomicSave_ReadOnlyTarget(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "readonly.ini")

	// Create initial file
	if err := os.WriteFile(target, []byte("[old]\nk=v\n"), 0o444); err != nil {
		t.Fatal(err)
	}

	cfg := ini.Empty()
	cfg.Section("new").Key("k").SetValue("v")

	// AtomicSave should still succeed because rename replaces the target
	err := AtomicSave(cfg, target)
	if err != nil {
		t.Fatalf("AtomicSave() error = %v", err)
	}

	loaded, err := ini.Load(target)
	if err != nil {
		t.Fatalf("failed to reload: %v", err)
	}
	if got := loaded.Section("new").Key("k").String(); got != "v" {
		t.Errorf("key = %q, want %q", got, "v")
	}
}

func TestIsFileExists_Symlink(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	realFile := filepath.Join(tmpDir, "real.ini")
	if err := os.WriteFile(realFile, []byte("[s]\nk=v\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	symlink := filepath.Join(tmpDir, "link.ini")
	if err := os.Symlink(realFile, symlink); err != nil {
		t.Skip("symlinks not supported")
	}

	if !IsFileExists(symlink) {
		t.Error("IsFileExists(symlink) = false, want true")
	}

	// Broken symlink
	brokenLink := filepath.Join(tmpDir, "broken.ini")
	if err := os.Symlink(filepath.Join(tmpDir, "nonexistent"), brokenLink); err != nil {
		t.Skip("symlinks not supported")
	}

	if IsFileExists(brokenLink) {
		t.Error("IsFileExists(broken symlink) = true, want false")
	}
}

func TestValidateKey_CloseBracket(t *testing.T) {
	t.Parallel()
	err := ValidateKey("key]")
	if err == nil {
		t.Error("ValidateKey(\"key]\") expected error")
	}
}

func TestValidateSection_CloseBracketOnly(t *testing.T) {
	t.Parallel()
	err := ValidateSection("]")
	if err == nil {
		t.Error("ValidateSection(\"]\") expected error")
	}
}

// containsError checks if err's message contains the target error's message.
func containsError(err, target error) bool {
	if err == nil || target == nil {
		return err == target
	}
	return contains(err.Error(), target.Error())
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
