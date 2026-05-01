<script lang="ts">
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { toast } from '$lib/components/ui/sonner';
	import { onMount } from 'svelte';
	import { readClientApiData, readClientJson } from '$lib/client/api';

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

	let dispositionsPromise = $state<Promise<DispositionRow[]> | null>(null);
	let dispositionsRequestId = 0;
	let filterStatus = $state('');
	let updateBusy = $state<Record<string, boolean>>({});
	let refreshBusy = $state(false);

	async function fetchDispositions(): Promise<DispositionRow[]> {
		const params = new URLSearchParams();
		if (filterStatus) params.set('status', filterStatus);
		const res = await fetch(`/api/tu/surat/disposisi?${params}`);
		return readClientApiData<DispositionRow[]>(res, 'Gagal memuat disposisi');
	}

	function loadDispositions() {
		const requestId = ++dispositionsRequestId;
		dispositionsPromise = fetchDispositions()
			.then((rows) => (requestId === dispositionsRequestId ? (rows ?? []) : []))
			.catch((error: unknown) => {
				if (requestId === dispositionsRequestId) throw error;
				return [];
			});
		return dispositionsPromise;
	}

	async function refreshDispositionsList() {
		refreshBusy = true;
		const promise = loadDispositions();
		const requestId = dispositionsRequestId;
		try {
			await promise;
		} catch {
			// AsyncContent owns the visible error state for the table body.
		} finally {
			if (requestId === dispositionsRequestId) refreshBusy = false;
		}
	}

	async function refreshDispositions() {
		const requestId = ++dispositionsRequestId;
		const rows = await fetchDispositions();
		if (requestId === dispositionsRequestId) {
			dispositionsPromise = Promise.resolve(rows ?? []);
		}
	}

	function retryDispositions(reset?: () => void) {
		reset?.();
		loadDispositions();
	}

	function dispositionsErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Gagal memuat disposisi';
	}

	function handleDispositionsRenderError(error: unknown) {
		console.error('TU dispositions render failed', error);
	}

	function mutationErrorMessage(error: unknown, fallback: string) {
		if (error instanceof Error && error.message.trim() && !error.message.toLowerCase().includes('fetch')) return error.message;
		return fallback;
	}

	async function refreshDispositionsAfterMutation() {
		try {
			await refreshDispositions();
		} catch (error) {
			toast.error(dispositionsErrorMessage(error));
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
			await readClientJson<unknown>(res);
			await refreshDispositionsAfterMutation();
			toast.success('Status disposisi diperbarui');
		} catch (error) {
			toast.error(mutationErrorMessage(error, 'Terjadi kesalahan jaringan'));
		} finally {
			updateBusy = { ...updateBusy, [disp.id]: false };
		}
	}

	function formatDate(raw: string) {
		if (!raw) return '-';
		const d = new Date(raw);
		return d.toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric' });
	}
</script>

<div class="container mx-auto max-w-7xl space-y-6 p-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-bold text-gray-900">Disposisi Surat</h1>
			<p class="text-sm text-gray-500">Daftar semua disposisi surat masuk</p>
		</div>
		<LoadingButton variant="outline" onclick={() => void refreshDispositionsList()} loading={refreshBusy} loadingLabel="Memuat...">Refresh</LoadingButton>
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

	<Card.Root>
		<Card.Content class="p-0">
			<AsyncContent promise={dispositionsPromise} onerror={handleDispositionsRenderError}>
				{#snippet pending()}
					<div class="space-y-2 p-4">
						{#each [1, 2, 3] as row (row)}
							<Skeleton class="h-12 w-full" />
						{/each}
					</div>
				{/snippet}

				{#snippet failed(error, reset)}
					<div class="p-4">
						<RecoveryPanel
							compact
							title="Disposisi Belum Tersaji"
							message={dispositionsErrorMessage(error)}
							onRetry={() => retryDispositions(reset)}
						/>
					</div>
				{/snippet}

				{#snippet children(value)}
					{@const currentDispositions = value as DispositionRow[]}
					{#if currentDispositions.length === 0}
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
								{#each currentDispositions as d, i (d.id)}
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
				{/snippet}
			</AsyncContent>
		</Card.Content>
	</Card.Root>
</div>
