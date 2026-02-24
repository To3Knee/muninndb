package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"time"
)

func main() {
	debug.SetGCPercent(400)

	full := flag.Bool("full", false, "Load full Bible (default: NT only)")
	dataDir := flag.String("data", "", "Persistent data directory (default: temp dir)")
	kjvPath := flag.String("kjv", "testdata/bible/kjv.json", "Path to kjv.json")
	xrefPath := flag.String("xref", "testdata/bible/cross-refs.csv", "Path to cross-refs.csv (or .tsv)")
	resultsFile := flag.String("results-file", "", "File to append results to")
	seedCount := flag.Int("seeds", 100, "Number of seed verses to evaluate")
	minXRefs := flag.Int("min-xrefs", 5, "Minimum cross-references per seed verse")
	skipLoad := flag.Bool("skip-load", false, "Skip corpus load (reuse existing data dir)")
	importFrom := flag.String("import-from", "", "Import vault from .muninn file before eval (skips corpus load)")
	exportTo := flag.String("export-to", "", "Export vault to .muninn file after corpus load")
	flag.Parse()

	// Determine mode label for the report
	mode := "fts"
	if *full {
		mode = "fts-full"
	}

	// Setup storage directory
	if *dataDir == "" {
		tmpDir, err := os.MkdirTemp("", "muninndb-bible-eval-*")
		if err != nil {
			log.Fatalf("create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)
		*dataDir = tmpDir
	}

	fmt.Printf("MuninnDB Bible Eval Harness\n")
	fmt.Printf("════════════════════════════════════════════════════════════════════\n")
	fmt.Printf("Mode:       %s\n", mode)
	fmt.Printf("Data dir:   %s\n", *dataDir)
	fmt.Printf("KJV file:   %s\n", *kjvPath)
	fmt.Printf("XRef file:  %s\n", *xrefPath)
	fmt.Printf("Seeds:      %d (min xrefs: %d)\n", *seedCount, *minXRefs)
	fmt.Printf("NT only:    %v\n\n", !*full)

	// Read corpus files
	kjvData, err := os.ReadFile(*kjvPath)
	if err != nil {
		log.Fatalf("read KJV file %q: %v\n(run scripts/eval-bible-setup.sh to download)", *kjvPath, err)
	}
	xrefData, err := os.ReadFile(*xrefPath)
	if err != nil {
		log.Fatalf("read xref file %q: %v\n(run scripts/eval-bible-setup.sh to download)", *xrefPath, err)
	}

	// Parse KJV corpus
	ntOnly := !*full
	corpus, err := parseKJV(kjvData, ntOnly)
	if err != nil {
		log.Fatalf("parse KJV: %v", err)
	}
	fmt.Printf("Corpus: %d verses loaded\n", len(corpus))

	// Parse cross-references
	xrefs, err := parseXRef(xrefData)
	if err != nil {
		log.Fatalf("parse cross-refs: %v", err)
	}
	fmt.Printf("Cross-refs: %d source verses mapped\n", len(xrefs))

	// Select evaluation seeds
	seeds := selectSeeds(xrefs, *minXRefs, *seedCount)
	fmt.Printf("Seeds selected: %d\n\n", len(seeds))

	// Initialise engine
	ctx := context.Background()
	ee, err := newEvalEngine(ctx, *dataDir)
	if err != nil {
		log.Fatalf("init engine: %v", err)
	}
	defer ee.close()

	var loadDur time.Duration

	// Fast-path: import vault from .muninn file
	if *importFrom != "" {
		fmt.Printf("Importing vault from %s...\n", *importFrom)
		f, openErr := os.Open(*importFrom)
		if openErr != nil {
			log.Fatalf("open import file: %v", openErr)
		}
		if importErr := ee.importVault(ctx, f); importErr != nil {
			f.Close()
			log.Fatalf("import vault: %v", importErr)
		}
		f.Close()
		fmt.Println("Vault imported. Skipping corpus load.")
	} else if !*skipLoad {
		// Standard corpus load
		fmt.Printf("Loading %d verses into vault 'bible'...\n", len(corpus))
		loadStart := time.Now()
		loaded := 0
		loadErrors := 0
		for i, req := range corpus {
			if _, writeErr := ee.writeVerse(ctx, req); writeErr != nil {
				loadErrors++
				if loadErrors <= 5 {
					log.Printf("  write error at index %d (%s): %v", i, req.Concept, writeErr)
				}
				continue
			}
			loaded++
			if (i+1)%1000 == 0 {
				elapsed := time.Since(loadStart).Seconds()
				fmt.Printf("  %d/%d loaded (%.0f writes/sec, %d errors)...\n",
					i+1, len(corpus), float64(i+1)/elapsed, loadErrors)
			}
		}
		loadDur = time.Since(loadStart)
		fmt.Printf("Loaded %d/%d verses in %v (%d errors)\n\n",
			loaded, len(corpus), loadDur.Round(time.Millisecond), loadErrors)

		// Give FTS worker time to settle
		fmt.Println("Waiting for FTS indexing to settle (2s)...")
		time.Sleep(2 * time.Second)

		// Export vault to .muninn file if requested
		if *exportTo != "" {
			fmt.Printf("Exporting vault to %s...\n", *exportTo)
			f, createErr := os.Create(*exportTo)
			if createErr != nil {
				log.Printf("create export file: %v", createErr)
			} else {
				if exportErr := ee.exportVault(ctx, f); exportErr != nil {
					log.Printf("export vault: %v", exportErr)
				}
				f.Close()
				fmt.Printf("Vault exported to %s\n\n", *exportTo)
			}
		}
	}

	// Build lookup map for quick text access
	corpusTexts := buildCorpusTextMap(corpus)

	// Phase 1: Retrieval Quality
	fmt.Printf("── Phase 1: Retrieval Quality ──────────────────────\n")
	p1 := RunPhase1(ctx, ee, seeds, xrefs, corpusTexts)
	fmt.Printf("Phase 1 complete: seeds=%d recall@10=%.4f ndcg@10=%.4f\n\n",
		p1.SeedsEvaluated, p1.RecallAtK, p1.NDCGAtK)

	// Phase 2: Cognitive Properties
	fmt.Printf("── Phase 2: Cognitive Properties ───────────────────\n")
	p2 := RunPhase2(ctx, ee)

	// Report
	writeReport(os.Stdout, p1, p2, mode, len(corpus), loadDur)

	if *resultsFile != "" {
		if saveErr := saveReport(*resultsFile, p1, p2, mode, len(corpus), loadDur); saveErr != nil {
			log.Printf("save report: %v", saveErr)
		} else {
			fmt.Printf("\nResults appended to: %s\n", *resultsFile)
		}
	}
}
