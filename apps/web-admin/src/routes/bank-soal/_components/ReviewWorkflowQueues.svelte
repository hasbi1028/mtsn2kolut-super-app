<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import { displayName } from '$lib/utils/display-name';

	type RevisionSourceFilter = '' | 'item_analysis' | 'reviewer' | 'workflow';
	type OptionItem = {
		label: string;
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
		difficulty: string;
		workflow_status: string;
		status: string;
		options: OptionItem[];
		answer_key: string;
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
		usage?: { package_count?: number; answer_count?: number; is_locked?: boolean };
	};
	type RevisionSourceOption = { id: RevisionSourceFilter; label: string; desc: string };
	type QueueTone = 'red' | 'amber' | 'green';
	type QueueAction = 'revision' | 'review' | 'approved';

	let {
		revisionQueue,
		revisionTotal,
		reviewQueue,
		reviewTotal,
		approvedQueue,
		approvedTotal,
		revisionSourceOptions,
		revisionSourceFilter,
		canReviewWorkflow,
		workflowBusyId,
		questionTypeLabel,
		stemPreview,
		revisionSourceLabel,
		revisionReason,
		canSubmitRevisionReview,
		canDecideReview,
		canPublishQuestion,
		onRevisionSourceFilter,
		onShowAllRevisions,
		onShowPendingReviews,
		onShowApprovedQuestions,
		onOpenQuestion,
		onOpenReviewDecision,
		onSubmitRevisionForReview,
		onPublishQuestion,
		difficultyLabel
	}: {
		revisionQueue: Question[];
		revisionTotal: number;
		reviewQueue: Question[];
		reviewTotal: number;
		approvedQueue: Question[];
		approvedTotal: number;
		revisionSourceOptions: RevisionSourceOption[];
		revisionSourceFilter: RevisionSourceFilter;
		canReviewWorkflow: boolean;
		workflowBusyId: string;
		questionTypeLabel: (questionType: string) => string;
		stemPreview: (question: Question) => string;
		revisionSourceLabel: (question: Question) => string;
		revisionReason: (question: Question) => string;
		canSubmitRevisionReview: (question: Question) => boolean;
		canDecideReview: (question: Question) => boolean;
		canPublishQuestion: (question: Question) => boolean;
		onRevisionSourceFilter: (source: RevisionSourceFilter) => void;
		onShowAllRevisions: () => void;
		onShowPendingReviews: () => void;
		onShowApprovedQuestions: () => void;
		onOpenQuestion: (question: Question) => void;
		onOpenReviewDecision: (question: Question, decision: 'mark_reviewed' | 'request_revision' | 'reject') => void;
		onSubmitRevisionForReview: (question: Question) => void;
		onPublishQuestion: (question: Question) => void;
		difficultyLabel: Record<string, string>;
	} = $props();

	const queueClasses: Record<QueueTone, { section: string; title: string; badge: string; card: string; button: string; empty: string }> = {
		red: {
			section: 'border-destructive/30 bg-destructive/10',
			title: 'text-destructive',
			badge: 'text-destructive',
			card: 'border-destructive/30 hover:border-destructive/30 hover:bg-destructive/10',
			button: 'border-destructive/30 bg-card text-destructive hover:bg-destructive/15',
			empty: 'border-destructive/30 text-destructive'
		},
		amber: {
			section: 'border-warning/30 bg-warning/10',
			title: 'text-warning',
			badge: 'text-warning',
			card: 'border-warning/30 hover:border-warning/30 hover:bg-warning/10',
			button: 'border-warning/30 bg-card text-warning hover:bg-warning/15',
			empty: 'border-warning/30 text-warning'
		},
		green: {
			section: 'border-success/20 bg-success/10',
			title: 'text-success',
			badge: 'text-success',
			card: 'border-success/20 hover:border-success/20 hover:bg-success/10',
			button: 'border-success/20 bg-card text-success hover:bg-success/15',
			empty: 'border-success/20 text-success'
		}
	};

	function queueMeta(action: QueueAction) {
		if (action === 'revision') {
			return {
				tone: 'red' as const,
				title: 'Antrian Revisi Soal',
				badge: `${revisionTotal} ${revisionSourceFilter ? 'sesuai filter' : 'perlu diperbaiki'}`,
				description: 'Buka, koreksi isi/kunci/rubrik, lalu ajukan pemeriksaan ulang.',
				buttonLabel: 'Lihat Semua Revisi',
				empty: 'Tidak ada revisi pada sumber ini.',
				items: revisionQueue,
				onShowAll: onShowAllRevisions
			};
		}
		if (action === 'review') {
			return {
				tone: 'amber' as const,
				title: 'Antrian Pemeriksaan Soal',
				badge: `${reviewTotal} menunggu keputusan`,
				description: 'Periksa soal yang diajukan guru, setujui, atau kembalikan dengan catatan revisi.',
				buttonLabel: 'Lihat Semua Diperiksa',
				empty: 'Belum ada soal yang menunggu pemeriksaan pada filter ini.',
				items: reviewQueue,
				onShowAll: onShowPendingReviews
			};
		}
		return {
			tone: 'green' as const,
				title: 'Siap Pakai',
				badge: `${approvedTotal} siap dipakai`,
				description: 'Soal sudah layak dan bisa dipakai di paket asesmen.',
				buttonLabel: 'Lihat Semua Siap Pakai',
				empty: 'Belum ada soal siap pakai pada filter ini.',
			items: approvedQueue,
			onShowAll: onShowApprovedQuestions
		};
	}
</script>

{#snippet queuePanel(action: QueueAction)}
	{@const meta = queueMeta(action)}
	{@const classes = queueClasses[meta.tone]}
	<section class={`rounded-lg border p-3 ${classes.section}`}>
		<div class="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
			<div class="min-w-0">
				<div class="flex flex-wrap items-center gap-2">
					<h2 class={`text-sm font-bold uppercase tracking-wider ${classes.title}`}>{meta.title}</h2>
					<span class={`rounded-full bg-card px-2 py-0.5 text-xs font-semibold ${classes.badge}`}>{meta.badge}</span>
				</div>
				<p class={`mt-1 text-xs ${classes.title}`}>{meta.description}</p>
			</div>
			<Button variant="outline" size="sm" class={`shrink-0 ${classes.button}`} onclick={meta.onShowAll}>
				{meta.buttonLabel}
			</Button>
		</div>
		{#if action === 'revision'}
			<div class="mt-3 flex flex-wrap gap-1.5">
				{#each revisionSourceOptions as source (source.id)}
					<button
						type="button"
						onclick={() => onRevisionSourceFilter(source.id)}
						class="rounded-md border px-2.5 py-1.5 text-left text-xs transition-colors {revisionSourceFilter === source.id
							? 'border-destructive/30 bg-card font-semibold text-destructive shadow-sm'
							: 'border-destructive/30 bg-destructive/10 text-destructive hover:bg-card'}"
					>
						<span>{source.label}</span>
						<span class="ml-1 text-[10px] font-normal opacity-70">{source.desc}</span>
					</button>
				{/each}
			</div>
		{/if}
		{#if meta.items.length > 0}
			<div class="mt-3 grid gap-2 md:grid-cols-2 xl:grid-cols-3">
				{#each meta.items as q (q.id)}
					<article class={`min-w-0 rounded-md border bg-card px-3 py-2 text-left shadow-sm transition-colors ${classes.card}`}>
						<div class="mb-1 flex items-center gap-1">
							<span class="rounded bg-card px-1.5 py-0.5 text-[10px] font-semibold text-success">{questionTypeLabel(q.question_type)}</span>
							<span class="truncate text-[11px] text-muted-foreground">{q.subject_name || q.subject_code || 'Mapel belum ada'}</span>
						</div>
						<button
							type="button"
							onclick={() => (action === 'review' && canReviewWorkflow ? onOpenReviewDecision(q, 'mark_reviewed') : onOpenQuestion(q))}
							class="block w-full text-left"
						>
							<p class="line-clamp-2 text-sm font-medium text-foreground hover:text-success">{stemPreview(q)}</p>
						</button>
						<div class="mt-1 flex flex-wrap items-center gap-x-2 gap-y-1 text-[11px] text-muted-foreground">
							<span>{q.code || 'Tanpa kode'}</span>
							{#if q.author_username || q.author_display_name}<span>Guru: {displayName({ display_name: q.author_display_name, username: q.author_username }, 'Guru')}</span>{/if}
							{#if q.reviewer_username || q.reviewer_display_name}<span>Reviewer: {displayName({ display_name: q.reviewer_display_name, username: q.reviewer_username }, 'Reviewer')}</span>{/if}
							{#if action !== 'approved'}<span>{difficultyLabel[q.difficulty] ?? q.difficulty ?? 'Sedang'}</span>{/if}
						</div>
						{#if action === 'revision'}
							<div class="mt-2 rounded border border-destructive/30 bg-destructive/10 px-2 py-1.5">
								<div class="mb-0.5 flex flex-wrap items-center gap-1">
									<span class="rounded bg-card px-1.5 py-0.5 text-[10px] font-semibold uppercase tracking-wider text-destructive">{revisionSourceLabel(q)}</span>
									{#if q.reviewed_at}<span class="text-[10px] text-destructive">{new Date(q.reviewed_at).toLocaleDateString('id-ID')}</span>{/if}
								</div>
								<p class="line-clamp-2 text-[11px] leading-relaxed text-destructive">{revisionReason(q)}</p>
							</div>
						{:else if action === 'review' && q.review_notes}
							<div class="mt-2 rounded border border-warning/30 bg-warning/10 px-2 py-1.5 text-[11px] leading-relaxed text-warning">
								<span class="font-semibold">Catatan sebelumnya:</span> {revisionReason(q)}
							</div>
						{/if}
						<div class="mt-2 flex flex-wrap gap-1.5">
							{#if action === 'revision'}
								<Button variant="outline" size="sm" class="h-7 bg-card text-xs" onclick={() => onOpenQuestion(q)}>Edit Revisi</Button>
								<LoadingButton
									variant="outline"
									size="sm"
									class="h-7 border-success/20 bg-success/10 text-xs text-success hover:bg-success/15"
									onclick={() => onSubmitRevisionForReview(q)}
									loading={workflowBusyId === q.id}
									loadingLabel="Mengajukan..."
									disabled={!canSubmitRevisionReview(q) || (workflowBusyId !== '' && workflowBusyId !== q.id)}
								>
									Ajukan Ulang
								</LoadingButton>
							{:else if action === 'review'}
								{#if canReviewWorkflow}
									<Button variant="outline" size="sm" class="h-7 border-success/20 bg-success/10 text-xs text-success hover:bg-success/15" onclick={() => onOpenReviewDecision(q, 'mark_reviewed')} disabled={!canDecideReview(q) || (workflowBusyId !== '' && workflowBusyId !== q.id)}>Tandai Layak</Button>
									<Button variant="outline" size="sm" class="h-7 border-destructive/30 bg-destructive/10 text-xs text-destructive hover:bg-destructive/15" onclick={() => onOpenReviewDecision(q, 'request_revision')} disabled={!canDecideReview(q) || (workflowBusyId !== '' && workflowBusyId !== q.id)}>Minta Revisi</Button>
								{:else}
									<Button variant="outline" size="sm" class="h-7 bg-card text-xs text-warning" onclick={() => onOpenQuestion(q)}>Periksa</Button>
								{/if}
							{:else}
								<LoadingButton
									variant="outline"
									size="sm"
									class="h-7 border-success/20 bg-success text-xs text-background hover:bg-success"
									onclick={() => onPublishQuestion(q)}
									loading={workflowBusyId === q.id}
									loadingLabel="Menerbitkan..."
									disabled={!canPublishQuestion(q) || (workflowBusyId !== '' && workflowBusyId !== q.id)}
								>
									Terbitkan
								</LoadingButton>
								{#if !canReviewWorkflow}<span class="text-[11px] text-success">Menunggu admin menerbitkan.</span>{/if}
							{/if}
						</div>
						{#if action === 'review' && !canReviewWorkflow}
							<p class="mt-2 text-[11px] text-warning">Keputusan Setujui/Minta Revisi hanya tersedia untuk admin.</p>
						{/if}
					</article>
				{/each}
			</div>
		{:else}
			<div class={`mt-3 rounded-md border bg-card px-3 py-3 text-sm ${classes.empty}`}>
				{meta.empty}
			</div>
		{/if}
	</section>
{/snippet}

<div class="space-y-3">
	{@render queuePanel('revision')}
	{@render queuePanel('review')}
	{@render queuePanel('approved')}
</div>
