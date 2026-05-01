<script lang="ts">
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import { toast } from '$lib/components/ui/sonner';
	import { onMount } from 'svelte';

	type DispositionRow = {
		id: string;
		incoming_letter_id: string;
		assignee_name: string;
		instruksi: string;
		catatan_tindak_lanjut: string;
		status: string;
		disposed_by_name: string;
		disposed_at: string;
		completed_at: string;
		nomor_agenda: string;
		letter_perihal: string;
		letter_asal: string;
	};

	let loading = $state(true);
	let error = $state('');
	let dispositions = $state<DispositionRow[]>([]);
	let filterStatus = $state('');
	let updateBusy = $state<Record<string, boolean>>({});

	async function loadDispositions() {
		loading = true;
		error = '';
		try {
			const params = new URLSearchParams();
			if (filterStatus) params.set('status', filterStatus);
			const res = await fetch(`/api/tu/surat/disposisi?${params}`);
			if (!res.ok) throw new Error((await res.json()).error ?? `HTTP ${res.status}`);
			dispositions = (await res.json()) as DispositionRow[];
		} catch (e) {
			error = e instanceof Error ? e.message : 'Gagal memuat disposisi';
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		loadDispositions();
	});

	async function updateStatus(disp: DispositionRow, newStatus: string) {
		if (disp.status === newStatus) return;
		updateBusy = { ...updateBusy, [disp.id]: true };
		try {
			const res = await fetch(`/api/tu/surat/disposisi/${disp.id}`, {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					instruksi: disp.instruksi,
					catatan_tindak_lanjut: disp.catatan_tindak_lanjut,
					status: newStatus
				})
			});
			if (!res.ok) { toast.error((await res.json()).error ?? 'Gagal mengubah status'); return; }
			await loadDispositions();
			toast.success('Status disposisi diperbarui');
		} catch {
			toast.error('Terjadi kesalahan jaringan');
		} finally {
			updateBusy = { ...updateBusy, [disp.id]: false };
		}
	}

	function formatDate(raw: string) {
		if (!raw) return '-';
		const d = new Date(raw);
		return d.toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric' });
	}

	const statusColors: Record<string, string> = {
		terkirim: 'bg-sky-100 text-sky-800 hover:bg-sky-100',
		dibaca: 'bg-purple-100 text-purple-800 hover:bg-purple-100',
		ditindaklanjuti: 'bg-yellow-100 text-yellow-800 hover:bg-yellow-100',
		selesai: 'bg-green-100 text-green-800 hover:bg-green-100'
	};
</script>

<div class="container mx-auto max-w-7xl space-y-6 p-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-bold text-gray-900">Disposisi Surat</h1>
			<p class="text-sm text-gray-500">Daftar semua disposisi surat masuk</p>
		</div>
		<Button variant="outline" onclick={() => loadDispositions()}>Refresh</Button>
	</div>

	<!-- Filter -->
	<Card.Root>
		<Card.Content class="pt-4">
			<div class="flex items-center gap-3">
				<label for="filter-status" class="text-sm font-medium text-gray-700 shrink-0">Status</label>
				<select
					id="filter-status"
					bind:value={filterStatus}
					onchange={() => loadDispositions()}
					class="rounded-md border border-gray-300 bg-white px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-green-600"
				>
					<option value="">Semua Status</option>
					<option value="terkirim">Terkirim</option>
					<option value="dibaca">Dibaca</option>
					<option value="ditindaklanjuti">Ditindaklanjuti</option>
					<option value="selesai">Selesai</option>
				</select>
			</div>
		</Card.Content>
	</Card.Root>

	{#if error}
		<div class="rounded-md border border-red-200 bg-red-50 p-4 text-sm text-red-700">{error}</div>
	{/if}

	<Card.Root>
		<Card.Content class="p-0">
			{#if loading}
				<div class="space-y-2 p-4">
					{#each [1, 2, 3] as _}
						<Skeleton class="h-12 w-full" />
					{/each}
				</div>
			{:else if dispositions.length === 0}
				<div class="p-8 text-center text-sm text-gray-500">
					{filterStatus ? 'Tidak ada disposisi dengan status ini.' : 'Belum ada disposisi.'}
				</div>
			{:else}
				<Table.Root>
					<Table.Header>
						<Table.Row>
							<Table.Head class="w-12">No</Table.Head>
							<Table.Head class="w-28">No. Agenda</Table.Head>
							<Table.Head>Perihal Surat</Table.Head>
							<Table.Head>Penerima</Table.Head>
							<Table.Head>Instruksi</Table.Head>
							<Table.Head class="w-28">Tanggal</Table.Head>
							<Table.Head class="w-36">Status</Table.Head>
						</Table.Row>
					</Table.Header>
					<Table.Body>
						{#each dispositions as d, i}
							<Table.Row>
								<Table.Cell class="text-gray-500">{i + 1}</Table.Cell>
								<Table.Cell class="font-mono text-xs font-medium text-gray-700">{d.nomor_agenda}</Table.Cell>
								<Table.Cell>
									<p class="text-sm font-medium text-gray-800 max-w-xs truncate">{d.letter_perihal}</p>
									<p class="text-xs text-gray-400">{d.letter_asal}</p>
								</Table.Cell>
								<Table.Cell class="text-sm text-gray-700">{d.assignee_name || '–'}</Table.Cell>
								<Table.Cell class="max-w-xs truncate text-sm text-gray-600">
									{#if d.instruksi}{d.instruksi}{:else}<span class="italic text-gray-400">–</span>{/if}
								</Table.Cell>
								<Table.Cell class="text-sm text-gray-600">{formatDate(d.disposed_at)}</Table.Cell>
								<Table.Cell>
									<select
										value={d.status}
										onchange={(e) => updateStatus(d, (e.target as HTMLSelectElement).value)}
										disabled={updateBusy[d.id]}
										class="rounded border border-gray-200 px-2 py-1 text-xs focus:outline-none focus:ring-1 focus:ring-green-600"
									>
										<option value="terkirim">Terkirim</option>
										<option value="dibaca">Dibaca</option>
										<option value="ditindaklanjuti">Ditindaklanjuti</option>
										<option value="selesai">Selesai</option>
									</select>
								</Table.Cell>
							</Table.Row>
						{/each}
					</Table.Body>
				</Table.Root>
			{/if}
		</Card.Content>
	</Card.Root>
</div>
