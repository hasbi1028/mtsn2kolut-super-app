<script lang="ts">
  import { onMount } from 'svelte';

  interface Backup {
    id: string; filename: string; name?: string;
    size_bytes?: number;
    created_at: string; status: string; type?: string;
  }

  interface BackupStatus {
    last_backup?: Backup;
    total_backups?: number; total_size_bytes?: number;
    next_scheduled?: string; is_running?: boolean;
  }

  let status_data = $state<BackupStatus | null>(null);
  let backups = $state<Backup[]>([]);
  let loading = $state(true);
  let error = $state('');

  onMount(loadData);

  async function loadData() {
    loading = true; error = '';
    try {
      const [statusRes, listRes] = await Promise.all([
        fetch('/api/system/backups/status'),
        fetch('/api/system/backups')
      ]);
      if (statusRes.ok) status_data = await statusRes.json();
      if (listRes.ok) {
        const d = await listRes.json();
        backups = d.backups || d.data || [];
      }
    } catch { error = 'Gagal memuat data backup'; }
    finally { loading = false; }
  }

  function formatSize(bytes?: number) {
    if (!bytes) return '—';
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1024*1024) return (bytes/1024).toFixed(1) + ' KB';
    return (bytes/(1024*1024)).toFixed(2) + ' MB';
  }

  function formatDate(ts: string) {
    if (!ts) return '—';
    return new Date(ts).toLocaleString('id-ID', { timeZone: 'Asia/Makassar', dateStyle: 'short', timeStyle: 'short' });
  }
</script>

<svelte:head><title>Backup Data — MTsN 2 Kolut</title></svelte:head>

<div class="container-fluid px-0">
  <div class="d-flex align-items-center justify-content-between mb-1">
    <h4 class="fw-black mb-0">Backup Data</h4>
    <button class="btn btn-outline-success btn-sm" onclick={loadData}>
      <i class="bi bi-arrow-repeat me-1"></i>Refresh
    </button>
  </div>
  <p class="text-secondary mb-3" style="font-size:0.85rem;">Cadangan database dan file sistem</p>

  {#if loading}
    <div class="text-center py-5">
      <div class="spinner-border text-success" role="status"></div>
    </div>
  {:else if error}
    <div class="alert alert-warning">{error}</div>
  {:else}
    <!-- Status overview -->
    {#if status_data}
      <div class="row g-3 mb-4">
        <div class="col-sm-3">
          <div class="stat-card bg-white text-center">
            <div class="fw-black fs-4 text-success">{status_data.total_backups ?? 0}</div>
            <div class="text-secondary small">Total Backup</div>
          </div>
        </div>
        <div class="col-sm-3">
          <div class="stat-card bg-white text-center">
            <div class="fw-black fs-4 text-primary">{formatSize(status_data.total_size_bytes)}</div>
            <div class="text-secondary small">Total Ukuran</div>
          </div>
        </div>
        <div class="col-sm-3">
          <div class="stat-card bg-white text-center">
            <div class="fw-black fs-4 text-warning">{status_data.is_running ? 'Berjalan' : 'Idle'}</div>
            <div class="text-secondary small">Status</div>
          </div>
        </div>
        <div class="col-sm-3">
          <div class="stat-card bg-white text-center">
            <div class="fw-black fs-4 text-info">{status_data.last_backup ? formatDate(status_data.last_backup.created_at) : '—'}</div>
            <div class="text-secondary small">Backup Terakhir</div>
          </div>
        </div>
      </div>
    {/if}

    <!-- Backup List -->
    <div class="card border shadow-sm">
      <div class="card-header bg-white fw-bold small py-2">Riwayat Backup</div>
      <div class="card-body p-0">
        {#if backups.length === 0}
          <div class="text-center py-4 text-secondary small">Belum ada backup</div>
        {:else}
          <div class="table-responsive">
            <table class="table table-sm mb-0 small">
              <thead class="table-light">
                <tr>
                  <th>#</th>
                  <th>File</th>
                  <th>Tipe</th>
                  <th>Ukuran</th>
                  <th>Status</th>
                  <th>Tanggal</th>
                </tr>
              </thead>
              <tbody>
                {#each backups as b, i}
                  <tr>
                    <td class="text-muted">{i + 1}</td>
                    <td class="fw-semibold" style="font-family:monospace;font-size:0.75rem;">{b.filename || b.name || '—'}</td>
                    <td><span class="badge bg-light text-dark">{b.type || '—'}</span></td>
                    <td>{formatSize(b.size_bytes)}</td>
                    <td>
                      <span class="badge {b.status === 'success' || b.status === 'completed' ? 'bg-success' : b.status === 'running' ? 'bg-primary' : b.status === 'failed' ? 'bg-danger' : 'bg-secondary'}">
                        {b.status || '—'}
                      </span>
                    </td>
                    <td class="text-muted">{formatDate(b.created_at)}</td>
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
