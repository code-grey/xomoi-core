#include <Arduino.h>
#include "XomoiCore.h"
// #include "XomoiPubSubAdapter.h" // Hypothetical wrapper for Arduino PubSubClient

XomoiCore xomoi;
// XomoiPubSubAdapter mqttAdapter(wifiClient, "192.168.1.100", 1883);

void setup() {
    Serial.begin(115200);
    
    // Setup WiFi here...

    // Zero-friction initialization
    // Parameter 1: MAC Address (MQTT Username)
    // Parameter 2: Secret Key (Used for HMAC Password generation)
    // Parameter 3: The Transport Adapter (Hardware agnostic)
    // xomoi.begin("00:1A:2B:3C:4D:5E", "SuperSecretKey123", &mqttAdapter);
}

void loop() {
    xomoi.loop();
    
    // Read your sensor (e.g., DHT22)
    float temp = 25.4; 
    
    // Push telemetry to the sovereign edge node.
    // This internally encodes to NanoPB and publishes over MQTT.
    xomoi.sendTemperature(temp);
    
    delay(5000); // 5 second interval
}
