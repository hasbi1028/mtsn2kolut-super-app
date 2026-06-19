<script lang="ts">
  import { onMount } from 'svelte';

  interface EmployeeSummary {
    employee_id: string; employee_nama: string; employee_nip: string;
    total_days: number; complete_days: number;
    missing_checkout: number; missing_checkin: number;
  }

  interface SummaryItem {
    jumlah_hari: number; hadir: number;
    sakit: number; izin: number; alpa: number;
    tanpa_keterangan: number;
  }

  let employees = $state<EmployeeSummary[]>([]);
  let summary = $state<SummaryItem | null>(null);
  let bulan = $state(String(new Date().getMonth() + 1).padStart(2, '0'));
  let tahun = $state(String(new Date().getFullYear()));
  let loading = $state(true);
  let error = $state('');

  const bulanList = [
    { value: '01', label: 'Januari' }, { value: '02', label: 'Februari' },
    { value: '03', label: 'Maret' }, { value: '04', label: 'April' },
    { value: '05', label: 'Mei' }, { value: '06', label: 'Juni' },
    { value: '07', label: 'Juli' }, { value: '08', label: 'Agustus' },
    { value: '09', label: 'September' }, { value: '10', label: 'Oktober' },
    { value: '11', label: 'November' }, { value: '12', label: 'Desember' }
  ];

  onMount(loadData);

  async function loadData() {
    loading = true; error = '';
    try {
      const startDate = `${tahun}-${bulan}-01`;
      const lastDay = new Date(parseInt(tahun), parseInt(bulan), 0).getDate();
      const endDate = `${tahun}-${bulan}-${String(lastDay).padStart(2, '0')}`;
      const res = await fetch(`/api/pusaka/attendance/summary?start_date=${startDate}&end_date=${endDate}`);
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const json = await res.json();
      const rows: EmployeeSummary[] = json?.data ?? json ?? [];
      employees = rows;

      const maxDays = rows.reduce((m, r) => Math.max(m, r.total_days), 0);
      const hadir = rows.reduce((s, r) => s + r.complete_days, 0);
      const missingCheckout = rows.reduce((s, r) => s + r.missing_checkout, 0);
      const missingCheckin = rows.reduce((s, r) => s + r.missing_checkin, 0);
      const tidakHadir = rows.reduce((s, r) => s + Math.max(0, r.total_days - r.complete_days - r.missing_checkout - r.missing_checkin), 0);

      summary = {
        jumlah_hari: maxDays,
        hadir,
        sakit: 0,
        izin: 0,
        alpa: tidakHadir,
        tanpa_keterangan: missingCheckout + missingCheckin,
      };
    } catch (e: any) {
      error = 'Gagal memuat ringkasan';
    } finally { loading = false; }
  }
</script>

<svelte:head><title>Ringkasan Kehadiran — MTsN 2 Kolut</title></svelte:head>

<div class="container-fluid px-0">
  <h4 class="fw-black mb-1">Ringkasan Kehadiran</h4>
  <p class="text-secondary mb-3" style="font-size:0.85rem;">Akumulasi kehadiran per periode</p>

  <!-- Filter -->
  <div class="card border shadow-sm mb-4">
    <div class="card-body py-3">
      <div class="row g-2 align-items-end">
        <div class="col-sm-4">
          <label class="form-label small fw-semibold mb-1">Bulan</label>
          <select class="form-select form-select-sm" bind:value={bulan}>
            {#each bulanList as b}
              <option value={b.value}>{b.label}</option>
            {/each}
          </select>
        </div>
        <div class="col-sm-3">
          <label class="form-label small fw-semibold mb-1">Tahun</label>
          <input type="number" class="form-control form-control-sm" bind:value={tahun} min="2020" max="2035" />
        </div>
        <div class="col-sm-3 d-grid">
          <button class="btn btn-primary btn-sm" onclick={loadData} disabled={loading}>
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
      <button class="btn btn-sm btn-outline-secondary ms-auto" onclick={loadData}>
        <i class="bi bi-arrow-repeat me-1"></i>Ulang
      </button>
    </div>
  {:else if summary}
    <!-- Summary Stats -->
    <div class="row g-3 mb-4">
      <div class="col-4 col-lg-2">
        <div class="stat-card bg-white text-center">
          <div class="fw-black fs-4 text-primary">{summary.jumlah_hari || 0}</div>
          <div class="text-secondary small">Hari Kerja</div>
        </div>
      </div>
      <div class="col-4 col-lg-2">
        <div class="stat-card bg-white text-center">
          <div class="fw-black fs-4 text-success">{summary.hadir || 0}</div>
          <div class="text-secondary small">Hadir</div>
        </div>
      </div>
      <div class="col-4 col-lg-2">
        <div class="stat-card bg-white text-center">
          <div class="fw-black fs-4 text-warning">{summary.sakit || 0}</div>
          <div class="text-secondary small">Sakit</div>
        </div>
      </div>
      <div class="col-4 col-lg-2">
        <div class="stat-card bg-white text-center">
          <div class="fw-black fs-4 text-info">{summary.izin || 0}</div>
          <div class="text-secondary small">Izin</div>
        </div>
      </div>
      <div class="col-4 col-lg-2">
        <div class="stat-card bg-white text-center">
          <div class="fw-black fs-4 text-danger">{summary.alpa || 0}</div>
          <div class="text-secondary small">Alpha</div>
        </div>
      </div>
      <div class="col-4 col-lg-2">
        <div class="stat-card bg-white text-center">
          <div class="fw-black fs-4 text-secondary">{summary.tanpa_keterangan || 0}</div>
          <div class="text-secondary small">Tanpa Ket.</div>
        </div>
      </div>
    </div>

    <!-- Alerts -->
    {#if alerts.length > 0}
      <div class="card border shadow-sm mb-4">
        <div class="card-header bg-white fw-bold small py-2">
          <i class="bi bi-exclamation-triangle text-warning me-1"></i>Perhatian
        </div>
        <div class="card-body p-0">
          <div class="table-responsive">
            <table class="table table-sm mb-0 small">
              <thead class="table-light">
                <tr><th>Pegawai</th><th>Masalah</th><th>Periode</th><th>Tingkat</th></tr>
              </thead>
              <tbody>
                {#each alerts as a}
                  <tr>
                    <td class="fw-semibold">{a.employee_nama}</td>
                    <td>{a.masalah}</td>
                    <td>{a.periode}</td>
                    <td>
                      <span class="badge {a.tingkat === 'tinggi' ? 'bg-danger' : a.tingkat === 'sedang' ? 'bg-warning text-dark' : 'bg-secondary'}">
                        {a.tingkat}
                      </span>
                    </td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    {/if}
  {:else}
    <div class="text-center py-5 text-secondary small">
      <i class="bi bi-inbox" style="font-size:2rem;"></i>
      <p class="mt-2">Pilih periode untuk melihat ringkasan</p>
    </div>
  {/if}
</div>
