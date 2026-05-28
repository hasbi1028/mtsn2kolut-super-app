<script lang="ts">
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import * as Card from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';

	type PersiapanRoute =
		| '/asesmen/paket'
		| '/asesmen/kegiatan'
		| '/asesmen/sesi'
		| '/asesmen';

	type PreparationArea = {
		step: string;
		title: string;
		description: string;
		href: PersiapanRoute;
		cta: string;
		meta: string;
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

	const preparationAreas: PreparationArea[] = [
		{
			step: '01',
			title: 'Kegiatan',
			description: 'Buat atau pilih konteks asesmen seperti simulasi, gladi, UAS, atau ujian madrasah.',
			href: '/asesmen/kegiatan',
			cta: 'Buka Kegiatan',
			meta: 'Konteks ujian'
		},
		{
			step: '02',
			title: 'Paket Asesmen',
			description: 'Pilih paket dari soal Bank Soal yang sudah siap pakai. Penyusunan dan verifikasi soal tetap berada di modul Bank Soal.',
			href: '/asesmen/paket',
			cta: 'Buka Paket',
			meta: 'Materi ujian'
		},
		{
			step: '03',
			title: 'Sesi & Ruang',
			description: 'Atur jadwal, peserta, token, ruang, kapasitas, dan pembagian peserta per sesi. Jumlah ruang mengikuti data ruang yang ditambahkan, bukan angka tetap.',
			href: '/asesmen/sesi',
			cta: 'Buka Sesi',
			meta: 'Jadwal dan ruang'
		}
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
					Siapkan tiga hal inti sebelum hari ujian: kegiatan, paket asesmen, lalu sesi beserta ruangnya.
				</p>
			</div>

		</div>
	</section>

	<section aria-labelledby="persiapan-area-title" class="space-y-4">
		<div>
			<p class="text-xs font-semibold uppercase tracking-[0.22em] text-primary">Alur ringkas</p>
			<h2 id="persiapan-area-title" class="mt-1 text-xl font-semibold tracking-tight text-foreground">Langkah Persiapan</h2>
		</div>

		<div class="grid gap-3 md:grid-cols-3">
			{#each preparationAreas as area (area.href)}
				<Card.Root class="border-border shadow-sm">
					<Card.Header class="space-y-3">
						<div class="flex items-center justify-between gap-3">
							<span class="inline-flex h-8 min-w-8 items-center justify-center rounded-md border border-primary/20 bg-primary/10 px-2 text-xs font-semibold text-primary">{area.step}</span>
							<Badge variant="outline" class="bg-card text-[11px]">{area.meta}</Badge>
						</div>
						<div class="space-y-1">
							<Card.Title class="text-base">{area.title}</Card.Title>
							<Card.Description class="leading-6">{area.description}</Card.Description>
						</div>
					</Card.Header>
					<Card.Content>
						<Button href={resolve(area.href)} variant="outline" size="sm" class="w-full border-primary/20 text-primary hover:bg-primary/10">{area.cta}</Button>
					</Card.Content>
				</Card.Root>
			{/each}
		</div>
	</section>

	<section class="rounded-lg border border-primary/20 bg-primary/5 p-4 text-sm leading-6 text-muted-foreground">
		<p class="font-medium text-foreground">Pembagian ruang bersifat dinamis per sesi.</p>
		<p class="mt-1">
			Tambah ruang dari detail sesi, isi kapasitas sesuai master ruangan, lalu gunakan pembagian otomatis atau manual dari tab Ruangan. Sistem akan membaca jumlah ruang yang tersedia pada sesi tersebut.
		</p>
	</section>

	<p class="rounded-lg border border-dashed border-border bg-muted/40 px-3 py-2 text-xs text-muted-foreground">
		Catatan: penyusunan soal tetap berada di modul Bank Soal. DEMO hanya contoh lokal; Simulasi/Gladi/Ujian nyata harus memakai kegiatan, paket, dan sesi server.
	</p>
</div>
{:else}
	<div class="flex min-h-[60vh] flex-col items-center justify-center gap-4 px-4 py-16 text-center">
		<div class="max-w-lg rounded-2xl border border-border bg-card p-8 shadow-sm">
			<h2 class="text-xl font-semibold text-foreground">Akses terbatas</h2>
			<p class="mt-3 text-sm leading-6 text-muted-foreground">
				Fase persiapan ujian hanya tersedia untuk panitia/operator. Silakan kembali ke Ringkasan Asesmen atau gunakan menu asesmen lain sesuai tugas.
			</p>
			<div class="mt-6">
				<Button href={resolve('/asesmen')} variant="outline">Ringkasan Asesmen</Button>
			</div>
		</div>
	</div>
{/if}
