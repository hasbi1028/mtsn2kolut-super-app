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
	let loadedRangeKey = $state('');
	let attendanceRequestId = 0;

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
		if (!val) return '—';
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

	onMount(() => {
		const today = todayWita();
		startDate = today;
		endDate   = today;
		loadInitial();
	});
</script>

<svelte:head><title>Data Kehadiran PUSAKA — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6">

	<div class="flex items-center gap-2 text-sm text-muted-foreground">
		<a href={resolve('/pusaka')} class="hover:text-foreground">PUSAKA</a>
		<span>/</span>
		<span class="text-foreground font-medium">Data Kehadiran</span>
	</div>

	<div>
		<h1 class="text-2xl font-semibold text-foreground">Data Kehadiran Pegawai</h1>
		<p class="text-sm text-muted-foreground mt-1">Rekap kehadiran harian dari sistem PUSAKA Kemenag</p>
	</div>

	<div class="grid gap-4 md:grid-cols-3">
		<Card.Root class="border-primary/20 bg-gradient-to-br from-card via-card to-primary/10 md:col-span-2">
			<Card.Content class="flex items-start justify-between gap-4 p-5">
				<div class="space-y-1">
					<p class="text-xs font-semibold uppercase tracking-[0.2em] text-primary">Rekap Harian</p>
					<p class="text-2xl font-semibold text-foreground">{total ?? records.length}</p>
					<p class="text-sm text-muted-foreground">Rekaman kehadiran pada rentang tanggal terpilih</p>
				</div>
				<div class="rounded-2xl border border-primary/20 bg-card/80 px-4 py-3 text-right shadow-sm">
					<p class="text-xs uppercase tracking-[0.18em] text-muted-foreground">Status dominan</p>
					<p class="mt-1 text-base font-semibold text-primary">
						{records.some((r) => attendanceStatus(r) === 'lengkap') ? 'Lengkap tersedia' : 'Mayoritas check-in'}
					</p>
				</div>
			</Card.Content>
		</Card.Root>

		<Card.Root class="border-border bg-card">
			<Card.Content class="space-y-2 p-5">
				<p class="text-xs font-semibold uppercase tracking-[0.2em] text-muted-foreground">Periode</p>
				<p class="text-base font-semibold text-foreground">{startDate || '—'}</p>
				<p class="text-sm text-muted-foreground">sampai {endDate || startDate || '—'}</p>
			</Card.Content>
		</Card.Root>
	</div>

	<Card.Root class="overflow-hidden border-border shadow-sm">
		<Card.Header class="border-b border-border bg-gradient-to-r from-card to-primary/10">
			<div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
				<div class="grow">
					<Card.Title>Rekap Kehadiran</Card.Title>
					<Card.Description>
						{#if refreshing}
							Memuat data...
						{:else if total !== null}
							{total} rekaman ditemukan
						{:else}
							—
						{/if}
					</Card.Description>
				</div>
				<div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-[1fr_auto_auto] xl:items-end">
					<div class="grid gap-2 sm:grid-cols-[1fr_auto_1fr] sm:items-center">
						<Input type="date" bind:value={startDate} class="h-10 min-w-0 bg-card" />
						<span class="text-center text-sm text-muted-foreground">s/d</span>
						<Input type="date" bind:value={endDate} class="h-10 min-w-0 bg-card" />
					</div>
					<LoadingButton class="h-10 w-full sm:w-auto" size="sm" onclick={() => void load()} loading={refreshing} loadingLabel="Memuat..." label="Terapkan" />
					<div class="grid grid-cols-2 gap-2 sm:flex sm:flex-wrap sm:justify-end">
						<LoadingButton variant="outline" class="h-10 w-full bg-card sm:w-auto" size="sm" onclick={exportCSV} disabled={records.length === 0} label="↓ CSV" />
						<div class="col-span-2 flex h-10 overflow-hidden rounded-md border border-border bg-card sm:col-span-1">
							<button
								class="flex flex-1 items-center justify-center gap-1.5 px-3 text-xs font-medium transition-colors {viewMode === 'normal' ? 'bg-primary text-primary-foreground' : 'text-muted-foreground hover:bg-muted/50'}"
								onclick={() => viewMode = 'normal'}
							>
								<svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="18" height="18" rx="2"/><path d="M3 9h18M9 21V9"/></svg>
								Normal
							</button>
							<div class="w-px bg-border"></div>
							<button
								class="flex flex-1 items-center justify-center gap-1.5 px-3 text-xs font-medium transition-colors {viewMode === 'compact' ? 'bg-primary text-primary-foreground' : 'text-muted-foreground hover:bg-muted/50'}"
								onclick={() => viewMode = 'compact'}
							>
								<svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="18" height="18" rx="2"/><path d="M3 8h18M3 13h18M3 18h18"/></svg>
								Ringkas
							</button>
						</div>
						<LoadingButton variant="outline" class="h-10 w-full bg-card sm:w-auto" size="sm" href={resolve('/pusaka/antrian')} label="Antrian →" />
					</div>
				</div>
			</div>
		</Card.Header>

		<Card.Content class="p-0">
			{#if viewMode === 'compact'}
				<!-- Tampilan Ringkas: langsung pakai state records, tidak perlu tunggu promise -->
				{@const displayRecords = records.length > 0 ? records : SAMPLE_RECORDS}
				{@const isSample = records.length === 0}
				{#if isSample}
				<div class="flex items-center gap-1.5 border-b border-warning/30 bg-warning/10 px-3 py-1.5 text-[11px] text-warning">
					<span class="font-semibold">Contoh tampilan</span>
					<span class="text-warning">— data di bawah adalah sampel. Terapkan filter tanggal untuk memuat data nyata.</span>
				</div>
				{/if}
				<div class="overflow-x-auto">
					<table class="w-full border-collapse text-[9px] leading-tight">
						<thead>
							<tr class="border-b border-border bg-muted/50">
								<th class="w-4 py-1 pl-0.5 pr-0 text-left font-semibold text-muted-foreground">#</th>
								<th class="whitespace-nowrap px-0 py-1 text-left font-semibold text-muted-foreground">Tanggal</th>
								<th class="px-0 py-1 text-left font-semibold text-muted-foreground">Nama Pegawai</th>
								<th class="hidden px-0 py-1 text-left font-semibold text-muted-foreground sm:table-cell">NIP</th>
								<th class="px-0 py-1 text-center font-semibold text-muted-foreground">Masuk</th>
								<th class="px-0 py-1 text-center font-semibold text-muted-foreground">Pulang</th>
								<th class="py-1 pl-0 pr-0.5 text-center font-semibold text-muted-foreground">Status</th>
							</tr>
						</thead>
						<tbody>
							{#each displayRecords as r, i (r.id)}
								{@const s = attendanceStatus(r)}
								<tr class="border-b border-border {i % 2 === 1 ? 'bg-muted/50' : 'bg-card'} hover:bg-primary/10 {isSample ? 'opacity-75' : ''}">
									<td class="py-0.5 pl-0.5 pr-0 text-muted-foreground">{i + 1}</td>
									<td class="whitespace-nowrap px-0 py-0.5 text-muted-foreground">{r.tanggal}</td>
									<td class="px-0 py-0.5 font-medium text-foreground">{r.employee_nama}</td>
									<td class="hidden px-0 py-0.5 font-mono text-muted-foreground sm:table-cell">{r.employee_nip}</td>
									<td class="px-0 py-0.5 text-center text-foreground">{stripWita(r.jam_masuk)}</td>
									<td class="px-0 py-0.5 text-center text-foreground">{stripWita(r.jam_pulang)}</td>
									<td class="py-0.5 pl-0 pr-0.5 text-center">
										{#if s === 'lengkap'}
											<span class="inline-block rounded-sm px-0 py-0 text-[7px] font-semibold bg-primary/15 text-primary">Lengkap</span>
										{:else if s === 'masuk'}
											<span class="inline-block rounded-sm px-0 py-0 text-[7px] font-semibold bg-warning/15 text-warning">Masuk</span>
										{:else}
											<span class="inline-block rounded-sm px-0 py-0 text-[7px] font-semibold bg-muted text-muted-foreground">Belum</span>
										{/if}
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
					<div class="border-t border-border bg-muted/50 px-3 py-1.5 text-right text-[10px] text-muted-foreground">
						{#if isSample}
							Contoh data (10 sampel)
						{:else}
							{displayRecords.length} rekaman · {startDate}{endDate && endDate !== startDate ? ' s/d ' + endDate : ''}
						{/if}
					</div>
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
						{@const currentRecords = (value as AttendanceOverview).records}
					<div class="hidden overflow-x-auto lg:block">
					<Table.Root>
						<Table.Header>
							<Table.Row>
								<Table.Head class="w-10">#</Table.Head>
								<Table.Head>Tanggal</Table.Head>
								<Table.Head>Nama Pegawai</Table.Head>
								<Table.Head class="hidden sm:table-cell">NIP</Table.Head>
								<Table.Head class="text-center">Masuk</Table.Head>
								<Table.Head class="text-center">Pulang</Table.Head>
								<Table.Head class="text-center">Status</Table.Head>
							</Table.Row>
						</Table.Header>
						<Table.Body>
							{#each currentRecords as r, i (r.id)}
							{@const s = attendanceStatus(r)}
							<Table.Row>
								<Table.Cell class="text-muted-foreground">{i + 1}</Table.Cell>
								<Table.Cell class="text-sm whitespace-nowrap">{r.tanggal}</Table.Cell>
								<Table.Cell class="font-medium">{r.employee_nama}</Table.Cell>
								<Table.Cell class="hidden sm:table-cell text-muted-foreground font-mono text-xs">{r.employee_nip}</Table.Cell>
								<Table.Cell class="text-center text-sm">{stripWita(r.jam_masuk)}</Table.Cell>
								<Table.Cell class="text-center text-sm">{stripWita(r.jam_pulang)}</Table.Cell>
								<Table.Cell class="text-center">
									{#if s === 'lengkap'}
										<Badge variant="default" class="bg-primary">Lengkap</Badge>
									{:else if s === 'masuk'}
										<Badge variant="outline" class="text-warning border-warning/30">Masuk</Badge>
									{:else}
										<Badge variant="secondary">Belum</Badge>
									{/if}
								</Table.Cell>
							</Table.Row>
						{:else}
							<Table.Row>
								<Table.Cell colspan={7} class="py-12 text-center text-muted-foreground">
									Tidak ada data kehadiran untuk rentang tanggal ini.
								</Table.Cell>
							</Table.Row>
						{/each}
						</Table.Body>
					</Table.Root>
					</div>

					<div class="grid gap-3 p-4 lg:hidden">
						{#each currentRecords as r, i (r.id)}
						{@const s = attendanceStatus(r)}
						<div class="rounded-2xl border border-border bg-card p-4 shadow-sm">
							<div class="flex items-start justify-between gap-3">
								<div class="min-w-0">
									<p class="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">#{i + 1} • {r.tanggal}</p>
									<p class="mt-1 text-base font-semibold text-foreground">{r.employee_nama}</p>
									<p class="mt-1 break-all font-mono text-xs text-muted-foreground">{r.employee_nip}</p>
								</div>
								<div class="shrink-0">
									{#if s === 'lengkap'}
										<Badge variant="default" class="bg-primary">Lengkap</Badge>
									{:else if s === 'masuk'}
										<Badge variant="outline" class="text-warning border-warning/30">Masuk</Badge>
									{:else}
										<Badge variant="secondary">Belum</Badge>
									{/if}
								</div>
							</div>
							<div class="mt-4 grid grid-cols-2 gap-3">
								<div class="rounded-xl border border-border bg-muted/50 px-3 py-2">
									<p class="text-[11px] font-semibold uppercase tracking-[0.16em] text-muted-foreground">Masuk</p>
									<p class="mt-1 text-sm font-medium text-foreground">{stripWita(r.jam_masuk)}</p>
								</div>
								<div class="rounded-xl border border-border bg-muted/50 px-3 py-2">
									<p class="text-[11px] font-semibold uppercase tracking-[0.16em] text-muted-foreground">Pulang</p>
									<p class="mt-1 text-sm font-medium text-foreground">{stripWita(r.jam_pulang)}</p>
								</div>
							</div>
						</div>
					{:else}
						<div class="rounded-2xl border border-dashed border-border bg-muted/50 px-4 py-10 text-center text-sm text-muted-foreground">
							Tidak ada data kehadiran untuk rentang tanggal ini.
						</div>
					{/each}
					</div>
					{/snippet}
				</AsyncContent>
			{/if}
		</Card.Content>
	</Card.Root>

</div>
