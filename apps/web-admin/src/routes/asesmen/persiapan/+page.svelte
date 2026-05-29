<script lang="ts">
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import * as Card from '$lib/components/ui/card';

	type PersiapanRoute = '/asesmen/paket' | '/asesmen/kegiatan' | '/asesmen/sesi' | '/asesmen';
	type PreparationStep = {
		step: string;
		title: string;
		description: string;
		href: PersiapanRoute;
		cta: string;
		note: string;
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

	const preparationSteps: PreparationStep[] = [
		{
			step: '01',
			title: 'Kegiatan',
			description: 'Pilih konteks ujian terlebih dulu: simulasi, gladi, UAS, atau ujian madrasah.',
			href: '/asesmen/kegiatan',
			cta: 'Buka kegiatan',
			note: 'Mulai dari konteks ujian.'
		},
		{
			step: '02',
			title: 'Paket asesmen',
			description: 'Ambil paket dari Bank Soal yang sudah siap dipakai di kegiatan.',
			href: '/asesmen/paket',
			cta: 'Buka paket',
			note: 'Soal tetap disusun di Bank Soal.'
		},
		{
			step: '03',
			title: 'Sesi & ruang',
			description: 'Atur jadwal, peserta, token, ruang, kapasitas, dan pembagian peserta per sesi.',
			href: '/asesmen/sesi',
			cta: 'Buka sesi',
			note: 'Ruang mengikuti data per sesi.'
		}
	];
</script>

<svelte:head>
	<title>Persiapan Ujian — MTsN 2 Kolaka Utara</title>
</svelte:head>

{#if canAccess}
	<div class="space-y-5">
		<section class="rounded-2xl border border-border bg-card p-4 shadow-sm">
			<div class="space-y-2">
				<Badge class="border-primary/20 bg-primary/10 text-primary" variant="outline">Ujian Digital · Persiapan</Badge>
				<h1 class="text-2xl font-semibold tracking-tight text-foreground">Persiapan Ujian</h1>
				<p class="max-w-2xl text-sm leading-6 text-muted-foreground">
					Kerjakan berurutan: kegiatan, paket, lalu sesi dan ruang. Satu langkah selesai, lanjut ke langkah berikutnya.
				</p>
			</div>
		</section>

		<section aria-labelledby="persiapan-area-title" class="space-y-3">
			<div>
				<p class="text-xs font-semibold uppercase tracking-[0.22em] text-primary">Alur ringkas</p>
				<h2 id="persiapan-area-title" class="mt-1 text-xl font-semibold tracking-tight text-foreground">Langkah persiapan</h2>
			</div>

			<ol class="space-y-3">
				{#each preparationSteps as step}
					<li>
						<Card.Root class="border-border shadow-sm">
							<Card.Content class="p-4">
								<div class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
									<div class="flex min-w-0 gap-3">
										<span class="inline-flex h-9 min-w-9 items-center justify-center rounded-md border border-primary/20 bg-primary/10 px-2 text-xs font-semibold text-primary">{step.step}</span>
										<div class="min-w-0 space-y-1">
											<div class="flex flex-wrap items-center gap-2">
												<Card.Title class="text-base">{step.title}</Card.Title>
												<Badge variant="outline" class="bg-card text-[11px]">{step.note}</Badge>
											</div>
											<Card.Description class="max-w-2xl leading-6">{step.description}</Card.Description>
										</div>
									</div>
									<Button href={resolve(step.href)} variant="outline" size="sm" class="shrink-0 border-primary/20 text-primary hover:bg-primary/10">{step.cta}</Button>
								</div>
							</Card.Content>
						</Card.Root>
					</li>
				{/each}
			</ol>
		</section>

		<section class="rounded-lg border border-primary/20 bg-primary/5 p-4 text-sm leading-6 text-muted-foreground">
			<p class="font-medium text-foreground">Pembagian ruang mengikuti data sesi.</p>
			<p class="mt-1">
				Tambahkan ruang dari detail sesi, isi kapasitas sesuai master ruangan, lalu gunakan pembagian otomatis atau manual dari area Ruangan.
			</p>
		</section>

		<p class="rounded-lg border border-dashed border-border bg-muted/40 px-3 py-2 text-xs text-muted-foreground">
			Catatan: penyusunan soal tetap berada di modul Bank Soal. Simulasi, gladi, dan ujian nyata tetap memakai data kegiatan, paket, dan sesi server.
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
