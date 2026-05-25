<script lang="ts">
	type ScheduleItem = {
		participant_id: string;
		session_title: string;
		package_title: string;
		scheduled_start: string;
		scheduled_end: string;
		room_name?: string | null;
		seat_no?: number | null;
		status: string;
		can_start: boolean;
	};
	type Student = { id: string; nis: string; nisn: string; nama: string; class_name: string; class_code: string };
	type Question = { id: string; code: string; question_text: string; question_type: string; options?: unknown };
	type ExamResult = {
		participant_id: string;
		student: { nis: string; nama: string };
		session: { title: string; scheduled_end: string };
		questions: Question[];
		total_questions: number;
		time_remaining_seconds: number;
	};

	let nisn = $state('');
	let code = $state('');
	let loading = $state(false);
	let errorMessage = $state('');
	let token = $state('');
	let student = $state<Student | null>(null);
	let schedule = $state<ScheduleItem[]>([]);
	let exam = $state<ExamResult | null>(null);
	let answers = $state<Record<string, string>>({});
	let saveStatus = $state<Record<string, string>>({});

	function data<T>(payload: { data?: T; error?: string }): T {
		if (payload.error) throw new Error(payload.error);
		return payload.data as T;
	}

	async function login() {
		loading = true;
		errorMessage = '';
		try {
			const response = await fetch('/api/cbt-portal/login', {
				method: 'POST',
				headers: { 'content-type': 'application/json' },
				body: JSON.stringify({ nisn, code })
			});
			const payload = await response.json();
			if (!response.ok) throw new Error(payload.error ?? 'Login gagal');
			const result = data<{ access_token: string; student: Student; schedule: ScheduleItem[] }>(payload);
			token = result.access_token;
			student = result.student;
			schedule = result.schedule ?? [];
			localStorage.setItem('cbt_portal_token', token);
		} catch (error) {
			errorMessage = error instanceof Error ? error.message : 'Login gagal';
		} finally {
			loading = false;
		}
	}

	async function refreshSchedule() {
		if (!token) return;
		const response = await fetch('/api/cbt-portal/schedule', { headers: { authorization: `Bearer ${token}` } });
		const payload = await response.json();
		if (!response.ok) throw new Error(payload.error ?? 'Jadwal tidak tersedia');
		schedule = data<{ schedule: ScheduleItem[] }>(payload).schedule ?? [];
	}

	function fingerprint() {
		const raw = [navigator.userAgent, navigator.language, screen.width, screen.height, Intl.DateTimeFormat().resolvedOptions().timeZone].join('|');
		return btoa(unescape(encodeURIComponent(raw))).slice(0, 128);
	}

	async function startExam(item: ScheduleItem) {
		loading = true;
		errorMessage = '';
		try {
			const response = await fetch(`/api/cbt-portal/participants/${encodeURIComponent(item.participant_id)}/start`, {
				method: 'POST',
				headers: { authorization: `Bearer ${token}`, 'content-type': 'application/json' },
				body: JSON.stringify({ device_fingerprint: fingerprint(), browser_fingerprint: fingerprint() })
			});
			const payload = await response.json();
			if (!response.ok) throw new Error(payload.error ?? 'Ujian belum dapat dimulai');
			exam = data<ExamResult>(payload);
			answers = {};
		} catch (error) {
			errorMessage = error instanceof Error ? error.message : 'Ujian belum dapat dimulai';
		} finally {
			loading = false;
		}
	}


	async function saveAnswer(questionID: string) {
		if (!token || !exam) return;
		saveStatus = { ...saveStatus, [questionID]: 'Menyimpan...' };
		try {
			const response = await fetch(`/api/cbt-portal/participants/${encodeURIComponent(exam.participant_id)}/answer`, {
				method: 'POST',
				headers: { authorization: `Bearer ${token}`, 'content-type': 'application/json' },
				body: JSON.stringify({ question_id: questionID, answer: answers[questionID] ?? '' })
			});
			const payload = await response.json();
			if (!response.ok) throw new Error(payload.error ?? 'Jawaban gagal disimpan');
			saveStatus = { ...saveStatus, [questionID]: 'Tersimpan' };
		} catch (error) {
			saveStatus = { ...saveStatus, [questionID]: error instanceof Error ? error.message : 'Gagal disimpan' };
		}
	}

	async function submitExam() {
		if (!token || !exam) return;
		const ok = confirm('Kumpulkan ujian sekarang? Setelah dikumpulkan jawaban tidak dapat diubah.');
		if (!ok) return;
		loading = true;
		errorMessage = '';
		try {
			for (const question of exam.questions) {
				await saveAnswer(question.id);
			}
			const response = await fetch(`/api/cbt-portal/participants/${encodeURIComponent(exam.participant_id)}/submit`, {
				method: 'POST',
				headers: { authorization: `Bearer ${token}` }
			});
			const payload = await response.json();
			if (!response.ok) throw new Error(payload.error ?? 'Ujian gagal dikumpulkan');
			exam = null;
			await refreshSchedule();
		} catch (error) {
			errorMessage = error instanceof Error ? error.message : 'Ujian gagal dikumpulkan';
		} finally {
			loading = false;
		}
	}


	function logout() {
		token = '';
		student = null;
		schedule = [];
		exam = null;
		answers = {};
		saveStatus = {};
		localStorage.removeItem('cbt_portal_token');
	}

	function formatDate(value: string) {
		if (!value) return '-';
		return new Intl.DateTimeFormat('id-ID', { dateStyle: 'medium', timeStyle: 'short', timeZone: 'Asia/Makassar' }).format(new Date(value));
	}

	$effect(() => {
		const saved = localStorage.getItem('cbt_portal_token');
		if (saved && !token) {
			token = saved;
			refreshSchedule().catch(() => logout());
		}
	});
</script>

<svelte:head>
	<title>Portal CBT MTsN 2 Kolaka Utara</title>
</svelte:head>

<main class="min-h-screen bg-emerald-950 text-slate-100">
	<section class="mx-auto flex min-h-screen max-w-5xl flex-col px-4 py-6 sm:px-6 lg:px-8">
		<header class="mb-6 rounded-3xl border border-emerald-400/20 bg-white/10 p-5 shadow-2xl shadow-black/20">
			<p class="text-sm font-semibold uppercase tracking-[0.35em] text-amber-200">MTsN 2 Kolaka Utara</p>
			<div class="mt-3 flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
				<div>
					<h1 class="text-3xl font-bold tracking-tight sm:text-4xl">Portal CBT Siswa</h1>
					<p class="mt-2 max-w-2xl text-sm text-emerald-50/80">Masuk khusus simulasi CBT. Isi NISN pada kolom NISN dan ulangi NISN pada kolom kode masuk.</p>
				</div>
				{#if token}
					<button class="rounded-full border border-white/20 px-4 py-2 text-sm font-semibold hover:bg-white/10" onclick={logout}>Keluar</button>
				{/if}
			</div>
		</header>

		{#if errorMessage}
			<div class="mb-4 rounded-2xl border border-red-300/30 bg-red-500/20 p-4 text-sm text-red-50" role="alert">{errorMessage}</div>
		{/if}

		{#if exam}
			<section class="rounded-3xl bg-slate-50 p-5 text-slate-950 shadow-2xl">
				<div class="mb-5 flex flex-col gap-2 border-b pb-4 sm:flex-row sm:items-start sm:justify-between">
					<div>
						<p class="text-sm font-medium text-emerald-700">Sedang mengerjakan</p>
						<h2 class="text-2xl font-bold">{exam.session.title}</h2>
						<p class="text-sm text-slate-600">Peserta: {exam.student.nama} · {exam.total_questions} soal</p>
					</div>
					<button class="rounded-xl bg-slate-900 px-4 py-2 text-sm font-semibold text-white" onclick={() => (exam = null)}>Kembali</button>
				</div>
				<div class="space-y-4">
					{#each exam.questions as question, index (question.id)}
						<article class="rounded-2xl border bg-white p-4">
							<p class="text-xs font-semibold uppercase text-slate-500">Soal {index + 1} · {question.question_type}</p>
							<div class="mt-2 whitespace-pre-wrap text-base font-medium">{@html question.question_text || question.code}</div>
							<textarea class="mt-3 min-h-24 w-full rounded-xl border p-3 text-sm" placeholder="Tulis jawaban di sini..." bind:value={answers[question.id]} onblur={() => saveAnswer(question.id)}></textarea>
							<div class="mt-2 flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
								<p class="text-xs text-slate-500">{saveStatus[question.id] ?? 'Belum disimpan'}</p>
								<button class="rounded-xl border px-4 py-2 text-sm font-semibold" onclick={() => saveAnswer(question.id)}>Simpan Jawaban</button>
							</div>
						</article>
					{/each}
				</div>
				<div class="mt-5 flex flex-col gap-3 rounded-2xl bg-amber-50 p-4 text-sm text-amber-900 sm:flex-row sm:items-center sm:justify-between">
					<p>Periksa jawaban sebelum dikumpulkan. Jawaban disimpan saat tombol Simpan diklik atau kolom jawaban ditinggalkan.</p>
					<button class="rounded-xl bg-emerald-700 px-4 py-2 font-semibold text-white" disabled={loading} onclick={submitExam}>Kumpulkan Ujian</button>
				</div>
			</section>
		{:else if token}
			<section class="rounded-3xl bg-white p-5 text-slate-950 shadow-2xl">
				<div class="mb-4 flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
					<div>
						<p class="text-sm text-slate-500">Login sebagai</p>
						<h2 class="text-2xl font-bold">{student?.nama ?? 'Siswa'}</h2>
						<p class="text-sm text-slate-600">{student?.class_name ?? 'Kelas'} · NISN {student?.nisn ?? '-'}</p>
					</div>
					<button class="rounded-xl border px-4 py-2 text-sm font-semibold" onclick={() => refreshSchedule().catch((e) => (errorMessage = e.message))}>Muat ulang</button>
				</div>
				<h3 class="mb-3 text-lg font-semibold">Ujian Saya</h3>
				{#if schedule.length === 0}
					<p class="rounded-2xl border border-dashed p-5 text-sm text-slate-600">Belum ada sesi CBT simulasi yang aktif untuk akun ini.</p>
				{:else}
					<div class="grid gap-3">
						{#each schedule as item (item.participant_id)}
							<article class="rounded-2xl border p-4">
								<div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
									<div>
										<p class="text-xs font-semibold uppercase text-emerald-700">{item.status}</p>
										<h4 class="text-lg font-bold">{item.session_title}</h4>
										<p class="text-sm text-slate-600">{item.package_title}</p>
										<p class="mt-1 text-xs text-slate-500">{formatDate(item.scheduled_start)} – {formatDate(item.scheduled_end)} · {item.room_name ?? 'Ruang belum ditetapkan'}</p>
									</div>
									<button class="rounded-xl px-4 py-2 text-sm font-semibold {item.can_start ? 'bg-emerald-700 text-white' : 'bg-slate-200 text-slate-500'}" disabled={!item.can_start || loading} onclick={() => startExam(item)}>Mulai Ujian</button>
								</div>
							</article>
						{/each}
					</div>
				{/if}
			</section>
		{:else}
			<form class="mx-auto w-full max-w-md rounded-3xl bg-white p-6 text-slate-950 shadow-2xl" onsubmit={(event) => { event.preventDefault(); login(); }}>
				<h2 class="text-2xl font-bold">Login CBT</h2>
				<p class="mt-1 text-sm text-slate-600">Untuk simulasi, kode masuk sama dengan NISN.</p>
				<label class="mt-5 block text-sm font-semibold">
					NISN
					<input class="mt-2 w-full rounded-xl border px-4 py-3 text-base" inputmode="numeric" autocomplete="username" bind:value={nisn} required />
				</label>
				<label class="mt-4 block text-sm font-semibold">
					Kode Masuk
					<input class="mt-2 w-full rounded-xl border px-4 py-3 text-base" inputmode="numeric" autocomplete="current-password" bind:value={code} required />
				</label>
				<button class="mt-6 w-full rounded-xl bg-emerald-700 px-4 py-3 font-bold text-white disabled:opacity-60" disabled={loading}>{loading ? 'Memproses...' : 'Masuk'}</button>
			</form>
		{/if}
	</section>
</main>

<style>
	:global(body) { margin: 0; font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; }
	.min-h-screen { min-height: 100vh; }
	.bg-emerald-950 { background: radial-gradient(circle at top left, #047857 0, #064e3b 36%, #020617 100%); }
	.text-slate-100 { color: #f1f5f9; }
	.mx-auto { margin-left: auto; margin-right: auto; }
	.flex { display: flex; }
	.grid { display: grid; }
	.w-full { width: 100%; }
	.max-w-5xl { max-width: 64rem; }
	.max-w-md { max-width: 28rem; }
	.flex-col { flex-direction: column; }
	.gap-2 { gap: .5rem; } .gap-3 { gap: .75rem; }
	.space-y-4 > * + * { margin-top: 1rem; }
	.rounded-xl { border-radius: .75rem; } .rounded-2xl { border-radius: 1rem; } .rounded-3xl { border-radius: 1.5rem; } .rounded-full { border-radius: 9999px; }
	.border { border: 1px solid #e2e8f0; } .border-b { border-bottom: 1px solid #e2e8f0; }
	.border-dashed { border-style: dashed; }
	.bg-white { background: white; } .bg-slate-50 { background: #f8fafc; } .bg-slate-900 { background: #0f172a; } .bg-emerald-700 { background: #047857; } .bg-slate-200 { background: #e2e8f0; } .bg-amber-50 { background: #fffbeb; }
	.bg-white\/10 { background: rgba(255,255,255,.1); } .bg-red-500\/20 { background: rgba(239,68,68,.2); }
	.p-3 { padding: .75rem; } .p-4 { padding: 1rem; } .p-5 { padding: 1.25rem; } .p-6 { padding: 1.5rem; }
	.px-4 { padding-left: 1rem; padding-right: 1rem; }
	.py-2 { padding-top: .5rem; padding-bottom: .5rem; } .py-3 { padding-top: .75rem; padding-bottom: .75rem; } .py-6 { padding-top: 1.5rem; padding-bottom: 1.5rem; }
	.pb-4 { padding-bottom: 1rem; }
	.mb-3 { margin-bottom: .75rem; } .mb-4 { margin-bottom: 1rem; } .mb-5 { margin-bottom: 1.25rem; } .mb-6 { margin-bottom: 1.5rem; }
	.mt-1 { margin-top: .25rem; } .mt-2 { margin-top: .5rem; } .mt-3 { margin-top: .75rem; } .mt-4 { margin-top: 1rem; } .mt-5 { margin-top: 1.25rem; } .mt-6 { margin-top: 1.5rem; }
	.text-xs { font-size: .75rem; } .text-sm { font-size: .875rem; } .text-base { font-size: 1rem; } .text-lg { font-size: 1.125rem; } .text-2xl { font-size: 1.5rem; } .text-3xl { font-size: 1.875rem; }
	.font-medium { font-weight: 500; } .font-semibold { font-weight: 600; } .font-bold { font-weight: 700; }
	.uppercase { text-transform: uppercase; } .tracking-tight { letter-spacing: -.025em; } .tracking-\[0\.35em\] { letter-spacing: .35em; }
	.text-white { color: white; } .text-slate-950 { color: #020617; } .text-slate-600 { color: #475569; } .text-slate-500 { color: #64748b; } .text-emerald-700 { color: #047857; } .text-amber-200 { color: #fde68a; } .text-amber-900 { color: #78350f; } .text-red-50 { color: #fef2f2; }
	.shadow-2xl { box-shadow: 0 25px 50px -12px rgba(0,0,0,.25); }
	.min-h-24 { min-height: 6rem; }
	.whitespace-pre-wrap { white-space: pre-wrap; }
	button { cursor: pointer; border: 0; } button:disabled { cursor: not-allowed; }
	input, textarea { box-sizing: border-box; font: inherit; }
	@media (min-width: 640px) { .sm\:flex-row { flex-direction: row; } .sm\:items-center { align-items: center; } .sm\:items-start { align-items: flex-start; } .sm\:items-end { align-items: flex-end; } .sm\:justify-between { justify-content: space-between; } .sm\:text-4xl { font-size: 2.25rem; } .sm\:px-6 { padding-left: 1.5rem; padding-right: 1.5rem; } }
	@media (min-width: 1024px) { .lg\:px-8 { padding-left: 2rem; padding-right: 2rem; } }
</style>
