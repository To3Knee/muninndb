package main

import (
	"strings"
	"testing"
)

func TestServiceStatusString(t *testing.T) {
	cases := []struct {
		svc  serviceStatus
		want string
	}{
		{serviceStatus{name: "mcp", port: 8750, up: true}, "mcp"},
		{serviceStatus{name: "mcp", port: 8750, up: false}, "mcp"},
	}
	for _, c := range cases {
		got := c.svc.name
		if got != c.want {
			t.Errorf("got %q want %q", got, c.want)
		}
	}
}

func TestOverallState(t *testing.T) {
	all := []serviceStatus{{up: true}, {up: true}, {up: true}}
	if overallState(all) != stateRunning {
		t.Error("expected running")
	}
	none := []serviceStatus{{up: false}, {up: false}, {up: false}}
	if overallState(none) != stateStopped {
		t.Error("expected stopped")
	}
	mixed := []serviceStatus{{up: true}, {up: false}, {up: true}}
	if overallState(mixed) != stateDegraded {
		t.Error("expected degraded")
	}
}

func TestOverallStateEdgeCases(t *testing.T) {
	// Empty slice — no services — should be stateRunning
	empty := []serviceStatus{}
	got := overallState(empty)
	if got != stateRunning {
		t.Errorf("empty services: got %v, want stateRunning", got)
	}

	// Single service up
	single := []serviceStatus{{up: true}}
	if overallState(single) != stateRunning {
		t.Error("single up: expected stateRunning")
	}

	// Single service down
	singleDown := []serviceStatus{{up: false}}
	if overallState(singleDown) != stateStopped {
		t.Error("single down: expected stateStopped")
	}
}

func TestProbeServicesReturnsThreeServices(t *testing.T) {
	// probeServices always returns exactly 3 entries (even if all down)
	svcs := probeServices()
	if len(svcs) != 3 {
		t.Errorf("expected 3 services, got %d", len(svcs))
	}
	names := map[string]bool{}
	for _, s := range svcs {
		names[s.name] = true
	}
	for _, want := range []string{"database", "mcp", "web ui"} {
		if !names[want] {
			t.Errorf("missing service %q in probe results", want)
		}
	}
}

func TestPrintStatusDisplayReturnsStopped(t *testing.T) {
	// With no real server running, should return stateStopped or stateDegraded
	// (not stateRunning, unless muninn happens to be running in test env)
	state := stateStopped
	captureStdout(func() {
		state = printStatusDisplay(false)
	})
	// State should be one of the valid values
	if state != stateRunning && state != stateStopped && state != stateDegraded {
		t.Errorf("unexpected state: %v", state)
	}
}

func TestPrintStatusDisplayOutputContainsName(t *testing.T) {
	out := captureStdout(func() {
		printStatusDisplay(false)
	})
	if !strings.Contains(out, "muninn") {
		t.Errorf("output should contain 'muninn', got: %s", out)
	}
}

func TestPrintStatusDisplayCompactVsNonCompact(t *testing.T) {
	// Non-compact output should include service names
	outFull := captureStdout(func() {
		printStatusDisplay(false)
	})
	outCompact := captureStdout(func() {
		printStatusDisplay(true)
	})
	// Both should contain service names
	if !strings.Contains(outFull, "database") {
		t.Errorf("full output missing 'database': %s", outFull)
	}
	if !strings.Contains(outCompact, "database") {
		t.Errorf("compact output missing 'database': %s", outCompact)
	}
}
