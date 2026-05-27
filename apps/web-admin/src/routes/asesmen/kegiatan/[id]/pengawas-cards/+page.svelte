<script lang="ts">
	import PrinterIcon from '@lucide/svelte/icons/printer';
	import RefreshCwIcon from '@lucide/svelte/icons/refresh-cw';
	import QrCodeIcon from '@lucide/svelte/icons/qr-code';
	import { onMount } from 'svelte';
	import QRCode from 'qrcode';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { clientApiPath, readClientApiData } from '$lib/client/api';
	import { fetchSchoolProfile, schoolAddressLine, type SchoolProfile } from '$lib/school-profile';

	type EventInfo = { id: string; title: string; exam_type?: string; academic_year_name?: string; status?: string };
	type SupervisorCard = {
		id: string;
		eventTitle: string;
		sessionId: string;
		sessionTitle: string;
		scheduledStart: string;
		packageTitle: string;
		roomId: string;
		roomName: string;
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

	const eventId = page.params.id ?? '';
	let loading = $state(true);
	let loadError = $state<unknown>(null);
	let schoolProfile = $state<SchoolProfile | null>(null);
	let eventInfo = $state<EventInfo | null>(null);
	let cards = $state<SupervisorCard[]>([]);
	let issueBusy = $state(false);
	let sessionId = $state('');
	let roomName = $state('');
	let statusFilter = $state<'all' | 'ready' | 'needs_check'>('all');
	let search = $state('');

	let sessions = $derived(uniqueSorted(cards.map((card) => card.sessionId ? `${card.sessionId}|||${card.sessionTitle || 'Sesi'}` : '').filter(Boolean)));
	let rooms = $derived(uniqueSorted(cards.map((card) => card.roomName)));
	let filteredCards = $derived(sortSupervisorCards(cards.filter((card) => {
		if (sessionId && card.sessionId !== sessionId) return false;
		if (roomName && card.roomName !== roomName) return false;
		if (statusFilter === 'ready' && !supervisorReady(card)) return false;
		if (statusFilter === 'needs_check' && supervisorReady(card)) return false;
		const keyword = search.trim().toLocaleLowerCase('id-ID');
		if (keyword) {
			const haystack = [card.roomName, card.sessionTitle, card.packageTitle, card.proctorName].join(' ').toLocaleLowerCase('id-ID');
			if (!haystack.includes(keyword)) return false;
		}
		return true;
	})));
	let readyCount = $derived(cards.filter(supervisorReady).length);
	let selectedReadyCount = $derived(filteredCards.filter(supervisorReady).length);
	let printDisabled = $derived(filteredCards.length === 0 || selectedReadyCount !== filteredCards.length || issueBusy);

	function uniqueSorted(values: string[]) {
		return [...new Set(values.map((value) => value.trim()).filter(Boolean))].sort((a, b) => a.localeCompare(b, 'id-ID', { numeric: true, sensitivity: 'base' }));
	}

	function sortSupervisorCards(items: SupervisorCard[]) {
		return [...items].sort((a, b) => a.sessionTitle.localeCompare(b.sessionTitle, 'id-ID', { numeric: true }) || a.roomName.localeCompare(b.roomName, 'id-ID', { numeric: true }));
	}

	function supervisorReady(card: SupervisorCard) {
		return Boolean(card.token && card.pin && card.roomName);
	}

	async function fetchSupervisorCards() {
		const response = await fetch(clientApiPath`/api/asesmen/events/${eventId}/supervisor-access-cards`);
		const payload = await readClientApiData<unknown>(response, 'Gagal memuat lembar pengawas ruang QR+PIN');
		const rows = Array.isArray(payload)
			? payload
			: payload && typeof payload === 'object' && Array.isArray((payload as { cards?: unknown[] }).cards)
				? (payload as { cards: unknown[] }).cards
				: [];
		return Promise.all(rows.map(mapSupervisorCard));
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

	async function loadCards() {
		loading = true;
		loadError = null;
		try {
			const [profile, info, rows] = await Promise.all([fetchSchoolProfile(), fetchEventInfo(), fetchSupervisorCards()]);
			schoolProfile = profile;
			eventInfo = info;
			cards = rows;
		} catch (error) {
			loadError = error;
		} finally {
			loading = false;
		}
	}

	async function issueSupervisorCards(regenerate = false) {
		if (regenerate && !confirm('Reset ulang semua QR+PIN lembar pengawas ruang kegiatan ini? PIN lama tidak berlaku.')) return;
		issueBusy = true;
		try {
			const response = await fetch(clientApiPath`/api/asesmen/events/${eventId}/supervisor-access-cards/issue`, {
				method: 'POST',
				headers: { 'content-type': 'application/json' },
				body: JSON.stringify({ regenerate, expires_hours: 0 })
			});
			const payload = await readClientApiData<unknown>(response, 'Gagal menerbitkan lembar pengawas ruang QR+PIN');
			const rows = Array.isArray(payload)
				? payload
				: payload && typeof payload === 'object' && Array.isArray((payload as { cards?: unknown[] }).cards)
					? (payload as { cards: unknown[] }).cards
					: null;
			if (rows) cards = await Promise.all(rows.map(mapSupervisorCard));
			else await loadCards();
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
			pin: String(card.pin ?? ''),
			portalUrl,
			qrDataUrl: token ? await QRCode.toDataURL(portalUrl, { margin: 1, width: 160 }) : '',
			proctorName: String(card.employee_name ?? 'Lembar Pengawas Ruang'),
			proctorRole: String(card.proctor_role ?? 'pengawas_ruang'),
			participantCount: Number(card.participant_count ?? 0),
			status: String(card.status ?? 'not_issued'),
			token,
			qrPath
		};
	}

	function absoluteUrl(path: string) {
		if (typeof window === 'undefined') return path;
		return new URL(path, window.location.origin).toString();
	}

	function roleLabel(role: string) {
		const labels: Record<string, string> = { pengawas_ruang: 'Akses ruang fleksibel', utama: 'Pengawas utama', pendamping: 'Pengawas pendamping', cadangan: 'Pengawas cadangan' };
		return labels[role] ?? role;
	}

	function fmtDt(value: string) {
		if (!value) return '—';
		return new Date(value).toLocaleString('id-ID', { timeZone: 'Asia/Makassar', year: 'numeric', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' }) + ' WITA';
	}

	function resetFilters() {
		sessionId = '';
		roomName = '';
		statusFilter = 'all';
		search = '';
	}

	function errorMessage(error: unknown) {
		return error instanceof Error && error.message.trim() ? error.message : 'Lembar pengawas ruang belum dapat dimuat.';
	}

	onMount(() => {
		void loadCards();
	});
</script>

<svelte:head>
	<title>Lembar Pengawas Ruang</title>
</svelte:head>

{#if loading}
	<div class="mx-auto max-w-7xl space-y-4 p-6">
		<Skeleton class="h-8 w-72" />
		<Skeleton class="h-28 rounded-xl" />
		<div class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
			{#each Array.from({ length: 6 }) as _, index (`supervisor-card-skeleton-${index}`)}<Skeleton class="h-80 rounded-lg" />{/each}
		</div>
	</div>
{:else if loadError}
	<div class="mx-auto max-w-7xl p-6">
		<RecoveryPanel title="Lembar Pengawas Ruang Belum Tersaji" message={errorMessage(loadError)} onRetry={loadCards} />
	</div>
{:else}
	<div class="mx-auto max-w-7xl space-y-6 p-6 print:p-0">
		<div class="flex flex-col gap-3 print:hidden md:flex-row md:items-center md:justify-between">
			<div>
				<p class="text-sm font-medium text-primary">Dokumen & Cetak</p>
				<h1 class="text-2xl font-semibold text-foreground">Lembar Pengawas Ruang</h1>
				<p class="text-sm text-muted-foreground">Cetak satu QR+PIN per ruang. Lembar ini tidak melekat ke nama pengawas sehingga guru pengganti tetap bisa masuk ruang yang sama.</p>
			</div>
			<div class="flex flex-wrap gap-2">
				<a href={resolve(`/asesmen/kegiatan/${eventId}/cetak`)} class="inline-flex items-center rounded-md border border-border bg-card px-3 py-2 text-sm font-semibold text-foreground hover:bg-muted">Dokumen & Cetak</a>
				<Button variant="outline" onclick={loadCards}><RefreshCwIcon class="mr-2 size-4" />Refresh</Button>
				<Button variant="outline" onclick={() => issueSupervisorCards(false)} disabled={issueBusy}><QrCodeIcon class="mr-2 size-4" />Terbitkan QR+PIN Ruang</Button>
				<Button variant="outline" onclick={() => issueSupervisorCards(true)} disabled={issueBusy}>Reset QR+PIN</Button>
				<Button onclick={() => window.print()} disabled={printDisabled}><PrinterIcon class="mr-2 size-4" />Cetak Semua Lembar</Button>
			</div>
		</div>

		<div class="grid gap-3 print:hidden md:grid-cols-3">
			<div class="rounded-xl border border-primary/20 bg-primary/10 p-4"><p class="text-xs font-medium uppercase tracking-wide text-primary">Total lembar</p><p class="mt-1 text-2xl font-semibold text-primary">{cards.length}</p></div>
			<div class="rounded-xl border border-success/20 bg-success/10 p-4"><p class="text-xs font-medium uppercase tracking-wide text-success">Siap cetak</p><p class="mt-1 text-2xl font-semibold text-success">{readyCount}</p></div>
			<div class="rounded-xl border border-border bg-card p-4"><p class="text-xs font-medium uppercase tracking-wide text-muted-foreground">Terpilih</p><p class="mt-1 text-2xl font-semibold text-foreground">{filteredCards.length}</p></div>
		</div>

		<div class="space-y-4 rounded-2xl border bg-muted/30 p-4 print:hidden">
			<div class="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
				<div><h2 class="font-semibold text-foreground">Filter lembar pengawas</h2><p class="text-sm text-muted-foreground">Gunakan bila hanya ingin mencetak sesi atau ruang tertentu.</p></div>
				<Button variant="outline" size="sm" onclick={resetFilters}>Reset Filter</Button>
			</div>
			<div class="grid gap-3 md:grid-cols-4">
				<label class="space-y-1.5"><span class="text-xs font-medium text-muted-foreground">Cari</span><input class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" placeholder="Ruang, sesi, pengawas..." bind:value={search} /></label>
				<label class="space-y-1.5"><span class="text-xs font-medium text-muted-foreground">Sesi</span><select class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={sessionId}><option value="">Semua sesi</option>{#each sessions as item}{@const parts = item.split('|||')}<option value={parts[0]}>{parts[1]}</option>{/each}</select></label>
				<label class="space-y-1.5"><span class="text-xs font-medium text-muted-foreground">Ruang</span><select class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={roomName}><option value="">Semua ruang</option>{#each rooms as item}<option value={item}>{item}</option>{/each}</select></label>
				<label class="space-y-1.5"><span class="text-xs font-medium text-muted-foreground">Status</span><select class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={statusFilter}><option value="all">Semua status</option><option value="ready">QR+PIN siap</option><option value="needs_check">Perlu cek</option></select></label>
			</div>
		</div>

		{#if cards.length === 0}
			<div class="rounded-xl border border-warning/30 bg-warning/10 p-4 text-sm text-warning print:hidden">
				<p class="font-semibold">Belum ada lembar pengawas ruang.</p>
				<p class="mt-1">Pastikan kegiatan memiliki sesi dan ruang. Lembar diterbitkan per ruang agar tetap fleksibel bila pengawas berhalangan.</p>
			</div>
		{:else if selectedReadyCount !== filteredCards.length}
			<div class="rounded-xl border border-warning/30 bg-warning/10 p-4 text-sm text-warning print:hidden" role="alert">
				<p class="font-semibold">Sebagian lembar terpilih belum punya QR/PIN.</p>
				<p class="mt-1">Klik Terbitkan QR+PIN Ruang, lalu cetak ulang hasil terbitan.</p>
			</div>
		{:else}
			<div class="rounded-xl border border-primary/20 bg-primary/10 p-4 text-sm text-primary print:hidden">
				{filteredCards.length} lembar pengawas ruang siap dicetak untuk {eventInfo?.title ?? 'kegiatan ini'}.
			</div>
		{/if}

		<div class="grid gap-4 md:grid-cols-2 xl:grid-cols-3 print:grid-cols-2">
			{#each filteredCards as card (card.id)}
				<article class="break-inside-avoid rounded-lg border border-primary/20 bg-card p-5 shadow-sm print:shadow-none">
					<div class="border-b border-dashed border-primary/20 pb-3">
						{#if schoolProfile}
							<p class="text-xs font-semibold uppercase tracking-[0.16em] text-primary">{schoolProfile.ministry_line}</p>
							<p class="mt-1 text-sm font-semibold uppercase text-foreground">{schoolProfile.name}</p>
							<p class="mt-1 text-[11px] leading-4 text-muted-foreground">{schoolAddressLine(schoolProfile) || schoolProfile.office_line}</p>
						{/if}
						<h2 class="mt-2 text-lg font-semibold text-foreground">LEMBAR PENGAWAS RUANG</h2>
						<p class="text-sm text-muted-foreground">{card.eventTitle}</p>
					</div>
					<div class="mt-4 grid gap-4 sm:grid-cols-[1fr_150px]">
						<div class="space-y-1.5 text-sm text-foreground">
							<p><span class="font-medium">Akses:</span> Lembar ruang, bukan akun personal</p>
							<p><span class="font-medium">Pengawas tercatat:</span> {card.proctorName}</p>
							<p><span class="font-medium">Tugas:</span> {roleLabel(card.proctorRole)}</p>
							<p><span class="font-medium">Ruang:</span> {card.roomName}</p>
							<p><span class="font-medium">Sesi:</span> {card.sessionTitle}</p>
							<p><span class="font-medium">Mapel/Paket:</span> {card.packageTitle || '—'}</p>
							<p><span class="font-medium">Tanggal:</span> {fmtDt(card.scheduledStart)}</p>
							<p><span class="font-medium">Peserta:</span> {card.participantCount}</p>
						</div>
						<div class="text-center">
							{#if card.qrDataUrl}<img class="mx-auto size-36 rounded border border-border bg-white p-1" src={card.qrDataUrl} alt={`QR Ruang Saya ${card.roomName}`} />{:else}<div class="mx-auto grid size-36 place-items-center rounded border border-warning/30 bg-warning/10 p-2 text-xs font-semibold text-warning">QR belum tersedia</div>{/if}
							<Badge variant="outline" class="mt-2 border-primary/20 bg-primary/10 text-primary">Ruang Saya</Badge>
						</div>
					</div>
					<div class="mt-4 rounded-md bg-primary/10 px-4 py-3">
						<p class="text-xs uppercase tracking-[0.2em] text-primary">PIN Ruang</p>
						<p class="mt-1 font-mono text-2xl font-bold text-primary">{card.pin || '—'}</p>
					</div>
					<ol class="mt-4 list-decimal space-y-1 pl-5 text-xs leading-5 text-muted-foreground">
						<li>Scan QR pada lembar ini untuk membuka Ruang Saya.</li>
						<li>Tekan tombol besar <span class="font-semibold text-foreground">Mulai Ujian</span> saat peserta siap.</li>
						<li>Jika muncul merah/masalah, tekan <span class="font-semibold text-foreground">Hubungi Admin</span>.</li>
					</ol>
				</article>
			{/each}
		</div>
	</div>
{/if}
