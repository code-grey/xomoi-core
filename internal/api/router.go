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

package api

import (
	"net/http"

	"github.com/code-grey/xomoi-core/internal/api/handlers"
	"github.com/code-grey/xomoi-core/internal/api/middleware"
	"github.com/code-grey/xomoi-core/internal/repository"
	"github.com/code-grey/xomoi-core/internal/worker"
	mqtt "github.com/mochi-mqtt/server/v2"
)

// Server holds the dependencies for the HTTP server.
type Server struct {
	userRepo         repository.UserRepository
	sessionRepo      repository.SessionRepository
	otaHandler       *handlers.OTAHandler
	configHandler    *handlers.ConfigHandler
	rpcHandler       *handlers.RPCHandler
	claimHandler     *handlers.ClaimHandler
	telemetryHandler *handlers.TelemetryHandler
	rulesHandler     *handlers.RulesHandler
	broker           *mqtt.Server
}

// NewServer creates a new API Server instance.
func NewServer(uRepo repository.UserRepository, sRepo repository.SessionRepository, dRepo repository.DeviceRepository, tsdb repository.TelemetryRepository, ruleRepo repository.AlertRuleRepository, broker *mqtt.Server, pub handlers.MQTTPublisher, rulesEngine *worker.RulesEngine) *Server {
	return &Server{
		userRepo:         uRepo,
		sessionRepo:      sRepo,
		otaHandler:       handlers.NewOTAHandler(pub, "data/ota"),
		configHandler:    handlers.NewConfigHandler(pub),
		rpcHandler:       handlers.NewRPCHandler(pub),
		claimHandler:     handlers.NewClaimHandler(dRepo),
		telemetryHandler: handlers.NewTelemetryHandler(tsdb),
		rulesHandler:     handlers.NewRulesHandler(ruleRepo, rulesEngine),
		broker:           broker,
	}
}

// SetupRouter configures the Go 1.26 ServeMux with all endpoints.
func (s *Server) SetupRouter() http.Handler {
	mux := http.NewServeMux()

	authHandler := handlers.NewAuthHandler(s.userRepo, s.sessionRepo)

	// Public Endpoints
	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)

	// Protected Endpoints
	mux.Handle("POST /api/v1/auth/logout", middleware.SessionCheck(s.sessionRepo, http.HandlerFunc(authHandler.Logout)))

	// Device Management Endpoints
	mux.HandleFunc("GET /api/v1/devices", s.claimHandler.List)
	mux.HandleFunc("POST /api/v1/devices/claim", s.claimHandler.Claim)

	// Real-time WebSockets
	mux.HandleFunc("GET /api/v1/ws/health", handlers.HealthWebSocket)
	mux.HandleFunc("GET /api/v1/ws/telemetry", handlers.TelemetryWebSocket)
	
	// MQTT-over-WSS Multiplexer (for Render / single-port deployments)
	mux.HandleFunc("GET /mqtt", handlers.MQTTWebSocket(s.broker))

	// OTA (Over-The-Air) Firmware Endpoints
	mux.Handle("POST /api/v1/devices/{mac}/ota", middleware.SessionCheck(s.sessionRepo, http.HandlerFunc(s.otaHandler.UploadFirmware)))
	// The download endpoint is public so the hardware device can pull it without session cookies
	mux.HandleFunc("GET /api/v1/devices/{mac}/ota/download", s.otaHandler.DownloadFirmware)

	// Dynamic NVS Config Endpoints
	mux.Handle("POST /api/v1/devices/{mac}/config", middleware.SessionCheck(s.sessionRepo, http.HandlerFunc(s.configHandler.UpdateDeviceConfig)))

	// TSDB History
	mux.HandleFunc("GET /api/v1/devices/{mac}/history", s.telemetryHandler.GetHistory)

	// Generic RPC Actuation Endpoints
	mux.Handle("POST /api/v1/devices/{mac}/rpc", middleware.SessionCheck(s.sessionRepo, http.HandlerFunc(s.rpcHandler.ExecuteCommand)))

	// Apply global middlewares: CORS and Panic Recovery
	return middleware.CORS(middleware.PanicRecovery(mux))
}
