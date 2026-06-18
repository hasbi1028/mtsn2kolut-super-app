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
	import {
		hasInternalAnalyticsData,
		internalAnalyticsExportFilename,
		normalizeInternalAnalyticsOverview,
		type InternalAnalyticsDailyResult,
		type InternalAnalyticsDailyItem,
		type InternalAnalyticsOverview,
		type InternalAnalyticsSummary
	} from '$lib/analytics/internal-analytics-dashboard';
	import { trackInternalAnalyticsEvent } from '$lib/analytics/internal-analytics';
	import { clientApiPathWithQuery, readClientApiData } from '$lib/client/api';

	type AnalyticsTab = 'summary' | 'visitors' | 'modules' | 'security' | 'audit';
	type AnalyticsDetail = InternalAnalyticsOverview & {
		publicDaily: InternalAnalyticsDailyResult;
		adminDaily: InternalAnalyticsDailyResult;
		securityDaily: InternalAnalyticsDailyResult;
	};

	let analyticsPromise = $state<Promise<AnalyticsDetail> | null>(null);
	let activeTab = $state<AnalyticsTab>('summary');
	let selectedGroup = $state('');
	let selectedSource = $state('');
	let selectedRole = $state('');
	let selectedResult = $state('');
	let days = $state(30);
	let exportBusy = $state(false);

	const groupOptions = [
		{ value: '', label: 'Semua' },
		{ value: 'public', label: 'Publik' },
		{ value: 'auth', label: 'Auth' },
		{ value: 'dashboard', label: 'Dashboard' },
		{ value: 'pusaka', label: 'PUSAKA' },
		{ value: 'users', label: 'Users' },
		{ value: 'rbac', label: 'RBAC' },
		{ value: 'security', label: 'Security' }
	];

	const sourceOptions = [
		{ value: '', label: 'Semua source' },
		{ value: 'public_website', label: 'Public website' },
		{ value: 'web_admin', label: 'Web Admin' },
		{ value: 'core_api', label: 'Layanan Utama' },
		{ value: 'mobile_app', label: 'Mobile App' },
		{ value: 'system', label: 'System' }
	];

	const roleOptions = [
		{ value: '', label: 'Semua peran' },
		{ value: 'admin', label: 'Admin' },
		{ value: 'guru', label: 'Guru' },
		{ value: 'staf', label: 'Staf' },
		{ value: 'kesiswaan', label: 'Kesiswaan' },
		{ value: 'siswa', label: 'Siswa' },
		{ value: 'ortu', label: 'Orang tua' }
	];

	const resultOptions = [
		{ value: '', label: 'Semua hasil' },
		{ value: 'success', label: 'Sukses' },
		{ value: 'failed', label: 'Gagal' },
		{ value: 'blocked', label: 'Diblokir' },
		{ value: 'validation_failed', label: 'Validasi gagal' },
		{ value: 'started', label: 'Dimulai' }
	];

	const tabOptions: Array<{ value: AnalyticsTab; label: string }> = [
		{ value: 'summary', label: 'Ringkasan' },
		{ value: 'visitors', label: 'Pengunjung' },
		{ value: 'modules', label: 'Aktivitas Modul' },
		{ value: 'security', label: 'Security' },
		{ value: 'audit', label: 'Audit/Export' }
	];

	function analyticsParams(limit: string, overrides: Partial<Record<'event_group' | 'source_surface' | 'role' | 'result', string>> = {}) {
		const params = new URLSearchParams({ days: String(days), limit });
		const eventGroup = overrides.event_group ?? selectedGroup;
		const source = overrides.source_surface ?? selectedSource;
		const role = overrides.role ?? selectedRole;
		const result = overrides.result ?? selectedResult;
		if (eventGroup) params.set('event_group', eventGroup);
		if (source) params.set('source_surface', source);
		if (role) params.set('role', role);
		if (result) params.set('result', result);
		return params;
	}

	function loadAnalytics() {
		const activeDays = days;
		const activeGroup = selectedGroup;
		const params = analyticsParams('60');
		analyticsPromise = Promise.all([
			fetch(clientApiPathWithQuery('/api/internal-analytics/summary', params)).then((response) => readClientApiData<InternalAnalyticsSummary>(response, 'Gagal memuat ringkasan penggunaan internal.')),
			fetch(clientApiPathWithQuery('/api/internal-analytics/daily', params)).then((response) => readClientApiData<InternalAnalyticsDailyResult>(response, 'Gagal memuat tren harian ringkasan penggunaan.')),
			fetch(clientApiPathWithQuery('/api/internal-analytics/daily', analyticsParams('60', { event_group: 'public', source_surface: '', role: '', result: '' }))).then((response) => readClientApiData<InternalAnalyticsDailyResult>(response, 'Gagal memuat agregat pengunjung.')),
			fetch(clientApiPathWithQuery('/api/internal-analytics/daily', analyticsParams('80', { event_group: '', source_surface: 'web_admin' }))).then((response) => readClientApiData<InternalAnalyticsDailyResult>(response, 'Gagal memuat aktivitas modul.')),
			fetch(clientApiPathWithQuery('/api/internal-analytics/daily', analyticsParams('60', { event_group: 'security', source_surface: '', role: '' }))).then((response) => readClientApiData<InternalAnalyticsDailyResult>(response, 'Gagal memuat sinyal security.'))
		]).then(([summary, daily, publicDaily, adminDaily, securityDaily]) => ({
			...normalizeInternalAnalyticsOverview(summary, daily, activeDays, activeGroup),
			publicDaily: normalizeInternalAnalyticsOverview(summary, publicDaily, activeDays, 'public').daily,
			adminDaily: normalizeInternalAnalyticsOverview(summary, adminDaily, activeDays, '').daily,
			securityDaily: normalizeInternalAnalyticsOverview(summary, securityDaily, activeDays, 'security').daily
		}));
	}

	function analyticsErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Ringkasan penggunaan internal belum dapat dimuat.';
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

	function sourceLabel(value: string) {
		return sourceOptions.find((option) => option.value === value)?.label ?? (value || 'Semua source');
	}

	function resultLabel(value: string) {
		return resultOptions.find((option) => option.value === value)?.label ?? (value || 'Semua hasil');
	}

	function totalCount(items: InternalAnalyticsDailyItem[]) {
		return items.reduce((sum, item) => sum + item.count, 0);
	}

	function topEvent(items: InternalAnalyticsDailyItem[]) {
		const counts = new Map<string, number>();
		for (const item of items) counts.set(item.event_name, (counts.get(item.event_name) ?? 0) + item.count);
		return [...counts.entries()].sort((a, b) => b[1] - a[1] || a[0].localeCompare(b[0]))[0] ?? null;
	}

	function reloadWithFilter() {
		void trackInternalAnalyticsEvent('security.analytics_filter', {
			pathname: window.location.pathname,
			module: 'internal_analytics',
			result: 'success'
		});
		loadAnalytics();
	}

	async function exportAggregates() {
		exportBusy = true;
		void trackInternalAnalyticsEvent('security.export_requested', {
			pathname: window.location.pathname,
			module: 'internal_analytics',
			result: 'success'
		});
		try {
			const response = await fetch(clientApiPathWithQuery('/api/internal-analytics/export', analyticsParams('10000')));
			if (!response.ok) {
				const payload = await response.json().catch(() => null) as { error?: string; message?: string } | null;
				throw new Error(payload?.error || payload?.message || 'Ekspor ringkasan penggunaan internal gagal.');
			}
			const blob = await response.blob();
			const url = URL.createObjectURL(blob);
			const anchor = document.createElement('a');
			anchor.href = url;
			anchor.download = internalAnalyticsExportFilename(response.headers.get('content-disposition'));
			document.body.appendChild(anchor);
			anchor.click();
			anchor.remove();
			URL.revokeObjectURL(url);
			toast.success('Unduhan rekap penggunaan disiapkan.');
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
		void trackInternalAnalyticsEvent('security.analytics_view', {
			pathname: window.location.pathname,
			module: 'internal_analytics',
			result: 'success'
		});
		loadAnalytics();
	});
</script>

<svelte:head>
	<title>Ringkasan Penggunaan Internal</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
		<div>
			<p class="text-sm font-medium uppercase tracking-wide text-base-content/70">Sistem</p>
			<h1 class="text-2xl font-semibold text-base-content">Ringkasan Penggunaan Internal</h1>
			<p class="mt-2 max-w-3xl text-sm text-base-content/70">
				Ringkasan agregat pemakaian modul untuk evaluasi layanan sekolah. Data mentah dan detail sensitif tidak ditampilkan di halaman ini.
			</p>
		</div>
		<div class="flex flex-col gap-3 lg:items-end">
			<LoadingButton variant="outline" onclick={() => void exportAggregates()} loading={exportBusy} loadingLabel="Export..." label="Export CSV Agregat">
				<DownloadIcon class="size-4" />
				Export CSV Agregat
			</LoadingButton>
		</div>
	</div>

	<div class="grid gap-3 rounded-lg border border-base-300 bg-base-100 p-4 md:grid-cols-5">
		<div>
			<label for="analytics-days" class="mb-1 block text-xs font-medium text-base-content/70">Periode</label>
			<select id="analytics-days" bind:value={days} onchange={reloadWithFilter} class="h-9 w-full rounded-md border border-base-300 bg-base-200 px-2 text-sm">
				<option value={7}>7 hari</option>
				<option value={14}>14 hari</option>
				<option value={30}>30 hari</option>
				<option value={90}>90 hari</option>
				<option value={180}>180 hari</option>
			</select>
		</div>
		<div>
			<label for="analytics-group" class="mb-1 block text-xs font-medium text-base-content/70">Group</label>
			<select id="analytics-group" bind:value={selectedGroup} onchange={reloadWithFilter} class="h-9 w-full rounded-md border border-base-300 bg-base-200 px-2 text-sm">
				{#each groupOptions as option (option.value)}
					<option value={option.value}>{option.label}</option>
				{/each}
			</select>
		</div>
		<div>
			<label for="analytics-source" class="mb-1 block text-xs font-medium text-base-content/70">Source</label>
			<select id="analytics-source" bind:value={selectedSource} onchange={reloadWithFilter} class="h-9 w-full rounded-md border border-base-300 bg-base-200 px-2 text-sm">
				{#each sourceOptions as option (option.value)}
					<option value={option.value}>{option.label}</option>
				{/each}
			</select>
		</div>
		<div>
			<label for="analytics-role" class="mb-1 block text-xs font-medium text-base-content/70">Peran</label>
			<select id="analytics-role" bind:value={selectedRole} onchange={reloadWithFilter} class="h-9 w-full rounded-md border border-base-300 bg-base-200 px-2 text-sm">
				{#each roleOptions as option (option.value)}
					<option value={option.value}>{option.label}</option>
				{/each}
			</select>
		</div>
		<div>
			<label for="analytics-result" class="mb-1 block text-xs font-medium text-base-content/70">Hasil</label>
			<select id="analytics-result" bind:value={selectedResult} onchange={reloadWithFilter} class="h-9 w-full rounded-md border border-base-300 bg-base-200 px-2 text-sm">
				{#each resultOptions as option (option.value)}
					<option value={option.value}>{option.label}</option>
				{/each}
			</select>
		</div>
	</div>

	<div class="flex flex-wrap gap-2" role="tablist" aria-label="Section analytics internal">
		{#each tabOptions as tab (tab.value)}
			<Button type="button" size="sm" variant={activeTab === tab.value ? 'default' : 'outline'} onclick={() => (activeTab = tab.value)} aria-pressed={activeTab === tab.value}>
				{tab.label}
			</Button>
		{/each}
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
			<RecoveryPanel title="Ringkasan Penggunaan Internal Belum Tersaji" message={analyticsErrorMessage(error)} onRetry={() => retryAnalytics(reset)} />
		{/snippet}

		{#snippet children(value)}
			{@const overview = value as AnalyticsDetail}
			{@const summary = overview.summary}
			{@const daily = overview.daily}

			{#if activeTab === 'summary'}
				<div class="grid gap-4 md:grid-cols-3">
					<Card.Root>
						<Card.Header class="pb-2">
							<Card.Description>Total Aktivitas Tercatat</Card.Description>
						</Card.Header>
						<Card.Content>
							<p class="text-3xl font-semibold text-primary">{formatNumber(summary.total_count)}</p>
							<p class="mt-1 text-xs text-base-content/70">{summary.days} hari terakhir</p>
						</Card.Content>
					</Card.Root>
					<Card.Root>
						<Card.Header class="pb-2">
							<Card.Description>Group Teratas</Card.Description>
						</Card.Header>
						<Card.Content>
							<p class="text-3xl font-semibold text-primary">{summary.groups[0] ? groupLabel(summary.groups[0].event_group) : '—'}</p>
							<p class="mt-1 text-xs text-base-content/70">{summary.groups[0] ? `${formatNumber(summary.groups[0].count)} event` : 'Belum ada agregat'}</p>
						</Card.Content>
					</Card.Root>
					<Card.Root>
						<Card.Header class="pb-2">
							<Card.Description>Filter Aktif</Card.Description>
						</Card.Header>
						<Card.Content>
							<p class="text-2xl font-semibold text-primary">{groupLabel(selectedGroup)}</p>
							<p class="mt-1 text-xs text-base-content/70">{sourceLabel(selectedSource)} · {resultLabel(selectedResult)}</p>
						</Card.Content>
					</Card.Root>
				</div>

				{#if !hasInternalAnalyticsData(overview)}
					<Card.Root class="border-dashed bg-base-200/30">
						<Card.Content class="space-y-2 pt-6">
							<p class="text-sm font-medium text-base-content">Belum ada ringkasan penggunaan internal.</p>
							<p class="text-sm text-base-content/70">
								Halaman siap digunakan, tetapi periode atau filter ini belum memiliki data agregat yang dapat ditampilkan.
							</p>
						</Card.Content>
					</Card.Root>
				{/if}

				<div class="grid gap-4 lg:grid-cols-[0.9fr,1.1fr]">
					<Card.Root>
						<Card.Header>
							<Card.Title class="text-base">Distribusi Group</Card.Title>
							<Card.Description>Jumlah aktivitas yang sudah diringkas per area modul.</Card.Description>
						</Card.Header>
						<Card.Content class="space-y-3">
							{#if summary.groups.length > 0}
								{#each summary.groups as item (item.event_group)}
									<div class="flex items-center justify-between gap-3 rounded-md border border-base-300 px-3 py-2">
										<span class="text-sm font-medium text-base-content">{groupLabel(item.event_group)}</span>
										<Badge variant="outline">{formatNumber(item.count)}</Badge>
									</div>
								{/each}
							{:else}
								<p class="rounded-md border border-dashed border-base-300 p-4 text-sm text-base-content/70">Belum ada agregat untuk periode ini.</p>
							{/if}
						</Card.Content>
					</Card.Root>

					<Card.Root>
						<Card.Header>
							<Card.Title class="text-base">Aktivitas Teratas</Card.Title>
							<Card.Description>Nama aktivitas yang tercatat dengan jumlah tertinggi.</Card.Description>
						</Card.Header>
						<Card.Content class="space-y-3">
							{#if summary.top_events.length > 0}
								{#each summary.top_events as item (item.event_name)}
									<div class="flex items-center justify-between gap-3 rounded-md border border-base-300 px-3 py-2">
										<div>
											<p class="text-sm font-medium text-base-content">{item.event_name}</p>
											<p class="text-xs text-base-content/70">{groupLabel(item.event_group)}</p>
										</div>
										<Badge variant="outline">{formatNumber(item.count)}</Badge>
									</div>
								{/each}
							{:else}
								<p class="rounded-md border border-dashed border-base-300 p-4 text-sm text-base-content/70">Belum ada aktivitas teratas untuk periode ini.</p>
							{/if}
						</Card.Content>
					</Card.Root>
				</div>
			{:else if activeTab === 'visitors'}
				{@const topPublic = topEvent(overview.publicDaily.items)}
				<div class="grid gap-4 md:grid-cols-3">
					<Card.Root><Card.Content class="pt-6"><p class="text-sm text-base-content/70">Aktivitas Pengunjung</p><p class="mt-2 text-3xl font-semibold text-primary">{formatNumber(totalCount(overview.publicDaily.items))}</p></Card.Content></Card.Root>
					<Card.Root><Card.Content class="pt-6"><p class="text-sm text-base-content/70">Aktivitas Teratas</p><p class="mt-2 text-xl font-semibold text-primary">{topPublic ? topPublic[0] : '—'}</p></Card.Content></Card.Root>
					<Card.Root><Card.Content class="pt-6"><p class="text-sm text-base-content/70">Source</p><p class="mt-2 text-xl font-semibold text-primary">Public website</p></Card.Content></Card.Root>
				</div>
			{:else if activeTab === 'modules'}
				{@const topAdmin = topEvent(overview.adminDaily.items)}
				<div class="grid gap-4 md:grid-cols-3">
					<Card.Root><Card.Content class="pt-6"><p class="text-sm text-base-content/70">Aktivitas Web Admin</p><p class="mt-2 text-3xl font-semibold text-primary">{formatNumber(totalCount(overview.adminDaily.items))}</p></Card.Content></Card.Root>
					<Card.Root><Card.Content class="pt-6"><p class="text-sm text-base-content/70">Aktivitas Teratas</p><p class="mt-2 text-xl font-semibold text-primary">{topAdmin ? topAdmin[0] : '—'}</p></Card.Content></Card.Root>
					<Card.Root><Card.Content class="pt-6"><p class="text-sm text-base-content/70">Filter Peran</p><p class="mt-2 text-xl font-semibold text-primary">{selectedRole || 'Semua'}</p></Card.Content></Card.Root>
				</div>
			{:else if activeTab === 'security'}
				{@const topSecurity = topEvent(overview.securityDaily.items)}
				<div class="grid gap-4 md:grid-cols-3">
					<Card.Root><Card.Content class="pt-6"><p class="text-sm text-base-content/70">Sinyal Security</p><p class="mt-2 text-3xl font-semibold text-primary">{formatNumber(totalCount(overview.securityDaily.items))}</p></Card.Content></Card.Root>
					<Card.Root><Card.Content class="pt-6"><p class="text-sm text-base-content/70">Sinyal Teratas</p><p class="mt-2 text-xl font-semibold text-primary">{topSecurity ? topSecurity[0] : '—'}</p></Card.Content></Card.Root>
					<Card.Root><Card.Content class="pt-6"><p class="text-sm text-base-content/70">Hasil</p><p class="mt-2 text-xl font-semibold text-primary">{resultLabel(selectedResult)}</p></Card.Content></Card.Root>
				</div>
			{:else}
				<Card.Root>
					<Card.Header>
						<Card.Title class="text-base">Riwayat dan Ekspor Agregat</Card.Title>
						<Card.Description>Ekspor memakai data agregat harian sesuai filter aktif. Permintaan ekspor ikut dicatat sebagai aktivitas keamanan agregat.</Card.Description>
					</Card.Header>
					<Card.Content class="space-y-4">
						<div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
							<Badge variant="outline">Periode {days} hari</Badge>
							<Badge variant="outline">{groupLabel(selectedGroup)}</Badge>
							<Badge variant="outline">{sourceLabel(selectedSource)}</Badge>
							<Badge variant="outline">{resultLabel(selectedResult)}</Badge>
						</div>
						<LoadingButton variant="outline" onclick={() => void exportAggregates()} loading={exportBusy} loadingLabel="Export..." label="Export CSV Agregat">
							<DownloadIcon class="size-4" />
							Export CSV Agregat
						</LoadingButton>
					</Card.Content>
				</Card.Root>
			{/if}

			<Card.Root>
				<Card.Header>
					<Card.Title class="text-base">Tren Harian</Card.Title>
					<Card.Description>
						{activeTab === 'visitors' ? 'Agregat kunjungan publik website.' : activeTab === 'modules' ? 'Agregat aktivitas Web Admin.' : activeTab === 'security' ? 'Agregat sinyal security.' : 'Agregat harian berdasarkan filter aktif.'}
					</Card.Description>
				</Card.Header>
				<Card.Content class="space-y-3">
					{@const activeItems = activeTab === 'visitors' ? overview.publicDaily.items : activeTab === 'modules' ? overview.adminDaily.items : activeTab === 'security' ? overview.securityDaily.items : daily.items}
					{#if activeItems.length > 0}
						{#each activeItems as item (`${item.aggregate_date}-${item.event_name}-${item.source_surface}-${item.role}-${item.result}`)}
							<div class="grid gap-2 rounded-md border border-base-300 px-3 py-3 md:grid-cols-[1fr,1.2fr,0.7fr,0.5fr] md:items-center">
								<div>
									<p class="text-sm font-medium text-base-content">{formatDate(item.aggregate_date)}</p>
									<p class="text-xs text-base-content/70">{groupLabel(item.event_group)}</p>
								</div>
								<div class="text-sm text-base-content">{item.event_name}</div>
								<div class="text-xs text-base-content/70">{item.source_surface}{item.role ? ` · ${item.role}` : ''}{item.result ? ` · ${item.result}` : ''}</div>
								<div class="text-right text-sm font-semibold text-primary">{formatNumber(item.count)}</div>
							</div>
						{/each}
					{:else}
						<p class="rounded-md border border-dashed border-base-300 p-4 text-sm text-base-content/70">Belum ada tren harian untuk filter ini.</p>
					{/if}
				</Card.Content>
			</Card.Root>
		{/snippet}
	</AsyncContent>
</div>
