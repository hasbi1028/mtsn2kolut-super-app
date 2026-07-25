<script lang="ts">
  import { Button } from '$lib/components/ui/button';
  import * as Card from '$lib/components/ui/card';
  import * as Dialog from '$lib/components/ui/dialog';
  import { Input } from '$lib/components/ui/input';
  import { toast } from '$lib/components/ui/sonner';
  import LoadingButton from '$lib/components/LoadingButton.svelte';

  let { data } = $props();

  let assignment = $state(data.assignment);
  let sessions = $state<any[]>([]);
  let loading = $state(false);

  // — Tab state
  let activeTab = $state<'journal' | 'rekap'>('journal');

  // — Create session
  let showCreate = $state(false);
  let createForm = $state({ tanggal: new Date().toISOString().slice(0,10), materi: '', kegiatan: '', catatan: '', guru_hadir: true });
  let createLoading = $state(false);

  // — Summary
  let summaryData = $state<any[]>([]);
  let summaryLoading = $state(false);

  async function loadSessions() {
    if (!assignment) return;
    loading = true;
    try {
      const r = await fetch(`/api/class-journal?assignment_id=${assignment.id}`);
      if (r.ok) { const p = await r.json(); sessions = Array.isArray(p) ? p : (p?.data ?? []); }
    } catch(e) { console.log('loadSessions: error', e); }
    finally { loading = false; }
  }

  async function loadSummary() {
    if (!assignment) return;
    summaryLoading = true;
    try {
      const r = await fetch(`/api/class-journal/summary?assignment_id=${assignment.id}`);
      if (r.ok) { const p = await r.json(); summaryData = Array.isArray(p) ? p : (p?.data ?? []); }
    } catch {}
    finally { summaryLoading = false; }
  }

  $effect(() => {
    if (assignment) { loadSessions(); loadSummary(); }
  });

  $effect(() => {
    if (activeTab === 'rekap' && assignment) loadSummary();
  });

  // — Summary helpers
  let totalStudents = $derived(summaryData.length);
  let totalHadir = $derived(summaryData.reduce((s: number, i: any) => s + i.hadir, 0));
  let totalSakit = $derived(summaryData.reduce((s: number, i: any) => s + i.sakit, 0));
  let totalIzin = $derived(summaryData.reduce((s: number, i: any) => s + i.izin, 0));
  let totalAlpha = $derived(summaryData.reduce((s: number, i: any) => s + i.alpha, 0));
  let totalPertemuan = $derived(summaryData.reduce((s: number, i: any) => s + i.total_pertemuan, 0));
  let avgAttendanceRate = $derived(totalPertemuan > 0 ? Math.round((totalHadir / totalPertemuan) * 100) : 0);

  function rateClass(rate: number): string {
    if (rate >= 95) return 'text-green-600';
    if (rate >= 80) return 'text-yellow-600';
    return 'text-red-600';
  }

  async function submitCreate() {
    if (!assignment) return;
    createLoading = true;
    try {
      const r = await fetch('/api/class-journal', {
        method: 'POST', headers: { 'content-type': 'application/json' },
        body: JSON.stringify({ assignment_id: assignment.id, ...createForm }),
      });
      if (r.ok) {
        const p = await r.json();
        toast.success(`Pertemuan ke-${p?.pertemuan_ke ?? '?'} tersimpan`);
        showCreate = false;
        createForm = { tanggal: new Date().toISOString().slice(0,10), materi: '', kegiatan: '', catatan: '', guru_hadir: true };
        await Promise.all([loadSessions(), loadSummary()]);
      } else { const e = await r.json().catch(() => ({})); toast.error(e?.error || 'Gagal'); }
    } catch { toast.error('Gagal'); }
    finally { createLoading = false; }
  }

  async function deleteSession(id: string, tanggal: string) {
    if (!confirm(`Hapus jurnal tanggal ${tanggal}?`)) return;
    try {
      const r = await fetch(`/api/class-journal/sessions/${id}`, { method: 'DELETE' });
      if (r.ok || r.status === 204) { toast.success('Jurnal dihapus'); await Promise.all([loadSessions(), loadSummary()]); }
      else toast.error('Gagal');
    } catch { toast.error('Gagal'); }
  }
</script>

<svelte:head><title>Jurnal {assignment?.class_code ?? ''} — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-4 max-w-screen-xl mx-auto">
  <!-- Header -->
  <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2">
    <div>
      <a href="/academic/class-journal" class="text-sm text-primary hover:underline">&larr; Dashboard Jurnal</a>
      {#if assignment}
        <h1 class="text-xl font-black mt-1">{assignment.class_code} — {assignment.subject_name}</h1>
        <p class="text-sm text-muted-foreground">👨‍🏫 {assignment.teacher_name} · {sessions.length} pertemuan</p>
      {/if}
    </div>
    <div class="flex gap-2">
      <Button onclick={() => showCreate = true} size="sm">+ Catat Pertemuan</Button>
    </div>
  </div>

  <!-- Tabs -->
  <div class="flex border-b border-border gap-1">
    <button class="px-4 py-2 text-sm font-semibold transition-colors border-b-2 -mb-[1px] {activeTab === 'journal' ? 'border-primary text-primary' : 'border-transparent text-muted-foreground hover:text-foreground'}" onclick={() => activeTab = 'journal'}>
      📝 Catatan Jurnal
    </button>
    <button class="px-4 py-2 text-sm font-semibold transition-colors border-b-2 -mb-[1px] {activeTab === 'rekap' ? 'border-primary text-primary' : 'border-transparent text-muted-foreground hover:text-foreground'}" onclick={() => activeTab = 'rekap'}>
      📊 Rekap Absensi
    </button>
  </div>

  <!-- Tab: Jurnal -->
  {#if activeTab === 'journal'}
    {#if loading}
      <p class="text-sm text-muted-foreground">Memuat jurnal...</p>
    {:else if sessions.length === 0}
      <Card.Root><Card.Content class="p-8 text-center"><p class="text-sm text-muted-foreground">Belum ada catatan jurnal. Klik "+ Catat Pertemuan" untuk memulai.</p></Card.Content></Card.Root>
    {:else}
      <!-- Desktop table -->
      <div class="hidden lg:block overflow-x-auto rounded-lg border border-border">
        <table class="w-full text-sm">
          <thead><tr class="bg-muted/30 text-muted-foreground text-xs uppercase"><th class="px-3 py-2 text-left">Tanggal</th><th class="px-3 py-2 text-left">#</th><th class="px-3 py-2 text-left">Materi</th><th class="px-3 py-2 text-left hidden md:table-cell">Kegiatan</th><th class="px-3 py-2 text-center">Guru</th><th class="px-3 py-2 text-right">Aksi</th></tr></thead>
          <tbody class="divide-y divide-border">
            {#each sessions as s (s.id)}<tr class="hover:bg-muted/10">
              <td class="px-3 py-2 text-xs whitespace-nowrap">{s.tanggal}</td>
              <td class="px-3 py-2 text-xs text-muted-foreground">{s.pertemuan_ke}</td>
              <td class="px-3 py-2 text-xs">{s.materi || '—'}</td>
              <td class="px-3 py-2 text-xs text-muted-foreground hidden md:table-cell max-w-[200px] truncate">{s.kegiatan || '—'}</td>
              <td class="px-3 py-2 text-center">{#if s.guru_hadir}<span class="text-[10px] text-green-600 font-semibold bg-green-50 px-1.5 py-0.5 rounded">Hadir</span>{:else}<span class="text-[10px] text-muted-foreground">—</span>{/if}</td>
              <td class="px-3 py-2 text-right whitespace-nowrap space-x-1">
                <a href={`/academic/class-journal/attendance/${s.id}`} class="text-xs text-primary hover:underline">Absensi</a>
                <button class="text-xs text-destructive hover:underline" onclick={() => deleteSession(s.id, s.tanggal)}>Hapus</button>
              </td>
            </tr>{/each}
          </tbody>
        </table>
      </div>
      <!-- Mobile cards -->
      <div class="lg:hidden space-y-2">
        {#each sessions as s (s.id)}
          <div class="rounded-xl border border-border bg-base-100 shadow-sm p-3 space-y-1.5">
            <div class="flex items-center justify-between">
              <span class="text-sm font-semibold">{s.tanggal}</span>
              <span class="text-xs font-bold text-primary bg-primary/10 px-2 py-0.5 rounded-full">#{s.pertemuan_ke}</span>
            </div>
            <p class="text-xs text-muted-foreground">{s.materi || '—'}</p>
            <div class="flex justify-between items-center pt-0.5">
              <button class="text-xs text-destructive hover:underline" onclick={() => deleteSession(s.id, s.tanggal)}>Hapus</button>
              <a href={`/academic/class-journal/attendance/${s.id}`} class="text-xs text-primary hover:underline">Absensi →</a>
            </div>
          </div>
        {/each}
      </div>
    {/if}
  {/if}

  <!-- Tab: Rekap -->
  {#if activeTab === 'rekap'}
    {#if summaryLoading}
      <p class="text-sm text-muted-foreground">Memuat rekap...</p>
    {:else if summaryData.length === 0}
      <Card.Root><Card.Content class="p-8 text-center"><p class="text-sm text-muted-foreground">Belum ada data kehadiran. Catat pertemuan dan isi absensi dulu.</p></Card.Content></Card.Root>
    {:else}
      <!-- Stats -->
      <div class="grid grid-cols-2 sm:grid-cols-5 gap-3">
        <Card.Root class="p-3 text-center bg-green-50 border-green-200"><p class="text-lg font-black text-green-700">{totalHadir}</p><p class="text-[10px] font-semibold uppercase text-green-600">Hadir</p></Card.Root>
        <Card.Root class="p-3 text-center bg-yellow-50 border-yellow-200"><p class="text-lg font-black text-yellow-700">{totalSakit}</p><p class="text-[10px] font-semibold uppercase text-yellow-600">Sakit</p></Card.Root>
        <Card.Root class="p-3 text-center bg-blue-50 border-blue-200"><p class="text-lg font-black text-blue-700">{totalIzin}</p><p class="text-[10px] font-semibold uppercase text-blue-600">Izin</p></Card.Root>
        <Card.Root class="p-3 text-center bg-red-50 border-red-200"><p class="text-lg font-black text-red-700">{totalAlpha}</p><p class="text-[10px] font-semibold uppercase text-red-600">Alpha</p></Card.Root>
        <Card.Root class="p-3 text-center bg-primary-50 border-primary-200"><p class="text-lg font-black {rateClass(avgAttendanceRate)}">{avgAttendanceRate}%</p><p class="text-[10px] font-semibold uppercase text-muted-foreground">Kehadiran</p></Card.Root>
      </div>
      <div class="flex items-center justify-between text-xs text-muted-foreground">
        <span>{totalStudents} siswa · {totalPertemuan} total pertemuan</span>
        <span class="text-primary">{sessions.length} sesi tercatat</span>
      </div>
      <!-- Desktop table -->
      <div class="hidden md:block overflow-x-auto rounded-lg border border-border">
        <table class="w-full text-sm">
          <thead><tr class="bg-muted/30 text-muted-foreground text-xs uppercase"><th class="px-3 py-2 text-left">Murid</th><th class="px-3 py-2 text-center">NIS</th><th class="px-3 py-2 text-center">Pertemuan</th><th class="px-3 py-2 text-center">Hadir</th><th class="px-3 py-2 text-center">Sakit</th><th class="px-3 py-2 text-center">Izin</th><th class="px-3 py-2 text-center">Alpha</th><th class="px-3 py-2 text-center">%</th></tr></thead>
          <tbody class="divide-y divide-border">
            {#each summaryData as s (s.student_id)}<tr class="hover:bg-muted/10">
              <td class="px-3 py-2 text-xs font-medium">{s.nama}</td>
              <td class="px-3 py-2 text-xs text-muted-foreground text-center">{s.nis}</td>
              <td class="px-3 py-2 text-xs text-center">{s.total_pertemuan}</td>
              <td class="px-3 py-2 text-xs text-center text-green-600 font-semibold">{s.hadir}</td>
              <td class="px-3 py-2 text-xs text-center text-yellow-600">{s.sakit}</td>
              <td class="px-3 py-2 text-xs text-center text-blue-600">{s.izin}</td>
              <td class="px-3 py-2 text-xs text-center text-red-600">{s.alpha}</td>
              <td class="px-3 py-2 text-xs text-center font-semibold {s.total_pertemuan > 0 ? rateClass(Math.round(s.hadir / s.total_pertemuan * 100)) : ''}">
                {s.total_pertemuan > 0 ? Math.round(s.hadir / s.total_pertemuan * 100) + '%' : '—'}
              </td>
            </tr>{/each}
          </tbody>
        </table>
      </div>
      <!-- Mobile cards -->
      <div class="md:hidden space-y-2">
        {#each summaryData as s (s.student_id)}
          <div class="rounded-xl border border-border bg-base-100 shadow-sm p-3 space-y-1.5">
            <div class="flex items-center justify-between">
              <span class="text-sm font-semibold">{s.nama}</span>
              <span class="text-xs font-bold {s.total_pertemuan > 0 ? rateClass(Math.round(s.hadir / s.total_pertemuan * 100)) : ''} bg-muted/20 px-2 py-0.5 rounded-full">
                {s.total_pertemuan > 0 ? Math.round(s.hadir / s.total_pertemuan * 100) + '%' : '—'}
              </span>
            </div>
            <div class="flex items-center gap-3 text-xs text-muted-foreground"><span>📋 {s.total_pertemuan} pertemuan</span><span>🆔 {s.nis}</span></div>
            <div class="flex gap-2 pt-0.5">
              <span class="text-xs font-semibold text-green-700 bg-green-50 px-2 py-0.5 rounded">H {s.hadir}</span>
              <span class="text-xs font-semibold text-yellow-700 bg-yellow-50 px-2 py-0.5 rounded">S {s.sakit}</span>
              <span class="text-xs font-semibold text-blue-700 bg-blue-50 px-2 py-0.5 rounded">I {s.izin}</span>
              <span class="text-xs font-semibold text-red-700 bg-red-50 px-2 py-0.5 rounded">A {s.alpha}</span>
            </div>
          </div>
        {/each}
      </div>
    {/if}
  {/if}

  <!-- Dialog create -->
  <Dialog.Root bind:open={showCreate}>
    <Dialog.Content><div class="space-y-4">
      <h2 class="text-base font-semibold">Catat Pertemuan Baru</h2>
      {#if assignment}
        <p class="text-xs text-muted-foreground">{assignment.class_code} — {assignment.subject_name}</p>
      {/if}
      <div class="grid gap-3 sm:grid-cols-2">
        <div class="space-y-1"><label class="text-xs font-medium text-muted-foreground">Tanggal</label><Input type="date" bind:value={createForm.tanggal} /></div>
        <div class="space-y-1"><label class="text-xs font-medium text-muted-foreground">Guru Hadir</label>
          <label class="flex items-center gap-2 h-9"><input type="checkbox" bind:checked={createForm.guru_hadir} class="toggle" /> <span class="text-sm">{createForm.guru_hadir ? 'Hadir' : 'Tidak Hadir'}</span></label>
        </div>
        <div class="space-y-1 sm:col-span-2"><label class="text-xs font-medium text-muted-foreground">Materi</label><Input bind:value={createForm.materi} placeholder="Materi yang diajarkan" /></div>
        <div class="space-y-1 sm:col-span-2"><label class="text-xs font-medium text-muted-foreground">Kegiatan</label>
          <textarea bind:value={createForm.kegiatan} placeholder="Deskripsi kegiatan" class="flex w-full rounded-lg border border-input bg-background px-3 py-2 text-sm min-h-[60px]"></textarea>
        </div>
        <div class="space-y-1 sm:col-span-2"><label class="text-xs font-medium text-muted-foreground">Catatan</label><Input bind:value={createForm.catatan} placeholder="Catatan tambahan" /></div>
      </div>
      <div class="flex justify-end gap-2 pt-2">
        <Button variant="outline" onclick={() => showCreate = false}>Batal</Button>
        <LoadingButton onclick={() => void submitCreate()} loading={createLoading}>Simpan</LoadingButton>
      </div>
    </div></Dialog.Content>
  </Dialog.Root>
</div>
