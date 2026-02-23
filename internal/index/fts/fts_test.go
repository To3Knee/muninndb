package fts

import (
	"context"
	"os"
	"testing"

	"github.com/cockroachdb/pebble"
	"github.com/scrypster/muninndb/internal/storage"
)

func openTestDB(t *testing.T) (*pebble.DB, func()) {
	t.Helper()
	dir, err := os.MkdirTemp("", "muninndb-fts-*")
	if err != nil {
		t.Fatal(err)
	}
	db, err := storage.OpenPebble(dir, storage.DefaultOptions())
	if err != nil {
		os.RemoveAll(dir)
		t.Fatal(err)
	}
	return db, func() {
		db.Close()
		os.RemoveAll(dir)
	}
}

// TestIndexEngramUpdatesStats verifies that IndexEngram updates per-term document
// frequency and global stats so that BM25 search returns non-zero scores.
//
// Regression: IndexEngram wrote posting keys but did not write TermStats (df)
// or call UpdateStats. This caused getIDF to return 0 for all terms, making
// every BM25 score 0 and Search return no results.
func TestIndexEngramUpdatesStats(t *testing.T) {
	db, cleanup := openTestDB(t)
	defer cleanup()

	idx := New(db)
	store := storage.NewPebbleStore(db, 100)
	ws := store.VaultPrefix("test")
	ctx := context.Background()

	id := [16]byte{1}
	err := idx.IndexEngram(ws, id, "Go programming language", "", "Go is a compiled language", []string{"golang"})
	if err != nil {
		t.Fatalf("IndexEngram: %v", err)
	}

	// Stats must be updated
	stats := idx.readStats(ws)
	if stats.TotalEngrams == 0 {
		t.Errorf("TotalEngrams = 0, want >= 1 after IndexEngram")
	}
	if stats.AvgDocLen == 0 {
		t.Errorf("AvgDocLen = 0, want > 0 after IndexEngram")
	}

	// Search must return results with non-zero scores
	results, err := idx.Search(ctx, ws, "compiled language", 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("Search returned 0 results, want >= 1")
	}
	if results[0].Score <= 0 {
		t.Errorf("results[0].Score = %v, want > 0", results[0].Score)
	}
}

// TestFTSRankingOrder verifies that the most relevant document ranks first.
func TestFTSRankingOrder(t *testing.T) {
	db, cleanup := openTestDB(t)
	defer cleanup()

	idx := New(db)
	store := storage.NewPebbleStore(db, 100)
	ws := store.VaultPrefix("test")
	ctx := context.Background()

	id1 := [16]byte{1}
	id2 := [16]byte{2}
	id3 := [16]byte{3}

	_ = idx.IndexEngram(ws, id1, "Go programming language", "", "Go is a statically typed compiled language", []string{"golang", "compiled"})
	_ = idx.IndexEngram(ws, id2, "PostgreSQL database", "", "PostgreSQL is a relational database system", []string{"database", "sql"})
	_ = idx.IndexEngram(ws, id3, "Machine learning", "", "Machine learning algorithms learn from data", []string{"ml", "ai"})

	// Query about compiled language should rank Go first
	results, err := idx.Search(ctx, ws, "compiled programming language", 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("Search returned 0 results")
	}
	if results[0].ID != id1 {
		t.Errorf("top result ID = %x, want %x (Go engram)", results[0].ID, id1)
	}

	// Query about database should rank PostgreSQL first
	results, err = idx.Search(ctx, ws, "relational database SQL", 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("Search returned 0 results")
	}
	if results[0].ID != id2 {
		t.Errorf("top result ID = %x, want %x (PostgreSQL engram)", results[0].ID, id2)
	}
}

// TestFTSMultipleEngrams verifies df increments correctly across multiple engrams.
func TestFTSMultipleEngrams(t *testing.T) {
	db, cleanup := openTestDB(t)
	defer cleanup()

	idx := New(db)
	store := storage.NewPebbleStore(db, 100)
	ws := store.VaultPrefix("test")
	ctx := context.Background()

	// Index 3 engrams all containing the word "system"
	for i := 0; i < 3; i++ {
		id := [16]byte{byte(i + 1)}
		_ = idx.IndexEngram(ws, id, "system concept", "", "this is a system component", nil)
	}

	stats := idx.readStats(ws)
	if stats.TotalEngrams != 3 {
		t.Errorf("TotalEngrams = %d, want 3", stats.TotalEngrams)
	}

	results, err := idx.Search(ctx, ws, "system component", 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("Search returned %d results, want 3", len(results))
	}
}
