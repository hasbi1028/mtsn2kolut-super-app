<script lang="ts">
  import { onMount } from 'svelte';
  import EmployeeForm from '$lib/components/EmployeeForm.svelte';
  import EmployeeList from '$lib/components/EmployeeList.svelte';

  let employees = $state<any[]>([]);
  let toast     = $state('');

  async function load() {
    try {
      const res  = await fetch('/api/employees');
      const data = await res.json();
      if (data.error) { console.error('[employees]', data.error); return; }
      employees = data.items ?? [];
    } catch (e) {
      console.error('[employees] load failed:', e);
    }
  }

  async function runNow(employee_id: string, run_type: string) {
    try {
      const res  = await fetch('/api/jobs/run-now', {
        method: 'POST',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify({ employee_id, run_type }),
      });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) console.error('[run-now]', data.error);
    } catch (e) {
      console.error('[run-now] failed:', e);
    }
    const labels: Record<string, string> = { morning: 'rekap', afternoon: 'rekap', checkin: 'absensi masuk', checkout: 'absensi pulang' };
    showToast(`Job ${labels[run_type] ?? run_type} berhasil di-queue.`);
  }

  function handleStop(_id: string, cancelled: number) {
    showToast(cancelled > 0 ? `${cancelled} job dibatalkan.` : 'Tidak ada job aktif untuk pegawai ini.');
  }

  function showToast(msg: string) {
    toast = msg;
    setTimeout(() => (toast = ''), 3500);
  }

  onMount(() => {
    load();
    const itv = setInterval(load, 4000);
    return () => clearInterval(itv);
  });
</script>

<svelte:head><title>Pegawai — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6">
  <div>
    <h1 class="text-2xl font-semibold text-slate-800">Manajemen Pegawai</h1>
    <p class="text-sm text-muted-foreground mt-1">Data pegawai dan kontrol job absensi Pusaka Kemenag</p>
  </div>

  {#if toast}
    <div class="rounded-md border border-green-200 bg-green-50 px-4 py-3 text-sm text-green-800">
      {toast}
    </div>
  {/if}

  <EmployeeForm onadd={load} />
  <EmployeeList {employees} onrun={runNow} onstop={handleStop} ondelete={load} />
</div>
