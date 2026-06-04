<script lang="ts">
	import { onMount, tick } from 'svelte';
	import { page } from '$app/state';
	import { clientApiPath, readClientApiData } from '$lib/client/api';
	import { summarizeDocumentPrintStatus } from '$lib/asesmen/document-print-readiness';

	type KegiatanStatus = 'Draft' | 'Siap' | 'Berlangsung' | 'Selesai' | 'Arsip';
	type DetailFeatureKey = 'paket' | 'ruang' | 'sesi' | 'cetak' | 'hasil' | 'manual';

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

	type AssignmentClassSummary = {
		class_code: string;
		class_name: string;
		grade_level: number;
		count: number;
	};

	type AssignmentRoom = {
		code: string;
		name: string;
		capacity: number;
		assigned_count: number;
		grade_levels?: number[];
		class_summary?: AssignmentClassSummary[];
	};

	type AssignmentMixPolicy = 'mixed' | 'class_grouped';
	type AssignmentUiMode = 'balanced_all' | 'mixed_rombel' | 'class_grouped' | 'ordered_participant' | 'csv_manual';

	type AssignmentResult = {
		exam_id: string;
		session_id?: string;
		mix_policy: AssignmentMixPolicy;
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
		session_id?: string;
		room_id?: string;
		student_id?: string;
		student_name: string;
		nis?: string;
		nisn?: string;
		class_code: string;
		class_name: string;
		room_code?: string;
		room_name?: string;
		room_capacity?: number;
		seat_no?: number;
	};

	type AssessmentPackageOption = {
		id: string;
		event_id?: string;
		subject_id: string;
		subject_code?: string;
		subject_name: string;
		title: string;
		description?: string;
		duration_minutes: number;
		question_count: number;
		session_count: number;
		locked?: boolean;
		snapshot_version?: number;
	};

	type AssessmentPackageMap = {
		id: string;
		exam_id?: string;
		class_id: string;
		class_code?: string;
		class_name?: string;
		grade_level?: number;
		subject_id: string;
		subject_code?: string;
		subject_name: string;
		package_id: string;
		package_title?: string;
		duration_minutes?: number;
		slot_label?: string;
		notes?: string;
	};
	type EditablePackageMap = AssessmentPackageMap & { local_id: string; is_new?: boolean };
	type PackageClassGroup = { class_id: string; local_id: string; rows: EditablePackageMap[] };

	type AssessmentSessionStatus = 'draft' | 'scheduled' | 'active' | 'finished' | 'archived';
	type AssessmentSessionReadiness = {
		id: string;
		event_id?: string;
		package_id: string;
		package_title: string;
		subject_id?: string;
		subject_name?: string;
		subject_code?: string;
		class_id?: string;
		class_name?: string;
		class_code?: string;
		title: string;
		scheduled_start?: string;
		scheduled_end?: string;
		status: AssessmentSessionStatus;
		question_count?: number;
		published_question_count?: number;
		participant_count?: number;
		assigned_participant_count?: number;
		missing_seat_count?: number;
		token_ready_count?: number;
		room_count?: number;
		total_capacity?: number;
		rooms_without_proctor?: number;
	};

	type SessionDraft = {
		packageMapLocalId: string;
		date: string;
		startTime: string;
		durationMinutes: number;
		title: string;
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
	let rombelOptions = $state<AcademicRombel[]>([]);
	let showCreateForm = $state(false);
	let selectedKegiatanId = $state<string | null>(null);
	let activeDetailFeature = $state<DetailFeatureKey | null>(null);
	let draft = $state<DraftKegiatan>({ ...emptyDraft });
	let selectedClassIds = $state<string[]>([]);
	let roomCount = $state(8);
	let capacityPerRoom = $state(30);
	let assignmentMode = $state<AssignmentUiMode>('balanced_all');
	let balanceRooms = $state(true);
	let spreadRombel = $state(true);
	let assignmentPreview = $state<AssignmentResult | null>(null);
	let participantPlacements = $state<ParticipantPlacement[]>([]);
	let packageOptions = $state<AssessmentPackageOption[]>([]);
	let packageMaps = $state<EditablePackageMap[]>([]);
	let workingParticipantId = $state('');
	let formError = $state('');
	let formNotice = $state('');
	let listError = $state('');
	let assignmentError = $state('');
	let assignmentNotice = $state('');
	let placementError = $state('');
	let packageError = $state('');
	let packageNotice = $state('');
	let csvImportNotice = $state('');
	let documentNotice = $state('');
	let documentError = $state('');
	let printableCards = $state<ParticipantCard[]>([]);
	let qrImages = $state<Record<string, string>>({});
	let printMode = $state<'cards' | 'supervisor' | 'checklist' | null>(null);
	let issuingCards = $state(false);
	let checkingCards = $state(false);
	let loading = $state(true);
	let saving = $state(false);
	let loadingRombel = $state(false);
	let workingAssignment = $state(false);
	let loadingPlacements = $state(false);
	let loadingPackages = $state(false);
	let savingPackages = $state(false);
	let sessionRows = $state<AssessmentSessionReadiness[]>([]);
	let sessionDraft = $state<SessionDraft>({ packageMapLocalId: '', date: '', startTime: '07:30', durationMinutes: 90, title: '' });
	let loadingSessions = $state(false);
	let savingSession = $state(false);
	let workingSessionId = $state('');
	let sessionError = $state('');
	let sessionNotice = $state('');

	const totalPeserta = $derived(kegiatan.reduce((total, item) => total + item.peserta, 0));
	const totalRuang = $derived(kegiatan.reduce((total, item) => total + item.ruang, 0));
	const totalSesi = $derived(kegiatan.reduce((total, item) => total + item.sesi, 0));
	const selectedKegiatan = $derived(kegiatan.find((item) => item.id === selectedKegiatanId) ?? null);
	const selectedRombel = $derived(rombelOptions.filter((item) => selectedClassIds.includes(item.id)));
	const selectedStudentTotal = $derived(selectedRombel.reduce((total, item) => total + Number(item.total_students ?? 0), 0));
	const manualRoomOptions = $derived(Array.from(new Map(participantPlacements.filter((item) => item.room_id).map((item) => [item.room_id, item])).values()));
	const documentPrintSummary = $derived(summarizeDocumentPrintStatus({ participantCount: selectedKegiatan?.peserta ?? 0, roomCount: selectedKegiatan?.ruang ?? 0, cardCount: selectedKegiatan?.kartu ?? 0 }));
	const packageReadyCount = $derived(packageMaps.filter((item) => item.class_id && item.package_id).length);
	const packageSubjectCount = $derived(new Set(packageMaps.filter((item) => item.subject_id).map((item) => item.subject_id)).size);
	const packageClassCount = $derived(new Set(packageMaps.filter((item) => item.class_id && item.package_id).map((item) => item.class_id)).size);
	const packageSelectedCount = $derived(new Set(packageMaps.filter((item) => item.package_id).map((item) => item.package_id)).size);
	const packageMaxDuration = $derived(packageMaps.reduce((max, item) => Math.max(max, Number(item.duration_minutes || selectedPackageOption(item.package_id)?.duration_minutes || 0)), 0));
	const packageClassGroups = $derived(buildPackageClassGroups());
	const packageGateReady = $derived(packageReadyCount > 0);
	const packageGateMessage = 'Pilih minimal satu paket siap dari Modul Paket Soal sebelum lanjut ke ruang, sesi, cetak kartu, atau pelaksanaan.';
	const sessionPackageMapOptions = $derived(packageMaps.filter((item) => item.class_id && item.package_id));
	const selectedSessionPackageMap = $derived(sessionPackageMapOptions.find((item) => item.local_id === sessionDraft.packageMapLocalId) ?? sessionPackageMapOptions[0] ?? null);
	const selectedSessionPackageReused = $derived(Boolean(selectedSessionPackageMap && packageReusedFromAnotherKegiatan(selectedSessionPackageMap.package_id)));
	const sessionGateMessage = 'Sesi wajib terikat ke Kegiatan, memakai paket siap, punya peserta, ruang/kursi lengkap, dan jadwal valid sebelum diaktifkan.';
	const routeKegiatanId = $derived(page.params.id ?? '');

	type StepState = { label: string; tone: string; helper: string };

	function readinessScore(item: KegiatanUjian) {
		let score = 0;
		if (packageReadyCount > 0) score += 25;
		if (item.peserta > 0) score += 20;
		if (item.ruang > 0) score += 20;
		if (item.sesi > 0) score += 15;
		if (item.kartu > 0) score += 20;
		return Math.min(100, score);
	}

	function readinessTone(score: number) {
		if (score >= 80) return 'bg-emerald-500';
		if (score >= 45) return 'bg-amber-500';
		return 'bg-slate-400';
	}

	function readinessChecks(item: KegiatanUjian) {
		return [
			{ label: 'Paket soal dipilih', ready: packageReadyCount > 0 },
			{ label: 'Peserta masuk', ready: item.peserta > 0 },
			{ label: 'Ruang tersusun', ready: item.ruang > 0 },
			{ label: 'Jadwal sesi dibuat', ready: item.sesi > 0 },
			{ label: 'QR+PIN/kartu terbit', ready: item.kartu > 0 }
		];
	}

	function nextDetailActionKey(item: KegiatanUjian): DetailFeatureKey {
		if (packageReadyCount <= 0) return 'paket';
		if (item.peserta <= 0 || item.ruang <= 0) return 'ruang';
		if (item.sesi <= 0) return 'sesi';
		if (item.kartu <= 0) return 'cetak';
		return item.status === 'Selesai' || item.status === 'Arsip' ? 'hasil' : 'cetak';
	}

	function nextDetailActionLabel(item: KegiatanUjian) {
		const key = nextDetailActionKey(item);
		return detailFeatures.find((feature) => feature.key === key)?.label ?? 'Review Kegiatan';
	}

	function stepState(key: DetailFeatureKey, item: KegiatanUjian): StepState {
		if (key === 'paket') return packageReadyCount > 0
			? { label: 'Selesai', tone: 'border-emerald-200 bg-emerald-50 text-emerald-700', helper: `${packageReadyCount} paket dipilih` }
			: { label: 'Berikutnya', tone: 'border-amber-200 bg-amber-50 text-amber-700', helper: 'Pilih paket siap dulu' };
		if (!packageGateReady && key !== 'manual') return { label: 'Terkunci', tone: 'border-slate-200 bg-slate-50 text-slate-500', helper: 'Pilih paket siap dulu' };
		if (key === 'ruang') return item.peserta > 0 && item.ruang > 0
			? { label: 'Selesai', tone: 'border-emerald-200 bg-emerald-50 text-emerald-700', helper: `${item.peserta} peserta · ${item.ruang} ruang` }
			: { label: nextDetailActionKey(item) === 'ruang' ? 'Berikutnya' : 'Belum lengkap', tone: 'border-amber-200 bg-amber-50 text-amber-700', helper: 'Pilih rombel dan susun kursi' };
		if (key === 'sesi') return item.sesi > 0
			? { label: 'Selesai', tone: 'border-emerald-200 bg-emerald-50 text-emerald-700', helper: `${item.sesi} sesi` }
			: { label: nextDetailActionKey(item) === 'sesi' ? 'Berikutnya' : 'Belum lengkap', tone: 'border-amber-200 bg-amber-50 text-amber-700', helper: 'Jadwal/token belum final' };
		if (key === 'cetak') return item.kartu > 0
			? { label: 'Selesai', tone: 'border-emerald-200 bg-emerald-50 text-emerald-700', helper: `${item.kartu} kartu terbit` }
			: { label: nextDetailActionKey(item) === 'cetak' ? 'Berikutnya' : 'Belum lengkap', tone: 'border-amber-200 bg-amber-50 text-amber-700', helper: 'QR+PIN belum terbit' };
		if (key === 'hasil') return item.status === 'Selesai' || item.status === 'Arsip'
			? { label: 'Siap dibuka', tone: 'border-emerald-200 bg-emerald-50 text-emerald-700', helper: 'Rekap dan arsip' }
			: { label: 'Belum mulai', tone: 'border-slate-200 bg-slate-50 text-slate-500', helper: 'Dibuka setelah pelaksanaan' };
		return { label: 'Admin', tone: 'border-sky-200 bg-sky-50 text-sky-700', helper: 'CSV dan edit teknis' };
	}

	const assignmentModeDescriptions: Record<AssignmentUiMode, string> = {
		balanced_all: 'Rekomendasi default: peserta disebar seimbang ke semua ruang dan rombel diusahakan tidak berkumpul.',
		mixed_rombel: 'Sistem mengacak peserta dengan target komposisi rombel bercampur di setiap ruang.',
		class_grouped: 'Peserta dari rombel yang sama ditempatkan berdekatan/seruang selama kapasitas cukup.',
		ordered_participant: 'Peserta ditempatkan mengikuti urutan data peserta untuk administrasi yang mudah dicari.',
		csv_manual: 'Manual dari CSV: Download Template CSV, isi ruang dan urutan, lalu Import CSV untuk divalidasi sebelum disimpan.'
	};

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
		{ key: 'paket', label: 'Paket Siap', description: 'Pilih dari Modul Paket Soal.' },
		{ key: 'ruang', label: 'Peserta & Ruang', description: 'Pilih rombel, ruang, dan kursi.' },
		{ key: 'sesi', label: 'Sesi', description: 'Rancang jadwal dan alur masuk ujian.' },
		{ key: 'cetak', label: 'Cetak', description: 'Kartu peserta dan lembar pengawas.' },
		{ key: 'hasil', label: 'Hasil', description: 'Pantau nilai, berita acara, dan arsip.' },
		{ key: 'manual', label: 'Mode Lengkap', description: 'CSV dan edit panitia lanjutan.' }
	];


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

	onMount(async () => {
		await loadKegiatan();
		await loadRombelOptions();
		if (routeKegiatanId) openKegiatanDetail(routeKegiatanId);
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
		csvImportNotice = '';
		documentNotice = '';
		documentError = '';
		packageError = '';
		packageNotice = '';
		sessionError = '';
		sessionNotice = '';
		sessionRows = [];
	}

	function toggleCreateForm() {
		showCreateForm = !showCreateForm;
		if (showCreateForm) selectedKegiatanId = null;
		formNotice = '';
		if (showCreateForm) formError = '';
	}

	function openKegiatanDetail(id: string) {
		selectedKegiatanId = id;
		activeDetailFeature = 'paket';
		showCreateForm = false;
		formNotice = '';
		resetAssignmentState();
		if (rombelOptions.length === 0) void loadRombelOptions();
		void loadPackageOptions();
		void loadPackageMaps(id);
		void loadSessions(id);
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

	function gradeLabel(level: number) {
		return level > 0 ? `Kelas ${level}` : 'Tanpa tingkat';
	}

	function classSummaryLabel(item: AssignmentClassSummary) {
		const code = item.class_code || item.class_name || 'Tanpa Rombel';
		return `${code} ${item.count}`;
	}

	function backendMixPolicy(): AssignmentMixPolicy {
		if (!spreadRombel || assignmentMode === 'class_grouped' || assignmentMode === 'ordered_participant') return 'class_grouped';
		return 'mixed';
	}

	function assignmentPayload() {
		return {
			room_count: Number(roomCount),
			capacity_per_room: Number(capacityPerRoom),
			mix_policy: backendMixPolicy(),
			class_ids: selectedClassIds
		};
	}

	function csvCell(value: unknown) {
		const text = String(value ?? '');
		return /[",\n]/.test(text) ? `"${text.replaceAll('"', '""')}"` : text;
	}

	function downloadCsv(filename: string, content: string) {
		const blob = new Blob([content], { type: 'text/csv;charset=utf-8' });
		const url = URL.createObjectURL(blob);
		const link = document.createElement('a');
		link.href = url;
		link.download = filename;
		document.body.appendChild(link);
		link.click();
		link.remove();
		URL.revokeObjectURL(url);
	}

	function downloadPlacementTemplateCsv() {
		const header = 'student_id,nomor_peserta,nama,rombel,ruang,urutan';
		const rows = participantPlacements.length > 0
			? participantPlacements.map((item) => [
				item.student_id || item.participant_id,
				item.nis || item.nisn || '',
				item.student_name,
				item.class_code || item.class_name,
				item.room_code || 'R01',
				item.seat_no || 1
			].map(csvCell).join(','))
			: ['contoh-student-id,250001,Nama Siswa,9A,R01,1'];
		downloadCsv(`template-penempatan-${selectedKegiatan?.nama || 'asesmen'}.csv`, [header, ...rows].join('\n'));
		csvImportNotice = 'Template CSV siap. Kolom ruang memakai kode ruang seperti R01, R02; urutan adalah nomor kursi.';
	}

	function parseCsvLine(line: string) {
		const cells: string[] = [];
		let current = '';
		let quoted = false;
		for (let index = 0; index < line.length; index++) {
			const char = line[index];
			if (char === '"') {
				if (quoted && line[index + 1] === '"') {
					current += '"';
					index++;
				} else {
					quoted = !quoted;
				}
			} else if (char === ',' && !quoted) {
				cells.push(current.trim());
				current = '';
			} else {
				current += char;
			}
		}
		cells.push(current.trim());
		return cells;
	}

	async function handlePlacementCsvImport(event: Event) {
		const input = event.currentTarget as HTMLInputElement;
		const file = input.files?.[0];
		if (!file) return;
		placementError = '';
		csvImportNotice = '';
		try {
			if (participantPlacements.length === 0) await loadParticipantPlacements();
			if (participantPlacements.length === 0) throw new Error('Ambil/simpan peserta ruang dulu sebelum import CSV.');
			const text = await file.text();
			const lines = text.split(/\r?\n/).filter((line) => line.trim());
			if (lines.length < 2) throw new Error('CSV wajib berisi header dan minimal satu baris peserta.');
			const headers = parseCsvLine(lines[0]).map((item) => item.toLowerCase());
			const indexOf = (name: string) => headers.indexOf(name);
			const studentIndex = indexOf('student_id');
			const numberIndex = indexOf('nomor_peserta');
			const nameIndex = indexOf('nama');
			const roomIndex = indexOf('ruang');
			const seatIndex = indexOf('urutan');
			if (roomIndex < 0 || seatIndex < 0) throw new Error('CSV wajib memiliki kolom ruang dan urutan.');
			const roomByCode = new Map(manualRoomOptions.map((room) => [String(room.room_code || '').toUpperCase(), room]));
			const seen = new Set<string>();
			const moves: Array<{ participant: ParticipantPlacement; roomId: string; seatNo: number }> = [];
			for (const [lineIndex, line] of lines.slice(1).entries()) {
				const cells = parseCsvLine(line);
				const studentId = studentIndex >= 0 ? cells[studentIndex] : '';
				const number = numberIndex >= 0 ? cells[numberIndex] : '';
				const name = nameIndex >= 0 ? cells[nameIndex] : '';
				const roomCode = String(cells[roomIndex] || '').toUpperCase();
				const seatNo = Number(cells[seatIndex]);
				const participant = participantPlacements.find((item) =>
					(studentId && (item.student_id === studentId || item.participant_id === studentId)) ||
					(number && (item.nis === number || item.nisn === number)) ||
					(name && item.student_name.trim().toLowerCase() === name.trim().toLowerCase())
				);
				if (!participant) throw new Error(`Baris ${lineIndex + 2}: peserta tidak ditemukan.`);
				if (seen.has(participant.participant_id)) throw new Error(`Baris ${lineIndex + 2}: peserta dobel di CSV.`);
				seen.add(participant.participant_id);
				const room = roomByCode.get(roomCode);
				if (!room?.room_id) throw new Error(`Baris ${lineIndex + 2}: kode ruang ${roomCode || '-'} tidak valid.`);
				const capacity = room.room_capacity || capacityPerRoom;
				if (!Number.isInteger(seatNo) || seatNo < 1 || seatNo > capacity) throw new Error(`Baris ${lineIndex + 2}: urutan harus 1 sampai ${capacity}.`);
				moves.push({ participant, roomId: room.room_id, seatNo });
			}
			for (const move of moves) {
				await moveParticipantSeat(move.participant, move.roomId, move.seatNo);
			}
			await loadParticipantPlacements();
			csvImportNotice = `Import CSV valid dan ${moves.length} penempatan tersimpan. Import → Validasi → Preview → Simpan selesai.`;
		} catch (error) {
			placementError = error instanceof Error ? error.message : 'Import CSV belum dapat diproses.';
		} finally {
			input.value = '';
		}
	}


	function packageOptionLabel(option: AssessmentPackageOption) {
		const count = `${option.question_count ?? 0} soal`;
		const duration = option.duration_minutes ? ` · ${option.duration_minutes} menit` : '';
		const reuse = packageReusedFromAnotherKegiatan(option.id) ? ' · Reuse' : '';
		return `${option.subject_name} · ${option.title} (${count}${duration}${reuse})`;
	}

	function selectedPackageOption(packageId: string) {
		return packageOptions.find((option) => option.id === packageId) ?? null;
	}

	function packageEventId(packageId: string) {
		return selectedPackageOption(packageId)?.event_id || '';
	}

	function packageReusedFromAnotherKegiatan(packageId: string) {
		const eventId = packageEventId(packageId);
		return Boolean(eventId && selectedKegiatan?.id && eventId !== selectedKegiatan.id);
	}

	function packageScopeLabel(packageId: string) {
		const eventId = packageEventId(packageId);
		if (!eventId) return 'Paket umum';
		return packageReusedFromAnotherKegiatan(packageId) ? 'Reuse dari kegiatan lain' : 'Paket kegiatan ini';
	}

	function classOption(classId: string) {
		return rombelOptions.find((item) => item.id === classId) ?? null;
	}

	function classLabel(classId: string) {
		const item = classOption(classId);
		return item ? `${item.code || item.name} · ${item.total_students ?? 0} siswa` : 'Pilih rombel';
	}

	function buildPackageClassGroups(): PackageClassGroup[] {
		const byClass = new Map<string, EditablePackageMap[]>();
		for (const row of packageMaps) {
			const key = row.class_id || row.local_id;
			const rows = byClass.get(key) ?? [];
			rows.push(row);
			byClass.set(key, rows);
		}
		return Array.from(byClass.entries()).map(([key, rows]) => ({
			class_id: rows[0]?.class_id ?? '',
			local_id: key,
			rows
		}));
	}

	function packageRowsForClass(classId: string) {
		return packageMaps.filter((item) => item.class_id === classId && item.package_id);
	}

	function packageCheckedForClass(classId: string, packageId: string) {
		return packageMaps.some((item) => item.class_id === classId && item.package_id === packageId);
	}

	function packageOptionForSubjectDuplicate(classId: string, option: AssessmentPackageOption) {
		return packageMaps.some((item) => item.class_id === classId && item.subject_id === option.subject_id && item.package_id !== option.id);
	}

	function addPackageClassGroup() {
		const used = new Set(packageMaps.map((item) => item.class_id).filter(Boolean));
		const firstClass = selectedClassIds.find((id) => !used.has(id)) || rombelOptions.find((item) => !used.has(item.id))?.id || selectedClassIds[0] || rombelOptions[0]?.id || '';
		packageMaps = [
			...packageMaps,
			{
				local_id: crypto.randomUUID(),
				id: '',
				class_id: firstClass,
				subject_id: '',
				subject_name: '',
				package_id: '',
				package_title: '',
				slot_label: 'Sesi Utama',
				notes: '',
				is_new: true
			}
		];
		packageNotice = '';
	}

	function updatePackageClassGroup(localId: string, classId: string) {
		const group = packageClassGroups.find((item) => item.local_id === localId);
		if (!group) return;
		packageMaps = packageMaps.map((item) => group.rows.some((row) => row.local_id === item.local_id) ? { ...item, class_id: classId } : item);
	}

	function packageRowFromOption(classId: string, option: AssessmentPackageOption): EditablePackageMap {
		return {
			local_id: crypto.randomUUID(),
			id: '',
			class_id: classId,
			subject_id: option.subject_id,
			subject_code: option.subject_code,
			subject_name: option.subject_name,
			package_id: option.id,
			package_title: option.title,
			duration_minutes: option.duration_minutes,
			slot_label: 'Sesi Utama',
			notes: '',
			is_new: true
		};
	}

	async function togglePackageForClass(classId: string, option: AssessmentPackageOption, checked: boolean) {
		if (!classId) {
			packageError = 'Pilih rombel dulu sebelum memilih paket.';
			return;
		}
		packageError = '';
		packageNotice = '';
		if (checked) {
			if (packageOptionForSubjectDuplicate(classId, option)) {
				packageError = `Rombel ${classLabel(classId)} sudah punya paket untuk mapel ${option.subject_name}. Hapus paket mapel itu dulu jika ingin mengganti.`;
				return;
			}
			if (!packageCheckedForClass(classId, option.id)) {
				const empty = packageMaps.find((item) => item.class_id === classId && !item.package_id);
				if (empty) {
					updatePackageMapRow(empty.local_id, { package_id: option.id });
				} else {
					packageMaps = [...packageMaps, packageRowFromOption(classId, option)];
				}
			}
			return;
		}
		const row = packageMaps.find((item) => item.class_id === classId && item.package_id === option.id);
		if (!row) return;
		if (row.id) {
			await deletePackageMap(row);
		} else {
			removePackageMapRow(row.local_id);
		}
	}

	function mapPackageRows(items: AssessmentPackageMap[]): EditablePackageMap[] {
		return items.map((item) => ({ ...item, local_id: item.id || crypto.randomUUID(), is_new: false }));
	}

	async function loadPackageOptions() {
		loadingPackages = true;
		packageError = '';
		try {
			const response = await fetch('/api/asesmen/package-options');
			packageOptions = await readClientApiData<AssessmentPackageOption[]>(response);
		} catch (error) {
			packageError = error instanceof Error ? error.message : 'Daftar paket soal belum dapat dibuka.';
			packageOptions = [];
		} finally {
			loadingPackages = false;
		}
	}

	async function loadPackageMaps(examId = selectedKegiatan?.id) {
		if (!examId) return;
		loadingPackages = true;
		packageError = '';
		try {
			const response = await fetch(clientApiPath`/api/asesmen/exams/${examId}/package-maps`);
			const items = await readClientApiData<AssessmentPackageMap[]>(response);
			packageMaps = mapPackageRows(items);
		} catch (error) {
			packageError = error instanceof Error ? error.message : 'Pemetaan paket belum dapat dibuka.';
			packageMaps = [];
		} finally {
			loadingPackages = false;
		}
	}

	function addPackageMapRow() {
		addPackageClassGroup();
	}

	function updatePackageMapRow(localId: string, patch: Partial<EditablePackageMap>) {
		packageMaps = packageMaps.map((item) => {
			if (item.local_id !== localId) return item;
			const next = { ...item, ...patch };
			if (patch.package_id !== undefined) {
				const option = selectedPackageOption(patch.package_id);
				next.subject_id = option?.subject_id ?? '';
				next.subject_name = option?.subject_name ?? '';
				next.subject_code = option?.subject_code ?? '';
				next.package_title = option?.title ?? '';
				next.duration_minutes = option?.duration_minutes ?? 0;
			}
			return next;
		});
	}

	function removePackageMapRow(localId: string) {
		packageMaps = packageMaps.filter((item) => item.local_id !== localId);
	}

	async function deletePackageMap(row: EditablePackageMap) {
		if (!selectedKegiatan || !row.id) {
			removePackageMapRow(row.local_id);
			return;
		}
		const ok = window.confirm('Hapus tautan paket soal untuk rombel ini?');
		if (!ok) return;
		packageError = '';
		try {
			await fetch(clientApiPath`/api/asesmen/exams/${selectedKegiatan.id}/package-maps/${row.id}`, { method: 'DELETE' }).then((response) => readClientApiData<unknown>(response));
			packageNotice = 'Tautan paket dihapus.';
			await loadPackageMaps();
		} catch (error) {
			packageError = error instanceof Error ? error.message : 'Tautan paket belum dapat dihapus.';
		}
	}

	async function savePackageMaps() {
		if (!selectedKegiatan || savingPackages) return;
		packageError = '';
		packageNotice = '';
		const items = packageMaps
			.filter((item) => item.class_id && item.package_id)
			.map((item) => ({
				class_id: item.class_id,
				subject_id: item.subject_id,
				package_id: item.package_id,
				slot_label: item.slot_label || 'Sesi Utama',
				notes: item.notes || ''
			}));
		if (items.length === 0) {
			packageError = 'Tambahkan minimal satu rombel dan paket soal.';
			return;
		}
		if (items.some((item) => !item.class_id || !item.subject_id || !item.package_id)) {
			packageError = 'Setiap rombel yang disimpan wajib punya paket soal terpilih.';
			return;
		}
		savingPackages = true;
		try {
			const response = await fetch(clientApiPath`/api/asesmen/exams/${selectedKegiatan.id}/package-maps`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ items })
			});
			const result = await readClientApiData<{ count: number; message: string }>(response);
			packageNotice = result.message || `${result.count} tautan paket tersimpan.`;
			await loadPackageMaps();
		} catch (error) {
			packageError = error instanceof Error ? error.message : 'Paket soal belum dapat disimpan.';
		} finally {
			savingPackages = false;
		}
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
		if (!packageGateReady) {
			assignmentError = packageGateMessage;
			return;
		}
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
		if (!packageGateReady) {
			assignmentError = packageGateMessage;
			return;
		}
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

	type ParticipantCard = {
		card_id?: string;
		participant_id: string;
		session_id?: string;
		student_id?: string;
		student_name?: string;
		nis?: string;
		nisn?: string;
		class_code?: string;
		class_name?: string;
		room_code?: string;
		room_name?: string;
		seat_no?: number;
		status?: string;
		token?: string;
		pin?: string;
		qr_path?: string;
	};
	type IssueCardsResult = { count: number; cards: ParticipantCard[]; message: string };

	const printableCardsSorted = $derived([...printableCards].sort((a, b) => {
		const room = String(a.room_code || '').localeCompare(String(b.room_code || ''));
		if (room !== 0) return room;
		return Number(a.seat_no || 0) - Number(b.seat_no || 0);
	}));
	const printableCardPages = $derived(Array.from(
		{ length: Math.ceil(printableCardsSorted.length / 6) },
		(_, index) => printableCardsSorted.slice(index * 6, index * 6 + 6)
	));
	const printableRoomGroups = $derived(Array.from(printableCardsSorted.reduce((groups, card) => {
		const key = card.room_code || 'Tanpa ruang';
		if (!groups.has(key)) groups.set(key, []);
		groups.get(key)?.push(card);
		return groups;
	}, new Map<string, ParticipantCard[]>()).entries()));
	const printableHasSecrets = $derived(printableCards.some((card) => card.token || card.pin || card.qr_path));

	function cardPrintKey(card: ParticipantCard) {
		return card.card_id || card.participant_id || card.student_id || `${card.student_name}-${card.seat_no}`;
	}

	function cardQrValue(card: ParticipantCard) {
		const rawPath = card.qr_path || (card.token ? `/ujian?card=${card.token}` : '');
		if (!rawPath) return '';
		if (/^https?:\/\//.test(rawPath)) return rawPath;
		return `${window.location.origin}${rawPath.startsWith('/') ? rawPath : `/${rawPath}`}`;
	}

	async function generateQrImages(cards: ParticipantCard[]) {
		const next: Record<string, string> = {};
		const QRCode = await import('qrcode');
		for (const card of cards) {
			const value = cardQrValue(card);
			if (!value) continue;
			next[cardPrintKey(card)] = await QRCode.toDataURL(value, { errorCorrectionLevel: 'M', margin: 1, width: 180 });
		}
		qrImages = next;
	}

	async function setPrintableCards(cards: ParticipantCard[]) {
		printableCards = cards;
		await generateQrImages(cards);
	}

	function enableDocumentPrintMode() {
		if (typeof document === 'undefined') return;
		document.body.classList.add('cbt-document-print');
	}

	function disableDocumentPrintMode() {
		if (typeof document === 'undefined') return;
		document.body.classList.remove('cbt-document-print');
	}

	onMount(() => {
		const keepPrintModeActive = () => {
			if (printMode) enableDocumentPrintMode();
		};
		window.addEventListener('beforeprint', keepPrintModeActive);
		return () => {
			window.removeEventListener('beforeprint', keepPrintModeActive);
			disableDocumentPrintMode();
		};
	});

	async function printDocument(mode: 'cards' | 'supervisor' | 'checklist') {
		if ((mode === 'cards' || mode === 'supervisor') && printableCards.some((card) => cardQrValue(card)) && Object.keys(qrImages).length === 0) {
			await generateQrImages(printableCards);
		}
		printMode = mode;
		await tick();
		enableDocumentPrintMode();
		await tick();
		setTimeout(() => {
			enableDocumentPrintMode();
			window.print();
		}, 150);
	}

	async function printParticipantCards() {
		if (!selectedKegiatan) return;
		if (printableCards.length === 0) await checkParticipantCards(false);
		if (printableCards.length === 0) return;
		if (!printableHasSecrets) {
			documentError = 'PIN lama tidak bisa dibuka dari hash. Pilih PIN Baru & Cetak untuk mengganti PIN dan langsung mencetak kartu.';
			return;
		}
		await printDocument('cards');
	}

	async function regenerateParticipantCardsAndPrint() {
		await issueParticipantCards(true);
		if (printableCards.length > 0 && printableHasSecrets) await printDocument('cards');
	}

	async function printSupervisorSheets() {
		if (printableCards.length === 0) await checkParticipantCards(false);
		if (printableCards.length === 0) {
			documentError = 'Daftar peserta/kartu belum terbaca untuk lembar pengawas.';
			return;
		}
		await printDocument('supervisor');
	}

	async function checkParticipantCards(showNotice = true) {
		if (!selectedKegiatan) return;
		if (!packageGateReady) {
			documentError = packageGateMessage;
			return;
		}
		documentError = '';
		documentNotice = '';
		checkingCards = true;
		try {
			const response = await fetch(clientApiPath`/api/asesmen/exams/${selectedKegiatan.id}/cards`);
			const cards = await readClientApiData<ParticipantCard[]>(response);
			await setPrintableCards(cards);
			if (showNotice) documentNotice = `Daftar kartu terbaca: ${cards.length} peserta. Jika PIN tidak muncul, gunakan PIN Baru & Cetak.`;
		} catch (error) {
			documentError = error instanceof Error ? error.message : 'Daftar kartu peserta belum dapat dibuka.';
		} finally {
			checkingCards = false;
		}
	}


	function localDateTimeToIso(date: string, time: string) {
		if (!date || !time) return '';
		return new Date(`${date}T${time}:00+08:00`).toISOString();
	}

	function sessionEndIso(date: string, time: string, durationMinutes: number) {
		const startIso = localDateTimeToIso(date, time);
		if (!startIso) return '';
		return new Date(new Date(startIso).getTime() + Math.max(1, Number(durationMinutes || 0)) * 60_000).toISOString();
	}

	function formatDateTimeLabel(value?: string) {
		if (!value) return 'Belum diisi';
		const date = new Date(value);
		if (Number.isNaN(date.getTime())) return value;
		return new Intl.DateTimeFormat('id-ID', { dateStyle: 'medium', timeStyle: 'short', timeZone: 'Asia/Makassar' }).format(date) + ' WITA';
	}

	function sessionStatusLabel(status: AssessmentSessionStatus) {
		return ({ draft: 'Draft', scheduled: 'Terjadwal', active: 'Aktif', finished: 'Selesai', archived: 'Arsip' } as Record<AssessmentSessionStatus, string>)[status] ?? status;
	}

	function sessionStatusTone(status: AssessmentSessionStatus) {
		if (status === 'active') return 'border-sky-200 bg-sky-50 text-sky-700';
		if (status === 'scheduled') return 'border-emerald-200 bg-emerald-50 text-emerald-700';
		if (status === 'finished') return 'border-slate-200 bg-slate-50 text-slate-700';
		if (status === 'archived') return 'border-zinc-200 bg-zinc-50 text-zinc-600';
		return 'border-amber-200 bg-amber-50 text-amber-700';
	}

	function sessionReadinessChecks(row: AssessmentSessionReadiness) {
		const participantCount = Number(row.participant_count ?? 0);
		const roomCount = Number(row.room_count ?? 0);
		const assignedCount = Number(row.assigned_participant_count ?? 0);
		const missingSeatCount = Number(row.missing_seat_count ?? 0);
		const scheduleValid = Boolean(row.scheduled_start && row.scheduled_end && new Date(row.scheduled_end).getTime() > new Date(row.scheduled_start).getTime());
		return [
			{ label: 'Terikat ke Kegiatan', ready: Boolean(row.event_id), helper: 'Badge Kegiatan tampil, bukan sesi mandiri.' },
			{ label: 'Paket Soal dipilih', ready: Boolean(row.package_id), helper: row.package_title || 'Pilih paket dari mapping rombel.' },
			{ label: 'Peserta masuk', ready: participantCount > 0, helper: `${participantCount} peserta` },
			{ label: 'Ruang tersusun', ready: roomCount > 0, helper: `${roomCount} ruang · kapasitas ${row.total_capacity ?? 0}` },
			{ label: 'Kursi lengkap', ready: participantCount > 0 && assignedCount >= participantCount && missingSeatCount === 0, helper: `${assignedCount}/${participantCount} ditempatkan · ${missingSeatCount} tanpa kursi` },
			{ label: 'Jadwal valid', ready: scheduleValid, helper: `${formatDateTimeLabel(row.scheduled_start)} – ${formatDateTimeLabel(row.scheduled_end)}` }
		];
	}

	function sessionReadyToActivate(row: AssessmentSessionReadiness) {
		return sessionReadinessChecks(row).every((check) => check.ready);
	}

	function sessionBlockingChecks(row: AssessmentSessionReadiness) {
		return sessionReadinessChecks(row).filter((check) => !check.ready);
	}

	function sessionActivationSummary(row: AssessmentSessionReadiness) {
		const blockers = sessionBlockingChecks(row);
		if (blockers.length === 0) return 'Siap diaktifkan.';
		return blockers.map((check) => check.label).join(', ');
	}

	function sessionNextAction(row: AssessmentSessionReadiness) {
		const blockers = sessionBlockingChecks(row).map((check) => check.label);
		if (blockers.includes('Peserta masuk') || blockers.includes('Ruang tersusun') || blockers.includes('Kursi lengkap')) {
			return 'Lengkapi Peserta & Ruang dulu, lalu kembali ke Sesi.';
		}
		if (blockers.includes('Jadwal valid')) return 'Periksa tanggal, jam mulai, dan durasi.';
		return 'Review data sesi sebelum aktivasi.';
	}

	function defaultSessionTitle(row: EditablePackageMap | null) {
		if (!row) return selectedKegiatan ? `${selectedKegiatan.nama} · Sesi` : 'Sesi Asesmen';
		const subject = row.subject_name || selectedPackageOption(row.package_id)?.subject_name || 'Mapel';
		const kelas = row.class_code || classOption(row.class_id)?.code || classOption(row.class_id)?.name || 'Rombel';
		return `${subject} ${kelas}`;
	}

	async function loadSessions(kegiatanId = selectedKegiatanId) {
		if (!kegiatanId) return;
		loadingSessions = true;
		sessionError = '';
		try {
			const response = await fetch(clientApiPath`/api/asesmen/exams/${kegiatanId}/sessions`);
			sessionRows = await readClientApiData<AssessmentSessionReadiness[]>(response);
		} catch (error) {
			sessionError = error instanceof Error ? error.message : 'Daftar sesi belum dapat dibuka.';
			sessionRows = [];
		} finally {
			loadingSessions = false;
		}
	}

	async function createSession() {
		if (!selectedKegiatan) return;
		const row = selectedSessionPackageMap;
		if (!row) {
			sessionError = 'Pilih dan simpan minimal satu mapping Paket Soal/Rombel dulu.';
			return;
		}
		const scheduledStart = localDateTimeToIso(sessionDraft.date, sessionDraft.startTime);
		const scheduledEnd = sessionEndIso(sessionDraft.date, sessionDraft.startTime, sessionDraft.durationMinutes);
		if (!scheduledStart || !scheduledEnd) {
			sessionError = 'Tanggal dan jam mulai wajib diisi.';
			return;
		}
		savingSession = true;
		sessionError = '';
		sessionNotice = '';
		try {
			const response = await fetch(clientApiPath`/api/asesmen/exams/${selectedKegiatan.id}/sessions`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					package_id: row.package_id,
					class_id: row.class_id,
					scope_type: 'class',
					scope_ref: row.class_id,
					mix_policy: 'class_grouped',
					assignment_mode: 'balanced',
					title: sessionDraft.title.trim() || defaultSessionTitle(row),
					scheduled_start: scheduledStart,
					scheduled_end: scheduledEnd,
					status: 'draft'
				})
			});
			await readClientApiData<unknown>(response);
			sessionNotice = 'Sesi draft tersimpan dan sudah terikat ke Kegiatan. Review checklist sebelum aktif.';
			sessionDraft = { ...sessionDraft, title: '' };
			await loadSessions(selectedKegiatan.id);
			await loadKegiatan();
		} catch (error) {
			sessionError = error instanceof Error ? error.message : 'Sesi belum dapat dibuat.';
		} finally {
			savingSession = false;
		}
	}

	async function updateSessionStatus(row: AssessmentSessionReadiness, status: AssessmentSessionStatus) {
		if (status === 'active' && !sessionReadyToActivate(row)) {
			sessionError = sessionGateMessage;
			return;
		}
		workingSessionId = row.id;
		sessionError = '';
		sessionNotice = '';
		try {
			const response = await fetch(clientApiPath`/api/asesmen/sessions/${row.id}/status`, {
				method: 'PATCH',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ status })
			});
			await readClientApiData<unknown>(response);
			sessionNotice = status === 'active' ? 'Sesi berhasil diaktifkan.' : 'Status sesi diperbarui.';
			await loadSessions(selectedKegiatan?.id);
			await loadKegiatan();
		} catch (error) {
			sessionError = error instanceof Error ? error.message : 'Status sesi belum dapat diperbarui.';
		} finally {
			workingSessionId = '';
		}
	}

	async function issueParticipantCards(regenerate = false) {
		if (!selectedKegiatan || issuingCards) return;
		if (!packageGateReady) {
			documentError = packageGateMessage;
			return;
		}
		if (selectedKegiatan.peserta <= 0 || selectedKegiatan.ruang <= 0) {
			documentError = 'Lengkapi peserta dan simpan ruang sebelum menerbitkan QR+PIN.';
			return;
		}
		const ok = window.confirm(regenerate
			? 'Buat PIN baru untuk semua kartu peserta? PIN lama akan diganti.'
			: 'Buat QR+PIN kartu peserta sekarang? Setelah tampil, langsung cetak atau simpan PDF.');
		if (!ok) return;
		documentError = '';
		documentNotice = '';
		issuingCards = true;
		try {
			const response = await fetch(clientApiPath`/api/asesmen/exams/${selectedKegiatan.id}/issue-cards`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ regenerate })
			});
			const result = await readClientApiData<IssueCardsResult>(response);
			await setPrintableCards(result.cards || []);
			documentNotice = result.message || `${result.count} kartu peserta siap. PIN hanya tampil pada hasil terbitkan ini.`;
			kegiatan = kegiatan.map((item) => item.id === selectedKegiatan.id ? { ...item, kartu: result.count, catatan: `Data tersimpan di database. Kartu peserta terbit: ${result.count}.` } : item);
			await loadKegiatan();
		} catch (error) {
			documentError = error instanceof Error ? error.message : 'Kartu peserta belum dapat diterbitkan.';
		} finally {
			issuingCards = false;
		}
	}

</script>

<svelte:head>
	<title>Detail Kegiatan Asesmen | MTsN 2 Kolut</title>
</svelte:head>

<div class="app-screen space-y-5 pb-16">
	<section class="rounded-2xl border border-border bg-card p-5 shadow-sm">
		<div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
			<div class="min-w-0 space-y-2">
				<p class="text-xs font-semibold tracking-[0.22em] text-muted-foreground uppercase">Asesmen / Detail Kegiatan</p>
				<h1 class="truncate text-2xl font-bold tracking-tight text-foreground md:text-3xl">{selectedKegiatan?.nama ?? 'Memuat kegiatan…'}</h1>
				<p class="max-w-3xl text-sm leading-6 text-muted-foreground">
					Workspace satu kegiatan untuk Paket Soal, peserta, ruang, sesi, cetak, dan hasil. Halaman daftar hanya menjadi launcher.
				</p>
			</div>
			<div class="flex flex-wrap gap-2">
				<a href="/asesmen" class="inline-flex h-10 items-center justify-center rounded-md border px-4 text-sm font-semibold text-foreground hover:bg-muted">← Kembali ke Daftar</a>
				<button type="button" class="inline-flex h-10 items-center justify-center rounded-md border px-4 text-sm font-semibold text-foreground hover:bg-muted" onclick={loadKegiatan} disabled={loading}>{loading ? 'Memuat…' : 'Muat Ulang'}</button>
				{#if selectedKegiatan}
					<button type="button" class="inline-flex h-10 items-center justify-center rounded-md bg-primary px-4 text-sm font-semibold text-primary-foreground shadow-sm transition hover:bg-primary/90" onclick={() => (activeDetailFeature = nextDetailActionKey(selectedKegiatan))}>Lanjut: {nextDetailActionLabel(selectedKegiatan)}</button>
				{/if}
			</div>
		</div>
	</section>

	{#if listError}<p class="rounded-xl border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive" role="alert">{listError}</p>{/if}

	{#if selectedKegiatan}
		<section class="rounded-2xl border border-primary/20 bg-card shadow-sm" aria-labelledby="detail-kegiatan-title">
			<div class="border-b border-border bg-muted/20 px-5 py-4">
				<div class="flex items-start justify-between gap-3">
					<div class="min-w-0 space-y-2"><p class="text-xs font-semibold tracking-[0.18em] text-primary uppercase">Workspace Kegiatan</p><h2 id="detail-kegiatan-title" class="truncate text-xl font-bold text-foreground">{selectedKegiatan.nama}</h2><div class="flex flex-wrap items-center gap-2"><span class={`rounded-full border px-2.5 py-1 text-xs font-semibold ${statusTone[selectedKegiatan.status]}`}>{selectedKegiatan.status}</span><span class="rounded-full border bg-muted px-2.5 py-1 text-xs font-medium text-muted-foreground">{selectedKegiatan.mode}</span><span class="rounded-full border border-primary/20 bg-primary/10 px-2.5 py-1 text-xs font-semibold text-primary">Halaman penuh</span><span class="rounded-full border bg-background px-2.5 py-1 text-xs font-semibold text-muted-foreground">Kesiapan {readinessScore(selectedKegiatan)}%</span></div></div>
					<button type="button" class="rounded-md border px-3 py-2 text-xs font-semibold text-muted-foreground hover:bg-muted" onclick={() => (activeDetailFeature = nextDetailActionKey(selectedKegiatan))}>Lanjut: {nextDetailActionLabel(selectedKegiatan)}</button>
				</div>
			</div>
			<div class="grid gap-4 px-5 pt-4 pb-5 xl:grid-cols-[20rem_minmax(0,1fr)]">
				<aside class="space-y-4">
					<section class="rounded-xl border bg-background p-4"><h3 class="text-sm font-semibold text-foreground">Ringkasan</h3><div class="mt-3 grid gap-2 text-sm"><div class="flex justify-between gap-3"><span class="text-muted-foreground">Tanggal</span><strong class="text-right font-semibold">{selectedKegiatan.periode}</strong></div><div class="flex justify-between gap-3"><span class="text-muted-foreground">Peserta</span><strong>{selectedKegiatan.peserta}</strong></div><div class="flex justify-between gap-3"><span class="text-muted-foreground">Ruang</span><strong>{selectedKegiatan.ruang}</strong></div><div class="flex justify-between gap-3"><span class="text-muted-foreground">Sesi</span><strong>{selectedKegiatan.sesi}</strong></div></div><p class="mt-3 rounded-lg bg-muted/50 px-3 py-2 text-xs leading-5 text-muted-foreground">{selectedKegiatan.catatan}</p></section>

					<section class="rounded-xl border bg-background p-4">
						<div class="flex items-center justify-between gap-2"><h3 class="text-sm font-semibold text-foreground">Kesiapan Asesmen</h3><span class="text-sm font-bold text-foreground">{readinessScore(selectedKegiatan)}%</span></div>
						<div class="mt-3 h-2 overflow-hidden rounded-full bg-muted"><div class={`h-full rounded-full ${readinessTone(readinessScore(selectedKegiatan))}`} style={`width: ${readinessScore(selectedKegiatan)}%`}></div></div>
						<div class="mt-3 space-y-2">{#each readinessChecks(selectedKegiatan) as check}<div class="flex items-center gap-2 text-xs"><span class={`inline-flex h-5 w-5 items-center justify-center rounded-full border text-[10px] font-bold ${check.ready ? 'border-emerald-200 bg-emerald-50 text-emerald-700' : 'border-slate-200 bg-slate-50 text-slate-500'}`}>{check.ready ? '✓' : '!'}</span><span class={check.ready ? 'text-foreground' : 'text-muted-foreground'}>{check.label}</span></div>{/each}</div>
						<p class="mt-3 rounded-lg bg-muted/40 px-3 py-2 text-xs leading-5 text-muted-foreground">Langkah berikutnya: <strong class="text-foreground">{nextDetailActionLabel(selectedKegiatan)}</strong></p>
					</section>

					<section class="rounded-xl border bg-background p-3">
						<div class="flex items-center justify-between gap-2">
							<h3 class="text-sm font-semibold text-foreground">Alur Kegiatan</h3>
							<p class="text-[11px] text-muted-foreground">UI ringkas</p>
						</div>
						<div class="mt-3 space-y-2">{#each detailFeatures as feature, index}
							{@const state = stepState(feature.key, selectedKegiatan)}
							<button type="button" class={`flex w-full gap-3 rounded-lg border px-3 py-3 text-left text-sm transition ${activeDetailFeature === feature.key ? 'border-primary bg-primary/10 text-primary shadow-sm' : 'bg-card text-foreground hover:bg-muted'}`} aria-pressed={activeDetailFeature === feature.key} onclick={() => (activeDetailFeature = feature.key)}><span class={`mt-0.5 inline-flex size-6 shrink-0 items-center justify-center rounded-full border text-[11px] font-bold ${activeDetailFeature === feature.key ? 'border-primary bg-primary text-primary-foreground' : 'bg-background text-muted-foreground'}`}>{index + 1}</span><span class="min-w-0 flex-1"><span class="flex items-center justify-between gap-2"><span class="block text-sm font-semibold leading-5">{feature.label}</span><span class={`shrink-0 rounded-full border px-2 py-0.5 text-[10px] font-semibold ${state.tone}`}>{state.label}</span></span><span class="mt-0.5 block text-[11px] leading-4 text-muted-foreground">{state.helper || feature.description}</span></span></button>
						{/each}</div>
					</section>
				</aside>

				<div class="min-w-0 space-y-4">
					{#if activeDetailFeature === 'paket'}
					<section class="rounded-xl border border-indigo-200 bg-indigo-50/50 p-4">
						<div class="flex flex-col gap-2 sm:flex-row sm:items-start sm:justify-between">
							<div>
								<p class="text-xs font-semibold tracking-[0.16em] text-indigo-700 uppercase">Prasyarat · Paket Siap</p>
								<h3 class="text-base font-bold text-foreground">Pilih paket dari Modul Paket Soal</h3>
								<p class="mt-1 text-xs leading-5 text-muted-foreground">Bank Soal dan Paket Soal bukan bagian dari Asesmen. Di sini hanya memilih paket yang sudah siap untuk rombel/mapel kegiatan ini.</p>
							</div>
							<div class="flex flex-wrap gap-2">
								<a class="rounded-md border bg-background px-3 py-2 text-xs font-semibold text-foreground hover:bg-muted" href="/paket-soal">Buka Modul Paket Soal</a>
								<button type="button" class="rounded-md border bg-background px-3 py-2 text-xs font-semibold text-foreground hover:bg-muted" onclick={() => void loadPackageOptions()} disabled={loadingPackages}>{loadingPackages ? 'Memuat…' : 'Refresh Paket'}</button>
								<button type="button" class="rounded-md border bg-background px-3 py-2 text-xs font-semibold text-foreground hover:bg-muted" onclick={addPackageMapRow}>Tambah Rombel</button>
							</div>
						</div>
						<div class="mt-3 grid gap-2 sm:grid-cols-3">
								<div class="rounded-lg border bg-background px-3 py-2"><p class="text-xs text-muted-foreground">Paket dipilih</p><p class="text-xl font-bold text-foreground">{packageSelectedCount}</p></div>
							<div class="rounded-lg border bg-background px-3 py-2"><p class="text-xs text-muted-foreground">Rombel terhubung</p><p class="text-xl font-bold text-foreground">{packageClassCount}</p></div>
							<div class="rounded-lg border bg-background px-3 py-2"><p class="text-xs text-muted-foreground">Durasi terpanjang</p><p class="text-xl font-bold text-foreground">{packageMaxDuration}<span class="text-xs font-medium text-muted-foreground"> menit</span></p></div>
						</div>
						{#if packageError}<p class="mt-3 rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive" role="alert">{packageError}</p>{/if}
						{#if packageNotice}<p class="mt-3 rounded-md border border-indigo-300 bg-indigo-100 px-3 py-2 text-sm font-medium text-indigo-800" role="status">{packageNotice}</p>{/if}
						<div class="mt-4 divide-y rounded-xl border bg-background">
							{#if loadingPackages && packageMaps.length === 0}
								<p class="p-4 text-sm text-muted-foreground">Memuat paket soal…</p>
							{:else if packageClassGroups.length === 0}
								<div class="p-4 text-sm text-muted-foreground"><p class="font-semibold text-foreground">Belum ada paket dipilih.</p><p class="mt-1 text-xs leading-5">Klik Tambah Rombel, lalu pilih paket siap dari Modul Paket Soal. Satu rombel boleh memiliki beberapa mapel; satu mapel tetap satu paket.</p></div>
							{:else}
								{#each packageClassGroups as group (group.local_id)}
									<div class="space-y-3 p-3 text-sm">
										<div class="grid gap-3 lg:grid-cols-[1fr_1fr_5rem] lg:items-end">
											<label class="space-y-1.5"><span class="text-xs font-medium text-muted-foreground">Rombel</span><select class="min-w-0 w-full rounded-md border bg-card px-3 py-2 text-sm disabled:opacity-70" value={group.class_id} onchange={(event) => updatePackageClassGroup(group.local_id, event.currentTarget.value)} disabled={group.rows.some((row) => row.id)} title={group.rows.some((row) => row.id) ? 'Hapus mapping tersimpan lalu tambah rombel baru jika ingin pindah rombel.' : undefined}>{#each rombelOptions as rombel}<option value={rombel.id}>{rombel.code || rombel.name} · {rombel.total_students ?? 0} siswa</option>{/each}</select></label>
											<label class="space-y-1.5"><span class="text-xs font-medium text-muted-foreground">Sesi/slot default</span><input class="min-w-0 w-full rounded-md border bg-card px-3 py-2 text-sm" value={group.rows[0]?.slot_label || ''} oninput={(event) => group.rows.forEach((row) => updatePackageMapRow(row.local_id, { slot_label: event.currentTarget.value }))} placeholder="Sesi Utama" /></label>
											<button type="button" class="rounded-md border bg-card px-3 py-2 text-xs font-semibold text-foreground hover:bg-muted" onclick={() => group.rows.forEach((row) => void deletePackageMap(row))}>Hapus</button>
										</div>
										<div class="rounded-lg border bg-muted/20 p-3">
											<div class="flex flex-wrap items-center justify-between gap-2"><p class="text-xs font-semibold text-foreground">Paket soal untuk {classLabel(group.class_id)}</p><p class="text-[11px] text-muted-foreground">Dipilih: {packageRowsForClass(group.class_id).length} paket</p></div>
											<div class="mt-3 grid max-h-72 gap-2 overflow-y-auto pr-1 md:grid-cols-2">
												{#each packageOptions as option (option.id)}
													<label class={`flex cursor-pointer items-start gap-2 rounded-lg border px-3 py-2 text-xs hover:bg-background ${packageCheckedForClass(group.class_id, option.id) ? 'border-indigo-300 bg-indigo-50 text-indigo-900' : 'bg-card text-muted-foreground'}`}>
														<input type="checkbox" class="mt-1" checked={packageCheckedForClass(group.class_id, option.id)} onchange={(event) => void togglePackageForClass(group.class_id, option, event.currentTarget.checked)} />
														<span class="min-w-0"><span class="block font-semibold text-foreground">{option.subject_name}</span><span class="block truncate">{option.title}</span><span class="mt-1 block text-[11px]">{option.question_count ?? 0} soal · {option.duration_minutes || 0} menit{packageOptionForSubjectDuplicate(group.class_id, option) ? ' · mapel sudah dipilih' : ''}</span><span class={`mt-1 inline-flex rounded-full border px-2 py-0.5 text-[10px] font-semibold ${packageReusedFromAnotherKegiatan(option.id) ? 'border-amber-300 bg-amber-50 text-amber-800' : 'border-slate-200 bg-white text-slate-600'}`}>{packageScopeLabel(option.id)}</span></span>
													</label>
												{/each}
											</div>
										</div>
										{#if packageRowsForClass(group.class_id).length > 0}<p class="text-xs leading-5 text-muted-foreground">Mapping otomatis: {classLabel(group.class_id)} → {packageRowsForClass(group.class_id).map((row) => `${row.subject_name || selectedPackageOption(row.package_id)?.subject_name || 'Mapel'} (${row.package_title || selectedPackageOption(row.package_id)?.title || 'Paket'} · ${packageScopeLabel(row.package_id)})`).join(', ')}</p>{/if}
									</div>
								{/each}
							{/if}
						</div>
						<div class="mt-4 flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
							<p class="text-xs leading-5 text-muted-foreground">Rekomendasi: pilih rombel, centang semua paket yang dipakai, lalu lanjut ke peserta/ruang. Sistem tetap menolak dua paket untuk mapel yang sama pada rombel yang sama.</p>
							<button type="button" class="rounded-md bg-primary px-4 py-2 text-sm font-semibold text-primary-foreground hover:bg-primary/90 disabled:opacity-60" onclick={() => void savePackageMaps()} disabled={savingPackages || packageMaps.length === 0}>{savingPackages ? 'Menyimpan…' : 'Simpan Paket Soal'}</button>
						</div>
					</section>
					{/if}

					{#if !packageGateReady && activeDetailFeature !== 'paket'}
						<div class="rounded-xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm leading-6 text-amber-800" role="status">
							<strong class="block text-amber-900">Paket Soal belum siap</strong>
							{packageGateMessage}
						</div>
					{/if}

					{#if activeDetailFeature === 'ruang'}
					<section class="rounded-xl border border-emerald-200 bg-emerald-50/60 p-4">
						<div class="flex flex-col gap-2 sm:flex-row sm:items-start sm:justify-between">
							<div><p class="text-xs font-semibold tracking-[0.16em] text-emerald-700 uppercase">Langkah 3 · Peserta & Ruang</p><h3 class="text-base font-bold text-foreground">Pilih rombel dan susun ruang</h3><p class="mt-1 text-xs leading-5 text-muted-foreground">Alur baru: pilih rombel → atur pola acak → review peta ruang visual → edit manual bila perlu. Kartu/QR+PIN belum diterbitkan.</p></div>
							<button type="button" class="rounded-md border bg-background px-3 py-2 text-xs font-semibold text-foreground hover:bg-muted" onclick={loadRombelOptions} disabled={loadingRombel}>{loadingRombel ? 'Memuat…' : 'Refresh Rombel'}</button>
						</div>

						<div class="mt-4 grid gap-2 sm:grid-cols-4"><div class="rounded-lg border border-emerald-300 bg-background px-3 py-2"><p class="text-xs font-bold text-emerald-700">① Pilih Rombel</p><p class="text-[11px] text-muted-foreground">Tentukan peserta</p></div><div class="rounded-lg border border-emerald-300 bg-background px-3 py-2"><p class="text-xs font-bold text-emerald-700">② Atur Acak</p><p class="text-[11px] text-muted-foreground">Ruang & pola</p></div><div class="rounded-lg border border-emerald-300 bg-background px-3 py-2"><p class="text-xs font-bold text-emerald-700">③ Review Ruang</p><p class="text-[11px] text-muted-foreground">Cek kapasitas</p></div><div class="rounded-lg border border-emerald-300 bg-background px-3 py-2"><p class="text-xs font-bold text-emerald-700">④ Manual</p><p class="text-[11px] text-muted-foreground">Pindah peserta</p></div></div>

						<div class="mt-4 rounded-lg border bg-background p-3">
							<div class="flex flex-wrap items-center justify-between gap-2"><h4 class="text-sm font-semibold text-foreground">① Pilih Rombel Peserta</h4><p class="text-xs text-muted-foreground">Terpilih: {selectedClassIds.length} rombel · ±{selectedStudentTotal} siswa</p></div>
							{#if loadingRombel}<p class="mt-3 text-sm text-muted-foreground">Memuat rombel…</p>{:else if rombelOptions.length === 0}<p class="mt-3 text-sm text-muted-foreground">Belum ada rombel aktif yang bisa dipilih.</p>{:else}<div class="mt-3 grid max-h-56 gap-2 overflow-y-auto pr-1 sm:grid-cols-2">{#each rombelOptions as rombel}<label class="flex cursor-pointer items-start gap-2 rounded-lg border px-3 py-2 text-sm hover:bg-muted/60"><input type="checkbox" class="mt-1" checked={selectedClassIds.includes(rombel.id)} onchange={() => toggleClass(rombel.id)} /><span class="min-w-0"><span class="block font-semibold text-foreground">{rombel.code || rombel.name}</span><span class="block text-xs text-muted-foreground">{rombel.name} · {rombel.total_students ?? 0} siswa</span></span></label>{/each}</div>{/if}
						</div>

						<div class="mt-4 rounded-lg border bg-background p-3">
							<div class="flex flex-wrap items-start justify-between gap-2">
								<div>
									<h4 class="text-sm font-semibold text-foreground">② Atur Acak</h4>
									<p class="mt-1 text-xs leading-5 text-muted-foreground">Pilih metode pembagian. Default terbaik: Acak merata ke seluruh ruang.</p>
								</div>
								<span class="rounded-full border border-emerald-300 bg-emerald-50 px-2.5 py-1 text-[11px] font-semibold text-emerald-700">Rekomendasi: merata</span>
							</div>
							<div class="mt-3 grid gap-3 sm:grid-cols-3">
								<label class="space-y-1"><span class="text-xs font-medium text-muted-foreground">Jumlah ruang</span><input type="number" min="1" max="20" class="w-full rounded-md border bg-background px-3 py-2 text-sm" bind:value={roomCount} /></label>
								<label class="space-y-1"><span class="text-xs font-medium text-muted-foreground">Kapasitas/ruang</span><input type="number" min="1" max="50" class="w-full rounded-md border bg-background px-3 py-2 text-sm" bind:value={capacityPerRoom} /></label>
								<label class="space-y-1"><span class="text-xs font-medium text-muted-foreground">Metode pembagian</span><select class="w-full rounded-md border bg-background px-3 py-2 text-sm" bind:value={assignmentMode}><option value="balanced_all">Acak merata ke seluruh ruang</option><option value="mixed_rombel">Acak campur rombel</option><option value="class_grouped">Kelompok per rombel</option><option value="ordered_participant">Urut nomor peserta</option><option value="csv_manual">Manual dari CSV</option></select></label>
							</div>
							<p class="mt-2 rounded-lg bg-muted/40 px-3 py-2 text-xs leading-5 text-muted-foreground">{assignmentModeDescriptions[assignmentMode]}</p>
							<div class="mt-3 grid gap-2 sm:grid-cols-2">
								<label class="flex items-start gap-2 rounded-lg border bg-card px-3 py-2 text-xs text-muted-foreground"><input type="checkbox" class="mt-1" bind:checked={balanceRooms} /><span><strong class="block text-foreground">Seimbangkan jumlah peserta per ruang</strong><span>Ruang dibuat tidak timpang selama kapasitas masih cukup.</span></span></label>
								<label class="flex items-start gap-2 rounded-lg border bg-card px-3 py-2 text-xs text-muted-foreground"><input type="checkbox" class="mt-1" bind:checked={spreadRombel} /><span><strong class="block text-foreground">Usahakan rombel tidak berkumpul</strong><span>Sistem memakai pola campur bila mode memungkinkan.</span></span></label>
							</div>
							{#if assignmentMode === 'csv_manual'}
								<div class="mt-3 rounded-lg border border-dashed bg-muted/20 px-3 py-3">
									<p class="text-xs font-semibold text-foreground">Manual dari CSV</p>
									<p class="mt-1 text-xs leading-5 text-muted-foreground">Import → Validasi → Preview → Simpan. Gunakan tombol di Mode Manual untuk Download Template CSV dan Import CSV setelah peserta/ruang tersedia.</p>
								</div>
							{/if}
						</div>

						{#if assignmentError}<p class="mt-3 rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive" role="alert">{assignmentError}</p>{/if}
						{#if assignmentNotice}<p class="mt-3 rounded-md border border-emerald-300 bg-emerald-100 px-3 py-2 text-sm font-medium text-emerald-800" role="status">{assignmentNotice}</p>{/if}

						<div class="mt-4 flex flex-col gap-2 sm:flex-row sm:justify-end"><button type="button" class="rounded-md border bg-background px-4 py-2 text-sm font-semibold text-foreground hover:bg-muted disabled:opacity-60" onclick={() => void previewAssignment()} disabled={workingAssignment || !packageGateReady} title={!packageGateReady ? packageGateMessage : undefined}>{workingAssignment ? 'Memproses…' : 'Preview Pembagian Ruang'}</button><button type="button" class="rounded-md bg-primary px-4 py-2 text-sm font-semibold text-primary-foreground hover:bg-primary/90 disabled:opacity-60" onclick={() => void applyAssignment()} disabled={workingAssignment || !packageGateReady} title={!packageGateReady ? packageGateMessage : undefined}>{workingAssignment ? 'Menyimpan…' : 'Simpan Penempatan'}</button></div>

						<div class="mt-4 rounded-lg border bg-background p-3"><div class="flex flex-wrap items-center justify-between gap-2"><h4 class="text-sm font-semibold text-foreground">③ Review Ruang · Peta Ruang Visual</h4><p class="text-xs text-muted-foreground">Klik preview untuk melihat isi ruang sebelum simpan.</p></div>{#if assignmentPreview}{#if assignmentPreview.applied}<div class="mt-3 rounded-lg border border-emerald-200 bg-emerald-50 px-3 py-2 text-xs leading-5 text-emerald-800"><strong class="block text-emerald-900">Ringkasan hasil penempatan tersimpan</strong>Peserta sudah ditempatkan ke ruang/kursi. QR+PIN dan kartu peserta belum diterbitkan dari tahap ini.</div>{/if}<div class="mt-3 grid gap-2 sm:grid-cols-2">{#each assignmentPreview.rooms as room}<div class="rounded-xl border bg-card p-3 text-sm"><div class="flex items-center justify-between gap-2"><strong>{room.code}</strong><span class="rounded-full bg-muted px-2 py-0.5 text-[11px] font-semibold">{room.assigned_count}/{room.capacity}</span></div><div class="mt-2 h-2 overflow-hidden rounded-full bg-muted"><div class="h-full rounded-full bg-emerald-500" style={`width: ${Math.min(100, Math.round((room.assigned_count / Math.max(1, room.capacity)) * 100))}%`}></div></div><p class="mt-2 text-xs text-muted-foreground">{room.name}</p>{#if room.grade_levels?.length}<div class="mt-2 flex flex-wrap gap-1">{#each room.grade_levels as level}<span class="rounded-full border bg-background px-2 py-0.5 text-[11px] font-medium text-muted-foreground">{gradeLabel(level)}</span>{/each}</div>{/if}{#if room.class_summary?.length}<div class="mt-2 flex flex-wrap gap-1">{#each room.class_summary as summary}<span class="rounded-full bg-emerald-50 px-2 py-0.5 text-[11px] font-semibold text-emerald-700" title={`${summary.class_name || summary.class_code} · ${summary.count} siswa`}>{classSummaryLabel(summary)}</span>{/each}</div>{:else}<p class="mt-2 text-[11px] text-muted-foreground">Komposisi rombel tampil setelah penempatan disimpan.</p>{/if}</div>{/each}</div><div class="mt-3 grid gap-2 text-sm sm:grid-cols-4"><div><p class="text-xs text-muted-foreground">Peserta</p><p class="text-xl font-bold">{assignmentPreview.total_participants}</p></div><div><p class="text-xs text-muted-foreground">Tertampung</p><p class="text-xl font-bold">{assignmentPreview.assigned_total}</p></div><div><p class="text-xs text-muted-foreground">Sisa</p><p class="text-xl font-bold">{assignmentPreview.unassigned_total}</p></div><div><p class="text-xs text-muted-foreground">Status</p><p class="text-sm font-semibold">{assignmentPreview.applied ? 'Tersimpan' : 'Preview'}</p></div></div><p class="mt-2 text-xs leading-5 text-muted-foreground">{assignmentPreview.message}</p>{:else}<p class="mt-3 rounded-lg border border-dashed bg-muted/30 px-3 py-3 text-sm text-muted-foreground">Belum ada preview. Pilih rombel dan klik Preview Pembagian Ruang.</p>{/if}</div>
					</section>
					{/if}

					{#if activeDetailFeature === 'sesi'}
					<section class="rounded-xl border bg-background p-4">
						<div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
							<div>
								<p class="text-xs font-semibold tracking-[0.14em] text-muted-foreground uppercase">Langkah 4 · Sesi CBT</p>
								<h3 class="text-base font-bold text-foreground">Kelola jadwal di modul Sesi CBT</h3>
								<p class="mt-1 text-xs leading-5 text-muted-foreground">Agar halaman kegiatan tidak ramai, pembuatan jadwal, matrix sesi, aktivasi, dan monitor dipindahkan ke submodul operasional seperti CBT lama.</p>
							</div>
							<div class="flex flex-wrap gap-2">
								<a class="rounded-md bg-primary px-3 py-2 text-xs font-semibold text-primary-foreground hover:bg-primary/90" href={`/asesmen/sesi?exam_id=${encodeURIComponent(selectedKegiatan.id)}`}>Buka Sesi CBT</a>
								<a class="rounded-md border bg-background px-3 py-2 text-xs font-semibold text-foreground hover:bg-muted" href="/asesmen/sesi-lite">Mode Cepat HP</a>
								<button type="button" class="rounded-md border bg-background px-3 py-2 text-xs font-semibold text-foreground hover:bg-muted" onclick={() => void loadSessions(selectedKegiatan.id)} disabled={loadingSessions}>{loadingSessions ? 'Memuat…' : 'Refresh Ringkasan'}</button>
							</div>
						</div>

						{#if sessionError}<p class="mt-3 rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive" role="alert">{sessionError}</p>{/if}
						{#if sessionNotice}<p class="mt-3 rounded-md border border-blue-300 bg-blue-100 px-3 py-2 text-sm font-medium text-blue-800" role="status">{sessionNotice}</p>{/if}

						<div class="mt-4 grid gap-3 sm:grid-cols-3">
							<div class="rounded-lg border bg-card p-3"><p class="text-xs font-medium text-muted-foreground">Sesi dibuat</p><p class="mt-1 text-2xl font-bold text-foreground">{sessionRows.length}</p><p class="text-[11px] text-muted-foreground">Untuk kegiatan ini</p></div>
							<div class="rounded-lg border bg-card p-3"><p class="text-xs font-medium text-muted-foreground">Aktif/Terjadwal</p><p class="mt-1 text-2xl font-bold text-foreground">{sessionRows.filter((row) => row.status === 'active' || row.status === 'scheduled').length}</p><p class="text-[11px] text-muted-foreground">Siap operasional</p></div>
							<div class="rounded-lg border bg-card p-3"><p class="text-xs font-medium text-muted-foreground">Belum lengkap</p><p class="mt-1 text-2xl font-bold text-foreground">{sessionRows.filter((row) => sessionBlockingChecks(row).length > 0).length}</p><p class="text-[11px] text-muted-foreground">Cek di Sesi CBT</p></div>
						</div>

						{#if sessionRows.length === 0}
							<p class="mt-4 rounded-lg border border-dashed bg-muted/20 px-3 py-3 text-sm text-muted-foreground">Belum ada sesi untuk kegiatan ini. Klik <strong>Buka Sesi CBT</strong> agar operator menjadwalkan dari modul yang lebih sederhana.</p>
						{:else}
							<div class="mt-4 space-y-2">
								{#each sessionRows.slice(0, 4) as row (row.id)}
									<div class="flex flex-col gap-2 rounded-lg border bg-card px-3 py-2 text-sm sm:flex-row sm:items-center sm:justify-between">
										<div class="min-w-0"><p class="truncate font-semibold text-foreground">{row.subject_name || row.title}</p><p class="truncate text-xs text-muted-foreground">{row.class_code || row.class_name || 'Rombel'} · {formatDateTimeLabel(row.scheduled_start)}</p></div>
										<span class={`w-fit rounded-full border px-2 py-0.5 text-[11px] font-semibold ${sessionStatusTone(row.status)}`}>{sessionStatusLabel(row.status)}</span>
									</div>
								{/each}
								{#if sessionRows.length > 4}<p class="text-xs text-muted-foreground">+{sessionRows.length - 4} sesi lain tersedia di modul Sesi CBT.</p>{/if}
							</div>
						{/if}
					</section>
					{/if}

					{#if activeDetailFeature === 'hasil'}
					<section class="rounded-xl border border-slate-200 bg-slate-50/70 p-4">
						<div>
							<p class="text-xs font-semibold tracking-[0.16em] text-slate-700 uppercase">Langkah 6 · Hasil</p>
							<h3 class="text-base font-bold text-foreground">Ringkasan hasil dan arsip</h3>
							<p class="mt-1 text-xs leading-5 text-muted-foreground">Belum ada data hasil yang dibuka pada P1. Untuk UI/UX, area ini disiapkan sebagai pintu hasil, berita acara, dan arsip.</p>
						</div>
						<div class="mt-4 grid gap-2 sm:grid-cols-3">
							<div class="rounded-lg border bg-background p-3"><p class="text-xs font-semibold text-foreground">Nilai</p><p class="mt-1 text-xs leading-5 text-muted-foreground">Rekap nilai per paket dan rombel.</p></div>
							<div class="rounded-lg border bg-background p-3"><p class="text-xs font-semibold text-foreground">Berita Acara</p><p class="mt-1 text-xs leading-5 text-muted-foreground">Catatan pelaksanaan dan kendala ruang.</p></div>
							<div class="rounded-lg border bg-background p-3"><p class="text-xs font-semibold text-foreground">Arsip</p><p class="mt-1 text-xs leading-5 text-muted-foreground">Dokumen akhir kegiatan ujian.</p></div>
						</div>
					</section>
					{/if}


					{#if activeDetailFeature === 'manual'}
					<section class="rounded-xl border border-sky-200 bg-sky-50/50 p-4">
						<div class="flex flex-col gap-2 sm:flex-row sm:items-start sm:justify-between">
							<div><p class="text-xs font-semibold tracking-[0.16em] text-sky-700 uppercase">Mode Manual · Acak Sendiri</p><h3 class="text-base font-bold text-foreground">Atur peserta per ruang dan nomor kursi</h3><p class="mt-1 text-xs leading-5 text-muted-foreground">Ambil daftar peserta setelah simpan ruang. Tombol Acak Tampilan hanya mengubah urutan tampil agar Bapak bisa memilih manual; penyimpanan tetap per peserta lewat tombol Pindah.</p></div>
							<div class="flex flex-wrap gap-2"><button type="button" class="rounded-md border bg-background px-3 py-2 text-xs font-semibold text-foreground hover:bg-muted" onclick={() => void loadParticipantPlacements()} disabled={loadingPlacements}>{loadingPlacements ? 'Memuat…' : 'Ambil Peserta'}</button><button type="button" class="rounded-md border bg-background px-3 py-2 text-xs font-semibold text-foreground hover:bg-muted" onclick={shuffleManualView} disabled={participantPlacements.length === 0}>Acak Tampilan</button><button type="button" class="rounded-md border bg-background px-3 py-2 text-xs font-semibold text-foreground hover:bg-muted" onclick={downloadPlacementTemplateCsv}>Download Template CSV</button><label class="cursor-pointer rounded-md border bg-background px-3 py-2 text-xs font-semibold text-foreground hover:bg-muted">Import CSV<input class="sr-only" type="file" accept=".csv,text/csv" onchange={(event) => void handlePlacementCsvImport(event)} /></label></div>
						</div>
						{#if placementError}<p class="mt-3 rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive" role="alert">{placementError}</p>{/if}
						{#if csvImportNotice}<p class="mt-3 rounded-md border border-sky-300 bg-sky-100 px-3 py-2 text-sm font-medium text-sky-800" role="status">{csvImportNotice}</p>{/if}
						{#if participantPlacements.length === 0}<p class="mt-3 rounded-lg border border-dashed bg-background px-3 py-3 text-sm text-muted-foreground">Belum ada peserta dimuat. Simpan ruang & peserta dulu, lalu klik Ambil Peserta untuk edit manual.</p>{:else}
							<div class="mt-3 max-h-80 space-y-2 overflow-y-auto pr-1">{#each participantPlacements as placement (placement.participant_id)}<div class="grid gap-2 rounded-lg border bg-background p-3 text-sm sm:grid-cols-[1fr_8rem_6rem_5rem]"><div class="min-w-0"><p class="truncate font-semibold text-foreground">{placement.student_name}</p><p class="text-xs text-muted-foreground">{placement.class_code} · {placement.room_code || 'Belum ruang'} · Kursi {placement.seat_no || '-'}</p></div><select class="rounded-md border bg-card px-2 py-2 text-xs" value={placement.room_id || ''} onchange={(event) => (placement.room_id = event.currentTarget.value)}>{#each manualRoomOptions as room}<option value={room.room_id}>{room.room_code} · {room.room_name}</option>{/each}</select><input class="rounded-md border bg-card px-2 py-2 text-xs" type="number" min="1" max={placement.room_capacity || capacityPerRoom} value={placement.seat_no || 1} oninput={(event) => (placement.seat_no = Number(event.currentTarget.value))} /><button type="button" class="rounded-md bg-primary px-3 py-2 text-xs font-semibold text-primary-foreground disabled:opacity-60" onclick={() => void moveParticipantSeat(placement, placement.room_id || '', Number(placement.seat_no || 1))} disabled={workingParticipantId === placement.participant_id}>{workingParticipantId === placement.participant_id ? '...' : 'Pindah'}</button></div>{/each}</div>
						{/if}
					</section>
					{/if}

					{#if activeDetailFeature === 'cetak'}
					<section class="rounded-xl border border-emerald-200 bg-emerald-50/40 p-4">
						<div>
							<p class="text-xs font-semibold tracking-[0.16em] text-primary uppercase">Cetak</p>
							<h3 class="text-base font-bold text-foreground">Dokumen CBT</h3>
							<p class="mt-1 text-xs leading-5 text-muted-foreground">Pilih dokumen yang ingin dicetak.</p>
						</div>
						{#if documentError}<p class="mt-3 rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive" role="alert">{documentError}</p>{/if}
						{#if documentNotice}<p class="mt-3 rounded-md border border-emerald-300 bg-emerald-100 px-3 py-2 text-sm font-medium text-emerald-900" role="status">{documentNotice}</p>{/if}
						<div class="mt-4 space-y-3">
							<div class="rounded-xl border bg-background p-3 text-sm">
								<div class="flex items-start justify-between gap-3">
									<div class="min-w-0"><p class="font-semibold text-foreground">Kartu Peserta</p><p class="mt-1 text-xs text-muted-foreground">{selectedKegiatan.kartu || printableCards.length} kartu</p></div>
									<span class="rounded-full border bg-card px-2.5 py-1 text-[11px] font-semibold text-muted-foreground">{documentPrintSummary.participantCards.label}</span>
								</div>
								<p class="mt-3 rounded-md bg-muted/60 px-3 py-2 text-xs leading-5 text-muted-foreground">Masalah PIN lama: kalau PIN tidak muncul, buat PIN baru lalu cetak kartu.</p>
								<div class="mt-3 grid gap-2 sm:grid-cols-2">
									<button type="button" class="rounded-md bg-primary px-4 py-3 text-sm font-semibold text-primary-foreground hover:bg-primary/90 disabled:opacity-60" onclick={() => void (documentPrintSummary.participantCards.state === 'ready' ? printParticipantCards() : issueParticipantCards(false))} disabled={issuingCards || checkingCards || !packageGateReady || documentPrintSummary.participantCards.state === 'blocked'} title={!packageGateReady ? packageGateMessage : undefined}>{issuingCards || checkingCards ? 'Memproses…' : documentPrintSummary.participantCards.state === 'ready' ? 'Cetak Kartu' : 'Buat QR+PIN'}</button>
									<button type="button" class="rounded-md border border-primary/30 bg-background px-4 py-3 text-sm font-semibold text-primary hover:bg-primary/5 disabled:opacity-60" onclick={() => void regenerateParticipantCardsAndPrint()} disabled={issuingCards || !packageGateReady || documentPrintSummary.participantCards.state === 'blocked'} title="Ganti PIN lama yang tidak bisa dibuka dari hash">PIN Baru & Cetak</button>
								</div>
							</div>
							<div class="rounded-xl border bg-background p-3 text-sm">
								<div class="flex items-start justify-between gap-3"><div class="min-w-0"><p class="font-semibold text-foreground">Lembar Pengawas</p><p class="mt-1 text-xs text-muted-foreground">Daftar hadir per ruang</p></div><span class="rounded-full border bg-card px-2.5 py-1 text-[11px] font-semibold text-muted-foreground">{documentPrintSummary.supervisorSheets.label}</span></div>
								<button type="button" class="mt-3 w-full rounded-md border bg-card px-4 py-3 text-sm font-semibold text-foreground hover:bg-muted disabled:opacity-60" disabled={documentPrintSummary.supervisorSheets.state === 'blocked'} onclick={() => void printSupervisorSheets()}>Cetak Lembar Pengawas</button>
							</div>
							<div class="rounded-xl border bg-background p-3 text-sm">
								<div class="flex items-start justify-between gap-3"><div class="min-w-0"><p class="font-semibold text-foreground">Checklist Arsip</p><p class="mt-1 text-xs text-muted-foreground">Kelengkapan dokumen</p></div><span class="rounded-full border bg-card px-2.5 py-1 text-[11px] font-semibold text-muted-foreground">{documentPrintSummary.archiveChecklist.label}</span></div>
								<button type="button" class="mt-3 w-full rounded-md border bg-card px-4 py-3 text-sm font-semibold text-foreground hover:bg-muted" onclick={() => void printDocument('checklist')}>Cetak Checklist</button>
							</div>
						</div>
					</section>
					{/if}

					{#if activeDetailFeature === 'manual'}
					<section class="rounded-xl border bg-background p-4"><h3 class="text-sm font-semibold text-foreground">Checklist Persiapan</h3><div class="mt-3 space-y-2">{#each preparationChecklist as label, index}<div class="flex items-center gap-3 rounded-lg border bg-card px-3 py-2 text-sm"><span class="inline-flex h-6 w-6 shrink-0 items-center justify-center rounded-full border text-[11px] font-semibold text-muted-foreground">{index + 1}</span><span class="min-w-0 flex-1 text-foreground">{label}</span><span class="rounded-full bg-muted px-2 py-0.5 text-[11px] font-medium text-muted-foreground">Bertahap</span></div>{/each}</div></section>
					{/if}
				</div>
			</div>
		</section>
	{:else if !loading}
		<section class="rounded-2xl border border-dashed border-border bg-card p-6 text-center shadow-sm"><p class="text-sm font-semibold text-foreground">Kegiatan tidak ditemukan atau belum bisa dimuat.</p><p class="mt-1 text-xs text-muted-foreground">Kembali ke daftar asesmen, lalu pilih kegiatan yang tersedia.</p><a href="/asesmen" class="mt-4 inline-flex rounded-md bg-primary px-4 py-2 text-sm font-semibold text-primary-foreground">Kembali ke Daftar</a></section>
	{/if}
</div>

{#if selectedKegiatan && printMode}
	<section class="print-area">
		<header class="print-header">
			<p>Kementerian Agama · MTsN 2 Kolaka Utara</p>
			<h1>{printMode === 'cards' ? 'Kartu Peserta CBT' : printMode === 'supervisor' ? 'Lembar Pengawas Ruang' : 'Checklist Arsip Asesmen'}</h1>
			<div>{selectedKegiatan.nama} · {selectedKegiatan.periode}</div>
		</header>

		{#if printMode === 'cards'}
			<div class="card-pages">
				{#each printableCardPages as pageCards}
					<section class="card-page">
						{#each pageCards as card}
							<article class="participant-card-print">
								<div class="card-title">KARTU PESERTA CBT</div>
								<h2>{card.student_name || 'Nama peserta'}</h2>
								<div class="card-row"><span>Rombel</span><strong>{card.class_code || card.class_name || '-'}</strong></div>
								<div class="card-row"><span>Ruang/Kursi</span><strong>{card.room_code || '-'} / {card.seat_no || '-'}</strong></div>
								<div class="card-row"><span>No. Induk</span><strong>{card.nis || card.nisn || '-'}</strong></div>
								<div class="qr-box">
									{#if qrImages[cardPrintKey(card)]}
										<img src={qrImages[cardPrintKey(card)]} alt={`QR login ${card.student_name || 'peserta'}`} />
									{:else}
										<span>QR code tidak tersedia. Buat PIN baru untuk mencetak kartu login.</span>
									{/if}
								</div>
								<div class="pin-row"><span>PIN</span><strong>{card.pin || '—'}</strong></div>
								<p class="print-note">Gunakan QR/PIN ini hanya untuk pelaksanaan resmi. Simpan kartu dengan aman.</p>
							</article>
						{/each}
					</section>
				{/each}
			</div>
		{:else if printMode === 'supervisor'}
			{#each printableRoomGroups as [roomCode, cards]}
				<section class="room-sheet">
					<h2>Ruang {roomCode}</h2>
					<table>
						<thead><tr><th>No</th><th>Nama</th><th>Rombel</th><th>Kursi</th><th>QR</th><th>Tanda Tangan</th></tr></thead>
						<tbody>{#each cards as card, index}<tr><td>{index + 1}</td><td>{card.student_name || '-'}</td><td>{card.class_code || card.class_name || '-'}</td><td>{card.seat_no || '-'}</td><td>{#if qrImages[cardPrintKey(card)]}<img class="sheet-qr" src={qrImages[cardPrintKey(card)]} alt="QR" />{:else}<span class="qr-missing">-</span>{/if}</td><td></td></tr>{/each}</tbody>
					</table>
					<div class="signature-row"><span>Pengawas Ruang</span><span>Panitia</span></div>
				</section>
			{/each}
		{:else}
			<section class="archive-checklist-print">
				<h2>Checklist Arsip</h2>
				{#each preparationChecklist as label, index}
					<div class="check-row"><span>{index + 1}</span><strong>{label}</strong><em>□ Ada / □ Belum</em></div>
				{/each}
				<p class="print-note">Checklist dicetak untuk arsip panitia. Cocokkan kembali jumlah peserta, ruang, sesi, kartu, dan lembar pengawas sebelum pelaksanaan.</p>
			</section>
		{/if}
	</section>
{/if}

<style>
	.print-area { display: none; }
	@media print {
		@page { size: A4 portrait; margin: 8mm; }
		.print-area { display: none !important; }
		:global(body) { background: white !important; color: #111827 !important; }
		:global(body.cbt-document-print) { margin: 0 !important; }
		:global(body.cbt-document-print *) { visibility: hidden !important; }
		:global(body.cbt-document-print .print-area),
		:global(body.cbt-document-print .print-area *) { visibility: visible !important; }
		:global(body.cbt-document-print) .app-screen { display: none !important; }
		:global(body.cbt-document-print .print-area) {
			display: block !important;
			position: absolute !important;
			inset: 0 auto auto 0 !important;
			width: 100% !important;
			padding: 0;
			font-family: Arial, sans-serif;
			color: #111827;
		}
		.print-header { border-bottom: 1.5px solid #111827; margin-bottom: 4mm; padding-bottom: 2mm; text-align: center; }
		.print-header p { margin: 0; font-size: 8px; text-transform: uppercase; letter-spacing: .08em; }
		.print-header h1 { margin: 2px 0; font-size: 13px; }
		.print-header div { font-size: 9px; }
		.card-pages { display: block; }
		.card-page { break-after: page; display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); grid-template-rows: repeat(3, 1fr); gap: 3mm; min-height: 248mm; page-break-after: always; }
		.card-page:last-child { break-after: auto; page-break-after: auto; }
		.participant-card-print { break-inside: avoid; page-break-inside: avoid; border: 1px solid #111827; border-radius: 7px; box-sizing: border-box; height: 80mm; overflow: hidden; padding: 3.5mm; }
		.participant-card-print .card-title { font-size: 7px; font-weight: 700; letter-spacing: .1em; text-align: center; }
		.participant-card-print h2 { margin: 2mm 0; font-size: 11px; line-height: 1.2; text-align: center; }
		.card-row, .pin-row { display: flex; justify-content: space-between; gap: 6px; border-bottom: 1px dashed #9ca3af; padding: 1.1mm 0; font-size: 7.5px; line-height: 1.15; }
		.pin-row strong { font-size: 13px; letter-spacing: .1em; }
		.qr-box { align-items: center; border: 1px solid #111827; display: flex; font-size: 7px; justify-content: center; margin: 2.5mm auto; min-height: 22mm; padding: 1mm; text-align: center; width: 22mm; }
		.qr-box img { display: block; height: 20mm; width: 20mm; }
		.print-note { color: #4b5563; font-size: 6.5px; line-height: 1.25; margin-top: 1.5mm; }
		.room-sheet { page-break-after: always; }
		.room-sheet h2, .archive-checklist-print h2 { font-size: 16px; margin: 12px 0 8px; }
		table { border-collapse: collapse; width: 100%; }
		th, td { border: 1px solid #111827; font-size: 11px; padding: 5px; text-align: left; vertical-align: middle; }
		th:first-child, td:first-child, th:nth-child(4), td:nth-child(4) { text-align: center; width: 38px; }
		th:nth-child(5), td:nth-child(5) { text-align: center; width: 54px; }
		.sheet-qr { display: inline-block; height: 42px; width: 42px; }
		.qr-missing { color: #6b7280; font-size: 10px; }
		.signature-row { display: flex; justify-content: space-between; margin-top: 42px; padding: 0 48px; font-size: 12px; }
		.check-row { align-items: center; border: 1px solid #d1d5db; display: grid; grid-template-columns: 32px 1fr 110px; gap: 8px; margin-bottom: 6px; padding: 8px; font-size: 12px; }
	}
</style>
