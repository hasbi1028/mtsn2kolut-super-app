<script lang="ts">
	import type { HTMLAttributes } from 'svelte/elements';
	import { cn, type WithElementRef } from '$lib/utils.js';

	let {
		ref = $bindable(null),
		class: className,
		children,
		...restProps
	}: WithElementRef<HTMLAttributes<HTMLSpanElement>> = $props();
</script>

<span
	bind:this={ref}
	aria-current="page"
	class={cn('text-foreground inline-flex max-w-[14rem] truncate font-medium sm:max-w-[20rem]', className)}
	{...restProps}
>
	{@render children?.()}
</span>
