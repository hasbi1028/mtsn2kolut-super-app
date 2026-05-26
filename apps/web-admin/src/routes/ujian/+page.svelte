<script lang="ts">
	import { browser } from '$app/environment';
	import { page } from '$app/stores';

	type Question = {
		id: string;
		type: 'multiple_choice' | 'essay' | 'short_answer';
		text: string;
		options?: Array<{ label: string; text: string }>;
	};
	type StudentIdentity = { nama?: string; name?: string; nis?: string; class_code?: string; class_name?: string; room_name?: string; seat_no?: number | string | null };
	type SessionIdentity = { title?: string; subject?: string; scheduled_start?: string; scheduled_end?: string; status?: string; started?: boolean; is_started?: boolean; waiting?: boolean };
	type ExamPayload = {
		student?: StudentIdentity;
		participant?: StudentIdentity;
		session?: SessionIdentity;
		access_token?: string;
		exam_token?: string;
		token?: string;
		room_token?: string;
		questions?: Question[];
		total_questions?: number;
		time_remaining_seconds?: number;
	};
	type PortalStep = 'login' | 'confirm' | 'waiting' | 'exam';

	const demoQuestions: Question[] = [
		{
			id: 'demo-informatika-1',
			type: 'multiple_choice',
			text: 'Dalam Informatika, data yang dikirim dari browser ke server melalui jaringan sebaiknya dilindungi dengan...',
			options: [
				{ label: 'A', text: 'Enkripsi dan autentikasi yang sesuai' },
				{ label: 'B', text: 'Membagikan PIN kepada teman' },
				{ label: 'C', text: 'Menonaktifkan semua pembaruan keamanan' },
				{ label: 'D', text: 'Menyimpan kata sandi di catatan umum' }
			]
		},
		{
			id: 'demo-informatika-2',
			type: 'short_answer',
			text: 'Tuliskan satu contoh jejak digital yang dapat tercatat saat peserta memakai Portal Ujian Web.'
		},
		{
			id: 'demo-informatika-3',
			type: 'essay',
			text: 'Jelaskan mengapa sinkronisasi jawaban bertahap penting dalam sistem ujian berbasis web.'
		}
	];

	let cardToken = $state('');
	let pin = $state('');
	let examToken = $state('');
	let roomToken = $state('');
	let showLegacyTokenLogin = $state(false);
	let deviceFingerprint = $state('');
	let loading = $state(false);
	let errorMessage = $state('');
	let payload = $state<ExamPayload | null>(null);
	let portalStep = $state<PortalStep>('login');
	let portalAuthToken = $state('');
	let activeParticipantId = $state('');
	let authenticatedByCard = $state(false);
	let answers = $state<Record<string, string>>({});
	let pendingAnswers = $state<Record<string, string>>({});
	let submitted = $state(false);
	let telemetry = $state<string[]>([]);

	let demoMode = $derived($page.url.searchParams.get('demo') === '1');
	let queryCard = $derived($page.url.searchParams.get('card') ?? $page.url.searchParams.get('token') ?? '');
	let questions = $derived(payload?.questions ?? []);
	let answeredCount = $derived(Object.values(answers).filter((answer) => answer.trim().length > 0).length);
	let student = $derived(payload?.student ?? payload?.participant ?? {});
	let session = $derived(payload?.session ?? {});
	let studentName = $derived(student.nama ?? student.name ?? 'Peserta');
	let examAccessToken = $derived(payload?.access_token ?? payload?.exam_token ?? payload?.token ?? examToken);

	function friendlyError(error: unknown) {
		const message = error instanceof Error ? error.message : String(error || '');
		const lower = message.toLowerCase();
		if (lower.includes('pin') || lower.includes('invalid') || lower.includes('token') || lower.includes('unauthorized')) return 'Kartu ujian tidak cocok atau PIN salah. Silakan panggil pengawas.';
		if (lower.includes('not active') || lower.includes('belum') || lower.includes('session')) return 'Ujian belum dibuka oleh pengawas. Silakan tunggu di ruangan.';
		if (lower.includes('network') || lower.includes('fetch') || lower.includes('502')) return 'Koneksi ke layanan ujian sedang bermasalah. Coba lagi atau panggil pengawas.';
		return message || 'Masuk ujian gagal. Silakan panggil pengawas.';
	}

	function addTelemetry(label: string) {
		const stamp = new Intl.DateTimeFormat('id-ID', { timeStyle: 'medium', timeZone: 'Asia/Makassar' }).format(new Date());
		telemetry = [`${stamp} — ${label}`, ...telemetry].slice(0, 8);
	}

	function makeFingerprint() {
		if (!browser) return 'server-render';
		const raw = [navigator.userAgent, navigator.language, screen.width, screen.height, Intl.DateTimeFormat().resolvedOptions().timeZone].join('|');
		return btoa(unescape(encodeURIComponent(raw))).slice(0, 128);
	}

	function resetExamState() {
		submitted = false;
		answers = {};
		pendingAnswers = {};
		activeParticipantId = '';
		portalAuthToken = '';
		authenticatedByCard = false;
	}

	function normalizePortalPayload(body: unknown): ExamPayload {
		const wrapped = body as { data?: ExamPayload; payload?: ExamPayload };
		return wrapped?.data ?? wrapped?.payload ?? (body as ExamPayload);
	}

	function normalizeCardVerifyPayload(body: unknown): ExamPayload {
		const wrapped = body as { data?: { access_token?: string; card?: Record<string, unknown> }; access_token?: string; card?: Record<string, unknown> };
		const source = wrapped.data ?? wrapped;
		const card = (source.card ?? {}) as Record<string, unknown>;
		portalAuthToken = String(source.access_token ?? '');
		authenticatedByCard = true;
		activeParticipantId = String(card.participant_id ?? '');
		return {
			access_token: portalAuthToken,
			student: {
				nama: String(card.student_name ?? 'Peserta'),
				nis: String(card.nis ?? ''),
				class_name: String(card.class_name ?? ''),
				room_name: String(card.room_name ?? ''),
				seat_no: (card.seat_no as number | string | null | undefined) ?? null
			},
			session: {
				title: String(card.session_title ?? 'Sesi Ujian'),
				subject: String(card.package_title ?? ''),
				status: String(card.session_status ?? ''),
				scheduled_start: String(card.scheduled_start ?? ''),
				scheduled_end: String(card.scheduled_end ?? '')
			},
			questions: [],
			total_questions: 0
		};
	}

	function sessionAlreadyOpen(data: ExamPayload) {
		const state = (data.session?.status ?? '').toLowerCase();
		return Boolean(data.questions?.length || data.session?.started || data.session?.is_started || ['active', 'running', 'started', 'open', 'berjalan'].includes(state));
	}

	async function portalLogin() {
		loading = true;
		errorMessage = '';
		resetExamState();
		try {
			const fingerprint = deviceFingerprint.trim() || makeFingerprint();
			deviceFingerprint = fingerprint;
			if (demoMode) {
				payload = {
					student: { nama: 'Siswa Demo Portal', nis: 'DEMO-WEB', class_code: 'IX Demo', room_name: 'Ruang Simulasi', seat_no: 12 },
					session: { title: 'MODE DEMO Portal Ujian Peserta', subject: 'Informatika — Contoh lokal', status: 'waiting' },
					questions: demoQuestions,
					total_questions: demoQuestions.length,
					time_remaining_seconds: 45 * 60
				};
				portalStep = 'confirm';
				addTelemetry('Demo dibuka — tidak ada API produksi yang dipanggil');
				return;
			}

			if (showLegacyTokenLogin) {
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
				if (!response.ok) throw new Error(body.error ?? body.message ?? 'Login gagal');
				payload = normalizePortalPayload(body);
				portalStep = 'confirm';
				addTelemetry('Kredensial ujian diterima');
				return;
			}

			const response = await fetch('/api/exam/portal/card/verify', {
				method: 'POST',
				headers: { 'content-type': 'application/json' },
				body: JSON.stringify({ token: cardToken.trim(), pin: pin.trim() })
			});
			const body = await response.json();
			if (!response.ok) throw new Error(body.error ?? body.message ?? 'Verifikasi kartu gagal');
			payload = normalizeCardVerifyPayload(body);
			portalStep = 'confirm';
			addTelemetry('QR/kode kartu dan PIN diterima');
		} catch (error) {
			errorMessage = friendlyError(error);
		} finally {
			loading = false;
		}
	}

	async function confirmIdentity() {
		if (!payload) return;
		if (authenticatedByCard && !demoMode) {
			await startPortalExam();
			return;
		}
		portalStep = sessionAlreadyOpen(payload) ? 'exam' : 'waiting';
		addTelemetry(portalStep === 'exam' ? 'Identitas dikonfirmasi — ujian dibuka' : 'Identitas dikonfirmasi — menunggu pengawas membuka ujian');
	}

	async function startPortalExam() {
		if (!portalAuthToken || !activeParticipantId) {
			errorMessage = 'Data kartu belum lengkap. Silakan scan ulang atau panggil pengawas.';
			return;
		}
		loading = true;
		errorMessage = '';
		try {
			const response = await fetch(`/api/cbt-portal/participants/${encodeURIComponent(activeParticipantId)}/start`, {
				method: 'POST',
				headers: { 'content-type': 'application/json', authorization: `Bearer ${portalAuthToken}` },
				body: JSON.stringify({ device_fingerprint: deviceFingerprint.trim() || makeFingerprint(), browser_fingerprint: deviceFingerprint.trim() || makeFingerprint() })
			});
			const body = await response.json().catch(() => ({}));
			if (!response.ok) {
				const message = String((body as { error?: string; message?: string }).error ?? (body as { message?: string }).message ?? 'Ujian belum dibuka oleh pengawas.');
				if (response.status === 403 || response.status === 423 || message.toLowerCase().includes('belum')) {
					portalStep = 'waiting';
					addTelemetry('Identitas dikonfirmasi — menunggu pengawas membuka ujian');
					return;
				}
				throw new Error(message);
			}
			payload = normalizePortalPayload(body);
			portalStep = 'exam';
			addTelemetry('Identitas dikonfirmasi — ujian dibuka');
		} catch (error) {
			errorMessage = friendlyError(error);
		} finally {
			loading = false;
		}
	}

	function enterDemoExamFromWaiting() {
		if (!demoMode) return;
		if (payload?.session) payload.session.status = 'running';
		portalStep = 'exam';
		addTelemetry('Demo: pengawas membuka ujian');
	}

	async function saveAnswer(questionId: string, answer: string) {
		answers = { ...answers, [questionId]: answer };
		if (demoMode) {
			addTelemetry(`Demo menyimpan jawaban ${questionId} secara lokal`);
			return;
		}
		pendingAnswers = { ...pendingAnswers, [questionId]: answer };
		try {
			const response = await fetch(authenticatedByCard ? `/api/cbt-portal/participants/${encodeURIComponent(activeParticipantId)}/answer` : '/api/exam/answer', {
				method: 'POST',
				headers: authenticatedByCard
					? { 'content-type': 'application/json', authorization: `Bearer ${portalAuthToken}` }
					: {
						'content-type': 'application/json',
						'x-exam-token': examAccessToken.trim(),
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
				const response = await fetch(authenticatedByCard ? `/api/cbt-portal/participants/${encodeURIComponent(activeParticipantId)}/submit` : '/api/exam/submit', {
					method: 'POST',
					headers: authenticatedByCard
						? { 'content-type': 'application/json', authorization: `Bearer ${portalAuthToken}` }
						: {
							'content-type': 'application/json',
							'x-exam-token': examAccessToken.trim(),
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
		const token = queryCard;
		if (token && !cardToken) cardToken = token;
	});

	$effect(() => {
		if (!browser) return;
		deviceFingerprint = deviceFingerprint || makeFingerprint();
		const onBlur = () => addTelemetry(demoMode ? 'Demo: halaman ujian tidak aktif sesaat' : 'Halaman ujian tidak aktif sesaat');
		const onFocus = () => addTelemetry(demoMode ? 'Demo: halaman ujian aktif kembali' : 'Halaman ujian aktif kembali');
		const onOffline = () => addTelemetry('Koneksi terputus, jawaban disimpan sementara');
		const onOnline = () => addTelemetry('Koneksi kembali tersambung');
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
	<title>{demoMode ? 'MODE DEMO Portal Ujian' : 'Portal Ujian Peserta'} — MTsN 2 Kolaka Utara</title>
</svelte:head>

<main class="min-h-screen bg-slate-950 px-4 py-5 text-slate-100">
	<section class="mx-auto max-w-5xl space-y-4">
		<header class="rounded-2xl border border-emerald-300/20 bg-white/10 p-4">
			<p class="text-xs font-semibold uppercase tracking-[0.3em] text-amber-200">MTsN 2 Kolaka Utara</p>
			<h1 class="mt-2 text-2xl font-bold">{demoMode ? 'MODE DEMO Portal Ujian Peserta' : 'Portal Ujian Peserta'}</h1>
			<p class="mt-2 max-w-3xl text-sm text-slate-200/85">
				{demoMode
					? 'Mode DEMO meniru alur ujian nyata dengan soal contoh Informatika lokal. Jawaban tidak dikirim ke server dan tidak menjadi nilai.'
					: 'Scan QR pada Kartu Peserta Ujian, masukkan PIN, cek identitas, lalu tunggu pengawas membuka ujian.'}
			</p>
		</header>

		{#if errorMessage}
			<div class="rounded-xl border border-red-300/30 bg-red-500/20 p-3 text-sm text-red-50" role="alert">{errorMessage}</div>
		{/if}

		{#if submitted}
			<section class="rounded-2xl bg-white p-5 text-slate-950">
				<h2 class="text-xl font-bold">Selesai</h2>
				<p class="mt-2 text-sm text-slate-600">{demoMode ? 'Mode demo selesai lokal.' : 'Ujian telah dikumpulkan.'}</p>
				<button class="mt-4 rounded-lg border px-4 py-2 text-sm font-semibold" onclick={() => { payload = null; submitted = false; portalStep = 'login'; }}>Kembali</button>
			</section>
		{:else if payload && portalStep === 'confirm'}
			<section class="rounded-2xl bg-white p-5 text-slate-950">
				<p class="text-xs font-semibold uppercase tracking-[0.2em] text-emerald-700">Konfirmasi Identitas</p>
				<h2 class="mt-2 text-2xl font-bold">Apakah data ini benar?</h2>
				<div class="mt-4 grid gap-3 rounded-xl bg-emerald-50 p-4 text-sm sm:grid-cols-2">
					<p><span class="font-semibold">Nama:</span> {studentName}</p>
					<p><span class="font-semibold">NIS:</span> {student.nis ?? '—'}</p>
					<p><span class="font-semibold">Kelas:</span> {student.class_name ?? student.class_code ?? '—'}</p>
					<p><span class="font-semibold">Ruang/Meja:</span> {student.room_name ?? '—'} / {student.seat_no ?? '—'}</p>
					<p><span class="font-semibold">Ujian:</span> {session.title ?? session.subject ?? 'Sesi Ujian'}</p>
					<p><span class="font-semibold">Status:</span> {sessionAlreadyOpen(payload) ? 'Sudah dibuka' : 'Menunggu pengawas'}</p>
				</div>
				<p class="mt-3 text-sm text-slate-600">Jika nama/kelas/ruang tidak sesuai, jangan lanjut. Segera panggil pengawas.</p>
				<button class="mt-4 w-full rounded-xl bg-emerald-700 px-4 py-3 font-semibold text-white" onclick={confirmIdentity}>Ya, Masuk Ujian</button>
				<button class="mt-2 w-full rounded-xl border px-4 py-3 font-semibold" onclick={() => { payload = null; portalStep = 'login'; }}>Data Tidak Sesuai</button>
			</section>
		{:else if payload && portalStep === 'waiting'}
			<section class="rounded-2xl bg-white p-5 text-center text-slate-950">
				<div class="mx-auto grid size-16 place-items-center rounded-full bg-amber-100 text-3xl">⏳</div>
				<h2 class="mt-4 text-2xl font-bold">Ujian belum dimulai</h2>
				<p class="mx-auto mt-2 max-w-xl text-sm text-slate-600">Identitas sudah benar. Tetap di halaman ini dan tunggu pengawas menekan tombol <b>Mulai Ujian</b>. Jangan menutup browser.</p>
				<div class="mt-4 rounded-xl bg-slate-100 p-3 text-sm"><b>{studentName}</b> · {session.title ?? 'Sesi Ujian'} · {student.room_name ?? 'Ruang belum tercatat'}</div>
				{#if demoMode}
					<button class="mt-4 w-full rounded-xl bg-emerald-700 px-4 py-3 font-semibold text-white" onclick={enterDemoExamFromWaiting}>Demo: Simulasikan Pengawas Mulai Ujian</button>
				{:else if authenticatedByCard}
					<button class="mt-4 w-full rounded-xl bg-emerald-700 px-4 py-3 font-semibold text-white disabled:opacity-60" disabled={loading} onclick={startPortalExam}>{loading ? 'Mengecek...' : 'Cek Lagi: Pengawas Sudah Mulai'}</button>
				{/if}
			</section>
		{:else if payload && portalStep === 'exam'}
			<section class="rounded-2xl bg-white p-4 text-slate-950">
				<div class="flex flex-col gap-3 border-b pb-3 sm:flex-row sm:items-center sm:justify-between">
					<div>
						<p class="text-xs font-semibold uppercase text-emerald-700">{demoMode ? 'MODE DEMO — DATA CONTOH' : 'Portal Ujian Peserta'}</p>
						<h2 class="text-xl font-bold">{session.title ?? session.subject ?? 'Sesi Ujian'}</h2>
						<p class="text-sm text-slate-600">{studentName} · {answeredCount}/{questions.length} terjawab</p>
					</div>
					<button class="rounded-lg border px-4 py-2 text-sm font-semibold" onclick={() => { payload = null; portalStep = 'login'; }}>Keluar</button>
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
				<div class="mt-4 rounded-xl bg-emerald-50 p-3 text-sm text-emerald-900">{demoMode ? 'Mode DEMO: jawaban disimpan lokal di halaman ini saja. Simulasi/Gladi resmi harus memakai kegiatan dan paket server.' : 'Jawaban disimpan bertahap. Jika koneksi putus, tetap di halaman ini dan panggil pengawas.'}</div>
				<button class="mt-4 w-full rounded-xl bg-emerald-700 px-4 py-3 font-semibold text-white disabled:opacity-60" disabled={loading} onclick={submitExam}>Kumpulkan</button>
			</section>
		{:else}
			<section class="rounded-2xl bg-white p-5 text-slate-950">
				<h2 class="text-xl font-bold">{demoMode ? 'Mulai DEMO' : 'Masuk dengan Kartu Peserta Ujian'}</h2>
				<p class="mt-1 text-sm text-slate-600">{demoMode ? 'Tidak perlu QR/PIN. Klik mulai untuk memakai soal contoh Informatika lokal; tidak ada API atau database yang dimutasi.' : 'Scan QR pada kartu. Jika kamera perangkat tidak tersedia, ketik kode kartu dan PIN secara manual.'}</p>
				{#if !demoMode}
					{#if showLegacyTokenLogin}
						<div class="mt-4 grid gap-3 sm:grid-cols-2">
							<label class="space-y-1 text-sm font-medium">Token Ujian<input class="w-full rounded-lg border px-3 py-2" bind:value={examToken} autocomplete="off" /></label>
							<label class="space-y-1 text-sm font-medium">Token Ruang<input class="w-full rounded-lg border px-3 py-2" bind:value={roomToken} autocomplete="off" /></label>
						</div>
					{:else}
						<div class="mt-4 grid gap-3 sm:grid-cols-2">
							<label class="space-y-1 text-sm font-medium">Kode Kartu / QR Token<input class="w-full rounded-lg border px-3 py-2" bind:value={cardToken} autocomplete="off" placeholder="Terisi otomatis setelah scan QR" /></label>
							<label class="space-y-1 text-sm font-medium">PIN<input class="w-full rounded-lg border px-3 py-2 text-center text-xl tracking-[0.4em]" bind:value={pin} inputmode="numeric" autocomplete="one-time-code" maxlength="8" placeholder="••••" /></label>
						</div>
					{/if}
				{/if}
				<button class="mt-4 w-full rounded-xl bg-emerald-700 px-4 py-3 font-semibold text-white disabled:opacity-60" disabled={loading} onclick={portalLogin}>{loading ? 'Memproses...' : demoMode ? 'Mulai Mode DEMO' : 'Lanjutkan'}</button>
				<div class="mt-3 flex flex-col gap-2 text-center text-sm sm:flex-row sm:justify-center">
					{#if !demoMode}
						<button class="text-emerald-700 underline" type="button" onclick={() => (showLegacyTokenLogin = !showLegacyTokenLogin)}>{showLegacyTokenLogin ? 'Kembali ke QR + PIN' : 'Mode bantuan pengawas: token lama'}</button>
						<a class="text-emerald-700 underline" href="/ujian?demo=1">Buka Mode DEMO</a>
					{/if}
				</div>
			</section>
		{/if}

		<section class="rounded-2xl border border-white/10 bg-white/5 p-4 text-sm">
			<h2 class="font-semibold">Aktivitas perangkat</h2>
			{#if telemetry.length === 0}
				<p class="mt-2 text-slate-300">Belum ada aktivitas tercatat.</p>
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
