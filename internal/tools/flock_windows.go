//go:build windows

package tools

// LockFile on Windows is a no-op. Advisory file locking is only supported on
// Unix systems via syscall.Flock. Concurrent access on Windows is not protected.
func LockFile(_ string, _ LockMode) (*FileLock, error) {
	return &FileLock{}, nil
}
