<script lang="ts">
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import * as Card from '$lib/components/ui/card';
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

	function taskHref(task: DayTask): string {
		return `${resolve(task.href)}${task.query ?? ''}`;
	}

	function kindLabel(kind: TaskKind): string {
		switch (kind) {
			case 'panitia':
				return 'Panitia';
			case 'pengawas':
				return 'Pengawas';
			case 'bantuan':
				return 'Bantuan';
			case 'hasil':
				return 'Akhir';
		}
	}

	function kindClass(kind: TaskKind): string {
		switch (kind) {
			case 'panitia':
				return 'border-primary/20 bg-primary/10 text-primary';
			case 'pengawas':
				return 'border-success/30 bg-success/10 text-success';
			case 'bantuan':
				return 'border-border bg-muted text-muted-foreground';
			case 'hasil':
				return 'border-warning/30 bg-warning/10 text-warning';
		}
	}
</script>

<svelte:head>
	<title>Pelaksanaan Ujian — MTsN 2 Kolaka Utara</title>
</svelte:head>

{#if canAccess}
	<div class="space-y-4">
		<section class="rounded-2xl border border-border bg-card p-4 shadow-sm">
			<div class="flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between">
				<div class="min-w-0 space-y-2">
					<div class="flex flex-wrap items-center gap-2">
						<Badge class="border-primary/20 bg-primary/10 text-primary" variant="outline">Ujian Digital</Badge>
						<Badge class="border-border bg-muted text-muted-foreground" variant="outline">{roleName}</Badge>
					</div>
					<h1 class="text-2xl font-semibold tracking-tight text-foreground md:text-3xl">{heroTitle}</h1>
					<p class="max-w-2xl text-sm leading-6 text-muted-foreground">{heroSubtitle}</p>
				</div>
				<a class="inline-flex items-center rounded-md px-1 text-sm font-medium text-muted-foreground underline-offset-4 hover:underline" href={resolve('/asesmen')}>Ringkasan</a>
			</div>
		</section>

		<section aria-labelledby="pelaksanaan-focus-title" class="space-y-3">
			<div class="flex flex-wrap items-end justify-between gap-3">
				<div>
					<p class="text-xs font-semibold uppercase tracking-[0.22em] text-primary">Alur Pelaksanaan</p>
					<h2 id="pelaksanaan-focus-title" class="mt-1 text-xl font-semibold tracking-tight text-foreground">
						{isAdminMode ? 'Pilih jalur kerja' : 'Tugas pengawas ruang'}
					</h2>
				</div>
				<p class="max-w-md text-xs leading-5 text-muted-foreground">
					{isAdminMode ? 'Sesi untuk panitia, Ruang Saya untuk pengawas. Jangan masuk ke panel teknis kalau hanya perlu mengawasi ruang.' : 'Mulai dari Ruang Saya. Panel teknis sesi disediakan untuk panitia.'}
				</p>
			</div>

			<div class="grid gap-3 lg:grid-cols-3">
				{#each visibleTasks as task (task.title)}
					<Card.Root class={task.kind === 'pengawas' ? 'border-success/30 shadow-sm' : 'border-border shadow-sm'}>
						<Card.Header class="space-y-3">
							<div class="flex items-center justify-between gap-3">
								<span class="inline-flex h-8 min-w-8 items-center justify-center rounded-md border border-primary/20 bg-primary/10 px-2 text-xs font-semibold text-primary">{task.step}</span>
								<Badge class={kindClass(task.kind)} variant="outline">{kindLabel(task.kind)}</Badge>
							</div>
							<div>
								<Card.Title class="text-base">{task.title}</Card.Title>
								<Card.Description class="mt-1 leading-6">{task.description}</Card.Description>
							</div>
							<p class="text-xs font-medium text-muted-foreground">{task.meta}</p>
						</Card.Header>
						<Card.Content>
							<Button href={taskHref(task)} size="sm" variant={task.kind === 'pengawas' || task.kind === 'panitia' ? 'default' : 'outline'} class={task.kind === 'pengawas' || task.kind === 'panitia' ? 'w-full' : 'w-full border-primary/20 text-primary hover:bg-primary/10'}>
								{task.cta}
							</Button>
						</Card.Content>
					</Card.Root>
				{/each}
			</div>
		</section>

		<section class="rounded-lg border border-dashed border-border bg-muted/40 p-3 text-xs leading-5 text-muted-foreground">
			<p class="font-semibold text-foreground">Batas sederhana:</p>
			<p><span class="font-semibold">Sesi</span> = kendali panitia/operator. <span class="font-semibold">Ruang Saya</span> = layar kerja pengawas. <span class="font-semibold">Pengawasan</span> = panel teknis yang dibuka dari sesi atau ruang saat perlu tindakan.</p>
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
