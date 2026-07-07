package broker

import (
	"bytes"
	"log"

	"github.com/code-grey/xomoi-core/internal/repository"
	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/packets"
)

// HMACAuthHook implements the Mochi-MQTT Auth Hook for HMAC-Lite authentication.
type HMACAuthHook struct {
	mqtt.HookBase
	deviceRepo repository.DeviceRepository
}

// NewHMACAuthHook creates a new hook using the provided device repository.
func NewHMACAuthHook(repo repository.DeviceRepository) *HMACAuthHook {
	return &HMACAuthHook{deviceRepo: repo}
}

// ID returns the hook identifier.
func (h *HMACAuthHook) ID() string {
	return "xomoi-hmac-auth"
}

// Provides indicates which hook methods this hook implements.
func (h *HMACAuthHook) Provides(b byte) bool {
	return bytes.Contains([]byte{
		mqtt.OnConnectAuthenticate,
		mqtt.OnACLCheck,
	}, []byte{b})
}

// OnConnectAuthenticate verifies the HMAC-Lite token provided as the password.
func (h *HMACAuthHook) OnConnectAuthenticate(cl *mqtt.Client, pk packets.Packet) bool {
	macAddress := string(pk.Connect.Username)
	providedToken := string(pk.Connect.Password)

	log.Printf("Authenticating device MAC: %s", macAddress)
	
	// Phase 3.1: To be verified against device.SecretKey retrieved via deviceRepo.
	// The client generates an HMAC-SHA256 signature using their SecretKey.
	// We will re-compute it here and compare.
	_ = providedToken 

	// Default allow for skeleton structure. Hardening logic will lock this down.
	return true 
}

// OnACLCheck controls publish/subscribe permissions.
func (h *HMACAuthHook) OnACLCheck(cl *mqtt.Client, topic string, write bool) bool {
	// Devices can only publish/subscribe to their own topics: e.g., "telemetry/{mac}"
	// UI/Admin clients can subscribe to "#"
	
	// Default allow for testing the skeleton. Hardening comes in Phase 9.
	return true
}
