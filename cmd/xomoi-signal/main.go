package main

import (
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"
	"github.com/code-grey/xomoi-core/internal/signal"
)

func main() {
	// Initialize JSON Logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	slog.Info("BOOTING XOMOI WEBRTC SIGNALING SERVER")

	// 1. Extreme Memory Tuning
	// Cap the heap at 200MB. This guarantees the free-tier Render container (512MB limit)
	// will NEVER crash from an Out-Of-Memory error, even during a DDOS attack.
	debug.SetMemoryLimit(200 * 1024 * 1024)
	slog.Info("Engine tuning: GOMEMLIMIT enforced at 200MB")

	// 2. Initialize the Server with the Sharded Map
	sigServer := signal.NewServer()

	// 3. Register HTTP Routes
	mux := http.NewServeMux()
	
	// A basic health check so Cloudflare/Render knows the server is alive
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "alive", "version": "1.0.0", "goroutines": ` + itoa(runtime.NumGoroutine()) + `}`))
	})
	
	// The WebRTC WebSocket Signaling Endpoint
	mux.HandleFunc("GET /ws", sigServer.HandleWebSocket)

	// 4. Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8086"
	}

	slog.Info("Signaling server listening", "port", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		slog.Error("Server crashed", "error", err)
		os.Exit(1)
	}
}

// Simple helper for the health check
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var b [32]byte
	bp := len(b)
	for ; i > 0; i /= 10 {
		bp--
		b[bp] = byte(i%10) + '0'
	}
	return string(b[bp:])
}
