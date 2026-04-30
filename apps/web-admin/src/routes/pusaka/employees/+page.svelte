<script lang="ts">
  import { onMount } from 'svelte';
  import { toast } from '$lib/components/ui/sonner';
  import EmployeeList from '$lib/components/EmployeeList.svelte';

  let employees = $state<any[]>([]);

  async function load() {
    try {
      const res = await fetch('/api/pusaka/employees');
      const data = await res.json();
      if (data.error) {
        console.error('[pusaka/employees]', data.error);
        return;
      }
      employees = data.items ?? [];
    } catch (e) {
      console.error('[pusaka/employees] load failed:', e);
    }
  }

  async function runNow(employee_id: string, run_type: string) {
    try {
      const res  = await fetch('/api/pusaka/jobs/run-now', {
        method: 'POST',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify({ employee_id, run_type }),
      });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) console.error('[pusaka/run-now]', data.error);
    } catch (e) {
      console.error('[pusaka/run-now] failed:', e);
    }
    const labels: Record<string, string> = { morning: 'rekap', afternoon: 'rekap', checkin: 'absensi masuk', checkout: 'absensi pulang' };
    toast.success(`Job ${labels[run_type] ?? run_type} berhasil di-queue.`);
  }

  function handleStop(_id: string, cancelled: number) {
    toast.success(cancelled > 0 ? `${cancelled} job dibatalkan.` : 'Tidak ada job aktif untuk pegawai ini.');
  }

  onMount(() => {
    load();
    const itv = setInterval(load, 4000);
    return () => clearInterval(itv);
  });
</script>

<svelte:head><title>Pegawai PUSAKA — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6">
  <div>
    <h1 class="text-2xl font-semibold text-slate-800">Pegawai PUSAKA</h1>
    <p class="mt-1 text-sm text-muted-foreground">Area khusus pegawai PNS dan PPPK yang mengikuti integrasi PUSAKA Kemenag.</p>
  </div>

  <EmployeeList {employees} onrun={runNow} onstop={handleStop} ondelete={load} />
</div>
