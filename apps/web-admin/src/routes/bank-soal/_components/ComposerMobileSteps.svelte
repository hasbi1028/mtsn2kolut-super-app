<script lang="ts">
	import { Button } from '$lib/components/ui/button';

	export type ComposerMobileStep = {
		id: 'metadata' | 'question' | 'answer' | 'preview';
		label: string;
		helper: string;
		complete: boolean;
		targetId: string;
	};

	let {
		steps,
		currentStepId,
		onSelect,
		onPrevious,
		onNext,
	}: {
		steps: ComposerMobileStep[];
		currentStepId: ComposerMobileStep['id'];
		onSelect: (step: ComposerMobileStep) => void;
		onPrevious: () => void;
		onNext: () => void;
	} = $props();

	let currentIndex = $derived(Math.max(0, steps.findIndex((step) => step.id === currentStepId)));
	let currentStep = $derived(steps[currentIndex] ?? steps[0]);
</script>

<section class="lg:hidden rounded-xl border border-primary/20 bg-card p-3 shadow-sm" aria-label="Langkah penyusunan soal mobile">
	<div class="flex items-center justify-between gap-3">
		<div>
			<p class="text-xs font-black uppercase tracking-[0.18em] text-primary">Langkah {currentIndex + 1}/{steps.length}</p>
			<p class="mt-0.5 text-sm font-semibold text-foreground">{currentStep?.label ?? 'Identitas Soal'}</p>
		</div>
		<div class="flex gap-1">
			<Button type="button" variant="outline" size="sm" class="h-8 px-2 text-xs" aria-label="Langkah sebelumnya" onclick={onPrevious} disabled={currentIndex <= 0}>
				Sebelumnya
			</Button>
			<Button type="button" size="sm" class="h-8 px-2 text-xs" aria-label="Langkah berikutnya" onclick={onNext} disabled={currentIndex >= steps.length - 1}>
				Berikutnya
			</Button>
		</div>
	</div>
	<div class="mt-3 grid grid-cols-4 gap-1.5">
		{#each steps as step, index (step.id)}
			<button
				type="button"
				aria-label={`Buka langkah ${step.label}`}
				aria-current={step.id === currentStepId ? 'step' : undefined}
				onclick={() => onSelect(step)}
				class="rounded-lg border px-2 py-2 text-left transition {step.id === currentStepId
					? 'border-primary bg-primary text-primary-foreground'
					: step.complete
						? 'border-success/20 bg-success/10 text-success'
						: 'border-border bg-muted/50 text-muted-foreground'}"
			>
				<span class="block text-xs font-black">{index + 1}</span>
				<span class="mt-0.5 block truncate text-xs font-semibold">{step.label}</span>
			</button>
		{/each}
	</div>
	{#if currentStep}
		<p class="mt-2 text-xs text-muted-foreground">{currentStep.helper}</p>
	{/if}
</section>
