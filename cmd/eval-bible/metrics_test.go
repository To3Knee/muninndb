package main

import (
	"math"
	"testing"
)

func TestRecallAtK_Perfect(t *testing.T) {
	results := []string{"a", "b", "c"}
	relevant := map[string]bool{"a": true, "b": true, "c": true}
	got := recallAtK(results, relevant, 3)
	if got != 1.0 {
		t.Errorf("want 1.0, got %f", got)
	}
}

func TestRecallAtK_Partial(t *testing.T) {
	results := []string{"a", "x", "b", "y"}
	relevant := map[string]bool{"a": true, "b": true, "c": true, "d": true}
	// top-4: a, x, b, y — 2 relevant out of 4 relevant total = 0.5
	got := recallAtK(results, relevant, 4)
	if math.Abs(got-0.5) > 1e-9 {
		t.Errorf("want 0.5, got %f", got)
	}
}

func TestRecallAtK_None(t *testing.T) {
	results := []string{"x", "y"}
	relevant := map[string]bool{"a": true}
	got := recallAtK(results, relevant, 2)
	if got != 0.0 {
		t.Errorf("want 0.0, got %f", got)
	}
}

func TestNDCGAtK_Perfect(t *testing.T) {
	results := []string{"a", "b", "c"}
	relevant := map[string]bool{"a": true, "b": true, "c": true}
	got := ndcgAtK(results, relevant, 3)
	if math.Abs(got-1.0) > 1e-9 {
		t.Errorf("want 1.0, got %f", got)
	}
}

func TestNDCGAtK_Degraded(t *testing.T) {
	// Relevant items at ranks 2 and 3 instead of 1 and 2 — NDCG < 1.0
	results := []string{"x", "a", "b"}
	relevant := map[string]bool{"a": true, "b": true}
	got := ndcgAtK(results, relevant, 3)
	if got >= 1.0 {
		t.Errorf("want < 1.0 for degraded ranking, got %f", got)
	}
	if got <= 0.0 {
		t.Errorf("want > 0.0 for partially relevant ranking, got %f", got)
	}
}

func TestNDCGAtK_NoRelevant(t *testing.T) {
	results := []string{"x", "y", "z"}
	relevant := map[string]bool{"a": true}
	got := ndcgAtK(results, relevant, 3)
	if got != 0.0 {
		t.Errorf("want 0.0, got %f", got)
	}
}

func TestRecallAtK_KCap(t *testing.T) {
	results := []string{"a", "b", "c", "d"}
	relevant := map[string]bool{"a": true, "b": true, "c": true, "d": true}
	// Only top-2 checked
	got := recallAtK(results, relevant, 2)
	if math.Abs(got-0.5) > 1e-9 {
		t.Errorf("want 0.5 (2 found of 4 relevant), got %f", got)
	}
}
