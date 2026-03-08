package lock

import (
	"fmt"
	"os"
	"syscall"
	"time"
)

// FileLock はファイルベースのアドバイザリロック
type FileLock struct {
	file *os.File
	path string
}

// Acquire はロックファイルを作成・取得する
// timeout で指定した時間内に取得できなければエラーを返す
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

// Release はロックを解放しファイルを閉じる
func (l *FileLock) Release() error {
	if err := syscall.Flock(int(l.file.Fd()), syscall.LOCK_UN); err != nil {
		l.file.Close()
		return fmt.Errorf("unlock %s: %w", l.path, err)
	}
	return l.file.Close()
}
