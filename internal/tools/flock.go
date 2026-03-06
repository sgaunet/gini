package tools

// LockMode represents the type of file lock to acquire.
type LockMode int

const (
	// SharedLock allows multiple readers but blocks writers.
	SharedLock LockMode = iota
	// ExclusiveLock blocks all other readers and writers.
	ExclusiveLock
)

// FileLock represents an acquired file lock that must be released.
type FileLock struct {
	release func() error
}

// Unlock releases the file lock.
func (fl *FileLock) Unlock() error {
	if fl.release != nil {
		return fl.release()
	}
	return nil
}
