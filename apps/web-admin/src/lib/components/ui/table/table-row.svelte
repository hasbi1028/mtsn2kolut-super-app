<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { HTMLAttributes } from 'svelte/elements';
	import { cn } from '$lib/utils';

	let {
		class: className,
		children,
		...restProps
	}: HTMLAttributes<HTMLTableRowElement> & { children?: Snippet } = $props();
</script>

<tr class={cn('even:bg-base-200/50 hover:bg-base-200/80', className)} {...restProps}>
	{@render children?.()}
</tr>
