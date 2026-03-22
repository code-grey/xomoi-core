# XOMOI PROJECT RULEBOOK: MANDATORY DIRECTIVES

## CORE PHILOSOPHY (XOMOI-CORE)
Xomoi-Core is a sovereign, open-source edge node dedicated to **Digital Freedom**. It is built to liberate users from predatory proprietary giants who harvest and sell personal data. 

- **The Sovereign Monolith:** One binary. One process. Zero external dependencies. Everything (Broker, DB, UI) is `go:embed`ed.
- **Privacy First:** Zero telemetry. No "phone-home" logic. All data stays local by default.
- **Autonomous & Resilient:** Designed to be self-sufficient on low-power hardware.
- **DePIN Isolation:** Blockchain-specific logic is strictly a secondary, optional extension.
- **Efficiency:** Resource-conscious architecture ensures longevity on the edge.

## 1. THE ANTI-TRAPS (THE RED LINES)
- **The SD-Card Killer:** Never write to SQLite on every sensor update. High-frequency data *must* live in volatile memory (`sync.Map`) and only snapshot to disk in bulk.
- **Microservice Sprawl:** Xomoi must never be split into multiple binaries. If a feature requires a separate process, it doesn't belong in Xomoi-Core.
- **The Dependency Leak:** No Node.js, Python, or heavy C++ runtimes allowed in the production edge binary. 
- **The Configuration Friction:** No "Required" external YAML/JSON files. Xomoi must be able to boot with zero config and provide a "Local-First" setup experience.

## 2. BACKEND ARCHITECTURE (THE GHOST)
- **Zero-Dependency Mandate:** Never use 3rd-party routers (Gin/Fiber). Use Go 1.26 `net/http` ServeMux.
- **Zero Goroutine Sprawl:** Never spawn unmanaged goroutines in the ingestion path. All MQTT `OnPublish` events must flow into a fixed-size **Worker Pool** via a buffered channel.
- **Explicit Backpressure:** If the ingestion channel is full, the system must drop packets (QoS 0) or delay ACKs (QoS 1) to prevent memory exhaustion (OOM).
- **Repository Pattern:** All data access must be defined as an interface. The business logic must not know it is talking to SQLite. This allows seamless future transitions to Postgres or TimescaleDB for enterprise versions.
- **Strict Interface Purity:** Database-specific syntax (e.g., SQLite `strftime` or `json_extract`) must be contained entirely within the repository implementation. Business logic only receives and returns clean Go structs.
- **Hot-Path Memory Policy:** Use `sync.Pool` for high-frequency objects (Protobuf packets). Aim for **zero heap allocations** in the telemetry ingestion path.

## 2. PERSISTENCE & THE JANITOR (THE GHOST)
- **Volatile vs. Persistent:** High-frequency status updates (last_seen, current values) live in-memory (`sync.Map`). Telemetry is batched and flushed to SQLite WAL every 5 minutes or on shutdown.
- **Flash Memory Protection:** Minimize SQLite writes. Never write a single sensor reading synchronously.
- **The Janitor Service:** Implement a background "Janitor" for:
    - **Aggregation:** Bucketing raw telemetry into 1m, 5m, and 1h summary tables for O(1) charting performance.
    - **Pruning:** Deleting raw data older than $X$ days to preserve disk space.
    - **Optimization:** Running `PRAGMA optimize` and managing WAL checkpoints.
- **Circular Logging:** System logs must be stored in a fixed-size SQLite table (e.g., 10,000 rows). New logs overwrite the oldest entries to prevent disk-full crashes.
- **I/O Watchdog:** Monitor write latency. If I/O wait exceeds safe thresholds (e.g., failing SD card), trigger **Read-Only Safe Mode** to prevent binary deadlock and allow user notification.

## 3. PROTOCOL & SENSOR ECOSYSTEM (THE BLACKSMITH)
- **Universal Tag-Based Proto:** Telemetry uses `uint32` tags for efficiency. Registry (mapping tags to strings) is a low-frequency handshake.
- **Absolute Sensor Support:** Support any number of identical or diverse sensors via a **Component-ID** model.
- **Firmware Proxy (OTA):** Xomoi acts as a local OTA update server. Sensors pull `.bin` firmware updates directly from the Xomoi node over the local network via MQTT.
- **Static Edge Memory:** The C++ SDK must use static buffers and `PROGMEM` (Flash) for sensor metadata. **No Malloc** allowed to prevent heap fragmentation.

## 4. NETWORKING & PROVISIONING (THE SHRINK & THE GHOST)
- **mDNS Discovery:** Implement Multicast DNS (Bonjour). The node must be reachable via `xomoi.local` without the user knowing the IP.
- **Air-Gap Time Sync:** If no NTP is available, the UI must "Push" the browser's current Unix time to the node on first login to sync the system clock.
- **Captive Portal:** If no Wi-Fi is found, Xomoi must enter **AP-Fallback Mode**, serving a captive portal for Wi-Fi and Admin credentials setup.
- **MQTT Rate Limiting:** Implement broker-level throttling per `device_id` to prevent malfunctioning or hostile sensors from triggering an OOM kill.

## 5. SECURITY & ENTERPRISE SCALE (THE WARDEN)
- **HMAC-Lite:** Replace heavy mTLS/JWT with light HMAC-based signatures for edge device authentication to bypass heavy SSL buffers.
- **Encrypted Storage:** Use **SQLCipher** or a Master Key encryption strategy to ensure SQLite data cannot be extracted from a stolen SD card.
- **Hybrid-Sync (The Bridge):** Implement an optional upstream worker for forwarding local data to a central Enterprise Postgres instance (Store-and-Forward architecture).
- **Isolated DePIN Logic:** All blockchain-specific logic (Merkle trees, on-chain proof) must be isolated behind the `depin` build tag.

## 6. FRONTEND UX (THE SHRINK)
- **Embedded & Static:** UI compiled to static HTML/JS and embedded into Go via `go:embed`. No Node.js at runtime.
- **Native Rendering:** No heavy JS charting libraries (Chart.js/Recharts). Render telemetry using native SVG, Canvas, or D3 primitives for legacy hardware.
- **State via SSE/WS:** Use Mochi-MQTT's internal WebSockets for live status. Fallback to Server-Sent Events (SSE) if necessary. Avoid polling.
