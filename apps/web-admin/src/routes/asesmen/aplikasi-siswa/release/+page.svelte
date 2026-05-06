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
	<section class="rounded-3xl border border-primary/20 bg-gradient-to-br from-primary/10 via-card to-primary/10 p-6 shadow-sm">
		<div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
			<div class="max-w-3xl space-y-3">
				<p class="text-xs font-semibold uppercase tracking-[0.24em] text-primary">Perangkat & Kesiapan · Monitoring</p>
				<h1 class="text-3xl font-semibold tracking-tight text-foreground">Checklist Rilis CBT Mobile BYOD</h1>
				<p class="max-w-2xl text-sm leading-6 text-muted-foreground">
					Bagian pendukung sebelum backend exam atau APK mobile dipakai di gelombang berikutnya.
					Monitoring hari-H tetap dimulai dari sesi aktif dan dashboard ruang.
				</p>
			</div>
			<div class="flex flex-wrap gap-3">
				<Button href="/asesmen/aplikasi-siswa">Kembali ke Monitoring</Button>
				<Button href="/asesmen/pengawasan" variant="outline">Dashboard Ruang</Button>
				<Button href="/asesmen/aplikasi-siswa/matrix" variant="outline">Buka Matriks Perangkat</Button>
			</div>
		</div>
	</section>

	<Card.Root class="border-primary/20 bg-primary/10 shadow-sm">
		<Card.Content class="grid gap-4 pt-6 md:grid-cols-3">
			<div>
				<p class="text-sm font-semibold text-primary">Pantau Ujian</p>
				<p class="mt-1 text-sm leading-6 text-muted-foreground">Hari-H tetap diarahkan ke Monitoring, sesi hari ini, dan dashboard ruang.</p>
			</div>
			<div>
				<p class="text-sm font-semibold text-primary">Panduan BYOD</p>
				<p class="mt-1 text-sm leading-6 text-muted-foreground">Status guide dan checklist submit dipakai saat pengawas membaca kondisi siswa.</p>
			</div>
			<div>
				<p class="text-sm font-semibold text-primary">Perangkat & Kesiapan</p>
				<p class="mt-1 text-sm leading-6 text-muted-foreground">Halaman ini fokus pada readiness backend, APK, operator, dan artefak rollout.</p>
			</div>
		</Card.Content>
	</Card.Root>

	<div class="grid gap-6 xl:grid-cols-3">
		<Card.Root class="border-border shadow-sm">
			<Card.Header>
				<Card.Title class="text-lg text-foreground">Checklist Backend</Card.Title>
				<Card.Description>
					Pastikan perubahan backend exam tidak diam-diam merusak kontrak yang dipakai Flutter.
				</Card.Description>
			</Card.Header>
			<Card.Content>
				<ul class="space-y-3">
					{#each backendChecks as item (item)}
						<li class="flex gap-3 rounded-2xl border border-border bg-muted/50 px-4 py-3 text-sm leading-6 text-foreground">
							<span class="mt-0.5 inline-flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-primary/15 text-xs font-semibold text-primary">API</span>
							<span>{item}</span>
						</li>
					{/each}
				</ul>
			</Card.Content>
		</Card.Root>

		<Card.Root class="border-border shadow-sm">
			<Card.Header>
				<Card.Title class="text-lg text-foreground">Checklist Mobile</Card.Title>
				<Card.Description>
					Pastikan app Flutter dan uji perangkat inti masih sehat sebelum APK diteruskan ke lapangan.
				</Card.Description>
			</Card.Header>
			<Card.Content>
				<ul class="space-y-3">
					{#each mobileChecks as item (item)}
						<li class="flex gap-3 rounded-2xl border border-border bg-muted/50 px-4 py-3 text-sm leading-6 text-foreground">
							<span class="mt-0.5 inline-flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-primary/15 text-xs font-semibold text-primary">APP</span>
							<span>{item}</span>
						</li>
					{/each}
				</ul>
			</Card.Content>
		</Card.Root>

		<Card.Root class="border-border shadow-sm">
			<Card.Header>
				<Card.Title class="text-lg text-foreground">Checklist Rollout</Card.Title>
				<Card.Description>
					Pastikan operator, pengawas, dan artefak trial sudah benar-benar siap dipakai di gelombang uji berikutnya.
				</Card.Description>
			</Card.Header>
			<Card.Content>
				<ul class="space-y-3">
					{#each rolloutChecks as item (item)}
						<li class="flex gap-3 rounded-2xl border border-border bg-muted/50 px-4 py-3 text-sm leading-6 text-foreground">
							<span class="mt-0.5 inline-flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-primary/15 text-xs font-semibold text-primary">OPS</span>
							<span>{item}</span>
						</li>
					{/each}
				</ul>
			</Card.Content>
		</Card.Root>
	</div>

	<Card.Root class="border-border shadow-sm">
		<Card.Header>
			<Card.Title class="text-lg text-foreground">Artefak Rilis yang Harus Dicek</Card.Title>
			<Card.Description>
				Buka dokumen-dokumen ini sebelum menyatakan perubahan backend atau APK mobile aman untuk gelombang BYOD berikutnya.
			</Card.Description>
		</Card.Header>
		<Card.Content class="grid gap-4 lg:grid-cols-2">
			{#each releaseArtifacts as item (item.path)}
				<div class="rounded-2xl border border-border bg-card p-4">
					<div class="flex items-center gap-3">
						<Badge class="border-primary/20 bg-primary/10 text-primary">Dokumen</Badge>
						<p class="text-sm font-semibold text-foreground">{item.title}</p>
					</div>
					<p class="mt-3 text-sm leading-6 text-muted-foreground">{item.description}</p>
					<p class="mt-3 font-mono text-xs text-muted-foreground">{item.path}</p>
				</div>
			{/each}
		</Card.Content>
	</Card.Root>
</div>
