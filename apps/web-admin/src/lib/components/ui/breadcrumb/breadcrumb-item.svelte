<script lang="ts">
	import type { HTMLLiAttributes } from 'svelte/elements';
	import { cn, type WithElementRef } from '$lib/utils.js';

	let {
		ref = $bindable(null),
		class: className,
		children,
		...restProps
	}: WithElementRef<HTMLLiAttributes> = $props();
</script>

<li bind:this={ref} class={cn('inline-flex min-w-0 items-center gap-1.5', className)} {...restProps}>
	{@render children?.()}
</li>
