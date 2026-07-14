#ifndef ESP32_WIFI_TRANSPORT_H
#define ESP32_WIFI_TRANSPORT_H

#include <WiFi.h>
#include "XomoiTransport.h"

namespace xomoi {

/**
 * @brief ESP32 Wi-Fi Implementation of the Xomoi Transport Layer
 * 
 * This is an example of the Adapter Pattern. It wraps the standard Arduino 
 * WiFiClient into our agnostic ITransport interface. If a user wants to use 
 * an Ethernet shield instead, they just write an EthernetTransport that 
 * implements the same 5 methods.
 */
class ESP32WiFiTransport : public ITransport {
private:
    WiFiClient _client;
    const char* _ssid;
    const char* _password;
    const char* _host;
    uint16_t _port;

public:
    ESP32WiFiTransport(const char* ssid, const char* password, const char* host, uint16_t port)
        : _ssid(ssid), _password(password), _host(host), _port(port) {}

    ~ESP32WiFiTransport() {
        disconnect();
    }

    bool connect() override {
        // 1. Connect to Wi-Fi if not already connected
        if (WiFi.status() != WL_CONNECTED) {
            WiFi.begin(_ssid, _password);
            int attempts = 0;
            while (WiFi.status() != WL_CONNECTED && attempts < 20) {
                delay(500);
                attempts++;
            }
            if (WiFi.status() != WL_CONNECTED) return false;
        }

        // 2. Open TCP socket to the Xomoi-Core Server
        if (!_client.connected()) {
            return _client.connect(_host, _port);
        }
        
        return true;
    }

    bool disconnect() override {
        if (_client.connected()) {
            _client.stop();
        }
        // Note: We don't drop the Wi-Fi connection, just the TCP socket
        return true;
    }

    bool isConnected() override {
        return (WiFi.status() == WL_CONNECTED) && _client.connected();
    }

    size_t write(const uint8_t* buffer, size_t size) override {
        if (!isConnected()) return 0;
        return _client.write(buffer, size);
    }

    size_t read(uint8_t* buffer, size_t size) override {
        if (!isConnected() || !_client.available()) return 0;
        
        size_t bytesRead = 0;
        while (bytesRead < size && _client.available()) {
            buffer[bytesRead++] = _client.read();
        }
        return bytesRead;
    }
};

} // namespace xomoi

#endif // ESP32_WIFI_TRANSPORT_H
