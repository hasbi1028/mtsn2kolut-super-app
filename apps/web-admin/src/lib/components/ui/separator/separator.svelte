<script lang="ts">
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/utils';

	let {
		ref = $bindable(null),
		orientation = 'horizontal',
		class: className,
		decorative = false,
		...restProps
	}: {
		ref?: HTMLElement | null;
		orientation?: 'horizontal' | 'vertical';
		class?: string;
		decorative?: boolean;
		[key: string]: unknown;
	} = $props();
</script>

{#if orientation === 'vertical'}
	<div
		role={decorative ? 'none' : 'separator'}
		aria-orientation="vertical"
		bind:this={ref}
		class={cn('bg-border w-px shrink-0', className)}
		{...restProps}
	></div>
{:else}
	<div
		role={decorative ? 'none' : 'separator'}
		aria-orientation="horizontal"
		bind:this={ref}
		class={cn('bg-border h-px shrink-0', className)}
		{...restProps}
	></div>
{/if}
