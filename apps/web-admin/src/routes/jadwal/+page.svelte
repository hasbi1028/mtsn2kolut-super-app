<script lang="ts">
	import { onMount } from 'svelte';
	import * as Card from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import EmptyStatePanel from '$lib/components/EmptyStatePanel.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';

	type UserData = {
		role?: string;
		roles?: string[];
	};

	type TimetableEntry = {
		id: string;
		day_of_week: number;
		start_time: string;
		end_time: string;
		room_label: string;
		notes: string;
		class_name: string;
		class_code: string;
		subject_name: string;
		subject_code: string;
		teacher_name: string;
	};

	type StudentPortalData = {
		student: {
			nama: string;
			nis: string;
			class_name: string;
			class_code: string;
			status: string;
		};
		parents: Array<{ id: string; nama: string; phone: string }>;
		timetable: TimetableEntry[];
	};

	type ParentPortalData = {
		parent: { nama: string; phone: string; address: string };
		children: Array<{ id: string; nama: string; nis: string; class_name: string }>;
		timetable: Array<TimetableEntry & { student_id: string; student_name: string }>;
	};

	type ScheduleViewData =
		| { kind: 'guru'; timetable: TimetableEntry[] }
		| { kind: 'siswa'; portal: StudentPortalData | null }
		| { kind: 'ortu'; portal: ParentPortalData | null }
		| { kind: 'empty' };

	type ApiEnvelope<T> = {
		data?: T;
		error?: string;
		message?: string;
	};

	let { data }: { data: { user?: UserData } } = $props();

	let schedulePromise = $state<Promise<ScheduleViewData> | null>(null);
	let studentPortal = $state<StudentPortalData | null>(null);
	let parentPortal = $state<ParentPortalData | null>(null);
	let guruTimetable = $state<TimetableEntry[]>([]);
	let selectedDay = $state('all');
	let selectedGuruClass = $state('all');
	let selectedGuruSubject = $state('all');

	const roles = $derived(data.user?.roles || (data.user?.role ? [data.user.role] : []));
	const isGuru = $derived(roles.includes('guru'));
	const isSiswa = $derived(roles.includes('siswa'));
	const isParent = $derived(roles.includes('ortu'));

	const dayLabels: Record<number, string> = {
		1: 'Senin',
		2: 'Selasa',
		3: 'Rabu',
		4: 'Kamis',
		5: 'Jumat',
		6: 'Sabtu',
	};

	function fmtTime(value: string) {
		return value?.slice(0, 5) || '—';
	}

	function isRecord(value: unknown): value is Record<string, unknown> {
		return typeof value === 'object' && value !== null;
	}

	function apiErrorMessage(payload: unknown) {
		if (!isRecord(payload)) return '';
		const error = payload.error;
		if (typeof error === 'string' && error.trim()) return error;
		const message = payload.message;
		if (typeof message === 'string' && message.trim()) return message;
		return '';
	}

	async function readApi<T>(response: Response, fallbackMessage: string): Promise<T> {
		const payload = (await response.json().catch(() => null)) as ApiEnvelope<T> | T | null;
		const message = apiErrorMessage(payload);
		if (!response.ok) throw new Error(message || fallbackMessage);
		if (isRecord(payload) && typeof payload.error === 'string' && payload.error.trim()) throw new Error(payload.error);
		if (isRecord(payload) && 'data' in payload) {
			const envelope = payload as ApiEnvelope<T>;
			if (envelope.data === undefined) throw new Error(fallbackMessage);
			return envelope.data;
		}
		if (payload === null) throw new Error(fallbackMessage);
		return payload as T;
	}

	function csvEscape(value: string | number | null | undefined) {
		const text = String(value ?? '');
		if (text.includes('"') || text.includes(',') || text.includes('\n')) {
			return `"${text.replaceAll('"', '""')}"`;
		}
		return text;
	}

	function downloadCsv(filename: string, rows: Array<Array<string | number | null | undefined>>) {
		const csv = rows.map((row) => row.map(csvEscape).join(',')).join('\n');
		const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' });
		const href = URL.createObjectURL(blob);
		const link = document.createElement('a');
		link.href = href;
		link.download = filename;
		link.click();
		URL.revokeObjectURL(href);
	}

	function groupByDay<T extends { day_of_week: number }>(items: T[]) {
		return Object.entries(dayLabels).map(([day, label]) => ({
			day: Number(day),
			label,
			slots: items.filter((item) => item.day_of_week === Number(day)),
		}));
	}

	function applyDayFilter<T extends { day_of_week: number }>(items: T[]) {
		if (selectedDay === 'all') return items;
		const day = Number(selectedDay);
		return items.filter((item) => item.day_of_week === day);
	}

	const guruClassOptions = $derived.by(() =>
		Array.from(new Map(guruTimetable.map((slot) => [slot.class_code, { code: slot.class_code, name: slot.class_name }])).values())
	);

	const guruSubjectOptions = $derived.by(() =>
		Array.from(new Map(guruTimetable.map((slot) => [slot.subject_code, { code: slot.subject_code, name: slot.subject_name }])).values())
	);

	const visibleGuruTimetable = $derived.by(() => {
		const dayFiltered = applyDayFilter(guruTimetable);
		return dayFiltered.filter((slot) => {
			if (selectedGuruClass !== 'all' && slot.class_code !== selectedGuruClass) return false;
			if (selectedGuruSubject !== 'all' && slot.subject_code !== selectedGuruSubject) return false;
			return true;
		});
	});
	const visibleStudentTimetable = $derived.by(() => applyDayFilter(studentPortal?.timetable ?? []));
	const visibleParentTimetable = $derived.by(() => applyDayFilter(parentPortal?.timetable ?? []));

	const guruGroups = $derived.by(() => groupByDay(visibleGuruTimetable));
	const studentGroups = $derived.by(() => groupByDay(visibleStudentTimetable));
	const parentGroups = $derived.by(() => {
		const portal = parentPortal;
		if (!portal) return [];
		return portal.children.map((child) => ({
			child,
			days: groupByDay(visibleParentTimetable.filter((slot) => slot.student_id === child.id)),
		}));
	});

	const guruSummary = $derived.by(() => ({
		totalSlots: guruTimetable.length,
		totalClasses: new Set(guruTimetable.map((slot) => slot.class_code)).size,
		activeDays: guruGroups.filter((group) => group.slots.length > 0).length,
	}));

	const studentSummary = $derived.by(() => ({
		totalSlots: studentPortal?.timetable.length ?? 0,
		activeDays: studentGroups.filter((group) => group.slots.length > 0).length,
		totalParents: studentPortal?.parents.length ?? 0,
	}));

	const parentSummary = $derived.by(() => ({
		totalChildren: parentPortal?.children.length ?? 0,
		totalSlots: parentPortal?.timetable.length ?? 0,
		activeChildren: parentGroups.filter((group) => group.days.some((day) => day.slots.length > 0)).length,
	}));

	function applySchedule(schedule: ScheduleViewData) {
		if (schedule.kind === 'guru') {
			guruTimetable = schedule.timetable ?? [];
			studentPortal = null;
			parentPortal = null;
			return;
		}
		if (schedule.kind === 'siswa') {
			studentPortal = schedule.portal;
			guruTimetable = [];
			parentPortal = null;
			return;
		}
		if (schedule.kind === 'ortu') {
			parentPortal = schedule.portal;
			guruTimetable = [];
			studentPortal = null;
			return;
		}
		guruTimetable = [];
		studentPortal = null;
		parentPortal = null;
	}

	async function fetchSchedule(): Promise<ScheduleViewData> {
		if (isGuru) {
			const res = await fetch('/api/portal/guru/timetable');
			const payload = await readApi<{ timetable: TimetableEntry[] }>(res, 'Gagal memuat jadwal mengajar.');
			return { kind: 'guru', timetable: payload.timetable ?? [] };
		}

		if (isSiswa) {
			const res = await fetch('/api/portal/student/me');
			const portal = await readApi<StudentPortalData>(res, 'Gagal memuat jadwal pelajaran.');
			return { kind: 'siswa', portal };
		}

		if (isParent) {
			const res = await fetch('/api/portal/parent/me');
			const portal = await readApi<ParentPortalData>(res, 'Gagal memuat jadwal anak.');
			return { kind: 'ortu', portal };
		}

		return { kind: 'empty' };
	}

	function load() {
		schedulePromise = fetchSchedule().then((schedule) => {
			applySchedule(schedule);
			return schedule;
		});
	}

	function retrySchedule(reset?: () => void) {
		reset?.();
		load();
	}

	function scheduleErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Gagal memuat jadwal. Coba lagi untuk mengambil data terbaru.';
	}

	function handleScheduleRenderError(error: unknown) {
		console.error('Schedule render failed', error);
	}

	onMount(() => {
		void load();
	});

	function exportGuruTimetable() {
		downloadCsv('jadwal-guru.csv', [
			['Hari', 'Mulai', 'Selesai', 'Kelas', 'Mapel', 'Guru', 'Ruang', 'Catatan'],
			...visibleGuruTimetable.map((slot) => [
				dayLabels[slot.day_of_week] ?? `Hari ${slot.day_of_week}`,
				fmtTime(slot.start_time),
				fmtTime(slot.end_time),
				slot.class_name,
				slot.subject_name,
				slot.teacher_name,
				slot.room_label,
				slot.notes,
			]),
		]);
	}

	function exportStudentTimetable() {
		const timetable = visibleStudentTimetable;
		downloadCsv('jadwal-siswa.csv', [
			['Hari', 'Mulai', 'Selesai', 'Kelas', 'Mapel', 'Guru', 'Ruang', 'Catatan'],
			...timetable.map((slot) => [
				dayLabels[slot.day_of_week] ?? `Hari ${slot.day_of_week}`,
				fmtTime(slot.start_time),
				fmtTime(slot.end_time),
				slot.class_name,
				slot.subject_name,
				slot.teacher_name,
				slot.room_label,
				slot.notes,
			]),
		]);
	}

	function exportParentTimetable() {
		const timetable = visibleParentTimetable;
		downloadCsv('jadwal-anak.csv', [
			['Nama Anak', 'Hari', 'Mulai', 'Selesai', 'Kelas', 'Mapel', 'Guru', 'Ruang', 'Catatan'],
			...timetable.map((slot) => [
				slot.student_name,
				dayLabels[slot.day_of_week] ?? `Hari ${slot.day_of_week}`,
				fmtTime(slot.start_time),
				fmtTime(slot.end_time),
				slot.class_name,
				slot.subject_name,
				slot.teacher_name,
				slot.room_label,
				slot.notes,
			]),
		]);
	}
</script>

<svelte:head>
	<title>Jadwal — MTSN 2 Kolut</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
		{#if isGuru}
			<div>
				<h1 class="text-2xl font-semibold text-slate-800">Jadwal Mengajar</h1>
				<p class="mt-1 text-sm text-muted-foreground">Lihat slot mengajar mingguan Anda dalam tampilan yang lebih penuh daripada ringkasan dashboard.</p>
			</div>
			<Button variant="outline" onclick={exportGuruTimetable} disabled={guruTimetable.length === 0}>Ekspor CSV</Button>
		{:else if isSiswa}
			<div>
				<h1 class="text-2xl font-semibold text-slate-800">Jadwal Pelajaran</h1>
				<p class="mt-1 text-sm text-muted-foreground">Pantau jadwal pelajaran mingguan berdasarkan kelas yang sedang aktif.</p>
			</div>
			<Button variant="outline" onclick={exportStudentTimetable} disabled={(studentPortal?.timetable.length ?? 0) === 0}>Ekspor CSV</Button>
		{:else}
			<div>
				<h1 class="text-2xl font-semibold text-slate-800">Jadwal Anak</h1>
				<p class="mt-1 text-sm text-muted-foreground">Lihat ringkasan jadwal pelajaran setiap anak yang sudah terhubung ke akun orang tua ini.</p>
			</div>
			<Button variant="outline" onclick={exportParentTimetable} disabled={(parentPortal?.timetable.length ?? 0) === 0}>Ekspor CSV</Button>
		{/if}
	</div>

	<Card.Root class="border-slate-200">
		<Card.Content class="flex flex-col gap-3 p-4 sm:flex-row sm:items-end sm:justify-between">
			<div>
				<p class="text-xs font-medium uppercase tracking-[0.2em] text-slate-500">Filter Hari</p>
				<p class="mt-1 text-sm text-slate-500">Fokuskan tampilan dan ekspor ke satu hari tertentu bila diperlukan.</p>
			</div>
			<div class="grid w-full gap-3 sm:w-auto sm:grid-cols-2 lg:grid-cols-3">
				<div class="w-full sm:w-56">
					<label for="day-filter" class="mb-1 block text-xs font-medium text-slate-600">Hari</label>
					<select id="day-filter" bind:value={selectedDay} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
						<option value="all">Semua Hari</option>
						{#each Object.entries(dayLabels) as [day, label] (`day-option-${day}`)}
							<option value={day}>{label}</option>
						{/each}
					</select>
				</div>
				{#if isGuru}
					<div class="w-full sm:w-56">
						<label for="guru-class-filter" class="mb-1 block text-xs font-medium text-slate-600">Kelas</label>
						<select id="guru-class-filter" bind:value={selectedGuruClass} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
							<option value="all">Semua Kelas</option>
							{#each guruClassOptions as option (option.code)}
								<option value={option.code}>{option.name}</option>
							{/each}
						</select>
					</div>
					<div class="w-full sm:w-56">
						<label for="guru-subject-filter" class="mb-1 block text-xs font-medium text-slate-600">Mapel</label>
						<select id="guru-subject-filter" bind:value={selectedGuruSubject} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
							<option value="all">Semua Mapel</option>
							{#each guruSubjectOptions as option (option.code)}
								<option value={option.code}>{option.name}</option>
							{/each}
						</select>
					</div>
				{/if}
			</div>
		</Card.Content>
	</Card.Root>

	<AsyncContent promise={schedulePromise} onerror={handleScheduleRenderError}>
		{#snippet pending()}
			<div class="grid gap-4 md:grid-cols-3">
				{#each Array.from({ length: 3 }) as _, i (`loading-summary-${i}`)}
					<Card.Root class="border-slate-200">
						<Card.Content class="space-y-3 p-5">
							<Skeleton class="h-4 w-24" />
							<Skeleton class="h-8 w-16" />
							<Skeleton class="h-3 w-32" />
						</Card.Content>
					</Card.Root>
				{/each}
			</div>

			<Card.Root class="border-slate-200">
				<Card.Content class="space-y-4 p-5">
					{#each Array.from({ length: 4 }) as _, i (`loading-day-${i}`)}
						<div class="space-y-3">
							<Skeleton class="h-5 w-32" />
							<div class="grid gap-3 lg:grid-cols-2">
								<Skeleton class="h-28 w-full" />
								<Skeleton class="h-28 w-full" />
							</div>
						</div>
					{/each}
				</Card.Content>
			</Card.Root>
		{/snippet}

		{#snippet failed(error, reset)}
			<RecoveryPanel
				title="Jadwal Belum Tersaji"
				message={scheduleErrorMessage(error)}
				onRetry={() => retrySchedule(reset)}
			/>
		{/snippet}

		{#snippet children(value)}
			{@const loadedSchedule = value as ScheduleViewData}
			{#if loadedSchedule.kind === 'empty'}
				<EmptyStatePanel compact title="Jadwal belum tersedia" description="Akun ini belum memiliki peran portal jadwal yang aktif." />
			{:else if isGuru}
		<div class="grid gap-4 md:grid-cols-3">
			<Card.Root class="border-slate-200"><Card.Content class="p-5"><p class="text-xs font-medium uppercase tracking-[0.2em] text-slate-500">Total Slot</p><p class="mt-2 text-3xl font-semibold text-slate-900">{guruSummary.totalSlots}</p></Card.Content></Card.Root>
			<Card.Root class="border-slate-200"><Card.Content class="p-5"><p class="text-xs font-medium uppercase tracking-[0.2em] text-slate-500">Kelas Diajar</p><p class="mt-2 text-3xl font-semibold text-slate-900">{guruSummary.totalClasses}</p></Card.Content></Card.Root>
			<Card.Root class="border-slate-200"><Card.Content class="p-5"><p class="text-xs font-medium uppercase tracking-[0.2em] text-slate-500">Hari Aktif</p><p class="mt-2 text-3xl font-semibold text-slate-900">{guruSummary.activeDays}</p></Card.Content></Card.Root>
		</div>

		<Card.Root class="border-slate-200">
			<Card.Header>
				<Card.Title class="text-base">Minggu Mengajar</Card.Title>
				<Card.Description>Disusun per hari agar lebih cepat dipindai saat mempersiapkan pembelajaran.</Card.Description>
			</Card.Header>
			<Card.Content class="space-y-5">
				{#if visibleGuruTimetable.length > 0}
					{#each guruGroups as group (group.day)}
						<div class="space-y-3">
							<div class="flex items-center justify-between gap-3">
								<h2 class="text-sm font-semibold text-slate-900">{group.label}</h2>
								<Badge variant="outline">{group.slots.length} slot</Badge>
							</div>
							{#if group.slots.length > 0}
								<div class="grid gap-3 lg:grid-cols-2">
									{#each group.slots as slot (slot.id)}
										<div class="rounded-2xl border border-slate-200 bg-slate-50 p-4">
											<p class="text-sm font-semibold text-slate-900">{slot.subject_name}</p>
											<p class="mt-1 text-xs text-slate-500">{slot.class_name}</p>
											<p class="mt-2 text-sm text-slate-600">{fmtTime(slot.start_time)}–{fmtTime(slot.end_time)} · {slot.room_label || 'Ruang belum diisi'}</p>
											{#if slot.notes}
												<p class="mt-2 text-xs text-slate-500">{slot.notes}</p>
											{/if}
										</div>
									{/each}
								</div>
							{:else}
								<div class="rounded-xl border border-dashed border-slate-300 px-4 py-5 text-sm text-slate-500">Tidak ada slot mengajar pada hari ini.</div>
							{/if}
						</div>
					{/each}
				{:else}
					<EmptyStatePanel compact title="Jadwal mengajar belum tersedia" description="Jadwal akan muncul di sini setelah operator akademik menyusun slot kelas-mapel-guru." />
				{/if}
			</Card.Content>
		</Card.Root>
	{:else if isSiswa}
		<div class="grid gap-4 md:grid-cols-3">
			<Card.Root class="border-slate-200"><Card.Content class="p-5"><p class="text-xs font-medium uppercase tracking-[0.2em] text-slate-500">Total Slot</p><p class="mt-2 text-3xl font-semibold text-slate-900">{studentSummary.totalSlots}</p></Card.Content></Card.Root>
			<Card.Root class="border-slate-200"><Card.Content class="p-5"><p class="text-xs font-medium uppercase tracking-[0.2em] text-slate-500">Hari Aktif</p><p class="mt-2 text-3xl font-semibold text-slate-900">{studentSummary.activeDays}</p></Card.Content></Card.Root>
			<Card.Root class="border-slate-200"><Card.Content class="p-5"><p class="text-xs font-medium uppercase tracking-[0.2em] text-slate-500">Kontak Wali</p><p class="mt-2 text-3xl font-semibold text-slate-900">{studentSummary.totalParents}</p></Card.Content></Card.Root>
		</div>

		<Card.Root class="border-slate-200">
			<Card.Header>
				<Card.Title class="text-base">Informasi Siswa</Card.Title>
				<Card.Description>{studentPortal?.student.nama} · {studentPortal?.student.class_name || 'Belum ada kelas'}</Card.Description>
			</Card.Header>
			<Card.Content class="grid gap-3 md:grid-cols-3">
				<div class="rounded-xl border border-slate-200 bg-slate-50 p-4"><p class="text-xs font-medium uppercase tracking-[0.2em] text-slate-500">NIS</p><p class="mt-2 text-sm font-semibold text-slate-900">{studentPortal?.student.nis || '—'}</p></div>
				<div class="rounded-xl border border-slate-200 bg-slate-50 p-4"><p class="text-xs font-medium uppercase tracking-[0.2em] text-slate-500">Kelas</p><p class="mt-2 text-sm font-semibold text-slate-900">{studentPortal?.student.class_name || 'Belum ada kelas'}</p></div>
				<div class="rounded-xl border border-slate-200 bg-slate-50 p-4"><p class="text-xs font-medium uppercase tracking-[0.2em] text-slate-500">Status</p><p class="mt-2 text-sm font-semibold text-slate-900">{studentPortal?.student.status || '—'}</p></div>
			</Card.Content>
		</Card.Root>

		<Card.Root class="border-slate-200">
			<Card.Header>
				<Card.Title class="text-base">Minggu Pelajaran</Card.Title>
				<Card.Description>Jadwal pelajaran disusun per hari untuk memudahkan persiapan belajar.</Card.Description>
			</Card.Header>
			<Card.Content class="space-y-5">
				{#if visibleStudentTimetable.length > 0}
					{#each studentGroups as group (group.day)}
						<div class="space-y-3">
							<div class="flex items-center justify-between gap-3">
								<h2 class="text-sm font-semibold text-slate-900">{group.label}</h2>
								<Badge variant="outline">{group.slots.length} slot</Badge>
							</div>
							{#if group.slots.length > 0}
								<div class="grid gap-3 lg:grid-cols-2">
									{#each group.slots as slot (slot.id)}
										<div class="rounded-2xl border border-slate-200 bg-slate-50 p-4">
											<div class="flex items-start justify-between gap-3">
												<div>
													<p class="text-sm font-semibold text-slate-900">{slot.subject_name}</p>
													<p class="mt-1 text-xs text-slate-500">{slot.teacher_name}</p>
												</div>
												<Badge variant="outline">{slot.class_code}</Badge>
											</div>
											<p class="mt-2 text-sm text-slate-600">{fmtTime(slot.start_time)}–{fmtTime(slot.end_time)} · {slot.room_label || 'Ruang belum diisi'}</p>
											{#if slot.notes}
												<p class="mt-2 text-xs text-slate-500">{slot.notes}</p>
											{/if}
										</div>
									{/each}
								</div>
							{:else}
								<div class="rounded-xl border border-dashed border-slate-300 px-4 py-5 text-sm text-slate-500">Tidak ada pelajaran pada hari ini.</div>
							{/if}
						</div>
					{/each}
				{:else}
					<EmptyStatePanel compact title="Jadwal pelajaran belum tersedia" description="Jadwal akan tampil setelah operator akademik menyusun slot kelas untuk siswa ini." />
				{/if}
			</Card.Content>
		</Card.Root>
	{:else}
		<div class="grid gap-4 md:grid-cols-3">
			<Card.Root class="border-slate-200"><Card.Content class="p-5"><p class="text-xs font-medium uppercase tracking-[0.2em] text-slate-500">Anak Terhubung</p><p class="mt-2 text-3xl font-semibold text-slate-900">{parentSummary.totalChildren}</p></Card.Content></Card.Root>
			<Card.Root class="border-slate-200"><Card.Content class="p-5"><p class="text-xs font-medium uppercase tracking-[0.2em] text-slate-500">Total Slot</p><p class="mt-2 text-3xl font-semibold text-slate-900">{parentSummary.totalSlots}</p></Card.Content></Card.Root>
			<Card.Root class="border-slate-200"><Card.Content class="p-5"><p class="text-xs font-medium uppercase tracking-[0.2em] text-slate-500">Anak dengan Jadwal</p><p class="mt-2 text-3xl font-semibold text-slate-900">{parentSummary.activeChildren}</p></Card.Content></Card.Root>
		</div>

		<Card.Root class="border-slate-200">
			<Card.Header>
				<Card.Title class="text-base">Jadwal per Anak</Card.Title>
				<Card.Description>Setiap anak ditampilkan terpisah agar wali lebih mudah memeriksa ritme belajar mingguan.</Card.Description>
			</Card.Header>
			<Card.Content class="space-y-5">
				{#if visibleParentTimetable.length > 0 && parentGroups.length > 0}
					{#each parentGroups as group (`child-${group.child.id}`)}
						<div class="rounded-2xl border border-slate-200 bg-slate-50 p-4">
							<div class="flex items-start justify-between gap-3">
								<div>
									<p class="text-sm font-semibold text-slate-900">{group.child.nama}</p>
									<p class="mt-1 text-xs text-slate-500">{group.child.class_name || 'Belum ada kelas'}</p>
								</div>
								<Badge variant="outline">{group.days.reduce((sum, day) => sum + day.slots.length, 0)} slot</Badge>
							</div>

							<div class="mt-4 space-y-4">
								{#each group.days as day (day.day)}
									<div class="space-y-2">
										<div class="flex items-center justify-between gap-3">
											<h2 class="text-xs font-semibold uppercase tracking-[0.2em] text-slate-500">{day.label}</h2>
											<Badge variant="outline">{day.slots.length}</Badge>
										</div>
										{#if day.slots.length > 0}
											<div class="grid gap-3 lg:grid-cols-2">
												{#each day.slots as slot (slot.id)}
													<div class="rounded-xl border border-slate-200 bg-white p-3">
														<p class="text-sm font-semibold text-slate-900">{slot.subject_name}</p>
														<p class="mt-1 text-xs text-slate-500">{slot.teacher_name}</p>
														<p class="mt-2 text-sm text-slate-600">{fmtTime(slot.start_time)}–{fmtTime(slot.end_time)} · {slot.room_label || 'Ruang belum diisi'}</p>
														{#if slot.notes}
															<p class="mt-2 text-xs text-slate-500">{slot.notes}</p>
														{/if}
													</div>
												{/each}
											</div>
										{:else}
											<div class="rounded-xl border border-dashed border-slate-300 bg-white px-4 py-4 text-sm text-slate-500">Tidak ada jadwal untuk hari ini.</div>
										{/if}
									</div>
								{/each}
							</div>
						</div>
					{/each}
				{:else}
					<EmptyStatePanel compact title="Jadwal anak belum tersedia" description="Jadwal akan muncul di sini setelah anak terhubung dan operator akademik menyusun slot kelasnya." />
				{/if}
			</Card.Content>
		</Card.Root>
			{/if}
		{/snippet}
	</AsyncContent>
</div>
