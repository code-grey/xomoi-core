#ifndef XOMOI_TRANSPORT_H
#define XOMOI_TRANSPORT_H

#include <stdint.h>
#include <stddef.h>

// IXomoiTransport is the "Adapter Interface". 
// This decouples our SDK from the hardware. It allows XomoiCore to run on ANY microcontroller 
// by wrapping the underlying network library (PubSubClient, AsyncMQTT, Ethernet, Cellular GSM, etc.)
class IXomoiTransport {
public:
    virtual ~IXomoiTransport() {}
    
    // Connect to the broker using HMAC-Lite credentials
    virtual bool connect(const char* username, const char* hmac_password) = 0;
    
    // Check connection status
    virtual bool connected() = 0;
    
    // Publish raw binary NanoPB payload
    virtual bool publish(const char* topic, const uint8_t* payload, size_t length) = 0;
    
    // Keep connection alive
    virtual void loop() = 0;
};

#endif // XOMOI_TRANSPORT_H
