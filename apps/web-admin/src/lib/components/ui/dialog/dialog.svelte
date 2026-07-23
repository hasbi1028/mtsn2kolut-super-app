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
	<!-- Backdrop + modal with explicit Tailwind classes (no DaisyUI dependency) -->
	<div
		class={cn(
			'fixed inset-0 z-[999] flex items-end sm:items-center justify-center bg-black/40 p-0 sm:p-4',
			className
		)}
		onclick={handleBackdropClick}
		role="dialog"
		aria-modal="true"
		{...restProps}
	>
		<div class="w-full sm:max-w-lg bg-white dark:bg-neutral-900 rounded-t-2xl sm:rounded-2xl shadow-2xl max-h-[90dvh] overflow-y-auto">
			{@render children?.()}
		</div>
	</div>
{/if}
