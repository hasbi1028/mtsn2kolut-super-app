<script lang="ts">
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import * as Card from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';

	type PersiapanRoute =
		| '/asesmen/paket'
		| '/asesmen/paket/new'
		| '/asesmen/kegiatan'
		| '/asesmen/kegiatan/new'
		| '/asesmen/sesi'
		| '/asesmen/sesi/new'
		| '/asesmen/pelaksanaan'
		| '/asesmen/hasil'
		| '/asesmen';

	type PreparationTask = {
		step: string;
		title: string;
		description: string;
		href: PersiapanRoute;
		cta: string;
	};

	const userRoles = $derived(page.data.user?.roles ?? (page.data.user?.role ? [page.data.user.role] : []));
	const canAccess = $derived(userRoles.includes('admin') || userRoles.includes('guru'));

	const adminTasks: PreparationTask[] = [
		{
			step: '01',
			title: 'Manajemen Paket',
			description: 'Pilih soal siap pakai, atur komposisi, dan siapkan Paket Soal untuk Kegiatan Asesmen.',
			href: '/asesmen/paket',
			cta: 'Kelola Paket'
		},
		{
			step: '02',
			title: 'Buat Kegiatan Asesmen',
			description: 'Daftarkan Kegiatan Asesmen agar sesi, peserta, dan kartu ujian punya konteks yang jelas.',
			href: '/asesmen/kegiatan/new',
			cta: 'Buat Kegiatan'
		},
		{
			step: '03',
			title: 'Atur Sesi/Token Ujian',
			description: 'Tetapkan jadwal, ruang, peserta, dan token ujian sebelum ujian masuk hari pelaksanaan.',
			href: '/asesmen/sesi/new',
			cta: 'Atur Sesi'
		},
		{
			step: '04',
			title: 'Lanjut Pelaksanaan',
			description: 'Setelah paket, kegiatan, dan sesi siap, masuk ke ruang pemantauan hari-H.',
			href: '/asesmen/pelaksanaan',
			cta: 'Ke Pelaksanaan'
		}
	];

	const guruTasks: PreparationTask[] = [
		{
			step: '01',
			title: 'Pantau Pelaksanaan',
			description: 'Jika ditugaskan pada hari-H, masuk ke ruang pelaksanaan untuk melihat status ujian.',
			href: '/asesmen/pelaksanaan',
			cta: 'Ke Pelaksanaan'
		},
		{
			step: '02',
			title: 'Lihat Hasil',
			description: 'Masuk ke pintu hasil asesmen yang tersedia untuk akun guru.',
			href: '/asesmen/hasil',
			cta: 'Buka Hasil'
		}
	];

	const tasks = $derived(userRoles.includes('admin') ? adminTasks : guruTasks);
</script>

<svelte:head>
	<title>Persiapan Asesmen CBT — MTsN 2 Kolaka Utara</title>
</svelte:head>

{#if canAccess}
	<div class="space-y-6">
	<section class="rounded-3xl border border-primary/20 bg-gradient-to-br from-primary/10 via-card to-primary/10 p-6 shadow-sm">
		<div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
			<div class="max-w-3xl space-y-3">
				<div class="flex flex-wrap items-center gap-2">
					<Badge class="border-primary/20 bg-card text-primary" variant="outline">Asesmen · Fase Persiapan</Badge>
					<Badge class="border-border bg-card text-muted-foreground" variant="outline">Tanpa perubahan data</Badge>
				</div>
				<h1 class="text-3xl font-semibold tracking-tight text-foreground md:text-4xl">Persiapan Asesmen CBT</h1>
				<p class="max-w-2xl text-sm leading-6 text-muted-foreground md:text-base">
					Mulai dari pekerjaan operasional sebelum hari ujian: paket, kegiatan, sesi, token ujian, lalu pelaksanaan.
					Penyusunan soal berada di modul Bank Soal, sementara CBT memakai soal terbit untuk paket dan sesi ujian.
				</p>
			</div>
			<div class="flex flex-wrap gap-3">
				<Button href={resolve('/asesmen')} variant="outline">Beranda Asesmen</Button>
				<Button href={resolve('/asesmen/pelaksanaan')}>Ke Pelaksanaan</Button>
			</div>
		</div>
	</section>

	<section aria-labelledby="persiapan-tasks-title" class="space-y-4">
		<div class="flex flex-wrap items-end justify-between gap-3">
			<div>
				<p class="text-xs font-semibold uppercase tracking-[0.22em] text-primary">Daftar tugas</p>
				<h2 id="persiapan-tasks-title" class="mt-1 text-2xl font-semibold tracking-tight text-foreground">Selesaikan berurutan</h2>
			</div>
			<p class="max-w-lg text-sm leading-6 text-muted-foreground">Kartu ini menjaga operator tetap fokus pada jalur persiapan asesmen tanpa masuk ke penyusunan soal.</p>
		</div>

		<div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
			{#each tasks as task (task.href)}
				<Card.Root class="flex h-full flex-col border-border bg-card shadow-sm transition hover:-translate-y-0.5 hover:border-primary/20 hover:shadow-md">
					<Card.Header class="space-y-4">
						<span class="flex h-12 w-12 items-center justify-center rounded-2xl border border-primary/20 bg-primary/10 text-lg font-semibold text-primary">{task.step}</span>
						<div>
							<Card.Title class="text-xl text-foreground">{task.title}</Card.Title>
							<Card.Description class="mt-2 leading-6">{task.description}</Card.Description>
						</div>
					</Card.Header>
					<Card.Footer class="mt-auto">
						<Button href={resolve(task.href)} variant="outline" class="w-full border-primary/20 text-primary hover:bg-primary/10">{task.cta}</Button>
					</Card.Footer>
				</Card.Root>
			{/each}
		</div>
	</section>

	<Card.Root class="border-primary/20 bg-primary/10 shadow-sm">
		<Card.Header>
			<Card.Title class="text-lg text-foreground">Prinsip fase persiapan</Card.Title>
			<Card.Description>
				{#if userRoles.includes('admin')}
					Bank Soal berdiri sebagai modul terpisah. Di sini fokuskan pekerjaan pada paket, kegiatan, sesi,
					ruang, peserta, token ujian, dan kesiapan masuk hari-H.
				{:else}
					Untuk guru, penyusunan dan verifikasi soal ada di modul Bank Soal. Halaman ini dipakai untuk membaca paket,
					hasil, dan akses pelaksanaan jika ditugaskan.
				{/if}
			</Card.Description>
		</Card.Header>
	</Card.Root>
</div>
{:else}
	<div class="flex min-h-[60vh] flex-col items-center justify-center gap-4 px-4 py-16 text-center">
		<div class="max-w-lg rounded-2xl border border-border bg-card p-8 shadow-sm">
			<h2 class="text-xl font-semibold text-foreground">Akses terbatas</h2>
			<p class="mt-3 text-sm leading-6 text-muted-foreground">
				Fase persiapan asesmen hanya tersedia untuk admin dan guru. Silakan kembali ke Beranda Asesmen untuk memilih pekerjaan lain.
			</p>
			<div class="mt-6">
				<Button href={resolve('/asesmen')} variant="outline">Kembali ke Beranda Asesmen</Button>
			</div>
		</div>
	</div>
{/if}
