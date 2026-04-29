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
    active_status: string;
    active_run_type: string;
    last_status: string;
    last_run_type: string;
  }

  interface EmployeeSchedule {
    id: string;
    run_type: string;
    run_time: string;
    is_enabled: boolean;
  }

  type RunType = 'morning' | 'afternoon' | 'checkin' | 'checkout';

  let { employees, onrun, onstop, ondelete }: {
    employees: Employee[];
    onrun: (id: string, type: string) => void;
    onstop: (id: string, cancelled: number) => void;
    ondelete: () => void;
  } = $props();

  let confirmId        = $state<string | null>(null);
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
  let showScheduleDialog      = $state(false);
  let scheduleEmployee        = $state<Employee | null>(null);
  let empSchedules            = $state<EmployeeSchedule[]>([]);
  let scheduleCheckinTime     = $state('');
  let scheduleCheckoutTime    = $state('');
  let scheduleCheckinEnabled  = $state(true);
  let scheduleCheckoutEnabled = $state(true);
  let scheduleLoading = $state(false);
  let scheduleSaving  = $state(false);

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
      const res = await fetch(`/api/employees/${selectedEmployee.id}`, {
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
      const res  = await fetch(`/api/employees/${emp.id}/test-pusaka`, { method: 'POST' });
      const data = await res.json() as { message?: string };
      alert(data.message || 'Test selesai');
    } catch { alert('Test gagal'); } finally { testing = false; }
  }

  async function doStop(id: string) {
    busyId = id;
    const res  = await fetch('/api/jobs/cancel', {
      method: 'POST', headers: { 'content-type': 'application/json' },
      body: JSON.stringify({ employee_id: id }),
    });
    const data = await res.json().catch(() => ({})) as { cancelled?: number };
    busyId = null;
    onstop?.(id, data.cancelled ?? 0);
  }

  async function doDelete(id: string) {
    busyId = id;
    await fetch('/api/employees', {
      method: 'DELETE', headers: { 'content-type': 'application/json' },
      body: JSON.stringify({ id }),
    });
    busyId = null;
    confirmId = null;
    ondelete?.();
  }

  async function openScheduleDialog(emp: Employee) {
    scheduleEmployee = emp;
    showScheduleDialog = true;
    scheduleLoading = true;
    scheduleCheckinTime = '';
    scheduleCheckoutTime = '';
    scheduleCheckinEnabled = true;
    scheduleCheckoutEnabled = true;
    try {
      const res  = await fetch(`/api/employees/${emp.id}/schedules`);
      const data = await res.json().catch(() => ([])) as EmployeeSchedule[];
      empSchedules = Array.isArray(data) ? data : [];
      const ci = empSchedules.find(s => s.run_type === 'checkin');
      const co = empSchedules.find(s => s.run_type === 'checkout');
      if (ci) { scheduleCheckinTime = ci.run_time; scheduleCheckinEnabled = ci.is_enabled; }
      if (co) { scheduleCheckoutTime = co.run_time; scheduleCheckoutEnabled = co.is_enabled; }
    } catch { empSchedules = []; }
    finally { scheduleLoading = false; }
  }

  async function saveEmployeeSchedule(runType: 'checkin' | 'checkout') {
    if (!scheduleEmployee) return;
    const runTime   = runType === 'checkin' ? scheduleCheckinTime : scheduleCheckoutTime;
    const isEnabled = runType === 'checkin' ? scheduleCheckinEnabled : scheduleCheckoutEnabled;
    if (!runTime) return;
    scheduleSaving = true;
    try {
      await fetch(`/api/employees/${scheduleEmployee.id}/schedules`, {
        method: 'POST',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify({ run_type: runType, run_time: runTime, is_enabled: isEnabled }),
      });
      // Reload schedules to get updated id
      const res  = await fetch(`/api/employees/${scheduleEmployee.id}/schedules`);
      empSchedules = await res.json().catch(() => ([])) as EmployeeSchedule[];
    } catch { /* silent */ }
    finally { scheduleSaving = false; }
  }

  async function deleteEmployeeSchedule(runType: 'checkin' | 'checkout') {
    if (!scheduleEmployee) return;
    const sched = empSchedules.find(s => s.run_type === runType);
    if (!sched) return;
    try {
      await fetch(`/api/employees/${scheduleEmployee.id}/schedules/${sched.id}`, { method: 'DELETE' });
      if (runType === 'checkin') scheduleCheckinTime = '';
      else scheduleCheckoutTime = '';
      empSchedules = empSchedules.filter(s => s.run_type !== runType);
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

  const canConfirmRun = $derived(runConfirmInput.trim() === 'SURE');
</script>

<Card.Root>
  <Card.Header class="pb-3">
    <div class="flex items-center justify-between">
      <Card.Title class="text-base">Daftar Pegawai</Card.Title>
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
        {#each employees as e}
          {@const si = statusInfo(e)}
          <Table.Row class={e.active_status === 'running' ? 'bg-amber-50' : ''}>
            <Table.Cell>
              <div class="font-medium">{e.nama}</div>
              <div class="text-xs text-muted-foreground font-mono">{e.nip}</div>
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
              {#if confirmId === e.id}
                <div class="flex items-center justify-end gap-2 flex-wrap">
                  <span class="text-xs text-amber-700">Hapus beserta semua data?</span>
                  <Button size="sm" variant="destructive" onclick={() => doDelete(e.id)} disabled={busyId === e.id}>
                    {busyId === e.id ? '...' : 'Ya, Hapus'}
                  </Button>
                  <Button size="sm" variant="ghost" onclick={() => (confirmId = null)}>Batal</Button>
                </div>
              {:else}
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
                    class="border-slate-300 text-slate-600 hover:bg-slate-50">
                    Jadwal
                  </Button>
                  <Button size="sm" variant="ghost" onclick={() => doStop(e.id)} disabled={busyId === e.id || !e.active_status}
                    class="text-amber-700 hover:text-amber-800">
                    ■ Stop
                  </Button>
                  <Button size="sm" variant="ghost" onclick={() => (confirmId = e.id)}
                    class="text-destructive hover:text-destructive">
                    Hapus
                  </Button>
                </div>
              {/if}
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

<!-- Dialog Jadwal Per-Pegawai -->
<Dialog.Root bind:open={showScheduleDialog}>
  <Dialog.Content>
    <Dialog.Header>
      <Dialog.Title>Jadwal Absensi — {scheduleEmployee?.nama ?? ''}</Dialog.Title>
      <Dialog.Description>
        Atur jam otomatis absen masuk dan pulang untuk pegawai ini
      </Dialog.Description>
    </Dialog.Header>

    {#if scheduleLoading}
      <div class="py-8 text-center text-sm text-muted-foreground">Memuat...</div>
    {:else}
      <div class="space-y-4 py-4">

        <!-- Checkin -->
        <div class="rounded-md border p-4 space-y-3">
          <div class="flex items-center justify-between">
            <span class="text-sm font-medium">☀ Absen Masuk</span>
            <div class="flex items-center gap-2">
              <input type="checkbox" id="ci-enabled" bind:checked={scheduleCheckinEnabled}
                class="h-4 w-4 rounded border-input accent-green-700" />
              <label for="ci-enabled" class="text-xs text-muted-foreground">Aktif</label>
            </div>
          </div>
          <div class="flex gap-2">
            <Input type="time" bind:value={scheduleCheckinTime} class="font-mono flex-1" />
            <Button size="sm" variant="outline"
              onclick={() => saveEmployeeSchedule('checkin')}
              disabled={scheduleSaving || !scheduleCheckinTime}>
              Simpan
            </Button>
            {#if empSchedules.find(s => s.run_type === 'checkin')}
              <Button size="sm" variant="ghost" class="text-destructive hover:text-destructive"
                onclick={() => deleteEmployeeSchedule('checkin')}>
                Hapus
              </Button>
            {/if}
          </div>
          {#if !empSchedules.find(s => s.run_type === 'checkin')}
            <p class="text-xs text-muted-foreground">Belum ada jadwal masuk terpasang</p>
          {/if}
        </div>

        <!-- Checkout -->
        <div class="rounded-md border p-4 space-y-3">
          <div class="flex items-center justify-between">
            <span class="text-sm font-medium">🌙 Absen Pulang</span>
            <div class="flex items-center gap-2">
              <input type="checkbox" id="co-enabled" bind:checked={scheduleCheckoutEnabled}
                class="h-4 w-4 rounded border-input accent-green-700" />
              <label for="co-enabled" class="text-xs text-muted-foreground">Aktif</label>
            </div>
          </div>
          <div class="flex gap-2">
            <Input type="time" bind:value={scheduleCheckoutTime} class="font-mono flex-1" />
            <Button size="sm" variant="outline"
              onclick={() => saveEmployeeSchedule('checkout')}
              disabled={scheduleSaving || !scheduleCheckoutTime}>
              Simpan
            </Button>
            {#if empSchedules.find(s => s.run_type === 'checkout')}
              <Button size="sm" variant="ghost" class="text-destructive hover:text-destructive"
                onclick={() => deleteEmployeeSchedule('checkout')}>
                Hapus
              </Button>
            {/if}
          </div>
          {#if !empSchedules.find(s => s.run_type === 'checkout')}
            <p class="text-xs text-muted-foreground">Belum ada jadwal pulang terpasang</p>
          {/if}
        </div>

      </div>
    {/if}

    <Dialog.Footer>
      <Button variant="outline" onclick={() => (showScheduleDialog = false)}>Tutup</Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
