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
		jam_masuk: string | null;
		jam_pulang: string | null;
	}

	let records   = $state<AttendanceRecord[]>([]);
	let total     = $state<number | null>(null);
	let startDate = $state('');
	let endDate   = $state('');
	let loading   = $state(false);
	let error     = $state('');

	function todayWita() {
		return new Intl.DateTimeFormat('en-CA', {
			timeZone: 'Asia/Makassar',
			year: 'numeric', month: '2-digit', day: '2-digit',
		}).format(new Date());
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

	async function load() {
		if (!startDate) return;
		loading = true;
		error = '';
		try {
			const params = new URLSearchParams();
			params.set('start_date', startDate);
			params.set('end_date', endDate || startDate);

			const res  = await fetch(`/api/attendance?${params}`);
			const data = await res.json();

			if (!res.ok) {
				error = data?.error ?? `Error ${res.status}`;
				return;
			}

			records = data.data ?? [];
			total   = data.meta?.total ?? records.length;
		} catch {
			error = 'Gagal memuat data kehadiran';
		} finally {
			loading = false;
		}
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
		load();
	});
</script>

<svelte:head><title>Data Kehadiran PUSAKA — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6">

	<div class="flex items-center gap-2 text-sm text-slate-500">
		<a href="/pusaka" class="hover:text-slate-700">PUSAKA</a>
		<span>/</span>
		<span class="text-slate-700 font-medium">Data Kehadiran</span>
	</div>

	<div>
		<h1 class="text-2xl font-semibold text-slate-800">Data Kehadiran Pegawai</h1>
		<p class="text-sm text-muted-foreground mt-1">Rekap kehadiran harian dari sistem PUSAKA Kemenag</p>
	</div>

	<Card.Root>
		<Card.Header>
			<div class="flex flex-wrap items-start gap-3">
				<div class="grow">
					<Card.Title>Rekap Kehadiran</Card.Title>
					<Card.Description>
						{#if loading}
							Memuat data...
						{:else if total !== null}
							{total} rekaman ditemukan
						{:else}
							—
						{/if}
					</Card.Description>
				</div>
				<div class="flex flex-wrap items-center gap-2">
					<div class="flex items-center gap-2">
						<Input type="date" bind:value={startDate} class="w-auto h-9" />
						<span class="text-muted-foreground text-sm">s/d</span>
						<Input type="date" bind:value={endDate} class="w-auto h-9" />
					</div>
					<Button size="sm" onclick={load} disabled={loading}>
						{loading ? 'Memuat...' : 'Terapkan'}
					</Button>
					<Button variant="outline" size="sm" onclick={exportCSV} disabled={records.length === 0}>
						↓ CSV
					</Button>
					<Button variant="outline" size="sm" href="/pusaka/antrian">
						Antrian →
					</Button>
				</div>
			</div>
		</Card.Header>

		{#if error}
			<Card.Content>
				<div class="rounded-md border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800">
					{error}
				</div>
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
								{loading ? 'Memuat...' : 'Tidak ada data kehadiran untuk rentang tanggal ini.'}
							</Table.Cell>
						</Table.Row>
					{/each}
				</Table.Body>
			</Table.Root>
		</Card.Content>
	</Card.Root>

</div>
