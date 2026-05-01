import { fetchSchoolProfile, type SchoolProfile } from '$lib/school-profile';

export type Fetcher = (input: RequestInfo | URL, init?: RequestInit) => Promise<Response>;
export type CsvValue = string | number | boolean | null | undefined;

export type GovernanceStats = {
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
	work_plan_budget_amount: number | string;
	work_plan_realization_amount: number | string;
	compliance_actions?: number;
	open_compliance_actions?: number;
	completed_compliance_actions?: number;
	critical_compliance_actions?: number;
};

export type GovernanceUnit = {
	id: string;
	code: string;
	name: string;
	unit_type: string;
	parent_id?: string;
	parent_unit_name?: string;
	description: string;
	is_active: boolean;
	sort_order: number;
};

export type GovernancePosition = {
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
};

export type GovernanceAssignment = {
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
};

export type GovernanceDocument = {
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
};

export type GovernanceProgram = {
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
};

export type GovernanceWorkPlanItem = {
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
	budget_amount: number | string;
	realization_amount: number | string;
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
};

export type GovernancePerformanceTarget = {
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
};

export type GovernanceEvidenceItem = {
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
};

export type GovernanceComplianceAction = {
	id: string;
	period_year: number;
	source_type: string;
	source_ref_id?: string;
	snp_standard: string;
	program_id?: string;
	document_id?: string;
	performance_target_id?: string;
	evidence_item_id?: string;
	owner_unit_id?: string;
	responsible_employee_id?: string;
	title: string;
	description: string;
	priority: string;
	status: string;
	due_date?: string;
	completed_at?: string;
	follow_up_notes: string;
	evidence_url: string;
	updated_at: string;
	program_code?: string;
	program_name?: string;
	document_title?: string;
	performance_target_title?: string;
	evidence_item_title?: string;
	owner_unit_name?: string;
	responsible_employee_name?: string;
	responsible_employee_nip?: string;
};

export type GovernanceSNPRow = {
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
};

export type GovernancePrintPackData = {
	schoolProfile: SchoolProfile;
	stats: GovernanceStats;
	units: GovernanceUnit[];
	positions: GovernancePosition[];
	assignments: GovernanceAssignment[];
	documents: GovernanceDocument[];
	programs: GovernanceProgram[];
	workPlanItems: GovernanceWorkPlanItem[];
	performanceTargets: GovernancePerformanceTarget[];
	evidenceItems: GovernanceEvidenceItem[];
	complianceActions: GovernanceComplianceAction[];
	snpMatrix: GovernanceSNPRow[];
};

type ApiEnvelope<T> = {
	data?: T;
	error?: string;
	message?: string;
};

function isRecord(value: unknown): value is Record<string, unknown> {
	return typeof value === 'object' && value !== null;
}

function apiErrorMessage(payload: unknown) {
	if (!isRecord(payload)) return '';
	const error = payload.error;
	if (typeof error === 'string' && error.trim()) return error;
	const message = payload.message;
	if (typeof message === 'string' && message.trim()) return message;
	return '';
}

async function readApi<T>(response: Response, fallbackMessage: string): Promise<T> {
	const payload = (await response.json().catch(() => null)) as ApiEnvelope<T> | T | null;
	const message = apiErrorMessage(payload);
	if (!response.ok || message) throw new Error(message || fallbackMessage);
	if (isRecord(payload) && 'data' in payload) {
		const envelope = payload as ApiEnvelope<T>;
		if (envelope.data === undefined) throw new Error(fallbackMessage);
		return envelope.data;
	}
	if (payload === null) throw new Error(fallbackMessage);
	return payload as T;
}

export async function fetchGovernancePrintPackData(fetcher: Fetcher = fetch): Promise<GovernancePrintPackData> {
	const [
		schoolProfile,
		statsRes,
		unitsRes,
		positionsRes,
		assignmentsRes,
		documentsRes,
		programsRes,
		workPlanRes,
		performanceRes,
		evidenceRes,
		complianceActionRes,
		snpRes,
	] = await Promise.all([
		fetchSchoolProfile(fetcher),
		fetcher('/api/governance/stats'),
		fetcher('/api/governance/units'),
		fetcher('/api/governance/positions'),
		fetcher('/api/governance/assignments?active_only=true'),
		fetcher('/api/governance/documents'),
		fetcher('/api/governance/programs'),
		fetcher('/api/governance/work-plan-items'),
		fetcher('/api/governance/performance-targets'),
		fetcher('/api/governance/evidence-items'),
		fetcher('/api/governance/compliance-actions'),
		fetcher('/api/governance/snp-matrix'),
	]);

	const [stats, units, positions, assignments, documents, programs, workPlanItems, performanceTargets, evidenceItems, complianceActions, snpMatrix] =
		await Promise.all([
			readApi<GovernanceStats>(statsRes, 'Gagal memuat statistik tata kelola.'),
			readApi<GovernanceUnit[]>(unitsRes, 'Gagal memuat unit kerja.'),
			readApi<GovernancePosition[]>(positionsRes, 'Gagal memuat jabatan.'),
			readApi<GovernanceAssignment[]>(assignmentsRes, 'Gagal memuat pejabat aktif.'),
			readApi<GovernanceDocument[]>(documentsRes, 'Gagal memuat dokumen tata kelola.'),
			readApi<GovernanceProgram[]>(programsRes, 'Gagal memuat program.'),
			readApi<GovernanceWorkPlanItem[]>(workPlanRes, 'Gagal memuat RKT/RKJM.'),
			readApi<GovernancePerformanceTarget[]>(performanceRes, 'Gagal memuat target SKP.'),
			readApi<GovernanceEvidenceItem[]>(evidenceRes, 'Gagal memuat bukti mutu.'),
			readApi<GovernanceComplianceAction[]>(complianceActionRes, 'Gagal memuat tindak lanjut kepatuhan.'),
			readApi<GovernanceSNPRow[]>(snpRes, 'Gagal memuat matriks 8 SNP.'),
		]);

	return {
		schoolProfile,
		stats,
		units: units ?? [],
		positions: positions ?? [],
		assignments: assignments ?? [],
		documents: documents ?? [],
		programs: programs ?? [],
		workPlanItems: workPlanItems ?? [],
		performanceTargets: performanceTargets ?? [],
		evidenceItems: evidenceItems ?? [],
		complianceActions: complianceActions ?? [],
		snpMatrix: snpMatrix ?? [],
	};
}

export function numberValue(value: number | string | undefined) {
	if (typeof value === 'number') return value;
	const parsed = Number(value ?? 0);
	return Number.isFinite(parsed) ? parsed : 0;
}

export function dateValue(value?: string) {
	if (!value) return 0;
	const time = new Date(value).getTime();
	return Number.isFinite(time) ? time : 0;
}

export function formatDate(value?: string) {
	if (!value) return '-';
	return new Intl.DateTimeFormat('id-ID', {
		timeZone: 'Asia/Makassar',
		day: '2-digit',
		month: 'short',
		year: 'numeric',
	}).format(new Date(value));
}

export function formatGeneratedAt(value: Date) {
	return new Intl.DateTimeFormat('id-ID', {
		timeZone: 'Asia/Makassar',
		day: '2-digit',
		month: 'long',
		year: 'numeric',
		hour: '2-digit',
		minute: '2-digit',
	}).format(value);
}

export function formatCurrency(value: number | string | undefined) {
	return new Intl.NumberFormat('id-ID', {
		style: 'currency',
		currency: 'IDR',
		maximumFractionDigits: 0,
	}).format(numberValue(value));
}

export function latestByDate<T>(items: T[], getDate: (item: T) => string | undefined, limit: number) {
	return [...items].sort((a, b) => dateValue(getDate(b)) - dateValue(getDate(a))).slice(0, limit);
}

export function csvEscape(value: CsvValue) {
	return `"${String(value ?? '').replaceAll('"', '""')}"`;
}

export function downloadCsv(filename: string, rows: CsvValue[][]) {
	const csv = rows.map((row) => row.map(csvEscape).join(',')).join('\n');
	const blob = new Blob([csv], { type: 'text/csv;charset=utf-8' });
	const url = URL.createObjectURL(blob);
	const link = document.createElement('a');
	link.href = url;
	link.download = filename;
	link.click();
	URL.revokeObjectURL(url);
}

export function docTypeLabel(value: string) {
	switch (value) {
		case 'visi_misi':
			return 'Visi Misi';
		case 'rkjm':
			return 'RKJM';
		case 'rkt':
			return 'RKT';
		case 'renstra':
			return 'Renstra';
		case 'perkin':
			return 'Perkin';
		case 'iku':
			return 'IKU';
		case 'sk':
			return 'SK';
		case 'sop':
			return 'SOP';
		case 'snp':
			return '8 SNP';
		default:
			return value || '-';
	}
}

export function statusLabel(value: string) {
	switch (value) {
		case 'draft':
			return 'Draft';
		case 'final':
			return 'Final';
		case 'arsip':
			return 'Arsip';
		case 'planned':
			return 'Direncanakan';
		case 'in_progress':
			return 'Berjalan';
		case 'done':
			return 'Selesai';
		case 'blocked':
			return 'Terkendala';
		case 'needed':
			return 'Dibutuhkan';
		case 'collected':
			return 'Terkumpul';
		case 'verified':
			return 'Terverifikasi';
		case 'gap':
			return 'Kurang';
		default:
			return value || '-';
	}
}

export function snpLabel(value: string) {
	switch (value) {
		case 'skl':
			return 'SKL';
		case 'isi':
			return 'Isi';
		case 'proses':
			return 'Proses';
		case 'penilaian':
			return 'Penilaian';
		case 'ptk':
			return 'PTK';
		case 'sarpras':
			return 'Sarpras';
		case 'pengelolaan':
			return 'Pengelolaan';
		case 'pembiayaan':
			return 'Pembiayaan';
		default:
			return 'Tidak dikaitkan';
	}
}
