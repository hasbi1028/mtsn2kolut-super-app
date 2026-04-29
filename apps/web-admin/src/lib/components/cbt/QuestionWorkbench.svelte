<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Textarea } from '$lib/components/ui/textarea';
	import { Badge } from '$lib/components/ui/badge';
	import { toast } from '$lib/components/ui/sonner';
	import TiptapEditor from '$lib/components/TiptapEditor.svelte';
	import LatexBlock from '$lib/components/LatexBlock.svelte';
	import {
		QUESTION_VARIANT_MAP,
		evaluationStorageKey,
		type QuestionVariantConfig,
		type QuestionVariantEvaluation,
		type QuestionVariantId,
	} from '$lib/cbt/question-experiments';

	let { variant = 'studio' }: { variant?: QuestionVariantId } = $props();

	type Subject = { id: string; name: string; code: string };
	type OptionItem = { label: string; text?: string; html?: string; latex?: string; asset_id?: string };
	type UploadedAsset = { id: string; original_name: string; mime_type: string; purpose: string; url: string; file_size: number };
	type Question = {
		id: string;
		authoring_mode?: string;
		suggested_mode?: string;
		subject_id: string;
		subject_name: string;
		subject_code: string;
		code: string;
		question_text: string;
		question_type: string;
		options: OptionItem[];
		answer_key: string;
		explanation: string;
		difficulty: string;
		status: string;
		stem_html: string;
		stem_latex: string;
		stimulus_html: string;
		stimulus_latex: string;
		explanation_html: string;
		rubric_html: string;
		academic_phase: string;
		grade_level: number | null;
		cp_ref: string;
		tp_ref: string;
		kd_ref: string;
		indicator_ref: string;
		material_topic: string;
		cognitive_level: string;
		hots_flag: boolean;
		media_asset_ids: string[];
		workflow_status: string;
		version: number;
		author_username: string;
		reviewer_username: string;
		approver_username: string;
		writer_notes: string;
		review_notes: string;
		created_at: string;
		updated_at?: string;
		reviewed_at?: string | null;
		approved_at?: string | null;
	};
	type QuestionListResponse = {
		items: Question[];
		meta?: { total: number; limit: number; offset: number };
	};

	const pageSize = 10;
	const questionTypeLabel: Record<string, string> = {
		multiple_choice: 'Pilihan Ganda',
		multiple_answer: 'Jawaban Ganda',
		true_false: 'Benar / Salah',
		short_answer: 'Isian Singkat',
		essay: 'Uraian / Essay',
	};
	const difficultyLabel: Record<string, string> = { easy: 'Mudah', medium: 'Sedang', hard: 'Sulit' };
	const workflowLabel: Record<string, string> = { draft: 'Draft', review: 'Review', approved: 'Approved', rejected: 'Revisi' };
	const statusLabel: Record<string, string> = { draft: 'Draft', published: 'Published', archived: 'Arsip' };
	const wizardStepLabels = ['Dasar', 'Konten', 'Jawaban', 'Kurikulum & Review'];

	let loading = $state(true);
	let error = $state('');
	let questions = $state<Question[]>([]);
	let subjects = $state<Subject[]>([]);
	let uploadedAssets = $state<UploadedAsset[]>([]);
	let selectedDetail = $state<Question | null>(null);
	let detailBusy = $state(false);
	let totalItems = $state(0);
	let currentPage = $state(1);

	let search = $state('');
	let filterSubject = $state('');
	let filterWorkflow = $state('');
	let filterType = $state('');
	let filterHots = $state('');
	let showForm = $state(false);
	let editId = $state<string | null>(null);
	let assetFile = $state<File | null>(null);
	let assetPurpose = $state('general');
	let fAuthoringMode = $state<'beginner' | 'advance'>('beginner');

	let fSubjectId = $state('');
	let fCode = $state('');
	let fQuestionText = $state('');
	let fQuestionType = $state('multiple_choice');
	let fOptions = $state<OptionItem[]>(defaultOptions());
	let fAnswerKey = $state('A');
	let fExplanation = $state('');
	let fDifficulty = $state('medium');
	let fStatus = $state('draft');
	let fStemHTML = $state('');
	let fStemLatex = $state('');
	let fStimulusHTML = $state('');
	let fStimulusLatex = $state('');
	let fExplanationHTML = $state('');
	let fRubricHTML = $state('');
	let fAcademicPhase = $state('');
	let fGradeLevel = $state<number | null>(7);
	let fCPRef = $state('');
	let fTPRef = $state('');
	let fKDRef = $state('');
	let fIndicatorRef = $state('');
	let fMaterialTopic = $state('');
	let fCognitiveLevel = $state('');
	let fHotsFlag = $state(false);
	let fWorkflowStatus = $state('draft');
	let fWriterNotes = $state('');
	let fReviewNotes = $state('');
	let fMediaAssetIds = $state<string[]>([]);
	let fBusy = $state(false);
	let assetBusy = $state(false);
	let isAdvanceMode = $derived(fAuthoringMode === 'advance');
	let userRoles = $derived(page.data.user?.roles ?? (page.data.user?.role ? [page.data.user.role] : []));
	let canSubmitReview = $derived(userRoles.includes('admin') || userRoles.includes('guru'));
	let canApproveWorkflow = $derived(userRoles.includes('admin'));
	let variantConfig = $derived(QUESTION_VARIANT_MAP[variant]);
	let isWizardVariant = $derived(variant === 'wizard');
	let isGridVariant = $derived(variant === 'grid');
	let isDocumentVariant = $derived(variant === 'document');
	let isReviewVariant = $derived(variant === 'review');
	let isPackageVariant = $derived(variant === 'package-fit');
	let questionDraftCount = $derived(questions.filter((question) => question.workflow_status === 'draft').length);
	let questionReviewCount = $derived(questions.filter((question) => question.workflow_status === 'review').length);
	let questionPublishedCount = $derived(questions.filter((question) => question.status === 'published').length);
	let wizardStep = $state(1);
	let evaluationEase = $state(3);
	let evaluationSpeed = $state(3);
	let evaluationReview = $state(3);
	let evaluationFit = $state(3);
	let evaluationNote = $state('');
	let evaluationSavedAt = $state('');

	let pageCount = $derived(Math.max(1, Math.ceil(totalItems / pageSize)));
	let rangeStart = $derived(totalItems === 0 ? 0 : (currentPage - 1) * pageSize + 1);
	let rangeEnd = $derived(Math.min(totalItems, currentPage * pageSize));
	let qualityWarnings = $derived.by(() => {
		const warnings: string[] = [];
		if (!isAdvanceMode) {
			if (fQuestionType === 'essay' && !fRubricHTML.trim()) warnings.push('Rubrik essay belum diisi. Ini masih boleh di mode beginner dan bisa dilengkapi nanti di advance.');
			if (!fMaterialTopic.trim()) warnings.push('Topik materi belum diisi. Soal tetap bisa disimpan dan dilengkapi nanti di advance.');
			if (!fGradeLevel) warnings.push('Tingkat belum diisi. Ini opsional di beginner, tetapi akan membantu saat filtering.');
			return warnings;
		}
		if (!fMaterialTopic.trim()) warnings.push('Topik materi belum diisi. Ini penting untuk pencarian dan blueprint paket ujian.');
		if (!fCPRef.trim() || !fKDRef.trim()) warnings.push('Acuan kurikulum belum lengkap. Minimal CP dan KD sebaiknya terisi.');
		if (fQuestionType === 'essay' && !fRubricHTML.trim()) warnings.push('Soal essay sebaiknya memiliki rubrik HTML sebelum diajukan review.');
		if (fStatus === 'published' && fWorkflowStatus !== 'approved') warnings.push('Item published idealnya hanya keluar dari status workflow approved.');
		if (fHotsFlag && !fCognitiveLevel.trim()) warnings.push('Item HOTS sebaiknya disertai level kognitif agar mudah dikurasi.');
		if (objectiveType(fQuestionType) && fOptions.some((option) => !option.text?.trim() && !option.html?.trim() && !option.latex?.trim())) {
			warnings.push('Masih ada opsi objektif yang kosong.');
		}
		if (requiresAnswerKey(fQuestionType) && !fAnswerKey.trim()) warnings.push('Kunci jawaban belum diisi.');
		return warnings;
	});

	function loadEvaluation() {
		const raw = localStorage.getItem(evaluationStorageKey(variant));
		if (!raw) return;
		try {
			const parsed = JSON.parse(raw) as Partial<QuestionVariantEvaluation>;
			evaluationEase = parsed.ease ?? 3;
			evaluationSpeed = parsed.speed ?? 3;
			evaluationReview = parsed.review ?? 3;
			evaluationFit = parsed.fit ?? 3;
			evaluationNote = parsed.note ?? '';
			evaluationSavedAt = parsed.updatedAt ?? '';
		} catch {
			// Ignore malformed local experiment notes.
		}
	}

	function saveEvaluation() {
		const payload: QuestionVariantEvaluation = {
			ease: evaluationEase,
			speed: evaluationSpeed,
			review: evaluationReview,
			fit: evaluationFit,
			note: evaluationNote.trim(),
			updatedAt: new Date().toISOString(),
		};
		localStorage.setItem(evaluationStorageKey(variant), JSON.stringify(payload));
		evaluationSavedAt = payload.updatedAt;
		showToast('Catatan eksperimen disimpan di browser ini');
	}

	function resetEvaluation() {
		localStorage.removeItem(evaluationStorageKey(variant));
		evaluationEase = 3;
		evaluationSpeed = 3;
		evaluationReview = 3;
		evaluationFit = 3;
		evaluationNote = '';
		evaluationSavedAt = '';
		showToast('Catatan eksperimen dihapus dari browser ini');
	}

	function ratingText(value: number) {
		return ['Sangat kurang', 'Kurang', 'Cukup', 'Baik', 'Sangat baik'][value - 1] ?? 'Belum dinilai';
	}

	function defaultOptions(): OptionItem[] {
		return [
			{ label: 'A', text: '' },
			{ label: 'B', text: '' },
			{ label: 'C', text: '' },
			{ label: 'D', text: '' },
		];
	}

	function objectiveType(type: string) {
		return type === 'multiple_choice' || type === 'multiple_answer' || type === 'true_false';
	}

	function requiresAnswerKey(type: string) {
		return type !== 'essay';
	}

	function wizardSectionVisible(section: 'identity' | 'content' | 'answer' | 'review') {
		if (!isWizardVariant) return true;
		if (section === 'identity') return wizardStep === 1;
		if (section === 'content') return wizardStep === 2;
		if (section === 'answer') return wizardStep === 3;
		return wizardStep === 4;
	}

	function nextWizardStep() {
		wizardStep = Math.min(wizardStep + 1, 4);
	}

	function previousWizardStep() {
		wizardStep = Math.max(wizardStep - 1, 1);
	}

	function showToast(message: string) {
		toast.success(message);
	}

	function showError(message: string) {
		toast.error(message);
	}

	function buildListQuery(page = currentPage) {
		const params = new URLSearchParams();
		params.set('limit', String(pageSize));
		params.set('offset', String((page - 1) * pageSize));
		if (search.trim()) params.set('q', search.trim());
		if (filterSubject) params.set('subject_id', filterSubject);
		if (filterWorkflow) params.set('workflow_status', filterWorkflow);
		if (filterType) params.set('question_type', filterType);
		if (filterHots) params.set('hots', filterHots);
		return params.toString();
	}

	async function load(page = currentPage) {
		loading = true;
		error = '';
		try {
			const query = buildListQuery(page);
			const [qRes, aRes] = await Promise.all([fetch(`/api/cbt/questions?${query}`), fetch('/api/academic')]);
			const qJson = (await qRes.json().catch(() => ({}))) as QuestionListResponse & { error?: string };
			const aJson = await aRes.json();
			if (!qRes.ok || qJson.error) {
				error = qJson.error ?? 'Gagal memuat bank soal';
				return;
			}
			questions = qJson.items ?? [];
			totalItems = qJson.meta?.total ?? questions.length;
			currentPage = page;
			subjects = ((aJson.data ?? aJson)?.subjects ?? []) as Subject[];
		} catch (err) {
			error = (err as Error).message || 'Gagal memuat bank soal';
		} finally {
			loading = false;
		}
	}

	async function loadDetail(id: string) {
		detailBusy = true;
		try {
			const res = await fetch(`/api/cbt/questions/${id}`);
			const payload = (await res.json().catch(() => ({}))) as Question & { error?: string };
			if (!res.ok) {
				showError(payload.error ?? 'Gagal memuat detail soal');
				return null;
			}
			selectedDetail = payload;
			return payload;
		} finally {
			detailBusy = false;
		}
	}

	async function loadAssets(questionId: string) {
		try {
			const res = await fetch(`/api/cbt/assets?question_id=${encodeURIComponent(questionId)}`);
			const payload = (await res.json().catch(() => ([]))) as UploadedAsset[] | { error?: string };
			if (!res.ok) {
				showError((payload as { error?: string }).error ?? 'Gagal memuat asset soal');
				return;
			}
			uploadedAssets = Array.isArray(payload) ? payload : [];
		} catch (err) {
			showError((err as Error).message || 'Gagal memuat asset soal');
		}
	}

	function resetForm() {
		editId = null;
		showForm = false;
		fAuthoringMode = 'beginner';
		uploadedAssets = [];
		assetFile = null;
		assetPurpose = 'general';
		fSubjectId = '';
		fCode = '';
		fQuestionText = '';
		fQuestionType = 'multiple_choice';
		fOptions = defaultOptions();
		fAnswerKey = 'A';
		fExplanation = '';
		fDifficulty = 'medium';
		fStatus = 'draft';
		fStemHTML = '';
		fStemLatex = '';
		fStimulusHTML = '';
		fStimulusLatex = '';
		fExplanationHTML = '';
		fRubricHTML = '';
		fAcademicPhase = '';
		fGradeLevel = 7;
		fCPRef = '';
		fTPRef = '';
		fKDRef = '';
		fIndicatorRef = '';
		fMaterialTopic = '';
		fCognitiveLevel = '';
		fHotsFlag = false;
		fWorkflowStatus = 'draft';
		fWriterNotes = '';
		fReviewNotes = '';
		fMediaAssetIds = [];
	}

	function applyQuestionToForm(question: Question) {
		editId = question.id;
		showForm = true;
		fAuthoringMode = (question.suggested_mode === 'advance' || question.authoring_mode === 'advance') ? 'advance' : 'beginner';
		assetFile = null;
		assetPurpose = 'general';
		fSubjectId = question.subject_id;
		fCode = question.code;
		fQuestionText = question.question_text;
		fQuestionType = question.question_type || 'multiple_choice';
		fOptions = question.options?.length ? question.options.map((item) => ({ ...item })) : defaultOptions();
		fAnswerKey = question.answer_key || 'A';
		fExplanation = question.explanation;
		fDifficulty = question.difficulty;
		fStatus = question.status;
		fStemHTML = question.stem_html;
		fStemLatex = question.stem_latex;
		fStimulusHTML = question.stimulus_html;
		fStimulusLatex = question.stimulus_latex;
		fExplanationHTML = question.explanation_html;
		fRubricHTML = question.rubric_html;
		fAcademicPhase = question.academic_phase;
		fGradeLevel = question.grade_level ?? 7;
		fCPRef = question.cp_ref;
		fTPRef = question.tp_ref;
		fKDRef = question.kd_ref;
		fIndicatorRef = question.indicator_ref;
		fMaterialTopic = question.material_topic;
		fCognitiveLevel = question.cognitive_level;
		fHotsFlag = question.hots_flag;
		fWorkflowStatus = question.workflow_status;
		fWriterNotes = question.writer_notes;
		fReviewNotes = question.review_notes;
		fMediaAssetIds = question.media_asset_ids ?? [];
	}

	function openCreate() {
		if (showForm) {
			resetForm();
			return;
		}
		resetForm();
		showForm = true;
		selectedDetail = null;
	}

	async function openEdit(question: Question, forceMode?: 'beginner' | 'advance') {
		const detail = (await loadDetail(question.id)) ?? question;
		applyQuestionToForm(detail);
		if (forceMode) {
			fAuthoringMode = forceMode;
		}
		await loadAssets(question.id);
		window.scrollTo({ top: 0, behavior: 'smooth' });
	}

	async function openDetail(id: string) {
		await loadDetail(id);
		showForm = false;
		window.scrollTo({ top: 0, behavior: 'smooth' });
	}

	function addOption() {
		const nextLabel = String.fromCharCode(65 + fOptions.length);
		fOptions = [...fOptions, { label: nextLabel, text: '' }];
	}

	function removeOption(index: number) {
		if (fOptions.length <= 2) return;
		fOptions = fOptions
			.filter((_, idx) => idx !== index)
			.map((opt, idx) => ({ ...opt, label: String.fromCharCode(65 + idx) }));
	}

	function setTrueFalseTemplate() {
		fOptions = [
			{ label: 'A', text: 'Benar' },
			{ label: 'B', text: 'Salah' },
		];
		fAnswerKey = 'A';
	}

	function ensureObjectiveOptions() {
		if (!objectiveType(fQuestionType)) return;
		if (fQuestionType === 'true_false') {
			setTrueFalseTemplate();
			return;
		}
		if (fOptions.length < 4) fOptions = defaultOptions();
	}

	$effect(() => {
		if (!isAdvanceMode) {
			if (fQuestionType !== 'multiple_choice' && fQuestionType !== 'essay') {
				fQuestionType = 'multiple_choice';
			}
			fDifficulty = 'medium';
			fStatus = 'draft';
			fWorkflowStatus = 'draft';
		}
		ensureObjectiveOptions();
		if (fQuestionType === 'essay') fAnswerKey = '';
	});

	function insertAsset(tag: string, field: 'stem' | 'stimulus' | 'explanation' | 'rubric') {
		if (field === 'stem') fStemHTML += `${fStemHTML ? '\n' : ''}${tag}`;
		if (field === 'stimulus') fStimulusHTML += `${fStimulusHTML ? '\n' : ''}${tag}`;
		if (field === 'explanation') fExplanationHTML += `${fExplanationHTML ? '\n' : ''}${tag}`;
		if (field === 'rubric') fRubricHTML += `${fRubricHTML ? '\n' : ''}${tag}`;
	}

	async function uploadAsset() {
		if (!assetFile) return;
		assetBusy = true;
		try {
			const form = new FormData();
			form.set('file', assetFile);
			form.set('purpose', assetPurpose);
			if (editId) form.set('question_id', editId);
			const res = await fetch('/api/cbt/assets', { method: 'POST', body: form });
			const payload = (await res.json().catch(() => ({}))) as UploadedAsset & { error?: string };
			if (!res.ok) {
				showError(payload.error ?? 'Upload asset gagal');
				return;
			}
			uploadedAssets = [payload, ...uploadedAssets];
			fMediaAssetIds = Array.from(new Set([...fMediaAssetIds, payload.id]));
			assetFile = null;
			showToast('Asset berhasil diupload');
		} finally {
			assetBusy = false;
		}
	}

	async function uploadImageInEditor(file: File): Promise<string> {
		const form = new FormData();
		form.set('file', file);
		form.set('purpose', 'general');
		if (editId) form.set('question_id', editId);
		const res = await fetch('/api/cbt/assets', { method: 'POST', body: form });
		const payload = (await res.json().catch(() => ({}))) as UploadedAsset & { error?: string };
		if (!res.ok) throw new Error(payload.error ?? 'Upload gambar gagal');
		uploadedAssets = [payload, ...uploadedAssets];
		fMediaAssetIds = Array.from(new Set([...fMediaAssetIds, payload.id]));
		return payload.url;
	}

	function buildPayload() {
		return {
			authoring_mode: fAuthoringMode,
			subject_id: fSubjectId,
			code: fCode,
			question_text: fQuestionText,
			question_type: fQuestionType,
			options: fOptions,
			answer_key: fAnswerKey,
			explanation: fExplanation,
			difficulty: fDifficulty,
			status: fStatus,
			stem_html: fStemHTML,
			stem_latex: fStemLatex,
			stimulus_html: fStimulusHTML,
			stimulus_latex: fStimulusLatex,
			explanation_html: fExplanationHTML,
			rubric_html: fRubricHTML,
			academic_phase: fAcademicPhase,
			grade_level: fGradeLevel,
			cp_ref: fCPRef,
			tp_ref: fTPRef,
			kd_ref: fKDRef,
			indicator_ref: fIndicatorRef,
			material_topic: fMaterialTopic,
			cognitive_level: fCognitiveLevel,
			hots_flag: fHotsFlag,
			media_asset_ids: fMediaAssetIds,
			workflow_status: fWorkflowStatus,
			writer_notes: fWriterNotes,
			review_notes: fReviewNotes,
		};
	}

	async function saveQuestion() {
		if (!fSubjectId || (!fQuestionText && !fStemHTML && !fStemLatex)) {
			showError('Mapel dan isi soal utama wajib diisi');
			return;
		}
		if (!isAdvanceMode && fQuestionType === 'multiple_choice' && fOptions.filter((option) => option.text?.trim() || option.html?.trim() || option.latex?.trim()).length < 4) {
			showError('Mode beginner untuk pilihan ganda membutuhkan minimal 4 opsi terisi');
			return;
		}
		fBusy = true;
		try {
			const res = await fetch(editId ? `/api/cbt/questions/${editId}` : '/api/cbt/questions', {
				method: editId ? 'PUT' : 'POST',
				headers: { 'content-type': 'application/json' },
				body: JSON.stringify(buildPayload()),
			});
			const payload = await res.json().catch(() => ({}));
			if (!res.ok) {
				showError(payload.error ?? 'Gagal menyimpan soal');
				return;
			}
			showToast(editId ? 'Item bank soal diperbarui' : 'Item bank soal ditambahkan');
			resetForm();
			await load(editId && currentPage > pageCount ? pageCount : currentPage);
		} finally {
			fBusy = false;
		}
	}

	async function deleteQuestion(id: string) {
		if (!confirm('Hapus item bank soal ini?')) return;
		const res = await fetch(`/api/cbt/questions/${id}`, { method: 'DELETE' });
		if (!res.ok) {
			const payload = await res.json().catch(() => ({}));
			showError(payload.error ?? 'Gagal menghapus');
			return;
		}
		if (selectedDetail?.id === id) selectedDetail = null;
		showToast('Item bank soal dihapus');
		await load(currentPage);
	}

	async function duplicateQuestion(id: string) {
		const res = await fetch(`/api/cbt/questions/${id}/duplicate`, { method: 'POST' });
		const payload = (await res.json().catch(() => ({}))) as Question & { error?: string };
		if (!res.ok) {
			showError(payload.error ?? 'Gagal menggandakan item');
			return;
		}
		showToast('Draft hasil duplikasi berhasil dibuat');
		await load(1);
		await openEdit(payload);
	}

	function workflowBadgeClass(value: string) {
		if (value === 'approved') return 'bg-emerald-100 text-emerald-800 border-emerald-200';
		if (value === 'review') return 'bg-amber-100 text-amber-800 border-amber-200';
		if (value === 'rejected') return 'bg-rose-100 text-rose-800 border-rose-200';
		return 'bg-slate-100 text-slate-700 border-slate-200';
	}

	function publicationBadgeClass(value: string) {
		if (value === 'published') return 'bg-green-100 text-green-800 border-green-200';
		if (value === 'archived') return 'bg-slate-100 text-slate-500 border-slate-200';
		return 'bg-blue-50 text-blue-700 border-blue-200';
	}

	function difficultyBadgeClass(value: string) {
		if (value === 'easy') return 'bg-emerald-100 text-emerald-700 border-emerald-200';
		if (value === 'hard') return 'bg-rose-100 text-rose-700 border-rose-200';
		return 'bg-amber-100 text-amber-700 border-amber-200';
	}

	async function workflowAction(id: string, action: 'submit_review' | 'approve' | 'publish' | 'archive') {
		const res = await fetch(`/api/cbt/questions/${id}/workflow`, {
			method: 'PATCH',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({ action }),
		});
		const payload = await res.json().catch(() => ({}));
		if (!res.ok) {
			showError(payload.error ?? 'Aksi workflow gagal');
			return;
		}
		showToast('Status item diperbarui');
		await load(currentPage);
		if (selectedDetail?.id === id) {
			await loadDetail(id);
		}
	}

	function openAdvanceCompletion(question: Question) {
		void openEdit(question, 'advance');
	}

	async function applyFilters() {
		selectedDetail = null;
		await load(1);
	}

	async function resetFilters() {
		search = '';
		filterSubject = '';
		filterWorkflow = '';
		filterType = '';
		filterHots = '';
		selectedDetail = null;
		await load(1);
	}

	async function goToPage(page: number) {
		if (page < 1 || page > pageCount || page === currentPage) return;
		await load(page);
	}

	function handleSearchKeydown(event: KeyboardEvent) {
		if (event.key === 'Enter') {
			event.preventDefault();
			void applyFilters();
		}
	}

	onMount(() => {
		loadEvaluation();
		if (variant === 'review') {
			filterWorkflow = 'review';
		}
		if (variant === 'wizard' || variant === 'document') {
			showForm = true;
		}
		void load();
	});
</script>

<div class="space-y-6">
	<div class="flex flex-wrap items-start justify-between gap-4">
		<div class="space-y-1">
			<div class="flex flex-wrap items-center gap-2 text-xs font-semibold uppercase tracking-[0.22em] text-emerald-700">
				<button type="button" class="rounded-full border border-emerald-200 bg-emerald-50 px-3 py-1 text-left transition hover:border-emerald-300 hover:bg-emerald-100" onclick={() => goto(resolve('/cbt/questions'))}>
					Lab Eksperimen Bank Soal
				</button>
				<Badge variant="outline">{variantConfig.shortTitle}</Badge>
			</div>
			<h1 class="text-2xl font-semibold tracking-tight text-slate-900">{variantConfig.title} Bank Soal CBT</h1>
			<p class="max-w-3xl text-sm text-slate-600">
				{variantConfig.description}
			</p>
			<p class="max-w-2xl text-sm text-slate-500">{variantConfig.persona}</p>
		</div>
		<div class="flex flex-wrap gap-2">
			<Button variant="outline" onclick={() => goto(resolve('/cbt/questions'))}>Lihat Semua Varian</Button>
			<Button onclick={openCreate}>{showForm ? 'Tutup Form' : '+ Tambah Item'}</Button>
		</div>
	</div>

	<div class="grid gap-4 xl:grid-cols-[minmax(0,1.45fr),minmax(320px,0.95fr)]">
		<Card.Root class="border-emerald-200/70 bg-white/90 shadow-sm">
			<Card.Header class="pb-4">
				<Card.Title class="text-base text-slate-900">Karakter Pendekatan</Card.Title>
				<Card.Description>Setiap route ini tetap memakai backend yang sama. Yang berubah hanya pengalaman kerja untuk guru dan admin.</Card.Description>
			</Card.Header>
			<Card.Content class="grid gap-4 md:grid-cols-2">
				<div class="space-y-2">
					<p class="text-xs font-semibold uppercase tracking-[0.18em] text-emerald-800">Kelebihan Utama</p>
					<ul class="space-y-2 text-sm text-slate-700">
						{#each variantConfig.highlights as highlight (highlight)}
							<li class="rounded-lg border border-emerald-100 bg-emerald-50/60 px-3 py-2">{highlight}</li>
						{/each}
					</ul>
				</div>
				<div class="space-y-2">
					<p class="text-xs font-semibold uppercase tracking-[0.18em] text-slate-600">Hal yang Perlu Diperhatikan</p>
					<ul class="space-y-2 text-sm text-slate-700">
						{#each variantConfig.cautions as caution (caution)}
							<li class="rounded-lg border border-slate-200 bg-slate-50 px-3 py-2">{caution}</li>
						{/each}
					</ul>
				</div>
			</Card.Content>
		</Card.Root>

		<Card.Root class="border-slate-200 shadow-sm">
			<Card.Header class="pb-4">
				<Card.Title class="text-base text-slate-900">Catatan Uji Guru</Card.Title>
				<Card.Description>Disimpan lokal di browser ini agar guru bisa membandingkan keenam varian tanpa mengubah backend.</Card.Description>
			</Card.Header>
			<Card.Content class="space-y-4">
				<div class="grid gap-3 sm:grid-cols-2">
					<div>
						<label for={`eval-ease-${variant}`} class="mb-1 block text-xs font-medium text-slate-600">Mudah dipahami</label>
						<select id={`eval-ease-${variant}`} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={evaluationEase}>
							{#each [1, 2, 3, 4, 5] as value (value)}
								<option value={value}>{value} — {ratingText(value)}</option>
							{/each}
						</select>
					</div>
					<div>
						<label for={`eval-speed-${variant}`} class="mb-1 block text-xs font-medium text-slate-600">Cepat untuk input</label>
						<select id={`eval-speed-${variant}`} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={evaluationSpeed}>
							{#each [1, 2, 3, 4, 5] as value (value)}
								<option value={value}>{value} — {ratingText(value)}</option>
							{/each}
						</select>
					</div>
					<div>
						<label for={`eval-review-${variant}`} class="mb-1 block text-xs font-medium text-slate-600">Nyaman untuk review</label>
						<select id={`eval-review-${variant}`} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={evaluationReview}>
							{#each [1, 2, 3, 4, 5] as value (value)}
								<option value={value}>{value} — {ratingText(value)}</option>
							{/each}
						</select>
					</div>
					<div>
						<label for={`eval-fit-${variant}`} class="mb-1 block text-xs font-medium text-slate-600">Cocok untuk guru</label>
						<select id={`eval-fit-${variant}`} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={evaluationFit}>
							{#each [1, 2, 3, 4, 5] as value (value)}
								<option value={value}>{value} — {ratingText(value)}</option>
							{/each}
						</select>
					</div>
				</div>
				<div>
					<label for={`eval-note-${variant}`} class="mb-1 block text-xs font-medium text-slate-600">Catatan guru / admin</label>
					<Textarea id={`eval-note-${variant}`} rows={3} placeholder="Tuliskan apa yang terasa paling mudah, paling membingungkan, atau paling cocok untuk guru mapel..." bind:value={evaluationNote} />
				</div>
				<div class="flex flex-wrap items-center justify-between gap-3 text-xs text-slate-500">
					<p>{evaluationSavedAt ? `Tersimpan terakhir: ${new Date(evaluationSavedAt).toLocaleString('id-ID')}` : 'Belum ada catatan tersimpan'}</p>
					<div class="flex gap-2">
						<Button variant="outline" size="sm" onclick={resetEvaluation}>Reset Catatan</Button>
						<Button size="sm" onclick={saveEvaluation}>Simpan Catatan</Button>
					</div>
				</div>
			</Card.Content>
		</Card.Root>
	</div>

	<div class={`grid gap-4 ${isGridVariant || isReviewVariant ? 'xl:grid-cols-3' : 'xl:grid-cols-4'}`}>
		<div class="rounded-2xl border border-emerald-200 bg-gradient-to-br from-white via-emerald-50/40 to-emerald-100/50 px-4 py-4 shadow-sm">
			<p class="text-xs font-semibold uppercase tracking-[0.18em] text-emerald-800">Total Item</p>
			<p class="mt-3 text-3xl font-semibold text-slate-900">{totalItems}</p>
			<p class="mt-1 text-sm text-slate-600">Seluruh item yang cocok dengan filter saat ini.</p>
		</div>
		<div class="rounded-2xl border border-slate-200 bg-white px-4 py-4 shadow-sm">
			<p class="text-xs font-semibold uppercase tracking-[0.18em] text-slate-600">Draft</p>
			<p class="mt-3 text-3xl font-semibold text-slate-900">{questionDraftCount}</p>
			<p class="mt-1 text-sm text-slate-600">Cocok untuk dipakai saat menguji ritme authoring beginner.</p>
		</div>
		<div class="rounded-2xl border border-slate-200 bg-white px-4 py-4 shadow-sm">
			<p class="text-xs font-semibold uppercase tracking-[0.18em] text-slate-600">Review</p>
			<p class="mt-3 text-3xl font-semibold text-slate-900">{questionReviewCount}</p>
			<p class="mt-1 text-sm text-slate-600">Paling berguna untuk varian review dan package-fit.</p>
		</div>
		<div class="rounded-2xl border border-slate-200 bg-white px-4 py-4 shadow-sm">
			<p class="text-xs font-semibold uppercase tracking-[0.18em] text-slate-600">Published</p>
			<p class="mt-3 text-3xl font-semibold text-slate-900">{questionPublishedCount}</p>
			<p class="mt-1 text-sm text-slate-600">Menunjukkan item yang sudah siap dipakai ke paket ujian.</p>
		</div>
	</div>

	{#if error}
		<div class="rounded-lg border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-800">{error}</div>
	{/if}

	{#if isWizardVariant}
		<div class="rounded-2xl border border-emerald-200 bg-emerald-50/70 p-4 shadow-sm">
			<div class="flex flex-wrap items-center justify-between gap-3">
				<div>
					<p class="text-xs font-semibold uppercase tracking-[0.18em] text-emerald-800">Alur Wizard</p>
					<p class="mt-1 text-sm text-slate-700">Guru diarahkan lewat empat langkah kecil. Catalog tetap tersedia di bawah untuk referensi dan edit cepat.</p>
				</div>
				<div class="flex flex-wrap gap-2">
					{#each wizardStepLabels as label, index (`${label}-${index}`)}
						<div class={`rounded-full border px-3 py-1 text-xs font-semibold ${wizardStep === index + 1 ? 'border-emerald-300 bg-white text-emerald-800' : 'border-emerald-200 bg-emerald-100/60 text-slate-600'}`}>
							{index + 1}. {label}
						</div>
					{/each}
				</div>
			</div>
		</div>
	{:else if isReviewVariant}
		<div class="rounded-2xl border border-amber-200 bg-amber-50/70 p-4 shadow-sm">
			<div class="grid gap-4 md:grid-cols-3">
				<div>
					<p class="text-xs font-semibold uppercase tracking-[0.18em] text-amber-800">Antrian Review</p>
					<p class="mt-2 text-2xl font-semibold text-slate-900">{questionReviewCount}</p>
					<p class="text-sm text-slate-600">Item yang siap dibaca reviewer.</p>
				</div>
				<div>
					<p class="text-xs font-semibold uppercase tracking-[0.18em] text-amber-800">Draft Siap Naik</p>
					<p class="mt-2 text-2xl font-semibold text-slate-900">{questionDraftCount}</p>
					<p class="text-sm text-slate-600">Gunakan tombol review cepat pada katalog.</p>
				</div>
				<div>
					<p class="text-xs font-semibold uppercase tracking-[0.18em] text-amber-800">Fokus Route</p>
					<p class="mt-2 text-sm text-slate-700">Cocok untuk admin/guru senior yang lebih sering kurasi dibanding menulis dari nol.</p>
				</div>
			</div>
		</div>
	{:else if isPackageVariant}
		<div class="rounded-2xl border border-sky-200 bg-sky-50/70 p-4 shadow-sm">
			<div class="grid gap-4 lg:grid-cols-3">
				<div>
					<p class="text-xs font-semibold uppercase tracking-[0.18em] text-sky-800">Konteks Paket</p>
					<p class="mt-2 text-sm text-slate-700">Isi topik, tingkat, dan workflow agar item mudah dipilih saat menyusun UTS, UAS, atau event CBT per tingkat.</p>
				</div>
				<div>
					<p class="text-xs font-semibold uppercase tracking-[0.18em] text-sky-800">Readiness Cepat</p>
					<p class="mt-2 text-sm text-slate-700">Mode ini menekankan kecocokan item terhadap paket, bukan hanya kebenaran isi soal.</p>
				</div>
				<div>
					<p class="text-xs font-semibold uppercase tracking-[0.18em] text-sky-800">Saran Praktis</p>
					<p class="mt-2 text-sm text-slate-700">Minimal lengkapi mapel, tingkat, topik, dan workflow sebelum soal dipakai untuk penyusunan paket.</p>
				</div>
			</div>
		</div>
	{/if}

	{#if selectedDetail}
		<Card.Root class="border-slate-200 shadow-sm">
			<Card.Header class="border-b bg-slate-50/80">
				<div class="flex flex-wrap items-start justify-between gap-3">
					<div class="space-y-2">
						<Card.Title class="text-lg text-slate-900">Detail Item: {selectedDetail.code || 'Tanpa kode'}</Card.Title>
						<Card.Description>
							{selectedDetail.subject_code} · Versi {selectedDetail.version} · {selectedDetail.author_username || 'penulis tidak tercatat'}
						</Card.Description>
					</div>
					<div class="flex flex-wrap gap-2">
						{#if selectedDetail.suggested_mode !== 'advance'}
							<Button variant="outline" size="sm" onclick={() => selectedDetail && openEdit(selectedDetail, 'advance')}>Lengkapi di Advanced</Button>
						{/if}
						<Button variant="outline" size="sm" onclick={() => selectedDetail && openEdit(selectedDetail)}>Edit</Button>
						<Button variant="outline" size="sm" onclick={() => (selectedDetail = null)}>Tutup</Button>
					</div>
				</div>
			</Card.Header>
			<Card.Content class="space-y-4 pt-6">
				<div class="flex flex-wrap gap-2">
					<Badge class={workflowBadgeClass(selectedDetail.workflow_status)}>{workflowLabel[selectedDetail.workflow_status] ?? selectedDetail.workflow_status}</Badge>
					<Badge class={publicationBadgeClass(selectedDetail.status)}>{statusLabel[selectedDetail.status] ?? selectedDetail.status}</Badge>
					<Badge class={difficultyBadgeClass(selectedDetail.difficulty)}>{difficultyLabel[selectedDetail.difficulty] ?? selectedDetail.difficulty}</Badge>
					<Badge variant="outline">{questionTypeLabel[selectedDetail.question_type] ?? selectedDetail.question_type}</Badge>
					{#if selectedDetail.hots_flag}
						<Badge class="border-emerald-200 bg-emerald-100 text-emerald-800">HOTS</Badge>
					{/if}
				</div>
				<div class="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
					<div class="rounded-lg border bg-slate-50 p-3 text-sm">
						<p class="text-xs uppercase tracking-[0.18em] text-slate-500">Kurikulum</p>
						<p class="mt-2 text-slate-900">CP: {selectedDetail.cp_ref || '—'}</p>
						<p class="text-slate-900">KD: {selectedDetail.kd_ref || '—'}</p>
						<p class="text-slate-900">TP: {selectedDetail.tp_ref || '—'}</p>
					</div>
					<div class="rounded-lg border bg-slate-50 p-3 text-sm">
						<p class="text-xs uppercase tracking-[0.18em] text-slate-500">Blueprint</p>
						<p class="mt-2 text-slate-900">{selectedDetail.material_topic || 'Topik belum diisi'}</p>
						<p class="text-slate-900">Level: {selectedDetail.cognitive_level || '—'}</p>
						<p class="text-slate-900">Kelas: {selectedDetail.grade_level ?? '—'}</p>
					</div>
					<div class="rounded-lg border bg-slate-50 p-3 text-sm">
						<p class="text-xs uppercase tracking-[0.18em] text-slate-500">Review</p>
						<p class="mt-2 text-slate-900">Reviewer: {selectedDetail.reviewer_username || '—'}</p>
						<p class="text-slate-900">Approver: {selectedDetail.approver_username || '—'}</p>
					</div>
					<div class="rounded-lg border bg-slate-50 p-3 text-sm">
						<p class="text-xs uppercase tracking-[0.18em] text-slate-500">Asset</p>
						<p class="mt-2 text-slate-900">{selectedDetail.media_asset_ids.length} asset direferensikan</p>
						<p class="text-slate-900">Workflow: {workflowLabel[selectedDetail.workflow_status] ?? selectedDetail.workflow_status}</p>
					</div>
				</div>
				<div class="rounded-xl border border-emerald-200 bg-white p-4 shadow-sm">
					{#if selectedDetail.stimulus_html}
						<div class="prose prose-sm max-w-none rounded-lg border bg-slate-50 p-3">{@html selectedDetail.stimulus_html}</div>
					{/if}
					{#if selectedDetail.stimulus_latex}
						<LatexBlock src={selectedDetail.stimulus_latex} class="mt-3" />
					{/if}
					{#if selectedDetail.stem_html}
						<div class="prose prose-sm mt-4 max-w-none text-slate-900">{@html selectedDetail.stem_html}</div>
					{:else}
						<p class="mt-4 text-sm text-slate-900">{selectedDetail.question_text || 'Tanpa ringkasan teks'}</p>
					{/if}
					{#if selectedDetail.stem_latex}
						<LatexBlock src={selectedDetail.stem_latex} class="mt-3" />
					{/if}
					{#if objectiveType(selectedDetail.question_type)}
						<div class="mt-4 space-y-2">
							{#each selectedDetail.options as option (option.label)}
								<div class="rounded-lg border px-3 py-2 text-sm">
									<span class="mr-2 font-semibold text-emerald-800">{option.label}.</span>
									{#if option.html}
										<span>{@html option.html}</span>
									{:else if option.latex}
										<LatexBlock src={option.latex} display={false} class="inline-block" />
									{:else}
										<span>{option.text || '—'}</span>
									{/if}
								</div>
							{/each}
						</div>
					{/if}
				</div>
				<div class="rounded-2xl border border-slate-200 bg-slate-950 p-5 text-slate-50 shadow-sm">
					<div class="flex flex-wrap items-center justify-between gap-3 border-b border-slate-800 pb-3">
						<div>
							<p class="text-xs font-semibold uppercase tracking-[0.22em] text-emerald-300">Preview Peserta</p>
							<p class="mt-1 text-sm text-slate-300">Simulasi tampilan soal saat dibuka oleh siswa pada client ujian.</p>
						</div>
						<div class="rounded-full border border-slate-700 px-3 py-1 text-xs text-slate-300">
							{questionTypeLabel[selectedDetail.question_type] ?? selectedDetail.question_type}
						</div>
					</div>
					<div class="mt-4 space-y-4">
						{#if selectedDetail.stimulus_html}
							<div class="rounded-xl border border-slate-800 bg-slate-900 p-4 text-slate-100">{@html selectedDetail.stimulus_html}</div>
						{/if}
						{#if selectedDetail.stimulus_latex}
							<div class="rounded-xl border border-slate-800 bg-slate-900 p-4"><LatexBlock src={selectedDetail.stimulus_latex} class="text-white" /></div>
						{/if}
						<div class="rounded-xl border border-slate-800 bg-slate-900 p-4">
							{#if selectedDetail.stem_html}
								<div class="prose prose-invert max-w-none text-slate-100">{@html selectedDetail.stem_html}</div>
							{:else}
								<p class="text-sm text-slate-100">{selectedDetail.question_text || 'Tanpa ringkasan teks'}</p>
							{/if}
							{#if selectedDetail.stem_latex}
								<LatexBlock src={selectedDetail.stem_latex} class="mt-3 text-white" />
							{/if}
						</div>
						{#if objectiveType(selectedDetail.question_type)}
							<div class="grid gap-3">
								{#each selectedDetail.options as option (option.label)}
									<label class="flex items-start gap-3 rounded-xl border border-slate-800 bg-slate-900 px-4 py-3">
										<input type="radio" disabled class="mt-1 size-4 accent-emerald-500" />
										<div class="text-sm text-slate-100">
											<span class="mr-2 font-semibold text-emerald-300">{option.label}.</span>
											{#if option.html}
												<span>{@html option.html}</span>
											{:else if option.latex}
												<LatexBlock src={option.latex} display={false} class="inline-block text-white" />
											{:else}
												<span>{option.text || '—'}</span>
											{/if}
										</div>
									</label>
								{/each}
							</div>
						{:else}
							<div class="rounded-xl border border-dashed border-slate-700 bg-slate-900 px-4 py-4 text-sm text-slate-300">
								Peserta akan melihat editor jawaban essay / uraian pada area ini.
							</div>
						{/if}
					</div>
				</div>
			</Card.Content>
		</Card.Root>
	{/if}

	{#if showForm}
		<Card.Root class="border-emerald-200/70 shadow-sm">
			<Card.Header class="border-b bg-emerald-50/50">
				<Card.Title class="text-lg text-slate-900">{editId ? 'Edit Item Bank Soal' : 'Item Bank Soal Baru'}</Card.Title>
				<Card.Description>
					{#if isAdvanceMode}
						Mode advance mendukung konten teks polos, HTML terformat, LaTeX, stimulus, dan metadata kurikulum hybrid CP/TP + KD.
					{:else}
						Mode beginner dirancang agar guru bisa menulis soal lebih cepat tanpa harus langsung mengisi KD, CP, workflow, atau metadata lengkap lainnya.
					{/if}
				</Card.Description>
			</Card.Header>
			<Card.Content class="space-y-6 pt-6">
				{#if wizardSectionVisible('identity')}
				<section class="space-y-3">
					<h2 class="text-sm font-semibold uppercase tracking-[0.18em] text-emerald-800">Mode Authoring</h2>
					<div class="grid gap-3 md:grid-cols-2">
						<button
							type="button"
							class={`rounded-xl border p-4 text-left transition ${!isAdvanceMode ? 'border-emerald-300 bg-emerald-50 shadow-sm' : 'border-slate-200 bg-white hover:border-emerald-200'}`}
							onclick={() => (fAuthoringMode = 'beginner')}
						>
							<p class="text-sm font-semibold text-slate-900">Beginner</p>
							<p class="mt-1 text-sm text-slate-600">Fokus pada mapel, isi soal, opsi, jawaban, dan gambar sederhana. Cocok untuk guru yang ingin input cepat.</p>
						</button>
						<button
							type="button"
							class={`rounded-xl border p-4 text-left transition ${isAdvanceMode ? 'border-emerald-300 bg-emerald-50 shadow-sm' : 'border-slate-200 bg-white hover:border-emerald-200'}`}
							onclick={() => (fAuthoringMode = 'advance')}
						>
							<p class="text-sm font-semibold text-slate-900">Advance</p>
							<p class="mt-1 text-sm text-slate-600">Buka seluruh fitur blueprint kurikulum, workflow review, asset reuse, rich text, dan LaTeX.</p>
						</button>
					</div>
					{#if !isAdvanceMode}
						<div class="flex flex-wrap gap-2">
							<Button variant="outline" size="sm" onclick={() => (fAuthoringMode = 'advance')}>Lengkapi di Advanced</Button>
							{#if canSubmitReview}
								<Button variant="outline" size="sm" onclick={async () => {
									if (!editId) {
										showError('Simpan item terlebih dahulu sebelum diajukan review');
										return;
									}
									await workflowAction(editId, 'submit_review');
								}}>Ajukan Review</Button>
							{/if}
						</div>
					{/if}
				</section>
				{/if}

				{#if qualityWarnings.length > 0}
					<div class="rounded-xl border border-amber-200 bg-amber-50 px-4 py-3">
						<p class="text-sm font-semibold text-amber-900">Peringatan mutu ringan</p>
						<ul class="mt-2 space-y-1 text-sm text-amber-800">
							{#each qualityWarnings as warning (warning)}
								<li>{warning}</li>
							{/each}
						</ul>
					</div>
				{/if}

				<section class="space-y-3">
					<h2 class="text-sm font-semibold uppercase tracking-[0.18em] text-emerald-800">Identitas dan Blueprint</h2>
					<div class="grid gap-3 md:grid-cols-4">
						<div>
							<label for="q-subject" class="mb-1 block text-xs font-medium text-slate-600">Mata Pelajaran</label>
							<select id="q-subject" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fSubjectId}>
								<option value="">Pilih mapel</option>
								{#each subjects as subject (subject.id)}
									<option value={subject.id}>{subject.code} — {subject.name}</option>
								{/each}
							</select>
						</div>
						<div>
							<label for="q-code" class="mb-1 block text-xs font-medium text-slate-600">Kode Item</label>
							<Input id="q-code" placeholder="MTK-VII-HOTS-001" bind:value={fCode} />
						</div>
						<div>
							<label for="q-type" class="mb-1 block text-xs font-medium text-slate-600">Tipe Soal</label>
							<select id="q-type" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fQuestionType}>
								<option value="multiple_choice">Pilihan Ganda</option>
								{#if isAdvanceMode}
									<option value="multiple_answer">Jawaban Ganda</option>
									<option value="true_false">Benar / Salah</option>
									<option value="short_answer">Isian Singkat</option>
								{/if}
								<option value="essay">Essay</option>
							</select>
						</div>
						<div>
							<label for="q-grade" class="mb-1 block text-xs font-medium text-slate-600">Tingkat</label>
							<Input id="q-grade" type="number" min={1} max={12} bind:value={fGradeLevel} />
						</div>
					</div>

					{#if isAdvanceMode}
						<div class="grid gap-3 md:grid-cols-4">
							<div>
								<label for="q-phase" class="mb-1 block text-xs font-medium text-slate-600">Fase Akademik</label>
								<Input id="q-phase" placeholder="Fase D" bind:value={fAcademicPhase} />
							</div>
							<div>
								<label for="q-difficulty" class="mb-1 block text-xs font-medium text-slate-600">Kesulitan</label>
								<select id="q-difficulty" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fDifficulty}>
									<option value="easy">Mudah</option>
									<option value="medium">Sedang</option>
									<option value="hard">Sulit</option>
								</select>
							</div>
							<div>
								<label for="q-workflow" class="mb-1 block text-xs font-medium text-slate-600">Workflow</label>
								<select id="q-workflow" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fWorkflowStatus}>
									<option value="draft">Draft</option>
									<option value="review">Review</option>
									<option value="approved">Approved</option>
									<option value="rejected">Perlu Revisi</option>
								</select>
							</div>
							<div>
								<label for="q-status" class="mb-1 block text-xs font-medium text-slate-600">Publikasi</label>
								<select id="q-status" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fStatus}>
									<option value="draft">Draft</option>
									<option value="published">Published</option>
									<option value="archived">Arsip</option>
								</select>
							</div>
						</div>

						<div class="grid gap-3 md:grid-cols-2">
							<div>
								<label for="q-topic" class="mb-1 block text-xs font-medium text-slate-600">Topik Materi</label>
								<Input id="q-topic" placeholder="Persamaan Linear Satu Variabel" bind:value={fMaterialTopic} />
							</div>
							<div>
								<label for="q-cognitive" class="mb-1 block text-xs font-medium text-slate-600">Level Kognitif</label>
								<Input id="q-cognitive" placeholder="C3 / Analisis / HOTS" bind:value={fCognitiveLevel} />
							</div>
						</div>

						<div class="grid gap-3 md:grid-cols-4">
							<div>
								<label for="q-cp" class="mb-1 block text-xs font-medium text-slate-600">CP Ref</label>
								<Input id="q-cp" placeholder="CP-MTK-FD-2" bind:value={fCPRef} />
							</div>
							<div>
								<label for="q-tp" class="mb-1 block text-xs font-medium text-slate-600">TP Ref</label>
								<Input id="q-tp" placeholder="TP-7.1.3" bind:value={fTPRef} />
							</div>
							<div>
								<label for="q-kd" class="mb-1 block text-xs font-medium text-slate-600">KD Ref</label>
								<Input id="q-kd" placeholder="3.2" bind:value={fKDRef} />
							</div>
							<div>
								<label for="q-indicator" class="mb-1 block text-xs font-medium text-slate-600">Indikator</label>
								<Input id="q-indicator" placeholder="Menentukan solusi..." bind:value={fIndicatorRef} />
							</div>
						</div>

						<label class="flex items-center gap-2 rounded-md border border-emerald-200 bg-emerald-50 px-3 py-2 text-sm text-emerald-900">
							<input type="checkbox" bind:checked={fHotsFlag} class="size-4 accent-emerald-700" />
							Tandai item ini sebagai HOTS
						</label>
					{:else}
						<div class="rounded-lg border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-900">
							Soal mode beginner akan otomatis disimpan sebagai <strong>draft</strong> dengan workflow <strong>draft</strong>. KD, CP, TP, HOTS, dan metadata blueprint lain bisa dilengkapi nanti di mode advance.
						</div>
					{/if}
				</section>

				{#if wizardSectionVisible('content')}
				<section class="space-y-3">
					<h2 class="text-sm font-semibold uppercase tracking-[0.18em] text-emerald-800">Konten Soal</h2>
					<div>
						<label for="q-plain" class="mb-1 block text-xs font-medium text-slate-600">Ringkasan Teks Polos / Fallback</label>
						<Textarea id="q-plain" rows={2} placeholder="Teks ringkas untuk pencarian, fallback, dan kompatibilitas." bind:value={fQuestionText} />
					</div>
					<div class="space-y-4">
						{#if isAdvanceMode}
							<div>
								<label for="q-stimulus-html" class="mb-1 block text-xs font-medium text-slate-600">
									Stimulus <span class="font-normal text-slate-400">(narasi, tabel, gambar, atau rumus pendukung)</span>
								</label>
								<TiptapEditor
									id="q-stimulus-html"
									minRows={5}
									placeholder="Stimulus, narasi, tabel, atau gambar pendukung..."
									onImageUpload={uploadImageInEditor}
									bind:value={fStimulusHTML}
								/>
							</div>
							<div>
								<label for="q-stimulus-latex" class="mb-1 block text-xs font-medium text-slate-600">
									Stimulus LaTeX <span class="font-normal text-slate-400">(opsional — untuk rumus mandiri di luar editor)</span>
								</label>
								<Textarea id="q-stimulus-latex" rows={2} placeholder={'\\frac{2x+3}{5}=7'} bind:value={fStimulusLatex} class="font-mono text-sm" />
								{#if fStimulusLatex}
									<div class="mt-2 rounded-lg border bg-slate-50 p-3"><LatexBlock src={fStimulusLatex} /></div>
								{/if}
							</div>
						{/if}
						<div>
							<label for="q-stem-html" class="mb-1 block text-xs font-medium text-slate-600">
								{isAdvanceMode ? 'Stem / Pokok Soal' : 'Isi Soal'}
							</label>
							<TiptapEditor
								id="q-stem-html"
								minRows={isAdvanceMode ? 6 : 5}
								placeholder={isAdvanceMode ? 'Tuliskan pokok soal utama. Gunakan toolbar untuk format teks, rumus LaTeX (∑), atau sisipkan gambar.' : 'Tuliskan soal dengan mudah. Di mode beginner kamu tetap bisa menyisipkan gambar dari toolbar editor.'}
								onImageUpload={uploadImageInEditor}
								bind:value={fStemHTML}
							/>
						</div>
						{#if isAdvanceMode}
							<div>
								<label for="q-stem-latex" class="mb-1 block text-xs font-medium text-slate-600">
									Stem LaTeX <span class="font-normal text-slate-400">(opsional — untuk rumus mandiri di luar editor)</span>
								</label>
								<Textarea id="q-stem-latex" rows={2} placeholder="x^2 + 5x + 6 = 0" bind:value={fStemLatex} class="font-mono text-sm" />
								{#if fStemLatex}
									<div class="mt-2 rounded-lg border bg-slate-50 p-3"><LatexBlock src={fStemLatex} /></div>
								{/if}
							</div>
						{/if}
					</div>
				</section>
				{/if}

				{#if wizardSectionVisible('answer') && objectiveType(fQuestionType)}
					<section class="space-y-3">
						<div class="flex items-center justify-between gap-3">
							<h2 class="text-sm font-semibold uppercase tracking-[0.18em] text-emerald-800">Opsi Jawaban</h2>
							<div class="flex gap-2">
								{#if fQuestionType === 'true_false'}
									<Button variant="outline" size="sm" onclick={setTrueFalseTemplate}>Set Benar / Salah</Button>
								{/if}
								{#if fQuestionType !== 'true_false'}
									<Button variant="outline" size="sm" onclick={addOption}>+ Opsi</Button>
								{/if}
							</div>
						</div>
						<div class="space-y-3">
							{#each fOptions as option, idx (option.label + idx)}
								<div class="rounded-lg border bg-slate-50/60 p-3">
									<div class="mb-2 flex items-center justify-between gap-3">
										<div class="text-sm font-semibold text-slate-800">Opsi {option.label}</div>
										{#if fQuestionType !== 'true_false'}
											<Button variant="outline" size="sm" onclick={() => removeOption(idx)}>Hapus</Button>
										{/if}
									</div>
									<div class="grid gap-3 md:grid-cols-2">
										<div>
											<label for={`opt-text-${idx}`} class="mb-1 block text-xs font-medium text-slate-600">Teks Opsi</label>
											<Input id={`opt-text-${idx}`} bind:value={option.text} />
										</div>
										{#if isAdvanceMode}
											<div>
												<label for={`opt-latex-${idx}`} class="mb-1 block text-xs font-medium text-slate-600">LaTeX Opsi</label>
												<Input id={`opt-latex-${idx}`} bind:value={option.latex} />
											</div>
											<div class="md:col-span-2">
												<label for={`opt-html-${idx}`} class="mb-1 block text-xs font-medium text-slate-600">HTML Opsi</label>
												<Textarea id={`opt-html-${idx}`} rows={2} bind:value={option.html} />
											</div>
										{/if}
									</div>
								</div>
							{/each}
						</div>
					</section>
				{/if}

				{#if wizardSectionVisible('answer') && requiresAnswerKey(fQuestionType)}
					<section class="space-y-3">
						<h2 class="text-sm font-semibold uppercase tracking-[0.18em] text-emerald-800">Kunci dan Pembahasan</h2>
						<div class="grid gap-3 lg:grid-cols-2">
							<div>
								<label for="q-answer" class="mb-1 block text-xs font-medium text-slate-600">Kunci Jawaban</label>
								<Input id="q-answer" placeholder={fQuestionType === 'short_answer' ? 'jawaban singkat yang diharapkan' : 'A atau A,C'} bind:value={fAnswerKey} />
							</div>
							<div>
								<label for="q-explain" class="mb-1 block text-xs font-medium text-slate-600">Pembahasan Teks Polos</label>
								<Input id="q-explain" bind:value={fExplanation} />
							</div>
						</div>
						{#if isAdvanceMode}
							<div>
								<label for="q-explain-html" class="mb-1 block text-xs font-medium text-slate-600">Pembahasan HTML</label>
								<TiptapEditor
									id="q-explain-html"
									minRows={4}
									placeholder="Pembahasan lengkap dengan langkah-langkah penyelesaian..."
									onImageUpload={uploadImageInEditor}
									bind:value={fExplanationHTML}
								/>
							</div>
						{/if}
					</section>
				{/if}

				{#if wizardSectionVisible('answer') && fQuestionType === 'essay'}
					<section class="space-y-3">
						<h2 class="text-sm font-semibold uppercase tracking-[0.18em] text-emerald-800">Rubrik Essay</h2>
						{#if isAdvanceMode}
							<div>
								<label for="q-rubric" class="mb-1 block text-xs font-medium text-slate-600">Rubrik HTML</label>
								<TiptapEditor
									id="q-rubric"
									minRows={5}
									placeholder="Kriteria penilaian, bobot skor tiap aspek, dan panduan mengoreksi..."
									onImageUpload={uploadImageInEditor}
									bind:value={fRubricHTML}
								/>
							</div>
						{:else}
							<div class="rounded-lg border border-dashed border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-900">
								Rubrik essay tidak wajib di mode beginner. Setelah soal tersimpan, buka mode advance untuk melengkapi rubrik penilaian jika diperlukan.
							</div>
						{/if}
					</section>
				{/if}

				{#if wizardSectionVisible('review') && isAdvanceMode}
				<section class="space-y-3">
					<div class="flex items-center justify-between gap-3">
						<h2 class="text-sm font-semibold uppercase tracking-[0.18em] text-emerald-800">Asset dan Lampiran</h2>
						<Badge variant="outline">{fMediaAssetIds.length} asset direferensikan</Badge>
					</div>
					<div class="grid gap-3 lg:grid-cols-[1fr_auto_auto] lg:items-end">
						<div>
							<label for="asset-file" class="mb-1 block text-xs font-medium text-slate-600">Upload gambar / audio / PDF</label>
							<input id="asset-file" type="file" accept="image/*,audio/*,application/pdf" onchange={(e) => (assetFile = (e.currentTarget as HTMLInputElement).files?.[0] ?? null)} class="block w-full rounded-md border border-input bg-background px-3 py-2 text-sm text-slate-700" />
						</div>
						<div>
							<label for="asset-purpose" class="mb-1 block text-xs font-medium text-slate-600">Peruntukan</label>
							<select id="asset-purpose" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={assetPurpose}>
								<option value="general">Umum</option>
								<option value="stimulus">Stimulus</option>
								<option value="option">Opsi</option>
								<option value="explanation">Pembahasan</option>
								<option value="rubric">Rubrik</option>
							</select>
						</div>
						<Button onclick={uploadAsset} disabled={assetBusy || !assetFile}>{assetBusy ? 'Uploading...' : 'Upload'}</Button>
					</div>
					{#if uploadedAssets.length > 0}
						<div class="space-y-2">
							<p class="text-xs uppercase tracking-[0.18em] text-slate-500">Asset reusable untuk item ini</p>
							{#each uploadedAssets as asset (asset.id)}
								<div class="flex flex-wrap items-center justify-between gap-3 rounded-lg border bg-slate-50 px-3 py-2">
									<div class="min-w-0">
										<p class="truncate text-sm font-medium text-slate-900">{asset.original_name}</p>
										<p class="text-xs text-slate-500">{asset.mime_type} · {asset.purpose}</p>
									</div>
									<div class="flex flex-wrap gap-2">
										<Button variant="outline" size="sm" onclick={() => insertAsset(`<img src="${asset.url}" alt="${asset.original_name}" />`, 'stimulus')}>Ke Stimulus</Button>
										<Button variant="outline" size="sm" onclick={() => insertAsset(`<img src="${asset.url}" alt="${asset.original_name}" />`, 'stem')}>Ke Stem</Button>
										<Button variant="outline" size="sm" onclick={() => insertAsset(`<a href="${asset.url}" target="_blank" rel="noreferrer">${asset.original_name}</a>`, 'explanation')}>Ke Pembahasan</Button>
									</div>
								</div>
							{/each}
						</div>
					{/if}
				</section>
				{:else if wizardSectionVisible('review')}
				<section class="space-y-3">
					<h2 class="text-sm font-semibold uppercase tracking-[0.18em] text-emerald-800">Gambar Sederhana</h2>
					<div class="rounded-lg border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-900">
						Di mode beginner, gunakan tombol gambar pada editor untuk menyisipkan ilustrasi langsung ke soal. Asset reuse, PDF, dan pengelolaan lampiran lanjutan tersedia di mode advance.
					</div>
				</section>
				{/if}

				{#if wizardSectionVisible('review') && isAdvanceMode}
				<section class="space-y-3">
					<h2 class="text-sm font-semibold uppercase tracking-[0.18em] text-emerald-800">Catatan Penulis dan Reviewer</h2>
					<div class="grid gap-3 lg:grid-cols-2">
						<div>
							<label for="q-writer-notes" class="mb-1 block text-xs font-medium text-slate-600">Catatan Penulis</label>
							<Textarea id="q-writer-notes" rows={3} bind:value={fWriterNotes} />
						</div>
						<div>
							<label for="q-review-notes" class="mb-1 block text-xs font-medium text-slate-600">Catatan Review</label>
							<Textarea id="q-review-notes" rows={3} bind:value={fReviewNotes} />
						</div>
					</div>
				</section>
				{/if}

				<section class="space-y-3">
					<h2 class="text-sm font-semibold uppercase tracking-[0.18em] text-emerald-800">Preview</h2>
					{#if isDocumentVariant}
						<div class="rounded-lg border border-sky-200 bg-sky-50 px-4 py-3 text-sm text-sky-900">
							Mode document memprioritaskan rasa menulis. Gunakan area preview ini untuk mengecek apakah narasi, stimulus, dan tampilan peserta sudah enak dibaca.
						</div>
					{/if}
					<div class="rounded-xl border border-emerald-200 bg-white p-4 shadow-sm">
						<div class="mb-3 flex flex-wrap gap-2">
							<Badge class={workflowBadgeClass(fWorkflowStatus)}>{workflowLabel[fWorkflowStatus] ?? fWorkflowStatus}</Badge>
							<Badge class={publicationBadgeClass(fStatus)}>{statusLabel[fStatus] ?? fStatus}</Badge>
							<Badge class={difficultyBadgeClass(fDifficulty)}>{difficultyLabel[fDifficulty] ?? fDifficulty}</Badge>
							<Badge variant="outline">{questionTypeLabel[fQuestionType] ?? fQuestionType}</Badge>
							{#if fHotsFlag}
								<Badge class="border-emerald-200 bg-emerald-100 text-emerald-800">HOTS</Badge>
							{/if}
						</div>
						{#if fStimulusHTML}
							<div class="prose prose-sm max-w-none rounded-lg border bg-slate-50 p-3">{@html fStimulusHTML}</div>
						{/if}
						{#if fStimulusLatex}
							<LatexBlock src={fStimulusLatex} class="mt-3" />
						{/if}
						{#if fStemHTML}
							<div class="prose prose-sm mt-4 max-w-none text-slate-900">{@html fStemHTML}</div>
						{:else}
							<p class="mt-4 text-sm text-slate-900">{fQuestionText}</p>
						{/if}
						{#if fStemLatex}
							<LatexBlock src={fStemLatex} class="mt-3" />
						{/if}
						{#if objectiveType(fQuestionType)}
							<div class="mt-4 space-y-2">
								{#each fOptions as option (option.label)}
									<div class="rounded-lg border px-3 py-2 text-sm">
										<span class="mr-2 font-semibold text-emerald-800">{option.label}.</span>
										{#if option.html}
											<span>{@html option.html}</span>
										{:else if option.latex}
											<LatexBlock src={option.latex} display={false} class="inline-block" />
										{:else}
											<span>{option.text || '—'}</span>
										{/if}
									</div>
								{/each}
							</div>
						{/if}
					</div>
				</section>

				<div class="flex flex-wrap gap-2 border-t pt-4">
					{#if isWizardVariant}
						<Button variant="outline" onclick={previousWizardStep} disabled={wizardStep <= 1}>Langkah Sebelumnya</Button>
						<Button variant="outline" onclick={nextWizardStep} disabled={wizardStep >= 4}>Langkah Berikutnya</Button>
					{/if}
					<Button onclick={saveQuestion} disabled={fBusy}>{fBusy ? 'Menyimpan...' : editId ? 'Perbarui Item' : 'Simpan Item'}</Button>
					<Button variant="outline" onclick={resetForm}>Batal</Button>
				</div>
			</Card.Content>
		</Card.Root>
	{/if}

	{#if loading}
		<p class="text-sm text-slate-600">Memuat data bank soal...</p>
	{:else}
		<Card.Root class={isGridVariant ? 'border-slate-200 shadow-sm' : ''}>
			<Card.Header class="pb-3">
				<div class="flex flex-wrap items-center justify-between gap-3">
					<div>
						<Card.Title class="text-lg">Katalog Item</Card.Title>
						<Card.Description>
							{#if isGridVariant}
								Mode grid menekankan scanning cepat. Menampilkan {rangeStart}-{rangeEnd} dari {totalItems} item dengan fokus operasional.
							{:else if isReviewVariant}
								Mode review menonjolkan antrian dan status agar reviewer bisa bergerak cepat. Menampilkan {rangeStart}-{rangeEnd} dari {totalItems} item.
							{:else if isDocumentVariant}
								Mode document tetap memakai katalog yang sama, tetapi editor dan preview ditujukan untuk penulisan naratif. Menampilkan {rangeStart}-{rangeEnd} dari {totalItems} item.
							{:else if isPackageVariant}
								Mode package-fit membantu melihat item yang siap dibawa ke paket ujian. Menampilkan {rangeStart}-{rangeEnd} dari {totalItems} item.
							{:else}
								Menampilkan {rangeStart}-{rangeEnd} dari {totalItems} item. Filter dan pencarian sudah dipindah ke backend agar ringan saat bank soal membesar.
							{/if}
						</Card.Description>
					</div>
					<div class="flex flex-wrap gap-2">
						<Button variant="outline" onclick={resetFilters}>Reset</Button>
						<Button onclick={applyFilters}>Terapkan Filter</Button>
					</div>
				</div>
				<div class="grid gap-3 pt-3 lg:grid-cols-[minmax(0,2fr),repeat(4,minmax(0,1fr))]">
					<Input placeholder="Cari kode, ringkasan, topik, CP, atau KD..." bind:value={search} onkeydown={handleSearchKeydown} />
					<select id="filter-subject" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={filterSubject}>
						<option value="">Semua mapel</option>
						{#each subjects as subject (subject.id)}
							<option value={subject.id}>{subject.code}</option>
						{/each}
					</select>
					<select id="filter-workflow" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={filterWorkflow}>
						<option value="">Semua workflow</option>
						<option value="draft">Draft</option>
						<option value="review">Review</option>
						<option value="approved">Approved</option>
						<option value="rejected">Revisi</option>
					</select>
					<select id="filter-type" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={filterType}>
						<option value="">Semua tipe</option>
						<option value="multiple_choice">Pilihan Ganda</option>
						<option value="multiple_answer">Jawaban Ganda</option>
						<option value="true_false">Benar/Salah</option>
						<option value="short_answer">Isian Singkat</option>
						<option value="essay">Essay</option>
					</select>
					<select id="filter-hots" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={filterHots}>
						<option value="">Semua level</option>
						<option value="yes">HOTS</option>
						<option value="no">Non-HOTS</option>
					</select>
				</div>
			</Card.Header>
			<Card.Content class={`overflow-x-auto p-0 ${isGridVariant ? 'bg-slate-50/60' : ''}`}>
				<Table.Root>
					<Table.Header>
						<Table.Row>
							<Table.Head class="min-w-28">Kode</Table.Head>
							<Table.Head class="min-w-[22rem]">Item</Table.Head>
							<Table.Head>Mapel</Table.Head>
							<Table.Head>Tipe</Table.Head>
							<Table.Head>Mutu</Table.Head>
							<Table.Head>Status</Table.Head>
							<Table.Head class="text-right">Aksi</Table.Head>
						</Table.Row>
					</Table.Header>
					<Table.Body>
						{#each questions as q (q.id)}
							<Table.Row class={isGridVariant ? 'bg-white' : ''}>
								<Table.Cell class="align-top">
									<div class="font-mono text-xs text-slate-600">{q.code || '—'}</div>
									<div class="mt-1 text-xs text-slate-500">v{q.version} · Kelas {q.grade_level ?? '—'}</div>
								</Table.Cell>
								<Table.Cell class="align-top">
									<div class="space-y-1">
										<p class="line-clamp-2 text-sm font-medium text-slate-900">{q.question_text || 'Tanpa ringkasan teks'}</p>
										<div class="flex flex-wrap gap-2 text-xs text-slate-500">
											{#if q.material_topic}<span>{q.material_topic}</span>{/if}
											{#if q.cp_ref}<span>CP: {q.cp_ref}</span>{/if}
											{#if q.kd_ref}<span>KD: {q.kd_ref}</span>{/if}
										</div>
									</div>
								</Table.Cell>
								<Table.Cell class="align-top">
									<Badge variant="outline">{q.subject_code}</Badge>
								</Table.Cell>
								<Table.Cell class="align-top">
									<div class="space-y-1">
										<Badge variant="outline">{questionTypeLabel[q.question_type] ?? q.question_type}</Badge>
										<div class="text-xs text-slate-500">{q.options?.length ?? 0} opsi</div>
									</div>
								</Table.Cell>
								<Table.Cell class="align-top">
									<div class="space-y-1">
										<Badge class={difficultyBadgeClass(q.difficulty)}>{difficultyLabel[q.difficulty] ?? q.difficulty}</Badge>
										{#if q.hots_flag}
											<Badge class="border-emerald-200 bg-emerald-100 text-emerald-800">HOTS</Badge>
										{/if}
									</div>
								</Table.Cell>
								<Table.Cell class="align-top">
									<div class="space-y-1">
										<Badge class={workflowBadgeClass(q.workflow_status)}>{workflowLabel[q.workflow_status] ?? q.workflow_status}</Badge>
										<Badge class={publicationBadgeClass(q.status)}>{statusLabel[q.status] ?? q.status}</Badge>
										{#if q.suggested_mode === 'beginner'}
											<Badge variant="outline">Beginner</Badge>
										{/if}
									</div>
								</Table.Cell>
								<Table.Cell class="align-top">
									<div class="flex flex-wrap justify-end gap-2">
										<Button size="sm" variant="outline" onclick={() => openDetail(q.id)} disabled={detailBusy && selectedDetail?.id === q.id}>Detail</Button>
										<Button size="sm" variant="outline" onclick={() => duplicateQuestion(q.id)}>Duplikat</Button>
										{#if q.suggested_mode !== 'advance'}
											<Button size="sm" variant="outline" onclick={() => openAdvanceCompletion(q)}>Lengkapi</Button>
										{/if}
										{#if canSubmitReview && (q.workflow_status === 'draft' || q.workflow_status === 'rejected')}
											<Button size="sm" variant="outline" onclick={() => workflowAction(q.id, 'submit_review')}>Review</Button>
										{/if}
										{#if canApproveWorkflow && q.workflow_status === 'review'}
											<Button size="sm" variant="outline" onclick={() => workflowAction(q.id, 'approve')}>Approve</Button>
										{/if}
										{#if canApproveWorkflow && q.workflow_status === 'approved' && q.status !== 'published'}
											<Button size="sm" variant="outline" onclick={() => workflowAction(q.id, 'publish')}>Publish</Button>
										{/if}
										{#if canApproveWorkflow && q.status === 'published'}
											<Button size="sm" variant="outline" onclick={() => workflowAction(q.id, 'archive')}>Arsipkan</Button>
										{/if}
										<Button size="sm" variant="outline" onclick={() => openEdit(q)}>Edit</Button>
										<Button size="sm" variant="destructive" onclick={() => deleteQuestion(q.id)}>Hapus</Button>
									</div>
								</Table.Cell>
							</Table.Row>
						{/each}
						{#if questions.length === 0}
							<Table.Row>
								<Table.Cell colspan={7} class="py-10 text-center text-sm text-slate-500">Belum ada item yang cocok dengan filter.</Table.Cell>
							</Table.Row>
						{/if}
					</Table.Body>
				</Table.Root>
			</Card.Content>
			<div class="flex flex-wrap items-center justify-between gap-3 border-t px-6 py-4 text-sm text-slate-600">
				<p>Halaman {currentPage} dari {pageCount}</p>
				<div class="flex gap-2">
					<Button variant="outline" size="sm" onclick={() => goToPage(currentPage - 1)} disabled={currentPage <= 1}>Sebelumnya</Button>
					<Button variant="outline" size="sm" onclick={() => goToPage(currentPage + 1)} disabled={currentPage >= pageCount}>Berikutnya</Button>
				</div>
			</div>
		</Card.Root>
	{/if}
</div>
