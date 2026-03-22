# XOMOI PROTOCOL & TRANSPILER SPECIFICATION (v1.0)

## 1. The "Bigass" vs. "Lite" Philosophy
Xomoi-Core maintains a single, comprehensive "Bigass" Protobuf source in `proto/v1/`. However, we never force an 8-bit microcontroller to compile this entire file. 

The **Xomoi-Transpiler (`xomoi-ctl`)** acts as the intelligent bridge:
- **Backend:** Compiles the full Proto for 100% visibility.
- **Edge:** Generates a "Lite" C++ header by pruning unused fields from the `oneof` union.

## 2. The Transpiler Engine (`xomoi-ctl`)
The transpiler uses a **"Holey Header"** generation strategy:
1. **Field ID Locking:** `xomoi-ctl` ensures that even if a field is pruned (e.g., `json_val` at ID 8), the remaining fields KEEP their original IDs. This prevents the "Silent Killer" of field re-numbering.
2. **Adaptive Union Pruning:** In C++, a `oneof` is a `union`. By commenting out `bytes` or `string` fields in the generated header, `xomoi-ctl` reduces the RAM footprint of every `TelemetryPacket` from ~1KB down to ~8 bytes.
3. **Static Allocation Hints:** The transpiler calculates the `PB_MAX_SIZE` for NanoPB, ensuring the C++ stack is never over-allocated.

## 3. The Modular Protocol Structure
- **`common.proto`**: Shared DNA. Enums and Vector3 types.
- **`telemetry.proto`**: The Master Ingestion format (Signed `TelemetryBatch`).
- **`registry.proto`**: The Handshake mapping (Tag ID -> Metadata).
- **`command.proto`**: Bidirectional Control (Node -> Device).

## 4. Security: HMAC-Lite
Every packet MUST be signed. 
- **Signature:** `HMAC-SHA256(Payload + Timestamp, AuthToken)`.
- **Why?** It's 10x faster than TLS on an ESP32 and fits in 32 bytes of RAM.

## 5. OS-Agnostic Contribution
Xomoi is built for the "Sovereign Developer" on any machine:
- **Taskfile.yml**: Replaces complex Makefiles with a cross-platform runner (Windows/Linux/Mac).
- **Go 1.26**: Compiles to a single binary for any architecture.
- **Protoc**: We use standard `protoc` but wrap it in `xomoi-ctl` to handle the "Lite" generation.

### Build Commands:
- `task proto:go`: Generate Go models (Full).
- `task proto:sdk`: Generate C++ headers (Lite/Adaptive).
- `task build`: Build the Xomoi-Core monolith.
