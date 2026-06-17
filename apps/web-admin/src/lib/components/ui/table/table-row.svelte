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

<tr class={cn('border-b transition-colors hover:bg-muted data-[state=selected]:bg-muted', className)} {...restProps}>
	{@render children?.()}
</tr>
