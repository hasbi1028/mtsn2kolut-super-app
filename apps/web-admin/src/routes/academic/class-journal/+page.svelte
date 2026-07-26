<script lang="ts">
  import * as Card from '$lib/components/ui/card';
  import { journalFlow } from '$lib/stores/journal-flow.svelte';

  let { data } = $props();

  type Class = { id: string; code: string; name: string; level: string };

  let classes = $state<Class[]>(data.classes ?? []);

  let groupedClasses = $derived(
    ['VII', 'VIII', 'IX']
      .map(level => ({
        level,
        items: classes
          .filter(c => c.level === level)
          .sort((a, b) => a.code.localeCompare(b.code))
      }))
      .filter(g => g.items.length > 0)
  );

  function selectClass(c: Class) {
    $journalFlow = { ...$journalFlow, selectedClassId: c.id, selectedClassCode: c.code };
  }
</script>

<svelte:head><title>Pilih Kelas — Jurnal Belajar — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6 max-w-screen-xl mx-auto">
  <div>
    <h1 class="text-xl font-black">📚 Jurnal Belajar Harian</h1>
    <p class="text-sm text-muted-foreground">Pilih kelas untuk mulai mencatat jurnal dan absensi</p>
  </div>

  {#if groupedClasses.length === 0}
    <Card.Root>
      <Card.Content class="p-10 text-center">
        <p class="text-lg mb-1">📭</p>
        <p class="text-sm text-muted-foreground">Belum ada kelas tersedia.</p>
      </Card.Content>
    </Card.Root>
  {:else}
    {#each groupedClasses as group}
      <div class="space-y-2">
        <h2 class="text-base font-bold text-foreground">Kelas {group.level}</h2>
        <div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
          {#each group.items as c}
            <a href={`/academic/class-journal/kelas/${c.id}`} onclick={() => selectClass(c)} class="block no-underline">
              <Card.Root class="hover:shadow-md hover:border-primary/30 transition-all cursor-pointer h-full">
                <Card.Content class="p-5 flex flex-col items-center justify-center text-center gap-1 h-full">
                  <span class="text-3xl">🏫</span>
                  <p class="text-lg font-black text-foreground">{c.code}</p>
                  <p class="text-xs text-muted-foreground">{c.name}</p>
                  <span class="text-[10px] font-semibold text-primary mt-1">Pilih Mapel →</span>
                </Card.Content>
              </Card.Root>
            </a>
          {/each}
        </div>
      </div>
    {/each}
  {/if}
</div>
