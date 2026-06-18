<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { HTMLTableAttributes } from 'svelte/elements';
	import { cn } from '$lib/utils';

	let {
		class: className,
		children,
		...restProps
	}: HTMLTableAttributes & { children?: Snippet } = $props();
</script>

<div class="relative w-full overflow-auto">
	<table class={cn('table table-zebra w-full', className)} {...restProps}>
		{@render children?.()}
	</table>
</div>
