<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import { AlertTriangle, CalendarDays, Edit3, ExternalLink, Plus, RefreshCw, Search, Trash2, X } from '@lucide/svelte';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
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
	import { clientApiPath, readClientApiData, readClientJson } from '$lib/client/api';

	type WeeklyClass = {
		id: string;
		code: string;
		name: string;
		level: string;
	};

	type WeeklyTeacher = {
		id: string;
		nip: string;
		nama: string;
		unit_kerja: string;
	};

	type WeeklySubject = {
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

	type WeeklyAssignment = {
		id: string;
		class_id: string;
		class_code: string;
		class_name: string;
		class_level: string;
		subject_id: string;
		subject_code: string;
		subject_name: string;
		subject_category: string;
		is_schedule_activity: boolean;
		default_weekly_hours: number;
		teacher_employee_id: string;
		teacher_name: string;
		teacher_nip: string;
	};

	type WeeklySlot = {
		id: string;
		assignment_id: string;
		day_of_week: number;
		start_time: string;
		end_time: string;
		room_label: string;
		notes: string;
		class_id: string;
		class_name: string;
		class_code: string;
		class_level: string;
		subject_id: string;
		subject_name: string;
		subject_code: string;
		subject_category: string;
		is_schedule_activity: boolean;
		teacher_employee_id: string;
		teacher_name: string;
		conflict_status: 'ok' | 'conflict' | 'invalid_time_range' | string;
		conflict_label: string;
		conflict_count: number;
	};

	type TimetableConflict = {
		conflict_type: 'same_class' | 'same_teacher' | 'same_room' | 'invalid_time_range' | string;
		message: string;
		slot_id: string;
		assignment_id: string;
		day_of_week: number;
		start_time: string;
		end_time: string;
		room_label: string;
		class_id: string;
		class_code: string;
		class_name: string;
		subject_code: string;
		subject_name: string;
		teacher_employee_id: string;
		teacher_name: string;
		related_slot_id: string | null;
		related_day_of_week: number;
		related_start_time: string;
		related_end_time: string;
		related_room_label: string;
		related_class_code: string;
		related_class_name: string;
		related_subject_code: string;
		related_subject_name: string;
		related_teacher_name: string;
	};

	type WeeklyTimetable = {
		active_academic_year_id: string;
		active_academic_year_name: string;
		classes: WeeklyClass[];
		teachers: WeeklyTeacher[];
		subjects: WeeklySubject[];
		assignments: WeeklyAssignment[];
		slots: WeeklySlot[];
		conflicts: TimetableConflict[];
	};

	type TimetableMode = 'class' | 'teacher';

	const dayLabels: Record<number, string> = {
		1: 'Senin',
		2: 'Selasa',
		3: 'Rabu',
		4: 'Kamis',
		5: 'Jumat',
		6: 'Sabtu',
	};

	let timetable = $state<WeeklyTimetable | null>(null);
	let weeklyPromise = $state<Promise<WeeklyTimetable> | null>(null);
	let requestId = 0;
	let refreshBusy = $state(false);
	let mode = $state<TimetableMode>('class');
	let selectedClassId = $state('all');
	let selectedTeacherId = $state('all');
	let selectedDay = $state('all');
	let searchQuery = $state('');

	let editingSlot = $state<WeeklySlot | null>(null);
	let formClassId = $state('');
	let formAssignmentId = $state('');
	let formDay = $state('1');
	let formStart = $state('');
	let formEnd = $state('');
	let formRoom = $state('');
	let formNotes = $state('');
	let formError = $state('');
	let slotSaveBusy = $state(false);
	let deleteBusyId = $state('');

	const classes = $derived(timetable?.classes ?? []);
	const teachers = $derived(timetable?.teachers ?? []);
	const assignments = $derived(timetable?.assignments ?? []);
	const slots = $derived(timetable?.slots ?? []);
	const conflicts = $derived(timetable?.conflicts ?? []);

	const filteredSlots = $derived.by(() => {
		const query = searchQuery.trim().toLowerCase();
		return slots.filter((slot) => {
			if (selectedDay !== 'all' && slot.day_of_week !== Number(selectedDay)) return false;
			if (mode === 'class' && selectedClassId !== 'all' && slot.class_id !== selectedClassId) return false;
			if (mode === 'teacher' && selectedTeacherId !== 'all' && slot.teacher_employee_id !== selectedTeacherId) return false;
			if (!query) return true;
			return [
				slot.class_code,
				slot.class_name,
				slot.subject_code,
				slot.subject_name,
				slot.teacher_name,
				slot.room_label,
				slot.notes,
			].some((value) => value.toLowerCase().includes(query));
		});
	});

	const dayGroups = $derived.by(() =>
		Object.entries(dayLabels).map(([day, label]) => ({
			day: Number(day),
			label,
			slots: filteredSlots.filter((slot) => slot.day_of_week === Number(day)),
		}))
	);

	const conflictSlotIds = $derived.by(() => {
		const ids = new Set<string>();
		for (const conflict of conflicts) {
			ids.add(conflict.slot_id);
			if (conflict.related_slot_id) ids.add(conflict.related_slot_id);
		}
		return ids;
	});

	const summary = $derived.by(() => ({
		totalSlots: slots.length,
		totalConflicts: conflicts.length,
		conflictedSlots: conflictSlotIds.size,
		totalClasses: classes.length,
		totalTeachers: new Set(slots.map((slot) => slot.teacher_employee_id)).size,
	}));

	const formAssignments = $derived.by(() => assignments.filter((assignment) => assignment.class_id === formClassId));
	const selectedAssignment = $derived(formAssignments.find((assignment) => assignment.id === formAssignmentId) ?? null);
	const canSaveSlot = $derived(Boolean(formClassId && formAssignmentId && formDay && formStart && formEnd && formStart < formEnd));

	async function fetchWeeklyTimetable(): Promise<WeeklyTimetable> {
		return await fetch('/api/academic/timetable/weekly').then((response) =>
			readClientApiData<WeeklyTimetable>(response, 'Gagal memuat jadwal mingguan')
		);
	}

	function loadWeekly() {
		const current = ++requestId;
		weeklyPromise = fetchWeeklyTimetable().then((payload) => {
			if (current === requestId) applyWeekly(payload);
			return payload;
		});
	}

	function applyWeekly(payload: WeeklyTimetable) {
		timetable = payload;
		if (selectedClassId !== 'all' && !payload.classes.some((item) => item.id === selectedClassId)) selectedClassId = 'all';
		if (selectedTeacherId !== 'all' && !payload.teachers.some((item) => item.id === selectedTeacherId)) selectedTeacherId = 'all';
		if (!formClassId || !payload.classes.some((item) => item.id === formClassId)) {
			formClassId = payload.classes[0]?.id ?? '';
		}
		normalizeFormAssignment(payload.assignments);
		return payload;
	}

	async function refreshWeekly() {
		refreshBusy = true;
		try {
			loadWeekly();
			await weeklyPromise;
		} finally {
			refreshBusy = false;
		}
	}

	function normalizeFormAssignment(source = assignments) {
		const options = source.filter((assignment) => assignment.class_id === formClassId);
		if (!options.some((assignment) => assignment.id === formAssignmentId)) {
			formAssignmentId = options[0]?.id ?? '';
		}
	}

	function resetForm(classId = selectedClassId !== 'all' ? selectedClassId : classes[0]?.id ?? '') {
		editingSlot = null;
		formClassId = classId;
		formDay = selectedDay !== 'all' ? selectedDay : '1';
		formStart = '';
		formEnd = '';
		formRoom = '';
		formNotes = '';
		formError = '';
		normalizeFormAssignment();
	}

	function editSlot(slot: WeeklySlot) {
		editingSlot = slot;
		formClassId = slot.class_id;
		formAssignmentId = slot.assignment_id;
		formDay = String(slot.day_of_week);
		formStart = fmtTime(slot.start_time);
		formEnd = fmtTime(slot.end_time);
		formRoom = slot.room_label ?? '';
		formNotes = slot.notes ?? '';
		formError = '';
		setTimeout(() => document.getElementById('timetable-slot-form')?.scrollIntoView({ block: 'start', behavior: 'smooth' }), 0);
	}

	function onFormClassChange() {
		normalizeFormAssignment();
	}

	function validateSlotForm() {
		if (!formClassId) return 'Rombel wajib dipilih.';
		if (!formAssignmentId) return 'Guru mapel wajib dipilih.';
		if (!formDay || Number(formDay) < 1 || Number(formDay) > 6) return 'Hari wajib dipilih.';
		if (!formStart || !formEnd) return 'Jam mulai dan selesai wajib diisi.';
		if (formStart >= formEnd) return 'Jam selesai harus setelah jam mulai.';
		return '';
	}

	async function saveSlot() {
		const validation = validateSlotForm();
		if (validation) {
			formError = validation;
			toast.error(validation);
			return;
		}
		const classId = editingSlot?.class_id ?? formClassId;
		const path = editingSlot
			? clientApiPath`/api/academic/rombel/${classId}/timetable-slots/${editingSlot.id}`
			: clientApiPath`/api/academic/rombel/${classId}/timetable-slots`;
		slotSaveBusy = true;
		try {
			const response = await fetch(path, {
				method: editingSlot ? 'PUT' : 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					assignment_id: formAssignmentId,
					day_of_week: Number(formDay),
					start_time: formStart,
					end_time: formEnd,
					room_label: formRoom.trim(),
					notes: formNotes.trim(),
				}),
			});
			await readClientJson<unknown>(response);
			toast.success(editingSlot ? 'Jam pelajaran diperbarui' : 'Jam pelajaran ditambahkan');
			resetForm(classId);
			await refreshWeekly();
		} catch (error) {
			const message = errorMessage(error, 'Gagal menyimpan jam pelajaran.');
			formError = message;
			toast.error(message);
		} finally {
			slotSaveBusy = false;
		}
	}

	async function deleteSlot(slot: WeeklySlot) {
		if (!(await confirmAction({
			title: 'Hapus Jam Pelajaran',
			message: `Hapus ${slot.subject_name} ${slot.class_code} pada ${dayLabels[slot.day_of_week] ?? 'hari ini'} pukul ${fmtTime(slot.start_time)}-${fmtTime(slot.end_time)}?`,
			confirmLabel: 'Hapus',
			tone: 'danger',
		}))) return;
		deleteBusyId = slot.id;
		try {
			const response = await fetch(clientApiPath`/api/academic/rombel/${slot.class_id}/timetable-slots/${slot.id}`, { method: 'DELETE' });
			await readClientJson<unknown>(response);
			if (editingSlot?.id === slot.id) resetForm(slot.class_id);
			toast.success('Jam pelajaran dihapus');
			await refreshWeekly();
		} catch (error) {
			toast.error(errorMessage(error, 'Gagal menghapus jam pelajaran.'));
		} finally {
			deleteBusyId = '';
		}
	}

	function slotConflicts(slot: WeeklySlot) {
		return conflicts.filter((conflict) => conflict.slot_id === slot.id || conflict.related_slot_id === slot.id);
	}

	function conflictTypeLabel(type: string) {
		const labels: Record<string, string> = {
			same_class: 'Rombel bentrok',
			same_teacher: 'Guru bentrok',
			same_room: 'Ruang bentrok',
			invalid_time_range: 'Jam tidak valid',
		};
		return labels[type] ?? type;
	}

	function teacherLabel(teacher: WeeklyTeacher) {
		return teacher.nip ? `${teacher.nama} (${teacher.nip})` : teacher.nama;
	}

	function assignmentLabel(assignment: WeeklyAssignment) {
		const activity = assignment.is_schedule_activity ? 'Kegiatan' : assignment.subject_category || 'Mapel';
		return `${assignment.subject_name} - ${assignment.teacher_name} (${activity})`;
	}

	function fmtTime(value: string | null | undefined) {
		return value ? value.slice(0, 5) : '-';
	}

	function slotTimeLabel(slot: WeeklySlot) {
		return `${fmtTime(slot.start_time)}-${fmtTime(slot.end_time)}`;
	}

	function errorMessage(error: unknown, fallback: string) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return fallback;
	}

	function weeklyErrorMessage(error: unknown) {
		return errorMessage(error, 'Jadwal mingguan belum dapat dimuat. Periksa layanan sistem lalu coba lagi.');
	}

	function handleRenderError(error: unknown) {
		console.error('Weekly timetable render failed', error);
	}

	onMount(() => {
		loadWeekly();
	});
</script>

<svelte:head>
	<title>Jadwal Pelajaran | MTsN 2 Kolaka Utara</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex flex-col gap-3 md:flex-row md:items-end md:justify-between">
		<div>
			<p class="text-sm font-medium text-primary">Akademik</p>
			<h1 class="text-2xl font-semibold tracking-normal text-foreground">Jadwal Pelajaran</h1>
			<p class="mt-1 max-w-2xl text-sm text-muted-foreground">
				Pengaturan jadwal mingguan untuk rombel dan guru pada tahun ajaran aktif.
			</p>
		</div>
		<div class="flex flex-wrap gap-2">
			<Button variant="outline" onclick={() => resetForm()} disabled={slotSaveBusy}>
				<Plus class="mr-2 size-4" />
				Tambah jam pelajaran
			</Button>
			<Button variant="outline" onclick={() => void refreshWeekly()} disabled={refreshBusy || slotSaveBusy} aria-label="Muat ulang jadwal">
				<RefreshCw class={`mr-2 size-4 ${refreshBusy ? 'animate-spin' : ''}`} />
				Muat Ulang
			</Button>
		</div>
	</div>

	<AsyncContent promise={weeklyPromise} onerror={handleRenderError}>
		{#snippet pending()}
			<div class="space-y-4">
				<div class="grid gap-3 md:grid-cols-4">
					<Skeleton class="h-24 w-full" />
					<Skeleton class="h-24 w-full" />
					<Skeleton class="h-24 w-full" />
					<Skeleton class="h-24 w-full" />
				</div>
				<Skeleton class="h-[34rem] w-full" />
			</div>
		{/snippet}

		{#snippet failed(error, reset)}
			<RecoveryPanel
				title="Jadwal Belum Tersaji"
				message={weeklyErrorMessage(error)}
				onRetry={() => {
					reset?.();
					loadWeekly();
				}}
			/>
		{/snippet}

		<div class="grid gap-3 md:grid-cols-4">
			<Card.Root>
				<Card.Header class="pb-2">
					<Card.Description>Tahun Ajaran Aktif</Card.Description>
					<Card.Title class="text-xl">{timetable?.active_academic_year_name || 'Belum ada'}</Card.Title>
				</Card.Header>
			</Card.Root>
			<Card.Root>
				<Card.Header class="pb-2">
					<Card.Description>Total jam pelajaran</Card.Description>
					<Card.Title class="text-2xl">{summary.totalSlots}</Card.Title>
				</Card.Header>
			</Card.Root>
			<Card.Root>
				<Card.Header class="pb-2">
					<Card.Description>Rombel Aktif</Card.Description>
					<Card.Title class="text-2xl">{summary.totalClasses}</Card.Title>
				</Card.Header>
			</Card.Root>
			<Card.Root>
				<Card.Header class="pb-2">
					<Card.Description>Jadwal perlu diperiksa</Card.Description>
					<Card.Title class={`text-2xl ${summary.totalConflicts > 0 ? 'text-destructive' : ''}`}>{summary.totalConflicts}</Card.Title>
				</Card.Header>
			</Card.Root>
		</div>

		<Card.Root>
			<Card.Content class="grid gap-3 p-4 lg:grid-cols-[15rem_15rem_12rem_1fr]">
				<div class="space-y-1.5">
					<p class="text-sm font-medium">Tampilan</p>
					<div id="timetable-mode" class="grid grid-cols-2 rounded-md border border-input bg-background p-1">
						<button
							type="button"
							class={`rounded px-3 py-2 text-sm font-medium ${mode === 'class' ? 'bg-primary text-primary-foreground' : 'text-muted-foreground hover:bg-muted'}`}
							aria-pressed={mode === 'class'}
							onclick={() => (mode = 'class')}
						>
							Per Rombel
						</button>
						<button
							type="button"
							class={`rounded px-3 py-2 text-sm font-medium ${mode === 'teacher' ? 'bg-primary text-primary-foreground' : 'text-muted-foreground hover:bg-muted'}`}
							aria-pressed={mode === 'teacher'}
							onclick={() => (mode = 'teacher')}
						>
							Per Guru
						</button>
					</div>
				</div>
				{#if mode === 'class'}
					<div class="space-y-1.5">
						<label for="class-filter" class="text-sm font-medium">Rombel</label>
						<select id="class-filter" bind:value={selectedClassId} class="h-10 w-full rounded-md border border-input bg-background px-3 text-sm">
							<option value="all">Semua rombel</option>
							{#each classes as klass (klass.id)}
								<option value={klass.id}>{klass.code} - {klass.name}</option>
							{/each}
						</select>
					</div>
				{:else}
					<div class="space-y-1.5">
						<label for="teacher-filter" class="text-sm font-medium">Guru</label>
						<select id="teacher-filter" bind:value={selectedTeacherId} class="h-10 w-full rounded-md border border-input bg-background px-3 text-sm">
							<option value="all">Semua guru</option>
							{#each teachers as teacher (teacher.id)}
								<option value={teacher.id}>{teacherLabel(teacher)}</option>
							{/each}
						</select>
					</div>
				{/if}
				<div class="space-y-1.5">
					<label for="day-filter" class="text-sm font-medium">Hari</label>
					<select id="day-filter" bind:value={selectedDay} class="h-10 w-full rounded-md border border-input bg-background px-3 text-sm">
						<option value="all">Semua hari</option>
						{#each Object.entries(dayLabels) as [day, label] (`day-${day}`)}
							<option value={day}>{label}</option>
						{/each}
					</select>
				</div>
				<div class="space-y-1.5">
					<label for="slot-search" class="text-sm font-medium">Cari</label>
					<div class="relative">
						<Search class="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
						<Input id="slot-search" class="pl-9" bind:value={searchQuery} placeholder="Mapel, guru, ruang, atau rombel" />
					</div>
				</div>
			</Card.Content>
		</Card.Root>

		{#if !timetable?.active_academic_year_id}
			<EmptyStatePanel compact title="Tahun ajaran aktif belum ada" description="Aktifkan tahun ajaran sebelum menyusun jadwal pelajaran." />
		{:else}
			<div class="grid gap-4 xl:grid-cols-[minmax(0,1fr)_24rem]">
				<div class="space-y-4">
					<Card.Root>
						<Card.Header>
							<div class="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
								<div>
									<Card.Title class="text-base">Jadwal Mingguan</Card.Title>
									<Card.Description>{filteredSlots.length} jam pelajaran tampil dari {summary.totalSlots} jam pelajaran aktif.</Card.Description>
								</div>
								{#if summary.conflictedSlots > 0}
									<Badge variant="destructive">
										<AlertTriangle class="mr-1 size-3" />
										{summary.conflictedSlots} jadwal perlu diperiksa
									</Badge>
								{/if}
							</div>
						</Card.Header>
						<Card.Content class="space-y-5">
							{#if filteredSlots.length === 0}
								<EmptyStatePanel compact title="Jam pelajaran belum ditemukan" description="Ubah pilihan pencarian atau tambahkan jam pelajaran pertama untuk rombel yang dipilih." />
							{:else}
								{#each dayGroups as group (group.day)}
									<div class="space-y-3">
										<div class="flex items-center justify-between gap-3">
											<div class="flex items-center gap-2">
												<CalendarDays class="size-4 text-primary" />
												<h2 class="text-sm font-semibold text-foreground">{group.label}</h2>
											</div>
											<Badge variant="outline">{group.slots.length} jam</Badge>
										</div>
										{#if group.slots.length > 0}
											<div class="grid gap-3 lg:grid-cols-2">
												{#each group.slots as slot (slot.id)}
													<div class={`rounded-lg border p-4 ${slot.conflict_status !== 'ok' ? 'border-destructive/40 bg-destructive/5' : 'border-border bg-card'}`}>
														<div class="flex items-start justify-between gap-3">
															<div class="min-w-0">
																<p class="truncate text-sm font-semibold text-foreground">{slot.subject_name}</p>
																<p class="mt-1 text-xs text-muted-foreground">{slot.class_code} - {slot.teacher_name}</p>
															</div>
															<Badge variant={slot.conflict_status === 'ok' ? 'outline' : 'destructive'}>{slot.conflict_label}</Badge>
														</div>
														<div class="mt-3 flex flex-wrap gap-2 text-xs text-muted-foreground">
															<Badge variant="secondary">{slotTimeLabel(slot)}</Badge>
															<Badge variant="outline">{slot.room_label || 'Ruang belum diisi'}</Badge>
															{#if slot.is_schedule_activity}
																<Badge class="border-primary/20 bg-primary/15 text-primary">Kegiatan</Badge>
															{/if}
														</div>
														{#if slot.notes}
															<p class="mt-3 text-xs text-muted-foreground">{slot.notes}</p>
														{/if}
														{#if slot.conflict_count > 0}
															<div class="mt-3 space-y-1 rounded-md border border-destructive/20 bg-background p-2">
																{#each slotConflicts(slot) as conflict (`${slot.id}-${conflict.conflict_type}-${conflict.related_slot_id ?? 'self'}`)}
																	<p class="text-xs text-destructive">
																		<span class="font-medium">{conflictTypeLabel(conflict.conflict_type)}:</span>
																		{conflict.related_slot_id === slot.id
																			? `${conflict.class_code} ${conflict.subject_name}`
																			: (conflict.related_slot_id ? `${conflict.related_class_code} ${conflict.related_subject_name}` : conflict.message)}
																	</p>
																{/each}
															</div>
														{/if}
														<div class="mt-4 flex flex-wrap justify-end gap-2">
															<Button variant="outline" size="xs" href={resolve(`/akademik/rombel/${slot.class_id}` as '/')}>
																<ExternalLink class="mr-1 size-3.5" />
																Detail Rombel
															</Button>
															<Button variant="outline" size="xs" onclick={() => editSlot(slot)} disabled={slotSaveBusy || deleteBusyId !== ''}>
																<Edit3 class="mr-1 size-3.5" />
																Edit
															</Button>
															<LoadingButton
																variant="destructive"
																size="xs"
																loading={deleteBusyId === slot.id}
																loadingLabel="Hapus..."
																disabled={slotSaveBusy || (deleteBusyId !== '' && deleteBusyId !== slot.id)}
																onclick={() => void deleteSlot(slot)}
															>
																<Trash2 class="mr-1 size-3.5" />
																Hapus
															</LoadingButton>
														</div>
													</div>
												{/each}
											</div>
										{:else}
											<div class="rounded-lg border border-dashed border-border px-4 py-5 text-sm text-muted-foreground">
												Tidak ada jam pelajaran pada hari ini.
											</div>
										{/if}
									</div>
								{/each}
							{/if}
						</Card.Content>
					</Card.Root>

					<Card.Root>
						<Card.Header>
							<Card.Title class="text-base">Daftar jadwal yang perlu diperiksa</Card.Title>
							<Card.Description>{conflicts.length} jadwal perlu diperiksa pada tahun ajaran aktif.</Card.Description>
						</Card.Header>
						<Card.Content class="p-0">
							{#if conflicts.length === 0}
								<div class="p-4">
									<EmptyStatePanel compact title="Tidak ada jadwal yang perlu diperiksa" description="Jadwal aktif belum memiliki bentrok rombel, guru, atau ruang." />
								</div>
							{:else}
								<div class="overflow-x-auto">
									<Table.Root>
										<Table.Header>
											<Table.Row>
												<Table.Head>Jenis</Table.Head>
												<Table.Head>Jadwal utama</Table.Head>
												<Table.Head>Jadwal terkait</Table.Head>
												<Table.Head>Hari/Jam</Table.Head>
											</Table.Row>
										</Table.Header>
										<Table.Body>
											{#each conflicts as conflict (`${conflict.conflict_type}-${conflict.slot_id}-${conflict.related_slot_id ?? 'self'}`)}
												<Table.Row>
													<Table.Cell>
														<Badge variant="destructive">{conflictTypeLabel(conflict.conflict_type)}</Badge>
													</Table.Cell>
													<Table.Cell>
														<p class="font-medium">{conflict.class_code} - {conflict.subject_name}</p>
														<p class="text-xs text-muted-foreground">{conflict.teacher_name}</p>
													</Table.Cell>
													<Table.Cell>
														{#if conflict.related_slot_id}
															<p class="font-medium">{conflict.related_class_code} - {conflict.related_subject_name}</p>
															<p class="text-xs text-muted-foreground">{conflict.related_teacher_name}</p>
														{:else}
															<span class="text-muted-foreground">{conflict.message}</span>
														{/if}
													</Table.Cell>
													<Table.Cell class="text-muted-foreground">
														{dayLabels[conflict.day_of_week] ?? `Hari ${conflict.day_of_week}`} {fmtTime(conflict.start_time)}-{fmtTime(conflict.end_time)}
													</Table.Cell>
												</Table.Row>
											{/each}
										</Table.Body>
									</Table.Root>
								</div>
							{/if}
						</Card.Content>
					</Card.Root>
				</div>

				<Card.Root id="timetable-slot-form">
					<Card.Header>
						<div class="flex items-start justify-between gap-3">
							<div>
								<Card.Title class="text-base">{editingSlot ? 'Edit jam pelajaran' : 'Tambah jam pelajaran'}</Card.Title>
								<Card.Description>{selectedAssignment ? assignmentLabel(selectedAssignment) : 'Pilih rombel dan guru mapel aktif.'}</Card.Description>
							</div>
							{#if editingSlot}
								<Button variant="ghost" size="icon-sm" aria-label="Batalkan edit jam pelajaran" onclick={() => resetForm(editingSlot?.class_id ?? formClassId)} disabled={slotSaveBusy}>
									<X class="size-4" />
								</Button>
							{/if}
						</div>
					</Card.Header>
					<Card.Content class="space-y-4">
						{#if formError}
							<div class="rounded-md border border-destructive/30 bg-destructive/10 p-3 text-sm text-destructive">{formError}</div>
						{/if}
						<div class="space-y-1.5">
							<label for="slot-class" class="text-sm font-medium">Rombel</label>
							<select
								id="slot-class"
								bind:value={formClassId}
								onchange={onFormClassChange}
								disabled={Boolean(editingSlot) || slotSaveBusy}
								class="h-10 w-full rounded-md border border-input bg-background px-3 text-sm"
							>
								<option value="">Pilih rombel</option>
								{#each classes as klass (klass.id)}
									<option value={klass.id}>{klass.code} - {klass.name}</option>
								{/each}
							</select>
						</div>
						<div class="space-y-1.5">
							<label for="slot-assignment" class="text-sm font-medium">Guru Mapel</label>
							<select
								id="slot-assignment"
								bind:value={formAssignmentId}
								disabled={formAssignments.length === 0 || slotSaveBusy}
								class="h-10 w-full rounded-md border border-input bg-background px-3 text-sm"
							>
								<option value="">Pilih guru mapel</option>
								{#each formAssignments as assignment (assignment.id)}
									<option value={assignment.id}>{assignmentLabel(assignment)}</option>
								{/each}
							</select>
							{#if formClassId && formAssignments.length === 0}
								<p class="text-xs text-destructive">Rombel ini belum punya guru mapel aktif.</p>
							{/if}
						</div>
						<div class="grid gap-3 sm:grid-cols-2">
							<div class="space-y-1.5">
								<label for="slot-day" class="text-sm font-medium">Hari</label>
								<select id="slot-day" bind:value={formDay} disabled={slotSaveBusy} class="h-10 w-full rounded-md border border-input bg-background px-3 text-sm">
									{#each Object.entries(dayLabels) as [day, label] (`slot-day-${day}`)}
										<option value={day}>{label}</option>
									{/each}
								</select>
							</div>
							<div class="space-y-1.5">
								<label for="slot-room" class="text-sm font-medium">Ruang</label>
								<Input id="slot-room" bind:value={formRoom} placeholder="Mis. VII A / Lab IPA" disabled={slotSaveBusy} />
							</div>
						</div>
						<div class="grid gap-3 sm:grid-cols-2">
							<div class="space-y-1.5">
								<label for="slot-start" class="text-sm font-medium">Mulai</label>
								<Input id="slot-start" type="time" bind:value={formStart} disabled={slotSaveBusy} />
							</div>
							<div class="space-y-1.5">
								<label for="slot-end" class="text-sm font-medium">Selesai</label>
								<Input id="slot-end" type="time" bind:value={formEnd} disabled={slotSaveBusy} />
							</div>
						</div>
						<div class="space-y-1.5">
							<label for="slot-notes" class="text-sm font-medium">Catatan</label>
							<Textarea id="slot-notes" bind:value={formNotes} rows={3} disabled={slotSaveBusy} />
						</div>
						<div class="flex flex-wrap justify-end gap-2">
							{#if editingSlot}
								<Button variant="outline" onclick={() => resetForm(editingSlot?.class_id ?? formClassId)} disabled={slotSaveBusy}>Batal</Button>
							{/if}
							<LoadingButton loading={slotSaveBusy} loadingLabel="Menyimpan..." disabled={!canSaveSlot} onclick={() => void saveSlot()}>
								{editingSlot ? 'Simpan perubahan' : 'Tambah jam pelajaran'}
							</LoadingButton>
						</div>
					</Card.Content>
				</Card.Root>
			</div>
		{/if}
	</AsyncContent>
</div>
