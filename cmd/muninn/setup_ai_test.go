package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestGenerateToken verifies token format and uniqueness.
func TestGenerateToken(t *testing.T) {
	tok1, err := generateToken()
	if err != nil {
		t.Fatalf("generateToken: %v", err)
	}
	if !strings.HasPrefix(tok1, "mdb_") {
		t.Errorf("token should start with mdb_, got %q", tok1)
	}
	// prefix (4) + 48 hex chars = 52 total
	if len(tok1) != 52 {
		t.Errorf("expected token length 52, got %d (%s)", len(tok1), tok1)
	}
	tok2, _ := generateToken()
	if tok1 == tok2 {
		t.Error("two generated tokens should not be equal")
	}
}

// TestLoadOrGenerateToken_NewToken verifies a fresh token is created when none exists.
func TestLoadOrGenerateToken_NewToken(t *testing.T) {
	dir := t.TempDir()
	tok, isNew, err := loadOrGenerateToken(filepath.Join(dir, "data"))
	if err != nil {
		t.Fatalf("loadOrGenerateToken: %v", err)
	}
	if !isNew {
		t.Error("expected isNew=true for fresh token")
	}
	if !strings.HasPrefix(tok, "mdb_") {
		t.Errorf("token should start with mdb_, got %q", tok)
	}
	// Verify file was written
	tokenFile := filepath.Join(dir, "mcp.token")
	b, err := os.ReadFile(tokenFile)
	if err != nil {
		t.Fatalf("token file not written: %v", err)
	}
	if strings.TrimSpace(string(b)) != tok {
		t.Errorf("token file content mismatch")
	}
	// Verify file permissions
	info, _ := os.Stat(tokenFile)
	if info.Mode().Perm() != 0600 {
		t.Errorf("token file should be 0600, got %o", info.Mode().Perm())
	}
}

// TestLoadOrGenerateToken_ExistingToken verifies an existing token is reused.
func TestLoadOrGenerateToken_ExistingToken(t *testing.T) {
	dir := t.TempDir()
	tokenFile := filepath.Join(dir, "mcp.token")
	os.WriteFile(tokenFile, []byte("mdb_existingtoken1234567890abcdef1234567890abcde\n"), 0600)

	tok, isNew, err := loadOrGenerateToken(filepath.Join(dir, "data"))
	if err != nil {
		t.Fatalf("loadOrGenerateToken: %v", err)
	}
	if isNew {
		t.Error("expected isNew=false when token file already exists")
	}
	if tok != "mdb_existingtoken1234567890abcdef1234567890abcde" {
		t.Errorf("unexpected token: %q", tok)
	}
}

// TestWriteAIToolConfig_NewFile verifies config creation when no file exists.
func TestWriteAIToolConfig_NewFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "claude_desktop_config.json")

	summary, err := writeAIToolConfig(path, func(cfg map[string]any) {
		mergeMCPServers(cfg, "http://localhost:8750/mcp", "mdb_testtoken")
	})
	if err != nil {
		t.Fatalf("writeAIToolConfig: %v", err)
	}
	if !strings.Contains(summary, "mcpServers.muninn") {
		t.Errorf("unexpected summary: %q", summary)
	}

	// Read back and verify
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("config file not written: %v", err)
	}
	var cfg map[string]any
	if err := json.Unmarshal(b, &cfg); err != nil {
		t.Fatalf("config file is not valid JSON: %v", err)
	}
	servers, ok := cfg["mcpServers"].(map[string]any)
	if !ok {
		t.Fatal("mcpServers not found in config")
	}
	entry, ok := servers["muninn"].(map[string]any)
	if !ok {
		t.Fatal("muninn entry not found in mcpServers")
	}
	if entry["url"] != "http://localhost:8750/mcp" {
		t.Errorf("unexpected URL in config: %v", entry["url"])
	}
}

// TestWriteAIToolConfig_PreservesExistingServers verifies other mcpServers are not clobbered.
func TestWriteAIToolConfig_PreservesExistingServers(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	// Write existing config with another server
	existing := map[string]any{
		"mcpServers": map[string]any{
			"other-tool": map[string]any{"url": "http://other:9999"},
		},
		"someOtherKey": "someValue",
	}
	b, _ := json.Marshal(existing)
	os.WriteFile(path, b, 0600)

	_, err := writeAIToolConfig(path, func(cfg map[string]any) {
		mergeMCPServers(cfg, "http://localhost:8750/mcp", "")
	})
	if err != nil {
		t.Fatalf("writeAIToolConfig: %v", err)
	}

	// Read back
	b2, _ := os.ReadFile(path)
	var cfg map[string]any
	json.Unmarshal(b2, &cfg)

	servers := cfg["mcpServers"].(map[string]any)
	if _, ok := servers["other-tool"]; !ok {
		t.Error("other-tool server was removed")
	}
	if _, ok := servers["muninn"]; !ok {
		t.Error("muninn server not added")
	}
	if cfg["someOtherKey"] != "someValue" {
		t.Error("top-level key was removed")
	}
}

// TestWriteAIToolConfig_InvalidExistingJSON verifies graceful error on corrupt config.
func TestWriteAIToolConfig_InvalidExistingJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	os.WriteFile(path, []byte("this is not json {{{{"), 0644)

	_, err := writeAIToolConfig(path, func(cfg map[string]any) {
		mergeMCPServers(cfg, "http://localhost:8750/mcp", "")
	})
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
	if !strings.Contains(err.Error(), "invalid JSON") {
		t.Errorf("expected 'invalid JSON' in error, got: %v", err)
	}
}

// TestWriteAIToolConfig_CreatesParentDir verifies missing parent directories are created.
func TestWriteAIToolConfig_CreatesParentDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "Claude", "claude_desktop_config.json")
	// Parent dir does NOT exist yet

	_, err := writeAIToolConfig(path, func(cfg map[string]any) {
		mergeMCPServers(cfg, "http://localhost:8750/mcp", "")
	})
	if err != nil {
		t.Fatalf("writeAIToolConfig should create parent dirs: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Error("config file not created")
	}
}

// TestWriteAIToolConfig_BackupCreated verifies .bak is created for existing files.
func TestWriteAIToolConfig_BackupCreated(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	original := []byte(`{"mcpServers":{}}`)
	os.WriteFile(path, original, 0644)

	writeAIToolConfig(path, func(cfg map[string]any) {
		mergeMCPServers(cfg, "http://localhost:8750/mcp", "")
	})

	bak, err := os.ReadFile(path + ".bak")
	if err != nil {
		t.Fatal("backup file not created")
	}
	if string(bak) != string(original) {
		t.Error("backup content does not match original")
	}
}

// TestWriteAIToolConfig_AtomicTempCleaned verifies temp file is cleaned up after success.
func TestWriteAIToolConfig_AtomicTempCleaned(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	writeAIToolConfig(path, func(cfg map[string]any) {
		mergeMCPServers(cfg, "http://localhost:8750/mcp", "")
	})

	// No temp files should remain
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if strings.Contains(e.Name(), ".tmp.") {
			t.Errorf("temp file not cleaned up: %s", e.Name())
		}
	}
}

// TestMCPServerEntry_WithToken verifies token is included when provided.
func TestMCPServerEntry_WithToken(t *testing.T) {
	entry := mcpServerEntry("http://localhost:8750/mcp", "mdb_abc123")
	if entry["url"] != "http://localhost:8750/mcp" {
		t.Errorf("unexpected url: %v", entry["url"])
	}
	headers, ok := entry["headers"].(map[string]any)
	if !ok {
		t.Fatal("headers not found")
	}
	if headers["Authorization"] != "Bearer mdb_abc123" {
		t.Errorf("unexpected Authorization: %v", headers["Authorization"])
	}
}

// TestMCPServerEntry_NoToken verifies no headers when token is empty.
func TestMCPServerEntry_NoToken(t *testing.T) {
	entry := mcpServerEntry("http://localhost:8750/mcp", "")
	if _, ok := entry["headers"]; ok {
		t.Error("headers should not be present when token is empty")
	}
}

// TestParseToolNumbers verifies tool selection parsing.
func TestParseToolNumbers(t *testing.T) {
	tests := []struct {
		input    string
		expected []int
	}{
		{"1", []int{1}},
		{"1 2 3", []int{1, 2, 3}},
		{"1,2,3", []int{1, 2, 3}},
		{"1 1 2", []int{1, 2}}, // deduplication
		{"", nil},
		{"6 7 8", []int{6, 7, 8}}, // valid range is 1-9
		{"abc", nil},    // non-numeric
	}
	for _, tt := range tests {
		got := parseToolNumbers(tt.input)
		if len(got) != len(tt.expected) {
			t.Errorf("parseToolNumbers(%q): got %v, want %v", tt.input, got, tt.expected)
			continue
		}
		for i := range got {
			if got[i] != tt.expected[i] {
				t.Errorf("parseToolNumbers(%q)[%d]: got %d, want %d", tt.input, i, got[i], tt.expected[i])
			}
		}
	}
}

// TestOpenClawConfigPath verifies OpenClaw config path is set correctly.
func TestOpenClawConfigPath(t *testing.T) {
	path := openClawConfigPath()
	if path == "" {
		t.Error("openClawConfigPath returned empty string")
	}
	home, _ := os.UserHomeDir()
	if !strings.HasPrefix(path, home) {
		t.Errorf("path %q should start with home dir", path)
	}
}

// Helper to override HOME in tests
func withTempHome(t *testing.T) (string, func()) {
	t.Helper()
	tmp := t.TempDir()
	orig := os.Getenv("HOME")
	os.Setenv("HOME", tmp)
	// Also set XDG_CONFIG_HOME to temp dir for Linux tests
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmp)
	// Also set APPDATA for Windows tests
	origAPPDATA := os.Getenv("APPDATA")
	os.Setenv("APPDATA", tmp)
	return tmp, func() {
		os.Setenv("HOME", orig)
		os.Setenv("XDG_CONFIG_HOME", origXDG)
		os.Setenv("APPDATA", origAPPDATA)
	}
}

// TestConfigureClaudeDesktopWritesConfig verifies Claude Desktop config is written at correct path with correct JSON.
func TestConfigureClaudeDesktopWritesConfig(t *testing.T) {
	home, cleanup := withTempHome(t)
	defer cleanup()

	mcpURL := "http://localhost:8750/mcp"
	token := "mdb_testtoken123"

	out := captureStdout(func() {
		err := configureClaudeDesktop(mcpURL, token)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	// Path should be inside temp home
	configPath := claudeDesktopConfigPath()
	if !strings.HasPrefix(configPath, home) {
		t.Errorf("config path %q should be inside temp home %q", configPath, home)
	}

	// Read and verify the written JSON
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("config file not written: %v", err)
	}
	var cfg map[string]any
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("invalid JSON written: %v\ncontents: %s", err, data)
	}

	servers, ok := cfg["mcpServers"].(map[string]any)
	if !ok {
		t.Fatalf("mcpServers not found in config: %s", data)
	}
	muninn, ok := servers["muninn"].(map[string]any)
	if !ok {
		t.Fatalf("mcpServers.muninn not found: %s", data)
	}
	if muninn["url"] != mcpURL {
		t.Errorf("url = %v, want %q", muninn["url"], mcpURL)
	}
	headers, ok := muninn["headers"].(map[string]any)
	if !ok {
		t.Fatalf("headers not found when token supplied: %s", data)
	}
	if headers["Authorization"] != "Bearer "+token {
		t.Errorf("Authorization = %v, want %q", headers["Authorization"], "Bearer "+token)
	}

	// Output should contain success marker
	if !strings.Contains(out, "✓") {
		t.Errorf("output missing success marker '✓': %s", out)
	}
}

// TestConfigureClaudeDesktopNoToken verifies no auth header is written when token is empty.
func TestConfigureClaudeDesktopNoToken(t *testing.T) {
	_, cleanup := withTempHome(t)
	defer cleanup()

	captureStdout(func() {
		if err := configureClaudeDesktop("http://localhost:8750/mcp", ""); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	data, _ := os.ReadFile(claudeDesktopConfigPath())
	var cfg map[string]any
	json.Unmarshal(data, &cfg)
	servers := cfg["mcpServers"].(map[string]any)
	muninn := servers["muninn"].(map[string]any)

	if _, hasHeaders := muninn["headers"]; hasHeaders {
		t.Error("headers should not be present when token is empty")
	}
}

// TestConfigureClaudeDesktopPreservesExistingKeys verifies existing config keys are not lost.
func TestConfigureClaudeDesktopPreservesExistingKeys(t *testing.T) {
	_, cleanup := withTempHome(t)
	defer cleanup()

	// Pre-populate with an existing MCP server
	path := claudeDesktopConfigPath()
	os.MkdirAll(filepath.Dir(path), 0755)
	existing := `{"mcpServers":{"other-tool":{"url":"http://other.example"}},"someOtherKey":"preserved"}`
	os.WriteFile(path, []byte(existing), 0644)

	captureStdout(func() {
		configureClaudeDesktop("http://localhost:8750/mcp", "tok123")
	})

	data, _ := os.ReadFile(path)
	var cfg map[string]any
	json.Unmarshal(data, &cfg)

	// Original key preserved
	if cfg["someOtherKey"] != "preserved" {
		t.Errorf("someOtherKey was lost: %s", data)
	}
	servers := cfg["mcpServers"].(map[string]any)
	// Original server preserved
	if _, ok := servers["other-tool"]; !ok {
		t.Errorf("other-tool was lost: %s", data)
	}
	// muninn added
	if _, ok := servers["muninn"]; !ok {
		t.Errorf("muninn not added: %s", data)
	}
}

// TestConfigureCursorWritesConfig verifies Cursor config is written correctly.
func TestConfigureCursorWritesConfig(t *testing.T) {
	_, cleanup := withTempHome(t)
	defer cleanup()

	out := captureStdout(func() {
		if err := configureCursor("http://localhost:8750/mcp", "tok"); err != nil {
			t.Fatalf("error: %v", err)
		}
	})

	data, err := os.ReadFile(cursorConfigPath())
	if err != nil {
		t.Fatalf("file not written: %v", err)
	}
	if !strings.Contains(string(data), `"muninn"`) {
		t.Errorf("muninn not in config: %s", data)
	}
	if !strings.Contains(string(data), "8750") {
		t.Errorf("MCP port not in config: %s", data)
	}
	if !strings.Contains(out, "✓") {
		t.Errorf("output missing success marker: %s", out)
	}
}

// TestConfigureWindsurfWritesConfig verifies Windsurf config is written correctly.
func TestConfigureWindsurfWritesConfig(t *testing.T) {
	_, cleanup := withTempHome(t)
	defer cleanup()

	out := captureStdout(func() {
		if err := configureWindsurf("http://localhost:8750/mcp", "tok"); err != nil {
			t.Fatalf("error: %v", err)
		}
	})

	data, err := os.ReadFile(windsurfConfigPath())
	if err != nil {
		t.Fatalf("file not written: %v", err)
	}
	if !strings.Contains(string(data), `"muninn"`) {
		t.Errorf("muninn not in config: %s", data)
	}
	if !strings.Contains(string(data), "8750") {
		t.Errorf("MCP port not in config: %s", data)
	}
	if !strings.Contains(out, "✓") {
		t.Errorf("output missing success marker: %s", out)
	}
}

// TestConfigureOpenClawWritesConfig verifies OpenClaw config is written correctly.
func TestConfigureOpenClawWritesConfig(t *testing.T) {
	_, cleanup := withTempHome(t)
	defer cleanup()

	out := captureStdout(func() {
		if err := configureOpenClaw("http://localhost:8750/mcp", "tok"); err != nil {
			t.Fatalf("error: %v", err)
		}
	})

	data, err := os.ReadFile(openClawConfigPath())
	if err != nil {
		t.Fatalf("file not written: %v", err)
	}
	if !strings.Contains(string(data), `"muninn"`) {
		t.Errorf("muninn not in config: %s", data)
	}
	if !strings.Contains(string(data), "8750") {
		t.Errorf("MCP port not in config: %s", data)
	}
	if !strings.Contains(out, "✓") {
		t.Errorf("output missing success marker: %s", out)
	}
}

// TestPrintVSCodeInstructions verifies VS Code instructions contain required elements.
func TestPrintVSCodeInstructions(t *testing.T) {
	out := captureStdout(func() {
		printVSCodeInstructions("http://localhost:8750/mcp", "mdb_mytoken")
	})
	if !strings.Contains(out, `"muninn"`) {
		t.Errorf("missing muninn key: %s", out)
	}
	if !strings.Contains(out, "8750") {
		t.Errorf("missing MCP URL: %s", out)
	}
	if !strings.Contains(out, "mdb_mytoken") {
		t.Errorf("missing token: %s", out)
	}
	if !strings.Contains(out, "Bearer") {
		t.Errorf("missing Bearer auth: %s", out)
	}
	// VS Code uses "servers" not "mcpServers"
	if !strings.Contains(out, `"servers"`) {
		t.Errorf("VS Code format should use 'servers' key: %s", out)
	}
}

// TestPrintVSCodeInstructionsNoToken verifies no auth header without token.
func TestPrintVSCodeInstructionsNoToken(t *testing.T) {
	out := captureStdout(func() {
		printVSCodeInstructions("http://localhost:8750/mcp", "")
	})
	if strings.Contains(out, "Bearer") {
		t.Errorf("should not have auth header without token: %s", out)
	}
}

// TestPrintManualInstructions verifies manual instructions contain required elements.
func TestPrintManualInstructions(t *testing.T) {
	out := captureStdout(func() {
		printManualInstructions("http://localhost:8750/mcp", "mdb_secrettoken")
	})
	if !strings.Contains(out, "mcpServers") {
		t.Errorf("missing mcpServers: %s", out)
	}
	if !strings.Contains(out, "mdb_secrettoken") {
		t.Errorf("missing token: %s", out)
	}
	if !strings.Contains(out, "curl") {
		t.Errorf("missing curl test command: %s", out)
	}
	if !strings.Contains(out, "Bearer mdb_secrettoken") {
		t.Errorf("missing auth in curl: %s", out)
	}
}

// TestPrintManualInstructionsNoToken verifies curl command appears without token.
func TestPrintManualInstructionsNoToken(t *testing.T) {
	out := captureStdout(func() {
		printManualInstructions("http://localhost:8750/mcp", "")
	})
	if strings.Contains(out, "Bearer") {
		t.Errorf("should not have auth header without token: %s", out)
	}
	// curl command should still appear
	if !strings.Contains(out, "curl") {
		t.Errorf("missing curl command: %s", out)
	}
}

// TestConfigureNamedToolsClaudeDesktop verifies claude alias configures Claude Desktop.
func TestConfigureNamedToolsClaudeDesktop(t *testing.T) {
	_, cleanup := withTempHome(t)
	defer cleanup()

	out := captureStdout(func() {
		configureNamedTools([]string{"claude"}, "http://localhost:8750/mcp", "tok123")
	})
	if !strings.Contains(out, "✓") {
		t.Errorf("expected success marker for claude tool, got: %s", out)
	}

	// Verify file was written
	path := claudeDesktopConfigPath()
	if _, err := os.ReadFile(path); err != nil {
		t.Errorf("claude Desktop config file not written: %v", err)
	}
}

// TestConfigureNamedToolsClaudeDesktopAlias verifies claude-desktop alias works.
func TestConfigureNamedToolsClaudeDesktopAlias(t *testing.T) {
	_, cleanup := withTempHome(t)
	defer cleanup()

	out := captureStdout(func() {
		configureNamedTools([]string{"claude-desktop"}, "http://localhost:8750/mcp", "tok")
	})
	if !strings.Contains(out, "✓") {
		t.Errorf("claude-desktop alias should work: %s", out)
	}
}

// TestConfigureNamedToolsCursor verifies cursor tool configures Cursor.
func TestConfigureNamedToolsCursor(t *testing.T) {
	_, cleanup := withTempHome(t)
	defer cleanup()

	out := captureStdout(func() {
		configureNamedTools([]string{"cursor"}, "http://localhost:8750/mcp", "tok123")
	})
	if !strings.Contains(out, "✓") {
		t.Errorf("expected success marker for cursor tool, got: %s", out)
	}

	path := cursorConfigPath()
	if _, err := os.ReadFile(path); err != nil {
		t.Errorf("cursor config file not written: %v", err)
	}
}

// TestConfigureNamedToolsWindsurf verifies windsurf tool configures Windsurf.
func TestConfigureNamedToolsWindsurf(t *testing.T) {
	_, cleanup := withTempHome(t)
	defer cleanup()

	out := captureStdout(func() {
		configureNamedTools([]string{"windsurf"}, "http://localhost:8750/mcp", "tok123")
	})
	if !strings.Contains(out, "✓") {
		t.Errorf("expected success marker for windsurf tool, got: %s", out)
	}

	path := windsurfConfigPath()
	if _, err := os.ReadFile(path); err != nil {
		t.Errorf("windsurf config file not written: %v", err)
	}
}

// TestConfigureNamedToolsOpenClaw verifies openclaw tool configures OpenClaw.
func TestConfigureNamedToolsOpenClaw(t *testing.T) {
	_, cleanup := withTempHome(t)
	defer cleanup()

	out := captureStdout(func() {
		configureNamedTools([]string{"openclaw"}, "http://localhost:8750/mcp", "tok123")
	})
	if !strings.Contains(out, "✓") {
		t.Errorf("expected success marker for openclaw tool, got: %s", out)
	}

	path := openClawConfigPath()
	if _, err := os.ReadFile(path); err != nil {
		t.Errorf("openclaw config file not written: %v", err)
	}
}

// TestConfigureNamedToolsVSCode verifies vscode tool shows instructions.
func TestConfigureNamedToolsVSCode(t *testing.T) {
	_, cleanup := withTempHome(t)
	defer cleanup()

	out := captureStdout(func() {
		configureNamedTools([]string{"vscode"}, "http://localhost:8750/mcp", "")
	})
	if !strings.Contains(out, "VS Code") {
		t.Errorf("expected VS Code instructions, got: %s", out)
	}
	if !strings.Contains(out, `"servers"`) {
		t.Errorf("expected VS Code format with 'servers' key: %s", out)
	}
}

// TestConfigureNamedToolsVSCodeAlias verifies vs-code alias works.
func TestConfigureNamedToolsVSCodeAlias(t *testing.T) {
	_, cleanup := withTempHome(t)
	defer cleanup()

	out := captureStdout(func() {
		configureNamedTools([]string{"vs-code"}, "http://localhost:8750/mcp", "")
	})
	if !strings.Contains(out, "VS Code") {
		t.Errorf("expected VS Code instructions with vs-code alias: %s", out)
	}
}

// TestConfigureNamedToolsManual verifies manual tool shows manual instructions.
func TestConfigureNamedToolsManual(t *testing.T) {
	_, cleanup := withTempHome(t)
	defer cleanup()

	out := captureStdout(func() {
		configureNamedTools([]string{"manual"}, "http://localhost:8750/mcp", "")
	})
	if !strings.Contains(out, "mcpServers") {
		t.Errorf("expected manual instructions, got: %s", out)
	}
	if !strings.Contains(out, "curl") {
		t.Errorf("expected curl test command: %s", out)
	}
}

// TestConfigureNamedToolsOtherAlias verifies other alias works for manual.
func TestConfigureNamedToolsOtherAlias(t *testing.T) {
	_, cleanup := withTempHome(t)
	defer cleanup()

	out := captureStdout(func() {
		configureNamedTools([]string{"other"}, "http://localhost:8750/mcp", "")
	})
	if !strings.Contains(out, "mcpServers") {
		t.Errorf("expected manual instructions with 'other' alias: %s", out)
	}
}

// TestConfigureNamedToolsMultiple verifies multiple tools can be configured in one call.
func TestConfigureNamedToolsMultiple(t *testing.T) {
	_, cleanup := withTempHome(t)
	defer cleanup()

	out := captureStdout(func() {
		configureNamedTools([]string{"claude", "cursor"}, "http://localhost:8750/mcp", "tok123")
	})

	// Both should succeed
	if strings.Count(out, "✓") < 2 {
		t.Errorf("expected 2 success markers for 2 tools, got: %s", out)
	}

	// Both files should exist
	claudePath := claudeDesktopConfigPath()
	cursorPath := cursorConfigPath()
	if _, err := os.ReadFile(claudePath); err != nil {
		t.Errorf("claude config not written: %v", err)
	}
	if _, err := os.ReadFile(cursorPath); err != nil {
		t.Errorf("cursor config not written: %v", err)
	}
}

// TestConfigureNamedToolsUnknownToolSetupAI verifies unknown tools are gracefully ignored with error message.
func TestConfigureNamedToolsUnknownToolSetupAI(t *testing.T) {
	_, cleanup := withTempHome(t)
	defer cleanup()

	stderr := captureStderr(func() {
		configureNamedTools([]string{"nonexistent"}, "http://localhost:8750/mcp", "")
	})
	if !strings.Contains(stderr, "unknown tool") {
		t.Errorf("expected error for unknown tool, got stderr: %s", stderr)
	}
}
