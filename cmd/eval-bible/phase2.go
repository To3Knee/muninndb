package main

import (
	"context"
	"fmt"
	"time"

	"github.com/scrypster/muninndb/internal/transport/mbp"
)

// anchorResult captures per-query MRR scores for Sub-experiment A.
type anchorResult struct {
	Label    string
	FreshMRR float64
	StaleMRR float64
}

// pairResult captures per-pair rank data for Sub-experiment B.
type pairResult struct {
	OTVerse     string
	NTVerse     string
	Control     string
	LinkedRank  int // rank of OT verse queried via NT verse name (1-indexed; k+1 = not found)
	ControlRank int // rank of unlinked control OT verse under same query
	Delta       int // ControlRank - LinkedRank; positive = Hebbian is lifting the linked verse
}

// Phase2Result captures the cognitive properties evaluation results.
type Phase2Result struct {
	// Sub-experiment A: decay bias (fresh NT > stale OT?)
	FreshMRR   float64
	StaleMRR   float64
	DecayScore float64 // (FreshMRR - StaleMRR) / FreshMRR; positive = decay working

	// Sub-experiment B: Hebbian lift (linked stale ranks higher than unlinked control?)
	HebbianScore float64 // avg(ControlRank - LinkedRank); positive = Hebbian lifting stale verse

	AnchorResults []anchorResult
	PairResults   []pairResult
}

// hebbianControls maps each linked OT verse to an unlinked OT control with the same DaysAgo.
// Controls verify that Hebbian links — not just thematic content — drive the rank lift.
var hebbianControls = map[string]string{
	"Genesis 1:1":  "Exodus 20:3",     // both DaysAgo=3650
	"Isaiah 53:5":  "Jeremiah 29:11",  // both DaysAgo=2190
	"Psalm 23:1":   "Psalm 23:4",      // both DaysAgo=90
	"Genesis 3:15": "Deuteronomy 6:4", // both DaysAgo=3650
	"Isaiah 7:14":  "Jeremiah 29:11",  // both DaysAgo=2190
}

// RunPhase2 evaluates cognitive properties using the curated fixture corpus.
//
// Process:
//  1. Write all curated fixtures via the engine and install hand-set cognitive state.
//  2. Co-activate each Hebbian pair to seed cross-testament associations.
//  3. Wait for the Hebbian worker to process accumulated activations.
//  4. Sub-experiment A: measure FreshMRR vs StaleMRR across 6 anchor queries.
//  5. Sub-experiment B: compare linked-stale rank vs unlinked-control rank for 5 pairs.
func RunPhase2(ctx context.Context, ee *evalEngine) Phase2Result {
	const k = 10

	fmt.Println("Phase 2: writing curated fixture corpus...")
	if err := loadFixtures(ctx, ee); err != nil {
		fmt.Printf("  WARNING: fixture load error: %v\n", err)
	}

	fmt.Println("Phase 2: co-activating Hebbian pairs...")
	for _, pair := range hebbianPairs {
		_, _ = ee.activate(ctx, []string{pair[0], pair[1]})
	}

	fmt.Println("Phase 2: waiting 3s for Hebbian worker to process...")
	time.Sleep(3 * time.Second)

	// Sub-experiment A: anchor queries — does FreshMRR > StaleMRR (decay is working)?
	fmt.Println("Phase 2: Sub-experiment A — decay bias (FreshMRR vs StaleMRR)...")
	anchorResults := make([]anchorResult, len(anchorQueries))
	for i, q := range anchorQueries {
		acts, err := ee.activate(ctx, []string{q.Context})
		if err != nil || len(acts) == 0 {
			anchorResults[i] = anchorResult{Label: q.Label}
			continue
		}
		concepts := extractConcepts(acts)
		anchorResults[i] = anchorResult{
			Label:    q.Label,
			FreshMRR: mrrAtK(concepts, q.FreshRefs, k),
			StaleMRR: mrrAtK(concepts, q.StaleRefs, k),
		}
	}

	avgFreshMRR := avgFieldMRR(anchorResults, true)
	avgStaleMRR := avgFieldMRR(anchorResults, false)
	decayScore := 0.0
	if avgFreshMRR > 0 {
		decayScore = (avgFreshMRR - avgStaleMRR) / avgFreshMRR
	}
	fmt.Printf("  avg FreshMRR=%.4f  StaleMRR=%.4f  decay_score=%.4f\n",
		avgFreshMRR, avgStaleMRR, decayScore)

	// Sub-experiment B: Hebbian lift — does linked stale OT verse rank higher than control?
	fmt.Println("Phase 2: Sub-experiment B — Hebbian lift (linked vs control rank)...")
	pairResults := make([]pairResult, len(hebbianPairs))
	var totalDelta int
	for i, pair := range hebbianPairs {
		otVerse := pair[0]
		ntVerse := pair[1]
		control := hebbianControls[otVerse]

		acts, err := ee.activate(ctx, []string{ntVerse})
		if err != nil {
			pairResults[i] = pairResult{OTVerse: otVerse, NTVerse: ntVerse, Control: control,
				LinkedRank: k + 1, ControlRank: k + 1}
			continue
		}
		concepts := extractConcepts(acts)
		linkedRank := rankOf(concepts, otVerse, k)
		controlRank := rankOf(concepts, control, k)
		delta := controlRank - linkedRank

		pairResults[i] = pairResult{
			OTVerse:     otVerse,
			NTVerse:     ntVerse,
			Control:     control,
			LinkedRank:  linkedRank,
			ControlRank: controlRank,
			Delta:       delta,
		}
		totalDelta += delta
		fmt.Printf("  %-20s linked=%2d  control=%-20s rank=%2d  delta=%+d\n",
			otVerse, linkedRank, control, controlRank, delta)
	}
	hebbianScore := 0.0
	if len(hebbianPairs) > 0 {
		hebbianScore = float64(totalDelta) / float64(len(hebbianPairs))
	}
	fmt.Printf("  avg Hebbian delta=%.2f (positive = linked ranks higher than control)\n", hebbianScore)

	return Phase2Result{
		FreshMRR:      avgFreshMRR,
		StaleMRR:      avgStaleMRR,
		DecayScore:    decayScore,
		HebbianScore:  hebbianScore,
		AnchorResults: anchorResults,
		PairResults:   pairResults,
	}
}

// loadFixtures writes all curated verses into the vault and sets their cognitive state.
func loadFixtures(ctx context.Context, ee *evalEngine) error {
	for _, group := range phase2Fixtures {
		for _, fix := range group.Fixtures {
			req := mbp.WriteRequest{
				Concept: fix.Concept,
				Content: fix.Content,
				Tags:    fix.Tags,
				Vault:   "bible",
			}
			if _, err := ee.writeVerse(ctx, req); err != nil {
				return fmt.Errorf("write %q: %w", fix.Concept, err)
			}
			lastAccess := daysAgoToTime(fix.DaysAgo)
			if err := ee.setEngramState(ctx, fix.Concept, lastAccess, fix.AccessCount, fix.Stability); err != nil {
				// Log but don't abort — Phase 1 data may have already loaded these
				fmt.Printf("  setEngramState %q: %v\n", fix.Concept, err)
			}
		}
	}
	return nil
}

// extractConcepts returns the Concept field from each activation item in order.
func extractConcepts(acts []mbp.ActivationItem) []string {
	out := make([]string, len(acts))
	for i, a := range acts {
		out[i] = a.Concept
	}
	return out
}

// rankOf returns the 1-indexed rank of concept in results, or k+1 if not found.
func rankOf(results []string, concept string, k int) int {
	limit := k
	if limit > len(results) {
		limit = len(results)
	}
	for i := 0; i < limit; i++ {
		if results[i] == concept {
			return i + 1
		}
	}
	return k + 1
}

// avgFieldMRR computes the average FreshMRR (fresh=true) or StaleMRR (fresh=false)
// across all anchor results.
func avgFieldMRR(results []anchorResult, fresh bool) float64 {
	if len(results) == 0 {
		return 0
	}
	var sum float64
	for _, r := range results {
		if fresh {
			sum += r.FreshMRR
		} else {
			sum += r.StaleMRR
		}
	}
	return sum / float64(len(results))
}
