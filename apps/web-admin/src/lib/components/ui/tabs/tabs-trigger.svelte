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

<button
	class={cn('tab', className)}
	role="tab"
	{...restProps}
>
	{@render children?.()}
</button>
