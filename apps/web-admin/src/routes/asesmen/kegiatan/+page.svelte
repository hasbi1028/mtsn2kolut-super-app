<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import * as Card from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Badge } from '$lib/components/ui/badge';
	import { toast } from '$lib/components/ui/sonner';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import MicroActionTable from '$lib/components/ops/MicroActionTable.svelte';
	import { confirmAction } from '$lib/confirm-dialog';
	import { clientApiPath, readClientApiData, readClientJson } from '$lib/client/api';

	type CbtEvent = {
		id: string; title: string; exam_type: string; scope: string;
		target_levels: string[];
		academic_year_id: string; academic_year_name: string;
		status: string; created_at: string; session_count: number;
	};
	type AcademicYear = { id: string; name: string; is_active: boolean; };
	type EventQuestionReadiness = {
		complete_rows: number;
		total_rows: number;
		incomplete_rows: number;
		missing_pg: number;
		missing_essay: number;
	};
	type EventsOverview = {
		events: CbtEvent[];
		years: AcademicYear[];
		questionReadiness: Record<string, EventQuestionReadiness>;
	};
	type StatusFilter = 'all' | 'draft' | 'active' | 'finished';

	let events = $state<CbtEvent[]>([]);
	let years = $state<AcademicYear[]>([]);
	let eventsPromise = $state<Promise<EventsOverview> | null>(null);
	let showForm = $state(false);
	let editId = $state<string | null>(null);
	let deleteBusyId = $state('');

	let fTitle = $state('');
	let fType = $state('uts');
	let fScope = $state('grade');
	let fYearId = $state('');
	let fStatus = $state('draft');
	let originalStatus = $state('draft');
	let fTargetLevels = $state<string[]>([]);
	let fBusy = $state(false);
	let statusFilter = $state<StatusFilter>('all');
	let eventsRequestId = 0;
	const gradeOptions = ['VII', 'VIII', 'IX'];
	const statusFilters: Array<{ value: StatusFilter; label: string }> = [
		{ value: 'all', label: 'Semua' },
		{ value: 'active', label: 'Aktif' },
		{ value: 'draft', label: 'Konsep' },
		{ value: 'finished', label: 'Selesai' },
	];
	const eventColumns = [
		{ key: 'event', label: 'Kegiatan', class: 'min-w-[16rem]' },
		{ key: 'type', label: 'Tipe' },
		{ key: 'scope', label: 'Cakupan' },
		{ key: 'sessions', label: 'Sesi', class: 'text-center' },
		{ key: 'status', label: 'Status' },
	];

	const typeLabel: Record<string, string> = {
		ulangan: 'Ulangan', uts: 'UTS', uas: 'UAS', uam: 'UAM', tryout: 'Try Out', lainnya: 'Lainnya'
	};
	const scopeLabel: Record<string, string> = { class: 'Per Kelas', grade: 'Per Tingkat', school: 'Seluruh Sekolah' };
	const statusLabel: Record<string, string> = { draft: 'Konsep', active: 'Aktif', finished: 'Selesai' };
	const eventHomeCopy = 'Daftar ringkas kegiatan ujian. Buka satu kegiatan untuk mengikuti alur kesiapan paket, sesi, pemantauan, dan hasil.';

	function statusClass(status: string) {
		if (status === 'active') return 'bg-primary/15 text-primary border-primary/20';
		if (status === 'finished') return 'bg-muted text-muted-foreground border-border';
		return 'bg-warning/15 text-warning border-warning/30';
	}

	function eventProgressLabel(event: CbtEvent, readiness?: EventQuestionReadiness) {
		if (readiness && readiness.total_rows > 0) {
			const pct = Math.round((readiness.complete_rows / readiness.total_rows) * 100);
			return readiness.incomplete_rows > 0 ? `Soal ${pct}% lengkap · ${readiness.incomplete_rows} belum lengkap` : 'Soal 100% lengkap';
		}
		if (event.session_count > 0) return `${event.session_count} sesi tersusun`;
		if (event.status === 'draft') return 'Siapkan paket dan sesi';
		return 'Belum ada sesi';
	}

	function questionReadinessClass(readiness?: EventQuestionReadiness) {
		if (!readiness || readiness.total_rows === 0) return 'bg-muted text-muted-foreground border-border';
		if (readiness.incomplete_rows === 0) return 'bg-success/10 text-success border-success/20';
		return 'bg-warning/10 text-warning border-warning/30';
	}

	function questionReadinessLabel(readiness?: EventQuestionReadiness) {
		if (!readiness || readiness.total_rows === 0) return 'Soal belum dihitung';
		const pct = Math.round((readiness.complete_rows / readiness.total_rows) * 100);
		return readiness.incomplete_rows > 0 ? `Soal ${pct}%` : 'Soal lengkap';
	}

	function statusCount(items: CbtEvent[], status: StatusFilter) {
		if (status === 'all') return items.length;
		return items.filter((event) => event.status === status).length;
	}

	function filteredEvents(items: CbtEvent[]) {
		if (statusFilter === 'all') return items;
		return items.filter((event) => event.status === statusFilter);
	}

	function isRecord(value: unknown): value is Record<string, unknown> {
		return typeof value === 'object' && value !== null;
	}

	function parseAcademicYears(payload: unknown) {
		if (!isRecord(payload)) return [];
		if (Array.isArray(payload.years)) return payload.years as AcademicYear[];
		return [];
	}

	async function fetchEventQuestionReadiness(event: CbtEvent): Promise<[string, EventQuestionReadiness] | null> {
		try {
			const payload = await fetch(clientApiPath`/api/asesmen/events/${event.id}/question-completeness`).then((response) => readClientApiData<unknown>(response));
			const summary = isRecord(payload) && isRecord(payload.summary) ? payload.summary : null;
			if (!summary) return null;
			return [event.id, {
				complete_rows: Number(summary.complete_rows) || 0,
				total_rows: Number(summary.total_rows) || 0,
				incomplete_rows: Number(summary.incomplete_rows) || 0,
				missing_pg: Number(summary.missing_pg) || 0,
				missing_essay: Number(summary.missing_essay) || 0,
			}];
		} catch {
			return null;
		}
	}

	async function fetchOverview(): Promise<EventsOverview> {
		const [eventsData, academicData] = await Promise.all([
			fetch('/api/asesmen/events').then((response) => readClientApiData<CbtEvent[]>(response, 'Gagal memuat data kegiatan ujian')),
			fetch('/api/academic').then((response) => readClientApiData<unknown>(response, 'Gagal memuat data akademik')),
		]);
		const parsedEvents = Array.isArray(eventsData) ? eventsData : [];
		const readinessEntries = (await Promise.all(parsedEvents.slice(0, 12).map((event) => fetchEventQuestionReadiness(event)))).filter((entry): entry is [string, EventQuestionReadiness] => Boolean(entry));
		return {
			events: parsedEvents,
			years: parseAcademicYears(academicData),
			questionReadiness: Object.fromEntries(readinessEntries),
		};
	}

	function applyOverview(overview: EventsOverview) {
		events = overview.events;
		years = overview.years;
		if (years.length > 0 && !fYearId) {
			fYearId = years.find((year) => year.is_active)?.id || years[0].id;
		}
	}

	function loadInitial() {
		const requestId = ++eventsRequestId;
		events = [];
		years = [];
		eventsPromise = fetchOverview().then((overview) => {
			if (requestId !== eventsRequestId) return { events, years, questionReadiness: {} };
			applyOverview(overview);
			return overview;
		}).catch((error: unknown) => {
			if (requestId === eventsRequestId) throw error;
			return { events, years, questionReadiness: {} };
		});
	}

	async function refreshOverview() {
		if (!eventsPromise) {
			loadInitial();
			return;
		}
		const requestId = ++eventsRequestId;
		try {
			const overview = await fetchOverview();
			if (requestId !== eventsRequestId) return;
			applyOverview(overview);
			eventsPromise = Promise.resolve(overview);
		} catch (error) {
			if (requestId === eventsRequestId) {
				eventsPromise = Promise.resolve({ events, years, questionReadiness: {} });
				toast.error(overviewErrorMessage(error));
			}
		}
	}

	function retryOverview(reset?: () => void) {
		reset?.();
		loadInitial();
	}

	function overviewErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Gagal memuat data kegiatan ujian';
	}

	function handleOverviewRenderError(error: unknown) {
		console.error('Daftar kegiatan asesmen belum dapat ditampilkan', error);
	}

	function showToast(msg: string) {
		toast.success(msg);
	}

	function showError(msg: string) {
		toast.error(msg);
	}

	function mutationErrorMessage(error: unknown, fallback: string) {
		if (error instanceof Error && error.message.trim() && !error.message.toLowerCase().includes('fetch')) return error.message;
		return fallback;
	}

	function resetForm() {
		fTitle = ''; fType = 'uts'; fScope = 'grade'; fStatus = 'draft';
		originalStatus = 'draft';
		fTargetLevels = [];
		editId = null; showForm = false;
	}

	function openEdit(e: CbtEvent) {
		fTitle = e.title;
		fType = e.exam_type;
		fScope = e.scope;
		fYearId = e.academic_year_id;
		fStatus = e.status;
		originalStatus = e.status;
		fTargetLevels = [...(e.target_levels ?? [])];
		editId = e.id;
		showForm = true;
	}

	async function updateEventStatusIfNeeded(id: string) {
		if (fStatus === originalStatus) return;
		const res = await fetch(clientApiPath`/api/asesmen/events/${id}/status`, {
			method: 'PATCH',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ status: fStatus }),
		});
		await readClientJson<unknown>(res);
	}

	function toggleTargetLevel(level: string, checked: boolean) {
		if (checked) {
			fTargetLevels = Array.from(new Set([...fTargetLevels, level])).sort();
			return;
		}
		fTargetLevels = fTargetLevels.filter((item) => item !== level);
	}

	async function saveEvent() {
		if (!fTitle || !fType || !fYearId) return;
		fBusy = true;
		const currentEditId = editId;
		let fieldUpdateSucceeded = false;
		try {
			const method = currentEditId ? 'PUT' : 'POST';
			const path = currentEditId ? clientApiPath`/api/asesmen/events/${currentEditId}` : '/api/asesmen/events';
			const body = currentEditId
				? {
					title: fTitle, exam_type: fType, scope: fScope,
					target_levels: fTargetLevels,
					academic_year_id: fYearId,
				}
				: {
					title: fTitle, exam_type: fType, scope: fScope,
					target_levels: fTargetLevels,
					academic_year_id: fYearId, status: fStatus,
				};
			const res = await fetch(path, {
				method,
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(body),
			});
			await readClientJson<unknown>(res);
			fieldUpdateSucceeded = true;
			if (currentEditId) await updateEventStatusIfNeeded(currentEditId);
			showToast(currentEditId ? 'Kegiatan diperbarui' : 'Kegiatan berhasil dibuat');
			resetForm();
			await refreshOverview();
		} catch (error) {
			showError(mutationErrorMessage(error, 'Gagal menyimpan kegiatan ujian. Periksa koneksi lalu coba lagi.'));
			if (currentEditId && fieldUpdateSucceeded) await refreshOverview();
		} finally { fBusy = false; }
	}

	async function deleteEvent(id: string) {
		if (!(await confirmAction({
			title: 'Hapus Kegiatan Ujian',
			message: 'Hapus kegiatan ini? Sesi di dalamnya tidak akan terhapus, tetapi relasinya akan dilepas.',
			confirmLabel: 'Hapus Kegiatan',
			tone: 'danger'
		}))) return;
		deleteBusyId = id;
		try {
			const res = await fetch(clientApiPath`/api/asesmen/events/${id}`, { method: 'DELETE' });
			await readClientJson<unknown>(res);
			showToast('Kegiatan dihapus');
			await refreshOverview();
		} catch (error) {
			showError(mutationErrorMessage(error, 'Gagal menghapus kegiatan ujian. Periksa koneksi lalu coba lagi.'));
		} finally {
			deleteBusyId = '';
		}
	}

	onMount(() => {
		void loadInitial();
	});
</script>

<svelte:head><title>Kegiatan & Sesi Ujian — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-5">
	<section class="overflow-hidden rounded-2xl border border-border bg-card shadow-sm">
		<div class="grid gap-4 p-5 lg:grid-cols-[minmax(0,1fr)_auto] lg:items-center">
			<div class="min-w-0">
				<p class="text-xs font-bold uppercase tracking-[0.18em] text-primary">Kegiatan & Sesi Ujian</p>
				<h1 class="mt-1 text-2xl font-semibold tracking-tight text-foreground">Kegiatan & Sesi Ujian</h1>
				<p class="mt-2 max-w-3xl text-sm leading-6 text-muted-foreground">{eventHomeCopy}</p>
			</div>
			<div class="flex flex-wrap gap-2 lg:justify-end">
				<a href={resolve('/asesmen')} class="inline-flex rounded-md border border-input bg-card px-3 py-2 text-sm font-medium text-foreground hover:bg-muted/50">Beranda Ujian</a>
				<a href={resolve('/asesmen/kegiatan/new')} class="inline-flex rounded-md bg-primary px-3 py-2 text-sm font-semibold text-primary-foreground hover:bg-primary/90">Buat Kegiatan</a>
			</div>
		</div>
	</section>

	{#if showForm}
		<Card.Root class="border-border shadow-sm">
			<Card.Header class="pb-2">
				<Card.Title class="text-base">{editId ? 'Edit Kegiatan' : 'Buat Kegiatan'}</Card.Title>
				<Card.Description>Isi identitas kegiatan sekali, lalu lanjutkan ke alur kesiapan untuk paket, sesi, pemantauan, dan hasil.</Card.Description>
			</Card.Header>
			<Card.Content class="space-y-4">
				<div class="grid gap-3 sm:grid-cols-2">
					<div class="sm:col-span-2">
						<label for="e-title" class="text-xs text-muted-foreground mb-1 block">Judul Kegiatan <span class="text-destructive">*</span></label>
						<Input id="e-title" placeholder="mis: UTS Semester Ganjil 2025/2026" bind:value={fTitle} />
					</div>
					<div>
						<label for="e-year" class="text-xs text-muted-foreground mb-1 block">Tahun Ajaran <span class="text-destructive">*</span></label>
						<select id="e-year" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fYearId}>
							{#each years as y (y.id)}
								<option value={y.id}>{y.name} {y.is_active ? '(Aktif)' : ''}</option>
							{/each}
						</select>
					</div>
					<div>
						<label for="e-type" class="text-xs text-muted-foreground mb-1 block">Jenis Ujian</label>
						<select id="e-type" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fType}>
							{#each Object.entries(typeLabel) as [val, label] (val)}
								<option value={val}>{label}</option>
							{/each}
						</select>
					</div>
					<div>
						<label for="e-scope" class="text-xs text-muted-foreground mb-1 block">Cakupan</label>
						<select id="e-scope" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fScope}>
							{#each Object.entries(scopeLabel) as [val, label] (val)}
								<option value={val}>{label}</option>
							{/each}
						</select>
					</div>
					<fieldset class="sm:col-span-2">
						<legend class="mb-2 block text-xs text-muted-foreground">Tingkat yang diikutkan</legend>
						<div class="grid gap-2 sm:grid-cols-3">
							{#each gradeOptions as level (level)}
								<label class="flex items-center gap-2 rounded-md border border-input px-3 py-2 text-sm text-foreground">
									<input
										type="checkbox"
										checked={fTargetLevels.includes(level)}
										onchange={(event) => toggleTargetLevel(level, (event.currentTarget as HTMLInputElement).checked)}
										class="size-4 accent-primary"
									/>
									<span>Tingkat {level}</span>
								</label>
							{/each}
						</div>
						<p class="mt-2 text-xs text-muted-foreground">
							Kosong berarti mengikuti cakupan biasa. Isi ini untuk kasus seperti UAS genap yang hanya berlaku bagi tingkat tertentu.
						</p>
					</fieldset>
					<div>
						<label for="e-status" class="text-xs text-muted-foreground mb-1 block">Status</label>
						<select id="e-status" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fStatus}>
							{#each Object.entries(statusLabel) as [val, label] (val)}
								<option value={val}>{label}</option>
							{/each}
						</select>
					</div>
				</div>

				<div class="flex gap-2">
					<LoadingButton disabled={fBusy || !fTitle || !fYearId} onclick={() => void saveEvent()} loading={fBusy} loadingLabel="Menyimpan...">
						{editId ? 'Perbarui' : 'Simpan Kegiatan'}
					</LoadingButton>
					<LoadingButton variant="outline" onclick={resetForm}>Batal</LoadingButton>
				</div>
			</Card.Content>
		</Card.Root>
	{/if}

	<AsyncContent promise={eventsPromise} onerror={handleOverviewRenderError}>
		{#snippet pending()}
			<Card.Root class="overflow-hidden border-border shadow-sm">
				<Card.Content class="space-y-3 p-6">
					{#each Array.from({ length: 5 }) as _, index (`event-row-skeleton-${index}`)}
						<div class="grid gap-3 lg:grid-cols-[1.2fr_0.7fr_0.8fr_0.5fr_0.6fr_auto] lg:items-center">
							<Skeleton class="h-5 w-40" />
							<Skeleton class="h-6 w-20" />
							<Skeleton class="h-5 w-20" />
							<Skeleton class="h-6 w-16" />
							<Skeleton class="h-6 w-16" />
							<Skeleton class="h-9 w-28 justify-self-end" />
						</div>
					{/each}
				</Card.Content>
			</Card.Root>
		{/snippet}

		{#snippet failed(error, reset)}
			<RecoveryPanel
				title="Kegiatan Ujian Belum Tersaji"
				message={overviewErrorMessage(error)}
				onRetry={() => retryOverview(reset)}
			/>
		{/snippet}

		{#snippet children(value)}
			{@const overview = value as EventsOverview}
		{@const visibleEvents = filteredEvents(overview.events)}
		<section class="rounded-2xl border border-border bg-card p-4 shadow-sm">
			<div class="flex flex-wrap items-center justify-between gap-3">
				<div>
					<p class="text-sm font-semibold text-foreground">Kegiatan & Sesi</p>
					<p class="mt-1 text-xs text-muted-foreground">{overview.events.length} kegiatan, {overview.events.reduce((sum, event) => sum + (event.session_count || 0), 0)} sesi tersusun</p>
				</div>
				<div class="flex flex-wrap gap-2" aria-label="Filter status kegiatan">
					{#each statusFilters as filter (filter.value)}
						<button
							type="button"
							class={`rounded-full border px-3 py-1.5 text-xs font-semibold transition ${statusFilter === filter.value ? 'border-primary bg-primary/10 text-primary' : 'border-border bg-card text-muted-foreground hover:bg-muted/50'}`}
							aria-pressed={statusFilter === filter.value}
							onclick={() => statusFilter = filter.value}
						>
							{filter.label} <span class="ml-1 font-bold">{statusCount(overview.events, filter.value)}</span>
						</button>
					{/each}
				</div>
			</div>
		</section>
		<MicroActionTable
			title="Daftar kegiatan"
			description="Tampilan ringkas untuk memindai status, sesi, kesiapan soal, dan aksi pengelolaan."
			columns={eventColumns}
			rows={visibleEvents}
			rowKey={(row) => (row as CbtEvent).id}
			tableClass="min-w-[860px]"
			emptyTitle={overview.events.length === 0 ? 'Belum ada kegiatan ujian.' : 'Tidak ada kegiatan pada filter ini.'}
		>
			{#snippet cell(row, column)}
				{@const e = row as CbtEvent}
				{#if column.key === 'event'}
					<div class="font-medium text-foreground">{e.title}</div>
					<div class="text-xs text-muted-foreground">{e.academic_year_name} · {e.target_levels?.length ? `Tingkat ${e.target_levels.join(', ')}` : 'Target mengikuti cakupan'}</div>
				{:else if column.key === 'type'}
					<Badge variant="outline" class="text-xs capitalize">{typeLabel[e.exam_type] ?? e.exam_type}</Badge>
				{:else if column.key === 'scope'}
					<span class="text-sm text-muted-foreground">{scopeLabel[e.scope] ?? e.scope}</span>
				{:else if column.key === 'sessions'}
					<div class="flex flex-col items-center gap-1">
						<Badge variant="secondary">{e.session_count} Sesi</Badge>
						<Badge class={questionReadinessClass(overview.questionReadiness[e.id])}>{questionReadinessLabel(overview.questionReadiness[e.id])}</Badge>
					</div>
				{:else}
					<Badge class={statusClass(e.status)}>{statusLabel[e.status] ?? e.status}</Badge>
					<div class="mt-1 text-xs text-muted-foreground">{eventProgressLabel(e, overview.questionReadiness[e.id])}</div>
				{/if}
			{/snippet}
			{#snippet actions(row)}
				{@const e = row as CbtEvent}
				<a href={resolve(`/asesmen/kegiatan/${e.id}`)} class="inline-flex h-8 items-center rounded-md bg-primary px-3 text-xs font-semibold text-primary-foreground hover:bg-primary/90">Kelola</a>
				<LoadingButton variant="ghost" size="sm" class="text-muted-foreground" onclick={() => openEdit(e)}>Edit</LoadingButton>
				<LoadingButton
					variant="ghost"
					size="sm"
					class="text-destructive hover:bg-destructive/10 hover:text-destructive"
					onclick={() => deleteEvent(e.id)}
					loading={deleteBusyId === e.id}
					loadingLabel="Menghapus..."
					disabled={deleteBusyId !== '' && deleteBusyId !== e.id}
				>Hapus</LoadingButton>
			{/snippet}
			{#snippet mobile(row)}
				{@const e = row as CbtEvent}
				<div class="space-y-3">
					<div class="flex items-start justify-between gap-3">
						<div class="min-w-0">
							<p class="text-sm font-semibold text-foreground">{e.title}</p>
							<p class="mt-1 text-xs text-muted-foreground">{e.academic_year_name}</p>
						</div>
						<Badge class={statusClass(e.status)}>{statusLabel[e.status] ?? e.status}</Badge>
					</div>
					<div class="flex flex-wrap items-center gap-2">
						<Badge variant="outline" class="text-xs capitalize">{typeLabel[e.exam_type] ?? e.exam_type}</Badge>
						<Badge variant="outline" class="text-xs">{scopeLabel[e.scope] ?? e.scope}</Badge>
						<Badge variant="secondary">{e.session_count} sesi</Badge>
						<Badge class={questionReadinessClass(overview.questionReadiness[e.id])}>{questionReadinessLabel(overview.questionReadiness[e.id])}</Badge>
					</div>
					<p class="text-xs text-muted-foreground">{eventProgressLabel(e, overview.questionReadiness[e.id])}</p>
					<div class="flex flex-wrap gap-2">
						<a href={resolve(`/asesmen/kegiatan/${e.id}`)} class="inline-flex h-8 items-center rounded-md bg-primary px-3 text-xs font-semibold text-primary-foreground hover:bg-primary/90">Kelola</a>
						<LoadingButton variant="ghost" size="sm" onclick={() => openEdit(e)}>Edit</LoadingButton>
						<LoadingButton variant="ghost" size="sm" class="text-destructive hover:bg-destructive/10 hover:text-destructive" onclick={() => deleteEvent(e.id)} loading={deleteBusyId === e.id} loadingLabel="Menghapus..." disabled={deleteBusyId !== '' && deleteBusyId !== e.id}>Hapus</LoadingButton>
					</div>
				</div>
			{/snippet}
		</MicroActionTable>
		{/snippet}
	</AsyncContent>
</div>
