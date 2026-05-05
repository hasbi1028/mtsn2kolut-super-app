<script lang="ts">
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import * as Card from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';

	const userRoles = $derived(page.data.user?.roles ?? (page.data.user?.role ? [page.data.user.role] : []));
	const canAccess = $derived(userRoles.includes('admin'));

	type PaketSoalRoute = '/cbt/packages' | '/cbt/kegiatan';

	type PaketSoalCard = {
		title: string;
		description: string;
		href: PaketSoalRoute;
		label: string;
		tone: 'assembly' | 'operation';
	};

	const cards: PaketSoalCard[] = [
		{
			title: 'Kelola Paket',
			label: 'Perakitan',
			description: 'Rakit paket soal dari bank soal yang sudah diterbitkan. Atur komposisi, bobot, dan aturan paket ujian.',
			href: '/cbt/packages',
			tone: 'assembly'
		},
		{
			title: 'Kegiatan & Sesi',
			label: 'Lintas Modul',
			description: 'Lanjutkan ke Kegiatan & Sesi untuk menghubungkan paket soal dengan jadwal, peserta, dan ruang ujian.',
			href: '/cbt/kegiatan',
			tone: 'operation'
		}
	];

	function toneClass(tone: PaketSoalCard['tone']): string {
		switch (tone) {
			case 'assembly':
				return 'border-lime-200 bg-lime-50/70 text-lime-800';
			case 'operation':
				return 'border-teal-200 bg-teal-50/70 text-teal-800';
		}
	}
</script>

<svelte:head>
	<title>Paket Soal — MTsN 2 Kolaka Utara</title>
</svelte:head>

{#if canAccess}
	<div class="space-y-6">
	<section class="rounded-3xl border border-emerald-100 bg-gradient-to-br from-emerald-50 via-white to-lime-50 p-6 shadow-sm">
		<div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
			<div class="max-w-3xl space-y-3">
				<p class="text-xs font-semibold uppercase tracking-[0.24em] text-emerald-700">CBT · Paket Soal</p>
				<h1 class="text-3xl font-semibold tracking-tight text-slate-900">Paket Soal</h1>
				<p class="max-w-2xl text-sm leading-6 text-slate-600">
					Rakit paket ujian dari bank soal yang sudah diterbitkan, lalu hubungkan dengan kegiatan dan sesi ujian.
				</p>
			</div>
			<div class="flex flex-wrap gap-3">
				<Button href={resolve('/cbt')} variant="outline">Beranda CBT</Button>
				<Button href={resolve('/cbt/packages')}>Kelola Paket</Button>
			</div>
		</div>
	</section>

	<section class="grid gap-4 lg:grid-cols-2" aria-label="Pilihan utama Paket Soal">
		{#each cards as card (card.title)}
			<Card.Root class="flex h-full flex-col border-slate-200 bg-white shadow-sm transition hover:-translate-y-0.5 hover:border-emerald-200 hover:shadow-md">
				<Card.Header class="space-y-3">
					<Badge class={toneClass(card.tone)} variant="outline">{card.label}</Badge>
					<div>
						<Card.Title class="text-lg text-slate-900">{card.title}</Card.Title>
						<Card.Description class="mt-2 leading-6">{card.description}</Card.Description>
					</div>
				</Card.Header>
				<Card.Footer>
					<Button href={resolve(card.href)} variant="outline" class="border-emerald-200 text-emerald-800 hover:bg-emerald-50">Buka</Button>
				</Card.Footer>
			</Card.Root>
		{/each}
	</section>

	<Card.Root class="border-emerald-200 bg-emerald-50/60 shadow-sm">
		<Card.Header>
			<Card.Title class="text-lg text-slate-900">Rujukan silang</Card.Title>
			<Card.Description>
				Paket Soal terhubung erat dengan Kegiatan & Sesi. Gunakan tautan silang di atas untuk berpindah antar modul
				tanpa kembali ke Beranda CBT.
			</Card.Description>
		</Card.Header>
	</Card.Root>
	</div>
{:else}
	<div class="flex min-h-[60vh] flex-col items-center justify-center gap-4 px-4 py-16 text-center">
		<div class="rounded-2xl border border-slate-200 bg-white p-8 shadow-sm max-w-lg">
			<h2 class="text-xl font-semibold text-slate-900">Akses terbatas</h2>
			<p class="mt-3 text-sm leading-6 text-slate-600">
				Halaman Paket Soal hanya tersedia untuk admin. Silakan hubungi administrator jika Anda memerlukan akses.
			</p>
			<div class="mt-6">
				<a href={resolve('/cbt')} class="inline-flex rounded-md border border-emerald-200 bg-white px-4 py-2 text-sm font-semibold text-emerald-800 transition hover:bg-emerald-50">Kembali ke Beranda CBT</a>
			</div>
		</div>
	</div>
{/if}
