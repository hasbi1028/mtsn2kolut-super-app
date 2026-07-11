<script lang="ts">
  import { onMount, onDestroy } from 'svelte';

  interface QueueJob {
    id: string;
    created_at: string;
    run_type: string;
    status: string;
    error_message?: string;
    nama?: string;
    nip?: string;
    claimed_by?: string;
    attempts?: number;
    max_attempts?: number;
  }

  let jobs = $state<QueueJob[]>([]);
  let stats = $state<any>(null);
  let loading = $state(true);
  let error = $state('');
  let busy = $state<Record<string, boolean>>({});
  let pollingActive = $state(false);
  let lastRefreshTs = $state<number | null>(null);

  // Format time without using toLocaleTimeString (SSR-safe)
  const lastRefreshText = $derived(lastRefreshTs ? formatTime(new Date(lastRefreshTs)) : '');

  function formatTime(d: Date) {
    const pad = (n: number) => String(n).padStart(2, '0');
    return `${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())} WITA`;
  }

  let pollTimer: ReturnType<typeof setInterval> | null = null;
  const POLL_INTERVAL = 30000; // 30 detik

  onMount(() => {
    loadData();
    startPolling();
  });

  onDestroy(() => {
    stopPolling();
  });

  function startPolling() {
    if (typeof document === 'undefined') return;
    if (pollTimer) return;
    document.addEventListener('visibilitychange', handleVisibility);
    pollTimer = setInterval(() => {
      if (document.visibilityState === 'visible') {
        loadData({ silent: true });
      }
    }, POLL_INTERVAL);
    pollingActive = true;
  }

  function stopPolling() {
    if (typeof document === 'undefined') return;
    document.removeEventListener('visibilitychange', handleVisibility);
    if (pollTimer) {
      clearInterval(pollTimer);
      pollTimer = null;
    }
    pollingActive = false;
  }

  function handleVisibility() {
    if (typeof document !== 'undefined' && document.visibilityState === 'visible') {
      loadData({ silent: true });
    }
  }

  async function loadData(opts: { silent?: boolean } = {}) {
    if (!opts.silent) {
      loading = true;
    }
    error = '';
    try {
      const [jobsRes, statsRes] = await Promise.all([
        fetch('/api/pusaka/jobs'),
        fetch('/api/pusaka/jobs/stats')
      ]);
      if (jobsRes.ok) {
        const jd = await jobsRes.json();
        jobs = jd.items || jd.data || [];
      }
      if (statsRes.ok) stats = await statsRes.json();
      lastRefreshTs = Date.now();
    } catch (e: any) {
      error = 'Gagal memuat antrian';
    } finally {
      loading = false;
    }
  }

  async function cancelJob(id: string) {
    busy = { ...busy, [id]: true };
    try {
      await fetch('/api/pusaka/jobs/cancel', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ job_id: id })
      });
      await loadData();
    } catch {} finally {
      busy = { ...busy, [id]: false };
    }
  }

  function statusBadge(s: string) {
    const m: Record<string, string> = {
      queued: 'bg-secondary', running: 'bg-primary',
      success: 'bg-success', failed: 'bg-danger',
      cancelled: 'bg-warning text-dark'
    };
    return m[s] || 'bg-secondary';
  }

  function formatDate(ts: string) {
    if (!ts) return '—';
    return new Date(ts).toLocaleString('id-ID', { timeZone: 'Asia/Makassar', dateStyle: 'short', timeStyle: 'short' });
  }
</script>

<svelte:head><title>Antrian Sinkronisasi — MTsN 2 Kolut</title></svelte:head>

<div class="container-fluid px-0">
  <div class="d-flex align-items-center justify-content-between mb-1">
    <h4 class="fw-black mb-0">Antrian Sinkronisasi</h4>
    <div class="d-flex align-items-center gap-2">
      {#if pollingActive}
        <span class="badge bg-success-subtle text-success small" title="Auto-refresh setiap {POLL_INTERVAL / 1000} detik">
          <i class="bi bi-arrow-clockwise"></i> Auto: {POLL_INTERVAL / 1000}s
        </span>
        {#if lastRefreshText}
          <span class="text-muted" style="font-size:0.75rem;" title="Update terakhir">
            {lastRefreshText}
          </span>
        {/if}
      {/if}
      <button class="btn btn-outline-success btn-sm" onclick={() => loadData()} disabled={loading}>
        <i class="bi {loading ? 'bi-arrow-repeat spin' : 'bi-arrow-repeat'} me-1"></i>Refresh
      </button>
    </div>
  </div>
  <p class="text-secondary mb-3" style="font-size:0.85rem;">Daftar job sinkronisasi PUSAKA</p>

  {#if stats}
    <div class="row g-2 mb-4">
      <div class="col-auto">
        <span class="badge bg-secondary fs-6 p-2">Antri: {stats.queued || 0}</span>
      </div>
      <div class="col-auto">
        <span class="badge bg-primary fs-6 p-2">Berjalan: {stats.running || 0}</span>
      </div>
      <div class="col-auto">
        <span class="badge bg-success fs-6 p-2">Sukses: {stats.success || 0}</span>
      </div>
      <div class="col-auto">
        <span class="badge bg-danger fs-6 p-2">Gagal: {stats.failed || 0}</span>
      </div>
    </div>
  {/if}

  {#if loading}
    <div class="text-center py-5">
      <div class="spinner-border text-success" role="status"></div>
      <p class="mt-2 text-secondary small">Memuat...</p>
    </div>
  {:else if error}
    <div class="alert alert-warning">{error}</div>
  {:else}
    <div class="card border shadow-sm">
      <div class="card-body p-0">
        {#if jobs.length === 0}
          <div class="text-center py-4 text-secondary small">Tidak ada job dalam antrian</div>
        {:else}
          <div class="table-responsive">
            <table class="table table-sm mb-0 small">
              <thead class="table-light">
                <tr>
                  <th>#</th>
                  <th>Pegawai (NIP)</th>
                  <th>Tipe</th>
                  <th>Status</th>
                  <th>Waktu</th>
                  <th></th>
                </tr>
              </thead>
              <tbody>
                {#each jobs as job, i}
                  <tr>
                    <td class="text-muted">{i + 1}</td>
                    <td>
                      {#if job.nama}
                        <span class="fw-semibold">{job.nama}</span>
                        {#if job.nip}
                          <small class="text-muted d-block">{job.nip}</small>
                        {/if}
                      {:else}
                        —
                      {/if}
                      {#if (job.status === 'failed' || job.status === 'queued') && job.attempts != null}
                        <small class="text-muted d-block">Percobaan {job.attempts}/{job.max_attempts || 3}</small>
                      {/if}
                    </td>
                    <td><span class="badge bg-light text-dark">{job.run_type}</span></td>
                    <td><span class="badge {statusBadge(job.status)}">{job.status}</span></td>
                    <td class="text-muted">{formatDate(job.created_at)}</td>
                    <td>
                      {#if job.status === 'queued' || job.status === 'running'}
                        <button class="btn btn-outline-danger btn-sm py-0 px-1" onclick={() => cancelJob(job.id)} disabled={busy[job.id]} aria-label="Batalkan job">
                          <i class="bi bi-x-circle"></i>
                        </button>
                      {/if}
                    </td>
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
