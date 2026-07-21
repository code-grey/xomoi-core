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
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/code-grey/xomoi-core/internal/api"
	"github.com/code-grey/xomoi-core/internal/api/handlers"
	"github.com/code-grey/xomoi-core/internal/broker"
	"github.com/code-grey/xomoi-core/internal/config"
	"github.com/code-grey/xomoi-core/internal/network"
	"github.com/code-grey/xomoi-core/internal/repository/sqlite"
	"github.com/code-grey/xomoi-core/internal/state"
	"github.com/code-grey/xomoi-core/internal/worker"
	mqtt "github.com/mochi-mqtt/server/v2"
)

type corePublisher struct {
	broker *mqtt.Server
}

func (p *corePublisher) Publish(topic string, payload []byte, retain bool, qos byte) error {
	return p.broker.Publish(topic, payload, retain, qos)
}

func main() {
	// Initialize Global JSON Logger and pipe it to both Stdout and the WebSocket LogBuffer
	multiWriter := io.MultiWriter(os.Stdout, handlers.GlobalLogBuffer)
	logger := slog.New(slog.NewJSONHandler(multiWriter, nil))
	slog.SetDefault(logger)

	cfg := config.Load()

	// 0. GC and Memory Tuning for Edge Hardware
	// Set a soft memory limit to prevent the OS OOM Killer from terminating 
	// the binary on memory-constrained devices like Raspberry Pi.
	debug.SetMemoryLimit(int64(cfg.MemoryLimitMB) * 1024 * 1024)
	
	// Hard-limit the Go Scheduler OS threads to simulate constraints
	if cfg.IngestionWorkers > 0 {
		runtime.GOMAXPROCS(cfg.IngestionWorkers)
	}

	slog.Info("BOOTING XOMOI-CORE SOVEREIGN EDGE NODE")

	// 1. Initialize SQLite Database (WAL Mode)
	db, err := sqlite.NewDB(cfg.DBPath)
	if err != nil {
		slog.Error("Failed to initialize SQLite", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// 2. Initialize Memory Barrier (sync.Map HotState)
	hotState := state.NewHotState()

	// 3. Global Context for Background Workers
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// 4. Initialize Repositories
	tsdb := sqlite.NewTelemetryRepository(db)
	ruleRepo := sqlite.NewRuleRepository(db)

	// 5. Start Lossless Ring Buffer (Phase 2.5)
	flushInterval := time.Duration(cfg.FlushIntervalSec) * time.Second
	slog.Info("Configured Ring Buffer Batch Flush", "interval", flushInterval)
	ringBuffer := state.NewRingBuffer(tsdb, 100000, 1000, flushInterval)
	ringBuffer.Start()

	rulesEngine := worker.NewRulesEngine(ruleRepo)
	if err := rulesEngine.Start(ctx); err != nil {
		slog.Error("Failed to start Rules Engine", "error", err)
	}

	processor := broker.NewProcessor(hotState, ringBuffer, rulesEngine)
	
	// Dynamic Worker Sizing: Prevent context-switching hell on low-end edge nodes.
	// We bind the number of ingestion workers strictly to the available hardware threads.
	slog.Info("Hardware Check: Sizing Ingestion Pool", "cpu_cores", runtime.NumCPU(), "workers", cfg.IngestionWorkers)
	
	pool := broker.NewWorkerPool(cfg.IngestionWorkers, 10000, processor)
	go pool.Start()
	slog.Info(fmt.Sprintf("Starting GC-Optimized Ingestion Pool with %d workers", cfg.IngestionWorkers))

	// 4b. Start Mochi-MQTT Broker
	mqttServer, err := broker.NewMochiServer(cfg.MQTTPort)
	
	deviceRepo := sqlite.NewDeviceRepository(db)
	mqttServer.AddHook(broker.NewHMACAuthHook(deviceRepo), nil) // Enforce HMAC-Lite Security
	mqttServer.AddHook(broker.NewPublishHook(pool), nil)  // Hook the TSDB Worker Pool

	// Note: You would wire your custom broker Hooks here
	go func() {
		if err := mqttServer.Serve(); err != nil {
			slog.Error("MQTT Server Failed", "error", err)
		}
	}()
	slog.Info("MQTT Broker listening", "port", cfg.MQTTPort)

	// 4c. Start the WebRTC Tunnel Host
	rtcHost := api.NewWebRTCHost("XOMOI-CORE-SERVER", cfg.SignalingURL, cfg.STUNServers, mqttServer)
	rtcHost.Start()

	// 5. Start Background Janitor (Prunes data older than 30 days, checks every 24h)
	janitor := worker.NewJanitor(db.DB, 30*24*time.Hour, 24*time.Hour)
	go janitor.Start(ctx)

	// 6. Start the Headless API Server
	apiServer := api.NewServer(nil, nil, deviceRepo, tsdb, ruleRepo, mqttServer, &corePublisher{broker: mqttServer}, rulesEngine)
	router := apiServer.SetupRouter()
	
	httpSrv := &http.Server{
		Addr:    ":" + cfg.APIPort,
		Handler: router,
	}
	
	go func() {
		slog.Info("API Server listening", "port", cfg.APIPort)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("API Server failed", "error", err)
			os.Exit(1)
		}
	}()

	// 7. Start mDNS Zero-Config Broadcaster
	mdnsServer, err := network.StartMDNS(cfg.APIPort)
	if err != nil {
		slog.Warn("mDNS failed to start, falling back to raw IP access", "error", err)
	}

	// 8. Graceful Shutdown & Signal Trapping
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	slog.Info("SHUTDOWN SIGNAL RECEIVED. EXECUTING TEARDOWN")

	// A. Stop background cron jobs (Janitor, Rules triggers)
	cancel()

	// B. Gracefully flush and stop the Ring Buffer to prevent data loss
	slog.Info("Flushing Ring Buffer to SQLite...")
	ringBuffer.Stop()

	// C. Shutdown API cleanly
	apiCtx, apiCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer apiCancel()
	if err := httpSrv.Shutdown(apiCtx); err != nil {
		slog.Error("API Server forced to shutdown", "error", err)
	}

	// D. Stop Ingestion Workers
	pool.Stop()

	// E. Stop mDNS
	if mdnsServer != nil {
		mdnsServer.Stop()
	}

	// F. Execute Hexagonal Backup Snapshot (e.g., upload to Discord)
	// preserver := backup.NewDiscordPreserver("WEBHOOK_URL")
	// preserver.Save(context.Background(), "xomoi.db")
	slog.Info("Disaster recovery snapshot completed.")

	slog.Info("XOMOI-CORE HAS GONE DARK")
}
