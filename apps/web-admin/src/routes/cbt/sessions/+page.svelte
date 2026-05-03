<script lang="ts">
	import { onMount } from 'svelte';
	import { SvelteMap } from 'svelte/reactivity';
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
	import { confirmChallenge } from '$lib/confirm-dialog';
	import { clientApiPath, clientApiPathWithQuery, readClientApiData, readClientJson } from '$lib/client/api';

	type ExamSession = {
		id: string; package_id: string; package_title: string;
		class_id: string; class_name: string; class_code: string;
		scope_type: string; scope_ref: string; mix_policy: string; assignment_mode: string;
		allow_cross_grade: boolean; is_special_event: boolean;
		title: string; scheduled_start: string; scheduled_end: string;
		status: string; participant_count: number; created_at: string;
	};
	type CbtPackage = {
		id: string; title: string; subject_code: string; subject_name: string;
		question_count: number; is_active: boolean;
	};
	type PackageQuestion = {
		package_id: string; question_id: string;
		question_type?: string; status?: string; cp_ref?: string; tp_ref?: string; kd_ref?: string;
		cognitive_level?: string; hots_flag?: boolean;
	};
	type PackageQualitySummary = {
		questions: PackageQuestion[];
		typeBuckets: { label: string; count: number }[];
		hotsCount: number;
		missingCount: number;
		unpublishedCount: number;
		totalCount: number;
	};
	type SchoolClass = { id: string; name: string; code: string; level: string; };
	type SessionsOverview = {
		sessions: ExamSession[];
		packages: CbtPackage[];
		packageQuestions: PackageQuestion[];
		classes: SchoolClass[];
	};
	type CbtPackagesPayload = {
		packages?: CbtPackage[];
		questions?: unknown[];
		error?: string;
		message?: string;
	};
	type AcademicPayload = {
		classes?: SchoolClass[];
		error?: string;
		message?: string;
	};

	let sessions = $state<ExamSession[]>([]);
	let packages = $state<CbtPackage[]>([]);
	let packageQuestions = $state<PackageQuestion[]>([]);
	let classes = $state<SchoolClass[]>([]);
	let sessionsPromise = $state<Promise<SessionsOverview> | null>(null);
	let showForm = $state(false);

	let fPackageId = $state('');
	let fScopeType = $state('class');
	let fClassId = $state('');
	let fGradeLevel = $state('VII');
	let fMixPolicy = $state('same_grade');
	let fAssignmentMode = $state('random_balanced');
	let fAllowCrossGrade = $state(false);
	let fIsSpecialEvent = $state(false);
	let fTitle = $state('');
	let fStart = $state('');
	let fEnd = $state('');
	let fBusy = $state(false);
	let statusBusyId = $state('');
	let deleteBusyId = $state('');
	let operationState = $state<{ tone: 'success' | 'error' | 'warning' | 'info'; title: string; message: string } | null>(null);
	let sessionsRequestId = 0;

	// Enroll modal
	let enrollSession = $state<ExamSession | null>(null);
	let enrollScopeType = $state('class');
	let enrollClassId = $state('');
	let enrollGradeLevel = $state('VII');
	let enrollBusy = $state(false);
	let selectedPackage = $derived(packages.find((pkg) => pkg.id === fPackageId) ?? null);
	let selectedPackageQuality = $derived(packageQualitySummary(fPackageId));
	let sessionReadinessIssues = $derived(buildSessionReadinessIssues());
	let canCreateSession = $derived(!fBusy && sessionReadinessIssues.length === 0);

	const statusLabel: Record<string, string> = {
		draft: 'Draft', scheduled: 'Terjadwal', active: 'Berlangsung',
		finished: 'Selesai', cancelled: 'Dibatalkan',
	};

	function statusClass(s: string) {
		if (s === 'active') return 'bg-emerald-100 text-emerald-700 border-emerald-200';
		if (s === 'finished') return 'bg-slate-100 text-slate-500 border-slate-200';
		if (s === 'cancelled') return 'bg-red-100 text-red-700 border-red-200';
		if (s === 'scheduled') return 'bg-green-100 text-green-800 border-green-200';
		return 'bg-amber-100 text-amber-700 border-amber-200';
	}

	function fmtDt(iso: string) {
		if (!iso) return '—';
		return new Date(iso).toLocaleString('id-ID', {
			timeZone: 'Asia/Makassar', year: 'numeric', month: 'short',
			day: 'numeric', hour: '2-digit', minute: '2-digit',
		}) + ' WITA';
	}

	function toRFC3339(localDt: string): string {
		if (!localDt) return '';
		return new Date(localDt).toISOString();
	}

	function scopeSummary(session: ExamSession) {
		if (session.scope_type === 'grade') return `Tingkat ${session.scope_ref || '—'}`;
		if (session.scope_type === 'school') return 'Seluruh sekolah';
		if (session.scope_type === 'custom') return session.scope_ref || 'Peserta khusus';
		return session.class_code || session.class_name || 'Per kelas';
	}

	function mixPolicyLabel(value: string) {
		if (value === 'same_class') return 'Tetap per kelas';
		if (value === 'mixed_scope') return 'Campur lintas cakupan';
		return 'Campur dalam tingkat';
	}

	function adaptiveMixPolicy(scopeType: string) {
		if (scopeType === 'class') return 'same_class';
		if (scopeType === 'grade') return 'same_grade';
		return 'mixed_scope';
	}

	function sessionActionLabel(status: string) {
		if (status === 'draft') return 'Draft';
		if (status === 'scheduled') return 'Terjadwal';
		if (status === 'active') return 'Berlangsung';
		if (status === 'finished') return 'Selesai';
		return 'Dibatalkan';
	}

	function isRecord(value: unknown): value is Record<string, unknown> {
		return typeof value === 'object' && value !== null;
	}

	function parsePackageQuestions(payload: CbtPackagesPayload | unknown) {
		return isRecord(payload) && Array.isArray(payload.questions) ? (payload.questions as PackageQuestion[]) : [];
	}

	function compactValue(value: string | number | null | undefined, fallback: string) {
		const text = value === null || value === undefined ? '' : String(value).trim();
		return text || fallback;
	}

	function questionTypeLabel(value: string | null | undefined) {
		const labels: Record<string, string> = {
			multiple_choice: 'PG',
			multiple_answer: 'PG Kompleks',
			true_false: 'Benar/Salah',
			agree_disagree: 'Setuju/Tidak',
			matching: 'Menjodohkan',
			short_answer: 'Isian',
			essay: 'Essay',
		};
		const normalized = compactValue(value, '');
		return labels[normalized] ?? (normalized ? normalized.replaceAll('_', ' ') : 'Belum tipe');
	}

	function packageQuestionHasBlueprintGap(question: PackageQuestion) {
		return !compactValue(question.cp_ref, '')
			|| (!compactValue(question.tp_ref, '') && !compactValue(question.kd_ref, ''))
			|| !compactValue(question.cognitive_level, '');
	}

	function countByLabel<T>(items: T[], selector: (item: T) => string) {
		const counts = new SvelteMap<string, number>();
		for (const item of items) {
			const label = selector(item);
			counts.set(label, (counts.get(label) ?? 0) + 1);
		}
		return Array.from(counts.entries())
			.map(([label, count]) => ({ label, count }))
			.sort((a, b) => b.count - a.count || a.label.localeCompare(b.label));
	}

	function packageQualitySummary(packageID: string): PackageQualitySummary {
		const pkg = packages.find((item) => item.id === packageID);
		const questions = packageQuestions.filter((question) => question.package_id === packageID);
		return {
			questions,
			typeBuckets: countByLabel(questions, (question) => questionTypeLabel(question.question_type)),
			hotsCount: questions.filter((question) => question.hots_flag).length,
			missingCount: questions.filter(packageQuestionHasBlueprintGap).length,
			unpublishedCount: questions.filter((question) => question.status !== 'published').length,
			totalCount: pkg?.question_count ?? questions.length,
		};
	}

	function packageQualityIssues(packageID: string) {
		const issues: string[] = [];
		const pkg = packages.find((item) => item.id === packageID);
		const quality = packageQualitySummary(packageID);
		if (pkg && !pkg.is_active) issues.push('paket nonaktif');
		if (packageID && quality.totalCount === 0) issues.push('paket kosong');
		if (packageID && quality.questions.length > 0 && quality.unpublishedCount > 0) {
			issues.push(`${quality.unpublishedCount} belum terbit`);
		}
		return issues;
	}

	function buildSessionReadinessIssues() {
		const issues: string[] = [];
		if (!fPackageId) issues.push('Pilih paket soal');
		for (const issue of packageQualityIssues(fPackageId)) {
			if (issue === 'paket nonaktif') issues.push('Paket soal tidak aktif');
			else if (issue === 'paket kosong') issues.push('Paket belum memiliki soal');
			else issues.push(`${issue} di paket soal`);
		}
		if (!fTitle.trim()) issues.push('Isi nama sesi');
		if (!fStart) issues.push('Isi jadwal mulai');
		if (!fEnd) issues.push('Isi jadwal selesai');
		if (fStart && fEnd && new Date(fEnd) <= new Date(fStart)) issues.push('Jadwal selesai harus setelah mulai');
		if (fScopeType === 'class' && !fClassId) issues.push('Pilih kelas peserta');
		if (fScopeType === 'grade' && !fGradeLevel) issues.push('Pilih tingkat peserta');
		if (fAllowCrossGrade && !fIsSpecialEvent) issues.push('Lintas tingkat hanya boleh untuk sesi khusus');
		return issues;
	}

	async function fetchOverview(): Promise<SessionsOverview> {
		const [nextSessions, packagePayload, academicPayload] = await Promise.all([
			fetch('/api/cbt/sessions').then((response) => readClientApiData<ExamSession[]>(response, 'Gagal memuat data sesi')),
			fetch('/api/cbt/packages').then((response) => readClientApiData<CbtPackagesPayload>(response, 'Gagal memuat paket ujian')),
			fetch('/api/academic').then((response) => readClientApiData<AcademicPayload>(response, 'Gagal memuat data akademik')),
		]);
		return {
			sessions: Array.isArray(nextSessions) ? nextSessions : [],
			packages: packagePayload.packages ?? [],
			packageQuestions: parsePackageQuestions(packagePayload),
			classes: academicPayload.classes ?? [],
		};
	}

	function applyOverview(overview: SessionsOverview) {
		sessions = overview.sessions;
		packages = overview.packages;
		packageQuestions = overview.packageQuestions;
		classes = overview.classes;
	}

	function loadInitial() {
		const requestId = ++sessionsRequestId;
		sessions = [];
		packages = [];
		packageQuestions = [];
		classes = [];
		sessionsPromise = fetchOverview().then((overview) => {
			if (requestId !== sessionsRequestId) return { sessions, packages, packageQuestions, classes };
			applyOverview(overview);
			return overview;
		}).catch((error: unknown) => {
			if (requestId === sessionsRequestId) throw error;
			return { sessions, packages, packageQuestions, classes };
		});
	}

	async function refreshSessions() {
		if (!sessionsPromise) {
			loadInitial();
			return;
		}
		const requestId = ++sessionsRequestId;
		try {
			const overview = await fetchOverview();
			if (requestId !== sessionsRequestId) return;
			applyOverview(overview);
			sessionsPromise = Promise.resolve(overview);
		} catch (error) {
			if (requestId === sessionsRequestId) {
				sessionsPromise = Promise.resolve({ sessions, packages, packageQuestions, classes });
				toast.error(sessionsErrorMessage(error));
			}
		}
	}

	function retrySessions(reset?: () => void) {
		reset?.();
		loadInitial();
	}

	function sessionsErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Gagal memuat data sesi';
	}

	function handleSessionsRenderError(error: unknown) {
		console.error('CBT sessions render failed', error);
	}

	function showToast(msg: string, ok = true) {
		if (ok) toast.success(msg);
		else toast.error(msg);
	}

	function mutationErrorMessage(error: unknown, fallback: string) {
		if (error instanceof Error && error.message.trim() && !error.message.toLowerCase().includes('fetch')) return error.message;
		return fallback;
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

	function sessionLegacyMutationPath(id: string) {
		return clientApiPathWithQuery('/api/cbt/sessions', new URLSearchParams({ id }));
	}

	async function createSession() {
		if (sessionReadinessIssues.length > 0) {
			setOperationState('warning', 'Sesi Belum Siap', sessionReadinessIssues[0] ?? 'Lengkapi sesi ujian terlebih dahulu.');
			return;
		}
		fBusy = true;
		try {
			const scopeRef = fScopeType === 'class' ? fClassId : fScopeType === 'grade' ? fGradeLevel : '';
			const res = await fetch('/api/cbt/sessions', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					package_id: fPackageId,
					class_id: fScopeType === 'class' ? fClassId : '',
					scope_type: fScopeType,
					scope_ref: scopeRef,
					mix_policy: adaptiveMixPolicy(fScopeType),
					assignment_mode: fAssignmentMode,
					allow_cross_grade: fAllowCrossGrade,
					is_special_event: fIsSpecialEvent,
					title: fTitle,
					scheduled_start: toRFC3339(fStart), scheduled_end: toRFC3339(fEnd),
					status: 'draft',
				}),
			});
			await readClientJson<unknown>(res);
			fPackageId = ''; fScopeType = 'class'; fClassId = ''; fGradeLevel = 'VII';
			fMixPolicy = 'same_class'; fAssignmentMode = 'random_balanced';
			fAllowCrossGrade = false; fIsSpecialEvent = false;
			fTitle = ''; fStart = ''; fEnd = '';
			showForm = false;
			setOperationState('success', 'Sesi Tersimpan Sebagai Draft', 'Sesi baru sudah dibuat. Daftarkan peserta dan cek ruang sebelum menjadwalkan atau memulai sesi.');
			showToast('Sesi ujian berhasil dibuat');
			await refreshSessions();
		} catch (error) {
			showToast(mutationErrorMessage(error, 'Gagal membuat sesi ujian. Periksa koneksi lalu coba lagi.'), false);
		} finally { fBusy = false; }
	}

	async function deleteSession(id: string, title: string) {
		if (!(await confirmPhrase('Hapus Sesi Ujian', `Sesi "${title}" hanya boleh dihapus jika masih Draft. Penghapusan akan membuang konfigurasi sesi dari daftar operator.`, 'HAPUS'))) return;
		deleteBusyId = id;
		try {
			const res = await fetch(sessionLegacyMutationPath(id), { method: 'DELETE' });
			await readClientJson<unknown>(res);
			setOperationState('warning', 'Sesi Dihapus', `Sesi "${title}" sudah dihapus dari daftar sesi ujian.`);
			showToast('Sesi dihapus');
			await refreshSessions();
		} catch (error) {
			setOperationState('error', 'Sesi Gagal Dihapus', 'Hanya sesi berstatus Draft yang dapat dihapus. Ubah alur kerja sesi atau periksa statusnya terlebih dahulu.');
			showToast(mutationErrorMessage(error, 'Gagal menghapus sesi ujian. Periksa koneksi lalu coba lagi.'), false);
		} finally {
			deleteBusyId = '';
		}
	}

	async function updateStatus(id: string, status: string) {
		const challenge = status === 'cancelled' ? 'BATALKAN' : status === 'finished' ? 'SELESAI' : '';
		if (challenge && !(await confirmPhrase('Konfirmasi Perubahan Status', `Perubahan ini akan mengubah status sesi menjadi "${statusLabel[status] ?? status}" dan memengaruhi operasi ujian berikutnya.`, challenge))) {
			return;
		}
		statusBusyId = id;
		try {
			const res = await fetch(clientApiPath`/api/cbt/sessions/${id}/status`, {
				method: 'PATCH',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ status }),
			});
			await readClientJson<unknown>(res);
			setOperationState('success', 'Status Sesi Diperbarui', `Sesi sekarang berstatus "${statusLabel[status] ?? status}". Pastikan langkah operator berikutnya sudah sesuai.`);
			showToast('Status diperbarui');
			await refreshSessions();
		} catch (error) {
			const message = mutationErrorMessage(error, 'Gagal mengubah status sesi. Periksa kesiapan ruang, peserta, kursi, dan pengawas lalu coba lagi.');
			setOperationState('error', 'Status Gagal Diperbarui', message);
			showToast(message, false);
		} finally {
			statusBusyId = '';
		}
	}

	async function enrollParticipants() {
		if (!enrollSession) return;
		if (enrollScopeType === 'class' && !enrollClassId) return;
		if (enrollScopeType === 'grade' && !enrollGradeLevel) return;
		enrollBusy = true;
		try {
			const payload =
				enrollScopeType === 'class'
					? { scope_type: 'class', class_id: enrollClassId }
					: enrollScopeType === 'grade'
						? { scope_type: 'grade', level: enrollGradeLevel }
						: { scope_type: 'school' };
			const res = await fetch(clientApiPath`/api/cbt/sessions/${enrollSession.id}/enroll`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(payload),
			});
			await readClientJson<unknown>(res);
			setOperationState('success', 'Peserta Berhasil Didaftarkan', `Kelompok peserta untuk sesi "${enrollSession.title}" sudah masuk. Lanjutkan ke pengaturan ruangan jika diperlukan.`);
			showToast(`Peserta berhasil didaftarkan ke sesi "${enrollSession.title}"`);
			enrollSession = null;
			enrollScopeType = 'class';
			enrollClassId = '';
			enrollGradeLevel = 'VII';
			await refreshSessions();
		} catch (error) {
			showToast(mutationErrorMessage(error, 'Gagal mendaftarkan siswa. Periksa koneksi lalu coba lagi.'), false);
		} finally { enrollBusy = false; }
	}

	onMount(() => {
		void loadInitial();
	});
</script>

<svelte:head><title>Sesi Ujian CBT — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6">
	<div class="flex flex-wrap items-start justify-between gap-4">
		<div>
			<h1 class="text-2xl font-semibold text-slate-800">Sesi Ujian CBT</h1>
			<p class="text-sm text-slate-500 mt-1">Jadwalkan sesi per kelas, tingkat, atau seluruh sekolah dengan rooming yang fleksibel</p>
		</div>
		<Button onclick={() => (showForm = !showForm)}>
			{showForm ? 'Batal' : '+ Buat Sesi'}
		</Button>
	</div>

	{#if operationState}
		<OperationStatusPanel {...operationState} />
	{/if}

	{#if showForm}
		<Card.Root>
			<Card.Header class="pb-2">
				<Card.Title class="text-base">Buat Sesi Ujian Baru</Card.Title>
			</Card.Header>
			<Card.Content class="space-y-4">
				<div class="grid gap-3 sm:grid-cols-2">
					<div>
						<label for="session-package" class="text-xs text-slate-500 mb-1 block">Paket Soal <span class="text-red-500">*</span></label>
						<select id="session-package" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fPackageId}>
							<option value="">-- Pilih Paket --</option>
							{#each packages as p (p.id)}
								<option value={p.id}>{p.title} ({p.subject_code}){p.is_active ? '' : ' - nonaktif'}</option>
							{/each}
						</select>
					</div>
					{#if fPackageId}
						{@const quality = selectedPackageQuality}
						<div class="sm:col-span-2 rounded-lg border border-slate-200 bg-slate-50 px-3 py-2">
							<div class="flex flex-wrap items-start justify-between gap-3">
								<div>
									<p class="text-xs font-semibold uppercase tracking-[0.16em] text-slate-700">Quality Gate Paket</p>
									<p class="mt-1 text-sm font-medium text-slate-900">{selectedPackage?.title ?? 'Paket dipilih'}</p>
								</div>
								<div class="flex flex-wrap gap-1.5">
									<Badge variant="outline" class="bg-white text-xs">{quality.totalCount} soal</Badge>
									{#each quality.typeBuckets.slice(0, 3) as bucket (bucket.label)}
										<Badge variant="outline" class="bg-white text-xs">{bucket.label}: {bucket.count}</Badge>
									{/each}
									{#if quality.hotsCount > 0}
										<Badge class="border-amber-200 bg-amber-50 text-amber-700 text-xs">{quality.hotsCount} HOTS</Badge>
									{/if}
									{#if quality.missingCount > 0}
										<Badge class="border-amber-200 bg-amber-50 text-amber-700 text-xs">{quality.missingCount} metadata kurang</Badge>
									{/if}
									{#if quality.unpublishedCount > 0}
										<Badge class="border-red-200 bg-red-50 text-red-700 text-xs">{quality.unpublishedCount} belum terbit</Badge>
									{/if}
								</div>
							</div>
							{#if selectedPackage && !selectedPackage.is_active}
								<p class="mt-2 text-xs font-medium text-red-700">Paket nonaktif tidak boleh dijadikan sesi ujian.</p>
							{:else if quality.totalCount === 0}
								<p class="mt-2 text-xs font-medium text-red-700">Paket ini belum memiliki soal, sehingga sesi tidak bisa dibuat.</p>
							{:else if quality.unpublishedCount > 0}
								<p class="mt-2 text-xs font-medium text-red-700">Rapikan paket dulu. Flutter hanya menyajikan soal terbit, jadi soal belum terbit akan membuat jumlah soal sesi tidak konsisten.</p>
							{:else if quality.missingCount > 0}
								<p class="mt-2 text-xs font-medium text-amber-700">Sesi masih boleh dibuat, tetapi {quality.missingCount} soal belum lengkap CP/TP/KD atau level kognitif.</p>
							{:else}
								<p class="mt-2 text-xs font-medium text-emerald-700">Paket siap dipakai untuk draft sesi CBT.</p>
							{/if}
						</div>
					{/if}
					<div>
						<label for="session-scope" class="text-xs text-slate-500 mb-1 block">Cakupan peserta <span class="text-red-500">*</span></label>
						<select id="session-scope" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fScopeType}>
							<option value="class">Per kelas</option>
							<option value="grade">Per tingkat</option>
							<option value="school">Seluruh sekolah</option>
						</select>
					</div>
					{#if fScopeType === 'class'}
						<div>
							<label for="session-class" class="text-xs text-slate-500 mb-1 block">Kelas <span class="text-red-500">*</span></label>
							<select id="session-class" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fClassId}>
								<option value="">-- Pilih Kelas --</option>
								{#each classes as c (c.id)}
									<option value={c.id}>{c.code} — {c.name}</option>
								{/each}
							</select>
						</div>
					{:else if fScopeType === 'grade'}
						<div>
							<label for="session-grade" class="text-xs text-slate-500 mb-1 block">Tingkat <span class="text-red-500">*</span></label>
							<select id="session-grade" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fGradeLevel}>
								<option value="VII">VII</option>
								<option value="VIII">VIII</option>
								<option value="IX">IX</option>
							</select>
						</div>
					{:else}
						<div class="rounded-md border border-emerald-200 bg-emerald-50 px-3 py-2 text-sm text-emerald-900">
							Semua siswa aktif di sekolah dapat menjadi peserta sesi ini.
						</div>
					{/if}
					<div>
						<label for="session-mix-policy" class="text-xs text-slate-500 mb-1 block">Mix policy</label>
						<select id="session-mix-policy" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fMixPolicy}>
							<option value="same_class">Tetap per kelas</option>
							<option value="same_grade">Campur dalam tingkat</option>
							<option value="mixed_scope">Campur lintas cakupan</option>
						</select>
					</div>
					<div>
						<label for="session-assignment-mode" class="text-xs text-slate-500 mb-1 block">Mode alokasi ruangan</label>
						<select id="session-assignment-mode" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fAssignmentMode}>
							<option value="random_balanced">Acak seimbang</option>
							<option value="manual">Manual</option>
							<option value="random_by_gender">Acak per gender</option>
							<option value="random_by_accommodation">Acak akomodasi khusus</option>
						</select>
					</div>
					<div class="sm:col-span-2">
						<label for="session-title" class="text-xs text-slate-500 mb-1 block">Nama Sesi <span class="text-red-500">*</span></label>
						<Input id="session-title" placeholder="mis: UTS Matematika VII A - Semester 1 2025" bind:value={fTitle} />
					</div>
					<div>
						<label for="session-start" class="text-xs text-slate-500 mb-1 block">Mulai <span class="text-red-500">*</span></label>
						<Input id="session-start" type="datetime-local" bind:value={fStart} />
					</div>
					<div>
						<label for="session-end" class="text-xs text-slate-500 mb-1 block">Selesai <span class="text-red-500">*</span></label>
						<Input id="session-end" type="datetime-local" bind:value={fEnd} />
					</div>
					<div class="sm:col-span-2 grid gap-3 sm:grid-cols-2">
						<label class="flex items-center gap-2 rounded-md border border-input px-3 py-2 text-sm text-slate-700">
							<input type="checkbox" bind:checked={fIsSpecialEvent} class="size-4 accent-emerald-700" />
							Tandai sebagai sesi khusus
						</label>
						<label class="flex items-center gap-2 rounded-md border border-input px-3 py-2 text-sm text-slate-700">
							<input type="checkbox" bind:checked={fAllowCrossGrade} class="size-4 accent-emerald-700" />
							Izinkan lintas tingkat
						</label>
					</div>
				</div>
				<div class="flex gap-2">
					<LoadingButton
						disabled={!canCreateSession}
						onclick={() => void createSession()}
						loading={fBusy}
						loadingLabel="Menyimpan..."
					>
						Buat Sesi
					</LoadingButton>
					<Button variant="outline" onclick={() => (showForm = false)}>Batal</Button>
				</div>
				{#if sessionReadinessIssues.length > 0}
					<div class="rounded-md border border-amber-200 bg-amber-50 px-3 py-2 text-xs text-amber-900">
						<span class="font-semibold">Belum siap dibuat:</span>
						<span>{sessionReadinessIssues.join(', ')}</span>
					</div>
				{/if}
			</Card.Content>
		</Card.Root>
	{/if}

	<!-- Enroll modal -->
	{#if enrollSession}
		<Card.Root class="border-green-200 bg-green-50">
			<Card.Header class="pb-2">
				<Card.Title class="text-base text-green-900">Daftarkan Siswa ke Sesi</Card.Title>
				<p class="text-sm text-green-700 mt-0.5">{enrollSession.title}</p>
			</Card.Header>
			<Card.Content class="space-y-3">
				<p class="text-sm text-slate-600">Tentukan kelompok peserta untuk sesi ini. Ruangan tetap bisa diacak terpisah setelah peserta terdaftar.</p>
				<div>
					<label for="enroll-scope" class="text-xs text-slate-500 mb-1 block">Cakupan peserta</label>
					<select id="enroll-scope" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={enrollScopeType}>
						<option value="class">Per kelas</option>
						<option value="grade">Per tingkat</option>
						<option value="school">Seluruh sekolah</option>
					</select>
				</div>
				{#if enrollScopeType === 'class'}
					<div>
						<label for="enroll-class" class="text-xs text-slate-500 mb-1 block">Pilih Kelas</label>
						<select id="enroll-class" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={enrollClassId}>
							<option value="">-- Pilih Kelas --</option>
							{#each classes as c (c.id)}
								<option value={c.id}>{c.code} — {c.name}</option>
							{/each}
						</select>
					</div>
				{:else if enrollScopeType === 'grade'}
					<div>
						<label for="enroll-grade" class="text-xs text-slate-500 mb-1 block">Pilih Tingkat</label>
						<select id="enroll-grade" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={enrollGradeLevel}>
							<option value="VII">VII</option>
							<option value="VIII">VIII</option>
							<option value="IX">IX</option>
						</select>
					</div>
				{:else}
					<div class="rounded-md border border-emerald-200 bg-white px-3 py-2 text-sm text-slate-700">
						Semua siswa aktif di sekolah akan didaftarkan ke sesi ini.
					</div>
				{/if}
				<div class="flex gap-2">
					<LoadingButton
						disabled={enrollBusy || (enrollScopeType === 'class' && !enrollClassId) || (enrollScopeType === 'grade' && !enrollGradeLevel)}
						onclick={() => void enrollParticipants()}
						loading={enrollBusy}
						loadingLabel="Mendaftarkan..."
					>
						Daftarkan Siswa
					</LoadingButton>
					<Button variant="outline" onclick={() => { enrollSession = null; enrollScopeType = 'class'; enrollClassId = ''; enrollGradeLevel = 'VII'; }}>Batal</Button>
				</div>
			</Card.Content>
		</Card.Root>
	{/if}

	<AsyncContent promise={sessionsPromise} onerror={handleSessionsRenderError}>
		{#snippet pending()}
			<div class="space-y-4">
				<div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
					{#each Array.from({ length: 4 }) as _, index (`cbt-session-stat-skeleton-${index}`)}
						<Card.Root class="border-slate-200">
							<Card.Content class="space-y-2 p-4">
								<Skeleton class="h-4 w-24" />
								<Skeleton class="h-7 w-16" />
							</Card.Content>
						</Card.Root>
					{/each}
				</div>
				<Card.Root class="overflow-hidden border-slate-200 shadow-sm">
					<Card.Content class="space-y-3 p-6">
						{#each Array.from({ length: 5 }) as _, index (`cbt-session-row-skeleton-${index}`)}
							<div class="grid gap-3 lg:grid-cols-[1.2fr_1fr_0.9fr_1fr_0.5fr_0.7fr_auto] lg:items-center">
								<Skeleton class="h-5 w-40" />
								<Skeleton class="h-5 w-32" />
								<Skeleton class="h-5 w-28" />
								<Skeleton class="h-5 w-36" />
								<Skeleton class="h-5 w-12" />
								<Skeleton class="h-6 w-20" />
								<Skeleton class="h-9 w-36 justify-self-end" />
							</div>
						{/each}
					</Card.Content>
				</Card.Root>
			</div>
		{/snippet}

		{#snippet failed(error, reset)}
			<RecoveryPanel
				title="Sesi Ujian Belum Tersaji"
				message={sessionsErrorMessage(error)}
				onRetry={() => retrySessions(reset)}
			/>
		{/snippet}

		{#snippet children(value)}
			{@const overview = value as SessionsOverview}
			{@const currentSessions = overview.sessions}
		<Card.Root class="overflow-hidden border-slate-200 shadow-sm">
			<Card.Header class="pb-2">
				<Card.Title class="text-base">Daftar Sesi ({currentSessions.length})</Card.Title>
			</Card.Header>
			<Card.Content class="p-0">
				<div class="hidden overflow-x-auto lg:block">
				<Table.Root>
					<Table.Header>
						<Table.Row>
							<Table.Head>Nama Sesi</Table.Head>
							<Table.Head>Paket</Table.Head>
							<Table.Head>Cakupan</Table.Head>
							<Table.Head>Jadwal Mulai</Table.Head>
							<Table.Head>Peserta</Table.Head>
							<Table.Head>Status</Table.Head>
							<Table.Head></Table.Head>
						</Table.Row>
					</Table.Header>
					<Table.Body>
						{#each currentSessions as s (s.id)}
							{@const rowPackageQuality = packageQualitySummary(s.package_id)}
							{@const rowPackageIssues = packageQualityIssues(s.package_id)}
							<Table.Row>
								<Table.Cell class="font-medium max-w-48">
									<p class="truncate">{s.title}</p>
								</Table.Cell>
								<Table.Cell class="max-w-44">
									<p class="truncate text-sm text-slate-600">{s.package_title}</p>
									<div class="mt-1 flex flex-wrap gap-1">
										<Badge variant="outline" class="bg-white text-[11px]">{rowPackageQuality.totalCount} soal</Badge>
										{#if rowPackageIssues.length > 0}
											<Badge class="border-red-200 bg-red-50 text-[11px] text-red-700">{rowPackageIssues.join(', ')}</Badge>
										{:else if rowPackageQuality.missingCount > 0}
											<Badge class="border-amber-200 bg-amber-50 text-[11px] text-amber-700">{rowPackageQuality.missingCount} metadata kurang</Badge>
										{:else}
											<Badge class="border-emerald-200 bg-emerald-50 text-[11px] text-emerald-700">Paket siap</Badge>
										{/if}
									</div>
								</Table.Cell>
								<Table.Cell>
									<div class="space-y-1">
										<Badge variant="outline" class="text-xs">{scopeSummary(s)}</Badge>
										<p class="text-[11px] text-slate-500">{mixPolicyLabel(s.mix_policy)}</p>
									</div>
								</Table.Cell>
								<Table.Cell class="text-slate-500 text-xs whitespace-nowrap">{fmtDt(s.scheduled_start)}</Table.Cell>
								<Table.Cell>
									<span class="font-mono text-sm">{s.participant_count}</span>
								</Table.Cell>
								<Table.Cell>
									<Badge class={statusClass(s.status)}>{statusLabel[s.status] ?? s.status}</Badge>
								</Table.Cell>
								<Table.Cell>
									<div class="flex gap-1 flex-wrap">
										{#if s.status === 'draft'}
											<Button
												size="xs"
												variant="outline"
												onclick={() => {
													enrollSession = s;
													enrollScopeType = s.scope_type || 'class';
													enrollClassId = s.class_id;
													enrollGradeLevel = s.scope_type === 'grade' ? s.scope_ref : 'VII';
												}}
											>
												Daftarkan Siswa
											</Button>
											<LoadingButton size="xs" onclick={() => updateStatus(s.id, 'scheduled')} loading={statusBusyId === s.id} disabled={(statusBusyId !== '' && statusBusyId !== s.id) || rowPackageIssues.length > 0} loadingLabel="Memproses...">
												Jadwalkan
											</LoadingButton>
											<LoadingButton size="xs" variant="destructive" onclick={() => deleteSession(s.id, s.title)} loading={deleteBusyId === s.id} disabled={deleteBusyId !== '' && deleteBusyId !== s.id} loadingLabel="Menghapus...">
												Hapus
											</LoadingButton>
										{:else if s.status === 'scheduled'}
											<LoadingButton size="xs" onclick={() => updateStatus(s.id, 'active')} loading={statusBusyId === s.id} disabled={(statusBusyId !== '' && statusBusyId !== s.id) || rowPackageIssues.length > 0} loadingLabel="Memproses...">Mulai</LoadingButton>
											<LoadingButton size="xs" variant="outline" onclick={() => updateStatus(s.id, 'cancelled')} loading={statusBusyId === s.id} disabled={statusBusyId !== '' && statusBusyId !== s.id} loadingLabel="Memproses...">Batalkan</LoadingButton>
										{:else if s.status === 'active'}
											<LoadingButton size="xs" onclick={() => updateStatus(s.id, 'finished')} loading={statusBusyId === s.id} disabled={statusBusyId !== '' && statusBusyId !== s.id} loadingLabel="Memproses...">Selesaikan</LoadingButton>
										{/if}
										{#if s.status === 'finished' || s.status === 'active'}
											<a href={resolve(`/cbt/sessions/${s.id}`)} class="inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium border border-input bg-background hover:bg-muted text-slate-700 transition-colors">
												Lihat Hasil
											</a>
										{/if}
									</div>
								</Table.Cell>
							</Table.Row>
						{:else}
							<Table.Row>
								<Table.Cell colspan={7} class="text-center text-slate-400 py-8">Belum ada sesi ujian</Table.Cell>
							</Table.Row>
						{/each}
					</Table.Body>
				</Table.Root>
				</div>

				<div class="grid gap-3 p-4 lg:hidden">
					{#each currentSessions as s (s.id)}
						{@const rowPackageQuality = packageQualitySummary(s.package_id)}
						{@const rowPackageIssues = packageQualityIssues(s.package_id)}
						<div class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
							<div class="flex items-start justify-between gap-3">
								<div class="min-w-0">
									<p class="text-sm font-semibold text-slate-900">{s.title}</p>
									<p class="mt-1 text-xs text-slate-500">{s.package_title}</p>
								</div>
								<Badge class={statusClass(s.status)}>{sessionActionLabel(s.status)}</Badge>
							</div>
							<div class="mt-3 flex flex-wrap items-center gap-2">
								<Badge variant="outline" class="text-xs">{scopeSummary(s)}</Badge>
								<Badge variant="outline" class="text-xs">{mixPolicyLabel(s.mix_policy)}</Badge>
								<Badge variant="secondary">{s.participant_count} peserta</Badge>
							</div>
							<div class="mt-2 flex flex-wrap items-center gap-1.5">
								<Badge variant="outline" class="bg-white text-xs">{rowPackageQuality.totalCount} soal</Badge>
								{#if rowPackageIssues.length > 0}
									<Badge class="border-red-200 bg-red-50 text-red-700 text-xs">{rowPackageIssues.join(', ')}</Badge>
								{:else if rowPackageQuality.missingCount > 0}
									<Badge class="border-amber-200 bg-amber-50 text-amber-700 text-xs">{rowPackageQuality.missingCount} metadata kurang</Badge>
								{:else}
									<Badge class="border-emerald-200 bg-emerald-50 text-emerald-700 text-xs">Paket siap</Badge>
								{/if}
							</div>
							<p class="mt-3 text-xs text-slate-500">{fmtDt(s.scheduled_start)}</p>
							<div class="mt-4 flex flex-wrap gap-2">
								{#if s.status === 'draft'}
									<Button
										size="sm"
										variant="outline"
										onclick={() => {
											enrollSession = s;
											enrollScopeType = s.scope_type || 'class';
											enrollClassId = s.class_id;
											enrollGradeLevel = s.scope_type === 'grade' ? s.scope_ref : 'VII';
										}}
									>
										Daftarkan
									</Button>
									<LoadingButton size="sm" onclick={() => updateStatus(s.id, 'scheduled')} loading={statusBusyId === s.id} disabled={(statusBusyId !== '' && statusBusyId !== s.id) || rowPackageIssues.length > 0} loadingLabel="Memproses...">Jadwalkan</LoadingButton>
									<LoadingButton size="sm" variant="destructive" onclick={() => deleteSession(s.id, s.title)} loading={deleteBusyId === s.id} disabled={deleteBusyId !== '' && deleteBusyId !== s.id} loadingLabel="Menghapus...">Hapus</LoadingButton>
								{:else if s.status === 'scheduled'}
									<LoadingButton size="sm" onclick={() => updateStatus(s.id, 'active')} loading={statusBusyId === s.id} disabled={(statusBusyId !== '' && statusBusyId !== s.id) || rowPackageIssues.length > 0} loadingLabel="Memproses...">Mulai</LoadingButton>
									<LoadingButton size="sm" variant="outline" onclick={() => updateStatus(s.id, 'cancelled')} loading={statusBusyId === s.id} disabled={statusBusyId !== '' && statusBusyId !== s.id} loadingLabel="Memproses...">Batalkan</LoadingButton>
								{:else if s.status === 'active'}
									<LoadingButton size="sm" onclick={() => updateStatus(s.id, 'finished')} loading={statusBusyId === s.id} disabled={statusBusyId !== '' && statusBusyId !== s.id} loadingLabel="Memproses...">Selesaikan</LoadingButton>
								{/if}
								{#if s.status === 'finished' || s.status === 'active'}
									<a href={resolve(`/cbt/sessions/${s.id}`)} class="inline-flex items-center rounded-md px-3 py-1.5 text-sm font-medium border border-input bg-background hover:bg-muted text-slate-700 transition-colors">
										Lihat Hasil
									</a>
								{/if}
							</div>
						</div>
					{:else}
						<div class="rounded-2xl border border-dashed border-slate-300 bg-slate-50 px-4 py-10 text-center text-sm text-slate-500">
							Belum ada sesi ujian
						</div>
					{/each}
				</div>
			</Card.Content>
		</Card.Root>
		{/snippet}
	</AsyncContent>
</div>
