<script lang="ts">
  import { onMount } from 'svelte';

  interface QueueStats {
    queued: number; running: number; success: number;
    failed: number; retry_due: number; total: number;
  }

  interface WorkerEntry {
    worker_id: string; active_consumers: number;
    target_concurrency: number; headless: boolean; reported_at: string;
  }

  interface WorkerStatus {
    active_workers: WorkerEntry[];
    total: number;
    queue: { queued: number; running: number; success: number; failed: number };
  }

  interface RecentJob {
    id: string; created_at: string; nama?: string;
    employee_nama?: string; run_type: string; status: string;
  }

  let loading = $state(true);
  let error = $state('');
  let queueStats = $state<QueueStats | null>(null);
  let workerStatus = $state<WorkerStatus | null>(null);
  let recentJobs = $state<RecentJob[]>([]);
  let busy = $state<Record<string, boolean>>({});

  onMount(() => loadOverview());

  async function loadOverview() {
    loading = true; error = '';
    try {
      const [queueRes, workerRes, jobsRes] = await Promise.all([
        fetch('/api/pusaka/jobs/stats'),
        fetch('/api/pusaka/worker/status'),
        fetch('/api/pusaka/jobs?limit=10')
      ]);
      if (queueRes.ok) queueStats = await queueRes.json();
      if (workerRes.ok) workerStatus = await workerRes.json();
      if (jobsRes.ok) {
        const jd = await jobsRes.json();
        recentJobs = jd.items || jd.data || [];
      }
    } catch (e: any) {
      error = 'Gagal memuat data PUSAKA';
    } finally { loading = false; }
  }

  async function runAllJobs() {
    busy = { ...busy, runAll: true };
    try {
      const res = await fetch('/api/pusaka/jobs/run-all', { method: 'POST' });
      if (res.ok) await loadOverview();
    } catch {} finally { busy = { ...busy, runAll: false }; }
  }

  async function cancelAllJobs() {
    busy = { ...busy, cancelAll: true };
    try {
      const res = await fetch('/api/pusaka/jobs/cancel-all', { method: 'POST' });
      if (res.ok) await loadOverview();
    } catch {} finally { busy = { ...busy, cancelAll: false }; }
  }

  function statusBadge(status: string) {
    const map: Record<string, string> = {
      queued: 'bg-secondary', running: 'bg-primary',
      success: 'bg-success', failed: 'bg-danger',
      cancelled: 'bg-warning text-dark', retry_due: 'bg-info text-dark'
    };
    return map[status] || 'bg-secondary';
  }

  function formatDate(ts: string) {
    if (!ts) return '—';
    return new Date(ts).toLocaleString('id-ID', { timeZone: 'Asia/Makassar', dateStyle: 'short', timeStyle: 'short' });
  }
</script>

<svelte:head><title>Monitor Kehadiran — MTsN 2 Kolut</title></svelte:head>

<div class="container-fluid px-0">
  <h4 class="fw-black mb-1">Monitor Kehadiran</h4>
  <p class="text-secondary mb-4" style="font-size:0.85rem;">Pantau status sinkronisasi kehadiran PUSAKA</p>

  {#if loading}
    <div class="text-center py-5">
      <div class="spinner-border text-success" role="status"></div>
      <p class="mt-2 text-secondary small">Memuat data...</p>
    </div>
  {:else if error}
    <div class="alert alert-warning d-flex align-items-center gap-2">
      <i class="bi bi-exclamation-triangle"></i>
      <span>{error}</span>
      <button class="btn btn-sm btn-outline-secondary ms-auto" onclick={loadOverview}>
        <i class="bi bi-arrow-repeat me-1"></i>Muat Ulang
      </button>
    </div>
  {:else}

    <!-- Stat Cards -->
    <div class="row g-3 mb-4">
      <div class="col-6 col-lg-2">
        <div class="stat-card bg-white">
          <div class="stat-icon bg-info bg-opacity-10 text-info mb-2"><i class="bi bi-hourglass-split"></i></div>
          <div class="fw-black fs-5">{queueStats?.queued ?? 0}</div>
          <div class="text-secondary small">Antrian</div>
        </div>
      </div>
      <div class="col-6 col-lg-2">
        <div class="stat-card bg-white">
          <div class="stat-icon bg-primary bg-opacity-10 text-primary mb-2"><i class="bi bi-arrow-repeat"></i></div>
          <div class="fw-black fs-5">{queueStats?.running ?? 0}</div>
          <div class="text-secondary small">Berjalan</div>
        </div>
      </div>
      <div class="col-6 col-lg-2">
        <div class="stat-card bg-white">
          <div class="stat-icon bg-success bg-opacity-10 text-success mb-2"><i class="bi bi-check-circle"></i></div>
          <div class="fw-black fs-5">{queueStats?.success ?? 0}</div>
          <div class="text-secondary small">Sukses</div>
        </div>
      </div>
      <div class="col-6 col-lg-2">
        <div class="stat-card bg-white">
          <div class="stat-icon bg-danger bg-opacity-10 text-danger mb-2"><i class="bi bi-x-circle"></i></div>
          <div class="fw-black fs-5">{queueStats?.failed ?? 0}</div>
          <div class="text-secondary small">Gagal</div>
        </div>
      </div>
      <div class="col-6 col-lg-2">
        <div class="stat-card bg-white">
          <div class="stat-icon bg-warning bg-opacity-10 text-warning mb-2"><i class="bi bi-clock-history"></i></div>
          <div class="fw-black fs-5">{queueStats?.retry_due ?? 0}</div>
          <div class="text-secondary small">Retry</div>
        </div>
      </div>
      <div class="col-6 col-lg-2">
        <div class="stat-card bg-white">
          <div class="stat-icon bg-secondary bg-opacity-10 text-secondary mb-2"><i class="bi bi-list-ol"></i></div>
          <div class="fw-black fs-5">{queueStats?.total ?? 0}</div>
          <div class="text-secondary small">Total</div>
        </div>
      </div>
    </div>

    <!-- Action Buttons -->
    <div class="d-flex gap-2 mb-4">
      <button class="btn btn-success btn-sm" onclick={runAllJobs} disabled={busy.runAll}>
        {#if busy.runAll}
          <span class="spinner-border spinner-border-sm me-1"></span>
        {:else}
          <i class="bi bi-play-fill me-1"></i>
        {/if}
        Jalankan Semua
      </button>
      <button class="btn btn-outline-danger btn-sm" onclick={cancelAllJobs} disabled={busy.cancelAll}>
        {#if busy.cancelAll}
          <span class="spinner-border spinner-border-sm me-1"></span>
        {:else}
          <i class="bi bi-stop-fill me-1"></i>
        {/if}
        Batalkan Semua
      </button>
      <button class="btn btn-outline-secondary btn-sm ms-auto" onclick={loadOverview}>
        <i class="bi bi-arrow-repeat me-1"></i>Refresh
      </button>
    </div>

    <!-- Worker Status -->
    {#if workerStatus?.active_workers?.length}
      <div class="card border shadow-sm mb-4">
        <div class="card-header bg-white fw-bold small py-2">Worker Status
          <span class="badge bg-success ms-2">{workerStatus.total} aktif</span>
        </div>
        <div class="card-body p-0">
          <div class="table-responsive">
            <table class="table table-sm mb-0 small">
              <thead class="table-light">
                <tr>
                  <th>Worker ID</th>
                  <th>Consumer</th>
                  <th>Concurrency</th>
                  <th>Headless</th>
                  <th>Reported</th>
                </tr>
              </thead>
              <tbody>
                {#each workerStatus.active_workers as w}
                  <tr>
                    <td class="text-muted" style="font-family:monospace;font-size:0.75rem;">{w.worker_id?.slice(0,12)}...</td>
                    <td>{w.active_consumers}</td>
                    <td>{w.target_concurrency}</td>
                    <td>
                      <span class="badge {w.headless ? 'bg-secondary' : 'bg-success'}">{w.headless ? 'Ya' : 'Tidak'}</span>
                    </td>
                    <td>{formatDate(w.reported_at)}</td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    {:else if workerStatus}
      <div class="alert alert-secondary small mb-4" role="alert">
        <i class="bi bi-info-circle me-2"></i>Tidak ada worker aktif
      </div>
    {/if}

    <!-- Recent Jobs -->
    <div class="card border shadow-sm">
      <div class="card-header bg-white fw-bold small py-2">Job Terbaru (10 terakhir)</div>
      <div class="card-body p-0">
        {#if recentJobs.length === 0}
          <div class="text-center py-4 text-secondary small">Belum ada job</div>
        {:else}
          <div class="table-responsive">
            <table class="table table-sm mb-0 small">
              <thead class="table-light">
                <tr>
                  <th>#</th>
                  <th>Nama</th>
                  <th>Pegawai</th>
                  <th>Tipe</th>
                  <th>Status</th>
                  <th>Waktu</th>
                </tr>
              </thead>
              <tbody>
                {#each recentJobs as job, i}
                  <tr>
                    <td class="text-muted">{i + 1}</td>
                    <td class="fw-semibold">{job.nama || '—'}</td>
                    <td>{job.employee_nama || '—'}</td>
                    <td><span class="badge bg-light text-dark">{job.run_type}</span></td>
                    <td><span class="badge {statusBadge(job.status)}">{job.status}</span></td>
                    <td class="text-muted">{formatDate(job.created_at)}</td>
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
