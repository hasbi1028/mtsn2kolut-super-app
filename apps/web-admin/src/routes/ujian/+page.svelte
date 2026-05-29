<script lang="ts">
	import { browser } from '$app/environment';
	import { page } from '$app/stores';
	import { onDestroy } from 'svelte';

	type Question = {
		id: string;
		type: 'multiple_choice' | 'essay' | 'short_answer';
		text: string;
		options?: Array<{ label: string; text: string }>;
	};
	type StudentIdentity = { nama?: string; name?: string; nis?: string; class_code?: string; class_name?: string; room_name?: string; seat_no?: number | string | null; participant_id?: string };
	type SessionIdentity = { id?: string; title?: string; subject?: string; scheduled_start?: string; scheduled_end?: string; status?: string; started?: boolean; is_started?: boolean; waiting?: boolean };
	type ExamPayload = {
		participant_id?: string;
		student?: StudentIdentity;
		participant?: StudentIdentity;
		session?: SessionIdentity;
		room?: { room_name?: string };
		access_token?: string;
		exam_token?: string;
		token?: string;
		room_token?: string;
		questions?: Question[];
		total_questions?: number;
		time_remaining_seconds?: number;
	};
	type PortalStep = 'login' | 'confirm' | 'waiting' | 'exam';

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
	let activeQuestionIndex = $state(0);
	let doubtfulQuestions = $state(new Set<string>());
	let watermarkTime = $state('');
	let participantPollInterval: ReturnType<typeof setInterval> | undefined;
	let countdownInterval: ReturnType<typeof setInterval> | undefined;
	let timeRemainingSeconds = $state<number | null>(null);
	let waitingAutoCheckBusy = false;
	let syncingPendingAnswers = $state(false);
	let seenCommandIds = $state(new Set<string>());

	let queryCard = $derived($page.url.searchParams.get('card') ?? $page.url.searchParams.get('token') ?? '');
	let questions = $derived(payload?.questions ?? []);
	let currentQuestion = $derived(questions[Math.min(activeQuestionIndex, Math.max(questions.length - 1, 0))]);
	let answeredCount = $derived(Object.values(answers).filter((answer) => answer.trim().length > 0).length);
	let student = $derived(payload?.student ?? payload?.participant ?? {});
	let session = $derived(payload?.session ?? {});
	let studentName = $derived(student.nama ?? student.name ?? 'Peserta');
	let examAccessToken = $derived(payload?.access_token ?? payload?.exam_token ?? payload?.token ?? examToken);
	let classLabel = $derived(student.class_name ?? student.class_code ?? 'Kelas belum tercatat');
	let roomLabel = $derived(student.room_name ?? payload?.room?.room_name ?? 'Ruang belum tercatat');
	let seatLabel = $derived(student.seat_no ? `Kursi ${student.seat_no}` : 'Kursi belum tercatat');
	let participantLabel = $derived(payload?.participant_id ?? student.participant_id ?? activeParticipantId);
	let participantWatermark = $derived(participantLabel ? `Peserta ${participantLabel}` : 'Peserta belum tercatat');
	let sessionTitleLabel = $derived(session.title ?? session.subject ?? 'Sesi ujian');
	let sessionIdLabel = $derived(session.id ? `Sesi ${session.id}` : '');
	let pendingAnswerCount = $derived(Object.keys(pendingAnswers).length);
	let timeRemainingLabel = $derived(formatTimeRemaining(timeRemainingSeconds));
	let timeRemainingClass = $derived(timeRemainingSeconds === null ? 'border-slate-200 bg-slate-50 text-slate-700' : timeRemainingSeconds <= 5 * 60 ? 'border-red-200 bg-red-50 text-red-800' : timeRemainingSeconds <= 10 * 60 ? 'border-amber-200 bg-amber-50 text-amber-900' : 'border-emerald-200 bg-emerald-50 text-emerald-900');
	let watermarkLine = $derived(
		[
			studentName,
			classLabel,
			roomLabel,
			seatLabel,
			sessionTitleLabel,
			participantWatermark,
			sessionIdLabel,
			watermarkTime
		].filter(Boolean).join(' · ')
	);

	function friendlyError(error: unknown) {
		const message = error instanceof Error ? error.message : String(error || '');
		const lower = message.toLowerCase();
		if (lower.includes('terikat perangkat') || lower.includes('device mismatch') || lower.includes('another device')) return 'Akun ujian sudah terikat perangkat lain. Minta pengawas/admin membuka ulang akses peserta.';
		if (lower.includes('terkunci') || lower.includes('locked')) return 'Akses ujian terkunci oleh kebijakan pengawasan. Silakan panggil pengawas.';
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

	function formatTimeRemaining(seconds: number | null) {
		if (seconds === null || !Number.isFinite(seconds)) return 'Waktu menunggu';
		const safe = Math.max(0, Math.floor(seconds));
		const h = Math.floor(safe / 3600);
		const m = Math.floor((safe % 3600) / 60);
		const sec = safe % 60;
		if (h > 0) return `${h}:${String(m).padStart(2, '0')}:${String(sec).padStart(2, '0')}`;
		return `${m}:${String(sec).padStart(2, '0')}`;
	}

	function setExamPayload(nextPayload: ExamPayload | null) {
		payload = nextPayload;
		timeRemainingSeconds = typeof nextPayload?.time_remaining_seconds === 'number' ? nextPayload.time_remaining_seconds : null;
	}

	function startCountdown() {
		if (!browser) return;
		stopCountdown();
		if (timeRemainingSeconds === null) return;
		countdownInterval = setInterval(() => {
			if (portalStep !== 'exam' || submitted) return;
			timeRemainingSeconds = Math.max(0, (timeRemainingSeconds ?? 0) - 1);
		}, 1000);
	}

	function stopCountdown() {
		if (countdownInterval) clearInterval(countdownInterval);
		countdownInterval = undefined;
	}

	function resetExamState() {
		submitted = false;
		answers = {};
		pendingAnswers = {};
		activeQuestionIndex = 0;
		doubtfulQuestions = new Set();
		activeParticipantId = '';
		portalAuthToken = '';
		authenticatedByCard = false;
		stopParticipantRuntimePolling();
		stopCountdown();
		timeRemainingSeconds = null;
	}

	function normalizePortalPayload(body: unknown): ExamPayload {
		const wrapped = body as { data?: ExamPayload; payload?: ExamPayload };
		const source = wrapped?.data ?? wrapped?.payload ?? (body as ExamPayload);
		return {
			...source,
			participant_id: source.participant_id,
			room: source.room,
			student: {
				...(source.student ?? source.participant ?? {}),
				room_name: (source.student ?? source.participant ?? {}).room_name ?? source.room?.room_name,
				participant_id: source.participant_id
			},
			questions: (source.questions ?? []).map((question) => ({
				...question,
				text: question.text ?? (question as Question & { question_text?: string }).question_text ?? '',
				type: question.type ?? (question as Question & { question_type?: Question['type'] }).question_type ?? 'multiple_choice'
			}))
		};
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
				setExamPayload(normalizePortalPayload(body));
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
			setExamPayload(normalizeCardVerifyPayload(body));
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
		if (authenticatedByCard) {
			await startPortalExam();
			return;
		}
		portalStep = sessionAlreadyOpen(payload) ? 'exam' : 'waiting';
		if (portalStep === 'exam') startCountdown();
		addTelemetry(portalStep === 'exam' ? 'Identitas dikonfirmasi — ujian dibuka' : 'Identitas dikonfirmasi — menunggu pengawas membuka ujian');
	}

	async function startPortalExam(silent = false) {
		if (!portalAuthToken || !activeParticipantId) {
			errorMessage = 'Data kartu belum lengkap. Silakan scan ulang atau panggil pengawas.';
			return;
		}
		if (!silent) loading = true;
		if (!silent) errorMessage = '';
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
					startParticipantRuntimePolling();
					return;
				}
				throw new Error(message);
			}
			const startedPayload = normalizePortalPayload(body);
			const priorPayload = payload;
			setExamPayload({
				...startedPayload,
				student: {
					...(priorPayload?.student ?? priorPayload?.participant ?? {}),
					...(startedPayload.student ?? startedPayload.participant ?? {})
				},
				room: startedPayload.room ?? priorPayload?.room
			});
			portalStep = 'exam';
			startCountdown();
			addTelemetry('Identitas dikonfirmasi — ujian dibuka');
			startParticipantRuntimePolling();
		} catch (error) {
			if (!silent) errorMessage = friendlyError(error);
		} finally {
			if (!silent) loading = false;
		}
	}


	function portalAuthHeaders() {
		return { 'content-type': 'application/json', authorization: `Bearer ${portalAuthToken}` };
	}

	function startParticipantRuntimePolling() {
		if (!browser || !authenticatedByCard || !portalAuthToken || !activeParticipantId) return;
		stopParticipantRuntimePolling();
		void sendPortalHeartbeat();
		void pollPortalCommands();
		if (portalStep === 'waiting') void autoCheckPortalStart();
		participantPollInterval = setInterval(() => {
			void sendPortalHeartbeat();
			void pollPortalCommands();
			if (portalStep === 'waiting') void autoCheckPortalStart();
			if (portalStep === 'exam' && Object.keys(pendingAnswers).length > 0) void syncPendingAnswers();
		}, 5000);
	}

	function stopParticipantRuntimePolling() {
		if (participantPollInterval) clearInterval(participantPollInterval);
		participantPollInterval = undefined;
	}

	async function autoCheckPortalStart() {
		if (waitingAutoCheckBusy || portalStep !== 'waiting') return;
		waitingAutoCheckBusy = true;
		try {
			await startPortalExam(true);
		} finally {
			waitingAutoCheckBusy = false;
		}
	}

	async function sendPortalHeartbeat() {
		if (!authenticatedByCard || !portalAuthToken || !activeParticipantId) return;
		await fetch(`/api/cbt-portal/participants/${encodeURIComponent(activeParticipantId)}/heartbeat`, {
			method: 'POST',
			headers: portalAuthHeaders(),
			body: '{}'
		}).catch(() => undefined);
	}

	async function reportPortalEvent(eventType: string, data: Record<string, unknown> = {}) {
		if (!authenticatedByCard || !portalAuthToken || !activeParticipantId) return;
		await fetch(`/api/cbt-portal/participants/${encodeURIComponent(activeParticipantId)}/event`, {
			method: 'POST',
			headers: portalAuthHeaders(),
			body: JSON.stringify({ event_type: eventType, data: { ...data, client_time: new Date().toISOString(), portal_step: portalStep } })
		}).catch(() => undefined);
	}

	async function pollPortalCommands() {
		if (!authenticatedByCard || !portalAuthToken || !activeParticipantId) return;
		try {
			const response = await fetch(`/api/cbt-portal/participants/${encodeURIComponent(activeParticipantId)}/commands`, {
				headers: { authorization: `Bearer ${portalAuthToken}` }
			});
			const body = await response.json().catch(() => ({}));
			if (!response.ok) return;
			const commands = ((body.data?.commands ?? body.commands ?? []) as Array<{ id: string; message?: string; type?: string; issued_at?: string }>);
			const known = new Set(seenCommandIds);
			for (const command of commands) {
				if (!command.id || known.has(command.id)) continue;
				known.add(command.id);
				const message = command.message || 'Ada instruksi dari pengawas. Ikuti arahan pengawas ruang.';
				addTelemetry(`Instruksi pengawas diterima — ${message}`);
				window.alert(`Instruksi Pengawas:\n\n${message}`);
				void acknowledgePortalCommand(command.id);
			}
			seenCommandIds = known;
		} catch {
			// polling command bersifat tambahan; jangan ganggu ujian.
		}
	}

	async function acknowledgePortalCommand(commandId: string) {
		if (!authenticatedByCard || !portalAuthToken || !activeParticipantId) return;
		await fetch(`/api/cbt-portal/participants/${encodeURIComponent(activeParticipantId)}/commands/${encodeURIComponent(commandId)}/ack`, {
			method: 'POST',
			headers: portalAuthHeaders(),
			body: JSON.stringify({ status: 'seen' })
		}).catch(() => undefined);
	}

	async function syncAnswer(questionId: string, answer: string, quiet = false) {
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
		if (!quiet) addTelemetry(`Jawaban ${questionId} tersinkron`);
	}

	async function saveAnswer(questionId: string, answer: string) {
		answers = { ...answers, [questionId]: answer };
		pendingAnswers = { ...pendingAnswers, [questionId]: answer };
		try {
			await syncAnswer(questionId, answer);
		} catch {
			addTelemetry(`Jawaban ${questionId} aman lokal, menunggu sinkron`);
		}
	}

	async function syncPendingAnswers() {
		const entries = Object.entries(pendingAnswers);
		if (entries.length === 0 || syncingPendingAnswers) return;
		syncingPendingAnswers = true;
		try {
			for (const [questionId, answer] of entries) {
				try {
					await syncAnswer(questionId, answer, true);
				} catch {
					// Tetap simpan lokal; retry berikutnya akan mencoba lagi.
				}
			}
			if (Object.keys(pendingAnswers).length === 0) addTelemetry('Semua jawaban tertunda berhasil disinkron ulang');
		} finally {
			syncingPendingAnswers = false;
		}
	}

	async function submitExam() {
		if (Object.keys(pendingAnswers).length > 0) {
			await syncPendingAnswers();
		}
		if (Object.keys(pendingAnswers).length > 0) {
			errorMessage = 'Masih ada jawaban yang belum tersinkron. Tekan Sinkron ulang jawaban atau panggil pengawas sebelum kirim jawaban.';
			return;
		}
		if (!confirm('Kirim jawaban sekarang? Periksa kembali soal yang masih ragu-ragu sebelum lanjut.')) return;
		loading = true;
		errorMessage = '';
		try {
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
			submitted = true;
			stopParticipantRuntimePolling();
			addTelemetry('Ujian dikumpulkan');
		} catch (error) {
			errorMessage = error instanceof Error ? error.message : 'Submit gagal';
		} finally {
			loading = false;
		}
	}

	function toggleDoubtful(questionId: string) {
		const next = new Set(doubtfulQuestions);
		if (next.has(questionId)) next.delete(questionId);
		else next.add(questionId);
		doubtfulQuestions = next;
	}

	function goQuestion(delta: number) {
		activeQuestionIndex = Math.min(Math.max(activeQuestionIndex + delta, 0), Math.max(questions.length - 1, 0));
	}

	function isExamRuntime() {
		return portalStep === 'exam' && Boolean(payload) && !submitted;
	}

	function eventTargetName(target: EventTarget | null) {
		if (!(target instanceof HTMLElement)) return 'unknown';
		return target.tagName.toLowerCase();
	}

	function isEditableTarget(target: EventTarget | null) {
		return target instanceof HTMLInputElement || target instanceof HTMLTextAreaElement || (target instanceof HTMLElement && target.isContentEditable);
	}

	function blockExamClipboard(event: Event, eventType: 'copy_attempt' | 'cut_attempt' | 'paste_attempt' | 'context_menu_attempt' | 'drop_attempt') {
		if (!isExamRuntime()) return;
		event.preventDefault();
		event.stopPropagation();
		const target = eventTargetName(event.target);
		const label = eventType === 'paste_attempt' ? 'Tempel/paste diblokir' : eventType === 'copy_attempt' ? 'Salin/copy diblokir' : eventType === 'cut_attempt' ? 'Potong/cut diblokir' : eventType === 'drop_attempt' ? 'Tarik-lepas teks diblokir' : 'Menu klik kanan/tahan diblokir';
		addTelemetry(label);
		void reportPortalEvent(eventType, { reason: eventType, target, blocked: true });
	}

	function blockExamSelection(event: Event) {
		if (!isExamRuntime() || isEditableTarget(event.target)) return;
		event.preventDefault();
	}

	function blockPasteBeforeInput(event: InputEvent) {
		if (!isExamRuntime()) return;
		if (event.inputType === 'insertFromPaste' || event.inputType === 'insertFromDrop') {
			blockExamClipboard(event, event.inputType === 'insertFromPaste' ? 'paste_attempt' : 'drop_attempt');
		}
	}

	function blockExamShortcut(event: KeyboardEvent) {
		if (!isExamRuntime()) return;
		const key = event.key.toLowerCase();
		if ((event.ctrlKey || event.metaKey) && ['c', 'x', 'v', 'a', 's', 'p'].includes(key)) {
			event.preventDefault();
			event.stopPropagation();
			if (key === 'v') void reportPortalEvent('paste_attempt', { reason: 'keyboard_shortcut', blocked: true });
			else if (key === 'c') void reportPortalEvent('copy_attempt', { reason: 'keyboard_shortcut', blocked: true });
			else if (key === 'x') void reportPortalEvent('cut_attempt', { reason: 'keyboard_shortcut', blocked: true });
			else void reportPortalEvent('anti_cheat_keyboard_shortcut', { reason: `ctrl_${key}`, blocked: true });
			addTelemetry('Pintasan keyboard diblokir');
		}
	}

	$effect(() => {
		const token = queryCard;
		if (token && !cardToken) cardToken = token;
	});

	$effect(() => {
		if (!browser) return;
		deviceFingerprint = deviceFingerprint || makeFingerprint();
		const updateWatermarkTime = () => {
			watermarkTime = new Intl.DateTimeFormat('id-ID', { timeStyle: 'medium', timeZone: 'Asia/Makassar' }).format(new Date()) + ' WITA';
		};
		updateWatermarkTime();
		const watermarkTimer = setInterval(updateWatermarkTime, 30000);
		const onBlur = () => { addTelemetry('Halaman ujian tidak aktif sesaat'); void reportPortalEvent('web_focus_lost', { reason: 'window_blur' }); };
		const onFocus = () => { addTelemetry('Halaman ujian aktif kembali'); void reportPortalEvent('web_focus_restored', { reason: 'window_focus' }); };
		const onOffline = () => addTelemetry('Koneksi terputus, jawaban disimpan sementara');
		const onOnline = () => { addTelemetry('Koneksi kembali tersambung'); void syncPendingAnswers(); };
		const onVisibility = () => void reportPortalEvent(document.hidden ? 'web_visibility_hidden' : 'web_visibility_visible', { reason: document.hidden ? 'document_hidden' : 'document_visible' });
		const onCopy = (event: ClipboardEvent) => blockExamClipboard(event, 'copy_attempt');
		const onCut = (event: ClipboardEvent) => blockExamClipboard(event, 'cut_attempt');
		const onPaste = (event: ClipboardEvent) => blockExamClipboard(event, 'paste_attempt');
		const onContextMenu = (event: MouseEvent) => blockExamClipboard(event, 'context_menu_attempt');
		const onDrop = (event: DragEvent) => blockExamClipboard(event, 'drop_attempt');
		window.addEventListener('blur', onBlur);
		window.addEventListener('focus', onFocus);
		window.addEventListener('offline', onOffline);
		window.addEventListener('online', onOnline);
		window.addEventListener('keydown', blockExamShortcut, true);
		document.addEventListener('visibilitychange', onVisibility);
		document.addEventListener('copy', onCopy, true);
		document.addEventListener('cut', onCut, true);
		document.addEventListener('paste', onPaste, true);
		document.addEventListener('contextmenu', onContextMenu, true);
		document.addEventListener('drop', onDrop, true);
		document.addEventListener('selectstart', blockExamSelection, true);
		document.addEventListener('beforeinput', blockPasteBeforeInput, true);
		return () => {
			clearInterval(watermarkTimer);
			window.removeEventListener('blur', onBlur);
			window.removeEventListener('focus', onFocus);
			window.removeEventListener('offline', onOffline);
			window.removeEventListener('online', onOnline);
			window.removeEventListener('keydown', blockExamShortcut, true);
			document.removeEventListener('visibilitychange', onVisibility);
			document.removeEventListener('copy', onCopy, true);
			document.removeEventListener('cut', onCut, true);
			document.removeEventListener('paste', onPaste, true);
			document.removeEventListener('contextmenu', onContextMenu, true);
			document.removeEventListener('drop', onDrop, true);
			document.removeEventListener('selectstart', blockExamSelection, true);
			document.removeEventListener('beforeinput', blockPasteBeforeInput, true);
		};
	});

	onDestroy(() => { stopParticipantRuntimePolling(); stopCountdown(); });
</script>

<svelte:head>
	<title>Portal Ujian Peserta — MTsN 2 Kolaka Utara</title>
	<meta name="color-scheme" content="light" />
	<meta name="theme-color" content="#f7fbf5" />
</svelte:head>

<!-- Sacred space background with subtle geometric pattern -->
<main class="exam-light-scope min-h-dvh bg-[#f7fbf5] px-3 py-4 text-slate-950" style="color-scheme: light;">
	<section class="mx-auto flex min-h-[calc(100dvh-2rem)] max-w-md flex-col overflow-hidden rounded-[2rem] border border-emerald-100 bg-white shadow-xl">
		<header class="sticky top-0 z-20 border-b border-emerald-100 bg-emerald-800 px-4 pb-4 pt-[calc(1rem+env(safe-area-inset-top))] text-white">
			<p class="text-[10px] font-bold uppercase tracking-[0.25em] text-emerald-100">MTsN 2 Kolaka Utara</p>
			<h1 class="mt-1 text-2xl font-black">Portal Ujian Peserta</h1>
			<p class="mt-1 text-sm text-emerald-50">QR + PIN, cek identitas, lalu kerjakan ujian di halaman ini.</p>
		</header>

		<div class="flex-1 overflow-y-auto px-4 py-4 pb-[calc(5rem+env(safe-area-inset-bottom))]">
			{#if errorMessage}
				<div class="mb-3 rounded-2xl border border-red-200 bg-red-50 p-3 text-sm text-red-700" role="alert">{errorMessage}</div>
			{/if}

			{#if submitted}
				<section class="rounded-[1.5rem] border border-emerald-200 bg-emerald-50 p-5 text-center">
					<div class="mx-auto grid size-16 place-items-center rounded-full bg-emerald-700 text-2xl font-black text-white">✓</div>
					<h2 class="mt-4 text-2xl font-black">Jawaban terkirim</h2>
					<p class="mt-2 text-sm text-slate-600">Jawaban sudah terkirim. Tetap di tempat dan tunjukkan layar ini kepada pengawas bila diminta.</p>
					<button class="mt-5 min-h-12 w-full rounded-2xl border border-slate-300 bg-white px-4 text-sm font-bold" onclick={() => { payload = null; submitted = false; portalStep = 'login'; }}>Kembali ke awal</button>
				</section>
			{:else if payload && portalStep === 'confirm'}
				<section class="space-y-4">
					<div class="rounded-[1.5rem] border border-slate-200 bg-slate-50 p-4">
						<p class="text-xs font-bold uppercase tracking-[0.18em] text-emerald-700">Konfirmasi Identitas</p>
						<h2 class="mt-1 text-2xl font-black">Apakah data ini benar?</h2>
					</div>
					<div class="rounded-[1.5rem] border border-emerald-100 bg-white p-4 text-sm shadow-sm">
						<div class="space-y-3">
							<p><span class="block text-xs font-bold uppercase tracking-wide text-slate-500">Nama</span><b class="text-lg">{studentName}</b></p>
							<p><span class="block text-xs font-bold uppercase tracking-wide text-slate-500">Kelas / NIS</span>{student.class_name ?? student.class_code ?? '—'} · {student.nis ?? '—'}</p>
							<p><span class="block text-xs font-bold uppercase tracking-wide text-slate-500">Ruang / Meja</span>{student.room_name ?? '—'} / {student.seat_no ?? '—'}</p>
							<p><span class="block text-xs font-bold uppercase tracking-wide text-slate-500">Ujian</span>{session.title ?? session.subject ?? 'Sesi Ujian'}</p>
						</div>
					</div>
					<p class="rounded-2xl border border-amber-200 bg-amber-50 p-3 text-sm text-amber-950">Jika data tidak sesuai, jangan lanjut. Panggil pengawas ruang.</p>
					<button class="min-h-14 w-full rounded-2xl bg-emerald-700 px-4 text-base font-black text-white" onclick={confirmIdentity}>Ya, Masuk Ujian</button>
					<button class="min-h-12 w-full rounded-2xl border border-slate-300 px-4 text-sm font-bold" onclick={() => { payload = null; portalStep = 'login'; }}>Bukan Saya</button>
				</section>
			{:else if payload && portalStep === 'waiting'}
				<section class="rounded-[1.5rem] border border-amber-200 bg-amber-50 p-5 text-center">
					<div class="mx-auto grid size-16 place-items-center rounded-full bg-amber-400 text-sm font-black text-amber-950">SIAP</div>
					<h2 class="mt-4 text-2xl font-black">Menunggu pengawas</h2>
					<p class="mt-2 text-sm text-amber-950">Identitas sudah benar. Tetap di halaman ini. Sistem akan mengecek otomatis sampai pengawas menekan <b>Mulai Ujian</b>.</p>
					<div class="mt-4 rounded-2xl bg-white/70 p-3 text-sm"><b>{studentName}</b><br />{session.title ?? 'Sesi Ujian'} · {student.room_name ?? 'Ruang belum tercatat'}</div>
					{#if authenticatedByCard}
						<button class="mt-4 min-h-12 w-full rounded-2xl bg-emerald-700 px-4 font-black text-white disabled:opacity-60" disabled={loading} onclick={() => startPortalExam()}>{loading ? 'Mengecek...' : 'Cek Lagi Sekarang'}</button>
						<p class="mt-2 text-xs font-semibold text-amber-900">Cek otomatis berjalan tiap 5 detik.</p>
					{/if}
				</section>
			{:else if payload && portalStep === 'exam'}
				<section class="space-y-4 select-none" data-exam-anti-cheat-scope="active">
					<div class="sticky top-0 z-10 -mx-4 border-b border-slate-200 bg-white/95 px-4 py-3 backdrop-blur">
						<div class="flex items-center justify-between gap-3">
							<div class="min-w-0">
								<p class="text-xs font-bold uppercase tracking-[0.18em] text-emerald-700">{session.title ?? session.subject ?? 'Sesi Ujian'}</p>
								<h2 class="truncate text-lg font-black">Soal {Math.min(activeQuestionIndex + 1, questions.length)} dari {questions.length}</h2>
							</div>
							<div class="rounded-2xl border border-emerald-200 bg-emerald-50 px-3 py-2 text-right text-xs font-bold text-emerald-900">{answeredCount}/{questions.length}<br />terjawab</div>
						</div>
						<div class="mt-2 rounded-xl border border-slate-200 bg-slate-50 px-3 py-2 text-[11px] font-semibold leading-4 text-slate-500">
							{watermarkLine}
						</div>
					</div>

					{#if currentQuestion}
						<article class="relative overflow-hidden rounded-[1.5rem] border border-slate-200 bg-white p-4 shadow-sm">
							<div class="pointer-events-none absolute inset-x-3 top-16 -rotate-6 select-none text-center text-[11px] font-black uppercase tracking-[0.16em] text-slate-900/5" aria-hidden="true">
								{studentName} · {classLabel} · {roomLabel} · {seatLabel} · {participantWatermark}
							</div>
							<div class="mb-3 rounded-xl border border-slate-100 bg-slate-50/80 px-3 py-2 text-[10px] font-semibold leading-4 text-slate-500">
								{watermarkLine}
							</div>
							<div class="flex items-start justify-between gap-3">
								<p class="text-xs font-bold uppercase tracking-[0.18em] text-slate-500">{currentQuestion.type === 'multiple_choice' ? 'Pilihan Ganda' : 'Uraian'}</p>
								{#if doubtfulQuestions.has(currentQuestion.id)}<span class="rounded-full bg-amber-100 px-2 py-1 text-[11px] font-bold text-amber-900">Ragu-ragu</span>{/if}
							</div>
							<p class="mt-3 whitespace-pre-wrap text-base font-semibold leading-7 select-none">{currentQuestion.text}</p>
							{#if currentQuestion.options?.length}
								<div class="mt-4 grid gap-2">
									{#each currentQuestion.options as option}
										<label class="flex min-h-14 select-none items-start gap-3 rounded-2xl border p-3 text-sm {answers[currentQuestion.id] === option.label ? 'border-emerald-500 bg-emerald-50' : 'border-slate-200 bg-white'}">
											<input class="mt-1" type="radio" name={currentQuestion.id} value={option.label} checked={answers[currentQuestion.id] === option.label} onchange={() => saveAnswer(currentQuestion.id, option.label)} />
											<span class="select-none"><b>{option.label}.</b> {option.text}</span>
										</label>
									{/each}
								</div>
							{:else}
								<textarea class="mt-4 min-h-40 w-full select-text rounded-2xl border border-slate-200 bg-white p-3 text-sm text-slate-950" placeholder="Tulis jawaban..." value={answers[currentQuestion.id] ?? ''} onpaste={(event) => blockExamClipboard(event, 'paste_attempt')} ondrop={(event) => blockExamClipboard(event, 'drop_attempt')} onblur={(event) => saveAnswer(currentQuestion.id, event.currentTarget.value)}></textarea>
							{/if}
							<p class="mt-3 rounded-xl bg-slate-50 p-2 text-xs text-slate-600">{pendingAnswers[currentQuestion.id] ? 'Aman lokal, menunggu sinkron' : answers[currentQuestion.id] ? 'Tersimpan' : 'Belum dijawab'}</p>
							{#if pendingAnswerCount > 0}
								<button class="mt-2 min-h-10 w-full rounded-xl border border-amber-300 bg-amber-50 px-3 text-xs font-black text-amber-950 disabled:opacity-60" disabled={syncingPendingAnswers} onclick={syncPendingAnswers}>{syncingPendingAnswers ? 'Menyinkron...' : `Sinkron ulang ${pendingAnswerCount} jawaban`}</button>
							{/if}
						</article>
					{/if}

					<div class="rounded-2xl border border-amber-200 bg-amber-50 p-3 text-xs leading-5 text-amber-950">Soal ditampilkan satu per layar. Periksa tombol <b>Ragu-ragu</b> bila masih ingin meninjau lagi. Jika koneksi tidak stabil, jawaban disimpan sementara lalu dikirim ulang saat tersambung.</div>
				</section>
			{:else}
				<section class="space-y-4">
					<div class="rounded-[1.5rem] border border-slate-200 bg-slate-50 p-4">
						<h2 class="text-xl font-black">Masuk dengan Kartu Peserta Ujian</h2>
						<p class="mt-1 text-sm text-slate-600">Scan QR pada kartu. Jika kamera perangkat tidak tersedia, ketik kode kartu dan PIN secara manual.</p>
					</div>
					{#if showLegacyTokenLogin}
						<label class="block space-y-1 text-sm font-bold">Token Ujian<input class="min-h-12 w-full rounded-2xl border border-slate-300 bg-white px-3 text-slate-950" bind:value={examToken} autocomplete="off" /></label>
						<label class="block space-y-1 text-sm font-bold">Token Ruang<input class="min-h-12 w-full rounded-2xl border border-slate-300 bg-white px-3 text-slate-950" bind:value={roomToken} autocomplete="off" /></label>
					{:else}
						<label class="block space-y-1 text-sm font-bold">Kode Kartu / QR Token<input class="min-h-12 w-full rounded-2xl border border-slate-300 bg-white px-3 text-slate-950" bind:value={cardToken} autocomplete="off" placeholder="Terisi otomatis setelah scan QR" /></label>
						<label class="block space-y-1 text-sm font-bold">PIN<input class="min-h-14 w-full rounded-2xl border border-slate-300 bg-white px-3 text-center text-2xl tracking-[0.45em] text-slate-950" bind:value={pin} inputmode="numeric" autocomplete="one-time-code" maxlength="8" placeholder="••••" /></label>
					{/if}
					<button class="min-h-14 w-full rounded-2xl bg-emerald-700 px-4 text-base font-black text-white disabled:opacity-60" disabled={loading} onclick={portalLogin}>{loading ? 'Memproses...' : 'Lanjutkan'}</button>
					<button class="w-full text-sm font-bold text-emerald-700 underline" type="button" onclick={() => (showLegacyTokenLogin = !showLegacyTokenLogin)}>{showLegacyTokenLogin ? 'Kembali ke QR + PIN' : 'Cara lain bila QR belum bisa dipakai'}</button>
				</section>
			{/if}

			<section class="mt-4 rounded-2xl border border-slate-200 bg-slate-50 p-3 text-xs text-slate-600">
				<p class="font-bold text-slate-900">Aktivitas perangkat</p>
				{#if telemetry.length === 0}<p class="mt-1">Belum ada aktivitas tercatat.</p>{:else}<ul class="mt-2 space-y-1">{#each telemetry.slice(0, 4) as item}<li>• {item}</li>{/each}</ul>{/if}
			</section>
		</div>

		{#if payload && portalStep === 'exam' && currentQuestion}
			<nav class="fixed inset-x-0 bottom-0 z-30 mx-auto max-w-md border-t border-slate-200 bg-white/95 px-3 pb-[calc(0.55rem+env(safe-area-inset-bottom))] pt-2 backdrop-blur">
				<div class="grid grid-cols-3 gap-2">
					<button class="min-h-12 rounded-2xl bg-slate-100 text-xs font-black text-slate-700 disabled:opacity-40" disabled={activeQuestionIndex === 0} onclick={() => goQuestion(-1)}>Sebelumnya</button>
					<button class="min-h-12 rounded-2xl text-xs font-black {doubtfulQuestions.has(currentQuestion.id) ? 'bg-amber-500 text-amber-950' : 'bg-slate-100 text-slate-700'}" onclick={() => toggleDoubtful(currentQuestion.id)}>Ragu-ragu</button>
					{#if activeQuestionIndex >= questions.length - 1}
						<button class="min-h-12 rounded-2xl bg-emerald-700 text-xs font-black text-white disabled:opacity-60" disabled={loading} onclick={submitExam}>Kirim Jawaban</button>
					{:else}
						<button class="min-h-12 rounded-2xl bg-emerald-700 text-xs font-black text-white" onclick={() => goQuestion(1)}>Berikutnya</button>
					{/if}
				</div>
			</nav>
		{/if}
	</section>
</main>
