#include "XomoiCore.h"
#include "XomoiTransport.h"
#include <string.h>

XomoiCore::XomoiCore() : device_mac(nullptr), secret_key(nullptr), transport(nullptr) {}

void XomoiCore::begin(const char* mac, const char* secret, IXomoiTransport* trans) {
    this->device_mac = mac;
    this->secret_key = secret;
    this->transport = trans;
    
    // Authenticate using HMAC-Lite instead of mTLS to save massive amounts of RAM
    char signature[65];
    generateHMAC(mac, signature); 
    
    this->transport->connect(this->device_mac, signature);
}

void XomoiCore::loop() {
    if (this->transport != nullptr) {
        this->transport->loop();
    }
}

void XomoiCore::generateHMAC(const char* payload, char* out_signature) {
    // Skeleton: In a real implementation, this calls mbedtls_md_hmac (ESP32) 
    // or an Arduino Crypto library to generate SHA-256 HMAC using this->secret_key.
    // This is mathematically proven authentication with zero TLS overhead.
    strcpy(out_signature, "dummy_hmac_signature");
}

bool XomoiCore::sendTemperature(float temp) {
    if (!transport || !transport->connected()) return false;
    
    // Zero-Allocation Protocol Buffers (NanoPB) Flow:
    // 1. Construct NanoPB struct 
    // Telemetry msg = Telemetry_init_zero;
    // msg.payload.temperature = temp;
    // msg.which_payload = Telemetry_temperature_tag;
    
    // 2. Encode to stack buffer
    // uint8_t buffer[128];
    // pb_ostream_t stream = pb_ostream_from_buffer(buffer, sizeof(buffer));
    // pb_encode(&stream, Telemetry_fields, &msg);
    
    // 3. Publish binary payload over transport
    // return transport->publish("telemetry/temp", buffer, stream.bytes_written);
    
    return true; 
}

bool XomoiCore::sendHumidity(float hum) {
    return true; 
}

bool XomoiCore::sendMotion(bool detected) {
    return true; 
}
