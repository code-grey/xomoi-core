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
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/code-grey/xomoi-core/internal/api/response"
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
	
	response.JSON(w, http.StatusOK, devices)
}

type ClaimRequest struct {
	MACAddress string `json:"mac_address"`
	DeviceName string `json:"device_name"`
}

// Claim executes the zero-friction HMAC-Lite binding process.
func (h *ClaimHandler) Claim(w http.ResponseWriter, r *http.Request) {
	var req ClaimRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// 1. Check if device exists and is Unclaimed
	device, err := h.deviceRepo.GetByMAC(r.Context(), req.MACAddress)
	if err != nil {
		response.Error(w, http.StatusNotFound, "Device not found. Please ensure it is powered on and connected to the Dark Grid once.")
		return
	}

	if device.SecretKey != "xomoi-factory-secret" {
		response.Error(w, http.StatusConflict, "Device is already claimed.")
		return
	}

	// 2. Generate an unbreakable 32-byte HMAC Secret for this specific device
	secretBytes := make([]byte, 32)
	rand.Read(secretBytes)
	newSecretKey := hex.EncodeToString(secretBytes)

	// 3. Update Device in SQLite
	if err := h.deviceRepo.ClaimDevice(r.Context(), req.MACAddress, req.DeviceName, newSecretKey); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to claim device")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"status": "success",
		"message": "Device successfully claimed.",
		"private_key": newSecretKey,
	})
}

// List returns all registered devices.
func (h *ClaimHandler) List(w http.ResponseWriter, r *http.Request) {
	devices, err := h.deviceRepo.GetAll(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to fetch devices")
		return
	}

	response.JSON(w, http.StatusOK, devices)
}
