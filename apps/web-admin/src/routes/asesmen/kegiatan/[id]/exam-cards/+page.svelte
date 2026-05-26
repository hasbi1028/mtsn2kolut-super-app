<script lang="ts">
	import PrinterIcon from '@lucide/svelte/icons/printer';
	import { onMount } from 'svelte';
	import QRCode from 'qrcode';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import MicroActionTable from '$lib/components/ops/MicroActionTable.svelte';
	import { fetchSchoolProfile, schoolAddressLine, type SchoolProfile } from '$lib/school-profile';
	import { clientApiPath, readClientApiData } from '$lib/client/api';

	type ExamCard = {
		event_id: string;
		event_title?: string;
		exam_type?: string;
		event_scope?: string;
		academic_year_name?: string;
		session_id: string;
		session_title: string;
		scheduled_start: string;
		participant_id: string;
		token: string;
		card_token?: string;
		exam_card_token?: string;
		qr_token?: string;
		portal_url?: string;
		pin?: string;
		card_pin?: string;
		exam_card_pin?: string;
		seat_no: number | null;
		nis: string;
		student_nama?: string;
		student_name?: string;
		gender?: string;
		class_code: string;
		class_name?: string;
		room_name: string;
		status?: string;
		card_id?: string;
		qr_path?: string;
		package_title?: string;
	};
	type ExamCardPrintData = {
		schoolProfile: SchoolProfile;
		cards: ExamCard[];
	};

	const eventId = page.params.id ?? '';
	const cardColumns = [
		{ key: 'student', label: 'Peserta', class: 'min-w-[14rem]' },
		{ key: 'session', label: 'Sesi' },
		{ key: 'room', label: 'Ruang/Meja' },
		{ key: 'token', label: 'QR/PIN' },
		{ key: 'status', label: 'Status' },
	];
	let cardsPromise = $state<Promise<ExamCardPrintData> | null>(null);
	let qrDataUrls = $state<Record<string, string>>({});
	let issueBusy = $state(false);
	let qrGenerationRun = 0;

	async function fetchCardRows() {
		const response = await fetch(clientApiPath`/api/asesmen/events/${eventId}/exam-access-cards`);
		const payload = await readClientApiData<unknown>(response, 'Gagal memuat kartu ujian QR+PIN');
		if (Array.isArray(payload)) return payload as ExamCard[];
		if (payload && typeof payload === 'object' && Array.isArray((payload as { cards?: unknown[] }).cards)) return (payload as { cards: ExamCard[] }).cards;
		return [];
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
			let cards: ExamCard[] = [];
			if (Array.isArray(payload)) cards = payload as ExamCard[];
			else if (payload && typeof payload === 'object' && Array.isArray((payload as { cards?: unknown[] }).cards)) cards = (payload as { cards: ExamCard[] }).cards;
			cardsPromise = Promise.resolve({ schoolProfile: await fetchSchoolProfile(), cards });
		} finally {
			issueBusy = false;
		}
	}

	async function fetchCards(): Promise<ExamCardPrintData> {
		const [schoolProfile, cards] = await Promise.all([fetchSchoolProfile(), fetchCardRows()]);
		return { schoolProfile, cards };
	}

	function loadCards() {
		cardsPromise = fetchCards();
	}

	function retryCards(reset?: () => void) {
		reset?.();
		loadCards();
	}

	function cardsErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Gagal memuat kartu ujian';
	}

	function handleCardsRenderError(error: unknown) {
		console.error('Kartu ujian kegiatan asesmen belum dapat ditampilkan', error);
	}

	function fmtDt(value: string) {
		return new Date(value).toLocaleString('id-ID', {
			timeZone: 'Asia/Makassar',
			year: 'numeric',
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		}) + ' WITA';
	}

	function cardReadinessIssues(cards: ExamCard[]) {
		const missingToken = cards.filter((card) => !cardPortalToken(card)).length;
		const missingPin = cards.filter((card) => !cardPin(card)).length;
		const missingRoom = cards.filter((card) => !card.room_name).length;
		const missingSeat = cards.filter((card) => card.seat_no === null || card.seat_no === undefined).length;
		const issues: string[] = [];
		if (missingToken > 0) issues.push(`${missingToken} kartu belum punya QR/kode kartu`);
		if (missingPin > 0) issues.push(`${missingPin} kartu belum punya PIN`);
		if (missingRoom > 0) issues.push(`${missingRoom} peserta belum punya ruangan`);
		if (missingSeat > 0) issues.push(`${missingSeat} peserta belum punya nomor meja`);
		return issues;
	}

	function cardPortalToken(card: ExamCard) {
		return card.card_token || card.exam_card_token || card.qr_token || card.token || '';
	}

	function cardPin(card: ExamCard) {
		const explicit = card.pin || card.card_pin || card.exam_card_pin;
		if (explicit) return explicit;
		const token = cardPortalToken(card);
		return token.length >= 4 ? token.slice(-4).toUpperCase() : '';
	}

	function cardPortalUrl(card: ExamCard) {
		const token = cardPortalToken(card);
		if (card.portal_url) return card.portal_url;
		if (card.qr_path) {
			const origin = typeof window === 'undefined' ? '' : window.location.origin;
			return `${origin}${card.qr_path.startsWith('/') ? card.qr_path : `/${card.qr_path}`}`;
		}
		if (!token) return '';
		const origin = typeof window === 'undefined' ? '' : window.location.origin;
		return `${origin}/ujian?card=${encodeURIComponent(token)}`;
	}

	function maskedToken(token: string) {
		if (!token) return 'Belum ada';
		if (token.length <= 4) return '••••';
		return `${token.slice(0, 2)}••••${token.slice(-2)}`;
	}

	function cardReady(card: ExamCard) {
		return Boolean(cardPortalToken(card) && cardPin(card) && card.room_name && card.seat_no !== null && card.seat_no !== undefined);
	}

	function studentName(card: ExamCard) {
		return card.student_nama || card.student_name || 'Peserta';
	}

	function displayClass(card: ExamCard) {
		return card.class_code || card.class_name || 'Kelas belum ada';
	}

	async function generateQrDataUrls(cards: ExamCard[], runId: number) {
		const entries = await Promise.all(
			cards.map(async (card) => {
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
		const promise = cardsPromise;
		if (!promise) return;
		const runId = ++qrGenerationRun;
		qrDataUrls = {};
		promise.then((data) => generateQrDataUrls(data.cards, runId)).catch(() => {
			if (runId === qrGenerationRun) qrDataUrls = {};
		});
	});
</script>

<svelte:head>
	<title>Kartu Ujian Kegiatan</title>
</svelte:head>

<AsyncContent promise={cardsPromise} onerror={handleCardsRenderError}>
	{#snippet pending()}
		<div class="mx-auto max-w-7xl space-y-6 p-6">
			<div class="space-y-2">
				<Skeleton class="h-8 w-56" />
				<Skeleton class="h-4 w-80" />
			</div>
			<div class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
				{#each Array.from({ length: 6 }) as _, index (`exam-card-skeleton-${index}`)}
					<div class="rounded-lg border border-primary/20 bg-card p-5 shadow-sm">
						<div class="space-y-2 border-b border-dashed border-primary/20 pb-3">
							<Skeleton class="h-4 w-36" />
							<Skeleton class="h-6 w-48" />
							<Skeleton class="h-4 w-40" />
						</div>
						<div class="mt-4 space-y-2">
							<Skeleton class="h-4 w-44" />
							<Skeleton class="h-4 w-32" />
							<Skeleton class="h-4 w-28" />
							<Skeleton class="h-4 w-36" />
							<Skeleton class="h-4 w-24" />
							<Skeleton class="h-4 w-40" />
						</div>
						<div class="mt-5 rounded-md bg-primary/10 px-4 py-3">
							<Skeleton class="h-4 w-24" />
							<Skeleton class="mt-2 h-8 w-40" />
						</div>
					</div>
				{/each}
			</div>
		</div>
	{/snippet}

	{#snippet failed(error, reset)}
		<div class="mx-auto max-w-7xl p-6">
			<RecoveryPanel
				title="Kartu Ujian Belum Tersaji"
				message={cardsErrorMessage(error)}
				onRetry={() => retryCards(reset)}
			/>
		</div>
	{/snippet}

	{#snippet children(value)}
		{@const data = value as ExamCardPrintData}
		{@const currentCards = data.cards}
		{@const schoolProfile = data.schoolProfile}
		{@const readinessIssues = cardReadinessIssues(currentCards)}
		{@const printDisabled = currentCards.length === 0 || readinessIssues.length > 0}
	<div class="mx-auto max-w-7xl space-y-6 p-6 print:p-0">
		<div class="flex items-center justify-between print:hidden">
			<div>
				<h1 class="text-2xl font-semibold text-foreground">Kartu Ujian Kegiatan</h1>
				<p class="text-sm text-muted-foreground">Cetak per peserta dengan QR Portal Ujian, PIN pendek, ruangan, dan nomor meja.</p>
			</div>
			<div class="flex flex-wrap gap-2">
				<a href={resolve(`/asesmen/kegiatan/${eventId}`)} class="inline-flex items-center rounded-md border border-success/20 bg-success/10 px-3 py-2 text-sm font-semibold text-success hover:bg-success/15">Kembali ke Kegiatan</a>
			<Button onclick={() => window.print()} disabled={printDisabled}>
				<PrinterIcon class="mr-2 size-4" />
				Cetak
			</Button>
			</div>
		</div>

		{#if currentCards.length === 0}
			<div class="rounded-xl border border-dashed border-warning/30 bg-warning/10 px-4 py-5 text-sm text-warning print:hidden">
				<p class="font-semibold">Belum ada kartu ujian untuk dicetak.</p>
				<p class="mt-1">Daftarkan peserta, terbitkan kartu QR+PIN, lalu lengkapi ruangan dan nomor meja pada sesi kegiatan sebelum cetak massal.</p>
				<div class="mt-3 flex flex-wrap gap-2">
					<a href={resolve(`/asesmen/sesi?event_id=${eventId}&readiness=not_ready`)} class="inline-flex rounded-md border border-warning/30 bg-card px-3 py-2 text-sm font-semibold text-warning hover:bg-warning/15">Cek sesi kegiatan</a>
					<a href={resolve(`/asesmen/kegiatan/${eventId}/members`)} class="inline-flex rounded-md border border-warning/30 bg-card px-3 py-2 text-sm font-semibold text-warning hover:bg-warning/15">Cek penugasan/peserta</a>
				</div>
			</div>
		{:else if readinessIssues.length > 0}
			<div class="rounded-xl border border-warning/30 bg-warning/10 px-4 py-3 text-sm text-warning print:hidden">
				<p class="font-semibold">Kartu belum siap dicetak massal.</p>
				<p class="mt-1">{readinessIssues.join(', ')}. Rapikan QR/PIN, ruang, dan kursi dari detail sesi sebelum cetak final.</p>
				<a href={resolve(`/asesmen/sesi?event_id=${eventId}&readiness=not_ready`)} class="mt-3 inline-flex rounded-md border border-warning/30 bg-card px-3 py-2 text-sm font-semibold text-warning hover:bg-warning/15">Cek sesi dan ruang</a>
			</div>
		{:else if currentCards.length > 0}
			<div class="rounded-xl border border-success/20 bg-success/10 px-4 py-3 text-sm text-success print:hidden">
				QR Portal Ujian, PIN, ruang, dan nomor meja pada data kartu yang termuat sudah lengkap. Cetak hanya saat distribusi kartu siap.
			</div>
		{/if}

		<MicroActionTable
			title="Status kartu peserta"
			description="Ringkasan layar memakai kode kartu tersamarkan; token mentah tidak dijadikan informasi utama."
			columns={cardColumns}
			rows={currentCards}
			rowKey={(row) => (row as ExamCard).participant_id}
			tableClass="min-w-[820px]"
			class="print:hidden"
			emptyTitle="Belum ada kartu ujian untuk dicetak."
		>
			{#snippet cell(row, column)}
				{@const card = row as ExamCard}
				{#if column.key === 'student'}
					<div class="font-semibold text-foreground">{studentName(card)}</div>
					<div class="text-xs text-muted-foreground">NIS {card.nis} · {displayClass(card)}</div>
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

		<div class="hidden gap-4 md:grid-cols-2 xl:grid-cols-3 print:grid">
			{#each currentCards as card (card.participant_id)}
				<article class="break-inside-avoid rounded-lg border border-primary/20 bg-card p-5 shadow-sm print:shadow-none">
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
						<p><span class="font-medium">Nomor Peserta/NIS:</span> {card.nis}</p>
						<p><span class="font-medium">Kelas:</span> {displayClass(card)}</p>
						<p><span class="font-medium">Ruang:</span> {card.room_name || 'Belum ditentukan'}</p>
						<p><span class="font-medium">No Meja:</span> {card.seat_no ?? '—'}</p>
						<p><span class="font-medium">Mapel/Jadwal:</span> {card.session_title} · {fmtDt(card.scheduled_start)}</p>
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
	</div>
	{/snippet}
</AsyncContent>
