<script lang="ts">
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/utils';

	let {
		class: className,
		children,
		...restProps
	}: {
		class?: string;
		children?: Snippet;
		[key: string]: unknown;
	} = $props();
</script>

<div class={cn('inline-flex h-10 items-center justify-center rounded-lg border border-border bg-muted p-1 text-muted-foreground', className)} role="tablist" {...restProps}>
	{@render children?.()}
</div>
