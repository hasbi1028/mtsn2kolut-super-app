<script lang="ts">
  import { onMount } from 'svelte';

  interface User {
    id: string; username: string; role?: string;
    roles?: string[]; employee_nama?: string;
    status?: string; created_at?: string;
  }

  let users = $state<User[]>([]);
  let loading = $state(true);
  let error = $state('');
  let search = $state('');

  onMount(loadUsers);

  async function loadUsers() {
    loading = true; error = '';
    try {
      const res = await fetch('/api/users');
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const data = await res.json();
      users = data.users || data.data || [];
    } catch (e: any) {
      error = 'Gagal memuat data pengguna';
    } finally { loading = false; }
  }

  let filtered = $derived(search
    ? users.filter(u =>
        (u.username || '').toLowerCase().includes(search.toLowerCase()) ||
        (u.employee_nama || '').toLowerCase().includes(search.toLowerCase()))
    : users);
</script>

<svelte:head><title>Pengguna — MTsN 2 Kolut</title></svelte:head>

<div class="container-fluid px-0">
  <div class="d-flex align-items-center justify-content-between mb-1">
    <h4 class="fw-black mb-0">Manajemen Pengguna</h4>
    <button class="btn btn-outline-success btn-sm" onclick={loadUsers}>
      <i class="bi bi-arrow-repeat me-1"></i>Refresh
    </button>
  </div>
  <p class="text-secondary mb-3" style="font-size:0.85rem;">Kelola akun pengguna sistem</p>

  <div class="mb-3">
    <div class="input-group input-group-sm" style="max-width:320px;">
      <span class="input-group-text"><i class="bi bi-search"></i></span>
      <input type="text" class="form-control" placeholder="Cari username atau nama..." bind:value={search} />
    </div>
  </div>

  {#if loading}
    <div class="text-center py-5">
      <div class="spinner-border text-success" role="status"></div>
    </div>
  {:else if error}
    <div class="alert alert-warning">{error}</div>
  {:else}
    <div class="card border shadow-sm">
      <div class="card-body p-0">
        {#if filtered.length === 0}
          <div class="text-center py-4 text-secondary small">
            {#if search}Tidak ditemukan{:else}Belum ada pengguna{/if}
          </div>
        {:else}
          <div class="table-responsive">
            <table class="table table-sm mb-0 small">
              <thead class="table-light">
                <tr>
                  <th>#</th>
                  <th>Username</th>
                  <th>Nama Pegawai</th>
                  <th>Role</th>
                  <th>Status</th>
                </tr>
              </thead>
              <tbody>
                {#each filtered as u, i}
                  <tr>
                    <td class="text-muted">{i + 1}</td>
                    <td class="fw-semibold">{u.username}</td>
                    <td>{u.employee_nama || '—'}</td>
                    <td>
                      {#if u.roles}
                        {#each u.roles as r}
                          <span class="badge bg-primary me-1">{r}</span>
                        {/each}
                      {:else}
                        <span class="badge bg-secondary">{u.role || '—'}</span>
                      {/if}
                    </td>
                    <td>
                      <span class="badge {u.status === 'active' || u.status === 'aktif' ? 'bg-success' : 'bg-secondary'}">
                        {u.status || '—'}
                      </span>
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
