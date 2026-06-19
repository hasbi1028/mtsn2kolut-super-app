<script lang="ts">
  import { onMount } from 'svelte';

  interface AuditLog {
    id: string; user_id?: string; username?: string;
    action: string; resource: string; details?: string;
    ip_address?: string; created_at: string;
  }

  let logs = $state<AuditLog[]>([]);
  let loading = $state(true);
  let error = $state('');
  let pageNum = $state(1);
  const perPage = 25;

  onMount(loadLogs);

  async function loadLogs() {
    loading = true; error = '';
    try {
      const res = await fetch(`/api/users/audit-logs?limit=${perPage}&offset=${(pageNum - 1) * perPage}`);
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const data = await res.json();
      logs = data.logs || data.data || data.audit_logs || [];
    } catch { error = 'Gagal memuat audit log'; }
    finally { loading = false; }
  }
</script>

<svelte:head><title>Audit Aktivitas — MTsN 2 Kolut</title></svelte:head>

<div class="container-fluid px-0">
  <div class="d-flex align-items-center justify-content-between mb-1">
    <h4 class="fw-black mb-0">Audit Aktivitas</h4>
    <button class="btn btn-outline-success btn-sm" onclick={loadLogs}>
      <i class="bi bi-arrow-repeat me-1"></i>Refresh
    </button>
  </div>
  <p class="text-secondary mb-3" style="font-size:0.85rem;">Jejak aktivitas pengguna dalam sistem</p>

  {#if loading}
    <div class="text-center py-5">
      <div class="spinner-border text-success" role="status"></div>
    </div>
  {:else if error}
    <div class="alert alert-warning">{error}</div>
  {:else}
    <div class="card border shadow-sm">
      <div class="card-body p-0">
        {#if logs.length === 0}
          <div class="text-center py-4 text-secondary small">Belum ada aktivitas tercatat</div>
        {:else}
          <div class="table-responsive">
            <table class="table table-sm mb-0 small">
              <thead class="table-light">
                <tr>
                  <th>#</th>
                  <th>Pengguna</th>
                  <th>Aksi</th>
                  <th>Resource</th>
                  <th>Detail</th>
                  <th>IP</th>
                  <th>Waktu</th>
                </tr>
              </thead>
              <tbody>
                {#each logs as log, i}
                  <tr>
                    <td class="text-muted">{(pageNum - 1) * perPage + i + 1}</td>
                    <td class="fw-semibold">{log.username || log.user_id || '—'}</td>
                    <td><span class="badge bg-light text-dark">{log.action}</span></td>
                    <td class="text-muted" style="font-size:0.7rem;">{log.resource}</td>
                    <td class="text-muted" style="max-width:200px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;">{log.details || '—'}</td>
                    <td class="text-muted" style="font-family:monospace;font-size:0.7rem;">{log.ip_address || '—'}</td>
                    <td class="text-muted">{log.created_at ? new Date(log.created_at).toLocaleString('id-ID', {timeZone:'Asia/Makassar', dateStyle:'short', timeStyle:'short'}) : '—'}</td>
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
