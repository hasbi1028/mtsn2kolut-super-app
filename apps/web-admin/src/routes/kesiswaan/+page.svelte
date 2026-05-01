<script lang="ts">
	import { onMount } from 'svelte';
	import type { PageData } from './$types';
	import * as Card from '$lib/components/ui/card';
	import * as Dialog from '$lib/components/ui/dialog';
	import * as Table from '$lib/components/ui/table';
	import * as Tabs from '$lib/components/ui/tabs';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { Textarea } from '$lib/components/ui/textarea';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import EmptyStatePanel from '$lib/components/EmptyStatePanel.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { toast } from '$lib/components/ui/sonner';
	import { confirmAction } from '$lib/confirm-dialog';

	let { data }: { data: PageData } = $props();

	type Stats = {
		active_students: number;
		prospective_students: number;
		students_with_points: number;
		total_violation_points: number;
		open_violations: number;
		achievements_this_year: number;
		active_categories: number;
		active_extracurriculars: number;
		open_counseling_sessions: number;
		transfers_this_year: number;
	};

	type ClassOption = { id: string; code: string; name: string; level: string };

	type StudentRow = {
		id: string;
		nis: string;
		nisn: string;
		nama: string;
		gender: string;
		parent_name: string;
		parent_phone: string;
		class_id?: string | null;
		class_name?: string | null;
		class_code?: string | null;
		status: string;
		nik: string;
		tempat_lahir: string;
		tanggal_lahir?: string | null;
		alamat: string;
		agama: string;
		anak_ke?: number | null;
		phone: string;
		photo_url: string;
		total_violation_points: number;
		violation_count: number;
		achievement_count: number;
	};

	type CategoryRow = {
		id: string;
		code: string;
		name: string;
		point: number;
		severity: string;
		description: string;
		is_active: boolean;
	};

	type ViolationRow = {
		id: string;
		student_id: string;
		category_id?: string | null;
		incident_date: string;
		points: number;
		description: string;
		action_taken: string;
		status: string;
		student_name: string;
		student_nis: string;
		class_name?: string | null;
		class_code?: string | null;
		category_code?: string | null;
		category_name?: string | null;
		reported_by_name?: string | null;
	};

	type AchievementRow = {
		id: string;
		student_id: string;
		achievement_date: string;
		title: string;
		level: string;
		category: string;
		organizer: string;
		description: string;
		document_url: string;
		student_name: string;
		student_nis: string;
		class_name?: string | null;
		class_code?: string | null;
	};

	type ExtracurricularRow = {
		id: string;
		code: string;
		name: string;
		category: string;
		description: string;
		supervisor_employee_id?: string | null;
		supervisor_name: string;
		schedule_text: string;
		is_active: boolean;
		active_member_count: number;
	};

	type ExtracurricularMemberRow = {
		id: string;
		extracurricular_id: string;
		student_id: string;
		joined_at: string;
		role: string;
		status: string;
		notes: string;
		extracurricular_code: string;
		extracurricular_name: string;
		student_name: string;
		student_nis: string;
		class_name?: string | null;
		class_code?: string | null;
	};

	type CounselingRow = {
		id: string;
		student_id: string;
		session_date: string;
		topic: string;
		summary: string;
		follow_up: string;
		status: string;
		is_confidential: boolean;
		counselor_name: string;
		student_name: string;
		student_nis: string;
		class_name?: string | null;
		class_code?: string | null;
	};

	type TransferRow = {
		id: string;
		student_id: string;
		transfer_date: string;
		transfer_type: string;
		previous_school: string;
		destination_school: string;
		reason: string;
		document_ref: string;
		notes: string;
		status: string;
		student_name: string;
		student_nis: string;
		student_status: string;
		class_name?: string | null;
		class_code?: string | null;
	};

	type KesiswaanData = {
		stats: Stats;
		classes: ClassOption[];
		students: StudentRow[];
		categories: CategoryRow[];
		violations: ViolationRow[];
		achievements: AchievementRow[];
		extracurriculars: ExtracurricularRow[];
		extracurricularMembers: ExtracurricularMemberRow[];
		counselingSessions: CounselingRow[];
		transfers: TransferRow[];
	};

	type ProfileForm = {
		nik: string;
		tempat_lahir: string;
		tanggal_lahir: string;
		alamat: string;
		agama: string;
		anak_ke: number | null;
		phone: string;
		parent_name: string;
		parent_phone: string;
	};

	type CategoryForm = {
		code: string;
		name: string;
		point: number;
		severity: string;
		description: string;
		is_active: boolean;
	};

	type ViolationForm = {
		student_id: string;
		category_id: string;
		incident_date: string;
		points: number;
		description: string;
		action_taken: string;
		status: string;
	};

	type AchievementForm = {
		student_id: string;
		achievement_date: string;
		title: string;
		level: string;
		category: string;
		organizer: string;
		description: string;
		document_url: string;
	};

	type ExtracurricularForm = {
		code: string;
		name: string;
		category: string;
		description: string;
		supervisor_employee_id: string;
		schedule_text: string;
		is_active: boolean;
	};

	type ExtracurricularMemberForm = {
		extracurricular_id: string;
		student_id: string;
		joined_at: string;
		role: string;
		status: string;
		notes: string;
	};

	type CounselingForm = {
		student_id: string;
		session_date: string;
		topic: string;
		summary: string;
		follow_up: string;
		status: string;
		is_confidential: boolean;
	};

	type TransferForm = {
		student_id: string;
		transfer_date: string;
		transfer_type: string;
		previous_school: string;
		destination_school: string;
		reason: string;
		document_ref: string;
		notes: string;
	};

	const canManage = $derived(data.canManage);
	const STUDENT_STATUSES = [
		['prospective', 'Calon'],
		['active', 'Aktif'],
		['alumni', 'Alumni'],
		['mutated', 'Mutasi'],
	] as const;
	const SEVERITIES = [
		['ringan', 'Ringan'],
		['sedang', 'Sedang'],
		['berat', 'Berat'],
	] as const;
	const VIOLATION_STATUSES = [
		['open', 'Perlu tindak lanjut'],
		['resolved', 'Selesai'],
		['canceled', 'Dibatalkan'],
	] as const;
	const ACHIEVEMENT_LEVELS = [
		['school', 'Sekolah'],
		['district', 'Kab/Kota'],
		['province', 'Provinsi'],
		['national', 'Nasional'],
		['international', 'Internasional'],
	] as const;
	const EXTRACURRICULAR_ROLES = [
		['member', 'Anggota'],
		['leader', 'Ketua'],
		['assistant', 'Pengurus'],
	] as const;
	const EXTRACURRICULAR_STATUSES = [
		['active', 'Aktif'],
		['inactive', 'Nonaktif'],
		['alumni', 'Alumni'],
	] as const;
	const COUNSELING_STATUSES = [
		['open', 'Terbuka'],
		['monitoring', 'Monitoring'],
		['resolved', 'Selesai'],
		['referred', 'Dirujuk'],
		['canceled', 'Dibatalkan'],
	] as const;
	const TRANSFER_TYPES = [
		['out', 'Mutasi Keluar'],
		['in', 'Mutasi Masuk'],
	] as const;

	let pagePromise = $state<Promise<KesiswaanData> | null>(null);
	let snapshot = $state<KesiswaanData>(emptyData());
	let activeTab = $state('siswa');
	let studentSearch = $state('');
	let violationSearch = $state('');
	let achievementSearch = $state('');
	let categorySearch = $state('');
	let extracurricularSearch = $state('');
	let memberSearch = $state('');
	let counselingSearch = $state('');
	let transferSearch = $state('');
	let dialogOpen = $state(false);
	let dialogKind = $state<'profile' | 'category' | 'violation' | 'achievement' | 'extracurricular' | 'member' | 'counseling' | 'transfer' | null>(null);
	let editingId = $state<string | null>(null);
	let busy = $state(false);
	let selectedPhotoFile = $state<File | null>(null);

	let profileForm = $state<ProfileForm>(emptyProfileForm());
	let categoryForm = $state<CategoryForm>(emptyCategoryForm());
	let violationForm = $state<ViolationForm>(emptyViolationForm());
	let achievementForm = $state<AchievementForm>(emptyAchievementForm());
	let extracurricularForm = $state<ExtracurricularForm>(emptyExtracurricularForm());
	let memberForm = $state<ExtracurricularMemberForm>(emptyExtracurricularMemberForm());
	let counselingForm = $state<CounselingForm>(emptyCounselingForm());
	let transferForm = $state<TransferForm>(emptyTransferForm());

	const filteredStudents = $derived.by(() => {
		const q = studentSearch.trim().toLowerCase();
		if (!q) return snapshot.students;
		return snapshot.students.filter((student) =>
			[student.nama, student.nis, student.nisn, student.nik, student.class_name ?? '', student.class_code ?? ''].some((value) =>
				value.toLowerCase().includes(q)
			)
		);
	});

	const filteredViolations = $derived.by(() => {
		const q = violationSearch.trim().toLowerCase();
		if (!q) return snapshot.violations;
		return snapshot.violations.filter((item) =>
			[item.student_name, item.student_nis, item.description, item.action_taken, item.category_name ?? '', item.class_name ?? ''].some((value) =>
				value.toLowerCase().includes(q)
			)
		);
	});

	const filteredAchievements = $derived.by(() => {
		const q = achievementSearch.trim().toLowerCase();
		if (!q) return snapshot.achievements;
		return snapshot.achievements.filter((item) =>
			[item.student_name, item.student_nis, item.title, item.category, item.organizer, item.class_name ?? ''].some((value) =>
				value.toLowerCase().includes(q)
			)
		);
	});

	const filteredCategories = $derived.by(() => {
		const q = categorySearch.trim().toLowerCase();
		if (!q) return snapshot.categories;
		return snapshot.categories.filter((item) =>
			[item.code, item.name, item.severity, item.description].some((value) => value.toLowerCase().includes(q))
		);
	});

	const filteredExtracurriculars = $derived.by(() => {
		const q = extracurricularSearch.trim().toLowerCase();
		if (!q) return snapshot.extracurriculars;
		return snapshot.extracurriculars.filter((item) =>
			[item.code, item.name, item.category, item.description, item.schedule_text, item.supervisor_name].some((value) => value.toLowerCase().includes(q))
		);
	});

	const filteredMembers = $derived.by(() => {
		const q = memberSearch.trim().toLowerCase();
		if (!q) return snapshot.extracurricularMembers;
		return snapshot.extracurricularMembers.filter((item) =>
			[item.extracurricular_name, item.extracurricular_code, item.student_name, item.student_nis, item.class_name ?? '', item.notes].some((value) =>
				value.toLowerCase().includes(q)
			)
		);
	});

	const filteredCounseling = $derived.by(() => {
		const q = counselingSearch.trim().toLowerCase();
		if (!q) return snapshot.counselingSessions;
		return snapshot.counselingSessions.filter((item) =>
			[item.student_name, item.student_nis, item.topic, item.summary, item.follow_up, item.counselor_name].some((value) => value.toLowerCase().includes(q))
		);
	});

	const filteredTransfers = $derived.by(() => {
		const q = transferSearch.trim().toLowerCase();
		if (!q) return snapshot.transfers;
		return snapshot.transfers.filter((item) =>
			[item.student_name, item.student_nis, item.previous_school, item.destination_school, item.reason, item.document_ref].some((value) =>
				value.toLowerCase().includes(q)
			)
		);
	});

	function emptyData(): KesiswaanData {
		return {
			stats: {
				active_students: 0,
				prospective_students: 0,
				students_with_points: 0,
				total_violation_points: 0,
				open_violations: 0,
				achievements_this_year: 0,
				active_categories: 0,
				active_extracurriculars: 0,
				open_counseling_sessions: 0,
				transfers_this_year: 0,
			},
			classes: [],
			students: [],
			categories: [],
			violations: [],
			achievements: [],
			extracurriculars: [],
			extracurricularMembers: [],
			counselingSessions: [],
			transfers: [],
		};
	}

	function emptyProfileForm(): ProfileForm {
		return { nik: '', tempat_lahir: '', tanggal_lahir: '', alamat: '', agama: '', anak_ke: null, phone: '', parent_name: '', parent_phone: '' };
	}

	function emptyCategoryForm(): CategoryForm {
		return { code: '', name: '', point: 0, severity: 'ringan', description: '', is_active: true };
	}

	function emptyViolationForm(): ViolationForm {
		return {
			student_id: '',
			category_id: '',
			incident_date: new Date().toISOString().slice(0, 10),
			points: 0,
			description: '',
			action_taken: '',
			status: 'open',
		};
	}

	function emptyAchievementForm(): AchievementForm {
		return {
			student_id: '',
			achievement_date: new Date().toISOString().slice(0, 10),
			title: '',
			level: 'school',
			category: '',
			organizer: '',
			description: '',
			document_url: '',
		};
	}

	function emptyExtracurricularForm(): ExtracurricularForm {
		return { code: '', name: '', category: '', description: '', supervisor_employee_id: '', schedule_text: '', is_active: true };
	}

	function emptyExtracurricularMemberForm(): ExtracurricularMemberForm {
		return {
			extracurricular_id: '',
			student_id: '',
			joined_at: new Date().toISOString().slice(0, 10),
			role: 'member',
			status: 'active',
			notes: '',
		};
	}

	function emptyCounselingForm(): CounselingForm {
		return {
			student_id: '',
			session_date: new Date().toISOString().slice(0, 10),
			topic: '',
			summary: '',
			follow_up: '',
			status: 'open',
			is_confidential: false,
		};
	}

	function emptyTransferForm(): TransferForm {
		return {
			student_id: '',
			transfer_date: new Date().toISOString().slice(0, 10),
			transfer_type: 'out',
			previous_school: '',
			destination_school: '',
			reason: '',
			document_ref: '',
			notes: '',
		};
	}

	function readApi<T>(response: Response, fallback: string): Promise<T> {
		return response.json().then((payload: unknown) => {
			if (!response.ok) {
				const message = typeof payload === 'object' && payload !== null && 'error' in payload ? String((payload as { error?: unknown }).error) : fallback;
				throw new Error(message || fallback);
			}
			return payload as T;
		});
	}

	async function fetchData(): Promise<KesiswaanData> {
		const [
			statsRes,
			classesRes,
			studentsRes,
			categoriesRes,
			violationsRes,
			achievementsRes,
			extracurricularsRes,
			membersRes,
			counselingRes,
			transfersRes,
		] = await Promise.all([
			fetch('/api/kesiswaan/stats'),
			fetch('/api/kesiswaan/classes'),
			fetch('/api/kesiswaan/students'),
			fetch('/api/kesiswaan/violation-categories'),
			fetch('/api/kesiswaan/violations'),
			fetch('/api/kesiswaan/achievements'),
			fetch('/api/kesiswaan/extracurriculars'),
			fetch('/api/kesiswaan/extracurricular-members'),
			fetch('/api/kesiswaan/counseling-sessions'),
			fetch('/api/kesiswaan/student-transfers'),
		]);
		const next = {
			stats: await readApi<Stats>(statsRes, 'Gagal memuat statistik kesiswaan.'),
			classes: await readApi<ClassOption[]>(classesRes, 'Gagal memuat kelas.'),
			students: await readApi<StudentRow[]>(studentsRes, 'Gagal memuat siswa.'),
			categories: await readApi<CategoryRow[]>(categoriesRes, 'Gagal memuat kategori pelanggaran.'),
			violations: await readApi<ViolationRow[]>(violationsRes, 'Gagal memuat pelanggaran.'),
			achievements: await readApi<AchievementRow[]>(achievementsRes, 'Gagal memuat prestasi.'),
			extracurriculars: await readApi<ExtracurricularRow[]>(extracurricularsRes, 'Gagal memuat ekskul.'),
			extracurricularMembers: await readApi<ExtracurricularMemberRow[]>(membersRes, 'Gagal memuat anggota ekskul.'),
			counselingSessions: await readApi<CounselingRow[]>(counselingRes, 'Gagal memuat catatan BK.'),
			transfers: await readApi<TransferRow[]>(transfersRes, 'Gagal memuat mutasi siswa.'),
		};
		snapshot = next;
		return next;
	}

	function load() {
		pagePromise = fetchData();
	}

	function retry(reset?: () => void) {
		reset?.();
		load();
	}

	function errorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Data kesiswaan belum dapat dimuat.';
	}

	function handleRenderError(error: unknown) {
		console.error('Kesiswaan render failed', error);
	}

	function openProfile(student: StudentRow) {
		if (!canManage) return;
		editingId = student.id;
		dialogKind = 'profile';
		selectedPhotoFile = null;
		profileForm = {
			nik: student.nik,
			tempat_lahir: student.tempat_lahir,
			tanggal_lahir: dateValue(student.tanggal_lahir),
			alamat: student.alamat,
			agama: student.agama,
			anak_ke: student.anak_ke ?? null,
			phone: student.phone,
			parent_name: student.parent_name,
			parent_phone: student.parent_phone,
		};
		dialogOpen = true;
	}

	function openCreate(kind: 'category' | 'violation' | 'achievement' | 'extracurricular' | 'member' | 'counseling' | 'transfer') {
		if (!canManage) return;
		editingId = null;
		dialogKind = kind;
		if (kind === 'category') categoryForm = emptyCategoryForm();
		if (kind === 'violation') violationForm = emptyViolationForm();
		if (kind === 'achievement') achievementForm = emptyAchievementForm();
		if (kind === 'extracurricular') extracurricularForm = emptyExtracurricularForm();
		if (kind === 'member') memberForm = emptyExtracurricularMemberForm();
		if (kind === 'counseling') counselingForm = emptyCounselingForm();
		if (kind === 'transfer') transferForm = emptyTransferForm();
		dialogOpen = true;
	}

	function openEditCategory(category: CategoryRow) {
		if (!canManage) return;
		editingId = category.id;
		dialogKind = 'category';
		categoryForm = {
			code: category.code,
			name: category.name,
			point: category.point,
			severity: category.severity,
			description: category.description,
			is_active: category.is_active,
		};
		dialogOpen = true;
	}

	function openEditViolation(item: ViolationRow) {
		if (!canManage) return;
		editingId = item.id;
		dialogKind = 'violation';
		violationForm = {
			student_id: item.student_id,
			category_id: item.category_id ?? '',
			incident_date: dateValue(item.incident_date),
			points: item.points,
			description: item.description,
			action_taken: item.action_taken,
			status: item.status,
		};
		dialogOpen = true;
	}

	function openEditAchievement(item: AchievementRow) {
		if (!canManage) return;
		editingId = item.id;
		dialogKind = 'achievement';
		achievementForm = {
			student_id: item.student_id,
			achievement_date: dateValue(item.achievement_date),
			title: item.title,
			level: item.level,
			category: item.category,
			organizer: item.organizer,
			description: item.description,
			document_url: item.document_url,
		};
		dialogOpen = true;
	}

	function openEditExtracurricular(item: ExtracurricularRow) {
		if (!canManage) return;
		editingId = item.id;
		dialogKind = 'extracurricular';
		extracurricularForm = {
			code: item.code,
			name: item.name,
			category: item.category,
			description: item.description,
			supervisor_employee_id: item.supervisor_employee_id ?? '',
			schedule_text: item.schedule_text,
			is_active: item.is_active,
		};
		dialogOpen = true;
	}

	function openEditMember(item: ExtracurricularMemberRow) {
		if (!canManage) return;
		editingId = item.id;
		dialogKind = 'member';
		memberForm = {
			extracurricular_id: item.extracurricular_id,
			student_id: item.student_id,
			joined_at: dateValue(item.joined_at),
			role: item.role,
			status: item.status,
			notes: item.notes,
		};
		dialogOpen = true;
	}

	function openEditCounseling(item: CounselingRow) {
		if (!canManage) return;
		editingId = item.id;
		dialogKind = 'counseling';
		counselingForm = {
			student_id: item.student_id,
			session_date: dateValue(item.session_date),
			topic: item.topic,
			summary: item.summary,
			follow_up: item.follow_up,
			status: item.status,
			is_confidential: item.is_confidential,
		};
		dialogOpen = true;
	}

	function applyCategoryPoint(categoryId: string) {
		violationForm.category_id = categoryId;
		const category = snapshot.categories.find((item) => item.id === categoryId);
		if (category) violationForm.points = category.point;
	}

	function handlePhotoChange(event: Event) {
		const input = event.currentTarget as HTMLInputElement;
		selectedPhotoFile = input.files?.[0] ?? null;
	}

	async function saveDialog() {
		if (!dialogKind) return;
		busy = true;
		try {
			let ok = false;
			if (dialogKind === 'profile' && editingId) ok = await saveProfile(editingId);
			if (dialogKind === 'category') ok = await submitJson('/api/kesiswaan/violation-categories', categoryForm);
			if (dialogKind === 'violation') ok = await submitJson('/api/kesiswaan/violations', violationForm);
			if (dialogKind === 'achievement') ok = await submitJson('/api/kesiswaan/achievements', achievementForm);
			if (dialogKind === 'extracurricular') ok = await submitJson('/api/kesiswaan/extracurriculars', extracurricularForm);
			if (dialogKind === 'member') ok = await submitJson('/api/kesiswaan/extracurricular-members', memberForm);
			if (dialogKind === 'counseling') ok = await submitJson('/api/kesiswaan/counseling-sessions', counselingForm);
			if (dialogKind === 'transfer') ok = await submitJson('/api/kesiswaan/student-transfers', transferForm, 'POST', false);
			if (!ok) return;
			dialogOpen = false;
			toast.success('Data kesiswaan disimpan');
			load();
		} finally {
			busy = false;
		}
	}

	async function saveProfile(id: string) {
		const ok = await submitJson(`/api/kesiswaan/students/${id}/profile`, profileForm, 'PUT', false);
		if (!ok) return false;
		if (selectedPhotoFile) {
			const form = new FormData();
			form.set('file', selectedPhotoFile);
			const res = await fetch(`/api/kesiswaan/students/${id}/photo`, { method: 'POST', body: form });
			if (!res.ok) {
				const j = await res.json().catch(() => ({ error: 'Gagal mengunggah foto siswa' })) as { error?: string };
				toast.error(j.error ?? 'Gagal mengunggah foto siswa');
				return false;
			}
		}
		return true;
	}

	async function submitJson(
		path: string,
		payload: ProfileForm | CategoryForm | ViolationForm | AchievementForm | ExtracurricularForm | ExtracurricularMemberForm | CounselingForm | TransferForm,
		method = editingId ? 'PUT' : 'POST',
		appendId = true
	) {
		const url = editingId && appendId ? `${path}/${editingId}` : path;
		const res = await fetch(url, {
			method,
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(payload),
		});
		if (!res.ok) {
			const j = await res.json().catch(() => ({ error: 'Gagal menyimpan data' })) as { error?: string };
			toast.error(j.error ?? 'Gagal menyimpan data');
			return false;
		}
		return true;
	}

	async function deleteEntity(kind: 'violation-categories' | 'violations' | 'achievements' | 'extracurriculars' | 'extracurricular-members' | 'counseling-sessions', id: string) {
		if (!canManage) return;
		if (!(await confirmAction({
			title: 'Hapus Data Kesiswaan',
			message: 'Hapus data kesiswaan ini?',
			confirmLabel: 'Hapus Data',
			tone: 'danger'
		}))) return;
		const res = await fetch(`/api/kesiswaan/${kind}/${id}`, { method: 'DELETE' });
		if (!res.ok) {
			const j = await res.json().catch(() => ({ error: 'Gagal menghapus data' })) as { error?: string };
			toast.error(j.error ?? 'Gagal menghapus data');
			return;
		}
		toast.success('Data kesiswaan dihapus');
		load();
	}

	function labelOf<T extends readonly (readonly [string, string])[]>(items: T, value: string) {
		return items.find(([key]) => key === value)?.[1] ?? value;
	}

	function dateValue(value?: string | null) {
		if (!value) return '';
		return value.slice(0, 10);
	}

	function formatDate(value?: string | null) {
		if (!value) return '-';
		return new Date(value).toLocaleDateString('id-ID', { timeZone: 'Asia/Makassar', day: '2-digit', month: 'short', year: 'numeric' });
	}

	function openURL(url: string) {
		const trimmed = url.trim();
		if (!trimmed) return;
		window.open(trimmed, '_blank', 'noopener,noreferrer');
	}

	onMount(() => {
		void load();
	});
</script>

<svelte:head><title>Kesiswaan - MTSN 2 Kolaka Utara</title></svelte:head>

<div class="space-y-6">
	<div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
		<div>
			<h1 class="text-lg font-semibold text-slate-800">Kesiswaan</h1>
			<p class="text-sm text-slate-500">Profil siswa, poin pelanggaran, prestasi, dan administrasi pembinaan.</p>
		</div>
		{#if canManage}
			<div class="flex flex-wrap gap-2">
				<Button size="sm" onclick={() => openCreate('violation')}>Tambah Pelanggaran</Button>
				<Button size="sm" variant="outline" onclick={() => openCreate('achievement')}>Tambah Prestasi</Button>
				<Button size="sm" variant="outline" onclick={() => openCreate('counseling')}>Catatan BK</Button>
				<Button size="sm" variant="outline" onclick={() => openCreate('member')}>Anggota Ekskul</Button>
				<Button size="sm" variant="outline" onclick={() => openCreate('transfer')}>Mutasi Siswa</Button>
				<Button size="sm" variant="outline" onclick={() => openCreate('category')}>Kategori Poin</Button>
			</div>
		{/if}
	</div>

	<AsyncContent promise={pagePromise} onerror={handleRenderError}>
		{#snippet pending()}
			<div class="grid grid-cols-2 gap-3 lg:grid-cols-4">
				{#each Array.from({ length: 7 }) as _, index (`stat-skeleton-${index}`)}
					<Card.Root class="border-slate-200">
						<Card.Content class="p-4"><Skeleton class="h-4 w-28" /><Skeleton class="mt-3 h-8 w-16" /></Card.Content>
					</Card.Root>
				{/each}
			</div>
		{/snippet}

		{#snippet failed(error, reset)}
			<RecoveryPanel title="Data Kesiswaan Belum Tersaji" message={errorMessage(error)} onRetry={() => retry(reset)} />
		{/snippet}

		{#snippet children(value)}
			{@const overview = value as KesiswaanData}
			<div class="grid grid-cols-2 gap-3 lg:grid-cols-4">
				{#each [
					{ label: 'Siswa Aktif', value: overview.stats.active_students },
					{ label: 'Calon Siswa', value: overview.stats.prospective_students },
					{ label: 'Siswa Berpoin', value: overview.stats.students_with_points },
					{ label: 'Total Poin', value: overview.stats.total_violation_points },
					{ label: 'Pelanggaran Aktif', value: overview.stats.open_violations },
					{ label: 'Prestasi Tahun Ini', value: overview.stats.achievements_this_year },
					{ label: 'Kategori Aktif', value: overview.stats.active_categories },
					{ label: 'Ekskul Aktif', value: overview.stats.active_extracurriculars },
					{ label: 'BK Terbuka', value: overview.stats.open_counseling_sessions },
					{ label: 'Mutasi Tahun Ini', value: overview.stats.transfers_this_year },
				] as item (item.label)}
					<Card.Root class="border-slate-200">
						<Card.Content class="p-4">
							<p class="text-xs text-slate-500">{item.label}</p>
							<p class="mt-1 text-2xl font-bold text-emerald-800">{item.value}</p>
						</Card.Content>
					</Card.Root>
				{/each}
			</div>
		{/snippet}
	</AsyncContent>

	<Tabs.Root bind:value={activeTab}>
		<Tabs.List class="w-full overflow-x-auto">
			<Tabs.Trigger value="siswa">Siswa</Tabs.Trigger>
			<Tabs.Trigger value="pelanggaran">Pelanggaran</Tabs.Trigger>
			<Tabs.Trigger value="prestasi">Prestasi</Tabs.Trigger>
			<Tabs.Trigger value="kategori">Kategori Poin</Tabs.Trigger>
			<Tabs.Trigger value="ekskul">Ekskul</Tabs.Trigger>
			<Tabs.Trigger value="bk">BK</Tabs.Trigger>
			<Tabs.Trigger value="mutasi">Mutasi</Tabs.Trigger>
		</Tabs.List>

		<AsyncContent promise={pagePromise} onerror={handleRenderError}>
			{#snippet pending()}
				<Card.Root class="border-slate-200"><Card.Content class="space-y-3 p-4">{#each Array.from({ length: 6 }) as _, index (`table-skeleton-${index}`)}<Skeleton class="h-10 w-full" />{/each}</Card.Content></Card.Root>
			{/snippet}

			{#snippet failed(error, reset)}
				<RecoveryPanel title="Data Kesiswaan Belum Tersaji" message={errorMessage(error)} onRetry={() => retry(reset)} />
			{/snippet}

			{#snippet children(value)}
				{@const overview = value as KesiswaanData}
				<Tabs.Content value="siswa" class="space-y-4">
					<Input placeholder="Cari nama, NIS, NISN, NIK, atau kelas" bind:value={studentSearch} class="sm:max-w-sm" />
					<Card.Root class="border-slate-200">
						<Card.Content class="p-0">
							{#if overview.students.length === 0 || filteredStudents.length === 0}
								<div class="p-4"><EmptyStatePanel compact title="Belum ada data siswa" description="Data siswa mengikuti master akademik. Lengkapi profil kesiswaan ketika siswa sudah tersedia." /></div>
							{:else}
								<Table.Root>
									<Table.Header><Table.Row><Table.Head>Siswa</Table.Head><Table.Head>Profil</Table.Head><Table.Head>Poin</Table.Head><Table.Head>Aktivitas</Table.Head>{#if canManage}<Table.Head class="w-24">Aksi</Table.Head>{/if}</Table.Row></Table.Header>
									<Table.Body>
										{#each filteredStudents as student (student.id)}
											<Table.Row>
												<Table.Cell>
													<div class="flex items-center gap-3">
														{#if student.photo_url}
															<img src={student.photo_url} alt={student.nama} class="h-10 w-10 rounded-md border border-slate-200 object-cover" />
														{:else}
															<div class="flex h-10 w-10 items-center justify-center rounded-md border border-slate-200 bg-emerald-50 text-xs font-semibold text-emerald-800">{student.nama.slice(0, 2).toUpperCase()}</div>
														{/if}
														<div>
															<p class="text-sm font-medium text-slate-800">{student.nama}</p>
															<p class="text-xs text-slate-500">{student.nis} · {student.class_code || student.class_name || 'Tanpa kelas'}</p>
														</div>
													</div>
												</Table.Cell>
												<Table.Cell class="text-sm">
													<p>{student.nik || 'NIK belum diisi'}</p>
													<p class="text-xs text-slate-500">{student.tempat_lahir || '-'}{student.tanggal_lahir ? `, ${formatDate(student.tanggal_lahir)}` : ''}</p>
												</Table.Cell>
												<Table.Cell><Badge variant={student.total_violation_points > 0 ? 'destructive' : 'outline'}>{student.total_violation_points} poin</Badge></Table.Cell>
												<Table.Cell class="text-sm">{student.violation_count} pelanggaran · {student.achievement_count} prestasi</Table.Cell>
												{#if canManage}
													<Table.Cell><Button size="sm" variant="outline" onclick={() => openProfile(student)}>Edit</Button></Table.Cell>
												{/if}
											</Table.Row>
										{/each}
									</Table.Body>
								</Table.Root>
							{/if}
						</Card.Content>
					</Card.Root>
				</Tabs.Content>

				<Tabs.Content value="pelanggaran" class="space-y-4">
					<div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
						<Input placeholder="Cari siswa, kategori, tindakan, atau catatan" bind:value={violationSearch} class="sm:max-w-sm" />
						{#if canManage}<Button size="sm" onclick={() => openCreate('violation')}>Tambah Pelanggaran</Button>{/if}
					</div>
					<Card.Root class="border-slate-200">
						<Card.Content class="p-0">
							{#if filteredViolations.length === 0}
								<div class="p-4"><EmptyStatePanel compact title="Belum ada catatan pelanggaran" description="Catatan pelanggaran akan otomatis memperbarui total poin siswa." /></div>
							{:else}
								<Table.Root>
									<Table.Header><Table.Row><Table.Head>Siswa</Table.Head><Table.Head>Kategori</Table.Head><Table.Head>Tanggal</Table.Head><Table.Head>Status</Table.Head>{#if canManage}<Table.Head class="w-32">Aksi</Table.Head>{/if}</Table.Row></Table.Header>
									<Table.Body>
										{#each filteredViolations as item (item.id)}
											<Table.Row>
												<Table.Cell><p class="text-sm font-medium">{item.student_name}</p><p class="text-xs text-slate-500">{item.student_nis} · {item.class_code || item.class_name || '-'}</p></Table.Cell>
												<Table.Cell><p class="text-sm">{item.category_name || 'Tanpa kategori'} · {item.points} poin</p><p class="text-xs text-slate-500">{item.description || '-'}</p></Table.Cell>
												<Table.Cell class="text-sm">{formatDate(item.incident_date)}</Table.Cell>
												<Table.Cell><Badge variant={item.status === 'open' ? 'destructive' : item.status === 'resolved' ? 'default' : 'outline'}>{labelOf(VIOLATION_STATUSES, item.status)}</Badge></Table.Cell>
												{#if canManage}
													<Table.Cell>
														<div class="flex gap-1"><Button size="sm" variant="outline" onclick={() => openEditViolation(item)}>Edit</Button><Button size="sm" variant="destructive" onclick={() => deleteEntity('violations', item.id)}>Hapus</Button></div>
													</Table.Cell>
												{/if}
											</Table.Row>
										{/each}
									</Table.Body>
								</Table.Root>
							{/if}
						</Card.Content>
					</Card.Root>
				</Tabs.Content>

				<Tabs.Content value="prestasi" class="space-y-4">
					<div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
						<Input placeholder="Cari siswa, prestasi, kategori, atau penyelenggara" bind:value={achievementSearch} class="sm:max-w-sm" />
						{#if canManage}<Button size="sm" onclick={() => openCreate('achievement')}>Tambah Prestasi</Button>{/if}
					</div>
					<Card.Root class="border-slate-200">
						<Card.Content class="p-0">
							{#if filteredAchievements.length === 0}
								<div class="p-4"><EmptyStatePanel compact title="Belum ada prestasi siswa" description="Catat prestasi siswa dari tingkat sekolah sampai internasional." /></div>
							{:else}
								<Table.Root>
									<Table.Header><Table.Row><Table.Head>Prestasi</Table.Head><Table.Head>Siswa</Table.Head><Table.Head>Tingkat</Table.Head><Table.Head>Dokumen</Table.Head>{#if canManage}<Table.Head class="w-32">Aksi</Table.Head>{/if}</Table.Row></Table.Header>
									<Table.Body>
										{#each filteredAchievements as item (item.id)}
											<Table.Row>
												<Table.Cell><p class="text-sm font-medium">{item.title}</p><p class="text-xs text-slate-500">{formatDate(item.achievement_date)} · {item.category || '-'}</p></Table.Cell>
												<Table.Cell><p class="text-sm">{item.student_name}</p><p class="text-xs text-slate-500">{item.student_nis} · {item.class_code || item.class_name || '-'}</p></Table.Cell>
												<Table.Cell><Badge variant={item.level === 'national' || item.level === 'international' ? 'default' : 'outline'}>{labelOf(ACHIEVEMENT_LEVELS, item.level)}</Badge></Table.Cell>
												<Table.Cell>{#if item.document_url}<Button size="sm" variant="outline" onclick={() => openURL(item.document_url)}>Buka</Button>{:else}<span class="text-sm text-slate-400">-</span>{/if}</Table.Cell>
												{#if canManage}
													<Table.Cell><div class="flex gap-1"><Button size="sm" variant="outline" onclick={() => openEditAchievement(item)}>Edit</Button><Button size="sm" variant="destructive" onclick={() => deleteEntity('achievements', item.id)}>Hapus</Button></div></Table.Cell>
												{/if}
											</Table.Row>
										{/each}
									</Table.Body>
								</Table.Root>
							{/if}
						</Card.Content>
					</Card.Root>
				</Tabs.Content>

				<Tabs.Content value="kategori" class="space-y-4">
					<div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
						<Input placeholder="Cari kode, kategori, atau deskripsi" bind:value={categorySearch} class="sm:max-w-sm" />
						{#if canManage}<Button size="sm" onclick={() => openCreate('category')}>Tambah Kategori</Button>{/if}
					</div>
					<Card.Root class="border-slate-200">
						<Card.Content class="p-0">
							<Table.Root>
								<Table.Header><Table.Row><Table.Head>Kategori</Table.Head><Table.Head>Poin</Table.Head><Table.Head>Tingkat</Table.Head><Table.Head>Status</Table.Head>{#if canManage}<Table.Head class="w-32">Aksi</Table.Head>{/if}</Table.Row></Table.Header>
								<Table.Body>
									{#each filteredCategories as item (item.id)}
										<Table.Row>
											<Table.Cell><p class="text-sm font-medium">{item.code} · {item.name}</p><p class="text-xs text-slate-500">{item.description || '-'}</p></Table.Cell>
											<Table.Cell>{item.point}</Table.Cell>
											<Table.Cell><Badge variant={item.severity === 'berat' ? 'destructive' : item.severity === 'sedang' ? 'default' : 'outline'}>{labelOf(SEVERITIES, item.severity)}</Badge></Table.Cell>
											<Table.Cell><Badge variant={item.is_active ? 'outline' : 'destructive'}>{item.is_active ? 'Aktif' : 'Nonaktif'}</Badge></Table.Cell>
											{#if canManage}<Table.Cell><div class="flex gap-1"><Button size="sm" variant="outline" onclick={() => openEditCategory(item)}>Edit</Button><Button size="sm" variant="destructive" onclick={() => deleteEntity('violation-categories', item.id)}>Hapus</Button></div></Table.Cell>{/if}
										</Table.Row>
									{:else}
										<Table.Row><Table.Cell colspan={canManage ? 5 : 4}><EmptyStatePanel compact title="Belum ada kategori poin" description="Tambahkan kategori agar pelanggaran memiliki standar poin yang konsisten." /></Table.Cell></Table.Row>
									{/each}
								</Table.Body>
							</Table.Root>
						</Card.Content>
					</Card.Root>
				</Tabs.Content>

				<Tabs.Content value="ekskul" class="space-y-4">
					<div class="grid gap-4 xl:grid-cols-[minmax(0,1fr)_minmax(0,1.25fr)]">
						<Card.Root class="border-slate-200">
							<Card.Header>
								<div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
									<div>
										<Card.Title>Program Ekstrakurikuler</Card.Title>
										<Card.Description>Daftar ekskul, pembina, jadwal, dan jumlah anggota aktif.</Card.Description>
									</div>
									{#if canManage}<Button size="sm" onclick={() => openCreate('extracurricular')}>Tambah Ekskul</Button>{/if}
								</div>
								<Input placeholder="Cari ekskul, kategori, atau jadwal" bind:value={extracurricularSearch} />
							</Card.Header>
							<Card.Content class="p-0">
								{#if filteredExtracurriculars.length === 0}
									<div class="p-4"><EmptyStatePanel compact title="Belum ada ekskul" description="Daftarkan program ekstrakurikuler sebelum mencatat keanggotaan siswa." /></div>
								{:else}
									<Table.Root>
										<Table.Header><Table.Row><Table.Head>Ekskul</Table.Head><Table.Head>Anggota</Table.Head><Table.Head>Status</Table.Head>{#if canManage}<Table.Head class="w-32">Aksi</Table.Head>{/if}</Table.Row></Table.Header>
										<Table.Body>
											{#each filteredExtracurriculars as item (item.id)}
												<Table.Row>
													<Table.Cell><p class="text-sm font-medium">{item.code} · {item.name}</p><p class="text-xs text-slate-500">{item.category || '-'} · {item.schedule_text || 'Jadwal belum diisi'}</p></Table.Cell>
													<Table.Cell><Badge variant="outline">{item.active_member_count} aktif</Badge></Table.Cell>
													<Table.Cell><Badge variant={item.is_active ? 'outline' : 'destructive'}>{item.is_active ? 'Aktif' : 'Nonaktif'}</Badge></Table.Cell>
													{#if canManage}
														<Table.Cell><div class="flex gap-1"><Button size="sm" variant="outline" onclick={() => openEditExtracurricular(item)}>Edit</Button><Button size="sm" variant="destructive" onclick={() => deleteEntity('extracurriculars', item.id)}>Hapus</Button></div></Table.Cell>
													{/if}
												</Table.Row>
											{/each}
										</Table.Body>
									</Table.Root>
								{/if}
							</Card.Content>
						</Card.Root>

						<Card.Root class="border-slate-200">
							<Card.Header>
								<div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
									<div>
										<Card.Title>Keanggotaan Ekskul</Card.Title>
										<Card.Description>Riwayat siswa sebagai anggota, pengurus, atau ketua ekskul.</Card.Description>
									</div>
									{#if canManage}<Button size="sm" onclick={() => openCreate('member')}>Tambah Anggota</Button>{/if}
								</div>
								<Input placeholder="Cari siswa atau ekskul" bind:value={memberSearch} />
							</Card.Header>
							<Card.Content class="p-0">
								{#if filteredMembers.length === 0}
									<div class="p-4"><EmptyStatePanel compact title="Belum ada anggota ekskul" description="Catat keikutsertaan siswa agar pembinaan minat bakat lebih terukur." /></div>
								{:else}
									<Table.Root>
										<Table.Header><Table.Row><Table.Head>Siswa</Table.Head><Table.Head>Ekskul</Table.Head><Table.Head>Peran</Table.Head><Table.Head>Status</Table.Head>{#if canManage}<Table.Head class="w-32">Aksi</Table.Head>{/if}</Table.Row></Table.Header>
										<Table.Body>
											{#each filteredMembers as item (item.id)}
												<Table.Row>
													<Table.Cell><p class="text-sm font-medium">{item.student_name}</p><p class="text-xs text-slate-500">{item.student_nis} · {item.class_code || item.class_name || '-'}</p></Table.Cell>
													<Table.Cell><p class="text-sm">{item.extracurricular_name}</p><p class="text-xs text-slate-500">{formatDate(item.joined_at)}</p></Table.Cell>
													<Table.Cell><Badge variant="outline">{labelOf(EXTRACURRICULAR_ROLES, item.role)}</Badge></Table.Cell>
													<Table.Cell><Badge variant={item.status === 'active' ? 'default' : 'outline'}>{labelOf(EXTRACURRICULAR_STATUSES, item.status)}</Badge></Table.Cell>
													{#if canManage}
														<Table.Cell><div class="flex gap-1"><Button size="sm" variant="outline" onclick={() => openEditMember(item)}>Edit</Button><Button size="sm" variant="destructive" onclick={() => deleteEntity('extracurricular-members', item.id)}>Hapus</Button></div></Table.Cell>
													{/if}
												</Table.Row>
											{/each}
										</Table.Body>
									</Table.Root>
								{/if}
							</Card.Content>
						</Card.Root>
					</div>
				</Tabs.Content>

				<Tabs.Content value="bk" class="space-y-4">
					<div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
						<Input placeholder="Cari siswa, topik, ringkasan, atau tindak lanjut" bind:value={counselingSearch} class="sm:max-w-sm" />
						{#if canManage}<Button size="sm" onclick={() => openCreate('counseling')}>Tambah Catatan BK</Button>{/if}
					</div>
					<Card.Root class="border-slate-200">
						<Card.Content class="p-0">
							{#if filteredCounseling.length === 0}
								<div class="p-4"><EmptyStatePanel compact title="Belum ada catatan BK" description="Catatan rahasia hanya tampil untuk admin dan petugas kesiswaan." /></div>
							{:else}
								<Table.Root>
									<Table.Header><Table.Row><Table.Head>Siswa</Table.Head><Table.Head>Topik</Table.Head><Table.Head>Status</Table.Head><Table.Head>Kerahasiaan</Table.Head>{#if canManage}<Table.Head class="w-32">Aksi</Table.Head>{/if}</Table.Row></Table.Header>
									<Table.Body>
										{#each filteredCounseling as item (item.id)}
											<Table.Row>
												<Table.Cell><p class="text-sm font-medium">{item.student_name}</p><p class="text-xs text-slate-500">{item.student_nis} · {item.class_code || item.class_name || '-'}</p></Table.Cell>
												<Table.Cell><p class="text-sm font-medium">{item.topic}</p><p class="text-xs text-slate-500">{formatDate(item.session_date)} · {item.follow_up || item.summary || '-'}</p></Table.Cell>
												<Table.Cell><Badge variant={item.status === 'open' || item.status === 'monitoring' ? 'default' : 'outline'}>{labelOf(COUNSELING_STATUSES, item.status)}</Badge></Table.Cell>
												<Table.Cell><Badge variant={item.is_confidential ? 'destructive' : 'outline'}>{item.is_confidential ? 'Rahasia' : 'Umum'}</Badge></Table.Cell>
												{#if canManage}
													<Table.Cell><div class="flex gap-1"><Button size="sm" variant="outline" onclick={() => openEditCounseling(item)}>Edit</Button><Button size="sm" variant="destructive" onclick={() => deleteEntity('counseling-sessions', item.id)}>Hapus</Button></div></Table.Cell>
												{/if}
											</Table.Row>
										{/each}
									</Table.Body>
								</Table.Root>
							{/if}
						</Card.Content>
					</Card.Root>
				</Tabs.Content>

				<Tabs.Content value="mutasi" class="space-y-4">
					<div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
						<Input placeholder="Cari siswa, sekolah asal/tujuan, alasan, atau dokumen" bind:value={transferSearch} class="sm:max-w-sm" />
						{#if canManage}<Button size="sm" onclick={() => openCreate('transfer')}>Catat Mutasi</Button>{/if}
					</div>
					<Card.Root class="border-slate-200">
						<Card.Content class="p-0">
							{#if filteredTransfers.length === 0}
								<div class="p-4"><EmptyStatePanel compact title="Belum ada riwayat mutasi" description="Mutasi keluar otomatis mengubah status siswa menjadi mutasi, sedangkan mutasi masuk mengaktifkan siswa." /></div>
							{:else}
								<Table.Root>
									<Table.Header><Table.Row><Table.Head>Siswa</Table.Head><Table.Head>Jenis</Table.Head><Table.Head>Sekolah</Table.Head><Table.Head>Dokumen</Table.Head><Table.Head>Status Siswa</Table.Head></Table.Row></Table.Header>
									<Table.Body>
										{#each filteredTransfers as item (item.id)}
											<Table.Row>
												<Table.Cell><p class="text-sm font-medium">{item.student_name}</p><p class="text-xs text-slate-500">{item.student_nis} · {item.class_code || item.class_name || '-'}</p></Table.Cell>
												<Table.Cell><Badge variant={item.transfer_type === 'out' ? 'destructive' : 'default'}>{labelOf(TRANSFER_TYPES, item.transfer_type)}</Badge><p class="mt-1 text-xs text-slate-500">{formatDate(item.transfer_date)}</p></Table.Cell>
												<Table.Cell class="text-sm">{item.transfer_type === 'out' ? item.destination_school : item.previous_school}</Table.Cell>
												<Table.Cell class="text-sm">{item.document_ref || '-'}</Table.Cell>
												<Table.Cell><Badge variant="outline">{labelOf(STUDENT_STATUSES, item.student_status)}</Badge></Table.Cell>
											</Table.Row>
										{/each}
									</Table.Body>
								</Table.Root>
							{/if}
						</Card.Content>
					</Card.Root>
				</Tabs.Content>
			{/snippet}
		</AsyncContent>
	</Tabs.Root>
</div>

<Dialog.Root bind:open={dialogOpen}>
	<Dialog.Content>
		<Dialog.Header>
			<Dialog.Title>{editingId ? 'Edit Data Kesiswaan' : 'Tambah Data Kesiswaan'}</Dialog.Title>
			<Dialog.Description>Lengkapi data sesuai administrasi kesiswaan madrasah.</Dialog.Description>
		</Dialog.Header>

		<div class="mt-4 space-y-4">
			{#if dialogKind === 'profile'}
				<div class="grid gap-3 sm:grid-cols-2">
					<div><label for="profile-nik" class="text-sm font-medium">NIK</label><Input id="profile-nik" bind:value={profileForm.nik} /></div>
					<div><label for="profile-phone" class="text-sm font-medium">No. Telepon Siswa</label><Input id="profile-phone" bind:value={profileForm.phone} /></div>
					<div><label for="profile-place" class="text-sm font-medium">Tempat Lahir</label><Input id="profile-place" bind:value={profileForm.tempat_lahir} /></div>
					<div><label for="profile-date" class="text-sm font-medium">Tanggal Lahir</label><Input id="profile-date" type="date" bind:value={profileForm.tanggal_lahir} /></div>
					<div><label for="profile-religion" class="text-sm font-medium">Agama</label><Input id="profile-religion" bind:value={profileForm.agama} /></div>
					<div><label for="profile-child-no" class="text-sm font-medium">Anak Ke</label><Input id="profile-child-no" type="number" min="1" bind:value={profileForm.anak_ke} /></div>
					<div><label for="profile-parent" class="text-sm font-medium">Nama Orang Tua/Wali</label><Input id="profile-parent" bind:value={profileForm.parent_name} /></div>
					<div><label for="profile-parent-phone" class="text-sm font-medium">Telepon Orang Tua/Wali</label><Input id="profile-parent-phone" bind:value={profileForm.parent_phone} /></div>
					<div class="sm:col-span-2"><label for="profile-photo" class="text-sm font-medium">Foto Siswa</label><input id="profile-photo" type="file" accept="image/*" onchange={handlePhotoChange} class="block w-full rounded-md border border-input bg-background px-3 py-2 text-sm" /></div>
				</div>
				<div><label for="profile-address" class="text-sm font-medium">Alamat</label><Textarea id="profile-address" bind:value={profileForm.alamat} /></div>
			{:else if dialogKind === 'category'}
				<div class="grid gap-3 sm:grid-cols-2">
					<div><label for="cat-code" class="text-sm font-medium">Kode</label><Input id="cat-code" bind:value={categoryForm.code} /></div>
					<div><label for="cat-name" class="text-sm font-medium">Nama Kategori</label><Input id="cat-name" bind:value={categoryForm.name} /></div>
					<div><label for="cat-point" class="text-sm font-medium">Poin</label><Input id="cat-point" type="number" min="0" bind:value={categoryForm.point} /></div>
					<div>
						<label for="cat-severity" class="text-sm font-medium">Tingkat</label>
						<select id="cat-severity" bind:value={categoryForm.severity} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">{#each SEVERITIES as [value, label] (value)}<option {value}>{label}</option>{/each}</select>
					</div>
				</div>
				<div><label for="cat-desc" class="text-sm font-medium">Deskripsi</label><Textarea id="cat-desc" bind:value={categoryForm.description} /></div>
			{:else if dialogKind === 'violation'}
				<div class="grid gap-3 sm:grid-cols-2">
					<div>
						<label for="vio-student" class="text-sm font-medium">Siswa</label>
						<select id="vio-student" bind:value={violationForm.student_id} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"><option value="">Pilih siswa</option>{#each snapshot.students as student (student.id)}<option value={student.id}>{student.nama} · {student.nis}</option>{/each}</select>
					</div>
					<div>
						<label for="vio-category" class="text-sm font-medium">Kategori</label>
						<select id="vio-category" value={violationForm.category_id} onchange={(event) => applyCategoryPoint((event.currentTarget as HTMLSelectElement).value)} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"><option value="">Tanpa kategori</option>{#each snapshot.categories as category (category.id)}<option value={category.id}>{category.code} · {category.name}</option>{/each}</select>
					</div>
					<div><label for="vio-date" class="text-sm font-medium">Tanggal Kejadian</label><Input id="vio-date" type="date" bind:value={violationForm.incident_date} /></div>
					<div><label for="vio-points" class="text-sm font-medium">Poin</label><Input id="vio-points" type="number" min="0" bind:value={violationForm.points} /></div>
					<div>
						<label for="vio-status" class="text-sm font-medium">Status</label>
						<select id="vio-status" bind:value={violationForm.status} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">{#each VIOLATION_STATUSES as [value, label] (value)}<option {value}>{label}</option>{/each}</select>
					</div>
				</div>
				<div><label for="vio-desc" class="text-sm font-medium">Catatan Kejadian</label><Textarea id="vio-desc" bind:value={violationForm.description} /></div>
				<div><label for="vio-action" class="text-sm font-medium">Tindakan Pembinaan</label><Textarea id="vio-action" bind:value={violationForm.action_taken} /></div>
			{:else if dialogKind === 'achievement'}
				<div class="grid gap-3 sm:grid-cols-2">
					<div>
						<label for="ach-student" class="text-sm font-medium">Siswa</label>
						<select id="ach-student" bind:value={achievementForm.student_id} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"><option value="">Pilih siswa</option>{#each snapshot.students as student (student.id)}<option value={student.id}>{student.nama} · {student.nis}</option>{/each}</select>
					</div>
					<div><label for="ach-date" class="text-sm font-medium">Tanggal Prestasi</label><Input id="ach-date" type="date" bind:value={achievementForm.achievement_date} /></div>
					<div class="sm:col-span-2"><label for="ach-title" class="text-sm font-medium">Judul Prestasi</label><Input id="ach-title" bind:value={achievementForm.title} /></div>
					<div>
						<label for="ach-level" class="text-sm font-medium">Tingkat</label>
						<select id="ach-level" bind:value={achievementForm.level} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">{#each ACHIEVEMENT_LEVELS as [value, label] (value)}<option {value}>{label}</option>{/each}</select>
					</div>
					<div><label for="ach-category" class="text-sm font-medium">Kategori</label><Input id="ach-category" bind:value={achievementForm.category} /></div>
					<div><label for="ach-organizer" class="text-sm font-medium">Penyelenggara</label><Input id="ach-organizer" bind:value={achievementForm.organizer} /></div>
					<div><label for="ach-doc" class="text-sm font-medium">URL/Lokasi Dokumen</label><Input id="ach-doc" bind:value={achievementForm.document_url} /></div>
				</div>
				<div><label for="ach-desc" class="text-sm font-medium">Deskripsi</label><Textarea id="ach-desc" bind:value={achievementForm.description} /></div>
			{:else if dialogKind === 'extracurricular'}
				<div class="grid gap-3 sm:grid-cols-2">
					<div><label for="extra-code" class="text-sm font-medium">Kode Ekskul</label><Input id="extra-code" bind:value={extracurricularForm.code} /></div>
					<div><label for="extra-name" class="text-sm font-medium">Nama Ekskul</label><Input id="extra-name" bind:value={extracurricularForm.name} /></div>
					<div><label for="extra-category" class="text-sm font-medium">Kategori</label><Input id="extra-category" bind:value={extracurricularForm.category} placeholder="Olahraga, seni, keagamaan" /></div>
					<div><label for="extra-schedule" class="text-sm font-medium">Jadwal</label><Input id="extra-schedule" bind:value={extracurricularForm.schedule_text} placeholder="Jumat 15.30 WITA" /></div>
					<div class="sm:col-span-2"><label for="extra-supervisor" class="text-sm font-medium">ID Pembina Pegawai</label><Input id="extra-supervisor" bind:value={extracurricularForm.supervisor_employee_id} placeholder="Opsional" /></div>
					<label class="flex items-center gap-2 text-sm font-medium sm:col-span-2" for="extra-active">
						<input id="extra-active" type="checkbox" bind:checked={extracurricularForm.is_active} class="size-4 rounded border-input" />
						Ekskul aktif
					</label>
				</div>
				<div><label for="extra-desc" class="text-sm font-medium">Deskripsi</label><Textarea id="extra-desc" bind:value={extracurricularForm.description} /></div>
			{:else if dialogKind === 'member'}
				<div class="grid gap-3 sm:grid-cols-2">
					<div>
						<label for="member-extra" class="text-sm font-medium">Ekskul</label>
						<select id="member-extra" bind:value={memberForm.extracurricular_id} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"><option value="">Pilih ekskul</option>{#each snapshot.extracurriculars as item (item.id)}<option value={item.id}>{item.code} · {item.name}</option>{/each}</select>
					</div>
					<div>
						<label for="member-student" class="text-sm font-medium">Siswa</label>
						<select id="member-student" bind:value={memberForm.student_id} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"><option value="">Pilih siswa</option>{#each snapshot.students as student (student.id)}<option value={student.id}>{student.nama} · {student.nis}</option>{/each}</select>
					</div>
					<div><label for="member-date" class="text-sm font-medium">Tanggal Bergabung</label><Input id="member-date" type="date" bind:value={memberForm.joined_at} /></div>
					<div>
						<label for="member-role" class="text-sm font-medium">Peran</label>
						<select id="member-role" bind:value={memberForm.role} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">{#each EXTRACURRICULAR_ROLES as [value, label] (value)}<option {value}>{label}</option>{/each}</select>
					</div>
					<div>
						<label for="member-status" class="text-sm font-medium">Status</label>
						<select id="member-status" bind:value={memberForm.status} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">{#each EXTRACURRICULAR_STATUSES as [value, label] (value)}<option {value}>{label}</option>{/each}</select>
					</div>
				</div>
				<div><label for="member-notes" class="text-sm font-medium">Catatan</label><Textarea id="member-notes" bind:value={memberForm.notes} /></div>
			{:else if dialogKind === 'counseling'}
				<div class="grid gap-3 sm:grid-cols-2">
					<div>
						<label for="bk-student" class="text-sm font-medium">Siswa</label>
						<select id="bk-student" bind:value={counselingForm.student_id} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"><option value="">Pilih siswa</option>{#each snapshot.students as student (student.id)}<option value={student.id}>{student.nama} · {student.nis}</option>{/each}</select>
					</div>
					<div><label for="bk-date" class="text-sm font-medium">Tanggal Konseling</label><Input id="bk-date" type="date" bind:value={counselingForm.session_date} /></div>
					<div class="sm:col-span-2"><label for="bk-topic" class="text-sm font-medium">Topik</label><Input id="bk-topic" bind:value={counselingForm.topic} /></div>
					<div>
						<label for="bk-status" class="text-sm font-medium">Status</label>
						<select id="bk-status" bind:value={counselingForm.status} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">{#each COUNSELING_STATUSES as [value, label] (value)}<option {value}>{label}</option>{/each}</select>
					</div>
					<label class="flex items-center gap-2 text-sm font-medium" for="bk-confidential">
						<input id="bk-confidential" type="checkbox" bind:checked={counselingForm.is_confidential} class="size-4 rounded border-input" />
						Catatan rahasia
					</label>
				</div>
				<div><label for="bk-summary" class="text-sm font-medium">Ringkasan</label><Textarea id="bk-summary" bind:value={counselingForm.summary} /></div>
				<div><label for="bk-follow-up" class="text-sm font-medium">Tindak Lanjut</label><Textarea id="bk-follow-up" bind:value={counselingForm.follow_up} /></div>
			{:else if dialogKind === 'transfer'}
				<div class="grid gap-3 sm:grid-cols-2">
					<div>
						<label for="transfer-student" class="text-sm font-medium">Siswa</label>
						<select id="transfer-student" bind:value={transferForm.student_id} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"><option value="">Pilih siswa</option>{#each snapshot.students as student (student.id)}<option value={student.id}>{student.nama} · {student.nis}</option>{/each}</select>
					</div>
					<div><label for="transfer-date" class="text-sm font-medium">Tanggal Mutasi</label><Input id="transfer-date" type="date" bind:value={transferForm.transfer_date} /></div>
					<div>
						<label for="transfer-type" class="text-sm font-medium">Jenis Mutasi</label>
						<select id="transfer-type" bind:value={transferForm.transfer_type} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">{#each TRANSFER_TYPES as [value, label] (value)}<option {value}>{label}</option>{/each}</select>
					</div>
					<div><label for="transfer-doc" class="text-sm font-medium">Nomor/Dokumen Rujukan</label><Input id="transfer-doc" bind:value={transferForm.document_ref} /></div>
					<div><label for="transfer-prev" class="text-sm font-medium">Sekolah Asal</label><Input id="transfer-prev" bind:value={transferForm.previous_school} /></div>
					<div><label for="transfer-dest" class="text-sm font-medium">Sekolah Tujuan</label><Input id="transfer-dest" bind:value={transferForm.destination_school} /></div>
				</div>
				<div><label for="transfer-reason" class="text-sm font-medium">Alasan</label><Textarea id="transfer-reason" bind:value={transferForm.reason} /></div>
				<div><label for="transfer-notes" class="text-sm font-medium">Catatan</label><Textarea id="transfer-notes" bind:value={transferForm.notes} /></div>
			{/if}
		</div>

		<Dialog.Footer>
			<Button variant="outline" onclick={() => (dialogOpen = false)}>Batal</Button>
			<LoadingButton loading={busy} onclick={() => void saveDialog()}>Simpan</LoadingButton>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
