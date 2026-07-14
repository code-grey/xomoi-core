#ifndef XOMOI_TRANSPORT_H
#define XOMOI_TRANSPORT_H

#include <stddef.h>
#include <stdint.h>

namespace xomoi {

/**
 * @brief ITransport defines the hardware-agnostic network layer.
 * 
 * By using this interface (The Adapter Pattern), the Xomoi SDK never 
 * hardcodes <WiFi.h> or <Ethernet.h>. It doesn't care if the device is 
 * an ESP32 on Wi-Fi, an Arduino on Ethernet, or an STM32 on LoRaWAN.
 * As long as the user provides an object that implements these 5 methods,
 * the Xomoi engine will function perfectly.
 */
class ITransport {
public:
    virtual ~ITransport() {}
    
    // Network lifecycle
    virtual bool connect() = 0;
    virtual bool disconnect() = 0;
    virtual bool isConnected() = 0;
    
    // Raw byte stream (used by the MQTT parser)
    virtual size_t write(const uint8_t* buffer, size_t size) = 0;
    virtual size_t read(uint8_t* buffer, size_t size) = 0;
};

} // namespace xomoi

#endif // XOMOI_TRANSPORT_H
