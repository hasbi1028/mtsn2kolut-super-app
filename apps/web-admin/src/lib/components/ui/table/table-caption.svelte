<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { HTMLAttributes } from 'svelte/elements';
	import { cn } from '$lib/utils';

	let {
		class: className,
		children,
		...restProps
	}: HTMLAttributes<HTMLElement> & { children?: Snippet } = $props();
</script>

<caption class={cn('text-on-surface/60 mt-4 text-sm', className)} {...restProps}>
	{@render children?.()}
</caption>
