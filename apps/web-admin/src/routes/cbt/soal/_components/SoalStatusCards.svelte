<script lang="ts">
	type StatusCard = { label: string; value: number; tone: string; helper: string };

	let { cards }: { cards: StatusCard[] } = $props();

	function statusCardClass(tone: string) {
		const base = 'rounded-xl border bg-white p-4 shadow-sm';
		if (tone === 'red') return `${base} border-red-100`;
		if (tone === 'amber') return `${base} border-amber-100`;
		if (tone === 'green' || tone === 'emerald') return `${base} border-emerald-100`;
		return `${base} border-slate-200`;
	}
</script>

<section class="grid gap-3 md:grid-cols-5" aria-label="Ringkasan status bank soal">
	{#each cards as card (card.label)}
		<div class={statusCardClass(card.tone)}>
			<p class="text-xs font-semibold uppercase tracking-wider text-slate-500">{card.label}</p>
			<p class="mt-2 text-2xl font-bold text-slate-900">{card.value}</p>
			<p class="mt-1 text-[11px] text-slate-500">{card.helper}</p>
		</div>
	{/each}
</section>
