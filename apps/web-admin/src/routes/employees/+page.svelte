<script lang="ts">
  import { onMount } from 'svelte';
  import EmployeeForm from '$lib/components/EmployeeForm.svelte';
  import GeneralEmployeeList from '$lib/components/GeneralEmployeeList.svelte';
  import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';

  let employees = $state<any[]>([]);
  let error = $state('');

  async function load() {
    try {
      error = '';
      const res  = await fetch('/api/employees');
      const data = await res.json();
      if (data.error) { error = data.error; return; }
      employees = data.items ?? [];
    } catch (e) {
      error = 'Gagal memuat data pegawai. Coba lagi untuk mengambil master pegawai terbaru.';
    }
  }

  onMount(() => {
    load();
  });
</script>

<svelte:head><title>Pegawai — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6">
  <div>
    <h1 class="text-2xl font-semibold text-slate-800">Master Pegawai</h1>
    <p class="mt-1 text-sm text-muted-foreground">Data seluruh pegawai sekolah. Integrasi akun, jadwal, dan job PUSAKA dikelola terpisah dari area ini.</p>
  </div>

  <div class="grid gap-3 md:grid-cols-3">
    <div class="rounded-2xl border border-emerald-100 bg-emerald-50 px-4 py-4">
      <p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-emerald-700">Total Pegawai</p>
      <p class="mt-2 text-2xl font-semibold text-slate-900">{employees.length}</p>
      <p class="text-sm text-slate-600">seluruh profil pegawai yang tercatat</p>
    </div>
    <div class="rounded-2xl border border-sky-100 bg-sky-50 px-4 py-4">
      <p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-sky-700">Pegawai Aktif</p>
      <p class="mt-2 text-2xl font-semibold text-slate-900">{employees.filter((item) => item.is_active).length}</p>
      <p class="text-sm text-slate-600">siap dipakai untuk akun, akademik, dan operasional</p>
    </div>
    <div class="rounded-2xl border border-amber-100 bg-amber-50 px-4 py-4">
      <p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-amber-700">Eligible PUSAKA</p>
      <p class="mt-2 text-2xl font-semibold text-slate-900">
        {employees.filter((item) => item.employment_type === 'pns' || item.employment_type === 'pppk').length}
      </p>
      <p class="text-sm text-slate-600">subset yang dapat dikelola di area PUSAKA</p>
    </div>
  </div>

  {#if error}
    <RecoveryPanel title="Data Pegawai Belum Tersaji" message={error} onRetry={load} />
  {/if}

  <EmployeeForm onadd={load} />
  <GeneralEmployeeList {employees} onreload={load} />
</div>
