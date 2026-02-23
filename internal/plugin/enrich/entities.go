package enrich

import (
	"context"

	"github.com/scrypster/muninndb/internal/plugin"
)

// StoreEntities persists extracted entities and links them to an engram.
func StoreEntities(ctx context.Context, store plugin.PluginStore, engramID plugin.ULID, entities []plugin.ExtractedEntity) error {
	for _, entity := range entities {
		if err := store.UpsertEntity(ctx, entity); err != nil {
			return err
		}
		if err := store.LinkEngramToEntity(ctx, engramID, entity.Name); err != nil {
			return err
		}
	}
	return nil
}

// StoreRelationships persists extracted relationships for an engram.
func StoreRelationships(ctx context.Context, store plugin.PluginStore, engramID plugin.ULID, relationships []plugin.ExtractedRelation) error {
	for _, rel := range relationships {
		if err := store.UpsertRelationship(ctx, engramID, rel); err != nil {
			return err
		}
	}
	return nil
}
