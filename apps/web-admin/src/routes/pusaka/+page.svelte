<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { toast } from '$lib/components/ui/sonner';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import QueueMonitor from '$lib/components/QueueMonitor.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import EmptyStatePanel from '$lib/components/EmptyStatePanel.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import OperationStatusPanel from '$lib/components/OperationStatusPanel.svelte';
	import { confirmChallenge } from '$lib/confirm-dialog';

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

	interface RecentJob {
		id: string;
		created_at: string;
		nama?: string;
		employee_nama?: string;
		run_type: string;
		status: string;
		claimed_by?: string;
	}

	interface PusakaOverview {
		queueStats: QueueStats;
		workerStatus: WorkerStatus | null;
		recentJobs: RecentJob[];
	}

	type ApiEnvelope<T> = {
		data?: T;
		error?: string;
		message?: string;
	};

	type JobListPayload = {
		items?: RecentJob[];
		data?: RecentJob[];
		error?: string;
		message?: string;
	};

	type PusakaActionResponse = {
		cancelled?: number;
		inserted?: number;
		error?: string;
		message?: string;
	};

	let overviewPromise = $state<Promise<PusakaOverview> | null>(null);
	let queueStats   = $state<QueueStats>(emptyQueueStats());
	let workerStatus = $state<WorkerStatus | null>(null);
	let recentJobs   = $state<RecentJob[]>([]);
	let busy         = $state<Record<string, boolean>>({});
	let confirmKey   = $state('');
	let operationState = $state<{ tone: 'success' | 'error' | 'warning' | 'info'; title: string; message: string } | null>(null);

	function emptyQueueStats(): QueueStats {
		return { queued: 0, running: 0, success: 0, failed: 0, retry_due: 0, total: 0 };
	}

	function isRecord(value: unknown): value is Record<string, unknown> {
		return typeof value === 'object' && value !== null;
	}

	function apiErrorMessage(payload: unknown) {
		if (!isRecord(payload)) return '';
		const error = payload.error;
		if (typeof error === 'string' && error.trim()) return error;
		const message = payload.message;
		if (typeof message === 'string' && message.trim()) return message;
		return '';
	}

	async function readApi<T>(response: Response, fallbackMessage: string): Promise<T> {
		const payload = (await response.json().catch(() => null)) as ApiEnvelope<T> | T | null;
		const message = apiErrorMessage(payload);
		if (!response.ok) throw new Error(message || fallbackMessage);
		if (isRecord(payload) && typeof payload.error === 'string' && payload.error.trim()) {
			throw new Error(payload.error);
		}
		if (isRecord(payload) && 'data' in payload) {
			const envelope = payload as ApiEnvelope<T>;
			if (envelope.data === undefined) throw new Error(fallbackMessage);
			return envelope.data;
		}
		if (payload === null) throw new Error(fallbackMessage);
		return payload as T;
	}

	function normalizeQueueStats(value: Partial<QueueStats> | null | undefined): QueueStats {
		return { ...emptyQueueStats(), ...(value ?? {}) };
	}

	function normalizeJobs(value: JobListPayload | RecentJob[] | null | undefined): RecentJob[] {
		if (Array.isArray(value)) return value;
		if (!value) return [];
		if (Array.isArray(value.items)) return value.items;
		if (Array.isArray(value.data)) return value.data;
		return [];
	}

	function applyOverview(overview: PusakaOverview) {
		queueStats = overview.queueStats;
		workerStatus = overview.workerStatus;
		recentJobs = overview.recentJobs;
	}

	function currentOverview(): PusakaOverview {
		return { queueStats, workerStatus, recentJobs };
	}

	async function fetchOverview(): Promise<PusakaOverview> {
		const [queueData, jobsData, workerData] = await Promise.all([
			fetch('/api/queue/stats').then((response) => readApi<Partial<QueueStats>>(response, 'Gagal memuat ringkasan antrian PUSAKA')),
			fetch('/api/pusaka/jobs?limit=5').then((response) => readApi<JobListPayload | RecentJob[]>(response, 'Gagal memuat job terbaru PUSAKA')),
			fetch('/api/pusaka/worker/status').then((response) => readApi<WorkerStatus | null>(response, 'Gagal memuat status worker PUSAKA')),
		]);

		return {
			queueStats: normalizeQueueStats(queueData),
			recentJobs: normalizeJobs(jobsData),
			workerStatus: workerData ?? null,
		};
	}

	function loadOverview() {
		overviewPromise = fetchOverview().then((overview) => {
			applyOverview(overview);
			return overview;
		});
		return overviewPromise;
	}

	async function refreshOverview(showFailureToast = false) {
		if (!overviewPromise) {
			await loadOverview();
			return;
		}
		try {
			const overview = await fetchOverview();
			applyOverview(overview);
			overviewPromise = Promise.resolve(overview);
		} catch (error) {
			overviewPromise = Promise.resolve(currentOverview());
			if (showFailureToast) toast.error(overviewErrorMessage(error));
		}
	}

	function retryOverview(reset?: () => void) {
		reset?.();
		loadOverview();
	}

	function overviewErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		if (typeof error === 'string' && error.trim()) return error;
		return 'Gagal memuat status worker, ringkasan antrian, atau job terbaru. Periksa backend dan worker PUSAKA, lalu coba lagi.';
	}

	function handleOverviewRenderError(error: unknown, reset: () => void) {
		console.error('PUSAKA overview render failed', error);
		reset();
	}

	async function act(key: string, fn: () => Promise<Response>, successMsg: string) {
		busy = { ...busy, [key]: true };
		try {
			const res  = await fn();
			const data = (await res.json().catch(() => ({}))) as PusakaActionResponse;
			if (!res.ok) throw new Error(apiErrorMessage(data) || 'Gagal');
			operationState = {
				tone: key === 'cancel_all' ? 'warning' : 'success',
				title: key === 'cancel_all' ? 'Antrian Dibatalkan' : 'Operasi PUSAKA Berhasil',
				message: successMsg + (data.cancelled != null ? ` (${data.cancelled} job)` : ''),
			};
			showToast(successMsg + (data.cancelled != null ? ` (${data.cancelled} job)` : ''), 'ok');
		} catch (error) {
			operationState = {
				tone: 'error',
				title: 'Operasi PUSAKA Gagal',
				message: overviewErrorMessage(error),
			};
			showToast(overviewErrorMessage(error), 'err');
		} finally {
			busy = { ...busy, [key]: false };
			confirmKey = '';
			await refreshOverview(true);
		}
	}

	async function runRekap() {
		if (!(await confirmChallenge({
			title: 'Mulai Rekap Massal PUSAKA',
			message: 'Rekap massal akan membuat job untuk seluruh akun PUSAKA yang aktif. Gunakan hanya saat operator siap memantau antrian.',
			challenge: 'REKAP',
			confirmLabel: 'Mulai Rekap',
			tone: 'warning'
		}))) return;
		busy = { ...busy, rekap: true };
		try {
			const res  = await fetch('/api/pusaka/jobs/run-all', { method: 'POST', headers: { 'content-type': 'application/json' }, body: JSON.stringify({ run_type: 'morning' }) });
			const data = (await res.json().catch(() => ({}))) as PusakaActionResponse;
			if (!res.ok) throw new Error(apiErrorMessage(data) || 'Gagal');
			operationState = {
				tone: 'warning',
				title: 'Rekap Massal Diantrekan',
				message: `Sistem menambahkan ${data.inserted ?? 0} job baru. Pantau hasilnya di antrian dan worker status sebelum mengulangi operasi ini.`,
			};
			showToast(`Rekap di-queue: ${data.inserted ?? 0} job baru`, 'ok');
		} catch (error) {
			operationState = {
				tone: 'error',
				title: 'Rekap Massal Gagal',
				message: overviewErrorMessage(error),
			};
			showToast(overviewErrorMessage(error), 'err');
		} finally {
			busy = { ...busy, rekap: false };
			await refreshOverview(true);
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
		void loadOverview();
		const itv = setInterval(() => void refreshOverview(false), 10_000);
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
			<LoadingButton variant="outline" size="sm" onclick={() => void triggerSched()} loading={busy.sched} loadingLabel="Memproses..." label="⚡ Jalankan Scheduler" />
			<LoadingButton size="sm" onclick={() => void runRekap()} loading={busy.rekap} loadingLabel="Memproses..." label="▶ Mulai Rekap" />
			{#if confirmKey === 'cancel_all'}
				<span class="self-center text-xs text-amber-700">Batalkan semua antrian?</span>
				<LoadingButton size="sm" variant="destructive" onclick={() => void cancelAll()} loading={busy.cancel_all} loadingLabel="Membatalkan..." label="Ya" />
				<Button size="sm" variant="ghost" onclick={() => (confirmKey = '')}>Tidak</Button>
			{:else}
				<LoadingButton size="sm" variant="outline" onclick={() => (confirmKey = 'cancel_all')} loading={busy.cancel_all} loadingLabel="Memproses..." disabled={busy.cancel_all}
					class="text-destructive border-destructive/40 hover:bg-destructive/10">
					✕ Batalkan Semua
				</LoadingButton>
			{/if}
		</div>
	</div>

	{#if operationState}
		<OperationStatusPanel {...operationState} />
	{/if}

	<AsyncContent promise={overviewPromise} onerror={handleOverviewRenderError}>
		{#snippet pending()}
			<div class="space-y-4 rounded-2xl border border-slate-200 bg-white p-5">
				<div class="grid gap-4 sm:grid-cols-3">
					{#each Array.from({ length: 3 }) as _, index (`pusaka-worker-skeleton-${index}`)}
						<div class="space-y-3 rounded-xl border border-slate-100 p-4">
							<Skeleton class="h-4 w-28" />
							<Skeleton class="h-8 w-20" />
						</div>
					{/each}
				</div>
				<Skeleton class="h-28 w-full" />
				<div class="space-y-3">
					<Skeleton class="h-10 w-40" />
					{#each Array.from({ length: 5 }) as _, index (`pusaka-job-skeleton-${index}`)}
						<Skeleton class="h-11 w-full" />
					{/each}
				</div>
			</div>
		{/snippet}
		{#snippet failed(error, reset)}
			<RecoveryPanel title="PUSAKA Belum Merespons Penuh" message={overviewErrorMessage(error)} onRetry={() => retryOverview(reset)} />
		{/snippet}
		{#snippet children(value)}
			{@const overview = value as PusakaOverview}
			{@const currentQueueStats = overview.queueStats}
			{@const currentWorkerStatus = overview.workerStatus}
			{@const currentRecentJobs = overview.recentJobs}

	<!-- Worker status -->
	<div class="grid gap-4 sm:grid-cols-3">
		<Card.Root>
			<Card.Header class="pb-2">
				<Card.Description>Worker Aktif</Card.Description>
			</Card.Header>
			<Card.Content class="pt-0">
				<div class="flex items-center gap-2">
					<span class="text-2xl font-bold">{currentWorkerStatus?.total ?? '—'}</span>
					{#if currentWorkerStatus}
						<Badge variant={currentWorkerStatus.total > 0 ? 'default' : 'destructive'}>
							{currentWorkerStatus.total > 0 ? 'Online' : 'Offline'}
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
					{currentWorkerStatus?.active_workers?.reduce((s, w) => s + w.active_consumers, 0) ?? '—'}
				</span>
				{#if currentWorkerStatus}
					<span class="text-sm text-muted-foreground ml-1">
						/ {currentWorkerStatus.active_workers?.reduce((s, w) => s + w.target_concurrency, 0)} kapasitas
					</span>
				{/if}
			</Card.Content>
		</Card.Root>

		<Card.Root>
			<Card.Header class="pb-2">
			<Card.Description>Antrian Job</Card.Description>
			</Card.Header>
			<Card.Content class="pt-0">
				{#if currentWorkerStatus?.queue}
					{@const q = currentWorkerStatus.queue}
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
	<QueueMonitor stats={currentQueueStats} />

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
					{#each currentRecentJobs as j (j.id)}
						<Table.Row>
							<Table.Cell class="text-xs text-muted-foreground whitespace-nowrap">{fmtDt(j.created_at)}</Table.Cell>
							<Table.Cell class="font-medium">{j.nama || j.employee_nama || '—'}</Table.Cell>
							<Table.Cell><Badge variant="outline">{runTypeLabel(j.run_type)}</Badge></Table.Cell>
							<Table.Cell><Badge variant={statusVariant(j.status)}>{statusLabel(j.status)}</Badge></Table.Cell>
							<Table.Cell class="hidden sm:table-cell text-xs text-muted-foreground">{j.claimed_by || '—'}</Table.Cell>
						</Table.Row>
					{:else}
						<Table.Row>
							<Table.Cell colspan={5} class="p-4">
								<EmptyStatePanel
									compact
									title="Belum ada job"
									description="Jalankan rekap atau trigger scheduler untuk mulai membentuk antrean kerja PUSAKA di dashboard ini."
								/>
							</Table.Cell>
						</Table.Row>
					{/each}
				</Table.Body>
			</Table.Root>
			</div>

			<div class="grid gap-3 p-4 lg:hidden">
				{#each currentRecentJobs as j (j.id)}
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
					<EmptyStatePanel
						title="Belum ada job"
						description="Jalankan rekap atau trigger scheduler untuk mulai membentuk antrean kerja PUSAKA di dashboard ini."
					/>
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
		{/snippet}
	</AsyncContent>

</div>
