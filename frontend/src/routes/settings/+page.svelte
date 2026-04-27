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
    try {
      const [stRes, sRes] = await Promise.all([
        fetch('/api/settings'),
        fetch('/api/schedules')
      ]);
      const st = await stRes.json();
      const s  = await sRes.json();
      if (st.error) console.error('[pusaka] settings:', st.error);
      else          appSettings = st;
      if (s.error)  console.error('[pusaka] schedules:', s.error);
      else          schedules = s.items ?? [];
    } catch (e) {
      console.error('[pusaka] settings load failed:', e);
    }
  }

  async function saveSettings() {
    try {
      const res  = await fetch('/api/settings', {
        method: 'PUT',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify(appSettings)
      });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) { console.error('[pusaka] save settings error:', data.error); showToast('Gagal menyimpan: ' + (data.error ?? res.status)); return; }
      showToast('Pengaturan worker disimpan.');
    } catch (e) {
      console.error('[pusaka] save settings failed:', e);
      showToast('Gagal menyimpan pengaturan');
    }
  }

  async function saveSchedules() {
    try {
      const res  = await fetch('/api/schedules', {
        method: 'PUT',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify({ schedules })
      });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) { console.error('[pusaka] save schedules error:', data.error); showToast('Gagal menyimpan jadwal'); return; }
      showToast('Jadwal otomatis disimpan.');
      await load();
    } catch (e) {
      console.error('[pusaka] save schedules failed:', e);
      showToast('Gagal menyimpan jadwal');
    }
  }

  // Backup & Restore
  let restoreFile = $state<File | null>(null);
  let restoreLoading = $state(false);
  let restoreError = $state('');

  function downloadBackup() {
    window.location.href = '/api/backup';
  }

  async function doRestore() {
    if (!restoreFile) return;
    restoreError = '';
    restoreLoading = true;
    try {
      const form = new FormData();
      form.append('file', restoreFile);
      const res = await fetch('/api/restore', { method: 'POST', body: form });
      const data = await res.json();
      if (!res.ok) { restoreError = data.error ?? 'Restore gagal'; return; }
      showToast(data.message ?? 'Restore berhasil. App sedang restart...');
      restoreFile = null;
    } finally {
      restoreLoading = false;
    }
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

<!-- Backup & Restore -->
<section class="card">
  <h3>Backup & Restore Database</h3>
  <div class="backup-row">
    <div class="backup-col">
      <p class="hint">Unduh salinan database SQLite saat ini.</p>
      <button class="btn-save btn-backup" onclick={downloadBackup}>Unduh Backup</button>
    </div>
    <div class="backup-divider"></div>
    <div class="backup-col">
      <p class="hint">Pulihkan database dari file backup. <strong>App akan restart otomatis.</strong></p>
      {#if restoreError}
        <div class="pw-error">{restoreError}</div>
      {/if}
      <div class="restore-row">
        <label class="file-label">
          <span>{restoreFile ? restoreFile.name : 'Pilih file .sqlite…'}</span>
          <input type="file" accept=".sqlite,.db" onchange={(e) => restoreFile = (e.target as HTMLInputElement).files?.[0] ?? null} />
        </label>
        <button class="btn-save btn-restore" onclick={doRestore} disabled={!restoreFile || restoreLoading}>
          {restoreLoading ? 'Memulihkan…' : 'Restore'}
        </button>
      </div>
    </div>
  </div>
</section>

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
  .backup-row { display: flex; gap: 20px; align-items: flex-start; flex-wrap: wrap; }
  .backup-col { flex: 1; min-width: 200px; display: flex; flex-direction: column; gap: 10px; }
  .backup-divider { width: 1px; background: rgba(130,157,204,0.12); align-self: stretch; }
  .hint { margin: 0; font-size: 0.82rem; color: #9db2d1; line-height: 1.4; }
  .hint strong { color: #ffb74d; }
  .btn-backup { background: #2d6a4f; }
  .btn-backup:hover { background: #40916c; }
  .btn-restore { background: #7b2d2d; white-space: nowrap; }
  .btn-restore:hover:not(:disabled) { background: #a03030; }
  .restore-row { display: flex; gap: 8px; align-items: center; flex-wrap: wrap; }
  .file-label {
    flex: 1; display: flex; align-items: center;
    background: rgba(7,11,19,0.5);
    border: 1px solid rgba(130,157,204,0.2);
    border-radius: 8px; padding: 7px 12px;
    font-size: 0.85rem; color: #9db2d1; cursor: pointer;
    min-width: 0; overflow: hidden;
  }
  .file-label span { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .file-label input { display: none; }
  .pw-error {
    background: rgba(248, 81, 73, 0.12);
    border: 1px solid rgba(248, 81, 73, 0.35);
    color: #ff7b72; border-radius: 8px;
    padding: 7px 12px; font-size: 0.85rem; margin-bottom: 12px;
  }
</style>
