<script>
  import { onMount } from 'svelte';
  import EmployeeForm from '$lib/components/EmployeeForm.svelte';
  import EmployeeList from '$lib/components/EmployeeList.svelte';

  let employees = $state([]);
  let toast     = $state('');

  async function load() {
    const res = await fetch('/api/employees');
    const data = await res.json();
    employees = data.items;
  }

  async function runNow(employee_id, run_type) {
    await fetch('/api/jobs/run-now', {
      method: 'POST',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify({ employee_id, run_type })
    });
    const labels = { morning: 'pagi', afternoon: 'sore', checkin: 'absensi masuk', checkout: 'absensi pulang' };
    showToast(`Job ${labels[run_type] || run_type} berhasil di-queue.`);
  }

  function handleStop(_id, cancelled) {
    showToast(cancelled > 0 ? `${cancelled} job dibatalkan.` : 'Tidak ada job aktif untuk pegawai ini.');
  }

  function showToast(msg) {
    toast = msg;
    setTimeout(() => (toast = ''), 3000);
  }

  onMount(() => {
    load();
    const itv = setInterval(load, 4000);
    return () => clearInterval(itv);
  });
</script>

<svelte:head><title>Pegawai — Pusaka Worker</title></svelte:head>

<h2 class="page-title">Manajemen Pegawai</h2>

{#if toast}
  <div class="toast">{toast}</div>
{/if}

<EmployeeForm onadd={load} />

<EmployeeList {employees} onrun={runNow} onstop={handleStop} ondelete={load} />

<style>
  .page-title { margin: 0 0 16px; font-size: 1.4rem; color: #e7edf7; }
  .toast {
    background: rgba(31, 170, 112, 0.2);
    border: 1px solid rgba(31, 170, 112, 0.4);
    color: #7ff0b7;
    border-radius: 10px;
    padding: 10px 16px;
    font-size: 0.9rem;
    margin-bottom: 4px;
  }
</style>
