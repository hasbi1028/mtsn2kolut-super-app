<script lang="ts">
	import { resolve } from '$app/paths';
	import * as Card from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';

	type UserRole = 'admin' | 'guru' | 'staf' | 'kesiswaan' | 'siswa' | 'ortu' | 'reviewer' | string;
	type CbtRoute = '/cbt/soal' | '/cbt/packages' | '/cbt/events' | '/cbt/byod' | '/cbt/hasil' | '/';

	type ModuleCard = {
		title: string;
		label: string;
		description: string;
		href: CbtRoute;
		roles: UserRole[];
		tone: 'authoring' | 'assembly' | 'operation' | 'monitoring' | 'reporting';
	};

	type FlowStep = {
		title: string;
		description: string;
	};

	type RoleMode = {
		eyebrow: string;
		title: string;
		description: string;
		ctaLabel: string;
		ctaHref: CbtRoute;
		badge: string;
	};

	let { data }: {
		data: {
			user?: {
				role?: string;
				roles?: string[];
			};
		};
	} = $props();

	const modules: ModuleCard[] = [
		{
			title: 'Bank Soal',
			label: 'Komposer',
			description: 'Tulis, susun, dan rapikan butir soal dari jalur /cbt/soal yang aktif.',
			href: '/cbt/soal',
			roles: ['admin', 'guru', 'reviewer'],
			tone: 'authoring'
		},
		{
			title: 'Paket Soal',
			label: 'Admin',
			description: 'Rakit paket dan komposisi ujian. Disembunyikan dari guru karena route ini admin-only.',
			href: '/cbt/packages',
			roles: ['admin'],
			tone: 'assembly'
		},
		{
			title: 'Kegiatan & Sesi',
			label: 'Admin',
			description: 'Atur event, jadwal sesi, peserta, ruang, token, dan kesiapan operasional.',
			href: '/cbt/events',
			roles: ['admin'],
			tone: 'operation'
		},
		{
			title: 'Monitoring',
			label: 'Hari-H',
			description: 'Baca panduan BYOD, status koneksi, dan akses ringkas untuk pantauan ujian.',
			href: '/cbt/byod',
			roles: ['admin', 'guru', 'staf'],
			tone: 'monitoring'
		},
		{
			title: 'Hasil & Analisis',
			label: 'Rekap',
			description: 'Masuk ke hasil ujian dan analisis yang aman untuk admin dan guru.',
			href: '/cbt/hasil',
			roles: ['admin', 'guru'],
			tone: 'reporting'
		}
	];

	const flowSteps: FlowStep[] = [
		{
			title: '1. Siapkan Soal',
			description: 'Guru mulai dari Bank Soal; admin dapat meninjau kelengkapan sebelum paket dibuat.'
		},
		{
			title: '2. Rakit Sesi',
			description: 'Admin menghubungkan paket, jadwal, peserta, ruang, dan kartu ujian.'
		},
		{
			title: '3. Pantau Hari-H',
			description: 'Operator membuka monitoring, membaca status BYOD, lalu menindaklanjuti hasil.'
		}
	];

	const roles = $derived(data.user?.roles ?? (data.user?.role ? [data.user.role] : []));
	const roleSet = $derived(new Set(roles));
	const visibleModules = $derived(modules.filter((module) => module.roles.some((role) => roleSet.has(role))));
	const roleLabel = $derived(roles.length > 0 ? roles.join(' / ') : 'pengguna');
	const primaryMode = $derived.by<RoleMode>(() => {
		if (roleSet.has('admin')) {
			return {
				eyebrow: 'Beranda CBT · Mode Operasional',
				title: 'Pusat Kendali CBT Madrasah',
				description: 'Arahkan pekerjaan CBT dari penyusunan soal, paket, sesi ujian, monitoring BYOD, sampai hasil dalam urutan yang jelas.',
				ctaLabel: 'Kelola Kegiatan & Sesi',
				ctaHref: '/cbt/events',
				badge: 'Admin melihat 5 modul'
			};
		}

		if (roleSet.has('guru')) {
			return {
				eyebrow: 'Beranda CBT · Mode Guru',
				title: 'Ruang CBT Sederhana untuk Guru',
				description: 'Fokus pada pekerjaan aman untuk guru: menyiapkan Bank Soal, melihat monitoring yang relevan, dan membuka hasil ujian.',
				ctaLabel: 'Buka Bank Soal',
				ctaHref: '/cbt/soal',
				badge: 'Paket disembunyikan'
			};
		}

		if (roleSet.has('staf')) {
			return {
				eyebrow: 'Beranda CBT · Mode Staf',
				title: 'Dukungan Monitoring CBT',
				description: 'Akses CBT staf dibuat ringan: masuk ke monitoring BYOD untuk membaca status hari-H tanpa membuka area admin.',
				ctaLabel: 'Buka Monitoring',
				ctaHref: '/cbt/byod',
				badge: 'Staf melihat monitoring'
			};
		}

		return {
			eyebrow: 'Beranda CBT · Tidak Ada Modul',
			title: 'CBT Belum Tersedia untuk Peran Ini',
			description: 'Peran kesiswaan, siswa, dan orang tua tidak memiliki kartu CBT di beranda admin ini. Gunakan menu utama sesuai tugas masing-masing.',
			ctaLabel: 'Kembali ke Beranda',
			ctaHref: '/',
			badge: 'Tanpa kartu CBT'
		};
	});

	function moduleToneClass(tone: ModuleCard['tone']): string {
		switch (tone) {
			case 'authoring':
				return 'border-emerald-200 bg-emerald-50/70 text-emerald-800';
			case 'assembly':
				return 'border-lime-200 bg-lime-50 text-lime-800';
			case 'operation':
				return 'border-teal-200 bg-teal-50 text-teal-800';
			case 'monitoring':
				return 'border-amber-200 bg-amber-50 text-amber-800';
			case 'reporting':
				return 'border-sky-200 bg-sky-50 text-sky-800';
		}
	}
</script>

<svelte:head>
	<title>Beranda CBT — MTsN 2 Kolaka Utara</title>
</svelte:head>

<div class="space-y-6">
	<section class="overflow-hidden rounded-3xl border border-emerald-100 bg-gradient-to-br from-emerald-50 via-white to-lime-50 shadow-sm">
		<div class="grid gap-6 p-6 lg:grid-cols-[1fr_20rem] lg:p-8">
			<div class="space-y-4">
				<div class="flex flex-wrap items-center gap-2">
					<Badge class="border-emerald-200 bg-white text-emerald-700" variant="outline">{primaryMode.eyebrow}</Badge>
					<Badge class="border-slate-200 bg-white text-slate-600" variant="outline">Role: {roleLabel}</Badge>
				</div>
				<div class="max-w-3xl space-y-3">
					<h1 class="text-3xl font-semibold tracking-tight text-slate-950 md:text-4xl">{primaryMode.title}</h1>
					<p class="text-sm leading-6 text-slate-600 md:text-base">{primaryMode.description}</p>
				</div>
				<div class="flex flex-wrap items-center gap-3">
					<Button href={resolve(primaryMode.ctaHref)}>{primaryMode.ctaLabel}</Button>
					<span class="text-sm font-medium text-emerald-800">{primaryMode.badge}</span>
				</div>
			</div>

			<Card.Root class="border-emerald-200 bg-white/85 shadow-sm">
				<Card.Header>
					<Card.Title class="text-lg text-slate-950">Mode sederhana</Card.Title>
					<Card.Description>Beranda ini hanya menampilkan rute yang aman untuk role aktif.</Card.Description>
				</Card.Header>
				<Card.Content class="space-y-3 text-sm text-slate-600">
					<div class="rounded-2xl border border-emerald-100 bg-emerald-50/70 p-4">
						<p class="font-semibold text-emerald-900">{visibleModules.length} modul terlihat</p>
						<p class="mt-1 leading-6">Tidak ada data live dan tidak ada panggilan API dari halaman ini.</p>
					</div>
				</Card.Content>
			</Card.Root>
		</div>
	</section>

	<section aria-labelledby="cbt-flow-title" class="rounded-3xl border border-slate-200 bg-white p-4 shadow-sm md:p-5">
		<div class="mb-4 flex flex-wrap items-center justify-between gap-3">
			<div>
				<p class="text-xs font-semibold uppercase tracking-[0.22em] text-emerald-700">Alur kerja</p>
				<h2 id="cbt-flow-title" class="mt-1 text-xl font-semibold text-slate-950">Dari komposisi soal sampai hasil</h2>
			</div>
			<Badge class="border-emerald-200 text-emerald-700" variant="outline">Dashboard operasional</Badge>
		</div>

		<div class="grid gap-3 md:grid-cols-3">
			{#each flowSteps as step (step.title)}
				<div class="rounded-2xl border border-emerald-100 bg-emerald-50/50 p-4">
					<p class="font-semibold text-emerald-950">{step.title}</p>
					<p class="mt-2 text-sm leading-6 text-slate-600">{step.description}</p>
				</div>
			{/each}
		</div>
	</section>

	<section aria-labelledby="cbt-modules-title" class="space-y-4">
		<div class="flex flex-wrap items-end justify-between gap-3">
			<div>
				<p class="text-xs font-semibold uppercase tracking-[0.22em] text-emerald-700">Modul CBT</p>
				<h2 id="cbt-modules-title" class="mt-1 text-2xl font-semibold tracking-tight text-slate-950">Kartu sesuai peran</h2>
			</div>
			<p class="max-w-xl text-sm leading-6 text-slate-600">Admin melihat seluruh rangkaian. Guru dan staf hanya melihat jalur yang selaras dengan route guard yang sudah ada.</p>
		</div>

		{#if visibleModules.length > 0}
			<div class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
				{#each visibleModules as module (module.href)}
					<Card.Root class="group border-slate-200 bg-white shadow-sm transition hover:-translate-y-0.5 hover:border-emerald-200 hover:shadow-md">
						<Card.Header class="space-y-3">
							<Badge class={moduleToneClass(module.tone)} variant="outline">{module.label}</Badge>
							<div>
								<Card.Title class="text-lg text-slate-950">{module.title}</Card.Title>
								<Card.Description class="mt-2 leading-6">{module.description}</Card.Description>
							</div>
						</Card.Header>
						<Card.Footer>
							<Button href={resolve(module.href)} variant="outline" class="border-emerald-200 text-emerald-800 hover:bg-emerald-50">Buka Modul</Button>
						</Card.Footer>
					</Card.Root>
				{/each}
			</div>
		{:else}
			<Card.Root class="border-dashed border-slate-300 bg-slate-50 shadow-sm">
				<Card.Header>
					<Card.Title class="text-lg text-slate-950">Tidak ada kartu CBT untuk role ini</Card.Title>
					<Card.Description>Beranda CBT tetap aman dibuka, tetapi tidak menawarkan pintasan ke modul yang tidak relevan.</Card.Description>
				</Card.Header>
				<Card.Footer>
					<Button href={resolve('/')} variant="outline">Kembali ke Beranda</Button>
				</Card.Footer>
			</Card.Root>
		{/if}
	</section>
</div>
