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
			note: 'Stabil untuk baseline uji awal'
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
			note: 'Perlu cek lagi saat app dibawa ke background'
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
		'Login token berhasil pada koneksi yang dipakai siswa.',
		'Restore sesi tetap berjalan setelah app ditutup lalu dibuka lagi.',
		'Status Waspada dan Menurun muncul sesuai simulasi gangguan.',
		'Submit hanya dilakukan saat koneksi kembali sehat.'
	];

	function badgeClass(value: string) {
		if (value === 'Lulus') return 'border-emerald-200 bg-emerald-50 text-emerald-700';
		if (value === 'Perlu perhatian') return 'border-amber-200 bg-amber-50 text-amber-700';
		return 'border-rose-200 bg-rose-50 text-rose-700';
	}
</script>

<svelte:head>
	<title>Matriks Perangkat BYOD — MTsN 2 Kolaka Utara</title>
</svelte:head>

<div class="space-y-6">
	<section class="rounded-3xl border border-emerald-100 bg-gradient-to-br from-emerald-50 via-white to-lime-50 p-6 shadow-sm">
		<div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
			<div class="max-w-3xl space-y-3">
				<p class="text-xs font-semibold uppercase tracking-[0.24em] text-emerald-700">Perangkat & Kesiapan · Monitoring</p>
				<h1 class="text-3xl font-semibold tracking-tight text-slate-900">Perbandingan Perangkat BYOD</h1>
				<p class="max-w-2xl text-sm leading-6 text-slate-600">
					Bagian pendukung setelah monitoring hari-H. Gunakan untuk membaca kesiapan perangkat siswa:
					install, login, restore, media, koneksi, dan submit.
				</p>
			</div>
			<div class="flex flex-wrap gap-3">
				<Button href="/asesmen/aplikasi-siswa">Kembali ke Monitoring</Button>
				<Button href="/asesmen/pengawasan" variant="outline">Dashboard Ruang</Button>
				<Button href="/asesmen/sesi?schedule=today" variant="outline">Sesi Hari Ini</Button>
			</div>
		</div>
	</section>

	<Card.Root class="border-emerald-200 bg-emerald-50/60 shadow-sm">
		<Card.Content class="grid gap-4 pt-6 md:grid-cols-3">
			<div>
				<p class="text-sm font-semibold text-emerald-950">Pantau Ujian</p>
				<p class="mt-1 text-sm leading-6 text-slate-600">Untuk hari-H, mulai dari Monitoring, sesi hari ini, atau dashboard ruang.</p>
			</div>
			<div>
				<p class="text-sm font-semibold text-emerald-950">Panduan BYOD</p>
				<p class="mt-1 text-sm leading-6 text-slate-600">Status guide dan checklist submit tetap berada di halaman Monitoring.</p>
			</div>
			<div>
				<p class="text-sm font-semibold text-emerald-950">Perangkat & Kesiapan</p>
				<p class="mt-1 text-sm leading-6 text-slate-600">Halaman ini hanya untuk pembanding perangkat dan catatan kesiapan teknis.</p>
			</div>
		</Card.Content>
	</Card.Root>

	<div class="grid gap-6 xl:grid-cols-[1.15fr_0.85fr]">
		<Card.Root class="border-slate-200 shadow-sm">
			<Card.Header>
				<Card.Title class="text-lg text-slate-900">Template Contoh Matriks Perangkat</Card.Title>
				<Card.Description>
					Baris di bawah adalah contoh/template, bukan hasil sertifikasi perangkat resmi sekolah.
				</Card.Description>
			</Card.Header>
			<Card.Content class="space-y-4">
				<div class="overflow-x-auto rounded-2xl border border-slate-200">
					<Table.Root>
						<Table.Header>
							<Table.Row>
								<Table.Head>Vendor</Table.Head>
								<Table.Head>Model</Table.Head>
								<Table.Head>Android</Table.Head>
								<Table.Head>RAM</Table.Head>
								<Table.Head>Koneksi</Table.Head>
								<Table.Head>Install</Table.Head>
								<Table.Head>Login</Table.Head>
								<Table.Head>Restore</Table.Head>
								<Table.Head>Audio</Table.Head>
								<Table.Head>Submit</Table.Head>
								<Table.Head>Catatan</Table.Head>
							</Table.Row>
						</Table.Header>
						<Table.Body>
							{#each templateRows as row (row.vendor + row.model)}
								<Table.Row>
									<Table.Cell class="font-medium text-slate-900">{row.vendor}</Table.Cell>
									<Table.Cell>{row.model}</Table.Cell>
									<Table.Cell>{row.android}</Table.Cell>
									<Table.Cell>{row.ram}</Table.Cell>
									<Table.Cell>{row.connection}</Table.Cell>
									<Table.Cell><Badge class={badgeClass(row.install)}>{row.install}</Badge></Table.Cell>
									<Table.Cell><Badge class={badgeClass(row.login)}>{row.login}</Badge></Table.Cell>
									<Table.Cell><Badge class={badgeClass(row.restore)}>{row.restore}</Badge></Table.Cell>
									<Table.Cell><Badge class={badgeClass(row.audio)}>{row.audio}</Badge></Table.Cell>
									<Table.Cell><Badge class={badgeClass(row.submit)}>{row.submit}</Badge></Table.Cell>
									<Table.Cell class="min-w-56 text-sm text-slate-600">{row.note}</Table.Cell>
								</Table.Row>
							{/each}
						</Table.Body>
					</Table.Root>
				</div>
				<p class="text-xs leading-5 text-slate-500">
					Format sumber resminya tetap ada di <span class="font-mono">apps/mobile/DEVICE_TEST_MATRIX.md</span>.
					Halaman ini disediakan agar pengawas dan operator bisa membaca struktur penilaian tanpa keluar dari web admin.
				</p>
			</Card.Content>
		</Card.Root>

		<div class="space-y-6">
			<Card.Root class="border-slate-200 shadow-sm">
				<Card.Header>
					<Card.Title class="text-lg text-slate-900">Fokus Uji Minimal</Card.Title>
					<Card.Description>
						Lima poin ini yang paling penting saat membandingkan perangkat siswa sebelum masuk uji yang lebih besar.
					</Card.Description>
				</Card.Header>
				<Card.Content>
					<ul class="space-y-3">
						{#each focusChecks as item (item)}
							<li class="flex gap-3 rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3 text-sm leading-6 text-slate-700">
								<span class="mt-0.5 inline-flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-emerald-100 text-xs font-semibold text-emerald-700">UJI</span>
								<span>{item}</span>
							</li>
						{/each}
					</ul>
				</Card.Content>
			</Card.Root>

			<Card.Root class="border-slate-200 shadow-sm">
				<Card.Header>
					<Card.Title class="text-lg text-slate-900">Interpretasi Hasil</Card.Title>
					<Card.Description>
						Gunakan klasifikasi ini agar keputusan perangkat yang layak dipakai tetap konsisten antar operator.
					</Card.Description>
				</Card.Header>
				<Card.Content class="space-y-4">
					<div class="rounded-2xl border border-emerald-200 bg-emerald-50 p-4">
						<p class="text-sm font-semibold text-emerald-800">Layak dipakai produksi awal</p>
						<p class="mt-2 text-sm leading-6 text-emerald-900/80">
							Fungsi inti lulus, restore stabil, submit sehat, dan perangkat tidak sering masuk status Menurun.
						</p>
					</div>
					<div class="rounded-2xl border border-amber-200 bg-amber-50 p-4">
						<p class="text-sm font-semibold text-amber-800">Layak dengan catatan</p>
						<p class="mt-2 text-sm leading-6 text-amber-900/80">
							Fungsi inti jalan, tetapi ada issue minor seperti audio lambat atau perlu refresh manual sesekali.
						</p>
					</div>
					<div class="rounded-2xl border border-rose-200 bg-rose-50 p-4">
						<p class="text-sm font-semibold text-rose-800">Tidak direkomendasikan</p>
						<p class="mt-2 text-sm leading-6 text-rose-900/80">
							Login, restore, atau submit sering gagal walau perangkat lain pada jaringan yang sama berjalan baik.
						</p>
					</div>
				</Card.Content>
			</Card.Root>
		</div>
	</div>
</div>
