<script lang="ts">
  import { ShieldAlert, CheckCircle, Activity, Cpu, Thermometer, Droplets, Zap, Settings } from 'lucide-svelte';
  import WebFlasher from '../WebFlasher.svelte';
  
  // Simulated State for the Fleet Matrix
  let fleet = $state([
    { id: 'ESP32-A1', type: 'DHT11 Env', location: 'Greenhouse Core', status: 'healthy', temp: 24.5, hum: 45, uptime: '72h' },
    { id: 'ESP32-B2', type: 'DHT11 Env', location: 'Server Rack 01', status: 'alert', temp: 34.2, hum: 20, uptime: '12h' }, // Overheating!
    { id: 'PICO-C3', type: 'PIR Motion', location: 'Garage Door', status: 'healthy', motion: false, uptime: '240h' },
    { id: 'ESP8266-D4', type: 'Relay Switch', location: 'Exhaust Fan', status: 'healthy', state: 'OFF', uptime: '15h' },
  ]);

  let selectedDevice = $state<any>(null);
  let showFlasher = $state(false);

  function getStatusColor(status: string) {
    return status === 'alert' ? 'var(--accent-orange)' : 'var(--accent-cyan)';
  }
</script>

<div class="fleet-container">
  {#if !selectedDevice}
    <!-- 1. THE DEVICE MATRIX -->
    <div class="matrix-header">
      <div class="title-group">
        <h2>Device Fleet Matrix</h2>
        <p class="subtitle">Real-time topology of connected endpoints</p>
      </div>
      <div class="header-actions">
        <button class="provision-btn glass-panel" onclick={() => showFlasher = true}>+ Provision New (Web-Flasher)</button>
        <div class="legend glass-panel">
          <span class="dot cyan"></span> Healthy
          <span class="dot orange"></span> Alerting
        </div>
      </div>
    </div>

    <div class="matrix-grid">
      {#each fleet as device}
        <button 
          class="device-node glass-panel {device.status}" 
          onclick={() => selectedDevice = device}
        >
          <div class="node-icon">
            {#if device.status === 'alert'}
              <ShieldAlert color="var(--accent-orange)" size={28} />
            {:else}
              <CheckCircle color="var(--accent-cyan)" size={28} />
            {/if}
          </div>
          <div class="node-info">
            <span class="node-id mono">{device.id}</span>
            <span class="node-loc">{device.location}</span>
          </div>
        </button>
      {/each}
    </div>

  {:else}
    <!-- 2. THE DRILL-DOWN DASHBOARD -->
    <div class="drill-down">
      <div class="drill-header">
        <button class="back-btn glass-panel" onclick={() => selectedDevice = null}>
          ← Return to Matrix
        </button>
        <div class="status-badge glass-panel {selectedDevice.status}">
          {#if selectedDevice.status === 'alert'} <ShieldAlert size={16} /> {:else} <CheckCircle size={16} /> {/if}
          {selectedDevice.status.toUpperCase()}
        </div>
      </div>
      
      <div class="profile-card glass-panel">
        <div class="device-identity">
          <div class="avatar" style="border-color: {getStatusColor(selectedDevice.status)}">
            <Cpu size={32} color={getStatusColor(selectedDevice.status)} />
          </div>
          <div class="identity-text">
            <h2 class="mono">{selectedDevice.id}</h2>
            <p>{selectedDevice.type} • {selectedDevice.location}</p>
          </div>
        </div>
        <div class="uptime-badge">
          <Zap size={14} /> Uptime: {selectedDevice.uptime}
        </div>
      </div>

      <!-- Dynamic Sensor Widgets -->
      <h3 class="section-title">Live Telemetry</h3>
      <div class="widgets-grid">
        {#if selectedDevice.type === 'DHT11 Env'}
          <div class="sensor-widget glass-panel">
            <Thermometer size={24} color={selectedDevice.status === 'alert' ? 'var(--accent-orange)' : 'var(--text-secondary)'} />
            <div class="widget-data">
              <span class="val mono">{selectedDevice.temp}°C</span>
              <span class="lbl">Temperature</span>
            </div>
          </div>
          <div class="sensor-widget glass-panel">
            <Droplets size={24} color="var(--text-secondary)" />
            <div class="widget-data">
              <span class="val mono">{selectedDevice.hum}%</span>
              <span class="lbl">Humidity</span>
            </div>
          </div>
        {:else if selectedDevice.type === 'PIR Motion'}
          <div class="sensor-widget glass-panel">
            <Activity size={24} color="var(--accent-cyan)" />
            <div class="widget-data">
              <span class="val mono">{selectedDevice.motion ? 'MOTION DETECTED' : 'CLEAR'}</span>
              <span class="lbl">PIR Status</span>
            </div>
          </div>
        {:else if selectedDevice.type === 'Relay Switch'}
          <button class="sensor-widget glass-panel interactive-relay">
            <Zap size={24} color={selectedDevice.state === 'ON' ? 'var(--accent-orange)' : 'var(--text-secondary)'} />
            <div class="widget-data">
              <span class="val mono">{selectedDevice.state}</span>
              <span class="lbl">Toggle Relay</span>
            </div>
          </button>
        {/if}
      </div>

      <!-- 3. THE ALERT RULES ENGINE -->
      <div class="rules-engine glass-panel">
        <div class="rules-header">
          <Settings size={20} color="var(--accent-purple)" />
          <h3>Alert Rules Engine</h3>
        </div>
        <div class="rule-builder">
          <span>IF</span>
          <select class="rule-select">
            <option>Temperature</option>
            <option>Humidity</option>
          </select>
          <select class="rule-select">
            <option>&gt;</option>
            <option>&lt;</option>
            <option>==</option>
          </select>
          <input type="number" value={selectedDevice.type === 'DHT11 Env' ? 30 : 1} class="rule-input mono" />
          <span>THEN</span>
          <select class="rule-select action">
            <option>Trigger Alert Status</option>
            <option>Toggle Relay (ESP8266-D4)</option>
            <option>Send Discord Webhook</option>
          </select>
        </div>
        <div class="rule-actions">
          <button class="save-btn">+ Add Rule</button>
        </div>
      </div>
    </div>
  {/if}
</div>

{#if showFlasher}
  <WebFlasher close={() => showFlasher = false} />
{/if}

<style>
  .fleet-container {
    display: flex;
    flex-direction: column;
    gap: 24px;
    height: 100%;
  }

  /* Matrix Layout */
  .matrix-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-end;
  }
  .title-group h2 { color: var(--text-primary); margin-bottom: 4px; }
  .subtitle { color: var(--text-secondary); font-size: 0.9rem;  }
  .header-actions {
    display: flex;
    gap: 16px;
    align-items: center;
  }
  .provision-btn {
    padding: 8px 16px;
    color: var(--accent-cyan);
    font-weight: 600;
    font-size: 0.85rem;
    letter-spacing: 0.05em;
    border-color: var(--accent-cyan-dim);
  }
  .provision-btn:hover {
    background: var(--accent-cyan-dim);
    box-shadow: 0 0 16px rgba(0, 255, 204, 0.2);
  }
  
  .legend {
    display: flex;
    gap: 12px;
    padding: 8px 16px;
    align-items: center;
    font-size: 0.8rem;
    font-weight: 600;
  }
  .dot { width: 10px; height: 10px; border-radius: 50%; display: inline-block; }
  .dot.cyan { background: var(--accent-cyan); box-shadow: 0 0 8px rgba(0, 255, 204, 0.4); }
  .dot.orange { background: var(--accent-orange); box-shadow: 0 0 8px rgba(255, 85, 0, 0.4); }

  .matrix-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
    gap: 16px;
  }

  .device-node {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    padding: 20px;
    gap: 16px;
    text-align: left;
    background: rgba(0,0,0,0.4);
  }
  .device-node:hover {
    transform: translateY(-4px);
    background: rgba(255,255,255,0.05);
  }
  .device-node.alert {
    border-color: rgba(255, 85, 0, 0.3);
    background: rgba(255, 85, 0, 0.05);
  }
  .device-node.alert:hover {
    box-shadow: 0 8px 24px rgba(255, 85, 0, 0.15);
  }
  
  .node-info {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  .node-id { font-weight: 700; color: var(--text-primary); font-size: 1.1rem; }
  .node-loc { color: var(--text-secondary); font-size: 0.8rem; }
  
  /* Drill Down Layout */
  .drill-down {
    display: flex;
    flex-direction: column;
    gap: 24px;
    animation: fadein 0.2s ease;
  }
  @keyframes fadein { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }

  .drill-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .back-btn {
    padding: 8px 16px;
    color: var(--text-primary);
    font-weight: 600;
  }
  .status-badge {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 16px;
    font-weight: 700;
    font-size: 0.85rem;
    letter-spacing: 0.05em;
  }
  .status-badge.alert { color: var(--accent-orange); border-color: var(--accent-orange); }
  .status-badge.healthy { color: var(--accent-cyan); border-color: var(--accent-cyan); }

  .profile-card {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 24px;
  }
  .device-identity {
    display: flex;
    align-items: center;
    gap: 20px;
  }
  .avatar {
    width: 64px;
    height: 64px;
    border-radius: 50%;
    border: 2px solid;
    display: flex;
    justify-content: center;
    align-items: center;
    background: rgba(0,0,0,0.5);
  }
  .identity-text h2 { font-size: 2rem; margin-bottom: 4px; }
  .identity-text p { color: var(--text-secondary); }
  .uptime-badge {
    display: flex;
    align-items: center;
    gap: 6px;
    color: var(--accent-purple);
    font-family: var(--font-mono);
    font-size: 0.9rem;
  }

  .section-title {
    font-size: 1.1rem;
    color: var(--text-secondary);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    margin-top: 8px;
  }

  .widgets-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 16px;
  }
  .sensor-widget {
    padding: 24px;
    display: flex;
    align-items: center;
    gap: 20px;
  }
  .widget-data {
    display: flex;
    flex-direction: column;
  }
  .widget-data .val { font-size: 1.8rem; font-weight: 700; color: var(--text-primary); }
  .widget-data .lbl { font-size: 0.85rem; color: var(--text-secondary); }

  .interactive-relay {
    cursor: pointer;
    border-color: rgba(255, 255, 255, 0.2);
  }
  .interactive-relay:active { transform: scale(0.98); }

  /* Rules Engine */
  .rules-engine {
    padding: 24px;
    display: flex;
    flex-direction: column;
    gap: 20px;
    border-color: rgba(178, 102, 255, 0.2);
  }
  .rules-header {
    display: flex;
    align-items: center;
    gap: 12px;
  }
  .rules-header h3 { color: var(--text-primary); }
  
  .rule-builder {
    display: flex;
    align-items: center;
    gap: 12px;
    flex-wrap: wrap;
    background: rgba(0,0,0,0.4);
    padding: 16px;
    border-radius: 8px;
    border: 1px solid var(--bg-panel-border);
    font-family: var(--font-mono);
    color: var(--text-secondary);
  }
  .rule-select, .rule-input {
    background: var(--bg-base);
    color: var(--text-primary);
    border: 1px solid var(--bg-panel-border);
    padding: 8px 12px;
    border-radius: 6px;
    font-family: inherit;
    font-size: 0.9rem;
  }
  .rule-select:focus, .rule-input:focus {
    outline: none;
    border-color: var(--accent-cyan);
  }
  .rule-select.action {
    color: var(--accent-orange);
  }
  .rule-input { width: 80px; }

  .rule-actions {
    display: flex;
    justify-content: flex-end;
  }
  .save-btn {
    background: var(--accent-cyan-dim);
    color: var(--accent-cyan);
    border: 1px solid var(--accent-cyan);
    padding: 10px 20px;
    border-radius: 6px;
    font-weight: 600;
  }
  .save-btn:hover {
    background: var(--accent-cyan);
    color: #000;
  }
</style>
