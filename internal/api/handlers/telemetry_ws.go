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
	"encoding/json"
	"math/rand"
	"net/http"
	"time"
)

type TelemetryData struct {
	Temperature   float64 `json:"temperature"`
	Humidity      float64 `json:"humidity"`
	Pressure      float64 `json:"pressure"`
	ActiveDevices int     `json:"active_devices"`
}

func TelemetryWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// Base mock values to create smooth sine-wave-like charts
	baseTemp := 24.5
	baseHum := 45.0

	for {
		select {
		case <-ticker.C:
			// Add slight random drift
			baseTemp += (rand.Float64() - 0.5) * 0.5
			baseHum += (rand.Float64() - 0.5) * 1.5

			payload := TelemetryData{
				Temperature:   baseTemp,
				Humidity:      baseHum,
				Pressure:      1012.5 + (rand.Float64()-0.5)*2,
				ActiveDevices: 3, // Mock endpoint microcontrollers
			}

			msg, _ := json.Marshal(payload)
			if err := conn.WriteMessage(1, msg); err != nil {
				return // Client disconnected
			}
		}
	}
}
