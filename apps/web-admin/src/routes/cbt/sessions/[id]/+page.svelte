<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Badge } from '$lib/components/ui/badge';
	import { toast } from '$lib/components/ui/sonner';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import OperationStatusPanel from '$lib/components/OperationStatusPanel.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { confirmAction, confirmChallenge } from '$lib/confirm-dialog';
	import { clientApiPath, clientApiPathWithQuery, readClientApiData, readClientJson } from '$lib/client/api';

	type SessionInfo = {
		id: string; title: string; package_title: string; duration_minutes: number;
		class_name: string; class_code: string; event_id: string | null;
		scope_type?: string; scope_ref?: string;
		scheduled_start: string; scheduled_end: string; status: string;
	};
	type ResultRow = {
		participant_id: string; nis: string; nama: string; gender: string;
		submitted_at: string | null; score: string | null;
		room_name?: string; seat_no?: number | null;
		total_answers: number; correct_answers: number;
	};
	type Participant = {
		id: string; nis: string; nama: string; gender: string;
		token: string; room_name: string; room_id: string | null;
		seat_no?: number | null;
		submitted_at: string | null; score: string | null;
		app_switch_count: number; screenshot_attempt: number; suspicious_flag: boolean;
		last_heartbeat: string | null;
	};
	type Room = { id: string; room_name: string; capacity: number; participant_count: number; };
	type ProctoringRow = {
		participant_id: string; nis: string; nama: string;
		token: string; room_name: string; seat_no?: number | null;
		submitted_at: string | null; last_heartbeat: string | null;
		app_switch_count: number; screenshot_attempt: number;
		suspicious_flag: boolean; answered_count: number; score: string | null;
	};
	type ProctoringEvent = {
		id: string;
		participant_id: string;
		student_id: string;
		nis: string;
		nama: string;
		room_name: string;
		event_type: string;
		event_data: unknown;
		created_at: string;
	};
	type UngradedEssay = {
		id: string; nis: string; nama: string;
		question_text: string; answer: string;
	};
	type SessionResultsDetail = {
		session: SessionInfo;
		results: ResultRow[];
	};
	type ResultsPayload = {
		session?: SessionInfo | null;
		results?: ResultRow[];
		error?: string;
		message?: string;
	};

	const sessionId = page.params.id ?? '';
	type ActiveTab = 'hasil' | 'peserta' | 'ruangan' | 'proctoring' | 'essay';

	let activeTab = $state<ActiveTab>('hasil');
	let session = $state<SessionInfo | null>(null);
	let results = $state<ResultRow[]>([]);
	let participants = $state<Participant[]>([]);
	let rooms = $state<Room[]>([]);
	let proctoring = $state<ProctoringRow[]>([]);
	let proctoringEvents = $state<ProctoringEvent[]>([]);
	let essays = $state<UngradedEssay[]>([]);
	let detailPromise = $state<Promise<SessionResultsDetail> | null>(null);
	let scoreBusy = $state(false);
	let shuffleBusy = $state(false);
	let newRoomName = $state('');
	let newRoomCap = $state(30);
	let roomBusy = $state(false);
	let seatBusy = $state(false);
	let tokenBusy = $state(false);
	let participantRefreshBusy = $state(false);
	let proctoringRefreshBusy = $state(false);
	let eventRefreshBusy = $state(false);
	let regenBusyId = $state('');
	let roomDeleteBusyId = $state('');
	let seatSaveBusyId = $state('');
	let gradeBusyId = $state('');
	let flagBusyId = $state('');
	let resetAccessBusyId = $state('');
	let forceSubmitBusyId = $state('');
	let eventPanelParticipantId = $state('');
	let procInterval: ReturnType<typeof setInterval> | null = null;
	let gradeInput = $state<Record<string, number>>({});
	let seatInput = $state<Record<string, number>>({});
	let roomInput = $state<Record<string, string>>({});
	let operationState = $state<{ tone: 'success' | 'error' | 'warning' | 'info'; title: string; message: string } | null>(null);
	let detailRequestId = 0;
	let participantsRequestId = 0;
	let roomsRequestId = 0;
	let proctoringRequestId = 0;
	let proctoringEventsRequestId = 0;
	let essaysRequestId = 0;

	const statusLabel: Record<string, string> = {
		draft: 'Draft', scheduled: 'Terjadwal', active: 'Berlangsung',
		finished: 'Selesai', cancelled: 'Dibatalkan',
	};
	const detailTabs: { id: ActiveTab; label: string }[] = [
		{ id: 'hasil', label: 'Hasil Ujian' },
		{ id: 'peserta', label: 'Peserta & Token' },
		{ id: 'ruangan', label: 'Ruangan' },
		{ id: 'proctoring', label: 'Proctoring' },
		{ id: 'essay', label: 'Koreksi Uraian' },
	];

	function statusClass(s: string) {
		if (s === 'active') return 'bg-emerald-100 text-emerald-700 border-emerald-200';
		if (s === 'finished') return 'bg-slate-100 text-slate-500 border-slate-200';
		if (s === 'cancelled') return 'bg-red-100 text-red-700 border-red-200';
		if (s === 'scheduled') return 'bg-green-100 text-green-800 border-green-200';
		return 'bg-amber-100 text-amber-700 border-amber-200';
	}

	function fmtDt(iso: string | null) {
		if (!iso) return '—';
		return new Date(iso).toLocaleString('id-ID', {
			timeZone: 'Asia/Makassar', year: 'numeric', month: 'short',
			day: 'numeric', hour: '2-digit', minute: '2-digit',
		}) + ' WITA';
	}

	function fmtScore(score: string | null) {
		if (score === null || score === undefined || score === '') return '—';
		const n = parseFloat(score);
		return isNaN(n) ? '—' : n.toFixed(1);
	}

	function scoreClass(score: string | null) {
		if (!score) return 'text-slate-400';
		const n = parseFloat(score);
		if (n >= 75) return 'text-emerald-600 font-semibold';
		if (n >= 60) return 'text-amber-600 font-semibold';
		return 'text-red-600 font-semibold';
	}

	function heartbeatBucket(hb: string | null): 'none' | 'online' | 'slow' | 'offline' {
		if (!hb) return 'none';
		const diff = (Date.now() - new Date(hb).getTime()) / 1000;
		if (diff < 60) return 'online';
		if (diff < 180) return 'slow';
		return 'offline';
	}

	function heartbeatStatus(hb: string | null): { label: string; cls: string } {
		const bucket = heartbeatBucket(hb);
		if (bucket === 'online') return { label: 'Online', cls: 'text-emerald-600' };
		if (bucket === 'slow') return { label: 'Lambat', cls: 'text-amber-600' };
		if (bucket === 'offline') return { label: 'Offline', cls: 'text-red-600' };
		return { label: 'Belum login', cls: 'text-slate-400' };
	}

	let stats = $derived({
		total: results.length,
		submitted: results.filter(r => r.submitted_at).length,
		avgScore: results.length > 0
			? results.reduce((sum, r) => sum + (r.score ? parseFloat(r.score) : 0), 0) / results.length
			: 0,
		passing: results.filter(r => r.score && parseFloat(r.score) >= 75).length,
	});

	let proctoringStats = $derived.by(() => {
		let online = 0;
		let slow = 0;
		let offline = 0;
		let submitted = 0;
		let suspicious = 0;
		let appSwitches = 0;
		let screenshots = 0;
		for (const row of proctoring) {
			const bucket = heartbeatBucket(row.last_heartbeat);
			if (row.submitted_at) submitted += 1;
			else if (bucket === 'online') online += 1;
			else if (bucket === 'slow') slow += 1;
			else if (bucket === 'offline') offline += 1;
			if (row.suspicious_flag) suspicious += 1;
			appSwitches += row.app_switch_count;
			screenshots += row.screenshot_attempt;
		}
		return { online, slow, offline, submitted, suspicious, appSwitches, screenshots };
	});

	function showToast(msg: string, ok = true) {
		if (ok) toast.success(msg);
		else toast.error(msg);
	}

	function setOperationState(
		tone: 'success' | 'error' | 'warning' | 'info',
		title: string,
		message: string,
	) {
		operationState = { tone, title, message };
	}

	function confirmPhrase(title: string, detail: string, challenge: string) {
		return confirmChallenge({
			title,
			message: detail,
			challenge,
			confirmLabel: 'Konfirmasi',
			tone: 'danger'
		});
	}

	async function fetchSessionDetail(): Promise<SessionResultsDetail> {
		const payload = await fetch(`/api/cbt/sessions/${sessionId}/results`)
			.then((response) => readClientApiData<ResultsPayload>(response, 'Gagal memuat hasil ujian'));
		if (!payload.session) throw new Error('Data sesi tidak ditemukan');
		return {
			session: payload.session,
			results: payload.results ?? [],
		};
	}

	function applySessionDetail(detail: SessionResultsDetail) {
		session = detail.session;
		results = detail.results;
	}

	function loadInitial() {
		const requestId = ++detailRequestId;
		session = null;
		results = [];
		detailPromise = fetchSessionDetail().then((detail) => {
			if (requestId !== detailRequestId) {
				if (!session) throw new Error('Permintaan detail sesi dibatalkan');
				return { session, results };
			}
			applySessionDetail(detail);
			return detail;
		}).catch((error: unknown) => {
			if (requestId === detailRequestId || !session) throw error;
			return { session, results };
		});
	}

	async function refreshSessionDetail() {
		if (!detailPromise) {
			loadInitial();
			return;
		}
		const requestId = ++detailRequestId;
		try {
			const detail = await fetchSessionDetail();
			if (requestId !== detailRequestId) return;
			applySessionDetail(detail);
			detailPromise = Promise.resolve(detail);
		} catch (error) {
			if (requestId !== detailRequestId) return;
			if (session) {
				detailPromise = Promise.resolve({ session, results });
				toast.error(detailErrorMessage(error));
			} else {
				detailPromise = Promise.reject(error);
			}
		}
	}

	function retryDetail(reset?: () => void) {
		reset?.();
		loadInitial();
	}

	function detailErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Gagal memuat hasil ujian';
	}

	function handleDetailRenderError(error: unknown) {
		console.error('CBT session detail render failed', error);
	}

	function mutationErrorMessage(error: unknown, fallback: string) {
		if (error instanceof Error && error.message.trim() && !error.message.toLowerCase().includes('fetch')) return error.message;
		return fallback;
	}

	function proctoringEventLabel(type: string): string {
		const map: Record<string, string> = {
			login: 'Login',
			heartbeat: 'Heartbeat',
			answer: 'Jawaban',
			submit: 'Submit',
			app_switch: 'Pindah Aplikasi',
			screenshot_attempt: 'Screenshot',
			copy_attempt: 'Salin Teks',
			paste_attempt: 'Tempel Teks',
			cut_attempt: 'Potong Teks',
			proctor_reset_access: 'Reset Akses',
			proctor_force_submit: 'Paksa Submit',
		};
		return map[type] ?? type.replaceAll('_', ' ');
	}

	function eventDataText(data: unknown): string {
		if (data === null || data === undefined || data === '') return '—';
		const text = typeof data === 'string' ? data : JSON.stringify(data);
		if (!text) return '—';
		return text.length > 140 ? `${text.slice(0, 140)}...` : text;
	}

	async function loadParticipants() {
		const requestId = ++participantsRequestId;
		try {
			const res = await fetch(`/api/cbt/sessions/${sessionId}/participants`);
			const rows = await readClientApiData<Participant[]>(res, 'Gagal memuat peserta');
			if (requestId !== participantsRequestId) return;
			participants = Array.isArray(rows) ? rows : [];
			seatInput = Object.fromEntries(participants.map((participant) => [participant.id, participant.seat_no ?? 0]));
			roomInput = Object.fromEntries(participants.map((participant) => [participant.id, participant.room_id ?? '']));
		} catch (error) {
			if (requestId === participantsRequestId) throw error;
		}
	}

	async function loadRooms() {
		const requestId = ++roomsRequestId;
		try {
			const res = await fetch(`/api/cbt/sessions/${sessionId}/rooms`);
			const rows = await readClientApiData<Room[]>(res, 'Gagal memuat ruangan');
			if (requestId !== roomsRequestId) return;
			rooms = Array.isArray(rows) ? rows : [];
		} catch (error) {
			if (requestId === roomsRequestId) throw error;
		}
	}

	async function loadProctoring() {
		const requestId = ++proctoringRequestId;
		try {
			const res = await fetch(`/api/cbt/sessions/${sessionId}/proctoring`);
			const rows = await readClientApiData<ProctoringRow[]>(res, 'Gagal memuat proctoring');
			if (requestId !== proctoringRequestId) return;
			proctoring = Array.isArray(rows) ? rows : [];
		} catch (error) {
			if (requestId === proctoringRequestId) throw error;
		}
	}

	async function loadProctoringEvents(participantId = eventPanelParticipantId) {
		const requestId = ++proctoringEventsRequestId;
		const params = new URLSearchParams({ limit: '100' });
		if (participantId) params.set('participant_id', participantId);
		try {
			const res = await fetch(clientApiPathWithQuery(clientApiPath`/api/cbt/sessions/${sessionId}/proctoring/events`, params));
			const rows = await readClientApiData<ProctoringEvent[]>(res, 'Gagal memuat log proctoring');
			if (requestId !== proctoringEventsRequestId) return;
			proctoringEvents = Array.isArray(rows) ? rows : [];
		} catch (error) {
			if (requestId === proctoringEventsRequestId) throw error;
		}
	}

	async function showParticipantEvents(pid: string) {
		eventPanelParticipantId = eventPanelParticipantId === pid ? '' : pid;
		eventRefreshBusy = true;
		try {
			await loadProctoringEvents();
		} catch (error) {
			toast.error(detailErrorMessage(error));
		} finally {
			eventRefreshBusy = false;
		}
	}

	async function refreshParticipants() {
		participantRefreshBusy = true;
		try {
			await loadParticipants();
		} catch (error) {
			toast.error(detailErrorMessage(error));
		} finally {
			participantRefreshBusy = false;
		}
	}

	async function refreshProctoring() {
		proctoringRefreshBusy = true;
		try {
			await Promise.all([loadProctoring(), loadProctoringEvents()]);
		} catch (error) {
			toast.error(detailErrorMessage(error));
		} finally {
			proctoringRefreshBusy = false;
		}
	}

	async function refreshProctoringEvents() {
		eventRefreshBusy = true;
		try {
			await loadProctoringEvents();
		} catch (error) {
			toast.error(detailErrorMessage(error));
		} finally {
			eventRefreshBusy = false;
		}
	}

	async function loadEssays() {
		const requestId = ++essaysRequestId;
		try {
			const res = await fetch(`/api/cbt/sessions/${sessionId}/ungraded-essays`);
			const rows = await readClientApiData<UngradedEssay[]>(res, 'Gagal memuat esai belum dinilai');
			if (requestId !== essaysRequestId) return;
			essays = Array.isArray(rows) ? rows : [];
		} catch (error) {
			if (requestId === essaysRequestId) throw error;
		}
	}

	async function switchTab(tab: ActiveTab) {
		activeTab = tab;
		try {
			if (tab === 'peserta') { await loadRooms(); await loadParticipants(); }
			if (tab === 'ruangan') { await loadRooms(); await loadParticipants(); }
			if (tab === 'essay') await loadEssays();
			if (tab === 'proctoring') {
				await Promise.all([loadProctoring(), loadProctoringEvents()]);
				if (!procInterval) {
					procInterval = setInterval(() => {
						void loadProctoring().catch((error) => toast.error(detailErrorMessage(error)));
					}, 15000);
				}
			} else {
				if (procInterval) { clearInterval(procInterval); procInterval = null; }
			}
		} catch (error) {
			toast.error(detailErrorMessage(error));
		}
	}

	async function triggerScoring() {
		if (!(await confirmPhrase('Hitung Ulang Skor', 'Sistem akan menghitung ulang skor seluruh peserta berdasarkan jawaban yang masuk. Gunakan setelah koreksi uraian atau sinkronisasi nilai.', 'NILAI ULANG'))) return;
		scoreBusy = true;
		try {
			const res = await fetch(`/api/cbt/sessions/${sessionId}/score`, { method: 'POST' });
			await readClientJson<unknown>(res);
			setOperationState('success', 'Skor Diperbarui', 'Perhitungan nilai sesi sudah disegarkan. Tinjau kembali hasil akhir sebelum menutup sesi.');
			showToast('Penilaian selesai — skor diperbarui');
			await refreshSessionDetail();
		} catch (error) {
			setOperationState('error', 'Skor Gagal Dihitung Ulang', 'Perhitungan ulang belum berhasil. Periksa data jawaban atau ulangi beberapa saat lagi.');
			showToast(mutationErrorMessage(error, 'Gagal menghitung skor'), false);
		} finally { scoreBusy = false; }
	}

	async function generateTokens() {
		if (!(await confirmPhrase('Buat Token Massal', 'Token baru akan dibuat untuk seluruh peserta sesi ini. Gunakan hanya saat token awal belum dibagikan atau harus direset terkontrol.', 'TOKEN'))) return;
		tokenBusy = true;
		try {
			const res = await fetch(`/api/cbt/sessions/${sessionId}/generate-tokens`, { method: 'POST' });
			await readClientJson<unknown>(res);
			setOperationState('success', 'Token Massal Berhasil Dibuat', 'Token peserta sudah diperbarui. Bagikan ulang token hanya ke pengawas atau peserta yang berwenang.');
			showToast('Token berhasil digenerate');
			await loadParticipants();
		} catch (error) {
			setOperationState('error', 'Token Gagal Dibuat', 'Pembuatan token massal belum berhasil. Ulangi setelah memeriksa daftar peserta sesi ini.');
			showToast(mutationErrorMessage(error, 'Gagal generate token'), false);
		} finally {
			tokenBusy = false;
		}
	}

	async function regenerateToken(pid: string) {
		if (!(await confirmAction({
			title: 'Buat Ulang Token Peserta',
			message: 'Buat ulang token peserta ini? Token lama tidak sebaiknya dipakai lagi setelah tindakan ini.',
			confirmLabel: 'Buat Ulang Token',
			tone: 'warning'
		}))) return;
		regenBusyId = pid;
		try {
			const res = await fetch(`/api/cbt/sessions/${sessionId}/participants/${pid}/regenerate-token`, { method: 'POST' });
			await readClientJson<unknown>(res);
			setOperationState('warning', 'Token Peserta Diperbarui', 'Token lama untuk peserta terkait sebaiknya tidak dipakai lagi. Pastikan pengawas membagikan token terbaru.');
			showToast('Token diperbarui');
			await loadParticipants();
		} catch (error) {
			setOperationState('error', 'Token Gagal Diperbarui', 'Pembuatan ulang token peserta belum berhasil. Coba ulang beberapa saat lagi.');
			showToast(mutationErrorMessage(error, 'Gagal regenerate token'), false);
		} finally {
			regenBusyId = '';
		}
	}

	function copyToken(token: string) {
		navigator.clipboard.writeText(token).then(() => showToast('Token disalin'));
	}

	async function createRoom() {
		if (!newRoomName.trim()) return;
		roomBusy = true;
		try {
			const res = await fetch(`/api/cbt/sessions/${sessionId}/rooms`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ room_name: newRoomName.trim(), capacity: newRoomCap }),
			});
			await readClientJson<unknown>(res);
			setOperationState('success', 'Ruangan Ditambahkan', `Ruangan "${newRoomName}" sudah tersimpan. Pastikan kapasitasnya sesuai sebelum peserta diacak.`);
			showToast(`Ruangan "${newRoomName}" ditambahkan`);
			newRoomName = ''; newRoomCap = 30;
			await loadRooms();
		} catch (error) {
			setOperationState('error', 'Ruangan Gagal Ditambahkan', 'Ruangan baru belum berhasil disimpan. Periksa nama atau kapasitas lalu coba lagi.');
			showToast(mutationErrorMessage(error, 'Gagal tambah ruangan'), false);
		} finally {
			roomBusy = false;
		}
	}

	async function deleteRoom(rid: string, roomName: string) {
		if (!(await confirmPhrase('Hapus Ruangan Sesi', `Ruangan "${roomName}" akan dihapus dari sesi dan peserta di dalamnya akan dilepas dari alokasi ruangan.`, 'RUANGAN'))) return;
		roomDeleteBusyId = rid;
		try {
			const res = await fetch(`/api/cbt/sessions/${sessionId}/rooms/${rid}`, { method: 'DELETE' });
			await readClientJson<unknown>(res);
			setOperationState('warning', 'Ruangan Dihapus', `Ruangan "${roomName}" dihapus dan peserta yang terkait perlu dialokasikan ulang.`);
			await loadRooms(); await loadParticipants();
		} catch (error) {
			setOperationState('error', 'Ruangan Gagal Dihapus', 'Ruangan belum berhasil dihapus. Pastikan sesi masih bisa diubah lalu coba lagi.');
			showToast(mutationErrorMessage(error, 'Gagal menghapus ruangan'), false);
		} finally {
			roomDeleteBusyId = '';
		}
	}

	async function shuffleRooms() {
		if (!(await confirmPhrase('Acak Peserta ke Ruangan', 'Sistem akan menghapus alokasi ruangan sebelumnya dan membagikan ulang peserta secara otomatis. Pastikan daftar ruangan dan kapasitas sudah final.', 'ACAK'))) return;
		shuffleBusy = true;
		try {
			const res = await fetch(`/api/cbt/sessions/${sessionId}/shuffle-rooms`, { method: 'POST' });
			await readClientJson<unknown>(res);
			setOperationState('warning', 'Alokasi Ruangan Diperbarui', 'Peserta sudah diacak ulang ke ruangan. Periksa kembali pembagian sebelum ujian dimulai.');
			showToast('Peserta berhasil diacak ke ruangan');
			await loadRooms(); await loadParticipants();
		} catch (error) {
			setOperationState('error', 'Pengacakan Ruangan Gagal', 'Sistem belum berhasil mengacak peserta ke ruangan. Cek kapasitas ruangan atau ulangi lagi.');
			showToast(mutationErrorMessage(error, 'Gagal mengacak ruangan'), false);
		} finally {
			shuffleBusy = false;
		}
	}

	async function autoAssignSeats() {
		if (!(await confirmPhrase('Atur Nomor Meja Otomatis', 'Nomor meja peserta akan diurutkan ulang per ruangan. Gunakan setelah alokasi ruangan sudah final.', 'MEJA'))) return;
		seatBusy = true;
		try {
			const res = await fetch(`/api/cbt/sessions/${sessionId}/seats/auto`, { method: 'POST' });
			await readClientJson<unknown>(res);
			setOperationState('success', 'Nomor Meja Diatur Otomatis', 'Nomor meja peserta sudah diperbarui. Lanjutkan ke pengecekan akhir per ruangan jika diperlukan.');
			showToast('Nomor meja berhasil diurutkan otomatis');
			await loadParticipants();
		} catch (error) {
			setOperationState('error', 'Nomor Meja Gagal Diatur', 'Pengaturan otomatis belum berhasil. Pastikan peserta sudah punya ruangan lalu coba lagi.');
			showToast(mutationErrorMessage(error, 'Gagal mengatur nomor meja otomatis'), false);
		} finally {
			seatBusy = false;
		}
	}

	async function assignSeat(pid: string) {
		const roomId = roomInput[pid];
		const seatNo = seatInput[pid];
		if (!roomId || !seatNo || seatNo <= 0) {
			showToast('Pilih ruangan dan isi nomor meja yang valid', false);
			return;
		}
		seatSaveBusyId = pid;
		try {
			const res = await fetch(`/api/cbt/sessions/${sessionId}/participants/${pid}/seat`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ room_id: roomId, seat_no: seatNo }),
			});
			await readClientJson<unknown>(res);
			setOperationState('success', 'Nomor Meja Peserta Disimpan', 'Ruangan dan nomor meja peserta sudah diperbarui sesuai pengaturan operator.');
			showToast('No meja peserta diperbarui');
			await loadParticipants();
		} catch (error) {
			setOperationState('error', 'Nomor Meja Gagal Disimpan', 'Perubahan ruangan atau nomor meja belum berhasil. Periksa input dan ulangi lagi.');
			showToast(mutationErrorMessage(error, 'Gagal menyimpan nomor meja'), false);
		} finally {
			seatSaveBusyId = '';
		}
	}

	async function flagParticipant(pid: string, flag: boolean) {
		flagBusyId = pid;
		try {
			const res = await fetch(`/api/cbt/sessions/${sessionId}/participants/${pid}/flag`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ flag }),
			});
			await readClientJson<unknown>(res);
			setOperationState(
				flag ? 'warning' : 'success',
				flag ? 'Peserta Diberi Tanda' : 'Tanda Peserta Dihapus',
				flag
					? 'Peserta ditandai untuk perhatian pengawas. Tinjau kembali aktivitas proctoring sebelum mengambil langkah lanjutan.'
					: 'Tanda kecurigaan pada peserta sudah dibersihkan dari daftar proctoring.',
			);
			await loadProctoring();
		} catch (error) {
			setOperationState('error', 'Tanda Peserta Gagal Diperbarui', 'Perubahan tanda proctoring belum berhasil. Coba ulang beberapa saat lagi.');
			showToast(mutationErrorMessage(error, 'Gagal memperbarui tanda peserta'), false);
		} finally {
			flagBusyId = '';
		}
	}

	async function resetParticipantAccess(pid: string, nama: string) {
		if (!(await confirmPhrase('Reset Akses Ujian', `Akses perangkat untuk ${nama} akan dilepas sehingga peserta dapat login ulang. Gunakan hanya setelah diverifikasi oleh pengawas.`, 'RESET AKSES'))) return;
		resetAccessBusyId = pid;
		try {
			const res = await fetch(clientApiPath`/api/cbt/sessions/${sessionId}/participants/${pid}/reset-access`, { method: 'POST' });
			await readClientJson<unknown>(res);
			setOperationState('warning', 'Akses Peserta Direset', `${nama} dapat login ulang setelah pengawas memastikan perangkat yang dipakai benar.`);
			toast.success('Akses peserta direset');
			eventPanelParticipantId = pid;
			await Promise.all([loadProctoring(), loadProctoringEvents(pid)]);
		} catch (error) {
			setOperationState('error', 'Reset Akses Gagal', 'Akses peserta belum berhasil direset. Ulangi setelah memeriksa status sesi.');
			toast.error(mutationErrorMessage(error, 'Gagal reset akses peserta'));
		} finally {
			resetAccessBusyId = '';
		}
	}

	async function forceSubmitParticipant(pid: string, nama: string) {
		if (!(await confirmPhrase('Paksa Submit Peserta', `Jawaban ${nama} akan dikunci dan skor objektif dihitung dari jawaban yang sudah tersimpan. Tindakan ini untuk kondisi darurat operasional.`, 'PAKSA SUBMIT'))) return;
		forceSubmitBusyId = pid;
		try {
			const res = await fetch(clientApiPath`/api/cbt/sessions/${sessionId}/participants/${pid}/force-submit`, { method: 'POST' });
			await readClientJson<unknown>(res);
			setOperationState('warning', 'Peserta Dipaksa Submit', `${nama} sudah ditandai submit oleh proktor. Periksa hasil akhir sebelum menutup sesi.`);
			toast.success('Peserta disubmit oleh proktor');
			eventPanelParticipantId = pid;
			await Promise.all([loadProctoring(), loadProctoringEvents(pid), refreshSessionDetail()]);
		} catch (error) {
			setOperationState('error', 'Paksa Submit Gagal', 'Sistem belum berhasil mengunci submit peserta. Periksa status waktu ujian dan ulangi bila perlu.');
			toast.error(mutationErrorMessage(error, 'Gagal paksa submit peserta'));
		} finally {
			forceSubmitBusyId = '';
		}
	}

	async function submitGrade(aid: string) {
		const score = gradeInput[aid];
		if (score === undefined || score < 0 || score > 100) {
			showToast('Nilai harus 0-100', false);
			return;
		}
		gradeBusyId = aid;
		try {
			const res = await fetch(`/api/cbt/sessions/${sessionId}/answers/${aid}/grade-essay`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ manual_score: score }),
			});
			await readClientJson<unknown>(res);
			setOperationState('success', 'Nilai Uraian Tersimpan', 'Koreksi uraian sudah masuk ke sistem. Lanjutkan ke jawaban uraian berikutnya bila masih ada.');
			showToast('Nilai berhasil disimpan');
			await loadEssays();
		} catch (error) {
			setOperationState('error', 'Nilai Uraian Gagal Disimpan', 'Koreksi belum berhasil tersimpan. Ulangi setelah memastikan skor sudah valid.');
			showToast(mutationErrorMessage(error, 'Gagal menyimpan nilai'), false);
		} finally {
			gradeBusyId = '';
		}
	}

	function exportCSV() {
		if (!session || results.length === 0) return;
		const header = 'NIS,Nama,L/P,Jawaban Masuk,Benar,Skor,Waktu Submit';
		const rows = results.map(r =>
			[r.nis, `"${r.nama}"`, r.gender, r.total_answers, r.correct_answers,
			 fmtScore(r.score), r.submitted_at ? fmtDt(r.submitted_at) : ''].join(',')
		);
		const csv = [header, ...rows].join('\n');
		const blob = new Blob([csv], { type: 'text/csv' });
		const url = URL.createObjectURL(blob);
		const a = document.createElement('a');
		a.href = url;
		a.download = `hasil_${session.title.replace(/\s+/g, '_')}.csv`;
		a.click();
		URL.revokeObjectURL(url);
	}

	onMount(() => {
		void loadInitial();
	});
	onDestroy(() => { if (procInterval) clearInterval(procInterval); });
</script>

<svelte:head>
	<title>{session?.title ?? 'Detail Sesi'} — MTSN 2 Kolut</title>
</svelte:head>

<div class="space-y-5 p-6">
	<!-- Breadcrumb -->
	<div class="flex items-center gap-2 text-sm text-slate-500">
		<a href={resolve('/cbt/sessions')} class="hover:text-slate-700">Sesi Ujian</a>
		<span>/</span>
		<span class="text-slate-700 font-medium truncate max-w-xs">{session?.title ?? '...'}</span>
	</div>

	{#if operationState}
		<OperationStatusPanel {...operationState} />
	{/if}

	<AsyncContent promise={detailPromise} onerror={handleDetailRenderError}>
		{#snippet pending()}
			<div class="space-y-4">
				<div class="space-y-2">
					<Skeleton class="h-4 w-56" />
					<Skeleton class="h-8 w-80" />
					<Skeleton class="h-4 w-96" />
				</div>
				<div class="grid grid-cols-2 gap-3 sm:grid-cols-4">
					{#each Array.from({ length: 4 }) as _, index (`cbt-session-detail-stat-${index}`)}
						<Card.Root class="border-green-100">
							<Card.Content class="space-y-2 px-4 pb-3 pt-4">
								<Skeleton class="h-4 w-24" />
								<Skeleton class="h-8 w-16" />
							</Card.Content>
						</Card.Root>
					{/each}
				</div>
				<Card.Root>
					<Card.Content class="space-y-3 p-6">
						{#each Array.from({ length: 5 }) as _, index (`cbt-session-detail-row-${index}`)}
							<div class="grid gap-3 lg:grid-cols-[1fr_0.8fr_0.8fr_0.8fr_0.6fr_0.6fr_0.7fr_auto] lg:items-center">
								<Skeleton class="h-5 w-24" />
								<Skeleton class="h-5 w-32" />
								<Skeleton class="h-5 w-10" />
								<Skeleton class="h-5 w-12" />
								<Skeleton class="h-5 w-12" />
								<Skeleton class="h-5 w-12" />
								<Skeleton class="h-5 w-28" />
								<Skeleton class="h-9 w-24 justify-self-end" />
							</div>
						{/each}
					</Card.Content>
				</Card.Root>
			</div>
		{/snippet}

		{#snippet failed(error, reset)}
			<RecoveryPanel
				title="Detail Sesi Belum Tersaji"
				message={detailErrorMessage(error)}
				onRetry={() => retryDetail(reset)}
			/>
		{/snippet}

		{#snippet children(value)}
			{@const detail = value as SessionResultsDetail}
			{@const currentSession = detail.session}
			{@const currentResults = detail.results}
		<!-- Session header -->
		<div class="flex items-start justify-between gap-4 flex-wrap">
			<div>
				<h1 class="text-2xl font-semibold text-[oklch(0.38_0.13_145)]">{currentSession.title}</h1>
				<div class="flex flex-wrap gap-2 mt-2 text-sm text-slate-500">
					<span>{currentSession.package_title}</span>
					{#if currentSession.class_code}<span>· Kelas {currentSession.class_code}</span>{/if}
					<span>· {currentSession.duration_minutes} menit</span>
					<span>· {fmtDt(currentSession.scheduled_start)}</span>
				</div>
			</div>
				<div class="flex items-center gap-2 flex-wrap">
					<Badge class={statusClass(currentSession.status)}>{statusLabel[currentSession.status] ?? currentSession.status}</Badge>
				{#if currentSession.status === 'finished' || currentSession.status === 'active'}
					<LoadingButton size="sm" variant="outline" disabled={scoreBusy} onclick={triggerScoring} loading={scoreBusy} loadingLabel="Menghitung...">
						⟳ Hitung Skor
					</LoadingButton>
				{/if}
					{#if currentResults.length > 0}
						<Button size="sm" variant="outline" onclick={exportCSV}>↓ CSV</Button>
					{/if}
					<a href={resolve(`/cbt/sessions/${sessionId}/minutes`)} class="inline-flex items-center rounded-md border border-input bg-background px-3 py-2 text-sm font-medium text-slate-700 hover:bg-muted">
						Berita Acara
					</a>
					{#if currentSession.event_id}
						<a href={resolve(`/cbt/events/${currentSession.event_id}/exam-cards`)} class="inline-flex items-center rounded-md border border-input bg-background px-3 py-2 text-sm font-medium text-slate-700 hover:bg-muted">
							Kartu Ujian Event
						</a>
					{/if}
				</div>
		</div>

		<!-- Stats -->
		<div class="grid grid-cols-2 gap-3 sm:grid-cols-4">
			{#each [
				{ label: 'Total Peserta', val: stats.total.toString() },
				{ label: 'Sudah Submit', val: stats.submitted.toString() },
				{ label: 'Rata-rata Nilai', val: stats.total > 0 ? stats.avgScore.toFixed(1) : '—' },
				{ label: 'Lulus (≥75)', val: `${stats.passing} / ${stats.submitted}` },
			] as s (s.label)}
				<Card.Root class="border-green-100">
					<Card.Content class="pt-4 pb-3 px-4">
						<p class="text-xs text-slate-500 mb-1">{s.label}</p>
						<p class="text-2xl font-bold text-[oklch(0.38_0.13_145)]">{s.val}</p>
					</Card.Content>
				</Card.Root>
			{/each}
		</div>

		<!-- Tabs -->
		<div class="border-b border-green-100">
			<nav class="flex gap-1">
				{#each detailTabs as tab (tab.id)}
					<button
						onclick={() => switchTab(tab.id)}
						class="px-4 py-2 text-sm font-medium border-b-2 transition-colors {activeTab === tab.id
							? 'border-[oklch(0.38_0.13_145)] text-[oklch(0.38_0.13_145)]'
							: 'border-transparent text-slate-500 hover:text-slate-700 hover:border-slate-300'}"
					>
						{tab.label}
					</button>
				{/each}
			</nav>
		</div>

		<!-- Tab: Hasil -->
		{#if activeTab === 'hasil'}
			<Card.Root>
				<Card.Header class="pb-2">
				<Card.Title class="text-base">Daftar Nilai ({results.length} peserta)</Card.Title>
				</Card.Header>
				<Card.Content class="p-0 overflow-x-auto">
					<Table.Root>
						<Table.Header>
							<Table.Row class="bg-green-50">
								<Table.Head class="w-8">#</Table.Head>
								<Table.Head>NIS</Table.Head>
								<Table.Head>Nama</Table.Head>
								<Table.Head>L/P</Table.Head>
								<Table.Head class="text-center">Jawaban</Table.Head>
								<Table.Head class="text-center">Benar</Table.Head>
								<Table.Head class="text-center">Skor</Table.Head>
								<Table.Head>Waktu Submit</Table.Head>
							</Table.Row>
						</Table.Header>
						<Table.Body>
							{#each currentResults as r, i (r.participant_id)}
								<Table.Row>
									<Table.Cell class="text-slate-400 text-xs">{i + 1}</Table.Cell>
									<Table.Cell class="font-mono text-sm">{r.nis}</Table.Cell>
									<Table.Cell class="font-medium">{r.nama}</Table.Cell>
									<Table.Cell><Badge variant="outline" class="text-xs">{r.gender}</Badge></Table.Cell>
									<Table.Cell class="text-center text-sm">{r.total_answers}</Table.Cell>
									<Table.Cell class="text-center text-sm">{r.correct_answers}</Table.Cell>
									<Table.Cell class="text-center"><span class={scoreClass(r.score)}>{fmtScore(r.score)}</span></Table.Cell>
									<Table.Cell class="text-slate-500 text-xs whitespace-nowrap">
										{r.submitted_at ? fmtDt(r.submitted_at) : '<span class="text-slate-300">Belum submit</span>'}
									</Table.Cell>
								</Table.Row>
							{:else}
								<Table.Row>
									<Table.Cell colspan={8} class="text-center text-slate-400 py-10">Belum ada data nilai</Table.Cell>
								</Table.Row>
							{/each}
						</Table.Body>
					</Table.Root>
				</Card.Content>
			</Card.Root>

		<!-- Tab: Peserta & Token -->
		{:else if activeTab === 'peserta'}
			<div class="flex gap-2 flex-wrap">
				<LoadingButton variant="outline" size="sm" onclick={() => void generateTokens()} loading={tokenBusy} disabled={tokenBusy} loadingLabel="Membuat token...">⚡ Buat Token Massal</LoadingButton>
				<LoadingButton variant="outline" size="sm" onclick={() => void refreshParticipants()} loading={participantRefreshBusy} loadingLabel="Memuat..." disabled={participantRefreshBusy}>↻ Refresh</LoadingButton>
			</div>
			<OperationStatusPanel
				tone="warning"
				compact
				title="Aksi Sensitif Peserta"
				message="Pembuatan token massal, ubah token, dan simpan nomor meja akan langsung mengubah data operasional ujian. Pastikan pengawas sudah siap menerima perubahan terbaru."
			/>
			<Card.Root>
				<Card.Content class="p-0 overflow-x-auto">
					<Table.Root>
						<Table.Header>
							<Table.Row class="bg-green-50">
								<Table.Head>NIS</Table.Head>
								<Table.Head>Nama</Table.Head>
								<Table.Head>L/P</Table.Head>
								<Table.Head>Ruangan</Table.Head>
								<Table.Head>No Meja</Table.Head>
								<Table.Head>Token</Table.Head>
								<Table.Head>Status</Table.Head>
								<Table.Head class="text-right">Aksi</Table.Head>
							</Table.Row>
						</Table.Header>
						<Table.Body>
								{#each participants as p (p.id)}
									<Table.Row class={p.suspicious_flag ? 'bg-red-50' : 'hover:bg-green-50/30'}>
									<Table.Cell class="font-mono text-sm">{p.nis}</Table.Cell>
									<Table.Cell class="font-medium">
										{p.nama}
										{#if p.suspicious_flag}<span class="ml-1 text-red-500 text-xs">⚑ Dicurigai</span>{/if}
									</Table.Cell>
										<Table.Cell><Badge variant="outline" class="text-xs">{p.gender}</Badge></Table.Cell>
										<Table.Cell class="text-sm text-muted-foreground">{p.room_name || '—'}</Table.Cell>
										<Table.Cell class="text-sm text-muted-foreground">{p.seat_no ?? '—'}</Table.Cell>
										<Table.Cell>
										{#if p.token}
											<code class="bg-green-50 text-green-800 px-2 py-0.5 rounded text-xs font-mono border border-green-200">
												{p.token}
											</code>
										{:else}
											<span class="text-slate-400 text-xs">—</span>
										{/if}
									</Table.Cell>
									<Table.Cell>
										{#if p.submitted_at}
											<Badge variant="outline" class="text-xs bg-slate-100 text-slate-500">Submit</Badge>
										{:else}
											<Badge variant="outline" class="text-xs bg-amber-50 text-amber-700 border-amber-200">Belum</Badge>
										{/if}
									</Table.Cell>
										<Table.Cell class="text-right">
											<div class="flex items-center justify-end gap-1">
												<select bind:value={roomInput[p.id]} class="h-8 rounded-md border border-input bg-background px-2 text-xs">
													<option value="">Ruangan</option>
													{#each rooms as room (room.id)}
														<option value={room.id}>{room.room_name}</option>
													{/each}
												</select>
												<Input bind:value={seatInput[p.id]} type="number" min="1" class="h-8 w-16" />
												<LoadingButton variant="outline" size="sm" onclick={() => assignSeat(p.id)} loading={seatSaveBusyId === p.id} disabled={(seatSaveBusyId !== '' && seatSaveBusyId !== p.id) || seatBusy} loadingLabel="Menyimpan...">Simpan</LoadingButton>
												{#if p.token}
													<Button variant="outline" size="sm" onclick={() => copyToken(p.token)}>Salin</Button>
												{/if}
											<LoadingButton variant="outline" size="sm" onclick={() => regenerateToken(p.id)} loading={regenBusyId === p.id} disabled={regenBusyId !== '' && regenBusyId !== p.id} loadingLabel="Membuat ulang...">Buat Ulang</LoadingButton>
										</div>
									</Table.Cell>
								</Table.Row>
								{:else}
									<Table.Row>
										<Table.Cell colspan={8} class="text-center text-slate-400 py-8">Belum ada peserta</Table.Cell>
									</Table.Row>
								{/each}
						</Table.Body>
					</Table.Root>
				</Card.Content>
			</Card.Root>

		<!-- Tab: Ruangan -->
		{:else if activeTab === 'ruangan'}
			<Card.Root class="border-green-200">
				<Card.Header class="pb-3">
					<Card.Title class="text-base">Tambah Ruangan Baru</Card.Title>
				</Card.Header>
				<Card.Content>
					<div class="flex gap-3 flex-wrap items-end">
						<div>
							<label for="r-name" class="block text-sm font-medium mb-1">Nama Ruangan</label>
							<Input id="r-name" bind:value={newRoomName} placeholder="Ruang 1 / Lab Komputer A" class="w-48" />
						</div>
						<div>
							<label for="r-cap" class="block text-sm font-medium mb-1">Kapasitas</label>
							<Input id="r-cap" type="number" bind:value={newRoomCap} min={1} max={100} class="w-24" />
						</div>
						<LoadingButton onclick={() => void createRoom()} loading={roomBusy} loadingLabel="Menyimpan..." disabled={roomBusy || !newRoomName.trim()}>
							+ Tambah Ruangan
						</LoadingButton>
							{#if rooms.length > 0}
								<LoadingButton variant="outline" loading={shuffleBusy} loadingLabel="Mengacak..." disabled={shuffleBusy} onclick={shuffleRooms}
									class="border-amber-300 text-amber-700 hover:bg-amber-50">
									🔀 Acak Peserta ke Ruangan
								</LoadingButton>
								<LoadingButton variant="outline" loading={seatBusy} loadingLabel="Mengatur..." disabled={seatBusy} onclick={autoAssignSeats}>
									🪑 Atur Nomor Meja
								</LoadingButton>
							{/if}
					</div>
				</Card.Content>
			</Card.Root>
			<OperationStatusPanel
				tone="warning"
				compact
				title="Aksi Sensitif Ruangan"
				message="Pengacakan ruangan dan pengaturan nomor meja otomatis akan menimpa penempatan sebelumnya. Gunakan hanya setelah kapasitas dan daftar ruangan sudah final."
			/>

			{#if rooms.length > 0}
				<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 mt-4">
					{#each rooms as room (room.id)}
						<Card.Root class="border-green-100">
							<Card.Content class="p-4">
								<div class="flex items-start justify-between">
									<div>
										<div class="font-semibold text-[oklch(0.38_0.13_145)]">{room.room_name}</div>
										<div class="text-sm text-muted-foreground mt-1">
											Kapasitas: {room.capacity} · Terisi: {room.participant_count}
										</div>
										<div class="mt-2 h-2 rounded-full bg-green-100 overflow-hidden">
											<div class="h-full bg-[oklch(0.38_0.13_145)] rounded-full transition-all"
												style="width: {Math.min(100, (room.participant_count / room.capacity) * 100)}%">
											</div>
										</div>
									</div>
									<LoadingButton variant="outline" size="sm"
										class="border-red-200 text-red-600 hover:bg-red-50 ml-3"
										onclick={() => deleteRoom(room.id, room.room_name)}
										loading={roomDeleteBusyId === room.id}
										disabled={roomDeleteBusyId !== '' && roomDeleteBusyId !== room.id}
										loadingLabel="Menghapus...">Hapus</LoadingButton>
								</div>
							</Card.Content>
						</Card.Root>
					{/each}
				</div>

				<!-- Participants by room -->
				<Card.Root class="mt-4">
					<Card.Header class="pb-2">
						<Card.Title class="text-base">Peserta per Ruangan</Card.Title>
					</Card.Header>
					<Card.Content class="p-0 overflow-x-auto">
						<Table.Root>
							<Table.Header>
								<Table.Row class="bg-green-50">
									<Table.Head>NIS</Table.Head>
									<Table.Head>Nama</Table.Head>
										<Table.Head>L/P</Table.Head>
										<Table.Head>Ruangan</Table.Head>
										<Table.Head>No Meja</Table.Head>
										<Table.Head>Token</Table.Head>
								</Table.Row>
							</Table.Header>
							<Table.Body>
								{#each participants as p (p.id)}
									<Table.Row>
										<Table.Cell class="font-mono text-sm">{p.nis}</Table.Cell>
										<Table.Cell class="font-medium">{p.nama}</Table.Cell>
										<Table.Cell><Badge variant="outline" class="text-xs">{p.gender}</Badge></Table.Cell>
										{#if p.room_name}
											<Table.Cell class="text-sm">{p.room_name}</Table.Cell>
											{:else}
												<Table.Cell class="text-sm text-slate-400">Belum ditentukan</Table.Cell>
											{/if}
											<Table.Cell class="text-sm text-slate-600">{p.seat_no ?? '—'}</Table.Cell>
											<Table.Cell>
											{#if p.token}
												<code class="bg-green-50 text-green-800 px-2 py-0.5 rounded text-xs font-mono border border-green-200">{p.token}</code>
											{:else}
												<span class="text-slate-400 text-xs">—</span>
											{/if}
										</Table.Cell>
									</Table.Row>
								{/each}
							</Table.Body>
						</Table.Root>
					</Card.Content>
				</Card.Root>
			{:else}
				<div class="rounded-lg border border-dashed border-green-200 p-8 text-center text-muted-foreground text-sm mt-4">
					Belum ada ruangan. Tambah ruangan di atas, lalu klik "Acak Peserta ke Ruangan".
				</div>
			{/if}

		<!-- Tab: Proctoring -->
		{:else if activeTab === 'proctoring'}
			{@const selectedEventParticipant = proctoring.find((row) => row.participant_id === eventPanelParticipantId)}
			<div class="flex items-center justify-between gap-3 mb-4 flex-wrap">
				<div>
					<p class="text-sm font-medium text-slate-700">Monitoring proctoring live</p>
					<p class="text-xs text-muted-foreground">Pembaruan otomatis setiap 15 detik</p>
				</div>
				<LoadingButton variant="outline" size="sm" onclick={() => void refreshProctoring()} loading={proctoringRefreshBusy} loadingLabel="Memuat..." disabled={proctoringRefreshBusy}>↻ Refresh Sekarang</LoadingButton>
			</div>
			<div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
				{#each [
					{ label: 'Online Aktif', value: proctoringStats.online.toString(), className: 'text-emerald-700' },
					{ label: 'Lambat / Offline', value: `${proctoringStats.slow + proctoringStats.offline}`, className: 'text-amber-700' },
					{ label: 'Perlu Atensi', value: proctoringStats.suspicious.toString(), className: 'text-red-700' },
					{ label: 'App Switch / Screenshot', value: `${proctoringStats.appSwitches} / ${proctoringStats.screenshots}`, className: 'text-slate-700' },
				] as item (item.label)}
					<Card.Root class="border-green-100">
						<Card.Content class="px-4 pb-3 pt-4">
							<p class="mb-1 text-xs text-slate-500">{item.label}</p>
							<p class={`text-2xl font-bold ${item.className}`}>{item.value}</p>
						</Card.Content>
					</Card.Root>
				{/each}
			</div>
			<Card.Root>
				<Card.Content class="p-0 overflow-x-auto">
					<Table.Root>
						<Table.Header>
							<Table.Row class="bg-green-50">
								<Table.Head>Nama</Table.Head>
								<Table.Head>Ruangan</Table.Head>
								<Table.Head>Status</Table.Head>
								<Table.Head class="text-center">Dijawab</Table.Head>
								<Table.Head class="text-center">App Switch</Table.Head>
								<Table.Head class="text-center">Screenshot</Table.Head>
								<Table.Head>Submit</Table.Head>
								<Table.Head class="text-right">Aksi Proktor</Table.Head>
							</Table.Row>
						</Table.Header>
						<Table.Body>
								{#each proctoring as p (p.participant_id)}
									{@const hb = heartbeatStatus(p.last_heartbeat)}
									<Table.Row class={p.suspicious_flag ? 'bg-red-50' : p.app_switch_count >= 3 ? 'bg-amber-50/50' : 'hover:bg-green-50/30'}>
									<Table.Cell class="font-medium">
										{p.nama}
										<div class="text-xs text-muted-foreground font-mono">{p.nis}</div>
									</Table.Cell>
									<Table.Cell class="text-sm text-muted-foreground">{p.room_name || '—'}</Table.Cell>
									<Table.Cell>
										<span class="text-sm font-medium {hb.cls}">{p.submitted_at ? '✅ Submit' : hb.label}</span>
									</Table.Cell>
									<Table.Cell class="text-center font-mono text-sm">{p.answered_count}</Table.Cell>
									<Table.Cell class="text-center">
										<span class="font-mono text-sm {p.app_switch_count >= 3 ? 'text-red-600 font-bold' : 'text-slate-600'}">
											{p.app_switch_count}x
										</span>
									</Table.Cell>
									<Table.Cell class="text-center">
										<span class="font-mono text-sm {p.screenshot_attempt > 0 ? 'text-amber-600 font-semibold' : 'text-slate-400'}">
											{p.screenshot_attempt}x
										</span>
									</Table.Cell>
									<Table.Cell class="text-xs text-muted-foreground whitespace-nowrap">
										{p.submitted_at ? fmtDt(p.submitted_at) : '—'}
									</Table.Cell>
									<Table.Cell class="text-right">
										<div class="flex flex-wrap items-center justify-end gap-1">
											<Button
												variant="outline" size="sm"
												class={eventPanelParticipantId === p.participant_id ? 'border-green-500 text-green-800 bg-green-50' : 'border-slate-200 text-slate-600'}
												onclick={() => showParticipantEvents(p.participant_id)}
												disabled={eventRefreshBusy}>
												Log
											</Button>
											<LoadingButton
												variant="outline" size="sm"
												onclick={() => resetParticipantAccess(p.participant_id, p.nama)}
												loading={resetAccessBusyId === p.participant_id}
												disabled={resetAccessBusyId !== '' && resetAccessBusyId !== p.participant_id}
												loadingLabel="Reset...">
												Reset Akses
											</LoadingButton>
											<LoadingButton
												variant="outline" size="sm"
												class="border-amber-300 text-amber-700 hover:bg-amber-50"
												onclick={() => forceSubmitParticipant(p.participant_id, p.nama)}
												loading={forceSubmitBusyId === p.participant_id}
												disabled={!!p.submitted_at || (forceSubmitBusyId !== '' && forceSubmitBusyId !== p.participant_id)}
												loadingLabel="Submit...">
												Paksa Submit
											</LoadingButton>
											<Button
												variant="outline" size="sm"
												class={p.suspicious_flag ? 'border-red-400 text-red-700 bg-red-50' : 'border-slate-200 text-slate-500'}
												onclick={() => flagParticipant(p.participant_id, !p.suspicious_flag)}
												disabled={flagBusyId === p.participant_id}>
											{p.suspicious_flag ? '⚑ Hapus Tanda' : '⚐ Tandai'}
										</Button>
										</div>
									</Table.Cell>
								</Table.Row>
							{:else}
								<Table.Row>
									<Table.Cell colspan={8} class="text-center text-slate-400 py-8">Belum ada data proctoring</Table.Cell>
								</Table.Row>
							{/each}
						</Table.Body>
					</Table.Root>
				</Card.Content>
			</Card.Root>
			<Card.Root class="border-green-100">
				<Card.Header class="pb-3">
					<div class="flex items-start justify-between gap-3 flex-wrap">
						<div>
							<Card.Title class="text-base">
								Log Aktivitas {selectedEventParticipant ? selectedEventParticipant.nama : 'Semua Peserta'}
							</Card.Title>
							<p class="text-xs text-muted-foreground mt-1">
								{selectedEventParticipant
									? `${selectedEventParticipant.nis} · ${selectedEventParticipant.room_name || 'Tanpa ruangan'}`
									: '100 aktivitas terbaru dari sesi ini'}
							</p>
						</div>
						<div class="flex gap-2">
							{#if eventPanelParticipantId}
								<Button
									variant="outline"
									size="sm"
									onclick={() => {
										eventPanelParticipantId = '';
										void refreshProctoringEvents();
									}}
								>
									Semua Log
								</Button>
							{/if}
							<LoadingButton variant="outline" size="sm" onclick={() => void refreshProctoringEvents()} loading={eventRefreshBusy} loadingLabel="Memuat..." disabled={eventRefreshBusy}>
								↻ Refresh Log
							</LoadingButton>
						</div>
					</div>
				</Card.Header>
				<Card.Content class="p-0 overflow-x-auto">
					<Table.Root>
						<Table.Header>
							<Table.Row class="bg-slate-50">
								<Table.Head>Waktu</Table.Head>
								<Table.Head>Peserta</Table.Head>
								<Table.Head>Aktivitas</Table.Head>
								<Table.Head>Ruangan</Table.Head>
								<Table.Head>Data</Table.Head>
							</Table.Row>
						</Table.Header>
						<Table.Body>
							{#each proctoringEvents as ev (ev.id)}
								<Table.Row>
									<Table.Cell class="whitespace-nowrap text-xs text-muted-foreground">{fmtDt(ev.created_at)}</Table.Cell>
									<Table.Cell>
										<div class="font-medium text-sm">{ev.nama}</div>
										<div class="font-mono text-xs text-muted-foreground">{ev.nis}</div>
									</Table.Cell>
									<Table.Cell>
										<Badge variant="outline" class="text-xs">{proctoringEventLabel(ev.event_type)}</Badge>
									</Table.Cell>
									<Table.Cell class="text-sm text-muted-foreground">{ev.room_name || '—'}</Table.Cell>
									<Table.Cell class="max-w-md break-words font-mono text-xs text-slate-500">{eventDataText(ev.event_data)}</Table.Cell>
								</Table.Row>
							{:else}
								<Table.Row>
									<Table.Cell colspan={5} class="py-8 text-center text-sm text-slate-400">Belum ada log aktivitas</Table.Cell>
								</Table.Row>
							{/each}
						</Table.Body>
					</Table.Root>
				</Card.Content>
			</Card.Root>

		<!-- Tab: Essay Grading -->
		{:else if activeTab === 'essay'}
			<Card.Root>
				<Card.Header class="pb-2">
					<Card.Title class="text-base">Koreksi Jawaban Uraian ({essays.length} belum dinilai)</Card.Title>
				</Card.Header>
				<Card.Content class="p-0 overflow-x-auto">
					<Table.Root>
						<Table.Header>
							<Table.Row class="bg-green-50">
								<Table.Head>Siswa</Table.Head>
								<Table.Head>Pertanyaan</Table.Head>
								<Table.Head>Jawaban Siswa</Table.Head>
								<Table.Head class="w-32">Nilai (0-100)</Table.Head>
								<Table.Head></Table.Head>
							</Table.Row>
						</Table.Header>
						<Table.Body>
							{#each essays as e (e.id)}
								<Table.Row>
									<Table.Cell>
										<div class="font-medium">{e.nama}</div>
										<div class="text-xs text-slate-500 font-mono">{e.nis}</div>
									</Table.Cell>
									<Table.Cell class="max-w-xs text-sm">{e.question_text}</Table.Cell>
									<Table.Cell class="max-w-sm">
										<div class="rounded bg-slate-50 p-2 text-sm border border-slate-200 whitespace-pre-wrap">{e.answer}</div>
									</Table.Cell>
									<Table.Cell>
										<Input type="number" min="0" max="100" bind:value={gradeInput[e.id]} placeholder="0-100" class="w-24 h-8" />
									</Table.Cell>
									<Table.Cell>
										<LoadingButton size="sm" onclick={() => submitGrade(e.id)} loading={gradeBusyId === e.id} disabled={gradeBusyId !== '' && gradeBusyId !== e.id} loadingLabel="Menyimpan...">Simpan</LoadingButton>
									</Table.Cell>
								</Table.Row>
							{:else}
								<Table.Row>
									<Table.Cell colspan={5} class="text-center text-slate-400 py-12">
										Tidak ada jawaban uraian yang perlu dikoreksi.
									</Table.Cell>
								</Table.Row>
							{/each}
						</Table.Body>
					</Table.Root>
				</Card.Content>
			</Card.Root>
		{/if}
		{/snippet}
	</AsyncContent>
</div>
