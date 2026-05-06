<script lang="ts">
	import ArrowLeftIcon from '@lucide/svelte/icons/arrow-left';
	import CalendarDaysIcon from '@lucide/svelte/icons/calendar-days';
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
	type CalendarBucketKey = 'overdue' | 'today' | 'week' | 'month' | 'later' | 'no_due';
	type CalendarBucket = {
		key: CalendarBucketKey;
		label: string;
		description: string;
		actions: GovernanceComplianceAction[];
	};

	let calendarPromise = $state<Promise<GovernancePrintPackData> | null>(null);
	let snapshot = $state<GovernancePrintPackData | null>(null);
	let generatedAt = $state(new Date());
	let refreshBusy = $state(false);

	function load() {
		generatedAt = new Date();
		const promise = fetchGovernancePrintPackData();
		calendarPromise = promise;
		promise.then((data) => (snapshot = data)).catch(() => {});
		return promise;
	}

	async function refreshCalendar() {
		refreshBusy = true;
		try {
			await load();
		} catch {
			// AsyncContent owns the visible failed state.
		} finally {
			refreshBusy = false;
		}
	}

	function retryCalendar(reset?: () => void) {
		reset?.();
		load();
	}

	function errorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Kalender tenggat tindak lanjut belum dapat dimuat.';
	}

	function todayDate() {
		return new Date(new Date().toISOString().slice(0, 10));
	}

	function addDays(date: Date, days: number) {
		return new Date(date.getTime() + days * 24 * 60 * 60 * 1000);
	}

	function dayDiff(value?: string) {
		if (!value) return Number.MAX_SAFE_INTEGER;
		const today = todayDate().getTime();
		const due = new Date(value.slice(0, 10)).getTime();
		if (!Number.isFinite(due)) return Number.MAX_SAFE_INTEGER;
		return Math.ceil((due - today) / (24 * 60 * 60 * 1000));
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
		if (value === 'done') return 'default';
		if (value === 'cancelled') return 'outline';
		if (value === 'waiting_evidence') return 'secondary';
		return 'destructive';
	}

	function priorityRank(value: string) {
		if (value === 'urgent') return 0;
		if (value === 'high') return 1;
		if (value === 'medium') return 2;
		return 3;
	}

	function dueRank(value?: string) {
		if (!value) return Number.MAX_SAFE_INTEGER;
		const time = new Date(value).getTime();
		return Number.isFinite(time) ? time : Number.MAX_SAFE_INTEGER;
	}

	function sortedActions(actions: GovernanceComplianceAction[]) {
		return [...actions].sort(
			(a, b) =>
				dueRank(a.due_date) - dueRank(b.due_date) ||
				priorityRank(a.priority) - priorityRank(b.priority) ||
				a.title.localeCompare(b.title)
		);
	}

	function bucketActions(data: GovernancePrintPackData): CalendarBucket[] {
		const openActions = data.complianceActions.filter(isOpenAction);
		const today = todayDate();
		const weekEnd = addDays(today, 7);
		const monthEnd = addDays(today, 30);
		const buckets: CalendarBucket[] = [
			{ key: 'overdue', label: 'Lewat Tenggat', description: 'Butuh eskalasi atau penetapan ulang tenggat.', actions: [] },
			{ key: 'today', label: 'Hari Ini', description: 'Perlu diputuskan atau dicek hari ini.', actions: [] },
			{ key: 'week', label: '7 Hari ke Depan', description: 'Target kerja pekanan untuk PIC/unit.', actions: [] },
			{ key: 'month', label: '30 Hari ke Depan', description: 'Rencana tindak lanjut bulan berjalan.', actions: [] },
			{ key: 'later', label: 'Setelah 30 Hari', description: 'Agenda pemantauan berikutnya.', actions: [] },
			{ key: 'no_due', label: 'Tanpa Tenggat', description: 'Perlu tanggal agar jalur komando jelas.', actions: [] },
		];

		for (const action of openActions) {
			if (!action.due_date) {
				buckets.find((bucket) => bucket.key === 'no_due')?.actions.push(action);
				continue;
			}
			const due = new Date(action.due_date.slice(0, 10));
			if (due.getTime() < today.getTime()) buckets.find((bucket) => bucket.key === 'overdue')?.actions.push(action);
			else if (due.getTime() === today.getTime()) buckets.find((bucket) => bucket.key === 'today')?.actions.push(action);
			else if (due.getTime() <= weekEnd.getTime()) buckets.find((bucket) => bucket.key === 'week')?.actions.push(action);
			else if (due.getTime() <= monthEnd.getTime()) buckets.find((bucket) => bucket.key === 'month')?.actions.push(action);
			else buckets.find((bucket) => bucket.key === 'later')?.actions.push(action);
		}

		return buckets.map((bucket) => ({ ...bucket, actions: sortedActions(bucket.actions) }));
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

	function ownerLabel(action: GovernanceComplianceAction) {
		return action.responsible_employee_name || action.owner_unit_name || 'Belum ada PIC';
	}

	function ownerUnitLabel(action: GovernanceComplianceAction) {
		if (action.owner_unit_name) return action.owner_unit_name;
		if (action.responsible_employee_nip) return action.responsible_employee_nip;
		return 'Unit belum dicatat';
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

	function dueLabel(action: GovernanceComplianceAction) {
		if (!action.due_date) return 'Tanpa tenggat';
		const diff = dayDiff(action.due_date);
		if (diff < 0) return `${Math.abs(diff)} hari lewat`;
		if (diff === 0) return 'Hari ini';
		if (diff === 1) return 'Besok';
		return `${diff} hari lagi`;
	}

	function bucketVariant(key: CalendarBucketKey): BadgeVariant {
		if (key === 'overdue') return 'destructive';
		if (key === 'today' || key === 'week') return 'secondary';
		return 'outline';
	}

	function exportCsv() {
		if (!snapshot) return;
		const buckets = bucketActions(snapshot);
		const rows = [
			['Bucket', 'Tenggat', 'Sisa Hari', 'Judul', 'PIC', 'Unit', 'Prioritas', 'Status', 'SNP', 'Kaitan', 'Bukti'],
			...buckets.flatMap((bucket) =>
				bucket.actions.map((action) => [
					bucket.label,
					formatDate(action.due_date),
					dueLabel(action),
					action.title,
					ownerLabel(action),
					ownerUnitLabel(action),
					priorityLabel(action.priority),
					actionStatusLabel(action.status),
					snpLabel(action.snp_standard),
					linkedLabel(action),
					action.evidence_url || action.evidence_item_title || '',
				])
			),
		];
		downloadCsv(`kalender-tenggat-tindak-lanjut-${new Date().toISOString().slice(0, 10)}.csv`, rows);
	}

	function printCalendar() {
		window.print();
	}

	function handleRenderError(error: unknown) {
		console.error('Governance action calendar render failed', error);
	}

	onMount(() => {
		void load();
	});
</script>

<svelte:head><title>Kalender Tindak Lanjut - MTsN 2 Kolaka Utara</title></svelte:head>

<div class="space-y-6">
	<div class="no-print flex flex-col gap-3 rounded-md border border-border bg-card px-4 py-3 lg:flex-row lg:items-center lg:justify-between">
		<div>
			<p class="text-sm font-semibold text-foreground">Kalender Tindak Lanjut</p>
			<p class="text-sm text-muted-foreground">Timeline tenggat untuk gap Renstra/IKU/RKT/SKP/8 SNP.</p>
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
			<LoadingButton variant="outline" size="sm" loading={refreshBusy} loadingLabel="Memuat..." onclick={() => void refreshCalendar()}>
				<RefreshCcwIcon class="mr-2 size-4" />
				Refresh
			</LoadingButton>
			<Button size="sm" onclick={printCalendar}>
				<PrinterIcon class="mr-2 size-4" />
				Cetak
			</Button>
		</div>
	</div>

	<AsyncContent promise={calendarPromise} onerror={handleRenderError}>
		{#snippet pending()}
			<div class="space-y-4">
				<Skeleton class="h-28 w-full" />
				<div class="grid gap-3 md:grid-cols-3 xl:grid-cols-6">
					{#each Array.from({ length: 6 }) as _, index (`calendar-bucket-skeleton-${index}`)}
						<Skeleton class="h-24 w-full" />
					{/each}
				</div>
				<Skeleton class="h-96 w-full" />
			</div>
		{/snippet}

		{#snippet failed(error, reset)}
			<RecoveryPanel message={errorMessage(error)} onRetry={() => retryCalendar(reset)} />
		{/snippet}

		{#snippet children(value)}
			{@const data = value as GovernancePrintPackData}
			{@const buckets = bucketActions(data)}
			{@const openActions = data.complianceActions.filter(isOpenAction)}
			{@const criticalActions = openActions.filter(isCriticalAction)}
			{@const overdueActions = openActions.filter(isOverdue)}
			<main class="print-root mx-auto max-w-6xl space-y-6 bg-card text-foreground">
				<section class="print-section rounded-md border border-border p-5">
					<div class="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
						<div>
							<p class="text-xs font-semibold uppercase tracking-[0.16em] text-primary">{data.schoolProfile.ministry_line}</p>
							<h1 class="mt-2 flex items-center gap-2 text-2xl font-bold text-foreground">
								<CalendarDaysIcon class="size-6 text-primary" />
								Kalender Tindak Lanjut Kepatuhan
							</h1>
							<p class="mt-1 text-sm text-muted-foreground">{data.schoolProfile.name}</p>
							<p class="mt-1 max-w-2xl text-xs text-muted-foreground">{schoolAddressLine(data.schoolProfile) || data.schoolProfile.office_line}</p>
						</div>
						<div class="rounded-md border border-border px-4 py-3 text-sm">
							<p class="font-medium text-foreground">Waktu cetak</p>
							<p class="text-muted-foreground">{formatGeneratedAt(generatedAt)} WITA</p>
						</div>
					</div>
				</section>

				<section class="print-section rounded-md border border-border p-5">
					<h2 class="text-base font-semibold text-foreground">Ringkasan Timeline</h2>
					<div class="mt-4 grid gap-3 md:grid-cols-3 xl:grid-cols-6">
						{#each [
							{ label: 'Terbuka', value: openActions.length, note: 'Belum selesai' },
							{ label: 'Prioritas Tinggi', value: criticalActions.length, note: 'High/urgent' },
							{ label: 'Lewat Tenggat', value: overdueActions.length, note: 'Butuh eskalasi' },
							...buckets.slice(1, 4).map((bucket) => ({ label: bucket.label, value: bucket.actions.length, note: bucket.description })),
						] as item (item.label)}
							<div class="rounded-md border border-border p-4">
								<p class="text-xs text-muted-foreground">{item.label}</p>
								<p class="mt-2 text-2xl font-semibold text-foreground">{item.value}</p>
								<p class="mt-1 text-xs text-muted-foreground">{item.note}</p>
							</div>
						{/each}
					</div>
				</section>

				<section class="print-section grid gap-4 lg:grid-cols-2">
					{#each buckets as bucket (bucket.key)}
						<div class="rounded-md border border-border">
							<div class="border-b border-border px-5 py-4">
								<div class="flex items-start justify-between gap-3">
									<div>
										<h2 class="text-base font-semibold text-foreground">{bucket.label}</h2>
										<p class="mt-1 text-xs text-muted-foreground">{bucket.description}</p>
									</div>
									<Badge variant={bucketVariant(bucket.key)}>{bucket.actions.length}</Badge>
								</div>
							</div>
							<div class="space-y-2 p-5">
								{#each bucket.actions.slice(0, 8) as action (action.id)}
									<div class="rounded-md border border-border px-3 py-2">
										<div class="flex items-start justify-between gap-3">
											<div>
												<p class="text-sm font-medium text-foreground">{action.title}</p>
												<p class="mt-1 text-xs text-muted-foreground">{ownerLabel(action)} - {snpLabel(action.snp_standard)}</p>
											</div>
											<div class="flex flex-col items-end gap-1">
												<Badge variant={priorityVariant(action.priority)}>{priorityLabel(action.priority)}</Badge>
												<Badge variant={statusVariant(action.status)}>{actionStatusLabel(action.status)}</Badge>
											</div>
										</div>
										<p class={isOverdue(action) ? 'mt-2 text-xs font-medium text-destructive' : 'mt-2 text-xs text-muted-foreground'}>
											{formatDate(action.due_date)} - {dueLabel(action)}
										</p>
									</div>
								{:else}
									<EmptyStatePanel compact title="Tidak ada item" description="Tidak ada tindak lanjut pada bucket ini." />
								{/each}
								{#if bucket.actions.length > 8}
									<p class="text-xs text-muted-foreground">+{bucket.actions.length - 8} item lain di export CSV.</p>
								{/if}
							</div>
						</div>
					{/each}
				</section>

				<section class="print-section rounded-md border border-border">
					<div class="border-b border-border px-5 py-4">
						<h2 class="text-base font-semibold text-foreground">Timeline Detail</h2>
					</div>
					<div class="overflow-x-auto">
						<Table.Root>
							<Table.Header>
								<Table.Row>
									<Table.Head>Tenggat</Table.Head>
									<Table.Head>Tindak Lanjut</Table.Head>
									<Table.Head>PIC/Unit</Table.Head>
									<Table.Head>Status</Table.Head>
									<Table.Head>Kaitan</Table.Head>
								</Table.Row>
							</Table.Header>
							<Table.Body>
								{#each sortedActions(openActions).slice(0, 40) as action (action.id)}
									<Table.Row>
										<Table.Cell class="whitespace-nowrap">
											<p class={isOverdue(action) ? 'text-sm font-medium text-destructive' : 'text-sm text-foreground'}>{formatDate(action.due_date)}</p>
											<p class="text-xs text-muted-foreground">{dueLabel(action)}</p>
										</Table.Cell>
										<Table.Cell class="min-w-64">
											<p class="text-sm font-medium text-foreground">{action.title}</p>
											<p class="text-xs text-muted-foreground">{sourceLabel(action.source_type)} - {snpLabel(action.snp_standard)}</p>
										</Table.Cell>
										<Table.Cell class="min-w-44">
											<p class="text-sm">{ownerLabel(action)}</p>
											<p class="text-xs text-muted-foreground">{ownerUnitLabel(action)}</p>
										</Table.Cell>
										<Table.Cell>
											<div class="flex flex-col gap-1">
												<Badge variant={statusVariant(action.status)}>{actionStatusLabel(action.status)}</Badge>
												<Badge variant={priorityVariant(action.priority)}>{priorityLabel(action.priority)}</Badge>
											</div>
										</Table.Cell>
										<Table.Cell class="max-w-md text-sm text-muted-foreground">
											<p>{linkedLabel(action)}</p>
											<p class="mt-1 text-xs text-muted-foreground">{action.evidence_url || action.evidence_item_title || 'Bukti belum dicatat'}</p>
										</Table.Cell>
									</Table.Row>
								{:else}
									<Table.Row><Table.Cell colspan={5} class="text-center text-sm text-muted-foreground">Tidak ada tindak lanjut aktif.</Table.Cell></Table.Row>
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
