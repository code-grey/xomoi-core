<!--
Xomoi-Core: Sovereign Edge Node
Copyright (C) 2026 Adrish Bora (@code-grey) & Simanjit Hujuri (@code-zephyrus)

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
-->

<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import MetricCard from '../MetricCard.svelte';
  import HistoricalExplorer from './HistoricalExplorer.svelte';
  import { Thermometer, Droplets, Network, Gauge, Zap, Sun, Fan, Activity, Power, Lightbulb } from 'lucide-svelte';
  import { globalState } from '../store.svelte';

  let activeMetric = $state('temperature');

  // RPC Demo State
  let rpcMainPower = $state(true);
  let rpcAuxPower = $state(false);
  let rpcFanSpeed = $state(45);
  let rpcIrrigation = $state(false);
  let rpcDimmer = $state(80);

  let fleet = $derived(globalState.fleet);
  let activeDevices = $derived(fleet.length.toString());

  // Aggregate current values from the real fleet
  let tempDevices = $derived(fleet.filter(d => d.temp !== undefined && d.temp !== 0));
  let humDevices = $derived(fleet.filter(d => d.hum !== undefined && d.hum !== 0));
  let pressDevices = $derived(fleet.filter(d => d.pressure !== undefined && d.pressure !== 0));
  let voltDevices = $derived(fleet.filter(d => d.voltage !== undefined && d.voltage !== 0));
  let luxDevices = $derived(fleet.filter(d => d.lux !== undefined && d.lux !== 0));
  let fanDevices = $derived(fleet.filter(d => d.fan_speed !== undefined && d.fan_speed !== 0));
  let imuDevices = $derived(fleet.filter(d => d.imu_hz !== undefined && d.imu_hz !== 0));

  let temperature = $derived(
    tempDevices.length > 0 
      ? (tempDevices.reduce((acc, d) => acc + d.temp, 0) / tempDevices.length).toFixed(1)
      : '0.0'
  );

  let humidity = $derived(
    humDevices.length > 0 
      ? (humDevices.reduce((acc, d) => acc + d.hum, 0) / humDevices.length).toFixed(1)
      : '0.0'
  );

  let pressure = $derived(pressDevices.length > 0 ? (pressDevices.reduce((acc, d) => acc + d.pressure, 0) / pressDevices.length).toFixed(1) : '0.0');
  let voltage = $derived(voltDevices.length > 0 ? (voltDevices.reduce((acc, d) => acc + d.voltage, 0) / voltDevices.length).toFixed(2) : '0.00');
  let lux = $derived(luxDevices.length > 0 ? Math.floor(luxDevices.reduce((acc, d) => acc + d.lux, 0) / luxDevices.length).toString() : '0');
  let fan_speed = $derived(fanDevices.length > 0 ? Math.floor(fanDevices.reduce((acc, d) => acc + d.fan_speed, 0) / fanDevices.length).toString() : '0');
  let imu_hz = $derived(imuDevices.length > 0 ? Math.floor(imuDevices.reduce((acc, d) => acc + d.imu_hz, 0) / imuDevices.length).toString() : '0');

  // Use the first valid sensor's history for the top-level sparklines
  let tempHistory = $derived(tempDevices.length > 0 ? tempDevices[0].tempHistory : Array(40).fill(0));
  let humHistory = $derived(humDevices.length > 0 ? humDevices[0].humHistory : Array(40).fill(0));
  let pressHistory = $derived(pressDevices.length > 0 ? pressDevices[0].pressureHistory : Array(40).fill(0));
  let voltHistory = $derived(voltDevices.length > 0 ? voltDevices[0].voltageHistory : Array(40).fill(0));
  let luxHistory = $derived(luxDevices.length > 0 ? luxDevices[0].luxHistory : Array(40).fill(0));
  let fanHistory = $derived(fanDevices.length > 0 ? fanDevices[0].fanSpeedHistory : Array(40).fill(0));
  let imuHistory = $derived(imuDevices.length > 0 ? imuDevices[0].imuHistory : Array(40).fill(0)); 

  function buildPoints(history: number[], min: number, max: number) {
    if (!history) return '';
    return history.map((val, i) => {
      const x = (i / 39) * 100;
      const clampedVal = Math.max(min, Math.min(max, val));
      const y = 100 - ((clampedVal - min) / (max - min)) * 100; 
      return `${x},${y}`;
    }).join(' ');
  }

  let tempPoints = $derived(buildPoints(tempHistory, 0, 50));
  let humPoints = $derived(buildPoints(humHistory, 0, 100));
  let pressPoints = $derived(buildPoints(pressHistory, 900, 1100));
  let voltPoints = $derived(buildPoints(voltHistory, 0, 5));
  let luxPoints = $derived(buildPoints(luxHistory, 0, 1000));
  let fanPoints = $derived(buildPoints(fanHistory, 0, 3000));
  let imuPoints = $derived(buildPoints(imuHistory, 0, 2000));

  let activeColor = $derived(
    activeMetric === 'temperature' ? 'var(--accent-orange)' :
    activeMetric === 'humidity' ? 'var(--accent-cyan)' :
    activeMetric === 'pressure' ? 'var(--accent-purple)' :
    activeMetric === 'voltage' ? '#eccc68' :
    activeMetric === 'lux' ? '#ffeaa7' :
    activeMetric === 'fan_speed' ? '#7bed9f' :
    '#ff4757' // imu_hz
  );

  // Map the real WebRTC fleet into the chart components
  let realSensors = $derived(
    activeMetric === 'temperature' ? tempDevices.map(d => ({ id: d.id, name: d.friendlyName || d.id, val: d.temp.toFixed(1), hist: d.tempHistory, min: 0, max: 50, unit: '°C' })) :
    activeMetric === 'humidity' ? humDevices.map(d => ({ id: d.id, name: d.friendlyName || d.id, val: d.hum.toFixed(1), hist: d.humHistory, min: 0, max: 100, unit: '%' })) :
    activeMetric === 'pressure' ? pressDevices.map(d => ({ id: d.id, name: d.friendlyName || d.id, val: d.pressure.toFixed(1), hist: d.pressureHistory, min: 900, max: 1100, unit: 'hPa' })) :
    activeMetric === 'voltage' ? voltDevices.map(d => ({ id: d.id, name: d.friendlyName || d.id, val: d.voltage.toFixed(2), hist: d.voltageHistory, min: 0, max: 5, unit: 'V' })) :
    activeMetric === 'lux' ? luxDevices.map(d => ({ id: d.id, name: d.friendlyName || d.id, val: d.lux, hist: d.luxHistory, min: 0, max: 1000, unit: 'lx' })) :
    activeMetric === 'fan_speed' ? fanDevices.map(d => ({ id: d.id, name: d.friendlyName || d.id, val: d.fan_speed, hist: d.fanSpeedHistory, min: 0, max: 3000, unit: 'RPM' })) :
    imuDevices.map(d => ({ id: d.id, name: d.friendlyName || d.id, val: d.imu_hz, hist: d.imuHistory, min: 0, max: 2000, unit: 'Hz' }))
  );

  let selectedSensor = $state<any>(null);
</script>

{#if fleet.length > 0}
  <div class="view-container">
    <div class="metrics-grid">
      <MetricCard title="Temperature" value={temperature} unit="°C" Icon={Thermometer} sparkline={tempPoints} active={activeMetric === 'temperature'} onclick={() => activeMetric = 'temperature'} />
      <MetricCard title="Humidity" value={humidity} unit="%" Icon={Droplets} sparkline={humPoints} active={activeMetric === 'humidity'} onclick={() => activeMetric = 'humidity'} />
      <MetricCard title="Pressure" value={pressure} unit="hPa" Icon={Gauge} sparkline={pressPoints} active={activeMetric === 'pressure'} onclick={() => activeMetric = 'pressure'} />
      <MetricCard title="Voltage" value={voltage} unit="V" Icon={Zap} sparkline={voltPoints} active={activeMetric === 'voltage'} onclick={() => activeMetric = 'voltage'} />
      <MetricCard title="Ambient Light" value={lux} unit="lx" Icon={Sun} sparkline={luxPoints} active={activeMetric === 'lux'} onclick={() => activeMetric = 'lux'} />
      <MetricCard title="Fan Speed" value={fan_speed} unit="RPM" Icon={Fan} sparkline={fanPoints} active={activeMetric === 'fan_speed'} onclick={() => activeMetric = 'fan_speed'} />
      <MetricCard title="IMU Polling" value={imu_hz} unit="Hz" Icon={Activity} sparkline={imuPoints} active={activeMetric === 'imu_hz'} onclick={() => activeMetric = 'imu_hz'} />
      <MetricCard title="Active Devices" value={activeDevices} unit="Sensors" Icon={Network} onclick={() => window.location.hash = 'fleet'} />
    </div>

    <!-- RPC Remote Controls -->
    <div class="rpc-controls-panel glass-panel">
      <div class="rpc-panel-header">
        <h3 class="panel-title">Quick Actions (RPC)</h3>
        <span class="rpc-status glow-dot green"></span>
      </div>
      <div class="rpc-cards-grid">
        <div class="rpc-card {rpcMainPower ? 'active' : ''}">
          <div class="rpc-header"><Power size={18} /> Main Relay</div>
          <button class="rpc-toggle {rpcMainPower ? 'on' : 'off'}" onclick={() => rpcMainPower = !rpcMainPower}>
            {rpcMainPower ? 'ON' : 'OFF'}
          </button>
        </div>
        <div class="rpc-card {rpcAuxPower ? 'active' : ''}">
          <div class="rpc-header"><Zap size={18} /> Aux Power</div>
          <button class="rpc-toggle {rpcAuxPower ? 'on' : 'off'}" onclick={() => rpcAuxPower = !rpcAuxPower}>
            {rpcAuxPower ? 'ON' : 'OFF'}
          </button>
        </div>
        <div class="rpc-card {rpcIrrigation ? 'active' : ''}">
          <div class="rpc-header"><Droplets size={18} /> Irrigation</div>
          <button class="rpc-toggle {rpcIrrigation ? 'on' : 'off'}" onclick={() => rpcIrrigation = !rpcIrrigation}>
            {rpcIrrigation ? 'ON' : 'OFF'}
          </button>
        </div>
        <div class="rpc-card slider-card">
          <div class="rpc-header"><Fan size={18} /> HVAC Speed <span class="rpc-val">{rpcFanSpeed}%</span></div>
          <input type="range" min="0" max="100" bind:value={rpcFanSpeed} class="rpc-slider" />
        </div>
        <div class="rpc-card slider-card">
          <div class="rpc-header"><Lightbulb size={18} /> Dimmer <span class="rpc-val">{rpcDimmer}%</span></div>
          <input type="range" min="0" max="100" bind:value={rpcDimmer} class="rpc-slider" />
        </div>
      </div>
    </div>

    <div class="content-layout">
      <!-- Grid of all sensors reporting the active metric -->
      <div class="charts-grid">
        {#each realSensors as sensor (sensor.id)}
          <!-- svelte-ignore a11y_click_events_have_key_events -->
          <!-- svelte-ignore a11y_no_static_element_interactions -->
          <div class="mini-chart-panel glass-panel interactive" style="--chart-color: {activeColor}" onclick={() => selectedSensor = sensor}>
            <div class="mini-header">
              <span class="sensor-name">{sensor.name}</span>
              <span class="sensor-val" style="color: var(--chart-color)">{sensor.val}{sensor.unit}</span>
            </div>
            <div class="mini-body">
              <svg viewBox="0 0 100 100" preserveAspectRatio="none">
                <defs>
                  <linearGradient id="grad-{sensor.id}" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="0%" stop-color="var(--chart-color)" stop-opacity="0.4" />
                    <stop offset="100%" stop-color="var(--chart-color)" stop-opacity="0.0" />
                  </linearGradient>
                </defs>
                <polygon points="0,100 {buildPoints(sensor.hist, sensor.min, sensor.max)} 100,100" fill="url(#grad-{sensor.id})" />
                <polyline points={buildPoints(sensor.hist, sensor.min, sensor.max)} fill="none" stroke="var(--chart-color)" stroke-width="2.5" vector-effect="non-scaling-stroke" />
              </svg>
            </div>
          </div>
        {/each}
      </div>

      <!-- Alert History Sidebar -->
      <div class="alerts-sidebar glass-panel">
        <h3 class="sidebar-title">Alert History</h3>
        <div class="alerts-list">
          <div class="alert-item critical">
            <div class="alert-time">09:12 AM</div>
            <div class="alert-msg">Outdoor Node offline</div>
          </div>
          <div class="alert-item warning">
            <div class="alert-time">14:32 PM</div>
            <div class="alert-msg">Greenhouse {activeMetric} spike</div>
          </div>
          <div class="alert-item info">
            <div class="alert-time">18:05 PM</div>
            <div class="alert-msg">System OTA Success</div>
          </div>
        </div>
      </div>
    </div>
  </div>
{:else}
  <div class="empty-state glass-panel">
    <div class="empty-icon glow-orb"></div>
    <h2>No Devices Detected</h2>
    <p>Add devices to see what they are doing!</p>
    <button class="btn primary" onclick={() => window.location.hash = 'fleet'}>Go to Device Fleet</button>
  </div>
{/if}

{#if selectedSensor}
  <HistoricalExplorer 
    sensor={selectedSensor}
    color={activeColor}
    onclose={() => selectedSensor = null}
  />
{/if}

<style>
  .view-container {
    display: flex;
    flex-direction: column;
    gap: 24px;
    height: 100%;
  }

  .metrics-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 24px;
  }

  /* RPC Controls */
  .rpc-controls-panel {
    padding: 20px;
  }
  .rpc-panel-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;
    border-bottom: 1px solid var(--bg-panel-border);
    padding-bottom: 12px;
  }
  .panel-title {
    color: var(--text-primary);
    font-size: 1.1rem;
    font-weight: 500;
    margin: 0;
  }
  .glow-dot.green {
    width: 8px; height: 8px;
    border-radius: 50%;
    background: #00ff66;
    box-shadow: 0 0 8px #00ff66;
  }
  .rpc-cards-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
    gap: 16px;
  }
  .rpc-card {
    background: rgba(0,0,0,0.2);
    border: 1px solid var(--bg-panel-border);
    border-radius: 12px;
    padding: 16px;
    display: flex;
    flex-direction: column;
    gap: 16px;
    transition: all 0.2s ease;
  }
  .rpc-card.active {
    border-color: rgba(0, 255, 102, 0.4);
    box-shadow: 0 4px 12px rgba(0, 255, 102, 0.05);
  }
  .rpc-header {
    display: flex;
    align-items: center;
    gap: 8px;
    color: var(--text-secondary);
    font-size: 0.9rem;
    font-weight: 600;
  }
  .rpc-toggle {
    background: transparent;
    border: 2px solid var(--bg-panel-border);
    color: var(--text-secondary);
    padding: 8px;
    border-radius: 8px;
    font-weight: 700;
    cursor: pointer;
    transition: all 0.2s;
  }
  .rpc-toggle.on {
    background: rgba(0, 255, 102, 0.1);
    border-color: #00ff66;
    color: #00ff66;
    text-shadow: 0 0 8px rgba(0,255,102,0.5);
  }
  .rpc-toggle:hover {
    border-color: var(--text-secondary);
  }
  .rpc-toggle.on:hover {
    border-color: #00cc52;
  }
  .rpc-val {
    margin-left: auto;
    color: var(--accent-cyan);
  }
  .rpc-slider {
    width: 100%;
    accent-color: var(--accent-cyan);
    cursor: pointer;
  }

  .content-layout {
    display: flex;
    gap: 24px;
    flex-grow: 1;
    min-height: 0; /* Important for flex children scrolling */
  }

  .charts-grid {
    flex-grow: 1;
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: 20px;
    align-content: flex-start;
    overflow-y: auto;
    padding: 12px 12px 12px 4px; /* Space for hover shadow/transform */
    margin: -12px -12px -12px -4px; /* Offset padding to maintain alignment */
  }

  .mini-chart-panel {
    display: flex;
    flex-direction: column;
    height: 180px;
    padding: 16px;
  }

  .interactive {
    cursor: pointer;
    transition: transform 0.2s, box-shadow 0.2s;
  }
  .interactive:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.4);
    border-color: var(--chart-color);
  }

  .mini-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 12px;
  }

  .sensor-name {
    color: var(--text-secondary);
    font-size: 0.9rem;
    font-weight: 500;
  }

  .sensor-val {
    font-size: 1.2rem;
    font-weight: 700;
    font-family: var(--font-mono);
  }

  .mini-body {
    flex-grow: 1;
    position: relative;
    overflow: hidden;
    border-radius: 4px;
  }

  svg {
    width: 100%;
    height: 100%;
    display: block;
    overflow: visible;
  }

  /* Sidebar */
  .alerts-sidebar {
    width: 300px;
    flex-shrink: 0;
    padding: 20px;
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .sidebar-title {
    color: var(--text-primary);
    font-size: 1.1rem;
    font-weight: 500;
    margin: 0;
    padding-bottom: 12px;
    border-bottom: 1px solid var(--bg-panel-border);
  }

  .alerts-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
    overflow-y: auto;
  }

  .alert-item {
    display: flex;
    flex-direction: column;
    gap: 4px;
    padding: 12px;
    background: rgba(255, 255, 255, 0.02);
    border-radius: 8px;
    border-left: 3px solid transparent;
  }

  .alert-item.critical { border-left-color: #ff4757; }
  .alert-item.warning { border-left-color: var(--accent-orange); }
  .alert-item.info { border-left-color: var(--accent-cyan); }

  .alert-time {
    font-size: 0.75rem;
    color: var(--text-secondary);
    font-family: var(--font-mono);
  }

  .alert-msg {
    font-size: 0.85rem;
    color: var(--text-primary);
  }

  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    min-height: 400px;
    text-align: center;
    gap: 16px;
  }

  .empty-icon {
    width: 24px;
    height: 24px;
    border-radius: 50%;
    background: var(--accent-cyan);
    box-shadow: 0 0 24px var(--accent-cyan);
    margin-bottom: 16px;
  }

  .empty-state h2 {
    font-size: 1.5rem;
    color: var(--text-primary);
    margin: 0;
  }

  .empty-state p {
    color: var(--text-secondary);
    font-size: 1rem;
    margin: 0 0 24px 0;
  }

  .btn {
    padding: 10px 24px;
    border: none;
    border-radius: 6px;
    font-family: inherit;
    font-size: 0.95rem;
    font-weight: 600;
    cursor: pointer;
    transition: var(--transition-smooth);
  }

  .btn.primary {
    background: rgba(0, 255, 204, 0.1);
    color: var(--accent-cyan);
    border: 1px solid rgba(0, 255, 204, 0.2);
  }

  .btn.primary:hover {
    background: rgba(0, 255, 204, 0.2);
    box-shadow: 0 0 12px rgba(0, 255, 204, 0.3);
  }

  /* Mobile Responsive */
  @media (max-width: 768px) {
    .content-layout {
      flex-direction: column;
    }
    .alerts-sidebar {
      width: 100%;
      border-left: none;
      border-top: 1px solid var(--bg-panel-border);
    }
    .charts-grid {
      grid-template-columns: 1fr;
    }
  }
</style>
