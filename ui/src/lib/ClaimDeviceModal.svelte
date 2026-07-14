<script lang="ts">
  import { Shield, Key, Check, X, AlertTriangle } from 'lucide-svelte';

  let { onclose } = $props<{ onclose: () => void }>();

  let macAddress = $state('');
  let deviceName = $state('');
  
  let isClaiming = $state(false);
  let errorMsg = $state('');
  let successKey = $state('');

  async function handleClaim() {
    if (!macAddress || !deviceName) {
      errorMsg = "Both MAC Address and Device Name are required.";
      return;
    }
    
    errorMsg = '';
    isClaiming = true;
    
    try {
      const res = await fetch('http://localhost:8085/api/v1/devices/claim', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ mac_address: macAddress, device_name: deviceName })
      });
      
      const data = await res.json();
      
      if (!res.ok) {
        errorMsg = data.message || data.error || await res.text() || "Failed to claim device.";
      } else {
        successKey = data.private_key;
      }
    } catch (err) {
      errorMsg = "Network error. Is the backend running?";
    } finally {
      isClaiming = false;
    }
  }

  function copyKey() {
    navigator.clipboard.writeText(successKey);
  }
</script>

<div class="modal-backdrop" onclick={onclose}>
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div class="modal-content glass-panel" onclick={(e) => e.stopPropagation()}>
    <button class="close-btn" onclick={onclose}><X size={20} /></button>
    
    {#if !successKey}
      <div class="modal-header">
        <div class="icon-orb"><Shield size={24} /></div>
        <h2>Claim New Node</h2>
        <p>Register a factory device to your Dark Grid and generate its secure Private Key.</p>
      </div>

      <div class="modal-body">
        {#if errorMsg}
          <div class="error-banner">
            <AlertTriangle size={16} />
            {errorMsg}
          </div>
        {/if}

        <div class="input-group">
          <label for="mac">Device MAC Address</label>
          <input id="mac" type="text" bind:value={macAddress} placeholder="e.g. 00:1A:2B:3C:4D:5E" />
          <span class="hint">The device must be powered on and connected to the Dark Grid once.</span>
        </div>

        <div class="input-group">
          <label for="name">Friendly Name</label>
          <input id="name" type="text" bind:value={deviceName} placeholder="e.g. Living Room Node" />
        </div>
      </div>

      <div class="modal-footer">
        <button class="btn secondary" onclick={onclose}>Cancel</button>
        <button class="btn primary" onclick={handleClaim} disabled={isClaiming}>
          {isClaiming ? 'Claiming...' : 'Claim Device'}
        </button>
      </div>
    {:else}
      <div class="modal-header success">
        <div class="icon-orb success-orb"><Check size={24} /></div>
        <h2>Device Claimed Successfully!</h2>
        <p>The factory secret has been wiped. This node is now cryptographically bound to you.</p>
      </div>

      <div class="modal-body">
        <div class="key-display">
          <label>HMAC-SHA256 Private Key</label>
          <div class="key-box">
            <Key size={16} class="key-icon" />
            <code class="secret-key">{successKey}</code>
          </div>
          <p class="warning-text">
            <AlertTriangle size={14} />
            Copy this key immediately. You will never see it again. Update your simulator or firmware code with this new key!
          </p>
        </div>
      </div>

      <div class="modal-footer center">
        <button class="btn secondary" onclick={copyKey}>Copy Key</button>
        <button class="btn primary" onclick={onclose}>Done</button>
      </div>
    {/if}
  </div>
</div>

<style>
  /* Base Modal Styles */
  .modal-backdrop {
    position: fixed;
    top: 0; left: 0; right: 0; bottom: 0;
    background: rgba(0, 0, 0, 0.7);
    backdrop-filter: blur(4px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 9999;
    padding: 16px;
  }

  .modal-content {
    width: 100%;
    max-width: 480px;
    background: var(--bg-panel);
    border: 1px solid var(--bg-panel-border);
    border-radius: 12px;
    box-shadow: 0 12px 48px rgba(0,0,0,0.5);
    position: relative;
    display: flex;
    flex-direction: column;
    animation: slideUp 0.3s cubic-bezier(0.16, 1, 0.3, 1);
  }

  @keyframes slideUp {
    from { opacity: 0; transform: translateY(20px); }
    to { opacity: 1; transform: translateY(0); }
  }

  .close-btn {
    position: absolute;
    top: 16px; right: 16px;
    background: transparent;
    border: none;
    color: var(--text-secondary);
    cursor: pointer;
    padding: 4px;
    border-radius: 4px;
    transition: var(--transition-smooth);
  }
  .close-btn:hover {
    background: rgba(255,255,255,0.1);
    color: var(--text-primary);
  }

  .modal-header {
    padding: 32px 32px 16px;
    text-align: center;
    display: flex;
    flex-direction: column;
    align-items: center;
  }

  .icon-orb {
    width: 48px; height: 48px;
    border-radius: 50%;
    background: rgba(0, 255, 204, 0.1);
    color: var(--accent-cyan);
    display: flex; align-items: center; justify-content: center;
    margin-bottom: 16px;
    box-shadow: 0 0 24px rgba(0, 255, 204, 0.2);
  }
  .icon-orb.success-orb {
    background: rgba(46, 213, 115, 0.1);
    color: #2ed573;
    box-shadow: 0 0 24px rgba(46, 213, 115, 0.2);
  }

  .modal-header h2 {
    margin: 0 0 8px;
    font-size: 1.4rem;
    color: var(--text-primary);
  }
  .modal-header p {
    margin: 0;
    color: var(--text-secondary);
    font-size: 0.95rem;
    line-height: 1.5;
  }

  .modal-body {
    padding: 16px 32px;
    display: flex;
    flex-direction: column;
    gap: 20px;
  }

  .error-banner {
    display: flex;
    align-items: center;
    gap: 8px;
    background: rgba(255, 71, 87, 0.1);
    border: 1px solid rgba(255, 71, 87, 0.3);
    color: #ff4757;
    padding: 12px;
    border-radius: 8px;
    font-size: 0.9rem;
  }

  .input-group {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .input-group label {
    font-size: 0.85rem;
    color: var(--text-secondary);
    font-weight: 500;
  }
  .input-group input {
    background: rgba(0,0,0,0.2);
    border: 1px solid var(--bg-panel-border);
    padding: 12px 16px;
    border-radius: 8px;
    color: var(--text-primary);
    font-family: inherit;
    font-size: 1rem;
    transition: var(--transition-smooth);
  }
  .input-group input:focus {
    outline: none;
    border-color: var(--accent-cyan);
    box-shadow: 0 0 0 2px rgba(0, 255, 204, 0.1);
  }
  .input-group .hint {
    font-size: 0.75rem;
    color: rgba(255,255,255,0.4);
  }

  .modal-footer {
    padding: 24px 32px 32px;
    display: flex;
    justify-content: flex-end;
    gap: 12px;
  }
  .modal-footer.center {
    justify-content: center;
  }

  .btn {
    padding: 10px 24px;
    border-radius: 8px;
    font-family: inherit;
    font-weight: 600;
    font-size: 0.95rem;
    cursor: pointer;
    border: none;
    transition: var(--transition-smooth);
  }
  .btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
  .btn.secondary {
    background: transparent;
    color: var(--text-secondary);
  }
  .btn.secondary:hover {
    background: rgba(255,255,255,0.05);
    color: var(--text-primary);
  }
  .btn.primary {
    background: var(--accent-cyan);
    color: #000;
  }
  .btn.primary:hover:not(:disabled) {
    background: #00e6b8;
    box-shadow: 0 0 16px rgba(0, 255, 204, 0.4);
  }

  /* Success State Styles */
  .key-display {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  .key-display label {
    font-size: 0.85rem;
    color: var(--text-secondary);
  }
  .key-box {
    display: flex;
    align-items: center;
    gap: 12px;
    background: rgba(0,0,0,0.3);
    border: 1px solid rgba(0, 255, 204, 0.3);
    padding: 16px;
    border-radius: 8px;
  }
  .key-icon {
    color: var(--accent-cyan);
  }
  .secret-key {
    color: var(--accent-cyan);
    font-size: 1.1rem;
    letter-spacing: 0.05em;
    word-break: break-all;
  }
  .warning-text {
    display: flex;
    align-items: flex-start;
    gap: 8px;
    color: var(--accent-orange);
    font-size: 0.85rem;
    background: rgba(255, 171, 0, 0.1);
    padding: 12px;
    border-radius: 8px;
    margin-top: 8px;
  }
</style>
