package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type LatencyPayload struct {
	Temp      float64 `json:"temp"`
	Hum       float64 `json:"hum"`
	State     string  `json:"state"`
	Timestamp int64   `json:"timestamp"`
}

func main() {
	targetIP := flag.String("ip", "localhost", "The Tailscale IP or localhost")
	workers := flag.Int("workers", 50, "Number of concurrent connections")
	duration := flag.Int("time", 15, "Duration in seconds")
	qos := flag.Int("qos", 1, "MQTT QoS level (0 or 1)")
	mode := flag.String("mode", "ingest", "Benchmark mode: 'ingest', 'fanout', or 'latency'")
	flag.Parse()

	factorySecret := "xomoi-factory-secret"
	brokerURL := fmt.Sprintf("tcp://%s:1883", *targetIP)
	
	log.Printf("Starting Xomoi-Core Benchmark Suite")
	log.Printf("Target: %s | QoS: %d | Mode: %s | Workers: %d | Time: %ds", brokerURL, *qos, *mode, *workers, *duration)

	switch *mode {
	case "ingest":
		runIngestBenchmark(brokerURL, factorySecret, *workers, *duration, byte(*qos))
	case "fanout":
		runFanoutBenchmark(brokerURL, factorySecret, *workers, *duration, byte(*qos))
	case "latency":
		runLatencyBenchmark(brokerURL, factorySecret, *duration, byte(*qos))
	default:
		log.Fatalf("Unknown mode: %s. Use 'ingest', 'fanout', or 'latency'", *mode)
	}
}

func runIngestBenchmark(brokerURL, secret string, workers, duration int, qos byte) {
	var totalSent uint64
	var wg sync.WaitGroup

	payload := []byte(`{"temp":42.5,"hum":88.1,"state":"ACTIVE","tensor":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20],"padding":"xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"}`)
	start := time.Now()

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			macAddress := fmt.Sprintf("AA:BB:CC:DD:EE:%02X", workerID)
			client := connectClient(brokerURL, macAddress, secret)
			if client == nil {
				return
			}
			defer client.Disconnect(250)

			topic := fmt.Sprintf("/xomoi/%s/telemetry", macAddress)
			timer := time.NewTimer(time.Duration(duration) * time.Second)

			for {
				select {
				case <-timer.C:
					return
				default:
					token := client.Publish(topic, qos, false, payload)
					token.Wait()
					if token.Error() == nil {
						atomic.AddUint64(&totalSent, 1)
					}
				}
			}
		}(i)
	}

	wg.Wait()
	printResults(totalSent, start, len(payload))
}

func runFanoutBenchmark(brokerURL, secret string, subs, duration int, qos byte) {
	var totalReceived uint64
	var wg sync.WaitGroup

	// 1. Connect all subscribers
	macPublisher := "AA:BB:CC:DD:EE:FF" // Publisher
	topic := fmt.Sprintf("/xomoi/%s/telemetry", macPublisher)

	log.Printf("Connecting %d subscribers...", subs)
	for i := 0; i < subs; i++ {
		wg.Add(1)
		go func(subID int) {
			defer wg.Done()
			macSub := fmt.Sprintf("SUB:BB:CC:DD:EE:%02X", subID)
			client := connectClient(brokerURL, macSub, secret)
			if client == nil {
				return
			}
			defer client.Disconnect(250)

			// Xomoi ACL check: Subscribers might need special auth to read another's topic.
			// Assuming for now they can subscribe.
			client.Subscribe(topic, qos, func(c mqtt.Client, m mqtt.Message) {
				atomic.AddUint64(&totalReceived, 1)
			}).Wait()
			
			time.Sleep(time.Duration(duration+2) * time.Second)
		}(i)
	}
	
	time.Sleep(1 * time.Second) // Wait for subs to connect

	// 2. Start Publisher
	payload := []byte(`{"temp":42.5,"hum":88.1,"state":"FANOUT"}`)
	client := connectClient(brokerURL, macPublisher, secret)
	timer := time.NewTimer(time.Duration(duration) * time.Second)
	start := time.Now()

	go func() {
		for {
			select {
			case <-timer.C:
				return
			default:
				client.Publish(topic, qos, false, payload).Wait()
			}
		}
	}()

	wg.Wait()
	log.Printf("\n--- FANOUT RESULTS ---")
	log.Printf("Subscribers: %d", subs)
	log.Printf("Total Packets Received by Subs: %d", totalReceived)
	log.Printf("Fan-out Throughput: %.2f msgs/sec", float64(totalReceived)/time.Since(start).Seconds())
}

func runLatencyBenchmark(brokerURL, secret string, duration int, qos byte) {
	macAddress := "LATENCY:CC:DD:EE:FF"
	topic := fmt.Sprintf("/xomoi/%s/telemetry", macAddress)

	client := connectClient(brokerURL, macAddress, secret)
	if client == nil { return }
	defer client.Disconnect(250)

	var totalLatency int64
	var count int64

	client.Subscribe(topic, qos, func(c mqtt.Client, m mqtt.Message) {
		var p LatencyPayload
		if err := json.Unmarshal(m.Payload(), &p); err == nil {
			latency := time.Now().UnixMilli() - p.Timestamp
			atomic.AddInt64(&totalLatency, latency)
			atomic.AddInt64(&count, 1)
		}
	}).Wait()

	timer := time.NewTimer(time.Duration(duration) * time.Second)
	go func() {
		for {
			select {
			case <-timer.C:
				return
			default:
				p := LatencyPayload{Temp: 25.5, Hum: 60.0, State: "PING", Timestamp: time.Now().UnixMilli()}
				b, _ := json.Marshal(p)
				client.Publish(topic, qos, false, b).Wait()
				time.Sleep(10 * time.Millisecond) // Pace it to measure stable latency
			}
		}
	}()

	<-timer.C
	time.Sleep(500 * time.Millisecond) // Flush incoming

	avg := float64(totalLatency) / float64(count)
	log.Printf("\n--- LATENCY RESULTS ---")
	log.Printf("Packets Sampled: %d", count)
	log.Printf("Average End-to-End Latency: %.2f ms", avg)
}

func connectClient(brokerURL, macAddress, secret string) mqtt.Client {
	macHmac := hmac.New(sha256.New, []byte(secret))
	macHmac.Write([]byte(macAddress))
	password := hex.EncodeToString(macHmac.Sum(nil))

	opts := mqtt.NewClientOptions()
	opts.AddBroker(brokerURL)
	opts.SetClientID(macAddress)
	opts.SetUsername(macAddress)
	opts.SetPassword(password)
	
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Printf("Failed to connect %s: %v", macAddress, token.Error())
		return nil
	}
	return client
}

func printResults(totalSent uint64, start time.Time, payloadSize int) {
	durationSec := time.Since(start).Seconds()
	throughput := float64(totalSent) / durationSec
	bytesPerSec := float64(uint64(payloadSize)*totalSent) / durationSec
	mbPerSec := bytesPerSec / (1024 * 1024)

	fmt.Printf("\n--- INGESTION RESULTS ---\n")
	fmt.Printf("Total Packets Sent: %d\n", totalSent)
	fmt.Printf("Duration: %.2f seconds\n", durationSec)
	fmt.Printf("Throughput: %.2f msgs/sec\n", throughput)
	fmt.Printf("Payload size: %d bytes\n", payloadSize)
	fmt.Printf("Data Rate: %.2f MB/sec\n", mbPerSec)
}
