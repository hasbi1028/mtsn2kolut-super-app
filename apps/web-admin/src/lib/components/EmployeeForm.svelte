<script lang="ts">
  import * as Card from '$lib/components/ui/card';
  import { Input } from '$lib/components/ui/input';
  import { Button } from '$lib/components/ui/button';

  let { onadd }: { onadd?: () => void } = $props();

  let form = $state({ nip: '', nama: '', unit_kerja: '', pusaka_username: '', pusaka_password: '' });
  let error   = $state('');
  let loading = $state(false);

  async function submit() {
    if (!form.nip || !form.nama || !form.pusaka_username || !form.pusaka_password) {
      error = 'NIP, Nama, Username, dan Password wajib diisi.';
      return;
    }
    loading = true;
    error = '';
    const res = await fetch('/api/employees', {
      method: 'POST',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify(form),
    });
    loading = false;
    if (!res.ok) {
      const data = await res.json().catch(() => ({}));
      error = data.error || 'Gagal menyimpan pegawai.';
      return;
    }
    form = { nip: '', nama: '', unit_kerja: '', pusaka_username: '', pusaka_password: '' };
    onadd?.();
  }
</script>

<Card.Root>
  <Card.Header class="pb-3">
    <Card.Title class="text-base">Tambah Pegawai</Card.Title>
  </Card.Header>
  <Card.Content>
    {#if error}
      <p class="mb-3 text-sm text-destructive">{error}</p>
    {/if}
    <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-6">
      <div class="xl:col-span-1">
        <label class="mb-1 block text-xs font-medium text-muted-foreground">NIP <span class="text-destructive">*</span></label>
        <Input placeholder="NIP Pegawai" bind:value={form.nip} />
      </div>
      <div class="xl:col-span-1">
        <label class="mb-1 block text-xs font-medium text-muted-foreground">Nama <span class="text-destructive">*</span></label>
        <Input placeholder="Nama Lengkap" bind:value={form.nama} />
      </div>
      <div class="xl:col-span-1">
        <label class="mb-1 block text-xs font-medium text-muted-foreground">Unit Kerja</label>
        <Input placeholder="Unit Kerja" bind:value={form.unit_kerja} />
      </div>
      <div class="xl:col-span-1">
        <label class="mb-1 block text-xs font-medium text-muted-foreground">Username Pusaka <span class="text-destructive">*</span></label>
        <Input placeholder="Username Pusaka" bind:value={form.pusaka_username} />
      </div>
      <div class="xl:col-span-1">
        <label class="mb-1 block text-xs font-medium text-muted-foreground">Password Pusaka <span class="text-destructive">*</span></label>
        <Input type="password" placeholder="Password Pusaka" bind:value={form.pusaka_password} />
      </div>
      <div class="xl:col-span-1 flex items-end">
        <Button class="w-full" onclick={submit} disabled={loading}>
          {loading ? 'Menyimpan...' : 'Simpan'}
        </Button>
      </div>
    </div>
  </Card.Content>
</Card.Root>
