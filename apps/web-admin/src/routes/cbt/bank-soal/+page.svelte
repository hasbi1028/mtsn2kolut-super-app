<script lang="ts">
	import { resolve } from '$app/paths';
	import * as Card from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';

	type BankSoalRoute = '/cbt/soal' | '/cbt/soal/review' | '/';

	type BankSoalCard = {
		title: string;
		description: string;
		href: BankSoalRoute;
		tone: 'authoring' | 'review' | 'import';
		label: string;
	};

	const cards: BankSoalCard[] = [
		{
			title: 'Komposer Soal',
			label: 'Penyusunan',
			description: 'Tulis, susun, dan rapikan butir soal melalui jalur /cbt/soal yang aktif. Guru memulai dari sini.',
			href: '/cbt/soal',
			tone: 'authoring'
		},
		{
			title: 'Review & Penerbitan',
			label: 'Peninjauan',
			description: 'Tinjau ulang soal yang menunggu review, beri catatan, lalu terbitkan ke bank soal yang sudah siap dipakai.',
			href: '/cbt/soal/review',
			tone: 'review'
		},
		{
			title: 'Upload CSV',
			label: 'Impor',
			description: 'Impor bank soal dari file CSV dengan mode impor di Komposer Soal. Cocok untuk migrasi soal lama.',
			href: '/cbt/soal',
			tone: 'import'
		}
	];

	function toneClass(tone: BankSoalCard['tone']): string {
		switch (tone) {
			case 'authoring':
				return 'border-emerald-200 bg-emerald-50/70 text-emerald-800';
			case 'review':
				return 'border-amber-200 bg-amber-50/70 text-amber-800';
			case 'import':
				return 'border-slate-200 bg-slate-100 text-slate-700';
		}
	}
</script>

<svelte:head>
	<title>Bank Soal — MTsN 2 Kolaka Utara</title>
</svelte:head>

<div class="space-y-6">
	<section class="rounded-3xl border border-emerald-100 bg-gradient-to-br from-emerald-50 via-white to-lime-50 p-6 shadow-sm">
		<div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
			<div class="max-w-3xl space-y-3">
				<p class="text-xs font-semibold uppercase tracking-[0.24em] text-emerald-700">CBT · Bank Soal</p>
				<h1 class="text-3xl font-semibold tracking-tight text-slate-900">Bank Soal</h1>
				<p class="max-w-2xl text-sm leading-6 text-slate-600">
					Pintu masuk untuk menyusun, meninjau, dan mengimpor butir soal. Komposer Soal tetap menjadi jalur utama
					penulisan soal dari rute /cbt/soal.
				</p>
			</div>
			<div class="flex flex-wrap gap-3">
				<Button href={resolve('/cbt')} variant="outline">Beranda CBT</Button>
				<Button href={resolve('/cbt/soal')}>Komposer Soal</Button>
			</div>
		</div>
	</section>

	<section class="grid gap-4 lg:grid-cols-3" aria-label="Pilihan utama Bank Soal">
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
					{#if card.tone === 'import'}
						<Button href={resolve('/cbt/soal') + '?mode=import'} variant="outline" class="border-emerald-200 text-emerald-800 hover:bg-emerald-50">Buka Impor</Button>
					{:else}
						<Button href={resolve(card.href)} variant="outline" class="border-emerald-200 text-emerald-800 hover:bg-emerald-50">Buka</Button>
					{/if}
				</Card.Footer>
			</Card.Root>
		{/each}
	</section>

	<Card.Root class="border-emerald-200 bg-emerald-50/60 shadow-sm">
		<Card.Header>
			<Card.Title class="text-lg text-slate-900">Catatan ringkas</Card.Title>
			<Card.Description>
				Semua jalur Bank Soal kembali ke Komposer Soal di /cbt/soal. Halaman ini hanya ringkasan navigasi;
				data soal, review, dan impor tetap dikelola melalui rute Komposer Soal yang aktif.
			</Card.Description>
		</Card.Header>
	</Card.Root>
</div>
