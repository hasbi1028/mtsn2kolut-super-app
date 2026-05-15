<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Textarea } from '$lib/components/ui/textarea';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import RichContent from '$lib/components/RichContent.svelte';
	import ComposerVersionPanel from './ComposerVersionPanel.svelte';
	import {
		DIFFICULTY_LABEL,
		WORKFLOW_LABEL,
		answerKeyLabelForQuestion,
		optionPrimaryContent,
		optionSecondaryContent,
		questionTypeLabel,
		questionUsageText,
		revisionReason,
		workflowClass,
		type Question,
		type QuestionVersion,
		type TimelineItem,
	} from './soal-workspace.model';
	import {
		composerDateTimeLabel,
		questionVersionLabel,
	} from './soal-workspace.navigation';

	type ReviewDecision = 'approve' | 'reject';

	let {
		question,
		backHref,
		versions,
		versionsLoading,
		timeline,
		timelineLoading,
		workflowBusyId,
		canEdit,
		canReturnRevision,
		canCreateRevision,
		canReview,
		canPublish,
		lockMessage,
		hrefForVersion,
		onEdit,
		onReturnRevision,
		onCreateRevision,
		onReviewDecision,
		onPublish,
	}: {
		question: Question;
		backHref: string;
		versions: QuestionVersion[];
		versionsLoading: boolean;
		timeline: TimelineItem[];
		timelineLoading: boolean;
		workflowBusyId: string;
		canEdit: boolean;
		canReturnRevision: boolean;
		canCreateRevision: boolean;
		canReview: boolean;
		canPublish: boolean;
		lockMessage: string;
		hrefForVersion: (id: string) => string;
		onEdit: () => void;
		onReturnRevision: () => void;
		onCreateRevision: () => void;
		onReviewDecision: (question: Question, decision: ReviewDecision, notes?: string) => void;
		onPublish: (question: Question) => void;
	} = $props();

	let reviewerNotes = $state('');
	let reviewerChecklist = $state({
		materi: false,
		opsi: false,
		kunci: false,
		level: false,
		bahasa: false,
	});

	let options = $derived(question.options ?? []);
	let reviewerChecklistComplete = $derived(Object.values(reviewerChecklist).every(Boolean));
	let authorName = $derived((question.author_display_name || '').trim());
	let authorUsername = $derived((question.author_username || '').trim());
	let reviewerName = $derived((question.reviewer_display_name || '').trim());
	let reviewerUsername = $derived((question.reviewer_username || '').trim());
	let approverName = $derived((question.approver_display_name || '').trim());
	let approverUsername = $derived((question.approver_username || '').trim());

	const reviewItems = [
		{ key: 'materi', label: 'Materi benar' },
		{ key: 'opsi', label: 'Opsi tidak ambigu' },
		{ key: 'kunci', label: 'Kunci/rubrik benar' },
		{ key: 'level', label: 'Level kognitif sesuai' },
		{ key: 'bahasa', label: 'Bahasa jelas' },
	] as const;
</script>

<article class="space-y-4 rounded-2xl border border-border bg-card p-4 shadow-sm md:p-5">
	<nav class="flex flex-wrap items-center gap-2 text-xs text-muted-foreground" aria-label="Breadcrumb detail soal">
		<a href={backHref} class="font-semibold text-primary underline-offset-2 hover:underline">Bank Soal</a>
		<span aria-hidden="true">/</span>
		<span>Detail Soal</span>
		<span aria-hidden="true">/</span>
		<span class="font-semibold text-foreground">{questionVersionLabel(question)}</span>
	</nav>

	<header class="flex flex-col gap-3 border-b border-border pb-4 lg:flex-row lg:items-start lg:justify-between">
		<div class="min-w-0">
			<p class="text-xs font-black uppercase tracking-[0.2em] text-primary">Detail Soal</p>
			<h2 class="mt-1 text-xl font-black text-foreground">
				{question.code || `Butir Soal ${questionVersionLabel(question)}`}
			</h2>
			<p class="mt-2 max-w-3xl text-sm leading-6 text-muted-foreground">{lockMessage}</p>
			<p class="mt-2 rounded-lg border border-primary/20 bg-primary/10 px-3 py-2 text-sm font-medium text-primary">
				Mode lihat. Draft lokal tidak digunakan di halaman detail ini.
			</p>
		</div>
		<div class="flex flex-wrap gap-2 lg:justify-end">
			<a href={backHref} class="inline-flex h-9 items-center rounded-md border border-border bg-card px-3 text-sm font-semibold text-foreground hover:bg-muted/50">
				Kembali ke Daftar
			</a>
			{#if canEdit}
				<Button type="button" class="h-9 bg-success text-background hover:bg-success" onclick={onEdit} aria-label="Edit soal ini">
					Edit
				</Button>
			{/if}
			{#if canReturnRevision}
				<Button type="button" class="h-9" variant="outline" onclick={onReturnRevision} aria-label="Kembalikan soal ke revisi">
					Kembalikan ke Revisi
				</Button>
			{/if}
			{#if canCreateRevision}
				<Button type="button" class="h-9" onclick={onCreateRevision} aria-label="Buat draft revisi baru">
					Buat Revisi Baru
				</Button>
			{/if}
		</div>
	</header>

	<section class="grid gap-2 sm:grid-cols-2 xl:grid-cols-4" aria-label="Metadata ringkas soal">
		<div class="rounded-lg border border-border bg-muted/50 px-3 py-2">
			<p class="text-xs font-semibold text-muted-foreground">Mata pelajaran</p>
			<p class="mt-0.5 text-sm font-bold text-foreground">{question.subject_name || question.subject_code || '-'}</p>
		</div>
		<div class="rounded-lg border border-border bg-muted/50 px-3 py-2">
			<p class="text-xs font-semibold text-muted-foreground">Tipe dan tingkat</p>
			<p class="mt-0.5 text-sm font-bold text-foreground">{questionTypeLabel(question.question_type)} · Kelas {question.grade_level ?? '-'}</p>
		</div>
		<div class="rounded-lg border border-border bg-muted/50 px-3 py-2">
			<p class="text-xs font-semibold text-muted-foreground">Status</p>
			<p class="mt-1">
				<span class="rounded px-1.5 py-0.5 text-xs font-semibold {workflowClass(question.workflow_status)}">
					{WORKFLOW_LABEL[question.workflow_status] ?? question.workflow_status ?? '-'}
				</span>
				<span class="ml-1 text-xs text-muted-foreground">{question.status || 'draft'}</span>
			</p>
		</div>
		<div class="rounded-lg border border-border bg-muted/50 px-3 py-2">
			<p class="text-xs font-semibold text-muted-foreground">Versi dan pemakaian</p>
			<p class="mt-0.5 text-sm font-bold text-foreground">{questionVersionLabel(question)} · {questionUsageText(question)}</p>
		</div>
		<div class="rounded-lg border border-border bg-muted/50 px-3 py-2">
			<p class="text-xs font-semibold text-muted-foreground">Kesulitan</p>
			<p class="mt-0.5 text-sm font-bold text-foreground">{DIFFICULTY_LABEL[question.difficulty] ?? question.difficulty ?? '-'}</p>
		</div>
		<div class="rounded-lg border border-border bg-muted/50 px-3 py-2">
			<p class="text-xs font-semibold text-muted-foreground">Topik / Kognitif</p>
			<p class="mt-0.5 text-sm font-bold text-foreground">{question.material_topic || '-'} · {question.cognitive_level || '-'}</p>
		</div>
		<div class="rounded-lg border border-border bg-muted/50 px-3 py-2">
			<p class="text-xs font-semibold text-muted-foreground">Penulis</p>
			<p class="mt-0.5 text-sm font-bold text-foreground">{authorName || authorUsername || '-'}</p>
			{#if authorName && authorUsername && authorName !== authorUsername}
				<p class="mt-0.5 text-[11px] font-semibold text-muted-foreground">{authorUsername}</p>
			{/if}
		</div>
		<div class="rounded-lg border border-border bg-muted/50 px-3 py-2">
			<p class="text-xs font-semibold text-muted-foreground">Diperbarui</p>
			<p class="mt-0.5 text-sm font-bold text-foreground">{composerDateTimeLabel(question.reviewed_at ?? question.created_at)}</p>
		</div>
	</section>

	<div class="grid gap-4 xl:grid-cols-[minmax(0,1.45fr)_minmax(19rem,0.55fr)]">
		<section class="space-y-4 rounded-xl border border-border bg-muted/40 p-4" aria-label="Isi soal">
			{#if question.stimulus_html}
				<div class="rounded-lg border border-border bg-card p-3">
					<p class="mb-2 text-xs font-black uppercase tracking-[0.18em] text-muted-foreground">Stimulus</p>
					<RichContent html={question.stimulus_html} class="prose prose-sm max-w-none text-foreground latex-preview" />
				</div>
			{/if}

			<div class="rounded-lg border border-border bg-card p-3">
				<p class="mb-2 text-xs font-black uppercase tracking-[0.18em] text-muted-foreground">Pertanyaan</p>
				<RichContent html={question.stem_html || question.question_text || 'Isi soal belum tersedia'} class="prose prose-sm max-w-none text-foreground latex-preview" />
			</div>

			{#if options.length > 0}
				<div class="rounded-lg border border-border bg-card p-3">
					<p class="mb-2 text-xs font-black uppercase tracking-[0.18em] text-muted-foreground">Opsi / Pasangan</p>
					<div class="space-y-2">
						{#each options as option, index (`detail-option-${question.id}-${index}`)}
							{@const secondary = optionSecondaryContent(option)}
							<div class="rounded-lg border border-border bg-muted/50 px-3 py-2 text-sm text-foreground">
								<div class="flex gap-2">
									<span class="shrink-0 font-black text-success">{option.label || option.match_label || index + 1}.</span>
									<div class="min-w-0 flex-1">
										<RichContent html={optionPrimaryContent(option)} class="latex-preview" />
										{#if secondary}
											<div class="mt-2 border-t border-border pt-2 text-muted-foreground">
												<RichContent html={secondary} class="latex-preview" />
											</div>
										{/if}
									</div>
								</div>
							</div>
						{/each}
					</div>
				</div>
			{/if}

			<div class="rounded-lg border border-success/20 bg-success/10 p-3 text-sm text-success">
				<p class="font-semibold">Kunci / Rubrik</p>
				<p class="mt-1">{answerKeyLabelForQuestion(question)}</p>
				{#if question.rubric_html}
					<div class="mt-3 rounded-lg border border-success/20 bg-card p-3">
						<RichContent html={question.rubric_html} class="prose prose-sm max-w-none text-success latex-preview" />
					</div>
				{/if}
			</div>

			{#if question.explanation_html}
				<div class="rounded-lg border border-border bg-card p-3">
					<p class="mb-2 text-xs font-black uppercase tracking-[0.18em] text-muted-foreground">Pembahasan / Catatan Internal</p>
					<RichContent html={question.explanation_html} class="prose prose-sm max-w-none text-foreground latex-preview" />
				</div>
			{/if}

			{#if question.review_notes}
				<div class="rounded-lg border border-warning/30 bg-warning/10 p-3 text-sm text-warning">
					<p class="font-semibold">Catatan reviewer terakhir</p>
					<p class="mt-1">{revisionReason(question)}</p>
				</div>
			{/if}
		</section>

		<aside class="space-y-4">
			{#if canReview}
				<section class="rounded-xl border border-warning/30 bg-warning/10 p-4 text-warning shadow-sm" aria-label="Panel reviewer">
					<div class="flex items-start justify-between gap-3">
						<div>
							<p class="text-xs font-black uppercase tracking-[0.18em]">Telaah Reviewer</p>
							<h3 class="mt-1 text-sm font-bold">Checklist sebelum keputusan</h3>
						</div>
						<span class="rounded-full bg-card px-2 py-1 text-xs font-semibold">
							{reviewerChecklistComplete ? 'Lengkap' : 'Cek'}
						</span>
					</div>
					<div class="mt-3 space-y-2">
						{#each reviewItems as item (item.key)}
							<label class="flex cursor-pointer items-center gap-2 rounded-lg border border-warning/30 bg-card px-3 py-2 text-sm">
								<input
									type="checkbox"
									bind:checked={reviewerChecklist[item.key]}
									class="h-4 w-4 rounded accent-green-700"
								/>
								<span>{item.label}</span>
							</label>
						{/each}
					</div>
					<div class="mt-3">
						<label for="detail-reviewer-notes" class="mb-1 block text-xs font-semibold">Catatan reviewer</label>
						<Textarea
							id="detail-reviewer-notes"
							rows={3}
							bind:value={reviewerNotes}
							placeholder="Catatan ini akan dibawa ke dialog keputusan."
							class="bg-card"
						/>
					</div>
					<div class="mt-3 grid grid-cols-2 gap-2">
						<LoadingButton
							variant="outline"
							size="sm"
							class="h-9 bg-card text-success"
							loading={workflowBusyId === question.id}
							loadingLabel="Memproses..."
							disabled={!reviewerChecklistComplete || (workflowBusyId !== '' && workflowBusyId !== question.id)}
							onclick={() => onReviewDecision(question, 'approve', reviewerNotes)}
						>
							Setujui
						</LoadingButton>
						<LoadingButton
							variant="outline"
							size="sm"
							class="h-9 bg-card text-destructive"
							loading={workflowBusyId === question.id}
							loadingLabel="Memproses..."
							disabled={workflowBusyId !== '' && workflowBusyId !== question.id}
							onclick={() => onReviewDecision(question, 'reject', reviewerNotes)}
						>
							Minta Revisi
						</LoadingButton>
					</div>
				</section>
			{/if}

			{#if canPublish}
				<section class="rounded-xl border border-success/20 bg-success/10 p-4 text-success shadow-sm">
					<p class="text-xs font-black uppercase tracking-[0.18em]">Publikasi</p>
					<p class="mt-1 text-sm">Soal sudah disetujui dan bisa diterbitkan ke paket ujian.</p>
					<LoadingButton
						size="sm"
						class="mt-3 h-9 bg-success text-background hover:bg-success"
						loading={workflowBusyId === question.id}
						loadingLabel="Terbit..."
						disabled={workflowBusyId !== '' && workflowBusyId !== question.id}
						onclick={() => onPublish(question)}
					>
						Terbitkan
					</LoadingButton>
				</section>
			{/if}

			<section class="rounded-xl border border-border bg-card p-4 shadow-sm" aria-label="Timeline soal">
				<div class="flex items-center justify-between gap-2">
					<p class="text-xs font-black uppercase tracking-[0.18em] text-muted-foreground">Timeline</p>
					{#if timelineLoading}
						<span class="text-xs text-muted-foreground">Memuat...</span>
					{/if}
				</div>
				{#if timeline.length > 0}
					<div class="mt-3 space-y-2">
						{#each timeline.slice(0, 8) as item, index (`detail-timeline-${item.id ?? index}`)}
							<div class="rounded-lg border border-border bg-muted/50 px-3 py-2 text-xs">
								<div class="flex flex-wrap items-center gap-1.5">
									<span class="font-semibold text-success">{item.action ?? item.status ?? 'Perubahan'}</span>
									{#if item.actor_username}<span class="text-muted-foreground">oleh {item.actor_username}</span>{/if}
									{#if item.created_at}<span class="text-muted-foreground">{composerDateTimeLabel(item.created_at)}</span>{/if}
								</div>
								{#if item.notes}<p class="mt-1 text-muted-foreground">{item.notes}</p>{/if}
							</div>
						{/each}
					</div>
				{:else}
					<p class="mt-3 text-sm text-muted-foreground">Timeline belum tersedia dari backend.</p>
				{/if}
			</section>
		</aside>
	</div>

	<ComposerVersionPanel
		currentQuestion={question}
		versions={versions}
		loading={versionsLoading}
		{hrefForVersion}
	/>
</article>
