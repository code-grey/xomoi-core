#ifndef XOMOI_CLIENT_H
#define XOMOI_CLIENT_H

#include "XomoiTransport.h"
// #include "telemetry.pb.h" // Will be included once NanoPB generates it

namespace xomoi {

/**
 * @brief XomoiClient is the Core Engine of the SDK.
 * 
 * It takes in any ITransport adapter. It handles packing the telemetry 
 * using NanoPB, connecting to the Go backend via MQTT, and listening 
 * for Generic RPC actuations.
 */
class XomoiClient {
private:
    ITransport* _transport;
    char _macAddress[18]; // Fixed size for MAC (e.g. "AA:BB:CC:DD:EE:FF\0")
    char _secretKey[64];  // Fixed max size for the JWT/Secret Key

public:
    // Dependency Injection: Inject the hardware adapter into the Core Engine
    XomoiClient(ITransport* transport, const char* macAddress, const char* secretKey)
        : _transport(transport) {
        
        // SECURITY FIX: Deep copy the strings to prevent dangling pointers.
        // If a user passed `String(WiFi.macAddress()).c_str()`, the pointer would 
        // become invalid instantly after the function ends. Deep copying saves us.
        strncpy(_macAddress, macAddress, sizeof(_macAddress) - 1);
        _macAddress[sizeof(_macAddress) - 1] = '\0'; // Guarantee null termination

        strncpy(_secretKey, secretKey, sizeof(_secretKey) - 1);
        _secretKey[sizeof(_secretKey) - 1] = '\0';
    }

    // Boot up the connection to the Go backend
    bool begin() {
        if (!_transport->isConnected()) {
            if (!_transport->connect()) {
                return false;
            }
        }
        // TODO: Initialize MQTT Handshake over the transport
        return true;
    }

    // Publish a NanoPB compressed telemetry packet
    bool publishTelemetry(uint32_t fieldId, float value) {
        if (!_transport->isConnected()) return false;

        // 1. Pack data into the NanoPB Struct
        // xomoi_TelemetryPayload payload = xomoi_TelemetryPayload_init_zero;
        // payload.field_id = fieldId;
        // payload.value = value;
        // payload.timestamp = getCurrentTime();

        // 2. Encode to binary buffer
        // uint8_t buffer[32];
        // pb_ostream_t stream = pb_ostream_from_buffer(buffer, sizeof(buffer));
        // pb_encode(&stream, xomoi_TelemetryPayload_fields, &payload);

        // 3. Send over MQTT -> _transport->write(mqttPacket, size)

        return true; // placeholder
    }

    // Must be called in the main loop to process incoming RPCs and MQTT Keep-Alives
    void loop() {
        // TODO: Read from _transport, parse MQTT Acks and RPC commands
    }
};

} // namespace xomoi

#endif // XOMOI_CLIENT_H
