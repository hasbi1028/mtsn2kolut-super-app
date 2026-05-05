<script lang="ts">
	import { resolve } from '$app/paths';
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';

	type ResultTarget = 'events' | 'sessions';

	type ResultPath = {
		title: string;
		description: string;
		target: ResultTarget;
		cta: string;
		note: string;
	};

	const resultPaths: ResultPath[] = [
		{
			title: 'Hasil per Kegiatan',
			description: 'Mulai dari daftar kegiatan untuk membuka detail kegiatan dan membaca rekap nilai gabungan seluruh sesi.',
			target: 'events',
			cta: 'Buka Kegiatan',
			note: 'Gunakan tab Hasil di detail kegiatan untuk rekap lintas sesi dan ekspor yang tersedia.'
		},
		{
			title: 'Hasil per Sesi',
			description: 'Pilih sesi ujian untuk melihat nilai peserta, status pengerjaan, dan kebutuhan tindak lanjut per ruang atau kelas.',
			target: 'sessions',
			cta: 'Buka Sesi',
			note: 'Gunakan tab Hasil di detail sesi untuk daftar nilai peserta pada sesi tersebut.'
		},
		{
			title: 'Analisis Soal',
			description: 'Analisis butir dibaca dari konteks sesi agar kualitas soal tetap terhubung dengan paket dan peserta yang mengerjakan.',
			target: 'sessions',
			cta: 'Pilih Sesi',
			note: 'Buka detail sesi lalu gunakan tab Butir untuk matriks analisis soal; tab Uraian dipakai untuk koreksi esai.'
		}
	];
</script>

<svelte:head>
	<title>Hasil & Analisis CBT — MTsN 2 Kolaka Utara</title>
</svelte:head>

<div class="space-y-6">
	<section class="rounded-3xl border border-emerald-100 bg-gradient-to-br from-emerald-50 via-white to-lime-50 p-6 shadow-sm">
		<div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
			<div class="max-w-3xl space-y-3">
				<p class="text-xs font-semibold uppercase tracking-[0.24em] text-emerald-700">CBT · Hasil & Analisis</p>
				<h1 class="text-3xl font-semibold tracking-tight text-slate-900">Hasil & Analisis</h1>
				<p class="max-w-2xl text-sm leading-6 text-slate-600">
					Halaman ini adalah pintu masuk sederhana untuk membaca hasil CBT. Data rinci masih berada di detail kegiatan dan
					detail sesi sampai endpoint agregat tersedia, sehingga nilai, analisis soal, dan koreksi tetap jujur pada konteks ujian.
				</p>
			</div>
			<div class="flex flex-wrap gap-3">
				<Button href={resolve('/cbt')} variant="outline">Beranda CBT</Button>
				<Button href="/cbt/events">Hasil per Kegiatan</Button>
				<Button href="/cbt/sessions" variant="outline">Hasil per Sesi</Button>
			</div>
		</div>
	</section>

	<section class="grid gap-4 lg:grid-cols-3" aria-label="Pilihan utama Hasil & Analisis CBT">
		{#each resultPaths as path (path.title)}
			<Card.Root class="flex h-full flex-col border-slate-200 bg-white shadow-sm">
				<Card.Header>
					<Card.Title class="text-lg text-slate-900">{path.title}</Card.Title>
					<Card.Description class="leading-6">{path.description}</Card.Description>
				</Card.Header>
				<Card.Content class="flex flex-1 flex-col justify-between gap-5">
					<p class="rounded-2xl border border-emerald-100 bg-emerald-50/70 p-3 text-sm leading-6 text-emerald-900">
						{path.note}
					</p>
					{#if path.target === 'events'}
						<a href={resolve('/cbt/events')} class="inline-flex rounded-md border border-emerald-200 bg-white px-3 py-2 text-sm font-semibold text-emerald-800 transition hover:bg-emerald-50">
							{path.cta}
						</a>
					{:else}
						<a href={resolve('/cbt/sessions')} class="inline-flex rounded-md border border-emerald-200 bg-white px-3 py-2 text-sm font-semibold text-emerald-800 transition hover:bg-emerald-50">
							{path.cta}
						</a>
					{/if}
				</Card.Content>
			</Card.Root>
		{/each}
	</section>

	<Card.Root class="border-emerald-200 bg-emerald-50/60 shadow-sm">
		<Card.Header>
			<Card.Title class="text-lg text-slate-900">Catatan sementara</Card.Title>
			<Card.Description>
				Belum ada halaman agregat khusus untuk seluruh hasil dan analisis. Untuk saat ini, gunakan detail kegiatan untuk rekap
				lintas sesi, lalu gunakan detail sesi untuk tab Hasil, Butir, dan Uraian.
			</Card.Description>
		</Card.Header>
	</Card.Root>
</div>
