<svelte:head>
	<title>Command Center CBT</title>
	<meta
		name="description"
		content="Command Center CBT web-first untuk alur Asesmen Ujian sederhana MTsN 2 Kolaka Utara."
	/>
</svelte:head>

<script lang="ts">
	const steps = [
		{
			label: 'Siapkan Ujian',
			detail: 'Buat nama ujian, pilih mapel/kelas, pilih paket dari Bank Soal, lalu tetapkan tanggal dan jam.',
			status: 'Tahap berikutnya'
		},
		{
			label: 'Atur Peserta & Ruang',
			detail: 'Pilih rombel, bentuk Ruang Ujian otomatis, cek preview pembagian peserta, lalu simpan saat backend siap.',
			status: 'Belum aktif'
		},
		{
			label: 'Cetak Kartu & Lembar Pengawas',
			detail: 'Cetak Kartu Peserta Ujian, Lembar Pengawas Ruang, dan daftar hadir sederhana berbasis QR+PIN.',
			status: 'Belum aktif'
		},
		{
			label: 'Pelaksanaan & Hasil',
			detail: 'Pantau status ruang, peserta masuk, peserta selesai, lalu lihat Rekap Nilai PG minimal.',
			status: 'Belum aktif'
		}
	];
</script>

<div class="min-h-screen bg-slate-50 text-slate-950">
	<div class="mx-auto flex w-full max-w-6xl flex-col gap-5 px-4 py-6 sm:px-6 lg:px-8">
		<header class="rounded-3xl border border-emerald-200 bg-white p-5 shadow-sm">
			<div class="flex flex-col gap-4 md:flex-row md:items-end md:justify-between">
				<div class="space-y-3">
					<span class="inline-flex w-fit rounded-full border border-emerald-200 bg-emerald-50 px-3 py-1 text-xs font-black uppercase tracking-[0.2em] text-emerald-800">
						CBT Web
					</span>
					<div class="space-y-2">
						<h1 class="text-3xl font-black tracking-tight text-slate-950 md:text-5xl">Command Center CBT</h1>
						<p class="max-w-3xl text-base leading-7 text-slate-600">
							Alur sederhana untuk panitia: siapkan ujian, jalankan ruang, buka portal peserta, lalu tutup hasil.
						</p>
					</div>
				</div>
				<div class="flex flex-wrap gap-2">
					<a href="/asesmen/prototype" class="rounded-2xl border border-slate-300 bg-white px-4 py-2 text-sm font-black text-slate-700 shadow-sm hover:border-emerald-300 hover:text-emerald-800">
						Lihat Prototype
					</a>
					<a href="/bank-soal" class="rounded-2xl bg-emerald-700 px-4 py-2 text-sm font-black text-white shadow-sm hover:bg-emerald-800">
						Buka Bank Soal
					</a>
				</div>
			</div>
		</header>

		<section class="rounded-3xl border border-slate-200 bg-white p-4 shadow-sm">
			<div class="mb-3 flex flex-col gap-1 border-b border-slate-200 pb-3 sm:flex-row sm:items-end sm:justify-between">
				<div>
					<p class="text-xs font-bold uppercase tracking-[0.18em] text-slate-500">Asesmen Ujian</p>
					<h2 class="text-xl font-black text-slate-950">Checklist ringkas panitia</h2>
				</div>
				<p class="text-sm font-semibold text-slate-500">Produksi shell minimal — belum ada mutasi data.</p>
			</div>

			<div class="divide-y divide-slate-200 overflow-hidden rounded-2xl border border-slate-200">
				{#each steps as step, index}
					<div class="grid gap-3 p-4 md:grid-cols-[3rem_1fr_auto] md:items-center">
						<div class="flex h-10 w-10 items-center justify-center rounded-2xl bg-emerald-50 text-sm font-black text-emerald-800">
							{index + 1}
						</div>
						<div class="min-w-0">
							<h3 class="font-black text-slate-950">{step.label}</h3>
							<p class="mt-1 text-sm leading-6 text-slate-600">{step.detail}</p>
						</div>
						<span class="w-fit rounded-full bg-slate-100 px-3 py-1 text-xs font-black text-slate-600">{step.status}</span>
					</div>
				{/each}
			</div>
		</section>

		<section class="grid gap-4 lg:grid-cols-[1fr_0.8fr]">
			<div class="rounded-3xl border border-slate-200 bg-white p-4 shadow-sm">
				<p class="text-xs font-bold uppercase tracking-[0.18em] text-slate-500">Batas aman</p>
				<h2 class="mt-1 text-lg font-black text-slate-950">Bank Soal tetap modul terpisah</h2>
				<p class="mt-2 text-sm leading-6 text-slate-600">
					Halaman ini hanya shell produksi Asesmen. Pembuatan, review, import, penerbitan, dan arsip soal tetap dilakukan di Bank Soal. Asesmen nanti hanya memilih paket yang sudah siap.
				</p>
			</div>

			<div class="rounded-3xl border border-dashed border-slate-300 bg-slate-100/80 p-4 shadow-sm">
				<p class="text-xs font-bold uppercase tracking-[0.18em] text-slate-500">Mode Lengkap Admin</p>
				<h2 class="mt-1 text-lg font-black text-slate-950">Disiapkan nanti setelah backend minimal</h2>
				<p class="mt-2 text-sm leading-6 text-slate-600">
					Pengaturan teknis seperti reset, koreksi manual, audit, dan detail ruang tidak ditampilkan dulu agar panitia melihat alur utama saja.
				</p>
			</div>
		</section>
	</div>
</div>
