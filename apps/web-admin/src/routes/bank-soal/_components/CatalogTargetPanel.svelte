<script lang="ts">
	import { Input } from '$lib/components/ui/input';
	import LoadingButton from '$lib/components/LoadingButton.svelte';

	type QuestionTarget = {
		subject_id: string;
		subject_name?: string;
		subject_code?: string;
		target_questions: number;
		draft: number;
		review: number;
		rejected: number;
		approved: number;
		published: number;
	};

	let {
		selectedEventTitle,
		questionTargets,
		eventTargetTotal,
		eventPublishedTotal,
		filterSubject,
		canReviewWorkflow,
		targetBusy,
		targetQuestionsInput = $bindable(0),
		selectedTarget,
		selectedTargetShortage,
		onSaveTarget
	}: {
		selectedEventTitle: string;
		questionTargets: QuestionTarget[];
		eventTargetTotal: number;
		eventPublishedTotal: number;
		filterSubject: string;
		canReviewWorkflow: boolean;
		targetBusy: boolean;
		targetQuestionsInput: number;
		selectedTarget: QuestionTarget | null;
		selectedTargetShortage: number;
		onSaveTarget: () => void;
	} = $props();
</script>

<details class="rounded-xl border border-slate-200 bg-white/80 shadow-sm">
	<summary class="flex cursor-pointer flex-col gap-2 px-4 py-3 marker:text-slate-400 lg:flex-row lg:items-center lg:justify-between">
		<span class="min-w-0">
			<span class="block text-[11px] font-bold uppercase tracking-wider text-slate-500">Target kegiatan</span>
			<span class="mt-0.5 block truncate text-sm font-semibold text-slate-800">{selectedEventTitle}</span>
		</span>
		<span class="flex flex-wrap gap-2 text-xs text-slate-600">
			<span class="rounded-full border border-slate-200 bg-white px-2 py-1">Target {eventTargetTotal}</span>
			<span class="rounded-full border border-emerald-100 bg-white px-2 py-1 text-emerald-800">Terbit {eventPublishedTotal}</span>
			<span class="rounded-full border border-amber-100 bg-white px-2 py-1 text-amber-800">Kurang {Math.max(0, eventTargetTotal - eventPublishedTotal)}</span>
		</span>
	</summary>
	<div class="border-t border-slate-100 px-4 pb-4 pt-3">
		<p class="mb-3 text-xs text-slate-500">Target ini hanya membantu kesiapan paket/event; daftar Bank Soal tetap menjadi repositori reusable.</p>
		<div class="grid gap-3 lg:grid-cols-[minmax(0,1fr)_minmax(18rem,0.42fr)]">
			<div class="grid gap-2 md:grid-cols-2 xl:grid-cols-3">
				{#each questionTargets as target (target.subject_id)}
					<div class="rounded-lg border border-slate-200 bg-slate-50 p-3">
						<div class="flex items-start justify-between gap-2">
							<div>
								<p class="text-sm font-semibold text-slate-900">{target.subject_name ?? target.subject_code ?? target.subject_id}</p>
								<p class="text-[11px] text-slate-500">Target {target.target_questions} · kurang {Math.max(0, target.target_questions - target.published)}</p>
							</div>
							<span class="rounded bg-white px-2 py-1 text-xs font-bold text-green-700">{target.published}/{target.target_questions}</span>
						</div>
						<div class="mt-2 h-2 overflow-hidden rounded-full bg-white">
							<div class="h-2 rounded-full bg-green-600" style:width={`${target.target_questions > 0 ? Math.min(100, Math.round((target.published / target.target_questions) * 100)) : 0}%`}></div>
						</div>
						<p class="mt-2 text-[11px] text-slate-500">Draft {target.draft} · Review {target.review} · Revisi {target.rejected} · Disetujui {target.approved}</p>
					</div>
				{:else}
					<div class="rounded-lg border border-dashed border-slate-200 px-3 py-4 text-sm text-slate-400">Belum ada target kebutuhan soal untuk event ini.</div>
				{/each}
			</div>
			<div class="rounded-lg border border-slate-200 bg-slate-50 p-3">
				<p class="text-xs font-bold uppercase tracking-wider text-green-800">Atur Kebutuhan Mapel</p>
				<p class="mt-1 text-xs text-slate-600">Pilih mapel di filter konteks, lalu isi jumlah soal terbit yang dibutuhkan event.</p>
				<div class="mt-3 grid gap-2 sm:grid-cols-[1fr_auto]">
					<Input id="target-questions" type="number" min="0" bind:value={targetQuestionsInput} aria-label="Jumlah target soal terbit untuk mapel terpilih" disabled={!filterSubject || !canReviewWorkflow} class="h-9 bg-white" />
					<LoadingButton onclick={onSaveTarget} loading={targetBusy} loadingLabel="Simpan..." disabled={targetBusy || !filterSubject || !canReviewWorkflow} class="bg-green-700 text-white hover:bg-green-800 disabled:opacity-50">Simpan Kebutuhan</LoadingButton>
				</div>
				{#if selectedTarget}
					<p class="mt-2 text-xs text-green-900">Mapel terpilih kurang <span class="font-bold">{selectedTargetShortage}</span> soal terbit dari target {selectedTarget.target_questions}.</p>
				{:else if filterSubject}
					<p class="mt-2 text-xs text-green-900">Mapel ini belum punya kebutuhan. Simpan untuk membuat kebutuhan baru.</p>
				{:else}
					<p class="mt-2 text-xs text-green-900">Pilih mapel spesifik untuk melihat shortage dan mengatur target.</p>
				{/if}
			</div>
		</div>
	</div>
</details>
