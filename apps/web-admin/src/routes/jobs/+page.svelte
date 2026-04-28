<script lang="ts">
  import { onMount } from 'svelte';
  import * as Card from '$lib/components/ui/card';
  import * as Table from '$lib/components/ui/table';
  import { Badge } from '$lib/components/ui/badge';
  import { Button } from '$lib/components/ui/button';

  let jobs         = $state<any[]>([]);
  let filterStatus = $state('');
  let limit        = $state(50);
  let autoRefresh  = $state(true);
  let interval: ReturnType<typeof setInterval>;

  async function load() {
    try {
      const params = new URLSearchParams({ limit: String(limit) });
      if (filterStatus) params.set('status', filterStatus);
      const res  = await fetch(`/api/jobs?${params}`);
      const data = await res.json();
      if (!data.error) jobs = data.items ?? [];
    } catch { /* silent */ }
  }

  function statusVariant(s: string): 'default' | 'destructive' | 'outline' | 'secondary' {
    if (s === 'success') return 'default';
    if (s === 'failed')  return 'destructive';
    if (s === 'running') return 'outline';
    return 'secondary';
  }

  function statusLabel(s: string) {
    return { queued: 'Antre', running: 'Berjalan', success: 'Sukses', failed: 'Gagal', cancelled: 'Batal' }[s] ?? s;
  }

  function runTypeLabel(t: string) {
    return { morning: 'Pagi', afternoon: 'Sore', checkin: 'Masuk', checkout: 'Pulang' }[t] ?? t;
  }

  function fmtDt(iso: string) {
    if (!iso) return '—';
    return new Date(iso).toLocaleString('id-ID', {
      timeZone: 'Asia/Makassar', day: 'numeric', month: 'short',
      hour: '2-digit', minute: '2-digit',
    });
  }

  $effect(() => {
    clearInterval(interval);
    if (autoRefresh) interval = setInterval(load, 5000);
    return () => clearInterval(interval);
  });

  onMount(load);
</script>

<svelte:head><title>Riwayat Job — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6">

  <div class="flex flex-wrap items-start justify-between gap-4">
    <div>
      <h1 class="text-2xl font-semibold text-slate-800">Riwayat Job</h1>
      <p class="text-sm text-muted-foreground mt-1">Log eksekusi job absensi Pusaka Kemenag</p>
    </div>
    <Button variant="outline" size="sm" href="/pusaka">← PUSAKA</Button>
  </div>

  <!-- Filter bar -->
  <Card.Root>
    <Card.Content class="pt-4 pb-3">
      <div class="flex flex-wrap gap-3 items-center">
        <span class="text-sm text-muted-foreground">Filter:</span>
        <select
          bind:value={filterStatus}
          onchange={load}
          class="h-9 rounded-md border border-input bg-background px-3 text-sm"
        >
          <option value="">Semua Status</option>
          <option value="queued">Antre</option>
          <option value="running">Berjalan</option>
          <option value="success">Sukses</option>
          <option value="failed">Gagal</option>
        </select>
        <select
          bind:value={limit}
          onchange={load}
          class="h-9 rounded-md border border-input bg-background px-3 text-sm"
        >
          <option value={20}>20 baris</option>
          <option value={50}>50 baris</option>
          <option value={100}>100 baris</option>
          <option value={200}>200 baris</option>
        </select>
        <label class="flex items-center gap-2 text-sm cursor-pointer">
          <input type="checkbox" bind:checked={autoRefresh} class="h-4 w-4 accent-green-700" />
          Auto-refresh
        </label>
        <Button variant="outline" size="sm" onclick={load}>↺ Refresh</Button>
        <span class="text-sm text-muted-foreground">{jobs.length} job</span>
      </div>
    </Card.Content>
  </Card.Root>

  <Card.Root>
    <Card.Content class="p-0 overflow-x-auto">
      <Table.Root>
        <Table.Header>
          <Table.Row>
            <Table.Head>Waktu</Table.Head>
            <Table.Head>Nama</Table.Head>
            <Table.Head>Tipe</Table.Head>
            <Table.Head>Status</Table.Head>
            <Table.Head class="text-center">Percobaan</Table.Head>
            <Table.Head class="hidden md:table-cell">Worker</Table.Head>
            <Table.Head class="hidden lg:table-cell">Retry At</Table.Head>
            <Table.Head class="hidden lg:table-cell">Error</Table.Head>
          </Table.Row>
        </Table.Header>
        <Table.Body>
          {#each jobs as j}
            <Table.Row>
              <Table.Cell class="text-xs text-muted-foreground whitespace-nowrap">{fmtDt(j.created_at)}</Table.Cell>
              <Table.Cell class="font-medium">{j.nama || j.employee_nama || '—'}</Table.Cell>
              <Table.Cell><Badge variant="outline">{runTypeLabel(j.run_type)}</Badge></Table.Cell>
              <Table.Cell><Badge variant={statusVariant(j.status)}>{statusLabel(j.status)}</Badge></Table.Cell>
              <Table.Cell class="text-center text-sm">{j.attempts}/{j.max_attempts}</Table.Cell>
              <Table.Cell class="hidden md:table-cell text-xs text-muted-foreground">{j.claimed_by || '—'}</Table.Cell>
              <Table.Cell class="hidden lg:table-cell text-xs text-muted-foreground whitespace-nowrap">{fmtDt(j.next_retry_at)}</Table.Cell>
              <Table.Cell class="hidden lg:table-cell text-xs text-muted-foreground max-w-48 truncate">{j.error_message || '—'}</Table.Cell>
            </Table.Row>
          {:else}
            <Table.Row>
              <Table.Cell colspan={8} class="py-12 text-center text-muted-foreground">
                Tidak ada job ditemukan.
              </Table.Cell>
            </Table.Row>
          {/each}
        </Table.Body>
      </Table.Root>
    </Card.Content>
  </Card.Root>

</div>
