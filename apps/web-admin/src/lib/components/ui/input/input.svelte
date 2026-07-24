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
		class={cn('flex h-10 w-full rounded-lg border border-input bg-background px-3 py-2 text-sm text-foreground file:border-0 file:bg-transparent file:text-sm file:font-medium file:text-foreground placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50', className)}
		type="file"
		bind:this={ref}
		bind:files
		{...restProps}
	/>
{:else}
	<input
		class={cn('flex h-10 w-full rounded-lg border border-input bg-background px-3 py-2 text-sm text-foreground file:border-0 file:bg-transparent file:text-sm file:font-medium file:text-foreground placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50', className)}
		{type}
		bind:this={ref}
		bind:value
		{...restProps}
	/>
{/if}
