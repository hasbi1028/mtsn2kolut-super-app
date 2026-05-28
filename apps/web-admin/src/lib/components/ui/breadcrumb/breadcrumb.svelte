<script lang="ts">
	import type { HTMLAttributes } from 'svelte/elements';
	import { cn, type WithElementRef } from '$lib/utils.js';

	let {
		ref = $bindable(null),
		class: className,
		children,
		'aria-label': ariaLabel = 'Breadcrumb',
		...restProps
	}: WithElementRef<HTMLAttributes<HTMLElement>> = $props();
</script>

<nav bind:this={ref} aria-label={ariaLabel} class={cn('min-w-0', className)} {...restProps}>
	{@render children?.()}
</nav>
