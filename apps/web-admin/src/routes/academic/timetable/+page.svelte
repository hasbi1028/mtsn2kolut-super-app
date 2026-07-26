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
  const dayMap: Record<number, number> = { 1: 0, 2: 1, 3: 2, 4: 3, 5: 4, 6: 5 };

  let classes = $state<{ id: string; code: string; name: string; level: string }[]>(data.classes ?? []);
  let subjects = $state<any[]>(data.subjects ?? []);
  let teachers = $state<any[]>(data.teachers ?? []);
  let assignments = $state<any[]>(data.assignments ?? []);

  // ── Dialog form
  let showSlotDialog = $state(false);
  let editMode = $state(false);
  let slotForm = $state({
    assignment_id: '', day_of_week: 1, start_time: '07:00', end_time: '07:40',
    room_label: '', slot_type: 'pelajaran', lesson_hours: 1,
  });
  let slotLoading = $state(false);

  // Helper: format time range
  function timeLabel(s: Slot): string {
    if (!s.start_time && !s.end_time) return '—';
    const st = s.start_time?.slice(0, 5) || '??';
    const et = s.end_time?.slice(0, 5) || '??';
    return `${st}–${et}`;
  }

  function classForAssignment(assignmentId: string) {
    const a = assignments.find(as => as.id === assignmentId);
    if (!a) return null;
    return classes.find(c => c.id === a.class_id);
  }

  function openAdd() {
    editMode = false;
    slotForm = { assignment_id: '', day_of_week: 1, start_time: '07:00', end_time: '07:40', room_label: '', slot_type: 'pelajaran', lesson_hours: 1 };
    showSlotDialog = true;
  }
  function openEdit(slot: Slot) {
    editMode = true;
    slotForm = {
      assignment_id: slot.assignment_id, day_of_week: slot.day_of_week,
      start_time: slot.start_time, end_time: slot.end_time,
      room_label: slot.room_label, slot_type: 'pelajaran', lesson_hours: 1,
    };
    showSlotDialog = true;
  }

  async function submitSlot() {
    slotLoading = true;
    try {
      const r = await fetch('/api/academic/timetable/slots', {
        method: 'POST', headers: { 'content-type': 'application/json' },
        body: JSON.stringify(slotForm),
      });
      if (r.ok) { toast.success('Slot tersimpan'); showSlotDialog = false; await refresh(); }
      else { const e = await r.json().catch(() => ({})); toast.error(e?.error || 'Gagal'); }
    } catch { toast.error('Gagal'); }
    finally { slotLoading = false; }
  }

  async function deleteSlot(id: string) {
    if (!confirm('Hapus slot ini?')) return;
    const r = await fetch(`/api/academic/timetable/slots/${id}`, { method: 'DELETE' });
    if (r.ok || r.status === 204) { toast.success('Slot dihapus'); await refresh(); }
    else toast.error('Gagal');
  }

  async function refresh() {
    loading = true;
    try {
      const r = await fetch('/api/academic/timetable/weekly');
      if (r.ok) { const p = await r.json(); const d = p.data ?? {}; classes = d.classes ?? []; subjects = d.subjects ?? []; teachers = d.teachers ?? []; assignments = d.assignments ?? []; slots = d.slots ?? []; }
    } catch {}
    finally { loading = false; }
  }

  // ── Conflict detection ──

  type Conflict = {
    type: 'teacher_overlap' | 'class_overlap' | 'room_overlap';
    teacher_name?: string;
    class_name?: string;
    day: number;
    time_range: string;
    slots: Slot[];
    description: string;
  };

  let conflicts = $derived.by(() => {
    const result: Conflict[] = [];

    // Group slots by (day, teacher_name)
    const byTeacher: Record<string, Slot[]> = {};
    // Group by (day, class_code)
    const byClass: Record<string, Slot[]> = {};
    // Group by (day, room)
    const byRoom: Record<string, Slot[]> = {};

    function groupKey(day: number, name: string): string {
      return `${day}|${name}`;
    }

    function timeOverlap(a: Slot, b: Slot): boolean {
      if (!a.start_time || !a.end_time || !b.start_time || !b.end_time) return false;
      return a.start_time < b.end_time && b.start_time < a.end_time;
    }

    for (const s of slots) {
      if (!s.start_time || !s.end_time) continue;

      // Teacher overlap
      if (s.teacher_name) {
        const k = groupKey(s.day_of_week, s.teacher_name);
        if (!byTeacher[k]) byTeacher[k] = [];
        byTeacher[k].push(s);
      }

      // Class overlap (two subjects at same time)
      const classCode = s.class_code;
      if (classCode) {
        const k = groupKey(s.day_of_week, classCode);
        if (!byClass[k]) byClass[k] = [];
        byClass[k].push(s);
      }

      // Room overlap
      if (s.room_label) {
        const k = groupKey(s.day_of_week, s.room_label);
        if (!byRoom[k]) byRoom[k] = [];
        byRoom[k].push(s);
      }
    }

    // Check teacher overlaps
    for (const [key, items] of Object.entries(byTeacher)) {
      for (let i = 0; i < items.length; i++) {
        for (let j = i + 1; j < items.length; j++) {
          if (timeOverlap(items[i], items[j])) {
            const day = Number(key.split('|')[0]);
            result.push({
              type: 'teacher_overlap',
              teacher_name: items[i].teacher_name,
              day,
              time_range: `${timeLabel(items[i])} / ${timeLabel(items[j])}`,
              slots: [items[i], items[j]],
              description: `⚠️ ${items[i].teacher_name} mengajar ${items[i].subject_name} (${items[i].class_code}) dan ${items[j].subject_name} (${items[j].class_code}) di jam yang sama!`,
            });
          }
        }
      }
    }

    // Check class overlaps
    for (const [key, items] of Object.entries(byClass)) {
      for (let i = 0; i < items.length; i++) {
        for (let j = i + 1; j < items.length; j++) {
          if (timeOverlap(items[i], items[j])) {
            const day = Number(key.split('|')[0]);
            result.push({
              type: 'class_overlap',
              class_name: items[i].class_code,
              day,
              time_range: `${timeLabel(items[i])} / ${timeLabel(items[j])}`,
              slots: [items[i], items[j]],
              description: `⚠️ ${items[i].class_code} diajar ${items[i].subject_name} (${items[i].teacher_name}) dan ${items[j].subject_name} (${items[j].teacher_name}) di jam yang sama!`,
            });
          }
        }
      }
    }

    // Check room overlaps
    for (const [key, items] of Object.entries(byRoom)) {
      for (let i = 0; i < items.length; i++) {
        for (let j = i + 1; j < items.length; j++) {
          if (timeOverlap(items[i], items[j])) {
            // Only flag if same room, different class
            if (items[i].class_code !== items[j].class_code) {
              const day = Number(key.split('|')[0]);
              result.push({
                type: 'room_overlap',
                class_name: items[i].class_code,
                day,
                time_range: `${timeLabel(items[i])} / ${timeLabel(items[j])}`,
                slots: [items[i], items[j]],
                description: `⚠️ Ruang ${items[i].room_label} dipakai ${items[i].class_code} (${items[i].subject_name}) dan ${items[j].class_code} (${items[j].subject_name}) di jam yang sama!`,
              });
            }
          }
        }
      }
    }

    return result;
  });

  // Set of slot IDs that are in conflict
  const conflictSlotIds = $derived(new Set(conflicts.flatMap(c => c.slots.map(s => s.id))));

  function dayName(day: number): string {
    return days[day - 1] || '?';
  }
</script>

<svelte:head><title>Jadwal Pelajaran — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-4 max-w-screen-xl mx-auto">
  <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2">
    <div><h1 class="text-xl font-black">📅 Jadwal Pelajaran</h1><p class="text-sm text-muted-foreground">Atur jadwal mingguan per rombel.</p></div>
  </div>

  <!-- ── Conflict Alert ── -->
  {#if conflicts.length > 0}
    <div class="rounded-xl border-2 border-destructive/20 bg-destructive/5 p-4 space-y-2">
      <div class="flex items-center gap-2 text-destructive">
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" /></svg>
        <h3 class="text-sm font-black">{conflicts.length} Bentrokan Jadwal Ditemukan</h3>
      </div>
      <div class="space-y-1.5">
        {#each conflicts as c, idx (idx)}
          <div class="flex items-start gap-2 text-xs px-1">
            <span class="shrink-0 mt-0.5">
              {#if c.type === 'teacher_overlap'}👨‍🏫
              {:else if c.type === 'class_overlap'}🏫
              {:else}🚪
              {/if}
            </span>
            <div>
              <span class="text-foreground font-medium">{dayName(c.day)}</span>
              <span class="text-muted-foreground"> — {c.description}</span>
            </div>
          </div>
        {/each}
      </div>
    </div>
  {/if}

  {#if classes.length === 0}
    <Card.Root><Card.Content class="p-8 text-center"><p class="text-sm text-muted-foreground">Belum ada data jadwal. Pastikan assign guru sudah dilakukan.</p></Card.Content></Card.Root>
  {:else}
    <div class="space-y-6">
      {#each classes as cls (cls.id)}
        <Card.Root>
          <Card.Header class="px-4 pt-3 pb-1 flex flex-row items-center justify-between">
            <Card.Title class="text-sm font-semibold">{cls.name} <span class="text-muted-foreground font-normal">({cls.level})</span></Card.Title>
            <span class="text-[10px] text-muted-foreground">{slots.filter(s => s.assignment_id && assignments.find(a => a.id === s.assignment_id)?.class_id === cls.id).length} slot</span>
          </Card.Header>
          <Card.Content class="px-4 pb-3">
            <!-- Desktop: Table -->
            <div class="hidden lg:block overflow-x-auto">
              <table class="w-full text-xs">
                <thead><tr class="bg-muted/30 text-muted-foreground"><th class="px-2 py-1.5 text-left">Jam</th>{#each days as d}<th class="px-2 py-1.5 text-center border-l border-border">{d}</th>{/each}</tr></thead>
                <tbody class="divide-y divide-border">
                  {#each Array(10) as _, period}
                    {@const rowNum = period + 1}
                    <tr class="hover:bg-muted/10">
                      <td class="px-2 py-1.5 text-muted-foreground whitespace-nowrap text-[9px] font-mono font-semibold align-top pt-2">
                        #{rowNum}
                      </td>
                      {#each days as _, dayIdx}
                        {@const day = dayIdx + 1}
                        {@const daySlots = slots.filter(s => s.assignment_id && assignments.find(a => a.id === s.assignment_id)?.class_id === cls.id && s.day_of_week === day)}
                        {@const slot = daySlots[rowNum - 1]}
                        <td class="px-1 py-1 align-top border-l border-border {slot && conflictSlotIds.has(slot.id) ? 'bg-destructive/5' : ''}">
                          {#if slot}
                            <div class="rounded-md border px-1.5 py-1.5 {conflictSlotIds.has(slot.id) ? 'border-destructive/30 bg-destructive/10' : 'border-primary/10 bg-primary/5'}">
                              <div class="font-semibold text-[10px] leading-tight text-foreground">{slot.subject_name}</div>
                              <div class="text-[8px] text-muted-foreground mt-0.5">{slot.teacher_name}</div>
                              <div class="text-[8px] font-mono text-primary font-bold mt-0.5">
                                {timeLabel(slot)}
                              </div>
                              {#if slot.room_label}
                                <div class="text-[8px] text-muted-foreground">📍{slot.room_label}</div>
                              {/if}
                              <div class="flex gap-1.5 justify-end mt-1">
                                <button class="text-[8px] text-primary hover:underline font-semibold" onclick={() => openEdit(slot)}>Edit</button>
                                <button class="text-[8px] text-destructive hover:underline font-semibold" onclick={() => deleteSlot(slot.id)}>Hapus</button>
                              </div>
                            </div>
                          {:else}
                            <button class="w-full py-3 text-[9px] text-muted-foreground/40 hover:text-primary/60 transition-colors rounded hover:bg-primary/5" onclick={() => { slotForm.day_of_week = day; openAdd(); }}>+ Tambah</button>
                          {/if}
                        </td>
                      {/each}
                    </tr>
                  {/each}
                </tbody>
              </table>
            </div>
            <!-- Mobile: day list -->
            <div class="lg:hidden space-y-3">
              {#each days as d, idx}
                {@const dayNum = idx + 1}
                {@const daySlots = slots.filter(s => s.assignment_id && assignments.find(a => a.id === s.assignment_id)?.class_id === cls.id && s.day_of_week === dayNum).sort((a, b) => (a.start_time || '').localeCompare(b.start_time || ''))}
                <div>
                  <h4 class="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-1.5 flex items-center gap-2">
                    {d}
                    {#if daySlots.some(s => conflictSlotIds.has(s.id))}
                      <span class="text-destructive text-[9px]">⚠️ bentrok</span>
                    {/if}
                  </h4>
                  {#if daySlots.length === 0}
                    <p class="text-[10px] text-muted-foreground/50 italic px-1">—</p>
                  {:else}
                    <div class="space-y-1">
                      {#each daySlots as s (s.id)}
                        <div class="flex items-center justify-between gap-2 rounded-lg border px-2.5 py-2 {conflictSlotIds.has(s.id) ? 'border-destructive/30 bg-destructive/5' : 'border-border'}">
                          <div class="min-w-0 flex-1">
                            <div class="flex items-center gap-1.5">
                              <span class="text-xs font-semibold">{s.subject_name}</span>
                              {#if conflictSlotIds.has(s.id)}
                                <span class="text-destructive text-[9px]">⚠️</span>
                              {/if}
                            </div>
                            <div class="text-[10px] text-muted-foreground">{s.teacher_name}</div>
                            <div class="text-[10px] font-mono text-primary font-bold">{timeLabel(s)}</div>
                            {#if s.room_label}
                              <div class="text-[9px] text-muted-foreground">📍{s.room_label}</div>
                            {/if}
                          </div>
                          <button class="text-xs text-destructive hover:underline shrink-0" onclick={() => deleteSlot(s.id)}>Hapus</button>
                        </div>
                      {/each}
                    </div>
                  {/if}
                </div>
              {/each}
            </div>
          </Card.Content>
        </Card.Root>
      {/each}
    </div>
  {/if}

  <!-- Show all conflicts summary at bottom -->
  {#if conflicts.length > 0}
    <div class="rounded-xl border border-border bg-card p-4 space-y-2">
      <h3 class="text-xs font-black text-foreground uppercase tracking-wider">📋 Ringkasan Bentrokan</h3>
      <table class="w-full text-xs">
        <thead><tr class="text-muted-foreground"><th class="text-left py-1 pr-2">Hari</th><th class="text-left py-1 pr-2">Jenis</th><th class="text-left py-1 pr-2">Detail</th><th class="text-left py-1">Jam</th></tr></thead>
        <tbody>
          {#each conflicts as c, idx (idx)}
            <tr class="{idx % 2 === 0 ? '' : 'bg-muted/20'}">
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
  {:else if slots.length > 0}
    <div class="rounded-xl border border-green-200 bg-green-50 p-4 text-center">
      <p class="text-sm font-semibold text-green-700">✅ Tidak ada bentrokan jadwal. Semua slot aman.</p>
    </div>
  {/if}
</div>

<!-- Dialog -->
<Dialog.Root bind:open={showSlotDialog}>
  <Dialog.Content><div class="space-y-4"><h2 class="text-base font-semibold">{editMode ? 'Edit' : 'Tambah'} Slot Jadwal</h2>
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
      <LoadingButton onclick={() => void submitSlot()} loading={slotLoading}>Simpan</LoadingButton>
    </div>
  </div></Dialog.Content>
</Dialog.Root>
