<script lang="ts">
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/utils';

	let {
		title = '',
		description = '',
		children,
		class: className = '',
	}: {
		title?: string;
		description?: string;
		children?: Snippet;
		class?: string;
	} = $props();
</script>

{#if title || description || children}
	<section class={cn('rounded-lg border border-border bg-card p-3 shadow-sm', className)}>
		<div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
			{#if title || description}
				<div class="min-w-0">
					{#if title}
						<p class="text-sm font-semibold text-foreground">{title}</p>
					{/if}
					{#if description}
						<p class="mt-1 text-xs leading-5 text-muted-foreground">{description}</p>
					{/if}
				</div>
			{/if}
			{#if children}
				<div class="flex flex-wrap gap-2 md:justify-end">
					{@render children()}
				</div>
			{/if}
		</div>
	</section>
{/if}
