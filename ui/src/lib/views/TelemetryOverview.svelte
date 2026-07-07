<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import MetricCard from '../MetricCard.svelte';
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

  let activePoints = $derived(
    activeMetric === 'temperature' ? tempPoints :
    activeMetric === 'humidity' ? humPoints : pressPoints
  );

  let activeColor = $derived(
    activeMetric === 'temperature' ? 'var(--accent-orange)' :
    activeMetric === 'humidity' ? 'var(--accent-cyan)' : 'var(--accent-purple)'
  );

  let activeTitle = $derived(
    activeMetric === 'temperature' ? 'Live Temperature Feed' :
    activeMetric === 'humidity' ? 'Live Humidity Feed' : 'Live Pressure Feed'
  );

  let yAxisLabels = $derived(
    activeMetric === 'temperature' ? ['30°C', '25°C', '20°C'] :
    activeMetric === 'humidity' ? ['50%', '45%', '40%'] : ['1015hPa', '1012hPa', '1010hPa']
  );
</script>

<div class="view-container">
  <div class="metrics-grid">
    <MetricCard title="Temperature" value={temperature} unit="°C" Icon={Thermometer} sparkline={tempPoints} active={activeMetric === 'temperature'} onclick={() => activeMetric = 'temperature'} />
    <MetricCard title="Humidity" value={humidity} unit="%" Icon={Droplets} sparkline={humPoints} active={activeMetric === 'humidity'} onclick={() => activeMetric = 'humidity'} />
    <MetricCard title="Pressure" value={pressure} unit="hPa" Icon={Gauge} sparkline={pressPoints} active={activeMetric === 'pressure'} onclick={() => activeMetric = 'pressure'} />
    <MetricCard title="Active Devices" value={activeDevices} unit="Sensors" Icon={Network} onclick={() => window.location.hash = 'fleet'} />
  </div>

  <div class="chart-panel glass-panel" style="--chart-color: {activeColor}">
    <div class="chart-header">
      {#if activeMetric === 'temperature'} <Thermometer size={18} color="var(--chart-color)" />
      {:else if activeMetric === 'humidity'} <Droplets size={18} color="var(--chart-color)" />
      {:else} <Gauge size={18} color="var(--chart-color)" /> {/if}
      <h3>{activeTitle}</h3>
    </div>
    <div class="chart-body">
      <div class="y-axis">
        {#each yAxisLabels as label}
          <span>{label}</span>
        {/each}
      </div>
      <div class="svg-container">
        <!-- Background Grid -->
        <div class="grid-lines">
          <div class="line"></div>
          <div class="line"></div>
          <div class="line"></div>
        </div>
        <svg viewBox="0 0 100 100" preserveAspectRatio="none">
          <defs>
            <linearGradient id="chartGrad" x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stop-color="var(--chart-color)" stop-opacity="0.3" />
              <stop offset="100%" stop-color="var(--chart-color)" stop-opacity="0.0" />
            </linearGradient>
          </defs>
          <polygon points="0,100 {activePoints} 100,100" fill="url(#chartGrad)" />
          <polyline points={activePoints} fill="none" stroke="var(--chart-color)" stroke-width="2" vector-effect="non-scaling-stroke" />
        </svg>
      </div>
    </div>
  </div>
</div>

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

  .chart-panel {
    display: flex;
    flex-direction: column;
    padding: 20px;
    height: 250px; /* Fixed smaller height */
    overflow: hidden; /* Fix layout breakout */
  }

  .chart-header {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 20px;
    flex-shrink: 0;
  }

  h3 {
    color: var(--text-primary);
    font-size: 1.1rem;
    font-weight: 500;
  }

  .chart-body {
    flex-grow: 1;
    display: flex;
    gap: 16px;
    position: relative;
    height: 100%;
    min-height: 0; /* Important for flex children scrolling/overflow */
  }

  .y-axis {
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    color: var(--text-secondary);
    font-size: 0.8rem;
    font-family: var(--font-mono);
  }

  .svg-container {
    flex-grow: 1;
    position: relative;
    height: 100%;
    overflow: hidden; /* Prevent SVG from spilling */
  }

  .grid-lines {
    position: absolute;
    inset: 0;
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    pointer-events: none;
  }

  .line {
    width: 100%;
    height: 1px;
    background: var(--bg-panel-border);
    opacity: 0.5;
  }

  svg {
    width: 100%;
    height: 100%;
    display: block; /* Remove ghost margins under SVGs */
    overflow: visible; /* We hide overflow in .svg-container instead */
  }
</style>
