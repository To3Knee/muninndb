package main

import (
	"path/filepath"
	"testing"
)

func TestPIDFileWriteRead(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "muninn.pid")

	if err := writePID(path, 12345); err != nil {
		t.Fatalf("writePID: %v", err)
	}
	pid, err := readPID(path)
	if err != nil {
		t.Fatalf("readPID: %v", err)
	}
	if pid != 12345 {
		t.Errorf("pid = %d, want 12345", pid)
	}
}

func TestReadPIDMissingFile(t *testing.T) {
	_, err := readPID("/nonexistent/path/pid")
	if err == nil {
		t.Error("expected error for missing PID file")
	}
}
