<script lang="ts">
	import { onMount } from 'svelte';
	import {
		groupSesiCbtMatrix,
		normalizeSesiCbtRow,
		sessionStatusLabel,
		sessionStatusTone,
		summarizeSessionRows,
		type SesiCbtRow,
		type SesiCbtStatus
	} from '$lib/asesmen/sesi-cbt';

	type SessionsResponse = {
		items?: unknown[];
		summary?: ReturnType<typeof summarizeSessionRows>;
		generated_at?: string;
	};
	type ExamOption = { id: string; title?: string; nama?: string };
	type PackageMapOption = {
		id?: string;
		local_id?: string;
		class_id: string;
		class_code?: string;
		class_name?: string;
		subject_id: string;
		subject_name?: string;
		package_id: string;
		package_title?: string;
		duration_minutes?: number;
	};
	type SessionDraft = { examId: string; packageMapKey: string; date: string; startTime: string; durationMinutes: number; title: string };

	let rows = $state<SesiCbtRow[]>([]);
	let loading = $state(true);
	let error = $state('');
	let notice = $state('');
	let viewMode = $state<'matrix' | 'detail'>('matrix');
	let statusFilter = $state('');
	let dateFilter = $state('');
	let searchFilter = $state('');
	let includeArchived = $state(false);
	let workingId = $state('');
	let creating = $state(false);
	let generatedAt = $state('');
	let exams = $state<ExamOption[]>([]);
	let packageMaps = $state<PackageMapOption[]>([]);
	let loadingMaps = $state(false);
	let savingSession = $state(false);

	const queryExamId = typeof window !== 'undefined' ? new URLSearchParams(window.location.search).get('exam_id') ?? '' : '';
	let examFilter = $state(queryExamId);
	let draft = $state<SessionDraft>({ examId: queryExamId, packageMapKey: '', date: '', startTime: '07:30', durationMinutes: 90, title: '' });

	const filteredRows = $derived(rows.filter((row) => {
		const keyword = searchFilter.trim().toLowerCase();
		if (!includeArchived && row.status === 'archived') return false;
		if (statusFilter && row.status !== statusFilter) return false;
		if (dateFilter && row.scheduled_start?.slice(0, 10) !== dateFilter) return false;
		if (keyword) {
			const text = [row.exam_title, row.title, row.subject_name, row.package_title, row.class_code, row.class_name, row.room_labels.join(' ')]
				.filter(Boolean)
				.join(' ')
				.toLowerCase();
			if (!text.includes(keyword)) return false;
		}
		return true;
	}));
	const summary = $derived(summarizeSessionRows(filteredRows));
	const matrixGroups = $derived(groupSesiCbtMatrix(filteredRows));
	const selectedPackageMap = $derived(packageMaps.find((item) => packageMapKey(item) === draft.packageMapKey) ?? packageMaps[0] ?? null);

	function readClientPayload(payload: unknown): SessionsResponse {
		if (payload && typeof payload === 'object' && 'data' in payload) {
			return (payload as { data: SessionsResponse }).data ?? {};
		}
		return (payload ?? {}) as SessionsResponse;
	}

	function asItems(payload: unknown): unknown[] {
		const data = readClientPayload(payload);
		if (Array.isArray(data.items)) return data.items;
		if (Array.isArray(payload)) return payload;
		return [];
	}

	function packageMapKey(item: PackageMapOption) {
		return item.local_id || item.id || `${item.package_id}:${item.class_id}`;
	}

	function packageMapLabel(item: PackageMapOption) {
		const subject = item.subject_name || 'Mapel';
		const kelas = item.class_code || item.class_name || 'Rombel';
		const paket = item.package_title || 'Paket siap';
		return `${subject} · ${kelas} · ${paket}`;
	}

	function defaultSessionTitle(item: PackageMapOption | null) {
		if (!item) return 'Sesi CBT';
		return `${item.subject_name || 'Mapel'} ${item.class_code || item.class_name || 'Rombel'}`;
	}

	function localDateTimeToIso(date: string, time: string) {
		if (!date || !time) return '';
		return new Date(`${date}T${time}:00+08:00`).toISOString();
	}

	function sessionEndIso(date: string, time: string, durationMinutes: number) {
		const start = localDateTimeToIso(date, time);
		if (!start) return '';
		return new Date(new Date(start).getTime() + Math.max(1, Number(durationMinutes || 0)) * 60_000).toISOString();
	}

	async function loadSessions() {
		loading = true;
		error = '';
		notice = '';
		try {
			const params = new URLSearchParams({ limit: '200' });
			if (examFilter.trim()) params.set('exam_id', examFilter.trim());
			if (includeArchived) params.set('include_archived', '1');
			if (statusFilter) params.set('status', statusFilter);
			if (dateFilter) params.set('date', dateFilter);
			const response = await fetch(`/api/asesmen/sessions?${params}`);
			const payload = readClientPayload(await response.json().catch(() => ({})));
			if (!response.ok) throw new Error((payload as { error?: string }).error ?? `HTTP ${response.status}`);
			rows = Array.isArray(payload.items) ? payload.items.map((item) => normalizeSesiCbtRow(item)) : [];
			generatedAt = payload.generated_at ?? new Date().toISOString();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Sesi CBT belum dapat dimuat.';
			rows = [];
		} finally {
			loading = false;
		}
	}

	async function loadExams() {
		try {
			const response = await fetch('/api/asesmen/exams?limit=80');
			const payload = await response.json().catch(() => ({}));
			if (!response.ok) throw new Error('Kegiatan belum dapat dimuat.');
			exams = asItems(payload).map((item) => item as ExamOption).filter((item) => item.id);
			if (!draft.examId && exams[0]?.id) draft = { ...draft, examId: exams[0].id };
			if (draft.examId) await loadPackageMaps(draft.examId);
		} catch {
			exams = [];
		}
	}

	async function loadPackageMaps(examId = draft.examId) {
		if (!examId) {
			packageMaps = [];
			return;
		}
		loadingMaps = true;
		try {
			const response = await fetch(`/api/asesmen/exams/${encodeURIComponent(examId)}/package-maps`);
			const payload = await response.json().catch(() => ({}));
			if (!response.ok) throw new Error('Paket siap belum dapat dimuat.');
			packageMaps = asItems(payload).map((item) => item as PackageMapOption).filter((item) => item.class_id && item.package_id);
			draft = { ...draft, examId, packageMapKey: packageMaps[0] ? packageMapKey(packageMaps[0]) : '' };
		} catch (err) {
			error = err instanceof Error ? err.message : 'Paket siap belum dapat dimuat.';
			packageMaps = [];
		} finally {
			loadingMaps = false;
		}
	}

	async function createSession() {
		const map = selectedPackageMap;
		if (!draft.examId) {
			error = 'Pilih kegiatan ujian dulu.';
			return;
		}
		if (!map) {
			error = 'Belum ada Paket Siap/Rombel. Lengkapi Paket Soal dan peserta/ruang dulu.';
			return;
		}
		const scheduledStart = localDateTimeToIso(draft.date, draft.startTime);
		const scheduledEnd = sessionEndIso(draft.date, draft.startTime, draft.durationMinutes);
		if (!scheduledStart || !scheduledEnd) {
			error = 'Tanggal dan jam mulai wajib diisi.';
			return;
		}
		savingSession = true;
		error = '';
		notice = '';
		try {
			const response = await fetch(`/api/asesmen/exams/${encodeURIComponent(draft.examId)}/sessions`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					package_id: map.package_id,
					class_id: map.class_id,
					scope_type: 'class',
					scope_ref: map.class_id,
					mix_policy: 'class_grouped',
					assignment_mode: 'balanced',
					title: draft.title.trim() || defaultSessionTitle(map),
					scheduled_start: scheduledStart,
					scheduled_end: scheduledEnd,
					status: 'draft'
				})
			});
			const payload = await response.json().catch(() => ({}));
			if (!response.ok) throw new Error(payload.error ?? `HTTP ${response.status}`);
			notice = 'Sesi draft tersimpan. Review readiness lalu jadwalkan/aktifkan dari daftar.';
			draft = { ...draft, title: '' };
			examFilter = draft.examId;
			await loadSessions();
			creating = false;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Sesi belum dapat dibuat.';
		} finally {
			savingSession = false;
		}
	}

	async function updateStatus(row: SesiCbtRow, status: SesiCbtStatus) {
		workingId = row.id;
		error = '';
		notice = '';
		try {
			const response = await fetch(`/api/asesmen/sessions/${encodeURIComponent(row.id)}/status`, {
				method: 'PATCH',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ status })
			});
			const payload = await response.json().catch(() => ({}));
			if (!response.ok) throw new Error(payload.error ?? `HTTP ${response.status}`);
			notice = `Status ${row.title} diubah menjadi ${sessionStatusLabel(status)}.`;
			await loadSessions();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Status sesi belum dapat diubah.';
		} finally {
			workingId = '';
		}
	}

	function formatDateTime(value?: string) {
		if (!value) return 'Belum dijadwalkan';
		const date = new Date(value);
		if (Number.isNaN(date.getTime())) return value;
		return new Intl.DateTimeFormat('id-ID', {
			dateStyle: 'medium',
			timeStyle: 'short',
			timeZone: 'Asia/Makassar'
		}).format(date) + ' WITA';
	}

	function statusAction(row: SesiCbtRow): { label: string; next?: SesiCbtStatus; disabled: boolean; title?: string } {
		if (row.status === 'active') return { label: 'Aktif', disabled: true };
		if (row.status === 'finished') return { label: 'Selesai', disabled: true };
		if (row.status === 'archived') return { label: 'Arsip', disabled: true };
		if (row.status === 'draft') return { label: 'Jadwalkan', next: 'scheduled', disabled: false };
		if (!row.ready_to_activate) return { label: 'Aktifkan', next: 'active', disabled: true, title: row.blockers.join(' · ') };
		return { label: 'Aktifkan', next: 'active', disabled: false };
	}

	onMount(() => {
		void loadSessions();
		void loadExams();
	});
</script>

<svelte:head><title>Sesi CBT · MTsN 2 Kolaka Utara</title></svelte:head>

<div class="space-y-5 p-4 sm:p-6">
	<section class="rounded-2xl border bg-card p-4 shadow-sm sm:p-5">
		<div class="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
			<div>
				<p class="text-xs font-semibold tracking-[0.18em] text-primary uppercase">Sesi CBT</p>
				<h1 class="mt-1 text-2xl font-bold tracking-tight text-foreground">Jadwal Sesi Ujian</h1>
				<p class="mt-2 max-w-3xl text-sm leading-6 text-muted-foreground">Modul operasional jadwal dan pelaksanaan CBT. Bank Soal dan Paket Soal tetap modul sendiri; di sini operator fokus melihat sesi, jadwal, ruang, status, dan monitor.</p>
			</div>
			<div class="flex flex-wrap gap-2">
				<a class="rounded-md border bg-background px-3 py-2 text-sm font-semibold text-foreground hover:bg-muted" href="/asesmen">Kegiatan Ujian</a>
				<a class="rounded-md border bg-background px-3 py-2 text-sm font-semibold text-foreground hover:bg-muted" href="/asesmen/sesi-lite">Mode Cepat HP</a>
				<button type="button" class="rounded-md border bg-background px-3 py-2 text-sm font-semibold text-foreground hover:bg-muted" onclick={() => (creating = !creating)}>{creating ? 'Tutup Form' : '+ Buat Sesi'}</button>
				<button type="button" class="rounded-md bg-primary px-3 py-2 text-sm font-semibold text-primary-foreground hover:bg-primary/90 disabled:opacity-60" onclick={() => void loadSessions()} disabled={loading}>{loading ? 'Memuat…' : 'Refresh'}</button>
			</div>
		</div>
	</section>

	{#if creating}
		<section class="rounded-xl border bg-background p-4">
			<div class="flex flex-col gap-2 sm:flex-row sm:items-start sm:justify-between">
				<div>
					<p class="text-xs font-semibold tracking-[0.16em] text-primary uppercase">Buat sesi</p>
					<h2 class="text-base font-bold text-foreground">Form sederhana seperti CBT lama</h2>
					<p class="mt-1 text-xs leading-5 text-muted-foreground">Pilih kegiatan dan Paket Siap/Rombel, isi tanggal-jam, lalu simpan sebagai draft. Aktivasi tetap lewat checklist kesiapan.</p>
				</div>
				<button type="button" class="rounded-md border px-3 py-2 text-xs font-semibold hover:bg-muted" onclick={() => draft.examId && void loadPackageMaps(draft.examId)} disabled={loadingMaps}>{loadingMaps ? 'Memuat…' : 'Refresh Paket Siap'}</button>
			</div>
			<div class="mt-4 grid gap-2 lg:grid-cols-[1.3fr_1.7fr_9rem_7rem_7rem_1fr_auto]">
				<select class="rounded-md border bg-card px-3 py-2 text-sm" bind:value={draft.examId} onchange={(event) => void loadPackageMaps(event.currentTarget.value)} aria-label="Kegiatan ujian">
					<option value="">Pilih kegiatan</option>
					{#each exams as exam (exam.id)}<option value={exam.id}>{exam.title || exam.nama || exam.id}</option>{/each}
				</select>
				<select class="rounded-md border bg-card px-3 py-2 text-sm" bind:value={draft.packageMapKey} disabled={packageMaps.length === 0} aria-label="Paket siap dan rombel">
					{#each packageMaps as item (packageMapKey(item))}<option value={packageMapKey(item)}>{packageMapLabel(item)}</option>{/each}
				</select>
				<input class="rounded-md border bg-card px-3 py-2 text-sm" type="date" bind:value={draft.date} aria-label="Tanggal" />
				<input class="rounded-md border bg-card px-3 py-2 text-sm" type="time" bind:value={draft.startTime} aria-label="Jam mulai" />
				<input class="rounded-md border bg-card px-3 py-2 text-sm" type="number" min="15" max="240" bind:value={draft.durationMinutes} aria-label="Durasi menit" />
				<input class="rounded-md border bg-card px-3 py-2 text-sm" placeholder="Judul opsional" bind:value={draft.title} />
				<button type="button" class="rounded-md bg-primary px-4 py-2 text-sm font-semibold text-primary-foreground disabled:opacity-60" onclick={() => void createSession()} disabled={savingSession || !draft.examId || !selectedPackageMap}>{savingSession ? 'Simpan…' : '+ Buat'}</button>
			</div>
			{#if packageMaps.length === 0}<p class="mt-2 text-xs text-amber-700">Belum ada Paket Siap/Rombel untuk kegiatan ini. Lengkapi dulu di Paket Soal dan Peserta & Ruang.</p>{/if}
		</section>
	{/if}

	<section class="grid gap-3 sm:grid-cols-2 lg:grid-cols-6">
		{#each [
			['Total', summary.total, 'Semua sesi'],
			['Aktif', summary.active, 'Terjadwal/aktif'],
			['Berlangsung', summary.running, 'Sedang ujian'],
			['Selesai', summary.finished, 'Sudah selesai'],
			['Arsip', summary.archived, 'Diarsipkan'],
			['Insiden', summary.incident, 'Perlu perhatian']
		] as item}
			<div class="rounded-xl border bg-background p-3">
				<p class="text-xs font-medium text-muted-foreground">{item[0]}</p>
				<p class="mt-1 text-2xl font-bold text-foreground">{item[1]}</p>
				<p class="text-[11px] text-muted-foreground">{item[2]}</p>
			</div>
		{/each}
	</section>

	<section class="rounded-xl border bg-background p-3">
		<div class="grid gap-2 md:grid-cols-[1.4fr_10rem_10rem_10rem_auto]">
			<input class="rounded-md border bg-card px-3 py-2 text-sm" placeholder="Cari kegiatan/mapel/rombel/ruang" bind:value={searchFilter} />
			<input class="rounded-md border bg-card px-3 py-2 text-sm" placeholder="ID kegiatan opsional" bind:value={examFilter} />
			<input class="rounded-md border bg-card px-3 py-2 text-sm" type="date" bind:value={dateFilter} />
			<select class="rounded-md border bg-card px-3 py-2 text-sm" bind:value={statusFilter}>
				<option value="">Semua status</option>
				<option value="draft">Draft</option>
				<option value="scheduled">Terjadwal</option>
				<option value="active">Aktif</option>
				<option value="finished">Selesai</option>
				<option value="archived">Arsip</option>
			</select>
			<label class="flex items-center gap-2 rounded-md border bg-card px-3 py-2 text-sm text-muted-foreground"><input type="checkbox" bind:checked={includeArchived} /> Arsip</label>
		</div>
		<div class="mt-3 flex flex-wrap items-center justify-between gap-2">
			<div class="flex rounded-lg border bg-muted/30 p-1 text-xs font-semibold">
				<button type="button" class={`rounded-md px-3 py-1.5 ${viewMode === 'matrix' ? 'bg-background shadow-sm' : 'text-muted-foreground'}`} onclick={() => (viewMode = 'matrix')}>Grup Matrix</button>
				<button type="button" class={`rounded-md px-3 py-1.5 ${viewMode === 'detail' ? 'bg-background shadow-sm' : 'text-muted-foreground'}`} onclick={() => (viewMode = 'detail')}>Detail Sesi</button>
			</div>
			<p class="text-xs text-muted-foreground">{generatedAt ? `Terakhir dimuat ${formatDateTime(generatedAt)}` : 'Belum dimuat'}</p>
		</div>
	</section>

	{#if error}<p class="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive" role="alert">{error}</p>{/if}
	{#if notice}<p class="rounded-md border border-blue-300 bg-blue-50 px-3 py-2 text-sm font-medium text-blue-800" role="status">{notice}</p>{/if}

	{#if loading && rows.length === 0}
		<p class="rounded-xl border bg-background p-6 text-center text-sm text-muted-foreground">Memuat sesi CBT…</p>
	{:else if filteredRows.length === 0}
		<div class="rounded-xl border border-dashed bg-background p-6 text-center">
			<p class="font-semibold text-foreground">Belum ada sesi CBT</p>
			<p class="mt-1 text-sm text-muted-foreground">Buat kegiatan ujian, pilih paket siap, susun peserta/ruang, lalu jadwalkan sesi.</p>
			<a class="mt-3 inline-flex rounded-md bg-primary px-3 py-2 text-sm font-semibold text-primary-foreground" href="/asesmen">Buka Kegiatan Ujian</a>
		</div>
	{:else if viewMode === 'matrix'}
		<section class="grid gap-3 lg:grid-cols-2">
			{#each matrixGroups as group (group.key)}
				<div class="rounded-xl border bg-background p-4">
					<div class="flex items-start justify-between gap-3">
						<div class="min-w-0">
							<p class="truncate text-base font-bold text-foreground">{group.anchor.subject_name || group.anchor.title}</p>
							<p class="mt-1 text-xs text-muted-foreground">{formatDateTime(group.anchor.scheduled_start)} · {group.rows.length} sesi</p>
						</div>
						<span class={`rounded-full border px-2.5 py-1 text-[11px] font-semibold ${sessionStatusTone(group.anchor.status)}`}>{sessionStatusLabel(group.anchor.status)}</span>
					</div>
					<div class="mt-3 grid grid-cols-3 gap-2 text-xs">
						<div class="rounded-lg bg-muted/40 p-2"><strong class="block text-foreground">{group.totalParticipants}</strong><span class="text-muted-foreground">Peserta</span></div>
						<div class="rounded-lg bg-muted/40 p-2"><strong class="block text-foreground">{group.roomLabels.length}</strong><span class="text-muted-foreground">Ruang</span></div>
						<div class="rounded-lg bg-muted/40 p-2"><strong class="block text-foreground">{group.classLabels.length}</strong><span class="text-muted-foreground">Rombel</span></div>
					</div>
					<div class="mt-3 flex flex-wrap gap-1">
						{#each group.classLabels.slice(0, 6) as label}<span class="rounded-full border bg-card px-2 py-0.5 text-[11px] text-muted-foreground">{label}</span>{/each}
						{#if group.classLabels.length > 6}<span class="text-[11px] text-muted-foreground">+{group.classLabels.length - 6}</span>{/if}
					</div>
				</div>
			{/each}
		</section>
	{:else}
		<section class="overflow-hidden rounded-xl border bg-background">
			<div class="hidden grid-cols-[1.5fr_1.1fr_9rem_8rem_8rem] gap-3 bg-muted/40 px-4 py-2 text-[11px] font-bold tracking-wide text-muted-foreground uppercase lg:grid">
				<span>Sesi</span><span>Kegiatan</span><span>Peserta/Ruang</span><span>Status</span><span class="text-right">Aksi</span>
			</div>
			{#each filteredRows as row (row.id)}
				{@const action = statusAction(row)}
				<div class="grid gap-3 border-t px-4 py-3 text-sm lg:grid-cols-[1.5fr_1.1fr_9rem_8rem_8rem] lg:items-center">
					<div class="min-w-0">
						<p class="truncate font-semibold text-foreground">{row.subject_name || row.title}</p>
						<p class="truncate text-xs text-muted-foreground">{row.package_title || 'Paket belum terbaca'} · {row.class_code || row.class_name || 'Rombel'}</p>
						<p class="mt-1 text-xs text-muted-foreground">{formatDateTime(row.scheduled_start)}</p>
					</div>
					<p class="text-xs leading-5 text-muted-foreground"><strong class="block text-foreground">{row.exam_title}</strong>{row.room_labels.join(', ') || 'Ruang belum terbaca'}</p>
					<p class="text-xs text-muted-foreground"><strong class="block text-foreground">{row.participant_count}</strong>{row.room_count} ruang</p>
					<span class={`w-fit rounded-full border px-2.5 py-1 text-[11px] font-semibold ${sessionStatusTone(row.status)}`}>{sessionStatusLabel(row.status)}</span>
					<div class="flex flex-wrap justify-start gap-2 lg:justify-end">
						{#if row.monitor_token}<a class="rounded-md border px-2 py-1 text-xs font-semibold hover:bg-muted" href={`/asesmen/pelaksanaan?monitor=${encodeURIComponent(row.monitor_token)}`}>Monitor</a>{/if}
						<button type="button" class="rounded-md border px-2 py-1 text-xs font-semibold hover:bg-muted disabled:opacity-50" disabled={workingId === row.id || action.disabled || !action.next} title={action.title} onclick={() => action.next && void updateStatus(row, action.next)}>{workingId === row.id ? '...' : action.label}</button>
					</div>
				</div>
			{/each}
		</section>
	{/if}
</div>
