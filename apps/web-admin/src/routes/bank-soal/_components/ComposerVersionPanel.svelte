<script lang="ts">
	import { WORKFLOW_LABEL, workflowClass, type Question, type QuestionVersion } from './soal-workspace.model';
	import {
		compactVersionNote,
		composerDateTimeLabel,
		questionVersionLabel,
	} from './soal-workspace.navigation';

	let {
		currentQuestion,
		versions,
		loading,
		hrefForVersion,
	}: {
		currentQuestion: Question | null;
		versions: QuestionVersion[];
		loading: boolean;
		hrefForVersion: (id: string) => string;
	} = $props();

	let latestVersion = $derived(versions.find((version) => version.is_latest_version) ?? versions[0] ?? null);

	function isCurrent(version: QuestionVersion): boolean {
		return version.id === currentQuestion?.id;
	}

	function actorLabel(version: QuestionVersion): string {
		const reviewer = (version.reviewer_display_name || version.reviewer_username || '').trim();
		if (reviewer) return `Reviewer: ${reviewer}`;
		const approver = (version.approver_display_name || version.approver_username || '').trim();
		if (approver) return `Approver: ${approver}`;
		const author = (version.author_display_name || version.author_username || '').trim();
		return author ? `Penulis: ${author}` : 'Aktor belum tercatat';
	}
</script>

<section class="rounded-xl border border-border bg-card p-4 shadow-sm" aria-label="Riwayat versi soal">
	<div class="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
		<div>
			<p class="text-xs font-black uppercase tracking-[0.2em] text-muted-foreground">Riwayat Versi</p>
			<h3 class="mt-1 text-sm font-bold text-foreground">
				{currentQuestion ? `${questionVersionLabel(currentQuestion)} aktif` : 'Versi aktif belum tersedia'}
			</h3>
		</div>
		<div class="flex flex-wrap gap-2 text-xs">
			<span class="rounded-full border border-primary/20 bg-primary/10 px-2.5 py-1 font-semibold text-primary">
				Current {currentQuestion ? questionVersionLabel(currentQuestion) : '-'}
			</span>
			<span class="rounded-full border border-success/20 bg-success/10 px-2.5 py-1 font-semibold text-success">
				Latest {latestVersion ? questionVersionLabel(latestVersion) : '-'}
			</span>
		</div>
	</div>

	{#if loading}
		<p class="mt-3 text-sm text-muted-foreground">Memuat riwayat versi...</p>
	{:else if versions.length === 0}
		<p class="mt-3 text-sm text-muted-foreground">Belum ada riwayat versi lain dari backend.</p>
	{:else}
		<div class="mt-4 space-y-2">
			{#each versions as version (version.id)}
				<a
					href={hrefForVersion(version.id)}
					aria-current={isCurrent(version) ? 'page' : undefined}
					class="block rounded-lg border p-3 transition hover:border-primary/30 hover:bg-primary/5 {isCurrent(version) ? 'border-primary/30 bg-primary/10' : 'border-border bg-card'}"
				>
					<div class="flex flex-col gap-2 md:flex-row md:items-start md:justify-between">
						<div class="min-w-0">
							<div class="flex flex-wrap items-center gap-2">
								<span class="text-sm font-black text-foreground">{questionVersionLabel(version)}</span>
								{#if isCurrent(version)}
									<span class="rounded-full bg-primary px-2 py-0.5 text-xs font-semibold text-primary-foreground">current</span>
								{/if}
								{#if version.is_latest_version}
									<span class="rounded-full bg-success/10 px-2 py-0.5 text-xs font-semibold text-success">latest</span>
								{/if}
								<span class="rounded px-1.5 py-0.5 text-xs font-semibold {workflowClass(version.workflow_status ?? '')}">
									{WORKFLOW_LABEL[version.workflow_status ?? ''] ?? version.workflow_status ?? version.status ?? '-'}
								</span>
							</div>
							<p class="mt-1 text-sm leading-5 text-muted-foreground">{compactVersionNote(version)}</p>
						</div>
						<div class="shrink-0 text-left text-xs text-muted-foreground md:text-right">
							<p>{composerDateTimeLabel(version.updated_at ?? version.created_at)}</p>
							<p class="mt-1">{actorLabel(version)}</p>
						</div>
					</div>
				</a>
			{/each}
		</div>
	{/if}
</section>
