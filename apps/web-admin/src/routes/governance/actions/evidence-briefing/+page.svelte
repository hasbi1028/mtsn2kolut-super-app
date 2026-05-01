<script lang="ts">
	import ArrowLeftIcon from '@lucide/svelte/icons/arrow-left';
	import ClipboardCheckIcon from '@lucide/svelte/icons/clipboard-check';
	import DownloadIcon from '@lucide/svelte/icons/download';
	import PrinterIcon from '@lucide/svelte/icons/printer';
	import RefreshCcwIcon from '@lucide/svelte/icons/refresh-ccw';
	import { resolve } from '$app/paths';
	import { onMount } from 'svelte';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import EmptyStatePanel from '$lib/components/EmptyStatePanel.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import * as Table from '$lib/components/ui/table';
	import {
		dateValue,
		downloadCsv,
		fetchGovernancePrintPackData,
		formatDate,
		formatGeneratedAt,
		snpLabel,
		type GovernanceComplianceAction,
		type GovernancePrintPackData,
	} from '$lib/governance/print-pack-data';
	import { schoolAddressLine } from '$lib/school-profile';

	type BadgeVariant = 'default' | 'secondary' | 'destructive' | 'outline';
	type EvidenceBucketKey = 'waiting' | 'missing_overdue' | 'missing_critical' | 'missing' | 'recorded';

	type EvidenceBucket = {
		key: EvidenceBucketKey;
		label: string;
		description: string;
		actions: GovernanceComplianceAction[];
	};

	type EvidenceTotals = {
		open: number;
		waiting: number;
		missing: number;
		missingOverdue: number;
		missingCritical: number;
		recorded: number;
	};

	let briefingPromise = $state<Promise<GovernancePrintPackData> | null>(null);
	let snapshot = $state<GovernancePrintPackData | null>(null);
	let generatedAt = $state(new Date());
	let refreshBusy = $state(false);
	let bucketFilter = $state('');

	function load() {
		generatedAt = new Date();
		const promise = fetchGovernancePrintPackData();
		briefingPromise = promise;
		promise.then((data) => (snapshot = data)).catch(() => {});
		return promise;
	}

	async function refreshBriefing() {
		refreshBusy = true;
		try {
			await load();
		} catch {
			// AsyncContent owns the visible failed state.
		} finally {
			refreshBusy = false;
		}
	}

	function retryBriefing(reset?: () => void) {
		reset?.();
		load();
	}

	function errorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Briefing bukti tindak lanjut belum dapat dimuat.';
	}

	function isOpenAction(action: GovernanceComplianceAction) {
		return !['done', 'cancelled'].includes(action.status);
	}

	function isCriticalAction(action: GovernanceComplianceAction) {
		return ['high', 'urgent'].includes(action.priority);
	}

	function isOverdue(action: GovernanceComplianceAction) {
		return !!action.due_date && isOpenAction(action) && dateValue(action.due_date) < dateValue(new Date().toISOString().slice(0, 10));
	}

	function hasEvidence(action: GovernanceComplianceAction) {
		return !!action.evidence_url.trim() || !!action.evidence_item_id || !!action.evidence_item_title;
	}

	function numericDate(value?: string) {
		if (!value) return Number.MAX_SAFE_INTEGER;
		const time = new Date(value).getTime();
		return Number.isFinite(time) ? time : Number.MAX_SAFE_INTEGER;
	}

	function priorityRank(value: string) {
		if (value === 'urgent') return 0;
		if (value === 'high') return 1;
		if (value === 'medium') return 2;
		return 3;
	}

	function sortedActions(actions: GovernanceComplianceAction[]) {
		return [...actions].sort(
			(a, b) =>
				Number(isOverdue(b)) - Number(isOverdue(a)) ||
				priorityRank(a.priority) - priorityRank(b.priority) ||
				numericDate(a.due_date) - numericDate(b.due_date) ||
				a.title.localeCompare(b.title)
		);
	}

	function actionStatusLabel(value: string) {
		switch (value) {
			case 'open':
				return 'Terbuka';
			case 'in_progress':
				return 'Dikerjakan';
			case 'waiting_evidence':
				return 'Menunggu Bukti';
			case 'done':
				return 'Selesai';
			case 'cancelled':
				return 'Dibatalkan';
			default:
				return value || '-';
		}
	}

	function priorityLabel(value: string) {
		switch (value) {
			case 'urgent':
				return 'Mendesak';
			case 'high':
				return 'Tinggi';
			case 'medium':
				return 'Sedang';
			case 'low':
				return 'Rendah';
			default:
				return value || '-';
		}
	}

	function priorityVariant(value: string): BadgeVariant {
		if (value === 'urgent' || value === 'high') return 'destructive';
		if (value === 'medium') return 'secondary';
		return 'outline';
	}

	function statusVariant(value: string): BadgeVariant {
		if (value === 'waiting_evidence') return 'secondary';
		if (value === 'done') return 'default';
		if (value === 'cancelled') return 'outline';
		return 'destructive';
	}

	function bucketVariant(key: EvidenceBucketKey): BadgeVariant {
		if (key === 'missing_overdue' || key === 'missing_critical') return 'destructive';
		if (key === 'waiting') return 'secondary';
		return 'outline';
	}

	function sourceLabel(value: string) {
		switch (value) {
			case 'alignment_gap':
				return 'Gap Peta IKU';
			case 'snp_gap':
				return 'Gap 8 SNP';
			case 'audit':
				return 'Review/Audit';
			case 'document':
				return 'Dokumen';
			case 'program':
				return 'Program';
			case 'performance':
				return 'SKP/Kinerja';
			case 'evidence':
				return 'Bukti Mutu';
			case 'manual':
				return 'Manual';
			default:
				return value || '-';
		}
	}

	function ownerKey(action: GovernanceComplianceAction) {
		if (action.responsible_employee_id) return `employee:${action.responsible_employee_id}`;
		if (action.owner_unit_id) return `unit:${action.owner_unit_id}`;
		return 'unassigned';
	}

	function ownerLabel(action: GovernanceComplianceAction) {
		return action.responsible_employee_name || action.owner_unit_name || 'Belum ada PIC';
	}

	function ownerUnitLabel(action: GovernanceComplianceAction) {
		if (action.owner_unit_name) return action.owner_unit_name;
		if (action.responsible_employee_nip) return action.responsible_employee_nip;
		return ownerKey(action) === 'unassigned' ? 'Perlu penugasan' : 'Unit belum dicatat';
	}

	function linkedLabel(action: GovernanceComplianceAction) {
		const parts = [
			action.program_code || action.program_name ? `Program: ${[action.program_code, action.program_name].filter(Boolean).join(' - ')}` : '',
			action.document_title ? `Dokumen: ${action.document_title}` : '',
			action.performance_target_title ? `SKP: ${action.performance_target_title}` : '',
			action.evidence_item_title ? `Bukti: ${action.evidence_item_title}` : '',
		].filter(Boolean);
		return parts.join(' | ') || 'Belum dikaitkan';
	}

	function evidenceLabel(action: GovernanceComplianceAction) {
		if (action.evidence_url.trim()) return action.evidence_url;
		if (action.evidence_item_title) return action.evidence_item_title;
		return 'Belum ada bukti';
	}

	function dueLabel(action: GovernanceComplianceAction) {
		if (!action.due_date) return 'Tanpa tenggat';
		const today = dateValue(new Date().toISOString().slice(0, 10));
		const diff = Math.ceil((dateValue(action.due_date) - today) / (24 * 60 * 60 * 1000));
		if (diff < 0) return `${Math.abs(diff)} hari lewat`;
		if (diff === 0) return 'Hari ini';
		if (diff === 1) return 'Besok';
		return `${diff} hari lagi`;
	}

	function bucketActions(data: GovernancePrintPackData): EvidenceBucket[] {
		const buckets: EvidenceBucket[] = [
			{ key: 'waiting', label: 'Menunggu Bukti', description: 'Status sudah menunggu bukti dan perlu validasi lokasi/item bukti.', actions: [] },
			{ key: 'missing_overdue', label: 'Lewat Tenggat Tanpa Bukti', description: 'Action terbuka sudah lewat tenggat dan belum punya bukti.', actions: [] },
			{ key: 'missing_critical', label: 'Prioritas Tinggi Tanpa Bukti', description: 'High/urgent yang belum memiliki URL atau item bukti.', actions: [] },
			{ key: 'missing', label: 'Tanpa Bukti', description: 'Action terbuka lain yang belum memiliki bukti.', actions: [] },
			{ key: 'recorded', label: 'Bukti Tercatat', description: 'Action terbuka yang sudah punya tautan atau item bukti.', actions: [] },
		];

		for (const action of data.complianceActions.filter(isOpenAction)) {
			const evidenceReady = hasEvidence(action);
			if (action.status === 'waiting_evidence') {
				buckets.find((bucket) => bucket.key === 'waiting')?.actions.push(action);
			} else if (!evidenceReady && isOverdue(action)) {
				buckets.find((bucket) => bucket.key === 'missing_overdue')?.actions.push(action);
			} else if (!evidenceReady && isCriticalAction(action)) {
				buckets.find((bucket) => bucket.key === 'missing_critical')?.actions.push(action);
			} else if (!evidenceReady) {
				buckets.find((bucket) => bucket.key === 'missing')?.actions.push(action);
			} else {
				buckets.find((bucket) => bucket.key === 'recorded')?.actions.push(action);
			}
		}

		return buckets.map((bucket) => ({ ...bucket, actions: sortedActions(bucket.actions) }));
	}

	function filterBuckets(buckets: EvidenceBucket[]) {
		if (!bucketFilter) return buckets;
		return buckets.filter((bucket) => bucket.key === bucketFilter);
	}

	function buildTotals(data: GovernancePrintPackData): EvidenceTotals {
		const openActions = data.complianceActions.filter(isOpenAction);
		return {
			open: openActions.length,
			waiting: openActions.filter((action) => action.status === 'waiting_evidence').length,
			missing: openActions.filter((action) => !hasEvidence(action)).length,
			missingOverdue: openActions.filter((action) => !hasEvidence(action) && isOverdue(action)).length,
			missingCritical: openActions.filter((action) => !hasEvidence(action) && isCriticalAction(action)).length,
			recorded: openActions.filter(hasEvidence).length,
		};
	}

	function exportCsv() {
		if (!snapshot) return;
		const buckets = filterBuckets(bucketActions(snapshot));
		const rows = [
			['Bucket', 'Tenggat', 'Sisa Hari', 'Judul', 'PIC', 'Unit/NIP', 'Sumber', 'SNP', 'Prioritas', 'Status', 'Kaitan', 'Bukti', 'Catatan'],
			...buckets.flatMap((bucket) =>
				bucket.actions.map((action) => [
					bucket.label,
					formatDate(action.due_date),
					dueLabel(action),
					action.title,
					ownerLabel(action),
					ownerUnitLabel(action),
					sourceLabel(action.source_type),
					snpLabel(action.snp_standard),
					priorityLabel(action.priority),
					actionStatusLabel(action.status),
					linkedLabel(action),
					evidenceLabel(action),
					action.follow_up_notes,
				])
			),
		];
		downloadCsv(`briefing-bukti-tindak-lanjut-${new Date().toISOString().slice(0, 10)}.csv`, rows);
	}

	function printBriefing() {
		window.print();
	}

	function handleRenderError(error: unknown) {
		console.error('Governance evidence briefing render failed', error);
	}

	onMount(() => {
		void load();
	});
</script>

<svelte:head><title>Briefing Bukti Tindak Lanjut - MTsN 2 Kolaka Utara</title></svelte:head>

<div class="space-y-6">
	<div class="no-print flex flex-col gap-3 rounded-md border border-slate-200 bg-white px-4 py-3 lg:flex-row lg:items-center lg:justify-between">
		<div>
			<p class="text-sm font-semibold text-slate-900">Briefing Bukti Tindak Lanjut</p>
			<p class="text-sm text-slate-500">Lembar kerja untuk menutup kekurangan bukti pada action kepatuhan.</p>
		</div>
		<div class="flex flex-wrap gap-2">
			<Button href={resolve('/governance/actions')} variant="outline" size="sm">
				<ArrowLeftIcon class="mr-2 size-4" />
				Tindak Lanjut
			</Button>
			<Button variant="outline" size="sm" onclick={exportCsv}>
				<DownloadIcon class="mr-2 size-4" />
				Export CSV
			</Button>
			<LoadingButton variant="outline" size="sm" loading={refreshBusy} loadingLabel="Memuat..." onclick={() => void refreshBriefing()}>
				<RefreshCcwIcon class="mr-2 size-4" />
				Refresh
			</LoadingButton>
			<Button size="sm" onclick={printBriefing}>
				<PrinterIcon class="mr-2 size-4" />
				Cetak
			</Button>
		</div>
	</div>

	<AsyncContent promise={briefingPromise} onerror={handleRenderError}>
		{#snippet pending()}
			<div class="space-y-4">
				<Skeleton class="h-28 w-full" />
				<div class="grid gap-3 md:grid-cols-3 lg:grid-cols-6">
					{#each Array.from({ length: 6 }) as _, index (`evidence-briefing-stat-skeleton-${index}`)}
						<Skeleton class="h-24 w-full" />
					{/each}
				</div>
				<Skeleton class="h-96 w-full" />
			</div>
		{/snippet}

		{#snippet failed(error, reset)}
			<RecoveryPanel message={errorMessage(error)} onRetry={() => retryBriefing(reset)} />
		{/snippet}

		{#snippet children(value)}
			{@const data = value as GovernancePrintPackData}
			{@const buckets = bucketActions(data)}
			{@const visibleBuckets = filterBuckets(buckets)}
			{@const totals = buildTotals(data)}
			<main class="print-root mx-auto max-w-6xl space-y-6 bg-white text-slate-950">
				<section class="print-section rounded-md border border-slate-200 p-5">
					<div class="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
						<div>
							<p class="text-xs font-semibold uppercase tracking-[0.16em] text-emerald-700">{data.schoolProfile.ministry_line}</p>
							<h1 class="mt-2 flex items-center gap-2 text-2xl font-bold text-slate-950">
								<ClipboardCheckIcon class="size-6 text-emerald-700" />
								Briefing Bukti Tindak Lanjut
							</h1>
							<p class="mt-1 text-sm text-slate-600">{data.schoolProfile.name}</p>
							<p class="mt-1 max-w-2xl text-xs text-slate-500">{schoolAddressLine(data.schoolProfile) || data.schoolProfile.office_line}</p>
						</div>
						<div class="rounded-md border border-slate-200 px-4 py-3 text-sm">
							<p class="font-medium text-slate-900">Waktu cetak</p>
							<p class="text-slate-600">{formatGeneratedAt(generatedAt)} WITA</p>
						</div>
					</div>
				</section>

				<section class="no-print rounded-md border border-slate-200 p-4">
					<div class="grid gap-3 lg:grid-cols-[1fr_auto] lg:items-end">
						<div>
							<label for="evidence-bucket-filter" class="text-sm font-medium text-slate-700">Fokus bukti</label>
							<select id="evidence-bucket-filter" bind:value={bucketFilter} class="mt-1 w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
								<option value="">Semua fokus</option>
								{#each buckets as bucket (bucket.key)}
									<option value={bucket.key}>{bucket.label}</option>
								{/each}
							</select>
						</div>
						{#if bucketFilter}
							<Button variant="outline" onclick={() => (bucketFilter = '')}>Tampilkan Semua</Button>
						{/if}
					</div>
				</section>

				<section class="print-section rounded-md border border-slate-200 p-5">
					<h2 class="text-base font-semibold text-slate-950">Ringkasan Bukti</h2>
					<div class="mt-4 grid gap-3 md:grid-cols-3 lg:grid-cols-6">
						{#each [
							{ label: 'Action Terbuka', value: totals.open, note: 'Belum selesai' },
							{ label: 'Menunggu Bukti', value: totals.waiting, note: 'Perlu validasi' },
							{ label: 'Tanpa Bukti', value: totals.missing, note: 'Belum ada tautan/item' },
							{ label: 'Lewat Tanpa Bukti', value: totals.missingOverdue, note: 'Prioritas review' },
							{ label: 'Kritis Tanpa Bukti', value: totals.missingCritical, note: 'High/urgent' },
							{ label: 'Bukti Tercatat', value: totals.recorded, note: 'Masih terbuka' },
						] as item (item.label)}
							<div class="rounded-md border border-slate-200 p-4">
								<p class="text-xs text-slate-500">{item.label}</p>
								<p class="mt-2 text-2xl font-semibold text-slate-950">{item.value}</p>
								<p class="mt-1 text-xs text-slate-500">{item.note}</p>
							</div>
						{/each}
					</div>
				</section>

				<section class="print-section grid gap-4 lg:grid-cols-2">
					{#each visibleBuckets as bucket (bucket.key)}
						<div class="rounded-md border border-slate-200">
							<div class="border-b border-slate-200 px-5 py-4">
								<div class="flex items-start justify-between gap-3">
									<div>
										<h2 class="text-base font-semibold text-slate-950">{bucket.label}</h2>
										<p class="mt-1 text-xs text-slate-500">{bucket.description}</p>
									</div>
									<Badge variant={bucketVariant(bucket.key)}>{bucket.actions.length}</Badge>
								</div>
							</div>
							<div class="space-y-2 p-5">
								{#each bucket.actions.slice(0, 8) as action (action.id)}
									<div class="rounded-md border border-slate-200 px-3 py-2">
										<div class="flex items-start justify-between gap-3">
											<div>
												<p class="text-sm font-medium text-slate-900">{action.title}</p>
												<p class="mt-1 text-xs text-slate-500">{ownerLabel(action)} - {snpLabel(action.snp_standard)}</p>
											</div>
											<div class="flex flex-col items-end gap-1">
												<Badge variant={priorityVariant(action.priority)}>{priorityLabel(action.priority)}</Badge>
												<Badge variant={statusVariant(action.status)}>{actionStatusLabel(action.status)}</Badge>
											</div>
										</div>
										<p class={isOverdue(action) ? 'mt-2 text-xs font-medium text-red-700' : 'mt-2 text-xs text-slate-500'}>
											{formatDate(action.due_date)} - {dueLabel(action)}
										</p>
										<p class="mt-1 text-xs text-slate-500">{evidenceLabel(action)}</p>
									</div>
								{:else}
									<EmptyStatePanel compact title="Tidak ada item" description="Tidak ada tindak lanjut pada fokus bukti ini." />
								{/each}
								{#if bucket.actions.length > 8}
									<p class="text-xs text-slate-500">+{bucket.actions.length - 8} item lain di tabel detail/export CSV.</p>
								{/if}
							</div>
						</div>
					{/each}
				</section>

				<section class="print-section rounded-md border border-slate-200">
					<div class="border-b border-slate-200 px-5 py-4">
						<h2 class="text-base font-semibold text-slate-950">Daftar Validasi Bukti</h2>
					</div>
					<div class="overflow-x-auto">
						<Table.Root>
							<Table.Header>
								<Table.Row>
									<Table.Head>Fokus</Table.Head>
									<Table.Head>Tindak Lanjut</Table.Head>
									<Table.Head>PIC/Unit</Table.Head>
									<Table.Head>Status</Table.Head>
									<Table.Head>Bukti/Kaitan</Table.Head>
									<Table.Head>Validasi</Table.Head>
								</Table.Row>
							</Table.Header>
							<Table.Body>
								{#each visibleBuckets.flatMap((bucket) => bucket.actions.map((action) => ({ bucket, action }))).slice(0, 50) as row (`${row.bucket.key}-${row.action.id}`)}
									<Table.Row>
										<Table.Cell class="min-w-40">
											<Badge variant={bucketVariant(row.bucket.key)}>{row.bucket.label}</Badge>
										</Table.Cell>
										<Table.Cell class="min-w-64">
											<p class="text-sm font-medium text-slate-900">{row.action.title}</p>
											<p class="text-xs text-slate-500">{row.action.period_year} - {sourceLabel(row.action.source_type)} - {snpLabel(row.action.snp_standard)}</p>
										</Table.Cell>
										<Table.Cell class="min-w-44">
											<p class="text-sm text-slate-900">{ownerLabel(row.action)}</p>
											<p class="text-xs text-slate-500">{ownerUnitLabel(row.action)}</p>
										</Table.Cell>
										<Table.Cell>
											<div class="flex flex-col gap-1">
												<Badge variant={statusVariant(row.action.status)}>{actionStatusLabel(row.action.status)}</Badge>
												<Badge variant={priorityVariant(row.action.priority)}>{priorityLabel(row.action.priority)}</Badge>
											</div>
										</Table.Cell>
										<Table.Cell class="max-w-md text-sm text-slate-600">
											<p>{evidenceLabel(row.action)}</p>
											<p class="mt-1 text-xs text-slate-500">{linkedLabel(row.action)}</p>
											<p class={isOverdue(row.action) ? 'mt-1 text-xs font-medium text-red-700' : 'mt-1 text-xs text-slate-500'}>
												{formatDate(row.action.due_date)} - {dueLabel(row.action)}
											</p>
										</Table.Cell>
										<Table.Cell class="min-w-56 text-sm text-slate-400">........................................</Table.Cell>
									</Table.Row>
								{:else}
									<Table.Row><Table.Cell colspan={6} class="text-center text-sm text-slate-500">Tidak ada tindak lanjut aktif untuk divalidasi.</Table.Cell></Table.Row>
								{/each}
							</Table.Body>
						</Table.Root>
					</div>
				</section>
			</main>
		{/snippet}
	</AsyncContent>
</div>

<style>
	@media print {
		:global(body) {
			background: white;
		}

		:global(body *) {
			visibility: hidden;
		}

		.no-print {
			display: none !important;
		}

		.print-root,
		.print-root * {
			visibility: visible;
		}

		.print-root {
			position: absolute;
			left: 0;
			top: 0;
			width: 100%;
			max-width: none;
			margin: 0;
			padding: 0;
		}

		.print-section {
			break-inside: avoid;
			box-shadow: none !important;
		}
	}
</style>
