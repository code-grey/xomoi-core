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

package broker

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"sync"

	"github.com/code-grey/xomoi-core/internal/config"
	"github.com/code-grey/xomoi-core/internal/core"
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

// sessionCache stores the validated HMAC string for each MAC address in RAM.
// This completely bypasses SQLite Disk I/O and SHA256 computations on reconnects.
var sessionCache sync.Map

func (h *HMACAuthHook) OnConnectAuthenticate(cl *mqtt.Client, pk packets.Packet) bool {
	macAddress := string(pk.Connect.Username)
	providedToken := string(pk.Connect.Password)

	// 1. Lightning-Fast Session Cache Check (0 Allocations, RAM only)
	if cachedToken, ok := sessionCache.Load(macAddress); ok {
		if cachedToken.(string) == providedToken {
			return true
		}
	}

	var secretKey string

	device, err := h.deviceRepo.GetByMAC(context.Background(), macAddress)
	if err != nil {
		// Device not found in SQLite! 
		// Auto-provision as Unclaimed if it authenticates with the Factory Secret
		factorySecret := config.Load().FactorySecret
		
		macHmac := hmac.New(sha256.New, []byte(factorySecret))
		macHmac.Write([]byte(macAddress))
		expectedFactoryHash := hex.EncodeToString(macHmac.Sum(nil))

		if providedToken != expectedFactoryHash {
			log.Printf("AUTH REJECTED: Unknown device with invalid factory signature. MAC: %s", macAddress)
			return false
		}

		// Valid factory signature. Auto-provision in SQLite
		newDevice := &core.Device{
			ID:         macAddress, // Use MAC as ID for now
			Name:       "Unclaimed Node",
			MACAddress: macAddress,
			SecretKey:  factorySecret, // Must be rotated during "Claiming" (Phase 3)
		}
		
		if createErr := h.deviceRepo.Create(context.Background(), newDevice); createErr != nil {
			log.Printf("Failed to auto-provision device: %v", createErr)
			return false
		}
		
		log.Printf("AUTH SUCCESS: Auto-provisioned new device %s", macAddress)
		sessionCache.Store(macAddress, providedToken)
		return true
	}

	secretKey = device.SecretKey

	macHmac := hmac.New(sha256.New, []byte(secretKey))
	macHmac.Write([]byte(macAddress))
	expectedHash := hex.EncodeToString(macHmac.Sum(nil))

	if providedToken != expectedHash {
		log.Printf("AUTH REJECTED: Invalid signature for claimed device. MAC: %s", macAddress)
		return false
	}

	sessionCache.Store(macAddress, providedToken)
	return true
}

// OnACLCheck controls publish/subscribe permissions.
func (h *HMACAuthHook) OnACLCheck(cl *mqtt.Client, topic string, write bool) bool {
	username := string(cl.Properties.Username)
	if username == "" {
		return false
	}

	// Strict ACL: Device can only publish/sub to its own namespace
	allowedPrefix := "/xomoi/" + username + "/"
	if len(topic) >= len(allowedPrefix) && topic[:len(allowedPrefix)] == allowedPrefix {
		return true
	}
	
	log.Printf("ACL REJECTED: Client %s attempted to access restricted topic %s", username, topic)
	return false
}
