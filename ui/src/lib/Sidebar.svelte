<!--
Xomoi-Core: Sovereign Edge Node
Copyright (C) 2026 Adrish Bora (@code-grey) & Simanjit Hujuri (@code-zephyrus)

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
-->

<script lang="ts">
  import { LayoutDashboard, Radio, Cpu } from 'lucide-svelte';
  import { globalState } from './store.svelte';

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
    <a href="#overview" class:active={activeTab === 'overview'}>
      <span class="icon"><LayoutDashboard size={20} /></span>
      Overview
    </a>
    <a href="#fleet" class:active={activeTab === 'fleet'}>
      <span class="icon"><Radio size={20} /></span>
      Device Fleet
    </a>
    <a href="#health" class:active={activeTab === 'health'}>
      <span class="icon"><Cpu size={20} /></span>
      Node Health
    </a>
  </nav>

  <div class="system-status-container">
    <div class="system-status">
      <div class="status-indicator live"></div>
      <span class="mono">DARK GRID :1883</span>
    </div>
    <div class="system-status">
      <div class="status-indicator {globalState.webrtcStatus === 'connected' ? 'live' : 'error'}"></div>
      <span class="mono">WEBRTC :WSS</span>
    </div>
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

  .system-status-container {
    padding: 16px 24px;
    border-top: 1px solid var(--bg-panel-border);
    display: flex;
    flex-direction: column;
    gap: 8px;
    margin-top: auto;
  }

  .system-status {
    display: flex;
    align-items: center;
    gap: 12px;
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

  /* Mobile Responsive */
  @media (max-width: 768px) {
    .sidebar {
      width: 100%;
      height: 60px;
      padding: 0;
      position: fixed;
      bottom: 0;
      left: 0;
      z-index: 1000;
      border-top: 1px solid var(--bg-panel-border);
      border-right: none;
      border-radius: 0;
      background: rgba(10, 10, 12, 0.95);
      backdrop-filter: blur(10px);
    }

    .logo-container, .system-status-container {
      display: none;
    }

    .nav-links {
      flex-direction: row;
      justify-content: space-around;
      align-items: center;
      padding: 0;
      height: 100%;
    }

    a {
      padding: 8px;
      flex-direction: column;
      justify-content: center;
      gap: 4px;
      font-size: 0; /* Hide text */
      border: none !important;
      background: transparent !important;
    }

    a.active .icon {
      color: var(--accent-cyan);
      filter: drop-shadow(0 0 8px var(--accent-cyan));
    }

    .icon {
      font-size: 1.5rem;
    }
  }
</style>
