package db

import (
	"path/filepath"
	"testing"
)

// NewTestDB creates a temporary in-memory database for use in tests.
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
