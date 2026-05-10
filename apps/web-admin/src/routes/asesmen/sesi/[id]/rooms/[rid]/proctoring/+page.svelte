<script lang="ts">
	import FileDownIcon from '@lucide/svelte/icons/file-down';
	import PrinterIcon from '@lucide/svelte/icons/printer';
	import { onDestroy, onMount } from 'svelte';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { toast } from '$lib/components/ui/sonner';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { Textarea } from '$lib/components/ui/textarea';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import OperationStatusPanel from '$lib/components/OperationStatusPanel.svelte';
	import { confirmAction } from '$lib/confirm-dialog';
	import { clientApiPath, readClientApiData, readClientJson } from '$lib/client/api';
	import { csvRow } from '$lib/csv';
	import {
		PROCTOR_EVIDENCE_CATEGORIES,
		buildProctorEvidenceCsvRows,
		classifyProctorEvent,
		proctorEventLabel,
		proctorEvidenceCategoryLabel,
		proctorEvidenceCategorySummary,
		proctorOperatorGuidance,
		summarizeProctorEvidence,
		type ProctorEvidenceCategory
	} from '$lib/cbt/proctor-evidence';

	type RoomDashboard = {
		id: string;
		session_id: string;
		room_name: string;
		capacity: number;
		room_token: string;
		status: string;
		is_locked: boolean;
		session_title: string;
		session_status: string;
		scheduled_start: string;
		scheduled_end: string;
		package_title: string;
		duration_minutes: number;
		school_room_code: string;
		school_room_name: string;
		school_room_building: string;
		school_room_location_note: string;
		participant_count: number;
		submitted_count: number;
		online_count: number;
		suspicious_count: number;
		missing_seat_count: number;
	};
	type RoomProctor = {
		id: string;
		exam_room_id: string;
		employee_id: string;
		nip: string;
		nama: string;
		role: string;
		assigned_at: string;
	};
	type ProctoringRow = {
		participant_id: string;
		student_id: string;
		nis: string;
		nama: string;
		token: string;
		room_id: string | null;
		room_name: string;
		seat_no?: number | null;
		submitted_at: string | null;
		last_heartbeat: string | null;
		app_switch_count: number;
		screenshot_attempt: number;
		suspicious_flag: boolean;
		violation_count: number;
		risk_score: number;
		risk_level: string;
		locked_at: string | null;
		locked_reason: string | null;
		recent_violation_count: number;
		last_violation_at: string | null;
		last_violation_reason: string;
		answered_count: number;
		score: string | null;
	};
	type ProctoringEvent = {
		id: string;
		participant_id: string;
		student_id: string;
		nis: string;
		nama: string;
		room_id: string | null;
		room_name: string;
		event_type: string;
		event_data: unknown;
		created_at: string;
	};
	type DashboardPayload = {
		room: RoomDashboard;
		proctors: RoomProctor[];
		participants: ProctoringRow[];
		events: ProctoringEvent[];
	};
	type RoomHandover = {
		room_id: string;
		session_id: string;
		room_name: string;
		room_token: string;
		room_status: string;
		room_is_locked: boolean;
		session_title: string;
		session_status: string;
		scheduled_start: string;
		scheduled_end: string;
		package_title: string;
		participant_count: number;
		submitted_count: number;
		suspicious_count: number;
		missing_seat_count: number;
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
		locked_by: string | null;
		updated_by: string | null;
		handover_created_at: string | null;
		handover_updated_at: string | null;
	};

	const sessionId = page.params.id ?? '';
	const roomId = page.params.rid ?? '';

	let dashboardPromise = $state<Promise<DashboardPayload> | null>(null);
	let room = $state<RoomDashboard | null>(null);
	let proctors = $state<RoomProctor[]>([]);
	let participants = $state<ProctoringRow[]>([]);
	let events = $state<ProctoringEvent[]>([]);
	let handover = $state<RoomHandover | null>(null);
	let handoverLoadBusy = $state(false);
	let handoverBusy = $state(false);
	let handoverLockBusy = $state(false);
	let refreshBusy = $state(false);
	let backgroundBusy = $state(false);
	let actionBusyId = $state('');
	let operationState = $state<{ tone: 'success' | 'error' | 'warning' | 'info'; title: string; message: string } | null>(null);
	let interval: ReturnType<typeof setInterval> | undefined;
	let liveSource = $state<EventSource | null>(null);
	let liveMode = $state<'connecting' | 'sse' | 'polling'>('connecting');
	let seenEventIds = $state(new Set<string>());
	let recentAlertEvents = $state<ProctoringEvent[]>([]);
	let highlightedParticipantIds = $state(new Map<string, number>());
	let audioAlertsEnabled = $state(false);
	let hasPrimedLiveEvents = false;
	let handoverLocked = $derived(Boolean(handover?.locked_at));

	let participantStats = $derived.by(() => {
		let online = 0;
		let stale = 0;
		let offline = 0;
		let submitted = 0;
		let locked = 0;
		let highRisk = 0;
		for (const row of participants) {
			const state = heartbeatState(row);
			if (state === 'online') online += 1;
			else if (state === 'stale') stale += 1;
			else offline += 1;
			if (row.submitted_at) submitted += 1;
			if (row.locked_at || row.risk_level === 'locked') locked += 1;
			if (row.risk_level === 'high' || row.risk_level === 'locked') highRisk += 1;
		}
		return { online, stale, offline, submitted, locked, highRisk };
	});
	let evidenceSummary = $derived.by(() => summarizeProctorEvidence({ participants, events, hasPrintPack: true }));

	onMount(() => {
		dashboardPromise = loadDashboard();
		void loadHandover();
		connectLiveStream();
		interval = setInterval(() => {
			if (typeof document !== 'undefined' && document.hidden) return;
			if (liveMode === 'sse') return;
			void refreshDashboard(true);
		}, 5000);
	});

	onDestroy(() => {
		if (interval) clearInterval(interval);
		liveSource?.close();
	});

	async function loadDashboard() {
		const res = await fetch(clientApiPath`/api/asesmen/sessions/${sessionId}/rooms/${roomId}/proctoring`);
		const payload = await readClientApiData<DashboardPayload>(res, 'Gagal memuat dashboard pengawas ruang');
		mergeDashboardPayload(payload);
		return payload;
	}


	function mergeDashboardPayload(payload: DashboardPayload) {
		room = payload.room;
		proctors = Array.isArray(payload.proctors) ? payload.proctors : [];
		participants = Array.isArray(payload.participants) ? payload.participants : [];
		const nextEvents = Array.isArray(payload.events) ? payload.events : [];
		if (hasPrimedLiveEvents) notifyNewEvents(nextEvents);
		else {
			primeSeenEvents(nextEvents);
			hasPrimedLiveEvents = true;
		}
		events = nextEvents;
	}

	function primeSeenEvents(items: ProctoringEvent[]) {
		const next = new Set(seenEventIds);
		for (const event of items) next.add(event.id);
		seenEventIds = next;
	}

	function importantEvent(event: ProctoringEvent) {
		return ['anti_cheat_violation', 'app_switch', 'screenshot_attempt', 'proctor_force_submit', 'proctor_reset_access', 'proctor_unlock', 'proctor_acknowledge'].includes(event.event_type);
	}


	function shouldPlayAlertSound(event: ProctoringEvent) {
		if (!audioAlertsEnabled) return false;
		if (event.event_type !== 'anti_cheat_violation') return event.event_type === 'app_switch' || event.event_type === 'screenshot_attempt';
		const reason = eventReason(event).toLowerCase();
		return reason.includes('locked') || reason.includes('high') || reason.includes('split') || reason.includes('switch');
	}

	function playAlertSound() {
		try {
			const Ctx = window.AudioContext || (window as unknown as { webkitAudioContext?: typeof AudioContext }).webkitAudioContext;
			if (!Ctx) return;
			const ctx = new Ctx();
			const oscillator = ctx.createOscillator();
			const gain = ctx.createGain();
			oscillator.type = 'sine';
			oscillator.frequency.setValueAtTime(880, ctx.currentTime);
			gain.gain.setValueAtTime(0.0001, ctx.currentTime);
			gain.gain.exponentialRampToValueAtTime(0.15, ctx.currentTime + 0.02);
			gain.gain.exponentialRampToValueAtTime(0.0001, ctx.currentTime + 0.28);
			oscillator.connect(gain).connect(ctx.destination);
			oscillator.start();
			oscillator.stop(ctx.currentTime + 0.3);
		} catch (error) {
			console.warn('Audio alert gagal diputar', error);
		}
	}

	function eventReason(event: ProctoringEvent) {
		if (event.event_data && typeof event.event_data === 'object' && 'reason' in event.event_data) {
			const reason = (event.event_data as { reason?: unknown }).reason;
			if (typeof reason === 'string' && reason.trim()) return reason.replaceAll('_', ' ');
		}
		return proctorEventLabel(event);
	}

	function notifyNewEvents(nextEvents: ProctoringEvent[]) {
		const known = new Set(seenEventIds);
		const fresh = nextEvents
			.filter((event) => event.id && !known.has(event.id))
			.sort((a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime());
		if (fresh.length === 0) return;
		for (const event of fresh) known.add(event.id);
		seenEventIds = known;
		for (const event of fresh.filter(importantEvent)) {
			recentAlertEvents = [event, ...recentAlertEvents.filter((item) => item.id !== event.id)].slice(0, 12);
			highlightParticipant(event.participant_id);
			const message = `${event.nama}: ${eventReason(event)}`;
			if (event.event_type === 'anti_cheat_violation') toast.warning(message);
			else if (event.event_type === 'app_switch' || event.event_type === 'screenshot_attempt') toast.warning(message);
			else toast.info(message);
			if (shouldPlayAlertSound(event)) playAlertSound();
		}
	}

	function handleLiveEvent(event: ProctoringEvent) {
		notifyNewEvents([event]);
		events = [event, ...events.filter((item) => item.id !== event.id)].slice(0, 100);
		void refreshDashboard(true);
	}

	function connectLiveStream() {
		if (typeof EventSource === 'undefined') {
			liveMode = 'polling';
			return;
		}
		liveSource?.close();
		liveMode = 'connecting';
		const source = new EventSource(clientApiPath`/api/asesmen/sessions/${sessionId}/rooms/${roomId}/proctoring/stream`);
		liveSource = source;
		source.addEventListener('ready', () => {
			liveMode = 'sse';
		});
		source.addEventListener('proctor_event', (message) => {
			liveMode = 'sse';
			try {
				handleLiveEvent(JSON.parse((message as MessageEvent).data) as ProctoringEvent);
			} catch (error) {
				console.error('Invalid proctoring SSE payload', error);
			}
		});
		source.onerror = () => {
			liveMode = 'polling';
			source.close();
			liveSource = null;
			setTimeout(() => {
				if (typeof document !== 'undefined' && document.hidden) return;
				connectLiveStream();
			}, 5000);
		};
	}

	function highlightParticipant(participantId: string) {
		if (!participantId) return;
		const next = new Map(highlightedParticipantIds);
		next.set(participantId, Date.now() + 15000);
		highlightedParticipantIds = next;
		setTimeout(() => {
			const current = highlightedParticipantIds.get(participantId) ?? 0;
			if (current <= Date.now()) {
				const cleared = new Map(highlightedParticipantIds);
				cleared.delete(participantId);
				highlightedParticipantIds = cleared;
			}
		}, 15500);
	}

	function liveModeLabel() {
		if (liveMode === 'sse') return 'Live connected';
		if (liveMode === 'connecting') return 'Menghubungkan live';
		return 'Fallback polling 5 detik';
	}

	function liveModeClass() {
		if (liveMode === 'sse') return 'border-primary/20 bg-primary/10 text-primary';
		if (liveMode === 'connecting') return 'border-warning/30 bg-warning/10 text-warning';
		return 'border-destructive/30 bg-destructive/10 text-destructive';
	}

	async function loadHandover() {
		handoverLoadBusy = true;
		try {
			const res = await fetch(clientApiPath`/api/asesmen/sessions/${sessionId}/rooms/${roomId}/handover`);
			handover = await readClientApiData<RoomHandover>(res, 'Gagal memuat serah terima ruang');
		} catch (error) {
			const message = detailErrorMessage(error);
			operationState = { tone: 'error', title: 'Serah Terima Tidak Termuat', message };
		} finally {
			handoverLoadBusy = false;
		}
	}

	async function refreshDashboard(background = false) {
		if (background) backgroundBusy = true;
		else refreshBusy = true;
		try {
			await loadDashboard();
		} catch (error) {
			if (!background) {
				const message = detailErrorMessage(error);
				operationState = { tone: 'error', title: 'Refresh Gagal', message };
				toast.error(message);
			}
		} finally {
			refreshBusy = false;
			backgroundBusy = false;
		}
	}

	function detailErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Operasi belum berhasil. Periksa koneksi lalu coba lagi.';
	}

	function fmtDate(value: string | null | undefined) {
		if (!value) return '—';
		const date = new Date(value);
		if (Number.isNaN(date.getTime())) return value;
		return date.toLocaleString('id-ID', { dateStyle: 'medium', timeStyle: 'short' });
	}

	function fmtScore(value: string | null) {
		if (!value) return '—';
		const parsed = Number(value);
		if (Number.isNaN(parsed)) return value;
		return parsed.toFixed(2);
	}

	function minutesSince(value: string | null) {
		if (!value) return Number.POSITIVE_INFINITY;
		const date = new Date(value);
		if (Number.isNaN(date.getTime())) return Number.POSITIVE_INFINITY;
		return (Date.now() - date.getTime()) / 60000;
	}

	function heartbeatState(row: ProctoringRow): 'submitted' | 'online' | 'stale' | 'offline' {
		if (row.submitted_at) return 'submitted';
		const minutes = minutesSince(row.last_heartbeat);
		if (minutes <= 2) return 'online';
		if (minutes <= 7) return 'stale';
		return 'offline';
	}

	function heartbeatLabel(row: ProctoringRow) {
		const state = heartbeatState(row);
		if (state === 'submitted') return 'Submit';
		if (state === 'online') return 'Online';
		if (state === 'stale') return 'Waspada';
		return 'Offline';
	}

	function heartbeatClass(row: ProctoringRow) {
		const state = heartbeatState(row);
		if (state === 'submitted') return 'border-border bg-muted text-muted-foreground';
		if (state === 'online') return 'border-primary/20 bg-primary/10 text-primary';
		if (state === 'stale') return 'border-warning/30 bg-warning/10 text-warning';
		return 'border-destructive/30 bg-destructive/10 text-destructive';
	}

	function riskLabel(row: ProctoringRow) {
		if (row.locked_at || row.risk_level === 'locked') return 'Locked';
		if (row.risk_level === 'high') return 'High';
		if (row.risk_level === 'warning') return 'Warning';
		return 'Normal';
	}

	function riskClass(row: ProctoringRow) {
		if (row.locked_at || row.risk_level === 'locked') return 'border-destructive/40 bg-destructive/15 text-destructive';
		if (row.risk_level === 'high') return 'border-warning/40 bg-warning/15 text-warning';
		if (row.risk_level === 'warning') return 'border-accent bg-accent/60 text-accent-foreground';
		return 'border-primary/20 bg-primary/10 text-primary';
	}

	function rowAttentionClass(row: ProctoringRow) {
		if (row.locked_at || row.risk_level === 'locked') return 'bg-destructive/15';
		if (row.risk_level === 'high' || row.suspicious_flag) return 'bg-destructive/10';
		if (row.risk_level === 'warning' || row.app_switch_count >= 3 || row.violation_count > 0) return 'bg-warning/10';
		return '';
	}

	function evidenceCategoryLabel(category: ProctorEvidenceCategory | null) {
		return proctorEvidenceCategoryLabel(category);
	}

	function evidenceCategoryDescription(category: ProctorEvidenceCategory | null) {
		return proctorEvidenceCategorySummary(category);
	}

	function evidenceCategoryClass(category: ProctorEvidenceCategory | null) {
		if (category === 'device_mismatch' || category === 'submit_guard' || category === 'stale_connection') {
			return 'border-warning/30 bg-warning/10 text-warning';
		}
		if (category === 'force_submit' || category === 'reset_access' || category === 'anti_cheat') return 'border-destructive/30 bg-destructive/10 text-destructive';
		if (category === 'heartbeat' || category === 'export_print') return 'border-primary/20 bg-primary/10 text-primary';
		return 'border-border bg-card text-muted-foreground';
	}

	function exportEvidenceCSV() {
		if (!room) return;
		const rows = buildProctorEvidenceCsvRows({
			sessionTitle: room.session_title,
			roomName: room.room_name,
			proctors: proctors.map((proctor) => proctor.nama),
			participants,
			events,
		});
		const csv = rows.map((row) => csvRow(row)).join('\n');
		const blob = new Blob([csv], { type: 'text/csv;charset=utf-8' });
		const url = URL.createObjectURL(blob);
		const link = document.createElement('a');
		link.href = url;
		link.download = `bukti_pengawas_${safeFileName(room.session_title)}_${safeFileName(room.room_name)}.csv`;
		link.click();
		URL.revokeObjectURL(url);
	}

	function safeFileName(value: string) {
		const safe = value.trim().replace(/[^A-Za-z0-9_-]+/g, '_').replace(/^_+|_+$/g, '');
		return safe || 'cbt';
	}

	async function flagParticipant(row: ProctoringRow, flag: boolean) {
		actionBusyId = `flag-${row.participant_id}`;
		try {
			const res = await fetch(clientApiPath`/api/asesmen/sessions/${sessionId}/rooms/${roomId}/participants/${row.participant_id}/flag`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ flag }),
			});
			await readClientJson<unknown>(res);
			operationState = {
				tone: flag ? 'warning' : 'success',
				title: flag ? 'Peserta Ditandai' : 'Tanda Dibersihkan',
				message: `${row.nama} ${flag ? 'masuk daftar atensi pengawas.' : 'tidak lagi ditandai sebagai atensi.'}`,
			};
			await refreshDashboard(true);
		} catch (error) {
			toast.error(detailErrorMessage(error));
		} finally {
			actionBusyId = '';
		}
	}

	async function resetAccess(row: ProctoringRow) {
		if (!(await confirmAction({
			title: 'Reset Akses Peserta',
			message: `Reset perangkat dan heartbeat ${row.nama}. Gunakan ini saat siswa perlu login ulang di ruang ini.`,
			confirmLabel: 'Reset Akses',
			tone: 'warning',
		}))) return;
		actionBusyId = `reset-${row.participant_id}`;
		try {
			const res = await fetch(clientApiPath`/api/asesmen/sessions/${sessionId}/rooms/${roomId}/participants/${row.participant_id}/reset-access`, { method: 'POST' });
			await readClientJson<unknown>(res);
			operationState = { tone: 'success', title: 'Akses Direset', message: `${row.nama} dapat login ulang setelah diverifikasi pengawas.` };
			await refreshDashboard(true);
		} catch (error) {
			toast.error(detailErrorMessage(error));
		} finally {
			actionBusyId = '';
		}
	}


	async function unlockParticipant(row: ProctoringRow) {
		const notes = window.prompt(`Catatan unlock untuk ${row.nama}`, 'Sudah diverifikasi pengawas ruang') ?? '';
		if (!(await confirmAction({
			title: 'Unlock Peserta',
			message: `Buka status terkunci ${row.nama} tanpa menghapus histori pelanggaran.`,
			confirmLabel: 'Unlock',
			tone: 'warning',
		}))) return;
		actionBusyId = `unlock-${row.participant_id}`;
		try {
			const res = await fetch(clientApiPath`/api/asesmen/sessions/${sessionId}/rooms/${roomId}/participants/${row.participant_id}/unlock`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ notes }),
			});
			await readClientJson<unknown>(res);
			operationState = { tone: 'success', title: 'Peserta Di-unlock', message: `${row.nama} sudah dibuka kembali dan tetap masuk atensi pengawas.` };
			await refreshDashboard(true);
		} catch (error) {
			toast.error(detailErrorMessage(error));
		} finally {
			actionBusyId = '';
		}
	}

	async function acknowledgeEvent(event: ProctoringEvent) {
		const row = participants.find((item) => item.participant_id === event.participant_id);
		if (!row) return;
		const notes = window.prompt(`Catatan pemeriksaan untuk ${event.nama}`, eventReason(event)) ?? '';
		actionBusyId = `ack-${event.id}`;
		try {
			const res = await fetch(clientApiPath`/api/asesmen/sessions/${sessionId}/rooms/${roomId}/participants/${event.participant_id}/acknowledge`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ event_id: event.id, notes }),
			});
			await readClientJson<unknown>(res);
			toast.success(`${event.nama} ditandai sudah diperiksa`);
			await refreshDashboard(true);
		} catch (error) {
			toast.error(detailErrorMessage(error));
		} finally {
			actionBusyId = '';
		}
	}

	async function forceSubmit(row: ProctoringRow) {
		if (!(await confirmAction({
			title: 'Paksa Submit Peserta',
			message: `Paksa submit jawaban ${row.nama}. Tindakan ini dipakai hanya saat ujian ruang sudah harus ditutup.`,
			confirmLabel: 'Paksa Submit',
			tone: 'danger',
		}))) return;
		actionBusyId = `submit-${row.participant_id}`;
		try {
			const res = await fetch(clientApiPath`/api/asesmen/sessions/${sessionId}/rooms/${roomId}/participants/${row.participant_id}/force-submit`, { method: 'POST' });
			await readClientJson<unknown>(res);
			operationState = { tone: 'warning', title: 'Peserta Disubmit', message: `${row.nama} sudah dipaksa submit dari dashboard ruang.` };
			await refreshDashboard(true);
		} catch (error) {
			toast.error(detailErrorMessage(error));
		} finally {
			actionBusyId = '';
		}
	}

	function handoverPayload() {
		return {
			attendance_checked: handover?.attendance_checked === true,
			all_submitted_checked: handover?.all_submitted_checked === true,
			device_issue_checked: handover?.device_issue_checked === true,
			room_clean_checked: handover?.room_clean_checked === true,
			token_returned_checked: handover?.token_returned_checked === true,
			assets_returned_checked: handover?.assets_returned_checked === true,
			incident_notes: handover?.incident_notes ?? '',
			operator_notes: handover?.operator_notes ?? '',
			handover_notes: handover?.handover_notes ?? '',
		};
	}

	async function saveHandover() {
		handoverBusy = true;
		try {
			const res = await fetch(clientApiPath`/api/asesmen/sessions/${sessionId}/rooms/${roomId}/handover`, {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(handoverPayload()),
			});
			await readClientJson<unknown>(res);
			operationState = { tone: 'success', title: 'Serah Terima Tersimpan', message: 'Checklist dan catatan akhir ruang sudah disimpan.' };
			await loadHandover();
		} catch (error) {
			toast.error(detailErrorMessage(error));
		} finally {
			handoverBusy = false;
		}
	}

	async function lockHandover() {
		if (!(await confirmAction({
			title: 'Kunci Serah Terima Ruang',
			message: 'Setelah dikunci, checklist dan catatan ruang menjadi arsip akhir dan tidak dapat diedit dari dashboard pengawas.',
			confirmLabel: 'Kunci',
			tone: 'warning',
		}))) return;
		handoverLockBusy = true;
		try {
			const res = await fetch(clientApiPath`/api/asesmen/sessions/${sessionId}/rooms/${roomId}/handover/lock`, { method: 'POST' });
			await readClientJson<unknown>(res);
			operationState = { tone: 'success', title: 'Serah Terima Dikunci', message: 'Ruang sudah memiliki bukti penutupan digital.' };
			await loadHandover();
		} catch (error) {
			toast.error(detailErrorMessage(error));
		} finally {
			handoverLockBusy = false;
		}
	}

	function handleRenderError(error: unknown) {
		console.error('CBT room proctoring dashboard render failed', error);
	}
</script>

<svelte:head>
	<title>Dashboard Pengawas Ruang | CBT</title>
</svelte:head>

<div class="space-y-5 p-4 md:p-6">
	<div class="flex flex-col gap-3 border-b border-primary/20 pb-4 md:flex-row md:items-start md:justify-between">
		<div>
			<a href={resolve(`/asesmen/sesi/${sessionId}`)} class="text-xs font-semibold uppercase tracking-[0.18em] text-primary">Kembali ke detail sesi</a>
			<h1 class="mt-2 text-2xl font-bold tracking-tight text-foreground">Dashboard Pengawas Ruang</h1>
			<p class="text-sm text-muted-foreground">{room?.session_title ?? 'Memuat sesi'} · {room?.room_name ?? 'Memuat ruang'}</p>
		</div>
		<div class="flex flex-wrap items-center gap-2">
			<Badge variant="outline" class={liveModeClass()}>{liveModeLabel()}</Badge>
			{#if backgroundBusy}
				<Badge variant="outline" class="border-primary/20 text-primary">Memperbarui</Badge>
			{/if}
			<Button variant="outline" onclick={exportEvidenceCSV} disabled={!room}>
				<FileDownIcon class="mr-2 size-4" />
				Export Evidence CSV (tanpa token)
			</Button>
			<Button variant={audioAlertsEnabled ? 'default' : 'outline'} onclick={() => audioAlertsEnabled = !audioAlertsEnabled}>
				Audio {audioAlertsEnabled ? 'ON' : 'OFF'}
			</Button>
			<Button variant="outline" href={resolve(`/asesmen/sesi/${sessionId}/rooms/${roomId}/print-pack`)}>
				<PrinterIcon class="mr-2 size-4" />
				Paket Cetak
			</Button>
			<LoadingButton variant="outline" onclick={() => void refreshDashboard()} loading={refreshBusy} loadingLabel="Memuat...">
				Refresh
			</LoadingButton>
		</div>
	</div>

	{#if operationState}
		<OperationStatusPanel {...operationState} />
	{/if}

	{#if recentAlertEvents.length > 0}
		<Card.Root class="border-warning/30 bg-warning/5">
			<Card.Header class="pb-2">
				<Card.Title class="text-base">Live Alert Kecurangan</Card.Title>
			</Card.Header>
			<Card.Content class="grid gap-2 md:grid-cols-2 xl:grid-cols-3">
				{#each recentAlertEvents.slice(0, 6) as event (event.id)}
					<div class="rounded-lg border border-warning/20 bg-card p-3 text-sm">
						<div class="font-semibold text-foreground">{event.nama}</div>
						<div class="text-muted-foreground">{eventReason(event)} · {fmtDate(event.created_at)}</div>
					</div>
				{/each}
			</Card.Content>
		</Card.Root>
	{/if}

	<AsyncContent promise={dashboardPromise} onerror={handleRenderError}>
		{#snippet pending()}
			<div class="grid gap-4 lg:grid-cols-4">
				{#each Array.from({ length: 4 }) as _, index (index)}
					<Skeleton class="h-28 rounded-lg" />
				{/each}
			</div>
			<Skeleton class="h-96 rounded-lg" />
		{/snippet}

		{#if room}
				<div class="grid gap-3 md:grid-cols-2 xl:grid-cols-5">
					<Card.Root class="border-primary/20">
						<Card.Header class="pb-2">
							<Card.Title class="text-sm text-muted-foreground">Peserta</Card.Title>
						</Card.Header>
						<Card.Content>
							<div class="text-2xl font-bold text-foreground">{room.participant_count}</div>
							<p class="text-xs text-muted-foreground">Kapasitas {room.capacity}</p>
						</Card.Content>
					</Card.Root>
					<Card.Root class="border-primary/20">
						<Card.Header class="pb-2">
							<Card.Title class="text-sm text-muted-foreground">Online</Card.Title>
						</Card.Header>
						<Card.Content>
							<div class="text-2xl font-bold text-primary">{participantStats.online}</div>
							<p class="text-xs text-muted-foreground">Heartbeat 2 menit terakhir</p>
						</Card.Content>
					</Card.Root>
					<Card.Root class="border-primary/20">
						<Card.Header class="pb-2">
							<Card.Title class="text-sm text-muted-foreground">Waspada / Offline</Card.Title>
						</Card.Header>
						<Card.Content>
							<div class="text-2xl font-bold text-warning">{participantStats.stale + participantStats.offline}</div>
							<p class="text-xs text-muted-foreground">Perlu dicek pengawas</p>
						</Card.Content>
					</Card.Root>
					<Card.Root class="border-primary/20">
						<Card.Header class="pb-2">
							<Card.Title class="text-sm text-muted-foreground">Submit</Card.Title>
						</Card.Header>
						<Card.Content>
							<div class="text-2xl font-bold text-foreground">{participantStats.submitted}</div>
							<p class="text-xs text-muted-foreground">Dari {room.participant_count} peserta</p>
						</Card.Content>
					</Card.Root>
					<Card.Root class="border-primary/20">
						<Card.Header class="pb-2">
							<Card.Title class="text-sm text-muted-foreground">Anti-cheat</Card.Title>
						</Card.Header>
						<Card.Content>
							<div class="text-2xl font-bold text-destructive">{participantStats.highRisk}</div>
							<p class="text-xs text-muted-foreground">{participantStats.locked} terkunci lokal/server</p>
						</Card.Content>
					</Card.Root>
				</div>

				<Card.Root class="border-primary/20">
					<Card.Header>
						<div class="flex flex-col gap-2 md:flex-row md:items-start md:justify-between">
							<div>
								<Card.Title>{room.room_name}</Card.Title>
								<Card.Description>
									{room.package_title} · {fmtDate(room.scheduled_start)} - {fmtDate(room.scheduled_end)}
								</Card.Description>
							</div>
							<div class="flex flex-wrap items-center gap-2">
								<Badge variant="outline" class="border-primary/20 text-primary">Token ruang {room.room_token || '—'}</Badge>
								<Badge variant="outline">{room.session_status}</Badge>
							</div>
						</div>
					</Card.Header>
					<Card.Content class="grid gap-4 md:grid-cols-3">
						<div class="rounded-lg border border-border p-3">
							<p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Lokasi</p>
							<p class="mt-1 text-sm font-medium text-foreground">{room.school_room_code ? `${room.school_room_code} · ${room.school_room_name}` : 'Ruang manual sesi'}</p>
							<p class="text-xs text-muted-foreground">{room.school_room_building || room.school_room_location_note || 'Lokasi belum dicatat'}</p>
						</div>
						<div class="rounded-lg border border-border p-3">
							<p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Pengawas</p>
							<div class="mt-1 space-y-1">
								{#each proctors as proctor (proctor.id)}
									<p class="text-sm text-foreground">{proctor.nama} <span class="text-xs text-muted-foreground">({proctor.role})</span></p>
								{:else}
									<p class="text-sm text-warning">Belum ada pengawas</p>
								{/each}
							</div>
						</div>
						<div class="rounded-lg border border-border p-3">
							<p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Status Ruang</p>
							<p class="mt-1 text-sm text-foreground">Durasi paket {room.duration_minutes} menit</p>
							<p class="text-xs text-muted-foreground">{room.is_locked ? 'Ruang dikunci' : 'Ruang masih dapat diperbarui operator'}</p>
						</div>
					</Card.Content>
				</Card.Root>

				<Card.Root class="border-primary/20">
					<Card.Header class="pb-3">
						<Card.Title>Mode Bukti Pengawas</Card.Title>
						<Card.Description>Ringkasan evidence ruang dari heartbeat, warning BYOD, tindakan pengawas, dan paket export/print.</Card.Description>
					</Card.Header>
					<Card.Content class="grid gap-2 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5">
						{#each PROCTOR_EVIDENCE_CATEGORIES as category (category)}
							<div class="rounded-lg border border-border bg-card p-3">
								<p class="text-[11px] font-semibold uppercase tracking-wide text-muted-foreground">{evidenceCategoryLabel(category)}</p>
								<p class="mt-1 text-xl font-bold text-foreground">{evidenceSummary.counts[category]}</p>
								<p class="text-[11px] text-muted-foreground">{evidenceCategoryDescription(category)}</p>
								<p class="mt-1 text-[11px] text-muted-foreground">{evidenceSummary.missingCategories.includes(category) ? 'Belum ada bukti di data aktif' : 'Tercatat di evidence ruang'}</p>
							</div>
						{/each}
					</Card.Content>
				</Card.Root>

				<Card.Root class="border-primary/20">
					<Card.Header class="pb-3">
						<Card.Title>Panduan Tindakan Pengawas</Card.Title>
						<Card.Description>Kapan memberi peringatan, reset akses, atau paksa submit saat rehearsal ruang.</Card.Description>
					</Card.Header>
					<Card.Content class="grid gap-3 md:grid-cols-3">
						{#each proctorOperatorGuidance as item (item.title)}
							<div class="rounded-lg border border-border bg-muted/40 p-3">
								<p class="text-sm font-semibold text-foreground">{item.title}</p>
								<p class="mt-1 text-xs leading-5 text-muted-foreground">{item.description}</p>
							</div>
						{/each}
					</Card.Content>
				</Card.Root>

				<Card.Root class="border-primary/20">
					<Card.Header class="pb-3">
						<div class="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
							<div>
								<Card.Title>Serah Terima Akhir Ruang</Card.Title>
								<Card.Description>Checklist penutupan, catatan insiden, dan bukti penguncian ruang setelah ujian.</Card.Description>
							</div>
							<div class="flex flex-wrap items-center gap-2">
								<Badge variant="outline" class={handoverLocked ? 'border-primary/20 bg-primary/10 text-primary' : 'border-warning/30 bg-warning/10 text-warning'}>
									{handoverLocked ? 'Terkunci' : 'Belum dikunci'}
								</Badge>
								{#if handoverLoadBusy}
									<Badge variant="outline">Memuat</Badge>
								{/if}
							</div>
						</div>
					</Card.Header>
					<Card.Content>
						{#if handover}
							<div class="grid gap-4 xl:grid-cols-[minmax(0,1fr)_320px]">
								<div class="grid gap-2 sm:grid-cols-2 xl:grid-cols-3">
									<label for="handover-attendance" class="flex min-h-14 items-start gap-3 rounded-lg border border-border bg-card p-3 text-sm text-foreground">
										<input id="handover-attendance" type="checkbox" class="mt-0.5 size-4 accent-primary" checked={handover.attendance_checked} disabled={handoverLocked || handoverBusy} onchange={(event) => handover && (handover.attendance_checked = event.currentTarget.checked)} />
										<span><span class="font-semibold text-foreground">Daftar hadir</span><br /><span class="text-xs text-muted-foreground">Paraf/kehadiran peserta sudah dicek.</span></span>
									</label>
									<label for="handover-submitted" class="flex min-h-14 items-start gap-3 rounded-lg border border-border bg-card p-3 text-sm text-foreground">
										<input id="handover-submitted" type="checkbox" class="mt-0.5 size-4 accent-primary" checked={handover.all_submitted_checked} disabled={handoverLocked || handoverBusy} onchange={(event) => handover && (handover.all_submitted_checked = event.currentTarget.checked)} />
										<span><span class="font-semibold text-foreground">Submit akhir</span><br /><span class="text-xs text-muted-foreground">{participantStats.submitted}/{room.participant_count} peserta tercatat.</span></span>
									</label>
									<label for="handover-device" class="flex min-h-14 items-start gap-3 rounded-lg border border-border bg-card p-3 text-sm text-foreground">
										<input id="handover-device" type="checkbox" class="mt-0.5 size-4 accent-primary" checked={handover.device_issue_checked} disabled={handoverLocked || handoverBusy} onchange={(event) => handover && (handover.device_issue_checked = event.currentTarget.checked)} />
										<span><span class="font-semibold text-foreground">Gangguan dicatat</span><br /><span class="text-xs text-muted-foreground">{room.suspicious_count} atensi, {room.missing_seat_count} meja kosong.</span></span>
									</label>
									<label for="handover-clean" class="flex min-h-14 items-start gap-3 rounded-lg border border-border bg-card p-3 text-sm text-foreground">
										<input id="handover-clean" type="checkbox" class="mt-0.5 size-4 accent-primary" checked={handover.room_clean_checked} disabled={handoverLocked || handoverBusy} onchange={(event) => handover && (handover.room_clean_checked = event.currentTarget.checked)} />
										<span><span class="font-semibold text-foreground">Ruang rapi</span><br /><span class="text-xs text-muted-foreground">Meja, kursi, listrik, dan jaringan dicek.</span></span>
									</label>
									<label for="handover-token" class="flex min-h-14 items-start gap-3 rounded-lg border border-border bg-card p-3 text-sm text-foreground">
										<input id="handover-token" type="checkbox" class="mt-0.5 size-4 accent-primary" checked={handover.token_returned_checked} disabled={handoverLocked || handoverBusy} onchange={(event) => handover && (handover.token_returned_checked = event.currentTarget.checked)} />
										<span><span class="font-semibold text-foreground">Token/berkas</span><br /><span class="text-xs text-muted-foreground">Token ruang dan berkas pengawas dikembalikan.</span></span>
									</label>
									<label for="handover-assets" class="flex min-h-14 items-start gap-3 rounded-lg border border-border bg-card p-3 text-sm text-foreground">
										<input id="handover-assets" type="checkbox" class="mt-0.5 size-4 accent-primary" checked={handover.assets_returned_checked} disabled={handoverLocked || handoverBusy} onchange={(event) => handover && (handover.assets_returned_checked = event.currentTarget.checked)} />
										<span><span class="font-semibold text-foreground">Aset cadangan</span><br /><span class="text-xs text-muted-foreground">Perangkat pinjaman/cadangan sudah kembali.</span></span>
									</label>
								</div>
								<div class="space-y-2 rounded-lg border border-border bg-muted/50 p-3">
									<div class="grid grid-cols-3 gap-2 text-center">
										<div>
											<p class="text-lg font-bold text-foreground">{participantStats.submitted}</p>
											<p class="text-[11px] uppercase tracking-wide text-muted-foreground">Submit</p>
										</div>
										<div>
											<p class="text-lg font-bold text-destructive">{room.suspicious_count}</p>
											<p class="text-[11px] uppercase tracking-wide text-muted-foreground">Atensi</p>
										</div>
										<div>
											<p class="text-lg font-bold text-warning">{participantStats.stale + participantStats.offline}</p>
											<p class="text-[11px] uppercase tracking-wide text-muted-foreground">Cek ulang</p>
										</div>
									</div>
									<p class="border-t border-border pt-2 text-xs text-muted-foreground">
										Terakhir diperbarui {fmtDate(handover.handover_updated_at)}. {handoverLocked ? `Dikunci ${fmtDate(handover.locked_at)}.` : 'Simpan draft sebelum mengunci.'}
									</p>
								</div>
							</div>
							<div class="mt-4 grid gap-3 lg:grid-cols-3">
								<div>
									<label for="handover-incident-notes" class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Catatan Kejadian</label>
									<Textarea id="handover-incident-notes" class="mt-1 min-h-24" bind:value={handover.incident_notes} disabled={handoverLocked || handoverBusy} placeholder="Gangguan perangkat, jaringan, keterlambatan, atau kejadian ruang." />
								</div>
								<div>
									<label for="handover-operator-notes" class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Catatan Operator</label>
									<Textarea id="handover-operator-notes" class="mt-1 min-h-24" bind:value={handover.operator_notes} disabled={handoverLocked || handoverBusy} placeholder="Tindak lanjut operator, reset akses, atau verifikasi submit." />
								</div>
								<div>
									<label for="handover-notes" class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Catatan Serah Terima</label>
									<Textarea id="handover-notes" class="mt-1 min-h-24" bind:value={handover.handover_notes} disabled={handoverLocked || handoverBusy} placeholder="Ringkasan akhir untuk kepala madrasah/panitia." />
								</div>
							</div>
							<div class="mt-4 flex flex-col gap-2 border-t border-border pt-4 sm:flex-row sm:items-center sm:justify-between">
								<p class="text-xs text-muted-foreground">Penguncian membuat catatan menjadi arsip akhir ruang. Perubahan setelah itu dilakukan melalui prosedur operator.</p>
								<div class="flex flex-wrap gap-2">
									<LoadingButton variant="outline" onclick={() => void saveHandover()} loading={handoverBusy} disabled={handoverLocked || handoverLockBusy} loadingLabel="Menyimpan...">
										Simpan Draft
									</LoadingButton>
									<LoadingButton onclick={() => void lockHandover()} loading={handoverLockBusy} disabled={handoverLocked || handoverBusy} loadingLabel="Mengunci...">
										Kunci Handover
									</LoadingButton>
								</div>
							</div>
						{:else}
							<div class="grid gap-3 md:grid-cols-3">
								<Skeleton class="h-20 rounded-lg" />
								<Skeleton class="h-20 rounded-lg" />
								<Skeleton class="h-20 rounded-lg" />
							</div>
						{/if}
					</Card.Content>
				</Card.Root>

				<div class="grid gap-4 xl:grid-cols-[minmax(0,1fr)_380px]">
					<Card.Root class="border-primary/20">
						<Card.Header>
							<Card.Title>Peserta Ruang</Card.Title>
							<Card.Description>Monitoring heartbeat, submit, dan tindakan pengawas terbatas pada ruang ini.</Card.Description>
						</Card.Header>
						<Card.Content class="overflow-x-auto">
							<Table.Root>
								<Table.Header>
									<Table.Row class="bg-success/10">
										<Table.Head>Peserta</Table.Head>
										<Table.Head class="text-center">Meja</Table.Head>
										<Table.Head>Status</Table.Head>
										<Table.Head class="text-center">Jawab</Table.Head>
										<Table.Head class="text-center">Switch</Table.Head>
										<Table.Head class="text-center">SS</Table.Head>
										<Table.Head>Risk</Table.Head>
										<Table.Head class="text-center">Skor</Table.Head>
										<Table.Head class="text-right">Aksi</Table.Head>
									</Table.Row>
								</Table.Header>
								<Table.Body>
									{#each participants as row (row.participant_id)}
											<Table.Row class={`${rowAttentionClass(row)} ${highlightedParticipantIds.has(row.participant_id) ? 'ring-2 ring-warning/60 bg-warning/15' : ''}`}>
											<Table.Cell>
												<div class="font-medium text-foreground">{row.nama}</div>
												<div class="text-xs text-muted-foreground">{row.nis}</div>
											</Table.Cell>
											<Table.Cell class="text-center font-mono">{row.seat_no ?? '—'}</Table.Cell>
											<Table.Cell>
												<Badge variant="outline" class={heartbeatClass(row)}>{heartbeatLabel(row)}</Badge>
												<p class="mt-1 text-[11px] text-muted-foreground">{fmtDate(row.last_heartbeat)}</p>
											</Table.Cell>
											<Table.Cell class="text-center font-mono">{row.answered_count}</Table.Cell>
											<Table.Cell class="text-center font-mono">{row.app_switch_count}</Table.Cell>
											<Table.Cell class="text-center font-mono">{row.screenshot_attempt}</Table.Cell>
											<Table.Cell>
												<Badge variant="outline" class={riskClass(row)}>{riskLabel(row)} · {row.risk_score}</Badge>
												<p class="mt-1 text-[11px] text-muted-foreground">{row.violation_count} violation{row.last_violation_reason ? ` · ${row.last_violation_reason.replaceAll('_', ' ')}` : ''}</p>
											</Table.Cell>
											<Table.Cell class="text-center font-mono">{fmtScore(row.score)}</Table.Cell>
											<Table.Cell class="min-w-[260px] text-right">
												<div class="flex flex-wrap justify-end gap-2">
													<Button size="sm" variant="outline" onclick={() => void flagParticipant(row, !row.suspicious_flag)}>
														{row.suspicious_flag ? 'Bersihkan' : 'Tandai'}
													</Button>
													<LoadingButton size="sm" variant="outline" onclick={() => void resetAccess(row)} loading={actionBusyId === `reset-${row.participant_id}`} disabled={actionBusyId !== '' && actionBusyId !== `reset-${row.participant_id}`} loadingLabel="Reset...">
														Reset
													</LoadingButton>
													<LoadingButton size="sm" onclick={() => void forceSubmit(row)} loading={actionBusyId === `submit-${row.participant_id}`} disabled={!!row.submitted_at || (actionBusyId !== '' && actionBusyId !== `submit-${row.participant_id}`)} loadingLabel="Submit...">
														Submit
													</LoadingButton>
												</div>
											</Table.Cell>
										</Table.Row>
									{:else}
										<Table.Row>
											<Table.Cell colspan={9} class="py-10 text-center text-muted-foreground">Belum ada peserta di ruang ini</Table.Cell>
										</Table.Row>
									{/each}
								</Table.Body>
							</Table.Root>
						</Card.Content>
					</Card.Root>

					<Card.Root class="border-primary/20">
						<Card.Header>
							<Card.Title>Log Ruang</Card.Title>
							<Card.Description>Aktivitas terakhir dari peserta ruang ini, dikategorikan sebagai evidence operasional BYOD.</Card.Description>
						</Card.Header>
						<Card.Content class="space-y-3">
							{#each events.slice(0, 15) as event (event.id)}
								{@const eventCategory = classifyProctorEvent(event)}
								<div class="rounded-lg border border-border p-3">
									<div class="flex items-center justify-between gap-2">
										<p class="text-sm font-medium text-foreground">{event.nama}</p>
										<div class="flex flex-wrap justify-end gap-1">
											<Badge variant="outline" class={evidenceCategoryClass(eventCategory)}>{evidenceCategoryLabel(eventCategory)}</Badge>
											<Badge variant="outline" class="text-[11px]">{proctorEventLabel(event)}</Badge>
										</div>
									</div>
									<p class="mt-1 text-xs text-muted-foreground">{event.nis} · {fmtDate(event.created_at)}</p>
					<div class="mt-2 flex justify-end">
						<LoadingButton size="sm" variant="outline" onclick={() => void acknowledgeEvent(event)} loading={actionBusyId === `ack-${event.id}`} disabled={actionBusyId !== '' && actionBusyId !== `ack-${event.id}`} loadingLabel="Menyimpan...">
							Tandai diperiksa
						</LoadingButton>
					</div>
								</div>
							{:else}
								<p class="rounded-lg border border-dashed border-border p-5 text-center text-sm text-muted-foreground">Belum ada log ruang</p>
							{/each}
						</Card.Content>
					</Card.Root>
				</div>
		{/if}
	</AsyncContent>
</div>
