<script lang="ts">
  import { Button } from '$lib/components/ui/button';
  import * as Card from '$lib/components/ui/card';
  import * as Dialog from '$lib/components/ui/dialog';
  import { Input } from '$lib/components/ui/input';
  import { toast } from '$lib/components/ui/sonner';
  import LoadingButton from '$lib/components/LoadingButton.svelte';

  let { data } = $props();

  type Class = { id: string; code: string; name: string; level: string; };
  type Assignment = { id: string; class_id: string; class_code: string; class_name: string; subject_name: string; teacher_name: string; };
  type Session = { id: string; assignment_id: string; tanggal: string; pertemuan_ke: number; materi: string; kegiatan: string; catatan: string; guru_hadir: boolean; class_name: string; subject_name: string; teacher_name: string; };
  type SummaryItem = { student_id: string; nis: string; nisn: string; nama: string; total_pertemuan: number; hadir: number; sakit: number; izin: number; alpha: number; };

  let classes = $state<Class[]>(data.classes ?? []);
  let assignments = $state<Assignment[]>(data.assignments ?? []);
  let sessions = $state<Session[]>([]);
  let loading = $state(false);

  // — Tab state
  let activeTab = $state<'journal' | 'rekap'>('journal');
  let lastSessionCount = $state(0);

  // — Step 1a: Select class
  let selectedClassId = $state('');
  let selectedClass = $derived(classes.find(c => c.id === selectedClassId));

  // — Step 1b: Filtered assignments by selected class
  let filteredAssignments = $derived(
    selectedClassId
      ? assignments.filter(a => a.class_id === selectedClassId)
      : []
  );

  // — Step 1c: Select assignment
  let selectedAssignmentId = $state('');
  let selectedAssignment = $derived(assignments.find(a => a.id === selectedAssignmentId));

  // — Step 2: Create session
  let showCreate = $state(false);
  let createForm = $state({ tanggal: new Date().toISOString().slice(0,10), materi: '', kegiatan: '', catatan: '', guru_hadir: true });
  let createLoading = $state(false);

  // — Step 3: Attendance (per-session)
  let selectedSession = $state<Session | null>(null);
  let attendances = $state<any[]>([]);
  let attLoading = $state(false);

  // — Summary data
  let summaryData = $state<SummaryItem[]>([]);
  let summaryLoading = $state(false);

  async function selectClass(id: string) {
    selectedClassId = id;
    selectedAssignmentId = '';
    sessions = [];
    summaryData = [];
  }

  async function selectAssignment(id: string) {
    selectedAssignmentId = id;
    await Promise.all([loadSessions(), loadSummary()]);
  }

  async function loadSessions() {
    if (!selectedAssignmentId) return;
    loading = true;
    try {
      const r = await fetch(`/api/class-journal?assignment_id=${selectedAssignmentId}`);
      if (r.ok) {
        const p = await r.json();
        sessions = Array.isArray(p) ? p : (p?.data ?? []);
        lastSessionCount = sessions.length;
      }
    } catch(e) { console.log('loadSessions: error', e); }
    finally { loading = false; }
  }

  async function loadSummary() {
    if (!selectedAssignmentId) return;
    summaryLoading = true;
    try {
      const r = await fetch(`/api/class-journal/summary?assignment_id=${selectedAssignmentId}`);
      if (r.ok) { const p = await r.json(); summaryData = Array.isArray(p) ? p : (p?.data ?? []); }
    } catch {}
    finally { summaryLoading = false; }
  }

  // — Auto-refresh saat pindah tab
  $effect(() => {
    if (activeTab === 'rekap' && selectedAssignmentId) {
      loadSummary();
    }
    if (activeTab === 'journal' && selectedAssignmentId && sessions.length === 0) {
      loadSessions();
    }
  });

  // — Summary helpers
  let totalStudents = $derived(summaryData.length);
  let totalHadir = $derived(summaryData.reduce((s, i) => s + i.hadir, 0));
  let totalSakit = $derived(summaryData.reduce((s, i) => s + i.sakit, 0));
  let totalIzin = $derived(summaryData.reduce((s, i) => s + i.izin, 0));
  let totalAlpha = $derived(summaryData.reduce((s, i) => s + i.alpha, 0));
  let totalPertemuan = $derived(summaryData.reduce((s, i) => s + i.total_pertemuan, 0));
  let avgAttendanceRate = $derived(
    totalPertemuan > 0 ? Math.round((totalHadir / totalPertemuan) * 100) : 0
  );

  function rateClass(rate: number): string {
    if (rate >= 95) return 'text-green-600';
    if (rate >= 80) return 'text-yellow-600';
    return 'text-red-600';
  }

  async function submitCreate() {
    createLoading = true;
    try {
      const r = await fetch('/api/class-journal', {
        method: 'POST', headers: { 'content-type': 'application/json' },
        body: JSON.stringify({ assignment_id: selectedAssignmentId, ...createForm }),
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

  async function openAttendance(session: Session) {
    selectedSession = session;
    attLoading = true;
    try {
      const r = await fetch(`/api/class-journal/sessions/${session.id}/attendances`);
      if (r.ok) { const p = await r.json(); attendances = Array.isArray(p) ? p : (p?.data ?? []); }
    } catch {}
    finally { attLoading = false; }
  }

  function setStatus(idx: number, status: string) {
    attendances[idx] = { ...attendances[idx], status };
  }

  async function saveAttendances() {
    if (!selectedSession) return;
    attLoading = true;
    try {
      const r = await fetch(`/api/class-journal/sessions/${selectedSession.id}/attendances`, {
        method: 'PUT', headers: { 'content-type': 'application/json' },
        body: JSON.stringify({ records: attendances.map(a => ({ student_id: a.student_id, status: a.status, catatan: a.catatan || '' })) }),
      });
      if (r.ok) {
        toast.success('Kehadiran tersimpan');
        // Refresh summary & session list after saving attendance
        await Promise.all([loadSessions(), loadSummary()]);
        selectedSession = null;
      }
      else { const e = await r.json().catch(() => ({})); toast.error(e?.error || 'Gagal'); }
    } catch { toast.error('Gagal'); }
    finally { attLoading = false; }
  }

  let rekapInfo = $derived(
    sessions.length > lastSessionCount
      ? { text: `+${sessions.length - lastSessionCount} pertemuan baru`, type: 'info' as const }
      : summaryData.length > 0
      ? { text: `${totalPertemuan} pertemuan`, type: 'default' as const }
      : null
  );

  // — Status badge helpers
  const statusMeta: Record<string, { label: string; activeClass: string; inactiveClass: string }> = {
    hadir:  { label: 'H', activeClass: 'bg-green-600 text-white ring-2 ring-green-300', inactiveClass: 'text-green-700 bg-green-50 hover:bg-green-100 border-green-200' },
    sakit:  { label: 'S', activeClass: 'bg-yellow-500 text-white ring-2 ring-yellow-300', inactiveClass: 'text-yellow-700 bg-yellow-50 hover:bg-yellow-100 border-yellow-200' },
    izin:   { label: 'I', activeClass: 'bg-blue-600 text-white ring-2 ring-blue-300', inactiveClass: 'text-blue-700 bg-blue-50 hover:bg-blue-100 border-blue-200' },
    alpha:  { label: 'A', activeClass: 'bg-red-600 text-white ring-2 ring-red-300', inactiveClass: 'text-red-700 bg-red-50 hover:bg-red-100 border-red-200' },
  };
</script>

<svelte:head><title>Jurnal Belajar Harian — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-4 max-w-screen-xl mx-auto">
  <!-- Header -->
  <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2">
    <div><h1 class="text-xl font-black">Jurnal Belajar Harian</h1><p class="text-sm text-muted-foreground">Catat kegiatan belajar & kehadiran murid per pertemuan.</p></div>
  </div>

  <!-- Step 1: Pilih Kelas & Mapel -->
  <Card.Root>
    <Card.Header class="px-4 pt-3 pb-1"><Card.Title class="text-sm font-semibold">Pilih Kelas & Mata Pelajaran</Card.Title></Card.Header>
    <Card.Content class="px-4 pb-3 space-y-3">
      <div class="flex flex-col sm:flex-row gap-3">
        <div class="flex-1 min-w-0">
          <label class="text-xs font-medium text-muted-foreground mb-1 block">Kelas</label>
          <select class="flex h-9 w-full rounded-lg border border-input bg-background px-3 text-sm" bind:value={selectedClassId} onchange={() => selectClass(selectedClassId)}>
            <option value="">— Pilih Kelas —</option>
            {#each classes as c}
              <option value={c.id}>{c.code}</option>
            {/each}
          </select>
        </div>
        <div class="flex-1 min-w-0">
          <label class="text-xs font-medium text-muted-foreground mb-1 block">Mata Pelajaran</label>
          <select class="flex h-9 w-full rounded-lg border border-input bg-background px-3 text-sm" disabled={!selectedClassId} bind:value={selectedAssignmentId} onchange={() => selectAssignment(selectedAssignmentId)}>
            <option value="">{selectedClassId ? '— Pilih Mapel —' : 'Pilih kelas terlebih dahulu'}</option>
            {#each filteredAssignments as a}
              <option value={a.id}>{a.subject_name} ({a.teacher_name})</option>
            {/each}
          </select>
        </div>
      </div>
    </Card.Content>
  </Card.Root>

  {#if selectedAssignment}
    <!-- Tabs -->
    <div class="flex border-b border-border gap-1">
      <button class="px-4 py-2 text-sm font-semibold transition-colors border-b-2 -mb-[1px] {activeTab === 'journal' ? 'border-primary text-primary' : 'border-transparent text-muted-foreground hover:text-foreground'}" onclick={() => activeTab = 'journal'}>
        📝 Catatan Jurnal
      </button>
      <button class="px-4 py-2 text-sm font-semibold transition-colors border-b-2 -mb-[1px] {activeTab === 'rekap' ? 'border-primary text-primary' : 'border-transparent text-muted-foreground hover:text-foreground'}" onclick={() => activeTab = 'rekap'}>
        📊 Rekap Absensi
        {#if rekapInfo && activeTab !== 'rekap'}
          <span class="ml-1.5 inline-flex items-center justify-center px-1.5 py-0.5 text-[10px] font-bold rounded-full bg-primary/10 text-primary">{totalPertemuan}</span>
        {/if}
      </button>
    </div>

    <!-- Tab: Catatan Jurnal -->
    {#if activeTab === 'journal'}
      <div class="flex justify-between items-center">
        <div>
          <p class="text-sm font-medium">{selectedAssignment.class_code} — {selectedAssignment.subject_name}</p>
          <p class="text-xs text-muted-foreground">{sessions.length} pertemuan</p>
        </div>
        <Button onclick={() => showCreate = true} size="sm">+ Catat Pertemuan</Button>
      </div>

      {#if loading}
        <p class="text-sm text-muted-foreground">Memuat jurnal...</p>
      {:else if sessions.length === 0}
        <Card.Root><Card.Content class="p-8 text-center"><p class="text-sm text-muted-foreground">Belum ada catatan jurnal untuk mapel ini. Klik "+ Catat Pertemuan" untuk memulai.</p></Card.Content></Card.Root>
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
                  <button class="text-xs text-primary hover:underline" onclick={() => openAttendance(s)}>Absensi</button>
                  <button class="text-xs text-destructive hover:underline" onclick={() => deleteSession(s.id, s.tanggal)}>Hapus</button>
                </td>
              </tr>{/each}
            </tbody>
          </table>
        </div>
        <!-- Mobile card list -->
        <div class="lg:hidden space-y-2">
          {#each sessions as s (s.id)}
            <div class="rounded-xl border border-border bg-base-100 shadow-sm p-3 space-y-1.5" role="button" onclick={() => openAttendance(s)}>
              <div class="flex items-center justify-between">
                <span class="text-sm font-semibold">{s.tanggal}</span>
                <span class="text-xs font-bold text-primary bg-primary/10 px-2 py-0.5 rounded-full">#{s.pertemuan_ke}</span>
              </div>
              <p class="text-xs text-muted-foreground">{s.materi || '—'}</p>
              <div class="flex justify-between items-center pt-0.5">
                <button class="text-xs text-destructive hover:underline" onclick={(e) => { e.stopPropagation(); deleteSession(s.id, s.tanggal); }}>Hapus</button>
                {#if s.guru_hadir}<span class="text-[10px] text-green-600">Guru Hadir</span>{/if}
              </div>
            </div>
          {/each}
        </div>
      {/if}
    {/if}

    <!-- Tab: Rekap Absensi -->
    {#if activeTab === 'rekap'}
      {#if summaryLoading}
        <p class="text-sm text-muted-foreground">Memuat rekap...</p>
      {:else if summaryData.length === 0}
        <Card.Root><Card.Content class="p-8 text-center"><p class="text-sm text-muted-foreground">Belum ada data kehadiran. Catat pertemuan dan isi absensi dulu.</p></Card.Content></Card.Root>
      {:else}
        <!-- Summary stats -->
        <div class="grid grid-cols-2 sm:grid-cols-5 gap-3">
          <Card.Root class="p-3 text-center bg-green-50 border-green-200">
            <p class="text-lg font-black text-green-700">{totalHadir}</p>
            <p class="text-[10px] font-semibold uppercase text-green-600">Hadir</p>
          </Card.Root>
          <Card.Root class="p-3 text-center bg-yellow-50 border-yellow-200">
            <p class="text-lg font-black text-yellow-700">{totalSakit}</p>
            <p class="text-[10px] font-semibold uppercase text-yellow-600">Sakit</p>
          </Card.Root>
          <Card.Root class="p-3 text-center bg-blue-50 border-blue-200">
            <p class="text-lg font-black text-blue-700">{totalIzin}</p>
            <p class="text-[10px] font-semibold uppercase text-blue-600">Izin</p>
          </Card.Root>
          <Card.Root class="p-3 text-center bg-red-50 border-red-200">
            <p class="text-lg font-black text-red-700">{totalAlpha}</p>
            <p class="text-[10px] font-semibold uppercase text-red-600">Alpha</p>
          </Card.Root>
          <Card.Root class="p-3 text-center bg-primary-50 border-primary-200">
            <p class="text-lg font-black {rateClass(avgAttendanceRate)}">{avgAttendanceRate}%</p>
            <p class="text-[10px] font-semibold uppercase text-muted-foreground">Kehadiran</p>
          </Card.Root>
        </div>

        <!-- Info bar -->
        <div class="flex items-center justify-between text-xs text-muted-foreground">
          <span>{totalStudents} siswa · {totalPertemuan} total pertemuan</span>
          {#if sessions.length > 0}
            <span class="text-primary">{sessions.length} sesi tercatat</span>
          {/if}
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
              <div class="flex items-center gap-3 text-xs text-muted-foreground">
                <span>📋 {s.total_pertemuan} pertemuan</span>
                <span>🆔 {s.nis}</span>
              </div>
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
  {:else if selectedClassId && filteredAssignments.length === 0}
    <Card.Root><Card.Content class="p-8 text-center"><p class="text-sm text-muted-foreground">Tidak ada mata pelajaran untuk kelas ini. Silakan assign guru terlebih dahulu.</p></Card.Content></Card.Root>
  {/if}
</div>

<!-- Dialog tambah sesi jurnal -->
<Dialog.Root bind:open={showCreate}>
  <Dialog.Content><div class="space-y-4">
    <h2 class="text-base font-semibold">Catat Pertemuan Baru</h2>
    {#if selectedAssignment}
      <p class="text-xs text-muted-foreground">{selectedAssignment.class_code} — {selectedAssignment.subject_name}</p>
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

<!-- Dialog absensi murid -->
<Dialog.Root open={!!selectedSession} onOpenChange={(v: boolean) => { if (!v) { selectedSession = null; } }}>
  <Dialog.Content class="max-w-2xl"><div class="space-y-4">
    <div class="flex items-center justify-between">
      <h2 class="text-base font-semibold">Absensi Jurnal</h2>
      {#if selectedSession}
        <span class="text-xs font-bold bg-primary/10 text-primary px-2 py-0.5 rounded-full">#{selectedSession.pertemuan_ke}</span>
      {/if}
    </div>
    {#if selectedSession}
      <div class="flex items-center gap-2 text-xs text-muted-foreground bg-muted/20 px-3 py-2 rounded-lg">
        <span class="font-semibold">{selectedSession.tanggal}</span>
        <span class="text-muted-foreground/50">|</span>
        <span>{selectedSession.materi || '—'}</span>
      </div>
    {/if}

    {#if attLoading}
      <p class="text-sm text-muted-foreground py-4 text-center">Memuat data siswa...</p>
    {:else if attendances.length === 0}
      <p class="text-sm text-muted-foreground py-4 text-center">Belum ada data kehadiran. Pastikan ada murid terdaftar di rombel ini.</p>
    {:else}
      <div class="max-h-80 overflow-y-auto space-y-1.5 pr-1">
        {#each attendances as a, idx}
          <div class="flex items-center justify-between gap-2 rounded-lg border border-border px-3 py-2 hover:bg-muted/10 transition-colors">
            <div class="min-w-0 flex-1 flex items-center gap-2">
              <span class="text-xs font-medium text-muted-foreground w-5 shrink-0">{idx + 1}.</span>
              <span class="text-sm font-medium truncate">{a.nama}</span>
              <span class="text-[10px] text-muted-foreground hidden sm:inline">({a.nis})</span>
            </div>
            <div class="flex gap-1.5 shrink-0">
              {#each ['hadir', 'sakit', 'izin', 'alpha'] as st}
                <button
                  class="w-8 h-8 text-xs font-bold rounded-lg border transition-all {a.status === st ? statusMeta[st].activeClass : statusMeta[st].inactiveClass}"
                  onclick={() => setStatus(idx, st)}
                  title={st}
                >{statusMeta[st].label}</button>
              {/each}
            </div>
          </div>
        {/each}
      </div>
      <div class="flex items-center justify-between pt-2 border-t border-border">
        <p class="text-xs text-muted-foreground">{attendances.length} siswa</p>
        <div class="flex gap-2">
          <Button variant="outline" onclick={() => selectedSession = null} size="sm">Tutup</Button>
          <LoadingButton onclick={() => void saveAttendances()} loading={attLoading} size="sm">Simpan Kehadiran</LoadingButton>
        </div>
      </div>
    {/if}
  </div></Dialog.Content>
</Dialog.Root>
