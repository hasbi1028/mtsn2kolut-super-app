<script lang="ts">
	import { onMount } from 'svelte';
	import { RefreshCw, Search } from '@lucide/svelte';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { toast } from '$lib/components/ui/sonner';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import DirtyChangeBar from '$lib/components/editable/DirtyChangeBar.svelte';
	import EditableSelectCell from '$lib/components/editable/EditableSelectCell.svelte';
	import EmptyStatePanel from '$lib/components/EmptyStatePanel.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { academicCopy } from '$lib/academic/copy';
	import { bindBeforeUnload, confirmDiscardChanges } from '$lib/client/unsaved-changes';
	import { readClientApiData, readClientJson } from '$lib/client/api';
	import type { EditableOption } from '$lib/components/editable/types';

	type MatrixClass = {
		id: string;
		code: string;
		name: string;
		level: string;
	};

	type MatrixSubject = {
		id: string;
		code: string;
		name: string;
		category: string;
		is_assessment_subject: boolean;
		is_report_subject: boolean;
		is_schedule_activity: boolean;
		default_weekly_hours: number;
		display_order: number;
	};

	type MatrixTeacher = {
		id: string;
		nip: string;
		nama: string;
		unit_kerja: string;
	};

	type MatrixCell = {
		class_id: string;
		subject_id: string;
		assignment_id: string | null;
		teacher_employee_id: string | null;
		teacher_name: string;
		status: 'complete' | 'missing_teacher' | 'missing_assignment';
	};

	type SubjectAssignmentMatrix = {
		academic_year_id: string;
		academic_year_name: string;
		classes: MatrixClass[];
		subjects: MatrixSubject[];
		teachers: MatrixTeacher[];
		cells: MatrixCell[];
	};

	type DirtyCell = {
		key: string;
		class_id: string;
		subject_id: string;
		teacher_employee_id: string;
		subject_code: string;
		class_code: string;
	};

	let matrix = $state<SubjectAssignmentMatrix | null>(null);
	let matrixPromise = $state<Promise<SubjectAssignmentMatrix> | null>(null);
	let cellDrafts = $state<Record<string, string>>({});
	let cellErrors = $state<Record<string, string>>({});
	let refreshBusy = $state(false);
	let saveBusy = $state(false);
	let savingCellKey = $state('');
	let levelFilter = $state('');
	let subjectSearch = $state('');
	let requestId = 0;

	const teacherOptions = $derived.by<EditableOption[]>(() =>
		(matrix?.teachers ?? []).map((teacher) => ({
			value: teacher.id,
			label: teacher.nip ? `${teacher.nama} (${teacher.nip})` : teacher.nama,
			description: teacher.unit_kerja,
		}))
	);
	const levelOptions = $derived.by(() => Array.from(new Set((matrix?.classes ?? []).map((item) => item.level))).sort());
	const visibleClasses = $derived.by(() => (matrix?.classes ?? []).filter((item) => !levelFilter || item.level === levelFilter));
	const visibleSubjects = $derived.by(() => {
		const query = subjectSearch.trim().toLowerCase();
		return (matrix?.subjects ?? []).filter((subject) => {
			if (!query) return true;
			return [subject.code, subject.name, subject.category].some((value) => value.toLowerCase().includes(query));
		});
	});
	const cellByKey = $derived.by(() => {
		const items = new Map<string, MatrixCell>();
		for (const cell of matrix?.cells ?? []) {
			items.set(cellKey(cell.class_id, cell.subject_id), cell);
		}
		return items;
	});
	const dirtyCells = $derived.by(() => {
		if (!matrix) return [] as DirtyCell[];
		const rows: DirtyCell[] = [];
		for (const subject of matrix.subjects) {
			for (const klass of matrix.classes) {
				const key = cellKey(klass.id, subject.id);
				const draft = getCellDraft(klass.id, subject.id);
				if (draft !== currentTeacherId(klass.id, subject.id)) {
					rows.push({
						key,
						class_id: klass.id,
						subject_id: subject.id,
						teacher_employee_id: draft,
						subject_code: subject.code,
						class_code: klass.code,
					});
				}
			}
		}
		return rows;
	});
	const dirtyCount = $derived(dirtyCells.length);
	const summary = $derived.by(() => {
		const cells = matrix?.cells ?? [];
		return {
			complete: cells.filter((cell) => cell.status === 'complete').length,
			missing: cells.filter((cell) => cell.status !== 'complete').length,
			teachers: matrix?.teachers.length ?? 0,
		};
	});

	async function fetchMatrix(): Promise<SubjectAssignmentMatrix> {
		return await fetch('/api/academic/subject-assignment-matrix').then((response) =>
			readClientApiData<SubjectAssignmentMatrix>(response, 'Gagal memuat tabel penugasan guru mapel')
		);
	}

	function cellKey(classId: string, subjectId: string) {
		return `${classId}:${subjectId}`;
	}

	function currentTeacherId(classId: string, subjectId: string) {
		return cellByKey.get(cellKey(classId, subjectId))?.teacher_employee_id ?? '';
	}

	function getCellDraft(classId: string, subjectId: string) {
		const key = cellKey(classId, subjectId);
		if (key in cellDrafts) return cellDrafts[key];
		return currentTeacherId(classId, subjectId);
	}

	function statusFor(classId: string, subjectId: string) {
		return cellByKey.get(cellKey(classId, subjectId))?.status ?? 'missing_assignment';
	}

	function teacherNameFor(classId: string, subjectId: string) {
		return cellByKey.get(cellKey(classId, subjectId))?.teacher_name ?? '';
	}

	function resetCellDrafts(source = matrix) {
		cellDrafts = Object.fromEntries((source?.cells ?? []).map((cell) => [cellKey(cell.class_id, cell.subject_id), cell.teacher_employee_id ?? '']));
		cellErrors = {};
	}

	function applyMatrix(payload: SubjectAssignmentMatrix) {
		matrix = payload;
		resetCellDrafts(payload);
		return payload;
	}

	function loadMatrix() {
		const current = ++requestId;
		matrixPromise = fetchMatrix().then((payload) => {
			if (current === requestId) applyMatrix(payload);
			return payload;
		});
	}

	async function refreshMatrix() {
		if (dirtyCount > 0 && !(await confirmDiscardChanges(true))) return;
		refreshBusy = true;
		try {
			loadMatrix();
			await matrixPromise;
		} finally {
			refreshBusy = false;
		}
	}

	function updateCellDraft(classId: string, subjectId: string, value: string) {
		const key = cellKey(classId, subjectId);
		cellDrafts = { ...cellDrafts, [key]: value };
		if (cellErrors[key]) {
			const { [key]: _ignored, ...rest } = cellErrors;
			cellErrors = rest;
		}
	}

	function markCellError(key: string, error: string) {
		cellErrors = { ...cellErrors, [key]: error };
	}

	function mergeCell(updated: MatrixCell) {
		if (!matrix) return;
		const key = cellKey(updated.class_id, updated.subject_id);
		const exists = matrix.cells.some((cell) => cellKey(cell.class_id, cell.subject_id) === key);
		const cells = exists
			? matrix.cells.map((cell) => (cellKey(cell.class_id, cell.subject_id) === key ? updated : cell))
			: [...matrix.cells, updated];
		matrix = { ...matrix, cells };
		cellDrafts = { ...cellDrafts, [key]: updated.teacher_employee_id ?? '' };
		const { [key]: _ignored, ...rest } = cellErrors;
		cellErrors = rest;
	}

	async function saveChangedCells() {
		const rows = [...dirtyCells];
		saveBusy = true;
		try {
			for (const cell of rows) {
				savingCellKey = cell.key;
				const response = await fetch('/api/academic/subject-assignment-matrix', {
					method: 'PUT',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify({
						class_id: cell.class_id,
						subject_id: cell.subject_id,
						teacher_employee_id: cell.teacher_employee_id,
					}),
				});
				try {
					const updated = await readClientJson<MatrixCell>(response);
					mergeCell(updated);
				} catch (error) {
					const message = error instanceof Error && error.message.trim() ? error.message : 'Gagal menyimpan guru mapel.';
					markCellError(cell.key, message);
					toast.error(`${cell.subject_code} ${cell.class_code}: ${message}`);
					return;
				}
			}
			toast.success(`${rows.length} perubahan guru mapel tersimpan`);
		} finally {
			saveBusy = false;
			savingCellKey = '';
		}
	}

	function statusBadge(status: MatrixCell['status']) {
		if (status === 'complete') return 'Lengkap';
		if (status === 'missing_teacher') return 'Guru perlu diperiksa';
		return 'Belum ada';
	}

	function matrixErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Tabel penugasan guru mapel belum dapat dimuat. Periksa koneksi layanan sistem lalu coba lagi.';
	}

	function handleRenderError(error: unknown) {
		console.error('Subject assignment matrix render failed', error);
	}

	onMount(() => {
		loadMatrix();
		return bindBeforeUnload(() => dirtyCount > 0);
	});
</script>

<svelte:head>
	<title>Penugasan Guru Mapel | MTsN 2 Kolaka Utara</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex flex-col gap-3 md:flex-row md:items-end md:justify-between">
		<div>
			<p class="text-sm font-medium text-primary">Akademik</p>
			<h1 class="text-2xl font-semibold tracking-normal text-foreground">Penugasan Guru Mapel</h1>
			<p class="mt-1 max-w-2xl text-sm text-muted-foreground">
				Tabel mapel dan rombel tahun ajaran aktif untuk menetapkan guru pengampu.
			</p>
		</div>
		<Button variant="outline" onclick={() => void refreshMatrix()} disabled={refreshBusy || saveBusy} aria-label="Muat ulang penugasan guru mapel">
			<RefreshCw class={`mr-2 size-4 ${refreshBusy ? 'animate-spin' : ''}`} />
			Muat Ulang
		</Button>
	</div>

	<AsyncContent promise={matrixPromise} onerror={handleRenderError}>
		{#snippet pending()}
			<div class="space-y-4">
				<div class="grid gap-3 md:grid-cols-3">
					<Skeleton class="h-24 w-full" />
					<Skeleton class="h-24 w-full" />
					<Skeleton class="h-24 w-full" />
				</div>
				<Skeleton class="h-[28rem] w-full" />
			</div>
		{/snippet}

		{#snippet failed(error, reset)}
			<RecoveryPanel
				title="Penugasan Guru Mapel Belum Tersaji"
				message={matrixErrorMessage(error)}
				onRetry={() => {
					reset?.();
					loadMatrix();
				}}
			/>
		{/snippet}

		<div class="grid gap-3 md:grid-cols-3">
			<Card.Root>
				<Card.Header class="pb-2">
					<Card.Description>{academicCopy.labels.activeAcademicYear}</Card.Description>
					<Card.Title class="text-xl">{matrix?.academic_year_name || 'Belum ada'}</Card.Title>
				</Card.Header>
			</Card.Root>
			<Card.Root>
				<Card.Header class="pb-2">
					<Card.Description>Penugasan lengkap</Card.Description>
					<Card.Title class="text-2xl">{summary.complete}</Card.Title>
				</Card.Header>
			</Card.Root>
			<Card.Root>
				<Card.Header class="pb-2">
					<Card.Description>Perlu dilengkapi</Card.Description>
					<Card.Title class="text-2xl">{summary.missing}</Card.Title>
				</Card.Header>
			</Card.Root>
		</div>

		<Card.Root>
			<Card.Header>
				<Card.Title class="text-base">{academicCopy.labels.teacherAssignmentTable}</Card.Title>
				<Card.Description>{summary.teachers} guru aktif tersedia sebagai pilihan pengampu.</Card.Description>
			</Card.Header>
			<Card.Content class="space-y-4">
				<div class="grid gap-3 md:grid-cols-[12rem_1fr]">
					<div class="space-y-1">
						<label for="matrix-level-filter" class="text-sm font-medium">{academicCopy.labels.classLevel}</label>
						<select
							id="matrix-level-filter"
							class="h-10 w-full rounded-md border border-input bg-background px-3 text-sm text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
							bind:value={levelFilter}
							disabled={saveBusy}
						>
							<option value="">Semua tingkat</option>
							{#each levelOptions as level (level)}
								<option value={level}>{level}</option>
							{/each}
						</select>
					</div>
					<div class="space-y-1">
						<label for="matrix-subject-search" class="text-sm font-medium">Cari mapel</label>
						<div class="relative">
							<Search class="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
							<Input id="matrix-subject-search" class="pl-9" bind:value={subjectSearch} placeholder="Nama, kode, atau kategori" disabled={saveBusy} />
						</div>
					</div>
				</div>

				{#if !matrix?.academic_year_id}
					<EmptyStatePanel compact title="Tahun ajaran aktif belum ada" description="Aktifkan tahun ajaran sebelum mengatur guru mapel." />
				{:else if visibleClasses.length === 0 || visibleSubjects.length === 0}
					<EmptyStatePanel compact title="Tabel penugasan masih kosong" description="Pastikan rombel dan mapel aktif tersedia untuk tahun ajaran berjalan." />
				{:else}
					<div class="overflow-x-auto rounded-md border">
						<Table.Root>
							<Table.Header>
								<Table.Row>
									<Table.Head class="min-w-56 bg-muted/40">Mapel</Table.Head>
									{#each visibleClasses as klass (klass.id)}
										<Table.Head class="min-w-64 bg-muted/40">
											<div class="space-y-0.5">
												<p class="font-medium text-foreground">{klass.code}</p>
												<p class="text-xs font-normal text-muted-foreground">{klass.name}</p>
											</div>
										</Table.Head>
									{/each}
								</Table.Row>
							</Table.Header>
							<Table.Body>
								{#each visibleSubjects as subject (subject.id)}
									<Table.Row>
										<Table.Cell class="align-top">
											<div class="space-y-1">
												<p class="font-medium text-foreground">{subject.name}</p>
												<p class="text-xs text-muted-foreground">{subject.code} · {subject.category}</p>
												<div class="flex flex-wrap gap-1">
													{#if subject.is_assessment_subject}
														<Badge class="border-primary/20 bg-primary/15 text-primary">Asesmen</Badge>
													{/if}
													{#if subject.is_report_subject}
														<Badge variant="secondary">Rapor</Badge>
													{/if}
													{#if subject.is_schedule_activity}
														<Badge variant="outline">Aktivitas jadwal</Badge>
													{/if}
												</div>
											</div>
										</Table.Cell>
										{#each visibleClasses as klass (klass.id)}
											{@const key = cellKey(klass.id, subject.id)}
											{@const status = statusFor(klass.id, subject.id)}
											{@const draft = getCellDraft(klass.id, subject.id)}
											{@const current = currentTeacherId(klass.id, subject.id)}
											<Table.Cell class={`align-top ${draft !== current ? 'bg-primary/5' : ''}`}>
												<div class="space-y-1">
													<EditableSelectCell
														value={draft}
														options={teacherOptions}
														emptyLabel="Belum ada"
														placeholder="Belum ada"
														ariaLabel={`Pilih guru ${subject.name} ${klass.code}`}
														dirty={draft !== current}
														loading={savingCellKey === key}
														disabled={saveBusy && savingCellKey !== key}
														oninput={(value) => updateCellDraft(klass.id, subject.id, value)}
														oncancel={(event) => updateCellDraft(klass.id, subject.id, event.originalValue)}
													/>
													<div class="flex flex-wrap items-center gap-1">
														{#if status === 'complete'}
															<Badge class="border-primary/20 bg-primary/15 text-primary">{statusBadge(status)}</Badge>
														{:else if status === 'missing_teacher'}
															<Badge variant="destructive">{statusBadge(status)}</Badge>
														{:else}
															<Badge variant="secondary">{statusBadge(status)}</Badge>
														{/if}
														{#if teacherNameFor(klass.id, subject.id)}
															<span class="text-xs text-muted-foreground">{teacherNameFor(klass.id, subject.id)}</span>
														{/if}
													</div>
													{#if cellErrors[key]}
														<p class="max-w-56 text-xs text-destructive">{cellErrors[key]}</p>
													{/if}
												</div>
											</Table.Cell>
										{/each}
									</Table.Row>
								{/each}
							</Table.Body>
						</Table.Root>
					</div>
				{/if}
			</Card.Content>
		</Card.Root>

		<DirtyChangeBar
			count={dirtyCount}
			saving={saveBusy}
			saveLabel={saveBusy ? 'Menyimpan...' : academicCopy.actions.saveAll}
			discardLabel={academicCopy.actions.cancel}
			onsave={() => void saveChangedCells()}
			ondiscard={() => resetCellDrafts()}
		/>
	</AsyncContent>
</div>
