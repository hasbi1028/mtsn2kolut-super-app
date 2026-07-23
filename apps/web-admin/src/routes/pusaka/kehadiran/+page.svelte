<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Badge } from '$lib/components/ui/badge';
	import { Input } from '$lib/components/ui/input';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { toast } from '$lib/components/ui/sonner';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import Pagination from '$lib/components/Pagination.svelte';
	import { readClientJson } from '$lib/client/api';

	interface AttendanceRecord {
		id: string;
		employee_nama: string;
		employee_nip: string;
		tanggal: string;
		jam_masuk: string | null;
		jam_pulang: string | null;
	}

	type AttendanceOverview = {
		records: AttendanceRecord[];
		total: number;
	};

	type ViewMode = 'normal' | 'compact';

	const SAMPLE_RECORDS: AttendanceRecord[] = [
		{ id: 's1', employee_nama: 'Ahmad Fauzi, S.Pd.',       employee_nip: '197501012005011001', tanggal: '2026-05-05', jam_masuk: '07:12:04 WITA', jam_pulang: '14:05:22 WITA' },
		{ id: 's2', employee_nama: 'Siti Rahayu, S.Pd.I.',     employee_nip: '198003152006042002', tanggal: '2026-05-05', jam_masuk: '07:18:31 WITA', jam_pulang: '14:02:47 WITA' },
		{ id: 's3', employee_nama: 'Muhammad Ilham, S.Pd.',    employee_nip: '199205102019031003', tanggal: '2026-05-05', jam_masuk: '07:23:09 WITA', jam_pulang: '' },
		{ id: 's4', employee_nama: 'Fitriani, S.Pd.',           employee_nip: '198807202015041004', tanggal: '2026-05-05', jam_masuk: '07:30:00 WITA', jam_pulang: '14:00:00 WITA' },
		{ id: 's5', employee_nama: 'Hasmawati, S.Ag.',          employee_nip: '197912052003122005', tanggal: '2026-05-05', jam_masuk: '07:09:55 WITA', jam_pulang: '13:58:11 WITA' },
		{ id: 's6', employee_nama: 'Rahmat Hidayat, S.Pd.',    employee_nip: '200001102022031006', tanggal: '2026-05-05', jam_masuk: '',              jam_pulang: '' },
		{ id: 's7', employee_nama: 'Nurjannah, S.Pd.',          employee_nip: '198504182009012007', tanggal: '2026-05-05', jam_masuk: '07:14:22 WITA', jam_pulang: '14:10:03 WITA' },
		{ id: 's8', employee_nama: 'Abdul Karim, S.Pd.I.',     employee_nip: '197806092001121008', tanggal: '2026-05-05', jam_masuk: '07:40:17 WITA', jam_pulang: '14:01:59 WITA' },
		{ id: 's9', employee_nama: 'Dewi Anggraini, S.Pd.',    employee_nip: '199308152020122009', tanggal: '2026-05-05', jam_masuk: '07:27:44 WITA', jam_pulang: '' },
		{ id: 's10', employee_nama: 'Syarifuddin, S.Pd.',       employee_nip: '198111302007011010', tanggal: '2026-05-05', jam_masuk: '07:33:06 WITA', jam_pulang: '14:07:38 WITA' },
	];

	let records   = $state<AttendanceRecord[]>([]);
	let total     = $state<number | null>(null);
	let startDate = $state('');
	let endDate   = $state('');
	let viewMode  = $state<ViewMode>('normal');
	let recordsPromise = $state<Promise<AttendanceOverview> | null>(null);
	let refreshing = $state(false);
	let sendingTelegram = $state(false);
	let loadedRangeKey = $state('');
	let attendanceRequestId = 0;
	let page = $state(1);
	let perPage = $state(15);
	let pagedRecords = $derived(perPage === 0 ? records : records.slice((page - 1) * perPage, page * perPage));

	function todayWita() {
		return new Intl.DateTimeFormat('en-CA', {
			timeZone: 'Asia/Makassar',
			year: 'numeric', month: '2-digit', day: '2-digit',
		}).format(new Date());
	}

	function rangeKey() {
		return `${startDate}:${endDate || startDate}`;
	}

	function stripWita(val: string | null) {
		if (!val) return 'Belum';
		return val.replace(/\s*WITA$/i, '').trim();
	}

	function attendanceStatus(r: AttendanceRecord) {
		if (r.jam_masuk && r.jam_pulang) return 'lengkap';
		if (r.jam_masuk) return 'masuk';
		return 'belum';
	}

	function isRecord(value: unknown): value is Record<string, unknown> {
		return typeof value === 'object' && value !== null;
	}

	function parseAttendancePayload(payload: unknown): AttendanceOverview {
		if (Array.isArray(payload)) {
			const nextRecords = payload as AttendanceRecord[];
			return { records: nextRecords, total: nextRecords.length };
		}
		if (!isRecord(payload)) return { records: [], total: 0 };
		const nextRecords = Array.isArray(payload.data) ? (payload.data as AttendanceRecord[]) : [];
		const meta = isRecord(payload.meta) ? payload.meta : {};
		const metaTotal = typeof meta.total === 'number' ? meta.total : nextRecords.length;
		return { records: nextRecords, total: metaTotal };
	}

	async function fetchAttendance(): Promise<AttendanceOverview> {
		if (!startDate) return { records: [], total: 0 };
		const params = new URLSearchParams();
		params.set('start_date', startDate);
		params.set('end_date', endDate || startDate);

		const res = await fetch(`/api/pusaka/attendance?${params}`);
		const payload = await readClientJson<unknown>(res);
		if (isRecord(payload)) {
			const error = payload.error;
			if (typeof error === 'string' && error.trim()) throw new Error(error);
		}
		return parseAttendancePayload(payload);
	}

	function applyOverview(overview: AttendanceOverview) {
		records = overview.records ?? [];
		total = overview.total ?? records.length;
	}

	function loadInitial() {
		if (!startDate) return;
		const requestId = ++attendanceRequestId;
		const nextRangeKey = rangeKey();
		records = [];
		total = null;
		recordsPromise = fetchAttendance()
			.then((overview) => {
				if (requestId === attendanceRequestId) {
					applyOverview(overview);
					loadedRangeKey = nextRangeKey;
					return overview;
				}
				return { records, total: total ?? records.length };
			})
			.catch((error: unknown) => {
				if (requestId === attendanceRequestId) throw error;
				return { records, total: total ?? records.length };
			});
	}

	async function load() {
		if (!startDate) return;
		page = 1;
		perPage = 15;
		if (!recordsPromise) {
			loadInitial();
			return;
		}
		const requestId = ++attendanceRequestId;
		const nextRangeKey = rangeKey();
		refreshing = true;
		try {
			const overview = await fetchAttendance();
			if (requestId === attendanceRequestId) {
				applyOverview(overview);
				loadedRangeKey = nextRangeKey;
				recordsPromise = Promise.resolve(overview);
			}
		} catch (error) {
			if (requestId !== attendanceRequestId) return;
			if (loadedRangeKey === nextRangeKey) {
				recordsPromise = Promise.resolve({ records, total: total ?? records.length });
				toast.error(attendanceErrorMessage(error));
			} else {
				recordsPromise = Promise.reject(error);
			}
		} finally {
			if (requestId === attendanceRequestId) {
				refreshing = false;
			}
		}
	}

	function retryAttendance(reset?: () => void) {
		reset?.();
		loadInitial();
	}

	function attendanceErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Gagal memuat data kehadiran';
	}

	function handleAttendanceRenderError(error: unknown) {
		console.error('PUSAKA attendance render failed', error);
	}

	function exportCSV() {
		if (records.length === 0) return;
		const header = 'Tanggal,NIP,Nama,Masuk,Pulang,Status';
		const rows = records.map(r => {
			const s = attendanceStatus(r);
			return [r.tanggal, r.employee_nip, `"${r.employee_nama}"`,
				stripWita(r.jam_masuk), stripWita(r.jam_pulang), s].join(',');
		});
		const csv = [header, ...rows].join('\n');
		const blob = new Blob([csv], { type: 'text/csv' });
		const url = URL.createObjectURL(blob);
		const a = document.createElement('a');
		a.href = url;
		a.download = `kehadiran_${startDate}${endDate && endDate !== startDate ? '_sd_' + endDate : ''}.csv`;
		a.click();
		URL.revokeObjectURL(url);
	}

	async function sendTelegramReport() {
		if (!startDate) {
			toast.error('Pilih tanggal laporan lebih dulu');
			return;
		}
		sendingTelegram = true;
		try {
			const res = await fetch('/api/pusaka/attendance-telegram/send', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ date: startDate, include_caption: true, include_image: true })
			});
			const payload = await readClientJson<Record<string, unknown>>(res);
			if (payload?.error) throw new Error(String(payload.error));
			const data = isRecord(payload?.data) ? payload.data : payload;
			const target = typeof data.target_chat_id_masked === 'string' ? data.target_chat_id_masked : 'Telegram';
			toast.success(`Laporan daftar hadir terkirim ke ${target}`);
		} catch (error) {
			toast.error(attendanceErrorMessage(error));
		} finally {
			sendingTelegram = false;
		}
	}

	onMount(() => {
		const today = todayWita();
		startDate = today;
		endDate   = today;
		loadInitial();
	});
</script>

<svelte:head><title>Data Kehadiran PUSAKA — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6">

	<div class="flex items-center gap-2 text-sm text-base-content/70">
		<a href={resolve('/pusaka')} class="hover:text-base-content">PUSAKA</a>
		<span>/</span>
		<span class="text-base-content font-medium">Data Kehadiran</span>
	</div>

	<div>
		<h1 class="text-2xl font-black text-base-content">Data Kehadiran Pegawai</h1>
		<p class="text-sm text-base-content/70 mt-1">Rekap kehadiran harian dari sistem PUSAKA Kemenag</p>
	</div>

	{#if total !== null}
	<div class="grid gap-3 md:grid-cols-3">
		<div class="rounded-xl border border-base-300 bg-base-100 px-4 py-3 flex flex-col gap-1">
			<p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground">Rekap Harian</p>
			<p class="text-2xl font-black text-foreground">{total ?? records.length}</p>
			<p class="text-xs text-muted-foreground">Rekaman kehadiran pada rentang tanggal terpilih</p>
		</div>
		<div class="rounded-xl border border-base-300 bg-base-100 px-4 py-3 flex flex-col gap-1">
			<p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground">Status Dominan</p>
			<p class="text-base font-black text-primary">
				{records.some((r) => attendanceStatus(r) === 'lengkap') ? 'Lengkap tersedia' : 'Mayoritas check-in'}
			</p>
			<p class="text-xs text-muted-foreground">Mayoritas pegawai lengkap masuk & pulang</p>
		</div>
		<div class="rounded-xl border border-base-300 bg-base-100 px-4 py-3 flex flex-col gap-1">
			<p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground">Periode</p>
			<p class="text-sm font-bold text-foreground">{startDate || '—'}</p>
			<p class="text-xs text-muted-foreground">sampai {endDate || startDate || '—'}</p>
		</div>
	</div>
	{:else}
	<div class="grid gap-3 md:grid-cols-3">
		<div class="skeleton h-24 w-full rounded-xl"></div>
		<div class="skeleton h-24 w-full rounded-xl"></div>
		<div class="skeleton h-24 w-full rounded-xl"></div>
	</div>
	{/if}

	<Card.Root class="overflow-hidden border-base-300 shadow-sm">
		<Card.Header class="border-b border-base-300 py-3">
			<div class="flex flex-col gap-3">
				<!-- Title + Count inline -->
				<div class="flex items-center justify-between gap-2">
					<Card.Title class="text-base">Rekap Kehadiran</Card.Title>
					<Card.Description class="!mt-0">
						{#if refreshing || total === null}
							<div class="skeleton h-4 w-20"></div>
						{:else if total !== null}
							<span class="text-xs font-semibold text-muted-foreground">{total} rekaman</span>
						{:else}
							—
						{/if}
					</Card.Description>
				</div>

				<!-- Date Range + Terapkan (inline) -->
				<div class="flex flex-wrap items-center gap-2">
					<div class="flex items-center gap-1.5">
						<input type="date" bind:value={startDate} class="input input-bordered h-9 min-w-0 border border-input bg-background px-2 text-xs" />
						<span class="text-xs text-muted-foreground font-medium" aria-hidden="true">⟶</span>
						<input type="date" bind:value={endDate} class="input input-bordered h-9 min-w-0 border border-input bg-background px-2 text-xs" />
					</div>
					<button class="inline-flex items-center justify-center rounded-lg h-9 px-4 text-sm font-bold text-primary-foreground bg-primary border border-primary hover:bg-primary/90 transition-colors" onclick={() => void load()}>
						{#if refreshing}<span class="loading loading-spinner loading-xs"></span>{/if}
						Terapkan
					</button>
				</div>

				<!-- Quick date shortcuts -->
				<div class="flex flex-wrap items-center gap-1.5">
					<button class="inline-flex items-center justify-center rounded-lg h-8 px-3 text-[11px] font-semibold text-primary border border-primary/30 bg-primary/5 hover:bg-primary/10 transition-colors" onclick={() => { const t = todayWita(); startDate = t; endDate = t; void load(); }}>
						Hari Ini
					</button>
					<button class="inline-flex items-center justify-center rounded-lg h-8 px-3 text-[11px] font-semibold text-primary border border-primary/30 bg-primary/5 hover:bg-primary/10 transition-colors" onclick={() => { const t = todayWita(); const d = new Date(t); d.setDate(d.getDate() - 6); startDate = d.toISOString().slice(0,10); endDate = t; void load(); }}>
						7 Hari
					</button>
					<button class="inline-flex items-center justify-center rounded-lg h-8 px-3 text-[11px] font-semibold text-primary border border-primary/30 bg-primary/5 hover:bg-primary/10 transition-colors" onclick={() => { const t = todayWita(); const d = new Date(t); d.setDate(1); startDate = d.toISOString().slice(0,10); endDate = t; void load(); }}>
						Bulan Ini
					</button>
				</div>

				<!-- Divider + Action buttons -->
				<hr class="border-border -mx-4" />
				<div class="flex flex-wrap items-center gap-1.5">
					<button class="inline-flex items-center justify-center rounded-lg h-8 px-3 text-xs font-medium border border-input bg-background text-foreground hover:bg-accent transition-colors disabled:opacity-40" onclick={exportCSV} disabled={records.length === 0}>
						<svg class="h-3.5 w-3.5 mr-1.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" /></svg>
						CSV
					</button>
					<button class="inline-flex items-center justify-center rounded-lg h-8 px-3 text-xs font-medium border border-input bg-background text-foreground hover:bg-accent transition-colors disabled:opacity-40" onclick={() => void sendTelegramReport()} disabled={sendingTelegram}>
						{#if sendingTelegram}<span class="loading loading-spinner loading-xs"></span>{/if}
						<svg class="h-3.5 w-3.5 mr-1.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" /></svg>
						Telegram
					</button>
					<a href={resolve('/pusaka/telegram-laporan')} class="inline-flex items-center justify-center rounded-lg h-8 px-3 text-xs font-medium border border-input bg-background text-foreground hover:bg-accent transition-colors">Atur Jadwal</a>
					<a href={resolve('/pusaka/antrian')} class="inline-flex items-center justify-center rounded-lg h-8 px-3 text-xs font-medium border border-input bg-background text-foreground hover:bg-accent transition-colors">Antrian</a>
					<span class="mx-1 text-xs text-border" aria-hidden="true">|</span>
					<div class="flex h-8 overflow-hidden rounded-lg border border-base-300 bg-base-100">
						<button
							class="flex items-center justify-center gap-1 px-2.5 text-[11px] font-medium transition-colors {viewMode === 'normal' ? 'bg-primary text-primary-content' : 'text-base-content/70 hover:bg-base-200/50'}"
							onclick={() => viewMode = 'normal'}
						>Normal</button>
						<div class="w-px bg-border"></div>
						<button
							class="flex items-center justify-center gap-1 px-2.5 text-[11px] font-medium transition-colors {viewMode === 'compact' ? 'bg-primary text-primary-content' : 'text-base-content/70 hover:bg-base-200/50'}"
							onclick={() => viewMode = 'compact'}
						>Ringkas</button>
					</div>
				</div>
			</div>
		</Card.Header>

		<Card.Content class="p-0">
			{#if viewMode === 'compact'}
				<!-- Tampilan Ringkas: langsung pakai state records, tidak perlu tunggu promise -->
				{@const displayRecords = pagedRecords.length > 0 ? pagedRecords : SAMPLE_RECORDS.slice((page - 1) * PER_PAGE, page * PER_PAGE)}
				{@const isSample = records.length === 0}
				{#if isSample}
				<div class="flex items-center gap-1.5 border-b border-warning/30 bg-warning/10 px-3 py-1.5 text-[11px] text-warning">
					<span class="font-semibold">Contoh tampilan</span>
					<span class="text-warning">— data di bawah adalah sampel. Terapkan filter tanggal untuk memuat data nyata.</span>
				</div>
				{/if}
				<div class="overflow-x-auto">
					<table class="table table-zebra table-xs text-[10px]">
						<thead>
							<tr>
								<th>Tanggal</th>
								<th>Nama Pegawai</th>
								<th class="hidden sm:table-cell">NIP</th>
								<th class="text-center">Masuk</th>
								<th class="text-center">Pulang</th>
								<th class="text-center">Status</th>
							</tr>
						</thead>
						<tbody>
							{#each displayRecords as r, i (r.id)}
								{@const s = attendanceStatus(r)}
								<tr class="{isSample ? 'opacity-75' : ''}">
									<td class="whitespace-nowrap text-base-content/70 text-[9px]">{r.tanggal}</td>
									<td class="font-medium text-base-content">{r.employee_nama}</td>
									<td class="hidden sm:table-cell font-mono text-base-content/70 text-[9px]">{r.employee_nip}</td>
									<td class="text-center text-base-content text-[9px]">{stripWita(r.jam_masuk)}</td>
									<td class="text-center text-base-content text-[9px]">{stripWita(r.jam_pulang)}</td>
									<td class="text-center">
										{#if s === 'lengkap'}
											<span class="badge badge-xs badge-success">Lengkap</span>
										{:else if s === 'masuk'}
											<span class="badge badge-xs badge-warning">Masuk</span>
										{:else}
											<span class="badge badge-xs badge-ghost">Belum</span>
										{/if}
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
					<div class="border-t border-base-300 bg-base-200/50 px-3 py-1 text-right text-[9px] text-base-content/70">
						{#if isSample}
							Contoh data (10 sampel)
						{:else}
							{displayRecords.length} rekaman · {startDate}{endDate && endDate !== startDate ? ' s/d ' + endDate : ''}
						{/if}
					</div>
				</div>
				<div class="px-3 py-2 border-t border-base-300">
					<Pagination bind:page total={records.length || SAMPLE_RECORDS.length} bind:perPage />
				</div>
			{:else}
				<!-- Tampilan Normal: pakai AsyncContent seperti semula -->
				<AsyncContent promise={recordsPromise} onerror={handleAttendanceRenderError}>
					{#snippet pending()}
					<div class="space-y-3 p-4">
						<Skeleton class="h-12 w-full" />
						<Skeleton class="h-14 w-full" />
						<Skeleton class="h-14 w-full" />
						<Skeleton class="h-14 w-full" />
					</div>
					{/snippet}

					{#snippet failed(error, reset)}
						<div class="p-4">
							<RecoveryPanel
								compact
								title="Kehadiran Belum Tersaji"
								message={attendanceErrorMessage(error)}
								onRetry={() => retryAttendance(reset)}
							/>
						</div>
					{/snippet}

					{#snippet children(value)}
						{@const currentRecords = pagedRecords}
					<div class="hidden overflow-x-auto lg:block">
					<table class="table table-zebra table-xs">
						<thead>
							<tr>
								<th>Tanggal</th>
								<th>Nama Pegawai</th>
								<th class="hidden sm:table-cell">NIP</th>
								<th class="text-center">Masuk</th>
								<th class="text-center">Pulang</th>
								<th class="text-center">Status</th>
							</tr>
						</thead>
						<tbody>
							{#each currentRecords as r, i (r.id)}
							{@const s = attendanceStatus(r)}
							<tr>
								<td class="whitespace-nowrap text-base-content/70">{r.tanggal}</td>
								<td class="font-medium">{r.employee_nama}</td>
								<td class="hidden sm:table-cell text-base-content/70 font-mono text-xs">{r.employee_nip}</td>
								<td class="text-center">{stripWita(r.jam_masuk)}</td>
								<td class="text-center">{stripWita(r.jam_pulang)}</td>
								<td class="text-center">
									{#if s === 'lengkap'}
										<span class="badge badge-sm badge-success">Lengkap</span>
									{:else if s === 'masuk'}
										<span class="badge badge-sm badge-warning">Masuk</span>
									{:else}
										<span class="badge badge-sm badge-ghost">Belum</span>
									{/if}
								</td>
							</tr>
							{:else}
							<tr>
								<td colspan={6} class="py-12 text-center text-base-content/70">
									Tidak ada data kehadiran untuk rentang tanggal ini.
								</td>
							</tr>
							{/each}
						</tbody>
					</table>
					</div>
					<div class="hidden lg:block border-t border-base-300 px-3 py-2">
						<Pagination bind:page total={records.length} bind:perPage />
					</div>

					<div class="lg:hidden">
						{#if currentRecords.length === 0}
							<div class="mx-4 my-4 rounded-2xl border border-dashed border-base-300 bg-base-200/50 px-4 py-10 text-center text-sm text-base-content/70">
								Tidak ada data kehadiran untuk rentang tanggal ini.
							</div>
						{:else}
							<ul class="divide-y divide-border border-b border-base-300">
								{#each currentRecords as r, i (r.id)}
									{@const s = attendanceStatus(r)}
									<li class="flex items-center gap-3 px-4 py-2.5">
										<span class="w-7 shrink-0 text-xs font-medium tabular-nums text-base-content/70">#{i + 1}</span>
										<div class="min-w-0 grow">
											<p class="truncate text-sm font-medium text-base-content">{r.employee_nama}</p>
											<p class="truncate font-mono text-[11px] text-base-content/70">{r.employee_nip} · {r.tanggal}</p>
										</div>
										<div class="flex shrink-0 items-center gap-2 text-[11px] tabular-nums text-base-content/70">
											<span class="hidden min-[360px]:inline">{stripWita(r.jam_masuk)}</span>
											<span aria-hidden="true" class="hidden min-[360px]:inline">→</span>
											<span>{stripWita(r.jam_pulang)}</span>
										</div>
										<span class="shrink-0">
											{#if s === 'lengkap'}
												<span class="inline-flex h-2 w-2 rounded-full bg-primary" title="Lengkap" aria-label="Lengkap"></span>
											{:else if s === 'masuk'}
												<span class="inline-flex h-2 w-2 rounded-full bg-warning" title="Masuk" aria-label="Masuk"></span>
											{:else}
												<span class="inline-flex h-2 w-2 rounded-full bg-base-200-foreground/40" title="Belum" aria-label="Belum"></span>
											{/if}
										</span>
									</li>
								{/each}
							</ul>
							<div class="border-t border-base-300 bg-base-200/50 px-4 py-2 text-right text-[11px] text-base-content/70">
								{currentRecords.length} rekaman · {startDate}{endDate && endDate !== startDate ? ' s/d ' + endDate : ''}
							</div>
						{/if}
					</div>
					{/snippet}
				</AsyncContent>
			{/if}
		</Card.Content>
	</Card.Root>

</div>
