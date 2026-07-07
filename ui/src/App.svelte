<script lang="ts">
  import Sidebar from './lib/Sidebar.svelte';
  import TelemetryOverview from './lib/views/TelemetryOverview.svelte';
  import DeviceFleet from './lib/views/DeviceFleet.svelte';
  import NodeHealth from './lib/views/NodeHealth.svelte';

  // Svelte 5 Rune for reactive state
  let activeTab = $state('overview');
</script>

<main class="app-layout">
  <Sidebar bind:activeTab={activeTab} />
  
  <div class="content-area">
    <header class="top-bar">
      <h1>
        {#if activeTab === 'overview'} Dashboard <span class="neon-text">/ Overview</span>
        {:else if activeTab === 'fleet'} Dashboard <span class="neon-text">/ Device Fleet</span>
        {:else if activeTab === 'health'} Dashboard <span class="neon-text">/ Node Health</span>
        {/if}
      </h1>
      <button class="action-btn glass-panel">
        + Claim Device
      </button>
    </header>

    <!-- Tab Routing -->
    <div class="tab-content">
      {#if activeTab === 'overview'}
        <TelemetryOverview />
      {:else if activeTab === 'fleet'}
        <DeviceFleet />
      {:else if activeTab === 'health'}
        <NodeHealth />
      {/if}
    </div>
  </div>
</main>

<style>
  .app-layout {
    display: flex;
    gap: 24px;
    padding: 16px;
    height: 100vh;
  }

  .content-area {
    flex-grow: 1;
    display: flex;
    flex-direction: column;
    gap: 24px;
    overflow-y: auto;
    padding-right: 16px; 
  }

  .top-bar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 8px 0;
  }

  h1 {
    font-size: 1.5rem;
    color: var(--text-primary);
  }

  .action-btn {
    padding: 10px 20px;
    color: var(--accent-cyan);
    font-weight: 600;
    font-size: 0.9rem;
    letter-spacing: 0.05em;
  }

  .action-btn:hover {
    background: var(--accent-cyan-dim);
    box-shadow: 0 0 16px rgba(0, 255, 204, 0.2);
  }

  .tab-content {
    flex-grow: 1;
    display: flex;
    flex-direction: column;
  }
</style>
