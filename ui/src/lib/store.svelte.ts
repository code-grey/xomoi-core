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
                temp: 0, hum: 0, pressure: 0, voltage: 0, lux: 0, fan_speed: 0, imu_hz: 0,
                tempHistory: Array(40).fill(0),
                humHistory: Array(40).fill(0),
                pressureHistory: Array(40).fill(0),
                voltageHistory: Array(40).fill(0),
                luxHistory: Array(40).fill(0),
                fanSpeedHistory: Array(40).fill(0),
                imuHistory: Array(40).fill(0)
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
        if (msg.pressure !== undefined) {
            device.pressure = msg.pressure;
            device.pressureHistory = [...device.pressureHistory.slice(1), msg.pressure];
        }
        if (msg.voltage !== undefined) {
            device.voltage = msg.voltage;
            device.voltageHistory = [...device.voltageHistory.slice(1), msg.voltage];
        }
        if (msg.lux !== undefined) {
            device.lux = msg.lux;
            device.luxHistory = [...device.luxHistory.slice(1), msg.lux];
        }
        if (msg.fan_speed !== undefined) {
            device.fan_speed = msg.fan_speed;
            device.fanSpeedHistory = [...device.fanSpeedHistory.slice(1), msg.fan_speed];
        }
        if (msg.imu_hz !== undefined) {
            device.imu_hz = msg.imu_hz;
            device.imuHistory = [...device.imuHistory.slice(1), msg.imu_hz];
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
                temp: 42.5, hum: 50.0, pressure: 1013.2, voltage: 3.3, lux: 500, fan_speed: 1500, imu_hz: 1000,
                tempHistory: Array.from({length: 40}, () => parseFloat((35 + Math.random() * 15).toFixed(1))),
                humHistory: Array.from({length: 40}, () => parseFloat((45 + Math.random() * 10).toFixed(1))),
                pressureHistory: Array.from({length: 40}, () => parseFloat((1010 + Math.random() * 5).toFixed(1))),
                voltageHistory: Array.from({length: 40}, () => parseFloat((3.2 + Math.random() * 0.2).toFixed(2))),
                luxHistory: Array.from({length: 40}, () => Math.floor(400 + Math.random() * 200)),
                fanSpeedHistory: Array.from({length: 40}, () => Math.floor(1400 + Math.random() * 200)),
                imuHistory: Array.from({length: 40}, () => Math.floor(980 + Math.random() * 40))
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
                    pressure: parseFloat((1010 + Math.random() * 5).toFixed(1)),
                    voltage: parseFloat((3.2 + Math.random() * 0.2).toFixed(2)),
                    lux: Math.floor(400 + Math.random() * 200),
                    fan_speed: Math.floor(1400 + Math.random() * 200),
                    imu_hz: Math.floor(980 + Math.random() * 40),
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
