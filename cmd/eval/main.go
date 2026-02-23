package main

import (
	"context"
	"flag"
	"fmt"
	"hash/fnv"
	"log"
	"math"
	"math/rand"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/scrypster/muninndb/internal/cognitive"
	"github.com/scrypster/muninndb/internal/engine"
	"github.com/scrypster/muninndb/internal/engine/activation"
	"github.com/scrypster/muninndb/internal/engine/trigger"
	"github.com/scrypster/muninndb/internal/index/fts"
	hnswpkg "github.com/scrypster/muninndb/internal/index/hnsw"
	"github.com/scrypster/muninndb/internal/plugin"
	embedpkg "github.com/scrypster/muninndb/internal/plugin/embed"
	"github.com/scrypster/muninndb/internal/storage"
	"github.com/scrypster/muninndb/internal/transport/mbp"
)

func main() {
	debug.SetGCPercent(400)

	targetCount := flag.Int("count", 15000, "Target number of memories to load")
	skipLoad := flag.Bool("skip-load", false, "Skip loading memories (use existing data dir)")
	dataDir := flag.String("data", "", "Persistent data directory (default: temp dir)")
	verbose := flag.Bool("verbose", false, "Print detailed per-query results")
	mode := flag.String("mode", "fts", "Eval mode: 'fts' (hash embedder, nil HNSW), 'semantic' (Ollama + HNSW), or 'local' (bundled MiniLM + HNSW)")
	ollamaURL := flag.String("ollama", "ollama://localhost:11434/nomic-embed-text", "Ollama provider URL for semantic mode")
	flag.Parse()

	// Cap memories for semantic/local mode — embed time is the bottleneck
	if (*mode == "semantic" || *mode == "local") && *targetCount > 2000 {
		*targetCount = 2000
		fmt.Printf("Note: capping to 2000 memories in %s mode (embedding budget)\n", *mode)
	}

	// Setup storage directory
	var tmpDir string
	var err error
	if *dataDir == "" {
		tmpDir, err = os.MkdirTemp("", "muninndb-eval-*")
		if err != nil {
			log.Fatalf("create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)
		*dataDir = tmpDir
	}

	fmt.Printf("MuninnDB Relevance Evaluation\n")
	fmt.Printf("══════════════════════════════════════════════════\n")
	fmt.Printf("Mode:    %s\n", *mode)
	fmt.Printf("Data dir: %s\n", *dataDir)
	fmt.Printf("Target memories: %d\n\n", *targetCount)

	ctx := context.Background()

	// Initialize storage
	db, err := storage.OpenPebble(*dataDir, storage.DefaultOptions())
	if err != nil {
		log.Fatalf("open pebble: %v", err)
	}
	store := storage.NewPebbleStore(db, 100_000)
	ftsIdx := fts.New(db)

	// HNSW registry (always created; wired in for semantic mode only)
	hnswReg := hnswpkg.NewRegistry(db)

	// Choose embedder and HNSW index based on mode
	var embedder activation.Embedder
	var hnswIdx activation.HNSWIndex
	var useHNSW bool

	switch *mode {
	case "semantic":
		svc, initErr := embedpkg.NewEmbedService(*ollamaURL)
		if initErr != nil {
			log.Fatalf("create embed service: %v", initErr)
		}
		if initErr = svc.Init(ctx, plugin.PluginConfig{}); initErr != nil {
			log.Fatalf("init embed service: %v — is Ollama running at %s?", initErr, *ollamaURL)
		}
		fmt.Printf("Embedder: Ollama nomic-embed-text (dim=%d)\n\n", svc.Dimension())
		embedder = &ollamaEmbedder{svc: svc}
		hnswIdx = &hnswActAdapter{r: hnswReg}
		useHNSW = true

	case "local":
		svc, initErr := embedpkg.NewEmbedService("local://all-MiniLM-L6-v2")
		if initErr != nil {
			log.Fatalf("create local embed service: %v", initErr)
		}
		if initErr = svc.Init(ctx, plugin.PluginConfig{DataDir: *dataDir}); initErr != nil {
			log.Fatalf("init local embed service: %v — did you run `make fetch-assets`?", initErr)
		}
		fmt.Printf("Embedder: bundled all-MiniLM-L6-v2 INT8 ONNX (dim=%d)\n\n", svc.Dimension())
		embedder = &ollamaEmbedder{svc: svc} // reuse same adapter (Embed+Tokenize contract)
		hnswIdx = &hnswActAdapter{r: hnswReg}
		useHNSW = true

	default: // "fts"
		embedder = &hashEmbedder{}
		hnswIdx = nil // weight renorm kicks in when nil
		useHNSW = false
		fmt.Printf("Embedder: hash (deterministic, 384-dim)\n")
		fmt.Printf("HNSW:     disabled — semantic weight renormalized to FTS+Decay\n\n")
	}

	actEngine := activation.New(store, &ftsAdapter{ftsIdx}, hnswIdx, embedder)
	trigSystem := trigger.New(store, &ftsTrigAdapter{ftsIdx}, nil, embedder)

	hebbianWorker := cognitive.NewHebbianWorker(&benchHebbianAdapter{store})
	decayWorker := cognitive.NewDecayWorker(&benchDecayAdapter{store})
	contradictWorker := cognitive.NewContradictWorker(&benchContradictAdapter{store})
	confidenceWorker := cognitive.NewConfidenceWorker(&benchConfidenceAdapter{store})

	benchCtx, benchCancel := context.WithCancel(context.Background())
	go decayWorker.Worker.Run(benchCtx)
	go contradictWorker.Worker.Run(benchCtx)
	go confidenceWorker.Worker.Run(benchCtx)

	eng := engine.NewEngine(store, nil, ftsIdx, actEngine, trigSystem,
		hebbianWorker, decayWorker,
		contradictWorker.Worker, confidenceWorker.Worker,
		embedder, hnswReg,
	)
	defer func() {
		benchCancel()
		hebbianWorker.Stop()
		eng.Stop()
		store.Close()
	}()

	vault := "knowledge"

	// --- Phase 1: Load memories ---
	if !*skipLoad {
		memories := expandedMemories()
		if len(memories) > *targetCount {
			memories = memories[:*targetCount]
		}

		fmt.Printf("Loading %d memories into vault '%s'...\n", len(memories), vault)
		start := time.Now()
		loaded := 0
		errors := 0
		hnswErrors := 0

		// Pre-compute vault prefix once (same for all writes)
		ws := store.ResolveVaultPrefix(vault)

		for i, mem := range memories {
			mem.Vault = vault
			resp, writeErr := eng.Write(ctx, &mem)
			if writeErr != nil {
				errors++
				if errors <= 5 {
					log.Printf("write error at index %d: %v", i, writeErr)
				}
				continue
			}
			loaded++

			// If HNSW mode: embed and index the written memory
			if useHNSW {
				id, parseErr := storage.ParseULID(resp.ID)
				if parseErr == nil {
					text := mem.Concept
					if mem.Content != "" {
						text += ". " + mem.Content
					}
					vec, embedErr := embedder.Embed(ctx, []string{text})
					if embedErr == nil {
						if insertErr := hnswReg.Insert(ctx, ws, [16]byte(id), vec); insertErr != nil {
							hnswErrors++
						}
					} else {
						hnswErrors++
					}
				}
			}

			if (i+1)%500 == 0 {
				elapsed := time.Since(start).Seconds()
				rate := float64(i+1) / elapsed
				if useHNSW {
					fmt.Printf("  %d/%d loaded (%.0f writes/sec, %d errors, %d hnsw errors, %d vectors indexed)...\n",
						i+1, len(memories), rate, errors, hnswErrors, hnswReg.TotalVectors())
				} else {
					fmt.Printf("  %d/%d loaded (%.0f writes/sec, %d errors)...\n",
						i+1, len(memories), rate, errors)
				}
			}
		}

		elapsed := time.Since(start)
		if useHNSW {
			fmt.Printf("\n✓ Loaded %d memories in %v (%.0f writes/sec, %d HNSW vectors)\n\n",
				loaded, elapsed.Round(time.Millisecond), float64(loaded)/elapsed.Seconds(), hnswReg.TotalVectors())
		} else {
			fmt.Printf("\n✓ Loaded %d memories in %v (%.0f writes/sec, %d errors)\n\n",
				loaded, elapsed.Round(time.Millisecond), float64(loaded)/elapsed.Seconds(), errors)
		}

		// Give FTS worker time to index
		fmt.Println("Waiting for FTS indexing to settle...")
		time.Sleep(2 * time.Second)
	}

	// --- Phase 2: Activation queries ---
	fmt.Printf("Running %d activation queries...\n\n", len(evalQueries))

	results := make([]queryResult, 0, len(evalQueries))

	for _, q := range evalQueries {
		resp, err := eng.Activate(ctx, &mbp.ActivateRequest{
			Context:    q.Context,
			MaxResults: 10,
			Vault:      vault,
			BriefMode:  "off",
		})

		qr := queryResult{query: q}
		if err != nil {
			qr.err = err
		} else {
			qr.resp = resp
			qr.relevantCount = countRelevant(resp, q.ExpectedTags)
		}
		results = append(results, qr)
	}

	// --- Phase 3: Report ---
	printReport(results, *verbose, *mode)

	// --- Phase 4: Hebbian demonstration (FTS mode only) ---
	if *mode == "fts" {
		runHebbianDemo(ctx, eng, vault)
	}
}

// queryResult holds activation response + relevance assessment
type queryResult struct {
	query         evalQuery
	resp          *mbp.ActivateResponse
	relevantCount int
	err           error
}

func printReport(results []queryResult, verbose bool, mode string) {
	fmt.Printf("═══════════════════════════════════════════════════════════════\n")
	fmt.Printf("RELEVANCE EVALUATION REPORT  [mode: %s]\n", mode)
	fmt.Printf("═══════════════════════════════════════════════════════════════\n\n")

	totalQueries := 0
	totalRelevant := 0
	totalResults := 0
	successQueries := 0
	zeroResultQueries := 0

	for _, qr := range results {
		totalQueries++
		if qr.err != nil {
			fmt.Printf("✗ [ERROR] %q: %v\n\n", strings.Join(qr.query.Context, " "), qr.err)
			continue
		}

		n := len(qr.resp.Activations)
		totalResults += n
		totalRelevant += qr.relevantCount

		if n == 0 {
			zeroResultQueries++
		} else {
			successQueries++
		}

		// Relevance rate for this query
		relRate := 0.0
		if n > 0 {
			relRate = float64(qr.relevantCount) / float64(n)
		}

		statusIcon := "✓"
		if n == 0 {
			statusIcon = "✗"
		} else if relRate < 0.5 {
			statusIcon = "~"
		}

		fmt.Printf("%s [%s] %q\n", statusIcon, qr.query.Domain, strings.Join(qr.query.Context, " | "))
		fmt.Printf("   Results: %d  |  Relevant: %d/%d (%.0f%%)  |  Top score: %.3f\n",
			n, qr.relevantCount, n, relRate*100,
			topScore(qr.resp))

		if verbose && n > 0 {
			fmt.Printf("   Top results:\n")
			limit := 5
			if len(qr.resp.Activations) < limit {
				limit = len(qr.resp.Activations)
			}
			for i, act := range qr.resp.Activations[:limit] {
				snippet := act.Content
				if len(snippet) > 100 {
					snippet = snippet[:97] + "..."
				}
				fmt.Printf("     %d. [%.3f] %s\n", i+1, act.Score, act.Concept)
				fmt.Printf("        %s\n", snippet)
			}
		}
		fmt.Println()
	}

	// Summary statistics
	fmt.Printf("═══════════════════════════════════════════════════════════════\n")
	fmt.Printf("SUMMARY\n")
	fmt.Printf("═══════════════════════════════════════════════════════════════\n")
	fmt.Printf("Total queries:          %d\n", totalQueries)
	fmt.Printf("Queries with results:   %d (%.0f%%)\n", successQueries, pct(successQueries, totalQueries))
	fmt.Printf("Zero-result queries:    %d (%.0f%%)\n", zeroResultQueries, pct(zeroResultQueries, totalQueries))
	fmt.Printf("Avg results/query:      %.1f\n", avg(totalResults, successQueries))

	overallRelRate := 0.0
	if totalResults > 0 {
		overallRelRate = float64(totalRelevant) / float64(totalResults) * 100
	}
	fmt.Printf("Overall relevance rate: %.1f%%  (%d/%d results tagged relevant)\n",
		overallRelRate, totalRelevant, totalResults)

	// Domain breakdown
	domainStats := make(map[string][2]int)
	for _, qr := range results {
		if qr.err != nil || len(qr.resp.Activations) == 0 {
			continue
		}
		s := domainStats[qr.query.Domain]
		s[0] += qr.relevantCount
		s[1] += len(qr.resp.Activations)
		domainStats[qr.query.Domain] = s
	}

	fmt.Printf("\nDomain breakdown:\n")
	for domain, s := range domainStats {
		if s[1] > 0 {
			fmt.Printf("  %-25s %d/%d relevant (%.0f%%)\n",
				domain+":", s[0], s[1], float64(s[0])/float64(s[1])*100)
		}
	}

	fmt.Printf("\n")
	if overallRelRate >= 70 {
		fmt.Printf("★ VERDICT: MuninnDB produces HIGH-QUALITY relevant results (%.0f%% relevance)\n", overallRelRate)
	} else if overallRelRate >= 50 {
		fmt.Printf("◆ VERDICT: MuninnDB produces GOOD relevant results (%.0f%% relevance)\n", overallRelRate)
	} else if overallRelRate >= 30 {
		fmt.Printf("◇ VERDICT: MuninnDB produces ACCEPTABLE results (%.0f%% relevance) — may need tuning\n", overallRelRate)
	} else {
		fmt.Printf("✗ VERDICT: Relevance needs improvement (%.0f%%) — check FTS indexing and scoring weights\n", overallRelRate)
	}
}

func topScore(resp *mbp.ActivateResponse) float64 {
	if resp == nil || len(resp.Activations) == 0 {
		return 0
	}
	return float64(resp.Activations[0].Score)
}

func countRelevant(resp *mbp.ActivateResponse, expectedTags []string) int {
	if resp == nil {
		return 0
	}
	count := 0
	for _, act := range resp.Activations {
		if isRelevant(act, expectedTags) {
			count++
		}
	}
	return count
}

// isRelevant checks if an activation result is topically relevant to the query.
// A result is relevant if its concept or content contains any of the expected domain tags.
func isRelevant(act mbp.ActivationItem, expectedTags []string) bool {
	text := strings.ToLower(act.Concept + " " + act.Content)
	for _, tag := range expectedTags {
		if strings.Contains(text, strings.ToLower(tag)) {
			return true
		}
	}
	return false
}

func pct(a, b int) float64 {
	if b == 0 {
		return 0
	}
	return float64(a) / float64(b) * 100
}

func avg(total, count int) float64 {
	if count == 0 {
		return 0
	}
	return float64(total) / float64(count)
}

// runHebbianDemo demonstrates Hebbian associative learning.
// It measures HebbianBoost scores before and after a "warm-up" phase of repeated
// activations on a topic, proving that repeated co-activation builds association weights
// that boost related memories on subsequent queries.
func runHebbianDemo(ctx context.Context, eng *engine.Engine, vault string) {
	fmt.Printf("\n═══════════════════════════════════════════════════════════════\n")
	fmt.Printf("HEBBIAN ASSOCIATIVE LEARNING DEMONSTRATION\n")
	fmt.Printf("═══════════════════════════════════════════════════════════════\n\n")
	fmt.Printf("Theory: Memories that fire together wire together.\n")
	fmt.Printf("Test: Repeated activation of a topic should boost related memories\n")
	fmt.Printf("      on subsequent queries via accumulated Hebbian weights.\n\n")

	// Pick a probe query and a warm-up topic
	probeCtx := []string{"quantum entanglement spooky action distance"}
	warmupTopics := [][]string{
		{"quantum entanglement spooky action distance"},
		{"special relativity time dilation speed of light"},
		{"thermodynamics entropy heat engine Carnot"},
		{"quantum mechanics wave function probability"},
		{"black hole event horizon singularity Hawking radiation"},
	}

	// Measure baseline HebbianBoost (before any warm-up)
	baseResp, err := eng.Activate(ctx, &mbp.ActivateRequest{
		Context: probeCtx, MaxResults: 10, Vault: vault, BriefMode: "off",
	})
	if err != nil {
		fmt.Printf("Baseline activation error: %v\n", err)
		return
	}
	baseHebbian := avgHebbianBoost(baseResp)
	baseScore := topScore(baseResp)
	fmt.Printf("BEFORE warm-up:\n")
	fmt.Printf("  Query: %q\n", strings.Join(probeCtx, " "))
	fmt.Printf("  Top score:      %.4f\n", baseScore)
	fmt.Printf("  Avg Hebbian:    %.6f (baseline — no prior activations)\n\n", baseHebbian)

	// Warm-up: run related topics repeatedly to build association log
	fmt.Printf("Running %d warm-up activations across related physics topics...\n", len(warmupTopics)*3)
	for round := 0; round < 3; round++ {
		for _, topic := range warmupTopics {
			_, _ = eng.Activate(ctx, &mbp.ActivateRequest{
				Context: topic, MaxResults: 10, Vault: vault, BriefMode: "off",
			})
		}
	}

	// Re-measure after warm-up
	afterResp, err := eng.Activate(ctx, &mbp.ActivateRequest{
		Context: probeCtx, MaxResults: 10, Vault: vault, BriefMode: "off",
	})
	if err != nil {
		fmt.Printf("Post-warmup activation error: %v\n", err)
		return
	}
	afterHebbian := avgHebbianBoost(afterResp)
	afterScore := topScore(afterResp)

	fmt.Printf("\nAFTER %d warm-up activations:\n", len(warmupTopics)*3)
	fmt.Printf("  Query: %q\n", strings.Join(probeCtx, " "))
	fmt.Printf("  Top score:      %.4f  (delta: %+.4f)\n", afterScore, afterScore-baseScore)
	fmt.Printf("  Avg Hebbian:    %.6f  (delta: %+.6f)\n", afterHebbian, afterHebbian-baseHebbian)

	if afterHebbian > baseHebbian {
		fmt.Printf("\n★ HEBBIAN LEARNING CONFIRMED: Repeated co-activation built association\n")
		fmt.Printf("  weights that boosted related memories. HebbianBoost rose by %.4f%%\n",
			(afterHebbian-baseHebbian)*100)
	} else if afterHebbian == baseHebbian {
		fmt.Printf("\n◆ Hebbian boost unchanged — associations may not yet be wired (requires\n")
		fmt.Printf("  cross-memory association records from the contradiction/autoassoc workers)\n")
	} else {
		fmt.Printf("\n◇ Note: Unexpected result — Hebbian boost lower after warm-up\n")
	}

	// Show score component breakdown for top results
	fmt.Printf("\nTop result score components after warm-up:\n")
	fmt.Printf("  %-40s  FTS     Decay   Hebbian Access  Recency\n", "Concept")
	fmt.Printf("  %-40s  ------  ------  ------  ------  -------\n", strings.Repeat("-", 40))
	limit := 5
	if len(afterResp.Activations) < limit {
		limit = len(afterResp.Activations)
	}
	for _, act := range afterResp.Activations[:limit] {
		concept := act.Concept
		if len(concept) > 40 {
			concept = concept[:37] + "..."
		}
		sc := act.ScoreComponents
		fmt.Printf("  %-40s  %.4f  %.4f  %.4f  %.4f  %.4f\n",
			concept, sc.FullTextRelevance, sc.DecayFactor,
			sc.HebbianBoost, sc.AccessFrequency, sc.Recency)
	}
	fmt.Printf("\n")
}

func avgHebbianBoost(resp *mbp.ActivateResponse) float64 {
	if resp == nil || len(resp.Activations) == 0 {
		return 0
	}
	var sum float64
	for _, act := range resp.Activations {
		sum += float64(act.ScoreComponents.HebbianBoost)
	}
	return sum / float64(len(resp.Activations))
}

// hashEmbedder produces deterministic 384-dim unit vectors from text.
type hashEmbedder struct{}

func (e *hashEmbedder) Embed(_ context.Context, texts []string) ([]float32, error) {
	const dims = 384
	vec := make([]float64, dims)
	for _, text := range texts {
		for _, word := range strings.Fields(strings.ToLower(text)) {
			h := fnv.New64a()
			h.Write([]byte(word))
			rng := rand.New(rand.NewSource(int64(h.Sum64()))) //nolint:gosec
			for i := range vec {
				vec[i] += rng.NormFloat64()
			}
		}
	}
	var norm float64
	for _, v := range vec {
		norm += v * v
	}
	norm = math.Sqrt(norm)
	out := make([]float32, dims)
	if norm > 0 {
		for i, v := range vec {
			out[i] = float32(v / norm)
		}
	}
	return out, nil
}

func (e *hashEmbedder) Tokenize(text string) []string {
	return strings.Fields(strings.ToLower(text))
}

// ollamaEmbedder wraps EmbedService to implement activation.Embedder.
// Multiple context strings are joined before embedding so HNSW gets a single
// dim-sized vector (not a concatenated flat array).
type ollamaEmbedder struct {
	svc *embedpkg.EmbedService
}

func (e *ollamaEmbedder) Embed(ctx context.Context, texts []string) ([]float32, error) {
	combined := strings.Join(texts, ". ")
	return e.svc.Embed(ctx, []string{combined})
}

func (e *ollamaEmbedder) Tokenize(text string) []string {
	return strings.Fields(strings.ToLower(text))
}

// hnswActAdapter adapts hnsw.Registry to activation.HNSWIndex.
type hnswActAdapter struct{ r *hnswpkg.Registry }

func (a *hnswActAdapter) Search(ctx context.Context, ws [8]byte, vec []float32, topK int) ([]activation.ScoredID, error) {
	results, err := a.r.Search(ctx, ws, vec, topK)
	if err != nil {
		return nil, err
	}
	out := make([]activation.ScoredID, len(results))
	for i, r := range results {
		out[i] = activation.ScoredID{ID: storage.ULID(r.ID), Score: r.Score}
	}
	return out, nil
}

type ftsAdapter struct{ idx *fts.Index }

func (a *ftsAdapter) Search(ctx context.Context, ws [8]byte, query string, topK int) ([]activation.ScoredID, error) {
	results, err := a.idx.Search(ctx, ws, query, topK)
	if err != nil {
		return nil, err
	}
	out := make([]activation.ScoredID, len(results))
	for i, r := range results {
		out[i] = activation.ScoredID{ID: storage.ULID(r.ID), Score: r.Score}
	}
	return out, nil
}

type ftsTrigAdapter struct{ idx *fts.Index }

func (a *ftsTrigAdapter) Search(ctx context.Context, ws [8]byte, query string, topK int) ([]trigger.ScoredID, error) {
	results, err := a.idx.Search(ctx, ws, query, topK)
	if err != nil {
		return nil, err
	}
	out := make([]trigger.ScoredID, len(results))
	for i, r := range results {
		out[i] = trigger.ScoredID{ID: storage.ULID(r.ID), Score: r.Score}
	}
	return out, nil
}
