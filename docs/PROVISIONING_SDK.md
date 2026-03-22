# Xomoi-Core: Provisioning and Sensor Ecosystem

## Introduction

This document defines the "Smooth AF" user experience for provisioning new hardware and the low-code/no-code sensor ecosystem. Xomoi aims to be accessible to hobbyists while remaining robust enough for off-grid enterprise security.

## 1. The "Xomoi-Claim" Provisioning Protocol

To eliminate the friction of physical certificates (mTLS) and manual configuration, Xomoi uses a "Claim-and-Bind" flow.

### Workflow
1. **Beacon Mode:** An unprovisioned device (ESP32/Pico) starts a captive portal Wi-Fi AP named `Xomoi-Claim-XXXX`.
2. **Detection:** The Xomoi-Core dashboard (or mobile app) detects the AP via an mDNS-style scan.
3. **Claiming:** The user clicks "Claim" in the UI. 
4. **Handshake:** 
    - The provisioning client (browser/phone) connects to the sensor's AP.
    - It pushes the `WIFI_SSID`, `WIFI_PASS`, and a unique `HMAC_LITE_TOKEN`.
    - The sensor stores these in NVS (Non-Volatile Storage) and reboots.
5. **Binding:** The sensor joins the main Wi-Fi and connects to the Xomoi-Core MQTT broker using its MAC address as the username and the `HMAC_LITE_TOKEN` as the password.

### Security
The `HMAC_LITE_TOKEN` is generated on the Xomoi-Core node and is specific to the device's MAC address. It is signed with the node's secret key, preventing unmanaged sensors from joining the broker.

## 2. The "Xomoi-Catalyst" Sensor Ecosystem

Xomoi provides a "Sensors Directory" to enable low-code/no-code deployment.

### The Sensors Directory (`templates/sensors/`)
A collection of JSON-based blueprints defining:
- **Identifier:** e.g., `dht22_temp_hum`.
- **Dependencies:** Required C++ libraries.
- **Protocol:** The Tag-Based IDs used for telemetry.
- **Default Pins:** Recommended GPIOs for common MCUs (ESP32, Pico).

### Generation Options
- **CLI (xomoi-ctl):** For developers. `xomoi-ctl generate --mcu esp32 --sensor bme280`. Generates a complete PlatformIO or Arduino project.
- **Web-Flasher (No-Code):** For beginners. The Xomoi dashboard uses WebSerial/WebUSB to flash pre-compiled binaries directly to a connected microcontroller based on the selected sensor blueprint.

### The Xomoi C++ SDK

The SDK is a zero-allocation, static-memory library that abstracts the complexity of Protobuf and MQTT.

#### Transpiler-Driven SDK Generation (`xomoi-ctl`)
To solve the "Memory Trap" of complex Protobuf files on 8-bit/32-bit MCUs, Xomoi does not use a generic Protobuf header. Instead, the `xomoi-ctl` tool generates a **Tailored SDK** for each device:
- **Adaptive Pruning:** If a device only needs simple telemetry (Temp/Humidity), `xomoi-ctl` prunes the "heavy" fields (like `bytes raw_val` or `string json_val`) from the Protobuf definition, drastically reducing the RAM required for static message buffers.
- **Static Buffer Sizing:** The tool calculates the exact maximum size of the Protobuf message and pre-allocates a static stack buffer, ensuring the device never hits a heap fragmentation error.
- **Unified Handshake:** Generates the C++ constants for `tag_id` mapping automatically, ensuring the firmware and the dashboard are always in sync.

### Core Features
- **Static Memory:** No `malloc()` or `new` allowed.
- **Auto-Registration:** The SDK automatically sends a "Registry" packet on boot to map Tag IDs to human-readable names (e.g., `Tag 101` -> `Living Room Temperature`).
- **OTA Ready:** Built-in support for pulling firmware updates from the Xomoi-Core node.
- **Low-Power Sleep:** Native support for deep-sleep cycles to preserve battery life on solar-powered nodes.

---
*This ecosystem ensures that Xomoi is as easy to use as a proprietary hub while remaining 100% open and sovereign.*
