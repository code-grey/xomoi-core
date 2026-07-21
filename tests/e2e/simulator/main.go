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

package main

import (
	"bufio"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func main() {
	// Seed random generator for Virtual MACs
	rand.Seed(time.Now().UnixNano())

	fmt.Println("=========================================")
	fmt.Println("    XOMOI UNIVERSAL EDGE SIMULATOR       ")
	fmt.Println("=========================================")

	reader := bufio.NewReader(os.Stdin)

	// 1. Interactive Device Selection
	fmt.Println("Select Microcontroller Profile:")
	fmt.Println("  1. ESP32 (DHT11 Env Sensor)")
	fmt.Println("  2. Raspberry Pi Pico W (PIR Motion Sensor)")
	fmt.Println("  3. ESP8266 (Relay Switch)")
	fmt.Print("Choice [1/2/3]: ")
	typeChoice, _ := reader.ReadString('\n')
	typeChoice = strings.TrimSpace(typeChoice)

	devType := "Unknown"
	switch typeChoice {
	case "1":
		devType = "DHT11 Env"
	case "2":
		devType = "PIR Motion"
	case "3":
		devType = "Relay Switch"
	default:
		devType = "DHT11 Env"
	}

	// 2. Network Auth Selection (For Phase 2 Testing)
	fmt.Println("\nSelect Security Mode:")
	fmt.Println("  1. Valid Credentials (Authorized)")
	fmt.Println("  2. Malicious Actor (Bad HMAC Signature)")
	fmt.Print("Choice [1/2]: ")
	secChoice, _ := reader.ReadString('\n')
	secChoice = strings.TrimSpace(secChoice)

	// 3. Generate or Input Virtual MAC Address
	fmt.Print("\nEnter MAC Address (Leave empty to Auto-Generate): ")
	macInput, _ := reader.ReadString('\n')
	macInput = strings.TrimSpace(macInput)
	mac := macInput

	if mac == "" {
		mac = fmt.Sprintf("%02X:%02X:%02X:%02X:%02X:%02X", 
			rand.Intn(256), rand.Intn(256), rand.Intn(256), 
			rand.Intn(256), rand.Intn(256), rand.Intn(256))
		fmt.Printf("Generated Virtual MAC: %s\n", mac)
	} else {
		fmt.Printf("Using provided MAC: %s\n", mac)
	}

	// 3.5 Ask for Secret Key
	fmt.Print("\nEnter HMAC Secret Key (Leave empty for Factory Secret): ")
	secretInput, _ := reader.ReadString('\n')
	secretInput = strings.TrimSpace(secretInput)
	if secretInput == "" {
		secretInput = "xomoi-factory-secret"
	}

	// 4. Connect to Broker
	fmt.Print("Target Broker [localhost:1883]: ")
	brokerInput, _ := reader.ReadString('\n')
	brokerInput = strings.TrimSpace(brokerInput)
	if brokerInput == "" {
		brokerInput = "localhost:1883"
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s", brokerInput))
	opts.SetClientID(mac)
	
	// Mock Auth injection based on selection
	if secChoice == "2" {
		opts.SetUsername("malicious_actor")
		opts.SetPassword("invalid_hash")
	} else {
		// Calculate valid HMAC-Lite token using the provided Secret
		h := hmac.New(sha256.New, []byte(secretInput))
		h.Write([]byte(mac))
		validHash := hex.EncodeToString(h.Sum(nil))

		opts.SetUsername(mac)
		opts.SetPassword(validHash)
	}

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("❌ Failed to connect to broker: %v", token.Error())
	}
	fmt.Println("✅ Connected to Xomoi-Core Dark Grid (Port 1883)")

	// 5. Announce Presence
	presencePayload := fmt.Sprintf(`{"device_id":"%s", "type":"%s", "status":"healthy", "state":"OFF"}`, mac, devType)
	client.Publish(fmt.Sprintf("/xomoi/%s/status", mac), 1, true, presencePayload)

	// 6. Listen for RPCs if it's an Actuator
	if devType == "Relay Switch" {
		rpcTopic := fmt.Sprintf("/xomoi/%s/rpc/command", mac)
		client.Subscribe(rpcTopic, 1, func(client mqtt.Client, msg mqtt.Message) {
			fmt.Printf("\n⚡ [RPC RECEIVED from Svelte Dashboard]: %s\n", msg.Payload())
			
			var payload map[string]interface{}
			json.Unmarshal(msg.Payload(), &payload)
			
			if payload["command"] == "toggle_relay" {
				fmt.Println("⚙️ Executing Relay Toggle (300ms)...")
				time.Sleep(300 * time.Millisecond)
				
				ackPayload := fmt.Sprintf(`{"device_id":"%s", "ack":"relay_success", "state":"ON"}`, mac)
				client.Publish(fmt.Sprintf("/xomoi/%s/rpc/ack", mac), 1, false, ackPayload)
				fmt.Println("✅ WebRTC ACK Sent!")
			}
		})
	}

	// 7. The Interactive Command Loop
	fmt.Println("\n--- INTERACTIVE COMMAND TERMINAL ---")
	fmt.Println("Press [Enter] to publish telemetry event.")
	fmt.Println("Press [Q] + [Enter] to gracefully shutdown.")

	for {
		fmt.Print("xomoi-sim> ")
		cmd, _ := reader.ReadString('\n')
		cmd = strings.TrimSpace(strings.ToUpper(cmd))

		if cmd == "Q" {
			fmt.Println("Publishing LWT Offline packet and shutting down...")
			client.Publish(fmt.Sprintf("/xomoi/%s/status", mac), 1, true, fmt.Sprintf(`{"device_id":"%s", "status":"offline"}`, mac))
			client.Disconnect(250)
			break
		}

		if devType == "DHT11 Env" {
			temp := 20.0 + rand.Float64()*15.0
			hum := 30.0 + rand.Float64()*40.0
			payload := fmt.Sprintf(`{"device_id":"%s", "temp":%.2f, "hum":%.2f}`, mac, temp, hum)
			client.Publish(fmt.Sprintf("/xomoi/%s/telemetry", mac), 0, false, payload)
			fmt.Printf("📤 Published to Dark Grid: %s\n", payload)
		} else if devType == "PIR Motion" {
			payload := fmt.Sprintf(`{"device_id":"%s", "motion":true}`, mac)
			client.Publish(fmt.Sprintf("/xomoi/%s/telemetry", mac), 0, false, payload)
			fmt.Printf("📤 Published Intrusion Event: %s\n", payload)
		} else {
			fmt.Println("Relay Switch telemetry is event-based. (It's waiting for RPCs from the UI)")
		}

		// Prevent infinite loops when running headless via pipes
		time.Sleep(2 * time.Second)
	}
}
