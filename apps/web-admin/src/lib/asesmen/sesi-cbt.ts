export type SesiCbtStatus = 'draft' | 'scheduled' | 'active' | 'finished' | 'archived';

export type SesiCbtRow = {
	id: string;
	exam_id: string;
	exam_title: string;
	title: string;
	subject_name?: string;
	package_title?: string;
	class_code?: string;
	class_name?: string;
	room_count: number;
	room_labels: string[];
	participant_count: number;
	assigned_participant_count: number;
	working_count: number;
	finished_count: number;
	incident_count: number;
	scheduled_start?: string;
	scheduled_end?: string;
	status: SesiCbtStatus;
	monitor_token?: string;
	ready_to_activate: boolean;
	blockers: string[];
};

export type SesiCbtSummary = {
	total: number;
	active: number;
	running: number;
	finished: number;
	archived: number;
	incident: number;
};

export type SesiCbtMatrixGroup = {
	key: string;
	anchor: SesiCbtRow;
	rows: SesiCbtRow[];
	classLabels: string[];
	roomLabels: string[];
	packageLabels: string[];
	totalParticipants: number;
};

type RawRecord = Record<string, unknown>;

const STATUSES = new Set<SesiCbtStatus>(['draft', 'scheduled', 'active', 'finished', 'archived']);

function isRecord(value: unknown): value is RawRecord {
	return typeof value === 'object' && value !== null;
}

function stringValue(value: unknown): string | undefined {
	return typeof value === 'string' && value.trim() ? value.trim() : undefined;
}

function numberValue(value: unknown): number {
	const parsed = Number(value ?? 0);
	return Number.isFinite(parsed) && parsed > 0 ? parsed : 0;
}

function statusValue(value: unknown): SesiCbtStatus {
	const raw = stringValue(value)?.toLowerCase();
	if (raw === 'ready') return 'scheduled';
	if (raw === 'running') return 'active';
	if (raw && STATUSES.has(raw as SesiCbtStatus)) return raw as SesiCbtStatus;
	return 'draft';
}

function roomLabels(raw: RawRecord): string[] {
	const direct = raw.room_labels ?? raw.rooms ?? raw.jadwal_ruang_labels;
	if (Array.isArray(direct)) {
		return direct
			.map((item) => {
				if (typeof item === 'string') return item.trim();
				if (isRecord(item)) return stringValue(item.code) ?? stringValue(item.name) ?? stringValue(item.room_code) ?? stringValue(item.room_name) ?? '';
				return '';
			})
			.filter(Boolean);
	}
	const single = stringValue(raw.room_label) ?? stringValue(raw.room_code) ?? stringValue(raw.room_name);
	return single ? [single] : [];
}

export function sessionStatusLabel(status: SesiCbtStatus | string): string {
	const normalized = statusValue(status);
	if (normalized === 'scheduled') return 'Terjadwal';
	if (normalized === 'active') return 'Aktif';
	if (normalized === 'finished') return 'Selesai';
	if (normalized === 'archived') return 'Arsip';
	return 'Draft';
}

export function sessionStatusTone(status: SesiCbtStatus | string): string {
	const normalized = statusValue(status);
	if (normalized === 'active') return 'border-emerald-200 bg-emerald-50 text-emerald-700';
	if (normalized === 'scheduled') return 'border-sky-200 bg-sky-50 text-sky-700';
	if (normalized === 'finished') return 'border-slate-200 bg-slate-50 text-slate-700';
	if (normalized === 'archived') return 'border-zinc-200 bg-zinc-50 text-zinc-600';
	return 'border-amber-200 bg-amber-50 text-amber-700';
}

export function sessionIsActiveWindow(row: Pick<SesiCbtRow, 'scheduled_start' | 'scheduled_end' | 'status'>, now = new Date()): boolean {
	if (row.status === 'active') return true;
	if (!row.scheduled_start || !row.scheduled_end) return false;
	const start = new Date(row.scheduled_start).getTime();
	const end = new Date(row.scheduled_end).getTime();
	const current = now.getTime();
	return Number.isFinite(start) && Number.isFinite(end) && start <= current && current <= end;
}

export function sessionBlockers(row: Pick<SesiCbtRow, 'exam_id' | 'package_title' | 'participant_count' | 'assigned_participant_count' | 'room_count' | 'scheduled_start' | 'scheduled_end'>): string[] {
	const blockers: string[] = [];
	if (!row.exam_id) blockers.push('Kegiatan belum terikat');
	if (!row.package_title) blockers.push('Paket soal belum dipilih');
	if (row.participant_count <= 0) blockers.push('Peserta belum masuk');
	if (row.room_count <= 0) blockers.push('Ruang belum tersusun');
	if (row.participant_count > 0 && row.assigned_participant_count < row.participant_count) blockers.push('Kursi peserta belum lengkap');
	const start = row.scheduled_start ? new Date(row.scheduled_start).getTime() : NaN;
	const end = row.scheduled_end ? new Date(row.scheduled_end).getTime() : NaN;
	if (!Number.isFinite(start) || !Number.isFinite(end) || end <= start) blockers.push('Jadwal belum valid');
	return blockers;
}

export function normalizeSesiCbtRow(rawValue: unknown, examValue?: unknown): SesiCbtRow {
	const raw = isRecord(rawValue) ? rawValue : {};
	const exam = isRecord(examValue) ? examValue : {};
	const labels = roomLabels(raw);
	const participantCount = numberValue(raw.participant_count ?? raw.peserta_total ?? raw.total_participants ?? raw.stats_total);
	const assignedCount = numberValue(raw.assigned_participant_count ?? raw.assigned_total ?? raw.seated_count);
	const roomCount = numberValue(raw.room_count ?? labels.length);
	const status = statusValue(raw.status ?? raw.operational_status);
	const row: SesiCbtRow = {
		id: stringValue(raw.id) ?? stringValue(raw.session_id) ?? '',
		exam_id: stringValue(raw.exam_id) ?? stringValue(raw.event_id) ?? stringValue(exam.id) ?? '',
		exam_title: stringValue(raw.exam_title) ?? stringValue(raw.event_title) ?? stringValue(exam.title) ?? stringValue(exam.nama) ?? 'Kegiatan ujian',
		title: stringValue(raw.title) ?? stringValue(raw.name) ?? stringValue(raw.slot_label) ?? 'Sesi CBT',
		subject_name: stringValue(raw.subject_name) ?? stringValue(raw.mapel),
		package_title: stringValue(raw.package_title) ?? stringValue(raw.paket_soal) ?? stringValue(raw.paketSoal),
		class_code: stringValue(raw.class_code) ?? stringValue(raw.kelas_code),
		class_name: stringValue(raw.class_name) ?? stringValue(raw.kelas_name) ?? stringValue(raw.kelasTarget),
		room_count: roomCount,
		room_labels: labels,
		participant_count: participantCount,
		assigned_participant_count: assignedCount,
		working_count: numberValue(raw.working_count ?? raw.mengerjakan ?? (isRecord(raw.stats) ? raw.stats.mengerjakan : undefined)),
		finished_count: numberValue(raw.finished_count ?? raw.selesai ?? (isRecord(raw.stats) ? raw.stats.selesai : undefined)),
		incident_count: numberValue(raw.incident_count ?? raw.insiden ?? raw.unresolved_incident_count),
		scheduled_start: stringValue(raw.scheduled_start) ?? stringValue(raw.starts_at) ?? stringValue(raw.start_at) ?? stringValue(raw.startAt),
		scheduled_end: stringValue(raw.scheduled_end) ?? stringValue(raw.ends_at) ?? stringValue(raw.end_at) ?? stringValue(raw.endAt),
		status,
		monitor_token: stringValue(raw.monitor_token) ?? stringValue(raw.monitorToken),
		ready_to_activate: false,
		blockers: []
	};
	row.blockers = Array.isArray(raw.blockers) ? raw.blockers.map((item) => String(item)).filter(Boolean) : sessionBlockers(row);
	row.ready_to_activate = Boolean(raw.ready_to_activate) || row.blockers.length === 0;
	return row;
}

export function summarizeSessionRows(rows: SesiCbtRow[]): SesiCbtSummary {
	return rows.reduce<SesiCbtSummary>((summary, row) => {
		summary.total += 1;
		if (row.status === 'active' || row.status === 'scheduled') summary.active += 1;
		if (sessionIsActiveWindow(row)) summary.running += 1;
		if (row.status === 'finished') summary.finished += 1;
		if (row.status === 'archived') summary.archived += 1;
		summary.incident += row.incident_count;
		return summary;
	}, { total: 0, active: 0, running: 0, finished: 0, archived: 0, incident: 0 });
}

function uniqueSorted(values: Array<string | undefined>): string[] {
	return Array.from(new Set(values.map((value) => value?.trim()).filter(Boolean) as string[])).sort((a, b) => a.localeCompare(b));
}

export function groupSesiCbtMatrix(rows: SesiCbtRow[]): SesiCbtMatrixGroup[] {
	const groups = new Map<string, SesiCbtRow[]>();
	for (const row of rows) {
		const startKey = row.scheduled_start ? row.scheduled_start.slice(0, 16) : 'tanpa-jadwal';
		const key = [row.exam_id, row.subject_name ?? row.title, startKey].join('::');
		const current = groups.get(key) ?? [];
		current.push(row);
		groups.set(key, current);
	}
	return Array.from(groups.entries()).map(([key, groupRows]) => {
		const anchor = groupRows[0] ?? normalizeSesiCbtRow({ id: key });
		return {
			key,
			anchor,
			rows: groupRows,
			classLabels: uniqueSorted(groupRows.map((row) => row.class_code ?? row.class_name)),
			roomLabels: uniqueSorted(groupRows.flatMap((row) => row.room_labels)),
			packageLabels: uniqueSorted(groupRows.map((row) => row.package_title)),
			totalParticipants: groupRows.reduce((total, row) => total + row.participant_count, 0)
		};
	});
}
