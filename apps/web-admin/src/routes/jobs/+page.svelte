<script>
  import { onMount } from 'svelte';

  let jobs        = $state([]);
  let filterStatus = $state('');
  let limit        = $state(50);
  let autoRefresh  = $state(true);
  let interval;

  async function load() {
    try {
      const params = new URLSearchParams({ limit: String(limit) });
      if (filterStatus) params.set('status', filterStatus);
      const res  = await fetch(`/api/jobs?${params}`);
      const data = await res.json();
      if (data.error) { console.error('[mtsn2kolut] jobs:', data.error); return; }
      jobs = data.items ?? [];
    } catch (e) {
      console.error('[mtsn2kolut] jobs load failed:', e);
    }
  }

  function statusClass(s) {
    if (s === 'success') return 'ok';
    if (s === 'failed')  return 'bad';
    if (s === 'running') return 'run';
    return 'idle';
  }

  $effect(() => {
    clearInterval(interval);
    if (autoRefresh) interval = setInterval(load, 5000);
    return () => clearInterval(interval);
  });

  onMount(() => { load(); });
</script>

<svelte:head><title>Jobs — MTSN 2 Kolut Super App</title></svelte:head>

<div class="toolbar">
  <h2 class="page-title">Riwayat Job</h2>
  <div class="controls">
    <select bind:value={filterStatus} onchange={load}>
      <option value="">Semua Status</option>
      <option value="queued">Queued</option>
      <option value="running">Running</option>
      <option value="success">Success</option>
      <option value="failed">Failed</option>
    </select>
    <select bind:value={limit} onchange={load}>
      <option value={20}>20 baris</option>
      <option value={50}>50 baris</option>
      <option value={100}>100 baris</option>
      <option value={200}>200 baris</option>
    </select>
    <label class="check">
      <input type="checkbox" bind:checked={autoRefresh} />
      Auto-refresh
    </label>
    <button class="btn small" onclick={load}>↺ Refresh</button>
  </div>
</div>

<section class="card">
  <div class="table-wrap">
    <table>
      <thead>
        <tr>
          <th>Waktu</th><th>Nama</th><th>Tipe</th><th>Status</th>
          <th>Attempt</th><th>Worker</th><th>Retry At</th><th>Error</th>
        </tr>
      </thead>
      <tbody>
        {#each jobs as j}
          <tr>
            <td class="nowrap">{j.created_at_wita || j.created_at}</td>
            <td>{j.nama}</td>
            <td>{j.run_type}</td>
            <td><span class={`pill ${statusClass(j.status)}`}>{j.status}</span></td>
            <td>{j.attempts}/{j.max_attempts}</td>
            <td class="muted nowrap">{j.claimed_by || '-'}</td>
            <td class="muted nowrap">{j.next_retry_at || '-'}</td>
            <td class="muted err">{j.error_message || '-'}</td>
          </tr>
        {/each}
        {#if jobs.length === 0}
          <tr><td colspan="8" class="muted">Tidak ada job.</td></tr>
        {/if}
      </tbody>
    </table>
  </div>
</section>

<style>
  .toolbar {
    display: flex; justify-content: space-between;
    align-items: center; flex-wrap: wrap; gap: 12px;
    margin-bottom: 4px;
  }
  .page-title { margin: 0; font-size: 1.4rem; color: #e7edf7; }
  .controls { display: flex; gap: 8px; align-items: center; flex-wrap: wrap; }
  .controls select { width: auto; padding: 7px 10px; }
  .nowrap { white-space: nowrap; }
  .err { max-width: 220px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
</style>
