<script lang="ts">
	import { cn } from '$lib/utils';

	let {
		ref = $bindable(null),
		class: className,
		value = $bindable(),
		type = 'text',
		files = $bindable(),
		...restProps
	}: {
		ref?: HTMLInputElement | null;
		class?: string;
		value?: unknown;
		type?: string;
		files?: FileList;
		[key: string]: unknown;
	} = $props();
</script>

{#if type === 'file'}
	<input
		class={cn('input', className)}
		type="file"
		bind:this={ref}
		bind:files
		{...restProps}
	/>
{:else}
	<input
		class={cn('input', className)}
		{type}
		bind:this={ref}
		bind:value
		{...restProps}
	/>
{/if}
