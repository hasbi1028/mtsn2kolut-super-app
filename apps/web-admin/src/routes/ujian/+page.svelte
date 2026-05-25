<script lang="ts">
	import { browser } from '$app/environment';
	import { page } from '$app/stores';

	type Question = {
		id: string;
		type: 'multiple_choice' | 'essay' | 'short_answer';
		text: string;
		options?: Array<{ label: string; text: string }>;
	};
	type ExamPayload = {
		student?: { nama?: string; nis?: string };
		session?: { title?: string; scheduled_end?: string };
		questions?: Question[];
		total_questions?: number;
		time_remaining_seconds?: number;
	};

	const demoQuestions: Question[] = [
		{
			id: 'demo-web-1',
			type: 'multiple_choice',
			text: 'Mode Browser Darurat sebaiknya dipakai ketika...',
			options: [
				{ label: 'A', text: 'APK Flutter tidak dapat dipakai pada perangkat tertentu' },
				{ label: 'B', text: 'Semua siswa ingin bebas membuka tab lain' },
				{ label: 'C', text: 'Panitia tidak ingin memakai token ruang' },
				{ label: 'D', text: 'Ujian sudah selesai' }
			]
		},
		{
			id: 'demo-web-2',
			type: 'short_answer',
			text: 'Tuliskan satu hal yang wajib dilakukan pengawas saat Browser Darurat aktif.'
		},
		{
			id: 'demo-web-3',
			type: 'essay',
			text: 'Jelaskan perbedaan singkat antara jalur utama Flutter APK dan Mode Darurat / Browser.'
		}
	];

	let examToken = $state('');
	let roomToken = $state('');
	let deviceFingerprint = $state('');
	let loading = $state(false);
	let errorMessage = $state('');
	let payload = $state<ExamPayload | null>(null);
	let answers = $state<Record<string, string>>({});
	let pendingAnswers = $state<Record<string, string>>({});
	let submitted = $state(false);
	let telemetry = $state<string[]>([]);

	let demoMode = $derived($page.url.searchParams.get('demo') === '1');
	let questions = $derived(payload?.questions ?? []);
	let answeredCount = $derived(Object.values(answers).filter((answer) => answer.trim().length > 0).length);

	function addTelemetry(label: string) {
		const stamp = new Intl.DateTimeFormat('id-ID', { timeStyle: 'medium', timeZone: 'Asia/Makassar' }).format(new Date());
		telemetry = [`${stamp} — ${label}`, ...telemetry].slice(0, 8);
	}

	function makeFingerprint() {
		if (!browser) return 'server-render';
		const raw = [navigator.userAgent, navigator.language, screen.width, screen.height, Intl.DateTimeFormat().resolvedOptions().timeZone].join('|');
		return btoa(unescape(encodeURIComponent(raw))).slice(0, 128);
	}

	async function login() {
		loading = true;
		errorMessage = '';
		submitted = false;
		answers = {};
		pendingAnswers = {};
		try {
			if (demoMode) {
				payload = {
					student: { nama: 'Siswa Demo Browser', nis: 'DEMO-WEB' },
					session: { title: 'MODE DEMO Browser Darurat — Data Contoh' },
					questions: demoQuestions,
					total_questions: demoQuestions.length,
					time_remaining_seconds: 45 * 60
				};
				addTelemetry('Demo dibuka — tidak ada API produksi yang dipanggil');
				return;
			}
			const fingerprint = deviceFingerprint.trim() || makeFingerprint();
			deviceFingerprint = fingerprint;
			const response = await fetch('/api/exam/login', {
				method: 'POST',
				headers: { 'content-type': 'application/json' },
				body: JSON.stringify({
					token: examToken.trim(),
					room_token: roomToken.trim(),
					device_fingerprint: fingerprint,
					browser_fingerprint: fingerprint,
					client_type: 'web_fallback'
				})
			});
			const body = await response.json();
			if (!response.ok) throw new Error(body.error ?? body.message ?? 'Login Browser Darurat gagal');
			payload = body.data ?? body;
			addTelemetry('Login Browser Darurat berhasil');
		} catch (error) {
			errorMessage = error instanceof Error ? error.message : 'Login gagal';
		} finally {
			loading = false;
		}
	}

	async function saveAnswer(questionId: string, answer: string) {
		answers = { ...answers, [questionId]: answer };
		if (demoMode) {
			addTelemetry(`Demo menyimpan jawaban ${questionId} secara lokal`);
			return;
		}
		pendingAnswers = { ...pendingAnswers, [questionId]: answer };
		try {
			const response = await fetch('/api/exam/answer', {
				method: 'POST',
				headers: {
					'content-type': 'application/json',
					'x-exam-token': examToken.trim(),
					'x-device-fingerprint': deviceFingerprint.trim()
				},
				body: JSON.stringify({ question_id: questionId, answer })
			});
			if (!response.ok) throw new Error('Jawaban belum tersinkron');
			const next = { ...pendingAnswers };
			delete next[questionId];
			pendingAnswers = next;
			addTelemetry(`Jawaban ${questionId} tersinkron`);
		} catch {
			addTelemetry(`Jawaban ${questionId} aman lokal, menunggu sinkron`);
		}
	}

	async function submitExam() {
		if (Object.keys(pendingAnswers).length > 0) {
			errorMessage = 'Masih ada jawaban yang belum tersinkron. Coba simpan ulang sebelum kumpulkan.';
			return;
		}
		if (!confirm('Kumpulkan ujian sekarang?')) return;
		loading = true;
		errorMessage = '';
		try {
			if (!demoMode) {
				const response = await fetch('/api/exam/submit', {
					method: 'POST',
					headers: {
						'content-type': 'application/json',
						'x-exam-token': examToken.trim(),
						'x-device-fingerprint': deviceFingerprint.trim()
					},
					body: '{}'
				});
				if (!response.ok) throw new Error('Ujian gagal dikumpulkan');
			}
			submitted = true;
			addTelemetry(demoMode ? 'Demo selesai lokal' : 'Ujian dikumpulkan');
		} catch (error) {
			errorMessage = error instanceof Error ? error.message : 'Submit gagal';
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		if (!browser) return;
		deviceFingerprint = deviceFingerprint || makeFingerprint();
		const onBlur = () => addTelemetry(demoMode ? 'Demo: window blur' : 'Window blur tercatat');
		const onFocus = () => addTelemetry(demoMode ? 'Demo: window focus' : 'Window focus tercatat');
		const onOffline = () => addTelemetry('Perangkat offline');
		const onOnline = () => addTelemetry('Perangkat online kembali');
		window.addEventListener('blur', onBlur);
		window.addEventListener('focus', onFocus);
		window.addEventListener('offline', onOffline);
		window.addEventListener('online', onOnline);
		return () => {
			window.removeEventListener('blur', onBlur);
			window.removeEventListener('focus', onFocus);
			window.removeEventListener('offline', onOffline);
			window.removeEventListener('online', onOnline);
		};
	});
</script>

<svelte:head>
	<title>{demoMode ? 'MODE DEMO Browser CBT' : 'Browser Darurat CBT'} — MTsN 2 Kolaka Utara</title>
</svelte:head>

<main class="min-h-screen bg-slate-950 px-4 py-5 text-slate-100">
	<section class="mx-auto max-w-5xl space-y-4">
		<header class="rounded-2xl border border-emerald-300/20 bg-white/10 p-4">
			<p class="text-xs font-semibold uppercase tracking-[0.3em] text-amber-200">MTsN 2 Kolaka Utara</p>
			<h1 class="mt-2 text-2xl font-bold">{demoMode ? 'MODE DEMO Browser Darurat' : 'Mode Darurat / Browser'}</h1>
			<p class="mt-2 max-w-3xl text-sm text-slate-200/85">
				{demoMode
					? 'Data contoh lokal untuk tes cepat tampilan dan alur. Jawaban tidak dikirim ke server dan tidak menjadi nilai.'
					: 'Fallback hanya ketika APK Flutter tidak dapat dipakai. Pengawasan fisik wajib dan akses tetap memakai Token Ujian + Token Ruang.'}
			</p>
		</header>

		{#if errorMessage}
			<div class="rounded-xl border border-red-300/30 bg-red-500/20 p-3 text-sm text-red-50" role="alert">{errorMessage}</div>
		{/if}

		{#if submitted}
			<section class="rounded-2xl bg-white p-5 text-slate-950">
				<h2 class="text-xl font-bold">Selesai</h2>
				<p class="mt-2 text-sm text-slate-600">{demoMode ? 'Mode demo selesai lokal.' : 'Ujian telah dikumpulkan.'}</p>
				<button class="mt-4 rounded-lg border px-4 py-2 text-sm font-semibold" onclick={() => { payload = null; submitted = false; }}>Kembali</button>
			</section>
		{:else if payload}
			<section class="rounded-2xl bg-white p-4 text-slate-950">
				<div class="flex flex-col gap-3 border-b pb-3 sm:flex-row sm:items-center sm:justify-between">
					<div>
						<p class="text-xs font-semibold uppercase text-emerald-700">{demoMode ? 'MODE DEMO — DATA CONTOH' : 'Browser Darurat — Pengawasan Wajib'}</p>
						<h2 class="text-xl font-bold">{payload.session?.title ?? 'Sesi Ujian'}</h2>
						<p class="text-sm text-slate-600">{payload.student?.nama ?? 'Siswa'} · {answeredCount}/{questions.length} terjawab</p>
					</div>
					<button class="rounded-lg border px-4 py-2 text-sm font-semibold" onclick={() => (payload = null)}>Keluar</button>
				</div>
				<div class="mt-4 space-y-3">
					{#each questions as question, index (question.id)}
						<article class="rounded-xl border p-4">
							<p class="text-xs font-semibold uppercase text-slate-500">Soal {index + 1} · {question.type}</p>
							<p class="mt-2 font-medium">{question.text}</p>
							{#if question.options?.length}
								<div class="mt-3 grid gap-2">
									{#each question.options as option}
										<label class="flex gap-2 rounded-lg border p-2 text-sm">
											<input type="radio" name={question.id} value={option.label} checked={answers[question.id] === option.label} onchange={() => saveAnswer(question.id, option.label)} />
											<span><b>{option.label}.</b> {option.text}</span>
										</label>
									{/each}
								</div>
							{:else}
								<textarea class="mt-3 min-h-24 w-full rounded-lg border p-3 text-sm" placeholder="Tulis jawaban..." value={answers[question.id] ?? ''} onblur={(event) => saveAnswer(question.id, event.currentTarget.value)}></textarea>
							{/if}
							<p class="mt-2 text-xs text-slate-500">{pendingAnswers[question.id] ? 'Aman lokal, menunggu sinkron' : answers[question.id] ? 'Tersimpan' : 'Belum dijawab'}</p>
						</article>
					{/each}
				</div>
				<div class="mt-4 rounded-xl bg-amber-50 p-3 text-sm text-amber-900">
					Browser Darurat tidak setara keamanan APK Flutter. Gunakan hanya atas arahan panitia/pengawas.
				</div>
				<button class="mt-4 w-full rounded-xl bg-emerald-700 px-4 py-3 font-semibold text-white disabled:opacity-60" disabled={loading} onclick={submitExam}>Kumpulkan</button>
			</section>
		{:else}
			<section class="rounded-2xl bg-white p-5 text-slate-950">
				<h2 class="text-xl font-bold">{demoMode ? 'Mulai DEMO' : 'Masuk Browser Darurat'}</h2>
				<p class="mt-1 text-sm text-slate-600">{demoMode ? 'Tidak perlu token. Klik mulai untuk memakai soal contoh.' : 'Isi token ujian dan token ruang dari pengawas.'}</p>
				{#if !demoMode}
					<div class="mt-4 grid gap-3 sm:grid-cols-2">
						<label class="space-y-1 text-sm font-medium">Token Ujian<input class="w-full rounded-lg border px-3 py-2" bind:value={examToken} /></label>
						<label class="space-y-1 text-sm font-medium">Token Ruang<input class="w-full rounded-lg border px-3 py-2" bind:value={roomToken} /></label>
					</div>
				{/if}
				<button class="mt-4 w-full rounded-xl bg-emerald-700 px-4 py-3 font-semibold text-white disabled:opacity-60" disabled={loading} onclick={login}>{loading ? 'Memproses...' : demoMode ? 'Mulai Mode DEMO' : 'Masuk Ujian'}</button>
				{#if !demoMode}
					<a class="mt-3 block text-center text-sm text-emerald-700 underline" href="/ujian?demo=1">Buka Mode DEMO untuk tes cepat</a>
				{/if}
			</section>
		{/if}

		<section class="rounded-2xl border border-white/10 bg-white/5 p-4 text-sm">
			<h2 class="font-semibold">Log perangkat lokal</h2>
			{#if telemetry.length === 0}
				<p class="mt-2 text-slate-300">Belum ada event.</p>
			{:else}
				<ul class="mt-2 space-y-1 text-slate-200">
					{#each telemetry as item}
						<li>• {item}</li>
					{/each}
				</ul>
			{/if}
		</section>
	</section>
</main>
