<script lang="ts">
  import * as Card from '$lib/components/ui/card';
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
</script>

<div class="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-6">
  <Card.Root>
    <Card.Content class="pt-4 pb-3">
      <p class="text-xs text-muted-foreground">Antre</p>
      <p class="text-2xl font-bold mt-1">{stats.queued}</p>
    </Card.Content>
  </Card.Root>
  <Card.Root class={stats.running > 0 ? 'border-green-300' : ''}>
    <Card.Content class="pt-4 pb-3">
      <p class="text-xs text-muted-foreground">Berjalan</p>
      <div class="flex items-center gap-2 mt-1">
        <p class="text-2xl font-bold">{stats.running}</p>
        {#if stats.running > 0}
          <Badge class="text-xs">Aktif</Badge>
        {/if}
      </div>
    </Card.Content>
  </Card.Root>
  <Card.Root>
    <Card.Content class="pt-4 pb-3">
      <p class="text-xs text-muted-foreground">Sukses</p>
      <p class="text-2xl font-bold text-green-700 mt-1">{stats.success}</p>
    </Card.Content>
  </Card.Root>
  <Card.Root class={stats.failed > 0 ? 'border-red-200' : ''}>
    <Card.Content class="pt-4 pb-3">
      <p class="text-xs text-muted-foreground">Gagal</p>
      <p class="text-2xl font-bold {stats.failed > 0 ? 'text-destructive' : ''} mt-1">{stats.failed}</p>
    </Card.Content>
  </Card.Root>
  <Card.Root>
    <Card.Content class="pt-4 pb-3">
      <p class="text-xs text-muted-foreground">Retry</p>
      <p class="text-2xl font-bold mt-1">{stats.retry_due}</p>
    </Card.Content>
  </Card.Root>
  <Card.Root>
    <Card.Content class="pt-4 pb-3">
      <p class="text-xs text-muted-foreground">Total</p>
      <p class="text-2xl font-bold mt-1">{stats.total}</p>
    </Card.Content>
  </Card.Root>
</div>
