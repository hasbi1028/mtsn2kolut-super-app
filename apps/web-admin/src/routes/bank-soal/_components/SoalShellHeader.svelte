<script lang="ts">
	import { resolve } from '$app/paths';
	import { Button } from '$lib/components/ui/button';
	import LoadingButton from '$lib/components/LoadingButton.svelte';

	let {
		reviewHref,
		exportButtonLabel,
		exportBusy,
		totalItems,
		onCreate,
		onExport
	}: {
		reviewHref: '/bank-soal/verifikasi' | `/bank-soal/verifikasi?${string}`;
		exportButtonLabel: string;
		exportBusy: boolean;
		totalItems: number;
		onCreate: () => void;
		onExport: () => void;
	} = $props();
</script>

<section class="overflow-hidden rounded-2xl border border-border bg-card shadow-sm">
	<div class="flex items-center gap-3 border-b border-border px-5 py-2.5 text-xs">
		<a href={resolve('/bank-soal')} class="text-muted-foreground hover:text-foreground">← Beranda Bank Soal</a>
		<span class="text-muted-foreground">|</span>
		<a href={resolve(reviewHref)} class="text-primary hover:text-primary">Ruang Review Fokus →</a>
	</div>
	<div class="grid gap-4 p-5 lg:grid-cols-[minmax(0,1fr)_auto] lg:items-center">
		<div class="min-w-0">
			<p class="text-xs font-bold uppercase tracking-[0.18em] text-primary">Bank Soal</p>
			<h1 class="mt-1 text-2xl font-semibold tracking-tight text-foreground">Kelola bank soal pakai ulang</h1>
			<p class="mt-2 max-w-3xl text-sm leading-6 text-muted-foreground">
				Bank Soal berdiri sebagai repositori bersama. Kegiatan CBT hanya menjadi konteks atau filter opsional saat soal dipakai untuk paket/event tertentu.
			</p>
		</div>
		<div class="flex flex-wrap gap-2 lg:justify-end">
			<Button onclick={onCreate} class="bg-primary text-primary-foreground hover:bg-primary/90">Buat Soal</Button>
			<LoadingButton
				variant="outline"
				onclick={onExport}
				loading={exportBusy}
				loadingLabel="Export..."
				disabled={exportBusy || totalItems === 0}
				class="bg-card text-foreground hover:bg-muted/50"
			>
				{exportButtonLabel}
			</LoadingButton>
		</div>
	</div>
</section>
