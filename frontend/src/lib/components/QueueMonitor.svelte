<script>
  let { stats } = $props();
</script>

<section class="card">
  <h2>Queue Monitor</h2>
  <div class="queue-grid">
    <div class="qbox"><span>Queued</span><strong>{stats.queued}</strong></div>
    <div class="qbox"><span>Running</span><strong>{stats.running}</strong></div>
    <div class="qbox"><span>Success</span><strong>{stats.success}</strong></div>
    <div class="qbox"><span>Failed</span><strong>{stats.failed}</strong></div>
    <div class="qbox"><span>Retry Due</span><strong>{stats.retry_due}</strong></div>
    <div class="qbox"><span>Total</span><strong>{stats.total}</strong></div>
  </div>
</section>

<style>
  .queue-grid {
    display: grid;
    grid-template-columns: repeat(6, minmax(0, 1fr));
    gap: 10px;
  }
  .qbox {
    border: 1px solid #2f4668;
    border-radius: 10px;
    padding: 10px;
    background: rgba(13, 21, 34, 0.6);
    display: grid;
    gap: 4px;
  }
  .qbox span { color: #95a9c7; font-size: 0.85rem; }
  .qbox strong { font-size: 1.2rem; }

  @media (max-width: 900px) {
    .queue-grid { grid-template-columns: 1fr 1fr; }
  }
</style>
