//go:build integration

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// muninnBin is the path to the built binary, set by TestMain.
var muninnBin string

// TestMain builds the muninn binary once and runs all integration tests.
// It exits 0 without running tests if muninn is already listening on :8750,
// since these tests require exclusive use of that port.
func TestMain(m *testing.M) {
	// Guard: skip if something is already on the MCP port.
	if resp, err := http.Get("http://localhost:8750/mcp/health"); err == nil {
		resp.Body.Close()
		fmt.Fprintln(os.Stderr, "integration: muninn already running on :8750 — stop it first")
		os.Exit(0)
	}

	// Build the binary into a temp file.
	tmp, err := os.CreateTemp("", "muninn-integ-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "integration: create temp: %v\n", err)
		os.Exit(1)
	}
	tmp.Close()
	muninnBin = tmp.Name()

	out, err := exec.Command("go", "build", "-o", muninnBin, ".").CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "integration: build failed: %v\n%s\n", err, out)
		os.Remove(muninnBin)
		os.Exit(1)
	}

	code := m.Run()
	os.Remove(muninnBin)
	os.Exit(code)
}

// muninnCmd creates a command with MUNINNDB_DATA set to the given directory.
func muninnCmd(dataDir string, args ...string) *exec.Cmd {
	cmd := exec.Command(muninnBin, args...)
	cmd.Env = append(os.Environ(), "MUNINNDB_DATA="+dataDir)
	return cmd
}

// waitForHealth polls localhost:8750/mcp/health until 200 or timeout.
func waitForHealth(timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get("http://localhost:8750/mcp/health")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			return true
		}
		time.Sleep(200 * time.Millisecond)
	}
	return false
}

// waitForDead polls until localhost:8750 refuses connections (port is free).
func waitForDead(timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get("http://localhost:8750/mcp/health")
		if err != nil {
			return true // connection refused — port is free
		}
		resp.Body.Close()
		time.Sleep(100 * time.Millisecond)
	}
	return false
}

// startDaemon runs `muninn start`, waits for the health endpoint, and
// registers a cleanup that stops the daemon and waits for the port to free.
func startDaemon(t *testing.T, dataDir string) {
	t.Helper()
	out, err := muninnCmd(dataDir, "start").CombinedOutput()
	if err != nil {
		t.Fatalf("muninn start: %v\n%s", err, out)
	}
	if !waitForHealth(10 * time.Second) {
		t.Fatal("daemon did not become healthy within 10s")
	}
	t.Cleanup(func() {
		muninnCmd(dataDir, "stop").Run() //nolint:errcheck — ignore if already stopped
		waitForDead(5 * time.Second)    // ensure port is free before the next test
	})
}

// TestHelpExitsZero verifies that `muninn help` exits 0 and produces output.
func TestHelpExitsZero(t *testing.T) {
	out, err := exec.Command(muninnBin, "help").CombinedOutput()
	if err != nil {
		t.Fatalf("muninn help: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "muninn") {
		t.Errorf("help output missing 'muninn':\n%s", out)
	}
}

// TestUnknownCommandExitsOne verifies that unknown subcommands exit non-zero
// and print a helpful error.
func TestUnknownCommandExitsOne(t *testing.T) {
	cmd := exec.Command(muninnBin, "boguscommand-xyz")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected non-zero exit for unknown command, got success\n%s", out)
	}
	if !strings.Contains(string(out), "Unknown command") {
		t.Errorf("expected 'Unknown command' in output:\n%s", out)
	}
}

// TestInitNoStart verifies that `muninn init --yes --no-start --no-token` exits
// 0 and does not start a daemon.
func TestInitNoStart(t *testing.T) {
	dataDir := t.TempDir()
	out, err := muninnCmd(dataDir, "init", "--yes", "--no-start", "--no-token").CombinedOutput()
	if err != nil {
		t.Fatalf("muninn init --yes --no-start --no-token: %v\n%s", err, out)
	}
	// Port 8750 must still be closed.
	resp, hErr := http.Get("http://localhost:8750/mcp/health")
	if hErr == nil {
		resp.Body.Close()
		t.Error("daemon should not be running after --no-start, but :8750 responded")
	}
}

// TestStartStop verifies the full start → health check → stop lifecycle.
func TestStartStop(t *testing.T) {
	dataDir := t.TempDir()
	startDaemon(t, dataDir)

	// PID file must exist while the daemon is running.
	pidPath := filepath.Join(dataDir, "muninn.pid")
	if _, err := os.Stat(pidPath); err != nil {
		t.Errorf("PID file missing at %s: %v", pidPath, err)
	}

	// `muninn status` must exit 0 and eventually report "running".
	// runStart only waits for the MCP port (8750); the REST API (8475) used
	// by the database probe may still be warming up. Poll until settled.
	var statusOut string
	statusDeadline := time.Now().Add(8 * time.Second)
	for time.Now().Before(statusDeadline) {
		out, err := muninnCmd(dataDir, "status").CombinedOutput()
		if err != nil {
			t.Errorf("muninn status returned non-zero: %v\n%s", err, out)
			break
		}
		statusOut = string(out)
		if strings.Contains(statusOut, "running") {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	if !strings.Contains(statusOut, "running") {
		t.Errorf("status never reached 'running' (last output):\n%s", statusOut)
	}
}

// TestRestartNoRace verifies that `muninn restart` cleanly stops the old
// process before starting a new one (no "address already in use" race).
func TestRestartNoRace(t *testing.T) {
	dataDir := t.TempDir()
	startDaemon(t, dataDir)

	// `muninn restart` should stop, then start, then return healthy.
	out, err := muninnCmd(dataDir, "restart").CombinedOutput()
	if err != nil {
		t.Fatalf("muninn restart: %v\n%s", err, out)
	}

	// Confirm the daemon is healthy after the restart.
	if !waitForHealth(10 * time.Second) {
		t.Fatal("daemon did not become healthy after restart")
	}
	// Cleanup registered by startDaemon will stop the restarted daemon.
}

// mcpTool sends a JSON-RPC 2.0 tools/call request to the local MCP server and
// returns the parsed text payload. token may be empty if auth is not configured.
func mcpTool(t *testing.T, token, toolName string, args map[string]any) map[string]any {
	t.Helper()
	body, _ := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      toolName,
			"arguments": args,
		},
	})
	req, _ := http.NewRequest("POST", "http://localhost:8750/mcp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("mcpTool %s: %v", toolName, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("mcpTool %s: HTTP %d", toolName, resp.StatusCode)
	}
	rawBody, _ := io.ReadAll(resp.Body)
	// result is []{"type":"text","text":"<json>"} (textContent envelope)
	var rpcResp struct {
		Result []struct {
			Text string `json:"text"`
		} `json:"result"`
		Error *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(bytes.NewReader(rawBody)).Decode(&rpcResp); err != nil {
		t.Fatalf("mcpTool %s: decode: %v", toolName, err)
	}
	if rpcResp.Error != nil {
		t.Fatalf("mcpTool %s: RPC error %d: %s", toolName, rpcResp.Error.Code, rpcResp.Error.Message)
	}
	if len(rpcResp.Result) == 0 {
		t.Fatalf("mcpTool %s: empty result array", toolName)
	}
	var result map[string]any
	if err := json.Unmarshal([]byte(rpcResp.Result[0].Text), &result); err != nil {
		t.Fatalf("mcpTool %s: parse text payload: %v", toolName, err)
	}
	return result
}

// TestInitAndStart verifies that `muninn init --yes --no-token` (the first-run
// path without interactive prompts) successfully starts the daemon and all three
// services become healthy. This is the most important gap in first-time user
// confidence: init delegates to runStart internally, so if that path is broken
// the very first command a new user runs fails.
func TestInitAndStart(t *testing.T) {
	dataDir := t.TempDir()
	out, err := muninnCmd(dataDir, "init", "--yes", "--no-token").CombinedOutput()
	if err != nil {
		t.Fatalf("muninn init --yes --no-token: %v\n%s", err, out)
	}
	// init calls runStart internally and blocks until the MCP port is up, so
	// health should be nearly immediate — but give the REST API (8475) a moment.
	if !waitForHealth(10 * time.Second) {
		t.Fatal("daemon did not become healthy within 10s after init")
	}
	t.Cleanup(func() {
		muninnCmd(dataDir, "stop").Run() //nolint:errcheck
		waitForDead(5 * time.Second)
	})

	// Confirm status reports "running" (all 3 services up).
	var statusOut string
	deadline := time.Now().Add(8 * time.Second)
	for time.Now().Before(deadline) {
		out, err := muninnCmd(dataDir, "status").CombinedOutput()
		if err != nil {
			t.Errorf("muninn status returned non-zero: %v\n%s", err, out)
			break
		}
		statusOut = string(out)
		if strings.Contains(statusOut, "running") {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if !strings.Contains(statusOut, "running") {
		t.Errorf("status never reached 'running' after init (last output):\n%s", statusOut)
	}
}

// TestMCPRoundTrip verifies the core product path: store a memory via the MCP
// tool, then fetch it back by ID and confirm the content is preserved.
// This test exercises the full stack: MCP server → engine → storage → retrieval.
func TestMCPRoundTrip(t *testing.T) {
	dataDir := t.TempDir()
	startDaemon(t, dataDir)

	// Use whatever token the daemon was started with (readTokenFile reads the
	// same path as runStart, so they are always consistent).
	tok := readTokenFile()

	// Store a memory.
	writeResult := mcpTool(t, tok, "muninn_remember", map[string]any{
		"vault":   "default",
		"content": "integration test memory — hello from TestMCPRoundTrip",
		"concept": "integration test",
	})
	id, ok := writeResult["id"].(string)
	if !ok || id == "" {
		t.Fatalf("muninn_remember: expected non-empty id, got: %v", writeResult)
	}

	// Fetch it back by ID.
	readResult := mcpTool(t, tok, "muninn_read", map[string]any{
		"vault": "default",
		"id":    id,
	})
	content, _ := readResult["content"].(string)
	if content == "" {
		t.Fatalf("muninn_read: content field missing or empty: %v", readResult)
	}
	const want = "integration test memory"
	if !strings.Contains(content, want) {
		t.Errorf("muninn_read: content %q does not contain %q", content, want)
	}
}
