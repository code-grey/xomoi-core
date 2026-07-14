package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"
	"sync/atomic"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/joho/godotenv"
)

var totalPublished uint64

func main() {
	_ = godotenv.Load("../../.env")
	
	numDevices := 1000
	msgDelay := 10 * time.Microsecond // 100,000 msgs/sec per device -> 100,000,000 msgs/sec total

	fmt.Printf("🔥 STARTING XOMOI-CORE STRESS TEST 🔥\n")
	fmt.Printf("Virtual Devices: %d\n", numDevices)
	fmt.Printf("Target Publish Rate: ~%d msgs/sec\n", int64(numDevices)*(time.Second.Nanoseconds()/msgDelay.Nanoseconds()))

	var wg sync.WaitGroup
	for i := 0; i < numDevices; i++ {
		wg.Add(1)
		go spawnDevice(i, msgDelay, &wg)
		// Stagger connections by 2ms to prevent overwhelming the TCP backlog queue
		time.Sleep(2 * time.Millisecond)
	}

	go func() {
		lastCount := uint64(0)
		for {
			time.Sleep(1 * time.Second)
			current := atomic.LoadUint64(&totalPublished)
			rate := current - lastCount
			lastCount = current
			fmt.Printf("📈 Throughput: %d msgs/sec\n", rate)
		}
	}()

	wg.Wait()
}

func spawnDevice(id int, delay time.Duration, wg *sync.WaitGroup) {
	defer wg.Done()
	
	mac := fmt.Sprintf("00:11:22:33:ST:%02X", id)
	
	secret := os.Getenv("XOMOI_FACTORY_SECRET")
	if secret == "" {
		secret = "xomoi-factory-secret"
	}

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(mac))
	validHash := hex.EncodeToString(h.Sum(nil))

	opts := mqtt.NewClientOptions()
	brokerHost := os.Getenv("XOMOI_BROKER_URL")
	if brokerHost == "" {
		brokerHost = "tcp://localhost:1883"
	}
	opts.AddBroker(brokerHost)
	opts.SetClientID(mac)
	opts.SetUsername(mac)
	opts.SetPassword(validHash) 

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Printf("Device %d failed to connect: %v", id, token.Error())
		return
	}

	topic := fmt.Sprintf("/xomoi/%s/telemetry", mac)
	
	for {
		temp := 20.0 + rand.Float64()*15.0
		hum := 30.0 + rand.Float64()*40.0
		payload := fmt.Sprintf(`{"device_id":"%s", "temp":%.2f, "hum":%.2f}`, mac, temp, hum)

		token := client.Publish(topic, 0, false, payload)
		token.Wait()
		
		if token.Error() != nil {
			log.Printf("Device %d publish failed: %v", id, token.Error())
		} else {
			atomic.AddUint64(&totalPublished, 1)
		}

		time.Sleep(delay)
	}
}
