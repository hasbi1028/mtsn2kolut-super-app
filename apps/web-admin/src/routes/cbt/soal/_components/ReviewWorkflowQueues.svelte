<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import LoadingButton from '$lib/components/LoadingButton.svelte';

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
		grade_level?: number | null;
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
		reviewer_username?: string;
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
		onOpenReviewDecision: (question: Question, decision: 'approve' | 'reject') => void;
		onSubmitRevisionForReview: (question: Question) => void;
		onPublishQuestion: (question: Question) => void;
		difficultyLabel: Record<string, string>;
	} = $props();

	const queueClasses: Record<QueueTone, { section: string; title: string; badge: string; card: string; button: string; empty: string }> = {
		red: {
			section: 'border-red-200 bg-red-50/60',
			title: 'text-red-900',
			badge: 'text-red-700',
			card: 'border-red-100 hover:border-red-300 hover:bg-red-50',
			button: 'border-red-200 bg-white text-red-800 hover:bg-red-100',
			empty: 'border-red-100 text-red-800'
		},
		amber: {
			section: 'border-amber-200 bg-amber-50/70',
			title: 'text-amber-950',
			badge: 'text-amber-800',
			card: 'border-amber-100 hover:border-amber-300 hover:bg-amber-50',
			button: 'border-amber-200 bg-white text-amber-900 hover:bg-amber-100',
			empty: 'border-amber-100 text-amber-900'
		},
		green: {
			section: 'border-green-200 bg-green-50/70',
			title: 'text-green-950',
			badge: 'text-green-800',
			card: 'border-green-100 hover:border-green-300 hover:bg-green-50',
			button: 'border-green-200 bg-white text-green-900 hover:bg-green-100',
			empty: 'border-green-100 text-green-900'
		}
	};

	function queueMeta(action: QueueAction) {
		if (action === 'revision') {
			return {
				tone: 'red' as const,
				title: 'Antrian Revisi Soal',
				badge: `${revisionTotal} ${revisionSourceFilter ? 'sesuai filter' : 'perlu diperbaiki'}`,
				description: 'Buka, koreksi isi/kunci/rubrik, lalu ajukan review ulang.',
				buttonLabel: 'Lihat Semua Revisi',
				empty: 'Tidak ada revisi pada sumber ini.',
				items: revisionQueue,
				onShowAll: onShowAllRevisions
			};
		}
		if (action === 'review') {
			return {
				tone: 'amber' as const,
				title: 'Antrian Review Soal',
				badge: `${reviewTotal} menunggu keputusan`,
				description: 'Periksa soal yang diajukan guru, setujui, atau kembalikan dengan catatan revisi.',
				buttonLabel: 'Lihat Semua Menunggu Review',
				empty: 'Belum ada soal yang menunggu review pada filter ini.',
				items: reviewQueue,
				onShowAll: onShowPendingReviews
			};
		}
		return {
			tone: 'green' as const,
			title: 'Siap Terbit ke Paket',
			badge: `${approvedTotal} disetujui`,
			description: 'Soal sudah lolos review tetapi belum berstatus terbit.',
			buttonLabel: 'Lihat Semua Disetujui',
			empty: 'Belum ada soal disetujui yang menunggu terbit.',
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
					<span class={`rounded-full bg-white px-2 py-0.5 text-xs font-semibold ${classes.badge}`}>{meta.badge}</span>
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
							? 'border-red-300 bg-white font-semibold text-red-800 shadow-sm'
							: 'border-red-100 bg-red-50 text-red-700 hover:bg-white'}"
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
					<article class={`min-w-0 rounded-md border bg-white px-3 py-2 text-left shadow-sm transition-colors ${classes.card}`}>
						<div class="mb-1 flex items-center gap-1">
							<span class="rounded bg-white px-1.5 py-0.5 text-[10px] font-semibold text-green-800">{questionTypeLabel(q.question_type)}</span>
							<span class="truncate text-[11px] text-slate-400">{q.subject_name || q.subject_code || 'Mapel belum ada'}</span>
						</div>
						<button
							type="button"
							onclick={() => (action === 'review' && canReviewWorkflow ? onOpenReviewDecision(q, 'approve') : onOpenQuestion(q))}
							class="block w-full text-left"
						>
							<p class="line-clamp-2 text-sm font-medium text-slate-800 hover:text-green-800">{stemPreview(q)}</p>
						</button>
						<div class="mt-1 flex flex-wrap items-center gap-x-2 gap-y-1 text-[11px] text-slate-400">
							<span>{q.code || 'Tanpa kode'}</span>
							{#if q.author_username}<span>Guru: {q.author_username}</span>{/if}
							{#if q.reviewer_username}<span>Reviewer: {q.reviewer_username}</span>{/if}
							{#if action !== 'approved'}<span>{difficultyLabel[q.difficulty] ?? q.difficulty ?? 'Sedang'}</span>{/if}
						</div>
						{#if action === 'revision'}
							<div class="mt-2 rounded border border-red-100 bg-red-50/70 px-2 py-1.5">
								<div class="mb-0.5 flex flex-wrap items-center gap-1">
									<span class="rounded bg-white px-1.5 py-0.5 text-[10px] font-semibold uppercase tracking-wider text-red-700">{revisionSourceLabel(q)}</span>
									{#if q.reviewed_at}<span class="text-[10px] text-red-500">{new Date(q.reviewed_at).toLocaleDateString('id-ID')}</span>{/if}
								</div>
								<p class="line-clamp-2 text-[11px] leading-relaxed text-red-900">{revisionReason(q)}</p>
							</div>
						{:else if action === 'review' && q.review_notes}
							<div class="mt-2 rounded border border-amber-100 bg-amber-50/70 px-2 py-1.5 text-[11px] leading-relaxed text-amber-900">
								<span class="font-semibold">Catatan sebelumnya:</span> {revisionReason(q)}
							</div>
						{/if}
						<div class="mt-2 flex flex-wrap gap-1.5">
							{#if action === 'revision'}
								<Button variant="outline" size="sm" class="h-7 bg-white text-xs" onclick={() => onOpenQuestion(q)}>Edit Revisi</Button>
								<LoadingButton
									variant="outline"
									size="sm"
									class="h-7 border-green-200 bg-green-50 text-xs text-green-800 hover:bg-green-100"
									onclick={() => onSubmitRevisionForReview(q)}
									loading={workflowBusyId === q.id}
									loadingLabel="Mengajukan..."
									disabled={!canSubmitRevisionReview(q) || (workflowBusyId !== '' && workflowBusyId !== q.id)}
								>
									Ajukan Review Ulang
								</LoadingButton>
							{:else if action === 'review'}
								{#if canReviewWorkflow}
									<Button variant="outline" size="sm" class="h-7 border-green-200 bg-green-50 text-xs text-green-800 hover:bg-green-100" onclick={() => onOpenReviewDecision(q, 'approve')} disabled={!canDecideReview(q) || (workflowBusyId !== '' && workflowBusyId !== q.id)}>Setujui</Button>
									<Button variant="outline" size="sm" class="h-7 border-red-200 bg-red-50 text-xs text-red-700 hover:bg-red-100" onclick={() => onOpenReviewDecision(q, 'reject')} disabled={!canDecideReview(q) || (workflowBusyId !== '' && workflowBusyId !== q.id)}>Minta Revisi</Button>
								{:else}
									<Button variant="outline" size="sm" class="h-7 bg-white text-xs text-amber-800" onclick={() => onOpenQuestion(q)}>Periksa</Button>
								{/if}
							{:else}
								<LoadingButton
									variant="outline"
									size="sm"
									class="h-7 border-green-200 bg-green-700 text-xs text-white hover:bg-green-800"
									onclick={() => onPublishQuestion(q)}
									loading={workflowBusyId === q.id}
									loadingLabel="Menerbitkan..."
									disabled={!canPublishQuestion(q) || (workflowBusyId !== '' && workflowBusyId !== q.id)}
								>
									Terbitkan
								</LoadingButton>
								{#if !canReviewWorkflow}<span class="text-[11px] text-green-800">Menunggu admin menerbitkan.</span>{/if}
							{/if}
						</div>
						{#if action === 'review' && !canReviewWorkflow}
							<p class="mt-2 text-[11px] text-amber-800">Keputusan Setujui/Minta Revisi hanya tersedia untuk admin.</p>
						{/if}
					</article>
				{/each}
			</div>
		{:else}
			<div class={`mt-3 rounded-md border bg-white px-3 py-3 text-sm ${classes.empty}`}>
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
