<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import { ArrowLeft, Pencil, RefreshCw, Search, Trash2, UserCheck } from '@lucide/svelte';
	import * as Card from '$lib/components/ui/card';
	import * as Tabs from '$lib/components/ui/tabs';
	import * as Table from '$lib/components/ui/table';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { Textarea } from '$lib/components/ui/textarea';
	import { toast } from '$lib/components/ui/sonner';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import EmptyStatePanel from '$lib/components/EmptyStatePanel.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { readClientApiData, readClientJson } from '$lib/client/api';
	import { confirmAction } from '$lib/confirm-dialog';
	import {
		buildHomeroomEmployeeOptions,
		employeeOptionSubtitle,
		employmentTypeLabel,
		filterHomeroomEmployeeOptions,
		type HomeroomEmployeeOption
	} from '$lib/client/rombel-homeroom';
	import {
		availableRombelSubjectOptions,
		buildRombelSubjectOptions,
		subjectOptionLabel
	} from '$lib/client/rombel-subject-assignments';

	type RombelDetail = {
		id: string;
		code: string;
		name: string;
		level: string;
		is_active: boolean;
		academic_year_name: string;
		homeroom_assignment_id: string | null;
		homeroom_teacher_name: string;
		homeroom_notes: string;
		total_students: number;
		total_linked_parents: number;
		total_subject_teachers: number;
		total_subject_assignments: number;
		total_timetable_slots: number;
	};

	type StudentParent = {
		id: string;
		nama: string;
		phone: string;
		address: string;
		occupation: string;
		income_band: string;
		relationship: string;
		is_primary_contact: boolean;
		notes: string;
	};

	type RombelStudent = {
		id: string;
		nis: string;
		nisn: string;
		nama: string;
		gender: string;
		parent_name: string;
		parent_phone: string;
		student_phone: string;
		student_address: string;
		is_active: boolean;
		status: string;
		parents: StudentParent[];
	};

	type SubjectAssignment = {
		id: string;
		subject_id: string;
		subject_name: string;
		subject_code: string;
		teacher_employee_id: string;
		teacher_name: string;
	};

	type TimetableSlot = {
		id: string;
		day_of_week: number;
		start_time: string;
		end_time: string;
		room_label: string;
		notes: string;
		subject_name: string;
		subject_code: string;
		teacher_name: string;
	};

	type HomeroomAssignment = {
		id: string;
		employee_id: string;
		employee_name: string;
		academic_year_name: string;
		start_date: string | null;
		end_date: string | null;
		is_active: boolean;
		notes: string;
	};

	type EmployeeRow = {
		id: string;
		pegawai_uid: string;
		nip: string;
		nama: string;
		employment_type: string;
		is_active: boolean;
	};

	type SubjectRow = {
		id: string;
		code: string;
		name: string;
		is_active: boolean;
	};

	type AcademicSubjectsPayload = {
		subjects?: SubjectRow[];
	};

	type RombelDetailPayload = {
		rombel: RombelDetail;
		students: RombelStudent[];
		subject_assignments: SubjectAssignment[];
		timetable_slots: TimetableSlot[];
		homeroom_assignments: HomeroomAssignment[];
	};

	const classId = $derived(page.params.id ?? '');
	const dayLabels: Record<number, string> = {
		1: 'Senin',
		2: 'Selasa',
		3: 'Rabu',
		4: 'Kamis',
		5: 'Jumat',
		6: 'Sabtu',
	};

	let detail = $state<RombelDetailPayload | null>(null);
	let detailPromise = $state<Promise<RombelDetailPayload> | null>(null);
	let requestId = 0;
	let employeeRequestId = 0;
	let refreshBusy = $state(false);
	let saveBusy = $state(false);
	let homeroomEmployeeId = $state('');
	let homeroomNotes = $state('');
	let employeeRows = $state<EmployeeRow[]>([]);
	let employeesLoading = $state(false);
	let employeesLoaded = $state(false);
	let employeesError = $state('');
	let employeeSearch = $state('');
	let subjectRows = $state<SubjectRow[]>([]);
	let subjectsLoading = $state(false);
	let subjectsLoaded = $state(false);
	let subjectsError = $state('');
	let subjectRequestId = 0;
	let subjectAssignmentSubjectId = $state('');
	let subjectAssignmentTeacherId = $state('');
	let editingSubjectAssignmentId = $state('');
	let subjectTeacherSearch = $state('');
	let subjectAssignmentSaveBusy = $state(false);
	let deleteSubjectAssignmentBusyId = $state('');

	const rombel = $derived(detail?.rombel ?? null);
	const activeHomeroom = $derived(detail?.homeroom_assignments.find((item) => item.is_active) ?? null);
	const employeeOptions = $derived(buildHomeroomEmployeeOptions({
		employees: employeeRows,
		subjectAssignments: detail?.subject_assignments ?? [],
		activeHomeroom
	}));
	const filteredEmployeeOptions = $derived(filterHomeroomEmployeeOptions(employeeOptions, employeeSearch));
	const selectedEmployeeOption = $derived(employeeOptions.find((option) => option.id === homeroomEmployeeId) ?? null);
	const canSaveHomeroom = $derived(Boolean(classId && selectedEmployeeOption));
	const subjectOptions = $derived(buildRombelSubjectOptions({
		subjects: subjectRows,
		assignments: detail?.subject_assignments ?? []
	}));
	const availableSubjectOptions = $derived(availableRombelSubjectOptions(
		subjectOptions,
		detail?.subject_assignments ?? [],
		editingSubjectAssignmentId
	));
	const selectedSubjectOption = $derived(subjectOptions.find((option) => option.id === subjectAssignmentSubjectId) ?? null);
	const filteredSubjectTeacherOptions = $derived(filterHomeroomEmployeeOptions(employeeOptions, subjectTeacherSearch));
	const selectedSubjectTeacherOption = $derived(employeeOptions.find((option) => option.id === subjectAssignmentTeacherId) ?? null);
	const canSaveSubjectAssignment = $derived(Boolean(classId && selectedSubjectOption && selectedSubjectTeacherOption));
	const parentRows = $derived.by(() =>
		(detail?.students ?? []).flatMap((student) =>
			student.parents.map((parent) => ({ student, parent }))
		)
	);

	async function fetchDetail(): Promise<RombelDetailPayload> {
		return await fetch(`/api/academic/rombel/${classId}`).then((response) =>
			readClientApiData<RombelDetailPayload>(response, 'Gagal memuat detail rombel')
		);
	}

	async function fetchEmployees(): Promise<EmployeeRow[]> {
		return await fetch('/api/employees').then((response) =>
			readClientApiData<EmployeeRow[]>(response, 'Gagal memuat daftar pegawai aktif')
		);
	}

	async function fetchSubjects(): Promise<SubjectRow[]> {
		const payload = await fetch('/api/academic').then((response) =>
			readClientApiData<AcademicSubjectsPayload>(response, 'Gagal memuat daftar mata pelajaran')
		);
		return Array.isArray(payload.subjects) ? payload.subjects : [];
	}

	function applyDetail(payload: RombelDetailPayload) {
		detail = payload;
		const active = payload.homeroom_assignments.find((item) => item.is_active);
		homeroomEmployeeId = active?.employee_id ?? '';
		homeroomNotes = active?.notes ?? '';
		if (editingSubjectAssignmentId && !payload.subject_assignments.some((item) => item.id === editingSubjectAssignmentId)) {
			resetSubjectAssignmentForm();
		}
	}

	function loadDetail() {
		if (!classId) return;
		const current = ++requestId;
		detailPromise = fetchDetail().then((payload) => {
			if (current === requestId) {
				applyDetail(payload);
			}
			return payload;
		});
	}

	async function loadEmployees() {
		const current = ++employeeRequestId;
		employeesLoading = true;
		employeesError = '';
		try {
			const items = await fetchEmployees();
			if (current === employeeRequestId) {
				employeeRows = items;
				employeesLoaded = true;
			}
		} catch (error) {
			if (current === employeeRequestId) {
				employeesError = errorMessage(error, 'Daftar pegawai aktif belum dapat dimuat.');
			}
		} finally {
			if (current === employeeRequestId) {
				employeesLoading = false;
			}
		}
	}

	async function loadSubjects() {
		const current = ++subjectRequestId;
		subjectsLoading = true;
		subjectsError = '';
		try {
			const items = await fetchSubjects();
			if (current === subjectRequestId) {
				subjectRows = items;
				subjectsLoaded = true;
			}
		} catch (error) {
			if (current === subjectRequestId) {
				subjectsError = errorMessage(error, 'Daftar mata pelajaran belum dapat dimuat.');
			}
		} finally {
			if (current === subjectRequestId) {
				subjectsLoading = false;
			}
		}
	}

	async function refreshDetail() {
		refreshBusy = true;
		try {
			loadDetail();
			const employeesPromise = loadEmployees();
			const subjectsPromise = loadSubjects();
			await detailPromise;
			await employeesPromise;
			await subjectsPromise;
		} finally {
			refreshBusy = false;
		}
	}

	async function saveHomeroom() {
		if (!canSaveHomeroom) return;
		saveBusy = true;
		try {
			const body = JSON.stringify({
				employee_id: homeroomEmployeeId,
				is_active: true,
				notes: homeroomNotes,
			});
			const active = activeHomeroom;
			const response = await fetch(
				active
					? `/api/academic/rombel/${classId}/homeroom-assignments/${active.id}`
					: `/api/academic/rombel/${classId}/homeroom-assignments`,
				{
					method: active ? 'PUT' : 'POST',
					headers: { 'Content-Type': 'application/json' },
					body,
				}
			);
			await readClientJson<unknown>(response);
			toast.success('Wali kelas tersimpan');
			await refreshDetail();
		} catch (error) {
			toast.error(errorMessage(error, 'Gagal menyimpan wali kelas'));
		} finally {
			saveBusy = false;
		}
	}

	function resetSubjectAssignmentForm() {
		editingSubjectAssignmentId = '';
		subjectAssignmentSubjectId = '';
		subjectAssignmentTeacherId = '';
		subjectTeacherSearch = '';
	}

	function editSubjectAssignment(assignment: SubjectAssignment) {
		editingSubjectAssignmentId = assignment.id;
		subjectAssignmentSubjectId = assignment.subject_id;
		subjectAssignmentTeacherId = assignment.teacher_employee_id;
		subjectTeacherSearch = '';
	}

	async function saveSubjectAssignment() {
		if (!canSaveSubjectAssignment) return;
		subjectAssignmentSaveBusy = true;
		try {
			const body = JSON.stringify({
				subject_id: subjectAssignmentSubjectId,
				teacher_employee_id: subjectAssignmentTeacherId,
			});
			const response = await fetch(
				editingSubjectAssignmentId
					? `/api/academic/rombel/${classId}/subject-assignments/${editingSubjectAssignmentId}`
					: `/api/academic/rombel/${classId}/subject-assignments`,
				{
					method: editingSubjectAssignmentId ? 'PUT' : 'POST',
					headers: { 'Content-Type': 'application/json' },
					body,
				}
			);
			await readClientJson<unknown>(response);
			toast.success(editingSubjectAssignmentId ? 'Guru mapel diperbarui' : 'Guru mapel ditambahkan');
			resetSubjectAssignmentForm();
			await refreshDetail();
		} catch (error) {
			toast.error(errorMessage(error, 'Gagal menyimpan guru mapel'));
		} finally {
			subjectAssignmentSaveBusy = false;
		}
	}

	async function deleteSubjectAssignment(assignment: SubjectAssignment) {
		if (!(await confirmAction({
			title: 'Hapus Guru Mapel',
			message: `Hapus penugasan ${assignment.subject_name} - ${assignment.teacher_name}?`,
			confirmLabel: 'Hapus',
			tone: 'danger'
		}))) return;
		deleteSubjectAssignmentBusyId = assignment.id;
		try {
			const response = await fetch(`/api/academic/rombel/${classId}/subject-assignments/${assignment.id}`, { method: 'DELETE' });
			await readClientJson<unknown>(response);
			toast.success('Guru mapel dihapus');
			if (editingSubjectAssignmentId === assignment.id) {
				resetSubjectAssignmentForm();
			}
			await refreshDetail();
		} catch (error) {
			toast.error(errorMessage(error, 'Gagal menghapus guru mapel'));
		} finally {
			deleteSubjectAssignmentBusyId = '';
		}
	}

	function errorMessage(error: unknown, fallback: string) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return fallback;
	}

	function fmtDate(value: string | null | undefined) {
		return value ? value.slice(0, 10) : '-';
	}

	function fmtTime(value: string) {
		return value.slice(0, 5);
	}

	function genderLabel(value: string) {
		return value === 'female' ? 'P' : 'L';
	}

	function relationshipLabel(value: string) {
		const labels: Record<string, string> = { ayah: 'Ayah', ibu: 'Ibu', wali: 'Wali', lainnya: 'Lainnya' };
		return labels[value] ?? (value || '-');
	}

	function employeeSourceLabel(option: HomeroomEmployeeOption) {
		return option.isActiveEmployee ? 'Pegawai aktif' : 'Fallback rombel';
	}

	function selectHomeroomEmployee(option: HomeroomEmployeeOption) {
		homeroomEmployeeId = option.id;
	}

	function handleRenderError(error: unknown) {
		console.error('Rombel detail render failed', error);
	}

	onMount(() => {
		loadDetail();
		void loadEmployees();
		void loadSubjects();
	});
</script>

<svelte:head>
	<title>{rombel?.name ?? 'Rombel'} | MTsN 2 Kolaka Utara</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex flex-col gap-3 md:flex-row md:items-end md:justify-between">
		<div class="space-y-2">
			<Button variant="ghost" size="sm" href={resolve('/akademik/rombel' as '/')}>
				<ArrowLeft class="mr-2 size-4" />
				Kembali
			</Button>
			<div>
				<p class="text-sm font-medium text-primary">Rombel</p>
				<h1 class="text-2xl font-semibold tracking-normal text-foreground">{rombel?.name ?? 'Memuat rombel'}</h1>
				<p class="mt-1 text-sm text-muted-foreground">
					{rombel ? `${rombel.code} · Tahun ajaran ${rombel.academic_year_name}` : 'Detail kelas, siswa, wali, mapel, dan jadwal.'}
				</p>
			</div>
		</div>
		<Button variant="outline" onclick={() => void refreshDetail()} disabled={refreshBusy} aria-label="Muat ulang detail rombel">
			<RefreshCw class={`mr-2 size-4 ${refreshBusy ? 'animate-spin' : ''}`} />
			Muat Ulang
		</Button>
	</div>

	<AsyncContent promise={detailPromise} onerror={handleRenderError}>
		{#snippet pending()}
			<div class="space-y-4">
				<div class="grid gap-3 md:grid-cols-5">
					<Skeleton class="h-20 w-full" />
					<Skeleton class="h-20 w-full" />
					<Skeleton class="h-20 w-full" />
					<Skeleton class="h-20 w-full" />
					<Skeleton class="h-20 w-full" />
				</div>
				<Skeleton class="h-96 w-full" />
			</div>
		{/snippet}

		{#snippet failed(error, reset)}
			<RecoveryPanel
				title="Detail Rombel Belum Tersaji"
				message={errorMessage(error, 'Detail rombel belum dapat dimuat.')}
				onRetry={() => {
					reset?.();
					loadDetail();
				}}
			/>
		{/snippet}

		{#if detail && rombel}
			<div class="grid gap-3 md:grid-cols-5">
				<Card.Root>
					<Card.Header class="pb-2">
						<Card.Description>Siswa</Card.Description>
						<Card.Title class="text-2xl">{rombel.total_students}</Card.Title>
					</Card.Header>
				</Card.Root>
				<Card.Root>
					<Card.Header class="pb-2">
						<Card.Description>Orang Tua Terhubung</Card.Description>
						<Card.Title class="text-2xl">{rombel.total_linked_parents}</Card.Title>
					</Card.Header>
				</Card.Root>
				<Card.Root>
					<Card.Header class="pb-2">
						<Card.Description>Guru Mapel</Card.Description>
						<Card.Title class="text-2xl">{rombel.total_subject_teachers}</Card.Title>
					</Card.Header>
				</Card.Root>
				<Card.Root>
					<Card.Header class="pb-2">
						<Card.Description>Assignment</Card.Description>
						<Card.Title class="text-2xl">{rombel.total_subject_assignments}</Card.Title>
					</Card.Header>
				</Card.Root>
				<Card.Root>
					<Card.Header class="pb-2">
						<Card.Description>Slot Jadwal</Card.Description>
						<Card.Title class="text-2xl">{rombel.total_timetable_slots}</Card.Title>
					</Card.Header>
				</Card.Root>
			</div>

			<Tabs.Root value="students">
				<div class="overflow-x-auto pb-1">
					<Tabs.List class="mb-4 min-w-max">
						<Tabs.Trigger value="students">Siswa ({detail.students.length})</Tabs.Trigger>
						<Tabs.Trigger value="parents">Orang Tua ({parentRows.length})</Tabs.Trigger>
						<Tabs.Trigger value="homeroom">Wali Kelas</Tabs.Trigger>
						<Tabs.Trigger value="teachers">Guru Mapel ({detail.subject_assignments.length})</Tabs.Trigger>
						<Tabs.Trigger value="schedule">Jadwal ({detail.timetable_slots.length})</Tabs.Trigger>
					</Tabs.List>
				</div>

				<Tabs.Content value="students">
					<Card.Root>
						<Card.Header>
							<Card.Title class="text-base">Siswa Rombel</Card.Title>
							<Card.Description>Daftar siswa aktif dan data orang tua lama dari master siswa.</Card.Description>
						</Card.Header>
						<Card.Content class="p-0">
							<div class="overflow-x-auto">
								<Table.Root>
									<Table.Header>
										<Table.Row>
											<Table.Head>Siswa</Table.Head>
											<Table.Head>NISN</Table.Head>
											<Table.Head>JK</Table.Head>
											<Table.Head>Orang Tua Legacy</Table.Head>
											<Table.Head class="text-right">Link Orang Tua</Table.Head>
											<Table.Head>Status</Table.Head>
										</Table.Row>
									</Table.Header>
									<Table.Body>
										{#each detail.students as student (student.id)}
											<Table.Row>
												<Table.Cell>
													<div class="space-y-0.5">
														<p class="font-medium">{student.nama}</p>
														<p class="text-xs text-muted-foreground">NIS {student.nis}</p>
													</div>
												</Table.Cell>
												<Table.Cell class="text-muted-foreground">{student.nisn || '-'}</Table.Cell>
												<Table.Cell>{genderLabel(student.gender)}</Table.Cell>
												<Table.Cell class="max-w-[18rem] text-muted-foreground">{student.parent_name || '-'}</Table.Cell>
												<Table.Cell class="text-right tabular-nums">{student.parents.length}</Table.Cell>
												<Table.Cell>
													<Badge variant={student.is_active ? 'default' : 'secondary'}>{student.status}</Badge>
												</Table.Cell>
											</Table.Row>
										{:else}
											<Table.Row>
												<Table.Cell colspan={6} class="p-4">
													<EmptyStatePanel compact title="Belum ada siswa di rombel ini" description="Siswa akan tampil setelah terhubung dengan kelas ini." />
												</Table.Cell>
											</Table.Row>
										{/each}
									</Table.Body>
								</Table.Root>
							</div>
						</Card.Content>
					</Card.Root>
				</Tabs.Content>

				<Tabs.Content value="parents">
					<Card.Root>
						<Card.Header>
							<Card.Title class="text-base">Orang Tua Terhubung</Card.Title>
							<Card.Description>Relasi ayah, ibu, wali, atau lainnya yang sudah dinormalisasi ke tabel orang tua.</Card.Description>
						</Card.Header>
						<Card.Content class="p-0">
							<div class="overflow-x-auto">
								<Table.Root>
									<Table.Header>
										<Table.Row>
											<Table.Head>Siswa</Table.Head>
											<Table.Head>Nama Orang Tua</Table.Head>
											<Table.Head>Relasi</Table.Head>
											<Table.Head>Kontak</Table.Head>
											<Table.Head>Pekerjaan</Table.Head>
											<Table.Head>Utama</Table.Head>
										</Table.Row>
									</Table.Header>
									<Table.Body>
										{#each parentRows as row (`${row.student.id}-${row.parent.id}`)}
											<Table.Row>
												<Table.Cell class="font-medium">{row.student.nama}</Table.Cell>
												<Table.Cell>{row.parent.nama}</Table.Cell>
												<Table.Cell>{relationshipLabel(row.parent.relationship)}</Table.Cell>
												<Table.Cell class="text-muted-foreground">{row.parent.phone || row.parent.address || '-'}</Table.Cell>
												<Table.Cell class="text-muted-foreground">{row.parent.occupation || '-'}</Table.Cell>
												<Table.Cell>
													{#if row.parent.is_primary_contact}
														<Badge class="border-primary/20 bg-primary/15 text-primary">Kontak utama</Badge>
													{:else}
														<span class="text-muted-foreground">-</span>
													{/if}
												</Table.Cell>
											</Table.Row>
										{:else}
											<Table.Row>
												<Table.Cell colspan={6} class="p-4">
													<EmptyStatePanel compact title="Belum ada relasi orang tua" description="Relasi ayah/ibu/wali akan tampil setelah normalisasi data orang tua." />
												</Table.Cell>
											</Table.Row>
										{/each}
									</Table.Body>
								</Table.Root>
							</div>
						</Card.Content>
					</Card.Root>
				</Tabs.Content>

				<Tabs.Content value="homeroom">
					<div class="grid gap-4 lg:grid-cols-3">
						<Card.Root class="lg:col-span-2">
							<Card.Header>
								<Card.Title class="text-base">Riwayat Wali Kelas</Card.Title>
							</Card.Header>
							<Card.Content class="p-0">
								<div class="overflow-x-auto">
									<Table.Root>
										<Table.Header>
											<Table.Row>
												<Table.Head>Guru</Table.Head>
												<Table.Head>Tahun Ajaran</Table.Head>
												<Table.Head>Mulai</Table.Head>
												<Table.Head>Selesai</Table.Head>
												<Table.Head>Status</Table.Head>
												<Table.Head>Catatan</Table.Head>
											</Table.Row>
										</Table.Header>
										<Table.Body>
											{#each detail.homeroom_assignments as item (item.id)}
												<Table.Row>
													<Table.Cell class="font-medium">{item.employee_name}</Table.Cell>
													<Table.Cell>{item.academic_year_name || '-'}</Table.Cell>
													<Table.Cell class="text-muted-foreground">{fmtDate(item.start_date)}</Table.Cell>
													<Table.Cell class="text-muted-foreground">{fmtDate(item.end_date)}</Table.Cell>
													<Table.Cell>
														<Badge variant={item.is_active ? 'default' : 'secondary'}>{item.is_active ? 'Aktif' : 'Nonaktif'}</Badge>
													</Table.Cell>
													<Table.Cell class="text-muted-foreground">{item.notes || '-'}</Table.Cell>
												</Table.Row>
											{:else}
												<Table.Row>
													<Table.Cell colspan={6} class="p-4">
														<EmptyStatePanel compact title="Belum ada wali kelas" description="Tetapkan wali kelas aktif agar tampil di ringkasan rombel." />
													</Table.Cell>
												</Table.Row>
											{/each}
										</Table.Body>
									</Table.Root>
								</div>
							</Card.Content>
						</Card.Root>

						<Card.Root>
							<Card.Header>
								<Card.Title class="text-base">{activeHomeroom ? 'Perbarui Wali Kelas' : 'Tetapkan Wali Kelas'}</Card.Title>
								<Card.Description>Pilih dari pegawai aktif. Guru mapel dan wali aktif saat ini tetap muncul sebagai fallback.</Card.Description>
							</Card.Header>
							<Card.Content class="space-y-3">
								<div class="space-y-1.5">
									<label for="homeroom-employee-search" class="block text-xs font-medium text-muted-foreground">Cari Pegawai</label>
									<div class="relative">
										<Search class="pointer-events-none absolute left-2.5 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
										<Input
											id="homeroom-employee-search"
											bind:value={employeeSearch}
											placeholder="Nama, UID, NIP, atau tipe pegawai"
											class="pl-8"
											disabled={employeesLoading && !employeesLoaded && employeeOptions.length === 0}
										/>
									</div>
								</div>

								{#if selectedEmployeeOption}
									<div class="rounded-lg border border-primary/20 bg-primary/10 p-3 text-sm">
										<div class="flex items-start gap-2">
											<UserCheck class="mt-0.5 size-4 shrink-0 text-primary" />
											<div class="min-w-0 space-y-1">
												<p class="font-medium text-foreground">{selectedEmployeeOption.name}</p>
												<p class="text-xs text-muted-foreground">{employeeOptionSubtitle(selectedEmployeeOption)}</p>
												<Badge class="border-primary/20 bg-primary/15 text-primary">{employeeSourceLabel(selectedEmployeeOption)}</Badge>
											</div>
										</div>
									</div>
								{/if}

								{#if employeesError}
									<div class="rounded-lg border border-destructive/30 bg-destructive/10 p-3 text-sm text-destructive">
										<p class="font-medium">Daftar pegawai aktif belum dapat dimuat.</p>
										<p class="mt-1 text-xs text-destructive/80">
											{employeeOptions.length > 0
												? 'Opsi sementara memakai guru mapel dan wali kelas yang sudah tercatat pada rombel ini.'
												: employeesError}
										</p>
										<Button
											type="button"
											variant="outline"
											size="sm"
											class="mt-3 border-destructive/30 text-destructive hover:bg-destructive/10 hover:text-destructive"
											disabled={employeesLoading}
											onclick={() => void loadEmployees()}
										>
											<RefreshCw class={`mr-2 size-4 ${employeesLoading ? 'animate-spin' : ''}`} />
											Muat Ulang Pegawai
										</Button>
									</div>
								{/if}

								<div class="space-y-2">
									<div class="flex items-center justify-between gap-3">
										<p class="text-xs font-medium text-muted-foreground">Opsi Wali Kelas</p>
										{#if employeesLoading && employeesLoaded}
											<span class="inline-flex items-center gap-1.5 text-xs text-muted-foreground">
												<span class="size-3 animate-spin rounded-full border-2 border-current border-t-transparent"></span>
												Memperbarui
											</span>
										{/if}
									</div>
									{#if employeesLoading && !employeesLoaded && employeeOptions.length === 0}
										<div class="space-y-2 rounded-lg border border-border p-2">
											<Skeleton class="h-12 w-full" />
											<Skeleton class="h-12 w-full" />
											<Skeleton class="h-12 w-full" />
										</div>
									{:else if employeeOptions.length === 0}
										<EmptyStatePanel
											compact
											title="Belum ada opsi wali kelas"
											description={employeesError
												? 'Muat ulang daftar pegawai, atau hubungi admin jika endpoint pegawai tidak dapat diakses.'
												: 'Tambahkan pegawai aktif lebih dulu agar wali kelas bisa dipilih.'}
										/>
									{:else if filteredEmployeeOptions.length === 0}
										<EmptyStatePanel compact title="Pegawai tidak ditemukan" description="Ubah kata kunci pencarian untuk melihat opsi wali kelas lain." />
									{:else}
										<div class="max-h-72 overflow-y-auto rounded-lg border border-border p-1">
											{#each filteredEmployeeOptions as option (option.id)}
												<button
													type="button"
													class={`flex w-full items-start justify-between gap-3 rounded-md border px-3 py-2 text-left text-sm transition-colors ${
														homeroomEmployeeId === option.id
															? 'border-primary bg-primary/10 text-foreground'
															: 'border-transparent text-foreground hover:bg-accent hover:text-accent-foreground'
													}`}
													aria-pressed={homeroomEmployeeId === option.id}
													aria-label={`Pilih ${option.name} sebagai wali kelas`}
													onclick={() => selectHomeroomEmployee(option)}
												>
													<span class="min-w-0">
														<span class="block truncate font-medium">{option.name}</span>
														<span class="block truncate text-xs text-muted-foreground">{employeeOptionSubtitle(option)}</span>
													</span>
													<span class="flex shrink-0 flex-col items-end gap-1">
														<Badge variant={option.isActiveEmployee ? 'secondary' : 'outline'}>{employeeSourceLabel(option)}</Badge>
														{#if employmentTypeLabel(option.employmentType)}
															<span class="text-xs text-muted-foreground">{employmentTypeLabel(option.employmentType)}</span>
														{/if}
													</span>
												</button>
											{/each}
										</div>
									{/if}
								</div>
								<div>
									<label for="homeroom-notes" class="mb-1 block text-xs font-medium text-muted-foreground">Catatan</label>
									<Textarea
										id="homeroom-notes"
										bind:value={homeroomNotes}
										rows={3}
									/>
								</div>
								<LoadingButton
									class="w-full"
									loading={saveBusy}
									loadingLabel="Menyimpan..."
									disabled={!canSaveHomeroom}
									onclick={() => void saveHomeroom()}
									label="Simpan Wali Kelas"
								/>
							</Card.Content>
						</Card.Root>
					</div>
				</Tabs.Content>

				<Tabs.Content value="teachers">
					<div class="grid gap-4 lg:grid-cols-3">
						<Card.Root class="lg:col-span-2">
							<Card.Header>
								<Card.Title class="text-base">Guru Mapel</Card.Title>
								<Card.Description>Pasangan mata pelajaran dan guru pengampu pada rombel ini.</Card.Description>
							</Card.Header>
							<Card.Content class="p-0">
								<div class="overflow-x-auto">
									<Table.Root>
										<Table.Header>
											<Table.Row>
												<Table.Head>Mata Pelajaran</Table.Head>
												<Table.Head>Kode</Table.Head>
												<Table.Head>Guru</Table.Head>
												<Table.Head class="text-right">Aksi</Table.Head>
											</Table.Row>
										</Table.Header>
										<Table.Body>
											{#each detail.subject_assignments as assignment (assignment.id)}
												<Table.Row>
													<Table.Cell class="font-medium">{assignment.subject_name}</Table.Cell>
													<Table.Cell class="text-muted-foreground">{assignment.subject_code}</Table.Cell>
													<Table.Cell>{assignment.teacher_name}</Table.Cell>
													<Table.Cell>
														<div class="flex justify-end gap-2">
															<Button
																type="button"
																variant="outline"
																size="xs"
																onclick={() => editSubjectAssignment(assignment)}
																disabled={subjectAssignmentSaveBusy || deleteSubjectAssignmentBusyId !== ''}
															>
																<Pencil class="mr-1 size-3.5" />
																Edit
															</Button>
															<LoadingButton
																type="button"
																variant="destructive"
																size="xs"
																loading={deleteSubjectAssignmentBusyId === assignment.id}
																loadingLabel="Hapus..."
																disabled={subjectAssignmentSaveBusy || (deleteSubjectAssignmentBusyId !== '' && deleteSubjectAssignmentBusyId !== assignment.id)}
																onclick={() => void deleteSubjectAssignment(assignment)}
															>
																<Trash2 class="mr-1 size-3.5" />
																Hapus
															</LoadingButton>
														</div>
													</Table.Cell>
												</Table.Row>
											{:else}
												<Table.Row>
													<Table.Cell colspan={4} class="p-4">
														<EmptyStatePanel compact title="Belum ada guru mapel" description="Tambahkan mata pelajaran dan guru pengampu untuk rombel ini." />
													</Table.Cell>
												</Table.Row>
											{/each}
										</Table.Body>
									</Table.Root>
								</div>
							</Card.Content>
						</Card.Root>

						<Card.Root>
							<Card.Header>
								<Card.Title class="text-base">{editingSubjectAssignmentId ? 'Perbarui Guru Mapel' : 'Tambah Guru Mapel'}</Card.Title>
								<Card.Description>Pilih mapel aktif dan pegawai aktif sebagai guru pengampu.</Card.Description>
							</Card.Header>
							<Card.Content class="space-y-3">
								{#if subjectsError}
									<div class="rounded-lg border border-destructive/30 bg-destructive/10 p-3 text-sm text-destructive">
										<p class="font-medium">Daftar mata pelajaran belum dapat dimuat.</p>
										<p class="mt-1 text-xs text-destructive/80">
											{subjectOptions.length > 0
												? 'Opsi sementara memakai mata pelajaran yang sudah tercatat pada rombel ini.'
												: subjectsError}
										</p>
										<Button
											type="button"
											variant="outline"
											size="sm"
											class="mt-3 border-destructive/30 text-destructive hover:bg-destructive/10 hover:text-destructive"
											disabled={subjectsLoading}
											onclick={() => void loadSubjects()}
										>
											<RefreshCw class={`mr-2 size-4 ${subjectsLoading ? 'animate-spin' : ''}`} />
											Muat Ulang Mapel
										</Button>
									</div>
								{/if}

								<div class="space-y-1.5">
									<label for="subject-assignment-subject" class="block text-xs font-medium text-muted-foreground">Mata Pelajaran</label>
									{#if subjectsLoading && !subjectsLoaded && subjectOptions.length === 0}
										<Skeleton class="h-10 w-full" />
									{:else}
										<select
											id="subject-assignment-subject"
											bind:value={subjectAssignmentSubjectId}
											class="h-10 w-full rounded-md border border-input bg-background px-3 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring"
											disabled={availableSubjectOptions.length === 0 || subjectAssignmentSaveBusy}
										>
											<option value="">Pilih mapel</option>
											{#each availableSubjectOptions as subject (subject.id)}
												<option value={subject.id}>{subjectOptionLabel(subject)}{subject.isFallback ? ' - dari rombel' : ''}</option>
											{/each}
										</select>
									{/if}
								</div>

								<div class="space-y-1.5">
									<label for="subject-teacher-search" class="block text-xs font-medium text-muted-foreground">Cari Guru</label>
									<div class="relative">
										<Search class="pointer-events-none absolute left-2.5 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
										<Input
											id="subject-teacher-search"
											bind:value={subjectTeacherSearch}
											placeholder="Nama, UID, NIP, atau tipe pegawai"
											class="pl-8"
											disabled={employeesLoading && !employeesLoaded && employeeOptions.length === 0}
										/>
									</div>
								</div>

								<div class="space-y-1.5">
									<label for="subject-assignment-teacher" class="block text-xs font-medium text-muted-foreground">Guru Pengampu</label>
									{#if employeesLoading && !employeesLoaded && employeeOptions.length === 0}
										<Skeleton class="h-10 w-full" />
									{:else}
										<select
											id="subject-assignment-teacher"
											bind:value={subjectAssignmentTeacherId}
											class="h-10 w-full rounded-md border border-input bg-background px-3 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring"
											disabled={filteredSubjectTeacherOptions.length === 0 || subjectAssignmentSaveBusy}
										>
											<option value="">Pilih guru</option>
											{#each filteredSubjectTeacherOptions as option (option.id)}
												<option value={option.id}>{option.name} - {employeeOptionSubtitle(option)}</option>
											{/each}
										</select>
									{/if}
								</div>

								{#if selectedSubjectOption || selectedSubjectTeacherOption}
									<div class="rounded-lg border border-primary/20 bg-primary/10 p-3 text-sm">
										<div class="flex items-start gap-2">
											<UserCheck class="mt-0.5 size-4 shrink-0 text-primary" />
											<div class="min-w-0 space-y-1">
												<p class="font-medium text-foreground">{selectedSubjectOption ? subjectOptionLabel(selectedSubjectOption) : 'Mapel belum dipilih'}</p>
												<p class="text-xs text-muted-foreground">
													{selectedSubjectTeacherOption ? `${selectedSubjectTeacherOption.name} - ${employeeOptionSubtitle(selectedSubjectTeacherOption)}` : 'Guru pengampu belum dipilih'}
												</p>
											</div>
										</div>
									</div>
								{/if}

								{#if employeesError && employeeOptions.length === 0}
									<div class="rounded-lg border border-destructive/30 bg-destructive/10 p-3 text-sm text-destructive">
										<p class="font-medium">Daftar pegawai aktif belum dapat dimuat.</p>
										<p class="mt-1 text-xs text-destructive/80">{employeesError}</p>
										<Button
											type="button"
											variant="outline"
											size="sm"
											class="mt-3 border-destructive/30 text-destructive hover:bg-destructive/10 hover:text-destructive"
											disabled={employeesLoading}
											onclick={() => void loadEmployees()}
										>
											<RefreshCw class={`mr-2 size-4 ${employeesLoading ? 'animate-spin' : ''}`} />
											Muat Ulang Pegawai
										</Button>
									</div>
								{/if}

								{#if availableSubjectOptions.length === 0 && !subjectsLoading}
									<EmptyStatePanel compact title="Tidak ada mapel tersedia" description="Semua mata pelajaran aktif sudah memiliki guru pada rombel ini." />
								{:else if subjectTeacherSearch && filteredSubjectTeacherOptions.length === 0}
									<EmptyStatePanel compact title="Guru tidak ditemukan" description="Ubah kata kunci pencarian untuk melihat opsi guru lain." />
								{/if}

								<div class="flex gap-2">
									<LoadingButton
										class="flex-1"
										loading={subjectAssignmentSaveBusy}
										loadingLabel="Menyimpan..."
										disabled={!canSaveSubjectAssignment || deleteSubjectAssignmentBusyId !== ''}
										onclick={() => void saveSubjectAssignment()}
										label={editingSubjectAssignmentId ? 'Simpan Perubahan' : 'Tambah Guru Mapel'}
									/>
									{#if editingSubjectAssignmentId}
										<Button
											type="button"
											variant="outline"
											disabled={subjectAssignmentSaveBusy}
											onclick={resetSubjectAssignmentForm}
										>
											Batal
										</Button>
									{/if}
								</div>
							</Card.Content>
						</Card.Root>
					</div>
				</Tabs.Content>

				<Tabs.Content value="schedule">
					<Card.Root>
						<Card.Header>
							<Card.Title class="text-base">Jadwal Rombel</Card.Title>
						</Card.Header>
						<Card.Content class="p-0">
							<div class="overflow-x-auto">
								<Table.Root>
									<Table.Header>
										<Table.Row>
											<Table.Head>Hari</Table.Head>
											<Table.Head>Waktu</Table.Head>
											<Table.Head>Mata Pelajaran</Table.Head>
											<Table.Head>Guru</Table.Head>
											<Table.Head>Ruang</Table.Head>
										</Table.Row>
									</Table.Header>
									<Table.Body>
										{#each detail.timetable_slots as slot (slot.id)}
											<Table.Row>
												<Table.Cell class="font-medium">{dayLabels[slot.day_of_week] ?? `Hari ${slot.day_of_week}`}</Table.Cell>
												<Table.Cell class="text-muted-foreground">{fmtTime(slot.start_time)}-{fmtTime(slot.end_time)}</Table.Cell>
												<Table.Cell>{slot.subject_name}</Table.Cell>
												<Table.Cell class="text-muted-foreground">{slot.teacher_name}</Table.Cell>
												<Table.Cell class="text-muted-foreground">{slot.room_label || '-'}</Table.Cell>
											</Table.Row>
										{:else}
											<Table.Row>
												<Table.Cell colspan={5} class="p-4">
													<EmptyStatePanel compact title="Belum ada slot jadwal" description="Jadwal mingguan akan tampil setelah slot dibuat." />
												</Table.Cell>
											</Table.Row>
										{/each}
									</Table.Body>
								</Table.Root>
							</div>
						</Card.Content>
					</Card.Root>
				</Tabs.Content>
			</Tabs.Root>
		{/if}
	</AsyncContent>
</div>
