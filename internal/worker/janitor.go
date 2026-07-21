// Xomoi-Core: Sovereign Edge Node
// Copyright (C) 2026 Adrish Bora (@code-grey) & Simanjit Hujuri (@code-zephyrus)
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package worker

import (
	"context"
	"database/sql"
	"log"
	"time"
)

// Janitor is responsible for pruning old telemetry data to prevent disk exhaustion on edge nodes.
type Janitor struct {
	db        *sql.DB
	retention time.Duration
	interval  time.Duration
}

// NewJanitor creates a new Janitor worker. 
// retention defines how long data is kept (e.g., 30 days).
// interval defines how often the janitor runs (e.g., every 24 hours).
func NewJanitor(db *sql.DB, retention time.Duration, interval time.Duration) *Janitor {
	return &Janitor{
		db:        db,
		retention: retention,
		interval:  interval,
	}
}

// Start begins the periodic pruning cycle. Blocks, so run in a goroutine.
func (j *Janitor) Start(ctx context.Context) {
	log.Printf("Background Janitor online. Pruning data older than %v every %v", j.retention, j.interval)
	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Janitor received shutdown signal. Exiting.")
			return
		case <-ticker.C:
			j.Prune()
		}
	}
}

// Prune executes the DELETE query for old telemetry data.
func (j *Janitor) Prune() {
	cutoff := time.Now().Add(-j.retention)
	
	// Execute deletion on the new TSDB table
	res, err := j.db.Exec("DELETE FROM telemetry_history WHERE timestamp < ?", cutoff)
	if err != nil {
		log.Printf("[JANITOR] Failed to prune telemetry: %v", err)
		return
	}

	rows, _ := res.RowsAffected()
	if rows > 0 {
		log.Printf("[JANITOR] Successfully pruned %d old telemetry records.", rows)
	}
}
