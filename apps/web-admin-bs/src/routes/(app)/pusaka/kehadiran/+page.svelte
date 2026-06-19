<script lang="ts">
  import { onMount } from 'svelte';

  interface AttendanceRecord {
    id: string; employee_nama: string; employee_nip: string;
    tanggal: string; jam_masuk: string | null; jam_pulang: string | null;
  }

  let records = $state<AttendanceRecord[]>([]);
  let total = $state<number | null>(null);
  let startDate = $state(todayWita());
  let endDate = $state('');
  let loading = $state(true);
  let error = $state('');
  let viewMode = $state<'normal' | 'compact'>('normal');

  function todayWita() {
    return new Intl.DateTimeFormat('en-CA', {
      timeZone: 'Asia/Makassar', year: 'numeric', month: '2-digit', day: '2-digit'
    }).format(new Date());
  }

  function stripWita(val: string | null) {
    if (!val) return '—';
    return val.replace(/\s*WITA$/i, '').trim();
  }

  function attendanceStatus(r: AttendanceRecord) {
    if (r.jam_masuk && r.jam_pulang) return 'lengkap';
    if (r.jam_masuk) return 'masuk';
    return 'belum';
  }

  function statusBadge(status: string) {
    const map: Record<string, string> = {
      lengkap: 'bg-success', masuk: 'bg-primary', belum: 'bg-secondary'
    };
    return map[status] || 'bg-secondary';
  }

  async function loadRecords() {
    const sd = startDate || todayWita();
    const ed = endDate || sd;
    loading = true; error = '';
    try {
      const res = await fetch(`/api/pusaka/attendance?start_date=${sd}&end_date=${ed}`);
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const data = await res.json();
      records = data.records || data.data || [];
      total = data.total || records.length;
    } catch (e: any) {
      error = 'Gagal memuat data kehadiran';
      records = [];
    } finally { loading = false; }
  }

  onMount(loadRecords);
</script>

<svelte:head><title>Data Kehadiran — MTsN 2 Kolut</title></svelte:head>

<div class="container-fluid px-0">
  <div class="d-flex align-items-center justify-content-between mb-1">
    <h4 class="fw-black mb-0">Data Kehadiran</h4>
    <div class="btn-group btn-group-sm">
      <button class="btn {viewMode === 'normal' ? 'btn-primary' : 'btn-outline-primary'}" onclick={() => viewMode = 'normal'}>
        <i class="bi bi-list-ul"></i>
      </button>
      <button class="btn {viewMode === 'compact' ? 'btn-primary' : 'btn-outline-primary'}" onclick={() => viewMode = 'compact'}>
        <i class="bi bi-justify"></i>
      </button>
    </div>
  </div>
  <p class="text-secondary mb-3" style="font-size:0.85rem;">Rekap kehadiran harian dari PUSAKA Kemenag</p>

  <!-- Filter -->
  <div class="card border shadow-sm mb-4">
    <div class="card-body py-3">
      <div class="row g-2 align-items-end">
        <div class="col-sm-4">
          <label class="form-label small fw-semibold mb-1">Tanggal Mulai</label>
          <input type="date" class="form-control form-control-sm" bind:value={startDate} />
        </div>
        <div class="col-sm-4">
          <label class="form-label small fw-semibold mb-1">Tanggal Akhir</label>
          <input type="date" class="form-control form-control-sm" bind:value={endDate} />
        </div>
        <div class="col-sm-4 d-grid">
          <button class="btn btn-primary btn-sm" onclick={loadRecords} disabled={loading}>
            {#if loading}
              <span class="spinner-border spinner-border-sm me-1"></span>
            {:else}
              <i class="bi bi-search me-1"></i>
            {/if}
            Tampilkan
          </button>
        </div>
      </div>
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
      <button class="btn btn-sm btn-outline-secondary ms-auto" onclick={loadRecords}>
        <i class="bi bi-arrow-repeat me-1"></i>Ulang
      </button>
    </div>
  {:else}
    <div class="card border shadow-sm">
      <div class="card-header bg-white d-flex justify-content-between align-items-center py-2">
        <span class="fw-bold small">
          {records.length} data
          {#if total !== null && total !== records.length}
            <span class="text-muted fw-normal">(total {total})</span>
          {/if}
        </span>
        <button class="btn btn-outline-success btn-sm" onclick={loadRecords}>
          <i class="bi bi-arrow-repeat me-1"></i>Refresh
        </button>
      </div>
      <div class="card-body p-0">
        {#if records.length === 0}
          <div class="text-center py-4 text-secondary small">Tidak ada data kehadiran untuk rentang ini</div>
        {:else}
          <div class="table-responsive">
            <table class="table table-sm mb-0 {viewMode === 'compact' ? 'table-sm' : ''}">
              <thead class="table-light">
                <tr>
                  <th>#</th>
                  <th>Nama</th>
                  {#if viewMode === 'normal'}
                    <th>NIP</th>
                  {/if}
                  <th>Tanggal</th>
                  <th>Masuk</th>
                  <th>Pulang</th>
                  <th>Status</th>
                </tr>
              </thead>
              <tbody>
                {#each records as r, i}
                  <tr>
                    <td class="text-muted">{i + 1}</td>
                    <td class="fw-semibold">{r.employee_nama}</td>
                    {#if viewMode === 'normal'}
                      <td class="text-muted" style="font-family:monospace;font-size:0.75rem;">{r.employee_nip || '—'}</td>
                    {/if}
                    <td>{r.tanggal ? new Date(r.tanggal).toLocaleDateString('id-ID', { timeZone: 'Asia/Makassar' }) : '—'}</td>
                    <td class="fw-semibold">{stripWita(r.jam_masuk)}</td>
                    <td>{stripWita(r.jam_pulang)}</td>
                    <td><span class="badge {statusBadge(attendanceStatus(r))}">{attendanceStatus(r)}</span></td>
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
