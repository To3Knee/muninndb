package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

func runVault(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: muninn vault <delete|clear|clone|merge>")
		fmt.Println("  delete <name> [--yes] [--force]              Delete a vault and all its memories")
		fmt.Println("  clear  <name> [--yes] [--force]              Remove all memories from a vault")
		fmt.Println("  clone  <source> <new-name>                   Clone a vault into a new vault")
		fmt.Println("  merge  <source> <target> [--delete-source] [--yes]  Merge source into target vault")
		return
	}
	switch args[0] {
	case "delete":
		runVaultDelete(args[1:])
	case "clear":
		runVaultClear(args[1:])
	case "clone":
		runVaultClone(args[1:])
	case "merge":
		runVaultMerge(args[1:])
	default:
		fmt.Printf("Unknown vault command: %q\n", args[0])
		fmt.Println("Available: delete, clear, clone, merge")
	}
}

func runVaultDelete(args []string) {
	name, yes, force := parseVaultArgs(args, "delete")
	if name == "" {
		return
	}
	if !yes && !confirmVaultAction(name, "delete") {
		fmt.Println("Cancelled.")
		return
	}
	doVaultRequestForce("DELETE",
		fmt.Sprintf("http://localhost:8475/api/admin/vaults/%s", url.PathEscape(name)),
		"Vault deleted.", force)
}

func runVaultClear(args []string) {
	name, yes, force := parseVaultArgs(args, "clear")
	if name == "" {
		return
	}
	if !yes && !confirmVaultAction(name, "clear") {
		fmt.Println("Cancelled.")
		return
	}
	doVaultRequestForce("POST",
		fmt.Sprintf("http://localhost:8475/api/admin/vaults/%s/clear", url.PathEscape(name)),
		"Vault cleared.", force)
}

// parseVaultArgs parses: <name> [--yes|-y] [--force|-f]
func parseVaultArgs(args []string, cmd string) (name string, yes bool, force bool) {
	for _, a := range args {
		switch a {
		case "--yes", "-y":
			yes = true
		case "--force", "-f":
			force = true
		default:
			if !strings.HasPrefix(a, "-") {
				name = a
			}
		}
	}
	if name == "" {
		fmt.Printf("Usage: muninn vault %s <vault-name> [--yes] [--force]\n", cmd)
	}
	return
}

// confirmVaultAction prompts the user to type the vault name to confirm.
func confirmVaultAction(name, action string) bool {
	fmt.Printf("\n  WARNING: This will %s vault %q and all its memories.\n", action, name)
	fmt.Printf("  Type the vault name to confirm: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	typed := strings.TrimSpace(scanner.Text())
	if typed != name {
		fmt.Printf("  Confirmation did not match %q.\n", name)
		return false
	}
	return true
}

func doVaultRequestForce(method, reqURL, successMsg string, force bool) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(method, reqURL, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	if force {
		req.Header.Set("X-Allow-Default", "true")
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error connecting to MuninnDB: %v\n", err)
		fmt.Println("Is muninn running? Try: muninn status")
		return
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNoContent:
		fmt.Println(" ", successMsg)
	case http.StatusNotFound:
		fmt.Println("  Vault not found.")
	case http.StatusConflict:
		if !force {
			fmt.Println("  Protected vault. Use --force to operate on the default vault.")
		} else {
			fmt.Println("  Protected vault. Cannot override.")
		}
	case http.StatusUnauthorized:
		fmt.Println("  Not authenticated. Open the web UI to manage vaults: http://localhost:8476")
	default:
		fmt.Printf("  Error: HTTP %d\n", resp.StatusCode)
	}
}

// ---------------------------------------------------------------------------
// vault clone
// ---------------------------------------------------------------------------

func runVaultClone(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: muninn vault clone <source-vault> <new-name>")
		return
	}
	source := args[0]
	newName := args[1]

	fmt.Printf("Cloning vault %q to %q...\n", source, newName)

	bodyBytes, err := json.Marshal(map[string]any{"new_name": newName})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	req, err := http.NewRequest("POST",
		fmt.Sprintf("http://localhost:8475/api/admin/vaults/%s/clone", url.PathEscape(source)),
		bytes.NewReader(bodyBytes))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error connecting to MuninnDB: %v\n", err)
		fmt.Println("Is muninn running? Try: muninn status")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		printHTTPError(resp)
		return
	}

	var result struct {
		JobID string `json:"job_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil || result.JobID == "" {
		fmt.Println("  Error: could not read job ID from response.")
		return
	}

	pollProgressBar(result.JobID, source)
}

// ---------------------------------------------------------------------------
// vault merge
// ---------------------------------------------------------------------------

func runVaultMerge(args []string) {
	var source, target string
	var deleteSource, yes bool

	for i, a := range args {
		switch {
		case a == "--delete-source":
			deleteSource = true
		case a == "--yes" || a == "-y":
			yes = true
		case source == "" && !strings.HasPrefix(a, "-"):
			source = a
		case target == "" && !strings.HasPrefix(a, "-") && i > 0:
			target = a
		}
	}

	if source == "" || target == "" {
		fmt.Println("Usage: muninn vault merge <source> <target> [--delete-source] [--yes]")
		return
	}

	if source == target {
		fmt.Fprintln(os.Stderr, "Error: cannot merge a vault into itself")
		os.Exit(1)
	}

	if !yes {
		fmt.Printf("\n  Merge Vault Wizard\n")
		fmt.Printf("  Source: %q\n", source)
		fmt.Printf("  Target: %q\n", target)
		if deleteSource {
			fmt.Printf("  Source vault will be deleted after merge.\n")
		}
		fmt.Printf("\n  Type 'merge' to confirm: ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		if strings.TrimSpace(scanner.Text()) != "merge" {
			fmt.Println("Cancelled.")
			return
		}
	}

	fmt.Printf("Merging %q into %q...\n", source, target)

	bodyBytes, err := json.Marshal(map[string]any{"target": target, "delete_source": deleteSource})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	req, err := http.NewRequest("POST",
		fmt.Sprintf("http://localhost:8475/api/admin/vaults/%s/merge-into", url.PathEscape(source)),
		bytes.NewReader(bodyBytes))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error connecting to MuninnDB: %v\n", err)
		fmt.Println("Is muninn running? Try: muninn status")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		printHTTPError(resp)
		return
	}

	var result struct {
		JobID string `json:"job_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil || result.JobID == "" {
		fmt.Println("  Error: could not read job ID from response.")
		return
	}

	pollProgressBar(result.JobID, source)
}

// ---------------------------------------------------------------------------
// progress bar
// ---------------------------------------------------------------------------

type statusSnap struct {
	Status       string  `json:"status"`
	Phase        string  `json:"phase"`
	CopyTotal    int64   `json:"copy_total"`
	CopyCurrent  int64   `json:"copy_current"`
	IndexTotal   int64   `json:"index_total"`
	IndexCurrent int64   `json:"index_current"`
	Pct          float64 `json:"pct"`
	Error        string  `json:"error,omitempty"`
}

// printHTTPError decodes and prints the JSON error body from a non-success response.
func printHTTPError(resp *http.Response) {
	var errResp struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil && errResp.Error.Message != "" {
		fmt.Printf("  Error: %s\n", errResp.Error.Message)
		return
	}
	fmt.Printf("  Error: HTTP %d\n", resp.StatusCode)
}

const pollTimeout = 30 * time.Minute

func pollProgressBar(jobID, vaultName string) {
	isTTY := isTerminal()
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	deadline := time.After(pollTimeout)

	for {
		select {
		case <-deadline:
			if isTTY {
				fmt.Println()
			}
			fmt.Printf("Timed out after %s waiting for job to complete.\n", pollTimeout)
			fmt.Printf("The job may still be running on the server.\n")
			fmt.Printf("Check status: muninn vault job-status %s\n", jobID)
			os.Exit(1)
		case <-ticker.C:
			snap := fetchJobStatus(jobID, vaultName)
			if snap == nil {
				fmt.Println("Job not found.")
				return
			}

			bar := renderBar(*snap)
			if isTTY {
				fmt.Printf("\r%s", bar)
			} else {
				fmt.Printf("%s\n", bar)
			}

			if snap.Status == "done" {
				if isTTY {
					fmt.Println()
				}
				fmt.Println("Done!")
				return
			}
			if snap.Status == "error" {
				if isTTY {
					fmt.Println()
				}
				fmt.Printf("Error: %s\n", snap.Error)
				return
			}
		}
	}
}

func fetchJobStatus(jobID, vaultName string) *statusSnap {
	u := fmt.Sprintf("http://localhost:8475/api/admin/vaults/%s/job-status?job_id=%s",
		url.PathEscape(vaultName), url.QueryEscape(jobID))
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(u)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil
	}
	var snap statusSnap
	if err := json.NewDecoder(resp.Body).Decode(&snap); err != nil {
		return nil
	}
	return &snap
}

func renderBar(snap statusSnap) string {
	pct := snap.Pct
	filled := int(pct / 5) // 20-char bar
	if filled > 20 {
		filled = 20
	}
	bar := strings.Repeat("█", filled) + strings.Repeat("░", 20-filled)
	phase := "Copying"
	current, total := snap.CopyCurrent, snap.CopyTotal
	if snap.Phase == "indexing" {
		phase = "Re-indexing"
		current, total = snap.IndexCurrent, snap.IndexTotal
	}
	return fmt.Sprintf("[%s] %5.1f%%  %s  (%d/%d)",
		bar, pct, phase, current, total)
}

func isTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}
