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
    
    // Internal method to generate HMAC-Lite signature for ultra-secure Auth
    void generateHMAC(const char* payload, char* out_signature);

public:
    XomoiCore();
    
    // Initialize the SDK with credentials and network transport
    void begin(const char* mac, const char* secret, IXomoiTransport* trans);
    
    // Keep-alive and incoming message processor
    void loop();

    // High-level telemetry methods (The "Blynk-killer" simple UX)
    bool sendTemperature(float temp);
    bool sendHumidity(float hum);
    bool sendMotion(bool detected);
};

#endif // XOMOI_CORE_H
