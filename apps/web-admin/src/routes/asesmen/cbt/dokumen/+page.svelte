<script lang="ts">
	import { onMount } from 'svelte';
	import { readClientApiData } from '$lib/client/api';
	import {
		availableRooms,
		cardHasPrintableSecret,
		deriveCardReadiness,
		filterAndSortExamCards,
		summarizeCardsByRoom,
		type ExamCardFilter,
		type ExamCardPrintItem
	} from '$lib/asesmen/exam-card-print';

	type Exam = {
		id: string;
		title: string;
		status: string;
		participant_count?: number;
		room_count?: number;
		card_count?: number;
	};
	type IssueResult = {
		exam_id: string;
		count?: number;
		cards?: ExamCardPrintItem[];
		message?: string;
	};

	let exams = $state<Exam[]>([]);
	let selectedExamId = $state('');
	let cards = $state<ExamCardPrintItem[]>([]);
	let loading = $state(true);
	let cardsLoading = $state(false);
	let issuing = $state(false);
	let regenerate = $state(false);
	let expiresHours = $state(12);
	let search = $state('');
	let roomFilter = $state('all');
	let statusFilter = $state<ExamCardFilter['status']>('all');
	let error = $state('');
	let notice = $state('');
	let lastIssueHadSecrets = $state(false);

	let selectedExam = $derived(exams.find((exam) => exam.id === selectedExamId) ?? null);
	let roomOptions = $derived(availableRooms(cards));
	let visibleCards = $derived(filterAndSortExamCards(cards, { search, room: roomFilter, status: statusFilter }));
	let roomSheets = $derived(summarizeCardsByRoom(cards));
	let readiness = $derived(
		deriveCardReadiness({
			participantCount: selectedExam?.participant_count ?? cards.length,
			roomCount: selectedExam?.room_count ?? roomOptions.filter((room) => room !== 'Belum ruang').length,
			cardCount: selectedExam?.card_count ?? cards.filter((card) => card.status && card.status !== 'not_issued').length
		})
	);
	let readyCards = $derived(cards.filter(cardHasPrintableSecret).length);
	let missingSecrets = $derived(cards.length - readyCards);
	let canLoadCards = $derived(Boolean(selectedExamId) && !cardsLoading);
	let canIssueCards = $derived(Boolean(selectedExamId && readiness.canIssue && !issuing));
	let canPrintCards = $derived(visibleCards.length > 0 && readyCards > 0);

	onMount(() => {
		void loadExams();
	});

	async function loadExams() {
		loading = true;
		error = '';
		try {
			exams = await fetch('/api/asesmen/exams?limit=20').then((response) =>
				readClientApiData<Exam[]>(response, 'Daftar ujian belum dapat dimuat')
			);
			if (exams.length > 0 && !selectedExamId) {
				selectedExamId = exams[0].id;
				await loadCards();
			}
		} catch (err) {
			error = err instanceof Error ? err.message : 'Daftar ujian belum dapat dimuat';
			exams = [];
		} finally {
			loading = false;
		}
	}

	async function loadCards() {
		if (!selectedExamId) return;
		cardsLoading = true;
		error = '';
		lastIssueHadSecrets = false;
		try {
			cards = await fetch(`/api/asesmen/exams/${encodeURIComponent(selectedExamId)}/participant-cards`).then((response) =>
				readClientApiData<ExamCardPrintItem[]>(response, 'Daftar kartu peserta belum dapat dimuat')
			);
			notice = cards.length > 0 ? 'Daftar kartu peserta dimuat. PIN mentah hanya muncul langsung setelah Terbitkan QR+PIN.' : 'Belum ada peserta/kartu untuk ujian ini.';
		} catch (err) {
			error = err instanceof Error ? err.message : 'Daftar kartu peserta gagal dimuat';
			cards = [];
		} finally {
			cardsLoading = false;
		}
	}

	async function issueCards() {
		if (!selectedExamId || !canIssueCards) return;
		const ok = window.confirm('Terbitkan QR+PIN peserta sekarang? PIN mentah hanya tampil pada hasil ini, jadi langsung cetak atau simpan PDF setelah berhasil.');
		if (!ok) return;
		issuing = true;
		error = '';
		notice = '';
		try {
			const result = await fetch(`/api/asesmen/exams/${encodeURIComponent(selectedExamId)}/issue-cards`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ regenerate, expires_hours: Number(expiresHours) })
			}).then((response) => readClientApiData<IssueResult>(response, 'Terbitkan QR+PIN gagal'));
			if (Array.isArray(result.cards)) {
				cards = result.cards;
				lastIssueHadSecrets = result.cards.some(cardHasPrintableSecret);
			} else {
				await loadCards();
			}
			notice = result.message ?? 'Kartu peserta siap. Cetak/simpan PDF sekarang.';
		} catch (err) {
			error = err instanceof Error ? err.message : 'Terbitkan QR+PIN gagal';
		} finally {
			issuing = false;
		}
	}

	function onExamChange() {
		cards = [];
		search = '';
		roomFilter = 'all';
		statusFilter = 'all';
		notice = '';
		void loadCards();
	}

	function resetFilters() {
		search = '';
		roomFilter = 'all';
		statusFilter = 'all';
	}

	function printPage() {
		window.print();
	}

	function cardCredential(card: ExamCardPrintItem) {
		if (cardHasPrintableSecret(card)) return `${card.token} / PIN ${card.pin}`;
		if (card.status && card.status !== 'not_issued') return 'Sudah terbit · PIN tidak ditampilkan ulang';
		return 'Belum terbit';
	}
</script>

<svelte:head>
	<title>Dokumen & Cetak CBT</title>
	<meta name="description" content="Panel operasional dokumen CBT familiar: terbitkan QR+PIN, cetak kartu peserta, lembar pengawas ruang, daftar hadir, denah, dan checklist arsip." />
</svelte:head>

<div class="space-y-4 pb-10">
	<header class="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm print:hidden">
		<p class="text-xs font-black uppercase tracking-[0.2em] text-emerald-700">Laporan · Operasional Penuh</p>
		<div class="mt-2 flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between">
			<div>
				<h1 class="text-3xl font-black tracking-tight md:text-4xl">Dokumen & Cetak</h1>
				<p class="mt-2 max-w-3xl text-sm leading-6 text-slate-600">
					Pilih ujian, muat kartu, terbitkan QR+PIN secara eksplisit, lalu cetak kartu peserta dan lembar pengawas ruang dari satu panel. QR+PIN tidak dibuat otomatis dari ruang.
				</p>
			</div>
			<div class="flex flex-wrap gap-2">
				<a href="/asesmen/cbt/ruang" class="rounded-2xl border border-slate-300 bg-white px-4 py-2 text-sm font-black text-slate-700 shadow-sm hover:border-emerald-300 hover:text-emerald-800">Ruang Ujian</a>
				<a href="/asesmen/dokumen" class="rounded-2xl border border-slate-300 bg-white px-4 py-2 text-sm font-black text-slate-700 shadow-sm hover:border-emerald-300 hover:text-emerald-800">Dokumen ringkas</a>
				<button class="rounded-2xl border border-slate-300 bg-white px-4 py-2 text-sm font-black text-slate-700 shadow-sm hover:border-emerald-300 hover:text-emerald-800" onclick={printPage}>Cetak halaman</button>
			</div>
		</div>
	</header>

	{#if error}
		<p class="rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-sm font-bold text-red-800 print:hidden" role="alert">{error}</p>
	{/if}
	{#if notice}
		<p class="rounded-2xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm font-bold text-emerald-900 print:hidden">{notice}</p>
	{/if}
	{#if lastIssueHadSecrets}
		<p class="rounded-2xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm font-black text-amber-900 print:hidden">PIN mentah sedang tampil dari hasil Terbitkan QR+PIN. Cetak atau simpan PDF sekarang sebelum memuat ulang halaman.</p>
	{/if}

	<section class="grid gap-4 lg:grid-cols-[0.9fr_1.1fr] print:hidden">
		<div class="rounded-3xl border border-slate-200 bg-white p-4 shadow-sm">
			<h2 class="text-lg font-black text-slate-950">1. Pilih ujian & status cetak</h2>
			<label class="mt-4 block space-y-1.5 text-sm font-bold text-slate-700">
				<span>Ujian</span>
				<select class="w-full rounded-2xl border border-slate-300 bg-white px-3 py-2 text-sm outline-none focus:border-emerald-500" bind:value={selectedExamId} onchange={onExamChange} disabled={loading}>
					<option value="">{loading ? 'Memuat ujian...' : 'Pilih ujian'}</option>
					{#each exams as exam (exam.id)}
						<option value={exam.id}>{exam.title} · {exam.status}</option>
					{/each}
				</select>
			</label>
			<div class="mt-4 grid gap-3 sm:grid-cols-3">
				<div class="rounded-2xl bg-slate-50 p-3"><p class="text-xs font-black uppercase text-slate-500">Ruang</p><p class="text-2xl font-black text-slate-950">{selectedExam?.room_count ?? roomSheets.length}</p></div>
				<div class="rounded-2xl bg-slate-50 p-3"><p class="text-xs font-black uppercase text-slate-500">Peserta</p><p class="text-2xl font-black text-slate-950">{selectedExam?.participant_count ?? cards.length}</p></div>
				<div class="rounded-2xl bg-emerald-50 p-3"><p class="text-xs font-black uppercase text-emerald-700">Kartu</p><p class="text-2xl font-black text-emerald-950">{selectedExam?.card_count ?? readyCards}</p></div>
			</div>
			<p class="mt-3 rounded-2xl bg-amber-50 p-3 text-xs font-bold leading-5 text-amber-900">{readiness.message} Simpan Ruang tidak menerbitkan QR+PIN/token.</p>
			<div class="mt-4 flex flex-wrap gap-2">
				<button class="rounded-2xl border border-slate-300 px-4 py-2 text-sm font-black text-slate-700 hover:border-emerald-300 hover:text-emerald-800 disabled:cursor-not-allowed disabled:opacity-50" disabled={!canLoadCards} onclick={loadCards}>{cardsLoading ? 'Memuat...' : 'Muat Daftar Kartu'}</button>
				<button class="rounded-2xl bg-emerald-700 px-4 py-2 text-sm font-black text-white hover:bg-emerald-800 disabled:cursor-not-allowed disabled:bg-slate-300" disabled={!canIssueCards} onclick={issueCards}>{issuing ? 'Menerbitkan...' : 'Terbitkan QR+PIN'}</button>
				<button class="rounded-2xl border border-slate-300 px-4 py-2 text-sm font-black text-slate-700 hover:border-emerald-300 hover:text-emerald-800 disabled:cursor-not-allowed disabled:opacity-50" disabled={!canPrintCards} onclick={printPage}>Cetak Kartu</button>
			</div>
			<div class="mt-4 grid gap-3 sm:grid-cols-2">
				<label class="flex items-center gap-2 rounded-2xl border border-slate-200 bg-slate-50 p-3 text-xs font-bold text-slate-700">
					<input type="checkbox" bind:checked={regenerate} />
					Regenerasi sengaja bila kartu lama harus diganti
				</label>
				<label class="space-y-1.5 text-xs font-bold text-slate-700">
					<span>Masa berlaku PIN/jam</span>
					<input type="number" min="1" max="72" bind:value={expiresHours} class="w-full rounded-2xl border border-slate-300 bg-white px-3 py-2 text-sm outline-none focus:border-emerald-500" />
				</label>
			</div>
		</div>

		<div class="rounded-3xl border border-slate-200 bg-white p-4 shadow-sm">
			<h2 class="text-lg font-black text-slate-950">2. Checklist dokumen</h2>
			<div class="mt-4 divide-y divide-slate-200 rounded-2xl border border-slate-200">
				<div class="grid gap-2 p-3 sm:grid-cols-[1fr_auto] sm:items-center">
					<div><p class="font-black text-slate-950">Kartu peserta QR+PIN</p><p class="text-xs font-bold text-slate-500">{readyCards} siap cetak · {missingSecrets} belum menampilkan PIN/token</p></div>
					<span class="rounded-full bg-emerald-50 px-3 py-1 text-xs font-black text-emerald-800">{readiness.actionLabel}</span>
				</div>
				<div class="grid gap-2 p-3 sm:grid-cols-[1fr_auto] sm:items-center">
					<div><p class="font-black text-slate-950">Lembar pengawas ruang</p><p class="text-xs font-bold text-slate-500">Daftar peserta per ruang, tidak memuat PIN rahasia peserta.</p></div>
					<span class="rounded-full bg-slate-100 px-3 py-1 text-xs font-black text-slate-700">{roomSheets.length} ruang</span>
				</div>
				<div class="grid gap-2 p-3 sm:grid-cols-[1fr_auto] sm:items-center">
					<div><p class="font-black text-slate-950">Denah, daftar hadir, BA, arsip</p><p class="text-xs font-bold text-slate-500">Template ringkas ikut tercetak bersama ringkasan ruang.</p></div>
					<span class="rounded-full bg-slate-100 px-3 py-1 text-xs font-black text-slate-700">Checklist Arsip</span>
				</div>
			</div>
		</div>
	</section>

	<section class="rounded-3xl border border-slate-200 bg-white p-4 shadow-sm print:hidden">
		<div class="flex flex-col gap-3 border-b border-slate-200 pb-3 lg:flex-row lg:items-end lg:justify-between">
			<div>
				<h2 class="text-lg font-black text-slate-950">3. Daftar Kartu Peserta</h2>
				<p class="text-sm font-semibold text-slate-500">Filter, cek status PIN, lalu cetak. Baris tanpa PIN menandakan data dari GET lama atau kartu belum diterbitkan.</p>
			</div>
			<span class="rounded-full bg-slate-100 px-3 py-1 text-xs font-black text-slate-700">{visibleCards.length}/{cards.length} tampil</span>
		</div>
		<div class="mt-4 grid gap-3 md:grid-cols-[1.2fr_0.8fr_0.8fr_auto]">
			<input class="min-w-0 rounded-2xl border border-slate-300 px-3 py-2 text-sm outline-none focus:border-emerald-500" placeholder="Cari nama, kelas, NIS, ruang..." bind:value={search} />
			<select class="min-w-0 rounded-2xl border border-slate-300 bg-white px-3 py-2 text-sm outline-none focus:border-emerald-500" bind:value={roomFilter}>
				<option value="all">Semua ruang</option>
				{#each roomOptions as room (room)}<option value={room}>{room}</option>{/each}
			</select>
			<select class="min-w-0 rounded-2xl border border-slate-300 bg-white px-3 py-2 text-sm outline-none focus:border-emerald-500" bind:value={statusFilter}>
				<option value="all">Semua status</option>
				<option value="ready">Siap cetak</option>
				<option value="missing-secret">Tanpa PIN/token</option>
				<option value="not-issued">Belum terbit</option>
			</select>
			<button class="rounded-2xl border border-slate-300 px-4 py-2 text-sm font-black text-slate-700 hover:border-emerald-300 hover:text-emerald-800" onclick={resetFilters}>Reset</button>
		</div>
		{#if cards.length === 0}
			<p class="mt-4 rounded-2xl border border-dashed border-slate-300 bg-slate-50 p-4 text-sm text-slate-600">Belum ada kartu. Pilih ujian lalu klik Muat Daftar Kartu atau Terbitkan QR+PIN setelah peserta dan ruang siap.</p>
		{:else}
			<div class="mt-4 overflow-x-auto rounded-2xl border border-slate-200">
				<table class="min-w-[860px] w-full text-left text-sm">
					<thead class="bg-slate-50 text-xs uppercase text-slate-500"><tr><th class="px-3 py-2">Peserta</th><th class="px-3 py-2">Kelas</th><th class="px-3 py-2">Ruang/Kursi</th><th class="px-3 py-2">QR+PIN</th><th class="px-3 py-2">Status</th></tr></thead>
					<tbody class="divide-y divide-slate-100">
						{#each visibleCards as card (card.participant_id)}
							<tr>
								<td class="px-3 py-2"><p class="font-black text-slate-950">{card.student_name}</p><p class="text-xs font-bold text-slate-500">{card.nis || card.nisn || 'tanpa NIS'}</p></td>
								<td class="px-3 py-2 font-bold text-slate-700">{card.class_code || card.class_name || '—'}</td>
								<td class="px-3 py-2 font-black text-slate-950">{card.room_code || '—'} / {card.seat_no || '-'}</td>
								<td class="px-3 py-2 text-xs font-bold text-slate-600">{cardCredential(card)}</td>
								<td class="px-3 py-2"><span class={`rounded-full px-2 py-1 text-xs font-black ${cardHasPrintableSecret(card) ? 'bg-emerald-50 text-emerald-800' : 'bg-amber-50 text-amber-900'}`}>{card.status || 'not_issued'}</span></td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	</section>

	<section class="rounded-3xl border border-slate-200 bg-white p-4 shadow-sm print:hidden">
		<h2 class="text-lg font-black text-slate-950">4. Lembar Pengawas Ruang</h2>
		<p class="text-sm font-semibold text-slate-500">Ringkasan ruang untuk pengawas/substitusi. Tidak menampilkan PIN peserta.</p>
		<div class="mt-4 grid gap-2 md:grid-cols-2 xl:grid-cols-4">
			{#each roomSheets as sheet (sheet.roomCode)}
				<div class="rounded-2xl border border-slate-200 p-3">
					<div class="flex items-center justify-between gap-2"><p class="font-black text-slate-950">{sheet.roomCode}</p><span class="rounded-full bg-slate-100 px-2 py-0.5 text-xs font-black text-slate-600">{sheet.total} peserta</span></div>
					<p class="mt-1 text-xs font-bold text-slate-500">Kursi {sheet.firstSeat || '-'}–{sheet.lastSeat || '-'} · siap {sheet.ready} · cek {sheet.missingSecret}</p>
				</div>
			{/each}
		</div>
	</section>

	<section class="hidden print:block">
		<h1 class="text-xl font-black">Dokumen & Cetak CBT</h1>
		<p class="text-sm">{selectedExam?.title ?? 'Ujian belum dipilih'}</p>
		<h2 class="mt-4 text-lg font-black">Kartu Peserta</h2>
		<div class="grid grid-cols-2 gap-2">
			{#each visibleCards as card (card.participant_id)}
				<div class="break-inside-avoid border border-slate-300 p-3 text-sm">
					<strong>{card.student_name}</strong><br />
					{card.class_code || card.class_name || '—'} · {card.room_code || '—'} / {card.seat_no || '-'}<br />
					{cardCredential(card)}<br />
					{card.qr_path || (card.token ? `/ujian?card=${card.token}` : 'QR belum tersedia')}
				</div>
			{/each}
		</div>
		<h2 class="mt-4 text-lg font-black">Lembar Pengawas Ruang</h2>
		<div class="grid grid-cols-2 gap-2">
			{#each roomSheets as sheet (sheet.roomCode)}
				<div class="border border-slate-300 p-2 text-sm"><strong>{sheet.roomCode}</strong> · {sheet.total} peserta · Kursi {sheet.firstSeat || '-'}–{sheet.lastSeat || '-'}</div>
			{/each}
		</div>
		<h2 class="mt-4 text-lg font-black">Checklist Arsip</h2>
		<ul class="text-sm"><li>Kartu peserta tersimpan PDF</li><li>Lembar pengawas ruang tercetak</li><li>Daftar hadir/BA disiapkan</li><li>QR+PIN tidak dibagikan di luar panitia</li></ul>
	</section>
</div>
