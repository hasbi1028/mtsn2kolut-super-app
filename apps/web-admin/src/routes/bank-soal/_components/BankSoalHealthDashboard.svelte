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
	import { ContextStrip, MetricCard, PageHeader, WorkflowCard } from '$lib/components/ops';
	import { clientApiPathWithQuery, readClientApiData } from '$lib/client/api';
	import { canCreateBankSoal, canManageBankSoalSettings, canReviewBankSoal } from '$lib/bank-soal/access';
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
	let canCreate = $derived(canCreateBankSoal(data.user));
	let canReview = $derived(canReviewBankSoal(data.user));
	let canSettings = $derived(canManageBankSoalSettings(data.user));
	let canUseQuality = $derived(canAccessQuality(data.user));

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
		if (href === '/bank-soal/daftar') return resolve('/bank-soal');
		if (href === '/bank-soal/verifikasi') return resolve('/bank-soal/verifikasi');
		if (href === '/bank-soal/tambah') return resolve('/bank-soal/tambah');
		if (href === '/bank-soal/impor') return resolve('/bank-soal/impor');
		if (href === '/bank-soal/analisis-butir') return resolve('/bank-soal/analisis-butir');
		if (href === '/bank-soal/pengaturan') return resolve('/bank-soal/pengaturan');
		if (href === '/bank-soal/penerbitan') return resolve('/bank-soal/penerbitan');
		return href;
	}

	function statusValue(model: BankSoalHealthModel, key: string): number | null {
		return model.statusCards.find((card) => card.key === key)?.value ?? null;
	}

	function totalQuestionValue(model: BankSoalHealthModel): number {
		return model.totalQuestions ?? model.sampleSize;
	}

	function canAccessQuality(user: PageData['user']): boolean {
		const roles = user?.roles ?? (user?.role ? [user.role] : []);
		const permissions = user?.permissions ?? [];
		return roles.includes('admin') || permissions.includes('bank_soal.analytics');
	}

	function handleRenderError(error: unknown, reset: () => void) {
		console.error('Ringkasan kesehatan Bank Soal belum dapat ditampilkan', error);
		reset();
	}

	onMount(() => {
		loadHealth();
	});
</script>

<svelte:head>
	<title>Alat Bank Soal - MTsN 2 Kolaka Utara</title>
</svelte:head>

<div class="space-y-5">
	<PageHeader
		eyebrow="Admin Bank Soal"
		title="Alat Bank Soal"
		subtitle="Pantau mutu, penerbitan, referensi, laporan, dan pengaturan Bank Soal dari satu tempat admin."
		context="Bank Soal MTsN 2 Kolaka Utara"
		primaryAction={canCreate ? { label: 'Tambah Soal', href: actionHref('/bank-soal/tambah') } : undefined}
		secondaryAction={{ label: 'Refresh', onclick: loadHealth }}
	/>

	<ContextStrip
		items={[
			{ label: 'Alur', value: 'Daftar → Verifikasi → Mutu → Pengaturan', tone: 'success' },
			{ label: 'Sumber data', value: 'Data Bank Soal', tone: 'muted' },
			{ label: 'Mode', value: 'Operasional' }
		]}
	/>

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
			<section class="grid gap-3 md:grid-cols-3" aria-label="Ringkasan utama Bank Soal">
				<MetricCard label="Soal tersedia" value={formatNumber(totalQuestionValue(model))} helper="Total dari ringkasan atau contoh data Bank Soal." tone="success" />
				<MetricCard label="Menunggu verifikasi" value={formatNumber(model.reviewBacklog.value)} helper={model.reviewBacklog.evidenceLabel} tone={model.reviewBacklog.value && model.reviewBacklog.value > 0 ? 'warning' : 'success'} />
				<MetricCard label="Mutu perlu dicek" value={model.warnings.length} helper="Catatan operasional dan kelengkapan data soal." tone={model.warnings.length > 0 ? 'warning' : 'success'} />
			</section>

			<section class="grid gap-3 md:grid-cols-2 xl:grid-cols-4" aria-label="Workflow Bank Soal">
				<WorkflowCard
					title="Kelola Soal"
					description="Cari, filter, edit, tambah, atau impor soal tanpa membuka semua fitur teknis."
					href={actionHref('/bank-soal/daftar')}
					actionLabel="Buka Kelola"
					status="Aktif"
				/>
				<WorkflowCard
					title="Verifikasi & Terbitkan"
					description="Antrean pemeriksa soal dan penerbitan dipusatkan sebagai satu alur keputusan."
					href={canReview ? actionHref('/bank-soal/verifikasi') : ''}
					actionLabel={canReview ? 'Buka Verifikasi' : 'Perlu izin pemeriksa soal'}
					status={`${formatNumber(model.reviewBacklog.value)} antrean`}
					tone={model.reviewBacklog.value && model.reviewBacklog.value > 0 ? 'warning' : 'default'}
				/>
				<WorkflowCard
					title="Mutu Soal"
					description="Cek cakupan mapel, kelengkapan data soal, HOTS, pemakaian paket, dan prioritas revisi."
					href={canUseQuality ? actionHref('/bank-soal/analisis-butir') : ''}
					actionLabel={canUseQuality ? 'Buka Mutu' : 'Perlu izin mutu'}
					status={`${formatNumber(statusValue(model, 'approved'))} siap`}
				/>
				<WorkflowCard
					title="Pengaturan"
					description="Standar kualitas, mapel/KD, impor, pemeriksa soal, dan SOP ada di area pengaturan."
					href={canSettings ? actionHref('/bank-soal/pengaturan') : ''}
					actionLabel={canSettings ? 'Buka Pengaturan' : 'Admin'}
					status="Lanjutan"
				/>
			</section>

			<details class="rounded-xl border border-border bg-card p-4 shadow-sm">
				<summary class="cursor-pointer text-sm font-semibold text-foreground">Rincian lengkap: mutu dan data operasional</summary>
				<div class="mt-4 space-y-5">
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
							<h2 class="text-base font-semibold text-foreground">Kesiapan Bank Soal</h2>
							<p class="mt-1 text-xs text-muted-foreground">{model.readiness.evidenceLabel}</p>
						</div>
						<ShieldCheckIcon class="size-5 text-primary" />
					</div>
					<div class="mt-5 grid gap-4 sm:grid-cols-[auto_1fr] sm:items-center">
						<div class="flex size-28 items-center justify-center rounded-full border border-primary/20 bg-primary/10">
							<div class="text-center">
								<p class="text-3xl font-semibold text-primary">{model.readiness.score ?? '-'}</p>
								<p class="text-xs font-medium text-muted-foreground">Nilai kesiapan {model.readiness.grade}</p>
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
							<h2 class="text-base font-semibold text-foreground">Antrean Verifikasi</h2>
							<p class="mt-1 text-xs text-muted-foreground">{model.reviewBacklog.evidenceLabel}</p>
						</div>
						<ClipboardCheckIcon class="size-5 text-warning" />
					</div>
					<p class="mt-5 text-3xl font-semibold text-foreground">{formatNumber(model.reviewBacklog.value)}</p>
					<p class="mt-1 text-sm text-muted-foreground">
						{#if model.reviewBacklog.percent === null}
							Perlu data pendukung total untuk membaca rasio antrean.
						{:else}
							{model.reviewBacklog.percent}% dari stok terukur masih menunggu verifikasi atau revisi.
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
										perlu data pendukung
									{:else}
										{metric.missing} tanpa data
									{/if}
								</p>
							</div>
						{/each}
					</div>
				</div>

				<div class="rounded-xl border border-border bg-card p-4 shadow-sm">
					<h2 class="text-base font-semibold text-foreground">Kelengkapan Data Soal</h2>
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
							<h2 class="text-base font-semibold text-foreground">Catatan Operasional</h2>
							<p class="mt-1 text-xs text-muted-foreground">Impor, aset, dan tata kelola hanya dinilai saat bukti pendukung tersedia.</p>
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
				</div>
			</details>
		{/snippet}
	</AsyncContent>
</div>
