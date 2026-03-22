# Xomoi-Core

Xomoi-Core is a sovereign, open-source edge node designed for digital freedom and data sovereignty. It provides a single-binary infrastructure for IoT telemetry, liberating users from proprietary data-harvesting platforms by keeping all intelligence and storage local.

## Core Mandates

1. Sovereign Monolith: A single Go binary containing the MQTT broker, database, and user interface. No external dependencies or microservices required.
2. Privacy by Design: Zero telemetry. No phone-home functionality. All data is owned and stored by the user on their own hardware.
3. Hardware Longevity: Specifically engineered to protect edge storage (SD cards/eMMC) via in-memory state management and batched persistence.
4. Protocol Efficiency: Utilizes a tag-based component ID model for low-bandwidth, high-frequency telemetry.

## Architecture

Xomoi-Core is built on a high-concurrency Go 1.26 backend utilizing:
- Mochi-MQTT: Embedded MQTT broker.
- SQLite (WAL Mode): Embedded relational storage with flash-wear protection.
- Svelte 5: Embedded static dashboard for low-power visualization.
- Repository Pattern: Interface-driven data access for strict architectural purity.

## Status

Xomoi is currently in active development (Phase 1: The Sovereign Contract). DePIN and blockchain features are strictly isolated via build tags and are not part of the Xomoi-Core runtime by default.

## Developing & Compiling Protobufs

Xomoi uses an OS-Agnostic toolchain managed by `Taskfile.yml`. The core data structures are defined in `proto/v1/`.

To compile the protocol into Go models or the adaptive C++ SDK:
1. Ensure you have `go 1.26+`, `protoc`, and `task` installed.
2. Run `task proto:go` to generate the backend models.
3. For detailed protocol architecture, refer to the [Protocol Specification](docs/PROTOCOL_SPEC.md).

For full contribution guidelines, see [CONTRIBUTING.md](CONTRIBUTING.md).

## Licensing

Xomoi-Core is released under the GNU Affero General Public License v3.0 (AGPLv3). This ensures the software remains a public good and protects it from proprietary, closed-source cloud capture.
