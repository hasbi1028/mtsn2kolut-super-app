import { htmlToPlainText } from '$lib/utils/html-text';

export type Subject = { id: string; name: string; code: string };
export type CbtEvent = {
	id: string;
	title: string;
	exam_type?: string;
	status?: string;
	academic_year_name?: string;
};
export type UserOption = {
	id: string;
	username: string;
	roles?: string[];
	employee_id?: string | null;
	profile_nama?: string | null;
};
export type EventMemberRole = 'panitia' | 'pembuat_soal' | 'reviewer' | 'proktor' | 'pengawas' | 'korektor';
export type EventMember = {
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
export type OptionItem = {
	label: string;
	text?: string;
	html?: string;
	latex?: string;
	match_label?: string;
	match_text?: string;
	match_html?: string;
	is_distractor?: boolean;
};
export type MatchingPair = { left: string; right: string };
export type ModuleMode = 'catalog' | 'composer' | 'review' | 'import';
export type WorkspaceRouteMode = 'composer' | 'import';
export type WorkspaceRoute = '/bank-soal' | '/bank-soal/tambah' | '/bank-soal/impor' | '/bank-soal/verifikasi';
export type AuthoringMode = 'beginner' | 'advance';
export type ComposerIssueHint = { message: string; targetId: string };
export type ComposerStageCard = {
	label: string;
	desc: string;
	status: string;
	tone: 'green' | 'amber' | 'red' | 'slate';
	targetId: string;
};
export type RevisionSourceFilter = '' | 'item_analysis' | 'reviewer' | 'workflow';
export type ReviewDecision = 'approve' | 'reject';
export type BulkWorkflowAction = 'approve' | 'reject' | 'publish';
export type ComposerQuestionType =
	| 'multiple_choice'
	| 'multiple_answer'
	| 'true_false'
	| 'agree_disagree'
	| 'matching'
	| 'ordering'
	| 'short_answer'
	| 'essay';
export type ComposerSaveIntent = 'draft' | 'review';
export type AnswerMode =
	| 'single_option'
	| 'multi_option'
	| 'fixed_pair'
	| 'matching'
	| 'ordering'
	| 'short_text'
	| 'rubric';
export type QuestionTypeConfig = {
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
export type Question = {
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
export type QuestionListResponse = {
	items: Question[];
	meta?: { total: number; limit: number; offset: number };
};
export type SoalOverview = {
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
export type PageData = {
	user?: {
		role?: string;
		roles?: string[];
		permissions?: string[];
	};
};
export type AcademicPayload = {
	subjects?: Subject[];
	error?: string;
	message?: string;
};
export type EventsPayload = CbtEvent[] | { items?: CbtEvent[]; events?: CbtEvent[] };
export type UsersPayload = UserOption[] | { items?: UserOption[]; users?: UserOption[] };
export type ComposerMetadataMemory = {
	eventId?: string;
	specialEventMode?: boolean;
	subjectId: string;
	questionType: ComposerQuestionType;
	authoringMode: AuthoringMode;
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
	savedAt: string;
};
export type DraftPayload = ComposerMetadataMemory & {
	stem: string;
	stimulus: string;
	rubric: string;
	explanation: string;
	options: string[];
	matchingPairs: MatchingPair[];
	matchingDistractors: string[];
	answerKey: string;
	workflowStatus: string;
};
export type QuestionPayloadOption = {
	label: string;
	text: string;
	html: string;
	match_label?: string;
	match_text?: string;
	match_html?: string;
	is_distractor?: boolean;
};
export type QuestionSavePayload = {
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
export type OptionLabel = 'A' | 'B' | 'C' | 'D' | 'E' | 'F';
export type LegacyImportResult = {
	total_rows: number;
	valid?: number;
	would_import?: number;
	imported: number;
	skipped: number;
	errors: string[];
	duplicate_codes: string[];
};
export type QuestionTarget = {
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
export type TimelineItem = {
	id?: string;
	action?: string;
	status?: string;
	notes?: string;
	actor_username?: string;
	created_at?: string;
};
export type BulkWorkflowResult = {
	question_id?: string;
	id?: string;
	ok?: boolean;
	success?: boolean;
	status?: string;
	error?: string;
	message?: string;
};
export type BulkWorkflowResponse = { results?: BulkWorkflowResult[] } | BulkWorkflowResult[];
export type FocusedEditor = 'stem' | 'stimulus' | 'rubric' | 'explanation' | OptionLabel;

export const PAGE_SIZE = 15;
export const MIN_OPTION_COUNT = 4;
export const MAX_OPTION_COUNT = 6;
export const MIN_MATCHING_PAIR_COUNT = 2;
export const DEFAULT_MATCHING_PAIR_COUNT = 4;
export const MAX_MATCHING_PAIR_COUNT = 6;
export const MAX_MATCHING_DISTRACTOR_COUNT = 4;
export const ANSWER_LABELS: OptionLabel[] = ['A', 'B', 'C', 'D', 'E', 'F'];
export const TRUE_FALSE_OPTIONS = ['Benar', 'Salah'];
export const AGREE_DISAGREE_OPTIONS = ['Setuju', 'Tidak Setuju'];
export const QUESTION_TYPE_CONFIGS: QuestionTypeConfig[] = [
	{
		id: 'multiple_choice',
		label: 'Pilihan Ganda',
		shortLabel: 'PG',
		desc: 'Satu jawaban benar dari beberapa opsi.',
		studentHint: 'Siswa memilih satu jawaban.',
		answerMode: 'single_option',
		minOptions: MIN_OPTION_COUNT,
		maxOptions: MAX_OPTION_COUNT
	},
	{
		id: 'multiple_answer',
		label: 'Jawaban Ganda',
		shortLabel: 'Ganda',
		desc: 'Lebih dari satu opsi dapat menjadi kunci.',
		studentHint: 'Siswa memilih semua jawaban benar.',
		answerMode: 'multi_option',
		minOptions: MIN_OPTION_COUNT,
		maxOptions: MAX_OPTION_COUNT
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
		fixedOptions: TRUE_FALSE_OPTIONS
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
		fixedOptions: AGREE_DISAGREE_OPTIONS
	},
	{
		id: 'matching',
		label: 'Menjodohkan',
		shortLabel: 'Jodoh',
		desc: 'Pasangkan pernyataan kiri dengan jawaban kanan.',
		studentHint: 'Siswa memilih pasangan yang sesuai untuk setiap baris.',
		answerMode: 'matching',
		minOptions: MIN_MATCHING_PAIR_COUNT,
		maxOptions: MAX_MATCHING_PAIR_COUNT
	},
	{
		id: 'ordering',
		label: 'Mengurutkan',
		shortLabel: 'Urutan',
		desc: 'Siswa menyusun opsi sesuai urutan kunci.',
		studentHint: 'Siswa mengurutkan semua opsi jawaban.',
		answerMode: 'ordering',
		minOptions: MIN_OPTION_COUNT,
		maxOptions: MAX_OPTION_COUNT
	},
	{
		id: 'short_answer',
		label: 'Isian Singkat',
		shortLabel: 'Isian',
		desc: 'Jawaban pendek dengan kunci teks.',
		studentHint: 'Siswa mengetik jawaban singkat.',
		answerMode: 'short_text',
		minOptions: 0,
		maxOptions: 0
	},
	{
		id: 'essay',
		label: 'Essay',
		shortLabel: 'Essay',
		desc: 'Jawaban uraian dikoreksi manual dengan pedoman.',
		studentHint: 'Siswa menulis jawaban uraian.',
		answerMode: 'rubric',
		minOptions: 0,
		maxOptions: 0
	}
];
export const WORKFLOW_LABEL: Record<string, string> = {
	draft: 'Draft',
	review: 'Menunggu Review',
	approved: 'Disetujui',
	rejected: 'Perlu Revisi'
};
export const DIFFICULTY_LABEL: Record<string, string> = { easy: 'Mudah', medium: 'Sedang', hard: 'Sulit' };
export const revisionSourceOptions: Array<{ id: RevisionSourceFilter; label: string; desc: string }> = [
	{ id: '', label: 'Semua Revisi', desc: 'Semua sumber' },
	{ id: 'item_analysis', label: 'Analisis Butir', desc: 'Dari hasil ujian' },
	{ id: 'reviewer', label: 'Reviewer', desc: 'Catatan penelaah' },
	{ id: 'workflow', label: 'Workflow', desc: 'Tanpa reviewer' }
];

export const DRAFT_KEY = (id: string | null, eventId: string) => `mtsn2-soal-komposer:${eventId || 'global'}:${id ?? 'new'}`;
export const LAST_METADATA_KEY = (eventId: string) => `mtsn2-soal-komposer:last-metadata:${eventId || 'global'}`;
export const EVENT_MEMBER_ROLES: Array<{ value: EventMemberRole; label: string; desc: string }> = [
	{ value: 'panitia', label: 'Panitia', desc: 'Koordinasi kegiatan' },
	{ value: 'pembuat_soal', label: 'Pembuat Soal', desc: 'Menyusun bank soal' },
	{ value: 'reviewer', label: 'Reviewer', desc: 'Menelaah mutu soal' },
	{ value: 'proktor', label: 'Proktor', desc: 'Teknis sesi ujian' },
	{ value: 'pengawas', label: 'Pengawas', desc: 'Pengawasan ruang' },
	{ value: 'korektor', label: 'Korektor', desc: 'Koreksi uraian' }
];

export function richTextHasContent(html: string): boolean {
	return htmlToPlainText(html).length > 0 || html.includes('<img');
}

export function createEmptyMatchingPairs(count = DEFAULT_MATCHING_PAIR_COUNT): MatchingPair[] {
	return Array.from({ length: count }, () => ({ left: '', right: '' }));
}

export function isComposerQuestionType(value: string | undefined): value is ComposerQuestionType {
	return QUESTION_TYPE_CONFIGS.some((config) => config.id === value);
}

export function getQuestionTypeConfig(type: ComposerQuestionType): QuestionTypeConfig {
	return QUESTION_TYPE_CONFIGS.find((config) => config.id === type) ?? QUESTION_TYPE_CONFIGS[0]!;
}

export function normalizeQuestionType(value: string | undefined): ComposerQuestionType {
	const normalized = (value ?? '').trim().toLowerCase();
	return isComposerQuestionType(normalized) ? normalized : 'multiple_choice';
}

export function questionTypeLabel(value: string | undefined): string {
	return getQuestionTypeConfig(normalizeQuestionType(value)).shortLabel;
}

export function normalizeAuthoringMode(value: string | undefined): AuthoringMode {
	return value === 'advance' ? 'advance' : 'beginner';
}

export function normalizeWorkflowStatus(value: string | undefined): string {
	return value === 'review' ? 'review' : 'draft';
}

export function defaultAnswerKeyForQuestionType(type: ComposerQuestionType): string {
	const config = getQuestionTypeConfig(type);
	if (config.answerMode === 'single_option' || config.answerMode === 'fixed_pair') return 'A';
	if (config.answerMode === 'matching') return buildMatchingAnswerKey(DEFAULT_MATCHING_PAIR_COUNT);
	if (config.answerMode === 'ordering') return normalizeOrderingAnswerKey('', config.minOptions);
	return '';
}

export function defaultOptionsForQuestionType(type: ComposerQuestionType): string[] {
	const config = getQuestionTypeConfig(type);
	if (config.answerMode === 'fixed_pair') return [...(config.fixedOptions ?? [])];
	if (config.answerMode !== 'single_option' && config.answerMode !== 'multi_option' && config.answerMode !== 'ordering') return [];
	return Array.from({ length: config.minOptions }, () => '');
}

export function normalizeOptionCount(options: string[], type: ComposerQuestionType): string[] {
	const config = getQuestionTypeConfig(type);
	if (config.answerMode === 'fixed_pair') return [...(config.fixedOptions ?? [])];
	if (config.answerMode !== 'single_option' && config.answerMode !== 'multi_option' && config.answerMode !== 'ordering') return [];
	let normalized = options.slice(0, config.maxOptions);
	while (normalized.length < config.minOptions) normalized = [...normalized, ''];
	return normalized;
}

export function normalizeMatchingPairs(pairs: MatchingPair[], type: ComposerQuestionType): MatchingPair[] {
	const config = getQuestionTypeConfig(type);
	if (config.answerMode !== 'matching') return createEmptyMatchingPairs();
	let normalized = pairs
		.slice(0, config.maxOptions)
		.map((pair) => ({ left: pair.left ?? '', right: pair.right ?? '' }));
	const minCount = pairs.length === 0 ? DEFAULT_MATCHING_PAIR_COUNT : config.minOptions;
	while (normalized.length < minCount) normalized = [...normalized, { left: '', right: '' }];
	return normalized;
}

export function optionsToMatchingPairs(options: OptionItem[]): MatchingPair[] {
	return options.filter((option) => !option.is_distractor).map((option) => ({
		left: option.html || option.text || option.latex || '',
		right: option.match_html || option.match_text || ''
	}));
}

export function normalizeMatchingDistractors(distractors: string[], type: ComposerQuestionType): string[] {
	const config = getQuestionTypeConfig(type);
	if (config.answerMode !== 'matching') return [];
	return distractors.slice(0, MAX_MATCHING_DISTRACTOR_COUNT);
}

export function optionsToMatchingDistractors(options: OptionItem[]): string[] {
	return options
		.filter((option) => option.is_distractor)
		.map((option) => option.match_html || option.match_text || '')
		.filter((option) => option.trim().length > 0);
}

export function optionLabelAt(index: number): OptionLabel {
	return ANSWER_LABELS[index] ?? 'A';
}

export function answerKeyLabels(value: string | undefined, labels: OptionLabel[]): OptionLabel[] {
	const keys: OptionLabel[] = [];
	for (const raw of (value ?? '').split(',')) {
		const label = raw.trim().toUpperCase() as OptionLabel;
		if (!labels.includes(label) || keys.includes(label)) continue;
		keys.push(label);
	}
	return keys.sort((a, b) => ANSWER_LABELS.indexOf(a) - ANSWER_LABELS.indexOf(b));
}

export function orderingAnswerKeyLabels(value: string | undefined, labels: OptionLabel[]): OptionLabel[] {
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

export function normalizeOrderingAnswerKey(value: string | undefined, optionCount: number): string {
	return orderingAnswerKeyLabels(value, ANSWER_LABELS.slice(0, optionCount)).join(',');
}

export function buildMatchingAnswerKey(count: number): string {
	return Array.from({ length: count }, (_, index) => `${optionLabelAt(index)}=${index + 1}`).join(';');
}

export function normalizeMatchingAnswerKey(value: string | undefined, count: number): string {
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

export function normalizeShortAnswerComparable(value: string): string {
	return value.replace(/\u00a0/g, ' ').trim().replace(/\s+/g, ' ').toLocaleLowerCase('id-ID');
}

export function parseShortAnswerAliases(value: string | undefined): string[] {
	const aliases: string[] = [];
	for (const raw of (value ?? '').split('|')) {
		const alias = raw.replace(/\u00a0/g, ' ').trim().replace(/\s+/g, ' ');
		const comparable = normalizeShortAnswerComparable(alias);
		if (!comparable || aliases.some((item) => normalizeShortAnswerComparable(item) === comparable)) continue;
		aliases.push(alias);
	}
	return aliases;
}

export function normalizeAnswerKey(value: string | undefined, type: ComposerQuestionType, optionCount: number): string {
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

export function editorAllowedForQuestionType(editor: FocusedEditor, type: ComposerQuestionType): boolean {
	if (editor === 'stem' || editor === 'stimulus' || editor === 'explanation') return true;
	if (editor === 'rubric') return getQuestionTypeConfig(type).answerMode === 'rubric';
	const answerMode = getQuestionTypeConfig(type).answerMode;
	return answerMode === 'single_option' || answerMode === 'multi_option' || answerMode === 'ordering';
}

export function questionUsageLocked(q: Question): boolean {
	const packages = q.package_count ?? q.usage?.package_count ?? 0;
	const answers = q.answer_count ?? q.usage?.answer_count ?? 0;
	return Boolean(q.is_locked ?? q.usage?.is_locked ?? (packages > 0 || answers > 0));
}

export function questionUsageText(q: Question): string {
	const packages = q.package_count ?? q.usage?.package_count ?? 0;
	const answers = q.answer_count ?? q.usage?.answer_count ?? 0;
	if (packages > 0 && answers > 0) return `${packages} paket, ${answers} jawaban`;
	if (packages > 0) return `${packages} paket`;
	if (answers > 0) return `${answers} jawaban`;
	return 'Belum dipakai';
}

export function revisionSourceLabel(q: Question): string {
	const note = (q.review_notes ?? '').toLowerCase();
	if (note.includes('analisis butir')) return 'Analisis Butir';
	if ((q.reviewer_username ?? '').trim()) return `Reviewer: ${q.reviewer_username}`;
	return 'Alur Verifikasi';
}

export function revisionReason(q: Question): string {
	const note = (q.review_notes ?? '').replace(/\s+/g, ' ').trim();
	if (!note) return 'Belum ada catatan alasan revisi.';
	return note.length > 180 ? `${note.slice(0, 180)}...` : note;
}

export function canSubmitRevisionReview(q: Question): boolean {
	return q.workflow_status === 'rejected' && q.status === 'draft' && !questionUsageLocked(q);
}

export function isQuickEditable(q: Question): boolean {
	return isComposerQuestionType(q.question_type)
		&& (q.workflow_status === 'draft' || q.workflow_status === 'rejected')
		&& q.status === 'draft'
		&& !questionUsageLocked(q);
}

export function explainQuickEditBlocked(q: Question): string {
	if (questionUsageLocked(q)) return 'Soal sudah dipakai. Gunakan Duplikat untuk membuat revisi draft.';
	if ((q.workflow_status !== 'draft' && q.workflow_status !== 'rejected') || q.status !== 'draft') {
		return 'Soal sudah masuk alur review/publikasi. Gunakan Duplikat untuk revisi.';
	}
	if (!isComposerQuestionType(q.question_type)) {
		return 'Tipe soal ini belum masuk komposer utama. Gunakan Duplikat setelah tipe ini dimigrasikan.';
	}
	return 'Buat revisi lewat Duplikat agar riwayat soal tetap aman.';
}

export function stemPreview(q: Question): string {
	const text = htmlToPlainText(q.stem_html || q.question_text || '');
	return text.length > 90 ? `${text.slice(0, 90)}...` : text || '(kosong)';
}

export function workflowClass(status: string): string {
	const map: Record<string, string> = {
		draft: 'bg-muted text-muted-foreground',
		review: 'bg-warning/15 text-warning',
		approved: 'bg-success/15 text-success',
		rejected: 'bg-destructive/15 text-destructive'
	};
	return map[status] ?? 'bg-muted text-muted-foreground';
}

export function answerKeyLabelForQuestion(q: Question): string {
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

export function optionPrimaryContent(option: OptionItem): string {
	if (option.is_distractor) return option.match_html || option.match_text || '-';
	return option.html || option.text || option.latex || '-';
}

export function optionSecondaryContent(option: OptionItem): string {
	if (option.is_distractor) return '';
	return option.match_html || option.match_text || '';
}

export function composerStageToneClass(tone: ComposerStageCard['tone']): string {
	const map: Record<ComposerStageCard['tone'], string> = {
		green: 'border-primary/20 bg-primary/10 text-primary',
		amber: 'border-warning/30 bg-warning/10 text-warning',
		red: 'border-destructive/30 bg-destructive/10 text-destructive',
		slate: 'border-border bg-muted/50 text-foreground'
	};
	return map[tone];
}
