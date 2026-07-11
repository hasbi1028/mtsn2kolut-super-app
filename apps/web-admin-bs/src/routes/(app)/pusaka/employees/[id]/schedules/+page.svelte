<script lang="ts">
  import { page } from '$app/stores';
  import { onMount } from 'svelte';

  interface Schedule {
    id: string;
    employee_id: string;
    run_type: 'checkin' | 'checkout';
    run_time: string;
    is_enabled: boolean;
    day_of_week: number;
    random_window_minutes?: number;
  }

  interface Employee {
    id: string;
    nama: string;
    nip: string;
    unit_kerja?: string;
  }

  let employee = $state<Employee | null>(null);
  let schedules = $state<Schedule[]>([]);
  let loading = $state(true);
  let saving = $state(false);
  let error = $state('');
  let successMsg = $state('');
  let employeeId = $derived($page.params.id || '');

  const DAYS = [
    { idx: 1, label: 'Senin' },
    { idx: 2, label: 'Selasa' },
    { idx: 3, label: 'Rabu' },
    { idx: 4, label: 'Kamis' },
    { idx: 5, label: 'Jumat' },
    { idx: 6, label: 'Sabtu' },
    { idx: 0, label: 'Minggu' }
  ];

  function scheduleFor(day: number, type: 'checkin' | 'checkout'): Schedule | undefined {
    return schedules.find(s => s.day_of_week === day && s.run_type === type);
  }

  async function loadData() {
    loading = true; error = '';
    try {
      const [empRes, schedRes] = await Promise.all([
        fetch(`/api/pusaka/employees/${employeeId}`),
        fetch(`/api/pusaka/employees/${employeeId}/schedules`)
      ]);
      if (empRes.ok) {
        const d = await empRes.json();
        employee = d.data || d.employee || d;
      }
      if (schedRes.ok) {
        const d = await schedRes.json();
        schedules = d.data || d.schedules || d || [];
      } else {
        schedules = [];
      }
    } catch (e: any) {
      error = 'Gagal memuat data jadwal';
    } finally { loading = false; }
  }

  async function saveSchedule(day: number, type: 'checkin' | 'checkout', time: string, enabled: boolean) {
    saving = true; error = ''; successMsg = '';
    try {
      const res = await fetch(`/api/pusaka/employees/${employeeId}/schedules`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          run_type: type,
          run_time: time,
          is_enabled: enabled,
          day_of_week: day,
          random_window_minutes: 10
        })
      });
      if (res.ok) {
        successMsg = 'Jadwal tersimpan';
        await loadData();
        setTimeout(() => { successMsg = ''; }, 2000);
      } else {
        const d = await res.json().catch(() => ({}));
        error = d.error || 'Gagal menyimpan jadwal';
      }
    } catch {
      error = 'Gagal terhubung ke server';
    } finally { saving = false; }
  }

  async function deleteSchedule(id: string) {
    if (!confirm('Hapus jadwal ini?')) return;
    saving = true; error = '';
    try {
      const res = await fetch(`/api/pusaka/employees/${employeeId}/schedules/${id}`, {
        method: 'DELETE'
      });
      if (res.ok) {
        successMsg = 'Jadwal dihapus';
        await loadData();
        setTimeout(() => { successMsg = ''; }, 2000);
      } else {
        error = 'Gagal menghapus jadwal';
      }
    } catch {
      error = 'Gagal terhubung ke server';
    } finally { saving = false; }
  }

  async function copyToAllDays(type: 'checkin' | 'checkout', time: string) {
    if (!confirm(`Set jadwal ${type} ke ${time} untuk semua hari (Senin-Minggu)?`)) return;
    saving = true; error = '';
    try {
      for (const day of DAYS) {
        await fetch(`/api/pusaka/employees/${employeeId}/schedules`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            run_type: type,
            run_time: time,
            is_enabled: true,
            day_of_week: day.idx,
            random_window_minutes: 10
          })
        });
      }
      successMsg = 'Jadwal disalin ke semua hari';
      await loadData();
      setTimeout(() => { successMsg = ''; }, 2000);
    } catch {
      error = 'Gagal menyalin jadwal';
    } finally { saving = false; }
  }

  onMount(loadData);
</script>

<svelte:head><title>Atur Jadwal: {employee?.nama || 'Pegawai'} — MTsN 2 Kolut</title></svelte:head>

<div class="container-fluid px-0">
  <div class="d-flex align-items-center justify-content-between mb-3">
    <div>
      <a href="/pusaka/employees" class="btn btn-link p-0 text-decoration-none small">
        <i class="bi bi-arrow-left me-1"></i>Kembali ke Pegawai
      </a>
      <h4 class="fw-black mt-2 mb-0">Atur Jadwal Auto Absensi</h4>
      {#if employee}
        <p class="text-secondary mb-0 small">
          <span class="fw-semibold">{employee.nama}</span>
          {#if employee.nip}<span class="text-muted ms-2">{employee.nip}</span>{/if}
          {#if employee.unit_kerja}<span class="text-muted ms-2">— {employee.unit_kerja}</span>{/if}
        </p>
      {/if}
    </div>
  </div>

  {#if successMsg}
    <div class="alert alert-success py-2 small">{successMsg}</div>
  {/if}
  {#if error}
    <div class="alert alert-danger py-2 small">{error}</div>
  {/if}

  {#if loading}
    <div class="text-center py-5">
      <div class="spinner-border text-success" role="status"></div>
      <p class="mt-2 text-secondary small">Memuat...</p>
    </div>
  {:else}
    <div class="card border shadow-sm">
      <div class="card-header bg-white d-flex justify-content-between align-items-center py-2">
        <span class="fw-bold small">Jadwal Check-in & Check-out per Hari</span>
        <div class="d-flex gap-2">
          <button class="btn btn-outline-success btn-sm" onclick={() => copyToAllDays('checkin', '07:00')} disabled={saving}>
            <i class="bi bi-copy me-1"></i>Copy 07:00 ke semua hari
          </button>
          <button class="btn btn-outline-warning btn-sm" onclick={() => copyToAllDays('checkout', '15:00')} disabled={saving}>
            <i class="bi bi-copy me-1"></i>Copy 15:00 ke semua hari
          </button>
        </div>
      </div>
      <div class="card-body p-0">
        <div class="table-responsive">
          <table class="table table-sm mb-0 small align-middle">
            <thead class="table-light">
              <tr>
                <th style="width: 18%;">Hari</th>
                <th class="text-center">Check-in</th>
                <th class="text-center">Check-out</th>
                <th class="text-center" style="width: 80px;">Aksi</th>
              </tr>
            </thead>
            <tbody>
              {#each DAYS as day (day.idx)}
                {@const checkin = scheduleFor(day.idx, 'checkin')}
                {@const checkout = scheduleFor(day.idx, 'checkout')}
                <tr>
                  <td class="fw-semibold">{day.label}</td>
                  <td>
                    {#if checkin}
                      <div class="d-flex align-items-center gap-2">
                        <input
                          type="time"
                          class="form-control form-control-sm"
                          style="max-width: 110px;"
                          value={checkin.run_time?.slice(0, 5) || '07:00'}
                          onchange={(e) => saveSchedule(day.idx, 'checkin', e.currentTarget.value, checkin.is_enabled)}
                          disabled={saving}
                        />
                        <div class="form-check form-switch mb-0">
                          <input
                            class="form-check-input"
                            type="checkbox"
                            checked={checkin.is_enabled}
                            onchange={(e) => saveSchedule(day.idx, 'checkin', checkin.run_time, e.currentTarget.checked)}
                            disabled={saving}
                            title="Aktif/Non-aktif"
                          />
                        </div>
                      </div>
                    {:else}
                      <button
                        class="btn btn-outline-success btn-sm py-0"
                        onclick={() => saveSchedule(day.idx, 'checkin', '07:00', true)}
                        disabled={saving}
                      >
                        <i class="bi bi-plus-circle me-1"></i>07:00
                      </button>
                    {/if}
                  </td>
                  <td>
                    {#if checkout}
                      <div class="d-flex align-items-center gap-2">
                        <input
                          type="time"
                          class="form-control form-control-sm"
                          style="max-width: 110px;"
                          value={checkout.run_time?.slice(0, 5) || '15:00'}
                          onchange={(e) => saveSchedule(day.idx, 'checkout', e.currentTarget.value, checkout.is_enabled)}
                          disabled={saving}
                        />
                        <div class="form-check form-switch mb-0">
                          <input
                            class="form-check-input"
                            type="checkbox"
                            checked={checkout.is_enabled}
                            onchange={(e) => saveSchedule(day.idx, 'checkout', checkout.run_time, e.currentTarget.checked)}
                            disabled={saving}
                            title="Aktif/Non-aktif"
                          />
                        </div>
                      </div>
                    {:else}
                      <button
                        class="btn btn-outline-warning btn-sm py-0"
                        onclick={() => saveSchedule(day.idx, 'checkout', '15:00', true)}
                        disabled={saving}
                      >
                        <i class="bi bi-plus-circle me-1"></i>15:00
                      </button>
                    {/if}
                  </td>
                  <td class="text-center">
                    {#if checkin || checkout}
                      <div class="btn-group btn-group-sm">
                        {#if checkin}
                          <button class="btn btn-outline-danger py-0 px-1" onclick={() => deleteSchedule(checkin.id)} disabled={saving} title="Hapus check-in" aria-label="Hapus check-in">
                            <i class="bi bi-trash"></i>
                          </button>
                        {/if}
                        {#if checkout}
                          <button class="btn btn-outline-danger py-0 px-1" onclick={() => deleteSchedule(checkout.id)} disabled={saving} title="Hapus check-out" aria-label="Hapus check-out">
                            <i class="bi bi-trash"></i>
                          </button>
                        {/if}
                      </div>
                    {:else}
                      <span class="text-muted">—</span>
                    {/if}
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      </div>
      <div class="card-footer bg-white small text-muted">
        <i class="bi bi-info-circle me-1"></i>
        Jadwal digunakan oleh Pusaka Worker untuk auto check-in/check-out via SIMPATIKA.
        Random window otomatis 10 menit untuk menghindari deteksi bot.
      </div>
    </div>
  {/if}
</div>
