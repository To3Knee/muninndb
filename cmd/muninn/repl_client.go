package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func mcpHealthCheck(baseURL string) bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(baseURL + "/mcp/health")
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == 200
}

func mcpCall(baseURL, toolName string, args map[string]any) (map[string]any, error) {
	body, _ := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"method":  "tools/call",
		"id":      1,
		"params": map[string]any{
			"name":      toolName,
			"arguments": args,
		},
	})
	resp, err := http.Post(baseURL+"/mcp", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if errObj, ok := result["error"]; ok {
		return nil, fmt.Errorf("MCP error: %v", errObj)
	}
	res, _ := result["result"].(map[string]any)
	return res, nil
}

func (r *replState) cmdShowVaults() {
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", "http://localhost:8475/api/vaults", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}
	if r.sessionCookie != "" {
		req.AddCookie(&http.Cookie{Name: "session", Value: r.sessionCookie})
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to server: %v\n", err)
		fmt.Println("Is muninn running? Try: muninn start")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		// Fallback: show static message
		fmt.Println("  default   (built-in)")
		fmt.Println()
		fmt.Println("  For full vault list, open: http://localhost:8476")
		return
	}

	var result any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding response: %v\n", err)
		return
	}

	// The API may return []vault or {vaults: []} — handle both
	var vaults []map[string]any
	switch v := result.(type) {
	case []any:
		for _, item := range v {
			if m, ok := item.(map[string]any); ok {
				vaults = append(vaults, m)
			}
		}
	case map[string]any:
		if list, ok := v["vaults"].([]any); ok {
			for _, item := range list {
				if m, ok := item.(map[string]any); ok {
					vaults = append(vaults, m)
				}
			}
		}
	}

	if len(vaults) == 0 {
		fmt.Println("  No vaults found.")
		fmt.Println("  Vaults are created automatically when your AI tools store their first memory.")
		return
	}

	formatVaultTable(vaults)
}

func (r *replState) cmdShowMemories() {
	since := time.Now().Add(-24 * time.Hour).Format(time.RFC3339)
	result, err := mcpCall(r.mcpURL, "muninn_session", map[string]any{
		"vault": r.vault,
		"since": since,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}
	// Check for empty result and provide a helpful message
	if isEmptyMCPResult(result) {
		fmt.Println("No memories in vault '" + r.vault + "' in the last 24 hours.")
		fmt.Println()
		fmt.Println("Memories are created automatically when your AI tools are connected.")
		fmt.Println("Run 'muninn init' to connect Claude Desktop, Cursor, or other tools.")
		return
	}
	prettyPrint(result)
}

func (r *replState) cmdSearch(query string) {
	result, err := mcpCall(r.mcpURL, "muninn_recall", map[string]any{
		"vault":   r.vault,
		"context": []string{query},
		"limit":   10,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}
	if isEmptyMCPResult(result) {
		fmt.Printf("No memories match '%s'.\n", query)
		fmt.Println()
		fmt.Println("Tips:")
		fmt.Println("  • Try broader terms or synonyms")
		fmt.Println("  • Semantic search works best with natural language phrases")
		fmt.Println("  • Check Settings → Vault in the web UI to verify an embedder is configured")
		return
	}
	prettyPrint(result)
}

func (r *replState) cmdGet(id string) {
	result, err := mcpCall(r.mcpURL, "muninn_read", map[string]any{
		"vault": r.vault,
		"id":    id,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}
	prettyPrint(result)
}

func (r *replState) cmdForget(id string) {
	_, err := mcpCall(r.mcpURL, "muninn_forget", map[string]any{
		"vault": r.vault,
		"id":    id,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}
	fmt.Printf("  Soft-deleted memory %s\n", id)
	fmt.Printf("  Undo with: restore %s\n", id)
}

func (r *replState) cmdShowContradictions() {
	result, err := mcpCall(r.mcpURL, "muninn_contradictions", map[string]any{
		"vault": r.vault,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}
	prettyPrint(result)
}

func (r *replState) cmdShowStats() {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(r.mcpURL + "/mcp/health")
	if err != nil || resp.StatusCode != 200 {
		fmt.Println("Server: not running")
		fmt.Println("Start with: muninn start")
		return
	}
	resp.Body.Close()
	fmt.Println("Server: running")
	fmt.Println("  MBP  :8474   binary protocol")
	fmt.Println("  REST :8475   JSON API")
	fmt.Println("  MCP  :8750   AI tool integration")
	fmt.Println("  UI   :8476   http://localhost:8476")
	if r.vault != "" {
		fmt.Println()
		result, err := mcpCall(r.mcpURL, "muninn_status", map[string]any{"vault": r.vault})
		if err == nil {
			prettyPrint(result)
		}
	}
}

// isEmptyMCPResult returns true if the MCP result has no meaningful content.
func isEmptyMCPResult(result map[string]any) bool {
	if result == nil {
		return true
	}
	content, ok := result["content"].([]any)
	if !ok || len(content) == 0 {
		return true
	}
	item, ok := content[0].(map[string]any)
	if !ok {
		return true
	}
	text, _ := item["text"].(string)
	// Check for empty JSON arrays/objects or empty strings
	trimmed := strings.TrimSpace(text)
	return trimmed == "" || trimmed == "[]" || trimmed == "{}" || trimmed == "null"
}

func runShowVaults() {
	r := &replState{mcpURL: "http://localhost:8750"}
	r.cmdShowVaults()
}

// prettyPrint extracts text content from an MCP result and prints it.
func prettyPrint(result map[string]any) {
	if result == nil {
		fmt.Println("(no result)")
		return
	}
	// MCP text content format: result["content"][0]["text"]
	if content, ok := result["content"].([]any); ok && len(content) > 0 {
		if item, ok := content[0].(map[string]any); ok {
			if text, ok := item["text"].(string); ok {
				var v any
				if json.Unmarshal([]byte(text), &v) == nil {
					b, _ := json.MarshalIndent(v, "", "  ")
					fmt.Println(string(b))
					return
				}
				fmt.Println(text)
				return
			}
		}
	}
	b, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(b))
}

// Config persistence

func configPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".muninn", "config")
}

func loadDefaultVault() string {
	b, err := os.ReadFile(configPath())
	if err != nil {
		return ""
	}
	var cfg map[string]string
	if json.Unmarshal(b, &cfg) == nil {
		return cfg["default_vault"]
	}
	return ""
}

func saveDefaultVault(vault string) {
	path := configPath()
	os.MkdirAll(filepath.Dir(path), 0700)
	cfg := map[string]string{"default_vault": vault}
	b, _ := json.Marshal(cfg)
	os.WriteFile(path, b, 0644)
}

// formatVaultTable prints vaults in a table format.
func formatVaultTable(vaults []map[string]any) {
	fmt.Printf("\n  %-20s  %-10s  %s\n", "NAME", "MEMORIES", "LAST ACTIVE")
	fmt.Printf("  %-20s  %-10s  %s\n", "────────────────────", "──────────", "───────────")
	for _, v := range vaults {
		name, _ := v["name"].(string)
		count := 0
		if c, ok := v["memory_count"].(float64); ok {
			count = int(c)
		}
		lastActive := "—"
		if la, ok := v["last_active"].(string); ok && la != "" {
			lastActive = humanizeTime(la)
		}
		fmt.Printf("  %-20s  %-10d  %s\n", name, count, lastActive)
	}
	fmt.Println()
}

// humanizeTime converts an RFC3339 timestamp to a human-friendly relative string.
func humanizeTime(ts string) string {
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return ts
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%d minutes ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%d hours ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%d days ago", int(d.Hours()/24))
	}
}
