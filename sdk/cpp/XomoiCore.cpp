#include "XomoiCore.h"
#include "XomoiTransport.h"
#include "../proto/xomoi.pb.h"
#include <pb_encode.h>
#include <mbedtls/md.h>
#include <string.h>
#include <stdio.h>
#include <stdlib.h>

XomoiCore* XomoiCore::_instance = nullptr;

XomoiCore::XomoiCore() : device_mac(nullptr), secret_key(nullptr), transport(nullptr), discovery_state(nullptr), batch_state(nullptr), rpc_callback(nullptr) {
    _instance = this;
}

void XomoiCore::begin(const char* mac, const char* secret, IXomoiTransport* trans) {
    this->device_mac = mac;
    this->secret_key = secret;
    this->transport = trans;
    
    // Register the internal static callback for incoming MQTT RPC commands
    this->transport->setCallback(XomoiCore::_internalMessageCallback);
    
    // Authenticate using HMAC-Lite instead of mTLS to save massive amounts of RAM
    char signature[65];
    generateHMAC(mac, signature); 
    
    if (this->transport->connect(this->device_mac, signature)) {
        // Subscribe to RPC commands immediately upon connection
        char rpcTopic[64];
        snprintf(rpcTopic, sizeof(rpcTopic), "/xomoi/%s/rpc", this->device_mac);
        this->transport->subscribe(rpcTopic);
    }
}

void XomoiCore::_internalMessageCallback(const char* topic, const uint8_t* payload, size_t length) {
    if (_instance && _instance->rpc_callback) {
        char cmdBuf[128] = {0};
        size_t copyLen = length < 127 ? length : 127;
        memcpy(cmdBuf, payload, copyLen);
        _instance->rpc_callback(topic, cmdBuf);
    }
}

void XomoiCore::loop() {
    if (this->transport != nullptr) {
        this->transport->loop();
    }
}

void XomoiCore::generateHMAC(const char* payload, char* out_signature) {
    // Hardware-accelerated SHA-256 HMAC using mbedtls (ESP32/ESP8266 Native)
    unsigned char hmacResult[32];
    mbedtls_md_context_t ctx;
    mbedtls_md_type_t md_type = MBEDTLS_MD_SHA256;
    
    mbedtls_md_init(&ctx);
    mbedtls_md_setup(&ctx, mbedtls_md_info_from_type(md_type), 1); // 1 = HMAC
    mbedtls_md_hmac_starts(&ctx, (const unsigned char*)this->secret_key, strlen(this->secret_key));
    mbedtls_md_hmac_update(&ctx, (const unsigned char*)payload, strlen(payload));
    mbedtls_md_hmac_finish(&ctx, hmacResult);
    mbedtls_md_free(&ctx);
    
    // Convert 32-byte binary hash to 64-character hex string
    for (int i = 0; i < 32; i++) {
        sprintf(&out_signature[i * 2], "%02x", hmacResult[i]);
    }
    out_signature[64] = '\0';
}

bool XomoiCore::publishTelemetry(xomoi_Telemetry* msg) {
    if (!transport || !transport->connected()) return false;
    
    uint8_t buffer[xomoi_Telemetry_size];
    pb_ostream_t stream = pb_ostream_from_buffer(buffer, sizeof(buffer));
    
    if (!pb_encode(&stream, xomoi_Telemetry_fields, msg)) {
        return false;
    }
    
    char topic[64];
    snprintf(topic, sizeof(topic), "/xomoi/%s/telemetry", this->device_mac);
    return transport->publish(topic, buffer, stream.bytes_written);
}

bool XomoiCore::addSensor(const char* key, const char* display_name, const char* unit, uint32_t data_type) {
    if (!this->discovery_state) {
        this->discovery_state = (xomoi_Discovery*)malloc(sizeof(xomoi_Discovery));
        *this->discovery_state = xomoi_Discovery_init_zero;
    }
    
    if (this->discovery_state->sensors_count >= 8) return false; // Max sensors reached
    
    int idx = this->discovery_state->sensors_count;
    strncpy(this->discovery_state->sensors[idx].key, key, sizeof(this->discovery_state->sensors[idx].key));
    strncpy(this->discovery_state->sensors[idx].display_name, display_name, sizeof(this->discovery_state->sensors[idx].display_name));
    strncpy(this->discovery_state->sensors[idx].unit, unit, sizeof(this->discovery_state->sensors[idx].unit));
    this->discovery_state->sensors[idx].data_type = data_type;
    
    this->discovery_state->sensors_count++;
    return true;
}

bool XomoiCore::publishDiscovery(const char* firmware_version) {
    if (!transport || !transport->connected() || !this->discovery_state) return false;
    
    strncpy(this->discovery_state->mac_address, this->device_mac, sizeof(this->discovery_state->mac_address));
    strncpy(this->discovery_state->firmware_version, firmware_version, sizeof(this->discovery_state->firmware_version));
    
    uint8_t buffer[xomoi_Discovery_size];
    pb_ostream_t stream = pb_ostream_from_buffer(buffer, sizeof(buffer));
    
    if (!pb_encode(&stream, xomoi_Discovery_fields, this->discovery_state)) {
        return false;
    }
    
    char topic[64];
    snprintf(topic, sizeof(topic), "/xomoi/%s/discovery", this->device_mac);
    
    // Cleanup state after publish
    free(this->discovery_state);
    this->discovery_state = nullptr;
    
    return transport->publish(topic, buffer, stream.bytes_written);
}

bool XomoiCore::sendTemperature(float temp) {
    xomoi_Telemetry msg = xomoi_Telemetry_init_zero;
    msg.timestamp = 0;
    strncpy(msg.device_type, "GENERIC", sizeof(msg.device_type));
    
    msg.data_count = 1;
    strncpy(msg.data[0].key, "temperature", sizeof(msg.data[0].key));
    msg.data[0].value = temp;
    
    return publishTelemetry(&msg);
}

bool XomoiCore::sendHumidity(float hum) {
    xomoi_Telemetry msg = xomoi_Telemetry_init_zero;
    msg.timestamp = 0;
    strncpy(msg.device_type, "GENERIC", sizeof(msg.device_type));
    
    msg.data_count = 1;
    strncpy(msg.data[0].key, "humidity", sizeof(msg.data[0].key));
    msg.data[0].value = hum;
    
    return publishTelemetry(&msg);
}

bool XomoiCore::sendMotion(bool detected) {
    xomoi_Telemetry msg = xomoi_Telemetry_init_zero;
    msg.timestamp = 0;
    strncpy(msg.device_type, "GENERIC", sizeof(msg.device_type));
    
    msg.data_count = 1;
    strncpy(msg.data[0].key, "motion", sizeof(msg.data[0].key));
    msg.data[0].value = detected ? 1.0f : 0.0f; // Cast bool to float
    
    return publishTelemetry(&msg);
}

void XomoiCore::beginBatch() {
    if (!this->batch_state) {
        this->batch_state = (xomoi_Telemetry*)malloc(sizeof(xomoi_Telemetry));
    }
    *this->batch_state = xomoi_Telemetry_init_zero;
    this->batch_state->timestamp = 0;
    strncpy(this->batch_state->device_type, "GENERIC", sizeof(this->batch_state->device_type));
}

bool XomoiCore::addBatchFloat(const char* key, float value) {
    if (!this->batch_state) return false;
    if (this->batch_state->data_count >= 8) return false; // Max NanoPB array size reached
    
    int idx = this->batch_state->data_count;
    strncpy(this->batch_state->data[idx].key, key, sizeof(this->batch_state->data[idx].key));
    this->batch_state->data[idx].value = value;
    this->batch_state->data_count++;
    
    return true;
}

bool XomoiCore::publishBatch() {
    if (!this->batch_state) return false;
    
    bool success = publishTelemetry(this->batch_state);
    
    free(this->batch_state);
    this->batch_state = nullptr;
    
    return success;
}

void XomoiCore::onCommand(RpcCallback callback) {
    this->rpc_callback = callback;
}
