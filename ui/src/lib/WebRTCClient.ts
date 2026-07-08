// WebRTCClient.ts
// Handles the P2P NAT hole-punching for sovereign remote access.

type SignalMessage = {
    type: string;
    target: string;
    source?: string;
    data: any;
};

export class WebRTCClient {
    private pc: RTCPeerConnection;
    private ws: WebSocket | null = null;
    private dataChannel: RTCDataChannel | null = null;
    
    private readonly clientId: string;
    private readonly targetNodeId: string;
    private readonly signalingUrl: string;

    public onData: (msg: string) => void = () => {};
    public onStatusChange: (status: 'connecting' | 'connected' | 'disconnected' | 'error') => void = () => {};

    constructor(targetNodeId: string, signalingUrl: string = 'ws://localhost:8086/ws') {
        this.clientId = 'CLIENT-' + Math.random().toString(36).substring(7);
        this.targetNodeId = targetNodeId;
        this.signalingUrl = `${signalingUrl}?id=${this.clientId}`;
        
        // 1. Initialize Peer Connection with free public STUN servers
        this.pc = new RTCPeerConnection({
            iceServers: [
                { urls: 'stun:stun.l.google.com:19302' },
                { urls: 'stun:global.stun.twilio.com:3478' }
            ]
        });

        // 2. Setup ICE Candidate routing
        this.pc.onicecandidate = (event) => {
            if (event.candidate && this.ws?.readyState === WebSocket.OPEN) {
                this.sendSignal('ice', event.candidate);
            }
        };

        this.pc.onconnectionstatechange = () => {
            if (this.pc.connectionState === 'connected') {
                this.onStatusChange('connected');
                // We successfully punched the hole! Disconnect from signaling server.
                this.ws?.close();
            } else if (this.pc.connectionState === 'disconnected' || this.pc.connectionState === 'failed') {
                this.onStatusChange('disconnected');
            }
        };
    }

    public async connect() {
        this.onStatusChange('connecting');
        
        // 1. Connect to Signaling Server
        this.ws = new WebSocket(this.signalingUrl);
        
        this.ws.onmessage = async (event) => {
            const msg: SignalMessage = JSON.parse(event.data);
            
            if (msg.type === 'answer') {
                await this.pc.setRemoteDescription(new RTCSessionDescription(msg.data));
            } else if (msg.type === 'ice') {
                await this.pc.addIceCandidate(new RTCIceCandidate(msg.data));
            }
        };

        this.ws.onopen = async () => {
            // 2. Create the E2E Encrypted Data Channel
            this.dataChannel = this.pc.createDataChannel('xomoi-telemetry');
            
            this.dataChannel.onmessage = (event) => {
                this.onData(event.data);
            };

            // 3. Create the SDP Offer and send it to the Raspberry Pi
            const offer = await this.pc.createOffer();
            await this.pc.setLocalDescription(offer);
            
            this.sendSignal('offer', offer);
        };
    }

    private sendSignal(type: string, data: any) {
        if (!this.ws || this.ws.readyState !== WebSocket.OPEN) return;
        
        const msg: SignalMessage = {
            type,
            target: this.targetNodeId,
            data
        };
        this.ws.send(JSON.stringify(msg));
    }

    public sendData(payload: string) {
        if (this.dataChannel && this.dataChannel.readyState === 'open') {
            this.dataChannel.send(payload);
        }
    }

    public disconnect() {
        this.dataChannel?.close();
        this.pc.close();
        this.ws?.close();
        this.onStatusChange('disconnected');
    }
}
