<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import AlertTriangleIcon from '@lucide/svelte/icons/alert-triangle';
	import ArchiveIcon from '@lucide/svelte/icons/archive';
	import BarChart3Icon from '@lucide/svelte/icons/bar-chart-3';
	import CheckCircle2Icon from '@lucide/svelte/icons/check-circle-2';
	import ClipboardCheckIcon from '@lucide/svelte/icons/clipboard-check';
	import DatabaseIcon from '@lucide/svelte/icons/database';
	import FileQuestionIcon from '@lucide/svelte/icons/file-question';
	import ListChecksIcon from '@lucide/svelte/icons/list-checks';
	import PlusIcon from '@lucide/svelte/icons/plus';
	import RefreshCcwIcon from '@lucide/svelte/icons/refresh-ccw';
	import ShieldCheckIcon from '@lucide/svelte/icons/shield-check';
	import UploadIcon from '@lucide/svelte/icons/upload';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { clientApiPathWithQuery, readClientApiData } from '$lib/client/api';
	import {
		buildBankSoalHealthModel,
		filterBankSoalHealthActions,
		type BankSoalHealthModel,
		type BankSoalHealthQuestion,
		type BankSoalHealthSummary,
		type BankSoalRatioMetric,
		type BankSoalStatusCard,
		type BankSoalHealthWarning,
	} from '$lib/bank-soal/health';

	type PageData = {
		user?: {
			role?: string;
			roles?: string[];
			permissions?: string[];
		};
	};

	type QuestionListPayload = {
		items?: BankSoalHealthQuestion[];
		meta?: {
			total?: number;
			limit?: number;
			offset?: number;
		};
	};

	let { data }: { data: PageData } = $props();

	const sampleLimit = 50;
	let healthPromise = $state<Promise<BankSoalHealthModel> | null>(null);

	async function fetchHealth(): Promise<BankSoalHealthModel> {
		const params = new URLSearchParams({ limit: String(sampleLimit), offset: '0' });
		const [summary, questionPayload] = await Promise.all([
			fetch('/api/bank-soal/summary').then((response) =>
				readClientApiData<BankSoalHealthSummary>(response, 'Gagal memuat summary Bank Soal')
			),
			fetch(clientApiPathWithQuery('/api/bank-soal/questions', params)).then((response) =>
				readClientApiData<QuestionListPayload | BankSoalHealthQuestion[]>(response, 'Gagal memuat sampel soal')
			),
		]);

		const questions = Array.isArray(questionPayload) ? questionPayload : questionPayload.items ?? [];
		return buildBankSoalHealthModel({ summary, questions, sampleLimit });
	}

	function loadHealth() {
		healthPromise = fetchHealth();
	}

	function visibleActions(model: BankSoalHealthModel) {
		return filterBankSoalHealthActions(model.actions, data.user);
	}

	function formatNumber(value: number | null): string {
		return value === null ? 'Perlu evidence/data' : new Intl.NumberFormat('id-ID').format(value);
	}

	function formatPercent(value: number | null): string {
		return value === null ? 'Perlu evidence/data' : `${value}%`;
	}

	function statusCardClass(card: BankSoalStatusCard): string {
		if (card.tone === 'success') return 'border-primary/20 bg-primary/10';
		if (card.tone === 'warning') return 'border-warning/30 bg-warning/10';
		if (card.tone === 'danger') return 'border-destructive/30 bg-destructive/10';
		return 'border-border bg-card';
	}

	function statusValueClass(card: BankSoalStatusCard): string {
		if (card.value === null) return 'text-muted-foreground';
		if (card.tone === 'success') return 'text-primary';
		if (card.tone === 'warning') return 'text-warning';
		if (card.tone === 'danger') return 'text-destructive';
		return 'text-foreground';
	}

	function warningClass(warning: BankSoalHealthWarning): string {
		if (warning.severity === 'danger') return 'border-destructive/30 bg-destructive/10 text-destructive';
		if (warning.severity === 'warning') return 'border-warning/30 bg-warning/10 text-warning';
		return 'border-border bg-muted/50 text-muted-foreground';
	}

	function metricBarWidth(metric: BankSoalRatioMetric): number {
		return Math.max(metric.percent ?? 0, metric.value && metric.value > 0 ? 6 : 0);
	}

	function actionHref(href: string): string {
		if (href === '/bank-soal/daftar') return resolve('/bank-soal/daftar');
		if (href === '/bank-soal/verifikasi') return resolve('/bank-soal/verifikasi');
		if (href === '/bank-soal/tambah') return resolve('/bank-soal/tambah');
		if (href === '/bank-soal/impor') return resolve('/bank-soal/impor');
		return href;
	}

	function handleRenderError(error: unknown, reset: () => void) {
		console.error('Bank Soal health dashboard render failed', error);
		reset();
	}

	onMount(() => {
		loadHealth();
	});
</script>

<svelte:head>
	<title>Dashboard Kesehatan Bank Soal - MTsN 2 Kolaka Utara</title>
</svelte:head>

<div class="space-y-5">
	<section class="rounded-xl border border-primary/20 bg-card p-5 shadow-sm md:p-6">
		<div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
			<div class="max-w-3xl">
				<div class="mb-3 flex flex-wrap items-center gap-2">
					<span class="rounded-full border border-primary/20 bg-primary/10 px-3 py-1 text-xs font-semibold text-primary">Bank Soal</span>
					<span class="rounded-full border border-border bg-muted/60 px-3 py-1 text-xs text-muted-foreground">summary + sampel endpoint existing</span>
				</div>
				<h1 class="text-2xl font-semibold tracking-tight text-foreground md:text-3xl">Dashboard Kesehatan Bank Soal</h1>
				<p class="mt-2 max-w-2xl text-sm leading-6 text-muted-foreground">
					Pantau stok soal, readiness, cakupan kurikulum, mutu metadata, backlog review, dan warning operasional tanpa membuat analytics yang tidak punya evidence.
				</p>
			</div>
			<button
				type="button"
				class="inline-flex w-fit items-center gap-2 rounded-md border border-border bg-card px-3 py-2 text-sm font-medium text-foreground shadow-sm transition hover:border-primary/30 hover:bg-primary/10"
				onclick={loadHealth}
			>
				<RefreshCcwIcon class="size-4" />
				Refresh
			</button>
		</div>
	</section>

	<AsyncContent promise={healthPromise} onerror={handleRenderError}>
		{#snippet pending()}
			<div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
				{#each Array.from({ length: 8 }) as _, index (index)}
					<Skeleton class="h-28 rounded-lg" />
				{/each}
			</div>
			<div class="grid gap-4 xl:grid-cols-[1.2fr_0.8fr]">
				<Skeleton class="h-72 rounded-xl" />
				<Skeleton class="h-72 rounded-xl" />
			</div>
		{/snippet}

		{#snippet failed(error)}
			<RecoveryPanel compact message={error instanceof Error ? error.message : String(error)} onRetry={loadHealth} />
		{/snippet}

		{#snippet children(value)}
			{@const model = value as BankSoalHealthModel}
			<section class="grid gap-3 sm:grid-cols-2 xl:grid-cols-7" aria-label="Indikator status Bank Soal">
				{#each model.statusCards as card (card.key)}
					<div class={`rounded-lg border p-4 shadow-sm ${statusCardClass(card)}`}>
						<div class="flex items-center justify-between gap-3">
							<span class="text-xs font-semibold uppercase text-muted-foreground">{card.label}</span>
							{#if card.key === 'archived'}
								<ArchiveIcon class="size-4 text-muted-foreground" />
							{:else if card.key === 'published' || card.key === 'approved'}
								<CheckCircle2Icon class="size-4 text-primary" />
							{:else}
								<FileQuestionIcon class="size-4 text-muted-foreground" />
							{/if}
						</div>
						<p class={`mt-2 text-2xl font-semibold ${statusValueClass(card)}`}>{formatNumber(card.value)}</p>
						<p class="mt-1 text-xs text-muted-foreground">{card.evidenceLabel}</p>
					</div>
				{/each}
			</section>

			<section class="grid gap-4 xl:grid-cols-[1.1fr_0.9fr]">
				<div class="rounded-xl border border-border bg-card p-4 shadow-sm">
					<div class="flex items-start justify-between gap-3">
						<div>
							<h2 class="text-base font-semibold text-foreground">Readiness Bank Soal</h2>
							<p class="mt-1 text-xs text-muted-foreground">{model.readiness.evidenceLabel}</p>
						</div>
						<ShieldCheckIcon class="size-5 text-primary" />
					</div>
					<div class="mt-5 grid gap-4 sm:grid-cols-[auto_1fr] sm:items-center">
						<div class="flex size-28 items-center justify-center rounded-full border border-primary/20 bg-primary/10">
							<div class="text-center">
								<p class="text-3xl font-semibold text-primary">{model.readiness.score ?? '-'}</p>
								<p class="text-xs font-medium text-muted-foreground">Grade {model.readiness.grade}</p>
							</div>
						</div>
						<div class="space-y-2">
							{#each model.readiness.drivers as driver}
								<div class="flex items-center gap-2 rounded-md border border-border bg-muted/50 px-3 py-2 text-sm text-foreground">
									<DatabaseIcon class="size-4 shrink-0 text-muted-foreground" />
									<span>{driver}</span>
								</div>
							{/each}
						</div>
					</div>
				</div>

				<div class="rounded-xl border border-border bg-card p-4 shadow-sm">
					<div class="flex items-start justify-between gap-3">
						<div>
							<h2 class="text-base font-semibold text-foreground">Review Backlog</h2>
							<p class="mt-1 text-xs text-muted-foreground">{model.reviewBacklog.evidenceLabel}</p>
						</div>
						<ClipboardCheckIcon class="size-5 text-warning" />
					</div>
					<p class="mt-5 text-3xl font-semibold text-foreground">{formatNumber(model.reviewBacklog.value)}</p>
					<p class="mt-1 text-sm text-muted-foreground">
						{#if model.reviewBacklog.percent === null}
							Perlu evidence/data total untuk membaca rasio backlog.
						{:else}
							{model.reviewBacklog.percent}% dari stok terukur masih review atau revisi.
						{/if}
					</p>
				</div>
			</section>

			<section class="grid gap-4 xl:grid-cols-3">
				<div class="rounded-xl border border-border bg-card p-4 shadow-sm">
					<div class="flex items-start justify-between gap-3">
						<div>
							<h2 class="text-base font-semibold text-foreground">Cakupan Mapel</h2>
							<p class="mt-1 text-xs text-muted-foreground">{model.subjectCoverage.evidenceLabel}</p>
						</div>
						<BarChart3Icon class="size-5 text-primary" />
					</div>
					<p class="mt-5 text-3xl font-semibold text-foreground">
						{#if model.subjectCoverage.value === null}
							-
						{:else if model.subjectCoverage.total === null}
							{model.subjectCoverage.value}
						{:else}
							{model.subjectCoverage.value}/{model.subjectCoverage.total}
						{/if}
					</p>
					<div class="mt-4 h-2 rounded-full bg-muted">
						<div class="h-2 rounded-full bg-primary" style={`width: ${metricBarWidth(model.subjectCoverage)}%`}></div>
					</div>
				</div>

				<div class="rounded-xl border border-border bg-card p-4 shadow-sm">
					<h2 class="text-base font-semibold text-foreground">Cakupan KD/CP/TP</h2>
					<div class="mt-4 space-y-3">
						{#each model.curriculumCoverage as metric (metric.key)}
							<div>
								<div class="mb-1 flex items-center justify-between gap-3 text-xs">
									<span class="font-medium text-foreground">{metric.label}</span>
									<span class="text-muted-foreground">{formatPercent(metric.percent)}</span>
								</div>
								<div class="h-2 rounded-full bg-muted">
									<div class="h-2 rounded-full bg-primary" style={`width: ${metricBarWidth(metric)}%`}></div>
								</div>
								<p class="mt-1 text-xs text-muted-foreground">
									{#if metric.missing === null}
										perlu evidence/data
									{:else}
										{metric.missing} tanpa data
									{/if}
								</p>
							</div>
						{/each}
					</div>
				</div>

				<div class="rounded-xl border border-border bg-card p-4 shadow-sm">
					<h2 class="text-base font-semibold text-foreground">Kualitas Metadata</h2>
					<div class="mt-4 space-y-3">
						{#each model.metadataQuality as metric (metric.key)}
							<div class="rounded-md border border-border bg-muted/40 px-3 py-2">
								<div class="flex items-center justify-between gap-3">
									<span class="text-sm font-medium text-foreground">{metric.label}</span>
									<span class="font-mono text-sm text-foreground">{metric.missing ?? '-'}</span>
								</div>
								<p class="mt-1 text-xs text-muted-foreground">{metric.evidenceLabel}</p>
							</div>
						{/each}
					</div>
				</div>
			</section>

			<section class="grid gap-4 xl:grid-cols-[1fr_0.9fr]">
				<div class="rounded-xl border border-border bg-card p-4 shadow-sm">
					<div class="flex items-start justify-between gap-3">
						<div>
							<h2 class="text-base font-semibold text-foreground">Warning Operasional</h2>
							<p class="mt-1 text-xs text-muted-foreground">Import, aset, dan tata kelola hanya dinilai saat evidence tersedia.</p>
						</div>
						<AlertTriangleIcon class="size-5 text-warning" />
					</div>
					<div class="mt-4 space-y-2">
						{#each model.warnings as warning}
							<div class={`rounded-lg border p-3 ${warningClass(warning)}`}>
								<div class="flex items-center justify-between gap-3">
									<p class="text-sm font-semibold">{warning.label}</p>
									<span class="rounded-full border border-current/20 px-2 py-0.5 text-[11px]">{warning.evidenceLabel}</span>
								</div>
								<p class="mt-1 text-sm leading-5">{warning.message}</p>
							</div>
						{/each}
					</div>
				</div>

				<div class="rounded-xl border border-border bg-card p-4 shadow-sm">
					<h2 class="text-base font-semibold text-foreground">Rekomendasi Tindakan Cepat</h2>
					<div class="mt-4 grid gap-2">
						{#each visibleActions(model) as action (action.label)}
							<a
								href={actionHref(action.href)}
								class="group flex items-center gap-3 rounded-lg border border-border bg-muted/50 p-3 transition hover:border-primary/20 hover:bg-primary/10"
							>
								<span class="flex size-10 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary">
									{#if action.capability === 'review'}
										<ClipboardCheckIcon class="size-5" />
									{:else if action.capability === 'create'}
										<PlusIcon class="size-5" />
									{:else if action.capability === 'import'}
										<UploadIcon class="size-5" />
									{:else}
										<ListChecksIcon class="size-5" />
									{/if}
								</span>
								<span class="min-w-0 flex-1">
									<span class="block text-sm font-semibold text-foreground">{action.label}</span>
									<span class="block text-xs leading-5 text-muted-foreground">{action.description}</span>
								</span>
							</a>
						{/each}
					</div>
				</div>
			</section>
		{/snippet}
	</AsyncContent>
</div>
