<script lang="ts">
	import { onMount } from 'svelte';
	import * as Dialog from '$lib/components/ui/dialog';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { toast } from '$lib/components/ui/sonner';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import LegacyRichTextEditor from '$lib/components/LegacyRichTextEditor.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import RichContent from '$lib/components/RichContent.svelte';
	import { confirmAction } from '$lib/confirm-dialog';
	import { clientApiPath, clientApiPathWithQuery, readClientApiData, readClientJson } from '$lib/client/api';
	import { htmlToPlainText } from '$lib/utils/html-text';

	// ── Types ─────────────────────────────────────────────────────────────────
	type Subject = { id: string; name: string; code: string };
	type OptionItem = { label: string; text?: string; html?: string; latex?: string };
	type ModuleMode = 'catalog' | 'composer' | 'advanced' | 'review' | 'import';
	type Question = {
		id: string;
		authoring_mode?: string;
		suggested_mode?: string;
		subject_id: string;
		subject_name: string;
		subject_code: string;
		code: string;
		question_type: string;
		stem_html: string;
		question_text: string;
		options: OptionItem[];
		answer_key: string;
		difficulty: string;
		status: string;
		workflow_status: string;
		author_username: string;
		created_at: string;
		package_count?: number;
		answer_count?: number;
		is_locked?: boolean;
		usage?: {
			package_count?: number;
			answer_count?: number;
			is_locked?: boolean;
		};
	};
	type QuestionListResponse = {
		items: Question[];
		meta?: { total: number; limit: number; offset: number };
	};
	type SoalOverview = {
		questions: Question[];
		subjects: Subject[];
		totalItems: number;
		page: number;
	};
	type AcademicPayload = {
		subjects?: Subject[];
		error?: string;
		message?: string;
	};
	type DraftPayload = {
		subjectId: string;
		stem: string;
		options: string[];
		answerKey: string;
		weight: number;
		difficulty: string;
		isRtl: boolean;
		savedAt: string;
	};
	type Template = {
		id: string;
		label: string;
		desc: string;
		stem: string;
		options: string[];
		answerKey: 'A' | 'B' | 'C' | 'D';
		isRtl?: boolean;
	};
	type LegacyImportResult = {
		total_rows: number;
		imported: number;
		skipped: number;
		errors: string[];
		duplicate_codes: string[];
	};

	// ── Constants ─────────────────────────────────────────────────────────────
	const PAGE_SIZE = 15;
	const ANSWER_LABELS = ['A', 'B', 'C', 'D'] as const;
	const moduleModes: Array<{ id: ModuleMode; label: string; desc: string }> = [
		{ id: 'catalog', label: 'Katalog', desc: 'Daftar terpadu' },
		{ id: 'composer', label: 'Komposer Cepat', desc: 'Input PG legacy' },
		{ id: 'advanced', label: 'Editor Lanjutan', desc: 'Blueprint & asset' },
		{ id: 'review', label: 'Review', desc: 'Mutu & publikasi' },
		{ id: 'import', label: 'Import Legacy', desc: 'CSV lama' },
	];
	const WORKFLOW_LABEL: Record<string, string> = {
		draft: 'Draft',
		review: 'Ditinjau',
		approved: 'Disetujui',
		rejected: 'Revisi',
	};
	const DIFFICULTY_LABEL: Record<string, string> = { easy: 'Mudah', medium: 'Sedang', hard: 'Sulit' };

	const DRAFT_KEY = (id: string | null) => `mtsn2-soal-komposer:${id ?? 'new'}`;

	const templates: Template[] = [
		{
			id: 'standar',
			label: 'Soal Standar',
			desc: 'PG biasa',
			stem: '<p>Perhatikan pernyataan berikut!</p><p><br></p><p>Berdasarkan pernyataan di atas, yang paling tepat adalah...</p>',
			options: ['', '', '', ''],
			answerKey: 'A',
		},
		{
			id: 'stimulus',
			label: 'Dengan Stimulus',
			desc: 'Teks + pertanyaan',
			stem: '<p><strong>Bacalah teks berikut dengan saksama!</strong></p><p><br></p><p>[Teks/wacana/stimulus di sini]</p><p><br></p><p>Berdasarkan teks di atas, apa yang dimaksud dengan...?</p>',
			options: ['', '', '', ''],
			answerKey: 'A',
		},
		{
			id: 'matematika',
			label: 'Matematika',
			desc: 'Rumus LaTeX',
			stem: '<p>Diketahui $a = 3$ dan $b = 4$. Tentukan nilai dari:</p><p>$$a^2 + b^2 = \\ldots$$</p>',
			options: ['25', '12', '7', '49'],
			answerKey: 'A',
		},
		{
			id: 'cerita',
			label: 'Soal Cerita',
			desc: 'Narasi + hitung',
			stem: '<p>Pak Ahmad memiliki [jumlah] [barang]. Setelah [kejadian], sisanya menjadi [jumlah baru]. Berapa [yang ditanya]?</p>',
			options: ['', '', '', ''],
			answerKey: 'A',
		},
		{
			id: 'grafik',
			label: 'Grafik / Data',
			desc: 'Interpretasi data',
			stem: '<p>Perhatikan grafik/tabel berikut!</p><p><br></p><p>[Sisipkan gambar grafik di sini menggunakan tombol 🖼]</p><p><br></p><p>Berdasarkan data tersebut, kesimpulan yang paling tepat adalah...</p>',
			options: ['', '', '', ''],
			answerKey: 'A',
		},
		{
			id: 'arab-mufradat',
			label: 'Arab: Mufradat',
			desc: 'Kosakata Arab (RTL)',
			stem: '<p>ما معنى الكلمة التالية ؟</p><p><br></p><p><strong>[الكلمة]</strong></p>',
			options: ['', '', '', ''],
			answerKey: 'A',
			isRtl: true,
		},
		{
			id: 'arab-qiraah',
			label: 'Arab: Qiraah',
			desc: 'Bacaan Arab (RTL)',
			stem: '<p>اقرأ النص الآتي ثم أجب عن الأسئلة!</p><p><br></p><p>[النص هنا]</p><p><br></p><p>ما الموضوع الرئيسي للنص ؟</p>',
			options: ['', '', '', ''],
			answerKey: 'A',
			isRtl: true,
		},
		{
			id: 'dalil',
			label: 'Dalil & Tafsir',
			desc: 'Ayat/hadis + analisis',
			stem: '<p>Perhatikan dalil berikut!</p><p><br></p><p><em>[Tulis ayat atau hadis di sini]</em></p><p><br></p><p>Kandungan pokok dari dalil di atas adalah...</p>',
			options: ['', '', '', ''],
			answerKey: 'A',
		},
		{
			id: 'sains',
			label: 'Percobaan Sains',
			desc: 'Data eksperimen',
			stem: '<p>Perhatikan hasil percobaan berikut!</p><p><br></p><p>[Tabel / data percobaan]</p><p><br></p><p>Berdasarkan data tersebut, variabel yang paling berpengaruh adalah...</p>',
			options: ['', '', '', ''],
			answerKey: 'A',
		},
	];

	// ── Page state ─────────────────────────────────────────────────────────────
	let activeMode = $state<ModuleMode>('catalog');
	let questionsPromise = $state<Promise<SoalOverview> | null>(null);
	let questions = $state<Question[]>([]);
	let subjects = $state<Subject[]>([]);
	let totalItems = $state(0);
	let currentPage = $state(1);

	let search = $state('');
	let filterSubject = $state('');
	let filterWorkflow = $state('');

	// ── Composer state ─────────────────────────────────────────────────────────
	let showComposer = $state(false);
	let editingId = $state<string | null>(null);
	let composerBusy = $state(false);
	let draftStatus = $state('');
	let showTemplates = $state(true);
	let composerMobilePanel = $state<'write' | 'preview'>('write');
	let lastDraftSig = '';
	let questionsRequestId = 0;
	let showImport = $state(false);
	let importSubjectId = $state('');
	let importFile = $state<File | null>(null);
	let importBusy = $state(false);
	let importResult = $state<LegacyImportResult | null>(null);
	let duplicateBusyId = $state('');

	// ── Form fields ────────────────────────────────────────────────────────────
	let fSubjectId = $state('');
	let fStem = $state('');
	let fOptions = $state(['', '', '', '']);
	let fAnswerKey = $state<string>('A');
	let fWeight = $state(1);
	let fDifficulty = $state('medium');
	let fIsRtl = $state(false);
	let activeDraftKey = $derived(DRAFT_KEY(editingId));
	let draftSignature = $derived(JSON.stringify({
		fSubjectId,
		fStem,
		opts: fOptions,
		fAnswerKey,
		fWeight,
		fDifficulty,
		fIsRtl,
	}));

	// ── Derived ────────────────────────────────────────────────────────────────
	let pageCount = $derived(Math.max(1, Math.ceil(totalItems / PAGE_SIZE)));
	let lockedCount = $derived(questions.filter(questionUsageLocked).length);
	let reviewCount = $derived(questions.filter((item) => item.workflow_status === 'review').length);
	let draftCount = $derived(questions.filter((item) => item.workflow_status === 'draft').length);
	let publishedCount = $derived(questions.filter((item) => item.status === 'published').length);

	let stemText = $derived(htmlToPlainText(fStem));
	let hasImage = $derived(fStem.includes('<img'));
	let optionPlainTexts = $derived(fOptions.map((option) => htmlToPlainText(option)));
	let optionHasImages = $derived(fOptions.map((option) => option.includes('<img')));

	let readinessChecks = $derived({
		subject: !!fSubjectId,
		stem: stemText.length >= 5 || hasImage,
		optA: richTextHasContent(fOptions[0]),
		optB: richTextHasContent(fOptions[1]),
		optC: richTextHasContent(fOptions[2]),
		optD: richTextHasContent(fOptions[3]),
		answerKey: !!fAnswerKey,
		weight: Number.isFinite(fWeight) && fWeight >= 1,
	});

	let passedChecks = $derived(Object.values(readinessChecks).filter(Boolean).length);
	let totalChecks = $derived(Object.keys(readinessChecks).length);
	let readinessScore = $derived(Math.round((passedChecks / totalChecks) * 100));
	let canSave = $derived(readinessScore === 100 && !composerBusy);

	let qualitySignals = $derived.by(() => {
		const filled = optionPlainTexts.filter((text, index) => text || optionHasImages[index]);
		const unique = new Set(filled);
		const lengths = filled.map((o) => o.length);
		const lengthRange =
			lengths.length >= 2 ? Math.max(...lengths) - Math.min(...lengths) : 0;
		return [
			{
				label: 'Stem cukup jelas',
				status: stemText.length >= 35 ? 'good' : 'warn',
				desc: `${stemText.length} / 35 karakter minimum`,
			},
			{
				label: 'Distraktor bervariasi',
				status: unique.size === filled.length ? 'good' : 'warn',
				desc: unique.size < filled.length ? 'Ada opsi yang duplikat' : 'Semua opsi unik',
			},
			{
				label: 'Panjang opsi seimbang',
				status: lengthRange <= 80 ? 'good' : 'warn',
				desc: `Selisih panjang: ${lengthRange} karakter`,
			},
			{
				label: 'Media pendukung',
				status: hasImage || stemText.length >= 70 ? 'good' : 'warn',
				desc: hasImage ? 'Ada gambar' : `${stemText.length} / 70 karakter`,
			},
		];
	});

	let validationIssues = $derived.by(() => {
		const issues: string[] = [];
		if (!readinessChecks.subject) issues.push('Pilih mata pelajaran');
		if (!readinessChecks.stem) issues.push('Isi soal minimal 5 karakter');
		if (!readinessChecks.optA || !readinessChecks.optB || !readinessChecks.optC || !readinessChecks.optD)
			issues.push('Semua opsi (A–D) wajib diisi');
		if (!readinessChecks.answerKey) issues.push('Pilih kunci jawaban');
		if (!readinessChecks.weight) issues.push('Bobot nilai minimal 1');
		return issues;
	});

	// ── Draft autosave ─────────────────────────────────────────────────────────
	function buildDraftPayload(): DraftPayload {
		return {
			subjectId: fSubjectId,
			stem: fStem,
			options: [...fOptions],
			answerKey: fAnswerKey,
			weight: fWeight,
			difficulty: fDifficulty,
			isRtl: fIsRtl,
			savedAt: new Date().toISOString(),
		};
	}

	function markDraftAutosaved() {
		draftStatus = 'Draft tersimpan otomatis';
	}

	function saveDraftSnapshot(draftKey: string, signature: string) {
		lastDraftSig = signature;
		try {
			localStorage.setItem(draftKey, JSON.stringify(buildDraftPayload()));
			markDraftAutosaved();
		} catch {
			/* ignore storage errors */
		}
	}

	$effect(() => {
		if (!showComposer) return;
		const sig = draftSignature;
		if (sig === lastDraftSig) return;

		const draftKey = activeDraftKey;
		const timer = setTimeout(() => {
			saveDraftSnapshot(draftKey, sig);
		}, 700);
		return () => {
			clearTimeout(timer);
		};
	});

	function restoreDraft(): boolean {
		try {
			const raw = localStorage.getItem(activeDraftKey);
			if (!raw) return false;
			const d = JSON.parse(raw) as {
				subjectId?: string;
				stem?: string;
				options?: string[];
				answerKey?: string;
				weight?: number;
				difficulty?: string;
				isRtl?: boolean;
			};
			fSubjectId = d.subjectId ?? '';
			fStem = d.stem ?? '';
			fOptions = d.options ?? ['', '', '', ''];
			while (fOptions.length < 4) fOptions = [...fOptions, ''];
			fAnswerKey = d.answerKey ?? 'A';
			fWeight = d.weight ?? 1;
			fDifficulty = d.difficulty ?? 'medium';
			fIsRtl = d.isRtl ?? false;
			draftStatus = 'Draft lokal dipulihkan';
			return true;
		} catch {
			return false;
		}
	}

	function clearDraft() {
		try {
			localStorage.removeItem(activeDraftKey);
		} catch {
			/* ignore */
		}
		lastDraftSig = '';
		draftStatus = '';
	}

	// ── API ────────────────────────────────────────────────────────────────────
	function buildQuestionParams(page: number) {
		const params = new URLSearchParams();
		params.set('limit', String(PAGE_SIZE));
		params.set('offset', String((page - 1) * PAGE_SIZE));
		if (search.trim()) params.set('q', search.trim());
		if (filterSubject) params.set('subject_id', filterSubject);
		if (filterWorkflow) params.set('workflow_status', filterWorkflow);
		return params;
	}

	async function fetchOverview(page = currentPage): Promise<SoalOverview> {
		const params = buildQuestionParams(page);
		const [questionPayload, academicPayload] = await Promise.all([
			fetch(clientApiPathWithQuery('/api/cbt/questions', params)).then((response) =>
				readClientApiData<QuestionListResponse>(response, 'Gagal memuat soal')
			),
			fetch('/api/academic').then((response) =>
				readClientApiData<AcademicPayload>(response, 'Gagal memuat data akademik')
			),
		]);
		const loadedQuestions = questionPayload.items ?? [];
		return {
			questions: loadedQuestions,
			subjects: academicPayload.subjects ?? [],
			totalItems: questionPayload.meta?.total ?? loadedQuestions.length,
			page,
		};
	}

	function applyOverview(overview: SoalOverview) {
		questions = overview.questions;
		subjects = overview.subjects;
		totalItems = overview.totalItems;
		currentPage = overview.page;
	}

	function load(page = currentPage) {
		const requestId = ++questionsRequestId;
		questionsPromise = fetchOverview(page).then((overview) => {
			if (requestId !== questionsRequestId) return { questions, subjects, totalItems, page: currentPage };
			applyOverview(overview);
			return overview;
		}).catch((error: unknown) => {
			if (requestId === questionsRequestId) throw error;
			return { questions, subjects, totalItems, page: currentPage };
		});
	}

	async function refreshOverview(page = currentPage) {
		if (!questionsPromise) {
			load(page);
			return;
		}
		const requestId = ++questionsRequestId;
		try {
			const overview = await fetchOverview(page);
			if (requestId !== questionsRequestId) return;
			applyOverview(overview);
			questionsPromise = Promise.resolve(overview);
		} catch (error) {
			if (requestId === questionsRequestId) {
				questionsPromise = Promise.resolve({ questions, subjects, totalItems, page: currentPage });
				toast.error(soalErrorMessage(error));
			}
		}
	}

	function retryQuestions(reset?: () => void) {
		reset?.();
		load(currentPage);
	}

	function soalErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		if (typeof error === 'string' && error.trim()) return error;
		return 'Gagal memuat soal';
	}

	function handleQuestionsRenderError(error: unknown, reset: () => void) {
		console.error('Question composer render failed', error);
		reset();
	}

	function mutationErrorMessage(error: unknown, fallback: string) {
		if (error instanceof Error && error.message.trim() && !error.message.toLowerCase().includes('fetch')) return error.message;
		return fallback;
	}

	function validModuleMode(value: string | null): value is ModuleMode {
		return moduleModes.some((mode) => mode.id === value);
	}

	function setModuleMode(mode: ModuleMode) {
		activeMode = mode;
		if (typeof window === 'undefined') return;
		const url = new URL(window.location.href);
		url.searchParams.set('mode', mode);
		window.history.replaceState({}, '', `${url.pathname}?${url.searchParams.toString()}`);
	}

	function richTextHasContent(html: string): boolean {
		return htmlToPlainText(html).length > 0 || html.includes('<img');
	}

	function questionUsageLocked(q: Question): boolean {
		const packages = q.package_count ?? q.usage?.package_count ?? 0;
		const answers = q.answer_count ?? q.usage?.answer_count ?? 0;
		return Boolean(q.is_locked ?? q.usage?.is_locked ?? (packages > 0 || answers > 0));
	}

	function questionUsageText(q: Question): string {
		const packages = q.package_count ?? q.usage?.package_count ?? 0;
		const answers = q.answer_count ?? q.usage?.answer_count ?? 0;
		if (packages > 0 && answers > 0) return `${packages} paket, ${answers} jawaban`;
		if (packages > 0) return `${packages} paket`;
		if (answers > 0) return `${answers} jawaban`;
		return 'Belum dipakai';
	}

	function isQuickEditable(q: Question): boolean {
		const mode = q.suggested_mode ?? q.authoring_mode ?? 'beginner';
		return mode !== 'advance' && q.question_type === 'multiple_choice' && q.workflow_status === 'draft' && q.status === 'draft' && !questionUsageLocked(q);
	}

	function explainQuickEditBlocked(q: Question): string {
		if (questionUsageLocked(q)) return 'Soal sudah dipakai. Gunakan Duplikat untuk membuat revisi draft.';
		if ((q.suggested_mode ?? q.authoring_mode) === 'advance') return 'Soal ini memakai metadata lanjutan. Gunakan editor lanjutan agar metadata tidak hilang.';
		if (q.workflow_status !== 'draft' || q.status !== 'draft') return 'Soal sudah masuk alur review/publikasi. Gunakan Duplikat untuk revisi.';
		if (q.question_type !== 'multiple_choice') return 'Komposer cepat hanya untuk pilihan ganda A-D.';
		return 'Gunakan editor lanjutan untuk item ini.';
	}

	function openQuestion(q: Question) {
		if (isQuickEditable(q)) {
			void openEdit(q);
			return;
		}
		setModuleMode('advanced');
		if (typeof window !== 'undefined') {
			window.location.href = `/cbt/questions?question_id=${encodeURIComponent(q.id)}`;
			return;
		}
		toast.warning(explainQuickEditBlocked(q));
	}

	function resetForm() {
		fSubjectId = '';
		fStem = '';
		fOptions = ['', '', '', ''];
		fAnswerKey = 'A';
		fWeight = 1;
		fDifficulty = 'medium';
		fIsRtl = false;
		draftStatus = '';
		lastDraftSig = '';
	}

	function openCreate() {
		editingId = null;
		resetForm();
		showTemplates = true;
		composerMobilePanel = 'write';
		showComposer = true;
		// Delay to let state settle before restoring
		setTimeout(() => {
			if (!restoreDraft()) draftStatus = '';
		}, 50);
	}

	async function openEdit(q: Question) {
		if (!isQuickEditable(q)) {
			setModuleMode('advanced');
			toast.warning(explainQuickEditBlocked(q));
			return;
		}
		composerBusy = true;
		try {
			const res = await fetch(clientApiPath`/api/cbt/questions/${q.id}`);
			const d = await readClientApiData<Question>(res, 'Gagal memuat detail soal');

			editingId = d.id;
			resetForm();
			fSubjectId = d.subject_id ?? '';
			fStem = d.stem_html || d.question_text || '';
			const opts = d.options?.length
				? d.options.map((o) => o.html || o.text || o.latex || '')
				: ['', '', '', ''];
			while (opts.length < 4) opts.push('');
			fOptions = opts;
			fAnswerKey = d.answer_key || 'A';
			fDifficulty = d.difficulty || 'medium';
			showTemplates = false;
			composerMobilePanel = 'write';
			showComposer = true;
		} catch (error) {
			toast.error(mutationErrorMessage(error, 'Gagal memuat detail soal. Form memakai data ringkas dari daftar.'));
			editingId = q.id;
			resetForm();
			fSubjectId = q.subject_id;
			fStem = q.stem_html || q.question_text || '';
			fAnswerKey = q.answer_key || 'A';
			composerMobilePanel = 'write';
			showComposer = true;
		} finally {
			composerBusy = false;
		}
	}

	function applyTemplate(t: Template) {
		fStem = t.stem;
		fOptions = [...t.options];
		while (fOptions.length < 4) fOptions = [...fOptions, ''];
		fAnswerKey = t.answerKey;
		fIsRtl = t.isRtl ?? false;
		draftStatus = `Template "${t.label}" diterapkan`;
		showTemplates = false;
	}

	function closeComposer() {
		showComposer = false;
		editingId = null;
	}

	function openImport() {
		importSubjectId = filterSubject || fSubjectId || '';
		importFile = null;
		importResult = null;
		showImport = true;
	}

	function onImportFileChange(event: Event) {
		const input = event.target as HTMLInputElement;
		importFile = input.files?.[0] ?? null;
		importResult = null;
	}

	async function saveQuestion() {
		if (!canSave) return;
		if (editingId) {
			const current = questions.find((item) => item.id === editingId);
			if (current && !isQuickEditable(current)) {
				toast.error('Soal ini tidak aman diedit lewat komposer cepat. Gunakan Duplikat untuk revisi.');
				return;
			}
		}
		composerBusy = true;
		try {
			const payload = {
				authoring_mode: 'beginner',
				subject_id: fSubjectId,
				question_text: htmlToPlainText(fStem),
				question_type: 'multiple_choice',
				stem_html: fStem,
				options: fOptions.map((html, i) => ({
					label: ANSWER_LABELS[i],
					text: htmlToPlainText(html),
					html,
				})),
				answer_key: fAnswerKey,
				difficulty: fDifficulty,
				status: 'draft',
				workflow_status: 'draft',
			};

			const url = editingId ? clientApiPath`/api/cbt/questions/${editingId}` : '/api/cbt/questions';
			const method = editingId ? 'PUT' : 'POST';
			const res = await fetch(url, {
				method,
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(payload),
			});
			await readClientJson<unknown>(res);

			clearDraft();
			toast.success(editingId ? 'Soal berhasil diperbarui' : 'Soal berhasil dibuat');
			closeComposer();
			await refreshOverview(1);
		} catch (e) {
			toast.error(mutationErrorMessage(e, 'Gagal menyimpan soal'));
		} finally {
			composerBusy = false;
		}
	}

	async function importLegacyCSV() {
		if (!importSubjectId) {
			toast.error('Pilih mata pelajaran untuk import');
			return;
		}
		if (!importFile) {
			toast.error('Pilih file CSV terlebih dahulu');
			return;
		}
		importBusy = true;
		try {
			const form = new FormData();
			form.set('subject_id', importSubjectId);
			form.set('file', importFile);
			const res = await fetch('/api/cbt/questions/import-legacy', {
				method: 'POST',
				body: form,
			});
			const result = await readClientApiData<LegacyImportResult>(res, 'Import CSV gagal');
			importResult = {
				total_rows: result.total_rows ?? 0,
				imported: result.imported ?? 0,
				skipped: result.skipped ?? 0,
				errors: result.errors ?? [],
				duplicate_codes: result.duplicate_codes ?? [],
			};
			toast.success(`${importResult.imported} soal berhasil diimport`);
			await refreshOverview(1);
		} catch (error) {
			toast.error(mutationErrorMessage(error, 'Import CSV gagal'));
		} finally {
			importBusy = false;
		}
	}

	async function duplicateQuestion(id: string) {
		duplicateBusyId = id;
		try {
			const res = await fetch(clientApiPath`/api/cbt/questions/${id}/duplicate`, { method: 'POST' });
			await readClientJson<unknown>(res);
			toast.success('Soal diduplikasi sebagai draft');
			await refreshOverview(currentPage);
		} catch (e) {
			toast.error(mutationErrorMessage(e, 'Gagal menduplikasi soal'));
		} finally {
			duplicateBusyId = '';
		}
	}

	async function deleteQuestion(id: string) {
		const current = questions.find((item) => item.id === id);
		if (current && questionUsageLocked(current)) {
			toast.error('Soal yang sudah dipakai tidak bisa dihapus langsung. Duplikat untuk revisi atau arsipkan lewat alur review.');
			return;
		}
		if (!(await confirmAction({
			title: 'Hapus Soal',
			message: 'Hapus soal ini? Tindakan tidak bisa dibatalkan dari layar operator.',
			confirmLabel: 'Hapus Soal',
			tone: 'danger'
	}))) return;
		try {
			const res = await fetch(clientApiPathWithQuery('/api/cbt/questions', new URLSearchParams({ id })), { method: 'DELETE' });
			await readClientJson<unknown>(res);
			toast.success('Soal dihapus');
			await refreshOverview(currentPage);
		} catch (e) {
			toast.error(mutationErrorMessage(e, 'Gagal menghapus soal'));
		}
	}

	async function uploadImageInEditor(file: File): Promise<string> {
		const form = new FormData();
		form.set('file', file);
		form.set('purpose', 'general');
		if (editingId) form.set('question_id', editingId);
		const res = await fetch('/api/cbt/assets', { method: 'POST', body: form });
		const payload = await readClientApiData<{ url?: string }>(res, 'Upload gambar gagal');
		return payload.url ?? '';
	}

	let searchTimer: ReturnType<typeof setTimeout> | null = null;
	function onSearchInput(e: Event) {
		search = (e.target as HTMLInputElement).value;
		if (searchTimer) clearTimeout(searchTimer);
		searchTimer = setTimeout(() => load(1), 400);
	}

	function stemPreview(q: Question): string {
		const t = htmlToPlainText(q.stem_html || q.question_text || '');
		return t.length > 90 ? t.slice(0, 90) + '…' : t || '(kosong)';
	}

	function workflowClass(status: string): string {
		const map: Record<string, string> = {
			draft: 'bg-slate-100 text-slate-600',
			review: 'bg-yellow-100 text-yellow-700',
			approved: 'bg-green-100 text-green-700',
			rejected: 'bg-red-100 text-red-600',
		};
		return map[status] ?? 'bg-slate-100 text-slate-500';
	}

	onMount(() => {
		const params = new URLSearchParams(window.location.search);
		const mode = params.get('mode');
		if (validModuleMode(mode)) activeMode = mode;
		load();
	});
</script>

<!-- ── Main page ──────────────────────────────────────────────────────────── -->
<div class="space-y-4">
	<!-- Header -->
	<div class="flex flex-wrap items-start justify-between gap-3">
		<div>
			<h1 class="text-xl font-semibold text-slate-800">Bank Soal CBT</h1>
			<p class="text-sm text-slate-500 mt-0.5">
				Satu modul untuk katalog, komposer cepat, editor lanjutan, review, dan import legacy
			</p>
		</div>
		<div class="flex shrink-0 flex-wrap gap-2">
			<Button variant="outline" onclick={() => setModuleMode('import')}>
				Import CSV
			</Button>
			<Button onclick={openCreate} class="bg-green-700 hover:bg-green-800 text-white">
				+ Buat Soal
			</Button>
		</div>
	</div>

	<div class="grid gap-2 md:grid-cols-5">
		{#each moduleModes as mode (mode.id)}
			<button
				type="button"
				onclick={() => setModuleMode(mode.id)}
				class="rounded-lg border px-3 py-3 text-left transition-colors {activeMode === mode.id
					? 'border-green-500 bg-green-50 text-green-900'
					: 'border-slate-200 bg-white text-slate-600 hover:border-green-200 hover:bg-green-50/40'}"
			>
				<div class="text-xs font-bold uppercase tracking-wider">{mode.label}</div>
				<div class="mt-1 text-[11px] text-slate-500">{mode.desc}</div>
			</button>
		{/each}
	</div>

	{#if activeMode === 'composer'}
		<section class="rounded-lg border border-green-200 bg-green-50 p-4">
			<div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
				<div>
					<h2 class="text-sm font-bold uppercase tracking-wider text-green-900">Komposer Cepat PG Legacy</h2>
					<p class="mt-1 text-sm text-green-800">Dipakai untuk membuat draft pilihan ganda A-D. Soal yang sudah review/publish atau dipakai paket wajib direvisi lewat duplikasi.</p>
				</div>
				<Button onclick={openCreate} class="bg-green-700 text-white hover:bg-green-800">Buka Komposer</Button>
			</div>
		</section>
	{:else if activeMode === 'advanced'}
		<section class="rounded-lg border border-amber-200 bg-amber-50 p-4">
			<div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
				<div>
					<h2 class="text-sm font-bold uppercase tracking-wider text-amber-900">Editor Lanjutan sebagai Compatibility Bridge</h2>
					<p class="mt-1 text-sm text-amber-800">Gunakan untuk CP/TP/KD, HOTS, stimulus, rubrik, asset, dan tipe soal selain PG. Route lama tetap hidup sebagai jembatan aman pada sprint ini.</p>
				</div>
				<a href="/cbt/questions" class="inline-flex h-9 items-center justify-center rounded-md border border-amber-300 bg-white px-4 text-sm font-semibold text-amber-900 hover:bg-amber-100">
					Buka Editor Lanjutan
				</a>
			</div>
		</section>
	{:else if activeMode === 'review'}
		<section class="grid gap-3 md:grid-cols-4">
			<div class="rounded-lg border border-slate-200 bg-white p-4">
				<p class="text-xs uppercase tracking-wider text-slate-500">Draft</p>
				<p class="mt-2 text-2xl font-bold text-slate-900">{draftCount}</p>
			</div>
			<div class="rounded-lg border border-yellow-200 bg-yellow-50 p-4">
				<p class="text-xs uppercase tracking-wider text-yellow-700">Menunggu Review</p>
				<p class="mt-2 text-2xl font-bold text-yellow-900">{reviewCount}</p>
			</div>
			<div class="rounded-lg border border-green-200 bg-green-50 p-4">
				<p class="text-xs uppercase tracking-wider text-green-700">Terbit</p>
				<p class="mt-2 text-2xl font-bold text-green-900">{publishedCount}</p>
			</div>
			<div class="rounded-lg border border-red-200 bg-red-50 p-4">
				<p class="text-xs uppercase tracking-wider text-red-700">Terkunci</p>
				<p class="mt-2 text-2xl font-bold text-red-900">{lockedCount}</p>
			</div>
		</section>
	{:else if activeMode === 'import'}
		<section class="rounded-lg border border-slate-200 bg-white p-4">
			<div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
				<div>
					<h2 class="text-sm font-bold uppercase tracking-wider text-slate-800">Import CSV Bank Soal Legacy</h2>
					<p class="mt-1 text-sm text-slate-500">Import masuk sebagai draft, memakai kontrak Go API, dan akan menolak duplikat dalam file maupun duplikat stem di bank soal.</p>
				</div>
				<Button variant="outline" onclick={openImport}>Pilih CSV</Button>
			</div>
		</section>
	{/if}

	<!-- Filter bar -->
	<div class="flex flex-wrap gap-2">
		<Input
			placeholder="Cari soal..."
			value={search}
			oninput={onSearchInput}
			class="w-48 text-sm h-8"
		/>
		<select
			bind:value={filterSubject}
			onchange={() => load(1)}
			class="h-8 rounded-md border border-slate-200 bg-white px-2.5 text-sm text-slate-700 focus:outline-none focus:ring-2 focus:ring-green-500"
		>
			<option value="">Semua Mapel</option>
			{#each subjects as s (s.id)}
				<option value={s.id}>{s.name}</option>
			{/each}
		</select>
		<select
			bind:value={filterWorkflow}
			onchange={() => load(1)}
			class="h-8 rounded-md border border-slate-200 bg-white px-2.5 text-sm text-slate-700 focus:outline-none focus:ring-2 focus:ring-green-500"
		>
			<option value="">Semua Status</option>
			<option value="draft">Draft</option>
			<option value="review">Ditinjau</option>
			<option value="approved">Disetujui</option>
			<option value="rejected">Revisi</option>
		</select>
		{#if totalItems > 0}
			<span class="flex items-center text-xs text-slate-400">{totalItems} soal</span>
		{/if}
	</div>

	<!-- Table -->
	<div class="rounded-lg border border-slate-200 bg-white overflow-hidden">
		<Table.Root>
			<Table.Header>
				<Table.Row class="bg-slate-50 text-xs">
					<Table.Head class="w-10 text-slate-500">#</Table.Head>
					<Table.Head class="text-slate-500">Isi Soal</Table.Head>
					<Table.Head class="w-32 text-slate-500">Mapel</Table.Head>
					<Table.Head class="w-28 hidden sm:table-cell text-slate-500">Status</Table.Head>
					<Table.Head class="w-16 text-slate-500">Kunci</Table.Head>
					<Table.Head class="w-36 text-right text-slate-500">Aksi</Table.Head>
				</Table.Row>
			</Table.Header>
			<Table.Body>
				<AsyncContent promise={questionsPromise} onerror={handleQuestionsRenderError}>
					{#snippet pending()}
						{#each Array.from({ length: 6 }) as _, index (`composer-question-skeleton-${index}`)}
							<Table.Row>
								<Table.Cell><Skeleton class="h-4 w-4" /></Table.Cell>
								<Table.Cell><Skeleton class="h-5 w-full max-w-sm" /></Table.Cell>
								<Table.Cell><Skeleton class="h-5 w-24" /></Table.Cell>
								<Table.Cell class="hidden sm:table-cell"><Skeleton class="h-6 w-20" /></Table.Cell>
								<Table.Cell><Skeleton class="h-5 w-8" /></Table.Cell>
								<Table.Cell class="text-right"><Skeleton class="ml-auto h-8 w-16" /></Table.Cell>
							</Table.Row>
						{/each}
					{/snippet}
					{#snippet failed(error, reset)}
						<Table.Row>
							<Table.Cell colspan={6} class="py-6">
								<RecoveryPanel
									compact
									title="Soal Belum Tersaji"
									message={soalErrorMessage(error)}
									onRetry={() => retryQuestions(reset)}
								/>
							</Table.Cell>
						</Table.Row>
					{/snippet}
					{#snippet children(value)}
						{@const overview = value as SoalOverview}
						{@const currentQuestions = overview.questions}
						{#if currentQuestions.length === 0}
							<Table.Row>
								<Table.Cell colspan={6} class="py-10 text-center text-sm text-slate-400">
									Belum ada soal.
									<button onclick={openCreate} class="text-green-700 underline ml-1"
										>Buat soal pertama →</button
									>
								</Table.Cell>
							</Table.Row>
						{:else}
							{#each currentQuestions as q, i (q.id)}
								<Table.Row
									class="hover:bg-slate-50 cursor-pointer"
									onclick={() => openQuestion(q)}
								>
									<Table.Cell class="text-xs text-slate-400">
										{(overview.page - 1) * PAGE_SIZE + i + 1}
									</Table.Cell>
									<Table.Cell class="text-sm text-slate-700 max-w-xs">
										<div class="truncate">{stemPreview(q)}</div>
										{#if q.author_username}
											<div class="text-[10px] text-slate-400 mt-0.5">{q.author_username}</div>
										{/if}
										<div class="mt-1 flex flex-wrap gap-1">
											<span class="rounded bg-slate-100 px-1.5 py-0.5 text-[10px] font-medium text-slate-500">
												{q.suggested_mode ?? q.authoring_mode ?? 'beginner'}
											</span>
											{#if questionUsageLocked(q)}
												<span class="rounded bg-red-50 px-1.5 py-0.5 text-[10px] font-medium text-red-600">
													Terkunci: {questionUsageText(q)}
												</span>
											{/if}
										</div>
									</Table.Cell>
									<Table.Cell class="text-xs text-slate-500 truncate max-w-[8rem]">
										{q.subject_name || q.subject_code || '-'}
									</Table.Cell>
									<Table.Cell class="hidden sm:table-cell">
										<span
											class="inline-block rounded px-1.5 py-0.5 text-[10px] font-medium {workflowClass(q.workflow_status)}"
										>
											{WORKFLOW_LABEL[q.workflow_status] ?? q.workflow_status ?? '-'}
										</span>
									</Table.Cell>
									<Table.Cell class="text-sm font-bold text-green-700">
										{q.answer_key || '-'}
									</Table.Cell>
									<Table.Cell class="text-right">
										<button
											onclick={(e) => {
												e.stopPropagation();
												openQuestion(q);
											}}
											class="rounded px-2 py-1 text-xs text-slate-600 transition-colors hover:bg-slate-100"
										>
											{isQuickEditable(q) ? 'Edit' : 'Editor'}
										</button>
										<button
											onclick={(e) => {
												e.stopPropagation();
												duplicateQuestion(q.id);
											}}
											disabled={duplicateBusyId === q.id}
											class="rounded px-2 py-1 text-xs text-green-700 transition-colors hover:bg-green-50 disabled:opacity-50"
										>
											{duplicateBusyId === q.id ? 'Menyalin...' : 'Duplikat'}
										</button>
										<button
											onclick={(e) => {
												e.stopPropagation();
												deleteQuestion(q.id);
											}}
											disabled={questionUsageLocked(q)}
											class="text-xs text-red-400 hover:text-red-600 px-2 py-1 rounded hover:bg-red-50 transition-colors disabled:cursor-not-allowed disabled:opacity-40"
										>
											Hapus
										</button>
									</Table.Cell>
								</Table.Row>
							{/each}
						{/if}
					{/snippet}
				</AsyncContent>
			</Table.Body>
		</Table.Root>
	</div>

	<!-- Pagination -->
	{#if pageCount > 1}
		<div class="flex items-center justify-between text-sm text-slate-500">
			<span class="text-xs">{totalItems} soal total</span>
			<div class="flex items-center gap-1">
				<Button
					variant="outline"
					class="h-7 w-7 p-0 text-xs"
					disabled={currentPage <= 1}
					onclick={() => load(currentPage - 1)}
				>
					‹
				</Button>
				<span class="px-2 text-xs">Hal {currentPage} / {pageCount}</span>
				<Button
					variant="outline"
					class="h-7 w-7 p-0 text-xs"
					disabled={currentPage >= pageCount}
					onclick={() => load(currentPage + 1)}
				>
					›
				</Button>
			</div>
		</div>
	{/if}
</div>

<Dialog.Root bind:open={showImport}>
	<Dialog.Content>
		<div class="w-[min(92vw,34rem)] space-y-4 p-5">
			<div>
				<h2 class="text-base font-semibold text-slate-800">Import CSV Bank Soal Legacy</h2>
				<p class="mt-1 text-xs text-slate-500">Hasil import disimpan sebagai draft di model soal CBT.</p>
			</div>
			<div class="space-y-3">
				<div>
					<label for="legacy-import-subject" class="mb-1 block text-xs font-medium text-slate-600">
						Mata Pelajaran <span class="text-red-500">*</span>
					</label>
					<select
						id="legacy-import-subject"
						bind:value={importSubjectId}
						class="w-full rounded-md border border-slate-200 bg-white px-2.5 py-2 text-sm text-slate-700 focus:outline-none focus:ring-2 focus:ring-green-500"
					>
						<option value="">-- Pilih Mapel --</option>
						{#each subjects as s (s.id)}
							<option value={s.id}>{s.name}</option>
						{/each}
					</select>
				</div>
				<div>
					<label for="legacy-import-file" class="mb-1 block text-xs font-medium text-slate-600">
						File CSV <span class="text-red-500">*</span>
					</label>
					<input
						id="legacy-import-file"
						type="file"
						accept=".csv,text/csv"
						onchange={onImportFileChange}
						class="block w-full rounded-md border border-slate-200 bg-white px-2.5 py-2 text-sm text-slate-700 file:mr-3 file:rounded-md file:border-0 file:bg-green-50 file:px-3 file:py-1.5 file:text-sm file:font-medium file:text-green-800"
					/>
				</div>
			</div>
			{#if importResult}
				<div class="rounded-md border border-green-200 bg-green-50 p-3 text-sm text-green-900">
					<div class="grid grid-cols-3 gap-2 text-center">
						<div>
							<div class="text-lg font-bold">{importResult.imported}</div>
							<div class="text-[10px] uppercase text-green-700">Masuk</div>
						</div>
						<div>
							<div class="text-lg font-bold">{importResult.skipped}</div>
							<div class="text-[10px] uppercase text-green-700">Lewat</div>
						</div>
						<div>
							<div class="text-lg font-bold">{importResult.total_rows}</div>
							<div class="text-[10px] uppercase text-green-700">Baris</div>
						</div>
					</div>
					{#if importResult.errors.length > 0}
						<ul class="mt-3 space-y-1 border-t border-green-200 pt-2 text-xs text-amber-800">
							{#each importResult.errors.slice(0, 6) as err (`legacy-import-error-${err}`)}
								<li>{err}</li>
							{/each}
						</ul>
					{/if}
				</div>
			{/if}
			<div class="flex justify-end gap-2 border-t border-slate-100 pt-4">
				<Button variant="outline" onclick={() => (showImport = false)}>
					Tutup
				</Button>
				<LoadingButton
					onclick={() => void importLegacyCSV()}
					loading={importBusy}
					loadingLabel="Import..."
					disabled={importBusy || !importSubjectId || !importFile}
					class="bg-green-700 text-white hover:bg-green-800 disabled:opacity-50"
				>
					Import
				</LoadingButton>
			</div>
		</div>
	</Dialog.Content>
</Dialog.Root>

{#snippet composerPreview()}
	<div class="mb-4">
		<div class="mb-1 flex items-center justify-between">
			<span class="text-[10px] font-semibold uppercase tracking-wide text-slate-400">Kesiapan Soal</span>
			<span
				class="text-sm font-bold {readinessScore === 100
					? 'text-green-700'
					: readinessScore >= 50
						? 'text-amber-600'
						: 'text-red-500'}"
			>
				{readinessScore}%
			</span>
		</div>
		<div class="h-2 overflow-hidden rounded-full bg-slate-200">
			<div
				class="h-2 rounded-full transition-all duration-300 {readinessScore === 100
					? 'bg-green-600'
					: readinessScore >= 50
						? 'bg-amber-400'
						: 'bg-red-400'}"
				style="width: {readinessScore}%"
			></div>
		</div>
		{#if validationIssues.length > 0}
			<ul class="mt-1.5 space-y-0.5">
				{#each validationIssues as issue (`validation-${issue}`)}
					<li class="text-[10px] text-red-500">• {issue}</li>
				{/each}
			</ul>
		{:else}
			<p class="mt-1 text-[10px] text-green-600">Soal siap disimpan!</p>
		{/if}
	</div>

	<div class="mb-4">
		<p class="mb-2 text-[10px] font-semibold uppercase tracking-wide text-slate-400">
			Sinyal Kualitas
		</p>
		<div class="space-y-2">
			{#each qualitySignals as sig (sig.label)}
				<div class="flex items-start gap-2">
					<span class="mt-0.5 shrink-0 text-sm font-bold {sig.status === 'good' ? 'text-green-600' : 'text-amber-500'}">
						{sig.status === 'good' ? '✓' : '!'}
					</span>
					<div>
						<div class="text-xs font-medium text-slate-700">{sig.label}</div>
						<div class="text-[10px] text-slate-400">{sig.desc}</div>
					</div>
				</div>
			{/each}
		</div>
	</div>

	<div class="my-3 border-t border-slate-200"></div>

	<div>
		<p class="mb-2 text-[10px] font-semibold uppercase tracking-wide text-slate-400">
			Preview Siswa
		</p>
		<div class="rounded-lg border border-slate-200 bg-white p-3 space-y-3" dir={fIsRtl ? 'rtl' : undefined}>
			{#if fStem}
				<RichContent
					html={fStem}
					class="prose prose-sm max-w-none text-slate-800 latex-preview text-sm"
				/>
			{:else}
				<p class="text-xs text-slate-400 italic">Isi soal belum dimasukkan</p>
			{/if}

			{#if fOptions.some(richTextHasContent)}
				<div class="space-y-1.5 border-t border-slate-100 pt-2">
					{#each fOptions as opt, i (ANSWER_LABELS[i])}
						{@const label = ANSWER_LABELS[i]}
						{@const isAnswer = fAnswerKey === label}
						<div class="flex items-start gap-2 text-sm {isAnswer ? 'text-green-700 font-medium' : 'text-slate-700'}">
							<span class="shrink-0 font-bold">{label}.</span>
							{#if richTextHasContent(opt)}
								<RichContent html={opt} class="latex-preview min-w-0 flex-1" />
							{:else}
								<span class="italic text-slate-300">(kosong)</span>
							{/if}
							{#if isAnswer}
								<span class="ml-auto shrink-0 text-xs text-green-500">✓</span>
							{/if}
						</div>
					{/each}
				</div>
			{/if}
		</div>
	</div>
{/snippet}

<!-- ── Composer Dialog ─────────────────────────────────────────────────────── -->
<Dialog.Root bind:open={showComposer}>
	<Dialog.Content>
		<div class="soal-composer-modal flex h-[94vh] w-[min(96vw,110rem)] max-w-[96vw] flex-col overflow-hidden rounded-2xl bg-white text-slate-900 shadow-2xl">
			<div class="shrink-0 border-b border-green-100 bg-white px-4 py-2.5 md:px-6">
				<div class="flex flex-col gap-2 lg:flex-row lg:items-center lg:justify-between">
					<div class="min-w-0">
						<div class="flex flex-wrap items-baseline gap-x-2 gap-y-1">
							<p class="text-[10px] font-semibold uppercase tracking-[0.24em] text-green-700">
								Komposer Soal Legacy
							</p>
							<span class="text-xs text-slate-300">/</span>
							<h2 class="text-base font-black uppercase italic tracking-tight text-slate-900">
								{editingId ? 'Edit Butir Soal' : 'Penyusunan Soal Baru'}
							</h2>
						</div>
					</div>
					<div class="flex flex-wrap gap-2 lg:justify-end">
						<Button
							type="button"
							variant={composerMobilePanel === 'write' ? 'default' : 'outline'}
							size="sm"
							class="h-8 text-xs lg:hidden"
							onclick={() => (composerMobilePanel = 'write')}
						>
							Editor
						</Button>
						<Button
							type="button"
							variant={composerMobilePanel === 'preview' ? 'default' : 'outline'}
							size="sm"
							class="h-8 text-xs lg:hidden"
							onclick={() => (composerMobilePanel = 'preview')}
						>
							Preview
						</Button>
						<Button
							type="button"
							variant="outline"
							size="sm"
							class="h-8 text-xs"
							onclick={closeComposer}
						>
							Tutup
						</Button>
					</div>
				</div>
			</div>

			<div class="min-h-0 flex-1 overflow-y-auto bg-slate-50/70 p-4 md:p-6">
				<div class="grid grid-cols-1 gap-5 xl:grid-cols-[minmax(0,1.45fr)_minmax(22rem,0.72fr)]">
					<div class={`space-y-4 min-w-0 ${composerMobilePanel === 'preview' ? 'hidden lg:block' : 'block'}`}>
						<section class="rounded-xl border border-slate-200 bg-white px-3 py-2 shadow-sm">
							<div class="flex flex-col gap-2 lg:flex-row lg:items-center">
								<div class="flex shrink-0 flex-wrap items-center gap-x-2 gap-y-1">
									<h3 class="text-[10px] font-black uppercase tracking-[0.22em] text-slate-800">Template Cepat</h3>
									<span class="text-[11px] text-slate-500">Struktur awal</span>
								</div>
								{#if showTemplates}
									<div class="min-w-0 flex-1 overflow-x-auto">
										<div class="inline-flex min-w-max items-center gap-1 rounded-lg border border-slate-200 bg-slate-50 p-1">
											{#each templates as t (t.id)}
												<button
													type="button"
													onclick={() => applyTemplate(t)}
													title={t.desc}
													aria-label={`Pakai template ${t.label}: ${t.desc}`}
													class="inline-flex h-8 items-center gap-1.5 rounded-md px-2.5 text-[10px] font-bold uppercase tracking-wide text-slate-700 transition-colors hover:bg-green-50 hover:text-green-800 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-green-500"
												>
													<span>{t.label}</span>
													{#if t.isRtl}
														<span class="rounded bg-amber-100 px-1 py-0.5 text-[8px] font-black text-amber-700">RTL</span>
													{/if}
												</button>
											{/each}
										</div>
									</div>
								{/if}
								<button
									type="button"
									onclick={() => (showTemplates = !showTemplates)}
									class="shrink-0 self-start rounded-md border border-slate-200 px-2.5 py-1 text-[10px] font-semibold uppercase tracking-widest text-slate-500 hover:bg-slate-50 lg:ml-auto lg:self-auto"
								>
									{showTemplates ? 'Sembunyikan' : 'Tampilkan'}
								</button>
							</div>
						</section>

						<section class="rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
							<div class="mb-3 flex flex-wrap items-center justify-between gap-2">
								<h3 class="text-xs font-black uppercase tracking-[0.2em] text-slate-800">Metadata Soal</h3>
								<span class="text-[11px] text-slate-400">Poin final diatur saat soal masuk paket</span>
							</div>
							<div class="grid gap-3 lg:grid-cols-[minmax(0,1fr)_9rem_7rem_11rem] lg:items-end">
								<div>
									<label for="f-subject" class="mb-1 block text-[10px] font-semibold uppercase tracking-wider text-slate-600">
										Mata Pelajaran <span class="text-red-500">*</span>
									</label>
									<select
										id="f-subject"
										bind:value={fSubjectId}
										class="h-9 w-full rounded-md border border-slate-200 bg-white px-3 text-sm font-medium text-slate-700 focus:outline-none focus:ring-2 focus:ring-green-500"
									>
										<option value="">-- Pilih Mapel --</option>
										{#each subjects as s (s.id)}
											<option value={s.id}>{s.name}</option>
										{/each}
									</select>
								</div>
								<div>
									<label for="f-difficulty" class="mb-1 block text-[10px] font-semibold uppercase tracking-wider text-slate-600">
										Kesulitan
									</label>
									<select
										id="f-difficulty"
										bind:value={fDifficulty}
										class="h-9 w-full rounded-md border border-slate-200 bg-white px-3 text-sm font-medium text-slate-700 focus:outline-none focus:ring-2 focus:ring-green-500"
									>
										{#each Object.entries(DIFFICULTY_LABEL) as [val, lbl] (val)}
											<option value={val}>{lbl}</option>
										{/each}
									</select>
								</div>
								<div>
									<label for="f-weight" class="mb-1 block text-[10px] font-semibold uppercase tracking-wider text-slate-600">
										Bobot Paket
									</label>
									<Input id="f-weight" type="number" min="1" bind:value={fWeight} class="h-9 text-sm font-medium" />
								</div>
								<label for="f-rtl" class="flex h-9 cursor-pointer items-center justify-between gap-2 rounded-md border border-dashed border-green-200 bg-green-50 px-3">
									<span class="text-[10px] font-black uppercase tracking-wider text-green-800">Mode Arab / RTL</span>
									<input id="f-rtl" type="checkbox" bind:checked={fIsRtl} class="rounded accent-green-700" />
								</label>
							</div>
						</section>

						<section class="space-y-2 rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
							<div>
								<h3 class="text-xs font-black uppercase tracking-[0.2em] text-slate-800">Isi Pertanyaan</h3>
								<p class="mt-0.5 text-xs text-slate-500">Rich text legacy penuh untuk teks, gambar, daftar, dan formula.</p>
							</div>
							<LegacyRichTextEditor
								bind:value={fStem}
								id="soal-stem"
								placeholder="Tuliskan pertanyaan utama. Gambar bisa disisipkan langsung di antara teks."
								minRows={4}
								compact
								onImageUpload={uploadImageInEditor}
							/>
						</section>

						<section class="space-y-4 border-t border-slate-200 pt-5">
							<div class="text-center">
								<h3 class="text-xs font-black uppercase italic tracking-[0.26em] text-slate-500">Opsi & Kunci Jawaban</h3>
								<p class="mt-1 text-xs text-slate-500">Setiap opsi memakai kotak rich text legacy sendiri.</p>
							</div>
							<div class="grid grid-cols-1 items-start gap-3 xl:grid-cols-2">
								{#each fOptions as _, i (ANSWER_LABELS[i])}
									{@const label = ANSWER_LABELS[i]}
									{@const isAnswer = fAnswerKey === label}
									<section
										class="space-y-2 rounded-xl border bg-white p-3 shadow-sm transition-colors {isAnswer
											? 'border-green-500 ring-4 ring-green-100'
											: 'border-slate-200 hover:border-slate-300'}"
									>
										<div class="flex items-center justify-between gap-3 border-b border-slate-100 pb-2">
											<div>
												<p class="text-xs font-black uppercase tracking-[0.2em] text-slate-700">Opsi {label}</p>
												<p class="mt-0.5 text-[11px] text-slate-400">{isAnswer ? 'Ditandai sebagai kunci jawaban' : 'Pengecoh / alternatif jawaban'}</p>
											</div>
											<label class="flex cursor-pointer items-center gap-2 rounded-lg px-2 py-1.5 transition-colors hover:bg-green-50">
												<input
													type="radio"
													name="answer-key"
													value={label}
													checked={fAnswerKey === label}
													onchange={() => (fAnswerKey = label)}
													class="h-4 w-4 cursor-pointer accent-green-700"
												/>
												<span class="text-[10px] font-black uppercase tracking-widest {isAnswer ? 'text-green-700' : 'text-slate-500'}">Kunci</span>
											</label>
										</div>
										<div dir={fIsRtl ? 'rtl' : undefined}>
											<LegacyRichTextEditor
												bind:value={fOptions[i]}
												id={`soal-option-${label}`}
												placeholder={`Teks jawaban ${label}. Gambar bisa disisipkan langsung di dalam opsi.`}
												minRows={2}
												compact
												onImageUpload={uploadImageInEditor}
											/>
										</div>
									</section>
								{/each}
							</div>
						</section>
					</div>

					<aside class={`space-y-4 min-w-0 xl:sticky xl:top-0 xl:self-start ${composerMobilePanel === 'write' ? 'hidden lg:block' : 'block'}`}>
						<div class="rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
							{@render composerPreview()}
						</div>
					</aside>
				</div>
			</div>

			<div class="flex shrink-0 flex-col gap-2 border-t border-green-100 bg-white px-4 py-2.5 md:flex-row md:items-center md:justify-between md:px-6">
				<div class="min-w-0 text-xs text-slate-500">
					{#if draftStatus}
						<span class="font-semibold text-green-700">{draftStatus}</span>
					{:else}
						<span>Editor legacy siap untuk buat/edit soal.</span>
					{/if}
				</div>
				<div class="flex shrink-0 flex-wrap justify-end gap-2">
					<Button
						variant="outline"
						class="h-8 text-xs"
						onclick={() => {
							clearDraft();
							closeComposer();
						}}
					>
						Batalkan
					</Button>
					<LoadingButton
						onclick={() => void saveQuestion()}
						disabled={!canSave}
						loading={composerBusy}
						loadingLabel="Menyimpan..."
						class="h-8 bg-green-700 text-xs text-white hover:bg-green-800 disabled:opacity-50"
					>
						{editingId ? 'Simpan Perubahan' : 'Simpan Soal'}
					</LoadingButton>
				</div>
			</div>
		</div>
	</Dialog.Content>
</Dialog.Root>

<style>
	:global(.content:has(.soal-composer-modal)) {
		width: min(96vw, 110rem);
		min-width: min(96vw, 110rem);
		max-width: 96vw;
		max-height: 94vh;
		overflow: hidden;
		border: 0;
		border-radius: 1rem;
		padding: 0;
	}

	:global(.latex-preview .latex-display) {
		overflow-x: auto;
		padding: 0.25rem 0;
	}
	:global(.latex-preview .katex-display) {
		margin: 0.5rem 0;
	}
</style>
