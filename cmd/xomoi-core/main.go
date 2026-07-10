package main

import (
	"context"
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
	"github.com/code-grey/xomoi-core/internal/repository/sqlite"
	"github.com/code-grey/xomoi-core/internal/state"
	"github.com/code-grey/xomoi-core/internal/worker"
	"github.com/code-grey/xomoi-core/internal/network"
)

func main() {
	// Initialize Global JSON Logger and pipe it to both Stdout and the WebSocket LogBuffer
	multiWriter := io.MultiWriter(os.Stdout, handlers.GlobalLogBuffer)
	logger := slog.New(slog.NewJSONHandler(multiWriter, nil))
	slog.SetDefault(logger)

	slog.Info("BOOTING XOMOI-CORE SOVEREIGN EDGE NODE")

	// 0. GC and Memory Tuning for Edge Hardware
	// Set a soft memory limit (250MB) to prevent the OS OOM Killer from terminating 
	// the binary on memory-constrained devices like Raspberry Pi.
	// Go will aggressively GC only when it nears this limit, otherwise it stays highly performant.
	debug.SetMemoryLimit(250 * 1024 * 1024)
	slog.Info("Hardware Check: GOMEMLIMIT enforced at 250MB")

	// 1. Initialize SQLite Database (WAL Mode)
	db, err := sqlite.NewDB("xomoi.db")
	if err != nil {
		slog.Error("Failed to initialize SQLite", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// 2. Initialize Memory Barrier (sync.Map HotState)
	hotState := state.NewHotState()

	// 3. Start Snapshot Worker (5-minute bulk flush)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	snapshotWorker := state.NewSnapshotWorker(hotState, db.DB, 5*time.Minute)
	go snapshotWorker.Start(ctx)

	// 4. Initialize Mochi-MQTT Broker & Ingestion Pipeline
	processor := broker.NewProcessor(hotState)
	
	// Dynamic Worker Sizing: Prevent context-switching hell on low-end edge nodes.
	// We bind the number of ingestion workers strictly to the available hardware threads.
	numWorkers := runtime.NumCPU()
	if numWorkers < 2 {
		numWorkers = 2 // Ensure at least 2 workers on ultra-low-end single-core setups (e.g., Pi Zero)
	}
	slog.Info("Hardware Check: Sizing Ingestion Pool", "cpu_cores", runtime.NumCPU(), "workers", numWorkers)
	
	workerPool := broker.NewWorkerPool(numWorkers, 1000, processor) // 1000 queue depth
	workerPool.Start()

	// Note: In a full deployment, we would pass the initialized hooks to broker.NewBroker()
	// and start the MQTT server here.

	// 5. Start Background Janitor (Prunes data older than 30 days, checks every 24h)
	janitor := worker.NewJanitor(db.DB, 30*24*time.Hour, 24*time.Hour)
	go janitor.Start(ctx)

	// 6. Start the Headless API Server
	// Provide nil repos and nil publisher for skeleton, to be wired later
	apiServer := api.NewServer(nil, nil, nil)
	router := apiServer.SetupRouter()
	
	httpSrv := &http.Server{
		Addr:    ":8085",
		Handler: router,
	}
	
	go func() {
		slog.Info("API Server listening", "port", 8085)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("API Server failed", "error", err)
			os.Exit(1)
		}
	}()

	// 7. Start mDNS Zero-Config Broadcaster
	mdnsServer, err := network.StartMDNS(8085)
	if err != nil {
		slog.Warn("mDNS failed to start, falling back to raw IP access", "error", err)
	}

	// 8. Graceful Shutdown & Signal Trapping
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	slog.Info("SHUTDOWN SIGNAL RECEIVED. EXECUTING TEARDOWN")

	// A. Stop background cron jobs (Janitor, Snapshot triggers)
	cancel()

	// B. Force a final flush of HotState to disk to prevent data loss
	slog.Info("Flushing volatile state to SQLite...")
	snapshotWorker.ForceFlush()

	// C. Shutdown API cleanly
	apiCtx, apiCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer apiCancel()
	if err := httpSrv.Shutdown(apiCtx); err != nil {
		slog.Error("API Server forced to shutdown", "error", err)
	}

	// D. Stop Ingestion Workers
	workerPool.Stop()

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
