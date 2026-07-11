<script lang="ts">
  import { onMount } from 'svelte';

  interface LogRow {
    id: string;
    report_date: string;
    target_chat_id_masked: string;
    send_mode: string;
    schedule_time?: string;
    status: string;
    telegram_message_id?: string;
    error_message?: string;
    sent_at: string;
  }

  let laporanList = $state<LogRow[]>([]);
  let loading = $state(true);
  let error = $state('');
  let sending = $state(false);
  let statusMsg = $state('');

  onMount(loadLogs);

  async function loadLogs() {
    loading = true; error = '';
    try {
      const res = await fetch('/api/pusaka/attendance-telegram/logs');
      if (res.ok) {
        const data = await res.json();
        laporanList = data.logs || data.data || [];
      }
    } catch (e: any) {
      error = 'Gagal memuat laporan';
    } finally { loading = false; }
  }

  async function sendLaporan() {
    sending = true; statusMsg = '';
    try {
      const res = await fetch('/api/pusaka/attendance-telegram/send', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({})
      });
      const data = await res.json();
      statusMsg = res.ok
        ? 'Laporan berhasil dikirim!'
        : (data.error || 'Gagal mengirim laporan');
      if (res.ok) { loadLogs(); }
    } catch {
      statusMsg = 'Gagal terhubung ke server';
    } finally { sending = false; }
  }

  function formatDate(ts: string) {
    if (!ts) return '—';
    return new Date(ts).toLocaleString('id-ID', { timeZone: 'Asia/Makassar', dateStyle: 'short', timeStyle: 'short' });
  }
</script>

<svelte:head><title>Laporan Telegram — MTsN 2 Kolut</title></svelte:head>

<div class="container-fluid px-0">
  <h4 class="fw-black mb-1">Laporan Telegram</h4>
  <p class="text-secondary mb-3" style="font-size:0.85rem;">Kirim dan pantau laporan kehadiran via Telegram</p>

  <!-- Send Card -->
  <div class="card border shadow-sm mb-4">
    <div class="card-header bg-white fw-bold small py-2">Kirim Laporan Baru</div>
    <div class="card-body">
      <button class="btn btn-primary btn-sm" onclick={sendLaporan} disabled={sending}>
        {#if sending}
          <span class="spinner-border spinner-border-sm me-1"></span>
          Mengirim...
        {:else}
          <i class="bi bi-send me-1"></i>Kirim Laporan
        {/if}
      </button>
      {#if statusMsg}
        <span class="ms-3 small fw-semibold {statusMsg.includes('berhasil') ? 'text-success' : 'text-danger'}">
          {statusMsg}
        </span>
      {/if}
    </div>
  </div>

  <!-- Logs -->
  <div class="card border shadow-sm">
    <div class="card-header bg-white d-flex justify-content-between align-items-center py-2">
      <span class="fw-bold small">Riwayat Laporan</span>
      <button class="btn btn-outline-secondary btn-sm" onclick={loadLogs}>
        <i class="bi bi-arrow-repeat me-1"></i>Refresh
      </button>
    </div>
    <div class="card-body p-0">
      {#if loading}
        <div class="text-center py-5">
          <div class="spinner-border text-success" role="status"></div>
          <p class="mt-2 text-secondary small">Memuat...</p>
        </div>
      {:else if error}
        <div class="alert alert-warning m-3">{error}</div>
      {:else if laporanList.length === 0}
        <div class="text-center py-4 text-secondary small">
          <i class="bi bi-inbox" style="font-size:2rem;"></i>
          <p class="mt-2">Belum ada laporan terkirim</p>
        </div>
      {:else}
        <div class="table-responsive">
          <table class="table table-sm mb-0 small">
            <thead class="table-light">
              <tr>
                <th>#</th>
                <th>Tanggal</th>
                <th>Waktu</th>
                <th>Status</th>
              </tr>
            </thead>
            <tbody>
              {#each laporanList as log, i}
                <tr>
                  <td class="text-muted">{i + 1}</td>
                  <td>{log.report_date || '—'}</td>
                  <td>{formatDate(log.sent_at)}</td>
                  <td>
                    <span class="badge {log.status === 'success' || log.status === 'terkirim' ? 'bg-success' : log.status === 'failed' || log.status === 'gagal' ? 'bg-danger' : 'bg-secondary'}">
                      {log.status || '—'}
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
</div>
