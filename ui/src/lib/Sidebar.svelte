<script lang="ts">
  import { LayoutDashboard, Radio, Cpu } from 'lucide-svelte';

  let { activeTab = $bindable() } = $props<{
    activeTab: string;
  }>();
</script>

<aside class="sidebar glass-panel">
  <div class="logo-container">
    <div class="glow-orb"></div>
    <h2>XOMOI</h2>
  </div>

  <nav class="nav-links">
    <a href="#" class:active={activeTab === 'overview'} onclick={() => activeTab = 'overview'}>
      <span class="icon"><LayoutDashboard size={20} /></span>
      Overview
    </a>
    <a href="#" class:active={activeTab === 'fleet'} onclick={() => activeTab = 'fleet'}>
      <span class="icon"><Radio size={20} /></span>
      Device Fleet
    </a>
    <a href="#" class:active={activeTab === 'health'} onclick={() => activeTab = 'health'}>
      <span class="icon"><Cpu size={20} /></span>
      Node Health
    </a>
  </nav>

  <div class="system-status">
    <div class="status-indicator live"></div>
    <span class="mono">BROKER :1883</span>
  </div>
</aside>

<style>
  .sidebar {
    width: 260px;
    height: calc(100vh - 32px); /* Margin offset */
    display: flex;
    flex-direction: column;
    padding: 24px 0;
  }

  .logo-container {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 0 24px;
    margin-bottom: 48px;
  }

  .glow-orb {
    width: 12px;
    height: 12px;
    border-radius: 50%;
    background: var(--accent-cyan);
    box-shadow: 0 0 12px var(--accent-cyan);
  }

  h2 {
    font-size: 1.2rem;
    font-weight: 700;
    letter-spacing: 0.1em;
  }

  .nav-links {
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding: 0 16px;
    flex-grow: 1;
  }

  a {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px 16px;
    color: var(--text-secondary);
    text-decoration: none;
    border-radius: 8px;
    transition: var(--transition-smooth);
    font-weight: 500;
  }

  a:hover {
    background: var(--bg-panel-hover);
    color: var(--text-primary);
  }

  a.active {
    background: var(--accent-cyan-dim);
    color: var(--accent-cyan);
    border: 1px solid rgba(0, 255, 204, 0.2);
  }

  .icon {
    font-size: 1.2rem;
  }

  .system-status {
    padding: 16px 24px;
    border-top: 1px solid var(--bg-panel-border);
    display: flex;
    align-items: center;
    gap: 12px;
    margin-top: auto;
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

  .system-status span {
    font-size: 0.75rem;
    color: var(--text-code);
  }

  @keyframes pulse {
    0% { opacity: 1; }
    50% { opacity: 0.4; }
    100% { opacity: 1; }
  }
</style>
