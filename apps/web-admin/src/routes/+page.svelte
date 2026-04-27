<script lang="ts">
  import { onMount } from 'svelte';
  import * as Card from '$lib/components/ui/card';
  import * as Table from '$lib/components/ui/table';
  import { Badge } from '$lib/components/ui/badge';
  import { Button } from '$lib/components/ui/button';
  import QueueMonitor from '$lib/components/QueueMonitor.svelte';

  interface QueueStats {
    queued: number; running: number; success: number;
    failed: number; retry_due: number; total: number;
  }

  let queueStats = $state<QueueStats>({ queued: 0, running: 0, success: 0, failed: 0, retry_due: 0, total: 0 });
  let recentJobs = $state<any[]>([]);
  let busy       = $state<Record<string, boolean>>({});
  let toast      = $state({ msg: '', type: 'ok' as 'ok' | 'err' });
  let confirmKey = $state('');

  async function load() {
    try {
      const [qRes, jRes] = await Promise.all([
        fetch('/api/queue/stats'),
        fetch('/api/jobs?limit=5'),
      ]);
      const q = await qRes.json();
      const j = await jRes.json();
      if (!q.error) queueStats = q;
      if (!j.error) recentJobs = j.items ?? [];
    } catch { /* silent */ }
  }

  async function act(key: string, fn: () => Promise<Response>, successMsg: string) {
    busy = { ...busy, [key]: true };
    try {
      const res  = await fn();
      const data = await res.json().catch(() => ({}));
      if (!res.ok) throw new Error(data.error || 'Gagal');
      showToast(successMsg + (data.cancelled != null ? ` (${data.cancelled} job)` : ''), 'ok');
    } catch (e: any) {
      showToast(e.message, 'err');
    } finally {
      busy = { ...busy, [key]: false };
      confirmKey = '';
      await load();
    }
  }

  const runAll       = (t: string) => act(`run_${t}`,  () => fetch('/api/jobs/run-all',    { method: 'POST', headers: { 'content-type': 'application/json' }, body: JSON.stringify({ run_type: t }) }), `Job ${t === 'morning' ? 'pagi' : 'sore'} di-queue`);
  const triggerSched = ()          => act('sched',      () => fetch('/api/scheduler/tick',  { method: 'POST' }), 'Scheduler tick dijalankan');
  const cancelAll    = ()          => act('cancel_all', () => fetch('/api/jobs/cancel-all', { method: 'POST' }), 'Semua antrian dibatalkan');
  const restartWorker= ()          => act('restart',    () => fetch('/api/worker/restart',  { method: 'POST' }), 'Worker di-restart');

  function showToast(msg: string, type: 'ok' | 'err' = 'ok') {
    toast = { msg, type };
    setTimeout(() => (toast = { msg: '', type: 'ok' }), 3500);
  }

  function statusVariant(s: string): 'default' | 'destructive' | 'outline' | 'secondary' {
    if (s === 'success') return 'default';
    if (s === 'failed')  return 'destructive';
    if (s === 'running') return 'outline';
    return 'secondary';
  }

  function statusLabel(s: string) {
    return { queued: 'Antre', running: 'Berjalan', success: 'Sukses', failed: 'Gagal' }[s] ?? s;
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

  onMount(() => {
    load();
    const itv = setInterval(load, 5000);
    return () => clearInterval(itv);
  });
</script>

<svelte:head><title>Dashboard — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6">

  <!-- Header -->
  <div class="flex flex-wrap items-start justify-between gap-4">
    <div>
      <h1 class="text-2xl font-semibold text-slate-800">Dashboard Operasional</h1>
      <p class="text-sm text-muted-foreground mt-1">Kontrol absensi dan worker integrasi Pusaka Kemenag</p>
    </div>
    <div class="flex flex-wrap gap-2">
      <Button variant="outline" size="sm" onclick={triggerSched} disabled={busy.sched}>
        {busy.sched ? '...' : '⚡ Trigger Scheduler'}
      </Button>
      <Button size="sm" onclick={() => runAll('morning')} disabled={busy.run_morning}>
        {busy.run_morning ? '...' : '▶ Run All Pagi'}
      </Button>
      <Button variant="outline" size="sm" onclick={() => runAll('afternoon')} disabled={busy.run_afternoon}>
        {busy.run_afternoon ? '...' : '▶ Run All Sore'}
      </Button>

      {#if confirmKey === 'cancel_all'}
        <span class="self-center text-xs text-amber-700">Batalkan semua antrian?</span>
        <Button size="sm" variant="destructive" onclick={cancelAll}>Ya</Button>
        <Button size="sm" variant="ghost" onclick={() => (confirmKey = '')}>Tidak</Button>
      {:else}
        <Button size="sm" variant="outline" onclick={() => (confirmKey = 'cancel_all')} disabled={busy.cancel_all}
          class="text-destructive border-destructive/40 hover:bg-destructive/10">
          ✕ Cancel All
        </Button>
      {/if}

      {#if confirmKey === 'restart'}
        <span class="self-center text-xs text-amber-700">Restart worker sekarang?</span>
        <Button size="sm" variant="destructive" onclick={restartWorker}>Ya</Button>
        <Button size="sm" variant="ghost" onclick={() => (confirmKey = '')}>Tidak</Button>
      {:else}
        <Button size="sm" variant="outline" onclick={() => (confirmKey = 'restart')} disabled={busy.restart}
          class="text-destructive border-destructive/40 hover:bg-destructive/10">
          ↺ Restart Worker
        </Button>
      {/if}
    </div>
  </div>

  <!-- Toast -->
  {#if toast.msg}
    <div class="rounded-md px-4 py-3 text-sm
      {toast.type === 'ok'
        ? 'border border-green-200 bg-green-50 text-green-800'
        : 'border border-red-200 bg-red-50 text-destructive'}">
      {toast.msg}
    </div>
  {/if}

  <!-- Queue stats -->
  <QueueMonitor stats={queueStats} />

  <!-- Recent jobs -->
  <Card.Root>
    <Card.Header class="pb-3">
      <div class="flex items-center justify-between">
        <Card.Title class="text-base">Job Terbaru</Card.Title>
        <Button variant="ghost" size="sm" href="/jobs">Lihat semua →</Button>
      </div>
    </Card.Header>
    <Card.Content class="p-0 overflow-x-auto">
      <Table.Root>
        <Table.Header>
          <Table.Row>
            <Table.Head>Waktu</Table.Head>
            <Table.Head>Nama</Table.Head>
            <Table.Head>Tipe</Table.Head>
            <Table.Head>Status</Table.Head>
            <Table.Head class="hidden sm:table-cell">Worker</Table.Head>
          </Table.Row>
        </Table.Header>
        <Table.Body>
          {#each recentJobs as j}
            <Table.Row>
              <Table.Cell class="text-xs text-muted-foreground whitespace-nowrap">{fmtDt(j.created_at)}</Table.Cell>
              <Table.Cell class="font-medium">{j.nama || j.employee_nama || '—'}</Table.Cell>
              <Table.Cell><Badge variant="outline">{runTypeLabel(j.run_type)}</Badge></Table.Cell>
              <Table.Cell><Badge variant={statusVariant(j.status)}>{statusLabel(j.status)}</Badge></Table.Cell>
              <Table.Cell class="hidden sm:table-cell text-xs text-muted-foreground">{j.claimed_by || '—'}</Table.Cell>
            </Table.Row>
          {:else}
            <Table.Row>
              <Table.Cell colspan={5} class="py-10 text-center text-muted-foreground">Belum ada job.</Table.Cell>
            </Table.Row>
          {/each}
        </Table.Body>
      </Table.Root>
    </Card.Content>
  </Card.Root>

</div>
