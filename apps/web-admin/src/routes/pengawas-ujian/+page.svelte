<script lang="ts">
	import { browser } from '$app/environment';
	import { page } from '$app/stores';
	import { onDestroy, onMount } from 'svelte';
	import { toast } from '$lib/components/ui/sonner';
	import {
		proctorEventReasonLabel,
		proctorIncidentActionLabel,
		type ProctorEvidenceEvent
	} from '$lib/cbt/proctor-evidence';

	type ProctorCard = {
		card_id?: string;
		card_type?: string;
		event_title?: string;
		session_id?: string;
		session_title?: string;
		session_status?: string;
		scheduled_start?: string;
		scheduled_end?: string;
		room_id?: string;
		room_name?: string;
		proctor_name?: string;
		proctor_role?: string;
		package_title?: string;
	};
	type RoomDashboard = {
		room_name?: string;
		session_title?: string;
		session_status?: string;
		package_title?: string;
		participant_count?: number;
		submitted_count?: number;
		online_count?: number;
		suspicious_count?: number;
		missing_seat_count?: number;
	};
	type ProctoringRow = {
		participant_id: string;
		nama?: string;
		nis?: string;
		room_name?: string;
		seat_no?: number | null;
		submitted_at?: string | null;
		last_heartbeat?: string | null;
		risk_level?: string;
		locked_at?: string | null;
		suspicious_flag?: boolean;
		app_switch_count?: number;
		screenshot_attempt?: number;
		violation_count?: number;
	};
	type ProctoringEvent = ProctorEvidenceEvent & {
		id: string;
		participant_id: string;
	};
	type PortalDashboard = {
		card?: ProctorCard;
		room?: RoomDashboard;
		participants?: ProctoringRow[];
		events?: ProctoringEvent[];
	};

	let token = $state('');
	let pin = $state('');
	let loading = $state(false);
	let backgroundBusy = $state(false);
	let actionBusy = $state('');
	let errorMessage = $state('');
	let successMessage = $state('');
	let card = $state<ProctorCard | null>(null);
	let dashboard = $state<PortalDashboard | null>(null);
	let demoRoomStatus = $state<'Menunggu' | 'Ujian Dibuka' | 'Bantuan Admin Diminta' | 'Ujian Ditutup'>('Menunggu');
	let liveMode = $state<'idle' | 'polling' | 'paused'>('idle');
	let seenEventIds = $state(new Set<string>());
	let recentAlertEvents = $state<ProctoringEvent[]>([]);
	let highlightedParticipantIds = $state(new Map<string, number>());
	let audioAlertsEnabled = $state(false);
	let activeTab = $state<'ruang' | 'peringatan' | 'peserta'>('ruang');
	let alertFilter = $state<'all' | 'red' | 'yellow' | 'unchecked'>('all');
	let adminHelpText = $state('');
	let pollInterval: ReturnType<typeof setInterval> | undefined;
	let hasPrimedEvents = false;

	let queryCard = $derived($page.url.searchParams.get('card') ?? $page.url.searchParams.get('token') ?? '');
	let demoMode = $derived($page.url.searchParams.get('demo') === '1');
	let participants = $derived(dashboard?.participants ?? []);
	let events = $derived(dashboard?.events ?? []);
	let room = $derived(dashboard?.room ?? null);
	let onlineCount = $derived(room?.online_count ?? participants.filter((row) => heartbeatState(row) === 'online').length);
	let submittedCount = $derived(room?.submitted_count ?? participants.filter((row) => row.submitted_at).length);
	let lockedCount = $derived(participants.filter(participantLocked).length);
	let attentionParticipants = $derived(participants.filter(participantNeedsAttention));
	let attentionCount = $derived(attentionParticipants.length + (room?.suspicious_count ?? 0) + (room?.missing_seat_count ?? 0));
	let roomSignal = $derived(attentionCount > 0 || lockedCount > 0 ? 'red' : onlineCount > 0 ? 'green' : 'yellow');
	let alertEvents = $derived((recentAlertEvents.length > 0 ? recentAlertEvents : events.filter(importantEvent)).slice(0, 12));
	let visibleEvents = $derived.by<ProctoringEvent[]>(() => {
		const source = alertEvents;
		if (alertFilter === 'red') return source.filter((event) => eventSeverity(event) === 'red');
		if (alertFilter === 'yellow') return source.filter((event) => eventSeverity(event) !== 'red');
		if (alertFilter === 'unchecked') return source.filter((event) => !['proctor_acknowledge', 'proctor_incident_action'].includes(event.event_type));
		return source;
	});

	$effect(() => {
		if (queryCard && !token) token = queryCard;
		if (demoMode && !card) activateDemo();
	});

	onMount(() => {
		if (card && !demoMode) startLivePolling();
	});

	onDestroy(() => stopLivePolling());

	function activateDemo() {
		card = demoCard();
		dashboard = demoDashboard(card);
		primeSeenEvents(dashboard.events ?? []);
	}

	async function verifyCard() {
		loading = true;
		errorMessage = '';
		successMessage = '';
		card = null;
		dashboard = null;
		try {
			const response = await fetch('/api/exam/proctor/card/verify', {
				method: 'POST',
				headers: { 'content-type': 'application/json' },
				body: JSON.stringify({ token: token.trim(), pin: pin.trim() })
			});
			const body = await response.json().catch(() => ({}));
			if (!response.ok) throw new Error(String(body.error ?? body.message ?? 'Lembar pengawas ruang tidak cocok atau PIN salah.'));
			card = (body.data?.card ?? body.card ?? {}) as ProctorCard;
			if (browser && card.session_id && card.room_id) {
				sessionStorage.setItem('cbt_proctor_card_room', JSON.stringify({ token: token.trim(), verified_at: new Date().toISOString(), card }));
			}
			await loadPortalDashboard(false);
			successMessage = 'PIN benar. Portal pengawasan ruang real-time sudah terbuka tanpa login admin.';
			startLivePolling();
		} catch (error) {
			errorMessage = error instanceof Error ? error.message : 'Lembar pengawas ruang tidak cocok atau PIN salah.';
		} finally {
			loading = false;
		}
	}

	async function loadPortalDashboard(background = false) {
		if (demoMode) return;
		if (!background) loading = true;
		else backgroundBusy = true;
		errorMessage = '';
		try {
			const response = await fetch('/api/exam/proctor/portal/dashboard', {
				method: 'POST',
				headers: { 'content-type': 'application/json' },
				body: JSON.stringify({ token: token.trim(), pin: pin.trim() })
			});
			const payload = await readApiPayload<PortalDashboard>(response, 'Panel pengawasan ruang belum dapat dimuat.');
			mergeDashboard(payload);
		} finally {
			loading = false;
			backgroundBusy = false;
		}
	}

	function mergeDashboard(payload: PortalDashboard) {
		dashboard = payload;
		card = payload.card ?? card;
		const nextEvents = Array.isArray(payload.events) ? payload.events : [];
		if (hasPrimedEvents) notifyNewEvents(nextEvents);
		else {
			primeSeenEvents(nextEvents);
			hasPrimedEvents = true;
		}
	}

	function startLivePolling() {
		stopLivePolling();
		if (demoMode || !card) return;
		liveMode = 'polling';
		pollInterval = setInterval(() => {
			if (typeof document !== 'undefined' && document.hidden) {
				liveMode = 'paused';
				return;
			}
			liveMode = 'polling';
			void loadPortalDashboard(true);
		}, 3500);
	}

	function stopLivePolling() {
		if (pollInterval) clearInterval(pollInterval);
		pollInterval = undefined;
		liveMode = 'idle';
	}

	async function updateStatus(status: 'active' | 'finished') {
		if (demoMode) {
			demoRoomStatus = status === 'active' ? 'Ujian Dibuka' : 'Ujian Ditutup';
			return;
		}
		actionBusy = `status-${status}`;
		errorMessage = '';
		successMessage = '';
		try {
			const response = await fetch('/api/exam/proctor/portal/status', {
				method: 'POST',
				headers: { 'content-type': 'application/json' },
				body: JSON.stringify({ token: token.trim(), pin: pin.trim(), status })
			});
			await readApiPayload(response, status === 'active' ? 'Ujian belum dapat dimulai.' : 'Ujian belum dapat ditutup.');
			successMessage = status === 'active' ? 'Ujian dimulai. Peserta dapat masuk/mengerjakan.' : 'Ujian ditutup. Pastikan serah terima ke admin/operator selesai.';
			await loadPortalDashboard(true);
		} catch (error) {
			errorMessage = error instanceof Error ? error.message : 'Operasi portal pengawasan gagal.';
		} finally {
			actionBusy = '';
		}
	}

	async function acknowledgeEvent(event: ProctoringEvent) {
		const notes = window.prompt('Catatan pemeriksaan pengawas:', eventReason(event));
		if (notes === null) return;
		await portalAction(`/api/exam/proctor/portal/participants/${encodeURIComponent(event.participant_id)}/acknowledge`, { event_id: event.id, notes }, `Kejadian ${event.nama ?? 'peserta'} ditandai sudah diperiksa.`);
	}

	async function incidentAction(event: ProctoringEvent, action: 'warning_given' | 'cleared' | 'escalated') {
		const notes = window.prompt(`Catatan ${proctorIncidentActionLabel(action)}:`, eventReason(event));
		if (notes === null) return;
		await portalAction(`/api/exam/proctor/portal/participants/${encodeURIComponent(event.participant_id)}/incident-action`, { event_id: event.id, action, notes }, 'Tindakan insiden tersimpan.');
	}

	async function sendWarning(row: ProctoringRow) {
		const message = window.prompt('Instruksi/peringatan ke aplikasi siswa:', 'Tetap di aplikasi ujian dan ikuti arahan pengawas.');
		if (!message?.trim()) return;
		await portalAction(`/api/exam/proctor/portal/participants/${encodeURIComponent(row.participant_id)}/command`, { command_type: 'warning_message', message }, `Peringatan dikirim ke ${row.nama ?? 'peserta'}.`);
	}

	async function unlockParticipant(row: ProctoringRow) {
		const notes = window.prompt('Catatan verifikasi sebelum buka kunci:', 'Sudah diverifikasi pengawas ruang.');
		if (notes === null) return;
		await portalAction(`/api/exam/proctor/portal/participants/${encodeURIComponent(row.participant_id)}/unlock`, { notes }, `Kunci ${row.nama ?? 'peserta'} dibuka.`);
	}

	async function portalAction(url: string, payload: Record<string, unknown>, message: string) {
		actionBusy = url;
		try {
			const response = await fetch(url, {
				method: 'POST',
				headers: { 'content-type': 'application/json' },
				body: JSON.stringify({ token: token.trim(), pin: pin.trim(), ...payload })
			});
			await readApiPayload(response, 'Aksi pengawas belum berhasil.');
			toast.success(message);
			successMessage = message;
			await loadPortalDashboard(true);
		} catch (error) {
			const msg = error instanceof Error ? error.message : 'Aksi pengawas gagal.';
			toast.error(msg);
			errorMessage = msg;
		} finally {
			actionBusy = '';
		}
	}

	function contactAdmin() {
		const latest = alertEvents[0];
		const text = `Mohon bantuan admin CBT.\n\nUjian: ${card?.session_title ?? room?.session_title ?? 'CBT'}\nRuang: ${card?.room_name ?? room?.room_name ?? 'Ruang ujian'}\nStatus: ${signalLabel()}\nAtensi: ${attentionCount} peserta/peringatan\nTerkunci: ${lockedCount}\n${latest ? `Peserta terbaru: ${latest.nama ?? 'Peserta'} - ${eventReason(latest)}\nWaktu: ${fmtDt(latest.created_at)}` : 'Belum ada detail peserta terbaru.'}`;
		adminHelpText = text;
		if (browser) navigator.clipboard?.writeText(text).catch(() => undefined);
		successMessage = 'Pesan bantuan disiapkan/disalin. Kirim ke admin/operator CBT.';
		if (demoMode) demoRoomStatus = 'Bantuan Admin Diminta';
	}

	function copyAdminHelpText() {
		if (!adminHelpText) contactAdmin();
		if (browser && adminHelpText) navigator.clipboard?.writeText(adminHelpText).catch(() => undefined);
		successMessage = 'Teks bantuan admin disalin.';
	}

	async function readApiPayload<T = unknown>(response: Response, fallback: string): Promise<T> {
		const body = await response.json().catch(() => ({}));
		if (!response.ok) throw new Error(String(body.error ?? body.message ?? fallback));
		return (body.data ?? body) as T;
	}

	function primeSeenEvents(items: ProctoringEvent[]) {
		const next = new Set(seenEventIds);
		for (const event of items) if (event.id) next.add(event.id);
		seenEventIds = next;
	}

	function importantEvent(event: ProctoringEvent) {
		return ['anti_cheat_violation', 'app_switch', 'screenshot_attempt', 'proctor_force_submit', 'proctor_reset_access', 'proctor_unlock', 'proctor_acknowledge', 'proctor_incident_action', 'participant_command'].includes(event.event_type);
	}

	function notifyNewEvents(nextEvents: ProctoringEvent[]) {
		const known = new Set(seenEventIds);
		const fresh = nextEvents.filter((event) => event.id && !known.has(event.id)).sort((a, b) => new Date(a.created_at ?? '').getTime() - new Date(b.created_at ?? '').getTime());
		if (fresh.length === 0) return;
		for (const event of fresh) known.add(event.id);
		seenEventIds = known;
		for (const event of fresh.filter(importantEvent)) {
			recentAlertEvents = [event, ...recentAlertEvents.filter((item) => item.id !== event.id)].slice(0, 12);
			highlightParticipant(event.participant_id);
			const message = `${event.nama ?? 'Peserta'}: ${eventReason(event)}`;
			toast.warning(message);
			if (shouldPlayAlertSound(event)) playAlertSound();
		}
	}

	function shouldPlayAlertSound(event: ProctoringEvent) {
		if (!audioAlertsEnabled) return false;
		return ['anti_cheat_violation', 'app_switch', 'screenshot_attempt'].includes(event.event_type);
	}

	function playAlertSound() {
		try {
			const Ctx = window.AudioContext || (window as unknown as { webkitAudioContext?: typeof AudioContext }).webkitAudioContext;
			if (!Ctx) return;
			const ctx = new Ctx();
			const oscillator = ctx.createOscillator();
			const gain = ctx.createGain();
			oscillator.frequency.setValueAtTime(880, ctx.currentTime);
			gain.gain.setValueAtTime(0.0001, ctx.currentTime);
			gain.gain.exponentialRampToValueAtTime(0.18, ctx.currentTime + 0.02);
			gain.gain.exponentialRampToValueAtTime(0.0001, ctx.currentTime + 0.35);
			oscillator.connect(gain).connect(ctx.destination);
			oscillator.start();
			oscillator.stop(ctx.currentTime + 0.35);
		} catch {
			// Audio alert is optional.
		}
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

	function demoCard(): ProctorCard {
		return { card_type: 'proctor', event_title: 'MODE DEMO Pengawasan CBT', session_id: 'demo-session', room_id: 'demo-room', session_title: 'Simulasi Pengawasan Portal Web', session_status: 'demo', scheduled_start: new Date().toISOString(), scheduled_end: new Date(Date.now() + 45 * 60_000).toISOString(), room_name: 'Ruang DEMO 01', proctor_name: 'Lembar Pengawas Ruang', proctor_role: 'pengawas_ruang', package_title: 'Informatika — Contoh lokal' };
	}

	function demoDashboard(demoCardValue: ProctorCard): PortalDashboard {
		const now = new Date().toISOString();
		return {
			card: demoCardValue,
			room: { room_name: demoCardValue.room_name, session_title: demoCardValue.session_title, session_status: 'scheduled', package_title: demoCardValue.package_title, participant_count: 32, submitted_count: 0, online_count: 29, suspicious_count: 2 },
			participants: [
				{ participant_id: 'demo-1', nama: 'Ahmad Demo', nis: '24001', seat_no: 12, last_heartbeat: now, risk_level: 'high', suspicious_flag: true, app_switch_count: 4, screenshot_attempt: 1 },
				{ participant_id: 'demo-2', nama: 'Siti Demo', nis: '24002', seat_no: 18, last_heartbeat: new Date(Date.now() - 9 * 60_000).toISOString(), risk_level: 'warning', suspicious_flag: false, app_switch_count: 1, screenshot_attempt: 0 },
				{ participant_id: 'demo-3', nama: 'Budi Demo', nis: '24003', seat_no: 22, last_heartbeat: now, risk_level: 'locked', locked_at: now, suspicious_flag: true, app_switch_count: 5, screenshot_attempt: 2 }
			],
			events: [
				{ id: 'demo-e1', participant_id: 'demo-1', event_type: 'app_switch', event_data: { state: 'paused' }, created_at: now, nama: 'Ahmad Demo', nis: '24001', room_name: demoCardValue.room_name },
				{ id: 'demo-e2', participant_id: 'demo-3', event_type: 'anti_cheat_violation', event_data: { reason: 'anti_cheat_locked' }, created_at: now, nama: 'Budi Demo', nis: '24003', room_name: demoCardValue.room_name }
			]
		};
	}

	function heartbeatState(row: ProctoringRow): 'submitted' | 'online' | 'stale' | 'offline' {
		if (row.submitted_at) return 'submitted';
		if (!row.last_heartbeat) return 'offline';
		const minutes = (Date.now() - new Date(row.last_heartbeat).getTime()) / 60000;
		if (Number.isNaN(minutes)) return 'offline';
		if (minutes <= 2) return 'online';
		if (minutes <= 8) return 'stale';
		return 'offline';
	}

	function participantLocked(row: ProctoringRow) {
		return Boolean(row.locked_at || row.risk_level === 'locked');
	}

	function participantNeedsAttention(row: ProctoringRow) {
		const state = heartbeatState(row);
		return Boolean(participantLocked(row) || row.suspicious_flag || row.risk_level === 'high' || row.risk_level === 'warning' || state === 'offline' || state === 'stale' || (row.app_switch_count ?? 0) >= 3 || (row.screenshot_attempt ?? 0) > 0 || (row.violation_count ?? 0) > 0);
	}

	function eventReason(event: ProctoringEvent) {
		return proctorEventReasonLabel(event);
	}

	function eventSeverity(event: ProctoringEvent): 'red' | 'orange' | 'yellow' {
		const type = event.event_type.trim().toLowerCase();
		const reason = JSON.stringify(event.event_data ?? {}).toLowerCase();
		if (['anti_cheat_violation', 'device_mismatch', 'device_mismatch_strong', 'token_reuse_confirmed', 'screenshot_attempt', 'screenshot_attempt_valid'].includes(type)) return 'red';
		if (type.includes('device') || type.includes('token') || reason.includes('anti_cheat_locked') || reason.includes('device_mismatch')) return 'red';
		if (type.includes('app_switch') || type.includes('visibility') || type.includes('focus') || type.includes('stale') || type.includes('heartbeat_failed')) return 'orange';
		return 'yellow';
	}

	function eventSeverityLabel(event: ProctoringEvent) {
		const severity = eventSeverity(event);
		if (severity === 'red') return 'Merah — Perlu Admin';
		if (severity === 'orange') return 'Oranye — Perhatian';
		return 'Kuning — Perlu Dicek';
	}

	function eventSeverityClass(event: ProctoringEvent) {
		const severity = eventSeverity(event);
		if (severity === 'red') return 'border-red-200 bg-red-50 text-red-950';
		if (severity === 'orange') return 'border-orange-200 bg-orange-50 text-orange-950';
		return 'border-amber-200 bg-amber-50 text-amber-950';
	}

	function signalLabel() {
		if (roomSignal === 'green') return 'Hijau — aman';
		if (roomSignal === 'yellow') return 'Kuning — menunggu/cek koneksi';
		return 'Merah — perlu perhatian';
	}

	function signalClass() {
		if (roomSignal === 'green') return 'border-emerald-200 bg-emerald-50 text-emerald-950';
		if (roomSignal === 'yellow') return 'border-amber-200 bg-amber-50 text-amber-950';
		return 'border-red-200 bg-red-50 text-red-950';
	}

	function heartbeatClass(row: ProctoringRow) {
		const state = heartbeatState(row);
		if (state === 'online') return 'bg-emerald-50 text-emerald-800 border-emerald-200';
		if (state === 'stale') return 'bg-amber-50 text-amber-800 border-amber-200';
		if (state === 'submitted') return 'bg-slate-50 text-slate-700 border-slate-200';
		return 'bg-red-50 text-red-800 border-red-200';
	}

	function riskLabel(row: ProctoringRow) {
		if (participantLocked(row)) return 'Terkunci';
		if (row.risk_level === 'high') return 'Risiko tinggi';
		if (row.risk_level === 'warning' || row.suspicious_flag) return 'Perlu dicek';
		return 'Normal';
	}

	function fmtDt(value: string | undefined | null) {
		if (!value) return '—';
		const date = new Date(value);
		if (Number.isNaN(date.getTime())) return value;
		return date.toLocaleString('id-ID', { timeZone: 'Asia/Makassar', day: '2-digit', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit' }) + ' WITA';
	}
</script>

<svelte:head>
	<title>Portal Pengawasan Ruang — MTsN 2 Kolaka Utara</title>
	<meta name="color-scheme" content="light" />
	<meta name="theme-color" content="#f7fbf5" />
</svelte:head>

<main data-cbt-portal-root class="exam-light-scope min-h-dvh bg-[#f7fbf5] px-3 py-4 text-slate-950" style="background-color: #f7fbf5 !important; color-scheme: light;">
	<section class="mx-auto max-w-md space-y-3">
		<header class="rounded-[1.75rem] border border-emerald-200 bg-emerald-50 p-4 text-slate-950 shadow-sm">
			<p class="text-xs font-semibold uppercase tracking-[0.25em] text-emerald-700">MTsN 2 Kolaka Utara</p>
			<h1 class="mt-1 text-2xl font-black">Portal Pengawasan Ruang</h1>
			<p class="mt-2 text-sm text-slate-700">{demoMode ? 'MODE DEMO meniru alur pengawasan ruang ujian nyata tanpa API/database.' : 'Scan QR pada Lembar Pengawas Ruang, masukkan PIN ruang, lalu pantau peserta tanpa login admin.'}</p>
		</header>

		{#if card}
			<section class="mx-auto flex min-h-[calc(100dvh-2rem)] max-w-md flex-col overflow-hidden rounded-[2rem] border border-emerald-100 bg-white text-slate-950 shadow-xl sm:min-h-[760px]">
				<div class="sticky top-0 z-20 border-b border-emerald-100 bg-white/95 px-4 pb-3 pt-[calc(0.75rem+env(safe-area-inset-top))] backdrop-blur">
					<div class="flex items-start justify-between gap-3">
						<div class="min-w-0">
							<p class="text-[10px] font-bold uppercase tracking-[0.22em] text-emerald-700">Portal Pengawasan Ruang</p>
							<h2 class="mt-1 truncate text-xl font-black">{card.room_name ?? room?.room_name ?? 'Ruang Ujian'}</h2>
							<p class="truncate text-xs text-slate-600">{card.session_title ?? room?.session_title ?? 'Sesi CBT'}</p>
						</div>
						<div class="shrink-0 rounded-2xl border border-emerald-200 bg-emerald-50 px-3 py-2 text-right text-[10px] font-bold text-emerald-800">
							<p>{demoMode ? 'DEMO' : liveMode === 'polling' ? 'LIVE' : liveMode === 'paused' ? 'JEDA' : 'SIAP'}</p>
							<p class="font-medium normal-case">{backgroundBusy ? 'sinkron...' : '3–5 dtk'}</p>
						</div>
					</div>
					{#if adminHelpText}
						<div class="mt-3 rounded-2xl border border-amber-200 bg-amber-50 p-3 text-xs text-amber-950">
							<p class="font-bold">Pesan bantuan admin siap disalin.</p>
							<pre class="mt-2 max-h-28 overflow-auto whitespace-pre-wrap rounded-xl bg-white/70 p-2 text-[11px] leading-5">{adminHelpText}</pre>
							<button class="mt-2 min-h-10 rounded-xl bg-amber-500 px-3 text-xs font-bold text-amber-950" onclick={copyAdminHelpText}>Salin Lagi</button>
						</div>
					{/if}
				</div>

				<div class="flex-1 overflow-y-auto px-4 py-4 pb-[calc(5.5rem+env(safe-area-inset-bottom))]">
					{#if activeTab === 'ruang'}
						<div class="space-y-4">
							<div class={`rounded-[1.5rem] border p-4 ${signalClass()}`}>
								<p class="text-xs font-bold uppercase tracking-[0.18em]">Status Ruang</p>
								<p class="mt-1 text-2xl font-black">{demoMode ? demoRoomStatus : signalLabel()}</p>
								<p class="mt-2 text-sm opacity-80">{card.package_title ?? room?.package_title ?? 'Paket ujian'} · {fmtDt(card.scheduled_start)}</p>
							</div>

							<div class="grid grid-cols-2 gap-2">
								<div class="rounded-2xl border border-slate-200 bg-slate-50 p-3"><p class="text-[11px] text-slate-500">Peserta</p><p class="text-2xl font-black">{room?.participant_count ?? participants.length}</p></div>
								<div class="rounded-2xl border border-emerald-200 bg-emerald-50 p-3"><p class="text-[11px] text-emerald-700">Online</p><p class="text-2xl font-black text-emerald-900">{onlineCount}</p></div>
								<div class="rounded-2xl border border-amber-200 bg-amber-50 p-3"><p class="text-[11px] text-amber-700">Perlu dicek</p><p class="text-2xl font-black text-amber-950">{attentionCount}</p></div>
								<div class="rounded-2xl border border-slate-200 bg-white p-3"><p class="text-[11px] text-slate-500">Selesai</p><p class="text-2xl font-black">{submittedCount}</p></div>
							</div>

							<div class="grid gap-2">
								<button class="min-h-14 rounded-2xl bg-emerald-700 px-4 text-base font-black text-white disabled:opacity-60" disabled={Boolean(actionBusy) || room?.session_status === 'active'} onclick={() => void updateStatus('active')}>{actionBusy === 'status-active' ? 'Memproses...' : 'Mulai Ujian'}</button>
								<div class="grid grid-cols-2 gap-2">
									<button class="min-h-12 rounded-2xl bg-amber-500 px-3 text-sm font-black text-amber-950" onclick={contactAdmin}>Hubungi Admin</button>
									<button class="min-h-12 rounded-2xl border border-slate-300 px-3 text-sm font-black {audioAlertsEnabled ? 'bg-emerald-700 text-white' : 'bg-white text-slate-900'}" onclick={() => audioAlertsEnabled = !audioAlertsEnabled}>Audio {audioAlertsEnabled ? 'ON' : 'OFF'}</button>
								</div>
								<button class="min-h-12 rounded-2xl bg-slate-900 px-4 text-sm font-black text-white disabled:opacity-60" disabled={Boolean(actionBusy) || room?.session_status === 'finished'} onclick={() => void updateStatus('finished')}>Tutup Ujian</button>
							</div>

							<div class="rounded-2xl border border-slate-200 bg-slate-50 p-3 text-xs leading-5 text-slate-600">
								<p class="font-bold text-slate-900">Pegangan pengawas</p>
								<p>Jika ada peserta merah/kuning, buka tab <b>Peringatan</b>, cek nama peserta, lalu pilih Sudah Dicek, Beri Peringatan, atau Hubungi Admin.</p>
							</div>
						</div>
					{:else if activeTab === 'peringatan'}
						<div class="space-y-3">
							<div class="flex gap-2 overflow-x-auto pb-1">
								{#each [
									{ key: 'all', label: 'Semua' },
									{ key: 'red', label: 'Merah' },
									{ key: 'yellow', label: 'Kuning' },
									{ key: 'unchecked', label: 'Belum Dicek' }
								] as filter}
									<button class="shrink-0 rounded-full border px-3 py-2 text-xs font-bold {alertFilter === filter.key ? 'border-emerald-700 bg-emerald-700 text-white' : 'border-slate-200 bg-white text-slate-700'}" onclick={() => alertFilter = filter.key as typeof alertFilter}>{filter.label}</button>
								{/each}
							</div>

							{#each visibleEvents as event (event.id)}
								<article class={`rounded-[1.35rem] border p-4 shadow-sm ${eventSeverityClass(event)}`}>
									<div class="flex items-start justify-between gap-3">
										<div class="min-w-0">
											<p class="text-[11px] font-black uppercase tracking-[0.16em]">{eventSeverityLabel(event)}</p>
											<h3 class="mt-1 truncate text-lg font-black">{event.nama ?? 'Peserta'}</h3>
											<p class="text-sm font-semibold">{eventReason(event)}</p>
											<p class="mt-1 text-xs opacity-75">{fmtDt(event.created_at)} · {event.room_name ?? card.room_name ?? 'Ruang'}</p>
										</div>
									</div>
									<div class="mt-3 grid grid-cols-2 gap-2">
										<button class="min-h-11 rounded-xl border border-slate-300 bg-white/80 px-2 text-xs font-bold" disabled={Boolean(actionBusy) || demoMode} onclick={() => void acknowledgeEvent(event)}>Sudah Dicek</button>
										<button class="min-h-11 rounded-xl border border-amber-300 bg-amber-100 px-2 text-xs font-bold text-amber-950" disabled={Boolean(actionBusy) || demoMode} onclick={() => void incidentAction(event, 'warning_given')}>Beri Peringatan</button>
										<button class="col-span-2 min-h-11 rounded-xl bg-red-700 px-2 text-xs font-bold text-white" onclick={contactAdmin}>Hubungi Admin</button>
									</div>
								</article>
							{/each}
							{#if visibleEvents.length === 0}
								<div class="rounded-2xl border border-dashed border-slate-300 bg-slate-50 p-5 text-center text-sm text-slate-600">Belum ada peringatan pada filter ini.</div>
							{/if}
						</div>
					{:else}
						<div class="space-y-3">
							{#each participants as row (row.participant_id)}
								<article class="rounded-[1.25rem] border border-slate-200 bg-white p-4 shadow-sm {highlightedParticipantIds.has(row.participant_id) ? 'ring-2 ring-red-300' : ''}">
									<div class="flex items-start justify-between gap-3">
										<div class="min-w-0">
											<h3 class="truncate text-base font-black">{row.nama ?? 'Peserta'}</h3>
											<p class="text-xs text-slate-500">NIS {row.nis ?? '—'} · Kursi {row.seat_no ?? '—'}</p>
										</div>
										<span class={`shrink-0 rounded-full border px-2 py-1 text-[11px] font-bold ${heartbeatClass(row)}`}>{heartbeatState(row)}</span>
									</div>
									<div class="mt-3 flex flex-wrap gap-1 text-[11px]">
										<span class="rounded-full border border-slate-200 bg-slate-50 px-2 py-1">{riskLabel(row)}</span>
										<span class="rounded-full border border-slate-200 bg-slate-50 px-2 py-1">Keluar {row.app_switch_count ?? 0}x</span>
										<span class="rounded-full border border-slate-200 bg-slate-50 px-2 py-1">Screenshot {row.screenshot_attempt ?? 0}x</span>
									</div>
									{#if participantNeedsAttention(row)}
										<div class="mt-3 grid grid-cols-2 gap-2">
											<button class="min-h-10 rounded-xl border border-amber-300 bg-amber-50 px-2 text-xs font-bold text-amber-950" disabled={Boolean(actionBusy) || demoMode} onclick={() => void sendWarning(row)}>Beri Peringatan</button>
											<button class="min-h-10 rounded-xl border border-slate-300 px-2 text-xs font-bold" onclick={contactAdmin}>Hubungi Admin</button>
										</div>
									{/if}
								</article>
							{/each}
							{#if participants.length === 0}
								<div class="rounded-2xl border border-dashed border-slate-300 bg-slate-50 p-5 text-center text-sm text-slate-600">Belum ada peserta terbaca.</div>
							{/if}
						</div>
					{/if}
				</div>

				<nav class="fixed inset-x-0 bottom-0 z-30 mx-auto max-w-md border-t border-slate-200 bg-white/95 px-3 pb-[calc(0.55rem+env(safe-area-inset-bottom))] pt-2 backdrop-blur">
					<div class="grid grid-cols-3 gap-2">
						<button class="min-h-12 rounded-2xl text-xs font-black {activeTab === 'ruang' ? 'bg-emerald-700 text-white' : 'bg-slate-100 text-slate-700'}" onclick={() => activeTab = 'ruang'}>Ruang</button>
						<button class="relative min-h-12 rounded-2xl text-xs font-black {activeTab === 'peringatan' ? 'bg-emerald-700 text-white' : 'bg-slate-100 text-slate-700'}" onclick={() => activeTab = 'peringatan'}>Peringatan{#if alertEvents.length > 0}<span class="absolute -right-1 -top-1 grid size-6 place-items-center rounded-full bg-red-600 text-[10px] text-white">{alertEvents.length}</span>{/if}</button>
						<button class="min-h-12 rounded-2xl text-xs font-black {activeTab === 'peserta' ? 'bg-emerald-700 text-white' : 'bg-slate-100 text-slate-700'}" onclick={() => activeTab = 'peserta'}>Peserta</button>
					</div>
				</nav>
			</section>
		{:else}
			<section class="rounded-2xl border border-slate-200 bg-white p-5 text-slate-950 shadow-sm">
				<h2 class="text-xl font-bold">{demoMode ? 'MODE DEMO Portal Pengawasan' : 'Masuk dengan Lembar Pengawas Ruang'}</h2>
				<p class="mt-1 text-sm text-slate-600">{demoMode ? 'Demo langsung menampilkan ruang contoh.' : 'Kode QR biasanya terisi otomatis setelah scan. Jika tidak, ketik kode lembar ruang secara manual.'}</p>
				<div class="mt-4 grid gap-3 sm:grid-cols-2">
					<label class="space-y-1 text-sm font-medium">Kode Lembar / QR Token<input class="w-full rounded-lg border border-slate-300 bg-white px-3 py-2 text-slate-950" bind:value={token} autocomplete="off" /></label>
					<label class="space-y-1 text-sm font-medium">PIN<input class="w-full rounded-lg border border-slate-300 bg-white px-3 py-2 text-center text-xl tracking-[0.4em] text-slate-950" bind:value={pin} inputmode="numeric" autocomplete="one-time-code" maxlength="8" placeholder="••••" /></label>
				</div>
				{#if !demoMode}
					<button class="mt-4 w-full rounded-xl bg-emerald-700 px-4 py-3 font-semibold text-white disabled:opacity-60" disabled={loading} onclick={verifyCard}>{loading ? 'Memeriksa...' : 'Masuk Portal Pengawasan'}</button>
				{:else}
					<button class="mt-4 w-full rounded-xl bg-emerald-700 px-4 py-3 font-semibold text-white" onclick={activateDemo}>Mulai DEMO Pengawasan</button>
				{/if}
			</section>
		{/if}
	</section>
</main>
