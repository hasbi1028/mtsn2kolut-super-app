<script lang="ts">
	import * as Card from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { resolve } from '$app/paths';

	type StatusTone = 'good' | 'warning' | 'danger';

	type StatusGuide = {
		label: string;
		tone: StatusTone;
		meaning: string;
		intervention: string;
	};
	type QuickLink = {
		label: string;
		href: string;
		description: string;
	};
	type QuickLinkGroup = {
		title: string;
		description: string;
		links: QuickLink[];
	};

	const statuses: StatusGuide[] = [
		{
			label: 'Tersambung',
			tone: 'good',
			meaning: 'Perangkat baru saja berkomunikasi baik dengan server dan tidak ada jawaban lokal yang tertahan.',
			intervention: 'Siswa dapat lanjut mengerjakan soal seperti biasa.'
		},
		{
			label: 'Lokal',
			tone: 'warning',
			meaning: 'Sebagian jawaban masih aman di perangkat dan menunggu sinkron ulang.',
			intervention: 'Minta siswa tetap di layar ujian dan pantau sampai sinkron kembali normal.'
		},
		{
			label: 'Waspada',
			tone: 'warning',
			meaning: 'Kontak server mulai lama. Belum tentu gagal, tetapi perangkat perlu refresh status.',
			intervention: 'Minta siswa menekan sinkron ulang dan pastikan koneksi masih sehat.'
		},
		{
			label: 'Gangguan',
			tone: 'danger',
			meaning: 'Aplikasi baru saja gagal menyimpan jawaban atau memperbarui status ke server.',
			intervention: 'Pantau jaringan, jangan buru-buru submit, lalu coba sinkron ulang.'
		},
		{
			label: 'Menurun',
			tone: 'danger',
			meaning: 'Gangguan sinkron sudah berulang. Submit manual memang ditahan sampai sesi cukup pulih.',
			intervention: 'Pengawas harus intervensi. Siswa tetap di layar ujian sampai status membaik.'
		}
	];

	const preSubmitChecklist = [
		'Status bukan Menurun.',
		'Tidak ada jawaban lokal yang masih menunggu sinkron.',
		'Refresh status terakhir berhasil.',
		'Perangkat yang dipakai masih sama dengan perangkat saat login.',
		'Siswa tetap berada di layar ujian sebelum menekan Kirim Ujian.'
	];

	const trialFlow = [
		'Sebelum sesi, pastikan APK terpasang dan API base URL yang dipakai benar.',
		'Saat login, cek apakah ada kartu restore dan perhatikan label kesehatan sesi terakhir.',
		'Selama ujian, pantau badge status app bar dan panel kesehatan koneksi di layar siswa.',
		'Saat gangguan disimulasikan, minta siswa tetap berada di layar ujian sampai sinkron pulih.',
		'Sebelum submit, ulangi checklist pengawas dan jangan izinkan kirim jika status masih Menurun.'
	];

	const monitoringGroups: QuickLinkGroup[] = [
		{
			title: 'Pantau Ujian',
			description: 'Aksi utama hari-H: buka sesi hari ini, ruang pengawas, dan status koneksi siswa.',
			links: [
				{
					label: 'Sesi Hari Ini / Aktif',
					href: '/asesmen/sesi?schedule=today',
					description: 'Daftar sesi yang perlu dipantau hari ini.'
				},
				{
					label: 'Dashboard Ruang',
					href: '/asesmen/pengawasan',
					description: 'Buka rekap ruang/proctoring untuk dashboard live.'
				},
				{
					label: 'Status Guide',
					href: '#status-guide',
					description: 'Arti Tersambung, Lokal, Waspada, Gangguan, dan Menurun.'
				}
			]
		},
		{
			title: 'Panduan BYOD',
			description: 'Bahan pengawas saat perlu menjelaskan status, submit, dan alur trial kepada siswa.',
			links: [
				{
					label: 'Arti Status Koneksi',
					href: '#status-guide',
					description: 'Makna badge aplikasi mobile dan tindakan pengawas.'
				},
				{
					label: 'Checklist Submit',
					href: '#submit-checklist',
					description: 'Pemeriksaan singkat sebelum siswa menekan Kirim Ujian.'
				},
				{
					label: 'Alur Trial BYOD',
					href: '#trial-flow',
					description: 'Urutan latihan untuk operator dan pengawas.'
				}
			]
		},
		{
			title: 'Perangkat & Kesiapan',
			description: 'Dibuka setelah kebutuhan monitoring terpenuhi: matrix perangkat dan release checklist.',
			links: [
				{
					label: 'Matriks Perangkat',
					href: '/asesmen/aplikasi-siswa/matrix',
					description: 'Bandingkan vendor, model, koneksi, restore, audio, dan submit.'
				},
				{
					label: 'Release Checklist',
					href: '/asesmen/aplikasi-siswa/release',
					description: 'Cek kesiapan backend, APK, operator, dan artefak rollout.'
				}
			]
		}
	];

	function badgeClass(tone: StatusTone) {
		switch (tone) {
			case 'good':
				return 'border-primary/20 bg-primary/10 text-primary';
			case 'warning':
				return 'border-warning/30 bg-warning/10 text-warning';
			case 'danger':
				return 'border-destructive/30 bg-destructive/10 text-destructive';
		}
	}

</script>

<svelte:head>
	<title>Panduan BYOD CBT — MTsN 2 Kolaka Utara</title>
</svelte:head>

<div class="space-y-6">
	<section class="rounded-3xl border border-primary/20 bg-gradient-to-br from-primary/10 via-card to-primary/10 p-6 shadow-sm">
		<div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
			<div class="max-w-3xl space-y-3">
				<p class="text-xs font-semibold uppercase tracking-[0.24em] text-primary">Modul 4 dari 5 · Monitoring</p>
				<h1 class="text-3xl font-semibold tracking-tight text-foreground">Monitoring CBT Mobile BYOD</h1>
				<p class="max-w-2xl text-sm leading-6 text-muted-foreground">
					Mulai dari kebutuhan hari-H: buka sesi aktif, pantau dashboard ruang, lalu gunakan status guide
					hanya saat pengawas perlu membaca sinyal koneksi siswa.
				</p>
			</div>
			<div class="flex flex-wrap gap-3">
				<Button href={resolve('/asesmen')} variant="outline">Beranda CBT</Button>
				<Button href="/asesmen/sesi?schedule=today">Pantau Sesi Hari Ini</Button>
				<Button href="/asesmen/pengawasan" variant="outline">Dashboard Ruang</Button>
				<Button href="#status-guide" variant="outline">Status Guide</Button>
			</div>
		</div>
	</section>

	<Card.Root class="border-primary/20 bg-primary/10 shadow-sm">
		<Card.Header>
			<Card.Title class="text-lg text-foreground">Pantau Ujian Dulu</Card.Title>
			<Card.Description>
				Tiga grup sederhana untuk monitoring: hari-H dulu, panduan setelahnya, perangkat dan kesiapan sebagai dukungan.
			</Card.Description>
		</Card.Header>
		<Card.Content class="grid gap-4 lg:grid-cols-3">
			{#each monitoringGroups as group (group.title)}
				<div class="rounded-2xl border border-primary/20 bg-card p-4 shadow-sm">
					<p class="text-sm font-semibold text-primary">{group.title}</p>
					<p class="mt-1 text-sm leading-6 text-muted-foreground">{group.description}</p>
					<div class="mt-4 space-y-2">
						{#each group.links as link (link.href)}
							<a href={resolve((link.href.startsWith('#') ? `/asesmen/aplikasi-siswa${link.href}` : link.href) as '/')} class="block rounded-xl border border-primary/20 bg-primary/10 px-3 py-2 text-sm transition hover:border-primary/20 hover:bg-primary/10">
								<span class="font-semibold text-primary">{link.label}</span>
								<span class="mt-1 block leading-5 text-muted-foreground">{link.description}</span>
							</a>
						{/each}
					</div>
				</div>
			{/each}
		</Card.Content>
	</Card.Root>

	<div class="grid gap-6 xl:grid-cols-[1.2fr_0.8fr]">
		<Card.Root id="status-guide" class="border-border shadow-sm">
			<Card.Header>
				<Card.Title class="text-lg text-foreground">Arti Status Koneksi Mobile</Card.Title>
				<Card.Description>
					Gunakan arti status ini saat mendampingi siswa. Fokus utamanya adalah kapan pengawas cukup memantau dan kapan harus menahan submit.
				</Card.Description>
			</Card.Header>
			<Card.Content class="space-y-4">
				{#each statuses as status (status.label)}
					<div class="rounded-2xl border border-border bg-card p-4">
						<div class="flex flex-wrap items-center gap-3">
							<Badge class={badgeClass(status.tone)}>{status.label}</Badge>
							<p class="text-sm font-medium text-foreground">{status.meaning}</p>
						</div>
						<p class="mt-3 text-sm leading-6 text-muted-foreground">
							<span class="font-semibold text-foreground">Tindakan pengawas:</span> {status.intervention}
						</p>
					</div>
				{/each}
			</Card.Content>
		</Card.Root>

		<div class="space-y-6">
			<Card.Root id="submit-checklist" class="border-border shadow-sm">
				<Card.Header>
					<Card.Title class="text-lg text-foreground">Checklist Sebelum Submit</Card.Title>
					<Card.Description>
						Lima pemeriksaan singkat ini sebaiknya selalu diulang sebelum pengawas mengizinkan siswa menekan kirim ujian.
					</Card.Description>
				</Card.Header>
				<Card.Content>
					<ul class="space-y-3">
						{#each preSubmitChecklist as item (item)}
							<li class="flex gap-3 rounded-2xl border border-border bg-muted/50 px-4 py-3 text-sm leading-6 text-foreground">
								<span class="mt-0.5 inline-flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-primary/15 text-xs font-semibold text-primary">OK</span>
								<span>{item}</span>
							</li>
						{/each}
					</ul>
				</Card.Content>
			</Card.Root>

			<Card.Root id="trial-flow" class="border-border shadow-sm">
				<Card.Header>
					<Card.Title class="text-lg text-foreground">Alur Trial BYOD</Card.Title>
					<Card.Description>
						Gunakan urutan ini saat uji coba perangkat siswa agar hasil antar pengawas tetap konsisten.
					</Card.Description>
				</Card.Header>
				<Card.Content>
					<ol class="space-y-3">
						{#each trialFlow as item, index (item)}
							<li class="flex gap-3 rounded-2xl border border-border bg-card px-4 py-3 text-sm leading-6 text-foreground">
								<span class="mt-0.5 inline-flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-foreground text-xs font-semibold text-background">{index + 1}</span>
								<span>{item}</span>
							</li>
						{/each}
					</ol>
				</Card.Content>
			</Card.Root>
		</div>
	</div>

	<Card.Root class="border-border shadow-sm">
		<Card.Header>
			<Card.Title class="text-lg text-foreground">Artefak Operasional</Card.Title>
			<Card.Description>
				Gunakan dokumen ini di repo yang sama untuk trial lapangan dan review kompatibilitas backend-mobile.
			</Card.Description>
		</Card.Header>
		<Card.Content class="grid gap-4 lg:grid-cols-5">
			<div class="rounded-2xl border border-border bg-muted/50 p-4">
				<p class="text-sm font-semibold text-foreground">Operator Quick Start</p>
				<p class="mt-2 text-sm leading-6 text-muted-foreground">
					Panduan singkat pengawas saat mendampingi siswa, termasuk arti status dan langkah saat koneksi mulai terganggu.
				</p>
				<p class="mt-3 font-mono text-xs text-muted-foreground">apps/mobile/OPERATOR_QUICKSTART.md</p>
			</div>
			<div class="rounded-2xl border border-border bg-muted/50 p-4">
				<p class="text-sm font-semibold text-foreground">BYOD Trial Procedure</p>
				<p class="mt-2 text-sm leading-6 text-muted-foreground">
					Prosedur end-to-end untuk operator, pengawas, siswa, simulasi gangguan, dan keputusan submit readiness.
				</p>
				<p class="mt-3 font-mono text-xs text-muted-foreground">apps/mobile/BYOD_TRIAL_PROCEDURE.md</p>
			</div>
			<div class="rounded-2xl border border-border bg-muted/50 p-4">
				<p class="text-sm font-semibold text-foreground">Device Test Matrix</p>
				<p class="mt-2 text-sm leading-6 text-muted-foreground">
					Matriks vendor dan model perangkat untuk mencatat hasil uji install, restore, audio, gambar, dan submit.
				</p>
				<p class="mt-3 font-mono text-xs text-muted-foreground">apps/mobile/DEVICE_TEST_MATRIX.md</p>
			</div>
			<div class="rounded-2xl border border-border bg-muted/50 p-4">
				<p class="text-sm font-semibold text-foreground">Ringkasan Matriks di Admin</p>
				<p class="mt-2 text-sm leading-6 text-muted-foreground">
					Gunakan halaman matriks perangkat di admin untuk membaca struktur evaluasi vendor tanpa keluar dari dashboard.
				</p>
				<p class="mt-3 font-mono text-xs text-muted-foreground">/asesmen/aplikasi-siswa/matrix</p>
			</div>
			<div class="rounded-2xl border border-border bg-muted/50 p-4">
				<p class="text-sm font-semibold text-foreground">Readiness Release di Admin</p>
				<p class="mt-2 text-sm leading-6 text-muted-foreground">
					Buka ringkasan backend, mobile, dan rollout checklist sebelum perubahan backend exam atau APK dinyatakan siap uji lapangan.
				</p>
				<p class="mt-3 font-mono text-xs text-muted-foreground">/asesmen/aplikasi-siswa/release</p>
			</div>
		</Card.Content>
	</Card.Root>
</div>
