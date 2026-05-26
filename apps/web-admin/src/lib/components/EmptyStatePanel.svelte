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
	class={`parchment-texture page-enter rounded-[1.35rem] border border-dashed border-[var(--gold)]/30 bg-muted/50 text-center shadow-sm shadow-primary/5 ${
		compact ? 'px-4 py-8' : 'px-5 py-10'
	}`}
>
	<div class="mx-auto mb-4 flex size-14 items-center justify-center rounded-[1.15rem] border border-primary/15 bg-primary/10 text-primary">
		<svg class="size-7" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.8" d="M4 6.5C5.7 5.5 7.5 5 9.2 5c1 0 1.9.2 2.8.6.9-.4 1.8-.6 2.8-.6 1.7 0 3.5.5 5.2 1.5v11.7c-1.7-1-3.5-1.5-5.2-1.5-1 0-1.9.2-2.8.6-.9-.4-1.8-.6-2.8-.6-1.7 0-3.5.5-5.2 1.5V6.5Z" />
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.8" d="M12 5.6v11.7" />
		</svg>
	</div>
	<p class="text-[11px] font-semibold uppercase tracking-[0.24em] text-primary">{eyebrow}</p>
	<h3 class="mt-3 font-[var(--font-display)] text-xl font-semibold text-foreground">{title}</h3>
	<p class="mx-auto mt-2 max-w-2xl text-sm leading-6 text-muted-foreground">{description}</p>
	{#if children}
		<div class="mt-4 flex flex-wrap items-center justify-center gap-2">
			{@render children()}
		</div>
	{/if}
</div>
