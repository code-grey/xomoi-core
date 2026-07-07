package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/code-grey/xomoi-core/internal/core"
	"github.com/code-grey/xomoi-core/internal/repository"
)

type ClaimHandler struct {
	deviceRepo repository.DeviceRepository
}

func NewClaimHandler(dRepo repository.DeviceRepository) *ClaimHandler {
	return &ClaimHandler{deviceRepo: dRepo}
}

// Discover scans local network interfaces or mDNS for "Xomoi-Claim-XXXX" beacons.
func (h *ClaimHandler) Discover(w http.ResponseWriter, r *http.Request) {
	// Skeleton simulation of finding an unprovisioned ESP32 in AP mode
	devices := []map[string]string{
		{"ssid": "Xomoi-Claim-A1B2", "mac": "00:1A:2B:3C:4D:5E"},
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(devices)
}

type ClaimRequest struct {
	MACAddress string `json:"mac_address"`
	DeviceName string `json:"device_name"`
}

// Claim executes the zero-friction HMAC-Lite binding process.
func (h *ClaimHandler) Claim(w http.ResponseWriter, r *http.Request) {
	var req ClaimRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 1. Generate an unbreakable 32-byte HMAC Secret for this specific device
	secretBytes := make([]byte, 32)
	rand.Read(secretBytes)
	secretKey := hex.EncodeToString(secretBytes)

	// 2. Register Device in SQLite
	device := &core.Device{
		ID:         "dev_" + hex.EncodeToString(secretBytes[:4]),
		Name:       req.DeviceName,
		MACAddress: req.MACAddress,
		SecretKey:  secretKey,
	}
	
	// h.deviceRepo.Create(r.Context(), device)
	_ = device

	// 3. (Simulated) Connect to the Device's Captive AP, push the Wi-Fi credentials
	// and the HMAC-Lite secret token, then command it to reboot into Station mode.

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
		"message": "Device successfully claimed and bound via HMAC-Lite.",
	})
}
