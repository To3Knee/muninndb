package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/scrypster/muninndb/internal/transport/mbp"
)

// thematicQuery is a query used to measure cognitive/thematic coherence.
type thematicQuery struct {
	Context string
	Label   string
}

// thematicQueries is the fixed set of 10 thematic eval queries.
var thematicQueries = []thematicQuery{
	{Context: "forgiveness redemption grace mercy sins forgiven", Label: "forgiveness"},
	{Context: "eternal life salvation believe faith Jesus saved", Label: "eternal life"},
	{Context: "shepherd flock lost sheep pasture", Label: "shepherd"},
	{Context: "resurrection from the dead raised third day", Label: "resurrection"},
	{Context: "love one another neighbor commandment", Label: "love commandment"},
	{Context: "bread of life feeding five thousand loaves fish", Label: "feeding miracle"},
	{Context: "Holy Spirit Pentecost tongues fire descended", Label: "Holy Spirit"},
	{Context: "faith without works dead Abraham justified", Label: "faith and works"},
	{Context: "creation light darkness void earth heaven", Label: "creation"},
	{Context: "wisdom understanding proverbs fear of the Lord", Label: "wisdom"},
}

// queryDelta captures the NDCG change for one thematic query across phases.
type queryDelta struct {
	Label        string
	BaselineNDCG float64
	PostReadNDCG float64
	Delta        float64
}

// Phase2Result captures the cognitive properties evaluation results.
type Phase2Result struct {
	BaselineNDCG    float64
	PostReadingNDCG float64
	PostDecayNDCG   float64
	QueryDeltas     []queryDelta
}

// RunPhase2 evaluates cognitive properties of the engine:
//  1. Measure baseline NDCG for 10 thematic queries.
//  2. Simulate reading: activate 200 John verses to build Hebbian associations.
//  3. Wait 3 seconds for Hebbian worker to process.
//  4. Re-measure NDCG on same thematic queries.
//  5. Return results (PostDecayNDCG == PostReadingNDCG for v1).
func RunPhase2(ctx context.Context, ee *evalEngine, gospelJohnTexts []string) Phase2Result {
	fmt.Println("Phase 2: measuring baseline thematic NDCG...")
	baselineNDCGs := measureThematicNDCGs(ctx, ee)
	baselineAvg := avg64(baselineNDCGs)
	fmt.Printf("  baseline avg NDCG: %.4f\n", baselineAvg)

	// Simulate reading: activate up to 200 John verses
	limit := 200
	if len(gospelJohnTexts) < limit {
		limit = len(gospelJohnTexts)
	}
	fmt.Printf("Phase 2: activating %d John verses to build Hebbian associations...\n", limit)
	for i, text := range gospelJohnTexts[:limit] {
		_, _ = ee.activate(ctx, []string{text})
		if (i+1)%50 == 0 {
			fmt.Printf("  activated %d/%d John verses...\n", i+1, limit)
		}
	}

	// Wait for Hebbian worker to process accumulated activations
	fmt.Println("Phase 2: waiting 3s for Hebbian worker...")
	time.Sleep(3 * time.Second)

	fmt.Println("Phase 2: re-measuring thematic NDCG post-reading...")
	postNDCGs := measureThematicNDCGs(ctx, ee)
	postAvg := avg64(postNDCGs)
	fmt.Printf("  post-reading avg NDCG: %.4f\n", postAvg)

	deltas := make([]queryDelta, len(thematicQueries))
	for i, q := range thematicQueries {
		baseline := 0.0
		if i < len(baselineNDCGs) {
			baseline = baselineNDCGs[i]
		}
		post := 0.0
		if i < len(postNDCGs) {
			post = postNDCGs[i]
		}
		deltas[i] = queryDelta{
			Label:        q.Label,
			BaselineNDCG: baseline,
			PostReadNDCG: post,
			Delta:        post - baseline,
		}
	}

	return Phase2Result{
		BaselineNDCG:    baselineAvg,
		PostReadingNDCG: postAvg,
		// Skip genealogy decay for v1 — set PostDecayNDCG = PostReadingNDCG
		PostDecayNDCG: postAvg,
		QueryDeltas:   deltas,
	}
}

// measureThematicNDCGs runs all 10 thematic queries and returns per-query NDCG scores
// using keyword proximity as a proxy for relevance (no ground truth available).
func measureThematicNDCGs(ctx context.Context, ee *evalEngine) []float64 {
	scores := make([]float64, len(thematicQueries))
	for i, q := range thematicQueries {
		activations, err := ee.activate(ctx, []string{q.Context})
		if err != nil || len(activations) == 0 {
			scores[i] = 0
			continue
		}

		// Proxy relevance: result is relevant if concept+content contains any query keyword > 3 chars
		keywords := keywordsFromContext(q.Context)
		relevant := make(map[string]bool, len(activations))
		resultRefs := make([]string, len(activations))
		for j, act := range activations {
			resultRefs[j] = act.Concept
			text := strings.ToLower(act.Concept + " " + act.Content)
			for _, kw := range keywords {
				if strings.Contains(text, kw) {
					relevant[act.Concept] = true
					break
				}
			}
		}
		scores[i] = ndcgAtK(resultRefs, relevant, 10)
	}
	return scores
}

// keywordsFromContext extracts words longer than 3 characters from a context string.
func keywordsFromContext(context string) []string {
	words := strings.Fields(strings.ToLower(context))
	out := make([]string, 0, len(words))
	for _, w := range words {
		if len(w) > 3 {
			out = append(out, w)
		}
	}
	return out
}

// filterJohnVerses returns the content texts of all "John X:Y" verses.
func filterJohnVerses(reqs []mbp.WriteRequest) []string {
	out := make([]string, 0, 879) // Gospel of John has ~879 verses
	for _, r := range reqs {
		if strings.HasPrefix(r.Concept, "John ") {
			out = append(out, r.Content)
		}
	}
	return out
}

// avg64 returns the arithmetic mean of a float64 slice.
func avg64(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	var sum float64
	for _, v := range vals {
		sum += v
	}
	return sum / float64(len(vals))
}
