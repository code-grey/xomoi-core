package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"

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

	// 1. Check if device exists and is Unclaimed
	device, err := h.deviceRepo.GetByMAC(r.Context(), req.MACAddress)
	if err != nil {
		http.Error(w, "Device not found. Please ensure it is powered on and connected to the Dark Grid once.", http.StatusNotFound)
		return
	}

	if device.SecretKey != "xomoi-factory-secret" {
		http.Error(w, "Device is already claimed.", http.StatusConflict)
		return
	}

	// 2. Generate an unbreakable 32-byte HMAC Secret for this specific device
	secretBytes := make([]byte, 32)
	rand.Read(secretBytes)
	newSecretKey := hex.EncodeToString(secretBytes)

	// 3. Update Device in SQLite
	if err := h.deviceRepo.ClaimDevice(r.Context(), req.MACAddress, req.DeviceName, newSecretKey); err != nil {
		http.Error(w, "Failed to claim device", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
		"message": "Device successfully claimed.",
		"private_key": newSecretKey,
	})
}

// List returns all registered devices.
func (h *ClaimHandler) List(w http.ResponseWriter, r *http.Request) {
	devices, err := h.deviceRepo.GetAll(r.Context())
	if err != nil {
		http.Error(w, "Failed to fetch devices", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(devices)
}
