<svelte:head>
	<title>Command Center CBT</title>
	<meta
		name="description"
		content="Command Center CBT web-first untuk alur Asesmen Ujian sederhana MTsN 2 Kolaka Utara."
	/>
</svelte:head>

<script lang="ts">
	import { onMount } from 'svelte';
	import { readClientApiData } from '$lib/client/api';
	import { summarizeAssessmentPrintDocuments, type AssessmentPrintDocument } from '$lib/asesmen/document-print-readiness';
	type Exam = {
		id: string;
		title: string;
		grade_level?: number;
		status: string;
		starts_at?: string;
		ends_at?: string;
		session_count?: number;
		room_count?: number;
		participant_count?: number;
		card_count?: number;
	};

	type ActionResult = {
		message?: string;
		room_count?: number;
		participant_count?: number;
		card_count?: number;
	};

	type AssignmentClassSummary = {
		class_code: string;
		class_name?: string;
		grade_level?: number;
		count: number;
	};

	type AssignmentRoom = {
		code: string;
		name: string;
		capacity: number;
		assigned_count: number;
		grade_levels?: number[];
		class_summary?: AssignmentClassSummary[];
	};

	type AssignmentResult = ActionResult & {
		exam_id: string;
		session_id?: string;
		mix_policy: string;
		room_count: number;
		capacity_per_room: number;
		total_participants: number;
		assigned_total: number;
		unassigned_total: number;
		rooms: AssignmentRoom[];
		applied?: boolean;
	};

	type Rombel = {
		id: string;
		code: string;
		name: string;
		level?: number;
		is_active?: boolean;
		total_students?: number;
	};

	const steps = [
		{
			label: 'Siapkan Ujian',
			detail: 'Buat draft ujian dulu: nama, tingkat kelas, tanggal dan jam. Paket soal tetap di Bank Soal.',
			status: 'Aktif'
		},
		{
			label: 'Atur 8 Ruang',
			detail: 'Pilih ujian, cek preview R01–R08, lalu simpan pembagian ruang tanpa membuat token peserta.',
			status: 'Dikerjakan'
		},
		{
			label: 'Cetak Kartu & Pengawas',
			detail: 'Tetap aman: QR+PIN dan kartu belum diterbitkan sampai peserta ujian tersambung.',
			status: 'Berikutnya'
		},
		{
			label: 'Pelaksanaan & Hasil',
			detail: 'Nanti hanya tampilkan pantauan ruang dan rekap nilai penting, bukan konsol teknis yang ramai.',
			status: 'Berikutnya'
		}
	];

	let exams = $state<Exam[]>([]);
	let loading = $state(true);
	let saving = $state(false);
	let actionExamId = $state<string | null>(null);
	let error = $state('');
	let notice = $state('');
	let title = $state('');
	let gradeLevel = $state('7');
	let startsAt = $state('');
	let endsAt = $state('');
	let selectedExamId = $state<string | null>(null);
	let roomCount = $state(8);
	let capacityPerRoom = $state(30);
	let mixPolicy = $state<'mixed' | 'class_grouped'>('mixed');
	let roomPreview = $state<AssignmentResult | null>(null);
	let roomPreviewExamId = $state<string | null>(null);
	let roomBusy = $state(false);
	let rombels = $state<Rombel[]>([]);
	let selectedClassIds = $state<Set<string>>(new Set());
	let classError = $state('');

	let selectedExam = $derived(exams.find((exam) => exam.id === selectedExamId) ?? exams[0] ?? null);
	let activeRombels = $derived(rombels.filter((rombel) => rombel.is_active !== false));
	let selectedClassCount = $derived(selectedClassIds.size);
	let selectedStudentEstimate = $derived(
		activeRombels
			.filter((rombel) => selectedClassIds.has(rombel.id))
			.reduce((total, rombel) => total + (rombel.total_students ?? 0), 0)
	);
	let totalCapacity = $derived(roomCount * capacityPerRoom);
	let canSaveRooms = $derived(Boolean(selectedExam && roomPreview && roomPreview.exam_id === selectedExam.id && !roomBusy));
	let hasPlacementSummary = $derived(Boolean(roomPreview?.applied && roomPreview.rooms.some((room) => (room.class_summary?.length ?? 0) > 0)));
	let printDocuments = $derived(summarizeAssessmentPrintDocuments(selectedExam ?? {}));

	onMount(() => {
		void loadInitialData();
	});

	async function loadInitialData() {
		const [examResult, rombelResult] = await Promise.allSettled([loadExams(), loadRombels()]);
		if (examResult.status === 'rejected') {
			error = examResult.reason instanceof Error ? examResult.reason.message : 'Daftar asesmen belum dapat dimuat';
		}
		if (rombelResult.status === 'rejected') {
			classError = rombelResult.reason instanceof Error ? rombelResult.reason.message : 'Daftar rombel belum dapat dimuat';
		}
	}

	async function loadExams() {
		loading = true;
		error = '';
		try {
			exams = await fetch('/api/asesmen/exams?limit=12').then((response) =>
				readClientApiData<Exam[]>(response, 'Daftar asesmen belum dapat dimuat')
			);
			if (exams.length > 0 && (!selectedExamId || !exams.some((exam) => exam.id === selectedExamId))) {
				selectedExamId = exams[0].id;
			}
			if (exams.length === 0) {
				selectedExamId = null;
				roomPreview = null;
				roomPreviewExamId = null;
			}
		} catch (err) {
			error = err instanceof Error ? err.message : 'Daftar asesmen belum dapat dimuat';
			exams = [];
		} finally {
			loading = false;
		}
	}

	async function loadRombels() {
		classError = '';
		try {
			rombels = await fetch('/api/academic/rombel').then((response) =>
				readClientApiData<Rombel[]>(response, 'Daftar rombel belum dapat dimuat')
			);
		} catch (err) {
			classError = err instanceof Error ? err.message : 'Daftar rombel belum dapat dimuat';
			rombels = [];
		}
	}

	function toggleClass(id: string) {
		const next = new Set(selectedClassIds);
		if (next.has(id)) {
			next.delete(id);
		} else {
			next.add(id);
		}
		selectedClassIds = next;
		roomPreview = null;
		roomPreviewExamId = null;
	}

	function toISO(value: string) {
		if (!value) return undefined;
		const date = new Date(value);
		return Number.isNaN(date.getTime()) ? undefined : date.toISOString();
	}

	function assignmentPayload() {
		return {
			room_count: Number(roomCount),
			capacity_per_room: Number(capacityPerRoom),
			mix_policy: mixPolicy,
			class_ids: Array.from(selectedClassIds)
		};
	}

	function selectExamForRooms(exam: Exam) {
		selectedExamId = exam.id;
		error = '';
		notice = '';
		if (roomPreviewExamId !== exam.id) {
			roomPreview = null;
		}
	}

	async function createDraft() {
		const trimmedTitle = title.trim();
		if (!trimmedTitle || saving) return;
		saving = true;
		error = '';
		notice = '';
		try {
			const created = await fetch('/api/asesmen/exams', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					title: trimmedTitle,
					grade_level: Number(gradeLevel),
					starts_at: toISO(startsAt),
					ends_at: toISO(endsAt)
				})
			}).then((response) => readClientApiData<Exam>(response, 'Draft ujian belum dapat dibuat'));
			notice = 'Draft ujian tersimpan. Lanjutkan ke panel Atur Ruang.';
			title = '';
			selectedExamId = created.id;
			roomPreview = null;
			roomPreviewExamId = null;
			await loadExams();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Draft ujian belum dapat dibuat';
		} finally {
			saving = false;
		}
	}

	async function previewRooms(exam = selectedExam) {
		if (!exam || roomBusy) return;
		roomBusy = true;
		actionExamId = exam.id;
		error = '';
		notice = '';
		try {
			const result = await fetch(`/api/asesmen/exams/${encodeURIComponent(exam.id)}/assignment-preview`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(assignmentPayload())
			}).then((response) => readClientApiData<AssignmentResult>(response, 'Preview ruang belum dapat dibuat'));
			roomPreview = result;
			roomPreviewExamId = exam.id;
			notice = result.message ?? 'Preview ruang siap.';
		} catch (err) {
			error = err instanceof Error ? err.message : 'Preview ruang belum dapat dibuat';
		} finally {
			roomBusy = false;
			actionExamId = null;
		}
	}

	async function saveRooms(exam = selectedExam) {
		if (!exam || roomBusy) return;
		roomBusy = true;
		actionExamId = exam.id;
		error = '';
		notice = '';
		try {
			const result = await fetch(`/api/asesmen/exams/${encodeURIComponent(exam.id)}/assignment-apply`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(assignmentPayload())
			}).then((response) => readClientApiData<AssignmentResult>(response, 'Ruang ujian belum dapat disimpan'));
			roomPreview = result;
			roomPreviewExamId = exam.id;
			notice = result.message ?? 'Ruang ujian tersimpan.';
			await loadExams();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Ruang ujian belum dapat disimpan';
		} finally {
			roomBusy = false;
			actionExamId = null;
		}
	}

	async function runExamAction(exam: Exam, action: 'issue-cards') {
		if (actionExamId) return;
		actionExamId = exam.id;
		error = '';
		notice = '';
		try {
			const result = await fetch(`/api/asesmen/exams/${encodeURIComponent(exam.id)}/${action}`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' }
			}).then((response) => readClientApiData<ActionResult>(response, 'Aksi asesmen belum dapat dijalankan'));
			notice = result.message ?? 'Aksi asesmen selesai.';
			await loadExams();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Aksi asesmen belum dapat dijalankan';
		} finally {
			actionExamId = null;
		}
	}

	function handlePrintDocumentAction(document: AssessmentPrintDocument) {
		if (!selectedExam) {
			error = 'Pilih ujian dulu sebelum membuka Dokumen & Cetak.';
			return;
		}
		if (!document.ready) {
			notice = document.detail;
			return;
		}
		if (document.id === 'participant-cards' && document.dangerous) {
			const ok = window.confirm('Terbitkan QR+PIN untuk kartu peserta ujian ini? Aksi ini sengaja dipisah dari Simpan Ruang agar tidak membuat token tanpa sengaja.');
			if (!ok) return;
			void runExamAction(selectedExam, 'issue-cards');
			return;
		}
		if (document.id === 'participant-cards') {
			notice = 'Permukaan cetak kartu peserta sudah dipisah. Data kartu akan tampil setelah endpoint cetak penuh diaktifkan.';
			return;
		}
		if (document.id === 'supervisor-sheets') {
			notice = 'Lembar Pengawas Ruang siap sebagai dokumen per ruang. Tidak ada PIN peserta yang dibuat dari tombol ini.';
			return;
		}
		notice = 'Checklist arsip siap dibaca. Simpan kartu peserta, lembar pengawas, daftar hadir, dan berita acara setelah ujian selesai.';
	}

	function formatClassSummary(items?: AssignmentClassSummary[]) {
		if (!items || items.length === 0) return 'Belum ada peserta';
		return items.map((item) => `${item.class_code || item.class_name || 'Tanpa rombel'} ${item.count}`).join(' · ');
	}

	function formatGradeLevels(levels?: number[]) {
		if (!levels || levels.length === 0) return '—';
		return levels.map((level) => `Kelas ${level}`).join(', ');
	}

	function formatDate(value?: string) {
		if (!value) return 'Belum dijadwalkan';
		return new Intl.DateTimeFormat('id-ID', {
			dateStyle: 'medium',
			timeStyle: 'short'
		}).format(new Date(value));
	}
</script>

<div class="min-h-screen bg-slate-50 text-slate-950">
	<div class="mx-auto flex w-full max-w-6xl flex-col gap-5 px-4 py-6 sm:px-6 lg:px-8">
		<header class="rounded-3xl border border-emerald-200 bg-white p-5 shadow-sm">
			<div class="flex flex-col gap-4 md:flex-row md:items-end md:justify-between">
				<div class="space-y-3">
					<span class="inline-flex w-fit rounded-full border border-emerald-200 bg-emerald-50 px-3 py-1 text-xs font-black uppercase tracking-[0.2em] text-emerald-800">
						CBT Web
					</span>
					<div class="space-y-2">
						<h1 class="text-3xl font-black tracking-tight text-slate-950 md:text-5xl">Command Center CBT</h1>
						<p class="max-w-3xl text-base leading-7 text-slate-600">
							Alur sederhana untuk panitia: buat draft ujian, atur 8 ruang, lalu lanjut peserta dan kartu setelah fondasi ruang rapi.
						</p>
					</div>
				</div>
				<div class="flex flex-wrap gap-2">
					<a href="/asesmen/prototype" class="rounded-2xl border border-slate-300 bg-white px-4 py-2 text-sm font-black text-slate-700 shadow-sm hover:border-emerald-300 hover:text-emerald-800">
						Lihat Prototype
					</a>
					<a href="/bank-soal" class="rounded-2xl bg-emerald-700 px-4 py-2 text-sm font-black text-white shadow-sm hover:bg-emerald-800">
						Buka Bank Soal
					</a>
				</div>
			</div>
		</header>

		{#if error}
			<p class="rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-sm font-bold text-red-800" role="alert">{error}</p>
		{/if}
		{#if notice}
			<p class="rounded-2xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm font-bold text-emerald-800" role="status">{notice}</p>
		{/if}

		<section class="rounded-3xl border border-slate-200 bg-white p-4 shadow-sm">
			<div class="mb-3 flex flex-col gap-1 border-b border-slate-200 pb-3 sm:flex-row sm:items-end sm:justify-between">
				<div>
					<p class="text-xs font-bold uppercase tracking-[0.18em] text-slate-500">Asesmen Ujian</p>
					<h2 class="text-xl font-black text-slate-950">Checklist ringkas panitia</h2>
				</div>
				<p class="text-sm font-semibold text-slate-500">Produksi awal — fokus ruang ujian dulu.</p>
			</div>

			<div class="divide-y divide-slate-200 overflow-hidden rounded-2xl border border-slate-200">
				{#each steps as step, index}
					<div class="grid gap-3 p-4 md:grid-cols-[3rem_1fr_auto] md:items-center">
						<div class="flex h-10 w-10 items-center justify-center rounded-2xl bg-emerald-50 text-sm font-black text-emerald-800">
							{index + 1}
						</div>
						<div class="min-w-0">
							<h3 class="font-black text-slate-950">{step.label}</h3>
							<p class="mt-1 text-sm leading-6 text-slate-600">{step.detail}</p>
						</div>
						<span class="w-fit rounded-full bg-slate-100 px-3 py-1 text-xs font-black text-slate-600">{step.status}</span>
					</div>
				{/each}
			</div>
		</section>

		<section class="grid gap-4 lg:grid-cols-[0.85fr_1.15fr]">
			<form class="rounded-3xl border border-slate-200 bg-white p-4 shadow-sm" onsubmit={(event) => { event.preventDefault(); void createDraft(); }}>
				<p class="text-xs font-bold uppercase tracking-[0.18em] text-slate-500">Buat Draft</p>
				<h2 class="mt-1 text-lg font-black text-slate-950">Siapkan Ujian</h2>
				<div class="mt-4 grid gap-3">
					<label class="space-y-1.5 text-sm font-bold text-slate-700">
						<span>Nama ujian</span>
						<input class="w-full rounded-2xl border border-slate-300 px-3 py-2 text-sm outline-none focus:border-emerald-500" bind:value={title} placeholder="Contoh: PAT Bahasa Indonesia Kelas 8" />
					</label>
					<label class="space-y-1.5 text-sm font-bold text-slate-700">
						<span>Tingkat</span>
						<select class="w-full rounded-2xl border border-slate-300 px-3 py-2 text-sm outline-none focus:border-emerald-500" bind:value={gradeLevel}>
							<option value="7">Kelas 7</option>
							<option value="8">Kelas 8</option>
							<option value="9">Kelas 9</option>
						</select>
					</label>
					<div class="grid gap-3 sm:grid-cols-2">
						<label class="space-y-1.5 text-sm font-bold text-slate-700">
							<span>Mulai</span>
							<input type="datetime-local" class="w-full rounded-2xl border border-slate-300 px-3 py-2 text-sm outline-none focus:border-emerald-500" bind:value={startsAt} />
						</label>
						<label class="space-y-1.5 text-sm font-bold text-slate-700">
							<span>Selesai</span>
							<input type="datetime-local" class="w-full rounded-2xl border border-slate-300 px-3 py-2 text-sm outline-none focus:border-emerald-500" bind:value={endsAt} />
						</label>
					</div>
				</div>
				<button type="submit" disabled={saving || !title.trim()} class="mt-4 rounded-2xl bg-emerald-700 px-4 py-2 text-sm font-black text-white shadow-sm hover:bg-emerald-800 disabled:cursor-not-allowed disabled:bg-slate-300">
					{saving ? 'Menyimpan...' : 'Simpan Draft Ujian'}
				</button>
			</form>

			<div class="rounded-3xl border border-slate-200 bg-white p-4 shadow-sm">
				<div class="mb-3 flex items-center justify-between gap-3">
					<div>
						<p class="text-xs font-bold uppercase tracking-[0.18em] text-slate-500">Daftar Draft</p>
						<h2 class="mt-1 text-lg font-black text-slate-950">Ujian terbaru</h2>
					</div>
					<button class="rounded-2xl border border-slate-300 px-3 py-2 text-xs font-black text-slate-700 hover:border-emerald-300 hover:text-emerald-800" onclick={loadExams} disabled={loading}>Muat ulang</button>
				</div>

				{#if loading}
					<p class="rounded-2xl bg-slate-100 p-4 text-sm font-bold text-slate-600">Memuat daftar asesmen...</p>
				{:else if exams.length === 0}
					<p class="rounded-2xl border border-dashed border-slate-300 bg-slate-50 p-4 text-sm leading-6 text-slate-600">Belum ada draft ujian. Buat satu draft dulu dari form di sebelah kiri.</p>
				{:else}
					<div class="divide-y divide-slate-200 overflow-hidden rounded-2xl border border-slate-200">
						{#each exams as exam (exam.id)}
							<div class={`grid gap-3 p-4 lg:grid-cols-[1fr_auto] lg:items-center ${selectedExamId === exam.id ? 'bg-emerald-50/60' : ''}`}>
								<div class="min-w-0 space-y-1">
									<div class="flex flex-wrap items-center gap-2">
										<h3 class="font-black text-slate-950">{exam.title}</h3>
										<span class="rounded-full bg-emerald-50 px-2 py-0.5 text-[11px] font-black uppercase text-emerald-800">{exam.status}</span>
										{#if selectedExamId === exam.id}
											<span class="rounded-full bg-white px-2 py-0.5 text-[11px] font-black uppercase text-emerald-800">Dipilih</span>
										{/if}
									</div>
									<p class="text-sm text-slate-600">Kelas {exam.grade_level ?? '-'} · {formatDate(exam.starts_at)}</p>
									<p class="text-xs font-bold text-slate-500">Sesi {exam.session_count ?? 0} · Ruang {exam.room_count ?? 0} · Peserta {exam.participant_count ?? 0} · Kartu {exam.card_count ?? 0}</p>
								</div>
								<div class="flex flex-wrap gap-2">
									<button class="rounded-2xl border border-slate-300 px-3 py-2 text-xs font-black text-slate-700 hover:border-emerald-300 hover:text-emerald-800" onclick={() => selectExamForRooms(exam)}>Pilih Ruang</button>
									<button class="rounded-2xl border border-slate-300 px-3 py-2 text-xs font-black text-slate-700 hover:border-emerald-300 hover:text-emerald-800 disabled:cursor-wait disabled:opacity-60" disabled={actionExamId === exam.id} onclick={() => { selectExamForRooms(exam); notice = 'Ujian dipilih. Buka panel Dokumen & Cetak di bawah untuk aksi kartu yang aman.'; }}>Dokumen</button>
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</div>
		</section>

		<section class="rounded-3xl border border-emerald-200 bg-white p-4 shadow-sm">
			<div class="flex flex-col gap-3 border-b border-slate-200 pb-3 lg:flex-row lg:items-end lg:justify-between">
				<div>
					<p class="text-xs font-bold uppercase tracking-[0.18em] text-emerald-700">Ruang Ujian</p>
					<h2 class="mt-1 text-xl font-black text-slate-950">Atur ruang dulu, peserta menyusul</h2>
					<p class="mt-1 text-sm leading-6 text-slate-600">
						Preview ini membuat ruang R01 sampai R{String(roomCount).padStart(2, '0')}. Simpan ruang tidak membuat kartu, PIN, atau token peserta.
					</p>
				</div>
				{#if selectedExam}
					<div class="rounded-2xl bg-slate-50 px-3 py-2 text-sm font-bold text-slate-700">
						Dipilih: <span class="text-slate-950">{selectedExam.title}</span>
					</div>
				{/if}
			</div>

			<div class="mt-4 grid gap-4 lg:grid-cols-[0.9fr_1.1fr]">
				<div class="rounded-2xl border border-slate-200 bg-slate-50 p-4">
					<div class="grid gap-3 sm:grid-cols-2">
						<label class="space-y-1.5 text-sm font-bold text-slate-700">
							<span>Jumlah ruang</span>
							<input type="number" min="1" max="20" class="w-full rounded-2xl border border-slate-300 px-3 py-2 text-sm outline-none focus:border-emerald-500" bind:value={roomCount} />
						</label>
						<label class="space-y-1.5 text-sm font-bold text-slate-700">
							<span>Kapasitas/ruang</span>
							<input type="number" min="1" max="50" class="w-full rounded-2xl border border-slate-300 px-3 py-2 text-sm outline-none focus:border-emerald-500" bind:value={capacityPerRoom} />
						</label>
					</div>
					<label class="mt-3 block space-y-1.5 text-sm font-bold text-slate-700">
						<span>Kebijakan ruang</span>
						<select class="w-full rounded-2xl border border-slate-300 px-3 py-2 text-sm outline-none focus:border-emerald-500" bind:value={mixPolicy}>
							<option value="mixed">Campur seimbang</option>
							<option value="class_grouped">Kelompok per kelas</option>
						</select>
					</label>
					<div class="mt-3 rounded-2xl border border-slate-200 bg-white p-3">
						<div class="flex flex-col gap-1 sm:flex-row sm:items-end sm:justify-between">
							<div>
								<p class="text-sm font-black text-slate-800">Pilih rombel peserta</p>
								<p class="text-xs font-semibold text-slate-500">Opsional. Jika dipilih, siswa aktif dari rombel ini akan didaftarkan dan masuk ruang saat disimpan.</p>
							</div>
							<span class="rounded-full bg-slate-100 px-2 py-1 text-xs font-black text-slate-600">{selectedClassCount} rombel · ±{selectedStudentEstimate} siswa</span>
						</div>
						{#if classError}
							<p class="mt-2 rounded-xl bg-red-50 px-3 py-2 text-xs font-bold text-red-700">{classError}</p>
						{:else if activeRombels.length === 0}
							<p class="mt-2 text-xs font-semibold text-slate-500">Daftar rombel belum tersedia.</p>
						{:else}
							<div class="mt-3 grid max-h-44 gap-2 overflow-y-auto sm:grid-cols-2">
								{#each activeRombels as rombel (rombel.id)}
									<label class="flex cursor-pointer items-center gap-2 rounded-xl border border-slate-200 px-3 py-2 text-xs font-bold text-slate-700 hover:border-emerald-300">
										<input type="checkbox" class="h-4 w-4 rounded border-slate-300" checked={selectedClassIds.has(rombel.id)} onchange={() => toggleClass(rombel.id)} />
										<span class="min-w-0 flex-1 truncate">{rombel.code || rombel.name}</span>
										<span class="shrink-0 text-slate-400">{rombel.total_students ?? 0}</span>
									</label>
								{/each}
							</div>
						{/if}
					</div>
					<div class="mt-3 rounded-2xl border border-dashed border-slate-300 bg-white p-3 text-sm leading-6 text-slate-600">
						<p><strong class="text-slate-800">Total kapasitas:</strong> {totalCapacity} kursi.</p>
						<p>{selectedClassCount > 0 ? `Preview memakai ±${selectedStudentEstimate} siswa dari rombel terpilih.` : 'Tanpa rombel, preview hanya memakai peserta yang sudah pernah terdaftar.'}</p>
						<p>Default UAS: 8 ruang × 30 peserta. Validasi backend membatasi 1–20 ruang dan 1–50 kursi/ruang.</p>
					</div>
					<div class="mt-4 flex flex-wrap gap-2">
						<button class="rounded-2xl border border-emerald-300 bg-white px-4 py-2 text-sm font-black text-emerald-800 hover:bg-emerald-50 disabled:cursor-not-allowed disabled:opacity-50" disabled={!selectedExam || roomBusy} onclick={() => previewRooms()}>
							{roomBusy ? 'Memproses...' : 'Preview Ruang'}
						</button>
						<button class="rounded-2xl bg-emerald-700 px-4 py-2 text-sm font-black text-white hover:bg-emerald-800 disabled:cursor-not-allowed disabled:bg-slate-300" disabled={!canSaveRooms} onclick={() => saveRooms()}>
							Simpan Ruang
						</button>
					</div>
				</div>

				<div class="rounded-2xl border border-slate-200 bg-white">
					{#if !selectedExam}
						<p class="p-4 text-sm leading-6 text-slate-600">Pilih atau buat draft ujian dulu untuk melihat preview ruang.</p>
					{:else if !roomPreview || roomPreviewExamId !== selectedExam.id}
						<div class="p-4 text-sm leading-6 text-slate-600">
							<p class="font-bold text-slate-800">Belum ada preview untuk ujian ini.</p>
							<p>Klik <strong>Preview Ruang</strong> untuk melihat R01–R{String(roomCount).padStart(2, '0')} sebelum disimpan.</p>
						</div>
					{:else}
						<div class="grid grid-cols-2 gap-2 border-b border-slate-200 p-3 text-xs font-black text-slate-600 sm:grid-cols-4">
							<div class="rounded-xl bg-slate-50 p-2">Ruang<br /><span class="text-lg text-slate-950">{roomPreview.room_count}</span></div>
							<div class="rounded-xl bg-slate-50 p-2">Kapasitas<br /><span class="text-lg text-slate-950">{roomPreview.room_count * roomPreview.capacity_per_room}</span></div>
							<div class="rounded-xl bg-slate-50 p-2">Peserta<br /><span class="text-lg text-slate-950">{roomPreview.total_participants}</span></div>
							<div class={`rounded-xl p-2 ${roomPreview.unassigned_total > 0 ? 'bg-red-50 text-red-700' : 'bg-emerald-50 text-emerald-800'}`}>Belum masuk<br /><span class="text-lg">{roomPreview.unassigned_total}</span></div>
						</div>
						<div class="max-h-[360px] overflow-y-auto">
							<div class="divide-y divide-slate-200">
								{#each roomPreview.rooms as room}
									<div class="grid gap-3 p-3 text-sm sm:grid-cols-[4.5rem_1fr_auto] sm:items-start">
										<span class="w-fit rounded-xl bg-emerald-50 px-2 py-1 text-center font-black text-emerald-800">{room.code}</span>
										<div class="min-w-0">
											<p class="truncate font-black text-slate-950">{room.name}</p>
											<p class="text-xs font-bold text-slate-500">Terisi {room.assigned_count} dari {room.capacity} · {formatGradeLevels(room.grade_levels)}</p>
											<p class="mt-1 text-xs font-semibold leading-5 text-slate-600">{formatClassSummary(room.class_summary)}</p>
										</div>
										<span class="w-fit rounded-full bg-slate-100 px-2 py-1 text-xs font-black text-slate-600">{room.capacity} kursi</span>
									</div>
								{/each}
							</div>
						</div>
						{#if hasPlacementSummary}
							<div class="border-t border-emerald-100 bg-emerald-50 p-3 text-xs font-bold leading-5 text-emerald-900">
								<p class="font-black">Ringkasan hasil penempatan siap.</p>
								<p>Peserta sudah tersimpan per ruang dan kursi. Tahap berikutnya baru cetak kartu peserta dan lembar pengawas; token/QR+PIN belum dibuat dari panel ini.</p>
							</div>
						{/if}
						<p class="border-t border-slate-200 bg-slate-50 p-3 text-xs font-bold leading-5 text-slate-600">{roomPreview.message}</p>
					{/if}
				</div>
			</div>
		</section>

		<section class="rounded-3xl border border-slate-200 bg-white p-4 shadow-sm">
			<div class="flex flex-col gap-3 border-b border-slate-200 pb-3 lg:flex-row lg:items-end lg:justify-between">
				<div>
					<p class="text-xs font-bold uppercase tracking-[0.18em] text-emerald-700">Dokumen & Cetak</p>
					<h2 class="mt-1 text-xl font-black text-slate-950">Kartu peserta dan lembar pengawas</h2>
					<p class="mt-1 max-w-3xl text-sm leading-6 text-slate-600">
						Aksi cetak dibuat terpisah dari Simpan Ruang. QR+PIN hanya dicoba diterbitkan saat panitia menekan tombol khusus dan menyetujui konfirmasi.
					</p>
				</div>
				{#if selectedExam}
					<div class="rounded-2xl bg-slate-50 px-3 py-2 text-sm font-bold text-slate-700">
						Dokumen untuk: <span class="text-slate-950">{selectedExam.title}</span>
					</div>
				{/if}
			</div>

			<div class="mt-4 divide-y divide-slate-200 overflow-hidden rounded-2xl border border-slate-200">
				{#each printDocuments as document (document.id)}
					<div class="grid gap-3 p-4 lg:grid-cols-[1fr_auto] lg:items-center">
						<div class="min-w-0">
							<div class="flex flex-wrap items-center gap-2">
								<h3 class="font-black text-slate-950">{document.title}</h3>
								<span class={`rounded-full px-2 py-0.5 text-[11px] font-black uppercase ${document.ready ? 'bg-emerald-50 text-emerald-800' : 'bg-slate-100 text-slate-600'}`}>{document.status}</span>
								<span class="rounded-full bg-slate-100 px-2 py-0.5 text-[11px] font-black text-slate-600">{document.countLabel}</span>
							</div>
							<p class="mt-1 text-sm leading-6 text-slate-600">{document.description}</p>
							<p class="mt-1 text-xs font-bold leading-5 text-slate-500">{document.detail}</p>
						</div>
						<button
							class={`w-fit rounded-2xl px-4 py-2 text-sm font-black shadow-sm disabled:cursor-not-allowed disabled:opacity-50 ${document.dangerous ? 'border border-amber-300 bg-amber-50 text-amber-900 hover:bg-amber-100' : document.ready ? 'border border-emerald-300 bg-white text-emerald-800 hover:bg-emerald-50' : 'border border-slate-200 bg-slate-50 text-slate-500'}`}
							disabled={!selectedExam || actionExamId === selectedExam.id}
							onclick={() => handlePrintDocumentAction(document)}
						>
							{actionExamId === selectedExam?.id && document.id === 'participant-cards' ? 'Memproses...' : document.primaryAction}
						</button>
					</div>
				{/each}
			</div>
		</section>

		<section class="rounded-3xl border border-slate-200 bg-white p-4 shadow-sm">
			<p class="text-xs font-bold uppercase tracking-[0.18em] text-slate-500">Batas aman</p>
			<h2 class="mt-1 text-lg font-black text-slate-950">Bank Soal tetap modul terpisah</h2>
			<p class="mt-2 text-sm leading-6 text-slate-600">
				Halaman ini hanya menyederhanakan alur ujian. Pembuatan, review, import, penerbitan, dan arsip soal tetap dilakukan di Bank Soal.
			</p>
		</section>
	</div>
</div>
