<script lang="ts">
	import { onMount } from 'svelte';
	import { readClientApiData } from '$lib/client/api';

	type KegiatanStatus = 'Draft' | 'Siap' | 'Berlangsung' | 'Selesai' | 'Arsip';

	type AssessmentExam = {
		id: string;
		title: string;
		status: 'draft' | 'ready' | 'running' | 'finished' | 'archived';
		starts_at?: string;
		ends_at?: string;
		created_at: string;
		session_count: number;
		room_count: number;
		participant_count: number;
		card_count?: number;
	};

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
		kartu: number;
	};

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

	let kegiatan = $state<KegiatanUjian[]>([]);
	let showCreateForm = $state(false);
	let draft = $state<DraftKegiatan>({ ...emptyDraft });
	let formError = $state('');
	let formNotice = $state('');
	let listError = $state('');
	let loading = $state(true);
	let saving = $state(false);

	const totalPeserta = $derived(kegiatan.reduce((total, item) => total + item.peserta, 0));
	const totalRuang = $derived(kegiatan.reduce((total, item) => total + item.ruang, 0));
	const totalSesi = $derived(kegiatan.reduce((total, item) => total + item.sesi, 0));
	const kegiatanAktif = $derived(kegiatan.filter((item) => item.status === 'Draft' || item.status === 'Siap' || item.status === 'Berlangsung').length);

	const statusTone: Record<KegiatanStatus, string> = {
		Draft: 'border-amber-200 bg-amber-50 text-amber-700',
		Siap: 'border-emerald-200 bg-emerald-50 text-emerald-700',
		Berlangsung: 'border-sky-200 bg-sky-50 text-sky-700',
		Selesai: 'border-slate-200 bg-slate-50 text-slate-700',
		Arsip: 'border-zinc-200 bg-zinc-50 text-zinc-600'
	};

	const apiStatusLabel: Record<AssessmentExam['status'], KegiatanStatus> = {
		draft: 'Draft',
		ready: 'Siap',
		running: 'Berlangsung',
		finished: 'Selesai',
		archived: 'Arsip'
	};

	onMount(() => {
		void loadKegiatan();
	});

	function formatDateLabel(value?: string) {
		if (!value) return '';
		const datePart = value.includes('T') ? value.slice(0, 10) : value;
		const [year, month, day] = datePart.split('-');
		if (!year || !month || !day) return value;
		return `${day}/${month}/${year}`;
	}

	function periodeLabelFromDates(startsAt?: string, endsAt?: string) {
		if (!startsAt && !endsAt) return 'Belum dijadwalkan';
		if (startsAt && endsAt) return `${formatDateLabel(startsAt)}–${formatDateLabel(endsAt)}`;
		return formatDateLabel(startsAt || endsAt);
	}

	function dateToIso(value: string) {
		return value ? new Date(`${value}T00:00:00+08:00`).toISOString() : undefined;
	}

	function mapExamToKegiatan(item: AssessmentExam): KegiatanUjian {
		return {
			id: item.id,
			nama: item.title,
			jenis: 'Kegiatan Ujian',
			periode: periodeLabelFromDates(item.starts_at, item.ends_at),
			mode: 'CBT Web',
			status: apiStatusLabel[item.status] ?? 'Draft',
			peserta: Number(item.participant_count ?? 0),
			ruang: Number(item.room_count ?? 0),
			sesi: Number(item.session_count ?? 0),
			catatan: Number(item.card_count ?? 0) > 0
				? `Data tersimpan di database. Kartu peserta terbit: ${item.card_count}.`
				: 'Data tersimpan di database. Kartu/QR+PIN belum diterbitkan.',
			kartu: Number(item.card_count ?? 0)
		};
	}

	async function loadKegiatan() {
		loading = true;
		listError = '';
		try {
			const params = new URLSearchParams({ limit: '50' });
			const response = await fetch(`/api/asesmen/exams?${params}`);
			const items = await readClientApiData<AssessmentExam[]>(response);
			kegiatan = items.map(mapExamToKegiatan);
		} catch (error) {
			listError = error instanceof Error ? error.message : 'Daftar kegiatan belum dapat dibuka.';
			kegiatan = [];
		} finally {
			loading = false;
		}
	}

	function resetDraft() {
		draft = { ...emptyDraft };
		formError = '';
	}

	function toggleCreateForm() {
		showCreateForm = !showCreateForm;
		formNotice = '';
		if (showCreateForm) formError = '';
	}

	async function submitKegiatan() {
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

		saving = true;
		try {
			const title = `${draft.nama.trim()} ${draft.semester} ${draft.tahunAjaran}`.trim();
			const response = await fetch('/api/asesmen/exams', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					title,
					starts_at: dateToIso(draft.tanggalMulai),
					ends_at: dateToIso(draft.tanggalSelesai)
				})
			});
			await readClientApiData<AssessmentExam>(response);
			formNotice = 'Kegiatan tersimpan. Lanjutkan dari daftar kegiatan untuk mengatur peserta, ruang, sesi, cetak, dan hasil.';
			showCreateForm = false;
			resetDraft();
			await loadKegiatan();
		} catch (error) {
			formError = error instanceof Error ? error.message : 'Kegiatan belum dapat disimpan.';
		} finally {
			saving = false;
		}
	}

	function readinessScore(item: KegiatanUjian) {
		let score = 0;
		if (item.status !== 'Arsip') score += 10;
		if (item.peserta > 0) score += 20;
		if (item.ruang > 0) score += 20;
		if (item.sesi > 0) score += 20;
		if (item.kartu > 0) score += 30;
		return Math.min(100, score);
	}

	function readinessTone(score: number) {
		if (score >= 80) return 'bg-emerald-500';
		if (score >= 45) return 'bg-amber-500';
		return 'bg-slate-400';
	}

	function nextActionLabel(item: KegiatanUjian) {
		if (item.peserta <= 0 && item.ruang <= 0) return 'Hubungkan paket siap dan susun peserta/ruang';
		if (item.sesi <= 0) return 'Lengkapi sesi ujian';
		if (item.kartu <= 0) return 'Terbitkan QR+PIN dan kartu';
		if (item.status === 'Berlangsung') return 'Pantau pelaksanaan';
		if (item.status === 'Selesai' || item.status === 'Arsip') return 'Lihat hasil dan arsip';
		return 'Review kesiapan akhir';
	}

	function issueSummary(item: KegiatanUjian) {
		const issues: string[] = [];
		if (item.peserta <= 0) issues.push('peserta belum masuk');
		if (item.ruang <= 0) issues.push('ruang belum tersusun');
		if (item.sesi <= 0) issues.push('sesi belum dibuat');
		if (item.kartu <= 0) issues.push('kartu belum terbit');
		return issues.length ? issues.join(', ') : 'dokumen utama sudah siap';
	}
</script>

<svelte:head>
	<title>Asesmen CBT | MTsN 2 Kolut</title>
</svelte:head>

<div class="space-y-5 pb-16">
	<section class="rounded-2xl border border-border bg-card p-5 shadow-sm">
		<div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
			<div class="space-y-2">
				<p class="text-xs font-semibold tracking-[0.22em] text-muted-foreground uppercase">Asesmen / CBT</p>
				<h1 class="text-2xl font-bold tracking-tight text-foreground md:text-3xl">Daftar Kegiatan Ujian</h1>
				<p class="max-w-3xl text-sm leading-6 text-muted-foreground">
					Asesmen khusus untuk kegiatan ujian: peserta, ruang, sesi, cetak, pelaksanaan, dan hasil. Bank Soal serta Paket Soal berdiri sebagai modul tersendiri; di sini hanya memilih paket yang sudah siap.
				</p>
			</div>
			<div class="flex flex-wrap gap-2">
				<button type="button" class="inline-flex h-10 items-center justify-center rounded-md border px-4 text-sm font-semibold text-foreground hover:bg-muted" onclick={loadKegiatan} disabled={loading}>{loading ? 'Memuat…' : 'Muat Ulang'}</button>
				<button type="button" class="inline-flex h-10 items-center justify-center rounded-md bg-primary px-4 text-sm font-semibold text-primary-foreground shadow-sm transition hover:bg-primary/90" onclick={toggleCreateForm}>{showCreateForm ? 'Tutup Form' : 'Buat Kegiatan'}</button>
			</div>
		</div>
	</section>

	{#if listError}<p class="rounded-xl border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive" role="alert">{listError}</p>{/if}
	{#if formNotice}<p class="rounded-xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm font-medium text-emerald-700" role="status">{formNotice}</p>{/if}

	{#if showCreateForm}
		<section class="rounded-2xl border border-primary/20 bg-card shadow-sm" aria-labelledby="create-kegiatan-title">
			<div class="border-b border-border bg-muted/20 px-5 py-4">
				<div class="flex items-start justify-between gap-3">
					<div class="space-y-1"><p class="text-xs font-semibold tracking-[0.18em] text-primary uppercase">Langkah 1 · Kegiatan</p><h2 id="create-kegiatan-title" class="text-lg font-bold text-foreground">Buat Kegiatan Baru</h2><p class="text-xs leading-5 text-muted-foreground">Kegiatan tersimpan sebagai draft. Setelah itu lanjutkan ke workspace kegiatan untuk peserta, ruang, sesi, cetak, dan hasil.</p></div>
					<button type="button" class="rounded-md border px-3 py-2 text-xs font-semibold text-muted-foreground hover:bg-muted" aria-label="Tutup form" onclick={toggleCreateForm}>Tutup</button>
				</div>
			</div>
			<form onsubmit={(event) => { event.preventDefault(); void submitKegiatan(); }}>
				<div class="grid gap-4 px-5 py-4 lg:grid-cols-2">
					<div class="rounded-xl border border-emerald-200 bg-emerald-50 px-3 py-2 text-xs leading-5 text-emerald-800">Kegiatan akan tersimpan sebagai draft. Kartu/QR+PIN belum diterbitkan pada tahap ini.</div>
					<label class="space-y-1.5"><span class="text-xs font-medium text-muted-foreground">Nama kegiatan</span><input class="min-w-0 w-full rounded-md border border-input bg-background px-3 py-2 text-sm outline-none focus:border-primary" placeholder="Contoh: UAS Genap" bind:value={draft.nama} /></label>
					<div class="grid gap-3 sm:grid-cols-2"><label class="space-y-1.5"><span class="text-xs font-medium text-muted-foreground">Jenis kegiatan</span><select class="min-w-0 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={draft.jenis}><option>Ujian Semester</option><option>Gladi CBT</option><option>Tryout</option><option>Simulasi</option></select></label><label class="space-y-1.5"><span class="text-xs font-medium text-muted-foreground">Mode pelaksanaan</span><select class="min-w-0 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={draft.mode}><option>CBT Web</option><option>Android</option><option>Web / Android</option><option>Kertas / Campuran</option></select></label></div>
					<div class="grid gap-3 sm:grid-cols-2"><label class="space-y-1.5"><span class="text-xs font-medium text-muted-foreground">Tahun ajaran</span><input class="min-w-0 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={draft.tahunAjaran} /></label><label class="space-y-1.5"><span class="text-xs font-medium text-muted-foreground">Semester</span><select class="min-w-0 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={draft.semester}><option>Ganjil</option><option>Genap</option></select></label></div>
					<div class="grid gap-3 sm:grid-cols-2"><label class="space-y-1.5"><span class="text-xs font-medium text-muted-foreground">Tanggal mulai</span><input type="date" class="min-w-0 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={draft.tanggalMulai} /></label><label class="space-y-1.5"><span class="text-xs font-medium text-muted-foreground">Tanggal selesai</span><input type="date" class="min-w-0 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={draft.tanggalSelesai} /></label></div>
					<label class="space-y-1.5"><span class="text-xs font-medium text-muted-foreground">Catatan singkat</span><textarea class="min-h-24 w-full rounded-md border border-input bg-background px-3 py-2 text-sm outline-none focus:border-primary" placeholder="Opsional untuk operator; belum disimpan sebagai kolom khusus." bind:value={draft.catatan}></textarea></label>
					{#if formError}<p class="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive" role="alert">{formError}</p>{/if}
				</div>
				<div class="flex flex-col gap-2 border-t border-border bg-muted/20 px-5 py-4 sm:flex-row sm:justify-end"><button type="button" class="rounded-md border px-4 py-2 text-sm font-semibold text-foreground hover:bg-muted" onclick={resetDraft} disabled={saving}>Reset</button><button type="submit" class="rounded-md bg-primary px-4 py-2 text-sm font-semibold text-primary-foreground hover:bg-primary/90 disabled:opacity-60" disabled={saving}>{saving ? 'Menyimpan…' : 'Simpan Kegiatan'}</button></div>
			</form>
		</section>
	{/if}

	<section class="grid gap-3 md:grid-cols-4">
		<div class="rounded-2xl border bg-card p-4 shadow-sm"><p class="text-xs text-muted-foreground">Kegiatan aktif</p><p class="mt-1 text-2xl font-bold text-foreground">{kegiatanAktif}</p></div>
		<div class="rounded-2xl border bg-card p-4 shadow-sm"><p class="text-xs text-muted-foreground">Total peserta</p><p class="mt-1 text-2xl font-bold text-foreground">{totalPeserta}</p></div>
		<div class="rounded-2xl border bg-card p-4 shadow-sm"><p class="text-xs text-muted-foreground">Ruang tersusun</p><p class="mt-1 text-2xl font-bold text-foreground">{totalRuang}</p></div>
		<div class="rounded-2xl border bg-card p-4 shadow-sm"><p class="text-xs text-muted-foreground">Sesi dibuat</p><p class="mt-1 text-2xl font-bold text-foreground">{totalSesi}</p></div>
	</section>

	<section class="rounded-2xl border border-border bg-card shadow-sm">
		<div class="flex flex-col gap-2 border-b border-border px-4 py-3 sm:flex-row sm:items-center sm:justify-between">
			<div>
				<h2 class="text-base font-semibold text-foreground">Kegiatan Operasional</h2>
				<p class="text-xs text-muted-foreground">Pilih Lanjutkan untuk membuka workspace penuh satu kegiatan.</p>
			</div>
			<button type="button" class="rounded-md border bg-background px-3 py-2 text-xs font-semibold text-foreground hover:bg-muted" onclick={toggleCreateForm}>Buat Kegiatan</button>
		</div>
		{#if loading}
			<p class="px-4 py-8 text-center text-sm text-muted-foreground">Memuat kegiatan…</p>
		{:else if kegiatan.length === 0}
			<div class="px-4 py-8 text-center"><p class="text-sm font-semibold text-foreground">Belum ada kegiatan.</p><p class="mt-1 text-xs text-muted-foreground">Klik Buat Kegiatan untuk membuat draft pertama.</p></div>
		{:else}
			<div class="divide-y divide-border">
				{#each kegiatan as item (item.id)}
					{@const score = readinessScore(item)}
					<article class="grid gap-4 px-4 py-4 lg:grid-cols-[minmax(0,1fr)_14rem_8rem] lg:items-center">
						<div class="min-w-0 space-y-2">
							<div class="flex flex-wrap items-center gap-2"><span class={`rounded-full border px-2.5 py-1 text-xs font-semibold ${statusTone[item.status]}`}>{item.status}</span><span class="rounded-full border bg-muted px-2.5 py-1 text-xs font-medium text-muted-foreground">{item.mode}</span><span class="text-xs text-muted-foreground">{item.periode}</span></div>
							<h3 class="truncate text-base font-bold text-foreground">{item.nama}</h3>
							<p class="text-xs leading-5 text-muted-foreground">Langkah berikutnya: <strong class="text-foreground">{nextActionLabel(item)}</strong></p>
							<p class="text-xs leading-5 text-muted-foreground">Catatan: {issueSummary(item)}</p>
						</div>
						<div class="space-y-2">
							<div class="flex items-center justify-between text-xs"><span class="font-semibold text-foreground">Kesiapan</span><span class="text-muted-foreground">{score}%</span></div>
							<div class="h-2 overflow-hidden rounded-full bg-muted"><div class={`h-full rounded-full ${readinessTone(score)}`} style={`width: ${score}%`}></div></div>
							<div class="grid grid-cols-4 gap-1 text-center text-[11px] text-muted-foreground"><span>{item.peserta} peserta</span><span>{item.ruang} ruang</span><span>{item.sesi} sesi</span><span>{item.kartu} kartu</span></div>
						</div>
						<a class="inline-flex h-10 items-center justify-center rounded-md bg-primary px-4 text-sm font-semibold text-primary-foreground hover:bg-primary/90" aria-label={`Kelola ${item.nama}`} href={`/asesmen/kegiatan/${item.id}`}>Lanjutkan</a>
					</article>
				{/each}
			</div>
		{/if}
	</section>

	<section class="rounded-2xl border border-dashed border-border bg-muted/20 p-4"><h2 class="text-sm font-semibold text-foreground">Catatan alur</h2><p class="mt-1 text-sm leading-6 text-muted-foreground">Bank Soal dan Paket Soal adalah modul mandiri. Asesmen hanya memakai paket yang sudah siap untuk mengatur peserta, ruang, sesi, cetak, pelaksanaan, dan hasil.</p></section>
</div>
