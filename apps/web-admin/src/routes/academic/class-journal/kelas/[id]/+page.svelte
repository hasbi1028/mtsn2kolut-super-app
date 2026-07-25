<script lang="ts">
  import * as Card from '$lib/components/ui/card';

  let { data } = $props();

  type Assignment = { id: string; class_id: string; class_code: string; class_name: string; subject_name: string; teacher_name: string };
  type Kelas = { id: string; code: string; name: string; level: string };

  let kelas = $state<Kelas | null>(data.kelas);
  let mapels = $state<Assignment[]>(data.mapels ?? []);
</script>

<svelte:head><title>Pilih Mapel {kelas?.code ?? ''} — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6 max-w-screen-xl mx-auto">
  <!-- Header -->
  <div>
    <a href="/academic/class-journal" class="text-sm text-primary hover:underline">&larr; Pilih Kelas</a>
    {#if kelas}
      <h1 class="text-xl font-black mt-1">🏫 {kelas.code} — {kelas.name}</h1>
      <p class="text-sm text-muted-foreground">Pilih mata pelajaran untuk melihat jurnal dan absensi</p>
    {/if}
  </div>

  {#if mapels.length === 0}
    <Card.Root>
      <Card.Content class="p-10 text-center">
        <p class="text-lg mb-1">📭</p>
        <p class="text-sm text-muted-foreground">Belum ada mata pelajaran untuk kelas ini. Silakan atur di menu Akademik → Assign Guru.</p>
      </Card.Content>
    </Card.Root>
  {:else}
    <div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
      {#each mapels as m}
        <a href={`/academic/class-journal/kelas/${kelas?.id}/${m.id}`} class="block no-underline">
          <Card.Root class="hover:shadow-md hover:border-primary/30 transition-all cursor-pointer h-full">
            <Card.Content class="p-4 flex flex-col gap-2 h-full">
              <div class="flex items-start gap-3">
                <span class="text-2xl shrink-0">📖</span>
                <div class="min-w-0 flex-1">
                  <p class="text-sm font-semibold leading-tight">{m.subject_name}</p>
                  <p class="text-xs text-muted-foreground mt-0.5">{m.teacher_name}</p>
                </div>
              </div>
              <div class="mt-auto flex justify-end">
                <span class="text-[10px] font-semibold text-primary">Buka Jurnal →</span>
              </div>
            </Card.Content>
          </Card.Root>
        </a>
      {/each}
    </div>
  {/if}
</div>
