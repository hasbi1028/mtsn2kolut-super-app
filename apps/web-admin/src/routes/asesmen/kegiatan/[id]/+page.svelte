<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Badge } from '$lib/components/ui/badge';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { ContextStrip, MetricCard } from '$lib/components/ops';
	import { AssessmentPhaseHeader, AssessmentTaskCard } from '$lib/components/asesmen';
	import { sopStages, sopStatusLabels, type SopReadinessResponse, type SopStageKey, type SopStageReadiness, type SopStageStatus } from '$lib/asesmen/sop-stages';
	import { assessmentApprovalLabels, createApproval, listApprovals, revokeApproval, type AssessmentApprovalRecord, type AssessmentApprovalType } from '$lib/asesmen/approval-client';
	import { clientApiPath, readClientApiData } from '$lib/client/api';
	import { csvRow } from '$lib/csv';
	import { toast } from '$lib/components/ui/sonner';

	type EventInfo = {
		id: string; title: string; exam_type: string; scope: string;
		target_levels?: string[];
		academic_year_name: string; status: string;
	};
	type ResultRow = {
		participant_id: string; session_id: string; session_title: string;
		nis: string; student_nama: string; gender: string;
		class_code: string; score: string | null; submitted_at: string | null;
	};
	type EventSession = {
		id: string; title: string; status?: string; participant_count?: number; room_count?: number;
		missing_seat_count?: number; rooms_without_proctor?: number; unassigned_participant_count?: number;
	};
	type EventPackage = { id: string; title: string; question_count?: number; is_active?: boolean };
	type QuestionCompletenessRow = {
		level: string; class_id: string; class_code: string; class_name: string;
		subject_id: string; subject_name: string; subject_code: string;
		teacher_employee_id: string; teacher_name: string; teacher_username: string;
		scope_mode?: string; status_filter?: string;
		target_pg: number; available_pg: number; missing_pg: number;
		target_essay: number; available_essay: number; missing_essay: number;
		complete: boolean;
	};
	type QuestionRequirement = {
		scope_mode: 'per_rombel' | 'per_level' | 'pool_level_subject' | string;
		target_pg: number;
		target_essay: number;
		status_filter: 'published_only' | 'all_progress' | string;
	};
	type QuestionCompleteness = {
		requirements?: QuestionRequirement;
		summary: {
			total_rows: number; complete_rows: number; incomplete_rows: number;
			target_pg: number; available_pg: number; missing_pg: number;
			target_essay: number; available_essay: number; missing_essay: number;
		};
		rows: QuestionCompletenessRow[];
		contributions?: Array<{ level: string; subject_id: string; subject_name: string; teacher_name: string; teacher_username: string; available_pg: number; available_essay: number }>;
		excluded_levels?: string[];
	};
	type EventOverview = {
		member_count?: number; target_count?: number; review_count?: number; question_count?: number; published_question_count?: number;
		package_count?: number; session_count?: number; room_count?: number; token_count?: number; card_count?: number; result_count?: number;
	};
	type EventCommandDetail = {
		info: EventInfo;
		results: ResultRow[];
		overview: EventOverview | null;
		sessions: EventSession[];
		packages: EventPackage[];
		questionCompleteness: QuestionCompleteness | null;
		sopReadiness: SopReadinessResponse | null;
		sopDetail: SopDetail | null;
		sopTransitions: SopTransitionRecord[];
		approvals: AssessmentApprovalRecord[];
		approvalsAvailable: boolean;
	};
	type ChecklistHref =
		| `/asesmen/kegiatan/${string}/members`
		| `/bank-soal/tambah?${string}`
		| '/bank-soal/tambah'
		| '/bank-soal/verifikasi'
		| `/asesmen/paket?event_id=${string}`
		| `/asesmen/sesi?event_id=${string}`
		| `/asesmen/sesi?event_id=${string}&readiness=not_ready`
		| `/asesmen/sesi?event_id=${string}&readiness=needs_rooms`
		| `/asesmen/sesi?event_id=${string}&readiness=needs_proctors`
		| `/asesmen/kegiatan/${string}/cetak`
		| `/asesmen/kegiatan/${string}/exam-cards`
		| `/asesmen/kegiatan/${string}/pengawas-cards`
		| `/asesmen/kegiatan/${string}#hasil`
		| `/asesmen/kegiatan/${string}/archive`;
	type ChecklistItem = {
		label: string; helper: string; count: number | null; href: ChecklistHref; tone: 'success' | 'warning' | 'info'; action: string;
	};
	type EventSection = 'ringkasan' | 'persiapan' | 'pelaksanaan' | 'hasil' | 'arsip';
	type ReadinessGroup = {
		id: EventSection;
		title: string;
		description: string;
		items: ChecklistItem[];
	};
	type NextAction = ChecklistItem & { priority: string };
	type EventApprovalType = Extract<AssessmentApprovalType, 'package_ready' | 'participants_rooms_ready' | 'tokens_cards_ready' | 'results_verified' | 'final_archive'>;
	type SopApprovalMilestone = {
		approvalType: EventApprovalType;
		stageKey: SopStageKey;
		helper: string;
	};
	type SopTransitionRecord = {
		id?: string;
		from_stage?: string | null;
		to_stage: string;
		note?: string | null;
		actor?: string | null;
		created_at?: string | null;
	};
	type SopDetail = {
		event_id?: string;
		current_stage?: string;
		current_stage_label?: string;
		next_action?: string | { label?: string; href?: string } | null;
		blockers?: string[];
		warnings?: string[];
		transitions?: SopTransitionRecord[];
		report_only?: boolean;
	};

	const eventId = page.params.id ?? '';
	let info = $state<EventInfo | null>(null);
	let results = $state<ResultRow[]>([]);
	let detailPromise = $state<Promise<EventCommandDetail> | null>(null);
	let hasilSectionElement = $state<HTMLElement | null>(null);
	let hasilFocusRequest = $state(0);
	let detailRequestId = 0;
	const mainSections = [
		{ id: 'persiapan', code: '7.1', label: 'Persiapan', href: '/asesmen/persiapan', withEvent: true },
		{ id: 'pelaksanaan', code: '7.2', label: 'Pelaksanaan', href: '/asesmen/pelaksanaan', withEvent: true },
		{ id: 'hasil', code: '7.3', label: 'Hasil', href: '/asesmen/hasil', withEvent: false },
	] as const;
	let completenessLevel = $state('');
	let completenessStatus = $state('');
	let completenessSearch = $state('');
	let targetScopeMode = $state<'per_rombel' | 'per_level' | 'pool_level_subject'>('per_rombel');
	let targetStatusFilter = $state<'published_only' | 'all_progress'>('published_only');
	let targetPg = $state(20);
	let targetEssay = $state(5);
	let targetSettingsBusy = $state(false);
	let approvalBusyType = $state<EventApprovalType | null>(null);
	let approvalNotes = $state<Record<EventApprovalType, string>>({
		package_ready: '',
		participants_rooms_ready: '',
		tokens_cards_ready: '',
		results_verified: '',
		final_archive: '',
	});

	const statusLabel: Record<string, string> = { draft: 'Konsep', active: 'Aktif', finished: 'Selesai' };
	const scopeLabel: Record<string, string> = { class: 'Per Kelas', grade: 'Per Tingkat', school: 'Seluruh Sekolah' };
	const questionRequirementScopeLabel: Record<string, string> = {
		per_rombel: 'Per rombel + mapel + guru',
		per_level: 'Per tingkat + mapel + guru',
		pool_level_subject: 'Kumpulan tingkat + mapel',
	};
	const questionRequirementStatusLabel: Record<string, string> = {
		published_only: 'Hanya soal terbit',
		all_progress: 'Konsep/verifikasi/terbit dihitung',
	};
	const sopApprovalMilestones: SopApprovalMilestone[] = [
		{ approvalType: 'package_ready', stageKey: 'package_ready', helper: 'Paket kegiatan siap dipakai dan isi soal sudah dicek panitia.' },
		{ approvalType: 'participants_rooms_ready', stageKey: 'participants_rooms_ready', helper: 'Peserta, ruang, nomor meja, dan pengawas sudah siap untuk hari-H.' },
		{ approvalType: 'tokens_cards_ready', stageKey: 'tokens_cards_ready', helper: 'Token ujian/kartu peserta siap didistribusikan secara terbatas.' },
		{ approvalType: 'results_verified', stageKey: 'result_verification', helper: 'Rekap hasil sudah diperiksa sebelum masuk laporan akhir.' },
		{ approvalType: 'final_archive', stageKey: 'final_archive', helper: 'Kegiatan ditutup dan dokumen arsip akhir siap dirujuk kembali.' },
	];

	function isRecord(value: unknown): value is Record<string, unknown> {
		return typeof value === 'object' && value !== null;
	}

	function parseArrayPayload<T>(payload: unknown, key: string): T[] {
		if (Array.isArray(payload)) return payload as T[];
		if (isRecord(payload) && Array.isArray(payload[key])) return payload[key] as T[];
		return [];
	}

	function isSopStageStatus(value: unknown): value is SopStageStatus {
		return value === 'ready' || value === 'warning' || value === 'blocked' || value === 'running';
	}

	function parseSopReadiness(payload: unknown): SopReadinessResponse | null {
		if (!isRecord(payload) || typeof payload.event_id !== 'string' || !Array.isArray(payload.stages)) return null;
		const stages: SopStageReadiness[] = payload.stages
			.filter((stage): stage is Record<string, unknown> => isRecord(stage) && typeof stage.key === 'string' && isSopStageStatus(stage.status))
			.map((stage) => ({
				key: stage.key as SopStageReadiness['key'],
				label: typeof stage.label === 'string' ? stage.label : String(stage.key),
				owner: typeof stage.owner === 'string' ? stage.owner : 'Panitia',
				description: typeof stage.description === 'string' ? stage.description : 'Tahap SOP kegiatan asesmen.',
				status: stage.status as SopStageStatus,
				blocking_count: Number(stage.blocking_count ?? 0),
				warning_count: Number(stage.warning_count ?? 0),
				next_actions: parseArrayPayload<{ label: string; href: string }>(stage.next_actions, 'next_actions')
					.filter((action) => typeof action.label === 'string' && typeof action.href === 'string'),
			}));
		return { event_id: payload.event_id, stages };
	}

	function parseStringList(value: unknown): string[] {
		if (!Array.isArray(value)) return [];
		return value.map((item) => typeof item === 'string' ? item : isRecord(item) && typeof item.label === 'string' ? item.label : '').filter(Boolean);
	}

	function parseSopTransition(value: unknown): SopTransitionRecord | null {
		if (!isRecord(value)) return null;
		const toStage = typeof value.to_stage === 'string'
			? value.to_stage
			: typeof value.to === 'string'
				? value.to
				: typeof value.stage === 'string'
					? value.stage
					: '';
		if (!toStage) return null;
		const actor = typeof value.actor === 'string'
			? value.actor
			: typeof value.actor_name === 'string'
				? value.actor_name
				: typeof value.created_by_display_name === 'string'
					? value.created_by_display_name
					: typeof value.created_by_username === 'string'
						? value.created_by_username
						: null;
		return {
			id: typeof value.id === 'string' ? value.id : undefined,
			from_stage: typeof value.from_stage === 'string' ? value.from_stage : typeof value.from === 'string' ? value.from : null,
			to_stage: toStage,
			note: typeof value.note === 'string' ? value.note : typeof value.notes === 'string' ? value.notes : null,
			actor,
			created_at: typeof value.created_at === 'string' ? value.created_at : typeof value.transitioned_at === 'string' ? value.transitioned_at : null,
		};
	}

	function parseSopTransitions(payload: unknown): SopTransitionRecord[] {
		const values = Array.isArray(payload) ? payload : isRecord(payload) && Array.isArray(payload.transitions) ? payload.transitions : [];
		return values.map(parseSopTransition).filter((item): item is SopTransitionRecord => Boolean(item));
	}

	function parseSopDetail(payload: unknown): SopDetail | null {
		if (!isRecord(payload)) return null;
		const transitions = parseSopTransitions(payload);
		return {
			event_id: typeof payload.event_id === 'string' ? payload.event_id : undefined,
			current_stage: typeof payload.current_stage === 'string' ? payload.current_stage : typeof payload.stage === 'string' ? payload.stage : undefined,
			current_stage_label: typeof payload.current_stage_label === 'string' ? payload.current_stage_label : undefined,
			next_action: typeof payload.next_action === 'string' || isRecord(payload.next_action) ? payload.next_action as SopDetail['next_action'] : null,
			blockers: parseStringList(payload.blockers),
			warnings: parseStringList(payload.warnings),
			transitions,
			report_only: payload.report_only === true,
		};
	}

	function syncQuestionRequirementForm(completeness: QuestionCompleteness | null) {
		const requirements = completeness?.requirements;
		if (!requirements) return;
		if (requirements.scope_mode === 'per_level' || requirements.scope_mode === 'pool_level_subject') targetScopeMode = requirements.scope_mode;
		else targetScopeMode = 'per_rombel';
		targetStatusFilter = requirements.status_filter === 'all_progress' ? 'all_progress' : 'published_only';
		targetPg = Number.isFinite(Number(requirements.target_pg)) ? Number(requirements.target_pg) : 20;
		targetEssay = Number.isFinite(Number(requirements.target_essay)) ? Number(requirements.target_essay) : 5;
	}

	async function optionalApiData<T>(path: string, fallback: T): Promise<T> {
		try {
			return await fetch(path).then((response) => readClientApiData<T>(response));
		} catch {
			return fallback;
		}
	}

	async function fetchEventApprovals(): Promise<{ approvals: AssessmentApprovalRecord[]; available: boolean }> {
		try {
			return { approvals: await listApprovals('event', eventId), available: true };
		} catch {
			return { approvals: [], available: false };
		}
	}

	async function fetchDetail(): Promise<EventCommandDetail> {
		const [nextInfo, nextResults, overviewPayload, sessionPayload, packagePayload, completenessPayload, sopPayload, sopDetailPayload, sopTransitionsPayload, approvalPayload] = await Promise.all([
			fetch(clientApiPath`/api/asesmen/events/${eventId}`).then((response) => readClientApiData<EventInfo>(response, 'Gagal memuat kegiatan ujian')),
			fetch(clientApiPath`/api/asesmen/events/${eventId}/results`).then((response) => readClientApiData<ResultRow[]>(response, 'Gagal memuat rekap nilai kegiatan')),
			optionalApiData<unknown>(clientApiPath`/api/asesmen/events/${eventId}/overview`, null),
			optionalApiData<unknown>(clientApiPath`/api/asesmen/events/${eventId}/sessions`, []),
			optionalApiData<unknown>(clientApiPath`/api/asesmen/events/${eventId}/packages`, []),
			optionalApiData<unknown>(clientApiPath`/api/asesmen/events/${eventId}/question-completeness`, null),
			optionalApiData<unknown>(clientApiPath`/api/asesmen/events/${eventId}/sop-readiness`, null),
			optionalApiData<unknown>(clientApiPath`/api/asesmen/events/${eventId}/sop`, null),
			optionalApiData<unknown>(clientApiPath`/api/asesmen/events/${eventId}/transitions`, null),
			fetchEventApprovals(),
		]);
		const questionCompleteness = isRecord(completenessPayload) ? completenessPayload as QuestionCompleteness : null;
		syncQuestionRequirementForm(questionCompleteness);
		return {
			info: nextInfo,
			results: Array.isArray(nextResults) ? nextResults : [],
			overview: isRecord(overviewPayload) ? overviewPayload as EventOverview : null,
			sessions: parseArrayPayload<EventSession>(sessionPayload, 'sessions'),
			packages: parseArrayPayload<EventPackage>(packagePayload, 'packages'),
			questionCompleteness,
			sopReadiness: parseSopReadiness(sopPayload),
			sopDetail: parseSopDetail(sopDetailPayload),
			sopTransitions: parseSopTransitions(sopTransitionsPayload),
			approvals: approvalPayload.approvals,
			approvalsAvailable: approvalPayload.available,
		};
	}

	function applyDetail(detail: EventCommandDetail) {
		info = detail.info;
		results = detail.results;
	}

	function loadInitial() {
		const requestId = ++detailRequestId;
		info = null;
		results = [];
		detailPromise = fetchDetail().then((detail) => {
			if (requestId !== detailRequestId) {
				if (!info) throw new Error('Permintaan panel kegiatan dibatalkan');
				return { info, results, overview: null, sessions: [], packages: [], questionCompleteness: null, sopReadiness: null, sopDetail: null, sopTransitions: [], approvals: [], approvalsAvailable: false };
			}
			applyDetail(detail);
			return detail;
		}).catch((error: unknown) => {
			if (requestId === detailRequestId || !info) throw error;
			return { info, results, overview: null, sessions: [], packages: [], questionCompleteness: null, sopReadiness: null, sopDetail: null, sopTransitions: [], approvals: [], approvalsAvailable: false };
		});
	}

	function retryDetail(reset?: () => void) {
		reset?.();
		loadInitial();
	}

	function detailErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Gagal memuat alur kesiapan kegiatan';
	}

	function handleDetailRenderError(error: unknown) {
		console.error('Pemeriksaan kesiapan kegiatan asesmen belum dapat ditampilkan', error);
	}

	function fmtScore(score: string | null) {
		if (!score) return '-';
		const n = parseFloat(score);
		return isNaN(n) ? '-' : n.toFixed(1);
	}

	function countFrom(value: number | undefined, fallback: number | null) {
		return typeof value === 'number' ? value : fallback;
	}

	function checklistTone(count: number | null): ChecklistItem['tone'] {
		if (count === null) return 'info';
		return count > 0 ? 'success' : 'warning';
	}

	function checklistClass(tone: ChecklistItem['tone']) {
		if (tone === 'success') return 'border-primary/20 bg-primary/10';
		if (tone === 'warning') return 'border-warning/30 bg-warning/10';
		return 'border-border bg-card';
	}

	function phaseBadgeClass(tone: ChecklistItem['tone']) {
		if (tone === 'success') return 'border-primary/20 bg-primary/10 text-primary';
		if (tone === 'warning') return 'border-warning/30 bg-warning/10 text-warning';
		return 'border-border bg-card text-muted-foreground';
	}

	function statusClass(status: string) {
		if (status === 'active') return 'bg-primary/15 text-primary border-primary/20';
		if (status === 'finished') return 'bg-muted text-muted-foreground border-border';
		return 'bg-warning/15 text-warning border-warning/30';
	}

	function buildChecklist(detail: EventCommandDetail): ChecklistItem[] {
		const overview = detail.overview;
		const sessionFallback = detail.sessions.length > 0 ? detail.sessions.length : null;
		const packageFallback = detail.packages.length > 0 ? detail.packages.length : null;
		const roomIssues = detail.sessions.filter((session) => (session.room_count ?? 0) === 0 || (session.missing_seat_count ?? 0) > 0 || (session.rooms_without_proctor ?? 0) > 0 || (session.unassigned_participant_count ?? 0) > 0).length;
		const proctorIssues = detail.sessions.filter((session) => (session.rooms_without_proctor ?? 0) > 0).length;
		const items: Array<Omit<ChecklistItem, 'tone'>> = [
			{ label: 'Penugasan', helper: 'Guru pembuat soal dan reviewer kegiatan', count: countFrom(overview?.member_count, null), href: `/asesmen/kegiatan/${eventId}/members`, action: 'Atur penugasan' },
		{ label: 'Kebutuhan Soal', helper: 'Target kebutuhan kegiatan; Bank Soal tetap menjadi daftar soal mandiri sebelum dipakai paket', count: countFrom(overview?.published_question_count ?? overview?.question_count, null), href: `/bank-soal/tambah?event_id=${eventId}`, action: 'Cek target kebutuhan' },
			{ label: 'Verifikasi Bank Soal', helper: 'Antrean verifikasi dari Bank Soal sebelum soal diterbitkan dan masuk paket', count: countFrom(overview?.review_count, null), href: '/bank-soal/verifikasi', action: 'Verifikasi Bank Soal' },
			{ label: 'Paket Kegiatan', helper: 'Prioritas persiapan: paket yang tertaut kegiatan agar sesi ujian bisa memakai paket yang tepat', count: countFrom(overview?.package_count, packageFallback), href: `/asesmen/paket?event_id=${eventId}`, action: 'Kelola paket kegiatan' },
			{ label: 'Sesi/Jadwal', helper: 'Sesi, status, dan jadwal operasional', count: countFrom(overview?.session_count, sessionFallback), href: `/asesmen/sesi?event_id=${eventId}`, action: 'Kelola sesi' },
			{ label: 'Ruang/Pengawas/Kursi', helper: roomIssues > 0 ? `${roomIssues} sesi masih perlu dirapikan${proctorIssues > 0 ? `, ${proctorIssues} butuh pengawas` : ''}` : 'Cek ruang, pengawas, kapasitas, dan nomor meja', count: countFrom(overview?.room_count, detail.sessions.length > 0 ? detail.sessions.reduce((sum, session) => sum + (session.room_count ?? 0), 0) : null), href: `/asesmen/sesi?event_id=${eventId}&readiness=not_ready`, action: 'Cek ruang' },
			{ label: 'Token/Kartu', helper: 'Dokumen & Cetak: kartu peserta, QR+PIN, lembar pengawas ruang, dan validasi kesiapan cetak', count: countFrom(overview?.token_count ?? overview?.card_count, null), href: `/asesmen/kegiatan/${eventId}/cetak`, action: 'Buka Dokumen & Cetak' },
			{ label: 'Hasil', helper: 'Rekap nilai gabungan tersedia di ringkasan hasil', count: countFrom(overview?.result_count, detail.results.length), href: `/asesmen/kegiatan/${eventId}#hasil`, action: 'Lihat hasil' },
			{ label: 'Arsip', helper: 'Checklist kartu, daftar hadir, berita acara, hasil, insiden, dan catatan tindakan ringkas', count: null, href: `/asesmen/kegiatan/${eventId}/archive`, action: 'Lihat arsip' },
		];
		return items.map((item) => ({ ...item, tone: item.label === 'Ruang/Pengawas/Kursi' && roomIssues > 0 ? 'warning' : checklistTone(item.count) }));
	}

	function readinessStatusLabel(item: ChecklistItem) {
		if (item.tone === 'success') return item.count === null ? 'Tersedia' : `${item.count} siap`;
		if (item.tone === 'warning') return item.count === null ? 'Perlu dicek' : `${item.count} perlu dilengkapi`;
		return item.count === null ? 'Cek data' : `${item.count} terbaca`;
	}

	function readinessDotClass(tone: ChecklistItem['tone']) {
		if (tone === 'success') return 'bg-primary';
		if (tone === 'warning') return 'bg-warning';
		return 'bg-muted';
	}

	function buildReadinessGroups(checklist: ChecklistItem[]): ReadinessGroup[] {
		const byLabel = new Map(checklist.map((item) => [item.label, item]));
		return [
			{
				id: 'persiapan',
				title: 'Paket Soal',
				description: 'Tim, kebutuhan soal, verifikasi, dan paket yang akan dipakai sesi.',
				items: ['Penugasan', 'Kebutuhan Soal', 'Verifikasi Bank Soal', 'Paket Kegiatan'].map((label) => byLabel.get(label)).filter((item): item is ChecklistItem => Boolean(item)),
			},
			{
				id: 'pelaksanaan',
				title: 'Kegiatan & Sesi',
				description: 'Jadwal, ruang, pengawas, kursi, token ujian, dan kartu ujian.',
				items: ['Sesi/Jadwal', 'Token/Kartu'].map((label) => byLabel.get(label)).filter((item): item is ChecklistItem => Boolean(item)),
			},
			{
				id: 'pelaksanaan',
				title: 'Monitoring',
				description: 'Kesiapan ruang dan pengawasan saat ujian berlangsung.',
				items: ['Ruang/Pengawas/Kursi'].map((label) => byLabel.get(label)).filter((item): item is ChecklistItem => Boolean(item)),
			},
			{
				id: 'hasil',
				title: 'Hasil & Berita Acara',
				description: 'Rekap nilai gabungan, ekspor, dan berita acara saat data sudah masuk.',
				items: ['Hasil'].map((label) => byLabel.get(label)).filter((item): item is ChecklistItem => Boolean(item)),
			},
			{
				id: 'arsip',
				title: 'Arsip',
				description: 'Dokumen final, pengesahan SOP, dan catatan tindakan ringkas kegiatan.',
				items: ['Arsip'].map((label) => byLabel.get(label)).filter((item): item is ChecklistItem => Boolean(item)),
			},
		];
	}

	function buildNextActions(detail: EventCommandDetail, checklist: ChecklistItem[]): NextAction[] {
		const needsAttention = checklist.filter((item) => item.tone === 'warning');
		const preferredLabels = detail.info.status === 'finished'
			? ['Hasil', 'Token/Kartu', 'Sesi/Jadwal']
			: ['Paket Kegiatan', 'Sesi/Jadwal', 'Ruang/Pengawas/Kursi', 'Token/Kartu', 'Hasil'];
		const preferred = preferredLabels.map((label) => checklist.find((item) => item.label === label)).filter((item): item is ChecklistItem => Boolean(item));
		const ordered = [...needsAttention, ...preferred, ...checklist].filter((item, index, source) => source.findIndex((candidate) => candidate.label === item.label) === index);
		return ordered.slice(0, 3).map((item, index) => ({ ...item, priority: index === 0 ? 'Utama' : `Langkah ${index + 1}` }));
	}

	function buildSopTimeline(detail: EventCommandDetail, checklist: ChecklistItem[]): SopStageReadiness[] {
		const definitions = new Map(sopStages.map((stage) => [stage.key, stage]));
		if (detail.sopReadiness?.stages.length) {
			return detail.sopReadiness.stages.map((stage) => ({
				...definitions.get(stage.key),
				...stage,
				label: definitions.get(stage.key)?.label ?? stage.label,
				owner: definitions.get(stage.key)?.owner,
				description: definitions.get(stage.key)?.description,
			}));
		}

		const byLabel = new Map(checklist.map((item) => [item.label, item]));
		const statusFrom = (item: ChecklistItem | undefined, running = false): SopStageStatus => {
			if (running) return 'running';
			if (!item) return 'blocked';
			if (item.tone === 'success') return 'ready';
			if (item.tone === 'warning') return 'warning';
			return 'blocked';
		};
		const actionFrom = (item: ChecklistItem | undefined) => item ? [{ label: item.action, href: item.href }] : [];
		const packageItem = byLabel.get('Paket Kegiatan');
		const roomItem = byLabel.get('Ruang/Pengawas/Kursi');
		const tokenItem = byLabel.get('Token/Kartu');
		const resultItem = byLabel.get('Hasil');
		const archiveItem = byLabel.get('Arsip');
		return sopStages.map((definition) => {
			const status = definition.key === 'draft'
				? 'ready'
				: definition.key === 'question_authoring'
					? statusFrom(byLabel.get('Kebutuhan Soal'))
					: definition.key === 'question_verification'
						? statusFrom(byLabel.get('Verifikasi Bank Soal'))
						: definition.key === 'package_ready'
							? statusFrom(packageItem)
							: definition.key === 'participants_rooms_ready'
								? statusFrom(roomItem)
								: definition.key === 'tokens_cards_ready'
									? statusFrom(tokenItem)
									: definition.key === 'execution'
										? statusFrom(byLabel.get('Sesi/Jadwal'), detail.sessions.some((session) => session.status === 'active'))
										: definition.key === 'grading' || definition.key === 'result_verification'
											? statusFrom(resultItem)
											: statusFrom(archiveItem);
			const actions = definition.key === 'draft'
				? [{ label: 'Cek identitas kegiatan', href: `/asesmen/kegiatan/${eventId}` }]
				: definition.key === 'question_authoring'
					? actionFrom(byLabel.get('Kebutuhan Soal'))
					: definition.key === 'question_verification'
						? actionFrom(byLabel.get('Verifikasi Bank Soal'))
						: definition.key === 'package_ready'
							? actionFrom(packageItem)
							: definition.key === 'participants_rooms_ready'
								? actionFrom(roomItem)
								: definition.key === 'tokens_cards_ready'
									? actionFrom(tokenItem)
									: definition.key === 'execution'
										? actionFrom(byLabel.get('Sesi/Jadwal'))
										: definition.key === 'grading' || definition.key === 'result_verification'
											? actionFrom(resultItem)
											: actionFrom(archiveItem);
			return { ...definition, status, blocking_count: status === 'blocked' ? 1 : 0, warning_count: status === 'warning' ? 1 : 0, next_actions: actions };
		});
	}

	function sopStageClass(status: SopStageStatus) {
		if (status === 'ready') return 'border-primary/20 bg-primary/10 text-primary';
		if (status === 'running') return 'border-accent bg-accent/70 text-accent-foreground';
		if (status === 'warning') return 'border-warning/30 bg-warning/10 text-warning';
		return 'border-muted bg-muted/60 text-muted-foreground';
	}

	function sopStageLabel(key: string | null | undefined) {
		if (!key) return 'Belum tercatat';
		return sopStages.find((stage) => stage.key === key)?.label ?? key.replace(/_/g, ' ');
	}

	function sopTransitionTime(value: string | null | undefined) {
		if (!value) return 'Waktu belum tercatat';
		return new Date(value).toLocaleString('id-ID', {
			timeZone: 'Asia/Makassar',
			year: 'numeric',
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit',
		}) + ' WITA';
	}

	function currentSopStage(detail: EventCommandDetail, timeline: SopStageReadiness[]) {
		const explicit = detail.sopDetail?.current_stage;
		if (explicit) return timeline.find((stage) => stage.key === explicit) ?? null;
		return timeline.find((stage) => stage.status === 'running')
			?? timeline.find((stage) => stage.status === 'blocked' || stage.status === 'warning')
			?? [...timeline].reverse().find((stage) => stage.status === 'ready')
			?? timeline[0]
			?? null;
	}

	function currentSopLabel(detail: EventCommandDetail, timeline: SopStageReadiness[]) {
		return detail.sopDetail?.current_stage_label ?? currentSopStage(detail, timeline)?.label ?? 'Belum tercatat';
	}

	function currentSopNextAction(detail: EventCommandDetail, timeline: SopStageReadiness[]) {
		const explicit = detail.sopDetail?.next_action;
		if (typeof explicit === 'string' && explicit.trim()) return { label: explicit, href: '' };
		if (isRecord(explicit) && typeof explicit.label === 'string') return { label: explicit.label, href: typeof explicit.href === 'string' ? explicit.href : '' };
		const stage = currentSopStage(detail, timeline);
		return stage?.next_actions[0] ?? null;
	}

	function sopPanelIssues(detail: EventCommandDetail, timeline: SopStageReadiness[], checklist: ChecklistItem[]) {
		const blockers = detail.sopDetail?.blockers?.length
			? detail.sopDetail.blockers
			: checklist.filter((item) => item.tone === 'warning').map((item) => `${item.label}: ${item.helper}`);
		const warnings = detail.sopDetail?.warnings?.length
			? detail.sopDetail.warnings
			: timeline.filter((stage) => stage.status === 'warning').map((stage) => `${stage.label}: ${stage.warning_count} perhatian`);
		return { blockers, warnings };
	}

	function sopHistory(detail: EventCommandDetail) {
		return (detail.sopTransitions.length ? detail.sopTransitions : detail.sopDetail?.transitions ?? []).slice(0, 5);
	}

	function approvalRecordFor(approvals: AssessmentApprovalRecord[], approvalType: EventApprovalType) {
		return approvals.find((record) => record.approval_type === approvalType && record.status === 'approved')
			?? approvals.find((record) => record.approval_type === approvalType)
			?? null;
	}

	function approvalStatusLabel(record: AssessmentApprovalRecord | null) {
		if (!record) return 'Belum disahkan';
		if (record.status === 'approved') return 'Disahkan';
		if (record.status === 'revoked') return 'Dicabut';
		return record.status || 'Belum disahkan';
	}

	function approvalStatusClass(record: AssessmentApprovalRecord | null) {
		if (record?.status === 'approved') return 'border-primary/20 bg-primary/10 text-primary';
		if (record?.status === 'revoked') return 'border-destructive/20 bg-destructive/10 text-destructive';
		return 'border-warning/30 bg-warning/10 text-warning';
	}

	function approvalActor(record: AssessmentApprovalRecord | null) {
		if (!record) return 'Belum ada aktor';
		if (record.status === 'revoked') return record.revoked_by_display_name || record.revoked_by_username || record.approved_by_display_name || record.approved_by_username || 'Aktor tidak tercatat';
		return record.approved_by_display_name || record.approved_by_username || 'Aktor tidak tercatat';
	}

	function approvalTime(record: AssessmentApprovalRecord | null) {
		const value = record?.status === 'revoked' ? record.revoked_at ?? record.approved_at : record?.approved_at;
		if (!value) return 'Waktu belum tercatat';
		return new Date(value).toLocaleString('id-ID', {
			timeZone: 'Asia/Makassar',
			year: 'numeric',
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit',
		}) + ' WITA';
	}

	function approvalStageStatus(timeline: SopStageReadiness[], milestone: SopApprovalMilestone) {
		return timeline.find((stage) => stage.key === milestone.stageKey)?.status ?? 'blocked';
	}

	function approvalStageReady(status: SopStageStatus) {
		return status === 'ready';
	}

	async function approveMilestone(approvalType: EventApprovalType) {
		approvalBusyType = approvalType;
		try {
			await createApproval({
				entity_type: 'event',
				entity_id: eventId,
				approval_type: approvalType,
				notes: approvalNotes[approvalType].trim() || `${assessmentApprovalLabels[approvalType]} untuk kegiatan asesmen.`,
			});
			approvalNotes[approvalType] = '';
			toast.success('Pengesahan SOP disimpan');
			loadInitial();
		} catch (error) {
			toast.error(error instanceof Error ? error.message : 'Gagal menyimpan pengesahan SOP');
		} finally {
			approvalBusyType = null;
		}
	}

	async function revokeMilestone(record: AssessmentApprovalRecord, approvalType: EventApprovalType) {
		approvalBusyType = approvalType;
		try {
			await revokeApproval(record.id, approvalNotes[approvalType].trim() || `Cabut ${assessmentApprovalLabels[approvalType]}.`);
			approvalNotes[approvalType] = '';
			toast.success('Pengesahan SOP dicabut');
			loadInitial();
		} catch (error) {
			toast.error(error instanceof Error ? error.message : 'Gagal mencabut pengesahan SOP');
		} finally {
			approvalBusyType = null;
		}
	}

	function exportCSV() {
		if (!info || results.length === 0) return;
		const header = csvRow(['NIS', 'Nama', 'Kelas', 'Sesi', 'Skor', 'Waktu Kirim Jawaban']);
		const rows = results.map(r => csvRow([r.nis, r.student_nama, r.class_code, r.session_title, fmtScore(r.score), r.submitted_at ?? '']));
		const csv = [header, ...rows].join('\n');
		const blob = new Blob([csv], { type: 'text/csv' });
		const url = URL.createObjectURL(blob);
		const a = document.createElement('a');
		a.href = url;
		a.download = `rekap_${info.title.replace(/\s+/g, '_')}.csv`;
		a.click();
		URL.revokeObjectURL(url);
	}

	function filteredCompletenessRows(detail: EventCommandDetail) {
		const rows = detail.questionCompleteness?.rows ?? [];
		const search = completenessSearch.trim().toLowerCase();
		return rows.filter((row) => {
			if (completenessLevel && row.level !== completenessLevel) return false;
			if (completenessStatus === 'complete' && !row.complete) return false;
			if (completenessStatus === 'incomplete' && row.complete) return false;
			if (!search) return true;
			return [row.class_name, row.class_code, row.subject_name, row.subject_code, row.teacher_name, row.teacher_username]
				.some((value) => value.toLowerCase().includes(search));
		});
	}

	function completenessLevels(detail: EventCommandDetail) {
		return Array.from(new Set((detail.questionCompleteness?.rows ?? []).map((row) => row.level).filter(Boolean))).sort();
	}

	function bankSoalComposerHref(row: QuestionCompletenessRow, questionType: 'multiple_choice' | 'essay' = 'multiple_choice'): `/bank-soal/tambah?${string}` {
		const params = new URLSearchParams();
		params.set('event_id', eventId);
		params.set('subject_id', row.subject_id);
		if (row.level) params.set('target_level', row.level);
		if (row.teacher_username) params.set('teacher_username', row.teacher_username);
		params.set('question_type', questionType);
		return `/bank-soal/tambah?${params.toString()}`;
	}

	async function saveQuestionRequirements() {
		targetSettingsBusy = true;
		try {
			const target_pg = Math.max(0, Number(targetPg) || 0);
			const target_essay = Math.max(0, Number(targetEssay) || 0);
			await fetch(clientApiPath`/api/asesmen/events/${eventId}/question-requirements`, {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ scope_mode: targetScopeMode, target_pg, target_essay, status_filter: targetStatusFilter }),
			}).then((response) => readClientApiData<QuestionRequirement>(response, 'Gagal menyimpan pengaturan target soal'));
			toast.success('Pengaturan target kelengkapan soal disimpan');
			loadInitial();
		} catch (error) {
			toast.error(error instanceof Error ? error.message : 'Gagal menyimpan pengaturan target soal');
		} finally {
			targetSettingsBusy = false;
		}
	}

	function incompleteCompletenessRows(detail: EventCommandDetail) {
		return filteredCompletenessRows(detail).filter((row) => !row.complete);
	}

	function exportCompletenessCSV(detail: EventCommandDetail) {
		if (!info) return;
		const rows = filteredCompletenessRows(detail);
		const header = csvRow(['Tingkat', 'Rombel', 'Mapel', 'Guru', 'Username', 'Target PG', 'PG Ada', 'PG Kurang', 'Target Esai', 'Esai Ada', 'Esai Kurang', 'Status']);
		const body = rows.map((row) => csvRow([
			row.level,
			row.class_name || row.class_code,
			row.subject_name,
			row.teacher_name,
			row.teacher_username,
			row.target_pg,
			row.available_pg,
			row.missing_pg,
			row.target_essay,
			row.available_essay,
			row.missing_essay,
			row.complete ? 'Lengkap' : 'Belum lengkap',
		]));
		const csv = [header, ...body].join('\n');
		const blob = new Blob([csv], { type: 'text/csv' });
		const url = URL.createObjectURL(blob);
		const a = document.createElement('a');
		a.href = url;
		a.download = `kelengkapan_soal_${info.title.replace(/\s+/g, '_')}.csv`;
		a.click();
		URL.revokeObjectURL(url);
	}

	function exportIncompleteByTeacherCSV(detail: EventCommandDetail) {
		if (!info) return;
		const rows = incompleteCompletenessRows(detail).sort((a, b) => `${a.teacher_name}-${a.subject_name}-${a.level}`.localeCompare(`${b.teacher_name}-${b.subject_name}-${b.level}`));
		const header = csvRow(['Guru', 'Username', 'Mapel', 'Tingkat', 'Rombel/Scope', 'PG Ada', 'PG Kurang', 'Esai Ada', 'Esai Kurang', 'Catatan']);
		const body = rows.map((row) => csvRow([
			row.teacher_name,
			row.teacher_username,
			row.subject_name,
			row.level,
			row.class_name || row.class_code || questionRequirementScopeLabel[row.scope_mode ?? ''] || '-',
			row.available_pg,
			row.missing_pg,
			row.available_essay,
			row.missing_essay,
			`Kurang PG ${row.missing_pg}, esai ${row.missing_essay}`,
		]));
		const csv = [header, ...body].join('\n');
		const blob = new Blob([csv], { type: 'text/csv' });
		const url = URL.createObjectURL(blob);
		const a = document.createElement('a');
		a.href = url;
		a.download = `kekurangan_soal_per_guru_${info.title.replace(/\s+/g, '_')}.csv`;
		a.click();
		URL.revokeObjectURL(url);
	}

	function buildReminderDraft(detail: EventCommandDetail) {
		const rows = incompleteCompletenessRows(detail);
		if (!info || rows.length === 0) return '';
		const eventTitle = info.title;
		const grouped = new Map<string, QuestionCompletenessRow[]>();
		for (const row of rows) {
			const key = `${row.teacher_name || row.teacher_username || 'Guru'}|${row.teacher_username || ''}`;
			grouped.set(key, [...(grouped.get(key) ?? []), row]);
		}
		return Array.from(grouped.entries()).map(([key, teacherRows]) => {
			const [teacherName, username] = key.split('|');
			const lines = teacherRows.map((row, index) => `${index + 1}. ${row.subject_name} ${row.level}${row.class_name || row.class_code ? ` (${row.class_name || row.class_code})` : ''}: PG ${row.available_pg}/${row.target_pg} kurang ${row.missing_pg}, Esai ${row.available_essay}/${row.target_essay} kurang ${row.missing_essay}`);
			return `Assalamu'alaikum Bapak/Ibu ${teacherName}.\nMohon melengkapi soal untuk kegiatan ${eventTitle}${username ? ` (akun: ${username})` : ''}:\n${lines.join('\n')}\nSilakan buka Bank Soal/tautan Kelengkapan Soal pada dashboard. Terima kasih.`;
		}).join('\n\n---\n\n');
	}

	async function copyReminderDraft(detail: EventCommandDetail) {
		const text = buildReminderDraft(detail);
		if (!text) {
			toast.info('Tidak ada kekurangan soal pada filter saat ini');
			return;
		}
		try {
			await navigator.clipboard.writeText(text);
			toast.success('Konsep pengingat per guru disalin');
		} catch {
			toast.error('Gagal menyalin draft reminder');
		}
	}

	onMount(() => {
		void loadInitial();
	});
</script>

<svelte:head><title>Pusat Kegiatan Asesmen — {info?.title ?? 'Kegiatan Asesmen'}</title></svelte:head>


<div class="space-y-5">
	<div class="flex items-center gap-2 text-sm text-muted-foreground">
		<a href={resolve('/asesmen/kegiatan')} class="hover:text-foreground">Kegiatan Asesmen & Sesi Ujian</a>
		<span>/</span>
		<span class="text-foreground font-medium truncate max-w-xs">{info?.title ?? 'Pusat Kendali'}</span>
	</div>

	<AsyncContent promise={detailPromise} onerror={handleDetailRenderError}>
		{#snippet pending()}
			<div class="space-y-4">
				<Skeleton class="h-32 w-full" />
				<div class="grid gap-3 lg:grid-cols-[0.82fr_1.18fr]">
					<Skeleton class="h-56 w-full" />
					<Skeleton class="h-56 w-full" />
				</div>
			</div>
		{/snippet}

		{#snippet failed(error, reset)}
			<RecoveryPanel title="Wizard Kesiapan Belum Tersaji" message={detailErrorMessage(error)} onRetry={() => retryDetail(reset)} />
		{/snippet}

		{#snippet children(value)}
			{@const detail = value as EventCommandDetail}
			{@const currentInfo = detail.info}
			{@const currentResults = detail.results}
			{@const checklist = buildChecklist(detail)}
			{@const readinessGroups = buildReadinessGroups(checklist)}
			{@const nextActions = buildNextActions(detail, checklist)}
			{@const sopTimeline = buildSopTimeline(detail, checklist)}
			{@const currentSop = currentSopStage(detail, sopTimeline)}
			{@const currentSopAction = currentSopNextAction(detail, sopTimeline)}
			{@const sopIssues = sopPanelIssues(detail, sopTimeline, checklist)}
			{@const sopTransitionHistory = sopHistory(detail)}
			{@const sopBackendAvailable = detail.sopDetail?.report_only === false}
			{@const blockingItems = checklist.filter((item) => item.tone === 'warning')}
			{@const readyCount = checklist.filter((item) => item.tone === 'success').length}
			<AssessmentPhaseHeader
				code="7.1.1"
				badge={statusLabel[currentInfo.status] ?? currentInfo.status}
				context={`${currentInfo.academic_year_name} · ${scopeLabel[currentInfo.scope] ?? currentInfo.scope}`}
				title={currentInfo.title}
				description="Kelola kegiatan dari ringkasan, persiapan, pelaksanaan, hasil dan berita acara, sampai arsip final."
				primaryAction={{ label: 'Buka Mode Lengkap', href: '#mode-lengkap' }}
				secondaryActions={[{ label: 'Daftar Kegiatan', href: resolve('/asesmen/kegiatan'), variant: 'outline' }]}
			/>

			<ContextStrip
				items={[
					{ label: 'Jenis', value: currentInfo.exam_type, tone: 'muted' },
					{ label: 'Status', value: statusLabel[currentInfo.status] ?? currentInfo.status, tone: currentInfo.status === 'active' ? 'success' : currentInfo.status === 'finished' ? 'muted' : 'warning' },
					{ label: 'Cakupan', value: scopeLabel[currentInfo.scope] ?? currentInfo.scope }
				]}
			/>

			<section class="grid gap-3 md:grid-cols-3" aria-label="Langkah utama kegiatan asesmen">
				{#each mainSections as section (section.id)}
					<AssessmentTaskCard
						code={section.code}
						title={section.label}
						description={section.id === 'persiapan' ? 'Atur paket, sesi, peserta, dan kelengkapan sebelum hari-H.' : section.id === 'pelaksanaan' ? 'Pantau ruang, pengawas, dan kejadian saat ujian berjalan.' : 'Buka rekap nilai, BA, dan tindak lanjut hasil.'}
						href={`${resolve(section.href)}${section.withEvent ? `?event_id=${eventId}` : ''}`}
						cta={`Buka ${section.label}`}
						tone={section.id === 'persiapan' ? 'primary' : 'default'}
					/>
				{/each}
			</section>

			<section class="grid gap-4 lg:grid-cols-[0.9fr_1.1fr]" aria-label="Ringkasan kegiatan asesmen">
				<Card.Root class="border-primary/20 bg-primary/10 shadow-sm">
					<Card.Header class="pb-2">
						<Card.Title class="text-base">Langkah berikutnya</Card.Title>
						<Card.Description>Ringkasan cepat, bukan daftar panjang. Detail kerja tetap di halaman khusus.</Card.Description>
					</Card.Header>
					<Card.Content class="space-y-3">
						{#each nextActions as action (action.label)}
							<a href={resolve(action.href)} class="block rounded-xl border border-white bg-card p-4 shadow-sm transition hover:border-primary/20 hover:shadow-md">
								<div class="flex items-start justify-between gap-3">
									<div class="min-w-0">
										<p class="text-xs font-bold uppercase tracking-[0.16em] text-primary">{action.priority}</p>
										<p class="mt-1 text-sm font-semibold text-foreground">{action.action}</p>
										<p class="mt-1 text-xs leading-5 text-muted-foreground">{action.helper}</p>
									</div>
									<Badge variant={action.tone === 'warning' ? 'secondary' : 'outline'} class={phaseBadgeClass(action.tone)}>{readinessStatusLabel(action)}</Badge>
								</div>
							</a>
						{/each}
					</Card.Content>
				</Card.Root>

				<Card.Root class="border-border shadow-sm">
					<Card.Header class="pb-2">
						<Card.Title class="text-base">Kelengkapan data</Card.Title>
						<Card.Description>Angka praktis agar admin tahu harus masuk ke jalur mana dulu.</Card.Description>
					</Card.Header>
					<Card.Content class="grid gap-2 sm:grid-cols-2 lg:grid-cols-2">
						<div class="rounded-xl bg-muted/50 p-3"><p class="text-xs text-muted-foreground">Paket</p><p class="text-lg font-semibold text-foreground">{detail.packages.length}</p></div>
						<div class="rounded-xl bg-muted/50 p-3"><p class="text-xs text-muted-foreground">Sesi</p><p class="text-lg font-semibold text-foreground">{detail.sessions.length}</p></div>
						<div class="rounded-xl bg-muted/50 p-3"><p class="text-xs text-muted-foreground">Baris hasil</p><p class="text-lg font-semibold text-foreground">{currentResults.length}</p></div>
						<div class="rounded-xl bg-muted/50 p-3"><p class="text-xs text-muted-foreground">Kesiapan</p><p class="text-sm font-semibold text-foreground">{detail.overview ? 'Terbaca dari sistem' : 'Sebagian data tersedia'}</p></div>
					</Card.Content>
				</Card.Root>
			</section>

			<section id="mode-lengkap" class="rounded-2xl border border-border bg-card p-4 shadow-sm" aria-label="Mode lengkap kegiatan asesmen">
				<div class="flex flex-col gap-2 md:flex-row md:items-start md:justify-between">
					<div>
						<p class="text-sm font-semibold text-foreground">7.1.1 Mode Lengkap</p>
						<p class="text-sm leading-6 text-muted-foreground">Detail teknis, SOP, arsip, dan kelengkapan lengkap dipisah agar layar utama tetap ringan.</p>
					</div>
					<a href={resolve(`/asesmen/kegiatan/${eventId}/cetak`)} class="inline-flex rounded-md border border-primary/20 bg-primary/10 px-3 py-2 text-sm font-semibold text-primary hover:bg-primary/15">Dokumen & Cetak</a>
				</div>
				<div class="mt-4 grid gap-3 md:grid-cols-2 xl:grid-cols-4">
					<a href={resolve(`/asesmen/persiapan?event_id=${eventId}`)} class="rounded-xl border border-border bg-muted/20 p-4 transition hover:border-primary/20 hover:bg-primary/5">
						<p class="text-sm font-semibold text-foreground">Persiapan</p>
						<p class="mt-1 text-xs leading-5 text-muted-foreground">Kelola paket, sesi, peserta, dan cek kesiapan.</p>
					</a>
					<a href={resolve(`/asesmen/sesi?event_id=${eventId}`)} class="rounded-xl border border-border bg-muted/20 p-4 transition hover:border-primary/20 hover:bg-primary/5">
						<p class="text-sm font-semibold text-foreground">Pelaksanaan</p>
						<p class="mt-1 text-xs leading-5 text-muted-foreground">Lihat sesi, ruang, pengawas, dan status berjalan.</p>
					</a>
					<a href={resolve('/asesmen/hasil')} class="rounded-xl border border-border bg-muted/20 p-4 transition hover:border-primary/20 hover:bg-primary/5">
						<p class="text-sm font-semibold text-foreground">Hasil</p>
						<p class="mt-1 text-xs leading-5 text-muted-foreground">Buka rekap, BA, dan tindak lanjut nilai.</p>
					</a>
					<a href={resolve(`/asesmen/kegiatan/${eventId}/archive`)} class="rounded-xl border border-border bg-muted/20 p-4 transition hover:border-primary/20 hover:bg-primary/5">
						<p class="text-sm font-semibold text-foreground">Arsip</p>
						<p class="mt-1 text-xs leading-5 text-muted-foreground">Lihat pengesahan, BA, dan dokumen final.</p>
					</a>
				</div>
			</section>
{/snippet}
	</AsyncContent>
</div>
