<script lang="ts">
  import { onMount } from 'svelte';
  import WorkerSettings from '$lib/components/WorkerSettings.svelte';
  import ScheduleList   from '$lib/components/ScheduleList.svelte';

  let appSettings = $state({ max_concurrent: 1, headless: false });
  let schedules   = $state([]);
  let toast       = $state('');

  // Change password form
  let pwForm = $state({ current: '', next: '', confirm: '' });
  let pwError = $state('');
  let pwLoading = $state(false);

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
    showToast('Pengaturan worker disimpan.');
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

  async function changePassword() {
    pwError = '';
    if (pwForm.next !== pwForm.confirm) { pwError = 'Konfirmasi password tidak cocok'; return; }
    if (pwForm.next.length < 6)         { pwError = 'Password baru minimal 6 karakter'; return; }
    pwLoading = true;
    try {
      const res = await fetch('/api/auth/change-password', {
        method: 'POST',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify({ current_password: pwForm.current, new_password: pwForm.next })
      });
      const data = await res.json();
      if (!res.ok) { pwError = data.error ?? 'Gagal mengubah password'; return; }
      pwForm = { current: '', next: '', confirm: '' };
      showToast('Password berhasil diubah.');
    } finally {
      pwLoading = false;
    }
  }

  function showToast(msg: string) {
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

<!-- Ubah Password -->
<section class="card">
  <h3>Ubah Password</h3>
  {#if pwError}
    <div class="pw-error">{pwError}</div>
  {/if}
  <div class="pw-grid">
    <label>
      Password Saat Ini
      <input type="password" bind:value={pwForm.current} autocomplete="current-password" />
    </label>
    <label>
      Password Baru
      <input type="password" bind:value={pwForm.next} autocomplete="new-password" />
    </label>
    <label>
      Konfirmasi Password Baru
      <input type="password" bind:value={pwForm.confirm} autocomplete="new-password" />
    </label>
  </div>
  <button class="btn-save" onclick={changePassword} disabled={pwLoading}>
    {pwLoading ? 'Menyimpan…' : 'Simpan Password'}
  </button>
</section>

<style>
  .page-title { margin: 0 0 16px; font-size: 1.4rem; color: #e7edf7; }
  .toast {
    background: rgba(31, 170, 112, 0.2);
    border: 1px solid rgba(31, 170, 112, 0.4);
    color: #7ff0b7; border-radius: 10px;
    padding: 10px 16px; font-size: 0.9rem; margin-bottom: 8px;
  }
  .card {
    background: rgba(22, 30, 46, 0.6);
    border: 1px solid rgba(130, 157, 204, 0.12);
    border-radius: 12px; padding: 20px 22px; margin-top: 16px;
  }
  .card h3 { margin: 0 0 16px; font-size: 1rem; color: #e7edf7; font-weight: 600; }
  .pw-grid { display: flex; flex-direction: column; gap: 12px; margin-bottom: 16px; }
  label { display: flex; flex-direction: column; gap: 5px; font-size: 0.85rem; color: #9db2d1; }
  input {
    background: rgba(7, 11, 19, 0.5);
    border: 1px solid rgba(130, 157, 204, 0.2);
    border-radius: 8px; color: #e7edf7;
    padding: 8px 12px; font-size: 0.9rem; outline: none;
  }
  input:focus { border-color: #58a6ff; }
  .btn-save {
    background: #1f6feb; color: #fff; border: none;
    border-radius: 8px; padding: 8px 18px;
    font-size: 0.9rem; font-weight: 600; cursor: pointer;
  }
  .btn-save:hover:not(:disabled) { background: #388bfd; }
  .btn-save:disabled { opacity: 0.6; cursor: not-allowed; }
  .pw-error {
    background: rgba(248, 81, 73, 0.12);
    border: 1px solid rgba(248, 81, 73, 0.35);
    color: #ff7b72; border-radius: 8px;
    padding: 7px 12px; font-size: 0.85rem; margin-bottom: 12px;
  }
</style>
