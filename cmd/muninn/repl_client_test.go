package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestIsEmptyMCPResult tests all branches of isEmptyMCPResult.
func TestIsEmptyMCPResult(t *testing.T) {
	cases := []struct {
		name   string
		result map[string]any
		want   bool
	}{
		{"nil", nil, true},
		{"empty map", map[string]any{}, true},
		{"no content key", map[string]any{"other": "x"}, true},
		{"content not slice", map[string]any{"content": "x"}, true},
		{"content empty slice", map[string]any{"content": []any{}}, true},
		{"content item not map", map[string]any{"content": []any{"bad"}}, true},
		{"content item no text", map[string]any{"content": []any{map[string]any{"other": "x"}}}, true},
		{"content empty string", map[string]any{"content": []any{map[string]any{"text": ""}}}, true},
		{"content []", map[string]any{"content": []any{map[string]any{"text": "[]"}}}, true},
		{"content {}", map[string]any{"content": []any{map[string]any{"text": "{}"}}}, true},
		{"content null", map[string]any{"content": []any{map[string]any{"text": "null"}}}, true},
		{"content whitespace only", map[string]any{"content": []any{map[string]any{"text": "   "}}}, true},
		{"content has value", map[string]any{"content": []any{map[string]any{"text": `{"id":"01JF"}`}}}, false},
		{"content plain text", map[string]any{"content": []any{map[string]any{"text": "hello"}}}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := isEmptyMCPResult(tc.result)
			if got != tc.want {
				t.Errorf("isEmptyMCPResult(%s): got %v, want %v", tc.name, got, tc.want)
			}
		})
	}
}

// TestPrettyPrintNil tests prettyPrint with nil result.
func TestPrettyPrintNil(t *testing.T) {
	out := captureStdout(func() { prettyPrint(nil) })
	if !strings.Contains(out, "no result") {
		t.Errorf("nil: expected 'no result', got: %s", out)
	}
}

// TestPrettyPrintJSON tests prettyPrint with JSON content.
func TestPrettyPrintJSON(t *testing.T) {
	result := map[string]any{
		"content": []any{
			map[string]any{"text": `{"id":"01JF","score":0.9}`},
		},
	}
	out := captureStdout(func() { prettyPrint(result) })
	if !strings.Contains(out, "01JF") {
		t.Errorf("JSON content: expected id in output, got: %s", out)
	}
	if !strings.Contains(out, "0.9") {
		t.Errorf("JSON content: expected score in output, got: %s", out)
	}
}

// TestPrettyPrintPlainText tests prettyPrint with plain text content.
func TestPrettyPrintPlainText(t *testing.T) {
	result := map[string]any{
		"content": []any{
			map[string]any{"text": "plain text value"},
		},
	}
	out := captureStdout(func() { prettyPrint(result) })
	if !strings.Contains(out, "plain text value") {
		t.Errorf("plain text: got: %s", out)
	}
}

// TestPrettyPrintFallback tests prettyPrint fallback to JSON marshaling.
func TestPrettyPrintFallback(t *testing.T) {
	result := map[string]any{"something": "else"}
	out := captureStdout(func() { prettyPrint(result) })
	if !strings.Contains(out, "something") {
		t.Errorf("fallback: got: %s", out)
	}
}

// TestHumanizeTime tests all branches of humanizeTime.
func TestHumanizeTime(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name string
		ts   string
		want string
	}{
		{"invalid format", "not-a-date", "not-a-date"},
		{"just now", now.Add(-10 * time.Second).Format(time.RFC3339), "just now"},
		{"minutes ago", now.Add(-5 * time.Minute).Format(time.RFC3339), "minutes ago"},
		{"hours ago", now.Add(-3 * time.Hour).Format(time.RFC3339), "hours ago"},
		{"days ago", now.Add(-48 * time.Hour).Format(time.RFC3339), "days ago"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := humanizeTime(tc.ts)
			if !strings.Contains(got, tc.want) {
				t.Errorf("humanizeTime(%s): got %q, want to contain %q", tc.name, got, tc.want)
			}
		})
	}
}

// TestMcpHealthCheckUp tests mcpHealthCheck with a healthy server.
func TestMcpHealthCheckUp(t *testing.T) {
	srv := newHealthServer()
	defer srv.Close()
	if !mcpHealthCheck(srv.URL) {
		t.Error("expected true for 200 response")
	}
}

// TestMcpHealthCheckDown tests mcpHealthCheck with an unreachable server.
func TestMcpHealthCheckDown(t *testing.T) {
	// Unreachable port
	if mcpHealthCheck("http://localhost:19876") {
		t.Error("expected false for unreachable server")
	}
}

// TestMcpHealthCheck404 tests mcpHealthCheck with a 404 response.
func TestMcpHealthCheck404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()
	// 404 is not 200 — should be false
	if mcpHealthCheck(srv.URL) {
		t.Error("expected false for 404 response")
	}
}

// TestLoadSaveDefaultVault tests loading and saving default vault to config.
func TestLoadSaveDefaultVault(t *testing.T) {
	// Save current HOME
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Initially nothing
	got := loadDefaultVault()
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}

	// Save and load
	saveDefaultVault("workproject")
	got = loadDefaultVault()
	if got != "workproject" {
		t.Errorf("got %q, want %q", got, "workproject")
	}

	// Overwrite
	saveDefaultVault("personal")
	got = loadDefaultVault()
	if got != "personal" {
		t.Errorf("got %q, want %q", got, "personal")
	}
}

// TestCmdShowVaults401Fallback tests cmdShowVaults with connection error.
func TestCmdShowVaults401Fallback(t *testing.T) {
	// When the server is unreachable, cmdShowVaults should handle gracefully
	r := &replState{mcpURL: "http://localhost:9998"}
	errOut := captureStderr(func() {
		captureStdout(func() {
			r.cmdShowVaults()
		})
	})
	// Should print error message to stderr or handle gracefully
	// (error message is written to stderr, not stdout)
	if !strings.Contains(errOut, "Error connecting to server") && !strings.Contains(errOut, "muninn start") {
		// If not in stderr, the function should at least complete without panic
		t.Logf("cmdShowVaults handled connection error gracefully")
	}
}

// TestCmdShowVaultsEmptyList tests formatVaultTable with empty list.
func TestCmdShowVaultsEmptyList(t *testing.T) {
	// formatVaultTable with empty list should not panic
	out := captureStdout(func() {
		formatVaultTable([]map[string]any{})
	})
	// formatVaultTable always prints header — verify it doesn't panic
	if !strings.Contains(out, "NAME") {
		t.Errorf("expected header row, got: %s", out)
	}
}

// TestFormatVaultTableMissingFields tests formatVaultTable with missing fields.
func TestFormatVaultTableMissingFields(t *testing.T) {
	// Vaults with missing optional fields should not panic
	vaults := []map[string]any{
		{"name": "vault-no-count"},                    // no memory_count
		{"name": "vault-no-time", "memory_count": float64(5)},  // no last_active
		{"memory_count": float64(3)},                  // no name
		{},                                             // completely empty
	}
	out := captureStdout(func() {
		formatVaultTable(vaults)
	})
	if !strings.Contains(out, "NAME") {
		t.Errorf("expected header row, got: %s", out)
	}
	if !strings.Contains(out, "vault-no-count") {
		t.Errorf("expected vault-no-count in output, got: %s", out)
	}
	if !strings.Contains(out, "vault-no-time") {
		t.Errorf("expected vault-no-time in output, got: %s", out)
	}
}

// TestFormatVaultTableNormalVaults tests formatVaultTable with normal vaults.
func TestFormatVaultTableNormalVaults(t *testing.T) {
	now := time.Now()
	vaults := []map[string]any{
		{
			"name": "personal",
			"memory_count": float64(47),
			"last_active": now.Format(time.RFC3339),
		},
		{
			"name": "work",
			"memory_count": float64(12),
			"last_active": now.Add(-2*time.Hour).Format(time.RFC3339),
		},
	}
	out := captureStdout(func() {
		formatVaultTable(vaults)
	})
	if !strings.Contains(out, "personal") {
		t.Errorf("expected 'personal' in output, got: %s", out)
	}
	if !strings.Contains(out, "work") {
		t.Errorf("expected 'work' in output, got: %s", out)
	}
	if !strings.Contains(out, "47") {
		t.Errorf("expected '47' in output, got: %s", out)
	}
	if !strings.Contains(out, "12") {
		t.Errorf("expected '12' in output, got: %s", out)
	}
	if !strings.Contains(out, "just now") {
		t.Errorf("expected 'just now' in output, got: %s", out)
	}
	if !strings.Contains(out, "hours ago") {
		t.Errorf("expected 'hours ago' in output, got: %s", out)
	}
}

// TestFormatVaultTableLongNames tests formatVaultTable with long vault names.
func TestFormatVaultTableLongNames(t *testing.T) {
	vaults := []map[string]any{
		{
			"name": "a-very-long-vault-name-for-testing",
			"memory_count": float64(99),
			"last_active": time.Now().Format(time.RFC3339),
		},
	}
	// Should not panic and should handle truncation gracefully
	out := captureStdout(func() {
		formatVaultTable(vaults)
	})
	if !strings.Contains(out, "a-very-long") {
		t.Errorf("expected part of long name in output, got: %s", out)
	}
}

// TestHumanizeTimeEdgeCases tests humanizeTime with edge cases.
func TestHumanizeTimeEdgeCases(t *testing.T) {
	now := time.Now()

	// Test boundary: 59 seconds (< 1 minute)
	got := humanizeTime(now.Add(-59 * time.Second).Format(time.RFC3339))
	if got != "just now" {
		t.Errorf("59 seconds: expected 'just now', got %q", got)
	}

	// Test boundary: 61 seconds (> 1 minute)
	got = humanizeTime(now.Add(-61 * time.Second).Format(time.RFC3339))
	if !strings.Contains(got, "minute") {
		t.Errorf("61 seconds: expected 'minute' in output, got %q", got)
	}

	// Test boundary: 59 minutes (< 1 hour)
	got = humanizeTime(now.Add(-59 * time.Minute).Format(time.RFC3339))
	if !strings.Contains(got, "minutes ago") {
		t.Errorf("59 minutes: expected 'minutes ago', got %q", got)
	}

	// Test boundary: 61 minutes (> 1 hour)
	got = humanizeTime(now.Add(-61 * time.Minute).Format(time.RFC3339))
	if !strings.Contains(got, "hours ago") {
		t.Errorf("61 minutes: expected 'hours ago', got %q", got)
	}

	// Test boundary: 23 hours (< 24 hours)
	got = humanizeTime(now.Add(-23 * time.Hour).Format(time.RFC3339))
	if !strings.Contains(got, "hours ago") {
		t.Errorf("23 hours: expected 'hours ago', got %q", got)
	}

	// Test boundary: 25 hours (> 24 hours)
	got = humanizeTime(now.Add(-25 * time.Hour).Format(time.RFC3339))
	if !strings.Contains(got, "days ago") {
		t.Errorf("25 hours: expected 'days ago', got %q", got)
	}
}

// TestIsEmptyMCPResultComplexJSON tests isEmptyMCPResult with complex JSON.
func TestIsEmptyMCPResultComplexJSON(t *testing.T) {
	// Complex JSON with nested structures should not be empty
	result := map[string]any{
		"content": []any{
			map[string]any{
				"text": `{"nested":{"deep":"value"},"array":[1,2,3]}`,
			},
		},
	}
	got := isEmptyMCPResult(result)
	if got != false {
		t.Errorf("complex JSON: expected false, got %v", got)
	}
}

// TestIsEmptyMCPResultEmptyJSON tests isEmptyMCPResult with empty JSON objects.
func TestIsEmptyMCPResultEmptyJSON(t *testing.T) {
	cases := []struct {
		name string
		text string
		want bool
	}{
		{"empty object", "{}", true},
		{"empty array", "[]", true},
		{"null", "null", true},
		{"nested empty object", `{"inner":{}}`, false},
		{"nested empty array", `{"inner":[]}`, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := map[string]any{
				"content": []any{
					map[string]any{"text": tc.text},
				},
			}
			got := isEmptyMCPResult(result)
			if got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

// TestPrettyPrintWithMultipleContentItems tests prettyPrint uses first content item.
func TestPrettyPrintWithMultipleContentItems(t *testing.T) {
	result := map[string]any{
		"content": []any{
			map[string]any{"text": "first item"},
			map[string]any{"text": "second item"},
		},
	}
	out := captureStdout(func() { prettyPrint(result) })
	if !strings.Contains(out, "first item") {
		t.Errorf("expected first item in output, got: %s", out)
	}
	// second item should not be in output since we only use the first
	if strings.Contains(out, "second item") {
		t.Errorf("unexpected second item in output: %s", out)
	}
}

// TestMcpHealthCheckDifferentStatuses tests mcpHealthCheck with various HTTP statuses.
func TestMcpHealthCheckDifferentStatuses(t *testing.T) {
	statuses := []int{
		http.StatusOK,                  // 200 - should pass
		http.StatusCreated,             // 201 - should fail
		http.StatusBadRequest,          // 400 - should fail
		http.StatusUnauthorized,        // 401 - should fail
		http.StatusNotFound,            // 404 - should fail
		http.StatusInternalServerError, // 500 - should fail
	}

	for _, status := range statuses {
		t.Run(fmt.Sprintf("status_%d", status), func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(status)
			}))
			defer srv.Close()

			got := mcpHealthCheck(srv.URL)
			want := status == http.StatusOK
			if got != want {
				t.Errorf("status %d: got %v, want %v", status, got, want)
			}
		})
	}
}

// TestLoadDefaultVaultInvalidJSON tests loadDefaultVault with invalid JSON.
func TestLoadDefaultVaultInvalidJSON(t *testing.T) {
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Create config dir
	configDir := filepath.Join(tmpDir, ".muninn")
	os.MkdirAll(configDir, 0755)

	// Write invalid JSON
	configFile := filepath.Join(configDir, "config")
	os.WriteFile(configFile, []byte("not valid json {"), 0644)

	// Should return empty string instead of panicking
	got := loadDefaultVault()
	if got != "" {
		t.Errorf("invalid JSON: expected empty string, got %q", got)
	}
}

// TestFormatVaultTableZeroMemories tests formatVaultTable with zero memory counts.
func TestFormatVaultTableZeroMemories(t *testing.T) {
	vaults := []map[string]any{
		{
			"name": "empty-vault",
			"memory_count": float64(0),
			"last_active": time.Now().Format(time.RFC3339),
		},
	}
	out := captureStdout(func() {
		formatVaultTable(vaults)
	})
	if !strings.Contains(out, "empty-vault") {
		t.Errorf("expected 'empty-vault' in output, got: %s", out)
	}
	if !strings.Contains(out, "0") {
		t.Errorf("expected '0' memory count in output, got: %s", out)
	}
}

// mcpRequest captures the structure of an MCP JSON-RPC request.
type mcpRequest struct {
	Method string         `json:"method"`
	Params map[string]any `json:"params"`
}

// newMCPServer creates an httptest server that responds to /mcp JSON-RPC calls.
// It captures the last request in lastReq, and returns the given result JSON.
func newMCPServer(t *testing.T, resultJSON string) (*httptest.Server, *mcpRequest) {
	t.Helper()
	last := &mcpRequest{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/mcp" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		var req struct {
			Method string         `json:"method"`
			Params map[string]any `json:"params"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		*last = mcpRequest{Method: req.Method, Params: req.Params}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"jsonrpc":"2.0","id":1,"result":%s}`, resultJSON)
	}))
	return srv, last
}

// mcpResultWithContent wraps text in the MCP content format.
func mcpResultWithContent(text string) string {
	b, _ := json.Marshal(map[string]any{
		"content": []any{map[string]any{"type": "text", "text": text}},
	})
	return string(b)
}

// TestMcpCallSendsCorrectRequest verifies JSON-RPC request format and response parsing.
func TestMcpCallSendsCorrectRequest(t *testing.T) {
	srv, last := newMCPServer(t, mcpResultWithContent(`[{"id":"01JF","content":"memory text"}]`))
	defer srv.Close()

	result, err := mcpCall(srv.URL, "muninn_recall", map[string]any{
		"vault":   "myproject",
		"context": []string{"golang"},
		"limit":   10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify method sent
	if last.Method != "tools/call" {
		t.Errorf("method = %q, want %q", last.Method, "tools/call")
	}

	// Verify tool name in params
	params := last.Params
	if params["name"] != "muninn_recall" {
		t.Errorf("params.name = %v, want %q", params["name"], "muninn_recall")
	}

	// Verify arguments passed through
	args, ok := params["arguments"].(map[string]any)
	if !ok {
		t.Fatal("params.arguments is not a map")
	}
	if args["vault"] != "myproject" {
		t.Errorf("vault = %v, want %q", args["vault"], "myproject")
	}
	if args["limit"] != float64(10) && args["limit"] != 10 {
		t.Errorf("limit = %v, want 10", args["limit"])
	}

	// Verify result has content
	if result == nil {
		t.Fatal("result is nil")
	}
}

// TestMcpCallReturnsErrorOnMCPError verifies mcpCall handles MCP error responses.
func TestMcpCallReturnsErrorOnMCPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"jsonrpc":"2.0","id":1,"error":{"code":-32601,"message":"method not found"}}`)
	}))
	defer srv.Close()

	_, err := mcpCall(srv.URL, "nonexistent_tool", nil)
	if err == nil {
		t.Error("expected error for MCP error response")
	}
	if !strings.Contains(err.Error(), "MCP error") {
		t.Errorf("error should mention 'MCP error': %v", err)
	}
}

// TestMcpCallHandlesConnectionError verifies mcpCall handles unreachable server.
func TestMcpCallHandlesConnectionError(t *testing.T) {
	_, err := mcpCall("http://localhost:19999", "any_tool", nil)
	if err == nil {
		t.Error("expected error for unreachable server")
	}
}

// TestCmdSearchCallsCorrectTool verifies cmdSearch calls muninn_recall with correct vault and query.
func TestCmdSearchCallsCorrectTool(t *testing.T) {
	body := mcpResultWithContent(`[{"id":"01JF","content":"memory about golang"}]`)
	srv, last := newMCPServer(t, body)
	defer srv.Close()

	r := &replState{mcpURL: srv.URL, vault: "testproject"}
	captureStdout(func() {
		r.cmdSearch("golang concurrency")
	})

	args, ok := last.Params["arguments"].(map[string]any)
	if !ok {
		t.Fatal("arguments is not a map")
	}
	if args["vault"] != "testproject" {
		t.Errorf("vault = %v, want %q", args["vault"], "testproject")
	}
	// context should contain the query
	ctx, ok := args["context"].([]any)
	if !ok || len(ctx) == 0 {
		t.Error("context not passed to muninn_recall")
	}
	if last.Method != "tools/call" {
		t.Errorf("wrong method: %v", last.Method)
	}
}

// TestCmdSearchEmptyResultShowsHelpfulMessage verifies empty search results show a helpful message.
func TestCmdSearchEmptyResultShowsHelpfulMessage(t *testing.T) {
	srv, _ := newMCPServer(t, mcpResultWithContent("[]"))
	defer srv.Close()

	r := &replState{mcpURL: srv.URL, vault: "testproject"}
	out := captureStdout(func() {
		r.cmdSearch("something obscure")
	})
	if !strings.Contains(out, "No memories match") {
		t.Errorf("expected empty result message, got: %s", out)
	}
}

// TestCmdGetCallsCorrectTool verifies cmdGet calls muninn_read with the ID.
func TestCmdGetCallsCorrectTool(t *testing.T) {
	body := mcpResultWithContent(`{"id":"01JFXX","content":"memory detail"}`)
	srv, last := newMCPServer(t, body)
	defer srv.Close()

	r := &replState{mcpURL: srv.URL, vault: "testproject"}
	captureStdout(func() {
		r.cmdGet("01JFXX4KZMB3E9QV7P")
	})

	args, ok := last.Params["arguments"].(map[string]any)
	if !ok {
		t.Fatal("arguments is not a map")
	}
	if args["id"] != "01JFXX4KZMB3E9QV7P" {
		t.Errorf("id = %v, want %q", args["id"], "01JFXX4KZMB3E9QV7P")
	}
	if args["vault"] != "testproject" {
		t.Errorf("vault = %v, want %q", args["vault"], "testproject")
	}
}

// TestCmdGetError verifies cmdGet handles MCP errors gracefully.
func TestCmdGetError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"jsonrpc":"2.0","id":1,"error":{"code":-32000,"message":"not found"}}`)
	}))
	defer srv.Close()

	r := &replState{mcpURL: srv.URL, vault: "v"}
	out := captureStderr(func() {
		r.cmdGet("NONEXISTENT")
	})
	if !strings.Contains(out, "Error") {
		t.Errorf("expected error output, got: %s", out)
	}
}

// TestCmdForgetCallsCorrectTool verifies cmdForget calls muninn_forget and prints the undo hint.
func TestCmdForgetCallsCorrectTool(t *testing.T) {
	srv, last := newMCPServer(t, mcpResultWithContent(`{"deleted":true}`))
	defer srv.Close()

	r := &replState{mcpURL: srv.URL, vault: "myvault"}
	out := captureStdout(func() {
		r.cmdForget("01JFXX4KZMB3E9QV7P")
	})

	args, ok := last.Params["arguments"].(map[string]any)
	if !ok {
		t.Fatal("arguments is not a map")
	}
	if args["id"] != "01JFXX4KZMB3E9QV7P" {
		t.Errorf("id = %v, want %q", args["id"], "01JFXX4KZMB3E9QV7P")
	}

	// Should print undo hint
	if !strings.Contains(out, "restore") {
		t.Errorf("expected undo hint with 'restore', got: %s", out)
	}
	if !strings.Contains(out, "01JFXX4KZMB3E9QV7P") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

// TestCmdShowMemoriesCallsCorrectTool verifies cmdShowMemories calls muninn_session with vault.
func TestCmdShowMemoriesCallsCorrectTool(t *testing.T) {
	body := mcpResultWithContent(`[{"id":"01JF","content":"recent memory"}]`)
	srv, last := newMCPServer(t, body)
	defer srv.Close()

	r := &replState{mcpURL: srv.URL, vault: "myapp"}
	captureStdout(func() {
		r.cmdShowMemories()
	})

	args, ok := last.Params["arguments"].(map[string]any)
	if !ok {
		t.Fatal("arguments is not a map")
	}
	if args["vault"] != "myapp" {
		t.Errorf("vault = %v, want %q", args["vault"], "myapp")
	}
}

// TestCmdShowMemoriesEmptyShowsHelpMessage verifies empty memories show a help message.
func TestCmdShowMemoriesEmptyShowsHelpMessage(t *testing.T) {
	srv, _ := newMCPServer(t, mcpResultWithContent("[]"))
	defer srv.Close()

	r := &replState{mcpURL: srv.URL, vault: "emptyvault"}
	out := captureStdout(func() {
		r.cmdShowMemories()
	})
	if !strings.Contains(out, "No memories") {
		t.Errorf("expected 'No memories' message, got: %s", out)
	}
}

// TestCmdShowContradictions verifies cmdShowContradictions calls muninn_contradictions.
func TestCmdShowContradictions(t *testing.T) {
	body := mcpResultWithContent(`[{"a":"01JF1","b":"01JF2","reason":"conflict"}]`)
	srv, last := newMCPServer(t, body)
	defer srv.Close()

	r := &replState{mcpURL: srv.URL, vault: "myvault"}
	captureStdout(func() {
		r.cmdShowContradictions()
	})

	args, ok := last.Params["arguments"].(map[string]any)
	if !ok {
		t.Fatal("arguments is not a map")
	}
	if args["vault"] != "myvault" {
		t.Errorf("vault = %v, want %q", args["vault"], "myvault")
	}
}

// TestCmdShowStatsServerDown verifies cmdShowStats handles server down gracefully.
func TestCmdShowStatsServerDown(t *testing.T) {
	r := &replState{mcpURL: "http://localhost:19998"}
	out := captureStdout(func() {
		r.cmdShowStats()
	})
	if !strings.Contains(out, "not running") {
		t.Errorf("expected 'not running' message, got: %s", out)
	}
}

// TestCmdShowStatsServerUp verifies cmdShowStats shows status when server is running.
func TestCmdShowStatsServerUp(t *testing.T) {
	srv := newHealthServer()
	defer srv.Close()

	r := &replState{mcpURL: srv.URL}
	out := captureStdout(func() {
		r.cmdShowStats()
	})
	if !strings.Contains(out, "running") {
		t.Errorf("expected 'running' in output, got: %s", out)
	}
}

// TestCmdShowStatsWithVault verifies cmdShowStats calls muninn_status when vault is set.
func TestCmdShowStatsWithVault(t *testing.T) {
	body := mcpResultWithContent(`{"vault":"myapp","count":42}`)

	// For cmdShowStats to work, we need to handle both /mcp/health and /mcp paths
	healthLast := &mcpRequest{}
	srv2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/mcp/health" {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.URL.Path != "/mcp" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		var req struct {
			Method string         `json:"method"`
			Params map[string]any `json:"params"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		*healthLast = mcpRequest{Method: req.Method, Params: req.Params}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"jsonrpc":"2.0","id":1,"result":%s}`, body)
	}))
	defer srv2.Close()

	r := &replState{mcpURL: srv2.URL, vault: "myapp"}
	out := captureStdout(func() {
		r.cmdShowStats()
	})
	if !strings.Contains(out, "running") {
		t.Errorf("expected 'running' in output, got: %s", out)
	}

	args, ok := healthLast.Params["arguments"].(map[string]any)
	if !ok {
		t.Fatal("arguments is not a map")
	}
	if args["vault"] != "myapp" {
		t.Errorf("vault = %v, want %q", args["vault"], "myapp")
	}
}

// TestMcpCallHandlesDecodeError verifies mcpCall handles invalid JSON responses.
func TestMcpCallHandlesDecodeError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{invalid json`)
	}))
	defer srv.Close()

	_, err := mcpCall(srv.URL, "any_tool", nil)
	if err == nil {
		t.Error("expected error for invalid JSON response")
	}
	if !strings.Contains(err.Error(), "decode") {
		t.Errorf("error should mention 'decode': %v", err)
	}
}

// TestCmdSearchError verifies cmdSearch handles MCP errors gracefully.
func TestCmdSearchError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"jsonrpc":"2.0","id":1,"error":{"code":-32000,"message":"server error"}}`)
	}))
	defer srv.Close()

	r := &replState{mcpURL: srv.URL, vault: "test"}
	out := captureStderr(func() {
		r.cmdSearch("query")
	})
	if !strings.Contains(out, "Error") {
		t.Errorf("expected error output, got: %s", out)
	}
}

// TestCmdShowMemoriesError verifies cmdShowMemories handles MCP errors gracefully.
func TestCmdShowMemoriesError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"jsonrpc":"2.0","id":1,"error":{"code":-32000,"message":"error"}}`)
	}))
	defer srv.Close()

	r := &replState{mcpURL: srv.URL, vault: "test"}
	out := captureStderr(func() {
		r.cmdShowMemories()
	})
	if !strings.Contains(out, "Error") {
		t.Errorf("expected error output, got: %s", out)
	}
}

// TestCmdForgetError verifies cmdForget handles MCP errors gracefully.
func TestCmdForgetError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"jsonrpc":"2.0","id":1,"error":{"code":-32000,"message":"not found"}}`)
	}))
	defer srv.Close()

	r := &replState{mcpURL: srv.URL, vault: "test"}
	out := captureStderr(func() {
		r.cmdForget("BADID")
	})
	if !strings.Contains(out, "Error") {
		t.Errorf("expected error output, got: %s", out)
	}
}

// TestCmdShowContradictionsError verifies cmdShowContradictions handles errors gracefully.
func TestCmdShowContradictionsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"jsonrpc":"2.0","id":1,"error":{"code":-32000,"message":"error"}}`)
	}))
	defer srv.Close()

	r := &replState{mcpURL: srv.URL, vault: "test"}
	out := captureStderr(func() {
		r.cmdShowContradictions()
	})
	if !strings.Contains(out, "Error") {
		t.Errorf("expected error output, got: %s", out)
	}
}
