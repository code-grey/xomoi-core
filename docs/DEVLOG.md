# Xomoi-Core DevLogs

This document serves as a chronological record of the major architectural decisions and milestones achieved during the development of Xomoi-Core.

---

## 📅 DevLog: July 7-8, 2026
**Phases Completed:** Phase 6.1 (UI), 6.2 (Web-Flasher), 7.1 (SDK Foundation)
**Authors:** Adrish Bora (@code-grey) & Antigravity AI Architect

### 1. The Death of Flutter, The Rise of Svelte 5
We officially killed the Flutter Web Dashboard from the original spec. Flutter Web's canvas rendering is too heavy and SEO-hostile. We migrated the entire dashboard to **Svelte 5**. 
* **The Routing Masterclass:** We bypassed heavy SPA routers (like `svelte-routing` or `page.js`) and built a zero-dependency router that relies purely on native `window.location.hash`. Svelte 5's `$effect` runway reacts instantly to the hash changes, keeping the browser history intact without adding a single byte of dependencies.
* **SVG Telemetry:** We implemented raw SVG polylines for the real-time telemetry charts. This is computationally lighter than canvas-based libraries like Chart.js and allows us to use CSS variables (`--accent-cyan`) to create stunning, glowing UI aesthetics.

### 2. The "Golden Path" Web-Flasher (Phase 6.2)
To achieve the "3-lines of code / Grandma can use it" UX, we implemented a pure-browser Web-Flasher. 
* We installed Espressif's official `esptool-js`.
* We wired up the Chrome/Edge **WebSerial API** inside `ui/src/lib/WebFlasher.svelte`. 
* Result: A user can plug an ESP32 into their laptop, click a button on the Svelte dashboard, bypass the OS entirely, and directly negotiate with the ESP32 bootloader to read the MAC address and flash generic Xomoi firmware. Zero CLI tools required.

### 3. The Discovery Protobuf & NanoPB Engine (Phase 7.1)
We tackled the massive memory-bloat issue that plagues systems like Home Assistant (which forces microcontrollers to store and transmit gigantic JSON strings for sensor auto-discovery).
* **The Solution:** We designed a `Discovery` Protobuf payload inside `xomoi.proto`. 
* **The Float Standard:** We abstracted all sensor data (`DataPoint`) to generic floats (e.g., Motion = 1.0/0.0). The Discovery payload dictates exactly how the UI should render it (e.g., `FLOAT_METRIC` vs `BOOL_STATE`). This means you can invent a new sensor on the ESP32, and the Svelte UI will instantly build a chart for it without updating the frontend codebase.
* **Zero Fragmentation:** We compiled the Protobuf schema into C++ headers (`xomoi.pb.c`) using NanoPB. Crucially, we enforced strict string limits via `xomoi.options` (e.g., `SensorConfig.display_name max_size:32`). This guarantees the C++ compiler generates safe, static `char` arrays, completely eliminating the threat of heap fragmentation on the ESP32.

### 4. Hexagonal "Xomoi-Enterprise" Validation
We validated that our strict adherence to the Hexagonal Repository Pattern (`internal/repository/interfaces.go`) means Xomoi can instantly scale to enterprise levels. If a user outgrows SQLite and Mochi-MQTT, they can simply write a `timescaledb.go` adapter and deploy Xomoi on an AWS cluster without touching the core ingestion engine. We updated the `README.md` to reflect this massive selling point.

**Next Up:** Phase 7.2 (Writing the actual C++ mbedTLS Crypto & Protobuf streaming logic for the Blacksmith SDK).
