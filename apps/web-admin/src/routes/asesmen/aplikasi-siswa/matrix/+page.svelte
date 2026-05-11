<script lang="ts">
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';

	type DeviceRow = {
		vendor: string;
		model: string;
		android: string;
		ram: string;
		connection: string;
		install: string;
		login: string;
		restore: string;
		audio: string;
		submit: string;
		note: string;
	};

	const templateRows: DeviceRow[] = [
		{
			vendor: 'Samsung',
			model: 'Galaxy A14',
			android: '14',
			ram: '4 GB',
			connection: 'Wi-Fi',
			install: 'Lulus',
			login: 'Lulus',
			restore: 'Lulus',
			audio: 'Lulus',
			submit: 'Lulus',
			note: 'Stabil untuk acuan uji awal'
		},
		{
			vendor: 'Xiaomi',
			model: 'Redmi Note 11',
			android: '13',
			ram: '4 GB',
			connection: 'Data',
			install: 'Perlu perhatian',
			login: 'Lulus',
			restore: 'Perlu perhatian',
			audio: 'Lulus',
			submit: 'Lulus',
			note: 'Perlu cek lagi saat aplikasi dibawa ke latar belakang'
		},
		{
			vendor: 'Oppo',
			model: 'A57',
			android: '13',
			ram: '4 GB',
			connection: 'Wi-Fi',
			install: 'Lulus',
			login: 'Lulus',
			restore: 'Lulus',
			audio: 'Perlu perhatian',
			submit: 'Lulus',
			note: 'Audio perlu diuji ulang dengan file berbeda'
		}
	];

	const focusChecks = [
		'APK bisa dipasang tanpa langkah aneh tambahan.',
		'Masuk dengan kode ujian berhasil pada koneksi yang dipakai siswa.',
		'Pemulihan sesi tetap berjalan setelah aplikasi ditutup lalu dibuka lagi.',
		'Status Waspada dan Menurun muncul sesuai simulasi gangguan.',
		'Kirim ujian hanya dilakukan saat koneksi kembali sehat.'
	];

	function badgeClass(value: string) {
		if (value === 'Lulus') return 'border-primary/20 bg-primary/10 text-primary';
		if (value === 'Perlu perhatian') return 'border-warning/30 bg-warning/10 text-warning';
		return 'border-destructive/30 bg-destructive/10 text-destructive';
	}
</script>

<svelte:head>
	<title>Tabel Perangkat BYOD — MTsN 2 Kolaka Utara</title>
</svelte:head>

<div class="space-y-6">
	<section class="rounded-3xl border border-primary/20 bg-gradient-to-br from-primary/10 via-card to-primary/10 p-6 shadow-sm">
		<div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
			<div class="max-w-3xl space-y-3">
				<p class="text-xs font-semibold uppercase tracking-[0.24em] text-primary">Perangkat & Kesiapan · Pemantauan</p>
				<h1 class="text-3xl font-semibold tracking-tight text-foreground">Perbandingan Perangkat BYOD</h1>
				<p class="max-w-2xl text-sm leading-6 text-muted-foreground">
					Bagian pendukung setelah pemantauan hari-H. Gunakan untuk membaca kesiapan perangkat siswa:
					pemasangan, masuk ujian, pemulihan sesi, media, koneksi, dan kirim ujian.
				</p>
			</div>
			<div class="flex flex-wrap gap-3">
				<Button href="/asesmen/aplikasi-siswa">Kembali ke Pemantauan</Button>
				<Button href="/asesmen/pengawasan" variant="outline">Panel Ruang</Button>
				<Button href="/asesmen/sesi?schedule=today" variant="outline">Sesi Hari Ini</Button>
			</div>
		</div>
	</section>

	<Card.Root class="border-primary/20 bg-primary/10 shadow-sm">
		<Card.Content class="grid gap-4 pt-6 md:grid-cols-3">
			<div>
				<p class="text-sm font-semibold text-primary">Pantau Ujian</p>
				<p class="mt-1 text-sm leading-6 text-muted-foreground">Untuk hari-H, mulai dari Pemantauan, sesi hari ini, atau panel ruang.</p>
			</div>
			<div>
				<p class="text-sm font-semibold text-primary">Panduan BYOD</p>
				<p class="mt-1 text-sm leading-6 text-muted-foreground">Panduan status dan daftar pemeriksaan kirim tetap berada di halaman Pemantauan.</p>
			</div>
			<div>
				<p class="text-sm font-semibold text-primary">Perangkat & Kesiapan</p>
				<p class="mt-1 text-sm leading-6 text-muted-foreground">Halaman ini hanya untuk pembanding perangkat dan catatan kesiapan teknis.</p>
			</div>
		</Card.Content>
	</Card.Root>

	<div class="grid gap-6 xl:grid-cols-[1.15fr_0.85fr]">
		<Card.Root class="border-border shadow-sm">
			<Card.Header>
				<Card.Title class="text-lg text-foreground">Contoh Tabel Perangkat</Card.Title>
				<Card.Description>
					Baris di bawah adalah contoh format, bukan hasil sertifikasi perangkat resmi sekolah.
				</Card.Description>
			</Card.Header>
			<Card.Content class="space-y-4">
				<div class="overflow-x-auto rounded-2xl border border-border">
					<Table.Root>
						<Table.Header>
							<Table.Row>
								<Table.Head>Vendor</Table.Head>
								<Table.Head>Model</Table.Head>
								<Table.Head>Android</Table.Head>
								<Table.Head>RAM</Table.Head>
								<Table.Head>Koneksi</Table.Head>
								<Table.Head>Pasang</Table.Head>
								<Table.Head>Masuk</Table.Head>
								<Table.Head>Pulihkan</Table.Head>
								<Table.Head>Audio</Table.Head>
								<Table.Head>Kirim</Table.Head>
								<Table.Head>Catatan</Table.Head>
							</Table.Row>
						</Table.Header>
						<Table.Body>
							{#each templateRows as row (row.vendor + row.model)}
								<Table.Row>
									<Table.Cell class="font-medium text-foreground">{row.vendor}</Table.Cell>
									<Table.Cell>{row.model}</Table.Cell>
									<Table.Cell>{row.android}</Table.Cell>
									<Table.Cell>{row.ram}</Table.Cell>
									<Table.Cell>{row.connection}</Table.Cell>
									<Table.Cell><Badge class={badgeClass(row.install)}>{row.install}</Badge></Table.Cell>
									<Table.Cell><Badge class={badgeClass(row.login)}>{row.login}</Badge></Table.Cell>
									<Table.Cell><Badge class={badgeClass(row.restore)}>{row.restore}</Badge></Table.Cell>
									<Table.Cell><Badge class={badgeClass(row.audio)}>{row.audio}</Badge></Table.Cell>
									<Table.Cell><Badge class={badgeClass(row.submit)}>{row.submit}</Badge></Table.Cell>
									<Table.Cell class="min-w-56 text-sm text-muted-foreground">{row.note}</Table.Cell>
								</Table.Row>
							{/each}
						</Table.Body>
					</Table.Root>
				</div>
				<p class="text-xs leading-5 text-muted-foreground">
					Format sumber resminya tetap ada di <span class="font-mono">apps/mobile/DEVICE_TEST_MATRIX.md</span>.
					Halaman ini disediakan agar pengawas dan operator bisa membaca struktur penilaian tanpa keluar dari web admin.
				</p>
			</Card.Content>
		</Card.Root>

		<div class="space-y-6">
			<Card.Root class="border-border shadow-sm">
				<Card.Header>
					<Card.Title class="text-lg text-foreground">Fokus Uji Minimal</Card.Title>
					<Card.Description>
						Lima poin ini yang paling penting saat membandingkan perangkat siswa sebelum masuk uji yang lebih besar.
					</Card.Description>
				</Card.Header>
				<Card.Content>
					<ul class="space-y-3">
						{#each focusChecks as item (item)}
							<li class="flex gap-3 rounded-2xl border border-border bg-muted/50 px-4 py-3 text-sm leading-6 text-foreground">
								<span class="mt-0.5 inline-flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-primary/15 text-xs font-semibold text-primary">UJI</span>
								<span>{item}</span>
							</li>
						{/each}
					</ul>
				</Card.Content>
			</Card.Root>

			<Card.Root class="border-border shadow-sm">
				<Card.Header>
					<Card.Title class="text-lg text-foreground">Interpretasi Hasil</Card.Title>
					<Card.Description>
						Gunakan klasifikasi ini agar keputusan perangkat yang layak dipakai tetap konsisten antar operator.
					</Card.Description>
				</Card.Header>
				<Card.Content class="space-y-4">
					<div class="rounded-2xl border border-primary/20 bg-primary/10 p-4">
						<p class="text-sm font-semibold text-primary">Layak dipakai pada pemakaian awal</p>
						<p class="mt-2 text-sm leading-6 text-primary/80">
							Fungsi inti lulus, pemulihan sesi stabil, kirim ujian sehat, dan perangkat tidak sering masuk status Menurun.
						</p>
					</div>
					<div class="rounded-2xl border border-warning/30 bg-warning/10 p-4">
						<p class="text-sm font-semibold text-warning">Layak dengan catatan</p>
						<p class="mt-2 text-sm leading-6 text-warning/80">
							Fungsi inti jalan, tetapi ada catatan kecil seperti audio lambat atau perlu perbarui manual sesekali.
						</p>
					</div>
					<div class="rounded-2xl border border-destructive/30 bg-destructive/10 p-4">
						<p class="text-sm font-semibold text-destructive">Tidak direkomendasikan</p>
						<p class="mt-2 text-sm leading-6 text-destructive/80">
							Masuk ujian, pemulihan sesi, atau kirim ujian sering gagal walau perangkat lain pada jaringan yang sama berjalan baik.
						</p>
					</div>
				</Card.Content>
			</Card.Root>
		</div>
	</div>
</div>
