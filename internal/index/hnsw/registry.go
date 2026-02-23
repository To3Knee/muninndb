package hnsw

import (
	"context"
	"sync"

	"github.com/cockroachdb/pebble"
)

// Registry is a multi-vault HNSW index registry.
// It lazily creates and caches one *Index per vault workspace prefix.
// It implements both activation.HNSWIndex and trigger.HNSWIndex (both have the
// same Search signature: Search(ctx, ws [8]byte, vec []float32, topK int) ([]ScoredID, error)).
type Registry struct {
	mu      sync.RWMutex
	indexes map[[8]byte]*Index
	db      *pebble.DB
}

// NewRegistry creates a new Registry backed by the provided Pebble database.
func NewRegistry(db *pebble.DB) *Registry {
	return &Registry{
		indexes: make(map[[8]byte]*Index),
		db:      db,
	}
}

// getOrCreate returns the per-vault Index, creating it lazily if it doesn't exist.
// On creation it calls LoadFromPebble to restore any previously persisted nodes.
func (r *Registry) getOrCreate(ws [8]byte) *Index {
	// Fast path: read lock
	r.mu.RLock()
	idx, ok := r.indexes[ws]
	r.mu.RUnlock()
	if ok {
		return idx
	}

	// Slow path: create under write lock
	r.mu.Lock()
	defer r.mu.Unlock()

	// Double-check after acquiring write lock
	if idx, ok = r.indexes[ws]; ok {
		return idx
	}

	idx = New(r.db, ws)
	// Load any previously persisted nodes (ignore error — empty index is fine)
	_ = idx.LoadFromPebble()
	r.indexes[ws] = idx
	return idx
}

// Search implements activation.HNSWIndex and trigger.HNSWIndex.
// It delegates to the per-vault Index.
func (r *Registry) Search(ctx context.Context, ws [8]byte, vec []float32, topK int) ([]ScoredID, error) {
	idx := r.getOrCreate(ws)
	return idx.Search(ctx, vec, topK)
}

// TotalVectors returns the total number of indexed vectors across all vaults.
func (r *Registry) TotalVectors() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	total := 0
	for _, idx := range r.indexes {
		total += idx.Len()
	}
	return total
}

// ResetVault drops the in-memory HNSW index for the given vault workspace prefix.
// Called by ClearVault and DeleteVault to evict stale index state after the
// underlying storage has been wiped. The next Insert or Search call will
// recreate the index lazily (empty, since Pebble data is gone).
func (r *Registry) ResetVault(ws [8]byte) {
	r.mu.Lock()
	delete(r.indexes, ws)
	r.mu.Unlock()
}

// Insert adds a vector to the appropriate per-vault Index.
func (r *Registry) Insert(ctx context.Context, ws [8]byte, id [16]byte, vec []float32) error {
	idx := r.getOrCreate(ws)
	// Store vector first so the graph can fetch it during traversal.
	if err := idx.StoreVector(id, vec); err != nil {
		return err
	}
	// If the in-memory graph insertion panics or a future error path is added,
	// clean up the orphaned vector so it is never stranded in storage unreachable
	// by graph traversal.
	insertOK := false
	defer func() {
		if !insertOK {
			_ = idx.DeleteVector(id) // cleanup orphan on Insert failure
		}
	}()
	idx.Insert(id, vec)
	insertOK = true
	return nil
}
