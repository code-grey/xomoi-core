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
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	mqtt "github.com/mochi-mqtt/server/v2"
)

var mqttUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // Allow all for Edge Node
	Subprotocols: []string{"mqtt", "mqttv3.1"}, // Standard MQTT-over-WSS subprotocols
}

// wsConn adapts a gorilla/websocket to a standard net.Conn so Mochi-MQTT can use it.
type wsConn struct {
	*websocket.Conn
	r io.Reader
}

func (c *wsConn) Read(b []byte) (int, error) {
	for {
		if c.r == nil {
			var err error
			_, c.r, err = c.Conn.NextReader()
			if err != nil {
				return 0, err
			}
		}
		n, err := c.r.Read(b)
		if err == io.EOF {
			c.r = nil
			if n > 0 {
				return n, nil
			}
			continue // Read next message if this one was empty
		}
		return n, err
	}
}

func (c *wsConn) Write(b []byte) (int, error) {
	err := c.Conn.WriteMessage(websocket.BinaryMessage, b)
	if err != nil {
		return 0, err
	}
	return len(b), nil
}

func (c *wsConn) SetDeadline(t time.Time) error {
	if err := c.Conn.SetReadDeadline(t); err != nil {
		return err
	}
	return c.Conn.SetWriteDeadline(t)
}

// MQTTWebSocket multiplexes raw MQTT traffic over the same HTTP Port using WebSockets
func MQTTWebSocket(broker *mqtt.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ws, err := mqttUpgrader.Upgrade(w, r, nil)
		if err != nil {
			slog.Error("Failed to upgrade to MQTT over WSS", "error", err)
			return
		}

		slog.Info("New MQTT over WSS connection established", "ip", r.RemoteAddr)

		conn := &wsConn{Conn: ws}
		
		// Hand the raw upgraded WebSocket connection off to the Mochi-MQTT core!
		err = broker.EstablishConnection("WSS-MULTIPLEXER", conn)
		if err != nil {
			slog.Warn("Mochi-MQTT dropped WSS connection", "error", err)
		}
	}
}
