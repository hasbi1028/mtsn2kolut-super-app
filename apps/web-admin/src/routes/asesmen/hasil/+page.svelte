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

	type ResultRoute = '/asesmen/kegiatan' | '/asesmen/sesi' | '/asesmen/pelaksanaan' | '/asesmen/aplikasi-siswa' | '/asesmen';

	const adminResultPaths: ResultPath[] = [
		{
			title: 'Hasil Kegiatan',
			description: 'Lihat rekap hasil dari daftar kegiatan ujian.',
			href: '/asesmen/kegiatan'
		},
		{
			title: 'Hasil Sesi',
			description: 'Buka sesi ujian untuk nilai peserta dan status pengerjaan.',
			href: '/asesmen/sesi'
		},
		{
			title: 'Analisis Soal',
			description: 'Pilih sesi untuk membaca analisis butir soal.',
			href: '/asesmen/sesi'
		}
	];

	const guruResultPaths: ResultPath[] = [
		{
			title: 'Pantau Pelaksanaan',
			description: 'Buka ruang pelaksanaan bila hasil perlu dicocokkan dengan status sesi berjalan.',
			href: '/asesmen/pelaksanaan'
		},
		{
			title: 'Panduan Pemantauan',
			description: 'Gunakan panduan BYOD saat membaca kondisi ujian yang sedang berlangsung.',
			href: '/asesmen/aplikasi-siswa'
		},
		{
			title: 'Beranda Ujian',
			description: 'Kembali ke beranda ujian untuk memilih pekerjaan yang tersedia untuk akun guru.',
			href: '/asesmen'
		}
	];

	const resultPaths = $derived(userRoles.includes('admin') ? adminResultPaths : guruResultPaths);
</script>

<svelte:head>
	<title>Hasil Asesmen CBT — MTsN 2 Kolaka Utara</title>
</svelte:head>

{#if canAccess}
	<div class="mx-auto max-w-5xl space-y-6">
		<section class="space-y-4">
			<div class="space-y-2">
				<p class="text-xs font-semibold uppercase tracking-[0.18em] text-primary">Asesmen CBT · Hasil</p>
				<h1 class="text-3xl font-semibold tracking-tight text-foreground">Hasil Asesmen CBT</h1>
				<p class="max-w-2xl text-sm leading-6 text-muted-foreground">
					{#if userRoles.includes('admin')}
						Pilih pekerjaan hasil yang ingin dibuka. Rekap kegiatan, hasil sesi, dan analisis soal tetap berada di halaman detail yang sudah tersedia.
					{:else}
						Untuk guru, halaman ini menjadi pintu informasi hasil. Rekap detail dibuka melalui alur yang disediakan admin atau dari Beranda Asesmen CBT.
					{/if}
				</p>
			</div>
			<Button href={resolve('/asesmen')} variant="outline">Kembali ke Beranda Asesmen CBT</Button>
		</section>

		<section class="grid gap-4 lg:grid-cols-3" aria-label="Pilihan Hasil CBT">
			{#each resultPaths as path (path.title)}
				<a
					href={resolve(path.href)}
					class="group block h-full rounded-xl focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2"
				>
					<Card.Root
						class="h-full border-border bg-card shadow-sm transition group-hover:border-primary/20 group-hover:shadow-md"
					>
						<Card.Header class="space-y-3 p-6">
							<Card.Title class="text-2xl text-foreground">{path.title}</Card.Title>
							<Card.Description class="text-sm leading-6">{path.description}</Card.Description>
						</Card.Header>
					</Card.Root>
				</a>
			{/each}
		</section>
	</div>
{:else}
	<div class="flex min-h-[60vh] flex-col items-center justify-center gap-4 px-4 py-16 text-center">
		<div class="max-w-lg rounded-2xl border border-border bg-card p-8 shadow-sm">
			<h2 class="text-xl font-semibold text-foreground">Akses terbatas</h2>
			<p class="mt-3 text-sm leading-6 text-muted-foreground">
				Halaman Hasil hanya tersedia untuk admin dan guru. Silakan gunakan akun guru Anda atau hubungi administrator.
			</p>
			<div class="mt-6">
				<a
					href={resolve('/asesmen')}
					class="inline-flex rounded-md border border-primary/20 bg-card px-4 py-2 text-sm font-semibold text-primary transition hover:bg-primary/10"
				>
					Kembali ke Beranda Ujian
				</a>
			</div>
		</div>
	</div>
{/if}
