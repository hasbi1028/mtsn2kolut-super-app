<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Badge } from '$lib/components/ui/badge';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { toast } from '$lib/components/ui/sonner';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import { readClientJson } from '$lib/client/api';

	interface Job {
		id: string; nama: string; nip: string;
		run_type: string; status: string;
		attempts: number; max_attempts: number;
		created_at: string; updated_at: string; not_before?: string;
	}

	interface JobsResponse {
		error?: string;
		items?: Job[];
		data?: Job[];
	}

	let jobs         = $state<Job[]>([]);
	let jobsPromise  = $state<Promise<Job[]> | null>(null);
	let refreshing   = $state(false);
	let filterStatus = $state('');
	let filterType   = $state('');
	let loadedFilterKey = $state('');
	let jobsRequestId = 0;

	// Job counts per status for the stat bar
	const jobCounts = $derived({
		queued:    jobs.filter(j => j.status === 'queued').length,
		running:   jobs.filter(j => j.status === 'running').length,
		success:   jobs.filter(j => j.status === 'success').length,
		failed:    jobs.filter(j => j.status === 'failed').length,
		cancelled: jobs.filter(j => j.status === 'cancelled').length,
	});

	const statusOptions = [
		{ value: '', label: 'Semua Status' },
		{ value: 'queued',    label: 'Antri' },
		{ value: 'running',   label: 'Berjalan' },
		{ value: 'success',   label: 'Sukses' },
		{ value: 'failed',    label: 'Gagal' },
		{ value: 'cancelled', label: 'Dibatalkan' },
	];

	const typeOptions = [
		{ value: '',        label: 'Semua Tipe' },
		{ value: 'checkin', label: 'Masuk' },
		{ value: 'checkout',label: 'Pulang' },
	];

	function statusVariant(s: string): 'default' | 'destructive' | 'outline' | 'secondary' {
		if (s === 'success')   return 'default';
		if (s === 'failed')    return 'destructive';
		if (s === 'running')   return 'outline';
		return 'secondary';
	}

	function statusLabel(s: string) {
		return { queued: 'Antri', running: 'Berjalan', success: 'Sukses',
		         failed: 'Gagal', cancelled: 'Dibatalkan' }[s] ?? s;
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

	function buildJobParams() {
		const params = new URLSearchParams({ limit: '100' });
		if (filterStatus) params.set('status', filterStatus);
		if (filterType) params.set('run_type', filterType);
		return params;
	}

	function filterKey() {
		return `${filterStatus}:${filterType}`;
	}

	async function fetchJobs(): Promise<Job[]> {
		const res = await fetch(`/api/pusaka/jobs?${buildJobParams()}`);
		const payload = await readClientJson<JobsResponse>(res);
		if (payload.error) throw new Error(payload.error);
		return payload?.items ?? payload?.data ?? [];
	}

	function load() {
		const requestId = ++jobsRequestId;
		const nextFilterKey = filterKey();
		jobs = [];
		jobsPromise = fetchJobs()
			.then((nextJobs) => {
				if (requestId === jobsRequestId) {
					jobs = nextJobs;
					loadedFilterKey = nextFilterKey;
				}
				return requestId === jobsRequestId ? nextJobs : jobs;
			})
			.catch((error: unknown) => {
				if (requestId === jobsRequestId) throw error;
				return jobs;
			});
	}

	async function refreshJobs(showFailureToast = false) {
		const requestId = ++jobsRequestId;
		const nextFilterKey = filterKey();
		refreshing = true;
		try {
			const nextJobs = await fetchJobs();
			if (requestId === jobsRequestId) {
				jobs = nextJobs;
				loadedFilterKey = nextFilterKey;
				jobsPromise = Promise.resolve(nextJobs);
			}
		} catch (error) {
			if (requestId !== jobsRequestId) return;
			if (loadedFilterKey === nextFilterKey) {
				jobsPromise = Promise.resolve(jobs);
				if (showFailureToast) toast.error(jobErrorMessage(error));
			} else {
				jobsPromise = Promise.reject(error);
			}
		} finally {
			if (requestId === jobsRequestId) {
				refreshing = false;
			}
		}
	}

	function retryJobs(reset?: () => void) {
		reset?.();
		load();
	}

	function jobErrorMessage(error: unknown) {
		return error instanceof Error ? error.message : 'Gagal memuat antrian pekerjaan';
	}

	function handleJobRenderError(error: unknown) {
		console.error('Job queue render failed', error);
	}

	onMount(() => {
		load();
		const interval = setInterval(() => void refreshJobs(false), 10_000);
		return () => clearInterval(interval);
	});
</script>

<svelte:head><title>Antrian Job PUSAKA — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6">

	<div class="flex items-center gap-2 text-sm text-base-content/70">
		<a href={resolve('/pusaka')} class="hover:text-base-content">PUSAKA</a>
		<span>/</span>
		<span class="text-base-content font-medium">Antrian Job</span>
	</div>

	<div class="flex flex-wrap items-start justify-between gap-4">
		<div>
			<h1 class="text-2xl font-semibold text-base-content">Antrian Job PUSAKA</h1>
			<p class="text-sm text-base-content/70 mt-1">Auto-refresh setiap 10 detik</p>
		</div>
		<LoadingButton variant="outline" size="sm" onclick={() => void refreshJobs(true)} loading={refreshing} loadingLabel="Memuat..." label="↺ Refresh" />
	</div>

	<!-- Filters -->
	<Card.Root class="border-base-300 shadow-sm">
		<Card.Content class="pt-4 pb-3">
			<div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-[1fr_1fr_auto] xl:items-end">
				<select
					bind:value={filterStatus}
					onchange={load}
					class="h-10 rounded-lg border border-input bg-background px-3 text-sm text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
				>
					{#each statusOptions as o (o.value)}
						<option value={o.value}>{o.label}</option>
					{/each}
				</select>
				<select
					bind:value={filterType}
					onchange={load}
					class="h-10 rounded-lg border border-input bg-background px-3 text-sm text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
				>
					{#each typeOptions as o (o.value)}
						<option value={o.value}>{o.label}</option>
					{/each}
				</select>
				<div class="flex items-center gap-3 sm:col-span-2 xl:col-span-1 xl:justify-end">
					<span class="text-sm font-semibold text-base-content">{jobs.length} job</span>
				</div>
			</div>
		</Card.Content>
	</Card.Root>

	<!-- Stat bar: quick glance at job counts per status -->
	{#if jobs.length > 0}
		<div class="flex flex-wrap items-center gap-2">
			{#if jobCounts.queued > 0}
				<span class="inline-flex items-center gap-1.5 rounded-lg border border-base-300 bg-base-100 px-2.5 py-1 text-[11px] font-medium text-base-content/70">
					<span class="inline-flex h-2 w-2 rounded-full bg-info"></span>
					Antri {jobCounts.queued}
				</span>
			{/if}
			{#if jobCounts.running > 0}
				<span class="inline-flex items-center gap-1.5 rounded-lg border border-base-300 bg-base-100 px-2.5 py-1 text-[11px] font-medium text-base-content/70">
					<span class="inline-flex h-2 w-2 rounded-full bg-warning"></span>
					Berjalan {jobCounts.running}
				</span>
			{/if}
			{#if jobCounts.success > 0}
				<span class="inline-flex items-center gap-1.5 rounded-lg border border-base-300 bg-base-100 px-2.5 py-1 text-[11px] font-medium text-base-content/70">
					<span class="inline-flex h-2 w-2 rounded-full bg-success"></span>
					Sukses {jobCounts.success}
				</span>
			{/if}
			{#if jobCounts.failed > 0}
				<span class="inline-flex items-center gap-1.5 rounded-lg border border-base-300 bg-base-100 px-2.5 py-1 text-[11px] font-medium text-base-content/70">
					<span class="inline-flex h-2 w-2 rounded-full bg-destructive"></span>
					Gagal {jobCounts.failed}
				</span>
			{/if}
			{#if jobCounts.cancelled > 0}
				<span class="inline-flex items-center gap-1.5 rounded-lg border border-base-300 bg-base-100 px-2.5 py-1 text-[11px] font-medium text-base-content/70">
					<span class="inline-flex h-2 w-2 rounded-full bg-base-300"></span>
					Dibatalkan {jobCounts.cancelled}
				</span>
			{/if}
		</div>
	{/if}

	<Card.Root class="overflow-hidden border-base-300 shadow-sm">
		<Card.Content class="p-0">
			<AsyncContent promise={jobsPromise} onerror={handleJobRenderError}>
				{#snippet pending()}
					<div class="space-y-3 p-4">
						<Skeleton class="h-12 w-full" />
						<Skeleton class="h-14 w-full" />
						<Skeleton class="h-14 w-full" />
						<Skeleton class="h-14 w-full" />
					</div>
				{/snippet}

				{#snippet failed(error, reset)}
					<div class="p-4">
						<div class="rounded-2xl border border-destructive/20 bg-destructive/5 p-4">
							<p class="text-sm font-semibold text-destructive">Antrian job belum bisa dimuat</p>
							<p class="mt-1 text-sm text-base-content/70">{jobErrorMessage(error)}</p>
							<LoadingButton
								variant="outline"
								size="sm"
								onclick={() => retryJobs(reset)}
								loading={refreshing}
								loadingLabel="Mencoba..."
								label="Coba lagi"
								class="mt-4"
							/>
						</div>
					</div>
				{/snippet}

				{#snippet children(value)}
					{@const currentJobs = value as Job[]}
					<div class="hidden overflow-x-auto lg:block">
						<Table.Root>
							<Table.Header>
								<Table.Row>
									<Table.Head>Pegawai</Table.Head>
									<Table.Head>Tipe</Table.Head>
									<Table.Head>Status</Table.Head>
									<Table.Head class="text-center">Percobaan</Table.Head>
									<Table.Head class="hidden sm:table-cell">Dibuat</Table.Head>
									<Table.Head class="hidden md:table-cell">Mulai Setelah</Table.Head>
									<Table.Head class="hidden lg:table-cell">Diperbarui</Table.Head>
								</Table.Row>
							</Table.Header>
							<Table.Body>
								{#each currentJobs as job (job.id)}
									<Table.Row>
										<Table.Cell class="font-medium">{job.nama || job.nip || '—'}</Table.Cell>
										<Table.Cell>
											<Badge variant="outline">{runTypeLabel(job.run_type)}</Badge>
										</Table.Cell>
										<Table.Cell>
											<Badge variant={statusVariant(job.status)}>{statusLabel(job.status)}</Badge>
										</Table.Cell>
										<Table.Cell class="text-center text-sm">{job.attempts}/{job.max_attempts}</Table.Cell>
										<Table.Cell class="hidden sm:table-cell text-base-content/70 text-xs">{fmtDt(job.created_at)}</Table.Cell>
										<Table.Cell class="hidden md:table-cell text-base-content/70 text-xs">{fmtDt(job.not_before ?? '')}</Table.Cell>
										<Table.Cell class="hidden lg:table-cell text-base-content/70 text-xs">{fmtDt(job.updated_at)}</Table.Cell>
									</Table.Row>
								{:else}
									<Table.Row>
										<Table.Cell colspan={7} class="py-12 text-center text-base-content/70">
											Tidak ada job ditemukan.
										</Table.Cell>
									</Table.Row>
								{/each}
							</Table.Body>
						</Table.Root>
					</div>

					<div class="grid gap-3 p-4 lg:hidden">
						{#each currentJobs as job (job.id)}
							<div class="rounded-2xl border border-base-300 bg-base-100 p-4 shadow-sm">
								<div class="flex items-start justify-between gap-3">
									<div class="min-w-0">
										<p class="text-sm font-semibold text-base-content">{job.nama || job.nip || '—'}</p>
										<p class="mt-1 text-xs text-base-content/70">{fmtDt(job.created_at)}</p>
									</div>
									<Badge variant={statusVariant(job.status)}>{statusLabel(job.status)}</Badge>
								</div>
								<div class="mt-3 flex items-center gap-2">
									<Badge variant="outline">{runTypeLabel(job.run_type)}</Badge>
									<span class="text-xs text-base-content/70">Percobaan {job.attempts}/{job.max_attempts}</span>
								</div>
								<p class="mt-3 text-xs text-base-content/70">Mulai setelah {fmtDt(job.not_before ?? '')}</p>
								<p class="mt-1 text-xs text-base-content/70">Diperbarui {fmtDt(job.updated_at)}</p>
							</div>
						{:else}
							<div class="rounded-2xl border border-dashed border-base-300 bg-base-200/50 px-4 py-10 text-center text-sm text-base-content/70">
								Tidak ada job ditemukan.
							</div>
						{/each}
					</div>
				{/snippet}
			</AsyncContent>
		</Card.Content>
	</Card.Root>
</div>
