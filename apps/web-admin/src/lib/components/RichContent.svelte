<script lang="ts">
	import 'katex/dist/katex.min.css';
	import { renderRichMathHtml } from '$lib/utils/render-rich-math';

	type Props = {
		html?: string | null;
		tag?: 'div' | 'span';
		class?: string;
	};

	let { html = '', tag = 'div', class: className = '' }: Props = $props();

	let rendered = $derived(renderRichMathHtml(html ?? ''));
</script>

<svelte:element this={tag} class={className}>
	{@html rendered}
</svelte:element>
