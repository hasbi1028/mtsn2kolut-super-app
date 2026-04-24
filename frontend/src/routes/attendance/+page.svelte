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
    margin-bottom: 4px;
  }
  .title {
    font-size: 0.82rem;
    font-weight: 700;
    color: #cbe0ff;
  }
  .count {
    margin-left: 5px;
    font-size: 0.7rem;
    color: #58a6ff;
    background: rgba(88,166,255,0.12);
    padding: 0 6px;
    border-radius: 999px;
  }
  .controls { display: flex; gap: 4px; align-items: center; }
  .controls input {
    padding: 2px 6px;
    font-size: 0.72rem;
    border-radius: 6px;
    background: #0d1626;
    color: #e8eefc;
    border: 1px solid #2b3d5a;
    outline: none;
    height: 22px;
    width: auto;
  }
  .btn-sm {
    padding: 0 8px;
    height: 22px;
    font-size: 0.72rem;
    background: transparent;
    border: 1px solid #3f5f8a;
    border-radius: 6px;
    color: #9db2d1;
    cursor: pointer;
  }
  .btn-sm:hover { background: rgba(88,166,255,0.1); color: #e7edf7; }

  .tbl-wrap {
    background: rgba(12,18,30,0.9);
    border: 1px solid rgba(130,157,204,0.18);
    border-radius: 10px;
    overflow: hidden;
  }

  table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.8rem;
    line-height: 1;
  }

  thead tr { background: rgba(8,13,24,0.8); }
  th {
    padding: 5px 8px;
    text-align: left;
    color: #4e6d96;
    font-weight: 600;
    font-size: 0.65rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    border-bottom: 1px solid rgba(130,157,204,0.12);
    white-space: nowrap;
  }

  td {
    padding: 2px 8px;
    border-bottom: 1px solid rgba(30,46,70,0.5);
    white-space: nowrap;
    height: 22px;
  }
  tbody tr:last-child td { border-bottom: none; }
  tbody tr:hover { background: rgba(88,166,255,0.05); }
  .row-empty td { opacity: 0.55; }

  .c-no   { width: 26px; text-align: right; padding-right: 6px; }
  .c-nama { font-weight: 500; }
  .c-time { width: 76px; font-variant-numeric: tabular-nums; text-align: center; }

  .ok  { color: #6ee9b0; font-weight: 600; }
  .dim { color: #4e6d96; }

  .empty { padding: 16px; text-align: center; color: #4e6d96; }

  /* Mobile / screenshot smartphone */
  @media (max-width: 480px) {
    .bar { margin-bottom: 3px; }
    .controls input { font-size: 0.68rem; padding: 1px 4px; height: 20px; }
    .btn-sm { height: 20px; font-size: 0.68rem; }

    table { font-size: 0.75rem; }
    th { font-size: 0.6rem; padding: 4px 6px; }
    td { padding: 2px 6px; height: 20px; }

    .c-no   { width: 20px; }
    .c-time { width: 62px; }
  }
</style>
