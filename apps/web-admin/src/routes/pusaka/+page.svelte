<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { toast } from '$lib/components/ui/sonner';
	import QueueMonitor from '$lib/components/QueueMonitor.svelte';

	interface QueueStats {
		queued: number; running: number; success: number;
		failed: number; retry_due: number; total: number;
	}
	interface WorkerEntry {
		worker_id: string; active_consumers: number;
		target_concurrency: number; headless: boolean; reported_at: string;
	}
	interface WorkerStatus {
		active_workers: WorkerEntry[];
		total: number;
		queue: { queued: number; running: number; success: number; failed: number };
	}

	let queueStats   = $state<QueueStats>({ queued: 0, running: 0, success: 0, failed: 0, retry_due: 0, total: 0 });
	let workerStatus = $state<WorkerStatus | null>(null);
	let recentJobs   = $state<any[]>([]);
	let busy         = $state<Record<string, boolean>>({});
	let confirmKey   = $state('');

	async function load() {
		try {
			const [qRes, jRes, wRes] = await Promise.all([
				fetch('/api/queue/stats'),
				fetch('/api/pusaka/jobs?limit=5'),
				fetch('/api/pusaka/worker/status'),
			]);
			const q = await qRes.json().catch(() => ({}));
			const j = await jRes.json().catch(() => ({}));
			const w = await wRes.json().catch(() => ({}));
			if (!q.error) queueStats = q;
			recentJobs = j.items ?? j.data ?? [];
			workerStatus = w.data ?? null;
		} catch { /* silent */ }
	}

	async function act(key: string, fn: () => Promise<Response>, successMsg: string) {
		busy = { ...busy, [key]: true };
		try {
			const res  = await fn();
			const data = await res.json().catch(() => ({}));
			if (!res.ok) throw new Error((data as any).error || 'Gagal');
			showToast(successMsg + ((data as any).cancelled != null ? ` (${(data as any).cancelled} job)` : ''), 'ok');
		} catch (e: any) {
			showToast(e.message, 'err');
		} finally {
			busy = { ...busy, [key]: false };
			confirmKey = '';
			await load();
		}
	}

	async function runRekap() {
		busy = { ...busy, rekap: true };
		try {
			const res  = await fetch('/api/pusaka/jobs/run-all', { method: 'POST', headers: { 'content-type': 'application/json' }, body: JSON.stringify({ run_type: 'morning' }) });
			const data = await res.json().catch(() => ({}));
			if (!res.ok) throw new Error((data as any).error || 'Gagal');
			showToast(`Rekap di-queue: ${(data as any).inserted ?? 0} job baru`, 'ok');
		} catch (e: any) {
			showToast(e.message, 'err');
		} finally {
			busy = { ...busy, rekap: false };
			await load();
		}
	}

	const triggerSched  = ()          => act('sched',      () => fetch('/api/pusaka/scheduler/tick', { method: 'POST' }), 'Scheduler tick dijalankan');
	const cancelAll     = ()          => act('cancel_all', () => fetch('/api/pusaka/jobs/cancel-all',{ method: 'POST' }), 'Semua antrian dibatalkan');

	function showToast(msg: string, type: 'ok' | 'err' = 'ok') {
		if (type === 'ok') toast.success(msg);
		else toast.error(msg);
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
		return { morning: 'Rekap', afternoon: 'Rekap', checkin: 'Masuk', checkout: 'Pulang' }[t] ?? t;
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
		const itv = setInterval(load, 10_000);
		return () => clearInterval(itv);
	});
</script>

<svelte:head><title>PUSAKA Kemenag — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6">

	<!-- Header -->
	<div class="flex flex-wrap items-start justify-between gap-4">
		<div>
			<h1 class="text-2xl font-semibold text-slate-800">Kontrol PUSAKA Kemenag</h1>
			<p class="text-sm text-muted-foreground mt-1">Monitor dan kontrol sinkronisasi data kehadiran dari PUSAKA Kemenag</p>
		</div>
		<div class="flex flex-wrap gap-2">
			<Button variant="outline" size="sm" onclick={triggerSched} disabled={busy.sched}>
				{busy.sched ? '...' : '⚡ Trigger Scheduler'}
			</Button>
			<Button size="sm" onclick={runRekap} disabled={busy.rekap}>
				{busy.rekap ? '...' : '▶ Jalankan Rekap'}
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
		</div>
	</div>

	<!-- Worker status -->
	<div class="grid gap-4 sm:grid-cols-3">
		<Card.Root>
			<Card.Header class="pb-2">
				<Card.Description>Worker Aktif</Card.Description>
			</Card.Header>
			<Card.Content class="pt-0">
				<div class="flex items-center gap-2">
					<span class="text-2xl font-bold">{workerStatus?.total ?? '—'}</span>
					{#if workerStatus}
						<Badge variant={workerStatus.total > 0 ? 'default' : 'destructive'}>
							{workerStatus.total > 0 ? 'Online' : 'Offline'}
						</Badge>
					{/if}
				</div>
			</Card.Content>
		</Card.Root>

		<Card.Root>
			<Card.Header class="pb-2">
				<Card.Description>Consumers</Card.Description>
			</Card.Header>
			<Card.Content class="pt-0">
				<span class="text-2xl font-bold">
					{workerStatus?.active_workers?.reduce((s, w) => s + w.active_consumers, 0) ?? '—'}
				</span>
				{#if workerStatus}
					<span class="text-sm text-muted-foreground ml-1">
						/ {workerStatus.active_workers?.reduce((s, w) => s + w.target_concurrency, 0)} kapasitas
					</span>
				{/if}
			</Card.Content>
		</Card.Root>

		<Card.Root>
			<Card.Header class="pb-2">
				<Card.Description>Antrian Job</Card.Description>
			</Card.Header>
			<Card.Content class="pt-0">
				{#if workerStatus?.queue}
					{@const q = workerStatus.queue}
					<div class="flex flex-wrap gap-2">
						{#if q.running > 0}
							<Badge variant="default">{q.running} berjalan</Badge>
						{/if}
						{#if q.queued > 0}
							<Badge variant="outline">{q.queued} antre</Badge>
						{/if}
						{#if q.running === 0 && q.queued === 0}
							<Badge variant="secondary">Idle</Badge>
						{/if}
					</div>
					<p class="text-xs text-muted-foreground mt-1">{q.success} sukses · {q.failed} gagal</p>
				{:else}
					<span class="text-2xl font-bold">—</span>
				{/if}
			</Card.Content>
		</Card.Root>
	</div>

	<!-- Queue monitor -->
	<QueueMonitor stats={queueStats} />

	<!-- Recent jobs -->
	<Card.Root class="overflow-hidden border-slate-200 shadow-sm">
		<Card.Header class="pb-3">
			<div class="flex items-center justify-between">
				<Card.Title class="text-base">Job Terbaru</Card.Title>
			<Button variant="ghost" size="sm" href={resolve('/pusaka/antrian')}>Lihat semua →</Button>
			</div>
		</Card.Header>
		<Card.Content class="p-0">
			<div class="hidden overflow-x-auto lg:block">
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
					{#each recentJobs as j (j.id)}
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
			</div>

			<div class="grid gap-3 p-4 lg:hidden">
				{#each recentJobs as j (j.id)}
					<div class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
						<div class="flex items-start justify-between gap-3">
							<div class="min-w-0">
								<p class="text-xs font-semibold uppercase tracking-[0.16em] text-slate-400">{fmtDt(j.created_at)}</p>
								<p class="mt-1 text-sm font-semibold text-slate-900">{j.nama || j.employee_nama || '—'}</p>
								<p class="mt-1 text-xs text-slate-500">{j.claimed_by || 'Belum diklaim worker'}</p>
							</div>
							<Badge variant={statusVariant(j.status)}>{statusLabel(j.status)}</Badge>
						</div>
						<div class="mt-3 flex items-center gap-2">
							<Badge variant="outline">{runTypeLabel(j.run_type)}</Badge>
						</div>
					</div>
				{:else}
					<div class="rounded-2xl border border-dashed border-slate-300 bg-slate-50 px-4 py-10 text-center text-sm text-slate-500">
						Belum ada job.
					</div>
				{/each}
			</div>
		</Card.Content>
	</Card.Root>

	<!-- Quick navigation -->
	<div class="grid gap-3 sm:grid-cols-2">
		<Button variant="outline" href={resolve('/pusaka/employees')} class="h-auto py-4 flex-col items-start text-left gap-1">
			<span class="font-semibold">Pegawai PUSAKA</span>
			<span class="text-xs text-muted-foreground font-normal">Setup akun, jadwal, dan job untuk pegawai eligible</span>
		</Button>
		<Button variant="outline" href={resolve('/pusaka/kehadiran')} class="h-auto py-4 flex-col items-start text-left gap-1">
			<span class="font-semibold">Data Kehadiran</span>
			<span class="text-xs text-muted-foreground font-normal">Rekap harian dari PUSAKA Kemenag</span>
		</Button>
		<Button variant="outline" href={resolve('/pusaka/summary')} class="h-auto py-4 flex-col items-start text-left gap-1">
			<span class="font-semibold">Ringkasan Kehadiran</span>
			<span class="text-xs text-muted-foreground font-normal">Akumulasi per periode / bulan</span>
		</Button>
	</div>

</div>
