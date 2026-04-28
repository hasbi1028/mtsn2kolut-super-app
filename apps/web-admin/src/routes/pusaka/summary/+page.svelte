<script lang="ts">
	import { onMount } from 'svelte';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Badge } from '$lib/components/ui/badge';

	type SummaryRow = {
		employee_id: string; employee_nama: string; employee_nip: string;
		total_days: number; complete_days: number;
		missing_checkout: number; missing_checkin: number;
	};

	let startDate = $state('');
	let endDate   = $state('');
	let summary   = $state<SummaryRow[]>([]);
	let loading   = $state(false);
	let error     = $state('');

	function getDefaultDates() {
		const now = new Date();
		const start = new Date(now.getFullYear(), now.getMonth(), 1);
		const end   = new Date(now.getFullYear(), now.getMonth() + 1, 0);
		const fmt = (d: Date) => d.toISOString().split('T')[0];
		return { start: fmt(start), end: fmt(end) };
	}

	async function load() {
		if (!startDate || !endDate) return;
		loading = true; error = '';
		try {
			const res = await fetch(`/api/attendance/summary?start_date=${startDate}&end_date=${endDate}`);
			const data = await res.json();
			summary = data.data ?? data ?? [];
		} catch {
			error = 'Gagal memuat ringkasan kehadiran';
		} finally {
			loading = false;
		}
	}

	function exportCSV() {
		if (summary.length === 0) return;
		const header = 'Nama,NIP,Total Hari,Lengkap,Tanpa Pulang,Tanpa Masuk';
		const rows = summary.map(r =>
			[`"${r.employee_nama}"`, r.employee_nip, r.total_days, r.complete_days, r.missing_checkout, r.missing_checkin].join(',')
		);
		const csv = [header, ...rows].join('\n');
		const blob = new Blob([csv], { type: 'text/csv' });
		const url = URL.createObjectURL(blob);
		const a = document.createElement('a');
		a.href = url;
		a.download = `ringkasan_kehadiran_${startDate}_${endDate}.csv`;
		a.click();
		URL.revokeObjectURL(url);
	}

	onMount(() => {
		const d = getDefaultDates();
		startDate = d.start;
		endDate = d.end;
		load();
	});
</script>

<svelte:head><title>Ringkasan Kehadiran PUSAKA — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6">

	<div class="flex items-center gap-2 text-sm text-slate-500">
		<a href="/pusaka" class="hover:text-slate-700">PUSAKA</a>
		<span>/</span>
		<span class="text-slate-700 font-medium">Ringkasan Kehadiran</span>
	</div>

	<div class="flex flex-wrap items-end justify-between gap-4">
		<div>
			<h1 class="text-2xl font-semibold text-slate-800">Ringkasan Kehadiran</h1>
			<p class="text-sm text-slate-500 mt-1">Akumulasi kehadiran pegawai dari PUSAKA Kemenag dalam periode tertentu</p>
		</div>
		<div class="flex flex-wrap items-center gap-3">
			<div class="flex items-center gap-2">
				<Input type="date" bind:value={startDate} class="w-auto h-9" />
				<span class="text-muted-foreground text-sm">s/d</span>
				<Input type="date" bind:value={endDate} class="w-auto h-9" />
			</div>
			<Button onclick={load} disabled={loading}>{loading ? '...' : 'Tampilkan'}</Button>
			<Button variant="outline" onclick={exportCSV} disabled={summary.length === 0}>↓ CSV</Button>
		</div>
	</div>

	{#if error}
		<div class="rounded-md bg-red-50 border border-red-200 p-4 text-sm text-red-800">{error}</div>
	{/if}

	<Card.Root>
		<Card.Content class="p-0 overflow-x-auto">
			<Table.Root>
				<Table.Header>
					<Table.Row class="bg-slate-50">
						<Table.Head>Nama Pegawai</Table.Head>
						<Table.Head>NIP</Table.Head>
						<Table.Head class="text-center">Total Hari</Table.Head>
						<Table.Head class="text-center">Lengkap</Table.Head>
						<Table.Head class="text-center text-amber-600">Tanpa Pulang</Table.Head>
						<Table.Head class="text-center text-red-600">Tanpa Masuk</Table.Head>
						<Table.Head class="text-center">% Kehadiran</Table.Head>
					</Table.Row>
				</Table.Header>
				<Table.Body>
					{#each summary as r}
						{@const percent = r.total_days > 0 ? (r.complete_days / r.total_days) * 100 : 0}
						<Table.Row>
							<Table.Cell class="font-medium">{r.employee_nama}</Table.Cell>
							<Table.Cell class="font-mono text-xs text-slate-500">{r.employee_nip}</Table.Cell>
							<Table.Cell class="text-center">{r.total_days}</Table.Cell>
							<Table.Cell class="text-center font-semibold text-green-700">{r.complete_days}</Table.Cell>
							<Table.Cell class="text-center text-amber-600">{r.missing_checkout}</Table.Cell>
							<Table.Cell class="text-center text-red-600">{r.missing_checkin}</Table.Cell>
							<Table.Cell class="text-center">
								<div class="flex items-center justify-center gap-2">
									<div class="w-12 h-1.5 rounded-full bg-slate-100 overflow-hidden">
										<div class="h-full rounded-full bg-green-500" style="width:{Math.min(percent,100).toFixed(0)}%"></div>
									</div>
									<Badge variant={percent >= 80 ? 'default' : percent >= 50 ? 'outline' : 'destructive'}
										class={percent >= 80 ? 'bg-green-600' : ''}>
										{percent.toFixed(0)}%
									</Badge>
								</div>
							</Table.Cell>
						</Table.Row>
					{:else}
						<Table.Row>
							<Table.Cell colspan={7} class="py-12 text-center text-muted-foreground">
								{loading ? 'Memuat...' : 'Pilih rentang tanggal dan klik Tampilkan.'}
							</Table.Cell>
						</Table.Row>
					{/each}
				</Table.Body>
			</Table.Root>
		</Card.Content>
	</Card.Root>
</div>
