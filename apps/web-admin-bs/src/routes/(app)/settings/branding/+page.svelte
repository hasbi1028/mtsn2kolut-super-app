<script lang="ts">
  import { onMount } from 'svelte';

  let branding: any = null;
  let loading = $state(true);
  let error = $state('');

  onMount(async () => {
    try {
      const res = await fetch('/api/branding');
      if (res.ok) branding = await res.json();
    } catch { error = 'Gagal memuat branding'; }
    finally { loading = false; }
  });
</script>

<svelte:head><title>Logo & Branding — MTsN 2 Kolut</title></svelte:head>

<div class="container-fluid px-0">
  <h4 class="fw-black mb-1">Logo & Branding</h4>
  <p class="text-secondary mb-3" style="font-size:0.85rem;">Pengaturan logo dan tampilan portal publik</p>

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
          <dt class="col-sm-4 text-secondary">Nama Pendek</dt>
          <dd class="col-sm-8 fw-semibold">{branding?.short_name || '—'}</dd>
          <dt class="col-sm-4 text-secondary">Nama Panjang</dt>
          <dd class="col-sm-8">{branding?.long_name || branding?.longName || '—'}</dd>
          <dt class="col-sm-4 text-secondary">Warna Tema</dt>
          <dd class="col-sm-8">
            {#if branding?.theme_color || branding?.themeColor}
              <span class="d-inline-block rounded-circle me-1" style="width:16px;height:16px;background:{branding.theme_color || branding.themeColor};vertical-align:middle;"></span>
              {branding.theme_color || branding.themeColor}
            {:else}—{/if}
          </dd>
          <dt class="col-sm-4 text-secondary">Logo</dt>
          <dd class="col-sm-8">
            {#if branding?.mark_url || branding?.logo_url}
              <img src={branding.mark_url || branding.logo_url} alt="Logo" class="rounded border" style="max-height:64px;" />
            {:else}—{/if}
          </dd>
          <dt class="col-sm-4 text-secondary">Favicon</dt>
          <dd class="col-sm-8">{branding?.favicon_url || '—'}</dd>
          <dt class="col-sm-4 text-secondary">Versi</dt>
          <dd class="col-sm-8">{branding?.version || '—'}</dd>
        </dl>
      </div>
    </div>
  {/if}
</div>
