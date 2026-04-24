<script>
  let { schedules = $bindable(), onsave } = $props();
</script>

<section class="card">
  <h2>Jadwal Otomatis</h2>
  <div class="stack">
    {#each schedules as s, i}
      <div class="schedule-row">
        <strong>{s.label}</strong>
        <input class="time" bind:value={schedules[i].run_time} />
        <select bind:value={schedules[i].run_type}>
          <option value="morning">morning</option>
          <option value="afternoon">afternoon</option>
        </select>
        <label class="check">
          <input type="checkbox" bind:checked={schedules[i].is_enabled} /> aktif
        </label>
      </div>
    {/each}
  </div>
  <button class="btn" style="margin-top:12px" onclick={onsave}>Simpan Jadwal</button>
</section>

<style>
  .schedule-row {
    display: grid;
    grid-template-columns: 1fr 90px 140px 90px;
    gap: 8px;
    align-items: center;
    padding: 10px;
    border: 1px dashed #38527a;
    border-radius: 10px;
  }
</style>
