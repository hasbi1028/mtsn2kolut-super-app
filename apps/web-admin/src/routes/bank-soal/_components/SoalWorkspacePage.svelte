<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
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
	import ComposerDraftNotice from './ComposerDraftNotice.svelte';
	import CatalogTargetPanel from './CatalogTargetPanel.svelte';
	import ImportWorkflowPanel from './ImportWorkflowPanel.svelte';
	import ReviewWorkflowQueues from './ReviewWorkflowQueues.svelte';
	import SoalContextPanel from './SoalContextPanel.svelte';
	import SoalShellHeader from './SoalShellHeader.svelte';
	import SoalStatusCards from './SoalStatusCards.svelte';
	import {
		bankSoalQueueAuthRequiredMessage,
		buildBankSoalQuestionSyncInput,
		shouldQueueBankSoalQuestionSave,
	} from '$lib/client/bank-soal-composer-offline';
	import {
		clearBankSoalDraftPayloads,
		deleteBankSoalDraftPayload,
		enqueueBankSoalQuestionSync,
		isAuthExpiredSyncStatus,
		listBankSoalQuestionSyncQueue,
		loadBankSoalDraftPayload,
		markBankSoalQuestionSyncFailed,
		migrateLegacyBankSoalDrafts,
		removeBankSoalQuestionSyncItem,
		saveBankSoalDraftPayload,
		type BankSoalQuestionSyncItem,
	} from '$lib/client/bank-soal-offline';
	import { questionExportButtonLabel, questionExportSuccessMessage } from '$lib/cbt/question-export-ui';
	import { canDeleteBankSoal, canPublishBankSoal, canReviewBankSoal } from '$lib/bank-soal/access';
	import { confirmAction } from '$lib/confirm-dialog';
	import { clientApiPath, clientApiPathWithQuery, readClientApiData, readClientJson } from '$lib/client/api';
	import { htmlToPlainText } from '$lib/utils/html-text';

	// ── Types ─────────────────────────────────────────────────────────────────
	type Subject = { id: string; name: string; code: string };
	type CbtEvent = {
		id: string;
		title: string;
		exam_type?: string;
		status?: string;
		academic_year_name?: string;
	};
	type UserOption = {
		id: string;
		username: string;
		roles?: string[];
		employee_id?: string | null;
		profile_nama?: string | null;
	};
	type EventMemberRole = 'panitia' | 'pembuat_soal' | 'reviewer' | 'proktor' | 'pengawas' | 'korektor';
	type EventMember = {
		id: string;
		user_id: string;
		employee_id?: string | null;
		subject_id?: string | null;
		role: EventMemberRole;
		username?: string;
		employee_nama?: string;
		employee_name?: string;
		subject_name?: string;
		subject_code?: string;
	};
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
	type WorkspaceRouteMode = 'composer' | 'import';
	type WorkspaceRoute = '/bank-soal' | '/bank-soal/tambah' | '/bank-soal/impor' | '/bank-soal/verifikasi';
	type AuthoringMode = 'beginner' | 'advance';
	type ComposerIssueHint = { message: string; targetId: string };
type ComposerStageCard = { label: string; desc: string; status: string; tone: 'green' | 'amber' | 'red' | 'slate'; targetId: string };
	type RevisionSourceFilter = '' | 'item_analysis' | 'reviewer' | 'workflow';
	type ReviewDecision = 'approve' | 'reject';
	type BulkWorkflowAction = 'approve' | 'reject' | 'publish';
	type ComposerQuestionType = 'multiple_choice' | 'multiple_answer' | 'true_false' | 'agree_disagree' | 'matching' | 'ordering' | 'short_answer' | 'essay';
	type ComposerSaveIntent = 'draft' | 'review';
	type AnswerMode = 'single_option' | 'multi_option' | 'fixed_pair' | 'matching' | 'ordering' | 'short_text' | 'rubric';
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
		event_id?: string | null;
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
		events: CbtEvent[];
		revisionQueue: Question[];
		revisionTotal: number;
		reviewQueue: Question[];
		reviewTotal: number;
		approvedQueue: Question[];
		approvedTotal: number;
		questionTargets: QuestionTarget[];
		totalItems: number;
		page: number;
	};
	type PageData = {
		user?: {
			role?: string;
			roles?: string[];
			permissions?: string[];
		};
	};
	type AcademicPayload = {
		subjects?: Subject[];
		error?: string;
		message?: string;
	};
	type EventsPayload = CbtEvent[] | { items?: CbtEvent[]; events?: CbtEvent[] };
	type UsersPayload = UserOption[] | { items?: UserOption[]; users?: UserOption[] };
	type DraftPayload = {
		eventId?: string;
		specialEventMode?: boolean;
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
	type QuestionPayloadOption = {
		label: string;
		text: string;
		html: string;
		match_label?: string;
		match_text?: string;
		match_html?: string;
		is_distractor?: boolean;
	};
	type QuestionSavePayload = {
		event_id?: string;
		authoring_mode: AuthoringMode;
		subject_id: string;
		question_text: string;
		question_type: ComposerQuestionType;
		stem_html: string;
		stimulus_html: string;
		explanation_html: string;
		rubric_html: string;
		options: QuestionPayloadOption[];
		answer_key: string;
		difficulty: string;
		status: 'draft';
		workflow_status: 'draft' | 'review';
		grade_level: number;
		academic_phase: string;
		cp_ref: string;
		tp_ref: string;
		kd_ref: string;
		indicator_ref: string;
		material_topic: string;
		cognitive_level: string;
		hots_flag: boolean;
		writer_notes: string;
		review_notes: string;
	};
	type OptionLabel = 'A' | 'B' | 'C' | 'D' | 'E' | 'F';
	type LegacyImportResult = {
		total_rows: number;
		valid?: number;
		would_import?: number;
		imported: number;
		skipped: number;
		errors: string[];
		duplicate_codes: string[];
	};
	type QuestionTarget = {
		event_id?: string;
		subject_id: string;
		subject_name?: string;
		subject_code?: string;
		target_questions: number;
		total: number;
		draft: number;
		review: number;
		rejected: number;
		approved: number;
		published: number;
	};
	type TimelineItem = {
		id?: string;
		action?: string;
		status?: string;
		notes?: string;
		actor_username?: string;
		created_at?: string;
	};
	type BulkWorkflowResult = {
		question_id?: string;
		id?: string;
		ok?: boolean;
		success?: boolean;
		status?: string;
		error?: string;
		message?: string;
	};
	type BulkWorkflowResponse = { results?: BulkWorkflowResult[] } | BulkWorkflowResult[];
	type FocusedEditor = 'stem' | 'stimulus' | 'rubric' | 'explanation' | OptionLabel;

	let { data, routeMode }: { data: PageData; routeMode: WorkspaceRouteMode } = $props();

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
			id: 'ordering',
			label: 'Mengurutkan',
			shortLabel: 'Urutan',
			desc: 'Siswa menyusun opsi sesuai urutan kunci.',
			studentHint: 'Siswa mengurutkan semua opsi jawaban.',
			answerMode: 'ordering',
			minOptions: MIN_OPTION_COUNT,
			maxOptions: MAX_OPTION_COUNT,
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
	const WORKFLOW_LABEL: Record<string, string> = {
		draft: 'Draft',
		review: 'Menunggu Review',
		approved: 'Disetujui',
		rejected: 'Perlu Revisi',
	};
	const DIFFICULTY_LABEL: Record<string, string> = { easy: 'Mudah', medium: 'Sedang', hard: 'Sulit' };
	const revisionSourceOptions: Array<{ id: RevisionSourceFilter; label: string; desc: string }> = [
		{ id: '', label: 'Semua Revisi', desc: 'Semua sumber' },
		{ id: 'item_analysis', label: 'Analisis Butir', desc: 'Dari hasil ujian' },
		{ id: 'reviewer', label: 'Reviewer', desc: 'Catatan penelaah' },
		{ id: 'workflow', label: 'Workflow', desc: 'Tanpa reviewer' },
	];

	const DRAFT_KEY = (id: string | null, eventId: string) => `mtsn2-soal-komposer:${eventId || 'global'}:${id ?? 'new'}`;
	const EVENT_MEMBER_ROLES: Array<{ value: EventMemberRole; label: string; desc: string }> = [
		{ value: 'panitia', label: 'Panitia', desc: 'Koordinasi kegiatan' },
		{ value: 'pembuat_soal', label: 'Pembuat Soal', desc: 'Menyusun bank soal' },
		{ value: 'reviewer', label: 'Reviewer', desc: 'Menelaah mutu soal' },
		{ value: 'proktor', label: 'Proktor', desc: 'Teknis sesi ujian' },
		{ value: 'pengawas', label: 'Pengawas', desc: 'Pengawasan ruang' },
		{ value: 'korektor', label: 'Korektor', desc: 'Koreksi uraian' },
	];

	function currentRouteMode(): ModuleMode {
		return routeMode;
	}

	// ── Page state ─────────────────────────────────────────────────────────────
	const activeMode: ModuleMode = currentRouteMode();
	let questionsPromise = $state<Promise<SoalOverview> | null>(null);
	let questions = $state<Question[]>([]);
	let subjects = $state<Subject[]>([]);
	let events = $state<CbtEvent[]>([]);
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
	let selectedEventId = $state('');
	let specialEventQuestionMode = $state(false);
	let filterWorkflow = $state('');
	let filterStatus = $state('');
	let revisionSourceFilter = $state<RevisionSourceFilter>('');
	let visibleSelectionCheckbox = $state<HTMLInputElement | null>(null);
	let questionTargets = $state<QuestionTarget[]>([]);
	let targetBusy = $state(false);
	let targetQuestionsInput = $state(0);
	let eventMembers = $state<EventMember[]>([]);
	let eventMembersLoading = $state(false);
	let userOptions = $state<UserOption[]>([]);
	let userOptionsLoading = $state(false);
	let memberBusyId = $state('');
	let addMemberBusy = $state(false);
	let memberUserId = $state('');
	let memberSubjectId = $state('');
	let memberRole = $state<EventMemberRole>('pembuat_soal');

	// ── Composer state ─────────────────────────────────────────────────────────
	let editingId = $state<string | null>(null);
	let editingEventId = $state('');
	let composerBusy = $state(false);
	let composerAction = $state<ComposerSaveIntent | ''>('');
	let draftStatus = $state('');
	let draftSavedAt = $state<string | null>(null);
	let showInspector = $state(false);
	let composerMobilePanel = $state<'write' | 'preview'>('write');
	let focusedEditor = $state<FocusedEditor | null>(null);
	let lastDraftSig = '';
	let questionsRequestId = 0;
	let importSubjectId = $state('');
	let importFile = $state<File | null>(null);
	let importBusy = $state(false);
	let importDryRunDone = $state(false);
	let importResult = $state<LegacyImportResult | null>(null);
	let exportBusy = $state(false);
	let templateBusy = $state(false);
	let duplicateBusyId = $state('');
	let workflowBusyId = $state('');
	let reviewDecisionOpen = $state(false);
	let reviewDecisionQuestion = $state<Question | null>(null);
	let reviewDecision = $state<ReviewDecision>('approve');
	let reviewDecisionNotes = $state('');
	let questionPreviewOpen = $state(false);
	let questionPreview = $state<Question | null>(null);
	let questionPreviewLoading = $state(false);
	let questionTimeline = $state<TimelineItem[]>([]);
	let questionTimelineLoading = $state(false);
	let selectedQuestionIds = $state<string[]>([]);
	let bulkBusy = $state(false);
	let bulkNotes = $state('');
	let openMenuId = $state('');
	let isOnline = $state(true);
	let offlineQueueItems = $state<BankSoalQuestionSyncItem<QuestionSavePayload>[]>([]);
	let offlineSyncBusy = $state(false);
	let offlineStatus = $state('');
	let offlineAuthRequired = $state(false);
	let offlineLastSyncAt = $state<string | null>(null);

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
	let activeDraftKey = $derived(DRAFT_KEY(editingId, selectedEventId));
	let draftSignature = $derived(JSON.stringify({
		specialEventQuestionMode,
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
	let canReviewWorkflow = $derived(canReviewBankSoal(data.user));
	let canPublishWorkflow = $derived(canPublishBankSoal(data.user));
	let canDeleteQuestion = $derived(canDeleteBankSoal(data.user));
	let selectedEvent = $derived(events.find((event) => event.id === selectedEventId) ?? null);
	let selectedEventTitle = $derived(selectedEvent?.title ?? 'Bank soal reusable');
	let specialEventAttachId = $derived(specialEventQuestionMode && selectedEventId ? selectedEventId : '');
	let roleLabel = $derived.by(() => {
		if (roles.includes('admin')) return 'Admin bank soal';
		if (roles.includes('guru') || roles.includes('teacher')) return 'Pembuat soal / guru';
		return roles.length > 0 ? roles.join(', ') : 'Pengguna';
	});
	let selectedImportContext = $derived.by(() => {
		if (!selectedEvent) return 'Bank Soal reusable tanpa event';
		const eventLabel = `${selectedEvent.title}${selectedEvent.status ? ` (${selectedEvent.status})` : ''}`;
		if (specialEventQuestionMode) return `Soal khusus kegiatan untuk ${eventLabel}`;
		return `Bank Soal reusable; filter kegiatan aktif: ${eventLabel}. CSV tidak membawa event_id`;
	});
	let exportButtonLabel = $derived(questionExportButtonLabel(roles));
	let exportSuccessMessage = $derived(questionExportSuccessMessage(roles));
	let hasCatalogQuickFilter = $derived(Boolean(search.trim() || filterWorkflow || filterStatus));
	let reviewCount = $derived(reviewTotal);
	let visibleReviewCount = $derived(questions.filter((item) => item.workflow_status === 'review').length);
	let approvedCount = $derived(approvedTotal);
	let visibleApprovedCount = $derived(questions.filter((item) => item.workflow_status === 'approved' && item.status === 'draft').length);
	let draftCount = $derived(questions.filter((item) => item.workflow_status === 'draft').length);
	let visibleRevisionCount = $derived(questions.filter((item) => item.workflow_status === 'rejected').length);
	let publishedCount = $derived(questions.filter((item) => item.status === 'published').length);
	let selectedTarget = $derived(questionTargets.find((target) => target.subject_id === filterSubject) ?? null);
	let eventTargetTotal = $derived(questionTargets.reduce((sum, target) => sum + (target.target_questions || 0), 0));
	let eventPublishedTotal = $derived(questionTargets.reduce((sum, target) => sum + (target.published || 0), 0));
	let selectedTargetShortage = $derived(Math.max(0, (selectedTarget?.target_questions ?? 0) - (selectedTarget?.published ?? 0)));
	let visibleQuestionIds = $derived(questions.map((question) => question.id));
	let selectedVisibleCount = $derived(visibleQuestionIds.filter((id) => selectedQuestionIds.includes(id)).length);
	let allVisibleQuestionsSelected = $derived(visibleQuestionIds.length > 0 && selectedVisibleCount === visibleQuestionIds.length);
	let selectedQuestions = $derived(questions.filter((question) => selectedQuestionIds.includes(question.id)));
	let selectedReviewEligibleCount = $derived(selectedQuestions.filter(canDecideReview).length);
	let selectedPublishEligibleCount = $derived(selectedQuestions.filter(canPublishQuestion).length);
	let offlineQueueCount = $derived(offlineQueueItems.length);
	let offlineReviewQueueCount = $derived(offlineQueueItems.filter((item) => item.intent === 'review').length);
	let offlineDraftQueueCount = $derived(offlineQueueItems.filter((item) => item.intent === 'draft').length);
	let canSyncOfflineQueue = $derived(isOnline && offlineQueueCount > 0 && !offlineSyncBusy);
	let statusCards = $derived([
		{ label: 'Draft', value: draftCount, tone: 'slate', helper: 'soal masih disusun', workflowStatus: 'draft', status: '', active: filterWorkflow === 'draft' && !filterStatus },
		{ label: 'Perlu Revisi', value: revisionTotal, tone: 'red', helper: `${visibleRevisionCount} tampil`, workflowStatus: 'rejected', status: '', active: filterWorkflow === 'rejected' && !filterStatus },
		{ label: 'Menunggu Review', value: reviewCount, tone: 'amber', helper: `${visibleReviewCount} tampil`, workflowStatus: 'review', status: '', active: filterWorkflow === 'review' && !filterStatus },
		{ label: 'Disetujui', value: approvedCount, tone: 'green', helper: `${visibleApprovedCount} siap terbit`, workflowStatus: 'approved', status: 'draft', active: filterWorkflow === 'approved' && filterStatus === 'draft' },
		{ label: 'Terbit', value: publishedCount, tone: 'emerald', helper: 'siap dipakai paket', workflowStatus: '', status: 'published', active: !filterWorkflow && filterStatus === 'published' },
	]);
	let selectedSubject = $derived(subjects.find((subject) => subject.id === fSubjectId) ?? null);
	let composerModeLabel = $derived(fAuthoringMode === 'advance' ? 'Mode advance' : 'Mode pemula');
	let composerScopeLabel = $derived(specialEventQuestionMode ? `Khusus ${selectedEventTitle}` : 'Bank soal reusable');
	let composerStageCards = $derived.by<ComposerStageCard[]>(() => [
		{
			label: 'Metadata',
			desc: selectedSubject ? `${selectedSubject.name} · ${DIFFICULTY_LABEL[fDifficulty] ?? fDifficulty}` : 'Pilih mapel, jenis, kesulitan',
			status: readinessChecks.subject ? 'Lengkap' : 'Wajib',
			tone: readinessChecks.subject ? 'green' : 'red',
			targetId: 'composer-metadata',
		},
		{
			label: isAdvanceMode ? 'Blueprint' : 'Mode Singkat',
			desc: isAdvanceMode ? `${fCognitiveLevel.trim() || 'Kognitif belum diisi'} · ${fMaterialTopic.trim() || 'Topik belum diisi'}` : 'CP/TP/KD disembunyikan agar cepat',
			status: isAdvanceMode ? (fCognitiveLevel.trim() ? 'Terarah' : 'Opsional') : 'Pemula',
			tone: !isAdvanceMode || fCognitiveLevel.trim() ? 'green' : 'amber',
			targetId: isAdvanceMode ? 'composer-advanced' : 'composer-question',
		},
		{
			label: 'Naskah',
			desc: stemText ? `${stemText.length} karakter pertanyaan` : 'Tulis pertanyaan utama',
			status: readinessChecks.stem ? 'Terisi' : 'Wajib',
			tone: readinessChecks.stem ? 'green' : 'red',
			targetId: 'composer-question',
		},
		{
			label: isEssay ? 'Rubrik' : isShortAnswer ? 'Kunci' : isMatching ? 'Pasangan' : 'Opsi',
			desc: isEssay ? 'Pedoman koreksi uraian' : isShortAnswer ? `${shortAnswerAliases.length} alias jawaban` : isMatching ? `${fMatchingPairs.length} pasangan` : `${fOptions.length} opsi aktif`,
			status: readinessChecks.options && readinessChecks.answerKey && readinessChecks.rubric ? 'Valid' : 'Cek lagi',
			tone: readinessChecks.options && readinessChecks.answerKey && readinessChecks.rubric ? 'green' : 'amber',
			targetId: isMatching ? 'composer-matching' : isShortAnswer ? 'composer-answer' : isEssay ? 'composer-rubric' : 'composer-options',
		},
	]);

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
	let isOrdering = $derived(questionTypeConfig.answerMode === 'ordering');
	let isShortAnswer = $derived(questionTypeConfig.answerMode === 'short_text');
	let hasOptionSection = $derived(
		questionTypeConfig.answerMode === 'single_option' ||
			questionTypeConfig.answerMode === 'multi_option' ||
			questionTypeConfig.answerMode === 'ordering' ||
			isFixedPair
	);
	let hasEditableOptions = $derived(
		questionTypeConfig.answerMode === 'single_option' ||
			questionTypeConfig.answerMode === 'multi_option' ||
			questionTypeConfig.answerMode === 'ordering'
	);
	let requiresRubric = $derived(questionTypeConfig.answerMode === 'rubric');
	let isAdvanceMode = $derived(fAuthoringMode === 'advance');
	let activeOptionLabels = $derived(ANSWER_LABELS.slice(0, fOptions.length));
	let selectedAnswerLabels = $derived(isOrdering ? orderingAnswerKeyLabels(fAnswerKey, activeOptionLabels) : answerKeyLabels(fAnswerKey, activeOptionLabels));
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
		if (isOrdering) return selectedAnswerLabels.length === fOptions.length;
		return selectedAnswerLabels.length === 1;
	});
	let rubricReady = $derived(!requiresRubric || rubricText.length >= 5 || fRubric.includes('<img'));

	let readinessChecks = $derived({
		subject: !!fSubjectId,
		stem: stemText.length >= 5 || hasImage,
		options: isEssay || (isMatching ? matchingPairsReady : optionsReady),
		answerKey: answerKeyReady,
		rubric: rubricReady,
	});

	let passedChecks = $derived(Object.values(readinessChecks).filter(Boolean).length);
	let totalChecks = $derived(Object.keys(readinessChecks).length);
	let readinessScore = $derived(Math.round((passedChecks / totalChecks) * 100));
	let draftIssues = $derived.by(() => {
		const issues: string[] = [];
		if (!readinessChecks.subject) issues.push('Pilih mata pelajaran sebelum menyimpan draft');
		if (!readinessChecks.stem) issues.push('Isi pertanyaan minimal 5 karakter untuk draft');
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
		return issues;
	});
	let firstComposerIssue = $derived.by((): ComposerIssueHint | null => {
		if (!readinessChecks.subject) return { message: 'pilih mata pelajaran', targetId: 'composer-metadata' };
		if (!readinessChecks.stem) return { message: 'isi pertanyaan', targetId: 'composer-question' };
		if (!readinessChecks.options) {
			if (isMatching) return { message: 'lengkapi pasangan', targetId: 'composer-matching' };
			return { message: 'lengkapi opsi jawaban', targetId: 'composer-options' };
		}
		if (!readinessChecks.answerKey) {
			if (isMatching) return { message: 'cek pasangan jawaban', targetId: 'composer-matching' };
			if (isShortAnswer) return { message: 'isi kunci isian', targetId: 'composer-answer' };
			return { message: 'pilih kunci jawaban', targetId: 'composer-options' };
		}
		if (requiresRubric && !readinessChecks.rubric) return { message: 'isi rubrik essay', targetId: 'composer-rubric' };
		return null;
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

	$effect(() => {
		if (!visibleSelectionCheckbox) return;
		visibleSelectionCheckbox.indeterminate = selectedVisibleCount > 0 && !allVisibleQuestionsSelected;
	});

	// ── Draft autosave ─────────────────────────────────────────────────────────
	function buildDraftPayload(): DraftPayload {
		return {
			eventId: selectedEventId,
			specialEventMode: specialEventQuestionMode,
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
		void saveBankSoalDraftPayload(draftKey, buildDraftPayload())
			.then(() => markDraftAutosaved())
			.catch(() => {
				/* ignore storage errors */
			});
	}

	$effect(() => {
		if (activeMode !== 'composer') return;
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

	async function restoreDraft(): Promise<boolean> {
		try {
			const d = await loadBankSoalDraftPayload<{
				eventId?: string;
				specialEventMode?: boolean;
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
			}>(activeDraftKey);
			if (!d) return false;
			if (!selectedEventId && d.eventId) selectedEventId = d.eventId;
			specialEventQuestionMode = !editingId && Boolean(d.specialEventMode && (selectedEventId || d.eventId));
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

	async function clearDraft() {
		await deleteBankSoalDraftPayload(activeDraftKey);
		lastDraftSig = '';
		draftStatus = '';
		draftSavedAt = null;
	}

	async function clearAllLocalDrafts() {
		const count = await clearBankSoalDraftPayloads();
		lastDraftSig = '';
		draftStatus = '';
		draftSavedAt = null;
		toast.success(count > 0 ? `${count} draft lokal Bank Soal dihapus dari perangkat ini` : 'Tidak ada draft lokal Bank Soal di perangkat ini');
	}

	// ── API ────────────────────────────────────────────────────────────────────
	function applyQuestionScopeParams(params: URLSearchParams) {
		params.set('scope', selectedEventId ? 'event_pool' : 'global');
		if (selectedEventId) params.set('event_id', selectedEventId);
	}

	function buildQuestionParams(page: number) {
		const params = new URLSearchParams();
		params.set('limit', String(PAGE_SIZE));
		params.set('offset', String((page - 1) * PAGE_SIZE));
		applyQuestionScopeParams(params);
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
		applyQuestionScopeParams(params);
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
		applyQuestionScopeParams(params);
		params.set('workflow_status', 'review');
		if (search.trim()) params.set('q', search.trim());
		if (filterSubject) params.set('subject_id', filterSubject);
		return params;
	}

	function buildApprovedQueueParams() {
		const params = new URLSearchParams();
		params.set('limit', '6');
		params.set('offset', '0');
		applyQuestionScopeParams(params);
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
		const targetsPromise = selectedEventId
			? fetch(clientApiPath`/api/asesmen/events/${selectedEventId}/question-targets`).then((response) =>
				readClientApiData<QuestionTarget[]>(response, 'Gagal memuat target soal kegiatan')
			)
			: Promise.resolve([] as QuestionTarget[]);
		const [questionPayload, revisionPayload, reviewPayload, approvedPayload, academicPayload, eventsPayload, targetsPayload] = await Promise.all([
			fetch(clientApiPathWithQuery('/api/bank-soal/questions', params)).then((response) =>
				readClientApiData<QuestionListResponse>(response, 'Gagal memuat soal')
			),
			fetch(clientApiPathWithQuery('/api/bank-soal/questions', revisionParams)).then((response) =>
				readClientApiData<QuestionListResponse>(response, 'Gagal memuat antrian revisi')
			),
			fetch(clientApiPathWithQuery('/api/bank-soal/questions', reviewParams)).then((response) =>
				readClientApiData<QuestionListResponse>(response, 'Gagal memuat antrian review')
			),
			fetch(clientApiPathWithQuery('/api/bank-soal/questions', approvedParams)).then((response) =>
				readClientApiData<QuestionListResponse>(response, 'Gagal memuat antrian siap terbit')
			),
			fetch('/api/bank-soal/soal-support/subjects').then((response) =>
				readClientApiData<AcademicPayload>(response, 'Gagal memuat data akademik')
			),
			fetch('/api/asesmen/events').then((response) =>
				readClientApiData<EventsPayload>(response, 'Gagal memuat kegiatan CBT')
			),
			targetsPromise,
		]);
		const loadedQuestions = questionPayload.items ?? [];
		const loadedRevisions = revisionPayload.items ?? [];
		const loadedReviews = reviewPayload.items ?? [];
		const loadedApproved = approvedPayload.items ?? [];
		return {
			questions: loadedQuestions,
			subjects: academicPayload.subjects ?? [],
			events: normalizeEventsPayload(eventsPayload),
			revisionQueue: loadedRevisions,
			revisionTotal: revisionPayload.meta?.total ?? loadedRevisions.length,
			reviewQueue: loadedReviews,
			reviewTotal: reviewPayload.meta?.total ?? loadedReviews.length,
			approvedQueue: loadedApproved,
			approvedTotal: approvedPayload.meta?.total ?? loadedApproved.length,
			totalItems: questionPayload.meta?.total ?? loadedQuestions.length,
			page,
			questionTargets: Array.isArray(targetsPayload) ? targetsPayload : [],
		};
	}

	function applyOverview(overview: SoalOverview) {
		questions = overview.questions;
		const overviewQuestionIds = new Set(overview.questions.map((question) => question.id));
		selectedQuestionIds = selectedQuestionIds.filter((id) => overviewQuestionIds.has(id));
		subjects = overview.subjects;
		events = overview.events;
		revisionQueue = overview.revisionQueue;
		revisionTotal = overview.revisionTotal;
		reviewQueue = overview.reviewQueue;
		reviewTotal = overview.reviewTotal;
		approvedQueue = overview.approvedQueue;
		approvedTotal = overview.approvedTotal;
		questionTargets = overview.questionTargets;
		if (filterSubject) targetQuestionsInput = questionTargets.find((target) => target.subject_id === filterSubject)?.target_questions ?? 0;
		totalItems = overview.totalItems;
		currentPage = overview.page;
	}

	function load(page = currentPage) {
		const requestId = ++questionsRequestId;
		selectedQuestionIds = [];
		questionsPromise = fetchOverview(page).then((overview) => {
			if (requestId !== questionsRequestId) return { questions, subjects, events, revisionQueue, revisionTotal, reviewQueue, reviewTotal, approvedQueue, approvedTotal, questionTargets, totalItems, page: currentPage };
			applyOverview(overview);
			return overview;
		}).catch((error: unknown) => {
			if (requestId === questionsRequestId) throw error;
			return { questions, subjects, events, revisionQueue, revisionTotal, reviewQueue, reviewTotal, approvedQueue, approvedTotal, questionTargets, totalItems, page: currentPage };
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
				questionsPromise = Promise.resolve({ questions, subjects, events, revisionQueue, revisionTotal, reviewQueue, reviewTotal, approvedQueue, approvedTotal, questionTargets, totalItems, page: currentPage });
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

	function normalizeEventsPayload(payload: EventsPayload): CbtEvent[] {
		if (Array.isArray(payload)) return payload;
		return payload.items ?? payload.events ?? [];
	}

	function normalizeUsersPayload(payload: UsersPayload): UserOption[] {
		if (Array.isArray(payload)) return payload;
		return payload.items ?? payload.users ?? [];
	}

	function userDisplayName(user: UserOption): string {
		const profile = user.profile_nama?.trim();
		return profile ? `${profile} (${user.username})` : user.username;
	}

	function memberDisplayName(member: EventMember): string {
		return member.employee_nama ?? member.employee_name ?? member.username ?? member.user_id;
	}

	function memberRoleLabel(role: string): string {
		return EVENT_MEMBER_ROLES.find((item) => item.value === role)?.label ?? role;
	}

	function handleQuestionsRenderError(error: unknown, reset: () => void) {
		console.error('Question composer render failed', error);
		reset();
	}

	function mutationErrorMessage(error: unknown, fallback: string) {
		if (error instanceof Error && error.message.trim() && !error.message.toLowerCase().includes('fetch')) return error.message;
		return fallback;
	}

	function browserOnline(): boolean {
		if (typeof navigator === 'undefined') return true;
		return navigator.onLine;
	}

	function updateOnlineStatus() {
		isOnline = browserOnline();
	}

	function requireOnlineAction(label: string): boolean {
		updateOnlineStatus();
		if (isOnline) return true;
		toast.warning(`${label} harus online dan tidak masuk antrian offline.`);
		return false;
	}

	async function refreshOfflineQueueState() {
		const items = await listBankSoalQuestionSyncQueue<QuestionSavePayload>();
		offlineQueueItems = items;
		offlineAuthRequired = items.some((item) => item.authRequired);
	}

	async function queueQuestionSave(payload: QuestionSavePayload, intent: ComposerSaveIntent, authRequired = false) {
		const draftKey = activeDraftKey;
		await saveBankSoalDraftPayload(draftKey, buildDraftPayload());
		const input = buildBankSoalQuestionSyncInput({
			draftKey,
			editingId,
			intent,
			payload,
		});
		await enqueueBankSoalQuestionSync<QuestionSavePayload>({
			...input,
			lastError: authRequired ? 'Sesi perlu masuk ulang sebelum sinkronisasi.' : 'Menunggu koneksi untuk sinkronisasi.',
			authRequired,
		});
		await refreshOfflineQueueState();
		offlineStatus = authRequired ? bankSoalQueueAuthRequiredMessage(offlineQueueCount) : 'Perubahan disimpan di antrian lokal.';
		offlineAuthRequired = authRequired || offlineAuthRequired;
		draftStatus = authRequired ? 'Antrian menunggu login ulang' : 'Tersimpan di antrian offline';
		draftSavedAt = new Date().toISOString();
		toast.warning(authRequired ? offlineStatus : 'Perubahan Bank Soal disimpan lokal dan akan disinkronkan saat online.');
	}

	async function syncBankSoalOfflineQueue(manual = false) {
		if (offlineSyncBusy) return;
		updateOnlineStatus();
		if (!isOnline) {
			if (manual) toast.warning('Perangkat masih offline. Antrian Bank Soal tetap tersimpan lokal.');
			offlineStatus = 'Menunggu koneksi untuk sinkronisasi.';
			return;
		}
		const items = await listBankSoalQuestionSyncQueue<QuestionSavePayload>();
		if (items.length === 0) {
			offlineQueueItems = [];
			offlineAuthRequired = false;
			offlineStatus = manual ? 'Tidak ada antrian Bank Soal.' : offlineStatus;
			return;
		}
		offlineSyncBusy = true;
		offlineStatus = 'Menyinkronkan antrian Bank Soal...';
		let synced = 0;
		let failed = 0;
		let authBlocked = false;
		try {
			for (const item of items) {
				let responseReceived = false;
				try {
					const response = await fetch(item.endpoint, {
						method: item.method,
						headers: { 'Content-Type': 'application/json' },
						body: JSON.stringify(item.payload),
					});
					responseReceived = true;
					if (isAuthExpiredSyncStatus(response.status)) {
						await markBankSoalQuestionSyncFailed(item.id, 'Sesi berakhir. Masuk ulang lalu sinkronkan lagi.', true);
						authBlocked = true;
						break;
					}
					await readClientJson<unknown>(response);
					await removeBankSoalQuestionSyncItem(item.id);
					await deleteBankSoalDraftPayload(item.draftKey);
					synced += 1;
				} catch (error) {
					failed += 1;
					const message = responseReceived
						? mutationErrorMessage(error, 'Sinkronisasi Bank Soal gagal')
						: 'Koneksi belum stabil. Sinkronisasi ditunda.';
					await markBankSoalQuestionSyncFailed(item.id, message, false);
					if (!browserOnline()) break;
				}
			}
			await refreshOfflineQueueState();
			if (synced > 0) {
				offlineLastSyncAt = new Date().toISOString();
				toast.success(`${synced} perubahan Bank Soal berhasil disinkronkan`);
				await refreshOverview(1);
			}
			if (authBlocked) {
				offlineStatus = bankSoalQueueAuthRequiredMessage(offlineQueueCount);
				toast.warning(offlineStatus);
			} else if (failed > 0) {
				offlineStatus = `${failed} perubahan belum tersinkron. Coba lagi saat koneksi stabil.`;
				if (manual) toast.warning(offlineStatus);
			} else if (synced > 0) {
				offlineStatus = 'Antrian Bank Soal selesai disinkronkan.';
			}
		} finally {
			offlineSyncBusy = false;
		}
	}

	function moduleRoute(mode: ModuleMode): WorkspaceRoute {
		if (mode === 'catalog') return '/bank-soal';
		if (mode === 'review') return '/bank-soal/verifikasi';
		if (mode === 'import') return '/bank-soal/impor';
		return '/bank-soal/tambah';
	}

	function navigateToModuleMode(mode: ModuleMode) {
		if (typeof window === 'undefined') return;
		const target = new URL(resolve(moduleRoute(mode)), window.location.origin);
		if (selectedEventId && (mode === 'composer' || mode === 'import' || mode === 'review')) {
			target.searchParams.set('event_id', selectedEventId);
		}
		window.location.href = `${target.pathname}${target.search}`;
	}

	function setModuleMode(mode: ModuleMode) {
		if (mode === activeMode) return;
		navigateToModuleMode(mode);
	}

	function resetImportDryRunPreview() {
		importResult = null;
		importDryRunDone = false;
	}

	function setImportSubject(subjectId: string) {
		if (importSubjectId === subjectId) return;
		importSubjectId = subjectId;
		resetImportDryRunPreview();
	}

	function setSpecialEventQuestionMode(enabled: boolean) {
		if (specialEventQuestionMode === enabled) return;
		specialEventQuestionMode = enabled;
		resetImportDryRunPreview();
	}

	function handleSpecialEventModeChange(event: Event) {
		setSpecialEventQuestionMode((event.currentTarget as HTMLInputElement).checked);
	}

	function setSelectedEvent(eventId: string) {
		const eventChanged = selectedEventId !== eventId;
		selectedEventId = eventId;
		if (eventChanged) resetImportDryRunPreview();
		if (!eventId) specialEventQuestionMode = false;
		selectedQuestionIds = [];
		currentPage = 1;
		if (typeof window !== 'undefined') {
			const url = new URL(window.location.href);
			if (eventId) url.searchParams.set('event_id', eventId);
			else url.searchParams.delete('event_id');
			window.history.replaceState({}, '', `${url.pathname}?${url.searchParams.toString()}`);
		}
		load(1);
		void refreshEventMembers();
	}

	function setFilterSubject(subjectId: string) {
		filterSubject = subjectId;
		selectedQuestionIds = [];
		targetQuestionsInput = questionTargets.find((target) => target.subject_id === subjectId)?.target_questions ?? 0;
		load(1);
	}

	function reviewFocusHref(): '/bank-soal/verifikasi' | `/bank-soal/verifikasi?${string}` {
		const params = new URLSearchParams();
		if (selectedEventId) params.set('event_id', selectedEventId);
		if (filterSubject) params.set('subject_id', filterSubject);
		return params.toString() ? `/bank-soal/verifikasi?${params.toString()}` : '/bank-soal/verifikasi';
	}

	async function saveQuestionTarget() {
		if (!selectedEventId || !filterSubject) {
			toast.warning('Pilih konteks kegiatan dan mapel sebelum mengatur kebutuhan soal paket/event.');
			return;
		}
		targetBusy = true;
		try {
			await fetch(clientApiPath`/api/asesmen/events/${selectedEventId}/question-targets`, {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ subject_id: filterSubject, target_questions: Number(targetQuestionsInput) || 0 })
			}).then((response) => readClientJson<unknown>(response));
			toast.success('Target soal mapel diperbarui');
			await refreshOverview(currentPage);
		} catch (error) {
			toast.error(mutationErrorMessage(error, 'Gagal menyimpan target soal'));
		} finally {
			targetBusy = false;
		}
	}

	function toggleQuestionSelection(id: string) {
		selectedQuestionIds = selectedQuestionIds.includes(id)
			? selectedQuestionIds.filter((item) => item !== id)
			: [...selectedQuestionIds, id];
	}

	function setVisibleQuestionSelection(selected: boolean) {
		if (selected) {
			selectedQuestionIds = Array.from(new Set([...selectedQuestionIds, ...visibleQuestionIds]));
			return;
		}
		const visibleIds = new Set(visibleQuestionIds);
		selectedQuestionIds = selectedQuestionIds.filter((id) => !visibleIds.has(id));
	}

	function onVisibleQuestionSelectionChange(event: Event) {
		setVisibleQuestionSelection((event.currentTarget as HTMLInputElement).checked);
	}

	function clearSelection() {
		selectedQuestionIds = [];
	}

	async function runBulkWorkflow(action: BulkWorkflowAction) {
		if (!requireOnlineAction('Aksi review/publikasi')) return;
		const ids = action === 'publish'
			? selectedQuestions.filter(canPublishQuestion).map((question) => question.id)
			: selectedQuestions.filter(canDecideReview).map((question) => question.id);
		if (ids.length === 0) {
			toast.warning('Tidak ada soal terpilih yang memenuhi syarat aksi ini.');
			return;
		}
		const notes = bulkNotes.trim();
		if (action === 'reject' && notes.length < 8) {
			toast.warning('Catatan bulk reject minimal 8 karakter.');
			return;
		}
		bulkBusy = true;
		try {
			const res = await fetch('/api/bank-soal/questions/bulk-workflow', {
				method: 'PATCH',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ action, question_ids: ids, notes })
			});
			const payload = await readClientApiData<BulkWorkflowResponse>(res, 'Aksi bulk gagal');
			const results = Array.isArray(payload) ? payload : payload.results ?? [];
			const ok = results.filter((item) => item.ok ?? item.success ?? !item.error).length;
			const failed = Math.max(0, results.length - ok);
			toast.success(`${ok} soal berhasil diproses${failed ? `, ${failed} gagal` : ''}`);
			selectedQuestionIds = [];
			bulkNotes = '';
			await refreshOverview(currentPage);
		} catch (error) {
			toast.error(mutationErrorMessage(error, 'Aksi bulk gagal'));
		} finally {
			bulkBusy = false;
		}
	}

	async function ensureUserOptions() {
		if (!canReviewWorkflow || userOptions.length > 0 || userOptionsLoading) return;
		userOptionsLoading = true;
		try {
			const payload = await fetch('/api/users').then((response) => readClientApiData<UsersPayload>(response, 'Gagal memuat pengguna'));
			userOptions = normalizeUsersPayload(payload);
		} catch (error) {
			toast.error(mutationErrorMessage(error, 'Gagal memuat daftar pengguna'));
		} finally {
			userOptionsLoading = false;
		}
	}

	async function refreshEventMembers() {
		if (!selectedEventId || !canReviewWorkflow) {
			eventMembers = [];
			return;
		}
		eventMembersLoading = true;
		try {
			const members = await fetch(clientApiPath`/api/asesmen/events/${selectedEventId}/members`).then((response) =>
				readClientApiData<EventMember[]>(response, 'Gagal memuat panitia kegiatan')
			);
			eventMembers = Array.isArray(members) ? members : [];
			await ensureUserOptions();
		} catch (error) {
			toast.error(mutationErrorMessage(error, 'Gagal memuat panitia kegiatan'));
		} finally {
			eventMembersLoading = false;
		}
	}

	async function addEventMember() {
		if (!selectedEventId || !memberUserId) {
			toast.warning('Pilih kegiatan dan pengguna terlebih dahulu.');
			return;
		}
		addMemberBusy = true;
		try {
			const selectedUser = userOptions.find((user) => user.id === memberUserId);
			const payload = {
				user_id: memberUserId,
				employee_id: selectedUser?.employee_id || undefined,
				subject_id: memberSubjectId || undefined,
				role: memberRole,
			};
			await fetch(clientApiPath`/api/asesmen/events/${selectedEventId}/members`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(payload),
			}).then((response) => readClientJson<unknown>(response));
			memberUserId = '';
			memberSubjectId = '';
			memberRole = 'pembuat_soal';
			toast.success('Penugasan kegiatan ditambahkan');
			await refreshEventMembers();
		} catch (error) {
			toast.error(mutationErrorMessage(error, 'Gagal menambahkan penugasan'));
		} finally {
			addMemberBusy = false;
		}
	}

	async function removeEventMember(member: EventMember) {
		if (!selectedEventId) return;
		if (!(await confirmAction({
			title: 'Hapus Penugasan',
			message: `Hapus ${memberDisplayName(member)} dari panitia ${selectedEventTitle}?`,
			confirmLabel: 'Hapus Penugasan',
			tone: 'danger'
		}))) return;
		memberBusyId = member.id;
		try {
			await fetch(clientApiPath`/api/asesmen/events/${selectedEventId}/members/${member.id}`, { method: 'DELETE' }).then((response) => readClientJson<unknown>(response));
			toast.success('Penugasan kegiatan dihapus');
			await refreshEventMembers();
		} catch (error) {
			toast.error(mutationErrorMessage(error, 'Gagal menghapus penugasan'));
		} finally {
			memberBusyId = '';
		}
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

	function setCatalogStatusFilter(card: { workflowStatus?: string; status?: string }) {
		if (searchTimer) {
			clearTimeout(searchTimer);
			searchTimer = null;
		}
		search = '';
		filterWorkflow = card.workflowStatus ?? '';
		filterStatus = card.status ?? '';
		revisionSourceFilter = '';
		load(1);
	}

	function clearCatalogQuickFilters() {
		if (searchTimer) {
			clearTimeout(searchTimer);
			searchTimer = null;
		}
		search = '';
		filterWorkflow = '';
		filterStatus = '';
		revisionSourceFilter = '';
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
		if (config.answerMode === 'ordering') return normalizeOrderingAnswerKey('', config.minOptions);
		return '';
	}

	function defaultOptionsForQuestionType(type: ComposerQuestionType): string[] {
		const config = getQuestionTypeConfig(type);
		if (config.answerMode === 'fixed_pair') return [...(config.fixedOptions ?? [])];
		if (config.answerMode !== 'single_option' && config.answerMode !== 'multi_option' && config.answerMode !== 'ordering') return [];
		return Array.from({ length: config.minOptions }, () => '');
	}

	function normalizeOptionCount(options: string[], type: ComposerQuestionType = fQuestionType): string[] {
		const config = getQuestionTypeConfig(type);
		if (config.answerMode === 'fixed_pair') return [...(config.fixedOptions ?? [])];
		if (config.answerMode !== 'single_option' && config.answerMode !== 'multi_option' && config.answerMode !== 'ordering') return [];
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

	function orderingAnswerKeyLabels(value: string | undefined, labels: OptionLabel[]): OptionLabel[] {
		const ordered: OptionLabel[] = [];
		for (const raw of (value ?? '').split(',')) {
			const label = raw.trim().toUpperCase() as OptionLabel;
			if (!labels.includes(label) || ordered.includes(label)) continue;
			ordered.push(label);
		}
		for (const label of labels) {
			if (!ordered.includes(label)) ordered.push(label);
		}
		return ordered;
	}

	function normalizeOrderingAnswerKey(value: string | undefined, optionCount: number): string {
		return orderingAnswerKeyLabels(value, ANSWER_LABELS.slice(0, optionCount)).join(',');
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
		if (config.answerMode === 'ordering') return normalizeOrderingAnswerKey(value, optionCount);
		const labels = ANSWER_LABELS.slice(0, optionCount);
		const keys = answerKeyLabels(value, labels);
		if (config.answerMode === 'multi_option') return keys.join(',');
		return keys[0] ?? defaultAnswerKeyForQuestionType(type);
	}

	function isAnswerLabelSelected(label: OptionLabel): boolean {
		return selectedAnswerLabels.includes(label);
	}

	function toggleAnswerLabel(label: OptionLabel) {
		if (isOrdering) {
			fAnswerKey = normalizeOrderingAnswerKey(fAnswerKey || activeOptionLabels.join(','), fOptions.length);
			return;
		}
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

	function moveOrderingAnswerLabel(label: OptionLabel, delta: -1 | 1) {
		if (!isOrdering) return;
		const ordered = orderingAnswerKeyLabels(fAnswerKey, activeOptionLabels);
		const index = ordered.indexOf(label);
		const nextIndex = index + delta;
		if (index < 0 || nextIndex < 0 || nextIndex >= ordered.length) return;
		[ordered[index], ordered[nextIndex]] = [ordered[nextIndex], ordered[index]];
		fAnswerKey = ordered.join(',');
	}

	function editorAllowedForQuestionType(editor: FocusedEditor, type: ComposerQuestionType): boolean {
		if (editor === 'stem' || editor === 'stimulus' || editor === 'explanation') return true;
		if (editor === 'rubric') return getQuestionTypeConfig(type).answerMode === 'rubric';
		return getQuestionTypeConfig(type).answerMode === 'single_option' || getQuestionTypeConfig(type).answerMode === 'multi_option' || getQuestionTypeConfig(type).answerMode === 'ordering';
	}

	function addOption() {
		const config = getQuestionTypeConfig(fQuestionType);
		if (config.answerMode !== 'single_option' && config.answerMode !== 'multi_option' && config.answerMode !== 'ordering') return;
		if (fOptions.length >= config.maxOptions) return;
		fOptions = [...fOptions, ''];
		if (config.answerMode === 'ordering') fAnswerKey = normalizeOrderingAnswerKey(fAnswerKey, fOptions.length);
	}

	function removeLastOption() {
		const config = getQuestionTypeConfig(fQuestionType);
		if (config.answerMode !== 'single_option' && config.answerMode !== 'multi_option' && config.answerMode !== 'ordering') return;
		if (fOptions.length <= config.minOptions) return;
		const removedLabel = optionLabelAt(fOptions.length - 1);
		fOptions = fOptions.slice(0, -1);
		if (config.answerMode === 'ordering') {
			fAnswerKey = normalizeOrderingAnswerKey(fAnswerKey, fOptions.length);
		} else if (answerKeyLabels(fAnswerKey, ANSWER_LABELS).includes(removedLabel)) {
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
		return canPublishWorkflow && q.workflow_status === 'approved' && q.status === 'draft' && !questionUsageLocked(q);
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

	async function loadQuestionDetail(q: Question): Promise<Question> {
		const res = await fetch(clientApiPath`/api/bank-soal/questions/${q.id}`);
		return readClientApiData<Question>(res, 'Gagal memuat detail soal');
	}

	async function openQuestionPreview(q: Question) {
		questionPreview = q;
		questionPreviewOpen = true;
		questionPreviewLoading = true;
		questionTimeline = [];
		try {
			questionPreview = await loadQuestionDetail(q);
			void loadQuestionTimeline(questionPreview.id);
		} catch (error) {
			toast.warning(mutationErrorMessage(error, explainQuickEditBlocked(q)));
			void loadQuestionTimeline(q.id);
		} finally {
			questionPreviewLoading = false;
		}
	}

	async function loadQuestionTimeline(id: string) {
		questionTimelineLoading = true;
		try {
			const payload = await fetch(clientApiPath`/api/bank-soal/questions/${id}/timeline`).then((response) =>
				readClientApiData<TimelineItem[] | { items?: TimelineItem[] }>(response, 'Gagal memuat timeline soal')
			);
			questionTimeline = Array.isArray(payload) ? payload : payload.items ?? [];
		} catch {
			questionTimeline = [];
		} finally {
			questionTimelineLoading = false;
		}
	}

	function closeQuestionPreview() {
		questionPreviewOpen = false;
		questionPreview = null;
		questionPreviewLoading = false;
		questionTimeline = [];
	}

	function duplicatePreviewQuestion() {
		if (!questionPreview) return;
		void duplicateQuestion(questionPreview.id);
	}

	function openQuestion(q: Question) {
		if (isQuickEditable(q)) {
			void openEdit(q);
			return;
		}
		void openQuestionPreview(q);
	}

	async function openQuestionFromRouteParam(id: string) {
		try {
			const q = await loadQuestionDetail({ id } as Question);
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
		editingEventId = '';
		specialEventQuestionMode = false;
		resetForm();
		showInspector = false;
		focusedEditor = null;
		composerMobilePanel = 'write';
		setModuleMode('composer');
		// Delay to let state settle before restoring
		setTimeout(() => {
			void restoreDraft().then((restored) => {
				if (!restored) draftStatus = '';
			});
		}, 50);
	}

	async function openEdit(q: Question) {
		if (!isQuickEditable(q)) {
			setModuleMode('catalog');
			toast.warning(explainQuickEditBlocked(q));
			return;
		}
		composerBusy = true;
		specialEventQuestionMode = false;
		try {
			const d = await loadQuestionDetail(q);
			if (!selectedEventId && d.event_id) selectedEventId = d.event_id;

			editingId = d.id;
			editingEventId = d.event_id ?? '';
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
			setModuleMode('composer');
			setTimeout(() => {
				void restoreDraft().then((restored) => {
					if (restored) toast.info('Draft edit lokal dipulihkan otomatis.');
				});
			}, 50);
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
			setModuleMode('composer');
		} finally {
			composerBusy = false;
		}
	}

	function closeComposer() {
		setModuleMode('catalog');
		editingId = null;
		editingEventId = '';
		focusedEditor = null;
		showInspector = false;
	}

	async function requestCloseComposer() {
		if (hasDraftWork && draftStatus) {
			const confirmed = await confirmAction({
				title: 'Tutup Komposer?',
				message: 'Draft lokal tetap disimpan otomatis. Kembali ke daftar soal dan lanjutkan nanti?',
				confirmLabel: 'Tutup',
				tone: 'warning'
			});
			if (!confirmed) return;
		}
		closeComposer();
	}

	async function discardLocalDraftAndClose() {
		const confirmed = await confirmAction({
			title: 'Hapus Draft Lokal?',
			message: 'Draft autosave pada perangkat ini akan dihapus dan komposer kembali ke daftar soal. Soal yang sudah tersimpan di server tidak ikut dihapus.',
			confirmLabel: 'Hapus Draft Lokal',
			tone: 'danger'
		});
		if (!confirmed) return;
		await clearDraft();
		closeComposer();
	}

	function scrollComposerSection(id: string) {
		document.getElementById(id)?.scrollIntoView({ behavior: 'smooth', block: 'start' });
	}

	function scrollToFirstComposerIssue() {
		if (!firstComposerIssue) return;
		scrollComposerSection(firstComposerIssue.targetId);
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

	function offlineLastSyncLabel(): string {
		if (!offlineLastSyncAt) return '';
		const synced = new Date(offlineLastSyncAt);
		if (Number.isNaN(synced.getTime())) return '';
		return synced.toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' });
	}

	function offlineRetryLoginHref(): string {
		const from = typeof window === 'undefined' ? '/bank-soal/tambah' : `${window.location.pathname}${window.location.search}`;
		return `${resolve('/login')}?from=${encodeURIComponent(from)}`;
	}

	function focusTitle(editor: FocusedEditor) {
		if (editor === 'stem') return 'Isi Pertanyaan';
		if (editor === 'stimulus') return 'Stimulus';
		if (editor === 'rubric') return 'Rubrik Penilaian';
		if (editor === 'explanation') return 'Pembahasan';
		return `Opsi ${editor}`;
	}

	function handleComposerKeydown(event: KeyboardEvent) {
		if (activeMode !== 'composer') return;
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
		editingId = null;
		editingEventId = '';
		setImportSubject(filterSubject || fSubjectId || '');
		importFile = null;
		importResult = null;
		importDryRunDone = false;
		setModuleMode('import');
	}

	function importFileValidationMessage(file: File): string {
		if (file.size <= 0) return 'File CSV kosong. Pilih file yang berisi data soal.';
		if (!file.name.trim().toLowerCase().endsWith('.csv')) return 'File import harus berekstensi .csv.';
		return '';
	}

	function onImportFileChange(event: Event) {
		const input = event.target as HTMLInputElement;
		const file = input.files?.[0] ?? null;
		resetImportDryRunPreview();
		if (!file) {
			importFile = null;
			return;
		}
		const validationMessage = importFileValidationMessage(file);
		if (validationMessage) {
			importFile = null;
			input.value = '';
			toast.warning(validationMessage);
			return;
		}
		importFile = file;
	}

	function buildPayloadOptions(): QuestionPayloadOption[] {
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
		if (config.answerMode !== 'single_option' && config.answerMode !== 'multi_option' && config.answerMode !== 'ordering') {
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

	function buildQuestionSavePayload(isReview: boolean): QuestionSavePayload {
		return {
			...(!editingId && specialEventAttachId ? { event_id: specialEventAttachId } : {}),
			...(editingId && editingEventId ? { event_id: editingEventId } : {}),
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
		const payload = buildQuestionSavePayload(isReview);
		let responseReceived = false;
		let responseStatus: number | undefined;
		try {
			const syncInput = buildBankSoalQuestionSyncInput({
				draftKey: activeDraftKey,
				editingId,
				intent,
				payload,
			});
			if (!browserOnline()) {
				await queueQuestionSave(payload, intent, false);
				closeComposer();
				return;
			}
			const res = await fetch(syncInput.endpoint, {
				method: syncInput.method,
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(payload),
			});
			responseReceived = true;
			responseStatus = res.status;
			if (isAuthExpiredSyncStatus(res.status)) {
				await queueQuestionSave(payload, intent, true);
				closeComposer();
				return;
			}
			await readClientJson<unknown>(res);

			await clearDraft();
			toast.success(isReview ? 'Soal diajukan review' : editingId ? 'Draft soal berhasil diperbarui' : 'Draft soal berhasil dibuat');
			closeComposer();
			await refreshOverview(1);
		} catch (e) {
			if (shouldQueueBankSoalQuestionSave({ isOnline: browserOnline(), responseReceived, responseStatus })) {
				await queueQuestionSave(payload, intent, false);
				closeComposer();
				return;
			}
			toast.error(mutationErrorMessage(e, 'Gagal menyimpan soal'));
		} finally {
			composerBusy = false;
			composerAction = '';
		}
	}

	async function importLegacyCSV(dryRun = false) {
		if (!requireOnlineAction('Import CSV Bank Soal')) return;
		if (!importSubjectId) {
			toast.error('Pilih mata pelajaran untuk import');
			return;
		}
		if (!importFile) {
			toast.error('Pilih file CSV terlebih dahulu');
			return;
		}
		if (!dryRun) {
			const importErrors = importResult?.errors.length ?? 0;
			const readyImportCount = importResult?.would_import ?? importResult?.valid ?? importResult?.imported ?? 0;
			if (!importDryRunDone) {
				toast.warning('Jalankan Preview Dry-run sebelum konfirmasi import.');
				return;
			}
			if (importErrors > 0) {
				toast.warning('Perbaiki error dry-run sebelum konfirmasi import.');
				return;
			}
			if (readyImportCount <= 0) {
				toast.warning('Tidak ada baris valid untuk diimport.');
				return;
			}
		}
		importBusy = true;
		try {
			const form = new FormData();
			form.set('subject_id', importSubjectId);
			if (specialEventAttachId) form.set('event_id', specialEventAttachId);
			if (dryRun) form.set('dry_run', 'true');
			form.set('file', importFile);
			const res = await fetch('/api/bank-soal/questions/import-legacy', {
				method: 'POST',
				body: form,
			});
			const result = await readClientApiData<LegacyImportResult>(res, 'Import CSV gagal');
			importResult = {
				total_rows: result.total_rows ?? 0,
				valid: result.valid ?? result.would_import ?? result.imported ?? 0,
				would_import: result.would_import ?? result.valid ?? result.imported ?? 0,
				imported: result.imported ?? 0,
				skipped: result.skipped ?? 0,
				errors: result.errors ?? [],
				duplicate_codes: result.duplicate_codes ?? [],
			};
			if (dryRun) {
				importDryRunDone = true;
				toast.success(`${importResult.would_import ?? importResult.valid ?? 0} soal valid untuk diimport`);
			} else {
				toast.success(`${importResult.imported} soal berhasil diimport`);
				importDryRunDone = false;
				await refreshOverview(1);
			}
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
		if (!requireOnlineAction('Export Bank Soal')) return;
		exportBusy = true;
		try {
			const params = buildQuestionParams(1);
			params.delete('limit');
			params.delete('offset');
			const response = await fetch(clientApiPathWithQuery('/api/bank-soal/questions/export', params));
			await downloadCSVResponse(response, `bank-soal-cbt-${new Date().toISOString().slice(0, 10)}.csv`);
			toast.success(exportSuccessMessage);
		} catch (error) {
			toast.error(mutationErrorMessage(error, 'Export CSV gagal'));
		} finally {
			exportBusy = false;
		}
	}

	async function downloadQuestionsTemplateCSV() {
		if (!requireOnlineAction('Download template CSV')) return;
		templateBusy = true;
		try {
			const response = await fetch('/api/bank-soal/questions/template');
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
			const res = await fetch(clientApiPath`/api/bank-soal/questions/${id}/duplicate`, { method: 'POST' });
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
		if (!requireOnlineAction('Ajukan review ulang')) return;
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
			const res = await fetch(clientApiPath`/api/bank-soal/questions/${q.id}/workflow`, {
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
			toast.warning('Anda belum memiliki izin review soal.');
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
		questionTimeline = [];
		reviewDecisionOpen = true;
		void loadQuestionDetail(q).then((detail) => {
			if (reviewDecisionOpen && reviewDecisionQuestion?.id === q.id) {
				reviewDecisionQuestion = detail;
				void loadQuestionTimeline(detail.id);
			}
		}).catch((error) => {
			toast.warning(mutationErrorMessage(error, 'Detail lengkap soal belum dapat dimuat.'));
		});
		void loadQuestionTimeline(q.id);
	}

	function closeReviewDecision() {
		reviewDecisionOpen = false;
		reviewDecisionQuestion = null;
		reviewDecisionNotes = '';
		reviewDecision = 'approve';
		questionTimeline = [];
	}

	async function submitReviewDecision() {
		if (!requireOnlineAction('Keputusan reviewer')) return;
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
			const res = await fetch(clientApiPath`/api/bank-soal/questions/${q.id}/workflow`, {
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
		if (!requireOnlineAction('Publikasi soal')) return;
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
			const res = await fetch(clientApiPath`/api/bank-soal/questions/${q.id}/workflow`, {
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
			const res = await fetch(clientApiPathWithQuery('/api/bank-soal/questions', new URLSearchParams({ id })), { method: 'DELETE' });
			await readClientJson<unknown>(res);
			toast.success('Soal dihapus');
			await refreshOverview(currentPage);
		} catch (e) {
			toast.error(mutationErrorMessage(e, 'Gagal menghapus soal'));
		}
	}

	function toggleRowMenu(id: string, event: MouseEvent) {
		event.stopPropagation();
		openMenuId = openMenuId === id ? '' : id;
	}

	function closeRowMenu() {
		openMenuId = '';
	}

	async function uploadImageInEditor(file: File): Promise<string> {
		const form = new FormData();
		form.set('file', file);
		form.set('purpose', 'general');
		if (editingId) form.set('question_id', editingId);
		const res = await fetch('/api/bank-soal/assets', { method: 'POST', body: form });
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
			draft: 'bg-muted text-muted-foreground',
			review: 'bg-warning/15 text-warning',
			approved: 'bg-success/15 text-success',
			rejected: 'bg-destructive/15 text-destructive',
		};
		return map[status] ?? 'bg-muted text-muted-foreground';
	}

	function answerKeyLabelForQuestion(q: Question): string {
		const key = (q.answer_key ?? '').trim();
		const type = normalizeQuestionType(q.question_type);
		if (type === 'essay') return q.rubric_html ? 'Rubrik uraian tersedia' : 'Rubrik belum tersedia';
		if (!key) return 'Kunci belum tersedia atau dirahasiakan';
		if (type === 'multiple_answer') return `Kunci: ${answerKeyLabels(key, ANSWER_LABELS).join(', ') || key}`;
		if (type === 'matching') return `Pasangan: ${key.split(';').map((part) => part.trim()).filter(Boolean).join(' · ')}`;
		if (type === 'short_answer') return `Jawaban diterima: ${parseShortAnswerAliases(key).join(' / ') || key}`;
		if (type === 'true_false' || type === 'agree_disagree') {
			const config = getQuestionTypeConfig(type);
			const index = ANSWER_LABELS.indexOf(key.toUpperCase() as OptionLabel);
			return `Kunci: ${config.fixedOptions?.[index] ?? key}`;
		}
		return `Kunci: ${key}`;
	}

	function optionPrimaryContent(option: OptionItem): string {
		if (option.is_distractor) return option.match_html || option.match_text || '-';
		return option.html || option.text || option.latex || '-';
	}

	function optionSecondaryContent(option: OptionItem): string {
		if (option.is_distractor) return '';
		return option.match_html || option.match_text || '';
	}

	function composerStageToneClass(tone: ComposerStageCard['tone']): string {
		const map: Record<ComposerStageCard['tone'], string> = {
			green: 'border-primary/20 bg-primary/10 text-primary',
			amber: 'border-warning/30 bg-warning/10 text-warning',
			red: 'border-destructive/30 bg-destructive/10 text-destructive',
			slate: 'border-border bg-muted/50 text-foreground',
		};
		return map[tone];
	}

	function scrollToComposerSection(targetId: string) {
		document.getElementById(targetId)?.scrollIntoView({ behavior: 'smooth', block: 'start' });
	}

	function applyComposerQueryPrefill(params: URLSearchParams): boolean {
		if (routeMode !== 'composer') return false;
		let applied = false;
		const subjectId = params.get('subject_id') ?? '';
		const gradeLevel = Number(params.get('grade_level') ?? '');
		const questionType = params.get('question_type') ?? '';
		if (subjectId) {
			fSubjectId = subjectId;
			filterSubject = subjectId;
			applied = true;
		}
		if (Number.isFinite(gradeLevel) && gradeLevel > 0) {
			fGradeLevel = gradeLevel;
			applied = true;
		}
		if (questionType) {
			fQuestionType = normalizeQuestionType(questionType);
			fOptions = normalizeOptionCount(fOptions, fQuestionType);
			fAnswerKey = normalizeAnswerKey(fAnswerKey, fQuestionType, answerItemCountForType(fQuestionType));
			applied = true;
		}
		if (selectedEventId) {
			specialEventQuestionMode = true;
			applied = true;
		}
		if (applied) draftStatus = 'Konteks event/mapel diterapkan dari Kelengkapan Soal';
		return applied;
	}

	onMount(() => {
		const params = new URLSearchParams(window.location.search);
		const questionId = params.get('question_id');
		selectedEventId = params.get('event_id') ?? '';
		const hasComposerPrefill = applyComposerQueryPrefill(params);
		if (questionId) {
			params.delete('question_id');
			const query = params.toString();
			window.history.replaceState({}, '', query ? `${window.location.pathname}?${query}` : window.location.pathname);
		}
		load();
		void refreshEventMembers();
		updateOnlineStatus();
		void migrateLegacyBankSoalDrafts().then(() => refreshOfflineQueueState()).then(() => {
			if (browserOnline() && offlineQueueCount > 0) void syncBankSoalOfflineQueue(false);
		});
		if (routeMode === 'composer' && questionId) void openQuestionFromRouteParam(questionId);
		else if (routeMode === 'composer') {
			setTimeout(() => {
				void restoreDraft().then((restored) => {
					if (hasComposerPrefill) applyComposerQueryPrefill(params);
					else if (!restored) draftStatus = '';
				});
			}, 50);
		}
		const handleOnline = () => {
			updateOnlineStatus();
			offlineStatus = 'Koneksi kembali. Antrian dapat disinkronkan.';
			void syncBankSoalOfflineQueue(false);
		};
		const handleOffline = () => {
			updateOnlineStatus();
			offlineStatus = 'Mode offline aktif. Simpan draft masuk antrian lokal.';
		};
		const handleVisibility = () => {
			if (document.visibilityState !== 'visible') return;
			updateOnlineStatus();
			void refreshOfflineQueueState().then(() => {
				if (browserOnline() && offlineQueueCount > 0) void syncBankSoalOfflineQueue(false);
			});
		};
		window.addEventListener('keydown', handleComposerKeydown);
		window.addEventListener('online', handleOnline);
		window.addEventListener('offline', handleOffline);
		document.addEventListener('visibilitychange', handleVisibility);
		return () => {
			window.removeEventListener('keydown', handleComposerKeydown);
			window.removeEventListener('online', handleOnline);
			window.removeEventListener('offline', handleOffline);
			document.removeEventListener('visibilitychange', handleVisibility);
		};
	});
</script>

<!-- ── Main page ──────────────────────────────────────────────────────────── -->
<div class="space-y-4">
	<SoalShellHeader
		reviewHref={reviewFocusHref()}
		{exportButtonLabel}
		{exportBusy}
		{totalItems}
		onCreate={openCreate}
		onExport={() => void exportQuestionsCSV()}
	/>

	<SoalContextPanel
		{events}
		{subjects}
		{selectedEventId}
		{filterSubject}
		{selectedEventTitle}
		{roleLabel}
		onEventChange={setSelectedEvent}
		onSubjectChange={setFilterSubject}
	/>

	{#if activeMode === 'catalog'}
		<SoalStatusCards cards={statusCards} onSelect={setCatalogStatusFilter} />
	{/if}

	{#if activeMode === 'catalog' && selectedEventId}
		<CatalogTargetPanel
			{selectedEventTitle}
			{questionTargets}
			{eventTargetTotal}
			{eventPublishedTotal}
			{filterSubject}
			{canReviewWorkflow}
			{targetBusy}
			bind:targetQuestionsInput
			{selectedTarget}
			{selectedTargetShortage}
			onSaveTarget={() => void saveQuestionTarget()}
		/>
	{/if}

	{#if activeMode === 'composer'}
		<ComposerDraftNotice onClear={clearAllLocalDrafts} />
	{/if}

	{#if activeMode === 'composer' || offlineQueueCount > 0 || !isOnline}
		<section class="rounded-lg border p-3 text-sm shadow-sm {offlineAuthRequired
			? 'border-destructive/30 bg-destructive/10 text-destructive'
			: !isOnline
				? 'border-warning/30 bg-warning/10 text-warning'
				: offlineQueueCount > 0
					? 'border-primary/20 bg-primary/10 text-primary'
					: 'border-border bg-card text-muted-foreground'}">
			<div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
				<div class="min-w-0">
					<div class="flex flex-wrap items-center gap-2">
						<span class="rounded-full border border-current/25 bg-card px-2 py-0.5 text-[10px] font-bold uppercase tracking-wider">
							{isOnline ? 'Online' : 'Offline'}
						</span>
						<span class="rounded-full border border-current/25 bg-card px-2 py-0.5 text-[10px] font-bold uppercase tracking-wider">
							{offlineQueueCount} antrian
						</span>
						{#if offlineLastSyncLabel()}
							<span class="rounded-full border border-current/25 bg-card px-2 py-0.5 text-[10px] font-bold uppercase tracking-wider">
								Sinkron {offlineLastSyncLabel()}
							</span>
						{/if}
					</div>
					<p class="mt-1 font-semibold">
						{offlineStatus || (offlineQueueCount > 0 ? 'Perubahan lokal menunggu sinkronisasi.' : 'Komposer siap menyimpan draft lokal.')}
					</p>
					<p class="mt-0.5 text-xs opacity-80">
						Draft: {offlineDraftQueueCount} · Review: {offlineReviewQueueCount}. Token login tidak disimpan di IndexedDB; sinkronisasi tetap lewat sesi httpOnly.
					</p>
				</div>
				<div class="flex shrink-0 flex-wrap gap-2">
					{#if offlineAuthRequired}
						<a href={offlineRetryLoginHref()} class="inline-flex h-8 items-center rounded-md border border-current/30 bg-card px-3 text-xs font-semibold hover:bg-muted/50">
							Masuk Ulang
						</a>
					{/if}
					<LoadingButton
						variant="outline"
						size="sm"
						class="h-8 bg-card text-xs"
						onclick={() => void syncBankSoalOfflineQueue(true)}
						loading={offlineSyncBusy}
						loadingLabel="Sinkron..."
						disabled={!canSyncOfflineQueue}
					>
						Sinkronkan
					</LoadingButton>
				</div>
			</div>
		</section>
	{/if}

	{#if activeMode === 'composer' || activeMode === 'import'}
		<section class="rounded-xl border border-border bg-card p-3 shadow-sm">
			<div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
				<div class="min-w-0">
					<p class="text-[11px] font-bold uppercase tracking-wider text-muted-foreground">Cakupan soal</p>
					<h2 class="mt-1 text-sm font-semibold text-foreground">
						{specialEventQuestionMode ? `Khusus ${selectedEventTitle}` : 'Bank soal reusable'}
					</h2>
					<p class="mt-1 text-xs leading-5 text-muted-foreground">
						{specialEventQuestionMode
							? 'Soal disimpan untuk kegiatan terpilih dan tidak masuk stok reusable lintas kegiatan.'
							: 'Soal masuk repositori bersama agar bisa dipakai ulang di paket atau kegiatan lain.'}
					</p>
				</div>
				<label
					for="special-event-question-mode"
					class="flex min-w-[16rem] cursor-pointer items-center justify-between gap-3 rounded-lg border border-primary/20 bg-primary/10 px-3 py-2 text-sm font-semibold text-primary {(!selectedEventId || Boolean(editingId)) ? 'cursor-not-allowed opacity-60' : ''}"
				>
					<span>Khusus kegiatan ini</span>
					<input
						id="special-event-question-mode"
						type="checkbox"
						checked={specialEventQuestionMode}
						disabled={!selectedEventId || Boolean(editingId)}
						onchange={handleSpecialEventModeChange}
						class="rounded accent-green-700"
					/>
				</label>
			</div>
			{#if !selectedEventId}
				<p class="mt-2 text-xs text-muted-foreground">Pilih konteks kegiatan dulu bila soal atau CSV memang hanya untuk satu event.</p>
			{:else if editingId}
				<p class="mt-2 text-xs text-muted-foreground">Cakupan soal yang sudah tersimpan mengikuti data asalnya dan tidak diubah dari komposer cepat.</p>
			{/if}
		</section>
	{/if}

	{#if activeMode === 'review'}
		<ReviewWorkflowQueues
			{revisionQueue}
			{revisionTotal}
			{reviewQueue}
			{reviewTotal}
			{approvedQueue}
			{approvedTotal}
			{revisionSourceOptions}
			{revisionSourceFilter}
			{canReviewWorkflow}
			{workflowBusyId}
			{questionTypeLabel}
			{stemPreview}
			{revisionSourceLabel}
			{revisionReason}
			{canSubmitRevisionReview}
			{canDecideReview}
			{canPublishQuestion}
			onRevisionSourceFilter={setRevisionSourceFilter}
			onShowAllRevisions={showAllRevisions}
			onShowPendingReviews={showPendingReviews}
			onShowApprovedQuestions={showApprovedQuestions}
			onOpenQuestion={openQuestion}
			onOpenReviewDecision={openReviewDecision}
			onSubmitRevisionForReview={(question) => void submitRevisionForReview(question)}
			onPublishQuestion={(question) => void publishQuestion(question)}
			difficultyLabel={DIFFICULTY_LABEL}
		/>
	{/if}

	{#if activeMode === 'composer'}
		{@render composerPanel()}
	{:else if activeMode === 'import'}
		<ImportWorkflowPanel
			{subjects}
			{selectedImportContext}
			{specialEventQuestionMode}
			bind:importSubjectId
			onSubjectChange={setImportSubject}
			{importBusy}
			hasImportFile={importFile !== null}
			importFileName={importFile?.name ?? ''}
			importFileSize={importFile?.size ?? 0}
			{importDryRunDone}
			{importResult}
			{templateBusy}
			onTemplate={() => void downloadQuestionsTemplateCSV()}
			onFileChange={onImportFileChange}
			onDryRun={() => void importLegacyCSV(true)}
			onConfirmImport={() => void importLegacyCSV(false)}
			onBack={() => setModuleMode('catalog')}
		/>
	{/if}

	{#if activeMode === 'catalog'}
	<section class="rounded-xl border border-border bg-card p-3 shadow-sm">
		<div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
			<div>
				<h2 class="text-sm font-semibold text-foreground">Daftar Soal</h2>
				<p class="mt-0.5 text-xs text-muted-foreground">Cari, filter, lalu kelola soal dari tabel utama Bank Soal.</p>
			</div>
			{#if selectedEventId}
				<Button variant="outline" class="h-8 text-xs" onclick={() => setSelectedEvent('')}>Lepas Filter Kegiatan</Button>
			{/if}
		</div>
		<div class="mt-3 flex flex-wrap gap-2">
			<label for="question-search" class="sr-only">Cari soal berdasarkan isi atau kode</label>
			<Input
				id="question-search"
				placeholder="Cari soal..."
				value={search}
				oninput={onSearchInput}
				class="h-8 w-48 text-sm"
			/>
			<label for="question-workflow-filter" class="sr-only">Filter status workflow daftar soal</label>
				<select
					id="question-workflow-filter"
					bind:value={filterWorkflow}
					onchange={onWorkflowFilterChange}
					class="h-8 rounded-md border border-border bg-card px-2.5 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring"
			>
				<option value="">Semua Status</option>
				<option value="draft">Draft</option>
				<option value="review">Menunggu Review</option>
					<option value="approved">Disetujui</option>
					<option value="rejected">Perlu Revisi</option>
				</select>
				{#if hasCatalogQuickFilter}
					<Button variant="outline" class="h-8 text-xs" onclick={clearCatalogQuickFilters}>Bersihkan Filter</Button>
				{/if}
				{#if totalItems > 0}
					<span class="flex items-center text-xs text-muted-foreground">{totalItems} soal</span>
				{/if}
		</div>
			{#if selectedQuestionIds.length > 0}
				<div class="mt-3 rounded-lg border border-success/20 bg-success/10 p-3">
					<div class="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
						<div>
							<p class="text-sm font-semibold text-success">{selectedVisibleCount} soal di halaman ini dipilih</p>
							<p class="text-xs text-success">Siap diputuskan: {selectedReviewEligibleCount}; siap diterbitkan: {selectedPublishEligibleCount}. Aksi yang tidak memenuhi syarat otomatis dilewati.</p>
						</div>
					<div class="flex flex-wrap gap-2">
						<Input placeholder="Catatan untuk Setujui/Minta Revisi..." aria-label="Catatan aksi massal review soal" bind:value={bulkNotes} class="h-8 min-w-56 bg-card text-xs" />
						<LoadingButton variant="outline" size="sm" onclick={() => void runBulkWorkflow('approve')} loading={bulkBusy} loadingLabel="Memproses..." disabled={bulkBusy || selectedReviewEligibleCount === 0} class="h-8 bg-card text-success">Setujui</LoadingButton>
						<LoadingButton variant="outline" size="sm" onclick={() => void runBulkWorkflow('reject')} loading={bulkBusy} loadingLabel="Memproses..." disabled={bulkBusy || selectedReviewEligibleCount === 0} class="h-8 bg-card text-destructive">Minta Revisi</LoadingButton>
						<LoadingButton variant="outline" size="sm" onclick={() => void runBulkWorkflow('publish')} loading={bulkBusy} loadingLabel="Memproses..." disabled={bulkBusy || selectedPublishEligibleCount === 0} class="h-8 bg-card text-success">Terbitkan</LoadingButton>
						<Button variant="outline" size="sm" class="h-8 bg-card" onclick={clearSelection}>Bersihkan</Button>
					</div>
				</div>
			</div>
		{/if}
	</section>

	<!-- Table -->
	{#if openMenuId !== ''}
		<div class="fixed inset-0 z-10" onclick={closeRowMenu} aria-hidden="true"></div>
	{/if}
	<div class="rounded-lg border border-border bg-card overflow-hidden">
		<Table.Root>
				<Table.Header>
					<Table.Row class="bg-muted/50 text-xs">
							<Table.Head class="w-12 text-muted-foreground">
								<input
									bind:this={visibleSelectionCheckbox}
									type="checkbox"
									checked={allVisibleQuestionsSelected}
									disabled={visibleQuestionIds.length === 0}
								aria-label="Pilih semua soal di halaman ini"
								onchange={onVisibleQuestionSelectionChange}
								class="rounded accent-green-700"
							/>
						</Table.Head>
						<Table.Head class="text-muted-foreground">Isi Soal</Table.Head>
						<Table.Head class="w-32 text-muted-foreground">Mapel</Table.Head>
						<Table.Head class="w-28 hidden sm:table-cell text-muted-foreground">Status</Table.Head>
					<Table.Head class="w-44 text-right text-muted-foreground">Aksi</Table.Head>
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
									<Table.Cell class="text-right"><Skeleton class="ml-auto h-8 w-16" /></Table.Cell>
								</Table.Row>
							{/each}
						{/snippet}
						{#snippet failed(error, reset)}
							<Table.Row>
								<Table.Cell colspan={5} class="py-6">
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
									<Table.Cell colspan={5} class="py-10 text-center text-sm text-muted-foreground">
										Belum ada soal.
									<button onclick={openCreate} class="text-success underline ml-1"
										>Buat soal pertama →</button
									>
								</Table.Cell>
							</Table.Row>
						{:else}
							{#each currentQuestions as q, i (q.id)}
								<Table.Row
									class="hover:bg-muted/50 cursor-pointer"
									onclick={() => openQuestion(q)}
								>
								<Table.Cell class="text-xs text-muted-foreground">
									<input
										type="checkbox"
										checked={selectedQuestionIds.includes(q.id)}
										aria-label={`Pilih soal ${(overview.page - 1) * PAGE_SIZE + i + 1}`}
										onclick={(event) => event.stopPropagation()}
										onchange={() => toggleQuestionSelection(q.id)}
										class="rounded accent-green-700"
									/>
								</Table.Cell>
									<Table.Cell class="text-sm text-foreground max-w-xs">
										<div class="truncate">{stemPreview(q)}</div>
										{#if q.author_username}
											<div class="text-[10px] text-muted-foreground mt-0.5">{q.author_username}</div>
										{/if}
										<div class="mt-1 flex flex-wrap gap-1">
											<span class="rounded bg-success/10 px-1.5 py-0.5 text-[10px] font-medium text-success">
												{questionTypeLabel(q.question_type)}
											</span>
											<span class="rounded bg-muted px-1.5 py-0.5 text-[10px] font-medium text-muted-foreground">
												{q.suggested_mode ?? q.authoring_mode ?? 'beginner'}
											</span>
											{#if questionUsageLocked(q)}
												<span class="rounded bg-destructive/10 px-1.5 py-0.5 text-[10px] font-medium text-destructive">
													Terkunci: {questionUsageText(q)}
												</span>
											{/if}
										</div>
										{#if q.workflow_status === 'rejected'}
											<div class="mt-1 rounded border border-destructive/30 bg-destructive/10 px-2 py-1 text-[11px] leading-relaxed text-destructive">
												<span class="font-semibold">{revisionSourceLabel(q)}:</span> {revisionReason(q)}
											</div>
										{/if}
									</Table.Cell>
									<Table.Cell class="text-xs text-muted-foreground truncate max-w-[8rem]">
										{q.subject_name || q.subject_code || '-'}
									</Table.Cell>
									<Table.Cell class="hidden sm:table-cell">
										<span
											class="inline-block rounded px-1.5 py-0.5 text-[10px] font-medium {workflowClass(q.workflow_status)}"
										>
											{WORKFLOW_LABEL[q.workflow_status] ?? q.workflow_status ?? '-'}
										</span>
									</Table.Cell>
									<Table.Cell class="text-right">
										<button
											onclick={(e) => {
												e.stopPropagation();
												openQuestion(q);
											}}
											class="rounded px-2 py-1 text-xs text-muted-foreground transition-colors hover:bg-muted"
										>
										{isQuickEditable(q) ? 'Edit' : 'Lihat'}
										</button>
										{#if q.workflow_status === 'rejected'}
											<button
												onclick={(e) => {
													e.stopPropagation();
													void submitRevisionForReview(q);
												}}
												disabled={!canSubmitRevisionReview(q) || (workflowBusyId !== '' && workflowBusyId !== q.id)}
												class="rounded px-2 py-1 text-xs text-success transition-colors hover:bg-success/10 disabled:cursor-not-allowed disabled:opacity-40"
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
												class="rounded px-2 py-1 text-xs text-warning transition-colors hover:bg-warning/10 disabled:cursor-not-allowed disabled:opacity-40"
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
												class="rounded px-2 py-1 text-xs text-success transition-colors hover:bg-success/10 disabled:cursor-not-allowed disabled:opacity-40"
											>
												{workflowBusyId === q.id ? 'Terbit...' : 'Terbitkan'}
											</button>
										{/if}
										<div class="relative inline-block">
											<button
												type="button"
												onclick={(e) => toggleRowMenu(q.id, e)}
												aria-label="Aksi lainnya"
												class="rounded px-2 py-1 text-xs text-muted-foreground hover:bg-muted"
											>⋯</button>
											{#if openMenuId === q.id}
												<div
													class="absolute right-0 top-full z-20 min-w-[7.5rem] rounded-lg border border-border bg-card py-1 shadow-lg"
													role="menu"
												>
													<button
														type="button"
														onclick={(e) => { e.stopPropagation(); closeRowMenu(); void duplicateQuestion(q.id); }}
														disabled={duplicateBusyId === q.id}
														class="block w-full px-3 py-1.5 text-left text-xs text-success hover:bg-success/10 disabled:opacity-50"
														role="menuitem"
													>
														{duplicateBusyId === q.id ? 'Menyalin...' : 'Duplikat'}
													</button>
													{#if canDeleteQuestion}
														<button
															type="button"
															onclick={(e) => { e.stopPropagation(); closeRowMenu(); void deleteQuestion(q.id); }}
															disabled={questionUsageLocked(q)}
															class="block w-full px-3 py-1.5 text-left text-xs text-destructive hover:bg-destructive/10 disabled:cursor-not-allowed disabled:opacity-40"
															role="menuitem"
														>Hapus</button>
													{/if}
												</div>
											{/if}
										</div>
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
		<div class="flex items-center justify-between text-sm text-muted-foreground">
			<span class="text-xs">{totalItems} soal total</span>
			<div class="flex items-center gap-1">
				<Button
					variant="outline"
					class="h-7 w-7 p-0 text-xs"
					aria-label="Halaman sebelumnya"
					disabled={currentPage <= 1}
					onclick={() => load(currentPage - 1)}
				>
					‹
				</Button>
				<span class="px-2 text-xs">Hal {currentPage} / {pageCount}</span>
				<Button
					variant="outline"
					class="h-7 w-7 p-0 text-xs"
					aria-label="Halaman berikutnya"
					disabled={currentPage >= pageCount}
					onclick={() => load(currentPage + 1)}
				>
					›
				</Button>
			</div>
		</div>
	{/if}
	{/if}
</div>

<Dialog.Root bind:open={reviewDecisionOpen}>
	<Dialog.Content>
		<div class="w-[min(94vw,42rem)] space-y-4 p-5">
			<div>
				<p class="text-xs font-bold uppercase tracking-wider text-success">Periksa Satu Soal</p>
				<h2 class="mt-1 text-base font-semibold text-foreground">
					{reviewDecision === 'approve' ? 'Pilihan saat ini: Setujui Soal' : 'Pilihan saat ini: Minta Revisi Soal'}
				</h2>
				<p class="mt-1 text-xs text-muted-foreground">Baca satu soal ini sampai lengkap, lalu pilih salah satu keputusan yang jelas untuk guru dan admin.</p>
			</div>

			{#if reviewDecisionQuestion}
				<div class="grid gap-2 text-xs sm:grid-cols-3">
					<div class="rounded-md border border-warning/30 bg-warning/10 px-3 py-2 text-warning">
						<p class="font-semibold">Status sekarang</p>
						<p class="mt-0.5">{WORKFLOW_LABEL[reviewDecisionQuestion.workflow_status] ?? reviewDecisionQuestion.workflow_status}</p>
					</div>
					<div class="rounded-md border border-border bg-muted/50 px-3 py-2 text-foreground">
						<p class="font-semibold">Aksi aman</p>
						<p class="mt-0.5">Setujui atau kembalikan revisi.</p>
					</div>
					<div class="rounded-md border border-success/20 bg-success/10 px-3 py-2 text-success">
						<p class="font-semibold">Setelah disetujui</p>
						<p class="mt-0.5">Admin dapat menerbitkan ke paket ujian.</p>
					</div>
				</div>
				<div class="rounded-lg border border-border bg-muted/50 p-3">
					<div class="mb-2 flex flex-wrap items-center gap-1.5 text-[11px] text-muted-foreground">
						<span class="rounded bg-card px-1.5 py-0.5 font-semibold text-success">{questionTypeLabel(reviewDecisionQuestion.question_type)}</span>
						<span>{reviewDecisionQuestion.subject_name || reviewDecisionQuestion.subject_code || 'Mapel belum ada'}</span>
						<span>{reviewDecisionQuestion.code || 'Tanpa kode'}</span>
						{#if reviewDecisionQuestion.author_username}<span>Guru: {reviewDecisionQuestion.author_username}</span>{/if}
					</div>
					<div class="max-h-64 space-y-3 overflow-auto rounded-md border border-border bg-card p-3 text-sm">
						{#if reviewDecisionQuestion.stimulus_html}
							<div class="rounded border border-border bg-muted/50 p-2">
								<p class="mb-1 text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">Stimulus</p>
								<RichContent html={reviewDecisionQuestion.stimulus_html} class="prose prose-sm max-w-none text-foreground latex-preview" />
							</div>
						{/if}
						<div>
							<p class="mb-1 text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">Pertanyaan</p>
							<RichContent html={reviewDecisionQuestion.stem_html || reviewDecisionQuestion.question_text || 'Isi soal belum tersedia'} class="prose prose-sm max-w-none text-foreground latex-preview" />
						</div>
						{#if reviewDecisionQuestion.options.length > 0}
							<div class="space-y-1.5">
								<p class="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">Opsi / Pasangan</p>
								{#each reviewDecisionQuestion.options as opt, i (`review-option-${reviewDecisionQuestion.id}-${i}`)}
									{@const secondary = optionSecondaryContent(opt)}
									<div class="rounded border border-border bg-muted/50 px-2 py-1.5">
										<div class="flex gap-2">
											<span class="mt-0.5 text-xs font-bold text-success">{opt.label || opt.match_label || i + 1}</span>
											<div class="min-w-0 flex-1 text-xs text-foreground">
												<RichContent html={optionPrimaryContent(opt)} class="latex-preview" />
												{#if secondary}
													<div class="mt-1 border-t border-border pt-1 text-muted-foreground">
														<RichContent html={secondary} class="latex-preview" />
													</div>
												{/if}
											</div>
										</div>
									</div>
								{/each}
							</div>
						{/if}
						<div class="space-y-2 rounded border border-success/20 bg-success/10 px-2 py-1.5 text-xs text-success">
							<p><span class="font-semibold">Kunci/Rubrik:</span> {answerKeyLabelForQuestion(reviewDecisionQuestion)}</p>
							{#if reviewDecisionQuestion.rubric_html}
								<div class="rounded border border-success/20 bg-card p-2">
									<p class="mb-1 text-[10px] font-semibold uppercase tracking-wider text-success">Rubrik / Pedoman Koreksi</p>
									<RichContent html={reviewDecisionQuestion.rubric_html} class="prose prose-sm max-w-none text-success latex-preview" />
								</div>
							{/if}
							{#if reviewDecisionQuestion.explanation_html}
								<div class="rounded border border-border bg-card p-2">
									<p class="mb-1 text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">Pembahasan / Catatan Internal</p>
									<RichContent html={reviewDecisionQuestion.explanation_html} class="prose prose-sm max-w-none text-foreground latex-preview" />
								</div>
							{/if}
						</div>
						{#if reviewDecisionQuestion.review_notes}
							<div class="rounded border border-warning/30 bg-warning/10 px-2 py-1.5 text-xs leading-relaxed text-warning">
								<span class="font-semibold">Catatan sebelumnya:</span> {revisionReason(reviewDecisionQuestion)}
							</div>
						{/if}
						{@render questionTimelinePanel()}
					</div>
				</div>
			{/if}

			<div class="grid grid-cols-2 gap-2">
			<button
				type="button"
				onclick={() => (reviewDecision = 'approve')}
				aria-pressed={reviewDecision === 'approve'}
				aria-label="Pilih keputusan Setujui untuk soal ini"
				class="rounded-md border px-3 py-2 text-left text-sm transition-colors {reviewDecision === 'approve'
						? 'border-success/20 bg-success/10 font-semibold text-success'
						: 'border-border bg-card text-muted-foreground hover:bg-muted/50'}"
				>
					Setujui
					<span class="mt-0.5 block text-[11px] font-normal text-muted-foreground">Soal masuk status disetujui.</span>
				</button>
			<button
				type="button"
				onclick={() => (reviewDecision = 'reject')}
				aria-pressed={reviewDecision === 'reject'}
				aria-label="Pilih keputusan Minta Revisi untuk soal ini"
				class="rounded-md border px-3 py-2 text-left text-sm transition-colors {reviewDecision === 'reject'
						? 'border-destructive/30 bg-destructive/10 font-semibold text-destructive'
						: 'border-border bg-card text-muted-foreground hover:bg-muted/50'}"
				>
					Minta Revisi
					<span class="mt-0.5 block text-[11px] font-normal text-muted-foreground">Kembalikan ke guru dengan catatan.</span>
				</button>
			</div>

			<div>
				<label for="review-decision-notes" class="mb-1 block text-xs font-medium text-muted-foreground">
					Catatan Reviewer {#if reviewDecision === 'reject'}<span class="text-destructive">*</span>{/if}
				</label>
				<Textarea
					id="review-decision-notes"
					rows={3}
					bind:value={reviewDecisionNotes}
					placeholder={reviewDecision === 'approve' ? 'Opsional: catatan persetujuan.' : 'Tuliskan bagian yang harus diperbaiki guru.'}
				/>
			</div>

			<div class="flex justify-end gap-2 border-t border-border pt-4">
				<Button variant="outline" onclick={closeReviewDecision} disabled={workflowBusyId !== ''}>
					Batal
				</Button>
				<LoadingButton
					onclick={() => void submitReviewDecision()}
					loading={workflowBusyId !== ''}
					loadingLabel="Menyimpan..."
					disabled={!reviewDecisionQuestion || workflowBusyId !== ''}
					class={reviewDecision === 'approve' ? 'bg-success text-background hover:bg-success' : 'bg-destructive text-destructive-foreground hover:bg-destructive'}
				>
					{reviewDecision === 'approve' ? 'Setujui Soal' : 'Minta Revisi'}
				</LoadingButton>
			</div>
		</div>
	</Dialog.Content>
</Dialog.Root>

<Dialog.Root bind:open={questionPreviewOpen}>
	<Dialog.Content>
		<div class="w-[min(94vw,42rem)] space-y-4 p-5">
			<div class="flex items-start justify-between gap-3">
				<div>
					<p class="text-xs font-bold uppercase tracking-wider text-muted-foreground">Lihat Soal Terkunci</p>
					<h2 class="mt-1 text-base font-semibold text-foreground">Detail Read-only</h2>
					<p class="mt-1 text-xs text-muted-foreground">Soal tidak bisa diedit langsung karena sudah dipakai, masuk review/publikasi, atau tipe belum kompatibel dengan editor cepat.</p>
				</div>
				{#if questionPreviewLoading}<span class="rounded-full bg-muted px-2 py-1 text-xs text-muted-foreground">Memuat detail...</span>{/if}
			</div>

			{#if questionPreview}
				<div class="rounded-lg border border-border bg-muted/50 p-3">
					<div class="mb-2 flex flex-wrap items-center gap-1.5 text-[11px] text-muted-foreground">
						<span class="rounded bg-card px-1.5 py-0.5 font-semibold text-success">{questionTypeLabel(questionPreview.question_type)}</span>
						<span>{questionPreview.subject_name || questionPreview.subject_code || 'Mapel belum ada'}</span>
						<span>{WORKFLOW_LABEL[questionPreview.workflow_status] ?? questionPreview.workflow_status}</span>
						<span>{questionUsageText(questionPreview)}</span>
					</div>
					<div class="max-h-80 space-y-3 overflow-auto rounded-md border border-border bg-card p-3 text-sm">
						{#if questionPreview.stimulus_html}
							<div class="rounded border border-border bg-muted/50 p-2">
								<p class="mb-1 text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">Stimulus</p>
								<RichContent html={questionPreview.stimulus_html} class="prose prose-sm max-w-none text-foreground latex-preview" />
							</div>
						{/if}
						<div>
							<p class="mb-1 text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">Pertanyaan</p>
							<RichContent html={questionPreview.stem_html || questionPreview.question_text || 'Isi soal belum tersedia'} class="prose prose-sm max-w-none text-foreground latex-preview" />
						</div>
						{#if questionPreview.options.length > 0}
							<div class="space-y-1.5">
								<p class="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">Opsi / Pasangan</p>
							{#each questionPreview.options as opt, i (`preview-option-${questionPreview.id}-${i}`)}
								{@const secondary = optionSecondaryContent(opt)}
								<div class="rounded border border-border bg-muted/50 px-2 py-1.5 text-xs text-foreground">
									<span class="font-bold text-success">{opt.label || opt.match_label || i + 1}.</span>
									<RichContent html={optionPrimaryContent(opt)} class="mt-1 latex-preview" />
									{#if secondary}
										<div class="mt-1 border-t border-border pt-1 text-muted-foreground">
											<RichContent html={secondary} class="latex-preview" />
										</div>
									{/if}
								</div>
								{/each}
							</div>
						{/if}
						<div class="rounded border border-success/20 bg-success/10 px-2 py-1.5 text-xs text-success">
							<span class="font-semibold">Kunci/Rubrik:</span> {answerKeyLabelForQuestion(questionPreview)}
						</div>
						{#if questionPreview.rubric_html}
							<div class="rounded border border-warning/30 bg-warning/10 p-2 text-xs text-warning">
								<p class="mb-1 font-semibold">Rubrik / Pedoman Koreksi</p>
								<RichContent html={questionPreview.rubric_html} class="prose prose-sm max-w-none latex-preview" />
							</div>
						{/if}
						{#if questionPreview.explanation_html}
							<div class="rounded border border-border bg-muted/50 p-2 text-xs text-foreground">
								<p class="mb-1 font-semibold">Pembahasan</p>
								<RichContent html={questionPreview.explanation_html} class="prose prose-sm max-w-none latex-preview" />
							</div>
						{/if}
						{@render questionTimelinePanel()}
					</div>
				</div>
			{/if}

			<div class="flex justify-end gap-2 border-t border-border pt-4">
				<Button variant="outline" onclick={closeQuestionPreview}>Tutup</Button>
				{#if questionPreview}
					<Button class="bg-success text-background hover:bg-success" onclick={duplicatePreviewQuestion}>Duplikat untuk Revisi</Button>
				{/if}
			</div>
		</div>
	</Dialog.Content>
</Dialog.Root>

{#snippet questionTimelinePanel()}
	<div class="rounded border border-border bg-card p-2 text-xs">
		<div class="mb-2 flex items-center justify-between gap-2">
			<p class="font-semibold uppercase tracking-wider text-muted-foreground">Timeline Soal</p>
			{#if questionTimelineLoading}<span class="text-muted-foreground">Memuat...</span>{/if}
		</div>
		{#if questionTimeline.length > 0}
			<div class="space-y-2">
				{#each questionTimeline.slice(0, 8) as item, index (`timeline-${item.id ?? index}`)}
					<div class="rounded border border-border bg-muted/50 px-2 py-1.5">
						<div class="flex flex-wrap items-center gap-1.5">
							<span class="font-semibold text-success">{item.action ?? item.status ?? 'Perubahan'}</span>
							{#if item.actor_username}<span class="text-muted-foreground">oleh {item.actor_username}</span>{/if}
							{#if item.created_at}<span class="text-muted-foreground">{new Date(item.created_at).toLocaleString('id-ID')}</span>{/if}
						</div>
						{#if item.notes}<p class="mt-1 text-muted-foreground">{item.notes}</p>{/if}
					</div>
				{/each}
			</div>
		{:else}
			<p class="text-muted-foreground">Timeline belum tersedia dari backend.</p>
		{/if}
	</div>
{/snippet}

{#snippet composerPreview()}
	<div class="mb-4">
		<div class="mb-1 flex items-center justify-between">
			<span class="text-[10px] font-semibold uppercase tracking-wide text-muted-foreground">Kesiapan Review</span>
			<span
				class="text-sm font-bold {readinessScore === 100
					? 'text-success'
					: readinessScore >= 50
						? 'text-warning'
						: 'text-destructive'}"
			>
				{readinessScore}%
			</span>
		</div>
		<div class="h-2 overflow-hidden rounded-full bg-border">
			<div
				class="h-2 rounded-full transition-all duration-300 {readinessScore === 100
					? 'bg-success'
					: readinessScore >= 50
						? 'bg-warning'
						: 'bg-destructive'}"
				style="width: {readinessScore}%"
			></div>
		</div>
		{#if validationIssues.length > 0}
			<ul class="mt-1.5 space-y-0.5">
				{#each validationIssues as issue (`validation-${issue}`)}
					<li class="text-[10px] text-destructive">• {issue}</li>
				{/each}
			</ul>
		{:else}
			<p class="mt-1 text-[10px] text-success">Soal siap diajukan review.</p>
		{/if}
	</div>

	<div class="mb-4">
		<p class="mb-2 text-[10px] font-semibold uppercase tracking-wide text-muted-foreground">
			Sinyal Kualitas
		</p>
		<div class="space-y-2">
			{#each qualitySignals as sig (sig.label)}
				<div class="flex items-start gap-2">
					<span class="mt-0.5 shrink-0 text-sm font-bold {sig.status === 'good' ? 'text-success' : 'text-warning'}">
						{sig.status === 'good' ? '✓' : '!'}
					</span>
					<div>
						<div class="text-xs font-medium text-foreground">{sig.label}</div>
						<div class="text-[10px] text-muted-foreground">{sig.desc}</div>
					</div>
				</div>
			{/each}
		</div>
	</div>

	<div class="my-3 border-t border-border"></div>

	<div>
		<p class="mb-2 text-[10px] font-semibold uppercase tracking-wide text-muted-foreground">
			Preview Siswa
		</p>
		<div class="rounded-lg border border-border bg-card p-3 space-y-3" dir={fIsRtl ? 'rtl' : undefined}>
			{#if isAdvanceMode && fStimulus}
				<div class="rounded-md border border-border bg-muted/50 p-2">
					<p class="mb-1 text-[10px] font-semibold uppercase tracking-wide text-muted-foreground">Stimulus</p>
					<RichContent html={fStimulus} class="prose prose-sm max-w-none text-foreground latex-preview text-sm" />
				</div>
			{/if}
			{#if fStem}
				<RichContent
					html={fStem}
					class="prose prose-sm max-w-none text-foreground latex-preview text-sm"
				/>
			{:else}
				<p class="text-xs text-muted-foreground italic">Isi soal belum dimasukkan</p>
			{/if}

			{#if isEssay}
				<div class="space-y-2 border-t border-border pt-2">
					<div class="rounded-md border border-dashed border-success/20 bg-success/10 px-3 py-2 text-xs text-success">
						Siswa akan melihat kotak jawaban uraian pada aplikasi ujian.
					</div>
					{#if richTextHasContent(fRubric)}
						<div class="rounded-md border border-warning/30 bg-warning/10 p-2">
							<p class="mb-1 text-[10px] font-semibold uppercase tracking-wide text-warning">Pedoman koreksi guru</p>
							<RichContent html={fRubric} class="prose prose-sm max-w-none text-warning latex-preview text-sm" />
						</div>
					{/if}
				</div>
			{:else if isShortAnswer}
				<div class="space-y-2 border-t border-border pt-2">
					<div class="rounded-md border border-dashed border-success/20 bg-success/10 px-3 py-2 text-xs text-success">
						Siswa akan melihat kolom jawaban isian singkat.
					</div>
					{#if shortAnswerAliases.length > 0}
						<div class="rounded-md border border-border bg-muted/50 px-3 py-2 text-xs text-foreground">
							<span class="font-semibold text-muted-foreground">Jawaban diterima:</span> {shortAnswerAliases.join(' / ')}
						</div>
					{/if}
				</div>
			{:else if isMatching}
				<div class="space-y-3 border-t border-border pt-2">
					<div class="rounded-md border border-dashed border-success/20 bg-success/10 px-3 py-2 text-xs text-success">
						Siswa akan memilih nomor pasangan kanan untuk setiap item kiri.
					</div>
					<div class="grid grid-cols-1 gap-2 md:grid-cols-2">
						<div class="space-y-1.5">
							<p class="text-[10px] font-semibold uppercase tracking-wide text-muted-foreground">Kolom Kiri</p>
							{#each fMatchingPairs as pair, i (`preview-match-left-${i}`)}
								<div class="flex items-start gap-2 rounded-md border border-border bg-card px-2 py-1.5 text-xs">
									<span class="shrink-0 font-bold text-success">{optionLabelAt(i)}.</span>
									{#if richTextHasContent(pair.left)}
										<RichContent html={pair.left} class="latex-preview min-w-0 flex-1" />
									{:else}
										<span class="italic text-muted-foreground">(kiri kosong)</span>
									{/if}
								</div>
							{/each}
						</div>
						<div class="space-y-1.5">
							<p class="text-[10px] font-semibold uppercase tracking-wide text-muted-foreground">Pilihan Kanan</p>
							{#each fMatchingPairs as pair, i (`preview-match-right-${i}`)}
								<div class="flex items-start gap-2 rounded-md border border-border bg-muted/50 px-2 py-1.5 text-xs">
									<span class="shrink-0 font-bold text-muted-foreground">{i + 1}.</span>
									{#if richTextHasContent(pair.right)}
										<RichContent html={pair.right} class="latex-preview min-w-0 flex-1" />
									{:else}
										<span class="italic text-muted-foreground">(kanan kosong)</span>
									{/if}
								</div>
							{/each}
							{#each fMatchingDistractors as distractor, i (`preview-match-distractor-${i}`)}
								<div class="flex items-start gap-2 rounded-md border border-warning/30 bg-warning/10 px-2 py-1.5 text-xs">
									<span class="shrink-0 font-bold text-warning">{fMatchingPairs.length + i + 1}.</span>
									{#if richTextHasContent(distractor)}
										<RichContent html={distractor} class="latex-preview min-w-0 flex-1" />
									{:else}
										<span class="italic text-warning">(distraktor kosong)</span>
									{/if}
								</div>
							{/each}
						</div>
					</div>
					<div class="rounded-md border border-border bg-muted/50 px-3 py-2 text-xs text-foreground">
						<span class="font-semibold text-muted-foreground">Kunci otomatis:</span> {buildMatchingAnswerKey(fMatchingPairs.length)}
					</div>
				</div>
			{:else if fOptions.some(richTextHasContent)}
				<div class="space-y-1.5 border-t border-border pt-2">
					{#each fOptions as opt, i (`preview-option-${i}`)}
						{@const label = optionLabelAt(i)}
						{@const isAnswer = isAnswerLabelSelected(label)}
						<div class="flex items-start gap-2 text-sm {isAnswer ? 'text-success font-medium' : 'text-foreground'}">
							<span class="shrink-0 font-bold">{label}.</span>
							{#if richTextHasContent(opt)}
								<RichContent html={opt} class="latex-preview min-w-0 flex-1" />
							{:else}
								<span class="italic text-muted-foreground">(kosong)</span>
							{/if}
							{#if isAnswer}
								<span class="ml-auto shrink-0 text-xs text-success">✓</span>
							{/if}
						</div>
					{/each}
				</div>
			{/if}
		</div>
	</div>
{/snippet}

{#snippet composerStageRail()}
	<div class="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
		{#each composerStageCards as card (card.label)}
			<button
				type="button"
				onclick={() => scrollToComposerSection(card.targetId)}
				class={`rounded-xl border p-3 text-left transition hover:-translate-y-0.5 hover:shadow-sm ${composerStageToneClass(card.tone)}`}
			>
				<div class="flex items-start justify-between gap-3">
					<div class="min-w-0">
						<p class="text-[10px] font-black uppercase tracking-[0.2em] opacity-70">{card.label}</p>
						<p class="mt-1 truncate text-sm font-bold">{card.desc}</p>
					</div>
					<span class="shrink-0 rounded-full bg-card/70 px-2 py-1 text-[10px] font-bold uppercase tracking-wide">{card.status}</span>
				</div>
			</button>
		{/each}
	</div>
{/snippet}

<!-- ── Composer Inline Section ─────────────────────────────────────────────── -->
{#snippet composerPanel()}
	<section class="soal-composer-inline relative flex min-h-[42rem] flex-col overflow-hidden rounded-2xl border border-border bg-card text-foreground shadow-sm">
			<div class="shrink-0 border-b border-primary/20 bg-gradient-to-r from-primary/10 via-card to-warning/10 px-4 py-4 md:px-5">
				<div class="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
					<div class="min-w-0">
						<p class="text-[10px] font-black uppercase tracking-[0.28em] text-primary">Studio Komposer Bank Soal</p>
						<div class="mt-1 flex flex-wrap items-baseline gap-x-2 gap-y-1">
							<h2 class="text-xl font-black uppercase italic tracking-tight text-foreground">
								{editingId ? 'Edit Butir Soal' : 'Penyusunan Soal Baru'}
							</h2>
						</div>
						<p class="mt-2 max-w-3xl text-sm leading-6 text-muted-foreground">
							Susun metadata, naskah, kunci/rubrik, lalu cek preview siswa sebelum diajukan review. Autosave lokal dan Ctrl+S tetap aktif.
						</p>
						<div class="mt-3 flex flex-wrap gap-2">
							<span class="rounded-full border border-primary/20 bg-card px-2.5 py-1 text-[10px] font-bold uppercase tracking-wide text-primary">{composerScopeLabel}</span>
							<span class="rounded-full border border-border bg-card px-2.5 py-1 text-[10px] font-bold uppercase tracking-wide text-muted-foreground">{questionTypeConfig.label}</span>
							<span class="rounded-full border border-warning/30 bg-card px-2.5 py-1 text-[10px] font-bold uppercase tracking-wide text-warning">{composerModeLabel}</span>
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

			<div class="min-h-0 flex-1 overflow-y-auto bg-muted/50 p-4 md:p-5">
				<section class="mb-4 rounded-2xl border border-primary/20 bg-card p-4 shadow-sm">
					<div class="mb-3 flex flex-col gap-1 md:flex-row md:items-end md:justify-between">
						<div>
							<p class="text-[10px] font-black uppercase tracking-[0.24em] text-primary">Alur penyusunan</p>
							<h3 class="text-sm font-black uppercase text-foreground">Klik kartu untuk lompat ke bagian editor</h3>
						</div>
						<span class="text-xs font-semibold text-muted-foreground">Kesiapan review {readinessScore}%</span>
					</div>
					{@render composerStageRail()}
				</section>

				<section class="mb-4 rounded-xl border border-border bg-card p-4 shadow-sm">
					<div class="flex flex-wrap items-center gap-2">
						<div class="flex min-w-[11rem] items-center gap-2">
							<span class="text-[10px] font-black uppercase tracking-wider text-muted-foreground">Kesiapan Kirim</span>
							<div class="h-1.5 w-20 overflow-hidden rounded-full bg-border">
								<div
									class="h-1.5 rounded-full transition-all duration-300 {readinessScore === 100
										? 'bg-success'
										: readinessScore >= 50
											? 'bg-warning'
											: 'bg-destructive'}"
									style="width: {readinessScore}%"
								></div>
							</div>
							<span class="text-xs font-bold {readinessScore === 100 ? 'text-success' : 'text-destructive'}">{readinessScore}%</span>
						</div>
							<span class="rounded-full border px-2 py-1 text-[10px] font-semibold {validationIssues.length === 0 ? 'border-primary/20 bg-card text-primary' : 'border-destructive/30 bg-card text-destructive'}">
								{validationIssues.length === 0 ? 'Siap review' : `${validationIssues.length} wajib belum lengkap`}
							</span>
							{#if firstComposerIssue}
								<button
									type="button"
									onclick={scrollToFirstComposerIssue}
									class="rounded-full border border-destructive/30 bg-destructive/10 px-2 py-1 text-[10px] font-semibold text-destructive hover:bg-destructive/15"
								>
									Lengkapi: {firstComposerIssue.message}
								</button>
							{/if}
							<span class="rounded-full border border-warning/30 bg-card px-2 py-1 text-[10px] font-semibold text-warning">
								{qualityWarningCount} sinyal kualitas perlu cek
							</span>
						<span class="rounded-full border border-primary/20 bg-card px-2 py-1 text-[10px] font-semibold text-primary">
							{draftStatusLabel()}
						</span>
						{#if selectedEventId}
							<span class="rounded-full border border-primary/20 bg-card px-2 py-1 text-[10px] font-semibold text-primary">
								Konteks kegiatan: {selectedEventTitle}
							</span>
						{/if}
					</div>
				</section>

				<div class="grid grid-cols-1 gap-5 xl:grid-cols-[minmax(0,1.55fr)_minmax(18rem,0.5fr)]">
					<div class={`space-y-4 min-w-0 ${composerMobilePanel === 'preview' ? 'hidden lg:block' : 'block'}`}>
						<section id="composer-metadata" class="scroll-mt-4 rounded-lg border border-border bg-card p-3 shadow-sm">
							<div class="mb-3 flex flex-wrap items-center gap-2 border-b border-border pb-3">
								<span class="text-[10px] font-black uppercase tracking-[0.2em] text-foreground">Bentuk Soal</span>
								<div class="flex flex-wrap rounded-lg border border-border bg-muted/50 p-1">
									{#each QUESTION_TYPE_CONFIGS as typeConfig (typeConfig.id)}
										<button
											type="button"
											onclick={() => setQuestionType(typeConfig.id)}
											title={typeConfig.desc}
											class="h-7 rounded-md px-2.5 text-[10px] font-bold uppercase tracking-wide {fQuestionType === typeConfig.id ? 'bg-success text-background' : 'text-muted-foreground hover:bg-card'}"
										>
											{typeConfig.shortLabel}
										</button>
									{/each}
								</div>
								<span class="ml-2 text-[10px] font-black uppercase tracking-[0.2em] text-foreground">Mode</span>
								<div class="inline-flex rounded-lg border border-border bg-muted/50 p-1">
									<button
										type="button"
										onclick={() => setAuthoringMode('beginner')}
										class="h-7 rounded-md px-3 text-[10px] font-bold uppercase tracking-wide {fAuthoringMode === 'beginner' ? 'bg-success text-background' : 'text-muted-foreground hover:bg-card'}"
									>
										Pemula
									</button>
									<button
										type="button"
										onclick={() => setAuthoringMode('advance')}
										class="h-7 rounded-md px-3 text-[10px] font-bold uppercase tracking-wide {fAuthoringMode === 'advance' ? 'bg-success text-background' : 'text-muted-foreground hover:bg-card'}"
									>
										Advance
									</button>
								</div>
								<span class="text-[11px] text-muted-foreground">
									<span class="font-semibold text-foreground">{questionTypeConfig.label}:</span> {questionTypeConfig.desc}
								</span>
							</div>
							<div class="grid gap-2 lg:grid-cols-[8rem_minmax(0,1fr)_5.5rem_8rem_6.5rem_10.5rem] lg:items-end">
								<div class="self-center">
									<h3 class="text-[10px] font-black uppercase tracking-[0.2em] text-foreground">Metadata</h3>
									<p class="mt-0.5 text-[10px] text-muted-foreground">Data wajib</p>
								</div>
								<div>
									<div class="mb-1 flex items-center justify-between gap-2">
										<label for="f-subject" class="block text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">
										Mata Pelajaran <span class="text-destructive">*</span>
									</label>
										{#if !readinessChecks.subject}
											<span class="text-[10px] font-semibold text-destructive">Wajib</span>
										{/if}
									</div>
									<select
										id="f-subject"
										bind:value={fSubjectId}
										class="h-8 w-full rounded-md border border-border bg-card px-2.5 text-sm font-medium text-foreground focus:outline-none focus:ring-2 focus:ring-ring"
									>
										<option value="">-- Pilih Mapel --</option>
										{#each subjects as s (s.id)}
											<option value={s.id}>{s.name}</option>
										{/each}
									</select>
								</div>
								<div>
									<label for="f-grade-level" class="mb-1 block text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">Tingkat</label>
									<Input id="f-grade-level" type="number" min="1" max="12" bind:value={fGradeLevel} class="h-8 text-sm font-medium" />
								</div>
								<div>
									<label for="f-difficulty" class="mb-1 block text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">
										Kesulitan
									</label>
									<select
										id="f-difficulty"
										bind:value={fDifficulty}
										class="h-8 w-full rounded-md border border-border bg-card px-2.5 text-sm font-medium text-foreground focus:outline-none focus:ring-2 focus:ring-ring"
									>
										{#each Object.entries(DIFFICULTY_LABEL) as [val, lbl] (val)}
											<option value={val}>{lbl}</option>
										{/each}
									</select>
								</div>
								<div>
									<label for="f-weight" class="mb-1 block text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">
										Bobot lokal (panduan paket)
									</label>
									<Input id="f-weight" type="number" min="1" bind:value={fWeight} class="h-8 text-sm font-medium" />
									<p class="mt-1 text-[10px] text-muted-foreground">Tidak disimpan backend; hanya catatan saat menyusun paket.</p>
								</div>
								<label for="f-rtl" class="flex h-8 cursor-pointer items-center justify-between gap-2 rounded-md border border-dashed border-success/20 bg-success/10 px-2.5">
									<span class="text-[10px] font-black uppercase tracking-wider text-success">Mode Arab / RTL</span>
									<input id="f-rtl" type="checkbox" bind:checked={fIsRtl} class="rounded accent-green-700" />
								</label>
							</div>
						</section>

						{#if isAdvanceMode}
							<section id="composer-advanced" class="scroll-mt-4 rounded-xl border border-border bg-card p-4 shadow-sm">
								<div class="mb-3 flex flex-wrap items-center justify-between gap-2">
									<div>
										<h3 class="text-xs font-black uppercase tracking-[0.2em] text-foreground">Detail Advance</h3>
										<p class="mt-0.5 text-xs text-muted-foreground">Blueprint, kurikulum, dan alur review.</p>
									</div>
									<label for="f-hots" class="flex h-8 cursor-pointer items-center gap-2 rounded-md border border-success/20 bg-success/10 px-2.5">
										<input id="f-hots" type="checkbox" bind:checked={fHotsFlag} class="rounded accent-green-700" />
										<span class="text-[10px] font-black uppercase tracking-wider text-success">HOTS</span>
									</label>
								</div>
								<div class="grid gap-3 md:grid-cols-3 xl:grid-cols-4">
									<div>
										<label for="f-phase" class="mb-1 block text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">Fase</label>
										<Input id="f-phase" placeholder="Fase D" bind:value={fAcademicPhase} class="h-8 text-sm" />
									</div>
									<div>
										<label for="f-topic" class="mb-1 block text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">Topik</label>
										<Input id="f-topic" placeholder="Topik materi" bind:value={fMaterialTopic} class="h-8 text-sm" />
									</div>
									<div>
										<label for="f-cognitive" class="mb-1 block text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">Kognitif</label>
										<Input id="f-cognitive" placeholder="C3 / HOTS" bind:value={fCognitiveLevel} class="h-8 text-sm" />
									</div>
									<div>
										<label for="f-cp" class="mb-1 block text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">CP</label>
										<Input id="f-cp" placeholder="CP ref" bind:value={fCPRef} class="h-8 text-sm" />
									</div>
									<div>
										<label for="f-tp" class="mb-1 block text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">TP</label>
										<Input id="f-tp" placeholder="TP ref" bind:value={fTPRef} class="h-8 text-sm" />
									</div>
									<div>
										<label for="f-kd" class="mb-1 block text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">KD</label>
										<Input id="f-kd" placeholder="KD ref" bind:value={fKDRef} class="h-8 text-sm" />
									</div>
									<div>
										<label for="f-indicator" class="mb-1 block text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">Indikator</label>
										<Input id="f-indicator" placeholder="Indikator" bind:value={fIndicatorRef} class="h-8 text-sm" />
									</div>
									<div class="rounded-md border border-border bg-muted/50 px-3 py-2">
										<p class="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">Alur</p>
										<p class="mt-0.5 text-[11px] text-muted-foreground">Draft dan review dikendalikan dari tombol bawah.</p>
									</div>
								</div>
							</section>

							<section id="composer-stimulus" class="scroll-mt-4 space-y-2 rounded-xl border border-border bg-card p-4 shadow-sm">
								<div class="flex flex-wrap items-center justify-between gap-2">
									<div>
										<h3 class="text-xs font-black uppercase tracking-[0.2em] text-foreground">Stimulus</h3>
										<p class="mt-0.5 text-xs text-muted-foreground">Narasi, data, gambar, atau konteks pendukung.</p>
									</div>
									<button type="button" onclick={() => (focusedEditor = 'stimulus')} class="rounded-md border border-border px-2 py-1 text-[10px] font-semibold uppercase tracking-wider text-muted-foreground hover:bg-muted/50">Fokus</button>
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
							<section id="composer-question" class="scroll-mt-4 space-y-2 rounded-xl border border-border bg-card p-4 shadow-sm">
							<div class="flex flex-wrap items-center justify-between gap-2">
								<div>
									<h3 class="text-xs font-black uppercase tracking-[0.2em] text-foreground">{isEssay ? 'Pertanyaan Essay' : isShortAnswer ? 'Pertanyaan Isian Singkat' : isMatching ? 'Instruksi Menjodohkan' : isTrueFalse ? 'Pernyataan Benar/Salah' : isAgreeDisagree ? 'Pernyataan Setuju/Tidak Setuju' : 'Isi Pertanyaan'}</h3>
									<p class="mt-0.5 text-xs text-muted-foreground">{questionTypeConfig.studentHint}</p>
								</div>
								<button type="button" onclick={() => (focusedEditor = 'stem')} class="rounded-md border border-border px-2 py-1 text-[10px] font-semibold uppercase tracking-wider text-muted-foreground hover:bg-muted/50">Fokus</button>
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
								<p class="text-[10px] font-semibold text-destructive">Isi pertanyaan minimal 5 karakter atau sisipkan gambar.</p>
							{/if}
						</section>

						{#if isMatching}
							<section id="composer-matching" class="scroll-mt-4 space-y-4 border-t border-border pt-5">
								<div class="flex flex-wrap items-center justify-between gap-3">
									<div>
										<h3 class="text-xs font-black uppercase italic tracking-[0.26em] text-muted-foreground">Pasangan Menjodohkan</h3>
										<p class="mt-1 text-xs text-muted-foreground">Kolom kiri adalah pernyataan/istilah. Kolom kanan adalah pasangan benar yang akan dipilih siswa.</p>
									</div>
									<div class="flex flex-wrap items-center gap-2">
										<span class="rounded-full bg-muted px-2 py-1 text-[10px] font-semibold text-muted-foreground">{fMatchingPairs.length} pasangan</span>
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
										<section class="rounded-xl border border-border bg-card p-3 shadow-sm">
											<div class="mb-2 flex items-center justify-between gap-3 border-b border-border pb-2">
												<div>
													<p class="text-xs font-black uppercase tracking-[0.2em] text-foreground">Pasangan {leftLabel} = {rightLabel}</p>
													<p class="mt-0.5 text-[11px] text-muted-foreground">Kunci disimpan otomatis sebagai {leftLabel}={rightLabel}</p>
												</div>
											</div>
											<div class="grid grid-cols-1 gap-3 lg:grid-cols-2">
												<div>
													<label for={`matching-left-${i}`} class="mb-1 block text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">Kolom Kiri {leftLabel}</label>
													<LegacyRichTextEditor
														bind:value={fMatchingPairs[i].left}
														id={`matching-left-${i}`}
														placeholder={`Istilah/pernyataan ${leftLabel}`}
														minRows={2}
														compact
														onImageUpload={uploadImageInEditor}
													/>
													{#if !richTextHasContent(pair.left)}
														<p class="mt-1 text-[10px] font-semibold text-destructive">Kolom kiri wajib diisi.</p>
													{/if}
												</div>
												<div>
													<label for={`matching-right-${i}`} class="mb-1 block text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">Kolom Kanan {rightLabel}</label>
													<LegacyRichTextEditor
														bind:value={fMatchingPairs[i].right}
														id={`matching-right-${i}`}
														placeholder={`Pasangan jawaban ${rightLabel}`}
														minRows={2}
														compact
														onImageUpload={uploadImageInEditor}
													/>
													{#if !richTextHasContent(pair.right)}
														<p class="mt-1 text-[10px] font-semibold text-destructive">Kolom kanan wajib diisi.</p>
													{/if}
												</div>
											</div>
										</section>
									{/each}
								</div>
								<div class="rounded-xl border border-dashed border-border bg-muted/50 p-3">
									<div class="flex flex-wrap items-center justify-between gap-3">
										<div>
											<p class="text-xs font-black uppercase tracking-[0.2em] text-foreground">Distraktor Kanan Opsional</p>
											<p class="mt-0.5 text-[11px] text-muted-foreground">Tambahkan pilihan kanan ekstra agar siswa tidak hanya mencocokkan satu-ke-satu.</p>
										</div>
										<Button type="button" variant="outline" size="sm" class="h-7 px-2 text-[10px]" disabled={fMatchingDistractors.length >= MAX_MATCHING_DISTRACTOR_COUNT} onclick={addMatchingDistractor}>
											+ Distraktor
										</Button>
									</div>
									{#if fMatchingDistractors.length > 0}
										<div class="mt-3 grid grid-cols-1 gap-3 lg:grid-cols-2">
											{#each fMatchingDistractors as distractor, i (`matching-distractor-${i}`)}
												<div class="rounded-lg border border-border bg-card p-3">
													<div class="mb-2 flex items-center justify-between gap-3">
														<label for={`matching-distractor-${i}`} class="block text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">Distraktor Kanan {fMatchingPairs.length + i + 1}</label>
														<button type="button" onclick={() => removeMatchingDistractor(i)} class="rounded-md border border-border px-2 py-1 text-[10px] font-semibold uppercase tracking-wider text-muted-foreground hover:bg-muted/50">
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
														<p class="mt-1 text-[10px] font-semibold text-warning">Isi distraktor atau hapus jika tidak dipakai.</p>
													{/if}
												</div>
											{/each}
										</div>
									{/if}
								</div>
							</section>
						{:else if hasOptionSection}
							<section id="composer-options" class="scroll-mt-4 space-y-4 border-t border-border pt-5">
								<div class="text-center">
									<h3 class="text-xs font-black uppercase italic tracking-[0.26em] text-muted-foreground">
										{isOrdering ? 'Opsi & Urutan Kunci' : isMultipleAnswer ? 'Opsi & Kunci Jawaban Ganda' : isTrueFalse ? 'Kunci Benar/Salah' : isAgreeDisagree ? 'Kunci Setuju/Tidak Setuju' : 'Opsi & Kunci Jawaban'}
									</h3>
									<p class="mt-1 text-xs text-muted-foreground">{questionTypeConfig.studentHint}</p>
								</div>
								{#if hasEditableOptions}
									<div class="flex flex-wrap items-center justify-center gap-2">
										<span class="rounded-full bg-muted px-2 py-1 text-[10px] font-semibold text-muted-foreground">{fOptions.length} opsi aktif</span>
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
											class="space-y-2 rounded-xl border bg-card p-3 shadow-sm transition-colors {isAnswer
												? 'border-success ring-4 ring-success/30'
												: 'border-border hover:border-border'}"
										>
											<div class="flex items-center justify-between gap-3 border-b border-border pb-2">
												<div>
													<p class="text-xs font-black uppercase tracking-[0.2em] text-foreground">Opsi {label}</p>
													<p class="mt-0.5 text-[11px] text-muted-foreground">{isAnswer ? (isOrdering ? `Urutan kunci: ${selectedAnswerLabels.indexOf(label) + 1}` : 'Ditandai sebagai kunci jawaban') : isFixedPair ? 'Pilihan tetap' : 'Pengecoh / alternatif jawaban'}</p>
												</div>
												<div class="flex shrink-0 items-center gap-1">
													{#if isOrdering}
														<button type="button" onclick={() => moveOrderingAnswerLabel(label, -1)} disabled={selectedAnswerLabels.indexOf(label) <= 0} class="rounded-md border border-border px-2 py-1 text-[10px] font-semibold uppercase tracking-wider text-muted-foreground hover:bg-muted/50 disabled:cursor-not-allowed disabled:opacity-40">Naik</button>
														<button type="button" onclick={() => moveOrderingAnswerLabel(label, 1)} disabled={selectedAnswerLabels.indexOf(label) >= selectedAnswerLabels.length - 1} class="rounded-md border border-border px-2 py-1 text-[10px] font-semibold uppercase tracking-wider text-muted-foreground hover:bg-muted/50 disabled:cursor-not-allowed disabled:opacity-40">Turun</button>
													{/if}
													{#if hasEditableOptions}
														<button type="button" onclick={() => (focusedEditor = label)} class="rounded-md border border-border px-2 py-1 text-[10px] font-semibold uppercase tracking-wider text-muted-foreground hover:bg-muted/50">Fokus</button>
													{/if}
													{#if isOrdering}
														<span class="rounded-lg bg-success px-2 py-1.5 text-[10px] font-black uppercase tracking-widest text-white">Urutan {selectedAnswerLabels.indexOf(label) + 1}</span>
													{:else}
														<label class="flex cursor-pointer items-center gap-2 rounded-lg px-2 py-1.5 transition-colors {isAnswer ? 'bg-success text-background' : 'hover:bg-success/10'}">
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
															<span class="text-[10px] font-black uppercase tracking-widest {isAnswer ? 'text-white' : 'text-muted-foreground'}">Kunci</span>
														</label>
													{/if}
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
														<p class="mt-1 text-[10px] font-semibold text-destructive">Opsi {label} wajib diisi.</p>
													{/if}
												{:else}
													<div class="rounded-md border border-border bg-muted/50 px-3 py-2 text-sm font-semibold text-foreground">
														{opt}
													</div>
												{/if}
											</div>
										</section>
									{/each}
								</div>
								{#if isMultipleAnswer && !answerKeyReady}
									<p class="text-center text-[10px] font-semibold text-destructive">Pilih minimal dua opsi sebagai kunci jawaban ganda.</p>
								{/if}
							</section>
						{:else if isShortAnswer}
							<section id="composer-answer" class="scroll-mt-4 space-y-3 rounded-xl border border-border bg-card p-4 shadow-sm">
								<div>
									<h3 class="text-xs font-black uppercase tracking-[0.2em] text-foreground">Kunci Isian Singkat</h3>
									<p class="mt-0.5 text-xs text-muted-foreground">Pisahkan beberapa jawaban diterima dengan tanda |. Sistem mengabaikan besar/kecil huruf dan spasi ganda.</p>
								</div>
								<div>
									<label for="f-short-answer-key" class="mb-1 block text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">
										Kunci / Alias Jawaban <span class="text-destructive">*</span>
									</label>
									<Input
										id="f-short-answer-key"
										bind:value={fAnswerKey}
										placeholder="Contoh: Fotosintesis | foto sintesis"
										class="h-9 text-sm font-medium"
									/>
									{#if shortAnswerAliases.length > 0}
										<p class="mt-1 text-[10px] font-semibold text-success">{shortAnswerAliases.length} jawaban diterima: {shortAnswerAliases.join(' / ')}</p>
									{/if}
									{#if !readinessChecks.answerKey}
										<p class="mt-1 text-[10px] font-semibold text-destructive">Kunci isian singkat wajib diisi sebelum review.</p>
									{/if}
								</div>
							</section>
						{:else}
						<section id="composer-rubric" class="scroll-mt-4 space-y-2 rounded-xl border border-border bg-card p-4 shadow-sm">
							<div class="flex flex-wrap items-center justify-between gap-2">
								<div>
									<h3 class="text-xs font-black uppercase tracking-[0.2em] text-foreground">{isAdvanceMode ? 'Rubrik Penilaian' : 'Pedoman Jawaban'}</h3>
									<p class="mt-0.5 text-xs text-muted-foreground">{isAdvanceMode ? 'Kriteria koreksi, rentang skor, dan catatan penilai.' : 'Panduan singkat agar guru mudah mengoreksi jawaban.'}</p>
								</div>
								<button type="button" onclick={() => (focusedEditor = 'rubric')} class="rounded-md border border-border px-2 py-1 text-[10px] font-semibold uppercase tracking-wider text-muted-foreground hover:bg-muted/50">Fokus</button>
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
								<p class="text-[10px] font-semibold text-destructive">Pedoman/rubrik essay wajib diisi.</p>
							{/if}
						</section>
						{/if}

							{#if isAdvanceMode}
								<section id="composer-explanation" class="scroll-mt-4 space-y-2 rounded-xl border border-border bg-card p-4 shadow-sm">
								<div class="flex flex-wrap items-center justify-between gap-2">
									<div>
										<h3 class="text-xs font-black uppercase tracking-[0.2em] text-foreground">{isEssay ? 'Catatan Pembahasan' : 'Pembahasan'}</h3>
										<p class="mt-0.5 text-xs text-muted-foreground">{isEssay ? 'Catatan internal untuk guru/reviewer.' : 'Pembahasan yang membantu review dan bank soal.'}</p>
									</div>
									<button type="button" onclick={() => (focusedEditor = 'explanation')} class="rounded-md border border-border px-2 py-1 text-[10px] font-semibold uppercase tracking-wider text-muted-foreground hover:bg-muted/50">Fokus</button>
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
						<aside class={`space-y-4 min-w-0 xl:sticky xl:top-0 xl:self-start ${composerMobilePanel === 'write' ? 'hidden lg:block' : 'block'}`}>
						<div class="rounded-xl border border-border bg-card p-4 shadow-sm">
							{@render composerPreview()}
						</div>
					</aside>
				</div>
			</div>

				<div class="flex shrink-0 flex-col gap-2 border-t border-success/20 bg-card px-4 py-2.5 md:flex-row md:items-center md:justify-between md:px-6">
					<div class="min-w-0 text-xs text-muted-foreground">
						<span class="font-semibold text-success">{draftStatusLabel()}</span>
						<span class="ml-2 text-muted-foreground">· {draftIssues.length === 0 ? 'Draft bisa disimpan' : draftIssues[0]} · Review {validationIssues.length === 0 ? 'siap' : `${validationIssues.length} wajib belum lengkap`} · Ctrl+S</span>
						{#if firstComposerIssue}
							<button type="button" class="ml-2 text-destructive underline decoration-red-200 underline-offset-2" onclick={scrollToFirstComposerIssue}>
								Lengkapi: {firstComposerIssue.message}
							</button>
						{/if}
					</div>
				<div class="flex shrink-0 flex-wrap justify-end gap-2">
					<Button
						variant="outline"
						class="h-8 text-xs"
						onclick={() => void discardLocalDraftAndClose()}
					>
						Hapus Draft Lokal & Tutup
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
						class="h-8 bg-success text-xs text-background hover:bg-success disabled:opacity-50"
					>
						Ajukan Review
					</LoadingButton>
				</div>
			</div>

			{#if focusedEditor}
				<div class="absolute inset-0 z-20 flex bg-foreground/35 p-3 md:p-6">
					<div class="flex min-h-0 w-full flex-col overflow-hidden rounded-xl bg-card shadow-2xl">
						<div class="flex shrink-0 items-center justify-between gap-3 border-b border-success/20 px-4 py-3">
							<div>
								<p class="text-[10px] font-black uppercase tracking-[0.22em] text-success">Mode Fokus</p>
								<h3 class="text-base font-black uppercase italic text-foreground">{focusTitle(focusedEditor)}</h3>
							</div>
							<Button type="button" variant="outline" size="sm" class="h-8 text-xs" onclick={() => (focusedEditor = null)}>
								Tutup Fokus
							</Button>
						</div>
						<div class="min-h-0 flex-1 overflow-y-auto bg-muted/50 p-4">
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
	</section>
{/snippet}

<style>
	:global(.latex-preview .latex-display) {
		overflow-x: auto;
		padding: 0.25rem 0;
	}
	:global(.latex-preview .katex-display) {
		margin: 0.5rem 0;
	}
</style>
