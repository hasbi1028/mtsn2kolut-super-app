<script lang="ts">
  import { onMount } from 'svelte';

  interface Employee {
    id: string;
    nama?: string; employee_nama?: string;
    nip?: string; employee_nip?: string;
    jabatan?: string; golongan?: string;
    status?: string; unit?: string;
    [key: string]: any;
  }

  let employees = $state<Employee[]>([]);
  let loading = $state(true);
  let error = $state('');
  let search = $state('');

  onMount(loadEmployees);

  async function loadEmployees() {
    loading = true; error = '';
    try {
      const res = await fetch('/api/pusaka/employees');
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const data = await res.json();
      employees = data.employees || data.data || data || [];
    } catch (e: any) {
      error = 'Gagal memuat data pegawai';
    } finally { loading = false; }
  }

  let filtered = $derived(search
    ? employees.filter(e =>
        (e.nama || '').toLowerCase().includes(search.toLowerCase()) ||
        (e.nip || '').includes(search))
    : employees);
</script>

<svelte:head><title>Pegawai PUSAKA — MTsN 2 Kolut</title></svelte:head>

<div class="container-fluid px-0">
  <div class="d-flex align-items-center justify-content-between mb-1">
    <h4 class="fw-black mb-0">Pegawai PUSAKA</h4>
    <button class="btn btn-outline-success btn-sm" onclick={loadEmployees}>
      <i class="bi bi-arrow-repeat me-1"></i>Refresh
    </button>
  </div>
  <p class="text-secondary mb-3" style="font-size:0.85rem;">Daftar pegawai yang terdaftar di PUSAKA</p>

  <!-- Search -->
  <div class="mb-3">
    <div class="input-group input-group-sm" style="max-width:320px;">
      <span class="input-group-text"><i class="bi bi-search"></i></span>
      <input type="text" class="form-control" placeholder="Cari nama atau NIP..." bind:value={search} />
    </div>
  </div>

  {#if loading}
    <div class="text-center py-5">
      <div class="spinner-border text-success" role="status"></div>
      <p class="mt-2 text-secondary small">Memuat data...</p>
    </div>
  {:else if error}
    <div class="alert alert-warning d-flex align-items-center gap-2">
      <i class="bi bi-exclamation-triangle"></i>
      <span>{error}</span>
    </div>
  {:else}
    <div class="card border shadow-sm">
      <div class="card-header bg-white d-flex justify-content-between align-items-center py-2">
        <span class="fw-bold small">{filtered.length} pegawai</span>
      </div>
      <div class="card-body p-0">
        {#if filtered.length === 0}
          <div class="text-center py-4 text-secondary small">
            {#if search}
              Tidak ada pegawai dengan kata kunci "{search}"
            {:else}
              Belum ada data pegawai
            {/if}
          </div>
        {:else}
          <div class="table-responsive">
            <table class="table table-sm mb-0 small">
              <thead class="table-light">
                <tr>
                  <th>#</th>
                  <th>Nama</th>
                  <th>NIP</th>
                  <th>Jabatan</th>
                  <th>Golongan</th>
                  <th>Status</th>
                </tr>
              </thead>
              <tbody>
                {#each filtered as emp, i}
                  <tr>
                    <td class="text-muted">{i + 1}</td>
                    <td class="fw-semibold">{emp.nama || emp.employee_nama || '—'}</td>
                    <td class="text-muted" style="font-family:monospace;font-size:0.75rem;">{emp.nip || emp.employee_nip || '—'}</td>
                    <td>{emp.jabatan || '—'}</td>
                    <td>{emp.golongan || '—'}</td>
                    <td>
                      {#if emp.status}
                        <span class="badge {emp.status === 'aktif' ? 'bg-success' : 'bg-secondary'}">{emp.status}</span>
                      {:else}—{/if}
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
