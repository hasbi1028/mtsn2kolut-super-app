<script lang="ts">
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';

	const userRoles = $derived(page.data.user?.roles ?? (page.data.user?.role ? [page.data.user.role] : []));
	const canAccess = $derived(userRoles.includes('admin') || userRoles.includes('guru'));

	type ResultPath = {
		title: string;
		description: string;
		href: ResultRoute;
	};

	type ResultRoute = '/cbt/events' | '/cbt/sessions' | '/cbt/soal' | '/cbt/byod' | '/cbt';

	const adminResultPaths: ResultPath[] = [
		{
			title: 'Hasil Kegiatan',
			description: 'Lihat rekap hasil dari daftar kegiatan CBT.',
			href: '/cbt/events'
		},
		{
			title: 'Hasil Sesi',
			description: 'Buka sesi ujian untuk nilai peserta dan status pengerjaan.',
			href: '/cbt/sessions'
		},
		{
			title: 'Analisis Soal',
			description: 'Pilih sesi untuk membaca analisis butir soal.',
			href: '/cbt/sessions'
		}
	];

	const guruResultPaths: ResultPath[] = [
		{
			title: 'Bank Soal & Review',
			description: 'Mulai dari Bank Soal untuk menindaklanjuti kualitas soal setelah ujian.',
			href: '/cbt/soal'
		},
		{
			title: 'Panduan Monitoring',
			description: 'Gunakan panduan BYOD saat membaca kondisi ujian yang sedang berlangsung.',
			href: '/cbt/byod'
		},
		{
			title: 'Dashboard CBT',
			description: 'Kembali ke dashboard CBT untuk memilih pekerjaan yang tersedia untuk akun guru.',
			href: '/cbt'
		}
	];

	const resultPaths = $derived(userRoles.includes('admin') ? adminResultPaths : guruResultPaths);
</script>

<svelte:head>
	<title>Hasil CBT — MTsN 2 Kolaka Utara</title>
</svelte:head>

{#if canAccess}
	<div class="mx-auto max-w-5xl space-y-6">
		<section class="space-y-4">
			<div class="space-y-2">
				<p class="text-xs font-semibold uppercase tracking-[0.18em] text-emerald-700">CBT · Hasil</p>
				<h1 class="text-3xl font-semibold tracking-tight text-slate-900">Hasil CBT</h1>
				<p class="max-w-2xl text-sm leading-6 text-slate-600">
					{#if userRoles.includes('admin')}
						Pilih pekerjaan hasil yang ingin dibuka. Rekap kegiatan, hasil sesi, dan analisis soal tetap berada di halaman detail yang sudah tersedia.
					{:else}
						Untuk guru, halaman ini menjadi pintu informasi hasil. Rekap detail dibuka melalui alur yang disediakan admin atau dari dashboard CBT.
					{/if}
				</p>
			</div>
			<Button href={resolve('/cbt')} variant="outline">Kembali ke Dashboard CBT</Button>
		</section>

		<section class="grid gap-4 lg:grid-cols-3" aria-label="Pilihan Hasil CBT">
			{#each resultPaths as path (path.title)}
				<a
					href={resolve(path.href)}
					class="group block h-full rounded-xl focus:outline-none focus:ring-2 focus:ring-emerald-600 focus:ring-offset-2"
				>
					<Card.Root
						class="h-full border-slate-200 bg-white shadow-sm transition group-hover:border-emerald-300 group-hover:shadow-md"
					>
						<Card.Header class="space-y-3 p-6">
							<Card.Title class="text-2xl text-slate-900">{path.title}</Card.Title>
							<Card.Description class="text-sm leading-6">{path.description}</Card.Description>
						</Card.Header>
					</Card.Root>
				</a>
			{/each}
		</section>
	</div>
{:else}
	<div class="flex min-h-[60vh] flex-col items-center justify-center gap-4 px-4 py-16 text-center">
		<div class="max-w-lg rounded-2xl border border-slate-200 bg-white p-8 shadow-sm">
			<h2 class="text-xl font-semibold text-slate-900">Akses terbatas</h2>
			<p class="mt-3 text-sm leading-6 text-slate-600">
				Halaman Hasil hanya tersedia untuk admin dan guru. Silakan gunakan akun guru Anda atau hubungi administrator.
			</p>
			<div class="mt-6">
				<a
					href={resolve('/cbt')}
					class="inline-flex rounded-md border border-emerald-200 bg-white px-4 py-2 text-sm font-semibold text-emerald-800 transition hover:bg-emerald-50"
				>
					Kembali ke Dashboard CBT
				</a>
			</div>
		</div>
	</div>
{/if}
