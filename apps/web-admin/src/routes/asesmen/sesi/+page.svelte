<script lang="ts">
	import { onMount } from 'svelte';
	import {
		groupSesiCbtMatrix,
		makassarDateKey,
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
	type ScheduleDraft = { id: string; date: string; startTime: string; durationMinutes: number; title: string };
	type CsvSessionDraft = SessionDraft & { sourceLine: number };

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
	let selectedPackageMapKeys = $state<string[]>([]);
	let csvText = $state('');
	let csvFileName = $state('');
	let editSchedule = $state<ScheduleDraft | null>(null);

	const queryExamId = typeof window !== 'undefined' ? new URLSearchParams(window.location.search).get('exam_id') ?? '' : '';
	let examFilter = $state(queryExamId);
	let draft = $state<SessionDraft>({ examId: queryExamId, packageMapKey: '', date: '', startTime: '07:30', durationMinutes: 90, title: '' });

	const filteredRows = $derived(rows.filter((row) => {
		const keyword = searchFilter.trim().toLowerCase();
		if (!includeArchived && row.status === 'cancelled') return false;
		if (statusFilter && row.status !== statusFilter) return false;
		if (dateFilter && makassarDateKey(row.scheduled_start) !== dateFilter) return false;
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
	const selectedPackageMaps = $derived(packageMaps.filter((item) => selectedPackageMapKeys.includes(packageMapKey(item))));

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
			selectedPackageMapKeys = packageMaps[0] ? [packageMapKey(packageMaps[0])] : [];
		} catch (err) {
			error = err instanceof Error ? err.message : 'Paket siap belum dapat dimuat.';
			packageMaps = [];
		} finally {
			loadingMaps = false;
		}
	}

	async function createSessionForMap(input: SessionDraft, map: PackageMapOption) {
		const scheduledStart = localDateTimeToIso(input.date, input.startTime);
		const scheduledEnd = sessionEndIso(input.date, input.startTime, input.durationMinutes);
		if (!scheduledStart || !scheduledEnd) throw new Error('Tanggal dan jam mulai wajib diisi.');
		const response = await fetch(`/api/asesmen/exams/${encodeURIComponent(input.examId)}/sessions`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({
				package_id: map.package_id,
				class_id: map.class_id,
				scope_type: 'class',
				scope_ref: map.class_id,
				mix_policy: 'class_grouped',
				assignment_mode: 'balanced',
				title: input.title.trim() || defaultSessionTitle(map),
				scheduled_start: scheduledStart,
				scheduled_end: scheduledEnd,
				status: 'draft'
			})
		});
		const payload = await response.json().catch(() => ({}));
		if (!response.ok) throw new Error(payload.error ?? `HTTP ${response.status}`);
	}

	async function createSession() {
		if (!draft.examId) {
			error = 'Pilih kegiatan ujian dulu.';
			return;
		}
		if (!selectedPackageMap) {
			error = 'Belum ada Paket Siap/Rombel. Lengkapi Paket Soal dan peserta/ruang dulu.';
			return;
		}
		savingSession = true;
		error = '';
		notice = '';
		try {
			await createSessionForMap(draft, selectedPackageMap);
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

	async function createBatchSessions() {
		if (!draft.examId) {
			error = 'Pilih kegiatan ujian dulu.';
			return;
		}
		if (selectedPackageMaps.length === 0) {
			error = 'Pilih minimal satu Paket Siap/Rombel.';
			return;
		}
		savingSession = true;
		error = '';
		notice = '';
		let created = 0;
		try {
			for (const map of selectedPackageMaps) {
				await createSessionForMap({ ...draft, packageMapKey: packageMapKey(map) }, map);
				created += 1;
			}
			notice = `${created} sesi draft dibuat dari matrix Paket Siap/Rombel.`;
			examFilter = draft.examId;
			await loadSessions();
			creating = false;
		} catch (err) {
			error = err instanceof Error ? `${created} sesi tersimpan, proses berhenti: ${err.message}` : 'Batch sesi belum dapat dibuat.';
		} finally {
			savingSession = false;
		}
	}

	function togglePackageMap(key: string, checked: boolean) {
		selectedPackageMapKeys = checked ? Array.from(new Set([...selectedPackageMapKeys, key])) : selectedPackageMapKeys.filter((item) => item !== key);
		if (checked) draft = { ...draft, packageMapKey: key };
	}

	function parseCsvRows(): CsvSessionDraft[] {
		const lines = csvText.split(/\r?\n/).map((line) => line.trim()).filter(Boolean);
		if (lines.length < 2) return [];
		const headers = lines[0].split(',').map((header) => header.trim());
		return lines.slice(1).map((line, index) => {
			const cells = line.split(',').map((cell) => cell.trim());
			const record = Object.fromEntries(headers.map((header, cellIndex) => [header, cells[cellIndex] ?? '']));
			return {
				examId: draft.examId,
				packageMapKey: record.package_map_key ?? '',
				date: record.date ?? '',
				startTime: record.start_time ?? '',
				durationMinutes: Math.max(1, Number(record.duration || record.duration_minutes || 0)),
				title: record.title ?? '',
				sourceLine: index + 2
			};
		});
	}

	async function importCsvSessions() {
		const imports = parseCsvRows();
		if (!draft.examId) {
			error = 'Pilih kegiatan ujian dulu.';
			return;
		}
		if (imports.length === 0) {
			error = 'CSV minimal punya header dan satu baris data.';
			return;
		}
		savingSession = true;
		error = '';
		notice = '';
		let created = 0;
		try {
			for (const item of imports) {
				const map = packageMaps.find((candidate) => packageMapKey(candidate) === item.packageMapKey);
				if (!map) throw new Error(`Baris ${item.sourceLine}: package_map_key tidak ditemukan.`);
				await createSessionForMap(item, map);
				created += 1;
			}
			notice = `${created} sesi draft dibuat dari CSV${csvFileName ? ` ${csvFileName}` : ''}.`;
			await loadSessions();
			creating = false;
		} catch (err) {
			error = err instanceof Error ? `${created} sesi tersimpan, import berhenti: ${err.message}` : 'Import CSV belum dapat diproses.';
		} finally {
			savingSession = false;
		}
	}

	async function loadCsvFile(event: Event) {
		const input = event.currentTarget as HTMLInputElement;
		const file = input.files?.[0];
		if (!file) return;
		csvFileName = file.name;
		csvText = await file.text();
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

	async function deleteDraft(row: SesiCbtRow) {
		if (row.status !== 'draft') {
			error = 'Hanya sesi draft yang dapat dihapus dari halaman ini.';
			return;
		}
		workingId = row.id;
		error = '';
		notice = '';
		try {
			const response = await fetch(`/api/asesmen/sessions/${encodeURIComponent(row.id)}`, { method: 'DELETE' });
			const payload = await response.json().catch(() => ({}));
			if (!response.ok) throw new Error(payload.error ?? `HTTP ${response.status}`);
			notice = `Draft ${row.title} dihapus.`;
			await loadSessions();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Draft sesi belum dapat dihapus.';
		} finally {
			workingId = '';
		}
	}

	async function finalizeOverdue(row: SesiCbtRow) {
		workingId = row.id;
		error = '';
		notice = '';
		try {
			const response = await fetch(`/api/asesmen/sessions/${encodeURIComponent(row.id)}/finalize-overdue`, { method: 'POST' });
			const payload = await response.json().catch(() => ({}));
			if (!response.ok) throw new Error(payload.error ?? `HTTP ${response.status}`);
			notice = `Finalisasi overdue untuk ${row.title} diproses.`;
			await loadSessions();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Finalisasi overdue belum dapat diproses.';
		} finally {
			workingId = '';
		}
	}

	function openScheduleEdit(row: SesiCbtRow) {
		const start = row.scheduled_start ? new Date(row.scheduled_start) : null;
		const end = row.scheduled_end ? new Date(row.scheduled_end) : null;
		const duration = start && end ? Math.max(1, Math.round((end.getTime() - start.getTime()) / 60_000)) : 90;
		editSchedule = {
			id: row.id,
			date: makassarDateKey(row.scheduled_start),
			startTime: row.scheduled_start
				? new Intl.DateTimeFormat('en-GB', { timeZone: 'Asia/Makassar', hour: '2-digit', minute: '2-digit', hour12: false }).format(new Date(row.scheduled_start))
				: '07:30',
			durationMinutes: duration,
			title: row.title
		};
	}

	async function updateSchedule() {
		if (!editSchedule) return;
		const scheduledStart = localDateTimeToIso(editSchedule.date, editSchedule.startTime);
		const scheduledEnd = sessionEndIso(editSchedule.date, editSchedule.startTime, editSchedule.durationMinutes);
		if (!scheduledStart || !scheduledEnd) {
			error = 'Tanggal dan jam mulai wajib diisi.';
			return;
		}
		workingId = editSchedule.id;
		error = '';
		notice = '';
		try {
			const response = await fetch(`/api/asesmen/sessions/${encodeURIComponent(editSchedule.id)}/schedule`, {
				method: 'PATCH',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ scheduled_start: scheduledStart, scheduled_end: scheduledEnd })
			});
			const payload = await response.json().catch(() => ({}));
			if (!response.ok) throw new Error(payload.error ?? `HTTP ${response.status}`);
			notice = `Jadwal ${editSchedule.title} diperbarui.`;
			editSchedule = null;
			await loadSessions();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Jadwal sesi belum dapat diperbarui.';
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
		if (row.status === 'cancelled') return { label: 'Arsip/Batal', disabled: true };
		if (row.status === 'draft') return { label: 'Jadwalkan', next: 'scheduled', disabled: false };
		if (!row.ready_to_activate) return { label: 'Aktifkan', next: 'active', disabled: true, title: row.blockers.join(' · ') };
		return { label: 'Aktifkan', next: 'active', disabled: false };
	}

	function blockerRepairLink(blocker: string, row: SesiCbtRow) {
		if (blocker.includes('Kegiatan')) return '/asesmen';
		if (blocker.includes('Paket')) return row.exam_id ? `/paket-soal?exam_id=${encodeURIComponent(row.exam_id)}` : '/paket-soal';
		if (blocker.includes('Peserta') || blocker.includes('Kursi') || blocker.includes('Ruang')) return row.exam_id ? `/asesmen/kegiatan/${encodeURIComponent(row.exam_id)}` : '/asesmen/persiapan';
		if (blocker.includes('Jadwal')) return `/asesmen/sesi?exam_id=${encodeURIComponent(row.exam_id)}`;
		return '/asesmen/persiapan';
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
			<div class="mt-4 grid gap-2 lg:grid-cols-[1.3fr_1.7fr_9rem_7rem_7rem_1fr_auto_auto]">
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
				<button type="button" class="rounded-md border px-4 py-2 text-sm font-semibold hover:bg-muted disabled:opacity-60" onclick={() => void createBatchSessions()} disabled={savingSession || !draft.examId || selectedPackageMaps.length === 0}>Batch {selectedPackageMaps.length}</button>
			</div>
			{#if packageMaps.length > 0}
				<div class="mt-4 rounded-lg border bg-card p-3">
					<div class="flex flex-wrap items-center justify-between gap-2">
						<p class="text-xs font-semibold text-foreground">Matrix Paket Siap/Rombel</p>
						<button type="button" class="rounded-md border px-2 py-1 text-[11px] font-semibold hover:bg-muted" onclick={() => (selectedPackageMapKeys = packageMaps.map((item) => packageMapKey(item)))}>Pilih semua</button>
					</div>
					<div class="mt-2 grid gap-2 md:grid-cols-2 xl:grid-cols-3">
						{#each packageMaps as item (packageMapKey(item))}
							{@const key = packageMapKey(item)}
							<label class="flex items-start gap-2 rounded-md border bg-background p-2 text-xs">
								<input class="mt-0.5" type="checkbox" checked={selectedPackageMapKeys.includes(key)} onchange={(event) => togglePackageMap(key, event.currentTarget.checked)} />
								<span class="min-w-0">
									<span class="block truncate font-semibold text-foreground">{packageMapLabel(item)}</span>
									<span class="block text-[11px] text-muted-foreground">key: {key}</span>
								</span>
							</label>
						{/each}
					</div>
				</div>
				<div class="mt-4 rounded-lg border bg-card p-3">
					<div class="flex flex-col gap-2 lg:flex-row lg:items-start lg:justify-between">
						<div>
							<p class="text-xs font-semibold text-foreground">Import CSV jadwal</p>
							<p class="mt-1 text-[11px] text-muted-foreground">Kolom: package_map_key,date,start_time,duration,title. Semua baris memakai kegiatan yang sedang dipilih.</p>
						</div>
						<input class="text-xs" type="file" accept=".csv,text/csv,text/plain" onchange={(event) => void loadCsvFile(event)} aria-label="Pilih file CSV jadwal" />
					</div>
					<textarea class="mt-2 min-h-24 w-full rounded-md border bg-background px-3 py-2 font-mono text-xs" bind:value={csvText} placeholder="package_map_key,date,start_time,duration,title&#10;paket-7a,2026-05-12,07:30,90,IPA 7A"></textarea>
					<div class="mt-2 flex items-center justify-between gap-2">
						<p class="text-[11px] text-muted-foreground">{parseCsvRows().length} baris siap diproses{csvFileName ? ` dari ${csvFileName}` : ''}</p>
						<button type="button" class="rounded-md border px-3 py-2 text-xs font-semibold hover:bg-muted disabled:opacity-60" onclick={() => void importCsvSessions()} disabled={savingSession || parseCsvRows().length === 0}>Import CSV</button>
					</div>
				</div>
			{/if}
			{#if packageMaps.length === 0}<p class="mt-2 text-xs text-amber-700">Belum ada Paket Siap/Rombel untuk kegiatan ini. Lengkapi dulu di Paket Soal dan Peserta & Ruang.</p>{/if}
		</section>
	{/if}

	<section class="grid gap-3 sm:grid-cols-2 lg:grid-cols-6">
		{#each [
			['Total', summary.total, 'Semua sesi'],
			['Aktif', summary.active, 'Terjadwal/aktif'],
			['Berlangsung', summary.running, 'Sedang ujian'],
			['Selesai', summary.finished, 'Sudah selesai'],
			['Arsip/Batal', summary.cancelled, 'Dibatalkan'],
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
				<option value="cancelled">Arsip/Batal</option>
			</select>
			<label class="flex items-center gap-2 rounded-md border bg-card px-3 py-2 text-sm text-muted-foreground"><input type="checkbox" bind:checked={includeArchived} /> Arsip/Batal</label>
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
				<div class="grid gap-3 border-t px-4 py-3 text-sm lg:grid-cols-[1.5fr_1.1fr_9rem_8rem_14rem] lg:items-center">
					<div class="min-w-0">
						<p class="truncate font-semibold text-foreground">{row.subject_name || row.title}</p>
						<p class="truncate text-xs text-muted-foreground">{row.package_title || 'Paket belum terbaca'} · {row.class_code || row.class_name || 'Rombel'}</p>
						<p class="mt-1 text-xs text-muted-foreground">{formatDateTime(row.scheduled_start)}</p>
						<div class="mt-2 flex flex-wrap gap-1">
							<a class="rounded-md border px-2 py-0.5 text-[11px] font-semibold hover:bg-muted" href={`/asesmen/sesi/${encodeURIComponent(row.id)}/absen`}>Absen</a>
							<a class="rounded-md border px-2 py-0.5 text-[11px] font-semibold hover:bg-muted" href={`/asesmen/sesi/${encodeURIComponent(row.id)}/ba`}>BA</a>
							<a class="rounded-md border px-2 py-0.5 text-[11px] font-semibold hover:bg-muted" href={`/asesmen/sesi/${encodeURIComponent(row.id)}/kartu`}>Kartu</a>
						</div>
					</div>
					<p class="text-xs leading-5 text-muted-foreground"><strong class="block text-foreground">{row.exam_title}</strong>{row.room_labels.join(', ') || 'Ruang belum terbaca'}</p>
					<p class="text-xs text-muted-foreground"><strong class="block text-foreground">{row.participant_count}</strong>{row.room_count} ruang</p>
					<div>
						<span class={`w-fit rounded-full border px-2.5 py-1 text-[11px] font-semibold ${sessionStatusTone(row.status)}`}>{sessionStatusLabel(row.status)}</span>
						{#if row.blockers.length > 0}
							<div class="mt-2 space-y-1">
								{#each row.blockers as blocker}
									<a class="block text-[11px] font-medium text-amber-700 underline underline-offset-2" href={blockerRepairLink(blocker, row)}>{blocker}</a>
								{/each}
							</div>
						{:else}
							<p class="mt-2 text-[11px] font-medium text-emerald-700">Siap diaktifkan</p>
						{/if}
					</div>
					<div class="flex flex-wrap justify-start gap-2 lg:justify-end">
						<button type="button" class="rounded-md border px-2 py-1 text-xs font-semibold hover:bg-muted disabled:opacity-50" onclick={() => openScheduleEdit(row)} disabled={workingId === row.id}>Jadwal</button>
						<button type="button" class="rounded-md border px-2 py-1 text-xs font-semibold hover:bg-muted disabled:opacity-50" disabled={workingId === row.id || action.disabled || !action.next} title={action.title} onclick={() => action.next && void updateStatus(row, action.next)}>{workingId === row.id ? '...' : action.label}</button>
						{#if row.status === 'active'}<button type="button" class="rounded-md border px-2 py-1 text-xs font-semibold hover:bg-muted disabled:opacity-50" disabled={workingId === row.id} onclick={() => void updateStatus(row, 'finished')}>Selesai</button>{/if}
						{#if row.status !== 'finished' && row.status !== 'cancelled'}<button type="button" class="rounded-md border px-2 py-1 text-xs font-semibold hover:bg-muted disabled:opacity-50" disabled={workingId === row.id} onclick={() => void updateStatus(row, 'cancelled')}>Batal</button>{/if}
						<button type="button" class="rounded-md border px-2 py-1 text-xs font-semibold hover:bg-muted disabled:opacity-50" disabled={workingId === row.id} onclick={() => void finalizeOverdue(row)}>Finalize overdue</button>
						{#if row.status === 'draft'}<button type="button" class="rounded-md border border-destructive/30 px-2 py-1 text-xs font-semibold text-destructive hover:bg-destructive/10 disabled:opacity-50" disabled={workingId === row.id} onclick={() => void deleteDraft(row)}>Hapus draft</button>{/if}
					</div>
				</div>
			{/each}
		</section>
	{/if}

	{#if editSchedule}
		<div class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
			<section class="w-full max-w-md rounded-xl border bg-background p-4 shadow-xl">
				<div class="flex items-start justify-between gap-3">
					<div>
						<p class="text-xs font-semibold tracking-[0.16em] text-primary uppercase">Edit jadwal</p>
						<h2 class="text-base font-bold text-foreground">{editSchedule.title}</h2>
					</div>
					<button type="button" class="rounded-md border px-2 py-1 text-xs font-semibold hover:bg-muted" onclick={() => (editSchedule = null)}>Tutup</button>
				</div>
				<div class="mt-4 grid gap-2 sm:grid-cols-3">
					<label class="text-xs font-medium text-muted-foreground" for="edit-session-date">Tanggal<input id="edit-session-date" class="mt-1 w-full rounded-md border bg-card px-3 py-2 text-sm" type="date" bind:value={editSchedule.date} /></label>
					<label class="text-xs font-medium text-muted-foreground" for="edit-session-start">Mulai<input id="edit-session-start" class="mt-1 w-full rounded-md border bg-card px-3 py-2 text-sm" type="time" bind:value={editSchedule.startTime} /></label>
					<label class="text-xs font-medium text-muted-foreground" for="edit-session-duration">Durasi<input id="edit-session-duration" class="mt-1 w-full rounded-md border bg-card px-3 py-2 text-sm" type="number" min="15" max="240" bind:value={editSchedule.durationMinutes} /></label>
				</div>
				<button type="button" class="mt-4 w-full rounded-md bg-primary px-4 py-2 text-sm font-semibold text-primary-foreground disabled:opacity-60" disabled={workingId === editSchedule.id} onclick={() => void updateSchedule()}>{workingId === editSchedule.id ? 'Menyimpan…' : 'Simpan jadwal'}</button>
			</section>
		</div>
	{/if}
</div>
