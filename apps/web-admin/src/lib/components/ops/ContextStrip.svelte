<script lang="ts">
	import { cn } from '$lib/utils';

	type ContextItem = {
		label: string;
		value: string | number;
		tone?: 'default' | 'success' | 'warning' | 'danger' | 'muted';
	};

	let {
		items = [],
		class: className = '',
	}: {
		items?: ContextItem[];
		class?: string;
	} = $props();

	function toneClass(tone: ContextItem['tone']) {
		if (tone === 'success') return 'border-primary/20 bg-primary/10 text-primary';
		if (tone === 'warning') return 'border-warning/30 bg-warning/10 text-warning';
		if (tone === 'danger') return 'border-destructive/30 bg-destructive/10 text-destructive';
		if (tone === 'muted') return 'border-border bg-muted/50 text-muted-foreground';
		return 'border-border bg-card text-foreground';
	}
</script>

{#if items.length > 0}
	<dl class={cn('flex flex-wrap gap-2 rounded-lg border border-border bg-muted/40 p-2', className)}>
		{#each items.slice(0, 5) as item (item.label)}
			<div class={cn('rounded-md border px-3 py-2', toneClass(item.tone))}>
				<dt class="text-[10px] font-semibold uppercase tracking-[0.14em] opacity-70">{item.label}</dt>
				<dd class="mt-0.5 text-sm font-semibold">{item.value}</dd>
			</div>
		{/each}
	</dl>
{/if}
