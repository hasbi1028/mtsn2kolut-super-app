<script lang="ts">
	type Props = {
		eyebrow?: string;
		title: string;
		description: string;
		compact?: boolean;
		children?: import('svelte').Snippet;
	};

	let {
		eyebrow = 'Belum Ada Data',
		title,
		description,
		compact = false,
		children,
	}: Props = $props();
</script>

<div
	class={`rounded-2xl border border-dashed border-slate-300 bg-slate-50 text-center ${
		compact ? 'px-4 py-8' : 'px-5 py-10'
	}`}
>
	<p class="text-[11px] font-semibold uppercase tracking-[0.24em] text-emerald-700">{eyebrow}</p>
	<h3 class="mt-3 text-base font-semibold text-slate-900">{title}</h3>
	<p class="mx-auto mt-2 max-w-2xl text-sm leading-6 text-slate-500">{description}</p>
	{#if children}
		<div class="mt-4 flex flex-wrap items-center justify-center gap-2">
			{@render children()}
		</div>
	{/if}
</div>
