<script lang="ts">
	import ArrowLeftIcon from '@lucide/svelte/icons/arrow-left';
	import DownloadIcon from '@lucide/svelte/icons/download';
	import PrinterIcon from '@lucide/svelte/icons/printer';
	import RefreshCcwIcon from '@lucide/svelte/icons/refresh-ccw';
	import { onMount } from 'svelte';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import * as Table from '$lib/components/ui/table';
	import {
		dateValue,
		docTypeLabel,
		downloadCsv,
		fetchGovernancePrintPackData,
		formatCurrency,
		formatDate,
		formatGeneratedAt,
		latestByDate,
		numberValue,
		snpLabel,
		statusLabel,
		type GovernanceComplianceAction,
		type GovernancePosition,
		type GovernancePrintPackData,
		type GovernanceProgram,
	} from '$lib/governance/print-pack-data';
	import { schoolAddressLine } from '$lib/school-profile';

	type BadgeVariant = 'default' | 'secondary' | 'destructive' | 'outline';
	type PositionNode = {
		position: GovernancePosition;
		level: number;
	};
	type AlignmentRow = {
		key: string;
		periodYear: number;
		programCode: string;
		programName: string;
		ikuCode: string;
		snpStandard: string;
		sourceDocument: string;
		ownerLabel: string;
		workPlanCount: number;
		workPlanDone: number;
		performanceCount: number;
		performanceDone: number;
		evidenceCount: number;
		verifiedEvidenceCount: number;
		status: string;
		progressPercent: number;
		gaps: string[];
	};
	type ComplianceActionTotals = {
		total: number;
		open: number;
		critical: number;
		overdue: number;
		done: number;
	};

	const STRATEGIC_DOCUMENT_TYPES = new Set(['visi_misi', 'renstra', 'rkjm', 'rkt', 'perkin', 'iku']);

	let packPromise = $state<Promise<GovernancePrintPackData> | null>(null);
	let snapshot = $state<GovernancePrintPackData | null>(null);
	let generatedAt = $state(new Date());
	let refreshBusy = $state(false);

	function load() {
		generatedAt = new Date();
		const promise = fetchGovernancePrintPackData();
		packPromise = promise;
		promise.then((data) => (snapshot = data)).catch(() => {});
		return promise;
	}

	async function refreshPack() {
		refreshBusy = true;
		try {
			await load();
		} catch {
			// AsyncContent owns the visible error state for the print pack body.
		} finally {
			refreshBusy = false;
		}
	}

	function retryPack(reset?: () => void) {
		reset?.();
		load();
	}

	function errorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Paket tata kelola belum dapat dimuat. Coba lagi setelah koneksi backend tersedia.';
	}

	function statusVariant(value: string): BadgeVariant {
		if (value === 'blocked' || value === 'gap' || value === 'terkendala') return 'destructive';
		if (value === 'done' || value === 'final' || value === 'verified' || value === 'lengkap') return 'default';
		if (value === 'in_progress' || value === 'collected') return 'secondary';
		return 'outline';
	}

	function flattenPositions(positions: GovernancePosition[]): PositionNode[] {
		const active = positions.filter((position) => position.is_active);
		const children: Record<string, GovernancePosition[]> = {};
		const roots: GovernancePosition[] = [];
		for (const position of active) {
			if (position.parent_position_id) {
				const existing = children[position.parent_position_id] ?? [];
				children[position.parent_position_id] = [...existing, position];
			} else {
				roots.push(position);
			}
		}
		const result: PositionNode[] = [];
		const seen: Record<string, boolean> = {};
		const sortPositions = (items: GovernancePosition[]) =>
			items.sort((a, b) => a.unit_name.localeCompare(b.unit_name) || a.sort_order - b.sort_order || a.title.localeCompare(b.title));
		const visit = (position: GovernancePosition, level: number) => {
			if (seen[position.id]) return;
			seen[position.id] = true;
			result.push({ position, level });
			for (const child of sortPositions(children[position.id] ?? [])) {
				visit(child, level + 1);
			}
		};
		for (const root of sortPositions(roots)) visit(root, 0);
		for (const remaining of sortPositions(active.filter((position) => !seen[position.id]))) visit(remaining, 0);
		return result;
	}

	function strategicDocuments(data: GovernancePrintPackData) {
		return data.documents
			.filter((document) => STRATEGIC_DOCUMENT_TYPES.has(document.doc_type))
			.sort((a, b) => b.period_year - a.period_year || a.doc_type.localeCompare(b.doc_type) || a.title.localeCompare(b.title));
	}

	function buildAlignmentRows(data: GovernancePrintPackData): AlignmentRow[] {
		const documentsByID: Record<string, { title: string }> = {};
		for (const document of data.documents) {
			documentsByID[document.id] = document;
		}
		return data.programs
			.map((program) => buildAlignmentRow(program, data, documentsByID))
			.sort((a, b) => b.periodYear - a.periodYear || alignmentRank(a.status) - alignmentRank(b.status) || a.programCode.localeCompare(b.programCode));
	}

	function buildAlignmentRow(
		program: GovernanceProgram,
		data: GovernancePrintPackData,
		documentsByID: Record<string, { title: string }>
	): AlignmentRow {
		const sourceDocument = program.source_document_id ? documentsByID[program.source_document_id] : undefined;
		const workPlans = data.workPlanItems.filter((item) => item.program_id === program.id);
		const targets = data.performanceTargets.filter((target) => target.program_id === program.id);
		const targetIDs = targets.map((target) => target.id);
		const workPlanEvidenceIDs = workPlans.map((item) => item.evidence_item_id ?? '').filter(Boolean);
		const evidence = data.evidenceItems.filter(
			(item) =>
				item.program_id === program.id ||
				workPlanEvidenceIDs.includes(item.id) ||
				(item.performance_target_id ? targetIDs.includes(item.performance_target_id) : false)
		);
		const hasIKU = program.iku_code.trim() !== '';
		const hasSource = !!sourceDocument || !!program.source_document_title;
		const hasWorkPlan = workPlans.length > 0;
		const hasTarget = targets.length > 0;
		const hasEvidence =
			program.evidence_url.trim() !== '' ||
			workPlans.some((item) => item.evidence_url.trim() !== '' || item.evidence_item_title) ||
			targets.some((target) => target.evidence_url.trim() !== '') ||
			evidence.length > 0;
		const blocked = program.status === 'blocked' || workPlans.some((item) => item.status === 'blocked') || targets.some((target) => target.status === 'blocked');
		const gaps: string[] = [];
		if (!hasIKU) gaps.push('IKU');
		if (!hasSource) gaps.push('Dokumen');
		if (!hasWorkPlan) gaps.push('RKT/RKJM');
		if (!hasTarget) gaps.push('SKP');
		if (!hasEvidence) gaps.push('Bukti');
		if (blocked) gaps.push('Terkendala');
		const progressValues = [program.progress_percent, ...workPlans.map((item) => item.progress_percent), ...targets.map((target) => target.progress_percent)];
		const progressPercent = Math.round(progressValues.reduce((total, value) => total + Number(value || 0), 0) / progressValues.length);
		const readiness = [hasIKU, hasSource, hasWorkPlan, hasTarget, hasEvidence].filter(Boolean).length;
		return {
			key: program.id,
			periodYear: program.period_year,
			programCode: program.code,
			programName: program.name,
			ikuCode: program.iku_code,
			snpStandard: program.snp_standard,
			sourceDocument: sourceDocument?.title ?? program.source_document_title ?? '',
			ownerLabel: program.responsible_employee_name || program.responsible_position_title || program.owner_unit_name || 'Belum ditetapkan',
			workPlanCount: workPlans.length,
			workPlanDone: workPlans.filter((item) => item.status === 'done').length,
			performanceCount: targets.length,
			performanceDone: targets.filter((target) => target.status === 'done').length,
			evidenceCount: evidence.length + (program.evidence_url.trim() ? 1 : 0),
			verifiedEvidenceCount: evidence.filter((item) => item.status === 'verified').length,
			status: blocked ? 'terkendala' : readiness === 5 ? 'lengkap' : readiness >= 3 ? 'parsial' : 'perlu_pemetaan',
			progressPercent,
			gaps,
		};
	}

	function alignmentRank(value: string) {
		if (value === 'terkendala') return 0;
		if (value === 'perlu_pemetaan') return 1;
		if (value === 'parsial') return 2;
		return 3;
	}

	function alignmentStatusLabel(value: string) {
		if (value === 'lengkap') return 'Lengkap';
		if (value === 'parsial') return 'Parsial';
		if (value === 'terkendala') return 'Terkendala';
		return 'Perlu Dipetakan';
	}

	function workPlanTotals(data: GovernancePrintPackData) {
		return data.workPlanItems.reduce(
			(total, item) => ({
				budget: total.budget + numberValue(item.budget_amount),
				realization: total.realization + numberValue(item.realization_amount),
			}),
			{ budget: 0, realization: 0 }
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
				return statusLabel(value);
		}
	}

	function actionSourceLabel(value: string) {
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

	function priorityRank(value: string) {
		if (value === 'urgent') return 0;
		if (value === 'high') return 1;
		if (value === 'medium') return 2;
		return 3;
	}

	function actionStatusRank(value: string) {
		if (value === 'open') return 0;
		if (value === 'in_progress') return 1;
		if (value === 'waiting_evidence') return 2;
		if (value === 'done') return 3;
		return 4;
	}

	function actionDueRank(value?: string) {
		return value ? dateValue(value) : Number.MAX_SAFE_INTEGER;
	}

	function isComplianceActionOpen(action: GovernanceComplianceAction) {
		return !['done', 'cancelled'].includes(action.status);
	}

	function isComplianceActionOverdue(action: GovernanceComplianceAction) {
		if (!action.due_date || !isComplianceActionOpen(action)) return false;
		return dateValue(action.due_date) < dateValue(new Date().toISOString().slice(0, 10));
	}

	function complianceActionLinkLabel(action: GovernanceComplianceAction) {
		const parts = [
			action.program_code || action.program_name ? `Program: ${[action.program_code, action.program_name].filter(Boolean).join(' - ')}` : '',
			action.document_title ? `Dokumen: ${action.document_title}` : '',
			action.performance_target_title ? `SKP: ${action.performance_target_title}` : '',
			action.evidence_item_title ? `Bukti: ${action.evidence_item_title}` : '',
		].filter(Boolean);
		return parts.join(' | ') || 'Belum dikaitkan';
	}

	function sortedComplianceActions(actions: GovernanceComplianceAction[]) {
		return [...actions].sort(
			(a, b) =>
				actionStatusRank(a.status) - actionStatusRank(b.status) ||
				priorityRank(a.priority) - priorityRank(b.priority) ||
				actionDueRank(a.due_date) - actionDueRank(b.due_date) ||
				dateValue(b.updated_at) - dateValue(a.updated_at)
		);
	}

	function complianceActionTotals(data: GovernancePrintPackData): ComplianceActionTotals {
		const total = data.complianceActions.length;
		const open = data.complianceActions.filter(isComplianceActionOpen).length;
		const critical = data.complianceActions.filter((action) => isComplianceActionOpen(action) && ['high', 'urgent'].includes(action.priority)).length;
		const overdue = data.complianceActions.filter(isComplianceActionOverdue).length;
		const done = data.complianceActions.filter((action) => action.status === 'done').length;
		return { total, open, critical, overdue, done };
	}

	function exportCsv() {
		if (!snapshot) return;
		const rows = [
			['Bagian', 'Kode/ID', 'Uraian', 'Tahun/Unit', 'Status', 'Catatan'],
			...flattenPositions(snapshot.positions).map((node) => [
				'Jalur Komando',
				node.position.id,
				`${'  '.repeat(node.level)}${node.position.title}`,
				node.position.unit_name,
				node.position.active_employee_name || 'Belum ada pejabat',
				node.position.tupoksi || node.position.description,
			]),
			...strategicDocuments(snapshot).map((document) => [
				'Dokumen Strategis',
				document.id,
				document.title,
				document.period_year,
				statusLabel(document.status),
				`${docTypeLabel(document.doc_type)} - ${snpLabel(document.snp_standard)} - ${document.owner_unit_name || 'Tanpa unit'}`,
			]),
			...buildAlignmentRows(snapshot).map((row) => [
				'Peta IKU',
				row.programCode,
				row.programName,
				row.periodYear,
				alignmentStatusLabel(row.status),
				`IKU ${row.ikuCode || '-'}; RKT ${row.workPlanDone}/${row.workPlanCount}; SKP ${row.performanceDone}/${row.performanceCount}; Bukti ${row.verifiedEvidenceCount}/${row.evidenceCount}; Gap ${row.gaps.join('/') || '-'}`,
			]),
			...snapshot.snpMatrix.map((row) => [
				'8 SNP',
				row.code,
				row.name,
				'',
				`${row.verified_evidence_items} valid`,
				`Dokumen ${row.total_documents}; Program ${row.total_programs}; Gap ${row.gap_evidence_items}`,
			]),
			...sortedComplianceActions(snapshot.complianceActions).map((action) => [
				'Tindak Lanjut Kepatuhan',
				action.id,
				action.title,
				action.period_year,
				`${actionStatusLabel(action.status)} / ${priorityLabel(action.priority)}`,
				`PIC ${action.responsible_employee_name || '-'}; Tenggat ${formatDate(action.due_date)}; Kaitan ${complianceActionLinkLabel(action)}; Bukti ${action.evidence_url || '-'}`,
			]),
		];
		downloadCsv(`paket-tata-kelola-${new Date().toISOString().slice(0, 10)}.csv`, rows);
	}

	function printPack() {
		window.print();
	}

	function handleRenderError(error: unknown) {
		console.error('Governance print pack render failed', error);
	}

	onMount(() => {
		void load();
	});
</script>

<svelte:head><title>Paket Cetak Tata Kelola - MTsN 2 Kolaka Utara</title></svelte:head>

<div class="space-y-6">
	<div class="no-print flex flex-col gap-3 rounded-md border border-border bg-card px-4 py-3 lg:flex-row lg:items-center lg:justify-between">
		<div>
			<p class="text-sm font-semibold text-foreground">Paket Cetak Tata Kelola</p>
			<p class="text-sm text-muted-foreground">Struktur, komando, tupoksi, dokumen strategis, IKU, RKT/RKJM, SKP, dan 8 SNP.</p>
		</div>
		<div class="flex flex-wrap gap-2">
			<Button href="/governance" variant="outline" size="sm">
				<ArrowLeftIcon class="mr-2 size-4" />
				Tata Kelola
			</Button>
			<Button variant="outline" size="sm" onclick={exportCsv}>
				<DownloadIcon class="mr-2 size-4" />
				Export CSV
			</Button>
			<LoadingButton variant="outline" size="sm" onclick={() => void refreshPack()} loading={refreshBusy} loadingLabel="Memuat...">
				<RefreshCcwIcon class="mr-2 size-4" />
				Refresh
			</LoadingButton>
			<Button size="sm" onclick={printPack}>
				<PrinterIcon class="mr-2 size-4" />
				Cetak
			</Button>
		</div>
	</div>

	<AsyncContent promise={packPromise} onerror={handleRenderError}>
		{#snippet pending()}
			<div class="space-y-4">
				<Skeleton class="h-28 w-full" />
				<div class="grid gap-3 md:grid-cols-4">
					{#each Array.from({ length: 8 }) as _, index (`governance-pack-metric-skeleton-${index}`)}
						<Skeleton class="h-24 w-full" />
					{/each}
				</div>
				<Skeleton class="h-80 w-full" />
				<Skeleton class="h-80 w-full" />
			</div>
		{/snippet}

		{#snippet failed(error, reset)}
			<RecoveryPanel message={errorMessage(error)} onRetry={() => retryPack(reset)} />
		{/snippet}

		{#snippet children(value)}
			{@const data = value as GovernancePrintPackData}
			{@const nodes = flattenPositions(data.positions)}
			{@const docs = strategicDocuments(data)}
			{@const alignment = buildAlignmentRows(data)}
			{@const totals = workPlanTotals(data)}
			{@const actionTotals = complianceActionTotals(data)}
			<main class="print-root mx-auto max-w-6xl space-y-6 bg-card text-foreground">
				<section class="print-section rounded-md border border-border p-5">
					<div class="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
						<div>
							<p class="text-xs font-semibold uppercase tracking-[0.16em] text-primary">{data.schoolProfile.ministry_line}</p>
							<h1 class="mt-2 text-2xl font-bold text-foreground">Paket Tata Kelola Madrasah</h1>
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
					<h2 class="text-base font-semibold text-foreground">Ringkasan Administrasi</h2>
					<div class="mt-4 grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
						{#each [
							{ label: 'Unit Aktif', value: data.stats.total_units, note: `${data.stats.total_positions} jabatan` },
							{ label: 'Pejabat Aktif', value: data.stats.active_assignments, note: 'Berdasarkan SK aktif' },
							{ label: 'Dokumen Final', value: data.stats.final_documents, note: `${docs.length} dokumen strategis` },
							{ label: 'Program Aktif', value: data.stats.active_programs, note: `${data.stats.blocked_programs} terkendala` },
							{ label: 'Item RKT/RKJM', value: data.stats.work_plan_items, note: `Anggaran ${formatCurrency(totals.budget)}` },
							{ label: 'Realisasi RKT', value: formatCurrency(totals.realization), note: 'Serapan tercatat' },
							{ label: 'Target SKP', value: data.stats.performance_targets, note: `${data.stats.completed_performance_targets} selesai` },
							{ label: 'Bukti 8 SNP', value: data.stats.evidence_items, note: `${data.stats.verified_evidence_items} valid, ${data.stats.gap_evidence_items} gap` },
							{ label: 'Tindak Lanjut', value: data.stats.compliance_actions ?? actionTotals.total, note: `${data.stats.open_compliance_actions ?? actionTotals.open} terbuka` },
							{ label: 'Aksi Kritis', value: data.stats.critical_compliance_actions ?? actionTotals.critical, note: `${actionTotals.overdue} lewat tenggat` },
						] as item (item.label)}
							<div class="rounded-md border border-border p-4">
								<p class="text-xs text-muted-foreground">{item.label}</p>
								<p class="mt-2 text-2xl font-semibold text-foreground">{item.value}</p>
								<p class="mt-1 text-xs text-muted-foreground">{item.note}</p>
							</div>
						{/each}
					</div>
				</section>

				<section class="print-section rounded-md border border-border">
					<div class="border-b border-border px-5 py-4">
						<h2 class="text-base font-semibold text-foreground">Struktur Organisasi dan Jalur Komando</h2>
					</div>
					<div class="overflow-x-auto">
						<Table.Root>
							<Table.Header>
								<Table.Row>
									<Table.Head>Jabatan</Table.Head>
									<Table.Head>Unit</Table.Head>
									<Table.Head>Pejabat Aktif</Table.Head>
									<Table.Head>Tupoksi Ringkas</Table.Head>
								</Table.Row>
							</Table.Header>
							<Table.Body>
								{#each nodes as node (node.position.id)}
									<Table.Row>
										<Table.Cell class="min-w-64">
											<div style={`padding-left: ${node.level * 18}px`}>
												<p class="text-sm font-medium text-foreground">{node.position.title}</p>
												<p class="text-xs text-muted-foreground">{node.position.parent_position_title ? `Atasan: ${node.position.parent_position_title}` : 'Puncak/mandiri'}</p>
											</div>
										</Table.Cell>
										<Table.Cell class="text-sm">{node.position.unit_name}</Table.Cell>
										<Table.Cell>
											<p class="text-sm">{node.position.active_employee_name || '-'}</p>
											<p class="text-xs text-muted-foreground">{node.position.active_employee_nip || ''}</p>
										</Table.Cell>
										<Table.Cell class="max-w-lg text-sm text-muted-foreground">{node.position.tupoksi || node.position.description || '-'}</Table.Cell>
									</Table.Row>
								{:else}
									<Table.Row><Table.Cell colspan={4} class="text-center text-sm text-muted-foreground">Belum ada jabatan aktif.</Table.Cell></Table.Row>
								{/each}
							</Table.Body>
						</Table.Root>
					</div>
				</section>

				<section class="print-section rounded-md border border-border">
					<div class="border-b border-border px-5 py-4">
						<h2 class="text-base font-semibold text-foreground">Pejabat Aktif dan Dasar Penugasan</h2>
					</div>
					<div class="overflow-x-auto">
						<Table.Root>
							<Table.Header>
								<Table.Row>
									<Table.Head>Jabatan</Table.Head>
									<Table.Head>Pegawai</Table.Head>
									<Table.Head>Masa Tugas</Table.Head>
									<Table.Head>SK/Catatan</Table.Head>
								</Table.Row>
							</Table.Header>
							<Table.Body>
								{#each data.assignments as assignment (assignment.id)}
									<Table.Row>
										<Table.Cell><p class="text-sm font-medium">{assignment.position_title}</p><p class="text-xs text-muted-foreground">{assignment.unit_name}</p></Table.Cell>
										<Table.Cell><p class="text-sm">{assignment.employee_name}</p><p class="text-xs text-muted-foreground">{assignment.employee_nip}</p></Table.Cell>
										<Table.Cell class="text-sm">{formatDate(assignment.start_date)} - {assignment.end_date ? formatDate(assignment.end_date) : 'Aktif'}</Table.Cell>
										<Table.Cell class="text-sm">{assignment.decree_nomor_surat || assignment.notes || '-'}</Table.Cell>
									</Table.Row>
								{:else}
									<Table.Row><Table.Cell colspan={4} class="text-center text-sm text-muted-foreground">Belum ada pejabat aktif.</Table.Cell></Table.Row>
								{/each}
							</Table.Body>
						</Table.Root>
					</div>
				</section>

				<section class="print-section rounded-md border border-border">
					<div class="border-b border-border px-5 py-4">
						<h2 class="text-base font-semibold text-foreground">Dokumen Strategis</h2>
					</div>
					<div class="overflow-x-auto">
						<Table.Root>
							<Table.Header>
								<Table.Row>
									<Table.Head>Dokumen</Table.Head>
									<Table.Head>Periode</Table.Head>
									<Table.Head>Unit</Table.Head>
									<Table.Head>Status</Table.Head>
									<Table.Head>Ringkasan</Table.Head>
								</Table.Row>
							</Table.Header>
							<Table.Body>
								{#each docs as document (document.id)}
									<Table.Row>
										<Table.Cell><p class="text-sm font-medium">{document.title}</p><p class="text-xs text-muted-foreground">{docTypeLabel(document.doc_type)} · {snpLabel(document.snp_standard)}</p></Table.Cell>
										<Table.Cell class="text-sm">{document.period_year}{document.period_label ? ` · ${document.period_label}` : ''}</Table.Cell>
										<Table.Cell class="text-sm">{document.owner_unit_name || '-'}</Table.Cell>
										<Table.Cell><Badge variant={statusVariant(document.status)}>{statusLabel(document.status)}</Badge></Table.Cell>
										<Table.Cell class="max-w-lg text-sm text-muted-foreground">{document.summary || document.document_url || '-'}</Table.Cell>
									</Table.Row>
								{:else}
									<Table.Row><Table.Cell colspan={5} class="text-center text-sm text-muted-foreground">Belum ada dokumen strategis.</Table.Cell></Table.Row>
								{/each}
							</Table.Body>
						</Table.Root>
					</div>
				</section>

				<section class="print-section rounded-md border border-border">
					<div class="border-b border-border px-5 py-4">
						<h2 class="text-base font-semibold text-foreground">Peta Renstra/IKU ke RKT, SKP, dan Bukti</h2>
					</div>
					<div class="overflow-x-auto">
						<Table.Root>
							<Table.Header>
								<Table.Row>
									<Table.Head>Program</Table.Head>
									<Table.Head>Dokumen/IKU</Table.Head>
									<Table.Head>PJ</Table.Head>
									<Table.Head>RKT</Table.Head>
									<Table.Head>SKP</Table.Head>
									<Table.Head>Bukti</Table.Head>
									<Table.Head>Status</Table.Head>
								</Table.Row>
							</Table.Header>
							<Table.Body>
								{#each alignment as row (row.key)}
									<Table.Row>
										<Table.Cell class="min-w-64">
											<p class="text-sm font-medium text-foreground">{row.programCode} · {row.programName}</p>
											<p class="text-xs text-muted-foreground">{row.periodYear} · {snpLabel(row.snpStandard)}</p>
										</Table.Cell>
										<Table.Cell class="min-w-56">
											<p class="text-sm">{row.sourceDocument || '-'}</p>
											<p class="text-xs text-muted-foreground">{row.ikuCode || 'IKU belum diisi'}</p>
										</Table.Cell>
										<Table.Cell class="text-sm">{row.ownerLabel}</Table.Cell>
										<Table.Cell class="whitespace-nowrap text-sm">{row.workPlanDone}/{row.workPlanCount}</Table.Cell>
										<Table.Cell class="whitespace-nowrap text-sm">{row.performanceDone}/{row.performanceCount}</Table.Cell>
										<Table.Cell class="whitespace-nowrap text-sm">{row.verifiedEvidenceCount}/{row.evidenceCount}</Table.Cell>
										<Table.Cell>
											<Badge variant={statusVariant(row.status)}>{alignmentStatusLabel(row.status)}</Badge>
											<p class="mt-1 text-xs text-muted-foreground">{row.gaps.length ? `Gap: ${row.gaps.join(', ')}` : `${row.progressPercent}%`}</p>
										</Table.Cell>
									</Table.Row>
								{:else}
									<Table.Row><Table.Cell colspan={7} class="text-center text-sm text-muted-foreground">Belum ada program indikator.</Table.Cell></Table.Row>
								{/each}
							</Table.Body>
						</Table.Root>
					</div>
				</section>

				<section class="print-section rounded-md border border-border">
					<div class="border-b border-border px-5 py-4">
						<h2 class="text-base font-semibold text-foreground">Register Tindak Lanjut Kepatuhan</h2>
						<p class="mt-1 text-xs text-muted-foreground">
							{actionTotals.open} terbuka, {actionTotals.critical} prioritas tinggi/mendesak, {actionTotals.overdue} lewat tenggat, {actionTotals.done} selesai.
						</p>
					</div>
					<div class="overflow-x-auto">
						<Table.Root>
							<Table.Header>
								<Table.Row>
									<Table.Head>Tindak Lanjut</Table.Head>
									<Table.Head>Kaitan</Table.Head>
									<Table.Head>PIC/Unit</Table.Head>
									<Table.Head>Status</Table.Head>
									<Table.Head>Tenggat</Table.Head>
									<Table.Head>Bukti/Catatan</Table.Head>
								</Table.Row>
							</Table.Header>
							<Table.Body>
								{#each sortedComplianceActions(data.complianceActions) as action (action.id)}
									<Table.Row>
										<Table.Cell class="min-w-64">
											<p class="text-sm font-medium text-foreground">{action.title}</p>
											<p class="text-xs text-muted-foreground">{action.period_year} · {actionSourceLabel(action.source_type)} · {snpLabel(action.snp_standard)}</p>
											{#if action.description}
												<p class="mt-1 text-xs text-muted-foreground">{action.description}</p>
											{/if}
										</Table.Cell>
										<Table.Cell class="min-w-56 text-sm text-foreground">{complianceActionLinkLabel(action)}</Table.Cell>
										<Table.Cell class="min-w-44">
											<p class="text-sm">{action.responsible_employee_name || '-'}</p>
											<p class="text-xs text-muted-foreground">{action.owner_unit_name || action.responsible_employee_nip || 'Belum ditetapkan'}</p>
										</Table.Cell>
										<Table.Cell>
											<div class="flex flex-col gap-1">
												<Badge variant={statusVariant(action.status)}>{actionStatusLabel(action.status)}</Badge>
												<Badge variant={priorityVariant(action.priority)}>{priorityLabel(action.priority)}</Badge>
											</div>
										</Table.Cell>
										<Table.Cell class={isComplianceActionOverdue(action) ? 'whitespace-nowrap text-sm font-medium text-destructive' : 'whitespace-nowrap text-sm text-foreground'}>
											{formatDate(action.due_date)}
										</Table.Cell>
										<Table.Cell class="max-w-sm text-sm text-muted-foreground">
											<p>{action.evidence_url || '-'}</p>
											{#if action.follow_up_notes}
												<p class="mt-1 text-xs text-muted-foreground">{action.follow_up_notes}</p>
											{/if}
										</Table.Cell>
									</Table.Row>
								{:else}
									<Table.Row><Table.Cell colspan={6} class="text-center text-sm text-muted-foreground">Belum ada tindak lanjut kepatuhan.</Table.Cell></Table.Row>
								{/each}
							</Table.Body>
						</Table.Root>
					</div>
				</section>

				<section class="print-section rounded-md border border-border">
					<div class="border-b border-border px-5 py-4">
						<h2 class="text-base font-semibold text-foreground">Ringkasan Bukti 8 SNP</h2>
					</div>
					<div class="overflow-x-auto">
						<Table.Root>
							<Table.Header>
								<Table.Row>
									<Table.Head>Standar</Table.Head>
									<Table.Head>Dokumen</Table.Head>
									<Table.Head>Program</Table.Head>
									<Table.Head>Selesai</Table.Head>
									<Table.Head>Bukti Program</Table.Head>
									<Table.Head>Bukti Item</Table.Head>
									<Table.Head>Valid</Table.Head>
									<Table.Head>Gap</Table.Head>
								</Table.Row>
							</Table.Header>
							<Table.Body>
								{#each data.snpMatrix as row (row.code)}
									<Table.Row>
										<Table.Cell class="font-medium">{row.name}</Table.Cell>
										<Table.Cell>{row.total_documents}</Table.Cell>
										<Table.Cell>{row.total_programs}</Table.Cell>
										<Table.Cell>{row.completed_programs}</Table.Cell>
										<Table.Cell>{row.evidence_programs}</Table.Cell>
										<Table.Cell>{row.total_evidence_items}</Table.Cell>
										<Table.Cell>{row.verified_evidence_items}</Table.Cell>
										<Table.Cell>{row.gap_evidence_items}</Table.Cell>
									</Table.Row>
								{/each}
							</Table.Body>
						</Table.Root>
					</div>
				</section>

				<section class="print-section grid gap-6 lg:grid-cols-2">
					<div class="rounded-md border border-border">
						<div class="border-b border-border px-5 py-4">
							<h2 class="text-base font-semibold text-foreground">RKT/RKJM Terbaru</h2>
						</div>
						<div class="space-y-2 p-5">
							{#each latestByDate(data.workPlanItems, (item) => item.end_date || item.start_date, 8) as item (item.id)}
								<div class="rounded-md border border-border px-3 py-2">
									<p class="text-sm font-medium text-foreground">{item.activity_code} · {item.activity_name}</p>
									<p class="text-xs text-muted-foreground">{item.program_code} · {statusLabel(item.status)} · {formatCurrency(item.realization_amount)} / {formatCurrency(item.budget_amount)}</p>
								</div>
							{:else}
								<p class="text-sm text-muted-foreground">Belum ada item RKT/RKJM.</p>
							{/each}
						</div>
					</div>
					<div class="rounded-md border border-border">
						<div class="border-b border-border px-5 py-4">
							<h2 class="text-base font-semibold text-foreground">Target SKP Terbaru</h2>
						</div>
						<div class="space-y-2 p-5">
							{#each latestByDate(data.performanceTargets, (item) => item.due_date, 8) as item (item.id)}
								<div class="rounded-md border border-border px-3 py-2">
									<p class="text-sm font-medium text-foreground">{item.title}</p>
									<p class="text-xs text-muted-foreground">{item.employee_name} · {item.program_code || 'Tanpa program'} · {statusLabel(item.status)}</p>
								</div>
							{:else}
								<p class="text-sm text-muted-foreground">Belum ada target SKP.</p>
							{/each}
						</div>
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
