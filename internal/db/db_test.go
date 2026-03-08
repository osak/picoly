package db

import (
	"path/filepath"
	"testing"
)

// NewTestDB はテスト用の一時DBを作成する
func NewTestDB(t *testing.T) *DB {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	d, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open test DB: %v", err)
	}
	t.Cleanup(func() { d.Close() })
	return d
}

func TestOpen(t *testing.T) {
	d := NewTestDB(t)
	if d == nil {
		t.Fatal("expected non-nil DB")
	}
}

func TestOpenIdempotent(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	d1, err := Open(dbPath)
	if err != nil {
		t.Fatalf("first open: %v", err)
	}
	d1.Close()

	d2, err := Open(dbPath)
	if err != nil {
		t.Fatalf("second open: %v", err)
	}
	d2.Close()
}

func TestWALMode(t *testing.T) {
	d := NewTestDB(t)

	var mode string
	row := d.QueryRow("PRAGMA journal_mode")
	if err := row.Scan(&mode); err != nil {
		t.Fatalf("query journal_mode: %v", err)
	}
	if mode != "wal" {
		t.Errorf("journal_mode = %q, want %q", mode, "wal")
	}
}
