<script lang="ts">
  import { X, Clock } from 'lucide-svelte';

  let { sensor, color, onclose } = $props<{
    sensor: any;
    color: string;
    onclose: () => void;
  }>();

  let timeframes = ['1H', '3H', '6H', '12H', '24H', '7D', '30D'];
  let activeTimeframe = $state('24H');

  let mockHistory = $derived.by(() => {
    const t = activeTimeframe; 
    const points = 100;
    return Array.from({ length: points }, (_, i) => {
      const base = sensor.val ? parseFloat(sensor.val) : 25;
      // Smooth telemetry curves using slow sine/cosine combinations
      const wave1 = Math.sin((i / points) * Math.PI * 1.5) * 3;
      const wave2 = Math.cos((i / points) * Math.PI * 3.5) * 1.5;
      const tinyNoise = (Math.random() - 0.5) * 0.4;
      
      const multiplier = activeTimeframe.includes('D') ? 2 : 1;
      return base + (wave1 + wave2 + tinyNoise) * multiplier;
    });
  });

  let minVal = $derived(Math.min(...mockHistory));
  let maxVal = $derived(Math.max(...mockHistory));

  let yLabels = $derived([
    (maxVal).toFixed(1) + sensor.unit,
    ((maxVal + minVal) / 2).toFixed(1) + sensor.unit,
    (minVal).toFixed(1) + sensor.unit
  ]);

  let xLabels = $derived(
    activeTimeframe === '1H' ? ['60m ago', '30m ago', 'Now'] :
    activeTimeframe === '24H' ? ['24h ago', '12h ago', 'Now'] :
    activeTimeframe === '7D' ? ['7d ago', '3d ago', 'Now'] :
    ['30d ago', '15d ago', 'Now']
  );

  function buildPoints(history: number[]) {
    // Add 10% padding to top and bottom to prevent line clipping
    const rawMin = Math.min(...history);
    const rawMax = Math.max(...history);
    const range = Math.max(0.1, rawMax - rawMin);
    const min = rawMin - (range * 0.1);
    const max = rawMax + (range * 0.1);

    return history.map((val, i) => {
      const x = (i / 99) * 100;
      const y = 100 - ((val - min) / (max - min)) * 100;
      return `${x},${y}`;
    }).join(' ');
  }
</script>

<div class="explorer-overlay">
  <div class="explorer-modal glass-panel" style="--chart-color: {color}">
    <div class="header">
      <div class="title">
        <h2>{sensor.name} <span class="unit">({sensor.unit})</span></h2>
        <p class="subtitle">Historical Telemetry Analysis</p>
      </div>
      <button class="close-btn" onclick={onclose}>
        <X size={24} />
      </button>
    </div>

    <div class="timeframe-selector">
      <Clock size={16} color="var(--text-secondary)" />
      {#each timeframes as tf}
        <button 
          class="tf-btn {activeTimeframe === tf ? 'active' : ''}"
          onclick={() => activeTimeframe = tf}
        >
          {tf}
        </button>
      {/each}
    </div>

    <div class="big-chart">
      {#if sensor.unit === 'State' || sensor.unit === 'Bool'}
        <!-- Boolean Timeline Mock (Activity Bar Graph) -->
        <div class="bool-timeline">
           {#each Array(50) as _, i}
             <div class="bool-block {Math.random() > 0.8 ? 'tripped' : 'safe'}"></div>
           {/each}
        </div>
      {:else}
        <div class="chart-layout">
          <div class="y-axis-labels">
            {#each yLabels as label}
              <span>{label}</span>
            {/each}
          </div>
          <div class="chart-main-col">
            <div class="svg-container">
              <div class="grid-lines">
                <div class="line"></div>
                <div class="line"></div>
                <div class="line"></div>
              </div>
              <svg viewBox="0 0 100 100" preserveAspectRatio="none">
                <defs>
                    <linearGradient id="histGrad" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="0%" stop-color="var(--chart-color)" stop-opacity="0.3" />
                      <stop offset="100%" stop-color="var(--chart-color)" stop-opacity="0.0" />
                    </linearGradient>
                  </defs>
                <polygon points="0,100 {buildPoints(mockHistory)} 100,100" fill="url(#histGrad)" />
                <polyline points={buildPoints(mockHistory)} fill="none" stroke="var(--chart-color)" stroke-width="2" vector-effect="non-scaling-stroke" />
              </svg>
            </div>
            <div class="x-axis-labels">
              {#each xLabels as label}
                <span>{label}</span>
              {/each}
            </div>
          </div>
        </div>
      {/if}
    </div>
  </div>
</div>

<style>
  .explorer-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.7);
    backdrop-filter: blur(8px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    padding: 40px;
  }

  .explorer-modal {
    width: 100%;
    max-width: 1000px;
    height: 600px;
    display: flex;
    flex-direction: column;
    padding: 32px;
    gap: 24px;
    animation: slideUp 0.3s cubic-bezier(0.16, 1, 0.3, 1);
  }

  @keyframes slideUp {
    from { opacity: 0; transform: translateY(30px); }
    to { opacity: 1; transform: translateY(0); }
  }

  .header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
  }

  .title h2 {
    color: var(--text-primary);
    font-size: 1.8rem;
    font-weight: 600;
    margin: 0 0 4px 0;
  }

  .unit {
    color: var(--chart-color);
  }

  .subtitle {
    color: var(--text-secondary);
    margin: 0;
    font-size: 0.95rem;
  }

  .close-btn {
    background: none;
    border: none;
    color: var(--text-secondary);
    cursor: pointer;
    padding: 8px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.2s;
  }

  .close-btn:hover {
    background: rgba(255, 255, 255, 0.1);
    color: var(--text-primary);
  }

  .timeframe-selector {
    display: flex;
    align-items: center;
    gap: 12px;
    background: rgba(0, 0, 0, 0.2);
    padding: 8px 16px;
    border-radius: 8px;
    align-self: flex-start;
  }

  .tf-btn {
    background: none;
    border: none;
    color: var(--text-secondary);
    font-family: var(--font-mono);
    font-size: 0.85rem;
    cursor: pointer;
    padding: 4px 8px;
    border-radius: 4px;
    transition: all 0.2s;
  }

  .tf-btn:hover {
    color: var(--text-primary);
  }

  .tf-btn.active {
    background: var(--chart-color);
    color: #fff;
    font-weight: 700;
  }

  .big-chart {
    flex-grow: 1;
    display: flex;
    flex-direction: column;
    min-height: 0; /* Prevent flex children from blowing out container bounds */
    border: 1px solid var(--bg-panel-border);
    border-radius: 8px;
    padding: 16px;
    background: rgba(0, 0, 0, 0.2);
  }

  .chart-layout {
    display: flex;
    gap: 16px;
    flex-grow: 1;
    min-height: 0;
    width: 100%;
  }

  .y-axis-labels {
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    padding-bottom: 24px; /* offset for x-axis space */
    color: var(--text-secondary);
    font-size: 0.85rem;
    font-family: var(--font-mono);
    flex-shrink: 0;
  }

  .chart-main-col {
    display: flex;
    flex-direction: column;
    flex-grow: 1;
    min-width: 0;
  }

  .x-axis-labels {
    display: flex;
    justify-content: space-between;
    color: var(--text-secondary);
    font-size: 0.85rem;
    font-family: var(--font-mono);
    margin-top: 12px;
  }

  .svg-container {
    width: 100%;
    flex-grow: 1;
    position: relative;
    overflow: hidden; /* Fix graph overflow */
    border-radius: 4px;
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
    opacity: 0.3;
  }

  svg {
    width: 100%;
    height: 100%;
    display: block;
    overflow: visible;
  }

  .bool-timeline {
    width: 100%;
    height: 100%;
    display: flex;
    align-items: stretch;
    gap: 2px;
  }

  .bool-block {
    flex-grow: 1;
    border-radius: 2px;
  }
  .bool-block.safe { background: var(--accent-cyan); opacity: 0.4; }
  .bool-block.tripped { background: #ff4757; }
</style>
