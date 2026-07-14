<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import MetricCard from '../MetricCard.svelte';
  import HistoricalExplorer from './HistoricalExplorer.svelte';
  import { Thermometer, Droplets, Network } from 'lucide-svelte';
  import { globalState } from '../store.svelte';

  let activeMetric = $state('temperature');

  let fleet = $derived(globalState.fleet);
  let activeDevices = $derived(fleet.length.toString());

  // Aggregate current values from the real fleet
  let tempDevices = $derived(fleet.filter(d => d.temp !== undefined && d.temp !== 0));
  let humDevices = $derived(fleet.filter(d => d.hum !== undefined && d.hum !== 0));

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

  // Use the first valid sensor's history for the top-level sparklines
  let tempHistory = $derived(tempDevices.length > 0 ? tempDevices[0].tempHistory : Array(40).fill(0));
  let humHistory = $derived(humDevices.length > 0 ? humDevices[0].humHistory : Array(40).fill(0)); 

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

  let activeColor = $derived(
    activeMetric === 'temperature' ? 'var(--accent-orange)' :
    activeMetric === 'humidity' ? 'var(--accent-cyan)' : 'var(--accent-purple)'
  );

  // Map the real WebRTC fleet into the chart components
  let realSensors = $derived(
    activeMetric === 'temperature' ? tempDevices.map(d => ({
      id: d.id, name: d.friendlyName || d.id, val: d.temp.toFixed(1), hist: d.tempHistory, min: 0, max: 50, unit: '°C'
    })) : humDevices.map(d => ({
      id: d.id, name: d.friendlyName || d.id, val: d.hum.toFixed(1), hist: d.humHistory, min: 0, max: 100, unit: '%'
    }))
  );

  let selectedSensor = $state<any>(null);
</script>

{#if fleet.length > 0}
  <div class="view-container">
    <div class="metrics-grid">
      <MetricCard title="Temperature" value={temperature} unit="°C" Icon={Thermometer} sparkline={tempPoints} active={activeMetric === 'temperature'} onclick={() => activeMetric = 'temperature'} />
      <MetricCard title="Humidity" value={humidity} unit="%" Icon={Droplets} sparkline={humPoints} active={activeMetric === 'humidity'} onclick={() => activeMetric = 'humidity'} />
      <MetricCard title="Active Devices" value={activeDevices} unit="Sensors" Icon={Network} onclick={() => window.location.hash = 'fleet'} />
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
