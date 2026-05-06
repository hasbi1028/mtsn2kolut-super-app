<script lang="ts">
	import { resolve } from '$app/paths';
	import { Button } from '$lib/components/ui/button';
	import LoadingButton from '$lib/components/LoadingButton.svelte';

	let {
		reviewHref,
		membersHref,
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
		membersHref: `/asesmen/kegiatan/${string}/members` | '/asesmen/kegiatan';
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

<section class="rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
	<div class="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
		<div class="min-w-0">
			<p class="text-xs font-bold uppercase tracking-wider text-slate-600">Pintasan Bank Soal</p>
			<p class="mt-1 text-xs text-slate-500">Aksi penting untuk mengisi, meninjau, dan memakai repositori soal reusable.</p>
		</div>
		<div class="flex flex-wrap gap-2">
			<Button onclick={onCreate} class="bg-[oklch(0.38_0.13_145)] text-white hover:bg-[oklch(0.34_0.13_145)]">Buat Soal</Button>
			<Button variant="outline" onclick={onImport} class="bg-white">Upload CSV</Button>
			<Button variant="outline" onclick={onRevision} class="bg-white text-red-700 hover:bg-red-50">Perlu Revisi</Button>
			<Button variant="outline" onclick={onPendingReviews} disabled={!canUseReviewerTools} class="bg-white text-amber-800 hover:bg-amber-50 disabled:opacity-50">Antrian Review</Button>
			<a href={resolve(reviewHref)} class="inline-flex rounded-md border border-input bg-white px-3 py-2 text-sm font-medium text-slate-700 hover:bg-slate-50">Review Fokus</a>
			{#if isAdminRole}
				<a href={resolve(membersHref)} class="inline-flex rounded-md border border-input bg-white px-3 py-2 text-sm font-medium text-slate-700 hover:bg-slate-50">Penugasan Event</a>
				<Button variant="outline" onclick={onApproved} class="bg-white text-green-800 hover:bg-green-50">Siap Terbit</Button>
			{/if}
			<LoadingButton variant="outline" onclick={onExport} loading={exportBusy} loadingLabel="Export..." disabled={exportBusy || totalItems === 0} class="bg-white">Export</LoadingButton>
		</div>
	</div>
</section>
