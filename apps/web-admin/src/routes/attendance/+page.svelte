<script lang="ts">
  import { onMount } from 'svelte';
  import * as Card from '$lib/components/ui/card';
  import * as Table from '$lib/components/ui/table';
  import * as Badge from '$lib/components/ui/badge';
  import { Button } from '$lib/components/ui/button';
  import { Input } from '$lib/components/ui/input';
  import { Loader2 } from 'lucide-svelte';

  interface AttendanceRecord {
    id: string;
    employee_id: string;
    employee_nama: string;
    employee_nip: string;
    tanggal: string;
    jam_masuk: string;
    jam_pulang: string;
    source_job_id: string;
    created_at: string;
    updated_at: string;
  }

  interface WorkerStatus {
    active_workers: Array<{
      worker_id: string;
      active_consumers: number;
      target_concurrency: number;
      headless: boolean;
      last_seen_at: string;
    }>;
    total: number;
    last_checked: string;
  }

  let records = $state<AttendanceRecord[]>([]);
  let filterDate = $state('');
  let limit = $state(100);
  let workerStatus = $state<WorkerStatus | null>(null);
  let loadingStatus = $state(false);
  let loading = $state(false);

  async function load() {
    loading = true;
    try {
      const params = new URLSearchParams({ limit: String(limit) });
      if (filterDate) params.set('date', filterDate);
      const res = await fetch(`/api/attendance?${params}`);
      const data = await res.json();
      if (data.error) {
        console.error('[mtsn2kolut] attendance:', data.error);
        return;
      }
      records = data.items ?? [];
    } catch (e) {
      console.error('[mtsn2kolut] attendance load failed:', e);
    } finally {
      loading = false;
    }
  }

  async function loadWorkerStatus() {
    loadingStatus = true;
    try {
      const res = await fetch('/api/worker/status');
      const data = await res.json();
      workerStatus = data.data ?? data;
    } catch (e) {
      console.error('Worker status load failed:', e);
    } finally {
      loadingStatus = false;
    }
  }

  async function syncAttendance() {
    try {
      const res = await fetch('/api/jobs/sync-attendance', { method: 'POST' });
      const data = await res.json();
      if (res.ok) {
        alert(`Sinkronisasi berhasil! ${data.inserted || 0} job dibuat.`);
        loadWorkerStatus();
      } else {
        alert('Sinkronisasi gagal: ' + (data.error || 'Unknown error'));
      }
    } catch (e) {
      alert('Sinkronisasi gagal');
    }
  }

  function todayWita() {
    return new Intl.DateTimeFormat('en-CA', {
      timeZone: 'Asia/Makassar',
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
    }).format(new Date());
  }

  function stripWita(val: string) {
    return val ? val.replace(/\s*WITA$/i, '') : '–';
  }

  function getAttendanceStatus(record: AttendanceRecord) {
    if (record.jam_masuk && record.jam_pulang) return 'lengkap';
    if (record.jam_masuk) return 'masuk';
    return 'belum';
  }

  function getStatusVariant(status: string) {
    if (status === 'lengkap') return 'default';
    if (status === 'masuk') return 'outline';
    return 'destructive';
  }

  function getStatusLabel(status: string) {
    if (status === 'lengkap') return 'Lengkap';
    if (status === 'masuk') return 'Masuk';
    return 'Belum';
  }

  onMount(() => {
    filterDate = todayWita();
    load();
    loadWorkerStatus();
    const statusInterval = setInterval(loadWorkerStatus, 30000);
    return () => clearInterval(statusInterval);
  });
</script>

<svelte:head>
  <title>Absensi — MTSN 2 Kolut Super App</title>
</svelte:head>

<div class="space-y-6">
  <!-- Worker Status Panel -->
  <Card.Root>
    <Card.Header>
      <Card.Title>Status Worker Pusaka</Card.Title>
      <Card.Description>Monitoring status worker yang menangani absensi Pusaka Kemenag</Card.Description>
    </Card.Header>
    <Card.Content>
      {#if loadingStatus}
        <div class="flex items-center gap-2">
          <Loader2 class="h-4 w-4 animate-spin" />
          <span class="text-sm text-muted-foreground">Memuat status...</span>
        </div>
      {:else if workerStatus}
        <div class="grid grid-cols-3 gap-4">
          <div>
            <p class="text-sm text-muted-foreground">Active Workers</p>
            <p class="text-2xl font-bold">{workerStatus.active_workers?.length ?? 0}</p>
          </div>
          <div>
            <p class="text-sm text-muted-foreground">Total Consumers</p>
            <p class="text-2xl font-bold">
              {workerStatus.active_workers?.reduce((acc, w) => acc + w.active_consumers, 0) ?? 0}
            </p>
          </div>
          <div class="flex items-end">
            <Badge variant={workerStatus.total > 0 ? 'default' : 'destructive'}>
              {workerStatus.total > 0 ? 'Online' : 'Offline'}
            </Badge>
          </div>
        </div>
        {#if workerStatus.active_workers && workerStatus.active_workers.length > 0}
          <div class="mt-4 space-y-2">
            {#each workerStatus.active_workers as worker}
              <div class="text-sm text-muted-foreground flex items-center gap-2">
                <span class="font-mono">{worker.worker_id}</span>
                <Badge variant="outline">Consumers: {worker.active_consumers}</Badge>
                {#if worker.headless}
                  <Badge variant="secondary">Headless</Badge>
                {/if}
              </div>
            {/each}
          </div>
        {/if}
      {/if}
    </Card.Content>
  </Card.Root>

  <!-- Attendance Records -->
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <h2 class="text-2xl font-bold">Absensi</h2>
      <div class="flex items-center gap-2">
        <Button variant="default" size="sm" onclick={syncAttendance}>
          ↺ Sync Sekarang
        </Button>
        <Input
          type="date"
          value={filterDate}
          onchange={(e) => {
            filterDate = (e.target as HTMLInputElement).value;
            load();
          }}
          class="w-auto"
        />
        <Button variant="outline" size="sm" onclick={load} disabled={loading}>
          {#if loading}<Loader2 class="h-4 w-4 animate-spin" />{/if}
          ↺ Refresh
        </Button>
      </div>
    </div>

    <div class="rounded-md border">
      <Table.Root>
        <Table.Header>
          <Table.Row>
            <Table.Head class="w-[50px]">#</Table.Head>
            <Table.Head>Nama</Table.Head>
            <Table.Head class="text-center">NIP</Table.Head>
            <Table.Head class="text-center">Masuk</Table.Head>
            <Table.Head class="text-center">Pulang</Table.Head>
            <Table.Head class="text-center">Status</Table.Head>
          </Table.Row>
        </Table.Header>
        <Table.Body>
          {#each records as record, i}
            <Table.Row>
              <Table.Cell class="text-muted-foreground">{i + 1}</Table.Cell>
              <Table.Cell class="font-medium">{record.employee_nama}</Table.Cell>
              <Table.Cell class="text-center text-muted-foreground">{record.employee_nip}</Table.Cell>
              <Table.Cell class="text-center">{stripWita(record.jam_masuk)}</Table.Cell>
              <Table.Cell class="text-center">{stripWita(record.jam_pulang)}</Table.Cell>
              <Table.Cell class="text-center">
                <Badge variant={getStatusVariant(getAttendanceStatus(record))}>
                  {getStatusLabel(getAttendanceStatus(record))}
                </Badge>
              </Table.Cell>
            </Table.Row>
          {/each}
          {#if records.length === 0}
            <Table.Row>
              <Table.Cell colspan={6} class="text-center py-6 text-muted-foreground">
                Tidak ada data absensi untuk tanggal ini.
              </Table.Cell>
            </Table.Row>
          {/if}
        </Table.Body>
      </Table.Root>
    </div>
  </div>
</div>
