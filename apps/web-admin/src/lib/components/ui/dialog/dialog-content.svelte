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

<div class={cn('bg-surface-100-800 rounded-lg border border-surface-300-600 p-6 shadow-lg max-w-lg w-full', className)} {...restProps}>
	{@render children?.()}
</div>
