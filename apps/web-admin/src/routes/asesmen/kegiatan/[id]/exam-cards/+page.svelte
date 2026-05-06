<script lang="ts">
	import PrinterIcon from '@lucide/svelte/icons/printer';
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { Button } from '$lib/components/ui/button';
	import { fetchSchoolProfile, schoolAddressLine, type SchoolProfile } from '$lib/school-profile';
	import { clientApiPath, readClientApiData } from '$lib/client/api';

	type ExamCard = {
		event_id: string;
		event_title: string;
		exam_type: string;
		event_scope: string;
		academic_year_name: string;
		session_id: string;
		session_title: string;
		scheduled_start: string;
		participant_id: string;
		token: string;
		seat_no: number | null;
		nis: string;
		student_nama: string;
		gender: string;
		class_code: string;
		room_name: string;
	};
	type ExamCardPrintData = {
		schoolProfile: SchoolProfile;
		cards: ExamCard[];
	};

	const eventId = page.params.id ?? '';
	let cardsPromise = $state<Promise<ExamCardPrintData> | null>(null);

	async function fetchCardRows() {
		const response = await fetch(clientApiPath`/api/asesmen/events/${eventId}/exam-cards`);
		const rows = await readClientApiData<ExamCard[]>(response, 'Gagal memuat kartu ujian');
		return Array.isArray(rows) ? rows : [];
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
		console.error('CBT event exam cards render failed', error);
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
		const missingToken = cards.filter((card) => !card.token).length;
		const missingRoom = cards.filter((card) => !card.room_name).length;
		const missingSeat = cards.filter((card) => card.seat_no === null || card.seat_no === undefined).length;
		const issues: string[] = [];
		if (missingToken > 0) issues.push(`${missingToken} kartu belum punya token`);
		if (missingRoom > 0) issues.push(`${missingRoom} peserta belum punya ruangan`);
		if (missingSeat > 0) issues.push(`${missingSeat} peserta belum punya nomor meja`);
		return issues;
	}

	onMount(() => {
		void loadCards();
	});
</script>

<svelte:head>
	<title>Kartu Ujian Event</title>
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
					<div class="rounded-lg border border-emerald-200 bg-white p-5 shadow-sm">
						<div class="space-y-2 border-b border-dashed border-emerald-200 pb-3">
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
						<div class="mt-5 rounded-md bg-emerald-50 px-4 py-3">
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
				<h1 class="text-2xl font-semibold text-slate-900">Kartu Ujian Event</h1>
				<p class="text-sm text-slate-500">Cetak per peserta dengan token rahasia, ruangan, dan nomor meja.</p>
			</div>
			<div class="flex flex-wrap gap-2">
				<a href={resolve(`/asesmen/kegiatan/${eventId}`)} class="inline-flex items-center rounded-md border border-green-200 bg-green-50 px-3 py-2 text-sm font-semibold text-green-800 hover:bg-green-100">Kembali ke Event</a>
			<Button onclick={() => window.print()} disabled={printDisabled}>
				<PrinterIcon class="mr-2 size-4" />
				Cetak
			</Button>
			</div>
		</div>

		{#if currentCards.length === 0}
			<div class="rounded-xl border border-dashed border-amber-300 bg-amber-50 px-4 py-5 text-sm text-amber-950 print:hidden">
				<p class="font-semibold">Belum ada kartu ujian untuk dicetak.</p>
				<p class="mt-1">Daftarkan peserta, buat token, lalu lengkapi ruangan dan nomor meja pada sesi event sebelum cetak massal.</p>
				<div class="mt-3 flex flex-wrap gap-2">
					<a href={resolve(`/asesmen/sesi?event_id=${eventId}&readiness=not_ready`)} class="inline-flex rounded-md border border-amber-200 bg-white px-3 py-2 text-sm font-semibold text-amber-900 hover:bg-amber-100">Cek sesi event</a>
					<a href={resolve(`/asesmen/kegiatan/${eventId}/members`)} class="inline-flex rounded-md border border-amber-200 bg-white px-3 py-2 text-sm font-semibold text-amber-900 hover:bg-amber-100">Cek penugasan/peserta</a>
				</div>
			</div>
		{:else if readinessIssues.length > 0}
			<div class="rounded-xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-900 print:hidden">
				<p class="font-semibold">Kartu belum siap dicetak massal.</p>
				<p class="mt-1">{readinessIssues.join(', ')}. Rapikan token, ruang, dan kursi dari detail sesi sebelum cetak final.</p>
				<a href={resolve(`/asesmen/sesi?event_id=${eventId}&readiness=not_ready`)} class="mt-3 inline-flex rounded-md border border-amber-200 bg-white px-3 py-2 text-sm font-semibold text-amber-900 hover:bg-amber-100">Cek sesi dan ruang</a>
			</div>
		{:else if currentCards.length > 0}
			<div class="rounded-xl border border-green-200 bg-green-50 px-4 py-3 text-sm text-green-900 print:hidden">
				Token rahasia, ruang, dan nomor meja pada data kartu yang termuat sudah lengkap. Cetak hanya saat distribusi kartu siap dan jangan tampilkan token di layar umum.
			</div>
		{/if}

		<div class="rounded-xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-950 print:hidden">
			<p class="font-semibold">Token pada kartu ujian adalah kredensial rahasia peserta.</p>
			<p class="mt-1">Cetak dan simpan kartu melalui panitia/pengawas resmi. Hindari membagikan file cetak ke grup umum, layar proyektor, atau kanal yang dapat diakses peserta lain.</p>
		</div>

		<div class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
			{#each currentCards as card (card.participant_id)}
				<article class="break-inside-avoid rounded-lg border border-emerald-200 bg-white p-5 shadow-sm print:shadow-none">
					<div class="border-b border-dashed border-emerald-200 pb-3">
						<p class="text-xs font-semibold uppercase tracking-[0.16em] text-emerald-700">{schoolProfile.ministry_line}</p>
						<p class="mt-1 text-sm font-semibold uppercase text-slate-900">{schoolProfile.name}</p>
						<p class="mt-1 text-[11px] leading-4 text-slate-500">{schoolAddressLine(schoolProfile) || schoolProfile.office_line}</p>
						<h2 class="mt-2 text-lg font-semibold text-slate-900">{card.event_title}</h2>
						<p class="text-sm text-slate-500">{card.session_title}</p>
					</div>
					<div class="mt-4 space-y-2 text-sm text-slate-700">
						<p><span class="font-medium">Nama:</span> {card.student_nama}</p>
						<p><span class="font-medium">NIS:</span> {card.nis}</p>
						<p><span class="font-medium">Kelas:</span> {card.class_code || '—'}</p>
						<p><span class="font-medium">Ruangan:</span> {card.room_name || 'Belum ditentukan'}</p>
						<p><span class="font-medium">No Meja:</span> {card.seat_no ?? '—'}</p>
						<p><span class="font-medium">Jadwal:</span> {fmtDt(card.scheduled_start)}</p>
					</div>
					<div class="mt-5 rounded-md bg-emerald-50 px-4 py-3">
						<p class="text-xs uppercase tracking-[0.2em] text-emerald-700">Token Ujian Rahasia</p>
						<p class="mt-1 font-mono text-2xl font-bold text-emerald-950">{card.token || 'Belum digenerate'}</p>
						<p class="mt-1 text-[11px] text-emerald-800 print:hidden">Bagikan hanya kepada peserta terkait atau pengawas ruangan.</p>
					</div>
				</article>
			{/each}
		</div>
	</div>
	{/snippet}
</AsyncContent>
