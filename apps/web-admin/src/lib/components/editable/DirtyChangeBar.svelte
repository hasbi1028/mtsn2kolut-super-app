<script lang="ts">
	import { Check, Loader2, RotateCcw } from '@lucide/svelte';
	import { Button } from '$lib/components/ui/button';

	let {
		count = 0,
		saving = false,
		disabled = false,
		saveLabel = 'Simpan',
		discardLabel = 'Batalkan',
		ondiscard,
		onsave
	}: {
		count?: number;
		saving?: boolean;
		disabled?: boolean;
		saveLabel?: string;
		discardLabel?: string;
		ondiscard?: () => void;
		onsave?: () => void;
	} = $props();

	const visible = $derived(count > 0);
	const changeLabel = $derived(`${count} perubahan belum disimpan`);
</script>

{#if visible}
	<div class="sticky bottom-4 z-30 rounded-md border border-primary/20 bg-background/95 p-3 shadow-lg backdrop-blur">
		<div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
			<div class="space-y-0.5">
				<p class="text-sm font-medium text-foreground">{changeLabel}</p>
				<p class="text-xs text-muted-foreground">Periksa perubahan sebelum menyimpan ke sistem akademik.</p>
			</div>
			<div class="flex flex-wrap items-center gap-2">
				<Button type="button" variant="outline" disabled={saving || disabled} onclick={() => ondiscard?.()}>
					<RotateCcw class="mr-2 size-4" />
					{discardLabel}
				</Button>
				<Button type="button" disabled={saving || disabled} onclick={() => onsave?.()}>
					{#if saving}
						<Loader2 class="mr-2 size-4 animate-spin" />
					{:else}
						<Check class="mr-2 size-4" />
					{/if}
					{saveLabel}
				</Button>
			</div>
		</div>
	</div>
{/if}

