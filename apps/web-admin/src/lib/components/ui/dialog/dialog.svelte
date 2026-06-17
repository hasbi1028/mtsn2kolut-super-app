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
	<!--
		Backdrop: solid + solid black, high z-index agar selalu di atas sidebar/header/sticky
		toolbar. Pakai flex centering untuk child (Dialog.Content). Klik di backdrop tutup.
	-->
	<div
		class={cn(
			'fixed inset-0 z-[100] flex items-center justify-center overflow-y-auto bg-black/60 px-4 py-8 backdrop-blur-sm',
			className
		)}
		onclick={handleBackdropClick}
		role="dialog"
		aria-modal="true"
		{...restProps}
	>
		{@render children?.()}
	</div>
{/if}
