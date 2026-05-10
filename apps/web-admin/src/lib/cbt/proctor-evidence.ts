export const PROCTOR_EVIDENCE_CATEGORIES = [
	'heartbeat',
	'app_background_resume',
	'device_mismatch',
	'submit_guard',
	'stale_connection',
	'warning',
	'anti_cheat',
	'force_submit',
	'reset_access',
	'export_print'
] as const;

export type ProctorEvidenceCategory = (typeof PROCTOR_EVIDENCE_CATEGORIES)[number];

export type ProctorEvidenceEvent = {
	event_type: string;
	event_data: unknown;
	created_at?: string;
	nama?: string;
	nis?: string;
	room_name?: string;
};

export type ProctorEvidenceParticipant = {
	participant_id: string;
	nama: string;
	nis: string;
	submitted_at: string | null;
	last_heartbeat: string | null;
	app_switch_count: number;
	screenshot_attempt: number;
	suspicious_flag: boolean;
	answered_count: number;
	score: string | null;
};

export type ProctorEvidenceSummary = {
	counts: Record<ProctorEvidenceCategory, number>;
	staleParticipantCount: number;
	missingCategories: ProctorEvidenceCategory[];
};

export type ProctorOperatorGuidanceItem = {
	title: string;
	description: string;
};

const PROCTOR_EVIDENCE_CATEGORY_LABELS: Record<ProctorEvidenceCategory, string> = {
	heartbeat: 'Online/heartbeat OK',
	app_background_resume: 'Background/resume',
	device_mismatch: 'Device mismatch',
	submit_guard: 'Submit guard',
	stale_connection: 'Stale connection',
	warning: 'Peringatan pengawas',
	anti_cheat: 'Anti-cheat BYOD',
	force_submit: 'Submitted/force submitted',
	reset_access: 'Reset akses',
	export_print: 'Export/print token-free'
};

const PROCTOR_EVIDENCE_CATEGORY_SUMMARIES: Record<ProctorEvidenceCategory, string> = {
	heartbeat: 'Heartbeat aktif atau kontak server terakhir masih sehat.',
	app_background_resume: 'Aplikasi sempat background/resume; verifikasi siswa tetap mengikuti arahan ruang.',
	device_mismatch: 'Perangkat tidak sesuai binding sesi; reset akses hanya setelah verifikasi identitas.',
	submit_guard: 'Submit ditahan karena pending sinkron/degraded mode; tunggu jawaban aman terkirim.',
	stale_connection: 'Koneksi stale atau heartbeat tertunda; pengawas perlu cek jaringan/perangkat.',
	warning: 'Peringatan BYOD atau catatan manual dari runtime ujian.',
	anti_cheat: 'Split screen, PiP, fokus hilang, background, atau lock anti-cheat tercatat.',
	force_submit: 'Submit manual/force submit sudah tercatat sebagai tindakan pengawas.',
	reset_access: 'Akses perangkat direset oleh pengawas setelah verifikasi ruang.',
	export_print: 'Bukti cetak/export tersedia tanpa dump token mentah.'
};

export const proctorOperatorGuidance: ProctorOperatorGuidanceItem[] = [
	{
		title: 'Kapan memperingatkan siswa',
		description: 'Saat app background/resume, heartbeat mulai stale, atau ada percobaan screenshot/app switch berulang.'
	},
	{
		title: 'Kapan reset akses',
		description: 'Saat perangkat sah perlu login ulang setelah pengawas memverifikasi identitas, ruang, dan alasan gangguan.'
	},
	{
		title: 'Kapan paksa submit',
		description: 'Saat ruang harus ditutup dan pengawas sudah memastikan jawaban tersinkron atau prosedur manual dicatat.'
	}
];

export function proctorEvidenceCategoryLabel(category: ProctorEvidenceCategory | null): string {
	return category ? PROCTOR_EVIDENCE_CATEGORY_LABELS[category] : 'Event lain';
}

export function proctorEvidenceCategorySummary(category: ProctorEvidenceCategory | null): string {
	return category ? PROCTOR_EVIDENCE_CATEGORY_SUMMARIES[category] : 'Event tidak masuk kategori utama evidence.';
}

type ProctorEvidenceSummaryInput = {
	participants: ProctorEvidenceParticipant[];
	events: ProctorEvidenceEvent[];
	now?: Date;
	staleAfterMinutes?: number;
	hasPrintPack?: boolean;
};

type ProctorEvidenceCsvInput = {
	sessionTitle: string;
	roomName: string;
	proctors: string[];
	participants: ProctorEvidenceParticipant[];
	events: ProctorEvidenceEvent[];
	generatedAt?: Date;
};

type EventDataRecord = Record<string, unknown>;

const WARNING_REASON_LABELS: Record<string, string> = {
	resume_exam: 'Resume gate',
	repeat_resume_attempt: 'Resume berulang',
	answer_saved_local_only: 'Jawaban lokal',
	submit_blocked_pending_sync: 'Submit ditahan: pending sync',
	auto_submit_blocked_pending_sync: 'Auto-submit ditahan',
	submit_blocked_degraded_mode: 'Submit ditahan: koneksi menurun',
	degraded_mode_entered: 'Koneksi menurun',
	stale_connection_attention: 'Koneksi stale',
	stale_connection_escalated: 'Intervensi stale urgent',
	back_button_attempt: 'Tombol kembali',
	manual_submit: 'Submit manual',
	device_mismatch: 'Device mismatch',
	token_already_bound: 'Device mismatch'
};

export function classifyProctorEvent(event: ProctorEvidenceEvent): ProctorEvidenceCategory | null {
	const eventType = event.event_type.trim().toLowerCase();
	const data = eventDataRecord(event.event_data);
	const reason = stringValue(data.reason).toLowerCase();

	if (eventType === 'heartbeat') return 'heartbeat';
	if (eventType === 'app_switch' || eventType === 'app_backgrounded' || eventType === 'app_resumed' || eventType === 'resume') {
		return 'app_background_resume';
	}
	if (eventType === 'device_mismatch') return 'device_mismatch';
	if (eventType === 'submit_guard') return 'submit_guard';
	if (eventType === 'anti_cheat_violation') return 'anti_cheat';
	if (eventType === 'stale_connection' || eventType === 'heartbeat_failed') return 'stale_connection';
	if (eventType === 'proctor_force_submit') return 'force_submit';
	if (eventType === 'proctor_reset_access') return 'reset_access';

	if (eventType === 'warning') {
		if (reason === 'resume_exam' || reason === 'repeat_resume_attempt') return 'app_background_resume';
		if (reason.includes('device_mismatch') || reason.includes('token_already_bound')) return 'device_mismatch';
		if (reason.includes('submit_blocked') || reason.includes('auto_submit_blocked')) return 'submit_guard';
		if (reason.includes('stale_connection') || reason === 'degraded_mode_entered') return 'stale_connection';
		return 'warning';
	}

	return null;
}

export function proctorEventLabel(event: ProctorEvidenceEvent): string {
	const eventType = event.event_type.trim().toLowerCase();
	const data = eventDataRecord(event.event_data);
	const reason = stringValue(data.reason).toLowerCase();
	if (eventType === 'warning' && reason) return WARNING_REASON_LABELS[reason] ?? reason.replaceAll('_', ' ');
	if (eventType === 'anti_cheat_violation' && reason) return reason.replaceAll('_', ' ');

	const labels: Record<string, string> = {
		login: 'Login',
		heartbeat: 'Heartbeat',
		app_switch: 'Keluar/kembali aplikasi',
		app_backgrounded: 'App background',
		app_resumed: 'App resume',
		resume: 'Kembali ujian',
		device_mismatch: 'Device mismatch',
		answer: 'Simpan jawaban',
		answer_save: 'Simpan jawaban',
		submit: 'Submit',
		submit_guard: 'Submit guard',
		stale_connection: 'Koneksi stale',
		heartbeat_failed: 'Heartbeat gagal',
		screenshot_attempt: 'Percobaan screenshot',
		proctor_reset_access: 'Reset akses',
		anti_cheat_violation: 'Anti-cheat BYOD',
		proctor_force_submit: 'Paksa submit'
	};
	return labels[eventType] ?? event.event_type.replaceAll('_', ' ');
}

export function summarizeProctorEvidence(input: ProctorEvidenceSummaryInput): ProctorEvidenceSummary {
	const counts = emptyCounts();
	for (const event of input.events) {
		const category = classifyProctorEvent(event);
		if (!category) continue;
		counts[category] += 1;
		if (event.event_type.trim().toLowerCase() === 'warning') counts.warning += 1;
	}

	for (const participant of input.participants) {
		if (participant.last_heartbeat) counts.heartbeat += 1;
	}

	const staleParticipantCount = input.participants.filter((participant) =>
		isParticipantStale(participant, input.now ?? new Date(), input.staleAfterMinutes ?? 2)
	).length;
	counts.stale_connection += staleParticipantCount;

	if (input.hasPrintPack) counts.export_print += 1;

	return {
		counts,
		staleParticipantCount,
		missingCategories: PROCTOR_EVIDENCE_CATEGORIES.filter((category) => counts[category] === 0)
	};
}

export function buildProctorEvidenceCsvRows(input: ProctorEvidenceCsvInput): string[][] {
	const generatedAt = (input.generatedAt ?? new Date()).toISOString();
	const rows: string[][] = [['section', 'category', 'timestamp', 'participant', 'nis', 'detail']];
	rows.push(['summary', 'session', generatedAt, '', '', input.sessionTitle]);
	rows.push(['summary', 'room', generatedAt, '', '', input.roomName]);
	rows.push(['summary', 'proctors', generatedAt, '', '', input.proctors.filter(Boolean).join('; ') || 'Belum ada pengawas']);

	for (const participant of input.participants) {
		rows.push([
			'participant',
			participant.submitted_at ? 'submitted' : 'heartbeat',
			participant.last_heartbeat ?? '',
			participant.nama,
			participant.nis,
			`answered=${participant.answered_count}; app_switch=${participant.app_switch_count}; screenshot=${participant.screenshot_attempt}; suspicious=${participant.suspicious_flag ? 'yes' : 'no'}`
		]);
	}

	for (const event of input.events) {
		rows.push([
			'event',
			classifyProctorEvent(event) ?? 'other',
			event.created_at ?? '',
			event.nama ?? '',
			event.nis ?? '',
			proctorEventDetail(event)
		]);
	}

	return rows.map((row) => row.map((cell) => safeEvidenceCsvCell(cell)));
}

function proctorEventDetail(event: ProctorEvidenceEvent): string {
	const data = eventDataRecord(event.event_data);
	const reason = stringValue(data.reason);
	const parts = [proctorEventLabel(event)];
	if (reason && proctorEventLabel(event).toLowerCase() !== reason.toLowerCase()) parts.push(`reason=${reason}`);
	for (const key of [
		'state',
		'pending_count',
		'failure_count',
		'seconds_since_last_contact',
		'actor',
		'token',
		'access_token',
		'refresh_token',
		'exam_token',
		'password',
		'api_key',
		'secret',
		'authorization',
		'bearer',
		'device_fingerprint'
	]) {
		const value = data[key];
		if (value !== undefined && value !== null && value !== '') parts.push(`${key}=${redactEvidenceValue(key, value)}`);
	}
	return parts.join('; ');
}


const SENSITIVE_EVIDENCE_KEY_PATTERN = /(?:^|_)(?:token|access_token|refresh_token|exam_token|password|passwd|secret|api_key|bearer|authorization|device_fingerprint)(?:$|_)/i;
const SENSITIVE_EVIDENCE_VALUE_PATTERN = /\b(?:token|access_token|refresh_token|exam_token|password|passwd|secret|api[_-]?key|authorization|device_fingerprint)\s*[:=]\s*[^;\s,]+|\bBearer\s+[^;\s,]+/gi;
const CSV_INJECTION_PREFIX = /^[=+\-@\t\r\n]/;

function redactEvidenceValue(key: string, value: unknown): string {
	if (SENSITIVE_EVIDENCE_KEY_PATTERN.test(key)) return '[redacted]';
	return sanitizeSensitiveEvidenceString(String(value));
}

function safeEvidenceCsvCell(value: unknown): string {
	let cell = typeof value === 'string' ? value : String(value ?? '');
	cell = sanitizeSensitiveEvidenceString(cell);
	if (CSV_INJECTION_PREFIX.test(cell)) return `'${cell}`;
	return cell;
}

function sanitizeSensitiveEvidenceString(value: string): string {
	return value.replace(SENSITIVE_EVIDENCE_VALUE_PATTERN, (match) => {
		const separator = match.includes(':') ? ':' : match.includes('=') ? '=' : '';
		if (match.toLowerCase().startsWith('bearer ')) return 'Bearer [redacted]';
		const key = separator ? match.slice(0, match.indexOf(separator)).trim() : match;
		return `${key}${separator}[redacted]`;
	});
}

function emptyCounts(): Record<ProctorEvidenceCategory, number> {
	return Object.fromEntries(PROCTOR_EVIDENCE_CATEGORIES.map((category) => [category, 0])) as Record<ProctorEvidenceCategory, number>;
}

function isParticipantStale(participant: ProctorEvidenceParticipant, now: Date, staleAfterMinutes: number): boolean {
	if (participant.submitted_at || !participant.last_heartbeat) return false;
	const heartbeat = new Date(participant.last_heartbeat);
	if (Number.isNaN(heartbeat.getTime())) return false;
	return now.getTime() - heartbeat.getTime() > staleAfterMinutes * 60_000;
}

function eventDataRecord(value: unknown): EventDataRecord {
	if (typeof value === 'string') {
		const parsed = parseEventDataString(value);
		if (parsed) return parsed;
	}
	if (!value || typeof value !== 'object' || Array.isArray(value)) return {};
	return value as EventDataRecord;
}

function stringValue(value: unknown): string {
	return typeof value === 'string' ? value.trim() : '';
}

function parseEventDataString(value: string): EventDataRecord | null {
	const trimmed = value.trim();
	if (!trimmed) return null;
	const direct = parseJSONRecord(trimmed);
	if (direct) return direct;
	try {
		const decoded = globalThis.atob(trimmed);
		return parseJSONRecord(decoded);
	} catch {
		return null;
	}
}

function parseJSONRecord(value: string): EventDataRecord | null {
	try {
		const parsed: unknown = JSON.parse(value);
		if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) return parsed as EventDataRecord;
	} catch {
		return null;
	}
	return null;
}
