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
		return 'Gagal memuat status PUSAKA. Coba lagi.';
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
		busy = { ...busy, rekap: true };
		try {
			const res  = await fetch('/api/pusaka/jobs/run-all', { method: 'POST', headers: { 'content-type': 'application/json' }, body: JSON.stringify({ run_type: 'morning' }) });
			const data = await readClientApiData<PusakaActionResponse>(res, 'Gagal menjalankan rekap massal');
			operationState = {
				tone: 'warning',
				title: 'Rekap Diantrekan',
				message: `${data.inserted ?? 0} pekerjaan baru ditambahkan.`,
			};
			showToast(`Rekap di-queue: ${data.inserted ?? 0} pekerjaan`, 'ok');
		} catch (error) {
			operationState = {
				tone: 'error',
				title: 'Rekap Gagal',
				message: overviewErrorMessage(error),
			};
			showToast(overviewErrorMessage(error), 'err');
		} finally {
			busy = { ...busy, rekap: false };
			confirmKey = '';
			await refreshOverview(true);
		}
	}

	const triggerSched  = ()          => act('sched',      () => fetch('/api/pusaka/scheduler/tick', { method: 'POST' }), 'Jadwal otomatis dijalankan');
	const cancelAll     = ()          => act('cancel_all', () => fetch('/api/pusaka/jobs/cancel-all',{ method: 'POST' }), 'Semua antrian dibatalkan');

	// ── Schedule management ──
	type PusakaSchedule = { id: string; label: string; run_type: string; run_time: string; is_enabled: boolean };
	let schedules = $state<PusakaSchedule[] | null>(null);
	let saving = $state(false);

	async function loadSchedules() {
		schedules = null;
		try {
			const res = await fetch('/api/pusaka/schedules');
			const data = await readClientApiData<PusakaSchedule[]>(res, 'Gagal memuat jadwal');
			schedules = Array.isArray(data) ? data : [];
		} catch {
			schedules = [];
		}
	}

	async function createSchedule() {
		saving = true;
		try {
			const res = await fetch('/api/pusaka/schedules', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ label: 'Rekap Harian', run_time: '23:00', run_type: 'morning', is_enabled: true })
			});
			if (!res.ok) { const err = await res.json(); throw new Error(err.error || 'Gagal membuat jadwal'); }
			toast.success('Jadwal baru ditambahkan (jam 23:00 WITA)');
			await loadSchedules();
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Gagal membuat jadwal');
		} finally { saving = false; }
	}

	async function updateSchedule(sched: PusakaSchedule) {
		try {
			const res = await fetch(`/api/pusaka/schedules/${sched.id}`, {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ label: sched.label, run_time: sched.run_time, is_enabled: sched.is_enabled })
			});
			if (!res.ok) { const err = await res.json(); throw new Error(err.error || 'Gagal menyimpan'); }
			toast.success(`Jadwal diubah ke jam ${sched.run_time} WITA`);
			await loadSchedules();
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Gagal menyimpan jadwal');
		}
	}

	async function toggleSchedule(sched: PusakaSchedule) {
		sched.is_enabled = !sched.is_enabled;
		await updateSchedule(sched);
	}

	async function deleteSchedule(sched: PusakaSchedule) {
		try {
			const res = await fetch(`/api/pusaka/schedules/${sched.id}`, { method: 'DELETE' });
			if (!res.ok) { const err = await res.json(); throw new Error(err.error || 'Gagal menghapus'); }
			toast.success('Jadwal dihapus');
			await loadSchedules();
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Gagal menghapus jadwal');
		}
	}

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
		void loadSchedules();
		const itv = setInterval(() => void refreshOverview(false), 10_000);
		return () => clearInterval(itv);
	});
</script>

<svelte:head><title>PUSAKA Kemenag — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-4">

	<!-- Header: compact -->
	<div class="space-y-3">
		<h1 class="text-lg font-black tracking-tight text-base-content lg:text-2xl">Kontrol PUSAKA</h1>
		<div class="flex items-center gap-2 text-xs text-base-content/70">
			<a href={resolve('/')} class="hover:text-base-content hidden lg:inline">Beranda</a>
			<span class="hidden lg:inline">/</span>
			<span class="hidden lg:inline font-medium text-base-content">PUSAKA</span>
			<span class="hidden lg:inline">—</span>
			<span class="font-medium text-base-content">Pantau data kehadiran PUSAKA Kemenag</span>
		</div>

		<!-- Action buttons: horizontal scroll on mobile -->
		<div class="flex items-center gap-2 overflow-x-auto pb-1">
			<LoadingButton variant="outline" size="sm" onclick={() => void triggerSched()} loading={busy.sched} loadingLabel="Memproses..." label="" class="shrink-0">
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" /></svg>
				Jalankan Jadwal
			</LoadingButton>
			{#if confirmKey === 'rekap'}
				<span class="shrink-0 text-xs text-warning">Mulai rekap?</span>
				<LoadingButton size="sm" onclick={() => void runRekap()} loading={busy.rekap} loadingLabel="..." label="Ya" class="shrink-0" />
				<Button size="sm" variant="ghost" onclick={() => (confirmKey = '')} class="shrink-0">✕</Button>
			{:else}
				<LoadingButton size="sm" onclick={() => (confirmKey = 'rekap')} loading={busy.rekap} loadingLabel="..." label="" class="shrink-0">
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
					Mulai Rekap
				</LoadingButton>
			{/if}
			{#if confirmKey === 'cancel_all'}
				<span class="shrink-0 text-xs text-warning">Batalkan?</span>
				<LoadingButton size="sm" variant="destructive" onclick={() => void cancelAll()} loading={busy.cancel_all} loadingLabel="..." label="Ya" class="shrink-0" />
				<Button size="sm" variant="ghost" onclick={() => (confirmKey = '')} class="shrink-0">✕</Button>
			{:else}
				<LoadingButton size="sm" variant="outline" onclick={() => (confirmKey = 'cancel_all')} loading={busy.cancel_all} loadingLabel="..." disabled={busy.cancel_all} class="shrink-0 text-destructive border-destructive/40 hover:bg-destructive/10">
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
					Batalkan
				</LoadingButton>
			{/if}
		</div>
	</div>

	{#if operationState}
		<OperationStatusPanel {...operationState} />
	{/if}

	<AsyncContent promise={overviewPromise} onerror={handleOverviewRenderError}>
		{#snippet pending()}
			<div class="space-y-3 rounded-xl border border-base-300 bg-base-100 p-4">
				<div class="flex gap-2">{#each Array(4) as _, i}<Skeleton class="h-12 flex-1 rounded-lg" />{/each}</div>
				<div class="space-y-2">
					{#each Array(5) as _, i}<Skeleton class="h-10 w-full" />{/each}
				</div>
			</div>
		{/snippet}
		{#snippet failed(error, reset)}
			<RecoveryPanel title="PUSAKA Error" message={overviewErrorMessage(error)} onRetry={() => retryOverview(reset)} />
		{/snippet}
		{#snippet children(value)}
			{@const overview = value as PusakaOverview}
			{@const currentQueueStats = overview.queueStats}
			{@const currentWorkerStatus = overview.workerStatus}
			{@const currentRecentJobs = overview.recentJobs}

		<!-- Worker status: compact inline -->
		<div class="flex items-center gap-3 rounded-xl border border-base-300 bg-base-100 px-3 py-2.5 text-xs">
			{#if currentWorkerStatus}
				<Badge variant={currentWorkerStatus.total > 0 ? 'default' : 'destructive'}>
					{currentWorkerStatus.total > 0 ? 'Online' : 'Offline'}
				</Badge>
				<span class="text-base-content/70">
					Petugas: <span class="font-bold text-base-content">{currentWorkerStatus.total}</span>
					{#if currentWorkerStatus.active_workers?.length}
						<span class="ml-2">Proses: <span class="font-bold text-base-content">{currentWorkerStatus.active_workers.reduce((s, w) => s + w.active_consumers, 0)}</span>/{currentWorkerStatus.active_workers.reduce((s, w) => s + w.target_concurrency, 0)}</span>
					{/if}
				</span>
			{:else}
				<Badge variant="destructive">Offline</Badge>
				<span class="text-base-content/70">Petugas tidak aktif</span>
			{/if}
		</div>

		<!-- Queue monitor compact -->
		<QueueMonitor stats={currentQueueStats} />

		<!-- Recent jobs -->
		<Card.Root class="overflow-hidden border-base-300 shadow-sm">
			<Card.Header class="px-4 py-3">
				<div class="flex items-center justify-between">
					<Card.Title class="text-sm font-bold">Pekerjaan Terbaru</Card.Title>
					<Button variant="ghost" size="sm" href={resolve('/pusaka/antrian')} class="text-xs h-7 px-2">Semua →</Button>
				</div>
			</Card.Header>
			<Card.Content class="p-0">
				<!-- Desktop table -->
				<div class="hidden overflow-x-auto lg:block">
					<Table.Root>
						<Table.Header>
							<Table.Row>
								<Table.Head>Waktu</Table.Head>
								<Table.Head>Nama</Table.Head>
								<Table.Head>Tipe</Table.Head>
								<Table.Head>Status</Table.Head>
							</Table.Row>
						</Table.Header>
						<Table.Body>
							{#each currentRecentJobs.slice(0, 5) as j (j.id)}
								<Table.Row>
									<Table.Cell class="text-xs text-base-content/70 whitespace-nowrap">{fmtDt(j.created_at)}</Table.Cell>
									<Table.Cell class="font-medium">{j.nama || j.employee_nama || '—'}</Table.Cell>
									<Table.Cell><Badge variant="outline">{runTypeLabel(j.run_type)}</Badge></Table.Cell>
									<Table.Cell><Badge variant={statusVariant(j.status)}>{statusLabel(j.status)}</Badge></Table.Cell>
								</Table.Row>
							{:else}
								<Table.Row><Table.Cell colspan={4} class="p-4"><EmptyStatePanel compact title="Belum ada pekerjaan" description="Jalankan rekap untuk mulai." /></Table.Cell></Table.Row>
							{/each}
						</Table.Body>
					</Table.Root>
				</div>
				<!-- Mobile list -->
				<div class="lg:hidden">
					{#if currentRecentJobs.length === 0}
						<div class="p-4"><EmptyStatePanel compact title="Belum ada pekerjaan" description="Jalankan rekap untuk mulai." /></div>
					{:else}
						<ul class="divide-y divide-border">
							{#each currentRecentJobs.slice(0, 5) as j (j.id)}
								<li class="flex items-center gap-3 px-4 py-2.5">
									<div class="min-w-0 flex-1">
										<p class="truncate text-sm font-semibold text-base-content">{j.nama || j.employee_nama || '—'}</p>
										<p class="truncate text-[10px] text-base-content/60">{fmtDt(j.created_at)}</p>
									</div>
									<div class="flex shrink-0 flex-col items-end gap-0.5">
										<Badge variant={statusVariant(j.status)} class="text-[10px]">{statusLabel(j.status)}</Badge>
										<span class="text-[9px] text-base-content/50">{runTypeLabel(j.run_type)}</span>
									</div>
								</li>
							{/each}
						</ul>
					{/if}
				</div>
			</Card.Content>
		</Card.Root>

		<!-- Quick nav: compact 2x2 grid -->
		<div class="grid grid-cols-2 gap-2">
			<a href={resolve('/pusaka/employees')} class="group flex items-center gap-2 rounded-xl border border-base-300 bg-base-100 px-3 py-3 transition-all hover:bg-primary/5 hover:border-primary/30">
				<div class="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary">
					<svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 7.292 4 4 0 010-7.292zM15 21H9a2 2 0 01-2-2V12a2 2 0 012-2h6a2 2 0 012 2v7a2 2 0 01-2 2z" /></svg>
				</div>
				<div class="min-w-0">
					<p class="truncate text-xs font-bold text-base-content">Pegawai</p>
					<p class="truncate text-[10px] text-base-content/60">Akun & jadwal</p>
				</div>
			</a>
			<a href={resolve('/pusaka/kehadiran')} class="group flex items-center gap-2 rounded-xl border border-base-300 bg-base-100 px-3 py-3 transition-all hover:bg-emerald-500/5 hover:border-emerald-500/30">
				<div class="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-emerald-500/10 text-emerald-600">
					<svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" /></svg>
				</div>
				<div class="min-w-0">
					<p class="truncate text-xs font-bold text-base-content">Kehadiran</p>
					<p class="truncate text-[10px] text-base-content/60">Data harian</p>
				</div>
			</a>
			<a href={resolve('/pusaka/summary')} class="group flex items-center gap-2 rounded-xl border border-base-300 bg-base-100 px-3 py-3 transition-all hover:bg-amber-500/5 hover:border-amber-500/30">
				<div class="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-amber-500/10 text-amber-600">
					<svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" /></svg>
				</div>
				<div class="min-w-0">
					<p class="truncate text-xs font-bold text-base-content">Ringkasan</p>
					<p class="truncate text-[10px] text-base-content/60">Rekap periode</p>
				</div>
			</a>
			<a href={resolve('/pusaka/antrian')} class="group flex items-center gap-2 rounded-xl border border-base-300 bg-base-100 px-3 py-3 transition-all hover:bg-sky-500/5 hover:border-sky-500/30">
				<div class="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-sky-500/10 text-sky-600">
					<svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" /></svg>
				</div>
				<div class="min-w-0">
					<p class="truncate text-xs font-bold text-base-content">Antrian</p>
					<p class="truncate text-[10px] text-base-content/60">Semua pekerjaan</p>
				</div>
			</a>
		</div>

	{/snippet}
</AsyncContent>

		<!-- ═══ Jadwal Otomatis Rekap ═══ (di luar AsyncContent) -->
		<Card.Root class="overflow-hidden border-base-300 shadow-sm">
			<Card.Header class="px-4 py-3">
				<div class="flex items-center justify-between">
					<Card.Title class="text-sm font-bold">⏰ Jadwal Otomatis Rekap</Card.Title>
					<Button variant="ghost" size="sm" onclick={() => void loadSchedules()} class="text-xs h-7 px-2">🔄</Button>
				</div>
			</Card.Header>
			<Card.Content class="p-4 space-y-3">
				{#if schedules === null}
					<Skeleton class="h-10 w-full" />
				{:else if schedules.length === 0}
					<p class="text-xs text-base-content/60">Belum ada jadwal otomatis. Tambah jadwal untuk rekap harian.</p>
				{:else}
					{#each schedules as sched (sched.id)}
						<div class="flex items-center gap-3 rounded-lg border border-base-300 p-3">
							<div class="flex-1 min-w-0">
								<p class="text-sm font-semibold">{sched.label || 'Rekap Harian'}</p>
								<p class="text-xs text-base-content/60">Jam {sched.run_time} WITA · {sched.run_type === 'morning' ? 'Rekap' : sched.run_type}</p>
							</div>
							<input type="time" bind:value={sched.run_time}
								onchange={() => void updateSchedule(sched)}
								class="h-8 rounded-lg border border-input bg-background px-2 text-xs" />
							<button
								type="button"
								role="switch"
								aria-checked={sched.is_enabled}
								onclick={() => void toggleSchedule(sched)}
								class="relative inline-flex h-5 w-9 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors {sched.is_enabled ? 'bg-primary' : 'bg-base-300'}"
							>
								<span class="pointer-events-none inline-block h-4 w-4 rounded-full bg-white shadow transform ring-0 transition {sched.is_enabled ? 'translate-x-4' : 'translate-x-0'}" />
							</button>
							<button type="button" onclick={() => void deleteSchedule(sched)}
								class="text-xs text-destructive hover:underline shrink-0">Hapus</button>
						</div>
					{/each}
				{/if}
				<button type="button" onclick={() => void createSchedule()}
					disabled={saving}
					class="w-full text-xs font-medium text-primary hover:underline py-1">
					+ Tambah Jadwal Baru
				</button>
			</Card.Content>
		</Card.Root>

</div>

<!-- FAB Rekap: mobile only -->
<div class="fixed bottom-20 right-4 z-50 lg:hidden">
	{#if confirmKey === 'rekap'}
		<div class="flex items-center gap-2 rounded-2xl border border-warning/30 bg-warning/10 px-3 py-2 shadow-lg backdrop-blur">
			<span class="text-xs font-bold text-warning">Mulai rekap?</span>
			<LoadingButton size="sm" onclick={() => void runRekap()} loading={busy.rekap} loadingLabel="..." label="Ya" class="shrink-0" />
			<Button size="sm" variant="ghost" onclick={() => (confirmKey = '')} class="shrink-0 h-7 w-7 p-0">✕</Button>
		</div>
	{:else}
		<button
			type="button"
			onclick={() => (confirmKey = 'rekap')}
			disabled={busy.rekap}
			class="flex h-14 w-14 items-center justify-center rounded-full bg-primary text-primary-content shadow-lg shadow-primary/30 transition-all hover:bg-primary/90 hover:shadow-xl active:scale-95 disabled:opacity-50"
		>
			{#if busy.rekap}
				<span class="loading loading-spinner loading-sm"></span>
			{:else}
				<svg class="h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
			{/if}
		</button>
	{/if}
</div>
