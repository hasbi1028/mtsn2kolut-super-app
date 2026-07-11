<script lang="ts">
  import { onMount } from 'svelte';

  interface Loan {
    id: string;
    book_id: string;
    book_judul?: string;
    book_kode?: string;
    peminjam_nama?: string;
    peminjam_nip?: string;
    tanggal_pinjam: string;
    tanggal_kembali_rencana?: string;
    tanggal_kembali_aktual?: string;
    status: string;
    denda?: number;
    catatan?: string;
  }

  let loans = $state<Loan[]>([]);
  let loading = $state(true);
  let saving = $state(false);
  let error = $state('');
  let successMsg = $state('');
  let showReturnFor = $state<string | null>(null);
  let returnNote = $state('');

  onMount(loadData);

  async function loadData() {
    loading = true; error = '';
    try {
      const res = await fetch('/api/library/loans');
      if (res.ok) {
        const d = await res.json();
        loans = d.data || d.loans || d || [];
      } else {
        error = `Gagal memuat data (HTTP ${res.status})`;
      }
    } catch {
      error = 'Gagal terhubung ke server';
    } finally { loading = false; }
  }

  async function returnLoan(id: string) {
    saving = true; error = ''; successMsg = '';
    try {
      const res = await fetch(`/api/library/loans/${id}/return`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ catatan: returnNote || undefined })
      });
      if (res.ok) {
        successMsg = 'Pengembalian dicatat';
        showReturnFor = null;
        returnNote = '';
        await loadData();
        setTimeout(() => { successMsg = ''; }, 2000);
      } else {
        const d = await res.json().catch(() => ({}));
        error = d.error || 'Gagal mencatat pengembalian';
      }
    } catch {
      error = 'Gagal terhubung ke server';
    } finally { saving = false; }
  }

  function statusBadge(s: string) {
    const m: Record<string, string> = {
      dipinjam: 'bg-warning text-dark',
      kembali: 'bg-success',
      terlambat: 'bg-danger'
    };
    return m[s] || 'bg-secondary';
  }

  function formatDate(ts?: string) {
    if (!ts) return '—';
    return new Date(ts).toLocaleDateString('id-ID', { timeZone: 'Asia/Makassar', dateStyle: 'short' });
  }
</script>

<svelte:head><title>Peminjaman — Perpustakaan MTsN 2 Kolut</title></svelte:head>

<div class="container-fluid px-0">
  <div class="d-flex align-items-center justify-content-between mb-1">
    <h4 class="fw-black mb-0">Peminjaman Buku</h4>
    <div class="d-flex gap-2">
      <a class="btn btn-outline-primary btn-sm" href="/library">
        <i class="bi bi-book me-1"></i>Daftar Buku
      </a>
      <button class="btn btn-outline-success btn-sm" onclick={loadData} disabled={loading}>
        <i class="bi {loading ? 'bi-arrow-repeat spin' : 'bi-arrow-repeat'} me-1"></i>Refresh
      </button>
    </div>
  </div>
  <p class="text-secondary mb-3" style="font-size:0.85rem;">Daftar peminjaman dan pengembalian buku perpustakaan</p>

  {#if successMsg}
    <div class="alert alert-success py-2 small">{successMsg}</div>
  {/if}
  {#if error}
    <div class="alert alert-danger py-2 small">{error}</div>
  {/if}

  {#if loading}
    <div class="text-center py-5">
      <div class="spinner-border text-success" role="status"></div>
      <p class="mt-2 text-secondary small">Memuat...</p>
    </div>
  {:else}
    <div class="card border shadow-sm">
      <div class="card-header bg-white py-2">
        <span class="fw-bold small">{loans.length} peminjaman</span>
      </div>
      <div class="card-body p-0">
        {#if loans.length === 0}
          <div class="text-center py-4 text-secondary small">
            Belum ada data peminjaman.
          </div>
        {:else}
          <div class="table-responsive">
            <table class="table table-sm mb-0 small align-middle">
              <thead class="table-light">
                <tr>
                  <th>#</th>
                  <th>Buku</th>
                  <th>Peminjam</th>
                  <th>Tgl Pinjam</th>
                  <th>Jatuh Tempo</th>
                  <th>Tgl Kembali</th>
                  <th>Status</th>
                  <th></th>
                </tr>
              </thead>
              <tbody>
                {#each loans as loan, i (loan.id)}
                  <tr>
                    <td class="text-muted">{i + 1}</td>
                    <td>
                      <div class="fw-semibold">{loan.book_judul || '—'}</div>
                      {#if loan.book_kode}
                        <small class="text-muted" style="font-family:monospace;">{loan.book_kode}</small>
                      {/if}
                    </td>
                    <td>
                      <div class="fw-semibold">{loan.peminjam_nama || '—'}</div>
                      {#if loan.peminjam_nip}
                        <small class="text-muted">{loan.peminjam_nip}</small>
                      {/if}
                    </td>
                    <td class="text-muted">{formatDate(loan.tanggal_pinjam)}</td>
                    <td class="text-muted">{formatDate(loan.tanggal_kembali_rencana)}</td>
                    <td class="text-muted">{formatDate(loan.tanggal_kembali_aktual)}</td>
                    <td>
                      <span class="badge {statusBadge(loan.status)}">{loan.status || '—'}</span>
                      {#if loan.denda && loan.denda > 0}
                        <div class="small text-danger mt-1">Denda: Rp {loan.denda.toLocaleString('id-ID')}</div>
                      {/if}
                    </td>
                    <td>
                      {#if loan.status === 'dipinjam' || loan.status === 'terlambat'}
                        {#if showReturnFor === loan.id}
                          <div class="d-flex gap-1">
                            <input
                              class="form-control form-control-sm"
                              style="max-width:140px;"
                              placeholder="Catatan (opsional)"
                              bind:value={returnNote}
                            />
                            <button class="btn btn-success btn-sm py-0" onclick={() => returnLoan(loan.id)} disabled={saving}>
                              <i class="bi bi-check"></i>
                            </button>
                            <button class="btn btn-outline-secondary btn-sm py-0" onclick={() => { showReturnFor = null; returnNote = ''; }}>
                              <i class="bi bi-x"></i>
                            </button>
                          </div>
                        {:else}
                          <button class="btn btn-outline-success btn-sm py-0 px-2" onclick={() => showReturnFor = loan.id}>
                            <i class="bi bi-arrow-return-left me-1"></i>Kembalikan
                          </button>
                        {/if}
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
