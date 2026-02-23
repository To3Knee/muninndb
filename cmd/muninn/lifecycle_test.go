package main

import (
	"os"
	"strings"
	"testing"
)

// TestIsProcessRunningCurrentProcess checks if the current process is identified as running.
func TestIsProcessRunningCurrentProcess(t *testing.T) {
	pid := os.Getpid()
	if !isProcessRunning(pid) {
		t.Errorf("current process (pid %d) should be running", pid)
	}
}

// TestIsProcessRunningDeadProcess checks if a non-existent PID is correctly identified as not running.
func TestIsProcessRunningDeadProcess(t *testing.T) {
	// PID 99999999 almost certainly doesn't exist
	if isProcessRunning(99999999) {
		t.Error("pid 99999999 should not be running")
	}
}

// TestIsProcessRunningNegativePID checks that negative PIDs are handled gracefully.
func TestIsProcessRunningNegativePID(t *testing.T) {
	// Negative PID — should not panic, should return false
	if isProcessRunning(-1) {
		t.Error("negative pid should not be running")
	}
}

// TestIsProcessRunningZeroPID checks that PID 0 is handled correctly.
func TestIsProcessRunningZeroPID(t *testing.T) {
	// PID 0 is special — should return false
	if isProcessRunning(0) {
		t.Error("pid 0 should not be running")
	}
}

// TestDefaultDataDir checks that defaultDataDir returns a valid path under the home directory.
func TestDefaultDataDir(t *testing.T) {
	dir := defaultDataDir()
	if dir == "" {
		t.Error("defaultDataDir returned empty string")
	}
	home, _ := os.UserHomeDir()
	if !strings.HasPrefix(dir, home) {
		t.Errorf("defaultDataDir %q should be under home %q", dir, home)
	}
	if !strings.HasSuffix(dir, "data") {
		t.Errorf("defaultDataDir %q should end with 'data'", dir)
	}
}

// TestDefaultDataDirEnvOverride checks that MUNINNDB_DATA environment variable is respected.
func TestDefaultDataDirEnvOverride(t *testing.T) {
	oldVal := os.Getenv("MUNINNDB_DATA")
	defer os.Setenv("MUNINNDB_DATA", oldVal)

	testDir := "/tmp/test-muninn-data"
	os.Setenv("MUNINNDB_DATA", testDir)

	dir := defaultDataDir()
	if dir != testDir {
		t.Errorf("defaultDataDir = %q, want %q (from MUNINNDB_DATA)", dir, testDir)
	}
}
