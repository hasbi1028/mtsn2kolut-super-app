<script lang="ts">
  import * as Dialog from '$lib/components/ui/dialog';
  import { Button } from '$lib/components/ui/button';
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

  let showTambahDialog = $state(false);

  // Auto-tutup dialog setelah sukses tambah pegawai
  $effect(() => {
    if (form?.tambahSuccess) {
      showTambahDialog = false;
    }
  });
</script>

<svelte:head><title>Pegawai — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6">
  <div class="flex items-start justify-between gap-4">
    <div>
      <h1 class="text-2xl font-black text-base-content">Master Pegawai</h1>
      <p class="mt-1 text-sm text-base-content/70">Data seluruh pegawai sekolah. Integrasi akun, jadwal, dan job PUSAKA dikelola terpisah dari area ini.</p>
    </div>
    <Button onclick={() => showTambahDialog = true} class="shrink-0">
      + Tambah Pegawai
    </Button>
  </div>

  {#if form?.hapusSuccess}
    <div class="rounded-xl border border-success/20 bg-success/10 px-5 py-4 text-sm text-foreground">
      {form.hapusSuccess}
    </div>
  {/if}
  {#if form?.hapusError}
    <div class="rounded-xl border border-destructive/20 bg-destructive/10 px-5 py-4 text-sm text-foreground">
      {form.hapusError}
    </div>
  {/if}
  {#if form?.nonaktifSuccess}
    <div class="rounded-xl border border-success/20 bg-success/10 px-5 py-4 text-sm text-foreground">
      {form.nonaktifSuccess}
    </div>
  {/if}
  {#if form?.nonaktifError}
    <div class="rounded-xl border border-destructive/20 bg-destructive/10 px-5 py-4 text-sm text-foreground">
      {form.nonaktifError}
    </div>
  {/if}

  <div class="grid gap-4 md:grid-cols-3">
    <div class="rounded-xl border border-base-300 bg-base-100 px-5 py-4 flex flex-col gap-1.5">
      <p class="my-0 text-[10px] font-black uppercase tracking-widest text-muted-foreground">Total Pegawai</p>
      <p class="my-0 text-2xl font-black text-foreground">{employees.length}</p>
      <p class="my-0 text-xs text-muted-foreground">seluruh profil pegawai yang tercatat</p>
    </div>
    <div class="rounded-xl border border-base-300 bg-base-100 px-5 py-4 flex flex-col gap-1.5">
      <p class="my-0 text-[10px] font-black uppercase tracking-widest text-muted-foreground">Pegawai Aktif</p>
      <p class="my-0 text-2xl font-black text-foreground">{employees.filter((item) => item.is_active).length}</p>
      <p class="my-0 text-xs text-muted-foreground">siap dipakai untuk akun, akademik, dan operasional</p>
    </div>
    <div class="rounded-xl border border-base-300 bg-base-100 px-5 py-4 flex flex-col gap-1.5">
      <p class="my-0 text-[10px] font-black uppercase tracking-widest text-muted-foreground">Eligible PUSAKA</p>
      <p class="my-0 text-2xl font-black text-foreground">
        {employees.filter((item) => item.employment_type === 'pns' || item.employment_type === 'pppk').length}
      </p>
      <p class="my-0 text-xs text-muted-foreground">subset yang dapat dikelola di area PUSAKA</p>
    </div>
  </div>

  <GeneralEmployeeList {employees} {form} />
</div>

<Dialog.Root bind:open={showTambahDialog}>
  <Dialog.Content>
    <div class="space-y-4">
      <div>
        <h2 class="text-lg font-semibold text-foreground">Tambah Pegawai</h2>
        <p class="text-sm text-muted-foreground">Master data pegawai sekolah. Integrasi PUSAKA bersifat opsional dan hanya berlaku untuk pegawai PNS atau PPPK.</p>
      </div>
      <EmployeeForm {form} onclose={() => showTambahDialog = false} />
    </div>
  </Dialog.Content>
</Dialog.Root>
