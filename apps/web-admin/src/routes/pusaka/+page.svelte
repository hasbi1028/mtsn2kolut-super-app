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
	import { readClientApiData } from '$lib/client/api';
	import { trackInternalAnalyticsEvent } from '$lib/analytics/internal-analytics';

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
	let overviewRequestId = 0;

	function emptyQueueStats(): QueueStats {
		return { queued: 0, running: 0, success: 0, failed: 0, retry_due: 0, total: 0 };
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
			fetch('/api/pusaka/jobs/stats').then((response) => readClientApiData<Partial<QueueStats>>(response, 'Gagal memuat ringkasan antrian PUSAKA')),
			fetch('/api/pusaka/jobs?limit=5').then((response) => readClientApiData<JobListPayload | RecentJob[]>(response, 'Gagal memuat pekerjaan terbaru PUSAKA')),
			fetch('/api/pusaka/worker/status').then((response) => readClientApiData<WorkerStatus | null>(response, 'Gagal memuat status petugas sistem PUSAKA')),
		]);

		return {
			queueStats: normalizeQueueStats(queueData),
			recentJobs: normalizeJobs(jobsData),
			workerStatus: workerData ?? null,
		};
	}

	function loadOverview() {
		const requestId = ++overviewRequestId;
		overviewPromise = fetchOverview()
			.then((overview) => {
				if (requestId === overviewRequestId) {
					applyOverview(overview);
					return overview;
				}
				return currentOverview();
			})
			.catch((error: unknown) => {
				if (requestId === overviewRequestId) throw error;
				return currentOverview();
			});
		return overviewPromise;
	}

	async function refreshOverview(showFailureToast = false) {
		if (!overviewPromise) {
			await loadOverview();
			return;
		}
		const requestId = ++overviewRequestId;
		try {
			const overview = await fetchOverview();
			if (requestId === overviewRequestId) {
				applyOverview(overview);
				overviewPromise = Promise.resolve(overview);
			}
		} catch (error) {
			if (requestId !== overviewRequestId) return;
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
		return 'Gagal memuat status petugas sistem, ringkasan antrian, atau pekerjaan terbaru. Periksa layanan sistem PUSAKA, lalu coba lagi.';
	}

	function handleOverviewRenderError(error: unknown, reset: () => void) {
		console.error('Ringkasan PUSAKA gagal ditampilkan', error);
		reset();
	}

	async function act(key: string, fn: () => Promise<Response>, successMsg: string) {
		busy = { ...busy, [key]: true };
		try {
			const res  = await fn();
			const data = await readClientApiData<PusakaActionResponse>(res, 'Gagal menjalankan pekerjaan PUSAKA');
			operationState = {
				tone: key === 'cancel_all' ? 'warning' : 'success',
				title: key === 'cancel_all' ? 'Antrian Dibatalkan' : 'Pekerjaan PUSAKA Berhasil',
				message: successMsg + (data.cancelled != null ? ` (${data.cancelled} pekerjaan)` : ''),
			};
			showToast(successMsg + (data.cancelled != null ? ` (${data.cancelled} pekerjaan)` : ''), 'ok');
		} catch (error) {
			operationState = {
				tone: 'error',
				title: 'Pekerjaan PUSAKA Gagal',
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
			message: 'Rekap massal akan membuat pekerjaan untuk seluruh akun PUSAKA yang aktif. Gunakan hanya saat operator siap memantau antrian.',
			challenge: 'REKAP',
			confirmLabel: 'Mulai Rekap',
			tone: 'warning'
		}))) return;
		busy = { ...busy, rekap: true };
		try {
			const res  = await fetch('/api/pusaka/jobs/run-all', { method: 'POST', headers: { 'content-type': 'application/json' }, body: JSON.stringify({ run_type: 'morning' }) });
			const data = await readClientApiData<PusakaActionResponse>(res, 'Gagal menjalankan rekap massal');
			operationState = {
				tone: 'warning',
				title: 'Rekap Massal Diantrekan',
				message: `Sistem menambahkan ${data.inserted ?? 0} pekerjaan baru. Pantau hasilnya di antrian dan worker status sebelum mengulangi operasi ini.`,
			};
			showToast(`Rekap di-queue: ${data.inserted ?? 0} pekerjaan baru`, 'ok');
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

	const triggerSched  = ()          => act('sched',      () => fetch('/api/pusaka/scheduler/tick', { method: 'POST' }), 'Jadwal otomatis dijalankan');
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
		void trackInternalAnalyticsEvent('pusaka.dashboard_view', {
			pathname: window.location.pathname,
			metadata: { page_key: 'pusaka' }
		});
		void loadOverview();
		const itv = setInterval(() => void refreshOverview(false), 10_000);
		return () => clearInterval(itv);
	});
</script>

<svelte:head><title>PUSAKA Kemenag — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6">

	<!-- Breadcrumb -->
	<div class="flex items-center gap-2 text-sm text-base-content/70">
		<a href={resolve('/')} class="hover:text-base-content">Beranda</a>
		<span>/</span>
		<span class="text-base-content font-medium">PUSAKA</span>
	</div>

	<!-- Header -->
	<div class="flex flex-wrap items-start justify-between gap-4">
		<div>
			<h1 class="text-2xl font-semibold text-base-content">Kontrol PUSAKA Kemenag</h1>
			<p class="text-sm text-base-content/70 mt-1">Pantau dan kelola penarikan data kehadiran dari PUSAKA Kemenag</p>
		</div>
		<div class="flex flex-wrap gap-2">
			<LoadingButton variant="outline" size="sm" onclick={() => void triggerSched()} loading={busy.sched} loadingLabel="Memproses..." label="">
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
				</svg>
				Jalankan Jadwal Otomatis
			</LoadingButton>
			<LoadingButton size="sm" onclick={() => void runRekap()} loading={busy.rekap} loadingLabel="Memproses..." label="">
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
				</svg>
				Mulai Rekap
			</LoadingButton>
			{#if confirmKey === 'cancel_all'}
				<span class="self-center text-xs text-warning">Batalkan semua antrian?</span>
				<LoadingButton size="sm" variant="destructive" onclick={() => void cancelAll()} loading={busy.cancel_all} loadingLabel="Membatalkan..." label="Ya" />
				<Button size="sm" variant="ghost" onclick={() => (confirmKey = '')}>Tidak</Button>
			{:else}
				<LoadingButton size="sm" variant="outline" onclick={() => (confirmKey = 'cancel_all')} loading={busy.cancel_all} loadingLabel="Memproses..." disabled={busy.cancel_all}
					class="text-destructive border-destructive/40 hover:bg-destructive/10">
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
					Batalkan Semua
				</LoadingButton>
			{/if}
		</div>
	</div>

	{#if operationState}
		<OperationStatusPanel {...operationState} />
	{/if}

	<AsyncContent promise={overviewPromise} onerror={handleOverviewRenderError}>
		{#snippet pending()}
			<div class="space-y-4 rounded-2xl border border-base-300 bg-base-100 p-5">
				<div class="grid gap-4 sm:grid-cols-3">
					{#each Array.from({ length: 3 }) as _, index (`pusaka-worker-skeleton-${index}`)}
						<div class="space-y-3 rounded-xl border border-base-300 p-4">
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
				<Card.Description>Petugas Sistem Aktif</Card.Description>
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
				<Card.Description>Proses Aktif</Card.Description>
			</Card.Header>
			<Card.Content class="pt-0">
				<span class="text-2xl font-bold">
					{currentWorkerStatus?.active_workers?.reduce((s, w) => s + w.active_consumers, 0) ?? '—'}
				</span>
				{#if currentWorkerStatus}
					<span class="text-sm text-base-content/70 ml-1">
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
							<span class="badge badge-sm badge-info">{q.running} berjalan</span>
						{/if}
						{#if q.queued > 0}
							<span class="badge badge-sm badge-ghost">{q.queued} antre</span>
						{/if}
						{#if q.running === 0 && q.queued === 0}
							<span class="badge badge-sm badge-outline">Idle</span>
						{/if}
					</div>
					<p class="text-xs text-base-content/70 mt-1">
						<span class="badge badge-xs badge-success">{q.success} sukses</span>
						<span class="badge badge-xs badge-error">{q.failed} gagal</span>
					</p>
				{:else}
					<span class="text-2xl font-bold">—</span>
				{/if}
			</Card.Content>
		</Card.Root>
	</div>

	<!-- Queue monitor -->
	<QueueMonitor stats={currentQueueStats} />

	<!-- Recent jobs -->
	<Card.Root class="overflow-hidden border-base-300 shadow-sm">
		<Card.Header class="pb-3">
			<div class="flex items-center justify-between">
				<Card.Title class="text-base">Pekerjaan Terbaru</Card.Title>
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
						<Table.Head class="hidden sm:table-cell">Petugas Sistem</Table.Head>
					</Table.Row>
				</Table.Header>
				<Table.Body>
					{#each currentRecentJobs as j (j.id)}
						<Table.Row>
							<Table.Cell class="text-xs text-base-content/70 whitespace-nowrap">{fmtDt(j.created_at)}</Table.Cell>
							<Table.Cell class="font-medium">{j.nama || j.employee_nama || '—'}</Table.Cell>
							<Table.Cell><Badge variant="outline">{runTypeLabel(j.run_type)}</Badge></Table.Cell>
							<Table.Cell><Badge variant={statusVariant(j.status)}>{statusLabel(j.status)}</Badge></Table.Cell>
							<Table.Cell class="hidden sm:table-cell text-xs text-base-content/70">{j.claimed_by || '—'}</Table.Cell>
						</Table.Row>
					{:else}
						<Table.Row>
							<Table.Cell colspan={5} class="p-4">
								<EmptyStatePanel
									compact
									title="Belum ada pekerjaan"
									description="Jalankan rekap atau jalankan jadwal otomatis untuk mulai membentuk antrean kerja PUSAKA di halaman ini."
								/>
							</Table.Cell>
						</Table.Row>
					{/each}
				</Table.Body>
			</Table.Root>
			</div>

			<div class="lg:hidden">
				{#if currentRecentJobs.length === 0}
					<div class="p-4">
						<EmptyStatePanel
							title="Belum ada pekerjaan"
							description="Jalankan rekap atau jalankan jadwal otomatis untuk mulai membentuk antrean kerja PUSAKA di halaman ini."
						/>
					</div>
				{:else}
					<ul class="divide-y divide-border border-b border-base-300">
						{#each currentRecentJobs as j (j.id)}
							<li class="flex items-start gap-3 px-4 py-2.5">
								<div class="min-w-0 grow">
									<p class="truncate text-sm font-medium text-base-content">{j.nama || j.employee_nama || '—'}</p>
									<p class="truncate text-[11px] text-base-content/70">{fmtDt(j.created_at)} · {j.claimed_by || 'Belum diambil'}</p>
								</div>
								<div class="flex shrink-0 flex-col items-end gap-1">
									<Badge variant={statusVariant(j.status)}>{statusLabel(j.status)}</Badge>
									<Badge variant="outline" class="text-[10px]">{runTypeLabel(j.run_type)}</Badge>
								</div>
							</li>
						{/each}
					</ul>
				{/if}
			</div>
		</Card.Content>
	</Card.Root>

	<!-- Quick navigation -->
	<div class="grid gap-3 sm:grid-cols-2">
		<Button variant="outline" href={resolve('/pusaka/employees')} class="h-auto py-4 flex-col items-start text-left gap-1">
			<span class="font-semibold">Pegawai PUSAKA</span>
			<span class="text-xs text-base-content/70 font-normal">Atur akun, jadwal, dan pekerjaan untuk pegawai yang memenuhi syarat</span>
		</Button>
		<Button variant="outline" href={resolve('/pusaka/kehadiran')} class="h-auto py-4 flex-col items-start text-left gap-1">
			<span class="font-semibold">Data Kehadiran</span>
			<span class="text-xs text-base-content/70 font-normal">Rekap harian dari PUSAKA Kemenag</span>
		</Button>
		<Button variant="outline" href={resolve('/pusaka/summary')} class="h-auto py-4 flex-col items-start text-left gap-1">
			<span class="font-semibold">Ringkasan Kehadiran</span>
			<span class="text-xs text-base-content/70 font-normal">Akumulasi per periode / bulan</span>
		</Button>
	</div>
		{/snippet}
	</AsyncContent>

</div>
