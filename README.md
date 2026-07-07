# Xomoi-Core 

**The Open-Source, Single-Binary Assassin to Blynk.**

Xomoi is an industrial-grade, sovereign IoT telemetry engine. It bridges the massive gap between the "3-lines of code" simplicity of commercial platforms (like Blynk) and the grueling 15-hour DevOps nightmare of setting up your own open-source stack (Mosquitto, InfluxDB, Grafana, Docker).

## Why Xomoi Exists

If you are a maker, a hacker, or an indie hardware startup trying to build a custom fleet of sensors, you currently have three bad choices:
1. **Home Assistant:** Built for consumer smart homes (Philips Hue, Ring). It is a heavy Python monolith that chokes on high-throughput, fleet-scale raw telemetry.
2. **ThingsBoard / ChirpStack:** Enterprise-grade, but requires Java Virtual Machines (JVMs), PostgreSQL clusters, Cassandra, and 4GB of RAM just to boot up.
3. **Blynk (IoT):** A closed-source cloud platform. You pay monthly subscriptions, they strictly limit your devices, and you do not own your data.

**Xomoi is the "SQLite of IoT Platforms".**
Xomoi is a single 15MB Go executable. You drop it onto a $5 Raspberry Pi Zero in the middle of a forest, double-click it, and it instantly spins up an embedded Mochi-MQTT Broker, a SQLite (WAL) time-series database, and a stunning Flutter Web Dashboard on port 8080. Zero dependencies. Zero configuration.

## The Architecture

* **Zero-Allocation Routing:** The backend uses `vtprotobuf` (Go) and the C++ SDK uses `nanopb`. This prevents heap fragmentation and guarantees your ESP32s will never crash from memory leaks.
* **HMAC-Lite Security:** Bypasses the massive RAM overhead of mTLS certificates on edge microcontrollers. Devices authenticate using lightweight cryptographic signatures.
* **Context-Switching Immune:** The Go ingestion pipeline dynamically binds to your exact CPU core count (via `runtime.NumCPU()`).
* **Hexagonal Disaster Recovery:** Built-in Background Janitor and automatic SQLite snapshot uploads to free webhooks (like Discord) to prevent SD card exhaustion.

## The "Holy Grail" Networking (WebRTC)

No Port Forwarding. No VPN Apps. No Cloudflare Accounts. 
Xomoi utilizes **WebRTC Peer-to-Peer Data Channels** to punch through your home router's NAT firewall. When you open the mobile app at a coffee shop, a free, microscopic signaling server introduces your phone to your Raspberry Pi, then steps out of the way. The data streams end-to-end encrypted directly from your house to your phone. 

*(Note: For the ~10% of users behind highly restrictive corporate "Symmetric NATs" where UDP hole-punching fails, Xomoi provides 1-click fallback integrations for Tailscale and Cloudflare Tunnels).*

## Pros & Cons

### Pros
- **100% Free & Sovereign:** You own the data. No subscriptions, no cloud vendor lock-in.
- **Microscopic Footprint:** Runs flawlessly on any hardware (from a Pi Zero to a 64-core server).
- **Maker Simplicity:** The C++ SDK feels exactly like Blynk (`xomoi.sendTemperature(25.4)`).
- **Zero-Config Remote Access:** WebRTC P2P eliminates the need for DuckDNS or Port Forwarding.

### Cons
- **Not a Smart Home Hub:** Xomoi does not integrate with proprietary consumer brands (Alexa, Ring) out of the box. It is designed for custom-built hardware telemetry.
- **Requires Arduino Knowledge:** While we are building Web-Flashers for no-code users, getting the absolute most out of Xomoi requires basic C++ or Python knowledge.

## Roadmap Status
We are actively building V1.0. Check `docs/MASTER_TASKLIST.md` for current progress.
