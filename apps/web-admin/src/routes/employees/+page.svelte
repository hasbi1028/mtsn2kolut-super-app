<script lang="ts">
  import EmployeeForm from '$lib/components/EmployeeForm.svelte';
  import GeneralEmployeeList from '$lib/components/GeneralEmployeeList.svelte';

  let { data, form } = $props();

  type Employee = {
    id: string;
    pegawai_uid: string;
    nip: string;
    nama: string;
    unit_kerja: string;
    employment_type: string;
    tanggal_lahir: string;
    jenis_kelamin: string;
    tempat_lahir: string;
    pusaka_eligible: boolean;
    has_pusaka_account: boolean;
    pusaka_is_enabled: boolean;
    is_active: boolean;
  };

  let employees = $derived(data.employees as Employee[]);
</script>

<svelte:head><title>Pegawai — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6">
  <div>
    <h1 class="text-2xl font-semibold text-base-content">Master Pegawai</h1>
    <p class="mt-1 text-sm text-base-content/70">Data seluruh pegawai sekolah. Integrasi akun, jadwal, dan job PUSAKA dikelola terpisah dari area ini.</p>
  </div>

  {#if form?.tambahSuccess}
    <div class="rounded-2xl border border-success/20 bg-success/10 px-5 py-4 text-sm text-success-foreground">
      {form.tambahSuccess}
    </div>
  {/if}
  {#if form?.tambahError}
    <div class="rounded-2xl border border-destructive/20 bg-destructive/10 px-5 py-4 text-sm text-destructive-foreground">
      {form.tambahError}
    </div>
  {/if}

  <div class="grid gap-3 md:grid-cols-3">
    <div class="rounded-2xl border border-primary/20 bg-primary/10 px-4 py-4">
      <p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-primary">Total Pegawai</p>
      <p class="mt-2 text-2xl font-semibold text-base-content">{employees.length}</p>
      <p class="text-sm text-base-content/70">seluruh profil pegawai yang tercatat</p>
    </div>
    <div class="rounded-2xl border border-accent bg-accent/60 px-4 py-4">
      <p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-accent-foreground">Pegawai Aktif</p>
      <p class="mt-2 text-2xl font-semibold text-base-content">{employees.filter((item) => item.is_active).length}</p>
      <p class="text-sm text-base-content/70">siap dipakai untuk akun, akademik, dan operasional</p>
    </div>
    <div class="rounded-2xl border border-warning/30 bg-warning/10 px-4 py-4">
      <p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-warning">Eligible PUSAKA</p>
      <p class="mt-2 text-2xl font-semibold text-base-content">
        {employees.filter((item) => item.employment_type === 'pns' || item.employment_type === 'pppk').length}
      </p>
      <p class="text-sm text-base-content/70">subset yang dapat dikelola di area PUSAKA</p>
    </div>
  </div>

  <EmployeeForm />
  <GeneralEmployeeList {employees} />
</div>
