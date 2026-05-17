<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import type { RouteId } from '$app/types';
	import { Button } from '$lib/components/ui/button';
	import { ContextStrip, MetricCard, PageHeader, WorkflowCard } from '$lib/components/ops';
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

	type Workflow = {
		title: string;
		description: string;
		href: CbtRoute;
		actionLabel: string;
		status: string;
		roles: LauncherRole[];
		priority: Partial<Record<LauncherRole, number>>;
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

	const workflows: Workflow[] = [
		{
			title: 'Persiapan Ujian',
			description: 'Cek kegiatan, paket, peserta, sesi, ruang, token, dan kartu sebelum hari-H.',
			href: '/asesmen/persiapan',
			actionLabel: 'Buka Persiapan',
			status: 'Pra ujian',
			roles: ['admin', 'guru'],
			priority: { admin: 1, guru: 1 }
		},
		{
			title: 'Monitor Ujian',
			description: 'Pantau sesi aktif, ruang pengawasan, peserta bermasalah, dan insiden.',
			href: '/asesmen/pelaksanaan',
			actionLabel: 'Buka Monitor',
			status: 'Hari-H',
			roles: ['admin', 'staf'],
			priority: { admin: 2, staf: 1 }
		},
		{
			title: 'Hasil & BA',
			description: 'Buka rekap jawaban, koreksi, nilai, berita acara, dan unduhan operasional.',
			href: '/asesmen/hasil',
			actionLabel: 'Buka Hasil',
			status: 'Pasca ujian',
			roles: ['admin', 'guru'],
			priority: { admin: 3, guru: 2 }
		},
		{
			title: 'Arsip',
			description: 'Kunci dokumen final, cek pengesahan, dan simpan bukti kegiatan asesmen.',
			href: '/asesmen/kegiatan',
			actionLabel: 'Buka Kegiatan',
			status: 'Dokumen',
			roles: ['admin'],
			priority: { admin: 4 }
		}
	];

	const secondaryLinks: SecondaryLink[] = [
		{ label: 'Buat Kegiatan', href: '/asesmen/kegiatan', roles: ['admin'] },
		{ label: 'Cek Kesiapan', href: '/asesmen/persiapan', roles: ['admin', 'guru'] },
		{ label: 'Buka Monitor', href: '/asesmen/pelaksanaan', roles: ['admin', 'guru', 'staf'] },
		{ label: 'Lihat Hasil', href: '/asesmen/hasil', roles: ['admin', 'guru'] },
		{ label: 'Panduan BYOD', href: '/asesmen/aplikasi-siswa', roles: ['admin', 'guru'] }
	];

	const userRoles = $derived<KnownRole[]>(data.user?.roles ?? (data.user?.role ? [data.user.role] : []));
	const roleSet = $derived(new Set<KnownRole>(userRoles));
	const launcherRole = $derived<LauncherRole | undefined>(resolveLauncherRole(roleSet));
	const visibleWorkflows = $derived.by<Workflow[]>(() => {
		return workflows
			.filter((task) => task.roles.some((role) => roleSet.has(role)))
			.toSorted((firstTask, secondTask) => workflowPriority(firstTask, launcherRole) - workflowPriority(secondTask, launcherRole))
			.slice(0, 4);
	});
	const visibleSecondaryLinks = $derived(secondaryLinks.filter((link) => link.roles.some((role) => roleSet.has(role))));
	const roleName = $derived(launcherRole ? roleCopy[launcherRole].name : 'Peran ini');
	const roleDescription = $derived(launcherRole ? roleCopy[launcherRole].description : 'Belum ada pintasan ujian untuk peran aktif ini. Gunakan menu utama sesuai tugas masing-masing.');
	const primaryHref = $derived(launcherRole === 'admin' ? '/asesmen/kegiatan' : '/asesmen/pelaksanaan');
	const primaryLabel = $derived(launcherRole === 'admin' ? 'Buat Kegiatan' : 'Buka Monitor');

	function resolveLauncherRole(roleSetValue: ReadonlySet<KnownRole>): LauncherRole | undefined {
		if (roleSetValue.has('admin')) return 'admin';
		if (roleSetValue.has('guru')) return 'guru';
		if (roleSetValue.has('staf')) return 'staf';
		return undefined;
	}

	function workflowPriority(task: Workflow, role: LauncherRole | undefined): number {
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

<div class="space-y-5">
	<PageHeader
		eyebrow="Hari Ini / Dashboard CBT"
		title="Asesmen CBT"
		subtitle={`${roleName}: ${roleDescription}`}
		context="MTsN 2 Kolaka Utara"
		primaryAction={{ label: primaryLabel, href: primaryHref }}
		secondaryAction={{ label: 'Aplikasi Siswa', href: resolve('/asesmen/aplikasi-siswa') }}
	/>

	<ContextStrip
		items={[
			{ label: 'Fokus', value: 'Hari Ini', tone: 'success' },
			{ label: 'Alur', value: 'Persiapan → Monitor → Hasil → Arsip' },
			{ label: 'Peran', value: roleName, tone: 'muted' }
		]}
	/>

	<section class="grid gap-3 md:grid-cols-3" aria-label="Ringkasan operasional CBT">
		<MetricCard label="Ujian hari ini" value="Cek" helper="Buka Monitor untuk melihat sesi aktif atau yang akan berjalan." tone="success" />
		<MetricCard label="Sesi berjalan" value="Monitor" helper="Pengawas melihat peserta bermasalah lebih dulu pada hari-H." tone="warning" />
		<MetricCard label="Perlu tindakan" value="Kesiapan" helper="Paket, peserta, ruang, token, BA, dan arsip dibaca per kegiatan." />
	</section>

	<section aria-labelledby="cbt-tasks-title" class="space-y-4">
		<div class="flex flex-wrap items-end justify-between gap-3">
			<div class="space-y-1">
				<p class="text-sm font-semibold uppercase tracking-[0.18em] text-primary">Alur kerja</p>
				<h2 id="cbt-tasks-title" class="text-xl font-semibold tracking-tight text-foreground">Mulai dari kebutuhan operasional</h2>
			</div>
			<p class="text-sm text-muted-foreground">Fitur teknis tetap tersedia dari halaman detail.</p>
		</div>

		{#if visibleWorkflows.length > 0}
			<div class="grid auto-rows-fr gap-3 md:grid-cols-2 xl:grid-cols-4">
				{#each visibleWorkflows as task (task.title)}
					<WorkflowCard
						title={task.title}
						description={task.description}
						href={resolve(task.href)}
						status={task.status}
						actionLabel={task.actionLabel}
					/>
				{/each}
			</div>
		{:else}
			<section class="rounded-lg border border-dashed border-border bg-muted/50 p-5">
				<h2 class="text-lg font-semibold text-foreground">Tidak ada tugas CBT untuk peran ini</h2>
				<p class="mt-2 text-sm text-muted-foreground">Halaman ini tidak membuka modul yang tidak relevan dengan peran aktif.</p>
				<Button href={resolve('/')} variant="outline" class="mt-4">Kembali ke Beranda</Button>
			</section>
		{/if}
	</section>

	{#if visibleSecondaryLinks.length > 0}
		<nav aria-label="Aksi cepat CBT" class="flex flex-wrap items-center gap-2 rounded-lg border border-border bg-card p-3 text-sm shadow-sm">
			<span class="font-medium text-muted-foreground">Aksi cepat:</span>
			{#each visibleSecondaryLinks as link (link.href)}
				<a href={resolve(link.href)} class="rounded-md border border-border px-3 py-1.5 font-medium text-foreground hover:border-primary/30 hover:bg-primary/10">{link.label}</a>
			{/each}
		</nav>
	{/if}
</div>
