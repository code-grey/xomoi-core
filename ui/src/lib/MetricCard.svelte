<script lang="ts">
  import type { Component, Snippet } from 'svelte';
  
  let { title, value, unit, Icon, sparkline, children, onclick, active } = $props<{
    title: string;
    value: string | number;
    unit: string;
    Icon: Component<any>;
    sparkline?: string;
    children?: Snippet;
    onclick?: () => void;
    active?: boolean;
  }>();
</script>

<div class="metric-card glass-panel {active ? 'active' : ''} {onclick ? 'clickable' : ''}" {onclick} onkeydown={(e) => e.key === 'Enter' && onclick?.()} role={onclick ? 'button' : 'group'} tabindex={onclick ? 0 : -1}>
  <div class="card-header">
    <span class="title">{title}</span>
    <span class="icon"><Icon size={18} /></span>
  </div>
  <div class="card-body">
    <div class="value-group">
      <span class="value">{value}</span>
      <span class="unit">{unit}</span>
    </div>
    {#if sparkline}
      <div class="sparkline">
        <svg viewBox="0 0 100 100" preserveAspectRatio="none">
          <polyline points={sparkline} fill="none" stroke="var(--accent-cyan)" stroke-width="3" vector-effect="non-scaling-stroke" />
        </svg>
      </div>
    {/if}
  </div>
  {#if children}
    <div class="card-footer">
      {@render children()}
    </div>
  {/if}
</div>

<style>
  .metric-card {
    padding: 20px;
    display: flex;
    flex-direction: column;
    gap: 12px;
    min-width: 200px;
    transition: all 0.2s ease;
  }

  .clickable {
    cursor: pointer;
  }
  .clickable:hover {
    border-color: rgba(255, 255, 255, 0.2);
    transform: translateY(-2px);
  }
  .active {
    border-color: var(--accent-cyan);
    box-shadow: 0 0 16px rgba(0, 255, 204, 0.1);
  }

  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .title {
    color: var(--text-secondary);
    font-size: 0.75rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    font-weight: 600;
  }

  .icon {
    color: var(--accent-cyan);
    opacity: 0.8;
  }

  .card-body {
    display: flex;
    justify-content: space-between;
    align-items: flex-end;
  }

  .value-group {
    display: flex;
    align-items: baseline;
    gap: 6px;
  }

  .value {
    font-size: 2rem;
    font-weight: 700;
    color: var(--text-primary);
    font-family: var(--font-mono);
  }

  .unit {
    color: var(--text-secondary);
    font-size: 0.9rem;
    font-weight: 500;
  }

  .sparkline {
    width: 60px;
    height: 30px;
  }

  svg {
    width: 100%;
    height: 100%;
    overflow: visible;
  }

  .card-footer {
    margin-top: 8px;
    padding-top: 12px;
    border-top: 1px solid var(--bg-panel-border);
  }
</style>
