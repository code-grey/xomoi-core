<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import MetricCard from '../MetricCard.svelte';
  import HistoricalExplorer from './HistoricalExplorer.svelte';
  import { Thermometer, Droplets, Gauge, Network } from 'lucide-svelte';

  let temperature = $state('0.0');
  let humidity = $state('0.0');
  let pressure = $state('0.0');
  let activeDevices = $state('0');

  // Interactive Chart State
  let activeMetric = $state('temperature');
  
  let tempHistory = $state<number[]>(Array(40).fill(24)); 
  let humHistory = $state<number[]>(Array(40).fill(45)); 
  let pressHistory = $state<number[]>(Array(40).fill(1012)); 

  let ws: WebSocket;

  onMount(() => {
    ws = new WebSocket('ws://localhost:8085/api/v1/ws/telemetry');
    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      temperature = data.temperature.toFixed(1);
      humidity = data.humidity.toFixed(1);
      pressure = data.pressure.toFixed(1);
      activeDevices = data.active_devices.toString();

      tempHistory = [...tempHistory.slice(1), data.temperature];
      humHistory = [...humHistory.slice(1), data.humidity];
      pressHistory = [...pressHistory.slice(1), data.pressure];
    };
  });

  onDestroy(() => {
    if (ws) ws.close();
  });

  function buildPoints(history: number[], min: number, max: number) {
    return history.map((val, i) => {
      const x = (i / 39) * 100;
      const clampedVal = Math.max(min, Math.min(max, val));
      const y = 100 - ((clampedVal - min) / (max - min)) * 100; 
      return `${x},${y}`;
    }).join(' ');
  }

  let tempPoints = $derived(buildPoints(tempHistory, 20, 30));
  let humPoints = $derived(buildPoints(humHistory, 40, 50));
  let pressPoints = $derived(buildPoints(pressHistory, 1010, 1015));

  let activeColor = $derived(
    activeMetric === 'temperature' ? 'var(--accent-orange)' :
    activeMetric === 'humidity' ? 'var(--accent-cyan)' : 'var(--accent-purple)'
  );

  // Mocking multiple sensors to demonstrate the Grid UI
  let mockSensors = $derived(
    activeMetric === 'temperature' ? [
      { id: 's1', name: 'Garage Ambient', val: temperature, hist: tempHistory, min: 20, max: 30, unit: '°C' },
      { id: 's2', name: 'Living Room', val: (parseFloat(temperature) + 1.5).toFixed(1), hist: tempHistory.map(v => v + 1.5), min: 20, max: 30, unit: '°C' },
      { id: 's3', name: 'Outdoor Node', val: (parseFloat(temperature) - 8.2).toFixed(1), hist: tempHistory.map(v => v - 8.2), min: 10, max: 30, unit: '°C' }
    ] : activeMetric === 'humidity' ? [
      { id: 'h1', name: 'Garage Ambient', val: humidity, hist: humHistory, min: 40, max: 60, unit: '%' },
      { id: 'h2', name: 'Greenhouse', val: (parseFloat(humidity) + 25.0).toFixed(1), hist: humHistory.map(v => v + 25.0), min: 40, max: 100, unit: '%' }
    ] : [
      { id: 'p1', name: 'Basement', val: pressure, hist: pressHistory, min: 1010, max: 1020, unit: 'hPa' }
    ]
  );

  let selectedSensor = $state<any>(null);
</script>

<div class="view-container">
  <div class="metrics-grid">
    <MetricCard title="Temperature" value={temperature} unit="°C" Icon={Thermometer} sparkline={tempPoints} active={activeMetric === 'temperature'} onclick={() => activeMetric = 'temperature'} />
    <MetricCard title="Humidity" value={humidity} unit="%" Icon={Droplets} sparkline={humPoints} active={activeMetric === 'humidity'} onclick={() => activeMetric = 'humidity'} />
    <MetricCard title="Pressure" value={pressure} unit="hPa" Icon={Gauge} sparkline={pressPoints} active={activeMetric === 'pressure'} onclick={() => activeMetric = 'pressure'} />
    <MetricCard title="Active Devices" value={activeDevices} unit="Sensors" Icon={Network} onclick={() => window.location.hash = 'fleet'} />
  </div>

  <div class="content-layout">
    <!-- Grid of all sensors reporting the active metric -->
    <div class="charts-grid">
      {#each mockSensors as sensor (sensor.id)}
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
    padding-right: 8px; /* Scrollbar space */
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
</style>
