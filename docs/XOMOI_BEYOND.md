# XOMOI: BEYOND (THE SOVEREIGN COMMUNICATIONS FORK)

While `Xomoi-Core` focuses on high-speed telemetry and edge orchestration (TCP/Wi-Fi), this document drafts the architectural vision for a sovereign, decentralized communications fork (e.g., `Xomoi-Mesh` or `Xomoi-Chat`).

The philosophy is absolute independence from ISPs, Cloud Providers, and Cellular Towers.

## 1. THE TRANSPORT PROTOCOLS (Bypassing the Internet)
To be truly sovereign, a node must fall back to physical layers of communication when TCP/IP (the standard internet) is compromised or unavailable.

### A. LoRa (Long Range Sub-GHz Radio)
- **The Tech:** 433/868/915 MHz unregulated radio frequencies.
- **The Range:** 10 to 15+ kilometers in urban environments.
- **The Use Case:** The ultimate "City-Wide Mesh". A Xomoi node with a $15 LoRa module can broadcast heavily encrypted (AES-256) 256-byte chat messages across an entire city without a single cell tower.

### B. ESP-NOW (Connectionless Wi-Fi)
- **The Tech:** A protocol developed by Espressif that bypasses standard Wi-Fi routers.
- **The Range:** ~100 to 200 meters.
- **The Use Case:** Devices talk directly to each other via MAC addresses in a fraction of a millisecond. If the home router dies, the ESP32s automatically switch to ESP-NOW and form an unbreakable local mesh.

### C. BLE Mesh (Bluetooth Low Energy)
- **The Tech:** Device-to-device hopping via Bluetooth.
- **The Range:** ~10 to 50 meters per hop.
- **The Use Case:** "The Backpack Protocol". Your phone connects to a node via BLE, downloads encrypted messages intended for someone else, and silently passes them to other nodes as you walk through a city (Sneakernet).

### D. Audio-FSK (Ultrasonic Data Transfer)
- **The Tech:** Frequency-Shift Keying using soundwaves.
- **The Range:** Within a room.
- **The Use Case:** Air-gapped key exchange. Two devices that have their radios physically disabled can still exchange cryptographic keys using ultrasonic chirps (inaudible to human ears) via their microphones and speakers.

## 2. THE ENGINEERING PHILOSOPHY (The Adapter Pattern)
We will **not** rewrite physical layer protocols from scratch. That is a waste of engineering time. If Meshtastic and ESP-NOW already exist, we take the crust and bake our own toppings.

* **The Transport Interface:** The Xomoi C++ SDK and Go Backend will implement a universal `Transport` interface. 
* **Meshtastic Integration:** Instead of writing LoRa routing algorithms, a Xomoi Node will simply connect to a standard Meshtastic radio via USB/Serial. Meshtastic handles the 15km RF hops; Xomoi handles the SQLite storage, CRDT syncing, and the Svelte UI.
* **ESP-IDF Integration:** ESP-NOW is built natively into the Espressif IDF. Our C++ SDK will simply wrap our `xomoi.proto` binary payloads inside standard `esp_now_send()` calls.

## 3. THE SOFTWARE ARCHITECTURE
- **The Gossip Protocol:** Replacing standard MQTT routing with a CRDT (Conflict-Free Replicated Data Type) or `hashicorp/memberlist` engine. Nodes do not have "Clients"; they are all equal Peers.
- **Store-and-Forward SQLite:** The SQLite database transforms from a TSDB (Time-Series) into an Encrypted Message Spool. It holds packets indefinitely until the target node physically comes into radio range, then instantly flushes the queue.
- **Zero-Knowledge by Default:** The Go backend cannot read *any* payload. The C++ or Svelte SDK must encrypt the payload with X25519 (ECDH) and AES-256-GCM before the bytes ever hit the network layer.
