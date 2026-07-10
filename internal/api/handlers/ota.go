package handlers

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// isValidMAC provides ultra-fast, zero-allocation MAC address validation
func isValidMAC(mac string) bool {
	if len(mac) != 17 {
		return false
	}
	for i := 0; i < len(mac); i++ {
		c := mac[i]
		if i%3 == 2 {
			if c != ':' && c != '_' {
				return false
			}
		} else {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
				return false
			}
		}
	}
	return true
}

// MQTTPublisher defines the interface to interact with the embedded Mochi-MQTT broker
// or any external broker for the OTA RPC command.
type MQTTPublisher interface {
	Publish(topic string, payload []byte, retain bool, qos byte) error
}

type OTAHandler struct {
	publisher MQTTPublisher
	otaDir    string
}

func NewOTAHandler(pub MQTTPublisher, otaDir string) *OTAHandler {
	// Ensure the OTA binary storage directory exists
	if err := os.MkdirAll(otaDir, 0755); err != nil {
		slog.Error("Failed to create OTA directory", "dir", otaDir, "error", err)
	}
	return &OTAHandler{
		publisher: pub,
		otaDir:    otaDir,
	}
}

// UploadFirmware handles POST /api/v1/devices/{mac}/ota
// It saves the uploaded .bin file and fires an MQTT RPC command to the device.
func (h *OTAHandler) UploadFirmware(w http.ResponseWriter, r *http.Request) {
	// Simple path param extraction (Assuming Go 1.22+ ServeMux routing: /api/v1/devices/{mac}/ota)
	mac := r.PathValue("mac")
	if !isValidMAC(mac) {
		http.Error(w, "Invalid or missing MAC address", http.StatusBadRequest)
		return
	}

	// Limit upload size to 2MB (Standard for ESP32 OTA partitions)
	r.Body = http.MaxBytesReader(w, r.Body, 2<<20)
	if err := r.ParseMultipartForm(2 << 20); err != nil {
		http.Error(w, "File too large or invalid form", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("firmware")
	if err != nil {
		http.Error(w, "Missing 'firmware' file field", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Clean the MAC to prevent directory traversal
	safeMac := strings.ReplaceAll(mac, ":", "_")
	safeMac = filepath.Clean(safeMac)
	
	destPath := filepath.Join(h.otaDir, fmt.Sprintf("%s.bin", safeMac))
	destFile, err := os.Create(destPath)
	if err != nil {
		slog.Error("Failed to create OTA file", "path", destPath, "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, file); err != nil {
		slog.Error("Failed to write OTA file", "path", destPath, "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	slog.Info("OTA Firmware uploaded successfully", "mac", mac, "path", destPath)

	// Fire the MQTT RPC Command to trigger the ESP32 to download it
	if h.publisher != nil {
		topic := fmt.Sprintf("/xomoi/%s/rpc", mac)
		// The payload commands the SDK to download from this exact server
		// We pass a relative path; the SDK will prepend the Broker's IP automatically
		payload := []byte(fmt.Sprintf("OTA:/api/v1/devices/%s/ota/download", mac))
		
		err := h.publisher.Publish(topic, payload, false, 1)
		if err != nil {
			slog.Error("Failed to publish OTA trigger", "mac", mac, "error", err)
		} else {
			slog.Info("OTA Trigger sent via MQTT RPC", "topic", topic)
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success", "message":"Firmware uploaded and OTA triggered"}`))
}

// DownloadFirmware handles GET /api/v1/devices/{mac}/ota/download
// The ESP32 calls this endpoint after receiving the MQTT RPC trigger.
func (h *OTAHandler) DownloadFirmware(w http.ResponseWriter, r *http.Request) {
	mac := r.PathValue("mac")
	if !isValidMAC(mac) {
		http.Error(w, "Invalid or missing MAC address", http.StatusBadRequest)
		return
	}

	safeMac := strings.ReplaceAll(mac, ":", "_")
	safeMac = filepath.Clean(safeMac)
	filePath := filepath.Join(h.otaDir, fmt.Sprintf("%s.bin", safeMac))

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "Firmware not found", http.StatusNotFound)
		return
	}

	slog.Info("Device pulling OTA firmware", "mac", mac)
	http.ServeFile(w, r, filePath)
}
