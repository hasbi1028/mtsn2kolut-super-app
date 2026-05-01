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

	let records   = $state<AttendanceRecord[]>([]);
	let total     = $state<number | null>(null);
	let startDate = $state('');
	let endDate   = $state('');
	let recordsPromise = $state<Promise<AttendanceOverview> | null>(null);
	let refreshing = $state(false);
	let loadedRangeKey = $state('');

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

	function apiErrorMessage(payload: unknown) {
		if (!isRecord(payload)) return '';
		const error = payload.error;
		if (typeof error === 'string' && error.trim()) return error;
		const message = payload.message;
		if (typeof message === 'string' && message.trim()) return message;
		return '';
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
		const payload = await res.json().catch(() => null);
		const message = apiErrorMessage(payload);
		if (!res.ok || message) throw new Error(message || `Error ${res.status}`);
		return parseAttendancePayload(payload);
	}

	function applyOverview(overview: AttendanceOverview) {
		records = overview.records ?? [];
		total = overview.total ?? records.length;
	}

	function loadInitial() {
		if (!startDate) return;
		const nextRangeKey = rangeKey();
		records = [];
		total = null;
		recordsPromise = fetchAttendance().then((overview) => {
			applyOverview(overview);
			loadedRangeKey = nextRangeKey;
			return overview;
		});
	}

	async function load() {
		if (!startDate) return;
		if (!recordsPromise) {
			loadInitial();
			return;
		}
		const nextRangeKey = rangeKey();
		refreshing = true;
		try {
			const overview = await fetchAttendance();
			applyOverview(overview);
			loadedRangeKey = nextRangeKey;
			recordsPromise = Promise.resolve(overview);
		} catch (error) {
			if (loadedRangeKey === nextRangeKey) {
				recordsPromise = Promise.resolve({ records, total: total ?? records.length });
				toast.error(attendanceErrorMessage(error));
			} else {
				recordsPromise = Promise.reject(error);
			}
		} finally {
			refreshing = false;
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

	<div class="flex items-center gap-2 text-sm text-slate-500">
		<a href={resolve('/pusaka')} class="hover:text-slate-700">PUSAKA</a>
		<span>/</span>
		<span class="text-slate-700 font-medium">Data Kehadiran</span>
	</div>

	<div>
		<h1 class="text-2xl font-semibold text-slate-800">Data Kehadiran Pegawai</h1>
		<p class="text-sm text-muted-foreground mt-1">Rekap kehadiran harian dari sistem PUSAKA Kemenag</p>
	</div>

	<div class="grid gap-4 md:grid-cols-3">
		<Card.Root class="border-emerald-100 bg-gradient-to-br from-white via-white to-emerald-50/80 md:col-span-2">
			<Card.Content class="flex items-start justify-between gap-4 p-5">
				<div class="space-y-1">
					<p class="text-xs font-semibold uppercase tracking-[0.2em] text-emerald-700">Rekap Harian</p>
					<p class="text-2xl font-semibold text-slate-900">{total ?? records.length}</p>
					<p class="text-sm text-slate-600">Rekaman kehadiran pada rentang tanggal terpilih</p>
				</div>
				<div class="rounded-2xl border border-emerald-200 bg-white/80 px-4 py-3 text-right shadow-sm">
					<p class="text-xs uppercase tracking-[0.18em] text-slate-500">Status dominan</p>
					<p class="mt-1 text-base font-semibold text-emerald-800">
						{records.some((r) => attendanceStatus(r) === 'lengkap') ? 'Lengkap tersedia' : 'Mayoritas check-in'}
					</p>
				</div>
			</Card.Content>
		</Card.Root>

		<Card.Root class="border-slate-200 bg-white">
			<Card.Content class="space-y-2 p-5">
				<p class="text-xs font-semibold uppercase tracking-[0.2em] text-slate-500">Periode</p>
				<p class="text-base font-semibold text-slate-900">{startDate || '—'}</p>
				<p class="text-sm text-slate-500">sampai {endDate || startDate || '—'}</p>
			</Card.Content>
		</Card.Root>
	</div>

	<Card.Root class="overflow-hidden border-slate-200 shadow-sm">
		<Card.Header class="border-b border-slate-100 bg-gradient-to-r from-white to-emerald-50/40">
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
						<Input type="date" bind:value={startDate} class="h-10 min-w-0 bg-white" />
						<span class="text-center text-sm text-muted-foreground">s/d</span>
						<Input type="date" bind:value={endDate} class="h-10 min-w-0 bg-white" />
					</div>
					<LoadingButton class="h-10 w-full sm:w-auto" size="sm" onclick={() => void load()} loading={refreshing} loadingLabel="Memuat..." label="Terapkan" />
					<div class="grid grid-cols-2 gap-2 sm:flex sm:flex-wrap sm:justify-end">
						<LoadingButton variant="outline" class="h-10 w-full bg-white sm:w-auto" size="sm" onclick={exportCSV} disabled={records.length === 0} label="↓ CSV" />
						<LoadingButton variant="outline" class="h-10 w-full bg-white sm:w-auto" size="sm" href={resolve('/pusaka/antrian')} label="Antrian →" />
					</div>
				</div>
			</div>
		</Card.Header>

		<Card.Content class="p-0">
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
									<Badge variant="default" class="bg-emerald-600">Lengkap</Badge>
								{:else if s === 'masuk'}
									<Badge variant="outline" class="text-amber-600 border-amber-200">Masuk</Badge>
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
					<div class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
						<div class="flex items-start justify-between gap-3">
							<div class="min-w-0">
								<p class="text-xs font-semibold uppercase tracking-[0.18em] text-slate-400">#{i + 1} • {r.tanggal}</p>
								<p class="mt-1 text-base font-semibold text-slate-900">{r.employee_nama}</p>
								<p class="mt-1 break-all font-mono text-xs text-slate-500">{r.employee_nip}</p>
							</div>
							<div class="shrink-0">
								{#if s === 'lengkap'}
									<Badge variant="default" class="bg-emerald-600">Lengkap</Badge>
								{:else if s === 'masuk'}
									<Badge variant="outline" class="text-amber-600 border-amber-200">Masuk</Badge>
								{:else}
									<Badge variant="secondary">Belum</Badge>
								{/if}
							</div>
						</div>

						<div class="mt-4 grid grid-cols-2 gap-3">
							<div class="rounded-xl border border-slate-200 bg-slate-50 px-3 py-2">
								<p class="text-[11px] font-semibold uppercase tracking-[0.16em] text-slate-400">Masuk</p>
								<p class="mt-1 text-sm font-medium text-slate-800">{stripWita(r.jam_masuk)}</p>
							</div>
							<div class="rounded-xl border border-slate-200 bg-slate-50 px-3 py-2">
								<p class="text-[11px] font-semibold uppercase tracking-[0.16em] text-slate-400">Pulang</p>
								<p class="mt-1 text-sm font-medium text-slate-800">{stripWita(r.jam_pulang)}</p>
							</div>
						</div>
					</div>
				{:else}
					<div class="rounded-2xl border border-dashed border-slate-300 bg-slate-50 px-4 py-10 text-center text-sm text-slate-500">
						Tidak ada data kehadiran untuk rentang tanggal ini.
					</div>
				{/each}
			</div>
				{/snippet}
			</AsyncContent>
		</Card.Content>
	</Card.Root>

</div>
