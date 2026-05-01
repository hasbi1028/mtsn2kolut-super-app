<script lang="ts">
	import AsyncContent from '$lib/components/AsyncContent.svelte';

	let {
		promise,
		shouldThrow = false,
		onerror,
	}: {
		promise: Promise<unknown> | null;
		shouldThrow?: boolean;
		onerror?: (error: unknown, reset: () => void) => void;
	} = $props();

	function renderValue(value: unknown) {
		if (shouldThrow) {
			throw new Error('render failed');
		}
		return String(value);
	}

	function message(error: unknown) {
		return error instanceof Error ? error.message : String(error);
	}
</script>

<AsyncContent {promise} {onerror}>
	{#snippet pending()}
		<p>Memuat data...</p>
	{/snippet}

	{#snippet failed(error, reset)}
		<div role="alert">
			<p>{message(error)}</p>
			{#if reset}
				<button type="button" onclick={reset}>Reset</button>
			{/if}
		</div>
	{/snippet}

	{#snippet children(value)}
		<p>Nilai: {renderValue(value)}</p>
	{/snippet}
</AsyncContent>
