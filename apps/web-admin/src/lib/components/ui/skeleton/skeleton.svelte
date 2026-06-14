<script lang="ts">
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/utils';

	let {
		class: className,
		children,
		rounded = 'md',
		...restProps
	}: {
		class?: string;
		children?: Snippet;
		rounded?: string;
		[key: string]: unknown;
	} = $props();
</script>

<div
	class={cn('placeholder animate-pulse', className)}
	style={rounded ? `border-radius: var(--radius-${rounded})` : undefined}
	{...restProps}
>
	{@render children?.()}
</div>
