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

	type DetailFeatureKey = 'paket' | 'peserta' | 'ruang' | 'sesi' | 'cetak' | 'hasil';

	type DraftKegiatan = {
		nama: string;
		jenis: string;
		tahunAjaran: string;
		semester: string;
		tanggalMulai: string;
		tanggalSelesai: string;
		mode: string;
		catatan: string;
	};

	const emptyDraft: DraftKegiatan = {
		nama: '',
		jenis: 'Ujian Semester',
		tahunAjaran: '2025/2026',
		semester: 'Genap',
		tanggalMulai: '',
		tanggalSelesai: '',
		mode: 'CBT Web',
		catatan: ''
	};

	let kegiatan = $state<KegiatanUjian[]>([
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
	]);

	let showCreateForm = $state(false);
	let selectedKegiatanId = $state<string | null>(null);
	let activeDetailFeature = $state<DetailFeatureKey | null>(null);
	let draft = $state<DraftKegiatan>({ ...emptyDraft });
	let formError = $state('');
	let formNotice = $state('');

	const totalPeserta = $derived(kegiatan.reduce((total, item) => total + item.peserta, 0));
	const totalRuang = $derived(kegiatan.reduce((total, item) => total + item.ruang, 0));
	const totalSesi = $derived(kegiatan.reduce((total, item) => total + item.sesi, 0));
	const selectedKegiatan = $derived(kegiatan.find((item) => item.id === selectedKegiatanId) ?? null);

	const preparationChecklist = [
		'Data kegiatan lengkap',
		'Paket soal dipilih',
		'Peserta ditambahkan',
		'Ruang disiapkan',
		'Jadwal sesi dibuat',
		'Kartu peserta siap cetak',
		'Lembar pengawas siap cetak',
		'Siap pelaksanaan'
	] as const;

	const detailFeatures: Array<{ key: DetailFeatureKey; label: string; description: string }> = [
		{ key: 'paket', label: 'Paket Soal', description: 'Nanti untuk memilih/menautkan paket soal ke kegiatan.' },
		{ key: 'peserta', label: 'Peserta', description: 'Nanti untuk menambahkan siswa peserta kegiatan.' },
		{ key: 'ruang', label: 'Ruang', description: 'Nanti untuk menyiapkan ruang, kapasitas, dan tempat duduk.' },
		{ key: 'sesi', label: 'Sesi', description: 'Nanti untuk jadwal sesi ujian per ruang/paket.' },
		{ key: 'cetak', label: 'Cetak', description: 'Nanti untuk kartu peserta dan lembar pengawas.' },
		{ key: 'hasil', label: 'Hasil', description: 'Nanti untuk rekap nilai dan arsip pelaksanaan.' }
	];

	const activeFeature = $derived(detailFeatures.find((feature) => feature.key === activeDetailFeature) ?? null);

	const statusTone: Record<KegiatanStatus, string> = {
		Draft: 'border-amber-200 bg-amber-50 text-amber-700',
		Siap: 'border-emerald-200 bg-emerald-50 text-emerald-700',
		Berlangsung: 'border-sky-200 bg-sky-50 text-sky-700',
		Selesai: 'border-slate-200 bg-slate-50 text-slate-700',
		Arsip: 'border-zinc-200 bg-zinc-50 text-zinc-600'
	};

	function formatDateLabel(value: string) {
		if (!value) return '';
		const [year, month, day] = value.split('-');
		if (!year || !month || !day) return value;
		return `${day}/${month}/${year}`;
	}

	function periodeLabel() {
		if (!draft.tanggalMulai && !draft.tanggalSelesai) return 'Belum dijadwalkan';
		if (draft.tanggalMulai && draft.tanggalSelesai) {
			return `${formatDateLabel(draft.tanggalMulai)}–${formatDateLabel(draft.tanggalSelesai)}`;
		}
		return formatDateLabel(draft.tanggalMulai || draft.tanggalSelesai);
	}

	function slugify(value: string) {
		return value
			.toLowerCase()
			.trim()
			.replace(/[^a-z0-9]+/g, '-')
			.replace(/(^-|-$)/g, '') || `kegiatan-${Date.now()}`;
	}

	function resetDraft() {
		draft = { ...emptyDraft };
		formError = '';
	}

	function toggleCreateForm() {
		showCreateForm = !showCreateForm;
		if (showCreateForm) selectedKegiatanId = null;
		formNotice = '';
		if (showCreateForm) formError = '';
	}

	function openKegiatanDetail(id: string) {
		selectedKegiatanId = id;
		activeDetailFeature = null;
		showCreateForm = false;
		formNotice = '';
	}

	function closeKegiatanDetail() {
		selectedKegiatanId = null;
		activeDetailFeature = null;
	}

	function submitPreview() {
		formError = '';
		formNotice = '';

		if (!draft.nama.trim()) {
			formError = 'Nama kegiatan wajib diisi.';
			return;
		}
		if (draft.tanggalMulai && draft.tanggalSelesai && draft.tanggalMulai > draft.tanggalSelesai) {
			formError = 'Tanggal selesai tidak boleh lebih awal dari tanggal mulai.';
			return;
		}

		const baseId = slugify(draft.nama);
		const id = kegiatan.some((item) => item.id === baseId) ? `${baseId}-${kegiatan.length + 1}` : baseId;
		const namaDenganPeriode = `${draft.nama.trim()} ${draft.semester} ${draft.tahunAjaran}`.trim();

		kegiatan = [
			{
				id,
				nama: namaDenganPeriode,
				jenis: draft.jenis,
				periode: periodeLabel(),
				mode: draft.mode,
				status: 'Draft',
				peserta: 0,
				ruang: 0,
				sesi: 0,
				catatan: draft.catatan.trim() || 'Draft lokal dari form awal. Belum tersimpan ke database.'
			},
			...kegiatan
		];
		formNotice = 'Kegiatan ditambahkan sebagai preview lokal. Belum tersimpan ke database.';
		showCreateForm = false;
		resetDraft();
	}
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
				class="inline-flex h-10 items-center justify-center rounded-md bg-primary px-4 text-sm font-semibold text-primary-foreground shadow-sm transition hover:bg-primary/90"
				onclick={toggleCreateForm}
			>
				{showCreateForm ? 'Tutup Form' : 'Buat Kegiatan'}
			</button>
		</div>
	</section>

	{#if showCreateForm}
		<div class="fixed inset-0 z-50 flex justify-end" role="dialog" aria-modal="true" aria-labelledby="drawer-title">
			<button
				type="button"
				class="absolute inset-0 bg-slate-950/35 backdrop-blur-[1px]"
				aria-label="Tutup form buat kegiatan"
				onclick={toggleCreateForm}
			></button>

			<aside class="relative flex h-full w-full max-w-xl flex-col border-l border-border bg-card shadow-2xl sm:w-[34rem]">
				<div class="border-b border-border px-5 py-4">
					<div class="flex items-start justify-between gap-3">
						<div class="space-y-1">
							<p class="text-xs font-semibold tracking-[0.18em] text-muted-foreground uppercase">Step 2 · Preview Lokal</p>
							<h2 id="drawer-title" class="text-lg font-bold text-foreground">Buat Kegiatan Baru</h2>
							<p class="text-xs leading-5 text-muted-foreground">
								Isi data dasar kegiatan. Form ini belum menulis database/API, hanya menambah
								preview lokal di daftar.
							</p>
						</div>
						<button
							type="button"
							class="inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-full border text-sm font-bold text-muted-foreground hover:bg-muted"
							aria-label="Tutup"
							onclick={toggleCreateForm}
						>
							×
						</button>
					</div>
				</div>

				<form class="flex min-h-0 flex-1 flex-col" onsubmit={(event) => { event.preventDefault(); submitPreview(); }}>
					<div class="min-h-0 flex-1 space-y-4 overflow-y-auto px-5 py-4">
						<div class="rounded-xl border border-amber-200 bg-amber-50 px-3 py-2 text-xs leading-5 text-amber-800">
							Tahap ini untuk menyepakati bentuk input dulu. Data akan hilang saat halaman dimuat ulang.
						</div>

						<label class="space-y-1.5">
							<span class="text-xs font-medium text-muted-foreground">Nama kegiatan</span>
							<input
								class="min-w-0 w-full rounded-md border border-input bg-background px-3 py-2 text-sm outline-none focus:border-primary"
								placeholder="Contoh: UAS Genap"
								bind:value={draft.nama}
							/>
						</label>

						<div class="grid gap-3 sm:grid-cols-2">
							<label class="space-y-1.5">
								<span class="text-xs font-medium text-muted-foreground">Jenis kegiatan</span>
								<select class="min-w-0 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={draft.jenis}>
									<option>Ujian Semester</option>
									<option>Gladi CBT</option>
									<option>Tryout</option>
									<option>Simulasi</option>
								</select>
							</label>
							<label class="space-y-1.5">
								<span class="text-xs font-medium text-muted-foreground">Mode pelaksanaan</span>
								<select class="min-w-0 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={draft.mode}>
									<option>CBT Web</option>
									<option>Android</option>
									<option>Web / Android</option>
									<option>Kertas / Campuran</option>
								</select>
							</label>
						</div>

						<div class="grid gap-3 sm:grid-cols-2">
							<label class="space-y-1.5">
								<span class="text-xs font-medium text-muted-foreground">Tahun ajaran</span>
								<input class="min-w-0 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={draft.tahunAjaran} />
							</label>
							<label class="space-y-1.5">
								<span class="text-xs font-medium text-muted-foreground">Semester</span>
								<select class="min-w-0 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={draft.semester}>
									<option>Ganjil</option>
									<option>Genap</option>
								</select>
							</label>
						</div>

						<div class="grid gap-3 sm:grid-cols-2">
							<label class="space-y-1.5">
								<span class="text-xs font-medium text-muted-foreground">Tanggal mulai</span>
								<input type="date" class="min-w-0 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={draft.tanggalMulai} />
							</label>
							<label class="space-y-1.5">
								<span class="text-xs font-medium text-muted-foreground">Tanggal selesai</span>
								<input type="date" class="min-w-0 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={draft.tanggalSelesai} />
							</label>
						</div>

						<label class="space-y-1.5">
							<span class="text-xs font-medium text-muted-foreground">Catatan singkat</span>
							<textarea
								class="min-h-24 w-full rounded-md border border-input bg-background px-3 py-2 text-sm outline-none focus:border-primary"
								placeholder="Opsional: misalnya untuk kelas IX atau simulasi internal"
								bind:value={draft.catatan}
							></textarea>
						</label>

						{#if formError}
							<p class="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive" role="alert">{formError}</p>
						{/if}
					</div>

					<div class="flex flex-col gap-2 border-t border-border bg-card px-5 py-4 sm:flex-row sm:justify-end">
						<button type="button" class="rounded-md border px-4 py-2 text-sm font-semibold text-foreground hover:bg-muted" onclick={resetDraft}>Reset</button>
						<button type="submit" class="rounded-md bg-primary px-4 py-2 text-sm font-semibold text-primary-foreground hover:bg-primary/90">
							Tambah Preview Lokal
						</button>
					</div>
				</form>
			</aside>
		</div>
	{/if}

	{#if selectedKegiatan}
		<div class="fixed inset-0 z-50 flex justify-end" role="dialog" aria-modal="true" aria-labelledby="detail-drawer-title">
			<button
				type="button"
				class="absolute inset-0 bg-slate-950/35 backdrop-blur-[1px]"
				aria-label="Tutup detail kegiatan"
				onclick={closeKegiatanDetail}
			></button>

			<aside class="relative flex h-full w-full max-w-xl flex-col border-l border-border bg-card shadow-2xl sm:w-[34rem]">
				<div class="border-b border-border px-5 py-4">
					<div class="flex items-start justify-between gap-3">
						<div class="min-w-0 space-y-2">
							<p class="text-xs font-semibold tracking-[0.18em] text-muted-foreground uppercase">Step 4 · Detail Kegiatan</p>
							<h2 id="detail-drawer-title" class="truncate text-lg font-bold text-foreground">{selectedKegiatan.nama}</h2>
							<div class="flex flex-wrap items-center gap-2">
								<span class={`rounded-full border px-2.5 py-1 text-xs font-semibold ${statusTone[selectedKegiatan.status]}`}>{selectedKegiatan.status}</span>
								<span class="rounded-full border bg-muted px-2.5 py-1 text-xs font-medium text-muted-foreground">{selectedKegiatan.jenis}</span>
								<span class="rounded-full border bg-muted px-2.5 py-1 text-xs font-medium text-muted-foreground">{selectedKegiatan.mode}</span>
							</div>
						</div>
						<button
							type="button"
							class="inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-full border text-sm font-bold text-muted-foreground hover:bg-muted"
							aria-label="Tutup detail"
							onclick={closeKegiatanDetail}
						>
							×
						</button>
					</div>
				</div>

				<div class="min-h-0 flex-1 space-y-4 overflow-y-auto px-5 py-4">
					<section class="rounded-xl border bg-background p-4">
						<h3 class="text-sm font-semibold text-foreground">Ringkasan</h3>
						<div class="mt-3 grid gap-2 text-sm">
							<div class="flex justify-between gap-3"><span class="text-muted-foreground">Tanggal</span><strong class="text-right font-semibold">{selectedKegiatan.periode}</strong></div>
							<div class="flex justify-between gap-3"><span class="text-muted-foreground">Peserta</span><strong>{selectedKegiatan.peserta}</strong></div>
							<div class="flex justify-between gap-3"><span class="text-muted-foreground">Ruang</span><strong>{selectedKegiatan.ruang}</strong></div>
							<div class="flex justify-between gap-3"><span class="text-muted-foreground">Sesi</span><strong>{selectedKegiatan.sesi}</strong></div>
						</div>
						<p class="mt-3 rounded-lg bg-muted/50 px-3 py-2 text-xs leading-5 text-muted-foreground">{selectedKegiatan.catatan}</p>
					</section>

					<section class="rounded-xl border bg-background p-4">
						<h3 class="text-sm font-semibold text-foreground">Menu Dalam Kegiatan</h3>
						<p class="mt-1 text-xs leading-5 text-muted-foreground">Launcher kecil untuk menguji struktur turunan kegiatan. Semua masih placeholder, belum membuka route/API.</p>
						<div class="mt-3 grid gap-2 sm:grid-cols-2">
							{#each detailFeatures as feature}
								<button
									type="button"
									class={`rounded-lg border px-3 py-2 text-left text-sm transition ${activeDetailFeature === feature.key ? 'border-primary bg-primary/10 text-primary' : 'bg-card text-foreground hover:bg-muted'}`}
									onclick={() => (activeDetailFeature = feature.key)}
								>
									<span class="block font-semibold">{feature.label}</span>
									<span class="mt-1 block text-[11px] leading-4 text-muted-foreground">{feature.description}</span>
								</button>
							{/each}
						</div>

						{#if activeFeature}
							<div class="mt-3 rounded-lg border border-dashed bg-muted/30 px-3 py-3">
								<p class="text-sm font-semibold text-foreground">{activeFeature.label}</p>
								<p class="mt-1 text-xs leading-5 text-muted-foreground">{activeFeature.description} Pada step ini hanya dipakai untuk review arah menu, belum ada data asli.</p>
							</div>
						{/if}
					</section>

					<section class="rounded-xl border bg-background p-4">
						<h3 class="text-sm font-semibold text-foreground">Checklist Persiapan</h3>
						<p class="mt-1 text-xs leading-5 text-muted-foreground">Belum bisa dicentang permanen. Ini hanya kerangka alur sebelum backend dibuat.</p>
						<div class="mt-3 space-y-2">
							{#each preparationChecklist as label, index}
								<div class="flex items-center gap-3 rounded-lg border bg-card px-3 py-2 text-sm">
									<span class="inline-flex h-6 w-6 shrink-0 items-center justify-center rounded-full border text-[11px] font-semibold text-muted-foreground">{index + 1}</span>
									<span class="min-w-0 flex-1 text-foreground">{label}</span>
									<span class="rounded-full bg-muted px-2 py-0.5 text-[11px] font-medium text-muted-foreground">Nanti</span>
								</div>
							{/each}
						</div>
					</section>

					<section class="rounded-xl border border-dashed bg-muted/30 p-4">
						<h3 class="text-sm font-semibold text-foreground">Batas Step 4</h3>
						<p class="mt-1 text-xs leading-5 text-muted-foreground">Detail ini masih shell frontend-only. Menu paket, peserta, ruang, sesi, cetak, dan hasil sudah berupa placeholder untuk validasi alur, belum membuka route/API.</p>
					</section>
				</div>

				<div class="border-t border-border bg-card px-5 py-4">
					<button type="button" class="w-full rounded-md border px-4 py-2 text-sm font-semibold text-foreground hover:bg-muted" onclick={closeKegiatanDetail}>Tutup Detail</button>
				</div>
			</aside>
		</div>
	{/if}

	{#if formNotice}
		<p class="rounded-xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm font-medium text-emerald-700" role="status">{formNotice}</p>
	{/if}

	<section class="grid gap-3 md:grid-cols-4">
		<div class="rounded-xl border bg-background p-4">
			<p class="text-xs font-medium text-muted-foreground">Kegiatan aktif</p>
			<p class="mt-1 text-2xl font-bold">{kegiatan.length}</p>
		</div>
		<div class="rounded-xl border bg-background p-4">
			<p class="text-xs font-medium text-muted-foreground">Peserta terhubung</p>
			<p class="mt-1 text-2xl font-bold">{totalPeserta}</p>
		</div>
		<div class="rounded-xl border bg-background p-4">
			<p class="text-xs font-medium text-muted-foreground">Ruang disiapkan</p>
			<p class="mt-1 text-2xl font-bold">{totalRuang}</p>
		</div>
		<div class="rounded-xl border bg-background p-4">
			<p class="text-xs font-medium text-muted-foreground">Sesi dibuat</p>
			<p class="mt-1 text-2xl font-bold">{totalSesi}</p>
		</div>
	</section>

	<section class="rounded-2xl border border-border bg-card shadow-sm">
		<div class="border-b border-border px-4 py-3">
			<h2 class="text-base font-semibold text-foreground">Daftar Kegiatan</h2>
			<p class="text-xs text-muted-foreground">Step 4: klik Kelola untuk melihat detail, checklist, dan menu turunan placeholder.</p>
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

						<div class="space-y-2 sm:min-w-[18rem]">
							<div class="grid min-w-full grid-cols-3 gap-2 text-center">
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
							<button
								type="button"
								class="w-full rounded-md border bg-background px-4 py-2 text-sm font-semibold text-foreground hover:bg-muted"
								onclick={() => openKegiatanDetail(item.id)}
							>
								Kelola
							</button>
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
			<li>Form hanya membuat preview lokal di browser, belum tersimpan permanen.</li>
			<li>Detail kegiatan sudah berupa shell/drawer, belum menyimpan checklist permanen.</li>
			<li>Menu paket, peserta, ruang, sesi, cetak, dan hasil masih placeholder tanpa route/API.</li>
		</ul>
	</section>
</div>
