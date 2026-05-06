<script lang="ts">
	import katex from 'katex';

	let { src = '', display = true, class: cls = '' }: { src: string; display?: boolean; class?: string } = $props();

	let rendered = $derived.by(() => {
		if (!src.trim()) return '';
		try {
			return katex.renderToString(src.trim(), {
				displayMode: display,
				throwOnError: false,
				output: 'html',
				trust: false,
			});
		} catch {
			return `<span class="text-destructive font-mono text-xs">${src}</span>`;
		}
	});
</script>

{#if src.trim()}
	<div class="katex-wrapper overflow-x-auto {cls}">{@html rendered}</div>
{/if}
