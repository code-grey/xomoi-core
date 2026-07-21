package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func main() {
	targetIP := flag.String("ip", "localhost", "The Tailscale IP of the Fedora Pentium PC")
	workers := flag.Int("workers", 50, "Number of concurrent MQTT connections blasting data")
	duration := flag.Int("time", 15, "Duration to blast telemetry in seconds")
	flag.Parse()

	factorySecret := "xomoi-factory-secret"
	brokerURL := fmt.Sprintf("tcp://%s:1883", *targetIP)
	log.Printf("Connecting to Xomoi-Core at %s with %d workers...", brokerURL, *workers)

	var totalSent uint64
	var wg sync.WaitGroup

	// Payload is just a raw bytes tensor simulation (256 bytes)
	payload := make([]byte, 256)
	for i := 0; i < len(payload); i++ {
		payload[i] = byte(i % 255)
	}

	start := time.Now()

	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			macAddress := fmt.Sprintf("AA:BB:CC:DD:EE:%02X", workerID)
			
			// Generate HMAC-Lite Password using the factory secret
			macHmac := hmac.New(sha256.New, []byte(factorySecret))
			macHmac.Write([]byte(macAddress))
			password := hex.EncodeToString(macHmac.Sum(nil))

			opts := mqtt.NewClientOptions()
			opts.AddBroker(brokerURL)
			opts.SetClientID(macAddress)
			opts.SetUsername(macAddress)
			opts.SetPassword(password)
			opts.SetAutoReconnect(true)
			opts.SetCleanSession(true)

			client := mqtt.NewClient(opts)
			if token := client.Connect(); token.Wait() && token.Error() != nil {
				log.Printf("Worker %d failed to connect: %v", workerID, token.Error())
				return
			}
			defer client.Disconnect(250)

			topic := fmt.Sprintf("device/%s/telemetry", macAddress)

			// Tight blasting loop
			timer := time.NewTimer(time.Duration(*duration) * time.Second)
			for {
				select {
				case <-timer.C:
					return
				default:
					token := client.Publish(topic, 0, false, payload)
					token.Wait()
					if token.Error() == nil {
						atomic.AddUint64(&totalSent, 1)
					}
				}
			}
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start).Seconds()
	sent := atomic.LoadUint64(&totalSent)
	
	fmt.Printf("\n--- BENCHMARK RESULTS ---\n")
	fmt.Printf("Total Packets Sent: %d\n", sent)
	fmt.Printf("Duration: %.2f seconds\n", elapsed)
	fmt.Printf("Throughput: %.2f msgs/sec\n", float64(sent)/elapsed)
	fmt.Printf("Payload size: %d bytes\n", len(payload))
	fmt.Printf("Data Rate: %.2f MB/sec\n", (float64(sent)*float64(len(payload)))/1024/1024/elapsed)
}
