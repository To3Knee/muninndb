package main

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

func printHelp() {
	isTTY := term.IsTerminal(int(os.Stdout.Fd()))

	bold := func(s string) string {
		if isTTY {
			return "\033[1m" + s + "\033[0m"
		}
		return s
	}
	dim := func(s string) string {
		if isTTY {
			return "\033[2m" + s + "\033[0m"
		}
		return s
	}
	cyan := func(s string) string {
		if isTTY {
			return "\033[36m" + s + "\033[0m"
		}
		return s
	}

	fmt.Println()
	fmt.Println(bold("muninn") + " — a cognitive memory database")
	fmt.Println()

	fmt.Println(bold("QUICK START"))
	fmt.Println()
	fmt.Printf("  %s       %s\n", cyan("muninn"), dim("# status check / drop into shell"))
	fmt.Printf("  %s  %s\n", cyan("muninn init"), dim("# guided setup (or re-run to reconfigure)"))
	fmt.Printf("  %s %s\n", cyan("muninn start"), dim("# start all services in background"))
	fmt.Println()

	fmt.Println(bold("COMMANDS"))
	fmt.Println()
	fmt.Printf("  %-32s %s\n", cyan("muninn init"), "First-time setup wizard — connects your AI tools")
	fmt.Printf("  %-32s %s\n", cyan("muninn start"), "Start all services in background")
	fmt.Printf("  %-32s %s\n", cyan("muninn stop"), "Stop the running server")
	fmt.Printf("  %-32s %s\n", cyan("muninn restart"), "Stop and restart")
	fmt.Printf("  %-32s %s\n", cyan("muninn status"), "Show which services are running")
	fmt.Printf("  %-32s %s\n", cyan("muninn"), "Status check / drop into interactive shell")
	fmt.Printf("  %-32s %s\n", cyan("muninn shell"), "Interactive shell (alias: bare muninn when running)")
	fmt.Printf("  %-32s %s\n", cyan("muninn logs"), "Tail the server log file")
	fmt.Printf("  %-32s %s\n", cyan("muninn logs --last 50"), "Print last 50 lines and exit")
	fmt.Printf("  %-32s %s\n", cyan("muninn show vaults"), "List all vaults (requires server running)")
	fmt.Printf("  %-32s %s\n", cyan("muninn cluster"), "Cluster management (info, status, failover, add-node, remove-node)")
	fmt.Printf("  %-32s %s\n", cyan("muninn completion <shell>"), "Shell completion (bash/zsh/fish)")
	fmt.Printf("  %-32s %s\n", cyan("muninn upgrade"), "Check for and install updates")
	fmt.Printf("  %-32s %s\n", cyan("muninn help"), "Show this message")
	fmt.Println()

	fmt.Println(bold("SERVER FLAGS") + dim("  (used with: muninn start)"))
	fmt.Println()
	fmt.Printf("  %-28s %s\n", "--data <dir>", "Data directory (default: ~/.muninn/data)")
	fmt.Printf("  %-28s %s\n", "--mcp-addr <addr>", "MCP listen address (default: :8750)")
	fmt.Printf("  %-28s %s\n", "--mcp-token <tok>", "MCP bearer token for AI tool auth")
	fmt.Printf("  %-28s %s\n", "--dev", "Serve web assets from ./web (hot-reload, dev only)")
	fmt.Println()

	fmt.Println(bold("AI TOOL INTEGRATION") + dim("  (MCP — Model Context Protocol)"))
	fmt.Println()
	fmt.Println("  MuninnDB exposes an MCP server that AI tools connect to for memory.")
	fmt.Println("  Run " + cyan("muninn init") + " to configure Claude Desktop, Cursor, or Windsurf automatically.")
	fmt.Println()
	fmt.Println("  MCP endpoint: http://localhost:8750/mcp")
	fmt.Printf("  %-28s %s\n", "MUNINNDB_MCP_URL", "Override MCP server URL")
	fmt.Printf("  %-28s %s\n", "MUNINNDB_DATA", "Override default data directory")
	fmt.Println()

	fmt.Println(bold("PORTS"))
	fmt.Println()
	fmt.Printf("  %-8s %s\n", ":8474", "MBP  — binary protocol")
	fmt.Printf("  %-8s %s\n", ":8475", "REST — JSON API")
	fmt.Printf("  %-8s %s\n", ":8476", "UI   — web dashboard (http://localhost:8476)")
	fmt.Printf("  %-8s %s\n", ":8750", "MCP  — AI tool integration")
	fmt.Println()

	fmt.Println(bold("EMBEDDERS") + dim("  (optional — enable semantic similarity search)"))
	fmt.Println()
	fmt.Printf("  %-28s %s\n", "MUNINN_OLLAMA_URL", "Local Ollama embed model (e.g. ollama://localhost:11434/nomic-embed-text)")
	fmt.Printf("  %-28s %s\n", "MUNINN_OPENAI_KEY", "OpenAI embeddings API key (text-embedding-3-small, 1536d)")
	fmt.Printf("  %-28s %s\n", "MUNINN_VOYAGE_KEY", "Voyage AI embeddings API key (voyage-3, 1024d)")
	fmt.Println()

	fmt.Println(bold("LLM ENRICHMENT") + dim("  (optional — auto-extract entities, relationships, summaries)"))
	fmt.Println()
	fmt.Println("  Set MUNINN_ENRICH_URL to enable background LLM enrichment on every new memory.")
	fmt.Println("  One provider at a time. URL scheme selects the provider:")
	fmt.Println()
	fmt.Printf("  %-28s %s\n", "Ollama (local, no key):", "MUNINN_ENRICH_URL=ollama://localhost:11434/llama3.2")
	fmt.Printf("  %-28s %s\n", "OpenAI:", "MUNINN_ENRICH_URL=openai://gpt-4o-mini")
	fmt.Printf("  %-28s %s\n", "", "MUNINN_ENRICH_API_KEY=sk-...")
	fmt.Printf("  %-28s %s\n", "Anthropic:", "MUNINN_ENRICH_URL=anthropic://claude-haiku-4-5-20251001")
	fmt.Printf("  %-28s %s\n", "", "MUNINN_ANTHROPIC_KEY=sk-ant-...  (or MUNINN_ENRICH_API_KEY)")
	fmt.Println()
	fmt.Println("  Enrichment runs asynchronously and does not block memory writes.")
	fmt.Println()
}
