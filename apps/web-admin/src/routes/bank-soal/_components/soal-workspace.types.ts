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

export type QuestionVersion = {
	id: string;
	code?: string;
	workflow_status?: string;
	status?: string;
	version_number?: number;
	is_latest_version?: boolean;
	source_question_id?: string | null;
	supersedes_question_id?: string | null;
	version_note?: string;
	created_at?: string;
	updated_at?: string;
	author_username?: string;
	reviewer_username?: string;
	approver_username?: string;
};

export type WorkspaceRoute = '/bank-soal' | '/bank-soal/tambah' | '/bank-soal/impor' | '/bank-soal/verifikasi';
export type AuthoringMode = 'beginner' | 'advance';
export type ComposerIssueHint = { message: string; targetId: string };
export type ComposerStageCard = { label: string; desc: string; status: string; tone: 'green' | 'amber' | 'red' | 'slate'; targetId: string };
export type RevisionSourceFilter = '' | 'item_analysis' | 'reviewer' | 'workflow';
export type ReviewDecision = 'mark_reviewed' | 'request_revision' | 'reject';
export type BulkWorkflowAction = 'mark_reviewed' | 'request_revision' | 'reject' | 'approve' | 'publish' | 'archive';
export type ComposerQuestionType = 'multiple_choice' | 'multiple_answer' | 'true_false' | 'agree_disagree' | 'matching' | 'ordering' | 'short_answer' | 'essay';
export type ComposerSaveIntent = 'draft' | 'review';
export type ComposerAction = ComposerSaveIntent | 'draft_next' | 'review_next';

export type ComposerPostSaveAction = {
	questionId: string;
	intent: ComposerSaveIntent;
	metadata: ComposerMetadataMemory;
	wasEdit: boolean;
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
	target_level?: string | null;
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
	author_display_name?: string;
	reviewer_username?: string;
	reviewer_display_name?: string;
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
	version_group_id?: string | null;
	version_number?: number;
	source_question_id?: string | null;
	supersedes_question_id?: string | null;
	is_latest_version?: boolean;
	version_note?: string;
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
	limit: number;
};

export type PageData = {
	user?: {
		username?: string;
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
	targetLevel: string;
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
	target_level: string;
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

export type ServerBaselineHtml = {
	stem: string;
	stimulus: string;
	explanation: string;
	rubric: string;
	options: Record<string, string>;
};

export type GuardedHtmlField = { label: string; baseline: string; draft: string };
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
	from_status?: string;
	to_status?: string;
	note?: string;
	notes?: string;
	metadata?: Record<string, unknown>;
	actor_username?: string;
	actor_display_name?: string;
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
