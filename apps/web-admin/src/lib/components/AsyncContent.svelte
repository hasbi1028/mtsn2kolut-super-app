<script lang="ts">
	import type { Snippet } from 'svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';

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

	function errorMessage(error: unknown) {
		return error instanceof Error ? error.message : String(error);
	}
</script>

{#if promise}
	{#await promise}
		{@render pending?.()}
	{:then value}
		<svelte:boundary {onerror}>
			{@render children(value)}
			{#snippet failed(error, reset)}
				{#if failedSnippet}
					{@render failedSnippet(error, reset)}
				{:else}
					<RecoveryPanel compact message={errorMessage(error)} onRetry={reset} />
				{/if}
			{/snippet}
		</svelte:boundary>
	{:catch error}
		{#if failedSnippet}
			{@render failedSnippet(error, undefined)}
		{:else}
			<RecoveryPanel compact message={errorMessage(error)} />
		{/if}
	{/await}
{:else}
	{@render pending?.()}
{/if}
