<script lang="ts">
	import PrinterIcon from '@lucide/svelte/icons/printer';
	import RefreshCwIcon from '@lucide/svelte/icons/refresh-cw';
	import QrCodeIcon from '@lucide/svelte/icons/qr-code';
	import { onMount } from 'svelte';
	import QRCode from 'qrcode';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import MicroActionTable from '$lib/components/ops/MicroActionTable.svelte';
	import { AssessmentPhaseHeader } from '$lib/components/asesmen';
	import { fetchSchoolProfile, schoolAddressLine, type SchoolProfile } from '$lib/school-profile';
	import { clientApiPath, readClientApiData } from '$lib/client/api';
	import {
		cardPin,
		cardPortalToken,
		cardReadinessIssues,
		cardReady,
		displayClass,
		filterExamCards,
		maskedToken,
		sortExamCards,
		studentName,
		summarizeExamCards,
		uniqueSorted,
		type ExamCardFilters,
		type ExamCardLike,
		type ExamCardSortMode
	} from '$lib/asesmen/exam-card-print';

	type ExamCard = ExamCardLike & {
		event_id: string;
		session_id: string;
		session_title: string;
		scheduled_start: string;
		participant_id: string;
		seat_no: number | null;
		nis: string;
		class_code: string;
		room_name: string;
	};
	type PrintLayout = 'a4-8' | 'a4-4' | 'single';

	const eventId = page.params.id ?? '';
	const cardColumns = [
		{ key: 'student', label: 'Peserta', class: 'min-w-[14rem]' },
		{ key: 'session', label: 'Sesi' },
		{ key: 'room', label: 'Ruang/Meja' },
		{ key: 'token', label: 'QR/PIN' },
		{ key: 'status', label: 'Status' }
	];

	let loading = $state(true);
	let loadError = $state<unknown>(null);
	let schoolProfile = $state<SchoolProfile | null>(null);
	let cards = $state<ExamCard[]>([]);
	let qrDataUrls = $state<Record<string, string>>({});
	let issueBusy = $state(false);
	let search = $state('');
	let sessionId = $state('');
	let classCode = $state('');
	let roomName = $state('');
	let statusFilter = $state<'all' | 'ready' | 'needs_check'>('all');
	let sortMode = $state<ExamCardSortMode>('room-seat-name');
	let printLayout = $state<PrintLayout>('a4-8');
	let qrGenerationRun = 0;

	let sessions = $derived(uniqueSorted(cards.map((card) => card.session_id ? `${card.session_id}|||${card.session_title || 'Sesi'}` : '').filter(Boolean)));
	let classes = $derived(uniqueSorted(cards.map(displayClass)));
	let rooms = $derived(uniqueSorted(cards.map((card) => card.room_name)));
	let filteredCards = $derived(sortExamCards(filterExamCards(cards, currentFilters()), sortMode));
	let allSummary = $derived(summarizeExamCards(cards));
	let filteredSummary = $derived(summarizeExamCards(filteredCards));
	let readinessIssues = $derived(cardReadinessIssues(filteredCards));
	let printDisabled = $derived(filteredCards.length === 0 || readinessIssues.length > 0 || issueBusy);

	function currentFilters(): ExamCardFilters {
		return { sessionId, classCode, roomName, status: statusFilter, search };
	}

	async function fetchCardRows() {
		const response = await fetch(clientApiPath`/api/asesmen/events/${eventId}/exam-access-cards`);
		const payload = await readClientApiData<unknown>(response, 'Gagal memuat kartu ujian QR+PIN');
		if (Array.isArray(payload)) return payload as ExamCard[];
		if (payload && typeof payload === 'object' && Array.isArray((payload as { cards?: unknown[] }).cards)) return (payload as { cards: ExamCard[] }).cards;
		return [];
	}

	async function loadCards() {
		loading = true;
		loadError = null;
		try {
			const [profile, rows] = await Promise.all([fetchSchoolProfile(), fetchCardRows()]);
			schoolProfile = profile;
			cards = rows;
		} catch (error) {
			loadError = error;
		} finally {
			loading = false;
		}
	}

	async function issueCards(regenerate = false) {
		if (regenerate && !confirm('Reset ulang semua QR+PIN kartu peserta kegiatan ini? PIN lama tidak berlaku.')) return;
		issueBusy = true;
		try {
			const response = await fetch(clientApiPath`/api/asesmen/events/${eventId}/exam-access-cards/issue`, {
				method: 'POST',
				headers: { 'content-type': 'application/json' },
				body: JSON.stringify({ regenerate, expires_hours: 0 })
			});
			const payload = await readClientApiData<unknown>(response, 'Gagal menerbitkan kartu ujian QR+PIN');
			if (Array.isArray(payload)) cards = payload as ExamCard[];
			else if (payload && typeof payload === 'object' && Array.isArray((payload as { cards?: unknown[] }).cards)) cards = (payload as { cards: ExamCard[] }).cards;
			else await loadCards();
		} finally {
			issueBusy = false;
		}
	}

	function resetFilters() {
		search = '';
		sessionId = '';
		classCode = '';
		roomName = '';
		statusFilter = 'all';
		sortMode = 'room-seat-name';
	}

	function printCards() {
		if (printDisabled) return;
		window.print();
	}

	function cardPortalUrl(card: ExamCard) {
		const token = cardPortalToken(card);
		if (card.portal_url) return card.portal_url;
		if (card.qr_path) return `${window.location.origin}${card.qr_path.startsWith('/') ? card.qr_path : `/${card.qr_path}`}`;
		if (!token) return '';
		return `${window.location.origin}/ujian?card=${encodeURIComponent(token)}`;
	}

	function fmtDt(value: string | undefined) {
		if (!value) return '—';
		return new Date(value).toLocaleString('id-ID', { timeZone: 'Asia/Makassar', year: 'numeric', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' }) + ' WITA';
	}

	function layoutClass(layout: PrintLayout) {
		if (layout === 'single') return 'print:grid-cols-1';
		if (layout === 'a4-4') return 'print:grid-cols-2';
		return 'print:grid-cols-2';
	}

	function cardPrintClass(layout: PrintLayout) {
		if (layout === 'single') return 'print:min-h-[260mm] print:p-8';
		if (layout === 'a4-4') return 'print:min-h-[130mm]';
		return 'print:min-h-[92mm] print:p-4';
	}

	function cardsErrorMessage(error: unknown) {
		return error instanceof Error && error.message.trim() ? error.message : 'Gagal memuat kartu ujian';
	}

	async function generateQrDataUrls(currentCards: ExamCard[], runId: number) {
		const entries = await Promise.all(
			currentCards.map(async (card) => {
				const value = cardPortalUrl(card);
				if (!value) return [card.participant_id, ''] as const;
				try {
					const url = await QRCode.toDataURL(value, { errorCorrectionLevel: 'M', margin: 1, width: 192, color: { dark: '#052e16', light: '#ffffff' } });
					return [card.participant_id, url] as const;
				} catch {
					return [card.participant_id, ''] as const;
				}
			})
		);
		if (runId === qrGenerationRun) qrDataUrls = Object.fromEntries(entries);
	}

	onMount(() => {
		void loadCards();
	});

	$effect(() => {
		if (loading) return;
		const runId = ++qrGenerationRun;
		qrDataUrls = {};
		void generateQrDataUrls(filteredCards, runId);
	});
</script>

<svelte:head>
	<title>Kartu Peserta Ujian</title>
</svelte:head>

{#if loading}
	<div class="mx-auto max-w-7xl space-y-6 p-6">
		<Skeleton class="h-8 w-64" />
		<Skeleton class="h-28 rounded-xl" />
		<Skeleton class="h-72 rounded-xl" />
	</div>
{:else if loadError}
	<div class="mx-auto max-w-7xl p-6">
		<RecoveryPanel title="Kartu Ujian Belum Tersaji" message={cardsErrorMessage(loadError)} onRetry={loadCards} />
	</div>
{:else}
	<div class="mx-auto max-w-7xl space-y-6 p-6 print:max-w-none print:p-0">
		<div class="print:hidden">
			<AssessmentPhaseHeader
				code="7.4.0.a"
				badge="Kartu peserta"
				title="Kartu Peserta Ujian"
				description="Cetak massal kartu peserta per kegiatan dengan filter kelas, sesi, ruang, dan status kesiapan."
				primaryAction={{ label: 'Dokumen & Cetak', href: resolve(`/asesmen/kegiatan/${eventId}/cetak`), variant: 'outline' }}
				secondaryActions={[{ label: '7.4 Arsip', href: resolve(`/asesmen/kegiatan/${eventId}/archive`), variant: 'ghost' }]}
			/>
		</div>
		<div class="flex flex-wrap justify-end gap-2 print:hidden">
			<Button variant="outline" onclick={loadCards}><RefreshCwIcon class="mr-2 size-4" />Refresh</Button>
			<Button variant="outline" onclick={() => issueCards(false)} disabled={issueBusy}><QrCodeIcon class="mr-2 size-4" />Terbitkan QR+PIN</Button>
			<Button onclick={printCards} disabled={printDisabled}><PrinterIcon class="mr-2 size-4" />Cetak Massal</Button>
		</div>

		<div class="grid gap-3 print:hidden md:grid-cols-4">
			<div class="rounded-xl border border-primary/20 bg-primary/10 p-4">
				<p class="text-xs font-medium uppercase tracking-wide text-primary">Total kartu</p>
				<p class="mt-1 text-2xl font-semibold text-primary">{allSummary.total}</p>
			</div>
			<div class="rounded-xl border border-success/20 bg-success/10 p-4">
				<p class="text-xs font-medium uppercase tracking-wide text-success">Siap cetak</p>
				<p class="mt-1 text-2xl font-semibold text-success">{allSummary.ready}</p>
			</div>
			<div class="rounded-xl border border-warning/30 bg-warning/10 p-4">
				<p class="text-xs font-medium uppercase tracking-wide text-warning">Perlu cek</p>
				<p class="mt-1 text-2xl font-semibold text-warning">{allSummary.needsCheck}</p>
			</div>
			<div class="rounded-xl border border-border bg-card p-4">
				<p class="text-xs font-medium uppercase tracking-wide text-muted-foreground">Terpilih</p>
				<p class="mt-1 text-2xl font-semibold text-foreground">{filteredSummary.total}</p>
			</div>
		</div>

		<div class="space-y-4 rounded-2xl border bg-muted/30 p-4 print:hidden">
			<div class="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
				<div>
					<h2 class="font-semibold text-foreground">Filter cetak massal</h2>
					<p class="text-sm text-muted-foreground">Default distribusi: urut per ruang, nomor meja, lalu nama peserta.</p>
				</div>
				<Button variant="outline" size="sm" onclick={resetFilters}>Reset Filter</Button>
			</div>
			<div class="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
				<label class="space-y-1.5">
					<span class="text-xs font-medium text-muted-foreground">Cari peserta/NIS</span>
					<input class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" placeholder="Nama, NIS, kelas, ruang..." bind:value={search} />
				</label>
				<label class="space-y-1.5">
					<span class="text-xs font-medium text-muted-foreground">Sesi</span>
					<select class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={sessionId}>
						<option value="">Semua sesi</option>
						{#each sessions as item}
							{@const parts = item.split('|||')}
							<option value={parts[0]}>{parts[1]}</option>
						{/each}
					</select>
				</label>
				<label class="space-y-1.5">
					<span class="text-xs font-medium text-muted-foreground">Kelas</span>
					<select class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={classCode}>
						<option value="">Semua kelas</option>
						{#each classes as item}<option value={item}>{item}</option>{/each}
					</select>
				</label>
				<label class="space-y-1.5">
					<span class="text-xs font-medium text-muted-foreground">Ruang</span>
					<select class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={roomName}>
						<option value="">Semua ruang</option>
						{#each rooms as item}<option value={item}>{item}</option>{/each}
					</select>
				</label>
			</div>
			<div class="grid gap-3 md:grid-cols-3">
				<label class="space-y-1.5">
					<span class="text-xs font-medium text-muted-foreground">Status</span>
					<select class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={statusFilter}>
						<option value="all">Semua status</option>
						<option value="ready">Hanya siap cetak</option>
						<option value="needs_check">Perlu cek</option>
					</select>
				</label>
				<label class="space-y-1.5">
					<span class="text-xs font-medium text-muted-foreground">Urutan</span>
					<select class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={sortMode}>
						<option value="room-seat-name">Ruang → Meja → Nama</option>
						<option value="session-room-seat">Sesi → Ruang → Meja</option>
						<option value="class-name">Kelas → Nama</option>
						<option value="name">Nama peserta</option>
					</select>
				</label>
				<label class="space-y-1.5">
					<span class="text-xs font-medium text-muted-foreground">Layout kertas</span>
					<select class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={printLayout}>
						<option value="a4-8">A4 isi 8 kartu</option>
						<option value="a4-4">A4 isi 4 kartu besar</option>
						<option value="single">1 peserta per halaman</option>
					</select>
				</label>
			</div>
		</div>

		{#if cards.length === 0}
			<div class="rounded-xl border border-dashed border-warning/30 bg-warning/10 px-4 py-5 text-sm text-warning print:hidden">
				<p class="font-semibold">Belum ada kartu ujian untuk dicetak.</p>
				<p class="mt-1">Daftarkan peserta, terbitkan kartu QR+PIN, lalu lengkapi ruangan dan nomor meja pada sesi kegiatan sebelum cetak massal.</p>
			</div>
		{:else if readinessIssues.length > 0}
			<div class="rounded-xl border border-warning/30 bg-warning/10 px-4 py-3 text-sm text-warning print:hidden" role="alert">
				<p class="font-semibold">Kartu terpilih belum siap dicetak massal.</p>
				<p class="mt-1">{readinessIssues.join(', ')}. Rapikan QR/PIN, ruang, dan kursi dari detail sesi sebelum cetak final.</p>
				<div class="mt-3 flex flex-wrap gap-2">
					<a href={resolve(`/asesmen/sesi?event_id=${eventId}&readiness=not_ready`)} class="inline-flex rounded-md border border-warning/30 bg-card px-3 py-2 text-sm font-semibold text-warning hover:bg-warning/15">Cek sesi dan ruang</a>
					<Button variant="outline" size="sm" onclick={() => issueCards(false)} disabled={issueBusy}>Terbitkan QR+PIN</Button>
				</div>
			</div>
		{:else}
			<div class="rounded-xl border border-success/20 bg-success/10 px-4 py-3 text-sm text-success print:hidden">
				{filteredCards.length} kartu terpilih siap dicetak. Cetak hanya saat distribusi kartu siap.
			</div>
		{/if}

		<MicroActionTable title="7.4.0.a Status kartu peserta" description={`Menampilkan ${filteredCards.length} dari ${cards.length} kartu. Kode mentah disamarkan pada layar.`} columns={cardColumns} rows={filteredCards} rowKey={(row) => (row as ExamCard).participant_id} tableClass="min-w-[820px]" class="print:hidden" emptyTitle="Tidak ada kartu sesuai filter.">
			{#snippet cell(row, column)}
				{@const card = row as ExamCard}
				{#if column.key === 'student'}
					<div class="font-semibold text-foreground">{studentName(card)}</div>
					<div class="text-xs text-muted-foreground">NIS {card.nis || '—'} · {displayClass(card)}</div>
				{:else if column.key === 'session'}
					<div class="font-medium text-foreground">{card.session_title}</div>
					<div class="text-xs text-muted-foreground">{fmtDt(card.scheduled_start)}</div>
				{:else if column.key === 'room'}
					<span>{card.room_name || 'Belum ditentukan'} · Meja {card.seat_no ?? '—'}</span>
				{:else if column.key === 'token'}
					<div class="space-y-1">
						<div class="font-mono text-xs font-semibold text-muted-foreground">QR {maskedToken(cardPortalToken(card))}</div>
						<div class="text-xs text-muted-foreground">PIN {cardPin(card) || 'Belum ada'}</div>
					</div>
				{:else}
					<Badge variant="outline" class={cardReady(card) ? 'border-success/20 bg-success/10 text-success' : 'border-warning/30 bg-warning/10 text-warning'}>{cardReady(card) ? 'Siap cetak' : 'Perlu cek'}</Badge>
				{/if}
			{/snippet}
		</MicroActionTable>

		{#if schoolProfile}
			<div class={`hidden gap-4 ${layoutClass(printLayout)} print:grid`}>
				{#each filteredCards as card (card.participant_id)}
					<article class={`break-inside-avoid rounded-lg border border-primary/20 bg-card p-5 shadow-sm print:shadow-none ${cardPrintClass(printLayout)}`}>
						<div class="border-b border-dashed border-primary/20 pb-3">
							<p class="text-xs font-semibold uppercase tracking-[0.16em] text-primary">{schoolProfile.ministry_line}</p>
							<p class="mt-1 text-sm font-semibold uppercase text-foreground">{schoolProfile.name}</p>
							<p class="mt-1 text-[11px] leading-4 text-muted-foreground">{schoolAddressLine(schoolProfile) || schoolProfile.office_line}</p>
							<h2 class="mt-2 text-lg font-semibold text-foreground">KARTU PESERTA UJIAN</h2>
							<p class="text-sm font-medium text-foreground">{card.event_title || 'Kegiatan Ujian'}</p>
							<p class="text-sm text-muted-foreground">{card.session_title}</p>
						</div>
						<div class="mt-4 space-y-2 text-sm text-foreground">
							<p><span class="font-medium">Nama:</span> {studentName(card)}</p>
							<p><span class="font-medium">Nomor Peserta/NIS:</span> {card.nis || '—'}</p>
							<p><span class="font-medium">Kelas:</span> {displayClass(card)}</p>
							<p><span class="font-medium">Ruang:</span> {card.room_name || 'Belum ditentukan'}</p>
							<p><span class="font-medium">No Meja:</span> {card.seat_no ?? '—'}</p>
							<p><span class="font-medium">Jadwal:</span> {fmtDt(card.scheduled_start)}</p>
						</div>
						<div class="mt-5 grid grid-cols-[116px_1fr] gap-4 rounded-md bg-primary/10 px-4 py-3">
							<div class="grid size-28 place-items-center rounded-md bg-white p-2">
								{#if qrDataUrls[card.participant_id]}
									<img src={qrDataUrls[card.participant_id]} alt={`QR masuk ujian ${studentName(card)}`} class="size-24" />
								{:else}
									<span class="text-center text-[10px] font-semibold text-destructive">QR belum tersedia</span>
								{/if}
							</div>
							<div>
								<p class="text-xs uppercase tracking-[0.2em] text-primary">QR Masuk Ujian</p>
								<p class="mt-1 text-sm text-primary">Scan QR, lalu masukkan PIN:</p>
								<p class="mt-1 font-mono text-3xl font-bold tracking-[0.18em] text-primary">{cardPin(card) || '—'}</p>
								<p class="mt-2 text-[10px] text-primary">Kode kartu: {maskedToken(cardPortalToken(card))}. Token mentah tidak dicetak sebagai informasi utama.</p>
							</div>
						</div>
						<div class="mt-3 rounded-md border border-dashed border-primary/20 p-3 text-xs leading-5 text-foreground">
							<p class="font-semibold">Cara masuk:</p>
							<ol class="ml-4 list-decimal">
								<li>Scan QR pada kartu ini.</li>
								<li>Masukkan PIN di atas.</li>
								<li>Pastikan identitas benar, lalu tekan Masuk Ujian.</li>
							</ol>
						</div>
					</article>
				{/each}
			</div>
		{/if}
	</div>
{/if}
