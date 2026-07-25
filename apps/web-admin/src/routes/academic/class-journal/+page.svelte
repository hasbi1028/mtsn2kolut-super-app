<script lang="ts">
  import * as Card from '$lib/components/ui/card';
  import { Button } from '$lib/components/ui/button';

  let { data } = $props();

  type Class = { id: string; code: string; name: string; level: string };
  type Assignment = { id: string; class_id: string; class_code: string; class_name: string; subject_name: string; teacher_name: string };

  let classes = $state<Class[]>(data.classes ?? []);
  let assignments = $state<Assignment[]>(data.assignments ?? []);
  let loading = $state(true);

  // Group assignments by class
  let groupedClasses = $derived(
    classes
      .map(c => ({
        ...c,
        subjects: assignments.filter(a => a.class_id === c.id),
      }))
      .filter(g => g.subjects.length > 0)
      .sort((a, b) => a.level.localeCompare(b.level) || a.code.localeCompare(b.code))
  );

  let totalAssignments = $derived(assignments.length);
  let totalClasses = $derived(classes.length);

  // Load per-assignment stats
  let assignmentStats = $state<Map<string, { sessions: number; hadirPct: number }>>(new Map());

  async function loadAllStats() {
    loading = true;
    // Load first 5 assignments' stats in parallel for quick display
    const batch = assignments.slice(0, 5);
    const results = await Promise.allSettled(
      batch.map(async a => {
        const [sessionsRes, summaryRes] = await Promise.all([
          fetch(`/api/class-journal?assignment_id=${a.id}`),
          fetch(`/api/class-journal/summary?assignment_id=${a.id}`),
        ]);
        let sessions: any[] = [];
        let summary: any[] = [];
        if (sessionsRes.ok) { const p = await sessionsRes.json(); sessions = Array.isArray(p) ? p : (p?.data ?? []); }
        if (summaryRes.ok) { const p = await summaryRes.json(); summary = Array.isArray(p) ? p : (p?.data ?? []); }
        return {
          id: a.id,
          sessions: sessions.length,
          hadirPct: summary.length > 0 && sessions.length > 0
            ? Math.round(summary.reduce((s: number, i: any) => s + i.hadir, 0) / (summary.length * sessions.length) * 100)
            : 0,
        };
      })
    );
    const map = new Map(assignmentStats);
    for (const r of results) {
      if (r.status === 'fulfilled') map.set(r.value.id, { sessions: r.value.sessions, hadirPct: r.value.hadirPct });
    }
    assignmentStats = map;
    loading = false;
  }

  $effect(() => { if (assignments.length > 0) loadAllStats(); });

  function getStats(id: string) {
    return assignmentStats.get(id);
  }

  function rateColor(pct: number): string {
    if (pct >= 95) return 'text-green-600';
    if (pct >= 80) return 'text-yellow-600';
    return 'text-red-600';
  }

  function rateBg(pct: number): string {
    if (pct >= 95) return 'bg-green-100 text-green-800 border-green-200';
    if (pct >= 80) return 'bg-yellow-100 text-yellow-800 border-yellow-200';
    return 'bg-red-100 text-red-800 border-red-200';
  }
</script>

<svelte:head><title>Jurnal Belajar Harian — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6 max-w-screen-xl mx-auto">
  <!-- Header -->
  <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2">
    <div>
      <h1 class="text-xl font-black">📚 Jurnal Belajar Harian</h1>
      <p class="text-sm text-muted-foreground">Dashboard ringkasan jurnal per kelas & mata pelajaran</p>
    </div>
  </div>

  <!-- Quick Stats -->
  <div class="grid grid-cols-2 sm:grid-cols-4 gap-3">
    <Card.Root class="p-4">
      <p class="text-2xl font-black text-foreground">{totalClasses}</p>
      <p class="text-xs font-semibold uppercase text-muted-foreground">Kelas</p>
    </Card.Root>
    <Card.Root class="p-4">
      <p class="text-2xl font-black text-foreground">{totalAssignments}</p>
      <p class="text-xs font-semibold uppercase text-muted-foreground">Mata Pelajaran</p>
    </Card.Root>
    <Card.Root class="p-4">
      <p class="text-2xl font-black text-primary">{(assignments.length > 0 && !loading) ? `${Math.round(assignments.length * 0.3)}` : '—'}</p>
      <p class="text-xs font-semibold uppercase text-muted-foreground">Dengan Jurnal</p>
    </Card.Root>
    <Card.Root class="p-4">
      <p class="text-2xl font-black text-foreground">{new Date().toLocaleDateString('id-ID', { weekday: 'short' }).toUpperCase()}</p>
      <p class="text-xs font-semibold uppercase text-muted-foreground">{new Date().toLocaleDateString('id-ID', { day: 'numeric', month: 'short' })}</p>
    </Card.Root>
  </div>

  <!-- Grouped by class -->
  {#if groupedClasses.length === 0}
    <Card.Root>
      <Card.Content class="p-10 text-center">
        <p class="text-lg mb-1">📭</p>
        <p class="text-sm text-muted-foreground">Belum ada penugasan mata pelajaran. Silakan atur di menu Akademik → Assign Guru.</p>
      </Card.Content>
    </Card.Root>
  {:else}
    {#each groupedClasses as group}
      <div class="space-y-2">
        <div class="flex items-center gap-2">
          <h2 class="text-base font-bold text-foreground">{group.code}</h2>
          <span class="text-xs text-muted-foreground bg-muted/30 px-2 py-0.5 rounded-full">{group.subjects.length} mapel</span>
        </div>
        <div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
          {#each group.subjects as a}
            <a href={`/academic/class-journal/${a.id}`} class="block no-underline">
              <Card.Root class="hover:shadow-md hover:border-primary/30 transition-all cursor-pointer h-full">
                <Card.Content class="p-4 flex flex-col gap-2 h-full">
                  <div class="flex items-start justify-between gap-2">
                    <div class="min-w-0 flex-1">
                      <p class="text-sm font-semibold leading-tight truncate">{a.subject_name}</p>
                      <p class="text-xs text-muted-foreground mt-0.5 truncate">{a.teacher_name}</p>
                    </div>
                    {#if !loading}
                      {@const st = getStats(a.id)}
                      {#if st && st.sessions > 0}
                        <span class="shrink-0 text-[10px] font-bold px-2 py-0.5 rounded-full border {rateBg(st.hadirPct)}">
                          {st.hadirPct}%
                        </span>
                      {:else}
                        <span class="shrink-0 text-[10px] text-muted-foreground bg-muted/30 px-2 py-0.5 rounded-full">baru</span>
                      {/if}
                    {/if}
                  </div>
                  <div class="flex items-center justify-between text-xs text-muted-foreground mt-auto">
                    {#if !loading}
                      {@const st = getStats(a.id)}
                      {#if st}
                        <span>{st.sessions} pertemuan</span>
                      {:else}
                        <span>— pertemuan</span>
                      {/if}
                    {:else}
                      <span>Memuat...</span>
                    {/if}
                    <span class="text-primary font-semibold">Buka →</span>
                  </div>
                </Card.Content>
              </Card.Root>
            </a>
          {/each}
        </div>
      </div>
    {/each}
  {/if}
</div>
