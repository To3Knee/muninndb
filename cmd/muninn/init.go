package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// version is set at build time via -ldflags "-X main.version=vX.Y.Z"
var version string

// muninnVersion returns the binary version string. Falls back to "dev".
func muninnVersion() string {
	if version != "" {
		return version
	}
	return "dev"
}

// toolChoice represents an AI tool in the wizard selection list.
type toolChoice struct {
	key         string // internal key: "claude", "cursor", etc.
	displayName string // shown in wizard
	configPath  string // path detected (empty if not found or manual-only)
	detected    bool   // true if config path exists on disk
	selected    bool   // true = will be configured
}

// detectInstalledTools scans known config paths and returns toolChoices.
// Detected tools are pre-selected.
func detectInstalledTools() []toolChoice {
	tools := []toolChoice{
		{key: "claude", displayName: "Claude Desktop", configPath: claudeDesktopConfigPath()},
		{key: "claude-code", displayName: "Claude Code / CLI", configPath: claudeCodeConfigPath()},
		{key: "cursor", displayName: "Cursor", configPath: cursorConfigPath()},
		{key: "openclaw", displayName: "OpenClaw", configPath: openClawConfigPath()},
		{key: "windsurf", displayName: "Windsurf", configPath: windsurfConfigPath()},
		{key: "vscode", displayName: "VS Code", configPath: ""},
		{key: "manual", displayName: "Other / manual config", configPath: ""},
	}
	for i, t := range tools {
		if t.configPath != "" {
			if _, err := os.Stat(t.configPath); err == nil {
				tools[i].detected = true
				tools[i].selected = true
			}
		}
	}
	return tools
}

// runInit runs the first-time onboarding wizard (or non-interactive setup via flags).
func runInit() {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	toolFlag := fs.String("tool", "", "AI tools to configure, comma-separated: claude,claude-code,cursor,openclaw,windsurf,vscode,manual")
	tokenFlag := fs.String("token", "", "Use this specific token (skip generation)")
	noToken := fs.Bool("no-token", false, "Disable token authentication (open MCP endpoint)")
	noStart := fs.Bool("no-start", false, "Skip starting the server")
	yes := fs.Bool("yes", false, "Accept all defaults non-interactively")

	// Strip "init" from args before parsing
	args := os.Args[2:]
	fs.Parse(args)

	mcpURL := "http://localhost:8750/mcp"
	isInteractive := term.IsTerminal(int(os.Stdin.Fd()))

	if !isInteractive && !*yes && *toolFlag == "" {
		fmt.Fprintln(os.Stderr, `muninn init requires an interactive terminal.
For non-interactive setup, use flags:

  muninn init --tool claude --yes
  muninn init --tool cursor,claude --no-token --yes
  muninn init --yes   (manual instructions only)

  --tool <tools>   Comma-separated: claude, cursor, openclaw, windsurf, vscode, manual
  --token <tok>    Use specific token
  --no-token       Open MCP (no auth)
  --no-start       Skip starting server
  --yes            Accept defaults, non-interactive`)
		os.Exit(1)
	}

	if isInteractive && *toolFlag == "" && !*yes {
		runInteractiveInit(mcpURL, tokenFlag, noToken, noStart)
		return
	}

	runNonInteractiveInit(mcpURL, *toolFlag, *tokenFlag, *noToken, *noStart, *yes)
}

func runInteractiveInit(mcpURL string, tokenFlag *string, noToken *bool, noStart *bool) {
	printWelcomeBanner()

	// Step 1: Tool detection + multi-select
	tools := detectInstalledTools()
	fmt.Println("  Scanning for AI tools...")
	fmt.Println()
	fmt.Println("  Which AI tools would you like to configure?")
	fmt.Println("  (enter numbers to change selection, Enter to confirm)")
	fmt.Println()

	selectedTools := runToolMultiSelect(tools)

	// Step 2: Embedder selection
	fmt.Println()
	fmt.Println("  Which embedder should muninn use for memory search?")
	fmt.Println()
	fmt.Println("    ▶  1)  Local (bundled)  ·  offline, no setup required   (recommended)")
	fmt.Println("       2)  Ollama           ·  self-hosted")
	fmt.Println("       3)  OpenAI           ·  cloud, requires API key")
	fmt.Println("       4)  Voyage           ·  cloud, requires API key")
	fmt.Println()
	fmt.Print("  Choice [1]: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	embedderChoice := strings.TrimSpace(scanner.Text())
	if embedderChoice == "" {
		embedderChoice = "1"
	}
	printEmbedderNote(embedderChoice)

	// Auto: generate token (no prompt)
	var token string
	if !*noToken {
		if *tokenFlag != "" {
			token = *tokenFlag
		} else {
			dataDir := defaultDataDir()
			var isNew bool
			var err error
			token, isNew, err = loadOrGenerateToken(dataDir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "\n  warning: could not generate token: %v\n", err)
			} else if isNew {
				fmt.Println()
				fmt.Println("  Generating MCP access token...  ✓")
			}
		}
	}

	// Configure selected tools
	if len(selectedTools) > 0 {
		fmt.Println()
		toolErrs := configureNamedTools(selectedTools, mcpURL, token)
		if len(toolErrs) > 0 {
			fmt.Println()
			fmt.Printf("  ⚠  %d tool(s) failed to configure — check errors above.\n", len(toolErrs))
			fmt.Println("     You can re-run: muninn init")
		}
	}

	// Auto: start server (no "start now?" prompt)
	if !*noStart {
		fmt.Println()
		runStart(true)
	}

	// Success message
	fmt.Println()
	fmt.Println("  ────────────────────────────────────────────────────")
	fmt.Println()
	fmt.Println("  You're live. Your AI tools now have memory.")
	fmt.Println()
	fmt.Println("  Try it → open Claude Code or Cursor and ask:")
	fmt.Println(`    "What do you remember about me?"`)
	fmt.Println()
	fmt.Println("  Browse memories → http://localhost:8476")
	fmt.Println()
	fmt.Println("  ────────────────────────────────────────────────────")
	fmt.Println()
}

// runToolMultiSelect renders a checkbox list and returns selected tool keys.
func runToolMultiSelect(tools []toolChoice) []string {
	for i, t := range tools {
		check := "○"
		suffix := ""
		if t.selected {
			check = "✓"
		}
		if t.detected && t.configPath != "" {
			suffix = "   detected  ·  " + t.configPath
		}
		fmt.Printf("    %s  %d)  %-18s%s\n", check, i+1, t.displayName, suffix)
	}
	fmt.Println()
	fmt.Print("  Enter numbers to change selection, or Enter to confirm: ")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	input := strings.TrimSpace(scanner.Text())

	if input == "" {
		var keys []string
		for _, t := range tools {
			if t.selected {
				keys = append(keys, t.key)
			}
		}
		return keys
	}

	// Parse new explicit selection
	selected := map[int]bool{}
	for _, part := range strings.FieldsFunc(input, func(r rune) bool { return r == ',' || r == ' ' }) {
		for _, c := range part {
			if c >= '1' && c <= '9' {
				n := int(c-'0') - 1
				if n < len(tools) {
					selected[n] = true
				}
			}
		}
	}
	var keys []string
	for i, t := range tools {
		if selected[i] {
			keys = append(keys, t.key)
		}
	}
	return keys
}

func printEmbedderNote(choice string) {
	switch choice {
	case "2":
		fmt.Println()
		fmt.Println("  Ollama selected. Set MUNINN_OLLAMA_URL to configure.")
		fmt.Println("  Example: MUNINN_OLLAMA_URL=ollama://localhost:11434/nomic-embed-text")
	case "3":
		fmt.Println()
		fmt.Println("  OpenAI selected. Set MUNINN_OPENAI_KEY to configure.")
	case "4":
		fmt.Println()
		fmt.Println("  Voyage selected. Set MUNINN_VOYAGE_KEY to configure.")
	default:
		// Local bundled — works out of the box, no message needed
	}
}

func runNonInteractiveInit(mcpURL, toolStr, tokenStr string, noToken, noStart, yes bool) {
	printWelcomeBanner()

	var token string
	if !noToken {
		if tokenStr != "" {
			token = tokenStr
		} else {
			dataDir := defaultDataDir()
			var err error
			token, _, err = loadOrGenerateToken(dataDir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "warning: could not generate token: %v\nContinuing without token.\n", err)
			}
		}
	}

	if !noStart {
		runStart(true)
		fmt.Println()
	}

	var tools []string
	if toolStr != "" {
		for _, t := range strings.FieldsFunc(toolStr, func(r rune) bool { return r == ',' || r == ' ' }) {
			tools = append(tools, strings.ToLower(strings.TrimSpace(t)))
		}
	}

	if len(tools) > 0 {
		fmt.Println("Configuring AI tools:")
		toolErrs := configureNamedTools(tools, mcpURL, token)
		if len(toolErrs) > 0 {
			fmt.Printf("\n  ⚠  %d tool(s) failed to configure:\n", len(toolErrs))
			for _, e := range toolErrs {
				fmt.Printf("     • %s\n", e)
			}
			fmt.Println("  Re-run: muninn init --tool <toolname>")
		}
	}

	fmt.Println()
	fmt.Println("muninn is running.")
	fmt.Println("  MCP endpoint:   http://localhost:8750/mcp")
	if token != "" {
		fmt.Println("  Token:          ~/.muninn/mcp.token")
	}
	fmt.Println("  Web UI:         http://localhost:8476")
	fmt.Println()
}

func printWelcomeBanner() {
	fmt.Println()
	fmt.Println("  ┌────────────────────────────────────────────────────┐")
	fmt.Println("  │                                                    │")
	fmt.Printf("  │   muninn  ·  cognitive memory database  %-7s  │\n", muninnVersion())
	fmt.Println("  │                                                    │")
	fmt.Println("  └────────────────────────────────────────────────────┘")
	fmt.Println()
	fmt.Println("  First time here — let's get you set up.")
	fmt.Println()
}

// configureTools maps numbered selections to tool configuration.
func configureTools(selected []int, mcpURL, token string) []string {
	var errs []string
	for _, n := range selected {
		switch n {
		case 1:
			if err := configureClaudeDesktop(mcpURL, token); err != nil {
				errs = append(errs, fmt.Sprintf("Claude Desktop: %v", err))
				fmt.Fprintf(os.Stderr, "  ✗ Claude Desktop: %v\n", err)
			}
		case 2:
			if err := configureCursor(mcpURL, token); err != nil {
				errs = append(errs, fmt.Sprintf("Cursor: %v", err))
				fmt.Fprintf(os.Stderr, "  ✗ Cursor: %v\n", err)
			}
		case 3:
			printVSCodeInstructions(mcpURL, token)
		case 4:
			if err := configureWindsurf(mcpURL, token); err != nil {
				errs = append(errs, fmt.Sprintf("Windsurf: %v", err))
				fmt.Fprintf(os.Stderr, "  ✗ Windsurf: %v\n", err)
			}
		case 5:
			printManualInstructions(mcpURL, token)
		}
	}
	return errs
}

// configureNamedTools configures AI tools by name.
func configureNamedTools(tools []string, mcpURL, token string) []string {
	var errs []string
	for _, t := range tools {
		switch t {
		case "claude", "claude-desktop":
			if err := configureClaudeDesktop(mcpURL, token); err != nil {
				errs = append(errs, fmt.Sprintf("Claude Desktop: %v", err))
				fmt.Fprintf(os.Stderr, "  ✗ Claude Desktop: %v\n", err)
			}
		case "claude-code", "claudecode":
			if err := configureClaudeCode(mcpURL, token); err != nil {
				errs = append(errs, fmt.Sprintf("Claude Code: %v", err))
				fmt.Fprintf(os.Stderr, "  ✗ Claude Code: %v\n", err)
			}
		case "cursor":
			if err := configureCursor(mcpURL, token); err != nil {
				errs = append(errs, fmt.Sprintf("Cursor: %v", err))
				fmt.Fprintf(os.Stderr, "  ✗ Cursor: %v\n", err)
			}
		case "vscode", "vs-code":
			printVSCodeInstructions(mcpURL, token)
		case "windsurf":
			if err := configureWindsurf(mcpURL, token); err != nil {
				errs = append(errs, fmt.Sprintf("Windsurf: %v", err))
				fmt.Fprintf(os.Stderr, "  ✗ Windsurf: %v\n", err)
			}
		case "openclaw":
			if err := configureOpenClaw(mcpURL, token); err != nil {
				errs = append(errs, fmt.Sprintf("OpenClaw: %v", err))
				fmt.Fprintf(os.Stderr, "  ✗ OpenClaw: %v\n", err)
			}
		case "manual", "other":
			printManualInstructions(mcpURL, token)
		default:
			fmt.Fprintf(os.Stderr, "  unknown tool: %q (use: claude, claude-code, cursor, vscode, windsurf, openclaw, manual)\n", t)
		}
	}
	return errs
}

// parseToolNumbers parses "1 2 3" or "1,2,3" into deduplicated ints 1-9.
func parseToolNumbers(input string) []int {
	seen := map[int]bool{}
	var result []int
	for _, part := range strings.FieldsFunc(input, func(r rune) bool { return r == ',' || r == ' ' }) {
		n := 0
		for _, c := range part {
			if c >= '1' && c <= '9' {
				n = int(c - '0')
				break
			}
		}
		if n > 0 && !seen[n] {
			seen[n] = true
			result = append(result, n)
		}
	}
	return result
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
