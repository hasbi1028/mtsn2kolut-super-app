<script lang="ts">
  import { onMount } from 'svelte';
  import * as Card from '$lib/components/ui/card';
  import * as Table from '$lib/components/ui/table';
  import * as Badge from '$lib/components/ui/badge';
  import { Button } from '$lib/components/ui/button';
  import { Select } from '$lib/components/ui/select';
  import { Loader2 } from 'lucide-svelte';

  interface Job {
    id: string;
    employee_id: string;
    employee_nama: string;
    run_type: string;
    status: string;
    attempts: number;
    max_attempts: number;
    created_at: string;
    updated_at: string;
  }

  let jobs = $state<Job[]>([]);
  let loading = $state(true);
  let filterStatus = $state('');
  let filterType = $state('');

  async function loadJobs() {
    loading = true;
    try {
      const params = new URLSearchParams({ limit: '100' });
      if (filterStatus) params.set('status', filterStatus);
      if (filterType) params.set('run_type', filterType);
      
      const res = await fetch(`/api/jobs?${params}`);
      const data = await res.json();
      jobs = data.items ?? data ?? [];
    } catch (e) {
      console.error('Jobs load failed:', e);
    } finally {
      loading = false;
    }
  }

  function statusVariant(status: string) {
    if (status === 'completed') return 'default';
    if (status === 'failed') return 'destructive';
    if (status === 'running') return 'outline';
    return 'secondary';
  }

  function statusLabel(status: string) {
    const labels: Record<string, string> = {
      'queued': 'Antri',
      'running': 'Berjalan',
      'completed': 'Selesai',
      'failed': 'Gagal',
      'cancelled': 'Dibatalkan',
    };
    return labels[status] || status;
  }

  function runTypeLabel(type: string) {
    const labels: Record<string, string> = {
      'morning': 'Pagi',
      'afternoon': 'Sore',
      'checkin': 'Masuk',
      'checkout': 'Pulang',
    };
    return labels[type] || type;
  }

  function formatDate(iso: string) {
    if (!iso) return '–';
    return new Date(iso).toLocaleString('id-ID', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  }

  onMount(() => {
    loadJobs();
    const interval = setInterval(loadJobs, 10000);
    return () => clearInterval(interval);
  });
</script>

<svelte:head>
  <title>Antrian Job — MTSN 2 Kolut Super App</title>
</svelte:head>

<div class="space-y-6">
  <div class="flex items-center justify-between">
    <h2 class="text-2xl font-bold">Antrian Job Absensi</h2>
    <Button variant="outline" size="sm" onclick={loadJobs} disabled={loading}>
      {#if loading}<Loader2 class="h-4 w-4 animate-spin" />{/if}
      ↺ Refresh
    </Button>
  </div>

  <div class="flex items-center gap-4">
    <div class="flex items-center gap-2">
      <label for="filter-status" class="text-sm">Status:</label>
      <select id="filter-status" bind:value={filterStatus} onchange={loadJobs} class="text-sm border rounded px-2 py-1">
        <option value="">Semua</option>
        <option value="queued">Antri</option>
        <option value="running">Berjalan</option>
        <option value="completed">Selesai</option>
        <option value="failed">Gagal</option>
        <option value="cancelled">Dibatalkan</option>
      </select>
    </div>
    <div class="flex items-center gap-2">
      <label for="filter-type" class="text-sm">Tipe:</label>
      <select id="filter-type" bind:value={filterType} onchange={loadJobs} class="text-sm border rounded px-2 py-1">
        <option value="">Semua</option>
        <option value="morning">Pagi</option>
        <option value="afternoon">Sore</option>
        <option value="checkin">Masuk</option>
        <option value="checkout">Pulang</option>
      </select>
    </div>
  </div>

  <Card.Root>
    <Card.Content class="p-0">
      {#if loading && jobs.length === 0}
        <div class="flex items-center justify-center py-8">
          <Loader2 class="h-6 w-6 animate-spin" />
          <span class="ml-2">Memuat job...</span>
        </div>
      {:else if jobs.length === 0}
        <div class="text-center py-8 text-muted-foreground">
          Tidak ada job ditemukan.
        </div>
      {:else}
        <Table.Root>
          <Table.Header>
            <Table.Row>
              <Table.Head class="w-[80px]">ID</Table.Head>
              <Table.Head>Pegawai</Table.Head>
              <Table.Head>Tipe</Table.Head>
              <Table.Head>Status</Table.Head>
              <Table.Head class="text-center">Percobaan</Table.Head>
              <Table.Head>Terakhir Update</Table.Head>
            </Table.Row>
          </Table.Header>
          <Table.Body>
            {#each jobs as job}
              <Table.Row>
                <Table.Cell class="font-mono text-xs">{job.id.slice(0, 8)}...</Table.Cell>
                <Table.Cell class="font-medium">{job.employee_nama || job.employee_id}</Table.Cell>
                <Table.Cell>
                  <Badge variant="outline">{runTypeLabel(job.run_type)}</Badge>
                </Table.Cell>
                <Table.Cell>
                  <Badge variant={statusVariant(job.status)}>{statusLabel(job.status)}</Badge>
                </Table.Cell>
                <Table.Cell class="text-center">{job.attempts}/{job.max_attempts}</Table.Cell>
                <Table.Cell class="text-sm text-muted-foreground">
                  {formatDate(job.updated_at)}
                </Table.Cell>
              </Table.Row>
            {/each}
          </Table.Body>
        </Table.Root>
      {/if}
    </Card.Content>
  </Card.Root>
</div>
