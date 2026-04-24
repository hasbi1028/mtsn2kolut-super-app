<script>
  import { onMount } from 'svelte';

  let records    = $state([]);
  let filterDate = $state('');
  let limit      = $state(100);

  async function load() {
    const params = new URLSearchParams({ limit: String(limit) });
    if (filterDate) params.set('date', filterDate);
    const res  = await fetch(`/api/attendance?${params}`);
    const data = await res.json();
    records = data.items;
  }

  function todayWita() {
    return new Intl.DateTimeFormat('en-CA', {
      timeZone: 'Asia/Makassar',
      year: 'numeric', month: '2-digit', day: '2-digit'
    }).format(new Date());
  }

  function stripWita(val) {
    return val ? val.replace(/\s*WITA$/i, '') : '–';
  }

  onMount(() => {
    filterDate = todayWita();
    load();
  });
</script>

<svelte:head><title>Absensi — Pusaka Worker</title></svelte:head>

<div class="bar">
  <span class="title">
    Absensi
    {#if records.length > 0}<span class="count">{records.length}</span>{/if}
  </span>
  <div class="controls">
    <input type="date" bind:value={filterDate} onchange={load} />
    <button class="btn-sm" onclick={load}>↺</button>
  </div>
</div>

<div class="tbl-wrap">
  <table>
    <thead>
      <tr>
        <th class="c-no">#</th>
        <th>Nama</th>
        <th class="c-time">Masuk</th>
        <th class="c-time">Pulang</th>
      </tr>
    </thead>
    <tbody>
      {#each records as a, i}
        <tr class={!a.jam_masuk ? 'row-empty' : ''}>
          <td class="c-no dim">{i + 1}</td>
          <td class="c-nama">{a.nama}</td>
          <td class="c-time {a.jam_masuk ? 'ok' : 'dim'}">{stripWita(a.jam_masuk)}</td>
          <td class="c-time {a.jam_pulang ? 'ok' : 'dim'}">{stripWita(a.jam_pulang)}</td>
        </tr>
      {/each}
      {#if records.length === 0}
        <tr><td colspan="4" class="empty">Tidak ada data.</td></tr>
      {/if}
    </tbody>
  </table>
</div>

<style>
  .bar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 3px;
  }
  .title {
    font-size: 0.78rem;
    font-weight: 700;
    color: #cbe0ff;
  }
  .count {
    margin-left: 4px;
    font-size: 0.65rem;
    color: #58a6ff;
    background: rgba(88,166,255,0.12);
    padding: 0 5px;
    border-radius: 999px;
  }
  .controls { display: flex; gap: 3px; align-items: center; }
  .controls input {
    padding: 1px 4px;
    font-size: 0.65rem;
    border-radius: 5px;
    background: #0d1626;
    color: #e8eefc;
    border: 1px solid #2b3d5a;
    outline: none;
    height: 18px;
    width: auto;
  }
  .btn-sm {
    padding: 0 6px;
    height: 18px;
    font-size: 0.65rem;
    background: transparent;
    border: 1px solid #3f5f8a;
    border-radius: 5px;
    color: #9db2d1;
    cursor: pointer;
  }
  .btn-sm:hover { background: rgba(88,166,255,0.1); color: #e7edf7; }

  .tbl-wrap {
    background: rgba(12,18,30,0.9);
    border: 1px solid rgba(130,157,204,0.18);
    border-radius: 8px;
    overflow: hidden;
  }

  table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.72rem;
    line-height: 1;
  }

  thead tr { background: rgba(8,13,24,0.85); }
  th {
    padding: 3px 5px;
    text-align: left;
    color: #4e6d96;
    font-weight: 600;
    font-size: 0.58rem;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    border-bottom: 1px solid rgba(130,157,204,0.12);
    white-space: nowrap;
  }

  td {
    padding: 1px 5px;
    border-bottom: 1px solid rgba(30,46,70,0.4);
    white-space: nowrap;
    height: 16px;
  }
  tbody tr:last-child td { border-bottom: none; }
  .row-empty td { opacity: 0.5; }

  .c-no   { width: 22px; text-align: right; padding-right: 4px; }
  .c-nama { font-weight: 500; max-width: 170px; overflow: hidden; text-overflow: ellipsis; }
  .c-time { width: 68px; font-variant-numeric: tabular-nums; text-align: center; }

  .ok  { color: #6ee9b0; font-weight: 600; }
  .dim { color: #4e6d96; }

  .empty { padding: 12px; text-align: center; color: #4e6d96; }
</style>
