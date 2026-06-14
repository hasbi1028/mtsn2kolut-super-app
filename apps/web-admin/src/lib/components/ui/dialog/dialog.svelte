<script lang="ts">
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/utils';

	let {
		open = $bindable(false),
		class: className,
		children,
		...restProps
	}: {
		open?: boolean;
		class?: string;
		children?: Snippet;
		[key: string]: unknown;
	} = $props();

	function close() {
		open = false;
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') close();
	}

	function handleBackdropClick(e: MouseEvent) {
		if (e.target === e.currentTarget) close();
	}
</script>

<svelte:window onkeydown={handleKeydown} />

{#if open}
	<div
		class={cn('fixed inset-0 z-50 flex items-center justify-center bg-black/50', className)}
		onclick={handleBackdropClick}
		role="dialog"
		aria-modal="true"
		{...restProps}
	>
		{@render children?.()}
	</div>
{/if}
