<script lang="ts">
  import MetricCard from '../MetricCard.svelte';
  import { Server, ShieldAlert, Activity } from 'lucide-svelte';
</script>

<div class="view-container">
  <div class="metrics-grid">
    <MetricCard title="Total Connected" value="12" unit="Devices" Icon={Server} />
    <MetricCard title="Active Alerts" value="0" unit="Rules" Icon={ShieldAlert} />
    <MetricCard title="Ingestion Rate" value="145" unit="Msg/sec" Icon={Activity} />
  </div>

  <div class="chart-area glass-panel">
    <div class="chart-placeholder">
      <p class="mono">Live Fleet Telemetry Stream [ Awaiting Target... ]</p>
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
    grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
    gap: 24px;
  }
  .chart-area {
    flex-grow: 1;
    min-height: 400px;
    display: flex;
    align-items: center;
    justify-content: center;
    border: 1px dashed var(--bg-panel-border);
  }
  .chart-placeholder {
    color: var(--text-code);
    opacity: 0.5;
  }
</style>
