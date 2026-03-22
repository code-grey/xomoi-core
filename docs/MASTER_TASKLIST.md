# XOMOI-CORE v1.0 MASTER TASKLIST (5-MONTH SOVEREIGN ROADMAP)

## PHASE 1: THE SOVEREIGN CONTRACT (WEEKS 1-2)
- [ ] **1.1: Modular Protobuf "Core + Plugins" (`proto/v1/`)**
    - [ ] `common.proto`: Shared enums, Vector3, and Field ID locks.
    - [ ] `telemetry_base.proto`: Standard sensor types (Float, Int, Bool).
    - [ ] `telemetry_pro.proto`: Advanced types (JSON, Bytes) for AI/Imaging.
    - [ ] `registry.proto`: Tag-to-Name handshake mapping.
    - [ ] `command.proto`: Bidirectional control (Toggle, Value, Text).
- [ ] **1.2: Core Domain Models (`internal/core/`)**
    - [ ] `models.go`: Define `User`, `Device`, `Session`, `SensorTag`, and `AlertRule` Go structs.
- [ ] **1.3: Repository Interfaces (`internal/repository/`)**
    - [ ] `interfaces.go`: Define pure interfaces for User, Session, Device, and Telemetry.

## PHASE 2: STORAGE & MEMORY ENGINE (WEEKS 3-4)
- [ ] **2.1: SQLite Engine (`internal/repository/sqlite/`)**
    - [ ] `db.go`: Initialize `sql.DB` with WAL mode and `PRAGMA busy_timeout`.
    - [ ] `migrations/`: Create embedded SQL files for all core tables.
- [ ] **2.2: Volatile State Manager (`internal/state/`)**
    - [ ] `hot_state.go`: Implement `sync.Map` for O(1) real-time sensor status.
    - [ ] `snapshot.go`: Implement the 5-minute bulk-flush routine to SQLite.

## PHASE 3: THE EMBEDDED BROKER (WEEKS 5-6)
- [ ] **3.1: Mochi-MQTT Integration (`internal/broker/`)**
    - [ ] `mochi.go`: Initialize embedded server (TCP + WebSockets).
    - [ ] `auth.go`: Implement **HMAC-Lite Hook** (Verify MAC + Timestamp + Signature).
- [ ] **3.2: The Ingestion Pipeline**
    - [ ] `worker_pool.go`: Fixed-size Worker Pool for message processing.
    - [ ] `processor.go`: OnPublish -> Proto Unmarshal -> Worker Channel (Backpressure enabled).

## PHASE 4: SOVEREIGN API & SECURITY (WEEKS 7-8)
- [ ] **4.1: The Web Server (`internal/api/`)**
    - [ ] `router.go`: Go 1.26 `net/http` ServeMux initialization.
    - [ ] `middleware/`: Stateful Session check and Anti-CSRF.
- [ ] **4.2: Auth & Session Endpoints**
    - [ ] `POST /api/v1/auth/login`: Argon2ID verify + Session create.
    - [ ] `POST /api/v1/auth/logout`: Session deletion.

## PHASE 5: THE GRAND ARCHITECT (WEEKS 9-11)
- [ ] **5.1: Xomoi-Transpiler (`xomoi-ctl`)**
    - [ ] `parser`: Go-based Protobuf parser to read `proto/v1/`.
    - [ ] `generator/cpp`: Adaptive pruning logic to generate "Lite" C++ SDK headers.
    - [ ] `generator/svelte`: Metadata generation for UI widgets and icons.
- [ ] **5.2: Xomoi-Claim Flow**
    - [ ] `GET /api/v1/devices/discover`: Scan for `Xomoi-Claim-XXXX` signals.
    - [ ] `POST /api/v1/devices/claim`: Generate HMAC-Lite token and push to device.

## PHASE 6: THE SOVEREIGN DASHBOARD (WEEKS 12-14)
- [ ] **6.1: Svelte 5 Frontend (`ui/`)**
    - [ ] `realtime/`: SSE listener for hot state.
    - [ ] `charts/`: Native SVG/Canvas visualization (Zero-heavy-lib).
- [ ] **6.2: Low-Code Onboarding**
    - [ ] Web-Flasher: WebSerial/WebUSB integration for one-click sensor flashing.

## PHASE 7: THE BLACKSMITH SDK (WEEKS 15-16)
- [ ] **7.1: Xomoi C++ SDK (`sdk/`)**
    - [ ] `XomoiCore.h`: Static memory MQTT/NanoPB client.
    - [ ] Template Library: JSON blueprints for DHT, BME, PIR, and GPIO.

## PHASE 8: ALERTS & ENTERPRISE BRIDGE (WEEKS 17-18)
- [ ] **8.1: Alert Engine (`internal/alerts/`)**
    - [ ] `evaluator.go`: Match hot state against user-defined threshold rules.
- [ ] **8.2: Upstream Worker (`internal/bridge/`)**
    - [ ] Store-and-Forward logic for off-grid Enterprise syncing.

## PHASE 9: THE JANITOR & HARDENING (WEEKS 19-20)
- [ ] **9.1: Background Janitor (`internal/worker/`)**
    - [ ] 1m/5m/1h telemetry aggregation and raw data pruning.
- [ ] **9.2: Final Polish**
    - [ ] Security Audit, Binary Optimization (`ldflags`), and v1.0 Release.
