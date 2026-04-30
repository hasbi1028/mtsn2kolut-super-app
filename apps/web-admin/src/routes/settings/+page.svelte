<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/state';
  import * as Card from '$lib/components/ui/card';
  import { Input } from '$lib/components/ui/input';
  import { Button } from '$lib/components/ui/button';
  import { toast } from '$lib/components/ui/sonner';
  import WorkerSettings from '$lib/components/WorkerSettings.svelte';
  import ScheduleList   from '$lib/components/ScheduleList.svelte';

  type AuthSession = {
    id: string;
    user_id: string;
    expires_at: string;
    last_used_at: string;
    created_at: string;
    updated_at: string;
    ip_address: string;
    user_agent: string;
    device_label: string;
  };

  let appSettings = $state({ max_concurrent: 5, headless: false });
  let schedules   = $state<any[]>([]);
  let sessions    = $state<AuthSession[]>([]);
  let pwForm      = $state({ current: '', next: '', confirm: '' });
  let pwError     = $state('');
  let pwLoading   = $state(false);
  let logoutAllLoading = $state(false);
  let sessionsLoading = $state(false);
  let revokeSessionLoading = $state<string | null>(null);
  let renameSessionLoading = $state<string | null>(null);
  let labelDrafts = $state<Record<string, string>>({});

  const currentSessionId = $derived(page.data.user?.session_id ?? '');

  async function load() {
    try {
      sessionsLoading = true;
      const [stRes, sRes, sessRes] = await Promise.all([
        fetch('/api/pusaka/settings'),
        fetch('/api/pusaka/schedules'),
        fetch('/api/auth/sessions')
      ]);
      const st = await stRes.json();
      const s  = await sRes.json();
      const sess = await sessRes.json().catch(() => []);
      if (!st.error) appSettings = st;
      if (!s.error)  schedules   = s.items ?? [];
      if (!sess.error) {
        sessions = Array.isArray(sess) ? sess : (sess.data ?? []);
        labelDrafts = Object.fromEntries(
          sessions.map((session) => [session.id, session.device_label || ''])
        );
      }
    } catch { /* silent */ }
    finally { sessionsLoading = false; }
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

  async function revokeSession(sessionId: string) {
    revokeSessionLoading = sessionId;
    try {
      const res = await fetch(`/api/auth/sessions/${sessionId}`, { method: 'DELETE' });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) {
        pwError = data.error ?? 'Gagal mengakhiri sesi';
        return;
      }

      if (sessionId === currentSessionId) {
        await fetch('/api/auth/logout', { method: 'POST' });
        window.location.href = '/login';
        return;
      }

      sessions = sessions.filter((session) => session.id !== sessionId);
      showToast('Sesi berhasil diakhiri.');
    } finally {
      revokeSessionLoading = null;
    }
  }

  async function renameSession(sessionId: string) {
    const deviceLabel = (labelDrafts[sessionId] ?? '').trim();
    if (!deviceLabel) {
      showError('Nama perangkat tidak boleh kosong.');
      return;
    }

    renameSessionLoading = sessionId;
    try {
      const res = await fetch(`/api/auth/sessions/${sessionId}`, {
        method: 'PATCH',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify({ device_label: deviceLabel })
      });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) {
        showError(data.error ?? 'Gagal menyimpan nama perangkat');
        return;
      }

      sessions = sessions.map((session) =>
        session.id === sessionId ? { ...session, device_label: deviceLabel } : session
      );
      showToast('Nama perangkat berhasil disimpan.');
    } finally {
      renameSessionLoading = null;
    }
  }

  function formatDate(value: string) {
    return new Date(value).toLocaleString('id-ID', {
      dateStyle: 'medium',
      timeStyle: 'short',
      timeZone: 'Asia/Makassar'
    });
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

  <Card.Root>
    <Card.Header class="pb-3">
      <Card.Title class="text-base">Sesi Aktif</Card.Title>
      <Card.Description>
        Kelola sesi login yang masih aktif. Revoke sesi akan mencegah refresh token sesi itu dipakai lagi.
      </Card.Description>
    </Card.Header>
    <Card.Content>
      {#if sessionsLoading}
        <p class="text-sm text-muted-foreground">Memuat sesi…</p>
      {:else if sessions.length === 0}
        <p class="text-sm text-muted-foreground">Belum ada sesi aktif tercatat.</p>
      {:else}
        <div class="space-y-3">
          {#each sessions as session (session.id)}
            <div class="rounded-lg border border-slate-200 px-4 py-3">
              <div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                <div class="space-y-1">
                  <div class="flex items-center gap-2">
                    <p class="text-sm font-medium text-slate-800">
                      {session.device_label || `Sesi ${session.id.slice(0, 8)}`}
                    </p>
                    {#if session.id === currentSessionId}
                      <span class="rounded-full bg-green-100 px-2 py-0.5 text-[11px] font-medium text-green-800">
                        Perangkat Ini
                      </span>
                    {/if}
                  </div>
                  {#if session.ip_address}
                    <p class="text-xs text-muted-foreground">
                      IP: {session.ip_address}
                    </p>
                  {/if}
                  <p class="text-xs text-muted-foreground">
                    Terakhir aktif: {formatDate(session.last_used_at)}
                  </p>
                  <p class="text-xs text-muted-foreground">
                    Berlaku sampai: {formatDate(session.expires_at)}
                  </p>
                  {#if session.user_agent}
                    <p class="line-clamp-2 text-[11px] text-slate-400">
                      {session.user_agent}
                    </p>
                  {/if}
                  <div class="pt-2">
                    <label for={`session-label-${session.id}`} class="mb-1 block text-[11px] font-medium uppercase tracking-[0.16em] text-slate-500">
                      Nama perangkat
                    </label>
                    <div class="flex flex-col gap-2 sm:flex-row">
                      <Input
                        id={`session-label-${session.id}`}
                        bind:value={labelDrafts[session.id]}
                        maxlength={60}
                        placeholder="Mis. Laptop Ruang Guru"
                      />
                      <Button
                        variant="secondary"
                        size="sm"
                        onclick={() => renameSession(session.id)}
                        disabled={renameSessionLoading === session.id}
                      >
                        {renameSessionLoading === session.id ? 'Menyimpan…' : 'Simpan Nama'}
                      </Button>
                    </div>
                  </div>
                </div>
                <Button
                  variant="outline"
                  size="sm"
                  onclick={() => revokeSession(session.id)}
                  disabled={revokeSessionLoading === session.id}
                >
                  {revokeSessionLoading === session.id ? 'Memproses…' : 'Akhiri Sesi'}
                </Button>
              </div>
            </div>
          {/each}
        </div>
      {/if}
    </Card.Content>
  </Card.Root>

</div>
