package inifile

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sgaunet/gini/internal/tools"
)

func createTestINI(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.ini")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestValidateAndLoad(t *testing.T) {
	t.Parallel()

	iniContent := "[database]\nhost = localhost\nport = 5432\n"

	tests := []struct {
		name    string
		path    string
		section string
		key     string
		mode    tools.LockMode
		wantErr bool
	}{
		{
			name:    "valid load with shared lock",
			path:    "PLACEHOLDER",
			section: "database",
			key:     "host",
			mode:    tools.SharedLock,
			wantErr: false,
		},
		{
			name:    "valid load with exclusive lock",
			path:    "PLACEHOLDER",
			section: "database",
			key:     "host",
			mode:    tools.ExclusiveLock,
			wantErr: false,
		},
		{
			name:    "empty path",
			path:    "",
			section: "db",
			key:     "host",
			mode:    tools.SharedLock,
			wantErr: true,
		},
		{
			name:    "empty key",
			path:    "PLACEHOLDER",
			section: "db",
			key:     "",
			mode:    tools.SharedLock,
			wantErr: true,
		},
		{
			name:    "invalid key",
			path:    "PLACEHOLDER",
			section: "db",
			key:     "key=val",
			mode:    tools.SharedLock,
			wantErr: true,
		},
		{
			name:    "invalid section",
			path:    "PLACEHOLDER",
			section: "db[x]",
			key:     "host",
			mode:    tools.SharedLock,
			wantErr: true,
		},
		{
			name:    "non-existent file",
			path:    "/nonexistent/file.ini",
			section: "db",
			key:     "host",
			mode:    tools.SharedLock,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			path := tt.path
			if path == "PLACEHOLDER" {
				path = createTestINI(t, iniContent)
			}

			cfg, lock, err := ValidateAndLoad(path, tt.section, tt.key, tt.mode)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
					if lock != nil {
						_ = lock.Unlock()
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg == nil {
				t.Fatal("cfg is nil")
			}
			if lock == nil {
				t.Fatal("lock is nil")
			}
			_ = lock.Unlock()
		})
	}
}

func TestValidateSectionAndLoad(t *testing.T) {
	t.Parallel()

	iniContent := "[cache]\nttl = 60\n"

	tests := []struct {
		name    string
		path    string
		section string
		mode    tools.LockMode
		wantErr bool
	}{
		{
			name:    "valid load",
			path:    "PLACEHOLDER",
			section: "cache",
			mode:    tools.ExclusiveLock,
			wantErr: false,
		},
		{
			name:    "empty section (default)",
			path:    "PLACEHOLDER",
			section: "",
			mode:    tools.SharedLock,
			wantErr: false,
		},
		{
			name:    "empty path",
			path:    "",
			section: "cache",
			mode:    tools.SharedLock,
			wantErr: true,
		},
		{
			name:    "invalid section",
			path:    "PLACEHOLDER",
			section: "sect;ion",
			mode:    tools.SharedLock,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			path := tt.path
			if path == "PLACEHOLDER" {
				path = createTestINI(t, iniContent)
			}

			cfg, lock, err := ValidateSectionAndLoad(path, tt.section, tt.mode)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
					if lock != nil {
						_ = lock.Unlock()
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg == nil {
				t.Fatal("cfg is nil")
			}
			_ = lock.Unlock()
		})
	}
}

func TestSaveConfig(t *testing.T) {
	t.Parallel()

	t.Run("saves successfully", func(t *testing.T) {
		t.Parallel()
		path := createTestINI(t, "[db]\nhost = old\n")

		cfg, lock, err := ValidateAndLoad(path, "db", "host", tools.ExclusiveLock)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = lock.Unlock() }()

		cfg.Section("db").Key("host").SetValue("new")
		if err := SaveConfig(cfg, path); err != nil {
			t.Fatalf("SaveConfig() error = %v", err)
		}

		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if got := string(content); !contains(got, "new") {
			t.Errorf("file does not contain 'new': %s", got)
		}
	})

	t.Run("error on invalid path", func(t *testing.T) {
		t.Parallel()
		path := createTestINI(t, "[db]\nk = v\n")

		cfg, lock, err := ValidateAndLoad(path, "db", "k", tools.ExclusiveLock)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = lock.Unlock() }()

		err = SaveConfig(cfg, "/nonexistent/dir/file.ini")
		if err == nil {
			t.Error("expected error for invalid save path")
		}
	})
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
