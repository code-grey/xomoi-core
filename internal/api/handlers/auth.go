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
	"net/http"

	"github.com/code-grey/xomoi-core/internal/api/response"
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
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Phase 4.2 Skeleton:
	// 1. Fetch user by username from SQLite
	// 2. Verify password with Argon2ID
	// 3. Generate a secure random session token
	// 4. Save Session to SQLite
	
	// Simulated response for the skeleton
	response.JSON(w, http.StatusOK, map[string]string{
		"token": "simulated_session_token",
	})
}

// Logout deletes the active session from SQLite.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Extract the token from header (or context via middleware) and delete it from SQLite
	w.WriteHeader(http.StatusOK)
}
