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
	heartbeat: 'Koneksi aktif',
	app_background_resume: 'Aplikasi ditinggalkan',
	device_mismatch: 'Perangkat tidak sesuai',
	submit_guard: 'Pengiriman ditahan',
	stale_connection: 'Koneksi perlu dicek',
	warning: 'Peringatan pengawas',
	anti_cheat: 'Pelanggaran tata tertib aplikasi',
	force_submit: 'Jawaban dikirim oleh pengawas',
	reset_access: 'Reset akses',
	export_print: 'Dokumen pengawasan dicetak'
};

const PROCTOR_EVIDENCE_CATEGORY_SUMMARIES: Record<ProctorEvidenceCategory, string> = {
	heartbeat: 'Peserta masih tercatat terhubung dengan server ujian.',
	app_background_resume: 'Aplikasi ujian sempat tidak menjadi fokus; verifikasi siswa tetap mengikuti arahan ruang.',
	device_mismatch: 'Perangkat berbeda dari catatan sesi; reset akses hanya setelah verifikasi identitas.',
	submit_guard: 'Pengiriman jawaban ditahan karena sinkronisasi belum aman.',
	stale_connection: 'Kontak perangkat terlambat; pengawas perlu cek jaringan atau perangkat.',
	warning: 'Peringatan aplikasi atau catatan manual dari proses ujian.',
	anti_cheat: 'Aplikasi mencatat perilaku yang perlu ditindaklanjuti pengawas.',
	force_submit: 'Pengiriman jawaban oleh pengawas sudah tercatat sebagai tindakan resmi.',
	reset_access: 'Akses perangkat direset oleh pengawas setelah verifikasi ruang.',
	export_print: 'Bukti cetak atau unduhan tersedia tanpa data rahasia.'
};

export const proctorOperatorGuidance: ProctorOperatorGuidanceItem[] = [
	{
		title: 'Kapan memperingatkan siswa',
		description: 'Saat aplikasi ditinggalkan, koneksi mulai terlambat, atau ada percobaan tangkap layar/keluar aplikasi berulang.'
	},
	{
		title: 'Kapan reset akses',
		description: 'Saat perangkat sah perlu login ulang setelah pengawas memverifikasi identitas, ruang, dan alasan gangguan.'
	},
	{
		title: 'Kapan paksa kirim',
		description: 'Saat ruang harus ditutup dan pengawas sudah memastikan jawaban tersinkron atau prosedur manual dicatat.'
	}
];

export function proctorEvidenceCategoryLabel(category: ProctorEvidenceCategory | null): string {
	return category ? PROCTOR_EVIDENCE_CATEGORY_LABELS[category] : 'Kejadian lain';
}

export function proctorEvidenceCategorySummary(category: ProctorEvidenceCategory | null): string {
	return category ? PROCTOR_EVIDENCE_CATEGORY_SUMMARIES[category] : 'Kejadian tidak masuk kategori utama bukti pengawasan.';
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
	resume_exam: 'Masuk kembali ke ujian',
	repeat_resume_attempt: 'Masuk kembali berulang',
	answer_saved_local_only: 'Jawaban lokal',
	submit_blocked_pending_sync: 'Pengiriman ditahan: menunggu sinkronisasi',
	auto_submit_blocked_pending_sync: 'Pengiriman otomatis ditahan',
	submit_blocked_degraded_mode: 'Pengiriman ditahan: koneksi menurun',
	degraded_mode_entered: 'Koneksi menurun',
	stale_connection_attention: 'Koneksi perlu dicek',
	stale_connection_escalated: 'Koneksi perlu tindakan segera',
	back_button_attempt: 'Tombol kembali',
	manual_submit: 'Pengiriman manual',
	device_mismatch: 'Perangkat tidak sesuai',
	token_already_bound: 'Perangkat tidak sesuai'
};

const PROCTOR_EVENT_TYPE_LABELS: Record<string, string> = {
	login: 'Masuk ujian',
	heartbeat: 'Koneksi aktif',
	app_switch: 'Keluar/kembali aplikasi',
	app_backgrounded: 'Aplikasi ditinggalkan',
	app_resumed: 'Kembali ke aplikasi ujian',
	resume: 'Kembali ujian',
	device_mismatch: 'Perangkat tidak sesuai',
	answer: 'Simpan jawaban',
	answer_save: 'Simpan jawaban',
	submit: 'Kirim jawaban',
	submit_guard: 'Pengiriman ditahan',
	stale_connection: 'Koneksi perlu dicek',
	heartbeat_failed: 'Koneksi gagal',
	screenshot_attempt: 'Percobaan tangkap layar',
	focus_lost_short: 'Fokus aplikasi berpindah',
	app_switch_once: 'Aplikasi berpindah',
	app_switch_repeated: 'Aplikasi berpindah berulang',
	background_over_threshold: 'Aplikasi ditinggalkan terlalu lama',
	split_screen_detected: 'Layar terbagi',
	pip_detected: 'Jendela mengambang',
	overlay_suspicious_confirmed: 'Tampilan mencurigakan',
	screenshot_attempt_ambiguous: 'Indikasi tangkap layar',
	screenshot_attempt_valid: 'Percobaan tangkap layar tervalidasi',
	device_mismatch_weak: 'Perangkat tidak sesuai',
	device_mismatch_strong: 'Perangkat tidak sesuai kuat',
	token_reuse_confirmed: 'Akses ujian digunakan ulang',
	root_emulator_weak: 'Integritas perangkat lemah',
	root_emulator_strong: 'Integritas perangkat kuat',
	offline_short: 'Kontak perangkat terlambat',
	offline_mass: 'Gangguan teknis massal',
	pending_sync: 'Jawaban belum terkirim',
	submit_held_pending_sync: 'Pengiriman ditahan: menunggu sinkronisasi',
	web_fallback_used: 'Browser darurat digunakan',
	web_visibility_hidden: 'Browser tidak terlihat',
	web_visibility_visible: 'Browser terlihat kembali',
	web_focus_lost: 'Fokus browser berpindah',
	web_focus_restored: 'Fokus browser kembali',
	web_fullscreen_exit: 'Layar penuh browser keluar',
	web_fullscreen_restored: 'Layar penuh browser kembali',
	web_pending_answer_saved: 'Jawaban browser lokal',
	web_pending_answer_flushed: 'Jawaban browser terkirim',
	web_connection_degraded: 'Koneksi browser menurun',
	web_connection_restored: 'Koneksi browser pulih',
	proctor_reset_access: 'Reset akses',
	proctor_unlock: 'Buka kunci peserta',
	proctor_acknowledge: 'Tandai diperiksa',
	proctor_incident_action: 'Tindak lanjut insiden',
	participant_command: 'Instruksi ke aplikasi siswa',
	participant_command_ack: 'Instruksi diterima aplikasi siswa',
	student_portal_token_reveal: 'Akses ujian siswa dibuka portal',
	student_portal_room_token_mismatch: 'Kode ruang salah di portal',
	exam_room_token_mismatch: 'Kode ruang salah di aplikasi ujian',
	proctor_heartbeat: 'Koneksi pengawas aktif',
	anti_cheat_violation: 'Pelanggaran tata tertib aplikasi',
	proctor_force_submit: 'Jawaban dikirim oleh pengawas'
};

const DETAIL_KEY_LABELS: Record<string, string> = {
	action: 'Tindakan',
	status: 'Status',
	reason: 'Alasan',
	notes: 'Catatan',
	message: 'Pesan',
	command_type: 'Instruksi',
	state: 'Kondisi',
	pending_count: 'Belum tersinkron',
	failure_count: 'Gagal sinkron',
	seconds_since_last_contact: 'Detik sejak kontak terakhir',
	actor: 'Petugas'
};

const FORMAL_VALUE_LABELS: Record<string, string> = {
	paused: 'Aplikasi tidak aktif',
	resumed: 'Aplikasi aktif kembali',
	backgrounded: 'Aplikasi ditinggalkan',
	locked: 'Terkunci',
	high: 'Risiko tinggi',
	warning: 'Perlu perhatian',
	normal: 'Normal',
	reviewed: 'Sudah diperiksa',
	cleared: 'Selesai',
	warning_given: 'Peringatan diberikan',
	escalated: 'Dieskalasi',
	admin: 'Operator/admin',
	proctor: 'Pengawas',
	warning_message: 'Peringatan siswa',
	reconnect: 'Instruksi masuk ulang',
	unlock_notice: 'Pemberitahuan akses dibuka',
	split_screen: 'Layar terbagi',
	picture_in_picture: 'Jendela mengambang',
	focus_lost: 'Fokus aplikasi hilang',
	screenshot_attempt: 'Percobaan tangkap layar',
	app_switch: 'Keluar/kembali aplikasi',
	anti_cheat_locked: 'Akses dikunci karena pelanggaran'
};

export function classifyProctorEvent(event: ProctorEvidenceEvent): ProctorEvidenceCategory | null {
	const eventType = event.event_type.trim().toLowerCase();
	const data = eventDataRecord(event.event_data);
	const reason = stringValue(data.reason).toLowerCase();

	if (eventType === 'heartbeat' || eventType === 'proctor_heartbeat' || eventType === 'web_visibility_visible' || eventType === 'web_focus_restored' || eventType === 'web_fullscreen_restored' || eventType === 'web_pending_answer_flushed' || eventType === 'web_connection_restored') return 'heartbeat';
	if (
		eventType === 'app_switch' ||
		eventType === 'app_switch_once' ||
		eventType === 'app_switch_repeated' ||
		eventType === 'app_backgrounded' ||
		eventType === 'app_resumed' ||
		eventType === 'resume' ||
		eventType === 'focus_lost_short' ||
		eventType === 'background_over_threshold' ||
		eventType === 'web_visibility_hidden' ||
		eventType === 'web_focus_lost' ||
		eventType === 'web_fullscreen_exit'
	) {
		return 'app_background_resume';
	}
	if (
		eventType === 'device_mismatch' ||
		eventType === 'device_mismatch_weak' ||
		eventType === 'device_mismatch_strong' ||
		eventType === 'token_reuse_confirmed' ||
		eventType === 'exam_room_token_mismatch' ||
		eventType === 'student_portal_room_token_mismatch'
	) return 'device_mismatch';
	if (eventType === 'student_portal_token_reveal') return 'export_print';
	if (eventType === 'submit_guard' || eventType === 'submit_held_pending_sync' || eventType === 'pending_sync' || eventType === 'web_pending_answer_saved') return 'submit_guard';
	if (
		eventType === 'anti_cheat_violation' ||
		eventType === 'split_screen_detected' ||
		eventType === 'pip_detected' ||
		eventType === 'overlay_suspicious_confirmed' ||
		eventType === 'screenshot_attempt' ||
		eventType === 'screenshot_attempt_ambiguous' ||
		eventType === 'screenshot_attempt_valid' ||
		eventType === 'root_emulator_weak' ||
		eventType === 'root_emulator_strong'
	) return 'anti_cheat';
	if (eventType === 'stale_connection' || eventType === 'heartbeat_failed' || eventType === 'offline_short' || eventType === 'offline_mass' || eventType === 'web_connection_degraded') return 'stale_connection';
	if (eventType === 'proctor_force_submit') return 'force_submit';
	if (eventType === 'proctor_reset_access' || eventType === 'proctor_unlock') return 'reset_access';
	if (eventType === 'proctor_acknowledge' || eventType === 'proctor_incident_action' || eventType === 'participant_command') return 'warning';
	if (eventType === 'proctor_heartbeat') return 'heartbeat';

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
	if (eventType === 'warning' && reason) return proctorIncidentReasonLabel(reason);
	if (eventType === 'anti_cheat_violation' && reason) return proctorIncidentReasonLabel(reason);
	return PROCTOR_EVENT_TYPE_LABELS[eventType] ?? 'Kejadian pengawasan';
}

export function proctorEventReasonLabel(event: ProctorEvidenceEvent): string {
	const data = eventDataRecord(event.event_data);
	const reason = stringValue(data.reason);
	return reason ? proctorIncidentReasonLabel(reason) : proctorEventLabel(event);
}

export function proctorIncidentReasonLabel(value: string | null | undefined): string {
	if (!value) return 'Kejadian pengawasan';
	const normalized = value.trim().toLowerCase();
	if (!normalized) return 'Kejadian pengawasan';
	const redacted = sanitizeSensitiveEvidenceString(normalized);
	if (redacted.includes('[redacted]')) return 'Kejadian pengawasan';
	return WARNING_REASON_LABELS[normalized] ?? FORMAL_VALUE_LABELS[normalized] ?? 'Kejadian pengawasan';
}

export function proctorIncidentActionLabel(value: string | null | undefined): string {
	if (!value) return 'Tindakan pengawas';
	const normalized = value.trim().toLowerCase();
	return FORMAL_VALUE_LABELS[normalized] ?? 'Tindakan pengawas';
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

export function proctorEventDetail(event: ProctorEvidenceEvent): string {
	const data = eventDataRecord(event.event_data);
	const parts = [proctorEventLabel(event)];
	for (const key of [
		'action',
		'status',
		'reason',
		'notes',
		'message',
		'command_type',
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
		if (value !== undefined && value !== null && value !== '') parts.push(formatEvidenceDetail(key, value));
	}
	return parts.join('; ');
}

function formatEvidenceDetail(key: string, value: unknown): string {
	const label = DETAIL_KEY_LABELS[key] ?? 'Detail';
	const normalized = typeof value === 'string' ? value.trim().toLowerCase() : '';
	if (key === 'reason') return `${label}: ${proctorIncidentReasonLabel(String(value))}`;
	if (key === 'action') return `${label}: ${proctorIncidentActionLabel(String(value))}`;
	if (key === 'status' || key === 'command_type' || key === 'state' || key === 'actor') {
		return `${label}: ${FORMAL_VALUE_LABELS[normalized] ?? sanitizeSensitiveEvidenceString(String(value))}`;
	}
	return `${label}: ${redactEvidenceValue(key, value)}`;
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
