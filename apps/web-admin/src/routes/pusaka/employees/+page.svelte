<script lang="ts">
  import { onMount } from 'svelte';
  import AsyncContent from '$lib/components/AsyncContent.svelte';
  import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
  import { Skeleton } from '$lib/components/ui/skeleton';
  import { toast } from '$lib/components/ui/sonner';
  import EmployeeList from '$lib/components/EmployeeList.svelte';

  type Employee = {
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
  };

  type ApiEnvelope<T> = {
    data?: T;
    error?: string;
    message?: string;
  };

  type EmployeeListPayload = {
    items?: Employee[];
    data?: Employee[];
    error?: string;
    message?: string;
  };

  let employeesPromise = $state<Promise<Employee[]> | null>(null);
  let employees = $state<Employee[]>([]);

  function isRecord(value: unknown): value is Record<string, unknown> {
    return typeof value === 'object' && value !== null;
  }

  function apiErrorMessage(payload: unknown) {
    if (!isRecord(payload)) return '';
    const error = payload.error;
    if (typeof error === 'string' && error.trim()) return error;
    const message = payload.message;
    if (typeof message === 'string' && message.trim()) return message;
    return '';
  }

  async function readApi<T>(response: Response, fallbackMessage: string): Promise<T> {
    const payload = (await response.json().catch(() => null)) as ApiEnvelope<T> | T | null;
    const message = apiErrorMessage(payload);
    if (!response.ok) throw new Error(message || fallbackMessage);
    if (isRecord(payload) && typeof payload.error === 'string' && payload.error.trim()) {
      throw new Error(payload.error);
    }
    if (isRecord(payload) && 'data' in payload) {
      const envelope = payload as ApiEnvelope<T>;
      if (envelope.data === undefined) throw new Error(fallbackMessage);
      return envelope.data;
    }
    if (payload === null) throw new Error(fallbackMessage);
    return payload as T;
  }

  function normalizeEmployees(payload: EmployeeListPayload | Employee[] | null | undefined): Employee[] {
    if (Array.isArray(payload)) return payload;
    if (!payload) return [];
    if (Array.isArray(payload.items)) return payload.items;
    if (Array.isArray(payload.data)) return payload.data;
    return [];
  }

  async function fetchEmployees(): Promise<Employee[]> {
    const payload = await fetch('/api/pusaka/employees').then((response) =>
      readApi<EmployeeListPayload | Employee[]>(response, 'Gagal memuat pegawai PUSAKA')
    );
    return normalizeEmployees(payload);
  }

  function loadEmployees() {
    employeesPromise = fetchEmployees().then((nextEmployees) => {
      employees = nextEmployees;
      return nextEmployees;
    });
    return employeesPromise;
  }

  async function refreshEmployees(showFailureToast = false) {
    if (!employeesPromise) {
      await loadEmployees();
      return;
    }
    try {
      const nextEmployees = await fetchEmployees();
      employees = nextEmployees;
      employeesPromise = Promise.resolve(nextEmployees);
    } catch (error) {
      employeesPromise = Promise.resolve(employees);
      if (showFailureToast) toast.error(employeeErrorMessage(error));
    }
  }

  function retryEmployees(reset?: () => void) {
    reset?.();
    loadEmployees();
  }

  function employeeErrorMessage(error: unknown) {
    if (error instanceof Error && error.message.trim()) return error.message;
    if (typeof error === 'string' && error.trim()) return error;
    return 'Pegawai PUSAKA belum dapat dimuat. Periksa koneksi backend lalu coba lagi.';
  }

  function handleEmployeeRenderError(error: unknown, reset: () => void) {
    console.error('PUSAKA employees render failed', error);
    reset();
  }

  async function runNow(employee_id: string, run_type: string) {
    try {
      const res  = await fetch('/api/pusaka/jobs/run-now', {
        method: 'POST',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify({ employee_id, run_type }),
      });
      const data = await res.json().catch(() => null);
      if (!res.ok) {
        toast.error(apiErrorMessage(data) || 'Gagal mengantrekan job PUSAKA.');
        return;
      }
    } catch (error) {
      toast.error(employeeErrorMessage(error));
      return;
    }
    const labels: Record<string, string> = { morning: 'rekap', afternoon: 'rekap', checkin: 'absensi masuk', checkout: 'absensi pulang' };
    toast.success(`Job ${labels[run_type] ?? run_type} berhasil di-queue.`);
    await refreshEmployees(true);
  }

  function handleStop(_id: string, cancelled: number) {
    toast.success(cancelled > 0 ? `${cancelled} job dibatalkan.` : 'Tidak ada job aktif untuk pegawai ini.');
    void refreshEmployees(true);
  }

  onMount(() => {
    void loadEmployees();
    const itv = setInterval(() => void refreshEmployees(false), 4000);
    return () => clearInterval(itv);
  });
</script>

<svelte:head><title>Pegawai PUSAKA — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6">
  <div>
    <h1 class="text-2xl font-semibold text-slate-800">Pegawai PUSAKA</h1>
    <p class="mt-1 text-sm text-muted-foreground">Area khusus pegawai PNS dan PPPK yang mengikuti integrasi PUSAKA Kemenag.</p>
  </div>

  <AsyncContent promise={employeesPromise} onerror={handleEmployeeRenderError}>
    {#snippet pending()}
      <div class="rounded-2xl border border-slate-200 bg-white p-5">
        <div class="mb-5 grid gap-3 md:grid-cols-[1fr_1fr_auto]">
          <Skeleton class="h-10 w-full" />
          <Skeleton class="h-10 w-full" />
          <Skeleton class="h-10 w-28" />
        </div>
        <div class="space-y-3">
          {#each Array.from({ length: 6 }) as _, index (`pusaka-employee-skeleton-${index}`)}
            <Skeleton class="h-14 w-full" />
          {/each}
        </div>
      </div>
    {/snippet}
    {#snippet failed(error, reset)}
      <RecoveryPanel title="Pegawai PUSAKA Belum Tersaji" message={employeeErrorMessage(error)} onRetry={() => retryEmployees(reset)} />
    {/snippet}
    {#snippet children(_employees)}
      <EmployeeList {employees} onrun={runNow} onstop={handleStop} ondelete={() => void refreshEmployees(true)} />
    {/snippet}
  </AsyncContent>
</div>
