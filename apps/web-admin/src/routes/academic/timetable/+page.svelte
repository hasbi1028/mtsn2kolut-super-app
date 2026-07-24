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

  // Group slots by class then by day
  let classes = $state<{ id: string; code: string; name: string; level: string }[]>(data.classes ?? []);
  let subjects = $state<any[]>(data.subjects ?? []);
  let teachers = $state<any[]>(data.teachers ?? []);
  let assignments = $state<any[]>(data.assignments ?? []);

  // — Dialog tambah/edit slot
  let showSlotDialog = $state(false);
  let editMode = $state(false);
  let slotForm = $state({
    assignment_id: '', day_of_week: 1, start_time: '07:00', end_time: '07:40',
    room_label: '', slot_type: 'pelajaran', lesson_hours: 1,
  });
  let slotLoading = $state(false);

  function slotsForClassDay(classId: string, day: number): Slot[] {
    return slots.filter(s => dayMap[s.day_of_week] === day && s.assignment_id && assignments.find(a => a.id === s.assignment_id)?.class_id === classId);
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
      const url = editMode ? '' : '/api/academic/timetable/slots';
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

  // Group slots by class for easier rendering
  let classSlots = $derived.by(() => {
    const map: Record<string, Slot[]> = {};
    for (const s of slots) {
      const cls = classForAssignment(s.assignment_id);
      if (cls) {
        if (!map[cls.code]) map[cls.code] = [];
        map[cls.code].push(s);
      }
    }
    return map;
  });
</script>

<svelte:head><title>Jadwal Pelajaran — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-4 max-w-screen-xl mx-auto">
  <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2">
    <div><h1 class="text-xl font-black">Jadwal Pelajaran</h1><p class="text-sm text-muted-foreground">Atur jadwal mingguan per rombel.</p></div>
  </div>

  {#if classes.length === 0}
    <Card.Root><Card.Content class="p-8 text-center"><p class="text-sm text-muted-foreground">Belum ada data jadwal. Pastikan assign guru sudah dilakukan.</p></Card.Content></Card.Root>
  {:else}
    <!-- Per-class timetable cards -->
    <div class="space-y-6">
      {#each classes as cls (cls.id)}
        <Card.Root>
          <Card.Header class="px-4 pt-3 pb-1 flex flex-row items-center justify-between">
            <Card.Title class="text-sm font-semibold">{cls.name} <span class="text-muted-foreground font-normal">({cls.level})</span></Card.Title>
          </Card.Header>
          <Card.Content class="px-4 pb-3">
            <!-- Desktop: Table per class -->
            <div class="hidden lg:block overflow-x-auto">
              <table class="w-full text-xs">
                <thead><tr class="bg-muted/30 text-muted-foreground"><th class="px-2 py-1.5 text-left">Jam</th>{#each days as d}<th class="px-2 py-1.5 text-center border-l border-border">{d}</th>{/each}</tr></thead>
                <tbody class="divide-y divide-border">
                  <!-- 10 rows for periods -->
                  {#each Array(10) as _, period}
                    {@const rowNum = period + 1}
                    <tr class="hover:bg-muted/10">
                      <td class="px-2 py-1.5 text-muted-foreground whitespace-nowrap">#{rowNum}</td>
                      {#each [1,2,3,4,5,6] as day}
                        {@const daySlots = slots.filter(s => s.assignment_id && assignments.find(a => a.id === s.assignment_id)?.class_id === cls.id && s.day_of_week === day)}
                        {@const slot = daySlots[rowNum - 1]}
                        <td class="px-1.5 py-1 text-center border-l border-border">
                          {#if slot}
                            <div class="rounded bg-primary/5 border border-primary/10 px-1.5 py-1">
                              <div class="font-medium text-[10px] leading-tight">{slot.subject_name}</div>
                              <div class="text-[8px] text-muted-foreground truncate">{slot.teacher_name}</div>
                              <div class="flex gap-1 justify-center mt-0.5">
                                <button class="text-[8px] text-primary hover:underline" onclick={() => openEdit(slot)}>Edit</button>
                                <button class="text-[8px] text-destructive hover:underline" onclick={() => deleteSlot(slot.id)}>Hapus</button>
                              </div>
                            </div>
                          {:else}
                            <button class="w-full py-2 text-[9px] text-muted-foreground/40 hover:text-primary/60 transition-colors rounded hover:bg-primary/5" onclick={() => { slotForm.day_of_week = day; openAdd(); }}>+</button>
                          {/if}
                        </td>
                      {/each}
                    </tr>
                  {/each}
                </tbody>
              </table>
            </div>
            <!-- Mobile: day-by-day list -->
            <div class="lg:hidden space-y-3">
              {#each days as d, idx}
                {@const dayNum = idx + 1}
                {@const daySlots = slots.filter(s => s.assignment_id && assignments.find(a => a.id === s.assignment_id)?.class_id === cls.id && s.day_of_week === dayNum)}
                <div>
                  <h4 class="text-xs font-semibold text-muted-foreground uppercase mb-1">{d}</h4>
                  {#if daySlots.length === 0}
                    <p class="text-[10px] text-muted-foreground/50 italic">Belum ada jadwal</p>
                  {:else}
                    <div class="space-y-1">
                      {#each daySlots as s (s.id)}
                        <div class="flex items-center justify-between gap-2 rounded-md border border-border px-2.5 py-1.5">
                          <div class="min-w-0 flex-1">
                            <span class="text-xs font-medium">{s.subject_name}</span>
                            <span class="text-[10px] text-muted-foreground ml-1">({s.teacher_name})</span>
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
</div>

<!-- Dialog tambah slot -->
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
