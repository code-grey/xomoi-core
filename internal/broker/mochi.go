package broker

import (
	"log"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/listeners"
)

// Broker represents the embedded MQTT broker.
type Broker struct {
	Server *mqtt.Server
}

// NewBroker initializes the Mochi-MQTT embedded server.
func NewBroker(authHook mqtt.Hook, publishHook mqtt.Hook) *Broker {
	server := mqtt.New(nil)

	// Add our custom HMAC-Lite Auth Hook
	if err := server.AddHook(authHook, nil); err != nil {
		log.Fatalf("Failed to add auth hook to broker: %v", err)
	}

	// Add our Ingestion Publish Hook
	if err := server.AddHook(publishHook, nil); err != nil {
		log.Fatalf("Failed to add publish hook to broker: %v", err)
	}

	// Setup TCP Listener (Native MQTT for IoT Sensors)
	tcp := listeners.NewTCP(listeners.Config{
		ID:      "tcp1",
		Address: ":1883",
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

	return &Broker{Server: server}
}

// Start runs the broker in the background.
func (b *Broker) Start() error {
	log.Println("Embedded Mochi-MQTT broker starting on TCP :1883 and WS :1884")
	return b.Server.Serve()
}
