<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Badge } from '$lib/components/ui/badge';
	import { toast } from '$lib/components/ui/sonner';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import EmptyStatePanel from '$lib/components/EmptyStatePanel.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { confirmAction } from '$lib/confirm-dialog';
	import { clientApiPath, clientApiPathWithQuery, readClientApiData, readClientJson } from '$lib/client/api';

	type Assignment = {
		id: string;
		class_name: string;
		class_code: string;
		subject_name: string;
		subject_code: string;
		teacher_name: string;
	};

	type GradeComponent = {
		id: string;
		assignment_id: string;
		title: string;
		category: string;
		weight: number;
		max_score: number;
		is_published: boolean;
		source_non_test_assessment_id?: string | null;
		source_non_test_title?: string;
		source_non_test_type?: string;
		source_non_test_synced_at?: string | null;
		source_non_test_synced_by?: string;
	};

	type GradeSummary = {
		student_id: string;
		nis: string;
		nisn: string;
		nama: string;
		component_count: number;
		filled_count: number;
		final_score: number;
	};

	type GradeEntry = {
		student_id: string;
		nis: string;
		nisn: string;
		nama: string;
		entry_id?: string;
		score: number;
		notes?: string;
	};

	type GradeReadiness = {
		ready: boolean;
		published_component_count: number;
		draft_component_count: number;
		ready_student_count: number;
		incomplete_student_count: number;
		missing_grade_count: number;
	};

	type GradeFinalization = {
		assignment_id: string;
		finalized_by: string;
		notes: string;
		finalized_at: string;
	};

	type AssignmentStatus = {
		assignment_id: string;
		class_name: string;
		class_code: string;
		subject_name: string;
		subject_code: string;
		teacher_name: string;
		component_count: number;
		published_component_count: number;
		draft_component_count: number;
		student_count: number;
		ready_student_count: number;
		incomplete_student_count: number;
		missing_grade_count: number;
		ready: boolean;
		is_finalized: boolean;
		finalization?: GradeFinalization | null;
	};

	type ClassReadinessSummary = {
		class_name: string;
		class_code: string;
		assignment_count: number;
		ready_count: number;
		finalized_count: number;
		attention_count: number;
		missing_grade_count: number;
	};

	type TeacherReadinessSummary = {
		teacher_name: string;
		assignment_count: number;
		ready_count: number;
		finalized_count: number;
		attention_count: number;
		missing_grade_count: number;
	};

	type GradesPayload = {
		assignments?: Assignment[];
		assignment_statuses?: AssignmentStatus[];
		components?: GradeComponent[];
		summary?: GradeSummary[];
		entries?: GradeEntry[];
		readiness?: GradeReadiness | null;
		finalization?: GradeFinalization | null;
		error?: string;
		message?: string;
	};

	type GradesOverview = {
		assignments: Assignment[];
		assignmentStatuses: AssignmentStatus[];
		components: GradeComponent[];
		summary: GradeSummary[];
		entries: GradeEntry[];
		readiness: GradeReadiness | null;
		finalization: GradeFinalization | null;
	};

	type GradeWorkspace = 'triage' | 'gradebook' | 'finalization';

	const categoryOptions = [
		{ value: 'assignment', label: 'Tugas' },
		{ value: 'quiz', label: 'Kuis' },
		{ value: 'midterm', label: 'UTS' },
		{ value: 'final', label: 'UAS' },
		{ value: 'project', label: 'Proyek' },
		{ value: 'practice', label: 'Praktik' },
		{ value: 'attitude', label: 'Sikap' },
		{ value: 'attendance', label: 'Kehadiran' },
		{ value: 'other', label: 'Lainnya' }
	];

	let overviewPromise = $state<Promise<GradesOverview> | null>(null);
	let assignments = $state<Assignment[]>([]);
	let assignmentStatuses = $state<AssignmentStatus[]>([]);
	let components = $state<GradeComponent[]>([]);
	let summary = $state<GradeSummary[]>([]);
	let entries = $state<GradeEntry[]>([]);
	let readiness = $state<GradeReadiness | null>(null);
	let finalization = $state<GradeFinalization | null>(null);

	let assignmentId = $state('');
	let componentId = $state('');

	let createBusy = $state(false);
	let entryBusy = $state<Record<string, boolean>>({});
	let publishBusy = $state<Record<string, boolean>>({});
	let bulkSaveBusy = $state(false);
	let finalizationBusy = $state(false);
	let batchFinalizationBusy = $state(false);
	let batchReopenBusy = $state(false);
	let exportBusy = $state(false);
	let classExportBusy = $state(false);
	let teacherExportBusy = $state(false);
	let editingComponentId = $state('');
	let componentTitle = $state('');
	let componentCategory = $state('assignment');
	let componentWeight = $state(1);
	let componentMaxScore = $state(100);
	let quickFillScore = $state('');
	let quickFillNote = $state('');
	let finalizeNotes = $state('');
	let assignmentStatusFilter = $state<'all' | 'ready' | 'finalized' | 'attention'>('all');
	let assignmentStatusQuery = $state('');
	let assignmentTeacherFilter = $state('');
	let classFocusKey = $state('');
	let gradeWorkspace = $state<GradeWorkspace>('triage');
	let overviewRequestId = 0;

	let scoreInput = $state<Record<string, string>>({});
	let noteInput = $state<Record<string, string>>({});

	const selectedAssignment = $derived(assignments.find((item) => item.id === assignmentId) ?? null);
	const selectedAssignmentStatus = $derived(
		assignmentStatuses.find((item) => item.assignment_id === assignmentId) ?? null
	);
	const selectedComponent = $derived(components.find((item) => item.id === componentId) ?? null);
	const editingComponent = $derived(components.find((item) => item.id === editingComponentId) ?? null);
	const completionRate = $derived(
		readiness ? (summary.length === 0 ? 0 : Math.round((readiness.ready_student_count / summary.length) * 100)) : 0
	);
	const publishedComponentCount = $derived(readiness?.published_component_count ?? components.filter((item) => item.is_published).length);
	const draftComponentCount = $derived(readiness?.draft_component_count ?? components.filter((item) => !item.is_published).length);
	const readyStudentCount = $derived(readiness?.ready_student_count ?? summary.filter((row) => row.component_count > 0 && row.filled_count === row.component_count).length);
	const incompleteStudentCount = $derived(readiness?.incomplete_student_count ?? summary.filter((row) => row.component_count === 0 || row.filled_count < row.component_count).length);
	const missingGradeCount = $derived(readiness?.missing_grade_count ?? summary.reduce((total, row) => total + Math.max(row.component_count - row.filled_count, 0), 0));
	const dirtyEntryIds = $derived(
		entries
			.filter((row) => isEntryDirty(row.student_id))
			.map((row) => row.student_id)
	);
	const readyForRapor = $derived(readiness?.ready ?? false);
	const isFinalized = $derived(finalization !== null);
	const finalizedAssignmentCount = $derived(assignmentStatuses.filter((item) => item.is_finalized).length);
	const readyAssignmentCount = $derived(
		assignmentStatuses.filter((item) => item.ready && !item.is_finalized).length
	);
	const needsAttentionAssignmentCount = $derived(
		assignmentStatuses.filter((item) => !item.ready && !item.is_finalized).length
	);
	const teacherOptions = $derived(
		Array.from(new Set(assignmentStatuses.map((item) => item.teacher_name).filter(Boolean))).sort((a, b) => a.localeCompare(b, 'id'))
	);
	const filteredAssignmentStatuses = $derived.by(() => {
		const statusFiltered = (() => {
			switch (assignmentStatusFilter) {
				case 'ready':
					return assignmentStatuses.filter((item) => item.ready && !item.is_finalized);
				case 'finalized':
					return assignmentStatuses.filter((item) => item.is_finalized);
				case 'attention':
					return assignmentStatuses.filter((item) => !item.ready && !item.is_finalized);
				default:
					return assignmentStatuses;
			}
		})();
		const teacherFiltered = assignmentTeacherFilter
			? statusFiltered.filter((item) => item.teacher_name === assignmentTeacherFilter)
			: statusFiltered;
		const query = assignmentStatusQuery.trim().toLowerCase();
		if (!query) return teacherFiltered;
		return teacherFiltered.filter((item) =>
			[
				item.class_name,
				item.class_code,
				item.subject_name,
				item.subject_code,
				item.teacher_name,
				assignmentStatusLabel(item)
			]
				.join(' ')
				.toLowerCase()
				.includes(query)
		);
	});
	const nextReadyAssignment = $derived(
		assignmentStatuses.find((item) => item.ready && !item.is_finalized) ?? null
	);
	const filteredReadyAssignments = $derived(
		filteredAssignmentStatuses.filter((item) => item.ready && !item.is_finalized)
	);
	const filteredFinalizedAssignments = $derived(
		filteredAssignmentStatuses.filter((item) => item.is_finalized)
	);
	const filteredTeacherSummaries = $derived.by(() => {
		const grouped: Record<string, TeacherReadinessSummary> = {};
		for (const item of filteredAssignmentStatuses) {
			const key = item.teacher_name || 'Tanpa Guru';
			const current = grouped[key] ?? {
				teacher_name: key,
				assignment_count: 0,
				ready_count: 0,
				finalized_count: 0,
				attention_count: 0,
				missing_grade_count: 0
			};
			current.assignment_count += 1;
			current.missing_grade_count += item.missing_grade_count;
			if (item.is_finalized) {
				current.finalized_count += 1;
			} else if (item.ready) {
				current.ready_count += 1;
			} else {
				current.attention_count += 1;
			}
			grouped[key] = current;
		}
		return Object.values(grouped).sort((a, b) => a.teacher_name.localeCompare(b.teacher_name, 'id'));
	});
	const filteredClassSummaries = $derived.by(() => {
		const grouped: Record<string, ClassReadinessSummary> = {};
		for (const item of filteredAssignmentStatuses) {
			const key = `${item.class_code}::${item.class_name}`;
			const current = grouped[key] ?? {
				class_name: item.class_name,
				class_code: item.class_code,
				assignment_count: 0,
				ready_count: 0,
				finalized_count: 0,
				attention_count: 0,
				missing_grade_count: 0
			};
			current.assignment_count += 1;
			current.missing_grade_count += item.missing_grade_count;
			if (item.is_finalized) {
				current.finalized_count += 1;
			} else if (item.ready) {
				current.ready_count += 1;
			} else {
				current.attention_count += 1;
			}
			grouped[key] = current;
		}
		return Object.values(grouped).sort((a, b) => a.class_code.localeCompare(b.class_code, 'id'));
	});
	const focusedClassSummary = $derived(
		filteredClassSummaries.find((item) => `${item.class_code}::${item.class_name}` === classFocusKey) ?? filteredClassSummaries[0] ?? null
	);
	const focusedClassAssignments = $derived(
		focusedClassSummary
			? filteredAssignmentStatuses.filter((item) =>
				item.class_code === focusedClassSummary.class_code && item.class_name === focusedClassSummary.class_name
			)
			: []
	);
	const readinessLabel = $derived(
		isFinalized
			? 'Sudah Difinalisasi'
			: readyForRapor
			? 'Siap Rapor'
			: components.length === 0
				? 'Belum Siap'
				: draftComponentCount > 0
					? 'Masih Ada Draft'
					: 'Nilai Belum Lengkap'
	);
	const readinessDescription = $derived(
		isFinalized
			? 'Assignment ini sudah difinalisasi. Buka finalisasi terlebih dahulu jika ingin mengubah komponen atau nilai.'
			: readyForRapor
			? 'Semua komponen sudah terbit dan seluruh siswa telah memiliki isian nilai lengkap.'
			: components.length === 0
				? 'Tambahkan komponen penilaian terlebih dahulu sebelum kelas-mapel ini dapat difinalisasi.'
				: draftComponentCount > 0
					? 'Masih ada komponen yang belum diterbitkan untuk rapor.'
					: 'Masih ada siswa atau komponen yang belum terisi penuh.'
	);

	function showSuccess(message: string) {
		toast.success(message);
	}

	function showError(message: string) {
		toast.error(message);
	}

	function categoryLabel(value: string) {
		return categoryOptions.find((item) => item.value === value)?.label ?? value;
	}

	function nonTestTypeLabel(value: string | undefined) {
		switch (value) {
			case 'praktik':
				return 'Praktik';
			case 'portofolio':
				return 'Portofolio';
			case 'proyek':
				return 'Proyek';
			case 'penugasan':
				return 'Penugasan';
			case 'observasi':
				return 'Observasi';
			case 'lainnya':
				return 'Lainnya';
			default:
				return 'Non-Tes';
		}
	}

	function isNonTestComponent(component: GradeComponent | null | undefined) {
		return Boolean(component?.source_non_test_assessment_id);
	}

	function nonTestSourceHref(component: GradeComponent) {
		const sourceId = component.source_non_test_assessment_id;
		return sourceId ? resolve(`/cbt/non-test?assessment_id=${sourceId}`) : resolve('/cbt/non-test');
	}

	function assignmentStatusLabel(item: AssignmentStatus) {
		if (item.is_finalized) return 'Sudah Final';
		if (item.ready) return 'Siap Difinalkan';
		if (item.component_count === 0) return 'Belum Ada Komponen';
		if (item.draft_component_count > 0) return 'Masih Ada Draft';
		return 'Perlu Dilengkapi';
	}

	function assignmentStatusVariant(item: AssignmentStatus): 'default' | 'secondary' | 'outline' {
		if (item.is_finalized || item.ready) return 'default';
		if (item.component_count === 0) return 'outline';
		return 'secondary';
	}

	function assignmentStatusDescription(item: AssignmentStatus) {
		if (item.is_finalized) {
			return item.finalization?.finalized_by
				? `Difinalisasi oleh ${item.finalization.finalized_by}.`
				: 'Assignment ini sudah difinalisasi.';
		}
		if (item.ready) {
			return 'Semua komponen sudah terbit dan seluruh siswa sudah lengkap.';
		}
		if (item.component_count === 0) {
			return 'Belum ada komponen nilai.';
		}
		if (item.draft_component_count > 0) {
			return `${item.draft_component_count} komponen masih draft.`;
		}
		return `${item.missing_grade_count} slot nilai masih kosong.`;
	}

	function assignmentFilterLabel(value: 'all' | 'ready' | 'finalized' | 'attention') {
		switch (value) {
			case 'ready':
				return 'Siap Difinalkan';
			case 'finalized':
				return 'Sudah Final';
			case 'attention':
				return 'Perlu Dilengkapi';
			default:
				return 'Semua';
		}
	}

	async function focusAssignment(nextAssignmentId: string) {
		assignmentId = nextAssignmentId;
		componentId = '';
		gradeWorkspace = 'gradebook';
		resetComponentForm();
		await loadOverview();
	}

	async function selectAssignmentContext() {
		componentId = '';
		if (assignmentId && gradeWorkspace === 'triage') {
			gradeWorkspace = 'gradebook';
		}
		resetComponentForm();
		await loadOverview();
	}

	function csvCell(value: string | number) {
		const normalized = String(value ?? '').replaceAll('"', '""');
		return `"${normalized}"`;
	}

	async function exportAssignmentStatusSummary() {
		if (filteredAssignmentStatuses.length === 0) {
			showError('Tidak ada data rekap yang bisa diekspor untuk filter ini.');
			return;
		}
		exportBusy = true;
		try {
			const header = [
				'Kode Kelas',
				'Nama Kelas',
				'Kode Mapel',
				'Nama Mapel',
				'Guru',
				'Status',
				'Komponen Terbit',
				'Total Komponen',
				'Siswa Siap',
				'Total Siswa',
				'Nilai Kosong'
			];
			const rows = filteredAssignmentStatuses.map((item) => [
				item.class_code,
				item.class_name,
				item.subject_code,
				item.subject_name,
				item.teacher_name,
				assignmentStatusLabel(item),
				item.published_component_count,
				item.component_count,
				item.ready_student_count,
				item.student_count,
				item.missing_grade_count
			]);
			const csv = [header, ...rows].map((row) => row.map(csvCell).join(',')).join('\n');
			const blob = new Blob([csv], { type: 'text/csv;charset=utf-8' });
			const url = URL.createObjectURL(blob);
			const anchor = document.createElement('a');
			const filterSuffix = assignmentStatusFilter === 'all' ? 'semua' : assignmentStatusFilter;
			anchor.href = url;
			anchor.download = `rekap-finalisasi-grade-${filterSuffix}.csv`;
			document.body.append(anchor);
			anchor.click();
			anchor.remove();
			URL.revokeObjectURL(url);
			showSuccess('Rekap finalisasi berhasil diekspor.');
		} finally {
			exportBusy = false;
		}
	}

	async function exportFocusedClassReport() {
		if (!focusedClassSummary || focusedClassAssignments.length === 0) {
			showError('Tidak ada data kelas fokus yang bisa diekspor.');
			return;
		}
		classExportBusy = true;
		try {
			const header = [
				'Kode Kelas',
				'Nama Kelas',
				'Kode Mapel',
				'Nama Mapel',
				'Guru',
				'Status',
				'Komponen Terbit',
				'Total Komponen',
				'Nilai Kosong'
			];
			const rows = focusedClassAssignments.map((item) => [
				item.class_code,
				item.class_name,
				item.subject_code,
				item.subject_name,
				item.teacher_name,
				assignmentStatusLabel(item),
				item.published_component_count,
				item.component_count,
				item.missing_grade_count
			]);
			const csv = [header, ...rows].map((row) => row.map(csvCell).join(',')).join('\n');
			const blob = new Blob([csv], { type: 'text/csv;charset=utf-8' });
			const url = URL.createObjectURL(blob);
			const anchor = document.createElement('a');
			anchor.href = url;
			anchor.download = `report-kesiapan-${focusedClassSummary.class_code.toLowerCase()}.csv`;
			document.body.append(anchor);
			anchor.click();
			anchor.remove();
			URL.revokeObjectURL(url);
			showSuccess(`Report wali kelas ${focusedClassSummary.class_code} berhasil diekspor.`);
		} finally {
			classExportBusy = false;
		}
	}

	async function exportTeacherReadinessReport() {
		if (filteredTeacherSummaries.length === 0) {
			showError('Tidak ada data dashboard guru yang bisa diekspor.');
			return;
		}
		teacherExportBusy = true;
		try {
			const header = [
				'Guru',
				'Jumlah Kelas-Mapel',
				'Siap',
				'Final',
				'Perlu Dilengkapi',
				'Nilai Kosong'
			];
			const rows = filteredTeacherSummaries.map((item) => [
				item.teacher_name,
				item.assignment_count,
				item.ready_count,
				item.finalized_count,
				item.attention_count,
				item.missing_grade_count
			]);
			const csv = [header, ...rows].map((row) => row.map(csvCell).join(',')).join('\n');
			const blob = new Blob([csv], { type: 'text/csv;charset=utf-8' });
			const url = URL.createObjectURL(blob);
			const anchor = document.createElement('a');
			anchor.href = url;
			anchor.download = 'dashboard-kesiapan-guru.csv';
			document.body.append(anchor);
			anchor.click();
			anchor.remove();
			URL.revokeObjectURL(url);
			showSuccess('Dashboard kesiapan guru berhasil diekspor.');
		} finally {
			teacherExportBusy = false;
		}
	}

	async function finalizeAssignmentById(targetAssignmentId: string, notes: string) {
		const res = await fetch(clientApiPath`/api/grades/assignments/${targetAssignmentId}/finalize`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ notes })
		});
		await readClientJson<unknown>(res);
	}

	async function reopenAssignmentById(targetAssignmentId: string) {
		const res = await fetch(clientApiPath`/api/grades/assignments/${targetAssignmentId}/finalize`, {
			method: 'DELETE'
		});
		await readClientJson<unknown>(res);
	}

	function resetComponentForm() {
		editingComponentId = '';
		componentTitle = '';
		componentCategory = 'assignment';
		componentWeight = 1;
		componentMaxScore = 100;
	}

	function finalScoreLabel(value: number) {
		return value < 0 ? '—' : value.toFixed(2);
	}

	function normalizedEntryScore(studentId: string) {
		return (scoreInput[studentId] ?? '').trim();
	}

	function normalizedEntryNote(studentId: string) {
		return (noteInput[studentId] ?? '').trim();
	}

	function originalEntryScore(studentId: string) {
		const row = entries.find((item) => item.student_id === studentId);
		if (!row || row.score < 0) return '';
		return String(row.score);
	}

	function originalEntryNote(studentId: string) {
		const row = entries.find((item) => item.student_id === studentId);
		return (row?.notes ?? '').trim();
	}

	function isEntryDirty(studentId: string) {
		return normalizedEntryScore(studentId) !== originalEntryScore(studentId) || normalizedEntryNote(studentId) !== originalEntryNote(studentId);
	}

	function buildOverviewPath() {
		const params = new URLSearchParams();
		if (assignmentId) params.set('assignment_id', assignmentId);
		if (componentId) params.set('component_id', componentId);
		return clientApiPathWithQuery('/api/grades', params);
	}

	function applyGradeInputs(nextEntries: GradeEntry[]) {
		const nextScores: Record<string, string> = {};
		const nextNotes: Record<string, string> = {};
		for (const row of nextEntries) {
			nextScores[row.student_id] = row.score >= 0 ? String(row.score) : '';
			nextNotes[row.student_id] = row.notes ?? '';
		}
		scoreInput = nextScores;
		noteInput = nextNotes;
		resetQuickFill();
	}

	function normalizeOverview(payload: GradesPayload): GradesOverview {
		return {
			assignments: payload.assignments ?? [],
			assignmentStatuses: payload.assignment_statuses ?? [],
			components: payload.components ?? [],
			summary: payload.summary ?? [],
			entries: payload.entries ?? [],
			readiness: payload.readiness ?? null,
			finalization: payload.finalization ?? null
		};
	}

	function applyOverview(overview: GradesOverview) {
		assignments = overview.assignments;
		assignmentStatuses = overview.assignmentStatuses;
		components = overview.components;
		summary = overview.summary;
		entries = overview.entries;
		readiness = overview.readiness;
		finalization = overview.finalization;
		finalizeNotes = overview.finalization?.notes ?? '';
		applyGradeInputs(overview.entries);
	}

	function currentOverview(): GradesOverview {
		return {
			assignments,
			assignmentStatuses,
			components,
			summary,
			entries,
			readiness,
			finalization
		};
	}

	async function fetchOverview(): Promise<GradesOverview> {
		const payload = await fetch(buildOverviewPath()).then((response) =>
			readClientApiData<GradesPayload>(response, 'Gagal memuat data nilai')
		);
		return normalizeOverview(payload);
	}

	function loadOverview() {
		const requestId = ++overviewRequestId;
		overviewPromise = fetchOverview().then((overview) => {
			if (requestId !== overviewRequestId) return currentOverview();
			applyOverview(overview);
			return overview;
		}).catch((error: unknown) => {
			if (requestId === overviewRequestId) throw error;
			return currentOverview();
		});
		return overviewPromise;
	}

	async function refreshOverview() {
		if (!overviewPromise) {
			await loadOverview();
			return;
		}
		const requestId = ++overviewRequestId;
		try {
			const overview = await fetchOverview();
			if (requestId !== overviewRequestId) return;
			applyOverview(overview);
			overviewPromise = Promise.resolve(overview);
		} catch (error) {
			if (requestId === overviewRequestId) {
				overviewPromise = Promise.resolve(currentOverview());
				showError(overviewErrorMessage(error));
			}
		}
	}

	function retryOverview(reset?: () => void) {
		reset?.();
		loadOverview();
	}

	function overviewErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		if (typeof error === 'string' && error.trim()) return error;
		return 'Gagal memuat data nilai';
	}

	function handleOverviewRenderError(error: unknown, reset: () => void) {
		console.error('Grade overview render failed', error);
		reset();
	}

	function resetQuickFill() {
		quickFillScore = '';
		quickFillNote = '';
	}

	async function quickSelectFirstAssignment() {
		if (assignments.length === 0) return;
		assignmentId = assignments[0]?.id ?? '';
		componentId = '';
		gradeWorkspace = 'gradebook';
		resetComponentForm();
		await loadOverview();
	}

	function beginEditComponent(component: GradeComponent) {
		if (isFinalized) {
			showError('Assignment sudah difinalisasi. Buka finalisasi terlebih dahulu untuk mengubah komponen.');
			return;
		}
		editingComponentId = component.id;
		componentTitle = component.title;
		componentCategory = component.category;
		componentWeight = component.weight;
		componentMaxScore = component.max_score;
	}

	async function saveComponent() {
		if (isFinalized) {
			showError('Assignment sudah difinalisasi. Buka finalisasi terlebih dahulu untuk mengubah komponen.');
			return;
		}
		if ((!assignmentId && !editingComponentId) || !componentTitle) return;
		createBusy = true;
		try {
			const path = editingComponentId ? clientApiPath`/api/grades/components/${editingComponentId}` : '/api/grades/components';
			const method = editingComponentId ? 'PUT' : 'POST';
			const payload: Record<string, unknown> = {
				title: componentTitle,
				category: componentCategory,
				weight: componentWeight,
				max_score: componentMaxScore
			};
			if (!editingComponentId) {
				payload.assignment_id = assignmentId;
				payload.is_published = false;
			}
			const res = await fetch(path, {
				method,
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(payload)
			});
			try {
				await readClientJson<unknown>(res);
			} catch (error) {
				showError(overviewErrorMessage(error));
				return;
			}
			const wasEditing = editingComponentId !== '';
			resetComponentForm();
			showSuccess(wasEditing ? 'Komponen nilai diperbarui' : 'Komponen nilai ditambahkan');
			await refreshOverview();
		} finally {
			createBusy = false;
		}
	}

	async function deleteComponent(id: string) {
		if (isFinalized) {
			showError('Assignment sudah difinalisasi. Buka finalisasi terlebih dahulu untuk mengubah komponen.');
			return;
		}
		if (!(await confirmAction({
			title: 'Hapus Komponen Nilai',
			message: 'Hapus komponen nilai ini? Nilai terkait komponen ini tidak dapat dipakai lagi.',
			confirmLabel: 'Hapus Komponen',
			tone: 'danger'
		}))) return;
		const res = await fetch(clientApiPath`/api/grades/components/${id}`, { method: 'DELETE' });
		try {
			await readClientJson<unknown>(res);
		} catch (error) {
			showError(overviewErrorMessage(error));
			return;
		}
		if (editingComponentId === id) resetComponentForm();
		if (componentId === id) componentId = '';
		showSuccess('Komponen nilai dihapus');
		await refreshOverview();
	}

	async function togglePublish(component: GradeComponent) {
		if (isFinalized) {
			showError('Assignment sudah difinalisasi. Buka finalisasi terlebih dahulu untuk mengubah komponen.');
			return;
		}
		publishBusy = { ...publishBusy, [component.id]: true };
		try {
			const res = await fetch(clientApiPath`/api/grades/components/${component.id}/publish`, {
				method: 'PATCH',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					is_published: !component.is_published
				})
			});
			try {
				await readClientJson<unknown>(res);
			} catch (error) {
				showError(overviewErrorMessage(error));
				return;
			}
			showSuccess(component.is_published ? 'Komponen dikembalikan ke draft' : 'Komponen diterbitkan untuk rapor');
			await refreshOverview();
		} finally {
			publishBusy = { ...publishBusy, [component.id]: false };
		}
	}

	async function persistEntry(studentId: string, silent = false) {
		if (isFinalized) {
			throw new Error('Assignment sudah difinalisasi. Buka finalisasi terlebih dahulu untuk mengubah nilai.');
		}
		if (!componentId) return;
		const rawScore = normalizedEntryScore(studentId);
		if (rawScore === '') {
			throw new Error('Nilai wajib diisi');
		}
		const score = Number(rawScore);
		if (Number.isNaN(score)) {
			throw new Error('Nilai harus berupa angka');
		}
		entryBusy = { ...entryBusy, [studentId]: true };
		try {
			const res = await fetch(clientApiPath`/api/grades/components/${componentId}/entries`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					student_id: studentId,
					score,
					notes: noteInput[studentId] ?? ''
				})
			});
			await readClientJson<unknown>(res);
			if (!silent) {
				showSuccess('Nilai siswa diperbarui');
			}
		} finally {
			entryBusy = { ...entryBusy, [studentId]: false };
		}
	}

	async function saveEntry(studentId: string) {
		try {
			await persistEntry(studentId);
			await refreshOverview();
		} catch (err) {
			showError(err instanceof Error ? err.message : 'Gagal menyimpan nilai');
		}
	}

	async function saveAllDirtyEntries() {
		if (!selectedComponent || dirtyEntryIds.length === 0) return;
		bulkSaveBusy = true;
		const failedStudents: string[] = [];
		try {
			for (const studentId of dirtyEntryIds) {
				try {
					await persistEntry(studentId, true);
				} catch {
					const row = entries.find((item) => item.student_id === studentId);
					failedStudents.push(row?.nama ?? 'Siswa tanpa nama');
				}
			}
			if (failedStudents.length > 0) {
				showError(`Sebagian nilai gagal disimpan: ${failedStudents.slice(0, 3).join(', ')}${failedStudents.length > 3 ? ' dan lainnya' : ''}.`);
			} else {
				showSuccess(`Semua perubahan untuk ${dirtyEntryIds.length} siswa berhasil disimpan.`);
			}
			await refreshOverview();
		} finally {
			bulkSaveBusy = false;
		}
	}

	function applyQuickFill(mode: 'all' | 'empty') {
		if (isFinalized) {
			showError('Assignment sudah difinalisasi. Buka finalisasi terlebih dahulu untuk mengubah nilai.');
			return;
		}
		if (!selectedComponent) return;
		const normalizedScore = quickFillScore.trim();
		const normalizedNote = quickFillNote.trim();
		if (normalizedScore === '' && normalizedNote === '') {
			showError('Isi nilai atau catatan massal terlebih dahulu');
			return;
		}
		if (normalizedScore !== '') {
			const score = Number(normalizedScore);
			if (Number.isNaN(score)) {
				showError('Nilai massal harus berupa angka');
				return;
			}
			if (score < 0) {
				showError('Nilai massal tidak boleh negatif');
				return;
			}
			if (score > selectedComponent.max_score) {
				showError(`Nilai massal melebihi skor maksimum ${selectedComponent.max_score}`);
				return;
			}
		}

		const nextScores = { ...scoreInput };
		const nextNotes = { ...noteInput };
		let changedCount = 0;

		for (const row of entries) {
			const currentScore = normalizedEntryScore(row.student_id);
			const currentNote = normalizedEntryNote(row.student_id);
			const shouldApply = mode === 'all' || (currentScore === '' && currentNote === '');
			if (!shouldApply) continue;

			let changed = false;
			if (normalizedScore !== '' && currentScore !== normalizedScore) {
				nextScores[row.student_id] = normalizedScore;
				changed = true;
			}
			if (normalizedNote !== '' && currentNote !== normalizedNote) {
				nextNotes[row.student_id] = normalizedNote;
				changed = true;
			}
			if (changed) changedCount += 1;
		}

		scoreInput = nextScores;
		noteInput = nextNotes;
		if (changedCount === 0) {
			showError(mode === 'all' ? 'Tidak ada baris yang berubah dari quick fill' : 'Tidak ada baris kosong yang bisa diisi');
			return;
		}
		showSuccess(mode === 'all' ? `Quick fill diterapkan ke ${changedCount} siswa` : `Quick fill diterapkan ke ${changedCount} siswa yang masih kosong`);
	}

	async function finalizeAssignment() {
		if (!selectedAssignment || !readyForRapor) return;
		finalizationBusy = true;
		try {
			await finalizeAssignmentById(selectedAssignment.id, finalizeNotes);
			showSuccess('Assignment siap rapor sudah difinalisasi');
			await refreshOverview();
		} finally {
			finalizationBusy = false;
		}
	}

	async function finalizeFilteredReadyAssignments() {
		if (filteredReadyAssignments.length === 0) {
			showError('Tidak ada assignment siap-final pada filter aktif.');
			return;
		}
		if (!(await confirmAction({
			title: 'Finalisasi Batch Nilai',
			message: `Finalisasi ${filteredReadyAssignments.length} assignment siap-final dari hasil filter saat ini? Assignment yang difinalisasi akan masuk checkpoint rapor.`,
			confirmLabel: 'Finalisasi Batch',
			tone: 'warning'
		}))) {
			return;
		}

		batchFinalizationBusy = true;
		const failedAssignments: string[] = [];
		try {
			for (const item of filteredReadyAssignments) {
				try {
					await finalizeAssignmentById(item.assignment_id, finalizeNotes);
				} catch {
					failedAssignments.push(`${item.class_code} · ${item.subject_code}`);
				}
			}
			if (failedAssignments.length > 0) {
				showError(`Sebagian finalisasi gagal: ${failedAssignments.slice(0, 3).join(', ')}${failedAssignments.length > 3 ? ' dan lainnya' : ''}.`);
			} else {
				showSuccess(`${filteredReadyAssignments.length} assignment siap-final berhasil difinalisasi.`);
			}
			await refreshOverview();
		} finally {
			batchFinalizationBusy = false;
		}
	}

	async function reopenFilteredFinalizedAssignments() {
		if (filteredFinalizedAssignments.length === 0) {
			showError('Tidak ada assignment final pada filter aktif.');
			return;
		}
		if (!(await confirmAction({
			title: 'Buka Finalisasi Batch',
			message: `Buka kembali ${filteredFinalizedAssignments.length} assignment final dari hasil filter saat ini? Gunakan hanya untuk koreksi terkontrol.`,
			confirmLabel: 'Buka Finalisasi',
			tone: 'warning'
		}))) {
			return;
		}

		batchReopenBusy = true;
		const failedAssignments: string[] = [];
		try {
			for (const item of filteredFinalizedAssignments) {
				try {
					await reopenAssignmentById(item.assignment_id);
				} catch {
					failedAssignments.push(`${item.class_code} · ${item.subject_code}`);
				}
			}
			if (failedAssignments.length > 0) {
				showError(`Sebagian pembukaan finalisasi gagal: ${failedAssignments.slice(0, 3).join(', ')}${failedAssignments.length > 3 ? ' dan lainnya' : ''}.`);
			} else {
				showSuccess(`${filteredFinalizedAssignments.length} assignment final berhasil dibuka kembali.`);
			}
			await refreshOverview();
		} finally {
			batchReopenBusy = false;
		}
	}

	async function reopenFinalization() {
		if (!selectedAssignment) return;
		finalizationBusy = true;
		try {
			await reopenAssignmentById(selectedAssignment.id);
			showSuccess('Finalisasi assignment dibuka kembali');
			await refreshOverview();
		} finally {
			finalizationBusy = false;
		}
	}

	onMount(() => {
		void loadOverview();
	});
</script>

<svelte:head>
	<title>Nilai — MTSN 2 Kolaka Utara</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex flex-col gap-2 lg:flex-row lg:items-end lg:justify-between">
		<div>
			<p class="text-xs font-semibold uppercase tracking-[0.28em] text-emerald-700">Sprint 11</p>
			<h1 class="text-3xl font-semibold text-slate-900">Grade Management</h1>
			<p class="mt-1 max-w-3xl text-sm text-slate-600">Kelola komponen penilaian per kelas dan mata pelajaran, input nilai siswa, lalu pantau rekap capaian secara bertahap sebelum rapor final dibentuk.</p>
		</div>
		<div class="grid gap-2 sm:grid-cols-2 lg:w-[32rem]">
			<div>
				<label for="assignment-id" class="mb-1 block text-xs font-medium text-slate-500">Pilih Kelas-Mapel</label>
					<select
						id="assignment-id"
						class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
						bind:value={assignmentId}
						onchange={() => void selectAssignmentContext()}
					>
					<option value="">Pilih penugasan kelas-mapel</option>
					{#each assignments as item (item.id)}
						<option value={item.id}>{item.class_code} · {item.subject_code} · {item.teacher_name}</option>
					{/each}
				</select>
			</div>
			<div class="rounded-xl border border-emerald-100 bg-emerald-50 px-4 py-3">
				<p class="text-xs font-semibold uppercase tracking-[0.2em] text-emerald-700">Cakupan Isi</p>
				<p class="mt-2 text-2xl font-semibold text-emerald-900">{summary.length}</p>
				<p class="text-sm text-emerald-800">siswa aktif dalam gradebook terpilih</p>
			</div>
		</div>
	</div>

	<div class="grid gap-2 rounded-2xl border border-slate-200 bg-white p-2 shadow-sm sm:grid-cols-3">
		{#each [
			{ id: 'triage', label: 'Triase Rapor', desc: 'Pantau kesiapan lintas kelas-mapel' },
			{ id: 'gradebook', label: 'Gradebook', desc: 'Kelola komponen dan input nilai' },
			{ id: 'finalization', label: 'Finalisasi', desc: 'Kunci atau buka checkpoint rapor' }
		] as item (item.id)}
			<button
				type="button"
				class={`rounded-xl px-4 py-3 text-left transition-colors ${gradeWorkspace === item.id
					? 'bg-emerald-50 text-emerald-900 ring-1 ring-emerald-200'
					: 'text-slate-600 hover:bg-slate-50'}`}
				onclick={() => (gradeWorkspace = item.id as GradeWorkspace)}
			>
				<span class="block text-sm font-semibold">{item.label}</span>
				<span class="mt-1 block text-xs leading-5">{item.desc}</span>
			</button>
		{/each}
	</div>

	<AsyncContent promise={overviewPromise} onerror={handleOverviewRenderError}>
		{#snippet pending()}
			<div class="space-y-4 rounded-2xl border border-slate-200 bg-white p-5">
				<div class="grid gap-3 lg:grid-cols-[1.6fr,0.8fr]">
					<Skeleton class="h-14 w-full" />
					<Skeleton class="h-20 w-full" />
				</div>
				<div class="grid gap-4 xl:grid-cols-[0.95fr,1.05fr]">
					<div class="space-y-3">
						<Skeleton class="h-10 w-full" />
						<Skeleton class="h-24 w-full" />
						<Skeleton class="h-24 w-full" />
					</div>
					<div class="space-y-3">
						<Skeleton class="h-12 w-full" />
						<Skeleton class="h-14 w-full" />
						<Skeleton class="h-14 w-full" />
						<Skeleton class="h-14 w-full" />
					</div>
				</div>
			</div>
		{/snippet}
		{#snippet failed(error, reset)}
			<RecoveryPanel
				title="Data Nilai Belum Tersaji"
				message={overviewErrorMessage(error)}
				onRetry={() => retryOverview(reset)}
			/>
		{/snippet}
		{#snippet children(value)}
			{@const overview = value as GradesOverview}
			{@const currentAssignmentStatuses = overview.assignmentStatuses}
		{#if currentAssignmentStatuses.length > 0 && gradeWorkspace === 'triage'}
			<div class="grid gap-4 xl:grid-cols-[0.88fr_1.12fr]">
				<div class="grid gap-4 sm:grid-cols-3 xl:grid-cols-1">
					<Card.Root class="border-emerald-100 bg-white">
						<Card.Content class="pt-5">
							<p class="text-xs font-semibold uppercase tracking-[0.2em] text-emerald-700">Siap Difinalkan</p>
							<p class="mt-2 text-3xl font-semibold text-slate-900">{readyAssignmentCount}</p>
							<p class="text-sm text-slate-600">kelas-mapel yang siap masuk checkpoint finalisasi</p>
						</Card.Content>
					</Card.Root>
					<Card.Root class="border-sky-100 bg-white">
						<Card.Content class="pt-5">
							<p class="text-xs font-semibold uppercase tracking-[0.2em] text-sky-700">Sudah Final</p>
							<p class="mt-2 text-3xl font-semibold text-slate-900">{finalizedAssignmentCount}</p>
							<p class="text-sm text-slate-600">assignment yang sudah dikunci dan siap rapor</p>
						</Card.Content>
					</Card.Root>
					<Card.Root class="border-amber-100 bg-white">
						<Card.Content class="pt-5">
							<p class="text-xs font-semibold uppercase tracking-[0.2em] text-amber-700">Perlu Dilengkapi</p>
							<p class="mt-2 text-3xl font-semibold text-slate-900">{needsAttentionAssignmentCount}</p>
							<p class="text-sm text-slate-600">assignment yang masih perlu komponen, publish, atau isi nilai</p>
						</Card.Content>
					</Card.Root>
					{#if nextReadyAssignment}
						<Card.Root class="border-slate-200 bg-slate-50/80">
							<Card.Content class="space-y-3 pt-5">
								<div>
									<p class="text-xs font-semibold uppercase tracking-[0.2em] text-slate-500">Langkah Cepat</p>
									<p class="mt-2 text-sm font-medium text-slate-900">{nextReadyAssignment.class_code} · {nextReadyAssignment.subject_code}</p>
									<p class="text-sm text-slate-600">Buka assignment siap-final berikutnya agar operator bisa lanjut checkpoint tanpa mencari manual.</p>
								</div>
								<Button variant="outline" onclick={() => focusAssignment(nextReadyAssignment.assignment_id)}>
									Buka Assignment Siap Final
								</Button>
							</Card.Content>
						</Card.Root>
					{/if}
				</div>

				<Card.Root>
					<Card.Header class="pb-2">
						<Card.Title class="text-base">Rekap Finalisasi per Kelas-Mapel</Card.Title>
						<Card.Description>Gunakan ringkasan ini untuk melihat assignment mana yang sudah siap dikunci, mana yang sudah final, dan mana yang masih butuh tindak lanjut.</Card.Description>
					</Card.Header>
					<Card.Content class="space-y-4 p-4 pt-0">
						<div class="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
							<div class="flex flex-wrap gap-2">
								<Button
									variant={assignmentStatusFilter === 'all' ? 'default' : 'outline'}
									size="sm"
									onclick={() => { assignmentStatusFilter = 'all'; }}
								>
									Semua ({assignmentStatuses.length})
								</Button>
								<Button
									variant={assignmentStatusFilter === 'ready' ? 'default' : 'outline'}
									size="sm"
									onclick={() => { assignmentStatusFilter = 'ready'; }}
								>
									Siap Difinalkan ({readyAssignmentCount})
								</Button>
								<Button
									variant={assignmentStatusFilter === 'finalized' ? 'default' : 'outline'}
									size="sm"
									onclick={() => { assignmentStatusFilter = 'finalized'; }}
								>
									Sudah Final ({finalizedAssignmentCount})
								</Button>
								<Button
									variant={assignmentStatusFilter === 'attention' ? 'default' : 'outline'}
									size="sm"
									onclick={() => { assignmentStatusFilter = 'attention'; }}
								>
									Perlu Dilengkapi ({needsAttentionAssignmentCount})
								</Button>
							</div>
							<div class="flex flex-col gap-3 sm:flex-row sm:items-end">
								<div class="min-w-0 sm:w-56">
									<label for="assignment-teacher-filter" class="mb-1 block text-xs font-medium text-slate-500">Filter Guru</label>
									<select
										id="assignment-teacher-filter"
										class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
										bind:value={assignmentTeacherFilter}
									>
										<option value="">Semua guru</option>
										{#each teacherOptions as teacher (teacher)}
											<option value={teacher}>{teacher}</option>
										{/each}
									</select>
								</div>
								<div class="min-w-0 sm:w-72">
									<label for="assignment-status-query" class="mb-1 block text-xs font-medium text-slate-500">Cari Kelas, Mapel, Guru</label>
									<Input
										id="assignment-status-query"
										placeholder="Mis. VIIA, Matematika, Ibu Siti"
										bind:value={assignmentStatusQuery}
									/>
								</div>
								<LoadingButton
									variant="outline"
									loading={exportBusy}
									loadingLabel="Mengekspor..."
									disabled={filteredAssignmentStatuses.length === 0}
									onclick={() => void exportAssignmentStatusSummary()}
									label="Ekspor Rekap"
								/>
							</div>
						</div>
							<div class="flex flex-wrap items-center gap-2 text-xs text-slate-500">
								<Badge variant="outline">Filter: {assignmentFilterLabel(assignmentStatusFilter)}</Badge>
								<Badge variant="outline">Hasil: {filteredAssignmentStatuses.length}</Badge>
								<Badge variant="outline">Siap Final: {filteredReadyAssignments.length}</Badge>
								<Badge variant="outline">Sudah Final: {filteredFinalizedAssignments.length}</Badge>
								{#if assignmentTeacherFilter}
									<Badge variant="outline">Guru: {assignmentTeacherFilter}</Badge>
								{/if}
								{#if assignmentStatusQuery.trim()}
									<Badge variant="outline">Pencarian: {assignmentStatusQuery.trim()}</Badge>
								{/if}
						</div>
						{#if filteredReadyAssignments.length > 0}
							<div class="flex flex-col gap-3 rounded-xl border border-emerald-200 bg-emerald-50/60 px-4 py-4 lg:flex-row lg:items-center lg:justify-between">
								<div>
									<p class="text-xs font-semibold uppercase tracking-[0.2em] text-emerald-700">Finalisasi Batch</p>
									<p class="mt-2 text-sm font-medium text-slate-900">{filteredReadyAssignments.length} assignment siap-final ada di hasil filter aktif.</p>
									<p class="text-sm text-slate-600">Gunakan catatan finalisasi yang sama bila operator ingin menutup checkpoint rapor untuk beberapa kelas-mapel sekaligus.</p>
								</div>
								<LoadingButton
									loading={batchFinalizationBusy}
									loadingLabel="Memfinalisasi batch..."
									onclick={() => void finalizeFilteredReadyAssignments()}
									label="Finalisasi Semua yang Siap"
								/>
							</div>
						{/if}
						{#if filteredFinalizedAssignments.length > 0}
							<div class="flex flex-col gap-3 rounded-xl border border-sky-200 bg-sky-50/60 px-4 py-4 lg:flex-row lg:items-center lg:justify-between">
								<div>
									<p class="text-xs font-semibold uppercase tracking-[0.2em] text-sky-700">Buka Finalisasi Batch</p>
									<p class="mt-2 text-sm font-medium text-slate-900">{filteredFinalizedAssignments.length} assignment final ada di hasil filter aktif.</p>
									<p class="text-sm text-slate-600">Gunakan flow ini saat operator perlu membuka kembali beberapa assignment final untuk koreksi terkontrol.</p>
								</div>
								<LoadingButton
									variant="outline"
									loading={batchReopenBusy}
									loadingLabel="Membuka batch..."
									onclick={() => void reopenFilteredFinalizedAssignments()}
									label="Buka Semua yang Final"
								/>
							</div>
						{/if}
						{#if filteredClassSummaries.length > 0}
							<div class="rounded-xl border border-slate-200 bg-slate-50/70 p-4">
								<div class="mb-3">
									<p class="text-xs font-semibold uppercase tracking-[0.2em] text-slate-500">Ringkasan per Kelas</p>
									<p class="mt-1 text-sm text-slate-600">Gunakan rollup ini untuk membaca kelas mana yang sudah hampir siap rapor dan mana yang masih tertahan di beberapa mapel.</p>
								</div>
								<div class="overflow-x-auto">
									<Table.Root>
										<Table.Header>
											<Table.Row>
												<Table.Head>Kelas</Table.Head>
												<Table.Head>Mapel</Table.Head>
												<Table.Head>Siap</Table.Head>
												<Table.Head>Final</Table.Head>
												<Table.Head>Perlu Dilengkapi</Table.Head>
												<Table.Head>Nilai Kosong</Table.Head>
											</Table.Row>
										</Table.Header>
										<Table.Body>
											{#each filteredClassSummaries as item (`${item.class_code}-${item.class_name}`)}
												<Table.Row class={`${focusedClassSummary && focusedClassSummary.class_code === item.class_code && focusedClassSummary.class_name === item.class_name ? 'bg-emerald-50/70' : ''}`}>
													<Table.Cell>
														<button
															class="text-left"
															onclick={() => {
																classFocusKey = `${item.class_code}::${item.class_name}`;
															}}
														>
															<div class="font-medium text-slate-900">{item.class_name}</div>
															<div class="text-xs text-slate-500">{item.class_code}</div>
														</button>
													</Table.Cell>
													<Table.Cell>{item.assignment_count}</Table.Cell>
													<Table.Cell>{item.ready_count}</Table.Cell>
													<Table.Cell>{item.finalized_count}</Table.Cell>
													<Table.Cell>{item.attention_count}</Table.Cell>
													<Table.Cell>{item.missing_grade_count}</Table.Cell>
												</Table.Row>
											{/each}
										</Table.Body>
									</Table.Root>
								</div>
							</div>
						{/if}
						{#if filteredTeacherSummaries.length > 0}
							<div class="rounded-xl border border-sky-200 bg-sky-50/50 p-4">
								<div class="mb-3 flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
									<div>
										<p class="text-xs font-semibold uppercase tracking-[0.2em] text-sky-700">Dashboard Lintas Guru</p>
										<p class="mt-1 text-sm text-slate-600">Pantau distribusi kesiapan rapor per guru dari hasil filter aktif sebelum turun ke assignment atau kelas tertentu.</p>
									</div>
									<LoadingButton
										variant="outline"
										loading={teacherExportBusy}
										loadingLabel="Mengekspor..."
										onclick={() => void exportTeacherReadinessReport()}
										label="Ekspor Dashboard Guru"
									/>
								</div>
								<div class="overflow-x-auto">
									<Table.Root>
										<Table.Header>
											<Table.Row>
												<Table.Head>Guru</Table.Head>
												<Table.Head>Mapel/Kelas</Table.Head>
												<Table.Head>Siap</Table.Head>
												<Table.Head>Final</Table.Head>
												<Table.Head>Perlu Dilengkapi</Table.Head>
												<Table.Head>Nilai Kosong</Table.Head>
											</Table.Row>
										</Table.Header>
										<Table.Body>
											{#each filteredTeacherSummaries as item (item.teacher_name)}
												<Table.Row>
													<Table.Cell class="font-medium text-slate-900">{item.teacher_name}</Table.Cell>
													<Table.Cell>{item.assignment_count}</Table.Cell>
													<Table.Cell>{item.ready_count}</Table.Cell>
													<Table.Cell>{item.finalized_count}</Table.Cell>
													<Table.Cell>{item.attention_count}</Table.Cell>
													<Table.Cell>{item.missing_grade_count}</Table.Cell>
												</Table.Row>
											{/each}
										</Table.Body>
									</Table.Root>
								</div>
							</div>
						{/if}
						{#if focusedClassSummary}
							<div class="rounded-xl border border-emerald-200 bg-white p-4">
								<div class="mb-3 flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
									<div>
										<p class="text-xs font-semibold uppercase tracking-[0.2em] text-emerald-700">Fokus Wali Kelas</p>
										<h3 class="mt-1 text-base font-semibold text-slate-900">{focusedClassSummary.class_name} · {focusedClassSummary.class_code}</h3>
										<p class="text-sm text-slate-600">Ringkasan cepat semua mapel pada kelas aktif dari hasil filter saat ini.</p>
									</div>
									<div class="flex flex-wrap gap-2">
										<Badge variant="outline">{focusedClassSummary.assignment_count} mapel</Badge>
										<Badge variant="outline">{focusedClassSummary.ready_count} siap</Badge>
										<Badge variant="outline">{focusedClassSummary.finalized_count} final</Badge>
										<Badge variant="outline">{focusedClassSummary.attention_count} perlu dilengkapi</Badge>
										<LoadingButton
											variant="outline"
											loading={classExportBusy}
											loadingLabel="Mengekspor..."
											onclick={() => void exportFocusedClassReport()}
											label="Ekspor Report Kelas"
										/>
									</div>
								</div>
								<div class="overflow-x-auto">
									<Table.Root>
										<Table.Header>
											<Table.Row>
												<Table.Head>Mapel</Table.Head>
												<Table.Head>Guru</Table.Head>
												<Table.Head>Status</Table.Head>
												<Table.Head>Komponen</Table.Head>
												<Table.Head>Nilai Kosong</Table.Head>
											</Table.Row>
										</Table.Header>
										<Table.Body>
											{#each focusedClassAssignments as item (item.assignment_id)}
												<Table.Row class={item.assignment_id === assignmentId ? 'bg-emerald-50/70' : ''}>
													<Table.Cell>
														<button class="text-left" onclick={() => focusAssignment(item.assignment_id)}>
															<div class="font-medium text-slate-900">{item.subject_name}</div>
															<div class="text-xs text-slate-500">{item.subject_code}</div>
														</button>
													</Table.Cell>
													<Table.Cell>{item.teacher_name}</Table.Cell>
													<Table.Cell><Badge variant={assignmentStatusVariant(item)}>{assignmentStatusLabel(item)}</Badge></Table.Cell>
													<Table.Cell>{item.published_component_count}/{item.component_count}</Table.Cell>
													<Table.Cell>{item.missing_grade_count}</Table.Cell>
												</Table.Row>
											{/each}
										</Table.Body>
									</Table.Root>
								</div>
							</div>
						{/if}
						<div class="overflow-x-auto">
							<Table.Root>
								<Table.Header>
									<Table.Row>
										<Table.Head>Kelas-Mapel</Table.Head>
										<Table.Head>Status</Table.Head>
										<Table.Head>Komponen</Table.Head>
										<Table.Head>Siswa Siap</Table.Head>
										<Table.Head>Nilai Kosong</Table.Head>
									</Table.Row>
								</Table.Header>
								<Table.Body>
									{#each filteredAssignmentStatuses as item (item.assignment_id)}
										<Table.Row class={item.assignment_id === assignmentId ? 'bg-emerald-50/70' : ''}>
											<Table.Cell>
												<button
													class="text-left"
													onclick={() => focusAssignment(item.assignment_id)}
												>
													<div class="font-medium text-slate-900">{item.class_name} · {item.subject_name}</div>
													<div class="text-xs text-slate-500">{item.class_code} · {item.subject_code} · {item.teacher_name}</div>
												</button>
											</Table.Cell>
											<Table.Cell>
												<div class="space-y-2">
													<Badge variant={assignmentStatusVariant(item)}>{assignmentStatusLabel(item)}</Badge>
													<p class="text-xs text-slate-500">{assignmentStatusDescription(item)}</p>
												</div>
											</Table.Cell>
											<Table.Cell>{item.published_component_count}/{item.component_count}</Table.Cell>
											<Table.Cell>{item.ready_student_count}/{item.student_count}</Table.Cell>
											<Table.Cell>{item.missing_grade_count}</Table.Cell>
										</Table.Row>
									{:else}
										<Table.Row>
											<Table.Cell colspan={5} class="p-4">
												<EmptyStatePanel
													compact
													eyebrow="Filter Triase"
													title={`Tidak ada assignment pada kategori ${assignmentFilterLabel(assignmentStatusFilter)}`}
													description="Ubah filter rekap untuk melihat assignment lain yang sudah final, siap difinalkan, atau masih perlu dilengkapi."
												/>
											</Table.Cell>
										</Table.Row>
									{/each}
								</Table.Body>
							</Table.Root>
						</div>
					</Card.Content>
				</Card.Root>
			</div>
		{/if}

		{#if selectedAssignment && gradeWorkspace !== 'triage'}
			<div class="grid gap-4 md:grid-cols-3">
				<Card.Root class="border-emerald-100 bg-white">
					<Card.Content class="pt-5">
						<p class="text-xs font-semibold uppercase tracking-[0.2em] text-emerald-700">Kelas & Mapel</p>
						<p class="mt-2 text-lg font-semibold text-slate-900">{selectedAssignment.class_name}</p>
						<p class="text-sm text-slate-600">{selectedAssignment.subject_name} · {selectedAssignment.subject_code}</p>
						{#if selectedAssignmentStatus}
							<div class="mt-3 flex flex-wrap gap-2">
								<Badge variant={assignmentStatusVariant(selectedAssignmentStatus)}>
									{assignmentStatusLabel(selectedAssignmentStatus)}
								</Badge>
								<Badge variant="outline">{selectedAssignment.teacher_name}</Badge>
							</div>
						{/if}
					</Card.Content>
				</Card.Root>
				<Card.Root class="border-amber-100 bg-white">
					<Card.Content class="pt-5">
						<p class="text-xs font-semibold uppercase tracking-[0.2em] text-amber-700">Komponen Nilai</p>
						<p class="mt-2 text-3xl font-semibold text-slate-900">{publishedComponentCount}/{components.length}</p>
						<p class="text-sm text-slate-600">sudah terbit untuk rapor dari total komponen yang disusun</p>
					</Card.Content>
				</Card.Root>
				<Card.Root class="border-sky-100 bg-white">
					<Card.Content class="pt-5">
						<p class="text-xs font-semibold uppercase tracking-[0.2em] text-sky-700">Progress Pengisian</p>
						<p class="mt-2 text-3xl font-semibold text-slate-900">{completionRate}%</p>
						<p class="text-sm text-slate-600">siswa yang sudah punya minimal satu nilai</p>
					</Card.Content>
				</Card.Root>
			</div>

			<Card.Root class={readyForRapor ? 'border-emerald-200 bg-emerald-50/70' : 'border-amber-200 bg-amber-50/70'}>
				<Card.Header class="pb-2">
					<div class="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
						<div>
							<Card.Title class="text-base">Kesiapan Rapor</Card.Title>
							<Card.Description>{readinessDescription}</Card.Description>
						</div>
						<Badge variant={readyForRapor ? 'default' : 'secondary'}>{readinessLabel}</Badge>
					</div>
				</Card.Header>
				<Card.Content class="grid gap-4 md:grid-cols-4">
					<div>
						<p class="text-xs font-semibold uppercase tracking-[0.2em] text-slate-500">Komponen Terbit</p>
						<p class="mt-2 text-2xl font-semibold text-slate-900">{publishedComponentCount}</p>
						<p class="text-sm text-slate-600">{draftComponentCount} masih draft</p>
					</div>
					<div>
						<p class="text-xs font-semibold uppercase tracking-[0.2em] text-slate-500">Siswa Siap</p>
						<p class="mt-2 text-2xl font-semibold text-slate-900">{readyStudentCount}/{summary.length}</p>
						<p class="text-sm text-slate-600">{incompleteStudentCount} siswa belum lengkap</p>
					</div>
					<div>
						<p class="text-xs font-semibold uppercase tracking-[0.2em] text-slate-500">Nilai Belum Masuk</p>
						<p class="mt-2 text-2xl font-semibold text-slate-900">{missingGradeCount}</p>
						<p class="text-sm text-slate-600">slot nilai yang masih perlu diisi</p>
					</div>
					<div class="flex items-end">
						<div class="w-full space-y-3">
							{#if readyForRapor}
									<a
										href={resolve(`/grades/rapor?assignment_id=${assignmentId}`)}
										class="inline-flex w-full items-center justify-center rounded-md bg-emerald-700 px-4 py-2 text-sm font-medium text-white hover:bg-emerald-800"
									>
									Buka Cetak Rapor
								</a>
							{:else}
								<div class="rounded-xl border border-dashed border-amber-300 bg-white/70 px-4 py-3 text-sm text-amber-900">
									Selesaikan draft dan lengkapi semua nilai sebelum membuka rapor final.
								</div>
							{/if}
							<div class="space-y-2">
								<label for="finalize-notes" class="block text-xs font-medium text-slate-500">Catatan Finalisasi</label>
								<Input
									id="finalize-notes"
									placeholder="Opsional: catatan verifikasi guru atau wali kelas"
									bind:value={finalizeNotes}
									disabled={isFinalized}
								/>
							</div>
							{#if isFinalized}
								<div class="rounded-xl border border-emerald-200 bg-white/80 px-4 py-3 text-sm text-emerald-900">
									<p class="font-medium">Difinalisasi oleh {finalization?.finalized_by || 'operator'}.</p>
									<p class="mt-1 text-emerald-800">{new Date(finalization?.finalized_at ?? '').toLocaleString('id-ID')}</p>
									{#if finalization?.notes}
										<p class="mt-2 text-slate-700">{finalization.notes}</p>
									{/if}
								</div>
								<LoadingButton
									class="w-full"
									variant="outline"
									loading={finalizationBusy}
									loadingLabel="Membuka..."
									onclick={() => void reopenFinalization()}
									label="Buka Finalisasi"
								/>
							{:else}
								<LoadingButton
									class="w-full"
									loading={finalizationBusy}
									loadingLabel="Memfinalisasi..."
									disabled={!readyForRapor}
									onclick={() => void finalizeAssignment()}
									label="Finalisasi Assignment"
								/>
							{/if}
						</div>
					</div>
				</Card.Content>
			</Card.Root>

			{#if gradeWorkspace === 'gradebook'}
			<div class="grid gap-4 xl:grid-cols-[1.1fr_0.9fr]">
				<Card.Root>
					<Card.Header class="pb-2">
						<Card.Title class="text-base">Komponen Penilaian</Card.Title>
						<Card.Description>Bangun struktur penilaian per kelas-mapel sebelum nilai rapor dihitung.</Card.Description>
					</Card.Header>
					<Card.Content class="space-y-4">
						<div class="grid gap-3 md:grid-cols-4">
							<div class="md:col-span-2">
								<label for="component-title" class="mb-1 block text-xs font-medium text-slate-500">Judul Komponen</label>
								<Input id="component-title" placeholder="Mis: Tugas Bab 1" bind:value={componentTitle} />
							</div>
							<div>
								<label for="component-category" class="mb-1 block text-xs font-medium text-slate-500">Kategori</label>
								<select id="component-category" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={componentCategory}>
									{#each categoryOptions as item (item.value)}
										<option value={item.value}>{item.label}</option>
									{/each}
								</select>
							</div>
							<div>
								<label for="component-weight" class="mb-1 block text-xs font-medium text-slate-500">Bobot</label>
								<Input id="component-weight" type="number" min="0" step="0.1" bind:value={componentWeight} />
							</div>
						</div>
						<div class="grid gap-3 md:grid-cols-[14rem_auto]">
							<div>
								<label for="component-max-score" class="mb-1 block text-xs font-medium text-slate-500">Skor Maksimum</label>
								<Input id="component-max-score" type="number" min="1" step="0.1" bind:value={componentMaxScore} />
							</div>
							<div class="flex items-end gap-2">
								<LoadingButton
									class="w-full md:w-auto"
									loading={createBusy}
									loadingLabel="Menyimpan..."
									disabled={!componentTitle}
									onclick={() => void saveComponent()}
									label={editingComponent ? 'Simpan Perubahan' : 'Tambah Komponen'}
								/>
								{#if editingComponent}
									<Button variant="outline" onclick={resetComponentForm}>Batal Edit</Button>
								{/if}
							</div>
						</div>

						<div class="overflow-x-auto rounded-lg border border-slate-200">
							<Table.Root>
								<Table.Header>
									<Table.Row>
										<Table.Head>Judul</Table.Head>
										<Table.Head>Kategori</Table.Head>
										<Table.Head>Bobot</Table.Head>
										<Table.Head>Skor Max</Table.Head>
										<Table.Head></Table.Head>
									</Table.Row>
								</Table.Header>
								<Table.Body>
									{#each components as item (item.id)}
										<Table.Row class={componentId === item.id ? 'bg-emerald-50/70' : ''}>
											<Table.Cell>
												<button class="text-left font-medium text-slate-900 hover:text-emerald-700" onclick={async () => { componentId = item.id; await loadOverview(); }}>
													{item.title}
												</button>
												{#if isNonTestComponent(item)}
													<div class="mt-2 flex flex-wrap items-center gap-2">
														<Badge class="border border-emerald-200 bg-emerald-50 text-[10px] font-semibold uppercase tracking-wider text-emerald-800">
															Non-Tes
														</Badge>
														<a
															href={nonTestSourceHref(item)}
															class="text-xs font-medium text-emerald-700 hover:text-emerald-900 hover:underline"
														>
															{item.source_non_test_title || 'Buka asesmen asal'}
														</a>
													</div>
												{/if}
											</Table.Cell>
											<Table.Cell>
												<div class="flex flex-wrap gap-2">
													<Badge variant="outline">{categoryLabel(item.category)}</Badge>
													{#if isNonTestComponent(item)}
														<Badge variant="outline">{nonTestTypeLabel(item.source_non_test_type)}</Badge>
													{/if}
													<Badge variant={item.is_published ? 'default' : 'secondary'}>
														{item.is_published ? 'Terbit' : 'Draft'}
													</Badge>
												</div>
											</Table.Cell>
											<Table.Cell>{item.weight}</Table.Cell>
											<Table.Cell>{item.max_score}</Table.Cell>
											<Table.Cell class="text-right">
												<div class="flex justify-end gap-2">
													<Button variant="outline" size="sm" onclick={() => beginEditComponent(item)}>Edit</Button>
													<LoadingButton
														size="sm"
														variant={item.is_published ? 'outline' : 'default'}
														loading={publishBusy[item.id]}
														loadingLabel="Menyimpan..."
														onclick={() => togglePublish(item)}
														label={item.is_published ? 'Kembalikan ke Draft' : 'Terbitkan'}
													/>
													<Button variant="destructive" size="xs" onclick={() => deleteComponent(item.id)}>Hapus</Button>
												</div>
											</Table.Cell>
										</Table.Row>
										{:else}
											<Table.Row>
												<Table.Cell colspan={5} class="p-4">
													<EmptyStatePanel
														compact
														eyebrow="Bangun Struktur Nilai"
														title="Belum ada komponen penilaian"
														description="Tambahkan komponen seperti tugas, kuis, UTS, atau praktik agar guru bisa mulai mengisi capaian siswa."
													/>
												</Table.Cell>
											</Table.Row>
										{/each}
								</Table.Body>
							</Table.Root>
						</div>
					</Card.Content>
				</Card.Root>

				<Card.Root>
					<Card.Header class="pb-2">
						<Card.Title class="text-base">Rekap Hasil Sementara</Card.Title>
						<Card.Description>Nilai akhir sementara dihitung dari bobot komponen yang sudah terisi.</Card.Description>
					</Card.Header>
					<Card.Content class="p-0">
						<div class="overflow-x-auto">
							<Table.Root>
								<Table.Header>
									<Table.Row>
										<Table.Head>Siswa</Table.Head>
										<Table.Head>Terisi</Table.Head>
										<Table.Head>Nilai Akhir</Table.Head>
									</Table.Row>
								</Table.Header>
								<Table.Body>
									{#each summary as row (row.student_id)}
										<Table.Row>
											<Table.Cell>
												<div class="font-medium text-slate-900">{row.nama}</div>
												<div class="text-xs text-slate-500">{row.nis || row.nisn || 'Tanpa NIS/NISN'}</div>
											</Table.Cell>
											<Table.Cell>{row.filled_count}/{row.component_count}</Table.Cell>
											<Table.Cell class="font-semibold text-slate-900">{finalScoreLabel(row.final_score)}</Table.Cell>
										</Table.Row>
										{:else}
											<Table.Row>
												<Table.Cell colspan={3} class="p-4">
													<EmptyStatePanel
														compact
														eyebrow="Belum Ada Peserta"
														title="Gradebook ini belum memiliki siswa"
														description="Periksa penugasan kelas-mapel dan pastikan kelas terkait sudah berisi siswa aktif."
													/>
												</Table.Cell>
											</Table.Row>
										{/each}
								</Table.Body>
							</Table.Root>
						</div>
					</Card.Content>
				</Card.Root>
			</div>

			<Card.Root>
				<Card.Header class="pb-2">
					<Card.Title class="text-base">Input Nilai Komponen</Card.Title>
					<Card.Description>
						{#if selectedComponent}
							{selectedComponent.title} · {categoryLabel(selectedComponent.category)} · maksimum {selectedComponent.max_score}
							{#if isNonTestComponent(selectedComponent)}
								· sumber non-tes: {selectedComponent.source_non_test_title || nonTestTypeLabel(selectedComponent.source_non_test_type)}
							{/if}
						{:else}
							Pilih komponen penilaian untuk mulai mengisi nilai siswa.
						{/if}
					</Card.Description>
				</Card.Header>
				<Card.Content class="space-y-4 p-0">
					{#if selectedComponent}
						{#if isNonTestComponent(selectedComponent)}
							<div class="mx-6 mt-6 rounded-xl border border-emerald-200 bg-emerald-50/70 px-4 py-3">
								<div class="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
									<div>
										<p class="text-xs font-semibold uppercase tracking-[0.2em] text-emerald-700">Sumber Nilai Non-Tes</p>
										<p class="mt-1 text-sm font-medium text-slate-900">
											{selectedComponent.source_non_test_title || nonTestTypeLabel(selectedComponent.source_non_test_type)}
										</p>
										<p class="text-sm text-slate-600">
											Komponen ini dibuat dari sinkronisasi asesmen non-tes. Koreksi sumber nilai sebaiknya dilakukan dari modul asal lalu dikirim ulang ke nilai.
										</p>
									</div>
									<a
										href={nonTestSourceHref(selectedComponent)}
										class="inline-flex items-center justify-center rounded-md border border-emerald-200 bg-white px-3 py-2 text-sm font-medium text-emerald-800 hover:bg-emerald-50"
									>
										Buka Asesmen Asal
									</a>
								</div>
							</div>
						{/if}
						<div class="mx-6 mt-6 flex flex-col gap-3 rounded-xl border border-slate-200 bg-slate-50/80 px-4 py-4 lg:flex-row lg:items-center lg:justify-between">
							<div>
								<p class="text-xs font-semibold uppercase tracking-[0.2em] text-slate-500">Bulk Input</p>
								<p class="mt-2 text-lg font-semibold text-slate-900">{dirtyEntryIds.length} perubahan belum disimpan</p>
								<p class="text-sm text-slate-600">Guru dapat mengubah banyak nilai terlebih dahulu, lalu menyimpan seluruh perubahan untuk komponen ini sekaligus.</p>
							</div>
							<div class="flex flex-wrap gap-2">
								<Badge variant="outline">{entries.length} siswa</Badge>
								<Badge variant={dirtyEntryIds.length > 0 ? 'secondary' : 'outline'}>
									{dirtyEntryIds.length > 0 ? `${dirtyEntryIds.length} perlu disimpan` : 'Semua tersimpan'}
								</Badge>
								<LoadingButton
									loading={bulkSaveBusy}
									loadingLabel="Menyimpan semua..."
									disabled={dirtyEntryIds.length === 0}
									onclick={() => void saveAllDirtyEntries()}
									label="Simpan Semua Perubahan"
								/>
							</div>
						</div>
						<div class="mx-6 flex flex-col gap-3 rounded-xl border border-dashed border-emerald-200 bg-emerald-50/50 px-4 py-4">
							<div>
								<p class="text-xs font-semibold uppercase tracking-[0.2em] text-emerald-700">Quick Fill</p>
								<p class="mt-2 text-sm text-slate-700">Isi nilai atau catatan massal sebelum melakukan bulk save. Guru bisa menerapkan ke semua siswa atau hanya ke baris yang masih kosong.</p>
							</div>
							<div class="grid gap-3 lg:grid-cols-[12rem_1fr_auto]">
								<div>
									<label for="quick-fill-score" class="mb-1 block text-xs font-medium text-slate-500">Nilai Massal</label>
									<Input id="quick-fill-score" type="number" min="0" max={selectedComponent.max_score} step="0.1" bind:value={quickFillScore} />
								</div>
								<div>
									<label for="quick-fill-note" class="mb-1 block text-xs font-medium text-slate-500">Catatan Massal</label>
									<Input id="quick-fill-note" placeholder="Mis: remedial, observasi, atau catatan umum" bind:value={quickFillNote} />
								</div>
								<div class="flex items-end gap-2">
									<Button variant="outline" onclick={() => applyQuickFill('empty')}>Isi yang Kosong</Button>
									<Button variant="outline" onclick={() => applyQuickFill('all')}>Terapkan ke Semua</Button>
									<Button variant="ghost" onclick={resetQuickFill}>Reset</Button>
								</div>
							</div>
						</div>
					{/if}
					<div class="overflow-x-auto">
						<Table.Root>
							<Table.Header>
								<Table.Row>
									<Table.Head>Siswa</Table.Head>
									<Table.Head class="w-32">Nilai</Table.Head>
									<Table.Head>Catatan</Table.Head>
									<Table.Head class="w-28"></Table.Head>
								</Table.Row>
							</Table.Header>
							<Table.Body>
								{#if !selectedComponent}
									<Table.Row>
										<Table.Cell colspan={4} class="p-4">
											<EmptyStatePanel
												compact
												eyebrow="Siapkan Komponen"
												title="Pilih komponen penilaian"
												description="Pilih salah satu komponen di panel kiri agar lembar input nilai siswa terbuka."
											/>
										</Table.Cell>
									</Table.Row>
								{:else}
									{#each entries as row (row.student_id)}
										<Table.Row>
											<Table.Cell>
												<div class="font-medium text-slate-900">{row.nama}</div>
												<div class="text-xs text-slate-500">{row.nis || row.nisn || 'Tanpa NIS/NISN'}</div>
											</Table.Cell>
											<Table.Cell>
												<Input type="number" min="0" max={selectedComponent.max_score} step="0.1" bind:value={scoreInput[row.student_id]} />
											</Table.Cell>
											<Table.Cell>
												<Input placeholder="Catatan singkat" bind:value={noteInput[row.student_id]} />
											</Table.Cell>
											<Table.Cell class="text-right">
												<LoadingButton
													size="sm"
													loading={entryBusy[row.student_id]}
													loadingLabel="Menyimpan..."
													onclick={() => saveEntry(row.student_id)}
													label="Simpan"
												/>
											</Table.Cell>
										</Table.Row>
									{:else}
										<Table.Row>
											<Table.Cell colspan={4} class="p-4">
												<EmptyStatePanel
													compact
													eyebrow="Kelas Masih Kosong"
													title="Belum ada siswa aktif di kelas ini"
													description="Tambahkan atau aktifkan siswa pada kelas terkait supaya lembar input nilai bisa digunakan."
												/>
											</Table.Cell>
										</Table.Row>
									{/each}
								{/if}
							</Table.Body>
						</Table.Root>
					</div>
				</Card.Content>
			</Card.Root>
			{/if}
		{:else if gradeWorkspace !== 'triage'}
			<Card.Root class="border-dashed border-slate-300 bg-white">
				<Card.Content class="py-10 text-center">
						<EmptyStatePanel
							eyebrow="Buka Gradebook"
							title="Pilih penugasan kelas-mapel untuk membuka gradebook"
							description="Setelah konteks kelas dan mapel dipilih, komponen nilai, rekap sementara, dan lembar input siswa akan muncul dalam satu alur kerja."
						>
							{#if assignments.length > 0}
								<Button size="sm" onclick={quickSelectFirstAssignment}>Pilih penugasan pertama</Button>
							{/if}
						</EmptyStatePanel>
				</Card.Content>
			</Card.Root>
		{/if}
		{/snippet}
	</AsyncContent>
</div>
