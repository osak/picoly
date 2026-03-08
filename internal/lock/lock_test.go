package lock

import (
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestAcquireRelease(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	l, err := Acquire(dbPath, 5*time.Second)
	if err != nil {
		t.Fatalf("Acquire: %v", err)
	}
	if err := l.Release(); err != nil {
		t.Fatalf("Release: %v", err)
	}

	// Release後に再取得できることを確認
	l2, err := Acquire(dbPath, 5*time.Second)
	if err != nil {
		t.Fatalf("second Acquire: %v", err)
	}
	l2.Release()
}

func TestAcquireTimeout(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	// 最初のロックを取得
	l, err := Acquire(dbPath, 5*time.Second)
	if err != nil {
		t.Fatalf("Acquire: %v", err)
	}
	defer l.Release()

	// 別goroutineから短いtimeoutでロック取得を試みる
	done := make(chan error, 1)
	go func() {
		_, err := Acquire(dbPath, 100*time.Millisecond)
		done <- err
	}()

	select {
	case err := <-done:
		if err == nil {
			t.Error("expected timeout error, got nil")
		}
	case <-time.After(2 * time.Second):
		t.Error("goroutine did not return in time")
	}
}

func TestMutualExclusion(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	var mu sync.Mutex
	counter := 0
	const goroutines = 5

	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			l, err := Acquire(dbPath, 10*time.Second)
			if err != nil {
				t.Errorf("Acquire: %v", err)
				return
			}
			defer l.Release()

			// クリティカルセクション
			mu.Lock()
			counter++
			mu.Unlock()
		}()
	}
	wg.Wait()

	if counter != goroutines {
		t.Errorf("counter = %d, want %d", counter, goroutines)
	}
}
