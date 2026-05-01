<script lang="ts">
	import * as Card from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';

	const backendChecks = [
		'Response login masih memuat field siswa, sesi, ruang, dan progres yang dipakai Flutter.',
		'Perubahan field soal tidak merusak renderer pilihan ganda, uraian, rich text, gambar, dan audio.',
		'`time_remaining_seconds` tetap akurat untuk countdown, restore, dan submit guard.',
		'Perubahan event warning atau error code sudah ditinjau dampaknya ke restore flow mobile.',
		'Perubahan payload dicatat memakai template release note exam payload.'
	];

	const mobileChecks = [
		'`flutter analyze` hijau.',
		'`flutter test` hijau.',
		'Login token berhasil pada APK build terbaru.',
		'Restore sesi masih berjalan pada perangkat uji utama.',
		'Submit berhasil saat koneksi sehat.',
		'Status `Waspada` dan `Menurun` masih muncul sesuai simulasi gangguan.'
	];

	const rolloutChecks = [
		'APK release sudah dibagikan lewat kanal resmi sekolah.',
		'Pengawas sudah membaca Operator Quick Start.',
		'Perangkat uji tercatat di Device Test Matrix.',
		'BYOD trial procedure sudah diikuti untuk gelombang uji yang akan berjalan.',
		'Tidak ada perubahan backend exam yang belum direview terhadap mobile contract.'
	];

	const releaseArtifacts = [
		{
			title: 'Exam API Contract',
			path: 'docs/exam-api.md',
			description: 'Referensi kontrak endpoint exam dan checklist kompatibilitas payload mobile.'
		},
		{
			title: 'Payload Release Template',
			path: 'docs/exam-payload-release-template.md',
			description: 'Template release note setiap kali backend mengubah payload exam yang menyentuh Flutter.'
		},
		{
			title: 'Release Checklist',
			path: 'apps/mobile/RELEASE_CHECKLIST.md',
			description: 'Checklist operasional build, verifikasi, dan distribusi APK internal.'
		},
		{
			title: 'Operator Quick Start',
			path: 'apps/mobile/OPERATOR_QUICKSTART.md',
			description: 'Panduan singkat pengawas saat siswa mulai ujian dan ketika koneksi mulai bermasalah.'
		}
	];
</script>

<svelte:head>
	<title>Readiness Release Mobile — MTsN 2 Kolaka Utara</title>
</svelte:head>

<div class="space-y-6">
	<section class="rounded-3xl border border-sky-100 bg-gradient-to-br from-sky-50 via-white to-emerald-50 p-6 shadow-sm">
		<div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
			<div class="max-w-3xl space-y-3">
				<p class="text-xs font-semibold uppercase tracking-[0.24em] text-sky-700">Readiness Release Mobile</p>
				<h1 class="text-3xl font-semibold tracking-tight text-slate-900">Checklist Rilis CBT Mobile BYOD</h1>
				<p class="max-w-2xl text-sm leading-6 text-slate-600">
					Gunakan halaman ini sebelum backend exam atau APK mobile dirilis ke gelombang uji berikutnya.
					Fokusnya adalah kompatibilitas payload, stabilitas app, dan kesiapan operator lapangan.
				</p>
			</div>
			<div class="flex flex-wrap gap-3">
				<Button href="/cbt/byod">Kembali ke Panduan BYOD</Button>
				<Button href="/cbt/byod/matrix" variant="outline">Buka Matriks Perangkat</Button>
			</div>
		</div>
	</section>

	<div class="grid gap-6 xl:grid-cols-3">
		<Card.Root class="border-slate-200 shadow-sm">
			<Card.Header>
				<Card.Title class="text-lg text-slate-900">Checklist Backend</Card.Title>
				<Card.Description>
					Pastikan perubahan backend exam tidak diam-diam merusak kontrak yang dipakai Flutter.
				</Card.Description>
			</Card.Header>
			<Card.Content>
				<ul class="space-y-3">
					{#each backendChecks as item (item)}
						<li class="flex gap-3 rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3 text-sm leading-6 text-slate-700">
							<span class="mt-0.5 inline-flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-sky-100 text-xs font-semibold text-sky-700">API</span>
							<span>{item}</span>
						</li>
					{/each}
				</ul>
			</Card.Content>
		</Card.Root>

		<Card.Root class="border-slate-200 shadow-sm">
			<Card.Header>
				<Card.Title class="text-lg text-slate-900">Checklist Mobile</Card.Title>
				<Card.Description>
					Pastikan app Flutter dan uji perangkat inti masih sehat sebelum APK diteruskan ke lapangan.
				</Card.Description>
			</Card.Header>
			<Card.Content>
				<ul class="space-y-3">
					{#each mobileChecks as item (item)}
						<li class="flex gap-3 rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3 text-sm leading-6 text-slate-700">
							<span class="mt-0.5 inline-flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-emerald-100 text-xs font-semibold text-emerald-700">APP</span>
							<span>{item}</span>
						</li>
					{/each}
				</ul>
			</Card.Content>
		</Card.Root>

		<Card.Root class="border-slate-200 shadow-sm">
			<Card.Header>
				<Card.Title class="text-lg text-slate-900">Checklist Rollout</Card.Title>
				<Card.Description>
					Pastikan operator, pengawas, dan artefak trial sudah benar-benar siap dipakai di gelombang uji berikutnya.
				</Card.Description>
			</Card.Header>
			<Card.Content>
				<ul class="space-y-3">
					{#each rolloutChecks as item (item)}
						<li class="flex gap-3 rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3 text-sm leading-6 text-slate-700">
							<span class="mt-0.5 inline-flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-sky-100 text-xs font-semibold text-sky-700">OPS</span>
							<span>{item}</span>
						</li>
					{/each}
				</ul>
			</Card.Content>
		</Card.Root>
	</div>

	<Card.Root class="border-slate-200 shadow-sm">
		<Card.Header>
			<Card.Title class="text-lg text-slate-900">Artefak Rilis yang Harus Dicek</Card.Title>
			<Card.Description>
				Buka dokumen-dokumen ini sebelum menyatakan perubahan backend atau APK mobile aman untuk gelombang BYOD berikutnya.
			</Card.Description>
		</Card.Header>
		<Card.Content class="grid gap-4 lg:grid-cols-2">
			{#each releaseArtifacts as item (item.path)}
				<div class="rounded-2xl border border-slate-200 bg-white p-4">
					<div class="flex items-center gap-3">
						<Badge class="border-sky-200 bg-sky-50 text-sky-700">Dokumen</Badge>
						<p class="text-sm font-semibold text-slate-900">{item.title}</p>
					</div>
					<p class="mt-3 text-sm leading-6 text-slate-600">{item.description}</p>
					<p class="mt-3 font-mono text-xs text-slate-500">{item.path}</p>
				</div>
			{/each}
		</Card.Content>
	</Card.Root>
</div>
