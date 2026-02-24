package main

import "math"

// recallAtK returns the fraction of relevant items found in the top-k results.
// recall = |relevant ∩ top-k| / |relevant|
func recallAtK(results []string, relevant map[string]bool, k int) float64 {
	if len(relevant) == 0 {
		return 0
	}
	limit := k
	if limit > len(results) {
		limit = len(results)
	}
	found := 0
	for i := 0; i < limit; i++ {
		if relevant[results[i]] {
			found++
		}
	}
	return float64(found) / float64(len(relevant))
}

// ndcgAtK computes Normalized Discounted Cumulative Gain at rank k.
// NDCG = DCG / IDCG, where IDCG is the DCG of the ideal (perfect) ranking.
func ndcgAtK(results []string, relevant map[string]bool, k int) float64 {
	if len(relevant) == 0 {
		return 0
	}
	dcg := computeDCG(results, relevant, k)
	if dcg == 0 {
		return 0
	}

	// Build ideal ranking: all relevant items first
	ideal := make([]string, 0, len(relevant))
	for id := range relevant {
		ideal = append(ideal, id)
	}
	idcg := computeDCG(ideal, relevant, k)
	if idcg == 0 {
		return 0
	}
	return dcg / idcg
}

// computeDCG computes Discounted Cumulative Gain for the top-k results.
// DCG = sum over rank i (1-indexed) of: 1 / log2(i + 1) if result[i] is relevant.
func computeDCG(results []string, relevant map[string]bool, k int) float64 {
	limit := k
	if limit > len(results) {
		limit = len(results)
	}
	var dcg float64
	for i := 0; i < limit; i++ {
		if relevant[results[i]] {
			// rank is i+1 (1-indexed), gain = 1/log2(rank+1)
			dcg += 1.0 / math.Log2(float64(i+2))
		}
	}
	return dcg
}
