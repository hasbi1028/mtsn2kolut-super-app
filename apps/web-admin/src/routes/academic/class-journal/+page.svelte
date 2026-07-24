<script lang="ts">
  import { Badge } from '$lib/components/ui/badge';
  import { Button } from '$lib/components/ui/button';
  import * as Card from '$lib/components/ui/card';
  import * as Dialog from '$lib/components/ui/dialog';
  import * as Select from '$lib/components/ui/select';
  import { Input } from '$lib/components/ui/input';
  import { toast } from '$lib/components/ui/sonner';
  import LoadingButton from '$lib/components/LoadingButton.svelte';

  let { data } = $props();

  type Assignment = { id: string; class_id: string; class_code: string; class_name: string; subject_name: string; teacher_name: string; };
  type Session = { id: string; assignment_id: string; tanggal: string; pertemuan_ke: number; materi: string; kegiatan: string; catatan: string; guru_hadir: boolean; class_name: string; subject_name: string; teacher_name: string; };

  let assignments = $state<Assignment[]>(data.assignments ?? []);
  let sessions = $state<Session[]>([]);
  let loading = $state(false);

  // — Step 1: Select assignment
  let selectedAssignmentId = $state('');
  let selectedAssignment = $derived(assignments.find(a => a.id === selectedAssignmentId));

  // — Step 2: Create session
  let showCreate = $state(false);
  let createForm = $state({ tanggal: new Date().toISOString().slice(0,10), materi: '', kegiatan: '', catatan: '', guru_hadir: true });
  let createLoading = $state(false);

  // — Step 3: Attendance
  let selectedSession = $state<Session | null>(null);
  let attendances = $state<any[]>([]);
  let attLoading = $state(false);

  async function selectAssignment(id: string) {
    selectedAssignmentId = id;
    await loadSessions();
  }

  async function loadSessions() {
    if (!selectedAssignmentId) return;
    loading = true;
    try {
      const r = await fetch(`/api/class-journal?assignment_id=${selectedAssignmentId}`);
      if (r.ok) { const p = await r.json(); sessions = p.data ?? []; }
    } catch {}
    finally { loading = false; }
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
        toast.success(`Pertemuan ke-${p.data?.pertemuan_ke || '?'} tersimpan`);
        showCreate = false;
        createForm = { tanggal: new Date().toISOString().slice(0,10), materi: '', kegiatan: '', catatan: '', guru_hadir: true };
        await loadSessions();
      } else { const e = await r.json().catch(() => ({})); toast.error(e?.error || 'Gagal'); }
    } catch { toast.error('Gagal'); }
    finally { createLoading = false; }
  }

  async function openAttendance(session: Session) {
    selectedSession = session;
    attLoading = true;
    try {
      const r = await fetch(`/api/class-journal/sessions/${session.id}/attendances`);
      if (r.ok) { const p = await r.json(); attendances = p.data ?? []; }
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
      if (r.ok) { toast.success('Kehadiran tersimpan'); }
      else { const e = await r.json().catch(() => ({})); toast.error(e?.error || 'Gagal'); }
    } catch { toast.error('Gagal'); }
    finally { attLoading = false; }
  }
</script>

<svelte:head><title>Jurnal Belajar Harian — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-4 max-w-screen-xl mx-auto">
  <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2">
    <div><h1 class="text-xl font-black">Jurnal Belajar Harian</h1><p class="text-sm text-muted-foreground">Catat kegiatan belajar & kehadiran murid per pertemuan.</p></div>
  </div>

  <!-- Step 1: Pilih mapel -->
  <Card.Root>
    <Card.Header class="px-4 pt-3 pb-1"><Card.Title class="text-sm font-semibold">Pilih Mata Pelajaran</Card.Title></Card.Header>
    <Card.Content class="px-4 pb-3">
      <select class="flex h-9 w-full max-w-md rounded-lg border border-input bg-background px-3 text-sm" onchange={(e) => selectAssignment((e.target as HTMLSelectElement).value)}>
        <option value="">Pilih rombel & mapel...</option>
        {#each assignments as a}
          <option value={a.id}>{a.class_code} — {a.subject_name} ({a.teacher_name})</option>
        {/each}
      </select>
    </Card.Content>
  </Card.Root>

  {#if selectedAssignment}
    <div class="flex justify-between items-center">
      <p class="text-sm font-medium">{selectedAssignment.class_code} — {selectedAssignment.subject_name}</p>
      <Button onclick={() => showCreate = true} size="sm">+ Catat Pertemuan</Button>
    </div>

    <!-- Sessions List -->
    {#if loading}
      <p class="text-sm text-muted-foreground">Memuat jurnal...</p>
    {:else if sessions.length === 0}
      <Card.Root><Card.Content class="p-8 text-center"><p class="text-sm text-muted-foreground">Belum ada catatan jurnal untuk mapel ini.</p></Card.Content></Card.Root>
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
              <td class="px-3 py-2 text-center">{#if s.guru_hadir}<span class="text-[10px] text-primary">Hadir</span>{:else}<span class="text-[10px] text-muted-foreground">—</span>{/if}</td>
              <td class="px-3 py-2 text-right"><button class="text-xs text-primary hover:underline" onclick={() => openAttendance(s)}>Absensi</button></td>
            </tr>{/each}
          </tbody>
        </table>
      </div>
      <!-- Mobile card list -->
      <div class="lg:hidden space-y-2">
        {#each sessions as s (s.id)}
          <div class="rounded-xl border border-border bg-base-100 shadow-sm p-3 space-y-1" role="button" onclick={() => openAttendance(s)}>
            <div class="flex items-center justify-between"><span class="text-sm font-semibold">{s.tanggal}</span><span class="text-xs text-muted-foreground">#{s.pertemuan_ke}</span></div>
            <p class="text-xs">{s.materi || '—'}</p>
            <div class="flex justify-end"><span class="text-[10px] text-primary">Absensi →</span></div>
          </div>
        {/each}
      </div>
    {/if}
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
<Dialog.Root open={!!selectedSession} onOpenChange={(v) => { if (!v) selectedSession = null; }}>
  <Dialog.Content class="max-w-2xl"><div class="space-y-4">
    <h2 class="text-base font-semibold">Absensi Jurnal</h2>
    {#if selectedSession}
      <p class="text-xs text-muted-foreground">{selectedSession.tanggal} — Pertemuan #{selectedSession.pertemuan_ke}<br/>{selectedSession.materi}</p>
    {/if}

    {#if attLoading}
      <p class="text-sm text-muted-foreground">Memuat...</p>
    {:else if attendances.length === 0}
      <p class="text-sm text-muted-foreground">Belum ada data kehadiran. Pastikan ada murid terdaftar di rombel ini.</p>
    {:else}
      <div class="max-h-80 overflow-y-auto space-y-1">
        {#each attendances as a, idx}
          <div class="flex items-center justify-between gap-2 rounded-lg border border-border px-3 py-2 hover:bg-muted/10">
            <div class="min-w-0 flex-1"><span class="text-sm font-medium">{a.nama}</span><span class="text-xs text-muted-foreground ml-2">({a.nis})</span></div>
            <div class="flex gap-1">
              {#each ['hadir', 'sakit', 'izin', 'alpha'] as st}
                <button class="px-2 py-0.5 text-[10px] font-semibold uppercase rounded-md transition-colors {a.status === st ? 'bg-primary text-primary-foreground' : 'bg-muted/30 text-muted-foreground hover:bg-muted/50'}" onclick={() => setStatus(idx, st)}>{st}</button>
              {/each}
            </div>
          </div>
        {/each}
      </div>
      <div class="flex justify-end gap-2 pt-2">
        <Button variant="outline" onclick={() => selectedSession = null}>Tutup</Button>
        <LoadingButton onclick={() => void saveAttendances()} loading={attLoading}>Simpan Kehadiran</LoadingButton>
      </div>
    {/if}
  </div></Dialog.Content>
</Dialog.Root>
