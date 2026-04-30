<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';

	interface Job {
		id: string; nama: string; nip: string;
		run_type: string; status: string;
		attempts: number; max_attempts: number;
		created_at: string; updated_at: string;
	}

	let jobs         = $state<Job[]>([]);
	let loading      = $state(false);
	let error        = $state('');
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

	async function load() {
		loading = true;
		error = '';
		try {
			const params = new URLSearchParams({ limit: '100' });
			if (filterStatus) params.set('status', filterStatus);
			if (filterType)   params.set('run_type', filterType);
			const res  = await fetch(`/api/pusaka/jobs?${params}`);
			const data = await res.json();
			if (data.error) { error = data.error; return; }
			jobs = data.items ?? data.data ?? [];
		} catch {
			error = 'Gagal memuat antrian job';
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		load();
		const interval = setInterval(load, 10_000);
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
		<Button variant="outline" size="sm" onclick={load} disabled={loading}>
			{loading ? 'Memuat...' : '↺ Refresh'}
		</Button>
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
					{#each statusOptions as o}
						<option value={o.value}>{o.label}</option>
					{/each}
				</select>
				<select
					bind:value={filterType}
					onchange={load}
					class="h-10 rounded-md border border-input bg-background px-3 text-sm"
				>
					{#each typeOptions as o}
						<option value={o.value}>{o.label}</option>
					{/each}
				</select>
				<div class="flex items-center justify-between gap-3 sm:col-span-2 xl:col-span-1 xl:justify-end">
					<Button variant="outline" size="sm" onclick={load} disabled={loading} class="h-10 sm:w-auto">
						{loading ? 'Memuat...' : '↺ Refresh'}
					</Button>
					<span class="text-sm text-muted-foreground">{jobs.length} job</span>
				</div>
			</div>
		</Card.Content>
	</Card.Root>

	{#if error}
		<p class="text-sm text-destructive">{error}</p>
	{/if}

	<Card.Root class="overflow-hidden border-slate-200 shadow-sm">
		<Card.Content class="p-0">
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
					{#each jobs as job (job.id)}
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
								{loading ? 'Memuat...' : 'Tidak ada job ditemukan.'}
							</Table.Cell>
						</Table.Row>
					{/each}
				</Table.Body>
			</Table.Root>
			</div>

			<div class="grid gap-3 p-4 lg:hidden">
				{#each jobs as job (job.id)}
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
						{loading ? 'Memuat...' : 'Tidak ada job ditemukan.'}
					</div>
				{/each}
			</div>
		</Card.Content>
	</Card.Root>
</div>
