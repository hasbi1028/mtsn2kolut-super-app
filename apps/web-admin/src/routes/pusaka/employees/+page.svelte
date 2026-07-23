<script lang="ts">
  import { onMount } from 'svelte';
  import AsyncContent from '$lib/components/AsyncContent.svelte';
  import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
  import { Skeleton } from '$lib/components/ui/skeleton';
  import { toast } from '$lib/components/ui/sonner';
  import EmployeeList from '$lib/components/EmployeeList.svelte';
  import { readClientApiData } from '$lib/client/api';

  type Employee = {
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
  };

  type EmployeeListPayload = {
    items?: Employee[];
    data?: Employee[];
    error?: string;
    message?: string;
  };

  let employeesPromise = $state<Promise<Employee[]> | null>(null);
  let employees = $state<Employee[]>([]);
  let employeesRequestId = 0;

  function normalizeEmployees(payload: EmployeeListPayload | Employee[] | null | undefined): Employee[] {
    if (Array.isArray(payload)) return payload;
    if (!payload) return [];
    if (Array.isArray(payload.items)) return payload.items;
    if (Array.isArray(payload.data)) return payload.data;
    return [];
  }

  async function fetchEmployees(): Promise<Employee[]> {
    const payload = await fetch('/api/pusaka/employees').then((response) =>
      readClientApiData<EmployeeListPayload | Employee[]>(response, 'Gagal memuat pegawai PUSAKA')
    );
    return normalizeEmployees(payload);
  }

  function loadEmployees() {
    const requestId = ++employeesRequestId;
    employeesPromise = fetchEmployees()
      .then((nextEmployees) => {
        if (requestId === employeesRequestId) {
          employees = nextEmployees;
          return nextEmployees;
        }
        return employees;
      })
      .catch((error: unknown) => {
        if (requestId === employeesRequestId) throw error;
        return employees;
      });
    return employeesPromise;
  }

  async function refreshEmployees(showFailureToast = false) {
    if (!employeesPromise) {
      await loadEmployees();
      return;
    }
    const requestId = ++employeesRequestId;
    try {
      const nextEmployees = await fetchEmployees();
      if (requestId === employeesRequestId) {
        employees = nextEmployees;
        employeesPromise = Promise.resolve(nextEmployees);
      }
    } catch (error) {
      if (requestId !== employeesRequestId) return;
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
    return 'Pegawai PUSAKA belum dapat dimuat. Periksa koneksi layanan sistem lalu coba lagi.';
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
      await readClientApiData<unknown>(res, 'Gagal mengantrekan pekerjaan PUSAKA.');
    } catch (error) {
      toast.error(employeeErrorMessage(error));
      throw error;
    }
    const labels: Record<string, string> = { morning: 'rekap', afternoon: 'rekap', checkin: 'absensi masuk', checkout: 'absensi pulang' };
    toast.success(`Job ${labels[run_type] ?? run_type} berhasil di-queue.`);
    await refreshEmployees(true);
  }

  function handleStop(_id: string, cancelled: number) {
    toast.success(cancelled > 0 ? `${cancelled} job dibatalkan.` : 'Tidak ada pekerjaan aktif untuk pegawai ini.');
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
    <h1 class="text-2xl font-black text-base-content">Pegawai PUSAKA</h1>
    <p class="mt-1.5 text-sm text-base-content/70">Area khusus pegawai PNS dan PPPK yang mengikuti integrasi PUSAKA Kemenag.</p>
  </div>

  <AsyncContent promise={employeesPromise} onerror={handleEmployeeRenderError}>
    {#snippet pending()}
      <div class="rounded-2xl border border-base-300 bg-base-100 p-5">
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
