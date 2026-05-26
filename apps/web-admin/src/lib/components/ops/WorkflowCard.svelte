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
		if (tone === 'warning') return 'border-[var(--gold)]/30 bg-[var(--gold)]/10';
		if (tone === 'danger') return 'border-destructive/30 bg-destructive/10';
		return 'border-[var(--gold)]/20 bg-card';
	}
</script>

<article class={cn('parchment-texture page-enter-stagger-2 flex min-h-44 flex-col rounded-[1.15rem] border p-4 shadow-sm shadow-primary/5 transition duration-200 hover:-translate-y-0.5 hover:shadow-md', toneClass(), className)}>
	<div class="flex items-start justify-between gap-3">
		<div class="min-w-0">
			<h3 class="font-[var(--font-display)] text-lg font-semibold text-foreground">{title}</h3>
			<p class="mt-2 text-sm leading-6 text-muted-foreground">{description}</p>
		</div>
		{#if blockerCount > 0}
			<Badge variant="outline" class="border-[var(--gold)]/30 bg-[var(--gold)]/10 text-[var(--gold)]">{blockerCount} cek</Badge>
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
			<Button href={href} variant="outline" class="w-full justify-center border-[var(--gold)]/30 hover:bg-[var(--gold)]/10 hover:text-[var(--gold)]">{actionLabel}</Button>
		{:else}
			<p class="text-xs font-medium text-muted-foreground">{actionLabel}</p>
		{/if}
	</div>
</article>
