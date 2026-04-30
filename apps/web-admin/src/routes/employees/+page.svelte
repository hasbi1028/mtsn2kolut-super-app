<script lang="ts">
  import { onMount } from 'svelte';
  import EmployeeForm from '$lib/components/EmployeeForm.svelte';
  import GeneralEmployeeList from '$lib/components/GeneralEmployeeList.svelte';

  let employees = $state<any[]>([]);

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

  <EmployeeForm onadd={load} />
  <GeneralEmployeeList {employees} onreload={load} />
</div>
