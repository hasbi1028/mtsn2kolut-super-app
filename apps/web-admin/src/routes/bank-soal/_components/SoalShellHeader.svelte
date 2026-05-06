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

<section class="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm">
	<div class="flex items-center gap-3 border-b border-slate-100 px-5 py-2.5 text-xs">
		<a href={resolve('/bank-soal')} class="text-slate-500 hover:text-slate-700">← Beranda Bank Soal</a>
		<span class="text-slate-200">|</span>
		<a href={resolve(reviewHref)} class="text-emerald-600 hover:text-emerald-800">Ruang Review Fokus →</a>
	</div>
	<div class="grid gap-4 p-5 lg:grid-cols-[minmax(0,1fr)_auto] lg:items-center">
		<div class="min-w-0">
			<p class="text-xs font-bold uppercase tracking-[0.18em] text-emerald-700">Bank Soal</p>
			<h1 class="mt-1 text-2xl font-semibold tracking-tight text-slate-900">Kelola repositori soal reusable</h1>
			<p class="mt-2 max-w-3xl text-sm leading-6 text-slate-600">
				Bank Soal berdiri sebagai repositori bersama. Kegiatan CBT hanya menjadi konteks atau filter opsional saat soal dipakai untuk paket/event tertentu.
			</p>
		</div>
		<div class="flex flex-wrap gap-2 lg:justify-end">
			<Button onclick={onCreate} class="bg-[oklch(0.38_0.13_145)] text-white hover:bg-[oklch(0.34_0.13_145)]">Buat Soal</Button>
			<LoadingButton
				variant="outline"
				onclick={onExport}
				loading={exportBusy}
				loadingLabel="Export..."
				disabled={exportBusy || totalItems === 0}
				class="bg-white text-slate-700 hover:bg-slate-50"
			>
				{exportButtonLabel}
			</LoadingButton>
		</div>
	</div>
</section>
