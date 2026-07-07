package api

import (
	"net/http"

	"github.com/code-grey/xomoi-core/internal/api/handlers"
	"github.com/code-grey/xomoi-core/internal/api/middleware"
	"github.com/code-grey/xomoi-core/internal/repository"
)

// Server holds the dependencies for the HTTP server.
type Server struct {
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository
}

// NewServer creates a new API Server instance.
func NewServer(uRepo repository.UserRepository, sRepo repository.SessionRepository) *Server {
	return &Server{
		userRepo:    uRepo,
		sessionRepo: sRepo,
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

	// Real-time WebSockets
	mux.HandleFunc("GET /api/v1/ws/health", handlers.HealthWebSocket)
	mux.HandleFunc("GET /api/v1/ws/telemetry", handlers.TelemetryWebSocket)

	// Apply global Panic Recovery middleware to ensure the broker never crashes from an API panic.
	return middleware.PanicRecovery(mux)
}
