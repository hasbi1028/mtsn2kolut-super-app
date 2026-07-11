<script lang="ts">
  import { onMount } from 'svelte';

  interface Employee {
    id: string;
    nama?: string;
    nip?: string;
    unit_kerja?: string;
    employment_type?: string;
    is_active?: boolean;
    has_pusaka_account?: boolean;
    pusaka_is_enabled?: boolean;
    pusaka_username?: string;
    pegawai_uid?: string;
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
      employees = data.employees || data.data || data.items || [];
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
                  <th>Unit Kerja</th>
                  <th>Status</th>
                  <th>Pusaka</th>
                  <th>Jadwal</th>
                  <th></th>
                </tr>
              </thead>
              <tbody>
                {#each filtered as emp, i}
                  <tr>
                    <td class="text-muted">{i + 1}</td>
                    <td class="fw-semibold">{emp.nama || '—'}</td>
                    <td class="text-muted" style="font-family:monospace;font-size:0.75rem;">{emp.nip || '—'}</td>
                    <td>{emp.unit_kerja || '—'}</td>
                    <td>
                      <span class="badge {emp.is_active ? 'bg-success' : 'bg-secondary'}">
                        {emp.is_active ? 'Aktif' : 'Non-aktif'}
                      </span>
                      <small class="text-muted ms-1">{emp.employment_type?.toUpperCase() || ''}</small>
                    </td>
                    <td>
                      {#if emp.has_pusaka_account}
                        <span class="badge {emp.pusaka_is_enabled ? 'bg-success' : 'bg-warning text-dark'}" title={emp.pusaka_username || ''}>
                          {emp.pusaka_is_enabled ? 'Aktif' : 'Disabled'}
                        </span>
                      {:else}
                        <span class="badge bg-light text-dark">Belum</span>
                      {/if}
                    </td>
                    <td>
                      {#if emp.has_checkin_schedule || emp.has_checkout_schedule}
                        <span class="badge bg-info-subtle text-dark">
                          <i class="bi bi-clock me-1"></i>
                          {emp.has_checkin_schedule ? 'In' : ''}{emp.has_checkin_schedule && emp.has_checkout_schedule ? '+' : ''}{emp.has_checkout_schedule ? 'Out' : ''}
                        </span>
                      {:else}
                        <span class="text-muted">—</span>
                      {/if}
                    </td>
                    <td>
                      <a class="btn btn-outline-primary btn-sm py-0 px-2" href="/pusaka/employees/{emp.id}/schedules" title="Atur jadwal auto absensi" aria-label="Atur jadwal">
                        <i class="bi bi-clock-history"></i>
                      </a>
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
