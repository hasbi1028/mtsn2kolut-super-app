<script lang="ts">
	import { Button } from '$lib/components/ui/button';

	type Props = {
		title?: string;
		message: string;
		actionLabel?: string;
		onRetry?: (() => void | Promise<void>) | undefined;
		compact?: boolean;
	};

	let {
		title = 'Terjadi Kendala',
		message,
		actionLabel = 'Coba Lagi',
		onRetry,
		compact = false,
	}: Props = $props();
</script>

<div class={`rounded-2xl border border-destructive/30 bg-destructive/10 ${compact ? 'px-4 py-4' : 'px-5 py-5'}`}>
	<div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
		<div class="space-y-1">
			<p class="text-sm font-semibold text-destructive">{title}</p>
			<p class="text-sm leading-6 text-destructive">{message}</p>
		</div>
		{#if onRetry}
			<Button variant="outline" size="sm" onclick={onRetry} class="border-destructive/30 bg-card text-destructive hover:bg-destructive/15">
				{actionLabel}
			</Button>
		{/if}
	</div>
</div>
