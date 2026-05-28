<script lang="ts">
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { MicroActionTable } from '$lib/components/ops';
	import RoomAssignmentPanel from '$lib/asesmen/RoomAssignmentPanel.svelte';

	type PersiapanRoute =
		| '/asesmen/paket'
		| '/asesmen/paket/new'
		| '/asesmen/kegiatan'
		| '/asesmen/kegiatan/new'
		| '/asesmen/sesi'
		| '/asesmen/sesi/new'
		| '/asesmen/pelaksanaan'
		| '/asesmen/hasil'
		| '/asesmen';

	type PreparationTask = {
		step: string;
		title: string;
		description: string;
		href: PersiapanRoute;
		cta: string;
	};
	const userRoles = $derived(page.data.user?.roles ?? (page.data.user?.role ? [page.data.user.role] : []));
	const userPermissions = $derived(page.data.user?.permissions ?? []);
	const canAccess = $derived(
		userRoles.includes('admin')
			|| userPermissions.includes('asesmen.operator')
			|| userPermissions.includes('asesmen.event_manage')
			|| userPermissions.includes('asesmen.package_manage')
			|| userPermissions.includes('asesmen.session_manage')
			|| userPermissions.includes('asesmen.participant_manage')
	);
	const canManageRoomAssignment = $derived(userRoles.includes('admin'));

	const adminTasks: PreparationTask[] = [
		{
			step: '01',
			title: 'Pilih Kegiatan',
			description: 'Buat atau pilih kegiatan sebagai konteks ujian.',
			href: '/asesmen/kegiatan',
			cta: 'Kelola Kegiatan'
		},
		{
			step: '02',
			title: 'Siapkan Paket Asesmen',
			description: 'Pilih paket asesmen dari soal Bank Soal yang sudah siap pakai. Pembuatan dan verifikasi soal tetap dilakukan di modul Bank Soal.',
			href: '/asesmen/paket',
			cta: 'Kelola Paket'
		},
		{
			step: '03',
			title: 'Atur Sesi & Token',
			description: 'Tetapkan jadwal, ruang, peserta, dan token.',
			href: '/asesmen/sesi/new',
			cta: 'Atur Sesi'
		},
		{
			step: '04',
			title: 'Masuk Pelaksanaan',
			description: 'Pantau ruang ujian saat paket dan sesi sudah siap.',
			href: '/asesmen/pelaksanaan',
			cta: 'Ke Pelaksanaan'
		}
	];

	const tasks = $derived(adminTasks);

	const taskColumns = [
		{ key: 'step', label: '#', class: 'w-20' },
		{ key: 'task', label: 'Pekerjaan', class: 'min-w-[16rem]' },
		{ key: 'description', label: 'Catatan', class: 'min-w-[20rem]' }
	];


</script>

<svelte:head>
	<title>Persiapan Ujian — MTsN 2 Kolaka Utara</title>
</svelte:head>

{#if canAccess}
	<div class="space-y-6">
	<section class="rounded-2xl border border-border bg-card p-4 shadow-sm">
		<div class="flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between">
			<div class="max-w-3xl space-y-2">
				<Badge class="border-primary/20 bg-primary/10 text-primary" variant="outline">Ujian Digital · Persiapan</Badge>
				<h1 class="text-2xl font-semibold tracking-tight text-foreground">Persiapan Ujian</h1>
				<p class="max-w-2xl text-sm leading-6 text-muted-foreground">
					Siapkan kegiatan, paket, sesi, peserta, ruang, dan kode akses sebelum hari ujian.
				</p>
			</div>
			<div class="flex flex-wrap gap-2">
				<Button href={resolve('/asesmen')} variant="outline" size="sm">Dashboard Asesmen</Button>
			</div>
		</div>
	</section>

	<section aria-labelledby="persiapan-tasks-title" class="space-y-4">
		<div>
			<p class="text-xs font-semibold uppercase tracking-[0.22em] text-primary">Daftar tugas</p>
			<h2 id="persiapan-tasks-title" class="mt-1 text-xl font-semibold tracking-tight text-foreground">Checklist Persiapan</h2>
		</div>

		<MicroActionTable
			title="Checklist Persiapan"
			description="Pilih pekerjaan yang perlu diselesaikan."
			columns={taskColumns}
			rows={tasks}
			rowKey={(row) => (row as PreparationTask).href}
			tableClass="min-w-[760px]"
		>
			{#snippet cell(row, column)}
				{@const task = row as PreparationTask}
				{#if column.key === 'step'}
					<span class="inline-flex h-7 min-w-7 items-center justify-center rounded-md border border-primary/20 bg-primary/10 px-2 text-[11px] font-semibold text-primary">{task.step}</span>
				{:else if column.key === 'task'}
					<p class="font-medium text-foreground">{task.title}</p>
				{:else if column.key === 'description'}
					<p class="max-w-2xl leading-5 text-muted-foreground">{task.description}</p>
				{/if}
			{/snippet}
			{#snippet actions(row)}
				{@const task = row as PreparationTask}
				<Button href={resolve(task.href)} variant="outline" size="xs" class="border-primary/20 text-primary hover:bg-primary/10">{task.cta}</Button>
			{/snippet}
			{#snippet mobile(row)}
				{@const task = row as PreparationTask}
				<div class="space-y-2 text-xs">
					<div class="flex items-start gap-3">
						<span class="inline-flex h-7 min-w-7 shrink-0 items-center justify-center rounded-md border border-primary/20 bg-primary/10 px-2 text-[11px] font-semibold text-primary">{task.step}</span>
						<div class="min-w-0 flex-1">
							<p class="font-medium text-foreground">{task.title}</p>
							<p class="mt-1 leading-5 text-muted-foreground">{task.description}</p>
						</div>
					</div>
					<div class="flex justify-end pt-1">
						<Button href={resolve(task.href)} variant="outline" size="xs" class="border-primary/20 text-primary hover:bg-primary/10">{task.cta}</Button>
					</div>
				</div>
			{/snippet}
		</MicroActionTable>
	</section>

	{#if canManageRoomAssignment}
		<RoomAssignmentPanel />
	{/if}

	<p class="rounded-lg border border-dashed border-border bg-muted/40 px-3 py-2 text-xs text-muted-foreground">
		Catatan: penyusunan soal tetap berada di modul Bank Soal. DEMO hanya contoh lokal; Simulasi/Gladi/Ujian nyata harus memakai kegiatan, paket, dan sesi server.
	</p>
</div>
{:else}
	<div class="flex min-h-[60vh] flex-col items-center justify-center gap-4 px-4 py-16 text-center">
		<div class="max-w-lg rounded-2xl border border-border bg-card p-8 shadow-sm">
			<h2 class="text-xl font-semibold text-foreground">Akses terbatas</h2>
			<p class="mt-3 text-sm leading-6 text-muted-foreground">
				Fase persiapan ujian hanya tersedia untuk panitia/operator. Silakan kembali ke Dashboard Asesmen atau gunakan menu asesmen lain sesuai tugas.
			</p>
			<div class="mt-6">
				<Button href={resolve('/asesmen')} variant="outline">Dashboard Asesmen</Button>
			</div>
		</div>
	</div>
{/if}
