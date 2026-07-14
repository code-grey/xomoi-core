<script lang="ts">
  import { ShieldAlert, CheckCircle, Activity, Cpu, Thermometer, Droplets, Zap, Settings, UploadCloud } from 'lucide-svelte';
  import WebFlasher from '../WebFlasher.svelte';
  import { globalState } from '../store.svelte';
  
  let selectedDevice = $state<any>(null);
  let showFlasher = $state(false);

  // OTA Upload State
  let otaFile = $state<File | null>(null);
  let otaStatus = $state<'idle' | 'uploading' | 'success' | 'error'>('idle');
  let otaMessage = $state('');

  // Rules Engine State
  let rules = $state<any[]>([]);
  let newRuleTag = $state('temp');
  let newRuleCondition = $state('>');
  let newRuleThreshold = $state(30);

  // Fetch rules whenever the selected device changes
  $effect(() => {
    if (selectedDevice) {
      fetchRules();
    } else {
      rules = [];
    }
  });

  async function fetchRules() {
    if (!selectedDevice) return;
    try {
      const res = await fetch(`/api/v1/devices/${selectedDevice.id}/rules`);
      if (res.ok) {
        rules = await res.json() || [];
      }
    } catch (e) {
      console.error("Failed to fetch rules", e);
    }
  }

  async function addRule() {
    if (!selectedDevice) return;
    try {
      const payload = {
        tag_name: newRuleTag,
        condition: newRuleCondition,
        threshold: newRuleThreshold
      };
      const res = await fetch(`/api/v1/devices/${selectedDevice.id}/rules`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      });
      if (res.ok) {
        await fetchRules(); // Refresh list
      } else {
        alert("Failed to add rule");
      }
    } catch (e) {
      alert("Network error while adding rule");
    }
  }

  async function deleteRule(id: string) {
    try {
      const res = await fetch(`/api/v1/rules/${id}`, { method: 'DELETE' });
      if (res.ok) {
        rules = rules.filter(r => r.id !== id);
      }
    } catch (e) {
      console.error("Failed to delete rule", e);
    }
  }

  function getStatusColor(status: string) {
    return status === 'alert' ? 'var(--accent-orange)' : status === 'offline' ? 'var(--text-secondary)' : 'var(--accent-cyan)';
  }

  // --- RPC Command Execution ---
  async function toggleRelay() {
    if (!selectedDevice || selectedDevice.type !== 'Relay Switch') return;
    
    // Optimistic UI update (simulate loading/pending state)
    const previousState = selectedDevice.state;
    selectedDevice.state = 'PENDING...';

    // Real Hardware Execution Flow
    try {
      const payload = {
        command: "toggle_relay",
        params: { pin: 4 },
        retain: false // Critical: Physical actuations must not be retained if device is offline
      };

      const res = await fetch(`/api/v1/devices/${selectedDevice.id}/rpc`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      });

      if (!res.ok) throw new Error("Backend rejected command");

      // In real mode, we DO NOT revert the UI here. We leave it as 'PENDING...'
      // and wait for the real WebRTC tunnel to deliver the `/xomoi/mac/rpc/ack` 
      // payload which updates `globalState.fleet` automatically.

    } catch (err) {
      // Revert if network fails
      selectedDevice.state = previousState;
      alert("Failed to send RPC command to Edge Node");
    }
  }

  // --- OTA Firmware Execution ---
  async function handleOTAUpload() {
    if (!otaFile || !selectedDevice) return;
    
    otaStatus = 'uploading';
    otaMessage = 'Pushing firmware to Edge Node...';
    
    const formData = new FormData();
    formData.append('firmware', otaFile);
    
    try {
      const res = await fetch(`/api/v1/devices/${selectedDevice.id}/ota`, {
        method: 'POST',
        body: formData
      });
      
      if (res.ok) {
        otaStatus = 'success';
        otaMessage = 'OTA Triggered! Device is flashing and rebooting.';
        setTimeout(() => { otaStatus = 'idle'; otaFile = null; }, 5000);
      } else {
        throw new Error('Upload failed');
      }
    } catch (err) {
      otaStatus = 'error';
      otaMessage = 'Failed to push OTA update.';
    }
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
      {#each globalState.fleet as device}
        <button 
          class="device-node glass-panel {device.status}" 
          onclick={() => selectedDevice = device}
        >
          <div class="node-icon">
            {#if device.status === 'alert'}
              <ShieldAlert color="var(--accent-orange)" size={28} />
            {:else if device.status === 'offline'}
              <ShieldAlert color="var(--text-secondary)" size={28} />
            {:else}
              <CheckCircle color="var(--accent-cyan)" size={28} />
            {/if}
          </div>
          <div class="node-info">
            <span class="node-id mono">{device.friendlyName || device.id}</span>
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
          {#if selectedDevice.status === 'alert' || selectedDevice.status === 'offline'} <ShieldAlert size={16} /> {:else} <CheckCircle size={16} /> {/if}
          {selectedDevice.status.toUpperCase()}
        </div>
      </div>
      
      <div class="profile-card glass-panel">
        <div class="device-identity">
          <div class="avatar" style="border-color: {getStatusColor(selectedDevice.status)}">
            <Cpu size={32} color={getStatusColor(selectedDevice.status)} />
          </div>
          <div class="identity-text">
            <h2 class="mono">{selectedDevice.friendlyName || selectedDevice.id}</h2>
            <p>MAC: {selectedDevice.id} • {selectedDevice.type} • {selectedDevice.location}</p>
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
          <button class="sensor-widget glass-panel interactive-relay" onclick={toggleRelay}>
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

        <div class="existing-rules">
          {#each rules as rule}
            <div class="rule-item">
              <span class="rule-text mono">IF {rule.tag_name} {rule.condition} {rule.threshold} THEN Trigger Alert</span>
              <button class="delete-btn" onclick={() => deleteRule(rule.id)}>Delete</button>
            </div>
          {/each}
        </div>

        <div class="rule-builder">
          <span>IF</span>
          <select class="rule-select" bind:value={newRuleTag}>
            <option value="temp">Temperature</option>
            <option value="hum">Humidity</option>
            <option value="state">Relay/PIR State</option>
          </select>
          <select class="rule-select" bind:value={newRuleCondition}>
            <option value=">">&gt;</option>
            <option value="<">&lt;</option>
            <option value="==">==</option>
            <option value="!=">!=</option>
          </select>
          <input type="number" bind:value={newRuleThreshold} class="rule-input mono" />
          <span>THEN</span>
          <select class="rule-select action">
            <option>Trigger Alert Status</option>
          </select>
        </div>
        <div class="rule-actions">
          <button class="save-btn" onclick={addRule}>+ Add Rule</button>
        </div>
      </div>

      <!-- 4. OTA (OVER-THE-AIR) UPDATE ENGINE -->
      <div class="ota-engine glass-panel">
        <div class="ota-header">
          <UploadCloud size={20} color="var(--accent-cyan)" />
          <h3>Over-The-Air (OTA) Update</h3>
        </div>
        <p class="ota-desc">Push a new binary payload directly to this device over the sovereign network.</p>
        
        <div class="ota-controls">
          <input 
            type="file" 
            accept=".bin" 
            id="ota-file" 
            class="file-input" 
            onchange={(e) => otaFile = e.currentTarget.files?.[0] || null}
          />
          <label for="ota-file" class="file-label">
            {otaFile ? otaFile.name : 'Select .bin Firmware'}
          </label>
          
          <button 
            class="upload-btn" 
            disabled={!otaFile || otaStatus === 'uploading'} 
            onclick={handleOTAUpload}
          >
            {#if otaStatus === 'uploading'}
              Deploying...
            {:else}
              Flash Firmware
            {/if}
          </button>
        </div>
        
        {#if otaStatus !== 'idle'}
          <div class="ota-status {otaStatus}">
            {otaMessage}
          </div>
        {/if}
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
    transition: all 0.3s ease;
  }
  .device-node:hover {
    transform: translateY(-4px);
    background: rgba(255,255,255,0.05);
  }
  
  .device-node.healthy {
    border-color: rgba(0, 255, 204, 0.2);
    box-shadow: inset 0 0 20px rgba(0, 255, 204, 0.02), 0 0 15px rgba(0, 255, 204, 0.05);
  }
  .device-node.healthy:hover {
    border-color: rgba(0, 255, 204, 0.5);
    box-shadow: inset 0 0 20px rgba(0, 255, 204, 0.05), 0 8px 30px rgba(0, 255, 204, 0.2);
  }

  .device-node.alert {
    border-color: rgba(255, 85, 0, 0.4);
    background: rgba(255, 85, 0, 0.05);
    box-shadow: inset 0 0 20px rgba(255, 85, 0, 0.05), 0 0 15px rgba(255, 85, 0, 0.1);
  }
  .device-node.alert:hover {
    border-color: rgba(255, 85, 0, 0.8);
    box-shadow: inset 0 0 20px rgba(255, 85, 0, 0.1), 0 8px 30px rgba(255, 85, 0, 0.3);
  }
  .device-node.offline {
    border-color: rgba(255, 255, 255, 0.1);
    opacity: 0.6;
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
  .status-badge.offline { color: var(--text-secondary); border-color: var(--text-secondary); }

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
  
  .existing-rules {
    display: flex;
    flex-direction: column;
    gap: 8px;
    margin-bottom: 8px;
  }
  .rule-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    background: rgba(255, 255, 255, 0.03);
    padding: 12px 16px;
    border-radius: 6px;
    border: 1px solid rgba(255, 255, 255, 0.05);
  }
  .rule-text { color: var(--accent-cyan); font-size: 0.9rem; font-weight: 600; }
  .delete-btn {
    color: var(--accent-orange);
    background: rgba(255, 85, 0, 0.1);
    border: 1px solid rgba(255, 85, 0, 0.3);
    padding: 4px 12px;
    border-radius: 4px;
    font-size: 0.8rem;
    cursor: pointer;
  }
  .delete-btn:hover { background: rgba(255, 85, 0, 0.2); }

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

  /* OTA Engine */
  .ota-engine {
    padding: 24px;
    display: flex;
    flex-direction: column;
    gap: 16px;
    border-color: rgba(0, 255, 204, 0.2);
  }
  .ota-header {
    display: flex;
    align-items: center;
    gap: 12px;
  }
  .ota-header h3 { color: var(--text-primary); }
  .ota-desc { color: var(--text-secondary); font-size: 0.9rem; }
  
  .ota-controls {
    display: flex;
    gap: 16px;
    align-items: center;
  }
  
  .file-input { display: none; }
  .file-label {
    background: rgba(0,0,0,0.4);
    border: 1px dashed var(--accent-cyan-dim);
    color: var(--text-secondary);
    padding: 10px 20px;
    border-radius: 6px;
    cursor: pointer;
    font-family: var(--font-mono);
    font-size: 0.9rem;
    flex-grow: 1;
    text-align: center;
    transition: all 0.2s;
  }
  .file-label:hover {
    background: rgba(0, 255, 204, 0.05);
    color: var(--accent-cyan);
  }
  
  .upload-btn {
    background: var(--accent-cyan-dim);
    color: var(--accent-cyan);
    border: 1px solid var(--accent-cyan);
    padding: 10px 24px;
    border-radius: 6px;
    font-weight: 600;
    cursor: pointer;
    white-space: nowrap;
  }
  .upload-btn:hover:not(:disabled) {
    background: var(--accent-cyan);
    color: #000;
  }
  .upload-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
  
  .ota-status {
    padding: 12px;
    border-radius: 6px;
    font-size: 0.9rem;
    font-weight: 600;
    text-align: center;
  }
  .ota-status.uploading { background: rgba(255, 255, 255, 0.1); color: var(--text-primary); }
  .ota-status.success { background: rgba(0, 255, 204, 0.1); color: var(--accent-cyan); }
  .ota-status.error { background: rgba(255, 85, 0, 0.1); color: var(--accent-orange); }

  /* Mobile Responsive */
  @media (max-width: 768px) {
    .matrix-header {
      flex-direction: column;
      align-items: flex-start;
      gap: 16px;
    }
    .header-actions {
      width: 100%;
      flex-wrap: wrap;
    }
    .matrix-grid {
      grid-template-columns: 1fr; /* Single column on mobile */
    }
    .drill-header {
      flex-direction: column;
      align-items: flex-start;
      gap: 12px;
    }
    .profile-card {
      flex-direction: column;
      gap: 16px;
      align-items: flex-start;
    }
    .rule-builder {
      flex-direction: column;
      align-items: stretch;
    }
    .ota-controls {
      flex-direction: column;
      align-items: stretch;
    }
  }
</style>
