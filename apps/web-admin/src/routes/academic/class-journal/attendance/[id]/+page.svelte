<script lang="ts">
  import { Button } from '$lib/components/ui/button';
  import * as Card from '$lib/components/ui/card';
  import { Input } from '$lib/components/ui/input';
  import { toast } from '$lib/components/ui/sonner';
  import LoadingButton from '$lib/components/LoadingButton.svelte';

  let { data } = $props();

  type AttendanceItem = {
    id?: string;
    session_id?: string;
    student_id: string;
    status: string;
    catatan?: string;
    nis: string;
    nisn?: string;
    nama: string;
    gender?: string;
  };

  let session = $state(data.session);
  let attendances = $state<AttendanceItem[]>([]);
  let loading = $state(false);
  let saving = $state(false);
  let searchText = $state('');

  // Stats
  let totalHadir = $derived(attendances.filter(a => a.status === 'hadir').length);
  let totalSakit = $derived(attendances.filter(a => a.status === 'sakit').length);
  let totalIzin = $derived(attendances.filter(a => a.status === 'izin').length);
  let totalAlpha = $derived(attendances.filter(a => a.status === 'alpha').length);
  let totalTercatat = $derived(attendances.length);

  let filteredAttendances = $derived(
    searchText
      ? attendances.filter(a =>
          a.nama.toLowerCase().includes(searchText.toLowerCase()) ||
          a.nis.includes(searchText)
        )
      : attendances
  );

  async function loadAttendances() {
    if (!session) return;
    loading = true;
    try {
      const r = await fetch(`/api/class-journal/sessions/${session.id}/attendances`);
      if (r.ok) {
        const p = await r.json();
        attendances = Array.isArray(p) ? p : (p?.data ?? []);
      }
    } catch (e) { console.log('load error', e); }
    finally { loading = false; }
  }

  function setStatus(idx: number, status: string) {
    attendances[idx] = { ...attendances[idx], status };
  }

  function setAllStatus(status: string) {
    attendances = attendances.map(a => ({ ...a, status }));
    toast.success(`Semua siswa di-set ${status}`);
  }

  async function saveAttendances() {
    if (!session) return;
    saving = true;
    try {
      const r = await fetch(`/api/class-journal/sessions/${session.id}/attendances`, {
        method: 'PUT',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify({
          records: attendances.map(a => ({
            student_id: a.student_id,
            status: a.status,
            catatan: a.catatan || ''
          }))
        }),
      });
      if (r.ok) {
        toast.success('Kehadiran tersimpan!');
        await loadAttendances();
      } else {
        const e = await r.json().catch(() => ({}));
        toast.error(e?.error || 'Gagal menyimpan');
      }
    } catch { toast.error('Gagal'); }
    finally { saving = false; }
  }

  const statusMeta: Record<string, { label: string; icon: string; activeClass: string; inactiveClass: string }> = {
    hadir: { label: 'Hadir', icon: '✅', activeClass: 'bg-green-600 text-white ring-2 ring-green-300 shadow-sm', inactiveClass: 'text-green-700 bg-green-50 hover:bg-green-100 border-green-200' },
    sakit: { label: 'Sakit', icon: '🤒', activeClass: 'bg-yellow-500 text-white ring-2 ring-yellow-300 shadow-sm', inactiveClass: 'text-yellow-700 bg-yellow-50 hover:bg-yellow-100 border-yellow-200' },
    izin: { label: 'Izin', icon: '📝', activeClass: 'bg-blue-600 text-white ring-2 ring-blue-300 shadow-sm', inactiveClass: 'text-blue-700 bg-blue-50 hover:bg-blue-100 border-blue-200' },
    alpha: { label: 'Alpha', icon: '❌', activeClass: 'bg-red-600 text-white ring-2 ring-red-300 shadow-sm', inactiveClass: 'text-red-700 bg-red-50 hover:bg-red-100 border-red-200' },
  };

  // Recalculate stats when data changes
  let stats = $derived([
    { label: 'Hadir', count: totalHadir, pct: totalTercatat > 0 ? Math.round(totalHadir / totalTercatat * 100) : 0, color: 'bg-green-500' },
    { label: 'Sakit', count: totalSakit, pct: totalTercatat > 0 ? Math.round(totalSakit / totalTercatat * 100) : 0, color: 'bg-yellow-500' },
    { label: 'Izin', count: totalIzin, pct: totalTercatat > 0 ? Math.round(totalIzin / totalTercatat * 100) : 0, color: 'bg-blue-500' },
    { label: 'Alpha', count: totalAlpha, pct: totalTercatat > 0 ? Math.round(totalAlpha / totalTercatat * 100) : 0, color: 'bg-red-500' },
  ]);

  // Load on mount
  $effect(() => { if (session) loadAttendances(); });
</script>

<svelte:head><title>Absensi {session?.class_code ?? ''} — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-4 max-w-screen-xl mx-auto">
  <!-- Header -->
  <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2">
    <div>
      <a href="/academic/class-journal" class="text-sm text-primary hover:underline">&larr; Kembali ke Jurnal</a>
      {#if session}
        <h1 class="text-xl font-black mt-1">Absensi Jurnal</h1>
        <p class="text-sm text-muted-foreground">
          {session.class_code} — {session.subject_name}
        </p>
      {/if}
    </div>
    <div class="flex items-center gap-2">
      <Button variant="outline" size="sm" onclick={() => window.location.href = '/academic/class-journal'}>Tutup</Button>
      <LoadingButton onclick={() => void saveAttendances()} loading={saving} size="sm">
        💾 Simpan Kehadiran
      </LoadingButton>
    </div>
  </div>

  {#if session}
    <!-- Session info bar -->
    <Card.Root class="border-primary/10 bg-primary-50/30">
      <Card.Content class="p-3 flex flex-wrap items-center gap-x-6 gap-y-1 text-sm">
        <span class="font-semibold">📅 {session.tanggal}</span>
        <span class="text-xs bg-primary/10 text-primary font-bold px-2 py-0.5 rounded-full">Pertemuan #{session.pertemuan_ke}</span>
        <span class="text-muted-foreground">📖 {session.materi || '—'}</span>
        <span class="text-muted-foreground">👨‍🏫 {session.teacher_name}</span>
      </Card.Content>
    </Card.Root>

    <!-- Stats bar -->
    <div class="grid grid-cols-4 gap-2">
      {#each stats as st}
        <Card.Root class="p-2.5 text-center">
          <p class="text-lg font-black {st.count > 0 ? 'text-foreground' : 'text-muted-foreground'}">{st.count}</p>
          <p class="text-[10px] font-semibold uppercase text-muted-foreground">{st.label}</p>
          <div class="mt-1 h-1.5 w-full bg-muted/30 rounded-full overflow-hidden">
            <div class="h-full rounded-full transition-all {st.color}" style="width: {st.pct}%"></div>
          </div>
        </Card.Root>
      {/each}
    </div>

    <!-- Search & batch actions -->
    <div class="flex flex-col sm:flex-row gap-2 items-start sm:items-center justify-between">
      <div class="flex items-center gap-2 w-full sm:w-auto">
        <div class="relative w-full sm:w-64">
          <Input placeholder="🔍 Cari nama atau NIS..." bind:value={searchText} />
        </div>
        <span class="text-xs text-muted-foreground whitespace-nowrap">{filteredAttendances.length}/{totalTercatat} siswa</span>
      </div>
      <div class="flex gap-1">
        <button class="text-[10px] font-bold px-2 py-1 rounded bg-green-50 text-green-700 hover:bg-green-100 border border-green-200" onclick={() => setAllStatus('hadir')}>✅ Semua Hadir</button>
        <button class="text-[10px] font-bold px-2 py-1 rounded bg-red-50 text-red-700 hover:bg-red-100 border border-red-200" onclick={() => setAllStatus('alpha')}>❌ Semua Alpha</button>
      </div>
    </div>

    <!-- Student list -->
    {#if loading}
      <Card.Root><Card.Content class="p-8 text-center"><p class="text-sm text-muted-foreground">Memuat data siswa...</p></Card.Content></Card.Root>
    {:else if attendances.length === 0}
      <Card.Root><Card.Content class="p-8 text-center"><p class="text-sm text-muted-foreground">Belum ada data kehadiran. Pastikan ada murid terdaftar di rombel ini.</p></Card.Content></Card.Root>
    {:else}
      <!-- Desktop table -->
      <div class="hidden sm:block overflow-x-auto rounded-xl border border-border">
        <table class="w-full text-sm">
          <thead>
            <tr class="bg-muted/30 text-muted-foreground text-xs uppercase">
              <th class="px-3 py-2.5 text-left w-8">#</th>
              <th class="px-3 py-2.5 text-left">Nama</th>
              <th class="px-3 py-2.5 text-center hidden md:table-cell">NIS</th>
              <th class="px-3 py-2.5 text-center" colspan="4">Status Kehadiran</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-border">
            {#each filteredAttendances as a, idx (a.student_id)}
              <tr class="hover:bg-muted/10 transition-colors {a.status === 'alpha' ? 'bg-red-50/20' : a.status === 'sakit' ? 'bg-yellow-50/20' : ''}">
                <td class="px-3 py-2 text-xs text-muted-foreground">{idx + 1}</td>
                <td class="px-3 py-2">
                  <span class="text-sm font-medium">{a.nama}</span>
                  <span class="text-xs text-muted-foreground ml-1 md:hidden">({a.nis})</span>
                </td>
                <td class="px-3 py-2 text-xs text-muted-foreground text-center hidden md:table-cell">{a.nis}</td>
                <td class="px-3 py-2" colspan="4">
                  <div class="flex gap-1.5">
                    {#each ['hadir', 'sakit', 'izin', 'alpha'] as st}
                      <button
                        class="flex-1 text-xs font-bold py-1.5 px-2 rounded-lg border transition-all {a.status === st ? statusMeta[st].activeClass : statusMeta[st].inactiveClass}"
                        onclick={() => setStatus(
                          attendances.findIndex(x => x.student_id === a.student_id),
                          st
                        )}
                      >
                        <span class="hidden sm:inline">{statusMeta[st].label}</span>
                        <span class="sm:hidden">{statusMeta[st].icon}</span>
                      </button>
                    {/each}
                  </div>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>

      <!-- Mobile cards -->
      <div class="sm:hidden space-y-2">
        {#each filteredAttendances as a, idx (a.student_id)}
          <div class="rounded-xl border border-border bg-base-100 shadow-sm p-3 space-y-2 {a.status === 'alpha' ? 'border-red-200 bg-red-50/10' : a.status === 'sakit' ? 'border-yellow-200 bg-yellow-50/10' : ''}">
            <div class="flex items-center justify-between">
              <div>
                <span class="text-sm font-semibold">{idx + 1}. {a.nama}</span>
                <span class="text-xs text-muted-foreground ml-1">{a.nis}</span>
              </div>
              <span class="text-xs font-bold {a.status === 'hadir' ? 'text-green-600 bg-green-50 px-2 py-0.5 rounded-full' : a.status === 'alpha' ? 'text-red-600 bg-red-50 px-2 py-0.5 rounded-full' : ''}">
                {statusMeta[a.status]?.label ?? a.status}
              </span>
            </div>
            <div class="grid grid-cols-4 gap-1">
              {#each ['hadir', 'sakit', 'izin', 'alpha'] as st}
                <button
                  class="text-center text-xs font-bold py-2 rounded-lg border transition-all {a.status === st ? statusMeta[st].activeClass : statusMeta[st].inactiveClass}"
                  onclick={() => setStatus(
                    attendances.findIndex(x => x.student_id === a.student_id),
                    st
                  )}
                >
                  <div class="text-base">{statusMeta[st].icon}</div>
                  <div class="text-[10px]">{statusMeta[st].label}</div>
                </button>
              {/each}
            </div>
          </div>
        {/each}
      </div>
    {/if}
  {/if}
</div>
