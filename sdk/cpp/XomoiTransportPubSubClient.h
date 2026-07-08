#ifndef XOMOI_TRANSPORT_PUBSUBCLIENT_H
#define XOMOI_TRANSPORT_PUBSUBCLIENT_H

#include "XomoiTransport.h"
#include <PubSubClient.h>

// CRITICAL SAFETY CHECK: Prevent the "Silent Drop" bug!
// PubSubClient defaults to a 128 byte buffer. Our NanoPB Discovery packets are ~500 bytes.
// If the buffer is too small, PubSubClient silently drops the packet and your UI won't generate.
#if MQTT_MAX_PACKET_SIZE < 1024
#error "FATAL: MQTT_MAX_PACKET_SIZE is too small for Xomoi. You MUST edit PubSubClient.h and change MQTT_MAX_PACKET_SIZE to 1024."
#endif

class XomoiTransportPubSubClient : public IXomoiTransport {
private:
    PubSubClient* mqtt_client;
    MessageCallback user_callback;

public:
    XomoiTransportPubSubClient(PubSubClient* client) : mqtt_client(client), user_callback(nullptr) {}

    bool connect(const char* username, const char* hmac_password) override {
        return mqtt_client->connect(username, username, hmac_password);
    }

    bool connected() override {
        return mqtt_client->connected();
    }

    bool publish(const char* topic, const uint8_t* payload, size_t length) override {
        return mqtt_client->publish(topic, payload, length);
    }

    bool subscribe(const char* topic) override {
        return mqtt_client->subscribe(topic);
    }

    void setCallback(MessageCallback cb) override {
        this->user_callback = cb;
    }
    
    // Call this from your Arduino sketch's global PubSubClient callback
    void internalCallback(const char* topic, const uint8_t* payload, size_t length) {
        if (this->user_callback) {
            this->user_callback(topic, payload, length);
        }
    }

    void loop() override {
        mqtt_client->loop();
    }
};

#endif // XOMOI_TRANSPORT_PUBSUBCLIENT_H
