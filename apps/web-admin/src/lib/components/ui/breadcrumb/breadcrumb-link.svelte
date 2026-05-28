<script lang="ts">
	import type { HTMLAnchorAttributes } from 'svelte/elements';
	import { cn, type WithElementRef } from '$lib/utils.js';

	let {
		ref = $bindable(null),
		class: className,
		children,
		href,
		...restProps
	}: WithElementRef<HTMLAnchorAttributes> = $props();
</script>

<a
	bind:this={ref}
	{href}
	class={cn('hover:text-foreground inline-flex max-w-[12rem] items-center truncate transition-colors sm:max-w-[16rem]', className)}
	{...restProps}
>
	{@render children?.()}
</a>
