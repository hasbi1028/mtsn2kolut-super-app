<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { clientApiPathWithQuery, readClientApiData } from '$lib/client/api';
	import { htmlToPlainText } from '$lib/utils/html-text';

	type Question = {
		id: string;
		code?: string;
		subject_name?: string;
		subject_code?: string;
		question_text?: string;
		stem_html?: string;
		question_type?: string;
		workflow_status?: string;
		status?: string;
		difficulty?: string;
		cognitive_level?: string;
		hots_flag?: boolean;
		package_count?: number;
		answer_count?: number;
		usage?: { package_count?: number; answer_count?: number; is_locked?: boolean };
		review_notes?: string;
		created_at?: string;
	};
	type QuestionListResponse = { items?: Question[]; meta?: { total?: number } };
	type SummaryResponse = {
		counts?: Partial<Record<'all' | 'total' | 'konsep' | 'diperiksa' | 'siap_pakai' | 'draft' | 'review' | 'approved' | 'published' | 'revision' | 'package_usage', number>>;
		by_subject?: Array<{ subject_name?: string; subject_code?: string; total?: number }>;
		by_cognitive_level?: Array<{ cognitive_level?: string; total?: number }>;
	};
	type AnalysisPayload = { questions: Question[]; summary: SummaryResponse };
	type Insight = { label: string; value: number; desc: string; tone: string };

	let promise = $state<Promise<AnalysisPayload> | null>(null);
	let questions = $state<Question[]>([]);
	let summary = $state<SummaryResponse>({});
	let activeSubject = $state('');

	const limit = 200;

	function plain(question: Question) {
		return htmlToPlainText(question.stem_html || question.question_text || '').trim() || '(Soal kosong)';
	}

	function packageCount(question: Question) {
		return question.usage?.package_count ?? question.package_count ?? 0;
	}

	function answerCount(question: Question) {
		return question.usage?.answer_count ?? question.answer_count ?? 0;
	}

	function normalizeWorkflowStatus(value?: string | null) {
		const status = (value ?? '').trim().toLowerCase();
		if (['konsep', 'draft', 'revision', 'revision_needed', 'rejected', 'archived'].includes(status)) return 'konsep';
		if (['diperiksa', 'review', 'submitted', 'reviewed'].includes(status)) return 'diperiksa';
		if (['siap_pakai', 'approved', 'published'].includes(status)) return 'siap_pakai';
		return status;
	}
	function workflowLabel(value?: string | null) {
		const status = normalizeWorkflowStatus(value);
		if (status === 'konsep') return 'Konsep';
		if (status === 'diperiksa') return 'Diperiksa';
		if (status === 'siap_pakai') return 'Siap Pakai';
		return 'Belum ada status';
	}

	function subjectName(question: Question) {
		return question.subject_name || question.subject_code || 'Tanpa Mapel';
	}

	async function fetchAnalysis(): Promise<AnalysisPayload> {
		const params = new URLSearchParams({ limit: String(limit), offset: '0' });
		const [questionPayload, summaryPayload] = await Promise.all([
			fetch(clientApiPathWithQuery('/api/bank-soal/questions', params)).then((response) => readClientApiData<QuestionListResponse>(response, 'Gagal memuat soal untuk analisis')),
			fetch('/api/bank-soal/summary').then((response) => readClientApiData<SummaryResponse>(response, 'Gagal memuat ringkasan Bank Soal'))
		]);
		return { questions: questionPayload.items ?? [], summary: summaryPayload ?? {} };
	}

	function load() {
		promise = fetchAnalysis().then((payload) => {
			questions = payload.questions;
			summary = payload.summary;
			return payload;
		});
	}

	let filteredQuestions = $derived(activeSubject ? questions.filter((question) => subjectName(question) === activeSubject) : questions);
	let totalQuestions = $derived(summary.counts?.total ?? summary.counts?.all ?? questions.length);
	let reviewedQuestions = $derived(summary.counts?.siap_pakai ?? summary.counts?.approved ?? summary.counts?.published ?? 0);
	let usedQuestions = $derived(questions.filter((question) => packageCount(question) > 0 || answerCount(question) > 0).length || (summary.counts?.package_usage ?? 0));
	let revisionQuestions = $derived(questions.filter((question) => normalizeWorkflowStatus(question.workflow_status) === 'konsep' && Boolean(question.review_notes?.trim())));
	let untouchedQuestions = $derived(questions.filter((question) => packageCount(question) === 0 && answerCount(question) === 0 && normalizeWorkflowStatus(question.workflow_status ?? question.status) === 'siap_pakai'));
	let hotsQuestions = $derived(questions.filter((question) => question.hots_flag || ['C4', 'C5', 'C6'].includes((question.cognitive_level ?? '').toUpperCase())));
	let subjectOptions = $derived(Array.from(new Set(questions.map(subjectName))).sort((a, b) => a.localeCompare(b, 'id')));
	let typeBuckets = $derived.by(() => {
		const map = new Map<string, number>();
		for (const question of filteredQuestions) map.set(question.question_type || 'lainnya', (map.get(question.question_type || 'lainnya') ?? 0) + 1);
		return Array.from(map.entries()).sort((a, b) => b[1] - a[1]);
	});
	let cognitiveBuckets = $derived((summary.by_cognitive_level ?? []).filter((item) => item.total && item.total > 0));
	let insights = $derived<Insight[]>([
		{ label: 'Total bank soal', value: totalQuestions, desc: `${filteredQuestions.length} soal masuk sampel mutu`, tone: 'emerald' },
		{ label: 'Siap pakai', value: reviewedQuestions, desc: 'Lolos pemeriksaan dan boleh dipakai', tone: 'green' },
		{ label: 'Dipakai paket/jawaban', value: usedQuestions, desc: 'Soal yang sudah punya jejak pemakaian', tone: 'amber' },
		{ label: 'Perlu revisi', value: revisionQuestions.length, desc: 'Prioritas perbaikan guru/reviewer', tone: 'rose' }
	]);
	let priorityItems = $derived([...revisionQuestions, ...untouchedQuestions].slice(0, 8));

	onMount(load);
</script>

<svelte:head><title>Mutu Soal - Bank Soal</title></svelte:head>

<div class="space-y-5 p-4 md:p-6">
	<section class="overflow-hidden rounded-2xl border border-primary/20 bg-card shadow-sm">
		<div class="bg-gradient-to-r from-primary/10 via-card to-warning/10 p-4 md:p-5">
			<div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
				<div>
					<p class="text-[10px] font-black uppercase tracking-[0.28em] text-primary">Mutu Bank Soal</p>
					<h1 class="mt-1 text-2xl font-black uppercase italic tracking-tight text-foreground">Mutu Soal</h1>
					<p class="mt-2 max-w-3xl text-sm leading-6 text-muted-foreground">Pantau sinyal mutu dari data yang sudah tersedia: status verifikasi, pemakaian paket, level kognitif, tipe soal, dan antrean prioritas revisi.</p>
				</div>
				<div class="flex flex-wrap gap-2">
					<a href={resolve('/bank-soal')} class="rounded-md border border-border bg-card px-3 py-2 text-sm font-semibold text-foreground hover:bg-muted/50">Dashboard</a>
					<a href={resolve('/bank-soal')} class="rounded-md border border-border bg-card px-3 py-2 text-sm font-semibold text-foreground hover:bg-muted/50">Daftar Soal</a>
				</div>
			</div>
		</div>
	</section>

	<AsyncContent {promise}>
		{#snippet pending()}
			<Skeleton class="h-96 w-full" />
		{/snippet}
		{#snippet failed(error, reset)}
			<RecoveryPanel title="Mutu Soal belum tersedia" message={error instanceof Error ? error.message : 'Gagal memuat mutu soal.'} onRetry={() => { reset?.(); load(); }} />
		{/snippet}
		{#snippet children()}
			<section class="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
				{#each insights as item (item.label)}
					<article class="rounded-xl border border-border bg-card p-4 shadow-sm">
						<p class="text-[10px] font-bold uppercase tracking-[0.18em] text-muted-foreground">{item.label}</p>
						<p class="mt-2 text-3xl font-black text-foreground">{item.value}</p>
						<p class="mt-1 text-xs text-muted-foreground">{item.desc}</p>
					</article>
				{/each}
			</section>

			<section class="grid gap-4 xl:grid-cols-[minmax(0,1fr)_22rem]">
				<div class="space-y-4">
					<div class="rounded-xl border border-border bg-card p-4 shadow-sm">
						<div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
							<div>
								<h2 class="text-base font-bold text-foreground">Distribusi Tipe Soal</h2>
								<p class="text-xs text-muted-foreground">Berdasarkan sampel {filteredQuestions.length} soal terbaru.</p>
							</div>
							<select bind:value={activeSubject} class="h-9 rounded-md border border-border bg-card px-3 text-sm text-foreground">
								<option value="">Semua mapel</option>
								{#each subjectOptions as subject (subject)}<option value={subject}>{subject}</option>{/each}
							</select>
						</div>
						<div class="mt-4 space-y-3">
							{#each typeBuckets as [type, total] (type)}
								<div>
									<div class="mb-1 flex justify-between text-xs"><span class="font-semibold uppercase text-muted-foreground">{type}</span><span>{total}</span></div>
									<div class="h-2 rounded-full bg-muted"><div class="h-2 rounded-full bg-primary" style={`width: ${Math.min(100, Math.round((total / Math.max(1, filteredQuestions.length)) * 100))}%`}></div></div>
								</div>
							{/each}
						</div>
					</div>

					<div class="rounded-xl border border-border bg-card p-4 shadow-sm">
						<h2 class="text-base font-bold text-foreground">Prioritas Tindak Lanjut</h2>
						<p class="mt-1 text-xs text-muted-foreground">Revisi dan soal lolos verifikasi yang belum terlihat pemakaiannya.</p>
						<div class="mt-4 divide-y divide-border">
							{#each priorityItems as question (question.id)}
								<a href={resolve(`/bank-soal/tambah?question_id=${question.id}`)} class="block py-3 hover:bg-muted/50">
									<div class="flex flex-wrap items-center gap-2 text-[10px] font-bold uppercase tracking-wide text-muted-foreground"><span>{subjectName(question)}</span><span>{workflowLabel(question.workflow_status ?? question.status)}</span><span>{question.code || 'tanpa kode'}</span></div>
									<p class="mt-1 line-clamp-2 text-sm font-semibold text-foreground">{plain(question)}</p>
								</a>
							{:else}
								<p class="rounded-lg border border-dashed border-border bg-muted/50 p-4 text-sm text-muted-foreground">Belum ada prioritas tindak lanjut dari sampel saat ini.</p>
							{/each}
						</div>
					</div>
				</div>

				<aside class="space-y-4">
					<div class="rounded-xl border border-border bg-card p-4 shadow-sm">
						<h2 class="text-base font-bold text-foreground">Komposisi Kognitif</h2>
						<div class="mt-4 space-y-3">
							{#each cognitiveBuckets as item (item.cognitive_level ?? 'unset')}
								<div class="rounded-lg border border-border bg-muted/50 p-3"><div class="flex justify-between text-sm"><span class="font-semibold">{item.cognitive_level || 'Belum diisi'}</span><span>{item.total}</span></div></div>
							{:else}
								<p class="text-sm text-muted-foreground">Level kognitif belum cukup tersedia.</p>
							{/each}
						</div>
					</div>
					<div class="rounded-xl border border-warning/30 bg-warning/10 p-4 text-warning">
						<p class="text-sm font-bold">HOTS / C4-C6</p>
						<p class="mt-2 text-3xl font-black">{hotsQuestions.length}</p>
						<p class="mt-1 text-xs text-warning">Indikasi dari flag HOTS atau level kognitif C4-C6 pada sampel.</p>
					</div>
				</aside>
			</section>
		{/snippet}
	</AsyncContent>
</div>
