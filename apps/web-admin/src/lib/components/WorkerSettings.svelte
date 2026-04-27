<script lang="ts">
  let { settings = $bindable(), onsave } = $props();
</script>

<section class="card">
  <h2>Worker Settings</h2>
  <div class="settings-grid">
    <label>
      <span>Max Concurrent</span>
      <input type="number" min="1" max="20" bind:value={settings.max_concurrent} />
    </label>
    <label class="check">
      <span>Headless Mode</span>
      <input type="checkbox" bind:checked={settings.headless} />
    </label>
    <div class="settings-actions">
      <button class="btn" onclick={onsave}>Simpan Setting</button>
      <p class="muted small-note">Worker akan sinkron otomatis dari backend dalam sekitar 30 detik.</p>
    </div>
  </div>
</section>

<style>
  .settings-grid { display: grid; grid-template-columns: 1fr 1fr 2fr; gap: 16px; align-items: end; }
  .settings-grid label { display: grid; gap: 6px; }
  .settings-grid span { color: #9db2d1; font-size: 0.9rem; }
  .settings-actions { display: grid; gap: 8px; }
  .small-note { font-size: 0.86rem; margin: 0; }

  @media (max-width: 900px) {
    .settings-grid { grid-template-columns: 1fr; }
  }
</style>
