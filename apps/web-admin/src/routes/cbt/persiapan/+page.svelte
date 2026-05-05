<script lang="ts">
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import * as Card from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';

	type PersiapanRoute =
		| '/cbt/soal'
		| '/cbt/soal/new'
		| '/cbt/soal/review'
		| '/cbt/packages'
		| '/cbt/packages/new'
		| '/cbt/events'
		| '/cbt/events/new'
		| '/cbt/sessions'
		| '/cbt/sessions/new'
		| '/cbt/pelaksanaan'
		| '/cbt/hasil'
		| '/cbt';

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
			title: 'Buat Soal',
			description: 'Tulis dan rapikan butir soal di Komposer Soal sebelum paket disusun.',
			href: '/cbt/soal/new',
			cta: 'Buat Soal'
		},
		{
			step: '02',
			title: 'Susun Paket',
			description: 'Pilih soal siap pakai, atur komposisi, dan siapkan paket untuk kegiatan ujian.',
			href: '/cbt/packages/new',
			cta: 'Susun Paket'
		},
		{
			step: '03',
			title: 'Buat Kegiatan',
			description: 'Daftarkan kegiatan ujian agar sesi, peserta, dan kartu ujian punya konteks yang jelas.',
			href: '/cbt/events/new',
			cta: 'Buat Kegiatan'
		},
		{
			step: '04',
			title: 'Atur Sesi/Token',
			description: 'Tetapkan jadwal, ruang, peserta, dan token sebelum ujian masuk hari pelaksanaan.',
			href: '/cbt/sessions/new',
			cta: 'Atur Sesi'
		}
	];

	const guruTasks: PreparationTask[] = [
		{
			step: '01',
			title: 'Buat Soal',
			description: 'Tulis dan rapikan butir soal di Komposer Soal tanpa masuk ke pengaturan operasional admin.',
			href: '/cbt/soal/new',
			cta: 'Buat Soal'
		},
		{
			step: '02',
			title: 'Review Soal',
			description: 'Periksa antrean review dan tindak lanjuti soal yang perlu keputusan atau revisi.',
			href: '/cbt/soal/review',
			cta: 'Buka Review'
		},
		{
			step: '03',
			title: 'Lihat Hasil',
			description: 'Masuk ke pintu hasil CBT yang tersedia untuk akun guru.',
			href: '/cbt/hasil',
			cta: 'Buka Hasil'
		}
	];

	const tasks = $derived(userRoles.includes('admin') ? adminTasks : guruTasks);
</script>

<svelte:head>
	<title>Persiapan CBT — MTsN 2 Kolaka Utara</title>
</svelte:head>

{#if canAccess}
	<div class="space-y-6">
	<section class="rounded-3xl border border-emerald-100 bg-gradient-to-br from-emerald-50 via-white to-lime-50 p-6 shadow-sm">
		<div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
			<div class="max-w-3xl space-y-3">
				<div class="flex flex-wrap items-center gap-2">
					<Badge class="border-emerald-200 bg-white text-emerald-700" variant="outline">CBT · Fase Persiapan</Badge>
					<Badge class="border-slate-200 bg-white text-slate-600" variant="outline">UI-only</Badge>
				</div>
				<h1 class="text-3xl font-semibold tracking-tight text-slate-950 md:text-4xl">Persiapan CBT</h1>
				<p class="max-w-2xl text-sm leading-6 text-slate-600 md:text-base">
					Mulai dari pekerjaan yang paling penting sebelum hari ujian: soal, paket, kegiatan, lalu sesi dan token. Halaman ini
					hanya berisi pintasan tugas, tanpa membaca data ujian.
				</p>
			</div>
			<div class="flex flex-wrap gap-3">
				<Button href={resolve('/cbt')} variant="outline">Beranda CBT</Button>
				<Button href={resolve('/cbt/pelaksanaan')}>Ke Pelaksanaan</Button>
			</div>
		</div>
	</section>

	<section aria-labelledby="persiapan-tasks-title" class="space-y-4">
		<div class="flex flex-wrap items-end justify-between gap-3">
			<div>
				<p class="text-xs font-semibold uppercase tracking-[0.22em] text-emerald-700">Daftar tugas</p>
				<h2 id="persiapan-tasks-title" class="mt-1 text-2xl font-semibold tracking-tight text-slate-950">Selesaikan berurutan</h2>
			</div>
			<p class="max-w-lg text-sm leading-6 text-slate-600">Empat kartu besar ini menjaga operator tetap fokus pada jalur kerja CBT yang direkomendasikan.</p>
		</div>

		<div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
			{#each tasks as task (task.href)}
				<Card.Root class="flex h-full flex-col border-slate-200 bg-white shadow-sm transition hover:-translate-y-0.5 hover:border-emerald-200 hover:shadow-md">
					<Card.Header class="space-y-4">
						<span class="flex h-12 w-12 items-center justify-center rounded-2xl border border-emerald-200 bg-emerald-50 text-lg font-semibold text-emerald-800">{task.step}</span>
						<div>
							<Card.Title class="text-xl text-slate-950">{task.title}</Card.Title>
							<Card.Description class="mt-2 leading-6">{task.description}</Card.Description>
						</div>
					</Card.Header>
					<Card.Footer class="mt-auto">
						<Button href={resolve(task.href)} variant="outline" class="w-full border-emerald-200 text-emerald-800 hover:bg-emerald-50">{task.cta}</Button>
					</Card.Footer>
				</Card.Root>
			{/each}
		</div>
	</section>

	<Card.Root class="border-emerald-200 bg-emerald-50/60 shadow-sm">
		<Card.Header>
			<Card.Title class="text-lg text-slate-950">Prinsip fase persiapan</Card.Title>
			<Card.Description>
				{#if userRoles.includes('admin')}
					Jika ada bagian yang belum siap, tetap selesaikan dari kiri ke kanan. Hindari membuka monitoring hari-H sebelum sesi,
					ruang, peserta, dan token sudah jelas.
				{:else}
					Untuk guru, fase persiapan difokuskan pada pekerjaan soal: menulis, mereview, dan membaca hasil yang sudah tersedia.
				{/if}
			</Card.Description>
		</Card.Header>
	</Card.Root>
</div>
{:else}
	<div class="flex min-h-[60vh] flex-col items-center justify-center gap-4 px-4 py-16 text-center">
		<div class="max-w-lg rounded-2xl border border-slate-200 bg-white p-8 shadow-sm">
			<h2 class="text-xl font-semibold text-slate-900">Akses terbatas</h2>
			<p class="mt-3 text-sm leading-6 text-slate-600">
				Fase persiapan CBT hanya tersedia untuk admin dan guru. Silakan kembali ke Dashboard CBT untuk memilih pekerjaan lain.
			</p>
			<div class="mt-6">
				<Button href={resolve('/cbt')} variant="outline">Kembali ke Dashboard CBT</Button>
			</div>
		</div>
	</div>
{/if}
