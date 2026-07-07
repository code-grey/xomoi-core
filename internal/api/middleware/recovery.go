package middleware

import (
	"log"
	"net/http"
)

// PanicRecovery is a critical global middleware that prevents the entire Xomoi-Core 
// edge node from crashing if a single HTTP handler panics. 
// Because the broker and API share the same binary, this is mandatory.
func PanicRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[PANIC RECOVERED] API handler panicked: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
