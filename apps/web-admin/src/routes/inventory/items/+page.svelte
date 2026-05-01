<script lang="ts">
	import { onMount } from 'svelte';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import * as Dialog from '$lib/components/ui/dialog';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Badge } from '$lib/components/ui/badge';
	import { toast } from '$lib/components/ui/sonner';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import EmptyStatePanel from '$lib/components/EmptyStatePanel.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';

	interface Item {
		id: string;
		kode: string;
		nama: string;
		kategori: string;
		lokasi: string;
		kondisi: string;
		satuan: string;
		jumlah_total: number;
		jumlah_baik: number;
		min_stock: number;
		catatan: string;
	}

	interface ItemEvent {
		id: string;
		action: string;
		summary: string;
		actor_username: string;
		created_at: string;
	}

	const KATEGORI_LIST = ['umum', 'kelas', 'laboratorium', 'kantor', 'kebersihan', 'elektronik'];
	const KONDISI_LIST = ['baik', 'perlu-perawatan', 'rusak'];

	let items = $state<Item[]>([]);
	let loading = $state(true);
	let error = $state('');
	let search = $state('');
	let filterKategori = $state('');
	let filterKondisi = $state('');

	let showDialog = $state(false);
	let editingId = $state<string | null>(null);
	let busy = $state(false);
	let confirmDeleteId = $state<string | null>(null);
	let showDeleteDialog = $state(false);
	let showHistoryDialog = $state(false);
	let showBatchDialog = $state(false);
	let formMode = $state<'beginner' | 'advance'>('beginner');
	let historyLoading = $state(false);
	let historyItem = $state<Item | null>(null);
	let historyEvents = $state<ItemEvent[]>([]);
	let selectedIds = $state<string[]>([]);
	let batchBusy = $state(false);
	let batchLokasi = $state('');
	let batchKondisi = $state('');

	let fKode = $state('');
	let fNama = $state('');
	let fKategori = $state('umum');
	let fLokasi = $state('');
	let fKondisi = $state('baik');
	let fSatuan = $state('unit');
	let fJumlahTotal = $state(1);
	let fJumlahBaik = $state(1);
	let fMinStock = $state(0);
	let fCatatan = $state('');

	const filtered = $derived.by(() =>
		items.filter((item) => {
			const q = search.toLowerCase();
			const matchSearch =
				!q ||
				item.nama.toLowerCase().includes(q) ||
				item.kode.toLowerCase().includes(q) ||
				item.lokasi.toLowerCase().includes(q);
			const matchKategori = !filterKategori || item.kategori === filterKategori;
			const matchKondisi = !filterKondisi || item.kondisi === filterKondisi;
			return matchSearch && matchKategori && matchKondisi;
		})
	);

	const selectedItems = $derived.by(() => items.filter((item) => selectedIds.includes(item.id)));

	function csvEscape(value: string | number | null | undefined) {
		const text = String(value ?? '');
		if (text.includes('"') || text.includes(',') || text.includes('\n')) {
			return `"${text.replaceAll('"', '""')}"`;
		}
		return text;
	}

	function exportCsv() {
		const rows = [
			['Kode', 'Nama', 'Kategori', 'Lokasi', 'Kondisi', 'Satuan', 'Jumlah Total', 'Jumlah Baik', 'Min Stock', 'Catatan'],
			...filtered.map((item) => [
				item.kode,
				item.nama,
				item.kategori,
				item.lokasi,
				conditionLabel(item.kondisi),
				item.satuan,
				item.jumlah_total,
				item.jumlah_baik,
				item.min_stock,
				item.catatan,
			]),
		];
		const csv = rows.map((row) => row.map(csvEscape).join(',')).join('\n');
		const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' });
		const href = URL.createObjectURL(blob);
		const link = document.createElement('a');
		link.href = href;
		link.download = 'inventaris-filtered.csv';
		link.click();
		URL.revokeObjectURL(href);
		toast.success('Rekap inventaris berhasil diekspor');
	}

	function conditionLabel(value: string) {
		if (value === 'perlu-perawatan') return 'Perlu Perawatan';
		if (value === 'rusak') return 'Rusak';
		return 'Baik';
	}

	function isSelected(id: string) {
		return selectedIds.includes(id);
	}

	function toggleSelected(id: string, checked: boolean) {
		if (checked) {
			if (!selectedIds.includes(id)) selectedIds = [...selectedIds, id];
			return;
		}
		selectedIds = selectedIds.filter((value) => value !== id);
	}

	function toggleSelectAllVisible(checked: boolean) {
		if (!checked) {
			selectedIds = selectedIds.filter((id) => !filtered.some((item) => item.id === id));
			return;
		}
		selectedIds = Array.from(new Set([...selectedIds, ...filtered.map((item) => item.id)]));
	}

	function openBatchDialog() {
		if (selectedIds.length === 0) {
			toast.error('Pilih minimal satu barang untuk mutasi batch');
			return;
		}
		batchLokasi = '';
		batchKondisi = '';
		showBatchDialog = true;
	}

	function actionLabel(value: string) {
		if (value === 'create') return 'Dibuat';
		if (value === 'delete') return 'Dihapus';
		return 'Diperbarui';
	}

	function formatDateTime(value: string) {
		if (!value) return '—';
		return new Date(value).toLocaleString('id-ID', {
			timeZone: 'Asia/Makassar',
			day: '2-digit',
			month: 'short',
			year: 'numeric',
			hour: '2-digit',
			minute: '2-digit',
		}) + ' WITA';
	}

	async function load() {
		loading = true;
		try {
			error = '';
			const res = await fetch('/api/inventory/items');
			const j = await res.json();
			items = j.data ?? j ?? [];
		} catch {
			error = 'Gagal memuat daftar inventaris. Coba lagi untuk mengambil data barang terbaru.';
		} finally {
			loading = false;
		}
	}

	function openCreate() {
		editingId = null;
		fKode = '';
		fNama = '';
		fKategori = 'umum';
		fLokasi = '';
		fKondisi = 'baik';
		fSatuan = 'unit';
		fJumlahTotal = 1;
		fJumlahBaik = 1;
		fMinStock = 0;
		fCatatan = '';
		formMode = 'beginner';
		showDialog = true;
	}

	function openEdit(item: Item) {
		editingId = item.id;
		fKode = item.kode;
		fNama = item.nama;
		fKategori = item.kategori;
		fLokasi = item.lokasi;
		fKondisi = item.kondisi;
		fSatuan = item.satuan;
		fJumlahTotal = item.jumlah_total;
		fJumlahBaik = item.jumlah_baik;
		fMinStock = item.min_stock;
		fCatatan = item.catatan;
		formMode = 'advance';
		showDialog = true;
	}

	async function save() {
		if (!fKode.trim() || !fNama.trim()) {
			toast.error('Kode dan nama barang wajib diisi');
			return;
		}
		if (fJumlahTotal < 1) {
			toast.error('Jumlah total minimal 1');
			return;
		}
		if (fJumlahBaik < 0 || fJumlahBaik > fJumlahTotal) {
			toast.error('Jumlah kondisi baik harus antara 0 dan jumlah total');
			return;
		}

		busy = true;
		try {
			const body = {
				kode: fKode.trim(),
				nama: fNama.trim(),
				kategori: fKategori,
				lokasi: fLokasi.trim(),
				kondisi: fKondisi,
				satuan: fSatuan.trim(),
				jumlah_total: fJumlahTotal,
				jumlah_baik: fJumlahBaik,
				min_stock: fMinStock,
				catatan: fCatatan.trim(),
			};
			const res = editingId
				? await fetch(`/api/inventory/items/${editingId}`, { method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) })
				: await fetch('/api/inventory/items', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) });
			const j = await res.json();
			if (!res.ok) {
				toast.error(j.error ?? 'Gagal menyimpan barang');
				return;
			}
			toast.success(editingId ? 'Barang diperbarui' : 'Barang ditambahkan');
			showDialog = false;
			await load();
		} finally {
			busy = false;
		}
	}

	async function deleteItem() {
		if (!confirmDeleteId) return;
		busy = true;
		try {
			const res = await fetch(`/api/inventory/items/${confirmDeleteId}`, { method: 'DELETE' });
			if (!res.ok) {
				const j = await res.json();
				toast.error(j.error ?? 'Gagal menghapus barang');
				return;
			}
			toast.success('Barang dihapus');
			confirmDeleteId = null;
			showDeleteDialog = false;
			await load();
		} finally {
			busy = false;
		}
	}

	async function submitBatchUpdate() {
		if (selectedIds.length === 0) {
			toast.error('Pilih minimal satu barang untuk dimutasi');
			return;
		}
		const lokasi = batchLokasi.trim();
		const kondisi = batchKondisi.trim();
		if (!lokasi && !kondisi) {
			toast.error('Isi lokasi baru atau pilih kondisi baru');
			return;
		}

		batchBusy = true;
		try {
			const res = await fetch('/api/inventory/items', {
				method: 'PATCH',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					item_ids: selectedIds,
					lokasi: lokasi || undefined,
					kondisi: kondisi || undefined,
				}),
			});
			const j = await res.json();
			if (!res.ok) {
				toast.error(j.error ?? 'Gagal memproses mutasi batch');
				return;
			}
			toast.success(`${j.data?.updated ?? selectedIds.length} barang berhasil diperbarui`);
			showBatchDialog = false;
			selectedIds = [];
			await load();
		} finally {
			batchBusy = false;
		}
	}

	async function openHistory(item: Item) {
		historyItem = item;
		historyEvents = [];
		historyLoading = true;
		showHistoryDialog = true;
		try {
			const res = await fetch(`/api/inventory/items/${item.id}/history`);
			const j = await res.json();
			historyEvents = j.data ?? j ?? [];
		} catch {
			toast.error('Gagal memuat riwayat inventaris');
		} finally {
			historyLoading = false;
		}
	}

	onMount(load);
</script>

<svelte:head><title>Daftar Barang — Inventaris</title></svelte:head>

<div class="space-y-4">
	<div class="flex flex-wrap items-start justify-between gap-4">
		<div>
			<h1 class="text-2xl font-semibold text-slate-800">Daftar Barang</h1>
			<p class="text-sm text-slate-500">Kelola master inventaris sekolah, lokasi penyimpanan, kondisi, dan batas restok.</p>
		</div>
		<div class="flex gap-2">
			<Button variant="outline" onclick={exportCsv} size="sm" disabled={filtered.length === 0}>Ekspor CSV</Button>
			<Button variant="outline" onclick={openBatchDialog} size="sm" disabled={selectedIds.length === 0}>Mutasi Batch</Button>
			<Button onclick={openCreate} size="sm">+ Tambah Barang</Button>
		</div>
	</div>

	{#if selectedIds.length > 0}
		<Card.Root class="border-sky-200 bg-sky-50 shadow-sm">
			<Card.Content class="flex flex-wrap items-center justify-between gap-3 p-4">
				<div>
					<p class="text-xs font-semibold uppercase tracking-[0.18em] text-sky-700">Seleksi Batch Aktif</p>
					<p class="mt-1 text-sm text-slate-700">{selectedIds.length} barang dipilih untuk mutasi lokasi atau kondisi.</p>
				</div>
				<div class="flex gap-2">
					<Button variant="outline" size="sm" onclick={() => (selectedIds = [])}>Kosongkan Pilihan</Button>
					<Button size="sm" onclick={openBatchDialog}>Lanjut Mutasi Batch</Button>
				</div>
			</Card.Content>
		</Card.Root>
	{/if}

	<div class="grid gap-3 md:grid-cols-3">
		<div class="rounded-2xl border border-emerald-100 bg-emerald-50 px-4 py-4">
			<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-emerald-700">Total Jenis</p>
			<p class="mt-2 text-2xl font-semibold text-slate-900">{items.length}</p>
			<p class="text-sm text-slate-600">barang inventaris yang sudah tercatat</p>
		</div>
		<div class="rounded-2xl border border-sky-100 bg-sky-50 px-4 py-4">
			<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-sky-700">Hasil Filter</p>
			<p class="mt-2 text-2xl font-semibold text-slate-900">{filtered.length}</p>
			<p class="text-sm text-slate-600">barang yang cocok dengan filter aktif</p>
		</div>
		<div class="rounded-2xl border border-amber-100 bg-amber-50 px-4 py-4">
			<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-amber-700">Perlu Restok</p>
			<p class="mt-2 text-2xl font-semibold text-slate-900">{items.filter((item) => item.jumlah_baik <= item.min_stock).length}</p>
			<p class="text-sm text-slate-600">barang yang sudah menyentuh batas minimum</p>
		</div>
	</div>

	{#if error}
		<RecoveryPanel title="Daftar Inventaris Belum Tersaji" message={error} onRetry={load} />
	{/if}

	<Card.Root class="border-slate-200 shadow-sm">
		<Card.Content class="grid gap-3 p-4 md:grid-cols-3">
			<div>
				<p class="mb-1 text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">Cari Barang</p>
				<Input class="w-full" placeholder="Cari nama, kode, atau lokasi…" bind:value={search} />
			</div>
			<div>
				<p class="mb-1 text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">Filter Kategori</p>
				<select bind:value={filterKategori} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
					<option value="">Semua Kategori</option>
					{#each KATEGORI_LIST as kategori (kategori)}
						<option value={kategori}>{kategori.charAt(0).toUpperCase() + kategori.slice(1)}</option>
					{/each}
				</select>
			</div>
			<div>
				<p class="mb-1 text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">Filter Kondisi</p>
				<select bind:value={filterKondisi} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
					<option value="">Semua Kondisi</option>
					{#each KONDISI_LIST as kondisi (kondisi)}
						<option value={kondisi}>{conditionLabel(kondisi)}</option>
					{/each}
				</select>
			</div>
		</Card.Content>
	</Card.Root>

	<Card.Root class="overflow-hidden border-slate-200 shadow-sm">
		<Card.Content class="p-0">
			{#if loading}
				<div class="space-y-3 p-6">
					{#each Array.from({ length: 6 }) as _, index (`inventory-item-skeleton-${index}`)}
						<div class="grid gap-3 md:grid-cols-[0.8fr_1.5fr_0.8fr_0.8fr_0.8fr_auto] md:items-center">
							<Skeleton class="h-5 w-20" />
							<Skeleton class="h-5 w-full max-w-sm" />
							<Skeleton class="h-6 w-24" />
							<Skeleton class="h-5 w-20" />
							<Skeleton class="h-5 w-20" />
							<Skeleton class="h-9 w-28 justify-self-end" />
						</div>
					{/each}
				</div>
			{:else if filtered.length === 0}
				<div class="p-4">
					<EmptyStatePanel
						title={search || filterKategori || filterKondisi ? 'Tidak ada barang yang cocok' : 'Daftar inventaris masih kosong'}
						description={search || filterKategori || filterKondisi
							? 'Ubah kata kunci, kategori, atau kondisi untuk melihat barang lain yang sudah tercatat.'
							: 'Tambahkan barang inventaris pertama agar sekolah bisa mulai memantau stok, lokasi, dan kondisi.'}
						compact
					/>
				</div>
			{:else}
				<Table.Root>
					<Table.Header>
						<Table.Row class="bg-slate-50 text-xs uppercase tracking-wide text-slate-500">
							<Table.Head class="w-12">
								<input
									type="checkbox"
									class="h-4 w-4 rounded border-slate-300"
									checked={filtered.length > 0 && filtered.every((item) => selectedIds.includes(item.id))}
									onchange={(event) => toggleSelectAllVisible((event.currentTarget as HTMLInputElement).checked)}
								/>
							</Table.Head>
							<Table.Head>Kode</Table.Head>
							<Table.Head>Barang</Table.Head>
							<Table.Head>Kondisi</Table.Head>
							<Table.Head>Stok</Table.Head>
							<Table.Head>Lokasi</Table.Head>
							<Table.Head class="text-right">Aksi</Table.Head>
						</Table.Row>
					</Table.Header>
					<Table.Body>
						{#each filtered as item (item.id)}
							<Table.Row class="text-sm">
								<Table.Cell>
									<input
										type="checkbox"
										class="h-4 w-4 rounded border-slate-300"
										checked={isSelected(item.id)}
										onchange={(event) => toggleSelected(item.id, (event.currentTarget as HTMLInputElement).checked)}
									/>
								</Table.Cell>
								<Table.Cell class="font-mono text-xs text-slate-600">{item.kode}</Table.Cell>
								<Table.Cell>
									<p class="font-medium text-slate-900">{item.nama}</p>
									<p class="text-xs text-slate-500">{item.kategori} · {item.satuan}</p>
								</Table.Cell>
								<Table.Cell>
									<Badge variant={item.kondisi === 'rusak' ? 'destructive' : 'outline'}>
										{conditionLabel(item.kondisi)}
									</Badge>
								</Table.Cell>
								<Table.Cell>
									<p class="font-medium text-slate-900">{item.jumlah_baik} / {item.jumlah_total}</p>
									<p class="text-xs text-slate-500">min. {item.min_stock}</p>
								</Table.Cell>
								<Table.Cell>{item.lokasi || '—'}</Table.Cell>
								<Table.Cell class="text-right">
									<div class="flex justify-end gap-2">
										<Button variant="outline" size="sm" onclick={() => openHistory(item)}>Riwayat</Button>
										<Button variant="outline" size="sm" onclick={() => openEdit(item)}>Edit</Button>
										<Button variant="destructive" size="sm" onclick={() => { confirmDeleteId = item.id; showDeleteDialog = true; }}>Hapus</Button>
									</div>
								</Table.Cell>
							</Table.Row>
						{/each}
					</Table.Body>
				</Table.Root>
			{/if}
		</Card.Content>
	</Card.Root>
</div>

<Dialog.Root bind:open={showDialog}>
	{#if showDialog}
		<Dialog.Content>
			<Dialog.Header>
				<Dialog.Title>{editingId ? 'Edit Barang Inventaris' : 'Tambah Barang Inventaris'}</Dialog.Title>
				<Dialog.Description>
					{editingId ? 'Perbarui data lokasi, kondisi, dan stok inventaris.' : 'Catat barang inventaris baru agar stok dan lokasinya bisa dipantau.'}
				</Dialog.Description>
			</Dialog.Header>

			<div class="space-y-4">
				<div class="grid gap-2 md:grid-cols-2">
					<Button type="button" variant={formMode === 'beginner' ? 'default' : 'outline'} onclick={() => (formMode = 'beginner')}>Mode Dasar</Button>
					<Button type="button" variant={formMode === 'advance' ? 'default' : 'outline'} onclick={() => (formMode = 'advance')}>Mode Lanjutan</Button>
				</div>

				<div class="grid gap-4 md:grid-cols-2">
					<div>
						<label for="kode" class="mb-1 block text-xs font-medium text-slate-600">Kode</label>
						<Input id="kode" bind:value={fKode} placeholder="INV-001" />
					</div>
					<div>
						<label for="nama" class="mb-1 block text-xs font-medium text-slate-600">Nama Barang</label>
						<Input id="nama" bind:value={fNama} placeholder="Proyektor Kelas 8A" />
					</div>
				</div>

				<div class="grid gap-4 md:grid-cols-2">
					<div>
						<label for="kategori" class="mb-1 block text-xs font-medium text-slate-600">Kategori</label>
						<select id="kategori" bind:value={fKategori} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
							{#each KATEGORI_LIST as kategori (kategori)}
								<option value={kategori}>{kategori.charAt(0).toUpperCase() + kategori.slice(1)}</option>
							{/each}
						</select>
					</div>
					<div>
						<label for="lokasi" class="mb-1 block text-xs font-medium text-slate-600">Lokasi</label>
						<Input id="lokasi" bind:value={fLokasi} placeholder="Gudang Utama / Kelas / Lab" />
					</div>
				</div>

				{#if formMode === 'advance'}
					<div class="grid gap-4 md:grid-cols-3">
						<div>
							<label for="kondisi" class="mb-1 block text-xs font-medium text-slate-600">Kondisi</label>
							<select id="kondisi" bind:value={fKondisi} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
								{#each KONDISI_LIST as kondisi (kondisi)}
									<option value={kondisi}>{conditionLabel(kondisi)}</option>
								{/each}
							</select>
						</div>
						<div>
							<label for="satuan" class="mb-1 block text-xs font-medium text-slate-600">Satuan</label>
							<Input id="satuan" bind:value={fSatuan} placeholder="unit / pcs / set" />
						</div>
						<div>
							<label for="min-stock" class="mb-1 block text-xs font-medium text-slate-600">Batas Restok</label>
							<Input id="min-stock" type="number" bind:value={fMinStock} min="0" />
						</div>
					</div>
				{/if}

				<div class="grid gap-4 md:grid-cols-2">
					<div>
						<label for="jumlah-total" class="mb-1 block text-xs font-medium text-slate-600">Jumlah Total</label>
						<Input id="jumlah-total" type="number" bind:value={fJumlahTotal} min="1" />
					</div>
					<div>
						<label for="jumlah-baik" class="mb-1 block text-xs font-medium text-slate-600">Jumlah Kondisi Baik</label>
						<Input id="jumlah-baik" type="number" bind:value={fJumlahBaik} min="0" max={fJumlahTotal} />
					</div>
				</div>

				{#if formMode === 'advance'}
					<div>
						<label for="catatan" class="mb-1 block text-xs font-medium text-slate-600">Catatan</label>
						<Input id="catatan" bind:value={fCatatan} placeholder="Mis. butuh servis lampu atau dipakai bersama lintas kelas" />
					</div>
				{/if}

				<div class="flex justify-end gap-2">
					<Button variant="outline" onclick={() => (showDialog = false)}>Batal</Button>
					<LoadingButton onclick={save} loading={busy} loadingLabel="Menyimpan..." label={editingId ? 'Simpan Perubahan' : 'Tambah Barang'} />
				</div>
			</div>
		</Dialog.Content>
	{/if}
</Dialog.Root>

<Dialog.Root bind:open={showBatchDialog}>
	{#if showBatchDialog}
		<Dialog.Content>
			<Dialog.Header>
				<Dialog.Title>Mutasi Inventaris Batch</Dialog.Title>
				<Dialog.Description>
					Perbarui lokasi atau kondisi untuk {selectedIds.length} barang terpilih sekaligus. Setiap barang tetap akan mencatat riwayat perubahan masing-masing.
				</Dialog.Description>
			</Dialog.Header>

			<div class="space-y-4">
				<div class="rounded-xl border border-slate-200 bg-slate-50 p-3">
					<p class="text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">Barang Terpilih</p>
					<p class="mt-1 text-sm text-slate-700">
						{selectedItems.slice(0, 4).map((item) => item.nama).join(', ')}
						{#if selectedItems.length > 4}
							, dan {selectedItems.length - 4} barang lainnya
						{/if}
					</p>
				</div>

				<div class="grid gap-4 md:grid-cols-2">
					<div>
						<label for="batch-lokasi" class="mb-1 block text-xs font-medium text-slate-600">Lokasi Baru</label>
						<Input id="batch-lokasi" bind:value={batchLokasi} placeholder="Kosongkan jika lokasi tidak diubah" />
					</div>
					<div>
						<label for="batch-kondisi" class="mb-1 block text-xs font-medium text-slate-600">Kondisi Baru</label>
						<select id="batch-kondisi" bind:value={batchKondisi} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
							<option value="">Tidak diubah</option>
							{#each KONDISI_LIST as kondisi (kondisi)}
								<option value={kondisi}>{conditionLabel(kondisi)}</option>
							{/each}
						</select>
					</div>
				</div>

				<div class="rounded-xl border border-amber-200 bg-amber-50 p-3 text-sm text-slate-700">
					Mutasi batch ini cocok untuk perpindahan antarruang atau penyamaan status kondisi setelah pengecekan lapangan. Jika hanya satu barang yang berubah, tetap lebih aman memakai tombol edit per barang.
				</div>

				<div class="flex justify-end gap-2">
					<Button variant="outline" onclick={() => (showBatchDialog = false)}>Batal</Button>
					<LoadingButton onclick={submitBatchUpdate} loading={batchBusy} loadingLabel="Memproses..." label="Terapkan Mutasi Batch" />
				</div>
			</div>
		</Dialog.Content>
	{/if}
</Dialog.Root>

<Dialog.Root bind:open={showHistoryDialog}>
	{#if showHistoryDialog}
		<Dialog.Content>
			<Dialog.Header>
				<Dialog.Title>Riwayat Inventaris</Dialog.Title>
				<Dialog.Description>
					{#if historyItem}
						{historyItem.nama} · {historyItem.kode}
					{/if}
				</Dialog.Description>
			</Dialog.Header>

			<div class="space-y-3">
				{#if historyLoading}
					<div class="space-y-3">
						{#each Array.from({ length: 4 }) as _, index (`inventory-history-skeleton-${index}`)}
							<div class="rounded-xl border border-slate-200 p-3">
								<Skeleton class="h-4 w-24" />
								<Skeleton class="mt-2 h-4 w-full" />
								<Skeleton class="mt-2 h-3 w-36" />
							</div>
						{/each}
					</div>
				{:else if historyEvents.length === 0}
					<EmptyStatePanel compact title="Belum ada riwayat" description="Riwayat perubahan akan muncul setelah barang ini dibuat, diperbarui, atau dihapus." />
				{:else}
					<div class="max-h-[420px] space-y-3 overflow-y-auto pr-1">
						{#each historyEvents as event (event.id)}
							<div class="rounded-xl border border-slate-200 bg-slate-50 p-3">
								<div class="flex items-start justify-between gap-3">
									<Badge variant="outline">{actionLabel(event.action)}</Badge>
									<p class="text-xs text-slate-500">{formatDateTime(event.created_at)}</p>
								</div>
								<p class="mt-2 text-sm text-slate-700">{event.summary}</p>
								<p class="mt-2 text-xs text-slate-500">
									{event.actor_username ? `oleh ${event.actor_username}` : 'oleh sistem'}
								</p>
							</div>
						{/each}
					</div>
				{/if}
			</div>
		</Dialog.Content>
	{/if}
</Dialog.Root>

<Dialog.Root bind:open={showDeleteDialog}>
	{#if showDeleteDialog}
		<Dialog.Content>
			<Dialog.Header>
				<Dialog.Title>Hapus Barang</Dialog.Title>
				<Dialog.Description>Barang inventaris yang dihapus akan keluar dari daftar pemantauan stok dan kondisi.</Dialog.Description>
			</Dialog.Header>
			<div class="flex justify-end gap-2">
				<Button variant="outline" onclick={() => (showDeleteDialog = false)}>Batal</Button>
				<LoadingButton onclick={deleteItem} loading={busy} loadingLabel="Menghapus..." label="Hapus" variant="destructive" />
			</div>
		</Dialog.Content>
	{/if}
</Dialog.Root>
