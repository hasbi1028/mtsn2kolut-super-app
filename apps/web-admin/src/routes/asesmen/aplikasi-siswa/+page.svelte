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
			meaning: 'Perangkat baru saja berkomunikasi baik dengan layanan sistem dan tidak ada jawaban lokal yang tertahan.',
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
			meaning: 'Kontak layanan sistem mulai lama. Belum tentu gagal, tetapi perangkat perlu memperbarui status.',
			intervention: 'Minta siswa menekan sinkron ulang dan pastikan koneksi masih sehat.'
		},
		{
			label: 'Gangguan',
			tone: 'danger',
			meaning: 'Aplikasi baru saja gagal menyimpan jawaban atau memperbarui status ke layanan sistem.',
			intervention: 'Pantau jaringan, jangan buru-buru mengirim ujian, lalu coba sinkron ulang.'
		},
		{
			label: 'Menurun',
			tone: 'danger',
			meaning: 'Gangguan sinkron sudah berulang. Kirim ujian manual memang ditahan sampai sesi cukup pulih.',
			intervention: 'Pengawas harus intervensi. Siswa tetap di layar ujian sampai status membaik.'
		}
	];

	const preSubmitChecklist = [
		'Status bukan Menurun.',
		'Tidak ada jawaban lokal yang masih menunggu sinkron.',
		'Pembaruan status terakhir berhasil.',
		'Perangkat yang dipakai masih sama dengan perangkat saat masuk ujian.',
		'Siswa tetap berada di layar ujian sebelum menekan Kirim Ujian.'
	];

	const trialFlow = [
		'Sebelum sesi, pastikan APK terpasang dan alamat layanan sistem yang dipakai benar.',
		'Saat masuk ujian, cek apakah ada kartu pemulihan sesi dan perhatikan label kesehatan sesi terakhir.',
		'Selama ujian, pantau penanda status di bagian atas aplikasi dan panel kesehatan koneksi di layar siswa.',
		'Saat gangguan disimulasikan, minta siswa tetap berada di layar ujian sampai sinkron pulih.',
		'Sebelum kirim ujian, ulangi daftar pemeriksaan pengawas dan jangan izinkan kirim jika status masih Menurun.'
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
					label: 'Panel Ruang',
					href: '/asesmen/pengawasan',
					description: 'Buka rekap ruang dan pengawasan ujian untuk panel langsung.'
				},
				{
					label: 'Panduan Status',
					href: '#status-guide',
					description: 'Arti Tersambung, Lokal, Waspada, Gangguan, dan Menurun.'
				}
			]
		},
		{
			title: 'Panduan BYOD',
			description: 'Bahan pengawas saat perlu menjelaskan status, kirim ujian, dan alur uji coba kepada siswa.',
			links: [
				{
					label: 'Arti Status Koneksi',
					href: '#status-guide',
					description: 'Makna penanda status aplikasi siswa dan tindakan pengawas.'
				},
				{
					label: 'Daftar Pemeriksaan Kirim',
					href: '#submit-checklist',
					description: 'Pemeriksaan singkat sebelum siswa menekan Kirim Ujian.'
				},
				{
					label: 'Alur Uji Coba BYOD',
					href: '#trial-flow',
					description: 'Urutan latihan untuk operator dan pengawas.'
				}
			]
		},
		{
			title: 'Perangkat & Kesiapan',
			description: 'Dibuka setelah kebutuhan pemantauan terpenuhi: tabel perangkat dan daftar pemeriksaan rilis.',
			links: [
				{
					label: 'Tabel Perangkat',
					href: '/asesmen/aplikasi-siswa/matrix',
					description: 'Bandingkan vendor, model, koneksi, pemulihan sesi, audio, dan kirim ujian.'
				},
				{
					label: 'Daftar Pemeriksaan Rilis',
					href: '/asesmen/aplikasi-siswa/release',
					description: 'Cek kesiapan layanan sistem, APK, operator, dan bahan rilis.'
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
				<p class="text-xs font-semibold uppercase tracking-[0.24em] text-primary">Modul 4 dari 5 · Pemantauan</p>
				<h1 class="text-3xl font-semibold tracking-tight text-foreground">Panduan Aplikasi Siswa BYOD</h1>
				<p class="max-w-2xl text-sm leading-6 text-muted-foreground">
					Mulai dari kebutuhan hari-H: buka sesi aktif, pantau panel ruang, lalu gunakan panduan status
					hanya saat pengawas perlu membaca sinyal koneksi siswa.
				</p>
			</div>
			<div class="flex flex-wrap gap-3">
				<Button href={resolve('/asesmen')} variant="outline">Beranda Ujian</Button>
				<Button href="/asesmen/sesi?schedule=today">Pantau Sesi Hari Ini</Button>
				<Button href="/asesmen/pengawasan" variant="outline">Panel Ruang</Button>
				<Button href="#status-guide" variant="outline">Panduan Status</Button>
			</div>
		</div>
	</section>

	<Card.Root class="border-primary/20 bg-primary/10 shadow-sm">
		<Card.Header>
			<Card.Title class="text-lg text-foreground">Pantau Ujian Dulu</Card.Title>
			<Card.Description>
				Tiga grup sederhana untuk pemantauan: hari-H dulu, panduan setelahnya, perangkat dan kesiapan sebagai dukungan.
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
				<Card.Title class="text-lg text-foreground">Arti Status Koneksi Aplikasi Siswa</Card.Title>
				<Card.Description>
					Gunakan arti status ini saat mendampingi siswa. Fokus utamanya adalah kapan pengawas cukup memantau dan kapan harus menahan kirim ujian.
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
					<Card.Title class="text-lg text-foreground">Daftar Pemeriksaan Sebelum Kirim</Card.Title>
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
					<Card.Title class="text-lg text-foreground">Alur Uji Coba BYOD</Card.Title>
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
				Gunakan dokumen ini di repositori yang sama untuk uji coba lapangan dan pemeriksaan kesesuaian layanan sistem dengan aplikasi siswa.
			</Card.Description>
		</Card.Header>
		<Card.Content class="grid gap-4 lg:grid-cols-5">
			<div class="rounded-2xl border border-border bg-muted/50 p-4">
				<p class="text-sm font-semibold text-foreground">Panduan Cepat Operator</p>
				<p class="mt-2 text-sm leading-6 text-muted-foreground">
					Panduan singkat pengawas saat mendampingi siswa, termasuk arti status dan langkah saat koneksi mulai terganggu.
				</p>
				<p class="mt-3 font-mono text-xs text-muted-foreground">apps/mobile/OPERATOR_QUICKSTART.md</p>
			</div>
			<div class="rounded-2xl border border-border bg-muted/50 p-4">
				<p class="text-sm font-semibold text-foreground">Prosedur Uji Coba BYOD</p>
				<p class="mt-2 text-sm leading-6 text-muted-foreground">
					Prosedur lengkap untuk operator, pengawas, siswa, simulasi gangguan, dan keputusan kesiapan kirim ujian.
				</p>
				<p class="mt-3 font-mono text-xs text-muted-foreground">apps/mobile/BYOD_TRIAL_PROCEDURE.md</p>
			</div>
			<div class="rounded-2xl border border-border bg-muted/50 p-4">
				<p class="text-sm font-semibold text-foreground">Tabel Uji Perangkat</p>
				<p class="mt-2 text-sm leading-6 text-muted-foreground">
					Tabel vendor dan model perangkat untuk mencatat hasil uji pemasangan, pemulihan sesi, audio, gambar, dan kirim ujian.
				</p>
				<p class="mt-3 font-mono text-xs text-muted-foreground">apps/mobile/DEVICE_TEST_MATRIX.md</p>
			</div>
			<div class="rounded-2xl border border-border bg-muted/50 p-4">
				<p class="text-sm font-semibold text-foreground">Ringkasan Tabel di Admin</p>
				<p class="mt-2 text-sm leading-6 text-muted-foreground">
					Gunakan halaman tabel perangkat di admin untuk membaca struktur evaluasi vendor tanpa keluar dari beranda.
				</p>
				<p class="mt-3 font-mono text-xs text-muted-foreground">/asesmen/aplikasi-siswa/matrix</p>
			</div>
			<div class="rounded-2xl border border-border bg-muted/50 p-4">
				<p class="text-sm font-semibold text-foreground">Kesiapan Rilis di Admin</p>
				<p class="mt-2 text-sm leading-6 text-muted-foreground">
					Buka ringkasan layanan sistem, aplikasi siswa, dan daftar pemeriksaan rilis sebelum perubahan layanan ujian atau APK dinyatakan siap uji lapangan.
				</p>
				<p class="mt-3 font-mono text-xs text-muted-foreground">/asesmen/aplikasi-siswa/release</p>
			</div>
		</Card.Content>
	</Card.Root>
</div>
