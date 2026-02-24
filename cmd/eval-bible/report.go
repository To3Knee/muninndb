package main

import (
	"fmt"
	"io"
	"os"
	"time"
)

// writeReport writes a human-readable evaluation report to w.
func writeReport(w io.Writer, p1 Phase1Result, p2 Phase2Result, mode string, corpusSize int, loadDur time.Duration) {
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "════════════════════════════════════════════════════════════════\n")
	fmt.Fprintf(w, "MUNINNDB BIBLE EVAL REPORT\n")
	fmt.Fprintf(w, "════════════════════════════════════════════════════════════════\n")
	fmt.Fprintf(w, "Mode:          %s\n", mode)
	fmt.Fprintf(w, "Corpus size:   %d verses\n", corpusSize)
	fmt.Fprintf(w, "Load time:     %v\n", loadDur.Round(time.Millisecond))
	fmt.Fprintf(w, "\n")

	fmt.Fprintf(w, "── Phase 1: Retrieval Quality ──────────────────────────────────\n")
	fmt.Fprintf(w, "Seeds evaluated:   %d\n", p1.SeedsEvaluated)
	fmt.Fprintf(w, "Avg cross-refs:    %.1f\n", p1.AvgCrossRefs)
	fmt.Fprintf(w, "Recall@10:         %.4f\n", p1.RecallAtK)
	fmt.Fprintf(w, "NDCG@10:           %.4f\n", p1.NDCGAtK)
	fmt.Fprintf(w, "\n")

	fmt.Fprintf(w, "── Phase 2: Cognitive Properties ──────────────────────────────\n")

	fmt.Fprintf(w, "Sub-experiment A: Decay Bias\n")
	fmt.Fprintf(w, "  Fresh (NT) MRR@10:  %.4f\n", p2.FreshMRR)
	fmt.Fprintf(w, "  Stale (OT) MRR@10:  %.4f\n", p2.StaleMRR)
	fmt.Fprintf(w, "  Decay score:        %.4f  (positive = decay working)\n", p2.DecayScore)
	if len(p2.AnchorResults) > 0 {
		fmt.Fprintf(w, "\n  %-20s  %8s  %8s\n", "Query", "FreshMRR", "StaleMRR")
		fmt.Fprintf(w, "  %-20s  %8s  %8s\n", "--------------------", "--------", "--------")
		for _, a := range p2.AnchorResults {
			fmt.Fprintf(w, "  %-20s  %8.4f  %8.4f\n", a.Label, a.FreshMRR, a.StaleMRR)
		}
	}
	fmt.Fprintf(w, "\n")

	fmt.Fprintf(w, "Sub-experiment B: Hebbian Lift\n")
	fmt.Fprintf(w, "  Avg rank delta:     %.2f  (positive = linked OT ranks higher)\n", p2.HebbianScore)
	if len(p2.PairResults) > 0 {
		fmt.Fprintf(w, "\n  %-18s  %-18s  %6s  %7s  %5s\n",
			"OT (linked)", "Control", "Linked", "Control", "Delta")
		fmt.Fprintf(w, "  %-18s  %-18s  %6s  %7s  %5s\n",
			"------------------", "------------------", "------", "-------", "-----")
		for _, pr := range p2.PairResults {
			fmt.Fprintf(w, "  %-18s  %-18s  %6d  %7d  %+5d\n",
				pr.OTVerse, pr.Control, pr.LinkedRank, pr.ControlRank, pr.Delta)
		}
	}
	fmt.Fprintf(w, "\n")

	decayWorking := p2.DecayScore > 0
	hebbianWorking := p2.HebbianScore > 0
	fmt.Fprintf(w, "── Verdict ─────────────────────────────────────────────────────\n")
	fmt.Fprintf(w, "%s\n", verdictLine(p1.NDCGAtK, decayWorking, hebbianWorking))
	fmt.Fprintf(w, "════════════════════════════════════════════════════════════════\n")
}

// verdictLine returns a one-line summary verdict.
func verdictLine(ndcg float64, decayWorking, hebbianWorking bool) string {
	quality := "POOR"
	switch {
	case ndcg >= 0.5:
		quality = "HIGH"
	case ndcg >= 0.3:
		quality = "GOOD"
	case ndcg >= 0.1:
		quality = "ACCEPTABLE"
	}

	cogParts := ""
	switch {
	case decayWorking && hebbianWorking:
		cogParts = "decay + Hebbian confirmed"
	case decayWorking:
		cogParts = "decay confirmed, no Hebbian signal"
	case hebbianWorking:
		cogParts = "Hebbian confirmed, no decay signal"
	default:
		cogParts = "no cognitive signal"
	}

	return fmt.Sprintf("Retrieval quality: %s (NDCG@10=%.4f) | Cognitive: %s", quality, ndcg, cogParts)
}

// saveReport writes the report to a file, appending if it already exists.
func saveReport(path string, p1 Phase1Result, p2 Phase2Result, mode string, corpusSize int, loadDur time.Duration) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("open results file: %w", err)
	}
	defer f.Close()
	fmt.Fprintf(f, "\n# Run: %s\n", time.Now().Format(time.RFC3339))
	writeReport(f, p1, p2, mode, corpusSize, loadDur)
	return nil
}
