<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import ArrowLeftIcon from '@lucide/svelte/icons/arrow-left';
	import PrinterIcon from '@lucide/svelte/icons/printer';
	import RefreshCcwIcon from '@lucide/svelte/icons/refresh-ccw';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import RichContent from '$lib/components/RichContent.svelte';
	import { clientApiPathWithQuery, readClientApiData } from '$lib/client/api';
	import type { AccountIdentity } from '$lib/client/account';
	import type { AuthUser } from '$lib/server/auth';

	type PageData = {
		user?: AuthUser;
		account?: AccountIdentity | null;
	};

	type Subject = {
		id: string;
		name: string;
		code?: string;
	};

	type Author = {
		username: string;
		display_name?: string | null;
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
		subject_id?: string;
		subject_name?: string;
		subject_code?: string;
		code?: string;
		question_text?: string;
		question_type?: string;
		stem_html?: string;
		stem_latex?: string;
		stimulus_html?: string;
		options?: OptionItem[];
		option_a?: string;
		option_b?: string;
		option_c?: string;
		option_d?: string;
		option_e?: string;
		answer_key?: string;
		explanation?: string;
		explanation_html?: string;
		rubric_html?: string;
		difficulty?: string;
		status?: string;
		workflow_status?: string;
		author_username?: string;
		author_display_name?: string;
		academic_phase?: string;
		target_level?: string | null;
		cp_ref?: string;
		tp_ref?: string;
		kd_ref?: string;
		indicator_ref?: string;
		material_topic?: string;
		cognitive_level?: string;
		hots_flag?: boolean;
		created_at?: string;
		updated_at?: string;
	};

	type QuestionListResponse = {
		items?: Question[];
		meta?: { total?: number; limit?: number; offset?: number };
	};

	type AcademicPayload = {
		subjects?: Subject[];
	};

	let { data }: { data: PageData } = $props();

	let loadingReferences = $state(true);
	let loadingQuestions = $state(false);
	let hasLoadedQuestions = $state(false);
	let error = $state('');
	let questions = $state<Question[]>([]);
	let subjects = $state<Subject[]>([]);
	let authors = $state<Author[]>([]);
	let authorFilter = $state('');
	let subjectFilter = $state('');
	type WorkflowFilter = 'draft' | 'submitted' | 'review' | 'reviewed' | 'revision_needed' | 'approved' | 'published' | 'rejected' | 'archived';
	let workflowFilters = $state<WorkflowFilter[]>([]);
	let workflowDropdownOpen = $state(false);
	let typeFilter = $state('');
	let dateFrom = $state('');
	let dateTo = $state('');
	let sortOrder = $state('subject_code_asc');
	let resultLimit = $state('100');
	let loadedTotal = $state(0);
	let includeAnswer = $state(false);
	let includeExplanation = $state(false);
	let includeMetadata = $state(true);

	const listHref = resolve('/bank-soal/daftar');
	const printTitle = 'Cetak Soal Saya';

	const workflowOptions = [
		{ value: '', label: 'Semua status' },
		{ value: 'draft', label: 'Draft' },
		{ value: 'submitted', label: 'Diajukan' },
		{ value: 'review', label: 'Review' },
		{ value: 'revision_needed', label: 'Perlu Revisi' },
		{ value: 'reviewed', label: 'Sudah Direview' },
		{ value: 'approved', label: 'Disetujui' },
		{ value: 'published', label: 'Terbit' },
		{ value: 'rejected', label: 'Ditolak' },
		{ value: 'archived', label: 'Arsip' }
	];

	const sortOptions = [
		{ value: 'subject_code_asc', label: 'Mapel lalu kode soal (A-Z)' },
		{ value: 'code_asc', label: 'Kode soal (A-Z)' },
		{ value: 'code_desc', label: 'Kode soal (Z-A)' },
		{ value: 'newest', label: 'Tanggal input terbaru' },
		{ value: 'oldest', label: 'Tanggal input terlama' },
		{ value: 'author_subject_code_asc', label: 'Guru, mapel, lalu kode soal' },
		{ value: 'status_subject_code_asc', label: 'Status, mapel, lalu kode soal' },
		{ value: 'type_subject_code_asc', label: 'Tipe, mapel, lalu kode soal' }
	];

	const resultLimitOptions = [
		{ value: '50', label: '50 soal' },
		{ value: '100', label: '100 soal' },
		{ value: '200', label: '200 soal' },
		{ value: 'all', label: 'Semua hasil filter' }
	];

	const workflowSelectableOptions = workflowOptions.filter((option) => option.value) as Array<{
		value: WorkflowFilter;
		label: string;
	}>;

	const questionTypeOptions = [
		{ value: '', label: 'Semua tipe' },
		{ value: 'multiple_choice', label: 'Pilihan Ganda' },
		{ value: 'multiple_answer', label: 'Pilihan Kompleks' },
		{ value: 'true_false', label: 'Benar/Salah' },
		{ value: 'agree_disagree', label: 'Setuju/Tidak Setuju' },
		{ value: 'matching', label: 'Menjodohkan' },
		{ value: 'short_answer', label: 'Isian Singkat' },
		{ value: 'essay', label: 'Esai' }
	];

	let currentUsername = $derived(data.user?.username ?? data.account?.username ?? '');
	let userRoles = $derived(data.user?.roles ?? (data.user?.role ? [data.user.role] : []));
	let userPermissions = $derived(data.user?.permissions ?? []);
	let canPrintOtherAuthors = $derived(
		userRoles.includes('admin') || userPermissions.includes('bank_soal.manage') || userPermissions.includes('bank_soal.analytics')
	);
	let teacherName = $derived(
		data.account?.profile_nama || data.account?.display_name || data.user?.username || 'Guru'
	);
	let loading = $derived(loadingReferences || loadingQuestions);
	let filteredQuestions = $derived.by(() =>
		sortQuestions(
			questions.filter((question) => {
				if (subjectFilter && question.subject_id !== subjectFilter) return false;
				if (workflowFilters.length > 0 && !workflowFilters.includes((question.workflow_status ?? question.status ?? '') as WorkflowFilter)) return false;
				if (typeFilter && question.question_type !== typeFilter) return false;
				if (dateFrom && normalizeDate(question.created_at) < dateFrom) return false;
				if (dateTo && normalizeDate(question.created_at) > dateTo) return false;
				return true;
			})
		)
	);
	let selectedSubjectName = $derived(
		subjectFilter ? subjects.find((subject) => subject.id === subjectFilter)?.name || 'Mapel dipilih' : 'Semua mapel'
	);
	let selectedAuthorName = $derived.by(() => {
		if (!canPrintOtherAuthors) return teacherName;
		if (!authorFilter) return 'Semua guru';
		const author = authors.find((item) => item.username === authorFilter);
		return author?.display_name || author?.username || authorFilter;
	});
	let selectedSortName = $derived(sortOptions.find((option) => option.value === sortOrder)?.label ?? 'Urutan standar');
	let selectedLimitName = $derived(resultLimitOptions.find((option) => option.value === resultLimit)?.label ?? '100 soal');
	let isBroadAllSelection = $derived(
		canPrintOtherAuthors &&
		!authorFilter &&
		!subjectFilter &&
		workflowFilters.length === 0 &&
		!typeFilter &&
		!dateFrom &&
		!dateTo &&
		resultLimit === 'all'
	);
	let shouldWarnLargeSelection = $derived(
		resultLimit === 'all' || (canPrintOtherAuthors && !authorFilter && !subjectFilter && workflowFilters.length === 0 && !typeFilter)
	);
	let loadedHasMore = $derived(loadedTotal > questions.length);

	onMount(() => {
		void loadReferenceData();
	});

	async function loadReferenceData() {
		if (!currentUsername) {
			error = 'Akun belum terbaca. Silakan login ulang sebelum mencetak soal.';
			loadingReferences = false;
			return;
		}
		loadingReferences = true;
		error = '';
		try {
			const [subjectPayload, authorPayload] = await Promise.all([
				fetch('/api/bank-soal/soal-support/subjects').then((response) =>
					readClientApiData<AcademicPayload>(response, 'Gagal memuat daftar mapel')
				),
				canPrintOtherAuthors
					? fetch('/api/bank-soal/soal-support/authors').then((response) =>
						readClientApiData<Author[]>(response, 'Gagal memuat daftar guru')
					)
					: Promise.resolve([] as Author[])
			]);
			subjects = subjectPayload.subjects ?? [];
			authors = authorPayload ?? [];
		} catch (e) {
			error = e instanceof Error ? e.message : 'Data referensi cetak belum dapat dimuat';
		} finally {
			loadingReferences = false;
		}
	}

	async function loadPrintData() {
		if (!currentUsername) {
			error = 'Akun belum terbaca. Silakan login ulang sebelum mencetak soal.';
			return;
		}
		loadingQuestions = true;
		error = '';
		try {
			if (isBroadAllSelection) {
				const ok = window.confirm(
					'Bapak memilih Semua guru + semua filter + semua hasil. Ini bisa memuat ratusan soal dan membuat browser berat. Lanjutkan?'
				);
				if (!ok) {
					loadingQuestions = false;
					return;
				}
			}
			const params = new URLSearchParams({
				limit: resultLimit === 'all' ? '2000' : resultLimit,
				offset: '0',
				sort: backendSortOrder(sortOrder)
			});
			if (canPrintOtherAuthors) {
				if (authorFilter) params.set('author_username', authorFilter);
			} else {
				params.set('author_username', currentUsername);
			}
			if (subjectFilter) params.set('subject_id', subjectFilter);
			workflowFilters.forEach((status) => params.append('workflow_status', status));
			if (typeFilter) params.set('question_type', typeFilter);
			const questionPayload = await fetch(clientApiPathWithQuery('/api/bank-soal/questions', params)).then((response) =>
				readClientApiData<QuestionListResponse>(response, 'Gagal memuat soal untuk dicetak')
			);
			questions = questionPayload.items ?? [];
			loadedTotal = questionPayload.meta?.total ?? questions.length;
			hasLoadedQuestions = true;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Data cetak belum dapat dimuat';
		} finally {
			loadingQuestions = false;
		}
	}

	function normalizeDate(value: string | undefined): string {
		if (!value) return '';
		return value.slice(0, 10);
	}

	function backendSortOrder(value: string): string {
		if (value === 'oldest') return 'oldest';
		return 'newest';
	}

	function compareText(left: string | undefined | null, right: string | undefined | null): number {
		return compactText(left, '').localeCompare(compactText(right, ''), 'id', { numeric: true, sensitivity: 'base' });
	}

	function compareDate(left: string | undefined, right: string | undefined): number {
		return new Date(left ?? 0).getTime() - new Date(right ?? 0).getTime();
	}

	function toggleWorkflowFilter(value: WorkflowFilter) {
		workflowFilters = workflowFilters.includes(value)
			? workflowFilters.filter((item) => item !== value)
			: [...workflowFilters, value];
		clearLoadedQuestions();
	}

	function clearWorkflowFilters() {
		workflowFilters = [];
		clearLoadedQuestions();
	}

	function workflowFilterSummary(): string {
		if (workflowFilters.length === 0) return 'Semua status';
		if (workflowFilters.length === 1) return workflowLabel(workflowFilters[0]);
		return `${workflowFilters.length} status dipilih`;
	}

	function sortQuestions(items: Question[]): Question[] {
		const sorted = [...items];
		sorted.sort((left, right) => {
			if (sortOrder === 'oldest') return compareDate(left.created_at, right.created_at) || compareText(left.code, right.code);
			if (sortOrder === 'newest') return compareDate(right.created_at, left.created_at) || compareText(left.code, right.code);
			if (sortOrder === 'code_desc') return compareText(right.code, left.code);
			if (sortOrder === 'code_asc') return compareText(left.code, right.code);
			if (sortOrder === 'author_subject_code_asc') {
				return (
					compareText(left.author_display_name || left.author_username, right.author_display_name || right.author_username) ||
					compareText(left.subject_code || left.subject_name, right.subject_code || right.subject_name) ||
					compareText(left.code, right.code)
				);
			}
			if (sortOrder === 'status_subject_code_asc') {
				return (
					compareText(workflowLabel(left.workflow_status ?? left.status), workflowLabel(right.workflow_status ?? right.status)) ||
					compareText(left.subject_code || left.subject_name, right.subject_code || right.subject_name) ||
					compareText(left.code, right.code)
				);
			}
			if (sortOrder === 'type_subject_code_asc') {
				return (
					compareText(questionTypeLabel(left.question_type), questionTypeLabel(right.question_type)) ||
					compareText(left.subject_code || left.subject_name, right.subject_code || right.subject_name) ||
					compareText(left.code, right.code)
				);
			}
			return compareText(left.subject_code || left.subject_name, right.subject_code || right.subject_name) || compareText(left.code, right.code);
		});
		return sorted;
	}

	function formatDate(value: string | undefined): string {
		if (!value) return '-';
		const date = new Date(value);
		if (Number.isNaN(date.getTime())) return value;
		return new Intl.DateTimeFormat('id-ID', {
			day: '2-digit',
			month: 'short',
			year: 'numeric'
		}).format(date);
	}

	function formatDateTime(value: Date = new Date()): string {
		return new Intl.DateTimeFormat('id-ID', {
			day: '2-digit',
			month: 'long',
			year: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		}).format(value);
	}

	function compactText(value: string | undefined | null, fallback = '-'): string {
		const trimmed = (value ?? '').trim();
		return trimmed || fallback;
	}

	function escapeHtml(value: string): string {
		return value
			.replace(/&/g, '&amp;')
			.replace(/</g, '&lt;')
			.replace(/>/g, '&gt;')
			.replace(/"/g, '&quot;')
			.replace(/'/g, '&#39;')
			.replace(/\n/g, '<br>');
	}

	function richQuestionHtml(question: Question): string {
		return compactText(
			question.stem_html || question.stem_latex || question.question_text,
			'<em>Isi soal belum tersedia</em>'
		);
	}

	function richText(value: string | undefined | null): string {
		const trimmed = (value ?? '').trim();
		if (!trimmed) return '';
		if (/[<][a-zA-Z/][\s\S]*[>]/.test(trimmed) || trimmed.includes('$')) return trimmed;
		return `<p>${escapeHtml(trimmed)}</p>`;
	}

	function questionTypeLabel(value: string | undefined): string {
		return questionTypeOptions.find((option) => option.value === value)?.label ?? compactText(value, 'Tipe belum dipilih');
	}

	function workflowLabel(value: string | undefined): string {
		return workflowOptions.find((option) => option.value === value)?.label ?? compactText(value, 'Belum ada status');
	}

	function difficultyLabel(value: string | undefined): string {
		if (value === 'easy') return 'Mudah';
		if (value === 'medium') return 'Sedang';
		if (value === 'hard') return 'Sulit';
		return compactText(value, 'Kesulitan belum diisi');
	}

	function subjectLabel(question: Question): string {
		const code = compactText(question.subject_code, '');
		const name = compactText(question.subject_name, '');
		if (code && name) return `${code} - ${name}`;
		return name || code || 'Mapel belum tersedia';
	}

	function levelLabel(question: Question): string {
		return compactText(question.target_level || question.academic_phase, 'Tingkat belum diisi');
	}

	function optionRows(question: Question) {
		const rows = Array.isArray(question.options)
			? question.options
					.map((option, index) => ({
						label: compactText(option.label || String.fromCharCode(65 + index), String.fromCharCode(65 + index)),
						html: compactText(option.html || option.latex || option.text, '')
					}))
					.filter((option) => option.html)
			: [];
		if (rows.length > 0) return rows;
		return [
			['A', question.option_a],
			['B', question.option_b],
			['C', question.option_c],
			['D', question.option_d],
			['E', question.option_e]
		]
			.map(([label, html]) => ({ label: String(label), html: compactText(html, '') }))
			.filter((option) => option.html);
	}

	function metadataRows(question: Question) {
		return [
			['Materi', question.material_topic],
			['Level Kognitif', question.cognitive_level],
			['Kesulitan', difficultyLabel(question.difficulty)],
			['HOTS', question.hots_flag ? 'Ya' : 'Tidak'],
			['CP', question.cp_ref],
			['TP', question.tp_ref],
			['KD', question.kd_ref],
			['Indikator', question.indicator_ref]
		].filter(([, value]) => compactText(value as string | undefined, '') !== '');
	}

	function printPage() {
		if (loadedHasMore) {
			const ok = window.confirm(
				`Yang tampil baru ${questions.length} dari ${loadedTotal} soal. Cetak halaman ini saja? Pilih batas "Semua hasil filter" jika ingin mencetak seluruh hasil.`
			);
			if (!ok) return;
		}
		window.print();
	}

	function clearLoadedQuestions() {
		questions = [];
		loadedTotal = 0;
		hasLoadedQuestions = false;
	}
</script>

<svelte:head>
	<title>Cetak Soal Saya - Bank Soal - MTsN 2 Kolaka Utara</title>
</svelte:head>

<div class="print-shell space-y-5">
	<section class="no-print rounded-xl border border-border bg-card p-4 shadow-sm md:p-5">
		<div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
			<div>
				<p class="text-[11px] font-semibold uppercase tracking-[0.18em] text-primary">Bank Soal</p>
				<h1 class="mt-1 text-2xl font-semibold tracking-tight text-foreground md:text-3xl">Cetak Soal Saya</h1>
				<p class="mt-2 max-w-3xl text-sm leading-6 text-muted-foreground">
					{canPrintOtherAuthors
						? 'Admin dapat mencetak soal milik semua guru atau memilih guru tertentu.'
						: 'Cetak soal yang Bapak/Ibu input sendiri.'} Kunci jawaban, pembahasan, dan metadata bisa ditampilkan sesuai kebutuhan.
				</p>
			</div>
			<div class="flex flex-wrap gap-2">
				<Button href={listHref} variant="outline"><ArrowLeftIcon class="size-4" />Kembali</Button>
				<Button variant="outline" onclick={() => void loadReferenceData()} disabled={loading}>
					<RefreshCcwIcon class="size-4" />Muat Referensi
				</Button>
				<Button onclick={printPage} disabled={loading || filteredQuestions.length === 0}>
					<PrinterIcon class="size-4" />Print / Simpan PDF
				</Button>
			</div>
		</div>

		<div class="mt-4 grid gap-3 border-t border-border/60 pt-4 sm:grid-cols-2 lg:grid-cols-4">
			{#if canPrintOtherAuthors}
				<label class="space-y-1.5 text-sm">
					<span class="text-xs font-semibold text-muted-foreground">Guru/Pembuat</span>
					<select bind:value={authorFilter} onchange={clearLoadedQuestions} class="h-9 w-full rounded-md border border-border bg-background px-3 text-sm">
						<option value="">Semua guru</option>
						{#each authors as author (author.username)}
							<option value={author.username}>{author.display_name || author.username}</option>
						{/each}
					</select>
				</label>
			{/if}
			<label class="space-y-1.5 text-sm">
				<span class="text-xs font-semibold text-muted-foreground">Mapel</span>
				<select bind:value={subjectFilter} onchange={clearLoadedQuestions} class="h-9 w-full rounded-md border border-border bg-background px-3 text-sm">
					<option value="">Semua mapel</option>
					{#each subjects as subject (subject.id)}
						<option value={subject.id}>{subject.code ? `${subject.code} - ${subject.name}` : subject.name}</option>
					{/each}
				</select>
			</label>
			<div class="relative space-y-1.5 text-sm">
				<span id="print-workflow-filter-label" class="text-xs font-semibold text-muted-foreground">Status</span>
				<button
					type="button"
					class="flex h-9 w-full items-center justify-between rounded-md border border-border bg-background px-3 text-left text-sm"
					aria-labelledby="print-workflow-filter-label"
					aria-expanded={workflowDropdownOpen}
					onclick={() => (workflowDropdownOpen = !workflowDropdownOpen)}
				>
					<span class="truncate">{workflowFilterSummary()}</span>
					<span class="text-muted-foreground">▾</span>
				</button>
				{#if workflowDropdownOpen}
					<div class="absolute z-30 mt-1 max-h-72 w-full overflow-y-auto rounded-md border border-border bg-popover p-2 text-popover-foreground shadow-lg">
						<button type="button" class="mb-1 w-full rounded px-2 py-1.5 text-left text-sm hover:bg-muted" onclick={clearWorkflowFilters}>
							Semua status
						</button>
						{#each workflowSelectableOptions as option (option.value)}
							<label class="flex cursor-pointer items-center gap-2 rounded px-2 py-1.5 hover:bg-muted">
								<input
									type="checkbox"
									checked={workflowFilters.includes(option.value)}
									onchange={() => toggleWorkflowFilter(option.value)}
								/>
								<span>{option.label}</span>
							</label>
						{/each}
					</div>
				{/if}
			</div>
			<label class="space-y-1.5 text-sm">
				<span class="text-xs font-semibold text-muted-foreground">Tipe</span>
				<select bind:value={typeFilter} onchange={clearLoadedQuestions} class="h-9 w-full rounded-md border border-border bg-background px-3 text-sm">
					{#each questionTypeOptions as option (option.value)}
						<option value={option.value}>{option.label}</option>
					{/each}
				</select>
			</label>
			<div class="grid grid-cols-2 gap-2">
				<label class="space-y-1.5 text-sm">
					<span class="text-xs font-semibold text-muted-foreground">Dari</span>
					<input type="date" bind:value={dateFrom} onchange={clearLoadedQuestions} class="h-9 w-full rounded-md border border-border bg-background px-3 text-sm" />
				</label>
				<label class="space-y-1.5 text-sm">
					<span class="text-xs font-semibold text-muted-foreground">Sampai</span>
					<input type="date" bind:value={dateTo} onchange={clearLoadedQuestions} class="h-9 w-full rounded-md border border-border bg-background px-3 text-sm" />
				</label>
			</div>
			<label class="space-y-1.5 text-sm">
				<span class="text-xs font-semibold text-muted-foreground">Urutkan</span>
				<select bind:value={sortOrder} class="h-9 w-full rounded-md border border-border bg-background px-3 text-sm">
					{#each sortOptions as option (option.value)}
						<option value={option.value}>{option.label}</option>
					{/each}
				</select>
			</label>
			<label class="space-y-1.5 text-sm">
				<span class="text-xs font-semibold text-muted-foreground">Batas tampil/cetak</span>
				<select bind:value={resultLimit} onchange={clearLoadedQuestions} class="h-9 w-full rounded-md border border-border bg-background px-3 text-sm">
					{#each resultLimitOptions as option (option.value)}
						<option value={option.value}>{option.label}</option>
					{/each}
				</select>
			</label>
		</div>

		{#if workflowFilters.length > 0}
			<div class="mt-3 flex flex-wrap items-center gap-2 text-xs">
				<span class="font-semibold text-muted-foreground">Status dipilih:</span>
				{#each workflowFilters as status (status)}
					<button
						type="button"
						class="rounded-full border border-primary/30 bg-primary/10 px-2.5 py-1 text-primary hover:bg-primary/15"
						onclick={() => toggleWorkflowFilter(status)}
					>
						{workflowLabel(status)} ×
					</button>
				{/each}
				<button type="button" class="text-muted-foreground underline-offset-4 hover:underline" onclick={clearWorkflowFilters}>Reset status</button>
			</div>
		{/if}

		{#if shouldWarnLargeSelection}
			<div class="mt-4 rounded-lg border border-amber-300 bg-amber-50 p-3 text-sm leading-6 text-amber-900">
				Saran: untuk performa terbaik, pilih guru/mapel/status terlebih dahulu atau gunakan batas 50/100/200 soal.
				Opsi <span class="font-semibold">Semua hasil filter</span> hanya disarankan setelah filter cukup spesifik.
			</div>
		{/if}

		<div class="mt-4 flex flex-wrap gap-3 border-t border-border/60 pt-4 text-sm">
			<label class="inline-flex items-center gap-2 rounded-md border border-border px-3 py-2">
				<input type="checkbox" bind:checked={includeMetadata} />
				<span>Tampilkan metadata</span>
			</label>
			<label class="inline-flex items-center gap-2 rounded-md border border-border px-3 py-2">
				<input type="checkbox" bind:checked={includeAnswer} />
				<span>Tampilkan kunci jawaban</span>
			</label>
			<label class="inline-flex items-center gap-2 rounded-md border border-border px-3 py-2">
				<input type="checkbox" bind:checked={includeExplanation} />
				<span>Tampilkan pembahasan/rubrik</span>
			</label>
			<Button onclick={() => void loadPrintData()} disabled={loading} class="ml-auto">
				<RefreshCcwIcon class="size-4" />Terapkan & Muat Soal
			</Button>
		</div>
	</section>

	{#if error}
		<div class="no-print rounded-xl border border-destructive/30 bg-destructive/10 p-4 text-sm text-destructive">{error}</div>
	{:else if loading}
		<div class="no-print rounded-xl border border-border bg-card p-6 text-sm text-muted-foreground">Memuat data cetak...</div>
	{:else if !hasLoadedQuestions}
		<div class="no-print rounded-xl border border-dashed border-border bg-card p-6 text-sm leading-6 text-muted-foreground">
			Pilih guru/mapel/status/tipe/tanggal dan urutan, lalu klik <span class="font-semibold text-foreground">Terapkan & Muat Soal</span>.
			Halaman ini tidak lagi memuat dan merender semua soal otomatis agar tetap ringan. Default batas cetak adalah 100 soal; pilih batas lebih kecil/besar sesuai kebutuhan.
		</div>
	{:else}
		<section class="print-document rounded-xl border border-border bg-white p-6 text-slate-950 shadow-sm">
			<header class="border-b border-slate-300 pb-4 text-center">
				<p class="text-sm font-semibold uppercase tracking-[0.18em]">MTsN 2 Kolaka Utara</p>
				<h2 class="mt-1 text-xl font-bold">{printTitle}</h2>
				<p class="mt-1 text-sm text-slate-600">{selectedAuthorName} · {selectedSubjectName} · {selectedSortName} · Dicetak {formatDateTime()}</p>
			</header>

			<div class="mt-4 grid gap-2 text-sm sm:grid-cols-2 lg:grid-cols-4">
				<div><span class="font-semibold">Pembuat:</span> {selectedAuthorName}</div>
				<div><span class="font-semibold">Jumlah soal:</span> {filteredQuestions.length}{loadedTotal > filteredQuestions.length ? ` dari ${loadedTotal}` : ''}</div>
				<div><span class="font-semibold">Batas:</span> {selectedLimitName}</div>
				<div><span class="font-semibold">Status:</span> {workflowFilterSummary()}</div>
				<div><span class="font-semibold">Tipe:</span> {questionTypeOptions.find((option) => option.value === typeFilter)?.label ?? 'Semua tipe'}</div>
				<div><span class="font-semibold">Kunci:</span> {includeAnswer ? 'Ditampilkan' : 'Tidak ditampilkan'}</div>
				<div><span class="font-semibold">Urutan:</span> {selectedSortName}</div>
			</div>

			{#if loadedHasMore}
				<div class="no-print mt-4 rounded-lg border border-amber-300 bg-amber-50 p-3 text-sm leading-6 text-amber-900">
					Ditampilkan {questions.length} dari {loadedTotal} soal sesuai filter. Untuk mencetak semua, ubah batas ke <span class="font-semibold">Semua hasil filter</span> lalu muat ulang.
				</div>
			{/if}

			{#if filteredQuestions.length === 0}
				<div class="mt-8 rounded-lg border border-dashed border-slate-300 p-6 text-center text-sm text-slate-600">
					Belum ada soal sesuai filter cetak.
				</div>
			{:else}
				<ol class="mt-6 space-y-6">
					{#each filteredQuestions as question, index (question.id)}
						<li class="question-item break-inside-avoid rounded-lg border border-slate-300 p-4">
							<div class="flex flex-wrap items-center gap-2 text-xs text-slate-600">
								<Badge variant="outline" class="border-slate-300 text-slate-700">{compactText(question.code, `Soal ${index + 1}`)}</Badge>
								<span>{subjectLabel(question)}</span>
								<span>·</span>
								<span>{levelLabel(question)}</span>
								<span>·</span>
								<span>{questionTypeLabel(question.question_type)}</span>
								<span>·</span>
								<span>{workflowLabel(question.workflow_status ?? question.status)}</span>
								<span>·</span>
								<span>Dibuat {formatDate(question.created_at)}</span>
							</div>

							<div class="mt-3 flex gap-3">
								<div class="shrink-0 font-bold">{index + 1}.</div>
								<div class="min-w-0 flex-1">
									{#if question.stimulus_html}
										<div class="mb-3 rounded-md border border-slate-200 bg-slate-50 p-3 text-sm">
											<RichContent html={question.stimulus_html} />
										</div>
									{/if}
									<RichContent html={richQuestionHtml(question)} class="text-[15px] leading-7" />

									{#if optionRows(question).length > 0}
										<ol class="mt-3 space-y-2">
											{#each optionRows(question) as option (option.label)}
												<li class="flex gap-2 text-sm">
													<span class="font-semibold">{option.label}.</span>
													<RichContent html={richText(option.html)} class="min-w-0 flex-1" />
												</li>
											{/each}
										</ol>
									{:else if question.question_type === 'essay' || question.question_type === 'short_answer'}
										<div class="mt-4 space-y-2 text-sm text-slate-500">
											<div class="border-b border-dotted border-slate-400 pb-5">Jawaban:</div>
											<div class="border-b border-dotted border-slate-400 pb-5"></div>
										</div>
									{/if}

									{#if includeAnswer && question.answer_key}
										<div class="mt-3 rounded-md border border-emerald-300 bg-emerald-50 p-2 text-sm"><span class="font-semibold">Kunci:</span> {question.answer_key}</div>
									{/if}

									{#if includeExplanation && (question.explanation_html || question.explanation || question.rubric_html)}
										<div class="mt-3 rounded-md border border-slate-200 bg-slate-50 p-3 text-sm">
											<p class="mb-1 font-semibold">Pembahasan / Rubrik</p>
											{#if question.explanation_html || question.explanation}
												<RichContent html={richText(question.explanation_html || question.explanation)} />
											{/if}
											{#if question.rubric_html}
												<RichContent html={question.rubric_html} />
											{/if}
										</div>
									{/if}

									{#if includeMetadata && metadataRows(question).length > 0}
										<dl class="mt-3 grid gap-x-4 gap-y-1 rounded-md border border-slate-200 bg-slate-50 p-3 text-xs sm:grid-cols-[9rem_1fr]">
											{#each metadataRows(question) as row (`${question.id}-${row[0]}`)}
												<dt class="font-semibold text-slate-600">{row[0]}</dt>
												<dd>{row[1]}</dd>
											{/each}
										</dl>
									{/if}
								</div>
							</div>
						</li>
					{/each}
				</ol>
			{/if}
		</section>
	{/if}
</div>

<style>
	@media print {
		:global(body) {
			background: #fff !important;
		}

		.no-print,
		:global(header),
		:global(nav),
		:global(aside) {
			display: none !important;
		}

		.print-shell {
			margin: 0 !important;
			padding: 0 !important;
		}

		.print-document {
			border: 0 !important;
			box-shadow: none !important;
			padding: 0 !important;
		}

		.question-item {
			page-break-inside: avoid;
			break-inside: avoid;
		}
	}
</style>
