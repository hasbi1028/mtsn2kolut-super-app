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
	- max-h-[85vh] + overflow-y-auto → kalau isi panjang (mis. jadwal 7 hari),
	  scroll di dalam card, body tetap stabil.
	- my-8 → jarak atas/bawah minimum, tengah viewport.
	- shadow-2xl → pop jelas di atas backdrop.
	- relative → anchor untuk nested popover/tooltip bila ada.
-->
<div
	role="document"
	onclick={stop}
	class={cn(
				'relative z-[101] w-full max-w-lg rounded-lg border border-border bg-card p-6 shadow-2xl my-8 max-h-[85vh] overflow-y-auto',
		className
	)}
	{...restProps}
>
	{@render children?.()}
</div>
