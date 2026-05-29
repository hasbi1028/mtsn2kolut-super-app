<script lang="ts">
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';

	type PelaksanaanRoute =
		| '/asesmen/aplikasi-siswa'
		| '/asesmen/sesi'
		| '/asesmen/ruang-saya'
		| '/asesmen/kegiatan'
		| '/asesmen/persiapan'
		| '/asesmen/hasil'
		| '/asesmen';
	type RoleMode = 'admin' | 'guru' | 'staf';
	type TaskKind = 'panitia' | 'pengawas' | 'bantuan' | 'hasil';

	type DayTask = {
		step: string;
		title: string;
		description: string;
		meta: string;
		href: PelaksanaanRoute;
		query?: string;
		cta: string;
		kind: TaskKind;
		roles: RoleMode[];
	};

	const userRoles = $derived(page.data.user?.roles ?? (page.data.user?.role ? [page.data.user.role] : []));
	const userPermissions = $derived((page.data.user?.permissions ?? []).map((permission) => permission.trim()).filter(Boolean));
	const hasOperatorLane = $derived(
		userRoles.includes('admin')
			|| userPermissions.includes('asesmen.operator')
			|| userPermissions.includes('asesmen.event_manage')
			|| userPermissions.includes('asesmen.package_manage')
			|| userPermissions.includes('asesmen.session_manage')
			|| userPermissions.includes('asesmen.participant_manage')
	);
	const hasProctorLane = $derived(userPermissions.includes('asesmen.proctor'));
	const canOpenResults = $derived(userRoles.includes('admin') || userPermissions.includes('asesmen.result_read'));
	const canAccess = $derived(hasOperatorLane || hasProctorLane);
	const roleMode = $derived<RoleMode>(hasOperatorLane ? 'admin' : userRoles.includes('guru') ? 'guru' : 'staf');
	const isAdminMode = $derived(roleMode === 'admin');
	const roleName = $derived(isAdminMode ? 'Admin/Panitia' : roleMode === 'guru' ? 'Guru/Pengawas' : 'Staf/Operator');
	const heroTitle = $derived(isAdminMode ? 'Pelaksanaan Ujian' : 'Ruang Saya');
	const heroSubtitle = $derived(
		isAdminMode
			? 'Pilih jalur kerja hari-H: panitia mengelola sesi, pengawas membuka ruang, dan bantuan perangkat dipakai bila ada masalah siswa.'
			: 'Buka ruang yang ditugaskan. Pengawas tidak perlu masuk ke detail sesi panitia kecuali diminta operator.'
	);

	const baseDayTasks: DayTask[] = [
		{
			step: '1',
			title: 'Sesi Panitia',
			description: 'Untuk operator: cek jadwal hari ini, status sesi, ruang, peserta, pengawas, dan tindakan teknis panitia.',
			meta: 'Panitia / operator',
			href: '/asesmen/sesi',
			query: '?schedule=today',
			cta: 'Buka Sesi Hari Ini',
			kind: 'panitia',
			roles: ['admin']
		},
		{
			step: '2',
			title: 'Ruang Saya',
			description: 'Untuk pengawas: pilih ruang yang ditugaskan, lihat kode ruang, pantau peserta, dan tangani atensi.',
			meta: 'Pengawas ruang',
			href: '/asesmen/ruang-saya',
			cta: 'Buka Ruang Saya',
			kind: 'pengawas',
			roles: ['admin', 'guru', 'staf']
		},
		{
			step: '3',
			title: 'Perangkat Siswa',
			description: 'Panduan singkat saat siswa kesulitan masuk, status perangkat kuning/merah, atau butuh arahan operator.',
			meta: 'Bantuan lapangan',
			href: '/asesmen/aplikasi-siswa',
			cta: 'Buka Panduan',
			kind: 'bantuan',
			roles: ['admin', 'guru', 'staf']
		}
	];

	const dayTasks = $derived<DayTask[]>(
		canOpenResults
			? [
				...baseDayTasks,
				{
					step: '4',
					title: 'Hasil',
					description: 'Dipakai setelah ujian selesai untuk rekap nilai, berita acara, dan penutupan kegiatan.',
					meta: 'Akhir ujian',
					href: '/asesmen/hasil',
					cta: 'Lihat Hasil',
					kind: 'hasil',
					roles: ['admin', 'guru', 'staf']
				}
			]
			: baseDayTasks
	);

	const visibleTasks = $derived(dayTasks.filter((task) => task.roles.includes(roleMode)));
	const primaryTask = $derived(visibleTasks.find((task) => task.kind === (isAdminMode ? 'panitia' : 'pengawas')) ?? visibleTasks[0]);
	const secondaryTasks = $derived(visibleTasks.filter((task) => task !== primaryTask));

	function taskHref(task: DayTask): string {
		return `${resolve(task.href)}${task.query ?? ''}`;
	}

</script>

<svelte:head>
	<title>Pelaksanaan Ujian — MTsN 2 Kolaka Utara</title>
</svelte:head>

{#if canAccess}
	<div class="space-y-3">
		<section class="rounded-xl border border-primary/20 bg-card p-4 shadow-sm">
			<div class="grid gap-4 lg:grid-cols-[minmax(0,1fr)_18rem] lg:items-center">
				<div class="min-w-0">
					<div class="mb-2 flex flex-wrap items-center gap-2">
						<Badge class="border-primary/20 bg-primary/10 text-primary" variant="outline">Hari-H</Badge>
						<Badge class="border-border bg-muted text-muted-foreground" variant="outline">{roleName}</Badge>
					</div>
					<h1 class="text-2xl font-semibold tracking-tight text-foreground">{heroTitle}</h1>
					<p class="mt-2 max-w-2xl text-sm leading-6 text-muted-foreground">{heroSubtitle}</p>
				</div>
				{#if primaryTask}
					<div class="rounded-lg border border-primary/20 bg-primary/5 p-3">
						<p class="text-xs font-semibold uppercase tracking-[0.18em] text-primary">Mulai dari sini</p>
						<p class="mt-1 text-base font-semibold text-foreground">{primaryTask.title}</p>
						<p class="mt-1 text-xs leading-5 text-muted-foreground">{primaryTask.description}</p>
						<Button href={taskHref(primaryTask)} class="mt-3 w-full">
							{primaryTask.cta}
						</Button>
					</div>
				{/if}
			</div>
		</section>

		<section aria-labelledby="pelaksanaan-secondary-title" class="rounded-lg border border-border bg-muted/40 p-3">
			<div class="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
				<div>
					<h2 id="pelaksanaan-secondary-title" class="text-sm font-semibold text-foreground">Butuh yang lain?</h2>
					<p class="mt-1 text-xs leading-5 text-muted-foreground">Panel teknis dibuka hanya saat perlu tindakan. Pengawas cukup mulai dari Ruang Saya.</p>
				</div>
				<div class="flex flex-wrap gap-2">
					{#each secondaryTasks as task (task.title)}
						<a class="inline-flex min-h-9 items-center rounded-md border border-border bg-card px-3 text-sm font-medium text-foreground hover:border-primary/30 hover:text-primary" href={taskHref(task)}>
							<span class="mr-2 text-xs text-muted-foreground">{task.step}</span>{task.title}
						</a>
					{/each}
					<a class="inline-flex min-h-9 items-center rounded-md px-1 text-sm font-medium text-muted-foreground underline-offset-4 hover:underline" href={resolve('/asesmen')}>Ringkasan</a>
				</div>
			</div>
		</section>
	</div>
{:else}
	<div class="flex min-h-[60vh] flex-col items-center justify-center gap-4 px-4 py-16 text-center">
		<div class="max-w-lg rounded-2xl border border-border bg-card p-8 shadow-sm">
			<h2 class="text-xl font-semibold text-foreground">Akses terbatas</h2>
			<p class="mt-3 text-sm leading-6 text-muted-foreground">
				Halaman pelaksanaan ujian hanya tersedia untuk panitia atau pengawas yang diberi akses. Silakan kembali ke Ringkasan Asesmen.
			</p>
			<div class="mt-6">
				<Button href={resolve('/asesmen')} variant="outline">Ringkasan Asesmen</Button>
			</div>
		</div>
	</div>
{/if}
