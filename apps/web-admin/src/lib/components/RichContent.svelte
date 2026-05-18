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

<svelte:element this={tag} class={`rich-content ${className}`}>
	{@html rendered}
</svelte:element>

<style>
	:global(.rich-content ol) {
		list-style: decimal;
		margin: 0.5rem 0;
		padding-left: 1.5rem;
	}

	:global(.rich-content ul) {
		list-style: disc;
		margin: 0.5rem 0;
		padding-left: 1.5rem;
	}

	:global(.rich-content li) {
		margin: 0.25rem 0;
		padding-left: 0.125rem;
	}

	:global(.rich-content li > ol),
	:global(.rich-content li > ul) {
		margin-top: 0.25rem;
		margin-bottom: 0.25rem;
	}

	:global(.rich-content ol ol) {
		list-style: lower-alpha;
	}

	:global(.rich-content ol ol ol) {
		list-style: lower-roman;
	}

	:global(.rich-content ul ul) {
		list-style: circle;
	}

	:global(.rich-content ul ul ul) {
		list-style: square;
	}
</style>
