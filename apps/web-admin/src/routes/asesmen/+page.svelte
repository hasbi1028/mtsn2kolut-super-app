<script lang="ts">
	import { onMount } from 'svelte';
	import { clientApiPath, readClientApiData } from '$lib/client/api';

	type KegiatanStatus = 'Draft' | 'Siap' | 'Berlangsung' | 'Selesai' | 'Arsip';
	type DetailFeatureKey = 'paket' | 'peserta' | 'ruang' | 'sesi' | 'cetak' | 'hasil';

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

	type AcademicRombel = {
		id: string;
		code?: string;
		name?: string;
		level?: string;
		is_active?: boolean;
		total_students?: number;
	};

	type AssignmentRoom = {
		code: string;
		name: string;
		capacity: number;
		assigned_count: number;
	};

	type AssignmentResult = {
		exam_id: string;
		session_id?: string;
		mix_policy: 'mixed' | 'class_grouped';
		room_count: number;
		capacity_per_room: number;
		total_participants: number;
		assigned_total: number;
		unassigned_total: number;
		participant_count: number;
		card_count: number;
		rooms: AssignmentRoom[];
		message: string;
		applied: boolean;
	};

	type ParticipantPlacement = {
		participant_id: string;
		room_id?: string;
		student_name: string;
		nis?: string;
		class_code: string;
		class_name: string;
		room_code?: string;
		room_name?: string;
		room_capacity?: number;
		seat_no?: number;
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
	let rombelOptions = $state<AcademicRombel[]>([]);
	let showCreateForm = $state(false);
	let selectedKegiatanId = $state<string | null>(null);
	let activeDetailFeature = $state<DetailFeatureKey | null>(null);
	let draft = $state<DraftKegiatan>({ ...emptyDraft });
	let selectedClassIds = $state<string[]>([]);
	let roomCount = $state(8);
	let capacityPerRoom = $state(30);
	let mixPolicy = $state<'mixed' | 'class_grouped'>('mixed');
	let assignmentPreview = $state<AssignmentResult | null>(null);
	let participantPlacements = $state<ParticipantPlacement[]>([]);
	let workingParticipantId = $state('');
	let formError = $state('');
	let formNotice = $state('');
	let listError = $state('');
	let assignmentError = $state('');
	let assignmentNotice = $state('');
	let placementError = $state('');
	let loading = $state(true);
	let saving = $state(false);
	let loadingRombel = $state(false);
	let workingAssignment = $state(false);
	let loadingPlacements = $state(false);

	const totalPeserta = $derived(kegiatan.reduce((total, item) => total + item.peserta, 0));
	const totalRuang = $derived(kegiatan.reduce((total, item) => total + item.ruang, 0));
	const totalSesi = $derived(kegiatan.reduce((total, item) => total + item.sesi, 0));
	const selectedKegiatan = $derived(kegiatan.find((item) => item.id === selectedKegiatanId) ?? null);
	const selectedRombel = $derived(rombelOptions.filter((item) => selectedClassIds.includes(item.id)));
	const selectedStudentTotal = $derived(selectedRombel.reduce((total, item) => total + Number(item.total_students ?? 0), 0));
	const manualRoomOptions = $derived(Array.from(new Map(participantPlacements.filter((item) => item.room_id).map((item) => [item.room_id, item])).values()));

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
		{ key: 'paket', label: 'Paket Soal', description: 'Menautkan paket soal Bank Soal ke kegiatan.' },
		{ key: 'peserta', label: 'Peserta', description: 'Menambahkan siswa peserta dari rombel.' },
		{ key: 'ruang', label: 'Ruang', description: 'Menyiapkan ruang, kapasitas, dan tempat duduk.' },
		{ key: 'sesi', label: 'Sesi', description: 'Jadwal sesi ujian per ruang/paket.' },
		{ key: 'cetak', label: 'Cetak', description: 'Kartu peserta dan lembar pengawas.' },
		{ key: 'hasil', label: 'Hasil', description: 'Rekap nilai dan arsip pelaksanaan.' }
	];

	const activeFeature = $derived(detailFeatures.find((feature) => feature.key === activeDetailFeature) ?? null);

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
		void loadRombelOptions();
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
				: 'Data tersimpan di database. Kartu/QR+PIN belum diterbitkan.'
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

	async function loadRombelOptions() {
		loadingRombel = true;
		try {
			const response = await fetch('/api/academic/rombel');
			const items = await readClientApiData<AcademicRombel[]>(response);
			rombelOptions = items.filter((item) => item.id && item.is_active !== false);
		} catch (error) {
			assignmentError = error instanceof Error ? error.message : 'Daftar rombel belum dapat dibuka.';
			rombelOptions = [];
		} finally {
			loadingRombel = false;
		}
	}

	function resetDraft() {
		draft = { ...emptyDraft };
		formError = '';
	}

	function resetAssignmentState() {
		assignmentPreview = null;
		participantPlacements = [];
		assignmentError = '';
		assignmentNotice = '';
		placementError = '';
	}

	function toggleCreateForm() {
		showCreateForm = !showCreateForm;
		if (showCreateForm) selectedKegiatanId = null;
		formNotice = '';
		if (showCreateForm) formError = '';
	}

	function openKegiatanDetail(id: string) {
		selectedKegiatanId = id;
		activeDetailFeature = 'ruang';
		showCreateForm = false;
		formNotice = '';
		resetAssignmentState();
		if (rombelOptions.length === 0) void loadRombelOptions();
	}

	function closeKegiatanDetail() {
		selectedKegiatanId = null;
		activeDetailFeature = null;
		resetAssignmentState();
	}

	function toggleClass(id: string) {
		assignmentPreview = null;
		assignmentNotice = '';
		selectedClassIds = selectedClassIds.includes(id)
			? selectedClassIds.filter((item) => item !== id)
			: [...selectedClassIds, id];
	}

	function assignmentPayload() {
		return {
			room_count: Number(roomCount),
			capacity_per_room: Number(capacityPerRoom),
			mix_policy: mixPolicy,
			class_ids: selectedClassIds
		};
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
			formNotice = 'Kegiatan tersimpan ke database asesmen.';
			showCreateForm = false;
			resetDraft();
			await loadKegiatan();
		} catch (error) {
			formError = error instanceof Error ? error.message : 'Kegiatan belum dapat disimpan.';
		} finally {
			saving = false;
		}
	}

	async function previewAssignment() {
		if (!selectedKegiatan) return;
		assignmentError = '';
		assignmentNotice = '';
		workingAssignment = true;
		try {
			const response = await fetch(clientApiPath`/api/asesmen/exams/${selectedKegiatan.id}/assignment-preview`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(assignmentPayload())
			});
			assignmentPreview = await readClientApiData<AssignmentResult>(response);
		} catch (error) {
			assignmentError = error instanceof Error ? error.message : 'Preview pembagian ruang belum dapat dibuat.';
		} finally {
			workingAssignment = false;
		}
	}

	async function applyAssignment() {
		if (!selectedKegiatan) return;
		assignmentError = '';
		assignmentNotice = '';
		workingAssignment = true;
		try {
			const response = await fetch(clientApiPath`/api/asesmen/exams/${selectedKegiatan.id}/assignment-apply`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(assignmentPayload())
			});
			assignmentPreview = await readClientApiData<AssignmentResult>(response);
			assignmentNotice = 'Ruang dan peserta tersimpan. Kartu/QR+PIN belum diterbitkan.';
			await loadParticipantPlacements();
			await loadKegiatan();
		} catch (error) {
			assignmentError = error instanceof Error ? error.message : 'Ruang dan peserta belum dapat disimpan.';
		} finally {
			workingAssignment = false;
		}
	}

	async function loadParticipantPlacements() {
		if (!selectedKegiatan) return;
		placementError = '';
		loadingPlacements = true;
		try {
			const response = await fetch(clientApiPath`/api/asesmen/exams/${selectedKegiatan.id}/participants`);
			participantPlacements = await readClientApiData<ParticipantPlacement[]>(response);
		} catch (error) {
			placementError = error instanceof Error ? error.message : 'Daftar peserta ruang belum dapat dibuka.';
		} finally {
			loadingPlacements = false;
		}
	}

	function shuffleManualView() {
		participantPlacements = [...participantPlacements].sort(() => Math.random() - 0.5);
	}

	async function moveParticipantSeat(participant: ParticipantPlacement, roomId: string, seatNo: number) {
		if (!selectedKegiatan || !roomId || !seatNo) return;
		placementError = '';
		workingParticipantId = participant.participant_id;
		try {
			const response = await fetch(clientApiPath`/api/asesmen/exams/${selectedKegiatan.id}/participants/seat`, {
				method: 'PATCH',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ participant_id: participant.participant_id, room_id: roomId, seat_no: Number(seatNo) })
			});
			const updated = await readClientApiData<ParticipantPlacement>(response);
			participantPlacements = participantPlacements.map((item) => item.participant_id === updated.participant_id ? updated : item);
			assignmentNotice = `Peserta ${updated.student_name} dipindah ke ${updated.room_code || 'ruang baru'} kursi ${updated.seat_no || '-'}.`;
		} catch (error) {
			placementError = error instanceof Error ? error.message : 'Perubahan ruang/kursi belum dapat disimpan.';
		} finally {
			workingParticipantId = '';
		}
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
					Step 6: kegiatan, ruang, dan peserta sudah tersambung ke API Asesmen. Operator dapat preview pembagian ruang dulu sebelum menyimpan peserta dan kursi.
				</p>
			</div>
			<div class="flex flex-wrap gap-2">
				<button type="button" class="inline-flex h-10 items-center justify-center rounded-md border px-4 text-sm font-semibold text-foreground hover:bg-muted" onclick={loadKegiatan} disabled={loading}>{loading ? 'Memuat…' : 'Muat Ulang'}</button>
				<button type="button" class="inline-flex h-10 items-center justify-center rounded-md bg-primary px-4 text-sm font-semibold text-primary-foreground shadow-sm transition hover:bg-primary/90" onclick={toggleCreateForm}>{showCreateForm ? 'Tutup Form' : 'Buat Kegiatan'}</button>
			</div>
		</div>
	</section>

	{#if listError}<p class="rounded-xl border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive" role="alert">{listError}</p>{/if}

	{#if showCreateForm}
		<div class="fixed inset-0 z-50 flex justify-end" role="dialog" aria-modal="true" aria-labelledby="drawer-title">
			<button type="button" class="absolute inset-0 bg-slate-950/35 backdrop-blur-[1px]" aria-label="Tutup form buat kegiatan" onclick={toggleCreateForm}></button>
			<aside class="relative flex h-full w-full max-w-xl flex-col border-l border-border bg-card shadow-2xl sm:w-[34rem]">
				<div class="border-b border-border px-5 py-4">
					<div class="flex items-start justify-between gap-3">
						<div class="space-y-1"><p class="text-xs font-semibold tracking-[0.18em] text-muted-foreground uppercase">Step 5 · Database</p><h2 id="drawer-title" class="text-lg font-bold text-foreground">Buat Kegiatan Baru</h2><p class="text-xs leading-5 text-muted-foreground">Data dasar kegiatan akan tersimpan sebagai draft asesmen.</p></div>
						<button type="button" class="inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-full border text-sm font-bold text-muted-foreground hover:bg-muted" aria-label="Tutup" onclick={toggleCreateForm}>×</button>
					</div>
				</div>
				<form class="flex min-h-0 flex-1 flex-col" onsubmit={(event) => { event.preventDefault(); void submitKegiatan(); }}>
					<div class="min-h-0 flex-1 space-y-4 overflow-y-auto px-5 py-4">
						<div class="rounded-xl border border-emerald-200 bg-emerald-50 px-3 py-2 text-xs leading-5 text-emerald-800">Form ini sudah menulis database. Kartu/QR+PIN belum diterbitkan pada step ini.</div>
						<label class="space-y-1.5"><span class="text-xs font-medium text-muted-foreground">Nama kegiatan</span><input class="min-w-0 w-full rounded-md border border-input bg-background px-3 py-2 text-sm outline-none focus:border-primary" placeholder="Contoh: UAS Genap" bind:value={draft.nama} /></label>
						<div class="grid gap-3 sm:grid-cols-2"><label class="space-y-1.5"><span class="text-xs font-medium text-muted-foreground">Jenis kegiatan</span><select class="min-w-0 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={draft.jenis}><option>Ujian Semester</option><option>Gladi CBT</option><option>Tryout</option><option>Simulasi</option></select></label><label class="space-y-1.5"><span class="text-xs font-medium text-muted-foreground">Mode pelaksanaan</span><select class="min-w-0 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={draft.mode}><option>CBT Web</option><option>Android</option><option>Web / Android</option><option>Kertas / Campuran</option></select></label></div>
						<div class="grid gap-3 sm:grid-cols-2"><label class="space-y-1.5"><span class="text-xs font-medium text-muted-foreground">Tahun ajaran</span><input class="min-w-0 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={draft.tahunAjaran} /></label><label class="space-y-1.5"><span class="text-xs font-medium text-muted-foreground">Semester</span><select class="min-w-0 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={draft.semester}><option>Ganjil</option><option>Genap</option></select></label></div>
						<div class="grid gap-3 sm:grid-cols-2"><label class="space-y-1.5"><span class="text-xs font-medium text-muted-foreground">Tanggal mulai</span><input type="date" class="min-w-0 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={draft.tanggalMulai} /></label><label class="space-y-1.5"><span class="text-xs font-medium text-muted-foreground">Tanggal selesai</span><input type="date" class="min-w-0 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={draft.tanggalSelesai} /></label></div>
						<label class="space-y-1.5"><span class="text-xs font-medium text-muted-foreground">Catatan singkat</span><textarea class="min-h-24 w-full rounded-md border border-input bg-background px-3 py-2 text-sm outline-none focus:border-primary" placeholder="Opsional untuk operator; belum disimpan sebagai kolom khusus." bind:value={draft.catatan}></textarea></label>
						{#if formError}<p class="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive" role="alert">{formError}</p>{/if}
					</div>
					<div class="flex flex-col gap-2 border-t border-border bg-card px-5 py-4 sm:flex-row sm:justify-end"><button type="button" class="rounded-md border px-4 py-2 text-sm font-semibold text-foreground hover:bg-muted" onclick={resetDraft} disabled={saving}>Reset</button><button type="submit" class="rounded-md bg-primary px-4 py-2 text-sm font-semibold text-primary-foreground hover:bg-primary/90 disabled:opacity-60" disabled={saving}>{saving ? 'Menyimpan…' : 'Simpan Kegiatan'}</button></div>
				</form>
			</aside>
		</div>
	{/if}

	{#if selectedKegiatan}
		<div class="fixed inset-0 z-50 flex justify-end" role="dialog" aria-modal="true" aria-labelledby="detail-drawer-title">
			<button type="button" class="absolute inset-0 bg-slate-950/35 backdrop-blur-[1px]" aria-label="Tutup detail kegiatan" onclick={closeKegiatanDetail}></button>
			<aside class="relative flex h-full w-full max-w-2xl flex-col border-l border-border bg-card shadow-2xl sm:w-[42rem]">
				<div class="border-b border-border px-5 py-4">
					<div class="flex items-start justify-between gap-3">
						<div class="min-w-0 space-y-2"><p class="text-xs font-semibold tracking-[0.18em] text-muted-foreground uppercase">Detail Kegiatan</p><h2 id="detail-drawer-title" class="truncate text-lg font-bold text-foreground">{selectedKegiatan.nama}</h2><div class="flex flex-wrap items-center gap-2"><span class={`rounded-full border px-2.5 py-1 text-xs font-semibold ${statusTone[selectedKegiatan.status]}`}>{selectedKegiatan.status}</span><span class="rounded-full border bg-muted px-2.5 py-1 text-xs font-medium text-muted-foreground">{selectedKegiatan.mode}</span></div></div>
						<button type="button" class="inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-full border text-sm font-bold text-muted-foreground hover:bg-muted" aria-label="Tutup detail" onclick={closeKegiatanDetail}>×</button>
					</div>
				</div>
				<div class="min-h-0 flex-1 space-y-4 overflow-y-auto px-5 py-4">
					<section class="rounded-xl border bg-background p-4"><h3 class="text-sm font-semibold text-foreground">Ringkasan</h3><div class="mt-3 grid gap-2 text-sm"><div class="flex justify-between gap-3"><span class="text-muted-foreground">Tanggal</span><strong class="text-right font-semibold">{selectedKegiatan.periode}</strong></div><div class="flex justify-between gap-3"><span class="text-muted-foreground">Peserta</span><strong>{selectedKegiatan.peserta}</strong></div><div class="flex justify-between gap-3"><span class="text-muted-foreground">Ruang</span><strong>{selectedKegiatan.ruang}</strong></div><div class="flex justify-between gap-3"><span class="text-muted-foreground">Sesi</span><strong>{selectedKegiatan.sesi}</strong></div></div><p class="mt-3 rounded-lg bg-muted/50 px-3 py-2 text-xs leading-5 text-muted-foreground">{selectedKegiatan.catatan}</p></section>

					<section class="rounded-xl border bg-background p-4">
						<h3 class="text-sm font-semibold text-foreground">Menu Dalam Kegiatan</h3>
						<div class="mt-3 grid gap-2 sm:grid-cols-3">{#each detailFeatures as feature}<button type="button" class={`rounded-lg border px-3 py-2 text-left text-sm transition ${activeDetailFeature === feature.key ? 'border-primary bg-primary/10 text-primary' : 'bg-card text-foreground hover:bg-muted'}`} onclick={() => (activeDetailFeature = feature.key)}><span class="block font-semibold">{feature.label}</span><span class="mt-1 block text-[11px] leading-4 text-muted-foreground">{feature.description}</span></button>{/each}</div>
						{#if activeFeature && activeFeature.key !== 'ruang' && activeFeature.key !== 'peserta'}<div class="mt-3 rounded-lg border border-dashed bg-muted/30 px-3 py-3"><p class="text-sm font-semibold text-foreground">{activeFeature.label}</p><p class="mt-1 text-xs leading-5 text-muted-foreground">{activeFeature.description} Belum dibuka pada step ini agar migrasi tetap kecil dan aman.</p></div>{/if}
					</section>

					<section class="rounded-xl border border-emerald-200 bg-emerald-50/60 p-4">
						<div class="flex flex-col gap-2 sm:flex-row sm:items-start sm:justify-between">
							<div><p class="text-xs font-semibold tracking-[0.16em] text-emerald-700 uppercase">Step 6 · Ruang & Peserta</p><h3 class="text-base font-bold text-foreground">Wizard penempatan ruang</h3><p class="mt-1 text-xs leading-5 text-muted-foreground">Alur baru: pilih rombel → atur pola acak → review peta ruang visual → edit manual bila perlu. Kartu/QR+PIN belum diterbitkan.</p></div>
							<button type="button" class="rounded-md border bg-background px-3 py-2 text-xs font-semibold text-foreground hover:bg-muted" onclick={loadRombelOptions} disabled={loadingRombel}>{loadingRombel ? 'Memuat…' : 'Refresh Rombel'}</button>
						</div>

						<div class="mt-4 grid gap-2 sm:grid-cols-4"><div class="rounded-lg border border-emerald-300 bg-background px-3 py-2"><p class="text-xs font-bold text-emerald-700">① Pilih Rombel</p><p class="text-[11px] text-muted-foreground">Tentukan peserta</p></div><div class="rounded-lg border border-emerald-300 bg-background px-3 py-2"><p class="text-xs font-bold text-emerald-700">② Atur Acak</p><p class="text-[11px] text-muted-foreground">Ruang & pola</p></div><div class="rounded-lg border border-emerald-300 bg-background px-3 py-2"><p class="text-xs font-bold text-emerald-700">③ Review Ruang</p><p class="text-[11px] text-muted-foreground">Cek kapasitas</p></div><div class="rounded-lg border border-emerald-300 bg-background px-3 py-2"><p class="text-xs font-bold text-emerald-700">④ Manual</p><p class="text-[11px] text-muted-foreground">Pindah peserta</p></div></div>

						<div class="mt-4 rounded-lg border bg-background p-3">
							<div class="flex flex-wrap items-center justify-between gap-2"><h4 class="text-sm font-semibold text-foreground">① Pilih Rombel Peserta</h4><p class="text-xs text-muted-foreground">Terpilih: {selectedClassIds.length} rombel · ±{selectedStudentTotal} siswa</p></div>
							{#if loadingRombel}<p class="mt-3 text-sm text-muted-foreground">Memuat rombel…</p>{:else if rombelOptions.length === 0}<p class="mt-3 text-sm text-muted-foreground">Belum ada rombel aktif yang bisa dipilih.</p>{:else}<div class="mt-3 grid max-h-56 gap-2 overflow-y-auto pr-1 sm:grid-cols-2">{#each rombelOptions as rombel}<label class="flex cursor-pointer items-start gap-2 rounded-lg border px-3 py-2 text-sm hover:bg-muted/60"><input type="checkbox" class="mt-1" checked={selectedClassIds.includes(rombel.id)} onchange={() => toggleClass(rombel.id)} /><span class="min-w-0"><span class="block font-semibold text-foreground">{rombel.code || rombel.name}</span><span class="block text-xs text-muted-foreground">{rombel.name} · {rombel.total_students ?? 0} siswa</span></span></label>{/each}</div>{/if}
						</div>

						<div class="mt-4 rounded-lg border bg-background p-3"><h4 class="text-sm font-semibold text-foreground">② Atur Acak</h4><div class="mt-3 grid gap-3 sm:grid-cols-3"><label class="space-y-1"><span class="text-xs font-medium text-muted-foreground">Jumlah ruang</span><input type="number" min="1" max="20" class="w-full rounded-md border bg-background px-3 py-2 text-sm" bind:value={roomCount} /></label><label class="space-y-1"><span class="text-xs font-medium text-muted-foreground">Kapasitas/ruang</span><input type="number" min="1" max="50" class="w-full rounded-md border bg-background px-3 py-2 text-sm" bind:value={capacityPerRoom} /></label><label class="space-y-1"><span class="text-xs font-medium text-muted-foreground">Pola acak</span><select class="w-full rounded-md border bg-background px-3 py-2 text-sm" bind:value={mixPolicy}><option value="mixed">Campur rombel</option><option value="class_grouped">Kelompok kelas</option></select></label></div></div>

						{#if assignmentError}<p class="mt-3 rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive" role="alert">{assignmentError}</p>{/if}
						{#if assignmentNotice}<p class="mt-3 rounded-md border border-emerald-300 bg-emerald-100 px-3 py-2 text-sm font-medium text-emerald-800" role="status">{assignmentNotice}</p>{/if}

						<div class="mt-4 flex flex-col gap-2 sm:flex-row sm:justify-end"><button type="button" class="rounded-md border bg-background px-4 py-2 text-sm font-semibold text-foreground hover:bg-muted disabled:opacity-60" onclick={() => void previewAssignment()} disabled={workingAssignment}>{workingAssignment ? 'Memproses…' : 'Preview Pembagian Ruang'}</button><button type="button" class="rounded-md bg-primary px-4 py-2 text-sm font-semibold text-primary-foreground hover:bg-primary/90 disabled:opacity-60" onclick={() => void applyAssignment()} disabled={workingAssignment}>{workingAssignment ? 'Menyimpan…' : 'Simpan Penempatan'}</button></div>

						<div class="mt-4 rounded-lg border bg-background p-3"><div class="flex flex-wrap items-center justify-between gap-2"><h4 class="text-sm font-semibold text-foreground">③ Review Ruang · Peta Ruang Visual</h4><p class="text-xs text-muted-foreground">Klik preview untuk melihat isi ruang sebelum simpan.</p></div>{#if assignmentPreview}<div class="mt-3 grid gap-2 sm:grid-cols-2">{#each assignmentPreview.rooms as room}<div class="rounded-xl border bg-card p-3 text-sm"><div class="flex items-center justify-between gap-2"><strong>{room.code}</strong><span class="rounded-full bg-muted px-2 py-0.5 text-[11px] font-semibold">{room.assigned_count}/{room.capacity}</span></div><div class="mt-2 h-2 overflow-hidden rounded-full bg-muted"><div class="h-full rounded-full bg-emerald-500" style={`width: ${Math.min(100, Math.round((room.assigned_count / Math.max(1, room.capacity)) * 100))}%`}></div></div><p class="mt-2 text-xs text-muted-foreground">{room.name}</p></div>{/each}</div><div class="mt-3 grid gap-2 text-sm sm:grid-cols-4"><div><p class="text-xs text-muted-foreground">Peserta</p><p class="text-xl font-bold">{assignmentPreview.total_participants}</p></div><div><p class="text-xs text-muted-foreground">Tertampung</p><p class="text-xl font-bold">{assignmentPreview.assigned_total}</p></div><div><p class="text-xs text-muted-foreground">Sisa</p><p class="text-xl font-bold">{assignmentPreview.unassigned_total}</p></div><div><p class="text-xs text-muted-foreground">Status</p><p class="text-sm font-semibold">{assignmentPreview.applied ? 'Tersimpan' : 'Preview'}</p></div></div><p class="mt-2 text-xs leading-5 text-muted-foreground">{assignmentPreview.message}</p>{:else}<p class="mt-3 rounded-lg border border-dashed bg-muted/30 px-3 py-3 text-sm text-muted-foreground">Belum ada preview. Pilih rombel dan klik Preview Pembagian Ruang.</p>{/if}</div>
					</section>


					<section class="rounded-xl border border-sky-200 bg-sky-50/50 p-4">
						<div class="flex flex-col gap-2 sm:flex-row sm:items-start sm:justify-between">
							<div><p class="text-xs font-semibold tracking-[0.16em] text-sky-700 uppercase">Mode Manual · Acak Sendiri</p><h3 class="text-base font-bold text-foreground">Atur peserta per ruang dan nomor kursi</h3><p class="mt-1 text-xs leading-5 text-muted-foreground">Ambil daftar peserta setelah simpan ruang. Tombol Acak Tampilan hanya mengubah urutan tampil agar Bapak bisa memilih manual; penyimpanan tetap per peserta lewat tombol Pindah.</p></div>
							<div class="flex flex-wrap gap-2"><button type="button" class="rounded-md border bg-background px-3 py-2 text-xs font-semibold text-foreground hover:bg-muted" onclick={() => void loadParticipantPlacements()} disabled={loadingPlacements}>{loadingPlacements ? 'Memuat…' : 'Ambil Peserta'}</button><button type="button" class="rounded-md border bg-background px-3 py-2 text-xs font-semibold text-foreground hover:bg-muted" onclick={shuffleManualView} disabled={participantPlacements.length === 0}>Acak Tampilan</button></div>
						</div>
						{#if placementError}<p class="mt-3 rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive" role="alert">{placementError}</p>{/if}
						{#if participantPlacements.length === 0}<p class="mt-3 rounded-lg border border-dashed bg-background px-3 py-3 text-sm text-muted-foreground">Belum ada peserta dimuat. Simpan ruang & peserta dulu, lalu klik Ambil Peserta untuk edit manual.</p>{:else}
							<div class="mt-3 max-h-80 space-y-2 overflow-y-auto pr-1">{#each participantPlacements as placement (placement.participant_id)}<div class="grid gap-2 rounded-lg border bg-background p-3 text-sm sm:grid-cols-[1fr_8rem_6rem_5rem]"><div class="min-w-0"><p class="truncate font-semibold text-foreground">{placement.student_name}</p><p class="text-xs text-muted-foreground">{placement.class_code} · {placement.room_code || 'Belum ruang'} · Kursi {placement.seat_no || '-'}</p></div><select class="rounded-md border bg-card px-2 py-2 text-xs" value={placement.room_id || ''} onchange={(event) => (placement.room_id = event.currentTarget.value)}>{#each manualRoomOptions as room}<option value={room.room_id}>{room.room_code} · {room.room_name}</option>{/each}</select><input class="rounded-md border bg-card px-2 py-2 text-xs" type="number" min="1" max={placement.room_capacity || capacityPerRoom} value={placement.seat_no || 1} oninput={(event) => (placement.seat_no = Number(event.currentTarget.value))} /><button type="button" class="rounded-md bg-primary px-3 py-2 text-xs font-semibold text-primary-foreground disabled:opacity-60" onclick={() => void moveParticipantSeat(placement, placement.room_id || '', Number(placement.seat_no || 1))} disabled={workingParticipantId === placement.participant_id}>{workingParticipantId === placement.participant_id ? '...' : 'Pindah'}</button></div>{/each}</div>
						{/if}
					</section>

					<section class="rounded-xl border bg-background p-4"><h3 class="text-sm font-semibold text-foreground">Checklist Persiapan</h3><div class="mt-3 space-y-2">{#each preparationChecklist as label, index}<div class="flex items-center gap-3 rounded-lg border bg-card px-3 py-2 text-sm"><span class="inline-flex h-6 w-6 shrink-0 items-center justify-center rounded-full border text-[11px] font-semibold text-muted-foreground">{index + 1}</span><span class="min-w-0 flex-1 text-foreground">{label}</span><span class="rounded-full bg-muted px-2 py-0.5 text-[11px] font-medium text-muted-foreground">Bertahap</span></div>{/each}</div></section>
				</div>
				<div class="border-t border-border bg-card px-5 py-4"><button type="button" class="w-full rounded-md border px-4 py-2 text-sm font-semibold text-foreground hover:bg-muted" onclick={closeKegiatanDetail}>Tutup Detail</button></div>
			</aside>
		</div>
	{/if}

	{#if formNotice}<p class="rounded-xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm font-medium text-emerald-700" role="status">{formNotice}</p>{/if}

	<section class="grid gap-3 md:grid-cols-4"><div class="rounded-xl border bg-background p-4"><p class="text-xs font-medium text-muted-foreground">Kegiatan aktif</p><p class="mt-1 text-2xl font-bold">{kegiatan.length}</p></div><div class="rounded-xl border bg-background p-4"><p class="text-xs font-medium text-muted-foreground">Peserta terhubung</p><p class="mt-1 text-2xl font-bold">{totalPeserta}</p></div><div class="rounded-xl border bg-background p-4"><p class="text-xs font-medium text-muted-foreground">Ruang disiapkan</p><p class="mt-1 text-2xl font-bold">{totalRuang}</p></div><div class="rounded-xl border bg-background p-4"><p class="text-xs font-medium text-muted-foreground">Sesi dibuat</p><p class="mt-1 text-2xl font-bold">{totalSesi}</p></div></section>

	<section class="rounded-2xl border border-border bg-card shadow-sm"><div class="border-b border-border px-4 py-3"><h2 class="text-base font-semibold text-foreground">Daftar Kegiatan</h2><p class="text-xs text-muted-foreground">Data dibaca dari API Asesmen native.</p></div>{#if loading}<p class="px-4 py-8 text-center text-sm text-muted-foreground">Memuat kegiatan…</p>{:else if kegiatan.length === 0}<div class="px-4 py-8 text-center"><p class="text-sm font-semibold text-foreground">Belum ada kegiatan.</p><p class="mt-1 text-xs text-muted-foreground">Klik Buat Kegiatan untuk membuat draft pertama.</p></div>{:else}<div class="divide-y divide-border">{#each kegiatan as item (item.id)}<article class="p-4 transition-colors hover:bg-muted/30"><div class="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between"><div class="min-w-0 space-y-2"><div class="flex flex-wrap items-center gap-2"><span class={`rounded-full border px-2.5 py-1 text-xs font-semibold ${statusTone[item.status]}`}>{item.status}</span><span class="rounded-full border bg-muted px-2.5 py-1 text-xs font-medium text-muted-foreground">{item.jenis}</span><span class="rounded-full border bg-muted px-2.5 py-1 text-xs font-medium text-muted-foreground">{item.mode}</span></div><h3 class="truncate text-lg font-bold text-foreground">{item.nama}</h3><p class="text-sm text-muted-foreground">Tanggal: {item.periode}</p><p class="max-w-2xl text-xs leading-5 text-muted-foreground">{item.catatan}</p></div><div class="space-y-2 sm:min-w-[18rem]"><div class="grid min-w-full grid-cols-3 gap-2 text-center"><div class="rounded-lg border bg-background p-2"><p class="text-[11px] text-muted-foreground">Peserta</p><p class="text-lg font-bold">{item.peserta}</p></div><div class="rounded-lg border bg-background p-2"><p class="text-[11px] text-muted-foreground">Ruang</p><p class="text-lg font-bold">{item.ruang}</p></div><div class="rounded-lg border bg-background p-2"><p class="text-[11px] text-muted-foreground">Sesi</p><p class="text-lg font-bold">{item.sesi}</p></div></div><button type="button" class="w-full rounded-md border bg-background px-4 py-2 text-sm font-semibold text-foreground hover:bg-muted" onclick={() => openKegiatanDetail(item.id)}>Kelola</button></div></div></article>{/each}</div>{/if}</section>

	<section class="rounded-2xl border border-dashed border-border bg-muted/30 p-4"><h2 class="text-sm font-semibold text-foreground">Batas step 6</h2><ul class="mt-2 list-disc space-y-1 pl-5 text-sm leading-6 text-muted-foreground"><li>Tidak membuat tabel `kegiatan` baru; memakai tabel native `assessment_exams` yang sudah ada.</li><li>Ruang dan peserta disimpan melalui API assignment preview/apply yang sudah ada.</li><li>QR+PIN, cetak kartu, paket soal, sesi lanjutan, dan hasil tetap disambungkan bertahap agar aman.</li></ul></section>
</div>
