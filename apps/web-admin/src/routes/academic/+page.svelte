<script lang="ts">
	import { onMount } from 'svelte';
	import * as Card from '$lib/components/ui/card';
	import * as Tabs from '$lib/components/ui/tabs';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Badge } from '$lib/components/ui/badge';
	import { toast } from '$lib/components/ui/sonner';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import EmptyStatePanel from '$lib/components/EmptyStatePanel.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';

	type AcademicYear = {
		id: string; name: string; start_date: string; end_date: string;
		is_active: boolean; created_at: string;
	};
	type SchoolClass = {
		id: string; code: string; name: string; level: string;
		is_active: boolean; academic_year_id: string; academic_year_name: string;
	};
	type Subject = { id: string; code: string; name: string; is_active: boolean; };
	type Assignment = {
		id: string; class_id: string; class_name: string; class_code: string;
		subject_id: string; subject_name: string; subject_code: string;
		teacher_employee_id: string; teacher_name: string;
	};
	type TimetableSlot = {
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
		subject_id: string;
		subject_name: string;
		subject_code: string;
		teacher_employee_id: string;
		teacher_name: string;
	};

	let years = $state<AcademicYear[]>([]);
	let classes = $state<SchoolClass[]>([]);
	let subjects = $state<Subject[]>([]);
	let assignments = $state<Assignment[]>([]);
	let timetableSlots = $state<TimetableSlot[]>([]);
	let loading = $state(true);
	let error = $state('');

	// Year form
	let yearName = $state('');
	let yearStart = $state('');
	let yearEnd = $state('');
	let yearActive = $state(false);
	let yearBusy = $state(false);

	// Class form
	let className = $state('');
	let classCode = $state('');
	let classLevel = $state('');
	let classYearId = $state('');
	let classActive = $state(true);
	let classBusy = $state(false);

	// Subject form
	let subjectName = $state('');
	let subjectCode = $state('');
	let subjectActive = $state(true);
	let subjectBusy = $state(false);

	// Timetable form
	let timetableAssignmentId = $state('');
	let timetableDay = $state('1');
	let timetableStart = $state('');
	let timetableEnd = $state('');
	let timetableRoom = $state('');
	let timetableNotes = $state('');
	let timetableBusy = $state(false);
	let editingTimetableId = $state('');

	const dayLabels: Record<number, string> = {
		1: 'Senin',
		2: 'Selasa',
		3: 'Rabu',
		4: 'Kamis',
		5: 'Jumat',
		6: 'Sabtu',
	};

	const timetableSummary = $derived.by(() => {
		const classesCovered = new Set(timetableSlots.map((slot) => slot.class_id)).size;
		const teachersCovered = new Set(timetableSlots.map((slot) => slot.teacher_employee_id)).size;
		return {
			classesCovered,
			teachersCovered,
		};
	});

	async function load() {
		try {
			const res = await fetch('/api/academic');
			const json = await res.json();
			if (json.error) { error = json.error; return; }
			const d = json.data ?? json;
			years = d.years ?? [];
			classes = d.classes ?? [];
			subjects = d.subjects ?? [];
			assignments = d.assignments ?? [];
			timetableSlots = d.timetableSlots ?? [];
		} catch (e) {
			error = 'Gagal memuat data akademik';
		} finally {
			loading = false;
		}
	}

	function showToast(msg: string) {
		toast.success(msg);
	}

	function showError(msg: string) {
		toast.error(msg);
	}

	async function createYear() {
		if (!yearName || !yearStart || !yearEnd) return;
		yearBusy = true;
		try {
			const res = await fetch('/api/academic?entity=years', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ name: yearName, start_date: yearStart, end_date: yearEnd, is_active: yearActive }),
			});
			if (!res.ok) { const j = await res.json(); showError(j.error ?? 'Gagal'); return; }
			yearName = ''; yearStart = ''; yearEnd = ''; yearActive = false;
			showToast('Tahun ajaran berhasil ditambahkan');
			await load();
		} finally { yearBusy = false; }
	}

	async function deleteYear(id: string) {
		if (!confirm('Hapus tahun ajaran ini?')) return;
		await fetch(`/api/academic?entity=years&id=${id}`, { method: 'DELETE' });
		showToast('Tahun ajaran dihapus');
		await load();
	}

	async function createClass() {
		if (!className || !classCode || !classLevel || !classYearId) return;
		classBusy = true;
		try {
			const res = await fetch('/api/academic?entity=classes', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					name: className, code: classCode, level: classLevel,
					academic_year_id: classYearId, is_active: classActive,
				}),
			});
			if (!res.ok) { const j = await res.json(); showError(j.error ?? 'Gagal'); return; }
			className = ''; classCode = ''; classLevel = ''; classYearId = ''; classActive = true;
			showToast('Kelas berhasil ditambahkan');
			await load();
		} finally { classBusy = false; }
	}

	async function deleteClass(id: string) {
		if (!confirm('Hapus kelas ini?')) return;
		await fetch(`/api/academic?entity=classes&id=${id}`, { method: 'DELETE' });
		showToast('Kelas dihapus');
		await load();
	}

	async function createSubject() {
		if (!subjectName || !subjectCode) return;
		subjectBusy = true;
		try {
			const res = await fetch('/api/academic?entity=subjects', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ name: subjectName, code: subjectCode, is_active: subjectActive }),
			});
			if (!res.ok) { const j = await res.json(); showError(j.error ?? 'Gagal'); return; }
			subjectName = ''; subjectCode = ''; subjectActive = true;
			showToast('Mata pelajaran berhasil ditambahkan');
			await load();
		} finally { subjectBusy = false; }
	}

	async function deleteSubject(id: string) {
		if (!confirm('Hapus mata pelajaran ini?')) return;
		await fetch(`/api/academic?entity=subjects&id=${id}`, { method: 'DELETE' });
		showToast('Mata pelajaran dihapus');
		await load();
	}

	async function createTimetableSlot() {
		if (!timetableAssignmentId || !timetableStart || !timetableEnd) return;
		timetableBusy = true;
		try {
			const isEditing = Boolean(editingTimetableId);
			const res = await fetch(editingTimetableId ? `/api/academic?entity=timetables&id=${editingTimetableId}` : '/api/academic?entity=timetables', {
				method: editingTimetableId ? 'PUT' : 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					assignment_id: timetableAssignmentId,
					day_of_week: Number(timetableDay),
					start_time: timetableStart,
					end_time: timetableEnd,
					room_label: timetableRoom,
					notes: timetableNotes,
				}),
			});
			if (!res.ok) {
				const j = await res.json().catch(() => ({}));
				showError((j as { error?: string }).error ?? 'Gagal menyimpan jadwal');
				return;
			}
			resetTimetableForm();
			showToast(isEditing ? 'Slot jadwal pelajaran berhasil diperbarui' : 'Slot jadwal pelajaran berhasil ditambahkan');
			await load();
		} finally {
			timetableBusy = false;
		}
	}

	function editTimetableSlot(slot: TimetableSlot) {
		editingTimetableId = slot.id;
		timetableAssignmentId = slot.assignment_id;
		timetableDay = String(slot.day_of_week);
		timetableStart = fmtTime(slot.start_time);
		timetableEnd = fmtTime(slot.end_time);
		timetableRoom = slot.room_label ?? '';
		timetableNotes = slot.notes ?? '';
	}

	function resetTimetableForm() {
		editingTimetableId = '';
		timetableAssignmentId = '';
		timetableDay = '1';
		timetableStart = '';
		timetableEnd = '';
		timetableRoom = '';
		timetableNotes = '';
	}

	async function deleteTimetableSlot(id: string) {
		if (!confirm('Hapus slot jadwal ini?')) return;
		await fetch(`/api/academic?entity=timetables&id=${id}`, { method: 'DELETE' });
		showToast('Slot jadwal pelajaran dihapus');
		await load();
	}

	function fmtTime(value: string) {
		return value.slice(0, 5);
	}

	onMount(load);
</script>

<svelte:head><title>Data Akademik — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6">
	<div>
		<h1 class="text-2xl font-semibold text-slate-800">Data Akademik</h1>
		<p class="text-sm text-slate-500 mt-1">Kelola tahun ajaran, kelas, dan mata pelajaran</p>
	</div>

	<div class="grid gap-3 md:grid-cols-4">
		<div class="rounded-2xl border border-emerald-100 bg-emerald-50 px-4 py-4">
			<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-emerald-700">Tahun Ajaran</p>
			<p class="mt-2 text-2xl font-semibold text-slate-900">{years.length}</p>
			<p class="text-sm text-slate-600">periode akademik yang sudah tersusun</p>
		</div>
		<div class="rounded-2xl border border-sky-100 bg-sky-50 px-4 py-4">
			<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-sky-700">Kelas</p>
			<p class="mt-2 text-2xl font-semibold text-slate-900">{classes.length}</p>
			<p class="text-sm text-slate-600">rombel aktif yang siap dipakai modul lain</p>
		</div>
		<div class="rounded-2xl border border-amber-100 bg-amber-50 px-4 py-4">
			<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-amber-700">Mata Pelajaran</p>
			<p class="mt-2 text-2xl font-semibold text-slate-900">{subjects.length}</p>
			<p class="text-sm text-slate-600">mapel inti untuk jadwal, nilai, dan CBT</p>
		</div>
		<div class="rounded-2xl border border-violet-100 bg-violet-50 px-4 py-4">
			<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-violet-700">Slot Jadwal</p>
			<p class="mt-2 text-2xl font-semibold text-slate-900">{timetableSlots.length}</p>
			<p class="text-sm text-slate-600">jam pelajaran yang sudah disusun per kelas-mapel</p>
		</div>
	</div>

	{#if error}
		<RecoveryPanel message={error} onRetry={load} />
	{/if}

	{#if loading}
		<div class="space-y-4">
			<div class="overflow-x-auto pb-1">
				<div class="flex min-w-max gap-2">
					<Skeleton class="h-9 w-36 rounded-full" />
					<Skeleton class="h-9 w-28 rounded-full" />
					<Skeleton class="h-9 w-40 rounded-full" />
				</div>
			</div>
			<div class="grid gap-4 lg:grid-cols-3">
				<div class="space-y-4 lg:col-span-2">
					<Skeleton class="h-12 w-full" />
					<Skeleton class="h-14 w-full" />
					<Skeleton class="h-14 w-full" />
					<Skeleton class="h-14 w-full" />
				</div>
				<div class="space-y-3">
					<Skeleton class="h-10 w-full" />
					<Skeleton class="h-10 w-full" />
					<Skeleton class="h-10 w-full" />
					<Skeleton class="h-9 w-full" />
				</div>
			</div>
		</div>
	{:else}
		<Tabs.Root value="years">
			<div class="overflow-x-auto pb-1">
				<Tabs.List class="mb-4 min-w-max">
				<Tabs.Trigger value="years">Tahun Ajaran ({years.length})</Tabs.Trigger>
				<Tabs.Trigger value="classes">Kelas ({classes.length})</Tabs.Trigger>
				<Tabs.Trigger value="subjects">Mata Pelajaran ({subjects.length})</Tabs.Trigger>
				<Tabs.Trigger value="timetable">Jadwal ({timetableSlots.length})</Tabs.Trigger>
			</Tabs.List>
			</div>

			<!-- Tahun Ajaran Tab -->
			<Tabs.Content value="years">
				<div class="grid gap-4 lg:grid-cols-3">
					<div class="lg:col-span-2">
						<Card.Root>
							<Card.Header class="pb-2">
								<Card.Title class="text-base">Daftar Tahun Ajaran</Card.Title>
							</Card.Header>
							<Card.Content class="p-0 overflow-x-auto">
								<Table.Root>
									<Table.Header>
										<Table.Row>
											<Table.Head>Nama</Table.Head>
											<Table.Head>Mulai</Table.Head>
											<Table.Head>Selesai</Table.Head>
											<Table.Head>Status</Table.Head>
											<Table.Head></Table.Head>
										</Table.Row>
									</Table.Header>
									<Table.Body>
										{#each years as y (y.id)}
											<Table.Row>
												<Table.Cell class="font-medium">{y.name}</Table.Cell>
												<Table.Cell class="text-slate-500">{y.start_date?.slice(0,10)}</Table.Cell>
												<Table.Cell class="text-slate-500">{y.end_date?.slice(0,10)}</Table.Cell>
												<Table.Cell>
													{#if y.is_active}
														<Badge class="bg-emerald-100 text-emerald-700 border-emerald-200">Aktif</Badge>
													{:else}
														<Badge variant="secondary">Nonaktif</Badge>
													{/if}
												</Table.Cell>
												<Table.Cell>
													<Button variant="destructive" size="xs" onclick={() => deleteYear(y.id)}>Hapus</Button>
												</Table.Cell>
											</Table.Row>
										{:else}
											<Table.Row>
												<Table.Cell colspan={5} class="p-4">
													<EmptyStatePanel
														compact
														eyebrow="Mulai Dari Fondasi"
														title="Belum ada tahun ajaran"
														description="Tambahkan periode akademik terlebih dahulu agar kelas dan modul turunan bisa dihubungkan dengan rapi."
													/>
												</Table.Cell>
											</Table.Row>
										{/each}
									</Table.Body>
								</Table.Root>
							</Card.Content>
						</Card.Root>
					</div>

					<Card.Root>
						<Card.Header class="pb-2">
							<Card.Title class="text-base">Tambah Tahun Ajaran</Card.Title>
						</Card.Header>
						<Card.Content class="space-y-3">
							<Input id="academic-year-name" placeholder="Contoh: 2025/2026" bind:value={yearName} />
							<div>
								<label for="academic-year-start" class="text-xs text-slate-500 mb-1 block">Tanggal Mulai</label>
								<Input id="academic-year-start" type="date" bind:value={yearStart} />
							</div>
							<div>
								<label for="academic-year-end" class="text-xs text-slate-500 mb-1 block">Tanggal Selesai</label>
								<Input id="academic-year-end" type="date" bind:value={yearEnd} />
							</div>
							<label class="flex items-center gap-2 text-sm">
								<input type="checkbox" bind:checked={yearActive} class="rounded" />
								Jadikan aktif
							</label>
							<LoadingButton class="w-full" loading={yearBusy} loadingLabel="Menyimpan..." disabled={!yearName || !yearStart || !yearEnd} onclick={createYear} label="Simpan" />
						</Card.Content>
					</Card.Root>
				</div>
			</Tabs.Content>

			<Tabs.Content value="timetable">
				<div class="space-y-4">
					<div class="grid gap-3 md:grid-cols-3">
						<div class="rounded-2xl border border-emerald-100 bg-emerald-50 px-4 py-4">
							<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-emerald-700">Total Slot</p>
							<p class="mt-2 text-2xl font-semibold text-slate-900">{timetableSlots.length}</p>
							<p class="text-sm text-slate-600">seluruh jam pelajaran yang sudah dimasukkan</p>
						</div>
						<div class="rounded-2xl border border-sky-100 bg-sky-50 px-4 py-4">
							<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-sky-700">Kelas Tercakup</p>
							<p class="mt-2 text-2xl font-semibold text-slate-900">{timetableSummary.classesCovered}</p>
							<p class="text-sm text-slate-600">kelas yang sudah punya slot jadwal</p>
						</div>
						<div class="rounded-2xl border border-amber-100 bg-amber-50 px-4 py-4">
							<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-amber-700">Guru Tercakup</p>
							<p class="mt-2 text-2xl font-semibold text-slate-900">{timetableSummary.teachersCovered}</p>
							<p class="text-sm text-slate-600">guru yang sudah masuk ke jadwal pelajaran</p>
						</div>
					</div>

					<div class="grid gap-4 lg:grid-cols-3">
						<div class="lg:col-span-2">
							<Card.Root>
								<Card.Header class="pb-2">
									<Card.Title class="text-base">Daftar Slot Jadwal</Card.Title>
									<Card.Description>Kelola jam pelajaran berdasarkan assignment kelas, mapel, dan guru yang sudah aktif di sistem.</Card.Description>
								</Card.Header>
								<Card.Content class="overflow-x-auto p-0">
									<Table.Root>
										<Table.Header>
											<Table.Row>
												<Table.Head>Hari</Table.Head>
												<Table.Head>Waktu</Table.Head>
												<Table.Head>Kelas</Table.Head>
												<Table.Head>Mata Pelajaran</Table.Head>
												<Table.Head>Guru</Table.Head>
												<Table.Head>Ruang</Table.Head>
												<Table.Head></Table.Head>
											</Table.Row>
										</Table.Header>
										<Table.Body>
											{#each timetableSlots as slot (slot.id)}
												<Table.Row>
													<Table.Cell class="font-medium">{dayLabels[slot.day_of_week] ?? `Hari ${slot.day_of_week}`}</Table.Cell>
													<Table.Cell class="text-slate-600">{fmtTime(slot.start_time)}–{fmtTime(slot.end_time)}</Table.Cell>
													<Table.Cell>
														<div class="space-y-0.5">
															<p class="font-medium text-slate-900">{slot.class_name}</p>
															<p class="text-xs text-slate-500">{slot.class_code}</p>
														</div>
													</Table.Cell>
													<Table.Cell>
														<div class="space-y-0.5">
															<p class="font-medium text-slate-900">{slot.subject_name}</p>
															<p class="text-xs text-slate-500">{slot.subject_code}</p>
														</div>
													</Table.Cell>
													<Table.Cell class="text-slate-600">{slot.teacher_name}</Table.Cell>
													<Table.Cell class="text-slate-600">{slot.room_label || '—'}</Table.Cell>
													<Table.Cell>
														<div class="flex gap-2">
															<Button variant="outline" size="xs" onclick={() => editTimetableSlot(slot)}>Edit</Button>
															<Button variant="destructive" size="xs" onclick={() => deleteTimetableSlot(slot.id)}>Hapus</Button>
														</div>
													</Table.Cell>
												</Table.Row>
											{:else}
												<Table.Row>
													<Table.Cell colspan={7} class="p-4">
														<EmptyStatePanel
															title="Belum ada slot jadwal pelajaran"
															description="Tambahkan slot pertama agar kelas dan mapel mulai tersusun dalam jadwal mingguan."
															compact
														/>
													</Table.Cell>
												</Table.Row>
											{/each}
										</Table.Body>
									</Table.Root>
								</Card.Content>
							</Card.Root>
						</div>

						<Card.Root>
							<Card.Header class="pb-2">
								<Card.Title class="text-base">{editingTimetableId ? 'Edit Slot Jadwal' : 'Tambah Slot Jadwal'}</Card.Title>
								{#if editingTimetableId}
									<Card.Description>Perbarui assignment, hari, atau jam pelajaran. Sistem akan menolak bentrok dasar kelas/guru di waktu yang sama.</Card.Description>
								{/if}
							</Card.Header>
							<Card.Content class="space-y-3">
								<div>
									<label for="timetable-assignment" class="mb-1 block text-xs font-medium text-slate-600">Assignment Kelas-Mapel-Guru</label>
									<select id="timetable-assignment" bind:value={timetableAssignmentId} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
										<option value="">Pilih assignment...</option>
										{#each assignments as assignment (assignment.id)}
											<option value={assignment.id}>{assignment.class_code} · {assignment.subject_name} · {assignment.teacher_name}</option>
										{/each}
									</select>
								</div>
								<div class="grid gap-3 sm:grid-cols-2">
									<div>
										<label for="timetable-day" class="mb-1 block text-xs font-medium text-slate-600">Hari</label>
										<select id="timetable-day" bind:value={timetableDay} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
											{#each Object.entries(dayLabels) as [day, label] (`day-${day}`)}
												<option value={day}>{label}</option>
											{/each}
										</select>
									</div>
									<div>
										<label for="timetable-room" class="mb-1 block text-xs font-medium text-slate-600">Ruang</label>
										<Input id="timetable-room" bind:value={timetableRoom} placeholder="Opsional, mis. Lab IPA" />
									</div>
								</div>
								<div class="grid gap-3 sm:grid-cols-2">
									<div>
										<label for="timetable-start" class="mb-1 block text-xs font-medium text-slate-600">Mulai</label>
										<Input id="timetable-start" type="time" bind:value={timetableStart} />
									</div>
									<div>
										<label for="timetable-end" class="mb-1 block text-xs font-medium text-slate-600">Selesai</label>
										<Input id="timetable-end" type="time" bind:value={timetableEnd} />
									</div>
								</div>
								<div>
									<label for="timetable-notes" class="mb-1 block text-xs font-medium text-slate-600">Catatan</label>
									<Input id="timetable-notes" bind:value={timetableNotes} placeholder="Opsional, mis. blok bergantian dengan kelas lain" />
								</div>
								<div class="flex flex-wrap gap-2">
									<LoadingButton onclick={createTimetableSlot} loading={timetableBusy} loadingLabel="Menyimpan..." label={editingTimetableId ? 'Simpan Perubahan' : 'Tambah Slot'} disabled={!timetableAssignmentId || !timetableStart || !timetableEnd} />
									{#if editingTimetableId}
										<Button variant="outline" onclick={resetTimetableForm}>Batal Edit</Button>
									{/if}
								</div>
							</Card.Content>
						</Card.Root>
					</div>
				</div>
			</Tabs.Content>

			<!-- Kelas Tab -->
			<Tabs.Content value="classes">
				<div class="grid gap-4 lg:grid-cols-3">
					<div class="lg:col-span-2">
						<Card.Root>
							<Card.Header class="pb-2">
								<Card.Title class="text-base">Daftar Kelas</Card.Title>
							</Card.Header>
							<Card.Content class="p-0 overflow-x-auto">
								<Table.Root>
									<Table.Header>
										<Table.Row>
											<Table.Head>Kode</Table.Head>
											<Table.Head>Nama Kelas</Table.Head>
											<Table.Head>Tingkat</Table.Head>
											<Table.Head>Tahun Ajaran</Table.Head>
											<Table.Head>Status</Table.Head>
											<Table.Head></Table.Head>
										</Table.Row>
									</Table.Header>
									<Table.Body>
											{#each classes as c (c.id)}
											<Table.Row>
												<Table.Cell class="font-mono text-sm">{c.code}</Table.Cell>
												<Table.Cell class="font-medium">{c.name}</Table.Cell>
												<Table.Cell>{c.level}</Table.Cell>
												<Table.Cell class="text-slate-500">{c.academic_year_name}</Table.Cell>
												<Table.Cell>
													{#if c.is_active}
														<Badge class="bg-emerald-100 text-emerald-700 border-emerald-200">Aktif</Badge>
													{:else}
														<Badge variant="secondary">Nonaktif</Badge>
													{/if}
												</Table.Cell>
												<Table.Cell>
													<Button variant="destructive" size="xs" onclick={() => deleteClass(c.id)}>Hapus</Button>
												</Table.Cell>
											</Table.Row>
										{:else}
											<Table.Row>
												<Table.Cell colspan={6} class="p-4">
													<EmptyStatePanel
														compact
														eyebrow="Struktur Akademik"
														title="Belum ada kelas"
														description="Setelah tahun ajaran dibuat, tambahkan kelas atau rombel agar siswa, jadwal, dan nilai punya wadah yang jelas."
													/>
												</Table.Cell>
											</Table.Row>
										{/each}
									</Table.Body>
								</Table.Root>
							</Card.Content>
						</Card.Root>
					</div>

					<Card.Root>
						<Card.Header class="pb-2">
							<Card.Title class="text-base">Tambah Kelas</Card.Title>
						</Card.Header>
						<Card.Content class="space-y-3">
							<div>
								<label for="class-year-id" class="text-xs text-slate-500 mb-1 block">Tahun Ajaran</label>
								<select id="class-year-id" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={classYearId}>
									<option value="">-- Pilih --</option>
										{#each years as y (y.id)}
										<option value={y.id}>{y.name}</option>
									{/each}
								</select>
							</div>
							<Input id="class-code" placeholder="Kode, mis: 7A" bind:value={classCode} />
							<Input id="class-name" placeholder="Nama kelas, mis: VII A" bind:value={className} />
							<Input id="class-level" placeholder="Tingkat, mis: VII" bind:value={classLevel} />
							<label class="flex items-center gap-2 text-sm">
								<input type="checkbox" bind:checked={classActive} class="rounded" />
								Kelas aktif
							</label>
							<LoadingButton class="w-full" loading={classBusy} loadingLabel="Menyimpan..." disabled={!className || !classCode || !classLevel || !classYearId} onclick={createClass} label="Simpan" />
						</Card.Content>
					</Card.Root>
				</div>
			</Tabs.Content>

			<!-- Mata Pelajaran Tab -->
			<Tabs.Content value="subjects">
				<div class="grid gap-4 lg:grid-cols-3">
					<div class="lg:col-span-2">
						<Card.Root>
							<Card.Header class="pb-2">
								<Card.Title class="text-base">Daftar Mata Pelajaran</Card.Title>
							</Card.Header>
							<Card.Content class="p-0 overflow-x-auto">
								<Table.Root>
									<Table.Header>
										<Table.Row>
											<Table.Head>Kode</Table.Head>
											<Table.Head>Nama</Table.Head>
											<Table.Head>Status</Table.Head>
											<Table.Head></Table.Head>
										</Table.Row>
									</Table.Header>
									<Table.Body>
											{#each subjects as s (s.id)}
											<Table.Row>
												<Table.Cell class="font-mono text-sm">{s.code}</Table.Cell>
												<Table.Cell class="font-medium">{s.name}</Table.Cell>
												<Table.Cell>
													{#if s.is_active}
														<Badge class="bg-emerald-100 text-emerald-700 border-emerald-200">Aktif</Badge>
													{:else}
														<Badge variant="secondary">Nonaktif</Badge>
													{/if}
												</Table.Cell>
												<Table.Cell>
													<Button variant="destructive" size="xs" onclick={() => deleteSubject(s.id)}>Hapus</Button>
												</Table.Cell>
											</Table.Row>
										{:else}
											<Table.Row>
												<Table.Cell colspan={4} class="p-4">
													<EmptyStatePanel
														compact
														eyebrow="Kurikulum"
														title="Belum ada mata pelajaran"
														description="Buat daftar mapel inti lebih dulu supaya assignment guru, gradebook, dan bank soal bisa memakai referensi yang sama."
													/>
												</Table.Cell>
											</Table.Row>
										{/each}
									</Table.Body>
								</Table.Root>
							</Card.Content>
						</Card.Root>
					</div>

					<Card.Root>
						<Card.Header class="pb-2">
							<Card.Title class="text-base">Tambah Mata Pelajaran</Card.Title>
						</Card.Header>
						<Card.Content class="space-y-3">
							<Input placeholder="Kode, mis: MTK" bind:value={subjectCode} />
							<Input placeholder="Nama, mis: Matematika" bind:value={subjectName} />
							<label class="flex items-center gap-2 text-sm">
								<input type="checkbox" bind:checked={subjectActive} class="rounded" />
								Aktif
							</label>
							<LoadingButton class="w-full" loading={subjectBusy} loadingLabel="Menyimpan..." disabled={!subjectName || !subjectCode} onclick={createSubject} label="Simpan" />
						</Card.Content>
					</Card.Root>
				</div>
			</Tabs.Content>
		</Tabs.Root>
	{/if}
</div>
