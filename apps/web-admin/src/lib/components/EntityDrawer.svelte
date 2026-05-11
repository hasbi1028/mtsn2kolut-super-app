<script lang="ts">
	import type { Snippet } from 'svelte';
	import { X } from '@lucide/svelte';
	import { Button } from '$lib/components/ui/button';
	import { confirmDiscardChanges } from '$lib/client/unsaved-changes';

	let {
		open = $bindable(false),
		title,
		subtitle = '',
		hasUnsavedChanges = false,
		closeLabel = 'Tutup panel',
		actions,
		children,
		onbeforeclose
	}: {
		open?: boolean;
		title: string;
		subtitle?: string;
		hasUnsavedChanges?: boolean;
		closeLabel?: string;
		actions?: Snippet;
		children?: Snippet;
		onbeforeclose?: () => boolean | Promise<boolean>;
	} = $props();

	let panelRef: HTMLElement | null = $state(null);

	async function canClose() {
		if (onbeforeclose) return await onbeforeclose();
		return confirmDiscardChanges(hasUnsavedChanges);
	}

	async function requestClose() {
		if (!(await canClose())) return;
		open = false;
	}

	function handleBackdropClick(event: MouseEvent) {
		if (event.target !== event.currentTarget) return;
		void requestClose();
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key !== 'Escape') return;
		event.preventDefault();
		void requestClose();
	}

	$effect(() => {
		if (!open || typeof document === 'undefined') return;
		const previousOverflow = document.body.style.overflow;
		document.body.style.overflow = 'hidden';
		requestAnimationFrame(() => panelRef?.focus());
		return () => {
			document.body.style.overflow = previousOverflow;
		};
	});
</script>

{#if open}
	<div
		class="fixed inset-0 z-50 flex justify-end bg-black/40 p-0 sm:p-4"
		role="dialog"
		aria-modal="true"
		aria-label={title}
		tabindex="-1"
		onclick={handleBackdropClick}
		onkeydown={handleKeydown}
	>
		<section
			bind:this={panelRef}
			class="flex h-full w-full flex-col overflow-hidden bg-background shadow-2xl outline-none sm:max-w-xl sm:rounded-md sm:border"
			tabindex="-1"
		>
			<header class="flex shrink-0 items-start justify-between gap-3 border-b px-5 py-4">
				<div class="min-w-0 space-y-1">
					<h2 class="truncate text-lg font-semibold text-foreground">{title}</h2>
					{#if subtitle}
						<p class="text-sm text-muted-foreground">{subtitle}</p>
					{/if}
				</div>
				<Button type="button" variant="ghost" size="icon" class="size-9 shrink-0" aria-label={closeLabel} onclick={() => void requestClose()}>
					<X class="size-4" />
				</Button>
			</header>
			<div class="min-h-0 flex-1 overflow-y-auto px-5 py-4">
				{@render children?.()}
			</div>
			{#if actions}
				<footer class="shrink-0 border-t bg-muted/30 px-5 py-3">
					<div class="flex flex-wrap items-center justify-end gap-2">
						{@render actions()}
					</div>
				</footer>
			{/if}
		</section>
	</div>
{/if}

