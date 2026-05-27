<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import type { RouteId } from '$app/types';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { ContextStrip, MicroActionTable, PageHeader } from '$lib/components/ops';
	import { trackInternalAnalyticsEvent } from '$lib/analytics/internal-analytics';

	type AppRole = 'admin' | 'guru' | 'staf' | 'kesiswaan' | 'siswa' | 'ortu';
	type KnownRole = AppRole | (string & {});
	type CbtRoute = Extract<
		RouteId,
		| '/'
		| '/asesmen/ringkas'
		| '/asesmen/persiapan'
		| '/asesmen/kegiatan'
		| '/asesmen/pelaksanaan'
		| '/asesmen/pengawasan'
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
			description: 'Siapkan ujian, pantau hari-H, lalu buka hasil dari satu alur ringkas.'
		},
		guru: {
			name: 'Guru',
			description: 'Buka tugas pelaksanaan atau hasil jika sudah ditugaskan. Penulisan soal tetap di Bank Soal.'
		},
		staf: {
			name: 'Staf',
			description: 'Fokus membantu pantauan ruang dan perangkat siswa saat ujian.'
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
			title: 'Ringkasan Ujian',
			description: 'Buka halaman kerja ringkas untuk melihat sesi, paket, dan pembagian 8 ruang.',
			href: '/asesmen/ringkas',
			actionLabel: 'Buka Ringkasan',
			status: 'Utama',
			roles: ['admin', 'guru', 'staf'],
			priority: { admin: 1, guru: 1, staf: 1 }
		},
		{
			title: 'Persiapan',
			description: 'Siapkan kegiatan, paket, jadwal, peserta, ruang, token, dan kartu sebelum ujian.',
			href: '/asesmen/persiapan',
			actionLabel: 'Buka Persiapan',
			status: 'Pra ujian',
			roles: ['admin', 'guru'],
			priority: { admin: 2, guru: 2 }
		},
		{
			title: 'Pelaksanaan',
			description: 'Pantau sesi berjalan dan arahkan pengawas ke halaman pantau ruang.',
			href: '/asesmen/pelaksanaan',
			actionLabel: 'Buka Pelaksanaan',
			status: 'Hari-H',
			roles: ['admin', 'guru', 'staf'],
			priority: { admin: 3, guru: 3, staf: 2 }
		},
		{
			title: 'Pantau Ruang',
			description: 'Masuk ke daftar ruang, peserta perlu dibantu, dan catatan pengawasan.',
			href: '/asesmen/pengawasan',
			actionLabel: 'Pantau Ruang',
			status: 'Ruang',
			roles: ['admin', 'guru', 'staf'],
			priority: { admin: 4, guru: 4, staf: 3 }
		},
		{
			title: 'Hasil',
			description: 'Lihat rekap jawaban, koreksi, nilai, berita acara, dan unduhan akhir.',
			href: '/asesmen/hasil',
			actionLabel: 'Buka Hasil',
			status: 'Pasca ujian',
			roles: ['admin', 'guru'],
			priority: { admin: 5, guru: 5 }
		}
	];

	const secondaryLinks: SecondaryLink[] = [
		{ label: 'Persiapan', href: '/asesmen/persiapan', roles: ['admin', 'guru'] },
		{ label: 'Pantau Ruang', href: '/asesmen/pengawasan', roles: ['admin', 'guru', 'staf'] },
		{ label: 'Hasil', href: '/asesmen/hasil', roles: ['admin', 'guru'] }
	];

	const workflowColumns = [
		{ key: 'workflow', label: 'Pekerjaan', class: 'min-w-[15rem]' },
		{ key: 'status', label: 'Fase', class: 'w-32' },
		{ key: 'description', label: 'Catatan', class: 'min-w-[22rem]' }
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
	const primaryHref = $derived(launcherRole === 'admin' ? '/asesmen/ringkas' : '/asesmen/pelaksanaan');
	const primaryLabel = $derived(launcherRole === 'admin' ? 'Buka Ringkasan' : 'Buka Pelaksanaan');

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
		eyebrow="Asesmen / Alur Utama"
		title="Asesmen Ujian"
		subtitle={`${roleName}: ${roleDescription}`}
		context="Persiapan → Pelaksanaan → Pantau Ruang → Hasil"
		primaryAction={{ label: primaryLabel, href: primaryHref }}
	/>

	<ContextStrip
		items={[
			{ label: 'Alur', value: 'Ringkas → Persiapan → Hari-H → Hasil' },
			{ label: 'Peran', value: roleName, tone: 'muted' }
		]}
	/>

	<section aria-labelledby="cbt-tasks-title" class="space-y-4">
		<div class="flex flex-wrap items-end justify-between gap-3">
			<div class="space-y-1">
				<p class="text-sm font-semibold uppercase tracking-[0.18em] text-primary">Alur kerja</p>
				<h2 id="cbt-tasks-title" class="text-xl font-semibold tracking-tight text-foreground">Pilih pekerjaan</h2>
			</div>
		</div>

		{#if visibleWorkflows.length > 0}
			<MicroActionTable
				title="Pekerjaan Asesmen"
				description="Pilih sesuai fase."
				columns={workflowColumns}
				rows={visibleWorkflows}
				rowKey={(row) => (row as Workflow).href}
				tableClass="min-w-[780px]"
			>
				{#snippet cell(row, column)}
					{@const task = row as Workflow}
					{#if column.key === 'workflow'}
						<p class="font-medium text-foreground">{task.title}</p>
					{:else if column.key === 'status'}
						<Badge variant="outline" class="border-primary/20 bg-primary/10 text-xs text-primary">{task.status}</Badge>
					{:else if column.key === 'description'}
						<p class="max-w-2xl leading-5 text-muted-foreground">{task.description}</p>
					{/if}
				{/snippet}
				{#snippet actions(row)}
					{@const task = row as Workflow}
					<Button href={resolve(task.href)} variant="outline" size="xs" class="border-primary/20 text-primary hover:bg-primary/10">{task.actionLabel}</Button>
				{/snippet}
				{#snippet mobile(row)}
					{@const task = row as Workflow}
					<div class="space-y-2 text-xs">
						<div class="flex items-start justify-between gap-3">
							<div>
								<p class="font-medium text-foreground">{task.title}</p>
								<p class="mt-1 leading-5 text-muted-foreground">{task.description}</p>
							</div>
							<Badge variant="outline" class="shrink-0 border-primary/20 bg-primary/10 text-xs text-primary">{task.status}</Badge>
						</div>
						<div class="flex justify-end pt-1">
							<Button href={resolve(task.href)} variant="outline" size="xs" class="border-primary/20 text-primary hover:bg-primary/10">{task.actionLabel}</Button>
						</div>
					</div>
				{/snippet}
			</MicroActionTable>
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
			<span class="font-medium text-muted-foreground">Pintasan:</span>
			{#each visibleSecondaryLinks as link (link.href)}
				<a href={resolve(link.href)} class="rounded-md border border-border px-3 py-1.5 font-medium text-foreground hover:border-primary/30 hover:bg-primary/10">{link.label}</a>
			{/each}
		</nav>
	{/if}
</div>
