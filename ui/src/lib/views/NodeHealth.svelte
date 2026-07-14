<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import MetricCard from '../MetricCard.svelte';
  import { HardDrive, Cpu, Database, Clock, Activity, Zap } from 'lucide-svelte';

  let ramUsage = $state('0.00');
  let uptime = $state('0');
  let numWorkers = $state(0);
  let numCpu = $state(0);
  let walSize = $state('0.00');
  
  // Advanced Metrics
  let gcPauses = $state('0');
  let heapSys = $state('0.00');
  let goroutines = $state('0');
  let showAdvanced = $state(false);

  // Sparkline Chart State
  let ramHistory = $state<number[]>(Array(20).fill(0));
  let maxRam = $state(50);

  let ws: WebSocket;
  let logOutput = $state<string[]>([]);

  onMount(() => {
    ws = new WebSocket('ws://localhost:8085/api/v1/ws/health');
    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      ramUsage = data.ram_usage_mb.toFixed(2);
      uptime = data.uptime_sec;
      numWorkers = data.num_workers;
      numCpu = data.num_cpu || 1;
      walSize = (data.wal_size_mb || 0).toFixed(2);
      
      // Advanced
      gcPauses = (data.gc_pauses_ns / 1000000).toFixed(2); // ms
      heapSys = data.heap_sys_mb.toFixed(2);
      goroutines = data.goroutines;

      const ramVal = data.ram_usage_mb;
      ramHistory = [...ramHistory.slice(1), ramVal];
      if (ramVal > maxRam * 0.8) maxRam = maxRam + 20; 
      
      // Parse and inject real slog JSON
      if (data.new_logs && data.new_logs.length > 0) {
        let newHtmlLogs: string[] = [];
        data.new_logs.forEach((logStr: string) => {
          try {
            const logObj = JSON.parse(logStr);
            const time = new Date(logObj.time).toLocaleTimeString('en-US', { hour12: false });
            const colorClass = logObj.level === 'ERROR' ? 'warn' : 'info';
            
            // Format extra attributes cleanly
            let extraStr = '';
            for (const [k, v] of Object.entries(logObj)) {
              if (!['time', 'level', 'msg'].includes(k)) {
                extraStr += ` <span style="color:#888;">${k}=${v}</span>`;
              }
            }
            
            newHtmlLogs.push(`<span class="time">[${time}]</span> <span class="${colorClass}">${logObj.level}</span> ${logObj.msg}${extraStr}`);
          } catch (e) {
            newHtmlLogs.push(`<span class="time">[Raw]</span> ${logStr}`);
          }
        });
        
        logOutput = newHtmlLogs;
      }
    };
  });

  onDestroy(() => {
    if (ws) ws.close();
  });

  let points = $derived(ramHistory.map((val, i) => {
    const x = (i / 19) * 100;
    const y = 100 - (val / maxRam) * 100;
    return `${x},${y}`;
  }).join(' '));
</script>

<div class="view-container">
  <div class="metrics-grid">
    <MetricCard title="RAM Usage" value={ramUsage} unit="MB" Icon={HardDrive} sparkline={points} />
    
    <!-- Worker Pool with visual blocks -->
    <MetricCard title="Worker Pool" value={`${numWorkers}`} unit={`Workers (vs ${numCpu} CPU Cores)`} Icon={Cpu}>
      <div class="worker-grid" title={`${Math.round((numWorkers/numCpu)*100)}% CPU Thread Saturation`}>
        {#each Array(numWorkers) as _, i}
          <div class="worker-block active"></div>
        {/each}
      </div>
    </MetricCard>
    
    <MetricCard title="SQLite WAL" value={walSize} unit="MB" Icon={Database} />
    <MetricCard title="Uptime" value={uptime} unit="Sec" Icon={Clock} />
  </div>

  <div class="advanced-toggle">
    <button class="toggle-btn" onclick={() => showAdvanced = !showAdvanced}>
      {showAdvanced ? 'Hide Advanced Diagnostics' : 'Show Advanced Diagnostics'}
    </button>
  </div>

  {#if showAdvanced}
  <div class="metrics-grid advanced-grid">
    <MetricCard title="GC Latency" value={gcPauses} unit="ms" Icon={Activity} />
    <MetricCard title="Heap Sys" value={heapSys} unit="MB" Icon={HardDrive} />
    <MetricCard title="Goroutines" value={goroutines} unit="Live" Icon={Zap} />
  </div>
  {/if}

  <div class="terminal-area glass-panel">
    <div class="terminal-header">
      <span class="dot red"></span>
      <span class="dot yellow"></span>
      <span class="dot green"></span>
      <span class="title mono">xomoi-core.log</span>
    </div>
    <div class="terminal-body mono">
      {#if logOutput.length === 0}
        <p class="time">Awaiting log stream...</p>
      {/if}
      {#each logOutput as logStr}
        <p>{@html logStr}</p>
      {/each}
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
    grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
    gap: 24px;
  }
  .advanced-grid {
    margin-top: -8px;
    padding: 16px;
    border: 1px dashed var(--bg-panel-border);
    border-radius: 12px;
    background: rgba(0, 0, 0, 0.2);
  }
  .advanced-toggle {
    display: flex;
    justify-content: flex-end;
  }
  .toggle-btn {
    background: transparent;
    color: var(--text-secondary);
    font-size: 0.8rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    padding: 8px 16px;
    border: 1px solid var(--bg-panel-border);
    border-radius: 6px;
  }
  .toggle-btn:hover {
    color: var(--text-primary);
    border-color: var(--accent-cyan);
  }
  
  /* Worker Visualization */
  .worker-grid {
    display: flex;
    gap: 4px;
    flex-wrap: wrap;
  }
  .worker-block {
    width: 12px;
    height: 12px;
    background: #333;
    border-radius: 2px;
  }
  .worker-block.active {
    background: var(--accent-cyan);
    box-shadow: 0 0 8px rgba(0, 255, 204, 0.4);
  }

  /* Terminal */
  .terminal-area {
    flex-grow: 1;
    display: flex;
    flex-direction: column;
    background: #050507;
    height: 400px; /* Fixed height to prevent infinite page expansion */
  }
  .chart-header, .terminal-header {
    padding: 12px 16px;
    border-bottom: 1px solid var(--bg-panel-border);
    display: flex;
    align-items: center;
    gap: 8px;
  }
  .dot { width: 10px; height: 10px; border-radius: 50%; }
  .dot.red { background: #FF5F56; }
  .dot.yellow { background: #FFBD2E; }
  .dot.green { background: #27C93F; }
  .terminal-header .title {
    margin-left: 12px;
    color: var(--text-secondary);
    font-size: 0.8rem;
  }
  .terminal-body {
    flex-grow: 1;
    padding: 16px;
    font-size: 0.85rem;
    color: var(--text-code);
    line-height: 1.6;
    overflow-y: auto; /* Enable scrolling inside the widget */
  }
  .time { color: #666; }
  :global(.time) { color: #666; }
  :global(.info) { color: var(--accent-cyan); }
  :global(.warn) { color: var(--accent-orange); }
</style>
