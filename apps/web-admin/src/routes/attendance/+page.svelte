<script lang="ts">
	import { onMount } from 'svelte';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';

	interface AttendanceRecord {
		id: string;
		employee_nama: string;
		employee_nip: string;
		tanggal: string;
		jam_masuk: string;
		jam_pulang: string;
	}

	interface WorkerEntry {
		worker_id: string;
		active_consumers: number;
		target_concurrency: number;
		headless: boolean;
		reported_at: string;
	}

	interface WorkerStatus {
		active_workers: WorkerEntry[];
		total: number;
		queue: { queued: number; running: number; success: number; failed: number };
	}

	let records      = $state<AttendanceRecord[]>([]);
	let workerStatus = $state<WorkerStatus | null>(null);
	let startDate    = $state('');
	let endDate      = $state('');
	let loading      = $state(false);
	let error        = $state('');

	function todayWita() {
		return new Intl.DateTimeFormat('en-CA', {
			timeZone: 'Asia/Makassar',
			year: 'numeric', month: '2-digit', day: '2-digit',
		}).format(new Date());
	}

	function stripWita(val: string | null) {
		if (!val) return '—';
		return val.replace(/\s*WITA$/i, '');
	}

	function attendanceStatus(r: AttendanceRecord) {
		if (r.jam_masuk && r.jam_pulang) return 'lengkap';
		if (r.jam_masuk) return 'masuk';
		return 'belum';
	}

	async function load() {
		loading = true;
		error = '';
		try {
			const params = new URLSearchParams();
			if (startDate && endDate) {
				params.set('start_date', startDate);
				params.set('end_date', endDate);
			} else if (startDate) {
				params.set('date', startDate);
				params.set('limit', '200');
			}
			
			const res  = await fetch(`/api/attendance?${params}`);
			const data = await res.json();
			if (data.error) { error = data.error; return; }
			records = data.data ?? data.items ?? data ?? [];
		} catch {
			error = 'Gagal memuat data absensi';
		} finally {
			loading = false;
		}
	}

	async function loadWorkerStatus() {
		try {
			const res  = await fetch('/api/worker/status');
			const data = await res.json();
			workerStatus = data.data ?? null;
		} catch { /* non-critical */ }
	}

	function exportCSV() {
		if (records.length === 0) return;
		const header = 'Tanggal,NIP,Nama,Masuk,Pulang,Status';
		const rows = records.map(r => {
			const s = attendanceStatus(r);
			return [r.tanggal, r.employee_nip, `"${r.employee_nama}"`, stripWita(r.jam_masuk), stripWita(r.jam_pulang), s].join(',');
		});
		const csv = [header, ...rows].join('\n');
		const blob = new Blob([csv], { type: 'text/csv' });
		const url = URL.createObjectURL(blob);
		const a = document.createElement('a');
		a.href = url;
		a.download = `absensi_${startDate}${endDate ? '_sd_' + endDate : ''}.csv`;
		a.click();
		URL.revokeObjectURL(url);
	}

	onMount(() => {
		startDate = todayWita();
		endDate = todayWita();
		load();
		loadWorkerStatus();
		const interval = setInterval(loadWorkerStatus, 30_000);
		return () => clearInterval(interval);
	});
</script>

<svelte:head><title>Absensi — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6">

	<div>
		<h1 class="text-2xl font-semibold text-slate-800">Absensi Pegawai</h1>
		<p class="text-sm text-muted-foreground mt-1">Rekap kehadiran harian dari sistem Pusaka Kemenag</p>
	</div>

	<!-- Status cards -->
	<div class="grid gap-4 sm:grid-cols-3">
		<Card.Root>
			<Card.Header class="pb-2">
				<Card.Description>Worker Aktif</Card.Description>
			</Card.Header>
			<Card.Content class="pt-0">
				<div class="flex items-center gap-2">
					<span class="text-2xl font-bold">{workerStatus?.total ?? '—'}</span>
					{#if workerStatus}
						<Badge variant={workerStatus.total > 0 ? 'default' : 'destructive'}>
							{workerStatus.total > 0 ? 'Online' : 'Offline'}
						</Badge>
					{/if}
				</div>
			</Card.Content>
		</Card.Root>

		<Card.Root>
			<Card.Header class="pb-2">
				<Card.Description>Consumers</Card.Description>
			</Card.Header>
			<Card.Content class="pt-0">
				<span class="text-2xl font-bold">
					{workerStatus?.active_workers?.reduce((s, w) => s + w.active_consumers, 0) ?? '—'}
				</span>
				{#if workerStatus}
					<span class="text-sm text-muted-foreground ml-1">
						/ {workerStatus.active_workers?.reduce((s, w) => s + w.target_concurrency, 0)} kapasitas
					</span>
				{/if}
			</Card.Content>
		</Card.Root>

		<Card.Root>
			<Card.Header class="pb-2">
				<Card.Description>Antrian Job</Card.Description>
			</Card.Header>
			<Card.Content class="pt-0">
				{#if workerStatus?.queue}
					{@const q = workerStatus.queue}
					<div class="flex flex-wrap gap-2">
						{#if q.running > 0}
							<Badge variant="default">{q.running} berjalan</Badge>
						{/if}
						{#if q.queued > 0}
							<Badge variant="outline">{q.queued} antre</Badge>
						{/if}
						{#if q.running === 0 && q.queued === 0}
							<Badge variant="secondary">Idle</Badge>
						{/if}
					</div>
					<p class="text-xs text-muted-foreground mt-1">{q.success} sukses · {q.failed} gagal</p>
				{:else}
					<span class="text-2xl font-bold">—</span>
				{/if}
			</Card.Content>
		</Card.Root>
	</div>

	<!-- Attendance table -->
	<Card.Root>
		<Card.Header>
			<div class="flex flex-wrap items-start gap-3">
				<div class="grow">
					<Card.Title>Rekap Absensi</Card.Title>
					<Card.Description>{records.length} rekaman ditemukan</Card.Description>
				</div>
				<div class="flex flex-wrap items-center gap-2">
					<div class="flex items-center gap-2">
						<Input type="date" bind:value={startDate} class="w-auto h-9" />
						<span class="text-muted-foreground">s/d</span>
						<Input type="date" bind:value={endDate} class="w-auto h-9" />
					</div>
					<Button variant="default" size="sm" onclick={load} disabled={loading}>
						{loading ? 'Memuat...' : 'Terapkan'}
					</Button>
					<Button variant="outline" size="sm" onclick={exportCSV} disabled={records.length === 0}>
						Export CSV
					</Button>
					<Button variant="outline" size="sm" href="/attendance/queue">
						Antrian →
					</Button>
				</div>
			</div>
		</Card.Header>

		{#if error}
			<Card.Content>
				<p class="text-sm text-destructive">{error}</p>
			</Card.Content>
		{/if}

		<Card.Content class="p-0 overflow-x-auto">
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
					{#each records as r, i}
						{@const s = attendanceStatus(r)}
						<Table.Row>
							<Table.Cell class="text-muted-foreground">{i + 1}</Table.Cell>
							<Table.Cell class="text-sm whitespace-nowrap">{r.tanggal}</Table.Cell>
							<Table.Cell class="font-medium">{r.employee_nama}</Table.Cell>
							<Table.Cell class="hidden sm:table-cell text-muted-foreground font-mono text-xs">{r.employee_nip}</Table.Cell>
							<Table.Cell class="text-center">{stripWita(r.jam_masuk)}</Table.Cell>
							<Table.Cell class="text-center">{stripWita(r.jam_pulang)}</Table.Cell>
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
								{loading ? 'Memuat...' : 'Tidak ada data absensi untuk rentang tanggal ini.'}
							</Table.Cell>
						</Table.Row>
					{/each}
				</Table.Body>
			</Table.Root>
		</Card.Content>
	</Card.Root>
</div>
