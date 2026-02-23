package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		runDefault()
		return
	}

	// --daemon flag: run server inline (forked by runStart)
	for i, arg := range os.Args[1:] {
		if arg == "--daemon" {
			os.Args = append(os.Args[:i+1], os.Args[i+2:]...)
			runServer()
			return
		}
	}

	sub := parseSubcommand(os.Args[1:])
	switch sub {
	case "":
		runDefault()
	case "init":
		runInit()
	case "shell":
		runShell()
	case "start":
		runStart(true)
	case "start:web":
		runStartService("web")
	case "stop":
		runStop()
	case "stop:web":
		runStopService("web")
	case "status":
		runStatus()
	case "upgrade":
		runUpgrade(os.Args[2:])
	case "restart":
		runStop()
		runStart(true)
	case "show:vaults":
		runShowVaults()
	case "logs":
		runLogs(os.Args[2:])
	case "cluster":
		runCluster(os.Args[2:])
	case "completion:bash":
		printCompletion("bash")
	case "completion:zsh":
		printCompletion("zsh")
	case "completion:fish":
		printCompletion("fish")
	case "completion":
		fmt.Fprintln(os.Stderr, "Usage: muninn completion <bash|zsh|fish>")
		os.Exit(1)
	case "help", "--help", "-h":
		printHelp()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %q\n\n", sub)
		fmt.Fprintln(os.Stderr, "Run 'muninn help' to see all commands.")
		os.Exit(1)
	}
}

// runDefault is what happens when the user types bare `muninn`.
// Three states:
//  1. No data dir → first time → launch wizard
//  2. Data dir exists, not running → show status + exit
//  3. Running → show status flash + drop into shell
func runDefault() {
	_, err := os.Stat(defaultDataDir())
	if os.IsNotExist(err) {
		// First run — launch wizard directly
		runInit()
		return
	}

	// Show status to determine state
	state := printStatusDisplay(true)

	switch state {
	case stateStopped:
		// hints already printed by printStatusDisplay
	case stateDegraded:
		// fix command already printed by printStatusDisplay
	case stateRunning:
		// Drop into shell
		runShell()
	}
}
