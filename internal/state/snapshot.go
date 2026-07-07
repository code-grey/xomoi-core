package state

import (
	"context"
	"database/sql"
	"log"
	"time"
)

// SnapshotWorker is responsible for bulk-flushing the HotState to SQLite.
type SnapshotWorker struct {
	hotState *HotState
	db       *sql.DB
	interval time.Duration
}

// NewSnapshotWorker creates a new worker for flushing state.
func NewSnapshotWorker(hs *HotState, db *sql.DB, flushInterval time.Duration) *SnapshotWorker {
	return &SnapshotWorker{
		hotState: hs,
		db:       db,
		interval: flushInterval,
	}
}

// Start begins the periodic flush cycle. This blocks, so run in a goroutine.
func (s *SnapshotWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Snapshot worker received shutdown signal. Performing final flush...")
			s.ForceFlush()
			return
		case <-ticker.C:
			s.ForceFlush()
		}
	}
}

// ForceFlush takes the current state snapshot and writes it to SQLite in a single transaction.
func (s *SnapshotWorker) ForceFlush() {
	states := s.hotState.GetAll()
	if len(states) == 0 {
		return
	}

	tx, err := s.db.Begin()
	if err != nil {
		log.Printf("Failed to begin flush transaction: %v", err)
		return
	}
	defer tx.Rollback()

	// Prepared statement for fast bulk inserts
	stmt, err := tx.Prepare("INSERT INTO telemetry (device_id, timestamp, payload) VALUES (?, ?, ?)")
	if err != nil {
		log.Printf("Failed to prepare flush statement: %v", err)
		return
	}
	defer stmt.Close()

	for _, state := range states {
		_, err := stmt.Exec(state.DeviceID, state.LastUpdate, state.Payload)
		if err != nil {
			log.Printf("Failed to insert state for device %s: %v", state.DeviceID, err)
			// We intentionally do not abort the entire transaction for a single faulty row
		}
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Failed to commit flush transaction: %v", err)
		return
	}
	
	log.Printf("Successfully flushed %d device states to SQLite.", len(states))
}
