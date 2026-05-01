<script lang="ts">
	import ArrowLeftIcon from '@lucide/svelte/icons/arrow-left';
	import DownloadIcon from '@lucide/svelte/icons/download';
	import PrinterIcon from '@lucide/svelte/icons/printer';
	import RefreshCcwIcon from '@lucide/svelte/icons/refresh-ccw';
	import { resolve } from '$app/paths';
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

	type ActionTotals = {
		total: number;
		open: number;
		critical: number;
		overdue: number;
		waitingEvidence: number;
		unassigned: number;
		done: number;
	};

	type OwnerRow = {
		key: string;
		label: string;
		unitLabel: string;
		openCount: number;
		criticalCount: number;
		overdueCount: number;
		waitingEvidenceCount: number;
		nextDueDate?: string;
	};

	type SnpRow = {
		code: string;
		label: string;
		openCount: number;
		criticalCount: number;
		overdueCount: number;
		waitingEvidenceCount: number;
	};

	type AgendaRow = {
		key: string;
		agenda: string;
		count: number;
		direction: string;
	};

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
			// AsyncContent owns the visible failed state.
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
		return 'Paket rapat tindak lanjut belum dapat dimuat.';
	}

	function isOpenAction(action: GovernanceComplianceAction) {
		return !['done', 'cancelled'].includes(action.status);
	}

	function isCriticalAction(action: GovernanceComplianceAction) {
		return ['high', 'urgent'].includes(action.priority);
	}

	function numericDate(value?: string) {
		if (!value) return Number.MAX_SAFE_INTEGER;
		const time = new Date(value).getTime();
		return Number.isFinite(time) ? time : Number.MAX_SAFE_INTEGER;
	}

	function isOverdue(action: GovernanceComplianceAction) {
		if (!action.due_date || !isOpenAction(action)) return false;
		return dateValue(action.due_date) < dateValue(new Date().toISOString().slice(0, 10));
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

	function buildTotals(data: GovernancePrintPackData): ActionTotals {
		const actions = data.complianceActions;
		return {
			total: actions.length,
			open: actions.filter(isOpenAction).length,
			critical: actions.filter((action) => isOpenAction(action) && isCriticalAction(action)).length,
			overdue: actions.filter(isOverdue).length,
			waitingEvidence: actions.filter((action) => isOpenAction(action) && action.status === 'waiting_evidence').length,
			unassigned: actions.filter((action) => isOpenAction(action) && ownerKey(action) === 'unassigned').length,
			done: actions.filter((action) => action.status === 'done').length,
		};
	}

	function priorityRank(value: string) {
		if (value === 'urgent') return 0;
		if (value === 'high') return 1;
		if (value === 'medium') return 2;
		return 3;
	}

	function sortedMeetingActions(data: GovernancePrintPackData) {
		return data.complianceActions
			.filter(isOpenAction)
			.sort(
				(a, b) =>
					Number(isOverdue(b)) - Number(isOverdue(a)) ||
					priorityRank(a.priority) - priorityRank(b.priority) ||
					numericDate(a.due_date) - numericDate(b.due_date) ||
					a.title.localeCompare(b.title)
			);
	}

	function buildOwnerRows(data: GovernancePrintPackData): OwnerRow[] {
		const rows: OwnerRow[] = [];
		for (const action of data.complianceActions.filter(isOpenAction)) {
			const key = ownerKey(action);
			let row = rows.find((item) => item.key === key);
			if (!row) {
				row = {
					key,
					label: ownerLabel(action),
					unitLabel: ownerUnitLabel(action),
					openCount: 0,
					criticalCount: 0,
					overdueCount: 0,
					waitingEvidenceCount: 0,
					nextDueDate: undefined,
				};
				rows.push(row);
			}
			row.openCount += 1;
			if (isCriticalAction(action)) row.criticalCount += 1;
			if (isOverdue(action)) row.overdueCount += 1;
			if (action.status === 'waiting_evidence') row.waitingEvidenceCount += 1;
			if (numericDate(action.due_date) < numericDate(row.nextDueDate)) row.nextDueDate = action.due_date;
		}
		return rows.sort(
			(a, b) =>
				b.overdueCount - a.overdueCount ||
				b.criticalCount - a.criticalCount ||
				b.openCount - a.openCount ||
				numericDate(a.nextDueDate) - numericDate(b.nextDueDate) ||
				a.label.localeCompare(b.label)
		);
	}

	function buildSnpRows(data: GovernancePrintPackData): SnpRow[] {
		const rows: SnpRow[] = [];
		for (const action of data.complianceActions.filter(isOpenAction)) {
			const code = action.snp_standard || '';
			let row = rows.find((item) => item.code === code);
			if (!row) {
				row = {
					code,
					label: snpLabel(code),
					openCount: 0,
					criticalCount: 0,
					overdueCount: 0,
					waitingEvidenceCount: 0,
				};
				rows.push(row);
			}
			row.openCount += 1;
			if (isCriticalAction(action)) row.criticalCount += 1;
			if (isOverdue(action)) row.overdueCount += 1;
			if (action.status === 'waiting_evidence') row.waitingEvidenceCount += 1;
		}
		return rows.sort((a, b) => b.overdueCount - a.overdueCount || b.criticalCount - a.criticalCount || b.openCount - a.openCount || a.label.localeCompare(b.label));
	}

	function buildAgendaRows(totals: ActionTotals): AgendaRow[] {
		return [
			{
				key: 'overdue',
				agenda: 'Putuskan tindak lanjut lewat tenggat',
				count: totals.overdue,
				direction: 'Tetapkan ulang tenggat, hambatan, dan komando eskalasi.',
			},
			{
				key: 'critical',
				agenda: 'Kunci prioritas tinggi/mendesak',
				count: totals.critical,
				direction: 'Pastikan PIC, bukti, dan target penyelesaian pekan berjalan.',
			},
			{
				key: 'evidence',
				agenda: 'Validasi bukti mutu yang menunggu',
				count: totals.waitingEvidence,
				direction: 'Pastikan lokasi bukti, item bukti, dan status verifikasi.',
			},
			{
				key: 'unassigned',
				agenda: 'Tetapkan PIC yang masih kosong',
				count: totals.unassigned,
				direction: 'Tugaskan unit/pegawai agar setiap gap punya jalur komando.',
			},
		];
	}

	function exportCsv() {
		if (!snapshot) return;
		const actions = sortedMeetingActions(snapshot);
		const owners = buildOwnerRows(snapshot);
		const snpRows = buildSnpRows(snapshot);
		const rows = [
			['Bagian', 'Kode/ID', 'Uraian', 'PIC/Standar', 'Status', 'Catatan'],
			...actions.map((action) => [
				'Prioritas Rapat',
				action.id,
				action.title,
				ownerLabel(action),
				`${actionStatusLabel(action.status)} / ${priorityLabel(action.priority)}`,
				`Tenggat ${formatDate(action.due_date)}; ${linkedLabel(action)}; Bukti ${action.evidence_url || action.evidence_item_title || '-'}`,
			]),
			...owners.map((row) => [
				'Beban PIC/Unit',
				row.key,
				row.label,
				row.unitLabel,
				`${row.openCount} terbuka`,
				`${row.criticalCount} prioritas tinggi; ${row.overdueCount} lewat tenggat; ${row.waitingEvidenceCount} menunggu bukti`,
			]),
			...snpRows.map((row) => [
				'Sebaran 8 SNP',
				row.code || 'tanpa_snp',
				row.label,
				row.label,
				`${row.openCount} terbuka`,
				`${row.criticalCount} prioritas tinggi; ${row.overdueCount} lewat tenggat; ${row.waitingEvidenceCount} menunggu bukti`,
			]),
		];
		downloadCsv(`paket-rapat-tindak-lanjut-${new Date().toISOString().slice(0, 10)}.csv`, rows);
	}

	function printPack() {
		window.print();
	}

	function handleRenderError(error: unknown) {
		console.error('Governance action meeting pack render failed', error);
	}

	onMount(() => {
		void load();
	});
</script>

<svelte:head><title>Paket Rapat Tindak Lanjut - MTsN 2 Kolaka Utara</title></svelte:head>

<div class="space-y-6">
	<div class="no-print flex flex-col gap-3 rounded-md border border-slate-200 bg-white px-4 py-3 lg:flex-row lg:items-center lg:justify-between">
		<div>
			<p class="text-sm font-semibold text-slate-900">Paket Rapat Tindak Lanjut</p>
			<p class="text-sm text-slate-500">Agenda ringkas untuk menutup gap Renstra/IKU/RKT/SKP/8 SNP.</p>
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
			<LoadingButton variant="outline" size="sm" loading={refreshBusy} loadingLabel="Memuat..." onclick={() => void refreshPack()}>
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
					{#each Array.from({ length: 8 }) as _, index (`meeting-pack-skeleton-${index}`)}
						<Skeleton class="h-24 w-full" />
					{/each}
				</div>
				<Skeleton class="h-96 w-full" />
			</div>
		{/snippet}

		{#snippet failed(error, reset)}
			<RecoveryPanel message={errorMessage(error)} onRetry={() => retryPack(reset)} />
		{/snippet}

		{#snippet children(value)}
			{@const data = value as GovernancePrintPackData}
			{@const totals = buildTotals(data)}
			{@const agendaRows = buildAgendaRows(totals)}
			{@const meetingActions = sortedMeetingActions(data).slice(0, 18)}
			{@const ownerRows = buildOwnerRows(data).slice(0, 10)}
			{@const snpRows = buildSnpRows(data)}
			<main class="print-root mx-auto max-w-6xl space-y-6 bg-white text-slate-950">
				<section class="print-section rounded-md border border-slate-200 p-5">
					<div class="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
						<div>
							<p class="text-xs font-semibold uppercase tracking-[0.16em] text-emerald-700">{data.schoolProfile.ministry_line}</p>
							<h1 class="mt-2 text-2xl font-bold text-slate-950">Paket Rapat Tindak Lanjut Kepatuhan</h1>
							<p class="mt-1 text-sm text-slate-600">{data.schoolProfile.name}</p>
							<p class="mt-1 max-w-2xl text-xs text-slate-500">{schoolAddressLine(data.schoolProfile) || data.schoolProfile.office_line}</p>
						</div>
						<div class="rounded-md border border-slate-200 px-4 py-3 text-sm">
							<p class="font-medium text-slate-900">Waktu cetak</p>
							<p class="text-slate-600">{formatGeneratedAt(generatedAt)} WITA</p>
						</div>
					</div>
				</section>

				<section class="print-section rounded-md border border-slate-200 p-5">
					<h2 class="text-base font-semibold text-slate-950">Ringkasan Rapat</h2>
					<div class="mt-4 grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
						{#each [
							{ label: 'Total Register', value: totals.total, note: `${totals.done} selesai` },
							{ label: 'Masih Terbuka', value: totals.open, note: 'Perlu keputusan/PIC' },
							{ label: 'Prioritas Tinggi', value: totals.critical, note: 'High/urgent' },
							{ label: 'Lewat Tenggat', value: totals.overdue, note: 'Butuh eskalasi' },
							{ label: 'Menunggu Bukti', value: totals.waitingEvidence, note: 'Perlu validasi bukti' },
							{ label: 'Tanpa PIC', value: totals.unassigned, note: 'Perlu jalur komando' },
							{ label: 'Beban PIC/Unit', value: ownerRows.length, note: 'Unit/pegawai aktif' },
							{ label: 'Standar SNP', value: snpRows.length, note: 'Dengan action aktif' },
						] as item (item.label)}
							<div class="rounded-md border border-slate-200 p-4">
								<p class="text-xs text-slate-500">{item.label}</p>
								<p class="mt-2 text-2xl font-semibold text-slate-950">{item.value}</p>
								<p class="mt-1 text-xs text-slate-500">{item.note}</p>
							</div>
						{/each}
					</div>
				</section>

				<section class="print-section rounded-md border border-slate-200">
					<div class="border-b border-slate-200 px-5 py-4">
						<h2 class="text-base font-semibold text-slate-950">Agenda Keputusan</h2>
					</div>
					<div class="overflow-x-auto">
						<Table.Root>
							<Table.Header>
								<Table.Row>
									<Table.Head>Agenda</Table.Head>
									<Table.Head>Jumlah</Table.Head>
									<Table.Head>Arah Keputusan</Table.Head>
									<Table.Head>Catatan Rapat</Table.Head>
								</Table.Row>
							</Table.Header>
							<Table.Body>
								{#each agendaRows as row (row.key)}
									<Table.Row>
										<Table.Cell class="text-sm font-medium text-slate-900">{row.agenda}</Table.Cell>
										<Table.Cell class="text-sm">{row.count}</Table.Cell>
										<Table.Cell class="max-w-lg text-sm text-slate-600">{row.direction}</Table.Cell>
										<Table.Cell class="min-w-64 text-sm text-slate-400">................................................................</Table.Cell>
									</Table.Row>
								{/each}
							</Table.Body>
						</Table.Root>
					</div>
				</section>

				<section class="print-section rounded-md border border-slate-200">
					<div class="border-b border-slate-200 px-5 py-4">
						<h2 class="text-base font-semibold text-slate-950">Prioritas Pembahasan</h2>
						<p class="mt-1 text-xs text-slate-500">Diurutkan dari lewat tenggat, prioritas, dan tenggat terdekat.</p>
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
								</Table.Row>
							</Table.Header>
							<Table.Body>
								{#each meetingActions as action (action.id)}
									<Table.Row>
										<Table.Cell class="min-w-64">
											<p class="text-sm font-medium text-slate-900">{action.title}</p>
											<p class="text-xs text-slate-500">{action.period_year} - {sourceLabel(action.source_type)} - {snpLabel(action.snp_standard)}</p>
										</Table.Cell>
										<Table.Cell class="min-w-44">
											<p class="text-sm">{ownerLabel(action)}</p>
											<p class="text-xs text-slate-500">{ownerUnitLabel(action)}</p>
										</Table.Cell>
										<Table.Cell>
											<div class="flex flex-col gap-1">
												<Badge variant={statusVariant(action.status)}>{actionStatusLabel(action.status)}</Badge>
												<Badge variant={priorityVariant(action.priority)}>{priorityLabel(action.priority)}</Badge>
											</div>
										</Table.Cell>
										<Table.Cell class={isOverdue(action) ? 'whitespace-nowrap text-sm font-medium text-red-700' : 'whitespace-nowrap text-sm text-slate-700'}>
											{formatDate(action.due_date)}
										</Table.Cell>
										<Table.Cell class="max-w-md text-sm text-slate-600">
											<p>{linkedLabel(action)}</p>
											<p class="mt-1 text-xs text-slate-500">{action.evidence_url || action.evidence_item_title || 'Bukti belum dicatat'}</p>
										</Table.Cell>
									</Table.Row>
								{:else}
									<Table.Row><Table.Cell colspan={5} class="text-center text-sm text-slate-500">Belum ada tindak lanjut aktif.</Table.Cell></Table.Row>
								{/each}
							</Table.Body>
						</Table.Root>
					</div>
				</section>

				<section class="print-section grid gap-6 lg:grid-cols-2">
					<div class="rounded-md border border-slate-200">
						<div class="border-b border-slate-200 px-5 py-4">
							<h2 class="text-base font-semibold text-slate-950">Beban PIC/Unit</h2>
						</div>
						<div class="space-y-2 p-5">
							{#each ownerRows as row (row.key)}
								<div class="rounded-md border border-slate-200 px-3 py-2">
									<div class="flex items-start justify-between gap-3">
										<div>
											<p class="text-sm font-medium text-slate-900">{row.label}</p>
											<p class="text-xs text-slate-500">{row.unitLabel}</p>
										</div>
										<p class="text-sm font-semibold text-slate-900">{row.openCount}</p>
									</div>
									<p class="mt-1 text-xs text-slate-500">
										{row.criticalCount} tinggi, {row.overdueCount} lewat, {row.waitingEvidenceCount} tunggu bukti, tenggat {formatDate(row.nextDueDate)}
									</p>
								</div>
							{:else}
								<p class="text-sm text-slate-500">Tidak ada beban PIC/unit aktif.</p>
							{/each}
						</div>
					</div>

					<div class="rounded-md border border-slate-200">
						<div class="border-b border-slate-200 px-5 py-4">
							<h2 class="text-base font-semibold text-slate-950">Sebaran 8 SNP</h2>
						</div>
						<div class="space-y-2 p-5">
							{#each snpRows as row (row.code)}
								<div class="rounded-md border border-slate-200 px-3 py-2">
									<div class="flex items-start justify-between gap-3">
										<p class="text-sm font-medium text-slate-900">{row.label}</p>
										<p class="text-sm font-semibold text-slate-900">{row.openCount}</p>
									</div>
									<p class="mt-1 text-xs text-slate-500">{row.criticalCount} tinggi, {row.overdueCount} lewat, {row.waitingEvidenceCount} tunggu bukti</p>
								</div>
							{:else}
								<p class="text-sm text-slate-500">Belum ada tindak lanjut aktif yang dikaitkan ke SNP.</p>
							{/each}
						</div>
					</div>
				</section>

				<section class="print-section rounded-md border border-slate-200">
					<div class="border-b border-slate-200 px-5 py-4">
						<h2 class="text-base font-semibold text-slate-950">Keputusan Rapat</h2>
					</div>
					<div class="overflow-x-auto">
						<Table.Root>
							<Table.Header>
								<Table.Row>
									<Table.Head>Keputusan</Table.Head>
									<Table.Head>PIC</Table.Head>
									<Table.Head>Tenggat</Table.Head>
									<Table.Head>Paraf</Table.Head>
								</Table.Row>
							</Table.Header>
							<Table.Body>
								{#each Array.from({ length: 6 }) as _, index (`decision-row-${index}`)}
									<Table.Row>
										<Table.Cell class="h-14 text-slate-400">................................................................</Table.Cell>
										<Table.Cell class="text-slate-400">........................</Table.Cell>
										<Table.Cell class="text-slate-400">........................</Table.Cell>
										<Table.Cell class="text-slate-400">........................</Table.Cell>
									</Table.Row>
								{/each}
							</Table.Body>
						</Table.Root>
					</div>
				</section>

				<section class="print-section grid gap-6 md:grid-cols-3">
					{#each [
						{ role: 'Kepala Madrasah', name: data.schoolProfile.head_name || '................................', nip: data.schoolProfile.head_nip },
						{ role: 'Koordinator Tata Kelola', name: '................................', nip: '' },
						{ role: 'Notulis', name: '................................', nip: '' },
					] as signer (signer.role)}
						<div class="rounded-md border border-slate-200 p-5 text-center">
							<p class="text-sm font-medium text-slate-900">{signer.role}</p>
							<div class="h-20"></div>
							<p class="text-sm font-semibold text-slate-900">{signer.name}</p>
							<p class="text-xs text-slate-500">{signer.nip ? `NIP. ${signer.nip}` : 'NIP. ................................'}</p>
						</div>
					{/each}
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
