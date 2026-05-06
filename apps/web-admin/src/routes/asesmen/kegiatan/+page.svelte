<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Input } from '$lib/components/ui/input';
	import { Badge } from '$lib/components/ui/badge';
	import { toast } from '$lib/components/ui/sonner';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { confirmAction } from '$lib/confirm-dialog';
	import { clientApiPath, readClientApiData, readClientJson } from '$lib/client/api';

	type CbtEvent = {
		id: string; title: string; exam_type: string; scope: string;
		target_levels: string[];
		academic_year_id: string; academic_year_name: string;
		status: string; created_at: string; session_count: number;
	};
	type AcademicYear = { id: string; name: string; is_active: boolean; };
	type EventsOverview = {
		events: CbtEvent[];
		years: AcademicYear[];
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
		{ value: 'draft', label: 'Draft' },
		{ value: 'finished', label: 'Selesai' },
	];

	const typeLabel: Record<string, string> = {
		ulangan: 'Ulangan', uts: 'UTS', uas: 'UAS', uam: 'UAM', tryout: 'Try Out', lainnya: 'Lainnya'
	};
	const scopeLabel: Record<string, string> = { class: 'Per Kelas', grade: 'Per Tingkat', school: 'Seluruh Sekolah' };
	const statusLabel: Record<string, string> = { draft: 'Draft', active: 'Aktif', finished: 'Selesai' };
	const eventHomeCopy = 'Daftar ringkas kegiatan CBT. Buka satu kegiatan untuk mengikuti wizard kesiapan paket, sesi, monitoring, dan hasil.';

	function statusClass(status: string) {
		if (status === 'active') return 'bg-emerald-100 text-emerald-700 border-emerald-200';
		if (status === 'finished') return 'bg-slate-100 text-slate-600 border-slate-200';
		return 'bg-amber-100 text-amber-700 border-amber-200';
	}

	function eventProgressLabel(event: CbtEvent) {
		if (event.session_count > 0) return `${event.session_count} sesi tersusun`;
		if (event.status === 'draft') return 'Siapkan paket dan sesi';
		return 'Belum ada sesi';
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

	async function fetchOverview(): Promise<EventsOverview> {
		const [eventsData, academicData] = await Promise.all([
			fetch('/api/asesmen/events').then((response) => readClientApiData<CbtEvent[]>(response, 'Gagal memuat data kegiatan ujian')),
			fetch('/api/academic').then((response) => readClientApiData<unknown>(response, 'Gagal memuat data akademik')),
		]);
		return {
			events: Array.isArray(eventsData) ? eventsData : [],
			years: parseAcademicYears(academicData),
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
			if (requestId !== eventsRequestId) return { events, years };
			applyOverview(overview);
			return overview;
		}).catch((error: unknown) => {
			if (requestId === eventsRequestId) throw error;
			return { events, years };
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
				eventsPromise = Promise.resolve({ events, years });
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
		console.error('CBT events render failed', error);
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

<svelte:head><title>Kegiatan & Sesi CBT — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-5">
	<section class="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm">
		<div class="grid gap-4 p-5 lg:grid-cols-[minmax(0,1fr)_auto] lg:items-center">
			<div class="min-w-0">
				<p class="text-xs font-bold uppercase tracking-[0.18em] text-emerald-700">Kegiatan & Sesi CBT</p>
				<h1 class="mt-1 text-2xl font-semibold tracking-tight text-slate-900">Kegiatan & Sesi CBT</h1>
				<p class="mt-2 max-w-3xl text-sm leading-6 text-slate-600">{eventHomeCopy}</p>
			</div>
			<div class="flex flex-wrap gap-2 lg:justify-end">
				<a href={resolve('/asesmen')} class="inline-flex rounded-md border border-input bg-white px-3 py-2 text-sm font-medium text-slate-700 hover:bg-slate-50">Beranda CBT</a>
				<a href={resolve('/asesmen/kegiatan/new')} class="inline-flex rounded-md bg-[oklch(0.38_0.13_145)] px-3 py-2 text-sm font-semibold text-white hover:bg-[oklch(0.34_0.13_145)]">Buat Kegiatan</a>
			</div>
		</div>
	</section>

	{#if showForm}
		<Card.Root class="border-slate-200 shadow-sm">
			<Card.Header class="pb-2">
				<Card.Title class="text-base">{editId ? 'Edit Kegiatan' : 'Buat Kegiatan'}</Card.Title>
				<Card.Description>Isi identitas kegiatan sekali, lalu lanjutkan ke wizard kesiapan untuk paket, sesi, monitoring, dan hasil.</Card.Description>
			</Card.Header>
			<Card.Content class="space-y-4">
				<div class="grid gap-3 sm:grid-cols-2">
					<div class="sm:col-span-2">
						<label for="e-title" class="text-xs text-slate-500 mb-1 block">Judul Kegiatan <span class="text-red-500">*</span></label>
						<Input id="e-title" placeholder="mis: UTS Semester Ganjil 2025/2026" bind:value={fTitle} />
					</div>
					<div>
						<label for="e-year" class="text-xs text-slate-500 mb-1 block">Tahun Ajaran <span class="text-red-500">*</span></label>
						<select id="e-year" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fYearId}>
							{#each years as y (y.id)}
								<option value={y.id}>{y.name} {y.is_active ? '(Aktif)' : ''}</option>
							{/each}
						</select>
					</div>
					<div>
						<label for="e-type" class="text-xs text-slate-500 mb-1 block">Jenis Ujian</label>
						<select id="e-type" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fType}>
							{#each Object.entries(typeLabel) as [val, label] (val)}
								<option value={val}>{label}</option>
							{/each}
						</select>
					</div>
					<div>
						<label for="e-scope" class="text-xs text-slate-500 mb-1 block">Cakupan</label>
						<select id="e-scope" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fScope}>
							{#each Object.entries(scopeLabel) as [val, label] (val)}
								<option value={val}>{label}</option>
							{/each}
						</select>
					</div>
					<fieldset class="sm:col-span-2">
						<legend class="mb-2 block text-xs text-slate-500">Tingkat yang diikutkan</legend>
						<div class="grid gap-2 sm:grid-cols-3">
							{#each gradeOptions as level (level)}
								<label class="flex items-center gap-2 rounded-md border border-input px-3 py-2 text-sm text-slate-700">
									<input
										type="checkbox"
										checked={fTargetLevels.includes(level)}
										onchange={(event) => toggleTargetLevel(level, (event.currentTarget as HTMLInputElement).checked)}
										class="size-4 accent-emerald-700"
									/>
									<span>Tingkat {level}</span>
								</label>
							{/each}
						</div>
						<p class="mt-2 text-xs text-slate-500">
							Kosong berarti mengikuti cakupan biasa. Isi ini untuk kasus seperti UAS genap yang hanya berlaku bagi tingkat tertentu.
						</p>
					</fieldset>
					<div>
						<label for="e-status" class="text-xs text-slate-500 mb-1 block">Status</label>
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
			<Card.Root class="overflow-hidden border-slate-200 shadow-sm">
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
		<section class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
			<div class="flex flex-wrap items-center justify-between gap-3">
				<div>
					<p class="text-sm font-semibold text-slate-900">Kegiatan & Sesi</p>
					<p class="mt-1 text-xs text-slate-500">{overview.events.length} kegiatan, {overview.events.reduce((sum, event) => sum + (event.session_count || 0), 0)} sesi tersusun</p>
				</div>
				<div class="flex flex-wrap gap-2" aria-label="Filter status kegiatan">
					{#each statusFilters as filter (filter.value)}
						<button
							type="button"
							class={`rounded-full border px-3 py-1.5 text-xs font-semibold transition ${statusFilter === filter.value ? 'border-emerald-700 bg-emerald-50 text-emerald-800' : 'border-slate-200 bg-white text-slate-600 hover:bg-slate-50'}`}
							aria-pressed={statusFilter === filter.value}
							onclick={() => statusFilter = filter.value}
						>
							{filter.label} <span class="ml-1 font-bold">{statusCount(overview.events, filter.value)}</span>
						</button>
					{/each}
				</div>
			</div>
		</section>
		<Card.Root class="overflow-hidden border-slate-200 shadow-sm">
			<Card.Content class="p-0 overflow-x-auto">
				<div class="hidden overflow-x-auto lg:block">
					<Table.Root>
						<Table.Header>
							<Table.Row>
								<Table.Head>Judul Kegiatan</Table.Head>
								<Table.Head>Tipe</Table.Head>
								<Table.Head>Cakupan</Table.Head>
								<Table.Head class="text-center">Sesi</Table.Head>
								<Table.Head>Status</Table.Head>
								<Table.Head class="text-right">Aksi</Table.Head>
							</Table.Row>
						</Table.Header>
						<Table.Body>
							{#each visibleEvents as e (e.id)}
								<Table.Row>
									<Table.Cell>
										<div class="font-medium text-slate-800">{e.title}</div>
										<div class="text-xs text-slate-500">{e.academic_year_name} · {e.target_levels?.length ? `Tingkat ${e.target_levels.join(', ')}` : 'Target mengikuti cakupan'}</div>
									</Table.Cell>
									<Table.Cell>
										<Badge variant="outline" class="text-xs capitalize">{typeLabel[e.exam_type] ?? e.exam_type}</Badge>
									</Table.Cell>
									<Table.Cell class="text-sm text-slate-600">{scopeLabel[e.scope] ?? e.scope}</Table.Cell>
									<Table.Cell class="text-center">
										<Badge variant="secondary">{e.session_count} Sesi</Badge>
									</Table.Cell>
									<Table.Cell>
										<Badge class={statusClass(e.status)}>
											{statusLabel[e.status] ?? e.status}
										</Badge>
										<div class="mt-1 text-xs text-slate-500">{eventProgressLabel(e)}</div>
									</Table.Cell>
									<Table.Cell class="text-right">
										<div class="flex flex-wrap justify-end gap-1.5">
											<a href={resolve(`/asesmen/kegiatan/${e.id}`)} class="inline-flex items-center rounded-md bg-[oklch(0.38_0.13_145)] px-3 py-1.5 text-sm font-semibold text-white hover:bg-[oklch(0.34_0.13_145)]">Kelola</a>
											<LoadingButton variant="ghost" size="sm" class="text-slate-500" onclick={() => openEdit(e)}>Edit</LoadingButton>
											<LoadingButton
												variant="ghost"
												size="sm"
												class="text-red-600 hover:bg-red-50 hover:text-red-700"
												onclick={() => deleteEvent(e.id)}
												loading={deleteBusyId === e.id}
												loadingLabel="Menghapus..."
												disabled={deleteBusyId !== '' && deleteBusyId !== e.id}
											>Hapus</LoadingButton>
										</div>
									</Table.Cell>
								</Table.Row>
							{:else}
								<Table.Row>
									<Table.Cell colspan={6} class="text-center text-slate-400 py-12">
										{overview.events.length === 0 ? 'Belum ada kegiatan ujian.' : 'Tidak ada kegiatan pada filter ini.'}
									</Table.Cell>
								</Table.Row>
							{/each}
						</Table.Body>
					</Table.Root>
				</div>

				<div class="grid gap-3 p-4 lg:hidden">
					{#each visibleEvents as e (e.id)}
						<div class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
							<div class="flex items-start justify-between gap-3">
								<div class="min-w-0">
									<p class="text-sm font-semibold text-slate-900">{e.title}</p>
									<p class="mt-1 text-xs text-slate-500">{e.academic_year_name}</p>
								</div>
								<Badge class={statusClass(e.status)}>
									{statusLabel[e.status] ?? e.status}
								</Badge>
							</div>
							<div class="mt-3 flex flex-wrap items-center gap-2">
								<Badge variant="outline" class="text-xs capitalize">{typeLabel[e.exam_type] ?? e.exam_type}</Badge>
								<Badge variant="outline" class="text-xs">{scopeLabel[e.scope] ?? e.scope}</Badge>
								{#if e.target_levels?.length}
									<Badge variant="outline" class="text-xs">{e.target_levels.join(', ')}</Badge>
								{/if}
								<Badge variant="secondary">{e.session_count} sesi</Badge>
							</div>
							<p class="mt-3 text-xs text-slate-500">{eventProgressLabel(e)}</p>
							<div class="mt-4 grid grid-cols-2 gap-2">
								<a href={resolve(`/asesmen/kegiatan/${e.id}`)} class="col-span-2 inline-flex items-center justify-center rounded-md bg-[oklch(0.38_0.13_145)] px-3 py-2 text-sm font-semibold text-white hover:bg-[oklch(0.34_0.13_145)]">Kelola</a>
								<LoadingButton variant="ghost" size="sm" onclick={() => openEdit(e)}>Edit</LoadingButton>
								<LoadingButton
									variant="ghost"
									size="sm"
									class="text-red-600 hover:bg-red-50 hover:text-red-700"
									onclick={() => deleteEvent(e.id)}
									loading={deleteBusyId === e.id}
									loadingLabel="Menghapus..."
									disabled={deleteBusyId !== '' && deleteBusyId !== e.id}
								>Hapus</LoadingButton>
							</div>
						</div>
					{:else}
						<div class="rounded-2xl border border-dashed border-slate-300 bg-slate-50 px-4 py-10 text-center text-sm text-slate-500">
							{overview.events.length === 0 ? 'Belum ada kegiatan ujian.' : 'Tidak ada kegiatan pada filter ini.'}
						</div>
					{/each}
				</div>
			</Card.Content>
		</Card.Root>
		{/snippet}
	</AsyncContent>
</div>
