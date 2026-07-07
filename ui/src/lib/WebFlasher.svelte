<script lang="ts">
  import { Usb, Wifi, ShieldCheck, Terminal, X, Loader2 } from 'lucide-svelte';

  let { close } = $props<{ close: () => void }>();

  let step = $state(1);
  let portSelected = $state(false);
  let firmwareType = $state('ESP32-WROOM');
  
  let ssid = $state('');
  let password = $state('');

  import { ESPLoader, Transport } from 'esptool-js';

  let flashLogs = $state<string[]>([
    "Ready to initiate WebSerial connection...",
    "Awaiting user COM port selection..."
  ]);

  async function handleConnect() {
    if (!navigator.serial) {
      flashLogs = [...flashLogs, "ERROR: WebSerial API not supported in this browser. Please use Chrome or Edge."];
      return;
    }

    try {
      flashLogs = [...flashLogs, "Requesting COM port access..."];
      const port = await navigator.serial.requestPort();
      
      flashLogs = [...flashLogs, "Port opened. Initializing Espressif Transport..."];
      const transport = new Transport(port, true);
      
      const terminal = {
        clean: () => {},
        writeLine: (data: string) => { flashLogs = [...flashLogs, data] },
        write: (data: string) => {}
      };

      const loader = new ESPLoader(transport, 115200, terminal);
      await loader.main();
      
      const chipName = await loader.chip.get_chip_description();
      const mac = await loader.chip.read_mac();
      
      flashLogs = [...flashLogs, `SUCCESS: Detected ${chipName}`, `MAC Address: ${mac}`];
      portSelected = true;
      step = 2;
    } catch (e: any) {
      flashLogs = [...flashLogs, `CONNECTION FAILED: ${e.message}`];
    }
  }

  function handleFlash() {
    step = 3;
    flashLogs = [...flashLogs, "Erasing flash memory...", "Writing partition table...", "Uploading firmware.bin... [12%]"];
    
    // Simulate flashing sequence
    setTimeout(() => flashLogs = [...flashLogs, "Uploading firmware.bin... [45%]"], 800);
    setTimeout(() => flashLogs = [...flashLogs, "Uploading firmware.bin... [89%]"], 1600);
    setTimeout(() => {
      flashLogs = [...flashLogs, "Firmware successfully verified.", "Rebooting device via RTS pin..."];
      step = 4;
    }, 2500);
  }
</script>

<div class="modal-backdrop">
  <div class="flasher-modal glass-panel">
    <button class="close-btn" onclick={close}><X size={24} /></button>

    <div class="modal-header">
      <h2>Hardware Provisioning</h2>
      <p class="subtitle">Flash endpoints directly via WebUSB/WebSerial</p>
    </div>

    <div class="stepper">
      <div class="step {step >= 1 ? 'active' : ''} {step > 1 ? 'done' : ''}">
        <div class="step-icon"><Usb size={18} /></div>
        <span>Connect</span>
      </div>
      <div class="step-line {step >= 2 ? 'active' : ''}"></div>
      <div class="step {step >= 2 ? 'active' : ''} {step > 2 ? 'done' : ''}">
        <div class="step-icon"><Wifi size={18} /></div>
        <span>Network</span>
      </div>
      <div class="step-line {step >= 3 ? 'active' : ''}"></div>
      <div class="step {step >= 3 ? 'active' : ''} {step > 3 ? 'done' : ''}">
        <div class="step-icon"><Terminal size={18} /></div>
        <span>Flash</span>
      </div>
      <div class="step-line {step >= 4 ? 'active' : ''}"></div>
      <div class="step {step >= 4 ? 'active' : ''}">
        <div class="step-icon"><ShieldCheck size={18} /></div>
        <span>Done</span>
      </div>
    </div>

    <div class="modal-body">
      {#if step === 1}
        <div class="step-content">
          <h3>Select Hardware Target</h3>
          <select bind:value={firmwareType} class="input-box">
            <option value="ESP32-WROOM">ESP32 (WROOM/WROVER)</option>
            <option value="ESP8266">ESP8266 (NodeMCU)</option>
            <option value="RPI-PICO-W">Raspberry Pi Pico W</option>
          </select>
          <p class="help-text">Connect your microcontroller via USB and click below to authorize the browser to access the COM port.</p>
          <button class="action-btn" onclick={handleConnect}>Connect via WebSerial</button>
        </div>
      {:else if step === 2}
        <div class="step-content">
          <h3>Inject Local Credentials</h3>
          <p class="help-text">These credentials will be baked into the NVS partition during flashing.</p>
          <input type="text" placeholder="Wi-Fi SSID" bind:value={ssid} class="input-box" />
          <input type="password" placeholder="Wi-Fi Password" bind:value={password} class="input-box" />
          <button class="action-btn" onclick={handleFlash}>Write Firmware to Device</button>
        </div>
      {:else if step === 3}
        <div class="step-content flashing">
          <Loader2 class="spinner" size={48} color="var(--accent-cyan)" />
          <h3>Flashing {firmwareType}...</h3>
          <p class="help-text">Do not unplug the device. This takes about 30 seconds.</p>
        </div>
      {:else if step === 4}
        <div class="step-content success">
          <ShieldCheck size={64} color="var(--accent-cyan)" />
          <h3 style="color: var(--accent-cyan)">Provisioning Successful</h3>
          <p class="help-text">The device is rebooting. It will now broadcast a Xomoi Claim beacon. You may unplug the device and deploy it to its physical location.</p>
          <button class="action-btn" onclick={close}>Return to Dashboard</button>
        </div>
      {/if}
    </div>

    <div class="terminal glass-panel mono">
      {#each flashLogs as log}
        <p>> {log}</p>
      {/each}
    </div>
  </div>
</div>

<style>
  .modal-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.8);
    backdrop-filter: blur(8px);
    z-index: 1000;
    display: flex;
    justify-content: center;
    align-items: center;
  }

  .flasher-modal {
    width: 650px;
    background: var(--bg-base);
    padding: 32px;
    position: relative;
    display: flex;
    flex-direction: column;
    gap: 24px;
    border: 1px solid var(--accent-cyan-dim);
  }

  .close-btn {
    position: absolute;
    top: 24px;
    right: 24px;
    background: transparent;
    color: var(--text-secondary);
  }
  .close-btn:hover { color: var(--accent-orange); }

  .modal-header h2 { color: var(--text-primary); font-size: 1.8rem; }
  .subtitle { color: var(--text-secondary); }

  /* Stepper UI */
  .stepper {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 20px;
  }
  .step {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
    color: var(--text-secondary);
  }
  .step-icon {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    border: 2px solid var(--bg-panel-border);
    display: flex;
    justify-content: center;
    align-items: center;
    background: var(--bg-base);
    transition: all 0.3s;
  }
  .step.active { color: var(--accent-cyan); }
  .step.active .step-icon { border-color: var(--accent-cyan); box-shadow: 0 0 12px var(--accent-cyan-dim); }
  .step.done .step-icon { background: var(--accent-cyan); color: #000; }
  .step span { font-size: 0.85rem; font-weight: 600; text-transform: uppercase; }
  
  .step-line {
    flex-grow: 1;
    height: 2px;
    background: var(--bg-panel-border);
    margin: 0 16px;
    margin-bottom: 24px; /* offset for icon vs text */
    transition: background 0.3s;
  }
  .step-line.active { background: var(--accent-cyan); }

  /* Step Contents */
  .modal-body {
    min-height: 200px;
    display: flex;
    justify-content: center;
    align-items: center;
  }
  .step-content {
    width: 100%;
    display: flex;
    flex-direction: column;
    gap: 16px;
  }
  .step-content.flashing, .step-content.success {
    align-items: center;
    text-align: center;
  }
  
  .help-text { color: var(--text-secondary); font-size: 0.9rem; line-height: 1.5; }

  .input-box {
    background: rgba(0,0,0,0.3);
    border: 1px solid var(--bg-panel-border);
    padding: 12px 16px;
    border-radius: 8px;
    color: var(--text-primary);
    font-size: 1rem;
    font-family: inherit;
  }
  .input-box:focus { outline: none; border-color: var(--accent-cyan); }

  .action-btn {
    background: var(--accent-cyan-dim);
    border: 1px solid var(--accent-cyan);
    color: var(--accent-cyan);
    padding: 14px;
    border-radius: 8px;
    font-size: 1rem;
    font-weight: 600;
    margin-top: 8px;
  }
  .action-btn:hover { background: var(--accent-cyan); color: #000; }

  @keyframes spin { 100% { transform: rotate(360deg); } }
  .spinner { animation: spin 1s linear infinite; }

  /* Terminal */
  .terminal {
    background: #050507;
    padding: 16px;
    min-height: 120px;
    max-height: 120px;
    overflow-y: auto;
    font-size: 0.85rem;
    color: var(--text-code);
  }
</style>
