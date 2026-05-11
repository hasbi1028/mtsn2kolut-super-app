<script lang="ts">
	import { onMount } from 'svelte';
	import { Plus, RefreshCw } from '@lucide/svelte';
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
	import EditableTextCell from '$lib/components/editable/EditableTextCell.svelte';
	import EmptyStatePanel from '$lib/components/EmptyStatePanel.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { bindBeforeUnload, confirmDiscardChanges } from '$lib/client/unsaved-changes';
	import { clientApiPathWithQuery, readClientApiData, readClientJson } from '$lib/client/api';
	import type { EditableOption } from '$lib/components/editable/types';

	type Subject = {
		id: string;
		code: string;
		name: string;
		category: string;
		is_assessment_subject: boolean;
		is_report_subject: boolean;
		is_schedule_activity: boolean;
		default_weekly_hours: number;
		display_order: number;
		is_active: boolean;
	};

	type SubjectDraft = {
		code: string;
		name: string;
		category: string;
		is_assessment_subject: string;
		is_report_subject: string;
		is_schedule_activity: string;
		default_weekly_hours: string;
		display_order: string;
		is_active: string;
		error: string;
	};

	const categoryOptions: EditableOption[] = [
		{ value: 'intrakurikuler', label: 'Intrakurikuler' },
		{ value: 'muatan_lokal', label: 'Muatan Lokal' },
		{ value: 'kokurikuler', label: 'Kokurikuler' },
		{ value: 'kegiatan', label: 'Kegiatan' },
		{ value: 'lainnya', label: 'Lainnya' },
	];
	const booleanOptions: EditableOption[] = [
		{ value: 'true', label: 'Ya' },
		{ value: 'false', label: 'Tidak' },
	];
	const statusOptions: EditableOption[] = [
		{ value: 'true', label: 'Aktif' },
		{ value: 'false', label: 'Nonaktif' },
	];

	let subjects = $state<Subject[]>([]);
	let subjectDrafts = $state<Record<string, SubjectDraft>>({});
	let subjectPromise = $state<Promise<Subject[]> | null>(null);
	let refreshBusy = $state(false);
	let saveBusy = $state(false);
	let createBusy = $state(false);
	let savingSubjectId = $state('');
	let requestId = 0;

	let newCode = $state('');
	let newName = $state('');
	let newCategory = $state('intrakurikuler');
	let newWeeklyHours = $state('0');
	let newDisplayOrder = $state('0');

	const summary = $derived.by(() => ({
		active: subjects.filter((item) => item.is_active).length,
		assessment: subjects.filter((item) => item.is_assessment_subject && item.is_active).length,
		schedule: subjects.filter((item) => item.is_schedule_activity && item.is_active).length,
	}));
	const dirtySubjects = $derived(subjects.filter((subject) => isSubjectDirty(subject)));
	const dirtyCount = $derived(dirtySubjects.length);

	async function fetchSubjects(): Promise<Subject[]> {
		return await fetch('/api/academic/subjects').then((response) => readClientApiData<Subject[]>(response, 'Gagal memuat data mapel'));
	}

	function toDraft(subject: Subject): SubjectDraft {
		return {
			code: subject.code,
			name: subject.name,
			category: subject.category || 'intrakurikuler',
			is_assessment_subject: String(subject.is_assessment_subject),
			is_report_subject: String(subject.is_report_subject),
			is_schedule_activity: String(subject.is_schedule_activity),
			default_weekly_hours: String(subject.default_weekly_hours ?? 0),
			display_order: String(subject.display_order ?? 0),
			is_active: String(subject.is_active),
			error: '',
		};
	}

	function resetSubjectDrafts(source = subjects) {
		subjectDrafts = Object.fromEntries(source.map((subject) => [subject.id, toDraft(subject)]));
	}

	function applySubjects(items: Subject[]) {
		subjects = items;
		resetSubjectDrafts(items);
		return items;
	}

	function loadSubjects() {
		const current = ++requestId;
		subjectPromise = fetchSubjects().then((items) => {
			if (current === requestId) applySubjects(items);
			return items;
		});
	}

	async function refreshSubjects() {
		if (dirtyCount > 0 && !(await confirmDiscardChanges(true))) return;
		refreshBusy = true;
		try {
			loadSubjects();
			await subjectPromise;
		} finally {
			refreshBusy = false;
		}
	}

	function getSubjectDraft(subject: Subject) {
		return subjectDrafts[subject.id] ?? toDraft(subject);
	}

	function updateDraft(id: string, field: keyof Omit<SubjectDraft, 'error'>, value: string) {
		const current = subjectDrafts[id];
		if (!current) return;
		subjectDrafts = { ...subjectDrafts, [id]: { ...current, [field]: value, error: '' } };
	}

	function markDraftError(id: string, error: string) {
		const current = subjectDrafts[id];
		if (!current) return;
		subjectDrafts = { ...subjectDrafts, [id]: { ...current, error } };
	}

	function draftNumber(value: string) {
		const parsed = Number(value);
		if (!Number.isFinite(parsed)) return Number.NaN;
		return Math.trunc(parsed);
	}

	function isSubjectDirty(subject: Subject) {
		const draft = getSubjectDraft(subject);
		return (
			draft.code.trim() !== subject.code ||
			draft.name.trim() !== subject.name ||
			draft.category !== (subject.category || 'intrakurikuler') ||
			(draft.is_assessment_subject === 'true') !== subject.is_assessment_subject ||
			(draft.is_report_subject === 'true') !== subject.is_report_subject ||
			(draft.is_schedule_activity === 'true') !== subject.is_schedule_activity ||
			draftNumber(draft.default_weekly_hours) !== (subject.default_weekly_hours ?? 0) ||
			draftNumber(draft.display_order) !== (subject.display_order ?? 0) ||
			(draft.is_active === 'true') !== subject.is_active
		);
	}

	function validateDraft(draft: SubjectDraft) {
		const hours = draftNumber(draft.default_weekly_hours);
		const order = draftNumber(draft.display_order);
		if (!draft.code.trim()) return 'Kode mapel wajib diisi.';
		if (!draft.name.trim()) return 'Nama mapel wajib diisi.';
		if (!categoryOptions.some((option) => option.value === draft.category)) return 'Kategori mapel tidak valid.';
		if (!Number.isInteger(hours) || hours < 0 || hours > 60) return 'JP per pekan harus 0-60.';
		if (!Number.isInteger(order) || order < 0) return 'Urutan tampil tidak boleh negatif.';
		return '';
	}

	function subjectPayload(draft: SubjectDraft) {
		return {
			code: draft.code.trim(),
			name: draft.name.trim(),
			category: draft.category,
			is_assessment_subject: draft.is_assessment_subject === 'true',
			is_report_subject: draft.is_report_subject === 'true',
			is_schedule_activity: draft.is_schedule_activity === 'true',
			default_weekly_hours: draftNumber(draft.default_weekly_hours),
			display_order: draftNumber(draft.display_order),
			is_active: draft.is_active === 'true',
		};
	}

	async function saveChangedSubjects() {
		const rows = [...dirtySubjects];
		const invalidRow = rows.find((subject) => validateDraft(getSubjectDraft(subject)));
		if (invalidRow) {
			const message = validateDraft(getSubjectDraft(invalidRow));
			markDraftError(invalidRow.id, message);
			toast.error(`${invalidRow.code}: ${message}`);
			return;
		}
		saveBusy = true;
		try {
			for (const subject of rows) {
				const draft = getSubjectDraft(subject);
				savingSubjectId = subject.id;
				const response = await fetch(clientApiPathWithQuery('/api/academic/subjects', new URLSearchParams({ id: subject.id })), {
					method: 'PUT',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify(subjectPayload(draft)),
				});
				try {
					const updated = await readClientJson<Subject>(response);
					subjects = subjects.map((item) => (item.id === subject.id ? updated : item));
					subjectDrafts = { ...subjectDrafts, [subject.id]: toDraft(updated) };
				} catch (error) {
					const message = error instanceof Error && error.message.trim() ? error.message : 'Gagal menyimpan mapel.';
					markDraftError(subject.id, message);
					toast.error(`${subject.code}: ${message}`);
					return;
				}
			}
			toast.success(`${rows.length} perubahan mapel tersimpan`);
		} finally {
			saveBusy = false;
			savingSubjectId = '';
		}
	}

	async function createSubject() {
		const draft: SubjectDraft = {
			code: newCode,
			name: newName,
			category: newCategory,
			is_assessment_subject: 'true',
			is_report_subject: 'true',
			is_schedule_activity: 'false',
			default_weekly_hours: newWeeklyHours,
			display_order: newDisplayOrder,
			is_active: 'true',
			error: '',
		};
		const error = validateDraft(draft);
		if (error) {
			toast.error(error);
			return;
		}
		createBusy = true;
		try {
			const response = await fetch('/api/academic/subjects', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(subjectPayload(draft)),
			});
			const created = await readClientJson<Subject>(response);
			const next = [...subjects, created].sort((a, b) => a.display_order - b.display_order || a.name.localeCompare(b.name, 'id-ID'));
			applySubjects(next);
			newCode = '';
			newName = '';
			newCategory = 'intrakurikuler';
			newWeeklyHours = '0';
			newDisplayOrder = '0';
			toast.success('Mapel baru tersimpan');
		} catch (error) {
			toast.error(error instanceof Error && error.message.trim() ? error.message : 'Gagal menambah mapel.');
		} finally {
			createBusy = false;
		}
	}

	function subjectErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Data mapel belum dapat dimuat. Periksa koneksi backend lalu coba lagi.';
	}

	function handleRenderError(error: unknown) {
		console.error('Subject list render failed', error);
	}

	onMount(() => {
		loadSubjects();
		return bindBeforeUnload(() => dirtyCount > 0);
	});
</script>

<svelte:head>
	<title>Mapel | MTsN 2 Kolaka Utara</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex flex-col gap-3 md:flex-row md:items-end md:justify-between">
		<div>
			<p class="text-sm font-medium text-primary">Akademik</p>
			<h1 class="text-2xl font-semibold tracking-normal text-foreground">Mapel</h1>
			<p class="mt-1 max-w-2xl text-sm text-muted-foreground">Kelola kode, nama, kategori, penilaian, rapor, kegiatan jadwal, JP default, dan urutan tampil.</p>
		</div>
		<Button variant="outline" onclick={() => void refreshSubjects()} disabled={refreshBusy || saveBusy} aria-label="Muat ulang mapel">
			<RefreshCw class={`mr-2 size-4 ${refreshBusy ? 'animate-spin' : ''}`} />
			Muat Ulang
		</Button>
	</div>

	<AsyncContent promise={subjectPromise} onerror={handleRenderError}>
		{#snippet pending()}
			<div class="space-y-4">
				<div class="grid gap-3 md:grid-cols-3">
					<Skeleton class="h-24 w-full" />
					<Skeleton class="h-24 w-full" />
					<Skeleton class="h-24 w-full" />
				</div>
				<Skeleton class="h-96 w-full" />
			</div>
		{/snippet}

		{#snippet failed(error, reset)}
			<RecoveryPanel
				title="Mapel Belum Tersaji"
				message={subjectErrorMessage(error)}
				onRetry={() => {
					reset?.();
					loadSubjects();
				}}
			/>
		{/snippet}

		<div class="grid gap-3 md:grid-cols-3">
			<Card.Root>
				<Card.Header class="pb-2">
					<Card.Description>Mapel Aktif</Card.Description>
					<Card.Title class="text-2xl">{summary.active}</Card.Title>
				</Card.Header>
			</Card.Root>
			<Card.Root>
				<Card.Header class="pb-2">
					<Card.Description>Dipakai Asesmen</Card.Description>
					<Card.Title class="text-2xl">{summary.assessment}</Card.Title>
				</Card.Header>
			</Card.Root>
			<Card.Root>
				<Card.Header class="pb-2">
					<Card.Description>Kegiatan Jadwal</Card.Description>
					<Card.Title class="text-2xl">{summary.schedule}</Card.Title>
				</Card.Header>
			</Card.Root>
		</div>

		<Card.Root>
			<Card.Header>
				<Card.Title class="text-base">Tambah Mapel</Card.Title>
				<Card.Description>Gunakan kode singkat yang konsisten dengan Bank Soal, rapor, dan jadwal.</Card.Description>
			</Card.Header>
			<Card.Content>
				<div class="grid gap-3 md:grid-cols-6">
					<div class="space-y-1">
						<label for="subject-code" class="text-sm font-medium">Kode</label>
						<Input id="subject-code" bind:value={newCode} placeholder="MTK" disabled={createBusy} />
					</div>
					<div class="space-y-1 md:col-span-2">
						<label for="subject-name" class="text-sm font-medium">Nama Mapel</label>
						<Input id="subject-name" bind:value={newName} placeholder="Matematika" disabled={createBusy} />
					</div>
					<div class="space-y-1">
						<label for="subject-category" class="text-sm font-medium">Kategori</label>
						<select
							id="subject-category"
							class="h-10 w-full rounded-md border border-input bg-background px-3 text-sm text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
							bind:value={newCategory}
							disabled={createBusy}
						>
							{#each categoryOptions as option (option.value)}
								<option value={option.value}>{option.label}</option>
							{/each}
						</select>
					</div>
					<div class="space-y-1">
						<label for="subject-hours" class="text-sm font-medium">JP/Pekan</label>
						<Input id="subject-hours" type="number" min="0" max="60" bind:value={newWeeklyHours} disabled={createBusy} />
					</div>
					<div class="space-y-1">
						<label for="subject-order" class="text-sm font-medium">Urutan</label>
						<Input id="subject-order" type="number" min="0" bind:value={newDisplayOrder} disabled={createBusy} />
					</div>
				</div>
				<div class="mt-3 flex justify-end">
					<LoadingButton loading={createBusy} onclick={() => void createSubject()}>
						<Plus class="mr-2 size-4" />
						Tambah Mapel
					</LoadingButton>
				</div>
			</Card.Content>
		</Card.Root>

		<Card.Root>
			<Card.Header>
				<Card.Title class="text-base">Daftar Mapel</Card.Title>
				<Card.Description>Perubahan disimpan berurutan agar error per mapel tetap mudah dilacak.</Card.Description>
			</Card.Header>
			<Card.Content class="p-0">
				<div class="overflow-x-auto">
					<Table.Root>
						<Table.Header>
							<Table.Row>
								<Table.Head>Kode</Table.Head>
								<Table.Head>Nama</Table.Head>
								<Table.Head>Kategori</Table.Head>
								<Table.Head>Asesmen</Table.Head>
								<Table.Head>Rapor</Table.Head>
								<Table.Head>Kegiatan</Table.Head>
								<Table.Head>JP/Pekan</Table.Head>
								<Table.Head>Urutan</Table.Head>
								<Table.Head>Status</Table.Head>
							</Table.Row>
						</Table.Header>
						<Table.Body>
							{#each subjects as subject (subject.id)}
								{@const draft = getSubjectDraft(subject)}
								{@const rowDirty = isSubjectDirty(subject)}
								<Table.Row class={rowDirty ? 'bg-primary/5' : ''}>
									<Table.Cell>
										<EditableTextCell
											value={draft.code}
											ariaLabel={`Edit kode ${subject.name}`}
											dirty={draft.code.trim() !== subject.code}
											invalid={!draft.code.trim()}
											invalidMessage="Kode wajib diisi"
											disabled={saveBusy}
											oninput={(value) => updateDraft(subject.id, 'code', value)}
											oncancel={(event) => updateDraft(subject.id, 'code', event.originalValue)}
										/>
									</Table.Cell>
									<Table.Cell>
										<EditableTextCell
											value={draft.name}
											ariaLabel={`Edit nama ${subject.name}`}
											dirty={draft.name.trim() !== subject.name}
											invalid={!draft.name.trim()}
											invalidMessage="Nama wajib diisi"
											disabled={saveBusy}
											oninput={(value) => updateDraft(subject.id, 'name', value)}
											oncancel={(event) => updateDraft(subject.id, 'name', event.originalValue)}
										/>
										{#if draft.error}
											<p class="mt-1 max-w-xs text-xs text-destructive">{draft.error}</p>
										{/if}
									</Table.Cell>
									<Table.Cell>
										<EditableSelectCell
											value={draft.category}
											options={categoryOptions}
											allowEmpty={false}
											ariaLabel={`Edit kategori ${subject.name}`}
											dirty={draft.category !== (subject.category || 'intrakurikuler')}
											disabled={saveBusy}
											oninput={(value) => updateDraft(subject.id, 'category', value)}
											oncancel={(event) => updateDraft(subject.id, 'category', event.originalValue)}
										/>
									</Table.Cell>
									<Table.Cell>
										<EditableSelectCell
											value={draft.is_assessment_subject}
											options={booleanOptions}
											allowEmpty={false}
											ariaLabel={`Edit flag asesmen ${subject.name}`}
											dirty={(draft.is_assessment_subject === 'true') !== subject.is_assessment_subject}
											disabled={saveBusy}
											oninput={(value) => updateDraft(subject.id, 'is_assessment_subject', value)}
											oncancel={(event) => updateDraft(subject.id, 'is_assessment_subject', event.originalValue)}
										/>
									</Table.Cell>
									<Table.Cell>
										<EditableSelectCell
											value={draft.is_report_subject}
											options={booleanOptions}
											allowEmpty={false}
											ariaLabel={`Edit flag rapor ${subject.name}`}
											dirty={(draft.is_report_subject === 'true') !== subject.is_report_subject}
											disabled={saveBusy}
											oninput={(value) => updateDraft(subject.id, 'is_report_subject', value)}
											oncancel={(event) => updateDraft(subject.id, 'is_report_subject', event.originalValue)}
										/>
									</Table.Cell>
									<Table.Cell>
										<EditableSelectCell
											value={draft.is_schedule_activity}
											options={booleanOptions}
											allowEmpty={false}
											ariaLabel={`Edit flag kegiatan ${subject.name}`}
											dirty={(draft.is_schedule_activity === 'true') !== subject.is_schedule_activity}
											disabled={saveBusy}
											oninput={(value) => updateDraft(subject.id, 'is_schedule_activity', value)}
											oncancel={(event) => updateDraft(subject.id, 'is_schedule_activity', event.originalValue)}
										/>
									</Table.Cell>
									<Table.Cell>
										<EditableTextCell
											value={draft.default_weekly_hours}
											inputType="number"
											ariaLabel={`Edit JP per pekan ${subject.name}`}
											dirty={draftNumber(draft.default_weekly_hours) !== (subject.default_weekly_hours ?? 0)}
											invalid={!Number.isInteger(draftNumber(draft.default_weekly_hours)) || draftNumber(draft.default_weekly_hours) < 0 || draftNumber(draft.default_weekly_hours) > 60}
											invalidMessage="0-60"
											disabled={saveBusy}
											oninput={(value) => updateDraft(subject.id, 'default_weekly_hours', value)}
											oncancel={(event) => updateDraft(subject.id, 'default_weekly_hours', event.originalValue)}
										/>
									</Table.Cell>
									<Table.Cell>
										<EditableTextCell
											value={draft.display_order}
											inputType="number"
											ariaLabel={`Edit urutan ${subject.name}`}
											dirty={draftNumber(draft.display_order) !== (subject.display_order ?? 0)}
											invalid={!Number.isInteger(draftNumber(draft.display_order)) || draftNumber(draft.display_order) < 0}
											invalidMessage="Minimal 0"
											disabled={saveBusy}
											oninput={(value) => updateDraft(subject.id, 'display_order', value)}
											oncancel={(event) => updateDraft(subject.id, 'display_order', event.originalValue)}
										/>
									</Table.Cell>
									<Table.Cell>
										<div class="min-w-36">
											<EditableSelectCell
												value={draft.is_active}
												options={statusOptions}
												allowEmpty={false}
												ariaLabel={`Edit status ${subject.name}`}
												dirty={(draft.is_active === 'true') !== subject.is_active}
												loading={savingSubjectId === subject.id}
												disabled={saveBusy && savingSubjectId !== subject.id}
												oninput={(value) => updateDraft(subject.id, 'is_active', value)}
												oncancel={(event) => updateDraft(subject.id, 'is_active', event.originalValue)}
											/>
											{#if subject.is_active}
												<Badge class="mt-1 border-primary/20 bg-primary/15 text-primary">Aktif</Badge>
											{:else}
												<Badge class="mt-1" variant="secondary">Nonaktif</Badge>
											{/if}
										</div>
									</Table.Cell>
								</Table.Row>
							{:else}
								<Table.Row>
									<Table.Cell colspan={9} class="p-4">
										<EmptyStatePanel compact title="Belum ada mapel" description="Tambahkan mapel pertama untuk dipakai Bank Soal, rapor, guru mapel, dan jadwal." />
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
			saving={saveBusy}
			saveLabel={saveBusy ? 'Menyimpan...' : 'Simpan Perubahan'}
			discardLabel="Batalkan"
			onsave={() => void saveChangedSubjects()}
			ondiscard={() => resetSubjectDrafts()}
		/>
	</AsyncContent>
</div>
