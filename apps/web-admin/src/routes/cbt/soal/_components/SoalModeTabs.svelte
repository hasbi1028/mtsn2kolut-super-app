<script lang="ts">
	type ModuleMode = 'catalog' | 'composer' | 'review' | 'import';
	type ModeTab = { id: ModuleMode; label: string; desc: string };

	let {
		modes,
		activeMode,
		onSelect
	}: {
		modes: ModeTab[];
		activeMode: ModuleMode;
		onSelect: (mode: ModuleMode) => void;
	} = $props();

	function moduleModeClass(mode: ModuleMode) {
		const base = 'rounded-lg border px-3 py-2.5 text-left shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-emerald-700';
		if (activeMode === mode) return `${base} border-emerald-300 bg-emerald-50 text-emerald-950`;
		return `${base} border-slate-200 bg-white text-slate-600 hover:border-emerald-200 hover:bg-emerald-50`;
	}
</script>

<nav class="grid gap-2 md:grid-cols-4" aria-label="Mode kerja bank soal">
	{#each modes as mode (mode.id)}
		<button
			type="button"
			onclick={() => onSelect(mode.id)}
			class={moduleModeClass(mode.id)}
			aria-pressed={activeMode === mode.id}
			aria-label={`Buka mode ${mode.label}: ${mode.desc}`}
		>
			<div class="text-xs font-bold uppercase tracking-wider">{mode.label}</div>
			<div class="mt-1 text-[11px] text-slate-500">{mode.desc}</div>
		</button>
	{/each}
</nav>
