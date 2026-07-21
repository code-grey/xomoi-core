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

package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/code-grey/xomoi-core/internal/config"
	"github.com/gorilla/websocket"
)

type LogBuffer struct {
	mu   sync.Mutex
	logs []string
}

func (b *LogBuffer) Write(p []byte) (n int, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.logs = append(b.logs, string(bytes.TrimSpace(p)))
	if len(b.logs) > 50 {
		b.logs = b.logs[1:]
	}
	return len(p), nil
}

func (b *LogBuffer) GetAll() []string {
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.logs) == 0 {
		return nil
	}
	res := make([]string, len(b.logs))
	copy(res, b.logs)
	return res
}

var GlobalLogBuffer = &LogBuffer{}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all for UI development
	},
}

type HealthStats struct {
	RamUsageMB  float64  `json:"ram_usage_mb"`
	NumWorkers  int      `json:"num_workers"`
	NumCPU      int      `json:"num_cpu"`
	WalSizeMB   float64  `json:"wal_size_mb"`
	UptimeSec   int64    `json:"uptime_sec"`
	GcPausesNs  uint64   `json:"gc_pauses_ns"`
	NumGC       uint32   `json:"num_gc"`
	HeapSysMb   float64  `json:"heap_sys_mb"`
	Goroutines  int      `json:"goroutines"`
	NewLogs     []string `json:"new_logs"`
}

var serverStartTime = time.Now()

func HealthWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			var m runtime.MemStats
			runtime.ReadMemStats(&m)

			var walSize float64 = 0
			if info, err := os.Stat(config.Load().DBPath + "-wal"); err == nil {
				walSize = float64(info.Size()) / 1024 / 1024
			}

			stats := HealthStats{
				RamUsageMB:  float64(m.Alloc) / 1024 / 1024,
				NumWorkers:  config.Load().IngestionWorkers,
				NumCPU:      runtime.NumCPU(),
				WalSizeMB:   walSize,
				UptimeSec:   int64(time.Since(serverStartTime).Seconds()),
				GcPausesNs:  m.PauseNs[(m.NumGC+255)%256],
				NumGC:       m.NumGC,
				HeapSysMb:   float64(m.HeapSys) / 1024 / 1024,
				Goroutines:  runtime.NumGoroutine(),
				NewLogs:     GlobalLogBuffer.GetAll(),
			}

			msg, _ := json.Marshal(stats)
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return // Client disconnected
			}
		}
	}
}
