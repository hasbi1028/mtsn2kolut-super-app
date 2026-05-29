<script lang="ts">
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import { Button } from '$lib/components/ui/button';
	import { AssessmentPhaseHeader, AssessmentTaskCard } from '$lib/components/asesmen';

	type PersiapanRoute = '/asesmen/paket' | '/asesmen/kegiatan' | '/asesmen/sesi' | '/asesmen' | '/asesmen/sesi#ruang-peserta' | '/asesmen/sesi#pengawas' | '/asesmen/aplikasi-siswa';
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
			step: '7.1.1',
			title: 'Kegiatan Asesmen',
			description: 'Pilih konteks ujian terlebih dulu: simulasi, gladi, UAS, atau ujian madrasah.',
			href: '/asesmen/kegiatan',
			cta: 'Buka kegiatan',
			note: 'Mulai dari konteks ujian.'
		},
		{
			step: '7.1.2',
			title: 'Paket Soal',
			description: 'Ambil paket dari Bank Soal yang sudah siap dipakai di kegiatan.',
			href: '/asesmen/paket',
			cta: 'Buka paket',
			note: 'Soal tetap disusun di Bank Soal.'
		},
		{
			step: '7.1.3',
			title: 'Sesi Ujian',
			description: 'Atur jadwal, durasi, paket, dan cakupan peserta untuk tiap sesi ujian.',
			href: '/asesmen/sesi',
			cta: 'Buka sesi',
			note: 'Jadwal dan peserta.'
		},
		{
			step: '7.1.4',
			title: 'Ruang & Peserta',
			description: 'Bagi peserta ke ruang secara dinamis per siswa, bukan memindahkan rombel sebagai blok utuh.',
			href: '/asesmen/sesi#ruang-peserta',
			cta: 'Atur ruang',
			note: 'Pembagian per siswa.'
		},
		{
			step: '7.1.5',
			title: 'Pengawas Ruang',
			description: 'Tetapkan pengawas utama atau pendamping untuk ruang yang sudah siap.',
			href: '/asesmen/sesi#pengawas',
			cta: 'Atur pengawas',
			note: 'Penugasan ruang.'
		}
	];
</script>

<svelte:head>
	<title>Persiapan Ujian — MTsN 2 Kolaka Utara</title>
</svelte:head>

{#if canAccess}
	<div class="space-y-5">
		<AssessmentPhaseHeader
			code="7.1"
			badge="Asesmen"
			title="Persiapan Ujian"
			description="Checklist ringkas sebelum CBT: pilih kegiatan, siapkan paket, atur sesi-ruang-peserta, cetak kartu/lembar pengawas, lalu masuk pelaksanaan."
			primaryAction={{ label: 'Mulai dari Kegiatan', href: resolve('/asesmen/kegiatan') }}
			secondaryActions={[{ label: 'Command Center', href: resolve('/asesmen'), variant: 'outline' }]}
		/>

		<section aria-labelledby="persiapan-area-title" class="space-y-3">
			<div>
				<p class="text-xs font-semibold uppercase tracking-[0.22em] text-primary">7.1 Alur Persiapan</p>
				<h2 id="persiapan-area-title" class="mt-1 text-xl font-semibold tracking-tight text-foreground">Checklist persiapan CBT</h2>
			</div>

			<ol class="grid gap-3 lg:grid-cols-2">
				{#each preparationSteps as step}
					<li>
						<AssessmentTaskCard code={step.step} title={step.title} description={step.description} meta={step.note} href={resolve(step.href)} cta={step.cta} />
					</li>
				{/each}
			</ol>
		</section>

		<details class="rounded-lg border border-border bg-muted/40 p-3 text-sm leading-6 text-muted-foreground">
			<summary class="cursor-pointer font-semibold text-foreground">Catatan teknis persiapan</summary>
			<p class="mt-2">
				Penyusunan soal tetap berada di Bank Soal. Detail teknis kegiatan, paket, sesi, ruang, kartu peserta, dan lembar pengawas tetap dibuka dari halaman terkait agar sidebar harian tetap sederhana.
			</p>
		</details>
	</div>
{:else}
	<div class="flex min-h-[60vh] flex-col items-center justify-center gap-4 px-4 py-16 text-center">
		<div class="max-w-lg rounded-2xl border border-border bg-card p-8 shadow-sm">
			<h2 class="text-xl font-semibold text-foreground">Akses terbatas</h2>
			<p class="mt-3 text-sm leading-6 text-muted-foreground">
				Fase persiapan ujian hanya tersedia untuk panitia/operator. Silakan kembali ke Command Center CBT atau gunakan menu asesmen lain sesuai tugas.
			</p>
			<div class="mt-6">
				<Button href={resolve('/asesmen')} variant="outline">Command Center CBT</Button>
			</div>
		</div>
	</div>
{/if}
