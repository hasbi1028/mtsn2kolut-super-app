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
	let visibleEvents = $derived((recentAlertEvents.length > 0 ? recentAlertEvents : events.filter(importantEvent)).slice(0, 8));

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
		const text = `Mohon bantuan admin CBT. Ruang: ${card?.room_name ?? room?.room_name ?? 'Ruang ujian'}. Sesi: ${card?.session_title ?? room?.session_title ?? 'CBT'}. Status: ${signalLabel()}. Atensi: ${attentionCount}. Terkunci: ${lockedCount}.`;
		if (browser) navigator.clipboard?.writeText(text).catch(() => undefined);
		successMessage = 'Pesan bantuan disiapkan/disalin. Kirim ke admin/operator CBT.';
		if (demoMode) demoRoomStatus = 'Bantuan Admin Diminta';
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

<main data-cbt-portal-root class="exam-light-scope min-h-screen bg-[#f7fbf5] px-3 py-4 text-slate-950 sm:px-4 sm:py-6" style="background-color: #f7fbf5 !important; color-scheme: light;">
	<section class="mx-auto max-w-5xl space-y-3">
		<header class="rounded-2xl border border-emerald-200 bg-emerald-50 p-4 text-slate-950 shadow-sm sm:p-5">
			<div class="flex flex-wrap items-start justify-between gap-3">
				<div>
					<p class="text-xs font-semibold uppercase tracking-[0.25em] text-emerald-700">MTsN 2 Kolaka Utara</p>
					<h1 class="mt-1 text-2xl font-bold sm:text-3xl">Portal Pengawasan Ruang</h1>
					<p class="mt-2 max-w-3xl text-sm text-slate-700">{demoMode ? 'MODE DEMO meniru alur pengawasan ruang ujian nyata tanpa API/database.' : 'Scan QR pada Lembar Pengawas Ruang, masukkan PIN ruang, lalu pantau peserta secara real-time tanpa login admin.'}</p>
				</div>
				{#if card}
					<div class="rounded-xl border border-emerald-200 bg-white px-3 py-2 text-xs font-semibold text-emerald-800">
						{demoMode ? 'Demo Lokal' : liveMode === 'polling' ? 'Live 3,5 detik' : liveMode === 'paused' ? 'Jeda saat tab tersembunyi' : 'Siap'}
					</div>
				{/if}
			</div>
		</header>

		{#if errorMessage}
			<div class="rounded-xl border border-red-200 bg-red-50 p-3 text-sm text-red-700">{errorMessage}</div>
		{/if}
		{#if successMessage}
			<div class="rounded-xl border border-emerald-200 bg-emerald-50 p-3 text-sm text-emerald-800">{successMessage}</div>
		{/if}

		{#if card}
			<section class="rounded-2xl border border-slate-200 bg-white p-4 text-slate-950 shadow-sm sm:p-5">
				<div class="flex flex-wrap items-start justify-between gap-3">
					<div>
						<p class="text-xs font-semibold uppercase tracking-[0.2em] text-emerald-700">Lembar Pengawas Ruang Valid</p>
						<h2 class="mt-1 text-2xl font-bold">{card.room_name ?? room?.room_name ?? 'Ruang Ujian'}</h2>
						<p class="mt-1 text-sm text-slate-600">{card.session_title ?? room?.session_title ?? 'Sesi CBT'} · {card.package_title ?? room?.package_title ?? 'Paket ujian'}</p>
					</div>
					<button class="rounded-xl border border-slate-300 px-4 py-2 text-sm font-semibold" onclick={() => { stopLivePolling(); card = null; dashboard = null; pin = ''; successMessage = ''; }}>Ganti Lembar/PIN</button>
				</div>

				<div class="mt-4 grid gap-2 rounded-xl bg-emerald-50 p-3 text-sm sm:grid-cols-2 lg:grid-cols-4">
					<p><b>Akses:</b> {card.proctor_name ?? 'Lembar Pengawas Ruang'}</p>
					<p><b>Jadwal:</b> {fmtDt(card.scheduled_start)}</p>
					<p><b>Status:</b> {room?.session_status ?? card.session_status ?? '—'}</p>
					<p><b>Live:</b> {backgroundBusy ? 'Memperbarui...' : demoMode ? 'Demo' : 'Aktif'}</p>
				</div>

				<div class={`mt-4 rounded-2xl border p-4 ${signalClass()}`}>
					<p class="text-xs font-semibold uppercase tracking-[0.2em]">Status Ruang</p>
					<p class="mt-1 text-2xl font-bold">{demoMode ? demoRoomStatus : signalLabel()}</p>
					<div class="mt-4 grid grid-cols-2 gap-2 text-sm sm:grid-cols-5">
						<p><b>Peserta:</b> {room?.participant_count ?? participants.length}</p>
						<p><b>Online:</b> {onlineCount}</p>
						<p><b>Selesai:</b> {submittedCount}</p>
						<p><b>Atensi:</b> {attentionCount}</p>
						<p><b>Terkunci:</b> {lockedCount}</p>
					</div>
				</div>

				<div class="mt-4 grid gap-2 sm:grid-cols-4">
					<button class="rounded-xl bg-emerald-700 px-4 py-4 text-base font-bold text-white disabled:opacity-60" disabled={Boolean(actionBusy) || room?.session_status === 'active'} onclick={() => void updateStatus('active')}>{actionBusy === 'status-active' ? 'Memproses...' : 'Mulai Ujian'}</button>
					<button class="rounded-xl bg-amber-500 px-4 py-4 text-base font-bold text-amber-950" onclick={contactAdmin}>Hubungi Admin</button>
					<button class="rounded-xl bg-slate-800 px-4 py-4 text-base font-bold text-white disabled:opacity-60" disabled={Boolean(actionBusy) || room?.session_status === 'finished'} onclick={() => void updateStatus('finished')}>Tutup Ujian</button>
					<button class="rounded-xl border border-slate-300 px-4 py-4 text-base font-bold {audioAlertsEnabled ? 'bg-emerald-700 text-white' : 'bg-white text-slate-900'}" onclick={() => audioAlertsEnabled = !audioAlertsEnabled}>Audio {audioAlertsEnabled ? 'ON' : 'OFF'}</button>
				</div>

				{#if visibleEvents.length > 0}
					<div class="mt-5 overflow-hidden rounded-2xl border border-red-100">
						<div class="flex items-center justify-between bg-red-50 px-4 py-3">
							<p class="font-semibold text-red-900">Insiden Terbaru</p>
							<p class="text-xs text-red-700">Toast + highlight otomatis</p>
						</div>
						{#each visibleEvents as event (event.id)}
							<div class="border-t border-red-100 px-4 py-3 text-sm">
								<div class="flex flex-wrap items-start justify-between gap-2">
									<div class="min-w-0">
										<p class="font-semibold text-slate-950">{event.nama ?? 'Peserta'} · {eventReason(event)}</p>
										<p class="text-xs text-slate-600">{fmtDt(event.created_at)} · {event.event_type}</p>
									</div>
									<div class="flex flex-wrap gap-1">
										<button class="rounded-lg border border-slate-300 px-2 py-1 text-xs font-semibold" disabled={Boolean(actionBusy) || demoMode} onclick={() => void acknowledgeEvent(event)}>Sudah Dicek</button>
										<button class="rounded-lg border border-amber-300 bg-amber-50 px-2 py-1 text-xs font-semibold text-amber-900" disabled={Boolean(actionBusy) || demoMode} onclick={() => void incidentAction(event, 'warning_given')}>Beri Peringatan</button>
										<button class="rounded-lg border border-red-300 bg-red-50 px-2 py-1 text-xs font-semibold text-red-900" disabled={Boolean(actionBusy) || demoMode} onclick={() => void incidentAction(event, 'escalated')}>Eskalasi</button>
									</div>
								</div>
							</div>
						{/each}
					</div>
				{/if}

				<div class="mt-5 overflow-hidden rounded-2xl border border-slate-200">
					<div class="bg-slate-50 px-4 py-3 font-semibold">Peserta Perlu Dicek</div>
					{#each attentionParticipants.slice(0, 12) as row (row.participant_id)}
						<div class="border-t border-slate-100 px-4 py-3 text-sm {highlightedParticipantIds.has(row.participant_id) ? 'bg-red-50' : ''}">
							<div class="flex flex-wrap items-center justify-between gap-2">
								<div class="min-w-0">
									<p class="font-semibold text-slate-950">{row.nama ?? 'Peserta'} <span class="text-xs text-slate-500">Kursi {row.seat_no ?? '—'}</span></p>
									<div class="mt-1 flex flex-wrap gap-1 text-xs">
										<span class={`rounded-full border px-2 py-0.5 ${heartbeatClass(row)}`}>{heartbeatState(row)}</span>
										<span class="rounded-full border border-slate-200 bg-white px-2 py-0.5 text-slate-700">{riskLabel(row)}</span>
										<span class="rounded-full border border-slate-200 bg-white px-2 py-0.5 text-slate-700">Keluar {row.app_switch_count ?? 0}x</span>
										<span class="rounded-full border border-slate-200 bg-white px-2 py-0.5 text-slate-700">Screenshot {row.screenshot_attempt ?? 0}x</span>
									</div>
								</div>
								<div class="flex flex-wrap gap-1">
									<button class="rounded-lg border border-amber-300 bg-amber-50 px-2 py-1 text-xs font-semibold text-amber-900" disabled={Boolean(actionBusy) || demoMode} onclick={() => void sendWarning(row)}>Peringatkan</button>
									{#if participantLocked(row)}
										<button class="rounded-lg border border-emerald-300 bg-emerald-50 px-2 py-1 text-xs font-semibold text-emerald-900" disabled={Boolean(actionBusy) || demoMode} onclick={() => void unlockParticipant(row)}>Buka Kunci</button>
									{/if}
								</div>
							</div>
						</div>
					{/each}
					{#if attentionParticipants.length === 0}
						<div class="border-t border-slate-100 px-4 py-3 text-sm text-slate-600">Belum ada peserta yang perlu perhatian khusus.</div>
					{/if}
				</div>

				<div class="mt-4 flex flex-wrap gap-2">
					<button class="rounded-xl border border-slate-300 px-4 py-2 text-sm font-semibold" disabled={loading || demoMode} onclick={() => void loadPortalDashboard(false)}>{loading ? 'Memuat...' : 'Muat Ulang'}</button>
				</div>
				<p class="mt-3 text-center text-xs text-slate-500">Pengawas cukup memakai halaman ini. Admin tetap memantau semua ruang dari Command Center.</p>
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
