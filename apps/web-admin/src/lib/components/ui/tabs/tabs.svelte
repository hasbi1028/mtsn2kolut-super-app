<script lang="ts">
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/utils';

	let {
		value = $bindable(''),
		class: className,
		children,
		...restProps
	}: {
		value?: string;
		class?: string;
		children?: Snippet;
		[key: string]: unknown;
	} = $props();
</script>

<div class={cn('tabs', className)} data-value={value} {...restProps}>
	{@render children?.()}
</div>
