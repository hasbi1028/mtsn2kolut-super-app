<script lang="ts">
  import { onMount } from 'svelte';

  let profile: any = null;
  let loading = $state(true);
  let error = $state('');

  onMount(async () => {
    try {
      const res = await fetch('/api/school-profile');
      if (res.ok) profile = await res.json();
    } catch { error = 'Gagal memuat profil madrasah'; }
    finally { loading = false; }
  });
</script>

<svelte:head><title>Profil Madrasah — MTsN 2 Kolut</title></svelte:head>

<div class="container-fluid px-0">
  <h4 class="fw-black mb-1">Profil Madrasah</h4>
  <p class="text-secondary mb-3" style="font-size:0.85rem;">Informasi profil madrasah</p>

  {#if loading}
    <div class="text-center py-5">
      <div class="spinner-border text-success" role="status"></div>
    </div>
  {:else if error}
    <div class="alert alert-warning">{error}</div>
  {:else}
    <div class="card border shadow-sm">
      <div class="card-body">
        <dl class="row small mb-0">
          <dt class="col-sm-4 text-secondary">Nama Madrasah</dt>
          <dd class="col-sm-8 fw-semibold">{profile?.nama || profile?.name || '—'}</dd>
          <dt class="col-sm-4 text-secondary">NSM</dt>
          <dd class="col-sm-8">{profile?.nsm || '—'}</dd>
          <dt class="col-sm-4 text-secondary">NPSN</dt>
          <dd class="col-sm-8">{profile?.npsn || '—'}</dd>
          <dt class="col-sm-4 text-secondary">Alamat</dt>
          <dd class="col-sm-8">{profile?.alamat || profile?.address || '—'}</dd>
          <dt class="col-sm-4 text-secondary">Desa/Kelurahan</dt>
          <dd class="col-sm-8">{profile?.desa || profile?.village || '—'}</dd>
          <dt class="col-sm-4 text-secondary">Kecamatan</dt>
          <dd class="col-sm-8">{profile?.kecamatan || profile?.district || '—'}</dd>
          <dt class="col-sm-4 text-secondary">Kabupaten/Kota</dt>
          <dd class="col-sm-8">{profile?.kabupaten || profile?.city || '—'}</dd>
          <dt class="col-sm-4 text-secondary">Provinsi</dt>
          <dd class="col-sm-8">{profile?.provinsi || profile?.province || '—'}</dd>
          <dt class="col-sm-4 text-secondary">Telepon</dt>
          <dd class="col-sm-8">{profile?.telepon || profile?.phone || '—'}</dd>
          <dt class="col-sm-4 text-secondary">Email</dt>
          <dd class="col-sm-8">{profile?.email || '—'}</dd>
          <dt class="col-sm-4 text-secondary">Website</dt>
          <dd class="col-sm-8">{profile?.website || '—'}</dd>
        </dl>
      </div>
    </div>
  {/if}
</div>
