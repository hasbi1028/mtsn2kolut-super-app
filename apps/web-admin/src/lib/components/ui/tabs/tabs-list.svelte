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

<div class={cn('border-b border-surface-300-600 flex', className)} role="tablist" {...restProps}>
	{@render children?.()}
</div>
