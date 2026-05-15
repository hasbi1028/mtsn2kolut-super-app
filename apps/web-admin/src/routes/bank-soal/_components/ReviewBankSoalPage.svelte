<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { Textarea } from '$lib/components/ui/textarea';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import RichContent from '$lib/components/RichContent.svelte';
	import { toast } from '$lib/components/ui/sonner';
	import { canReviewBankSoal, type BankSoalAccessUser } from '$lib/bank-soal/access';
	import { clientApiPath, clientApiPathWithQuery, readClientApiData, readClientJson } from '$lib/client/api';
	import { htmlToPlainText } from '$lib/utils/html-text';
	import { displayName } from '$lib/utils/display-name';

	type PageData = { user?: BankSoalAccessUser };
	type OptionItem = { label?: string; text?: string; html?: string; latex?: string; match_label?: string; match_text?: string; match_html?: string; is_distractor?: boolean };
	type Question = {
		id: string;
		subject_name?: string;
		subject_code?: string;
		code?: string;
		question_type: string;
		stem_html?: string;
		question_text?: string;
		stimulus_html?: string;
		explanation_html?: string;
		rubric_html?: string;
		options?: OptionItem[];
		answer_key?: string;
		workflow_status: string;
		status: string;
		author_username?: string;
		author_display_name?: string;
		review_notes?: string;
	};
	type QuestionListResponse = { items: Question[]; meta?: { total: number } };
	type TimelineItem = { id?: string; action?: string; status?: string; notes?: string; actor_username?: string; actor_display_name?: string; created_at?: string };
	type EventContext = { id: string; title: string; status: string; academic_year_name?: string };

	let { data }: { data?: PageData } = $props();
	let queuePromise = $state<Promise<QuestionListResponse> | null>(null);
	let queue = $state<Question[]>([]);
	let eventContext = $state<EventContext | null>(null);
	let activeIndex = $state(0);
	let activeDetail = $state<Question | null>(null);
	let timeline = $state<TimelineItem[]>([]);
	let notes = $state('');
	let busyAction = $state<'approve' | 'reject' | ''>('');
	const eventId = page.url.searchParams.get('event_id') ?? '';
	const subjectId = page.url.searchParams.get('subject_id') ?? '';
	const requestedQuestionId = page.url.searchParams.get('question_id') ?? '';

	let activeQuestion = $derived(activeDetail ?? queue[activeIndex] ?? null);
	let canReview = $derived(canReviewBankSoal(data?.user));

	function errorMessage(error: unknown, fallback: string) {
		return error instanceof Error && error.message.trim() ? error.message : fallback;
	}

	function stemPratinjau(question: Question): string {
		const text = htmlToPlainText(question.stem_html || question.question_text || '');
		return text || '(Soal kosong)';
	}

	function optionContent(option: OptionItem): string {
		return option.html || option.text || option.latex || option.match_html || option.match_text || '-';
	}

	async function fetchQueue(): Promise<QuestionListResponse> {
		const params = new URLSearchParams({ workflow_status: 'review', status: 'draft', limit: '30', offset: '0' });
		if (eventId) params.set('event_id', eventId);
		if (subjectId) params.set('subject_id', subjectId);
		return fetch(clientApiPathWithQuery('/api/bank-soal/questions', params)).then((response) => readClientApiData<QuestionListResponse>(response, 'Gagal memuat antrean review'));
	}

	async function fetchEventContext() {
		if (!eventId) return null;
		try {
			return await fetch(clientApiPath`/api/asesmen/events/${eventId}`).then((response) => readClientApiData<EventContext>(response, 'Gagal memuat konteks kegiatan'));
		} catch {
			return null;
		}
	}

	function loadQueue() {
		queuePromise = fetchQueue().then((payload) => {
			queue = payload.items ?? [];
			const requestedIndex = requestedQuestionId
				? queue.findIndex((question) => question.id === requestedQuestionId)
				: -1;
			activeIndex = requestedIndex >= 0 ? requestedIndex : 0;
			void loadActiveDetail();
			return payload;
		});
	}

	async function loadActiveDetail() {
		const question = queue[activeIndex];
		activeDetail = question ?? null;
		timeline = [];
		notes = '';
		if (!question) return;
		try {
			activeDetail = await fetch(clientApiPath`/api/bank-soal/questions/${question.id}`).then((response) => readClientApiData<Question>(response, 'Gagal memuat detail soal'));
		} catch (error) {
			toast.warning(errorMessage(error, 'Detail lengkap belum dapat dimuat.'));
		}
		try {
			const payload = await fetch(clientApiPath`/api/bank-soal/questions/${question.id}/timeline`).then((response) => readClientApiData<TimelineItem[] | { items?: TimelineItem[] }>(response, 'Gagal memuat timeline'));
			timeline = Array.isArray(payload) ? payload : payload.items ?? [];
		} catch {
			timeline = [];
		}
	}

	function move(delta: number) {
		activeIndex = Math.max(0, Math.min(queue.length - 1, activeIndex + delta));
		void loadActiveDetail();
	}

	async function decide(action: 'approve' | 'reject') {
		if (!activeQuestion) return;
		if (!canReview) {
			toast.warning('Aksi pemeriksa membutuhkan izin akses verifikasi Bank Soal.');
			return;
		}
		const trimmed = notes.trim();
		if (action === 'reject' && trimmed.length < 8) {
			toast.warning('Catatan revisi minimal 8 karakter.');
			return;
		}
		busyAction = action;
		try {
			await fetch(clientApiPath`/api/bank-soal/questions/${activeQuestion.id}/workflow`, {
				method: 'PATCH',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ action, notes: trimmed })
			}).then((response) => readClientJson<unknown>(response));
			toast.success(action === 'approve' ? 'Soal disetujui' : 'Soal dikembalikan untuk revisi');
			queue = queue.filter((item) => item.id !== activeQuestion.id);
			activeIndex = Math.min(activeIndex, Math.max(0, queue.length - 1));
			void loadActiveDetail();
		} catch (error) {
			toast.error(errorMessage(error, 'Gagal menyimpan keputusan'));
		} finally {
			busyAction = '';
		}
	}

	let reviewChecklist = $derived.by(() => {
		const q = activeQuestion;
		return [
			{ label: 'Naskah', desc: q ? `${stemPratinjau(q).length} karakter` : 'Belum ada soal aktif', ok: Boolean(q && stemPratinjau(q).length >= 5) },
			{ label: 'Kunci/Rubrik', desc: q?.answer_key || q?.rubric_html ? 'Tersedia untuk reviewer' : 'Belum terlihat', ok: Boolean(q?.answer_key || q?.rubric_html) },
			{ label: 'Pembahasan', desc: q?.explanation_html ? 'Ada pembahasan/catatan' : 'Opsional', ok: Boolean(q?.explanation_html) },
			{ label: 'Catatan', desc: notes.trim() ? `${notes.trim().length} karakter catatan` : 'Opsional approve, wajib reject', ok: notes.trim().length >= 8 },
		];
	});
	let reviewReadyCount = $derived(reviewChecklist.filter((item) => item.ok).length);
	let reviewerDecisionHint = $derived(notes.trim().length >= 8 ? 'Catatan cukup untuk revisi' : 'Isi minimal 8 karakter bila meminta revisi');

	onMount(() => {
		void fetchEventContext().then((context) => { eventContext = context; });
		loadQueue();
	});
</script>

	<svelte:head><title>Review Bank Soal</title></svelte:head>

<div class="space-y-5 p-4 md:p-6">
	<div class="overflow-hidden rounded-2xl border border-primary/20 bg-card shadow-sm">
		<div class="bg-gradient-to-r from-primary/10 via-card to-warning/10 p-4 md:p-5">
			<div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
				<div class="min-w-0">
					<p class="text-[10px] font-black uppercase tracking-[0.28em] text-primary">Ruang Review Fokus</p>
					<h1 class="mt-1 text-2xl font-black uppercase italic tracking-tight text-foreground">Periksa Bank Soal</h1>
					<p class="mt-2 max-w-3xl text-sm leading-6 text-muted-foreground">Reviewer memeriksa naskah, opsi/kunci, rubrik, pembahasan, dan timeline sebelum menyetujui atau mengembalikan soal dengan catatan revisi.</p>
					<div class="mt-3 flex flex-wrap gap-2 text-[10px] font-bold uppercase tracking-wide">
						<span class="rounded-full border border-primary/20 bg-card px-2.5 py-1 text-primary">{queue.length} antrean aktif</span>
						<span class="rounded-full border border-warning/30 bg-card px-2.5 py-1 text-warning">Checklist {reviewReadyCount}/{reviewChecklist.length}</span>
						{#if eventId}<span class="rounded-full border border-border bg-card px-2.5 py-1 text-muted-foreground">Event scoped</span>{/if}
					</div>
				</div>
				<div class="flex flex-wrap gap-2">
					{#if eventId}
						<a href={resolve(`/asesmen/kegiatan/${eventId}`)} class="inline-flex rounded-md border border-success/20 bg-success/10 px-3 py-2 text-sm font-semibold text-success hover:bg-success/15">Kembali ke Event</a>
					{/if}
					<a href={resolve('/bank-soal/daftar')} class="inline-flex rounded-md border border-input bg-background px-3 py-2 text-sm font-semibold text-foreground hover:bg-muted">Daftar Soal</a>
					<a href={resolve('/bank-soal')} class="inline-flex rounded-md border border-input bg-background px-3 py-2 text-sm font-semibold text-foreground hover:bg-muted">Dashboard</a>
				</div>
			</div>
		</div>
	</div>

	{#if eventId}
		<div class="rounded-xl border border-success/20 bg-success/10 p-4 text-sm text-success">
			<p class="font-semibold">Konteks event terbaca: {eventContext?.title ?? eventId}</p>
			<p class="mt-1 text-success">Antrean review dibatasi ke soal yang terkait kegiatan ini. Reviewer mapel hanya dapat memutuskan soal sesuai scope event yang ditetapkan panitia.</p>
		</div>
	{/if}

	<AsyncContent promise={queuePromise}>
		{#snippet pending()}
			<Skeleton class="h-96 w-full" />
		{/snippet}
		{#snippet failed(error, reset)}
			<RecoveryPanel title="Antrean Verifikasi Belum Tersaji" message={errorMessage(error, 'Gagal memuat antrean review.')} onRetry={() => { reset?.(); loadQueue(); }} />
		{/snippet}
		{#snippet children()}
			{#if !activeQuestion}
				<div class="rounded-xl border border-success/20 bg-success/10 p-6 text-center text-success">Tidak ada soal yang menunggu review.</div>
			{:else}
				<section class="grid gap-4 xl:grid-cols-[minmax(0,1fr)_22rem]">
					<article class="overflow-hidden rounded-2xl border border-border bg-card shadow-sm">
						<div class="border-b border-border bg-muted/50 p-4">
							<div class="flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
								<span class="rounded bg-success/10 px-2 py-1 font-semibold text-success">{activeQuestion.question_type}</span>
								<span>{activeQuestion.subject_name ?? activeQuestion.subject_code ?? 'Mapel belum ada'}</span>
								<span>{activeQuestion.code ?? 'Tanpa kode'}</span>
								{#if activeQuestion.author_username || activeQuestion.author_display_name}<span>Guru: {displayName({ display_name: activeQuestion.author_display_name, username: activeQuestion.author_username }, 'Guru')}</span>{/if}
							</div>
							<h2 class="mt-2 line-clamp-2 text-lg font-bold text-foreground">{stemPratinjau(activeQuestion)}</h2>
						</div>
						<div class="space-y-4 bg-muted/50 p-4">
							{#if activeQuestion.stimulus_html}<div class="rounded-xl border border-border bg-card p-3"><p class="mb-1 text-xs font-semibold uppercase tracking-wide text-muted-foreground">Stimulus</p><RichContent html={activeQuestion.stimulus_html} class="prose prose-sm max-w-none latex-preview" /></div>{/if}
							<div class="rounded-xl border border-border bg-card p-3"><p class="mb-1 text-xs font-semibold uppercase tracking-wide text-muted-foreground">Pertanyaan</p><RichContent html={activeQuestion.stem_html || activeQuestion.question_text || stemPratinjau(activeQuestion)} class="prose prose-sm max-w-none latex-preview" /></div>
							{#if activeQuestion.options?.length}
								<div class="rounded-xl border border-border bg-card p-3">
									<p class="mb-2 text-xs font-semibold uppercase tracking-wide text-muted-foreground">Opsi / Pasangan</p>
									<div class="grid gap-2 md:grid-cols-2">
										{#each activeQuestion.options as option, index (`review-option-${activeQuestion.id}-${index}`)}
											<div class="rounded-lg border border-border bg-muted/50 px-2 py-1.5 text-sm"><span class="font-bold text-success">{option.label || option.match_label || index + 1}.</span><RichContent html={optionContent(option)} class="mt-1 latex-preview" /></div>
										{/each}
									</div>
								</div>
							{/if}
							<div class="rounded-xl border border-success/20 bg-success/10 p-3 text-sm text-success"><span class="font-semibold">Kunci/Rubrik:</span> {activeQuestion.answer_key || (activeQuestion.rubric_html ? 'Rubrik tersedia' : 'Tidak tersedia')}</div>
							{#if activeQuestion.rubric_html}<div class="rounded-xl border border-warning/30 bg-warning/10 p-3"><p class="mb-1 text-xs font-semibold uppercase tracking-wide text-warning">Rubrik</p><RichContent html={activeQuestion.rubric_html} class="prose prose-sm max-w-none latex-preview" /></div>{/if}
							{#if activeQuestion.explanation_html}<div class="rounded-xl border border-border bg-card p-3"><p class="mb-1 text-xs font-semibold uppercase tracking-wide text-muted-foreground">Pembahasan</p><RichContent html={activeQuestion.explanation_html} class="prose prose-sm max-w-none latex-preview" /></div>{/if}
						</div>
						<div class="border-t border-border bg-card p-4">
							{#if canReview}
								<label for="review-notes" class="mb-1 block text-sm font-medium text-foreground">Catatan keputusan</label>
								<Textarea id="review-notes" rows={3} bind:value={notes} placeholder="Wajib untuk reject, opsional untuk approve." />
								<p class="mt-1 text-xs text-muted-foreground">{reviewerDecisionHint}</p>
							{:else}
								<div class="rounded-xl border border-warning/30 bg-warning/10 p-3 text-sm text-warning">
									<p class="font-semibold">Mode baca antrean review</p>
									<p class="mt-1 text-xs">Keputusan approve/reject membutuhkan permission reviewer Bank Soal.</p>
								</div>
							{/if}
							<div class="mt-4 flex flex-wrap justify-between gap-2 border-t border-border pt-4">
								<div class="flex gap-2"><Button variant="outline" onclick={() => move(-1)} disabled={activeIndex === 0}>Sebelumnya</Button><Button variant="outline" onclick={() => move(1)} disabled={activeIndex >= queue.length - 1}>Berikutnya</Button></div>
								{#if canReview}
									<div class="flex gap-2"><LoadingButton variant="outline" onclick={() => void decide('reject')} loading={busyAction === 'reject'} loadingLabel="Mengirim..." disabled={busyAction !== ''} class="border-destructive/30 text-destructive hover:bg-destructive/10">Minta Revisi</LoadingButton><LoadingButton onclick={() => void decide('approve')} loading={busyAction === 'approve'} loadingLabel="Menyetujui..." disabled={busyAction !== ''} class="bg-success text-background hover:bg-success">Setujui</LoadingButton></div>
								{/if}
							</div>
						</div>
					</article>
					<aside class="space-y-3">
						<div class="rounded-xl border border-success/20 bg-success/10 p-4 text-success"><p class="text-sm font-semibold">Posisi Review</p><p class="mt-1 text-2xl font-bold">{activeIndex + 1}/{queue.length}</p><p class="mt-1 text-xs text-success">Gunakan tombol berikutnya/sebelumnya atau pilih antrean di bawah.</p></div>
						<div class="rounded-xl border border-border bg-card p-4">
							<p class="mb-2 text-sm font-semibold text-foreground">Checklist Reviewer</p>
							<div class="space-y-2">{#each reviewChecklist as item (item.label)}<div class="rounded-lg border px-3 py-2 text-xs {item.ok ? 'border-primary/20 bg-primary/10 text-primary' : 'border-warning/30 bg-warning/10 text-warning'}"><div class="flex items-center justify-between gap-2"><span class="font-semibold">{item.label}</span><span>{item.ok ? 'OK' : 'Cek'}</span></div><p class="mt-1 opacity-80">{item.desc}</p></div>{/each}</div>
						</div>
						<div class="rounded-xl border border-border bg-card p-4"><p class="mb-2 text-sm font-semibold text-foreground">Timeline</p>{#if timeline.length > 0}<div class="space-y-2">{#each timeline.slice(0, 8) as item, index (`timeline-${item.id ?? index}`)}<div class="rounded border border-border bg-muted/50 px-2 py-1.5 text-xs"><div class="flex flex-wrap items-center gap-1"><p class="font-semibold text-success">{item.action ?? item.status ?? 'Perubahan'}</p>{#if item.actor_username || item.actor_display_name}<span class="text-muted-foreground">oleh {displayName({ display_name: item.actor_display_name, username: item.actor_username }, 'Pengguna')}</span>{/if}</div>{#if item.notes}<p class="mt-1 text-muted-foreground">{item.notes}</p>{/if}{#if item.created_at}<p class="mt-1 text-muted-foreground">{new Date(item.created_at).toLocaleString('id-ID')}</p>{/if}</div>{/each}</div>{:else}<p class="text-xs text-muted-foreground">Timeline belum tersedia.</p>{/if}</div>
						<div class="overflow-hidden rounded-xl border border-border bg-card"><div class="border-b border-border px-3 py-2 text-sm font-semibold text-foreground">Antrean Verifikasi</div><Table.Root><Table.Body>{#each queue.slice(0, 8) as item, index (item.id)}<Table.Row class={index === activeIndex ? 'bg-success/10' : ''}><Table.Cell><button type="button" class="block w-full text-left text-xs" onclick={() => { activeIndex = index; void loadActiveDetail(); }}>{stemPratinjau(item).slice(0, 64)}</button></Table.Cell></Table.Row>{/each}</Table.Body></Table.Root></div>
					</aside>
				</section>
			{/if}
		{/snippet}
	</AsyncContent>
</div>
