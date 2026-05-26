<script lang="ts">
	import * as Card from '$lib/components/ui/card';
	import MicroActionTable from '$lib/components/ops/MicroActionTable.svelte';
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
		'Portal Ujian Web /ujian bisa dibuka tanpa instalasi APK.',
		'Masuk ujian web berhasil pada koneksi yang dipakai siswa.',
		'Pemulihan/sinkron sesi web tetap berjalan setelah tab/browser terganggu sesuai kemampuan browser.',
		'Status Waspada dan Menurun muncul sesuai simulasi gangguan.',
		'Kirim ujian hanya dilakukan saat koneksi kembali sehat.'
	];

	const deviceColumns = [
		{ key: 'device', label: 'Perangkat', class: 'min-w-44' },
		{ key: 'spec', label: 'Spesifikasi' },
		{ key: 'connection', label: 'Koneksi' },
		{ key: 'install', label: 'Pasang' },
		{ key: 'login', label: 'Masuk' },
		{ key: 'restore', label: 'Pulihkan' },
		{ key: 'audio', label: 'Audio' },
		{ key: 'submit', label: 'Kirim' },
		{ key: 'note', label: 'Catatan', class: 'min-w-56' }
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
					Bagian pendukung setelah pemantauan hari-H. Runtime resmi siswa tahun ini adalah Portal Ujian Web;
					gunakan halaman ini untuk membaca kesiapan browser/perangkat, media, koneksi, dan kirim ujian. APK Flutter hanya arsip nonaktif/tahap lanjutan.
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
				<p class="mt-1 text-sm leading-6 text-muted-foreground">Halaman ini hanya untuk pembanding perangkat/browser dan catatan teknis; APK Flutter nonaktif tahun ini.</p>
			</div>
		</Card.Content>
	</Card.Root>

	<div class="grid gap-6 xl:grid-cols-[1.15fr_0.85fr]">
		<div class="space-y-3">
			<MicroActionTable
				title="Contoh Tabel Perangkat"
				description="Baris di bawah adalah contoh format, bukan hasil sertifikasi perangkat resmi sekolah."
				columns={deviceColumns}
				rows={templateRows}
				rowKey={(row) => `${(row as DeviceRow).vendor}-${(row as DeviceRow).model}`}
				tableClass="min-w-[980px]"
			>
				{#snippet cell(row, column)}
					{@const device = row as DeviceRow}
					{#if column.key === 'device'}
						<div>
							<p class="font-medium text-foreground">{device.vendor}</p>
							<p class="text-xs text-muted-foreground">{device.model}</p>
						</div>
					{:else if column.key === 'spec'}
						<span>{device.android} · {device.ram}</span>
					{:else if column.key === 'connection'}
						<span>{device.connection}</span>
					{:else if column.key === 'install'}
						<Badge class={badgeClass(device.install)}>{device.install}</Badge>
					{:else if column.key === 'login'}
						<Badge class={badgeClass(device.login)}>{device.login}</Badge>
					{:else if column.key === 'restore'}
						<Badge class={badgeClass(device.restore)}>{device.restore}</Badge>
					{:else if column.key === 'audio'}
						<Badge class={badgeClass(device.audio)}>{device.audio}</Badge>
					{:else if column.key === 'submit'}
						<Badge class={badgeClass(device.submit)}>{device.submit}</Badge>
					{:else if column.key === 'note'}
						<span class="text-muted-foreground">{device.note}</span>
					{/if}
				{/snippet}
				{#snippet mobile(row)}
					{@const device = row as DeviceRow}
					<div class="space-y-2 text-xs">
						<div>
							<p class="font-medium text-foreground">{device.vendor} {device.model}</p>
							<p class="text-muted-foreground">Android {device.android} · {device.ram} · {device.connection}</p>
						</div>
						<div class="flex flex-wrap gap-1.5">
							<Badge class={badgeClass(device.install)}>Pasang: {device.install}</Badge>
							<Badge class={badgeClass(device.login)}>Masuk: {device.login}</Badge>
							<Badge class={badgeClass(device.restore)}>Pulihkan: {device.restore}</Badge>
							<Badge class={badgeClass(device.audio)}>Audio: {device.audio}</Badge>
							<Badge class={badgeClass(device.submit)}>Kirim: {device.submit}</Badge>
						</div>
						<p class="leading-5 text-muted-foreground">{device.note}</p>
					</div>
				{/snippet}
			</MicroActionTable>
			<p class="text-xs leading-5 text-muted-foreground">
				Format sumber APK lama tetap ada sebagai arsip di <span class="font-mono">apps/mobile/DEVICE_TEST_MATRIX.md</span>.
				Halaman ini disediakan agar pengawas dan operator bisa membaca struktur penilaian tanpa keluar dari web admin.
			</p>
		</div>

		<div class="space-y-6">
			<Card.Root class="border-border shadow-sm">
				<Card.Header>
					<Card.Title class="text-lg text-foreground">Fokus Uji Minimal</Card.Title>
					<Card.Description>
						Lima poin ini yang paling penting saat membandingkan perangkat/browser siswa untuk Portal Ujian Web.
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
