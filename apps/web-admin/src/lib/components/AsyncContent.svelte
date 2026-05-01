<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		promise,
		pending,
		failed: failedSnippet,
		children,
		onerror,
	}: {
		promise: Promise<unknown> | null;
		pending?: Snippet;
		failed?: Snippet<[error: unknown, reset: (() => void) | undefined]>;
		children: Snippet<[value: unknown]>;
		onerror?: (error: unknown, reset: () => void) => void;
	} = $props();
</script>

{#if promise}
	{#await promise}
		{@render pending?.()}
	{:then value}
		<svelte:boundary {onerror}>
			{@render children(value)}
			{#snippet failed(error, reset)}
				{@render failedSnippet?.(error, reset)}
			{/snippet}
		</svelte:boundary>
	{:catch error}
		{@render failedSnippet?.(error, undefined)}
	{/await}
{:else}
	{@render pending?.()}
{/if}
