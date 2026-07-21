# Xomoi-Core: Sovereign Edge Node

Xomoi-Core is a hyper-converged, single-binary edge node built entirely in Go. It is designed to act as a sovereign replacement for heavy, cloud-tethered IoT stacks (like KubeEdge + Mosquitto + InfluxDB) on severely resource-constrained hardware.

At a mere 15MB binary size, it orchestrates embedded Time-Series databases, high-speed telemetry ingestion pipelines, and isolated container executions without a heavy control plane or Docker daemon.

## Architecture

Xomoi is built for the extreme edge (e.g., Raspberry Pi 3/4, Industrial PCs) where memory is scarce, SD card IO is fragile, and network connectivity is highly intermittent.

### 1. The Lossless Ingestion Pipeline (360k+ Msg/Sec)
Instead of relying on external broker architectures, Xomoi embeds Mochi-MQTT directly into the Go binary. 
* Incoming MQTT packets are instantly parsed into a Zero-Allocation sync.Map utilizing a 256-shard FNV-1a hash algorithm to eliminate Mutex lock contention. 
* Packets are funneled into a high-speed Go Channel Ring Buffer to decouple fast network reads from slow disk IO.

### 2. The Native Zstd TSDB
Traditional databases chew through SD card lifespan. Xomoi intercepts all telemetry, natively compresses the JSON payloads using the Zstandard (Zstd) algorithm in memory, generates mathematically sortable ULIDs, and performs ultra-fast BulkInsert transactions into a WAL-enabled SQLite Time-Series Database. This reduces disk footprint by 80% and ensures zero data loss during high-frequency ingestion spikes.

### 3. telemetry_pro.proto (The Enterprise Escape Hatch)
Xomoi utilizes NanoPB Protobufs. For standard sensors (temperature, speed), it uses lightweight FLOAT/BOOL primitives. 
For complex enterprise arrays (e.g., a 1024-dimensional LiDAR point cloud, or an AI CV tensor), it uses the telemetry_pro.proto schema with the BYTES_RAW field. Xomoi acts as a lossless conduit, bypassing JSON decoding entirely to compress and store massive multidimensional byte matrices.

## Architectural Comparison

| Feature | Xomoi-Core | KubeEdge + Docker | Eclipse Mosquitto |
| :--- | :--- | :--- | :--- |
| Footprint | ~15MB (Single Binary) | > 1GB (Daemon + K3s) | ~5MB (Just the Broker) |
| Telemetry Storage | Embedded Zstd SQLite | Requires external DB | None |
| Ingestion Engine | Lossless Ring Buffer | N/A | Pass-through only |
| Web Dashboard | Embedded (Svelte 5) | None | None |
| Container Isolation | Silo (Linux Namespaces) | Docker / Containerd | None |
| Target Hardware | RPis, ARM SBCs | Multi-core Industrial Gateways | ESP32, Routers |
| Cloud Clustering | Store-and-Forward (Phase 2.5)| Natively supported | Requires bridging |

> Note: While Xomoi is exponentially more efficient on single-node hardware, KubeEdge remains superior for distributed, multi-node Kubernetes clustering. Xomoi is strictly an edge-first, offline-first sovereign node.

## Future Vision

We are actively hardening the project towards commercial deployment capabilities.

* Phase 10.1 (Hardware RTC Deep Sleep Scheduling): Integration with ESP32 Deep Sleep cycles, allowing the edge node to act as a buffer while low-power satellite endpoints go offline for days to preserve battery.
* Phase 11.1 (Immortal Gossip Mesh): Distributed High-Availability across multiple Xomoi nodes on a local LAN using UDP-based gossip protocols.
* Phase 13 (Silo Container Orchestration): Utilizing our custom silo-core engine (leveraging pure Linux cgroups and namespaces) to allow Xomoi to spin up untrusted 3rd-party AI plugins (like Python computer vision scripts) inside strict, memory-limited rootfs environments without installing Docker.

## License
This software is licensed under the GNU Affero General Public License v3.0 (AGPLv3). It is built to be sovereign. If you modify this software and run it as a cloud service over a network, you must open-source your modifications.
