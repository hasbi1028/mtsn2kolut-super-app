<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import { Pencil, RefreshCw } from '@lucide/svelte';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { toast } from '$lib/components/ui/sonner';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import DirtyChangeBar from '$lib/components/editable/DirtyChangeBar.svelte';
	import EditableSelectCell from '$lib/components/editable/EditableSelectCell.svelte';
	import EditableTextCell from '$lib/components/editable/EditableTextCell.svelte';
	import EmptyStatePanel from '$lib/components/EmptyStatePanel.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { bindBeforeUnload, confirmDiscardChanges } from '$lib/client/unsaved-changes';
	import { readClientApiData, readClientJson } from '$lib/client/api';
	import type { EditableOption } from '$lib/components/editable/types';

	type RombelSummary = {
		id: string;
		code: string;
		name: string;
		level: string;
		is_active: boolean;
		academic_year_name: string;
		homeroom_teacher_name: string;
		total_students: number;
		total_subject_teachers: number;
		total_subject_assignments: number;
		total_timetable_slots: number;
	};

	type RombelDraft = {
		code: string;
		name: string;
		level: string;
		is_active: string;
		error: string;
	};

	const levelOptions: EditableOption[] = [
		{ value: 'VII', label: 'VII' },
		{ value: 'VIII', label: 'VIII' },
		{ value: 'IX', label: 'IX' },
	];
	const statusOptions: EditableOption[] = [
		{ value: 'true', label: 'Aktif' },
		{ value: 'false', label: 'Nonaktif' },
	];

	let rombels = $state<RombelSummary[]>([]);
	let rombelDrafts = $state<Record<string, RombelDraft>>({});
	let rombelPromise = $state<Promise<RombelSummary[]> | null>(null);
	let refreshBusy = $state(false);
	let quickEdit = $state(false);
	let saveQuickBusy = $state(false);
	let quickSavingRowId = $state('');
	let requestId = 0;

	const summary = $derived.by(() => ({
		activeClasses: rombels.filter((item) => item.is_active).length,
		totalStudents: rombels.reduce((sum, item) => sum + item.total_students, 0),
		missingHomeroom: rombels.filter((item) => !item.homeroom_teacher_name).length,
	}));
	const dirtyRows = $derived(rombels.filter((rombel) => isRombelDirty(rombel)));
	const dirtyCount = $derived(dirtyRows.length);

	async function fetchRombels(): Promise<RombelSummary[]> {
		return await fetch('/api/academic/rombel').then((response) =>
			readClientApiData<RombelSummary[]>(response, 'Gagal memuat data rombel')
		);
	}

	function toDraft(rombel: RombelSummary): RombelDraft {
		return {
			code: rombel.code,
			name: rombel.name,
			level: rombel.level,
			is_active: String(rombel.is_active),
			error: '',
		};
	}

	function resetRombelDrafts(source = rombels) {
		rombelDrafts = Object.fromEntries(source.map((rombel) => [rombel.id, toDraft(rombel)]));
	}

	function applyRombels(items: RombelSummary[]) {
		rombels = items;
		if (quickEdit) resetRombelDrafts(items);
		return items;
	}

	function loadRombels() {
		const current = ++requestId;
		rombelPromise = fetchRombels().then((items) => {
			if (current === requestId) {
				applyRombels(items);
			}
			return items;
		});
	}

	async function refreshRombels() {
		if (dirtyCount > 0 && !(await confirmDiscardChanges(true))) return;
		refreshBusy = true;
		try {
			loadRombels();
			await rombelPromise;
		} finally {
			refreshBusy = false;
		}
	}

	async function toggleQuickEdit() {
		if (quickEdit && dirtyCount > 0 && !(await confirmDiscardChanges(true))) return;
		quickEdit = !quickEdit;
		resetRombelDrafts();
	}

	function getRombelDraft(rombel: RombelSummary) {
		return rombelDrafts[rombel.id] ?? toDraft(rombel);
	}

	function updateDraft(id: string, field: keyof Omit<RombelDraft, 'error'>, value: string) {
		const current = rombelDrafts[id];
		if (!current) return;
		rombelDrafts = {
			...rombelDrafts,
			[id]: { ...current, [field]: value, error: '' },
		};
	}

	function markDraftError(id: string, error: string) {
		const current = rombelDrafts[id];
		if (!current) return;
		rombelDrafts = { ...rombelDrafts, [id]: { ...current, error } };
	}

	function isRombelDirty(rombel: RombelSummary) {
		const draft = getRombelDraft(rombel);
		return (
			draft.code.trim() !== rombel.code ||
			draft.name.trim() !== rombel.name ||
			draft.level !== rombel.level ||
			(draft.is_active === 'true') !== rombel.is_active
		);
	}

	function validateDraft(draft: RombelDraft) {
		if (!draft.code.trim()) return 'Kode rombel wajib diisi.';
		if (!draft.name.trim()) return 'Nama rombel wajib diisi.';
		if (!levelOptions.some((option) => option.value === draft.level)) return 'Tingkat rombel harus VII, VIII, atau IX.';
		return '';
	}

	async function saveChangedRombels() {
		const rows = [...dirtyRows];
		const invalidRow = rows.find((rombel) => validateDraft(getRombelDraft(rombel)));
		if (invalidRow) {
			const message = validateDraft(getRombelDraft(invalidRow));
			markDraftError(invalidRow.id, message);
			toast.error(`${invalidRow.code}: ${message}`);
			return;
		}
		saveQuickBusy = true;
		try {
			for (const rombel of rows) {
				const draft = getRombelDraft(rombel);
				quickSavingRowId = rombel.id;
				const response = await fetch(`/api/academic/rombel/${rombel.id}`, {
					method: 'PUT',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify({
						code: draft.code.trim(),
						name: draft.name.trim(),
						level: draft.level,
						is_active: draft.is_active === 'true',
					}),
				});
				try {
					const updated = await readClientJson<RombelSummary>(response);
					const merged = { ...rombel, ...updated };
					rombels = rombels.map((item) => (item.id === rombel.id ? merged : item));
					rombelDrafts = { ...rombelDrafts, [rombel.id]: toDraft(merged) };
				} catch (error) {
					const message = error instanceof Error && error.message.trim() ? error.message : 'Gagal menyimpan rombel.';
					markDraftError(rombel.id, message);
					toast.error(`${rombel.code}: ${message}`);
					return;
				}
			}
			toast.success(`${rows.length} perubahan rombel tersimpan`);
		} finally {
			saveQuickBusy = false;
			quickSavingRowId = '';
		}
	}

	function rombelErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Data rombel belum dapat dimuat. Periksa layanan sistem lalu coba lagi.';
	}

	function handleRenderError(error: unknown) {
		console.error('Rombel list render failed', error);
	}

	onMount(() => {
		loadRombels();
		return bindBeforeUnload(() => dirtyCount > 0);
	});
</script>

<svelte:head>
	<title>Rombel | MTsN 2 Kolaka Utara</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex flex-col gap-3 md:flex-row md:items-end md:justify-between">
		<div>
			<p class="text-sm font-medium text-primary">Akademik</p>
			<h1 class="text-2xl font-semibold tracking-normal text-foreground">Rombel</h1>
			<p class="mt-1 max-w-2xl text-sm text-muted-foreground">
				Daftar rombel beserta wali kelas, jumlah siswa, penugasan guru mapel, dan jam pelajaran.
			</p>
		</div>
		<div class="flex flex-wrap gap-2">
			<Button variant={quickEdit ? 'default' : 'outline'} onclick={() => void toggleQuickEdit()} disabled={saveQuickBusy}>
				<Pencil class="mr-2 size-4" />
				{quickEdit ? 'Selesai Edit Cepat' : 'Edit Cepat'}
			</Button>
			<Button variant="outline" onclick={() => void refreshRombels()} disabled={refreshBusy || saveQuickBusy} aria-label="Muat ulang rombel">
				<RefreshCw class={`mr-2 size-4 ${refreshBusy ? 'animate-spin' : ''}`} />
				Muat Ulang
			</Button>
		</div>
	</div>

	<AsyncContent promise={rombelPromise} onerror={handleRenderError}>
		{#snippet pending()}
			<div class="space-y-4">
				<div class="grid gap-3 md:grid-cols-3">
					<Skeleton class="h-24 w-full" />
					<Skeleton class="h-24 w-full" />
					<Skeleton class="h-24 w-full" />
				</div>
				<Skeleton class="h-80 w-full" />
			</div>
		{/snippet}

		{#snippet failed(error, reset)}
			<RecoveryPanel
				title="Rombel Belum Tersaji"
				message={rombelErrorMessage(error)}
				onRetry={() => {
					reset?.();
					loadRombels();
				}}
			/>
		{/snippet}

		<div class="grid gap-3 md:grid-cols-3">
			<Card.Root>
				<Card.Header class="pb-2">
					<Card.Description>Rombel Aktif</Card.Description>
					<Card.Title class="text-2xl">{summary.activeClasses}</Card.Title>
				</Card.Header>
			</Card.Root>
			<Card.Root>
				<Card.Header class="pb-2">
					<Card.Description>Total Siswa Aktif</Card.Description>
					<Card.Title class="text-2xl">{summary.totalStudents}</Card.Title>
				</Card.Header>
			</Card.Root>
			<Card.Root>
				<Card.Header class="pb-2">
					<Card.Description>Belum Ada Wali Kelas</Card.Description>
					<Card.Title class="text-2xl">{summary.missingHomeroom}</Card.Title>
				</Card.Header>
			</Card.Root>
		</div>

		<Card.Root>
			<Card.Header>
				<Card.Title class="text-base">Daftar Rombel</Card.Title>
				<Card.Description>
					{quickEdit
						? 'Edit kode, nama, tingkat, dan status aktif. Perubahan disimpan berurutan agar masalah per rombel tetap jelas.'
						: 'Gunakan detail rombel untuk melihat siswa, orang tua, wali kelas, guru mapel, jadwal, dan menetapkan wali kelas yang masih kosong.'}
				</Card.Description>
			</Card.Header>
			<Card.Content class="p-0">
				<div class="overflow-x-auto">
					<Table.Root>
						<Table.Header>
							<Table.Row>
								{#if quickEdit}
									<Table.Head>Kode</Table.Head>
									<Table.Head>Nama Rombel</Table.Head>
									<Table.Head>Tingkat</Table.Head>
								{:else}
									<Table.Head>Kelas</Table.Head>
								{/if}
								<Table.Head>Tahun Ajaran</Table.Head>
								<Table.Head>Wali Kelas</Table.Head>
								<Table.Head class="text-right">Siswa</Table.Head>
								<Table.Head class="text-right">Penugasan Guru Mapel</Table.Head>
								<Table.Head class="text-right">Jadwal</Table.Head>
								<Table.Head>Status</Table.Head>
								<Table.Head></Table.Head>
							</Table.Row>
						</Table.Header>
						<Table.Body>
							{#each rombels as rombel (rombel.id)}
								{@const draft = getRombelDraft(rombel)}
								{@const rowDirty = isRombelDirty(rombel)}
								<Table.Row class={rowDirty ? 'bg-primary/5' : ''}>
									{#if quickEdit}
										<Table.Cell>
											<EditableTextCell
												value={draft.code}
												ariaLabel={`Edit kode ${rombel.name}`}
												dirty={draft.code.trim() !== rombel.code}
												invalid={!draft.code.trim()}
												invalidMessage="Kode wajib diisi"
												disabled={saveQuickBusy}
												oninput={(value) => updateDraft(rombel.id, 'code', value)}
												oncancel={(event) => updateDraft(rombel.id, 'code', event.originalValue)}
											/>
										</Table.Cell>
										<Table.Cell>
											<EditableTextCell
												value={draft.name}
												ariaLabel={`Edit nama ${rombel.name}`}
												dirty={draft.name.trim() !== rombel.name}
												invalid={!draft.name.trim()}
												invalidMessage="Nama wajib diisi"
												disabled={saveQuickBusy}
												oninput={(value) => updateDraft(rombel.id, 'name', value)}
												oncancel={(event) => updateDraft(rombel.id, 'name', event.originalValue)}
											/>
											{#if draft.error}
												<p class="mt-1 max-w-xs text-xs text-destructive">{draft.error}</p>
											{/if}
										</Table.Cell>
										<Table.Cell>
											<EditableSelectCell
												value={draft.level}
												options={levelOptions}
												allowEmpty={false}
												ariaLabel={`Edit tingkat ${rombel.name}`}
												dirty={draft.level !== rombel.level}
												disabled={saveQuickBusy}
												oninput={(value) => updateDraft(rombel.id, 'level', value)}
												oncancel={(event) => updateDraft(rombel.id, 'level', event.originalValue)}
											/>
										</Table.Cell>
									{:else}
										<Table.Cell>
											<div class="space-y-0.5">
												<p class="font-medium text-foreground">{rombel.name}</p>
												<p class="text-xs text-muted-foreground">{rombel.code} · Tingkat {rombel.level}</p>
											</div>
										</Table.Cell>
									{/if}
									<Table.Cell class="text-muted-foreground">{rombel.academic_year_name}</Table.Cell>
									<Table.Cell>
										{#if rombel.homeroom_teacher_name}
											{rombel.homeroom_teacher_name}
										{:else}
											<div class="space-y-0.5">
												<span class="block text-muted-foreground">Belum ditetapkan</span>
												<span class="block text-xs text-primary">Atur dari halaman detail</span>
											</div>
										{/if}
									</Table.Cell>
									<Table.Cell class="text-right tabular-nums">{rombel.total_students}</Table.Cell>
									<Table.Cell class="text-right tabular-nums">{rombel.total_subject_teachers}</Table.Cell>
									<Table.Cell class="text-right tabular-nums">{rombel.total_timetable_slots}</Table.Cell>
									<Table.Cell>
										{#if quickEdit}
											<div class="min-w-36">
												<EditableSelectCell
													value={draft.is_active}
													options={statusOptions}
													allowEmpty={false}
													ariaLabel={`Edit status ${rombel.name}`}
													dirty={(draft.is_active === 'true') !== rombel.is_active}
													loading={quickSavingRowId === rombel.id}
													disabled={saveQuickBusy && quickSavingRowId !== rombel.id}
													oninput={(value) => updateDraft(rombel.id, 'is_active', value)}
													oncancel={(event) => updateDraft(rombel.id, 'is_active', event.originalValue)}
												/>
												{#if draft.is_active === 'false' && rombel.total_students > 0}
													<p class="mt-1 max-w-40 text-xs text-muted-foreground">Layanan sistem akan menolak jika masih ada siswa aktif.</p>
												{/if}
											</div>
										{:else}
											{#if rombel.is_active}
												<Badge class="border-primary/20 bg-primary/15 text-primary">Aktif</Badge>
											{:else}
												<Badge variant="secondary">Nonaktif</Badge>
											{/if}
										{/if}
									</Table.Cell>
									<Table.Cell class="text-right">
										<Button variant="outline" size="sm" href={resolve(`/akademik/rombel/${rombel.id}` as '/')}>
											Detail
										</Button>
									</Table.Cell>
								</Table.Row>
							{:else}
								<Table.Row>
									<Table.Cell colspan={quickEdit ? 10 : 8} class="p-4">
										<EmptyStatePanel
											compact
											title="Belum ada rombel"
											description="Rombel mengikuti data kelas akademik yang sudah dibuat di Data Akademik."
										/>
									</Table.Cell>
								</Table.Row>
							{/each}
						</Table.Body>
					</Table.Root>
				</div>
			</Card.Content>
		</Card.Root>
		<DirtyChangeBar
			count={dirtyCount}
			saving={saveQuickBusy}
			saveLabel={saveQuickBusy ? 'Menyimpan...' : 'Simpan perubahan'}
			discardLabel="Batalkan"
			onsave={() => void saveChangedRombels()}
			ondiscard={() => resetRombelDrafts()}
		/>
	</AsyncContent>
</div>
