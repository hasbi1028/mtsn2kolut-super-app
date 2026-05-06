<script lang="ts">
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import * as Card from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';

	type PelaksanaanRoute = '/asesmen/aplikasi-siswa' | '/asesmen/sesi' | '/asesmen/pengawasan' | '/asesmen/kegiatan' | '/asesmen/persiapan' | '/asesmen';

	type DayTask = {
		title: string;
		description: string;
		href: PelaksanaanRoute;
		query?: string;
		cta: string;
		tone: 'monitor' | 'room' | 'guide' | 'print';
	};

	const userRoles = $derived(page.data.user?.roles ?? (page.data.user?.role ? [page.data.user.role] : []));
	const canAccess = $derived(userRoles.includes('admin') || userRoles.includes('guru') || userRoles.includes('staf'));
	const canOpenPersiapan = $derived(userRoles.includes('admin') || userRoles.includes('guru'));

	const adminDayTasks: DayTask[] = [
		{
			title: 'Pantau Sesi Hari Ini',
			description: 'Buka daftar sesi dengan fokus jadwal hari ini untuk memastikan ujian aktif dan token terkendali.',
			href: '/asesmen/sesi',
			query: '?schedule=today',
			cta: 'Pantau Sesi',
			tone: 'monitor'
		},
		{
			title: 'Dashboard Ruang',
			description: 'Masuk ke pantauan ruang untuk membantu pengawas membaca status peserta dan kebutuhan tindak lanjut.',
			href: '/asesmen/pengawasan',
			cta: 'Buka Ruang',
			tone: 'room'
		},
		{
			title: 'Panduan BYOD',
			description: 'Gunakan ringkasan status koneksi, submit readiness, dan panduan perangkat siswa saat ujian berlangsung.',
			href: '/asesmen/aplikasi-siswa',
			cta: 'Baca Panduan',
			tone: 'guide'
		},
		{
			title: 'Cetak Kartu via Kegiatan',
			description: 'Cetak kartu ujian dari detail kegiatan agar kartu tetap mengikuti peserta dan sesi yang benar.',
			href: '/asesmen/kegiatan',
			cta: 'Pilih Kegiatan',
			tone: 'print'
		}
	];

	const operatorDayTasks: DayTask[] = [
		{
			title: 'Dashboard Ruang',
			description: 'Masuk ke pantauan ruang untuk membantu pengawas membaca status peserta dan kebutuhan tindak lanjut.',
			href: '/asesmen/pengawasan',
			cta: 'Buka Ruang',
			tone: 'room'
		},
		{
			title: 'Panduan BYOD',
			description: 'Gunakan ringkasan status koneksi, submit readiness, dan panduan perangkat siswa saat ujian berlangsung.',
			href: '/asesmen/aplikasi-siswa',
			cta: 'Baca Panduan',
			tone: 'guide'
		},
		{
			title: 'Status Guide',
			description: 'Samakan bahasa status Tersambung, Lokal, Waspada, Gangguan, dan Menurun untuk siswa dan pengawas.',
			href: '/asesmen/aplikasi-siswa',
			cta: 'Buka Status Guide',
			tone: 'monitor'
		}
	];

	const dayTasks = $derived(userRoles.includes('admin') ? adminDayTasks : operatorDayTasks);

	function toneClass(tone: DayTask['tone']): string {
		switch (tone) {
			case 'monitor':
				return 'border-emerald-200 bg-emerald-50 text-emerald-800';
			case 'room':
				return 'border-teal-200 bg-teal-50 text-teal-800';
			case 'guide':
				return 'border-lime-200 bg-lime-50 text-lime-800';
			case 'print':
				return 'border-amber-200 bg-amber-50 text-amber-800';
		}
	}
</script>

<svelte:head>
	<title>Pelaksanaan CBT — MTsN 2 Kolaka Utara</title>
</svelte:head>

{#if canAccess}
	<div class="space-y-6">
	<section class="rounded-3xl border border-emerald-100 bg-gradient-to-br from-emerald-50 via-white to-lime-50 p-6 shadow-sm">
		<div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
			<div class="max-w-3xl space-y-3">
				<div class="flex flex-wrap items-center gap-2">
					<Badge class="border-emerald-200 bg-white text-emerald-700" variant="outline">CBT · Hari-H</Badge>
					<Badge class="border-slate-200 bg-white text-slate-600" variant="outline">Tanpa API</Badge>
				</div>
				<h1 class="text-3xl font-semibold tracking-tight text-slate-950 md:text-4xl">Pelaksanaan CBT</h1>
				<p class="max-w-2xl text-sm leading-6 text-slate-600 md:text-base">
					Fokus hari ujian dibuat ringkas: pantau sesi aktif, buka dashboard ruang, baca panduan BYOD, dan cetak kartu dari
					kegiatan bila diperlukan.
				</p>
			</div>
			<div class="flex flex-wrap gap-3">
				{#if canOpenPersiapan}
					<Button href={resolve('/asesmen/persiapan')} variant="outline">Kembali ke Persiapan</Button>
				{:else}
					<Button href={resolve('/asesmen')} variant="outline">Kembali ke Dashboard</Button>
				{/if}
				<Button href={resolve('/asesmen/aplikasi-siswa')}>Panduan BYOD</Button>
			</div>
		</div>
	</section>

	<section aria-labelledby="pelaksanaan-focus-title" class="grid gap-4 lg:grid-cols-[1fr_18rem]">
		<div class="space-y-4">
			<div>
				<p class="text-xs font-semibold uppercase tracking-[0.22em] text-emerald-700">Tugas hari-H</p>
				<h2 id="pelaksanaan-focus-title" class="mt-1 text-2xl font-semibold tracking-tight text-slate-950">Buka yang diperlukan saat ujian berjalan</h2>
			</div>

			<div class="grid gap-4 md:grid-cols-2">
				{#each dayTasks as task (task.title)}
					<Card.Root class="flex h-full flex-col border-slate-200 bg-white shadow-sm transition hover:-translate-y-0.5 hover:border-emerald-200 hover:shadow-md">
						<Card.Header class="space-y-3">
							<Badge class={toneClass(task.tone)} variant="outline">Hari-H</Badge>
							<div>
								<Card.Title class="text-xl text-slate-950">{task.title}</Card.Title>
								<Card.Description class="mt-2 leading-6">{task.description}</Card.Description>
							</div>
						</Card.Header>
						<Card.Footer class="mt-auto">
							<Button href={`${resolve(task.href)}${task.query ?? ''}`} variant="outline" class="w-full border-emerald-200 text-emerald-800 hover:bg-emerald-50">{task.cta}</Button>
						</Card.Footer>
					</Card.Root>
				{/each}
			</div>
		</div>

		<Card.Root class="h-fit border-emerald-200 bg-emerald-50/60 shadow-sm">
			<Card.Header>
				<Card.Title class="text-lg text-slate-950">Ritme operator</Card.Title>
			<Card.Description class="leading-6">
				{#if userRoles.includes('admin')}
					Mulai dari sesi hari ini, lanjutkan ke ruang bila ada peserta bermasalah, lalu gunakan BYOD sebagai bahasa bersama
					untuk siswa dan pengawas.
				{:else}
					Gunakan halaman ini untuk membuka dashboard ruang dan panduan BYOD tanpa masuk ke pengaturan sesi admin.
				{/if}
			</Card.Description>
			</Card.Header>
			<Card.Footer>
				<Button href={resolve('/asesmen')} variant="outline" class="w-full border-emerald-200 text-emerald-800 hover:bg-emerald-50">Beranda CBT</Button>
			</Card.Footer>
		</Card.Root>
	</section>
</div>
{:else}
	<div class="flex min-h-[60vh] flex-col items-center justify-center gap-4 px-4 py-16 text-center">
		<div class="max-w-lg rounded-2xl border border-slate-200 bg-white p-8 shadow-sm">
			<h2 class="text-xl font-semibold text-slate-900">Akses terbatas</h2>
			<p class="mt-3 text-sm leading-6 text-slate-600">
				Fase pelaksanaan CBT hanya tersedia untuk admin, guru, dan staf. Silakan kembali ke Dashboard CBT.
			</p>
			<div class="mt-6">
				<Button href={resolve('/asesmen')} variant="outline">Kembali ke Dashboard CBT</Button>
			</div>
		</div>
	</div>
{/if}
