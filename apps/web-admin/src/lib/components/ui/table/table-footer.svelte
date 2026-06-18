<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { HTMLAttributes } from 'svelte/elements';
	import { cn } from '$lib/utils';

	let {
		class: className,
		children,
		...restProps
	}: HTMLAttributes<HTMLTableSectionElement> & { children?: Snippet } = $props();
</script>

<tfoot class={cn('border-t bg-base-200 font-medium [&>tr]:last:border-b-0', className)} {...restProps}>
	{@render children?.()}
</tfoot>
