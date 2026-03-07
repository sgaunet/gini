//go:build !windows

package tools

import (
	"path/filepath"
	"testing"
)

func TestLockFile(t *testing.T) {
	t.Parallel()

	t.Run("shared lock", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		iniPath := filepath.Join(tmpDir, "test.ini")

		lock, err := LockFile(iniPath, SharedLock)
		if err != nil {
			t.Fatalf("LockFile(SharedLock) error = %v", err)
		}
		if err := lock.Unlock(); err != nil {
			t.Fatalf("Unlock() error = %v", err)
		}
	})

	t.Run("exclusive lock", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		iniPath := filepath.Join(tmpDir, "test.ini")

		lock, err := LockFile(iniPath, ExclusiveLock)
		if err != nil {
			t.Fatalf("LockFile(ExclusiveLock) error = %v", err)
		}
		if err := lock.Unlock(); err != nil {
			t.Fatalf("Unlock() error = %v", err)
		}
	})

	t.Run("multiple shared locks", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		iniPath := filepath.Join(tmpDir, "test.ini")

		lock1, err := LockFile(iniPath, SharedLock)
		if err != nil {
			t.Fatalf("first LockFile() error = %v", err)
		}
		lock2, err := LockFile(iniPath, SharedLock)
		if err != nil {
			t.Fatalf("second LockFile() error = %v", err)
		}

		if err := lock1.Unlock(); err != nil {
			t.Fatalf("Unlock lock1 error = %v", err)
		}
		if err := lock2.Unlock(); err != nil {
			t.Fatalf("Unlock lock2 error = %v", err)
		}
	})

	t.Run("invalid path", func(t *testing.T) {
		t.Parallel()
		_, err := LockFile("/nonexistent/dir/test.ini", SharedLock)
		if err == nil {
			t.Error("LockFile() expected error for invalid path")
		}
	})
}

func TestFileLock_Unlock_NilRelease(t *testing.T) {
	t.Parallel()
	fl := &FileLock{}
	if err := fl.Unlock(); err != nil {
		t.Errorf("Unlock() with nil release = %v, want nil", err)
	}
}
