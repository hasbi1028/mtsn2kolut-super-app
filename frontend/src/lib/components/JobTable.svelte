<script>
  let { jobs } = $props();

  function statusClass(s) {
    if (s === 'success') return 'ok';
    if (s === 'failed') return 'bad';
    if (s === 'running') return 'run';
    return 'idle';
  }
</script>

<section class="card">
  <h2>Job Terbaru</h2>
  <div class="table-wrap">
    <table>
      <thead>
        <tr>
          <th>Waktu</th><th>Nama</th><th>Tipe</th><th>Status</th>
          <th>Attempt</th><th>Worker</th><th>Error</th>
        </tr>
      </thead>
      <tbody>
        {#each jobs as j}
          <tr>
            <td>{j.created_at_wita || j.created_at}</td>
            <td>{j.nama}</td>
            <td>{j.run_type}</td>
            <td><span class={`pill ${statusClass(j.status)}`}>{j.status}</span></td>
            <td>{j.attempts}/{j.max_attempts}</td>
            <td>{j.claimed_by || '-'}</td>
            <td class="muted">{j.error_message || '-'}</td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</section>
