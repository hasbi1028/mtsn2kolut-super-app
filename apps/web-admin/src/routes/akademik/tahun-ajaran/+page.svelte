<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import { AlertTriangle, CalendarClock, CheckCircle2, Download, Eye, Plus, RefreshCw, Upload } from '@lucide/svelte';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import EmptyStatePanel from '$lib/components/EmptyStatePanel.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import * as Card from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import * as Table from '$lib/components/ui/table';
	import { toast } from '$lib/components/ui/sonner';
	import { confirmAction } from '$lib/confirm-dialog';
	import { clientApiPath, clientApiPathWithQuery, readClientApiData, readClientJson } from '$lib/client/api';

	type AcademicYear = {
		id: string;
		name: string;
		start_date: string;
		end_date: string;
		is_active: boolean;
		created_at: string;
		updated_at?: string;
	};

	type AcademicOverview = {
		years: AcademicYear[];
		classes?: unknown[];
		subjects?: unknown[];
		assignments?: unknown[];
		timetableSlots?: unknown[];
	};

	type RolloverCounts = {
		classes_to_create: number;
		students_to_promote: number;
		students_without_next_class: number;
		homeroom_assignments_to_copy: number;
		subject_assignments_to_copy: number;
		timetable_slots_to_copy: number;
	};

	type RolloverClassToCreate = {
		source_class_id: string;
		source_code: string;
		source_name: string;
		source_level: string;
		target_code: string;
		target_name: string;
		target_level: string;
	};

	type RolloverApplyCounts = {
		classes_created: number;
		classes_reused: number;
		students_promoted: number;
		students_skipped: number;
		homerooms_copied: number;
		assignments_copied: number;
		timetable_slots_copied: number;
	};

	type RolloverClassApply = RolloverClassToCreate & {
		target_class_id: string;
		action: 'created' | 'reused' | 'skipped_inactive_target' | string;
	};

	type RolloverStudentWarning = {
		student_id?: string;
		nis: string;
		nisn: string;
		nama: string;
		from_class_id?: string;
		from_class_code: string;
		reason: string;
	};

	type RolloverStudentMove = {
		student_id: string;
		nis: string;
		nisn: string;
		nama: string;
		from_class_id: string;
		from_class_code: string;
		to_class_id: string;
		to_class_code: string;
		to_class_name: string;
	};

	type RolloverPreview = {
		source_academic_year_id: string;
		source_academic_year_name: string;
		target_academic_year_id: string;
		target_academic_year_name: string;
		apply_challenge: string;
		counts: RolloverCounts;
		classes_to_create: RolloverClassToCreate[];
		students_to_promote: RolloverStudentMove[];
		students_without_next_class: RolloverStudentWarning[];
		warnings: string[];
	};

	type RolloverApplyResult = {
		source_academic_year_id: string;
		source_academic_year_name: string;
		target_academic_year_id: string;
		target_academic_year_name: string;
		counts: RolloverApplyCounts;
		classes: RolloverClassApply[];
		students_promoted: RolloverStudentMove[];
		students_skipped: RolloverStudentWarning[];
		warnings: string[];
	};

	type ImportKind = 'siswa' | 'rombel' | 'guru_mapel' | 'jadwal';

	type ImportDryRunRow = {
		row: number;
		action: 'add' | 'update' | 'skip' | 'error' | string;
		summary: string;
	};

	type ImportDryRunError = {
		row: number;
		field: string;
		message: string;
	};

	type ImportDryRunResult = {
		kind: ImportKind | string;
		total_rows: number;
		add_count: number;
		update_count: number;
		skip_count: number;
		error_count: number;
		rows: ImportDryRunRow[];
		row_errors: ImportDryRunError[];
	};

	const importKinds: Array<{ value: ImportKind; label: string; description: string }> = [
		{ value: 'siswa', label: 'Siswa', description: 'NIS, identitas, gender, rombel, dan lifecycle.' },
		{ value: 'rombel', label: 'Rombel', description: 'Kode rombel, tingkat, tahun ajaran, dan status aktif.' },
		{ value: 'guru_mapel', label: 'Guru Mapel', description: 'Pasangan rombel, mapel, dan guru pengampu.' },
		{ value: 'jadwal', label: 'Jadwal Pelajaran', description: 'Slot hari, jam, ruang, dan catatan.' }
	];

	let years = $state<AcademicYear[]>([]);
	let overviewPromise = $state<Promise<AcademicOverview> | null>(null);
	let overviewRequestId = 0;
	let refreshBusy = $state(false);

	let yearName = $state('');
	let startDate = $state('');
	let endDate = $state('');
	let createBusy = $state(false);
	let activateBusyId = $state('');

	let sourceYearId = $state('');
	let targetYearId = $state('');
	let previewBusy = $state(false);
	let rolloverPreview = $state<RolloverPreview | null>(null);
	let applyBusy = $state(false);
	let applyChallengeInput = $state('');
	let rolloverApplyResult = $state<RolloverApplyResult | null>(null);

	let importKind = $state<ImportKind>('siswa');
	let importFile = $state<File | null>(null);
	let importBusy = $state(false);
	let dryRunResult = $state<ImportDryRunResult | null>(null);

	const activeYear = $derived(years.find((year) => year.is_active) ?? null);
	const inactiveYears = $derived(years.filter((year) => !year.is_active));
	const targetYearOptions = $derived(years.filter((year) => year.id !== sourceYearId));
	const createDisabled = $derived(!yearName.trim() || !startDate || !endDate || createBusy);
	const canPreview = $derived(Boolean(targetYearId) && !previewBusy);
	const previewMatchesSelection = $derived(Boolean(
		rolloverPreview
			&& rolloverPreview.target_academic_year_id === targetYearId
			&& (!sourceYearId || rolloverPreview.source_academic_year_id === sourceYearId)
	));
	const canApplyRollover = $derived(Boolean(
		rolloverPreview
			&& previewMatchesSelection
			&& applyChallengeInput.trim() === rolloverPreview.apply_challenge
			&& !applyBusy
			&& !previewBusy
	));
	const selectedImportKind = $derived(importKinds.find((item) => item.value === importKind) ?? importKinds[0]);

	async function fetchOverview(): Promise<AcademicOverview> {
		const data = await fetch('/api/academic').then((response) =>
			readClientApiData<Partial<AcademicOverview>>(response, 'Gagal memuat tahun ajaran akademik')
		);
		return {
			years: data.years ?? [],
			classes: data.classes ?? [],
			subjects: data.subjects ?? [],
			assignments: data.assignments ?? [],
			timetableSlots: data.timetableSlots ?? []
		};
	}

	function applyOverview(data: AcademicOverview) {
		years = data.years;
		const active = data.years.find((year) => year.is_active);
		if (!sourceYearId && active) sourceYearId = active.id;
		if (!targetYearId) {
			targetYearId = data.years.find((year) => year.id !== (active?.id ?? ''))?.id ?? '';
		}
	}

	function loadOverview() {
		const requestId = ++overviewRequestId;
		overviewPromise = fetchOverview()
			.then((data) => {
				if (requestId === overviewRequestId) applyOverview(data);
				return data;
			})
			.catch((error: unknown) => {
				if (requestId === overviewRequestId) throw error;
				return { years };
			});
	}

	async function refreshOverview() {
		refreshBusy = true;
		try {
			const data = await fetchOverview();
			applyOverview(data);
			overviewPromise = Promise.resolve(data);
		} catch (error) {
			toast.error(errorMessage(error, 'Gagal memuat ulang tahun ajaran.'));
		} finally {
			refreshBusy = false;
		}
	}

	function retryOverview(reset?: () => void) {
		reset?.();
		loadOverview();
	}

	function errorMessage(error: unknown, fallback: string) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return fallback;
	}

	function handleRenderError(error: unknown) {
		console.error('Tahun ajaran render failed', error);
	}

	function resetRolloverApplyState() {
		applyChallengeInput = '';
		rolloverApplyResult = null;
	}

	function formatDate(value: string) {
		return value?.slice(0, 10) || '-';
	}

	function currentSemesterLabel() {
		const month = new Date().getMonth() + 1;
		return month >= 7 && month <= 12 ? 'Ganjil' : 'Genap';
	}

	function templateHref(kind: ImportKind) {
		return clientApiPath`/api/academic/import-export/templates/${kind}`;
	}

	function suggestNextYear() {
		const match = activeYear?.name.match(/^(\d{4})\/(\d{4})$/);
		if (!match) return;
		const start = Number(match[2]);
		if (!Number.isFinite(start)) return;
		yearName = `${start}/${start + 1}`;
		startDate = `${start}-07-01`;
		endDate = `${start + 1}-06-30`;
	}

	async function createYear() {
		if (createDisabled) return;
		createBusy = true;
		try {
			const params = new URLSearchParams({ entity: 'years' });
			const response = await fetch(clientApiPathWithQuery('/api/academic', params), {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					name: yearName.trim(),
					start_date: startDate,
					end_date: endDate,
					is_active: false
				})
			});
			await readClientJson<AcademicYear>(response);
			toast.success('Tahun ajaran baru tersimpan sebagai nonaktif.');
			yearName = '';
			startDate = '';
			endDate = '';
			await refreshOverview();
		} catch (error) {
			toast.error(errorMessage(error, 'Gagal membuat tahun ajaran.'));
		} finally {
			createBusy = false;
		}
	}

	async function activateYear(year: AcademicYear) {
		if (year.is_active || activateBusyId) return;
		const confirmed = await confirmAction({
			title: 'Aktifkan Tahun Ajaran',
			message: `Aktifkan ${year.name} sebagai tahun ajaran berjalan? Tahun aktif lama akan dinonaktifkan tanpa menghapus data.`,
			confirmLabel: 'Aktifkan',
			cancelLabel: 'Batal',
			tone: 'warning',
			challenge: 'AKTIFKAN'
		});
		if (!confirmed) return;
		activateBusyId = year.id;
		try {
			const response = await fetch(clientApiPath`/api/academic/years/${year.id}/activate`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ confirmation: 'AKTIFKAN' })
			});
			await readClientJson<AcademicYear>(response);
			toast.success(`${year.name} sudah menjadi tahun ajaran aktif.`);
			sourceYearId = year.id;
			rolloverPreview = null;
			resetRolloverApplyState();
			await refreshOverview();
		} catch (error) {
			toast.error(errorMessage(error, 'Gagal mengaktifkan tahun ajaran.'));
		} finally {
			activateBusyId = '';
		}
	}

	async function previewRollover() {
		if (!canPreview) return;
		previewBusy = true;
		resetRolloverApplyState();
		try {
			const response = await fetch('/api/academic/year-rollover/preview', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					source_academic_year_id: sourceYearId,
					target_academic_year_id: targetYearId
				})
			});
			rolloverPreview = await readClientApiData<RolloverPreview>(response, 'Preview kenaikan tidak tersedia.');
			toast.success('Preview kenaikan tahun ajaran selesai.');
		} catch (error) {
			rolloverPreview = null;
			resetRolloverApplyState();
			toast.error(errorMessage(error, 'Gagal membuat preview kenaikan.'));
		} finally {
			previewBusy = false;
		}
	}

	async function applyRollover() {
		if (!rolloverPreview || !canApplyRollover) return;
		applyBusy = true;
		try {
			const response = await fetch('/api/academic/year-rollover/apply', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					source_academic_year_id: rolloverPreview.source_academic_year_id,
					target_academic_year_id: rolloverPreview.target_academic_year_id,
					confirmation: applyChallengeInput.trim()
				})
			});
			rolloverApplyResult = await readClientApiData<RolloverApplyResult>(response, 'Apply rollover tidak tersedia.');
			applyChallengeInput = '';
			toast.success('Apply rollover tahun ajaran selesai.');
			await refreshOverview();
		} catch (error) {
			toast.error(errorMessage(error, 'Gagal apply rollover tahun ajaran.'));
		} finally {
			applyBusy = false;
		}
	}

	function onImportFileChange(event: Event) {
		const input = event.currentTarget as HTMLInputElement | null;
		importFile = input?.files?.[0] ?? null;
		dryRunResult = null;
	}

	async function dryRunImport() {
		if (!importFile || importBusy) return;
		importBusy = true;
		try {
			const csv = await importFile.text();
			const response = await fetch('/api/academic/import-export/dry-run', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ kind: importKind, csv })
			});
			dryRunResult = await readClientApiData<ImportDryRunResult>(response, 'Dry-run import tidak tersedia.');
			toast.success('Dry-run import selesai. Belum ada data yang diubah.');
		} catch (error) {
			dryRunResult = null;
			toast.error(errorMessage(error, 'Gagal melakukan dry-run import.'));
		} finally {
			importBusy = false;
		}
	}

	function actionBadgeClass(action: string) {
		switch (action) {
			case 'add':
				return 'border-primary/20 bg-primary/10 text-primary';
			case 'update':
				return 'border-warning/30 bg-warning/10 text-warning';
			case 'error':
				return 'border-destructive/30 bg-destructive/10 text-destructive';
			default:
				return 'border-border bg-muted text-muted-foreground';
		}
	}

	function actionLabel(action: string) {
		switch (action) {
			case 'add':
				return 'Tambah';
			case 'update':
				return 'Ubah';
			case 'skip':
				return 'Lewati';
			case 'error':
				return 'Error';
			default:
				return action;
		}
	}

	onMount(() => {
		loadOverview();
	});
</script>

<svelte:head>
	<title>Tahun Ajaran | MTsN 2 Kolaka Utara</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
		<div>
			<p class="text-sm font-medium text-primary">Akademik</p>
			<h1 class="text-2xl font-semibold tracking-normal text-foreground">Tahun Ajaran & Semester</h1>
			<p class="mt-1 text-sm leading-6 text-muted-foreground">Kelola periode akademik, preview rollover, dan import-export data dasar secara aman.</p>
		</div>
		<div class="flex flex-wrap gap-2">
			<Button variant="outline" href={resolve('/akademik')}>
				Dashboard Akademik
			</Button>
			<Button variant="outline" onclick={() => void refreshOverview()} disabled={refreshBusy} aria-label="Muat ulang tahun ajaran">
				<RefreshCw class={`mr-2 size-4 ${refreshBusy ? 'animate-spin' : ''}`} />
				Muat Ulang
			</Button>
		</div>
	</div>

	<AsyncContent promise={overviewPromise} onerror={handleRenderError}>
		{#snippet pending()}
			<div class="space-y-4">
				<div class="grid gap-3 md:grid-cols-4">
					{#each ['w-20', 'w-24', 'w-16', 'w-28'] as width}
						<div class="rounded-lg border border-border bg-card p-4">
							<Skeleton class="h-4 w-28" />
							<Skeleton class={`mt-3 h-8 ${width}`} />
							<Skeleton class="mt-2 h-4 w-full" />
						</div>
					{/each}
				</div>
				<Skeleton class="h-72 w-full" />
			</div>
		{/snippet}

		{#snippet failed(error, reset)}
			<div class="rounded-lg border border-border bg-card p-6">
				<EmptyStatePanel
					title="Tahun ajaran belum dapat dimuat"
					description={errorMessage(error, 'Periksa koneksi backend lalu coba lagi.')}
					compact
				/>
				<Button class="mt-4" variant="outline" onclick={() => retryOverview(reset)}>Coba Lagi</Button>
			</div>
		{/snippet}

		<div class="grid gap-3 md:grid-cols-4">
			<div class="rounded-lg border border-primary/20 bg-primary/10 px-4 py-4">
				<p class="text-xs font-semibold uppercase text-primary">Tahun Aktif</p>
				<p class="mt-2 text-xl font-semibold text-foreground">{activeYear?.name ?? 'Belum ada'}</p>
				<p class="text-sm text-muted-foreground">Periode berjalan untuk rombel dan jadwal.</p>
			</div>
			<div class="rounded-lg border border-border bg-card px-4 py-4">
				<p class="text-xs font-semibold uppercase text-muted-foreground">Semester Operasional</p>
				<p class="mt-2 text-xl font-semibold text-foreground">{currentSemesterLabel()}</p>
				<p class="text-sm text-muted-foreground">Model semester eksplisit belum tersedia.</p>
			</div>
			<div class="rounded-lg border border-border bg-card px-4 py-4">
				<p class="text-xs font-semibold uppercase text-muted-foreground">Total Tahun</p>
				<p class="mt-2 text-xl font-semibold text-foreground">{years.length}</p>
				<p class="text-sm text-muted-foreground">Termasuk periode lama dan persiapan.</p>
			</div>
			<div class="rounded-lg border border-warning/30 bg-warning/10 px-4 py-4">
				<p class="text-xs font-semibold uppercase text-warning">Rollover Aman</p>
				<p class="mt-2 text-xl font-semibold text-foreground">Preview + Challenge</p>
				<p class="text-sm text-muted-foreground">Apply berjalan lewat Go API dan transaksi.</p>
			</div>
		</div>

		<section class="grid gap-4 xl:grid-cols-[1fr_22rem]" aria-labelledby="year-list-title">
			<Card.Root class="overflow-hidden">
				<Card.Header class="pb-3">
					<div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
						<div>
							<Card.Title id="year-list-title" class="text-base">Daftar Tahun Ajaran</Card.Title>
							<Card.Description>Aktivasi wajib lewat konfirmasi eksplisit. Data tahun lama tidak dihapus.</Card.Description>
						</div>
						<Badge class="w-fit border-primary/20 bg-primary/10 text-primary" variant="outline">{inactiveYears.length} nonaktif</Badge>
					</div>
				</Card.Header>
				<Card.Content class="p-0">
					<div class="overflow-x-auto">
						<Table.Root>
							<Table.Header>
								<Table.Row>
									<Table.Head>Nama</Table.Head>
									<Table.Head>Mulai</Table.Head>
									<Table.Head>Selesai</Table.Head>
									<Table.Head>Status</Table.Head>
									<Table.Head class="text-right">Aksi</Table.Head>
								</Table.Row>
							</Table.Header>
							<Table.Body>
								{#each years as year (year.id)}
									<Table.Row>
										<Table.Cell class="font-medium">{year.name}</Table.Cell>
										<Table.Cell class="text-muted-foreground">{formatDate(year.start_date)}</Table.Cell>
										<Table.Cell class="text-muted-foreground">{formatDate(year.end_date)}</Table.Cell>
										<Table.Cell>
											{#if year.is_active}
												<Badge class="border-primary/20 bg-primary/10 text-primary" variant="outline">
													<CheckCircle2 class="mr-1 size-3" />
													Aktif
												</Badge>
											{:else}
												<Badge variant="secondary">Nonaktif</Badge>
											{/if}
										</Table.Cell>
										<Table.Cell class="text-right">
											<LoadingButton
												size="sm"
												variant="outline"
												loading={activateBusyId === year.id}
												loadingLabel="Mengaktifkan..."
												disabled={year.is_active || (activateBusyId !== '' && activateBusyId !== year.id)}
												onclick={() => void activateYear(year)}
											>
												Aktifkan
											</LoadingButton>
										</Table.Cell>
									</Table.Row>
								{:else}
									<Table.Row>
										<Table.Cell colspan={5} class="p-5">
											<EmptyStatePanel compact title="Belum ada tahun ajaran" description="Tambahkan periode akademik pertama sebelum menyusun rombel dan jadwal." />
										</Table.Cell>
									</Table.Row>
								{/each}
							</Table.Body>
						</Table.Root>
					</div>
				</Card.Content>
			</Card.Root>

			<Card.Root>
				<Card.Header class="pb-3">
					<Card.Title class="text-base">Tambah Tahun Ajaran</Card.Title>
					<Card.Description>Tahun baru selalu dibuat nonaktif lebih dulu.</Card.Description>
				</Card.Header>
				<Card.Content class="space-y-3">
					<Button variant="outline" class="w-full justify-start" onclick={suggestNextYear} disabled={!activeYear}>
						<CalendarClock class="mr-2 size-4" />
						Isi dari tahun aktif
					</Button>
					<div>
						<label for="year-name" class="mb-1 block text-sm font-medium">Tahun Ajaran</label>
						<Input id="year-name" bind:value={yearName} placeholder="2026/2027" />
					</div>
					<div>
						<label for="year-start" class="mb-1 block text-sm font-medium">Tanggal Mulai</label>
						<Input id="year-start" type="date" bind:value={startDate} />
					</div>
					<div>
						<label for="year-end" class="mb-1 block text-sm font-medium">Tanggal Selesai</label>
						<Input id="year-end" type="date" bind:value={endDate} />
					</div>
					<LoadingButton class="w-full" loading={createBusy} loadingLabel="Menyimpan..." disabled={createDisabled} onclick={() => void createYear()}>
						<Plus class="mr-2 size-4" />
						Simpan Nonaktif
					</LoadingButton>
				</Card.Content>
			</Card.Root>
		</section>

		<section class="grid gap-4 xl:grid-cols-[24rem_1fr]" aria-labelledby="rollover-title">
			<Card.Root>
				<Card.Header class="pb-3">
					<Card.Title id="rollover-title" class="text-base">Preview Kenaikan</Card.Title>
					<Card.Description>Jalankan preview sebelum apply. Data tahun lama tidak dihapus.</Card.Description>
				</Card.Header>
				<Card.Content class="space-y-3">
					<div>
						<label for="source-year" class="mb-1 block text-sm font-medium">Tahun Sumber</label>
						<select id="source-year" bind:value={sourceYearId} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
							<option value="">Tahun aktif saat ini</option>
							{#each years as year (year.id)}
								<option value={year.id}>{year.name}{year.is_active ? ' · aktif' : ''}</option>
							{/each}
						</select>
					</div>
					<div>
						<label for="target-year" class="mb-1 block text-sm font-medium">Tahun Tujuan</label>
						<select id="target-year" bind:value={targetYearId} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
							<option value="">Pilih tahun tujuan</option>
							{#each targetYearOptions as year (year.id)}
								<option value={year.id}>{year.name}{year.is_active ? ' · aktif' : ''}</option>
							{/each}
						</select>
					</div>
					<LoadingButton class="w-full" loading={previewBusy} loadingLabel="Membuat preview..." disabled={!canPreview} onclick={() => void previewRollover()}>
						<Eye class="mr-2 size-4" />
						Lihat Preview
					</LoadingButton>
					<div class="rounded-md border border-border bg-muted/40 p-3 text-sm">
						<p class="font-medium text-foreground">Apply Rollover</p>
						<p class="mt-1 text-muted-foreground">Challenge dibuat dari hasil preview dan wajib diketik persis.</p>
						<div class="mt-3 rounded-md border border-dashed border-border bg-background px-3 py-2 font-mono text-xs text-foreground">
							{rolloverPreview?.apply_challenge ?? 'Jalankan preview dulu'}
						</div>
						<div class="mt-3">
							<label for="rollover-apply-challenge" class="mb-1 block text-sm font-medium">Challenge Apply</label>
							<Input
								id="rollover-apply-challenge"
								bind:value={applyChallengeInput}
								placeholder="Ketik challenge dari preview"
								disabled={!rolloverPreview || !previewMatchesSelection || applyBusy}
							/>
							{#if rolloverPreview && !previewMatchesSelection}
								<p class="mt-1 text-xs text-warning">Preview tidak sesuai pilihan tahun saat ini. Jalankan preview ulang.</p>
							{/if}
						</div>
						<LoadingButton class="mt-3 w-full" loading={applyBusy} loadingLabel="Apply berjalan..." disabled={!canApplyRollover} onclick={() => void applyRollover()}>
							Terapkan Rollover
						</LoadingButton>
					</div>
				</Card.Content>
			</Card.Root>

			<Card.Root>
				<Card.Header class="pb-3">
					<Card.Title class="text-base">Hasil Preview</Card.Title>
					<Card.Description>Gunakan angka ini untuk meninjau dampak sebelum apply.</Card.Description>
				</Card.Header>
				<Card.Content class="space-y-4">
					{#if rolloverPreview}
						{#if rolloverApplyResult}
							<div class="rounded-lg border border-primary/20 bg-primary/10 p-4">
								<div class="flex flex-col gap-2 sm:flex-row sm:items-start sm:justify-between">
									<div>
										<p class="text-sm font-semibold text-primary">Apply Selesai</p>
										<p class="text-sm text-muted-foreground">{rolloverApplyResult.source_academic_year_name} ke {rolloverApplyResult.target_academic_year_name}</p>
									</div>
									<Badge class="w-fit border-primary/20 bg-background text-primary" variant="outline">
										<CheckCircle2 class="mr-1 size-3" />
										Transaksional
									</Badge>
								</div>
								<div class="mt-3 grid gap-2 md:grid-cols-4">
									{#each [
										['Rombel baru', rolloverApplyResult.counts.classes_created],
										['Rombel reuse', rolloverApplyResult.counts.classes_reused],
										['Siswa naik', rolloverApplyResult.counts.students_promoted],
										['Siswa dilewati', rolloverApplyResult.counts.students_skipped],
										['Wali kelas', rolloverApplyResult.counts.homerooms_copied],
										['Guru mapel', rolloverApplyResult.counts.assignments_copied],
										['Slot jadwal', rolloverApplyResult.counts.timetable_slots_copied]
									] as item}
										<div class="rounded-md border border-primary/20 bg-background/80 p-2">
											<p class="text-xs text-muted-foreground">{item[0]}</p>
											<p class="text-lg font-semibold text-foreground">{item[1]}</p>
										</div>
									{/each}
								</div>
								{#if rolloverApplyResult.warnings.length}
									<div class="mt-3 space-y-1 text-sm text-muted-foreground">
										{#each rolloverApplyResult.warnings as warning}
											<p>{warning}</p>
										{/each}
									</div>
								{/if}
							</div>
						{/if}
						<div class="grid gap-3 md:grid-cols-3">
							<div class="rounded-lg border border-border bg-muted/40 p-3">
								<p class="text-xs text-muted-foreground">Rombel perlu dibuat</p>
								<p class="mt-1 text-2xl font-semibold">{rolloverPreview.counts.classes_to_create}</p>
							</div>
							<div class="rounded-lg border border-primary/20 bg-primary/10 p-3">
								<p class="text-xs text-primary">Siswa siap naik</p>
								<p class="mt-1 text-2xl font-semibold">{rolloverPreview.counts.students_to_promote}</p>
							</div>
							<div class="rounded-lg border border-warning/30 bg-warning/10 p-3">
								<p class="text-xs text-warning">Tanpa kelas tujuan</p>
								<p class="mt-1 text-2xl font-semibold">{rolloverPreview.counts.students_without_next_class}</p>
							</div>
						</div>
						<div class="grid gap-3 md:grid-cols-3">
							<div class="rounded-md border border-border p-3 text-sm">
								<p class="text-muted-foreground">Wali kelas kandidat</p>
								<p class="mt-1 text-lg font-semibold">{rolloverPreview.counts.homeroom_assignments_to_copy}</p>
							</div>
							<div class="rounded-md border border-border p-3 text-sm">
								<p class="text-muted-foreground">Guru mapel kandidat</p>
								<p class="mt-1 text-lg font-semibold">{rolloverPreview.counts.subject_assignments_to_copy}</p>
							</div>
							<div class="rounded-md border border-border p-3 text-sm">
								<p class="text-muted-foreground">Slot jadwal kandidat</p>
								<p class="mt-1 text-lg font-semibold">{rolloverPreview.counts.timetable_slots_to_copy}</p>
							</div>
						</div>
						{#if rolloverPreview.warnings.length}
							<div class="space-y-2">
								{#each rolloverPreview.warnings as warning}
									<div class="flex gap-2 rounded-md border border-warning/30 bg-warning/10 px-3 py-2 text-sm">
										<AlertTriangle class="mt-0.5 size-4 shrink-0 text-warning" />
										<p>{warning}</p>
									</div>
								{/each}
							</div>
						{/if}
						{#if rolloverPreview.classes_to_create.length}
							<div class="overflow-x-auto rounded-md border border-border">
								<Table.Root>
									<Table.Header>
										<Table.Row>
											<Table.Head>Dari</Table.Head>
											<Table.Head>Target</Table.Head>
											<Table.Head>Tingkat</Table.Head>
										</Table.Row>
									</Table.Header>
									<Table.Body>
										{#each rolloverPreview.classes_to_create.slice(0, 8) as item}
											<Table.Row>
												<Table.Cell>{item.source_code} · {item.source_name}</Table.Cell>
												<Table.Cell class="font-medium">{item.target_code} · {item.target_name}</Table.Cell>
												<Table.Cell>{item.target_level}</Table.Cell>
											</Table.Row>
										{/each}
									</Table.Body>
								</Table.Root>
							</div>
						{/if}
						{#if rolloverPreview.students_without_next_class.length}
							<div class="overflow-x-auto rounded-md border border-border">
								<Table.Root>
									<Table.Header>
										<Table.Row>
											<Table.Head>Siswa</Table.Head>
											<Table.Head>Rombel</Table.Head>
											<Table.Head>Catatan</Table.Head>
										</Table.Row>
									</Table.Header>
									<Table.Body>
										{#each rolloverPreview.students_without_next_class.slice(0, 8) as item}
											<Table.Row>
												<Table.Cell>
													<p class="font-medium">{item.nama}</p>
													<p class="text-xs text-muted-foreground">NIS {item.nis || '-'}</p>
												</Table.Cell>
												<Table.Cell>{item.from_class_code}</Table.Cell>
												<Table.Cell class="text-muted-foreground">{item.reason}</Table.Cell>
											</Table.Row>
										{/each}
									</Table.Body>
								</Table.Root>
							</div>
						{/if}
					{:else}
						<EmptyStatePanel compact title="Belum ada preview" description="Pilih tahun tujuan lalu jalankan preview kenaikan." />
					{/if}
				</Card.Content>
			</Card.Root>
		</section>

		<section class="grid gap-4 xl:grid-cols-[24rem_1fr]" aria-labelledby="import-title">
			<Card.Root>
				<Card.Header class="pb-3">
					<Card.Title id="import-title" class="text-base">Import / Export CSV</Card.Title>
					<Card.Description>Template berbahasa Indonesia. Dry-run tidak mengubah data.</Card.Description>
				</Card.Header>
				<Card.Content class="space-y-3">
					<div>
						<label for="import-kind" class="mb-1 block text-sm font-medium">Jenis Data</label>
						<select id="import-kind" bind:value={importKind} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
							{#each importKinds as item (item.value)}
								<option value={item.value}>{item.label}</option>
							{/each}
						</select>
						<p class="mt-1 text-xs text-muted-foreground">{selectedImportKind.description}</p>
					</div>
					<Button variant="outline" class="w-full justify-start" href={templateHref(importKind)}>
						<Download class="mr-2 size-4" />
						Unduh Template CSV
					</Button>
					<div>
						<label for="import-file" class="mb-1 block text-sm font-medium">File CSV</label>
						<Input id="import-file" type="file" accept=".csv,text/csv" onchange={onImportFileChange} />
					</div>
					<LoadingButton class="w-full" loading={importBusy} loadingLabel="Validasi..." disabled={!importFile || importBusy} onclick={() => void dryRunImport()}>
						<Upload class="mr-2 size-4" />
						Dry-run Import
					</LoadingButton>
				</Card.Content>
			</Card.Root>

			<Card.Root>
				<Card.Header class="pb-3">
					<Card.Title class="text-base">Hasil Dry-run</Card.Title>
					<Card.Description>Baris valid hanya dihitung sebagai rencana tambah/ubah/lewati.</Card.Description>
				</Card.Header>
				<Card.Content class="space-y-4">
					{#if dryRunResult}
						<div class="grid gap-3 md:grid-cols-5">
							{#each [
								['Baris', dryRunResult.total_rows],
								['Tambah', dryRunResult.add_count],
								['Ubah', dryRunResult.update_count],
								['Lewati', dryRunResult.skip_count],
								['Error', dryRunResult.error_count]
							] as item}
								<div class="rounded-md border border-border p-3">
									<p class="text-xs text-muted-foreground">{item[0]}</p>
									<p class="mt-1 text-xl font-semibold">{item[1]}</p>
								</div>
							{/each}
						</div>
						{#if dryRunResult.row_errors.length}
							<div class="rounded-md border border-destructive/30 bg-destructive/10 p-3">
								<p class="text-sm font-medium text-destructive">Error baris</p>
								<div class="mt-2 space-y-1 text-sm">
									{#each dryRunResult.row_errors.slice(0, 10) as item}
										<p>Baris {item.row} · {item.field || 'Data'}: {item.message}</p>
									{/each}
								</div>
							</div>
						{/if}
						<div class="overflow-x-auto rounded-md border border-border">
							<Table.Root>
								<Table.Header>
									<Table.Row>
										<Table.Head>Baris</Table.Head>
										<Table.Head>Aksi</Table.Head>
										<Table.Head>Ringkasan</Table.Head>
									</Table.Row>
								</Table.Header>
								<Table.Body>
									{#each dryRunResult.rows.slice(0, 12) as item}
										<Table.Row>
											<Table.Cell>{item.row}</Table.Cell>
											<Table.Cell>
												<Badge class={actionBadgeClass(item.action)} variant="outline">{actionLabel(item.action)}</Badge>
											</Table.Cell>
											<Table.Cell class="text-muted-foreground">{item.summary}</Table.Cell>
										</Table.Row>
									{/each}
								</Table.Body>
							</Table.Root>
						</div>
					{:else}
						<EmptyStatePanel compact title="Belum ada dry-run" description="Unggah CSV dari template lalu jalankan dry-run untuk melihat tambah, ubah, skip, dan error." />
					{/if}
				</Card.Content>
			</Card.Root>
		</section>
	</AsyncContent>
</div>
