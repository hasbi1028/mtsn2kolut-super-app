<script lang="ts">
	import { browser } from '$app/environment';
	import { onDestroy, onMount } from 'svelte';

	type ApiEnvelope<T> = { data?: T; error?: string; message?: string };
	type ExamStudent = { nis?: string; nama?: string };
	type ExamSession = { id?: string; title?: string; scheduled_start?: string; scheduled_end?: string; duration_minutes?: number };
	type ExamRoom = { room_name?: string };
	type ExamQuestion = {
		id: string;
		code?: string;
		question_text?: string;
		question_type?: string;
		options?: unknown;
		stem_html?: string;
		stimulus_html?: string;
		stem_media_url?: string;
		stimulus_media_url?: string;
		stem_audio_url?: string;
		stimulus_audio_url?: string;
		option_a?: string;
		option_b?: string;
		option_c?: string;
		option_d?: string;
		option_e?: string;
	};
	type ExamLoginPayload = {
		participant_id: string;
		student: ExamStudent;
		session: ExamSession;
		room?: ExamRoom;
		questions: ExamQuestion[];
		answered_count: number;
		total_questions: number;
		time_remaining_seconds: number;
		anti_cheat?: { locked?: boolean; risk_level?: string; risk_score?: number; violation_count?: number; locked_reason?: string };
	};
	type ExamStatusPayload = {
		answered_count: number;
		total_questions: number;
		time_remaining_seconds: number;
		is_submitted: boolean;
		submitted_at?: string;
		anti_cheat?: ExamLoginPayload['anti_cheat'];
	};
	type PendingAnswer = { question_id: string; answer: string; updated_at: string; attempts: number };
	type OptionItem = { label: string; text: string };
	type NoticeTone = 'info' | 'success' | 'warning' | 'error';
	type Notice = { tone: NoticeTone; text: string };

	const STORAGE_PREFIX = 'mtsn2kolut-cbt-web-fallback:';
	const CLIENT_TYPE = 'web_fallback';
	const HEARTBEAT_MS = 30000;
	const FLUSH_MS = 12000;

	let token = $state('');
	let roomToken = $state('');
	let fingerprint = $state('');
	let payload = $state<ExamLoginPayload | null>(null);
	let activeIndex = $state(0);
	let answers = $state<Record<string, string>>({});
	let pendingAnswers = $state<Record<string, PendingAnswer>>({});
	let busyLogin = $state(false);
	let busySubmit = $state(false);
	let notice = $state<Notice | null>(null);
	let lastSyncAt = $state('');
	let offline = $state(false);
	let submitted = $state(false);
	let remainingSeconds = $state(0);
	let heartbeatTimer: ReturnType<typeof setInterval> | null = null;
	let flushTimer: ReturnType<typeof setInterval> | null = null;
	let countdownTimer: ReturnType<typeof setInterval> | null = null;
	let lastVisibilityState = '';
	let hasFullscreenWarning = $state(false);

	let questions = $derived(payload?.questions ?? []);
	let activeQuestion = $derived(questions[activeIndex] ?? null);
	let pendingCount = $derived(Object.keys(pendingAnswers).length);
	let answeredCount = $derived(Object.values(answers).filter((value) => value.trim() !== '').length);
	let durationLabel = $derived(formatDuration(remainingSeconds));

	onMount(() => {
		fingerprint = getOrCreateFingerprint();
		restoreSession();
		window.addEventListener('online', handleOnline);
		window.addEventListener('offline', handleOffline);
		document.addEventListener('visibilitychange', handleVisibilityChange);
		window.addEventListener('blur', handleBlur);
		window.addEventListener('focus', handleFocus);
		document.addEventListener('fullscreenchange', handleFullscreenChange);
		countdownTimer = setInterval(() => {
			if (remainingSeconds > 0) remainingSeconds -= 1;
		}, 1000);
	});

	onDestroy(() => {
		stopIntervals();
		if (!browser) return;
		window.removeEventListener('online', handleOnline);
		window.removeEventListener('offline', handleOffline);
		document.removeEventListener('visibilitychange', handleVisibilityChange);
		window.removeEventListener('blur', handleBlur);
		window.removeEventListener('focus', handleFocus);
		document.removeEventListener('fullscreenchange', handleFullscreenChange);
	});

	async function login() {
		const cleanToken = token.trim();
		const cleanRoomToken = roomToken.trim();
		if (!cleanToken || !cleanRoomToken) {
			notice = { tone: 'warning', text: 'Token ujian dan token ruang wajib diisi dari pengawas.' };
			return;
		}
		busyLogin = true;
		notice = null;
		try {
			const data = await examRequest<ExamLoginPayload>('/api/exam/login', {
				method: 'POST',
				body: {
					token: cleanToken,
					room_token: cleanRoomToken,
					device_fingerprint: fingerprint,
					browser_fingerprint: fingerprint,
					client_type: CLIENT_TYPE
				}
			});
			payload = data;
			token = cleanToken;
			roomToken = cleanRoomToken;
			remainingSeconds = Math.max(0, Number(data.time_remaining_seconds ?? 0));
			submitted = false;
			answers = { ...answers };
			restorePendingForToken(cleanToken);
			persistSession();
			startIntervals();
			await recordEvent('web_fallback_used', { reason: 'browser_darurat_login' });
			await flushPendingAnswers();
			notice = { tone: 'success', text: 'Masuk mode Browser Darurat. Pengawasan ruang wajib tetap berjalan.' };
		} catch (error) {
			notice = { tone: 'error', text: errorMessage(error) };
		} finally {
			busyLogin = false;
		}
	}

	async function saveAnswer(question: ExamQuestion, answer: string) {
		if (!payload || submitted) return;
		answers = { ...answers, [question.id]: answer };
		const pending: PendingAnswer = { question_id: question.id, answer, updated_at: new Date().toISOString(), attempts: pendingAnswers[question.id]?.attempts ?? 0 };
		pendingAnswers = { ...pendingAnswers, [question.id]: pending };
		persistSession();
		void sendAnswer(question.id, answer);
	}

	async function sendAnswer(questionId: string, answer: string) {
		try {
			await examRequest<{ status: string }>('/api/exam/answer', {
				method: 'POST',
				examToken: token,
				body: { question_id: questionId, answer }
			});
			const next = { ...pendingAnswers };
			delete next[questionId];
			pendingAnswers = next;
			lastSyncAt = new Date().toISOString();
			persistSession();
			await recordEvent('web_pending_answer_flushed', { question_id: questionId, pending_answer_count: pendingCount, last_synced_at: lastSyncAt, sync_state: 'synced' });
		} catch (error) {
			const current = pendingAnswers[questionId];
			if (current) {
				pendingAnswers = { ...pendingAnswers, [questionId]: { ...current, attempts: current.attempts + 1 } };
			}
			offline = true;
			persistSession();
			await recordEvent('web_pending_answer_saved', { question_id: questionId, pending_answer_count: pendingCount, last_local_save_at: new Date().toISOString(), sync_state: 'pending' });
			notice = { tone: 'warning', text: `Jawaban tersimpan sementara di sesi browser. Kirim ulang otomatis saat koneksi pulih. (${errorMessage(error)})` };
		}
	}

	async function flushPendingAnswers() {
		if (!payload || !token.trim()) return;
		const items = Object.values(pendingAnswers).sort((a, b) => a.updated_at.localeCompare(b.updated_at));
		for (const item of items) {
			await sendAnswer(item.question_id, item.answer);
		}
		if (items.length > 0 && Object.keys(pendingAnswers).length === 0) {
			offline = false;
			notice = { tone: 'success', text: 'Semua jawaban tertunda sudah tersinkron ke server.' };
		}
	}

	async function submitExam() {
		if (!payload || submitted) return;
		if (pendingCount > 0) {
			await flushPendingAnswers();
		}
		if (Object.keys(pendingAnswers).length > 0) {
			notice = { tone: 'warning', text: 'Masih ada jawaban lokal yang belum terkirim. Jangan submit sebelum sinkron selesai.' };
			return;
		}
		if (!confirm('Kirim akhir ujian sekarang? Setelah dikirim, jawaban tidak dapat diubah.')) return;
		busySubmit = true;
		try {
			await examRequest<{ status: string }>('/api/exam/submit', { method: 'POST', examToken: token, body: {} });
			submitted = true;
			clearSessionStorage();
			stopIntervals();
			notice = { tone: 'success', text: 'Ujian berhasil dikirim. Laporkan ke pengawas sebelum menutup browser.' };
		} catch (error) {
			notice = { tone: 'error', text: errorMessage(error) };
		} finally {
			busySubmit = false;
		}
	}

	async function refreshStatus() {
		if (!payload || !token.trim()) return;
		try {
			const status = await examRequest<ExamStatusPayload>('/api/exam/status', { method: 'GET', examToken: token });
			remainingSeconds = Math.max(0, Number(status.time_remaining_seconds ?? remainingSeconds));
			submitted = Boolean(status.is_submitted);
			if (submitted) {
				clearSessionStorage();
				stopIntervals();
			}
			offline = false;
		} catch {
			offline = true;
		}
	}

	async function sendHeartbeat() {
		if (!payload || !token.trim()) return;
		try {
			await examRequest<{ status: string }>('/api/exam/heartbeat', { method: 'POST', examToken: token, body: {} });
			offline = false;
		} catch {
			offline = true;
			await recordEvent('web_connection_degraded', { pending_answer_count: pendingCount, sync_state: pendingCount > 0 ? 'pending' : 'unknown' });
		}
	}

	async function recordEvent(eventType: string, data: Record<string, unknown> = {}) {
		if (!token.trim()) return;
		try {
			await examRequest<{ status: string }>('/api/exam/event', {
				method: 'POST',
				examToken: token,
				body: {
					event_type: eventType,
					data: {
						...data,
						client_type: CLIENT_TYPE,
						pending_answer_count: pendingCount,
						original_event_at: new Date().toISOString()
					}
				}
			});
		} catch {
			// Telemetry must not block the student in emergency browser mode.
		}
	}

	async function enterFullscreen() {
		try {
			await document.documentElement.requestFullscreen();
			hasFullscreenWarning = false;
			await recordEvent('web_fullscreen_restored');
		} catch {
			notice = { tone: 'warning', text: 'Browser tidak mengizinkan fullscreen. Pengawas wajib mencatat kondisi perangkat.' };
		}
	}

	function startIntervals() {
		stopIntervals();
		heartbeatTimer = setInterval(() => void sendHeartbeat(), HEARTBEAT_MS);
		flushTimer = setInterval(() => void flushPendingAnswers(), FLUSH_MS);
		void sendHeartbeat();
	}

	function stopIntervals() {
		if (heartbeatTimer) clearInterval(heartbeatTimer);
		if (flushTimer) clearInterval(flushTimer);
		heartbeatTimer = null;
		flushTimer = null;
	}

	function handleOnline() {
		offline = false;
		notice = { tone: 'info', text: 'Koneksi kembali. Mengirim jawaban tertunda...' };
		void recordEvent('web_connection_restored', { pending_answer_count: pendingCount });
		void flushPendingAnswers();
	}

	function handleOffline() {
		offline = true;
		notice = { tone: 'warning', text: 'Browser offline. Jawaban akan disimpan sementara di sesi browser ini.' };
		void recordEvent('web_connection_degraded', { pending_answer_count: pendingCount });
	}

	function handleVisibilityChange() {
		const state = document.visibilityState;
		if (state === lastVisibilityState) return;
		lastVisibilityState = state;
		if (state === 'hidden') void recordEvent('web_visibility_hidden', { reason: 'document_hidden' });
		if (state === 'visible') void recordEvent('web_visibility_visible', { reason: 'document_visible' });
	}

	function handleBlur() {
		void recordEvent('web_focus_lost', { reason: 'window_blur' });
	}

	function handleFocus() {
		void recordEvent('web_focus_restored', { reason: 'window_focus' });
	}

	function handleFullscreenChange() {
		if (!payload) return;
		if (!document.fullscreenElement) {
			hasFullscreenWarning = true;
			void recordEvent('web_fullscreen_exit', { reason: 'fullscreen_exit' });
		} else {
			hasFullscreenWarning = false;
			void recordEvent('web_fullscreen_restored', { reason: 'fullscreen_restored' });
		}
	}

	async function examRequest<T>(path: string, init: { method: string; examToken?: string; body?: unknown }): Promise<T> {
		const headers: Record<string, string> = { Accept: 'application/json' };
		if (init.body !== undefined) headers['Content-Type'] = 'application/json';
		if (init.examToken) headers['X-Exam-Token'] = init.examToken;
		if (fingerprint) headers['X-Device-Fingerprint'] = fingerprint;
		const response = await fetch(path, {
			method: init.method,
			headers,
			body: init.body === undefined ? undefined : JSON.stringify(init.body)
		});
		const json = (await response.json().catch(() => null)) as ApiEnvelope<T> | T | null;
		if (!response.ok) {
			if (json && typeof json === 'object' && ('error' in json || 'message' in json)) {
				throw new Error((json as ApiEnvelope<T>).error ?? (json as ApiEnvelope<T>).message ?? `HTTP ${response.status}`);
			}
			throw new Error(response.status >= 500 ? 'Server ujian belum dapat dihubungi.' : `HTTP ${response.status}`);
		}
		if (json && typeof json === 'object' && 'data' in json) return (json as ApiEnvelope<T>).data as T;
		return json as T;
	}

	function updateAnswerFromInput(question: ExamQuestion, value: string) {
		void saveAnswer(question, value);
	}

	function optionItems(question: ExamQuestion): OptionItem[] {
		const parsed = parseOptionItems(question.options);
		if (parsed.length > 0) return parsed;
		return [question.option_a, question.option_b, question.option_c, question.option_d, question.option_e]
			.map((text, index) => ({ label: String.fromCharCode(65 + index), text: String(text ?? '').trim() }))
			.filter((item) => item.text !== '');
	}

	function parseOptionItems(raw: unknown): OptionItem[] {
		if (!Array.isArray(raw)) return [];
		return raw.map((item, index) => {
			if (item && typeof item === 'object') {
				const record = item as Record<string, unknown>;
				return {
					label: String(record.label ?? String.fromCharCode(65 + index)).trim(),
					text: String(record.text ?? record.value ?? record.html ?? '').trim()
				};
			}
			return { label: String.fromCharCode(65 + index), text: String(item ?? '').trim() };
		}).filter((item) => item.text !== '');
	}

	function isObjective(question: ExamQuestion): boolean {
		return ['multiple_choice', 'true_false', 'agree_disagree'].includes(String(question.question_type ?? ''));
	}

	function isMultipleAnswer(question: ExamQuestion): boolean {
		return String(question.question_type ?? '') === 'multiple_answer';
	}

	function toggleMultipleAnswer(question: ExamQuestion, label: string, checked: boolean) {
		const current = new Set((answers[question.id] ?? '').split(',').map((item) => item.trim()).filter(Boolean));
		if (checked) current.add(label);
		else current.delete(label);
		void saveAnswer(question, Array.from(current).sort().join(','));
	}

	function choiceChecked(question: ExamQuestion, label: string): boolean {
		return (answers[question.id] ?? '').split(',').map((item) => item.trim()).includes(label);
	}

	function questionText(question: ExamQuestion): string {
		return question.stem_html || question.question_text || '';
	}

	function storageKey(cleanToken = token.trim()) {
		return `${STORAGE_PREFIX}${cleanToken || 'anonymous'}`;
	}

	function persistSession() {
		if (typeof sessionStorage === 'undefined' || !token.trim()) return;
		sessionStorage.setItem(storageKey(), JSON.stringify({ token, roomToken, fingerprint, payload, activeIndex, answers, pendingAnswers, submitted, remainingSeconds }));
	}

	function restoreSession() {
		if (typeof sessionStorage === 'undefined') return;
		const lastToken = sessionStorage.getItem(`${STORAGE_PREFIX}last-token`) ?? '';
		if (!lastToken) return;
		token = lastToken;
		restorePendingForToken(lastToken);
		if (payload && !submitted) startIntervals();
	}

	function restorePendingForToken(cleanToken: string) {
		if (typeof sessionStorage === 'undefined') return;
		const raw = sessionStorage.getItem(storageKey(cleanToken));
		if (!raw) {
			pendingAnswers = {};
			return;
		}
		try {
			const parsed = JSON.parse(raw) as Partial<{ token: string; roomToken: string; fingerprint: string; payload: ExamLoginPayload; activeIndex: number; answers: Record<string, string>; pendingAnswers: Record<string, PendingAnswer>; submitted: boolean; remainingSeconds: number }>;
			token = parsed.token ?? cleanToken;
			roomToken = parsed.roomToken ?? roomToken;
			fingerprint = parsed.fingerprint ?? fingerprint;
			payload = parsed.payload ?? payload;
			activeIndex = Math.max(0, Number(parsed.activeIndex ?? 0));
			answers = parsed.answers ?? {};
			pendingAnswers = parsed.pendingAnswers ?? {};
			submitted = Boolean(parsed.submitted);
			remainingSeconds = Math.max(0, Number(parsed.remainingSeconds ?? remainingSeconds));
		} catch {
			pendingAnswers = {};
		}
	}

	function clearSessionStorage() {
		if (typeof sessionStorage === 'undefined') return;
		sessionStorage.removeItem(storageKey());
		sessionStorage.removeItem(`${STORAGE_PREFIX}last-token`);
	}

	function getOrCreateFingerprint(): string {
		if (typeof sessionStorage === 'undefined') return `web-${crypto.randomUUID()}`;
		const key = `${STORAGE_PREFIX}fingerprint`;
		const existing = sessionStorage.getItem(key);
		if (existing) return existing;
		const value = `web-${crypto.randomUUID()}`;
		sessionStorage.setItem(key, value);
		return value;
	}

	$effect(() => {
		if (typeof sessionStorage !== 'undefined' && token.trim()) {
			sessionStorage.setItem(`${STORAGE_PREFIX}last-token`, token.trim());
		}
	});

	function formatDuration(total: number): string {
		const seconds = Math.max(0, Math.floor(total));
		const h = Math.floor(seconds / 3600);
		const m = Math.floor((seconds % 3600) / 60);
		const s = seconds % 60;
		return h > 0 ? `${h}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}` : `${m}:${String(s).padStart(2, '0')}`;
	}

	function formatDate(value?: string): string {
		if (!value) return '—';
		const date = new Date(value);
		if (Number.isNaN(date.getTime())) return value;
		return new Intl.DateTimeFormat('id-ID', { dateStyle: 'medium', timeStyle: 'short' }).format(date);
	}

	function errorMessage(error: unknown): string {
		return error instanceof Error ? error.message : 'Terjadi kendala pada browser ujian.';
	}

	function noticeClass(tone: NoticeTone): string {
		if (tone === 'success') return 'border-emerald-500/40 bg-emerald-50 text-emerald-900';
		if (tone === 'warning') return 'border-amber-500/40 bg-amber-50 text-amber-950';
		if (tone === 'error') return 'border-red-500/40 bg-red-50 text-red-900';
		return 'border-sky-500/40 bg-sky-50 text-sky-900';
	}
</script>

<svelte:head>
	<title>Ujian Browser Darurat</title>
	<meta name="robots" content="noindex,nofollow" />
</svelte:head>

<div class="min-h-screen bg-slate-950 text-slate-100">
	<header class="border-b border-amber-400/30 bg-slate-900/95 px-4 py-3 shadow-lg">
		<div class="mx-auto flex max-w-6xl flex-col gap-2 md:flex-row md:items-center md:justify-between">
			<div>
				<p class="text-xs font-bold uppercase tracking-[0.24em] text-amber-300">Mode Darurat / Browser — pengawasan wajib</p>
				<h1 class="text-xl font-bold">CBT MTsN 2 Kolaka Utara</h1>
			</div>
			<div class="flex flex-wrap items-center gap-2 text-xs font-semibold">
				<span class={`rounded-full border px-3 py-1 ${offline ? 'border-amber-300 text-amber-200' : 'border-emerald-300 text-emerald-200'}`}>{offline ? 'Koneksi waspada' : 'Online'}</span>
				<span class="rounded-full border border-slate-600 px-3 py-1">Sisa waktu {durationLabel}</span>
				{#if pendingCount > 0}<span class="rounded-full border border-amber-300 px-3 py-1 text-amber-200">{pendingCount} jawaban lokal</span>{/if}
			</div>
		</div>
	</header>

	<main class="mx-auto max-w-6xl space-y-4 px-4 py-5">
		{#if notice}
			<div class={`rounded-xl border p-3 text-sm ${noticeClass(notice.tone)}`}>{notice.text}</div>
		{/if}

		{#if !payload}
			<section class="grid gap-4 lg:grid-cols-[minmax(0,1fr)_360px]">
				<div class="rounded-2xl border border-amber-400/30 bg-slate-900 p-5 shadow-xl">
					<h2 class="text-2xl font-bold">Masuk Ujian Browser Darurat</h2>
					<p class="mt-2 text-sm leading-6 text-slate-300">Jalur ini hanya cadangan ketika aplikasi CBT Flutter/desktop bermasalah dan sudah diizinkan operator untuk ruang ini. Anti-cheat browser bersifat best-effort; pengawas wajib mencatat penggunaan mode ini.</p>
					<div class="mt-5 grid gap-3">
						<label class="grid gap-1 text-sm font-semibold">Token ujian peserta
							<input class="rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-base text-white outline-none focus:border-amber-300" bind:value={token} autocomplete="off" spellcheck="false" />
						</label>
						<label class="grid gap-1 text-sm font-semibold">Token ruang dari pengawas
							<input class="rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-base text-white outline-none focus:border-amber-300" bind:value={roomToken} autocomplete="off" spellcheck="false" />
						</label>
						<button type="button" class="rounded-lg bg-amber-400 px-4 py-2 font-bold text-slate-950 disabled:cursor-not-allowed disabled:opacity-60" disabled={busyLogin} onclick={login}>{busyLogin ? 'Memeriksa...' : 'Masuk Mode Browser Darurat'}</button>
					</div>
				</div>
				<aside class="rounded-2xl border border-slate-700 bg-slate-900 p-5 text-sm leading-6 text-slate-300">
					<h3 class="font-bold text-slate-100">SOP singkat</h3>
					<ol class="mt-2 list-decimal space-y-2 pl-5">
						<li>Pastikan pengawas/operator sudah mengaktifkan Browser Darurat untuk ruang ini.</li>
						<li>Gunakan perangkat di meja peserta, jangan berpindah tab/aplikasi.</li>
						<li>Jika koneksi putus, jangan tutup browser; tunggu sinkron otomatis.</li>
						<li>Setelah submit, lapor pengawas sebelum menutup halaman.</li>
					</ol>
				</aside>
			</section>
		{:else}
			<section class="grid gap-4 lg:grid-cols-[260px_minmax(0,1fr)]">
				<aside class="space-y-3 rounded-2xl border border-slate-700 bg-slate-900 p-3">
					<div class="rounded-xl bg-slate-950 p-3 text-sm">
						<p class="font-bold">{payload.student?.nama ?? 'Peserta'}</p>
						<p class="text-slate-400">{payload.student?.nis ?? '—'} · {payload.room?.room_name ?? 'Ruang ujian'}</p>
						<p class="mt-2 text-xs text-slate-400">{payload.session?.title ?? 'Sesi Ujian'}</p>
						<p class="text-xs text-slate-500">{formatDate(payload.session?.scheduled_start)} - {formatDate(payload.session?.scheduled_end)}</p>
					</div>
					<button type="button" class="w-full rounded-lg border border-amber-400/60 px-3 py-2 text-sm font-semibold text-amber-200" onclick={enterFullscreen}>Masuk Fullscreen</button>
					<button type="button" class="w-full rounded-lg border border-slate-600 px-3 py-2 text-sm font-semibold" onclick={() => void refreshStatus()}>Refresh Status</button>
					<div class="grid grid-cols-5 gap-1">
						{#each questions as question, index (question.id)}
							<button type="button" class={`rounded-md border px-2 py-1 text-xs font-bold ${activeIndex === index ? 'border-amber-300 bg-amber-300 text-slate-950' : answers[question.id]?.trim() ? 'border-emerald-400 text-emerald-200' : 'border-slate-700 text-slate-300'}`} onclick={() => { activeIndex = index; persistSession(); }}>{index + 1}</button>
						{/each}
					</div>
					<div class="rounded-xl bg-slate-950 p-3 text-xs text-slate-400">Terjawab {answeredCount}/{questions.length}. Pending lokal {pendingCount}. {lastSyncAt ? `Sinkron terakhir ${formatDate(lastSyncAt)}.` : ''}</div>
				</aside>

				<div class="space-y-4">
					{#if hasFullscreenWarning}
						<div class="rounded-xl border border-amber-400/40 bg-amber-300/10 p-3 text-sm text-amber-100">Fullscreen keluar. Tetap di halaman ujian dan panggil pengawas jika tidak sengaja.</div>
					{/if}
					{#if activeQuestion}
						<article class="rounded-2xl border border-slate-700 bg-slate-900 p-5 shadow-xl">
							<div class="mb-4 flex flex-col gap-2 md:flex-row md:items-start md:justify-between">
								<div>
									<p class="text-xs font-bold uppercase tracking-wide text-amber-300">Soal {activeIndex + 1} dari {questions.length}</p>
									<h2 class="text-lg font-bold">{activeQuestion.code || 'Soal'}</h2>
								</div>
								<span class="rounded-full border border-slate-600 px-3 py-1 text-xs">{activeQuestion.question_type ?? 'soal'}</span>
							</div>

							{#if activeQuestion.stimulus_html}<div class="prose prose-invert max-w-none rounded-xl bg-slate-950 p-3 text-sm">{@html activeQuestion.stimulus_html}</div>{/if}
							{#if activeQuestion.stimulus_media_url}<img class="mt-3 max-h-80 rounded-xl border border-slate-700 object-contain" src={activeQuestion.stimulus_media_url} alt="Stimulus soal" />{/if}
							{#if activeQuestion.stimulus_audio_url}<audio class="mt-3 w-full" controls src={activeQuestion.stimulus_audio_url}></audio>{/if}

							<div class="prose prose-invert mt-4 max-w-none rounded-xl bg-slate-950 p-4">{@html questionText(activeQuestion)}</div>
							{#if activeQuestion.stem_media_url}<img class="mt-3 max-h-80 rounded-xl border border-slate-700 object-contain" src={activeQuestion.stem_media_url} alt="Media soal" />{/if}
							{#if activeQuestion.stem_audio_url}<audio class="mt-3 w-full" controls src={activeQuestion.stem_audio_url}></audio>{/if}

							<div class="mt-5 space-y-3">
								{#if isMultipleAnswer(activeQuestion)}
									{#each optionItems(activeQuestion) as option (option.label)}
										<label class="flex gap-3 rounded-xl border border-slate-700 bg-slate-950 p-3 text-sm">
											<input type="checkbox" checked={choiceChecked(activeQuestion, option.label)} onchange={(event) => toggleMultipleAnswer(activeQuestion, option.label, event.currentTarget.checked)} />
											<span><b>{option.label}.</b> {@html option.text}</span>
										</label>
									{/each}
								{:else if isObjective(activeQuestion)}
									{#each optionItems(activeQuestion) as option (option.label)}
										<label class="flex gap-3 rounded-xl border border-slate-700 bg-slate-950 p-3 text-sm">
											<input type="radio" name={activeQuestion.id} value={option.label} checked={answers[activeQuestion.id] === option.label} onchange={() => updateAnswerFromInput(activeQuestion, option.label)} />
											<span><b>{option.label}.</b> {@html option.text}</span>
										</label>
									{/each}
								{:else}
									<textarea class="min-h-40 w-full rounded-xl border border-slate-700 bg-slate-950 p-3 text-base text-white outline-none focus:border-amber-300" value={answers[activeQuestion.id] ?? ''} oninput={(event) => updateAnswerFromInput(activeQuestion, event.currentTarget.value)} placeholder="Tulis jawaban di sini"></textarea>
								{/if}
							</div>
						</article>

						<div class="flex flex-col gap-2 rounded-2xl border border-slate-700 bg-slate-900 p-3 md:flex-row md:items-center md:justify-between">
							<div class="flex gap-2">
								<button type="button" class="rounded-lg border border-slate-600 px-4 py-2 font-semibold disabled:opacity-40" disabled={activeIndex <= 0} onclick={() => { activeIndex -= 1; persistSession(); }}>Sebelumnya</button>
								<button type="button" class="rounded-lg border border-slate-600 px-4 py-2 font-semibold disabled:opacity-40" disabled={activeIndex >= questions.length - 1} onclick={() => { activeIndex += 1; persistSession(); }}>Berikutnya</button>
							</div>
							<button type="button" class="rounded-lg bg-emerald-400 px-4 py-2 font-bold text-slate-950 disabled:cursor-not-allowed disabled:opacity-60" disabled={busySubmit || submitted} onclick={submitExam}>{busySubmit ? 'Mengirim...' : submitted ? 'Sudah dikirim' : 'Kirim Akhir Ujian'}</button>
						</div>
					{/if}
				</div>
			</section>
		{/if}
	</main>
</div>
