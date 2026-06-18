<script lang="ts">
	import { cn } from '$lib/utils';

	let {
		ref = $bindable(null),
		value = $bindable(),
		class: className,
		...restProps
	}: {
		ref?: HTMLElement | null;
		value?: unknown;
		class?: string;
		[key: string]: unknown;
	} = $props();
</script>

<textarea
	bind:this={ref}
	class={cn('textarea textarea-bordered w-full', className)}
	bind:value
	{...restProps}
></textarea>
