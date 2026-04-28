<script lang="ts">
	import { onMount } from 'svelte';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';

	interface Job {
		id: string;
		nama: string;
		nip: string;
		run_type: string;
		status: string;
		attempts: number;
		max_attempts: number;
		created_at: string;
		updated_at: string;
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
			const res  = await fetch(`/api/jobs?${params}`);
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

<svelte:head><title>Antrian Job — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6">

	<div class="flex flex-wrap items-start justify-between gap-4">
		<div>
			<h1 class="text-2xl font-semibold text-slate-800">Antrian Job Absensi</h1>
			<p class="text-sm text-muted-foreground mt-1">Auto-refresh setiap 10 detik</p>
		</div>
		<div class="flex gap-2">
			<Button variant="outline" size="sm" href="/attendance">← Absensi</Button>
			<Button variant="outline" size="sm" onclick={load} disabled={loading}>
				{loading ? 'Memuat...' : '↺ Refresh'}
			</Button>
		</div>
	</div>

	<!-- Filters -->
	<Card.Root>
		<Card.Content class="pt-4 pb-3">
			<div class="flex flex-wrap gap-3 items-center">
				<span class="text-sm text-muted-foreground">Filter:</span>
				<select
					bind:value={filterStatus}
					onchange={load}
					class="h-9 rounded-md border border-input bg-background px-3 text-sm"
				>
					{#each statusOptions as o}
						<option value={o.value}>{o.label}</option>
					{/each}
				</select>
				<select
					bind:value={filterType}
					onchange={load}
					class="h-9 rounded-md border border-input bg-background px-3 text-sm"
				>
					{#each typeOptions as o}
						<option value={o.value}>{o.label}</option>
					{/each}
				</select>
				<span class="text-sm text-muted-foreground">{jobs.length} job</span>
			</div>
		</Card.Content>
	</Card.Root>

	{#if error}
		<p class="text-sm text-destructive">{error}</p>
	{/if}

	<Card.Root>
		<Card.Content class="p-0 overflow-x-auto">
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
					{#each jobs as job}
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
		</Card.Content>
	</Card.Root>
</div>
