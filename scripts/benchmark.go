package main

import (
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
	"github.com/gorilla/websocket"
)

type LatencyPayload struct {
	Temp      float64 `json:"temp"`
	Hum       float64 `json:"hum"`
	State     string  `json:"state"`
	Timestamp int64   `json:"timestamp"`
}

type HealthStats struct {
	RamUsageMB float64 `json:"ram_usage_mb"`
	NumWorkers int     `json:"num_workers"`
	NumCPU     int     `json:"num_cpu"`
	WalSizeMB  float64 `json:"wal_size_mb"`
	UptimeSec  int64   `json:"uptime_sec"`
	GcPausesNs uint64  `json:"gc_pauses_ns"`
	HeapSysMb  float64 `json:"heap_sys_mb"`
	Goroutines int     `json:"goroutines"`
}

type BenchmarkMetrics struct {
	MaxRamMB    float64
	MaxGCPause  uint64
	MaxGorout   int
	MaxWalSize  float64
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
	wsURL := fmt.Sprintf("ws://%s:8085/api/v1/ws/health", *targetIP)
	
	log.Printf("Starting Xomoi-Core Benchmark Suite")
	log.Printf("Target: %s | QoS: %d | Mode: %s | Workers: %d | Time: %ds", brokerURL, *qos, *mode, *workers, *duration)

	metrics := monitorNodeHealth(wsURL, *duration)

	switch *mode {
	case "ingest":
		runIngestBenchmark(brokerURL, factorySecret, *workers, *duration, byte(*qos), metrics)
	case "fanout":
		runFanoutBenchmark(brokerURL, factorySecret, *workers, *duration, byte(*qos), metrics)
	case "latency":
		runLatencyBenchmark(brokerURL, factorySecret, *duration, byte(*qos), metrics)
	default:
		log.Fatalf("Unknown mode: %s. Use 'ingest', 'fanout', or 'latency'", *mode)
	}
}

func monitorNodeHealth(wsURL string, duration int) *BenchmarkMetrics {
	metrics := &BenchmarkMetrics{}
	
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		log.Printf("[WARNING] Could not connect to Node Health API (%v). Hardware stats will be unavailable.", err)
		return metrics
	}
	log.Printf("[INFO] Intercepted WebSockets Node Health stream...")

	timer := time.NewTimer(time.Duration(duration+2) * time.Second)

	go func() {
		defer conn.Close()
		for {
			select {
			case <-timer.C:
				return
			default:
				conn.SetReadDeadline(time.Now().Add(3 * time.Second))
				_, msg, err := conn.ReadMessage()
				if err != nil {
					return
				}
				var stats HealthStats
				if err := json.Unmarshal(msg, &stats); err == nil {
					if stats.RamUsageMB > metrics.MaxRamMB {
						metrics.MaxRamMB = stats.RamUsageMB
					}
					if stats.GcPausesNs > metrics.MaxGCPause {
						metrics.MaxGCPause = stats.GcPausesNs
					}
					if stats.Goroutines > metrics.MaxGorout {
						metrics.MaxGorout = stats.Goroutines
					}
					if stats.WalSizeMB > metrics.MaxWalSize {
						metrics.MaxWalSize = stats.WalSizeMB
					}
				}
			}
		}
	}()

	return metrics
}

func runIngestBenchmark(brokerURL, secret string, workers, duration int, qos byte, metrics *BenchmarkMetrics) {
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
	time.Sleep(1 * time.Second)
	
	durationSec := time.Since(start).Seconds()
	throughput := float64(totalSent) / durationSec
	bytesPerSec := float64(uint64(len(payload))*totalSent) / durationSec
	mbPerSec := bytesPerSec / (1024 * 1024)

	fmt.Printf("\n--- INGESTION RESULTS ---\n")
	fmt.Printf("Total Packets Sent: %d\n", totalSent)
	fmt.Printf("Duration: %.2f seconds\n", durationSec)
	fmt.Printf("Throughput: %.2f msgs/sec\n", throughput)
	fmt.Printf("Payload size: %d bytes\n", len(payload))
	fmt.Printf("Data Rate: %.2f MB/sec\n", mbPerSec)
	printHardwareMetrics(metrics)
}

func runFanoutBenchmark(brokerURL, secret string, subs, duration int, qos byte, metrics *BenchmarkMetrics) {
	var totalReceived uint64
	var wg sync.WaitGroup

	macPublisher := "AA:BB:CC:DD:EE:FF"
	topic := fmt.Sprintf("/xomoi/%s/telemetry", macPublisher)

	log.Printf("Connecting %d subscribers...", subs)
	for i := 0; i < subs; i++ {
		wg.Add(1)
		go func(subID int) {
			defer wg.Done()
			macSub := fmt.Sprintf("SUB:BB:CC:DD:EE:%02X", subID)
			client := connectClient(brokerURL, macSub, secret)
			if client == nil { return }
			defer client.Disconnect(250)

			client.Subscribe(topic, qos, func(c mqtt.Client, m mqtt.Message) {
				atomic.AddUint64(&totalReceived, 1)
			}).Wait()
			
			time.Sleep(time.Duration(duration+2) * time.Second)
		}(i)
	}
	
	time.Sleep(1 * time.Second)

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
	printHardwareMetrics(metrics)
}

func runLatencyBenchmark(brokerURL, secret string, duration int, qos byte, metrics *BenchmarkMetrics) {
	macAddress := "AA:BB:CC:DD:EE:FF"
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

	timeout := time.After(time.Duration(duration) * time.Second)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			default:
				p := LatencyPayload{Temp: 25.5, Hum: 60.0, State: "PING", Timestamp: time.Now().UnixMilli()}
				b, _ := json.Marshal(p)
				client.Publish(topic, qos, false, b).Wait()
				time.Sleep(20 * time.Millisecond) // Give network time to bounce
			}
		}
	}()

	<-timeout
	close(done)
	time.Sleep(500 * time.Millisecond) // Flush incoming

	avg := float64(totalLatency) / float64(count)
	fmt.Printf("\n--- LATENCY RESULTS ---\n")
	fmt.Printf("Packets Sampled: %d\n", count)
	fmt.Printf("Average End-to-End Latency: %.2f ms\n", avg)
	printHardwareMetrics(metrics)
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

func printHardwareMetrics(metrics *BenchmarkMetrics) {
	if metrics.MaxRamMB > 0 {
		fmt.Printf("\n--- BROKER HARDWARE LIMITS HIT ---\n")
		fmt.Printf("Max RAM Allocated: %.2f MB\n", metrics.MaxRamMB)
		fmt.Printf("Max GC Pause Latency: %d ns (%.3f ms)\n", metrics.MaxGCPause, float64(metrics.MaxGCPause)/1e6)
		fmt.Printf("Max Goroutines Sprawled: %d\n", metrics.MaxGorout)
		fmt.Printf("Max SQLite WAL Size: %.2f MB\n", metrics.MaxWalSize)
	}
}
