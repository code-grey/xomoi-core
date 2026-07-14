package network

import (
	"log/slog"
	"strconv"

	"github.com/grandcat/zeroconf"
)

// MDNSServer manages the local network Zero-Config broadcast
type MDNSServer struct {
	server *zeroconf.Server
}

// StartMDNS initializes the mDNS broadcast.
// It tells all devices on the local network that xomoi.local points to this IP.
func StartMDNS(portStr string) (*MDNSServer, error) {
	slog.Info("Starting mDNS Zero-Config Broadcaster...")

	port, err := strconv.Atoi(portStr)
	if err != nil {
		slog.Error("Invalid port for mDNS", "error", err)
		return nil, err
	}

	// Register a service pointing to xomoi.local
	// "_http._tcp" is the standard service type for web servers.
	// "local." is the standard mDNS domain.
	server, err := zeroconf.Register("xomoi", "_http._tcp", "local.", port, []string{"txtv=1", "app=xomoi"}, nil)
	if err != nil {
		slog.Error("Failed to initialize mDNS", "error", err)
		return nil, err
	}

	slog.Info("mDNS Broadcast active. You can now access the dashboard at http://xomoi.local:" + portStr)
	return &MDNSServer{server: server}, nil
}

// Stop cleanly unregisters the service from the network
func (m *MDNSServer) Stop() {
	if m.server != nil {
		slog.Info("Shutting down mDNS broadcaster...")
		m.server.Shutdown()
	}
}
