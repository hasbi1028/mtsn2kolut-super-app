<script lang="ts">
	import { onMount } from 'svelte';
	import { SvelteMap, SvelteSet } from 'svelte/reactivity';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import * as Card from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Textarea } from '$lib/components/ui/textarea';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { toast } from '$lib/components/ui/sonner';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import OperationStatusPanel from '$lib/components/OperationStatusPanel.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { clientApiPath, clientApiPathWithQuery, readClientApiData, readClientJson } from '$lib/client/api';

	type Question = {
		id: string;
		subject_id: string;
		subject_code: string;
		code: string;
		question_text: string;
		difficulty: string;
		status: string;
		event_id?: string | null;
		workflow_status?: string;
		question_type?: string;
		cp_ref?: string;
		tp_ref?: string;
		kd_ref?: string;
		material_topic?: string;
		cognitive_level?: string;
		hots_flag?: boolean;
	};
	type Subject = { id: string; name: string; code: string };
	type EventContext = { id: string; title: string; status: string; target_levels?: string[]; academic_year_name?: string };
	type AcademicPayload = { subjects?: Subject[] };
	type QuestionListPayload = { items?: Question[]; meta?: { total?: number; limit?: number; offset?: number } };
	type QuestionCompleteness = {
		requirements?: { scope_mode?: string; target_pg?: number; target_essay?: number; status_filter?: string };
		rows?: Array<{ subject_id: string; subject_name: string; level: string; scope_mode?: string; target_pg: number; available_pg: number; missing_pg: number; target_essay: number; available_essay: number; missing_essay: number; complete: boolean }>;
		contributions?: Array<{ subject_id: string; subject_name: string; level: string; teacher_name: string; teacher_username: string; available_pg: number; available_essay: number }>;
	};
	type FormData = { allQuestions: Question[]; subjects: Subject[]; eventContext: EventContext | null; completeness: QuestionCompleteness | null };
	type BlueprintBucket = { label: string; count: number };
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
	const eventId = page.url.searchParams.get('event_id') ?? '';

	let allQuestions = $state<Question[]>([]);
	let subjects = $state<Subject[]>([]);
	let eventContext = $state<EventContext | null>(null);
	let questionCompleteness = $state<QuestionCompleteness | null>(null);
	let formPromise = $state<Promise<FormData> | null>(null);
	let questionPoolTotal = $state(0);
	let fSubjectId = $state('');
	let fTitle = $state('');
	let fDescription = $state('');
	let fDuration = $state(60);
	let fRandomize = $state(false);
	let fRandomizeOptions = $state(false);
	let fSourceMode = $state<'teacher_class' | 'level_subject_teachers' | 'event_pool'>('teacher_class');
	let fDrawPgCount = $state(0);
	let fDrawEssayCount = $state(0);
	let fRandomSeed = $state('');
	let fActive = $state(true);
	let fSelectedIds = new SvelteSet<string>();
	let fQuestionWeights = new SvelteMap<string, number>();
	let fBusy = $state(false);
	let operationState = $state<{ tone: 'success' | 'error' | 'warning' | 'info'; title: string; message: string } | null>(null);

	let questionPool = $derived(
		fSubjectId ? allQuestions.filter((question) => question.subject_id === fSubjectId && question.status === 'published' && isQuestionAllowedForPackage(question)) : []
	);
	let hiddenScopedQuestionCount = $derived(
		fSubjectId ? allQuestions.filter((question) => question.subject_id === fSubjectId && question.status === 'published' && !isQuestionAllowedForPackage(question)).length : 0
	);
	let selectedQuestions = $derived(questionPool.filter((question) => fSelectedIds.has(question.id)));
	let selectedWeightTotal = $derived(selectedQuestions.reduce((sum, question) => sum + questionWeightValue(question.id), 0));
	let availableBlueprintMissingCount = $derived(questionPool.filter(questionHasBlueprintGap).length);
	let availableHotsCount = $derived(questionPool.filter((question) => question.hots_flag).length);
	let availableTypeBuckets = $derived(countByLabel(questionPool, (question) => questionTypeLabel(question.question_type)));
	let selectedBlueprintMatrix = $derived(buildBlueprintMatrix(selectedQuestions));
	let selectedTypeBuckets = $derived(countByLabel(selectedQuestions, (question) => questionTypeLabel(question.question_type)));
	let selectedCognitiveBuckets = $derived(countByLabel(selectedQuestions, (question) => compactValue(question.cognitive_level, 'Belum level')));
	let selectedBlueprintMissingCount = $derived(selectedQuestions.filter(questionHasBlueprintGap).length);
	let questionPoolCapped = $derived(questionPoolTotal > allQuestions.length);
	let packageReadinessIssues = $derived(buildPackageReadinessIssues());
	let canCreatePackage = $derived(packageReadinessIssues.length === 0 && !fBusy);
	let eventSubjectCompletenessRows = $derived((questionCompleteness?.rows ?? []).filter((row) => row.subject_id === fSubjectId));
	let eventSubjectMissingRows = $derived(eventSubjectCompletenessRows.filter((row) => !row.complete));
	let eventSubjectContributions = $derived((questionCompleteness?.contributions ?? []).filter((row) => row.subject_id === fSubjectId));
	let selectedPgCount = $derived(selectedQuestions.filter((question) => question.question_type === 'multiple_choice').length);
	let selectedEssayCount = $derived(selectedQuestions.filter((question) => question.question_type === 'essay').length);
	let listHref = $derived(`${resolve('/asesmen/paket')}${eventId ? `?event_id=${encodeURIComponent(eventId)}` : ''}`);

	function isRecord(value: unknown): value is Record<string, unknown> {
		return typeof value === 'object' && value !== null;
	}

	function parseQuestionPage(payload: unknown) {
		if (Array.isArray(payload)) return { items: payload as Question[], total: payload.length };
		if (!isRecord(payload)) return { items: [], total: 0 };
		const data = payload as QuestionListPayload;
		const items = Array.isArray(data.items) ? data.items : [];
		return { items, total: data.meta?.total ?? items.length };
	}

	function parseSubjects(payload: unknown): Subject[] {
		return isRecord(payload) && Array.isArray((payload as AcademicPayload).subjects) ? (payload as AcademicPayload).subjects ?? [] : [];
	}

	async function fetchQuestionsPage(offset: number) {
		const params = new URLSearchParams({ limit: String(questionPageSize), offset: String(offset), status: 'published' });
		params.set('scope', eventId ? 'event_pool' : 'global');
		if (eventId) params.set('event_id', eventId);
		const payload = await fetch(clientApiPathWithQuery('/api/bank-soal/questions', params)).then((response) => readClientApiData<unknown>(response, 'Gagal memuat bank soal'));
		return parseQuestionPage(payload);
	}

	async function fetchAllQuestions() {
		const firstPage = await fetchQuestionsPage(0);
		const questionsById = new SvelteMap(firstPage.items.map((question) => [question.id, question]));
		let loaded = firstPage.items.length;
		let pages = 1;
		while (loaded < firstPage.total && pages < maxQuestionPages) {
			const nextPage = await fetchQuestionsPage(loaded);
			if (nextPage.items.length === 0) break;
			for (const question of nextPage.items) questionsById.set(question.id, question);
			loaded += nextPage.items.length;
			pages += 1;
		}
		questionPoolTotal = Math.max(firstPage.total, questionsById.size);
		return Array.from(questionsById.values());
	}

	async function fetchEventContext() {
		if (!eventId) return null;
		try {
			return await fetch(clientApiPath`/api/asesmen/events/${eventId}`).then((response) => readClientApiData<EventContext>(response, 'Gagal memuat konteks kegiatan'));
		} catch {
			return null;
		}
	}

	async function fetchQuestionCompleteness() {
		if (!eventId) return null;
		try {
			return await fetch(clientApiPath`/api/asesmen/events/${eventId}/question-completeness`).then((response) => readClientApiData<QuestionCompleteness>(response, 'Gagal memuat kelengkapan soal'));
		} catch {
			return null;
		}
	}

	async function fetchFormData(): Promise<FormData> {
		const [questionItems, academicPayload, context, completeness] = await Promise.all([
			fetchAllQuestions(),
			fetch('/api/academic').then((response) => readClientApiData<unknown>(response, 'Gagal memuat data akademik')),
			fetchEventContext(),
			fetchQuestionCompleteness()
		]);
		return { allQuestions: questionItems, subjects: parseSubjects(academicPayload), eventContext: context, completeness };
	}

	function loadForm() {
		allQuestions = [];
		subjects = [];
		questionPoolTotal = 0;
		questionCompleteness = null;
		formPromise = fetchFormData().then((data) => {
			allQuestions = data.allQuestions;
			subjects = data.subjects;
			eventContext = data.eventContext;
			questionCompleteness = data.completeness;
			return data;
		});
	}

	function retryForm(reset?: () => void) {
		reset?.();
		loadForm();
	}

	function errorMessage(error: unknown, fallback: string) {
		if (error instanceof Error && error.message.trim() && !error.message.toLowerCase().includes('fetch')) return error.message;
		return fallback;
	}

	function handleRenderError(error: unknown) {
		console.error('CBT package create render failed', error);
	}

	function toggleQuestion(id: string) {
		if (fSelectedIds.has(id)) {
			fSelectedIds.delete(id);
			fQuestionWeights.delete(id);
			return;
		}
		fSelectedIds.add(id);
		if (!fQuestionWeights.has(id)) fQuestionWeights.set(id, 1);
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
			essay: 'Essay'
		};
		const normalized = compactValue(value, '');
		return labels[normalized] ?? (normalized ? normalized.replaceAll('_', ' ') : 'Belum tipe');
	}

	function difficultyLabel(value: string | null | undefined) {
		const labels: Record<string, string> = { easy: 'Mudah', medium: 'Sedang', hard: 'Sulit' };
		const normalized = compactValue(value, '');
		return labels[normalized] ?? (normalized || '-');
	}

	function questionHasBlueprintGap(question: { cp_ref?: string; tp_ref?: string; kd_ref?: string; cognitive_level?: string }) {
		return !compactValue(question.cp_ref, '') || (!compactValue(question.tp_ref, '') && !compactValue(question.kd_ref, '')) || !compactValue(question.cognitive_level, '');
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

	function countByLabel<T>(items: T[], selector: (item: T) => string): BlueprintBucket[] {
		const counts = new SvelteMap<string, number>();
		for (const item of items) counts.set(selector(item), (counts.get(selector(item)) ?? 0) + 1);
		return Array.from(counts.entries()).map(([label, count]) => ({ label, count })).sort((a, b) => b.count - a.count || a.label.localeCompare(b.label));
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
			const row = rows.get(key) ?? { key, cp, tp, kd, topic, cognitive, count: 0, hotsCount: 0, types: [], missing };
			row.count += 1;
			if (question.hots_flag) row.hotsCount += 1;
			const typeLabel = questionTypeLabel(question.question_type);
			if (!row.types.includes(typeLabel)) row.types.push(typeLabel);
			row.missing = row.missing || missing;
			rows.set(key, row);
		}
		return Array.from(rows.values()).sort((a, b) => Number(b.missing) - Number(a.missing) || b.count - a.count || a.cp.localeCompare(b.cp));
	}

	function normalizedScopeId(value: string | null | undefined) {
		return (value ?? '').trim();
	}

	function isQuestionAllowedForPackage(question: Question) {
		const questionEventId = normalizedScopeId(question.event_id);
		if (!eventId) return questionEventId === '';
		return questionEventId === '' || questionEventId === eventId;
	}

	function buildPackageReadinessIssues() {
		const issues: string[] = [];
		if (!fSubjectId) issues.push('Pilih mata pelajaran');
		if (!fTitle.trim()) issues.push('Isi nama paket');
		if (!fDuration || fDuration < 10) issues.push('Durasi minimal 10 menit');
		if (selectedQuestions.length === 0) issues.push('Pilih minimal 1 soal terbit');
		if (fDrawPgCount > 0 && selectedPgCount < fDrawPgCount) issues.push(`PG terpilih ${selectedPgCount}/${fDrawPgCount}`);
		if (fDrawEssayCount > 0 && selectedEssayCount < fDrawEssayCount) issues.push(`Esai terpilih ${selectedEssayCount}/${fDrawEssayCount}`);
		const invalidWeights = selectedQuestions.filter((question) => {
			const weight = questionWeightValue(question.id);
			return weight < 1 || weight > 100;
		});
		if (invalidWeights.length > 0) issues.push('Bobot setiap soal harus 1-100');
		return issues;
	}

	function selectQuestionDraw() {
		if (!fSubjectId) {
			toast.warning('Pilih mata pelajaran dulu');
			return;
		}
		const seed = fRandomSeed.trim() || `${Date.now()}`;
		const pick = (items: Question[], count: number) => seededShuffle(items, seed).slice(0, Math.max(0, count));
		const pg = fDrawPgCount > 0 ? pick(questionPool.filter((question) => question.question_type === 'multiple_choice'), fDrawPgCount) : [];
		const essay = fDrawEssayCount > 0 ? pick(questionPool.filter((question) => question.question_type === 'essay'), fDrawEssayCount) : [];
		const selected = [...pg, ...essay];
		if (selected.length === 0) {
			toast.warning('Isi target draw PG/Esai atau pilih soal manual');
			return;
		}
		fSelectedIds.clear();
		fQuestionWeights.clear();
		for (const question of selected) {
			fSelectedIds.add(question.id);
			fQuestionWeights.set(question.id, 1);
		}
		toast.success(`Preview draw siap: ${pg.length} PG, ${essay.length} esai`);
	}

	function seededShuffle<T>(items: T[], seed: string): T[] {
		let state = Array.from(seed).reduce((sum, char) => sum + char.charCodeAt(0), 0) || 1;
		const out = [...items];
		for (let i = out.length - 1; i > 0; i -= 1) {
			state = (state * 1664525 + 1013904223) % 4294967296;
			const j = state % (i + 1);
			[out[i], out[j]] = [out[j], out[i]];
		}
		return out;
	}

	async function createPackage() {
		if (packageReadinessIssues.length > 0) {
			operationState = { tone: 'warning', title: 'Paket Belum Siap', message: packageReadinessIssues[0] ?? 'Lengkapi paket ujian terlebih dahulu.' };
			return;
		}
		fBusy = true;
		try {
			const res = await fetch('/api/asesmen/packages', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					subject_id: fSubjectId,
					title: fTitle,
					description: fDescription,
					duration_minutes: fDuration,
					randomize_questions: fRandomize,
					randomize_options: fRandomizeOptions,
					source_mode: fSourceMode,
					draw_pg_count: Math.max(0, Number(fDrawPgCount) || 0),
					draw_essay_count: Math.max(0, Number(fDrawEssayCount) || 0),
					random_seed: fRandomSeed.trim(),
					is_active: fActive,
					...(eventId ? { event_id: eventId } : {}),
					question_ids: selectedQuestions.map((question) => question.id),
					question_weights: Object.fromEntries(selectedQuestions.map((question) => [question.id, questionWeightValue(question.id)]))
				})
			});
			await readClientJson<unknown>(res);
			toast.success('Paket ujian berhasil dibuat');
			await goto(eventId ? `${resolve('/asesmen/paket')}?event_id=${encodeURIComponent(eventId)}` : resolve('/asesmen/paket'));
		} catch (error) {
			toast.error(errorMessage(error, 'Gagal membuat paket. Periksa koneksi lalu coba lagi.'));
		} finally {
			fBusy = false;
		}
	}

	onMount(loadForm);
</script>

<svelte:head><title>{eventId ? 'Buat Paket Event CBT' : 'Buat Paket CBT'} — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6">
	<section class="rounded-3xl border border-primary/20 bg-gradient-to-br from-primary/10 via-card to-primary/10 p-6 shadow-sm">
		<div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
			<div class="max-w-3xl space-y-2">
				<p class="text-xs font-semibold uppercase tracking-[0.2em] text-primary">Keranjang Soal CBT</p>
				<h1 class="text-3xl font-semibold tracking-tight text-foreground">{eventId ? 'Buat Paket Event' : 'Buat Paket Soal'}</h1>
				<p class="text-sm leading-6 text-muted-foreground">Pilih soal terbit dari Bank Soal, atur bobot, lalu simpan paket untuk dipakai saat membuat sesi.</p>
			</div>
			<div class="flex flex-wrap gap-2">
				{#if eventId}
					<Button href={resolve(`/asesmen/kegiatan/${eventId}`)} variant="outline">Kembali ke Event</Button>
				{/if}
				<Button href={listHref} variant="outline">Batal</Button>
			</div>
		</div>
	</section>

	{#if eventId}
		<div class="rounded-xl border border-success/20 bg-success/10 p-4 text-sm text-success">
			<p class="font-semibold">Paket untuk kegiatan: {eventContext?.title ?? eventId}</p>
			<p class="mt-1 text-success">Payload pembuatan paket membawa <code class="rounded bg-card px-1">event_id</code>. Pool soal memakai soal reusable dan soal khusus kegiatan ini.</p>
		</div>
	{/if}

	{#if operationState}
		<OperationStatusPanel {...operationState} />
	{/if}

	<AsyncContent promise={formPromise} onerror={handleRenderError}>
		{#snippet pending()}
			<Card.Root class="border-border shadow-sm">
				<Card.Content class="space-y-3 p-6">
					<Skeleton class="h-10" />
					<Skeleton class="h-10" />
					<Skeleton class="h-48" />
				</Card.Content>
			</Card.Root>
		{/snippet}

		{#snippet failed(error, reset)}
			<RecoveryPanel title="Builder Paket Belum Siap" message={errorMessage(error, 'Gagal memuat data paket')} onRetry={() => retryForm(reset)} />
		{/snippet}

		{#snippet children()}
			<Card.Root class="border-primary/20 shadow-sm">
				<Card.Header>
					<Card.Title class="text-base">Builder Paket</Card.Title>
					<Card.Description>Form create-only ini memakai endpoint <code>/api/asesmen/packages</code> dan filter soal yang sama dengan builder lama.</Card.Description>
				</Card.Header>
				<Card.Content class="space-y-4">
					<div class="grid gap-3 sm:grid-cols-2">
						<div>
							<label for="package-subject-id" class="mb-1 block text-xs text-muted-foreground">Mata Pelajaran <span class="text-destructive">*</span></label>
							<select id="package-subject-id" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fSubjectId} onchange={handleSubjectChange}>
								<option value="">-- Pilih --</option>
								{#each subjects as subject (subject.id)}
									<option value={subject.id}>{subject.code} — {subject.name}</option>
								{/each}
							</select>
						</div>
						<div>
							<label for="package-title" class="mb-1 block text-xs text-muted-foreground">Nama Paket <span class="text-destructive">*</span></label>
							<Input id="package-title" placeholder="mis: UTS Matematika Sem 1 2025" bind:value={fTitle} />
						</div>
						<div>
							<label for="package-duration" class="mb-1 block text-xs text-muted-foreground">Durasi (menit) <span class="text-destructive">*</span></label>
							<Input id="package-duration" type="number" min={10} max={300} bind:value={fDuration} />
						</div>
						<div class="flex items-end gap-4 pb-1">
							<label class="flex items-center gap-2 text-sm"><input type="checkbox" bind:checked={fRandomize} class="rounded" /> Acak urutan soal</label>
							<label class="flex items-center gap-2 text-sm"><input type="checkbox" bind:checked={fRandomizeOptions} class="rounded" /> Acak opsi jawaban</label>
							<label class="flex items-center gap-2 text-sm"><input type="checkbox" bind:checked={fActive} class="rounded" /> Paket aktif</label>
						</div>
					</div>

					<div class="rounded-xl border border-border bg-muted/30 p-3">
						<div class="grid gap-3 md:grid-cols-5">
							<label class="space-y-1 text-xs text-muted-foreground">Mode sumber paket
								<select bind:value={fSourceMode} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground">
									<option value="teacher_class">Guru pengampu rombel</option>
									<option value="level_subject_teachers">Semua guru mapel tingkat</option>
									<option value="event_pool">Pool kegiatan</option>
								</select>
							</label>
							<label class="space-y-1 text-xs text-muted-foreground">Draw PG
								<Input type="number" min={0} bind:value={fDrawPgCount} />
							</label>
							<label class="space-y-1 text-xs text-muted-foreground">Draw esai
								<Input type="number" min={0} bind:value={fDrawEssayCount} />
							</label>
							<label class="space-y-1 text-xs text-muted-foreground">Seed opsional
								<Input placeholder="mis: VIIA-MTK-2026" bind:value={fRandomSeed} />
							</label>
							<div class="flex items-end"><Button type="button" variant="outline" onclick={selectQuestionDraw}>Preview Draw</Button></div>
						</div>
						<p class="mt-2 text-xs text-muted-foreground">Preview draw hanya memilih soal ke keranjang; sumber soal tidak dimutasi. Komposisi tersimpan di audit/log paket saat disimpan.</p>
					</div>

					<div>
						<label for="package-description" class="mb-1 block text-xs text-muted-foreground">Deskripsi (opsional)</label>
						<Textarea id="package-description" placeholder="Keterangan paket ujian..." rows={2} bind:value={fDescription} />
					</div>

					{#if fSubjectId}
						<div>
							<div class="mb-2 flex flex-wrap items-center justify-between gap-2">
								<div class="text-xs text-muted-foreground">Pilih Soal dari Bank Soal ({questionPool.length} soal terbit sesuai cakupan){#if selectedQuestions.length > 0} — <span class="font-medium text-success">{selectedQuestions.length} dipilih</span>{/if}</div>
								<details class="rounded-md border border-border bg-muted/50 px-3 py-2 text-xs text-foreground">
									<summary class="cursor-pointer font-medium">Info pool soal</summary>
									<div class="mt-2 space-y-2 leading-5">
										<p>{questionPool.length} soal terbit tersedia, {availableBlueprintMissingCount} metadata kurang, {availableHotsCount} HOTS.</p>
										{#if questionPoolCapped}<p>Pool soal dibatasi: termuat {allQuestions.length} dari {questionPoolTotal} soal terbit.</p>{/if}
										{#if availableTypeBuckets.length > 0}<p>Bentuk soal: {availableTypeBuckets.map((bucket) => `${bucket.label}: ${bucket.count}`).join(', ')}</p>{/if}
									</div>
								</details>
							</div>
							{#if hiddenScopedQuestionCount > 0}
								<div class="mb-2 rounded-md border border-accent bg-accent/60 px-3 py-2 text-xs text-accent-foreground">{hiddenScopedQuestionCount} soal terbit disembunyikan karena {eventId ? 'tertaut ke kegiatan lain' : 'khusus kegiatan tertentu'}.</div>
							{/if}
							{#if eventSubjectMissingRows.length > 0}
								<div class="mb-2 rounded-md border border-warning/30 bg-warning/10 px-3 py-2 text-xs text-warning"><span class="font-semibold">Warning kelengkapan:</span> {eventSubjectMissingRows.length} baris mapel ini masih kurang dari target. Paket tetap boleh dibuat setelah preview sumber soal diverifikasi.</div>
							{/if}
							{#if eventSubjectCompletenessRows.length > 0}
								<details class="mb-2 rounded-md border border-border bg-muted/40 px-3 py-2 text-xs text-foreground">
									<summary class="cursor-pointer font-medium">Readiness dari Kelengkapan Soal ({eventSubjectCompletenessRows.length} baris)</summary>
									<div class="mt-2 space-y-1">
										{#each eventSubjectCompletenessRows.slice(0, 8) as row (`${row.level}-${row.subject_id}-${row.scope_mode}`)}
											<p>{row.level} · {row.subject_name}: PG {row.available_pg}/{row.target_pg}, Esai {row.available_essay}/{row.target_essay} {row.complete ? '✓' : 'kurang'}</p>
										{/each}
										{#if eventSubjectContributions.length > 0}<p>Kontribusi pool: {eventSubjectContributions.map((row) => `${row.teacher_name}: PG ${row.available_pg}, Esai ${row.available_essay}`).join(' · ')}</p>{/if}
									</div>
								</details>
							{/if}
							{#if questionPool.length === 0}
								<p class="rounded-md border py-4 text-center text-sm text-muted-foreground">Belum ada soal berstatus "Terbit" untuk mata pelajaran ini dalam cakupan paket ini.</p>
							{:else}
								<div class="max-h-72 overflow-y-auto rounded-md border">
									{#each questionPool as question (question.id)}
										<label class="flex cursor-pointer items-start gap-3 border-b px-3 py-2 last:border-b-0 hover:bg-muted/50">
											<input type="checkbox" checked={fSelectedIds.has(question.id)} onchange={() => toggleQuestion(question.id)} class="mt-0.5 rounded" />
											<div class="min-w-0 flex-1">
												<p class="line-clamp-1 text-sm">{question.question_text}</p>
												<div class="mt-0.5 flex flex-wrap gap-1">
													{#if question.code}<span class="font-mono text-xs text-muted-foreground">{question.code}</span>{/if}
													<Badge variant="outline" class="py-0 text-xs">{questionTypeLabel(question.question_type)}</Badge>
													<Badge variant="outline" class="py-0 text-xs">{difficultyLabel(question.difficulty)}</Badge>
													<Badge class="border-success/20 bg-success/10 py-0 text-xs text-success">{questionReviewLabel(question)}</Badge>
													{#each questionReadinessIssues(question) as issue (`${question.id}-${issue}`)}<Badge class="border-warning/30 bg-warning/10 py-0 text-xs text-warning">{issue}</Badge>{/each}
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
											<p class="mt-1 text-xs text-primary">{selectedQuestions.length} soal dipilih dengan bobot total {selectedWeightTotal}.</p>
										</div>
										<span class="rounded-full border border-primary/20 bg-card px-2 py-1 text-xs font-medium text-primary">{selectedBlueprintMissingCount > 0 ? `${selectedBlueprintMissingCount} metadata kurang` : 'Metadata siap'}</span>
									</div>
									<div class="mt-3 overflow-hidden rounded-md border border-primary/20 bg-card">
										<div class="grid grid-cols-[1fr_5rem] gap-2 border-b bg-primary/10 px-2 py-1.5 text-[11px] font-semibold uppercase tracking-[0.12em] text-primary"><span>Soal Terpilih</span><span class="text-right">Bobot</span></div>
										<div class="max-h-40 overflow-y-auto">
											{#each selectedQuestions as question (question.id)}
												<div class="grid grid-cols-[1fr_5rem] items-center gap-2 border-b px-2 py-1.5 last:border-b-0">
													<div class="min-w-0"><p class="truncate text-xs font-medium text-foreground">{question.code || 'Tanpa kode'} · {question.question_text}</p><p class="text-[11px] text-muted-foreground">{questionTypeLabel(question.question_type)} · {difficultyLabel(question.difficulty)}</p></div>
													<Input aria-label={`Bobot ${question.code || question.question_text}`} class="h-8 text-right text-xs" type="number" min={1} max={100} value={questionWeightValue(question.id)} oninput={(event) => setQuestionWeight(question.id, Number(event.currentTarget.value))} />
												</div>
											{/each}
										</div>
									</div>
									<details class="mt-3 rounded-md border border-primary/20 bg-card px-3 py-2 text-xs text-foreground">
										<summary class="cursor-pointer font-semibold text-foreground">Ringkasan blueprint dan mutu</summary>
										<div class="mt-3 space-y-2">
											<p>Bentuk soal: {selectedTypeBuckets.map((bucket) => `${bucket.label}: ${bucket.count}`).join(', ') || '-'}</p>
											<p>Level kognitif: {selectedCognitiveBuckets.map((bucket) => `${bucket.label}: ${bucket.count}`).join(', ') || '-'}</p>
											<div class="max-h-40 overflow-y-auto rounded-md border border-primary/20 bg-card">
												{#each selectedBlueprintMatrix as row (row.key)}
													<div class="grid grid-cols-[1fr_1fr_0.6fr] gap-2 border-b px-2 py-1.5 text-xs last:border-b-0 {row.missing ? 'bg-warning/10' : ''}"><span class="truncate" title={row.cp}>{row.cp}</span><span class="truncate" title={row.cognitive}>{row.cognitive}</span><span class="text-right font-semibold">{row.count}</span></div>
												{/each}
											</div>
										</div>
									</details>
								</div>
							{/if}
							{#if packageReadinessIssues.length > 0}
								<div class="mt-3 rounded-md border border-warning/30 bg-warning/10 px-3 py-2 text-xs text-warning"><span class="font-semibold">Belum siap dibuat:</span> {packageReadinessIssues.join(', ')}</div>
							{:else}
								<div class="mt-3 rounded-md border border-success/20 bg-success/10 px-3 py-2 text-xs font-medium text-success">Paket siap dibuat dengan soal terbit yang sudah terpilih.</div>
							{/if}
						</div>
					{/if}

					<div class="flex flex-wrap gap-2">
						<LoadingButton disabled={!canCreatePackage} onclick={() => void createPackage()} loading={fBusy} loadingLabel="Menyimpan...">{`Buat Paket${selectedQuestions.length > 0 ? ` (${selectedQuestions.length} soal)` : ''}`}</LoadingButton>
						<Button href={listHref} variant="outline">Batal</Button>
					</div>
				</Card.Content>
			</Card.Root>
		{/snippet}
	</AsyncContent>
</div>
