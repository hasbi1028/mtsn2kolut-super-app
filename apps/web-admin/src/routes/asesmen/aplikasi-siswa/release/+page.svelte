<script lang="ts">
	import { onMount } from 'svelte';
	import * as Card from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';


	type CbtMobileArtifact = {
		fileName: string;
		label: string;
		description: string;
		variant: string;
		abi: string;
		recommendedFor: string;
		sizeBytes: number;
		updatedAt: string;
		sha256: string;
		downloadUrl: string;
		absoluteDownloadUrl: string;
	};

	type CbtMobileArtifactManifest = {
		title: string;
		primaryClient: string;
		fallbackClient: string;
		demoWebUrl: string;
		generatedAt: string;
		artifacts: CbtMobileArtifact[];
	};

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
		'Data masuk ujian masih memuat siswa, sesi, ruang, dan progres yang dipakai aplikasi siswa.',
		'Perubahan data soal tidak mengganggu tampilan pilihan ganda, uraian, teks kaya, gambar, dan audio.',
		'Sisa waktu tetap akurat untuk hitung mundur, pemulihan sesi, dan pengaman kirim ujian.',
		'Perubahan peringatan kegiatan atau kode masalah sudah ditinjau dampaknya ke alur pemulihan aplikasi siswa.',
		'Perubahan data yang dikirim dicatat memakai format catatan rilis data ujian.'
	];

	const mobileChecks = [
		'Pemeriksaan kode aplikasi siswa lulus.',
		'Tes aplikasi siswa lulus.',
		'Masuk dengan kode ujian berhasil pada APK rilis terbaru.',
		'Pemulihan sesi masih berjalan pada perangkat uji utama.',
		'Kirim ujian berhasil saat koneksi sehat.',
		'Status `Waspada` dan `Menurun` masih muncul sesuai simulasi gangguan.'
	];

	const rolloutChecks = [
		'APK rilis sudah dibagikan lewat kanal resmi sekolah.',
		'Pengawas sudah membaca Panduan Cepat Operator.',
		'Perangkat uji tercatat di Tabel Uji Perangkat.',
		'Prosedur uji coba BYOD sudah diikuti untuk gelombang uji yang akan berjalan.',
		'Tidak ada perubahan layanan ujian yang belum diperiksa terhadap kesesuaian aplikasi siswa.'
	];

	const releaseArtifacts = [
		{
			title: 'Kontrak Layanan Ujian',
			path: 'docs/exam-api.md',
			description: 'Referensi kontrak layanan ujian dan daftar pemeriksaan kesesuaian data yang dikirim ke aplikasi siswa.'
		},
		{
			title: 'Format Catatan Rilis Ujian',
			path: 'docs/exam-payload-release-template.md',
			description: 'Format catatan rilis setiap kali layanan sistem mengubah data ujian yang menyentuh aplikasi siswa.'
		},
		{
			title: 'Daftar Pemeriksaan Rilis',
			path: 'apps/mobile/RELEASE_CHECKLIST.md',
			description: 'Daftar pemeriksaan operasional rilis, verifikasi, dan distribusi APK internal.'
		},
		{
			title: 'Pusat Rilis APK Siswa',
			path: 'docs/mobile-apk-release-center.md',
			description: 'Panduan menerbitkan APK terbaru ke alamat layanan sistem tanpa membangun ulang web.'
		},
		{
			title: 'Panduan Cepat Operator',
			path: 'apps/mobile/OPERATOR_QUICKSTART.md',
			description: 'Panduan singkat pengawas saat siswa mulai ujian dan ketika koneksi mulai bermasalah.'
		}
	];

	let release = $state<MobileReleaseManifest | null>(null);
	let cbtArtifacts = $state<CbtMobileArtifactManifest | null>(null);
	let releaseError = $state('');
	let cbtArtifactsError = $state('');
	let copied = $state(false);

	onMount(async () => {
		try {
			const response = await fetch('/releases/mobile/latest.json', { cache: 'no-store' });
			if (!response.ok) throw new Error(`HTTP ${response.status}`);
			release = await response.json();
		} catch (error) {
			releaseError = error instanceof Error ? error.message : 'Data rilis belum tersedia.';
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
	<title>Kesiapan Rilis Aplikasi Siswa — MTsN 2 Kolaka Utara</title>
</svelte:head>

<div class="space-y-6">
	<section class="rounded-3xl border border-primary/20 bg-gradient-to-br from-primary/10 via-card to-primary/10 p-6 shadow-sm">
		<div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
			<div class="max-w-3xl space-y-3">
				<p class="text-xs font-semibold uppercase tracking-[0.24em] text-primary">Perangkat & Kesiapan · Pemantauan</p>
				<h1 class="text-3xl font-semibold tracking-tight text-foreground">Daftar Pemeriksaan Rilis Aplikasi Siswa BYOD</h1>
				<p class="max-w-2xl text-sm leading-6 text-muted-foreground">
					Bagian pendukung sebelum layanan ujian atau APK aplikasi siswa dipakai di gelombang berikutnya.
					Pemantauan hari-H tetap dimulai dari sesi aktif dan panel ruang.
				</p>
			</div>
			<div class="flex flex-wrap gap-3">
				<Button href="/asesmen/aplikasi-siswa">Kembali ke Pemantauan</Button>
				<Button href="/asesmen/pengawasan" variant="outline">Panel Ruang</Button>
				<Button href="/asesmen/aplikasi-siswa/matrix" variant="outline">Buka Tabel Perangkat</Button>
			</div>
		</div>
	</section>

	<Card.Root class="overflow-hidden border-primary/20 shadow-sm">
		<Card.Header class="bg-gradient-to-r from-primary/10 via-card to-card">
			<div class="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
				<div>
					<div class="flex flex-wrap items-center gap-2">
						<Badge class="border-primary/20 bg-primary/10 text-primary">APK Resmi</Badge>
						<Badge class="border-emerald-500/20 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300">Rilis utama</Badge>
						<Badge class="border-amber-500/20 bg-amber-500/10 text-amber-700 dark:text-amber-300">Pengawasan aktif</Badge>
					</div>
					<Card.Title class="mt-3 text-xl text-foreground">Unduh APK Aplikasi Siswa Terbaru</Card.Title>
					<Card.Description>
						Data diambil otomatis dari data rilis layanan sistem. Jika APK baru diterbitkan, info dan tautan ini ikut berubah tanpa edit halaman.
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
							<p class="text-xs uppercase tracking-wide text-muted-foreground">Nama Aplikasi</p>
							<p class="mt-1 text-sm font-semibold text-foreground">{release.app_name}</p>
						</div>
						<div class="rounded-2xl border border-border bg-muted/40 p-4">
							<p class="text-xs uppercase tracking-wide text-muted-foreground">Versi</p>
							<p class="mt-1 text-sm font-semibold text-foreground">v{release.version_name} · kode {release.version_code}</p>
						</div>
						<div class="rounded-2xl border border-border bg-muted/40 p-4">
							<p class="text-xs uppercase tracking-wide text-muted-foreground">Ukuran</p>
							<p class="mt-1 text-sm font-semibold text-foreground">{formatBytes(release.size_bytes)}</p>
						</div>
						<div class="rounded-2xl border border-border bg-muted/40 p-4">
							<p class="text-xs uppercase tracking-wide text-muted-foreground">Dibuat</p>
							<p class="mt-1 text-sm font-semibold text-foreground">{formatDate(release.build_time)}</p>
						</div>
					</div>

					<div class="rounded-2xl border border-border bg-muted/30 p-4">
						<p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">SHA256</p>
						<p class="mt-2 break-all font-mono text-xs leading-6 text-foreground">{release.sha256}</p>
					</div>

					<div class="flex flex-wrap gap-3">
						<Button href={release.download_url} class="h-10" download>Unduh APK Terbaru</Button>
						<Button href={release.checksum_url} class="h-10" variant="outline" download>Unduh Data Pemeriksaan</Button>
						<Button class="h-10" variant="outline" onclick={copyDownloadLink}>{copied ? 'Tautan Tersalin' : 'Salin Tautan'}</Button>
					</div>

					<div class="rounded-2xl border border-amber-500/20 bg-amber-500/10 p-4 text-sm leading-6 text-amber-900 dark:text-amber-100">
						<p class="font-semibold">Sumber resmi APK</p>
						<p>Pasang hanya dari domain <span class="font-mono">{release.server_url}</span>. Jangan gunakan APK dari sumber lain.</p>
					</div>
				</div>
				<div class="flex flex-col items-center justify-center rounded-2xl border border-border bg-muted/30 p-4 text-center">
					<img src={release.qr_url} alt="QR unduh APK CBT Mobile" class="h-44 w-44 rounded-xl bg-white p-2" />
					<p class="mt-3 text-sm font-semibold text-foreground">Pindai untuk unduh</p>
					<p class="mt-1 break-all text-xs leading-5 text-muted-foreground">{release.absolute_download_url}</p>
				</div>
			{:else}
				<div class="lg:col-span-2 rounded-2xl border border-dashed border-border bg-muted/30 p-6 text-sm leading-6 text-muted-foreground">
					<p class="font-semibold text-foreground">Data rilis APK belum terbaca.</p>
					<p class="mt-1">{releaseError || 'Memuat data rilis terbaru dari layanan sistem...'}</p>
				</div>
			{/if}
		</Card.Content>
	</Card.Root>


	<Card.Root class="overflow-hidden border-emerald-500/20 shadow-sm">
		<Card.Header class="bg-gradient-to-r from-emerald-500/10 via-card to-card">
			<div class="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
				<div>
					<div class="flex flex-wrap items-center gap-2">
						<Badge class="border-emerald-500/20 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300">Download Center</Badge>
						<Badge class="border-primary/20 bg-primary/10 text-primary">Flutter Utama</Badge>
						<Badge class="border-amber-500/20 bg-amber-500/10 text-amber-700 dark:text-amber-300">Demo & Browser Darurat</Badge>
					</div>
					<Card.Title class="mt-3 text-xl text-foreground">Download Center APK CBT</Card.Title>
					<Card.Description>
						Semua hasil build CBT terbaru disajikan dari server web-admin: universal, ABI spesifik, APK DEMO, dan checksum.
					</Card.Description>
				</div>
				<Button href="/ujian?demo=1" variant="outline">Buka Demo Browser</Button>
			</div>
		</Card.Header>
		<Card.Content class="space-y-4 pt-6">
			{#if cbtArtifacts && cbtArtifacts.artifacts.length > 0}
				<div class="grid gap-3 lg:grid-cols-2">
					{#each cbtArtifacts.artifacts as artifact (artifact.fileName)}
						<div class="min-w-0 rounded-2xl border border-border bg-muted/30 p-4">
							<div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
								<div class="min-w-0 space-y-2">
									<div class="flex flex-wrap items-center gap-2">
										<Badge class="border-border bg-background text-foreground">{artifact.abi}</Badge>
										<Badge class="border-border bg-background text-muted-foreground">{artifact.variant}</Badge>
									</div>
									<p class="truncate text-sm font-semibold text-foreground">{artifact.label}</p>
									<p class="text-sm leading-6 text-muted-foreground">{artifact.description}</p>
								</div>
								<Button href={artifact.downloadUrl} download class="shrink-0">Unduh</Button>
							</div>
							<div class="mt-4 grid gap-3 text-xs text-muted-foreground sm:grid-cols-3">
								<div>
									<p class="font-semibold uppercase tracking-wide">Ukuran</p>
									<p class="mt-1 text-foreground">{formatBytes(artifact.sizeBytes)}</p>
								</div>
								<div>
									<p class="font-semibold uppercase tracking-wide">Update</p>
									<p class="mt-1 text-foreground">{formatDate(artifact.updatedAt)}</p>
								</div>
								<div>
									<p class="font-semibold uppercase tracking-wide">Pemakaian</p>
									<p class="mt-1 text-foreground">{artifact.recommendedFor}</p>
								</div>
							</div>
							{#if artifact.sha256}
								<p class="mt-3 break-all rounded-xl bg-background px-3 py-2 font-mono text-[11px] leading-5 text-muted-foreground">SHA256: {artifact.sha256}</p>
							{/if}
						</div>
					{/each}
				</div>
				<div class="rounded-2xl border border-amber-500/20 bg-amber-500/10 p-4 text-sm leading-6 text-amber-900 dark:text-amber-100">
					<p class="font-semibold">Catatan distribusi</p>
					<p>Gunakan APK release untuk ujian resmi. APK DEMO debug hanya untuk tes manual cepat; jangan dibagikan sebagai APK ujian resmi.</p>
				</div>
			{:else}
				<div class="rounded-2xl border border-dashed border-border bg-muted/30 p-6 text-sm leading-6 text-muted-foreground">
					<p class="font-semibold text-foreground">Artifact APK CBT belum terbaca.</p>
					<p class="mt-1">{cbtArtifactsError || 'Memuat daftar artifact APK CBT dari server...'}</p>
				</div>
			{/if}
		</Card.Content>
	</Card.Root>

	<Card.Root class="border-primary/20 bg-primary/10 shadow-sm">
		<Card.Content class="grid gap-4 pt-6 md:grid-cols-3">
			<div>
				<p class="text-sm font-semibold text-primary">Pantau Ujian</p>
				<p class="mt-1 text-sm leading-6 text-muted-foreground">Hari-H tetap diarahkan ke Pemantauan, sesi hari ini, dan panel ruang.</p>
			</div>
			<div>
				<p class="text-sm font-semibold text-primary">Panduan BYOD</p>
				<p class="mt-1 text-sm leading-6 text-muted-foreground">Panduan status dan daftar pemeriksaan kirim dipakai saat pengawas membaca kondisi siswa.</p>
			</div>
			<div>
				<p class="text-sm font-semibold text-primary">Perangkat & Kesiapan</p>
				<p class="mt-1 text-sm leading-6 text-muted-foreground">Halaman ini fokus pada kesiapan layanan sistem, APK, operator, dan bahan rilis.</p>
			</div>
		</Card.Content>
	</Card.Root>

	<div class="grid gap-6 xl:grid-cols-3">
		<Card.Root class="border-border shadow-sm">
			<Card.Header>
				<Card.Title class="text-lg text-foreground">Pemeriksaan Layanan Sistem</Card.Title>
				<Card.Description>
					Pastikan perubahan layanan ujian tidak diam-diam merusak kontrak yang dipakai aplikasi siswa.
				</Card.Description>
			</Card.Header>
			<Card.Content>
				<ul class="space-y-3">
					{#each backendChecks as item (item)}
						<li class="flex gap-3 rounded-2xl border border-border bg-muted/50 px-4 py-3 text-sm leading-6 text-foreground">
							<span class="mt-0.5 inline-flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-primary/15 text-xs font-semibold text-primary">SISTEM</span>
							<span>{item}</span>
						</li>
					{/each}
				</ul>
			</Card.Content>
		</Card.Root>

		<Card.Root class="border-border shadow-sm">
			<Card.Header>
				<Card.Title class="text-lg text-foreground">Pemeriksaan APK</Card.Title>
				<Card.Description>
					Pastikan aplikasi siswa dan uji perangkat inti masih sehat sebelum APK diteruskan ke lapangan.
				</Card.Description>
			</Card.Header>
			<Card.Content>
				<ul class="space-y-3">
					{#each mobileChecks as item (item)}
						<li class="flex gap-3 rounded-2xl border border-border bg-muted/50 px-4 py-3 text-sm leading-6 text-foreground">
							<span class="mt-0.5 inline-flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-primary/15 text-xs font-semibold text-primary">APK</span>
							<span>{item}</span>
						</li>
					{/each}
				</ul>
			</Card.Content>
		</Card.Root>

		<Card.Root class="border-border shadow-sm">
			<Card.Header>
				<Card.Title class="text-lg text-foreground">Pemeriksaan Rilis</Card.Title>
				<Card.Description>
					Pastikan operator, pengawas, dan bahan uji coba sudah benar-benar siap dipakai di gelombang uji berikutnya.
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
				Buka dokumen-dokumen ini sebelum menyatakan perubahan layanan sistem atau APK aplikasi siswa aman untuk gelombang BYOD berikutnya.
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
