<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import CheckCircle2Icon from '@lucide/svelte/icons/check-circle-2';
	import EyeIcon from '@lucide/svelte/icons/eye';
	import RefreshCcwIcon from '@lucide/svelte/icons/refresh-ccw';
	import SearchIcon from '@lucide/svelte/icons/search';
	import SendIcon from '@lucide/svelte/icons/send';
	import ShieldCheckIcon from '@lucide/svelte/icons/shield-check';
	import * as Table from '$lib/components/ui/table';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { TablePagination } from '$lib/components/ui/pagination';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import RichContent from '$lib/components/RichContent.svelte';
	import { toast } from '$lib/components/ui/sonner';
	import { canPublishBankSoal, type BankSoalAccessUser } from '$lib/bank-soal/access';
	import { clientApiPath, clientApiPathWithQuery, readClientApiData, readClientJson } from '$lib/client/api';
	import { displayName } from '$lib/utils/display-name';
	import { htmlToPlainText } from '$lib/utils/html-text';
	import { DEFAULT_PAGE_SIZE_OPTIONS, normalizePage, normalizePageSize, type PaginationChange } from '$lib/utils/pagination';

	type PageData = { user?: BankSoalAccessUser };
	type OptionItem = { label?: string; text?: string; html?: string; latex?: string; match_label?: string; match_text?: string; match_html?: string; is_distractor?: boolean };
	type Question = {
		id: string;
		event_id?: string | null;
		subject_id?: string;
		subject_name?: string;
		subject_code?: string;
		code?: string;
		question_type?: string;
		stem_html?: string;
		question_text?: string;
		stimulus_html?: string;
		explanation_html?: string;
		rubric_html?: string;
		options?: OptionItem[];
		answer_key?: string;
		difficulty?: string;
		workflow_status?: string;
		status?: string;
		author_username?: string;
		author_display_name?: string;
		reviewer_username?: string;
		reviewer_display_name?: string;
		reviewed_at?: string | null;
		review_notes?: string;
		material_topic?: string;
		cognitive_level?: string;
		target_level?: string | null;
		hots_flag?: boolean;
		created_at?: string;
		package_count?: number;
		answer_count?: number;
		usage?: { package_count?: number; answer_count?: number; is_locked?: boolean };
	};
	type Subject = { id: string; name: string; code?: string };
	type QuestionListResponse = { items?: Question[]; meta?: { total?: number; limit?: number; offset?: number } };
	type AcademicPayload = { subjects?: Subject[] };
	type BulkWorkflowResult = { question_id?: string; id?: string; ok?: boolean; success?: boolean; error?: string; message?: string };
	type BulkWorkflowResponse = { results?: BulkWorkflowResult[] } | BulkWorkflowResult[];
	type PublishOverview = { questions: Question[]; subjects: Subject[]; total: number; page: number; limit: number };

	let { data }: { data?: PageData } = $props();

	const DEFAULT_PAGE_SIZE = DEFAULT_PAGE_SIZE_OPTIONS[0];
	const initialSubject = page.url.searchParams.get('subject_id') ?? '';
	const initialSearch = page.url.searchParams.get('q') ?? '';
	const initialLevel = page.url.searchParams.get('target_level') ?? '';
	const initialPage = normalizePage(page.url.searchParams.get('page'), 1);
	const initialLimit = normalizePageSize(page.url.searchParams.get('limit'), DEFAULT_PAGE_SIZE_OPTIONS, DEFAULT_PAGE_SIZE);

	let queuePromise = $state<Promise<PublishOverview> | null>(null);
	let questions = $state<Question[]>([]);
	let subjects = $state<Subject[]>([]);
	let totalItems = $state(0);
	let currentPage = $state(initialPage);
	let pageSize = $state<number>(initialLimit);
	let search = $state(initialSearch);
	let subjectFilter = $state(initialSubject);
	let targetLevelFilter = $state(initialLevel);
	let selectedIds = $state<Set<string>>(new Set());
	let busyId = $state('');
	let bulkBusy = $state(false);

	let canPublish = $derived(canPublishBankSoal(data?.user));
	let selectedCount = $derived(selectedIds.size);
	let selectedQuestions = $derived(questions.filter((question) => selectedIds.has(question.id)));
	let resultStart = $derived(totalItems === 0 ? 0 : (currentPage - 1) * pageSize + 1);
	let resultEnd = $derived(Math.min(totalItems, (currentPage - 1) * pageSize + questions.length));
	let readyWithMetadata = $derived(questions.filter((question) => publishChecklist(question).every((item) => item.ok)).length);

	const levelOptions = ['', 'VII', 'VIII', 'IX'];
	const difficultyLabels: Record<string, string> = { easy: 'Mudah', medium: 'Sedang', hard: 'Sulit' };
	const questionTypeLabels: Record<string, string> = {
		multiple_choice: 'PG',
		multiple_answer: 'PG Kompleks',
		true_false: 'Benar/Salah',
		agree_disagree: 'Setuju/Tidak',
		matching: 'Menjodohkan',
		short_answer: 'Isian Singkat',
		essay: 'Uraian'
	};

	function errorMessage(error: unknown, fallback: string) {
		return error instanceof Error && error.message.trim() ? error.message : fallback;
	}

	function questionTypeLabel(value?: string) {
		return questionTypeLabels[value ?? ''] ?? value ?? 'Soal';
	}

	function stemPreview(question: Question): string {
		return htmlToPlainText(question.stem_html || question.question_text || '').slice(0, 220) || '(Soal kosong)';
	}

	function authorLabel(question: Question): string {
		return displayName({ display_name: question.author_display_name, username: question.author_username }, 'Guru');
	}

	function reviewerLabel(question: Question): string {
		return displayName({ display_name: question.reviewer_display_name, username: question.reviewer_username }, 'Reviewer');
	}

	function usageCount(question: Question): number {
		return question.package_count ?? question.usage?.package_count ?? 0;
	}

	function formatDate(value?: string | null): string {
		if (!value) return '—';
		const date = new Date(value);
		if (Number.isNaN(date.getTime())) return '—';
		return date.toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' });
	}

	function publishChecklist(question: Question) {
		return [
			{ label: 'Siap Pakai', ok: question.workflow_status === 'siap_pakai' },
			{ label: 'Belum terbit', ok: (question.status ?? 'draft') !== 'published' },
			{ label: 'Mapel', ok: Boolean(question.subject_name || question.subject_code || question.subject_id) },
			{ label: 'Naskah', ok: stemPreview(question).length >= 5 },
			{ label: 'Kunci/Rubrik', ok: Boolean(question.answer_key || question.rubric_html) },
			{ label: 'Reviewer', ok: Boolean(question.reviewer_display_name || question.reviewer_username) }
		];
	}

	function buildParams(pageNumber: number, limit = pageSize): URLSearchParams {
		const params = new URLSearchParams({ workflow_status: 'siap_pakai', status: 'draft', limit: String(limit), offset: String(Math.max(0, (pageNumber - 1) * limit)) });
		if (search.trim()) params.set('q', search.trim());
		if (subjectFilter) params.set('subject_id', subjectFilter);
		if (targetLevelFilter) params.set('target_level', targetLevelFilter);
		return params;
	}

	async function fetchSubjects(): Promise<Subject[]> {
		const payload = await fetch('/api/bank-soal/soal-support/subjects').then((response) => readClientApiData<AcademicPayload>(response, 'Gagal memuat data mapel'));
		return payload.subjects ?? [];
	}

	async function fetchQueue(pageNumber: number, limit = pageSize): Promise<PublishOverview> {
		const [questionPayload, loadedSubjects] = await Promise.all([
			fetch(clientApiPathWithQuery('/api/bank-soal/questions', buildParams(pageNumber, limit))).then((response) => readClientApiData<QuestionListResponse>(response, 'Gagal memuat antrean penerbitan')),
			subjects.length > 0 ? Promise.resolve(subjects) : fetchSubjects()
		]);
		return {
			questions: questionPayload.items ?? [],
			subjects: loadedSubjects,
			total: questionPayload.meta?.total ?? questionPayload.items?.length ?? 0,
			page: pageNumber,
			limit: questionPayload.meta?.limit ?? limit
		};
	}

	function syncUrl(pageNumber: number, limit = pageSize) {
		if (typeof window === 'undefined') return;
		const params = new URLSearchParams();
		if (search.trim()) params.set('q', search.trim());
		if (subjectFilter) params.set('subject_id', subjectFilter);
		if (targetLevelFilter) params.set('target_level', targetLevelFilter);
		if (pageNumber > 1) params.set('page', String(pageNumber));
		if (limit !== DEFAULT_PAGE_SIZE) params.set('limit', String(limit));
		const query = params.toString();
		window.history.replaceState({}, '', `${window.location.pathname}${query ? `?${query}` : ''}`);
	}

	function load(pageNumber = currentPage, limit = pageSize) {
		pageSize = limit;
		queuePromise = fetchQueue(pageNumber, limit).then((overview) => {
			questions = overview.questions;
			subjects = overview.subjects;
			totalItems = overview.total;
			currentPage = overview.page;
			pageSize = normalizePageSize(overview.limit, DEFAULT_PAGE_SIZE_OPTIONS, limit);
			selectedIds = new Set([...selectedIds].filter((id) => overview.questions.some((question) => question.id === id)));
			syncUrl(overview.page, pageSize);
			return overview;
		});
	}

	function applyFilters() {
		currentPage = 1;
		selectedIds = new Set();
		load(1, pageSize);
	}

	function handlePagination(change: PaginationChange) {
		load(change.reason === 'limit' ? 1 : change.page, change.limit);
	}

	function reloadAfterRemoving(removedCount: number) {
		const remaining = Math.max(0, totalItems - removedCount);
		const nextPage = Math.min(currentPage, Math.max(1, Math.ceil(remaining / pageSize)));
		load(nextPage, pageSize);
	}

	function clearFilters() {
		search = '';
		subjectFilter = '';
		targetLevelFilter = '';
		applyFilters();
	}

	function toggleSelection(id: string, checked: boolean) {
		const next = new Set(selectedIds);
		if (checked) next.add(id);
		else next.delete(id);
		selectedIds = next;
	}

	function togglePageSelection(checked: boolean) {
		const next = new Set(selectedIds);
		for (const question of questions) {
			if (checked) next.add(question.id);
			else next.delete(question.id);
		}
		selectedIds = next;
	}

	async function publishOne(question: Question) {
		if (!canPublish) {
			toast.warning('Penerbitan soal membutuhkan izin publish Bank Soal.');
			return;
		}
		busyId = question.id;
		try {
			await fetch(clientApiPath`/api/bank-soal/questions/${question.id}/workflow`, {
				method: 'PATCH',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ action: 'publish' })
			}).then((response) => readClientJson<unknown>(response));
			toast.success('Soal diterbitkan');
			selectedIds = new Set([...selectedIds].filter((id) => id !== question.id));
			reloadAfterRemoving(1);
		} catch (error) {
			toast.error(errorMessage(error, 'Gagal menerbitkan soal'));
		} finally {
			busyId = '';
		}
	}

	async function publishSelected() {
		if (!canPublish) {
			toast.warning('Penerbitan soal membutuhkan izin publish Bank Soal.');
			return;
		}
		if (selectedIds.size === 0) return;
		bulkBusy = true;
		try {
			const payload = await fetch(clientApiPath`/api/bank-soal/questions/bulk-workflow`, {
				method: 'PATCH',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ action: 'publish', question_ids: [...selectedIds] })
			}).then((response) => readClientJson<BulkWorkflowResponse>(response));
			const results = Array.isArray(payload) ? payload : payload.results ?? [];
			const failed = results.filter((item) => item.ok === false || item.success === false || item.error || item.message?.toLowerCase().includes('gagal'));
			if (failed.length > 0) toast.warning(`${failed.length} soal belum bisa diterbitkan. Periksa ulang antrean.`);
			else toast.success(`${selectedIds.size} soal diterbitkan`);
			const removedCount = Math.max(0, selectedIds.size - failed.length);
			selectedIds = new Set();
			reloadAfterRemoving(removedCount);
		} catch (error) {
			toast.error(errorMessage(error, 'Gagal menerbitkan soal terpilih'));
		} finally {
			bulkBusy = false;
		}
	}

	onMount(() => load(currentPage));
</script>

<svelte:head><title>Penerbitan Bank Soal</title></svelte:head>

<div class="space-y-4">
	<section class="overflow-hidden rounded-2xl border border-success/20 bg-gradient-to-br from-success/10 via-card to-card p-4 shadow-sm md:p-6">
		<div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
			<div class="min-w-0">
				<div class="flex flex-wrap items-center gap-2">
					<span class="inline-flex size-10 items-center justify-center rounded-xl bg-success/15 text-success"><ShieldCheckIcon class="size-5" /></span>
					<div>
						<p class="text-[11px] font-black uppercase tracking-[0.28em] text-success">Ruang Penerbitan Soal</p>
						<h1 class="mt-1 text-2xl font-black text-foreground md:text-3xl">Terbitkan soal siap pakai</h1>
					</div>
				</div>
				<p class="mt-3 max-w-3xl text-sm leading-6 text-muted-foreground">
					Halaman ini khusus untuk antrean soal <strong>Siap Pakai</strong> yang belum berstatus <strong>Terbit</strong>. Pemeriksaan tetap dilakukan di ruang verifikasi; di sini admin/panitia memastikan kesiapan operasional sebelum soal boleh dipakai paket asesmen.
				</p>
			</div>
			<div class="flex flex-wrap gap-2 lg:justify-end">
				<a href={resolve('/bank-soal/verifikasi')} class="inline-flex h-9 items-center rounded-md border border-border bg-card px-3 text-sm font-semibold text-foreground hover:bg-muted/50">Pemeriksaan Soal</a>
				<a href={resolve('/bank-soal?workflow_status=siap_pakai&status=draft')} class="inline-flex h-9 items-center rounded-md border border-success/20 bg-card px-3 text-sm font-semibold text-success hover:bg-success/10">Daftar Siap Pakai</a>
			</div>
		</div>
	</section>

	<section class="grid gap-3 md:grid-cols-3">
		<div class="rounded-xl border border-border bg-card p-4 shadow-sm">
			<p class="text-xs font-semibold text-muted-foreground">Antrean siap terbit</p>
			<p class="mt-2 text-3xl font-black text-foreground">{totalItems}</p>
			<p class="mt-1 text-xs text-muted-foreground">Soal siap pakai + belum terbit</p>
		</div>
		<div class="rounded-xl border border-border bg-card p-4 shadow-sm">
			<p class="text-xs font-semibold text-muted-foreground">Siap secara metadata</p>
			<p class="mt-2 text-3xl font-black text-success">{readyWithMetadata}</p>
			<p class="mt-1 text-xs text-muted-foreground">Dari halaman antrean saat ini</p>
		</div>
		<div class="rounded-xl border border-border bg-card p-4 shadow-sm">
			<p class="text-xs font-semibold text-muted-foreground">Terpilih</p>
			<p class="mt-2 text-3xl font-black text-primary">{selectedCount}</p>
			<p class="mt-1 text-xs text-muted-foreground">Untuk bulk publish</p>
		</div>
	</section>

	<section class="rounded-xl border border-border bg-card p-4 shadow-sm">
		<form class="flex flex-col gap-3 lg:flex-row lg:items-end" onsubmit={(event) => { event.preventDefault(); applyFilters(); }}>
			<div class="min-w-0 flex-1">
				<label for="publish-search" class="mb-1 block text-xs font-semibold text-muted-foreground">Cari soal</label>
				<div class="relative">
					<SearchIcon class="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
					<Input id="publish-search" bind:value={search} placeholder="Kode, materi, atau isi soal..." class="pl-9" />
				</div>
			</div>
			<div class="min-w-[12rem]">
				<label for="publish-subject" class="mb-1 block text-xs font-semibold text-muted-foreground">Mapel</label>
				<select id="publish-subject" bind:value={subjectFilter} class="h-10 w-full rounded-md border border-border bg-card px-3 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring">
					<option value="">Semua mapel</option>
					{#each subjects as subject (subject.id)}
						<option value={subject.id}>{subject.name || subject.code}</option>
					{/each}
				</select>
			</div>
			<div class="min-w-[9rem]">
				<label for="publish-level" class="mb-1 block text-xs font-semibold text-muted-foreground">Tingkat</label>
				<select id="publish-level" bind:value={targetLevelFilter} class="h-10 w-full rounded-md border border-border bg-card px-3 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring">
					{#each levelOptions as level (level || 'all')}
						<option value={level}>{level || 'Semua tingkat'}</option>
					{/each}
				</select>
			</div>
			<div class="flex gap-2">
				<LoadingButton type="submit" loading={false} class="h-10">Terapkan</LoadingButton>
				<Button type="button" variant="outline" class="h-10" onclick={clearFilters}>Reset</Button>
				<LoadingButton type="button" variant="outline" class="h-10" loading={false} onclick={() => load(currentPage)} aria-label="Muat ulang antrean"><RefreshCcwIcon class="size-4" /></LoadingButton>
			</div>
		</form>
	</section>

	{#if selectedCount > 0}
		<section class="flex flex-col gap-3 rounded-xl border border-success/20 bg-success/10 p-3 text-success shadow-sm md:flex-row md:items-center md:justify-between">
			<div>
				<p class="text-sm font-bold">{selectedCount} soal dipilih</p>
				<p class="text-xs opacity-80">{selectedQuestions.slice(0, 3).map((question) => question.code || stemPreview(question).slice(0, 24)).join(', ')}{selectedCount > 3 ? '…' : ''}</p>
			</div>
			<div class="flex flex-wrap gap-2">
				<Button type="button" variant="outline" class="bg-card text-success" onclick={() => (selectedIds = new Set())}>Batal pilih</Button>
				<LoadingButton type="button" loading={bulkBusy} loadingLabel="Menerbitkan..." disabled={!canPublish || bulkBusy} onclick={() => void publishSelected()} class="bg-success text-background hover:bg-success">
					<SendIcon class="size-4" />
					Terbitkan Terpilih
				</LoadingButton>
			</div>
		</section>
	{/if}

	<AsyncContent promise={queuePromise}>
		{#snippet pending()}
			<section class="rounded-xl border border-border bg-card p-4 shadow-sm">
				<div class="space-y-3">
					{#each Array.from({ length: 5 }) as _, index (`publish-skeleton-${index}`)}
						<div class="rounded-lg border border-border p-3"><Skeleton class="h-4 w-36" /><Skeleton class="mt-3 h-5 max-w-2xl" /><Skeleton class="mt-3 h-4 w-64" /></div>
					{/each}
				</div>
			</section>
		{/snippet}
		{#snippet failed(error, reset)}
			<section class="rounded-xl border border-border bg-card p-5 shadow-sm">
				<RecoveryPanel title="Antrean penerbitan belum dapat dimuat" message={error instanceof Error ? error.message : 'Gagal memuat antrean penerbitan'} onRetry={() => { reset?.(); load(currentPage); }} />
			</section>
		{/snippet}
		{#snippet children(value)}
			{@const overview = value as PublishOverview}
			<section class="overflow-hidden rounded-xl border border-border bg-card shadow-sm">
				<div class="flex flex-col gap-2 border-b border-border p-4 md:flex-row md:items-center md:justify-between">
					<div>
						<h2 class="text-base font-semibold text-foreground">Antrean Penerbitan</h2>
						<p class="mt-1 text-xs text-muted-foreground">{overview.total === 0 ? 'Tidak ada soal siap terbit pada filter ini' : `${resultStart}-${resultEnd} dari ${overview.total} soal`}</p>
					</div>
					<Badge variant="outline" class="border-success/20 bg-success/10 text-success">Siap Pakai -> Terbit</Badge>
				</div>

				{#if overview.questions.length === 0}
					<div class="flex flex-col items-center px-6 py-12 text-center">
						<div class="flex size-12 items-center justify-center rounded-xl bg-success/10 text-success"><CheckCircle2Icon class="size-6" /></div>
						<h3 class="mt-4 text-lg font-bold text-foreground">Antrean penerbitan kosong</h3>
						<p class="mt-2 max-w-md text-sm leading-6 text-muted-foreground">Semua soal siap pakai sudah diterbitkan, atau belum ada soal yang siap diterbitkan.</p>
						<a href={resolve('/bank-soal/verifikasi')} class="mt-5 inline-flex h-9 items-center rounded-md border border-border bg-card px-3 text-sm font-semibold text-foreground hover:bg-muted/50">Buka Pemeriksaan</a>
					</div>
				{:else}
					<div class="hidden md:block">
						<Table.Root>
							<Table.Header>
								<Table.Row class="bg-muted/50">
									<Table.Head class="w-10"><input type="checkbox" aria-label="Pilih semua soal di halaman" checked={overview.questions.every((question) => selectedIds.has(question.id))} onchange={(event) => togglePageSelection((event.currentTarget as HTMLInputElement).checked)} /></Table.Head>
									<Table.Head class="w-[42%] text-muted-foreground">Soal</Table.Head>
									<Table.Head class="w-[18%] text-muted-foreground">Mapel</Table.Head>
									<Table.Head class="w-[20%] text-muted-foreground">Kesiapan</Table.Head>
									<Table.Head class="text-right text-muted-foreground">Aksi</Table.Head>
								</Table.Row>
							</Table.Header>
							<Table.Body>
								{#each overview.questions as question (question.id)}
									{@const checklist = publishChecklist(question)}
									<Table.Row>
										<Table.Cell><input type="checkbox" aria-label={`Pilih ${question.code || question.id}`} checked={selectedIds.has(question.id)} onchange={(event) => toggleSelection(question.id, (event.currentTarget as HTMLInputElement).checked)} /></Table.Cell>
										<Table.Cell>
											<div class="min-w-0">
												<div class="flex flex-wrap items-center gap-1.5">
													<Badge variant="outline" class="text-[10px]">{question.code || 'Tanpa kode'}</Badge>
													<Badge class="bg-success text-background text-[10px]">Siap Pakai</Badge>
													{#if question.hots_flag}<Badge variant="outline" class="text-[10px]">HOTS</Badge>{/if}
												</div>
												<p class="mt-1 line-clamp-2 text-sm font-medium text-foreground">{stemPreview(question)}</p>
												<p class="mt-1 text-xs text-muted-foreground">Penulis: {authorLabel(question)} · Pemeriksa: {reviewerLabel(question)} · Diperiksa {formatDate(question.reviewed_at)}</p>
											</div>
										</Table.Cell>
										<Table.Cell class="text-sm text-muted-foreground">
											<p class="font-medium text-foreground">{question.subject_name || question.subject_code || 'Mapel belum ada'}</p>
											<p>{questionTypeLabel(question.question_type)} · {question.target_level || 'Lintas tingkat'}</p>
											<p>{difficultyLabels[question.difficulty ?? ''] ?? question.difficulty ?? 'Sedang'} · Pakai {usageCount(question)} paket</p>
										</Table.Cell>
										<Table.Cell>
											<div class="flex flex-wrap gap-1.5">
												{#each checklist as item (item.label)}
													<span class="rounded px-1.5 py-0.5 text-[10px] font-semibold {item.ok ? 'bg-success/10 text-success' : 'bg-warning/10 text-warning'}">{item.ok ? '✓' : '!'} {item.label}</span>
												{/each}
											</div>
										</Table.Cell>
										<Table.Cell class="text-right">
											<div class="flex justify-end gap-1.5">
												<Button href={resolve(`/bank-soal/soal/${question.id}`)} variant="outline" size="icon-sm" aria-label="Lihat detail soal"><EyeIcon class="size-4" /></Button>
												<LoadingButton size="sm" loading={busyId === question.id} loadingLabel="Terbit..." disabled={!canPublish || (busyId !== '' && busyId !== question.id)} onclick={() => void publishOne(question)} class="bg-success text-background hover:bg-success">Terbitkan</LoadingButton>
											</div>
										</Table.Cell>
									</Table.Row>
								{/each}
							</Table.Body>
						</Table.Root>
					</div>

					<div class="divide-y divide-border md:hidden">
						{#each overview.questions as question (question.id)}
							{@const checklist = publishChecklist(question)}
							<article class="p-4">
								<div class="flex items-start gap-3">
									<input class="mt-1" type="checkbox" aria-label={`Pilih ${question.code || question.id}`} checked={selectedIds.has(question.id)} onchange={(event) => toggleSelection(question.id, (event.currentTarget as HTMLInputElement).checked)} />
									<div class="min-w-0 flex-1">
										<div class="flex flex-wrap gap-1.5"><Badge class="bg-success text-background text-[10px]">Siap Pakai</Badge><Badge variant="outline" class="text-[10px]">{question.code || 'Tanpa kode'}</Badge></div>
										<p class="mt-2 line-clamp-3 text-sm font-semibold text-foreground">{stemPreview(question)}</p>
										<p class="mt-1 text-xs text-muted-foreground">{question.subject_name || question.subject_code || 'Mapel'} · {questionTypeLabel(question.question_type)} · {question.target_level || 'Lintas tingkat'}</p>
										<p class="mt-1 text-xs text-muted-foreground">Penulis: {authorLabel(question)} · Reviewer: {reviewerLabel(question)}</p>
										<div class="mt-2 flex flex-wrap gap-1.5">
											{#each checklist as item (item.label)}<span class="rounded px-1.5 py-0.5 text-[10px] font-semibold {item.ok ? 'bg-success/10 text-success' : 'bg-warning/10 text-warning'}">{item.ok ? '✓' : '!'} {item.label}</span>{/each}
										</div>
										<div class="mt-3 flex flex-wrap gap-2">
											<a href={resolve(`/bank-soal/soal/${question.id}`)} class="inline-flex h-8 items-center rounded-md border border-border bg-card px-3 text-xs font-semibold text-foreground hover:bg-muted/50">Detail</a>
											<LoadingButton size="sm" loading={busyId === question.id} loadingLabel="Terbit..." disabled={!canPublish || (busyId !== '' && busyId !== question.id)} onclick={() => void publishOne(question)} class="h-8 bg-success text-xs text-background hover:bg-success">Terbitkan</LoadingButton>
										</div>
									</div>
								</div>
							</article>
						{/each}
					</div>
				{/if}
			</section>

			<TablePagination
				page={currentPage}
				limit={pageSize}
				total={overview.total}
				itemLabel="soal"
				ariaLabel="Navigasi halaman antrean penerbitan"
				onchange={handlePagination}
			/>
		{/snippet}
	</AsyncContent>

	<section class="rounded-xl border border-border bg-card p-4 text-sm text-muted-foreground shadow-sm">
		<p class="font-semibold text-foreground">Catatan alur</p>
		<p class="mt-1 leading-6">Penerbitan tidak menggantikan review. Soal harus tetap melewati verifikasi dulu. Setelah diterbitkan, soal masuk stok yang boleh dipakai oleh Paket Asesmen.</p>
		{#if !canPublish}
			<p class="mt-2 rounded-md border border-warning/30 bg-warning/10 px-3 py-2 text-warning">Akun ini bisa melihat antrean, tetapi tombol penerbitan membutuhkan izin <code>bank_soal.publish</code>.</p>
		{/if}
	</section>
</div>
