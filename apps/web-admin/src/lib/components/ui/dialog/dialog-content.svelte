<script lang="ts">
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/utils';

	let {
		class: className,
		children,
		...restProps
	}: {
		class?: string;
		children?: Snippet;
		[key: string]: unknown;
	} = $props();

	// Stop click propagation: kalau user klik di dalam content, jangan trigger
	// handleBackdropClick di parent (yang akan menutup dialog).
	function stop(e: MouseEvent) {
		e.stopPropagation();
	}
</script>

<!--
	Content:
	- p-5 → padding konsisten di semua ukuran
	- relative → anchor untuk nested popover/tooltip bila ada.
-->
<div
	role="document"
	onclick={stop}
	class={cn(
		'p-5',
		className
	)}
	{...restProps}
>
	{@render children?.()}
</div>
