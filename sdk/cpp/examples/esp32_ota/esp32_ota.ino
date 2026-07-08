#include <WiFi.h>
#include <HTTPClient.h>
#include <Update.h>
#include <PubSubClient.h>

// Xomoi SDK
#include "XomoiCore.h"
#include "XomoiTransportPubSubClient.h"

// Your Network Credentials
const char* ssid = "YOUR_WIFI_SSID";
const char* password = "YOUR_WIFI_PASSWORD";

// Your Xomoi Device Credentials
const char* device_mac = "AA:BB:CC:DD:EE:FF";
const char* secret_key = "my_super_secret_key";
const char* xomoi_broker_ip = "192.168.1.100"; // IP of the Raspberry Pi running Xomoi

WiFiClient espClient;
PubSubClient mqttClient(espClient);
XomoiTransportPubSubClient xomoiTransport(&mqttClient);
XomoiCore xomoi;

// Function to handle the actual HTTP OTA Download
void performOTAUpdate(const char* url) {
    Serial.printf("Starting OTA from: %s\n", url);

    HTTPClient http;
    http.begin(url);
    int httpCode = http.GET();

    if (httpCode == HTTP_CODE_OK) {
        int contentLength = http.getSize();
        bool canBegin = Update.begin(contentLength);

        if (canBegin) {
            Serial.println("Begin OTA. This may take 1-2 minutes to complete. Things might be quiet for a while.. Patience!");
            
            WiFiClient* client = http.getStreamPtr();
            size_t written = Update.writeStream(*client);

            if (written == contentLength) {
                Serial.println("Written : " + String(written) + " successfully");
            } else {
                Serial.println("Written only : " + String(written) + "/" + String(contentLength) + ". Retry?");
            }

            if (Update.end()) {
                Serial.println("OTA done!");
                if (Update.isFinished()) {
                    Serial.println("Update successfully completed. Rebooting.");
                    ESP.restart();
                } else {
                    Serial.println("Update not finished? Something went wrong!");
                }
            } else {
                Serial.println("Error Occurred. Error #: " + String(Update.getError()));
            }
        } else {
            Serial.println("Not enough space to begin OTA");
        }
    } else {
        Serial.printf("HTTP Download failed, error: %s\n", http.errorToString(httpCode).c_str());
    }
    http.end();
}

// Global RPC Callback to handle commands from the Xomoi Dashboard
void myRpcCallback(const char* command, const char* payload) {
    Serial.printf("Received RPC Command: %s | Payload: %s\n", command, payload);
    
    // Parse the OTA command
    // Format expected: "OTA:/api/v1/devices/{mac}/ota/download"
    if (strncmp(payload, "OTA:", 4) == 0) {
        const char* endpoint = payload + 4;
        
        // Construct the full HTTP URL
        char fullUrl[256];
        snprintf(fullUrl, sizeof(fullUrl), "http://%s:8085%s", xomoi_broker_ip, endpoint);
        
        // Perform the OTA update (Blocking operation)
        performOTAUpdate(fullUrl);
    }
}

void setup() {
    Serial.begin(115200);

    // 1. Connect to Wi-Fi
    WiFi.begin(ssid, password);
    while (WiFi.status() != WL_CONNECTED) {
        delay(500);
        Serial.print(".");
    }
    Serial.println("\nWiFi connected!");

    // 2. Setup MQTT Transport Wrapper
    mqttClient.setServer(xomoi_broker_ip, 1883);
    
    // Note: To use the PubSubClient Wrapper, you MUST route the callback like this:
    mqttClient.setCallback([](char* topic, byte* payload, unsigned int length) {
        xomoiTransport.internalCallback(topic, payload, length);
    });

    // 3. Initialize Xomoi Core
    xomoi.begin(device_mac, secret_key, &xomoiTransport);
    
    // 4. Register the RPC Callback for OTA and remote control
    xomoi.onCommand(myRpcCallback);
    
    // 5. Publish Discovery (Auto-generates UI)
    xomoi.addSensor("temp", "Temperature", "C", 0);
    xomoi.publishDiscovery("v1.1.0"); // Current firmware version
}

void loop() {
    // Keep Xomoi (and MQTT) alive
    xomoi.loop();
    
    // Send a telemetry reading every 10 seconds
    static unsigned long last_send = 0;
    if (millis() - last_send > 10000) {
        last_send = millis();
        xomoi.beginBatch();
        xomoi.addBatchFloat("temp", 26.5);
        xomoi.publishBatch();
    }
}
