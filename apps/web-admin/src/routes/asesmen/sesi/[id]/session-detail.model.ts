export type ActiveTab = 'hasil' | 'butir' | 'peserta' | 'ruangan' | 'operasional' | 'proctoring' | 'audit' | 'essay';
export type DetailNextAction = {
	title: string;
	message: string;
	label: string;
	tab?: ActiveTab;
	run?: 'auto_seats';
	tone: 'success' | 'warning' | 'info';
};
export type DetailTabGroup = {
	module: string;
	help: string;
	tabs: { id: ActiveTab; label: string }[];
};
export type RoomLike = {
	capacity: number;
	participant_count: number;
};
export type RoomReadinessLike = {
	participant_count: number;
	room_count: number;
	total_capacity: number;
	unassigned_participant_count: number;
	missing_seat_count: number;
	rooms_without_proctor: number;
};
export type OperationalRecapLike = {
	room_count: number;
	handover_locked_count: number;
	handover_missing_count: number;
	incident_room_count: number;
	participant_count: number;
	unassigned_participant_count: number;
	submitted_count: number;
	incident_event_count: number;
	force_submit_count: number;
};
export type OperationalRoomLike = {
	handover_id: string | null;
	locked_at: string | null;
};
export type ItemAnalysisRowLike = {
	answer_distribution: unknown;
	recommendation_tone: string;
};
export type AuditLogLike = {
	action: string;
	metadata: unknown;
};

export const statusLabel: Record<string, string> = {
	draft: 'Draft',
	scheduled: 'Terjadwal',
	active: 'Berlangsung',
	finished: 'Selesai',
	cancelled: 'Dibatalkan'
};

export const detailTabGroups: DetailTabGroup[] = [
	{
		module: 'Setup Panitia',
		help: 'Peserta dan ruang sebelum ujian',
		tabs: [
			{ id: 'peserta', label: 'Peserta' },
			{ id: 'ruangan', label: 'Ruang' }
		]
	},
	{
		module: 'Hari-H',
		help: 'Pantau ruang dan kejadian',
		tabs: [
			{ id: 'proctoring', label: 'Pengawasan' },
			{ id: 'operasional', label: 'Serah Terima' },
			{ id: 'audit', label: 'Log Tindakan' }
		]
	},
	{
		module: 'Penutupan',
		help: 'Nilai, analisis, uraian',
		tabs: [
			{ id: 'hasil', label: 'Hasil' },
			{ id: 'butir', label: 'Analisis Butir' },
			{ id: 'essay', label: 'Uraian' }
		]
	}
];

export const detailTabs = detailTabGroups.flatMap((group) => group.tabs);

export function tabFromQuery(value: string | null): ActiveTab | null {
	const found = detailTabs.find((tab) => tab.id === value);
	return found?.id ?? null;
}

export function statusClass(status: string) {
	if (status === 'active') return 'bg-primary/15 text-primary border-primary/20';
	if (status === 'finished') return 'bg-muted text-muted-foreground border-border';
	if (status === 'cancelled') return 'bg-destructive/15 text-destructive border-destructive/30';
	if (status === 'scheduled') return 'bg-success/15 text-success border-success/20';
	return 'bg-warning/15 text-warning border-warning/30';
}

export function fmtDt(iso: string | null) {
	if (!iso) return '—';
	return `${new Date(iso).toLocaleString('id-ID', {
		timeZone: 'Asia/Makassar',
		year: 'numeric',
		month: 'short',
		day: 'numeric',
		hour: '2-digit',
		minute: '2-digit'
	})} WITA`;
}

export function fmtScore(score: string | null) {
	if (score === null || score === undefined || score === '') return '—';
	const value = parseFloat(score);
	return Number.isNaN(value) ? '—' : value.toFixed(1);
}

export function roomSetupLocked(status: string | null | undefined) {
	return status === 'active' || status === 'finished';
}

export function roomCapacityRatio(room: RoomLike) {
	if (room.capacity <= 0) return 0;
	return Math.min(100, (room.participant_count / room.capacity) * 100);
}

export function nextDetailAction(readiness: RoomReadinessLike | null): DetailNextAction {
	if (!readiness) {
		return {
			title: 'Cek kesiapan operasional',
			message: 'Muat data ruang, peserta, kapasitas, nomor meja, dan pengawas sebelum sesi dimulai.',
			label: 'Cek Kesiapan',
			tab: 'ruangan',
			tone: 'info'
		};
	}
	if (readiness.participant_count === 0) {
		return {
			title: 'Peserta belum terdaftar',
			message: 'Daftarkan peserta dari daftar sesi, lalu kembali untuk mengatur ruangan dan pengawas.',
			label: 'Lihat Peserta',
			tab: 'peserta',
			tone: 'warning'
		};
	}
	if (readiness.room_count === 0 || readiness.total_capacity < readiness.participant_count) {
		return {
			title: 'Ruang ujian belum cukup',
			message: `${readiness.room_count} ruang tersedia dengan kapasitas ${readiness.total_capacity} untuk ${readiness.participant_count} peserta.`,
			label: 'Atur Ruang',
			tab: 'ruangan',
			tone: 'warning'
		};
	}
	if (readiness.unassigned_participant_count > 0) {
		return {
			title: 'Peserta belum masuk ruang',
			message: `${readiness.unassigned_participant_count} peserta belum punya ruang ujian.`,
			label: 'Acak Ruang',
			tab: 'ruangan',
			tone: 'warning'
		};
	}
	if (readiness.missing_seat_count > 0) {
		return {
			title: 'Nomor meja belum lengkap',
			message: `${readiness.missing_seat_count} peserta belum punya nomor meja.`,
			label: 'Atur Nomor Meja',
			run: 'auto_seats',
			tone: 'warning'
		};
	}
	if (readiness.rooms_without_proctor > 0) {
		return {
			title: 'Pengawas belum lengkap',
			message: `${readiness.rooms_without_proctor} ruang belum punya pengawas.`,
			label: 'Tetapkan Pengawas',
			tab: 'ruangan',
			tone: 'warning'
		};
	}
	return {
		title: 'Sesi siap dipantau',
		message: 'Peserta, ruang, kapasitas, nomor meja, dan pengawas sudah siap.',
		label: 'Buka Pengawasan',
		tab: 'proctoring',
		tone: 'success'
	};
}

export function detailActionPanelClass(tone: DetailNextAction['tone']) {
	if (tone === 'warning') return 'border-warning/30 bg-warning/10';
	if (tone === 'success') return 'border-primary/20 bg-primary/10';
	return 'border-border bg-muted/50';
}

export function operationalTone(recap: OperationalRecapLike | null) {
	if (!recap) return 'info';
	if (recap.handover_missing_count > 0 || recap.unassigned_participant_count > 0) return 'warning';
	if (recap.incident_room_count > 0 || recap.incident_event_count > 0 || recap.force_submit_count > 0) return 'warning';
	return 'success';
}

export function operationalMessage(recap: OperationalRecapLike | null) {
	if (!recap) return 'Rekap operasional belum dimuat.';
	return `${recap.handover_locked_count}/${recap.room_count} ruang sudah mengunci handover, ${recap.submitted_count}/${recap.participant_count} peserta submit, ${recap.incident_room_count} ruang punya catatan insiden, ${recap.force_submit_count} peserta dipaksa submit.`;
}

function isRecord(value: unknown): value is Record<string, unknown> {
	return typeof value === 'object' && value !== null;
}

const sensitiveEvidenceValuePattern = /\b(?:authorization)\s*[:=]\s*(?:Bearer\s+)?[^;\s,]+|\bBearer\s+[^;\s,]+|\b(?:token|access_token|refresh_token|exam_token|password|passwd|secret|api[_-]?key|api\s+key|device[_-]?fingerprint|deviceFingerprint)\s*[:=]\s*[^;\s,]+/gi;

export function sanitizeOperationalEvidenceNote(value: unknown) {
	if (value === null || value === undefined) return '';
	return String(value).replace(sensitiveEvidenceValuePattern, (match) => {
		if (match.toLowerCase().startsWith('bearer ')) return 'Bearer [redacted]';
		const colonIndex = match.indexOf(':');
		const equalsIndex = match.indexOf('=');
		const separatorIndex = colonIndex === -1
			? equalsIndex
			: equalsIndex === -1
				? colonIndex
				: Math.min(colonIndex, equalsIndex);
		if (separatorIndex === -1) return '[redacted]';
		const key = match.slice(0, separatorIndex).trim();
		return `${key}${match[separatorIndex]}[redacted]`;
	});
}

export function percent(value: number) {
	if (!Number.isFinite(value)) return '0%';
	return `${Math.round(value * 100)}%`;
}

export function questionTypeLabel(type: string) {
	const labels: Record<string, string> = {
		multiple_choice: 'PG',
		multiple_answer: 'PG Kompleks',
		true_false: 'Benar/Salah',
		agree_disagree: 'Setuju/Tidak',
		matching: 'Menjodohkan',
		short_answer: 'Isian',
		essay: 'Essay'
	};
	return labels[type] ?? type.replaceAll('_', ' ');
}

export function itemAnalysisToneClass(tone: string) {
	if (tone === 'success') return 'border-primary/20 bg-primary/10 text-primary';
	if (tone === 'danger') return 'border-destructive/30 bg-destructive/10 text-destructive';
	if (tone === 'info') return 'border-border bg-muted/50 text-muted-foreground';
	return 'border-warning/30 bg-warning/10 text-warning';
}

export function itemAnalysisSignalClass(row: ItemAnalysisRowLike) {
	if (row.recommendation_tone === 'danger') return 'bg-destructive/10';
	if (row.recommendation_tone === 'warning') return 'bg-warning/10';
	return '';
}

export function answerDistributionEntries(row: ItemAnalysisRowLike) {
	if (!isRecord(row.answer_distribution)) return [];
	return Object.entries(row.answer_distribution)
		.map(([answer, count]) => ({ answer, count: Number(count) || 0 }))
		.sort((a, b) => b.count - a.count || a.answer.localeCompare(b.answer));
}

export function shouldOfferQuestionRevision(row: ItemAnalysisRowLike) {
	return row.recommendation_tone === 'warning' || row.recommendation_tone === 'danger';
}

export function handoverStatusLabel(room: OperationalRoomLike) {
	if (room.locked_at) return 'Terkunci';
	if (room.handover_id) return 'Draft';
	return 'Belum ada';
}

export function handoverStatusClass(room: OperationalRoomLike) {
	if (room.locked_at) return 'border-primary/20 bg-primary/10 text-primary';
	if (room.handover_id) return 'border-warning/30 bg-warning/10 text-warning';
	return 'border-destructive/30 bg-destructive/10 text-destructive';
}

export function fmtEssayPoints(points: unknown) {
	if (typeof points === 'number') return points.toFixed(points % 1 === 0 ? 0 : 1);
	if (typeof points === 'string' && points.trim()) return points;
	return '1';
}

export function scoreClass(score: string | null) {
	if (!score) return 'text-muted-foreground';
	const value = parseFloat(score);
	if (value >= 75) return 'text-primary font-semibold';
	if (value >= 60) return 'text-warning font-semibold';
	return 'text-destructive font-semibold';
}

export function heartbeatBucket(hb: string | null): 'none' | 'online' | 'slow' | 'offline' {
	if (!hb) return 'none';
	const diff = (Date.now() - new Date(hb).getTime()) / 1000;
	if (diff < 60) return 'online';
	if (diff < 180) return 'slow';
	return 'offline';
}

export function heartbeatStatus(hb: string | null): { label: string; cls: string } {
	const bucket = heartbeatBucket(hb);
	if (bucket === 'online') return { label: 'Online', cls: 'text-primary' };
	if (bucket === 'slow') return { label: 'Lambat', cls: 'text-warning' };
	if (bucket === 'offline') return { label: 'Offline', cls: 'text-destructive' };
	return { label: 'Belum login', cls: 'text-muted-foreground' };
}

export function proctoringEventLabel(type: string): string {
	const map: Record<string, string> = {
		login: 'Login',
		heartbeat: 'Heartbeat',
		answer: 'Jawaban',
		submit: 'Submit',
		app_switch: 'Pindah Aplikasi',
		screenshot_attempt: 'Screenshot',
		copy_attempt: 'Salin Teks',
		paste_attempt: 'Tempel Teks',
		cut_attempt: 'Potong Teks',
		proctor_reset_access: 'Reset Akses',
		proctor_force_submit: 'Paksa Kirim Jawaban'
	};
	return map[type] ?? type.replaceAll('_', ' ');
}

export function eventDataText(data: unknown): string {
	if (data === null || data === undefined || data === '') return '—';
	const text = typeof data === 'string' ? data : JSON.stringify(data);
	if (!text) return '—';
	return text.length > 140 ? `${text.slice(0, 140)}...` : text;
}

export function auditMetadata(meta: unknown): Record<string, unknown> {
	if (typeof meta === 'object' && meta !== null && !Array.isArray(meta)) return meta as Record<string, unknown>;
	if (typeof meta === 'string' && meta.trim()) {
		try {
			const parsed = JSON.parse(meta);
			if (typeof parsed === 'object' && parsed !== null && !Array.isArray(parsed)) return parsed as Record<string, unknown>;
		} catch {
			return {};
		}
	}
	return {};
}

export function auditValue(log: AuditLogLike, key: string): string {
	const value = auditMetadata(log.metadata)[key];
	return typeof value === 'string' ? value : '';
}

export function auditActionLabel(action: string): string {
	const labels: Record<string, string> = {
		CBT_SESSION_SCHEDULE_UPDATE: 'Ubah Jadwal',
		CBT_SESSION_PARTICIPANT_FLAG: 'Flag Peserta',
		CBT_SESSION_PARTICIPANT_RESET_ACCESS: 'Reset Akses',
		CBT_SESSION_PARTICIPANT_FORCE_SUBMIT: 'Paksa Kirim Jawaban',
		CBT_SESSION_ESSAY_GRADE: 'Koreksi Uraian',
		CBT_SESSION_SCORE: 'Hitung Skor',
		CBT_SESSION_ROOM_HANDOVER_SAVE: 'Simpan Handover',
		CBT_SESSION_ROOM_HANDOVER_LOCK: 'Kunci Handover'
	};
	return labels[action] ?? action.replaceAll('_', ' ');
}

export function auditSummary(log: AuditLogLike): string {
	if (log.action === 'CBT_SESSION_SCHEDULE_UPDATE') {
		const beforeStart = auditValue(log, 'previous_scheduled_start');
		const beforeEnd = auditValue(log, 'previous_scheduled_end');
		const nextStart = auditValue(log, 'new_scheduled_start');
		const nextEnd = auditValue(log, 'new_scheduled_end');
		return `${fmtDt(beforeStart)} - ${fmtDt(beforeEnd)} -> ${fmtDt(nextStart)} - ${fmtDt(nextEnd)}`;
	}
	const participantId = auditValue(log, 'participant_id');
	const roomId = auditValue(log, 'room_id');
	const answerId = auditValue(log, 'answer_id');
	return [participantId && `Peserta ${participantId}`, roomId && `Ruang ${roomId}`, answerId && `Jawaban ${answerId}`]
		.filter(Boolean)
		.join(' · ') || 'Detail tersedia di riwayat pemeriksaan.';
}
