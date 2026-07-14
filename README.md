# Xomoi-Core

**The Open-Source, Single-Binary Assassin to Blynk.**

Xomoi is an industrial-grade, sovereign IoT telemetry engine. It bridges the massive gap between the "3-lines of code" simplicity of commercial platforms (like Blynk) and the grueling 15-hour DevOps nightmare of setting up your own open-source stack (Mosquitto, InfluxDB, Grafana, Docker).

## Why Xomoi Exists
If you are a maker, a hacker, or an indie hardware startup trying to build a custom fleet of sensors, you currently have three bad choices:
1. **Home Assistant:** Built for consumer smart homes (Philips Hue, Ring). It is a heavy Python monolith that chokes on high-throughput, fleet-scale raw telemetry.
2. **ThingsBoard / ChirpStack:** Enterprise-grade, but requires Java Virtual Machines (JVMs), PostgreSQL clusters, Cassandra, and 4GB of RAM just to boot up.
3. **Blynk (IoT):** A closed-source cloud platform. You pay monthly subscriptions, they strictly limit your devices, and you do not own your data.

**Xomoi is the "SQLite of IoT Platforms".**
Xomoi is a single 15MB Go executable. You drop it onto a $5 Raspberry Pi Zero, double-click it, and it instantly spins up an embedded Mochi-MQTT Broker, a SQLite (WAL) time-series database, and a stunning Svelte 5 Web Dashboard on port 8085. Zero dependencies. Zero configuration.

## Core Architectural Pillars
* **The "Golden Path" Web-Flasher:** Flash generic firmware directly to ESP32s from your browser using the built-in WebSerial API (`esptool-js`). No command line required.
* **The "Blacksmith SDK":** For advanced engineers, a NanoPB-backed C++ SDK to write custom firmware for any esoteric microcontroller on the market.
* **Protobuf Auto-Discovery:** Devices blast a tiny Protobuf Discovery struct over MQTT on boot. The Svelte UI parses this and instantly auto-generates Line Charts, Status Boxes, and Alert Rules—zero JSON memory bloat, zero manual UI configuration.
* **Zero-Allocation Routing:** The backend uses `vtprotobuf` (Go) and the C++ SDK uses `nanopb`. This prevents heap fragmentation and guarantees your ESP32s will never crash from memory leaks.
* **Hexagonal Disaster Recovery:** Built-in Background Janitor and automatic SQLite snapshot uploads to free webhooks (like Discord) to prevent SD card exhaustion.
* **Xomoi-Enterprise Ready:** Because we strictly enforce the Hexagonal Repository Pattern, you can swap out SQLite and Mochi-MQTT for TimescaleDB and EMQX with just a few lines of code to run Xomoi on a massive cloud VPS.
* **HMAC-Lite Security:** Bypasses the massive RAM overhead of mTLS certificates on edge microcontrollers. Devices authenticate using lightweight SHA-256 cryptographic signatures natively accelerated by ESP32 silicon.
* **Zero-Dependency SPA Routing:** The Svelte 5 dashboard uses native browser Hash state (`window.location.hash`) for mathematical perfection, avoiding massive third-party router libraries.

## The "Holy Grail" Networking (WebRTC)
No Port Forwarding. No VPN Apps. No Cloudflare Accounts. 
Xomoi utilizes **WebRTC Peer-to-Peer Data Channels** to punch through your home router's NAT firewall. When you open the mobile app at a coffee shop, a free, microscopic signaling server introduces your phone to your Raspberry Pi, then steps out of the way. The data streams end-to-end encrypted directly from your house to your phone. 

## Pros & Cons
### Pros
- **100% Free & Sovereign:** You own the data. No subscriptions, no cloud vendor lock-in.
- **Microscopic Footprint:** Runs flawlessly on any hardware (from a Pi Zero to a 64-core server).
- **Extreme Throughput:** Proven to handle over 128,000 telemetry messages per second on a single CPU core.
- **Maker Simplicity:** The C++ SDK feels exactly like Blynk (`node.addSensor(...)`).
- **Zero-Config Remote Access:** WebRTC P2P eliminates the need for DuckDNS or Port Forwarding.

### Cons
- **Not a Smart Home Hub:** Xomoi does not integrate with proprietary consumer brands (Alexa, Ring) out of the box. It is designed for custom-built hardware telemetry.
- **Requires Arduino Knowledge (For Custom Chips):** While we provide the Web-Flasher for standard ESP32s, getting the absolute most out of Xomoi requires basic C++ or PlatformIO knowledge.

## Getting Started (Quick Start)

Xomoi-Core is designed to run instantly without complex Docker setups or external databases.

### 1. Running the Broker & Dashboard
1. Clone the repository: `git clone https://github.com/code-grey/xomoi-core.git`
2. Install [Task](https://taskfile.dev/), Go 1.26+, and Node.js.
3. Build the Svelte UI: `task ui:build`
4. Build the core binary: `task build`
5. Run the executable: `./build/xomoi` (or `.\build\xomoi.exe` on Windows).

The Xomoi Engine is now running locally. 
- The embedded Mochi-MQTT Broker is active on Port 1883.
- The Svelte 5 Dashboard is live at `http://localhost:8085`.

### 2. Provisioning Devices (The Golden Path)
If you are using standard ESP32s, you do not need to write C++ code:
1. Open the Dashboard at `http://localhost:8085`.
2. Navigate to the **Device Fleet** tab and click **Provision New Device**.
3. Plug your ESP32 into your computer via USB.
4. Click **Connect via WebSerial** to flash the generic firmware directly from your browser.
5. The device will automatically connect to your Wi-Fi and appear on the dashboard.

### 3. Writing Custom Firmware (The Blacksmith Path)
If you are building custom hardware, you can use our zero-allocation C++ SDK:
1. Generate the NanoPB headers: `task proto:sdk` (requires Python).
2. Include the SDK in your PlatformIO or Arduino project (located in `sdk/cpp`).
3. Call `node.addSensor(...)` and send the telemetry!

## Roadmap Status
We are actively building V1.0. Check `docs/MASTER_TASKLIST.md` for current progress.
