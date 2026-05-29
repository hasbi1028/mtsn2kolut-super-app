<script lang="ts">
	import { page } from '$app/state';
	import { AssessmentPhaseHeader, AssessmentTaskCard } from '$lib/components/asesmen';

	type ResultRoute =
		| '/asesmen'
		| '/asesmen/kegiatan'
		| '/asesmen/sesi'
		| '/asesmen/pelaksanaan'
		| '/asesmen/hasil'
		| '/asesmen/persiapan';

	type ResultTask = {
		id: string;
		code: string;
		title: string;
		helper: string;
		status: string;
		coverage: string;
		href?: ResultRoute;
		action?: string;
	};

	const userRoles = $derived(page.data.user?.roles ?? (page.data.user?.role ? [page.data.user.role] : []));
	const userPermissions = $derived((page.data.user?.permissions ?? []).map((permission) => permission.trim()).filter(Boolean));
	const isOperator = $derived(
		userPermissions.includes('asesmen.operator') ||
		userPermissions.includes('asesmen.event_manage') ||
		userPermissions.includes('asesmen.package_manage') ||
		userPermissions.includes('asesmen.session_manage')
	);
	const isResultReader = $derived(userPermissions.includes('asesmen.result_read') || userPermissions.includes('asesmen.result_manage'));
	const isProctor = $derived(userPermissions.includes('asesmen.proctor'));
	const canAccess = $derived(isResultReader || userRoles.includes('admin'));
	const canOpenDashboard = $derived(
		userRoles.includes('admin')
			|| userPermissions.includes('asesmen.operator')
			|| userPermissions.includes('asesmen.event_manage')
			|| userPermissions.includes('asesmen.package_manage')
			|| userPermissions.includes('asesmen.session_manage')
	);
	const resultTasks = $derived<ResultTask[]>(
		isOperator
			? [
				{
					id: 'event-results',
					code: '7.3.1',
					title: 'Rekap Nilai',
					helper: 'Masuk dari daftar kegiatan untuk membaca rekap, detail sesi, BA, analisis, dan arsip final dari satu konteks kegiatan.',
					status: 'Pintu utama operator',
					coverage: 'Rekap · BA · analisis · arsip',
					href: '/asesmen/kegiatan',
					action: 'Buka kegiatan'
				},
				{
					id: 'session-results',
					code: '7.3.2',
					title: 'Status Submit',
					helper: 'Pilih sesi untuk melihat kiriman peserta, BA sesi, nilai, dan analisis butir.',
					status: 'Dipakai saat verifikasi',
					coverage: 'Per sesi · BA · analisis soal',
					href: '/asesmen/sesi',
					action: 'Buka sesi'
				},
				{
					id: 'essay-scoring',
					code: '7.3.3',
					title: 'Koreksi Uraian',
					helper: 'Buka sesi atau kegiatan yang memiliki soal uraian untuk melengkapi penilaian manual.',
					status: 'Jika ada uraian',
					coverage: 'Uraian · koreksi manual',
					href: '/asesmen/sesi',
					action: 'Cek sesi'
				},
				{
					id: 'item-analysis',
					code: '7.3.4',
					title: 'Analisis Butir',
					helper: 'Gunakan setelah peserta submit untuk melihat butir yang mudah, sulit, atau perlu ditinjau ulang.',
					status: 'Setelah submit',
					coverage: 'Mutu soal · tindak lanjut',
					href: '/asesmen/sesi',
					action: 'Buka analisis'
				},
				{
					id: 'publish-sync',
					code: '7.3.5',
					title: 'Publikasi / Sinkronisasi',
					helper: 'Lakukan setelah rekap, status submit, dan koreksi uraian sudah selesai diverifikasi.',
					status: 'Langkah akhir',
					coverage: 'Publikasi · rapor',
					href: '/asesmen/kegiatan',
					action: 'Buka kegiatan'
				}
			]
			: [
				{
					id: 'reader-results',
					code: '7.3.1',
					title: 'Rekap Hasil Kegiatan',
					helper: 'Baca hasil yang sudah dibuka oleh operator/panitia. Jika perlu detail kegiatan, minta operator membuka halaman lengkap.',
					status: isResultReader ? 'Baca hasil tersedia' : 'Ikuti arahan operator',
					coverage: 'Rekap akhir · status hasil',
				},
				{
					id: 'reader-monitoring',
					code: '7.2',
					title: 'Pantau Pelaksanaan',
					helper: 'Gunakan bila hasil perlu dicocokkan dengan ruang aktif, status kiriman, atau kejadian pengawasan.',
					status: isProctor ? 'Mode ruang aktif' : 'Koordinasi pengawas',
					coverage: 'Ruang berjalan · kejadian · status kiriman',
					href: isProctor ? '/asesmen/pelaksanaan' : undefined,
					action: isProctor ? 'Buka Pelaksanaan' : undefined,
				},
				{
					id: 'reader-followup',
					code: '7.4',
					title: 'Tindak Lanjut Operator',
					helper: 'Jika butuh BA sesi, analisis butir, atau arsip final, lanjutkan lewat operator/panitia.',
					status: 'Koordinasi panitia',
					coverage: 'BA sesi · analisis · arsip final',
				}
			]
	);

	function taskHref(task: ResultTask) {
		return task.href ?? '';
	}
</script>

<svelte:head>
	<title>Hasil & Penutupan Ujian — MTsN 2 Kolaka Utara</title>
</svelte:head>

{#if canAccess}
	<div class="space-y-5">
		<AssessmentPhaseHeader
			code="7.3"
			badge={isOperator ? 'Mode hasil panitia' : 'Mode baca hasil'}
			title="Hasil & Penutupan Kegiatan"
			description="Pilih satu alur hasil. Rekap detail tetap ada di halaman kegiatan, sesi, dan arsip; halaman ini hanya menjadi pintu kerja hasil."
			primaryAction={{ label: isOperator ? 'Buka Rekap Nilai' : 'Lihat Rekap', href: isOperator ? '/asesmen/kegiatan' : '/asesmen/hasil' }}
			secondaryActions={canOpenDashboard ? [{ label: 'Ringkasan', href: '/asesmen', variant: 'outline' }] : []}
		/>

		<section aria-labelledby="hasil-area-title" class="space-y-3">
			<div>
				<p class="text-xs font-semibold uppercase tracking-[0.22em] text-primary">7.3 Alur Hasil</p>
				<h2 id="hasil-area-title" class="mt-1 text-xl font-semibold tracking-tight text-foreground">Langkah hasil</h2>
			</div>
			<div class="grid gap-3 lg:grid-cols-2">
				{#each resultTasks as task (task.id)}
					<AssessmentTaskCard
						code={task.code}
						title={task.title}
						description={task.helper}
						meta={`${task.status} · ${task.coverage}`}
						href={taskHref(task)}
						cta={task.action ?? ''}
					/>
				{/each}
			</div>
		</section>

		<details class="rounded-lg border border-border bg-muted/40 p-3 text-sm leading-6 text-muted-foreground">
			<summary class="cursor-pointer font-semibold text-foreground">Catatan teknis hasil</summary>
			<p class="mt-2">
				Koreksi uraian, analisis butir, dan sinkronisasi dilakukan setelah status submit peserta jelas. Berita acara dan rekap pelaksanaan masuk area 7.4 Arsip.
			</p>
		</details>
	</div>
{:else}
	<div class="flex min-h-[60vh] flex-col items-center justify-center gap-4 px-4 py-16 text-center">
		<div class="max-w-lg rounded-2xl border border-border bg-card p-8 shadow-sm">
			<h2 class="text-xl font-semibold text-foreground">Akses terbatas</h2>
			<p class="mt-3 text-sm leading-6 text-muted-foreground">
				Halaman hasil hanya tersedia untuk akun yang diberi akses baca hasil asesmen.
			</p>
			<div class="mt-6">
				<a class="inline-flex items-center rounded-md px-1 text-sm font-medium text-muted-foreground underline-offset-4 hover:underline" href="/asesmen">Ringkasan</a>
			</div>
		</div>
	</div>
{/if}
