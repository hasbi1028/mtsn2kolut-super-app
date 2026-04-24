<script>
  let { employees, onrun, onstop, ondelete } = $props();

  let confirmId = $state(null);
  let busyId    = $state(null);

  async function doStop(id) {
    busyId = id;
    const res  = await fetch('/api/jobs/cancel', {
      method: 'POST',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify({ employee_id: id })
    });
    const data = await res.json().catch(() => ({}));
    busyId = null;
    onstop?.(id, data.cancelled ?? 0);
  }

  async function doDelete(id) {
    busyId = id;
    await fetch('/api/employees', {
      method: 'DELETE',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify({ id })
    });
    busyId    = null;
    confirmId = null;
    ondelete?.();
  }

  function statusLabel(e) {
    const s = e.active_status || e.last_status;
    const t = e.active_run_type || e.last_run_type;
    if (!s) return null;
    const tipe = t === 'morning' ? 'pagi' : t === 'afternoon' ? 'sore' : t;
    return { status: s, tipe };
  }

  function pillClass(status) {
    if (status === 'running') return 'pill-run';
    if (status === 'queued')  return 'pill-queue';
    if (status === 'success') return 'pill-ok';
    if (status === 'failed')  return 'pill-bad';
    return 'pill-idle';
  }

  function pillIcon(status) {
    if (status === 'running') return '⟳';
    if (status === 'queued')  return '…';
    if (status === 'success') return '✓';
    if (status === 'failed')  return '✕';
    return '';
  }
</script>

<section class="card">
  <h2>Daftar Pegawai <span class="badge">{employees.length}</span></h2>
  <div class="stack">
    {#if employees.length === 0}
      <p class="muted">Belum ada data pegawai.</p>
    {/if}

    {#each employees as e}
      {@const sl = statusLabel(e)}
      <div class="row" class:row-active={e.active_status === 'running'}>
        <div class="info">
          <div class="name-row">
            <span class="name">{e.nama}</span>
            {#if sl}
              <span class={`spill ${pillClass(sl.status)}`}>
                {pillIcon(sl.status)} {sl.status}{sl.tipe ? ' · ' + sl.tipe : ''}
              </span>
            {/if}
          </div>
          <div class="meta muted">{e.nip} • {e.unit_kerja}</div>
        </div>

        {#if confirmId === e.id}
          <div class="side">
            <span class="warn-text">Hapus beserta semua job & absensi?</span>
            <button class="btn small danger" onclick={() => doDelete(e.id)} disabled={busyId === e.id}>
              {busyId === e.id ? '...' : 'Ya, Hapus'}
            </button>
            <button class="btn small ghost" onclick={() => (confirmId = null)}>Batal</button>
          </div>
        {:else}
          <div class="side">
            <button class="btn small"       onclick={() => onrun(e.id, 'morning')}   disabled={busyId === e.id}>Pagi</button>
            <button class="btn small ghost" onclick={() => onrun(e.id, 'afternoon')} disabled={busyId === e.id}>Sore</button>
            <button class="btn small stop"  onclick={() => doStop(e.id)}             disabled={busyId === e.id || !e.active_status}>
              {busyId === e.id ? '...' : '■ Stop'}
            </button>
            <button class="btn small ghost danger-outline" onclick={() => (confirmId = e.id)}>Hapus</button>
          </div>
        {/if}
      </div>
    {/each}
  </div>
</section>

<style>
  .row {
    display: flex; justify-content: space-between; align-items: center;
    gap: 10px; padding: 8px 10px;
    border: 1px solid #2f4668; border-radius: 10px;
    background: rgba(13, 21, 34, 0.6);
    flex-wrap: wrap;
    transition: border-color 0.2s;
  }
  .row-active { border-color: rgba(250,180,34,0.4); background: rgba(250,180,34,0.04); }

  .name-row { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
  .name     { font-weight: 600; }
  .meta     { font-size: 0.85rem; margin-top: 2px; }

  /* status pill */
  .spill {
    display: inline-flex; align-items: center; gap: 3px;
    padding: 2px 8px; border-radius: 999px;
    font-size: 0.72rem; font-weight: 600;
    white-space: nowrap;
  }
  .pill-run   { background: rgba(250,180,34,.18);  color: #ffd788; border: 1px solid rgba(250,180,34,.3); }
  .pill-queue { background: rgba(88,166,255,.15);  color: #80b8ff; border: 1px solid rgba(88,166,255,.25); }
  .pill-ok    { background: rgba(31,170,112,.15);  color: #7ff0b7; border: 1px solid rgba(31,170,112,.25); }
  .pill-bad   { background: rgba(225,76,76,.15);   color: #ff9d9d; border: 1px solid rgba(225,76,76,.25); }
  .pill-idle  { background: rgba(130,157,204,.1);  color: #9db2d1; border: 1px solid rgba(130,157,204,.2); }

  .badge {
    background: rgba(88,166,255,0.15); color: #58a6ff;
    font-size: 0.78rem; font-weight: 600;
    padding: 2px 8px; border-radius: 999px;
    vertical-align: middle; margin-left: 6px;
  }

  .side { display: flex; align-items: center; gap: 6px; flex-wrap: wrap; }
  .warn-text { font-size: 0.85rem; color: #ffd788; }

  .btn.stop           { background: rgba(250,180,34,0.12); border: 1px solid #7a6010; color: #ffd788; }
  .btn.stop:hover     { background: rgba(250,180,34,0.22); }
  .btn.stop:disabled  { opacity: 0.3; cursor: not-allowed; }
  .btn.danger         { background: linear-gradient(180deg, #e05252, #c03030); }
  .btn.danger-outline { border-color: #7a3535; color: #ff9d9d; }
  .btn.danger-outline:hover { background: rgba(225,76,76,0.15); }
</style>
