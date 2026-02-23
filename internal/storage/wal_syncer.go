package storage

import (
	"errors"
	"log/slog"
	"time"

	"github.com/cockroachdb/pebble"
)

const walSyncInterval = 10 * time.Millisecond

// walSyncer periodically calls db.LogData(nil, pebble.Sync) to flush the WAL
// without triggering a memtable flush. This provides group-commit semantics:
// all batch.Commit(pebble.NoSync) writes accumulate in the WAL and are durably
// fsynced every walSyncInterval (default 10ms). Max data loss on crash: 10ms.
//
// This is the same trade-off as MySQL innodb_flush_log_at_trx_commit=2 or
// PostgreSQL synchronous_commit=off, and is safe because Pebble's own WAL
// provides crash recovery — the LogData sync covers all preceding NoSync writes.
type walSyncer struct {
	db   *pebble.DB
	stop chan struct{}
	done chan struct{}
}

func newWALSyncer(db *pebble.DB) *walSyncer {
	s := &walSyncer{
		db:   db,
		stop: make(chan struct{}),
		done: make(chan struct{}),
	}
	go s.run()
	return s
}

func (s *walSyncer) run() {
	defer close(s.done)
	// Recover from the "pebble: closed" panic that can occur if db.Close()
	// races with an in-flight ticker sync during shutdown.  Pebble panics with
	// pebble.ErrClosed (an error value), so we check via errors.Is.
	// Any other unexpected panic is re-panicked so it is not silently swallowed.
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok && errors.Is(err, pebble.ErrClosed) {
				return // expected during shutdown
			}
			panic(r) // unexpected — re-panic
		}
	}()

	ticker := time.NewTicker(walSyncInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := s.db.LogData(nil, pebble.Sync); err != nil {
				slog.Warn("storage: WAL sync failed", "err", err)
			}
		case <-s.stop:
			// Final sync before shutdown.
			_ = s.db.LogData(nil, pebble.Sync)
			return
		}
	}
}

// Close signals the syncer to stop and blocks until the final sync completes.
// Must be called before db.Close().
func (s *walSyncer) Close() {
	close(s.stop)
	<-s.done
}
