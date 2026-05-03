<script lang="ts">
	import { onMount } from 'svelte';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Input } from '$lib/components/ui/input';
	import { Textarea } from '$lib/components/ui/textarea';
	import { Badge } from '$lib/components/ui/badge';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { toast } from '$lib/components/ui/sonner';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import OperationStatusPanel from '$lib/components/OperationStatusPanel.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { confirmChallenge } from '$lib/confirm-dialog';
	import { clientApiPath, clientApiPathWithQuery, readClientApiData, readClientJson } from '$lib/client/api';

	type Subject = { id: string; code: string; name: string };
	type SchoolClass = { id: string; code: string; name: string; level: string };
	type NonTestAssessment = {
		id: string;
		subject_id: string;
		subject_name: string;
		subject_code: string;
		class_id: string;
		class_name: string;
		class_level: string;
		assessment_type: string;
		title: string;
		description: string;
		instruction_html: string;
		rubric_html: string;
		evidence_requirements: string;
		mode: 'beginner' | 'advance';
		scoring_scale: string;
		max_score: number;
		weight: number;
		due_at: string | null;
		status: string;
		assessor_username: string;
		checklist: string[];
		total_submissions: number;
		reviewed_submissions: number;
	};
	type AssessmentPayload = {
		items?: NonTestAssessment[];
		meta?: { total?: number; limit?: number; offset?: number };
		error?: string;
		message?: string;
	};
	type AcademicPayload = {
		subjects?: Subject[];
		classes?: SchoolClass[];
		error?: string;
		message?: string;
	};
	type AssessmentOverview = {
		assessments: NonTestAssessment[];
		subjects: Subject[];
		classes: SchoolClass[];
		totalItems: number;
	};
	type OperationState = {
		tone: 'success' | 'error' | 'warning' | 'info';
		title: string;
		message: string;
	};

	const ASSESSMENT_TYPES = [
		{ value: 'praktik', label: 'Praktik' },
		{ value: 'portofolio', label: 'Portofolio' },
		{ value: 'proyek', label: 'Proyek' },
		{ value: 'penugasan', label: 'Penugasan' },
		{ value: 'observasi', label: 'Observasi' },
		{ value: 'lainnya', label: 'Lainnya' },
	] as const;
	const STATUSES = [
		{ value: 'draft', label: 'Draft' },
		{ value: 'active', label: 'Aktif' },
		{ value: 'closed', label: 'Ditutup' },
		{ value: 'archived', label: 'Arsip' },
	] as const;

	let assessments = $state<NonTestAssessment[]>([]);
	let subjects = $state<Subject[]>([]);
	let classes = $state<SchoolClass[]>([]);
	let overviewPromise = $state<Promise<AssessmentOverview> | null>(null);
	let requestId = 0;

	let showForm = $state(false);
	let editingId = $state('');
	let saving = $state(false);
	let deleteBusyId = $state('');
	let operationState = $state<OperationState | null>(null);

	let filterSearch = $state('');
	let filterStatus = $state('');
	let filterType = $state('');
	let filterSubjectId = $state('');

	let formMode = $state<'beginner' | 'advance'>('beginner');
	let formSubjectId = $state('');
	let formClassId = $state('');
	let formType = $state('penugasan');
	let formTitle = $state('');
	let formDescription = $state('');
	let formInstruction = $state('');
	let formRubric = $state('');
	let formEvidence = $state('');
	let formStatus = $state('draft');
	let formDueAt = $state('');
	let formMaxScore = $state(100);
	let formWeight = $state(1);
	let formAssessor = $state('');
	let checklistText = $state('');

	let summary = $derived({
		total: assessments.length,
		active: assessments.filter((item) => item.status === 'active').length,
		draft: assessments.filter((item) => item.status === 'draft').length,
		closed: assessments.filter((item) => item.status === 'closed').length,
	});

	function assessmentTypeLabel(value: string) {
		return ASSESSMENT_TYPES.find((item) => item.value === value)?.label ?? value;
	}

	function statusLabel(value: string) {
		return STATUSES.find((item) => item.value === value)?.label ?? value;
	}

	function statusBadgeClass(value: string) {
		if (value === 'active') return 'border-emerald-200 bg-emerald-100 text-emerald-800';
		if (value === 'closed') return 'border-slate-200 bg-slate-100 text-slate-700';
		if (value === 'archived') return 'border-amber-200 bg-amber-100 text-amber-800';
		return 'border-blue-200 bg-blue-50 text-blue-800';
	}

	function fetchAssessmentsPath() {
		const params = new URLSearchParams();
		params.set('limit', '50');
		if (filterSearch.trim()) params.set('q', filterSearch.trim());
		if (filterStatus) params.set('status', filterStatus);
		if (filterType) params.set('assessment_type', filterType);
		if (filterSubjectId) params.set('subject_id', filterSubjectId);
		return clientApiPathWithQuery('/api/cbt/non-test-assessments', params);
	}

	async function fetchOverview(): Promise<AssessmentOverview> {
		const [assessmentPayload, academicPayload] = await Promise.all([
			fetch(fetchAssessmentsPath()).then((response) => readClientApiData<AssessmentPayload>(response, 'Gagal memuat asesmen non-tes.')),
			fetch('/api/academic').then((response) => readClientApiData<AcademicPayload>(response, 'Gagal memuat data akademik.')),
		]);
		return {
			assessments: assessmentPayload.items ?? [],
			totalItems: assessmentPayload.meta?.total ?? assessmentPayload.items?.length ?? 0,
			subjects: academicPayload.subjects ?? [],
			classes: academicPayload.classes ?? [],
		};
	}

	function applyOverview(overview: AssessmentOverview) {
		assessments = overview.assessments;
		subjects = overview.subjects;
		classes = overview.classes;
	}

	function load() {
		const currentRequestId = ++requestId;
		overviewPromise = fetchOverview().then((overview) => {
			if (currentRequestId !== requestId) return { assessments, subjects, classes, totalItems: assessments.length };
			applyOverview(overview);
			return overview;
		}).catch((error: unknown) => {
			if (currentRequestId === requestId) throw error;
			return { assessments, subjects, classes, totalItems: assessments.length };
		});
	}

	async function refresh() {
		const currentRequestId = ++requestId;
		try {
			const overview = await fetchOverview();
			if (currentRequestId !== requestId) return;
			applyOverview(overview);
			overviewPromise = Promise.resolve(overview);
		} catch (error) {
			if (currentRequestId !== requestId) return;
			overviewPromise = Promise.resolve({ assessments, subjects, classes, totalItems: assessments.length });
			toast.error(errorMessage(error, 'Gagal memuat ulang asesmen non-tes.'));
		}
	}

	function retry(reset?: () => void) {
		reset?.();
		load();
	}

	function errorMessage(error: unknown, fallback: string) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return fallback;
	}

	function handleRenderError(error: unknown) {
		console.error('CBT non-test assessment render failed', error);
	}

	function resetForm() {
		editingId = '';
		formMode = 'beginner';
		formSubjectId = '';
		formClassId = '';
		formType = 'penugasan';
		formTitle = '';
		formDescription = '';
		formInstruction = '';
		formRubric = '';
		formEvidence = '';
		formStatus = 'draft';
		formDueAt = '';
		formMaxScore = 100;
		formWeight = 1;
		formAssessor = '';
		checklistText = '';
	}

	function openCreateForm() {
		resetForm();
		showForm = true;
	}

	function editAssessment(item: NonTestAssessment) {
		editingId = item.id;
		formMode = item.mode ?? 'beginner';
		formSubjectId = item.subject_id;
		formClassId = item.class_id ?? '';
		formType = item.assessment_type;
		formTitle = item.title;
		formDescription = item.description ?? '';
		formInstruction = item.instruction_html ?? '';
		formRubric = item.rubric_html ?? '';
		formEvidence = item.evidence_requirements ?? '';
		formStatus = item.status ?? 'draft';
		formDueAt = toDateTimeLocal(item.due_at);
		formMaxScore = item.max_score ?? 100;
		formWeight = item.weight ?? 1;
		formAssessor = item.assessor_username ?? '';
		checklistText = Array.isArray(item.checklist) ? item.checklist.join('\n') : '';
		showForm = true;
	}

	function checklistItems() {
		return checklistText
			.split('\n')
			.map((item) => item.trim())
			.filter(Boolean);
	}

	function formPayload() {
		return {
			subject_id: formSubjectId,
			class_id: formClassId,
			assessment_type: formType,
			title: formTitle,
			description: formDescription,
			instruction_html: formInstruction,
			rubric_html: formRubric,
			evidence_requirements: formEvidence,
			mode: formMode,
			scoring_scale: '0_100',
			max_score: Number(formMaxScore) || 100,
			weight: Number(formWeight) || 1,
			due_at: formDueAt,
			status: formStatus,
			assessor_username: formAssessor,
			checklist: checklistItems(),
		};
	}

	async function saveAssessment() {
		if (!formSubjectId || !formTitle.trim()) {
			toast.error('Mata pelajaran dan judul wajib diisi.');
			return;
		}
		saving = true;
		try {
			const path = editingId
				? clientApiPath`/api/cbt/non-test-assessments/${editingId}`
				: '/api/cbt/non-test-assessments';
			const response = await fetch(path, {
				method: editingId ? 'PUT' : 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(formPayload()),
			});
			await readClientJson<unknown>(response);
			operationState = {
				tone: 'success',
				title: editingId ? 'Asesmen Diperbarui' : 'Asesmen Dibuat',
				message: editingId
					? 'Instruksi, rubrik, status, dan pengaturan asesmen non-tes sudah tersimpan.'
					: 'Asesmen non-tes baru sudah masuk daftar dan siap disiapkan untuk penilaian manual.',
			};
			toast.success(editingId ? 'Asesmen diperbarui' : 'Asesmen dibuat');
			showForm = false;
			resetForm();
			await refresh();
		} catch (error) {
			toast.error(errorMessage(error, 'Gagal menyimpan asesmen non-tes.'));
		} finally {
			saving = false;
		}
	}

	async function deleteAssessment(item: NonTestAssessment) {
		if (!(await confirmChallenge({
			title: 'Hapus Asesmen Non-Tes',
			message: `Asesmen "${item.title}" akan dihapus beserta data pengumpulan yang terkait.`,
			challenge: 'HAPUS',
			confirmLabel: 'Hapus',
			tone: 'danger'
		}))) return;
		deleteBusyId = item.id;
		try {
			const response = await fetch(clientApiPath`/api/cbt/non-test-assessments/${item.id}`, { method: 'DELETE' });
			await readClientJson<unknown>(response);
			operationState = {
				tone: 'warning',
				title: 'Asesmen Dihapus',
				message: `Asesmen "${item.title}" sudah dihapus dari modul non-tes.`,
			};
			toast.success('Asesmen dihapus');
			await refresh();
		} catch (error) {
			toast.error(errorMessage(error, 'Gagal menghapus asesmen non-tes.'));
		} finally {
			deleteBusyId = '';
		}
	}

	function applyFilters() {
		load();
	}

	function clearFilters() {
		filterSearch = '';
		filterStatus = '';
		filterType = '';
		filterSubjectId = '';
		load();
	}

	function toDateTimeLocal(value: string | null | undefined) {
		if (!value) return '';
		const date = new Date(value);
		if (Number.isNaN(date.getTime())) return '';
		const local = new Date(date.getTime() - date.getTimezoneOffset() * 60_000);
		return local.toISOString().slice(0, 16);
	}

	onMount(() => {
		load();
	});
</script>

<svelte:head><title>Asesmen Non-Tes CBT — MTsN 2 Kolut</title></svelte:head>

<div class="space-y-5">
	<div class="flex flex-wrap items-start justify-between gap-3">
		<div>
			<p class="text-xs font-semibold uppercase tracking-[0.24em] text-emerald-700">CBT / Penilaian Manual</p>
			<h1 class="mt-1 text-2xl font-semibold text-slate-900">Asesmen Non-Tes</h1>
			<p class="mt-1 max-w-3xl text-sm text-slate-600">
				Kelola praktik, portofolio, proyek, penugasan, dan observasi tanpa mencampurnya dengan bank soal ujian token.
			</p>
		</div>
		<LoadingButton onclick={openCreateForm}>+ Buat Asesmen</LoadingButton>
	</div>

	<div class="grid gap-3 md:grid-cols-4">
		<Card.Root class="border-emerald-100 shadow-sm">
			<Card.Content class="p-4">
				<p class="text-xs font-medium uppercase tracking-wider text-slate-500">Total Modul</p>
				<p class="mt-1 text-2xl font-semibold text-emerald-800">{summary.total}</p>
			</Card.Content>
		</Card.Root>
		<Card.Root class="border-emerald-100 shadow-sm">
			<Card.Content class="p-4">
				<p class="text-xs font-medium uppercase tracking-wider text-slate-500">Aktif</p>
				<p class="mt-1 text-2xl font-semibold text-emerald-800">{summary.active}</p>
			</Card.Content>
		</Card.Root>
		<Card.Root class="border-blue-100 shadow-sm">
			<Card.Content class="p-4">
				<p class="text-xs font-medium uppercase tracking-wider text-slate-500">Draft</p>
				<p class="mt-1 text-2xl font-semibold text-blue-800">{summary.draft}</p>
			</Card.Content>
		</Card.Root>
		<Card.Root class="border-slate-200 shadow-sm">
			<Card.Content class="p-4">
				<p class="text-xs font-medium uppercase tracking-wider text-slate-500">Ditutup</p>
				<p class="mt-1 text-2xl font-semibold text-slate-800">{summary.closed}</p>
			</Card.Content>
		</Card.Root>
	</div>

	{#if operationState}
		<OperationStatusPanel {...operationState} />
	{/if}

	{#if showForm}
		<Card.Root class="border-emerald-100 shadow-sm">
			<Card.Header class="border-b border-slate-100 pb-3">
				<div class="flex flex-wrap items-center justify-between gap-3">
					<div>
						<Card.Title class="text-base">{editingId ? 'Edit Asesmen Non-Tes' : 'Buat Asesmen Non-Tes'}</Card.Title>
						<Card.Description>Mode Pemula ringkas; mode Advance membuka rubrik, bukti, dan checklist observasi.</Card.Description>
					</div>
					<div class="inline-flex rounded-md border border-slate-200 bg-white p-1">
						<button
							type="button"
							class={`rounded px-3 py-1.5 text-xs font-semibold ${formMode === 'beginner' ? 'bg-emerald-700 text-white' : 'text-slate-600 hover:bg-slate-50'}`}
							onclick={() => (formMode = 'beginner')}
						>
							Pemula
						</button>
						<button
							type="button"
							class={`rounded px-3 py-1.5 text-xs font-semibold ${formMode === 'advance' ? 'bg-emerald-700 text-white' : 'text-slate-600 hover:bg-slate-50'}`}
							onclick={() => (formMode = 'advance')}
						>
							Advance
						</button>
					</div>
				</div>
			</Card.Header>
			<Card.Content class="space-y-4 p-4">
				<div class="grid gap-3 lg:grid-cols-[1.2fr_0.8fr_0.8fr_0.7fr]">
					<div>
						<label for="non-test-subject" class="mb-1 block text-xs font-semibold uppercase tracking-wide text-slate-600">Mata Pelajaran <span class="text-red-500">*</span></label>
						<select id="non-test-subject" class="h-9 w-full rounded-md border border-input bg-background px-3 text-sm" bind:value={formSubjectId}>
							<option value="">-- Pilih Mapel --</option>
							{#each subjects as subject (subject.id)}
								<option value={subject.id}>{subject.code} — {subject.name}</option>
							{/each}
						</select>
					</div>
					<div>
						<label for="non-test-class" class="mb-1 block text-xs font-semibold uppercase tracking-wide text-slate-600">Kelas</label>
						<select id="non-test-class" class="h-9 w-full rounded-md border border-input bg-background px-3 text-sm" bind:value={formClassId}>
							<option value="">Semua / belum ditentukan</option>
							{#each classes as kelas (kelas.id)}
								<option value={kelas.id}>{kelas.name}</option>
							{/each}
						</select>
					</div>
					<div>
						<label for="non-test-type" class="mb-1 block text-xs font-semibold uppercase tracking-wide text-slate-600">Bentuk</label>
						<select id="non-test-type" class="h-9 w-full rounded-md border border-input bg-background px-3 text-sm" bind:value={formType}>
							{#each ASSESSMENT_TYPES as item (item.value)}
								<option value={item.value}>{item.label}</option>
							{/each}
						</select>
					</div>
					<div>
						<label for="non-test-status" class="mb-1 block text-xs font-semibold uppercase tracking-wide text-slate-600">Status</label>
						<select id="non-test-status" class="h-9 w-full rounded-md border border-input bg-background px-3 text-sm" bind:value={formStatus}>
							{#each STATUSES as item (item.value)}
								<option value={item.value}>{item.label}</option>
							{/each}
						</select>
					</div>
				</div>

				<div class="grid gap-3 lg:grid-cols-[1fr_12rem_10rem]">
					<div>
						<label for="non-test-title" class="mb-1 block text-xs font-semibold uppercase tracking-wide text-slate-600">Judul <span class="text-red-500">*</span></label>
						<Input id="non-test-title" placeholder="Mis. Praktik membaca teks qiraah" bind:value={formTitle} />
					</div>
					<div>
						<label for="non-test-due" class="mb-1 block text-xs font-semibold uppercase tracking-wide text-slate-600">Tenggat</label>
						<Input id="non-test-due" type="datetime-local" bind:value={formDueAt} />
					</div>
					<div>
						<label for="non-test-weight" class="mb-1 block text-xs font-semibold uppercase tracking-wide text-slate-600">Bobot</label>
						<Input id="non-test-weight" type="number" min={0.1} step={0.1} bind:value={formWeight} />
					</div>
				</div>

				<div>
					<label for="non-test-instruction" class="mb-1 block text-xs font-semibold uppercase tracking-wide text-slate-600">Instruksi Tugas</label>
					<Textarea id="non-test-instruction" rows={3} placeholder="Instruksi singkat untuk guru/siswa..." bind:value={formInstruction} />
				</div>

				<div class="grid gap-3 lg:grid-cols-2">
					<div>
						<label for="non-test-rubric" class="mb-1 block text-xs font-semibold uppercase tracking-wide text-slate-600">Rubrik Ringkas</label>
						<Textarea id="non-test-rubric" rows={formMode === 'advance' ? 5 : 3} placeholder="Kriteria penilaian utama..." bind:value={formRubric} />
					</div>
					<div>
						<label for="non-test-evidence" class="mb-1 block text-xs font-semibold uppercase tracking-wide text-slate-600">Bukti yang Dikumpulkan</label>
						<Textarea id="non-test-evidence" rows={formMode === 'advance' ? 5 : 3} placeholder="Foto, dokumen, link portofolio, catatan observasi..." bind:value={formEvidence} />
					</div>
				</div>

				{#if formMode === 'advance'}
					<div class="grid gap-3 lg:grid-cols-[1fr_12rem_14rem]">
						<div>
							<label for="non-test-description" class="mb-1 block text-xs font-semibold uppercase tracking-wide text-slate-600">Deskripsi Internal</label>
							<Textarea id="non-test-description" rows={3} bind:value={formDescription} />
						</div>
						<div>
							<label for="non-test-max-score" class="mb-1 block text-xs font-semibold uppercase tracking-wide text-slate-600">Skor Maks</label>
							<Input id="non-test-max-score" type="number" min={1} bind:value={formMaxScore} />
						</div>
						<div>
							<label for="non-test-assessor" class="mb-1 block text-xs font-semibold uppercase tracking-wide text-slate-600">Penilai</label>
							<Input id="non-test-assessor" placeholder="username/penanggung jawab" bind:value={formAssessor} />
						</div>
					</div>
					<div>
						<label for="non-test-checklist" class="mb-1 block text-xs font-semibold uppercase tracking-wide text-slate-600">Checklist Observasi</label>
						<Textarea id="non-test-checklist" rows={4} placeholder="Satu indikator per baris" bind:value={checklistText} />
					</div>
				{/if}

				<div class="flex flex-wrap justify-end gap-2 border-t border-slate-100 pt-3">
					<LoadingButton variant="outline" onclick={() => { showForm = false; resetForm(); }}>Batal</LoadingButton>
					<LoadingButton onclick={saveAssessment} loading={saving} loadingLabel="Menyimpan...">
						{editingId ? 'Simpan Perubahan' : 'Simpan Asesmen'}
					</LoadingButton>
				</div>
			</Card.Content>
		</Card.Root>
	{/if}

	<Card.Root class="border-slate-200 shadow-sm">
		<Card.Content class="grid gap-3 p-4 lg:grid-cols-[1fr_12rem_12rem_14rem_auto_auto] lg:items-end">
			<div>
				<label for="non-test-search" class="mb-1 block text-xs font-semibold uppercase tracking-wide text-slate-600">Cari</label>
				<Input id="non-test-search" placeholder="Judul, bukti, mapel, kelas..." bind:value={filterSearch} onkeydown={(event) => { if (event.key === 'Enter') applyFilters(); }} />
			</div>
			<div>
				<label for="non-test-filter-status" class="mb-1 block text-xs font-semibold uppercase tracking-wide text-slate-600">Status</label>
				<select id="non-test-filter-status" class="h-9 w-full rounded-md border border-input bg-background px-3 text-sm" bind:value={filterStatus}>
					<option value="">Semua</option>
					{#each STATUSES as item (item.value)}
						<option value={item.value}>{item.label}</option>
					{/each}
				</select>
			</div>
			<div>
				<label for="non-test-filter-type" class="mb-1 block text-xs font-semibold uppercase tracking-wide text-slate-600">Bentuk</label>
				<select id="non-test-filter-type" class="h-9 w-full rounded-md border border-input bg-background px-3 text-sm" bind:value={filterType}>
					<option value="">Semua</option>
					{#each ASSESSMENT_TYPES as item (item.value)}
						<option value={item.value}>{item.label}</option>
					{/each}
				</select>
			</div>
			<div>
				<label for="non-test-filter-subject" class="mb-1 block text-xs font-semibold uppercase tracking-wide text-slate-600">Mapel</label>
				<select id="non-test-filter-subject" class="h-9 w-full rounded-md border border-input bg-background px-3 text-sm" bind:value={filterSubjectId}>
					<option value="">Semua</option>
					{#each subjects as subject (subject.id)}
						<option value={subject.id}>{subject.code} — {subject.name}</option>
					{/each}
				</select>
			</div>
			<LoadingButton onclick={applyFilters}>Terapkan</LoadingButton>
			<LoadingButton variant="outline" onclick={clearFilters}>Reset</LoadingButton>
		</Card.Content>
	</Card.Root>

	<AsyncContent promise={overviewPromise} onerror={handleRenderError}>
		{#snippet pending()}
			<Card.Root class="border-slate-200 shadow-sm">
				<Card.Content class="space-y-3 p-5">
					{#each Array.from({ length: 5 }) as _, index (`non-test-skeleton-${index}`)}
						<div class="grid gap-3 lg:grid-cols-[1fr_8rem_8rem_8rem_10rem_auto] lg:items-center">
							<Skeleton class="h-5 w-52" />
							<Skeleton class="h-5 w-20" />
							<Skeleton class="h-5 w-20" />
							<Skeleton class="h-5 w-20" />
							<Skeleton class="h-5 w-24" />
							<Skeleton class="h-9 w-28 justify-self-end" />
						</div>
					{/each}
				</Card.Content>
			</Card.Root>
		{/snippet}

		{#snippet failed(error, reset)}
			<RecoveryPanel
				title="Asesmen Non-Tes Belum Tersaji"
				message={errorMessage(error, 'Data asesmen non-tes belum dapat dimuat.')}
				onRetry={() => retry(reset)}
			/>
		{/snippet}

		{#snippet children(value)}
			{@const overview = value as AssessmentOverview}
			{@const currentAssessments = overview.assessments}
			<Card.Root class="overflow-hidden border-slate-200 shadow-sm">
				<Card.Header class="pb-2">
					<Card.Title class="text-base">Daftar Asesmen ({overview.totalItems})</Card.Title>
					<Card.Description>Semua item di sini dinilai manual dan tidak masuk runtime ujian token.</Card.Description>
				</Card.Header>
				<Card.Content class="p-0">
					<div class="hidden overflow-x-auto lg:block">
						<Table.Root>
							<Table.Header>
								<Table.Row>
									<Table.Head>Asesmen</Table.Head>
									<Table.Head>Mapel</Table.Head>
									<Table.Head>Bentuk</Table.Head>
									<Table.Head>Status</Table.Head>
									<Table.Head>Progress</Table.Head>
									<Table.Head></Table.Head>
								</Table.Row>
							</Table.Header>
							<Table.Body>
								{#each currentAssessments as item (item.id)}
									<Table.Row>
										<Table.Cell>
											<div>
												<p class="font-medium text-slate-900">{item.title}</p>
												<p class="mt-0.5 text-xs text-slate-500">
													{item.class_name || 'Lintas kelas'} · Bobot {item.weight} · Skor {item.max_score}
												</p>
											</div>
										</Table.Cell>
										<Table.Cell>
											<Badge variant="outline" class="text-xs">{item.subject_code}</Badge>
										</Table.Cell>
										<Table.Cell class="text-sm text-slate-700">{assessmentTypeLabel(item.assessment_type)}</Table.Cell>
										<Table.Cell>
											<Badge class={`border text-xs ${statusBadgeClass(item.status)}`}>{statusLabel(item.status)}</Badge>
										</Table.Cell>
										<Table.Cell class="text-sm text-slate-600">
											{item.reviewed_submissions}/{item.total_submissions} dinilai
										</Table.Cell>
										<Table.Cell>
											<div class="flex justify-end gap-2">
												<LoadingButton variant="outline" size="xs" onclick={() => editAssessment(item)}>Edit</LoadingButton>
												<LoadingButton
													variant="destructive"
													size="xs"
													onclick={() => deleteAssessment(item)}
													loading={deleteBusyId === item.id}
													disabled={deleteBusyId !== '' && deleteBusyId !== item.id}
													loadingLabel="Hapus..."
												>
													Hapus
												</LoadingButton>
											</div>
										</Table.Cell>
									</Table.Row>
								{:else}
									<Table.Row>
										<Table.Cell colspan={6} class="py-8 text-center text-slate-400">Belum ada asesmen non-tes.</Table.Cell>
									</Table.Row>
								{/each}
							</Table.Body>
						</Table.Root>
					</div>

					<div class="space-y-3 p-4 lg:hidden">
						{#each currentAssessments as item (item.id)}
							<div class="rounded-lg border border-slate-200 p-3">
								<div class="flex items-start justify-between gap-3">
									<div>
										<p class="font-medium text-slate-900">{item.title}</p>
										<p class="mt-1 text-xs text-slate-500">{item.subject_code} · {assessmentTypeLabel(item.assessment_type)}</p>
									</div>
									<Badge class={`border text-xs ${statusBadgeClass(item.status)}`}>{statusLabel(item.status)}</Badge>
								</div>
								<p class="mt-2 text-sm text-slate-600">{item.reviewed_submissions}/{item.total_submissions} dinilai</p>
								<div class="mt-3 flex justify-end gap-2">
									<LoadingButton variant="outline" size="xs" onclick={() => editAssessment(item)}>Edit</LoadingButton>
									<LoadingButton variant="destructive" size="xs" onclick={() => deleteAssessment(item)} loading={deleteBusyId === item.id}>Hapus</LoadingButton>
								</div>
							</div>
						{:else}
							<p class="py-6 text-center text-sm text-slate-400">Belum ada asesmen non-tes.</p>
						{/each}
					</div>
				</Card.Content>
			</Card.Root>
		{/snippet}
	</AsyncContent>
</div>
