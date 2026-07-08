#ifndef XOMOI_CORE_H
#define XOMOI_CORE_H

#include <stdint.h>
#include <stdbool.h>

// Forward declaration of the transport interface to decouple from specific WiFi/MQTT libs
class IXomoiTransport;

class XomoiCore {
private:
    const char* device_mac;
    const char* secret_key;
    IXomoiTransport* transport;
    
    // Internal state for building the discovery payload before publishing
    struct _xomoi_Discovery* discovery_state;
    
    // Internal state for telemetry batching
    struct _xomoi_Telemetry* batch_state;
    
    // RPC Callback pointer
    typedef void (*RpcCallback)(const char* command, const char* payload);
    RpcCallback rpc_callback;
    
    // Internal static callback to route from Transport to XomoiCore instance
    static void _internalMessageCallback(const char* topic, const uint8_t* payload, size_t length);
    static XomoiCore* _instance; // For static routing
    
    // Internal method to generate HMAC-Lite signature for ultra-secure Auth
    void generateHMAC(const char* payload, char* out_signature);
    
    // Internal helper to publish encoded protobufs
    bool publishTelemetry(struct _xomoi_Telemetry* msg);

public:
    XomoiCore();
    
    // Initialize the SDK with credentials and network transport
    void begin(const char* mac, const char* secret, IXomoiTransport* trans);
    
    // Discovery Payload Builder (The magic behind Svelte Auto-Generation)
    bool addSensor(const char* key, const char* display_name, const char* unit, uint32_t data_type);
    bool publishDiscovery(const char* firmware_version);
    
    // Keep-alive and incoming message processor
    void loop();

    // High-level telemetry methods (The "Blynk-killer" simple UX)
    bool sendTemperature(float temp);
    bool sendHumidity(float hum);
    bool sendMotion(bool detected);
    
    // Advanced Telemetry Batching API (Network & Battery Saver)
    void beginBatch();
    bool addBatchFloat(const char* key, float value);
    bool publishBatch();
    
    // RPC Control: Listen for commands from the dashboard
    void onCommand(RpcCallback callback);
};

#endif // XOMOI_CORE_H
