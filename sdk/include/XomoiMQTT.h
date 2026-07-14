#ifndef XOMOI_MQTT_H
#define XOMOI_MQTT_H

#include <stdint.h>
#include <stddef.h>
#include <string.h>

namespace xomoi {

/**
 * @brief Ultra-lightweight MQTT 3.1.1 Packet Builder
 * 
 * Instead of forcing users to install massive third-party MQTT libraries 
 * (like PubSubClient or coreMQTT), we provide a zero-dependency packet builder. 
 * This uses pure static memory to wrap NanoPB payloads into MQTT packets.
 */
class MQTTPacketBuilder {
public:
    // Builds an MQTT PUBLISH packet header directly into a buffer.
    // Returns the total size of the MQTT header written to the buffer.
    static size_t buildPublishHeader(
        uint8_t* buffer, 
        size_t maxBufferSize,
        const char* topic, 
        size_t payloadSize, 
        bool retain = false, 
        uint8_t qos = 1) 
    {
        size_t topicLen = strlen(topic);
        
        // Calculate Remaining Length (Topic length bytes + Topic String + Packet ID (if QoS > 0) + Payload)
        size_t remainingLength = 2 + topicLen + payloadSize;
        if (qos > 0) remainingLength += 2; // Packet Identifier takes 2 bytes

        // SECURITY CHECK: Buffer Overflow Prevention
        // Calculate maximum possible header size (1 byte fixed + max 4 bytes variable length + topic + packet ID)
        size_t maxHeaderSize = 1 + 4 + 2 + topicLen + 2; 
        if (maxHeaderSize > maxBufferSize) {
            return 0; // Buffer is too small. Abort instantly to prevent stack corruption.
        }

        // MQTT Fixed Header
        // 0x30 is the PUBLISH control packet type.
        // We bit-shift QoS and Retain flags into the first byte.
        buffer[0] = 0x30 | ((qos & 0x03) << 1) | (retain ? 0x01 : 0x00);
        
        // Calculate Remaining Length (Topic length bytes + Topic String + Packet ID (if QoS > 0) + Payload)
        size_t remainingLength = 2 + topicLen + payloadSize;
        if (qos > 0) remainingLength += 2; // Packet Identifier takes 2 bytes

        // Encode Remaining Length (MQTT uses a variable-length encoding scheme)
        size_t pos = 1;
        do {
            uint8_t encodedByte = remainingLength % 128;
            remainingLength /= 128;
            if (remainingLength > 0) {
                encodedByte |= 128;
            }
            buffer[pos++] = encodedByte;
        } while (remainingLength > 0);

        // Encode Topic Length (2 bytes, MSB first)
        buffer[pos++] = (topicLen >> 8) & 0xFF;
        buffer[pos++] = topicLen & 0xFF;

        // Copy Topic String
        memcpy(&buffer[pos], topic, topicLen);
        pos += topicLen;

        // If QoS > 0, inject a Packet Identifier (hardcoded to 0x00 0x01 for this stub)
        if (qos > 0) {
            buffer[pos++] = 0x00;
            buffer[pos++] = 0x01;
        }

        // At this point, the buffer contains a perfect MQTT header.
        // The XomoiClient can now write this buffer to the TCP socket, 
        // followed instantly by writing the NanoPB payload buffer.
        
        return pos;
    }
    
    // Builds a simple 2-byte MQTT PINGREQ for heartbeats
    static size_t buildPingReq(uint8_t* buffer, size_t maxBufferSize) {
        if (maxBufferSize < 2) return 0; // Bounds check
        buffer[0] = 0xC0; // PINGREQ Type
        buffer[1] = 0x00; // Remaining length 0
        return 2;
    }
};

} // namespace xomoi

#endif // XOMOI_MQTT_H
