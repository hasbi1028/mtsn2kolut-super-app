<script lang="ts">
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { MicroActionTable } from '$lib/components/ops';

	type PelaksanaanRoute =
		| '/asesmen/aplikasi-siswa'
		| '/asesmen/sesi'
		| '/asesmen/ruang-saya'
		| '/asesmen/pengawasan'
		| '/asesmen/kegiatan'
		| '/asesmen/persiapan'
		| '/asesmen/hasil'
		| '/asesmen/ringkas';
	type RoleMode = 'admin' | 'guru' | 'staf';
	type TaskKind = 'primary' | 'support' | 'result';

	type DayTask = {
		title: string;
		description: string;
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
			? 'Kelola sesi, ruang, kartu peserta, perangkat siswa, mode cadangan, dan hasil dari satu layar kerja.'
			: 'Buka ruang yang ditugaskan, pantau peserta, cek perangkat siswa, lalu hubungi panitia bila perlu.'
	);

	const baseDayTasks: DayTask[] = [
		{
			title: 'Persiapan',
			description: 'Kembali ke checklist kegiatan, paket, sesi, ruang, peserta, dan token.',
			href: '/asesmen/persiapan',
			cta: 'Buka Persiapan',
			kind: 'support',
			roles: ['admin']
		},
		{
			title: 'Sesi Hari Ini',
			description: 'Lihat jadwal/sesi aktif dan status ujian yang sedang berjalan.',
			href: '/asesmen/sesi',
			query: '?schedule=today',
			cta: 'Buka Sesi',
			kind: 'primary',
			roles: ['admin']
		},
		{
			title: 'Ruang Saya',
			description: 'Buka daftar ruang yang ditugaskan, status peserta, dan atensi yang perlu ditangani.',
			href: '/asesmen/ruang-saya',
			cta: 'Buka Ruang Saya',
			kind: 'primary',
			roles: ['admin', 'guru', 'staf']
		},
		{
			title: 'Cetak Kartu Peserta',
			description: 'Cetak kartu dari kegiatan ujian jika ada peserta yang membutuhkan salinan.',
			href: '/asesmen/kegiatan',
			cta: 'Pilih Kegiatan',
			kind: 'support',
			roles: ['admin']
		},
		{
			title: 'Perangkat Siswa',
			description: 'Panduan Portal Ujian Siswa, latihan lokal, dan mode cadangan untuk perangkat bermasalah.',
			href: '/asesmen/aplikasi-siswa',
			cta: 'Buka Panduan',
			kind: 'support',
			roles: ['admin', 'guru', 'staf']
		}
	];

	const dayTasks = $derived<DayTask[]>(
		canOpenResults
			? [
				...baseDayTasks,
				{
					title: 'Hasil',
					description: 'Buka rekap, nilai, dan hasil sesi setelah ujian selesai.',
					href: '/asesmen/hasil',
					cta: 'Lihat Hasil',
					kind: 'result',
					roles: ['admin', 'guru', 'staf']
				}
			]
			: baseDayTasks
	);

	const visibleTasks = $derived(dayTasks.filter((task) => task.roles.includes(roleMode)));
	const primaryTask = $derived(visibleTasks.find((task) => task.kind === 'primary') ?? visibleTasks[0]);
	const secondaryTasks = $derived(visibleTasks.filter((task) => task !== primaryTask));
	const dayTaskColumns = [
		{ key: 'task', label: 'Pekerjaan', class: 'min-w-56' },
		{ key: 'focus', label: 'Fokus', class: 'min-w-[20rem]' },
		{ key: 'kind', label: 'Jenis', headClass: 'text-right', class: 'text-right' }
	];

	function taskHref(task: DayTask): string {
		return `${resolve(task.href)}${task.query ?? ''}`;
	}

	function kindLabel(kind: TaskKind): string {
		switch (kind) {
			case 'primary':
				return 'Utama';
			case 'support':
				return 'Bantuan';
			case 'result':
				return 'Akhir';
		}
	}

	function kindClass(kind: TaskKind): string {
		switch (kind) {
			case 'primary':
				return 'border-primary/20 bg-primary/10 text-primary';
			case 'support':
				return 'border-border bg-muted text-muted-foreground';
			case 'result':
				return 'border-success/30 bg-success/10 text-success';
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
				{#if primaryTask}
					<div class="flex flex-wrap gap-2">
						<Button href={taskHref(primaryTask)} size="sm">{primaryTask.cta}</Button>
						<Button href={resolve('/asesmen/ringkas')} variant="outline" size="sm">Kembali ke Ringkasan</Button>
					</div>
				{/if}
			</div>
		</section>

		<section aria-labelledby="pelaksanaan-focus-title" class="space-y-3">
			<div class="flex flex-wrap items-end justify-between gap-3">
				<div>
					<p class="text-xs font-semibold uppercase tracking-[0.22em] text-primary">Alur Pelaksanaan</p>
					<h2 id="pelaksanaan-focus-title" class="mt-1 text-xl font-semibold tracking-tight text-foreground">
						{isAdminMode ? 'Kontrol panitia' : 'Tugas pengawasan'}
					</h2>
				</div>
				<p class="max-w-md text-xs leading-5 text-muted-foreground">
					{isAdminMode ? 'Admin melihat alur lengkap; guru/staf hanya melihat pekerjaan lapangan.' : 'Tampilan ini hanya menampilkan pekerjaan yang dibutuhkan pengawas.'}
				</p>
			</div>

			<MicroActionTable
				title={isAdminMode ? 'Pekerjaan Panitia' : 'Pekerjaan Pengawas'}
				description={isAdminMode ? 'Satu layar untuk persiapan akhir, pelaksanaan, dan hasil.' : 'Fokus pada ruang, peserta, perangkat, dan hasil.'}
				columns={dayTaskColumns}
				rows={visibleTasks}
				rowKey={(row) => (row as DayTask).title}
				tableClass="min-w-[720px]"
			>
				{#snippet cell(row, column)}
					{@const task = row as DayTask}
					{#if column.key === 'task'}
						<div class="font-semibold text-foreground">{task.title}</div>
					{:else if column.key === 'focus'}
						<p class="max-w-2xl text-xs leading-5 text-muted-foreground">{task.description}</p>
					{:else}
						<Badge class={kindClass(task.kind)} variant="outline">{kindLabel(task.kind)}</Badge>
					{/if}
				{/snippet}
				{#snippet actions(row)}
					{@const task = row as DayTask}
					<Button href={taskHref(task)} size="xs" variant={task.kind === 'primary' ? 'default' : 'outline'} class={task.kind === 'primary' ? '' : 'border-primary/20 text-primary hover:bg-primary/10'}>{task.cta}</Button>
				{/snippet}
				{#snippet mobile(row)}
					{@const task = row as DayTask}
					<div class="space-y-2">
						<div class="flex items-start justify-between gap-2">
							<div class="min-w-0">
								<p class="font-semibold text-foreground">{task.title}</p>
								<p class="mt-1 text-xs leading-5 text-muted-foreground">{task.description}</p>
							</div>
							<Badge class={`${kindClass(task.kind)} shrink-0`} variant="outline">{kindLabel(task.kind)}</Badge>
						</div>
						<Button href={taskHref(task)} size="sm" variant={task.kind === 'primary' ? 'default' : 'outline'} class={task.kind === 'primary' ? 'w-full' : 'w-full border-primary/20 text-primary hover:bg-primary/10'}>{task.cta}</Button>
					</div>
				{/snippet}
			</MicroActionTable>
		</section>

		{#if secondaryTasks.length > 0}
			<nav aria-label="Pintasan hari-H" class="flex flex-wrap items-center gap-2 rounded-lg border border-border bg-card p-3 text-sm shadow-sm">
				<span class="font-medium text-muted-foreground">Pintasan:</span>
				{#each secondaryTasks as task (task.title)}
					<a href={taskHref(task)} class="rounded-md border border-border px-3 py-1.5 font-medium text-foreground hover:border-primary/30 hover:bg-primary/10">{task.title}</a>
				{/each}
			</nav>
		{/if}
	</div>
{:else}
	<div class="flex min-h-[60vh] flex-col items-center justify-center gap-4 px-4 py-16 text-center">
		<div class="max-w-lg rounded-2xl border border-border bg-card p-8 shadow-sm">
			<h2 class="text-xl font-semibold text-foreground">Akses terbatas</h2>
			<p class="mt-3 text-sm leading-6 text-muted-foreground">
				Halaman pelaksanaan ujian hanya tersedia untuk panitia atau pengawas yang diberi akses. Silakan kembali ke Ringkasan.
			</p>
			<div class="mt-6">
				<Button href={resolve('/asesmen/ringkas')} variant="outline">Kembali ke Ringkasan</Button>
			</div>
		</div>
	</div>
{/if}
