<script lang="ts">
  import { Badge } from '$lib/components/ui/badge';

  interface QueueStats {
    queued: number;
    running: number;
    success: number;
    failed: number;
    retry_due: number;
    total: number;
  }

  let { stats }: { stats: QueueStats } = $props();

  const items = $derived([
    { label: 'Antre', value: stats.queued, tone: '' },
    { label: 'Jalan', value: stats.running, tone: stats.running > 0 ? 'text-primary font-bold' : '' },
    { label: 'Sukses', value: stats.success, tone: stats.success > 0 ? 'text-success' : '' },
    { label: 'Gagal', value: stats.failed, tone: stats.failed > 0 ? 'text-error' : '' },
  ]);
</script>

<!-- Compact horizontal stat bar -->
<div class="flex items-center gap-1 overflow-x-auto rounded-xl border border-base-300 bg-base-100 px-3 py-2.5 text-center">
  {#each items as item (item.label)}
    <div class="flex min-w-[60px] flex-1 flex-col items-center gap-0.5 px-1">
      <span class="text-lg font-black leading-none {item.tone}">{item.value}</span>
      <span class="text-[9px] font-semibold tracking-wide text-base-content/60 uppercase">{item.label}</span>
    </div>
    {#if item.label !== 'Gagal'}
      <div class="h-6 w-px bg-base-300/60"></div>
    {/if}
  {/each}
  {#if stats.total > 0}
    <div class="ml-1 shrink-0">
      <Badge class="badge-xs bg-base-200 text-[9px] font-bold text-base-content/70">{stats.total} total</Badge>
    </div>
  {/if}
</div>
