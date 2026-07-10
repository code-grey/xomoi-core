package api

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/gorilla/websocket"
	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/packets"
	"github.com/pion/webrtc/v4"
)

// SignalMessage matches the structure expected by the xomoi-signal server
type SignalMessage struct {
	Type   string          `json:"type"`
	Target string          `json:"target"`
	Source string          `json:"source,omitempty"`
	Data   json.RawMessage `json:"data"`
}

// WebRTCHost manages the Pi-side NAT hole-punching
type WebRTCHost struct {
	nodeID       string
	signalingURL string
	ws           *websocket.Conn
	api          *webrtc.API
	broker       *mqtt.Server
}

// NewWebRTCHost creates a new WebRTC host instance
func NewWebRTCHost(nodeID, signalingURL string, broker *mqtt.Server) *WebRTCHost {
	return &WebRTCHost{
		nodeID:       nodeID,
		signalingURL: signalingURL,
		api:          webrtc.NewAPI(),
		broker:       broker,
	}
}

// Start connects to the signaling server and listens for incoming connections
func (h *WebRTCHost) Start() {
	url := h.signalingURL + "?id=" + h.nodeID
	slog.Info("WebRTC Host connecting to signaling server", "url", url)

	var err error
	for {
		h.ws, _, err = websocket.DefaultDialer.Dial(url, nil)
		if err == nil {
			break
		}
		slog.Warn("Signaling server offline. Retrying in 5s...", "error", err)
		time.Sleep(5 * time.Second)
	}

	slog.Info("WebRTC Signaling Connected. Awaiting connections...")

	// Listen for incoming offers
	go h.listenLoop()
}

func (h *WebRTCHost) listenLoop() {
	// A single map to hold the active peer connection for now. 
	// In production, we'd manage multiple concurrent peer connections in a map.
	var pc *webrtc.PeerConnection

	for {
		_, msgBytes, err := h.ws.ReadMessage()
		if err != nil {
			slog.Error("Signaling WebSocket disconnected", "error", err)
			return // Ideally trigger a reconnect loop here
		}

		var msg SignalMessage
		if err := json.Unmarshal(msgBytes, &msg); err != nil {
			continue
		}

		switch msg.Type {
		case "offer":
			var offer webrtc.SessionDescription
			json.Unmarshal(msg.Data, &offer)

			slog.Info("Received WebRTC Offer", "from", msg.Source)
			pc = h.handleOffer(offer, msg.Source)

		case "ice":
			if pc != nil {
				var candidate webrtc.ICECandidateInit
				json.Unmarshal(msg.Data, &candidate)
				pc.AddICECandidate(candidate)
			}
		}
	}
}

func (h *WebRTCHost) handleOffer(offer webrtc.SessionDescription, clientID string) *webrtc.PeerConnection {
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{URLs: []string{"stun:stun.l.google.com:19302"}},
		},
	}

	pc, err := h.api.NewPeerConnection(config)
	if err != nil {
		slog.Error("Failed to create PeerConnection", "error", err)
		return nil
	}

	// 1. ICE Candidate Handler: Send our ICE candidates back to the client
	pc.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			return
		}
		
		candidateJSON, _ := json.Marshal(c.ToJSON())
		h.ws.WriteJSON(SignalMessage{
			Type:   "ice",
			Target: clientID,
			Data:   candidateJSON,
		})
	})

	// 2. Data Channel Handler: When the Svelte UI opens the encrypted tunnel
	pc.OnDataChannel(func(d *webrtc.DataChannel) {
		slog.Info("E2E Encrypted DataChannel opened!", "label", d.Label())

		// Subscribe the WebRTC Tunnel directly to the embedded MQTT Broker
		h.broker.Subscribe("/xomoi/+/telemetry", 1, func(cl *mqtt.Client, sub packets.Subscription, pk packets.Packet) {
			// Real-time: As soon as an ESP32 fires telemetry, stream it P2P to the Web UI
			if d.ReadyState() == webrtc.DataChannelStateOpen {
				d.SendText(string(pk.Payload))
			}
		})

		d.OnMessage(func(msg webrtc.DataChannelMessage) {
			slog.Info("Received P2P Command from Dashboard", "data", string(msg.Data))
			// TODO: Forward WebRTC commands into Mochi-MQTT to trigger OTA or relay toggles
		})
	})

	// 3. Connection State Logger
	pc.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		slog.Info("WebRTC Connection State", "state", s.String())
	})

	// 4. Accept the offer and create an Answer
	if err := pc.SetRemoteDescription(offer); err != nil {
		slog.Error("Failed to set remote description", "error", err)
		return nil
	}

	answer, err := pc.CreateAnswer(nil)
	if err != nil {
		slog.Error("Failed to create answer", "error", err)
		return nil
	}

	if err := pc.SetLocalDescription(answer); err != nil {
		slog.Error("Failed to set local description", "error", err)
		return nil
	}

	// 5. Send Answer back through the signaling server
	answerJSON, _ := json.Marshal(answer)
	h.ws.WriteJSON(SignalMessage{
		Type:   "answer",
		Target: clientID,
		Data:   answerJSON,
	})

	return pc
}
