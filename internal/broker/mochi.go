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

package broker

import (
	"log"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/listeners"
)

// NewMochiServer initializes an embedded Mochi-MQTT broker.
func NewMochiServer(port string) (*mqtt.Server, error) {
	server := mqtt.New(&mqtt.Options{
		InlineClient: true,
	})

	// Add standard TCP listener for standard ESP32 connections
	tcp := listeners.NewTCP(listeners.Config{
		ID:      "t1",
		Address: ":" + port,
	})
	if err := server.AddListener(tcp); err != nil {
		log.Fatalf("Failed to add TCP listener: %v", err)
	}

	// Setup WebSockets Listener (For Flutter UI real-time telemetry)
	ws := listeners.NewWebsocket(listeners.Config{
		ID:      "ws1",
		Address: ":1884",
	})
	if err := server.AddListener(ws); err != nil {
		log.Fatalf("Failed to add WS listener: %v", err)
	}

	return server, nil
}
