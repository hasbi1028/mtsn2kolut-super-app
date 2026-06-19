<script lang="ts">
  import { onMount } from 'svelte';

  let user: any = null;
  let loading = $state(true);
  let error = $state('');
  let sessions: any[] = [];
  let loadingSessions = $state(false);

  onMount(async () => {
    try {
      const res = await fetch('/api/auth/account');
      if (res.ok) user = await res.json();
    } catch { error = 'Gagal memuat data akun'; }
    loading = false;
    loadSessions();
  });

  async function loadSessions() {
    loadingSessions = true;
    try {
      const res = await fetch('/api/auth/sessions');
      if (res.ok) {
        const data = await res.json();
        sessions = data.sessions || data.data || [];
      }
    } catch {} finally { loadingSessions = false; }
  }

  async function logoutSession(id: string) {
    try {
      await fetch(`/api/auth/sessions/${id}`, { method: 'DELETE' });
      await loadSessions();
    } catch {}
  }

  function formatDate(ts: string) {
    if (!ts) return '—';
    return new Date(ts).toLocaleString('id-ID', { timeZone: 'Asia/Makassar', dateStyle: 'short', timeStyle: 'short' });
  }
</script>

<svelte:head><title>Akun Saya — MTsN 2 Kolut</title></svelte:head>

<div class="container-fluid px-0">
  <h4 class="fw-black mb-1">Akun Saya</h4>
  <p class="text-secondary mb-3" style="font-size:0.85rem;">Informasi akun dan sesi login</p>

  {#if loading}
    <div class="text-center py-5">
      <div class="spinner-border text-success" role="status"></div>
    </div>
  {:else if error}
    <div class="alert alert-warning">{error}</div>
  {:else}
    <!-- Profile Card -->
    <div class="card border shadow-sm mb-4">
      <div class="card-body">
        <div class="d-flex align-items-center gap-3 mb-3">
          <div class="bg-success text-white rounded-3 d-inline-flex align-items-center justify-content-center" style="width:56px;height:56px;">
            <span class="fw-black fs-4">{user?.username?.substring(0,2).toUpperCase() || '??'}</span>
          </div>
          <div>
            <h5 class="fw-bold mb-0">{user?.username || 'Pengguna'}</h5>
            <span class="badge bg-primary">{user?.role || '—'}</span>
            {#if user?.roles}
              {#each user.roles as r}
                <span class="badge bg-secondary ms-1">{r}</span>
              {/each}
            {/if}
          </div>
        </div>
        <dl class="row small mb-0">
          <dt class="col-sm-3 text-secondary">ID</dt>
          <dd class="col-sm-9">{user?.id || user?.uid || '—'}</dd>
          <dt class="col-sm-3 text-secondary">Username</dt>
          <dd class="col-sm-9">{user?.username || '—'}</dd>
          <dt class="col-sm-3 text-secondary">Role</dt>
          <dd class="col-sm-9">{user?.role || '—'}</dd>
        </dl>
      </div>
    </div>

    <!-- Active Sessions -->
    <div class="card border shadow-sm">
      <div class="card-header bg-white d-flex justify-content-between align-items-center py-2">
        <span class="fw-bold small">Sesi Aktif ({sessions.length})</span>
        <button class="btn btn-outline-secondary btn-sm" onclick={loadSessions}>
          <i class="bi bi-arrow-repeat me-1"></i>Refresh
        </button>
      </div>
      <div class="card-body p-0">
        {#if loadingSessions}
          <div class="text-center py-3">
            <div class="spinner-border spinner-border-sm text-success" role="status"></div>
          </div>
        {:else if sessions.length === 0}
          <div class="text-center py-4 text-secondary small">Tidak ada sesi aktif</div>
        {:else}
          <div class="table-responsive">
            <table class="table table-sm mb-0 small">
              <thead class="table-light">
                <tr>
                  <th>Device</th>
                  <th>IP</th>
                  <th>Login</th>
                  <th>Terakhir Aktif</th>
                  <th></th>
                </tr>
              </thead>
              <tbody>
                {#each sessions as s}
                  <tr>
                    <td class="fw-semibold">{s.user_agent?.slice(0,40) || s.device || '—'}</td>
                    <td class="text-muted">{s.ip_address || s.ip || '—'}</td>
                    <td>{formatDate(s.created_at || s.login_at)}</td>
                    <td>{formatDate(s.last_active || s.last_accessed_at)}</td>
                    <td>
                      <button class="btn btn-outline-danger btn-sm py-0 px-1" onclick={() => logoutSession(s.id)} title="Logout sesi ini">
                        <i class="bi bi-x-circle"></i>
                      </button>
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
