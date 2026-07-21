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
	"fmt"
	"log/slog"
	"net/http"

	"github.com/code-grey/xomoi-core/internal/api/response"
)

// isValidMAC provides ultra-fast, zero-allocation MAC address validation
func isValidMAC(mac string) bool {
	if len(mac) != 17 {
		return false
	}
	for i := 0; i < len(mac); i++ {
		c := mac[i]
		if i%3 == 2 {
			if c != ':' && c != '_' {
				return false
			}
		} else {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
				return false
			}
		}
	}
	return true
}

// DeviceConfig represents the dynamically updatable NVS parameters on the ESP32
type DeviceConfig struct {
	PingFrequencyMs int    `json:"ping_frequency_ms,omitempty"`
	Mode            string `json:"mode,omitempty"` // "batch", "realtime", "smart"
}

type ConfigHandler struct {
	publisher MQTTPublisher
}

func NewConfigHandler(pub MQTTPublisher) *ConfigHandler {
	return &ConfigHandler{
		publisher: pub,
	}
}

// UpdateDeviceConfig handles POST /api/v1/devices/{mac}/config
// It parses the desired config from the UI and fires an MQTT RPC command to the edge device.
func (h *ConfigHandler) UpdateDeviceConfig(w http.ResponseWriter, r *http.Request) {
	mac := r.PathValue("mac")
	if !isValidMAC(mac) {
		response.Error(w, http.StatusBadRequest, "Invalid or missing MAC address")
		return
	}

	var config DeviceConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	// Marshal the config back to JSON to send to the ESP32
	// (Or we could use Protobuf here if we want maximum efficiency)
	payload, err := json.Marshal(config)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to marshal config")
		return
	}

	// Fire the MQTT RPC Command
	if h.publisher != nil {
		topic := fmt.Sprintf("/xomoi/%s/rpc/config", mac)
		
		// Retain = true! This is crucial. If the ESP32 is offline or sleeping,
		// the broker will hold this config. The exact millisecond the ESP32 boots up
		// and subscribes to this topic, the broker will blast the new config down to it.
		err := h.publisher.Publish(topic, payload, true, 1) // QoS 1 ensures delivery
		if err != nil {
			slog.Error("Failed to publish Config RPC", "mac", mac, "error", err)
			response.Error(w, http.StatusInternalServerError, "Failed to send command to device")
			return
		}
		
		slog.Info("Config RPC sent successfully", "mac", mac, "topic", topic, "config", string(payload))
	} else {
		slog.Warn("Config RPC requested but MQTTPublisher is nil (Headless Mode)", "mac", mac)
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"status": "success",
		"message": "Configuration update sent to device",
	})
}
