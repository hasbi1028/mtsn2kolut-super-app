<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import type { RouteId } from '$app/types';
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { trackInternalAnalyticsEvent } from '$lib/analytics/internal-analytics';

	type AppRole = 'admin' | 'guru' | 'staf' | 'kesiswaan' | 'siswa' | 'ortu';
	type KnownRole = AppRole | (string & {});
	type CbtRoute = Extract<
		RouteId,
		| '/'
		| '/asesmen/persiapan'
		| '/asesmen/kegiatan'
		| '/asesmen/pelaksanaan'
		| '/asesmen/hasil'
		| '/asesmen/aplikasi-siswa'
		| '/asesmen/kegiatan'
		| '/asesmen/paket'
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
			name: 'Admin Ujian',
			description: 'Mulai dari persiapan asesmen, sesi ujian dan token ujian, pengawasan hari-H, lalu hasil.'
		},
		guru: {
			name: 'Guru',
			description: 'Ujian Berbasis Komputer dipakai untuk membaca persiapan, membantu pelaksanaan bila ditugaskan, dan membuka hasil. Penulisan soal ada di modul Bank Soal.'
		},
		staf: {
			name: 'Staf',
			description: 'Akses dibuat ringan untuk membantu Pengawasan Ruang dan membaca panduan BYOD.'
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
			description: 'Paket, kegiatan, jadwal, peserta, ruang, dan token ujian siap sebelum hari ujian.'
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
			phase: 'Evaluasi',
			title: 'Lihat Hasil',
			description: 'Buka rekap hasil ujian dan pemeriksaan yang relevan untuk guru.',
			href: '/asesmen/hasil',
			roles: ['guru', 'admin'],
			priority: { guru: 3, admin: 4 }
		},
		{
			phase: 'Siapkan',
			title: 'Baca Persiapan',
			description: 'Lihat jalur paket, kegiatan, sesi ujian, dan token ujian yang disiapkan untuk ujian.',
			href: '/asesmen/persiapan',
			roles: ['guru'],
			priority: { guru: 1 }
		},
		{
			phase: 'Jalankan',
			title: 'Pantau Pelaksanaan',
			description: 'Masuk ke ruang pelaksanaan bila ditugaskan membantu ujian hari-H.',
			href: '/asesmen/pelaksanaan',
			roles: ['guru'],
			priority: { guru: 2 }
		},
		{
			phase: 'Siapkan',
			title: 'Siapkan Asesmen',
			description: 'Mulai dari kegiatan ujian: paket, peserta, jadwal, dan ruang.',
			href: '/asesmen/persiapan',
			roles: ['admin'],
			priority: { admin: 1 }
		},
		{
			phase: 'Siapkan',
			title: 'Atur Sesi & Token Ujian',
			description: 'Kelola sesi ujian, kartu ujian, dan token ujian dari pusat Kegiatan Asesmen.',
			href: '/asesmen/kegiatan',
			roles: ['admin'],
			priority: { admin: 2 }
		},
		{
			phase: 'Jalankan',
			title: 'Pantau Ujian',
			description: 'Pantau pelaksanaan hari-H dan tindak lanjuti kebutuhan pengawas.',
			href: '/asesmen/pelaksanaan',
			roles: ['admin', 'staf'],
			priority: { admin: 3, staf: 1 }
		},
		{
			phase: 'Jalankan',
			title: 'Panduan BYOD',
			description: 'Baca ringkasan status perangkat, koneksi, dan kesiapan kirim ujian.',
			href: '/asesmen/aplikasi-siswa',
			roles: ['staf'],
			priority: { staf: 2 }
		}
	];

	const secondaryLinks: SecondaryLink[] = [
		{ label: 'Paket Soal', href: '/asesmen/paket', roles: ['admin'] },
		{ label: 'Kegiatan Asesmen', href: '/asesmen/kegiatan', roles: ['admin'] },
		{ label: 'Panduan BYOD', href: '/asesmen/aplikasi-siswa', roles: ['admin', 'guru'] }
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
	const roleDescription = $derived(launcherRole ? roleCopy[launcherRole].description : 'Belum ada pintasan ujian untuk peran aktif ini. Gunakan menu utama sesuai tugas masing-masing.');

	function resolveLauncherRole(roleSetValue: ReadonlySet<KnownRole>): LauncherRole | undefined {
		if (roleSetValue.has('admin')) return 'admin';
		if (roleSetValue.has('guru')) return 'guru';
		if (roleSetValue.has('staf')) return 'staf';
		return undefined;
	}

	function taskPriority(task: TaskCard, role: LauncherRole | undefined): number {
		return role ? (task.priority[role] ?? 99) : 99;
	}

	onMount(() => {
		void trackInternalAnalyticsEvent('asesmen.hub_view', {
			pathname: window.location.pathname,
			role: launcherRole,
			metadata: { page_key: 'asesmen' }
		});
	});
</script>

<svelte:head>
	<title>Beranda Asesmen CBT — MTsN 2 Kolaka Utara</title>
</svelte:head>

<div class="space-y-8">
	<section class="rounded-3xl border border-border bg-card p-6 shadow-sm md:p-8">
		<div class="max-w-4xl space-y-4">
			<p class="text-sm font-semibold uppercase tracking-[0.18em] text-primary">Beranda Asesmen CBT</p>
			<div class="space-y-3">
				<h1 class="text-3xl font-semibold tracking-tight text-foreground md:text-4xl">Apa yang perlu dikerjakan hari ini?</h1>
				<p class="max-w-2xl text-base leading-7 text-muted-foreground">
					{roleName}: {roleDescription}
				</p>
			</div>
		</div>
	</section>

	<section aria-labelledby="cbt-phases-title" class="space-y-4">
		<div class="space-y-1">
			<p class="text-sm font-semibold uppercase tracking-[0.18em] text-primary">3 fase besar</p>
			<h2 id="cbt-phases-title" class="text-2xl font-semibold tracking-tight text-foreground">Alur Asesmen CBT dibuat sederhana</h2>
		</div>

		<div class="grid gap-3 md:grid-cols-3">
			{#each phases as phase (phase.number)}
				<div class="rounded-2xl border border-border bg-card p-5">
					<div class="flex h-9 w-9 items-center justify-center rounded-full border border-primary/20 text-sm font-semibold text-primary">{phase.number}</div>
					<h3 class="mt-4 text-lg font-semibold text-foreground">{phase.title}</h3>
					<p class="mt-2 text-sm leading-6 text-muted-foreground">{phase.description}</p>
				</div>
			{/each}
		</div>
	</section>

	<section aria-labelledby="cbt-tasks-title" class="space-y-4">
		<div class="flex flex-wrap items-end justify-between gap-3">
			<div class="space-y-1">
				<p class="text-sm font-semibold uppercase tracking-[0.18em] text-primary">Pilih tugas</p>
				<h2 id="cbt-tasks-title" class="text-2xl font-semibold tracking-tight text-foreground">Pintasan sesuai peran</h2>
			</div>
			<p class="text-sm text-muted-foreground">Maksimal 4 tugas utama ditampilkan.</p>
		</div>

		{#if visibleTasks.length > 0}
			<div class="grid auto-rows-fr gap-4 md:grid-cols-2 xl:grid-cols-4">
				{#each visibleTasks as task (task.title)}
					<Card.Root class="flex h-full min-h-[17rem] flex-col border-border bg-card shadow-sm transition hover:border-primary/20 hover:shadow-md">
						<Card.Header class="flex-1 space-y-4 p-5">
							<p class="w-fit rounded-full border border-primary/20 bg-primary/10 px-3 py-1 text-xs font-semibold uppercase tracking-[0.14em] text-primary">
								{task.phase}
							</p>
							<div class="space-y-3">
								<Card.Title class="min-h-14 text-xl leading-7 text-foreground">{task.title}</Card.Title>
								<Card.Description class="min-h-20 text-sm leading-6">{task.description}</Card.Description>
							</div>
						</Card.Header>
						<Card.Footer class="mt-auto border-t border-border p-5 pt-4">
							<Button href={resolve(task.href)} class="w-full">Buka</Button>
						</Card.Footer>
					</Card.Root>
				{/each}
			</div>
		{:else}
			<Card.Root class="border-dashed border-border bg-muted/50 shadow-sm">
				<Card.Header>
					<Card.Title class="text-lg text-foreground">Tidak ada tugas CBT untuk peran ini</Card.Title>
					<Card.Description>Halaman ini tidak membuka modul yang tidak relevan dengan peran aktif.</Card.Description>
				</Card.Header>
				<Card.Footer>
					<Button href={resolve('/')} variant="outline">Kembali ke Beranda</Button>
				</Card.Footer>
			</Card.Root>
		{/if}
	</section>

	{#if visibleSecondaryLinks.length > 0}
		<nav aria-label="Tautan CBT lainnya" class="flex flex-wrap items-center gap-x-4 gap-y-2 border-t border-border pt-4 text-sm">
			<span class="font-medium text-muted-foreground">Tautan lain:</span>
			{#each visibleSecondaryLinks as link (link.href)}
				<a href={resolve(link.href)} class="font-medium text-primary underline-offset-4 hover:underline">{link.label}</a>
			{/each}
		</nav>
	{/if}
</div>
