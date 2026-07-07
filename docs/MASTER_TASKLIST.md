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
- [ ] **7.2: Xomoi C++ Engine Implementation**
    - [ ] Hardware SHA-256 HMAC-Lite Cryptography via `mbedtls`.
    - [ ] Implement Protobuf byte streaming and Discovery callbacks in `XomoiCore.cpp`.

## PHASE 8: ALERTS & HEXAGONAL BACKUP (WEEKS 17-18)
- [x] **8.1: Alert Engine (`internal/alerts/`)**
    - [x] `evaluator.go`: Match hot state against user-defined threshold rules.
- [x] **8.2: Hexagonal Backup Engine (`internal/backup/`)**
    - [x] `BackupProvider` Interface: Abstract `Save()` and `Restore()`.
    - [x] Implement `DiscordPreserver` (Webhook attachments).
    - [x] Implement `DrivePreserver` (Google Drive Service Account).
    - [x] Implement `PostgresPreserver` (Cloud syncing).

## PHASE 9: THE JANITOR & HARDENING (WEEKS 19-20)
- [x] **9.1: Background Janitor (`internal/worker/`)**
    - [x] 1m/5m/1h telemetry aggregation and raw data pruning for threshold control.
- [x] **9.2: Final Polish**
    - [x] Graceful Shutdown (force SQLite flush + Backup on SIGTERM).
    - [x] Security Audit, Binary Optimization (`ldflags`), and v1.0 Release.

## PHASE 9.5: THE TERMINAL DASHBOARD (TUI)
- [x] **9.5.1: Xomoi-CLI (`cmd/xomoi-cli/`)**
    - [x] `charmbracelet/bubbletea` integration for a zero-bloat standalone terminal UI.
    - [x] Matrix-style real-time Telemetry, Worker Pool health, and Active Claims viewing.

## PHASE 10: REMOTE OPERATIONS & OTA (WEEKS 21-22)
- [ ] **10.1: OTA (Over-The-Air) Engine**
    - [ ] MQTT-based binary stream for zero-downtime remote firmware updates.
    - [ ] Dynamic NVS (Non-Volatile Storage) config updates to change ping frequency without flashing.
- [ ] **10.2: Remote Access & Discovery**
    - [ ] WebRTC P2P Hole-punching (Zero-config, zero-port-forwarding, via free Render signaling server).
    - [ ] Fallback support for Tailscale / Custom Signaling Servers for advanced Homelab users.
	- [ ] mDNS (`xomoi.local`) zero-config auto-discovery for local network UX.

## PHASE 11: FEDERATION (MESH NETWORKING)
- [ ] **11.1: Node-to-Node Bridging**
    - [ ] Mochi-MQTT Bridge configuration to forward `/xomoi/+/telemetry` to a Primary Node.
    - [ ] Auto-discovery of local Satellite nodes.
- [ ] **11.2: Unified Federation Dashboard**
    - [ ] Display aggregate fleet health and sensor data from all Satellite nodes in one single UI.
