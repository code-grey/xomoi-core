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

package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/code-grey/xomoi-core/internal/repository"
)

type contextKey string

const UserIDKey contextKey = "user_id"

// SessionCheck verifies the bearer token against the SQLite database.
func SessionCheck(sessionRepo repository.SessionRepository, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Unauthorized: Missing Token", http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		
		// In a real implementation, we hash the token here before checking DB.
		// For the skeleton, we query the sessionRepo directly.
		session, err := sessionRepo.GetByID(r.Context(), token)
		if err != nil || session == nil {
			http.Error(w, "Unauthorized: Invalid or Expired Token", http.StatusUnauthorized)
			return
		}

		// Attach user ID to context for the downstream handler to use
		ctx := context.WithValue(r.Context(), UserIDKey, session.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
