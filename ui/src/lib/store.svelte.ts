import { WebRTCClient } from './WebRTCClient';

// Global reactive state
export const globalState = $state({
    webrtcStatus: 'disconnected' as 'connecting' | 'connected' | 'disconnected' | 'error',
    fleet: [] as any[] // Empty by default! Devices only appear when the WebRTC tunnel receives real MQTT data
});

let rtcClient: WebRTCClient | null = null;

function handleIncomingData(rawJson: string) {
    try {
        const msg = JSON.parse(rawJson);
        
        let device = globalState.fleet.find(d => d.id === msg.device_id);
        if (!device) {
            device = { 
                id: msg.device_id,
                friendlyName: msg.device_id,
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
        
        device.status = msg.status || 'healthy';
        
        if (msg.temp !== undefined) {
            device.temp = msg.temp;
            device.tempHistory = [...device.tempHistory.slice(1), msg.temp];
        }
        if (msg.hum !== undefined) {
            device.hum = msg.hum;
            device.humHistory = [...device.humHistory.slice(1), msg.hum];
        }
        if (msg.ack === 'relay_success' && msg.state) device.state = msg.state;
        
    } catch(e) {
        console.error("Failed to parse WebRTC stream", e);
    }
}

export function bootWebRTC() {
    if (rtcClient) return;

    if (import.meta.env.VITE_MOCK_MODE === 'true') {
        console.log("🔥 Mock Mode Enabled! Booting fake data generator...");
        globalState.webrtcStatus = 'connecting';
        
        const deviceIds = ['XOMOI-DEMO-01', 'XOMOI-DEMO-02', 'XOMOI-DEMO-03', 'XOMOI-DEMO-04'];
        deviceIds.forEach(id => {
            globalState.fleet.push({
                id: id,
                friendlyName: id,
                type: 'Render Edge Simulator',
                location: 'Local Network',
                status: 'healthy',
                uptime: '99h',
                state: 'ON',
                temp: 42.5,
                hum: 50.0,
                tempHistory: Array.from({length: 40}, () => parseFloat((35 + Math.random() * 15).toFixed(1))),
                humHistory: Array.from({length: 40}, () => parseFloat((45 + Math.random() * 10).toFixed(1)))
            });
        });
        
        setTimeout(() => {
            globalState.webrtcStatus = 'connected';
            
            setInterval(() => {
                const randomDevice = deviceIds[Math.floor(Math.random() * deviceIds.length)];
                
                const rawJson = JSON.stringify({
                    device_id: randomDevice,
                    type: 'Render Edge Simulator',
                    temp: parseFloat((35 + Math.random() * 15).toFixed(1)),
                    hum: parseFloat((45 + Math.random() * 10).toFixed(1)),
                    status: Math.random() > 0.98 ? 'warning' : 'healthy'
                });
                
                handleIncomingData(rawJson);
            }, 100); // 100ms updates
        }, 800);
        return;
    }

    console.log("Booting Live WebRTC Tunnel to Xomoi-Core...");
    rtcClient = new WebRTCClient('XOMOI-CORE-SERVER', 'ws://localhost:8086/ws');
    
    rtcClient.onStatusChange = (status) => {
        globalState.webrtcStatus = status;
    };
    
    rtcClient.onData = (rawJson) => {
        handleIncomingData(rawJson);
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
