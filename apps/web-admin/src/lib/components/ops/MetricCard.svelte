<script lang="ts">
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/utils';

	type MetricTone = 'default' | 'success' | 'warning' | 'danger' | 'muted';

	let {
		label,
		value,
		helper = '',
		tone = 'default',
		active = false,
		onclick,
		children,
		class: className = '',
	}: {
		label: string;
		value: string | number;
		helper?: string;
		tone?: MetricTone;
		active?: boolean;
		onclick?: () => void;
		children?: Snippet;
		class?: string;
	} = $props();

	function toneClass() {
		if (tone === 'success') return 'border-primary/20 bg-primary/10';
		if (tone === 'warning') return 'border-warning/30 bg-warning/10';
		if (tone === 'danger') return 'border-destructive/30 bg-destructive/10';
		if (tone === 'muted') return 'border-border bg-muted/50';
		return 'border-border bg-card';
	}

	function valueClass() {
		if (tone === 'success') return 'text-primary';
		if (tone === 'warning') return 'text-warning';
		if (tone === 'danger') return 'text-destructive';
		return 'text-foreground';
	}
</script>

{#if onclick}
	<button
		type="button"
		aria-pressed={active}
		onclick={onclick}
		class={cn(
			'rounded-lg border p-4 text-left shadow-sm transition hover:border-primary/30 hover:bg-primary/10 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring',
			active ? 'ring-2 ring-primary/30' : '',
			toneClass(),
			className
		)}
	>
		<div class="flex items-start justify-between gap-3">
			<div class="min-w-0">
				<p class="text-xs font-semibold uppercase tracking-[0.14em] text-muted-foreground">{label}</p>
				<p class={cn('mt-2 text-3xl font-semibold', valueClass())}>{value}</p>
				{#if helper}
					<p class="mt-1 text-xs leading-5 text-muted-foreground">{helper}</p>
				{/if}
			</div>
			{#if children}
				<div class="shrink-0 text-muted-foreground">
					{@render children()}
				</div>
			{/if}
		</div>
	</button>
{:else}
	<article class={cn('rounded-lg border p-4 text-left shadow-sm', toneClass(), className)}>
		<div class="flex items-start justify-between gap-3">
			<div class="min-w-0">
				<p class="text-xs font-semibold uppercase tracking-[0.14em] text-muted-foreground">{label}</p>
				<p class={cn('mt-2 text-3xl font-semibold', valueClass())}>{value}</p>
				{#if helper}
					<p class="mt-1 text-xs leading-5 text-muted-foreground">{helper}</p>
				{/if}
			</div>
			{#if children}
				<div class="shrink-0 text-muted-foreground">
					{@render children()}
				</div>
			{/if}
		</div>
	</article>
{/if}
