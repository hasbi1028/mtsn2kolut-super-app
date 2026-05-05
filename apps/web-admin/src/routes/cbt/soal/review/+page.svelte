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
	import { clientApiPath, clientApiPathWithQuery, readClientApiData, readClientJson } from '$lib/client/api';
	import { htmlToPlainText } from '$lib/utils/html-text';

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
		review_notes?: string;
	};
	type QuestionListResponse = { items: Question[]; meta?: { total: number } };
	type TimelineItem = { id?: string; action?: string; status?: string; notes?: string; actor_username?: string; created_at?: string };
	type EventContext = { id: string; title: string; status: string; academic_year_name?: string };

	let queuePromise = $state<Promise<QuestionListResponse> | null>(null);
	let queue = $state<Question[]>([]);
	let eventContext = $state<EventContext | null>(null);
	let activeIndex = $state(0);
	let activeDetail = $state<Question | null>(null);
	let timeline = $state<TimelineItem[]>([]);
	let notes = $state('');
	let busyAction = $state<'approve' | 'reject' | ''>('');
	const eventId = page.url.searchParams.get('event_id') ?? '';

	let activeQuestion = $derived(activeDetail ?? queue[activeIndex] ?? null);

	function errorMessage(error: unknown, fallback: string) {
		return error instanceof Error && error.message.trim() ? error.message : fallback;
	}

	function stemPreview(question: Question): string {
		const text = htmlToPlainText(question.stem_html || question.question_text || '');
		return text || '(Soal kosong)';
	}

	function optionContent(option: OptionItem): string {
		return option.html || option.text || option.latex || option.match_html || option.match_text || '-';
	}

	async function fetchQueue(): Promise<QuestionListResponse> {
		const params = new URLSearchParams({ workflow_status: 'review', status: 'draft', limit: '30', offset: '0' });
		return fetch(clientApiPathWithQuery('/api/cbt/questions', params)).then((response) => readClientApiData<QuestionListResponse>(response, 'Gagal memuat antrean review'));
	}

	async function fetchEventContext() {
		if (!eventId) return null;
		try {
			return await fetch(clientApiPath`/api/cbt/events/${eventId}`).then((response) => readClientApiData<EventContext>(response, 'Gagal memuat konteks kegiatan'));
		} catch {
			return null;
		}
	}

	function loadQueue() {
		queuePromise = fetchQueue().then((payload) => {
			queue = payload.items ?? [];
			activeIndex = 0;
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
			activeDetail = await fetch(clientApiPath`/api/cbt/questions/${question.id}`).then((response) => readClientApiData<Question>(response, 'Gagal memuat detail soal'));
		} catch (error) {
			toast.warning(errorMessage(error, 'Detail lengkap belum dapat dimuat.'));
		}
		try {
			const payload = await fetch(clientApiPath`/api/cbt/questions/${question.id}/timeline`).then((response) => readClientApiData<TimelineItem[] | { items?: TimelineItem[] }>(response, 'Gagal memuat timeline'));
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
		const trimmed = notes.trim();
		if (action === 'reject' && trimmed.length < 8) {
			toast.warning('Catatan revisi minimal 8 karakter.');
			return;
		}
		busyAction = action;
		try {
			await fetch(clientApiPath`/api/cbt/questions/${activeQuestion.id}/workflow`, {
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

	onMount(() => {
		void fetchEventContext().then((context) => { eventContext = context; });
		loadQueue();
	});
</script>

	<svelte:head><title>Review Bank Soal CBT</title></svelte:head>

<div class="space-y-5 p-6">
	<div class="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
		<div>
			<p class="text-xs font-bold uppercase tracking-[0.18em] text-green-700">Ruang Review Fokus</p>
			<h1 class="mt-1 text-2xl font-semibold text-slate-900">Periksa Bank Soal CBT</h1>
			<p class="mt-1 max-w-2xl text-sm text-slate-500">Tampilan ini memusatkan reviewer pada satu soal dari repositori reusable, kunci/rubrik, pembahasan, dan timeline sebelum mengambil keputusan.</p>
		</div>
		<div class="flex flex-wrap gap-2">
			{#if eventId}
				<a href={resolve(`/cbt/events/${eventId}`)} class="inline-flex rounded-md border border-green-200 bg-green-50 px-3 py-2 text-sm font-semibold text-green-800 hover:bg-green-100">Kembali ke Event</a>
			{/if}
			<a href={resolve('/cbt/soal?mode=review')} class="inline-flex rounded-md border border-input bg-background px-3 py-2 text-sm font-semibold text-slate-700 hover:bg-muted">Kembali ke Bank Soal</a>
		</div>
	</div>

	{#if eventId}
		<div class="rounded-xl border border-green-200 bg-green-50/70 p-4 text-sm text-green-950">
			<p class="font-semibold">Konteks event terbaca: {eventContext?.title ?? eventId}</p>
			<p class="mt-1 text-green-800">Antrean review tetap membaca repositori Bank Soal reusable tanpa mengirim <code class="rounded bg-white px-1">event_id</code>. Gunakan konteks event hanya untuk kembali ke pusat kegiatan atau target kebutuhan.</p>
		</div>
	{/if}

	<AsyncContent promise={queuePromise}>
		{#snippet pending()}
			<Skeleton class="h-96 w-full" />
		{/snippet}
		{#snippet failed(error, reset)}
			<RecoveryPanel title="Antrean Review Belum Tersaji" message={errorMessage(error, 'Gagal memuat antrean review.')} onRetry={() => { reset?.(); loadQueue(); }} />
		{/snippet}
		{#snippet children()}
			{#if !activeQuestion}
				<div class="rounded-xl border border-green-200 bg-green-50 p-6 text-center text-green-900">Tidak ada soal yang menunggu review.</div>
			{:else}
				<section class="grid gap-4 xl:grid-cols-[minmax(0,1fr)_20rem]">
					<article class="rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
						<div class="flex flex-wrap items-center gap-2 text-xs text-slate-500">
							<span class="rounded bg-green-50 px-2 py-1 font-semibold text-green-800">{activeQuestion.question_type}</span>
							<span>{activeQuestion.subject_name ?? activeQuestion.subject_code ?? 'Mapel belum ada'}</span>
							<span>{activeQuestion.code ?? 'Tanpa kode'}</span>
							{#if activeQuestion.author_username}<span>Guru: {activeQuestion.author_username}</span>{/if}
						</div>
						<div class="mt-4 space-y-4 rounded-lg border border-slate-200 bg-slate-50 p-3">
							{#if activeQuestion.stimulus_html}<div class="rounded border border-slate-100 bg-white p-3"><p class="mb-1 text-xs font-semibold text-slate-500">Stimulus</p><RichContent html={activeQuestion.stimulus_html} class="prose prose-sm max-w-none latex-preview" /></div>{/if}
							<div class="rounded border border-slate-100 bg-white p-3"><p class="mb-1 text-xs font-semibold text-slate-500">Pertanyaan</p><RichContent html={activeQuestion.stem_html || activeQuestion.question_text || stemPreview(activeQuestion)} class="prose prose-sm max-w-none latex-preview" /></div>
							{#if activeQuestion.options?.length}
								<div class="rounded border border-slate-100 bg-white p-3">
									<p class="mb-2 text-xs font-semibold text-slate-500">Opsi / Pasangan</p>
									<div class="space-y-2">
										{#each activeQuestion.options as option, index (`review-option-${activeQuestion.id}-${index}`)}
											<div class="rounded border border-slate-100 bg-slate-50 px-2 py-1.5 text-sm"><span class="font-bold text-green-700">{option.label || option.match_label || index + 1}.</span><RichContent html={optionContent(option)} class="mt-1 latex-preview" /></div>
										{/each}
									</div>
								</div>
							{/if}
							<div class="rounded border border-green-100 bg-green-50 p-3 text-sm text-green-950"><span class="font-semibold">Kunci/Rubrik:</span> {activeQuestion.answer_key || (activeQuestion.rubric_html ? 'Rubrik tersedia' : 'Tidak tersedia')}</div>
							{#if activeQuestion.rubric_html}<div class="rounded border border-amber-100 bg-amber-50 p-3"><p class="mb-1 text-xs font-semibold text-amber-700">Rubrik</p><RichContent html={activeQuestion.rubric_html} class="prose prose-sm max-w-none latex-preview" /></div>{/if}
							{#if activeQuestion.explanation_html}<div class="rounded border border-slate-100 bg-white p-3"><p class="mb-1 text-xs font-semibold text-slate-500">Pembahasan</p><RichContent html={activeQuestion.explanation_html} class="prose prose-sm max-w-none latex-preview" /></div>{/if}
						</div>
						<div class="mt-4">
							<label for="review-notes" class="mb-1 block text-sm font-medium text-slate-700">Catatan keputusan</label>
							<Textarea id="review-notes" rows={3} bind:value={notes} placeholder="Wajib untuk reject, opsional untuk approve." />
						</div>
						<div class="mt-4 flex flex-wrap justify-between gap-2 border-t border-slate-100 pt-4">
							<div class="flex gap-2"><Button variant="outline" onclick={() => move(-1)} disabled={activeIndex === 0}>Sebelumnya</Button><Button variant="outline" onclick={() => move(1)} disabled={activeIndex >= queue.length - 1}>Berikutnya</Button></div>
							<div class="flex gap-2"><LoadingButton variant="outline" onclick={() => void decide('reject')} loading={busyAction === 'reject'} loadingLabel="Mengirim..." disabled={busyAction !== ''} class="border-red-200 text-red-700 hover:bg-red-50">Minta Revisi</LoadingButton><LoadingButton onclick={() => void decide('approve')} loading={busyAction === 'approve'} loadingLabel="Menyetujui..." disabled={busyAction !== ''} class="bg-green-700 text-white hover:bg-green-800">Setujui</LoadingButton></div>
						</div>
					</article>
					<aside class="space-y-3">
						<div class="rounded-xl border border-green-200 bg-green-50 p-4 text-green-950"><p class="text-sm font-semibold">Posisi Review</p><p class="mt-1 text-2xl font-bold">{activeIndex + 1}/{queue.length}</p></div>
						<div class="rounded-xl border border-slate-200 bg-white p-4"><p class="mb-2 text-sm font-semibold text-slate-800">Timeline</p>{#if timeline.length > 0}<div class="space-y-2">{#each timeline.slice(0, 8) as item, index (`timeline-${item.id ?? index}`)}<div class="rounded border border-slate-100 bg-slate-50 px-2 py-1.5 text-xs"><p class="font-semibold text-green-800">{item.action ?? item.status ?? 'Perubahan'}</p>{#if item.notes}<p class="mt-1 text-slate-600">{item.notes}</p>{/if}{#if item.created_at}<p class="mt-1 text-slate-400">{new Date(item.created_at).toLocaleString('id-ID')}</p>{/if}</div>{/each}</div>{:else}<p class="text-xs text-slate-400">Timeline belum tersedia.</p>{/if}</div>
						<div class="rounded-xl border border-slate-200 bg-white p-0"><Table.Root><Table.Body>{#each queue.slice(0, 8) as item, index (item.id)}<Table.Row class={index === activeIndex ? 'bg-green-50' : ''}><Table.Cell><button type="button" class="block w-full text-left text-xs" onclick={() => { activeIndex = index; void loadActiveDetail(); }}>{stemPreview(item).slice(0, 64)}</button></Table.Cell></Table.Row>{/each}</Table.Body></Table.Root></div>
					</aside>
				</section>
			{/if}
		{/snippet}
	</AsyncContent>
</div>
