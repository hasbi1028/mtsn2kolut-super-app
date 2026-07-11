<script lang="ts">
  import { onMount } from 'svelte';

  interface Role { code: string; name: string; description?: string; }
  interface Permission { code: string; name: string; }

  let roles = $state<Role[]>([]);
  let permissions = $state<Permission[]>([]);
  let loading = $state(true);
  let error = $state('');

  onMount(loadData);

  async function loadData() {
    loading = true; error = '';
    try {
      const [rolesRes, permRes] = await Promise.all([
        fetch('/api/rbac/roles'),
        fetch('/api/rbac/permissions')
      ]);
      if (rolesRes.ok) {
        const d = await rolesRes.json();
        roles = d.roles || d.data || [];
      }
      if (permRes.ok) {
        const d = await permRes.json();
        permissions = d.permissions || d.data || [];
      }
    } catch { error = 'Gagal memuat data RBAC'; }
    finally { loading = false; }
  }
</script>

<svelte:head><title>Hak Akses — MTsN 2 Kolut</title></svelte:head>

<div class="container-fluid px-0">
  <div class="d-flex align-items-center justify-content-between mb-1">
    <h4 class="fw-black mb-0">Hak Akses (RBAC)</h4>
    <button class="btn btn-outline-success btn-sm" onclick={loadData}>
      <i class="bi bi-arrow-repeat me-1"></i>Refresh
    </button>
  </div>
  <p class="text-secondary mb-3" style="font-size:0.85rem;">Role dan permission sistem</p>

  {#if loading}
    <div class="text-center py-5">
      <div class="spinner-border text-success" role="status"></div>
    </div>
  {:else if error}
    <div class="alert alert-warning">{error}</div>
  {:else}
    <div class="row g-4">
      <!-- Roles -->
      <div class="col-md-6">
        <div class="card border shadow-sm h-100">
          <div class="card-header bg-white fw-bold small py-2">
            <i class="bi bi-shield me-1"></i>Roles ({roles.length})
          </div>
          <div class="card-body p-0">
            {#if roles.length === 0}
              <div class="text-center py-4 text-secondary small">Belum ada role</div>
            {:else}
              <div class="table-responsive">
                <table class="table table-sm mb-0 small">
                  <thead class="table-light">
                    <tr><th>Kode</th><th>Nama</th><th>Deskripsi</th></tr>
                  </thead>
                  <tbody>
                    {#each roles as r}
                      <tr>
                        <td><span class="badge bg-success">{r.code}</span></td>
                        <td class="fw-semibold">{r.name}</td>
                        <td class="text-muted">{r.description || '—'}</td>
                      </tr>
                    {/each}
                  </tbody>
                </table>
              </div>
            {/if}
          </div>
        </div>
      </div>

      <!-- Permissions -->
      <div class="col-md-6">
        <div class="card border shadow-sm h-100">
          <div class="card-header bg-white fw-bold small py-2">
            <i class="bi bi-key me-1"></i>Permissions ({permissions.length})
          </div>
          <div class="card-body p-0">
            {#if permissions.length === 0}
              <div class="text-center py-4 text-secondary small">Belum ada permission</div>
            {:else}
              <div class="table-responsive">
                <table class="table table-sm mb-0 small">
                  <thead class="table-light">
                    <tr><th>Kode</th><th>Nama</th></tr>
                  </thead>
                  <tbody>
                    {#each permissions as p}
                      <tr>
                        <td><span class="badge bg-secondary">{p.code}</span></td>
                        <td>{p.name}</td>
                      </tr>
                    {/each}
                  </tbody>
                </table>
              </div>
            {/if}
          </div>
        </div>
      </div>
    </div>
  {/if}
</div>
