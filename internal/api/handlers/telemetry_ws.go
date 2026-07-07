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
