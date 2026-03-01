package engine

import (
	"context"

	"github.com/scrypster/muninndb/internal/storage"
)

// WhereLeftOff returns the most recently accessed non-deleted, non-completed engrams
// sorted by LastAccess descending. Uses ListEngrams with sort="accessed" to avoid
// duplicating scan logic.
func (e *Engine) WhereLeftOff(ctx context.Context, vault string, limit int) ([]*storage.Engram, error) {
	result, err := e.ListEngrams(ctx, ListEngramsParams{
		Vault: vault,
		Limit: limit,
		Sort:  "accessed",
	})
	if err != nil {
		return nil, err
	}

	// ListEngrams already excludes soft-deleted; additionally exclude completed engrams.
	filtered := result.Engrams[:0]
	for _, eng := range result.Engrams {
		if eng.State != storage.StateCompleted {
			filtered = append(filtered, eng)
		}
	}
	return filtered, nil
}
