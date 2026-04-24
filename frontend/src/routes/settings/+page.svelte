<script>
  import { onMount } from 'svelte';
  import WorkerSettings from '$lib/components/WorkerSettings.svelte';
  import ScheduleList   from '$lib/components/ScheduleList.svelte';

  let appSettings = $state({ max_concurrent: 1, headless: false });
  let schedules   = $state([]);
  let toast       = $state('');

  async function load() {
    const [st, s] = await Promise.all([
      fetch('/api/settings').then((r) => r.json()),
      fetch('/api/schedules').then((r) => r.json())
    ]);
    appSettings = st;
    schedules   = s.items;
  }

  async function saveSettings() {
    await fetch('/api/settings', {
      method: 'PUT',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify(appSettings)
    });
    showToast('Pengaturan worker disimpan. Akan aktif dalam ~30 detik.');
  }

  async function saveSchedules() {
    await fetch('/api/schedules', {
      method: 'PUT',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify({ schedules })
    });
    showToast('Jadwal otomatis disimpan.');
    await load();
  }

  function showToast(msg) {
    toast = msg;
    setTimeout(() => (toast = ''), 3500);
  }

  onMount(load);
</script>

<svelte:head><title>Pengaturan — Pusaka Worker</title></svelte:head>

<h2 class="page-title">Pengaturan</h2>

{#if toast}
  <div class="toast">{toast}</div>
{/if}

<WorkerSettings bind:settings={appSettings} onsave={saveSettings} />

<ScheduleList bind:schedules onsave={saveSchedules} />

<style>
  .page-title { margin: 0 0 16px; font-size: 1.4rem; color: #e7edf7; }
  .toast {
    background: rgba(31, 170, 112, 0.2);
    border: 1px solid rgba(31, 170, 112, 0.4);
    color: #7ff0b7;
    border-radius: 10px;
    padding: 10px 16px;
    font-size: 0.9rem;
    margin-bottom: 4px;
  }
</style>
