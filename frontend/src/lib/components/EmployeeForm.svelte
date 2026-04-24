<script>
  let { onadd } = $props();

  let form = $state({ nip: '', nama: '', unit_kerja: '', pusaka_username: '', pusaka_password: '' });
  let error = $state('');
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
      body: JSON.stringify(form)
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

<section class="card">
  <h2>Tambah Pegawai</h2>
  {#if error}
    <p class="error">{error}</p>
  {/if}
  <div class="form-grid">
    <input placeholder="NIP" bind:value={form.nip} />
    <input placeholder="Nama" bind:value={form.nama} />
    <input placeholder="Unit Kerja" bind:value={form.unit_kerja} />
    <input placeholder="Username Pusaka" bind:value={form.pusaka_username} />
    <input type="password" placeholder="Password Pusaka" bind:value={form.pusaka_password} />
    <button class="btn" onclick={submit} disabled={loading}>
      {loading ? 'Menyimpan...' : 'Simpan'}
    </button>
  </div>
</section>

<style>
  .form-grid { display: grid; grid-template-columns: repeat(6, minmax(0, 1fr)); gap: 10px; }
  .error { color: #ff9d9d; font-size: 0.9rem; margin: 0 0 10px; }

  @media (max-width: 900px) {
    .form-grid { grid-template-columns: 1fr 1fr; }
  }
</style>
