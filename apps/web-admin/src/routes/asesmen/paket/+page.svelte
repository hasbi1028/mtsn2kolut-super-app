<script lang="ts">
	import { onMount } from 'svelte';
	import { SvelteMap, SvelteSet } from 'svelte/reactivity';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Input } from '$lib/components/ui/input';
	import { Textarea } from '$lib/components/ui/textarea';
	import { Badge } from '$lib/components/ui/badge';
	import { toast } from '$lib/components/ui/sonner';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import OperationStatusPanel from '$lib/components/OperationStatusPanel.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { confirmChallenge } from '$lib/confirm-dialog';
	import { clientApiPath, clientApiPathWithQuery, readClientApiData, readClientJson } from '$lib/client/api';

	type CbtPackage = {
		id: string; subject_id: string; subject_name: string; subject_code: string;
		title: string; description: string; duration_minutes: number;
		randomize_questions: boolean; is_active: boolean;
		question_count: number; created_at: string;
		event_id?: string | null;
	};
	type Question = {
		id: string; subject_id: string; subject_code: string;
		code: string; question_text: string; difficulty: string; status: string;
		event_id?: string | null;
		workflow_status?: string; question_type?: string; cp_ref?: string; tp_ref?: string; kd_ref?: string;
		material_topic?: string; cognitive_level?: string; hots_flag?: boolean;
	};
	type PackageQuestion = {
		package_id: string; question_id: string; position: number; points: number;
		question_code: string; question_text: string;
		question_type?: string; difficulty?: string; status?: string; workflow_status?: string;
		cp_ref?: string; tp_ref?: string; kd_ref?: string; material_topic?: string; cognitive_level?: string;
		hots_flag?: boolean;
	};
	type Subject = { id: string; name: string; code: string; };
	type EventContext = { id: string; title: string; status: string; target_levels?: string[]; academic_year_name?: string; };
	type PackagesOverview = {
		packages: CbtPackage[];
		packageQuestions: PackageQuestion[];
		allQuestions: Question[];
		subjects: Subject[];
	};
	type CbtPackagesPayload = {
		packages?: CbtPackage[];
		questions?: unknown[];
		error?: string;
		message?: string;
	};
	type AcademicPayload = {
		subjects?: Subject[];
		error?: string;
		message?: string;
	};
	type QuestionListPayload = {
		items?: Question[];
		meta?: {
			total?: number;
			limit?: number;
			offset?: number;
		};
	};
	type BlueprintBucket = {
		label: string;
		count: number;
	};
	type PackageQualitySummary = {
		questions: PackageQuestion[];
		typeBuckets: BlueprintBucket[];
		cognitiveBuckets: BlueprintBucket[];
		hotsCount: number;
		missingCount: number;
		unpublishedCount: number;
		totalPoints: number;
	};
	type BlueprintMatrixRow = {
		key: string;
		cp: string;
		tp: string;
		kd: string;
		topic: string;
		cognitive: string;
		count: number;
		hotsCount: number;
		types: string[];
		missing: boolean;
	};

	const questionPageSize = 100;
	const maxQuestionPages = 20;

	let packages = $state<CbtPackage[]>([]);
	let packageQuestions = $state<PackageQuestion[]>([]);
	let allQuestions = $state<Question[]>([]);
	let subjects = $state<Subject[]>([]);
	let packagesPromise = $state<Promise<PackagesOverview> | null>(null);
	let showForm = $state(false);

	let fSubjectId = $state('');
	let fTitle = $state('');
	let fDescription = $state('');
	let fDuration = $state(60);
	let fRandomize = $state(false);
	let fActive = $state(true);
	let fSelectedIds = new SvelteSet<string>();
	let fQuestionWeights = new SvelteMap<string, number>();
	let fBusy = $state(false);
	let deleteBusyId = $state('');
	let operationState = $state<{ tone: 'success' | 'error' | 'warning' | 'info'; title: string; message: string } | null>(null);
	let packagesRequestId = 0;
	let questionPoolTotal = $state(0);
	let hiddenEventPackageCount = $state(0);
	let eventContext = $state<EventContext | null>(null);
	const eventId = page.url.searchParams.get('event_id') ?? '';

	let questionPool = $derived(
		fSubjectId
			? allQuestions.filter(q => q.subject_id === fSubjectId && q.status === 'published' && isQuestionAllowedForPackage(q))
			: []
	);
	let hiddenScopedQuestionCount = $derived(
		fSubjectId
			? allQuestions.filter(q => q.subject_id === fSubjectId && q.status === 'published' && !isQuestionAllowedForPackage(q)).length
			: 0
	);
	let selectedQuestions = $derived(questionPool.filter((q) => fSelectedIds.has(q.id)));
	let selectedWeightTotal = $derived(selectedQuestions.reduce((sum, question) => sum + questionWeightValue(question.id), 0));
	let availableBlueprintMissingCount = $derived(questionPool.filter(questionHasBlueprintGap).length);
	let availableHotsCount = $derived(questionPool.filter((q) => q.hots_flag).length);
	let availableTypeBuckets = $derived(countByLabel(questionPool, (q) => questionTypeLabel(q.question_type)));
	let selectedBlueprintMatrix = $derived(buildBlueprintMatrix(selectedQuestions));
	let selectedTypeBuckets = $derived(countByLabel(selectedQuestions, (q) => questionTypeLabel(q.question_type)));
	let selectedCognitiveBuckets = $derived(countByLabel(selectedQuestions, (q) => compactValue(q.cognitive_level, 'Belum level')));
	let selectedHotsCount = $derived(selectedQuestions.filter((q) => q.hots_flag).length);
	let selectedBlueprintMissingCount = $derived(selectedQuestions.filter(questionHasBlueprintGap).length);
	let questionPoolCapped = $derived(questionPoolTotal > allQuestions.length);
	let packageReadinessIssues = $derived(buildPackageReadinessIssues());
	let canCreatePackage = $derived(packageReadinessIssues.length === 0 && !fBusy);
	let createPackageHref = $derived(`${resolve('/asesmen/paket/new')}${eventId ? `?event_id=${encodeURIComponent(eventId)}` : ''}`);

	function toggleQuestion(id: string) {
		if (fSelectedIds.has(id)) {
			fSelectedIds.delete(id);
			fQuestionWeights.delete(id);
		} else {
			fSelectedIds.add(id);
			if (!fQuestionWeights.has(id)) fQuestionWeights.set(id, 1);
		}
	}

	function handleSubjectChange() {
		fSelectedIds.clear();
		fQuestionWeights.clear();
	}

	function questionWeightValue(id: string) {
		const value = Number(fQuestionWeights.get(id) ?? 1);
		if (!Number.isFinite(value)) return 1;
		return Math.max(1, Math.min(100, Math.round(value)));
	}

	function setQuestionWeight(id: string, value: number) {
		const normalized = Number.isFinite(value) ? Math.max(1, Math.min(100, Math.round(value))) : 1;
		fQuestionWeights.set(id, normalized);
	}

	function handleQuestionWeightInput(id: string, event: Event) {
		const target = event.currentTarget as HTMLInputElement;
		setQuestionWeight(id, Number(target.value));
	}

	function isRecord(value: unknown): value is Record<string, unknown> {
		return typeof value === 'object' && value !== null;
	}

	function parseQuestionPage(payload: unknown) {
		if (Array.isArray(payload)) return { items: payload as Question[], total: payload.length };
		if (!isRecord(payload)) return { items: [], total: 0 };
		const data = payload as QuestionListPayload;
		const items = Array.isArray(data.items) ? data.items : [];
		const total = data.meta?.total ?? items.length;
		return { items, total };
	}

	function parseSubjects(payload: AcademicPayload | unknown) {
		return isRecord(payload) && Array.isArray(payload.subjects) ? (payload.subjects as Subject[]) : [];
	}

	function parsePackageQuestions(payload: CbtPackagesPayload | unknown) {
		return isRecord(payload) && Array.isArray(payload.questions) ? (payload.questions as PackageQuestion[]) : [];
	}

	async function fetchQuestionsPage(offset: number) {
		const params = new URLSearchParams({
			limit: String(questionPageSize),
			offset: String(offset),
			status: 'published',
		});
		params.set('scope', eventId ? 'event_pool' : 'global');
		if (eventId) params.set('event_id', eventId);
		const payload = await fetch(clientApiPathWithQuery('/api/bank-soal/questions', params))
			.then((response) => readClientApiData<unknown>(response, 'Gagal memuat bank soal'));
		return parseQuestionPage(payload);
	}

	async function fetchAllQuestions() {
		const firstPage = await fetchQuestionsPage(0);
		const questionsById = new SvelteMap(firstPage.items.map((question) => [question.id, question]));
		let loaded = firstPage.items.length;
		let pages = 1;
		while (loaded < firstPage.total && pages < maxQuestionPages) {
			const page = await fetchQuestionsPage(loaded);
			if (page.items.length === 0) break;
			for (const question of page.items) questionsById.set(question.id, question);
			loaded += page.items.length;
			pages += 1;
		}
		questionPoolTotal = Math.max(firstPage.total, questionsById.size);
		return Array.from(questionsById.values());
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

	function difficultyLabel(value: string | null | undefined) {
		const labels: Record<string, string> = {
			easy: 'Mudah',
			medium: 'Sedang',
			hard: 'Sulit',
		};
		const normalized = compactValue(value, '');
		return labels[normalized] ?? (normalized || '-');
	}

	function questionHasBlueprintGap(question: { cp_ref?: string; tp_ref?: string; kd_ref?: string; cognitive_level?: string }) {
		return !compactValue(question.cp_ref, '')
			|| (!compactValue(question.tp_ref, '') && !compactValue(question.kd_ref, ''))
			|| !compactValue(question.cognitive_level, '');
	}

	function questionReviewLabel(question: Question) {
		if (question.status === 'published') return 'Terbit';
		if (question.workflow_status === 'approved') return 'Disetujui';
		if (question.workflow_status === 'review') return 'Ditinjau';
		if (question.workflow_status === 'rejected') return 'Revisi';
		return 'Draft';
	}

	function questionReadinessIssues(question: Question) {
		const issues: string[] = [];
		if (question.status !== 'published') issues.push('belum terbit');
		if (questionHasBlueprintGap(question)) issues.push('metadata kurang');
		return issues;
	}

	function buildPackageReadinessIssues() {
		const issues: string[] = [];
		if (!fSubjectId) issues.push('Pilih mata pelajaran');
		if (!fTitle.trim()) issues.push('Isi nama paket');
		if (!fDuration || fDuration < 10) issues.push('Durasi minimal 10 menit');
		if (selectedQuestions.length === 0) issues.push('Pilih minimal 1 soal terbit');
		const blocked = selectedQuestions.filter((question) => question.status !== 'published');
		if (blocked.length > 0) issues.push(`${blocked.length} soal belum terbit`);
		const invalidWeights = selectedQuestions.filter((question) => {
			const weight = questionWeightValue(question.id);
			return weight < 1 || weight > 100;
		});
		if (invalidWeights.length > 0) issues.push('Bobot setiap soal harus 1-100');
		return issues;
	}

	function countByLabel<T>(questions: T[], selector: (question: T) => string): BlueprintBucket[] {
		const counts = new SvelteMap<string, number>();
		for (const question of questions) {
			const label = selector(question);
			counts.set(label, (counts.get(label) ?? 0) + 1);
		}
		return Array.from(counts.entries())
			.map(([label, count]) => ({ label, count }))
			.sort((a, b) => b.count - a.count || a.label.localeCompare(b.label));
	}

	function buildBlueprintMatrix(questions: Question[]): BlueprintMatrixRow[] {
		const rows = new SvelteMap<string, BlueprintMatrixRow>();
		for (const question of questions) {
			const cp = compactValue(question.cp_ref, 'Belum CP');
			const tp = compactValue(question.tp_ref, 'Belum TP');
			const kd = compactValue(question.kd_ref, 'Belum KD');
			const topic = compactValue(question.material_topic, 'Belum materi');
			const cognitive = compactValue(question.cognitive_level, 'Belum level');
			const missing = questionHasBlueprintGap(question);
			const key = [cp, tp, kd, topic, cognitive].join('|');
			const row = rows.get(key) ?? {
				key,
				cp,
				tp,
				kd,
				topic,
				cognitive,
				count: 0,
				hotsCount: 0,
				types: [],
				missing,
			};
			row.count += 1;
			if (question.hots_flag) row.hotsCount += 1;
			const typeLabel = questionTypeLabel(question.question_type);
			if (!row.types.includes(typeLabel)) row.types.push(typeLabel);
			row.missing = row.missing || missing;
			rows.set(key, row);
		}
		return Array.from(rows.values()).sort((a, b) => Number(b.missing) - Number(a.missing) || b.count - a.count || a.cp.localeCompare(b.cp));
	}

	function packageQualitySummary(packageID: string): PackageQualitySummary {
		const questions = packageQuestions.filter((question) => question.package_id === packageID);
		const typeBuckets = countByLabel(questions, (question) => questionTypeLabel(question.question_type));
		const cognitiveBuckets = countByLabel(questions, (question) => compactValue(question.cognitive_level, 'Belum level'));
		const hotsCount = questions.filter((question) => question.hots_flag).length;
		const missingCount = questions.filter(questionHasBlueprintGap).length;
		const unpublishedCount = questions.filter((question) => question.status !== 'published').length;
		const totalPoints = questions.reduce((sum, question) => sum + (Number(question.points) || 0), 0);
		return { questions, typeBuckets, cognitiveBuckets, hotsCount, missingCount, unpublishedCount, totalPoints };
	}

	function strictEventPackages(items: CbtPackage[]) {
		return eventId ? items.filter((pkg) => pkg.event_id === eventId) : items;
	}

	function hiddenPackageCount(items: CbtPackage[]) {
		return eventId ? items.filter((pkg) => pkg.event_id !== eventId).length : 0;
	}

	function normalizedScopeId(value: string | null | undefined) {
		return (value ?? '').trim();
	}

	function isGlobalQuestion(question: Question) {
		return normalizedScopeId(question.event_id) === '';
	}

	function isQuestionAllowedForPackage(question: Question) {
		const questionEventId = normalizedScopeId(question.event_id);
		if (!eventId) return questionEventId === '';
		return isGlobalQuestion(question) || questionEventId === eventId;
	}

	async function fetchEventContext() {
		if (!eventId) return null;
		try {
			return await fetch(clientApiPath`/api/asesmen/events/${eventId}`).then((response) => readClientApiData<EventContext>(response, 'Gagal memuat konteks kegiatan'));
		} catch {
			return null;
		}
	}

	async function fetchOverview(): Promise<PackagesOverview> {
		const packageParams = new URLSearchParams();
		if (eventId) packageParams.set('event_id', eventId);
		const [packagesPayload, questionItems, academicPayload] = await Promise.all([
			fetch(clientApiPathWithQuery('/api/asesmen/packages', packageParams)).then((response) => readClientApiData<CbtPackagesPayload>(response, 'Gagal memuat data paket')),
			fetchAllQuestions(),
			fetch('/api/academic').then((response) => readClientApiData<AcademicPayload>(response, 'Gagal memuat data akademik')),
		]);
		return {
			packages: packagesPayload.packages ?? [],
			packageQuestions: parsePackageQuestions(packagesPayload),
			allQuestions: questionItems,
			subjects: parseSubjects(academicPayload),
		};
	}

	function applyOverview(overview: PackagesOverview) {
		hiddenEventPackageCount = hiddenPackageCount(overview.packages);
		packages = strictEventPackages(overview.packages);
		packageQuestions = overview.packageQuestions;
		allQuestions = overview.allQuestions;
		subjects = overview.subjects;
	}

	function load() {
		const requestId = ++packagesRequestId;
		packages = [];
		packageQuestions = [];
		allQuestions = [];
		subjects = [];
		questionPoolTotal = 0;
		hiddenEventPackageCount = 0;
		void fetchEventContext().then((context) => { eventContext = context; });
		packagesPromise = fetchOverview().then((overview) => {
			if (requestId !== packagesRequestId) return { packages, packageQuestions, allQuestions, subjects };
			applyOverview(overview);
			return overview;
		}).catch((error: unknown) => {
			if (requestId === packagesRequestId) throw error;
			return { packages, packageQuestions, allQuestions, subjects };
		});
	}

	async function refreshPackages() {
		if (!packagesPromise) {
			load();
			return;
		}
		const requestId = ++packagesRequestId;
		try {
			const overview = await fetchOverview();
			if (requestId !== packagesRequestId) return;
			applyOverview(overview);
			packagesPromise = Promise.resolve(overview);
		} catch (error) {
			if (requestId === packagesRequestId) {
				packagesPromise = Promise.resolve({ packages, packageQuestions, allQuestions, subjects });
				toast.error(packagesErrorMessage(error));
			}
		}
	}

	function retryPackages(reset?: () => void) {
		reset?.();
		load();
	}

	function packagesErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Gagal memuat data paket';
	}

	function handlePackagesRenderError(error: unknown) {
		console.error('Daftar paket soal belum dapat ditampilkan', error);
	}

	function showToast(msg: string) {
		toast.success(msg);
	}

	function showError(msg: string) {
		toast.error(msg);
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

	function packageLegacyMutationPath(id: string) {
		return clientApiPathWithQuery('/api/asesmen/packages', new URLSearchParams({ id }));
	}

	async function createPackage() {
		if (packageReadinessIssues.length > 0) {
			setOperationState('warning', 'Paket Belum Siap', packageReadinessIssues[0] ?? 'Lengkapi paket ujian terlebih dahulu.');
			return;
		}
		fBusy = true;
		try {
			const res = await fetch('/api/asesmen/packages', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					subject_id: fSubjectId, title: fTitle, description: fDescription,
					duration_minutes: fDuration, randomize_questions: fRandomize,
					is_active: fActive,
					...(eventId ? { event_id: eventId } : {}),
					question_ids: selectedQuestions.map((question) => question.id),
					question_weights: Object.fromEntries(selectedQuestions.map((question) => [question.id, questionWeightValue(question.id)])),
				}),
			});
			await readClientJson<unknown>(res);
			fSubjectId = ''; fTitle = ''; fDescription = ''; fDuration = 60;
			fRandomize = false; fActive = true; fSelectedIds.clear(); fQuestionWeights.clear();
			showForm = false;
			setOperationState('success', 'Paket Berhasil Dibuat', 'Paket ujian baru sudah tersimpan dan siap dipakai untuk sesi ujian.');
			showToast('Paket ujian berhasil dibuat');
			await refreshPackages();
		} catch (error) {
			showError(mutationErrorMessage(error, 'Gagal membuat paket. Periksa koneksi lalu coba lagi.'));
		} finally { fBusy = false; }
	}

	async function deletePackage(id: string, title: string) {
		if (!(await confirmPhrase('Hapus Paket Ujian', `Paket "${title}" akan dihapus dari daftar. Tindakan ini tidak bisa dibatalkan dari layar operator.`, 'HAPUS'))) return;
		deleteBusyId = id;
		try {
			const res = await fetch(packageLegacyMutationPath(id), { method: 'DELETE' });
			await readClientJson<unknown>(res);
			setOperationState('warning', 'Paket Dihapus', `Paket "${title}" sudah dihapus dari daftar paket ujian.`);
			showToast('Paket dihapus');
			await refreshPackages();
		} catch (error) {
			setOperationState('error', 'Paket Gagal Dihapus', 'Periksa kembali apakah paket masih dipakai oleh sesi aktif atau coba ulang beberapa saat lagi.');
			showError(mutationErrorMessage(error, 'Gagal menghapus paket. Periksa koneksi lalu coba lagi.'));
		} finally {
			deleteBusyId = '';
		}
	}

	onMount(() => {
		void load();
	});
</script>

	<svelte:head><title>{eventId ? 'Paket Kegiatan Ujian' : 'Format Paket Ujian'} — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6">
	<section class="rounded-3xl border border-primary/20 bg-gradient-to-br from-primary/10 via-card to-primary/10 p-6 shadow-sm">
		<div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
			<div class="max-w-3xl space-y-3">
				<p class="text-xs font-semibold uppercase tracking-[0.2em] text-primary">Keranjang Soal CBT</p>
				<h1 class="text-3xl font-semibold tracking-tight text-foreground">{eventId ? 'Paket Soal Event' : 'Paket Soal'}</h1>
				<p class="max-w-2xl text-sm leading-6 text-muted-foreground">
					Pilih soal terbit dari Bank Soal, masukkan ke paket, lalu pakai paket itu saat membuat sesi event.
				</p>
			</div>
			<div class="flex flex-wrap gap-2">
				{#if eventId}
					<a href={resolve(`/asesmen/kegiatan/${eventId}`)} class="inline-flex items-center rounded-md border border-success/20 bg-success/10 px-3 py-2 text-sm font-semibold text-success hover:bg-success/15">Kembali ke Event</a>
				{/if}
				<a href={resolve('/asesmen')} class="inline-flex items-center rounded-md border border-success/20 bg-card px-3 py-2 text-sm font-semibold text-success hover:bg-success/10">Beranda Ujian</a>
				<a href={createPackageHref} class="inline-flex items-center rounded-md bg-primary px-3 py-2 text-sm font-semibold text-primary-foreground hover:bg-primary/90">Buat Paket</a>
			</div>
		</div>

		<div class="mt-5 grid gap-3 md:grid-cols-3">
			<a href="#paket-saya" class="rounded-2xl border border-primary/20 bg-card p-4 text-sm text-primary shadow-sm transition hover:border-primary">
				<p class="text-xs font-semibold uppercase tracking-[0.18em] text-primary">Paket Saya</p>
				<p class="mt-2 text-lg font-semibold">Lihat daftar paket</p>
				<p class="mt-1 leading-6 text-muted-foreground">Daftar paket menjadi pusat kerja utama.</p>
			</a>
			<a
				href={createPackageHref}
				class="rounded-2xl border border-primary/20 bg-card/70 p-4 text-left text-sm text-foreground shadow-sm transition hover:border-primary/20 hover:bg-card"
			>
				<p class="text-xs font-semibold uppercase tracking-[0.18em] text-primary">Buat Paket</p>
				<p class="mt-2 text-lg font-semibold">Buka builder soal</p>
				<p class="mt-1 leading-6 text-muted-foreground">Masuk ke halaman create-only agar alur utama tetap fokus.</p>
			</a>
			<a
				href={resolve(eventId ? `/asesmen/kegiatan/${eventId}` : '/asesmen/kegiatan')}
				class="rounded-2xl border border-primary/20 bg-card/70 p-4 text-sm text-foreground shadow-sm transition hover:border-primary/20 hover:bg-card"
			>
				<p class="text-xs font-semibold uppercase tracking-[0.18em] text-primary">Pakai untuk Event</p>
				<p class="mt-2 text-lg font-semibold">Buat/Cek Sesi Event</p>
				<p class="mt-1 leading-6 text-muted-foreground">Lanjutkan paket ke sesi, ruang, dan token.</p>
			</a>
		</div>
	</section>

	<details class="rounded-2xl border border-border bg-card px-4 py-3 text-sm text-foreground shadow-sm">
		<summary class="cursor-pointer font-semibold text-foreground">Catatan penggunaan paket dan cakupan soal</summary>
		<div class="mt-3 space-y-3 leading-6">
			{#if eventId}
				<p><span class="font-semibold text-primary">Paket event:</span> {eventContext?.title ?? eventId}. Sesi event membutuhkan paket yang membawa <code class="rounded bg-muted px-1">event_id</code> event ini.</p>
				<p>Pilihan soal tetap memakai filter Bank Soal reusable yang sudah ada, ditambah soal khusus event ini saja.</p>
			{:else}
				<p><span class="font-semibold text-warning">Paket global:</span> dapat dipakai sebagai template reusable atau paket standalone. Jika bekerja dari Kegiatan Ujian, buka builder paket dari event agar paket otomatis tertaut event.</p>
			{/if}
			{#if hiddenEventPackageCount > 0}
				<p>{hiddenEventPackageCount} template global atau paket event lain disembunyikan dari daftar event ini.</p>
			{/if}
			<a href={resolve('/bank-soal')} class="inline-flex rounded-md border border-success/20 bg-success/10 px-3 py-2 text-sm font-semibold text-success hover:bg-success/15">Buka Bank Soal</a>
		</div>
	</details>

	{#if operationState}
		<OperationStatusPanel {...operationState} />
	{/if}

	{#if showForm}
		<Card.Root id="buat-paket" class="border-primary/20 shadow-sm">
			<Card.Header class="pb-2">
				<Card.Title class="text-base">Buat Paket</Card.Title>
				<Card.Description>Builder sederhana untuk memilih mapel, mengisi identitas paket, lalu memasukkan soal seperti keranjang.</Card.Description>
			</Card.Header>
			<Card.Content class="space-y-4">
				<div class="grid gap-3 sm:grid-cols-2">
					<div>
						<label for="package-subject-id" class="text-xs text-muted-foreground mb-1 block">Mata Pelajaran <span class="text-destructive">*</span></label>
						<select id="package-subject-id" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fSubjectId} onchange={handleSubjectChange}>
							<option value="">-- Pilih --</option>
								{#each subjects as s (s.id)}
								<option value={s.id}>{s.code} — {s.name}</option>
							{/each}
						</select>
					</div>
					<div>
						<label for="package-title" class="text-xs text-muted-foreground mb-1 block">Nama Paket <span class="text-destructive">*</span></label>
						<Input id="package-title" placeholder="mis: UTS Matematika Sem 1 2025" bind:value={fTitle} />
					</div>
					<div>
						<label for="package-duration" class="text-xs text-muted-foreground mb-1 block">Durasi (menit) <span class="text-destructive">*</span></label>
						<Input id="package-duration" type="number" min={10} max={300} bind:value={fDuration} />
					</div>
					<div class="flex items-end gap-4 pb-1">
						<label class="flex items-center gap-2 text-sm">
							<input type="checkbox" bind:checked={fRandomize} class="rounded" />
							Acak urutan soal
						</label>
						<label class="flex items-center gap-2 text-sm">
							<input type="checkbox" bind:checked={fActive} class="rounded" />
							Paket aktif
						</label>
					</div>
				</div>

				<div>
					<label for="package-description" class="text-xs text-muted-foreground mb-1 block">Deskripsi (opsional)</label>
					<Textarea id="package-description" placeholder="Keterangan paket ujian..." rows={2} bind:value={fDescription} />
				</div>

				{#if fSubjectId}
					<div>
						<div class="mb-2 flex flex-wrap items-center justify-between gap-2">
							<div class="text-xs text-muted-foreground">
								Pilih Soal dari Bank Soal ({questionPool.length} soal terbit sesuai cakupan)
								{#if selectedQuestions.length > 0}
									— <span class="font-medium text-success">{selectedQuestions.length} dipilih</span>
								{/if}
							</div>
							<details class="rounded-md border border-border bg-muted/50 px-3 py-2 text-xs text-foreground">
								<summary class="cursor-pointer font-medium">Info pool soal</summary>
								<div class="mt-2 space-y-2 leading-5">
									<p>{questionPool.length} soal terbit tersedia, {availableBlueprintMissingCount} metadata kurang, {availableHotsCount} HOTS.</p>
									{#if questionPoolCapped}
										<p>Pool soal dibatasi: termuat {allQuestions.length} dari {questionPoolTotal} soal terbit. Rapikan status/mapel di Bank Soal bila soal belum muncul.</p>
									{/if}
									{#if availableTypeBuckets.length > 0}
										<p>Bentuk soal: {availableTypeBuckets.map((bucket) => `${bucket.label}: ${bucket.count}`).join(', ')}</p>
									{/if}
								</div>
							</details>
						</div>
						{#if questionPool.length === 0}
							<p class="text-sm text-muted-foreground py-4 text-center border rounded-md">
								Belum ada soal berstatus "Terbit" untuk mata pelajaran ini dalam cakupan paket ini. Paket global hanya memakai soal reusable; paket event memakai soal reusable dan soal khusus event yang sama.
							</p>
							{#if hiddenScopedQuestionCount > 0}
								<p class="mt-2 rounded-md border border-accent bg-accent/60 px-3 py-2 text-xs text-accent-foreground">
									{hiddenScopedQuestionCount} soal terbit disembunyikan karena {eventId ? 'tertaut ke kegiatan lain' : 'khusus kegiatan tertentu'}.
								</p>
							{/if}
						{:else}
							{#if hiddenScopedQuestionCount > 0}
								<div class="mb-2 rounded-md border border-accent bg-accent/60 px-3 py-2 text-xs text-accent-foreground">
									{hiddenScopedQuestionCount} soal terbit disembunyikan karena {eventId ? 'tertaut ke kegiatan lain' : 'khusus kegiatan tertentu'}. Pool ini hanya memakai {eventId ? 'soal pakai ulang dan soal kegiatan ini' : 'soal pakai ulang/global'}.
								</div>
							{/if}
							<div class="border rounded-md max-h-64 overflow-y-auto">
									{#each questionPool as q (q.id)}
									<label class="flex items-start gap-3 px-3 py-2 hover:bg-muted/50 cursor-pointer border-b last:border-b-0">
										<input type="checkbox" checked={fSelectedIds.has(q.id)} onchange={() => toggleQuestion(q.id)} class="mt-0.5 rounded" />
										<div class="flex-1 min-w-0">
											<p class="text-sm line-clamp-1">{q.question_text}</p>
											<div class="flex flex-wrap gap-1 mt-0.5">
												{#if q.code}
													<span class="text-xs text-muted-foreground font-mono">{q.code}</span>
												{/if}
												<Badge variant="outline" class="text-xs py-0">{questionTypeLabel(q.question_type)}</Badge>
												<Badge variant="outline" class="text-xs py-0">{difficultyLabel(q.difficulty)}</Badge>
												<Badge class="border-success/20 bg-success/10 text-success text-xs py-0">{questionReviewLabel(q)}</Badge>
												{#if q.cognitive_level}
													<Badge variant="secondary" class="text-xs py-0">{q.cognitive_level}</Badge>
												{/if}
												{#if q.hots_flag}
													<Badge class="border-warning/30 bg-warning/10 text-warning text-xs py-0">HOTS</Badge>
												{/if}
												{#each questionReadinessIssues(q) as issue (`${q.id}-${issue}`)}
													<Badge class="border-warning/30 bg-warning/10 text-warning text-xs py-0">{issue}</Badge>
												{/each}
											</div>
										</div>
									</label>
								{/each}
							</div>
						{/if}
						{#if selectedQuestions.length > 0}
							<div class="mt-3 rounded-lg border border-primary/20 bg-primary/10 p-3">
								<div class="flex flex-wrap items-start justify-between gap-2">
									<div>
										<p class="text-xs font-semibold uppercase tracking-[0.18em] text-primary">Keranjang Paket</p>
										<p class="mt-1 text-xs text-primary">
											{selectedQuestions.length} soal dipilih dengan bobot total {selectedWeightTotal}.
										</p>
									</div>
									<span class="rounded-full border border-primary/20 bg-card px-2 py-1 text-xs font-medium text-primary">
										{selectedBlueprintMissingCount > 0 ? `${selectedBlueprintMissingCount} metadata kurang` : 'Metadata siap'}
									</span>
								</div>

								<div class="mt-3 overflow-hidden rounded-md border border-primary/20 bg-card">
									<div class="grid grid-cols-[1fr_5rem] gap-2 border-b bg-primary/10 px-2 py-1.5 text-[11px] font-semibold uppercase tracking-[0.12em] text-primary">
										<span>Soal Terpilih</span>
										<span class="text-right">Bobot</span>
									</div>
									<div class="max-h-36 overflow-y-auto">
										{#each selectedQuestions as question (question.id)}
											<div class="grid grid-cols-[1fr_5rem] items-center gap-2 border-b px-2 py-1.5 last:border-b-0">
												<div class="min-w-0">
													<p class="truncate text-xs font-medium text-foreground">{question.code || 'Tanpa kode'} · {question.question_text}</p>
													<p class="text-[11px] text-muted-foreground">{questionTypeLabel(question.question_type)} · {difficultyLabel(question.difficulty)}</p>
												</div>
												<Input
													aria-label={`Bobot ${question.code || question.question_text}`}
													class="h-8 text-right text-xs"
													type="number"
													min={1}
													max={100}
													value={questionWeightValue(question.id)}
													oninput={(event) => handleQuestionWeightInput(question.id, event)}
												/>
											</div>
										{/each}
									</div>
								</div>

								<details class="mt-3 rounded-md border border-primary/20 bg-card px-3 py-2 text-xs text-foreground">
									<summary class="cursor-pointer font-semibold text-foreground">Lihat ringkasan blueprint dan mutu</summary>
									<div class="mt-3 grid gap-3 xl:grid-cols-[0.75fr_1.25fr]">
									<div class="space-y-2 text-xs">
										<div>
											<p class="mb-1 font-semibold text-muted-foreground">Bentuk soal</p>
											<div class="flex flex-wrap gap-1">
												{#each selectedTypeBuckets as bucket (bucket.label)}
													<Badge variant="outline" class="bg-card">{bucket.label}: {bucket.count}</Badge>
												{/each}
											</div>
										</div>
										<div>
											<p class="mb-1 font-semibold text-muted-foreground">Level kognitif</p>
											<div class="flex flex-wrap gap-1">
												{#each selectedCognitiveBuckets as bucket (bucket.label)}
													<Badge variant={bucket.label === 'Belum level' ? 'secondary' : 'outline'} class="bg-card">{bucket.label}: {bucket.count}</Badge>
												{/each}
											</div>
										</div>
										<p class="rounded-md border border-primary/20 bg-card px-2 py-1.5 text-muted-foreground">
											Gunakan panel ini untuk mencegah paket terlalu menumpuk di satu CP/KD sebelum sesi ujian dibuat.
										</p>
									</div>

									<div class="overflow-hidden rounded-md border border-primary/20 bg-card">
										<div class="grid grid-cols-[1fr_1fr_1fr_0.8fr_0.5fr] gap-2 border-b bg-primary/10 px-2 py-1.5 text-[11px] font-semibold uppercase tracking-[0.12em] text-primary">
											<span>CP</span>
											<span>TP</span>
											<span>KD / Materi</span>
											<span>Level</span>
											<span class="text-right">Soal</span>
										</div>
										<div class="max-h-40 overflow-y-auto">
											{#each selectedBlueprintMatrix as row (row.key)}
												<div class="grid grid-cols-[1fr_1fr_1fr_0.8fr_0.5fr] gap-2 border-b px-2 py-1.5 text-xs last:border-b-0 {row.missing ? 'bg-warning/10' : ''}">
													<span class="min-w-0 truncate font-medium text-foreground" title={row.cp}>{row.cp}</span>
													<span class="min-w-0 truncate text-muted-foreground" title={row.tp}>{row.tp}</span>
													<span class="min-w-0 truncate text-muted-foreground" title={`${row.kd} / ${row.topic}`}>{row.kd} / {row.topic}</span>
													<span class="min-w-0 truncate text-muted-foreground" title={row.types.join(', ')}>{row.cognitive}</span>
													<span class="text-right font-semibold text-foreground">
														{row.count}
														{#if row.hotsCount > 0}
															<span class="text-warning">/{row.hotsCount}</span>
														{/if}
													</span>
												</div>
											{/each}
										</div>
									</div>
									</div>
								</details>
							</div>
						{/if}
						{#if packageReadinessIssues.length > 0}
							<div class="mt-3 rounded-md border border-warning/30 bg-warning/10 px-3 py-2 text-xs text-warning">
								<span class="font-semibold">Belum siap dibuat:</span>
								<span>{packageReadinessIssues.join(', ')}</span>
							</div>
						{:else}
							<div class="mt-3 rounded-md border border-success/20 bg-success/10 px-3 py-2 text-xs font-medium text-success">
								Paket siap dibuat dengan soal terbit yang sudah terpilih.
							</div>
						{/if}
					</div>
				{/if}

				<div class="flex gap-2">
					<LoadingButton disabled={!canCreatePackage} onclick={() => void createPackage()} loading={fBusy} loadingLabel="Menyimpan...">
						{`Buat Paket${selectedQuestions.length > 0 ? ` (${selectedQuestions.length} soal)` : ''}`}
					</LoadingButton>
					<LoadingButton variant="outline" onclick={() => (showForm = false)}>Batal</LoadingButton>
				</div>
			</Card.Content>
		</Card.Root>
	{/if}

	<AsyncContent promise={packagesPromise} onerror={handlePackagesRenderError}>
		{#snippet pending()}
			<Card.Root class="overflow-hidden border-border shadow-sm">
				<Card.Content class="space-y-3 p-6">
					{#each Array.from({ length: 5 }) as _, index (`package-row-skeleton-${index}`)}
						<div class="grid gap-3 lg:grid-cols-[1.2fr_0.6fr_0.5fr_0.5fr_1fr_0.5fr_0.6fr_auto] lg:items-center">
							<Skeleton class="h-5 w-40" />
							<Skeleton class="h-6 w-16" />
							<Skeleton class="h-5 w-14" />
							<Skeleton class="h-5 w-12" />
							<Skeleton class="h-7 w-48" />
							<Skeleton class="h-6 w-12" />
							<Skeleton class="h-6 w-16" />
							<Skeleton class="h-9 w-20 justify-self-end" />
						</div>
					{/each}
				</Card.Content>
			</Card.Root>
		{/snippet}

		{#snippet failed(error, reset)}
			<RecoveryPanel
				title="Paket Ujian Belum Tersaji"
				message={packagesErrorMessage(error)}
				onRetry={() => retryPackages(reset)}
			/>
		{/snippet}

		{#snippet children(value)}
			{@const overview = value as PackagesOverview}
			{@const eventPackages = strictEventPackages(overview.packages)}
			{@const currentPackages = eventPackages}
			{@const hiddenPackages = hiddenPackageCount(overview.packages)}
		<Card.Root id="paket-saya" class="overflow-hidden border-border shadow-sm">
			<Card.Header class="pb-2">
				<div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
					<div>
						<p class="text-xs font-semibold uppercase tracking-[0.18em] text-primary">Paket Saya</p>
						<Card.Title class="mt-1 text-base">Daftar Paket ({currentPackages.length})</Card.Title>
					</div>
					<a href={createPackageHref} class="inline-flex items-center rounded-md bg-primary px-3 py-2 text-sm font-semibold text-primary-foreground hover:bg-primary/90">Buat Paket</a>
				</div>
				{#if hiddenPackages > 0}
					<Card.Description>{hiddenPackages} template global atau paket event lain disembunyikan dari daftar event ini.</Card.Description>
				{/if}
			</Card.Header>
			<Card.Content class="p-0">
				<div class="hidden overflow-x-auto lg:block">
				<Table.Root>
					<Table.Header>
						<Table.Row>
							<Table.Head>Nama Paket</Table.Head>
							<Table.Head>Mapel</Table.Head>
							<Table.Head>Durasi</Table.Head>
							<Table.Head>Jml Soal</Table.Head>
							<Table.Head>Mutu Paket</Table.Head>
							<Table.Head>Acak</Table.Head>
							<Table.Head>Status</Table.Head>
							<Table.Head></Table.Head>
						</Table.Row>
					</Table.Header>
					<Table.Body>
						{#each currentPackages as p (p.id)}
							{@const quality = packageQualitySummary(p.id)}
							<Table.Row>
								<Table.Cell class="font-medium">{p.title}</Table.Cell>
								<Table.Cell>
									<Badge variant="outline" class="text-xs">{p.subject_code}</Badge>
								</Table.Cell>
								<Table.Cell class="text-muted-foreground">{p.duration_minutes} mnt</Table.Cell>
								<Table.Cell>
									<span class="font-mono text-sm">{p.question_count}</span>
									{#if quality.totalPoints > 0}
										<span class="ml-1 text-xs text-muted-foreground">/{quality.totalPoints} poin</span>
									{/if}
								</Table.Cell>
								<Table.Cell>
									{#if quality.questions.length === 0}
										<span class="text-xs text-muted-foreground">Belum ada rincian</span>
									{:else}
										<details class="max-w-sm text-xs text-muted-foreground">
											<summary class="cursor-pointer font-medium text-foreground">
												{quality.typeBuckets.slice(0, 2).map((bucket) => `${bucket.label}: ${bucket.count}`).join(', ')}
											</summary>
											<div class="mt-2 space-y-1 rounded-md border border-border bg-muted/50 p-2 leading-5">
												<p>Bentuk: {quality.typeBuckets.map((bucket) => `${bucket.label}: ${bucket.count}`).join(', ')}</p>
												<p>Level: {quality.cognitiveBuckets.map((bucket) => `${bucket.label}: ${bucket.count}`).join(', ')}</p>
												<p>HOTS {quality.hotsCount}, metadata kurang {quality.missingCount}, belum terbit {quality.unpublishedCount}.</p>
											</div>
										</details>
									{/if}
								</Table.Cell>
								<Table.Cell>
									{#if p.randomize_questions}
										<Badge class="bg-primary/15 text-primary border-primary/20 text-xs">Ya</Badge>
									{:else}
										<Badge variant="secondary" class="text-xs">Tidak</Badge>
									{/if}
								</Table.Cell>
								<Table.Cell>
									{#if p.is_active}
										<Badge class="bg-primary/15 text-primary border-primary/20">Aktif</Badge>
									{:else}
										<Badge variant="secondary">Nonaktif</Badge>
									{/if}
								</Table.Cell>
								<Table.Cell>
					<a
						href={`${resolve('/asesmen/paket')}/${p.id}`}
						class="mr-2 inline-flex items-center rounded-md border border-border px-2 py-1 text-xs font-semibold hover:bg-muted"
					>
						Detail/Edit
					</a>
					<LoadingButton
						variant="destructive"
						size="xs"
						onclick={() => deletePackage(p.id, p.title)}
										loading={deleteBusyId === p.id}
										disabled={deleteBusyId !== '' && deleteBusyId !== p.id}
										loadingLabel="Menghapus..."
									>
										Hapus
									</LoadingButton>
								</Table.Cell>
							</Table.Row>
						{:else}
							<Table.Row>
								<Table.Cell colspan={8} class="text-center text-muted-foreground py-8">Belum ada paket ujian</Table.Cell>
							</Table.Row>
						{/each}
					</Table.Body>
				</Table.Root>
				</div>

				<div class="grid gap-3 p-4 lg:hidden">
					{#each currentPackages as p (p.id)}
						{@const quality = packageQualitySummary(p.id)}
						<div class="rounded-2xl border border-border bg-card p-4 shadow-sm">
							<div class="flex items-start justify-between gap-3">
								<div class="min-w-0">
									<p class="text-sm font-semibold text-foreground">{p.title}</p>
									<p class="mt-1 text-xs text-muted-foreground">{p.subject_name} ({p.subject_code})</p>
								</div>
								{#if p.is_active}
									<Badge class="bg-primary/15 text-primary border-primary/20">Aktif</Badge>
								{:else}
									<Badge variant="secondary">Nonaktif</Badge>
								{/if}
							</div>
							<div class="mt-3 flex flex-wrap items-center gap-2">
								<Badge variant="outline" class="text-xs">{p.duration_minutes} menit</Badge>
								<Badge variant="secondary">{p.question_count} soal</Badge>
								{#if quality.totalPoints > 0}
									<Badge variant="outline" class="text-xs">{quality.totalPoints} poin</Badge>
								{/if}
								{#if p.randomize_questions}
									<Badge class="bg-primary/10 text-primary border-primary/20 text-xs">Acak</Badge>
								{/if}
							</div>
							<details class="mt-3 rounded-lg border border-border bg-muted/50 px-3 py-2 text-xs text-muted-foreground">
								<summary class="cursor-pointer font-medium text-foreground">Mutu paket</summary>
								<div class="mt-2 flex flex-wrap items-center gap-1.5">
								{#if quality.questions.length === 0}
									<span class="text-xs text-muted-foreground">Rincian mutu belum tersedia</span>
								{:else}
									{#each quality.typeBuckets.slice(0, 3) as bucket (bucket.label)}
										<Badge variant="outline" class="bg-card text-xs">{bucket.label}: {bucket.count}</Badge>
									{/each}
									{#each quality.cognitiveBuckets.slice(0, 1) as bucket (bucket.label)}
										<Badge variant="secondary" class="text-xs">{bucket.label}: {bucket.count}</Badge>
									{/each}
									{#if quality.hotsCount > 0}
										<Badge class="border-warning/30 bg-warning/10 text-warning text-xs">{quality.hotsCount} HOTS</Badge>
									{/if}
									{#if quality.missingCount > 0}
										<Badge class="border-warning/30 bg-warning/10 text-warning text-xs">{quality.missingCount} metadata kurang</Badge>
									{/if}
									{#if quality.unpublishedCount > 0}
										<Badge class="border-destructive/30 bg-destructive/10 text-destructive text-xs">{quality.unpublishedCount} belum terbit</Badge>
									{/if}
								{/if}
								</div>
							</details>
							{#if p.description}
								<p class="mt-3 text-sm text-muted-foreground">{p.description}</p>
							{/if}
							<div class="mt-4 flex gap-2">
								<a
									href={`${resolve('/asesmen/paket')}/${p.id}`}
									class="inline-flex flex-1 items-center justify-center rounded-md border border-border px-3 py-2 text-sm font-semibold hover:bg-muted"
								>
									Detail/Edit
								</a>
								<LoadingButton
									variant="destructive"
									size="sm"
									class="w-full"
									onclick={() => deletePackage(p.id, p.title)}
									loading={deleteBusyId === p.id}
									disabled={deleteBusyId !== '' && deleteBusyId !== p.id}
									loadingLabel="Menghapus..."
								>
									Hapus
								</LoadingButton>
							</div>
						</div>
					{:else}
						<div class="rounded-2xl border border-dashed border-border bg-muted/50 px-4 py-10 text-center text-sm text-muted-foreground">
							Belum ada paket ujian
						</div>
					{/each}
				</div>
			</Card.Content>
		</Card.Root>
		{/snippet}
	</AsyncContent>
</div>
