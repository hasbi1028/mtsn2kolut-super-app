<script lang="ts">
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/utils';

	type RailItem = {
		label: string;
		value: string | number;
		helper?: string;
		tone?: 'default' | 'success' | 'warning' | 'danger';
	};

	let {
		title = 'Status',
		items = [],
		children,
		class: className = '',
	}: {
		title?: string;
		items?: RailItem[];
		children?: Snippet;
		class?: string;
	} = $props();

	function toneClass(tone: RailItem['tone']) {
		if (tone === 'success') return 'text-primary';
		if (tone === 'warning') return 'text-warning';
		if (tone === 'danger') return 'text-destructive';
		return 'text-foreground';
	}
</script>

{#if items.length > 0 || children}
	<aside class={cn('rounded-lg border border-border bg-card p-4 shadow-sm', className)} aria-labelledby="status-rail-title">
		<h2 id="status-rail-title" class="text-sm font-semibold text-foreground">{title}</h2>
		{#if items.length > 0}
			<dl class="mt-3 divide-y divide-border">
				{#each items as item (item.label)}
					<div class="py-3 first:pt-0 last:pb-0">
						<dt class="text-xs font-semibold uppercase tracking-[0.12em] text-muted-foreground">{item.label}</dt>
						<dd class={cn('mt-1 text-base font-semibold', toneClass(item.tone))}>{item.value}</dd>
						{#if item.helper}
							<p class="mt-1 text-xs leading-5 text-muted-foreground">{item.helper}</p>
						{/if}
					</div>
				{/each}
			</dl>
		{/if}
		{#if children}
			<div class="mt-4">
				{@render children()}
			</div>
		{/if}
	</aside>
{/if}
