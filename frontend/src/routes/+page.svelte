<script>
  import { onMount } from 'svelte';
  import QueueMonitor from '$lib/components/QueueMonitor.svelte';

  let queueStats  = $state({ queued: 0, running: 0, success: 0, failed: 0, retry_due: 0, total: 0 });
  let recentJobs  = $state([]);
  let busy        = $state({});
  let toast       = $state({ msg: '', type: 'ok' });
  let confirmKey  = $state('');

  async function load() {
    const [q, j] = await Promise.all([
      fetch('/api/queue/stats').then((r) => r.json()),
      fetch('/api/jobs?limit=5').then((r) => r.json())
    ]);
    queueStats = q;
    recentJobs = j.items;
  }

  async function act(key, fn, successMsg) {
    busy = { ...busy, [key]: true };
    try {
      const res  = await fn();
      const data = await res.json().catch(() => ({}));
      if (!res.ok) throw new Error(data.error || 'Gagal');
      showToast(successMsg + (data.cancelled != null ? ` (${data.cancelled} job)` : ''), 'ok');
    } catch (e) {
      showToast(e.message, 'err');
    } finally {
      busy = { ...busy, [key]: false };
      confirmKey = '';
      await load();
    }
  }

  const runAll        = (t)  => act(`run_${t}`,    () => fetch('/api/jobs/run-all',    { method:'POST', headers:{'content-type':'application/json'}, body: JSON.stringify({ run_type: t }) }), `Job ${t === 'morning' ? 'pagi' : 'sore'} di-queue`);
  const triggerSched  = ()   => act('sched',        () => fetch('/api/scheduler/tick',  { method:'POST' }), 'Scheduler tick dijalankan');
  const cancelAll     = ()   => act('cancel_all',   () => fetch('/api/jobs/cancel-all', { method:'POST' }), 'Semua antrian dibatalkan');
  const restartWorker = ()   => act('restart',      () => fetch('/api/worker/restart',  { method:'POST' }), 'Worker berhasil di-restart');

  function showToast(msg, type = 'ok') {
    toast = { msg, type };
    setTimeout(() => (toast = { msg: '', type: 'ok' }), 3500);
  }

  function statusClass(s) {
    if (s === 'success') return 'ok';
    if (s === 'failed')  return 'bad';
    if (s === 'running') return 'run';
    return 'idle';
  }

  onMount(() => {
    load();
    const itv = setInterval(load, 5000);
    return () => clearInterval(itv);
  });
</script>

<header class="card hero">
  <div>
    <h1>Pusaka Worker Manager</h1>
    <p>Dashboard monitoring scraping absensi pegawai</p>
  </div>
  <div class="btn-group">
    <button class="btn ghost"  onclick={triggerSched}          disabled={busy.sched}>
      {busy.sched ? '...' : '⚡ Trigger Scheduler'}
    </button>
    <button class="btn"        onclick={() => runAll('morning')} disabled={busy.run_morning}>
      {busy.run_morning ? '...' : '▶ Run All Pagi'}
    </button>
    <button class="btn ghost"  onclick={() => runAll('afternoon')} disabled={busy.run_afternoon}>
      {busy.run_afternoon ? '...' : '▶ Run All Sore'}
    </button>

    <div class="divider"></div>

    {#if confirmKey === 'cancel_all'}
      <span class="warn-inline">Batalkan semua antrian?</span>
      <button class="btn small danger" onclick={cancelAll}>Ya</button>
      <button class="btn small ghost"  onclick={() => (confirmKey = '')}>Tidak</button>
    {:else}
      <button class="btn ghost danger-outline" onclick={() => (confirmKey = 'cancel_all')} disabled={busy.cancel_all}>
        ✕ Cancel All Queue
      </button>
    {/if}

    {#if confirmKey === 'restart'}
      <span class="warn-inline">Restart worker sekarang?</span>
      <button class="btn small danger" onclick={restartWorker}>Ya</button>
      <button class="btn small ghost"  onclick={() => (confirmKey = '')}>Tidak</button>
    {:else}
      <button class="btn ghost danger-outline" onclick={() => (confirmKey = 'restart')} disabled={busy.restart}>
        ↺ Restart Worker
      </button>
    {/if}
  </div>
</header>

{#if toast.msg}
  <div class="toast {toast.type}">{toast.msg}</div>
{/if}

<QueueMonitor stats={queueStats} />

<section class="card">
  <div class="section-head">
    <h2>Job Terbaru</h2>
    <a href="/jobs" class="see-all">Lihat semua →</a>
  </div>
  <div class="table-wrap">
    <table>
      <thead>
        <tr><th>Waktu</th><th>Nama</th><th>Tipe</th><th>Status</th><th>Worker</th></tr>
      </thead>
      <tbody>
        {#each recentJobs as j}
          <tr>
            <td>{j.created_at_wita || j.created_at}</td>
            <td>{j.nama}</td>
            <td>{j.run_type}</td>
            <td><span class={`pill ${statusClass(j.status)}`}>{j.status}</span></td>
            <td class="muted">{j.claimed_by || '-'}</td>
          </tr>
        {/each}
        {#if recentJobs.length === 0}
          <tr><td colspan="5" class="muted">Belum ada job.</td></tr>
        {/if}
      </tbody>
    </table>
  </div>
</section>

<style>
  .hero {
    display: flex; justify-content: space-between;
    align-items: flex-start; flex-wrap: wrap; gap: 12px;
  }
  .btn-group {
    display: flex; flex-wrap: wrap; gap: 6px; align-items: center;
  }
  .divider {
    width: 1px; height: 24px;
    background: rgba(130,157,204,0.2);
    margin: 0 2px;
  }
  .warn-inline { font-size: 0.85rem; color: #ffd788; white-space: nowrap; }

  .btn.danger         { background: linear-gradient(180deg, #e05252, #c03030); }
  .btn.danger-outline { border-color: #7a3535; color: #ff9d9d; }
  .btn.danger-outline:hover { background: rgba(225,76,76,0.12); }

  .section-head {
    display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px;
  }
  .section-head h2 { margin: 0; }
  .see-all { color: #58a6ff; font-size: 0.9rem; text-decoration: none; }
  .see-all:hover { text-decoration: underline; }

  .toast {
    border-radius: 10px; padding: 10px 16px; font-size: 0.9rem;
  }
  .toast.ok  { background: rgba(31,170,112,0.2); border: 1px solid rgba(31,170,112,0.4); color: #7ff0b7; }
  .toast.err { background: rgba(225,76,76,0.2);  border: 1px solid rgba(225,76,76,0.4);  color: #ff9d9d; }
</style>
