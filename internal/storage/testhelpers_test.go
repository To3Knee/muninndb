package storage

import (
	"os"
	"testing"

	"github.com/cockroachdb/pebble"
)

// openTestPebble opens a Pebble DB in a temp directory and registers t.Cleanup to close it.
func openTestPebble(t *testing.T) *pebble.DB {
	t.Helper()
	dir, err := os.MkdirTemp("", "muninndb-test-*")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	db, err := OpenPebble(dir, DefaultOptions())
	if err != nil {
		os.RemoveAll(dir)
		t.Fatalf("open pebble: %v", err)
	}
	t.Cleanup(func() {
		db.Close()
		os.RemoveAll(dir)
	})
	return db
}
