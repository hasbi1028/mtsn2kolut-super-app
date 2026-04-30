<script lang="ts">
  import * as Card from '$lib/components/ui/card';
  import * as Table from '$lib/components/ui/table';
  import { Badge } from '$lib/components/ui/badge';
  import { Button } from '$lib/components/ui/button';
  import { Input } from '$lib/components/ui/input';
  import * as Dialog from '$lib/components/ui/dialog';
  import { toast } from '$lib/components/ui/sonner';
  import { Skeleton } from '$lib/components/ui/skeleton';
  import LoadingButton from '$lib/components/LoadingButton.svelte';

  interface Employee {
    id: string;
    nip: string;
    nama: string;
    unit_kerja: string;
    employment_type: string;
    pusaka_username: string;
    pusaka_is_enabled?: boolean;
    is_active: boolean;
    active_status: string;
    active_run_type: string;
    last_status: string;
    last_run_type: string;
    has_checkin_schedule: boolean;
    has_checkout_schedule: boolean;
  }

  interface EmployeeSchedule {
    id: string;
    run_type: 'checkin' | 'checkout';
    run_time: string;
    is_enabled: boolean;
    random_window_minutes: number;
    day_of_week: number;
  }

  interface AuditLog {
    id: string;
    username: string | null;
    action: string;
    entity_type: string;
    entity_id: string;
    metadata: unknown;
    created_at: string;
  }

  type DayConfig = {
    checkinId: string | null;
    checkoutId: string | null;
    checkinTime: string;
    checkoutTime: string;
    checkinEnabled: boolean;
    checkoutEnabled: boolean;
    randomWindow: number;
  };

  type RunType = 'morning' | 'afternoon' | 'checkin' | 'checkout';

  let { employees, onrun, onstop, ondelete }: {
    employees: Employee[];
    onrun: (id: string, type: string) => void;
    onstop: (id: string, cancelled: number) => void;
    ondelete: () => void;
  } = $props();

  let busyId           = $state<string | null>(null);
  let selectedEmployee = $state<Employee | null>(null);

  // Pusaka dialog
  let showPusakaDialog = $state(false);
  let pusakaUsername   = $state('');
  let pusakaPassword   = $state('');
  let saving           = $state(false);
  let testing          = $state(false);
  let accountToggling  = $state(false);
  let accountDeleting  = $state(false);
  let filterMode = $state<'all' | 'configured' | 'needs_setup' | 'disabled'>('all');
  let search = $state('');
  let showAuditDialog = $state(false);
  let auditLoading = $state(false);
  let auditLogs = $state<AuditLog[]>([]);
  let auditEmployee = $state<Employee | null>(null);

  // Run confirmation dialog
  let runConfirm = $state<{ emp: Employee; runType: RunType } | null>(null);
  let runConfirmInput = $state('');

  // Per-employee schedule dialog
  let showScheduleDialog = $state(false);
  let scheduleEmployee   = $state<Employee | null>(null);
  let scheduleLoading    = $state(false);
  let scheduleSaving     = $state(false);

  const makeDayConfig = (): DayConfig => ({
    checkinId: null, checkoutId: null,
    checkinTime: '', checkoutTime: '',
    checkinEnabled: true, checkoutEnabled: true, randomWindow: 0,
  });
  let dayConfigs = $state<DayConfig[]>(Array.from({ length: 7 }, makeDayConfig));
  let filteredEmployees = $derived.by(() => {
    const normalizedSearch = search.trim().toLowerCase();
    const scoped = filterMode === 'configured'
      ? employees.filter((employee) => !!employee.pusaka_username && employee.pusaka_is_enabled !== false)
      : filterMode === 'needs_setup'
        ? employees.filter((employee) => !employee.pusaka_username)
        : filterMode === 'disabled'
          ? employees.filter((employee) => !!employee.pusaka_username && employee.pusaka_is_enabled === false)
          : employees;
    if (!normalizedSearch) return scoped;
    return scoped.filter((employee) =>
      employee.nama.toLowerCase().includes(normalizedSearch) ||
      employee.nip.toLowerCase().includes(normalizedSearch)
    );
  });

  const runTypeLabel: Record<RunType, string> = {
    morning: 'Rekap', afternoon: 'Rekap', checkin: 'Masuk', checkout: 'Pulang',
  };

  const runTypeDesc: Record<RunType, string> = {
    morning:   'Rekap kehadiran via PUSAKA Kemenag',
    afternoon: 'Rekap kehadiran via PUSAKA Kemenag',
    checkin:   'Rekam absensi MASUK — akan langsung mengeksekusi login ke Pusaka',
    checkout:  'Rekam absensi PULANG — akan langsung mengeksekusi login ke Pusaka',
  };

  function openRunConfirm(emp: Employee, runType: RunType) {
    runConfirm = { emp, runType };
    runConfirmInput = '';
  }

  function submitRunConfirm() {
    if (!runConfirm || runConfirmInput.trim() !== 'SURE') return;
    onrun(runConfirm.emp.id, runConfirm.runType);
    runConfirm = null;
    runConfirmInput = '';
  }

  function closeRunConfirm() {
    runConfirm = null;
    runConfirmInput = '';
  }

  function openPusakaDialog(emp: Employee) {
    selectedEmployee = emp;
    pusakaUsername   = emp.pusaka_username || '';
    pusakaPassword   = '';
    showPusakaDialog = true;
  }

  async function savePusakaCredentials() {
    if (!selectedEmployee) return;
    saving = true;
    try {
      const res = await fetch(`/api/pusaka/employees/${selectedEmployee.id}`, {
        method: 'PUT',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify({ pusaka_username: pusakaUsername, pusaka_password: pusakaPassword }),
      });
      if (!res.ok) {
        const data = await res.json().catch(() => ({}));
        toast.error((data as { error?: string }).error || 'Gagal menyimpan kredensial PUSAKA');
        return;
      }
      toast.success('Kredensial PUSAKA berhasil diperbarui.');
      showPusakaDialog = false;
      ondelete?.();
    } catch {
      toast.error('Gagal menyimpan kredensial PUSAKA');
    } finally { saving = false; }
  }

  async function testPusakaCredentials(emp: Employee) {
    testing = true;
    try {
      const res  = await fetch(`/api/pusaka/employees/${emp.id}/test-pusaka`, { method: 'POST' });
      const data = await res.json() as { message?: string };
      toast.success(data.message || 'Test selesai');
    } catch { toast.error('Test gagal'); } finally { testing = false; }
  }

  async function togglePusakaAccount(emp: Employee, isEnabled: boolean) {
    accountToggling = true;
    try {
      const res = await fetch(`/api/pusaka/employees/${emp.id}`, {
        method: 'PATCH',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify({ is_enabled: isEnabled }),
      });
      if (!res.ok) {
        const data = await res.json().catch(() => ({}));
        toast.error((data as { error?: string }).error || 'Gagal memperbarui status akun PUSAKA');
        return;
      }
      toast.success(isEnabled ? 'Akun PUSAKA diaktifkan kembali.' : 'Akun PUSAKA dinonaktifkan.');
      ondelete?.();
    } finally {
      accountToggling = false;
    }
  }

  async function deletePusakaAccount(emp: Employee) {
    if (!confirm(`Hapus akun PUSAKA untuk ${emp.nama}? Jadwal tetap disimpan, tetapi akun integrasi akan dilepas.`)) return;
    accountDeleting = true;
    try {
      const res = await fetch(`/api/pusaka/employees/${emp.id}`, { method: 'DELETE' });
      if (!res.ok) {
        const data = await res.json().catch(() => ({}));
        toast.error((data as { error?: string }).error || 'Gagal menghapus akun PUSAKA');
        return;
      }
      toast.success('Akun PUSAKA berhasil dihapus.');
      ondelete?.();
    } finally {
      accountDeleting = false;
    }
  }

  function formatAuditDate(iso: string) {
    if (!iso) return '—';
    return new Date(iso).toLocaleString('id-ID', {
      timeZone: 'Asia/Makassar',
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    }) + ' WITA';
  }

  function auditActionLabel(action: string) {
    return {
      PUSAKA_ACCOUNT_UPDATE: 'Update akun',
      PUSAKA_ACCOUNT_TOGGLE: 'Ubah status akun',
      PUSAKA_ACCOUNT_DELETE: 'Hapus akun',
    }[action] ?? action;
  }

  function auditMeta(log: AuditLog): Record<string, unknown> {
    if (!log.metadata) return {};
    if (typeof log.metadata === 'string') {
      try {
        return JSON.parse(log.metadata) as Record<string, unknown>;
      } catch {
        return {};
      }
    }
    if (typeof log.metadata === 'object') {
      return log.metadata as Record<string, unknown>;
    }
    return {};
  }

  async function openAuditDialog(emp: Employee) {
    auditEmployee = emp;
    auditLogs = [];
    showAuditDialog = true;
    auditLoading = true;
    try {
      const res = await fetch(`/api/pusaka/employees/${emp.id}/audit-logs?per_page=10`);
      if (!res.ok) {
        toast.error('Gagal memuat riwayat akun PUSAKA');
        return;
      }
      auditLogs = await res.json() as AuditLog[];
    } catch {
      toast.error('Gagal memuat riwayat akun PUSAKA');
    } finally {
      auditLoading = false;
    }
  }

  async function doStop(id: string) {
    busyId = id;
    const res  = await fetch('/api/pusaka/jobs/cancel', {
      method: 'POST', headers: { 'content-type': 'application/json' },
      body: JSON.stringify({ employee_id: id }),
    });
    const data = await res.json().catch(() => ({})) as { cancelled?: number };
    busyId = null;
    onstop?.(id, data.cancelled ?? 0);
  }

  function populateDayConfigs(data: EmployeeSchedule[]): DayConfig[] {
    const configs = Array.from({ length: 7 }, makeDayConfig);
    for (const s of data) {
      const d = s.day_of_week;
      if (d < 0 || d > 6) continue;
      if (s.run_type === 'checkin') {
        configs[d].checkinId      = s.id;
        configs[d].checkinTime    = s.run_time;
        configs[d].checkinEnabled = s.is_enabled;
        configs[d].randomWindow   = s.random_window_minutes ?? 0;
      } else {
        configs[d].checkoutId      = s.id;
        configs[d].checkoutTime    = s.run_time;
        configs[d].checkoutEnabled = s.is_enabled;
        if (!configs[d].checkinId) configs[d].randomWindow = s.random_window_minutes ?? 0;
      }
    }
    return configs;
  }

  async function openScheduleDialog(emp: Employee) {
    scheduleEmployee   = emp;
    showScheduleDialog = true;
    scheduleLoading    = true;
    dayConfigs         = Array.from({ length: 7 }, makeDayConfig);
    try {
      const res  = await fetch(`/api/pusaka/employees/${emp.id}/schedules`);
      const data = await res.json().catch(() => []) as EmployeeSchedule[];
      if (Array.isArray(data)) dayConfigs = populateDayConfigs(data);
    } catch { /* silent */ }
    finally { scheduleLoading = false; }
  }

  async function saveDayRow(dow: number) {
    if (!scheduleEmployee) return;
    const cfg = dayConfigs[dow];
    scheduleSaving = true;
    try {
      if (cfg.checkinTime) {
        await fetch(`/api/pusaka/employees/${scheduleEmployee.id}/schedules`, {
          method: 'POST', headers: { 'content-type': 'application/json' },
          body: JSON.stringify({
            run_type: 'checkin', run_time: cfg.checkinTime,
            is_enabled: cfg.checkinEnabled,
            random_window_minutes: cfg.randomWindow,
            day_of_week: dow,
          }),
        });
      }
      if (cfg.checkoutTime) {
        await fetch(`/api/pusaka/employees/${scheduleEmployee.id}/schedules`, {
          method: 'POST', headers: { 'content-type': 'application/json' },
          body: JSON.stringify({
            run_type: 'checkout', run_time: cfg.checkoutTime,
            is_enabled: cfg.checkoutEnabled,
            random_window_minutes: cfg.randomWindow,
            day_of_week: dow,
          }),
        });
      }
      const res  = await fetch(`/api/pusaka/employees/${scheduleEmployee.id}/schedules`);
      const data = await res.json().catch(() => []) as EmployeeSchedule[];
      if (Array.isArray(data)) dayConfigs = populateDayConfigs(data);
      ondelete?.();
    } catch { /* silent */ }
    finally { scheduleSaving = false; }
  }

  async function deleteDaySchedule(dow: number, runType: 'checkin' | 'checkout') {
    if (!scheduleEmployee) return;
    const cfg    = dayConfigs[dow];
    const schedId = runType === 'checkin' ? cfg.checkinId : cfg.checkoutId;
    if (!schedId) return;
    try {
      await fetch(`/api/pusaka/employees/${scheduleEmployee.id}/schedules/${schedId}`, { method: 'DELETE' });
      if (runType === 'checkin') { cfg.checkinId = null; cfg.checkinTime = ''; }
      else                       { cfg.checkoutId = null; cfg.checkoutTime = ''; }
      ondelete?.();
    } catch { /* silent */ }
  }

  function statusInfo(e: Employee) {
    const s = e.active_status || e.last_status;
    const t = e.active_run_type || e.last_run_type;
    if (!s) return null;
    const labels: Record<string, string> = { morning: 'Rekap', afternoon: 'Rekap', checkin: 'Masuk', checkout: 'Pulang' };
    return { status: s, tipe: labels[t] ?? t };
  }

  function statusVariant(s: string): 'default' | 'destructive' | 'outline' | 'secondary' {
    if (s === 'running') return 'default';
    if (s === 'success') return 'default';
    if (s === 'failed')  return 'destructive';
    if (s === 'queued')  return 'secondary';
    return 'outline';
  }

  function statusLabel(s: string) {
    return { running: 'Berjalan', success: 'Sukses', failed: 'Gagal', queued: 'Antre' }[s] ?? s;
  }

  function isPusakaConfigured(emp: Employee) {
    return !!(emp.pusaka_username);
  }

  function pusakaStatusLabel(emp: Employee) {
    if (!emp.pusaka_username) return 'Belum setup';
    return emp.pusaka_is_enabled === false ? 'Dinonaktifkan' : 'Aktif';
  }

  function scheduleButtonClass(emp: Employee): string {
    const ci = emp.has_checkin_schedule;
    const co = emp.has_checkout_schedule;
    if (ci && co)  return 'border-green-400 text-green-700 bg-green-50 hover:bg-green-100';
    if (ci)        return 'border-amber-400 text-amber-700 bg-amber-50 hover:bg-amber-100';
    if (co)        return 'border-cyan-400 text-cyan-700 bg-cyan-50 hover:bg-cyan-100';
    return 'border-slate-300 text-slate-500 hover:bg-slate-50';
  }

  function scheduleButtonLabel(emp: Employee): string {
    const ci = emp.has_checkin_schedule;
    const co = emp.has_checkout_schedule;
    if (ci && co) return '📅 Jadwal ✓';
    if (ci)       return '📅 Masuk ✓';
    if (co)       return '📅 Pulang ✓';
    return '📅 Jadwal';
  }

  const canConfirmRun = $derived(runConfirmInput.trim() === 'SURE');
</script>

<Card.Root>
  <Card.Header class="pb-3">
    <div class="flex items-center justify-between">
      <div>
        <Card.Title class="text-base">Pegawai Eligible PUSAKA</Card.Title>
        <Card.Description>Hanya pegawai PNS dan PPPK yang dikelola di area ini untuk setup akun, jadwal, dan eksekusi job PUSAKA.</Card.Description>
      </div>
      <div class="flex items-center gap-2">
        <Input placeholder="Cari nama / NIP..." bind:value={search} class="w-44" />
        <select bind:value={filterMode} class="rounded-md border border-input bg-background px-3 py-2 text-sm">
          <option value="all">Semua</option>
          <option value="configured">Akun aktif</option>
          <option value="needs_setup">Belum setup</option>
          <option value="disabled">Dinonaktifkan</option>
        </select>
        <Badge variant="secondary">{filteredEmployees.length} pegawai</Badge>
      </div>
    </div>
  </Card.Header>
  <Card.Content class="p-0 overflow-x-auto">
    <Table.Root>
      <Table.Header>
        <Table.Row>
          <Table.Head>Nama / NIP</Table.Head>
          <Table.Head class="hidden sm:table-cell">Unit Kerja</Table.Head>
          <Table.Head class="text-center">Pusaka</Table.Head>
          <Table.Head class="text-center">Status</Table.Head>
          <Table.Head class="text-right">Aksi</Table.Head>
        </Table.Row>
      </Table.Header>
      <Table.Body>
        {#each filteredEmployees as e (e.id)}
          {@const si = statusInfo(e)}
          <Table.Row class={e.active_status === 'running' ? 'bg-amber-50' : ''}>
            <Table.Cell>
              <div class="font-medium">{e.nama}</div>
              <div class="text-xs text-muted-foreground font-mono">{e.nip}</div>
              <div class="mt-1">
                {#if e.is_active}
                  <Badge variant="outline" class="text-[11px] border-emerald-300 text-emerald-700">Pegawai aktif</Badge>
                {:else}
                  <Badge variant="secondary" class="text-[11px]">Nonaktif / rotasi</Badge>
                {/if}
              </div>
            </Table.Cell>
            <Table.Cell class="hidden sm:table-cell text-sm text-muted-foreground">
              {e.unit_kerja || '—'}
            </Table.Cell>
            <Table.Cell class="text-center">
              {#if isPusakaConfigured(e)}
                <Badge variant={e.pusaka_is_enabled === false ? 'secondary' : 'outline'} class={e.pusaka_is_enabled === false ? 'text-xs' : 'text-xs border-green-300 text-green-700'}>
                  {pusakaStatusLabel(e)}
                </Badge>
              {:else}
                <Badge variant="destructive" class="text-xs">Belum</Badge>
              {/if}
            </Table.Cell>
            <Table.Cell class="text-center">
              {#if si}
                <Badge variant={statusVariant(si.status)} class="text-xs">
                  {statusLabel(si.status)}{si.tipe ? ' · ' + si.tipe : ''}
                </Badge>
              {:else}
                <span class="text-xs text-muted-foreground">—</span>
              {/if}
            </Table.Cell>
            <Table.Cell class="text-right">
              <div class="flex items-center justify-end gap-1.5 flex-wrap">
                <Button size="sm" variant="outline" onclick={() => openPusakaDialog(e)}>
                  {isPusakaConfigured(e) ? 'Edit' : 'Setup'} Pusaka
                </Button>
                {#if isPusakaConfigured(e)}
                  <LoadingButton
                    size="sm"
                    variant="outline"
                    onclick={() => togglePusakaAccount(e, e.pusaka_is_enabled === false)}
                    loading={accountToggling}
                    loadingLabel="Memproses..."
                    disabled={accountToggling}
                  >
                    {e.pusaka_is_enabled === false ? 'Aktifkan Akun' : 'Nonaktifkan Akun'}
                  </LoadingButton>
                  <LoadingButton
                    size="sm"
                    variant="ghost"
                    class="text-destructive hover:text-destructive"
                    onclick={() => deletePusakaAccount(e)}
                    loading={accountDeleting}
                    loadingLabel="Menghapus..."
                    disabled={accountDeleting}
                  >
                    Hapus Akun
                  </LoadingButton>
                {/if}
                <Button size="sm" variant="outline" onclick={() => openAuditDialog(e)}>
                  Riwayat
                </Button>
                <LoadingButton size="sm" variant="ghost" onclick={() => testPusakaCredentials(e)} loading={testing} loadingLabel="Testing..." disabled={testing || !isPusakaConfigured(e)}>
                  Test
                </LoadingButton>
                <LoadingButton size="sm" variant="outline" onclick={() => onrun(e.id, 'morning')} loading={busyId === e.id} loadingLabel="Memproses..." disabled={busyId === e.id}>
                  Rekap
                </LoadingButton>
                <Button size="sm" variant="outline"
                  onclick={() => openRunConfirm(e, 'checkin')}
                  disabled={busyId === e.id}
                  class="border-amber-300 text-amber-700 hover:bg-amber-50">
                  ☀ Masuk
                </Button>
                <Button size="sm" variant="outline"
                  onclick={() => openRunConfirm(e, 'checkout')}
                  disabled={busyId === e.id}
                  class="border-amber-300 text-amber-700 hover:bg-amber-50">
                  🌙 Pulang
                </Button>
                <Button size="sm" variant="outline"
                  onclick={() => openScheduleDialog(e)}
                  class={scheduleButtonClass(e)}>
                  {scheduleButtonLabel(e)}
                </Button>
                <LoadingButton size="sm" variant="ghost" onclick={() => doStop(e.id)} loading={busyId === e.id} loadingLabel="Memproses..." disabled={busyId === e.id || !e.active_status}
                  class="text-amber-700 hover:text-amber-800">
                  ■ Stop
                </LoadingButton>
              </div>
            </Table.Cell>
          </Table.Row>
        {:else}
          <Table.Row>
            <Table.Cell colspan={5} class="py-12 text-center text-muted-foreground">
              Belum ada data pegawai.
            </Table.Cell>
          </Table.Row>
        {/each}
      </Table.Body>
    </Table.Root>
  </Card.Content>
</Card.Root>

<!-- Dialog Konfirmasi Run Masuk / Pulang -->
{#if runConfirm}
  <Dialog.Root open={true}>
    <Dialog.Content>
      <Dialog.Header>
        <Dialog.Title>
          Konfirmasi {runTypeLabel[runConfirm.runType]} — {runConfirm.emp.nama}
        </Dialog.Title>
        <Dialog.Description>
          {runTypeDesc[runConfirm.runType]}
        </Dialog.Description>
      </Dialog.Header>

      <div class="py-4 space-y-4">
        <div class="rounded-md border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-800">
          Aksi ini akan langsung menjalankan job ke Pusaka Kemenag dan <strong>tidak bisa dibatalkan</strong> setelah dieksekusi.
          Pastikan waktu dan pegawai sudah benar.
        </div>
        <div class="space-y-1.5">
          <label for="run-confirm-input" class="text-sm font-medium">
            Ketik <code class="rounded bg-slate-100 px-1.5 py-0.5 font-mono text-xs font-bold">SURE</code> untuk melanjutkan
          </label>
          <Input
            id="run-confirm-input"
            bind:value={runConfirmInput}
            placeholder="Ketik SURE"
            class="font-mono uppercase"
            onkeydown={(e) => { if (e.key === 'Enter' && canConfirmRun) submitRunConfirm(); }}
          />
        </div>
      </div>

      <Dialog.Footer>
        <Button variant="outline" onclick={closeRunConfirm}>Batal</Button>
        <Button
          onclick={submitRunConfirm}
          disabled={!canConfirmRun}
          class={canConfirmRun ? 'bg-amber-600 hover:bg-amber-700 text-white border-transparent' : ''}>
          Jalankan {runTypeLabel[runConfirm.runType]}
        </Button>
      </Dialog.Footer>
    </Dialog.Content>
  </Dialog.Root>
{/if}

<!-- Dialog Kredensial Pusaka -->
<Dialog.Root bind:open={showPusakaDialog}>
  <Dialog.Content>
    <Dialog.Header>
      <Dialog.Title>Kredensial Pusaka Kemenag</Dialog.Title>
      <Dialog.Description>
        {selectedEmployee ? `Pegawai: ${selectedEmployee.nama}` : ''}
      </Dialog.Description>
    </Dialog.Header>
    <div class="space-y-4 py-4">
      <div class="space-y-1.5">
        <label for="pusaka-username" class="text-sm font-medium">Username Pusaka</label>
        <Input id="pusaka-username" bind:value={pusakaUsername} placeholder="Username Pusaka Kemenag" />
      </div>
      <div class="space-y-1.5">
        <label for="pusaka-password" class="text-sm font-medium">Password Pusaka</label>
        <Input id="pusaka-password" type="password" bind:value={pusakaPassword} placeholder="Kosongkan jika tidak diubah" />
      </div>
    </div>
    <Dialog.Footer>
      <Button variant="outline" onclick={() => (showPusakaDialog = false)}>Batal</Button>
      <LoadingButton onclick={savePusakaCredentials} loading={saving} loadingLabel="Menyimpan..." disabled={saving}>Simpan</LoadingButton>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>

<Dialog.Root bind:open={showAuditDialog}>
  <Dialog.Content>
    <Dialog.Header>
      <Dialog.Title>Riwayat Akun PUSAKA</Dialog.Title>
      <Dialog.Description>
        {auditEmployee ? `10 aktivitas terakhir untuk ${auditEmployee.nama}` : ''}
      </Dialog.Description>
    </Dialog.Header>

    <div class="space-y-3 py-2">
      {#if auditLoading}
        <div class="space-y-3 rounded-md border border-slate-200 bg-slate-50 px-3 py-4">
          {#each Array.from({ length: 3 }) as _, index (`audit-skeleton-${index}`)}
            <div class="rounded-xl border border-slate-200 bg-white px-3 py-3">
              <div class="flex items-start justify-between gap-3">
                <div class="space-y-2">
                  <Skeleton class="h-5 w-28" />
                  <Skeleton class="h-4 w-36" />
                </div>
                <Skeleton class="h-6 w-16" />
              </div>
              <div class="mt-3 space-y-2">
                <Skeleton class="h-4 w-40" />
                <Skeleton class="h-4 w-28" />
              </div>
            </div>
          {/each}
        </div>
      {:else if auditLogs.length === 0}
        <div class="rounded-md border border-slate-200 bg-slate-50 px-3 py-6 text-center text-sm text-slate-500">Belum ada riwayat akun PUSAKA untuk pegawai ini.</div>
      {:else}
        <div class="space-y-2">
          {#each auditLogs as log (log.id)}
            {@const meta = auditMeta(log)}
            <div class="rounded-xl border border-slate-200 bg-white px-3 py-3">
              <div class="flex items-start justify-between gap-3">
                <div>
                  <p class="text-sm font-semibold text-slate-900">{auditActionLabel(log.action)}</p>
                  <p class="mt-1 text-xs text-slate-500">{formatAuditDate(log.created_at)}</p>
                </div>
                <Badge variant="outline" class="text-[11px]">{log.username ?? 'Sistem'}</Badge>
              </div>
              {#if meta.pusaka_username}
                <p class="mt-2 text-xs text-slate-600">Username: <span class="font-mono">{String(meta.pusaka_username)}</span></p>
              {/if}
              {#if typeof meta.is_enabled === 'boolean'}
                <p class="mt-1 text-xs text-slate-600">Status akun: {meta.is_enabled ? 'aktif' : 'dinonaktifkan'}</p>
              {/if}
            </div>
          {/each}
        </div>
      {/if}
    </div>

    <Dialog.Footer>
      <Button variant="outline" onclick={() => (showAuditDialog = false)}>Tutup</Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>

<!-- Dialog Jadwal Per-Pegawai (7-hari) -->
<Dialog.Root bind:open={showScheduleDialog}>
  <Dialog.Content>
    <Dialog.Header>
      <Dialog.Title>Jadwal Absensi</Dialog.Title>
      <Dialog.Description>{scheduleEmployee?.nama ?? ''}</Dialog.Description>
    </Dialog.Header>

    {#if scheduleLoading}
      <div class="space-y-3 py-2">
        {#each Array.from({ length: 4 }) as _, index (`schedule-skeleton-${index}`)}
          <div class="rounded-lg border bg-card px-3 py-3 space-y-3">
            <div class="flex items-center justify-between">
              <Skeleton class="h-5 w-20" />
              <Skeleton class="h-6 w-24" />
            </div>
            <div class="space-y-2">
              <Skeleton class="h-8 w-full" />
              <Skeleton class="h-8 w-full" />
            </div>
            <div class="flex justify-end">
              <Skeleton class="h-7 w-20" />
            </div>
          </div>
        {/each}
      </div>
    {:else}
      {@const dayLabels = ['Minggu', 'Senin', 'Selasa', 'Rabu', 'Kamis', 'Jumat', 'Sabtu']}
      <div class="space-y-2 max-h-[65vh] overflow-y-auto py-1 pr-1">
        {#each dayConfigs as cfg, dow (`${dow}-${cfg.checkinId ?? 'ci'}-${cfg.checkoutId ?? 'co'}`)}
          <div class="rounded-lg border bg-card px-3 py-2.5 space-y-2">

            <!-- Header baris hari -->
            <div class="flex items-center justify-between">
              <span class="text-sm font-semibold">{dayLabels[dow]}</span>
              <div class="flex items-center gap-1.5 text-xs text-muted-foreground">
                <span>Acak ±</span>
                <input type="number" min="0" max="60" bind:value={cfg.randomWindow}
                  class="w-12 rounded border border-input bg-background px-1.5 py-0.5 text-xs text-center" />
                <span>mnt</span>
              </div>
            </div>

            <!-- Baris masuk + pulang -->
            <div class="space-y-1.5">
              <div class="flex items-center gap-2">
                <span class="w-14 shrink-0 text-xs text-muted-foreground">☀ Masuk</span>
                <input type="time" bind:value={cfg.checkinTime}
                  class="flex-1 min-w-0 rounded border border-input bg-background px-2 py-1 font-mono text-xs" />
                <label class="flex items-center gap-1 text-xs text-muted-foreground cursor-pointer shrink-0">
                  <input type="checkbox" id="ci-{dow}" bind:checked={cfg.checkinEnabled}
                    class="h-3.5 w-3.5 rounded accent-green-700" />
                  Aktif
                </label>
                {#if cfg.checkinId}
                  <button type="button"
                    class="text-xs text-destructive hover:text-destructive/80 shrink-0"
                    onclick={() => deleteDaySchedule(dow, 'checkin')}
                    title="Hapus jadwal masuk">✕</button>
                {/if}
              </div>
              <div class="flex items-center gap-2">
                <span class="w-14 shrink-0 text-xs text-muted-foreground">🌙 Pulang</span>
                <input type="time" bind:value={cfg.checkoutTime}
                  class="flex-1 min-w-0 rounded border border-input bg-background px-2 py-1 font-mono text-xs" />
                <label class="flex items-center gap-1 text-xs text-muted-foreground cursor-pointer shrink-0">
                  <input type="checkbox" id="co-{dow}" bind:checked={cfg.checkoutEnabled}
                    class="h-3.5 w-3.5 rounded accent-green-700" />
                  Aktif
                </label>
                {#if cfg.checkoutId}
                  <button type="button"
                    class="text-xs text-destructive hover:text-destructive/80 shrink-0"
                    onclick={() => deleteDaySchedule(dow, 'checkout')}
                    title="Hapus jadwal pulang">✕</button>
                {/if}
              </div>
            </div>

            <!-- Tombol simpan -->
            <div class="flex justify-end">
              <LoadingButton size="sm" variant="outline"
                onclick={() => saveDayRow(dow)}
                loading={scheduleSaving}
                loadingLabel="Menyimpan..."
                disabled={scheduleSaving || (!cfg.checkinTime && !cfg.checkoutTime)}
                class="h-7 px-3 text-xs">
                Simpan
              </LoadingButton>
            </div>

          </div>
        {/each}
      </div>
    {/if}

    <Dialog.Footer>
      <Button variant="outline" onclick={() => (showScheduleDialog = false)}>Tutup</Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
