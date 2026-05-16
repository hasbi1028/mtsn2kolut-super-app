import {
	canCreateBankSoal,
	canImportBankSoal,
	canManageBankSoalSettings,
	canReviewBankSoal,
	type BankSoalAccessUser,
} from './access';

export type BankSoalHealthEvidence = 'summary' | 'sample' | 'missing';

export type BankSoalHealthQuestion = {
	id?: string | null;
	subject_id?: string | null;
	subject_name?: string | null;
	subject_code?: string | null;
	status?: string | null;
	workflow_status?: string | null;
	kd_ref?: string | null;
	cp_ref?: string | null;
	tp_ref?: string | null;
	cognitive_level?: string | null;
	difficulty?: string | null;
	explanation_html?: string | null;
	answer_explanation?: string | null;
	discussion_html?: string | null;
	solution_html?: string | null;
	explanation?: string | null;
	package_count?: number | null;
	asset_count?: number | null;
	assets?: unknown[] | null;
	stem_media_url?: string | null;
	stimulus_media_url?: string | null;
	stem_audio_url?: string | null;
	stimulus_audio_url?: string | null;
	import_batch_id?: string | null;
	import_source?: string | null;
	source_file?: string | null;
	source_filename?: string | null;
	imported_at?: string | null;
	[key: string]: unknown;
};

export type BankSoalHealthSummary = {
	counts?: Partial<Record<string, number>>;
	by_subject?: Array<{
		subject_id?: string | null;
		subject_name?: string | null;
		subject_code?: string | null;
		total?: number | null;
		published?: number | null;
	}>;
	by_cognitive_level?: Array<{ cognitive_level?: string | null; total?: number | null }>;
	recent?: BankSoalHealthQuestion[];
};

export type BankSoalHealthInput = {
	summary?: BankSoalHealthSummary | null;
	questions?: BankSoalHealthQuestion[] | null;
	sampleLimit?: number;
};

export type BankSoalStatusKey = 'total' | 'published' | 'approved' | 'review' | 'draft' | 'revision' | 'archived';

export type BankSoalStatusCard = {
	key: BankSoalStatusKey;
	label: string;
	value: number | null;
	evidence: BankSoalHealthEvidence;
	evidenceLabel: string;
	helper: string;
	tone: 'neutral' | 'success' | 'warning' | 'danger';
};

export type BankSoalRatioMetric = {
	key: string;
	label: string;
	value: number | null;
	total: number | null;
	missing: number | null;
	percent: number | null;
	evidence: BankSoalHealthEvidence;
	evidenceLabel: string;
};

export type BankSoalReadiness = {
	score: number | null;
	grade: 'A' | 'B' | 'C' | 'D' | 'Perlu evidence/data';
	evidenceLabel: string;
	drivers: string[];
};

export type BankSoalHealthWarning = {
	kind: 'review' | 'import' | 'asset' | 'governance';
	label: string;
	message: string;
	severity: 'info' | 'warning' | 'danger';
	evidenceLabel: string;
};

export type BankSoalHealthActionCapability = 'read' | 'create' | 'review' | 'import' | 'settings' | 'update';

export type BankSoalHealthAction = {
	label: string;
	description: string;
	href: string;
	capability: BankSoalHealthActionCapability;
	priority: number;
};

export type BankSoalRoleWorkflowCard = {
	key: string;
	label: string;
	value: number | null;
	evidenceLabel: string;
	helper: string;
	tone: 'neutral' | 'success' | 'warning' | 'danger';
};

export type BankSoalHealthModel = {
	statusCards: BankSoalStatusCard[];
	roleWorkflowCards: BankSoalRoleWorkflowCard[];
	subjectCoverage: BankSoalRatioMetric;
	curriculumCoverage: BankSoalRatioMetric[];
	metadataQuality: BankSoalRatioMetric[];
	readiness: BankSoalReadiness;
	reviewBacklog: BankSoalRatioMetric;
	warnings: BankSoalHealthWarning[];
	actions: BankSoalHealthAction[];
	sampleSize: number;
	sampleLimit: number;
	totalQuestions: number | null;
};

const statusLabels: Record<BankSoalStatusKey, string> = {
	total: 'Total',
	published: 'Terbit',
	approved: 'Disetujui',
	review: 'Review',
	draft: 'Draft',
	revision: 'Revisi',
	archived: 'Arsip',
};

const statusHelpers: Record<BankSoalStatusKey, string> = {
	total: 'seluruh stok yang tersedia dari summary atau sampel',
	published: 'sudah dapat dipakai di paket asesmen',
	approved: 'lulus review dan menunggu publikasi',
	review: 'menunggu keputusan reviewer',
	draft: 'masih disusun atau belum dikirim',
	revision: 'dikembalikan untuk perbaikan',
	archived: 'hanya ditampilkan saat evidence tersedia',
};

function numberValue(value: unknown): number | null {
	return typeof value === 'number' && Number.isFinite(value) ? value : null;
}

function normalizeText(value: unknown): string {
	return typeof value === 'string' ? value.trim() : '';
}

function isFilled(value: unknown): boolean {
	if (typeof value === 'string') return value.trim().length > 0;
	if (typeof value === 'number') return Number.isFinite(value) && value > 0;
	if (Array.isArray(value)) return value.length > 0;
	return value !== null && value !== undefined && value !== false;
}

function hasOwn(item: BankSoalHealthQuestion, key: string): boolean {
	return Object.prototype.hasOwnProperty.call(item, key);
}

function hasAnyField(items: BankSoalHealthQuestion[], keys: string[]): boolean {
	return items.some((item) => keys.some((key) => hasOwn(item, key)));
}

function hasAnyFilledField(item: BankSoalHealthQuestion, keys: string[]): boolean {
	return keys.some((key) => isFilled(item[key]));
}

function evidenceLabel(evidence: BankSoalHealthEvidence, sampleSize: number): string {
	if (evidence === 'summary') return 'summary';
	if (evidence === 'sample') return `sampel ${sampleSize} soal`;
	return 'perlu evidence/data';
}

function countFromSummary(summary: BankSoalHealthSummary | null | undefined, key: BankSoalStatusKey): number | null {
	const counts = summary?.counts ?? {};
	if (key === 'total') return numberValue(counts.total) ?? numberValue(counts.all);
	if (key === 'revision') return numberValue(counts.revision) ?? numberValue(counts.rejected);
	return numberValue(counts[key]);
}

function workflowOf(question: BankSoalHealthQuestion): string {
	return normalizeText(question.workflow_status).toLowerCase();
}

function publicationOf(question: BankSoalHealthQuestion): string {
	return normalizeText(question.status).toLowerCase();
}

function countFromSample(questions: BankSoalHealthQuestion[], key: BankSoalStatusKey): number | null {
	if (questions.length === 0) return null;
	if (key === 'total') return questions.length;
	if (key === 'published') return questions.filter((question) => publicationOf(question) === 'published').length;
	if (key === 'approved') return questions.filter((question) => workflowOf(question) === 'approved').length;
	if (key === 'review') return questions.filter((question) => workflowOf(question) === 'review').length;
	if (key === 'revision') {
		return questions.filter((question) => ['revision', 'rejected'].includes(workflowOf(question))).length;
	}
	if (key === 'draft') {
		return questions.filter((question) => {
			const workflow = workflowOf(question);
			const status = publicationOf(question);
			return workflow === 'draft' || (!workflow && status !== 'published' && status !== 'archived');
		}).length;
	}
	const archived = questions.filter((question) => publicationOf(question) === 'archived' || workflowOf(question) === 'archived').length;
	return archived > 0 ? archived : null;
}

function statusTone(key: BankSoalStatusKey, value: number | null): BankSoalStatusCard['tone'] {
	if (key === 'published' || key === 'approved') return 'success';
	if (key === 'review' || key === 'revision') return value && value > 0 ? 'warning' : 'neutral';
	if (key === 'archived') return value && value > 0 ? 'danger' : 'neutral';
	return 'neutral';
}

function buildRoleWorkflowCards(summary: BankSoalHealthSummary | null | undefined, sampleSize: number): BankSoalRoleWorkflowCard[] {
	const counts = summary?.counts ?? {};
	const configs = [
		{ key: 'my_draft', label: 'Draft saya', helper: 'soal pribadi yang masih bisa dilengkapi', tone: 'neutral' as const },
		{ key: 'my_review_waiting', label: 'Menunggu review saya', helper: 'antrean review sesuai role/scope', tone: 'warning' as const },
		{ key: 'revision_needed', label: 'Perlu revisi', helper: 'ditolak/dikembalikan untuk perbaikan', tone: 'warning' as const },
		{ key: 'approval_waiting', label: 'Menunggu approval', helper: 'sudah direview dan menunggu keputusan akhir', tone: 'warning' as const },
		{ key: 'package_ready', label: 'Siap paket', helper: 'approved/published dan boleh dipakai paket CBT', tone: 'success' as const },
		{ key: 'missing_metadata', label: 'Metadata kurang', helper: 'butuh mapel/tingkat/materi/level/CP-TP-KD', tone: 'danger' as const },
	];
	return configs.map((config) => ({
		...config,
		value: numberValue(counts[config.key]),
		evidenceLabel: evidenceLabel(summary?.counts ? 'summary' : 'missing', sampleSize),
	}));
}

function buildStatusCards(summary: BankSoalHealthSummary | null | undefined, questions: BankSoalHealthQuestion[]) {
	const keys: BankSoalStatusKey[] = ['total', 'published', 'approved', 'review', 'draft', 'revision', 'archived'];
	return keys.map((key): BankSoalStatusCard => {
		const summaryCount = countFromSummary(summary, key);
		if (summaryCount !== null) {
			return {
				key,
				label: statusLabels[key],
				value: summaryCount,
				evidence: 'summary',
				evidenceLabel: evidenceLabel('summary', questions.length),
				helper: statusHelpers[key],
				tone: statusTone(key, summaryCount),
			};
		}
		const sampleCount = countFromSample(questions, key);
		if (sampleCount !== null) {
			return {
				key,
				label: statusLabels[key],
				value: sampleCount,
				evidence: 'sample',
				evidenceLabel: evidenceLabel('sample', questions.length),
				helper: statusHelpers[key],
				tone: statusTone(key, sampleCount),
			};
		}
		return {
			key,
			label: statusLabels[key],
			value: null,
			evidence: 'missing',
			evidenceLabel: evidenceLabel('missing', questions.length),
			helper: statusHelpers[key],
			tone: 'neutral',
		};
	});
}

function ratioMetric(
	key: string,
	label: string,
	value: number | null,
	total: number | null,
	evidence: BankSoalHealthEvidence,
	sampleSize: number
): BankSoalRatioMetric {
	const missing = value !== null && total !== null ? Math.max(0, total - value) : null;
	const percent = value !== null && total !== null && total > 0 ? Math.round((value / total) * 100) : null;
	return {
		key,
		label,
		value,
		total,
		missing,
		percent,
		evidence,
		evidenceLabel: evidenceLabel(evidence, sampleSize),
	};
}

function buildSubjectCoverage(summary: BankSoalHealthSummary | null | undefined, questions: BankSoalHealthQuestion[]) {
	const subjectRows = summary?.by_subject ?? [];
	if (subjectRows.length > 0) {
		const covered = subjectRows.filter((subject) => (numberValue(subject.total) ?? 0) > 0).length;
		return ratioMetric('subject', 'Cakupan mapel', covered, subjectRows.length, 'summary', questions.length);
	}

	const subjectKeys = new Set(
		questions
			.map((question) => normalizeText(question.subject_id) || normalizeText(question.subject_name) || normalizeText(question.subject_code))
			.filter(Boolean)
	);
	if (subjectKeys.size > 0) {
		return ratioMetric('subject', 'Mapel terdeteksi di sampel', subjectKeys.size, null, 'sample', questions.length);
	}
	return ratioMetric('subject', 'Cakupan mapel', null, null, 'missing', questions.length);
}

function fieldCoverage(
	questions: BankSoalHealthQuestion[],
	key: string,
	label: string,
	fields: string[]
): BankSoalRatioMetric {
	if (questions.length === 0 || !hasAnyField(questions, fields)) {
		return ratioMetric(key, label, null, null, 'missing', questions.length);
	}
	const filled = questions.filter((question) => hasAnyFilledField(question, fields)).length;
	return ratioMetric(key, label, filled, questions.length, 'sample', questions.length);
}

function buildCurriculumCoverage(questions: BankSoalHealthQuestion[]) {
	return [
		fieldCoverage(questions, 'kd', 'KD terisi', ['kd_ref', 'kd', 'basic_competency']),
		fieldCoverage(questions, 'cp', 'CP terisi', ['cp_ref', 'cp', 'learning_outcome']),
		fieldCoverage(questions, 'tp', 'TP terisi', ['tp_ref', 'tp', 'learning_objective']),
	];
}

function buildMetadataQuality(questions: BankSoalHealthQuestion[]) {
	return [
		fieldCoverage(questions, 'kd', 'Tanpa KD', ['kd_ref', 'kd', 'basic_competency']),
		fieldCoverage(questions, 'cognitive_level', 'Tanpa level kognitif', ['cognitive_level', 'level', 'bloom_level']),
		fieldCoverage(questions, 'difficulty', 'Tanpa kesulitan', ['difficulty', 'difficulty_level']),
		fieldCoverage(questions, 'explanation', 'Tanpa pembahasan', [
			'explanation_html',
			'answer_explanation',
			'discussion_html',
			'solution_html',
			'explanation',
		]),
	];
}

function valueFor(cards: BankSoalStatusCard[], key: BankSoalStatusKey): number | null {
	return cards.find((card) => card.key === key)?.value ?? null;
}

function buildReviewBacklog(cards: BankSoalStatusCard[], sampleSize: number) {
	const reviewCard = cards.find((card) => card.key === 'review');
	const revisionCard = cards.find((card) => card.key === 'revision');
	if (reviewCard?.evidence === 'missing' && revisionCard?.evidence === 'missing') {
		return ratioMetric('review_backlog', 'Review backlog', null, null, 'missing', sampleSize);
	}
	const review = reviewCard?.value ?? 0;
	const revision = revisionCard?.value ?? 0;
	const total = valueFor(cards, 'total');
	const evidence = reviewCard?.evidence === 'summary' || revisionCard?.evidence === 'summary' ? 'summary' : 'sample';
	return ratioMetric('review_backlog', 'Review backlog', review + revision, total, evidence, sampleSize);
}

function averagePercent(metrics: BankSoalRatioMetric[]): number | null {
	const values = metrics.map((metric) => metric.percent).filter((value): value is number => value !== null);
	if (values.length === 0) return null;
	return Math.round(values.reduce((sum, value) => sum + value, 0) / values.length);
}

function readinessGrade(score: number | null): BankSoalReadiness['grade'] {
	if (score === null) return 'Perlu evidence/data';
	if (score >= 85) return 'A';
	if (score >= 70) return 'B';
	if (score >= 55) return 'C';
	return 'D';
}

function buildReadiness(
	cards: BankSoalStatusCard[],
	subjectCoverage: BankSoalRatioMetric,
	curriculumCoverage: BankSoalRatioMetric[],
	metadataQuality: BankSoalRatioMetric[],
	reviewBacklog: BankSoalRatioMetric,
	sampleSize: number
): BankSoalReadiness {
	const total = valueFor(cards, 'total');
	const approved = (valueFor(cards, 'approved') ?? 0) + (valueFor(cards, 'published') ?? 0);
	const approvalPercent = total && total > 0 ? Math.round((approved / total) * 100) : null;
	const metadataPercent = averagePercent([...curriculumCoverage, ...metadataQuality.filter((metric) => metric.key !== 'kd')]);
	const backlogHealth = total && total > 0 && reviewBacklog.value !== null
		? Math.max(0, Math.round((1 - Math.min(reviewBacklog.value / total, 1)) * 100))
		: null;
	const subjectPercent = subjectCoverage.percent;

	const weightedInputs = [
		{ value: approvalPercent, weight: 0.35 },
		{ value: metadataPercent, weight: 0.35 },
		{ value: backlogHealth, weight: 0.2 },
		{ value: subjectPercent, weight: 0.1 },
	].filter((item): item is { value: number; weight: number } => item.value !== null);

	const score = weightedInputs.length > 0
		? Math.round(weightedInputs.reduce((sum, item) => sum + item.value * item.weight, 0) / weightedInputs.reduce((sum, item) => sum + item.weight, 0))
		: null;

	const evidenceParts = new Set<string>();
	for (const card of cards) {
		if (card.evidence === 'summary') evidenceParts.add('summary');
		if (card.evidence === 'sample') evidenceParts.add(evidenceLabel('sample', sampleSize));
	}
	for (const metric of [...curriculumCoverage, ...metadataQuality, subjectCoverage]) {
		if (metric.evidence === 'sample') evidenceParts.add(evidenceLabel('sample', sampleSize));
		if (metric.evidence === 'summary') evidenceParts.add('summary');
	}

	return {
		score,
		grade: readinessGrade(score),
		evidenceLabel: evidenceParts.size > 0 ? Array.from(evidenceParts).join(' + ') : 'perlu evidence/data',
		drivers: [
			approvalPercent === null ? 'rasio approved/published perlu evidence/data' : `${approvalPercent}% approved/published`,
			metadataPercent === null ? 'kelengkapan metadata perlu evidence/data' : `${metadataPercent}% rata-rata metadata terisi`,
			backlogHealth === null ? 'backlog review perlu evidence/data' : `${reviewBacklog.value ?? 0} soal dalam review/revisi`,
		],
	};
}

function buildWarnings(
	cards: BankSoalStatusCard[],
	questions: BankSoalHealthQuestion[],
	reviewBacklog: BankSoalRatioMetric
): BankSoalHealthWarning[] {
	const warnings: BankSoalHealthWarning[] = [];
	if ((reviewBacklog.value ?? 0) > 0) {
		const percent = reviewBacklog.percent ?? 0;
		warnings.push({
			kind: 'review',
			label: 'Review backlog',
			message: `${reviewBacklog.value} soal menunggu review atau revisi.`,
			severity: percent >= 25 ? 'warning' : 'info',
			evidenceLabel: reviewBacklog.evidenceLabel,
		});
	}

	const importFields = ['import_batch_id', 'import_source', 'source_file', 'source_filename', 'imported_at'];
	if (!hasAnyField(questions, importFields)) {
		warnings.push({
			kind: 'import',
			label: 'Import',
			message: 'Riwayat import belum tersedia di summary atau sampel, jadi kesehatan import belum bisa dinilai.',
			severity: 'info',
			evidenceLabel: 'perlu evidence/data',
		});
	} else {
		const importedDrafts = questions.filter((question) => hasAnyFilledField(question, importFields) && publicationOf(question) !== 'published').length;
		if (importedDrafts > 0) {
			warnings.push({
				kind: 'import',
				label: 'Import',
				message: `${importedDrafts} soal sampel dari import belum terbit.`,
				severity: 'warning',
				evidenceLabel: evidenceLabel('sample', questions.length),
			});
		}
	}

	const assetFields = ['asset_count', 'assets', 'stem_media_url', 'stimulus_media_url', 'stem_audio_url', 'stimulus_audio_url'];
	if (!hasAnyField(questions, assetFields)) {
		warnings.push({
			kind: 'asset',
			label: 'Aset',
			message: 'Evidence aset belum tersedia, jadi dashboard tidak mengasumsikan risiko media.',
			severity: 'info',
			evidenceLabel: 'perlu evidence/data',
		});
	} else {
		const withoutAssetEvidence = questions.filter((question) => !hasAnyFilledField(question, assetFields)).length;
		if (withoutAssetEvidence > 0) {
			warnings.push({
				kind: 'asset',
				label: 'Aset',
				message: `${withoutAssetEvidence} soal sampel belum punya evidence aset terhubung.`,
				severity: 'warning',
				evidenceLabel: evidenceLabel('sample', questions.length),
			});
		}
	}

	const archivedCard = cards.find((card) => card.key === 'archived');
	if (archivedCard?.evidence === 'missing') {
		warnings.push({
			kind: 'governance',
			label: 'Arsip',
			message: 'Status arsip belum tersedia di summary; jangan simpulkan nol arsip tanpa data.',
			severity: 'info',
			evidenceLabel: 'perlu evidence/data',
		});
	}

	return warnings;
}

function buildActions(
	reviewBacklog: BankSoalRatioMetric,
	metadataQuality: BankSoalRatioMetric[],
	warnings: BankSoalHealthWarning[]
): BankSoalHealthAction[] {
	const actions: BankSoalHealthAction[] = [
		{
			label: 'Buka daftar soal',
			description: 'Audit sampel, filter status, dan cek metadata per butir.',
			href: '/bank-soal/daftar',
			capability: 'read',
			priority: 10,
		},
	];
	if ((reviewBacklog.value ?? 0) > 0) {
		actions.push({
			label: 'Review backlog',
			description: `${reviewBacklog.value} soal perlu keputusan reviewer.`,
			href: '/bank-soal/verifikasi',
			capability: 'review',
			priority: 20,
		});
	}
	if (metadataQuality.some((metric) => (metric.missing ?? 0) > 0)) {
		actions.push({
			label: 'Lengkapi metadata',
			description: 'Prioritaskan KD, level kognitif, kesulitan, dan pembahasan.',
			href: '/bank-soal/daftar',
			capability: 'update',
			priority: 30,
		});
	}
	actions.push({
		label: 'Tambah Soal',
		description: 'Isi stok baru dengan metadata lengkap sejak awal.',
		href: '/bank-soal/tambah',
		capability: 'create',
		priority: 40,
	});
	if (warnings.some((warning) => warning.kind === 'import')) {
		actions.push({
			label: 'Impor',
			description: 'Validasi hasil import dan mapping metadata sebelum publikasi.',
			href: '/bank-soal/impor',
			capability: 'import',
			priority: 50,
		});
	}
	return actions.sort((a, b) => a.priority - b.priority);
}

function hasPermission(user: BankSoalAccessUser | undefined, permission: string): boolean {
	return (user?.permissions ?? []).includes(permission);
}

function isAdmin(user: BankSoalAccessUser | undefined): boolean {
	return user?.role === 'admin' || (user?.roles ?? []).includes('admin');
}

function canUpdateBankSoal(user?: BankSoalAccessUser): boolean {
	return isAdmin(user) || hasPermission(user, 'bank_soal.update');
}

export function filterBankSoalHealthActions(actions: BankSoalHealthAction[], user?: BankSoalAccessUser): BankSoalHealthAction[] {
	return actions.filter((action) => {
		if (action.capability === 'read') return true;
		if (action.capability === 'create') return canCreateBankSoal(user);
		if (action.capability === 'review') return canReviewBankSoal(user);
		if (action.capability === 'import') return canImportBankSoal(user);
		if (action.capability === 'settings') return canManageBankSoalSettings(user);
		return canUpdateBankSoal(user);
	});
}

export function buildBankSoalHealthModel(input: BankSoalHealthInput): BankSoalHealthModel {
	const questions = input.questions ?? [];
	const sampleLimit = input.sampleLimit ?? 50;
	const statusCards = buildStatusCards(input.summary, questions);
	const roleWorkflowCards = buildRoleWorkflowCards(input.summary, questions.length);
	const subjectCoverage = buildSubjectCoverage(input.summary, questions);
	const curriculumCoverage = buildCurriculumCoverage(questions);
	const metadataQuality = buildMetadataQuality(questions);
	const reviewBacklog = buildReviewBacklog(statusCards, questions.length);
	const readiness = buildReadiness(statusCards, subjectCoverage, curriculumCoverage, metadataQuality, reviewBacklog, questions.length);
	const warnings = buildWarnings(statusCards, questions, reviewBacklog);
	const actions = buildActions(reviewBacklog, metadataQuality, warnings);

	return {
		statusCards,
		roleWorkflowCards,
		subjectCoverage,
		curriculumCoverage,
		metadataQuality,
		readiness,
		reviewBacklog,
		warnings,
		actions,
		sampleSize: questions.length,
		sampleLimit,
		totalQuestions: valueFor(statusCards, 'total'),
	};
}
