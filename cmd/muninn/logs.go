package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func logFilePath() string {
	return filepath.Join(defaultDataDir(), "muninn.log")
}

func runLogs(args []string) {
	fs := flag.NewFlagSet("logs", flag.ExitOnError)
	last := fs.Int("last", 0, "Print last N lines and exit")
	level := fs.String("level", "", "Filter by log level: debug, info, warn, error")
	fs.Parse(args)

	path := logFilePath()

	if *last > 0 {
		printLastN(path, *last, *level)
		return
	}

	// Capture os.Stdout/Stderr at the callsite so the tailLog goroutine never
	// reads the globals; tests redirect os.Stdout concurrently which would race.
	tailLog(path, *level, os.Stdout, os.Stderr)
}

// printLastN reads the last N lines from the log file (filtered by level if set).
func printLastN(path string, n int, levelFilter string) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("  No log file found at", path)
			fmt.Println("  Start muninn to begin logging: muninn start")
			return
		}
		fmt.Fprintf(os.Stderr, "Error opening log: %v\n", err)
		return
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if levelFilter == "" || matchesLevel(line, levelFilter) {
			lines = append(lines, line)
		}
	}

	start := 0
	if len(lines) > n {
		start = len(lines) - n
	}
	for _, l := range lines[start:] {
		fmt.Println(l)
	}
}

// tailLog continuously tails the log file until Ctrl+C.
// out and errOut are passed in by the caller (never read from os.Stdout/os.Stderr
// directly) so that concurrent tests that redirect those globals don't race.
func tailLog(path string, levelFilter string, out, errOut io.Writer) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintln(out, "  No log file found at", path)
			fmt.Fprintln(out, "  Start muninn to begin logging: muninn start")
			return
		}
		fmt.Fprintf(errOut, "Error opening log: %v\n", err)
		return
	}
	defer f.Close()

	fmt.Fprintf(out, "  tailing %s  (Ctrl+C to stop)\n", path)
	if levelFilter != "" {
		fmt.Fprintf(out, "  filter: %s\n", levelFilter)
	}
	fmt.Fprintln(out, "  "+strings.Repeat("─", 60))

	// Seek to end, then tail new lines
	f.Seek(0, io.SeekEnd)
	reader := bufio.NewReader(f)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			time.Sleep(200 * time.Millisecond)
			continue
		}
		line = strings.TrimRight(line, "\n")
		if levelFilter == "" || matchesLevel(line, levelFilter) {
			fmt.Fprintln(out, line)
		}
	}
}

func matchesLevel(line, level string) bool {
	return strings.Contains(strings.ToUpper(line), strings.ToUpper(level))
}
