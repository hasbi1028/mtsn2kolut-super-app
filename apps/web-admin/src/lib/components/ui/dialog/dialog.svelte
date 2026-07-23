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
	<!-- Backdrop + modal with explicit Tailwind classes -->
	<div
		class={cn(
			'fixed inset-0 z-[999] flex items-end sm:items-center justify-center bg-black/60 p-0 sm:p-6',
			className
		)}
		onclick={handleBackdropClick}
		role="dialog"
		aria-modal="true"
		{...restProps}
	>
		<div class="w-full sm:max-w-lg bg-base-100 border border-base-300 rounded-t-2xl sm:rounded-xl shadow-2xl ring-1 ring-black/5 max-h-[90dvh] overflow-y-auto">
			<div class="flex justify-center pt-1.5 pb-0 sm:hidden">
				<div class="h-1 w-10 rounded-full bg-base-300/70"></div>
			</div>
			{@render children?.()}
		</div>
	</div>
{/if}
