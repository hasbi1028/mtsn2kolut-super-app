<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/state';
  import * as Card from '$lib/components/ui/card';
  import { Input } from '$lib/components/ui/input';
  import { Button } from '$lib/components/ui/button';
  import { toast } from '$lib/components/ui/sonner';
  import { Skeleton } from '$lib/components/ui/skeleton';
  import AsyncContent from '$lib/components/AsyncContent.svelte';
  import WorkerSettings from '$lib/components/WorkerSettings.svelte';
  import ScheduleList   from '$lib/components/ScheduleList.svelte';
  import LoadingButton from '$lib/components/LoadingButton.svelte';
  import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
  import { clearCbtComposerDrafts } from '$lib/client/cbt-drafts';
  import { readClientApiData, readClientJson } from '$lib/client/api';

  type WorkerSettingState = {
    max_concurrent: number;
    headless: boolean;
    pusaka_geo_base_lat: number;
    pusaka_geo_base_lng: number;
    pusaka_geo_default_radius_m: number;
    pusaka_geo_checkin_radius_m: number;
    pusaka_geo_checkout_radius_m: number;
  };

  type Schedule = {
    id: string;
    label: string;
    run_time: string;
    run_type: string;
    is_enabled: boolean;
  };

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

  type SettingsOverview = {
    appSettings: WorkerSettingState;
    schedules: Schedule[];
    sessions: AuthSession[];
    pendingProfileChanges: number;
  };

  type PendingProfileChangesPayload = {
    pending?: number;
    count?: number;
  };

  type SchedulePayload = {
    items?: Schedule[];
    data?: Schedule[];
    error?: string;
    message?: string;
  };

  let settingsPromise = $state<Promise<SettingsOverview> | null>(null);
  let appSettings = $state<WorkerSettingState>(emptyWorkerSettings());
  let schedules   = $state<Schedule[]>([]);
  let sessions    = $state<AuthSession[]>([]);
  let pendingProfileChanges = $state(0);
  let pwForm      = $state({ current: '', next: '', confirm: '' });
  let pwLoading   = $state(false);
  let logoutAllLoading = $state(false);
  let revokeSessionLoading = $state<string | null>(null);
  let renameSessionLoading = $state<string | null>(null);
  let labelDrafts = $state<Record<string, string>>({});

  const currentSessionId = $derived(page.data.user?.session_id ?? '');
  const isAdmin = $derived(
    Boolean(page.data.user?.roles?.includes('admin') || page.data.user?.role === 'admin')
  );
  const canReviewProfileChanges = $derived(
    Boolean(isAdmin || page.data.user?.permissions?.includes('profile_changes.review'))
  );

  function emptyWorkerSettings(): WorkerSettingState {
    return {
      max_concurrent: 5,
      headless: false,
      pusaka_geo_base_lat: -3.2163111,
      pusaka_geo_base_lng: 121.0428659,
      pusaka_geo_default_radius_m: 50,
      pusaka_geo_checkin_radius_m: 55,
      pusaka_geo_checkout_radius_m: 28
    };
  }

  function normalizeSchedules(value: SchedulePayload | Schedule[] | null | undefined): Schedule[] {
    if (Array.isArray(value)) return value;
    if (!value) return [];
    if (Array.isArray(value.items)) return value.items;
    if (Array.isArray(value.data)) return value.data;
    return [];
  }

  function normalizeSessions(value: AuthSession[] | null | undefined): AuthSession[] {
    return Array.isArray(value) ? value : [];
  }

  function applySettingsOverview(overview: SettingsOverview) {
    appSettings = overview.appSettings;
    schedules = overview.schedules;
    sessions = overview.sessions;
    pendingProfileChanges = overview.pendingProfileChanges;
    labelDrafts = Object.fromEntries(
      overview.sessions.map((session) => [session.id, session.device_label || ''])
    );
  }

  function currentSettingsOverview(): SettingsOverview {
    return { appSettings, schedules, sessions, pendingProfileChanges };
  }

  async function fetchPendingProfileChanges() {
    if (!canReviewProfileChanges) return 0;
    const data = await fetch('/api/users/change-requests/pending-count').then((response) =>
      readClientApiData<PendingProfileChangesPayload>(response, 'Gagal memuat jumlah permintaan perubahan data')
    );
    const pending = Number(data.pending ?? data.count ?? 0);
    return Number.isFinite(pending) ? pending : 0;
  }

  async function fetchSettingsOverview(): Promise<SettingsOverview> {
    if (!isAdmin) {
      const [sessionData, pendingCount] = await Promise.all([
        fetch('/api/auth/sessions').then((response) =>
          readClientApiData<AuthSession[]>(response, 'Gagal memuat sesi aktif')
        ),
        fetchPendingProfileChanges()
      ]);
      return {
        appSettings: emptyWorkerSettings(),
        schedules: [],
        sessions: normalizeSessions(sessionData),
        pendingProfileChanges: pendingCount
      };
    }

    const [sessionData, settingsData, scheduleData, pendingCount] = await Promise.all([
      fetch('/api/auth/sessions').then((response) =>
        readClientApiData<AuthSession[]>(response, 'Gagal memuat sesi aktif')
      ),
      fetch('/api/pusaka/settings').then((response) =>
        readClientApiData<Partial<WorkerSettingState>>(response, 'Gagal memuat pengaturan worker')
      ),
      fetch('/api/pusaka/schedules').then((response) =>
        readClientApiData<SchedulePayload | Schedule[]>(response, 'Gagal memuat jadwal PUSAKA')
      ),
      fetchPendingProfileChanges()
    ]);

    return {
      appSettings: { ...emptyWorkerSettings(), ...(settingsData ?? {}) },
      schedules: normalizeSchedules(scheduleData),
      sessions: normalizeSessions(sessionData),
      pendingProfileChanges: pendingCount
    };
  }

  function loadSettings() {
    settingsPromise = fetchSettingsOverview().then((overview) => {
      applySettingsOverview(overview);
      return overview;
    });
    return settingsPromise;
  }

  async function refreshSettings(showFailureToast = false) {
    if (!settingsPromise) {
      await loadSettings();
      return;
    }
    try {
      const overview = await fetchSettingsOverview();
      applySettingsOverview(overview);
      settingsPromise = Promise.resolve(overview);
    } catch (error) {
      settingsPromise = Promise.resolve(currentSettingsOverview());
      if (showFailureToast) showError(settingsErrorMessage(error));
    }
  }

  function retrySettings(reset?: () => void) {
    reset?.();
    loadSettings();
  }

  function settingsErrorMessage(error: unknown) {
    if (error instanceof Error && error.message.trim()) return error.message;
    if (typeof error === 'string' && error.trim()) return error;
    return 'Data pengaturan belum dapat dimuat. Periksa koneksi backend lalu coba lagi.';
  }

  function mutationErrorMessage(error: unknown, fallbackMessage: string) {
    if (error instanceof Error && error.message.trim()) return error.message;
    return fallbackMessage;
  }

  function handleSettingsRenderError(error: unknown, reset: () => void) {
    console.error('Settings overview render failed', error);
    reset();
  }

  async function saveSettings() {
    try {
      const res  = await fetch('/api/pusaka/settings', {
        method: 'PUT', headers: { 'content-type': 'application/json' },
        body: JSON.stringify(appSettings),
      });
      await readClientJson<unknown>(res);
      showToast('Pengaturan worker disimpan.');
    } catch (error) { showError(mutationErrorMessage(error, 'Gagal menyimpan pengaturan')); }
  }

  async function saveSchedules() {
    try {
      const res  = await fetch('/api/pusaka/schedules', {
        method: 'PUT', headers: { 'content-type': 'application/json' },
        body: JSON.stringify({ schedules }),
      });
      await readClientJson<unknown>(res);
      showToast('Jadwal otomatis disimpan.');
      await refreshSettings(true);
    } catch (error) { showError(mutationErrorMessage(error, 'Gagal menyimpan jadwal')); }
  }

  async function changePassword() {
    if (pwForm.next !== pwForm.confirm) { showError('Konfirmasi password tidak cocok'); return; }
    if (pwForm.next.length < 8)         { showError('Password baru minimal 8 karakter'); return; }
    pwLoading = true;
    try {
      const res  = await fetch('/api/auth/change-password', {
        method: 'POST', headers: { 'content-type': 'application/json' },
        body: JSON.stringify({ current_password: pwForm.current, new_password: pwForm.next }),
      });
      await readClientJson<unknown>(res);
      pwForm = { current: '', next: '', confirm: '' };
      showToast('Password berhasil diubah.');
    } catch (error) {
      showError(mutationErrorMessage(error, 'Gagal mengubah password'));
    } finally { pwLoading = false; }
  }

  async function logoutAllSessions() {
    logoutAllLoading = true;
    try {
      const res = await fetch('/api/auth/logout-all', { method: 'POST' });
      await readClientJson<unknown>(res);
      clearCbtComposerDrafts();
      window.location.href = '/login';
    } catch (error) {
      showError(mutationErrorMessage(error, 'Gagal mengakhiri semua sesi'));
    } finally {
      logoutAllLoading = false;
    }
  }

  async function revokeSession(sessionId: string) {
    revokeSessionLoading = sessionId;
    try {
      const res = await fetch(`/api/auth/sessions/${sessionId}`, { method: 'DELETE' });
      await readClientJson<unknown>(res);

      if (sessionId === currentSessionId) {
        await fetch('/api/auth/logout', { method: 'POST' }).catch(() => undefined);
        clearCbtComposerDrafts();
        window.location.href = '/login';
        return;
      }

      sessions = sessions.filter((session) => session.id !== sessionId);
      settingsPromise = Promise.resolve(currentSettingsOverview());
      showToast('Sesi berhasil diakhiri.');
    } catch (error) {
      showError(mutationErrorMessage(error, 'Gagal mengakhiri sesi'));
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
      await readClientJson<unknown>(res);

      sessions = sessions.map((session) =>
        session.id === sessionId ? { ...session, device_label: deviceLabel } : session
      );
      settingsPromise = Promise.resolve(currentSettingsOverview());
      showToast('Nama perangkat berhasil disimpan.');
    } catch (error) {
      showError(mutationErrorMessage(error, 'Gagal menyimpan nama perangkat'));
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

  onMount(() => {
    void loadSettings();
  });
</script>

<svelte:head><title>Pengaturan — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6">

  <div>
    <h1 class="text-2xl font-semibold text-slate-800">Pengaturan</h1>
    <p class="text-sm text-muted-foreground mt-1">
      {#if isAdmin}
        Konfigurasi worker, jadwal absensi, dan keamanan akun.
      {:else}
        Keamanan akun dan pengelolaan sesi perangkat.
      {/if}
    </p>
  </div>

  <AsyncContent promise={settingsPromise} onerror={handleSettingsRenderError}>
    {#snippet pending()}
      <div class="space-y-4">
        <Card.Root>
          <Card.Header class="pb-3">
            <Skeleton class="h-5 w-40" />
            <Skeleton class="h-4 w-72" />
          </Card.Header>
          <Card.Content>
            <div class="grid grid-cols-1 gap-4 sm:grid-cols-3">
              <Skeleton class="h-10 w-full" />
              <Skeleton class="h-10 w-full" />
              <Skeleton class="h-10 w-36" />
            </div>
          </Card.Content>
        </Card.Root>
        <Card.Root>
          <Card.Header class="pb-3">
            <Skeleton class="h-5 w-44" />
            <Skeleton class="h-4 w-80" />
          </Card.Header>
          <Card.Content class="space-y-3">
            {#each Array.from({ length: 3 }) as _, index (`settings-session-skeleton-${index}`)}
              <div class="rounded-lg border border-slate-200 px-4 py-3">
                <div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                  <div class="space-y-2">
                    <Skeleton class="h-5 w-36" />
                    <Skeleton class="h-4 w-28" />
                    <Skeleton class="h-4 w-40" />
                  </div>
                  <Skeleton class="h-9 w-24" />
                </div>
              </div>
            {/each}
          </Card.Content>
        </Card.Root>
      </div>
    {/snippet}
    {#snippet failed(error, reset)}
      <RecoveryPanel title="Pengaturan Belum Tersaji" message={settingsErrorMessage(error)} onRetry={() => retrySettings(reset)} />
    {/snippet}
    {#snippet children(_value)}

  {#if canReviewProfileChanges}
    <Card.Root>
      <Card.Header class="flex flex-col gap-3 pb-3 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <Card.Title class="text-base">Permintaan Perubahan Data</Card.Title>
          <Card.Description>
            {pendingProfileChanges} permintaan resmi menunggu review admin.
          </Card.Description>
        </div>
        <Button href="/settings/user-change-requests" variant={pendingProfileChanges > 0 ? 'default' : 'outline'}>
          Buka Review
        </Button>
      </Card.Header>
    </Card.Root>
  {/if}

  {#if isAdmin}
    <WorkerSettings bind:settings={appSettings} onsave={saveSettings} />

    <ScheduleList bind:schedules onsave={saveSchedules} />
  {/if}

  <!-- Ubah Password -->
  <Card.Root>
      <Card.Header class="pb-3">
      <Card.Title class="text-base">Ubah Password</Card.Title>
      <Card.Description>
        {#if isAdmin}
          Ganti password login akun administrator.
        {:else}
          Ganti password akun Anda.
        {/if}
      </Card.Description>
      </Card.Header>
    <Card.Content>
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
      <LoadingButton class="mt-4" onclick={() => void changePassword()} loading={pwLoading} loadingLabel="Menyimpan..." label="Simpan Password" />
      <div class="mt-6 rounded-lg border border-amber-200 bg-amber-50 px-4 py-3">
        <p class="text-sm font-medium text-amber-900">Keluar dari semua perangkat</p>
        <p class="mt-1 text-xs text-amber-800">
          Semua sesi login lain akan diakhiri, termasuk token akses yang masih aktif.
        </p>
        <LoadingButton class="mt-3" variant="outline" onclick={() => void logoutAllSessions()} loading={logoutAllLoading} loadingLabel="Memproses..." label="Keluar dari Semua Sesi" />
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
      {#if sessions.length === 0}
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
                      <LoadingButton
                        variant="secondary"
                        size="sm"
                        onclick={() => renameSession(session.id)}
                        loading={renameSessionLoading === session.id}
                        loadingLabel="Menyimpan..."
                        label="Simpan Nama"
                      />
                    </div>
                  </div>
                </div>
                <LoadingButton
                  variant="outline"
                  size="sm"
                  onclick={() => revokeSession(session.id)}
                  loading={revokeSessionLoading === session.id}
                  loadingLabel="Memproses..."
                  label="Akhiri Sesi"
                />
              </div>
            </div>
          {/each}
        </div>
      {/if}
    </Card.Content>
  </Card.Root>
    {/snippet}
  </AsyncContent>

</div>
