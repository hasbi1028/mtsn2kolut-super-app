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
		| `/asesmen/kegiatan/${string}/exam-cards`
		| `/asesmen/kegiatan/${string}#hasil`
		| `/asesmen/kegiatan/${string}/archive`;
	type ChecklistItem = {
		label: string; helper: string; count: number | null; href: ChecklistHref; tone: 'success' | 'warning' | 'info'; action: string;
	};
	type EventSection = 'ringkasan' | 'persiapan' | 'kelengkapan-soal' | 'operasional' | 'hasil';
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

	const eventId = page.params.id ?? '';
	let info = $state<EventInfo | null>(null);
	let results = $state<ResultRow[]>([]);
	let detailPromise = $state<Promise<EventCommandDetail> | null>(null);
	let activeSection = $state<EventSection>('ringkasan');
	let hasilSectionElement = $state<HTMLElement | null>(null);
	let hasilFocusRequest = $state(0);
	let detailRequestId = 0;
	let handledHasilFocusRequest = 0;
	const sectionTabs: Array<{ id: EventSection; label: string }> = [
		{ id: 'ringkasan', label: 'Ringkasan' },
		{ id: 'persiapan', label: 'Persiapan' },
		{ id: 'kelengkapan-soal', label: 'Kelengkapan Soal' },
		{ id: 'operasional', label: 'Operasional' },
		{ id: 'hasil', label: 'Hasil' },
	];
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
		const [nextInfo, nextResults, overviewPayload, sessionPayload, packagePayload, completenessPayload, sopPayload, approvalPayload] = await Promise.all([
			fetch(clientApiPath`/api/asesmen/events/${eventId}`).then((response) => readClientApiData<EventInfo>(response, 'Gagal memuat kegiatan ujian')),
			fetch(clientApiPath`/api/asesmen/events/${eventId}/results`).then((response) => readClientApiData<ResultRow[]>(response, 'Gagal memuat rekap nilai kegiatan')),
			optionalApiData<unknown>(clientApiPath`/api/asesmen/events/${eventId}/overview`, null),
			optionalApiData<unknown>(clientApiPath`/api/asesmen/events/${eventId}/sessions`, []),
			optionalApiData<unknown>(clientApiPath`/api/asesmen/events/${eventId}/packages`, []),
			optionalApiData<unknown>(clientApiPath`/api/asesmen/events/${eventId}/question-completeness`, null),
			optionalApiData<unknown>(clientApiPath`/api/asesmen/events/${eventId}/sop-readiness`, null),
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
				return { info, results, overview: null, sessions: [], packages: [], questionCompleteness: null, sopReadiness: null, approvals: [], approvalsAvailable: false };
			}
			applyDetail(detail);
			return detail;
		}).catch((error: unknown) => {
			if (requestId === detailRequestId || !info) throw error;
			return { info, results, overview: null, sessions: [], packages: [], questionCompleteness: null, sopReadiness: null, approvals: [], approvalsAvailable: false };
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
		{ label: 'Kebutuhan Soal', helper: 'Target kebutuhan kegiatan; Bank Soal tetap repositori mandiri sebelum dipakai paket', count: countFrom(overview?.published_question_count ?? overview?.question_count, null), href: `/bank-soal/tambah?event_id=${eventId}`, action: 'Cek target kebutuhan' },
			{ label: 'Verifikasi Repositori', helper: 'Antrean verifikasi dari Bank Soal sebelum soal diterbitkan dan masuk paket', count: countFrom(overview?.review_count, null), href: '/bank-soal/verifikasi', action: 'Verifikasi repositori' },
			{ label: 'Paket Kegiatan', helper: 'Prioritas persiapan: paket yang tertaut kegiatan agar sesi ujian bisa memakai paket yang tepat', count: countFrom(overview?.package_count, packageFallback), href: `/asesmen/paket?event_id=${eventId}`, action: 'Kelola paket kegiatan' },
			{ label: 'Sesi/Jadwal', helper: 'Sesi, status, dan jadwal operasional', count: countFrom(overview?.session_count, sessionFallback), href: `/asesmen/sesi?event_id=${eventId}`, action: 'Kelola sesi' },
			{ label: 'Ruang/Pengawas/Kursi', helper: roomIssues > 0 ? `${roomIssues} sesi masih perlu dirapikan${proctorIssues > 0 ? `, ${proctorIssues} butuh pengawas` : ''}` : 'Cek ruang, pengawas, kapasitas, dan nomor meja', count: countFrom(overview?.room_count, detail.sessions.length > 0 ? detail.sessions.reduce((sum, session) => sum + (session.room_count ?? 0), 0) : null), href: `/asesmen/sesi?event_id=${eventId}&readiness=not_ready`, action: 'Cek ruang' },
			{ label: 'Token/Kartu', helper: 'Token ujian peserta dan kartu ujian siap cetak', count: countFrom(overview?.token_count ?? overview?.card_count, null), href: `/asesmen/kegiatan/${eventId}/exam-cards`, action: 'Cetak kartu' },
			{ label: 'Hasil', helper: 'Rekap nilai gabungan tersedia di tab Hasil', count: countFrom(overview?.result_count, detail.results.length), href: `/asesmen/kegiatan/${eventId}#hasil`, action: 'Buka tab hasil' },
			{ label: 'Arsip', helper: 'Checklist kartu, daftar hadir, berita acara, hasil, insiden, dan audit ringkas', count: null, href: `/asesmen/kegiatan/${eventId}/archive`, action: 'Buka checklist arsip' },
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
				items: ['Penugasan', 'Kebutuhan Soal', 'Verifikasi Repositori', 'Paket Kegiatan'].map((label) => byLabel.get(label)).filter((item): item is ChecklistItem => Boolean(item)),
			},
			{
				id: 'operasional',
				title: 'Kegiatan & Sesi',
				description: 'Jadwal, ruang, pengawas, kursi, token ujian, dan kartu ujian.',
				items: ['Sesi/Jadwal', 'Token/Kartu'].map((label) => byLabel.get(label)).filter((item): item is ChecklistItem => Boolean(item)),
			},
			{
				id: 'operasional',
				title: 'Monitoring',
				description: 'Kesiapan ruang dan pengawasan saat ujian berlangsung.',
				items: ['Ruang/Pengawas/Kursi'].map((label) => byLabel.get(label)).filter((item): item is ChecklistItem => Boolean(item)),
			},
			{
				id: 'hasil',
				title: 'Hasil',
				description: 'Rekap nilai gabungan, ekspor, dan arsip saat data sudah masuk.',
				items: ['Hasil', 'Arsip'].map((label) => byLabel.get(label)).filter((item): item is ChecklistItem => Boolean(item)),
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
						? statusFrom(byLabel.get('Verifikasi Repositori'))
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
						? actionFrom(byLabel.get('Verifikasi Repositori'))
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
			toast.success('Draft reminder per guru disalin');
		} catch {
			toast.error('Gagal menyalin draft reminder');
		}
	}

	function activateHasilHash() {
		if (window.location.hash !== '#hasil') return;
		activeSection = 'hasil';
		hasilFocusRequest += 1;
	}

	function handleHashChange() {
		activateHasilHash();
	}

	$effect(() => {
		if (hasilFocusRequest === handledHasilFocusRequest || activeSection !== 'hasil' || !hasilSectionElement) return;
		handledHasilFocusRequest = hasilFocusRequest;
		hasilSectionElement.focus({ preventScroll: true });
		hasilSectionElement.scrollIntoView({ block: 'start', behavior: 'smooth' });
	});

	onMount(() => {
		activateHasilHash();
		void loadInitial();
	});
</script>

<svelte:head><title>Pusat Kegiatan Asesmen — {info?.title ?? 'Kegiatan Asesmen'}</title></svelte:head>

<svelte:window onhashchange={handleHashChange} />

<div class="space-y-6 p-6">
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
			<section class="overflow-hidden rounded-2xl border border-border bg-card shadow-sm">
				<div class="flex flex-wrap items-start justify-between gap-4">
					<div class="max-w-3xl p-5">
						<p class="text-xs font-bold uppercase tracking-[0.18em] text-primary">Pusat Kegiatan Asesmen</p>
						<h1 class="mt-1 text-2xl font-semibold text-foreground">{currentInfo.title}</h1>
						<p class="mt-2 text-sm text-muted-foreground">{currentInfo.academic_year_name} · <span class="capitalize">{currentInfo.exam_type}</span> · {scopeLabel[currentInfo.scope] ?? currentInfo.scope}</p>
						<p class="mt-2 text-sm text-muted-foreground">Ikuti langkah kesiapan dari Paket Soal, Sesi Ujian, Pengawasan Ruang, sampai Hasil Asesmen tanpa membuka banyak halaman setara.</p>
					</div>
					<div class="flex flex-wrap items-center gap-2 p-5 lg:justify-end">
						<Badge class={statusClass(currentInfo.status)}>{statusLabel[currentInfo.status] ?? currentInfo.status}</Badge>
						{#each currentInfo.target_levels ?? [] as level (level)}
							<Badge variant="outline" class="bg-card">Tingkat {level}</Badge>
						{/each}
						{#if !currentInfo.target_levels?.length}
							<Badge variant="outline" class="bg-card">Target mengikuti cakupan</Badge>
						{/if}
					</div>
				</div>
				<nav class="flex gap-1 overflow-x-auto border-t border-border bg-muted/50 px-3 py-2" aria-label="Bagian pusat kegiatan">
					{#each sectionTabs as tab (tab.id)}
						<button
							type="button"
							class={`rounded-full px-3 py-1.5 text-sm font-semibold transition ${activeSection === tab.id ? 'bg-card text-primary shadow-sm ring-1 ring-primary/30' : 'text-muted-foreground hover:bg-card hover:text-foreground'}`}
							aria-current={activeSection === tab.id ? 'page' : undefined}
							aria-pressed={activeSection === tab.id}
							onclick={() => activeSection = tab.id}
						>
							{tab.label}
						</button>
					{/each}
				</nav>
			</section>

			<section class="grid gap-4 lg:grid-cols-[0.82fr_1.18fr]" aria-label="Wizard kesiapan kegiatan">
				<Card.Root class="border-primary/20 bg-primary/10 shadow-sm">
					<Card.Header class="pb-2">
						<Card.Title class="text-base">Langkah berikutnya</Card.Title>
						<Card.Description>Rekomendasi ringkas dari data kesiapan yang tersedia saat ini.</Card.Description>
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
						<Card.Title class="text-base">Kesiapan kegiatan</Card.Title>
						<Card.Description>Satu permukaan utama untuk membaca progres. Buka tab di bawah untuk rincian kerja atau hasil.</Card.Description>
					</Card.Header>
					<Card.Content class="space-y-4">
						{#each readinessGroups as group (group.title)}
							<div class="rounded-xl border border-border bg-card p-4">
								<div class="flex flex-wrap items-start justify-between gap-3">
									<div>
										<p class="text-sm font-semibold text-foreground">{group.title}</p>
										<p class="mt-1 text-xs text-muted-foreground">{group.description}</p>
									</div>
									<button type="button" class="text-xs font-semibold text-primary hover:text-primary" onclick={() => activeSection = group.id}>Buka tab</button>
								</div>
								<div class="mt-3 divide-y divide-border">
									{#each group.items as item (item.label)}
										<a href={resolve(item.href)} class="flex items-center justify-between gap-3 py-2 text-sm hover:text-primary">
											<span class="flex min-w-0 items-center gap-2">
												<span class={`size-2 rounded-full ${readinessDotClass(item.tone)}`}></span>
												<span class="truncate font-medium text-foreground">{item.label}</span>
											</span>
											<span class="shrink-0 text-xs font-semibold text-muted-foreground">{readinessStatusLabel(item)}</span>
										</a>
									{/each}
								</div>
							</div>
						{/each}
					</Card.Content>
				</Card.Root>
			</section>

			<section aria-label="Timeline SOP kegiatan asesmen">
				<Card.Root class="border-primary/20 shadow-sm">
					<Card.Header class="pb-3">
						<div class="flex flex-col gap-2 md:flex-row md:items-start md:justify-between">
							<div>
								<Card.Title class="text-base">Timeline SOP Kegiatan</Card.Title>
								<Card.Description>Sepuluh tahap formal dari persiapan soal sampai arsip akhir. Status bersifat panduan baca, belum mengunci alur kerja.</Card.Description>
							</div>
							<Badge variant="outline" class={detail.sopReadiness ? 'border-primary/20 bg-primary/10 text-primary' : 'border-warning/30 bg-warning/10 text-warning'}>
								{detail.sopReadiness ? 'Readiness backend' : 'Fallback halaman'}
							</Badge>
						</div>
					</Card.Header>
					<Card.Content>
						<div class="grid gap-3 lg:grid-cols-2">
							{#each sopTimeline as stage, index (stage.key)}
								<div class="flex gap-3 rounded-xl border border-border bg-card p-4">
									<div class="flex flex-col items-center">
										<div class={`flex size-8 items-center justify-center rounded-full border text-xs font-bold ${sopStageClass(stage.status)}`}>{index + 1}</div>
										{#if index < sopTimeline.length - 1}
											<div class="mt-2 h-full min-h-10 w-px bg-border"></div>
										{/if}
									</div>
									<div class="min-w-0 flex-1 space-y-2">
										<div class="flex flex-wrap items-start justify-between gap-2">
											<div>
												<p class="text-sm font-semibold text-foreground">{stage.label}</p>
												<p class="text-xs text-muted-foreground">{stage.owner ?? 'Panitia'} · {stage.description ?? 'Tahap SOP kegiatan asesmen.'}</p>
											</div>
											<Badge variant="outline" class={sopStageClass(stage.status)}>{sopStatusLabels[stage.status]}</Badge>
										</div>
										<div class="flex flex-wrap gap-2 text-[11px] text-muted-foreground">
											<span>{stage.blocking_count} penghambat</span>
											<span>{stage.warning_count} perhatian</span>
										</div>
										<div class="flex flex-wrap gap-2">
											{#each stage.next_actions.slice(0, 3) as action (action.href + action.label)}
												<a href={action.href} class="rounded-md border border-border px-2.5 py-1 text-xs font-semibold text-primary hover:border-primary/30 hover:bg-primary/10">{action.label}</a>
											{:else}
												<span class="text-xs text-muted-foreground">Tidak ada tindakan lanjutan dari data aktif.</span>
											{/each}
										</div>
									</div>
								</div>
							{/each}
						</div>
					</Card.Content>
				</Card.Root>
			</section>

			<section aria-label="Pengesahan SOP kegiatan asesmen">
				<Card.Root class="border-primary/20 shadow-sm">
					<Card.Header class="pb-3">
						<div class="flex flex-col gap-2 md:flex-row md:items-start md:justify-between">
							<div>
								<Card.Title class="text-base">Pengesahan SOP</Card.Title>
								<Card.Description>Pengesahan ini mencatat audit formal, belum memblokir alur lama.</Card.Description>
							</div>
							<a href={resolve(`/asesmen/kegiatan/${eventId}/archive`)} class="inline-flex rounded-md border border-primary/20 bg-primary/10 px-3 py-2 text-sm font-semibold text-primary hover:bg-primary/15">Buka Arsip</a>
						</div>
					</Card.Header>
					<Card.Content class="space-y-4">
						{#if !detail.approvalsAvailable}
							<p class="rounded-xl border border-warning/30 bg-warning/10 p-3 text-sm text-warning">Audit pengesahan SOP belum dapat dimuat untuk sesi ini. Pengelolaan pengesahan formal tersedia melalui akses admin.</p>
						{/if}
						<div class="grid gap-3 lg:grid-cols-2">
							{#each sopApprovalMilestones as milestone (milestone.approvalType)}
								{@const record = approvalRecordFor(detail.approvals, milestone.approvalType)}
								{@const stageStatus = approvalStageStatus(sopTimeline, milestone)}
								<div class="rounded-xl border border-border bg-card p-4">
									<div class="flex flex-wrap items-start justify-between gap-3">
										<div class="min-w-0">
											<p class="text-sm font-semibold text-foreground">{assessmentApprovalLabels[milestone.approvalType]}</p>
											<p class="mt-1 text-xs leading-5 text-muted-foreground">{milestone.helper}</p>
										</div>
										<div class="flex flex-wrap gap-2">
											<Badge variant="outline" class={sopStageClass(stageStatus)}>SOP {sopStatusLabels[stageStatus]}</Badge>
											<Badge variant="outline" class={approvalStatusClass(record)}>{approvalStatusLabel(record)}</Badge>
										</div>
									</div>
									<div class="mt-3 grid gap-2 rounded-lg bg-muted/40 p-3 text-xs text-muted-foreground sm:grid-cols-2">
										<p><span class="font-semibold text-foreground">Aktor:</span> {approvalActor(record)}</p>
										<p><span class="font-semibold text-foreground">Waktu:</span> {approvalTime(record)}</p>
									</div>
									{#if record?.notes}
										<p class="mt-3 rounded-lg border border-border bg-muted/30 p-3 text-xs leading-5 text-muted-foreground"><span class="font-semibold text-foreground">Catatan:</span> {record.notes}</p>
									{/if}
									{#if detail.approvalsAvailable}
										<label class="mt-3 block space-y-1 text-xs font-medium text-muted-foreground" for={`approval-note-${milestone.approvalType}`}>
											Catatan pengesahan/cabut
											<textarea
												id={`approval-note-${milestone.approvalType}`}
												bind:value={approvalNotes[milestone.approvalType]}
												rows="2"
												class="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground"
												placeholder="Opsional, misalnya hasil pemeriksaan panitia"
											></textarea>
										</label>
										<div class="mt-3 flex flex-wrap gap-2">
											{#if record?.status === 'approved'}
												<LoadingButton
													variant="destructive"
													size="sm"
													onclick={() => revokeMilestone(record, milestone.approvalType)}
													loading={approvalBusyType === milestone.approvalType}
													loadingLabel="Mencabut..."
													disabled={approvalBusyType !== null && approvalBusyType !== milestone.approvalType}
													label="Cabut pengesahan"
												/>
											{:else}
												<LoadingButton
													size="sm"
													onclick={() => approveMilestone(milestone.approvalType)}
													loading={approvalBusyType === milestone.approvalType}
													loadingLabel="Mengesahkan..."
													disabled={approvalBusyType !== null && approvalBusyType !== milestone.approvalType}
													label="Sahkan"
												/>
											{/if}
										</div>
									{/if}
								</div>
							{/each}
						</div>
					</Card.Content>
				</Card.Root>
			</section>

			{#if activeSection === 'ringkasan'}
				<section>
					<Card.Root>
				<Card.Header class="pb-2"><Card.Title class="text-base">Kelengkapan data</Card.Title><Card.Description>Angka praktis dari paket, sesi, dan hasil yang sudah terbaca.</Card.Description></Card.Header>
						<Card.Content class="grid gap-2 sm:grid-cols-2 lg:grid-cols-4">
							<div class="rounded-xl bg-muted/50 p-3"><p class="text-xs text-muted-foreground">Paket</p><p class="text-lg font-semibold text-foreground">{detail.packages.length}</p></div>
							<div class="rounded-xl bg-muted/50 p-3"><p class="text-xs text-muted-foreground">Sesi</p><p class="text-lg font-semibold text-foreground">{detail.sessions.length}</p></div>
							<div class="rounded-xl bg-muted/50 p-3"><p class="text-xs text-muted-foreground">Baris hasil</p><p class="text-lg font-semibold text-foreground">{currentResults.length}</p></div>
							<div class="rounded-xl bg-muted/50 p-3"><p class="text-xs text-muted-foreground">Ringkasan kesiapan</p><p class="text-sm font-semibold text-foreground">{detail.overview ? 'Lengkap dari sistem' : 'Sebagian data tersedia'}</p></div>
						</Card.Content>
					</Card.Root>
				</section>
			{/if}

			{#if activeSection === 'kelengkapan-soal'}
				{@const completeness = detail.questionCompleteness}
				{@const filteredRows = filteredCompletenessRows(detail)}
				<section class="space-y-4">
					<Card.Root>
						<Card.Header class="pb-2">
							<div class="flex flex-wrap items-start justify-between gap-3">
								<div>
									<Card.Title class="text-base">Kelengkapan Soal per Guru/Mapel/Rombel</Card.Title>
									<Card.Description>Target aktif: {questionRequirementScopeLabel[completeness?.requirements?.scope_mode ?? 'per_rombel'] ?? 'Per rombel + mapel + guru'} · PG {completeness?.requirements?.target_pg ?? 20} · Esai {completeness?.requirements?.target_essay ?? 5} · {questionRequirementStatusLabel[completeness?.requirements?.status_filter ?? 'published_only'] ?? 'Hanya soal terbit'}.</Card.Description>
								</div>
								<div class="flex flex-wrap gap-2">
						<LoadingButton variant="outline" onclick={() => exportIncompleteByTeacherCSV(detail)} disabled={incompleteCompletenessRows(detail).length === 0} label="Ekspor Kurang per Guru" />
						<LoadingButton variant="outline" onclick={() => copyReminderDraft(detail)} disabled={incompleteCompletenessRows(detail).length === 0} label="Salin Reminder" />
						<LoadingButton variant="outline" onclick={() => exportCompletenessCSV(detail)} disabled={filteredRows.length === 0} label="Ekspor CSV" />
					</div>
							</div>
						</Card.Header>
						<Card.Content class="space-y-4">
							<div class="rounded-2xl border border-border bg-muted/30 p-4">
								<div class="flex flex-wrap items-start justify-between gap-3">
									<div>
										<p class="text-sm font-semibold text-foreground">Pengaturan Target</p>
										<p class="text-xs text-muted-foreground">Atur cara pemantauan dan target minimal soal yang dihitung untuk kegiatan ini.</p>
									</div>
									<LoadingButton onclick={saveQuestionRequirements} loading={targetSettingsBusy} loadingLabel="Menyimpan..." label="Simpan Target" />
								</div>
								<div class="mt-3 grid gap-2 md:grid-cols-4">
									<label class="space-y-1 text-xs font-medium text-muted-foreground">Mode monitoring
										<select bind:value={targetScopeMode} class="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground">
											<option value="per_rombel">Per rombel + mapel + guru</option>
											<option value="per_level">Per tingkat + mapel + guru</option>
											<option value="pool_level_subject">Kumpulan tingkat + mapel</option>
										</select>
									</label>
									<label class="space-y-1 text-xs font-medium text-muted-foreground">Filter soal
										<select bind:value={targetStatusFilter} class="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground">
											<option value="published_only">Hanya soal terbit</option>
											<option value="all_progress">Draft/verifikasi/terbit dihitung</option>
										</select>
									</label>
									<label class="space-y-1 text-xs font-medium text-muted-foreground">Target PG
										<input bind:value={targetPg} type="number" min="0" class="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground" />
									</label>
									<label class="space-y-1 text-xs font-medium text-muted-foreground">Target esai
										<input bind:value={targetEssay} type="number" min="0" class="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground" />
									</label>
								</div>
							</div>
							{#if completeness}
								<div class="grid gap-2 sm:grid-cols-2 lg:grid-cols-4">
									<div class="rounded-xl bg-muted/50 p-3"><p class="text-xs text-muted-foreground">Penugasan</p><p class="text-lg font-semibold text-foreground">{completeness.summary.total_rows}</p></div>
									<div class="rounded-xl bg-success/10 p-3"><p class="text-xs text-muted-foreground">Lengkap</p><p class="text-lg font-semibold text-success">{completeness.summary.complete_rows}</p></div>
									<div class="rounded-xl bg-warning/10 p-3"><p class="text-xs text-muted-foreground">Belum lengkap</p><p class="text-lg font-semibold text-warning">{completeness.summary.incomplete_rows}</p></div>
									<div class="rounded-xl bg-muted/50 p-3"><p class="text-xs text-muted-foreground">Kekurangan total</p><p class="text-lg font-semibold text-foreground">PG {completeness.summary.missing_pg} · Esai {completeness.summary.missing_essay}</p></div>
								</div>
				{#if completeness.excluded_levels?.length}
					<p class="rounded-xl border border-warning/30 bg-warning/10 p-3 text-xs text-muted-foreground">Tingkat tidak dihitung karena di luar target event: {completeness.excluded_levels.join(', ')}.</p>
				{/if}
				{#if completeness.contributions?.length && completeness.requirements?.scope_mode === 'pool_level_subject'}
					<details class="rounded-xl border border-border bg-muted/30 p-3 text-xs text-foreground">
						<summary class="cursor-pointer font-semibold">Kontribusi guru ke pool mapel ({completeness.contributions.length})</summary>
						<div class="mt-2 grid gap-2 md:grid-cols-2 lg:grid-cols-3">
							{#each completeness.contributions.slice(0, 12) as row (`${row.level}-${row.subject_id}-${row.teacher_username}`)}
								<div class="rounded-lg border border-border bg-card p-2">
									<p class="font-medium">{row.level} · {row.subject_name}</p>
									<p class="text-muted-foreground">{row.teacher_name || row.teacher_username || 'Guru belum tertaut'}</p>
									<p class="mt-1 font-semibold">PG {row.available_pg} · Esai {row.available_essay}</p>
								</div>
							{/each}
						</div>
						{#if completeness.contributions.length > 12}<p class="mt-2 text-muted-foreground">Menampilkan 12 kontribusi pertama; ekspor CSV untuk data lengkap.</p>{/if}
					</details>
				{/if}
				<div class="grid gap-2 md:grid-cols-3">
									<select bind:value={completenessLevel} class="rounded-lg border border-border bg-background px-3 py-2 text-sm">
										<option value="">Semua tingkat</option>
										{#each completenessLevels(detail) as level (level)}<option value={level}>Tingkat {level}</option>{/each}
									</select>
									<select bind:value={completenessStatus} class="rounded-lg border border-border bg-background px-3 py-2 text-sm">
										<option value="">Semua status</option>
										<option value="incomplete">Belum lengkap</option>
										<option value="complete">Lengkap</option>
									</select>
									<input bind:value={completenessSearch} class="rounded-lg border border-border bg-background px-3 py-2 text-sm" placeholder="Cari rombel, mapel, atau guru" />
								</div>
							{:else}
								<p class="rounded-xl bg-muted/50 p-4 text-sm text-muted-foreground">Data kelengkapan soal belum tersedia dari API.</p>
							{/if}
						</Card.Content>
					</Card.Root>

					<Card.Root>
						<Card.Content class="p-0 overflow-x-auto">
							<Table.Root>
								<Table.Header><Table.Row class="bg-muted/50"><Table.Head>Tingkat</Table.Head><Table.Head>Rombel</Table.Head><Table.Head>Mapel</Table.Head><Table.Head>Guru</Table.Head><Table.Head class="text-center">PG</Table.Head><Table.Head class="text-center">Esai</Table.Head><Table.Head>Status</Table.Head></Table.Row></Table.Header>
								<Table.Body>
									{#each filteredRows as row (`${row.level}-${row.class_id}-${row.subject_id}-${row.teacher_employee_id}`)}
										<Table.Row>
											<Table.Cell><Badge variant="outline" class="bg-card">{row.level}</Badge></Table.Cell>
											<Table.Cell class="font-medium">{row.class_name || row.class_code}</Table.Cell>
											<Table.Cell>{row.subject_name}</Table.Cell>
											<Table.Cell><div class="font-medium">{row.teacher_name}</div><div class="text-xs text-muted-foreground">{row.teacher_username || 'username belum tertaut'}</div></Table.Cell>
											<Table.Cell class="text-center"><span class={row.missing_pg > 0 ? 'font-semibold text-warning' : 'font-semibold text-success'}>{row.available_pg}/{row.target_pg}</span>{#if row.missing_pg > 0}<div class="text-xs text-muted-foreground">kurang {row.missing_pg}</div>{/if}</Table.Cell>
											<Table.Cell class="text-center"><span class={row.missing_essay > 0 ? 'font-semibold text-warning' : 'font-semibold text-success'}>{row.available_essay}/{row.target_essay}</span>{#if row.missing_essay > 0}<div class="text-xs text-muted-foreground">kurang {row.missing_essay}</div>{/if}</Table.Cell>
											<Table.Cell>{#if row.complete}<Badge variant="outline" class="bg-success/10 text-success border-success/20">Lengkap</Badge>{:else}<div class="flex flex-col gap-1"><Badge variant="secondary" class="bg-warning/10 text-warning border-warning/30">Belum</Badge><div class="flex flex-wrap gap-1"><a class="text-xs font-semibold text-primary hover:underline" href={resolve(bankSoalComposerHref(row, 'multiple_choice'))}>Tambah PG</a><a class="text-xs font-semibold text-primary hover:underline" href={resolve(bankSoalComposerHref(row, 'essay'))}>Tambah esai</a></div></div>{/if}</Table.Cell>
										</Table.Row>
									{:else}
										<Table.Row><Table.Cell colspan={7} class="py-12 text-center text-muted-foreground">Tidak ada baris sesuai filter.</Table.Cell></Table.Row>
									{/each}
								</Table.Body>
							</Table.Root>
						</Card.Content>
					</Card.Root>
				</section>
			{/if}

			{#if activeSection === 'persiapan' || activeSection === 'operasional'}
				{@const activeGroups = readinessGroups.filter((item) => item.id === activeSection)}
				{#if activeGroups.length > 0}
					<section class="grid gap-3 md:grid-cols-2">
						{#each activeGroups as group (group.title)}
							{#each group.items as item (item.label)}
							<a href={resolve(item.href)} class={`block rounded-xl border p-4 shadow-sm transition hover:-translate-y-0.5 hover:shadow-md ${checklistClass(item.tone)}`}>
								<div class="flex items-start justify-between gap-3">
									<div>
										<p class="text-xs font-semibold uppercase tracking-[0.14em] text-muted-foreground">{group.title}</p>
										<p class="mt-1 text-sm font-semibold text-foreground">{item.label}</p>
										<p class="mt-1 text-xs text-muted-foreground">{item.helper}</p>
									</div>
									<Badge variant={item.tone === 'warning' ? 'secondary' : 'outline'} class="bg-card">{item.count ?? 'Cek'}</Badge>
								</div>
								<p class="mt-4 text-sm font-semibold text-success">{item.action}</p>
							</a>
							{/each}
						{/each}
					</section>
				{/if}
			{/if}

			{#if activeSection === 'hasil'}
			<Card.Root id="hasil" bind:ref={hasilSectionElement} tabindex={-1}>
				<Card.Header class="pb-2">
					<div class="flex flex-wrap items-start justify-between gap-3">
						<div><Card.Title class="text-base">Hasil & Analisis</Card.Title><Card.Description>Rekap nilai gabungan dari seluruh sesi dalam kegiatan ini; gunakan ekspor untuk analisis lanjutan.</Card.Description></div>
						<LoadingButton variant="outline" onclick={exportCSV} disabled={currentResults.length === 0} label="Ekspor CSV" />
					</div>
				</Card.Header>
				<Card.Content class="p-0 overflow-x-auto">
					<Table.Root>
						<Table.Header><Table.Row class="bg-muted/50"><Table.Head>NIS</Table.Head><Table.Head>Nama Siswa</Table.Head><Table.Head>Kelas</Table.Head><Table.Head>Sesi Ujian</Table.Head><Table.Head class="text-center">Skor</Table.Head><Table.Head>Status</Table.Head></Table.Row></Table.Header>
						<Table.Body>
							{#each currentResults as r (r.participant_id)}
								<Table.Row><Table.Cell class="font-mono text-sm">{r.nis}</Table.Cell><Table.Cell class="font-medium">{r.student_nama}</Table.Cell><Table.Cell><Badge variant="secondary" class="text-xs">{r.class_code || '-'}</Badge></Table.Cell><Table.Cell class="text-sm text-muted-foreground">{r.session_title}</Table.Cell><Table.Cell class="text-center font-bold text-success">{fmtScore(r.score)}</Table.Cell><Table.Cell>{#if r.submitted_at}<Badge variant="outline" class="bg-success/10 text-success border-success/20">Selesai</Badge>{:else}<Badge variant="outline" class="text-muted-foreground border-border">Belum</Badge>{/if}</Table.Cell></Table.Row>
							{:else}
								<Table.Row><Table.Cell colspan={6} class="py-12 text-center text-muted-foreground">Belum ada data nilai untuk kegiatan ini.</Table.Cell></Table.Row>
							{/each}
						</Table.Body>
					</Table.Root>
				</Card.Content>
			</Card.Root>
			{/if}
		{/snippet}
	</AsyncContent>
</div>
