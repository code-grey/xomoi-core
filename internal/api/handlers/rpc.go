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

// RPCPayload represents an arbitrary command sent from the Svelte UI to the edge device
type RPCPayload struct {
	Command string                 `json:"command"`
	Params  map[string]interface{} `json:"params,omitempty"`
	Retain  bool                   `json:"retain"` // If true, broker saves it for offline devices
}

type RPCHandler struct {
	publisher MQTTPublisher
}

func NewRPCHandler(pub MQTTPublisher) *RPCHandler {
	return &RPCHandler{
		publisher: pub,
	}
}

// ExecuteCommand handles POST /api/v1/devices/{mac}/rpc
// It takes a generic command from the Svelte UI and blasts it to the ESP32.
func (h *RPCHandler) ExecuteCommand(w http.ResponseWriter, r *http.Request) {
	mac := r.PathValue("mac")
	if !isValidMAC(mac) {
		response.Error(w, http.StatusBadRequest, "Invalid or missing MAC address")
		return
	}

	var payload RPCPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	// Marshal back to bytes for MQTT
	msgBytes, err := json.Marshal(payload)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to marshal payload")
		return
	}

	if h.publisher != nil {
		// We use a specific sub-topic for generic commands
		topic := fmt.Sprintf("/xomoi/%s/rpc/command", mac)
		
		err := h.publisher.Publish(topic, msgBytes, payload.Retain, 1) // QoS 1, Dynamic Retain
		if err != nil {
			slog.Error("Failed to publish Generic RPC", "mac", mac, "error", err)
			response.Error(w, http.StatusInternalServerError, "Failed to send command to device")
			return
		}
		
		slog.Info("Generic RPC command sent", "mac", mac, "command", payload.Command)
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"status": "success",
		"message": "Command sent to device",
	})
}
