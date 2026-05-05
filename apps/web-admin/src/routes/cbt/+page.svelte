<script lang="ts">
	import { resolve } from '$app/paths';
	import type { RouteId } from '$app/types';
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';

	type AppRole = 'admin' | 'guru' | 'staf' | 'kesiswaan' | 'siswa' | 'ortu';
	type KnownRole = AppRole | (string & {});
	type CbtRoute = Extract<
		RouteId,
		| '/'
		| '/cbt/soal'
		| '/cbt/soal/review'
		| '/cbt/persiapan'
		| '/cbt/events'
		| '/cbt/pelaksanaan'
		| '/cbt/hasil'
		| '/cbt/byod'
		| '/cbt/bank-soal'
		| '/cbt/kegiatan'
		| '/cbt/paket-soal'
	>;
	type LauncherRole = 'admin' | 'guru' | 'staf';

	type TaskCard = {
		phase: string;
		title: string;
		description: string;
		href: CbtRoute;
		roles: LauncherRole[];
		priority: Partial<Record<LauncherRole, number>>;
	};

	type Phase = {
		number: string;
		title: string;
		description: string;
	};

	type SecondaryLink = {
		label: string;
		href: CbtRoute;
		roles: LauncherRole[];
	};

	const roleCopy: Record<LauncherRole, { name: string; description: string }> = {
		admin: {
			name: 'Admin CBT',
			description: 'Mulai dari persiapan ujian, sesi dan token, pemantauan hari-H, lalu hasil.'
		},
		guru: {
			name: 'Guru',
			description: 'Fokus ke pekerjaan inti: kelola Bank Soal, review kesiapan soal, dan lihat hasil.'
		},
		staf: {
			name: 'Staf',
			description: 'Akses dibuat ringan untuk membantu pantauan ujian dan membaca panduan BYOD.'
		}
	};

	let { data }: {
		data: {
			user?: {
				role?: string;
				roles?: string[];
			};
		};
	} = $props();

	const phases: Phase[] = [
		{
			number: '1',
			title: 'Siapkan',
			description: 'Soal, paket, jadwal, peserta, ruang, dan token siap sebelum hari ujian.'
		},
		{
			number: '2',
			title: 'Jalankan',
			description: 'Operator memantau sesi, koneksi BYOD, dan kebutuhan pengawas saat ujian berlangsung.'
		},
		{
			number: '3',
			title: 'Evaluasi',
			description: 'Guru dan admin membuka hasil untuk pemeriksaan, rekap, dan tindak lanjut.'
		}
	];

	const tasks: TaskCard[] = [
		{
			phase: 'Bank Soal',
			title: 'Kelola Bank Soal',
			description: 'Masuk ke ruang Bank Soal untuk menulis, mengimpor, dan merapikan butir reusable.',
			href: '/cbt/bank-soal',
			roles: ['guru'],
			priority: { guru: 1 }
		},
		{
			phase: 'Bank Soal',
			title: 'Review Soal',
			description: 'Periksa antrean review agar soal siap digunakan saat paket ujian dirakit.',
			href: '/cbt/soal/review',
			roles: ['guru'],
			priority: { guru: 2 }
		},
		{
			phase: 'Evaluasi',
			title: 'Lihat Hasil',
			description: 'Buka rekap hasil ujian dan pemeriksaan yang relevan untuk guru.',
			href: '/cbt/hasil',
			roles: ['guru', 'admin'],
			priority: { guru: 3, admin: 4 }
		},
		{
			phase: 'Siapkan',
			title: 'Siapkan Ujian',
			description: 'Mulai dari kegiatan ujian: paket, peserta, jadwal, dan ruang.',
			href: '/cbt/persiapan',
			roles: ['admin'],
			priority: { admin: 1 }
		},
		{
			phase: 'Siapkan',
			title: 'Atur Sesi & Token',
			description: 'Kelola sesi, kartu ujian, dan token dari pusat kegiatan CBT.',
			href: '/cbt/events',
			roles: ['admin'],
			priority: { admin: 2 }
		},
		{
			phase: 'Jalankan',
			title: 'Pantau Ujian',
			description: 'Pantau pelaksanaan hari-H dan tindak lanjuti kebutuhan pengawas.',
			href: '/cbt/pelaksanaan',
			roles: ['admin', 'staf'],
			priority: { admin: 3, staf: 1 }
		},
		{
			phase: 'Jalankan',
			title: 'Panduan BYOD',
			description: 'Baca ringkasan status perangkat, koneksi, dan kesiapan submit.',
			href: '/cbt/byod',
			roles: ['staf'],
			priority: { staf: 2 }
		}
	];

	const secondaryLinks: SecondaryLink[] = [
		{ label: 'Bank Soal', href: '/cbt/bank-soal', roles: ['admin', 'guru'] },
		{ label: 'Paket Soal', href: '/cbt/paket-soal', roles: ['admin'] },
		{ label: 'Kegiatan lama', href: '/cbt/kegiatan', roles: ['admin'] },
		{ label: 'Panduan BYOD', href: '/cbt/byod', roles: ['admin', 'guru'] }
	];

	const userRoles = $derived<KnownRole[]>(data.user?.roles ?? (data.user?.role ? [data.user.role] : []));
	const roleSet = $derived(new Set<KnownRole>(userRoles));
	const launcherRole = $derived<LauncherRole | undefined>(resolveLauncherRole(roleSet));
	const visibleTasks = $derived.by<TaskCard[]>(() => {
		return tasks
			.filter((task) => task.roles.some((role) => roleSet.has(role)))
			.toSorted((firstTask, secondTask) => taskPriority(firstTask, launcherRole) - taskPriority(secondTask, launcherRole))
			.slice(0, 4);
	});
	const visibleSecondaryLinks = $derived(secondaryLinks.filter((link) => link.roles.some((role) => roleSet.has(role))));
	const roleName = $derived(launcherRole ? roleCopy[launcherRole].name : 'Peran ini');
	const roleDescription = $derived(launcherRole ? roleCopy[launcherRole].description : 'CBT belum menyediakan pintasan untuk peran aktif ini. Gunakan menu utama sesuai tugas masing-masing.');

	function resolveLauncherRole(roleSetValue: ReadonlySet<KnownRole>): LauncherRole | undefined {
		if (roleSetValue.has('admin')) return 'admin';
		if (roleSetValue.has('guru')) return 'guru';
		if (roleSetValue.has('staf')) return 'staf';
		return undefined;
	}

	function taskPriority(task: TaskCard, role: LauncherRole | undefined): number {
		return role ? (task.priority[role] ?? 99) : 99;
	}
</script>

<svelte:head>
	<title>Beranda CBT — MTsN 2 Kolaka Utara</title>
</svelte:head>

<div class="space-y-8">
	<section class="rounded-3xl border border-slate-200 bg-white p-6 shadow-sm md:p-8">
		<div class="max-w-4xl space-y-4">
			<p class="text-sm font-semibold uppercase tracking-[0.18em] text-emerald-800">Beranda CBT</p>
			<div class="space-y-3">
				<h1 class="text-3xl font-semibold tracking-tight text-slate-950 md:text-4xl">Apa yang perlu dikerjakan hari ini?</h1>
				<p class="max-w-2xl text-base leading-7 text-slate-600">
					{roleName}: {roleDescription}
				</p>
			</div>
		</div>
	</section>

	<section aria-labelledby="cbt-phases-title" class="space-y-4">
		<div class="space-y-1">
			<p class="text-sm font-semibold uppercase tracking-[0.18em] text-emerald-800">3 fase besar</p>
			<h2 id="cbt-phases-title" class="text-2xl font-semibold tracking-tight text-slate-950">Alur CBT dibuat sederhana</h2>
		</div>

		<div class="grid gap-3 md:grid-cols-3">
			{#each phases as phase (phase.number)}
				<div class="rounded-2xl border border-slate-200 bg-white p-5">
					<div class="flex h-9 w-9 items-center justify-center rounded-full border border-emerald-200 text-sm font-semibold text-emerald-800">{phase.number}</div>
					<h3 class="mt-4 text-lg font-semibold text-slate-950">{phase.title}</h3>
					<p class="mt-2 text-sm leading-6 text-slate-600">{phase.description}</p>
				</div>
			{/each}
		</div>
	</section>

	<section aria-labelledby="cbt-tasks-title" class="space-y-4">
		<div class="flex flex-wrap items-end justify-between gap-3">
			<div class="space-y-1">
				<p class="text-sm font-semibold uppercase tracking-[0.18em] text-emerald-800">Pilih tugas</p>
				<h2 id="cbt-tasks-title" class="text-2xl font-semibold tracking-tight text-slate-950">Pintasan sesuai peran</h2>
			</div>
			<p class="text-sm text-slate-500">Maksimal 4 tugas utama ditampilkan.</p>
		</div>

		{#if visibleTasks.length > 0}
			<div class="grid auto-rows-fr gap-4 md:grid-cols-2 xl:grid-cols-4">
				{#each visibleTasks as task (task.title)}
					<Card.Root class="flex h-full min-h-[17rem] flex-col border-slate-200 bg-white shadow-sm transition hover:border-emerald-300 hover:shadow-md">
						<Card.Header class="flex-1 space-y-4 p-5">
							<p class="w-fit rounded-full border border-emerald-100 bg-emerald-50 px-3 py-1 text-xs font-semibold uppercase tracking-[0.14em] text-emerald-800">
								{task.phase}
							</p>
							<div class="space-y-3">
								<Card.Title class="min-h-14 text-xl leading-7 text-slate-950">{task.title}</Card.Title>
								<Card.Description class="min-h-20 text-sm leading-6">{task.description}</Card.Description>
							</div>
						</Card.Header>
						<Card.Footer class="mt-auto border-t border-slate-100 p-5 pt-4">
							<Button href={resolve(task.href)} class="w-full">Buka</Button>
						</Card.Footer>
					</Card.Root>
				{/each}
			</div>
		{:else}
			<Card.Root class="border-dashed border-slate-300 bg-slate-50 shadow-sm">
				<Card.Header>
					<Card.Title class="text-lg text-slate-950">Tidak ada tugas CBT untuk peran ini</Card.Title>
					<Card.Description>Halaman ini tidak membuka modul yang tidak relevan dengan role aktif.</Card.Description>
				</Card.Header>
				<Card.Footer>
					<Button href={resolve('/')} variant="outline">Kembali ke Beranda</Button>
				</Card.Footer>
			</Card.Root>
		{/if}
	</section>

	{#if visibleSecondaryLinks.length > 0}
		<nav aria-label="Tautan CBT lainnya" class="flex flex-wrap items-center gap-x-4 gap-y-2 border-t border-slate-200 pt-4 text-sm">
			<span class="font-medium text-slate-500">Tautan lain:</span>
			{#each visibleSecondaryLinks as link (link.href)}
				<a href={resolve(link.href)} class="font-medium text-emerald-800 underline-offset-4 hover:underline">{link.label}</a>
			{/each}
		</nav>
	{/if}
</div>
