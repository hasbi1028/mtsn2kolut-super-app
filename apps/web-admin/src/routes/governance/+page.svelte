<script lang="ts">
	import CalendarDaysIcon from '@lucide/svelte/icons/calendar-days';
	import ClipboardCheckIcon from '@lucide/svelte/icons/clipboard-check';
	import FileTextIcon from '@lucide/svelte/icons/file-text';
	import PrinterIcon from '@lucide/svelte/icons/printer';
	import { resolve } from '$app/paths';
	import { onMount } from 'svelte';
	import * as Card from '$lib/components/ui/card';
	import * as Dialog from '$lib/components/ui/dialog';
	import * as Table from '$lib/components/ui/table';
	import * as Tabs from '$lib/components/ui/tabs';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { Textarea } from '$lib/components/ui/textarea';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import EmptyStatePanel from '$lib/components/EmptyStatePanel.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { toast } from '$lib/components/ui/sonner';
	import { confirmAction } from '$lib/confirm-dialog';
	import { readClientApiData, readClientJson } from '$lib/client/api';

	interface Stats {
		total_units: number;
		total_positions: number;
		active_assignments: number;
		final_documents: number;
		active_programs: number;
		blocked_programs: number;
		completed_programs: number;
		programs_with_evidence: number;
		performance_targets: number;
		completed_performance_targets: number;
		blocked_performance_targets: number;
		evidence_items: number;
		verified_evidence_items: number;
		gap_evidence_items: number;
		work_plan_items: number;
		completed_work_plan_items: number;
		blocked_work_plan_items: number;
		work_plan_budget_amount: number;
		work_plan_realization_amount: number;
		compliance_actions: number;
		open_compliance_actions: number;
		completed_compliance_actions: number;
		critical_compliance_actions: number;
	}

	interface UnitRow {
		id: string;
		code: string;
		name: string;
		unit_type: string;
		parent_id?: string;
		parent_unit_name?: string;
		description: string;
		is_active: boolean;
		sort_order: number;
	}

	interface PositionRow {
		id: string;
		unit_id: string;
		title: string;
		position_type: string;
		parent_position_id?: string;
		parent_position_title?: string;
		description: string;
		tupoksi: string;
		is_active: boolean;
		sort_order: number;
		unit_name: string;
		active_employee_name?: string;
		active_employee_nip?: string;
	}

	interface AssignmentRow {
		id: string;
		position_id: string;
		employee_id: string;
		start_date: string;
		end_date?: string;
		notes: string;
		position_title: string;
		unit_name: string;
		employee_name: string;
		employee_nip: string;
		decree_nomor_surat?: string;
	}

	interface DocumentRow {
		id: string;
		doc_type: string;
		title: string;
		period_year: number;
		period_label: string;
		owner_unit_id?: string;
		owner_unit_name?: string;
		snp_standard: string;
		status: string;
		document_url: string;
		summary: string;
		updated_at: string;
	}

	interface ProgramRow {
		id: string;
		period_year: number;
		code: string;
		name: string;
		source_document_id?: string;
		owner_unit_id?: string;
		responsible_position_id?: string;
		responsible_employee_id?: string;
		snp_standard: string;
		iku_code: string;
		indicator: string;
		target_value: string;
		target_unit: string;
		status: string;
		progress_percent: number;
		realization_summary: string;
		evidence_url: string;
		due_date?: string;
		source_document_title?: string;
		owner_unit_name?: string;
		responsible_position_title?: string;
		responsible_employee_name?: string;
	}

	interface StrategicAlignmentRow {
		key: string;
		period_year: number;
		program_code: string;
		program_name: string;
		iku_code: string;
		snp_standard: string;
		source_document_title: string;
		source_document_type: string;
		owner_label: string;
		work_plan_count: number;
		work_plan_done: number;
		work_plan_blocked: number;
		performance_count: number;
		performance_done: number;
		evidence_count: number;
		verified_evidence_count: number;
		progress_percent: number;
		readiness_status: string;
		gap_notes: string[];
	}

	interface PerformanceTargetRow {
		id: string;
		period_year: number;
		employee_id: string;
		position_id?: string;
		program_id?: string;
		parent_target_id?: string;
		aspect: string;
		title: string;
		indicator: string;
		target_value: string;
		target_unit: string;
		status: string;
		progress_percent: number;
		evidence_url: string;
		review_notes: string;
		due_date?: string;
		employee_name: string;
		employee_nip: string;
		position_title?: string;
		program_code?: string;
		program_name?: string;
		parent_target_title?: string;
	}

	interface WorkPlanItemRow {
		id: string;
		period_year: number;
		program_id: string;
		source_document_id?: string;
		owner_unit_id?: string;
		responsible_employee_id?: string;
		evidence_item_id?: string;
		activity_code: string;
		activity_name: string;
		output_indicator: string;
		target_volume: string;
		target_unit: string;
		budget_source: string;
		budget_amount: number;
		realization_amount: number;
		status: string;
		progress_percent: number;
		start_date?: string;
		end_date?: string;
		evidence_url: string;
		notes: string;
		program_code: string;
		program_name: string;
		program_snp_standard: string;
		source_document_title?: string;
		owner_unit_name?: string;
		responsible_employee_name?: string;
		responsible_employee_nip?: string;
		evidence_item_title?: string;
	}

	interface EvidenceItemRow {
		id: string;
		period_year: number;
		title: string;
		evidence_type: string;
		snp_standard: string;
		owner_unit_id?: string;
		document_id?: string;
		program_id?: string;
		performance_target_id?: string;
		source_module: string;
		evidence_url: string;
		status: string;
		notes: string;
		updated_at: string;
		owner_unit_name?: string;
		document_title?: string;
		program_code?: string;
		program_name?: string;
		performance_target_title?: string;
	}

	interface SNPRow {
		code: string;
		name: string;
		total_documents: number;
		total_programs: number;
		completed_programs: number;
		evidence_programs: number;
		total_evidence_items: number;
		verified_evidence_items: number;
		gap_evidence_items: number;
		last_updated_at?: string;
	}

	interface EmployeeOption {
		id: string;
		nip: string;
		nama: string;
		unit_kerja: string;
	}

	interface GovernanceOverview {
		stats: Stats;
		units: UnitRow[];
		positions: PositionRow[];
		assignments: AssignmentRow[];
		documents: DocumentRow[];
		programs: ProgramRow[];
		workPlanItems: WorkPlanItemRow[];
		performanceTargets: PerformanceTargetRow[];
		evidenceItems: EvidenceItemRow[];
		snpMatrix: SNPRow[];
		employees: EmployeeOption[];
	}

	interface UnitForm {
		code: string;
		name: string;
		unit_type: string;
		parent_id: string;
		description: string;
		is_active: boolean;
		sort_order: number;
	}

	interface PositionForm {
		unit_id: string;
		title: string;
		position_type: string;
		parent_position_id: string;
		description: string;
		tupoksi: string;
		is_active: boolean;
		sort_order: number;
	}

	interface AssignmentForm {
		position_id: string;
		employee_id: string;
		start_date: string;
		end_date: string;
		notes: string;
	}

	interface DocumentForm {
		doc_type: string;
		title: string;
		period_year: number;
		period_label: string;
		owner_unit_id: string;
		snp_standard: string;
		status: string;
		document_url: string;
		summary: string;
	}

	interface ProgramForm {
		period_year: number;
		code: string;
		name: string;
		source_document_id: string;
		owner_unit_id: string;
		responsible_position_id: string;
		responsible_employee_id: string;
		snp_standard: string;
		iku_code: string;
		indicator: string;
		target_value: string;
		target_unit: string;
		status: string;
		progress_percent: number;
		realization_summary: string;
		evidence_url: string;
		due_date: string;
	}

	interface PerformanceTargetForm {
		period_year: number;
		employee_id: string;
		position_id: string;
		program_id: string;
		parent_target_id: string;
		aspect: string;
		title: string;
		indicator: string;
		target_value: string;
		target_unit: string;
		status: string;
		progress_percent: number;
		evidence_url: string;
		review_notes: string;
		due_date: string;
	}

	interface WorkPlanItemForm {
		period_year: number;
		program_id: string;
		source_document_id: string;
		owner_unit_id: string;
		responsible_employee_id: string;
		evidence_item_id: string;
		activity_code: string;
		activity_name: string;
		output_indicator: string;
		target_volume: string;
		target_unit: string;
		budget_source: string;
		budget_amount: number;
		realization_amount: number;
		status: string;
		progress_percent: number;
		start_date: string;
		end_date: string;
		evidence_url: string;
		notes: string;
	}

	interface EvidenceItemForm {
		period_year: number;
		title: string;
		evidence_type: string;
		snp_standard: string;
		owner_unit_id: string;
		document_id: string;
		program_id: string;
		performance_target_id: string;
		source_module: string;
		evidence_url: string;
		status: string;
		notes: string;
	}

	const DOC_TYPES = [
		['visi_misi', 'Visi Misi'],
		['rkjm', 'RKJM'],
		['rkt', 'RKT'],
		['renstra', 'Renstra'],
		['perkin', 'Perkin'],
		['iku', 'IKU'],
		['sk', 'SK'],
		['sop', 'SOP'],
		['snp', '8 SNP'],
		['lainnya', 'Lainnya'],
	] as const;

	const SNP_STANDARDS = [
		['', 'Tidak dikaitkan'],
		['skl', 'SKL'],
		['isi', 'Isi'],
		['proses', 'Proses'],
		['penilaian', 'Penilaian'],
		['ptk', 'PTK'],
		['sarpras', 'Sarpras'],
		['pengelolaan', 'Pengelolaan'],
		['pembiayaan', 'Pembiayaan'],
	] as const;

	const PROGRAM_STATUSES = [
		['planned', 'Direncanakan'],
		['in_progress', 'Berjalan'],
		['done', 'Selesai'],
		['blocked', 'Terkendala'],
	] as const;

	const PERFORMANCE_ASPECTS = [
		['hasil_kerja', 'Hasil Kerja'],
		['perilaku_kerja', 'Perilaku Kerja'],
		['tambahan', 'Tambahan'],
	] as const;

	const EVIDENCE_TYPES = [
		['dokumen', 'Dokumen'],
		['foto', 'Foto'],
		['tautan', 'Tautan'],
		['laporan', 'Laporan'],
		['arsip', 'Arsip'],
		['lainnya', 'Lainnya'],
	] as const;

	const EVIDENCE_STATUSES = [
		['needed', 'Dibutuhkan'],
		['collected', 'Terkumpul'],
		['verified', 'Terverifikasi'],
		['gap', 'Kurang'],
	] as const;

	const DOCUMENT_STATUSES = [
		['draft', 'Draft'],
		['final', 'Final'],
		['arsip', 'Arsip'],
	] as const;

	const STRATEGIC_DOCUMENT_ORDER = ['visi_misi', 'renstra', 'rkjm', 'rkt', 'perkin', 'iku'];
	const STRATEGIC_DOCUMENT_TYPES = new Set(STRATEGIC_DOCUMENT_ORDER);

	let governancePromise = $state<Promise<GovernanceOverview> | null>(null);
	let governanceRequestId = 0;
	let stats = $state<Stats>(emptyStats());
	let units = $state<UnitRow[]>([]);
	let positions = $state<PositionRow[]>([]);
	let assignments = $state<AssignmentRow[]>([]);
	let documents = $state<DocumentRow[]>([]);
	let programs = $state<ProgramRow[]>([]);
	let workPlanItems = $state<WorkPlanItemRow[]>([]);
	let performanceTargets = $state<PerformanceTargetRow[]>([]);
	let evidenceItems = $state<EvidenceItemRow[]>([]);
	let snpMatrix = $state<SNPRow[]>([]);
	let employees = $state<EmployeeOption[]>([]);
	let activeTab = $state('organisasi');
	let dialogOpen = $state(false);
	let dialogKind = $state<'unit' | 'position' | 'assignment' | 'document' | 'program' | 'workplan' | 'performance' | 'evidence' | null>(null);
	let editingId = $state<string | null>(null);
	let busy = $state(false);
	let orgSearch = $state('');
	let documentSearch = $state('');
	let programSearch = $state('');
	let workPlanSearch = $state('');
	let performanceSearch = $state('');
	let evidenceSearch = $state('');
	let alignmentSearch = $state('');

	let unitForm = $state<UnitForm>(emptyUnitForm());
	let positionForm = $state<PositionForm>(emptyPositionForm());
	let assignmentForm = $state<AssignmentForm>(emptyAssignmentForm());
	let documentForm = $state<DocumentForm>(emptyDocumentForm());
	let programForm = $state<ProgramForm>(emptyProgramForm());
	let workPlanItemForm = $state<WorkPlanItemForm>(emptyWorkPlanItemForm());
	let performanceTargetForm = $state<PerformanceTargetForm>(emptyPerformanceTargetForm());
	let evidenceItemForm = $state<EvidenceItemForm>(emptyEvidenceItemForm());

	const filteredUnits = $derived.by(() => {
		const q = orgSearch.trim().toLowerCase();
		if (!q) return units;
		return units.filter((unit) =>
			[unit.code, unit.name, unit.unit_type, unit.parent_unit_name ?? ''].some((value) => value.toLowerCase().includes(q))
		);
	});

	const filteredPositions = $derived.by(() => {
		const q = orgSearch.trim().toLowerCase();
		if (!q) return positions;
		return positions.filter((position) =>
			[position.title, position.unit_name, position.position_type, position.active_employee_name ?? ''].some((value) =>
				value.toLowerCase().includes(q)
			)
		);
	});

	const filteredDocuments = $derived.by(() => {
		const q = documentSearch.trim().toLowerCase();
		if (!q) return documents;
		return documents.filter((document) =>
			[document.title, document.doc_type, document.owner_unit_name ?? '', document.summary].some((value) =>
				value.toLowerCase().includes(q)
			)
		);
	});

	const filteredPrograms = $derived.by(() => {
		const q = programSearch.trim().toLowerCase();
		if (!q) return programs;
		return programs.filter((program) =>
			[program.code, program.name, program.iku_code, program.indicator, program.owner_unit_name ?? ''].some((value) =>
				value.toLowerCase().includes(q)
			)
		);
	});

	const filteredWorkPlanItems = $derived.by(() => {
		const q = workPlanSearch.trim().toLowerCase();
		if (!q) return workPlanItems;
		return workPlanItems.filter((item) =>
			[
				item.activity_code,
				item.activity_name,
				item.output_indicator,
				item.program_code,
				item.program_name,
				item.owner_unit_name ?? '',
				item.responsible_employee_name ?? '',
				item.budget_source,
			].some((value) => value.toLowerCase().includes(q))
		);
	});

	const filteredPerformanceTargets = $derived.by(() => {
		const q = performanceSearch.trim().toLowerCase();
		if (!q) return performanceTargets;
		return performanceTargets.filter((target) =>
			[
				target.title,
				target.indicator,
				target.employee_name,
				target.employee_nip,
				target.program_code ?? '',
				target.program_name ?? '',
			].some((value) => value.toLowerCase().includes(q))
		);
	});

	const filteredEvidenceItems = $derived.by(() => {
		const q = evidenceSearch.trim().toLowerCase();
		if (!q) return evidenceItems;
		return evidenceItems.filter((item) =>
			[
				item.title,
				item.evidence_type,
				item.status,
				item.source_module,
				item.notes,
				item.owner_unit_name ?? '',
				item.document_title ?? '',
				item.program_code ?? '',
				item.program_name ?? '',
				item.performance_target_title ?? '',
			].some((value) => value.toLowerCase().includes(q))
		);
	});

	const strategicDocuments = $derived.by(() =>
		documents
			.filter((document) => STRATEGIC_DOCUMENT_TYPES.has(document.doc_type))
			.sort((a, b) => b.period_year - a.period_year || STRATEGIC_DOCUMENT_ORDER.indexOf(a.doc_type) - STRATEGIC_DOCUMENT_ORDER.indexOf(b.doc_type))
	);

	const alignmentRows = $derived.by(() => buildStrategicAlignmentRows());

	const filteredAlignmentRows = $derived.by(() => {
		const q = alignmentSearch.trim().toLowerCase();
		if (!q) return alignmentRows;
		return alignmentRows.filter((row) =>
			[
				row.program_code,
				row.program_name,
				row.iku_code,
				row.source_document_title,
				row.owner_label,
				row.readiness_status,
				row.gap_notes.join(' '),
			].some((value) => value.toLowerCase().includes(q))
		);
	});

	const alignmentSummary = $derived.by(() => ({
		total_programs: alignmentRows.length,
		final_strategic_documents: strategicDocuments.filter((document) => document.status === 'final').length,
		mapped_iku: alignmentRows.filter((row) => row.iku_code).length,
		complete: alignmentRows.filter((row) => row.readiness_status === 'lengkap').length,
		gaps: alignmentRows.filter((row) => row.gap_notes.length > 0).length,
		blocked: alignmentRows.filter((row) => row.readiness_status === 'terkendala').length,
	}));

	function emptyStats(): Stats {
		return {
			total_units: 0,
			total_positions: 0,
			active_assignments: 0,
			final_documents: 0,
			active_programs: 0,
			blocked_programs: 0,
			completed_programs: 0,
			programs_with_evidence: 0,
			performance_targets: 0,
			completed_performance_targets: 0,
			blocked_performance_targets: 0,
			evidence_items: 0,
			verified_evidence_items: 0,
			gap_evidence_items: 0,
			work_plan_items: 0,
			completed_work_plan_items: 0,
			blocked_work_plan_items: 0,
			work_plan_budget_amount: 0,
			work_plan_realization_amount: 0,
			compliance_actions: 0,
			open_compliance_actions: 0,
			completed_compliance_actions: 0,
			critical_compliance_actions: 0,
		};
	}

	function emptyUnitForm(): UnitForm {
		return { code: '', name: '', unit_type: 'madrasah', parent_id: '', description: '', is_active: true, sort_order: 0 };
	}

	function emptyPositionForm(): PositionForm {
		return {
			unit_id: '',
			title: '',
			position_type: 'struktural',
			parent_position_id: '',
			description: '',
			tupoksi: '',
			is_active: true,
			sort_order: 0,
		};
	}

	function emptyAssignmentForm(): AssignmentForm {
		return { position_id: '', employee_id: '', start_date: new Date().toISOString().slice(0, 10), end_date: '', notes: '' };
	}

	function emptyDocumentForm(): DocumentForm {
		return {
			doc_type: 'rkt',
			title: '',
			period_year: new Date().getFullYear(),
			period_label: '',
			owner_unit_id: '',
			snp_standard: '',
			status: 'draft',
			document_url: '',
			summary: '',
		};
	}

	function emptyProgramForm(): ProgramForm {
		return {
			period_year: new Date().getFullYear(),
			code: '',
			name: '',
			source_document_id: '',
			owner_unit_id: '',
			responsible_position_id: '',
			responsible_employee_id: '',
			snp_standard: '',
			iku_code: '',
			indicator: '',
			target_value: '',
			target_unit: '',
			status: 'planned',
			progress_percent: 0,
			realization_summary: '',
			evidence_url: '',
			due_date: '',
		};
	}

	function emptyWorkPlanItemForm(): WorkPlanItemForm {
		return {
			period_year: new Date().getFullYear(),
			program_id: programs[0]?.id ?? '',
			source_document_id: '',
			owner_unit_id: '',
			responsible_employee_id: '',
			evidence_item_id: '',
			activity_code: '',
			activity_name: '',
			output_indicator: '',
			target_volume: '',
			target_unit: '',
			budget_source: 'BOS',
			budget_amount: 0,
			realization_amount: 0,
			status: 'planned',
			progress_percent: 0,
			start_date: '',
			end_date: '',
			evidence_url: '',
			notes: '',
		};
	}

	function emptyPerformanceTargetForm(): PerformanceTargetForm {
		return {
			period_year: new Date().getFullYear(),
			employee_id: '',
			position_id: '',
			program_id: '',
			parent_target_id: '',
			aspect: 'hasil_kerja',
			title: '',
			indicator: '',
			target_value: '',
			target_unit: '',
			status: 'planned',
			progress_percent: 0,
			evidence_url: '',
			review_notes: '',
			due_date: '',
		};
	}

	function emptyEvidenceItemForm(): EvidenceItemForm {
		return {
			period_year: new Date().getFullYear(),
			title: '',
			evidence_type: 'dokumen',
			snp_standard: '',
			owner_unit_id: '',
			document_id: '',
			program_id: '',
			performance_target_id: '',
			source_module: 'governance',
			evidence_url: '',
			status: 'needed',
			notes: '',
		};
	}

	function normalizeStats(value: Partial<Stats> | null | undefined): Stats {
		return { ...emptyStats(), ...(value ?? {}) };
	}

	function normalizeRows<T>(value: T[] | null | undefined): T[] {
		return Array.isArray(value) ? value : [];
	}

	function applyGovernanceOverview(overview: GovernanceOverview) {
		stats = overview.stats;
		units = overview.units;
		positions = overview.positions;
		assignments = overview.assignments;
		documents = overview.documents;
		programs = overview.programs;
		workPlanItems = overview.workPlanItems;
		performanceTargets = overview.performanceTargets;
		evidenceItems = overview.evidenceItems;
		snpMatrix = overview.snpMatrix;
		employees = overview.employees;
	}

	function currentGovernanceOverview(): GovernanceOverview {
		return {
			stats,
			units,
			positions,
			assignments,
			documents,
			programs,
			workPlanItems,
			performanceTargets,
			evidenceItems,
			snpMatrix,
			employees,
		};
	}

	async function fetchGovernanceOverview(): Promise<GovernanceOverview> {
		const [statsData, unitsData, positionsData, assignmentsData, documentsData, programsData, workPlanData, performanceData, evidenceData, snpData, employeesData] =
			await Promise.all([
				fetch('/api/governance/stats').then((response) => readClientApiData<Partial<Stats>>(response, 'Gagal memuat statistik tata kelola')),
				fetch('/api/governance/units').then((response) => readClientApiData<UnitRow[]>(response, 'Gagal memuat unit kerja')),
				fetch('/api/governance/positions').then((response) => readClientApiData<PositionRow[]>(response, 'Gagal memuat jabatan')),
				fetch('/api/governance/assignments?active_only=true').then((response) => readClientApiData<AssignmentRow[]>(response, 'Gagal memuat pejabat aktif')),
				fetch('/api/governance/documents').then((response) => readClientApiData<DocumentRow[]>(response, 'Gagal memuat dokumen tata kelola')),
				fetch('/api/governance/programs').then((response) => readClientApiData<ProgramRow[]>(response, 'Gagal memuat program tata kelola')),
				fetch('/api/governance/work-plan-items').then((response) => readClientApiData<WorkPlanItemRow[]>(response, 'Gagal memuat RKT/RKJM')),
				fetch('/api/governance/performance-targets').then((response) => readClientApiData<PerformanceTargetRow[]>(response, 'Gagal memuat target kinerja')),
				fetch('/api/governance/evidence-items').then((response) => readClientApiData<EvidenceItemRow[]>(response, 'Gagal memuat bukti mutu')),
				fetch('/api/governance/snp-matrix').then((response) => readClientApiData<SNPRow[]>(response, 'Gagal memuat matriks 8 SNP')),
				fetch('/api/governance/employee-options').then((response) => readClientApiData<EmployeeOption[]>(response, 'Gagal memuat opsi pegawai')),
			]);

		return {
			stats: normalizeStats(statsData),
			units: normalizeRows(unitsData),
			positions: normalizeRows(positionsData),
			assignments: normalizeRows(assignmentsData),
			documents: normalizeRows(documentsData),
			programs: normalizeRows(programsData),
			workPlanItems: normalizeRows(workPlanData),
			performanceTargets: normalizeRows(performanceData),
			evidenceItems: normalizeRows(evidenceData),
			snpMatrix: normalizeRows(snpData),
			employees: normalizeRows(employeesData),
		};
	}

	function loadGovernance() {
		const requestId = ++governanceRequestId;
		governancePromise = fetchGovernanceOverview()
			.then((overview) => {
				if (requestId === governanceRequestId) {
					applyGovernanceOverview(overview);
					return overview;
				}
				return currentGovernanceOverview();
			})
			.catch((error: unknown) => {
				if (requestId === governanceRequestId) throw error;
				return currentGovernanceOverview();
			});
		return governancePromise;
	}

	async function refreshGovernance() {
		if (!governancePromise) {
			await loadGovernance();
			return;
		}
		const requestId = ++governanceRequestId;
		try {
			const overview = await fetchGovernanceOverview();
			if (requestId === governanceRequestId) {
				applyGovernanceOverview(overview);
				governancePromise = Promise.resolve(overview);
			}
		} catch (error) {
			if (requestId !== governanceRequestId) return;
			governancePromise = Promise.resolve(currentGovernanceOverview());
			toast.error(governanceErrorMessage(error));
		}
	}

	function retryGovernance(reset?: () => void) {
		reset?.();
		loadGovernance();
	}

	function governanceErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		if (typeof error === 'string' && error.trim()) return error;
		return 'Data tata kelola belum dapat dimuat. Periksa koneksi backend lalu coba lagi.';
	}

	function mutationErrorMessage(error: unknown, fallback: string) {
		if (error instanceof Error && error.message.trim()) return error.message;
		if (typeof error === 'string' && error.trim()) return error;
		return fallback;
	}

	function handleGovernanceRenderError(error: unknown, reset: () => void) {
		console.error('Governance overview render failed', error);
		reset();
	}

	function openCreate(kind: NonNullable<typeof dialogKind>) {
		editingId = null;
		dialogKind = kind;
		if (kind === 'unit') unitForm = emptyUnitForm();
		if (kind === 'position') positionForm = emptyPositionForm();
		if (kind === 'assignment') assignmentForm = emptyAssignmentForm();
		if (kind === 'document') documentForm = emptyDocumentForm();
		if (kind === 'program') programForm = emptyProgramForm();
		if (kind === 'workplan') workPlanItemForm = emptyWorkPlanItemForm();
		if (kind === 'performance') performanceTargetForm = emptyPerformanceTargetForm();
		if (kind === 'evidence') evidenceItemForm = emptyEvidenceItemForm();
		dialogOpen = true;
	}

	function openEditUnit(unit: UnitRow) {
		editingId = unit.id;
		dialogKind = 'unit';
		unitForm = {
			code: unit.code,
			name: unit.name,
			unit_type: unit.unit_type,
			parent_id: unit.parent_id ?? '',
			description: unit.description,
			is_active: unit.is_active,
			sort_order: unit.sort_order,
		};
		dialogOpen = true;
	}

	function openEditPosition(position: PositionRow) {
		editingId = position.id;
		dialogKind = 'position';
		positionForm = {
			unit_id: position.unit_id,
			title: position.title,
			position_type: position.position_type,
			parent_position_id: position.parent_position_id ?? '',
			description: position.description,
			tupoksi: position.tupoksi,
			is_active: position.is_active,
			sort_order: position.sort_order,
		};
		dialogOpen = true;
	}

	function openEditAssignment(assignment: AssignmentRow) {
		editingId = assignment.id;
		dialogKind = 'assignment';
		assignmentForm = {
			position_id: assignment.position_id,
			employee_id: assignment.employee_id,
			start_date: dateValue(assignment.start_date),
			end_date: dateValue(assignment.end_date),
			notes: assignment.notes,
		};
		dialogOpen = true;
	}

	function openEditDocument(document: DocumentRow) {
		editingId = document.id;
		dialogKind = 'document';
		documentForm = {
			doc_type: document.doc_type,
			title: document.title,
			period_year: document.period_year,
			period_label: document.period_label,
			owner_unit_id: document.owner_unit_id ?? '',
			snp_standard: document.snp_standard,
			status: document.status,
			document_url: document.document_url,
			summary: document.summary,
		};
		dialogOpen = true;
	}

	function openEditProgram(program: ProgramRow) {
		editingId = program.id;
		dialogKind = 'program';
		programForm = {
			period_year: program.period_year,
			code: program.code,
			name: program.name,
			source_document_id: program.source_document_id ?? '',
			owner_unit_id: program.owner_unit_id ?? '',
			responsible_position_id: program.responsible_position_id ?? '',
			responsible_employee_id: program.responsible_employee_id ?? '',
			snp_standard: program.snp_standard,
			iku_code: program.iku_code,
			indicator: program.indicator,
			target_value: program.target_value,
			target_unit: program.target_unit,
			status: program.status,
			progress_percent: program.progress_percent,
			realization_summary: program.realization_summary,
			evidence_url: program.evidence_url,
			due_date: dateValue(program.due_date),
		};
		dialogOpen = true;
	}

	function openEditWorkPlanItem(item: WorkPlanItemRow) {
		editingId = item.id;
		dialogKind = 'workplan';
		workPlanItemForm = {
			period_year: item.period_year,
			program_id: item.program_id,
			source_document_id: item.source_document_id ?? '',
			owner_unit_id: item.owner_unit_id ?? '',
			responsible_employee_id: item.responsible_employee_id ?? '',
			evidence_item_id: item.evidence_item_id ?? '',
			activity_code: item.activity_code,
			activity_name: item.activity_name,
			output_indicator: item.output_indicator,
			target_volume: item.target_volume,
			target_unit: item.target_unit,
			budget_source: item.budget_source,
			budget_amount: item.budget_amount,
			realization_amount: item.realization_amount,
			status: item.status,
			progress_percent: item.progress_percent,
			start_date: dateValue(item.start_date),
			end_date: dateValue(item.end_date),
			evidence_url: item.evidence_url,
			notes: item.notes,
		};
		dialogOpen = true;
	}

	function openEditPerformanceTarget(target: PerformanceTargetRow) {
		editingId = target.id;
		dialogKind = 'performance';
		performanceTargetForm = {
			period_year: target.period_year,
			employee_id: target.employee_id,
			position_id: target.position_id ?? '',
			program_id: target.program_id ?? '',
			parent_target_id: target.parent_target_id ?? '',
			aspect: target.aspect,
			title: target.title,
			indicator: target.indicator,
			target_value: target.target_value,
			target_unit: target.target_unit,
			status: target.status,
			progress_percent: target.progress_percent,
			evidence_url: target.evidence_url,
			review_notes: target.review_notes,
			due_date: dateValue(target.due_date),
		};
		dialogOpen = true;
	}

	function openEditEvidenceItem(item: EvidenceItemRow) {
		editingId = item.id;
		dialogKind = 'evidence';
		evidenceItemForm = {
			period_year: item.period_year,
			title: item.title,
			evidence_type: item.evidence_type,
			snp_standard: item.snp_standard,
			owner_unit_id: item.owner_unit_id ?? '',
			document_id: item.document_id ?? '',
			program_id: item.program_id ?? '',
			performance_target_id: item.performance_target_id ?? '',
			source_module: item.source_module,
			evidence_url: item.evidence_url,
			status: item.status,
			notes: item.notes,
		};
		dialogOpen = true;
	}

	async function saveDialog() {
		if (!dialogKind) return;
		busy = true;
		try {
			const result = await submitCurrentDialog();
			if (!result) return;
			toast.success(editingId ? 'Data tata kelola diperbarui' : 'Data tata kelola ditambahkan');
			dialogOpen = false;
			await refreshGovernance();
		} finally {
			busy = false;
		}
	}

	async function submitCurrentDialog() {
		if (dialogKind === 'unit') return submitEntity('/api/governance/units', unitForm);
		if (dialogKind === 'position') return submitEntity('/api/governance/positions', positionForm);
		if (dialogKind === 'assignment') return submitEntity('/api/governance/assignments', assignmentForm);
		if (dialogKind === 'document') return submitEntity('/api/governance/documents', documentForm);
		if (dialogKind === 'program') return submitEntity('/api/governance/programs', programForm);
		if (dialogKind === 'workplan') return submitEntity('/api/governance/work-plan-items', workPlanItemForm);
		if (dialogKind === 'performance') return submitEntity('/api/governance/performance-targets', performanceTargetForm);
		if (dialogKind === 'evidence') return submitEntity('/api/governance/evidence-items', evidenceItemForm);
		return false;
	}

	async function submitEntity(
		path: string,
		payload: UnitForm | PositionForm | AssignmentForm | DocumentForm | ProgramForm | WorkPlanItemForm | PerformanceTargetForm | EvidenceItemForm
	) {
		const res = await fetch(editingId ? `${path}/${editingId}` : path, {
			method: editingId ? 'PUT' : 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(payload),
		});
		try {
			await readClientJson<unknown>(res);
		} catch (error) {
			toast.error(mutationErrorMessage(error, 'Gagal menyimpan data tata kelola'));
			return false;
		}
		return true;
	}

	async function deleteEntity(kind: string, id: string) {
		if (!(await confirmAction({
			title: 'Hapus Data Tata Kelola',
			message: 'Hapus data tata kelola ini?',
			confirmLabel: 'Hapus Data',
			tone: 'danger'
		}))) return;
		const res = await fetch(`/api/governance/${kind}/${id}`, { method: 'DELETE' });
		try {
			await readClientJson<unknown>(res);
		} catch (error) {
			toast.error(mutationErrorMessage(error, 'Gagal menghapus data'));
			return;
		}
		toast.success('Data tata kelola dihapus');
		await refreshGovernance();
	}

	function docTypeLabel(value: string) {
		return DOC_TYPES.find(([key]) => key === value)?.[1] ?? value;
	}

	function snpLabel(value: string) {
		return SNP_STANDARDS.find(([key]) => key === value)?.[1] ?? 'Tidak dikaitkan';
	}

	function statusLabel(value: string) {
		return PROGRAM_STATUSES.find(([key]) => key === value)?.[1] ?? DOCUMENT_STATUSES.find(([key]) => key === value)?.[1] ?? value;
	}

	function performanceAspectLabel(value: string) {
		return PERFORMANCE_ASPECTS.find(([key]) => key === value)?.[1] ?? value;
	}

	function evidenceTypeLabel(value: string) {
		return EVIDENCE_TYPES.find(([key]) => key === value)?.[1] ?? value;
	}

	function evidenceStatusLabel(value: string) {
		return EVIDENCE_STATUSES.find(([key]) => key === value)?.[1] ?? value;
	}

	function evidenceReferenceLabel(item: EvidenceItemRow) {
		const references: string[] = [];
		if (item.document_title) references.push(`Dokumen: ${item.document_title}`);
		if (item.program_code || item.program_name) references.push(`Program: ${[item.program_code, item.program_name].filter(Boolean).join(' · ')}`);
		if (item.performance_target_title) references.push(`Target: ${item.performance_target_title}`);
		return references.length > 0 ? references.join(' | ') : 'Tidak dikaitkan';
	}

	function buildStrategicAlignmentRows(): StrategicAlignmentRow[] {
		const documentsByID = new Map(documents.map((document) => [document.id, document]));
		return programs
			.map((program) => {
				const sourceDocument = program.source_document_id ? documentsByID.get(program.source_document_id) : undefined;
				const linkedWorkPlan = workPlanItems.filter((item) => item.program_id === program.id);
				const linkedPerformance = performanceTargets.filter((target) => target.program_id === program.id);
				const linkedWorkPlanEvidenceIDs = new Set(linkedWorkPlan.map((item) => item.evidence_item_id ?? '').filter(Boolean));
				const linkedPerformanceIDs = new Set(linkedPerformance.map((target) => target.id));
				const linkedEvidence = evidenceItems.filter(
					(item) =>
						item.program_id === program.id ||
						linkedWorkPlanEvidenceIDs.has(item.id) ||
						(item.performance_target_id ? linkedPerformanceIDs.has(item.performance_target_id) : false)
				);
				const hasIKU = program.iku_code.trim() !== '';
				const hasSourceDocument = !!sourceDocument;
				const hasWorkPlan = linkedWorkPlan.length > 0;
				const hasPerformance = linkedPerformance.length > 0;
				const hasEvidence =
					program.evidence_url.trim() !== '' ||
					linkedWorkPlan.some((item) => item.evidence_url.trim() !== '' || item.evidence_item_title) ||
					linkedPerformance.some((target) => target.evidence_url.trim() !== '') ||
					linkedEvidence.length > 0;
				const blocked = program.status === 'blocked' || linkedWorkPlan.some((item) => item.status === 'blocked') || linkedPerformance.some((target) => target.status === 'blocked');
				const gapNotes: string[] = [];
				if (!hasIKU) gapNotes.push('IKU belum diisi');
				if (!hasSourceDocument) gapNotes.push('Dokumen sumber belum dipilih');
				if (!hasWorkPlan) gapNotes.push('Belum turun ke RKT/RKJM');
				if (!hasPerformance) gapNotes.push('Belum turun ke SKP');
				if (!hasEvidence) gapNotes.push('Bukti belum tersedia');
				if (blocked) gapNotes.push('Ada status terkendala');
				const readinessScore = [hasIKU, hasSourceDocument, hasWorkPlan, hasPerformance, hasEvidence].filter(Boolean).length;
				const progressValues = [
					program.progress_percent,
					...linkedWorkPlan.map((item) => item.progress_percent),
					...linkedPerformance.map((target) => target.progress_percent),
				];
				const progressPercent = Math.round(progressValues.reduce((total, value) => total + Number(value || 0), 0) / progressValues.length);
				return {
					key: program.id,
					period_year: program.period_year,
					program_code: program.code,
					program_name: program.name,
					iku_code: program.iku_code,
					snp_standard: program.snp_standard,
					source_document_title: sourceDocument?.title ?? program.source_document_title ?? '',
					source_document_type: sourceDocument?.doc_type ?? '',
					owner_label: program.responsible_employee_name || program.responsible_position_title || program.owner_unit_name || 'Belum ditetapkan',
					work_plan_count: linkedWorkPlan.length,
					work_plan_done: linkedWorkPlan.filter((item) => item.status === 'done').length,
					work_plan_blocked: linkedWorkPlan.filter((item) => item.status === 'blocked').length,
					performance_count: linkedPerformance.length,
					performance_done: linkedPerformance.filter((target) => target.status === 'done').length,
					evidence_count: linkedEvidence.length + (program.evidence_url.trim() ? 1 : 0),
					verified_evidence_count: linkedEvidence.filter((item) => item.status === 'verified').length,
					progress_percent: progressPercent,
					readiness_status: blocked ? 'terkendala' : readinessScore === 5 ? 'lengkap' : readinessScore >= 3 ? 'parsial' : 'perlu_pemetaan',
					gap_notes: gapNotes,
				};
			})
			.sort((a, b) => b.period_year - a.period_year || alignmentStatusRank(a.readiness_status) - alignmentStatusRank(b.readiness_status) || a.program_code.localeCompare(b.program_code));
	}

	function alignmentStatusRank(value: string) {
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

	function workPlanTotals(items: WorkPlanItemRow[]) {
		return items.reduce(
			(total, item) => ({
				budget: total.budget + Number(item.budget_amount || 0),
				realization: total.realization + Number(item.realization_amount || 0),
				done: total.done + (item.status === 'done' ? 1 : 0),
				blocked: total.blocked + (item.status === 'blocked' ? 1 : 0),
			}),
			{ budget: 0, realization: 0, done: 0, blocked: 0 }
		);
	}

	function formatCurrency(value: number) {
		return new Intl.NumberFormat('id-ID', {
			style: 'currency',
			currency: 'IDR',
			maximumFractionDigits: 0,
		}).format(Number(value || 0));
	}

	function csvEscape(value: string | number | undefined) {
		const text = String(value ?? '');
		if (/[",\n]/.test(text)) return `"${text.replaceAll('"', '""')}"`;
		return text;
	}

	function exportWorkPlanCsv() {
		const rows = [
			['Tahun', 'Program', 'Kode Kegiatan', 'Nama Kegiatan', 'Indikator Output', 'Target', 'Sumber Anggaran', 'Anggaran', 'Realisasi', 'Status', 'Progres', 'Penanggung Jawab', 'Bukti'],
			...filteredWorkPlanItems.map((item) => [
				item.period_year,
				`${item.program_code} ${item.program_name}`,
				item.activity_code,
				item.activity_name,
				item.output_indicator,
				`${item.target_volume} ${item.target_unit}`.trim(),
				item.budget_source,
				item.budget_amount,
				item.realization_amount,
				statusLabel(item.status),
				`${item.progress_percent}%`,
				item.responsible_employee_name ?? '',
				item.evidence_url || item.evidence_item_title || '',
			]),
		];
		const csv = rows.map((row) => row.map(csvEscape).join(',')).join('\n');
		const blob = new Blob([csv], { type: 'text/csv;charset=utf-8' });
		const url = URL.createObjectURL(blob);
		const link = document.createElement('a');
		link.href = url;
		link.download = `rkt-rkjm-${new Date().toISOString().slice(0, 10)}.csv`;
		link.click();
		URL.revokeObjectURL(url);
	}

	function exportAlignmentCsv() {
		const rows = [
			['Tahun', 'Program', 'IKU', 'Dokumen Sumber', 'Jenis Dokumen', '8 SNP', 'Penanggung Jawab', 'RKT', 'SKP', 'Bukti', 'Progres', 'Status', 'Gap'],
			...filteredAlignmentRows.map((row) => [
				row.period_year,
				`${row.program_code} ${row.program_name}`,
				row.iku_code || 'Belum diisi',
				row.source_document_title || 'Belum dipilih',
				row.source_document_type ? docTypeLabel(row.source_document_type) : '',
				snpLabel(row.snp_standard),
				row.owner_label,
				`${row.work_plan_done}/${row.work_plan_count}`,
				`${row.performance_done}/${row.performance_count}`,
				`${row.verified_evidence_count}/${row.evidence_count}`,
				`${row.progress_percent}%`,
				alignmentStatusLabel(row.readiness_status),
				row.gap_notes.join('; '),
			]),
		];
		const csv = rows.map((row) => row.map(csvEscape).join(',')).join('\n');
		const blob = new Blob([csv], { type: 'text/csv;charset=utf-8' });
		const url = URL.createObjectURL(blob);
		const link = document.createElement('a');
		link.href = url;
		link.download = `peta-renstra-iku-${new Date().toISOString().slice(0, 10)}.csv`;
		link.click();
		URL.revokeObjectURL(url);
	}

	function openEvidenceUrl(url: string) {
		const trimmed = url.trim();
		if (!trimmed) return;
		window.open(trimmed, '_blank', 'noopener,noreferrer');
	}

	function dateValue(value?: string) {
		if (!value) return '';
		return value.slice(0, 10);
	}

	function formatDate(value?: string) {
		if (!value) return '-';
		return new Date(value).toLocaleDateString('id-ID', { timeZone: 'Asia/Makassar', day: '2-digit', month: 'short', year: 'numeric' });
	}

	onMount(() => {
		void loadGovernance();
	});
</script>

<svelte:head><title>Tata Kelola Madrasah — MTSN 2 Kolaka Utara</title></svelte:head>

<div class="space-y-6">
	<div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
		<div>
			<h1 class="text-lg font-semibold text-foreground">Tata Kelola Madrasah</h1>
			<p class="text-sm text-muted-foreground">Struktur organisasi, dokumen perencanaan, indikator kinerja, dan bukti 8 SNP.</p>
		</div>
		<div class="flex flex-wrap gap-2">
			<Button size="sm" onclick={() => openCreate('unit')}>Tambah Unit</Button>
			<Button size="sm" variant="outline" onclick={() => openCreate('document')}>Tambah Dokumen</Button>
			<Button size="sm" variant="outline" onclick={() => openCreate('program')}>Tambah Program</Button>
			<Button size="sm" variant="outline" onclick={() => openCreate('workplan')}>Tambah RKT</Button>
			<Button size="sm" variant="outline" onclick={() => openCreate('performance')}>Tambah Target</Button>
			<Button size="sm" variant="outline" onclick={() => openCreate('evidence')}>Tambah Bukti</Button>
			<Button href={`${resolve('/document-cycles')}?domain_area=governance&tab=connections`} size="sm" variant="outline">
				<FileTextIcon class="mr-2 size-4" />
				Siklus Dokumen
			</Button>
			<Button href={resolve('/governance/print-pack')} size="sm" variant="outline">
				<PrinterIcon class="mr-2 size-4" />
				Paket Cetak
			</Button>
			<Button href={resolve('/governance/actions')} size="sm" variant="outline">
				<ClipboardCheckIcon class="mr-2 size-4" />
				Tindak Lanjut
			</Button>
		</div>
	</div>

	<AsyncContent promise={governancePromise} onerror={handleGovernanceRenderError}>
		{#snippet pending()}
			<div class="space-y-4">
				<div class="grid grid-cols-2 gap-3 lg:grid-cols-4">
					{#each Array.from({ length: 8 }) as _, index (`governance-stat-skeleton-${index}`)}
						<Card.Root class="border-border">
							<Card.Content class="space-y-3 p-4">
								<Skeleton class="h-3 w-24" />
								<Skeleton class="h-8 w-16" />
							</Card.Content>
						</Card.Root>
					{/each}
				</div>
				<div class="rounded-2xl border border-border bg-card p-5">
					<div class="mb-5 flex flex-wrap gap-2">
						{#each Array.from({ length: 8 }) as _, index (`governance-tab-skeleton-${index}`)}
							<Skeleton class="h-9 w-28 rounded-full" />
						{/each}
					</div>
					<div class="grid gap-4 xl:grid-cols-2">
						<div class="space-y-3">
							<Skeleton class="h-10 w-full" />
							{#each Array.from({ length: 5 }) as _, index (`governance-left-skeleton-${index}`)}
								<Skeleton class="h-12 w-full" />
							{/each}
						</div>
						<div class="space-y-3">
							<Skeleton class="h-10 w-full" />
							{#each Array.from({ length: 5 }) as _, index (`governance-right-skeleton-${index}`)}
								<Skeleton class="h-12 w-full" />
							{/each}
						</div>
					</div>
				</div>
			</div>
		{/snippet}
		{#snippet failed(error, reset)}
			<RecoveryPanel
				title="Data Tata Kelola Belum Tersaji"
				message={governanceErrorMessage(error)}
				onRetry={() => retryGovernance(reset)}
			/>
		{/snippet}
		{#snippet children(value)}
			{@const overview = value as GovernanceOverview}
			{@const loadedStats = overview.stats}

	<div class="grid grid-cols-2 gap-3 lg:grid-cols-4">
		{#each [
			{ label: 'Unit Aktif', value: loadedStats.total_units },
			{ label: 'Jabatan', value: loadedStats.total_positions },
			{ label: 'Pejabat Aktif', value: loadedStats.active_assignments },
			{ label: 'Dokumen Final', value: loadedStats.final_documents },
			{ label: 'Program Aktif', value: loadedStats.active_programs },
			{ label: 'Terkendala', value: loadedStats.blocked_programs },
			{ label: 'Selesai', value: loadedStats.completed_programs },
			{ label: 'Ada Bukti', value: loadedStats.programs_with_evidence },
			{ label: 'Target Pegawai', value: loadedStats.performance_targets },
			{ label: 'Target Selesai', value: loadedStats.completed_performance_targets },
			{ label: 'Target Terkendala', value: loadedStats.blocked_performance_targets },
			{ label: 'Bukti Mutu', value: loadedStats.evidence_items },
			{ label: 'Bukti Valid', value: loadedStats.verified_evidence_items },
			{ label: 'Gap Bukti', value: loadedStats.gap_evidence_items },
			{ label: 'Item RKT', value: loadedStats.work_plan_items },
			{ label: 'RKT Selesai', value: loadedStats.completed_work_plan_items },
			{ label: 'RKT Terkendala', value: loadedStats.blocked_work_plan_items },
			{ label: 'Tindak Lanjut', value: loadedStats.compliance_actions },
			{ label: 'TL Terbuka', value: loadedStats.open_compliance_actions },
			{ label: 'TL Kritis', value: loadedStats.critical_compliance_actions },
			{ label: 'TL Selesai', value: loadedStats.completed_compliance_actions },
		] as item (item.label)}
			<Card.Root class="border-border">
				<Card.Content class="p-4">
					<p class="text-xs text-muted-foreground">{item.label}</p>
					<p class="mt-1 text-2xl font-bold text-primary">{item.value}</p>
				</Card.Content>
			</Card.Root>
		{/each}
	</div>

	<Card.Root class="border-primary/20 bg-primary/10">
		<Card.Content class="flex flex-col gap-3 p-4 lg:flex-row lg:items-center lg:justify-between">
			<div>
				<p class="text-sm font-semibold text-primary">Kendali Tindak Lanjut Kepatuhan</p>
				<p class="text-sm text-primary">
					{loadedStats.open_compliance_actions} terbuka, {loadedStats.critical_compliance_actions} prioritas tinggi/mendesak, {loadedStats.completed_compliance_actions} selesai.
				</p>
			</div>
			<div class="flex flex-wrap gap-2">
				<Button href={resolve('/governance/actions')} size="sm">
					<ClipboardCheckIcon class="mr-2 size-4" />
					Register
				</Button>
				<Button href={resolve('/governance/actions/calendar')} size="sm" variant="outline">
					<CalendarDaysIcon class="mr-2 size-4" />
					Kalender
				</Button>
				<Button href={resolve('/governance/actions/meeting-pack')} size="sm" variant="outline">
					<FileTextIcon class="mr-2 size-4" />
					Paket Rapat
				</Button>
				<Button href={resolve('/governance/print-pack')} size="sm" variant="outline">
					<PrinterIcon class="mr-2 size-4" />
					Paket Cetak
				</Button>
			</div>
		</Card.Content>
	</Card.Root>

	<Tabs.Root bind:value={activeTab}>
		<Tabs.List class="w-full overflow-x-auto">
			<Tabs.Trigger value="organisasi">Organisasi</Tabs.Trigger>
			<Tabs.Trigger value="dokumen">Dokumen</Tabs.Trigger>
			<Tabs.Trigger value="alignment">Peta IKU</Tabs.Trigger>
			<Tabs.Trigger value="program">Program</Tabs.Trigger>
			<Tabs.Trigger value="rkt">RKT/RKJM</Tabs.Trigger>
			<Tabs.Trigger value="kinerja">SKP/Kinerja</Tabs.Trigger>
			<Tabs.Trigger value="bukti">Bukti Mutu</Tabs.Trigger>
			<Tabs.Trigger value="snp">8 SNP</Tabs.Trigger>
		</Tabs.List>

		<Tabs.Content value="organisasi" class="space-y-4">
			<div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
				<Input placeholder="Cari unit, jabatan, atau pegawai" bind:value={orgSearch} class="sm:max-w-sm" />
				<div class="flex gap-2">
					<Button size="sm" variant="outline" onclick={() => openCreate('position')}>Tambah Jabatan</Button>
					<Button size="sm" variant="outline" onclick={() => openCreate('assignment')}>Tetapkan Pejabat</Button>
				</div>
			</div>

			<div class="grid gap-4 xl:grid-cols-2">
				<Card.Root class="border-border">
					<Card.Header class="pb-2">
						<Card.Title class="text-sm font-medium text-foreground">Unit Kerja</Card.Title>
					</Card.Header>
					<Card.Content class="p-0">
						{#if filteredUnits.length === 0}
							<div class="p-4"><EmptyStatePanel compact title="Belum ada unit kerja" description="Tambahkan unit pertama untuk memulai struktur organisasi madrasah." /></div>
						{:else}
							<Table.Root>
								<Table.Header><Table.Row><Table.Head>Unit</Table.Head><Table.Head>Induk</Table.Head><Table.Head class="w-28">Aksi</Table.Head></Table.Row></Table.Header>
								<Table.Body>
									{#each filteredUnits as unit (unit.id)}
										<Table.Row>
											<Table.Cell>
												<p class="text-sm font-medium text-foreground">{unit.name}</p>
												<p class="text-xs text-muted-foreground">{unit.code} · {unit.unit_type}</p>
											</Table.Cell>
											<Table.Cell class="text-sm text-muted-foreground">{unit.parent_unit_name || '-'}</Table.Cell>
											<Table.Cell>
												<div class="flex gap-1">
													<Button size="sm" variant="outline" onclick={() => openEditUnit(unit)}>Edit</Button>
													<Button size="sm" variant="destructive" onclick={() => deleteEntity('units', unit.id)}>Hapus</Button>
												</div>
											</Table.Cell>
										</Table.Row>
									{/each}
								</Table.Body>
							</Table.Root>
						{/if}
					</Card.Content>
				</Card.Root>

				<Card.Root class="border-border">
					<Card.Header class="pb-2">
						<Card.Title class="text-sm font-medium text-foreground">Jabatan dan Jalur Komando</Card.Title>
					</Card.Header>
					<Card.Content class="p-0">
						{#if filteredPositions.length === 0}
							<div class="p-4"><EmptyStatePanel compact title="Belum ada jabatan" description="Tambahkan jabatan untuk memetakan komando dan uraian tugas." /></div>
						{:else}
							<Table.Root>
								<Table.Header><Table.Row><Table.Head>Jabatan</Table.Head><Table.Head>Pejabat Aktif</Table.Head><Table.Head class="w-28">Aksi</Table.Head></Table.Row></Table.Header>
								<Table.Body>
									{#each filteredPositions as position (position.id)}
										<Table.Row>
											<Table.Cell>
												<p class="text-sm font-medium text-foreground">{position.title}</p>
												<p class="text-xs text-muted-foreground">{position.unit_name}{position.parent_position_title ? ` · Atasan: ${position.parent_position_title}` : ''}</p>
											</Table.Cell>
											<Table.Cell class="text-sm text-muted-foreground">
												{position.active_employee_name || '-'}
												{#if position.active_employee_nip}<span class="block text-xs text-muted-foreground">{position.active_employee_nip}</span>{/if}
											</Table.Cell>
											<Table.Cell>
												<div class="flex gap-1">
													<Button size="sm" variant="outline" onclick={() => openEditPosition(position)}>Edit</Button>
													<Button size="sm" variant="destructive" onclick={() => deleteEntity('positions', position.id)}>Hapus</Button>
												</div>
											</Table.Cell>
										</Table.Row>
									{/each}
								</Table.Body>
							</Table.Root>
						{/if}
					</Card.Content>
				</Card.Root>
			</div>

			<Card.Root class="border-border">
				<Card.Header class="pb-2">
					<Card.Title class="text-sm font-medium text-foreground">Pejabat Aktif</Card.Title>
				</Card.Header>
				<Card.Content class="p-0">
					{#if assignments.length === 0}
						<div class="p-4"><EmptyStatePanel compact title="Belum ada pejabat aktif" description="Tetapkan pegawai pada jabatan untuk membangun struktur komando." /></div>
					{:else}
						<Table.Root>
							<Table.Header><Table.Row><Table.Head>Jabatan</Table.Head><Table.Head>Pegawai</Table.Head><Table.Head>Masa Tugas</Table.Head><Table.Head class="w-28">Aksi</Table.Head></Table.Row></Table.Header>
							<Table.Body>
								{#each assignments as assignment (assignment.id)}
									<Table.Row>
										<Table.Cell><p class="text-sm font-medium">{assignment.position_title}</p><p class="text-xs text-muted-foreground">{assignment.unit_name}</p></Table.Cell>
										<Table.Cell><p class="text-sm">{assignment.employee_name}</p><p class="text-xs text-muted-foreground">{assignment.employee_nip}</p></Table.Cell>
										<Table.Cell class="text-sm">{formatDate(assignment.start_date)} - {assignment.end_date ? formatDate(assignment.end_date) : 'Aktif'}</Table.Cell>
										<Table.Cell>
											<div class="flex gap-1">
												<Button size="sm" variant="outline" onclick={() => openEditAssignment(assignment)}>Edit</Button>
												<Button size="sm" variant="destructive" onclick={() => deleteEntity('assignments', assignment.id)}>Hapus</Button>
											</div>
										</Table.Cell>
									</Table.Row>
								{/each}
							</Table.Body>
						</Table.Root>
					{/if}
				</Card.Content>
			</Card.Root>
		</Tabs.Content>

		<Tabs.Content value="dokumen" class="space-y-4">
			<div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
				<Input placeholder="Cari dokumen, jenis, unit, atau ringkasan" bind:value={documentSearch} class="sm:max-w-sm" />
				<Button size="sm" onclick={() => openCreate('document')}>Tambah Dokumen</Button>
			</div>
			<Card.Root class="border-border">
				<Card.Content class="p-0">
					{#if filteredDocuments.length === 0}
						<div class="p-4"><EmptyStatePanel compact title="Belum ada dokumen tata kelola" description="Catat visi misi, RKJM, RKT, Renstra, Perkin, IKU, SK, SOP, dan bukti 8 SNP." /></div>
					{:else}
						<Table.Root>
							<Table.Header><Table.Row><Table.Head>Dokumen</Table.Head><Table.Head>Periode</Table.Head><Table.Head>Status</Table.Head><Table.Head>8 SNP</Table.Head><Table.Head class="w-28">Aksi</Table.Head></Table.Row></Table.Header>
							<Table.Body>
								{#each filteredDocuments as document (document.id)}
									<Table.Row>
										<Table.Cell><p class="text-sm font-medium">{document.title}</p><p class="text-xs text-muted-foreground">{docTypeLabel(document.doc_type)} · {document.owner_unit_name || 'Tanpa unit'}</p></Table.Cell>
										<Table.Cell class="text-sm">{document.period_year}{document.period_label ? ` · ${document.period_label}` : ''}</Table.Cell>
										<Table.Cell><Badge variant={document.status === 'final' ? 'default' : 'outline'}>{statusLabel(document.status)}</Badge></Table.Cell>
										<Table.Cell class="text-sm">{snpLabel(document.snp_standard)}</Table.Cell>
										<Table.Cell>
											<div class="flex gap-1">
												<Button size="sm" variant="outline" onclick={() => openEditDocument(document)}>Edit</Button>
												<Button size="sm" variant="destructive" onclick={() => deleteEntity('documents', document.id)}>Hapus</Button>
											</div>
										</Table.Cell>
									</Table.Row>
								{/each}
							</Table.Body>
						</Table.Root>
					{/if}
				</Card.Content>
			</Card.Root>
		</Tabs.Content>

		<Tabs.Content value="alignment" class="space-y-4">
			<div class="flex flex-col gap-2 lg:flex-row lg:items-center lg:justify-between">
				<Input placeholder="Cari program, IKU, dokumen, penanggung jawab, atau gap" bind:value={alignmentSearch} class="lg:max-w-md" />
				<div class="flex flex-wrap gap-2">
					<Button size="sm" variant="outline" onclick={exportAlignmentCsv}>Export CSV</Button>
					<Button size="sm" variant="outline" onclick={() => window.print()}>Print</Button>
					<Button size="sm" variant="outline" onclick={() => openCreate('document')}>Tambah Dokumen</Button>
					<Button size="sm" onclick={() => openCreate('program')}>Tambah Program</Button>
				</div>
			</div>

			<div class="grid gap-3 md:grid-cols-3 xl:grid-cols-6">
				{#each [
					{ label: 'Program', value: alignmentSummary.total_programs, note: 'Basis pemetaan' },
					{ label: 'Dokumen Final', value: alignmentSummary.final_strategic_documents, note: 'Visi/Renstra/Perkin/IKU' },
					{ label: 'IKU Terisi', value: alignmentSummary.mapped_iku, note: 'Program berkode IKU' },
					{ label: 'Lengkap', value: alignmentSummary.complete, note: 'RKT, SKP, bukti ada' },
					{ label: 'Gap', value: alignmentSummary.gaps, note: 'Perlu tindak lanjut' },
					{ label: 'Terkendala', value: alignmentSummary.blocked, note: 'Status blocked' },
				] as item (item.label)}
					<Card.Root class="border-border">
						<Card.Content class="p-4">
							<p class="text-xs text-muted-foreground">{item.label}</p>
							<p class="mt-1 text-2xl font-semibold text-foreground">{item.value}</p>
							<p class="text-xs text-muted-foreground">{item.note}</p>
						</Card.Content>
					</Card.Root>
				{/each}
			</div>

			<div class="grid gap-4 xl:grid-cols-[420px_1fr]">
				<Card.Root class="border-border">
					<Card.Header class="pb-2">
						<Card.Title class="text-sm font-medium text-foreground">Dokumen Strategis</Card.Title>
						<Card.Description>Visi/Misi, Renstra/RKJM, RKT, Perkin, dan IKU yang menjadi sumber pemetaan.</Card.Description>
					</Card.Header>
					<Card.Content class="space-y-2">
						{#if strategicDocuments.length === 0}
							<EmptyStatePanel compact title="Belum ada dokumen strategis" description="Tambahkan dokumen visi/misi, Renstra, RKJM, RKT, Perkin, atau IKU sebagai sumber peta." />
						{:else}
							{#each strategicDocuments.slice(0, 8) as document (document.id)}
								<div class="rounded-md border border-border px-3 py-2">
									<div class="flex items-start justify-between gap-3">
										<div>
											<p class="text-sm font-medium text-foreground">{document.title}</p>
											<p class="text-xs text-muted-foreground">{docTypeLabel(document.doc_type)} · {document.period_year}{document.period_label ? ` · ${document.period_label}` : ''}</p>
										</div>
										<Badge variant={document.status === 'final' ? 'default' : 'outline'}>{statusLabel(document.status)}</Badge>
									</div>
								</div>
							{/each}
						{/if}
					</Card.Content>
				</Card.Root>

				<Card.Root class="border-border">
					<Card.Header class="pb-2">
						<Card.Title class="text-sm font-medium text-foreground">Gap Pemetaan</Card.Title>
						<Card.Description>Program yang belum lengkap dari IKU sampai bukti mutu.</Card.Description>
					</Card.Header>
					<Card.Content class="space-y-2">
						{@const gapRows = filteredAlignmentRows.filter((row) => row.gap_notes.length > 0).slice(0, 8)}
							{#if gapRows.length === 0}
								<div class="rounded-md border border-primary/20 bg-primary/10 px-4 py-3 text-sm text-primary">Tidak ada gap utama pada filter saat ini.</div>
							{:else}
								{#each gapRows as row (row.key)}
									<div class="rounded-md border border-border px-3 py-2">
										<div class="flex items-start justify-between gap-3">
											<div>
												<p class="text-sm font-medium text-foreground">{row.program_code} · {row.program_name}</p>
												<p class="mt-1 text-xs text-muted-foreground">{row.gap_notes.join(' · ')}</p>
											</div>
											<Badge variant={row.readiness_status === 'terkendala' ? 'destructive' : 'outline'}>{alignmentStatusLabel(row.readiness_status)}</Badge>
										</div>
									</div>
								{/each}
							{/if}
					</Card.Content>
				</Card.Root>
			</div>

			<Card.Root class="border-border">
				<Card.Header class="pb-2">
					<Card.Title class="text-sm font-medium text-foreground">Peta Renstra/IKU ke RKT, SKP, dan Bukti</Card.Title>
					<Card.Description>Cascading internal sebelum data dirapikan ke administrasi Perkin, IKU, dan SIPKA.</Card.Description>
				</Card.Header>
				<Card.Content class="p-0">
					{#if filteredAlignmentRows.length === 0}
						<div class="p-4"><EmptyStatePanel compact title="Belum ada peta IKU" description="Tambahkan program indikator dan kaitkan ke dokumen sumber, RKT/RKJM, SKP, dan bukti mutu." /></div>
					{:else}
						<div class="overflow-x-auto">
							<Table.Root>
								<Table.Header>
									<Table.Row>
										<Table.Head>Program/IKU</Table.Head>
										<Table.Head>Dokumen</Table.Head>
										<Table.Head>Penanggung Jawab</Table.Head>
										<Table.Head>RKT</Table.Head>
										<Table.Head>SKP</Table.Head>
										<Table.Head>Bukti</Table.Head>
										<Table.Head>Status</Table.Head>
									</Table.Row>
								</Table.Header>
								<Table.Body>
									{#each filteredAlignmentRows as row (row.key)}
										<Table.Row>
											<Table.Cell class="min-w-[260px]">
												<p class="text-sm font-medium text-foreground">{row.program_code} · {row.program_name}</p>
												<p class="text-xs text-muted-foreground">{row.period_year} · {row.iku_code || 'IKU belum diisi'} · {snpLabel(row.snp_standard)}</p>
												{#if row.gap_notes.length > 0}
													<p class="mt-1 text-xs text-muted-foreground">{row.gap_notes.join(' · ')}</p>
												{/if}
											</Table.Cell>
											<Table.Cell class="min-w-[220px]">
												<p class="text-sm">{row.source_document_title || '-'}</p>
												<p class="text-xs text-muted-foreground">{row.source_document_type ? docTypeLabel(row.source_document_type) : 'Belum dikaitkan dokumen'}</p>
											</Table.Cell>
											<Table.Cell class="min-w-[180px] text-sm">{row.owner_label}</Table.Cell>
											<Table.Cell class="whitespace-nowrap text-sm">{row.work_plan_done}/{row.work_plan_count} selesai</Table.Cell>
											<Table.Cell class="whitespace-nowrap text-sm">{row.performance_done}/{row.performance_count} selesai</Table.Cell>
											<Table.Cell class="whitespace-nowrap text-sm">{row.verified_evidence_count}/{row.evidence_count} valid</Table.Cell>
											<Table.Cell>
												<div class="flex items-center gap-2">
													<div class="h-2 w-20 rounded-full bg-muted"><div class="h-2 rounded-full bg-primary" style={`width: ${row.progress_percent}%`}></div></div>
													<span class="text-xs text-muted-foreground">{row.progress_percent}%</span>
												</div>
												<Badge variant={row.readiness_status === 'terkendala' ? 'destructive' : row.readiness_status === 'lengkap' ? 'default' : 'outline'}>{alignmentStatusLabel(row.readiness_status)}</Badge>
											</Table.Cell>
										</Table.Row>
									{/each}
								</Table.Body>
							</Table.Root>
						</div>
					{/if}
				</Card.Content>
			</Card.Root>
		</Tabs.Content>

		<Tabs.Content value="program" class="space-y-4">
			<div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
				<Input placeholder="Cari program, IKU, indikator, atau unit" bind:value={programSearch} class="sm:max-w-sm" />
				<Button size="sm" onclick={() => openCreate('program')}>Tambah Program</Button>
			</div>
			<Card.Root class="border-border">
				<Card.Content class="p-0">
					{#if filteredPrograms.length === 0}
						<div class="p-4"><EmptyStatePanel compact title="Belum ada program indikator" description="Tambahkan program untuk memetakan RKT/RKJM ke penanggung jawab, target, dan bukti realisasi." /></div>
					{:else}
						<Table.Root>
							<Table.Header><Table.Row><Table.Head>Program</Table.Head><Table.Head>Penanggung Jawab</Table.Head><Table.Head>Target</Table.Head><Table.Head>Progres</Table.Head><Table.Head class="w-28">Aksi</Table.Head></Table.Row></Table.Header>
							<Table.Body>
								{#each filteredPrograms as program (program.id)}
									<Table.Row>
										<Table.Cell><p class="text-sm font-medium">{program.code} · {program.name}</p><p class="text-xs text-muted-foreground">{program.iku_code || 'Tanpa IKU'} · {snpLabel(program.snp_standard)}</p></Table.Cell>
										<Table.Cell><p class="text-sm">{program.responsible_employee_name || program.responsible_position_title || '-'}</p><p class="text-xs text-muted-foreground">{program.owner_unit_name || 'Tanpa unit'}</p></Table.Cell>
										<Table.Cell class="text-sm">{program.target_value || '-'} {program.target_unit}</Table.Cell>
										<Table.Cell>
											<div class="flex items-center gap-2">
												<div class="h-2 w-20 rounded-full bg-muted"><div class="h-2 rounded-full bg-primary" style={`width: ${program.progress_percent}%`}></div></div>
												<span class="text-xs text-muted-foreground">{program.progress_percent}%</span>
											</div>
											<Badge variant={program.status === 'blocked' ? 'destructive' : program.status === 'done' ? 'default' : 'outline'}>{statusLabel(program.status)}</Badge>
										</Table.Cell>
										<Table.Cell>
											<div class="flex gap-1">
												<Button size="sm" variant="outline" onclick={() => openEditProgram(program)}>Edit</Button>
												<Button size="sm" variant="destructive" onclick={() => deleteEntity('programs', program.id)}>Hapus</Button>
											</div>
										</Table.Cell>
									</Table.Row>
								{/each}
							</Table.Body>
						</Table.Root>
					{/if}
				</Card.Content>
			</Card.Root>
		</Tabs.Content>

		<Tabs.Content value="rkt" class="space-y-4">
			{@const totals = workPlanTotals(filteredWorkPlanItems)}
			<div class="flex flex-col gap-2 lg:flex-row lg:items-center lg:justify-between">
				<Input placeholder="Cari kegiatan, program, unit, pegawai, atau sumber anggaran" bind:value={workPlanSearch} class="lg:max-w-md" />
				<div class="flex flex-wrap gap-2">
					<Button size="sm" variant="outline" onclick={exportWorkPlanCsv}>Export CSV</Button>
					<Button size="sm" variant="outline" onclick={() => window.print()}>Print</Button>
					<Button size="sm" onclick={() => openCreate('workplan')}>Tambah Item RKT</Button>
				</div>
			</div>
			<div class="grid gap-3 md:grid-cols-4">
				<Card.Root class="border-border">
					<Card.Content class="p-4">
						<p class="text-xs text-muted-foreground">Item Tampil</p>
						<p class="mt-1 text-2xl font-semibold text-foreground">{filteredWorkPlanItems.length}</p>
						<p class="text-xs text-muted-foreground">{totals.done} selesai, {totals.blocked} terkendala</p>
					</Card.Content>
				</Card.Root>
				<Card.Root class="border-border">
					<Card.Content class="p-4">
						<p class="text-xs text-muted-foreground">Total Anggaran</p>
						<p class="mt-1 text-lg font-semibold text-foreground">{formatCurrency(totals.budget)}</p>
						<p class="text-xs text-muted-foreground">Berdasarkan filter aktif</p>
					</Card.Content>
				</Card.Root>
				<Card.Root class="border-border">
					<Card.Content class="p-4">
						<p class="text-xs text-muted-foreground">Realisasi</p>
						<p class="mt-1 text-lg font-semibold text-primary">{formatCurrency(totals.realization)}</p>
						<p class="text-xs text-muted-foreground">Serapan tercatat</p>
					</Card.Content>
				</Card.Root>
				<Card.Root class="border-border">
					<Card.Content class="p-4">
						<p class="text-xs text-muted-foreground">Sisa Anggaran</p>
						<p class="mt-1 text-lg font-semibold text-foreground">{formatCurrency(Math.max(totals.budget - totals.realization, 0))}</p>
						<p class="text-xs text-muted-foreground">Belum direalisasikan</p>
					</Card.Content>
				</Card.Root>
			</div>
			<Card.Root class="border-border">
				<Card.Header class="pb-2">
					<Card.Title class="text-sm font-medium text-foreground">Pelaksanaan RKT/RKJM</Card.Title>
					<Card.Description>Breakdown program ke kegiatan tahunan, jadwal, anggaran, realisasi, penanggung jawab, dan bukti.</Card.Description>
				</Card.Header>
				<Card.Content class="p-0">
					{#if filteredWorkPlanItems.length === 0}
						<div class="p-4"><EmptyStatePanel compact title="Belum ada item RKT/RKJM" description="Tambahkan kegiatan tahunan untuk menurunkan program menjadi jadwal, anggaran, dan bukti pelaksanaan." /></div>
					{:else}
						<div class="overflow-x-auto">
							<Table.Root>
								<Table.Header>
									<Table.Row>
										<Table.Head>Kegiatan</Table.Head>
										<Table.Head>Program</Table.Head>
										<Table.Head>Target</Table.Head>
										<Table.Head>Anggaran</Table.Head>
										<Table.Head>Progres</Table.Head>
										<Table.Head>Bukti</Table.Head>
										<Table.Head class="w-28">Aksi</Table.Head>
									</Table.Row>
								</Table.Header>
								<Table.Body>
									{#each filteredWorkPlanItems as item (item.id)}
										<Table.Row>
											<Table.Cell class="min-w-[260px]">
												<p class="text-sm font-medium text-foreground">{item.activity_code} · {item.activity_name}</p>
												<p class="text-xs text-muted-foreground">{item.period_year} · {item.owner_unit_name || 'Tanpa unit'} · {item.responsible_employee_name || 'Tanpa PJ'}</p>
												{#if item.notes}
													<p class="mt-1 text-xs text-muted-foreground">{item.notes}</p>
												{/if}
											</Table.Cell>
											<Table.Cell class="min-w-[220px]">
												<p class="text-sm text-foreground">{item.program_code}</p>
												<p class="text-xs text-muted-foreground">{item.program_name}</p>
											</Table.Cell>
											<Table.Cell class="min-w-[200px]">
												<p class="text-sm">{item.target_volume || '-'} {item.target_unit}</p>
												<p class="text-xs text-muted-foreground">{item.output_indicator || 'Tanpa indikator output'}</p>
												<p class="text-xs text-muted-foreground">{formatDate(item.start_date)} - {formatDate(item.end_date)}</p>
											</Table.Cell>
											<Table.Cell class="min-w-[170px]">
												<p class="text-sm">{formatCurrency(item.budget_amount)}</p>
												<p class="text-xs text-muted-foreground">Realisasi {formatCurrency(item.realization_amount)}</p>
												<p class="text-xs text-muted-foreground">{item.budget_source || 'Sumber belum diisi'}</p>
											</Table.Cell>
											<Table.Cell>
												<div class="flex items-center gap-2">
													<div class="h-2 w-20 rounded-full bg-muted"><div class="h-2 rounded-full bg-primary" style={`width: ${item.progress_percent}%`}></div></div>
													<span class="text-xs text-muted-foreground">{item.progress_percent}%</span>
												</div>
												<Badge variant={item.status === 'blocked' ? 'destructive' : item.status === 'done' ? 'default' : 'outline'}>{statusLabel(item.status)}</Badge>
											</Table.Cell>
											<Table.Cell>
												{#if item.evidence_url}
													<Button size="sm" variant="outline" onclick={() => openEvidenceUrl(item.evidence_url)}>Buka</Button>
												{:else if item.evidence_item_title}
													<p class="text-xs text-muted-foreground">{item.evidence_item_title}</p>
												{:else}
													<span class="text-muted-foreground">-</span>
												{/if}
											</Table.Cell>
											<Table.Cell>
												<div class="flex gap-1">
													<Button size="sm" variant="outline" onclick={() => openEditWorkPlanItem(item)}>Edit</Button>
													<Button size="sm" variant="destructive" onclick={() => deleteEntity('work-plan-items', item.id)}>Hapus</Button>
												</div>
											</Table.Cell>
										</Table.Row>
									{/each}
								</Table.Body>
							</Table.Root>
						</div>
					{/if}
				</Card.Content>
			</Card.Root>
		</Tabs.Content>

		<Tabs.Content value="kinerja" class="space-y-4">
			<div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
				<Input placeholder="Cari target, indikator, pegawai, atau program" bind:value={performanceSearch} class="sm:max-w-sm" />
				<Button size="sm" onclick={() => openCreate('performance')}>Tambah Target Kinerja</Button>
			</div>
			<Card.Root class="border-border">
				<Card.Header class="pb-2">
					<Card.Title class="text-sm font-medium text-foreground">SKP Mirror Internal</Card.Title>
					<Card.Description>Target kerja pegawai sebagai bahan pemetaan internal sebelum pengisian sistem resmi eksternal.</Card.Description>
				</Card.Header>
				<Card.Content class="p-0">
					{#if filteredPerformanceTargets.length === 0}
						<div class="p-4"><EmptyStatePanel compact title="Belum ada target kinerja" description="Tambahkan target kerja pegawai dari program/RKT/IKU agar cascading kinerja mulai terlihat." /></div>
					{:else}
						<Table.Root>
							<Table.Header>
								<Table.Row>
									<Table.Head>Target</Table.Head>
									<Table.Head>Pegawai</Table.Head>
									<Table.Head>Program</Table.Head>
									<Table.Head>Progres</Table.Head>
									<Table.Head class="w-28">Aksi</Table.Head>
								</Table.Row>
							</Table.Header>
							<Table.Body>
								{#each filteredPerformanceTargets as target (target.id)}
									<Table.Row>
										<Table.Cell>
											<p class="text-sm font-medium">{target.title}</p>
											<p class="text-xs text-muted-foreground">{performanceAspectLabel(target.aspect)} · {target.indicator || 'Tanpa indikator'}</p>
										</Table.Cell>
										<Table.Cell>
											<p class="text-sm">{target.employee_name}</p>
											<p class="text-xs text-muted-foreground">{target.employee_nip}{target.position_title ? ` · ${target.position_title}` : ''}</p>
										</Table.Cell>
										<Table.Cell>
											<p class="text-sm">{target.program_code || '-'}</p>
											<p class="text-xs text-muted-foreground">{target.program_name || 'Tidak dikaitkan program'}</p>
										</Table.Cell>
										<Table.Cell>
											<div class="flex items-center gap-2">
												<div class="h-2 w-20 rounded-full bg-muted"><div class="h-2 rounded-full bg-primary" style={`width: ${target.progress_percent}%`}></div></div>
												<span class="text-xs text-muted-foreground">{target.progress_percent}%</span>
											</div>
											<Badge variant={target.status === 'blocked' ? 'destructive' : target.status === 'done' ? 'default' : 'outline'}>{statusLabel(target.status)}</Badge>
										</Table.Cell>
										<Table.Cell>
											<div class="flex gap-1">
												<Button size="sm" variant="outline" onclick={() => openEditPerformanceTarget(target)}>Edit</Button>
												<Button size="sm" variant="destructive" onclick={() => deleteEntity('performance-targets', target.id)}>Hapus</Button>
											</div>
										</Table.Cell>
									</Table.Row>
								{/each}
							</Table.Body>
						</Table.Root>
					{/if}
				</Card.Content>
			</Card.Root>
		</Tabs.Content>

		<Tabs.Content value="bukti" class="space-y-4">
			<div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
				<Input placeholder="Cari bukti, status, unit, dokumen, program, atau target" bind:value={evidenceSearch} class="sm:max-w-sm" />
				<Button size="sm" onclick={() => openCreate('evidence')}>Tambah Bukti Mutu</Button>
			</div>
			<Card.Root class="border-border">
				<Card.Header class="pb-2">
					<Card.Title class="text-sm font-medium text-foreground">Register Bukti Mutu</Card.Title>
					<Card.Description>Daftar bukti pendukung 8 SNP yang terhubung ke dokumen, program, target kinerja, dan unit pemilik.</Card.Description>
				</Card.Header>
				<Card.Content class="p-0">
					{#if filteredEvidenceItems.length === 0}
						<div class="p-4"><EmptyStatePanel compact title="Belum ada bukti mutu" description="Tambahkan item bukti agar kesiapan 8 SNP tidak hanya bergantung pada tautan program atau target." /></div>
					{:else}
						<Table.Root>
							<Table.Header>
								<Table.Row>
									<Table.Head>Bukti</Table.Head>
									<Table.Head>Kaitan</Table.Head>
									<Table.Head>Status</Table.Head>
									<Table.Head>File</Table.Head>
									<Table.Head class="w-28">Aksi</Table.Head>
								</Table.Row>
							</Table.Header>
							<Table.Body>
								{#each filteredEvidenceItems as item (item.id)}
									<Table.Row>
										<Table.Cell>
											<p class="text-sm font-medium">{item.title}</p>
											<p class="text-xs text-muted-foreground">{item.period_year} · {evidenceTypeLabel(item.evidence_type)} · {snpLabel(item.snp_standard)}</p>
										</Table.Cell>
										<Table.Cell>
											<p class="text-sm">{evidenceReferenceLabel(item)}</p>
											<p class="text-xs text-muted-foreground">{item.owner_unit_name || 'Tanpa unit'} · {item.source_module}</p>
										</Table.Cell>
										<Table.Cell>
											<Badge variant={item.status === 'gap' ? 'destructive' : item.status === 'verified' ? 'default' : 'outline'}>{evidenceStatusLabel(item.status)}</Badge>
										</Table.Cell>
										<Table.Cell class="text-sm">
											{#if item.evidence_url}
												<Button size="sm" variant="outline" onclick={() => openEvidenceUrl(item.evidence_url)}>Buka</Button>
											{:else}
												<span class="text-muted-foreground">-</span>
											{/if}
										</Table.Cell>
										<Table.Cell>
											<div class="flex gap-1">
												<Button size="sm" variant="outline" onclick={() => openEditEvidenceItem(item)}>Edit</Button>
												<Button size="sm" variant="destructive" onclick={() => deleteEntity('evidence-items', item.id)}>Hapus</Button>
											</div>
										</Table.Cell>
									</Table.Row>
								{/each}
							</Table.Body>
						</Table.Root>
					{/if}
				</Card.Content>
			</Card.Root>
		</Tabs.Content>

		<Tabs.Content value="snp" class="space-y-4">
			<Card.Root class="border-border">
				<Card.Header class="pb-2">
					<Card.Title class="text-sm font-medium text-foreground">Matriks Bukti 8 SNP</Card.Title>
					<Card.Description>Ringkasan dokumen, program, dan bukti realisasi per standar nasional pendidikan.</Card.Description>
				</Card.Header>
				<Card.Content class="p-0">
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
							{#each snpMatrix as row (row.code)}
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
				</Card.Content>
			</Card.Root>
		</Tabs.Content>
	</Tabs.Root>
		{/snippet}
	</AsyncContent>
</div>

<Dialog.Root bind:open={dialogOpen}>
	<Dialog.Content>
		<Dialog.Header>
			<Dialog.Title>{editingId ? 'Edit Data Tata Kelola' : 'Tambah Data Tata Kelola'}</Dialog.Title>
			<Dialog.Description>Lengkapi data sesuai dokumen dan struktur resmi madrasah.</Dialog.Description>
		</Dialog.Header>

		<div class="mt-4 space-y-4">
			{#if dialogKind === 'unit'}
				<div class="grid gap-3 sm:grid-cols-2">
					<div><label for="unit-code" class="text-sm font-medium">Kode Unit</label><Input id="unit-code" bind:value={unitForm.code} /></div>
					<div><label for="unit-name" class="text-sm font-medium">Nama Unit</label><Input id="unit-name" bind:value={unitForm.name} /></div>
					<div><label for="unit-type" class="text-sm font-medium">Tipe Unit</label><Input id="unit-type" bind:value={unitForm.unit_type} /></div>
					<div>
						<label for="unit-parent" class="text-sm font-medium">Unit Induk</label>
						<select id="unit-parent" bind:value={unitForm.parent_id} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
							<option value="">Tanpa induk</option>
							{#each units.filter((unit) => unit.id !== editingId) as unit (unit.id)}<option value={unit.id}>{unit.name}</option>{/each}
						</select>
					</div>
				</div>
				<div><label for="unit-desc" class="text-sm font-medium">Deskripsi</label><Textarea id="unit-desc" bind:value={unitForm.description} /></div>
			{:else if dialogKind === 'position'}
				<div class="grid gap-3 sm:grid-cols-2">
					<div>
						<label for="position-unit" class="text-sm font-medium">Unit</label>
						<select id="position-unit" bind:value={positionForm.unit_id} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
							<option value="">Pilih unit</option>{#each units as unit (unit.id)}<option value={unit.id}>{unit.name}</option>{/each}
						</select>
					</div>
					<div><label for="position-title" class="text-sm font-medium">Nama Jabatan</label><Input id="position-title" bind:value={positionForm.title} /></div>
					<div><label for="position-type" class="text-sm font-medium">Tipe Jabatan</label><Input id="position-type" bind:value={positionForm.position_type} /></div>
					<div>
						<label for="position-parent" class="text-sm font-medium">Atasan Langsung</label>
						<select id="position-parent" bind:value={positionForm.parent_position_id} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
							<option value="">Tanpa atasan</option>{#each positions.filter((position) => position.id !== editingId) as position (position.id)}<option value={position.id}>{position.title}</option>{/each}
						</select>
					</div>
				</div>
				<div><label for="position-desc" class="text-sm font-medium">Deskripsi</label><Textarea id="position-desc" bind:value={positionForm.description} /></div>
				<div><label for="position-tupoksi" class="text-sm font-medium">Tupoksi</label><Textarea id="position-tupoksi" bind:value={positionForm.tupoksi} /></div>
			{:else if dialogKind === 'assignment'}
				<div class="grid gap-3 sm:grid-cols-2">
					<div>
						<label for="assignment-position" class="text-sm font-medium">Jabatan</label>
						<select id="assignment-position" bind:value={assignmentForm.position_id} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
							<option value="">Pilih jabatan</option>{#each positions as position (position.id)}<option value={position.id}>{position.title}</option>{/each}
						</select>
					</div>
					<div>
						<label for="assignment-employee" class="text-sm font-medium">Pegawai</label>
						<select id="assignment-employee" bind:value={assignmentForm.employee_id} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
							<option value="">Pilih pegawai</option>{#each employees as employee (employee.id)}<option value={employee.id}>{employee.nama} · {employee.nip}</option>{/each}
						</select>
					</div>
					<div><label for="assignment-start" class="text-sm font-medium">Tanggal Mulai</label><Input id="assignment-start" type="date" bind:value={assignmentForm.start_date} /></div>
					<div><label for="assignment-end" class="text-sm font-medium">Tanggal Selesai</label><Input id="assignment-end" type="date" bind:value={assignmentForm.end_date} /></div>
				</div>
				<div><label for="assignment-notes" class="text-sm font-medium">Catatan</label><Textarea id="assignment-notes" bind:value={assignmentForm.notes} /></div>
			{:else if dialogKind === 'document'}
				<div class="grid gap-3 sm:grid-cols-2">
					<div>
						<label for="document-type" class="text-sm font-medium">Jenis Dokumen</label>
						<select id="document-type" bind:value={documentForm.doc_type} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">{#each DOC_TYPES as [value, label] (value)}<option {value}>{label}</option>{/each}</select>
					</div>
					<div><label for="document-title" class="text-sm font-medium">Judul</label><Input id="document-title" bind:value={documentForm.title} /></div>
					<div><label for="document-year" class="text-sm font-medium">Tahun</label><Input id="document-year" type="number" bind:value={documentForm.period_year} /></div>
					<div><label for="document-period" class="text-sm font-medium">Label Periode</label><Input id="document-period" bind:value={documentForm.period_label} /></div>
					<div>
						<label for="document-unit" class="text-sm font-medium">Unit Pemilik</label>
						<select id="document-unit" bind:value={documentForm.owner_unit_id} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"><option value="">Tanpa unit</option>{#each units as unit (unit.id)}<option value={unit.id}>{unit.name}</option>{/each}</select>
					</div>
					<div>
						<label for="document-snp" class="text-sm font-medium">Standar SNP</label>
						<select id="document-snp" bind:value={documentForm.snp_standard} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">{#each SNP_STANDARDS as [value, label] (value)}<option {value}>{label}</option>{/each}</select>
					</div>
					<div>
						<label for="document-status" class="text-sm font-medium">Status</label>
						<select id="document-status" bind:value={documentForm.status} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">{#each DOCUMENT_STATUSES as [value, label] (value)}<option {value}>{label}</option>{/each}</select>
					</div>
					<div><label for="document-url" class="text-sm font-medium">URL/Lokasi File</label><Input id="document-url" bind:value={documentForm.document_url} /></div>
				</div>
				<div><label for="document-summary" class="text-sm font-medium">Ringkasan</label><Textarea id="document-summary" bind:value={documentForm.summary} /></div>
			{:else if dialogKind === 'program'}
				<div class="grid gap-3 sm:grid-cols-2">
					<div><label for="program-year" class="text-sm font-medium">Tahun</label><Input id="program-year" type="number" bind:value={programForm.period_year} /></div>
					<div><label for="program-code" class="text-sm font-medium">Kode Program</label><Input id="program-code" bind:value={programForm.code} /></div>
					<div class="sm:col-span-2"><label for="program-name" class="text-sm font-medium">Nama Program</label><Input id="program-name" bind:value={programForm.name} /></div>
					<div>
						<label for="program-doc" class="text-sm font-medium">Dokumen Sumber</label>
						<select id="program-doc" bind:value={programForm.source_document_id} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"><option value="">Tanpa dokumen</option>{#each documents as document (document.id)}<option value={document.id}>{document.title}</option>{/each}</select>
					</div>
					<div>
						<label for="program-unit" class="text-sm font-medium">Unit Pemilik</label>
						<select id="program-unit" bind:value={programForm.owner_unit_id} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"><option value="">Tanpa unit</option>{#each units as unit (unit.id)}<option value={unit.id}>{unit.name}</option>{/each}</select>
					</div>
					<div>
						<label for="program-position" class="text-sm font-medium">Jabatan Penanggung Jawab</label>
						<select id="program-position" bind:value={programForm.responsible_position_id} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"><option value="">Tanpa jabatan</option>{#each positions as position (position.id)}<option value={position.id}>{position.title}</option>{/each}</select>
					</div>
					<div>
						<label for="program-employee" class="text-sm font-medium">Pegawai Penanggung Jawab</label>
						<select id="program-employee" bind:value={programForm.responsible_employee_id} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"><option value="">Tanpa pegawai</option>{#each employees as employee (employee.id)}<option value={employee.id}>{employee.nama}</option>{/each}</select>
					</div>
					<div>
						<label for="program-snp" class="text-sm font-medium">Standar SNP</label>
						<select id="program-snp" bind:value={programForm.snp_standard} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">{#each SNP_STANDARDS as [value, label] (value)}<option {value}>{label}</option>{/each}</select>
					</div>
					<div><label for="program-iku" class="text-sm font-medium">Kode IKU</label><Input id="program-iku" bind:value={programForm.iku_code} /></div>
					<div class="sm:col-span-2"><label for="program-indicator" class="text-sm font-medium">Indikator</label><Textarea id="program-indicator" bind:value={programForm.indicator} /></div>
					<div><label for="program-target-value" class="text-sm font-medium">Target</label><Input id="program-target-value" bind:value={programForm.target_value} /></div>
					<div><label for="program-target-unit" class="text-sm font-medium">Satuan Target</label><Input id="program-target-unit" bind:value={programForm.target_unit} /></div>
					<div>
						<label for="program-status" class="text-sm font-medium">Status</label>
						<select id="program-status" bind:value={programForm.status} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">{#each PROGRAM_STATUSES as [value, label] (value)}<option {value}>{label}</option>{/each}</select>
					</div>
					<div><label for="program-progress" class="text-sm font-medium">Progres (%)</label><Input id="program-progress" type="number" min="0" max="100" bind:value={programForm.progress_percent} /></div>
					<div><label for="program-due" class="text-sm font-medium">Batas Waktu</label><Input id="program-due" type="date" bind:value={programForm.due_date} /></div>
					<div><label for="program-evidence" class="text-sm font-medium">URL/Lokasi Bukti</label><Input id="program-evidence" bind:value={programForm.evidence_url} /></div>
				</div>
				<div><label for="program-realization" class="text-sm font-medium">Realisasi</label><Textarea id="program-realization" bind:value={programForm.realization_summary} /></div>
			{:else if dialogKind === 'workplan'}
				<div class="grid gap-3 sm:grid-cols-2">
					<div>
						<label for="workplan-program" class="text-sm font-medium">Program Induk</label>
						<select id="workplan-program" bind:value={workPlanItemForm.program_id} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
							<option value="">Pilih program</option>
							{#each programs as program (program.id)}<option value={program.id}>{program.code} · {program.name}</option>{/each}
						</select>
					</div>
					<div><label for="workplan-year" class="text-sm font-medium">Tahun</label><Input id="workplan-year" type="number" bind:value={workPlanItemForm.period_year} /></div>
					<div><label for="workplan-code" class="text-sm font-medium">Kode Kegiatan</label><Input id="workplan-code" bind:value={workPlanItemForm.activity_code} /></div>
					<div><label for="workplan-name" class="text-sm font-medium">Nama Kegiatan</label><Input id="workplan-name" bind:value={workPlanItemForm.activity_name} /></div>
					<div>
						<label for="workplan-document" class="text-sm font-medium">Dokumen Sumber</label>
						<select id="workplan-document" bind:value={workPlanItemForm.source_document_id} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
							<option value="">Tanpa dokumen</option>
							{#each documents as document (document.id)}<option value={document.id}>{document.title}</option>{/each}
						</select>
					</div>
					<div>
						<label for="workplan-unit" class="text-sm font-medium">Unit Pemilik</label>
						<select id="workplan-unit" bind:value={workPlanItemForm.owner_unit_id} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
							<option value="">Tanpa unit</option>
							{#each units as unit (unit.id)}<option value={unit.id}>{unit.name}</option>{/each}
						</select>
					</div>
					<div>
						<label for="workplan-employee" class="text-sm font-medium">Penanggung Jawab</label>
						<select id="workplan-employee" bind:value={workPlanItemForm.responsible_employee_id} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
							<option value="">Tanpa pegawai</option>
							{#each employees as employee (employee.id)}<option value={employee.id}>{employee.nama} · {employee.nip}</option>{/each}
						</select>
					</div>
					<div>
						<label for="workplan-evidence-item" class="text-sm font-medium">Bukti Mutu Terkait</label>
						<select id="workplan-evidence-item" bind:value={workPlanItemForm.evidence_item_id} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
							<option value="">Tanpa item bukti</option>
							{#each evidenceItems as item (item.id)}<option value={item.id}>{item.title}</option>{/each}
						</select>
					</div>
					<div class="sm:col-span-2"><label for="workplan-output" class="text-sm font-medium">Indikator Output</label><Textarea id="workplan-output" bind:value={workPlanItemForm.output_indicator} /></div>
					<div><label for="workplan-target-volume" class="text-sm font-medium">Target Volume</label><Input id="workplan-target-volume" bind:value={workPlanItemForm.target_volume} /></div>
					<div><label for="workplan-target-unit" class="text-sm font-medium">Satuan Target</label><Input id="workplan-target-unit" bind:value={workPlanItemForm.target_unit} /></div>
					<div><label for="workplan-budget-source" class="text-sm font-medium">Sumber Anggaran</label><Input id="workplan-budget-source" bind:value={workPlanItemForm.budget_source} /></div>
					<div><label for="workplan-budget" class="text-sm font-medium">Anggaran</label><Input id="workplan-budget" type="number" min="0" bind:value={workPlanItemForm.budget_amount} /></div>
					<div><label for="workplan-realization" class="text-sm font-medium">Realisasi Anggaran</label><Input id="workplan-realization" type="number" min="0" bind:value={workPlanItemForm.realization_amount} /></div>
					<div>
						<label for="workplan-status" class="text-sm font-medium">Status</label>
						<select id="workplan-status" bind:value={workPlanItemForm.status} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">{#each PROGRAM_STATUSES as [value, label] (value)}<option {value}>{label}</option>{/each}</select>
					</div>
					<div><label for="workplan-progress" class="text-sm font-medium">Progres (%)</label><Input id="workplan-progress" type="number" min="0" max="100" bind:value={workPlanItemForm.progress_percent} /></div>
					<div><label for="workplan-start" class="text-sm font-medium">Tanggal Mulai</label><Input id="workplan-start" type="date" bind:value={workPlanItemForm.start_date} /></div>
					<div><label for="workplan-end" class="text-sm font-medium">Tanggal Selesai</label><Input id="workplan-end" type="date" bind:value={workPlanItemForm.end_date} /></div>
					<div class="sm:col-span-2"><label for="workplan-evidence-url" class="text-sm font-medium">URL/Lokasi Bukti</label><Input id="workplan-evidence-url" bind:value={workPlanItemForm.evidence_url} /></div>
				</div>
				<div><label for="workplan-notes" class="text-sm font-medium">Catatan</label><Textarea id="workplan-notes" bind:value={workPlanItemForm.notes} /></div>
			{:else if dialogKind === 'performance'}
				<div class="grid gap-3 sm:grid-cols-2">
					<div>
						<label for="performance-employee" class="text-sm font-medium">Pegawai</label>
						<select id="performance-employee" bind:value={performanceTargetForm.employee_id} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
							<option value="">Pilih pegawai</option>
							{#each employees as employee (employee.id)}<option value={employee.id}>{employee.nama} · {employee.nip}</option>{/each}
						</select>
					</div>
					<div>
						<label for="performance-position" class="text-sm font-medium">Jabatan</label>
						<select id="performance-position" bind:value={performanceTargetForm.position_id} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
							<option value="">Tanpa jabatan</option>
							{#each positions as position (position.id)}<option value={position.id}>{position.title}</option>{/each}
						</select>
					</div>
					<div><label for="performance-year" class="text-sm font-medium">Tahun</label><Input id="performance-year" type="number" bind:value={performanceTargetForm.period_year} /></div>
					<div>
						<label for="performance-aspect" class="text-sm font-medium">Aspek</label>
						<select id="performance-aspect" bind:value={performanceTargetForm.aspect} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
							{#each PERFORMANCE_ASPECTS as [value, label] (value)}<option {value}>{label}</option>{/each}
						</select>
					</div>
					<div class="sm:col-span-2"><label for="performance-title" class="text-sm font-medium">Target Kerja</label><Input id="performance-title" bind:value={performanceTargetForm.title} /></div>
					<div>
						<label for="performance-program" class="text-sm font-medium">Program Terkait</label>
						<select id="performance-program" bind:value={performanceTargetForm.program_id} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
							<option value="">Tanpa program</option>
							{#each programs as program (program.id)}<option value={program.id}>{program.code} · {program.name}</option>{/each}
						</select>
					</div>
					<div>
						<label for="performance-parent" class="text-sm font-medium">Target Induk</label>
						<select id="performance-parent" bind:value={performanceTargetForm.parent_target_id} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
							<option value="">Tanpa target induk</option>
							{#each performanceTargets.filter((target) => target.id !== editingId) as target (target.id)}<option value={target.id}>{target.employee_name} · {target.title}</option>{/each}
						</select>
					</div>
					<div class="sm:col-span-2"><label for="performance-indicator" class="text-sm font-medium">Indikator</label><Textarea id="performance-indicator" bind:value={performanceTargetForm.indicator} /></div>
					<div><label for="performance-target-value" class="text-sm font-medium">Target</label><Input id="performance-target-value" bind:value={performanceTargetForm.target_value} /></div>
					<div><label for="performance-target-unit" class="text-sm font-medium">Satuan Target</label><Input id="performance-target-unit" bind:value={performanceTargetForm.target_unit} /></div>
					<div>
						<label for="performance-status" class="text-sm font-medium">Status</label>
						<select id="performance-status" bind:value={performanceTargetForm.status} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">{#each PROGRAM_STATUSES as [value, label] (value)}<option {value}>{label}</option>{/each}</select>
					</div>
					<div><label for="performance-progress" class="text-sm font-medium">Progres (%)</label><Input id="performance-progress" type="number" min="0" max="100" bind:value={performanceTargetForm.progress_percent} /></div>
					<div><label for="performance-due" class="text-sm font-medium">Batas Waktu</label><Input id="performance-due" type="date" bind:value={performanceTargetForm.due_date} /></div>
					<div><label for="performance-evidence" class="text-sm font-medium">URL/Lokasi Bukti</label><Input id="performance-evidence" bind:value={performanceTargetForm.evidence_url} /></div>
				</div>
				<div><label for="performance-review" class="text-sm font-medium">Catatan Evaluasi</label><Textarea id="performance-review" bind:value={performanceTargetForm.review_notes} /></div>
			{:else if dialogKind === 'evidence'}
				<div class="grid gap-3 sm:grid-cols-2">
					<div><label for="evidence-year" class="text-sm font-medium">Tahun</label><Input id="evidence-year" type="number" bind:value={evidenceItemForm.period_year} /></div>
					<div>
						<label for="evidence-type" class="text-sm font-medium">Jenis Bukti</label>
						<select id="evidence-type" bind:value={evidenceItemForm.evidence_type} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
							{#each EVIDENCE_TYPES as [value, label] (value)}<option {value}>{label}</option>{/each}
						</select>
					</div>
					<div class="sm:col-span-2"><label for="evidence-title" class="text-sm font-medium">Judul Bukti</label><Input id="evidence-title" bind:value={evidenceItemForm.title} /></div>
					<div>
						<label for="evidence-snp" class="text-sm font-medium">Standar SNP</label>
						<select id="evidence-snp" bind:value={evidenceItemForm.snp_standard} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
							{#each SNP_STANDARDS as [value, label] (value)}<option {value}>{label}</option>{/each}
						</select>
					</div>
					<div>
						<label for="evidence-status" class="text-sm font-medium">Status</label>
						<select id="evidence-status" bind:value={evidenceItemForm.status} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
							{#each EVIDENCE_STATUSES as [value, label] (value)}<option {value}>{label}</option>{/each}
						</select>
					</div>
					<div>
						<label for="evidence-unit" class="text-sm font-medium">Unit Pemilik</label>
						<select id="evidence-unit" bind:value={evidenceItemForm.owner_unit_id} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
							<option value="">Tanpa unit</option>
							{#each units as unit (unit.id)}<option value={unit.id}>{unit.name}</option>{/each}
						</select>
					</div>
					<div>
						<label for="evidence-document" class="text-sm font-medium">Dokumen Terkait</label>
						<select id="evidence-document" bind:value={evidenceItemForm.document_id} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
							<option value="">Tanpa dokumen</option>
							{#each documents as document (document.id)}<option value={document.id}>{document.title}</option>{/each}
						</select>
					</div>
					<div>
						<label for="evidence-program" class="text-sm font-medium">Program Terkait</label>
						<select id="evidence-program" bind:value={evidenceItemForm.program_id} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
							<option value="">Tanpa program</option>
							{#each programs as program (program.id)}<option value={program.id}>{program.code} · {program.name}</option>{/each}
						</select>
					</div>
					<div>
						<label for="evidence-performance" class="text-sm font-medium">Target Kinerja Terkait</label>
						<select id="evidence-performance" bind:value={evidenceItemForm.performance_target_id} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
							<option value="">Tanpa target</option>
							{#each performanceTargets as target (target.id)}<option value={target.id}>{target.employee_name} · {target.title}</option>{/each}
						</select>
					</div>
					<div><label for="evidence-source" class="text-sm font-medium">Sumber</label><Input id="evidence-source" bind:value={evidenceItemForm.source_module} /></div>
					<div><label for="evidence-url" class="text-sm font-medium">URL/Lokasi File</label><Input id="evidence-url" bind:value={evidenceItemForm.evidence_url} /></div>
				</div>
				<div><label for="evidence-notes" class="text-sm font-medium">Catatan Review</label><Textarea id="evidence-notes" bind:value={evidenceItemForm.notes} /></div>
			{/if}
		</div>

		<Dialog.Footer>
			<Button variant="outline" onclick={() => (dialogOpen = false)}>Batal</Button>
			<LoadingButton loading={busy} onclick={() => void saveDialog()}>Simpan</LoadingButton>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
