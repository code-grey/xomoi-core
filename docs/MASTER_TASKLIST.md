# XOMOI-CORE v1.0 MASTER TASKLIST (5-MONTH SOVEREIGN ROADMAP)

## PHASE 1: THE SOVEREIGN CONTRACT (WEEKS 1-2)
- [x] **1.1: Modular Protobuf "Core + Plugins" (`proto/v1/`)**
    - [x] `common.proto`: Shared enums, Vector3, and Field ID locks.
    - [x] `telemetry_base.proto`: Standard sensor types (Float, Int, Bool).
    - [x] `telemetry_pro.proto`: Advanced types (JSON, Bytes) for AI/Imaging.
    - [x] `registry.proto`: Tag-to-Name handshake mapping.
    - [x] `command.proto`: Bidirectional control (Toggle, Value, Text).
- [x] **1.2: Core Domain Models (`internal/core/`)**
    - [x] `models.go`: Define `User`, `Device`, `Session`, `SensorTag`, and `AlertRule` Go structs.
- [x] **1.3: Repository Interfaces (`internal/repository/`)**
    - [x] `interfaces.go`: Define pure interfaces for User, Session, Device, and Telemetry.

## PHASE 2: STORAGE & MEMORY ENGINE (WEEKS 3-4)
- [x] **2.1: SQLite Engine (`internal/repository/sqlite/`)**
    - [x] `db.go`: Initialize `sql.DB` with WAL mode and `PRAGMA busy_timeout`.
    - [x] `migrations/`: Create embedded SQL files for all core tables.
- [x] **2.2: Volatile State Manager (`internal/state/`)**
    - [x] `hot_state.go`: Implement `sync.Map` for O(1) real-time sensor status.
    - [x] `snapshot.go`: Implement the 5-minute bulk-flush routine to SQLite.

## PHASE 2.5: ENTERPRISE TELEMETRY UPGRADE
- [x] **2.5.1: Database Decoupling & Hypertables**
    - [x] Isolate `HotState` strictly for O(1) WebRTC Dashboard reads.
    - [x] Update SQLite `telemetry_history` schema to use Auto-Incrementing IDs/ULIDs (Drop Composite PK).
    - [ ] Implement SQLite Hypertables (Time Partitioning via `ATTACH DATABASE` or month-suffixed tables) for TimescaleDB-like speed on massive datasets.
- [x] **2.5.2: Lossless Ingestion Pipeline**
    - [x] Implement high-speed in-memory Ring Buffer / Channel Queue for raw packets.
- [x] **2.5.3: Event-Driven TSDB Flusher**
    - [x] Replace `SnapshotWorker` with a batch flusher that triggers on queue limits.
- [x] **2.5.4: Zstd Payload Compression (Storage Layer)**
    - [x] Compress JSON payloads to BLOBs in Go before SQLite write to save 80% disk space.
- [ ] **2.5.5: Embedded Schema Migrations**
    - [ ] Integrate `golang-migrate` for automatic database schema updates on OTA reboot.
- [ ] **2.5.6: Store-and-Forward Cloud Sync**
    - [ ] Implement an incremental replication worker to push offline telemetry to the central cloud upon reconnection.

## PHASE 3: THE EMBEDDED BROKER (WEEKS 5-6)
- [x] **3.1: Mochi-MQTT Integration (`internal/broker/`)**
    - [x] `mochi.go`: Initialize embedded server (TCP + WebSockets).
    - [x] `auth.go`: Implement **HMAC-Lite Hook** (Verify MAC + Timestamp + Signature).
- [x] **3.2: The Ingestion Pipeline**
    - [x] `worker_pool.go`: Fixed-size Worker Pool for message processing.
    - [x] `processor.go`: OnPublish -> Proto Unmarshal -> Worker Channel (Backpressure enabled).

## PHASE 4: SOVEREIGN API & SECURITY (WEEKS 7-8)
- [x] **4.1: The Headless Web Server (`internal/api/`)**
    - [x] `router.go`: Go 1.26 `net/http` ServeMux initialization.
    - [x] `middleware/`: Stateful Session check, Anti-CSRF, and Panic Recovery.
- [x] **4.2: Auth & Session Endpoints**
    - [x] `POST /api/v1/auth/login`: Argon2ID verify + Session create.
    - [x] `POST /api/v1/auth/logout`: Session deletion.

## PHASE 5: THE GRAND ARCHITECT (WEEKS 9-11)
- [x] **5.1: Xomoi-Transpiler (`xomoi-ctl`)**
    - [x] `parser`: Go-based Protobuf parser to read `proto/v1/`.
    - [x] `generator/cpp`: Adaptive pruning logic to generate "Lite" C++ SDK headers.
- [x] **5.2: Xomoi-Claim Flow**
    - [x] `GET /api/v1/devices/discover`: Scan for `Xomoi-Claim-XXXX` signals.
    - [x] `POST /api/v1/devices/claim`: Generate HMAC-Lite token and push to device.

## PHASE 6: THE CROSS-PLATFORM UI (WEEKS 12-14)
- [x] **6.1: Svelte 5 Dashboard (`ui/`)**
    - [x] Zero-dependency SPA Hash Routing.
    - [x] Real-time telemetry via MQTT-over-WebSockets with SVG gradients.
    - [x] Embed Web build into Go binary via `go:embed`.
- [x] **6.2: Low-Code Onboarding**
    - [x] Web-Flasher: WebSerial integration (`esptool-js`) for one-click generic firmware flashing directly from Chrome.

## PHASE 7: THE BLACKSMITH SDK (WEEKS 15-16)
- [x] **7.1: Xomoi C++ SDK Foundation (`sdk/`)**
    - [x] Protobuf Discovery Schema (`xomoi.proto`) and NanoPB constraints (`xomoi.options`).
    - [x] Generated NanoPB C headers (`xomoi.pb.c`) for zero-allocation memory on the ESP32.
- [x] **7.2: Xomoi C++ Engine Implementation**
    - [x] Hardware SHA-256 HMAC-Lite Cryptography via `mbedtls`.
    - [x] Implement Protobuf byte streaming and Discovery callbacks in `XomoiCore.cpp`.
    - [x] Implement Telemetry Batching and two-way RPC hooks.

## PHASE 8: ALERTS & HEXAGONAL BACKUP (WEEKS 17-18)
- [x] **8.1: Alert Engine (`internal/worker/rules_engine.go`)**
    - [x] Zero-Allocation `sync.RWMutex` evaluation cache against hot state.
    - [x] Svelte UI Rule Builder and `GET/POST/DELETE` API endpoints (`handlers/rules.go`).
- [x] **8.2: Hexagonal Backup Engine (`internal/backup/`)**
    - [x] `BackupProvider` Interface: Abstract `Save()` and `Restore()`.
    - [x] Implement `DiscordPreserver` (Webhook attachments).
    - [x] Implement `DrivePreserver` (Google Drive Service Account).
    - [x] Implement `PostgresPreserver` (Cloud syncing).

## PHASE 9: THE JANITOR & HARDENING (WEEKS 19-20)
- [x] **9.1: Background Janitor (`internal/worker/`)**
    - [x] 1m/5m/1h telemetry aggregation and raw data pruning for threshold control.
- [ ] **9.2: Testing & QA**
    - [ ] Go unit tests (`go test`) for all core business logic (evaluators, parsers, state maps).
    - [ ] Svelte UI component tests using Vitest or Playwright.
    - [x] E2E integration tests (spin up broker -> mock C++ hardware -> verify UI state).
- [x] **9.3: Pre-Production Polish & Hardening**
    - [x] Graceful Shutdown (force SQLite flush + Backup on SIGTERM).
    - [x] Centralized Environment Configuration (`internal/config/config.go`).
    - [x] Migration Squashing (`001_init.sql`).
    - [x] Robust WebRTC Signaling (Exponential Backoff).
    - [x] Time-Series UI legend interpolation logic fixed.
    - [ ] Security Audit, Binary Optimization (`ldflags`), and v1.0 Release.

## PHASE 9.5: THE TERMINAL DASHBOARD (TUI)
- [x] **9.5.1: Xomoi-CLI (`cmd/xomoi-cli/`)**
    - [x] `charmbracelet/bubbletea` integration for a zero-bloat standalone terminal UI.
    - [x] Matrix-style real-time Telemetry, Worker Pool health, and Active Claims viewing.

## PHASE 10: REMOTE OPERATIONS & OTA (WEEKS 21-22)
- [ ] **10.1: OTA Engine & Device Management**
    - [x] HTTP-Pull based binary stream for zero-downtime remote firmware updates (with MQTT RPC Trigger).
    - [x] Dynamic NVS (Non-Volatile Storage) config updates via Retained MQTT RPCs.
    - [ ] Hardware RTC Deep Sleep Scheduling (Command ESP32 to sleep to save battery, wake on schedule).
- [ ] **10.2: Remote Access & Discovery**
    - [x] WebRTC P2P Hole-punching (Zero-config, zero-port-forwarding, via free Render signaling server).
    - [ ] Fallback support for Tailscale / Custom Signaling Servers for advanced Homelab users.
	- [x] mDNS (`xomoi.local`) zero-config auto-discovery for local network UX.

## PHASE 11: FEDERATION & MESH NETWORKING
- [ ] **11.1: The Immortal Gossip Mesh**
    - [ ] Implement `hashicorp/memberlist` for O(log N) decentralized epidemic routing.
    - [ ] Build CRDTs (Conflict-Free Replicated Data Types) for Split-Brain immunity during network partitions.
    - [ ] Cross-node Mochi-MQTT bridging for seamless global telemetry aggregation.

## PHASE 12: ENTERPRISE SECURITY & RBAC
- [ ] **12.1: Granular Role-Based Access Control**
    - [ ] Implement Owner, Editor, and Viewer roles for device sharing.
    - [ ] Map Roles dynamically to Mochi-MQTT ACLs (e.g., Viewers can subscribe, but cannot publish RPC configs).

## PHASE 13: SECURE EDGE ORCHESTRATION (SILO INTEGRATION)
- [ ] **13.1: The Silo Package Refactor**
    - [ ] Convert `silo-core` into a highly reusable Go library.
    - [ ] Implement Cgroups v2 support for CPU and Memory limits.
- [ ] **13.2: Xomoi Plugin Engine**
    - [ ] Build the OTA receiver to download and unpack `tar.gz` RootFS payloads.
    - [ ] Orchestrate Silo namespaces from within Xomoi to run untrusted AI/Edge workloads securely.
