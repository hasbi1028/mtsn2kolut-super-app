<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Input } from '$lib/components/ui/input';
	import { Badge } from '$lib/components/ui/badge';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { toast } from '$lib/components/ui/sonner';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { readClientApiData } from '$lib/client/api';

	type SummaryRow = {
		employee_id: string; employee_nama: string; employee_nip: string;
		total_days: number; complete_days: number;
		missing_checkout: number; missing_checkin: number;
	};

	let startDate = $state('');
	let endDate   = $state('');
	let summary   = $state<SummaryRow[]>([]);
	let summaryPromise = $state<Promise<SummaryRow[]> | null>(null);
	let refreshing = $state(false);
	let loadedRangeKey = $state('');
	let summaryRequestId = 0;

	function getDefaultDates() {
		const now = new Date();
		const start = new Date(now.getFullYear(), now.getMonth(), 1);
		const end   = new Date(now.getFullYear(), now.getMonth() + 1, 0);
		const fmt = (d: Date) => d.toISOString().split('T')[0];
		return { start: fmt(start), end: fmt(end) };
	}

	function rangeKey() {
		return `${startDate}:${endDate}`;
	}

	async function fetchSummary() {
		if (!startDate || !endDate) return [];
		const res = await fetch(`/api/pusaka/attendance/summary?start_date=${startDate}&end_date=${endDate}`);
		return readClientApiData<SummaryRow[]>(res, 'Gagal memuat ringkasan kehadiran');
	}

	function loadInitial() {
		if (!startDate || !endDate) return;
		const requestId = ++summaryRequestId;
		const nextRangeKey = rangeKey();
		summary = [];
		summaryPromise = fetchSummary()
			.then((rows) => {
				if (requestId === summaryRequestId) {
					summary = rows ?? [];
					loadedRangeKey = nextRangeKey;
					return summary;
				}
				return summary;
			})
			.catch((error: unknown) => {
				if (requestId === summaryRequestId) throw error;
				return summary;
			});
	}

	async function load() {
		if (!startDate || !endDate) return;
		if (!summaryPromise) {
			loadInitial();
			return;
		}
		const requestId = ++summaryRequestId;
		const nextRangeKey = rangeKey();
		refreshing = true;
		try {
			const rows = await fetchSummary();
			if (requestId === summaryRequestId) {
				summary = rows ?? [];
				loadedRangeKey = nextRangeKey;
				summaryPromise = Promise.resolve(summary);
			}
		} catch (error) {
			if (requestId !== summaryRequestId) return;
			if (loadedRangeKey === nextRangeKey) {
				summaryPromise = Promise.resolve(summary);
				toast.error(summaryErrorMessage(error));
			} else {
				summaryPromise = Promise.reject(error);
			}
		} finally {
			if (requestId === summaryRequestId) {
				refreshing = false;
			}
		}
	}

	function retrySummary(reset?: () => void) {
		reset?.();
		loadInitial();
	}

	function summaryErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Gagal memuat ringkasan kehadiran';
	}

	function handleSummaryRenderError(error: unknown) {
		console.error('PUSAKA attendance summary render failed', error);
	}

	function exportCSV() {
		if (summary.length === 0) return;
		const header = 'Nama,NIP,Total Hari,Lengkap,Tanpa Pulang,Tanpa Masuk';
		const rows = summary.map(r =>
			[`"${r.employee_nama}"`, r.employee_nip, r.total_days, r.complete_days, r.missing_checkout, r.missing_checkin].join(',')
		);
		const csv = [header, ...rows].join('\n');
		const blob = new Blob([csv], { type: 'text/csv' });
		const url = URL.createObjectURL(blob);
		const a = document.createElement('a');
		a.href = url;
		a.download = `ringkasan_kehadiran_${startDate}_${endDate}.csv`;
		a.click();
		URL.revokeObjectURL(url);
	}

	onMount(() => {
		const d = getDefaultDates();
		startDate = d.start;
		endDate = d.end;
		loadInitial();
	});
</script>

<svelte:head><title>Ringkasan Kehadiran PUSAKA — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6">

	<div class="flex items-center gap-2 text-sm text-base-content/70">
		<a href={resolve('/pusaka')} class="hover:text-base-content">PUSAKA</a>
		<span>/</span>
		<span class="text-base-content font-medium">Ringkasan Kehadiran</span>
	</div>

	<div class="flex flex-col gap-4 xl:flex-row xl:items-end xl:justify-between">
		<div>
			<h1 class="text-2xl font-semibold text-base-content">Ringkasan Kehadiran</h1>
			<p class="text-sm text-base-content/70 mt-1">Akumulasi kehadiran pegawai dari PUSAKA Kemenag dalam periode tertentu</p>
		</div>
		<div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-[1fr_auto_auto] xl:items-end">
			<div class="grid gap-2 sm:grid-cols-[1fr_auto_1fr] sm:items-center">
				<Input type="date" bind:value={startDate} class="h-10 min-w-0 bg-base-100" />
				<span class="text-center text-sm text-base-content/70">s/d</span>
				<Input type="date" bind:value={endDate} class="h-10 min-w-0 bg-base-100" />
			</div>
			<LoadingButton class="h-10 w-full sm:w-auto" onclick={() => void load()} loading={refreshing} loadingLabel="Memuat..." label="Tampilkan" />
			<LoadingButton class="h-10 w-full sm:w-auto" variant="outline" onclick={exportCSV} disabled={summary.length === 0} label="↓ CSV" />
		</div>
	</div>

	<Card.Root class="overflow-hidden border-base-300 shadow-sm">
		<Card.Content class="p-0">
			<AsyncContent promise={summaryPromise} onerror={handleSummaryRenderError}>
				{#snippet pending()}
				<div class="space-y-3 p-4">
					<Skeleton class="h-12 w-full" />
					<Skeleton class="h-14 w-full" />
					<Skeleton class="h-14 w-full" />
					<Skeleton class="h-14 w-full" />
					<Skeleton class="h-14 w-full" />
					<Skeleton class="h-14 w-full" />
				</div>
				{/snippet}

				{#snippet failed(error, reset)}
					<div class="p-4">
						<RecoveryPanel
							compact
							title="Ringkasan Belum Tersaji"
							message={summaryErrorMessage(error)}
							onRetry={() => retrySummary(reset)}
						/>
					</div>
				{/snippet}

				{#snippet children(value)}
					{@const currentSummary = value as SummaryRow[]}
			<div class="hidden overflow-x-auto lg:block">
			<Table.Root>
				<Table.Header>
					<Table.Row class="bg-base-200/50">
						<Table.Head>Nama Pegawai</Table.Head>
						<Table.Head>NIP</Table.Head>
						<Table.Head class="text-center">Total Hari</Table.Head>
						<Table.Head class="text-center">Lengkap</Table.Head>
						<Table.Head class="text-center text-warning">Tanpa Pulang</Table.Head>
						<Table.Head class="text-center text-destructive">Tanpa Masuk</Table.Head>
						<Table.Head class="text-center">% Kehadiran</Table.Head>
					</Table.Row>
				</Table.Header>
				<Table.Body>
					{#each currentSummary as r (r.employee_id)}
						{@const percent = r.total_days > 0 ? (r.complete_days / r.total_days) * 100 : 0}
						<Table.Row>
							<Table.Cell class="font-medium">{r.employee_nama}</Table.Cell>
							<Table.Cell class="font-mono text-xs text-base-content/70">{r.employee_nip}</Table.Cell>
							<Table.Cell class="text-center">{r.total_days}</Table.Cell>
							<Table.Cell class="text-center font-semibold text-success">{r.complete_days}</Table.Cell>
							<Table.Cell class="text-center text-warning">{r.missing_checkout}</Table.Cell>
							<Table.Cell class="text-center text-destructive">{r.missing_checkin}</Table.Cell>
							<Table.Cell class="text-center">
								<div class="flex items-center justify-center gap-2">
									<div class="w-12 h-1.5 rounded-full bg-base-200 overflow-hidden">
										<div class="h-full rounded-full bg-success" style="width:{Math.min(percent,100).toFixed(0)}%"></div>
									</div>
									<Badge variant={percent >= 80 ? 'default' : percent >= 50 ? 'outline' : 'destructive'}
										class={percent >= 80 ? 'bg-success' : ''}>
										{percent.toFixed(0)}%
									</Badge>
								</div>
							</Table.Cell>
						</Table.Row>
					{:else}
						<Table.Row>
							<Table.Cell colspan={7} class="py-12 text-center text-base-content/70">
								Pilih rentang tanggal dan klik Tampilkan.
							</Table.Cell>
						</Table.Row>
					{/each}
				</Table.Body>
			</Table.Root>
			</div>

			<div class="grid gap-3 p-4 lg:hidden">
				{#each currentSummary as r (r.employee_id)}
					{@const percent = r.total_days > 0 ? (r.complete_days / r.total_days) * 100 : 0}
					<div class="rounded-2xl border border-base-300 bg-base-100 p-4 shadow-sm">
						<div class="flex items-start justify-between gap-3">
							<div class="min-w-0">
								<p class="text-sm font-semibold text-base-content">{r.employee_nama}</p>
								<p class="mt-1 break-all font-mono text-xs text-base-content/70">{r.employee_nip}</p>
							</div>
							<Badge variant={percent >= 80 ? 'default' : percent >= 50 ? 'outline' : 'destructive'} class={percent >= 80 ? 'bg-success' : ''}>
								{percent.toFixed(0)}%
							</Badge>
						</div>
						<div class="mt-4 grid grid-cols-2 gap-3 text-sm">
							<div class="rounded-xl border border-base-300 bg-base-200/50 px-3 py-2"><span class="text-xs uppercase tracking-[0.16em] text-base-content/70">Total</span><p class="mt-1 font-semibold text-base-content">{r.total_days}</p></div>
							<div class="rounded-xl border border-base-300 bg-primary/10 px-3 py-2"><span class="text-xs uppercase tracking-[0.16em] text-primary">Lengkap</span><p class="mt-1 font-semibold text-primary">{r.complete_days}</p></div>
							<div class="rounded-xl border border-base-300 bg-warning/10 px-3 py-2"><span class="text-xs uppercase tracking-[0.16em] text-warning">Tanpa Pulang</span><p class="mt-1 font-semibold text-warning">{r.missing_checkout}</p></div>
							<div class="rounded-xl border border-base-300 bg-destructive/10 px-3 py-2"><span class="text-xs uppercase tracking-[0.16em] text-destructive">Tanpa Masuk</span><p class="mt-1 font-semibold text-destructive">{r.missing_checkin}</p></div>
						</div>
					</div>
				{:else}
					<div class="rounded-2xl border border-dashed border-base-300 bg-base-200/50 px-4 py-10 text-center text-sm text-base-content/70">
						Pilih rentang tanggal dan klik Tampilkan.
					</div>
				{/each}
			</div>
				{/snippet}
			</AsyncContent>
		</Card.Content>
	</Card.Root>
</div>
