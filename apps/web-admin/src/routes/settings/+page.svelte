<script lang="ts">
  import { onMount } from 'svelte';
  import * as Card from '$lib/components/ui/card';
  import { Input } from '$lib/components/ui/input';
  import { Button } from '$lib/components/ui/button';
  import { toast } from '$lib/components/ui/sonner';
  import WorkerSettings from '$lib/components/WorkerSettings.svelte';
  import ScheduleList   from '$lib/components/ScheduleList.svelte';

  let appSettings = $state({ max_concurrent: 5, headless: false });
  let schedules   = $state<any[]>([]);
  let pwForm      = $state({ current: '', next: '', confirm: '' });
  let pwError     = $state('');
  let pwLoading   = $state(false);
  let logoutAllLoading = $state(false);

  async function load() {
    try {
      const [stRes, sRes] = await Promise.all([fetch('/api/pusaka/settings'), fetch('/api/pusaka/schedules')]);
      const st = await stRes.json();
      const s  = await sRes.json();
      if (!st.error) appSettings = st;
      if (!s.error)  schedules   = s.items ?? [];
    } catch { /* silent */ }
  }

  async function saveSettings() {
    try {
      const res  = await fetch('/api/pusaka/settings', {
        method: 'PUT', headers: { 'content-type': 'application/json' },
        body: JSON.stringify(appSettings),
      });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) { showError('Gagal: ' + (data.error ?? res.status)); return; }
      showToast('Pengaturan worker disimpan.');
    } catch { showError('Gagal menyimpan pengaturan'); }
  }

  async function saveSchedules() {
    try {
      const res  = await fetch('/api/pusaka/schedules', {
        method: 'PUT', headers: { 'content-type': 'application/json' },
        body: JSON.stringify({ schedules }),
      });
      if (!res.ok) { showError('Gagal menyimpan jadwal'); return; }
      showToast('Jadwal otomatis disimpan.');
      await load();
    } catch { showError('Gagal menyimpan jadwal'); }
  }

  async function changePassword() {
    pwError = '';
    if (pwForm.next !== pwForm.confirm) { pwError = 'Konfirmasi password tidak cocok'; return; }
    if (pwForm.next.length < 8)         { pwError = 'Password baru minimal 8 karakter'; return; }
    pwLoading = true;
    try {
      const res  = await fetch('/api/auth/change-password', {
        method: 'POST', headers: { 'content-type': 'application/json' },
        body: JSON.stringify({ current_password: pwForm.current, new_password: pwForm.next }),
      });
      const data = await res.json();
      if (!res.ok) { pwError = data.error ?? 'Gagal mengubah password'; return; }
      pwForm = { current: '', next: '', confirm: '' };
      showToast('Password berhasil diubah.');
    } finally { pwLoading = false; }
  }

  async function logoutAllSessions() {
    logoutAllLoading = true;
    try {
      const res = await fetch('/api/auth/logout-all', { method: 'POST' });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) {
        pwError = data.error ?? 'Gagal mengakhiri semua sesi';
        return;
      }
      window.location.href = '/login';
    } finally {
      logoutAllLoading = false;
    }
  }

  function showToast(msg: string) {
    toast.success(msg);
  }

  function showError(msg: string) {
    toast.error(msg);
  }

  onMount(load);
</script>

<svelte:head><title>Pengaturan — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6">

  <div>
    <h1 class="text-2xl font-semibold text-slate-800">Pengaturan</h1>
    <p class="text-sm text-muted-foreground mt-1">Konfigurasi worker, jadwal absensi, dan akun admin</p>
  </div>

  <WorkerSettings bind:settings={appSettings} onsave={saveSettings} />

  <ScheduleList bind:schedules onsave={saveSchedules} />

  <!-- Ubah Password -->
  <Card.Root>
    <Card.Header class="pb-3">
      <Card.Title class="text-base">Ubah Password Admin</Card.Title>
      <Card.Description>Ganti password login akun administrator</Card.Description>
    </Card.Header>
    <Card.Content>
      {#if pwError}
        <p class="mb-3 text-sm text-destructive">{pwError}</p>
      {/if}
      <div class="grid grid-cols-1 gap-4 sm:grid-cols-3">
        <div>
          <label for="pw-current" class="mb-1.5 block text-sm font-medium">Password Saat Ini</label>
          <Input id="pw-current" type="password" bind:value={pwForm.current} autocomplete="current-password" />
        </div>
        <div>
          <label for="pw-next" class="mb-1.5 block text-sm font-medium">Password Baru</label>
          <Input id="pw-next" type="password" bind:value={pwForm.next} autocomplete="new-password" />
        </div>
        <div>
          <label for="pw-confirm" class="mb-1.5 block text-sm font-medium">Konfirmasi Password</label>
          <Input id="pw-confirm" type="password" bind:value={pwForm.confirm} autocomplete="new-password" />
        </div>
      </div>
      <Button class="mt-4" onclick={changePassword} disabled={pwLoading}>
        {pwLoading ? 'Menyimpan…' : 'Simpan Password'}
      </Button>
      <div class="mt-6 rounded-lg border border-amber-200 bg-amber-50 px-4 py-3">
        <p class="text-sm font-medium text-amber-900">Keluar dari semua perangkat</p>
        <p class="mt-1 text-xs text-amber-800">
          Semua sesi login lain akan diakhiri, termasuk token akses yang masih aktif.
        </p>
        <Button class="mt-3" variant="outline" onclick={logoutAllSessions} disabled={logoutAllLoading}>
          {logoutAllLoading ? 'Memproses…' : 'Keluar dari Semua Sesi'}
        </Button>
      </div>
    </Card.Content>
  </Card.Root>

</div>
