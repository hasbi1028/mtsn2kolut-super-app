<script lang="ts">
  import { Badge } from '$lib/components/ui/badge';
  import { Button } from '$lib/components/ui/button';
  import * as Card from '$lib/components/ui/card';
  import * as Dialog from '$lib/components/ui/dialog';
  import { Input } from '$lib/components/ui/input';
  import { toast } from '$lib/components/ui/sonner';
  import LoadingButton from '$lib/components/LoadingButton.svelte';

  let { data } = $props();

  type Slot = {
    id: string; assignment_id: string; day_of_week: number;
    start_time: string; end_time: string; room_label: string;
    subject_name: string; teacher_name: string; class_name: string;
    class_code: string; lesson_period_label: string;
  };

  let slots = $state<Slot[]>(data.slots ?? []);
  let loading = $state(false);

  const days = ['Senin', 'Selasa', 'Rabu', 'Kamis', "Jum'at", 'Sabtu'];

  let classes = $state<{ id: string; code: string; name: string; level: string }[]>(data.classes ?? []);
  let subjects = $state<any[]>(data.subjects ?? []);
  let teachers = $state<any[]>(data.teachers ?? []);
  let assignments = $state<any[]>(data.assignments ?? []);

  // ── Lookup maps (O(1) — dihitung sekali via $derived) ──
  let assignmentById = $derived(new Map(assignments.map(a => [a.id, a])));
  let classById = $derived(new Map(classes.map(c => [c.id, c])));

  // Map classId -> dayOfWeek -> Slot[]
  let slotsByClassDay = $derived.by(() => {
    const map = new Map<string, Slot[]>();
    for (const s of slots) {
      const a = assignmentById.get(s.assignment_id);
      if (!a) continue;
      const key = `${a.class_id}|${s.day_of_week}`;
      let arr = map.get(key);
      if (!arr) { arr = []; map.set(key, arr); }
      arr.push(s);
    }
    return map;
  });

  // Pagination: 1 class per page
  let kelasIndex = $state(0);
  const maxKelasIndex = $derived(Math.max(0, classes.length - 1));
  let kelasTerpilih = $derived(classes[kelasIndex]);

  // Dialog state
  let showSlotDialog = $state(false);
  let editMode = $state(false);
  let editSlotId = $state<string | null>(null);
  let slotForm = $state({
    assignment_id: '', day_of_week: 1, start_time: '07:00', end_time: '07:40',
    room_label: '', slot_type: 'pelajaran', lesson_hours: 1,
  });
  let slotLoading = $state(false);
  let slotError = $state('');

  // Delete confirmation
  let deleteConfirmId = $state<string | null>(null);
  let deleteLoading = $state(false);
  let deleteDialogOpen = $state(false);

  // Lazy conflict detection
  let computeConflicts = $state(false);
  let conflicts = $state<any[]>([]);

  function timeLabel(s: Slot): string {
    if (!s.start_time && !s.end_time) return '—';
    const st = s.start_time?.slice(0, 5) || '??';
    const et = s.end_time?.slice(0, 5) || '??';
    return `${st}–${et}`;
  }

  function validateForm(): boolean {
    if (!slotForm.assignment_id) { slotError = 'Pilih mata pelajaran'; return false; }
    if (!slotForm.start_time) { slotError = 'Isi jam mulai'; return false; }
    if (!slotForm.end_time) { slotError = 'Isi jam selesai'; return false; }
    if (slotForm.start_time >= slotForm.end_time) { slotError = 'Jam selesai harus lebih dari jam mulai'; return false; }
    slotError = '';
    return true;
  }

  function openAdd(day?: number) {
    editMode = false; editSlotId = null; slotError = '';
    slotForm = { assignment_id: '', day_of_week: day ?? 1, start_time: '07:00', end_time: '07:40', room_label: '', slot_type: 'pelajaran', lesson_hours: 1 };
    showSlotDialog = true;
  }

  function openEdit(slot: Slot) {
    editMode = true; editSlotId = slot.id; slotError = '';
    slotForm = { assignment_id: slot.assignment_id, day_of_week: slot.day_of_week, start_time: slot.start_time, end_time: slot.end_time, room_label: slot.room_label, slot_type: 'pelajaran', lesson_hours: 1 };
    showSlotDialog = true;
  }

  async function submitSlot() {
    if (!validateForm()) return;
    slotLoading = true;
    try {
      const url = editMode && editSlotId ? `/api/academic/timetable/slots/${editSlotId}` : '/api/academic/timetable/slots';
      const r = await fetch(url, {
        method: editMode && editSlotId ? 'PUT' : 'POST',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify(slotForm),
      });
      if (r.ok) {
        toast.success(editMode ? 'Slot diperbarui' : 'Slot tersimpan');
        showSlotDialog = false; computeConflicts = false; await refresh();
      } else {
        const e = await r.json().catch(() => ({}));
        slotError = e?.error || `Gagal ${editMode ? 'memperbarui' : 'menyimpan'} slot`;
      }
    } catch { slotError = 'Terjadi kesalahan, coba lagi'; }
    finally { slotLoading = false; }
  }

  async function executeDelete() {
    if (!deleteConfirmId) return;
    deleteLoading = true;
    try {
      const r = await fetch(`/api/academic/timetable/slots/${deleteConfirmId}`, { method: 'DELETE' });
      if (r.ok || r.status === 204) {
        toast.success('Slot dihapus'); deleteConfirmId = null; deleteDialogOpen = false;
        computeConflicts = false; await refresh();
      } else toast.error('Gagal menghapus slot');
    } catch { toast.error('Gagal'); }
    finally { deleteLoading = false; }
  }

  async function refresh() {
    loading = true;
    try {
      const r = await fetch('/api/academic/timetable/weekly');
      if (r.ok) {
        const p = await r.json(); const d = p.data ?? p ?? {};
        classes = d.classes ?? []; subjects = d.subjects ?? [];
        teachers = d.teachers ?? []; assignments = d.assignments ?? []; slots = d.slots ?? [];
      }
    } catch {}
    finally { loading = false; }
  }

  function dayName(day: number): string { return days[day - 1] || '?'; }

  // ── Lazy conflict computation ──
  function cekBentrokan() {
    const result: any[] = [];
    const byTeacher: Record<string, Slot[]> = {};
    const byClass: Record<string, Slot[]> = {};
    const byRoom: Record<string, Slot[]> = {};

    function groupKey(day: number, name: string): string { return `${day}|${name}`; }
    function timeOverlap(a: Slot, b: Slot): boolean {
      if (!a.start_time || !a.end_time || !b.start_time || !b.end_time) return false;
      return a.start_time < b.end_time && b.start_time < a.end_time;
    }

    // Group slots
    for (const s of slots) {
      if (!s.start_time || !s.end_time) continue;
      if (s.teacher_name) {
        (byTeacher[groupKey(s.day_of_week, s.teacher_name)] ??= []).push(s);
      }
      if (s.class_code) {
        (byClass[groupKey(s.day_of_week, s.class_code)] ??= []).push(s);
      }
      if (s.room_label) {
        (byRoom[groupKey(s.day_of_week, s.room_label)] ??= []).push(s);
      }
    }

    function check(map: Record<string, Slot[]>, type: string, labelKey: string) {
      for (const [key, items] of Object.entries(map)) {
        const day = Number(key.split('|')[0]);
        for (let i = 0; i < items.length; i++) {
          for (let j = i + 1; j < items.length; j++) {
            if (!timeOverlap(items[i], items[j])) continue;
            result.push({
              type, day, time_range: `${timeLabel(items[i])} / ${timeLabel(items[j])}`,
              slots: [items[i], items[j]],
              description: type === 'teacher_overlap'
                ? `👨‍🏫 ${items[i].teacher_name} mengajar "${items[i].subject_name}" (${items[i].class_code}) dan "${items[j].subject_name}" (${items[j].class_code}) di jam bersamaan!`
                : type === 'class_overlap'
                ? `🏫 ${items[i].class_code} diampu "${items[i].subject_name}" (${items[i].teacher_name}) dan "${items[j].subject_name}" (${items[j].teacher_name}) di jam bersamaan!`
                : `🚪 Ruang ${items[i].room_label} dipakai ${items[i].class_code} dan ${items[j].class_code} di jam bersamaan!`
            });
          }
        }
      }
    }

    check(byTeacher, 'teacher_overlap', 'teacher_name');
    check(byClass, 'class_overlap', 'class_code');
    check(byRoom, 'room_overlap', 'room_label');

    conflicts = result;
    computeConflicts = true;
  }

  const conflictSlotIds = $derived(computeConflicts ? new Set(conflicts.flatMap((c: any) => c.slots.map((s: Slot) => s.id))) : new Set());
</script>

<svelte:head><title>Jadwal Pelajaran — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-4 max-w-screen-xl mx-auto pb-8">
  <!-- Header -->
  <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2">
    <div>
      <h1 class="text-xl font-black">📅 Jadwal Pelajaran</h1>
      <p class="text-sm text-muted-foreground">Atur jadwal mingguan per rombel.</p>
      <p class="text-[10px] text-muted-foreground mt-0.5">{classes.length} kelas · {slots.length} slot</p>
    </div>
    <div class="flex gap-2">
      <Button size="sm" variant="outline" onclick={cekBentrokan}>
        {computeConflicts ? '🔍 Refresh Bentrokan' : '🔍 Cek Bentrokan'}
      </Button>
    </div>
  </div>

  <!-- Lazy Conflict Alert -->
  {#if computeConflicts && conflicts.length > 0}
    <div class="rounded-xl border-2 border-destructive/20 bg-destructive/5 p-4 space-y-2">
      <div class="flex items-center gap-2 text-destructive">
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" /></svg>
        <h3 class="text-sm font-black">{conflicts.length} Bentrokan Jadwal</h3>
      </div>
      <div class="space-y-1">
        {#each conflicts as c, idx (idx)}
          <div class="flex items-start gap-2 text-xs px-1">
            <span class="shrink-0 mt-0.5">{c.type === 'teacher_overlap' ? '👨‍🏫' : c.type === 'class_overlap' ? '🏫' : '🚪'}</span>
            <div><span class="text-foreground font-medium">{dayName(c.day)}</span><span class="text-muted-foreground"> — {c.description}</span></div>
          </div>
        {/each}
      </div>
    </div>
  {:else if computeConflicts && slots.length > 0}
    <div class="rounded-xl border border-green-200 bg-green-50 p-4 text-center">
      <p class="text-sm font-semibold text-green-700">✅ Tidak ada bentrokan jadwal. Semua slot aman.</p>
    </div>
  {/if}

  <!-- Empty state -->
  {#if classes.length === 0}
    <Card.Root><Card.Content class="p-10 text-center space-y-2">
      <p class="text-2xl">📭</p>
      <p class="text-sm text-muted-foreground">Belum ada data jadwal. Pastikan assign guru sudah dilakukan.</p>
      <a href="/academic/subject-assignments" class="inline-block mt-2 text-xs font-bold text-primary hover:underline">→ Buka Assign Guru</a>
    </Card.Content></Card.Root>
  {:else}
    <!-- ── Pagination Navigation ── -->
    <div class="flex items-center justify-between gap-2">
      <Button size="sm" variant="outline" disabled={kelasIndex === 0} onclick={() => kelasIndex = Math.max(0, kelasIndex - 1)}>
        ← {kelasTerpilih ? kelasTerpilih.name : ''}
      </Button>
      <span class="text-[11px] font-medium text-muted-foreground">{kelasIndex + 1} / {classes.length}</span>
      <Button size="sm" variant="outline" disabled={kelasIndex >= maxKelasIndex} onclick={() => kelasIndex = Math.min(maxKelasIndex, kelasIndex + 1)}>
        {classes[kelasIndex + 1]?.name || 'Selesai'} →
      </Button>
    </div>

    <!-- Single active class card -->
    {#if kelasTerpilih}
      {@const cls = kelasTerpilih}
      {@const totalSlots = slots.filter(s => {
        const a = assignmentById.get(s.assignment_id);
        return a && a.class_id === cls.id;
      }).length}

      <Card.Root>
        <Card.Header class="px-4 pt-3 pb-1 flex flex-row items-center justify-between">
          <Card.Title class="text-sm font-semibold">{cls.name} <span class="text-muted-foreground font-normal">({cls.level})</span></Card.Title>
          <span class="text-[10px] text-muted-foreground">{totalSlots} slot</span>
        </Card.Header>
        <Card.Content class="px-4 pb-3">
          <!-- Desktop: Table -->
          <div class="hidden lg:block overflow-x-auto">
            <table class="w-full text-xs">
              <thead><tr class="bg-muted/30 text-muted-foreground"><th class="px-2 py-1.5 text-left w-10">Jam</th>
                {#each days as d}<th class="px-2 py-1.5 text-center border-l border-border">{d}</th>{/each}
              </tr></thead>
              <tbody class="divide-y divide-border">
                {#each Array(10) as _, period}
                  {@const rowNum = period + 1}
                  <tr class="hover:bg-muted/10">
                    <td class="px-2 py-1.5 text-muted-foreground whitespace-nowrap text-[9px] font-mono font-semibold align-top pt-2">#{rowNum}</td>
                    {#each days as _, dayIdx}
                      {@const day = dayIdx + 1}
                      {@const key = `${cls.id}|${day}`}
                      {@const daySlots = slotsByClassDay.get(key) ?? []}
                      {@const slot = daySlots[rowNum - 1]}
                      <td class="px-1 py-1 align-top border-l border-border {slot && conflictSlotIds.has(slot.id) ? 'bg-destructive/5' : ''}">
                        {#if slot}
                          <div class="rounded-md border px-1.5 py-1.5 {conflictSlotIds.has(slot.id) ? 'border-destructive/30 bg-destructive/10' : 'border-primary/10 bg-primary/5'}">
                            <div class="font-semibold text-[10px] leading-tight text-foreground">{slot.subject_name}</div>
                            <div class="text-[8px] text-muted-foreground mt-0.5">{slot.teacher_name}</div>
                            <div class="text-[8px] font-mono text-primary font-bold mt-0.5">{timeLabel(slot)}</div>
                            {#if slot.room_label}<div class="text-[8px] text-muted-foreground">📍{slot.room_label}</div>{/if}
                            <div class="flex gap-1.5 justify-end mt-1">
                              <button class="text-[8px] text-primary hover:underline font-semibold" onclick={() => openEdit(slot)}>Edit</button>
                              <button class="text-[8px] text-destructive hover:underline font-semibold" onclick={() => { deleteConfirmId = slot.id; deleteDialogOpen = true; }}>Hapus</button>
                            </div>
                          </div>
                        {:else}
                          <button class="w-full py-3 text-[9px] text-muted-foreground/40 hover:text-primary/60 transition-colors rounded hover:bg-primary/5" onclick={() => openAdd(day)}>+ Tambah</button>
                        {/if}
                      </td>
                    {/each}
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
          <!-- Mobile: Day-by-day list -->
          <div class="lg:hidden space-y-3">
            {#each days as d, idx}
              {@const dayNum = idx + 1}
              {@const key = `${cls.id}|${dayNum}`}
              {@const daySlots = (slotsByClassDay.get(key) ?? []).sort((a, b) => (a.start_time || '').localeCompare(b.start_time || ''))}
              <div>
                <h4 class="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-1.5 flex items-center gap-2">
                  {d}
                  {#if conflictSlotIds.size > 0 && daySlots.some(s => conflictSlotIds.has(s.id))}
                    <span class="text-destructive text-[9px]">⚠️ bentrok</span>
                  {/if}
                </h4>
                {#if daySlots.length === 0}
                  <button class="w-full text-left text-[10px] text-muted-foreground/50 italic px-1 py-1 hover:text-primary/70 transition-colors" onclick={() => openAdd(dayNum)}>+ Tambah jadwal</button>
                {:else}
                  <div class="space-y-1">
                    {#each daySlots as s (s.id)}
                      <div class="flex items-center justify-between gap-2 rounded-lg border px-2.5 py-2.5 {conflictSlotIds.has(s.id) ? 'border-destructive/30 bg-destructive/5' : 'border-border'}">
                        <div class="min-w-0 flex-1">
                          <div class="flex items-center gap-1.5">
                            <span class="text-xs font-semibold">{s.subject_name}</span>
                            {#if conflictSlotIds.has(s.id)}<span class="text-destructive text-[9px]">⚠️</span>{/if}
                          </div>
                          <div class="text-[10px] text-muted-foreground">{s.teacher_name}</div>
                          <div class="flex gap-2 mt-0.5">
                            <span class="text-[10px] font-mono text-primary font-bold">{timeLabel(s)}</span>
                            {#if s.room_label}<span class="text-[9px] text-muted-foreground">📍{s.room_label}</span>{/if}
                          </div>
                        </div>
                        <div class="flex gap-1.5 shrink-0">
                          <button class="text-[10px] text-primary hover:underline font-semibold" onclick={() => openEdit(s)}>Edit</button>
                          <button class="text-[10px] text-destructive hover:underline font-semibold" onclick={() => { deleteConfirmId = s.id; deleteDialogOpen = true; }}>Hapus</button>
                        </div>
                      </div>
                    {/each}
                  </div>
                {/if}
              </div>
            {/each}
          </div>
        </Card.Content>
      </Card.Root>
    {/if}
  {/if}

  <!-- Conflict summary table -->
  {#if computeConflicts && conflicts.length > 0}
    <div class="rounded-xl border border-border bg-card p-4 space-y-2">
      <h3 class="text-xs font-black text-foreground uppercase tracking-wider">📋 Ringkasan Bentrokan</h3>
      <table class="w-full text-xs">
        <thead><tr class="text-muted-foreground"><th class="text-left py-1 pr-2">Hari</th><th class="text-left py-1 pr-2">Jenis</th><th class="text-left py-1 pr-2">Detail</th><th class="text-left py-1">Jam</th></tr></thead>
        <tbody>
          {#each conflicts as c, idx (idx)}
            <tr class={idx % 2 === 0 ? '' : 'bg-muted/20'}>
              <td class="py-1.5 pr-2 font-medium">{dayName(c.day)}</td>
              <td class="py-1.5 pr-2">
                {#if c.type === 'teacher_overlap'}<Badge variant="destructive">Guru</Badge>
                {:else if c.type === 'class_overlap'}<Badge variant="destructive">Kelas</Badge>
                {:else}<Badge variant="destructive">Ruang</Badge>{/if}
              </td>
              <td class="py-1.5 pr-2 text-muted-foreground">{c.description}</td>
              <td class="py-1.5 font-mono text-primary font-bold">{c.time_range}</td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</div>

<!-- Slot Dialog -->
<Dialog.Root bind:open={showSlotDialog}>
  <Dialog.Content>
    <div class="space-y-4">
      <h2 class="text-base font-semibold">{editMode ? '✏️ Edit' : '➕ Tambah'} Slot Jadwal</h2>
      {#if slotError}
        <div class="rounded-lg bg-destructive/10 border border-destructive/20 px-3 py-2 text-xs font-medium text-destructive">{slotError}</div>
      {/if}
      <div class="grid gap-3 sm:grid-cols-2">
        <div class="space-y-1 sm:col-span-2">
          <label class="text-xs font-medium text-muted-foreground">Mata Pelajaran</label>
          <select bind:value={slotForm.assignment_id} class="flex h-9 w-full rounded-lg border border-input bg-background px-3 text-sm">
            <option value="">Pilih mapel & kelas</option>
            {#each assignments as a}
              <option value={a.id}>{a.class_code} — {a.subject_name} ({a.teacher_name})</option>
            {/each}
          </select>
        </div>
        <div class="space-y-1"><label class="text-xs font-medium text-muted-foreground">Hari</label>
          <select bind:value={slotForm.day_of_week} class="flex h-9 w-full rounded-lg border border-input bg-background px-3 text-sm">
            {#each days as d, idx}<option value={idx + 1}>{d}</option>{/each}
          </select>
        </div>
        <div class="space-y-1"><label class="text-xs font-medium text-muted-foreground">Jam Mulai</label>
          <Input type="time" bind:value={slotForm.start_time} />
        </div>
        <div class="space-y-1"><label class="text-xs font-medium text-muted-foreground">Jam Selesai</label>
          <Input type="time" bind:value={slotForm.end_time} />
        </div>
        <div class="space-y-1"><label class="text-xs font-medium text-muted-foreground">Ruang</label>
          <Input bind:value={slotForm.room_label} placeholder="Kelas VII.A" />
        </div>
      </div>
      <div class="flex justify-end gap-2 pt-2">
        <Button variant="outline" onclick={() => showSlotDialog = false}>Batal</Button>
        <LoadingButton onclick={() => void submitSlot()} loading={slotLoading}>{editMode ? 'Simpan Perubahan' : 'Tambah Slot'}</LoadingButton>
      </div>
    </div>
  </Dialog.Content>
</Dialog.Root>

<!-- Delete Confirmation -->
<Dialog.Root bind:open={deleteDialogOpen}>
  <Dialog.Content class="max-w-sm">
    <div class="space-y-4 text-center">
      <div class="text-3xl">🗑️</div>
      <h2 class="text-base font-semibold">Hapus Slot?</h2>
      <p class="text-xs text-muted-foreground">Slot yang dihapus tidak bisa dikembalikan.</p>
      <div class="flex justify-center gap-2 pt-2">
        <Button variant="outline" onclick={() => { deleteConfirmId = null; deleteDialogOpen = false; }}>Batal</Button>
        <LoadingButton variant="destructive" onclick={() => void executeDelete()} loading={deleteLoading}>Ya, Hapus</LoadingButton>
      </div>
    </div>
  </Dialog.Content>
</Dialog.Root>
