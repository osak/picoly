package lock

import (
	"fmt"
	"os"
	"syscall"
	"time"
)

// FileLock is a file-based advisory lock using syscall.Flock.
type FileLock struct {
	file *os.File
	path string
}

// Acquire creates and obtains an exclusive lock on the file at dbPath+".lock".
// It returns an error if the lock cannot be obtained within the given timeout.
func Acquire(dbPath string, timeout time.Duration) (*FileLock, error) {
	lockPath := dbPath + ".lock"
	deadline := time.Now().Add(timeout)

	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, fmt.Errorf("open lock file %s: %w", lockPath, err)
	}

	for {
		err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			return &FileLock{file: f, path: lockPath}, nil
		}
		if err != syscall.EWOULDBLOCK {
			f.Close()
			return nil, fmt.Errorf("flock %s: %w", lockPath, err)
		}
		if time.Now().After(deadline) {
			f.Close()
			return nil, fmt.Errorf("timed out waiting for lock on %s after %v", lockPath, timeout)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

// Release unlocks the file and closes it.
func (l *FileLock) Release() error {
	if err := syscall.Flock(int(l.file.Fd()), syscall.LOCK_UN); err != nil {
		l.file.Close()
		return fmt.Errorf("unlock %s: %w", l.path, err)
	}
	return l.file.Close()
}
