<script lang="ts">
	type StatusCard = { label: string; value: number; tone: string; helper: string };

	let { cards }: { cards: StatusCard[] } = $props();

	function statusCardClass(tone: string) {
		const base = 'rounded-lg border bg-white px-3 py-2';
		if (tone === 'red') return `${base} border-red-100 text-red-800`;
		if (tone === 'amber') return `${base} border-amber-100 text-amber-800`;
		if (tone === 'green' || tone === 'emerald') return `${base} border-emerald-100 text-emerald-800`;
		return `${base} border-slate-200`;
	}
</script>

<section class="grid gap-2 md:grid-cols-5" aria-label="Ringkasan status bank soal">
	{#each cards as card (card.label)}
		<div class={statusCardClass(card.tone)}>
			<div class="flex items-baseline justify-between gap-2">
				<p class="truncate text-[11px] font-semibold uppercase tracking-wider text-slate-500">{card.label}</p>
				<p class="text-base font-bold text-slate-900">{card.value}</p>
			</div>
			<p class="mt-0.5 truncate text-[11px] text-slate-500">{card.helper}</p>
		</div>
	{/each}
</section>
