<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Badge } from '$lib/components/ui/badge';
	import { toast } from '$lib/components/ui/sonner';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import OperationStatusPanel from '$lib/components/OperationStatusPanel.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import RichContent from '$lib/components/RichContent.svelte';
	import { confirmAction, confirmChallenge } from '$lib/confirm-dialog';
	import { clientApiPath, clientApiPathWithQuery, readClientApiData, readClientJson } from '$lib/client/api';
	import { cbtRoomSetupErrorMessage, roomReadinessMessage, roomReadinessTone, type CbtRoomReadiness } from '$lib/client/cbt-room-readiness';
	import { csvRow } from '$lib/csv';

	type SessionInfo = {
		id: string; title: string; package_title: string; duration_minutes: number;
		class_name: string; class_code: string; event_id: string | null;
		scope_type?: string; scope_ref?: string;
		scheduled_start: string; scheduled_end: string; status: string;
	};
	type ResultRow = {
		participant_id: string; nis: string; nama: string; gender: string;
		submitted_at: string | null; score: string | null;
		room_name?: string; seat_no?: number | null;
		total_answers: number; correct_answers: number;
	};
	type Participant = {
		id: string; nis: string; nama: string; gender: string;
		token: string; room_name: string; room_id: string | null;
		seat_no?: number | null;
		submitted_at: string | null; score: string | null;
		app_switch_count: number; screenshot_attempt: number; suspicious_flag: boolean;
		last_heartbeat: string | null;
	};
	type Room = {
		id: string; room_name: string; room_name_snapshot?: string;
		school_room_id?: string | null; school_room_code?: string; school_room_name?: string;
		school_room_building?: string; school_room_location_note?: string;
		school_room_exam_capacity?: number; school_room_condition?: string; school_room_exam_eligible?: boolean;
		capacity: number; participant_count: number; proctor_count?: number;
		primary_proctor_id?: string | null; primary_proctor_name?: string;
	};
	type SchoolRoom = {
		id: string; code: string; name: string; building: string; floor: string;
		room_type: string; location_note: string; default_capacity: number; exam_capacity: number;
		condition: string; is_exam_eligible: boolean; network_ready: boolean; power_ready: boolean;
		notes: string;
	};
	type EmployeeOption = { id: string; nip: string; nama: string; unit_kerja?: string; is_active?: boolean; };
	type RoomReadiness = CbtRoomReadiness;
	type ProctoringRow = {
		participant_id: string; nis: string; nama: string;
		token: string; room_name: string; seat_no?: number | null;
		submitted_at: string | null; last_heartbeat: string | null;
		app_switch_count: number; screenshot_attempt: number;
		suspicious_flag: boolean; answered_count: number; score: string | null;
	};
	type ProctoringEvent = {
		id: string;
		participant_id: string;
		student_id: string;
		nis: string;
		nama: string;
		room_name: string;
		event_type: string;
		event_data: unknown;
		created_at: string;
	};
	type AuditLog = {
		id: string;
		user_id: string;
		username: string;
		action: string;
		entity_type: string;
		entity_id: string;
		metadata: unknown;
		created_at: string;
	};
	type OperationalRecap = {
		session_id: string;
		session_title: string;
		session_status: string;
		scheduled_start: string;
		scheduled_end: string;
		room_count: number;
		handover_locked_count: number;
		handover_draft_count: number;
		handover_missing_count: number;
		incident_room_count: number;
		participant_count: number;
		unassigned_participant_count: number;
		joined_count: number;
		submitted_count: number;
		no_show_count: number;
		suspicious_count: number;
		app_switch_count: number;
		screenshot_attempt_count: number;
		incident_event_count: number;
		force_submit_count: number;
		reset_access_count: number;
		app_switch_event_count: number;
		screenshot_event_count: number;
	};
	type OperationalRoom = {
		room_id: string;
		session_id: string;
		room_name: string;
		room_token: string;
		room_status: string;
		room_is_locked: boolean;
		participant_count: number;
		joined_count: number;
		submitted_count: number;
		no_show_count: number;
		suspicious_count: number;
		missing_seat_count: number;
		app_switch_count: number;
		screenshot_attempt_count: number;
		incident_event_count: number;
		force_submit_count: number;
		reset_access_count: number;
		handover_id: string | null;
		attendance_checked: boolean;
		all_submitted_checked: boolean;
		device_issue_checked: boolean;
		room_clean_checked: boolean;
		token_returned_checked: boolean;
		assets_returned_checked: boolean;
		incident_notes: string;
		operator_notes: string;
		handover_notes: string;
		locked_at: string | null;
		handover_updated_at: string | null;
	};
	type OperationalRecapPayload = {
		recap?: OperationalRecap;
		rooms?: OperationalRoom[];
	};
	type ItemAnalysisRow = {
		position: number;
		points: number;
		question_id: string;
		question_code: string;
		question_text: string;
		question_type: string;
		difficulty: string;
		answer_key: string;
		cp_ref: string;
		tp_ref: string;
		kd_ref: string;
		material_topic: string;
		cognitive_level: string;
		hots_flag: boolean;
		submitted_count: number;
		answered_count: number;
		blank_count: number;
		correct_count: number;
		incorrect_count: number;
		unscored_count: number;
		avg_manual_score: number;
		difficulty_index: number;
		top_group_count: number;
		top_correct_count: number;
		bottom_group_count: number;
		bottom_correct_count: number;
		discrimination_index: number;
		answer_distribution: unknown;
		recommendation: string;
		recommendation_tone: string;
	};
	type ItemAnalysisPayload = {
		items?: ItemAnalysisRow[];
	};
	type QuestionRevisionResponse = {
		id?: string;
		code?: string;
		workflow_status?: string;
	};
	type UngradedEssay = {
		answer_id: string;
		participant_id: string;
		question_id: string;
		nis: string;
		nama: string;
		room_name?: string;
		question_code: string;
		question_text: string;
		stem_html?: string;
		stimulus_html?: string;
		rubric_html?: string;
		points?: unknown;
		answer: string;
	};
	type SessionResultsDetail = {
		session: SessionInfo;
		results: ResultRow[];
	};
	type ResultsPayload = {
		session?: SessionInfo | null;
		results?: ResultRow[];
		error?: string;
		message?: string;
	};

	const sessionId = page.params.id ?? '';
	const userRoles = $derived(page.data.user?.roles ?? (page.data.user?.role ? [page.data.user.role] : []));
	const isAdmin = $derived(userRoles.includes('admin'));
	type ActiveTab = 'hasil' | 'butir' | 'peserta' | 'ruangan' | 'operasional' | 'proctoring' | 'audit' | 'essay';
	type DetailNextAction = {
		title: string;
		message: string;
		label: string;
		tab?: ActiveTab;
		run?: 'auto_seats';
		tone: 'success' | 'warning' | 'info';
	};

	let activeTab = $state<ActiveTab>('hasil');
	let session = $state<SessionInfo | null>(null);
	let results = $state<ResultRow[]>([]);
	let participants = $state<Participant[]>([]);
	let rooms = $state<Room[]>([]);
	let schoolRooms = $state<SchoolRoom[]>([]);
	let employeeOptions = $state<EmployeeOption[]>([]);
	let roomReadiness = $state<RoomReadiness | null>(null);
	let operationalRecap = $state<OperationalRecap | null>(null);
	let operationalRooms = $state<OperationalRoom[]>([]);
	let itemAnalysis = $state<ItemAnalysisRow[]>([]);
	let proctoring = $state<ProctoringRow[]>([]);
	let proctoringEvents = $state<ProctoringEvent[]>([]);
	let auditLogs = $state<AuditLog[]>([]);
	let essays = $state<UngradedEssay[]>([]);
	let detailPromise = $state<Promise<SessionResultsDetail> | null>(null);
	let scoreBusy = $state(false);
	let shuffleBusy = $state(false);
	let selectedSchoolRoomId = $state('');
	let newRoomName = $state('');
	let roomNameCustomized = $state(false);
	let newRoomCap = $state(30);
	let roomBusy = $state(false);
	let seatBusy = $state(false);
	let tokenBusy = $state(false);
	let participantRefreshBusy = $state(false);
	let proctoringRefreshBusy = $state(false);
	let operationalRefreshBusy = $state(false);
	let commandCenterBusy = $state(false);
	let itemAnalysisRefreshBusy = $state(false);
	let eventRefreshBusy = $state(false);
	let auditRefreshBusy = $state(false);
	let regenBusyId = $state('');
	let roomDeleteBusyId = $state('');
	let proctorSaveBusyId = $state('');
	let seatSaveBusyId = $state('');
	let gradeBusyId = $state('');
	let flagBusyId = $state('');
	let resetAccessBusyId = $state('');
	let forceSubmitBusyId = $state('');
	let revisionBusyId = $state('');
	let eventPanelParticipantId = $state('');
	let procInterval: ReturnType<typeof setInterval> | null = null;
	let gradeInput = $state<Record<string, number>>({});
	let seatInput = $state<Record<string, number>>({});
	let roomInput = $state<Record<string, string>>({});
	let roomProctorInput = $state<Record<string, string>>({});
	let operationState = $state<{ tone: 'success' | 'error' | 'warning' | 'info'; title: string; message: string } | null>(null);
	let detailRequestId = 0;
	let participantsRequestId = 0;
	let roomsRequestId = 0;
	let schoolRoomsRequestId = 0;
	let employeesRequestId = 0;
	let readinessRequestId = 0;
	let operationalRequestId = 0;
	let itemAnalysisRequestId = 0;
	let proctoringRequestId = 0;
	let proctoringEventsRequestId = 0;
	let auditLogsRequestId = 0;
	let essaysRequestId = 0;

	const statusLabel: Record<string, string> = {
		draft: 'Draft', scheduled: 'Terjadwal', active: 'Berlangsung',
		finished: 'Selesai', cancelled: 'Dibatalkan',
	};
	type DetailTabGroup = {
		module: string;
		help: string;
		tabs: { id: ActiveTab; label: string }[];
	};
	const detailTabGroups: DetailTabGroup[] = [
		{
			module: 'Kegiatan & Sesi',
			help: 'Peserta/Ruang/Kartu',
			tabs: [
				{ id: 'peserta', label: 'Peserta' },
				{ id: 'ruangan', label: 'Ruang' },
			],
		},
		{
			module: 'Monitoring',
			help: 'Pantau ujian berjalan',
			tabs: [
				{ id: 'proctoring', label: 'Proctoring' },
				{ id: 'operasional', label: 'Operasional' },
				{ id: 'audit', label: 'Audit' },
			],
		},
		{
			module: 'Hasil & Analisis',
			help: 'Nilai, butir, uraian',
			tabs: [
				{ id: 'hasil', label: 'Hasil' },
				{ id: 'butir', label: 'Butir' },
				{ id: 'essay', label: 'Uraian' },
			],
		},
	];
	const detailTabs = detailTabGroups.flatMap((group) => group.tabs);

	function tabFromQuery(value: string | null): ActiveTab | null {
		const found = detailTabs.find((tab) => tab.id === value);
		return found?.id ?? null;
	}

	function statusClass(s: string) {
		if (s === 'active') return 'bg-emerald-100 text-emerald-700 border-emerald-200';
		if (s === 'finished') return 'bg-slate-100 text-slate-500 border-slate-200';
		if (s === 'cancelled') return 'bg-red-100 text-red-700 border-red-200';
		if (s === 'scheduled') return 'bg-green-100 text-green-800 border-green-200';
		return 'bg-amber-100 text-amber-700 border-amber-200';
	}

	function fmtDt(iso: string | null) {
		if (!iso) return '—';
		return new Date(iso).toLocaleString('id-ID', {
			timeZone: 'Asia/Makassar', year: 'numeric', month: 'short',
			day: 'numeric', hour: '2-digit', minute: '2-digit',
		}) + ' WITA';
	}

	function fmtScore(score: string | null) {
		if (score === null || score === undefined || score === '') return '—';
		const n = parseFloat(score);
		return isNaN(n) ? '—' : n.toFixed(1);
	}

	function selectedSchoolRoom() {
		return schoolRooms.find((room) => room.id === selectedSchoolRoomId) ?? null;
	}

	function roomSetupLocked(status: string | null | undefined) {
		return status === 'active' || status === 'finished';
	}

	function selectSchoolRoom(value: string) {
		selectedSchoolRoomId = value;
		const room = schoolRooms.find((item) => item.id === value);
		if (!room) {
			if (!roomNameCustomized) newRoomName = '';
			return;
		}
		newRoomCap = room.exam_capacity || room.default_capacity || 30;
		if (!roomNameCustomized) newRoomName = room.name;
	}

	function updateNewRoomName(value: string) {
		newRoomName = value;
		roomNameCustomized = true;
	}

	function roomCapacityRatio(room: Room) {
		if (room.capacity <= 0) return 0;
		return Math.min(100, (room.participant_count / room.capacity) * 100);
	}

	function nextDetailAction(readiness: RoomReadiness | null): DetailNextAction {
		if (!readiness) {
			return {
				title: 'Cek kesiapan operasional',
				message: 'Muat data ruang, peserta, kapasitas, nomor meja, dan pengawas sebelum sesi dimulai.',
				label: 'Cek Kesiapan',
				tab: 'ruangan',
				tone: 'info',
			};
		}
		if (readiness.participant_count === 0) {
			return {
				title: 'Peserta belum terdaftar',
				message: 'Daftarkan peserta dari daftar sesi, lalu kembali untuk mengatur ruangan dan pengawas.',
				label: 'Lihat Peserta',
				tab: 'peserta',
				tone: 'warning',
			};
		}
		if (readiness.room_count === 0 || readiness.total_capacity < readiness.participant_count) {
			return {
				title: 'Ruang ujian belum cukup',
				message: `${readiness.room_count} ruang tersedia dengan kapasitas ${readiness.total_capacity} untuk ${readiness.participant_count} peserta.`,
				label: 'Atur Ruang',
				tab: 'ruangan',
				tone: 'warning',
			};
		}
		if (readiness.unassigned_participant_count > 0) {
			return {
				title: 'Peserta belum masuk ruang',
				message: `${readiness.unassigned_participant_count} peserta belum punya ruang ujian.`,
				label: 'Acak Ruang',
				tab: 'ruangan',
				tone: 'warning',
			};
		}
		if (readiness.missing_seat_count > 0) {
			return {
				title: 'Nomor meja belum lengkap',
				message: `${readiness.missing_seat_count} peserta belum punya nomor meja.`,
				label: 'Atur Nomor Meja',
				run: 'auto_seats',
				tone: 'warning',
			};
		}
		if (readiness.rooms_without_proctor > 0) {
			return {
				title: 'Pengawas belum lengkap',
				message: `${readiness.rooms_without_proctor} ruang belum punya pengawas.`,
				label: 'Tetapkan Pengawas',
				tab: 'ruangan',
				tone: 'warning',
			};
		}
		return {
			title: 'Sesi siap dipantau',
			message: 'Peserta, ruang, kapasitas, nomor meja, dan pengawas sudah siap.',
			label: 'Pantau Proctoring',
			tab: 'proctoring',
			tone: 'success',
		};
	}

	function detailActionPanelClass(tone: DetailNextAction['tone']) {
		if (tone === 'warning') return 'border-amber-200 bg-amber-50';
		if (tone === 'success') return 'border-emerald-200 bg-emerald-50';
		return 'border-slate-200 bg-slate-50';
	}

	function operationalTone(recap: OperationalRecap | null) {
		if (!recap) return 'info';
		if (recap.handover_missing_count > 0 || recap.unassigned_participant_count > 0) return 'warning';
		if (recap.incident_room_count > 0 || recap.incident_event_count > 0 || recap.force_submit_count > 0) return 'warning';
		return 'success';
	}

	function operationalMessage(recap: OperationalRecap | null) {
		if (!recap) return 'Rekap operasional belum dimuat.';
		return `${recap.handover_locked_count}/${recap.room_count} ruang sudah mengunci handover, ${recap.submitted_count}/${recap.participant_count} peserta submit, ${recap.incident_room_count} ruang punya catatan insiden, ${recap.force_submit_count} peserta dipaksa submit.`;
	}

	function isRecord(value: unknown): value is Record<string, unknown> {
		return typeof value === 'object' && value !== null;
	}

	function percent(value: number) {
		if (!Number.isFinite(value)) return '0%';
		return `${Math.round(value * 100)}%`;
	}

	function questionTypeLabel(type: string) {
		const labels: Record<string, string> = {
			multiple_choice: 'PG',
			multiple_answer: 'PG Kompleks',
			true_false: 'Benar/Salah',
			agree_disagree: 'Setuju/Tidak',
			matching: 'Menjodohkan',
			short_answer: 'Isian',
			essay: 'Essay',
		};
		return labels[type] ?? type.replaceAll('_', ' ');
	}

	function itemAnalysisToneClass(tone: string) {
		if (tone === 'success') return 'border-emerald-300 bg-emerald-50 text-emerald-700';
		if (tone === 'danger') return 'border-red-300 bg-red-50 text-red-700';
		if (tone === 'info') return 'border-slate-300 bg-slate-50 text-slate-600';
		return 'border-amber-300 bg-amber-50 text-amber-700';
	}

	function itemAnalysisSignalClass(row: ItemAnalysisRow) {
		if (row.recommendation_tone === 'danger') return 'bg-red-50/70';
		if (row.recommendation_tone === 'warning') return 'bg-amber-50/60';
		return '';
	}

	function answerDistributionEntries(row: ItemAnalysisRow) {
		if (!isRecord(row.answer_distribution)) return [];
		return Object.entries(row.answer_distribution)
			.map(([answer, count]) => ({ answer, count: Number(count) || 0 }))
			.sort((a, b) => b.count - a.count || a.answer.localeCompare(b.answer));
	}

	function shouldOfferQuestionRevision(row: ItemAnalysisRow) {
		return row.recommendation_tone === 'warning' || row.recommendation_tone === 'danger';
	}

	function itemRevisionNotes(row: ItemAnalysisRow) {
		const sessionTitle = session?.title || 'Sesi CBT';
		const answeredText = `${row.answered_count}/${row.submitted_count} dijawab`;
		const scoreSignal = row.question_type === 'essay'
			? `skor rata-rata ${row.avg_manual_score.toFixed(1)}, belum dinilai ${row.unscored_count}`
			: `benar ${row.correct_count}, salah ${row.incorrect_count}, kosong ${row.blank_count}`;
		return [
			`Revisi dari analisis butir sesi "${sessionTitle}".`,
			`Rekomendasi: ${row.recommendation}.`,
			`Kesukaran: ${percent(row.difficulty_index)}; daya pembeda: ${percent(row.discrimination_index)}; ${answeredText}; ${scoreSignal}.`,
		].join(' ');
	}

	function handoverStatusLabel(room: OperationalRoom) {
		if (room.locked_at) return 'Terkunci';
		if (room.handover_id) return 'Draft';
		return 'Belum ada';
	}

	function handoverStatusClass(room: OperationalRoom) {
		if (room.locked_at) return 'border-emerald-300 bg-emerald-50 text-emerald-700';
		if (room.handover_id) return 'border-amber-300 bg-amber-50 text-amber-700';
		return 'border-red-300 bg-red-50 text-red-700';
	}

	function fmtEssayPoints(points: unknown) {
		if (typeof points === 'number') return points.toFixed(points % 1 === 0 ? 0 : 1);
		if (typeof points === 'string' && points.trim()) return points;
		return '1';
	}

	function scoreClass(score: string | null) {
		if (!score) return 'text-slate-400';
		const n = parseFloat(score);
		if (n >= 75) return 'text-emerald-600 font-semibold';
		if (n >= 60) return 'text-amber-600 font-semibold';
		return 'text-red-600 font-semibold';
	}

	function heartbeatBucket(hb: string | null): 'none' | 'online' | 'slow' | 'offline' {
		if (!hb) return 'none';
		const diff = (Date.now() - new Date(hb).getTime()) / 1000;
		if (diff < 60) return 'online';
		if (diff < 180) return 'slow';
		return 'offline';
	}

	function heartbeatStatus(hb: string | null): { label: string; cls: string } {
		const bucket = heartbeatBucket(hb);
		if (bucket === 'online') return { label: 'Online', cls: 'text-emerald-600' };
		if (bucket === 'slow') return { label: 'Lambat', cls: 'text-amber-600' };
		if (bucket === 'offline') return { label: 'Offline', cls: 'text-red-600' };
		return { label: 'Belum login', cls: 'text-slate-400' };
	}

	let stats = $derived({
		total: results.length,
		submitted: results.filter(r => r.submitted_at).length,
		avgScore: results.length > 0
			? results.reduce((sum, r) => sum + (r.score ? parseFloat(r.score) : 0), 0) / results.length
			: 0,
		passing: results.filter(r => r.score && parseFloat(r.score) >= 75).length,
	});

	let proctoringStats = $derived.by(() => {
		let online = 0;
		let slow = 0;
		let offline = 0;
		let submitted = 0;
		let suspicious = 0;
		let appSwitches = 0;
		let screenshots = 0;
		for (const row of proctoring) {
			const bucket = heartbeatBucket(row.last_heartbeat);
			if (row.submitted_at) submitted += 1;
			else if (bucket === 'online') online += 1;
			else if (bucket === 'slow') slow += 1;
			else if (bucket === 'offline') offline += 1;
			if (row.suspicious_flag) suspicious += 1;
			appSwitches += row.app_switch_count;
			screenshots += row.screenshot_attempt;
		}
		return { online, slow, offline, submitted, suspicious, appSwitches, screenshots };
	});

	let commandCenterIssues = $derived.by(() => {
		const issues: { label: string; tone: 'danger' | 'warning' | 'info'; tab: ActiveTab }[] = [];
		if (!roomReadiness) {
			issues.push({ label: 'Kesiapan ruang belum dimuat', tone: 'info', tab: 'ruangan' });
		} else {
			if (roomReadiness.participant_count === 0) issues.push({ label: 'Belum ada peserta', tone: 'warning', tab: 'peserta' });
			if (roomReadiness.total_capacity < roomReadiness.participant_count) issues.push({ label: 'Kapasitas ruang kurang', tone: 'danger', tab: 'ruangan' });
			if (roomReadiness.unassigned_participant_count > 0) issues.push({ label: `${roomReadiness.unassigned_participant_count} peserta belum punya ruang`, tone: 'warning', tab: 'ruangan' });
			if (roomReadiness.missing_seat_count > 0) issues.push({ label: `${roomReadiness.missing_seat_count} nomor meja kosong`, tone: 'warning', tab: 'ruangan' });
			if (roomReadiness.rooms_without_proctor > 0) issues.push({ label: `${roomReadiness.rooms_without_proctor} ruang belum ada pengawas`, tone: 'danger', tab: 'ruangan' });
		}
		if (operationalRecap) {
			if (operationalRecap.incident_room_count > 0) issues.push({ label: `${operationalRecap.incident_room_count} ruang punya catatan insiden`, tone: 'warning', tab: 'operasional' });
			if (operationalRecap.handover_missing_count > 0) issues.push({ label: `${operationalRecap.handover_missing_count} handover belum dibuat`, tone: 'warning', tab: 'operasional' });
			if (operationalRecap.suspicious_count > 0) issues.push({ label: `${operationalRecap.suspicious_count} peserta perlu atensi`, tone: 'danger', tab: 'proctoring' });
		}
		if (proctoringStats.slow + proctoringStats.offline > 0) issues.push({ label: `${proctoringStats.slow + proctoringStats.offline} koneksi lambat/offline`, tone: 'warning', tab: 'proctoring' });
		return issues;
	});

	let commandCenterMetrics = $derived.by(() => [
		{
			label: 'Ruang',
			value: roomReadiness ? `${roomReadiness.room_count}` : '—',
			helper: roomReadiness ? `${roomReadiness.total_capacity} kursi / ${roomReadiness.participant_count} peserta` : 'Belum dimuat',
			tab: 'ruangan' as ActiveTab,
		},
		{
			label: 'Pengawas',
			value: roomReadiness ? `${Math.max(roomReadiness.room_count - roomReadiness.rooms_without_proctor, 0)}/${roomReadiness.room_count}` : '—',
			helper: roomReadiness ? `${roomReadiness.rooms_without_proctor} ruang kosong` : 'Belum dimuat',
			tab: 'ruangan' as ActiveTab,
		},
		{
			label: 'Submit',
			value: operationalRecap ? `${operationalRecap.submitted_count}/${operationalRecap.participant_count}` : `${stats.submitted}/${stats.total}`,
			helper: operationalRecap ? `${operationalRecap.joined_count} login, ${operationalRecap.no_show_count} belum hadir` : 'Dari hasil ujian',
			tab: 'hasil' as ActiveTab,
		},
		{
			label: 'Koneksi',
			value: `${proctoringStats.online}`,
			helper: `${proctoringStats.slow} lambat, ${proctoringStats.offline} offline`,
			tab: 'proctoring' as ActiveTab,
		},
		{
			label: 'Atensi',
			value: `${(operationalRecap?.suspicious_count ?? proctoringStats.suspicious) + (operationalRecap?.incident_room_count ?? 0)}`,
			helper: operationalRecap ? `${operationalRecap.force_submit_count} paksa submit, ${operationalRecap.reset_access_count} reset akses` : 'Dari proctoring aktif',
			tab: 'operasional' as ActiveTab,
		},
		{
			label: 'Handover',
			value: operationalRecap ? `${operationalRecap.handover_locked_count}/${operationalRecap.room_count}` : '—',
			helper: operationalRecap ? `${operationalRecap.handover_missing_count} belum dibuat` : 'Belum dimuat',
			tab: 'operasional' as ActiveTab,
		},
	]);

	function commandCenterClass() {
		if (commandCenterIssues.some((issue) => issue.tone === 'danger')) return 'border-red-200 bg-red-50/60';
		if (commandCenterIssues.some((issue) => issue.tone === 'warning')) return 'border-amber-200 bg-amber-50/70';
		return 'border-emerald-200 bg-emerald-50/60';
	}

	function commandIssueClass(tone: 'danger' | 'warning' | 'info') {
		if (tone === 'danger') return 'border-red-200 bg-white text-red-700';
		if (tone === 'warning') return 'border-amber-200 bg-white text-amber-700';
		return 'border-slate-200 bg-white text-slate-600';
	}

	let itemAnalysisStats = $derived.by(() => {
		const submittedRows = itemAnalysis.filter((row) => row.submitted_count > 0);
		const reviewRows = itemAnalysis.filter((row) => row.recommendation_tone === 'warning' || row.recommendation_tone === 'danger');
		const unscoredRows = itemAnalysis.filter((row) => row.unscored_count > 0);
		const avgDifficulty = submittedRows.length > 0
			? submittedRows.reduce((sum, row) => sum + row.difficulty_index, 0) / submittedRows.length
			: 0;
		const lowDiscrimination = itemAnalysis.filter((row) => row.submitted_count >= 3 && row.discrimination_index < 0.15).length;
		return {
			total: itemAnalysis.length,
			review: reviewRows.length,
			unscored: unscoredRows.length,
			avgDifficulty,
			lowDiscrimination,
		};
	});

	function showToast(msg: string, ok = true) {
		if (ok) toast.success(msg);
		else toast.error(msg);
	}

	function setOperationState(
		tone: 'success' | 'error' | 'warning' | 'info',
		title: string,
		message: string,
	) {
		operationState = { tone, title, message };
	}

	function confirmPhrase(title: string, detail: string, challenge: string) {
		return confirmChallenge({
			title,
			message: detail,
			challenge,
			confirmLabel: 'Konfirmasi',
			tone: 'danger'
		});
	}

	async function fetchSessionDetail(): Promise<SessionResultsDetail> {
		const payload = await fetch(`/api/asesmen/sessions/${sessionId}/results`)
			.then((response) => readClientApiData<ResultsPayload>(response, 'Gagal memuat hasil ujian'));
		if (!payload.session) throw new Error('Data sesi tidak ditemukan');
		return {
			session: payload.session,
			results: payload.results ?? [],
		};
	}

	function applySessionDetail(detail: SessionResultsDetail) {
		session = detail.session;
		results = detail.results;
	}

	function loadInitial() {
		const requestId = ++detailRequestId;
		session = null;
		results = [];
		detailPromise = fetchSessionDetail().then((detail) => {
			if (requestId !== detailRequestId) {
				if (!session) throw new Error('Permintaan detail sesi dibatalkan');
				return { session, results };
			}
			applySessionDetail(detail);
			return detail;
		}).catch((error: unknown) => {
			if (requestId === detailRequestId || !session) throw error;
			return { session, results };
		});
	}

	async function refreshSessionDetail() {
		if (!detailPromise) {
			loadInitial();
			return;
		}
		const requestId = ++detailRequestId;
		try {
			const detail = await fetchSessionDetail();
			if (requestId !== detailRequestId) return;
			applySessionDetail(detail);
			detailPromise = Promise.resolve(detail);
		} catch (error) {
			if (requestId !== detailRequestId) return;
			if (session) {
				detailPromise = Promise.resolve({ session, results });
				toast.error(detailErrorMessage(error));
			} else {
				detailPromise = Promise.reject(error);
			}
		}
	}

	function retryDetail(reset?: () => void) {
		reset?.();
		loadInitial();
	}

	function detailErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Gagal memuat hasil ujian';
	}

	function handleDetailRenderError(error: unknown) {
		console.error('CBT session detail render failed', error);
	}

	function mutationErrorMessage(error: unknown, fallback: string) {
		if (error instanceof Error && error.message.trim() && !error.message.toLowerCase().includes('fetch')) return error.message;
		return fallback;
	}

	function proctoringEventLabel(type: string): string {
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
			proctor_force_submit: 'Paksa Submit',
		};
		return map[type] ?? type.replaceAll('_', ' ');
	}

	function eventDataText(data: unknown): string {
		if (data === null || data === undefined || data === '') return '—';
		const text = typeof data === 'string' ? data : JSON.stringify(data);
		if (!text) return '—';
		return text.length > 140 ? `${text.slice(0, 140)}...` : text;
	}

	function auditMetadata(meta: unknown): Record<string, unknown> {
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

	function auditValue(log: AuditLog, key: string): string {
		const value = auditMetadata(log.metadata)[key];
		return typeof value === 'string' ? value : '';
	}

	function auditActionLabel(action: string): string {
		const labels: Record<string, string> = {
			CBT_SESSION_SCHEDULE_UPDATE: 'Ubah Jadwal',
			CBT_SESSION_PARTICIPANT_FLAG: 'Flag Peserta',
			CBT_SESSION_PARTICIPANT_RESET_ACCESS: 'Reset Akses',
			CBT_SESSION_PARTICIPANT_FORCE_SUBMIT: 'Paksa Submit',
			CBT_SESSION_ESSAY_GRADE: 'Koreksi Uraian',
			CBT_SESSION_SCORE: 'Hitung Skor',
			CBT_SESSION_ROOM_HANDOVER_SAVE: 'Simpan Handover',
			CBT_SESSION_ROOM_HANDOVER_LOCK: 'Kunci Handover',
		};
		return labels[action] ?? action.replaceAll('_', ' ');
	}

	function auditSummary(log: AuditLog): string {
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
			.join(' · ') || 'Detail tersedia di metadata audit.';
	}

	async function loadParticipants() {
		const requestId = ++participantsRequestId;
		try {
			const res = await fetch(`/api/asesmen/sessions/${sessionId}/participants`);
			const rows = await readClientApiData<Participant[]>(res, 'Gagal memuat peserta');
			if (requestId !== participantsRequestId) return;
			participants = Array.isArray(rows) ? rows : [];
			seatInput = Object.fromEntries(participants.map((participant) => [participant.id, participant.seat_no ?? 0]));
			roomInput = Object.fromEntries(participants.map((participant) => [participant.id, participant.room_id ?? '']));
		} catch (error) {
			if (requestId === participantsRequestId) throw error;
		}
	}

	async function loadRooms() {
		const requestId = ++roomsRequestId;
		try {
			const res = await fetch(`/api/asesmen/sessions/${sessionId}/rooms`);
			const rows = await readClientApiData<Room[]>(res, 'Gagal memuat ruangan');
			if (requestId !== roomsRequestId) return;
			rooms = Array.isArray(rows) ? rows : [];
			roomProctorInput = Object.fromEntries(rooms.map((room) => [room.id, room.primary_proctor_id ?? '']));
		} catch (error) {
			if (requestId === roomsRequestId) throw error;
		}
	}

	async function loadSchoolRooms() {
		const requestId = ++schoolRoomsRequestId;
		try {
			const res = await fetch('/api/inventory/rooms?exam_eligible=true');
			const rows = await readClientApiData<SchoolRoom[]>(res, 'Gagal memuat master ruangan');
			if (requestId !== schoolRoomsRequestId) return;
			schoolRooms = Array.isArray(rows) ? rows : [];
		} catch (error) {
			if (requestId === schoolRoomsRequestId) {
				schoolRooms = [];
				console.warn('Master ruangan tidak dapat dimuat dari inventaris', error);
			}
		}
	}

	async function loadEmployeeOptions() {
		const requestId = ++employeesRequestId;
		try {
			const res = await fetch('/api/employees');
			const data = await readClientApiData<{ items?: EmployeeOption[] } | EmployeeOption[]>(res, 'Gagal memuat pegawai');
			if (requestId !== employeesRequestId) return;
			employeeOptions = Array.isArray(data) ? data : Array.isArray(data?.items) ? data.items : [];
		} catch {
			if (requestId === employeesRequestId) employeeOptions = [];
		}
	}

	async function loadRoomReadiness() {
		const requestId = ++readinessRequestId;
		try {
			const res = await fetch(`/api/asesmen/sessions/${sessionId}/rooms/readiness`);
			const data = await readClientApiData<RoomReadiness>(res, 'Gagal memuat kesiapan ruangan');
			if (requestId !== readinessRequestId) return;
			roomReadiness = data;
		} catch (error) {
			if (requestId === readinessRequestId) throw error;
		}
	}

	async function loadOperationalRecap() {
		const requestId = ++operationalRequestId;
		try {
			const res = await fetch(`/api/asesmen/sessions/${sessionId}/operational-recap`);
			const data = await readClientApiData<OperationalRecapPayload>(res, 'Gagal memuat rekap operasional');
			if (requestId !== operationalRequestId) return;
			operationalRecap = data.recap ?? null;
			operationalRooms = Array.isArray(data.rooms) ? data.rooms : [];
		} catch (error) {
			if (requestId === operationalRequestId) throw error;
		}
	}

	async function loadCommandCenterSnapshot() {
		const results = await Promise.allSettled([
			loadRoomReadiness(),
			loadOperationalRecap(),
			loadProctoring(),
		]);
		const failed = results.find((result) => result.status === 'rejected');
		if (failed && !roomReadiness && !operationalRecap && proctoring.length === 0) {
			throw failed.reason;
		}
	}

	async function loadItemAnalysis() {
		const requestId = ++itemAnalysisRequestId;
		try {
			const res = await fetch(`/api/asesmen/sessions/${sessionId}/item-analysis`);
			const data = await readClientApiData<ItemAnalysisPayload>(res, 'Gagal memuat analisis butir');
			if (requestId !== itemAnalysisRequestId) return;
			itemAnalysis = Array.isArray(data.items) ? data.items : [];
		} catch (error) {
			if (requestId === itemAnalysisRequestId) throw error;
		}
	}

	async function loadProctoring() {
		const requestId = ++proctoringRequestId;
		try {
			const res = await fetch(`/api/asesmen/sessions/${sessionId}/proctoring`);
			const rows = await readClientApiData<ProctoringRow[]>(res, 'Gagal memuat proctoring');
			if (requestId !== proctoringRequestId) return;
			proctoring = Array.isArray(rows) ? rows : [];
		} catch (error) {
			if (requestId === proctoringRequestId) throw error;
		}
	}

	async function loadProctoringEvents(participantId = eventPanelParticipantId) {
		const requestId = ++proctoringEventsRequestId;
		const params = new URLSearchParams({ limit: '100' });
		if (participantId) params.set('participant_id', participantId);
		try {
			const res = await fetch(clientApiPathWithQuery(clientApiPath`/api/asesmen/sessions/${sessionId}/proctoring/events`, params));
			const rows = await readClientApiData<ProctoringEvent[]>(res, 'Gagal memuat log proctoring');
			if (requestId !== proctoringEventsRequestId) return;
			proctoringEvents = Array.isArray(rows) ? rows : [];
		} catch (error) {
			if (requestId === proctoringEventsRequestId) throw error;
		}
	}

	async function loadAuditLogs() {
		const requestId = ++auditLogsRequestId;
		const params = new URLSearchParams({ per_page: '30' });
		try {
			const res = await fetch(clientApiPathWithQuery(clientApiPath`/api/asesmen/sessions/${sessionId}/audit-logs`, params));
			const rows = await readClientApiData<AuditLog[]>(res, 'Gagal memuat audit operasi');
			if (requestId !== auditLogsRequestId) return;
			auditLogs = Array.isArray(rows) ? rows : [];
		} catch (error) {
			if (requestId === auditLogsRequestId) throw error;
		}
	}

	async function showParticipantEvents(pid: string) {
		eventPanelParticipantId = eventPanelParticipantId === pid ? '' : pid;
		eventRefreshBusy = true;
		try {
			await loadProctoringEvents();
		} catch (error) {
			toast.error(detailErrorMessage(error));
		} finally {
			eventRefreshBusy = false;
		}
	}

	async function refreshParticipants() {
		participantRefreshBusy = true;
		try {
			await loadParticipants();
		} catch (error) {
			toast.error(detailErrorMessage(error));
		} finally {
			participantRefreshBusy = false;
		}
	}

	async function refreshProctoring() {
		proctoringRefreshBusy = true;
		try {
			await Promise.all([loadProctoring(), loadProctoringEvents()]);
		} catch (error) {
			toast.error(detailErrorMessage(error));
		} finally {
			proctoringRefreshBusy = false;
		}
	}

	async function refreshOperationalRecap() {
		operationalRefreshBusy = true;
		try {
			await loadOperationalRecap();
		} catch (error) {
			toast.error(detailErrorMessage(error));
		} finally {
			operationalRefreshBusy = false;
		}
	}

	async function refreshCommandCenter() {
		commandCenterBusy = true;
		try {
			await loadCommandCenterSnapshot();
		} catch (error) {
			toast.error(detailErrorMessage(error));
		} finally {
			commandCenterBusy = false;
		}
	}

	async function refreshItemAnalysis() {
		itemAnalysisRefreshBusy = true;
		try {
			await loadItemAnalysis();
		} catch (error) {
			toast.error(detailErrorMessage(error));
		} finally {
			itemAnalysisRefreshBusy = false;
		}
	}

	async function markQuestionRevision(row: ItemAnalysisRow) {
		if (revisionBusyId) return;
		const questionLabel = row.question_code || `nomor ${row.position}`;
		if (!(await confirmAction({
			title: 'Buat Draft Revisi Soal',
			message: `Sistem akan menduplikasi butir ${questionLabel} menjadi draft revisi di bank soal. Soal asli tetap aman untuk riwayat ujian.`,
			confirmLabel: 'Buat Draft Revisi',
			tone: 'warning',
		}))) return;

		revisionBusyId = row.question_id;
		try {
			const res = await fetch(clientApiPath`/api/bank-soal/questions/${row.question_id}/revision`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ notes: itemRevisionNotes(row) }),
			});
			const created = await readClientJson<QuestionRevisionResponse>(res);
			const codeText = created.code ? ` (${created.code})` : '';
			setOperationState('success', 'Draft Revisi Dibuat', `Butir ${questionLabel} sudah dibuat sebagai draft revisi${codeText}. Buka Bank Soal untuk menyunting isi, opsi, kunci, atau rubriknya.`);
			toast.success('Draft revisi soal dibuat');
			await loadItemAnalysis();
		} catch (error) {
			const message = mutationErrorMessage(error, 'Gagal membuat draft revisi soal');
			setOperationState('error', 'Draft Revisi Gagal', message);
			toast.error(message);
		} finally {
			revisionBusyId = '';
		}
	}

	async function refreshProctoringEvents() {
		eventRefreshBusy = true;
		try {
			await loadProctoringEvents();
		} catch (error) {
			toast.error(detailErrorMessage(error));
		} finally {
			eventRefreshBusy = false;
		}
	}

	async function refreshAuditLogs() {
		auditRefreshBusy = true;
		try {
			await loadAuditLogs();
		} catch (error) {
			toast.error(detailErrorMessage(error));
		} finally {
			auditRefreshBusy = false;
		}
	}

	async function loadEssays() {
		const requestId = ++essaysRequestId;
		try {
			const res = await fetch(`/api/asesmen/sessions/${sessionId}/ungraded-essays`);
			const rows = await readClientApiData<UngradedEssay[]>(res, 'Gagal memuat esai belum dinilai');
			if (requestId !== essaysRequestId) return;
			essays = Array.isArray(rows) ? rows : [];
		} catch (error) {
			if (requestId === essaysRequestId) throw error;
		}
	}

	async function switchTab(tab: ActiveTab) {
		activeTab = tab;
		try {
			if (tab === 'peserta') { await loadRooms(); await loadParticipants(); }
			if (tab === 'ruangan') { await Promise.all([loadRooms(), loadParticipants(), loadSchoolRooms(), loadEmployeeOptions(), loadRoomReadiness()]); }
			if (tab === 'operasional') await loadOperationalRecap();
			if (tab === 'butir') await loadItemAnalysis();
			if (tab === 'audit') await loadAuditLogs();
			if (tab === 'essay') await loadEssays();
			if (tab === 'proctoring') {
				await Promise.all([loadProctoring(), loadProctoringEvents()]);
				if (!procInterval) {
					procInterval = setInterval(() => {
						void loadProctoring().catch((error) => toast.error(detailErrorMessage(error)));
					}, 15000);
				}
			} else {
				if (procInterval) { clearInterval(procInterval); procInterval = null; }
			}
		} catch (error) {
			toast.error(detailErrorMessage(error));
		}
	}

	async function triggerScoring() {
		if (!(await confirmPhrase('Hitung Ulang Skor', 'Sistem akan menghitung ulang skor seluruh peserta berdasarkan jawaban yang masuk. Gunakan setelah koreksi uraian atau sinkronisasi nilai.', 'NILAI ULANG'))) return;
		scoreBusy = true;
		try {
			const res = await fetch(`/api/asesmen/sessions/${sessionId}/score`, { method: 'POST' });
			await readClientJson<unknown>(res);
			setOperationState('success', 'Skor Diperbarui', 'Perhitungan nilai sesi sudah disegarkan. Tinjau kembali hasil akhir sebelum menutup sesi.');
			showToast('Penilaian selesai — skor diperbarui');
			await refreshSessionDetail();
			if (activeTab === 'butir') await loadItemAnalysis();
		} catch (error) {
			setOperationState('error', 'Skor Gagal Dihitung Ulang', 'Perhitungan ulang belum berhasil. Periksa data jawaban atau ulangi beberapa saat lagi.');
			showToast(mutationErrorMessage(error, 'Gagal menghitung skor'), false);
		} finally { scoreBusy = false; }
	}

	async function generateTokens() {
		if (!(await confirmPhrase('Buat Token Massal', 'Token baru akan dibuat untuk seluruh peserta sesi ini. Gunakan hanya saat token awal belum dibagikan atau harus direset terkontrol.', 'TOKEN'))) return;
		tokenBusy = true;
		try {
			const res = await fetch(`/api/asesmen/sessions/${sessionId}/generate-tokens`, { method: 'POST' });
			await readClientJson<unknown>(res);
			setOperationState('success', 'Token Massal Berhasil Dibuat', 'Token peserta sudah diperbarui. Bagikan ulang token hanya ke pengawas atau peserta yang berwenang.');
			showToast('Token berhasil digenerate');
			await loadParticipants();
		} catch (error) {
			setOperationState('error', 'Token Gagal Dibuat', 'Pembuatan token massal belum berhasil. Ulangi setelah memeriksa daftar peserta sesi ini.');
			showToast(mutationErrorMessage(error, 'Gagal generate token'), false);
		} finally {
			tokenBusy = false;
		}
	}

	async function regenerateToken(pid: string) {
		if (!(await confirmAction({
			title: 'Buat Ulang Token Peserta',
			message: 'Buat ulang token peserta ini? Token lama tidak sebaiknya dipakai lagi setelah tindakan ini.',
			confirmLabel: 'Buat Ulang Token',
			tone: 'warning'
		}))) return;
		regenBusyId = pid;
		try {
			const res = await fetch(`/api/asesmen/sessions/${sessionId}/participants/${pid}/regenerate-token`, { method: 'POST' });
			await readClientJson<unknown>(res);
			setOperationState('warning', 'Token Peserta Diperbarui', 'Token lama untuk peserta terkait sebaiknya tidak dipakai lagi. Pastikan pengawas membagikan token terbaru.');
			showToast('Token diperbarui');
			await loadParticipants();
		} catch (error) {
			setOperationState('error', 'Token Gagal Diperbarui', 'Pembuatan ulang token peserta belum berhasil. Coba ulang beberapa saat lagi.');
			showToast(mutationErrorMessage(error, 'Gagal regenerate token'), false);
		} finally {
			regenBusyId = '';
		}
	}

	async function createRoom() {
		if (roomSetupLocked(session?.status)) {
			setOperationState('warning', 'Ruangan Terkunci', 'Ruangan terkunci setelah sesi aktif/selesai.');
			return;
		}
		if (!newRoomName.trim() && !selectedSchoolRoomId) return;
		const masterRoom = selectedSchoolRoom();
		const roomLabel = newRoomName.trim() || masterRoom?.name || 'Ruangan';
		roomBusy = true;
		try {
			const res = await fetch(`/api/asesmen/sessions/${sessionId}/rooms`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					school_room_id: selectedSchoolRoomId || undefined,
					room_name: newRoomName.trim(),
					capacity: newRoomCap,
				}),
			});
			await readClientJson<unknown>(res);
			setOperationState('success', 'Ruangan Ditambahkan', `Ruangan "${roomLabel}" sudah tersimpan. Pastikan kapasitas dan pengawasnya siap sebelum peserta diacak.`);
			showToast(`Ruangan "${roomLabel}" ditambahkan`);
			selectedSchoolRoomId = ''; newRoomName = ''; roomNameCustomized = false; newRoomCap = 30;
			await Promise.all([loadRooms(), loadRoomReadiness()]);
		} catch (error) {
			const message = cbtRoomSetupErrorMessage(error, 'Ruangan baru belum berhasil disimpan. Periksa nama atau kapasitas lalu coba lagi.');
			setOperationState('error', 'Ruangan Gagal Ditambahkan', message);
			showToast(message, false);
		} finally {
			roomBusy = false;
		}
	}

	async function deleteRoom(rid: string, roomName: string) {
		if (roomSetupLocked(session?.status)) {
			setOperationState('warning', 'Ruangan Terkunci', 'Ruangan terkunci setelah sesi aktif/selesai.');
			return;
		}
		if (!(await confirmPhrase('Hapus Ruangan Sesi', `Ruangan "${roomName}" akan dihapus dari sesi dan peserta di dalamnya akan dilepas dari alokasi ruangan.`, 'RUANGAN'))) return;
		roomDeleteBusyId = rid;
		try {
			const res = await fetch(`/api/asesmen/sessions/${sessionId}/rooms/${rid}`, { method: 'DELETE' });
			await readClientJson<unknown>(res);
			setOperationState('warning', 'Ruangan Dihapus', `Ruangan "${roomName}" dihapus dan peserta yang terkait perlu dialokasikan ulang.`);
			await loadRooms(); await loadParticipants(); await loadRoomReadiness();
		} catch (error) {
			const message = cbtRoomSetupErrorMessage(error, 'Ruangan belum berhasil dihapus. Pastikan sesi masih bisa diubah lalu coba lagi.');
			setOperationState('error', 'Ruangan Gagal Dihapus', message);
			showToast(message, false);
		} finally {
			roomDeleteBusyId = '';
		}
	}

	async function saveRoomProctor(room: Room) {
		if (roomSetupLocked(session?.status)) {
			setOperationState('warning', 'Ruangan Terkunci', 'Ruangan terkunci setelah sesi aktif/selesai.');
			return;
		}
		const primaryEmployeeId = roomProctorInput[room.id] ?? '';
		proctorSaveBusyId = room.id;
		try {
			const res = await fetch(`/api/asesmen/sessions/${sessionId}/rooms/${room.id}/proctors`, {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					primary_employee_id: primaryEmployeeId,
					employee_ids: primaryEmployeeId ? [primaryEmployeeId] : [],
				}),
			});
			await readClientJson<unknown>(res);
			setOperationState('success', 'Pengawas Ruangan Diperbarui', `Penugasan pengawas untuk "${room.room_name}" sudah disimpan.`);
			showToast('Pengawas ruangan diperbarui');
			await Promise.all([loadRooms(), loadRoomReadiness()]);
		} catch (error) {
			const message = cbtRoomSetupErrorMessage(error, 'Penugasan pengawas belum berhasil. Akses pengaturan sesi hanya untuk admin/operator CBT.');
			setOperationState('error', 'Pengawas Gagal Disimpan', message);
			showToast(message, false);
		} finally {
			proctorSaveBusyId = '';
		}
	}

	async function shuffleRooms() {
		if (roomSetupLocked(session?.status)) {
			setOperationState('warning', 'Ruangan Terkunci', 'Ruangan terkunci setelah sesi aktif/selesai.');
			return;
		}
		if (!(await confirmPhrase('Acak Peserta ke Ruangan', 'Sistem akan menghapus alokasi ruangan sebelumnya dan membagikan ulang peserta secara otomatis. Pastikan daftar ruangan dan kapasitas sudah final.', 'ACAK'))) return;
		shuffleBusy = true;
		try {
			const res = await fetch(`/api/asesmen/sessions/${sessionId}/shuffle-rooms`, { method: 'POST' });
			await readClientJson<unknown>(res);
			setOperationState('warning', 'Alokasi Ruangan Diperbarui', 'Peserta sudah diacak ulang ke ruangan. Periksa kembali pembagian sebelum ujian dimulai.');
			showToast('Peserta berhasil diacak ke ruangan');
			await loadRooms(); await loadParticipants(); await loadRoomReadiness();
		} catch (error) {
			const message = cbtRoomSetupErrorMessage(error, 'Sistem belum berhasil mengacak peserta ke ruangan. Cek kapasitas ruangan atau ulangi lagi.');
			setOperationState('error', 'Pengacakan Ruangan Gagal', message);
			showToast(message, false);
		} finally {
			shuffleBusy = false;
		}
	}

	async function autoAssignSeats() {
		if (roomSetupLocked(session?.status)) {
			setOperationState('warning', 'Ruangan Terkunci', 'Ruangan terkunci setelah sesi aktif/selesai.');
			return;
		}
		if (!(await confirmPhrase('Atur Nomor Meja Otomatis', 'Nomor meja peserta akan diurutkan ulang per ruangan. Gunakan setelah alokasi ruangan sudah final.', 'MEJA'))) return;
		seatBusy = true;
		try {
			const res = await fetch(`/api/asesmen/sessions/${sessionId}/seats/auto`, { method: 'POST' });
			await readClientJson<unknown>(res);
			setOperationState('success', 'Nomor Meja Diatur Otomatis', 'Nomor meja peserta sudah diperbarui. Lanjutkan ke pengecekan akhir per ruangan jika diperlukan.');
			showToast('Nomor meja berhasil diurutkan otomatis');
			await Promise.all([loadParticipants(), loadRooms(), loadRoomReadiness()]);
		} catch (error) {
			const message = cbtRoomSetupErrorMessage(error, 'Pengaturan otomatis belum berhasil. Pastikan peserta sudah punya ruangan lalu coba lagi.');
			setOperationState('error', 'Nomor Meja Gagal Diatur', message);
			showToast(message, false);
		} finally {
			seatBusy = false;
		}
	}

	async function assignSeat(pid: string) {
		if (roomSetupLocked(session?.status)) {
			setOperationState('warning', 'Ruangan Terkunci', 'Ruangan terkunci setelah sesi aktif/selesai.');
			return;
		}
		const roomId = roomInput[pid];
		const seatNo = seatInput[pid];
		if (!roomId || !seatNo || seatNo <= 0) {
			showToast('Pilih ruangan dan isi nomor meja yang valid', false);
			return;
		}
		seatSaveBusyId = pid;
		try {
			const res = await fetch(`/api/asesmen/sessions/${sessionId}/participants/${pid}/seat`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ room_id: roomId, seat_no: seatNo }),
			});
			await readClientJson<unknown>(res);
			setOperationState('success', 'Nomor Meja Peserta Disimpan', 'Ruangan dan nomor meja peserta sudah diperbarui sesuai pengaturan operator.');
			showToast('No meja peserta diperbarui');
			await Promise.all([loadParticipants(), loadRooms(), loadRoomReadiness()]);
		} catch (error) {
			const message = cbtRoomSetupErrorMessage(error, 'Perubahan ruangan atau nomor meja belum berhasil. Periksa input dan ulangi lagi.');
			setOperationState('error', 'Nomor Meja Gagal Disimpan', message);
			showToast(message, false);
		} finally {
			seatSaveBusyId = '';
		}
	}

	async function flagParticipant(pid: string, flag: boolean) {
		flagBusyId = pid;
		try {
			const res = await fetch(`/api/asesmen/sessions/${sessionId}/participants/${pid}/flag`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ flag }),
			});
			await readClientJson<unknown>(res);
			setOperationState(
				flag ? 'warning' : 'success',
				flag ? 'Peserta Diberi Tanda' : 'Tanda Peserta Dihapus',
				flag
					? 'Peserta ditandai untuk perhatian pengawas. Tinjau kembali aktivitas proctoring sebelum mengambil langkah lanjutan.'
					: 'Tanda kecurigaan pada peserta sudah dibersihkan dari daftar proctoring.',
			);
			await loadProctoring();
		} catch (error) {
			setOperationState('error', 'Tanda Peserta Gagal Diperbarui', 'Perubahan tanda proctoring belum berhasil. Coba ulang beberapa saat lagi.');
			showToast(mutationErrorMessage(error, 'Gagal memperbarui tanda peserta'), false);
		} finally {
			flagBusyId = '';
		}
	}

	async function resetParticipantAccess(pid: string, nama: string) {
		if (!(await confirmPhrase('Reset Akses Ujian', `Akses perangkat untuk ${nama} akan dilepas sehingga peserta dapat login ulang. Gunakan hanya setelah diverifikasi oleh pengawas.`, 'RESET AKSES'))) return;
		resetAccessBusyId = pid;
		try {
			const res = await fetch(clientApiPath`/api/asesmen/sessions/${sessionId}/participants/${pid}/reset-access`, { method: 'POST' });
			await readClientJson<unknown>(res);
			setOperationState('warning', 'Akses Peserta Direset', `${nama} dapat login ulang setelah pengawas memastikan perangkat yang dipakai benar.`);
			toast.success('Akses peserta direset');
			eventPanelParticipantId = pid;
			await Promise.all([loadProctoring(), loadProctoringEvents(pid)]);
		} catch (error) {
			setOperationState('error', 'Reset Akses Gagal', 'Akses peserta belum berhasil direset. Ulangi setelah memeriksa status sesi.');
			toast.error(mutationErrorMessage(error, 'Gagal reset akses peserta'));
		} finally {
			resetAccessBusyId = '';
		}
	}

	async function forceSubmitParticipant(pid: string, nama: string) {
		if (!(await confirmPhrase('Paksa Submit Peserta', `Jawaban ${nama} akan dikunci dan skor objektif dihitung dari jawaban yang sudah tersimpan. Tindakan ini untuk kondisi darurat operasional.`, 'PAKSA SUBMIT'))) return;
		forceSubmitBusyId = pid;
		try {
			const res = await fetch(clientApiPath`/api/asesmen/sessions/${sessionId}/participants/${pid}/force-submit`, { method: 'POST' });
			await readClientJson<unknown>(res);
			setOperationState('warning', 'Peserta Dipaksa Submit', `${nama} sudah ditandai submit oleh proktor. Periksa hasil akhir sebelum menutup sesi.`);
			toast.success('Peserta disubmit oleh proktor');
			eventPanelParticipantId = pid;
			await Promise.all([loadProctoring(), loadProctoringEvents(pid), refreshSessionDetail()]);
		} catch (error) {
			setOperationState('error', 'Paksa Submit Gagal', 'Sistem belum berhasil mengunci submit peserta. Periksa status waktu ujian dan ulangi bila perlu.');
			toast.error(mutationErrorMessage(error, 'Gagal paksa submit peserta'));
		} finally {
			forceSubmitBusyId = '';
		}
	}

	async function submitGrade(aid: string) {
		const score = gradeInput[aid];
		if (score === undefined || score < 0 || score > 100) {
			showToast('Nilai harus 0-100', false);
			return;
		}
		gradeBusyId = aid;
		try {
			const res = await fetch(`/api/asesmen/sessions/${sessionId}/answers/${aid}/grade-essay`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ manual_score: score }),
			});
			await readClientJson<unknown>(res);
			setOperationState('success', 'Nilai Uraian Tersimpan', 'Koreksi uraian sudah masuk dan skor peserta sudah disegarkan.');
			showToast('Nilai berhasil disimpan');
			await Promise.all([loadEssays(), refreshSessionDetail()]);
		} catch (error) {
			setOperationState('error', 'Nilai Uraian Gagal Disimpan', 'Koreksi belum berhasil tersimpan. Ulangi setelah memastikan skor sudah valid.');
			showToast(mutationErrorMessage(error, 'Gagal menyimpan nilai'), false);
		} finally {
			gradeBusyId = '';
		}
	}

	function exportCSV() {
		if (!session || results.length === 0) return;
		const header = csvRow(['NIS', 'Nama', 'L/P', 'Jawaban Masuk', 'Benar', 'Skor', 'Waktu Submit']);
		const rows = results.map(r =>
			csvRow([r.nis, r.nama, r.gender, r.total_answers, r.correct_answers,
			 fmtScore(r.score), r.submitted_at ? fmtDt(r.submitted_at) : ''])
		);
		const csv = [header, ...rows].join('\n');
		const blob = new Blob([csv], { type: 'text/csv' });
		const url = URL.createObjectURL(blob);
		const a = document.createElement('a');
		a.href = url;
		a.download = `hasil_${session.title.replace(/\s+/g, '_')}.csv`;
		a.click();
		URL.revokeObjectURL(url);
	}

	function exportOperationalCSV() {
		if (!session || !operationalRecap) return;
		const header = csvRow(['Ruang', 'Handover', 'Peserta', 'Login', 'Submit', 'No Show', 'Atensi', 'Force Submit', 'Reset Akses', 'App Switch', 'Screenshot', 'Catatan Kejadian', 'Catatan Operator', 'Catatan Serah Terima']);
		const rows = operationalRooms.map((room) => csvRow([
			room.room_name,
			handoverStatusLabel(room),
			room.participant_count,
			room.joined_count,
			room.submitted_count,
			room.no_show_count,
			room.suspicious_count,
			room.force_submit_count,
			room.reset_access_count,
			room.app_switch_count,
			room.screenshot_attempt_count,
			room.incident_notes,
			room.operator_notes,
			room.handover_notes,
		]));
		const summary = csvRow([
			`TOTAL ${operationalRecap.session_title}`,
			`${operationalRecap.handover_locked_count}/${operationalRecap.room_count} terkunci`,
			operationalRecap.participant_count,
			operationalRecap.joined_count,
			operationalRecap.submitted_count,
			operationalRecap.no_show_count,
			operationalRecap.suspicious_count,
			operationalRecap.force_submit_count,
			operationalRecap.reset_access_count,
			operationalRecap.app_switch_count,
			operationalRecap.screenshot_attempt_count,
			`${operationalRecap.incident_room_count} ruang punya catatan`,
			`${operationalRecap.incident_event_count} event atensi`,
			`Generated ${new Date().toISOString()}`,
		]);
		const csv = [header, summary, ...rows].join('\n');
		const blob = new Blob([csv], { type: 'text/csv' });
		const url = URL.createObjectURL(blob);
		const a = document.createElement('a');
		a.href = url;
		a.download = `rekap_operasional_${session.title.replace(/\s+/g, '_')}.csv`;
		a.click();
		URL.revokeObjectURL(url);
	}

	function exportItemAnalysisCSV() {
		if (!session || itemAnalysis.length === 0) return;
		const header = csvRow(['No', 'Kode', 'Tipe', 'CP', 'TP', 'KD', 'Materi', 'Level', 'HOTS', 'Dijawab', 'Kosong', 'Benar', 'Salah', 'Kesukaran', 'Daya Pembeda', 'Rekomendasi']);
		const rows = itemAnalysis.map((row) => csvRow([
			row.position,
			row.question_code,
			questionTypeLabel(row.question_type),
			row.cp_ref,
			row.tp_ref,
			row.kd_ref,
			row.material_topic,
			row.cognitive_level,
			row.hots_flag ? 'Ya' : 'Tidak',
			row.answered_count,
			row.blank_count,
			row.correct_count,
			row.incorrect_count,
			percent(row.difficulty_index),
			percent(row.discrimination_index),
			row.recommendation,
		]));
		const csv = [header, ...rows].join('\n');
		const blob = new Blob([csv], { type: 'text/csv' });
		const url = URL.createObjectURL(blob);
		const a = document.createElement('a');
		a.href = url;
		a.download = `analisis_butir_${session.title.replace(/\s+/g, '_')}.csv`;
		a.click();
		URL.revokeObjectURL(url);
	}

	onMount(() => {
		void loadInitial();
		void loadCommandCenterSnapshot().catch((error) => console.warn('CBT command center snapshot failed', error));
		const requestedTab = tabFromQuery(page.url.searchParams.get('tab'));
		if (requestedTab && requestedTab !== activeTab) {
			void switchTab(requestedTab);
		}
	});
	onDestroy(() => { if (procInterval) clearInterval(procInterval); });
</script>

<svelte:head>
	<title>{session?.title ?? 'Detail Sesi'} — MTSN 2 Kolut</title>
</svelte:head>

<div class="space-y-5 p-6">
	<!-- Breadcrumb -->
	<div class="flex items-center gap-2 text-sm text-slate-500">
		<a href={resolve('/asesmen/sesi')} class="hover:text-slate-700">Kegiatan & Sesi</a>
		<span>/</span>
		<span class="text-slate-700 font-medium truncate max-w-xs">{session?.title ?? '...'}</span>
	</div>

	{#if operationState}
		<OperationStatusPanel {...operationState} />
	{/if}

	<AsyncContent promise={detailPromise} onerror={handleDetailRenderError}>
		{#snippet pending()}
			<div class="space-y-4">
				<div class="space-y-2">
					<Skeleton class="h-4 w-56" />
					<Skeleton class="h-8 w-80" />
					<Skeleton class="h-4 w-96" />
				</div>
				<div class="grid grid-cols-2 gap-3 sm:grid-cols-4">
					{#each Array.from({ length: 4 }) as _, index (`cbt-session-detail-stat-${index}`)}
						<Card.Root class="border-green-100">
							<Card.Content class="space-y-2 px-4 pb-3 pt-4">
								<Skeleton class="h-4 w-24" />
								<Skeleton class="h-8 w-16" />
							</Card.Content>
						</Card.Root>
					{/each}
				</div>
				<Card.Root>
					<Card.Content class="space-y-3 p-6">
						{#each Array.from({ length: 5 }) as _, index (`cbt-session-detail-row-${index}`)}
							<div class="grid gap-3 lg:grid-cols-[1fr_0.8fr_0.8fr_0.8fr_0.6fr_0.6fr_0.7fr_auto] lg:items-center">
								<Skeleton class="h-5 w-24" />
								<Skeleton class="h-5 w-32" />
								<Skeleton class="h-5 w-10" />
								<Skeleton class="h-5 w-12" />
								<Skeleton class="h-5 w-12" />
								<Skeleton class="h-5 w-12" />
								<Skeleton class="h-5 w-28" />
								<Skeleton class="h-9 w-24 justify-self-end" />
							</div>
						{/each}
					</Card.Content>
				</Card.Root>
			</div>
		{/snippet}

		{#snippet failed(error, reset)}
			<RecoveryPanel
				title="Detail Sesi Belum Tersaji"
				message={detailErrorMessage(error)}
				onRetry={() => retryDetail(reset)}
			/>
		{/snippet}

		{#snippet children(value)}
			{@const detail = value as SessionResultsDetail}
			{@const currentSession = detail.session}
			{@const currentResults = detail.results}
			{@const roomControlsLocked = roomSetupLocked(currentSession.status)}
		<!-- Session header -->
		<div class="flex items-start justify-between gap-4 flex-wrap">
			<div>
				<h1 class="text-2xl font-semibold text-[oklch(0.38_0.13_145)]">{currentSession.title}</h1>
				<div class="flex flex-wrap gap-2 mt-2 text-sm text-slate-500">
					<span>{currentSession.package_title}</span>
					{#if currentSession.class_code}<span>· Kelas {currentSession.class_code}</span>{/if}
					<span>· {currentSession.duration_minutes} menit</span>
					<span>· {fmtDt(currentSession.scheduled_start)}</span>
				</div>
			</div>
				<div class="flex items-center gap-2 flex-wrap">
					<Badge class={statusClass(currentSession.status)}>{statusLabel[currentSession.status] ?? currentSession.status}</Badge>
				{#if isAdmin && (currentSession.status === 'finished' || currentSession.status === 'active')}
					<LoadingButton size="sm" variant="outline" disabled={scoreBusy} onclick={triggerScoring} loading={scoreBusy} loadingLabel="Menghitung...">
						⟳ Hitung Skor
					</LoadingButton>
				{/if}
					{#if currentResults.length > 0}
						<Button size="sm" variant="outline" onclick={exportCSV}>↓ CSV</Button>
					{/if}
					<a href={resolve(`/asesmen/sesi/${sessionId}/minutes`)} class="inline-flex items-center rounded-md border border-input bg-background px-3 py-2 text-sm font-medium text-slate-700 hover:bg-muted">
						Berita Acara
					</a>
					{#if currentSession.event_id}
						<a href={resolve(`/asesmen/kegiatan/${currentSession.event_id}/exam-cards`)} class="inline-flex items-center rounded-md border border-input bg-background px-3 py-2 text-sm font-medium text-slate-700 hover:bg-muted">
							Kartu Ujian Event
						</a>
					{/if}
				</div>
		</div>

		<!-- Stats -->
		<div class="grid grid-cols-2 gap-3 sm:grid-cols-4">
			{#each [
				{ label: 'Total Peserta', val: stats.total.toString() },
				{ label: 'Sudah Submit', val: stats.submitted.toString() },
				{ label: 'Rata-rata Nilai', val: stats.total > 0 ? stats.avgScore.toFixed(1) : '—' },
				{ label: 'Lulus (≥75)', val: `${stats.passing} / ${stats.submitted}` },
			] as s (s.label)}
				<Card.Root class="border-green-100">
					<Card.Content class="pt-4 pb-3 px-4">
						<p class="text-xs text-slate-500 mb-1">{s.label}</p>
						<p class="text-2xl font-bold text-[oklch(0.38_0.13_145)]">{s.val}</p>
					</Card.Content>
				</Card.Root>
			{/each}
		</div>

		<Card.Root class={commandCenterClass()}>
			<Card.Header class="flex flex-row items-start justify-between gap-3 pb-2">
				<div>
					<Card.Title class="text-base">Monitoring Sesi</Card.Title>
					<p class="mt-1 text-xs text-slate-600">
						Snapshot ringkas untuk operator: ruang, pengawas, submit, koneksi, atensi, dan handover.
					</p>
				</div>
				<LoadingButton variant="outline" size="sm" loading={commandCenterBusy} loadingLabel="Memuat..." onclick={() => void refreshCommandCenter()}>
					Refresh
				</LoadingButton>
			</Card.Header>
			<Card.Content class="space-y-3">
				<div class="grid gap-2 md:grid-cols-3 xl:grid-cols-6">
					{#each commandCenterMetrics as metric (metric.label)}
						<button
							type="button"
							class="rounded-md border border-white/80 bg-white px-3 py-2 text-left shadow-sm transition hover:border-emerald-200 hover:bg-emerald-50"
							onclick={() => switchTab(metric.tab)}
						>
							<span class="block text-[11px] font-semibold uppercase tracking-[0.14em] text-slate-500">{metric.label}</span>
							<span class="mt-1 block text-xl font-bold text-[oklch(0.38_0.13_145)]">{metric.value}</span>
							<span class="mt-0.5 block text-[11px] text-slate-500">{metric.helper}</span>
						</button>
					{/each}
				</div>
				<div class="flex flex-wrap items-center gap-2">
					{#if commandCenterIssues.length === 0}
						<Badge class="border-emerald-200 bg-white text-emerald-700">Operasional terkendali</Badge>
					{:else}
						{#each commandCenterIssues.slice(0, 6) as issue (`${issue.tab}-${issue.label}`)}
							<button
								type="button"
								class="rounded-md border px-2.5 py-1 text-xs font-semibold transition hover:bg-slate-50 {commandIssueClass(issue.tone)}"
								onclick={() => switchTab(issue.tab)}
							>
								{issue.label}
							</button>
						{/each}
						{#if commandCenterIssues.length > 6}
							<Badge variant="outline" class="bg-white text-xs">+{commandCenterIssues.length - 6} atensi lain</Badge>
						{/if}
					{/if}
				</div>
			</Card.Content>
		</Card.Root>

		{@const detailNextAction = nextDetailAction(roomReadiness)}
		<Card.Root class={detailActionPanelClass(detailNextAction.tone)}>
			<Card.Content class="flex flex-wrap items-center justify-between gap-3 p-4">
				<div>
					<p class="text-xs font-semibold uppercase tracking-[0.16em] text-slate-600">Aksi Berikutnya</p>
					<p class="mt-1 text-sm font-semibold text-slate-900">{detailNextAction.title}</p>
					<p class="mt-0.5 text-xs text-slate-600">{detailNextAction.message}</p>
				</div>
				{#if detailNextAction.run === 'auto_seats'}
					<LoadingButton size="sm" loading={seatBusy} loadingLabel="Mengatur..." disabled={roomControlsLocked || seatBusy} onclick={autoAssignSeats}>
						{detailNextAction.label}
					</LoadingButton>
				{:else if detailNextAction.tab}
					<Button size="sm" onclick={() => switchTab(detailNextAction.tab ?? 'hasil')}>
						{detailNextAction.label}
					</Button>
				{/if}
			</Card.Content>
		</Card.Root>

		<!-- Tabs -->
		<div class="space-y-3 rounded-xl border border-green-100 bg-white p-3 shadow-sm">
			<div class="flex flex-wrap items-center justify-between gap-2">
				<div>
					<p class="text-xs font-semibold uppercase tracking-[0.16em] text-green-700">Alur 5 Modul CBT</p>
					<p class="mt-0.5 text-sm text-slate-600">Detail sesi dikelompokkan ke Kegiatan & Sesi, Monitoring, dan Hasil & Analisis.</p>
				</div>
				<div class="flex flex-wrap gap-2">
					<a href={resolve(`/asesmen/sesi/${sessionId}/minutes`)} class="inline-flex items-center rounded-md border border-input bg-background px-2.5 py-1.5 text-xs font-semibold text-slate-700 hover:bg-muted">
						Berita Acara
					</a>
					{#if currentSession.event_id}
						<a href={resolve(`/asesmen/kegiatan/${currentSession.event_id}/exam-cards`)} class="inline-flex items-center rounded-md border border-input bg-background px-2.5 py-1.5 text-xs font-semibold text-slate-700 hover:bg-muted">
							Kartu Ujian
						</a>
					{/if}
				</div>
			</div>
			<div class="grid gap-3 xl:grid-cols-3" role="tablist" aria-label="Navigasi detail sesi CBT">
				{#each detailTabGroups as group (group.module)}
					<section class="rounded-lg border border-slate-200 bg-slate-50/70 p-2">
						<div class="mb-2 px-1">
							<p class="text-xs font-semibold text-slate-800">{group.module}</p>
							<p class="text-[11px] text-slate-500">{group.help}</p>
						</div>
						<div class="flex flex-wrap gap-1">
							{#each group.tabs as tab (tab.id)}
								<button
									id={`tab-${tab.id}`}
									type="button"
									role="tab"
									aria-selected={activeTab === tab.id}
									aria-controls={`panel-${tab.id}`}
									onclick={() => switchTab(tab.id)}
									class="rounded-md border px-3 py-1.5 text-sm font-medium transition-colors {activeTab === tab.id
										? 'border-[oklch(0.38_0.13_145)] bg-green-50 text-[oklch(0.38_0.13_145)]'
										: 'border-slate-200 bg-white text-slate-600 hover:border-slate-300 hover:text-slate-800'}"
								>
									{tab.label}
								</button>
							{/each}
						</div>
					</section>
				{/each}
			</div>
		</div>

		<!-- Tab: Hasil -->
		{#if activeTab === 'hasil'}
			<Card.Root>
				<Card.Header class="pb-2">
				<Card.Title class="text-base">Daftar Nilai ({results.length} peserta)</Card.Title>
				</Card.Header>
				<Card.Content class="p-0 overflow-x-auto">
					<Table.Root>
						<Table.Header>
							<Table.Row class="bg-green-50">
								<Table.Head class="w-8">#</Table.Head>
								<Table.Head>NIS</Table.Head>
								<Table.Head>Nama</Table.Head>
								<Table.Head>L/P</Table.Head>
								<Table.Head class="text-center">Jawaban</Table.Head>
								<Table.Head class="text-center">Benar</Table.Head>
								<Table.Head class="text-center">Skor</Table.Head>
								<Table.Head>Waktu Submit</Table.Head>
							</Table.Row>
						</Table.Header>
						<Table.Body>
							{#each currentResults as r, i (r.participant_id)}
								<Table.Row>
									<Table.Cell class="text-slate-400 text-xs">{i + 1}</Table.Cell>
									<Table.Cell class="font-mono text-sm">{r.nis}</Table.Cell>
									<Table.Cell class="font-medium">{r.nama}</Table.Cell>
									<Table.Cell><Badge variant="outline" class="text-xs">{r.gender}</Badge></Table.Cell>
									<Table.Cell class="text-center text-sm">{r.total_answers}</Table.Cell>
									<Table.Cell class="text-center text-sm">{r.correct_answers}</Table.Cell>
										<Table.Cell class="text-center"><span class={scoreClass(r.score)}>{fmtScore(r.score)}</span></Table.Cell>
										<Table.Cell class="text-slate-500 text-xs whitespace-nowrap">
											{#if r.submitted_at}
												{fmtDt(r.submitted_at)}
											{:else}
												<span class="text-slate-300">Belum submit</span>
											{/if}
										</Table.Cell>
								</Table.Row>
							{:else}
								<Table.Row>
									<Table.Cell colspan={8} class="text-center text-slate-400 py-10">Belum ada data nilai</Table.Cell>
								</Table.Row>
							{/each}
						</Table.Body>
					</Table.Root>
				</Card.Content>
			</Card.Root>

		<!-- Tab: Analisis Butir -->
		{:else if activeTab === 'butir'}
			<div class="flex flex-wrap items-center justify-between gap-3">
				<div>
					<p class="text-sm font-medium text-slate-700">Analisis kualitas butir dari jawaban siswa</p>
					<p class="text-xs text-muted-foreground">Gunakan setelah skor dihitung untuk melihat kesukaran, daya pembeda, dan opsi jawaban.</p>
				</div>
				<div class="flex flex-wrap gap-2">
					<LoadingButton variant="outline" size="sm" onclick={() => void refreshItemAnalysis()} loading={itemAnalysisRefreshBusy} loadingLabel="Memuat..." disabled={itemAnalysisRefreshBusy}>
						↻ Refresh
					</LoadingButton>
					<Button variant="outline" size="sm" onclick={exportItemAnalysisCSV} disabled={itemAnalysis.length === 0}>
						↓ CSV Analisis
					</Button>
				</div>
			</div>

			<OperationStatusPanel
				tone={itemAnalysisStats.review > 0 ? 'warning' : itemAnalysisStats.total > 0 ? 'success' : 'info'}
				compact
				title="Kontrol Mutu Bank Soal"
				message={`${itemAnalysisStats.total} butir dianalisis, ${itemAnalysisStats.review} butir perlu ditinjau, ${itemAnalysisStats.lowDiscrimination} daya pembeda rendah, rata-rata kesukaran ${percent(itemAnalysisStats.avgDifficulty)}.`}
			/>

			<div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-5">
				{#each [
					{ label: 'Total Butir', value: itemAnalysisStats.total.toString(), hint: 'dalam paket' },
					{ label: 'Perlu Review', value: itemAnalysisStats.review.toString(), hint: 'kunci/rubrik/level' },
					{ label: 'Kesukaran Rata-rata', value: percent(itemAnalysisStats.avgDifficulty), hint: 'proporsi benar/skor' },
					{ label: 'Daya Pembeda Rendah', value: itemAnalysisStats.lowDiscrimination.toString(), hint: '< 15%' },
					{ label: 'Uraian Belum Dinilai', value: itemAnalysisStats.unscored.toString(), hint: 'butir essay' },
				] as item (item.label)}
					<Card.Root class="border-green-100">
						<Card.Content class="px-4 pb-3 pt-4">
							<p class="mb-1 text-xs text-slate-500">{item.label}</p>
							<p class="text-2xl font-bold text-[oklch(0.38_0.13_145)]">{item.value}</p>
							<p class="mt-1 text-xs text-slate-400">{item.hint}</p>
						</Card.Content>
					</Card.Root>
				{/each}
			</div>

			<Card.Root class="border-green-100">
				<Card.Header class="pb-2">
					<Card.Title class="text-base">Matriks Analisis Butir</Card.Title>
					<p class="text-xs text-muted-foreground">Butir bermasalah ditandai agar guru bisa memperbaiki bank soal setelah ujian.</p>
				</Card.Header>
				<Card.Content class="p-0 overflow-x-auto">
					<Table.Root>
						<Table.Header>
							<Table.Row class="bg-green-50">
								<Table.Head class="w-10">No</Table.Head>
								<Table.Head>Butir</Table.Head>
								<Table.Head>Blueprint</Table.Head>
								<Table.Head class="text-center">Dijawab</Table.Head>
								<Table.Head class="text-center">Benar / Salah</Table.Head>
								<Table.Head class="text-center">Kesukaran</Table.Head>
								<Table.Head class="text-center">Daya Pembeda</Table.Head>
								<Table.Head>Distribusi</Table.Head>
								<Table.Head>Rekomendasi</Table.Head>
							</Table.Row>
						</Table.Header>
						<Table.Body>
							{#each itemAnalysis as row (row.question_id)}
								<Table.Row class={itemAnalysisSignalClass(row)}>
									<Table.Cell class="text-xs text-slate-400">{row.position}</Table.Cell>
									<Table.Cell class="min-w-72">
										<div class="flex flex-wrap items-center gap-1">
											{#if row.question_code}<span class="font-mono text-xs text-slate-500">{row.question_code}</span>{/if}
											<Badge variant="outline" class="text-xs">{questionTypeLabel(row.question_type)}</Badge>
											<Badge variant="secondary" class="text-xs">{row.difficulty}</Badge>
											{#if row.hots_flag}<Badge class="border-amber-200 bg-amber-50 text-amber-700 text-xs">HOTS</Badge>{/if}
										</div>
										<p class="mt-1 line-clamp-2 text-sm text-slate-700">{row.question_text}</p>
										<p class="mt-1 text-[11px] text-slate-400">Kunci: {row.question_type === 'essay' ? 'Rubrik' : row.answer_key || '—'} · Bobot {row.points}</p>
									</Table.Cell>
									<Table.Cell class="min-w-56 text-xs text-slate-600">
										<div>CP: {row.cp_ref || '—'}</div>
										<div>TP/KD: {row.tp_ref || row.kd_ref || '—'}</div>
										<div>Materi: {row.material_topic || '—'} · {row.cognitive_level || '—'}</div>
									</Table.Cell>
									<Table.Cell class="text-center text-sm">
										{row.answered_count}/{row.submitted_count}
										{#if row.blank_count > 0}<div class="text-[11px] text-amber-700">{row.blank_count} kosong</div>{/if}
									</Table.Cell>
									<Table.Cell class="text-center text-sm">
										{#if row.question_type === 'essay'}
											{row.avg_manual_score.toFixed(1)} rata-rata
											{#if row.unscored_count > 0}<div class="text-[11px] text-red-700">{row.unscored_count} belum dinilai</div>{/if}
										{:else}
											<span class="font-mono">{row.correct_count}/{row.incorrect_count}</span>
										{/if}
									</Table.Cell>
									<Table.Cell class="text-center font-semibold text-slate-700">{percent(row.difficulty_index)}</Table.Cell>
									<Table.Cell class="text-center">
										<span class={row.discrimination_index < 0 ? 'font-semibold text-red-700' : row.discrimination_index < 0.15 ? 'font-semibold text-amber-700' : 'font-semibold text-emerald-700'}>
											{percent(row.discrimination_index)}
										</span>
										{#if row.top_group_count > 0 || row.bottom_group_count > 0}
											<div class="text-[11px] text-slate-400">atas {row.top_correct_count}/{row.top_group_count} · bawah {row.bottom_correct_count}/{row.bottom_group_count}</div>
										{/if}
									</Table.Cell>
									<Table.Cell class="min-w-40">
										<div class="flex flex-wrap gap-1">
											{#each answerDistributionEntries(row).slice(0, 5) as entry (entry.answer)}
												<Badge variant="outline" class="bg-white text-xs">{entry.answer}: {entry.count}</Badge>
											{:else}
												<span class="text-xs text-slate-400">Belum ada jawaban</span>
											{/each}
										</div>
									</Table.Cell>
									<Table.Cell>
										<div class="flex min-w-44 flex-col items-start gap-2">
											<Badge variant="outline" class={itemAnalysisToneClass(row.recommendation_tone)}>{row.recommendation}</Badge>
											{#if shouldOfferQuestionRevision(row)}
												<LoadingButton
													variant="outline"
													size="sm"
													onclick={() => void markQuestionRevision(row)}
													loading={revisionBusyId === row.question_id}
													loadingLabel="Membuat..."
													disabled={revisionBusyId !== '' && revisionBusyId !== row.question_id}
												>
													Buat Draft Revisi
												</LoadingButton>
											{/if}
										</div>
									</Table.Cell>
								</Table.Row>
							{:else}
								<Table.Row>
									<Table.Cell colspan={9} class="py-10 text-center text-slate-400">Analisis butir belum tersedia. Pastikan paket memiliki soal dan klik refresh setelah skor dihitung.</Table.Cell>
								</Table.Row>
							{/each}
						</Table.Body>
					</Table.Root>
				</Card.Content>
			</Card.Root>

		<!-- Tab: Peserta & Token -->
		{:else if activeTab === 'peserta'}
			<div id="panel-peserta" role="tabpanel" aria-labelledby="tab-peserta" class="space-y-4">
			<div class="flex gap-2 flex-wrap">
				<LoadingButton variant="outline" size="sm" onclick={() => void generateTokens()} loading={tokenBusy} disabled={tokenBusy} loadingLabel="Membuat token...">Buat Token Massal</LoadingButton>
				<LoadingButton variant="outline" size="sm" onclick={() => void refreshParticipants()} loading={participantRefreshBusy} loadingLabel="Memuat..." disabled={participantRefreshBusy}>↻ Refresh</LoadingButton>
			</div>
			<OperationStatusPanel
				tone="warning"
				compact
				title="Aksi Sensitif Peserta & Token Rahasia"
				message={roomControlsLocked ? 'Ruangan terkunci setelah sesi aktif/selesai. Token peserta tetap rahasia dan hanya boleh dibagikan ke pengawas atau peserta yang berwenang saat operasional ujian.' : 'Pembuatan token massal, ubah token, dan simpan nomor meja akan langsung mengubah data operasional ujian. Perlakukan token seperti kredensial ujian: jangan kirim ke kanal umum, jangan tampilkan di layar proyektor, dan bagikan hanya saat sesi siap.'}
			/>
			<Card.Root>
				<Card.Content class="p-0 overflow-x-auto">
					<Table.Root>
						<Table.Header>
							<Table.Row class="bg-green-50">
								<Table.Head>NIS</Table.Head>
								<Table.Head>Nama</Table.Head>
								<Table.Head>L/P</Table.Head>
								<Table.Head>Ruangan</Table.Head>
								<Table.Head>No Meja</Table.Head>
								<Table.Head>Token Rahasia</Table.Head>
								<Table.Head>Status</Table.Head>
								<Table.Head class="text-right">Aksi</Table.Head>
							</Table.Row>
						</Table.Header>
						<Table.Body>
								{#each participants as p (p.id)}
									<Table.Row class={p.suspicious_flag ? 'bg-red-50' : 'hover:bg-green-50/30'}>
									<Table.Cell class="font-mono text-sm">{p.nis}</Table.Cell>
									<Table.Cell class="font-medium">
										{p.nama}
										{#if p.suspicious_flag}<span class="ml-1 text-red-500 text-xs">⚑ Dicurigai</span>{/if}
									</Table.Cell>
										<Table.Cell><Badge variant="outline" class="text-xs">{p.gender}</Badge></Table.Cell>
										<Table.Cell class="text-sm text-muted-foreground">{p.room_name || '—'}</Table.Cell>
										<Table.Cell class="text-sm text-muted-foreground">{p.seat_no ?? '—'}</Table.Cell>
										<Table.Cell>
										{#if p.token}
											<Badge variant="outline" class="border-green-200 bg-green-50 text-xs text-green-800">Rahasia - siap kartu</Badge>
										{:else}
											<span class="text-slate-400 text-xs">—</span>
										{/if}
									</Table.Cell>
									<Table.Cell>
										{#if p.submitted_at}
											<Badge variant="outline" class="text-xs bg-slate-100 text-slate-500">Submit</Badge>
										{:else}
											<Badge variant="outline" class="text-xs bg-amber-50 text-amber-700 border-amber-200">Belum</Badge>
										{/if}
									</Table.Cell>
										<Table.Cell class="text-right">
											<div class="flex items-center justify-end gap-1">
										<select bind:value={roomInput[p.id]} aria-label={`Pilih ruangan untuk ${p.nama}`} class="h-8 rounded-md border border-input bg-background px-2 text-xs" disabled={roomControlsLocked}>
													<option value="">Ruangan</option>
													{#each rooms as room (room.id)}
														<option value={room.id}>{room.room_name}</option>
													{/each}
												</select>
										<Input bind:value={seatInput[p.id]} type="number" min="1" aria-label={`Nomor meja untuk ${p.nama}`} class="h-8 w-16" disabled={roomControlsLocked} />
										<LoadingButton variant="outline" size="sm" onclick={() => assignSeat(p.id)} loading={seatSaveBusyId === p.id} disabled={roomControlsLocked || (seatSaveBusyId !== '' && seatSaveBusyId !== p.id) || seatBusy} loadingLabel="Menyimpan...">Simpan</LoadingButton>
											<LoadingButton variant="outline" size="sm" onclick={() => regenerateToken(p.id)} loading={regenBusyId === p.id} disabled={regenBusyId !== '' && regenBusyId !== p.id} loadingLabel="Membuat ulang...">Buat Ulang</LoadingButton>
										</div>
									</Table.Cell>
								</Table.Row>
								{:else}
									<Table.Row>
										<Table.Cell colspan={8} class="text-center text-slate-400 py-8">Belum ada peserta</Table.Cell>
									</Table.Row>
								{/each}
						</Table.Body>
					</Table.Root>
				</Card.Content>
			</Card.Root>
			</div>

		<!-- Tab: Ruangan -->
		{:else if activeTab === 'ruangan'}
			<div id="panel-ruangan" role="tabpanel" aria-labelledby="tab-ruangan" class="space-y-4">
			{#if roomControlsLocked}
				<OperationStatusPanel
					tone="warning"
					compact
					title="Ruangan Terkunci"
					message="Ruangan terkunci setelah sesi aktif/selesai. Gunakan tab proctoring dan rekap operasional untuk pemantauan tanpa mengubah setup ruang."
				/>
			{/if}
			<Card.Root class="border-green-200">
				<Card.Header class="pb-3">
					<Card.Title class="text-base">Ruangan & Pengawas</Card.Title>
					<p class="text-xs text-muted-foreground">Pilih master ruangan fisik bila sudah tersedia, atau isi manual untuk transisi.</p>
				</Card.Header>
				<Card.Content>
					<div class="flex gap-3 flex-wrap items-end {roomControlsLocked ? 'opacity-70' : ''}">
						<div>
							<label for="r-master" class="block text-sm font-medium mb-1">Master Ruangan</label>
							<select
								id="r-master"
								value={selectedSchoolRoomId}
								class="h-10 w-64 rounded-md border border-input bg-background px-3 text-sm"
								disabled={roomControlsLocked}
								onchange={(event) => selectSchoolRoom((event.currentTarget as HTMLSelectElement).value)}
							>
								<option value="">Manual / belum terhubung aset</option>
								{#each schoolRooms as room (room.id)}
									<option value={room.id}>{room.code} — {room.name} ({room.exam_capacity} kursi)</option>
								{/each}
							</select>
						</div>
						<div>
							<label for="r-name" class="block text-sm font-medium mb-1">Label Ruang Ujian</label>
							<Input id="r-name" value={newRoomName} oninput={(event) => updateNewRoomName((event.currentTarget as HTMLInputElement).value)} placeholder="Ruang 1 / Lab Komputer A" class="w-56" disabled={roomControlsLocked} />
							{#if selectedSchoolRoomId && roomNameCustomized}
								<p class="mt-1 text-[11px] text-amber-700">Label sudah dikustom manual dan tidak mengikuti master ruangan.</p>
							{/if}
						</div>
						<div>
							<label for="r-cap" class="block text-sm font-medium mb-1">Kapasitas</label>
							<Input id="r-cap" type="number" bind:value={newRoomCap} min={1} max={100} class="w-24" disabled={roomControlsLocked} />
						</div>
						<LoadingButton onclick={() => void createRoom()} loading={roomBusy} loadingLabel="Menyimpan..." disabled={roomControlsLocked || roomBusy || (!newRoomName.trim() && !selectedSchoolRoomId)}>
							+ Tambah Ruangan
						</LoadingButton>
							{#if rooms.length > 0}
								<LoadingButton variant="outline" loading={shuffleBusy} loadingLabel="Mengacak..." disabled={roomControlsLocked || shuffleBusy} onclick={shuffleRooms}
									class="border-amber-300 text-amber-700 hover:bg-amber-50">
									Acak Peserta
								</LoadingButton>
								<LoadingButton variant="outline" loading={seatBusy} loadingLabel="Mengatur..." disabled={roomControlsLocked || seatBusy} onclick={autoAssignSeats}>
									Atur Nomor Meja
								</LoadingButton>
							{/if}
					</div>
				</Card.Content>
			</Card.Root>
			<OperationStatusPanel
				tone={roomReadinessTone(roomReadiness)}
				compact
				title="Kesiapan Ruangan"
				message={roomReadinessMessage(roomReadiness)}
			/>
			<OperationStatusPanel
				tone="warning"
				compact
				title="Aksi Sensitif Ruangan"
				message="Pengacakan ruangan dan pengaturan nomor meja otomatis akan menimpa penempatan sebelumnya. Gunakan setelah kapasitas, master ruangan, dan pengawas sudah final."
			/>

			{#if rooms.length > 0}
				<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 mt-4">
					{#each rooms as room (room.id)}
						<Card.Root class="border-green-100">
							<Card.Content class="p-4">
								<div class="space-y-3">
									<div class="flex items-start justify-between gap-3">
										<div class="min-w-0">
											<div class="font-semibold text-[oklch(0.38_0.13_145)]">{room.room_name}</div>
											<div class="text-xs text-muted-foreground mt-1">
												{room.school_room_name ? `${room.school_room_code} · ${room.school_room_name}` : 'Belum terhubung master ruangan'}
											</div>
											<div class="text-sm text-muted-foreground mt-1">
												Kapasitas {room.capacity} · Terisi {room.participant_count} · Pengawas {room.proctor_count ?? 0}
											</div>
											<div class="mt-2 h-2 rounded-full bg-green-100 overflow-hidden">
												<div class="h-full bg-[oklch(0.38_0.13_145)] rounded-full transition-all"
													style="width: {roomCapacityRatio(room)}%">
												</div>
											</div>
										</div>
										<div class="flex shrink-0 flex-col gap-2">
											<Button variant="outline" size="sm" href={resolve(`/asesmen/sesi/${sessionId}/rooms/${room.id}/proctoring`)}>
												Dashboard
											</Button>
											<Button variant="outline" size="sm" href={resolve(`/asesmen/sesi/${sessionId}/rooms/${room.id}/print-pack`)}>
												Cetak
											</Button>
										<LoadingButton variant="outline" size="sm"
											class="border-red-200 text-red-600 hover:bg-red-50"
											onclick={() => deleteRoom(room.id, room.room_name)}
											loading={roomDeleteBusyId === room.id}
											disabled={roomControlsLocked || (roomDeleteBusyId !== '' && roomDeleteBusyId !== room.id)}
											loadingLabel="Menghapus...">Hapus</LoadingButton>
										</div>
									</div>
									<div class="rounded-md border border-slate-200 bg-slate-50 p-2">
										<label for={`proctor-${room.id}`} class="mb-1 block text-[11px] font-semibold uppercase tracking-wide text-slate-500">Pengawas utama</label>
										<div class="flex gap-2">
											<select id={`proctor-${room.id}`} bind:value={roomProctorInput[room.id]} class="min-w-0 flex-1 rounded-md border border-input bg-white px-2 py-1.5 text-xs" disabled={roomControlsLocked}>
												<option value="">Belum ditugaskan</option>
												{#each employeeOptions as employee (employee.id)}
													<option value={employee.id}>{employee.nama}{employee.nip ? ` · ${employee.nip}` : ''}</option>
												{/each}
											</select>
											<LoadingButton
												variant="outline"
												size="sm"
												loading={proctorSaveBusyId === room.id}
												loadingLabel="..."
											disabled={roomControlsLocked || (proctorSaveBusyId !== '' && proctorSaveBusyId !== room.id)}
												onclick={() => void saveRoomProctor(room)}
											>
												Simpan
											</LoadingButton>
										</div>
										{#if room.primary_proctor_name}
											<p class="mt-1 text-[11px] text-slate-500">Aktif: {room.primary_proctor_name}</p>
										{/if}
									</div>
								</div>
							</Card.Content>
						</Card.Root>
					{/each}
				</div>

				<!-- Participants by room -->
				<Card.Root class="mt-4">
					<Card.Header class="pb-2">
						<Card.Title class="text-base">Peserta per Ruangan</Card.Title>
					</Card.Header>
					<Card.Content class="p-0 overflow-x-auto">
						<Table.Root>
							<Table.Header>
								<Table.Row class="bg-green-50">
									<Table.Head>NIS</Table.Head>
									<Table.Head>Nama</Table.Head>
										<Table.Head>L/P</Table.Head>
										<Table.Head>Ruangan</Table.Head>
										<Table.Head>No Meja</Table.Head>
										<Table.Head>Token Rahasia</Table.Head>
								</Table.Row>
							</Table.Header>
							<Table.Body>
								{#each participants as p (p.id)}
									<Table.Row>
										<Table.Cell class="font-mono text-sm">{p.nis}</Table.Cell>
										<Table.Cell class="font-medium">{p.nama}</Table.Cell>
										<Table.Cell><Badge variant="outline" class="text-xs">{p.gender}</Badge></Table.Cell>
										{#if p.room_name}
											<Table.Cell class="text-sm">{p.room_name}</Table.Cell>
											{:else}
												<Table.Cell class="text-sm text-slate-400">Belum ditentukan</Table.Cell>
											{/if}
											<Table.Cell class="text-sm text-slate-600">{p.seat_no ?? '—'}</Table.Cell>
											<Table.Cell>
											{#if p.token}
												<Badge variant="outline" class="border-green-200 bg-green-50 text-xs text-green-800">Rahasia - siap kartu</Badge>
											{:else}
												<span class="text-slate-400 text-xs">—</span>
											{/if}
										</Table.Cell>
									</Table.Row>
								{/each}
							</Table.Body>
						</Table.Root>
					</Card.Content>
				</Card.Root>
			{:else}
				<div class="rounded-lg border border-dashed border-green-200 p-8 text-center text-muted-foreground text-sm mt-4">
					Belum ada ruangan. Tambah ruangan di atas, lalu klik "Acak Peserta ke Ruangan".
				</div>
			{/if}
			</div>

		<!-- Tab: Rekap Operasional -->
		{:else if activeTab === 'operasional'}
			<div class="flex items-center justify-between gap-3 flex-wrap">
				<div>
					<p class="text-sm font-medium text-slate-700">Rekap handover, insiden, dan rekonsiliasi sesi</p>
					<p class="text-xs text-muted-foreground">Dipakai operator/panitia untuk memastikan semua ruang siap diarsipkan.</p>
				</div>
				<div class="flex flex-wrap gap-2">
					<LoadingButton variant="outline" size="sm" onclick={() => void refreshOperationalRecap()} loading={operationalRefreshBusy} loadingLabel="Memuat..." disabled={operationalRefreshBusy}>
						↻ Refresh
					</LoadingButton>
					<Button variant="outline" size="sm" onclick={exportOperationalCSV} disabled={!operationalRecap || operationalRooms.length === 0}>
						↓ CSV Rekap
					</Button>
				</div>
			</div>

			<OperationStatusPanel
				tone={operationalTone(operationalRecap)}
				compact
				title="Status Penutupan CBT"
				message={operationalMessage(operationalRecap)}
			/>

			{#if operationalRecap}
				<div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-5">
					{#each [
						{ label: 'Handover Terkunci', value: `${operationalRecap.handover_locked_count}/${operationalRecap.room_count}`, hint: `${operationalRecap.handover_missing_count} belum ada` },
						{ label: 'Submit Akhir', value: `${operationalRecap.submitted_count}/${operationalRecap.participant_count}`, hint: `${operationalRecap.no_show_count} belum login/no-show` },
						{ label: 'Ruang Berinsiden', value: operationalRecap.incident_room_count.toString(), hint: `${operationalRecap.incident_event_count} event atensi` },
						{ label: 'Paksa Submit', value: operationalRecap.force_submit_count.toString(), hint: `${operationalRecap.reset_access_count} reset akses` },
						{ label: 'Anti-Cheat', value: `${operationalRecap.app_switch_count}/${operationalRecap.screenshot_attempt_count}`, hint: 'app switch / screenshot' },
					] as item (item.label)}
						<Card.Root class="border-green-100">
							<Card.Content class="px-4 pb-3 pt-4">
								<p class="mb-1 text-xs text-slate-500">{item.label}</p>
								<p class="text-2xl font-bold text-[oklch(0.38_0.13_145)]">{item.value}</p>
								<p class="mt-1 text-xs text-slate-400">{item.hint}</p>
							</Card.Content>
						</Card.Root>
					{/each}
				</div>

				<Card.Root class="border-green-100">
					<Card.Header class="pb-2">
						<Card.Title class="text-base">Rekap Ruang & Register Insiden</Card.Title>
						<p class="text-xs text-muted-foreground">Ruang yang belum punya handover atau belum terkunci ditampilkan di atas.</p>
					</Card.Header>
					<Card.Content class="p-0 overflow-x-auto">
						<Table.Root>
							<Table.Header>
								<Table.Row class="bg-green-50">
									<Table.Head>Ruang</Table.Head>
									<Table.Head>Handover</Table.Head>
									<Table.Head class="text-center">Submit</Table.Head>
									<Table.Head class="text-center">Atensi</Table.Head>
									<Table.Head class="text-center">Force/Reset</Table.Head>
									<Table.Head>Catatan Kejadian</Table.Head>
									<Table.Head>Catatan Operator</Table.Head>
									<Table.Head class="text-right">Aksi</Table.Head>
								</Table.Row>
							</Table.Header>
							<Table.Body>
								{#each operationalRooms as room (room.room_id)}
									<Table.Row class={room.locked_at ? '' : 'bg-amber-50/50'}>
										<Table.Cell>
											<div class="font-medium text-slate-900">{room.room_name}</div>
											<div class="text-xs text-slate-500">Token rahasia {room.room_token || '—'} · {room.joined_count}/{room.participant_count} login</div>
										</Table.Cell>
										<Table.Cell>
											<Badge variant="outline" class={handoverStatusClass(room)}>{handoverStatusLabel(room)}</Badge>
											<p class="mt-1 text-[11px] text-slate-400">{room.locked_at ? fmtDt(room.locked_at) : room.handover_updated_at ? `Draft ${fmtDt(room.handover_updated_at)}` : 'Belum diisi'}</p>
										</Table.Cell>
										<Table.Cell class="text-center font-mono text-sm">
											{room.submitted_count}/{room.participant_count}
											{#if room.no_show_count > 0}<div class="text-[11px] text-amber-700">{room.no_show_count} no-show</div>{/if}
										</Table.Cell>
										<Table.Cell class="text-center">
											<div class={room.suspicious_count > 0 || room.incident_event_count > 0 ? 'font-semibold text-red-700' : 'text-slate-500'}>
												{room.suspicious_count} / {room.incident_event_count}
											</div>
											<div class="text-[11px] text-slate-400">flag / event</div>
										</Table.Cell>
										<Table.Cell class="text-center font-mono text-sm">{room.force_submit_count}/{room.reset_access_count}</Table.Cell>
										<Table.Cell class="max-w-sm text-xs text-slate-600">
											{room.incident_notes || '—'}
										</Table.Cell>
										<Table.Cell class="max-w-sm text-xs text-slate-600">
											{room.operator_notes || room.handover_notes || '—'}
										</Table.Cell>
										<Table.Cell class="text-right">
											<div class="flex flex-wrap justify-end gap-2">
												<Button variant="outline" size="sm" href={resolve(`/asesmen/sesi/${sessionId}/rooms/${room.room_id}/proctoring`)}>
													Dashboard
												</Button>
												<Button variant="outline" size="sm" href={resolve(`/asesmen/sesi/${sessionId}/rooms/${room.room_id}/print-pack`)}>
													Cetak
												</Button>
											</div>
										</Table.Cell>
									</Table.Row>
								{:else}
									<Table.Row>
										<Table.Cell colspan={8} class="py-10 text-center text-slate-400">Belum ada ruang untuk direkap</Table.Cell>
									</Table.Row>
								{/each}
							</Table.Body>
						</Table.Root>
					</Card.Content>
				</Card.Root>
			{:else}
				<Card.Root>
					<Card.Content class="grid gap-3 p-4 sm:grid-cols-3">
						<Skeleton class="h-24 rounded-lg" />
						<Skeleton class="h-24 rounded-lg" />
						<Skeleton class="h-24 rounded-lg" />
					</Card.Content>
				</Card.Root>
			{/if}

		<!-- Tab: Proctoring -->
		{:else if activeTab === 'proctoring'}
			{@const selectedEventParticipant = proctoring.find((row) => row.participant_id === eventPanelParticipantId)}
			<div class="flex items-center justify-between gap-3 mb-4 flex-wrap">
				<div>
					<p class="text-sm font-medium text-slate-700">Monitoring proctoring live</p>
					<p class="text-xs text-muted-foreground">Pembaruan otomatis setiap 15 detik</p>
				</div>
				<LoadingButton variant="outline" size="sm" onclick={() => void refreshProctoring()} loading={proctoringRefreshBusy} loadingLabel="Memuat..." disabled={proctoringRefreshBusy}>↻ Refresh Sekarang</LoadingButton>
			</div>
			<div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
				{#each [
					{ label: 'Online Aktif', value: proctoringStats.online.toString(), className: 'text-emerald-700' },
					{ label: 'Lambat / Offline', value: `${proctoringStats.slow + proctoringStats.offline}`, className: 'text-amber-700' },
					{ label: 'Perlu Atensi', value: proctoringStats.suspicious.toString(), className: 'text-red-700' },
					{ label: 'App Switch / Screenshot', value: `${proctoringStats.appSwitches} / ${proctoringStats.screenshots}`, className: 'text-slate-700' },
				] as item (item.label)}
					<Card.Root class="border-green-100">
						<Card.Content class="px-4 pb-3 pt-4">
							<p class="mb-1 text-xs text-slate-500">{item.label}</p>
							<p class={`text-2xl font-bold ${item.className}`}>{item.value}</p>
						</Card.Content>
					</Card.Root>
				{/each}
			</div>
			<Card.Root>
				<Card.Content class="p-0 overflow-x-auto">
					<Table.Root>
						<Table.Header>
							<Table.Row class="bg-green-50">
								<Table.Head>Nama</Table.Head>
								<Table.Head>Ruangan</Table.Head>
								<Table.Head>Status</Table.Head>
								<Table.Head class="text-center">Dijawab</Table.Head>
								<Table.Head class="text-center">App Switch</Table.Head>
								<Table.Head class="text-center">Screenshot</Table.Head>
								<Table.Head>Submit</Table.Head>
								<Table.Head class="text-right">Aksi Proktor</Table.Head>
							</Table.Row>
						</Table.Header>
						<Table.Body>
								{#each proctoring as p (p.participant_id)}
									{@const hb = heartbeatStatus(p.last_heartbeat)}
									<Table.Row class={p.suspicious_flag ? 'bg-red-50' : p.app_switch_count >= 3 ? 'bg-amber-50/50' : 'hover:bg-green-50/30'}>
									<Table.Cell class="font-medium">
										{p.nama}
										<div class="text-xs text-muted-foreground font-mono">{p.nis}</div>
									</Table.Cell>
									<Table.Cell class="text-sm text-muted-foreground">{p.room_name || '—'}</Table.Cell>
									<Table.Cell>
										<span class="text-sm font-medium {hb.cls}">{p.submitted_at ? '✅ Submit' : hb.label}</span>
									</Table.Cell>
									<Table.Cell class="text-center font-mono text-sm">{p.answered_count}</Table.Cell>
									<Table.Cell class="text-center">
										<span class="font-mono text-sm {p.app_switch_count >= 3 ? 'text-red-600 font-bold' : 'text-slate-600'}">
											{p.app_switch_count}x
										</span>
									</Table.Cell>
									<Table.Cell class="text-center">
										<span class="font-mono text-sm {p.screenshot_attempt > 0 ? 'text-amber-600 font-semibold' : 'text-slate-400'}">
											{p.screenshot_attempt}x
										</span>
									</Table.Cell>
									<Table.Cell class="text-xs text-muted-foreground whitespace-nowrap">
										{p.submitted_at ? fmtDt(p.submitted_at) : '—'}
									</Table.Cell>
									<Table.Cell class="text-right">
										<div class="flex flex-wrap items-center justify-end gap-1">
											<Button
												variant="outline" size="sm"
												class={eventPanelParticipantId === p.participant_id ? 'border-green-500 text-green-800 bg-green-50' : 'border-slate-200 text-slate-600'}
												onclick={() => showParticipantEvents(p.participant_id)}
												disabled={eventRefreshBusy}>
												Log
											</Button>
											<LoadingButton
												variant="outline" size="sm"
												onclick={() => resetParticipantAccess(p.participant_id, p.nama)}
												loading={resetAccessBusyId === p.participant_id}
												disabled={resetAccessBusyId !== '' && resetAccessBusyId !== p.participant_id}
												loadingLabel="Reset...">
												Reset Akses
											</LoadingButton>
											<LoadingButton
												variant="outline" size="sm"
												class="border-amber-300 text-amber-700 hover:bg-amber-50"
												onclick={() => forceSubmitParticipant(p.participant_id, p.nama)}
												loading={forceSubmitBusyId === p.participant_id}
												disabled={!!p.submitted_at || (forceSubmitBusyId !== '' && forceSubmitBusyId !== p.participant_id)}
												loadingLabel="Submit...">
												Paksa Submit
											</LoadingButton>
											<Button
												variant="outline" size="sm"
												class={p.suspicious_flag ? 'border-red-400 text-red-700 bg-red-50' : 'border-slate-200 text-slate-500'}
												onclick={() => flagParticipant(p.participant_id, !p.suspicious_flag)}
												disabled={flagBusyId === p.participant_id}>
											{p.suspicious_flag ? '⚑ Hapus Tanda' : '⚐ Tandai'}
										</Button>
										</div>
									</Table.Cell>
								</Table.Row>
							{:else}
								<Table.Row>
									<Table.Cell colspan={8} class="text-center text-slate-400 py-8">Belum ada data proctoring</Table.Cell>
								</Table.Row>
							{/each}
						</Table.Body>
					</Table.Root>
				</Card.Content>
			</Card.Root>
			<Card.Root class="border-green-100">
				<Card.Header class="pb-3">
					<div class="flex items-start justify-between gap-3 flex-wrap">
						<div>
							<Card.Title class="text-base">
								Log Aktivitas {selectedEventParticipant ? selectedEventParticipant.nama : 'Semua Peserta'}
							</Card.Title>
							<p class="text-xs text-muted-foreground mt-1">
								{selectedEventParticipant
									? `${selectedEventParticipant.nis} · ${selectedEventParticipant.room_name || 'Tanpa ruangan'}`
									: '100 aktivitas terbaru dari sesi ini'}
							</p>
						</div>
						<div class="flex gap-2">
							{#if eventPanelParticipantId}
								<Button
									variant="outline"
									size="sm"
									onclick={() => {
										eventPanelParticipantId = '';
										void refreshProctoringEvents();
									}}
								>
									Semua Log
								</Button>
							{/if}
							<LoadingButton variant="outline" size="sm" onclick={() => void refreshProctoringEvents()} loading={eventRefreshBusy} loadingLabel="Memuat..." disabled={eventRefreshBusy}>
								↻ Refresh Log
							</LoadingButton>
						</div>
					</div>
				</Card.Header>
				<Card.Content class="p-0 overflow-x-auto">
					<Table.Root>
						<Table.Header>
							<Table.Row class="bg-slate-50">
								<Table.Head>Waktu</Table.Head>
								<Table.Head>Peserta</Table.Head>
								<Table.Head>Aktivitas</Table.Head>
								<Table.Head>Ruangan</Table.Head>
								<Table.Head>Data</Table.Head>
							</Table.Row>
						</Table.Header>
						<Table.Body>
							{#each proctoringEvents as ev (ev.id)}
								<Table.Row>
									<Table.Cell class="whitespace-nowrap text-xs text-muted-foreground">{fmtDt(ev.created_at)}</Table.Cell>
									<Table.Cell>
										<div class="font-medium text-sm">{ev.nama}</div>
										<div class="font-mono text-xs text-muted-foreground">{ev.nis}</div>
									</Table.Cell>
									<Table.Cell>
										<Badge variant="outline" class="text-xs">{proctoringEventLabel(ev.event_type)}</Badge>
									</Table.Cell>
									<Table.Cell class="text-sm text-muted-foreground">{ev.room_name || '—'}</Table.Cell>
									<Table.Cell class="max-w-md break-words font-mono text-xs text-slate-500">{eventDataText(ev.event_data)}</Table.Cell>
								</Table.Row>
							{:else}
								<Table.Row>
									<Table.Cell colspan={5} class="py-8 text-center text-sm text-slate-400">Belum ada log aktivitas</Table.Cell>
								</Table.Row>
							{/each}
						</Table.Body>
					</Table.Root>
				</Card.Content>
			</Card.Root>

		<!-- Tab: Audit Ops -->
		{:else if activeTab === 'audit'}
			<div class="flex flex-wrap items-center justify-between gap-3">
				<div>
					<p class="text-sm font-medium text-slate-700">Jejak audit operasional sesi</p>
					<p class="text-xs text-muted-foreground">Mencatat perubahan jadwal, koreksi, proctoring, dan handover yang punya dampak operasional.</p>
				</div>
				<LoadingButton variant="outline" size="sm" onclick={() => void refreshAuditLogs()} loading={auditRefreshBusy} loadingLabel="Memuat..." disabled={auditRefreshBusy}>
					↻ Refresh Audit
				</LoadingButton>
			</div>

			<Card.Root class="border-green-100">
				<Card.Content class="p-0 overflow-x-auto">
					<Table.Root>
						<Table.Header>
							<Table.Row class="bg-green-50">
								<Table.Head>Waktu</Table.Head>
								<Table.Head>Aktor</Table.Head>
								<Table.Head>Aksi</Table.Head>
								<Table.Head>Ringkasan</Table.Head>
							</Table.Row>
						</Table.Header>
						<Table.Body>
							{#each auditLogs as log (log.id)}
								<Table.Row>
									<Table.Cell class="whitespace-nowrap text-xs text-muted-foreground">{fmtDt(log.created_at)}</Table.Cell>
									<Table.Cell>
										<div class="text-sm font-medium text-slate-800">{log.username || auditValue(log, 'username') || 'Sistem'}</div>
										<div class="font-mono text-[11px] text-slate-400">{log.user_id || auditValue(log, 'user_id') || '—'}</div>
									</Table.Cell>
									<Table.Cell>
										<Badge variant="outline" class="text-xs">{auditActionLabel(log.action)}</Badge>
									</Table.Cell>
									<Table.Cell class="max-w-2xl text-xs text-slate-600">{auditSummary(log)}</Table.Cell>
								</Table.Row>
							{:else}
								<Table.Row>
									<Table.Cell colspan={4} class="py-10 text-center text-sm text-slate-400">Belum ada audit operasional untuk sesi ini</Table.Cell>
								</Table.Row>
							{/each}
						</Table.Body>
					</Table.Root>
				</Card.Content>
			</Card.Root>

		<!-- Tab: Essay Grading -->
		{:else if activeTab === 'essay'}
			<Card.Root>
				<Card.Header class="pb-2">
					<Card.Title class="text-base">Koreksi Jawaban Uraian ({essays.length} belum dinilai)</Card.Title>
					<p class="text-sm text-slate-500">Nilai 0-100 akan dikalikan proporsional dengan bobot soal di paket.</p>
				</Card.Header>
				<Card.Content class="space-y-3">
					{#each essays as e (e.answer_id)}
						<section class="grid gap-3 rounded-lg border border-slate-200 bg-white p-3 shadow-sm xl:grid-cols-[minmax(0,0.85fr)_minmax(0,1fr)_16rem]">
							<div class="space-y-2">
								<div class="flex flex-wrap items-center gap-2">
									<Badge variant="outline">{e.question_code || 'Essay'}</Badge>
									<Badge variant="outline">Bobot {fmtEssayPoints(e.points)}</Badge>
								</div>
								<div>
									<p class="text-sm font-semibold text-slate-900">{e.nama}</p>
									<p class="text-xs text-slate-500">{e.nis}{e.room_name ? ` · ${e.room_name}` : ''}</p>
								</div>
								{#if e.stimulus_html}
									<div class="rounded-md border border-slate-100 bg-slate-50 p-2">
										<p class="mb-1 text-[10px] font-bold uppercase tracking-wide text-slate-400">Stimulus</p>
										<RichContent html={e.stimulus_html} class="prose prose-sm max-w-none text-slate-700" />
									</div>
								{/if}
								<div class="rounded-md border border-slate-100 bg-slate-50 p-2">
									<p class="mb-1 text-[10px] font-bold uppercase tracking-wide text-slate-400">Soal</p>
									{#if e.stem_html}
										<RichContent html={e.stem_html} class="prose prose-sm max-w-none text-slate-800" />
									{:else}
										<p class="text-sm text-slate-800">{e.question_text}</p>
									{/if}
								</div>
							</div>

							<div class="space-y-2">
								<p class="text-[10px] font-bold uppercase tracking-wide text-slate-500">Jawaban Siswa</p>
								<div class="min-h-32 whitespace-pre-wrap rounded-md border border-slate-200 bg-slate-50 p-3 text-sm text-slate-800">{e.answer || '(jawaban kosong)'}</div>
								{#if e.rubric_html}
									<div class="rounded-md border border-amber-100 bg-amber-50 p-3">
										<p class="mb-1 text-[10px] font-bold uppercase tracking-wide text-amber-700">Rubrik / Pedoman</p>
										<RichContent html={e.rubric_html} class="prose prose-sm max-w-none text-amber-950" />
									</div>
								{/if}
							</div>

							<div class="space-y-2 rounded-md border border-green-100 bg-green-50 p-3">
								<label for={`essay-score-${e.answer_id}`} class="block text-[10px] font-bold uppercase tracking-wide text-green-800">Nilai Manual</label>
								<Input id={`essay-score-${e.answer_id}`} type="number" min="0" max="100" bind:value={gradeInput[e.answer_id]} placeholder="0-100" class="h-9 bg-white" />
								<p class="text-xs text-green-800">0 berarti sudah dikoreksi dengan nilai nol. Kosong berarti belum bisa disimpan.</p>
								<LoadingButton
									size="sm"
									onclick={() => submitGrade(e.answer_id)}
									loading={gradeBusyId === e.answer_id}
									disabled={gradeBusyId !== '' && gradeBusyId !== e.answer_id}
									loadingLabel="Menyimpan..."
									class="w-full bg-green-700 text-white hover:bg-green-800"
								>
									Simpan Nilai
								</LoadingButton>
							</div>
						</section>
					{:else}
						<div class="rounded-lg border border-dashed border-slate-200 py-12 text-center text-sm text-slate-400">
							Tidak ada jawaban uraian yang perlu dikoreksi.
						</div>
					{/each}
				</Card.Content>
			</Card.Root>
		{/if}
		{/snippet}
	</AsyncContent>
</div>
