<script lang="ts">
  import { onMount } from 'svelte';

  let summary: any = null;
  let daily: any[] = [];
  let loading = $state(true);
  let error = $state('');

  onMount(loadData);

  async function loadData() {
    loading = true; error = '';
    try {
      const [sumRes, dailyRes] = await Promise.all([
        fetch('/api/internal-analytics/summary'),
        fetch('/api/internal-analytics/daily')
      ]);
      if (sumRes.ok) summary = await sumRes.json();
      if (dailyRes.ok) {
        const d = await dailyRes.json();
        daily = d.daily || d.data || [];
      }
    } catch { error = 'Gagal memuat statistik'; }
    finally { loading = false; }
  }

  function formatDate(ts: string) {
    if (!ts) return '—';
    return new Date(ts).toLocaleDateString('id-ID', { timeZone: 'Asia/Makassar' });
  }
</script>

<svelte:head><title>Statistik — MTsN 2 Kolut</title></svelte:head>

<div class="container-fluid px-0">
  <div class="d-flex align-items-center justify-content-between mb-1">
    <h4 class="fw-black mb-0">Statistik Sistem</h4>
    <button class="btn btn-outline-success btn-sm" onclick={loadData}>
      <i class="bi bi-arrow-repeat me-1"></i>Refresh
    </button>
  </div>
  <p class="text-secondary mb-3" style="font-size:0.85rem;">Ringkasan penggunaan dan aktivitas sistem</p>

  {#if loading}
    <div class="text-center py-5">
      <div class="spinner-border text-success" role="status"></div>
    </div>
  {:else if error}
    <div class="alert alert-warning">{error}</div>
  {:else}
    <!-- Summary Cards -->
    {#if summary}
      <div class="row g-3 mb-4">
        <div class="col-6 col-lg-3">
          <div class="stat-card bg-white text-center">
            <div class="fw-black fs-4 text-primary">{summary.total_users || summary.users || 0}</div>
            <div class="text-secondary small">Pengguna</div>
          </div>
        </div>
        <div class="col-6 col-lg-3">
          <div class="stat-card bg-white text-center">
            <div class="fw-black fs-4 text-success">{summary.total_employees || summary.employees || 0}</div>
            <div class="text-secondary small">Pegawai</div>
          </div>
        </div>
        <div class="col-6 col-lg-3">
          <div class="stat-card bg-white text-center">
            <div class="fw-black fs-4 text-info">{summary.total_students || summary.students || 0}</div>
            <div class="text-secondary small">Siswa</div>
          </div>
        </div>
        <div class="col-6 col-lg-3">
          <div class="stat-card bg-white text-center">
            <div class="fw-black fs-4 text-warning">{summary.active_sessions || summary.sessions || 0}</div>
            <div class="text-secondary small">Sesi Aktif</div>
          </div>
        </div>
      </div>
    {/if}

    <!-- Daily Activity -->
    <div class="card border shadow-sm">
      <div class="card-header bg-white fw-bold small py-2">Aktivitas Harian</div>
      <div class="card-body p-0">
        {#if daily.length === 0}
          <div class="text-center py-4 text-secondary small">Belum ada data aktivitas</div>
        {:else}
          <div class="table-responsive">
            <table class="table table-sm mb-0 small">
              <thead class="table-light">
                <tr>
                  <th>Tanggal</th>
                  <th>Login</th>
                  <th>Aksi</th>
                  <th>Pengguna Aktif</th>
                </tr>
              </thead>
              <tbody>
                {#each daily as d}
                  <tr>
                    <td class="fw-semibold">{formatDate(d.tanggal || d.date || d.day)}</td>
                    <td>{d.logins || d.login_count || 0}</td>
                    <td>{d.actions || d.action_count || 0}</td>
                    <td>{d.active_users || d.users || 0}</td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        {/if}
      </div>
    </div>
  {/if}
</div>
