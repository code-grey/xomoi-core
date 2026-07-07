<script lang="ts">
  // Mock data for the Grafana-style list
  const devices = [
    { id: 'esp32_greenhouse', type: 'DHT22', status: 'Online', lastPing: '2s ago', metric: '24.5 °C' },
    { id: 'pico_perimeter', type: 'PIR Motion', status: 'Online', lastPing: '5s ago', metric: 'SECURE' },
    { id: 'esp8266_attic', type: 'BME280', status: 'Offline', lastPing: '2h ago', metric: '31.2 °C' }
  ];
</script>

<div class="glass-panel table-container">
  <table class="fleet-table">
    <thead>
      <tr>
        <th>Device MAC / ID</th>
        <th>Hardware Profile</th>
        <th>Status</th>
        <th>Last Ping</th>
        <th>Primary Metric</th>
      </tr>
    </thead>
    <tbody>
      {#each devices as dev}
      <tr>
        <td class="mono highlight">{dev.id}</td>
        <td>{dev.type}</td>
        <td>
          <span class="status-badge {dev.status.toLowerCase()}">{dev.status}</span>
        </td>
        <td class="mono text-muted">{dev.lastPing}</td>
        <td class="mono highlight">{dev.metric}</td>
      </tr>
      {/each}
    </tbody>
  </table>
</div>

<style>
  .table-container {
    width: 100%;
    overflow-x: auto;
    padding: 8px;
  }
  .fleet-table {
    width: 100%;
    border-collapse: collapse;
    text-align: left;
  }
  th {
    color: var(--text-secondary);
    text-transform: uppercase;
    font-size: 0.75rem;
    letter-spacing: 0.05em;
    padding: 16px;
    border-bottom: 1px solid var(--bg-panel-border);
  }
  td {
    padding: 16px;
    border-bottom: 1px solid rgba(255,255,255,0.02);
    font-size: 0.9rem;
  }
  tr:hover td {
    background: var(--bg-panel-hover);
  }
  .highlight {
    color: var(--text-primary);
  }
  .text-muted {
    color: var(--text-code);
  }
  .status-badge {
    padding: 4px 8px;
    border-radius: 4px;
    font-size: 0.75rem;
    font-weight: 600;
  }
  .status-badge.online {
    background: var(--accent-cyan-dim);
    color: var(--accent-cyan);
    border: 1px solid rgba(0, 255, 204, 0.2);
  }
  .status-badge.offline {
    background: rgba(255, 85, 0, 0.15);
    color: var(--accent-orange);
    border: 1px solid rgba(255, 85, 0, 0.2);
  }
</style>
