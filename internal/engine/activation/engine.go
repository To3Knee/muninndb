package activation

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"runtime"
	"sort"
	"strings"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/scrypster/muninndb/internal/storage"
)

// DefaultWeights for composite scoring.
type DefaultWeights struct {
	SemanticSimilarity float32
	FullTextRelevance  float32
	DecayFactor        float32
	HebbianBoost       float32
	AccessFrequency    float32
	Recency            float32
}

// Weights is an optional client override.
type Weights struct {
	SemanticSimilarity float32
	FullTextRelevance  float32
	DecayFactor        float32
	HebbianBoost       float32
	AccessFrequency    float32
	Recency            float32
}

type resolvedWeights struct {
	SemanticSimilarity float64
	FullTextRelevance  float64
	DecayFactor        float64
	HebbianBoost       float64
	AccessFrequency    float64
	Recency            float64
}

// Filter is a query filter applied in Phase 6.
type Filter struct {
	Field string
	Op    string
	Value interface{}
}

// ScoredID is a search result from an index.
type ScoredID struct {
	ID    storage.ULID
	Score float64
}

// ScoreComponents breaks down how a score was computed.
type ScoreComponents struct {
	SemanticSimilarity float64
	FullTextRelevance  float64
	DecayFactor        float64
	HebbianBoost       float64
	AccessFrequency    float64
	Recency            float64
	Confidence         float64
	Raw                float64
	Final              float64
}

// ScoredEngram is one activation result.
type ScoredEngram struct {
	Engram     *storage.Engram
	Score      float64
	Components ScoreComponents
	Why        string
	HopPath    []storage.ULID
	Dormant    bool
}

// ActivateRequest is the internal activation request form.
type ActivateRequest struct {
	VaultID          uint32
	VaultPrefix      [8]byte // if set, used directly instead of VaultID
	Context          []string
	Embedding        []float32
	Threshold        float64
	MaxResults       int
	HopDepth         int
	IncludeWhy       bool
	Weights          *Weights
	Filters          []Filter
	ReadOnly         bool   // when true, skip all write side-effects (observe mode)
	Profile          string // traversal profile override: "default"|"causal"|"confirmatory"|"adversarial"|"structural"
	VaultDefault     string // vault Plasticity default profile (set by engine.go, not by callers)
	StructuredFilter interface{} // *query.Filter, applied as final post-retrieval predicate
}

// ActivateResult is what the transport layer serializes and returns.
type ActivateResult struct {
	QueryID     string
	Activations []ScoredEngram
	TotalFound  int
	LatencyMs   float64
	ProfileUsed string // resolved traversal profile name (e.g. "default", "causal")
}

// ActivateResponseFrame is one streaming frame of results.
type ActivateResponseFrame struct {
	QueryID     string
	TotalFound  int
	LatencyMs   float64
	Activations []ScoredEngram
	Frame       int
	TotalFrames int
}

// ActivationStore is the storage interface required by the activation engine.
type ActivationStore interface {
	GetMetadata(ctx context.Context, wsPrefix [8]byte, ids []storage.ULID) ([]*storage.EngramMeta, error)
	GetEngrams(ctx context.Context, wsPrefix [8]byte, ids []storage.ULID) ([]*storage.Engram, error)
	GetAssociations(ctx context.Context, wsPrefix [8]byte, ids []storage.ULID, maxPerNode int) (map[storage.ULID][]storage.Association, error)
	RecentActive(ctx context.Context, wsPrefix [8]byte, topK int) ([]storage.ULID, error)
	VaultPrefix(vault string) [8]byte
	// EngramLastAccessNs returns the nanosecond timestamp of the last cache access for id.
	// Returns 0 if not in cache; callers fall back to eng.LastAccess.
	EngramLastAccessNs(wsPrefix [8]byte, id storage.ULID) int64
}

// FTSIndex is the full-text search interface.
type FTSIndex interface {
	Search(ctx context.Context, ws [8]byte, query string, topK int) ([]ScoredID, error)
}

// HNSWIndex is the vector search interface.
type HNSWIndex interface {
	Search(ctx context.Context, ws [8]byte, vec []float32, topK int) ([]ScoredID, error)
}

// Embedder converts text to a vector embedding.
type Embedder interface {
	Embed(ctx context.Context, texts []string) ([]float32, error)
	Tokenize(text string) []string
}

// logItem is a queued activation log entry for the async drainer.
// activations is the already-allocated result slice — the drainer extracts
// ids and scores off the hot path, keeping Run() allocation-free for logging.
type logItem struct {
	vaultID     uint32
	activations []ScoredEngram
}

// ActivationEngine is the main ACTIVATE pipeline orchestrator.
type ActivationEngine struct {
	store    ActivationStore
	fts      FTSIndex
	hnsw     HNSWIndex
	embedder Embedder
	assocLog *ActivationLog
	weights  DefaultWeights
	// logCh is a buffered channel for async activation log entries.
	// A single drainer goroutine owns all writes to assocLog, eliminating
	// Lock() contention against Phase 4's concurrent RLock() calls.
	logCh   chan logItem
	logDone chan struct{}
}

// New creates a new ActivationEngine.
func New(store ActivationStore, fts FTSIndex, hnsw HNSWIndex, embedder Embedder) *ActivationEngine {
	w := DefaultWeights{
		SemanticSimilarity: 0.35,
		FullTextRelevance:  0.25,
		DecayFactor:        0.20,
		HebbianBoost:       0.10,
		AccessFrequency:    0.05,
		Recency:            0.05,
	}
	// When HNSW is unavailable, semantic similarity is always 0.
	// Redistribute that 0.35 budget to active components so the score
	// range isn't compressed by 35% of dead weight.
	if hnsw == nil {
		// Redistribute the 0.35 semantic budget proportionally across all remaining
		// active dimensions: FTS(0.25) + Decay(0.20) + Hebbian(0.10) + AccessFreq(0.05) + Recency(0.05) = 0.65
		// This keeps weights summing to exactly 1.0 so raw scores stay in [0,1].
		scale := float32(1.0 / 0.65)
		w.SemanticSimilarity = 0
		w.FullTextRelevance = 0.25 * scale  // ≈ 0.385
		w.DecayFactor = 0.20 * scale        // ≈ 0.308
		w.HebbianBoost = 0.10 * scale       // ≈ 0.154
		w.AccessFrequency = 0.05 * scale    // ≈ 0.077
		w.Recency = 0.05 * scale            // ≈ 0.077
	}
	e := &ActivationEngine{
		store:    store,
		fts:      fts,
		hnsw:     hnsw,
		embedder: embedder,
		assocLog: &ActivationLog{},
		weights:  w,
		logCh:    make(chan logItem, 4096),
		logDone:  make(chan struct{}),
	}
	go e.drainLog()
	return e
}

// drainLog is the single goroutine that writes to assocLog.
// Serializes all activation log writes, eliminating Lock contention against
// Phase 4's concurrent RecentForVault RLock calls. Eventual consistency:
// the log may lag by ~1ms but Hebbian decay half-life is 3600s — irrelevant.
func (e *ActivationEngine) drainLog() {
	defer close(e.logDone)
	for item := range e.logCh {
		// Extract ids/scores in the drainer goroutine — off the hot path.
		ids := make([]storage.ULID, len(item.activations))
		scores := make([]float64, len(item.activations))
		for i, a := range item.activations {
			ids[i] = a.Engram.ID
			scores[i] = a.Score
		}
		e.assocLog.Record(LogEntry{
			VaultID:   item.vaultID,
			At:        time.Now(),
			EngramIDs: ids,
			Scores:    scores,
		})
	}
}

// Close shuts down the async activation log drainer. Call before engine shutdown.
func (e *ActivationEngine) Close() {
	close(e.logCh)
	<-e.logDone
}

const candidatesPerIndex = 30
const minFloor = float32(0.05)
const frameSize = 100

// Run executes the 6-phase ACTIVATE pipeline.
func (e *ActivationEngine) Run(ctx context.Context, req *ActivateRequest) (*ActivateResult, error) {
	start := time.Now()

	if req.MaxResults <= 0 {
		req.MaxResults = 10
	}
	if req.Threshold <= 0 {
		req.Threshold = 0.05
	}

	// Phase 1: embed + tokenize
	p1, err := e.phase1(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("activation phase1: %w", err)
	}

	// Phase 2: parallel candidate retrieval
	ws := req.VaultPrefix
	if ws == ([8]byte{}) {
		ws = e.vaultWorkspace(req.VaultID)
	}
	sets, err := e.phase2(ctx, req, p1, ws)
	if err != nil {
		return nil, fmt.Errorf("activation phase2: %w", err)
	}

	// Phase 3: RRF fusion
	fused := phase3RRF(sets)

	// Phase 4: Hebbian boost (always sequential — fast, in-memory ring buffer read).
	e.phase4HebbianBoost(ctx, ws, req.VaultID, fused)

	// Resolve traversal profile for Phase 5 and for audit logging.
	// Always resolved so ProfileUsed is set on every activation, regardless of HopDepth.
	profileName, profile := resolveProfile(req)

	// Phase 5: BFS traversal — run sequentially after Phase 4.
	// Goroutine spawn overhead (~3-5µs) is not worth it for the common case where
	// the corpus has no associations (empty GetAssociations returns immediately from cache).
	// The early-exit in phase5Traverse handles the no-association case efficiently.
	var traversed []traversedCandidate
	if req.HopDepth > 0 {
		traversed = e.phase5Traverse(ctx, req, ws, profile, fused)
	}

	// Phase 6: final scoring, filter, response
	result, err := e.phase6Score(ctx, req, ws, fused, traversed, p1)
	if err != nil {
		return nil, fmt.Errorf("activation phase6: %w", err)
	}

	result.LatencyMs = float64(time.Since(start).Microseconds()) / 1000.0
	result.ProfileUsed = profileName

	slog.Info("activation complete", "profile", profileName, "results", len(result.Activations), "elapsed_ms", result.LatencyMs)

	// Submit activation log entry to the async drainer — zero hot-path allocations.
	// The drainer extracts ids/scores off the critical path.
	// Non-blocking: drops if channel full (Hebbian half-life=3600s, 1ms lag is negligible).
	if !req.ReadOnly && len(result.Activations) > 0 {
		select {
		case e.logCh <- logItem{vaultID: req.VaultID, activations: result.Activations}:
			// Yield to allow the drainer goroutine to process immediately.
			// Cost: ~1-5ns (no syscall). Ensures test consistency and reduces
			// drainer queue depth in production under bursty load.
			runtime.Gosched()
		default: // channel full — drop; eventual consistency accepted
		}
	}

	return result, nil
}

func (e *ActivationEngine) vaultWorkspace(vaultID uint32) [8]byte {
	var ws [8]byte
	ws[0] = byte(vaultID >> 24)
	ws[1] = byte(vaultID >> 16)
	ws[2] = byte(vaultID >> 8)
	ws[3] = byte(vaultID)
	return ws
}

// phase1 embeds context and tokenizes query.
type phase1Result struct {
	embedding []float32
	tokens    []string
	queryStr  string
}

func (e *ActivationEngine) phase1(ctx context.Context, req *ActivateRequest) (*phase1Result, error) {
	result := &phase1Result{}
	result.queryStr = strings.Join(req.Context, " ")

	if e.embedder != nil {
		result.tokens = e.embedder.Tokenize(result.queryStr)
	}

	if len(req.Embedding) > 0 {
		result.embedding = req.Embedding
		return result, nil
	}

	// Only compute embedding if HNSW is available — the embedding is used
	// exclusively for vector search in phase2.  When HNSW is nil (common in
	// benchmarks and lightweight deployments), this avoids the hashEmbedder
	// CPU cost entirely (~13% of activation CPU).
	if e.embedder != nil && e.hnsw != nil {
		vec, err := e.embedder.Embed(ctx, req.Context)
		if err != nil {
			return nil, fmt.Errorf("phase1 embed: %w", err)
		}
		result.embedding = vec
	}
	return result, nil
}

// phase2 retrieves candidates from FTS, HNSW, and decay pool in parallel.
type candidateSets struct {
	fts    []ScoredID
	vector []ScoredID
	decay  []storage.ULID
}

func (e *ActivationEngine) phase2(ctx context.Context, req *ActivateRequest, p1 *phase1Result, ws [8]byte) (*candidateSets, error) {
	var sets candidateSets

	// Fast path: when HNSW is nil, there is nothing to parallelize.
	// FTS and RecentActive are both in-memory with sub-10µs latency.
	// Eliminating the errgroup saves goroutine spawn + context derivation overhead
	// (~3-5µs per activation at 12+ concurrent goroutines).
	if e.hnsw == nil || len(p1.embedding) == 0 {
		if e.fts != nil {
			results, _ := e.fts.Search(ctx, ws, p1.queryStr, candidatesPerIndex)
			sets.fts = results
		}
		ids, _ := e.store.RecentActive(ctx, ws, candidatesPerIndex)
		sets.decay = ids
		return &sets, nil
	}

	// Full parallel path: FTS + HNSW + decay run concurrently.
	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		if e.fts == nil {
			return nil
		}
		results, err := e.fts.Search(gctx, ws, p1.queryStr, candidatesPerIndex)
		if err != nil {
			return nil
		}
		sets.fts = results
		return nil
	})

	g.Go(func() error {
		results, err := e.hnsw.Search(gctx, ws, p1.embedding, candidatesPerIndex)
		if err != nil {
			return nil
		}
		sets.vector = results
		return nil
	})

	g.Go(func() error {
		ids, err := e.store.RecentActive(gctx, ws, candidatesPerIndex)
		if err != nil {
			return nil
		}
		sets.decay = ids
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}
	return &sets, nil
}

// fusedCandidate is a candidate after RRF fusion.
type fusedCandidate struct {
	id           storage.ULID
	rrfScore     float64
	ftsScore     float64
	vectorScore  float64
	inDecayPool  bool
	hebbianBoost float64
}

const (
	rrfK_HNSW  = 40.0
	rrfK_FTS   = 60.0
	rrfK_Decay = 120.0
)

// phase3RRF merges the three candidate lists via Reciprocal Rank Fusion.
// Uses index-into-slice instead of map-of-pointers to reduce heap allocations.
func phase3RRF(sets *candidateSets) []fusedCandidate {
	totalCap := len(sets.fts) + len(sets.vector) + len(sets.decay)
	result := make([]fusedCandidate, 0, totalCap)
	index := make(map[storage.ULID]int, totalCap)

	getOrCreate := func(id storage.ULID) *fusedCandidate {
		if idx, ok := index[id]; ok {
			return &result[idx]
		}
		idx := len(result)
		result = append(result, fusedCandidate{id: id})
		index[id] = idx
		return &result[idx]
	}

	for rank, s := range sets.fts {
		c := getOrCreate(s.ID)
		c.rrfScore += 1.0 / (rrfK_FTS + float64(rank+1))
		c.ftsScore = s.Score
	}

	for rank, s := range sets.vector {
		c := getOrCreate(s.ID)
		c.rrfScore += 1.0 / (rrfK_HNSW + float64(rank+1))
		c.vectorScore = s.Score
	}

	for rank, id := range sets.decay {
		c := getOrCreate(id)
		c.rrfScore += 1.0 / (rrfK_Decay + float64(rank+1))
		c.inDecayPool = true
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].rrfScore > result[j].rrfScore
	})
	return result
}

// phase4HebbianBoost applies Hebbian association boost to candidates.
// vaultID is used to scope the activation log to the current vault, preventing
// Hebbian boosts from other vaults from bleeding into this vault's results.
func (e *ActivationEngine) phase4HebbianBoost(ctx context.Context, ws [8]byte, vaultID uint32, candidates []fusedCandidate) {
	recent := e.assocLog.RecentForVault(vaultID, 50)
	if len(recent) == 0 {
		return
	}

	now := time.Now().Unix()
	recentWeights := make(map[storage.ULID]float64, len(recent)*10)
	const halfLife = 3600.0
	for _, entry := range recent {
		age := float64(now - entry.At.Unix())
		recencyW := math.Exp(-age / halfLife)
		for _, id := range entry.EngramIDs {
			if w, ok := recentWeights[id]; !ok || recencyW > w {
				recentWeights[id] = recencyW
			}
		}
	}

	if len(recentWeights) == 0 {
		return
	}

	ids := make([]storage.ULID, len(candidates))
	for i, c := range candidates {
		ids[i] = c.id
	}

	assocMap, err := e.store.GetAssociations(ctx, ws, ids, 20)
	if err != nil {
		return
	}

	for i := range candidates {
		assocs := assocMap[candidates[i].id]
		var boost float64
		for _, a := range assocs {
			if rw, ok := recentWeights[a.TargetID]; ok {
				boost += float64(a.Weight) * rw
			}
		}
		if boost > 1.0 {
			boost = 1.0
		}
		candidates[i].hebbianBoost = boost
	}
}

// traversedCandidate is one node discovered via BFS.
type traversedCandidate struct {
	id        storage.ULID
	propagated float64
	hopPath   []storage.ULID
	relType   uint16
}

// resolveProfile implements the C-B-A traversal profile resolution chain:
//  1. Explicit per-request profile override (A) — if valid, use it.
//  2. Auto-inferred from context phrases (C) — if score >= 2, use inferred.
//  3. Vault Plasticity default (B) — if set, use it.
//  4. Hardcoded "default" profile.
//
// Returns both the resolved profile name and the profile pointer. Never returns nil.
func resolveProfile(req *ActivateRequest) (string, *TraversalProfile) {
	name := strings.ToLower(strings.TrimSpace(req.Profile))
	if name != "" && ValidProfileName(name) {
		return name, GetProfile(name)
	}
	inferredName := InferProfile(req.Context, req.VaultDefault)
	return inferredName, GetProfile(inferredName)
}

// phase5Traverse explores the association graph via level-by-level BFS from top candidates.
// Each BFS level issues a single batched GetAssociations call for all nodes at that depth,
// reducing Pebble iterator opens from O(nodes) to O(hops) — typically 2 calls instead of 200+.
func (e *ActivationEngine) phase5Traverse(
	ctx context.Context,
	req *ActivateRequest,
	ws [8]byte,
	profile *TraversalProfile,
	candidates []fusedCandidate,
) []traversedCandidate {
	if len(candidates) == 0 {
		return nil
	}

	const (
		hopPenalty      = 0.7
		minHopScore     = 0.05
		maxBFSNodes     = 500
		maxEdgesPerNode = 10
		maxSeeds        = 20
	)

	seedCount := maxSeeds
	if seedCount > len(candidates) {
		seedCount = len(candidates)
	}

	seen := make(map[storage.ULID]bool, len(candidates)+maxBFSNodes)
	for _, c := range candidates {
		seen[c.id] = true
	}

	type levelItem struct {
		id        storage.ULID
		baseScore float64
		hopDepth  int
		hopPath   []storage.ULID
	}

	// Seed the first level from top candidates.
	currentLevel := make([]levelItem, 0, seedCount)
	for _, seed := range candidates[:seedCount] {
		currentLevel = append(currentLevel, levelItem{
			id:        seed.id,
			baseScore: seed.rrfScore,
			hopDepth:  0,
			hopPath:   []storage.ULID{seed.id},
		})
	}

	var discovered []traversedCandidate
	expanded := 0

	for len(currentLevel) > 0 && expanded < maxBFSNodes {
		// Collect IDs eligible for expansion at this level.
		ids := make([]storage.ULID, 0, len(currentLevel))
		eligible := currentLevel[:0:len(currentLevel)]
		eligible = eligible[:0]
		for _, item := range currentLevel {
			if item.hopDepth < req.HopDepth {
				ids = append(ids, item.id)
				eligible = append(eligible, item)
			}
		}
		if len(ids) == 0 {
			break
		}

		// One batched Pebble call for the entire level.
		assocMap, err := e.store.GetAssociations(ctx, ws, ids, maxEdgesPerNode)
		if err != nil {
			break
		}

		// Fast exit: if no associations exist at this level, deeper levels won't either.
		// Avoids a second BFS round when the corpus has no Hebbian associations yet.
		hasAny := false
		for _, a := range assocMap {
			if len(a) > 0 {
				hasAny = true
				break
			}
		}
		if !hasAny {
			break
		}

		var nextLevel []levelItem
	outer:
		for _, curr := range eligible {
			for _, assoc := range assocMap[curr.id] {
				if seen[assoc.TargetID] {
					continue
				}

				// Profile filtering: skip edges excluded by the traversal profile.
				if !profile.AllowsEdge(assoc.RelType) {
					continue
				}

				boost := float64(profile.BoostFor(assoc.RelType))
				propagated := curr.baseScore * float64(assoc.Weight) * boost * math.Pow(hopPenalty, float64(curr.hopDepth+1))
				if propagated < minHopScore {
					// With per-type boost, weight order alone doesn't guarantee score order.
					// Use continue (not break) so a later low-weight/high-boost edge isn't skipped.
					continue
				}

				seen[assoc.TargetID] = true
				expanded++

				hopPath := make([]storage.ULID, len(curr.hopPath)+1)
				copy(hopPath, curr.hopPath)
				hopPath[len(curr.hopPath)] = assoc.TargetID

				discovered = append(discovered, traversedCandidate{
					id:         assoc.TargetID,
					propagated: propagated,
					hopPath:    hopPath,
					relType:    uint16(assoc.RelType),
				})

				if curr.hopDepth+1 < req.HopDepth {
					nextLevel = append(nextLevel, levelItem{
						id:        assoc.TargetID,
						baseScore: propagated,
						hopDepth:  curr.hopDepth + 1,
						hopPath:   hopPath,
					})
				}

				if expanded >= maxBFSNodes {
					break outer
				}
			}
		}
		currentLevel = nextLevel
	}
	return discovered
}

// phase6Score computes final scores, applies filters, and builds the result.
func (e *ActivationEngine) phase6Score(
	ctx context.Context,
	req *ActivateRequest,
	ws [8]byte,
	fused []fusedCandidate,
	traversed []traversedCandidate,
	p1 *phase1Result,
) (*ActivateResult, error) {

	w := resolveWeights(req.Weights, e.weights)

	type scoringCandidate struct {
		id           storage.ULID
		ftsScore     float64
		vectorScore  float64
		hebbianBoost float64
		rrfScore     float64
		hopPath      []storage.ULID
		relType      uint16
	}

	// Deduplicate: fused candidates take priority; traversed candidates are
	// only added if their ID has not already appeared in the fused set.
	seen := make(map[storage.ULID]struct{}, len(fused)+len(traversed))
	all := make([]scoringCandidate, 0, len(fused)+len(traversed))
	for _, c := range fused {
		if _, dup := seen[c.id]; dup {
			continue
		}
		seen[c.id] = struct{}{}
		all = append(all, scoringCandidate{
			id:           c.id,
			ftsScore:     c.ftsScore,
			vectorScore:  c.vectorScore,
			hebbianBoost: c.hebbianBoost,
			rrfScore:     c.rrfScore,
		})
	}
	for _, t := range traversed {
		if _, dup := seen[t.id]; dup {
			continue
		}
		seen[t.id] = struct{}{}
		all = append(all, scoringCandidate{
			id:       t.id,
			rrfScore: t.propagated,
			hopPath:  t.hopPath,
			relType:  t.relType,
		})
	}

	ids := make([]storage.ULID, len(all))
	for i, c := range all {
		ids[i] = c.id
	}

	// Load full engrams for all candidates in one pass.
	// Previously this was two passes: GetMetadata (all candidates) + GetEngrams (scored subset).
	// Loading full engrams upfront eliminates the second pass entirely — engrams are already
	// in hand when building the activation result. The extra bytes per candidate (~2-8KB vs ~46B
	// for metadata-only) are worth eliminating an entire Pebble read round-trip.
	allEngrams, err := e.store.GetEngrams(ctx, ws, ids)
	if err != nil {
		return nil, fmt.Errorf("phase6 get engrams: %w", err)
	}
	engramByID := make(map[storage.ULID]*storage.Engram, len(allEngrams))
	for _, eng := range allEngrams {
		if eng != nil {
			engramByID[eng.ID] = eng
		}
	}

	// Look up per-engram cache access time to use for recency scoring.
	// Cache hits return the time they were last recalled (≈ now); misses return 0
	// so computeComponents falls back to eng.LastAccess (the persisted write time).
	lastAccessNsByID := make(map[storage.ULID]int64, len(all))
	for _, c := range all {
		if ns := e.store.EngramLastAccessNs(ws, c.id); ns != 0 {
			lastAccessNsByID[c.id] = ns
		}
	}

	type scoredItem struct {
		id         storage.ULID
		final      float64
		components ScoreComponents
		hopPath    []storage.ULID
	}

	now := time.Now()
	scored := make([]scoredItem, 0, len(all))

	for _, c := range all {
		eng := engramByID[c.id]
		if eng == nil {
			continue
		}

		if !passesMetaFilter(eng, req.Filters) {
			continue
		}

		components := computeComponents(c.vectorScore, c.ftsScore, c.hebbianBoost, eng, lastAccessNsByID[c.id], now, w)
		final := components.Raw * components.Confidence

		if final < req.Threshold {
			continue
		}

		scored = append(scored, scoredItem{
			id:         c.id,
			final:      final,
			components: components,
			hopPath:    c.hopPath,
		})
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].final > scored[j].final
	})

	totalFound := len(scored)
	if len(scored) > req.MaxResults {
		scored = scored[:req.MaxResults]
	}

	// Apply structured filter if provided (post-retrieval predicate).
	// This is applied AFTER RRF scoring and confidence checks, as the final step.
	if req.StructuredFilter != nil {
		if qf, ok := req.StructuredFilter.(interface {
			Match(*storage.Engram) bool
		}); ok {
			filtered := make([]scoredItem, 0, len(scored))
			for _, s := range scored {
				eng := engramByID[s.id]
				if eng == nil {
					continue
				}
				if qf.Match(eng) {
					filtered = append(filtered, s)
				}
			}
			scored = filtered
		}
	}

	activations := make([]ScoredEngram, 0, len(scored))
	for _, s := range scored {
		eng := engramByID[s.id]
		if eng == nil {
			continue
		}
		var why string
		if req.IncludeWhy {
			why = buildWhy(eng, s.components, s.hopPath, p1.queryStr)
		}
		activations = append(activations, ScoredEngram{
			Engram:     eng,
			Score:      s.final,
			Components: s.components,
			Why:        why,
			HopPath:    append([]storage.ULID(nil), s.hopPath...),
			Dormant:    eng.Relevance <= minFloor*1.1,
		})
	}

	return &ActivateResult{
		QueryID:     newQueryID(),
		Activations: activations,
		TotalFound:  totalFound,
	}, nil
}

// computeComponents calculates all scoring components for a candidate engram.
// Accepts *storage.Engram directly — avoids a separate GetMetadata call in phase6.
// lastAccessNs is the nanosecond timestamp of last cache access (0 if not cached).
func computeComponents(vectorScore, ftsScore, hebbianBoost float64, eng *storage.Engram, lastAccessNs int64, now time.Time, w resolvedWeights) ScoreComponents {
	const accessFreqSaturation = 100.0
	const recencyHalfLifeDays = 7.0

	accessFreq := math.Log1p(float64(eng.AccessCount)) / math.Log1p(accessFreqSaturation)
	if accessFreq > 1.0 {
		accessFreq = 1.0
	}

	// Use cache lastAccess if available (reflects actual recall time); else use persisted eng.LastAccess.
	var lastAccess time.Time
	if lastAccessNs > 0 {
		lastAccess = time.Unix(0, lastAccessNs)
	} else {
		lastAccess = eng.LastAccess
	}
	daysSince := now.Sub(lastAccess).Hours() / 24.0
	recency := math.Exp(-daysSince * math.Log(2) / recencyHalfLifeDays)

	decayFactor := math.Max(0.05, math.Exp(-daysSince/float64(eng.Stability)))

	// Normalize BM25 score from [0, +∞) to [0, 1) using tanh.
	// Raw BM25 is unbounded and not comparable to cosine similarity [0,1].
	// tanh(0)=0, tanh(1)≈0.76, tanh(3)≈0.995 — preserves relative ordering,
	// prevents high BM25 scores from saturating the composite score via clamping.
	normalizedFTS := math.Tanh(ftsScore)

	raw := w.SemanticSimilarity*vectorScore +
		w.FullTextRelevance*normalizedFTS +
		w.DecayFactor*decayFactor +
		w.HebbianBoost*hebbianBoost +
		w.AccessFrequency*accessFreq +
		w.Recency*recency

	if raw > 1.0 {
		raw = 1.0
	}
	if raw < 0.0 {
		raw = 0.0
	}

	conf := float64(eng.Confidence)

	return ScoreComponents{
		SemanticSimilarity: vectorScore,
		FullTextRelevance:  normalizedFTS, // normalized [0,1), not raw BM25
		DecayFactor:        decayFactor,
		HebbianBoost:       hebbianBoost,
		AccessFrequency:    accessFreq,
		Recency:            recency,
		Confidence:         conf,
		Raw:                raw,
		Final:              raw * conf,
	}
}

// passesMetaFilter evaluates filter predicates against a full engram.
// Accepts *storage.Engram directly — avoids a separate GetMetadata call in phase6.
func passesMetaFilter(eng *storage.Engram, filters []Filter) bool {
	for _, f := range filters {
		switch f.Field {
		case "state":
			if s, ok := f.Value.(storage.LifecycleState); ok {
				switch f.Op {
				case "eq":
					if eng.State != s {
						return false
					}
				case "neq":
					if eng.State == s {
						return false
					}
				}
			}
		case "created_after":
			if t, ok := f.Value.(time.Time); ok {
				if !eng.CreatedAt.After(t) {
					return false
				}
			}
		case "created_before":
			if t, ok := f.Value.(time.Time); ok {
				if !eng.CreatedAt.Before(t) {
					return false
				}
			}
		}
	}
	return true
}

func resolveWeights(req *Weights, def DefaultWeights) resolvedWeights {
	if req == nil {
		return resolvedWeights{
			SemanticSimilarity: float64(def.SemanticSimilarity),
			FullTextRelevance:  float64(def.FullTextRelevance),
			DecayFactor:        float64(def.DecayFactor),
			HebbianBoost:       float64(def.HebbianBoost),
			AccessFrequency:    float64(def.AccessFrequency),
			Recency:            float64(def.Recency),
		}
	}
	return resolvedWeights{
		SemanticSimilarity: float64(req.SemanticSimilarity),
		FullTextRelevance:  float64(req.FullTextRelevance),
		DecayFactor:        float64(req.DecayFactor),
		HebbianBoost:       float64(req.HebbianBoost),
		AccessFrequency:    float64(req.AccessFrequency),
		Recency:            float64(req.Recency),
	}
}

func buildWhy(eng *storage.Engram, c ScoreComponents, hopPath []storage.ULID, queryStr string) string {
	var parts []string

	signals := map[string]float64{
		"semantic": c.SemanticSimilarity,
		"fts":      c.FullTextRelevance,
		"decay":    c.DecayFactor,
		"hebbian":  c.HebbianBoost,
	}
	best := ""
	bestVal := 0.0
	for k, v := range signals {
		if v > bestVal {
			bestVal = v
			best = k
		}
	}

	switch best {
	case "semantic":
		parts = append(parts, fmt.Sprintf("high semantic similarity (%.0f%%) to context", c.SemanticSimilarity*100))
	case "fts":
		q := queryStr
		if len(q) > 40 {
			q = q[:40] + "..."
		}
		parts = append(parts, fmt.Sprintf("strong full-text match (%.0f%%) to \"%s\"", c.FullTextRelevance*100, q))
	case "decay":
		parts = append(parts, "frequently accessed recently, high decay relevance")
	case "hebbian":
		parts = append(parts, "strongly associated with recently activated engrams")
	}

	if len(hopPath) > 1 {
		parts = append(parts, fmt.Sprintf("reached via %d association hop(s)", len(hopPath)-1))
	}

	if c.Confidence < 0.5 {
		parts = append(parts, fmt.Sprintf("confidence is low (%.0f%%)", c.Confidence*100))
	}

	if eng.Relevance <= minFloor*1.1 {
		parts = append(parts, "dormant (low decay relevance)")
	}

	return strings.Join(parts, "; ")
}

// queryIDSeq is a process-wide monotonic counter for query IDs.
// Replaces crypto/rand — the result is used for tracing only, not security.
var queryIDSeq atomic.Uint64

func newQueryID() string {
	return fmt.Sprintf("q-%016x", queryIDSeq.Add(1))
}

// Stream sends result frames to the provided send function.
func (e *ActivationEngine) Stream(
	ctx context.Context,
	result *ActivateResult,
	send func(frame *ActivateResponseFrame) error,
) error {
	activations := result.Activations
	totalFrames := (len(activations) + frameSize - 1) / frameSize
	if totalFrames == 0 {
		totalFrames = 1
	}

	for frame := 0; frame < totalFrames; frame++ {
		lo := frame * frameSize
		hi := lo + frameSize
		if hi > len(activations) {
			hi = len(activations)
		}

		f := &ActivateResponseFrame{
			QueryID:     result.QueryID,
			TotalFound:  result.TotalFound,
			LatencyMs:   result.LatencyMs,
			Activations: activations[lo:hi],
			Frame:       frame + 1,
			TotalFrames: totalFrames,
		}

		if err := send(f); err != nil {
			return fmt.Errorf("stream frame %d: %w", frame, err)
		}
	}
	return nil
}
