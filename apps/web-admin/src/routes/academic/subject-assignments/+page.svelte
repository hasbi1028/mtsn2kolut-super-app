<script lang="ts">
  import { Badge } from '$lib/components/ui/badge';
  import { Button } from '$lib/components/ui/button';
  import * as Card from '$lib/components/ui/card';
  import * as Dialog from '$lib/components/ui/dialog';
  import { Input } from '$lib/components/ui/input';
  import { toast } from '$lib/components/ui/sonner';
  import LoadingButton from '$lib/components/LoadingButton.svelte';

  let { data } = $props();

  type Class = { id: string; code: string; name: string; level: string; };
  type Subject = { id: string; code: string; name: string; category: string; };
  type Teacher = { id: string; nip: string; nama: string; unit_kerja: string; };
  type Cell = { class_id: string; subject_id: string; assignment_id?: string; teacher_employee_id?: string; teacher_name: string; status: string; intra_weekly_hours: number; total_weekly_hours: number; compliance_status: string; };

  let classes = $state<Class[]>(data.classes ?? []);
  let subjects = $state<Subject[]>(data.subjects ?? []);
  let teachers = $state<Teacher[]>(data.teachers ?? []);
  let cells = $state<Cell[]>(data.cells ?? []);
  let loading = $state(false);

  // — Assign dialog
  let showAssign = $state(false);
  let assignClassId = $state('');
  let assignSubjectId = $state('');
  let assignSubjectName = $state('');
  let assignTeacherId = $state('');
  let assignLoading = $state(false);

  function getCell(classId: string, subjectId: string): Cell | undefined {
    return cells.find(c => c.class_id === classId && c.subject_id === subjectId);
  }

  function statusClass(status: string): string {
    switch (status) {
      case 'complete': return 'bg-primary/10 text-primary';
      case 'missing_assignment': return 'bg-yellow-50 text-yellow-700 border-yellow-200';
      case 'missing_teacher': return 'bg-orange-50 text-orange-700 border-orange-200';
      case 'missing_curriculum': return 'bg-red-50 text-red-700 border-red-200';
      default: return 'bg-muted/30 text-muted-foreground';
    }
  }

  function statusLabel(status: string): string {
    switch (status) {
      case 'complete': return '✓';
      case 'missing_assignment': return '?';
      case 'missing_teacher': return '!';
      case 'missing_curriculum': return '✗';
      default: return '—';
    }
  }

  function openAssign(classId: string, subjectId: string, subjectName: string) {
    assignClassId = classId;
    assignSubjectId = subjectId;
    assignSubjectName = subjectName;
    const cell = getCell(classId, subjectId);
    assignTeacherId = cell?.teacher_employee_id ?? '';
    showAssign = true;
  }

  async function submitAssign() {
    if (!assignTeacherId) { toast.error('Pilih guru'); return; }
    assignLoading = true;
    try {
      const r = await fetch('/api/academic/subject-assignments', {
        method: 'POST', headers: { 'content-type': 'application/json' },
        body: JSON.stringify({ class_id: assignClassId, subject_id: assignSubjectId, teacher_employee_id: assignTeacherId }),
      });
      if (r.ok) {
        toast.success('Guru ditugaskan');
        showAssign = false;
        await refreshData();
      } else { const e = await r.json().catch(() => ({})); toast.error(e?.error || 'Gagal'); }
    } catch { toast.error('Gagal'); }
    finally { assignLoading = false; }
  }

  async function deleteAssign(cell: Cell) {
    if (!cell.assignment_id || !confirm(`Hapus penugasan?`)) return;
    const r = await fetch(`/api/academic/subject-assignments/${cell.assignment_id}`, { method: 'DELETE' });
    if (r.ok || r.status === 204) { toast.success('Penugasan dihapus'); await refreshData(); }
    else toast.error('Gagal');
  }

  async function refreshData() {
    loading = true;
    try {
      const r = await fetch('/api/academic/subject-assignments');
      if (r.ok) { const p = await r.json(); const d = p.data ?? {}; classes = d.classes ?? []; subjects = d.subjects ?? []; teachers = d.teachers ?? []; cells = d.cells ?? []; }
    } catch {}
    finally { loading = false; }
  }

  let filterLevel = $state('');
  function filteredSubjects() {
    return subjects;
  }
  function filteredClasses() {
    if (!filterLevel) return classes;
    return classes.filter(c => c.level === filterLevel);
  }

  let teacherMap: Record<string, string> = $derived(
    Object.fromEntries(teachers.map(t => [t.id, t.nama]))
  );
</script>

<svelte:head><title>Assign Guru — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-4 max-w-screen-xl mx-auto">
  <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2">
    <div>
      <h1 class="text-xl font-black">Assign Guru</h1>
      <p class="text-sm text-muted-foreground">Tentukan guru pengampu setiap mapel per rombel.</p>
    </div>
    {#if loading}
      <span class="text-sm text-muted-foreground">Memuat...</span>
    {/if}
  </div>

  <!-- Legend -->
  <div class="flex flex-wrap gap-3 text-xs text-muted-foreground">
    <span><span class="inline-block w-4 h-4 rounded bg-primary/10 text-primary text-center leading-4 mr-1">✓</span> Lengkap</span>
    <span><span class="inline-block w-4 h-4 rounded bg-yellow-50 text-yellow-700 border border-yellow-200 text-center leading-4 mr-1">?</span> Belum assign</span>
    <span><span class="inline-block w-4 h-4 rounded bg-orange-50 text-orange-700 border border-orange-200 text-center leading-4 mr-1">!</span> Guru tidak aktif</span>
    <span><span class="inline-block w-4 h-4 rounded bg-red-50 text-red-700 border border-red-200 text-center leading-4 mr-1">✗</span> Belum ada kurikulum</span>
  </div>

  <!-- Filter tingkat -->
  <div class="flex gap-1 flex-wrap">
    <button class="px-2.5 py-1 text-xs font-medium rounded-md transition-colors {!filterLevel ? 'bg-primary text-primary-foreground' : 'border border-border text-muted-foreground'}" onclick={() => filterLevel = ''}>Semua</button>
    {#each ['VII', 'VIII', 'IX'] as lv}
      <button class="px-2.5 py-1 text-xs font-medium rounded-md transition-colors {filterLevel === lv ? 'bg-primary text-primary-foreground' : 'border border-border text-muted-foreground'}" onclick={() => filterLevel = lv}>{lv}</button>
    {/each}
  </div>

  {#if classes.length === 0 || subjects.length === 0}
    <Card.Root><Card.Content class="p-8 text-center">
      <p class="text-sm text-muted-foreground">
        {#if classes.length === 0}Tidak ada rombel aktif.{:else}Tidak ada mapel terdaftar.{/if}
      </p>
    </Card.Content></Card.Root>
  {:else}
    <!-- Desktop: table matrix -->
    <div class="hidden lg:block overflow-x-auto rounded-lg border border-border">
      <table class="w-full text-sm border-collapse">
        <thead>
          <tr>
            <th class="sticky left-0 z-10 bg-muted/30 px-3 py-2 text-left text-xs font-semibold text-muted-foreground uppercase min-w-[140px]">Mapel</th>
            {#each filteredClasses() as c (c.id)}
              <th class="px-2 py-2 text-center text-xs font-semibold text-muted-foreground uppercase min-w-[120px] border-l border-border">
                <div>{c.code}</div>
                <div class="text-[10px] opacity-60">{c.level}</div>
              </th>
            {/each}
          </tr>
        </thead>
        <tbody class="divide-y divide-border">
          {#each filteredSubjects() as s (s.id)}
            <tr class="hover:bg-muted/10">
              <td class="sticky left-0 z-10 bg-white px-3 py-2 text-xs font-medium text-foreground border-r border-border whitespace-nowrap">
                <span class="text-[10px] text-muted-foreground">{s.code}</span>
                <span class="ml-1">{s.name}</span>
              </td>
              {#each filteredClasses() as c (c.id)}
                {@const cell = getCell(c.id, s.id)}
                <td class="px-2 py-1.5 text-center border-l border-border">
                  <button type="button" class="w-full text-left rounded-md px-2 py-1.5 text-xs transition-colors {statusClass(cell?.status ?? '')}" onclick={() => openAssign(c.id, s.id, s.name)}>
                    {#if cell?.status === 'complete'}
                      <span class="font-medium">{cell?.teacher_name || '—'}</span>
                      {#if cell.total_weekly_hours > 0}<span class="ml-1 text-[10px] opacity-60">{cell.total_weekly_hours}jp</span>{/if}
                    {:else}
                      <span class="italic opacity-70">{statusLabel(cell?.status ?? 'missing_assignment')}</span>
                    {/if}
                  </button>
                </td>
              {/each}
            </tr>
          {/each}
        </tbody>
      </table>
    </div>

    <!-- Mobile: list per class -->
    <div class="lg:hidden space-y-3">
      {#each filteredClasses() as c (c.id)}
        <Card.Root>
          <Card.Header class="px-4 pt-3 pb-1"><Card.Title class="text-sm font-semibold">{c.name} <span class="text-muted-foreground font-normal">({c.level})</span></Card.Title></Card.Header>
          <Card.Content class="px-4 pb-3 space-y-1">
            {#each filteredSubjects() as s (s.id)}
              {@const cell = getCell(c.id, s.id)}
              <div class="flex items-center justify-between gap-2 py-1.5 border-b border-border last:border-0">
                <div class="min-w-0 flex-1">
                  <span class="text-xs font-medium">{s.name}</span>
                  {#if cell?.status === 'complete'}<span class="text-[10px] text-muted-foreground ml-1">({cell?.teacher_name})</span>{/if}
                </div>
                <button class="text-xs text-primary hover:underline shrink-0" onclick={() => openAssign(c.id, s.id, s.name)}>
                  {#if cell?.status === 'complete'}Ganti{:else}Assign{/if}
                </button>
              </div>
            {/each}
          </Card.Content>
        </Card.Root>
      {/each}
    </div>
  {/if}
</div>

<!-- Dialog Assign -->
<Dialog.Root bind:open={showAssign}>
  <Dialog.Content>
    <div class="space-y-4">
      <h2 class="text-base font-semibold">Assign Guru</h2>
      <p class="text-sm text-muted-foreground">{assignSubjectName}</p>

      <div class="space-y-2 max-h-60 overflow-y-auto">
        <label class="text-xs font-medium text-muted-foreground">Pilih Guru</label>
        {#each teachers as t (t.id)}
          <label class="flex items-center gap-2 p-2 rounded-lg cursor-pointer transition-colors hover:bg-muted/20 {assignTeacherId === t.id ? 'bg-primary/5 border border-primary/20' : 'border border-transparent'}" role="radio" aria-checked={assignTeacherId === t.id} onclick={() => assignTeacherId = t.id}>
            <input type="radio" name="teacher" value={t.id} checked={assignTeacherId === t.id} onchange={() => assignTeacherId = t.id} class="hidden" />
            <div class="w-3.5 h-3.5 rounded-full border-2 flex items-center justify-center {assignTeacherId === t.id ? 'border-primary' : 'border-muted-foreground'}">
              {#if assignTeacherId === t.id}<div class="w-2 h-2 rounded-full bg-primary"></div>{/if}
            </div>
            <div>
              <p class="text-sm font-medium">{t.nama}</p>
              {#if t.nip}<p class="text-xs text-muted-foreground">{t.nip}</p>{/if}
            </div>
          </label>
        {/each}
      </div>

      <div class="flex justify-end gap-2 pt-2">
        <Button variant="outline" onclick={() => (showAssign = false)}>Batal</Button>
        <LoadingButton onclick={() => void submitAssign()} loading={assignLoading}>Simpan</LoadingButton>
      </div>
    </div>
  </Dialog.Content>
</Dialog.Root>
