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
