<script lang="ts">
	import type { Snippet } from 'svelte';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { cn } from '$lib/utils';

	type WorkflowTone = 'default' | 'success' | 'warning' | 'danger';

	let {
		title,
		description,
		status = '',
		actionLabel = 'Buka',
		href = '',
		blockerCount = 0,
		tone = 'default',
		children,
		class: className = '',
	}: {
		title: string;
		description: string;
		status?: string;
		actionLabel?: string;
		href?: string;
		blockerCount?: number;
		tone?: WorkflowTone;
		children?: Snippet;
		class?: string;
	} = $props();

	function toneClass() {
		if (tone === 'success') return 'border-primary/20 bg-primary/10';
		if (tone === 'warning') return 'border-warning/30 bg-warning/10';
		if (tone === 'danger') return 'border-destructive/30 bg-destructive/10';
		return 'border-border bg-card';
	}
</script>

<article class={cn('flex min-h-44 flex-col rounded-lg border p-4 shadow-sm', toneClass(), className)}>
	<div class="flex items-start justify-between gap-3">
		<div class="min-w-0">
			<h3 class="text-base font-semibold text-foreground">{title}</h3>
			<p class="mt-2 text-sm leading-6 text-muted-foreground">{description}</p>
		</div>
		{#if blockerCount > 0}
			<Badge variant="outline" class="border-warning/30 bg-warning/10 text-warning">{blockerCount} cek</Badge>
		{:else if status}
			<Badge variant="outline" class="border-primary/20 bg-primary/10 text-primary">{status}</Badge>
		{/if}
	</div>
	{#if children}
		<div class="mt-4 text-sm text-muted-foreground">
			{@render children()}
		</div>
	{/if}
	<div class="mt-auto pt-4">
		{#if href}
			<Button href={href} variant="outline" class="w-full justify-center">{actionLabel}</Button>
		{:else}
			<p class="text-xs font-medium text-muted-foreground">{actionLabel}</p>
		{/if}
	</div>
</article>
