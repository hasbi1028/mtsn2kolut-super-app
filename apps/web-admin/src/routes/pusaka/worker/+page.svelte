<script lang="ts">
	import { onMount } from 'svelte';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { toast } from '$lib/components/ui/sonner';
	import EmptyStatePanel from '$lib/components/EmptyStatePanel.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { readClientApiData } from '$lib/client/api';
	import { readUrlParam, writeUrlParam } from '$lib/stores/persistent';

	type WorkerInfo = {
		worker_id: string;
		status: 'active' | 'stale' | 'offline';
		enabled?: boolean;
		active_consumers: number;
		target_concurrency: number;
		headless: boolean;
		last_sync_at: string;
		reported_at: string;
	};

	type WorkerStatusData = {
		workers: WorkerInfo[];
		total: number;
		active_count: number;
		global_max_concurrent: string;
		queue: { queued: number; running: number; success: number; failed: number };
		last_checked: string;
	};

	type FilterKey = 'all' | 'active' | 'stale' | 'offline';
	type WorkerCap = { worker_id: string; cap: number };

	const REFRESH_MS = 15_000;

	let loading = $state(true);
	let error = $state('');
	let data = $state<WorkerStatusData | null>(null);
	let lastRefreshedAt = $state<Date | null>(null);
	let refreshTimer: ReturnType<typeof setInterval> | undefined;
	let filter = $state<FilterKey>(readUrlParam('status', 'all') as FilterKey);
	let isRefreshing = $state(false);
	let caps = $state<Record<string, number>>({});
	let capDrafts = $state<Record<string, string>>({});
	let capSaving = $state<Record<string, boolean>>({});
	let enabling = $state<Record<string, boolean>>({});

	const filterOptions: { key: FilterKey; label: string }[] = [
		{ key: 'all', label: 'Semua' },
		{ key: 'active', label: 'Aktif' },
		{ key: 'stale', label: 'Stale' },
		{ key: 'offline', label: 'Offline' }
	];

	const statusLabel: Record<WorkerInfo['status'], string> = {
		active: 'Aktif',
		stale: 'Stale',
		offline: 'Offline'
	};

	const statusBadgeClass: Record<WorkerInfo['status'], string> = {
		active: 'border-transparent bg-emerald-100 text-emerald-700',
		stale: 'border-transparent bg-amber-100 text-amber-700',
		offline: 'border-transparent bg-rose-100 text-rose-700'
	};

	async function loadWorkerStatus(background = false): Promise<void> {
		if (!background) {
			loading = true;
		} else {
			isRefreshing = true;
		}
		error = '';
		try {
			const res = await fetch('/api/pusaka/worker/status');
			const payload = await readClientApiData<WorkerStatusData>(res);
			data = payload;
			lastRefreshedAt = new Date();
		} catch (e) {
			error = (e as Error)?.message ?? 'Gagal memuat status worker';
		} finally {
			loading = false;
			isRefreshing = false;
		}
	}

	async function loadCaps(): Promise<void> {
		try {
			const res = await fetch('/api/pusaka/worker/caps');
			const payload = await readClientApiData<{ caps: WorkerCap[] }>(res);
			const map: Record<string, number> = {};
			const drafts: Record<string, string> = {};
			for (const c of payload.caps ?? []) {
				map[c.worker_id] = c.cap;
				drafts[c.worker_id] = String(c.cap);
			}
			caps = map;
			capDrafts = drafts;
		} catch {
			// Gagal memuat caps tidak boleh menggagalkan halaman utama.
		}
	}

	async function saveCap(workerId: string): Promise<void> {
		const raw = capDrafts[workerId];
		const cap = Number(raw);
		if (!Number.isInteger(cap) || cap < 1 || cap > 35) {
			toast.error('Cap harus bilangan bulat antara 1 dan 35');
			return;
		}
		capSaving = { ...capSaving, [workerId]: true };
		try {
			const res = await fetch(`/api/pusaka/worker/caps/${encodeURIComponent(workerId)}`, {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ cap })
			});
			const payload = await readClientApiData<{ cap: number }>(res);
			caps = { ...caps, [workerId]: payload.cap };
			capDrafts = { ...capDrafts, [workerId]: String(payload.cap) };
			toast.success(`Cap ${workerId} diatur ke ${payload.cap} (berlaku ≤30 detik)`);
			void loadWorkerStatus(true);
		} catch (e) {
			toast.error((e as Error)?.message ?? 'Gagal menyimpan cap');
		} finally {
			capSaving = { ...capSaving, [workerId]: false };
		}
	}

	async function resetCap(workerId: string): Promise<void> {
		capSaving = { ...capSaving, [workerId]: true };
		try {
			const res = await fetch(`/api/pusaka/worker/caps/${encodeURIComponent(workerId)}`, {
				method: 'DELETE'
			});
			await readClientApiData<{ ok: boolean }>(res);
			const next = { ...caps };
			delete next[workerId];
			caps = next;
			const drafts = { ...capDrafts };
			delete drafts[workerId];
			capDrafts = drafts;
			toast.success(`Cap ${workerId} direset — kembali ke env/global`);
			void loadWorkerStatus(true);
		} catch (e) {
			toast.error((e as Error)?.message ?? 'Gagal mereset cap');
		} finally {
			capSaving = { ...capSaving, [workerId]: false };
		}
	}

	async function toggleEnabled(w: WorkerInfo): Promise<void> {
		const next = !(w.enabled ?? true);
		enabling = { ...enabling, [w.worker_id]: true };
		try {
			const res = await fetch(
				`/api/pusaka/worker/enabled/${encodeURIComponent(w.worker_id)}`,
				{
					method: 'PUT',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify({ enabled: next })
				}
			);
			const payload = await readClientApiData<{ enabled: boolean }>(res);
			w.enabled = payload.enabled;
			toast.success(
				next
					? `${w.worker_id} diaktifkan — akan menerima job lagi`
					: `${w.worker_id} dinonaktifkan — tidak akan menerima job`
			);
		} catch (e) {
			toast.error((e as Error)?.message ?? 'Gagal mengubah status worker');
		} finally {
			enabling = { ...enabling, [w.worker_id]: false };
		}
	}

	function setFilter(key: FilterKey): void {
		filter = key;
		writeUrlParam('status', key);
	}

	function formatRelative(iso: string | undefined | null): string {
		if (!iso) return '—';
		const then = new Date(iso).getTime();
		if (Number.isNaN(then)) return '—';
		const diffSec = Math.max(0, Math.floor((Date.now() - then) / 1000));
		if (diffSec < 5) return 'baru saja';
		if (diffSec < 60) return `${diffSec} detik lalu`;
		const diffMin = Math.floor(diffSec / 60);
		if (diffMin < 60) return `${diffMin} menit lalu`;
		const diffHour = Math.floor(diffMin / 60);
		return `${diffHour} jam lalu`;
	}

	function formatClock(iso: string | undefined | null): string {
		if (!iso) return '—';
		const d = new Date(iso);
		if (Number.isNaN(d.getTime())) return '—';
		return d.toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit', second: '2-digit' });
	}

	const visibleWorkers = $derived(
		(data?.workers ?? []).filter((w) => filter === 'all' || w.status === filter)
	);

	const capacityActive = $derived(
		(data?.workers ?? [])
			.filter((w) => w.status === 'active')
			.reduce((sum, w) => sum + (w.target_concurrency ?? 0), 0)
	);

	onMount(() => {
		void loadWorkerStatus();
		void loadCaps();
		refreshTimer = setInterval(() => {
			if (!document.hidden) {
				void loadWorkerStatus(true);
			}
		}, REFRESH_MS);
		return () => {
			if (refreshTimer) clearInterval(refreshTimer);
		};
	});
</script>

<svelte:head>
	<title>Manajemen Worker — MTSN 2 Kolut</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex flex-wrap items-end justify-between gap-3">
		<div>
			<h1 class="text-2xl font-extrabold tracking-tight text-foreground">Manajemen Worker</h1>
			<p class="mt-1 text-sm text-muted-foreground">
				Pantau status seluruh worker PUSAKA (lokal & DomCloud) secara langsung.
			</p>
		</div>
		<div class="flex items-center gap-2">
			{#if lastRefreshedAt}
				<span class="text-xs text-muted-foreground">
					Terakhir: {formatClock(lastRefreshedAt.toISOString())}
					<span class="ml-1 inline-flex items-center gap-1 text-primary">
						{#if isRefreshing}↻ memperbarui…{:else}auto 15 dtk{/if}
					</span>
				</span>
			{/if}
			<Button
				variant="outline"
				size="sm"
				onclick={() => void loadWorkerStatus()}
				disabled={loading || isRefreshing}
			>
				↻ Muat Ulang
			</Button>
		</div>
	</div>

	{#if error}
		<RecoveryPanel
			title="Gagal memuat status worker"
			message={error}
			onRetry={() => void loadWorkerStatus()}
		/>
	{/if}

	<!-- Ringkasan -->
	<div class="grid grid-cols-2 gap-3 md:grid-cols-4">
		<Card.Root class="rounded-2xl border-[1.5px] border-slate-100 p-4 shadow-sm">
			<p class="text-xs font-bold uppercase tracking-widest text-muted-foreground">Worker Aktif</p>
			{#if loading}
				<Skeleton class="mt-2 h-7 w-16" />
			{:else}
				<p class="mt-1 text-2xl font-extrabold text-emerald-600">
					{data?.active_count ?? 0}<span class="text-base font-semibold text-muted-foreground"> / {data?.total ?? 0}</span>
				</p>
			{/if}
		</Card.Root>
		<Card.Root class="rounded-2xl border-[1.5px] border-slate-100 p-4 shadow-sm">
			<p class="text-xs font-bold uppercase tracking-widest text-muted-foreground">Kapasitas Aktif</p>
			{#if loading}
				<Skeleton class="mt-2 h-7 w-16" />
			{:else}
				<p class="mt-1 text-2xl font-extrabold text-foreground">{capacityActive}</p>
				<p class="text-xs text-muted-foreground">konsumen paralel</p>
			{/if}
		</Card.Root>
		<Card.Root class="rounded-2xl border-[1.5px] border-slate-100 p-4 shadow-sm">
			<p class="text-xs font-bold uppercase tracking-widest text-muted-foreground">Antrian Job</p>
			{#if loading}
				<Skeleton class="mt-2 h-7 w-16" />
			{:else}
				<p class="mt-1 text-2xl font-extrabold text-foreground">
					{data?.queue.queued ?? 0}<span class="text-base font-semibold text-muted-foreground"> antri</span>
				</p>
				<p class="text-xs text-muted-foreground">{data?.queue.running ?? 0} sedang diproses</p>
			{/if}
		</Card.Root>
		<Card.Root class="rounded-2xl border-[1.5px] border-slate-100 p-4 shadow-sm">
			<p class="text-xs font-bold uppercase tracking-widest text-muted-foreground">Max Concurrent</p>
			{#if loading}
				<Skeleton class="mt-2 h-7 w-16" />
			{:else}
				<p class="mt-1 text-2xl font-extrabold text-foreground">{data?.global_max_concurrent || '—'}</p>
				<p class="text-xs text-muted-foreground">setting global (UI)</p>
			{/if}
		</Card.Root>
	</div>

	<!-- Filter -->
	<div class="flex flex-wrap items-center gap-1.5">
		{#each filterOptions as opt (opt.key)}
			<button
				type="button"
				onclick={() => setFilter(opt.key)}
				class="rounded-full px-3 py-1 text-xs font-bold uppercase tracking-wide transition-colors {filter === opt.key
					? 'bg-primary text-primary-foreground'
					: 'bg-surface text-muted-foreground hover:bg-slate-100'}"
			>
				{opt.label}
			</button>
		{/each}
	</div>

	<!-- Tabel Worker -->
	<Card.Root class="overflow-hidden rounded-2xl border-[1.5px] border-slate-100 shadow-sm">
		{#if loading}
			<div class="space-y-3 p-4">
				<Skeleton class="h-10 w-full" />
				<Skeleton class="h-10 w-full" />
				<Skeleton class="h-10 w-full" />
			</div>
		{:else if visibleWorkers.length === 0}
			<div class="p-6">
				<EmptyStatePanel
					title={data?.workers.length ? 'Tidak ada worker dengan status ini' : 'Belum ada worker terhubung'}
					description={data?.workers.length
						? 'Coba filter lain.'
						: 'Worker akan muncul di sini setelah heartbeat pertama (≤30 detik setelah start).'}
				/>
			</div>
		{:else}
			<div class="overflow-x-auto">
				<Table.Root>
					<Table.Header>
						<Table.Row class="hover:bg-transparent">
							<Table.Head>Worker</Table.Head>
							<Table.Head>Status</Table.Head>
							<Table.Head>Aktif</Table.Head>
							<Table.Head class="text-right">Konsumen</Table.Head>
							<Table.Head class="text-right">Target</Table.Head>
							<Table.Head>Mode</Table.Head>
							<Table.Head>Cap (Override)</Table.Head>
							<Table.Head>Sync Config</Table.Head>
							<Table.Head>Heartbeat</Table.Head>
						</Table.Row>
					</Table.Header>
					<Table.Body>
						{#each visibleWorkers as w (w.worker_id)}
							<Table.Row>
								<Table.Cell class="font-mono text-xs font-semibold">{w.worker_id}</Table.Cell>
								<Table.Cell>
									<Badge class={statusBadgeClass[w.status]}>{statusLabel[w.status]}</Badge>
								</Table.Cell>
								<Table.Cell>
									<div class="flex items-center gap-2">
										<Badge
											class={w.enabled === false
												? 'border-transparent bg-slate-200 text-slate-600'
												: 'border-transparent bg-emerald-100 text-emerald-700'}
										>
											{w.enabled === false ? 'NONAKTIF' : 'AKTIF'}
										</Badge>
										<Button
											variant="outline"
											size="sm"
											disabled={enabling[w.worker_id]}
											onclick={() => void toggleEnabled(w)}
										>
											{enabling[w.worker_id]
												? '…'
												: w.enabled === false
													? 'Aktifkan'
													: 'Nonaktifkan'}
										</Button>
									</div>
								</Table.Cell>
								<Table.Cell class="text-right font-semibold tabular-nums">{w.active_consumers ?? 0}</Table.Cell>
								<Table.Cell class="text-right tabular-nums text-muted-foreground">{w.target_concurrency ?? 0}</Table.Cell>
								<Table.Cell>
									<span class="text-xs {w.headless ? 'text-muted-foreground' : 'text-amber-600'}">
										{w.headless ? 'Headless' : 'Terbuka'}
									</span>
								</Table.Cell>
								<Table.Cell>
									<div class="flex items-center gap-1">
										<Input
											type="number"
											min="1"
											max="35"
											placeholder={w.target_concurrency ? `env ${w.target_concurrency}` : '—'}
											value={capDrafts[w.worker_id] ?? ''}
											oninput={(e: { currentTarget: HTMLInputElement }) =>
												(capDrafts = { ...capDrafts, [w.worker_id]: e.currentTarget.value })}
											class="h-8 w-16 text-xs"
										/>
										<Button
											variant="outline"
											size="sm"
											class="h-8 px-2 text-xs"
											onclick={() => void saveCap(w.worker_id)}
											disabled={capSaving[w.worker_id]}
										>
											Simpan
										</Button>
										{#if caps[w.worker_id] !== undefined}
											<Button
												variant="ghost"
												size="sm"
												class="h-8 px-2 text-xs text-rose-600"
												onclick={() => void resetCap(w.worker_id)}
												disabled={capSaving[w.worker_id]}
											>
												Reset
											</Button>
										{/if}
									</div>
									{#if caps[w.worker_id] !== undefined}
										<p class="mt-0.5 text-[10px] text-muted-foreground">
											override aktif: {caps[w.worker_id]}
										</p>
									{/if}
								</Table.Cell>
								<Table.Cell class="text-xs text-muted-foreground">{formatRelative(w.last_sync_at)}</Table.Cell>
								<Table.Cell class="text-xs text-muted-foreground">{formatRelative(w.reported_at)}</Table.Cell>
							</Table.Row>
						{/each}
					</Table.Body>
				</Table.Root>
			</div>
			<div class="border-t border-slate-100 px-4 py-2 text-xs text-muted-foreground">
				Menampilkan {visibleWorkers.length} dari {data?.total ?? 0} worker · status aktif = heartbeat ≤ 2 menit,
				stale = 2–10 menit, offline = &gt; 10 menit
			</div>
		{/if}
	</Card.Root>
</div>
