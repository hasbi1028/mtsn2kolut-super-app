<script lang="ts">
	import { onMount } from 'svelte';
	import DownloadIcon from '@lucide/svelte/icons/download';
	import * as Card from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { toast } from '$lib/components/ui/sonner';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { clientApiPathWithQuery, readClientApiData } from '$lib/client/api';

	type SummaryGroup = {
		event_group: string;
		count: number;
	};

	type SummaryEvent = {
		event_name: string;
		event_group: string;
		count: number;
	};

	type Summary = {
		days: number;
		total_count: number;
		groups: SummaryGroup[];
		top_events: SummaryEvent[];
	};

	type DailyItem = {
		aggregate_date: string;
		event_group: string;
		event_name: string;
		source_surface: string;
		role?: string;
		result?: string;
		count: number;
	};

	type DailyResult = {
		days: number;
		event_group?: string;
		items: DailyItem[];
	};

	type AnalyticsOverview = {
		summary: Summary;
		daily: DailyResult;
	};

	let analyticsPromise = $state<Promise<AnalyticsOverview> | null>(null);
	let selectedGroup = $state('');
	let days = $state(30);
	let exportBusy = $state(false);

	const groupOptions = [
		{ value: '', label: 'Semua' },
		{ value: 'dashboard', label: 'Dashboard' },
		{ value: 'bank_soal', label: 'Bank Soal' },
		{ value: 'asesmen', label: 'Asesmen' },
		{ value: 'pusaka', label: 'PUSAKA' },
		{ value: 'users', label: 'Users' },
		{ value: 'rbac', label: 'RBAC' },
		{ value: 'security', label: 'Security' }
	];

	function analyticsParams(limit: string) {
		const params = new URLSearchParams({ days: String(days), limit });
		if (selectedGroup) params.set('event_group', selectedGroup);
		return params;
	}

	function loadAnalytics() {
		const params = analyticsParams('60');
		analyticsPromise = Promise.all([
			fetch(clientApiPathWithQuery('/api/internal-analytics/summary', params)).then((response) => readClientApiData<Summary>(response, 'Gagal memuat ringkasan analytics internal.')),
			fetch(clientApiPathWithQuery('/api/internal-analytics/daily', params)).then((response) => readClientApiData<DailyResult>(response, 'Gagal memuat tren harian analytics internal.'))
		]).then(([summary, daily]) => ({ summary, daily }));
	}

	function analyticsErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Analytics internal belum dapat dimuat.';
	}

	function retryAnalytics(reset?: () => void) {
		reset?.();
		loadAnalytics();
	}

	function formatNumber(value: number) {
		return new Intl.NumberFormat('id-ID').format(value);
	}

	function formatDate(value: string) {
		if (!value) return 'Tanggal tidak tersedia';
		const date = new Date(`${value}T00:00:00+08:00`);
		if (Number.isNaN(date.getTime())) return value;
		return date.toLocaleDateString('id-ID', { day: '2-digit', month: 'short', year: 'numeric' });
	}

	function groupLabel(value: string) {
		return groupOptions.find((option) => option.value === value)?.label ?? value;
	}

	function exportFilename(disposition: string | null) {
		const match = /filename="?([^";]+)"?/i.exec(disposition ?? '');
		return match?.[1] ?? 'internal-analytics-aggregate.csv';
	}

	async function exportAggregates() {
		exportBusy = true;
		try {
			const response = await fetch(clientApiPathWithQuery('/api/internal-analytics/export', analyticsParams('10000')));
			if (!response.ok) {
				const payload = await response.json().catch(() => null) as { error?: string; message?: string } | null;
				throw new Error(payload?.error || payload?.message || 'Export analytics internal gagal.');
			}
			const blob = await response.blob();
			const url = URL.createObjectURL(blob);
			const anchor = document.createElement('a');
			anchor.href = url;
			anchor.download = exportFilename(response.headers.get('content-disposition'));
			document.body.appendChild(anchor);
			anchor.click();
			anchor.remove();
			URL.revokeObjectURL(url);
			toast.success('Export CSV agregat disiapkan.');
		} catch (error) {
			toast.error(analyticsErrorMessage(error));
		} finally {
			exportBusy = false;
		}
	}

	function handleRenderError(error: unknown) {
		console.error('Internal analytics render failed', error);
	}

	onMount(() => {
		loadAnalytics();
	});
</script>

<svelte:head>
	<title>Analytics Internal</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
		<div>
			<p class="text-sm font-medium uppercase tracking-wide text-muted-foreground">Sistem</p>
			<h1 class="text-2xl font-semibold text-foreground">Analytics Internal</h1>
			<p class="mt-2 max-w-3xl text-sm text-muted-foreground">
				Ringkasan agregat pemakaian modul untuk evaluasi layanan sekolah. Data mentah dan detail sensitif tidak ditampilkan di halaman ini.
			</p>
		</div>
		<div class="flex flex-col gap-3 lg:items-end">
			<LoadingButton variant="outline" onclick={() => void exportAggregates()} loading={exportBusy} loadingLabel="Export..." label="Export CSV Agregat">
				<DownloadIcon class="size-4" />
				Export CSV Agregat
			</LoadingButton>
			<div class="flex flex-wrap items-center gap-2">
				{#each groupOptions as option (option.value)}
					<Button
						type="button"
						size="sm"
						variant={selectedGroup === option.value ? 'default' : 'outline'}
						onclick={() => {
							selectedGroup = option.value;
							loadAnalytics();
						}}
					>
						{option.label}
					</Button>
				{/each}
			</div>
		</div>
	</div>

	<AsyncContent promise={analyticsPromise} onerror={handleRenderError}>
		{#snippet pending()}
			<div class="space-y-4">
				<div class="grid gap-4 md:grid-cols-3">
					{#each Array.from({ length: 3 }) as _, index (`analytics-summary-skeleton-${index}`)}
						<Card.Root>
							<Card.Content class="space-y-3 pt-6">
								<Skeleton class="h-4 w-28" />
								<Skeleton class="h-8 w-20" />
							</Card.Content>
						</Card.Root>
					{/each}
				</div>
				<Card.Root><Card.Content class="space-y-3 pt-6"><Skeleton class="h-5 w-44" /><Skeleton class="h-40 w-full" /></Card.Content></Card.Root>
			</div>
		{/snippet}

		{#snippet failed(error, reset)}
			<RecoveryPanel title="Analytics Internal Belum Tersaji" message={analyticsErrorMessage(error)} onRetry={() => retryAnalytics(reset)} />
		{/snippet}

		{#snippet children(value)}
			{@const overview = value as AnalyticsOverview}
			{@const summary = overview.summary}
			{@const daily = overview.daily}

			<div class="grid gap-4 md:grid-cols-3">
				<Card.Root>
					<Card.Header class="pb-2">
						<Card.Description>Total Event Agregat</Card.Description>
					</Card.Header>
					<Card.Content>
						<p class="text-3xl font-semibold text-primary">{formatNumber(summary.total_count)}</p>
						<p class="mt-1 text-xs text-muted-foreground">{summary.days} hari terakhir</p>
					</Card.Content>
				</Card.Root>
				<Card.Root>
					<Card.Header class="pb-2">
						<Card.Description>Group Teratas</Card.Description>
					</Card.Header>
					<Card.Content>
						<p class="text-3xl font-semibold text-primary">{summary.groups[0] ? groupLabel(summary.groups[0].event_group) : '—'}</p>
						<p class="mt-1 text-xs text-muted-foreground">{summary.groups[0] ? `${formatNumber(summary.groups[0].count)} event` : 'Belum ada agregat'}</p>
					</Card.Content>
				</Card.Root>
				<Card.Root>
					<Card.Header class="pb-2">
						<Card.Description>Filter Aktif</Card.Description>
					</Card.Header>
					<Card.Content>
						<p class="text-3xl font-semibold text-primary">{groupLabel(selectedGroup)}</p>
						<p class="mt-1 text-xs text-muted-foreground">{days} hari, agregat harian</p>
					</Card.Content>
				</Card.Root>
			</div>

			<div class="grid gap-4 lg:grid-cols-[0.9fr,1.1fr]">
				<Card.Root>
					<Card.Header>
						<Card.Title class="text-base">Distribusi Group</Card.Title>
						<Card.Description>Jumlah event yang sudah diringkas per area modul.</Card.Description>
					</Card.Header>
					<Card.Content class="space-y-3">
						{#if summary.groups.length > 0}
							{#each summary.groups as item (item.event_group)}
								<div class="flex items-center justify-between gap-3 rounded-md border border-border px-3 py-2">
									<span class="text-sm font-medium text-foreground">{groupLabel(item.event_group)}</span>
									<Badge variant="outline">{formatNumber(item.count)}</Badge>
								</div>
							{/each}
						{:else}
							<p class="rounded-md border border-dashed border-border p-4 text-sm text-muted-foreground">Belum ada agregat untuk periode ini.</p>
						{/if}
					</Card.Content>
				</Card.Root>

				<Card.Root>
					<Card.Header>
						<Card.Title class="text-base">Event Teratas</Card.Title>
						<Card.Description>Nama event allowlisted dengan jumlah tertinggi.</Card.Description>
					</Card.Header>
					<Card.Content class="space-y-3">
						{#if summary.top_events.length > 0}
							{#each summary.top_events as item (item.event_name)}
								<div class="flex items-center justify-between gap-3 rounded-md border border-border px-3 py-2">
									<div>
										<p class="text-sm font-medium text-foreground">{item.event_name}</p>
										<p class="text-xs text-muted-foreground">{groupLabel(item.event_group)}</p>
									</div>
									<Badge variant="outline">{formatNumber(item.count)}</Badge>
								</div>
							{/each}
						{:else}
							<p class="rounded-md border border-dashed border-border p-4 text-sm text-muted-foreground">Belum ada event teratas untuk periode ini.</p>
						{/if}
					</Card.Content>
				</Card.Root>
			</div>

			<Card.Root>
				<Card.Header>
					<Card.Title class="text-base">Tren Harian</Card.Title>
					<Card.Description>Agregat harian berdasarkan filter group aktif.</Card.Description>
				</Card.Header>
				<Card.Content class="space-y-3">
					{#if daily.items.length > 0}
						{#each daily.items as item (`${item.aggregate_date}-${item.event_name}-${item.source_surface}-${item.role}-${item.result}`)}
							<div class="grid gap-2 rounded-md border border-border px-3 py-3 md:grid-cols-[1fr,1.2fr,0.7fr,0.5fr] md:items-center">
								<div>
									<p class="text-sm font-medium text-foreground">{formatDate(item.aggregate_date)}</p>
									<p class="text-xs text-muted-foreground">{groupLabel(item.event_group)}</p>
								</div>
								<div class="text-sm text-foreground">{item.event_name}</div>
								<div class="text-xs text-muted-foreground">{item.source_surface}{item.role ? ` · ${item.role}` : ''}{item.result ? ` · ${item.result}` : ''}</div>
								<div class="text-right text-sm font-semibold text-primary">{formatNumber(item.count)}</div>
							</div>
						{/each}
					{:else}
						<p class="rounded-md border border-dashed border-border p-4 text-sm text-muted-foreground">Belum ada tren harian untuk filter ini.</p>
					{/if}
				</Card.Content>
			</Card.Root>
		{/snippet}
	</AsyncContent>
</div>
