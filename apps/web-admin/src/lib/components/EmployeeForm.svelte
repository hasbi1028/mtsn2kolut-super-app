<script lang="ts">
  import * as Card from '$lib/components/ui/card';
  import { Input } from '$lib/components/ui/input';
  import PasswordInput from '$lib/components/PasswordInput.svelte';

  let { form }: { form?: { tambahError?: string; tambahSuccess?: string } } = $props();

  let pusakaEligible = $state(false);
</script>

<Card.Root>
  <Card.Header class="pb-3">
    <Card.Title class="text-base">Tambah Pegawai</Card.Title>
    <Card.Description>Master data pegawai sekolah. Integrasi PUSAKA bersifat opsional dan hanya berlaku untuk pegawai PNS atau PPPK.</Card.Description>
  </Card.Header>
  <Card.Content>
    {#if form?.tambahSuccess}
      <div class="mb-3 rounded-xl border border-success/20 bg-success/10 px-5 py-4 text-sm text-success-foreground">
        {form.tambahSuccess}
      </div>
    {/if}
    {#if form?.tambahError}
      <div class="mb-3 rounded-xl border border-destructive/20 bg-destructive/10 px-5 py-4 text-sm text-destructive-foreground">
        {form.tambahError}
      </div>
    {/if}

    <form method="POST" action="?/tambah">
      <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
        <div>
          <label for="f-nip" class="mb-1 block text-xs font-medium text-muted-foreground">NIP</label>
          <Input id="f-nip" name="nip" placeholder="Opsional untuk honorer" />
        </div>
        <div>
          <label for="f-nama" class="mb-1 block text-xs font-medium text-muted-foreground">Nama <span class="text-destructive">*</span></label>
          <Input id="f-nama" name="nama" placeholder="Masukkan nama lengkap pegawai" required />
        </div>
        <div>
          <label for="f-unit" class="mb-1 block text-xs font-medium text-muted-foreground">Unit Kerja</label>
          <Input id="f-unit" name="unit_kerja" placeholder="Contoh: Tata Usaha atau Kurikulum" />
        </div>

        <div>
          <label for="f-tempat-lahir" class="mb-1 block text-xs font-medium text-muted-foreground">Tempat Lahir</label>
          <Input id="f-tempat-lahir" name="tempat_lahir" placeholder="Contoh: Olo-oloho" />
        </div>
        <div>
          <label for="f-tanggal-lahir" class="mb-1 block text-xs font-medium text-muted-foreground">Tanggal Lahir</label>
          <Input id="f-tanggal-lahir" name="tanggal_lahir" type="date" />
          <p class="mt-1 text-[11px] text-muted-foreground">Dipakai untuk membuat nomor internal pegawai otomatis.</p>
        </div>
        <div>
          <label for="f-jenis-kelamin" class="mb-1 block text-xs font-medium text-muted-foreground">Jenis Kelamin</label>
          <select id="f-jenis-kelamin" name="jenis_kelamin" class="h-10 w-full rounded-lg border border-input bg-background px-3 text-sm text-foreground">
            <option value="">Belum diisi</option>
            <option value="L">Laki-laki</option>
            <option value="P">Perempuan</option>
          </select>
        </div>
        <div>
          <label for="f-employment-type" class="mb-1 block text-xs font-medium text-muted-foreground">Status Kepegawaian <span class="text-destructive">*</span></label>
          <select id="f-employment-type" name="employment_type" class="h-10 w-full rounded-lg border border-input bg-background px-3 text-sm text-foreground"
            onchange={(e) => { pusakaEligible = (e.target as HTMLSelectElement).value === 'pns' || (e.target as HTMLSelectElement).value === 'pppk'; }}>
            <option value="">Pilih status</option>
            <option value="pns">PNS</option>
            <option value="pppk">PPPK</option>
            <option value="honorer">Honorer</option>
            <option value="lainnya">Lainnya</option>
          </select>
        </div>
      </div>

      <div class="mt-4 rounded-xl border border-primary/20 bg-primary/10 p-4">
        <div class="space-y-1">
          <p class="text-sm font-semibold text-foreground">Integrasi PUSAKA</p>
          <p class="text-xs text-muted-foreground">Opsional saat tambah pegawai. Bisa diisi sekarang atau dilengkapi nanti dari halaman PUSAKA.</p>
        </div>

        {#if pusakaEligible}
          <div class="mt-4 grid grid-cols-1 gap-3 sm:grid-cols-2">
            <div>
              <label for="f-user" class="mb-1 block text-xs font-medium text-muted-foreground">Username PUSAKA</label>
              <Input id="f-user" name="pusaka_username" placeholder="Isi jika akun PUSAKA sudah tersedia" />
            </div>
            <div>
              <label for="f-pass" class="mb-1 block text-xs font-medium text-muted-foreground">Password PUSAKA</label>
              <PasswordInput id="f-pass" name="pusaka_password" placeholder="Isi bersama username PUSAKA" />
            </div>
          </div>
        {:else}
          <div class="mt-4 rounded-lg border border-border bg-card px-4 py-3 text-sm text-muted-foreground">
            Pegawai dengan status ini tidak otomatis eligible untuk integrasi PUSAKA. Simpan sebagai pegawai umum saja.
          </div>
        {/if}
      </div>

      <div class="mt-4 flex justify-end">
        <button type="submit" class="inline-flex min-w-36 items-center justify-center rounded-lg bg-primary px-5 py-2.5 text-sm font-semibold text-primary-foreground hover:bg-primary/90">
          Simpan Pegawai
        </button>
      </div>
    </form>
  </Card.Content>
</Card.Root>
