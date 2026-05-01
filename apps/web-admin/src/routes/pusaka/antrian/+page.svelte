<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Badge } from '$lib/components/ui/badge';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';

	interface Job {
		id: string; nama: string; nip: string;
		run_type: string; status: string;
		attempts: number; max_attempts: number;
		created_at: string; updated_at: string;
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

	async function fetchJobs(): Promise<Job[]> {
		const res = await fetch(`/api/pusaka/jobs?${buildJobParams()}`);
		const payload = (await res.json().catch(() => null)) as JobsResponse | null;
		if (!res.ok || payload?.error) {
			throw new Error(payload?.error || 'Gagal memuat antrian job');
		}
		return payload?.items ?? payload?.data ?? [];
	}

	function load() {
		jobs = [];
		jobsPromise = fetchJobs().then((nextJobs) => {
			jobs = nextJobs;
			return nextJobs;
		});
	}

	async function refreshJobs() {
		refreshing = true;
		try {
			const nextJobs = await fetchJobs();
			jobs = nextJobs;
			jobsPromise = Promise.resolve(nextJobs);
		} catch (error) {
			jobsPromise = Promise.reject(error);
		} finally {
			refreshing = false;
		}
	}

	function retryJobs(reset?: () => void) {
		reset?.();
		load();
	}

	function jobErrorMessage(error: unknown) {
		return error instanceof Error ? error.message : 'Gagal memuat antrian job';
	}

	function handleJobRenderError(error: unknown) {
		console.error('Job queue render failed', error);
	}

	onMount(() => {
		load();
		const interval = setInterval(refreshJobs, 10_000);
		return () => clearInterval(interval);
	});
</script>

<svelte:head><title>Antrian Job PUSAKA — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6">

	<div class="flex items-center gap-2 text-sm text-slate-500">
		<a href={resolve('/pusaka')} class="hover:text-slate-700">PUSAKA</a>
		<span>/</span>
		<span class="text-slate-700 font-medium">Antrian Job</span>
	</div>

	<div class="flex flex-wrap items-start justify-between gap-4">
		<div>
			<h1 class="text-2xl font-semibold text-slate-800">Antrian Job PUSAKA</h1>
			<p class="text-sm text-muted-foreground mt-1">Auto-refresh setiap 10 detik</p>
		</div>
		<LoadingButton variant="outline" size="sm" onclick={() => void refreshJobs()} loading={refreshing} loadingLabel="Memuat..." label="↺ Refresh" />
	</div>

	<!-- Filters -->
	<Card.Root class="border-slate-200 shadow-sm">
		<Card.Content class="pt-4 pb-3">
			<div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-[auto_1fr_1fr_auto] xl:items-end">
				<div class="text-sm text-muted-foreground">Filter</div>
				<select
					bind:value={filterStatus}
					onchange={load}
					class="h-10 rounded-md border border-input bg-background px-3 text-sm"
				>
					{#each statusOptions as o (o.value)}
						<option value={o.value}>{o.label}</option>
					{/each}
				</select>
				<select
					bind:value={filterType}
					onchange={load}
					class="h-10 rounded-md border border-input bg-background px-3 text-sm"
				>
					{#each typeOptions as o (o.value)}
						<option value={o.value}>{o.label}</option>
					{/each}
				</select>
				<div class="flex items-center justify-between gap-3 sm:col-span-2 xl:col-span-1 xl:justify-end">
					<LoadingButton variant="outline" size="sm" onclick={() => void refreshJobs()} loading={refreshing} loadingLabel="Memuat..." label="↺ Refresh" class="h-10 sm:w-auto" />
					<span class="text-sm text-muted-foreground">{jobs.length} job</span>
				</div>
			</div>
		</Card.Content>
	</Card.Root>

	<Card.Root class="overflow-hidden border-slate-200 shadow-sm">
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
							<p class="mt-1 text-sm text-muted-foreground">{jobErrorMessage(error)}</p>
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
									<Table.Head class="hidden sm:table-cell">Diperbarui</Table.Head>
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
										<Table.Cell class="hidden sm:table-cell text-muted-foreground text-xs">{fmtDt(job.created_at)}</Table.Cell>
										<Table.Cell class="hidden sm:table-cell text-muted-foreground text-xs">{fmtDt(job.updated_at)}</Table.Cell>
									</Table.Row>
								{:else}
									<Table.Row>
										<Table.Cell colspan={6} class="py-12 text-center text-muted-foreground">
											Tidak ada job ditemukan.
										</Table.Cell>
									</Table.Row>
								{/each}
							</Table.Body>
						</Table.Root>
					</div>

					<div class="grid gap-3 p-4 lg:hidden">
						{#each currentJobs as job (job.id)}
							<div class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
								<div class="flex items-start justify-between gap-3">
									<div class="min-w-0">
										<p class="text-sm font-semibold text-slate-900">{job.nama || job.nip || '—'}</p>
										<p class="mt-1 text-xs text-slate-500">{fmtDt(job.created_at)}</p>
									</div>
									<Badge variant={statusVariant(job.status)}>{statusLabel(job.status)}</Badge>
								</div>
								<div class="mt-3 flex items-center gap-2">
									<Badge variant="outline">{runTypeLabel(job.run_type)}</Badge>
									<span class="text-xs text-slate-500">Percobaan {job.attempts}/{job.max_attempts}</span>
								</div>
								<p class="mt-3 text-xs text-slate-500">Diperbarui {fmtDt(job.updated_at)}</p>
							</div>
						{:else}
							<div class="rounded-2xl border border-dashed border-slate-300 bg-slate-50 px-4 py-10 text-center text-sm text-slate-500">
								Tidak ada job ditemukan.
							</div>
						{/each}
					</div>
				{/snippet}
			</AsyncContent>
		</Card.Content>
	</Card.Root>
</div>
