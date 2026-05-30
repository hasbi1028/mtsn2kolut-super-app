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

	const steps = [
		{
			label: 'Siapkan Ujian',
			detail: 'Buat draft ujian dulu: nama, tingkat kelas, tanggal dan jam. Paket soal tetap di Bank Soal.',
			status: 'Aktif'
		},
		{
			label: 'Atur Ruang Awal',
			detail: 'Klik Siapkan Ruang untuk membuat fondasi sesi dan Ruang Ujian 1. Pembagian otomatis peserta menyusul pada tahap berikutnya.',
			status: 'Fondasi'
		},
		{
			label: 'Cetak Kartu & Pengawas',
			detail: 'Tombol sudah aman: belum menerbitkan token mentah sampai peserta ujian tersambung.',
			status: 'Aman'
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

	onMount(() => {
		void loadExams();
	});

	async function loadExams() {
		loading = true;
		error = '';
		try {
			exams = await fetch('/api/asesmen/exams?limit=12').then((response) =>
				readClientApiData<Exam[]>(response, 'Daftar asesmen belum dapat dimuat')
			);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Daftar asesmen belum dapat dimuat';
			exams = [];
		} finally {
			loading = false;
		}
	}

	function toISO(value: string) {
		if (!value) return undefined;
		const date = new Date(value);
		return Number.isNaN(date.getTime()) ? undefined : date.toISOString();
	}

	async function createDraft() {
		const trimmedTitle = title.trim();
		if (!trimmedTitle || saving) return;
		saving = true;
		error = '';
		notice = '';
		try {
			await fetch('/api/asesmen/exams', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					title: trimmedTitle,
					grade_level: Number(gradeLevel),
					starts_at: toISO(startsAt),
					ends_at: toISO(endsAt)
				})
			}).then((response) => readClientApiData<Exam>(response, 'Draft ujian belum dapat dibuat'));
			notice = 'Draft ujian tersimpan.';
			title = '';
			await loadExams();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Draft ujian belum dapat dibuat';
		} finally {
			saving = false;
		}
	}

	async function runExamAction(exam: Exam, action: 'prepare-rooms' | 'issue-cards') {
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
							Alur sederhana untuk panitia: buat draft ujian, siapkan ruang awal, lalu lanjutkan peserta dan kartu saat fondasi backend berikutnya siap.
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
				<p class="text-sm font-semibold text-slate-500">Produksi awal — sudah tersambung ke backend asesmen.</p>
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
							<div class="grid gap-3 p-4 lg:grid-cols-[1fr_auto] lg:items-center">
								<div class="min-w-0 space-y-1">
									<div class="flex flex-wrap items-center gap-2">
										<h3 class="font-black text-slate-950">{exam.title}</h3>
										<span class="rounded-full bg-emerald-50 px-2 py-0.5 text-[11px] font-black uppercase text-emerald-800">{exam.status}</span>
									</div>
									<p class="text-sm text-slate-600">Kelas {exam.grade_level ?? '-'} · {formatDate(exam.starts_at)}</p>
									<p class="text-xs font-bold text-slate-500">Sesi {exam.session_count ?? 0} · Ruang {exam.room_count ?? 0} · Peserta {exam.participant_count ?? 0} · Kartu {exam.card_count ?? 0}</p>
								</div>
								<div class="flex flex-wrap gap-2">
									<button class="rounded-2xl border border-slate-300 px-3 py-2 text-xs font-black text-slate-700 hover:border-emerald-300 hover:text-emerald-800 disabled:cursor-wait disabled:opacity-60" disabled={actionExamId === exam.id} onclick={() => runExamAction(exam, 'prepare-rooms')}>Siapkan Ruang</button>
									<button class="rounded-2xl border border-slate-300 px-3 py-2 text-xs font-black text-slate-700 hover:border-emerald-300 hover:text-emerald-800 disabled:cursor-wait disabled:opacity-60" disabled={actionExamId === exam.id} onclick={() => runExamAction(exam, 'issue-cards')}>Cek Kartu</button>
								</div>
							</div>
						{/each}
					</div>
				{/if}
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
