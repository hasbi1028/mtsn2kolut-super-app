<script lang="ts">
	type QualitySignal = {
		label: string;
		status: string;
		desc: string;
	};

	let {
		signals,
		actionForSignal,
	}: {
		signals: QualitySignal[];
		actionForSignal: (signal: QualitySignal) => string;
	} = $props();
</script>

<section aria-label="Checklist kualitas soal">
	<p class="mb-2 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
		Checklist Kualitas
	</p>
	<div class="space-y-2">
		{#each signals as signal (signal.label)}
			<article class="rounded-lg border p-2.5 {signal.status === 'good' ? 'border-success/20 bg-success/10' : 'border-warning/30 bg-warning/10'}">
				<div class="flex items-start gap-2">
					<span
						class="mt-0.5 flex h-5 w-5 shrink-0 items-center justify-center rounded-full text-xs font-black {signal.status === 'good' ? 'bg-success text-background' : 'bg-warning text-background'}"
						aria-hidden="true"
					>
						{signal.status === 'good' ? '✓' : '!'}
					</span>
					<div class="min-w-0">
						<h4 class="text-sm font-semibold text-foreground">{signal.label}</h4>
						<p class="mt-0.5 text-xs text-muted-foreground">{signal.desc}</p>
						<p class="mt-1 text-xs font-medium {signal.status === 'good' ? 'text-success' : 'text-warning'}">
							{actionForSignal(signal)}
						</p>
					</div>
				</div>
			</article>
		{/each}
	</div>
</section>
