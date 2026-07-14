import { WebRTCClient } from './WebRTCClient';

// Global reactive state
export const globalState = $state({
    webrtcStatus: 'disconnected' as 'connecting' | 'connected' | 'disconnected' | 'error',
    fleet: [] as any[] // Empty by default! Devices only appear when the WebRTC tunnel receives real MQTT data
});

let rtcClient: WebRTCClient | null = null;

export function bootWebRTC() {
    if (rtcClient) return;

    console.log("Booting Live WebRTC Tunnel to Xomoi-Core...");
    rtcClient = new WebRTCClient('XOMOI-CORE-SERVER', 'ws://localhost:8086/ws');
    
    rtcClient.onStatusChange = (status) => {
        globalState.webrtcStatus = status;
    };
    
    rtcClient.onData = (rawJson) => {
        try {
            console.log("🔥 WebRTC Received Data:", rawJson);
            const msg = JSON.parse(rawJson);
            
            // Auto-Discovery: If the device isn't in our array yet, add it dynamically!
            let device = globalState.fleet.find(d => d.id === msg.device_id);
            if (!device) {
                device = { 
                    id: msg.device_id,
                    friendlyName: msg.device_id, // Fallback if metadata not loaded 
                    type: msg.type || 'Simulated Edge Node', 
                    location: 'Local Network', 
                    status: 'healthy', 
                    uptime: '0h',
                    state: 'OFF',
                    temp: 0,
                    hum: 0,
                    tempHistory: Array(40).fill(0),
                    humHistory: Array(40).fill(0)
                };
                globalState.fleet.push(device);
            }
            
            // If we are actively receiving telemetry, the device is explicitly healthy
            device.status = 'healthy';
            
            // Update live telemetry
            if (msg.temp !== undefined) {
                device.temp = msg.temp;
                device.tempHistory = [...device.tempHistory.slice(1), msg.temp];
            }
            if (msg.hum !== undefined) {
                device.hum = msg.hum;
                device.humHistory = [...device.humHistory.slice(1), msg.hum];
            }
            if (msg.status !== undefined) device.status = msg.status;
            if (msg.ack === 'relay_success' && msg.state) device.state = msg.state;
            
        } catch(e) {
            console.error("Failed to parse WebRTC stream", e);
        }
    };
    
    rtcClient.connect();
}

export async function fetchDeviceMetadata() {
    try {
        const res = await fetch('/api/v1/devices');
        if (res.ok) {
            const devices = await res.json();
            // Pre-seed the fleet with metadata
            for (const d of devices) {
                let existing = globalState.fleet.find(f => f.id === d.mac_address);
                if (!existing) {
                    globalState.fleet.push({
                        id: d.mac_address,
                        friendlyName: d.name,
                        type: 'Simulated Edge Node',
                        location: 'Local Network',
                        status: 'offline', // will be overwritten by WebRTC
                        uptime: '0h',
                        state: 'OFF',
                        temp: 0,
                        hum: 0,
                        tempHistory: Array(40).fill(0),
                        humHistory: Array(40).fill(0)
                    });
                } else {
                    existing.friendlyName = d.name;
                }
            }
        }
    } catch(e) {
        console.error("Failed to fetch device metadata", e);
    }
}
