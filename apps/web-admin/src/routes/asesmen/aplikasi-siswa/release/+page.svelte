<script lang="ts">
	import { onMount } from 'svelte';
	import * as Card from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';

	type MobileReleaseManifest = {
		app_name: string;
		channel: string;
		platform: string;
		abi: string;
		version_name: string;
		version_code: number;
		commit: string;
		build_time: string;
		published_at: string;
		file_name: string;
		download_url: string;
		absolute_download_url: string;
		archive_url: string;
		checksum_url: string;
		qr_url: string;
		size_bytes: number;
		sha256: string;
		server_url: string;
		notes: string[];
	};

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
			title: 'Mobile APK Release Center',
			path: 'docs/mobile-apk-release-center.md',
			description: 'Runbook publish APK terbaru ke endpoint SvelteKit/cloudflared tanpa rebuild web.'
		},
		{
			title: 'Operator Quick Start',
			path: 'apps/mobile/OPERATOR_QUICKSTART.md',
			description: 'Panduan singkat pengawas saat siswa mulai ujian dan ketika koneksi mulai bermasalah.'
		}
	];

	let release = $state<MobileReleaseManifest | null>(null);
	let releaseError = $state('');
	let copied = $state(false);

	onMount(async () => {
		try {
			const response = await fetch('/releases/mobile/latest.json', { cache: 'no-store' });
			if (!response.ok) throw new Error(`HTTP ${response.status}`);
			release = await response.json();
		} catch (error) {
			releaseError = error instanceof Error ? error.message : 'Manifest rilis belum tersedia.';
		}
	});

	function formatBytes(bytes: number) {
		if (!Number.isFinite(bytes) || bytes <= 0) return '-';
		const units = ['B', 'KB', 'MB', 'GB'];
		let value = bytes;
		let unit = 0;
		while (value >= 1024 && unit < units.length - 1) {
			value /= 1024;
			unit += 1;
		}
		return `${value.toFixed(unit === 0 ? 0 : 1)} ${units[unit]}`;
	}

	function formatDate(value: string) {
		const date = new Date(value);
		if (Number.isNaN(date.getTime())) return value || '-';
		return new Intl.DateTimeFormat('id-ID', {
			dateStyle: 'medium',
			timeStyle: 'short',
			timeZone: 'Asia/Makassar'
		}).format(date);
	}

	async function copyDownloadLink() {
		if (!release) return;
		await navigator.clipboard?.writeText(release.absolute_download_url ?? release.download_url);
		copied = true;
		setTimeout(() => (copied = false), 1800);
	}
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

	<Card.Root class="overflow-hidden border-primary/20 shadow-sm">
		<Card.Header class="bg-gradient-to-r from-primary/10 via-card to-card">
			<div class="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
				<div>
					<div class="flex flex-wrap items-center gap-2">
						<Badge class="border-primary/20 bg-primary/10 text-primary">APK Resmi</Badge>
						<Badge class="border-emerald-500/20 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300">Production</Badge>
						<Badge class="border-amber-500/20 bg-amber-500/10 text-amber-700 dark:text-amber-300">Anti-cheat aktif</Badge>
					</div>
					<Card.Title class="mt-3 text-xl text-foreground">Download APK CBT Mobile Terbaru</Card.Title>
					<Card.Description>
						Data diambil otomatis dari manifest server. Jika APK baru dipublish, info dan link ini ikut berubah tanpa edit halaman.
					</Card.Description>
				</div>
				{#if release}
					<div class="rounded-2xl border border-border bg-background/80 px-4 py-3 text-sm">
						<p class="text-muted-foreground">Commit</p>
						<p class="font-mono font-semibold text-foreground">{release.commit}</p>
					</div>
				{/if}
			</div>
		</Card.Header>
		<Card.Content class="grid gap-6 pt-6 lg:grid-cols-[1fr_220px]">
			{#if release}
				<div class="space-y-5">
					<div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
						<div class="rounded-2xl border border-border bg-muted/40 p-4">
							<p class="text-xs uppercase tracking-wide text-muted-foreground">Nama App</p>
							<p class="mt-1 text-sm font-semibold text-foreground">{release.app_name}</p>
						</div>
						<div class="rounded-2xl border border-border bg-muted/40 p-4">
							<p class="text-xs uppercase tracking-wide text-muted-foreground">Versi</p>
							<p class="mt-1 text-sm font-semibold text-foreground">v{release.version_name} · code {release.version_code}</p>
						</div>
						<div class="rounded-2xl border border-border bg-muted/40 p-4">
							<p class="text-xs uppercase tracking-wide text-muted-foreground">Ukuran</p>
							<p class="mt-1 text-sm font-semibold text-foreground">{formatBytes(release.size_bytes)}</p>
						</div>
						<div class="rounded-2xl border border-border bg-muted/40 p-4">
							<p class="text-xs uppercase tracking-wide text-muted-foreground">Build</p>
							<p class="mt-1 text-sm font-semibold text-foreground">{formatDate(release.build_time)}</p>
						</div>
					</div>

					<div class="rounded-2xl border border-border bg-muted/30 p-4">
						<p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">SHA256</p>
						<p class="mt-2 break-all font-mono text-xs leading-6 text-foreground">{release.sha256}</p>
					</div>

					<div class="flex flex-wrap gap-3">
						<Button href={release.download_url} class="h-10" download>Download APK Terbaru</Button>
						<Button href={release.checksum_url} class="h-10" variant="outline" download>Download Checksum</Button>
						<Button class="h-10" variant="outline" onclick={copyDownloadLink}>{copied ? 'Link Tersalin' : 'Salin Link'}</Button>
					</div>

					<div class="rounded-2xl border border-amber-500/20 bg-amber-500/10 p-4 text-sm leading-6 text-amber-900 dark:text-amber-100">
						<p class="font-semibold">Sumber resmi APK</p>
						<p>Install hanya dari domain <span class="font-mono">{release.server_url}</span>. Jangan gunakan APK dari sumber lain.</p>
					</div>
				</div>
				<div class="flex flex-col items-center justify-center rounded-2xl border border-border bg-muted/30 p-4 text-center">
					<img src={release.qr_url} alt="QR download APK CBT Mobile" class="h-44 w-44 rounded-xl bg-white p-2" />
					<p class="mt-3 text-sm font-semibold text-foreground">Scan untuk download</p>
					<p class="mt-1 break-all text-xs leading-5 text-muted-foreground">{release.absolute_download_url}</p>
				</div>
			{:else}
				<div class="lg:col-span-2 rounded-2xl border border-dashed border-border bg-muted/30 p-6 text-sm leading-6 text-muted-foreground">
					<p class="font-semibold text-foreground">Manifest APK belum terbaca.</p>
					<p class="mt-1">{releaseError || 'Memuat latest.json dari server...'}</p>
				</div>
			{/if}
		</Card.Content>
	</Card.Root>

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
