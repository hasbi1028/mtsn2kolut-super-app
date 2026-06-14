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

<h3 class={cn('text-lg font-semibold leading-none tracking-tight text-on-surface', className)} {...restProps}>
	{@render children?.()}
</h3>
