<script lang="ts">
	import PrinterIcon from '@lucide/svelte/icons/printer';
	import { onMount } from 'svelte';
	import QRCode from 'qrcode';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { clientApiPath, readClientApiData } from '$lib/client/api';
	import { fetchSchoolProfile, schoolAddressLine, type SchoolProfile } from '$lib/school-profile';

	type EventInfo = { id: string; title: string; exam_type?: string; academic_year_name?: string; status?: string };
	type EventSession = { id: string; title: string; scheduled_start?: string; scheduled_end?: string; package_title?: string; status?: string };
	type SessionRoom = { id: string; room_name: string; room_token?: string; session_id?: string; school_room_code?: string; school_room_name?: string; participant_count?: number };
	type RoomProctor = { id: string; nama: string; nip?: string; role?: string; employee_id?: string };
	type SupervisorCard = {
		id: string;
		eventTitle: string;
		sessionId: string;
		sessionTitle: string;
		scheduledStart: string;
		packageTitle: string;
		roomId: string;
		roomName: string;
		roomToken: string;
		pin: string;
		portalUrl: string;
		qrDataUrl: string;
		proctorName: string;
		proctorRole: string;
		participantCount: number;
		status?: string;
		token?: string;
		qrPath?: string;
	};
	type PrintData = { schoolProfile: SchoolProfile; eventInfo: EventInfo | null; cards: SupervisorCard[] };

	const eventId = page.params.id ?? '';
	let cardsPromise = $state<Promise<PrintData> | null>(null);
	let issueBusy = $state(false);

	onMount(() => loadCards());

	function loadCards() {
		cardsPromise = fetchCards();
	}

	async function fetchCards(): Promise<PrintData> {
		const [schoolProfile, eventInfo, cards] = await Promise.all([fetchSchoolProfile(), fetchEventInfo(), fetchSupervisorCards()]);
		return { schoolProfile, eventInfo, cards };
	}

	async function fetchSupervisorCards() {
		const response = await fetch(clientApiPath`/api/asesmen/events/${eventId}/supervisor-access-cards`);
		const payload = await readClientApiData<unknown>(response, 'Gagal memuat kartu pengawas QR+PIN');
		const rows = Array.isArray(payload)
			? payload
			: payload && typeof payload === 'object' && Array.isArray((payload as { cards?: unknown[] }).cards)
				? (payload as { cards: unknown[] }).cards
				: [];
		return Promise.all(rows.map(mapSupervisorCard));
	}

	async function issueSupervisorCards(regenerate = false) {
		if (regenerate && !confirm('Reset ulang semua QR+PIN kartu pengawas kegiatan ini? PIN lama tidak berlaku.')) return;
		issueBusy = true;
		try {
			const response = await fetch(clientApiPath`/api/asesmen/events/${eventId}/supervisor-access-cards/issue`, {
				method: 'POST',
				headers: { 'content-type': 'application/json' },
				body: JSON.stringify({ regenerate, expires_hours: 0 })
			});
			const payload = await readClientApiData<unknown>(response, 'Gagal menerbitkan kartu pengawas QR+PIN');
			const rows = Array.isArray(payload)
				? payload
				: payload && typeof payload === 'object' && Array.isArray((payload as { cards?: unknown[] }).cards)
					? (payload as { cards: unknown[] }).cards
					: [];
			const [schoolProfile, eventInfo, cards] = await Promise.all([fetchSchoolProfile(), fetchEventInfo(), Promise.all(rows.map(mapSupervisorCard))]);
			cardsPromise = Promise.resolve({ schoolProfile, eventInfo, cards });
		} finally {
			issueBusy = false;
		}
	}

	async function mapSupervisorCard(row: unknown): Promise<SupervisorCard> {
		const card = row as Record<string, unknown>;
		const token = String(card.token ?? '');
		const qrPath = String(card.qr_path ?? '');
		const portalUrl = absoluteUrl(qrPath || (token ? `/pengawas-ujian?card=${encodeURIComponent(token)}` : `/pengawas-ujian`));
		return {
			id: String(card.room_proctor_id ?? card.card_id ?? `${card.session_id}-${card.room_id}`),
			eventTitle: String(card.event_title ?? 'Kegiatan Ujian'),
			sessionId: String(card.session_id ?? ''),
			sessionTitle: String(card.session_title ?? 'Sesi CBT'),
			scheduledStart: String(card.scheduled_start ?? ''),
			packageTitle: String(card.package_title ?? ''),
			roomId: String(card.room_id ?? ''),
			roomName: String(card.room_name ?? 'Ruang'),
			roomToken: '',
			pin: String(card.pin ?? ''),
			portalUrl,
			qrDataUrl: token ? await QRCode.toDataURL(portalUrl, { margin: 1, width: 160 }) : '',
			proctorName: String(card.employee_name ?? 'Pengawas Ruang'),
			proctorRole: String(card.proctor_role ?? 'utama'),
			participantCount: 0,
			status: String(card.status ?? 'not_issued'),
			token,
			qrPath
		};
	}

	async function fetchEventInfo() {
		try {
			const response = await fetch(clientApiPath`/api/asesmen/events/${eventId}`);
			return await readClientApiData<EventInfo>(response, 'Gagal memuat kegiatan');
		} catch (error) {
			console.warn('Kegiatan kartu pengawas tidak termuat', error);
			return null;
		}
	}

	async function fetchEventSessions() {
		const response = await fetch(clientApiPath`/api/asesmen/events/${eventId}/sessions`);
		const payload = await readClientApiData<unknown>(response, 'Gagal memuat sesi kegiatan');
		if (Array.isArray(payload)) return payload as EventSession[];
		if (payload && typeof payload === 'object' && Array.isArray((payload as { sessions?: unknown[] }).sessions)) return (payload as { sessions: EventSession[] }).sessions;
		return [];
	}

	async function fetchSessionRooms(sessionId: string) {
		const response = await fetch(clientApiPath`/api/asesmen/sessions/${sessionId}/rooms`);
		const payload = await readClientApiData<unknown>(response, 'Gagal memuat ruang sesi');
		if (Array.isArray(payload)) return payload as SessionRoom[];
		if (payload && typeof payload === 'object' && Array.isArray((payload as { rooms?: unknown[] }).rooms)) return (payload as { rooms: SessionRoom[] }).rooms;
		return [];
	}

	async function fetchRoomProctors(sessionId: string, roomId: string) {
		try {
			const response = await fetch(clientApiPath`/api/asesmen/sessions/${sessionId}/rooms/${roomId}/proctors`);
			const payload = await readClientApiData<unknown>(response, 'Gagal memuat pengawas ruang');
			if (Array.isArray(payload)) return payload as RoomProctor[];
			if (payload && typeof payload === 'object' && Array.isArray((payload as { proctors?: unknown[] }).proctors)) return (payload as { proctors: RoomProctor[] }).proctors;
		} catch (error) {
			console.warn('Pengawas ruang tidak termuat', error);
		}
		return [];
	}

	function absoluteUrl(path: string) {
		if (typeof window === 'undefined') return path;
		return new URL(path, window.location.origin).toString();
	}

	function supervisorPin(roomToken: string | undefined) {
		const compact = (roomToken ?? '').replace(/\D/g, '');
		if (compact.length >= 4) return compact.slice(-4);
		const raw = (roomToken ?? '').replace(/[^A-Za-z0-9]/g, '').toUpperCase();
		return raw.length >= 4 ? raw.slice(-4) : 'MINTA ADMIN';
	}

	function roleLabel(role: string) {
		const labels: Record<string, string> = { utama: 'Pengawas utama', pendamping: 'Pengawas pendamping', cadangan: 'Pengawas cadangan' };
		return labels[role] ?? role;
	}

	function fmtDt(value: string) {
		if (!value) return '—';
		return new Date(value).toLocaleString('id-ID', { timeZone: 'Asia/Makassar', year: 'numeric', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' }) + ' WITA';
	}

	function errorMessage(error: unknown) {
		return error instanceof Error && error.message.trim() ? error.message : 'Kartu pengawas belum dapat dimuat.';
	}
</script>

<svelte:head>
	<title>Kartu Pengawas Ujian</title>
</svelte:head>

<AsyncContent promise={cardsPromise}>
	{#snippet pending()}
		<div class="mx-auto max-w-7xl space-y-4 p-6">
			<Skeleton class="h-8 w-72" />
			<div class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
				{#each Array.from({ length: 6 }) as _, index (`supervisor-card-skeleton-${index}`)}
					<Skeleton class="h-80 rounded-lg" />
				{/each}
			</div>
		</div>
	{/snippet}

	{#snippet failed(error, reset)}
		<div class="mx-auto max-w-7xl p-6">
			<RecoveryPanel title="Kartu Pengawas Belum Tersaji" message={errorMessage(error)} onRetry={() => { reset?.(); loadCards(); }} />
		</div>
	{/snippet}

	{#snippet children(value)}
		{@const data = value as PrintData}
		<div class="mx-auto max-w-7xl space-y-6 p-6 print:p-0">
			<div class="flex flex-col gap-3 print:hidden md:flex-row md:items-center md:justify-between">
				<div>
					<h1 class="text-2xl font-semibold text-foreground">Kartu Pengawas Ujian</h1>
					<p class="text-sm text-muted-foreground">Cetak QR Portal Pengawasan + PIN singkat untuk pengawas ruang.</p>
				</div>
				<div class="flex flex-wrap gap-2">
					<Button href={resolve(`/asesmen/kegiatan/${eventId}`)} variant="outline">Kembali ke Kegiatan</Button>
					<Button variant="outline" onclick={() => issueSupervisorCards(false)} disabled={issueBusy}>Terbitkan QR+PIN</Button>
					<Button variant="outline" onclick={() => issueSupervisorCards(true)} disabled={issueBusy}>Reset QR+PIN</Button>
					<Button onclick={() => window.print()} disabled={data.cards.length === 0 || issueBusy}><PrinterIcon class="mr-2 size-4" />Cetak</Button>
				</div>
			</div>

			{#if data.cards.length === 0}
				<div class="rounded-xl border border-warning/30 bg-warning/10 p-4 text-sm text-warning print:hidden">
					<p class="font-semibold">Belum ada kartu pengawas.</p>
					<p class="mt-1">Pastikan kegiatan memiliki sesi, ruang, dan penugasan pengawas. Jika pengawas belum ditugaskan, kartu placeholder ruang tidak dibuat karena belum ada ruang.</p>
				</div>
			{:else}
				<div class="rounded-xl border border-primary/20 bg-primary/10 p-4 text-sm text-primary print:hidden">
					{data.cards.length} kartu pengawas termuat. Jika PIN/QR belum tampil, klik Terbitkan QR+PIN lalu langsung cetak hasil terbitan.
				</div>
			{/if}

			<div class="grid gap-4 md:grid-cols-2 xl:grid-cols-3 print:grid-cols-2">
				{#each data.cards as card (card.id)}
					<article class="break-inside-avoid rounded-lg border border-primary/20 bg-card p-5 shadow-sm print:shadow-none">
						<div class="border-b border-dashed border-primary/20 pb-3">
							<p class="text-xs font-semibold uppercase tracking-[0.16em] text-primary">{data.schoolProfile.ministry_line}</p>
							<p class="mt-1 text-sm font-semibold uppercase text-foreground">{data.schoolProfile.name}</p>
							<p class="mt-1 text-[11px] leading-4 text-muted-foreground">{schoolAddressLine(data.schoolProfile) || data.schoolProfile.office_line}</p>
							<h2 class="mt-2 text-lg font-semibold text-foreground">KARTU PENGAWAS UJIAN</h2>
							<p class="text-sm text-muted-foreground">{card.eventTitle}</p>
						</div>

						<div class="mt-4 grid gap-4 sm:grid-cols-[1fr_150px]">
							<div class="space-y-1.5 text-sm text-foreground">
								<p><span class="font-medium">Nama Pengawas:</span> {card.proctorName}</p>
								<p><span class="font-medium">Tugas:</span> {roleLabel(card.proctorRole)}</p>
								<p><span class="font-medium">Ruang:</span> {card.roomName}</p>
								<p><span class="font-medium">Sesi:</span> {card.sessionTitle}</p>
								<p><span class="font-medium">Mapel/Paket:</span> {card.packageTitle || '—'}</p>
								<p><span class="font-medium">Tanggal:</span> {fmtDt(card.scheduledStart)}</p>
								<p><span class="font-medium">Peserta:</span> {card.participantCount}</p>
							</div>
							<div class="text-center">
								<img class="mx-auto size-36 rounded border border-border bg-white p-1" src={card.qrDataUrl} alt={`QR Portal Pengawasan ${card.roomName}`} />
								<Badge variant="outline" class="mt-2 border-primary/20 bg-primary/10 text-primary">Portal Pengawasan</Badge>
							</div>
						</div>

						<div class="mt-4 rounded-md bg-primary/10 px-4 py-3">
							<p class="text-xs uppercase tracking-[0.2em] text-primary">PIN Pengawas</p>
							<p class="mt-1 font-mono text-2xl font-bold text-primary">{card.pin}</p>
						</div>

						<ol class="mt-4 list-decimal space-y-1 pl-5 text-xs leading-5 text-muted-foreground">
							<li>Scan QR untuk membuka Portal Pengawasan ruang.</li>
							<li>Tekan tombol besar <span class="font-semibold text-foreground">Mulai Ujian</span> saat peserta siap.</li>
							<li>Jika muncul merah/masalah, tekan <span class="font-semibold text-foreground">Hubungi Admin</span>.</li>
						</ol>
					</article>
				{/each}
			</div>
		</div>
	{/snippet}
</AsyncContent>
