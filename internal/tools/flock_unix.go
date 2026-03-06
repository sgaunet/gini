//go:build !windows

package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// LockFile acquires an advisory file lock on the given INI file path.
// Use SharedLock for read operations and ExclusiveLock for write operations.
// The returned FileLock must be released by calling Unlock.
func LockFile(iniPath string, mode LockMode) (*FileLock, error) {
	lockPath := filepath.Clean(iniPath) + ".lock"

	const lockFilePerms = 0o600
	// #nosec G304 - lock file path derived from user-provided INI path
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, lockFilePerms)
	if err != nil {
		return nil, fmt.Errorf("failed to open lock file: %w", err)
	}

	how := syscall.LOCK_SH
	if mode == ExclusiveLock {
		how = syscall.LOCK_EX
	}

	fd := int(f.Fd()) // #nosec G115 - file descriptor fits in int
	if err := syscall.Flock(fd, how); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("failed to acquire file lock: %w", err)
	}

	return &FileLock{
		release: func() error {
			lockErr := syscall.Flock(fd, syscall.LOCK_UN)
			closeErr := f.Close()
			if lockErr != nil {
				return fmt.Errorf("failed to release lock: %w", lockErr)
			}
			if closeErr != nil {
				return fmt.Errorf("failed to close lock file: %w", closeErr)
			}
			return nil
		},
	}, nil
}
