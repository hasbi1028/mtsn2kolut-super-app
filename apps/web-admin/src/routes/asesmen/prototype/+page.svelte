<svelte:head>
	<title>Prototype Asesmen CBT Sederhana</title>
	<meta
		name="description"
		content="Prototype UI Asesmen CBT sederhana MTsN 2 Kolaka Utara tanpa koneksi backend."
	/>
</svelte:head>

<div class="min-h-screen bg-slate-50 text-slate-950">
	<div class="mx-auto flex w-full max-w-7xl flex-col gap-6 px-4 py-6 sm:px-6 lg:px-8">
		<header class="rounded-3xl border border-emerald-200 bg-white p-5 shadow-sm">
			<div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
				<div class="space-y-3">
					<p class="inline-flex w-fit rounded-full border border-amber-300 bg-amber-50 px-3 py-1 text-xs font-bold uppercase tracking-[0.2em] text-amber-800">
						PROTOTYPE UI — belum terhubung backend
					</p>
					<div class="space-y-2">
						<p class="text-sm font-semibold text-emerald-700">A0 — Command Center CBT Sederhana</p>
						<h1 class="max-w-3xl text-3xl font-black tracking-tight text-slate-950 md:text-5xl">
							Asesmen Ujian dibuat seperti alur kerja panitia, bukan dashboard teknis.
						</h1>
						<p class="max-w-3xl text-base leading-7 text-slate-600">
							Prototype ini hanya untuk menilai rasa UI: siapkan ujian, atur peserta dan ruang, cetak kartu, jalankan hari-H, lalu lihat rekap. Bank Soal tetap modul terpisah.
						</p>
					</div>
				</div>
				<div class="rounded-2xl border border-slate-200 bg-slate-50 p-4 text-sm text-slate-700 lg:w-80">
					<p class="font-bold text-slate-950">Mengikuti CBT lama</p>
					<p class="mt-1 leading-6">
						Istilah tetap familiar: Ruang Ujian, Jadwal Sesi, Paket Ujian, Proctoring Live, Rekap Nilai. Detail teknis tetap disembunyikan dulu.
					</p>
				</div>
			</div>
		</header>

		<section class="grid gap-4 lg:grid-cols-[1.1fr_0.9fr]">
			<div class="rounded-3xl border border-slate-200 bg-white p-4 shadow-sm">
				<div class="mb-4 flex flex-wrap items-center justify-between gap-3 border-b border-slate-200 pb-3">
					<div>
						<p class="text-xs font-bold uppercase tracking-[0.18em] text-slate-500">Preview Panitia</p>
						<h2 class="text-xl font-black text-slate-950">Checklist ujian 4 langkah</h2>
					</div>
					<span class="rounded-full bg-emerald-100 px-3 py-1 text-xs font-bold text-emerald-800">Ringkas dulu</span>
				</div>

				<div class="divide-y divide-slate-200 overflow-hidden rounded-2xl border border-slate-200">
					<div class="grid gap-3 bg-emerald-50/70 p-4 md:grid-cols-[9rem_1fr_auto] md:items-center">
						<p class="text-sm font-black text-emerald-800">A1 — Siapkan Ujian</p>
						<div>
							<p class="font-bold text-slate-950">Nama ujian, mapel/kelas, paket dari Bank Soal, tanggal dan jam.</p>
							<p class="text-sm text-slate-600">Operator cukup mengisi yang wajib. Bank Soal tetap modul terpisah sebagai sumber paket.</p>
						</div>
						<span class="rounded-full bg-white px-3 py-1 text-xs font-bold text-emerald-700 shadow-sm">Mulai dari sini</span>
					</div>

					<div class="grid gap-3 p-4 md:grid-cols-[9rem_1fr_auto] md:items-center">
						<p class="text-sm font-black text-sky-800">A2 — Atur Peserta &amp; Ruang</p>
						<div>
							<p class="font-bold text-slate-950">Pilih rombel, buat Ruang Ujian otomatis, preview pembagian.</p>
							<p class="text-sm text-slate-600">Default: otomatis seimbang. Koreksi manual nanti hanya di Mode Lengkap Admin.</p>
						</div>
						<span class="rounded-full bg-sky-50 px-3 py-1 text-xs font-bold text-sky-800">Preview dulu</span>
					</div>

					<div class="grid gap-3 p-4 md:grid-cols-[9rem_1fr_auto] md:items-center">
						<p class="text-sm font-black text-violet-800">A3 — Cetak Kartu &amp; Lembar Pengawas</p>
						<div>
							<p class="font-bold text-slate-950">Kartu Peserta Ujian, Lembar Pengawas Ruang, daftar hadir.</p>
							<p class="text-sm text-slate-600">Semua berbasis QR+PIN, tetapi token mentah tidak ditampilkan di halaman kerja biasa.</p>
						</div>
						<span class="rounded-full bg-violet-50 px-3 py-1 text-xs font-bold text-violet-800">Siap cetak</span>
					</div>

					<div class="grid gap-3 p-4 md:grid-cols-[9rem_1fr_auto] md:items-center">
						<p class="text-sm font-black text-orange-800">A4 — Pelaksanaan Hari-H</p>
						<div>
							<p class="font-bold text-slate-950">Status ruang: belum mulai, berlangsung, selesai, perlu dicek.</p>
							<p class="text-sm text-slate-600">Panitia melihat ringkasan; pengawas ruang memakai portal mobile sederhana.</p>
						</div>
						<span class="rounded-full bg-orange-50 px-3 py-1 text-xs font-bold text-orange-800">Hari-H</span>
					</div>

					<div class="grid gap-3 p-4 md:grid-cols-[9rem_1fr_auto] md:items-center">
						<p class="text-sm font-black text-rose-800">A5 — Hasil &amp; Rekap</p>
						<div>
							<p class="font-bold text-slate-950">Jumlah peserta, submit, nilai PG otomatis, ekspor CSV sederhana.</p>
							<p class="text-sm text-slate-600">Koreksi esai dan sinkron rapor menyusul setelah alur dasar stabil.</p>
						</div>
						<span class="rounded-full bg-rose-50 px-3 py-1 text-xs font-bold text-rose-800">Rekap Nilai</span>
					</div>
				</div>
			</div>

			<aside class="space-y-4">
				<div class="rounded-3xl border border-slate-200 bg-white p-4 shadow-sm">
					<p class="text-xs font-bold uppercase tracking-[0.18em] text-slate-500">Struktur familiar</p>
					<div class="mt-3 grid grid-cols-2 gap-2 text-sm font-bold text-slate-800">
						<span class="rounded-xl bg-slate-100 px-3 py-2">Ruang Ujian</span>
						<span class="rounded-xl bg-slate-100 px-3 py-2">Jadwal Sesi</span>
						<span class="rounded-xl bg-slate-100 px-3 py-2">Paket Ujian</span>
						<span class="rounded-xl bg-slate-100 px-3 py-2">Kartu Peserta</span>
						<span class="rounded-xl bg-slate-100 px-3 py-2">Proctoring Live</span>
						<span class="rounded-xl bg-slate-100 px-3 py-2">Rekap Nilai</span>
					</div>
				</div>

				<div class="rounded-3xl border border-dashed border-slate-300 bg-slate-100/80 p-4 shadow-sm">
					<p class="text-sm font-black text-slate-950">A9 — Mode Lengkap Admin</p>
					<p class="mt-2 text-sm leading-6 text-slate-600">
						Jalur sekunder untuk admin teknis: audit, reset, koreksi manual, dan pengaturan lanjutan. Tidak menjadi menu utama panitia harian.
					</p>
				</div>
			</aside>
		</section>

		<section class="grid gap-4 lg:grid-cols-2">
			<div class="rounded-[2rem] border border-slate-200 bg-white p-4 shadow-sm">
				<div class="mb-4 flex items-center justify-between gap-3">
					<div>
						<p class="text-xs font-bold uppercase tracking-[0.18em] text-slate-500">Preview Peserta</p>
						<h2 class="text-lg font-black text-slate-950">Portal /ujian gaya mobile</h2>
					</div>
					<span class="rounded-full bg-blue-100 px-3 py-1 text-xs font-bold text-blue-800">QR + PIN</span>
				</div>
				<div class="mx-auto max-w-sm rounded-[2rem] border-4 border-slate-900 bg-slate-50 p-3 shadow-xl">
					<div class="rounded-[1.5rem] bg-white p-4">
						<p class="text-xs font-bold uppercase text-slate-500">Kartu Peserta Ujian</p>
						<h3 class="mt-1 text-xl font-black text-slate-950">Masuk Ujian</h3>
						<div class="mt-4 rounded-2xl border border-slate-200 bg-slate-50 p-4 text-center text-sm font-bold text-slate-500">Area scan QR</div>
						<div class="mt-4 text-sm font-bold text-slate-700">
							PIN kartu
							<span class="mt-1 block rounded-xl border border-slate-300 bg-slate-50 px-3 py-3 text-center tracking-[0.4em] text-slate-400">••••••</span>
						</div>
						<button class="mt-4 w-full rounded-2xl bg-emerald-600 px-4 py-3 text-sm font-black text-white shadow-sm">Lanjutkan</button>
						<p class="mt-3 text-xs leading-5 text-slate-500">Jika ujian belum dimulai, peserta melihat pesan tunggu pengawas.</p>
					</div>
				</div>
			</div>

			<div class="rounded-[2rem] border border-slate-200 bg-white p-4 shadow-sm">
				<div class="mb-4 flex items-center justify-between gap-3">
					<div>
						<p class="text-xs font-bold uppercase tracking-[0.18em] text-slate-500">Preview Pengawas</p>
						<h2 class="text-lg font-black text-slate-950">Portal /pengawas-ujian</h2>
					</div>
					<span class="rounded-full bg-emerald-100 px-3 py-1 text-xs font-bold text-emerald-800">Aksi aman</span>
				</div>
				<div class="mx-auto max-w-sm rounded-[2rem] border-4 border-slate-900 bg-slate-50 p-3 shadow-xl">
					<div class="rounded-[1.5rem] bg-white p-4">
						<div class="rounded-2xl bg-emerald-700 p-4 text-white">
							<p class="text-xs font-bold uppercase text-emerald-100">Ruang Ujian</p>
							<p class="text-2xl font-black">Ruang 03</p>
							<p class="text-sm text-emerald-100">Matematika • 32 peserta</p>
						</div>
						<button class="mt-4 w-full rounded-2xl bg-slate-950 px-4 py-3 text-sm font-black text-white">Mulai Ujian</button>
						<div class="mt-4 grid grid-cols-3 gap-2 text-center text-xs font-bold">
							<span class="rounded-xl bg-emerald-50 px-2 py-2 text-emerald-800">Ruang</span>
							<span class="rounded-xl bg-amber-50 px-2 py-2 text-amber-800">Peringatan</span>
							<span class="rounded-xl bg-slate-100 px-2 py-2 text-slate-700">Peserta</span>
						</div>
						<div class="mt-4 space-y-2 text-sm">
							<p class="rounded-xl border border-amber-200 bg-amber-50 px-3 py-2 text-amber-900">1 peserta perlu dicek koneksi.</p>
							<p class="rounded-xl border border-slate-200 px-3 py-2 text-slate-700">Aksi: Sudah Dicek, Beri Peringatan, Hubungi Admin.</p>
						</div>
					</div>
				</div>
			</div>
		</section>
	</div>
</div>
