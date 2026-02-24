package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/scrypster/muninndb/internal/cognitive"
	"github.com/scrypster/muninndb/internal/engine"
	"github.com/scrypster/muninndb/internal/engine/activation"
	"github.com/scrypster/muninndb/internal/engine/trigger"
	"github.com/scrypster/muninndb/internal/index/fts"
	hnswpkg "github.com/scrypster/muninndb/internal/index/hnsw"
	"github.com/scrypster/muninndb/internal/plugin"
	embedpkg "github.com/scrypster/muninndb/internal/plugin/embed"
	"github.com/scrypster/muninndb/internal/storage"
	"github.com/scrypster/muninndb/internal/transport/mbp"
)

// evalEngine bundles all the engine components needed for Bible eval.
type evalEngine struct {
	eng           *engine.Engine
	store         *storage.PebbleStore
	hnswReg       *hnswpkg.Registry
	embedder      activation.Embedder
	ws            [8]byte
	cancel        context.CancelFunc
	hebbianWorker *cognitive.HebbianWorker
}

// newEvalEngine initialises the full MuninnDB engine stack for evaluation.
// It follows cmd/eval/main.go exactly: OpenPebble → NewPebbleStore → fts.New →
// hnswpkg.NewRegistry → embedpkg.NewEmbedService → activation.New → engine.NewEngine.
func newEvalEngine(ctx context.Context, dataDir string) (*evalEngine, error) {
	db, err := storage.OpenPebble(dataDir, storage.DefaultOptions())
	if err != nil {
		return nil, fmt.Errorf("open pebble: %w", err)
	}

	store := storage.NewPebbleStore(db, 100_000)
	ftsIdx := fts.New(db)
	hnswReg := hnswpkg.NewRegistry(db)

	// Use bundled all-MiniLM-L6-v2 embedder (FTS-only fallback if unavailable)
	svc, initErr := embedpkg.NewEmbedService("local://all-MiniLM-L6-v2")
	if initErr != nil {
		// Fall back to no HNSW — FTS only
		svc = nil
	}
	var embedder activation.Embedder
	var hnswIdx activation.HNSWIndex
	if svc != nil {
		if initErr = svc.Init(ctx, plugin.PluginConfig{DataDir: dataDir}); initErr != nil {
			// Model assets not present — run FTS-only
			svc = nil
		}
	}
	if svc != nil {
		embedder = &bibleEmbedder{svc: svc}
		hnswIdx = &hnswActAdapter{r: hnswReg}
	} else {
		embedder = &hashEmbedder{}
		hnswIdx = nil
	}

	actEngine := activation.New(store, &ftsAdapter{ftsIdx}, hnswIdx, embedder)
	trigSystem := trigger.New(store, &ftsTrigAdapter{ftsIdx}, nil, embedder)

	hebbianWorker := cognitive.NewHebbianWorker(&bibleHebbianAdapter{store})
	decayWorker := cognitive.NewDecayWorker(&bibleDecayAdapter{store})
	contradictWorker := cognitive.NewContradictWorker(&bibleContradictAdapter{store})
	confidenceWorker := cognitive.NewConfidenceWorker(&bibleConfidenceAdapter{store})

	workerCtx, workerCancel := context.WithCancel(context.Background())
	go decayWorker.Worker.Run(workerCtx)
	go contradictWorker.Worker.Run(workerCtx)
	go confidenceWorker.Worker.Run(workerCtx)

	eng := engine.NewEngine(
		store, nil, ftsIdx, actEngine, trigSystem,
		hebbianWorker, decayWorker,
		contradictWorker.Worker, confidenceWorker.Worker,
		embedder, hnswReg,
	)

	ws := store.ResolveVaultPrefix("bible")

	return &evalEngine{
		eng:           eng,
		store:         store,
		hnswReg:       hnswReg,
		embedder:      embedder,
		ws:            ws,
		cancel:        workerCancel,
		hebbianWorker: hebbianWorker,
	}, nil
}

// close stops all background workers and releases storage.
// Order mirrors cmd/eval/main.go: cancel context → stop Hebbian → stop engine → close store.
func (ee *evalEngine) close() {
	ee.cancel()
	if ee.hebbianWorker != nil {
		ee.hebbianWorker.Stop()
	}
	ee.eng.Stop()
	ee.store.Close()
}

// writeVerse writes one verse WriteRequest to the engine and indexes it in HNSW.
func (ee *evalEngine) writeVerse(ctx context.Context, req mbp.WriteRequest) (storage.ULID, error) {
	req.Vault = "bible"
	resp, err := ee.eng.Write(ctx, &req)
	if err != nil {
		return storage.ULID{}, fmt.Errorf("write verse %q: %w", req.Concept, err)
	}

	id, parseErr := storage.ParseULID(resp.ID)
	if parseErr != nil {
		return storage.ULID{}, fmt.Errorf("parse ULID %q: %w", resp.ID, parseErr)
	}

	// Embed and insert into HNSW if index is available.
	if ee.hnswReg != nil {
		text := req.Concept
		if req.Content != "" {
			text += ". " + req.Content
		}
		vec, embedErr := ee.embedder.Embed(ctx, []string{text})
		if embedErr == nil {
			_ = ee.hnswReg.Insert(ctx, ee.ws, [16]byte(id), vec)
		}
	}

	return id, nil
}

// errStopScan is a sentinel used to stop ScanEngrams early once a target is found.
var errStopScan = errors.New("stop")

// setEngramState finds an engram by concept and overwrites its cognitive state fields.
// This is used by Phase 2 to install hand-crafted decay/access state into the corpus.
func (ee *evalEngine) setEngramState(ctx context.Context, concept string, lastAccess time.Time, accessCount uint32, stability float32) error {
	var targetID storage.ULID
	var found bool
	scanErr := ee.store.ScanEngrams(ctx, ee.ws, func(eng *storage.Engram) error {
		if eng.Concept == concept {
			targetID = eng.ID
			found = true
			return errStopScan
		}
		return nil
	})
	if scanErr != nil && !errors.Is(scanErr, errStopScan) {
		return fmt.Errorf("scan for %q: %w", concept, scanErr)
	}
	if !found {
		return fmt.Errorf("engram not found: %q", concept)
	}
	eng, err := ee.store.GetEngram(ctx, ee.ws, targetID)
	if err != nil {
		return fmt.Errorf("get engram %q: %w", concept, err)
	}
	eng.LastAccess = lastAccess
	eng.AccessCount = accessCount
	eng.Stability = stability
	_, err = ee.store.WriteEngram(ctx, ee.ws, eng)
	return err
}

// reloadHNSW re-inserts all vault engrams into the HNSW index.
// Used after vault import to rebuild the in-memory index from stored engrams.
func (ee *evalEngine) reloadHNSW(ctx context.Context) error {
	if ee.hnswReg == nil {
		return nil
	}
	return ee.store.ScanEngrams(ctx, ee.ws, func(eng *storage.Engram) error {
		text := eng.Concept
		if eng.Content != "" {
			text += ". " + eng.Content
		}
		vec, embedErr := ee.embedder.Embed(ctx, []string{text})
		if embedErr != nil {
			return nil // skip individual failures
		}
		_ = ee.hnswReg.Insert(ctx, ee.ws, [16]byte(eng.ID), vec)
		return nil
	})
}

// importVault imports a .muninn archive into the vault and reloads HNSW.
func (ee *evalEngine) importVault(ctx context.Context, r io.Reader) error {
	opts := storage.ImportOpts{}
	if _, err := ee.store.ImportVaultData(ctx, ee.ws, "bible", opts, r); err != nil {
		return fmt.Errorf("import vault data: %w", err)
	}
	return ee.reloadHNSW(ctx)
}

// exportVault writes the vault as a .muninn archive to w.
func (ee *evalEngine) exportVault(ctx context.Context, w io.Writer) error {
	opts := storage.ExportOpts{}
	_, err := ee.store.ExportVaultData(ctx, ee.ws, "bible", opts, w)
	return err
}

// activate queries the engine with the given context strings and returns the top-10 results.
func (ee *evalEngine) activate(ctx context.Context, contextStrs []string) ([]mbp.ActivationItem, error) {
	resp, err := ee.eng.Activate(ctx, &mbp.ActivateRequest{
		Context:    contextStrs,
		MaxResults: 10,
		Vault:      "bible",
		BriefMode:  "off",
	})
	if err != nil {
		return nil, err
	}
	return resp.Activations, nil
}
