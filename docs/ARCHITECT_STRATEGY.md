# Xomoi-Core: Architecture and Strategy for Sovereign Edge Computing

## Introduction

Xomoi-Core is an open-source, single-binary edge node designed to provide digital sovereignty and data freedom. It functions as a private, air-gapped IoT infrastructure that eliminates reliance on proprietary cloud platforms. This document outlines the hybrid architecture that allows Xomoi to scale from a single residential node to a high-security, off-grid Enterprise/Safe-House deployment.

## The Dual-Tier Architecture

Xomoi is architected to operate in two distinct modes without altering its core business logic, achieved through strict adherence to the Repository Pattern.

### 1. Xomoi-Core (Sovereign Local Node)
Designed for the individual user and the high-privacy residential environment.
- **Runtime:** Single Go 1.26 binary.
- **Broker:** Embedded Mochi-MQTT.
- **Persistence:** Embedded SQLite (WAL Mode) with Flash-Wear protection.
- **Hot-State:** In-Memory sync.Maps (Zero-latency status tracking).
- **Frontend:** Static Svelte 5 embedded via `go:embed`.
- **Primary Goal:** Local autonomy, zero telemetry, and hardware longevity on SD cards/eMMC.

### 2. Xomoi-Enterprise (Air-Gapped & Safe-House Grade)
Designed for high-security facilities, off-grid safe houses, and large-scale industrial deployments where data must be aggregated across multiple nodes or stored for long-term forensic analysis.
- **Runtime:** Orchestrated Go binaries (Xomoi-Core instances) acting as edge collectors.
- **Persistence Pivot:** Seamless transition from SQLite to **PostgreSQL/TimescaleDB** using the existing Repository interfaces.
- **Broker Scaling:** Transition to a clustered EMQX or RabbitMQ environment for high-availability MQTT.
- **State Management:** Optional Redis integration for distributed state across a cluster.
- **Security:** Enhanced HMAC-Lite with hardware-backed security modules (HSM) or TPM integration.

## Strategic Technical Stack

Xomoi utilizes a specialized stack chosen for its ability to operate in resource-constrained and network-isolated environments:

| Component | Technology | Rationale |
| :--- | :--- | :--- |
| **Language** | Go 1.26 | High-concurrency, static linking, and memory safety. |
| **Edge Storage** | SQLite (JSON Column) | Serverless, zero-config. JSON column slashes writes by 80% using batched data. |
| **Enterprise Storage** | PostgreSQL | Relational integrity and enterprise-grade backup/clustering. |
| **Telemetry Format** | NanoPB (Protobuf) | Wire-size efficiency, zero-allocation parsing, natively supports batched arrays. |
| **Identity/Auth** | HMAC-Lite | Bypasses heavy SSL/JWT buffers on 8-bit/32-bit microcontrollers. |
| **Real-time UI** | Svelte 5 + SSE | Low-power client-side rendering with no Node.js runtime required. |
| **Discovery** | mDNS (Bonjour) | Zero-configuration local network reachability (`xomoi.local`). |

## The Xomoi-Transpiler (`xomoi-ctl`)

To maintain our "Zero-Dependency" and "Static-Memory" mandates, Xomoi uses a custom Go-based orchestrator called `xomoi-ctl` instead of a standard raw Protobuf compiler.

- **SSOT (Single Source of Truth):** The master Protobuf files in `proto/v1/` define every possible sensor and command.
- **Adaptive Pruning:** `xomoi-ctl` generates "Lite" C++ headers by pruning unused or heavy fields (like `bytes raw_val` or `string json_val`) from the Protobuf `oneof` based on the target device's memory constraints. This ensures a simple PIR sensor doesn't allocate 1KB of RAM for an image buffer it will never use.
- **Unified Generation:** A single command generates Go models, C++ SDK headers, and Svelte UI metadata simultaneously, ensuring zero desync between the edge and the core.

## The C++ "Blacksmith" SDK Philosophy

The edge client code is designed to give industrial-grade performance with a "Blynk-killer" developer experience. 

### Telemetry Batching vs EAV Mapping
**The Anti-Pattern:** Firing `sendTemp()`, `sendHum()`, and `sendPres()` separately forces 3 TCP overheads and creates 3 rows in a traditional EAV (Entity-Attribute-Value) database.
**The Xomoi Solution:** The SDK provides a `beginBatch()` / `publishBatch()` API. Multiple readings are combined into a single NanoPB Protobuf array, encoded once, and sent via a single MQTT transmission.
**The Database Synergy:** The Go Backend receives the Protobuf and maps it to a single `JSON` column in SQLite (`{"temperature":25, "humidity":60}`). This architectural alignment reduces SD-Card write operations by up to 80%, massively extending hardware longevity.

### Remote Procedure Calls (RPC)
Xomoi acts as a two-way street. The SDK automatically subscribes to `/xomoi/{mac}/rpc`, allowing the Svelte dashboard to send command strings directly to the hardware via a zero-boilerplate `xomoi.onCommand()` callback hook.

## Modular Protocol: "Core + Plugins"

Xomoi uses a multi-file Protobuf structure to balance developer efficiency with edge-node performance:

1. **`common.proto`**: The "DNA" containing locked field numbers for enums and shared types (Vector3).
2. **`telemetry_base.proto`**: Small, fixed-size types (Float, Int, Bool) for ultra-low-power sensors.
3. **`telemetry_pro.proto`**: Large, variable-size types (JSON, Bytes) for AI and imaging sensors.
4. **`registry.proto` & `command.proto`**: Handshake and control definitions.

The Go backend imports the full suite for total visibility, while the C++ SDK includes only the specific "Plugin" necessary for the hardware.

## Off-Grid & Air-Gapped Capabilities

Xomoi is purpose-built for "Dark Sites" (locations with no internet connectivity):
- **Local Time Sync:** If NTP is unavailable, Xomoi uses a "Browser-Push" mechanism to sync the system clock from the user's device upon the first local login.
- **Store-and-Forward:** In hybrid deployments, edge nodes can cache telemetry in local SQLite and "burst" data to an Enterprise node whenever a secure link (satellite/burst radio) is established.
- **Zero Cloud Dependency:** No external identity providers (Ory, Firebase, Google) are permitted in the Core. Authentication is handled entirely by the local binary.

## Grant Suitability (NLnet / OTF / Sovereign Tech Fund)

Xomoi aligns with global digital sovereignty initiatives by addressing:
- **Anti-Surveillance:** Eliminates data harvesting at the architectural level.
- **Resource Efficiency:** Extends the life of hardware (SD cards) and operates on recycled/low-power computing.
- **Standardization:** Uses open protocols (MQTT, Protobuf, mDNS) to prevent vendor lock-in.
- **Code Auditability:** The "Single-Binary Monolith" ensures the entire system can be audited by a single security researcher without tracing through microservice meshes.

---
*For more information on contributing or implementing Xomoi in a high-security environment, refer to `CONTRIBUTING.md` and `rulebook.md`.*
