<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import ChevronLeftIcon from '@lucide/svelte/icons/chevron-left';
	import ChevronRightIcon from '@lucide/svelte/icons/chevron-right';
	import ClipboardCheckIcon from '@lucide/svelte/icons/clipboard-check';
	import EyeIcon from '@lucide/svelte/icons/eye';
	import FileQuestionIcon from '@lucide/svelte/icons/file-question';
	import ListFilterIcon from '@lucide/svelte/icons/list-filter';
	import PencilIcon from '@lucide/svelte/icons/pencil';
	import PlusIcon from '@lucide/svelte/icons/plus';
	import RefreshCcwIcon from '@lucide/svelte/icons/refresh-ccw';
	import SearchIcon from '@lucide/svelte/icons/search';
	import UploadIcon from '@lucide/svelte/icons/upload';
	import * as Table from '$lib/components/ui/table';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { clientApiPathWithQuery, readClientApiData } from '$lib/client/api';
	import { htmlToPlainText } from '$lib/utils/html-text';

	type PageData = {
		user?: {
			role?: string;
			roles?: string[];
		};
	};

	type Subject = {
		id: string;
		name: string;
		code?: string;
	};

	type OptionItem = {
		label?: string;
		text?: string;
		html?: string;
		latex?: string;
		match_label?: string;
		match_text?: string;
		match_html?: string;
		is_distractor?: boolean;
	};

	type Question = {
		id: string;
		event_id?: string | null;
		subject_id?: string;
		subject_name?: string;
		subject_code?: string;
		code?: string;
		question_text?: string;
		question_type?: string;
		stem_html?: string;
		stimulus_html?: string;
		options?: OptionItem[];
		answer_key?: string;
		difficulty?: string;
		status?: string;
		workflow_status?: string;
		author_username?: string;
		reviewer_username?: string;
		review_notes?: string;
		authoring_mode?: string;
		suggested_mode?: string;
		academic_phase?: string;
		grade_level?: number | null;
		cp_ref?: string;
		tp_ref?: string;
		kd_ref?: string;
		indicator_ref?: string;
		material_topic?: string;
		cognitive_level?: string;
		hots_flag?: boolean;
		created_at?: string;
		updated_at?: string;
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
		items?: Question[];
		meta?: {
			total?: number;
			limit?: number;
			offset?: number;
		};
	};

	type AcademicPayload = {
		subjects?: Subject[];
		error?: string;
		message?: string;
	};

	type WorkflowFilter = '' | 'draft' | 'review' | 'approved' | 'rejected';
	type PublicationFilter = '' | 'draft' | 'published';
	type QuestionTypeFilter =
		| ''
		| 'multiple_choice'
		| 'multiple_answer'
		| 'true_false'
		| 'agree_disagree'
		| 'matching'
		| 'short_answer'
		| 'essay';
	type HotsFilter = '' | 'yes' | 'no';
	type StatusKey = 'all' | 'draft' | 'review' | 'rejected' | 'approved' | 'published';

	type StatusCounts = Record<StatusKey, number>;

	type BankSoalOverview = {
		questions: Question[];
		subjects: Subject[];
		totalItems: number;
		counts: StatusCounts;
		page: number;
		limit: number;
		offset: number;
	};

	type SummaryCard = {
		key: StatusKey;
		label: string;
		helper: string;
		value: number;
		tone: 'slate' | 'amber' | 'red' | 'green' | 'emerald';
		active: boolean;
	};

	let { data }: { data: PageData } = $props();

	const PAGE_SIZE = 12;
	const emptyCounts: StatusCounts = {
		all: 0,
		draft: 0,
		review: 0,
		rejected: 0,
		approved: 0,
		published: 0
	};

	const workflowOptions: Array<{ value: WorkflowFilter; label: string }> = [
		{ value: '', label: 'Semua alur' },
		{ value: 'draft', label: 'Draft' },
		{ value: 'review', label: 'Menunggu review' },
		{ value: 'approved', label: 'Disetujui' },
		{ value: 'rejected', label: 'Perlu revisi' }
	];

	const publicationOptions: Array<{ value: PublicationFilter; label: string }> = [
		{ value: '', label: 'Semua publikasi' },
		{ value: 'draft', label: 'Belum terbit' },
		{ value: 'published', label: 'Terbit' }
	];

	const questionTypeOptions: Array<{ value: QuestionTypeFilter; label: string }> = [
		{ value: '', label: 'Semua tipe' },
		{ value: 'multiple_choice', label: 'PG' },
		{ value: 'multiple_answer', label: 'PG kompleks' },
		{ value: 'true_false', label: 'Benar/Salah' },
		{ value: 'agree_disagree', label: 'Setuju/Tidak' },
		{ value: 'matching', label: 'Menjodohkan' },
		{ value: 'short_answer', label: 'Isian singkat' },
		{ value: 'essay', label: 'Uraian' }
	];

	const hotsOptions: Array<{ value: HotsFilter; label: string }> = [
		{ value: '', label: 'Semua HOTS' },
		{ value: 'yes', label: 'HOTS' },
		{ value: 'no', label: 'Non-HOTS' }
	];

	const workflowLabels: Record<string, string> = {
		draft: 'Draft',
		review: 'Review',
		approved: 'Disetujui',
		rejected: 'Perlu Revisi'
	};

	const publicationLabels: Record<string, string> = {
		draft: 'Belum Terbit',
		published: 'Terbit',
		archived: 'Arsip'
	};

	const questionTypeLabels: Record<string, string> = {
		multiple_choice: 'PG',
		multiple_answer: 'PG Kompleks',
		true_false: 'Benar/Salah',
		agree_disagree: 'Setuju/Tidak',
		matching: 'Menjodohkan',
		short_answer: 'Isian Singkat',
		essay: 'Uraian'
	};

	const difficultyLabels: Record<string, string> = {
		easy: 'Mudah',
		medium: 'Sedang',
		hard: 'Sulit'
	};

	let search = $state('');
	let subjectFilter = $state('');
	let workflowFilter = $state<WorkflowFilter>('');
	let publicationFilter = $state<PublicationFilter>('');
	let questionTypeFilter = $state<QuestionTypeFilter>('');
	let hotsFilter = $state<HotsFilter>('');
	let currentPage = $state(1);
	let questions = $state<Question[]>([]);
	let subjects = $state<Subject[]>([]);
	let totalItems = $state(0);
	let counts = $state<StatusCounts>({ ...emptyCounts });
	let questionsPromise = $state<Promise<BankSoalOverview> | null>(null);
	let refreshing = $state(false);
	let requestId = 0;
	let searchTimer: ReturnType<typeof setTimeout> | null = null;

	let roles = $derived(data.user?.roles ?? (data.user?.role ? [data.user.role] : []));
	let roleLabel = $derived.by(() => {
		if (roles.includes('admin')) return 'Admin';
		if (roles.includes('guru')) return 'Guru';
		return roles.length > 0 ? roles.join(', ') : 'Pengguna';
	});
	let pageCount = $derived(Math.max(1, Math.ceil(totalItems / PAGE_SIZE)));
	let resultStart = $derived(totalItems === 0 ? 0 : (currentPage - 1) * PAGE_SIZE + 1);
	let resultEnd = $derived(Math.min(totalItems, currentPage * PAGE_SIZE));
	let hasFilters = $derived(Boolean(
		search.trim() ||
			subjectFilter ||
			workflowFilter ||
			publicationFilter ||
			questionTypeFilter ||
			hotsFilter
	));
	let selectedSubject = $derived(subjects.find((subject) => subject.id === subjectFilter) ?? null);
	let composerHref = $derived(resolve('/cbt/soal'));
	let importHref = $derived(resolve('/cbt/soal/import'));
	let summaryCards = $derived<SummaryCard[]>([
		{
			key: 'all',
			label: 'Semua',
			helper: 'stok sesuai filter',
			value: counts.all,
			tone: 'slate',
			active: !workflowFilter && !publicationFilter
		},
		{
			key: 'draft',
			label: 'Draft',
			helper: 'masih disusun',
			value: counts.draft,
			tone: 'slate',
			active: workflowFilter === 'draft' && !publicationFilter
		},
		{
			key: 'review',
			label: 'Review',
			helper: 'menunggu keputusan',
			value: counts.review,
			tone: 'amber',
			active: workflowFilter === 'review' && !publicationFilter
		},
		{
			key: 'rejected',
			label: 'Revisi',
			helper: 'perlu perbaikan',
			value: counts.rejected,
			tone: 'red',
			active: workflowFilter === 'rejected' && !publicationFilter
		},
		{
			key: 'approved',
			label: 'Disetujui',
			helper: 'siap diterbitkan',
			value: counts.approved,
			tone: 'green',
			active: workflowFilter === 'approved' && publicationFilter === 'draft'
		},
		{
			key: 'published',
			label: 'Terbit',
			helper: 'siap dipakai paket',
			value: counts.published,
			tone: 'emerald',
			active: !workflowFilter && publicationFilter === 'published'
		}
	]);

	function buildBaseParams(limit: number, offset: number): URLSearchParams {
		const params = new URLSearchParams();
		params.set('limit', String(limit));
		params.set('offset', String(offset));
		if (search.trim()) params.set('q', search.trim());
		if (subjectFilter) params.set('subject_id', subjectFilter);
		if (questionTypeFilter) params.set('question_type', questionTypeFilter);
		if (hotsFilter) params.set('hots', hotsFilter);
		return params;
	}

	function appendStatusParams(params: URLSearchParams, key: StatusKey) {
		if (key === 'draft' || key === 'review' || key === 'rejected') {
			params.set('workflow_status', key);
			return;
		}
		if (key === 'approved') {
			params.set('workflow_status', 'approved');
			params.set('status', 'draft');
			return;
		}
		if (key === 'published') {
			params.set('status', 'published');
		}
	}

	function buildListParams(page: number): URLSearchParams {
		const params = buildBaseParams(PAGE_SIZE, Math.max(0, (page - 1) * PAGE_SIZE));
		if (workflowFilter) params.set('workflow_status', workflowFilter);
		if (publicationFilter) params.set('status', publicationFilter);
		return params;
	}

	function buildCountParams(key: StatusKey): URLSearchParams {
		const params = buildBaseParams(1, 0);
		appendStatusParams(params, key);
		return params;
	}

	async function fetchQuestionList(page: number): Promise<QuestionListResponse> {
		return fetch(clientApiPathWithQuery('/api/cbt/questions', buildListParams(page))).then((response) =>
			readClientApiData<QuestionListResponse>(response, 'Gagal memuat daftar soal')
		);
	}

	async function fetchQuestionCount(key: StatusKey): Promise<number> {
		const payload = await fetch(clientApiPathWithQuery('/api/cbt/questions', buildCountParams(key))).then((response) =>
			readClientApiData<QuestionListResponse>(response, 'Gagal memuat ringkasan status soal')
		);
		return payload.meta?.total ?? payload.items?.length ?? 0;
	}

	async function fetchSubjects(): Promise<Subject[]> {
		const payload = await fetch('/api/cbt/soal-support/subjects').then((response) =>
			readClientApiData<AcademicPayload>(response, 'Gagal memuat data mata pelajaran')
		);
		return payload.subjects ?? [];
	}

	async function fetchOverview(page = currentPage): Promise<BankSoalOverview> {
		const countKeys: StatusKey[] = ['all', 'draft', 'review', 'rejected', 'approved', 'published'];
		const [questionPayload, loadedSubjects, ...countValues] = await Promise.all([
			fetchQuestionList(page),
			subjects.length > 0 ? Promise.resolve(subjects) : fetchSubjects(),
			...countKeys.map((key) => fetchQuestionCount(key))
		]);
		const nextCounts = { ...emptyCounts };
		countKeys.forEach((key, index) => {
			nextCounts[key] = countValues[index] ?? 0;
		});
		return {
			questions: questionPayload.items ?? [],
			subjects: loadedSubjects,
			totalItems: questionPayload.meta?.total ?? questionPayload.items?.length ?? 0,
			counts: nextCounts,
			page,
			limit: questionPayload.meta?.limit ?? PAGE_SIZE,
			offset: questionPayload.meta?.offset ?? Math.max(0, (page - 1) * PAGE_SIZE)
		};
	}

	function currentOverview(): BankSoalOverview {
		return {
			questions,
			subjects,
			totalItems,
			counts,
			page: currentPage,
			limit: PAGE_SIZE,
			offset: Math.max(0, (currentPage - 1) * PAGE_SIZE)
		};
	}

	function applyOverview(overview: BankSoalOverview) {
		questions = overview.questions;
		subjects = overview.subjects;
		totalItems = overview.totalItems;
		counts = overview.counts;
		currentPage = overview.page;
	}

	function syncUrl(page: number) {
		if (typeof window === 'undefined') return;
		const params = new URLSearchParams();
		if (search.trim()) params.set('q', search.trim());
		if (subjectFilter) params.set('subject_id', subjectFilter);
		if (workflowFilter) params.set('workflow_status', workflowFilter);
		if (publicationFilter) params.set('status', publicationFilter);
		if (questionTypeFilter) params.set('question_type', questionTypeFilter);
		if (hotsFilter) params.set('hots', hotsFilter);
		if (page > 1) params.set('page', String(page));
		const query = params.toString();
		window.history.replaceState({}, '', query ? `${window.location.pathname}?${query}` : window.location.pathname);
	}

	function load(page = currentPage, markRefreshing = false) {
		const nextPage = Math.max(1, page);
		currentPage = nextPage;
		syncUrl(nextPage);
		if (markRefreshing) refreshing = true;
		const activeRequestId = ++requestId;
		const promise = fetchOverview(nextPage)
			.then((overview) => {
				if (activeRequestId !== requestId) return currentOverview();
				applyOverview(overview);
				return overview;
			})
			.catch((error: unknown) => {
				if (activeRequestId !== requestId) return currentOverview();
				throw error;
			})
			.finally(() => {
				if (activeRequestId === requestId) refreshing = false;
			});
		questionsPromise = promise;
	}

	function applyFilters(event?: SubmitEvent) {
		event?.preventDefault();
		load(1, true);
	}

	function onSearchInput(event: Event) {
		search = (event.currentTarget as HTMLInputElement).value;
		if (searchTimer) clearTimeout(searchTimer);
		searchTimer = setTimeout(() => load(1, true), 350);
	}

	function clearFilters() {
		if (searchTimer) clearTimeout(searchTimer);
		search = '';
		subjectFilter = '';
		workflowFilter = '';
		publicationFilter = '';
		questionTypeFilter = '';
		hotsFilter = '';
		load(1, true);
	}

	function setSummaryFilter(key: StatusKey) {
		if (key === 'all') {
			workflowFilter = '';
			publicationFilter = '';
		} else if (key === 'approved') {
			workflowFilter = 'approved';
			publicationFilter = 'draft';
		} else if (key === 'published') {
			workflowFilter = '';
			publicationFilter = 'published';
		} else {
			workflowFilter = key;
			publicationFilter = '';
		}
		load(1, true);
	}

	function readInitialFilters() {
		if (typeof window === 'undefined') return;
		const params = new URLSearchParams(window.location.search);
		search = params.get('q') ?? '';
		subjectFilter = params.get('subject_id') ?? '';
		workflowFilter = normalizeWorkflowFilter(params.get('workflow_status'));
		publicationFilter = normalizePublicationFilter(params.get('status'));
		questionTypeFilter = normalizeQuestionTypeFilter(params.get('question_type'));
		hotsFilter = normalizeHotsFilter(params.get('hots'));
		const parsedPage = Number.parseInt(params.get('page') ?? '1', 10);
		currentPage = Number.isFinite(parsedPage) && parsedPage > 0 ? parsedPage : 1;
	}

	function normalizeWorkflowFilter(value: string | null): WorkflowFilter {
		return value === 'draft' || value === 'review' || value === 'approved' || value === 'rejected' ? value : '';
	}

	function normalizePublicationFilter(value: string | null): PublicationFilter {
		return value === 'draft' || value === 'published' ? value : '';
	}

	function normalizeQuestionTypeFilter(value: string | null): QuestionTypeFilter {
		return questionTypeOptions.some((option) => option.value === value) ? (value as QuestionTypeFilter) : '';
	}

	function normalizeHotsFilter(value: string | null): HotsFilter {
		return value === 'yes' || value === 'no' ? value : '';
	}

	function questionHref(question: Question): string {
		return resolve('/cbt/soal') + `?question_id=${encodeURIComponent(question.id)}`;
	}

	function reviewHref(): string {
		const params = new URLSearchParams();
		if (subjectFilter) params.set('subject_id', subjectFilter);
		return params.toString() ? `${resolve('/cbt/soal/review')}?${params.toString()}` : resolve('/cbt/soal/review');
	}

	function questionPreview(question: Question): string {
		const text = htmlToPlainText(question.stem_html || question.question_text || '').replace(/\s+/g, ' ').trim();
		if (!text) return '(Isi soal belum tersedia)';
		return text.length > 180 ? `${text.slice(0, 180)}...` : text;
	}

	function compactText(value: string | undefined | null, fallback = '-'): string {
		const trimmed = (value ?? '').trim();
		return trimmed || fallback;
	}

	function subjectLabel(question: Question): string {
		const code = compactText(question.subject_code, '');
		const name = compactText(question.subject_name, '');
		if (name && code) return `${code} - ${name}`;
		return name || code || 'Mapel belum tersedia';
	}

	function questionTypeLabel(value: string | undefined): string {
		return questionTypeLabels[value ?? ''] ?? compactText(value, 'Tipe belum dipilih');
	}

	function workflowLabel(value: string | undefined): string {
		return workflowLabels[value ?? ''] ?? compactText(value, 'Belum ada alur');
	}

	function publicationLabel(value: string | undefined): string {
		return publicationLabels[value ?? ''] ?? compactText(value, 'Belum ada status');
	}

	function difficultyLabel(value: string | undefined): string {
		return difficultyLabels[value ?? ''] ?? compactText(value, 'Kesulitan belum diisi');
	}

	function formatDate(value: string | undefined): string {
		if (!value) return '-';
		const date = new Date(value);
		if (Number.isNaN(date.getTime())) return '-';
		return new Intl.DateTimeFormat('id-ID', {
			day: '2-digit',
			month: 'short',
			year: 'numeric'
		}).format(date);
	}

	function gradeLabel(question: Question): string {
		const grade = question.grade_level ? `Kelas ${question.grade_level}` : '';
		const phase = compactText(question.academic_phase, '');
		if (grade && phase) return `${grade} / Fase ${phase}`;
		return grade || (phase ? `Fase ${phase}` : 'Level belum diisi');
	}

	function questionUsageLocked(question: Question): boolean {
		const packageCount = question.package_count ?? question.usage?.package_count ?? 0;
		const answerCount = question.answer_count ?? question.usage?.answer_count ?? 0;
		return Boolean(question.is_locked ?? question.usage?.is_locked ?? (packageCount > 0 || answerCount > 0));
	}

	function questionUsageText(question: Question): string {
		const packageCount = question.package_count ?? question.usage?.package_count ?? 0;
		const answerCount = question.answer_count ?? question.usage?.answer_count ?? 0;
		if (packageCount > 0 && answerCount > 0) return `${packageCount} paket, ${answerCount} jawaban`;
		if (packageCount > 0) return `${packageCount} paket`;
		if (answerCount > 0) return `${answerCount} jawaban`;
		return 'Belum dipakai';
	}

	function isQuickEditable(question: Question): boolean {
		return (
			(question.workflow_status === 'draft' || question.workflow_status === 'rejected') &&
			(question.status ?? 'draft') === 'draft' &&
			!questionUsageLocked(question)
		);
	}

	function workflowBadgeClass(value: string | undefined): string {
		switch (value) {
			case 'review':
				return 'border-amber-200 bg-amber-50 text-amber-800';
			case 'approved':
				return 'border-emerald-200 bg-emerald-50 text-emerald-800';
			case 'rejected':
				return 'border-red-200 bg-red-50 text-red-700';
			case 'draft':
				return 'border-slate-200 bg-slate-50 text-slate-700';
			default:
				return 'border-slate-200 bg-white text-slate-600';
		}
	}

	function publicationBadgeClass(value: string | undefined): string {
		if (value === 'published') return 'border-green-200 bg-green-50 text-green-800';
		return 'border-slate-200 bg-white text-slate-600';
	}

	function summaryCardClass(card: SummaryCard): string {
		const active = card.active ? 'border-emerald-500 bg-emerald-50 shadow-sm' : 'border-slate-200 bg-white hover:border-emerald-200 hover:bg-emerald-50/40';
		return `rounded-lg border p-4 text-left transition ${active}`;
	}

	function summaryValueClass(card: SummaryCard): string {
		if (card.tone === 'amber') return 'text-amber-800';
		if (card.tone === 'red') return 'text-red-700';
		if (card.tone === 'green') return 'text-green-800';
		if (card.tone === 'emerald') return 'text-emerald-800';
		return 'text-slate-900';
	}

	function handleQuestionsRenderError(error: unknown, reset: () => void) {
		console.error('Bank soal render failed', error);
		reset();
	}

	onMount(() => {
		readInitialFilters();
		load(currentPage);
		return () => {
			if (searchTimer) clearTimeout(searchTimer);
		};
	});
</script>

<svelte:head>
	<title>Daftar Soal Bank Soal - MTsN 2 Kolaka Utara</title>
</svelte:head>

<div class="space-y-5">
	<section class="rounded-lg border border-emerald-100 bg-white p-5 shadow-sm">
		<div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
			<div class="max-w-3xl space-y-2">
				<p class="text-xs font-semibold uppercase tracking-[0.18em] text-emerald-700">Bank Soal</p>
				<div class="flex flex-wrap items-center gap-3">
					<h1 class="text-2xl font-semibold tracking-tight text-slate-950 md:text-3xl">Daftar Soal</h1>
					<Badge variant="outline" class="border-emerald-200 bg-emerald-50 text-emerald-800">{roleLabel}</Badge>
				</div>
				<p class="max-w-2xl text-sm leading-6 text-slate-600">
					Inventaris soal CBT untuk guru dan admin, dengan status review, publikasi, dan pemakaian soal dalam satu daftar yang mudah dipindai.
				</p>
			</div>
			<div class="flex flex-wrap gap-2">
				<Button href={reviewHref()} variant="outline" class="border-amber-200 text-amber-800 hover:bg-amber-50">
					<ClipboardCheckIcon class="size-4" />
					Review
				</Button>
				<Button href={importHref} variant="outline" class="border-emerald-200 text-emerald-800 hover:bg-emerald-50">
					<UploadIcon class="size-4" />
					Impor
				</Button>
				<Button href={composerHref}>
					<PlusIcon class="size-4" />
					Tambah Soal
				</Button>
			</div>
		</div>
	</section>

	<section class="grid gap-3 sm:grid-cols-2 xl:grid-cols-6" aria-label="Ringkasan status soal">
		{#each summaryCards as card (card.key)}
			<button type="button" class={summaryCardClass(card)} aria-pressed={card.active} onclick={() => setSummaryFilter(card.key)}>
				<span class="text-xs font-semibold uppercase tracking-wide text-slate-500">{card.label}</span>
				<span class={`mt-2 block text-2xl font-semibold ${summaryValueClass(card)}`}>{card.value}</span>
				<span class="mt-1 block text-xs text-slate-500">{card.helper}</span>
			</button>
		{/each}
	</section>

	<section class="rounded-lg border border-slate-200 bg-white p-4 shadow-sm">
		<form class="grid gap-3 lg:grid-cols-[minmax(16rem,1.5fr)_repeat(5,minmax(9rem,1fr))_auto] lg:items-end" onsubmit={applyFilters}>
			<div class="space-y-1">
				<label for="bank-soal-search" class="text-xs font-semibold text-slate-600">Cari soal</label>
				<div class="relative">
					<SearchIcon class="pointer-events-none absolute left-2.5 top-1/2 size-4 -translate-y-1/2 text-slate-400" />
					<Input
						id="bank-soal-search"
						value={search}
						oninput={onSearchInput}
						placeholder="Kode, isi soal, materi..."
						class="h-9 pl-8"
					/>
				</div>
			</div>

			<div class="space-y-1">
				<label for="bank-soal-subject" class="text-xs font-semibold text-slate-600">Mata pelajaran</label>
				<select
					id="bank-soal-subject"
					bind:value={subjectFilter}
					onchange={() => load(1, true)}
					class="h-9 w-full rounded-md border border-slate-200 bg-white px-3 text-sm text-slate-700 focus:outline-none focus:ring-2 focus:ring-emerald-600"
				>
					<option value="">Semua mapel</option>
					{#each subjects as subject (subject.id)}
						<option value={subject.id}>{subject.code ? `${subject.code} - ${subject.name}` : subject.name}</option>
					{/each}
				</select>
			</div>

			<div class="space-y-1">
				<label for="bank-soal-workflow" class="text-xs font-semibold text-slate-600">Alur</label>
				<select
					id="bank-soal-workflow"
					bind:value={workflowFilter}
					onchange={() => load(1, true)}
					class="h-9 w-full rounded-md border border-slate-200 bg-white px-3 text-sm text-slate-700 focus:outline-none focus:ring-2 focus:ring-emerald-600"
				>
					{#each workflowOptions as option (option.value)}
						<option value={option.value}>{option.label}</option>
					{/each}
				</select>
			</div>

			<div class="space-y-1">
				<label for="bank-soal-publication" class="text-xs font-semibold text-slate-600">Publikasi</label>
				<select
					id="bank-soal-publication"
					bind:value={publicationFilter}
					onchange={() => load(1, true)}
					class="h-9 w-full rounded-md border border-slate-200 bg-white px-3 text-sm text-slate-700 focus:outline-none focus:ring-2 focus:ring-emerald-600"
				>
					{#each publicationOptions as option (option.value)}
						<option value={option.value}>{option.label}</option>
					{/each}
				</select>
			</div>

			<div class="space-y-1">
				<label for="bank-soal-type" class="text-xs font-semibold text-slate-600">Tipe</label>
				<select
					id="bank-soal-type"
					bind:value={questionTypeFilter}
					onchange={() => load(1, true)}
					class="h-9 w-full rounded-md border border-slate-200 bg-white px-3 text-sm text-slate-700 focus:outline-none focus:ring-2 focus:ring-emerald-600"
				>
					{#each questionTypeOptions as option (option.value)}
						<option value={option.value}>{option.label}</option>
					{/each}
				</select>
			</div>

			<div class="space-y-1">
				<label for="bank-soal-hots" class="text-xs font-semibold text-slate-600">HOTS</label>
				<select
					id="bank-soal-hots"
					bind:value={hotsFilter}
					onchange={() => load(1, true)}
					class="h-9 w-full rounded-md border border-slate-200 bg-white px-3 text-sm text-slate-700 focus:outline-none focus:ring-2 focus:ring-emerald-600"
				>
					{#each hotsOptions as option (option.value)}
						<option value={option.value}>{option.label}</option>
					{/each}
				</select>
			</div>

			<div class="flex gap-2">
				<Button type="submit" variant="outline" class="h-9">
					<ListFilterIcon class="size-4" />
					Terapkan
				</Button>
				<LoadingButton
					variant="outline"
					class="h-9"
					loading={refreshing}
					loadingLabel="Memuat"
					aria-label="Muat ulang daftar soal"
					onclick={() => load(currentPage, true)}
				>
					<RefreshCcwIcon class="size-4" />
				</LoadingButton>
			</div>
		</form>

		{#if hasFilters}
			<div class="mt-3 flex flex-wrap items-center gap-2 text-xs text-slate-500">
				<span>
					Filter aktif{selectedSubject ? `: ${selectedSubject.name}` : ''}
				</span>
				<Button variant="ghost" size="sm" class="h-7 text-emerald-800" onclick={clearFilters}>Bersihkan filter</Button>
			</div>
		{/if}
	</section>

	<AsyncContent promise={questionsPromise} onerror={handleQuestionsRenderError}>
		{#snippet pending()}
			<section class="rounded-lg border border-slate-200 bg-white p-4 shadow-sm">
				<div class="flex items-center justify-between">
					<Skeleton class="h-5 w-40" />
					<Skeleton class="h-5 w-24" />
				</div>
				<div class="mt-4 space-y-3">
					{#each Array.from({ length: 6 }) as _, index (`bank-soal-skeleton-${index}`)}
						<div class="rounded-lg border border-slate-100 p-4">
							<Skeleton class="h-4 w-28" />
							<Skeleton class="mt-3 h-5 w-full max-w-2xl" />
							<div class="mt-3 flex gap-2">
								<Skeleton class="h-5 w-20" />
								<Skeleton class="h-5 w-24" />
								<Skeleton class="h-5 w-16" />
							</div>
						</div>
					{/each}
				</div>
			</section>
		{/snippet}

		{#snippet failed(error, reset)}
			<section class="rounded-lg border border-slate-200 bg-white p-5 shadow-sm">
				<RecoveryPanel
					title="Daftar soal belum dapat dimuat"
					message={error instanceof Error ? error.message : 'Gagal memuat daftar soal'}
					onRetry={() => {
						reset?.();
						load(currentPage, true);
					}}
				/>
			</section>
		{/snippet}

		{#snippet children(value)}
			{@const overview = value as BankSoalOverview}
			{@const currentQuestions = overview.questions}
			<section class="overflow-hidden rounded-lg border border-slate-200 bg-white shadow-sm">
				<div class="flex flex-col gap-2 border-b border-slate-200 p-4 md:flex-row md:items-center md:justify-between">
					<div>
						<h2 class="text-base font-semibold text-slate-900">Soal Tersedia</h2>
						<p class="mt-1 text-xs text-slate-500">
							{totalItems === 0 ? 'Tidak ada soal pada filter ini' : `${resultStart}-${resultEnd} dari ${totalItems} soal`}
						</p>
					</div>
					<div class="flex flex-wrap gap-2">
						<Badge variant="outline" class="border-slate-200 bg-slate-50 text-slate-700">{PAGE_SIZE} per halaman</Badge>
						{#if selectedSubject}
							<Badge variant="outline" class="border-emerald-200 bg-emerald-50 text-emerald-800">{selectedSubject.name}</Badge>
						{/if}
					</div>
				</div>

				{#if currentQuestions.length === 0}
					<div class="flex flex-col items-center px-6 py-12 text-center">
						<div class="flex size-12 items-center justify-center rounded-lg bg-emerald-50 text-emerald-800">
							<FileQuestionIcon class="size-6" />
						</div>
						<h2 class="mt-4 text-lg font-semibold text-slate-900">
							{hasFilters ? 'Soal tidak ditemukan' : 'Belum ada soal di Bank Soal'}
						</h2>
						<p class="mt-2 max-w-md text-sm leading-6 text-slate-600">
							{hasFilters
								? 'Coba longgarkan filter atau cari dengan kode, materi, dan isi soal yang lebih umum.'
								: 'Mulai dari komposer untuk menulis soal pertama, atau impor CSV dari arsip soal lama.'}
						</p>
						<div class="mt-5 flex flex-wrap justify-center gap-2">
							{#if hasFilters}
								<Button variant="outline" onclick={clearFilters}>Bersihkan filter</Button>
							{/if}
							<Button href={composerHref}>
								<PlusIcon class="size-4" />
								Tambah Soal
							</Button>
							<Button href={importHref} variant="outline">
								<UploadIcon class="size-4" />
								Impor CSV
							</Button>
						</div>
					</div>
				{:else}
					<div class="hidden md:block">
						<Table.Root>
							<Table.Header>
								<Table.Row class="bg-slate-50">
									<Table.Head class="w-[44%] text-slate-600">Soal</Table.Head>
									<Table.Head class="w-[18%] text-slate-600">Mapel & Level</Table.Head>
									<Table.Head class="w-[16%] text-slate-600">Status</Table.Head>
									<Table.Head class="w-[12%] text-slate-600">Pemakaian</Table.Head>
									<Table.Head class="w-[10%] text-right text-slate-600">Aksi</Table.Head>
								</Table.Row>
							</Table.Header>
							<Table.Body>
								{#each currentQuestions as question (question.id)}
									<Table.Row class="align-top hover:bg-slate-50">
										<Table.Cell>
											<div class="space-y-2">
												<div class="flex flex-wrap items-center gap-2">
													<span class="font-mono text-xs font-semibold text-emerald-800">{compactText(question.code, 'Tanpa kode')}</span>
													<Badge variant="outline" class="border-slate-200 bg-white text-slate-600">
														{questionTypeLabel(question.question_type)}
													</Badge>
													{#if question.hots_flag}
														<Badge variant="outline" class="border-amber-200 bg-amber-50 text-amber-800">HOTS</Badge>
													{/if}
													{#if questionUsageLocked(question)}
														<Badge variant="outline" class="border-red-200 bg-red-50 text-red-700">Terkunci</Badge>
													{/if}
												</div>
												<p class="line-clamp-2 text-sm leading-6 text-slate-800">{questionPreview(question)}</p>
												<div class="flex flex-wrap gap-x-3 gap-y-1 text-xs text-slate-500">
													<span>{compactText(question.author_username, 'Penulis belum tercatat')}</span>
													<span>{difficultyLabel(question.difficulty)}</span>
													<span>{compactText(question.material_topic, 'Materi belum diisi')}</span>
													<span>Dibuat {formatDate(question.created_at)}</span>
												</div>
											</div>
										</Table.Cell>
										<Table.Cell class="text-sm text-slate-700">
											<div class="font-medium text-slate-900">{subjectLabel(question)}</div>
											<div class="mt-1 text-xs text-slate-500">{gradeLabel(question)}</div>
											{#if question.cognitive_level}
												<div class="mt-1 text-xs text-slate-500">{question.cognitive_level}</div>
											{/if}
										</Table.Cell>
										<Table.Cell>
											<div class="flex flex-col items-start gap-1.5">
												<Badge variant="outline" class={workflowBadgeClass(question.workflow_status)}>
													{workflowLabel(question.workflow_status)}
												</Badge>
												<Badge variant="outline" class={publicationBadgeClass(question.status)}>
													{publicationLabel(question.status)}
												</Badge>
											</div>
										</Table.Cell>
										<Table.Cell class="text-sm text-slate-600">
											{questionUsageText(question)}
										</Table.Cell>
										<Table.Cell class="text-right">
											<div class="flex justify-end gap-2">
												<Button href={questionHref(question)} variant="outline" size="sm">
													{#if isQuickEditable(question)}
														<PencilIcon class="size-3.5" />
														Edit
													{:else}
														<EyeIcon class="size-3.5" />
														Lihat
													{/if}
												</Button>
											</div>
										</Table.Cell>
									</Table.Row>
								{/each}
							</Table.Body>
						</Table.Root>
					</div>

					<div class="divide-y divide-slate-100 md:hidden">
						{#each currentQuestions as question (question.id)}
							<article class="space-y-3 p-4">
								<div class="flex items-start justify-between gap-3">
									<div class="min-w-0">
										<p class="font-mono text-xs font-semibold text-emerald-800">{compactText(question.code, 'Tanpa kode')}</p>
										<h2 class="mt-1 line-clamp-3 text-sm font-semibold leading-6 text-slate-900">{questionPreview(question)}</h2>
									</div>
									<Badge variant="outline" class={workflowBadgeClass(question.workflow_status)}>
										{workflowLabel(question.workflow_status)}
									</Badge>
								</div>
								<div class="flex flex-wrap gap-1.5">
									<Badge variant="outline" class="border-slate-200 bg-white text-slate-600">{questionTypeLabel(question.question_type)}</Badge>
									<Badge variant="outline" class={publicationBadgeClass(question.status)}>{publicationLabel(question.status)}</Badge>
									{#if question.hots_flag}
										<Badge variant="outline" class="border-amber-200 bg-amber-50 text-amber-800">HOTS</Badge>
									{/if}
								</div>
								<dl class="grid grid-cols-2 gap-2 text-xs text-slate-500">
									<div>
										<dt class="font-semibold text-slate-600">Mapel</dt>
										<dd class="mt-0.5">{subjectLabel(question)}</dd>
									</div>
									<div>
										<dt class="font-semibold text-slate-600">Level</dt>
										<dd class="mt-0.5">{gradeLabel(question)}</dd>
									</div>
									<div>
										<dt class="font-semibold text-slate-600">Pemakaian</dt>
										<dd class="mt-0.5">{questionUsageText(question)}</dd>
									</div>
									<div>
										<dt class="font-semibold text-slate-600">Dibuat</dt>
										<dd class="mt-0.5">{formatDate(question.created_at)}</dd>
									</div>
								</dl>
								<Button href={questionHref(question)} variant="outline" class="w-full">
									{#if isQuickEditable(question)}
										<PencilIcon class="size-4" />
										Edit di Komposer
									{:else}
										<EyeIcon class="size-4" />
										Lihat Soal
									{/if}
								</Button>
							</article>
						{/each}
					</div>
				{/if}
			</section>
		{/snippet}
	</AsyncContent>

	{#if totalItems > PAGE_SIZE}
		<nav class="flex flex-col gap-3 rounded-lg border border-slate-200 bg-white p-3 text-sm text-slate-600 shadow-sm sm:flex-row sm:items-center sm:justify-between" aria-label="Navigasi halaman daftar soal">
			<span>{resultStart}-{resultEnd} dari {totalItems} soal</span>
			<div class="flex items-center gap-2">
				<Button
					variant="outline"
					disabled={currentPage <= 1 || refreshing}
					onclick={() => load(currentPage - 1, true)}
				>
					<ChevronLeftIcon class="size-4" />
					Sebelumnya
				</Button>
				<span class="min-w-20 text-center text-xs font-semibold text-slate-500">Hal {currentPage} / {pageCount}</span>
				<Button
					variant="outline"
					disabled={currentPage >= pageCount || refreshing}
					onclick={() => load(currentPage + 1, true)}
				>
					Berikutnya
					<ChevronRightIcon class="size-4" />
				</Button>
			</div>
		</nav>
	{/if}
</div>
