package storage

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"sort"
	"time"

	"github.com/cockroachdb/pebble"
	"github.com/scrypster/muninndb/internal/provenance"
	"github.com/scrypster/muninndb/internal/storage/erf"
	"github.com/scrypster/muninndb/internal/storage/keys"
)

// GetEngram reads a full engram record by ID.
func (ps *PebbleStore) GetEngram(ctx context.Context, wsPrefix [8]byte, id ULID) (*Engram, error) {
	// Check L1 cache first (vault-scoped to prevent cross-vault cache hits).
	if eng, found := ps.cache.Get(wsPrefix, id); found {
		return eng, nil
	}

	// Get from pebble
	key := keys.EngramKey(wsPrefix, [16]byte(id))
	val, err := Get(ps.db, key)
	if err != nil {
		return nil, fmt.Errorf("get engram: %w", err)
	}
	if val == nil {
		return nil, fmt.Errorf("engram not found")
	}

	// Decode
	erfEng, err := erf.Decode(val)
	if err != nil {
		return nil, fmt.Errorf("decode engram: %w", err)
	}

	// Convert back to storage.Engram
	eng := fromERFEngram(erfEng)

	// Cache it (vault-scoped).
	ps.cache.Set(wsPrefix, id, eng)

	return eng, nil
}

// EngramLastAccessNs returns the nanosecond timestamp of the last time the engram
// was served from the L1 cache. Returns 0 if not cached (caller should fall back to eng.LastAccess).
func (ps *PebbleStore) EngramLastAccessNs(wsPrefix [8]byte, id ULID) int64 {
	return ps.cache.LastAccessNs(wsPrefix, id)
}

// GetEngrams batch-reads full engram records.
//
// Fast path: L1-cached engrams are served without touching Pebble.
// Slow path: cache-cold IDs are read with a SINGLE Pebble iterator using
// sorted forward seeks — O(1) iterator open + N seeks instead of N snapshot
// acquisitions and N separate bloom-filter probes. OS readahead also kicks in
// as the seeks are sequential in key order.
//
// Missing engrams (deleted or dangling association edges) are returned as nil.
// Callers must check for nil before dereferencing.
func (ps *PebbleStore) GetEngrams(ctx context.Context, wsPrefix [8]byte, ids []ULID) ([]*Engram, error) {
	result := make([]*Engram, len(ids))

	// Phase 1: serve L1-cached engrams without touching Pebble.
	type uncachedEntry struct {
		resultIdx int
		id        ULID
		key       []byte
	}
	var uncached []uncachedEntry
	for i, id := range ids {
		if eng, found := ps.cache.Get(wsPrefix, id); found {
			result[i] = eng
		} else {
			uncached = append(uncached, uncachedEntry{
				resultIdx: i,
				id:        id,
				key:       keys.EngramKey(wsPrefix, [16]byte(id)),
			})
		}
	}
	if len(uncached) == 0 {
		return result, nil
	}

	// Phase 2: sort by key order so all Pebble seeks are strictly forward.
	sort.Slice(uncached, func(i, j int) bool {
		return bytes.Compare(uncached[i].key, uncached[j].key) < 0
	})

	// Phase 3: open ONE iterator spanning the range of needed keys.
	lower := uncached[0].key
	// Upper bound: copy the last key and increment its last byte.
	lastKey := uncached[len(uncached)-1].key
	upper := make([]byte, len(lastKey)+1) // +1 guarantees we include lastKey
	copy(upper, lastKey)
	carried := true
	for i := len(lastKey) - 1; i >= 0; i-- {
		upper[i]++
		if upper[i] != 0 {
			upper = upper[:len(lastKey)]
			carried = false
			break
		}
	}
	if carried {
		// All bytes were 0xFF and wrapped; restore lastKey and keep the +1 trailing 0x00.
		copy(upper, lastKey)
	}

	iter, err := ps.db.NewIter(&pebble.IterOptions{
		LowerBound: lower,
		UpperBound: upper,
	})
	if err != nil {
		// Fallback: individual GetEngram calls.
		for _, u := range uncached {
			eng, _ := ps.GetEngram(ctx, wsPrefix, u.id)
			result[u.resultIdx] = eng
		}
		return result, nil
	}
	defer iter.Close()

	for _, u := range uncached {
		if iter.SeekGE(u.key); !iter.Valid() || !bytes.Equal(iter.Key(), u.key) {
			// Engram not found — dangling edge or soft-deleted; leave result[i] = nil.
			continue
		}
		val := make([]byte, len(iter.Value()))
		copy(val, iter.Value())
		erfEng, err := erf.Decode(val)
		if err != nil {
			continue
		}
		eng := fromERFEngram(erfEng)
		ps.cache.Set(wsPrefix, u.id, eng)
		result[u.resultIdx] = eng
	}

	return result, nil
}

// GetMetadata reads only the metadata fields for a batch of engrams.
// Uses a two-level cache: metaCache (metadata-only) → L1 engram cache → Pebble.
// Hot engrams (repeatedly activated) are served entirely from in-memory caches.
func (ps *PebbleStore) GetMetadata(ctx context.Context, wsPrefix [8]byte, ids []ULID) ([]*EngramMeta, error) {
	result := make([]*EngramMeta, len(ids))
	for i, id := range ids {
		// Level 1: metadata-only cache (populated after first Pebble read).
		if v, ok := ps.metaCache.Load([16]byte(id)); ok {
			result[i] = v.(*EngramMeta)
			continue
		}

		// Level 2: full engram L1 cache — extract metadata fields without Pebble read.
		if eng, found := ps.cache.Get(wsPrefix, id); found {
			meta := &EngramMeta{
				ID:          eng.ID,
				CreatedAt:   eng.CreatedAt,
				UpdatedAt:   eng.UpdatedAt,
				LastAccess:  eng.LastAccess,
				Confidence:  eng.Confidence,
				Relevance:   eng.Relevance,
				Stability:   eng.Stability,
				AccessCount: eng.AccessCount,
				State:       eng.State,
				AssocCount:  uint16(len(eng.Associations)),
				EmbedDim:    eng.EmbedDim,
				MemoryType:  eng.MemoryType,
			}
			ps.metaCache.Store([16]byte(id), meta)
			result[i] = meta
			continue
		}

		// Slow path: read compact metadata key from Pebble.
		key := keys.MetaKey(wsPrefix, [16]byte(id))
		val, err := Get(ps.db, key)
		if err != nil {
			return nil, fmt.Errorf("get metadata: %w", err)
		}
		if val == nil {
			return nil, fmt.Errorf("metadata not found")
		}

		erfMeta, err := erf.DecodeMeta(val)
		if err != nil {
			return nil, fmt.Errorf("decode metadata: %w", err)
		}

		meta := &EngramMeta{
			ID:          ULID(erfMeta.ID),
			CreatedAt:   erfMeta.CreatedAt,
			UpdatedAt:   erfMeta.UpdatedAt,
			LastAccess:  erfMeta.LastAccess,
			Confidence:  erfMeta.Confidence,
			Relevance:   erfMeta.Relevance,
			Stability:   erfMeta.Stability,
			AccessCount: erfMeta.AccessCount,
			State:       LifecycleState(erfMeta.State),
			AssocCount:  erfMeta.AssocCount,
			EmbedDim:    EmbedDimension(erfMeta.EmbedDim),
			MemoryType:  MemoryType(erfMeta.MemoryType),
		}
		// Populate metaCache so subsequent calls for this engram skip Pebble.
		ps.metaCache.Store([16]byte(id), meta)
		result[i] = meta
	}
	return result, nil
}

// UpdateMetadata writes only the metadata fields that changed.
// If the state changes, it also updates the 0x0B state secondary index.
// Patches the raw 0x01 bytes in-place (no full re-encode).
func (ps *PebbleStore) UpdateMetadata(ctx context.Context, wsPrefix [8]byte, id ULID, meta *EngramMeta) error {
	// Read slim metadata to detect state change (needed for index update).
	oldMetas, err := ps.GetMetadata(ctx, wsPrefix, []ULID{id})
	if err != nil {
		return err
	}
	if len(oldMetas) == 0 {
		return fmt.Errorf("engram not found")
	}
	oldState := oldMetas[0].State

	// Read raw 0x01 bytes without decoding the full ERF structure.
	engramKey := keys.EngramKey(wsPrefix, [16]byte(id))
	rawBytes, err := Get(ps.db, engramKey)
	if err != nil {
		return fmt.Errorf("get engram raw: %w", err)
	}
	if rawBytes == nil {
		return fmt.Errorf("engram not found")
	}

	// Patch all mutable metadata fields in-place and recompute CRC32.
	if err := erf.PatchAllMeta(rawBytes,
		meta.UpdatedAt, meta.LastAccess,
		meta.Confidence, meta.Relevance, meta.Stability,
		meta.AccessCount, uint8(meta.State),
	); err != nil {
		return fmt.Errorf("patch metadata: %w", err)
	}

	batch := ps.db.NewBatch()
	defer batch.Close()

	// Update state secondary index if the state changed.
	if oldState != meta.State {
		batch.Delete(keys.StateIndexKey(wsPrefix, uint8(oldState), [16]byte(id)), nil)
		batch.Set(keys.StateIndexKey(wsPrefix, uint8(meta.State), [16]byte(id)), []byte{}, nil)
	}

	batch.Set(engramKey, rawBytes, nil)
	metaKey := keys.MetaKey(wsPrefix, [16]byte(id))
	metaSlice248 := rawBytes
	if len(metaSlice248) > erf.MetaKeySize {
		metaSlice248 = metaSlice248[:erf.MetaKeySize]
	}
	batch.Set(metaKey, metaSlice248, nil)

	// Invalidate L1 cache and metadata cache BEFORE commit — cached structs are stale.
	ps.cache.Delete(wsPrefix, id)
	ps.metaCache.Delete([16]byte(id))

	if err := batch.Commit(pebble.NoSync); err != nil {
		return fmt.Errorf("commit batch: %w", err)
	}

	// Append provenance entry via persistent worker (best effort — drops if full).
	ps.provWork.Submit(wsPrefix, id, provenance.ProvenanceEntry{
		Timestamp: time.Now(),
		Source:    provenance.SourceInferred,
		AgentID:   "system:metadata-update",
		Operation: "update-meta",
		Note:      "",
	})

	return nil
}

// UpdateRelevance updates the relevance and stability of an engram.
// It moves the relevance bucket key (0x10) from the old bucket to the new bucket,
// and patches the raw 0x01 bytes in-place (no full re-encode).
func (ps *PebbleStore) UpdateRelevance(ctx context.Context, wsPrefix [8]byte, id ULID, relevance, stability float32) error {
	// Read slim metadata to get the old relevance for bucket key movement.
	metas, err := ps.GetMetadata(ctx, wsPrefix, []ULID{id})
	if err != nil {
		return err
	}
	if len(metas) == 0 {
		return fmt.Errorf("engram not found")
	}
	oldRelevance := metas[0].Relevance

	// Read raw 0x01 bytes without decoding the full ERF structure.
	engramKey := keys.EngramKey(wsPrefix, [16]byte(id))
	rawBytes, err := Get(ps.db, engramKey)
	if err != nil {
		return fmt.Errorf("get engram raw: %w", err)
	}
	if rawBytes == nil {
		return fmt.Errorf("engram not found")
	}

	// Patch relevance/stability/updatedAt in-place and recompute CRC32.
	if err := erf.PatchRelevance(rawBytes, time.Now(), relevance, stability); err != nil {
		return fmt.Errorf("patch relevance: %w", err)
	}

	batch := ps.db.NewBatch()
	defer batch.Close()

	// Move relevance bucket key.
	batch.Delete(keys.RelevanceBucketKey(wsPrefix, oldRelevance, [16]byte(id)), nil)
	batch.Set(keys.RelevanceBucketKey(wsPrefix, relevance, [16]byte(id)), []byte{}, nil)

	// Write patched records.
	batch.Set(engramKey, rawBytes, nil)
	metaKey := keys.MetaKey(wsPrefix, [16]byte(id))
	metaEnd := erf.MetaKeySize
	if metaEnd > len(rawBytes) {
		metaEnd = len(rawBytes)
	}
	batch.Set(metaKey, rawBytes[:metaEnd], nil)

	// Invalidate L1 cache and metadata cache BEFORE commit — cached structs are stale.
	ps.cache.Delete(wsPrefix, id)
	ps.metaCache.Delete([16]byte(id))

	if err := batch.Commit(pebble.NoSync); err != nil {
		return fmt.Errorf("commit batch: %w", err)
	}

	// Append provenance entry via persistent worker (best effort — drops if full).
	ps.provWork.Submit(wsPrefix, id, provenance.ProvenanceEntry{
		Timestamp: time.Now(),
		Source:    provenance.SourceInferred,
		AgentID:   "system:relevance-update",
		Operation: "update-relevance",
		Note:      "",
	})

	return nil
}

// DeleteEngram performs a hard delete: removes the engram, all association keys,
// and all secondary indexes. Reads the engram first to gather index data.
func (ps *PebbleStore) DeleteEngram(ctx context.Context, wsPrefix [8]byte, id ULID) error {
	// Read engram to collect secondary index data for cleanup.
	eng, err := ps.GetEngram(ctx, wsPrefix, id)
	if err != nil {
		// Not found or unreadable — attempt key-only delete as fallback.
		batch := ps.db.NewBatch()
		defer batch.Close()
		batch.Delete(keys.EngramKey(wsPrefix, [16]byte(id)), nil)
		batch.Delete(keys.MetaKey(wsPrefix, [16]byte(id)), nil)
		ps.cache.Delete(wsPrefix, id)
		return batch.Commit(pebble.NoSync)
	}

	batch := ps.db.NewBatch()
	defer batch.Close()

	// Primary records
	batch.Delete(keys.EngramKey(wsPrefix, [16]byte(id)), nil)
	batch.Delete(keys.MetaKey(wsPrefix, [16]byte(id)), nil)

	// Secondary indexes
	batch.Delete(keys.StateIndexKey(wsPrefix, uint8(eng.State), [16]byte(id)), nil)
	batch.Delete(keys.CreatorIndexKey(wsPrefix, keys.Hash(eng.CreatedBy), [16]byte(id)), nil)
	batch.Delete(keys.RelevanceBucketKey(wsPrefix, eng.Relevance, [16]byte(id)), nil)
	for _, tag := range eng.Tags {
		batch.Delete(keys.TagIndexKey(wsPrefix, keys.Hash(tag), [16]byte(id)), nil)
	}

	// Association forward/reverse keys
	for _, assoc := range eng.Associations {
		batch.Delete(keys.AssocFwdKey(wsPrefix, [16]byte(id), assoc.Weight, [16]byte(assoc.TargetID)), nil)
		batch.Delete(keys.AssocRevKey(wsPrefix, [16]byte(assoc.TargetID), assoc.Weight, [16]byte(id)), nil)
		batch.Delete(keys.AssocWeightIndexKey(wsPrefix, [16]byte(id), [16]byte(assoc.TargetID)), nil)
	}

	if err := batch.Commit(pebble.NoSync); err != nil {
		return fmt.Errorf("delete engram: %w", err)
	}

	ps.cache.Delete(wsPrefix, id)

	// Decrement vault count synchronously to avoid a race where callers
	// observe a stale count after DeleteEngram returns.
	vc := ps.getOrInitCounter(ctx, wsPrefix)
	newCount := vc.count.Add(-1)
	if newCount < 0 {
		vc.count.Store(0)
		newCount = 0
	}
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(newCount))
	_ = ps.db.Set(keys.VaultCountKey(wsPrefix), buf, pebble.NoSync)

	return nil
}

// SoftDelete sets state to StateSoftDeleted and updates the record.
// It also transitions the 0x0B state secondary index from the old state to StateSoftDeleted.
func (ps *PebbleStore) SoftDelete(ctx context.Context, wsPrefix [8]byte, id ULID) error {
	// Read engram
	eng, err := ps.GetEngram(ctx, wsPrefix, id)
	if err != nil {
		return err
	}

	oldState := eng.State

	// Set state to soft deleted
	eng.State = StateSoftDeleted
	eng.UpdatedAt = time.Now()

	// Re-encode
	erfEng := toERFEngram(eng)
	erfBytes, err := erf.Encode(erfEng)
	if err != nil {
		return fmt.Errorf("encode engram: %w", err)
	}

	batch := ps.db.NewBatch()
	defer batch.Close()

	// Move state index entry: delete old, write new.
	oldStateKey := keys.StateIndexKey(wsPrefix, uint8(oldState), [16]byte(id))
	batch.Delete(oldStateKey, nil)
	newStateKey := keys.StateIndexKey(wsPrefix, uint8(StateSoftDeleted), [16]byte(id))
	batch.Set(newStateKey, []byte{}, nil)

	engramKey := keys.EngramKey(wsPrefix, [16]byte(id))
	batch.Set(engramKey, erfBytes, nil)

	metaKey := keys.MetaKey(wsPrefix, [16]byte(id))
	metaSlice437 := erfBytes
	if len(metaSlice437) > erf.MetaKeySize {
		metaSlice437 = metaSlice437[:erf.MetaKeySize]
	}
	batch.Set(metaKey, metaSlice437, nil)

	if err := batch.Commit(pebble.NoSync); err != nil {
		return fmt.Errorf("commit batch: %w", err)
	}

	// Update cache (vault-scoped).
	ps.cache.Set(wsPrefix, id, eng)

	return nil
}

// UpdateTags replaces the tag list on an engram, re-encodes the full record,
// and adds any new tag index entries. Old tag index entries for tags no longer
// present are left as orphans (safe: they point to a valid engram, just stale).
// For the dedup use-case (tags are always a superset) there are no removals.
func (ps *PebbleStore) UpdateTags(ctx context.Context, wsPrefix [8]byte, id ULID, tags []string) error {
	eng, err := ps.GetEngram(ctx, wsPrefix, id)
	if err != nil {
		return err
	}

	eng.Tags = tags
	eng.UpdatedAt = time.Now()

	erfEng := toERFEngram(eng)
	erfBytes, err := erf.Encode(erfEng)
	if err != nil {
		return fmt.Errorf("encode engram: %w", err)
	}

	batch := ps.db.NewBatch()
	defer batch.Close()

	engramKey := keys.EngramKey(wsPrefix, [16]byte(id))
	batch.Set(engramKey, erfBytes, nil)

	metaKey := keys.MetaKey(wsPrefix, [16]byte(id))
	metaSlice := erfBytes
	if len(metaSlice) > erf.MetaKeySize {
		metaSlice = metaSlice[:erf.MetaKeySize]
	}
	batch.Set(metaKey, metaSlice, nil)

	// Write tag index entries for all tags (idempotent for existing tags).
	for _, tag := range tags {
		batch.Set(keys.TagIndexKey(wsPrefix, keys.Hash(tag), [16]byte(id)), []byte{}, nil)
	}

	// Invalidate L1 cache BEFORE commit — cached struct has stale tags.
	ps.cache.Delete(wsPrefix, id)
	ps.metaCache.Delete([16]byte(id))

	if err := batch.Commit(pebble.NoSync); err != nil {
		return fmt.Errorf("commit batch: %w", err)
	}

	return nil
}

// GetEmbedding retrieves the embedding for an engram.
func (ps *PebbleStore) GetEmbedding(ctx context.Context, wsPrefix [8]byte, id ULID) ([]float32, error) {
	eng, err := ps.GetEngram(ctx, wsPrefix, id)
	if err != nil {
		return nil, err
	}
	return eng.Embedding, nil
}

// GetConfidence reads the confidence value from 0x02 metadata for an engram.
func (ps *PebbleStore) GetConfidence(ctx context.Context, wsPrefix [8]byte, id ULID) (float32, error) {
	key := keys.MetaKey(wsPrefix, [16]byte(id))
	val, err := Get(ps.db, key)
	if err != nil {
		return 0.0, fmt.Errorf("get metadata: %w", err)
	}
	if val == nil {
		return 0.0, fmt.Errorf("metadata not found")
	}

	// Decode metadata to extract confidence
	erfMeta, err := erf.DecodeMeta(val)
	if err != nil {
		return 0.0, fmt.Errorf("decode metadata: %w", err)
	}

	return erfMeta.Confidence, nil
}

// UpdateConfidence updates the confidence in 0x02 metadata (and 0x01 full engram).
func (ps *PebbleStore) UpdateConfidence(ctx context.Context, wsPrefix [8]byte, id ULID, confidence float32) error {
	// Read current engram
	eng, err := ps.GetEngram(ctx, wsPrefix, id)
	if err != nil {
		return err
	}

	// Update confidence
	eng.Confidence = confidence
	eng.UpdatedAt = time.Now()

	// Re-encode full engram
	erfEng := toERFEngram(eng)
	erfBytes, err := erf.Encode(erfEng)
	if err != nil {
		return fmt.Errorf("encode engram: %w", err)
	}

	// Write both keys
	batch := ps.db.NewBatch()
	defer batch.Close()

	engramKey := keys.EngramKey(wsPrefix, [16]byte(id))
	batch.Set(engramKey, erfBytes, nil)

	metaKey := keys.MetaKey(wsPrefix, [16]byte(id))
	metaSlice505 := erfBytes
	if len(metaSlice505) > erf.MetaKeySize {
		metaSlice505 = metaSlice505[:erf.MetaKeySize]
	}
	batch.Set(metaKey, metaSlice505, nil)

	if err := batch.Commit(pebble.NoSync); err != nil {
		return fmt.Errorf("commit batch: %w", err)
	}

	// Update cache (vault-scoped).
	ps.cache.Set(wsPrefix, id, eng)
	// Invalidate metadata cache — cached metadata is stale.
	ps.metaCache.Delete([16]byte(id))

	return nil
}

// toERFEngram converts storage.Engram to erf.Engram.
func toERFEngram(eng *Engram) *erf.Engram {
	erfAssocs := make([]erf.Association, len(eng.Associations))
	for i, a := range eng.Associations {
		erfAssocs[i] = erf.Association{
			TargetID:      [16]byte(a.TargetID),
			RelType:       uint16(a.RelType),
			Weight:        a.Weight,
			Confidence:    a.Confidence,
			CreatedAt:     a.CreatedAt,
			LastActivated: a.LastActivated,
		}
	}

	return &erf.Engram{
		ID:             [16]byte(eng.ID),
		CreatedAt:      eng.CreatedAt,
		UpdatedAt:      eng.UpdatedAt,
		LastAccess:     eng.LastAccess,
		Confidence:     eng.Confidence,
		Relevance:      eng.Relevance,
		Stability:      eng.Stability,
		AccessCount:    eng.AccessCount,
		State:          uint8(eng.State),
		EmbedDim:       uint8(eng.EmbedDim),
		Concept:        eng.Concept,
		CreatedBy:      eng.CreatedBy,
		Content:        eng.Content,
		Tags:           eng.Tags,
		Associations:   erfAssocs,
		Embedding:      eng.Embedding,
		Summary:        eng.Summary,
		KeyPoints:      eng.KeyPoints,
		MemoryType:     uint8(eng.MemoryType),
		Classification: eng.Classification,
	}
}

// fromERFEngram converts erf.Engram back to storage.Engram.
func fromERFEngram(e *erf.Engram) *Engram {
	assocs := make([]Association, len(e.Associations))
	for i, a := range e.Associations {
		assocs[i] = Association{
			TargetID:      ULID(a.TargetID),
			RelType:       RelType(a.RelType),
			Weight:        a.Weight,
			Confidence:    a.Confidence,
			CreatedAt:     a.CreatedAt,
			LastActivated: a.LastActivated,
		}
	}

	return &Engram{
		ID:             ULID(e.ID),
		CreatedAt:      e.CreatedAt,
		UpdatedAt:      e.UpdatedAt,
		LastAccess:     e.LastAccess,
		Confidence:     e.Confidence,
		Relevance:      e.Relevance,
		Stability:      e.Stability,
		AccessCount:    e.AccessCount,
		State:          LifecycleState(e.State),
		EmbedDim:       EmbedDimension(e.EmbedDim),
		Concept:        e.Concept,
		CreatedBy:      e.CreatedBy,
		Content:        e.Content,
		Tags:           e.Tags,
		Associations:   assocs,
		Embedding:      e.Embedding,
		Summary:        e.Summary,
		KeyPoints:      e.KeyPoints,
		MemoryType:     MemoryType(e.MemoryType),
		Classification: e.Classification,
	}
}

// ScanEngrams iterates over all engrams in the given vault workspace, calling fn for each.
// Iteration stops early if fn returns a non-nil error.
// Corrupt ERF records are skipped with a warning log.
func (ps *PebbleStore) ScanEngrams(ctx context.Context, ws [8]byte, fn func(*Engram) error) error {
	wsNext := ws
	for i := 7; i >= 0; i-- {
		wsNext[i]++
		if wsNext[i] != 0 {
			break
		}
	}

	lo := make([]byte, 9)
	lo[0] = 0x01
	copy(lo[1:], ws[:])
	hi := make([]byte, 9)
	hi[0] = 0x01
	copy(hi[1:], wsNext[:])

	iter, err := ps.db.NewIter(&pebble.IterOptions{LowerBound: lo, UpperBound: hi})
	if err != nil {
		return fmt.Errorf("scan engrams: create iter: %w", err)
	}
	defer iter.Close()

	for valid := iter.First(); valid; valid = iter.Next() {
		k := iter.Key()
		if len(k) < 25 { // 1 prefix + 8 ws + 16 ULID minimum
			continue
		}

		rawVal := make([]byte, len(iter.Value()))
		copy(rawVal, iter.Value())

		erfEng, decErr := erf.Decode(rawVal)
		if decErr != nil {
			continue
		}

		eng := fromERFEngram(erfEng)
		if err := fn(eng); err != nil {
			return err
		}
	}
	return iter.Error()
}
