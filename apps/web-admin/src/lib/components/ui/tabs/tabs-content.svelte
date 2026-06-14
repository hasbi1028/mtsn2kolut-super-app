<script lang="ts">
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/utils';

	let {
		value,
		class: className,
		children,
		...restProps
	}: {
		value: string;
		class?: string;
		children?: Snippet;
		[key: string]: unknown;
	} = $props();
</script>

<div class={cn('mt-2', className)} role="tabpanel" {...restProps}>
	{@render children?.()}
</div>
