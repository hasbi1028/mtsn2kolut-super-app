<script lang="ts">
	type KegiatanStatus = 'Draft' | 'Siap' | 'Berlangsung' | 'Selesai' | 'Arsip';

	type KegiatanUjian = {
		id: string;
		nama: string;
		jenis: string;
		periode: string;
		mode: string;
		status: KegiatanStatus;
		peserta: number;
		ruang: number;
		sesi: number;
		catatan: string;
	};

	const kegiatan: KegiatanUjian[] = [
		{
			id: 'uas-genap-2026',
			nama: 'UAS Genap 2025/2026',
			jenis: 'Ujian Semester',
			periode: '3–8 Juni 2026',
			mode: 'CBT Web',
			status: 'Draft',
			peserta: 0,
			ruang: 0,
			sesi: 0,
			catatan: 'Contoh kegiatan awal. Tahap berikutnya akan disambungkan ke data asli.'
		},
		{
			id: 'gladi-cbt',
			nama: 'Gladi CBT Madrasah',
			jenis: 'Simulasi',
			periode: 'Belum dijadwalkan',
			mode: 'Web / Android',
			status: 'Draft',
			peserta: 0,
			ruang: 0,
			sesi: 0,
			catatan: 'Tempat awal untuk simulasi kecil sebelum pelaksanaan resmi.'
		}
	];

	const statusTone: Record<KegiatanStatus, string> = {
		Draft: 'border-amber-200 bg-amber-50 text-amber-700',
		Siap: 'border-emerald-200 bg-emerald-50 text-emerald-700',
		Berlangsung: 'border-sky-200 bg-sky-50 text-sky-700',
		Selesai: 'border-slate-200 bg-slate-50 text-slate-700',
		Arsip: 'border-zinc-200 bg-zinc-50 text-zinc-600'
	};
</script>

<svelte:head>
	<title>Kegiatan Ujian CBT | MTsN 2 Kolut</title>
</svelte:head>

<div class="space-y-5 pb-16">
	<section class="rounded-2xl border border-border bg-card p-5 shadow-sm">
		<div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
			<div class="space-y-2">
				<p class="text-xs font-semibold tracking-[0.22em] text-muted-foreground uppercase">Asesmen / CBT</p>
				<h1 class="text-2xl font-bold tracking-tight text-foreground md:text-3xl">Kegiatan Ujian</h1>
				<p class="max-w-3xl text-sm leading-6 text-muted-foreground">
					Tahap awal dibuat berbasis kegiatan. Operator memilih satu kegiatan ujian dulu, lalu
					nanti paket, peserta, ruang, sesi, cetak, pelaksanaan, dan hasil akan masuk di dalam
					kegiatan tersebut.
				</p>
			</div>
			<button
				type="button"
				class="inline-flex h-10 items-center justify-center rounded-md bg-primary px-4 text-sm font-semibold text-primary-foreground shadow-sm opacity-70"
				title="Tahap berikutnya: form Buat Kegiatan akan diaktifkan setelah alur daftar disetujui."
			>
				Buat Kegiatan
			</button>
		</div>
	</section>

	<section class="grid gap-3 md:grid-cols-4">
		<div class="rounded-xl border bg-background p-4">
			<p class="text-xs font-medium text-muted-foreground">Kegiatan aktif</p>
			<p class="mt-1 text-2xl font-bold">{kegiatan.length}</p>
		</div>
		<div class="rounded-xl border bg-background p-4">
			<p class="text-xs font-medium text-muted-foreground">Peserta terhubung</p>
			<p class="mt-1 text-2xl font-bold">0</p>
		</div>
		<div class="rounded-xl border bg-background p-4">
			<p class="text-xs font-medium text-muted-foreground">Ruang disiapkan</p>
			<p class="mt-1 text-2xl font-bold">0</p>
		</div>
		<div class="rounded-xl border bg-background p-4">
			<p class="text-xs font-medium text-muted-foreground">Sesi dibuat</p>
			<p class="mt-1 text-2xl font-bold">0</p>
		</div>
	</section>

	<section class="rounded-2xl border border-border bg-card shadow-sm">
		<div class="border-b border-border px-4 py-3">
			<h2 class="text-base font-semibold text-foreground">Daftar Kegiatan</h2>
			<p class="text-xs text-muted-foreground">Opsi A: satu halaman pusat kegiatan, belum masuk fitur detail yang kompleks.</p>
		</div>

		<div class="divide-y divide-border">
			{#each kegiatan as item (item.id)}
				<article class="p-4 transition-colors hover:bg-muted/30">
					<div class="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
						<div class="min-w-0 space-y-2">
							<div class="flex flex-wrap items-center gap-2">
								<span class={`rounded-full border px-2.5 py-1 text-xs font-semibold ${statusTone[item.status]}`}>
									{item.status}
								</span>
								<span class="rounded-full border bg-muted px-2.5 py-1 text-xs font-medium text-muted-foreground">
									{item.jenis}
								</span>
								<span class="rounded-full border bg-muted px-2.5 py-1 text-xs font-medium text-muted-foreground">
									{item.mode}
								</span>
							</div>
							<h3 class="truncate text-lg font-bold text-foreground">{item.nama}</h3>
							<p class="text-sm text-muted-foreground">Tanggal: {item.periode}</p>
							<p class="max-w-2xl text-xs leading-5 text-muted-foreground">{item.catatan}</p>
						</div>

						<div class="grid min-w-full grid-cols-3 gap-2 text-center sm:min-w-[18rem]">
							<div class="rounded-lg border bg-background p-2">
								<p class="text-[11px] text-muted-foreground">Peserta</p>
								<p class="text-lg font-bold">{item.peserta}</p>
							</div>
							<div class="rounded-lg border bg-background p-2">
								<p class="text-[11px] text-muted-foreground">Ruang</p>
								<p class="text-lg font-bold">{item.ruang}</p>
							</div>
							<div class="rounded-lg border bg-background p-2">
								<p class="text-[11px] text-muted-foreground">Sesi</p>
								<p class="text-lg font-bold">{item.sesi}</p>
							</div>
						</div>
					</div>
				</article>
			{/each}
		</div>
	</section>

	<section class="rounded-2xl border border-dashed border-border bg-muted/30 p-4">
		<h2 class="text-sm font-semibold text-foreground">Batas tahap ini</h2>
		<ul class="mt-2 list-disc space-y-1 pl-5 text-sm leading-6 text-muted-foreground">
			<li>Belum membuat tabel/database baru.</li>
			<li>Belum mengaktifkan form simpan kegiatan.</li>
			<li>Belum menampilkan paket, peserta, ruang, sesi, kartu, proctoring, atau hasil.</li>
			<li>Tujuannya hanya mengunci arah awal: CBT dimulai dari daftar Kegiatan.</li>
		</ul>
	</section>
</div>
