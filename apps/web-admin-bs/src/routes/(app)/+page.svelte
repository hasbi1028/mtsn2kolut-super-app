<script lang="ts">
  import { browser } from '$app/environment';

  let healthData: Record<string, any> | null = null;
  let loading = true;
  let error = '';

  async function loadHealth() {
    loading = true;
    error = '';
    try {
      const res = await fetch('/api/health');
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      healthData = await res.json();
    } catch (e: any) {
      error = 'Gagal memuat data dashboard';
    } finally {
      loading = false;
    }
  }

  if (browser) {
    loadHealth();
  }
</script>

<div>
  <h3 class="fw-black mb-1">Dashboard</h3>
  <p class="text-secondary mb-4">Selamat datang, Admin</p>

  {#if loading}
    <div class="text-center py-5">
      <div class="spinner-border text-success" role="status"></div>
      <p class="mt-2 text-secondary small">Memuat data...</p>
    </div>
  {:else if error}
    <div class="row justify-content-center">
      <div class="col-md-6 text-center py-5">
        <i class="bi bi-exclamation-circle text-warning" style="font-size:3rem;"></i>
        <p class="text-secondary mt-2">{error}</p>
        <button class="btn btn-primary btn-sm" onclick={loadHealth}>
          <i class="bi bi-arrow-repeat me-1"></i>Coba Muat Ulang
        </button>
      </div>
    </div>
  {:else}
    <div class="row g-3 mb-4">
      <div class="col-sm-6 col-lg-3">
        <div class="stat-card bg-white">
          <div class="stat-icon bg-success bg-opacity-10 text-success mb-3"><i class="bi bi-check-circle fs-5"></i></div>
          <div class="fw-black fs-5">{healthData?.status === 'ok' ? 'Online' : 'Tidak Diketahui'}</div>
          <div class="text-secondary small">Status Server</div>
          <div class="text-secondary" style="font-size:0.7rem;">API Backend</div>
        </div>
      </div>
      <div class="col-sm-6 col-lg-3">
        <div class="stat-card bg-white">
          <div class="stat-icon bg-success bg-opacity-10 text-success mb-3"><i class="bi bi-database fs-5"></i></div>
          <div class="fw-black fs-5">{healthData?.db === 'connected' ? 'Tersambung' : 'Putus'}</div>
          <div class="text-secondary small">Database</div>
          <div class="text-secondary" style="font-size:0.7rem;">PostgreSQL</div>
        </div>
      </div>
      <div class="col-sm-6 col-lg-3">
        <div class="stat-card bg-white">
          <div class="stat-icon bg-primary bg-opacity-10 text-primary mb-3"><i class="bi bi-grid fs-5"></i></div>
          <div class="fw-black fs-5">SMM {healthData?.server?.version ?? 'N/A'}</div>
          <div class="text-secondary small">Aplikasi</div>
          <div class="text-secondary" style="font-size:0.7rem;">Sistem Manajemen Madrasah</div>
        </div>
      </div>
      <div class="col-sm-6 col-lg-3">
        <div class="stat-card bg-white">
          <div class="stat-icon bg-warning bg-opacity-10 text-warning mb-3"><i class="bi bi-building fs-5"></i></div>
          <div class="fw-black fs-5">MTsN 2 Kolut</div>
          <div class="text-secondary small">Madrasah</div>
          <div class="text-secondary" style="font-size:0.7rem;">MTs Negeri 2 Kolaka Utara</div>
        </div>
      </div>
    </div>

    <div class="alert alert-success d-flex align-items-center gap-3 border-0" role="alert">
      <i class="bi bi-info-circle fs-4"></i>
      <div>
        <div class="fw-bold small">Sistem Manajemen Madrasah</div>
        <div class="small text-secondary">MTsN 2 Kolaka Utara &middot; WITA &middot; by Hasbi Awal</div>
      </div>
    </div>
  {/if}
</div>
