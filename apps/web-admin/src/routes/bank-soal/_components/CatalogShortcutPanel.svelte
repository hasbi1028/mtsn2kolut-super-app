<script lang="ts">
	import { resolve } from '$app/paths';
	import { Button } from '$lib/components/ui/button';
	import LoadingButton from '$lib/components/LoadingButton.svelte';

	let {
		reviewHref,
		canUseReviewerTools,
		isAdminRole,
		exportBusy,
		totalItems,
		onCreate,
		onImport,
		onRevision,
		onPendingReviews,
		onApproved,
		onExport
	}: {
		reviewHref: '/bank-soal/verifikasi' | `/bank-soal/verifikasi?${string}`;
		canUseReviewerTools: boolean;
		isAdminRole: boolean;
		exportBusy: boolean;
		totalItems: number;
		onCreate: () => void;
		onImport: () => void;
		onRevision: () => void;
		onPendingReviews: () => void;
		onApproved: () => void;
		onExport: () => void;
	} = $props();
</script>

<section class="rounded-xl border border-border bg-card p-4 shadow-sm">
	<div class="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
		<div class="min-w-0">
			<p class="text-xs font-bold uppercase tracking-wider text-muted-foreground">Pintasan Bank Soal</p>
			<p class="mt-1 text-xs text-muted-foreground">Aksi penting untuk mengisi, meninjau, dan memakai bank soal pakai ulang.</p>
		</div>
		<div class="flex flex-wrap gap-2">
			<Button onclick={onCreate} class="bg-primary text-primary-foreground hover:bg-primary/90">Buat Soal</Button>
			<Button variant="outline" onclick={onImport} class="bg-card">Unggah CSV</Button>
			<Button variant="outline" onclick={onRevision} class="bg-card text-destructive hover:bg-destructive/10">Perlu Revisi</Button>
			<Button variant="outline" onclick={onPendingReviews} disabled={!canUseReviewerTools} class="bg-card text-warning hover:bg-warning/10 disabled:opacity-50">Antrean Verifikasi</Button>
			<a href={resolve(reviewHref)} class="inline-flex rounded-md border border-input bg-card px-3 py-2 text-sm font-medium text-foreground hover:bg-muted/50">Verifikasi Fokus</a>
			{#if isAdminRole}
				<Button variant="outline" onclick={onApproved} class="bg-card text-success hover:bg-success/10">Siap Terbit</Button>
			{/if}
			<LoadingButton variant="outline" onclick={onExport} loading={exportBusy} loadingLabel="Export..." disabled={exportBusy || totalItems === 0} class="bg-card">Export</LoadingButton>
		</div>
	</div>
</section>
