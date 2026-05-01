<script lang="ts">
  import * as Card from '$lib/components/ui/card';
  import { Input } from '$lib/components/ui/input';
  import { toast } from '$lib/components/ui/sonner';
  import LoadingButton from '$lib/components/LoadingButton.svelte';
  import SuccessPanel from '$lib/components/SuccessPanel.svelte';
  import { readClientJson } from '$lib/client/api';

  let { onadd }: { onadd?: () => void } = $props();

  let form = $state({
    nip: '',
    nama: '',
    unit_kerja: '',
    employment_type: '',
    pusaka_username: '',
    pusaka_password: ''
  });
  let success = $state('');
  let saving = $state(false);
  let pusakaEligible = $derived(form.employment_type === 'pns' || form.employment_type === 'pppk');

  function showError(message: string) {
    toast.error(message);
  }

  function mutationErrorMessage(error: unknown, fallbackMessage: string) {
    if (error instanceof Error && error.message.trim()) return error.message;
    return fallbackMessage;
  }

  async function submit() {
    if (!form.nip || !form.nama || !form.employment_type) {
      success = '';
      showError('NIP, Nama, dan status kepegawaian wajib diisi.');
      return;
    }
    if ((form.pusaka_username && !form.pusaka_password) || (!form.pusaka_username && form.pusaka_password)) {
      success = '';
      showError('Username dan password PUSAKA harus diisi berpasangan.');
      return;
    }
    if (!pusakaEligible && (form.pusaka_username || form.pusaka_password)) {
      success = '';
      showError('Hanya pegawai PNS atau PPPK yang boleh memiliki akun PUSAKA.');
      return;
    }
    saving = true;
    success = '';
    try {
      const res = await fetch('/api/employees', {
        method: 'POST',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify(form),
      });
      await readClientJson<unknown>(res);
      form = { nip: '', nama: '', unit_kerja: '', employment_type: '', pusaka_username: '', pusaka_password: '' };
      success = 'Pegawai baru berhasil ditambahkan ke master data. Jika pegawai eligible PUSAKA, akun integrasinya bisa dilengkapi sekarang atau nanti dari menu PUSAKA.';
      onadd?.();
    } catch (error) {
      showError(mutationErrorMessage(error, 'Gagal menyimpan pegawai.'));
    } finally {
      saving = false;
    }
  }
</script>

<Card.Root>
  <Card.Header class="pb-3">
    <Card.Title class="text-base">Tambah Pegawai</Card.Title>
    <Card.Description>Master data pegawai sekolah. Integrasi PUSAKA bersifat opsional dan hanya berlaku untuk pegawai PNS atau PPPK.</Card.Description>
  </Card.Header>
  <Card.Content>
    {#if success}
      <div class="mb-3">
        <SuccessPanel title="Pegawai Berhasil Ditambahkan" message={success} compact />
      </div>
    {/if}
    <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
      <div>
        <label for="f-nip" class="mb-1 block text-xs font-medium text-muted-foreground">NIP <span class="text-destructive">*</span></label>
        <Input id="f-nip" placeholder="Masukkan NIP pegawai" bind:value={form.nip} />
      </div>
      <div>
        <label for="f-nama" class="mb-1 block text-xs font-medium text-muted-foreground">Nama <span class="text-destructive">*</span></label>
        <Input id="f-nama" placeholder="Masukkan nama lengkap pegawai" bind:value={form.nama} />
      </div>
      <div>
        <label for="f-unit" class="mb-1 block text-xs font-medium text-muted-foreground">Unit Kerja</label>
        <Input id="f-unit" placeholder="Contoh: Tata Usaha atau Kurikulum" bind:value={form.unit_kerja} />
      </div>
      <div>
        <label for="f-employment-type" class="mb-1 block text-xs font-medium text-muted-foreground">Status Kepegawaian <span class="text-destructive">*</span></label>
        <select id="f-employment-type" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={form.employment_type}>
          <option value="">Pilih status</option>
          <option value="pns">PNS</option>
          <option value="pppk">PPPK</option>
          <option value="honorer">Honorer</option>
          <option value="lainnya">Lainnya</option>
        </select>
      </div>
    </div>

    <div class="mt-4 rounded-xl border border-emerald-200/70 bg-emerald-50/40 p-4">
      <div class="space-y-1">
        <p class="text-sm font-semibold text-slate-900">Integrasi PUSAKA</p>
        <p class="text-xs text-slate-600">Opsional saat tambah pegawai. Bisa diisi sekarang atau dilengkapi nanti dari halaman PUSAKA.</p>
      </div>

      {#if pusakaEligible}
        <div class="mt-4 grid grid-cols-1 gap-3 sm:grid-cols-2">
          <div>
            <label for="f-user" class="mb-1 block text-xs font-medium text-muted-foreground">Username PUSAKA</label>
            <Input id="f-user" placeholder="Isi jika akun PUSAKA sudah tersedia" bind:value={form.pusaka_username} />
          </div>
          <div>
            <label for="f-pass" class="mb-1 block text-xs font-medium text-muted-foreground">Password PUSAKA</label>
            <Input id="f-pass" type="password" placeholder="Isi bersama username PUSAKA" bind:value={form.pusaka_password} />
          </div>
        </div>
      {:else}
        <div class="mt-4 rounded-lg border border-slate-200 bg-white px-4 py-3 text-sm text-slate-600">
          Pegawai dengan status ini tidak otomatis eligible untuk integrasi PUSAKA. Simpan sebagai pegawai umum saja.
        </div>
      {/if}
    </div>

    <div class="mt-4 flex justify-end">
      <LoadingButton class="min-w-36" onclick={() => void submit()} loading={saving} loadingLabel="Menyimpan..." disabled={saving}>
        Simpan Pegawai
      </LoadingButton>
    </div>
  </Card.Content>
</Card.Root>
