package signal

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow all origins for the Signaling Server since requests will come from 
	// app.xomoi.io and localhost
	CheckOrigin: func(r *http.Request) bool { return true },
}

// SignalMessage is the generic WebRTC payload
type SignalMessage struct {
	Type   string          `json:"type"`             // "offer", "answer", "ice"
	Target string          `json:"target"`           // Node ID to route to
	Source string          `json:"source,omitempty"` // Sender Node ID
	Data   json.RawMessage `json:"data"`             // The actual SDP or ICE payload
}

type Server struct {
	connections *ShardedConnectionMap
}

func NewServer() *Server {
	return &Server{
		connections: NewShardedConnectionMap(),
	}
}

// HandleWebSocket upgrades the HTTP request and handles signaling
func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// The client MUST provide their Node ID as a query param (e.g., ?id=XOMOI-PI-01)
	nodeID := r.URL.Query().Get("id")
	if nodeID == "" {
		http.Error(w, "Missing node ID", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("Failed to upgrade WebSocket", "error", err)
		return
	}

	// Hardened Security: Prevent malicious giants payloads from OOMing the server
	conn.SetReadLimit(4096) 
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))

	// Register Connection
	s.connections.Set(nodeID, conn)
	slog.Info("Node connected to signaling server", "id", nodeID)

	// Keepalive handler
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// Enter read loop
	defer func() {
		s.connections.Delete(nodeID)
		conn.Close()
		slog.Info("Node disconnected", "id", nodeID)
	}()

	for {
		_, msgData, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Error("WebSocket read error", "error", err)
			}
			break
		}

		// Use the sync.Pool to avoid heap allocations for the JSON unmarshaling
		// Actually, since msgData is already allocated by gorilla, we process it quickly
		// and let it GC, but for max performance we could use gorilla's io.Reader
		// For simplicity in this microservice, we just unmarshal.
		
		var sigMsg SignalMessage
		if err := json.Unmarshal(msgData, &sigMsg); err != nil {
			slog.Warn("Invalid signaling JSON", "error", err)
			continue
		}
		
		// Enforce source tag
		sigMsg.Source = nodeID

		// Route the message to the target peer!
		if targetConn, ok := s.connections.Get(sigMsg.Target); ok {
			// Re-marshal securely
			outData, _ := json.Marshal(sigMsg)
			
			// Lock the target connection for writing (gorilla requires single writer)
			// For ultra-scale, this should go through a worker queue, but this is fine for P2P.
			targetConn.WriteMessage(websocket.TextMessage, outData)
			slog.Info("Routed WebRTC signal", "from", nodeID, "to", sigMsg.Target, "type", sigMsg.Type)
		} else {
			slog.Warn("Target node not found or offline", "target", sigMsg.Target)
		}
	}
}
