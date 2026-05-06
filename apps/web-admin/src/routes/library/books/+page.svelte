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
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import EmptyStatePanel from '$lib/components/EmptyStatePanel.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { readClientApiData, readClientJson } from '$lib/client/api';

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
	let booksPromise = $state<Promise<Book[]> | null>(null);
	let booksRequestId = 0;
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

	function filterBooks(bookRows: Book[]) {
		return bookRows.filter((b) => {
			const q = search.toLowerCase();
			const matchSearch = !q || b.judul.toLowerCase().includes(q) || b.pengarang.toLowerCase().includes(q) || b.kode.toLowerCase().includes(q);
			const matchKat = !filterKategori || b.kategori === filterKategori;
			return matchSearch && matchKat;
		});
	}

	async function fetchBooks(): Promise<Book[]> {
		const res = await fetch('/api/library/books');
		return readClientApiData<Book[]>(res, 'Gagal memuat katalog buku.');
	}

	function load() {
		const requestId = ++booksRequestId;
		books = [];
		booksPromise = fetchBooks()
			.then((nextBooks) => {
				if (requestId === booksRequestId) {
					books = nextBooks ?? [];
					return books;
				}
				return books;
			})
			.catch((error: unknown) => {
				if (requestId === booksRequestId) throw error;
				return books;
			});
	}

	async function refreshBooks() {
		const requestId = ++booksRequestId;
		const nextBooks = await fetchBooks();
		if (requestId === booksRequestId) {
			books = nextBooks ?? [];
			booksPromise = Promise.resolve(books);
		}
	}

	function retryBooks(reset?: () => void) {
		reset?.();
		load();
	}

	function booksErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Gagal memuat katalog buku. Coba lagi untuk mengambil daftar buku terbaru.';
	}

	function mutationErrorMessage(error: unknown, fallback: string) {
		if (error instanceof Error && error.message.trim() && !error.message.toLowerCase().includes('fetch')) return error.message;
		return fallback;
	}

	async function refreshBooksAfterMutation() {
		try {
			await refreshBooks();
		} catch (error) {
			booksPromise = Promise.resolve(books);
			toast.error(booksErrorMessage(error));
		}
	}

	function handleBooksRenderError(error: unknown) {
		console.error('Library books render failed', error);
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
			await readClientJson<unknown>(res);
			toast.success(editingId ? 'Buku diperbarui' : 'Buku ditambahkan');
			showDialog = false;
			await refreshBooksAfterMutation();
		} catch (error) {
			toast.error(mutationErrorMessage(error, 'Gagal menyimpan katalog buku. Periksa koneksi lalu coba lagi.'));
		} finally {
			busy = false;
		}
	}

	async function deleteBook() {
		if (!confirmDeleteId) return;
		busy = true;
		try {
			const res = await fetch(`/api/library/books/${confirmDeleteId}`, { method: 'DELETE' });
			await readClientJson<unknown>(res);
			toast.success('Buku dihapus');
			confirmDeleteId = null;
			showDeleteDialog = false;
			await refreshBooksAfterMutation();
		} catch (error) {
			toast.error(mutationErrorMessage(error, 'Gagal menghapus katalog buku. Periksa koneksi lalu coba lagi.'));
		} finally {
			busy = false;
		}
	}

	onMount(() => {
		void load();
	});
</script>

<svelte:head><title>Katalog Buku — Perpustakaan</title></svelte:head>

<div class="space-y-4">
	<div class="flex flex-wrap items-start justify-between gap-4">
		<div>
			<h1 class="text-2xl font-semibold text-foreground">Katalog Buku</h1>
			<p class="text-sm text-muted-foreground">Kelola koleksi inti perpustakaan, stok eksemplar, dan klasifikasi rak.</p>
		</div>
		<Button onclick={openCreate} size="sm">+ Tambah Buku</Button>
	</div>

	<AsyncContent promise={booksPromise} onerror={handleBooksRenderError}>
		{#snippet pending()}
			<div class="grid gap-3 md:grid-cols-3">
				{#each ['Total Judul', 'Tersedia Dicari', 'Stok Habis'] as label (label)}
					<div class="rounded-2xl border border-border bg-muted/50 px-4 py-4">
						<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-muted-foreground">{label}</p>
						<Skeleton class="mt-3 h-8 w-16" />
						<Skeleton class="mt-2 h-4 w-48" />
					</div>
				{/each}
			</div>
		{/snippet}

		{#snippet failed(error, reset)}
			<RecoveryPanel
				title="Katalog Belum Tersaji"
				message={booksErrorMessage(error)}
				onRetry={() => retryBooks(reset)}
			/>
		{/snippet}

		{#snippet children(value)}
			{@const currentBooks = value as Book[]}
			{@const currentFiltered = filterBooks(currentBooks)}
			<div class="grid gap-3 md:grid-cols-3">
				<div class="rounded-2xl border border-primary/20 bg-primary/10 px-4 py-4">
					<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-primary">Total Judul</p>
					<p class="mt-2 text-2xl font-semibold text-foreground">{currentBooks.length}</p>
					<p class="text-sm text-muted-foreground">koleksi unik yang tercatat di katalog</p>
				</div>
				<div class="rounded-2xl border border-accent bg-accent/60 px-4 py-4">
					<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-accent-foreground">Tersedia Dicari</p>
					<p class="mt-2 text-2xl font-semibold text-foreground">{currentFiltered.length}</p>
					<p class="text-sm text-muted-foreground">hasil koleksi berdasarkan filter saat ini</p>
				</div>
				<div class="rounded-2xl border border-warning/30 bg-warning/10 px-4 py-4">
					<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-warning">Stok Habis</p>
					<p class="mt-2 text-2xl font-semibold text-foreground">{currentBooks.filter((item) => item.tersedia === 0).length}</p>
					<p class="text-sm text-muted-foreground">judul yang butuh penambahan atau pengembalian</p>
				</div>
			</div>
		{/snippet}
	</AsyncContent>

	<Card.Root class="border-border shadow-sm">
		<Card.Content class="grid gap-3 p-4 md:grid-cols-[1.2fr_0.8fr]">
			<div>
				<p class="mb-1 text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">Cari Koleksi</p>
				<Input class="w-full" placeholder="Cari judul, pengarang, atau kode buku…" bind:value={search} />
			</div>
			<div>
				<p class="mb-1 text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">Filter Kategori</p>
				<select
					bind:value={filterKategori}
					class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
				>
					<option value="">Semua Kategori</option>
					{#each KATEGORI_LIST as k (k)}
						<option value={k}>{k.charAt(0).toUpperCase() + k.slice(1)}</option>
					{/each}
				</select>
			</div>
		</Card.Content>
	</Card.Root>

	<Card.Root class="overflow-hidden border-border shadow-sm">
		<Card.Content class="p-0">
			<AsyncContent promise={booksPromise} onerror={handleBooksRenderError}>
				{#snippet pending()}
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
				{/snippet}

				{#snippet failed(error, reset)}
					<div class="p-4">
						<RecoveryPanel
							compact
							title="Katalog Belum Tersaji"
							message={booksErrorMessage(error)}
							onRetry={() => retryBooks(reset)}
						/>
					</div>
				{/snippet}

				{#snippet children(value)}
					{@const currentFiltered = filterBooks(value as Book[])}
					{#if currentFiltered.length === 0}
						<div class="p-4">
							<EmptyStatePanel
								title={search || filterKategori ? 'Tidak ada buku yang cocok' : 'Katalog buku masih kosong'}
								description={search || filterKategori
									? 'Ubah kata kunci atau kategori untuk melihat koleksi lain yang sudah ada.'
									: 'Tambahkan judul buku pertama agar perpustakaan bisa mulai dipakai untuk proses peminjaman.'}
								compact
							/>
						</div>
					{:else}
						<Table.Root>
							<Table.Header>
								<Table.Row class="bg-muted/50 text-xs">
									<Table.Head>Kode</Table.Head>
									<Table.Head>Judul & Pengarang</Table.Head>
									<Table.Head>Kategori</Table.Head>
									<Table.Head class="text-center">Tersedia</Table.Head>
									<Table.Head>Rak</Table.Head>
									<Table.Head class="text-right">Aksi</Table.Head>
								</Table.Row>
							</Table.Header>
							<Table.Body>
								{#each currentFiltered as b (b.id)}
									<Table.Row class="text-sm">
										<Table.Cell class="font-mono text-xs">{b.kode}</Table.Cell>
										<Table.Cell>
											<p class="font-medium text-foreground">{b.judul}</p>
											{#if b.pengarang}<p class="text-xs text-muted-foreground">{b.pengarang}</p>{/if}
										</Table.Cell>
										<Table.Cell>
											<Badge variant="secondary" class="text-xs capitalize">{b.kategori}</Badge>
										</Table.Cell>
										<Table.Cell class="text-center">
											<span class={b.tersedia === 0 ? 'text-destructive font-semibold' : 'text-success font-semibold'}>
												{b.tersedia}
											</span>
											<span class="text-muted-foreground">/{b.total_eksemplar}</span>
										</Table.Cell>
										<Table.Cell class="text-xs text-muted-foreground">{b.lokasi_rak || '-'}</Table.Cell>
										<Table.Cell class="text-right">
											<Button variant="ghost" size="sm" onclick={() => openEdit(b)}>Edit</Button>
											<Button variant="ghost" size="sm" class="text-destructive hover:text-destructive"
												onclick={() => { confirmDeleteId = b.id; showDeleteDialog = true; }}>Hapus</Button>
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

<!-- Book form dialog -->
<Dialog.Root bind:open={showDialog}>
	<Dialog.Content>
		<Dialog.Header>
			<Dialog.Title>{editingId ? 'Edit Buku' : 'Tambah Buku Baru'}</Dialog.Title>
		</Dialog.Header>
		<!-- Mode toggle -->
		<div class="flex items-center gap-2 border-b border-border pb-3">
			<span class="text-xs text-muted-foreground">Mode:</span>
			{#each [['beginner', 'Cepat'], ['advance', 'Lengkap']] as [val, label] (val)}
				<button
					class="rounded-full border px-3 py-1 text-xs transition-colors {formMode === val ? 'bg-success text-background border-success' : 'border-border text-muted-foreground hover:border-border'}"
					onclick={() => (formMode = val as 'beginner' | 'advance')}
				>{label}</button>
			{/each}
			{#if formMode === 'beginner'}
				<span class="text-xs text-muted-foreground">— hanya field wajib</span>
			{:else}
				<span class="text-xs text-muted-foreground">— semua informasi buku</span>
			{/if}
		</div>

		<div class="space-y-3 py-2">
			<div class="grid grid-cols-2 gap-3">
				<div class="space-y-1">
					<label for="b-kode" class="block text-sm font-medium">Kode Buku <span class="text-destructive">*</span></label>
					<Input id="b-kode" bind:value={fKode} placeholder="BK-001" />
				</div>
				<div class="space-y-1">
					<label for="b-eks" class="block text-sm font-medium">Jumlah Eksemplar <span class="text-destructive">*</span></label>
					<Input id="b-eks" type="number" min="1" bind:value={fEksemplar} />
				</div>
			</div>
			<div class="space-y-1">
				<label for="b-judul" class="block text-sm font-medium">Judul <span class="text-destructive">*</span></label>
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
							{#each KATEGORI_LIST as k (k)}
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
			<LoadingButton onclick={() => void save()} loading={busy} disabled={busy}>Simpan</LoadingButton>
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
			<LoadingButton variant="destructive" onclick={() => void deleteBook()} loading={busy} disabled={busy}>Hapus</LoadingButton>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
