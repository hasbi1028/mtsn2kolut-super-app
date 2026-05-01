<script lang="ts">
  import { onMount } from 'svelte';
  import EmployeeForm from '$lib/components/EmployeeForm.svelte';
  import GeneralEmployeeList from '$lib/components/GeneralEmployeeList.svelte';
  import AsyncContent from '$lib/components/AsyncContent.svelte';
  import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
  import { Skeleton } from '$lib/components/ui/skeleton';
  import { toast } from '$lib/components/ui/sonner';
  import { readClientApiData } from '$lib/client/api';

  type Employee = {
    id: string;
    nip: string;
    nama: string;
    unit_kerja: string;
    employment_type: string;
    pusaka_eligible: boolean;
    has_pusaka_account: boolean;
    pusaka_is_enabled: boolean;
    is_active: boolean;
  };
  type EmployeesPayload = {
    items?: Employee[];
    data?: {
      items?: Employee[];
    };
    error?: string;
    message?: string;
  };

  let employees = $state<Employee[]>([]);
  let employeesPromise = $state<Promise<Employee[]> | null>(null);
  let employeesRequestId = 0;

  function employeeRows(payload: EmployeesPayload | Employee[]) {
    if (Array.isArray(payload)) return payload;
    return payload.items ?? payload.data?.items ?? [];
  }

  async function fetchEmployees() {
    const res = await fetch('/api/employees');
    const payload = await readClientApiData<EmployeesPayload | Employee[]>(res, 'Gagal memuat data pegawai');
    return employeeRows(payload);
  }

  function load() {
    const requestId = ++employeesRequestId;
    employees = [];
    employeesPromise = fetchEmployees()
      .then((rows) => {
        if (requestId === employeesRequestId) {
          employees = rows;
          return rows;
        }
        return employees;
      })
      .catch((error: unknown) => {
        if (requestId === employeesRequestId) throw error;
        return employees;
      });
  }

  async function refreshEmployees() {
    if (!employeesPromise) {
      load();
      return;
    }
    try {
      const requestId = ++employeesRequestId;
      const rows = await fetchEmployees();
      if (requestId === employeesRequestId) {
        employees = rows;
        employeesPromise = Promise.resolve(rows);
      }
    } catch (error) {
      employeesPromise = Promise.resolve(employees);
      toast.error(employeeErrorMessage(error));
    }
  }

  function retryEmployees(reset?: () => void) {
    reset?.();
    load();
  }

  function employeeErrorMessage(error: unknown) {
    if (error instanceof Error && error.message.trim()) return error.message;
    return 'Gagal memuat data pegawai. Coba lagi untuk mengambil master pegawai terbaru.';
  }

  function handleEmployeeRenderError(error: unknown) {
    console.error('Employees render failed', error);
  }

  onMount(() => {
    void load();
  });
</script>

<svelte:head><title>Pegawai — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6">
  <div>
    <h1 class="text-2xl font-semibold text-slate-800">Master Pegawai</h1>
    <p class="mt-1 text-sm text-muted-foreground">Data seluruh pegawai sekolah. Integrasi akun, jadwal, dan job PUSAKA dikelola terpisah dari area ini.</p>
  </div>

  <AsyncContent promise={employeesPromise} onerror={handleEmployeeRenderError}>
    {#snippet pending()}
      <div class="grid gap-3 md:grid-cols-3">
        {#each Array.from({ length: 3 }) as _, index (`employee-stat-skeleton-${index}`)}
          <div class="rounded-2xl border border-slate-200 bg-white px-4 py-4">
            <Skeleton class="h-3 w-28" />
            <Skeleton class="mt-3 h-8 w-16" />
            <Skeleton class="mt-2 h-4 w-44" />
          </div>
        {/each}
      </div>
      <div class="rounded-2xl border border-slate-200 bg-white p-5">
        <Skeleton class="h-6 w-40" />
        <Skeleton class="mt-4 h-24 w-full" />
      </div>
    {/snippet}

    {#snippet failed(error, reset)}
      <RecoveryPanel
        title="Data Pegawai Belum Tersaji"
        message={employeeErrorMessage(error)}
        onRetry={() => retryEmployees(reset)}
      />
    {/snippet}

    {#snippet children(value)}
      {@const currentEmployees = value as Employee[]}
      <div class="grid gap-3 md:grid-cols-3">
        <div class="rounded-2xl border border-emerald-100 bg-emerald-50 px-4 py-4">
          <p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-emerald-700">Total Pegawai</p>
          <p class="mt-2 text-2xl font-semibold text-slate-900">{currentEmployees.length}</p>
          <p class="text-sm text-slate-600">seluruh profil pegawai yang tercatat</p>
        </div>
        <div class="rounded-2xl border border-sky-100 bg-sky-50 px-4 py-4">
          <p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-sky-700">Pegawai Aktif</p>
          <p class="mt-2 text-2xl font-semibold text-slate-900">{currentEmployees.filter((item) => item.is_active).length}</p>
          <p class="text-sm text-slate-600">siap dipakai untuk akun, akademik, dan operasional</p>
        </div>
        <div class="rounded-2xl border border-amber-100 bg-amber-50 px-4 py-4">
          <p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-amber-700">Eligible PUSAKA</p>
          <p class="mt-2 text-2xl font-semibold text-slate-900">
            {currentEmployees.filter((item) => item.employment_type === 'pns' || item.employment_type === 'pppk').length}
          </p>
          <p class="text-sm text-slate-600">subset yang dapat dikelola di area PUSAKA</p>
        </div>
      </div>

      <EmployeeForm onadd={refreshEmployees} />
      <GeneralEmployeeList employees={currentEmployees} onreload={refreshEmployees} />
    {/snippet}
  </AsyncContent>
</div>
