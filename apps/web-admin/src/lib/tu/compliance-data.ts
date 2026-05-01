import { fetchSchoolProfile, type SchoolProfile } from '$lib/school-profile';

export type Fetcher = (input: RequestInfo | URL, init?: RequestInit) => Promise<Response>;

export type IncomingLetter = {
	id: string;
	nomor_surat: string;
	nomor_agenda: string;
	tanggal_surat: string;
	tanggal_terima: string;
	asal: string;
	perihal: string;
	sifat: string;
	status: string;
	received_by_name: string;
	disposisi_count: number;
};

export type OutgoingLetter = {
	id: string;
	nomor_surat: string;
	classification_code: string;
	classification_name: string;
	tanggal_surat: string;
	tujuan: string;
	perihal: string;
	sifat: string;
	issued_by_name: string;
};

export type LetterDisposition = {
	id: string;
	status: string;
	instruksi: string;
	catatan_tindak_lanjut: string;
	disposed_at: string;
	completed_at?: string;
	assignee_name: string;
	disposed_by_name: string;
	nomor_agenda: string;
	letter_perihal: string;
	letter_asal: string;
};

export type StudentCertificate = {
	id: string;
	template_code: string;
	template_name: string;
	student_name: string;
	student_nis: string;
	class_name: string;
	nomor_surat: string;
	tanggal_surat: string;
	purpose: string;
	status: string;
	created_by_username: string;
};

export type ArchiveStats = {
	active_categories: number;
	total_documents: number;
	active_documents: number;
	borrowed_documents: number;
	disposed_documents: number;
	documents_this_year: number;
	total_file_size: number;
};

export type ArchiveDocument = {
	id: string;
	category_code: string;
	category_name: string;
	classification_code: string;
	title: string;
	archive_number: string;
	received_date: string;
	retention_until?: string;
	status: string;
	storage_location: string;
	original_name: string;
	file_size: number;
};

export type InventoryStats = {
	total_jenis: number;
	total_unit: number | string;
	total_layak: number | string;
	perlu_restok: number;
	perlu_perawatan: number;
};

export type InventoryItem = {
	id: string;
	kode: string;
	nama: string;
	kategori: string;
	lokasi: string;
	kondisi: string;
	satuan: string;
	jumlah_total: number;
	jumlah_baik: number;
	min_stock: number;
	catatan: string;
};

export type WorkPlanItem = {
	id: string;
	period_year: number;
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
	owner_unit_name?: string;
	responsible_employee_name?: string;
	evidence_item_title?: string;
};

export type EvidenceItem = {
	id: string;
	period_year: number;
	title: string;
	evidence_type: string;
	snp_standard: string;
	source_module: string;
	evidence_url: string;
	status: string;
	notes: string;
	owner_unit_name?: string;
	program_code?: string;
	program_name?: string;
	performance_target_title?: string;
};

export type TUComplianceData = {
	schoolProfile: SchoolProfile;
	incomingLetters: IncomingLetter[];
	outgoingLetters: OutgoingLetter[];
	dispositions: LetterDisposition[];
	certificates: StudentCertificate[];
	archiveStats: ArchiveStats;
	archiveDocuments: ArchiveDocument[];
	inventoryStats: InventoryStats;
	inventoryItems: InventoryItem[];
	workPlanItems: WorkPlanItem[];
	evidenceItems: EvidenceItem[];
};

export type CsvValue = string | number | boolean | null | undefined;

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

export async function readApi<T>(response: Response, fallbackMessage: string): Promise<T> {
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

export async function fetchTUComplianceData(fetcher: Fetcher = fetch): Promise<TUComplianceData> {
	const [
		schoolProfile,
		incomingRes,
		outgoingRes,
		dispositionRes,
		certificateRes,
		archiveStatsRes,
		archiveDocumentsRes,
		inventoryStatsRes,
		inventoryItemsRes,
		workPlanRes,
		evidenceRes,
	] = await Promise.all([
		fetchSchoolProfile(fetcher),
		fetcher('/api/tu/surat/incoming'),
		fetcher('/api/tu/surat/outgoing'),
		fetcher('/api/tu/surat/disposisi'),
		fetcher('/api/tu/surat-keterangan'),
		fetcher('/api/tu/archives/stats'),
		fetcher('/api/tu/archives/documents'),
		fetcher('/api/inventory/stats'),
		fetcher('/api/inventory/items'),
		fetcher('/api/governance/work-plan-items'),
		fetcher('/api/governance/evidence-items'),
	]);

	const [
		incomingLetters,
		outgoingLetters,
		dispositions,
		certificates,
		archiveStats,
		archiveDocuments,
		inventoryStats,
		inventoryItems,
		workPlanItems,
		evidenceItems,
	] = await Promise.all([
		readApi<IncomingLetter[]>(incomingRes, 'Gagal memuat surat masuk.'),
		readApi<OutgoingLetter[]>(outgoingRes, 'Gagal memuat surat keluar.'),
		readApi<LetterDisposition[]>(dispositionRes, 'Gagal memuat disposisi.'),
		readApi<StudentCertificate[]>(certificateRes, 'Gagal memuat surat keterangan.'),
		readApi<ArchiveStats>(archiveStatsRes, 'Gagal memuat statistik arsip.'),
		readApi<ArchiveDocument[]>(archiveDocumentsRes, 'Gagal memuat arsip.'),
		readApi<InventoryStats>(inventoryStatsRes, 'Gagal memuat statistik inventaris.'),
		readApi<InventoryItem[]>(inventoryItemsRes, 'Gagal memuat inventaris.'),
		readApi<WorkPlanItem[]>(workPlanRes, 'Gagal memuat RKT/RKJM.'),
		readApi<EvidenceItem[]>(evidenceRes, 'Gagal memuat bukti mutu.'),
	]);

	return {
		schoolProfile,
		incomingLetters: incomingLetters ?? [],
		outgoingLetters: outgoingLetters ?? [],
		dispositions: dispositions ?? [],
		certificates: certificates ?? [],
		archiveStats,
		archiveDocuments: archiveDocuments ?? [],
		inventoryStats,
		inventoryItems: inventoryItems ?? [],
		workPlanItems: workPlanItems ?? [],
		evidenceItems: evidenceItems ?? [],
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

export function formatCurrency(value: number | string | undefined) {
	return new Intl.NumberFormat('id-ID', {
		style: 'currency',
		currency: 'IDR',
		maximumFractionDigits: 0,
	}).format(numberValue(value));
}

export function formatBytes(value: number) {
	if (!Number.isFinite(value) || value <= 0) return '0 B';
	if (value < 1024) return `${value} B`;
	if (value < 1024 * 1024) return `${(value / 1024).toFixed(1)} KB`;
	return `${(value / (1024 * 1024)).toFixed(1)} MB`;
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

export function isOpenIncoming(letter: IncomingLetter) {
	return letter.status === 'baru' || letter.status === 'didisposisi';
}

export function isOpenDisposition(disposition: LetterDisposition) {
	return disposition.status !== 'selesai';
}

export function isInventoryAttention(item: InventoryItem) {
	return item.kondisi !== 'baik' || item.jumlah_baik <= item.min_stock;
}

export function isArchiveRetentionAttention(item: ArchiveDocument, days = 90) {
	if (!item.retention_until) return false;
	const now = Date.now();
	const target = dateValue(item.retention_until);
	return target > 0 && target - now <= days * 24 * 60 * 60 * 1000;
}

export function isWorkPlanAttention(item: WorkPlanItem) {
	return item.status === 'blocked' || (item.status !== 'done' && item.progress_percent < 50);
}

export function statusLabel(value: string) {
	switch (value) {
		case 'baru':
			return 'Baru';
		case 'didisposisi':
			return 'Didisposisi';
		case 'selesai':
			return 'Selesai';
		case 'arsip':
			return 'Arsip';
		case 'terkirim':
			return 'Terkirim';
		case 'dibaca':
			return 'Dibaca';
		case 'ditindaklanjuti':
			return 'Ditindaklanjuti';
		case 'issued':
			return 'Terbit';
		case 'cancelled':
			return 'Batal';
		case 'active':
			return 'Aktif';
		case 'borrowed':
			return 'Dipinjam';
		case 'disposed':
			return 'Dimusnahkan';
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
