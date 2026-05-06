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

	type SnpBriefing = {
		code: string;
		label: string;
		actions: GovernanceComplianceAction[];
		openCount: number;
		criticalCount: number;
		overdueCount: number;
		waitingEvidenceCount: number;
		withoutEvidenceCount: number;
		ownerCount: number;
		nextDueDate?: string;
	};

	type SnpTotals = {
		standards: number;
		open: number;
		critical: number;
		overdue: number;
		waitingEvidence: number;
		withoutEvidence: number;
		owners: number;
	};

	const SNP_ORDER = ['', 'skl', 'isi', 'proses', 'penilaian', 'ptk', 'sarpras', 'pengelolaan', 'pembiayaan'];

	let briefingPromise = $state<Promise<GovernancePrintPackData> | null>(null);
	let snapshot = $state<GovernancePrintPackData | null>(null);
	let generatedAt = $state(new Date());
	let refreshBusy = $state(false);
	let snpFilter = $state('');

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
		return 'Briefing 8 SNP belum dapat dimuat.';
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

	function snpRank(code: string) {
		const index = SNP_ORDER.indexOf(code || '');
		return index === -1 ? SNP_ORDER.length : index;
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

	function hasEvidence(action: GovernanceComplianceAction) {
		return !!action.evidence_url.trim() || !!action.evidence_item_id || !!action.evidence_item_title;
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

	function uniqueOwnerCount(actions: GovernanceComplianceAction[]) {
		const owners: string[] = [];
		for (const action of actions) {
			const key = ownerKey(action);
			if (!owners.includes(key)) owners.push(key);
		}
		return owners.length;
	}

	function buildSnpBriefings(data: GovernancePrintPackData): SnpBriefing[] {
		const rows: SnpBriefing[] = [];
		for (const action of data.complianceActions.filter(isOpenAction)) {
			const code = action.snp_standard || '';
			let row = rows.find((item) => item.code === code);
			if (!row) {
				row = {
					code,
					label: snpLabel(code),
					actions: [],
					openCount: 0,
					criticalCount: 0,
					overdueCount: 0,
					waitingEvidenceCount: 0,
					withoutEvidenceCount: 0,
					ownerCount: 0,
					nextDueDate: undefined,
				};
				rows.push(row);
			}
			row.actions.push(action);
			row.openCount += 1;
			if (isCriticalAction(action)) row.criticalCount += 1;
			if (isOverdue(action)) row.overdueCount += 1;
			if (action.status === 'waiting_evidence') row.waitingEvidenceCount += 1;
			if (!hasEvidence(action)) row.withoutEvidenceCount += 1;
			if (numericDate(action.due_date) < numericDate(row.nextDueDate)) row.nextDueDate = action.due_date;
		}
		return rows
			.map((row) => ({ ...row, actions: sortedActions(row.actions), ownerCount: uniqueOwnerCount(row.actions) }))
			.sort(
				(a, b) =>
					b.overdueCount - a.overdueCount ||
					b.criticalCount - a.criticalCount ||
					snpRank(a.code) - snpRank(b.code) ||
					a.label.localeCompare(b.label)
			);
	}

	function filterBriefings(briefings: SnpBriefing[]) {
		if (!snpFilter) return briefings;
		if (snpFilter === '__none') return briefings.filter((briefing) => !briefing.code);
		return briefings.filter((briefing) => briefing.code === snpFilter);
	}

	function snpOptionValue(code: string) {
		return code || '__none';
	}

	function buildTotals(briefings: SnpBriefing[]): SnpTotals {
		const ownerKeys: string[] = [];
		for (const briefing of briefings) {
			for (const action of briefing.actions) {
				const key = ownerKey(action);
				if (!ownerKeys.includes(key)) ownerKeys.push(key);
			}
		}
		return {
			standards: briefings.length,
			open: briefings.reduce((sum, briefing) => sum + briefing.openCount, 0),
			critical: briefings.reduce((sum, briefing) => sum + briefing.criticalCount, 0),
			overdue: briefings.reduce((sum, briefing) => sum + briefing.overdueCount, 0),
			waitingEvidence: briefings.reduce((sum, briefing) => sum + briefing.waitingEvidenceCount, 0),
			withoutEvidence: briefings.reduce((sum, briefing) => sum + briefing.withoutEvidenceCount, 0),
			owners: ownerKeys.length,
		};
	}

	function exportCsv() {
		if (!snapshot) return;
		const briefings = filterBriefings(buildSnpBriefings(snapshot));
		const rows = [
			['Standar SNP', 'Tenggat', 'Sisa Hari', 'Judul', 'PIC', 'Unit/NIP', 'Sumber', 'Prioritas', 'Status', 'Kaitan', 'Bukti', 'Catatan'],
			...briefings.flatMap((briefing) =>
				briefing.actions.map((action) => [
					briefing.label,
					formatDate(action.due_date),
					dueLabel(action),
					action.title,
					ownerLabel(action),
					ownerUnitLabel(action),
					sourceLabel(action.source_type),
					priorityLabel(action.priority),
					actionStatusLabel(action.status),
					linkedLabel(action),
					action.evidence_url || action.evidence_item_title || '',
					action.follow_up_notes,
				])
			),
		];
		downloadCsv(`briefing-8-snp-tindak-lanjut-${new Date().toISOString().slice(0, 10)}.csv`, rows);
	}

	function printBriefing() {
		window.print();
	}

	function handleRenderError(error: unknown) {
		console.error('Governance SNP briefing render failed', error);
	}

	onMount(() => {
		void load();
	});
</script>

<svelte:head><title>Briefing 8 SNP Tindak Lanjut - MTsN 2 Kolaka Utara</title></svelte:head>

<div class="space-y-6">
	<div class="no-print flex flex-col gap-3 rounded-md border border-border bg-card px-4 py-3 lg:flex-row lg:items-center lg:justify-between">
		<div>
			<p class="text-sm font-semibold text-foreground">Briefing 8 SNP Tindak Lanjut</p>
			<p class="text-sm text-muted-foreground">Lembar pembahasan tindak lanjut menurut standar nasional pendidikan.</p>
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
				<div class="grid gap-3 md:grid-cols-3 lg:grid-cols-7">
					{#each Array.from({ length: 7 }) as _, index (`snp-briefing-stat-skeleton-${index}`)}
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
			{@const briefings = buildSnpBriefings(data)}
			{@const visibleBriefings = filterBriefings(briefings)}
			{@const totals = buildTotals(visibleBriefings)}
			<main class="print-root mx-auto max-w-6xl space-y-6 bg-card text-foreground">
				<section class="print-section rounded-md border border-border p-5">
					<div class="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
						<div>
							<p class="text-xs font-semibold uppercase tracking-[0.16em] text-primary">{data.schoolProfile.ministry_line}</p>
							<h1 class="mt-2 flex items-center gap-2 text-2xl font-bold text-foreground">
								<ClipboardCheckIcon class="size-6 text-primary" />
								Briefing 8 SNP Tindak Lanjut
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

				<section class="no-print rounded-md border border-border p-4">
					<div class="grid gap-3 lg:grid-cols-[1fr_auto] lg:items-end">
						<div>
							<label for="snp-briefing-filter" class="text-sm font-medium text-foreground">Standar SNP</label>
							<select id="snp-briefing-filter" bind:value={snpFilter} class="mt-1 w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
								<option value="">Semua standar</option>
								{#each briefings as briefing (snpOptionValue(briefing.code))}
									<option value={snpOptionValue(briefing.code)}>{briefing.label}</option>
								{/each}
							</select>
						</div>
						{#if snpFilter}
							<Button variant="outline" onclick={() => (snpFilter = '')}>Tampilkan Semua</Button>
						{/if}
					</div>
				</section>

				<section class="print-section rounded-md border border-border p-5">
					<h2 class="text-base font-semibold text-foreground">Ringkasan 8 SNP</h2>
					<div class="mt-4 grid gap-3 md:grid-cols-3 lg:grid-cols-7">
						{#each [
							{ label: 'Standar', value: totals.standards, note: 'Dengan tugas aktif' },
							{ label: 'Tugas Terbuka', value: totals.open, note: 'Belum selesai' },
							{ label: 'PIC/Unit', value: totals.owners, note: 'Terlibat' },
							{ label: 'Prioritas Tinggi', value: totals.critical, note: 'High/urgent' },
							{ label: 'Lewat Tenggat', value: totals.overdue, note: 'Butuh eskalasi' },
							{ label: 'Menunggu Bukti', value: totals.waitingEvidence, note: 'Perlu validasi' },
							{ label: 'Tanpa Bukti', value: totals.withoutEvidence, note: 'Belum ada tautan/item' },
						] as item (item.label)}
							<div class="rounded-md border border-border p-4">
								<p class="text-xs text-muted-foreground">{item.label}</p>
								<p class="mt-2 text-2xl font-semibold text-foreground">{item.value}</p>
								<p class="mt-1 text-xs text-muted-foreground">{item.note}</p>
							</div>
						{/each}
					</div>
				</section>

				{#if visibleBriefings.length === 0}
					<EmptyStatePanel title="Tidak ada briefing aktif" description="Belum ada tindak lanjut terbuka untuk standar yang dipilih." />
				{:else}
					{#each visibleBriefings as briefing (snpOptionValue(briefing.code))}
						<section class="print-section rounded-md border border-border">
							<div class="border-b border-border px-5 py-4">
								<div class="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
									<div>
										<h2 class="text-base font-semibold text-foreground">{briefing.label}</h2>
										<p class="mt-1 text-sm text-muted-foreground">{briefing.ownerCount} PIC/unit terlibat · tenggat terdekat {formatDate(briefing.nextDueDate)}</p>
									</div>
									<div class="flex flex-wrap gap-1">
										<Badge variant="secondary">{briefing.openCount} terbuka</Badge>
										{#if briefing.criticalCount > 0}<Badge variant="destructive">{briefing.criticalCount} tinggi</Badge>{/if}
										{#if briefing.overdueCount > 0}<Badge variant="destructive">{briefing.overdueCount} lewat</Badge>{/if}
										{#if briefing.waitingEvidenceCount > 0}<Badge variant="outline">{briefing.waitingEvidenceCount} tunggu bukti</Badge>{/if}
										{#if briefing.withoutEvidenceCount > 0}<Badge variant="outline">{briefing.withoutEvidenceCount} tanpa bukti</Badge>{/if}
									</div>
								</div>
							</div>
							<div class="overflow-x-auto">
								<Table.Root>
									<Table.Header>
										<Table.Row>
											<Table.Head>Tindak Lanjut</Table.Head>
											<Table.Head>PIC/Unit</Table.Head>
											<Table.Head>Status</Table.Head>
											<Table.Head>Tenggat</Table.Head>
											<Table.Head>Kaitan/Bukti</Table.Head>
											<Table.Head>Keputusan Review</Table.Head>
										</Table.Row>
									</Table.Header>
									<Table.Body>
										{#each briefing.actions as action (action.id)}
											<Table.Row>
												<Table.Cell class="min-w-64">
													<p class="text-sm font-medium text-foreground">{action.title}</p>
													<p class="text-xs text-muted-foreground">{action.period_year} - {sourceLabel(action.source_type)}</p>
													{#if action.description}
														<p class="mt-1 text-xs text-muted-foreground">{action.description}</p>
													{/if}
												</Table.Cell>
												<Table.Cell class="min-w-44">
													<p class="text-sm text-foreground">{ownerLabel(action)}</p>
													<p class="text-xs text-muted-foreground">{ownerUnitLabel(action)}</p>
												</Table.Cell>
												<Table.Cell>
													<div class="flex flex-col gap-1">
														<Badge variant={statusVariant(action.status)}>{actionStatusLabel(action.status)}</Badge>
														<Badge variant={priorityVariant(action.priority)}>{priorityLabel(action.priority)}</Badge>
													</div>
												</Table.Cell>
												<Table.Cell class="whitespace-nowrap">
													<p class={isOverdue(action) ? 'text-sm font-medium text-destructive' : 'text-sm text-foreground'}>{formatDate(action.due_date)}</p>
													<p class="text-xs text-muted-foreground">{dueLabel(action)}</p>
												</Table.Cell>
												<Table.Cell class="max-w-md text-sm text-muted-foreground">
													<p>{linkedLabel(action)}</p>
													<p class="mt-1 text-xs text-muted-foreground">{action.evidence_url || action.evidence_item_title || 'Bukti belum dicatat'}</p>
													{#if action.follow_up_notes}
														<p class="mt-1 text-xs text-muted-foreground">Catatan: {action.follow_up_notes}</p>
													{/if}
												</Table.Cell>
												<Table.Cell class="min-w-56 text-sm text-muted-foreground">........................................</Table.Cell>
											</Table.Row>
										{/each}
									</Table.Body>
								</Table.Root>
							</div>
							<div class="grid gap-4 border-t border-border p-5 md:grid-cols-3">
								<div>
									<p class="text-sm font-medium text-foreground">Kesenjangan Standar</p>
									<p class="mt-6 text-sm text-muted-foreground">........................................</p>
									<p class="mt-3 text-sm text-muted-foreground">........................................</p>
								</div>
								<div>
									<p class="text-sm font-medium text-foreground">Bukti yang Harus Dilengkapi</p>
									<p class="mt-6 text-sm text-muted-foreground">........................................</p>
									<p class="mt-3 text-sm text-muted-foreground">........................................</p>
								</div>
								<div>
									<p class="text-sm font-medium text-foreground">Keputusan Review</p>
									<p class="mt-6 text-sm text-muted-foreground">........................................</p>
									<p class="mt-3 text-sm text-muted-foreground">........................................</p>
								</div>
							</div>
						</section>
					{/each}
				{/if}
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
