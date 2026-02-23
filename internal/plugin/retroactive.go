package plugin

import (
	"context"
	"log/slog"
	"runtime"
	"sync"
	"time"
)

// RetroactiveProcessor processes engrams asynchronously with a registered plugin.
// It scans for engrams missing a digest flag and calls the plugin to process them.
type RetroactiveProcessor struct {
	store    PluginStore
	plugin   Plugin
	flagBit  uint8              // DigestEmbed or DigestEnrich
	stats    RetroactiveStats
	statsMu  sync.RWMutex
	cancelFn context.CancelFunc
	wg       sync.WaitGroup
}

// NewRetroactiveProcessor creates a new processor for a plugin.
func NewRetroactiveProcessor(store PluginStore, p Plugin, flagBit uint8) *RetroactiveProcessor {
	return &RetroactiveProcessor{
		store:   store,
		plugin:  p,
		flagBit: flagBit,
	}
}

// Start launches the background processing goroutine.
func (rp *RetroactiveProcessor) Start(ctx context.Context) {
	// Create a cancellable context
	ctx, rp.cancelFn = context.WithCancel(ctx)

	rp.wg.Add(1)
	go rp.run(ctx)
}

// Stop gracefully shuts down the processor.
func (rp *RetroactiveProcessor) Stop() {
	if rp.cancelFn != nil {
		rp.cancelFn()
	}
	rp.wg.Wait()
}

// Stats returns a copy of the current processor statistics.
func (rp *RetroactiveProcessor) Stats() RetroactiveStats {
	rp.statsMu.RLock()
	defer rp.statsMu.RUnlock()
	return rp.stats
}

func (rp *RetroactiveProcessor) run(ctx context.Context) {
	defer rp.wg.Done()

	// Update initial stats
	rp.statsMu.Lock()
	rp.stats.PluginName = rp.plugin.Name()
	rp.stats.Status = "running"
	rp.stats.StartedAt = time.Now()
	rp.statsMu.Unlock()

	// Count total unprocessed engrams
	total, err := rp.store.CountWithoutFlag(ctx, rp.flagBit)
	if err != nil {
		slog.Error("retroactive processor: count failed", "plugin", rp.plugin.Name(), "error", err)
		rp.statsMu.Lock()
		rp.stats.Status = "failed"
		rp.statsMu.Unlock()
		return
	}

	rp.statsMu.Lock()
	rp.stats.Total = total
	rp.statsMu.Unlock()

	// If nothing to do, return early
	if total == 0 {
		slog.Info("retroactive processor: no unprocessed engrams", "plugin", rp.plugin.Name())
		rp.statsMu.Lock()
		rp.stats.Status = "complete"
		rp.statsMu.Unlock()
		return
	}

	slog.Info("retroactive processor: starting", "plugin", rp.plugin.Name(), "total", total)

	// Create iterator
	iter := rp.store.ScanWithoutFlag(ctx, rp.flagBit)
	if iter == nil {
		slog.Error("retroactive processor: failed to create iterator", "plugin", rp.plugin.Name())
		rp.statsMu.Lock()
		rp.stats.Status = "failed"
		rp.statsMu.Unlock()
		return
	}
	defer iter.Close()

	startTime := time.Now()
	batchCount := 0

	// Process engrams
	for iter.Next() {
		// Check context cancellation
		select {
		case <-ctx.Done():
			slog.Info("retroactive processor: cancelled", "plugin", rp.plugin.Name())
			rp.statsMu.Lock()
			rp.stats.Status = "paused"
			rp.statsMu.Unlock()
			return
		default:
		}

		eng := iter.Engram()
		if eng == nil {
			continue
		}

		// Process engram based on plugin type
		err := rp.processEngram(ctx, eng)
		if err != nil {
			slog.Warn("retroactive processor: failed to process engram",
				"plugin", rp.plugin.Name(),
				"engram_id", eng.ID.String(),
				"error", err)
			rp.statsMu.Lock()
			rp.stats.Errors++
			rp.statsMu.Unlock()
			continue
		}

		// Set the digest flag
		if err := rp.store.SetDigestFlag(ctx, eng.ID, rp.flagBit); err != nil {
			slog.Warn("retroactive processor: failed to set digest flag",
				"plugin", rp.plugin.Name(),
				"engram_id", eng.ID.String(),
				"error", err)
			rp.statsMu.Lock()
			rp.stats.Errors++
			rp.statsMu.Unlock()
			continue
		}

		// Increment processed counter
		rp.statsMu.Lock()
		rp.stats.Processed++
		processed := rp.stats.Processed
		rp.statsMu.Unlock()

		batchCount++

		// Yield to scheduler every 100 engrams
		if batchCount%100 == 0 {
			runtime.Gosched()
		}

		// Log progress every 1000 engrams
		if processed%1000 == 0 {
			elapsed := time.Since(startTime).Seconds()
			if elapsed > 0 {
				rate := float64(processed) / elapsed
				remaining := total - processed
				etaSeconds := int64(float64(remaining) / rate)

				rp.statsMu.Lock()
				rp.stats.RatePerSec = rate
				rp.stats.ETASeconds = etaSeconds
				rp.statsMu.Unlock()

				slog.Info("retroactive processor: progress",
					"plugin", rp.plugin.Name(),
					"processed", processed,
					"total", total,
					"rate_per_sec", rate,
					"eta_seconds", etaSeconds)
			}
		}
	}

	// Mark as complete
	rp.statsMu.Lock()
	rp.stats.Status = "complete"
	rp.statsMu.Unlock()

	slog.Info("retroactive processor: complete",
		"plugin", rp.plugin.Name(),
		"processed", rp.stats.Processed,
		"errors", rp.stats.Errors)
}

func (rp *RetroactiveProcessor) processEngram(ctx context.Context, eng *Engram) error {
	// Check if this is an embed plugin
	if embed, ok := rp.plugin.(EmbedPlugin); ok {
		// Call Embed with the concept and content
		text := eng.Concept + " " + eng.Content
		vec, err := embed.Embed(ctx, []string{text})
		if err != nil {
			return err
		}

		// Store the embedding
		if err := rp.store.UpdateEmbedding(ctx, eng.ID, vec); err != nil {
			return err
		}

		// Insert into HNSW index
		if err := rp.store.HNSWInsert(ctx, eng.ID, vec); err != nil {
			return err
		}

		// Auto-link by embedding
		if err := rp.store.AutoLinkByEmbedding(ctx, eng.ID, vec); err != nil {
			return err
		}

		return nil
	}

	// Check if this is an enrich plugin
	if enrich, ok := rp.plugin.(EnrichPlugin); ok {
		// Call Enrich
		result, err := enrich.Enrich(ctx, eng)
		if err != nil {
			return err
		}

		// Store the enrichment result
		if err := rp.store.UpdateDigest(ctx, eng.ID, result); err != nil {
			return err
		}

		// Upsert entities
		for _, entity := range result.Entities {
			if err := rp.store.UpsertEntity(ctx, entity); err != nil {
				// Log but don't fail the whole engram
				slog.Warn("failed to upsert entity", "error", err)
			}
		}

		// Upsert relationships
		for _, rel := range result.Relationships {
			if err := rp.store.UpsertRelationship(ctx, eng.ID, rel); err != nil {
				// Log but don't fail the whole engram
				slog.Warn("failed to upsert relationship", "error", err)
			}
		}

		return nil
	}

	return nil
}
