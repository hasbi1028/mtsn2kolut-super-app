<script lang="ts">
  import * as Card from '$lib/components/ui/card';
  import * as Table from '$lib/components/ui/table';
  import { Badge } from '$lib/components/ui/badge';
  import { Button } from '$lib/components/ui/button';
  import { Input } from '$lib/components/ui/input';
  import * as Dialog from '$lib/components/ui/dialog';

  interface Employee {
    id: string;
    nip: string;
    nama: string;
    unit_kerja: string;
    pusaka_username: string;
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
      if (res.ok) { showPusakaDialog = false; ondelete?.(); }
    } catch { /* ignore */ } finally { saving = false; }
  }

  async function testPusakaCredentials(emp: Employee) {
    testing = true;
    try {
      const res  = await fetch(`/api/pusaka/employees/${emp.id}/test-pusaka`, { method: 'POST' });
      const data = await res.json() as { message?: string };
      alert(data.message || 'Test selesai');
    } catch { alert('Test gagal'); } finally { testing = false; }
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
      <Badge variant="secondary">{employees.length} pegawai</Badge>
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
        {#each employees as e (e.id)}
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
                <Badge variant="outline" class="text-xs border-green-300 text-green-700">✓ Aktif</Badge>
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
                <Button size="sm" variant="ghost" onclick={() => testPusakaCredentials(e)} disabled={testing || !isPusakaConfigured(e)}>
                  Test
                </Button>
                <Button size="sm" variant="outline" onclick={() => onrun(e.id, 'morning')} disabled={busyId === e.id}>
                  Rekap
                </Button>
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
                <Button size="sm" variant="ghost" onclick={() => doStop(e.id)} disabled={busyId === e.id || !e.active_status}
                  class="text-amber-700 hover:text-amber-800">
                  ■ Stop
                </Button>
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
      <Button onclick={savePusakaCredentials} disabled={saving}>
        {saving ? 'Menyimpan...' : 'Simpan'}
      </Button>
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
      <div class="py-8 text-center text-sm text-muted-foreground">Memuat...</div>
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
              <Button size="sm" variant="outline"
                onclick={() => saveDayRow(dow)}
                disabled={scheduleSaving || (!cfg.checkinTime && !cfg.checkoutTime)}
                class="h-7 px-3 text-xs">
                {scheduleSaving ? 'Menyimpan...' : 'Simpan'}
              </Button>
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
