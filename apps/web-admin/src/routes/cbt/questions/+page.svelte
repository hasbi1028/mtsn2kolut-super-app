<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import * as Card from '$lib/components/ui/card';
	import {
		QUESTION_VARIANTS,
		evaluationStorageKey,
		type QuestionVariantEvaluation,
	} from '$lib/cbt/question-experiments';

	type EvaluationSummary = Partial<QuestionVariantEvaluation> & { average: number };

	let evaluations = $state<Record<string, EvaluationSummary>>({});

	function loadEvaluations() {
		const next: Record<string, EvaluationSummary> = {};
		for (const variant of QUESTION_VARIANTS) {
			const raw = localStorage.getItem(evaluationStorageKey(variant.id));
			if (!raw) continue;
			try {
				const parsed = JSON.parse(raw) as Partial<QuestionVariantEvaluation>;
				const scores = [parsed.ease, parsed.speed, parsed.review, parsed.fit].filter((value): value is number => typeof value === 'number');
				next[variant.id] = {
					...parsed,
					average: scores.length > 0 ? scores.reduce((sum, value) => sum + value, 0) / scores.length : 0,
				};
			} catch {
				// Ignore malformed local experiment notes.
			}
		}
		evaluations = next;
	}

	onMount(() => {
		loadEvaluations();
	});
</script>

<svelte:head>
	<title>Eksperimen Bank Soal CBT — MTsN 2 Kolaka Utara</title>
</svelte:head>

<div class="space-y-6">
	<div class="space-y-2">
		<div class="flex flex-wrap items-center gap-2 text-xs font-semibold uppercase tracking-[0.22em] text-emerald-700">
			<Badge variant="outline">6 Varian Frontend</Badge>
			<span>1 Backend Tetap</span>
		</div>
		<h1 class="text-2xl font-semibold tracking-tight text-slate-900">Lab Eksperimen Bank Soal CBT</h1>
		<p class="max-w-4xl text-sm text-slate-600">
			Halaman ini dipakai untuk membandingkan enam pendekatan UI authoring bank soal dengan kontrak backend yang sama. Guru bisa mencoba tiap varian, lalu menyimpan catatan penilaian lokal di browser.
		</p>
	</div>

	<div class="grid gap-4 lg:grid-cols-3">
		<div class="rounded-2xl border border-emerald-200 bg-gradient-to-br from-white via-emerald-50/40 to-emerald-100/60 px-4 py-4 shadow-sm">
			<p class="text-xs font-semibold uppercase tracking-[0.18em] text-emerald-800">Tujuan</p>
			<p class="mt-3 text-sm text-slate-700">Mencari pendekatan yang paling disukai guru tanpa memecah backend atau schema bank soal.</p>
		</div>
		<div class="rounded-2xl border border-slate-200 bg-white px-4 py-4 shadow-sm">
			<p class="text-xs font-semibold uppercase tracking-[0.18em] text-slate-600">Yang Sama</p>
			<p class="mt-3 text-sm text-slate-700">Semua varian memakai endpoint, validasi, mode beginner/advance, preview peserta, dan workflow yang sama.</p>
		</div>
		<div class="rounded-2xl border border-slate-200 bg-white px-4 py-4 shadow-sm">
			<p class="text-xs font-semibold uppercase tracking-[0.18em] text-slate-600">Yang Dibandingkan</p>
			<p class="mt-3 text-sm text-slate-700">Cara guru memahami alur, menulis cepat, mengedit ulang, dan mengajukan soal ke review.</p>
		</div>
	</div>

	<div class="grid gap-5 xl:grid-cols-2">
		{#each QUESTION_VARIANTS as variant (variant.id)}
			{@const evaluation = evaluations[variant.id]}
			<Card.Root class="border-slate-200 shadow-sm">
				<Card.Header class="space-y-3 border-b bg-white/90">
					<div class="flex flex-wrap items-start justify-between gap-3">
						<div class="space-y-2">
							<div class="flex flex-wrap items-center gap-2">
								<Card.Title class="text-lg text-slate-900">{variant.title}</Card.Title>
								<Badge variant="outline">Beginner + Advance</Badge>
							</div>
							<Card.Description>{variant.description}</Card.Description>
						</div>
						<div class="rounded-full border border-emerald-200 bg-emerald-50 px-3 py-1 text-xs font-semibold uppercase tracking-[0.18em] text-emerald-800">
							{variant.persona}
						</div>
					</div>
				</Card.Header>
				<Card.Content class="space-y-4 pt-5">
					<div class="grid gap-4 md:grid-cols-2">
						<div class="space-y-2">
							<p class="text-xs font-semibold uppercase tracking-[0.18em] text-emerald-800">Kelebihan</p>
							<ul class="space-y-2 text-sm text-slate-700">
								{#each variant.highlights as highlight (highlight)}
									<li class="rounded-lg border border-emerald-100 bg-emerald-50/60 px-3 py-2">{highlight}</li>
								{/each}
							</ul>
						</div>
						<div class="space-y-2">
							<p class="text-xs font-semibold uppercase tracking-[0.18em] text-slate-600">Risiko</p>
							<ul class="space-y-2 text-sm text-slate-700">
								{#each variant.cautions as caution (caution)}
									<li class="rounded-lg border border-slate-200 bg-slate-50 px-3 py-2">{caution}</li>
								{/each}
							</ul>
						</div>
					</div>

					<div class="rounded-xl border border-slate-200 bg-slate-50/80 px-4 py-3">
						{#if evaluation}
							<p class="text-sm font-semibold text-slate-900">Rata-rata lokal: {evaluation.average.toFixed(1)} / 5</p>
							<p class="mt-1 text-xs text-slate-500">
								Terakhir disimpan {evaluation.updatedAt ? new Date(evaluation.updatedAt).toLocaleString('id-ID') : '—'}
							</p>
							{#if evaluation.note}
								<p class="mt-3 text-sm text-slate-700">{evaluation.note}</p>
							{/if}
						{:else}
							<p class="text-sm text-slate-600">Belum ada catatan lokal untuk varian ini.</p>
						{/if}
					</div>

					<div class="flex flex-wrap gap-2">
						<a href={resolve(`/cbt/questions/${variant.id}`)}>
							<Button>Coba Halaman</Button>
						</a>
					</div>
				</Card.Content>
			</Card.Root>
		{/each}
	</div>
</div>
