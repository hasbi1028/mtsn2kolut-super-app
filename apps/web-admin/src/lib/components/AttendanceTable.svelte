<script>
  let { records } = $props();
</script>

<section class="card">
  <h2>Absensi Terakhir</h2>
  <div class="table-wrap">
    <table>
      <thead>
        <tr><th>Tanggal</th><th>Pegawai</th><th>Jam Masuk</th><th>Jam Pulang</th></tr>
      </thead>
      <tbody>
        {#each records as a}
          <tr>
            <td>{a.tanggal}</td>
            <td>{a.nama} <span class="muted">({a.nip})</span></td>
            <td>{a.jam_masuk || '-'}</td>
            <td>{a.jam_pulang || '-'}</td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</section>
