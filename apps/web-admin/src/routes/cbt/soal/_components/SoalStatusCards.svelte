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
		const base = 'rounded-lg border bg-white px-3 py-2 text-left transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-emerald-700';
		const activeClass = active ? ' ring-2 ring-emerald-600 ring-offset-1' : ' hover:border-emerald-200 hover:bg-emerald-50/40';
		if (tone === 'red') return `${base} border-red-100 text-red-800${activeClass}`;
		if (tone === 'amber') return `${base} border-amber-100 text-amber-800${activeClass}`;
		if (tone === 'green' || tone === 'emerald') return `${base} border-emerald-100 text-emerald-800${activeClass}`;
		return `${base} border-slate-200${activeClass}`;
	}
</script>

<section class="grid gap-2 md:grid-cols-5" aria-label="Ringkasan status bank soal">
	{#each cards as card (card.label)}
		<button type="button" class={statusCardClass(card.tone, card.active)} aria-pressed={card.active} onclick={() => onSelect?.(card)}>
			<div class="flex items-baseline justify-between gap-2">
				<p class="truncate text-[11px] font-semibold uppercase tracking-wider text-slate-500">{card.label}</p>
				<p class="text-base font-bold text-slate-900">{card.value}</p>
			</div>
			<p class="mt-0.5 truncate text-[11px] text-slate-500">{card.helper}</p>
		</button>
	{/each}
</section>
