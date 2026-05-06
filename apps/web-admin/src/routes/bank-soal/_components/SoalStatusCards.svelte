<script lang="ts">
	type StatusCard = {
		label: string;
		value: number;
		tone: string;
		helper: string;
		workflowStatus?: string;
		status?: string;
		active?: boolean;
	};

	let {
		cards,
		onSelect
	}: {
		cards: StatusCard[];
		onSelect?: (card: StatusCard) => void;
	} = $props();

	function statusCardClass(tone: string, active = false) {
		const base = 'rounded-lg border bg-card px-3 py-2 text-left transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/30';
		const activeClass = active ? ' ring-2 ring-primary/30 ring-offset-1' : ' hover:border-primary/20 hover:bg-primary/10';
		if (tone === 'red') return `${base} border-destructive/30 text-destructive${activeClass}`;
		if (tone === 'amber') return `${base} border-warning/30 text-warning${activeClass}`;
		if (tone === 'green' || tone === 'emerald') return `${base} border-primary/20 text-primary${activeClass}`;
		return `${base} border-border${activeClass}`;
	}
</script>

<section class="grid gap-2 md:grid-cols-5" aria-label="Ringkasan status bank soal">
	{#each cards as card (card.label)}
		<button type="button" class={statusCardClass(card.tone, card.active)} aria-pressed={card.active} onclick={() => onSelect?.(card)}>
			<div class="flex items-baseline justify-between gap-2">
				<p class="truncate text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">{card.label}</p>
				<p class="text-base font-bold text-foreground">{card.value}</p>
			</div>
			<p class="mt-0.5 truncate text-[11px] text-muted-foreground">{card.helper}</p>
		</button>
	{/each}
</section>
