<script lang="ts">
  import * as Card from '$lib/components/ui/card';
  import * as Table from '$lib/components/ui/table';
  import { Badge } from '$lib/components/ui/badge';
  import { Button } from '$lib/components/ui/button';
  import { Input } from '$lib/components/ui/input';
  import PasswordInput from '$lib/components/PasswordInput.svelte';
  import * as Dialog from '$lib/components/ui/dialog';
  import { toast } from '$lib/components/ui/sonner';
  import { Skeleton } from '$lib/components/ui/skeleton';
  import AsyncContent from '$lib/components/AsyncContent.svelte';
  import LoadingButton from '$lib/components/LoadingButton.svelte';
  import EmptyStatePanel from '$lib/components/EmptyStatePanel.svelte';
  import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
  import SuccessPanel from '$lib/components/SuccessPanel.svelte';
  import OperationStatusPanel from '$lib/components/OperationStatusPanel.svelte';
  import Pagination from '$lib/components/Pagination.svelte';
  import { confirmAction, confirmChallenge } from '$lib/confirm-dialog';
  import { readClientApiData, readClientJson } from '$lib/client/api';

  interface Employee {
    id: string;
    pegawai_uid: string;
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
    onrun: (id: string, type: string) => void | Promise<void>;
    onstop: (id: string, cancelled: number) => void | Promise<void>;
    ondelete: () => void | Promise<void>;
  } = $props();

  let busyId           = $state<string | null>(null);
  let selectedEmployee = $state<Employee | null>(null);

  // Pusaka dialog
  let showPusakaDialog = $state(false);
  let pusakaUsername   = $state('');
  let pusakaPassword   = $state('');
  let saving           = $state(false);
  let testingId        = $state<string | null>(null);
  let accountTogglingId = $state<string | null>(null);
  let accountDeletingId = $state<string | null>(null);
  let filterMode = $state<'all' | 'configured' | 'needs_setup' | 'disabled'>('all');
  let search = $state('');
  let page = $state(1);
  const PER_PAGE = 12;
  let success = $state('');
  let operationState = $state<{ tone: 'success' | 'error' | 'warning' | 'info'; title: string; message: string } | null>(null);
  let showAuditDialog = $state(false);
  let auditPromise = $state<Promise<AuditLog[]> | null>(null);
  let auditLogs = $state<AuditLog[]>([]);
  let auditEmployee = $state<Employee | null>(null);

  // Run confirmation dialog
  let runConfirm = $state<{ emp: Employee; runType: RunType } | null>(null);
  let runConfirmInput = $state('');

  // Per-employee schedule dialog
  let showScheduleDialog = $state(false);
  let scheduleEmployee   = $state<Employee | null>(null);
  let schedulePromise    = $state<Promise<DayConfig[]> | null>(null);
  let scheduleSaving     = $state(false);
  let scheduleDeletingKey = $state<string | null>(null);

  const makeDayConfig = (): DayConfig => ({
    checkinId: null, checkoutId: null,
    checkinTime: '', checkoutTime: '',
    checkinEnabled: true, checkoutEnabled: true, randomWindow: 0,
  });
  let dayConfigs = $state<DayConfig[]>(Array.from({ length: 7 }, makeDayConfig));
  const filteredEmployees = $derived.by(() => {
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
      employee.pegawai_uid.toLowerCase().includes(normalizedSearch) ||
      employee.nip.toLowerCase().includes(normalizedSearch)
    );
  });

  const pagedEmployees = $derived(filteredEmployees.slice((page - 1) * PER_PAGE, page * PER_PAGE));
  const totalFiltered = $derived(filteredEmployees.length);

  function handleSearch(val: string) { search = val; page = 1; }
  function handleFilter(val: 'all' | 'configured' | 'needs_setup' | 'disabled') { filterMode = val; page = 1; }

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
    success = `Job ${runTypeLabel[runConfirm.runType]} untuk ${runConfirm.emp.nama} berhasil diantrekan. Pantau statusnya di kolom operasi atau riwayat job terbaru.`;
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

  function confirmPhrase(title: string, detail: string, challenge: string) {
    return confirmChallenge({
      title,
      message: detail,
      challenge,
      confirmLabel: 'Konfirmasi',
      tone: 'danger'
    });
  }

  function showOperationError(title: string, message: string) {
    operationState = { tone: 'error', title, message };
    toast.error(message);
  }

  async function ensureMutationOk(response: Response, fallbackMessage: string) {
    try {
      return await readClientJson<unknown>(response);
    } catch (error) {
      throw new Error(employeeListErrorMessage(error, fallbackMessage));
    }
  }

  async function savePusakaCredentials() {
    if (!selectedEmployee) return;
    saving = true;
    success = '';
    try {
      const res = await fetch(`/api/pusaka/employees/${selectedEmployee.id}`, {
        method: 'PUT',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify({ pusaka_username: pusakaUsername, pusaka_password: pusakaPassword }),
      });
      await ensureMutationOk(res, 'Gagal menyimpan kredensial PUSAKA');
      toast.success('Kredensial PUSAKA berhasil diperbarui.');
      success = `Kredensial PUSAKA untuk ${selectedEmployee.nama} berhasil diperbarui.`;
      showPusakaDialog = false;
      await ondelete?.();
    } catch (error) {
      showOperationError(
        'Kredensial Gagal Diperbarui',
        employeeListErrorMessage(error, 'Gagal menyimpan kredensial PUSAKA')
      );
    } finally { saving = false; }
  }

  async function testPusakaCredentials(emp: Employee) {
    testingId = emp.id;
    try {
      const res  = await fetch(`/api/pusaka/employees/${emp.id}/test-pusaka`, { method: 'POST' });
      const data = await readClientApiData<{ message?: string }>(res, 'Test gagal');
      const message = isRecord(data) && typeof data.message === 'string' ? data.message : '';
      toast.success(message || 'Test selesai');
    } catch (error) {
      toast.error(employeeListErrorMessage(error, 'Test gagal'));
    } finally { testingId = null; }
  }

  async function togglePusakaAccount(emp: Employee, isEnabled: boolean) {
    const challenge = isEnabled ? 'AKTIFKAN' : 'NONAKTIFKAN';
    if (!(await confirmPhrase(
      isEnabled ? 'Aktifkan Akun PUSAKA' : 'Nonaktifkan Akun PUSAKA',
      isEnabled
        ? `Akun PUSAKA ${emp.nama} akan diaktifkan kembali dan bisa dipakai untuk job otomatis maupun manual.`
        : `Akun PUSAKA ${emp.nama} akan dinonaktifkan dari integrasi. Job otomatis sebaiknya tidak lagi dijalankan sampai akun diaktifkan kembali.`,
      challenge,
    ))) return;
    accountTogglingId = emp.id;
    success = '';
    try {
      const res = await fetch(`/api/pusaka/employees/${emp.id}`, {
        method: 'PATCH',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify({ is_enabled: isEnabled }),
      });
      await ensureMutationOk(res, 'Gagal memperbarui status akun PUSAKA');
      toast.success(isEnabled ? 'Akun PUSAKA diaktifkan kembali.' : 'Akun PUSAKA dinonaktifkan.');
      success = `Akun PUSAKA ${emp.nama} berhasil ${isEnabled ? 'diaktifkan kembali' : 'dinonaktifkan'}.`;
      await ondelete?.();
    } catch (error) {
      showOperationError(
        'Status Akun Gagal Diperbarui',
        employeeListErrorMessage(error, 'Gagal memperbarui status akun PUSAKA')
      );
    } finally {
      accountTogglingId = null;
    }
  }

  async function deletePusakaAccount(emp: Employee) {
    if (!(await confirmPhrase('Hapus Akun PUSAKA', `Akun PUSAKA untuk ${emp.nama} akan dilepas dari integrasi. Jadwal tetap tersimpan, tetapi kredensial dan status akun integrasi akan hilang dari pegawai ini.`, 'HAPUS'))) return;
    accountDeletingId = emp.id;
    success = '';
    try {
      const res = await fetch(`/api/pusaka/employees/${emp.id}`, { method: 'DELETE' });
      await ensureMutationOk(res, 'Gagal menghapus akun PUSAKA');
      toast.success('Akun PUSAKA berhasil dihapus.');
      success = `Akun PUSAKA ${emp.nama} berhasil dihapus dari integrasi.`;
      await ondelete?.();
    } catch (error) {
      showOperationError(
        'Akun Gagal Dihapus',
        employeeListErrorMessage(error, 'Gagal menghapus akun PUSAKA')
      );
    } finally {
      accountDeletingId = null;
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

  function isRecord(value: unknown): value is Record<string, unknown> {
    return typeof value === 'object' && value !== null;
  }

  function normalizeRows<T>(value: T[] | null | undefined): T[] {
    return Array.isArray(value) ? value : [];
  }

  function employeeListErrorMessage(error: unknown, fallbackMessage: string) {
    if (error instanceof Error && error.message.trim()) return error.message;
    if (typeof error === 'string' && error.trim()) return error;
    return fallbackMessage;
  }

  async function fetchAuditLogs(emp: Employee): Promise<AuditLog[]> {
    const logs = await fetch(`/api/pusaka/employees/${emp.id}/audit-logs?per_page=10`).then((response) =>
      readClientApiData<AuditLog[]>(response, 'Gagal memuat riwayat akun PUSAKA')
    );
    return normalizeRows(logs);
  }

  async function openAuditDialog(emp: Employee) {
    auditEmployee = emp;
    auditLogs = [];
    showAuditDialog = true;
    auditPromise = fetchAuditLogs(emp).then((logs) => {
      auditLogs = logs;
      return logs;
    });
  }

  function retryAudit(reset?: () => void) {
    if (!auditEmployee) return;
    reset?.();
    auditPromise = fetchAuditLogs(auditEmployee).then((logs) => {
      auditLogs = logs;
      return logs;
    });
  }

  function handleAuditRenderError(error: unknown, reset: () => void) {
    console.error('PUSAKA audit dialog render failed', error);
    reset();
  }

  async function doStop(id: string) {
    if (!(await confirmAction({
      title: 'Batalkan Job Aktif',
      message: 'Stop akan mencoba membatalkan job aktif untuk pegawai ini. Lanjutkan?',
      confirmLabel: 'Batalkan Job',
      tone: 'warning'
    }))) return;
    busyId = id;
    success = '';
    try {
      const res  = await fetch('/api/pusaka/jobs/cancel', {
        method: 'POST', headers: { 'content-type': 'application/json' },
        body: JSON.stringify({ employee_id: id }),
      });
      const data = await readClientApiData<{ cancelled?: number }>(res, 'Gagal membatalkan job aktif pegawai');
      const cancelled = data.cancelled ?? 0;
      success = cancelled > 0
        ? `${cancelled} job untuk pegawai ini berhasil dibatalkan.`
        : 'Tidak ada job aktif yang perlu dibatalkan untuk pegawai ini.';
      await onstop?.(id, cancelled);
    } catch (error) {
      showOperationError(
        'Job Gagal Dibatalkan',
        employeeListErrorMessage(error, 'Gagal membatalkan job aktif pegawai')
      );
    } finally {
      busyId = null;
    }
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

  async function fetchScheduleConfigs(emp: Employee): Promise<DayConfig[]> {
    const schedules = await fetch(`/api/pusaka/employees/${emp.id}/schedules`).then((response) =>
      readClientApiData<EmployeeSchedule[]>(response, 'Gagal memuat jadwal absensi pegawai')
    );
    return populateDayConfigs(normalizeRows(schedules));
  }

  async function openScheduleDialog(emp: Employee) {
    scheduleEmployee   = emp;
    showScheduleDialog = true;
    dayConfigs         = Array.from({ length: 7 }, makeDayConfig);
    schedulePromise = fetchScheduleConfigs(emp).then((configs) => {
      dayConfigs = configs;
      return configs;
    });
  }

  function retrySchedule(reset?: () => void) {
    if (!scheduleEmployee) return;
    reset?.();
    schedulePromise = fetchScheduleConfigs(scheduleEmployee).then((configs) => {
      dayConfigs = configs;
      return configs;
    });
  }

  function handleScheduleRenderError(error: unknown, reset: () => void) {
    console.error('PUSAKA schedule dialog render failed', error);
    reset();
  }

  async function saveDayRow(dow: number) {
    if (!scheduleEmployee) return;
    const cfg = dayConfigs[dow];
    scheduleSaving = true;
    success = '';
    try {
      if (cfg.checkinTime) {
        const res = await fetch(`/api/pusaka/employees/${scheduleEmployee.id}/schedules`, {
          method: 'POST', headers: { 'content-type': 'application/json' },
          body: JSON.stringify({
            run_type: 'checkin', run_time: cfg.checkinTime,
            is_enabled: cfg.checkinEnabled,
            random_window_minutes: cfg.randomWindow,
            day_of_week: dow,
          }),
        });
        await ensureMutationOk(res, 'Gagal menyimpan jadwal masuk pegawai');
      }
      if (cfg.checkoutTime) {
        const res = await fetch(`/api/pusaka/employees/${scheduleEmployee.id}/schedules`, {
          method: 'POST', headers: { 'content-type': 'application/json' },
          body: JSON.stringify({
            run_type: 'checkout', run_time: cfg.checkoutTime,
            is_enabled: cfg.checkoutEnabled,
            random_window_minutes: cfg.randomWindow,
            day_of_week: dow,
          }),
        });
        await ensureMutationOk(res, 'Gagal menyimpan jadwal pulang pegawai');
      }
      const configs = await fetchScheduleConfigs(scheduleEmployee);
      dayConfigs = configs;
      schedulePromise = Promise.resolve(configs);
      success = `Jadwal ${scheduleEmployee.nama} untuk ${dayLabels[dow]} berhasil diperbarui.`;
      await ondelete?.();
    } catch (error) {
      toast.error(employeeListErrorMessage(error, 'Gagal menyimpan jadwal absensi pegawai'));
    }
    finally { scheduleSaving = false; }
  }

  async function deleteDaySchedule(dow: number, runType: 'checkin' | 'checkout') {
    if (!scheduleEmployee) return;
    const cfg    = dayConfigs[dow];
    const schedId = runType === 'checkin' ? cfg.checkinId : cfg.checkoutId;
    if (!schedId) return;
    scheduleDeletingKey = `${dow}:${runType}`;
    try {
      const res = await fetch(`/api/pusaka/employees/${scheduleEmployee.id}/schedules/${schedId}`, { method: 'DELETE' });
      await ensureMutationOk(res, 'Gagal menghapus jadwal absensi pegawai');
      if (runType === 'checkin') { cfg.checkinId = null; cfg.checkinTime = ''; }
      else                       { cfg.checkoutId = null; cfg.checkoutTime = ''; }
      schedulePromise = Promise.resolve(dayConfigs);
      await ondelete?.();
    } catch (error) {
      toast.error(employeeListErrorMessage(error, 'Gagal menghapus jadwal absensi pegawai'));
    } finally {
      scheduleDeletingKey = null;
    }
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
    if (ci && co)  return 'border-success text-success bg-success/10 hover:bg-success/15';
    if (ci)        return 'border-warning text-warning bg-warning/10 hover:bg-warning/15';
    if (co)        return 'border-cyan-400 text-accent-foreground bg-accent/60 hover:bg-accent';
    return 'border-border text-muted-foreground hover:bg-muted/50';
  }

  function scheduleButtonLabel(emp: Employee): string {
    const ci = emp.has_checkin_schedule;
    const co = emp.has_checkout_schedule;
    if (ci && co) return 'Jadwal ✓';
    if (ci)       return 'Masuk ✓';
    if (co)       return 'Pulang ✓';
    return 'Jadwal';
  }

  const dayLabels = ['Minggu', 'Senin', 'Selasa', 'Rabu', 'Kamis', 'Jumat', 'Sabtu'];
  const canConfirmRun = $derived(runConfirmInput.trim() === 'SURE');
</script>

<div class="flex flex-col gap-4">
  <!-- Stat Cards -->
  <div class="grid grid-cols-3 gap-3">
    <div class="rounded-xl border border-base-300 bg-base-100 px-3 py-2.5 flex flex-col gap-0.5">
      <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground">Eligible</span>
      <span class="text-xl font-black text-foreground">{employees.length}</span>
      <span class="text-[10px] font-medium text-muted-foreground">pegawai PNS/PPPK</span>
    </div>
    <div class="rounded-xl border border-base-300 bg-base-100 px-3 py-2.5 flex flex-col gap-0.5">
      <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground">Aktif</span>
      <span class="text-xl font-black text-success">{employees.filter(e => e.pusaka_username && e.pusaka_is_enabled !== false).length}</span>
      <span class="text-[10px] font-medium text-muted-foreground">akun PUSAKA aktif</span>
    </div>
    <div class="rounded-xl border border-base-300 bg-base-100 px-3 py-2.5 flex flex-col gap-0.5">
      <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground">Butuh Setup</span>
      <span class="text-xl font-black text-warning">{employees.filter(e => !e.pusaka_username).length}</span>
      <span class="text-[10px] font-medium text-muted-foreground">pegawai belum setup</span>
    </div>
  </div>

  <!-- Filter Bar -->
  <div class="grid gap-3 md:grid-cols-[1.2fr_0.8fr_auto]">
    <div>
      <p class="mb-1 text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">Cari Pegawai</p>
      <Input placeholder="Cari nama / ID / NIP..." bind:value={search} class="w-full border border-input bg-background px-3 h-10" oninput={() => { page = 1; }} />
    </div>
    <div>
      <p class="mb-1 text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">Status Integrasi</p>
      <select bind:value={filterMode} class="select select-bordered w-full h-10 border border-input bg-background px-3" onchange={() => { page = 1; }}>
        <option value="all">Semua</option>
        <option value="configured">Akun aktif</option>
        <option value="needs_setup">Belum setup</option>
        <option value="disabled">Dinonaktifkan</option>
      </select>
    </div>
    <div class="flex items-end">
      <Badge variant="secondary" class="h-10 px-4 flex items-center text-sm">{filteredEmployees.length} pegawai</Badge>
    </div>
  </div>

  <!-- Success / Error Panels -->
  {#if success}
    <div>
      <SuccessPanel title="Operasi PUSAKA Berhasil" message={success} compact />
    </div>
  {/if}
  {#if operationState}
    <div>
      <OperationStatusPanel {...operationState} compact />
    </div>
  {/if}

  <!-- Card Grid -->
  {#if filteredEmployees.length > 0}
    <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-3">
      {#each pagedEmployees as e (e.id)}
        {@const si = statusInfo(e)}
        <div class="rounded-xl border border-base-300 bg-base-100 p-3 flex flex-col gap-2.5 transition-shadow hover:shadow-sm">
          <!-- Header: Name + Badges -->
          <div class="flex items-start justify-between gap-2">
            <div class="min-w-0 flex-1">
              <p class="text-sm font-bold text-foreground truncate">{e.nama}</p>
              <p class="text-[11px] text-muted-foreground truncate">NIP {e.nip || '—'}</p>
            </div>
            <div class="flex gap-1 shrink-0">
              {#if e.is_active}
                <Badge variant="outline" class="text-[10px] leading-none py-0.5 px-1.5 border-primary/20 text-primary">Aktif</Badge>
              {:else}
                <Badge variant="secondary" class="text-[10px] leading-none py-0.5 px-1.5">Nonaktif</Badge>
              {/if}
            </div>
          </div>

          <!-- Meta: Unit + Pusaka + Status -->
          <div class="flex flex-wrap gap-x-3 gap-y-1 text-[11px] text-muted-foreground">
            {#if e.unit_kerja}
              <span class="truncate">{e.unit_kerja}</span>
            {/if}
            <span>Pusaka:
              {#if isPusakaConfigured(e)}
                <span class="font-semibold {e.pusaka_is_enabled === false ? 'text-warning' : 'text-success'}">{pusakaStatusLabel(e)}</span>
              {:else}
                <span class="text-warning">Belum</span>
              {/if}
            </span>
            {#if si}
              <span>Status: <span class="font-semibold {si.status === 'success' ? 'text-success' : si.status === 'failed' ? 'text-destructive' : 'text-warning'}">{statusLabel(si.status)}{si.tipe ? ' · ' + si.tipe : ''}</span></span>
            {/if}
          </div>

          <!-- Action Buttons -->
          <div class="flex items-center gap-1.5 pt-1">
            <button onclick={() => openRunConfirm(e, 'morning')} disabled={busyId === e.id}
              class="inline-flex items-center justify-center rounded-lg border border-input bg-background px-2.5 h-8 text-xs font-medium text-foreground hover:bg-accent transition-colors disabled:opacity-50">
              Rekap
            </button>
            <button onclick={() => openRunConfirm(e, 'checkin')} disabled={busyId === e.id}
              class="inline-flex items-center justify-center rounded-lg border border-warning/30 px-2.5 h-8 text-xs font-medium text-warning hover:bg-warning/10 transition-colors disabled:opacity-50">
              <svg class="h-3 w-3 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
              Masuk
            </button>
            <button onclick={() => openRunConfirm(e, 'checkout')} disabled={busyId === e.id}
              class="inline-flex items-center justify-center rounded-lg border border-warning/30 px-2.5 h-8 text-xs font-medium text-warning hover:bg-warning/10 transition-colors disabled:opacity-50">
              <svg class="h-3 w-3 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" /></svg>
              Pulang
            </button>

            <!-- More dropdown -->
            <div class="relative dropdown ml-auto">
              <button
                class="inline-flex items-center justify-center rounded-lg border border-input bg-background px-2 h-8 text-xs font-medium text-muted-foreground hover:bg-accent transition-colors"
                onclick={(ev) => {
                  const btn = ev.currentTarget;
                  const menu = btn.nextElementSibling as HTMLElement;
                  if (menu) menu.classList.toggle('hidden');
                }}
                aria-label="Aksi lainnya"
              >
                <svg class="h-3.5 w-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                  <circle cx="12" cy="5" r="1.5" fill="currentColor" stroke="none" />
                  <circle cx="12" cy="12" r="1.5" fill="currentColor" stroke="none" />
                  <circle cx="12" cy="19" r="1.5" fill="currentColor" stroke="none" />
                </svg>
              </button>
              <div class="hidden absolute right-0 z-50 mt-1 min-w-[180px] rounded-lg border border-border bg-card p-1 shadow-xl" onclick={(ev) => ev.stopPropagation()}>
                <Button size="sm" variant="ghost" class="w-full justify-start gap-2 rounded-md px-3 py-1.5 text-xs font-medium" onclick={() => openPusakaDialog(e)}>
                  <svg class="h-3.5 w-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.066 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.573 1.066c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.066-2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                    <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                  </svg>
                  {isPusakaConfigured(e) ? 'Edit Akun' : 'Setup Akun'} Pusaka
                </Button>
                {#if isPusakaConfigured(e)}
                  <LoadingButton size="sm" variant="ghost" class="w-full justify-start gap-2 rounded-md px-3 py-1.5 text-xs font-medium"
                    onclick={() => togglePusakaAccount(e, e.pusaka_is_enabled === false)}
                    loading={accountTogglingId === e.id}
                    loadingLabel="Memproses..."
                    disabled={(accountTogglingId !== null && accountTogglingId !== e.id) || accountDeletingId !== null}
                  >
                    <svg class="h-3.5 w-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                    </svg>
                    {e.pusaka_is_enabled === false ? 'Aktifkan Akun' : 'Nonaktifkan Akun'}
                  </LoadingButton>
                {/if}
                <hr class="my-1 border-border" />
                <Button size="sm" variant="ghost" class="w-full justify-start gap-2 rounded-md px-3 py-1.5 text-xs font-medium" onclick={() => openScheduleDialog(e)}>
                  <svg class="h-3.5 w-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                  </svg>
                  Atur Jadwal
                </Button>
                <Button size="sm" variant="ghost" class="w-full justify-start gap-2 rounded-md px-3 py-1.5 text-xs font-medium" onclick={() => openAuditDialog(e)}>
                  <svg class="h-3.5 w-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                  Riwayat Audit
                </Button>
                <Button size="sm" variant="ghost" class="w-full justify-start gap-2 rounded-md px-3 py-1.5 text-xs font-medium" onclick={() => testPusakaCredentials(e)} disabled={!isPusakaConfigured(e)}>
                  <svg class="h-3.5 w-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                  </svg>
                  Test Koneksi
                </Button>
                {#if isPusakaConfigured(e)}
                  <hr class="my-1 border-border" />
                  <Button size="sm" variant="ghost" class="w-full justify-start gap-2 rounded-md px-3 py-1.5 text-xs font-medium text-destructive hover:text-destructive" onclick={() => deletePusakaAccount(e)}>
                    <svg class="h-3.5 w-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                    </svg>
                    Hapus Akun
                  </Button>
                {/if}
                <hr class="my-1 border-border" />
                <LoadingButton size="sm" variant="ghost" class="w-full justify-start gap-2 rounded-md px-3 py-1.5 text-xs font-medium text-warning hover:text-warning" onclick={() => doStop(e.id)} loading={busyId === e.id} loadingLabel="Memproses..." disabled={busyId === e.id || !e.active_status}>
                  <svg class="h-3.5 w-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2"><rect x="6" y="6" width="12" height="12" rx="2" /></svg>
                  Stop Job
                </LoadingButton>
              </div>
            </div>
          </div>
        </div>
      {/each}
    </div>

    <Pagination bind:page total={totalFiltered} perPage={PER_PAGE} />
  {:else}
    <div class="rounded-xl border border-dashed border-base-300 bg-base-100/50 px-6 py-12 text-center">
      <EmptyStatePanel
        compact
        title="Tidak ada pegawai yang cocok"
        description="Coba ubah filter atau kata kunci pencarian."
      />
    </div>
  {/if}
</div>

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
        <div class="rounded-md border border-warning/30 bg-warning/10 px-4 py-3 text-sm text-warning">
          Aksi ini akan langsung menjalankan job ke Pusaka Kemenag dan <strong>tidak bisa dibatalkan</strong> setelah dieksekusi.
          Pastikan waktu dan pegawai sudah benar.
        </div>
        <div class="space-y-1.5">
          <label for="run-confirm-input" class="text-sm font-medium">
            Ketik <code class="rounded bg-muted px-1.5 py-0.5 font-mono text-xs font-bold">SURE</code> untuk melanjutkan
          </label>
          <Input
            id="run-confirm-input"
            bind:value={runConfirmInput}
            placeholder="Ketik SURE"
            class="font-mono uppercase"
            onkeydown={(e: KeyboardEvent) => { if (e.key === 'Enter' && canConfirmRun) submitRunConfirm(); }}
          />
        </div>
      </div>

      <Dialog.Footer>
        <Button variant="outline" onclick={closeRunConfirm}>Batal</Button>
        <Button
          onclick={submitRunConfirm}
          disabled={!canConfirmRun}
          class={canConfirmRun ? 'bg-warning hover:bg-warning text-background border-transparent' : ''}>
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
        <PasswordInput id="pusaka-password" bind:value={pusakaPassword} placeholder="Kosongkan jika tidak diubah" />
      </div>
    </div>
    <Dialog.Footer>
      <Button variant="outline" onclick={() => (showPusakaDialog = false)}>Batal</Button>
      <LoadingButton onclick={() => void savePusakaCredentials()} loading={saving} loadingLabel="Menyimpan..." disabled={saving}>Simpan</LoadingButton>
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
      <AsyncContent promise={auditPromise} onerror={handleAuditRenderError}>
        {#snippet pending()}
          <div class="space-y-3 rounded-md border border-border bg-muted/50 px-3 py-4">
            {#each Array.from({ length: 3 }) as _, index (`audit-skeleton-${index}`)}
              <div class="rounded-xl border border-border bg-card px-3 py-3">
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
        {/snippet}
        {#snippet failed(error, reset)}
          <RecoveryPanel
            compact
            title="Riwayat PUSAKA Belum Tersaji"
            message={employeeListErrorMessage(error, 'Gagal memuat riwayat akun PUSAKA')}
            onRetry={() => retryAudit(reset)}
          />
        {/snippet}
        {#snippet children(_logs)}
          {#if auditLogs.length === 0}
            <div class="rounded-md border border-border bg-muted/50 px-3 py-6 text-center text-sm text-muted-foreground">Belum ada riwayat akun PUSAKA untuk pegawai ini.</div>
          {:else}
            <div class="space-y-2">
              {#each auditLogs as log (log.id)}
                {@const meta = auditMeta(log)}
                <div class="rounded-xl border border-border bg-card px-3 py-3">
                  <div class="flex items-start justify-between gap-3">
                    <div>
                      <p class="text-sm font-semibold text-foreground">{auditActionLabel(log.action)}</p>
                      <p class="mt-1 text-xs text-muted-foreground">{formatAuditDate(log.created_at)}</p>
                    </div>
                    <Badge variant="outline" class="text-[11px]">{log.username ?? 'Sistem'}</Badge>
                  </div>
                  {#if meta.pusaka_username}
                    <p class="mt-2 text-xs text-muted-foreground">Username: <span class="font-mono">{String(meta.pusaka_username)}</span></p>
                  {/if}
                  {#if typeof meta.is_enabled === 'boolean'}
                    <p class="mt-1 text-xs text-muted-foreground">Status akun: {meta.is_enabled ? 'aktif' : 'dinonaktifkan'}</p>
                  {/if}
                </div>
              {/each}
            </div>
          {/if}
        {/snippet}
      </AsyncContent>
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

    <AsyncContent promise={schedulePromise} onerror={handleScheduleRenderError}>
      {#snippet pending()}
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
      {/snippet}
      {#snippet failed(error, reset)}
        <div class="py-2">
          <RecoveryPanel
            compact
            title="Jadwal Absensi Belum Tersaji"
            message={employeeListErrorMessage(error, 'Gagal memuat jadwal absensi pegawai')}
            onRetry={() => retrySchedule(reset)}
          />
        </div>
      {/snippet}
      {#snippet children(_configs)}
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
                    <LoadingButton
                      size="xs"
                      variant="ghost"
                      class="text-xs text-destructive hover:text-destructive/80 shrink-0"
                      onclick={() => deleteDaySchedule(dow, 'checkin')}
                      loading={scheduleDeletingKey === `${dow}:checkin`}
                      loadingLabel="..."
                      disabled={scheduleDeletingKey !== null && scheduleDeletingKey !== `${dow}:checkin`}
                      title="Hapus jadwal masuk"
                    >✕</LoadingButton>
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
                    <LoadingButton
                      size="xs"
                      variant="ghost"
                      class="text-xs text-destructive hover:text-destructive/80 shrink-0"
                      onclick={() => deleteDaySchedule(dow, 'checkout')}
                      loading={scheduleDeletingKey === `${dow}:checkout`}
                      loadingLabel="..."
                      disabled={scheduleDeletingKey !== null && scheduleDeletingKey !== `${dow}:checkout`}
                      title="Hapus jadwal pulang"
                    >✕</LoadingButton>
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
      {/snippet}
    </AsyncContent>

    <Dialog.Footer>
      <Button variant="outline" onclick={() => (showScheduleDialog = false)}>Tutup</Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
