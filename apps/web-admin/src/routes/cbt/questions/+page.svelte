<script lang="ts">
	import { onMount } from 'svelte';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Textarea } from '$lib/components/ui/textarea';
	import { Badge } from '$lib/components/ui/badge';
	import { toast } from '$lib/components/ui/sonner';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { page } from '$app/state';
	import EditorWrapper from '$lib/components/EditorWrapper.svelte';
	import LatexBlock from '$lib/components/LatexBlock.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import EmptyStatePanel from '$lib/components/EmptyStatePanel.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';

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
	let pageCount = $derived(Math.max(1, Math.ceil(totalItems / pageSize)));
	let rangeStart = $derived(totalItems === 0 ? 0 : (currentPage - 1) * pageSize + 1);
	let rangeEnd = $derived(Math.min(totalItems, currentPage * pageSize));
	let qualityWarnings = $derived.by(() => {
		const warnings: string[] = [];
		if (!isAdvanceMode) {
			if (fQuestionType === 'essay' && !fRubricHTML.trim()) warnings.push('Rubrik essay belum diisi. Bisa dilengkapi nanti di mode advance.');
			if (!fMaterialTopic.trim()) warnings.push('Topik materi belum diisi. Bisa dilengkapi nanti di mode advance.');
			return warnings;
		}
		if (!fMaterialTopic.trim()) warnings.push('Topik materi belum diisi. Penting untuk blueprint paket ujian.');
		if (!fCPRef.trim() || !fKDRef.trim()) warnings.push('Acuan kurikulum belum lengkap. Minimal CP dan KD sebaiknya terisi.');
		if (fQuestionType === 'essay' && !fRubricHTML.trim()) warnings.push('Soal essay sebaiknya memiliki rubrik HTML sebelum diajukan review.');
		if (fStatus === 'published' && fWorkflowStatus !== 'approved') warnings.push('Item published idealnya berstatus workflow approved.');
		if (fHotsFlag && !fCognitiveLevel.trim()) warnings.push('Item HOTS sebaiknya disertai level kognitif.');
		if (objectiveType(fQuestionType) && fOptions.some((o) => !o.text?.trim() && !o.html?.trim() && !o.latex?.trim())) {
			warnings.push('Masih ada opsi objektif yang kosong.');
		}
		if (requiresAnswerKey(fQuestionType) && !fAnswerKey.trim()) warnings.push('Kunci jawaban belum diisi.');
		return warnings;
	});

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

	function showToast(message: string) { toast.success(message); }
	function showError(message: string) { toast.error(message); }

	function detectMode(q: Question): 'beginner' | 'advance' {
		if (q.authoring_mode === 'advance' || q.suggested_mode === 'advance') return 'advance';
		if (q.stem_latex || q.stimulus_latex || q.cp_ref || q.kd_ref || q.indicator_ref || q.academic_phase || q.hots_flag || (q.media_asset_ids?.length > 0)) return 'advance';
		return 'beginner';
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
			if (!qRes.ok || qJson.error) { error = qJson.error ?? 'Gagal memuat bank soal'; return; }
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
			if (!res.ok) { showError(payload.error ?? 'Gagal memuat detail soal'); return null; }
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
			if (!res.ok) { showError((payload as { error?: string }).error ?? 'Gagal memuat asset'); return; }
			uploadedAssets = Array.isArray(payload) ? payload : [];
		} catch (err) {
			showError((err as Error).message || 'Gagal memuat asset');
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
		fAuthoringMode = detectMode(question);
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
		resetForm();
		showForm = true;
		selectedDetail = null;
	}

	async function openEdit(question: Question, forceMode?: 'beginner' | 'advance') {
		const detail = (await loadDetail(question.id)) ?? question;
		applyQuestionToForm(detail);
		if (forceMode) fAuthoringMode = forceMode;
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
		fOptions = [{ label: 'A', text: 'Benar' }, { label: 'B', text: 'Salah' }];
		fAnswerKey = 'A';
	}

	$effect(() => {
		const type = fQuestionType;
		if (!isAdvanceMode && type !== 'multiple_choice' && type !== 'essay') {
			fQuestionType = 'multiple_choice';
		}
		if (type === 'true_false') {
			setTrueFalseTemplate();
		} else if (objectiveType(type) && fOptions.length < 4) {
			fOptions = defaultOptions();
		}
		if (type === 'essay') fAnswerKey = '';
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
			if (!res.ok) { showError(payload.error ?? 'Upload asset gagal'); return; }
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
		if (!fSubjectId || (!fQuestionText.trim() && !fStemHTML && !fStemLatex)) {
			showError('Mapel dan isi soal utama wajib diisi');
			return;
		}
		if (!isAdvanceMode && fQuestionType === 'multiple_choice' && fOptions.filter((o) => o.text?.trim() || o.html?.trim() || o.latex?.trim()).length < 4) {
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
			if (!res.ok) { showError(payload.error ?? 'Gagal menyimpan soal'); return; }
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
		if (!res.ok) { showError(payload.error ?? 'Gagal menggandakan item'); return; }
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
		if (!res.ok) { showError(payload.error ?? 'Aksi workflow gagal'); return; }
		showToast('Status item diperbarui');
		await load(currentPage);
		if (selectedDetail?.id === id) await loadDetail(id);
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

	function clearQuestionFilters() {
		void resetFilters();
	}

	async function goToPage(p: number) {
		if (p < 1 || p > pageCount || p === currentPage) return;
		await load(p);
	}

	function handleSearchKeydown(event: KeyboardEvent) {
		if (event.key === 'Enter') { event.preventDefault(); void applyFilters(); }
	}

	onMount(() => { void load(); });
</script>

<svelte:head>
	<title>Bank Soal CBT — MTsN 2 Kolaka Utara</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex flex-wrap items-start justify-between gap-4">
		<div class="space-y-1">
			<h1 class="text-2xl font-semibold tracking-tight text-slate-900">Bank Soal CBT</h1>
			<p class="max-w-3xl text-sm text-slate-600">
				Bangun item asesmen dengan rich text, LaTeX, stimulus, workflow review, dan metadata kurikulum. Mode <strong>beginner</strong> untuk input cepat, mode <strong>advance</strong> untuk metadata blueprint lengkap.
			</p>
		</div>
		<Button onclick={openCreate}>{showForm ? 'Tutup Form' : '+ Tambah Item'}</Button>
	</div>

	<div class="grid gap-3 md:grid-cols-4">
		<div class="rounded-2xl border border-emerald-100 bg-emerald-50 px-4 py-4">
			<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-emerald-700">Total Item</p>
			<p class="mt-2 text-2xl font-semibold text-slate-900">{totalItems}</p>
			<p class="text-sm text-slate-600">soal terarsip, draft, atau published dalam bank soal</p>
		</div>
		<div class="rounded-2xl border border-amber-100 bg-amber-50 px-4 py-4">
			<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-amber-700">Perlu Review</p>
			<p class="mt-2 text-2xl font-semibold text-slate-900">{questions.filter((item) => item.workflow_status === 'review').length}</p>
			<p class="text-sm text-slate-600">item pada halaman aktif yang menunggu approval</p>
		</div>
		<div class="rounded-2xl border border-sky-100 bg-sky-50 px-4 py-4">
			<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-sky-700">Mode Beginner</p>
			<p class="mt-2 text-2xl font-semibold text-slate-900">{questions.filter((item) => detectMode(item) === 'beginner').length}</p>
			<p class="text-sm text-slate-600">item cepat yang masih bisa disempurnakan di advance</p>
		</div>
		<div class="rounded-2xl border border-violet-100 bg-violet-50 px-4 py-4">
			<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-violet-700">Published</p>
			<p class="mt-2 text-2xl font-semibold text-slate-900">{questions.filter((item) => item.status === 'published').length}</p>
			<p class="text-sm text-slate-600">item siap dipakai dari halaman hasil filter saat ini</p>
		</div>
	</div>

	{#if error}
		<RecoveryPanel message={error} onRetry={() => load(currentPage)} />
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
						{#if detectMode(selectedDetail) !== 'advance'}
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
				<!-- Student preview -->
				<div class="rounded-2xl border border-slate-200 bg-slate-950 p-5 text-slate-50 shadow-sm">
					<div class="flex flex-wrap items-center justify-between gap-3 border-b border-slate-800 pb-3">
						<div>
							<p class="text-xs font-semibold uppercase tracking-[0.22em] text-emerald-300">Preview Peserta</p>
							<p class="mt-1 text-sm text-slate-300">Simulasi tampilan soal saat dibuka oleh siswa.</p>
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
						Mode advance: rich text, LaTeX, stimulus, metadata kurikulum CP/TP/KD, asset, dan workflow review.
					{:else}
						Mode beginner: isi soal, opsi, kunci jawaban. Metadata kurikulum dan fitur lanjutan bisa dilengkapi nanti di mode advance.
					{/if}
				</Card.Description>
			</Card.Header>
			<Card.Content class="space-y-6 pt-6">
				<!-- Mode selector -->
				<section class="space-y-3">
					<h2 class="text-sm font-semibold uppercase tracking-[0.18em] text-emerald-800">Mode Authoring</h2>
					<div class="grid gap-3 md:grid-cols-2">
						<button
							type="button"
							class={`rounded-xl border p-4 text-left transition ${!isAdvanceMode ? 'border-emerald-300 bg-emerald-50 shadow-sm' : 'border-slate-200 bg-white hover:border-emerald-200'}`}
							onclick={() => (fAuthoringMode = 'beginner')}
						>
							<p class="text-sm font-semibold text-slate-900">Beginner</p>
							<p class="mt-1 text-sm text-slate-600">Fokus pada mapel, isi soal, opsi, dan jawaban. Cocok untuk input cepat tanpa harus isi metadata.</p>
						</button>
						<button
							type="button"
							class={`rounded-xl border p-4 text-left transition ${isAdvanceMode ? 'border-emerald-300 bg-emerald-50 shadow-sm' : 'border-slate-200 bg-white hover:border-emerald-200'}`}
							onclick={() => (fAuthoringMode = 'advance')}
						>
							<p class="text-sm font-semibold text-slate-900">Advance</p>
							<p class="mt-1 text-sm text-slate-600">Buka seluruh fitur: blueprint kurikulum, workflow review, asset reuse, rich text, dan LaTeX.</p>
						</button>
					</div>
					{#if !isAdvanceMode}
						<div class="rounded-lg border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-900">
							Soal mode beginner disimpan otomatis sebagai <strong>draft</strong>. KD, CP, TP, HOTS, dan metadata blueprint bisa dilengkapi nanti di mode advance.
						</div>
						<div class="flex flex-wrap gap-2">
							<Button variant="outline" size="sm" onclick={() => (fAuthoringMode = 'advance')}>Lengkapi di Advanced</Button>
							{#if canSubmitReview && editId}
								<Button variant="outline" size="sm" onclick={async () => { await workflowAction(editId!, 'submit_review'); }}>Ajukan Review</Button>
							{/if}
						</div>
					{/if}
				</section>

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

				<!-- Identitas -->
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
						{#if isAdvanceMode}
							<div>
								<label for="q-code" class="mb-1 block text-xs font-medium text-slate-600">Kode Item</label>
								<Input id="q-code" placeholder="MTK-VII-HOTS-001" bind:value={fCode} />
							</div>
						{/if}
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
							<label for="q-grade" class="mb-1 block text-xs font-medium text-slate-600">Tingkat / Kelas</label>
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
					{/if}
				</section>

				<!-- Konten soal -->
				<section class="space-y-3">
					<h2 class="text-sm font-semibold uppercase tracking-[0.18em] text-emerald-800">Konten Soal</h2>
					<div>
						<label for="q-plain" class="mb-1 block text-xs font-medium text-slate-600">
							{isAdvanceMode ? 'Ringkasan Teks Polos / Fallback' : 'Ringkasan Teks Polos'}
						</label>
						<Textarea id="q-plain" rows={2} placeholder="Teks ringkas untuk pencarian, fallback, dan kompatibilitas." bind:value={fQuestionText} />
					</div>
					<div class="space-y-4">
						{#if isAdvanceMode}
							<div>
								<label for="q-stimulus-html" class="mb-1 block text-xs font-medium text-slate-600">
									Stimulus <span class="font-normal text-slate-400">(narasi, tabel, gambar, atau rumus pendukung)</span>
								</label>
								<EditorWrapper id="q-stimulus-html" minRows={5} placeholder="Stimulus, narasi, tabel, atau gambar pendukung..." onImageUpload={uploadImageInEditor} bind:value={fStimulusHTML} />
							</div>
							<div>
								<label for="q-stimulus-latex" class="mb-1 block text-xs font-medium text-slate-600">
									Stimulus LaTeX <span class="font-normal text-slate-400">(opsional)</span>
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
							<EditorWrapper
								id="q-stem-html"
								minRows={isAdvanceMode ? 6 : 5}
								placeholder={isAdvanceMode ? 'Tuliskan pokok soal utama. Gunakan toolbar untuk format teks, rumus LaTeX (∑), atau sisipkan gambar.' : 'Tuliskan soal. Gunakan toolbar editor untuk menyisipkan gambar, tabel, atau format teks.'}
								onImageUpload={uploadImageInEditor}
								bind:value={fStemHTML}
							/>
						</div>
						{#if isAdvanceMode}
							<div>
								<label for="q-stem-latex" class="mb-1 block text-xs font-medium text-slate-600">
									Stem LaTeX <span class="font-normal text-slate-400">(opsional)</span>
								</label>
								<Textarea id="q-stem-latex" rows={2} placeholder="x^2 + 5x + 6 = 0" bind:value={fStemLatex} class="font-mono text-sm" />
								{#if fStemLatex}
									<div class="mt-2 rounded-lg border bg-slate-50 p-3"><LatexBlock src={fStemLatex} /></div>
								{/if}
							</div>
						{/if}
					</div>
				</section>

				<!-- Opsi jawaban -->
				{#if objectiveType(fQuestionType)}
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
									<div class="grid gap-3 {isAdvanceMode ? 'md:grid-cols-2' : ''}">
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

				<!-- Kunci dan pembahasan -->
				{#if requiresAnswerKey(fQuestionType)}
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
								<EditorWrapper id="q-explain-html" minRows={4} placeholder="Pembahasan lengkap dengan langkah-langkah penyelesaian..." onImageUpload={uploadImageInEditor} bind:value={fExplanationHTML} />
							</div>
						{/if}
					</section>
				{/if}

				<!-- Rubrik essay -->
				{#if fQuestionType === 'essay'}
					<section class="space-y-3">
						<h2 class="text-sm font-semibold uppercase tracking-[0.18em] text-emerald-800">Rubrik Essay</h2>
						{#if isAdvanceMode}
							<div>
								<label for="q-rubric" class="mb-1 block text-xs font-medium text-slate-600">Rubrik HTML</label>
								<EditorWrapper id="q-rubric" minRows={5} placeholder="Kriteria penilaian, bobot skor tiap aspek, dan panduan mengoreksi..." onImageUpload={uploadImageInEditor} bind:value={fRubricHTML} />
							</div>
						{:else}
							<div class="rounded-lg border border-dashed border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-900">
								Rubrik essay tidak wajib di mode beginner. Lengkapi nanti di mode advance setelah soal tersimpan.
							</div>
						{/if}
					</section>
				{/if}

				<!-- Asset dan lampiran (advance only) -->
				{#if isAdvanceMode}
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
							<LoadingButton onclick={uploadAsset} loading={assetBusy} loadingLabel="Uploading..." disabled={!assetFile} label="Upload" />
						</div>
						{#if uploadedAssets.length > 0}
							<div class="space-y-2">
								<p class="text-xs uppercase tracking-[0.18em] text-slate-500">Asset untuk item ini</p>
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
				{:else}
					<section class="space-y-3">
						<h2 class="text-sm font-semibold uppercase tracking-[0.18em] text-emerald-800">Gambar Sederhana</h2>
						<div class="rounded-lg border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-900">
							Di mode beginner, gunakan tombol gambar pada editor untuk menyisipkan gambar langsung ke soal. Asset reuse, PDF, dan pengelolaan lampiran lanjutan tersedia di mode advance.
						</div>
					</section>
				{/if}

				<!-- Preview -->
				<section class="space-y-3">
					<h2 class="text-sm font-semibold uppercase tracking-[0.18em] text-emerald-800">Preview</h2>
					<div class="rounded-xl border border-emerald-200 bg-white p-4 shadow-sm">
						<div class="mb-3 flex flex-wrap gap-2">
							<Badge class={workflowBadgeClass(fWorkflowStatus)}>{workflowLabel[fWorkflowStatus] ?? fWorkflowStatus}</Badge>
							<Badge class={publicationBadgeClass(fStatus)}>{statusLabel[fStatus] ?? fStatus}</Badge>
							<Badge class={difficultyBadgeClass(fDifficulty)}>{difficultyLabel[fDifficulty] ?? fDifficulty}</Badge>
							<Badge variant="outline">{questionTypeLabel[fQuestionType] ?? fQuestionType}</Badge>
							{#if fHotsFlag}<Badge class="border-emerald-200 bg-emerald-100 text-emerald-800">HOTS</Badge>{/if}
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
					<LoadingButton onclick={saveQuestion} loading={fBusy} loadingLabel="Menyimpan..." label={editId ? 'Perbarui Item' : 'Simpan Item'} />
					<Button variant="outline" onclick={resetForm}>Batal</Button>
				</div>
			</Card.Content>
		</Card.Root>
	{/if}

	{#if loading}
		<div class="space-y-4">
			<div class="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
				<div class="space-y-2">
					<Skeleton class="h-8 w-44" />
					<Skeleton class="h-4 w-80" />
				</div>
				<div class="flex gap-2">
					<Skeleton class="h-9 w-28" />
					<Skeleton class="h-9 w-36" />
				</div>
			</div>
			<div class="grid gap-4 lg:grid-cols-[0.95fr_1.05fr]">
				<div class="space-y-3">
					<Skeleton class="h-10 w-full" />
					<Skeleton class="h-10 w-full" />
					<Skeleton class="h-10 w-full" />
					<Skeleton class="h-56 w-full" />
				</div>
				<div class="space-y-3">
					<Skeleton class="h-12 w-full" />
					<Skeleton class="h-16 w-full" />
					<Skeleton class="h-16 w-full" />
					<Skeleton class="h-16 w-full" />
				</div>
			</div>
		</div>
	{:else}
		<Card.Root>
			<Card.Header class="pb-3">
				<div class="flex flex-wrap items-center justify-between gap-3">
					<div>
						<Card.Title class="text-lg">Katalog Item</Card.Title>
						<Card.Description>
							Menampilkan {rangeStart}–{rangeEnd} dari {totalItems} item.
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
			<Card.Content class="overflow-x-auto p-0">
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
							<Table.Row>
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
									</div>
								</Table.Cell>
								<Table.Cell class="align-top">
									<div class="flex flex-wrap justify-end gap-1">
										<Button size="sm" variant="outline" onclick={() => openDetail(q.id)} disabled={detailBusy && selectedDetail?.id === q.id}>Detail</Button>
										<Button size="sm" variant="outline" onclick={() => openEdit(q)}>Edit</Button>
										<Button size="sm" variant="outline" onclick={() => duplicateQuestion(q.id)}>Duplikat</Button>
										{#if detectMode(q) !== 'advance'}
											<Button size="sm" variant="outline" onclick={() => openEdit(q, 'advance')}>Lengkapi</Button>
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
											<Button size="sm" variant="outline" onclick={() => workflowAction(q.id, 'archive')}>Arsip</Button>
										{/if}
										<Button size="sm" variant="destructive" onclick={() => deleteQuestion(q.id)}>Hapus</Button>
									</div>
								</Table.Cell>
							</Table.Row>
						{/each}
						{#if questions.length === 0}
							<Table.Row>
								<Table.Cell colspan={7} class="p-4">
									<EmptyStatePanel
										compact
										eyebrow={search || filterSubject || filterWorkflow || filterType || filterHots ? 'Filter Tidak Menemukan Hasil' : 'Mulai Bank Soal'}
										title={search || filterSubject || filterWorkflow || filterType || filterHots ? 'Belum ada item yang cocok' : 'Bank soal masih kosong'}
										description={search || filterSubject || filterWorkflow || filterType || filterHots
											? 'Ubah kombinasi filter atau reset pencarian untuk melihat item lain yang sudah tersedia.'
											: 'Tambahkan item pertama lewat mode beginner untuk input cepat, lalu lengkapi di advance bila dibutuhkan.'}
									>
										{#snippet children()}
											{#if search || filterSubject || filterWorkflow || filterType || filterHots}
												<Button variant="outline" size="sm" onclick={clearQuestionFilters}>Reset filter</Button>
											{:else}
												<Button size="sm" onclick={openCreate}>Tambah item pertama</Button>
											{/if}
										{/snippet}
									</EmptyStatePanel>
								</Table.Cell>
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
