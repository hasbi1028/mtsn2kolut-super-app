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

	interface Book {
		id: string;
		kode: string;
		judul: string;
		pengarang: string;
		isbn: string;
		kategori: string;
		penerbit: string;
		tahun_terbit: number | null;
		total_eksemplar: number;
		tersedia: number;
		lokasi_rak: string;
	}

	const KATEGORI_LIST = ['pelajaran', 'fiksi', 'referensi', 'agama', 'umum'];

	let books = $state<Book[]>([]);
	let loading = $state(true);
	let search = $state('');
	let filterKategori = $state('');

	let showDialog = $state(false);
	let editingId = $state<string | null>(null);
	let busy = $state(false);
	let confirmDeleteId = $state<string | null>(null);
	let showDeleteDialog = $state(false);
	let formMode = $state<'beginner' | 'advance'>('beginner');

	let fKode = $state('');
	let fJudul = $state('');
	let fPengarang = $state('');
	let fIsbn = $state('');
	let fKategori = $state('umum');
	let fPenerbit = $state('');
	let fTahun = $state('');
	let fEksemplar = $state(1);
	let fRak = $state('');

	const filtered = $derived(
		books.filter((b) => {
			const q = search.toLowerCase();
			const matchSearch = !q || b.judul.toLowerCase().includes(q) || b.pengarang.toLowerCase().includes(q) || b.kode.toLowerCase().includes(q);
			const matchKat = !filterKategori || b.kategori === filterKategori;
			return matchSearch && matchKat;
		})
	);

	async function load() {
		loading = true;
		try {
			const res = await fetch('/api/library/books');
			const j = await res.json();
			books = j.data ?? j ?? [];
		} finally {
			loading = false;
		}
	}

	function openCreate() {
		editingId = null;
		fKode = ''; fJudul = ''; fPengarang = ''; fIsbn = '';
		fKategori = 'umum'; fPenerbit = ''; fTahun = '';
		fEksemplar = 1; fRak = '';
		formMode = 'beginner';
		showDialog = true;
	}

	function openEdit(b: Book) {
		editingId = b.id;
		fKode = b.kode; fJudul = b.judul; fPengarang = b.pengarang; fIsbn = b.isbn;
		fKategori = b.kategori; fPenerbit = b.penerbit; fTahun = b.tahun_terbit?.toString() ?? '';
		fEksemplar = b.total_eksemplar; fRak = b.lokasi_rak;
		formMode = 'advance';
		showDialog = true;
	}

	async function save() {
		if (!fKode.trim() || !fJudul.trim()) { toast.error('Kode dan judul wajib diisi'); return; }
		if (fEksemplar < 1) { toast.error('Jumlah eksemplar minimal 1'); return; }
		busy = true;
		try {
			const body = {
				kode: fKode.trim(), judul: fJudul.trim(), pengarang: fPengarang.trim(),
				isbn: fIsbn.trim(), kategori: fKategori, penerbit: fPenerbit.trim(),
				tahun_terbit: parseInt(fTahun) || 0,
				total_eksemplar: fEksemplar, lokasi_rak: fRak.trim(),
			};
			const res = editingId
				? await fetch(`/api/library/books/${editingId}`, { method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) })
				: await fetch('/api/library/books', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) });
			const j = await res.json();
			if (!res.ok) { toast.error(j.error ?? 'Gagal menyimpan'); return; }
			toast.success(editingId ? 'Buku diperbarui' : 'Buku ditambahkan');
			showDialog = false;
			await load();
		} finally {
			busy = false;
		}
	}

	async function deleteBook() {
		if (!confirmDeleteId) return;
		busy = true;
		try {
			const res = await fetch(`/api/library/books/${confirmDeleteId}`, { method: 'DELETE' });
			if (!res.ok) {
				const j = await res.json();
				toast.error(j.error ?? 'Gagal menghapus');
			} else {
				toast.success('Buku dihapus');
				confirmDeleteId = null;
				showDeleteDialog = false;
				await load();
			}
		} finally {
			busy = false;
		}
	}

	onMount(load);
</script>

<svelte:head><title>Katalog Buku — Perpustakaan</title></svelte:head>

<div class="space-y-4">
	<div class="flex flex-wrap items-center justify-between gap-3">
		<div>
			<h1 class="text-lg font-semibold text-slate-800">Katalog Buku</h1>
			<p class="text-sm text-slate-500">{books.length} judul terdaftar</p>
		</div>
		<Button onclick={openCreate} size="sm">+ Tambah Buku</Button>
	</div>

	<!-- Filters -->
	<div class="flex flex-wrap gap-2">
		<Input class="w-60" placeholder="Cari judul, pengarang, kode…" bind:value={search} />
		<select
			bind:value={filterKategori}
			class="rounded-md border border-input bg-background px-3 py-2 text-sm"
		>
			<option value="">Semua Kategori</option>
			{#each KATEGORI_LIST as k}
				<option value={k}>{k.charAt(0).toUpperCase() + k.slice(1)}</option>
			{/each}
		</select>
	</div>

	<Card.Root class="overflow-hidden border-slate-200 shadow-sm">
		<Card.Content class="p-0">
			{#if loading}
				<div class="space-y-3 p-6">
					{#each Array.from({ length: 6 }) as _, index (`book-skeleton-${index}`)}
						<div class="grid gap-3 md:grid-cols-[0.7fr_1.7fr_0.8fr_0.8fr_0.6fr_auto] md:items-center">
							<Skeleton class="h-5 w-20" />
							<Skeleton class="h-5 w-full max-w-sm" />
							<Skeleton class="h-6 w-20" />
							<Skeleton class="h-5 w-20 justify-self-center" />
							<Skeleton class="h-5 w-16" />
							<Skeleton class="h-9 w-28 justify-self-end" />
						</div>
					{/each}
				</div>
			{:else if filtered.length === 0}
				<p class="p-6 text-sm text-slate-400">Tidak ada buku ditemukan.</p>
			{:else}
				<Table.Root>
					<Table.Header>
						<Table.Row class="bg-slate-50 text-xs">
							<Table.Head>Kode</Table.Head>
							<Table.Head>Judul & Pengarang</Table.Head>
							<Table.Head>Kategori</Table.Head>
							<Table.Head class="text-center">Tersedia</Table.Head>
							<Table.Head>Rak</Table.Head>
							<Table.Head class="text-right">Aksi</Table.Head>
						</Table.Row>
					</Table.Header>
					<Table.Body>
						{#each filtered as b (b.id)}
							<Table.Row class="text-sm">
								<Table.Cell class="font-mono text-xs">{b.kode}</Table.Cell>
								<Table.Cell>
									<p class="font-medium text-slate-800">{b.judul}</p>
									{#if b.pengarang}<p class="text-xs text-slate-500">{b.pengarang}</p>{/if}
								</Table.Cell>
								<Table.Cell>
									<Badge variant="secondary" class="text-xs capitalize">{b.kategori}</Badge>
								</Table.Cell>
								<Table.Cell class="text-center">
									<span class={b.tersedia === 0 ? 'text-red-600 font-semibold' : 'text-green-700 font-semibold'}>
										{b.tersedia}
									</span>
									<span class="text-slate-400">/{b.total_eksemplar}</span>
								</Table.Cell>
								<Table.Cell class="text-xs text-slate-500">{b.lokasi_rak || '-'}</Table.Cell>
								<Table.Cell class="text-right">
									<Button variant="ghost" size="sm" onclick={() => openEdit(b)}>Edit</Button>
									<Button variant="ghost" size="sm" class="text-red-600 hover:text-red-700"
										onclick={() => { confirmDeleteId = b.id; showDeleteDialog = true; }}>Hapus</Button>
								</Table.Cell>
							</Table.Row>
						{/each}
					</Table.Body>
				</Table.Root>
			{/if}
		</Card.Content>
	</Card.Root>
</div>

<!-- Book form dialog -->
<Dialog.Root bind:open={showDialog}>
	<Dialog.Content>
		<Dialog.Header>
			<Dialog.Title>{editingId ? 'Edit Buku' : 'Tambah Buku Baru'}</Dialog.Title>
		</Dialog.Header>
		<!-- Mode toggle -->
		<div class="flex items-center gap-2 border-b border-slate-100 pb-3">
			<span class="text-xs text-slate-500">Mode:</span>
			{#each [['beginner', 'Cepat'], ['advance', 'Lengkap']] as [val, label]}
				<button
					class="rounded-full border px-3 py-1 text-xs transition-colors {formMode === val ? 'bg-green-700 text-white border-green-700' : 'border-slate-300 text-slate-600 hover:border-slate-400'}"
					onclick={() => (formMode = val as 'beginner' | 'advance')}
				>{label}</button>
			{/each}
			{#if formMode === 'beginner'}
				<span class="text-xs text-slate-400">— hanya field wajib</span>
			{:else}
				<span class="text-xs text-slate-400">— semua informasi buku</span>
			{/if}
		</div>

		<div class="space-y-3 py-2">
			<div class="grid grid-cols-2 gap-3">
				<div class="space-y-1">
					<label for="b-kode" class="block text-sm font-medium">Kode Buku <span class="text-red-500">*</span></label>
					<Input id="b-kode" bind:value={fKode} placeholder="BK-001" />
				</div>
				<div class="space-y-1">
					<label for="b-eks" class="block text-sm font-medium">Jumlah Eksemplar <span class="text-red-500">*</span></label>
					<Input id="b-eks" type="number" min="1" bind:value={fEksemplar} />
				</div>
			</div>
			<div class="space-y-1">
				<label for="b-judul" class="block text-sm font-medium">Judul <span class="text-red-500">*</span></label>
				<Input id="b-judul" bind:value={fJudul} placeholder="Judul buku" />
			</div>

			{#if formMode === 'advance'}
				<div class="grid grid-cols-2 gap-3">
					<div class="space-y-1">
						<label for="b-pengarang" class="block text-sm font-medium">Pengarang</label>
						<Input id="b-pengarang" bind:value={fPengarang} placeholder="Nama pengarang" />
					</div>
					<div class="space-y-1">
						<label for="b-isbn" class="block text-sm font-medium">ISBN</label>
						<Input id="b-isbn" bind:value={fIsbn} placeholder="ISBN (opsional)" />
					</div>
				</div>
				<div class="grid grid-cols-2 gap-3">
					<div class="space-y-1">
						<label for="b-kat" class="block text-sm font-medium">Kategori</label>
						<select id="b-kat" bind:value={fKategori} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
							{#each KATEGORI_LIST as k}
								<option value={k}>{k.charAt(0).toUpperCase() + k.slice(1)}</option>
							{/each}
						</select>
					</div>
					<div class="space-y-1">
						<label for="b-tahun" class="block text-sm font-medium">Tahun Terbit</label>
						<Input id="b-tahun" type="number" bind:value={fTahun} placeholder="2024" />
					</div>
				</div>
				<div class="grid grid-cols-2 gap-3">
					<div class="space-y-1">
						<label for="b-penerbit" class="block text-sm font-medium">Penerbit</label>
						<Input id="b-penerbit" bind:value={fPenerbit} placeholder="Nama penerbit" />
					</div>
					<div class="space-y-1">
						<label for="b-rak" class="block text-sm font-medium">Lokasi Rak</label>
						<Input id="b-rak" bind:value={fRak} placeholder="A-1, B-3, dst." />
					</div>
				</div>
			{/if}
		</div>
		<Dialog.Footer>
			<Button variant="outline" onclick={() => (showDialog = false)}>Batal</Button>
			<LoadingButton onclick={save} loading={busy} disabled={busy}>Simpan</LoadingButton>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>

<!-- Delete confirm dialog -->
<Dialog.Root bind:open={showDeleteDialog}>
	<Dialog.Content>
		<Dialog.Header>
			<Dialog.Title>Hapus Buku</Dialog.Title>
			<Dialog.Description>Buku yang masih dipinjam tidak dapat dihapus. Tindakan ini tidak dapat dibatalkan.</Dialog.Description>
		</Dialog.Header>
		<Dialog.Footer>
			<Button variant="outline" onclick={() => { showDeleteDialog = false; confirmDeleteId = null; }}>Batal</Button>
			<LoadingButton variant="destructive" onclick={deleteBook} loading={busy} disabled={busy}>Hapus</LoadingButton>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
