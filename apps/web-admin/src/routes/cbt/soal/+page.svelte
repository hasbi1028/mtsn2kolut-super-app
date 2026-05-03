<script lang="ts">
	import { onMount } from 'svelte';
	import * as Dialog from '$lib/components/ui/dialog';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Textarea } from '$lib/components/ui/textarea';
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
	type OptionItem = {
		label: string;
		text?: string;
		html?: string;
		latex?: string;
		match_label?: string;
		match_text?: string;
		match_html?: string;
		is_distractor?: boolean;
	};
	type MatchingPair = { left: string; right: string };
	type ModuleMode = 'catalog' | 'composer' | 'review' | 'import';
	type AuthoringMode = 'beginner' | 'advance';
	type RevisionSourceFilter = '' | 'item_analysis' | 'reviewer' | 'workflow';
	type ReviewDecision = 'approve' | 'reject';
	type ComposerQuestionType = 'multiple_choice' | 'multiple_answer' | 'true_false' | 'agree_disagree' | 'matching' | 'short_answer' | 'essay';
	type ComposerSaveIntent = 'draft' | 'review';
	type AnswerMode = 'single_option' | 'multi_option' | 'fixed_pair' | 'matching' | 'short_text' | 'rubric';
	type QuestionTypeConfig = {
		id: ComposerQuestionType;
		label: string;
		shortLabel: string;
		desc: string;
		studentHint: string;
		answerMode: AnswerMode;
		minOptions: number;
		maxOptions: number;
		fixedOptions?: string[];
	};
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
		stimulus_html?: string;
		explanation_html?: string;
		rubric_html?: string;
		academic_phase?: string;
		grade_level?: number | null;
		cp_ref?: string;
		tp_ref?: string;
		kd_ref?: string;
		indicator_ref?: string;
		material_topic?: string;
		cognitive_level?: string;
		hots_flag?: boolean;
		workflow_status: string;
		options: OptionItem[];
		answer_key: string;
		difficulty: string;
		status: string;
		author_username: string;
		reviewer_username?: string;
		reviewed_at?: string | null;
		review_notes?: string;
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
		revisionQueue: Question[];
		revisionTotal: number;
		reviewQueue: Question[];
		reviewTotal: number;
		approvedQueue: Question[];
		approvedTotal: number;
		totalItems: number;
		page: number;
	};
	type PageData = {
		user?: {
			role?: string;
			roles?: string[];
		};
	};
	type AcademicPayload = {
		subjects?: Subject[];
		error?: string;
		message?: string;
	};
	type DraftPayload = {
		subjectId: string;
		questionType: ComposerQuestionType;
		authoringMode: AuthoringMode;
		stem: string;
		stimulus: string;
		rubric: string;
		explanation: string;
		options: string[];
		matchingPairs: MatchingPair[];
		matchingDistractors: string[];
		answerKey: string;
		weight: number;
		difficulty: string;
		isRtl: boolean;
		gradeLevel: number;
		academicPhase: string;
		cpRef: string;
		tpRef: string;
		kdRef: string;
		indicatorRef: string;
		materialTopic: string;
		cognitiveLevel: string;
		hotsFlag: boolean;
		workflowStatus: string;
		savedAt: string;
	};
	type OptionLabel = 'A' | 'B' | 'C' | 'D' | 'E' | 'F';
	type LegacyImportResult = {
		total_rows: number;
		imported: number;
		skipped: number;
		errors: string[];
		duplicate_codes: string[];
	};
	type FocusedEditor = 'stem' | 'stimulus' | 'rubric' | 'explanation' | OptionLabel;

	let { data }: { data: PageData } = $props();

	// ── Constants ─────────────────────────────────────────────────────────────
	const PAGE_SIZE = 15;
	const MIN_OPTION_COUNT = 4;
	const MAX_OPTION_COUNT = 6;
	const MIN_MATCHING_PAIR_COUNT = 2;
	const DEFAULT_MATCHING_PAIR_COUNT = 4;
	const MAX_MATCHING_PAIR_COUNT = 6;
	const MAX_MATCHING_DISTRACTOR_COUNT = 4;
	const ANSWER_LABELS: OptionLabel[] = ['A', 'B', 'C', 'D', 'E', 'F'];
	const TRUE_FALSE_OPTIONS = ['Benar', 'Salah'];
	const AGREE_DISAGREE_OPTIONS = ['Setuju', 'Tidak Setuju'];
	const QUESTION_TYPE_CONFIGS: QuestionTypeConfig[] = [
		{
			id: 'multiple_choice',
			label: 'Pilihan Ganda',
			shortLabel: 'PG',
			desc: 'Satu jawaban benar dari beberapa opsi.',
			studentHint: 'Siswa memilih satu jawaban.',
			answerMode: 'single_option',
			minOptions: MIN_OPTION_COUNT,
			maxOptions: MAX_OPTION_COUNT,
		},
		{
			id: 'multiple_answer',
			label: 'Jawaban Ganda',
			shortLabel: 'Ganda',
			desc: 'Lebih dari satu opsi dapat menjadi kunci.',
			studentHint: 'Siswa memilih semua jawaban benar.',
			answerMode: 'multi_option',
			minOptions: MIN_OPTION_COUNT,
			maxOptions: MAX_OPTION_COUNT,
		},
		{
			id: 'true_false',
			label: 'Benar/Salah',
			shortLabel: 'B/S',
			desc: 'Pernyataan dengan kunci benar atau salah.',
			studentHint: 'Siswa memilih Benar atau Salah.',
			answerMode: 'fixed_pair',
			minOptions: 2,
			maxOptions: 2,
			fixedOptions: TRUE_FALSE_OPTIONS,
		},
		{
			id: 'agree_disagree',
			label: 'Setuju/Tidak Setuju',
			shortLabel: 'S/TS',
			desc: 'Respons sikap/pendapat dengan kunci setuju atau tidak setuju.',
			studentHint: 'Siswa memilih Setuju atau Tidak Setuju.',
			answerMode: 'fixed_pair',
			minOptions: 2,
			maxOptions: 2,
			fixedOptions: AGREE_DISAGREE_OPTIONS,
		},
		{
			id: 'matching',
			label: 'Menjodohkan',
			shortLabel: 'Jodoh',
			desc: 'Pasangkan pernyataan kiri dengan jawaban kanan.',
			studentHint: 'Siswa memilih pasangan yang sesuai untuk setiap baris.',
			answerMode: 'matching',
			minOptions: MIN_MATCHING_PAIR_COUNT,
			maxOptions: MAX_MATCHING_PAIR_COUNT,
		},
		{
			id: 'short_answer',
			label: 'Isian Singkat',
			shortLabel: 'Isian',
			desc: 'Jawaban pendek dengan kunci teks.',
			studentHint: 'Siswa mengetik jawaban singkat.',
			answerMode: 'short_text',
			minOptions: 0,
			maxOptions: 0,
		},
		{
			id: 'essay',
			label: 'Essay',
			shortLabel: 'Essay',
			desc: 'Jawaban uraian dikoreksi manual dengan pedoman.',
			studentHint: 'Siswa menulis jawaban uraian.',
			answerMode: 'rubric',
			minOptions: 0,
			maxOptions: 0,
		},
	];
	const moduleModes: Array<{ id: ModuleMode; label: string; desc: string }> = [
		{ id: 'catalog', label: 'Katalog', desc: 'Daftar terpadu' },
		{ id: 'composer', label: 'Komposer Soal', desc: 'Semua tipe' },
		{ id: 'review', label: 'Review', desc: 'Mutu & publikasi' },
		{ id: 'import', label: 'Import CSV', desc: 'Multi-tipe' },
	];
	const WORKFLOW_LABEL: Record<string, string> = {
		draft: 'Draft',
		review: 'Ditinjau',
		approved: 'Disetujui',
		rejected: 'Revisi',
	};
	const DIFFICULTY_LABEL: Record<string, string> = { easy: 'Mudah', medium: 'Sedang', hard: 'Sulit' };
	const revisionSourceOptions: Array<{ id: RevisionSourceFilter; label: string; desc: string }> = [
		{ id: '', label: 'Semua Revisi', desc: 'Semua sumber' },
		{ id: 'item_analysis', label: 'Analisis Butir', desc: 'Dari hasil ujian' },
		{ id: 'reviewer', label: 'Reviewer', desc: 'Catatan penelaah' },
		{ id: 'workflow', label: 'Workflow', desc: 'Tanpa reviewer' },
	];

	const DRAFT_KEY = (id: string | null) => `mtsn2-soal-komposer:${id ?? 'new'}`;

	// ── Page state ─────────────────────────────────────────────────────────────
	let activeMode = $state<ModuleMode>('catalog');
	let questionsPromise = $state<Promise<SoalOverview> | null>(null);
	let questions = $state<Question[]>([]);
	let subjects = $state<Subject[]>([]);
	let revisionQueue = $state<Question[]>([]);
	let revisionTotal = $state(0);
	let reviewQueue = $state<Question[]>([]);
	let reviewTotal = $state(0);
	let approvedQueue = $state<Question[]>([]);
	let approvedTotal = $state(0);
	let totalItems = $state(0);
	let currentPage = $state(1);

	let search = $state('');
	let filterSubject = $state('');
	let filterWorkflow = $state('');
	let filterStatus = $state('');
	let revisionSourceFilter = $state<RevisionSourceFilter>('');

	// ── Composer state ─────────────────────────────────────────────────────────
	let showComposer = $state(false);
	let editingId = $state<string | null>(null);
	let composerBusy = $state(false);
	let composerAction = $state<ComposerSaveIntent | ''>('');
	let draftStatus = $state('');
	let draftSavedAt = $state<string | null>(null);
	let showInspector = $state(false);
	let composerMobilePanel = $state<'write' | 'preview'>('write');
	let focusedEditor = $state<FocusedEditor | null>(null);
	let lastDraftSig = '';
	let questionsRequestId = 0;
	let showImport = $state(false);
	let importSubjectId = $state('');
	let importFile = $state<File | null>(null);
	let importBusy = $state(false);
	let importResult = $state<LegacyImportResult | null>(null);
	let exportBusy = $state(false);
	let templateBusy = $state(false);
	let duplicateBusyId = $state('');
	let workflowBusyId = $state('');
	let reviewDecisionOpen = $state(false);
	let reviewDecisionQuestion = $state<Question | null>(null);
	let reviewDecision = $state<ReviewDecision>('approve');
	let reviewDecisionNotes = $state('');

	// ── Form fields ────────────────────────────────────────────────────────────
	let fSubjectId = $state('');
	let fQuestionType = $state<ComposerQuestionType>('multiple_choice');
	let fAuthoringMode = $state<AuthoringMode>('beginner');
	let fStem = $state('');
	let fStimulus = $state('');
	let fRubric = $state('');
	let fExplanation = $state('');
	let fOptions = $state(['', '', '', '']);
	let fMatchingPairs = $state<MatchingPair[]>(createEmptyMatchingPairs());
	let fMatchingDistractors = $state<string[]>([]);
	let fAnswerKey = $state('A');
	let fWeight = $state(1);
	let fDifficulty = $state('medium');
	let fIsRtl = $state(false);
	let fGradeLevel = $state(7);
	let fAcademicPhase = $state('');
	let fCPRef = $state('');
	let fTPRef = $state('');
	let fKDRef = $state('');
	let fIndicatorRef = $state('');
	let fMaterialTopic = $state('');
	let fCognitiveLevel = $state('');
	let fHotsFlag = $state(false);
	let fWorkflowStatus = $state('draft');
	let activeDraftKey = $derived(DRAFT_KEY(editingId));
	let draftSignature = $derived(JSON.stringify({
		fSubjectId,
		fQuestionType,
		fAuthoringMode,
		fStem,
		fStimulus,
		fRubric,
		fExplanation,
		opts: fOptions,
		matchingPairs: fMatchingPairs,
		matchingDistractors: fMatchingDistractors,
		fAnswerKey,
		fWeight,
		fDifficulty,
		fIsRtl,
		fGradeLevel,
		fAcademicPhase,
		fCPRef,
		fTPRef,
		fKDRef,
		fIndicatorRef,
		fMaterialTopic,
		fCognitiveLevel,
		fHotsFlag,
		fWorkflowStatus,
	}));

	// ── Derived ────────────────────────────────────────────────────────────────
	let pageCount = $derived(Math.max(1, Math.ceil(totalItems / PAGE_SIZE)));
	let lockedCount = $derived(questions.filter(questionUsageLocked).length);
	let roles = $derived(data.user?.roles ?? (data.user?.role ? [data.user.role] : []));
	let canReviewWorkflow = $derived(roles.includes('admin'));
	let reviewCount = $derived(reviewTotal);
	let visibleReviewCount = $derived(questions.filter((item) => item.workflow_status === 'review').length);
	let approvedCount = $derived(approvedTotal);
	let visibleApprovedCount = $derived(questions.filter((item) => item.workflow_status === 'approved' && item.status === 'draft').length);
	let draftCount = $derived(questions.filter((item) => item.workflow_status === 'draft').length);
	let visibleRevisionCount = $derived(questions.filter((item) => item.workflow_status === 'rejected').length);
	let publishedCount = $derived(questions.filter((item) => item.status === 'published').length);

	let stemText = $derived(htmlToPlainText(fStem));
	let hasImage = $derived(fStem.includes('<img'));
	let stimulusText = $derived(htmlToPlainText(fStimulus));
	let rubricText = $derived(htmlToPlainText(fRubric));
	let explanationText = $derived(htmlToPlainText(fExplanation));
	let questionTypeConfig = $derived(getQuestionTypeConfig(fQuestionType));
	let isEssay = $derived(fQuestionType === 'essay');
	let isMultipleAnswer = $derived(questionTypeConfig.answerMode === 'multi_option');
	let isTrueFalse = $derived(fQuestionType === 'true_false');
	let isAgreeDisagree = $derived(fQuestionType === 'agree_disagree');
	let isFixedPair = $derived(questionTypeConfig.answerMode === 'fixed_pair');
	let isMatching = $derived(questionTypeConfig.answerMode === 'matching');
	let isShortAnswer = $derived(questionTypeConfig.answerMode === 'short_text');
	let hasOptionSection = $derived(
		questionTypeConfig.answerMode === 'single_option' ||
			questionTypeConfig.answerMode === 'multi_option' ||
			isFixedPair
	);
	let hasEditableOptions = $derived(
		questionTypeConfig.answerMode === 'single_option' ||
			questionTypeConfig.answerMode === 'multi_option'
	);
	let requiresRubric = $derived(questionTypeConfig.answerMode === 'rubric');
	let isAdvanceMode = $derived(fAuthoringMode === 'advance');
	let activeOptionLabels = $derived(ANSWER_LABELS.slice(0, fOptions.length));
	let selectedAnswerLabels = $derived(answerKeyLabels(fAnswerKey, activeOptionLabels));
	let optionPlainTexts = $derived(fOptions.map((option) => htmlToPlainText(option)));
	let optionHasImages = $derived(fOptions.map((option) => option.includes('<img')));
	let matchingLeftTexts = $derived(fMatchingPairs.map((pair) => htmlToPlainText(pair.left)));
	let matchingRightTexts = $derived(fMatchingPairs.map((pair) => htmlToPlainText(pair.right)));
	let matchingDistractorTexts = $derived(fMatchingDistractors.map((distractor) => htmlToPlainText(distractor)));
	let matchingPairsReady = $derived(
		!isMatching ||
			(fMatchingPairs.length >= questionTypeConfig.minOptions &&
				fMatchingPairs.every((pair) => richTextHasContent(pair.left) && richTextHasContent(pair.right)))
	);
	let optionsReady = $derived(!hasEditableOptions || (fOptions.length >= questionTypeConfig.minOptions && fOptions.every(richTextHasContent)));
	let shortAnswerAliases = $derived(parseShortAnswerAliases(fAnswerKey));
	let answerKeyReady = $derived.by(() => {
		if (requiresRubric) return true;
		if (isMatching) return matchingPairsReady;
		if (isShortAnswer) return shortAnswerAliases.length > 0;
		if (isMultipleAnswer) return selectedAnswerLabels.length >= 2;
		return selectedAnswerLabels.length === 1;
	});
	let rubricReady = $derived(!requiresRubric || rubricText.length >= 5 || fRubric.includes('<img'));

	let readinessChecks = $derived({
		subject: !!fSubjectId,
		stem: stemText.length >= 5 || hasImage,
		options: isEssay || (isMatching ? matchingPairsReady : optionsReady),
		answerKey: answerKeyReady,
		rubric: rubricReady,
		weight: Number.isFinite(fWeight) && fWeight >= 1,
	});

	let passedChecks = $derived(Object.values(readinessChecks).filter(Boolean).length);
	let totalChecks = $derived(Object.keys(readinessChecks).length);
	let readinessScore = $derived(Math.round((passedChecks / totalChecks) * 100));
	let draftIssues = $derived.by(() => {
		const issues: string[] = [];
		if (!readinessChecks.subject) issues.push('Pilih mata pelajaran sebelum menyimpan draft');
		if (!readinessChecks.stem) issues.push('Isi pertanyaan minimal 5 karakter untuk draft');
		if (!readinessChecks.weight) issues.push('Bobot nilai minimal 1');
		return issues;
	});
	let canSaveDraft = $derived(draftIssues.length === 0 && !composerBusy);
	let canSubmitReview = $derived(readinessScore === 100 && !composerBusy);

	let qualitySignals = $derived.by(() => {
		if (isEssay) {
			return [
				{
					label: 'Pertanyaan terbuka jelas',
					status: stemText.length >= 35 ? 'good' : 'warn',
					desc: `${stemText.length} / 35 karakter minimum`,
				},
				{
					label: 'Pedoman penilaian tersedia',
					status: rubricReady ? 'good' : 'warn',
					desc: rubricReady ? 'Rubrik/pedoman terisi' : `${rubricText.length} / 5 karakter minimum`,
				},
				{
					label: 'Stimulus pendukung',
					status: !isAdvanceMode || stimulusText.length > 0 || hasImage || stemText.length >= 70 ? 'good' : 'warn',
					desc: isAdvanceMode ? (stimulusText ? 'Stimulus terisi' : `${stimulusText.length} karakter stimulus`) : 'Opsional di mode pemula',
				},
				{
					label: 'Level kognitif',
					status: !isAdvanceMode || fCognitiveLevel.trim().length > 0 ? 'good' : 'warn',
					desc: isAdvanceMode ? (fCognitiveLevel.trim() || 'Belum diisi') : 'Opsional di mode pemula',
				},
			];
		}
		if (isMatching) {
			const filledLeft = matchingLeftTexts.filter(Boolean);
			const filledRight = [...matchingRightTexts, ...matchingDistractorTexts].filter(Boolean);
			const uniqueRight = new Set(filledRight);
			const filledDistractors = matchingDistractorTexts.filter(Boolean).length;
			return [
				{
					label: 'Instruksi menjodohkan jelas',
					status: stemText.length >= 25 ? 'good' : 'warn',
					desc: `${stemText.length} / 25 karakter minimum`,
				},
				{
					label: 'Pasangan lengkap',
					status: matchingPairsReady ? 'good' : 'warn',
					desc: `${Math.min(filledLeft.length, filledRight.length)} / ${fMatchingPairs.length} pasangan terisi`,
				},
				{
					label: 'Jawaban kanan unik',
					status: filledRight.length > 0 && uniqueRight.size === filledRight.length ? 'good' : 'warn',
					desc: uniqueRight.size < filledRight.length ? 'Ada pasangan kanan yang sama' : 'Semua pasangan kanan berbeda',
				},
				{
					label: 'Distraktor kanan',
					status: fMatchingDistractors.length === 0 || filledDistractors === fMatchingDistractors.length ? 'good' : 'warn',
					desc: fMatchingDistractors.length === 0 ? 'Opsional' : `${filledDistractors} / ${fMatchingDistractors.length} distraktor terisi`,
				},
				{
					label: 'Skoring deterministik',
					status: matchingPairsReady ? 'good' : 'warn',
					desc: buildMatchingAnswerKey(fMatchingPairs.length) || 'Belum ada pasangan',
				},
			];
		}
		if (isShortAnswer) {
			return [
				{
					label: 'Pertanyaan singkat jelas',
					status: stemText.length >= 25 ? 'good' : 'warn',
					desc: `${stemText.length} / 25 karakter minimum`,
				},
				{
					label: 'Kunci jawaban tersedia',
					status: answerKeyReady ? 'good' : 'warn',
					desc: answerKeyReady ? `${shortAnswerAliases.length} jawaban diterima` : 'Belum ada kunci teks',
				},
				{
					label: 'Alias tidak berlebihan',
					status: shortAnswerAliases.length > 0 && shortAnswerAliases.length <= 5 ? 'good' : 'warn',
					desc: `${shortAnswerAliases.length} alias aktif`,
				},
				{
					label: 'Stimulus pendukung',
					status: !isAdvanceMode || stimulusText.length > 0 || hasImage || stemText.length >= 60 ? 'good' : 'warn',
					desc: isAdvanceMode ? (stimulusText ? 'Stimulus terisi' : `${stimulusText.length} karakter stimulus`) : 'Opsional di mode pemula',
				},
			];
		}
		if (isTrueFalse || isAgreeDisagree) {
			return [
				{
					label: isAgreeDisagree ? 'Sikap/pendapat jelas' : 'Pernyataan tegas',
					status: stemText.length >= 20 ? 'good' : 'warn',
					desc: `${stemText.length} / 20 karakter minimum`,
				},
				{
					label: isAgreeDisagree ? 'Kunci setuju/tidak setuju dipilih' : 'Kunci benar/salah dipilih',
					status: answerKeyReady ? 'good' : 'warn',
					desc: answerKeyReady ? `Kunci ${fixedOptionTextForKey(fAnswerKey)}` : 'Belum dipilih',
				},
				{
					label: isAgreeDisagree ? 'Berbasis sikap terukur' : 'Tidak ambigu',
					status: isAgreeDisagree || !stemText.includes('?') ? 'good' : 'warn',
					desc: isAgreeDisagree ? 'Format respons sikap' : stemText.includes('?') ? 'B/S lebih kuat sebagai pernyataan' : 'Format pernyataan',
				},
				{
					label: 'Media pendukung',
					status: hasImage || stemText.length >= 60 ? 'good' : 'warn',
					desc: hasImage ? 'Ada gambar' : `${stemText.length} / 60 karakter`,
				},
			];
		}
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
				label: isMultipleAnswer ? 'Kunci jawaban ganda' : 'Distraktor bervariasi',
				status: isMultipleAnswer ? (selectedAnswerLabels.length >= 2 ? 'good' : 'warn') : (unique.size === filled.length ? 'good' : 'warn'),
				desc: isMultipleAnswer
					? `${selectedAnswerLabels.length} kunci dipilih`
					: unique.size < filled.length ? 'Ada opsi yang duplikat' : 'Semua opsi unik',
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
		if (isMatching && !readinessChecks.options)
			issues.push('Lengkapi semua pasangan kiri dan kanan');
		if (hasEditableOptions && !readinessChecks.options)
			issues.push(`Semua opsi (${activeOptionLabels.join('–')}) wajib diisi`);
		if (!readinessChecks.answerKey) {
			if (isMatching) issues.push('Lengkapi pasangan menjodohkan');
			else if (isShortAnswer) issues.push('Isi kunci jawaban isian singkat');
			else if (isMultipleAnswer) issues.push('Pilih minimal dua kunci jawaban');
			else issues.push('Pilih kunci jawaban');
		}
		if (requiresRubric && !readinessChecks.rubric) issues.push('Isi pedoman/rubrik penilaian essay');
		if (!readinessChecks.weight) issues.push('Bobot nilai minimal 1');
		return issues;
	});
	let qualityWarningCount = $derived(qualitySignals.filter((signal) => signal.status !== 'good').length);
	let hasDraftWork = $derived(Boolean(
		fSubjectId ||
		richTextHasContent(fStem) ||
		richTextHasContent(fStimulus) ||
		richTextHasContent(fRubric) ||
		richTextHasContent(fExplanation) ||
		fOptions.some(richTextHasContent) ||
		fMatchingPairs.some((pair) => richTextHasContent(pair.left) || richTextHasContent(pair.right)) ||
		fMatchingDistractors.some(richTextHasContent) ||
		draftStatus
	));

	// ── Draft autosave ─────────────────────────────────────────────────────────
	function buildDraftPayload(): DraftPayload {
		return {
			subjectId: fSubjectId,
			questionType: fQuestionType,
			authoringMode: fAuthoringMode,
			stem: fStem,
			stimulus: fStimulus,
			rubric: fRubric,
			explanation: fExplanation,
			options: [...fOptions],
			matchingPairs: fMatchingPairs.map((pair) => ({ ...pair })),
			matchingDistractors: [...fMatchingDistractors],
			answerKey: fAnswerKey,
			weight: fWeight,
			difficulty: fDifficulty,
			isRtl: fIsRtl,
			gradeLevel: fGradeLevel,
			academicPhase: fAcademicPhase,
			cpRef: fCPRef,
			tpRef: fTPRef,
			kdRef: fKDRef,
			indicatorRef: fIndicatorRef,
			materialTopic: fMaterialTopic,
			cognitiveLevel: fCognitiveLevel,
			hotsFlag: fHotsFlag,
			workflowStatus: fWorkflowStatus,
			savedAt: new Date().toISOString(),
		};
	}

	function markDraftAutosaved() {
		draftStatus = 'Draft tersimpan otomatis';
		draftSavedAt = new Date().toISOString();
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
				questionType?: string;
				authoringMode?: string;
				stem?: string;
				stimulus?: string;
				rubric?: string;
				explanation?: string;
				options?: string[];
				matchingPairs?: MatchingPair[];
				matchingDistractors?: string[];
				answerKey?: string;
				weight?: number;
				difficulty?: string;
				isRtl?: boolean;
				gradeLevel?: number;
				academicPhase?: string;
				cpRef?: string;
				tpRef?: string;
				kdRef?: string;
				indicatorRef?: string;
				materialTopic?: string;
				cognitiveLevel?: string;
				hotsFlag?: boolean;
				workflowStatus?: string;
				savedAt?: string;
			};
			fSubjectId = d.subjectId ?? '';
			fQuestionType = normalizeQuestionType(d.questionType);
			fAuthoringMode = normalizeAuthoringMode(d.authoringMode);
			fStem = d.stem ?? '';
			fStimulus = d.stimulus ?? '';
			fRubric = d.rubric ?? '';
			fExplanation = d.explanation ?? '';
			fOptions = normalizeOptionCount(d.options ?? defaultOptionsForQuestionType(fQuestionType), fQuestionType);
			fMatchingPairs = normalizeMatchingPairs(d.matchingPairs ?? [], fQuestionType);
			fMatchingDistractors = normalizeMatchingDistractors(d.matchingDistractors ?? [], fQuestionType);
			fAnswerKey = normalizeAnswerKey(d.answerKey, fQuestionType, answerItemCountForType(fQuestionType));
			fWeight = d.weight ?? 1;
			fDifficulty = d.difficulty ?? 'medium';
			fIsRtl = d.isRtl ?? false;
			fGradeLevel = d.gradeLevel ?? 7;
			fAcademicPhase = d.academicPhase ?? '';
			fCPRef = d.cpRef ?? '';
			fTPRef = d.tpRef ?? '';
			fKDRef = d.kdRef ?? '';
			fIndicatorRef = d.indicatorRef ?? '';
			fMaterialTopic = d.materialTopic ?? '';
			fCognitiveLevel = d.cognitiveLevel ?? '';
			fHotsFlag = d.hotsFlag ?? false;
			fWorkflowStatus = normalizeWorkflowStatus(d.workflowStatus);
			draftStatus = 'Draft lokal dipulihkan';
			draftSavedAt = d.savedAt ?? null;
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
		draftSavedAt = null;
	}

	// ── API ────────────────────────────────────────────────────────────────────
	function buildQuestionParams(page: number) {
		const params = new URLSearchParams();
		params.set('limit', String(PAGE_SIZE));
		params.set('offset', String((page - 1) * PAGE_SIZE));
		if (search.trim()) params.set('q', search.trim());
		if (filterSubject) params.set('subject_id', filterSubject);
		if (filterWorkflow) params.set('workflow_status', filterWorkflow);
		if (filterStatus) params.set('status', filterStatus);
		if (filterWorkflow === 'rejected' && revisionSourceFilter) params.set('revision_source', revisionSourceFilter);
		return params;
	}

	function buildRevisionQueueParams() {
		const params = new URLSearchParams();
		params.set('limit', '6');
		params.set('offset', '0');
		params.set('workflow_status', 'rejected');
		if (revisionSourceFilter) params.set('revision_source', revisionSourceFilter);
		if (search.trim()) params.set('q', search.trim());
		if (filterSubject) params.set('subject_id', filterSubject);
		return params;
	}

	function buildReviewQueueParams() {
		const params = new URLSearchParams();
		params.set('limit', '6');
		params.set('offset', '0');
		params.set('workflow_status', 'review');
		if (search.trim()) params.set('q', search.trim());
		if (filterSubject) params.set('subject_id', filterSubject);
		return params;
	}

	function buildApprovedQueueParams() {
		const params = new URLSearchParams();
		params.set('limit', '6');
		params.set('offset', '0');
		params.set('workflow_status', 'approved');
		params.set('status', 'draft');
		if (search.trim()) params.set('q', search.trim());
		if (filterSubject) params.set('subject_id', filterSubject);
		return params;
	}

	async function fetchOverview(page = currentPage): Promise<SoalOverview> {
		const params = buildQuestionParams(page);
		const revisionParams = buildRevisionQueueParams();
		const reviewParams = buildReviewQueueParams();
		const approvedParams = buildApprovedQueueParams();
		const [questionPayload, revisionPayload, reviewPayload, approvedPayload, academicPayload] = await Promise.all([
			fetch(clientApiPathWithQuery('/api/cbt/questions', params)).then((response) =>
				readClientApiData<QuestionListResponse>(response, 'Gagal memuat soal')
			),
			fetch(clientApiPathWithQuery('/api/cbt/questions', revisionParams)).then((response) =>
				readClientApiData<QuestionListResponse>(response, 'Gagal memuat antrian revisi')
			),
			fetch(clientApiPathWithQuery('/api/cbt/questions', reviewParams)).then((response) =>
				readClientApiData<QuestionListResponse>(response, 'Gagal memuat antrian review')
			),
			fetch(clientApiPathWithQuery('/api/cbt/questions', approvedParams)).then((response) =>
				readClientApiData<QuestionListResponse>(response, 'Gagal memuat antrian siap terbit')
			),
			fetch('/api/academic').then((response) =>
				readClientApiData<AcademicPayload>(response, 'Gagal memuat data akademik')
			),
		]);
		const loadedQuestions = questionPayload.items ?? [];
		const loadedRevisions = revisionPayload.items ?? [];
		const loadedReviews = reviewPayload.items ?? [];
		const loadedApproved = approvedPayload.items ?? [];
		return {
			questions: loadedQuestions,
			subjects: academicPayload.subjects ?? [],
			revisionQueue: loadedRevisions,
			revisionTotal: revisionPayload.meta?.total ?? loadedRevisions.length,
			reviewQueue: loadedReviews,
			reviewTotal: reviewPayload.meta?.total ?? loadedReviews.length,
			approvedQueue: loadedApproved,
			approvedTotal: approvedPayload.meta?.total ?? loadedApproved.length,
			totalItems: questionPayload.meta?.total ?? loadedQuestions.length,
			page,
		};
	}

	function applyOverview(overview: SoalOverview) {
		questions = overview.questions;
		subjects = overview.subjects;
		revisionQueue = overview.revisionQueue;
		revisionTotal = overview.revisionTotal;
		reviewQueue = overview.reviewQueue;
		reviewTotal = overview.reviewTotal;
		approvedQueue = overview.approvedQueue;
		approvedTotal = overview.approvedTotal;
		totalItems = overview.totalItems;
		currentPage = overview.page;
	}

	function load(page = currentPage) {
		const requestId = ++questionsRequestId;
		questionsPromise = fetchOverview(page).then((overview) => {
			if (requestId !== questionsRequestId) return { questions, subjects, revisionQueue, revisionTotal, reviewQueue, reviewTotal, approvedQueue, approvedTotal, totalItems, page: currentPage };
			applyOverview(overview);
			return overview;
		}).catch((error: unknown) => {
			if (requestId === questionsRequestId) throw error;
			return { questions, subjects, revisionQueue, revisionTotal, reviewQueue, reviewTotal, approvedQueue, approvedTotal, totalItems, page: currentPage };
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
				questionsPromise = Promise.resolve({ questions, subjects, revisionQueue, revisionTotal, reviewQueue, reviewTotal, approvedQueue, approvedTotal, totalItems, page: currentPage });
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

	function showAllRevisions() {
		setRevisionSourceFilter('');
	}

	function showPendingReviews() {
		filterWorkflow = 'review';
		filterStatus = '';
		revisionSourceFilter = '';
		setModuleMode('review');
		load(1);
	}

	function showApprovedQuestions() {
		filterWorkflow = 'approved';
		filterStatus = 'draft';
		revisionSourceFilter = '';
		setModuleMode('review');
		load(1);
	}

	function setRevisionSourceFilter(source: RevisionSourceFilter) {
		revisionSourceFilter = source;
		filterWorkflow = 'rejected';
		filterStatus = '';
		setModuleMode('review');
		load(1);
	}

	function onWorkflowFilterChange() {
		if (filterWorkflow !== 'rejected') revisionSourceFilter = '';
		filterStatus = '';
		load(1);
	}

	function richTextHasContent(html: string): boolean {
		return htmlToPlainText(html).length > 0 || html.includes('<img');
	}

	function createEmptyMatchingPairs(count = DEFAULT_MATCHING_PAIR_COUNT): MatchingPair[] {
		return Array.from({ length: count }, () => ({ left: '', right: '' }));
	}

	function isComposerQuestionType(value: string | undefined): value is ComposerQuestionType {
		return QUESTION_TYPE_CONFIGS.some((config) => config.id === value);
	}

	function getQuestionTypeConfig(type: ComposerQuestionType): QuestionTypeConfig {
		return QUESTION_TYPE_CONFIGS.find((config) => config.id === type) ?? QUESTION_TYPE_CONFIGS[0]!;
	}

	function normalizeQuestionType(value: string | undefined): ComposerQuestionType {
		const normalized = (value ?? '').trim().toLowerCase();
		return isComposerQuestionType(normalized) ? normalized : 'multiple_choice';
	}

	function questionTypeLabel(value: string | undefined): string {
		return getQuestionTypeConfig(normalizeQuestionType(value)).shortLabel;
	}

	function normalizeAuthoringMode(value: string | undefined): AuthoringMode {
		return value === 'advance' ? 'advance' : 'beginner';
	}

	function normalizeWorkflowStatus(value: string | undefined): string {
		return value === 'review' ? 'review' : 'draft';
	}

	function setQuestionType(type: ComposerQuestionType) {
		const changed = fQuestionType !== type;
		fQuestionType = type;
		fOptions = normalizeOptionCount(changed ? [] : fOptions, type);
		fMatchingPairs = normalizeMatchingPairs(changed ? [] : fMatchingPairs, type);
		fMatchingDistractors = normalizeMatchingDistractors(changed ? [] : fMatchingDistractors, type);
		fAnswerKey = changed
			? defaultAnswerKeyForQuestionType(type)
			: normalizeAnswerKey(fAnswerKey, type, answerItemCountForType(type));
		if (focusedEditor && !editorAllowedForQuestionType(focusedEditor, type)) {
			focusedEditor = null;
		}
	}

	function setAuthoringMode(mode: AuthoringMode) {
		fAuthoringMode = mode;
		if (mode === 'beginner') fWorkflowStatus = 'draft';
	}

	function defaultAnswerKeyForQuestionType(type: ComposerQuestionType): string {
		const config = getQuestionTypeConfig(type);
		if (config.answerMode === 'single_option' || config.answerMode === 'fixed_pair') return 'A';
		if (config.answerMode === 'matching') return buildMatchingAnswerKey(DEFAULT_MATCHING_PAIR_COUNT);
		return '';
	}

	function defaultOptionsForQuestionType(type: ComposerQuestionType): string[] {
		const config = getQuestionTypeConfig(type);
		if (config.answerMode === 'fixed_pair') return [...(config.fixedOptions ?? [])];
		if (config.answerMode !== 'single_option' && config.answerMode !== 'multi_option') return [];
		return Array.from({ length: config.minOptions }, () => '');
	}

	function normalizeOptionCount(options: string[], type: ComposerQuestionType = fQuestionType): string[] {
		const config = getQuestionTypeConfig(type);
		if (config.answerMode === 'fixed_pair') return [...(config.fixedOptions ?? [])];
		if (config.answerMode !== 'single_option' && config.answerMode !== 'multi_option') return [];
		let normalized = options.slice(0, config.maxOptions);
		while (normalized.length < config.minOptions) normalized = [...normalized, ''];
		return normalized;
	}

	function normalizeMatchingPairs(pairs: MatchingPair[], type: ComposerQuestionType = fQuestionType): MatchingPair[] {
		const config = getQuestionTypeConfig(type);
		if (config.answerMode !== 'matching') return createEmptyMatchingPairs();
		let normalized = pairs
			.slice(0, config.maxOptions)
			.map((pair) => ({ left: pair.left ?? '', right: pair.right ?? '' }));
		const minCount = pairs.length === 0 ? DEFAULT_MATCHING_PAIR_COUNT : config.minOptions;
		while (normalized.length < minCount) normalized = [...normalized, { left: '', right: '' }];
		return normalized;
	}

	function optionsToMatchingPairs(options: OptionItem[]): MatchingPair[] {
		return options.filter((option) => !option.is_distractor).map((option) => ({
			left: option.html || option.text || option.latex || '',
			right: option.match_html || option.match_text || '',
		}));
	}

	function normalizeMatchingDistractors(distractors: string[], type: ComposerQuestionType = fQuestionType): string[] {
		const config = getQuestionTypeConfig(type);
		if (config.answerMode !== 'matching') return [];
		return distractors.slice(0, MAX_MATCHING_DISTRACTOR_COUNT);
	}

	function optionsToMatchingDistractors(options: OptionItem[]): string[] {
		return options
			.filter((option) => option.is_distractor)
			.map((option) => option.match_html || option.match_text || '')
			.filter((option) => option.trim().length > 0);
	}

	function answerItemCountForType(type: ComposerQuestionType = fQuestionType): number {
		return getQuestionTypeConfig(type).answerMode === 'matching' ? fMatchingPairs.length : fOptions.length;
	}

	function optionLabelAt(index: number): OptionLabel {
		return ANSWER_LABELS[index] ?? 'A';
	}

	function answerKeyLabels(value: string | undefined, labels: OptionLabel[]): OptionLabel[] {
		const keys: OptionLabel[] = [];
		for (const raw of (value ?? '').split(',')) {
			const label = raw.trim().toUpperCase() as OptionLabel;
			if (!labels.includes(label) || keys.includes(label)) continue;
			keys.push(label);
		}
		return keys.sort((a, b) => ANSWER_LABELS.indexOf(a) - ANSWER_LABELS.indexOf(b));
	}

	function fixedOptionTextForKey(key: string): string {
		const label = key.trim().toUpperCase() as OptionLabel;
		const index = activeOptionLabels.indexOf(label);
		return questionTypeConfig.fixedOptions?.[index] ?? label;
	}

	function buildMatchingAnswerKey(count: number): string {
		return Array.from({ length: count }, (_, index) => `${optionLabelAt(index)}=${index + 1}`).join(';');
	}

	function normalizeMatchingAnswerKey(value: string | undefined, count: number): string {
		const fallback = buildMatchingAnswerKey(count);
		const rawPairs = (value ?? '').split(';');
		const labels = ANSWER_LABELS.slice(0, count);
		const matches: Record<string, string> = {};
		for (const rawPair of rawPairs) {
			const [rawLeft, rawRight] = rawPair.split('=');
			const left = rawLeft?.trim().toUpperCase() ?? '';
			const right = rawRight?.trim() ?? '';
			if (!labels.includes(left as OptionLabel) || !/^[1-6]$/.test(right)) continue;
			const rightIndex = Number(right) - 1;
			if (rightIndex < 0 || rightIndex >= count) continue;
			matches[left] = String(rightIndex + 1);
		}
		if (Object.keys(matches).length !== count) return fallback;
		return labels.map((label) => `${label}=${matches[label]}`).join(';');
	}

	function normalizeShortAnswerComparable(value: string): string {
		return value.replace(/\u00a0/g, ' ').trim().replace(/\s+/g, ' ').toLocaleLowerCase('id-ID');
	}

	function parseShortAnswerAliases(value: string | undefined): string[] {
		const aliases: string[] = [];
		for (const raw of (value ?? '').split('|')) {
			const alias = raw.replace(/\u00a0/g, ' ').trim().replace(/\s+/g, ' ');
			const comparable = normalizeShortAnswerComparable(alias);
			if (!comparable || aliases.some((item) => normalizeShortAnswerComparable(item) === comparable)) continue;
			aliases.push(alias);
		}
		return aliases;
	}

	function normalizeAnswerKey(value: string | undefined, type: ComposerQuestionType = fQuestionType, optionCount = fOptions.length): string {
		const config = getQuestionTypeConfig(type);
		if (config.answerMode === 'rubric') return '';
		if (config.answerMode === 'short_text') return parseShortAnswerAliases(value).join('|');
		if (config.answerMode === 'matching') return normalizeMatchingAnswerKey(value, optionCount);
		const labels = ANSWER_LABELS.slice(0, optionCount);
		const keys = answerKeyLabels(value, labels);
		if (config.answerMode === 'multi_option') return keys.join(',');
		return keys[0] ?? defaultAnswerKeyForQuestionType(type);
	}

	function isAnswerLabelSelected(label: OptionLabel): boolean {
		return selectedAnswerLabels.includes(label);
	}

	function toggleAnswerLabel(label: OptionLabel) {
		if (!isMultipleAnswer) {
			fAnswerKey = label;
			return;
		}
		const selected = answerKeyLabels(fAnswerKey, activeOptionLabels);
		const next = selected.includes(label)
			? selected.filter((item) => item !== label)
			: [...selected, label];
		fAnswerKey = normalizeAnswerKey(next.join(','), fQuestionType, fOptions.length);
	}

	function editorAllowedForQuestionType(editor: FocusedEditor, type: ComposerQuestionType): boolean {
		if (editor === 'stem' || editor === 'stimulus' || editor === 'explanation') return true;
		if (editor === 'rubric') return getQuestionTypeConfig(type).answerMode === 'rubric';
		return getQuestionTypeConfig(type).answerMode === 'single_option' || getQuestionTypeConfig(type).answerMode === 'multi_option';
	}

	function addOption() {
		const config = getQuestionTypeConfig(fQuestionType);
		if (config.answerMode !== 'single_option' && config.answerMode !== 'multi_option') return;
		if (fOptions.length >= config.maxOptions) return;
		fOptions = [...fOptions, ''];
	}

	function removeLastOption() {
		const config = getQuestionTypeConfig(fQuestionType);
		if (config.answerMode !== 'single_option' && config.answerMode !== 'multi_option') return;
		if (fOptions.length <= config.minOptions) return;
		const removedLabel = optionLabelAt(fOptions.length - 1);
		fOptions = fOptions.slice(0, -1);
		if (answerKeyLabels(fAnswerKey, ANSWER_LABELS).includes(removedLabel)) {
			fAnswerKey = normalizeAnswerKey(fAnswerKey, fQuestionType, fOptions.length);
		}
	}

	function addMatchingPair() {
		if (!isMatching || fMatchingPairs.length >= questionTypeConfig.maxOptions) return;
		fMatchingPairs = [...fMatchingPairs, { left: '', right: '' }];
		fAnswerKey = buildMatchingAnswerKey(fMatchingPairs.length);
	}

	function removeLastMatchingPair() {
		if (!isMatching || fMatchingPairs.length <= questionTypeConfig.minOptions) return;
		fMatchingPairs = fMatchingPairs.slice(0, -1);
		fAnswerKey = buildMatchingAnswerKey(fMatchingPairs.length);
	}

	function addMatchingDistractor() {
		if (!isMatching || fMatchingDistractors.length >= MAX_MATCHING_DISTRACTOR_COUNT) return;
		fMatchingDistractors = [...fMatchingDistractors, ''];
	}

	function removeMatchingDistractor(index: number) {
		fMatchingDistractors = fMatchingDistractors.filter((_, i) => i !== index);
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

	function revisionSourceLabel(q: Question): string {
		const note = (q.review_notes ?? '').toLowerCase();
		if (note.includes('analisis butir')) return 'Analisis Butir';
		if ((q.reviewer_username ?? '').trim()) return `Reviewer: ${q.reviewer_username}`;
		return 'Workflow Review';
	}

	function revisionReason(q: Question): string {
		const note = (q.review_notes ?? '').replace(/\s+/g, ' ').trim();
		if (!note) return 'Belum ada catatan alasan revisi.';
		return note.length > 180 ? `${note.slice(0, 180)}...` : note;
	}

	function canSubmitRevisionReview(q: Question): boolean {
		return q.workflow_status === 'rejected' && q.status === 'draft' && !questionUsageLocked(q);
	}

	function canDecideReview(q: Question): boolean {
		return canReviewWorkflow && q.workflow_status === 'review' && q.status === 'draft' && !questionUsageLocked(q);
	}

	function canPublishQuestion(q: Question): boolean {
		return canReviewWorkflow && q.workflow_status === 'approved' && q.status === 'draft' && !questionUsageLocked(q);
	}

	function isQuickEditable(q: Question): boolean {
		return isComposerQuestionType(q.question_type) && (q.workflow_status === 'draft' || q.workflow_status === 'rejected') && q.status === 'draft' && !questionUsageLocked(q);
	}

	function explainQuickEditBlocked(q: Question): string {
		if (questionUsageLocked(q)) return 'Soal sudah dipakai. Gunakan Duplikat untuk membuat revisi draft.';
		if ((q.workflow_status !== 'draft' && q.workflow_status !== 'rejected') || q.status !== 'draft') return 'Soal sudah masuk alur review/publikasi. Gunakan Duplikat untuk revisi.';
		if (!isComposerQuestionType(q.question_type)) return 'Tipe soal ini belum masuk komposer utama. Gunakan Duplikat setelah tipe ini dimigrasikan.';
		return 'Buat revisi lewat Duplikat agar riwayat soal tetap aman.';
	}

	function openQuestion(q: Question) {
		if (isQuickEditable(q)) {
			void openEdit(q);
			return;
		}
		toast.warning(explainQuickEditBlocked(q));
	}

	async function openQuestionFromRouteParam(id: string) {
		try {
			const res = await fetch(clientApiPath`/api/cbt/questions/${id}`);
			const q = await readClientApiData<Question>(res, 'Gagal memuat detail soal');
			openQuestion(q);
		} catch (error) {
			toast.error(mutationErrorMessage(error, 'Gagal membuka soal dari tautan lama.'));
		}
	}

	function resetForm() {
		fSubjectId = '';
		fQuestionType = 'multiple_choice';
		fAuthoringMode = 'beginner';
		fStem = '';
		fStimulus = '';
		fRubric = '';
		fExplanation = '';
		fOptions = normalizeOptionCount([], fQuestionType);
		fMatchingPairs = normalizeMatchingPairs([], fQuestionType);
		fMatchingDistractors = normalizeMatchingDistractors([], fQuestionType);
		fAnswerKey = defaultAnswerKeyForQuestionType(fQuestionType);
		fWeight = 1;
		fDifficulty = 'medium';
		fIsRtl = false;
		fGradeLevel = 7;
		fAcademicPhase = '';
		fCPRef = '';
		fTPRef = '';
		fKDRef = '';
		fIndicatorRef = '';
		fMaterialTopic = '';
		fCognitiveLevel = '';
		fHotsFlag = false;
		fWorkflowStatus = 'draft';
		draftStatus = '';
		draftSavedAt = null;
		lastDraftSig = '';
	}

	function openCreate() {
		editingId = null;
		resetForm();
		showInspector = false;
		focusedEditor = null;
		composerMobilePanel = 'write';
		showComposer = true;
		// Delay to let state settle before restoring
		setTimeout(() => {
			if (!restoreDraft()) draftStatus = '';
		}, 50);
	}

	async function openEdit(q: Question) {
		if (!isQuickEditable(q)) {
			setModuleMode('catalog');
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
			fQuestionType = normalizeQuestionType(d.question_type);
			fAuthoringMode = normalizeAuthoringMode(d.suggested_mode ?? d.authoring_mode);
			fStem = d.stem_html || d.question_text || '';
			fStimulus = d.stimulus_html ?? '';
			fRubric = d.rubric_html ?? '';
			fExplanation = d.explanation_html ?? '';
			const opts = d.options?.length
				? d.options.map((o) => o.html || o.text || o.latex || '')
				: defaultOptionsForQuestionType(fQuestionType);
			fOptions = normalizeOptionCount(opts, fQuestionType);
			fMatchingPairs = normalizeMatchingPairs(optionsToMatchingPairs(d.options ?? []), fQuestionType);
			fMatchingDistractors = normalizeMatchingDistractors(optionsToMatchingDistractors(d.options ?? []), fQuestionType);
			fAnswerKey = normalizeAnswerKey(d.answer_key, fQuestionType, answerItemCountForType(fQuestionType));
			fDifficulty = d.difficulty || 'medium';
			fGradeLevel = d.grade_level ?? 7;
			fAcademicPhase = d.academic_phase ?? '';
			fCPRef = d.cp_ref ?? '';
			fTPRef = d.tp_ref ?? '';
			fKDRef = d.kd_ref ?? '';
			fIndicatorRef = d.indicator_ref ?? '';
			fMaterialTopic = d.material_topic ?? '';
			fCognitiveLevel = d.cognitive_level ?? '';
			fHotsFlag = d.hots_flag ?? false;
			fWorkflowStatus = normalizeWorkflowStatus(d.workflow_status);
			showInspector = false;
			focusedEditor = null;
			composerMobilePanel = 'write';
			showComposer = true;
		} catch (error) {
			toast.error(mutationErrorMessage(error, 'Gagal memuat detail soal. Form memakai data ringkas dari daftar.'));
			editingId = q.id;
			resetForm();
			fSubjectId = q.subject_id;
			fQuestionType = normalizeQuestionType(q.question_type);
			fAuthoringMode = normalizeAuthoringMode(q.suggested_mode ?? q.authoring_mode);
			fStem = q.stem_html || q.question_text || '';
			fStimulus = q.stimulus_html ?? '';
			fRubric = q.rubric_html ?? '';
			fExplanation = q.explanation_html ?? '';
			fOptions = normalizeOptionCount(q.options?.map((o) => o.html || o.text || o.latex || '') ?? defaultOptionsForQuestionType(fQuestionType), fQuestionType);
			fMatchingPairs = normalizeMatchingPairs(optionsToMatchingPairs(q.options ?? []), fQuestionType);
			fMatchingDistractors = normalizeMatchingDistractors(optionsToMatchingDistractors(q.options ?? []), fQuestionType);
			fAnswerKey = normalizeAnswerKey(q.answer_key, fQuestionType, answerItemCountForType(fQuestionType));
			fGradeLevel = q.grade_level ?? 7;
			fAcademicPhase = q.academic_phase ?? '';
			fCPRef = q.cp_ref ?? '';
			fTPRef = q.tp_ref ?? '';
			fKDRef = q.kd_ref ?? '';
			fIndicatorRef = q.indicator_ref ?? '';
			fMaterialTopic = q.material_topic ?? '';
			fCognitiveLevel = q.cognitive_level ?? '';
			fHotsFlag = q.hots_flag ?? false;
			fWorkflowStatus = normalizeWorkflowStatus(q.workflow_status);
			showInspector = false;
			focusedEditor = null;
			composerMobilePanel = 'write';
			showComposer = true;
		} finally {
			composerBusy = false;
		}
	}

	function closeComposer() {
		showComposer = false;
		editingId = null;
		focusedEditor = null;
		showInspector = false;
	}

	async function requestCloseComposer() {
		if (hasDraftWork && draftStatus) {
			const confirmed = await confirmAction({
				title: 'Tutup Komposer?',
				message: 'Draft lokal tetap disimpan otomatis. Tutup modal dan lanjutkan nanti?',
				confirmLabel: 'Tutup',
				tone: 'warning'
			});
			if (!confirmed) return;
		}
		closeComposer();
	}

	function scrollComposerSection(id: string) {
		document.getElementById(id)?.scrollIntoView({ behavior: 'smooth', block: 'start' });
	}

	function openInspector() {
		showInspector = true;
		composerMobilePanel = 'preview';
	}

	function draftStatusLabel() {
		if (draftSavedAt) {
			const saved = new Date(draftSavedAt);
			const clock = Number.isNaN(saved.getTime())
				? ''
				: saved.toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' });
			return `${draftStatus || 'Draft tersimpan'}${clock ? ` ${clock}` : ''}`;
		}
		return draftStatus || 'Editor legacy siap untuk buat/edit soal.';
	}

	function focusTitle(editor: FocusedEditor) {
		if (editor === 'stem') return 'Isi Pertanyaan';
		if (editor === 'stimulus') return 'Stimulus';
		if (editor === 'rubric') return 'Rubrik Penilaian';
		if (editor === 'explanation') return 'Pembahasan';
		return `Opsi ${editor}`;
	}

	function handleComposerKeydown(event: KeyboardEvent) {
		if (!showComposer) return;
		if (event.key === 'Escape' && focusedEditor) {
			event.preventDefault();
			focusedEditor = null;
			return;
		}
		if ((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === 's') {
			event.preventDefault();
			if (canSaveDraft) {
				void saveQuestion('draft');
				return;
			}
			toast.warning(draftIssues[0] ?? 'Lengkapi draft sebelum menyimpan');
		}
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

	function buildPayloadOptions() {
		const config = getQuestionTypeConfig(fQuestionType);
		if (config.answerMode === 'fixed_pair') {
			return (config.fixedOptions ?? []).map((text, i) => ({
				label: optionLabelAt(i),
				text,
				html: text,
			}));
		}
		if (config.answerMode === 'matching') {
			return [
				...fMatchingPairs.map((pair, i) => ({
					label: optionLabelAt(i),
					text: htmlToPlainText(pair.left),
					html: pair.left,
					match_label: String(i + 1),
					match_text: htmlToPlainText(pair.right),
					match_html: pair.right,
					is_distractor: false,
				})),
				...fMatchingDistractors.map((html, i) => ({
					label: '',
					text: '',
					html: '',
					match_label: String(fMatchingPairs.length + i + 1),
					match_text: htmlToPlainText(html),
					match_html: html,
					is_distractor: true,
				})),
			];
		}
		if (config.answerMode !== 'single_option' && config.answerMode !== 'multi_option') {
			return [];
		}
		return fOptions.map((html, i) => ({
			label: optionLabelAt(i),
			text: htmlToPlainText(html),
			html,
		}));
	}

	function buildPayloadAnswerKey(): string {
		const config = getQuestionTypeConfig(fQuestionType);
		if (config.answerMode === 'rubric') return '';
		if (config.answerMode === 'short_text') return normalizeAnswerKey(fAnswerKey, fQuestionType, fOptions.length);
		return normalizeAnswerKey(fAnswerKey, fQuestionType, answerItemCountForType(fQuestionType));
	}

	async function saveQuestion(intent: ComposerSaveIntent = 'draft') {
		const isReview = intent === 'review';
		if (isReview && !canSubmitReview) {
			toast.warning(validationIssues[0] ?? 'Lengkapi soal sebelum diajukan review');
			return;
		}
		if (!isReview && !canSaveDraft) {
			toast.warning(draftIssues[0] ?? 'Lengkapi draft sebelum menyimpan');
			return;
		}
		if (editingId) {
			const current = questions.find((item) => item.id === editingId);
			if (current && !isQuickEditable(current)) {
				toast.error('Soal ini tidak aman diedit lewat komposer cepat. Gunakan Duplikat untuk revisi.');
				return;
			}
		}
		composerBusy = true;
		composerAction = intent;
		try {
			const payload = {
				authoring_mode: fAuthoringMode,
				subject_id: fSubjectId,
				question_text: htmlToPlainText(fStem),
				question_type: fQuestionType,
				stem_html: fStem,
				stimulus_html: fStimulus,
				explanation_html: fExplanation,
				rubric_html: requiresRubric ? fRubric : '',
				options: buildPayloadOptions(),
				answer_key: buildPayloadAnswerKey(),
				difficulty: fDifficulty,
				status: 'draft',
				workflow_status: isReview ? 'review' : 'draft',
				grade_level: fGradeLevel,
				academic_phase: fAcademicPhase,
				cp_ref: fCPRef,
				tp_ref: fTPRef,
				kd_ref: fKDRef,
				indicator_ref: fIndicatorRef,
				material_topic: fMaterialTopic,
				cognitive_level: fCognitiveLevel,
				hots_flag: fHotsFlag,
				writer_notes: isAdvanceMode ? 'Disusun dari komposer soal mode advance.' : '',
				review_notes: '',
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
			toast.success(isReview ? 'Soal diajukan review' : editingId ? 'Draft soal berhasil diperbarui' : 'Draft soal berhasil dibuat');
			closeComposer();
			await refreshOverview(1);
		} catch (e) {
			toast.error(mutationErrorMessage(e, 'Gagal menyimpan soal'));
		} finally {
			composerBusy = false;
			composerAction = '';
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

	function exportFilenameFromResponse(response: Response) {
		const disposition = response.headers.get('content-disposition') ?? '';
		const match = /filename="?([^";]+)"?/i.exec(disposition);
		return match?.[1] ?? `bank-soal-cbt-${new Date().toISOString().slice(0, 10)}.csv`;
	}

	async function downloadCSVResponse(response: Response, fallbackFilename: string) {
		if (!response.ok) {
			const payload = await response.json().catch(() => null) as { error?: string; message?: string } | null;
			throw new Error(payload?.error ?? payload?.message ?? 'Download CSV gagal');
		}
		const blob = await response.blob();
		const url = URL.createObjectURL(blob);
		const link = document.createElement('a');
		link.href = url;
		link.download = exportFilenameFromResponse(response) || fallbackFilename;
		link.click();
		URL.revokeObjectURL(url);
	}

	async function exportQuestionsCSV() {
		exportBusy = true;
		try {
			const params = buildQuestionParams(1);
			params.delete('limit');
			params.delete('offset');
			const response = await fetch(clientApiPathWithQuery('/api/cbt/questions/export', params));
			await downloadCSVResponse(response, `bank-soal-cbt-${new Date().toISOString().slice(0, 10)}.csv`);
			toast.success('Export CSV bank soal berhasil dibuat');
		} catch (error) {
			toast.error(mutationErrorMessage(error, 'Export CSV gagal'));
		} finally {
			exportBusy = false;
		}
	}

	async function downloadQuestionsTemplateCSV() {
		templateBusy = true;
		try {
			const response = await fetch('/api/cbt/questions/template');
			await downloadCSVResponse(response, 'template-bank-soal-cbt.csv');
			toast.success('Template CSV bank soal berhasil diunduh');
		} catch (error) {
			toast.error(mutationErrorMessage(error, 'Download template CSV gagal'));
		} finally {
			templateBusy = false;
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

	async function submitRevisionForReview(q: Question) {
		if (!canSubmitRevisionReview(q)) {
			toast.warning('Revisi ini belum aman diajukan review ulang.');
			return;
		}
		if (!(await confirmAction({
			title: 'Ajukan Review Ulang',
			message: 'Ajukan revisi soal ini ke reviewer? Pastikan isi, kunci/rubrik, dan metadata sudah diperbaiki.',
			confirmLabel: 'Ajukan Review',
			tone: 'warning'
		}))) return;
		workflowBusyId = q.id;
		try {
			const res = await fetch(clientApiPath`/api/cbt/questions/${q.id}/workflow`, {
				method: 'PATCH',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ action: 'submit_review', notes: '' }),
			});
			await readClientJson<unknown>(res);
			toast.success('Revisi diajukan review ulang');
			await refreshOverview(currentPage);
		} catch (e) {
			toast.error(mutationErrorMessage(e, 'Gagal mengajukan review ulang'));
		} finally {
			workflowBusyId = '';
		}
	}

	function openReviewDecision(q: Question, decision: ReviewDecision) {
		if (!canReviewWorkflow) {
			toast.warning('Hanya admin yang dapat memutuskan review soal.');
			return;
		}
		if (q.workflow_status !== 'review' || q.status !== 'draft') {
			toast.warning('Soal ini tidak sedang menunggu review.');
			return;
		}
		if (questionUsageLocked(q)) {
			toast.warning('Soal sudah dipakai. Buat revisi baru sebelum mengubah keputusan review.');
			return;
		}
		reviewDecisionQuestion = q;
		reviewDecision = decision;
		reviewDecisionNotes = '';
		reviewDecisionOpen = true;
	}

	function closeReviewDecision() {
		reviewDecisionOpen = false;
		reviewDecisionQuestion = null;
		reviewDecisionNotes = '';
		reviewDecision = 'approve';
	}

	async function submitReviewDecision() {
		const q = reviewDecisionQuestion;
		if (!q) return;
		if (!canDecideReview(q)) {
			toast.warning('Keputusan review tidak tersedia untuk soal ini.');
			return;
		}
		const notes = reviewDecisionNotes.trim();
		if (reviewDecision === 'reject' && notes.length < 8) {
			toast.warning('Catatan revisi minimal 8 karakter agar guru tahu yang harus diperbaiki.');
			return;
		}
		workflowBusyId = q.id;
		try {
			const res = await fetch(clientApiPath`/api/cbt/questions/${q.id}/workflow`, {
				method: 'PATCH',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ action: reviewDecision, notes }),
			});
			await readClientJson<unknown>(res);
			toast.success(reviewDecision === 'approve' ? 'Soal disetujui' : 'Soal dikembalikan untuk revisi');
			closeReviewDecision();
			await refreshOverview(currentPage);
		} catch (e) {
			toast.error(mutationErrorMessage(e, 'Gagal menyimpan keputusan review'));
		} finally {
			workflowBusyId = '';
		}
	}

	async function publishQuestion(q: Question) {
		if (!canPublishQuestion(q)) {
			toast.warning('Soal ini belum siap diterbitkan.');
			return;
		}
		if (!(await confirmAction({
			title: 'Terbitkan Soal',
			message: 'Terbitkan soal ini agar bisa dipilih ke paket ujian CBT? Setelah diterbitkan, perubahan isi harus melalui duplikasi/revisi.',
			confirmLabel: 'Terbitkan',
			tone: 'warning'
		}))) return;
		workflowBusyId = q.id;
		try {
			const res = await fetch(clientApiPath`/api/cbt/questions/${q.id}/workflow`, {
				method: 'PATCH',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ action: 'publish', notes: '' }),
			});
			await readClientJson<unknown>(res);
			toast.success('Soal diterbitkan dan siap masuk paket ujian');
			await refreshOverview(currentPage);
		} catch (e) {
			toast.error(mutationErrorMessage(e, 'Gagal menerbitkan soal'));
		} finally {
			workflowBusyId = '';
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
		const questionId = params.get('question_id');
		if (validModuleMode(mode)) activeMode = mode;
		if (questionId) {
			params.delete('question_id');
			const query = params.toString();
			window.history.replaceState({}, '', query ? `${window.location.pathname}?${query}` : window.location.pathname);
		}
		load();
		if (questionId) void openQuestionFromRouteParam(questionId);
		window.addEventListener('keydown', handleComposerKeydown);
		return () => {
			window.removeEventListener('keydown', handleComposerKeydown);
		};
	});
</script>

<!-- ── Main page ──────────────────────────────────────────────────────────── -->
<div class="space-y-4">
	<!-- Header -->
	<div class="flex flex-wrap items-start justify-between gap-3">
		<div>
			<h1 class="text-xl font-semibold text-slate-800">Bank Soal CBT</h1>
			<p class="text-sm text-slate-500 mt-0.5">
				Satu modul untuk katalog, komposer terpadu, editor lanjutan, review, dan import CSV
			</p>
		</div>
		<div class="flex shrink-0 flex-wrap gap-2">
			<LoadingButton
				variant="outline"
				onclick={() => void exportQuestionsCSV()}
				loading={exportBusy}
				loadingLabel="Export..."
				disabled={exportBusy || totalItems === 0}
			>
				Export CSV
			</LoadingButton>
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

	{#if revisionTotal > 0 || revisionSourceFilter}
		<section class="rounded-lg border border-red-200 bg-red-50/60 p-3">
			<div class="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
				<div class="min-w-0">
					<div class="flex flex-wrap items-center gap-2">
						<h2 class="text-sm font-bold uppercase tracking-wider text-red-900">Antrian Revisi Soal</h2>
						<span class="rounded-full bg-white px-2 py-0.5 text-xs font-semibold text-red-700">{revisionTotal} {revisionSourceFilter ? 'sesuai filter' : 'perlu diperbaiki'}</span>
					</div>
					<p class="mt-1 text-xs text-red-800">Draft revisi dari analisis butir atau workflow reviewer. Buka, koreksi isi/kunci/rubrik, lalu simpan sebagai draft atau ajukan review lagi.</p>
				</div>
				<Button variant="outline" size="sm" class="shrink-0 border-red-200 bg-white text-red-800 hover:bg-red-100" onclick={showAllRevisions}>
					Lihat Semua Revisi
				</Button>
			</div>
			<div class="mt-3 flex flex-wrap gap-1.5">
				{#each revisionSourceOptions as source (source.id)}
					<button
						type="button"
						onclick={() => setRevisionSourceFilter(source.id)}
						class="rounded-md border px-2.5 py-1.5 text-left text-xs transition-colors {revisionSourceFilter === source.id
							? 'border-red-300 bg-white font-semibold text-red-800 shadow-sm'
							: 'border-red-100 bg-red-50 text-red-700 hover:bg-white'}"
					>
						<span>{source.label}</span>
						<span class="ml-1 text-[10px] font-normal opacity-70">{source.desc}</span>
					</button>
				{/each}
			</div>
			{#if revisionQueue.length > 0}
				<div class="mt-3 grid gap-2 md:grid-cols-2 xl:grid-cols-3">
					{#each revisionQueue as q (q.id)}
						<article class="min-w-0 rounded-md border border-red-100 bg-white px-3 py-2 text-left shadow-sm transition-colors hover:border-red-300 hover:bg-red-50">
							<div class="mb-1 flex items-center gap-1">
								<span class="rounded bg-red-50 px-1.5 py-0.5 text-[10px] font-semibold text-red-700">{questionTypeLabel(q.question_type)}</span>
								<span class="truncate text-[11px] text-slate-400">{q.subject_name || q.subject_code || 'Mapel belum ada'}</span>
							</div>
							<button type="button" onclick={() => openQuestion(q)} class="block w-full text-left">
								<p class="line-clamp-2 text-sm font-medium text-slate-800 hover:text-green-800">{stemPreview(q)}</p>
							</button>
							<div class="mt-2 rounded border border-red-100 bg-red-50/70 px-2 py-1.5">
								<div class="mb-0.5 flex flex-wrap items-center gap-1">
									<span class="rounded bg-white px-1.5 py-0.5 text-[10px] font-semibold uppercase tracking-wider text-red-700">{revisionSourceLabel(q)}</span>
									{#if q.reviewed_at}<span class="text-[10px] text-red-500">{new Date(q.reviewed_at).toLocaleDateString('id-ID')}</span>{/if}
								</div>
								<p class="line-clamp-2 text-[11px] leading-relaxed text-red-900">{revisionReason(q)}</p>
							</div>
							<div class="mt-1 flex flex-wrap items-center gap-x-2 gap-y-1 text-[11px] text-slate-400">
								<span>{q.code || 'Tanpa kode'}</span>
								{#if q.author_username}<span>{q.author_username}</span>{/if}
								<span>{DIFFICULTY_LABEL[q.difficulty] ?? q.difficulty ?? 'Sedang'}</span>
							</div>
							<div class="mt-2 flex flex-wrap gap-1.5">
								<Button variant="outline" size="sm" class="h-7 bg-white text-xs" onclick={() => openQuestion(q)}>
									Edit Revisi
								</Button>
								<LoadingButton
									variant="outline"
									size="sm"
									class="h-7 border-green-200 bg-green-50 text-xs text-green-800 hover:bg-green-100"
									onclick={() => void submitRevisionForReview(q)}
									loading={workflowBusyId === q.id}
									loadingLabel="Mengajukan..."
									disabled={!canSubmitRevisionReview(q) || (workflowBusyId !== '' && workflowBusyId !== q.id)}
								>
									Ajukan Review Ulang
								</LoadingButton>
							</div>
						</article>
					{/each}
				</div>
			{:else}
				<div class="mt-3 rounded-md border border-red-100 bg-white px-3 py-3 text-sm text-red-800">
					Tidak ada revisi pada sumber ini.
				</div>
			{/if}
		</section>
	{/if}

	{#if activeMode === 'review' || (canReviewWorkflow && reviewTotal > 0)}
		<section class="rounded-lg border border-amber-200 bg-amber-50/70 p-3">
			<div class="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
				<div class="min-w-0">
					<div class="flex flex-wrap items-center gap-2">
						<h2 class="text-sm font-bold uppercase tracking-wider text-amber-950">Antrian Review Soal</h2>
						<span class="rounded-full bg-white px-2 py-0.5 text-xs font-semibold text-amber-800">{reviewTotal} menunggu keputusan</span>
					</div>
					<p class="mt-1 text-xs text-amber-900">Meja kerja reviewer untuk memeriksa soal yang diajukan guru, menyetujui, atau mengembalikan dengan catatan revisi.</p>
				</div>
				<Button variant="outline" size="sm" class="shrink-0 border-amber-200 bg-white text-amber-900 hover:bg-amber-100" onclick={showPendingReviews}>
					Lihat Semua Ditinjau
				</Button>
			</div>
			{#if reviewQueue.length > 0}
				<div class="mt-3 grid gap-2 md:grid-cols-2 xl:grid-cols-3">
					{#each reviewQueue as q (q.id)}
						<article class="min-w-0 rounded-md border border-amber-100 bg-white px-3 py-2 text-left shadow-sm transition-colors hover:border-amber-300 hover:bg-amber-50">
							<div class="mb-1 flex items-center gap-1">
								<span class="rounded bg-amber-50 px-1.5 py-0.5 text-[10px] font-semibold text-amber-800">{questionTypeLabel(q.question_type)}</span>
								<span class="truncate text-[11px] text-slate-400">{q.subject_name || q.subject_code || 'Mapel belum ada'}</span>
							</div>
							<button type="button" onclick={() => openReviewDecision(q, 'approve')} class="block w-full text-left" disabled={!canReviewWorkflow}>
								<p class="line-clamp-2 text-sm font-medium text-slate-800 hover:text-green-800">{stemPreview(q)}</p>
							</button>
							<div class="mt-1 flex flex-wrap items-center gap-x-2 gap-y-1 text-[11px] text-slate-400">
								<span>{q.code || 'Tanpa kode'}</span>
								{#if q.author_username}<span>Guru: {q.author_username}</span>{/if}
								<span>{DIFFICULTY_LABEL[q.difficulty] ?? q.difficulty ?? 'Sedang'}</span>
							</div>
							{#if q.review_notes}
								<div class="mt-2 rounded border border-amber-100 bg-amber-50/70 px-2 py-1.5 text-[11px] leading-relaxed text-amber-900">
									<span class="font-semibold">Catatan sebelumnya:</span> {revisionReason(q)}
								</div>
							{/if}
							<div class="mt-2 flex flex-wrap gap-1.5">
								<Button
									variant="outline"
									size="sm"
									class="h-7 border-green-200 bg-green-50 text-xs text-green-800 hover:bg-green-100"
									onclick={() => openReviewDecision(q, 'approve')}
									disabled={!canDecideReview(q) || (workflowBusyId !== '' && workflowBusyId !== q.id)}
								>
									Setujui
								</Button>
								<Button
									variant="outline"
									size="sm"
									class="h-7 border-red-200 bg-red-50 text-xs text-red-700 hover:bg-red-100"
									onclick={() => openReviewDecision(q, 'reject')}
									disabled={!canDecideReview(q) || (workflowBusyId !== '' && workflowBusyId !== q.id)}
								>
									Minta Revisi
								</Button>
							</div>
							{#if !canReviewWorkflow}
								<p class="mt-2 text-[11px] text-amber-800">Menunggu keputusan admin/reviewer.</p>
							{/if}
						</article>
					{/each}
				</div>
			{:else}
				<div class="mt-3 rounded-md border border-amber-100 bg-white px-3 py-3 text-sm text-amber-900">
					Belum ada soal yang menunggu review pada filter ini.
				</div>
			{/if}
		</section>
	{/if}

	{#if activeMode === 'review' || (canReviewWorkflow && approvedTotal > 0)}
		<section class="rounded-lg border border-green-200 bg-green-50/70 p-3">
			<div class="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
				<div class="min-w-0">
					<div class="flex flex-wrap items-center gap-2">
						<h2 class="text-sm font-bold uppercase tracking-wider text-green-950">Siap Terbit ke Paket</h2>
						<span class="rounded-full bg-white px-2 py-0.5 text-xs font-semibold text-green-800">{approvedTotal} disetujui</span>
					</div>
					<p class="mt-1 text-xs text-green-900">Soal sudah lolos review tetapi belum berstatus terbit, sehingga belum bisa dipilih di pembuat paket ujian.</p>
				</div>
				<Button variant="outline" size="sm" class="shrink-0 border-green-200 bg-white text-green-900 hover:bg-green-100" onclick={showApprovedQuestions}>
					Lihat Semua Disetujui
				</Button>
			</div>
			{#if approvedQueue.length > 0}
				<div class="mt-3 grid gap-2 md:grid-cols-2 xl:grid-cols-3">
					{#each approvedQueue as q (q.id)}
						<article class="min-w-0 rounded-md border border-green-100 bg-white px-3 py-2 text-left shadow-sm transition-colors hover:border-green-300 hover:bg-green-50">
							<div class="mb-1 flex items-center gap-1">
								<span class="rounded bg-green-50 px-1.5 py-0.5 text-[10px] font-semibold text-green-800">{questionTypeLabel(q.question_type)}</span>
								<span class="truncate text-[11px] text-slate-400">{q.subject_name || q.subject_code || 'Mapel belum ada'}</span>
							</div>
							<p class="line-clamp-2 text-sm font-medium text-slate-800">{stemPreview(q)}</p>
							<div class="mt-1 flex flex-wrap items-center gap-x-2 gap-y-1 text-[11px] text-slate-400">
								<span>{q.code || 'Tanpa kode'}</span>
								{#if q.author_username}<span>Guru: {q.author_username}</span>{/if}
								{#if q.reviewer_username}<span>Reviewer: {q.reviewer_username}</span>{/if}
							</div>
							<div class="mt-2 flex flex-wrap items-center gap-1.5">
								<LoadingButton
									variant="outline"
									size="sm"
									class="h-7 border-green-200 bg-green-700 text-xs text-white hover:bg-green-800"
									onclick={() => void publishQuestion(q)}
									loading={workflowBusyId === q.id}
									loadingLabel="Menerbitkan..."
									disabled={!canPublishQuestion(q) || (workflowBusyId !== '' && workflowBusyId !== q.id)}
								>
									Terbitkan
								</LoadingButton>
								{#if !canReviewWorkflow}
									<span class="text-[11px] text-green-800">Menunggu admin menerbitkan.</span>
								{/if}
							</div>
						</article>
					{/each}
				</div>
			{:else}
				<div class="mt-3 rounded-md border border-green-100 bg-white px-3 py-3 text-sm text-green-900">
					Belum ada soal disetujui yang menunggu terbit.
				</div>
			{/if}
		</section>
	{/if}

	{#if activeMode === 'composer'}
		<section class="rounded-lg border border-green-200 bg-green-50 p-4">
			<div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
				<div>
					<h2 class="text-sm font-bold uppercase tracking-wider text-green-900">Komposer Soal PG & Essay</h2>
					<p class="mt-1 text-sm text-green-800">Buat draft pilihan ganda atau essay dengan mode Pemula dan Advance. Soal yang sudah review/publish atau dipakai paket wajib direvisi lewat duplikasi.</p>
				</div>
				<Button onclick={openCreate} class="bg-green-700 text-white hover:bg-green-800">Buka Komposer</Button>
			</div>
		</section>
	{:else if activeMode === 'review'}
		<section class="grid gap-3 md:grid-cols-6">
			<div class="rounded-lg border border-slate-200 bg-white p-4">
				<p class="text-xs uppercase tracking-wider text-slate-500">Draft</p>
				<p class="mt-2 text-2xl font-bold text-slate-900">{draftCount}</p>
			</div>
			<div class="rounded-lg border border-red-200 bg-red-50 p-4">
				<p class="text-xs uppercase tracking-wider text-red-700">Perlu Revisi</p>
				<p class="mt-2 text-2xl font-bold text-red-900">{revisionTotal}</p>
				<p class="mt-1 text-[11px] text-red-700">{visibleRevisionCount} tampil di halaman ini</p>
			</div>
			<div class="rounded-lg border border-yellow-200 bg-yellow-50 p-4">
				<p class="text-xs uppercase tracking-wider text-yellow-700">Menunggu Review</p>
				<p class="mt-2 text-2xl font-bold text-yellow-900">{reviewCount}</p>
				<p class="mt-1 text-[11px] text-yellow-700">{visibleReviewCount} tampil di halaman ini</p>
			</div>
			<div class="rounded-lg border border-green-200 bg-green-50 p-4">
				<p class="text-xs uppercase tracking-wider text-green-700">Siap Terbit</p>
				<p class="mt-2 text-2xl font-bold text-green-900">{approvedCount}</p>
				<p class="mt-1 text-[11px] text-green-700">{visibleApprovedCount} tampil di halaman ini</p>
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
					<h2 class="text-sm font-bold uppercase tracking-wider text-slate-800">Import CSV Bank Soal</h2>
					<p class="mt-1 text-sm text-slate-500">Mendukung PG lama serta kolom tipe untuk PG Kompleks, Benar/Salah, Setuju/Tidak Setuju, Isian, Essay, dan Menjodohkan.</p>
				</div>
				<div class="flex flex-wrap gap-2">
					<LoadingButton
						variant="outline"
						onclick={() => void downloadQuestionsTemplateCSV()}
						loading={templateBusy}
						loadingLabel="Mengunduh..."
					>
						Template CSV
					</LoadingButton>
					<Button variant="outline" onclick={openImport}>Pilih CSV</Button>
				</div>
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
			onchange={onWorkflowFilterChange}
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
									onclick={() => (canDecideReview(q) ? openReviewDecision(q, 'approve') : openQuestion(q))}
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
											<span class="rounded bg-green-50 px-1.5 py-0.5 text-[10px] font-medium text-green-700">
												{questionTypeLabel(q.question_type)}
											</span>
											<span class="rounded bg-slate-100 px-1.5 py-0.5 text-[10px] font-medium text-slate-500">
												{q.suggested_mode ?? q.authoring_mode ?? 'beginner'}
											</span>
											{#if questionUsageLocked(q)}
												<span class="rounded bg-red-50 px-1.5 py-0.5 text-[10px] font-medium text-red-600">
													Terkunci: {questionUsageText(q)}
												</span>
											{/if}
										</div>
										{#if q.workflow_status === 'rejected'}
											<div class="mt-1 rounded border border-red-100 bg-red-50 px-2 py-1 text-[11px] leading-relaxed text-red-800">
												<span class="font-semibold">{revisionSourceLabel(q)}:</span> {revisionReason(q)}
											</div>
										{/if}
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
										<Table.Cell class="text-sm font-medium text-slate-600">
											{questionTypeLabel(q.question_type)}
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
										{#if q.workflow_status === 'rejected'}
											<button
												onclick={(e) => {
													e.stopPropagation();
													void submitRevisionForReview(q);
												}}
												disabled={!canSubmitRevisionReview(q) || (workflowBusyId !== '' && workflowBusyId !== q.id)}
												class="rounded px-2 py-1 text-xs text-green-700 transition-colors hover:bg-green-50 disabled:cursor-not-allowed disabled:opacity-40"
											>
												{workflowBusyId === q.id ? 'Mengajukan...' : 'Review Ulang'}
											</button>
										{/if}
										{#if canDecideReview(q)}
											<button
												onclick={(e) => {
													e.stopPropagation();
													openReviewDecision(q, 'approve');
												}}
												disabled={workflowBusyId !== '' && workflowBusyId !== q.id}
												class="rounded px-2 py-1 text-xs text-amber-700 transition-colors hover:bg-amber-50 disabled:cursor-not-allowed disabled:opacity-40"
											>
												Review
											</button>
										{/if}
										{#if canPublishQuestion(q)}
											<button
												onclick={(e) => {
													e.stopPropagation();
													void publishQuestion(q);
												}}
												disabled={workflowBusyId !== '' && workflowBusyId !== q.id}
												class="rounded px-2 py-1 text-xs text-green-700 transition-colors hover:bg-green-50 disabled:cursor-not-allowed disabled:opacity-40"
											>
												{workflowBusyId === q.id ? 'Terbit...' : 'Terbitkan'}
											</button>
										{/if}
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
				<h2 class="text-base font-semibold text-slate-800">Import CSV Bank Soal</h2>
				<p class="mt-1 text-xs text-slate-500">Hasil import disimpan sebagai draft. Kolom tipe boleh kosong untuk PG lama, atau diisi: pg_kompleks, benar_salah, setuju_tidak_setuju, isian, essay, menjodohkan.</p>
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

<Dialog.Root bind:open={reviewDecisionOpen}>
	<Dialog.Content>
		<div class="w-[min(94vw,42rem)] space-y-4 p-5">
			<div>
				<p class="text-xs font-bold uppercase tracking-wider text-green-700">Review Soal CBT</p>
				<h2 class="mt-1 text-base font-semibold text-slate-800">
					{reviewDecision === 'approve' ? 'Setujui Soal' : 'Minta Revisi Soal'}
				</h2>
				<p class="mt-1 text-xs text-slate-500">Periksa isi, opsi, kunci/rubrik, dan catatan sebelum menyimpan keputusan.</p>
			</div>

			{#if reviewDecisionQuestion}
				<div class="rounded-lg border border-slate-200 bg-slate-50 p-3">
					<div class="mb-2 flex flex-wrap items-center gap-1.5 text-[11px] text-slate-500">
						<span class="rounded bg-white px-1.5 py-0.5 font-semibold text-green-700">{questionTypeLabel(reviewDecisionQuestion.question_type)}</span>
						<span>{reviewDecisionQuestion.subject_name || reviewDecisionQuestion.subject_code || 'Mapel belum ada'}</span>
						<span>{reviewDecisionQuestion.code || 'Tanpa kode'}</span>
						{#if reviewDecisionQuestion.author_username}<span>Guru: {reviewDecisionQuestion.author_username}</span>{/if}
					</div>
					<div class="max-h-64 space-y-3 overflow-auto rounded-md border border-slate-200 bg-white p-3 text-sm">
						{#if reviewDecisionQuestion.stimulus_html}
							<div class="rounded border border-slate-100 bg-slate-50 p-2">
								<p class="mb-1 text-[10px] font-semibold uppercase tracking-wider text-slate-400">Stimulus</p>
								<RichContent html={reviewDecisionQuestion.stimulus_html} class="prose prose-sm max-w-none text-slate-700 latex-preview" />
							</div>
						{/if}
						<div>
							<p class="mb-1 text-[10px] font-semibold uppercase tracking-wider text-slate-400">Pertanyaan</p>
							<RichContent html={reviewDecisionQuestion.stem_html || reviewDecisionQuestion.question_text || 'Isi soal belum tersedia'} class="prose prose-sm max-w-none text-slate-800 latex-preview" />
						</div>
						{#if reviewDecisionQuestion.options.length > 0}
							<div class="space-y-1.5">
								<p class="text-[10px] font-semibold uppercase tracking-wider text-slate-400">Opsi / Pasangan</p>
								{#each reviewDecisionQuestion.options as opt, i (`review-option-${reviewDecisionQuestion.id}-${i}`)}
									<div class="rounded border border-slate-100 bg-slate-50 px-2 py-1.5">
										<div class="flex gap-2">
											<span class="mt-0.5 text-xs font-bold text-green-700">{opt.label || opt.match_label || i + 1}</span>
											<div class="min-w-0 flex-1 text-xs text-slate-700">
												<RichContent html={opt.html || opt.text || opt.latex || opt.match_html || opt.match_text || '-'} class="latex-preview" />
												{#if opt.match_html || opt.match_text}
													<div class="mt-1 border-t border-slate-200 pt-1 text-slate-500">
														<RichContent html={opt.match_html || opt.match_text || '-'} class="latex-preview" />
													</div>
												{/if}
											</div>
										</div>
									</div>
								{/each}
							</div>
						{/if}
						<div class="rounded border border-green-100 bg-green-50 px-2 py-1.5 text-xs text-green-900">
							<span class="font-semibold">Kunci/Rubrik:</span>
							{reviewDecisionQuestion.question_type === 'essay' ? 'Periksa rubrik uraian pada isi soal.' : 'Periksa kunci dari detail soal sebelum keputusan final.'}
						</div>
						{#if reviewDecisionQuestion.review_notes}
							<div class="rounded border border-amber-100 bg-amber-50 px-2 py-1.5 text-xs leading-relaxed text-amber-900">
								<span class="font-semibold">Catatan sebelumnya:</span> {revisionReason(reviewDecisionQuestion)}
							</div>
						{/if}
					</div>
				</div>
			{/if}

			<div class="grid grid-cols-2 gap-2">
				<button
					type="button"
					onclick={() => (reviewDecision = 'approve')}
					class="rounded-md border px-3 py-2 text-left text-sm transition-colors {reviewDecision === 'approve'
						? 'border-green-300 bg-green-50 font-semibold text-green-800'
						: 'border-slate-200 bg-white text-slate-600 hover:bg-slate-50'}"
				>
					Setujui
					<span class="mt-0.5 block text-[11px] font-normal text-slate-500">Soal masuk status disetujui.</span>
				</button>
				<button
					type="button"
					onclick={() => (reviewDecision = 'reject')}
					class="rounded-md border px-3 py-2 text-left text-sm transition-colors {reviewDecision === 'reject'
						? 'border-red-300 bg-red-50 font-semibold text-red-700'
						: 'border-slate-200 bg-white text-slate-600 hover:bg-slate-50'}"
				>
					Minta Revisi
					<span class="mt-0.5 block text-[11px] font-normal text-slate-500">Kembalikan ke guru dengan catatan.</span>
				</button>
			</div>

			<div>
				<label for="review-decision-notes" class="mb-1 block text-xs font-medium text-slate-600">
					Catatan Reviewer {#if reviewDecision === 'reject'}<span class="text-red-500">*</span>{/if}
				</label>
				<Textarea
					id="review-decision-notes"
					rows={3}
					bind:value={reviewDecisionNotes}
					placeholder={reviewDecision === 'approve' ? 'Opsional: catatan persetujuan.' : 'Tuliskan bagian yang harus diperbaiki guru.'}
				/>
			</div>

			<div class="flex justify-end gap-2 border-t border-slate-100 pt-4">
				<Button variant="outline" onclick={closeReviewDecision} disabled={workflowBusyId !== ''}>
					Batal
				</Button>
				<LoadingButton
					onclick={() => void submitReviewDecision()}
					loading={workflowBusyId !== ''}
					loadingLabel="Menyimpan..."
					disabled={!reviewDecisionQuestion || workflowBusyId !== ''}
					class={reviewDecision === 'approve' ? 'bg-green-700 text-white hover:bg-green-800' : 'bg-red-600 text-white hover:bg-red-700'}
				>
					{reviewDecision === 'approve' ? 'Setujui Soal' : 'Kirim Revisi'}
				</LoadingButton>
			</div>
		</div>
	</Dialog.Content>
</Dialog.Root>

{#snippet composerPreview()}
	<div class="mb-4">
		<div class="mb-1 flex items-center justify-between">
			<span class="text-[10px] font-semibold uppercase tracking-wide text-slate-400">Kesiapan Review</span>
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
			<p class="mt-1 text-[10px] text-green-600">Soal siap diajukan review.</p>
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
			{#if isAdvanceMode && fStimulus}
				<div class="rounded-md border border-slate-100 bg-slate-50 p-2">
					<p class="mb-1 text-[10px] font-semibold uppercase tracking-wide text-slate-400">Stimulus</p>
					<RichContent html={fStimulus} class="prose prose-sm max-w-none text-slate-700 latex-preview text-sm" />
				</div>
			{/if}
			{#if fStem}
				<RichContent
					html={fStem}
					class="prose prose-sm max-w-none text-slate-800 latex-preview text-sm"
				/>
			{:else}
				<p class="text-xs text-slate-400 italic">Isi soal belum dimasukkan</p>
			{/if}

			{#if isEssay}
				<div class="space-y-2 border-t border-slate-100 pt-2">
					<div class="rounded-md border border-dashed border-green-200 bg-green-50 px-3 py-2 text-xs text-green-900">
						Siswa akan melihat kotak jawaban uraian pada aplikasi ujian.
					</div>
					{#if richTextHasContent(fRubric)}
						<div class="rounded-md border border-amber-100 bg-amber-50 p-2">
							<p class="mb-1 text-[10px] font-semibold uppercase tracking-wide text-amber-700">Pedoman koreksi guru</p>
							<RichContent html={fRubric} class="prose prose-sm max-w-none text-amber-950 latex-preview text-sm" />
						</div>
					{/if}
				</div>
			{:else if isShortAnswer}
				<div class="space-y-2 border-t border-slate-100 pt-2">
					<div class="rounded-md border border-dashed border-green-200 bg-green-50 px-3 py-2 text-xs text-green-900">
						Siswa akan melihat kolom jawaban isian singkat.
					</div>
					{#if shortAnswerAliases.length > 0}
						<div class="rounded-md border border-slate-100 bg-slate-50 px-3 py-2 text-xs text-slate-700">
							<span class="font-semibold text-slate-500">Jawaban diterima:</span> {shortAnswerAliases.join(' / ')}
						</div>
					{/if}
				</div>
			{:else if isMatching}
				<div class="space-y-3 border-t border-slate-100 pt-2">
					<div class="rounded-md border border-dashed border-green-200 bg-green-50 px-3 py-2 text-xs text-green-900">
						Siswa akan memilih nomor pasangan kanan untuk setiap item kiri.
					</div>
					<div class="grid grid-cols-1 gap-2 md:grid-cols-2">
						<div class="space-y-1.5">
							<p class="text-[10px] font-semibold uppercase tracking-wide text-slate-400">Kolom Kiri</p>
							{#each fMatchingPairs as pair, i (`preview-match-left-${i}`)}
								<div class="flex items-start gap-2 rounded-md border border-slate-100 bg-white px-2 py-1.5 text-xs">
									<span class="shrink-0 font-bold text-green-700">{optionLabelAt(i)}.</span>
									{#if richTextHasContent(pair.left)}
										<RichContent html={pair.left} class="latex-preview min-w-0 flex-1" />
									{:else}
										<span class="italic text-slate-300">(kiri kosong)</span>
									{/if}
								</div>
							{/each}
						</div>
						<div class="space-y-1.5">
							<p class="text-[10px] font-semibold uppercase tracking-wide text-slate-400">Pilihan Kanan</p>
							{#each fMatchingPairs as pair, i (`preview-match-right-${i}`)}
								<div class="flex items-start gap-2 rounded-md border border-slate-100 bg-slate-50 px-2 py-1.5 text-xs">
									<span class="shrink-0 font-bold text-slate-500">{i + 1}.</span>
									{#if richTextHasContent(pair.right)}
										<RichContent html={pair.right} class="latex-preview min-w-0 flex-1" />
									{:else}
										<span class="italic text-slate-300">(kanan kosong)</span>
									{/if}
								</div>
							{/each}
							{#each fMatchingDistractors as distractor, i (`preview-match-distractor-${i}`)}
								<div class="flex items-start gap-2 rounded-md border border-amber-100 bg-amber-50 px-2 py-1.5 text-xs">
									<span class="shrink-0 font-bold text-amber-700">{fMatchingPairs.length + i + 1}.</span>
									{#if richTextHasContent(distractor)}
										<RichContent html={distractor} class="latex-preview min-w-0 flex-1" />
									{:else}
										<span class="italic text-amber-400">(distraktor kosong)</span>
									{/if}
								</div>
							{/each}
						</div>
					</div>
					<div class="rounded-md border border-slate-100 bg-slate-50 px-3 py-2 text-xs text-slate-700">
						<span class="font-semibold text-slate-500">Kunci otomatis:</span> {buildMatchingAnswerKey(fMatchingPairs.length)}
					</div>
				</div>
			{:else if fOptions.some(richTextHasContent)}
				<div class="space-y-1.5 border-t border-slate-100 pt-2">
					{#each fOptions as opt, i (`preview-option-${i}`)}
						{@const label = optionLabelAt(i)}
						{@const isAnswer = isAnswerLabelSelected(label)}
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
		<div class="soal-composer-modal relative flex h-[94vh] w-[min(96vw,110rem)] max-w-[96vw] flex-col overflow-hidden rounded-2xl bg-white text-slate-900 shadow-2xl">
			<div class="shrink-0 border-b border-green-100 bg-white px-4 py-2.5 md:px-6">
				<div class="flex flex-col gap-2 lg:flex-row lg:items-center lg:justify-between">
					<div class="min-w-0">
						<div class="flex flex-wrap items-baseline gap-x-2 gap-y-1">
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
							onclick={openInspector}
						>
							Preview
						</Button>
						<Button
							type="button"
							variant="outline"
							size="sm"
							class="h-8 text-xs"
							onclick={() => void requestCloseComposer()}
						>
							Tutup
						</Button>
					</div>
				</div>
			</div>

			<div class="min-h-0 flex-1 overflow-y-auto bg-slate-50/70 p-4 md:p-6">
				<section class="mb-4 rounded-xl border border-slate-200 bg-white px-3 py-2 shadow-sm">
					<div class="flex flex-wrap items-center gap-2">
						<div class="flex min-w-[11rem] items-center gap-2">
							<span class="text-[10px] font-black uppercase tracking-wider text-slate-500">Review</span>
							<div class="h-1.5 w-20 overflow-hidden rounded-full bg-slate-200">
								<div
									class="h-1.5 rounded-full transition-all duration-300 {readinessScore === 100
										? 'bg-green-600'
										: readinessScore >= 50
											? 'bg-amber-400'
											: 'bg-red-400'}"
									style="width: {readinessScore}%"
								></div>
							</div>
							<span class="text-xs font-bold {readinessScore === 100 ? 'text-green-700' : 'text-red-500'}">{readinessScore}%</span>
						</div>
						<span class="rounded-full px-2 py-1 text-[10px] font-semibold {validationIssues.length === 0 ? 'bg-green-50 text-green-700' : 'bg-red-50 text-red-600'}">
							{validationIssues.length === 0 ? 'Siap review' : `${validationIssues.length} wajib belum lengkap`}
						</span>
						<span class="rounded-full bg-amber-50 px-2 py-1 text-[10px] font-semibold text-amber-700">
							{qualityWarningCount} sinyal kualitas perlu cek
						</span>
						<span class="rounded-full bg-green-50 px-2 py-1 text-[10px] font-semibold text-green-700">
							{draftStatusLabel()}
						</span>
						<div class="ml-auto flex flex-wrap items-center gap-1">
							<button type="button" onclick={() => scrollComposerSection('composer-metadata')} class="rounded-md px-2 py-1 text-[10px] font-semibold uppercase tracking-wider text-slate-500 hover:bg-slate-50">Metadata</button>
							{#if isAdvanceMode}
								<button type="button" onclick={() => scrollComposerSection('composer-advanced')} class="rounded-md px-2 py-1 text-[10px] font-semibold uppercase tracking-wider text-slate-500 hover:bg-slate-50">Advance</button>
							{/if}
							<button type="button" onclick={() => scrollComposerSection('composer-question')} class="rounded-md px-2 py-1 text-[10px] font-semibold uppercase tracking-wider text-slate-500 hover:bg-slate-50">Pertanyaan</button>
							<button type="button" onclick={() => scrollComposerSection(isEssay ? 'composer-rubric' : isShortAnswer ? 'composer-answer' : isMatching ? 'composer-matching' : 'composer-options')} class="rounded-md px-2 py-1 text-[10px] font-semibold uppercase tracking-wider text-slate-500 hover:bg-slate-50">{isEssay ? 'Rubrik' : isShortAnswer ? 'Kunci' : isMatching ? 'Pasangan' : 'Opsi'}</button>
							<button
								type="button"
								onclick={() => (showInspector = !showInspector)}
								class="rounded-md border px-2 py-1 text-[10px] font-semibold uppercase tracking-wider {showInspector
									? 'border-green-200 bg-green-50 text-green-800'
									: 'border-slate-200 text-slate-600 hover:bg-slate-50'}"
							>
								{showInspector ? 'Tutup Inspector' : 'Inspector'}
							</button>
						</div>
					</div>
				</section>

				<div class={`grid grid-cols-1 gap-5 ${showInspector || composerMobilePanel === 'preview' ? 'xl:grid-cols-[minmax(0,1.55fr)_minmax(18rem,0.5fr)]' : ''}`}>
					<div class={`space-y-4 min-w-0 ${composerMobilePanel === 'preview' ? 'hidden lg:block' : 'block'}`}>
						<section id="composer-metadata" class="scroll-mt-4 rounded-lg border border-slate-200 bg-white p-3 shadow-sm">
							<div class="mb-3 flex flex-wrap items-center gap-2 border-b border-slate-100 pb-3">
								<span class="text-[10px] font-black uppercase tracking-[0.2em] text-slate-800">Bentuk Soal</span>
								<div class="flex flex-wrap rounded-lg border border-slate-200 bg-slate-50 p-1">
									{#each QUESTION_TYPE_CONFIGS as typeConfig (typeConfig.id)}
										<button
											type="button"
											onclick={() => setQuestionType(typeConfig.id)}
											title={typeConfig.desc}
											class="h-7 rounded-md px-2.5 text-[10px] font-bold uppercase tracking-wide {fQuestionType === typeConfig.id ? 'bg-green-700 text-white' : 'text-slate-600 hover:bg-white'}"
										>
											{typeConfig.shortLabel}
										</button>
									{/each}
								</div>
								<span class="ml-2 text-[10px] font-black uppercase tracking-[0.2em] text-slate-800">Mode</span>
								<div class="inline-flex rounded-lg border border-slate-200 bg-slate-50 p-1">
									<button
										type="button"
										onclick={() => setAuthoringMode('beginner')}
										class="h-7 rounded-md px-3 text-[10px] font-bold uppercase tracking-wide {fAuthoringMode === 'beginner' ? 'bg-green-700 text-white' : 'text-slate-600 hover:bg-white'}"
									>
										Pemula
									</button>
									<button
										type="button"
										onclick={() => setAuthoringMode('advance')}
										class="h-7 rounded-md px-3 text-[10px] font-bold uppercase tracking-wide {fAuthoringMode === 'advance' ? 'bg-green-700 text-white' : 'text-slate-600 hover:bg-white'}"
									>
										Advance
									</button>
								</div>
								<span class="text-[11px] text-slate-500">
									<span class="font-semibold text-slate-700">{questionTypeConfig.label}:</span> {questionTypeConfig.desc}
								</span>
							</div>
							<div class="grid gap-2 lg:grid-cols-[8rem_minmax(0,1fr)_8rem_6.5rem_10.5rem] lg:items-end">
								<div class="self-center">
									<h3 class="text-[10px] font-black uppercase tracking-[0.2em] text-slate-800">Metadata</h3>
									<p class="mt-0.5 text-[10px] text-slate-400">Data wajib</p>
								</div>
								<div>
									<div class="mb-1 flex items-center justify-between gap-2">
										<label for="f-subject" class="block text-[10px] font-semibold uppercase tracking-wider text-slate-600">
										Mata Pelajaran <span class="text-red-500">*</span>
									</label>
										{#if !readinessChecks.subject}
											<span class="text-[10px] font-semibold text-red-500">Wajib</span>
										{/if}
									</div>
									<select
										id="f-subject"
										bind:value={fSubjectId}
										class="h-8 w-full rounded-md border border-slate-200 bg-white px-2.5 text-sm font-medium text-slate-700 focus:outline-none focus:ring-2 focus:ring-green-500"
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
										class="h-8 w-full rounded-md border border-slate-200 bg-white px-2.5 text-sm font-medium text-slate-700 focus:outline-none focus:ring-2 focus:ring-green-500"
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
									<Input id="f-weight" type="number" min="1" bind:value={fWeight} class="h-8 text-sm font-medium" />
									{#if !readinessChecks.weight}
										<p class="mt-1 text-[10px] font-semibold text-red-500">Minimal 1.</p>
									{/if}
								</div>
								<label for="f-rtl" class="flex h-8 cursor-pointer items-center justify-between gap-2 rounded-md border border-dashed border-green-200 bg-green-50 px-2.5">
									<span class="text-[10px] font-black uppercase tracking-wider text-green-800">Mode Arab / RTL</span>
									<input id="f-rtl" type="checkbox" bind:checked={fIsRtl} class="rounded accent-green-700" />
								</label>
							</div>
						</section>

						{#if isAdvanceMode}
							<section id="composer-advanced" class="scroll-mt-4 rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
								<div class="mb-3 flex flex-wrap items-center justify-between gap-2">
									<div>
										<h3 class="text-xs font-black uppercase tracking-[0.2em] text-slate-800">Detail Advance</h3>
										<p class="mt-0.5 text-xs text-slate-500">Blueprint, kurikulum, dan alur review.</p>
									</div>
									<label for="f-hots" class="flex h-8 cursor-pointer items-center gap-2 rounded-md border border-green-200 bg-green-50 px-2.5">
										<input id="f-hots" type="checkbox" bind:checked={fHotsFlag} class="rounded accent-green-700" />
										<span class="text-[10px] font-black uppercase tracking-wider text-green-800">HOTS</span>
									</label>
								</div>
								<div class="grid gap-3 md:grid-cols-3 xl:grid-cols-4">
									<div>
										<label for="f-grade-level" class="mb-1 block text-[10px] font-semibold uppercase tracking-wider text-slate-600">Tingkat</label>
										<Input id="f-grade-level" type="number" min="1" max="12" bind:value={fGradeLevel} class="h-8 text-sm" />
									</div>
									<div>
										<label for="f-phase" class="mb-1 block text-[10px] font-semibold uppercase tracking-wider text-slate-600">Fase</label>
										<Input id="f-phase" placeholder="Fase D" bind:value={fAcademicPhase} class="h-8 text-sm" />
									</div>
									<div>
										<label for="f-topic" class="mb-1 block text-[10px] font-semibold uppercase tracking-wider text-slate-600">Topik</label>
										<Input id="f-topic" placeholder="Topik materi" bind:value={fMaterialTopic} class="h-8 text-sm" />
									</div>
									<div>
										<label for="f-cognitive" class="mb-1 block text-[10px] font-semibold uppercase tracking-wider text-slate-600">Kognitif</label>
										<Input id="f-cognitive" placeholder="C3 / HOTS" bind:value={fCognitiveLevel} class="h-8 text-sm" />
									</div>
									<div>
										<label for="f-cp" class="mb-1 block text-[10px] font-semibold uppercase tracking-wider text-slate-600">CP</label>
										<Input id="f-cp" placeholder="CP ref" bind:value={fCPRef} class="h-8 text-sm" />
									</div>
									<div>
										<label for="f-tp" class="mb-1 block text-[10px] font-semibold uppercase tracking-wider text-slate-600">TP</label>
										<Input id="f-tp" placeholder="TP ref" bind:value={fTPRef} class="h-8 text-sm" />
									</div>
									<div>
										<label for="f-kd" class="mb-1 block text-[10px] font-semibold uppercase tracking-wider text-slate-600">KD</label>
										<Input id="f-kd" placeholder="KD ref" bind:value={fKDRef} class="h-8 text-sm" />
									</div>
									<div>
										<label for="f-indicator" class="mb-1 block text-[10px] font-semibold uppercase tracking-wider text-slate-600">Indikator</label>
										<Input id="f-indicator" placeholder="Indikator" bind:value={fIndicatorRef} class="h-8 text-sm" />
									</div>
									<div class="rounded-md border border-slate-200 bg-slate-50 px-3 py-2">
										<p class="text-[10px] font-semibold uppercase tracking-wider text-slate-600">Alur</p>
										<p class="mt-0.5 text-[11px] text-slate-500">Draft dan review dikendalikan dari tombol bawah.</p>
									</div>
								</div>
							</section>

							<section id="composer-stimulus" class="scroll-mt-4 space-y-2 rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
								<div class="flex flex-wrap items-center justify-between gap-2">
									<div>
										<h3 class="text-xs font-black uppercase tracking-[0.2em] text-slate-800">Stimulus</h3>
										<p class="mt-0.5 text-xs text-slate-500">Narasi, data, gambar, atau konteks pendukung.</p>
									</div>
									<button type="button" onclick={() => (focusedEditor = 'stimulus')} class="rounded-md border border-slate-200 px-2 py-1 text-[10px] font-semibold uppercase tracking-wider text-slate-600 hover:bg-slate-50">Fokus</button>
								</div>
								<LegacyRichTextEditor
									bind:value={fStimulus}
									id="soal-stimulus"
									placeholder="Opsional. Tambahkan wacana, data, tabel, gambar, atau konteks sebelum pertanyaan utama."
									minRows={3}
									compact
									onImageUpload={uploadImageInEditor}
								/>
							</section>
						{/if}

						<section id="composer-question" class="scroll-mt-4 space-y-2 rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
							<div class="flex flex-wrap items-center justify-between gap-2">
								<div>
									<h3 class="text-xs font-black uppercase tracking-[0.2em] text-slate-800">{isEssay ? 'Pertanyaan Essay' : isShortAnswer ? 'Pertanyaan Isian Singkat' : isMatching ? 'Instruksi Menjodohkan' : isTrueFalse ? 'Pernyataan Benar/Salah' : isAgreeDisagree ? 'Pernyataan Setuju/Tidak Setuju' : 'Isi Pertanyaan'}</h3>
									<p class="mt-0.5 text-xs text-slate-500">{questionTypeConfig.studentHint}</p>
								</div>
								<button type="button" onclick={() => (focusedEditor = 'stem')} class="rounded-md border border-slate-200 px-2 py-1 text-[10px] font-semibold uppercase tracking-wider text-slate-600 hover:bg-slate-50">Fokus</button>
							</div>
							<LegacyRichTextEditor
								bind:value={fStem}
								id="soal-stem"
								placeholder={isEssay ? 'Tuliskan instruksi essay/uraian. Contoh: Jelaskan alasan, uraikan langkah, atau analisis data berikut.' : isShortAnswer ? 'Tuliskan pertanyaan yang jawabannya singkat dan jelas.' : isMatching ? 'Tuliskan instruksi. Contoh: Jodohkan istilah pada kolom kiri dengan pengertian yang tepat pada kolom kanan.' : isTrueFalse ? 'Tuliskan satu pernyataan yang dapat dinilai benar atau salah.' : isAgreeDisagree ? 'Tuliskan pernyataan sikap yang dapat dijawab setuju atau tidak setuju.' : 'Tuliskan pertanyaan utama. Gambar bisa disisipkan langsung di antara teks.'}
								minRows={4}
								compact
								onImageUpload={uploadImageInEditor}
							/>
							{#if !readinessChecks.stem}
								<p class="text-[10px] font-semibold text-red-500">Isi pertanyaan minimal 5 karakter atau sisipkan gambar.</p>
							{/if}
						</section>

						{#if isMatching}
							<section id="composer-matching" class="scroll-mt-4 space-y-4 border-t border-slate-200 pt-5">
								<div class="flex flex-wrap items-center justify-between gap-3">
									<div>
										<h3 class="text-xs font-black uppercase italic tracking-[0.26em] text-slate-500">Pasangan Menjodohkan</h3>
										<p class="mt-1 text-xs text-slate-500">Kolom kiri adalah pernyataan/istilah. Kolom kanan adalah pasangan benar yang akan dipilih siswa.</p>
									</div>
									<div class="flex flex-wrap items-center gap-2">
										<span class="rounded-full bg-slate-100 px-2 py-1 text-[10px] font-semibold text-slate-500">{fMatchingPairs.length} pasangan</span>
										<Button type="button" variant="outline" size="sm" class="h-7 px-2 text-[10px]" disabled={fMatchingPairs.length <= questionTypeConfig.minOptions} onclick={removeLastMatchingPair}>
											Kurangi
										</Button>
										<Button type="button" variant="outline" size="sm" class="h-7 px-2 text-[10px]" disabled={fMatchingPairs.length >= questionTypeConfig.maxOptions} onclick={addMatchingPair}>
											+ Pasangan
										</Button>
									</div>
								</div>
								<div class="space-y-3">
									{#each fMatchingPairs as pair, i (`matching-pair-${i}`)}
										{@const leftLabel = optionLabelAt(i)}
										{@const rightLabel = i + 1}
										<section class="rounded-xl border border-slate-200 bg-white p-3 shadow-sm">
											<div class="mb-2 flex items-center justify-between gap-3 border-b border-slate-100 pb-2">
												<div>
													<p class="text-xs font-black uppercase tracking-[0.2em] text-slate-700">Pasangan {leftLabel} = {rightLabel}</p>
													<p class="mt-0.5 text-[11px] text-slate-400">Kunci disimpan otomatis sebagai {leftLabel}={rightLabel}</p>
												</div>
											</div>
											<div class="grid grid-cols-1 gap-3 lg:grid-cols-2">
												<div>
													<label for={`matching-left-${i}`} class="mb-1 block text-[10px] font-semibold uppercase tracking-wider text-slate-600">Kolom Kiri {leftLabel}</label>
													<LegacyRichTextEditor
														bind:value={fMatchingPairs[i].left}
														id={`matching-left-${i}`}
														placeholder={`Istilah/pernyataan ${leftLabel}`}
														minRows={2}
														compact
														onImageUpload={uploadImageInEditor}
													/>
													{#if !richTextHasContent(pair.left)}
														<p class="mt-1 text-[10px] font-semibold text-red-500">Kolom kiri wajib diisi.</p>
													{/if}
												</div>
												<div>
													<label for={`matching-right-${i}`} class="mb-1 block text-[10px] font-semibold uppercase tracking-wider text-slate-600">Kolom Kanan {rightLabel}</label>
													<LegacyRichTextEditor
														bind:value={fMatchingPairs[i].right}
														id={`matching-right-${i}`}
														placeholder={`Pasangan jawaban ${rightLabel}`}
														minRows={2}
														compact
														onImageUpload={uploadImageInEditor}
													/>
													{#if !richTextHasContent(pair.right)}
														<p class="mt-1 text-[10px] font-semibold text-red-500">Kolom kanan wajib diisi.</p>
													{/if}
												</div>
											</div>
										</section>
									{/each}
								</div>
								<div class="rounded-xl border border-dashed border-slate-300 bg-slate-50 p-3">
									<div class="flex flex-wrap items-center justify-between gap-3">
										<div>
											<p class="text-xs font-black uppercase tracking-[0.2em] text-slate-700">Distraktor Kanan Opsional</p>
											<p class="mt-0.5 text-[11px] text-slate-500">Tambahkan pilihan kanan ekstra agar siswa tidak hanya mencocokkan satu-ke-satu.</p>
										</div>
										<Button type="button" variant="outline" size="sm" class="h-7 px-2 text-[10px]" disabled={fMatchingDistractors.length >= MAX_MATCHING_DISTRACTOR_COUNT} onclick={addMatchingDistractor}>
											+ Distraktor
										</Button>
									</div>
									{#if fMatchingDistractors.length > 0}
										<div class="mt-3 grid grid-cols-1 gap-3 lg:grid-cols-2">
											{#each fMatchingDistractors as distractor, i (`matching-distractor-${i}`)}
												<div class="rounded-lg border border-slate-200 bg-white p-3">
													<div class="mb-2 flex items-center justify-between gap-3">
														<label for={`matching-distractor-${i}`} class="block text-[10px] font-semibold uppercase tracking-wider text-slate-600">Distraktor Kanan {fMatchingPairs.length + i + 1}</label>
														<button type="button" onclick={() => removeMatchingDistractor(i)} class="rounded-md border border-slate-200 px-2 py-1 text-[10px] font-semibold uppercase tracking-wider text-slate-500 hover:bg-slate-50">
															Hapus
														</button>
													</div>
													<LegacyRichTextEditor
														bind:value={fMatchingDistractors[i]}
														id={`matching-distractor-${i}`}
														placeholder={`Pilihan kanan ekstra ${fMatchingPairs.length + i + 1}`}
														minRows={2}
														compact
														onImageUpload={uploadImageInEditor}
													/>
													{#if !richTextHasContent(distractor)}
														<p class="mt-1 text-[10px] font-semibold text-amber-600">Isi distraktor atau hapus jika tidak dipakai.</p>
													{/if}
												</div>
											{/each}
										</div>
									{/if}
								</div>
							</section>
						{:else if hasOptionSection}
							<section id="composer-options" class="scroll-mt-4 space-y-4 border-t border-slate-200 pt-5">
								<div class="text-center">
									<h3 class="text-xs font-black uppercase italic tracking-[0.26em] text-slate-500">
										{isMultipleAnswer ? 'Opsi & Kunci Jawaban Ganda' : isTrueFalse ? 'Kunci Benar/Salah' : isAgreeDisagree ? 'Kunci Setuju/Tidak Setuju' : 'Opsi & Kunci Jawaban'}
									</h3>
									<p class="mt-1 text-xs text-slate-500">{questionTypeConfig.studentHint}</p>
								</div>
								{#if hasEditableOptions}
									<div class="flex flex-wrap items-center justify-center gap-2">
										<span class="rounded-full bg-slate-100 px-2 py-1 text-[10px] font-semibold text-slate-500">{fOptions.length} opsi aktif</span>
										<Button type="button" variant="outline" size="sm" class="h-7 px-2 text-[10px]" disabled={fOptions.length <= questionTypeConfig.minOptions} onclick={removeLastOption}>
											Kurangi
										</Button>
										<Button type="button" variant="outline" size="sm" class="h-7 px-2 text-[10px]" disabled={fOptions.length >= questionTypeConfig.maxOptions} onclick={addOption}>
											+ Opsi
										</Button>
									</div>
								{/if}
								<div class="grid grid-cols-1 items-start gap-3 xl:grid-cols-2">
									{#each fOptions as opt, i (`composer-option-${i}`)}
										{@const label = optionLabelAt(i)}
										{@const isAnswer = isAnswerLabelSelected(label)}
										<section
											class="space-y-2 rounded-xl border bg-white p-3 shadow-sm transition-colors {isAnswer
												? 'border-green-500 ring-4 ring-green-100'
												: 'border-slate-200 hover:border-slate-300'}"
										>
											<div class="flex items-center justify-between gap-3 border-b border-slate-100 pb-2">
												<div>
													<p class="text-xs font-black uppercase tracking-[0.2em] text-slate-700">Opsi {label}</p>
													<p class="mt-0.5 text-[11px] text-slate-400">{isAnswer ? 'Ditandai sebagai kunci jawaban' : isFixedPair ? 'Pilihan tetap' : 'Pengecoh / alternatif jawaban'}</p>
												</div>
												<div class="flex shrink-0 items-center gap-1">
													{#if hasEditableOptions}
														<button type="button" onclick={() => (focusedEditor = label)} class="rounded-md border border-slate-200 px-2 py-1 text-[10px] font-semibold uppercase tracking-wider text-slate-600 hover:bg-slate-50">Fokus</button>
													{/if}
													<label class="flex cursor-pointer items-center gap-2 rounded-lg px-2 py-1.5 transition-colors {isAnswer ? 'bg-green-600 text-white' : 'hover:bg-green-50'}">
														{#if isMultipleAnswer}
															<input
																type="checkbox"
																checked={isAnswer}
																onchange={() => toggleAnswerLabel(label)}
																class="h-4 w-4 cursor-pointer accent-green-700"
															/>
														{:else}
															<input
																type="radio"
																name="answer-key"
																value={label}
																checked={isAnswer}
																onchange={() => toggleAnswerLabel(label)}
																class="h-4 w-4 cursor-pointer accent-green-700"
															/>
														{/if}
														<span class="text-[10px] font-black uppercase tracking-widest {isAnswer ? 'text-white' : 'text-slate-500'}">Kunci</span>
													</label>
												</div>
											</div>
											<div dir={fIsRtl ? 'rtl' : undefined}>
												{#if hasEditableOptions}
													<LegacyRichTextEditor
														bind:value={fOptions[i]}
														id={`soal-option-${label}`}
														placeholder={`Teks jawaban ${label}. Gambar bisa disisipkan langsung di dalam opsi.`}
														minRows={2}
														compact
														onImageUpload={uploadImageInEditor}
													/>
													{#if !richTextHasContent(fOptions[i])}
														<p class="mt-1 text-[10px] font-semibold text-red-500">Opsi {label} wajib diisi.</p>
													{/if}
												{:else}
													<div class="rounded-md border border-slate-100 bg-slate-50 px-3 py-2 text-sm font-semibold text-slate-700">
														{opt}
													</div>
												{/if}
											</div>
										</section>
									{/each}
								</div>
								{#if isMultipleAnswer && !answerKeyReady}
									<p class="text-center text-[10px] font-semibold text-red-500">Pilih minimal dua opsi sebagai kunci jawaban ganda.</p>
								{/if}
							</section>
						{:else if isShortAnswer}
							<section id="composer-answer" class="scroll-mt-4 space-y-3 rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
								<div>
									<h3 class="text-xs font-black uppercase tracking-[0.2em] text-slate-800">Kunci Isian Singkat</h3>
									<p class="mt-0.5 text-xs text-slate-500">Pisahkan beberapa jawaban diterima dengan tanda |. Sistem mengabaikan besar/kecil huruf dan spasi ganda.</p>
								</div>
								<div>
									<label for="f-short-answer-key" class="mb-1 block text-[10px] font-semibold uppercase tracking-wider text-slate-600">
										Kunci / Alias Jawaban <span class="text-red-500">*</span>
									</label>
									<Input
										id="f-short-answer-key"
										bind:value={fAnswerKey}
										placeholder="Contoh: Fotosintesis | foto sintesis"
										class="h-9 text-sm font-medium"
									/>
									{#if shortAnswerAliases.length > 0}
										<p class="mt-1 text-[10px] font-semibold text-green-700">{shortAnswerAliases.length} jawaban diterima: {shortAnswerAliases.join(' / ')}</p>
									{/if}
									{#if !readinessChecks.answerKey}
										<p class="mt-1 text-[10px] font-semibold text-red-500">Kunci isian singkat wajib diisi sebelum review.</p>
									{/if}
								</div>
							</section>
						{:else}
						<section id="composer-rubric" class="scroll-mt-4 space-y-2 rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
							<div class="flex flex-wrap items-center justify-between gap-2">
								<div>
									<h3 class="text-xs font-black uppercase tracking-[0.2em] text-slate-800">{isAdvanceMode ? 'Rubrik Penilaian' : 'Pedoman Jawaban'}</h3>
									<p class="mt-0.5 text-xs text-slate-500">{isAdvanceMode ? 'Kriteria koreksi, rentang skor, dan catatan penilai.' : 'Panduan singkat agar guru mudah mengoreksi jawaban.'}</p>
								</div>
								<button type="button" onclick={() => (focusedEditor = 'rubric')} class="rounded-md border border-slate-200 px-2 py-1 text-[10px] font-semibold uppercase tracking-wider text-slate-600 hover:bg-slate-50">Fokus</button>
							</div>
							<LegacyRichTextEditor
								bind:value={fRubric}
								id="soal-rubric"
								placeholder={isAdvanceMode ? 'Tuliskan rubrik lengkap. Contoh: ketepatan konsep 40%, langkah/alasan 40%, bahasa/sistematika 20%.' : 'Tuliskan pedoman jawaban atau poin penting yang harus ada pada jawaban siswa.'}
								minRows={isAdvanceMode ? 4 : 3}
								compact
								onImageUpload={uploadImageInEditor}
							/>
							{#if !readinessChecks.rubric}
								<p class="text-[10px] font-semibold text-red-500">Pedoman/rubrik essay wajib diisi.</p>
							{/if}
						</section>
						{/if}

						{#if isAdvanceMode}
							<section id="composer-explanation" class="scroll-mt-4 space-y-2 rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
								<div class="flex flex-wrap items-center justify-between gap-2">
									<div>
										<h3 class="text-xs font-black uppercase tracking-[0.2em] text-slate-800">{isEssay ? 'Catatan Pembahasan' : 'Pembahasan'}</h3>
										<p class="mt-0.5 text-xs text-slate-500">{isEssay ? 'Catatan internal untuk guru/reviewer.' : 'Pembahasan yang membantu review dan bank soal.'}</p>
									</div>
									<button type="button" onclick={() => (focusedEditor = 'explanation')} class="rounded-md border border-slate-200 px-2 py-1 text-[10px] font-semibold uppercase tracking-wider text-slate-600 hover:bg-slate-50">Fokus</button>
								</div>
								<LegacyRichTextEditor
									bind:value={fExplanation}
									id="soal-explanation"
									placeholder={isEssay ? 'Opsional. Catatan koreksi, contoh jawaban ideal, atau alasan rubrik.' : 'Opsional. Tulis pembahasan atau langkah penyelesaian.'}
									minRows={3}
									compact
									onImageUpload={uploadImageInEditor}
								/>
							</section>
						{/if}
					</div>

					{#if showInspector || composerMobilePanel === 'preview'}
					<aside class={`space-y-4 min-w-0 xl:sticky xl:top-0 xl:self-start ${composerMobilePanel === 'write' ? 'hidden lg:block' : 'block'}`}>
						<div class="rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
							{@render composerPreview()}
						</div>
					</aside>
					{/if}
				</div>
			</div>

			<div class="flex shrink-0 flex-col gap-2 border-t border-green-100 bg-white px-4 py-2.5 md:flex-row md:items-center md:justify-between md:px-6">
				<div class="min-w-0 text-xs text-slate-500">
					<span class="font-semibold text-green-700">{draftStatusLabel()}</span>
					<span class="ml-2 text-slate-400">· {draftIssues.length === 0 ? 'Draft bisa disimpan' : draftIssues[0]} · Review {validationIssues.length === 0 ? 'siap' : `${validationIssues.length} wajib belum lengkap`} · Ctrl+S</span>
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
						onclick={() => void saveQuestion('draft')}
						disabled={!canSaveDraft}
						loading={composerAction === 'draft'}
						loadingLabel="Menyimpan draft..."
						variant="outline"
						class="h-8 text-xs disabled:opacity-50"
					>
						Simpan Draft
					</LoadingButton>
					<LoadingButton
						onclick={() => void saveQuestion('review')}
						disabled={!canSubmitReview}
						loading={composerAction === 'review'}
						loadingLabel="Mengajukan..."
						class="h-8 bg-green-700 text-xs text-white hover:bg-green-800 disabled:opacity-50"
					>
						Ajukan Review
					</LoadingButton>
				</div>
			</div>

			{#if focusedEditor}
				<div class="absolute inset-0 z-20 flex bg-slate-950/35 p-3 md:p-6">
					<div class="flex min-h-0 w-full flex-col overflow-hidden rounded-xl bg-white shadow-2xl">
						<div class="flex shrink-0 items-center justify-between gap-3 border-b border-green-100 px-4 py-3">
							<div>
								<p class="text-[10px] font-black uppercase tracking-[0.22em] text-green-700">Mode Fokus</p>
								<h3 class="text-base font-black uppercase italic text-slate-900">{focusTitle(focusedEditor)}</h3>
							</div>
							<Button type="button" variant="outline" size="sm" class="h-8 text-xs" onclick={() => (focusedEditor = null)}>
								Tutup Fokus
							</Button>
						</div>
						<div class="min-h-0 flex-1 overflow-y-auto bg-slate-50 p-4">
							{#if focusedEditor === 'stem'}
								<LegacyRichTextEditor
									bind:value={fStem}
									id="soal-stem-focus"
									placeholder="Tuliskan pertanyaan utama. Gambar bisa disisipkan langsung di antara teks."
									minRows={10}
									onImageUpload={uploadImageInEditor}
								/>
							{:else if focusedEditor === 'stimulus'}
								<LegacyRichTextEditor
									bind:value={fStimulus}
									id="soal-stimulus-focus"
									placeholder="Tambahkan stimulus, wacana, data, tabel, atau konteks pendukung."
									minRows={8}
									onImageUpload={uploadImageInEditor}
								/>
							{:else if focusedEditor === 'rubric'}
								<LegacyRichTextEditor
									bind:value={fRubric}
									id="soal-rubric-focus"
									placeholder="Tuliskan rubrik atau pedoman jawaban essay."
									minRows={8}
									onImageUpload={uploadImageInEditor}
								/>
							{:else if focusedEditor === 'explanation'}
								<LegacyRichTextEditor
									bind:value={fExplanation}
									id="soal-explanation-focus"
									placeholder="Tulis pembahasan atau catatan internal."
									minRows={8}
									onImageUpload={uploadImageInEditor}
								/>
							{:else if focusedEditor === 'A'}
								<LegacyRichTextEditor
									bind:value={fOptions[0]}
									id="soal-option-focus-a"
									placeholder="Teks jawaban A"
									minRows={8}
									onImageUpload={uploadImageInEditor}
								/>
							{:else if focusedEditor === 'B'}
								<LegacyRichTextEditor
									bind:value={fOptions[1]}
									id="soal-option-focus-b"
									placeholder="Teks jawaban B"
									minRows={8}
									onImageUpload={uploadImageInEditor}
								/>
							{:else if focusedEditor === 'C'}
								<LegacyRichTextEditor
									bind:value={fOptions[2]}
									id="soal-option-focus-c"
									placeholder="Teks jawaban C"
									minRows={8}
									onImageUpload={uploadImageInEditor}
								/>
							{:else if focusedEditor === 'D'}
								<LegacyRichTextEditor
									bind:value={fOptions[3]}
									id="soal-option-focus-d"
									placeholder="Teks jawaban D"
									minRows={8}
									onImageUpload={uploadImageInEditor}
								/>
							{:else if focusedEditor === 'E'}
								<LegacyRichTextEditor
									bind:value={fOptions[4]}
									id="soal-option-focus-e"
									placeholder="Teks jawaban E"
									minRows={8}
									onImageUpload={uploadImageInEditor}
								/>
							{:else if focusedEditor === 'F'}
								<LegacyRichTextEditor
									bind:value={fOptions[5]}
									id="soal-option-focus-f"
									placeholder="Teks jawaban F"
									minRows={8}
									onImageUpload={uploadImageInEditor}
								/>
							{/if}
						</div>
					</div>
				</div>
			{/if}
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
