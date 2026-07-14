<script lang="ts">
  import { onMount } from 'svelte';
  import Sidebar from './lib/Sidebar.svelte';
  import TelemetryOverview from './lib/views/TelemetryOverview.svelte';
  import DeviceFleet from './lib/views/DeviceFleet.svelte';
  import NodeHealth from './lib/views/NodeHealth.svelte';
  import ClaimDeviceModal from './lib/ClaimDeviceModal.svelte';
  import { globalState, bootWebRTC, fetchDeviceMetadata } from './lib/store.svelte';

  // Svelte 5 Rune for reactive state synced with URL Hash
  let activeTab = $state(window.location.hash.replace('#', '') || 'overview');
  let showClaimModal = $state(false);
  
  $effect(() => {
    window.location.hash = activeTab;
  });

  onMount(async () => {
    await fetchDeviceMetadata(); // Fetch metadata first
    bootWebRTC(); // Instantly establish P2P connection to Go backend on load

    const handleHashChange = () => {
      const hash = window.location.hash.replace('#', '') || 'overview';
      if (activeTab !== hash) {
        activeTab = hash;
      }
    };
    window.addEventListener('hashchange', handleHashChange);
    
    return () => window.removeEventListener('hashchange', handleHashChange);
  });
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
      
      <div class="header-tools">
        <div class="mobile-network-badges">
          <div class="mobile-badge-item">
            <div class="status-indicator live"></div>
            <span>MQTT</span>
          </div>
          <div class="mobile-badge-item">
            <div class="status-indicator {globalState.webrtcStatus === 'connected' ? 'live' : 'error'}"></div>
            <span>WSS</span>
          </div>
        </div>
        <div class="webrtc-badge glass-panel {globalState.webrtcStatus}">
          WebRTC: {globalState.webrtcStatus.toUpperCase()}
        </div>
        <button class="action-btn glass-panel" onclick={() => showClaimModal = true}>
          + Claim Device
        </button>
      </div>
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

  {#if showClaimModal}
    <ClaimDeviceModal onclose={() => showClaimModal = false} />
  {/if}
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

  .header-tools {
    display: flex;
    gap: 16px;
    align-items: center;
  }

  .webrtc-badge {
    padding: 8px 16px;
    font-size: 0.8rem;
    font-weight: 700;
    letter-spacing: 0.05em;
  }
  .webrtc-badge.connected { color: var(--accent-cyan); border-color: var(--accent-cyan); }
  .webrtc-badge.disconnected { color: var(--accent-orange); border-color: var(--accent-orange); }
  .webrtc-badge.connecting { color: var(--accent-purple); border-color: var(--accent-purple); }

  .tab-content {
    flex-grow: 1;
    display: flex;
    flex-direction: column;
  }

  .mobile-network-badges {
    display: none;
  }

  .status-indicator {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: #FF0000;
  }

  .status-indicator.live {
    background: var(--accent-cyan);
    box-shadow: 0 0 8px var(--accent-cyan);
    animation: pulse 2s infinite;
  }

  /* Mobile Responsive */
  @media (max-width: 768px) {
    .app-layout {
      flex-direction: column;
      padding: 0;
      gap: 0;
    }
    .content-area {
      padding: 16px;
      padding-bottom: 80px; /* Space for the bottom navbar */
    }
    .top-bar {
      flex-direction: column;
      align-items: flex-start;
      gap: 16px;
    }
    h1 {
      font-size: 1.25rem;
    }
    .header-tools {
      width: 100%;
      justify-content: space-between;
    }
    .webrtc-badge {
      display: none; /* Hide large badge on mobile */
    }
    .mobile-network-badges {
      display: flex;
      gap: 16px;
      padding: 6px 12px;
      background: var(--bg-panel);
      border: 1px solid var(--bg-panel-border);
      border-radius: 20px;
    }
    .mobile-badge-item {
      display: flex;
      align-items: center;
      gap: 6px;
    }
    .mobile-badge-item span {
      font-size: 0.7rem;
      font-family: var(--font-mono);
      font-weight: 600;
      color: var(--text-secondary);
    }
  }

  @keyframes pulse {
    0% { opacity: 1; }
    50% { opacity: 0.4; }
    100% { opacity: 1; }
  }
</style>
