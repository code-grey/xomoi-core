package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/code-grey/xomoi-core/internal/repository"
)

type AuthHandler struct {
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository
}

func NewAuthHandler(uRepo repository.UserRepository, sRepo repository.SessionRepository) *AuthHandler {
	return &AuthHandler{
		userRepo:    uRepo,
		sessionRepo: sRepo,
	}
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Login handles the Argon2ID verification and issues a Session Token.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Phase 4.2 Skeleton:
	// 1. Fetch user by username from SQLite
	// 2. Verify password with Argon2ID
	// 3. Generate a secure random session token
	// 4. Save Session to SQLite
	
	// Simulated response for the skeleton
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"token": "simulated_session_token",
	})
}

// Logout deletes the active session from SQLite.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Extract the token from header (or context via middleware) and delete it from SQLite
	w.WriteHeader(http.StatusOK)
}
