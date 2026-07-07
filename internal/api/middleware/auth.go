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
