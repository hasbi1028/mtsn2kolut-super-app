<script lang="ts">
	import { onMount } from 'svelte';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Badge } from '$lib/components/ui/badge';
	import { toast } from '$lib/components/ui/sonner';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import EmptyStatePanel from '$lib/components/EmptyStatePanel.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';

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

	let loading = $state(true);
	let error = $state('');
	let assignments = $state<Assignment[]>([]);
	let components = $state<GradeComponent[]>([]);
	let summary = $state<GradeSummary[]>([]);
	let entries = $state<GradeEntry[]>([]);

	let assignmentId = $state('');
	let componentId = $state('');

	let createBusy = $state(false);
	let entryBusy = $state<Record<string, boolean>>({});
	let publishBusy = $state<Record<string, boolean>>({});
	let bulkSaveBusy = $state(false);
	let editingComponentId = $state('');
	let componentTitle = $state('');
	let componentCategory = $state('assignment');
	let componentWeight = $state(1);
	let componentMaxScore = $state(100);

	let scoreInput = $state<Record<string, string>>({});
	let noteInput = $state<Record<string, string>>({});

	const selectedAssignment = $derived(assignments.find((item) => item.id === assignmentId) ?? null);
	const selectedComponent = $derived(components.find((item) => item.id === componentId) ?? null);
	const editingComponent = $derived(components.find((item) => item.id === editingComponentId) ?? null);
	const completionRate = $derived(summary.length === 0 ? 0 : Math.round((summary.filter((row) => row.filled_count > 0).length / summary.length) * 100));
	const publishedComponentCount = $derived(components.filter((item) => item.is_published).length);
	const draftComponentCount = $derived(components.filter((item) => !item.is_published).length);
	const readyStudentCount = $derived(summary.filter((row) => row.component_count > 0 && row.filled_count === row.component_count).length);
	const incompleteStudentCount = $derived(summary.filter((row) => row.component_count === 0 || row.filled_count < row.component_count).length);
	const missingGradeCount = $derived(summary.reduce((total, row) => total + Math.max(row.component_count - row.filled_count, 0), 0));
	const dirtyEntryIds = $derived(
		entries
			.filter((row) => isEntryDirty(row.student_id))
			.map((row) => row.student_id)
	);
	const readyForRapor = $derived(
		components.length > 0 &&
		draftComponentCount === 0 &&
		summary.length > 0 &&
		summary.every((row) => row.component_count > 0 && row.filled_count === row.component_count)
	);
	const readinessLabel = $derived(
		readyForRapor
			? 'Siap Rapor'
			: components.length === 0
				? 'Belum Siap'
				: draftComponentCount > 0
					? 'Masih Ada Draft'
					: 'Nilai Belum Lengkap'
	);
	const readinessDescription = $derived(
		readyForRapor
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

	async function quickSelectFirstAssignment() {
		if (assignments.length === 0) return;
		assignmentId = assignments[0]?.id ?? '';
		componentId = '';
		resetComponentForm();
		await loadOverview();
	}

	async function loadOverview() {
		loading = true;
		error = '';
		try {
			const params = new URLSearchParams();
			if (assignmentId) params.set('assignment_id', assignmentId);
			if (componentId) params.set('component_id', componentId);
			const url = params.size > 0 ? `/api/grades?${params.toString()}` : '/api/grades';
			const res = await fetch(url);
			const json = await res.json();
			if (!res.ok || json.error) {
				error = json.error ?? 'Gagal memuat data nilai';
				return;
			}
			const data = json.data ?? json;
			assignments = data.assignments ?? [];
			components = data.components ?? [];
			summary = data.summary ?? [];
			entries = data.entries ?? [];

			const nextScores: Record<string, string> = {};
			const nextNotes: Record<string, string> = {};
			for (const row of entries) {
				nextScores[row.student_id] = row.score >= 0 ? String(row.score) : '';
				nextNotes[row.student_id] = row.notes ?? '';
			}
			scoreInput = nextScores;
			noteInput = nextNotes;
		} catch {
			error = 'Gagal memuat data nilai';
		} finally {
			loading = false;
		}
	}

	function beginEditComponent(component: GradeComponent) {
		editingComponentId = component.id;
		componentTitle = component.title;
		componentCategory = component.category;
		componentWeight = component.weight;
		componentMaxScore = component.max_score;
	}

	async function saveComponent() {
		if ((!assignmentId && !editingComponentId) || !componentTitle) return;
		createBusy = true;
		try {
			const path = editingComponentId ? `/api/grades/components/${editingComponentId}` : '/api/grades/components';
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
			const json = await res.json().catch(() => ({}));
			if (!res.ok) {
				showError(json.error ?? (editingComponentId ? 'Gagal memperbarui komponen nilai' : 'Gagal menambah komponen nilai'));
				return;
			}
			const wasEditing = editingComponentId !== '';
			resetComponentForm();
			showSuccess(wasEditing ? 'Komponen nilai diperbarui' : 'Komponen nilai ditambahkan');
			await loadOverview();
		} finally {
			createBusy = false;
		}
	}

	async function deleteComponent(id: string) {
		if (!confirm('Hapus komponen nilai ini?')) return;
		const res = await fetch(`/api/grades/components/${id}`, { method: 'DELETE' });
		if (!res.ok) {
			const json = await res.json().catch(() => ({}));
			showError(json.error ?? 'Gagal menghapus komponen');
			return;
		}
		if (editingComponentId === id) resetComponentForm();
		if (componentId === id) componentId = '';
		showSuccess('Komponen nilai dihapus');
		await loadOverview();
	}

	async function togglePublish(component: GradeComponent) {
		publishBusy = { ...publishBusy, [component.id]: true };
		try {
			const res = await fetch(`/api/grades/components/${component.id}/publish`, {
				method: 'PATCH',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					is_published: !component.is_published
				})
			});
			const json = await res.json().catch(() => ({}));
			if (!res.ok) {
				showError(json.error ?? 'Gagal memperbarui status komponen');
				return;
			}
			showSuccess(component.is_published ? 'Komponen dikembalikan ke draft' : 'Komponen diterbitkan untuk rapor');
			await loadOverview();
		} finally {
			publishBusy = { ...publishBusy, [component.id]: false };
		}
	}

	async function persistEntry(studentId: string, silent = false) {
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
			const res = await fetch(`/api/grades/components/${componentId}/entries`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					student_id: studentId,
					score,
					notes: noteInput[studentId] ?? ''
				})
			});
			const json = await res.json().catch(() => ({}));
			if (!res.ok) {
				throw new Error(json.error ?? 'Gagal menyimpan nilai');
			}
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
			await loadOverview();
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
			await loadOverview();
		} finally {
			bulkSaveBusy = false;
		}
	}

	onMount(async () => {
		await loadOverview();
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
					onchange={async () => { componentId = ''; resetComponentForm(); await loadOverview(); }}
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

	{#if error}
		<RecoveryPanel message={error} onRetry={loadOverview} />
	{/if}

	{#if loading}
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
	{:else}
		{#if selectedAssignment}
			<div class="grid gap-4 md:grid-cols-3">
				<Card.Root class="border-emerald-100 bg-white">
					<Card.Content class="pt-5">
						<p class="text-xs font-semibold uppercase tracking-[0.2em] text-emerald-700">Kelas & Mapel</p>
						<p class="mt-2 text-lg font-semibold text-slate-900">{selectedAssignment.class_name}</p>
						<p class="text-sm text-slate-600">{selectedAssignment.subject_name} · {selectedAssignment.subject_code}</p>
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
						{#if readyForRapor}
							<a
								href={`/grades/rapor?assignment_id=${assignmentId}`}
								class="inline-flex w-full items-center justify-center rounded-md bg-emerald-700 px-4 py-2 text-sm font-medium text-white hover:bg-emerald-800"
							>
								Buka Cetak Rapor
							</a>
						{:else}
							<div class="rounded-xl border border-dashed border-amber-300 bg-white/70 px-4 py-3 text-sm text-amber-900">
								Selesaikan draft dan lengkapi semua nilai sebelum membuka rapor final.
							</div>
						{/if}
					</div>
				</Card.Content>
			</Card.Root>

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
									onclick={saveComponent}
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
											</Table.Cell>
											<Table.Cell>
												<div class="flex flex-wrap gap-2">
													<Badge variant="outline">{categoryLabel(item.category)}</Badge>
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
						{:else}
							Pilih komponen penilaian untuk mulai mengisi nilai siswa.
						{/if}
					</Card.Description>
				</Card.Header>
				<Card.Content class="space-y-4 p-0">
					{#if selectedComponent}
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
									onclick={saveAllDirtyEntries}
									label="Simpan Semua Perubahan"
								/>
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
		{:else}
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
	{/if}
</div>
