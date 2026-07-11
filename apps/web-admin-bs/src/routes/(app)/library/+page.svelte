<script lang="ts">
  import { onMount } from 'svelte';

  interface Book {
    id: string;
    kode: string;
    judul: string;
    pengarang: string;
    isbn?: string;
    kategori?: string;
    penerbit?: string;
    tahun_terbit?: number;
    total_eksemplar: number;
    tersedia: number;
    lokasi_rak?: string;
  }

  interface Stats {
    total_judul: number;
    total_eksemplar: number;
    total_tersedia: number;
    sedang_dipinjam: number;
    terlambat: number;
    denda_belum_lunas: number;
  }

  let books = $state<Book[]>([]);
  let stats = $state<Stats | null>(null);
  let loading = $state(true);
  let saving = $state(false);
  let error = $state('');
  let successMsg = $state('');
  let search = $state('');
  let showForm = $state(false);
  let editingId = $state<string | null>(null);

  let form = $state({
    kode: '',
    judul: '',
    pengarang: '',
    isbn: '',
    kategori: '',
    penerbit: '',
    tahun_terbit: '' as string | number,
    total_eksemplar: 1 as string | number,
    lokasi_rak: ''
  });

  onMount(async () => {
    await loadData();
  });

  async function loadData() {
    loading = true; error = '';
    try {
      const [booksRes, statsRes] = await Promise.all([
        fetch('/api/library/books'),
        fetch('/api/library/stats')
      ]);
      if (booksRes.ok) {
        const d = await booksRes.json();
        books = d.data || d.books || d || [];
      } else {
        error = `Gagal memuat buku (HTTP ${booksRes.status})`;
      }
      if (statsRes.ok) {
        const d = await statsRes.json();
        stats = d.data || d;
      }
    } catch (e: any) {
      error = 'Gagal terhubung ke server';
    } finally { loading = false; }
  }

  function openCreate() {
    editingId = null;
    form = { kode: '', judul: '', pengarang: '', isbn: '', kategori: '', penerbit: '', tahun_terbit: '', total_eksemplar: 1, lokasi_rak: '' };
    showForm = true;
  }

  function openEdit(b: Book) {
    editingId = b.id;
    form = {
      kode: b.kode,
      judul: b.judul,
      pengarang: b.pengarang,
      isbn: b.isbn || '',
      kategori: b.kategori || '',
      penerbit: b.penerbit || '',
      tahun_terbit: b.tahun_terbit || '',
      total_eksemplar: b.total_eksemplar,
      lokasi_rak: b.lokasi_rak || ''
    };
    showForm = true;
  }

  async function saveBook() {
    if (!form.kode || !form.judul || !form.pengarang) {
      error = 'Kode, judul, dan pengarang wajib diisi';
      return;
    }
    saving = true; error = ''; successMsg = '';
    try {
      const payload = {
        ...form,
        tahun_terbit: form.tahun_terbit ? Number(form.tahun_terbit) : 0,
        total_eksemplar: Number(form.total_eksemplar) || 1
      };
      const url = editingId ? `/api/library/books/${editingId}` : '/api/library/books';
      const method = editingId ? 'PUT' : 'POST';
      const res = await fetch(url, {
        method,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      });
      if (res.ok) {
        successMsg = editingId ? 'Buku diperbarui' : 'Buku ditambahkan';
        showForm = false;
        await loadData();
        setTimeout(() => { successMsg = ''; }, 2000);
      } else {
        const d = await res.json().catch(() => ({}));
        error = d.error || `Gagal (HTTP ${res.status})`;
      }
    } catch {
      error = 'Gagal terhubung ke server';
    } finally { saving = false; }
  }

  async function deleteBook(b: Book) {
    if (!confirm(`Hapus buku "${b.judul}"?`)) return;
    saving = true; error = '';
    try {
      const res = await fetch(`/api/library/books/${b.id}`, { method: 'DELETE' });
      if (res.ok) {
        successMsg = 'Buku dihapus';
        await loadData();
        setTimeout(() => { successMsg = ''; }, 2000);
      } else {
        error = 'Gagal menghapus buku';
      }
    } catch {
      error = 'Gagal terhubung ke server';
    } finally { saving = false; }
  }

  let filtered = $derived(search
    ? books.filter(b =>
        (b.judul || '').toLowerCase().includes(search.toLowerCase()) ||
        (b.kode || '').toLowerCase().includes(search.toLowerCase()) ||
        (b.pengarang || '').toLowerCase().includes(search.toLowerCase()))
    : books);
</script>

<svelte:head><title>Perpustakaan — MTsN 2 Kolut</title></svelte:head>

<div class="container-fluid px-0">
  <div class="d-flex align-items-center justify-content-between mb-1">
    <h4 class="fw-black mb-0">Perpustakaan</h4>
    <div class="d-flex gap-2">
      <a class="btn btn-outline-info btn-sm" href="/library/loans">
        <i class="bi bi-arrow-left-right me-1"></i>Peminjaman
      </a>
      <button class="btn btn-outline-success btn-sm" onclick={loadData} disabled={loading}>
        <i class="bi {loading ? 'bi-arrow-repeat spin' : 'bi-arrow-repeat'} me-1"></i>Refresh
      </button>
      <button class="btn btn-primary btn-sm" onclick={openCreate}>
        <i class="bi bi-plus-circle me-1"></i>Tambah Buku
      </button>
    </div>
  </div>
  <p class="text-secondary mb-3" style="font-size:0.85rem;">Daftar buku perpustakaan madrasah</p>

  {#if successMsg}
    <div class="alert alert-success py-2 small">{successMsg}</div>
  {/if}
  {#if error}
    <div class="alert alert-danger py-2 small">{error}</div>
  {/if}

  <!-- Stats -->
  {#if stats}
    <div class="row g-2 mb-3">
      <div class="col-auto">
        <span class="badge bg-primary fs-6 p-2">Judul: {stats.total_judul || 0}</span>
      </div>
      <div class="col-auto">
        <span class="badge bg-info fs-6 p-2">Eksemplar: {stats.total_eksemplar || 0}</span>
      </div>
      <div class="col-auto">
        <span class="badge bg-success fs-6 p-2">Tersedia: {stats.total_tersedia || 0}</span>
      </div>
      <div class="col-auto">
        <span class="badge bg-warning text-dark fs-6 p-2">Dipinjam: {stats.sedang_dipinjam || 0}</span>
      </div>
      <div class="col-auto">
        <span class="badge bg-danger fs-6 p-2">Terlambat: {stats.terlambat || 0}</span>
      </div>
    </div>
  {/if}

  <!-- Form Modal -->
  {#if showForm}
    <div class="card border shadow-sm mb-3">
      <div class="card-header bg-white d-flex justify-content-between align-items-center py-2">
        <span class="fw-bold small">{editingId ? 'Edit' : 'Tambah'} Buku</span>
        <button class="btn-close" aria-label="Tutup" onclick={() => { showForm = false; error = ''; }}></button>
      </div>
      <div class="card-body">
        <div class="row g-2">
          <div class="col-md-3">
            <label class="form-label small fw-semibold mb-1">Kode <span class="text-danger">*</span></label>
            <input class="form-control form-control-sm" bind:value={form.kode} placeholder="BK-001" />
          </div>
          <div class="col-md-5">
            <label class="form-label small fw-semibold mb-1">Judul <span class="text-danger">*</span></label>
            <input class="form-control form-control-sm" bind:value={form.judul} />
          </div>
          <div class="col-md-4">
            <label class="form-label small fw-semibold mb-1">Pengarang <span class="text-danger">*</span></label>
            <input class="form-control form-control-sm" bind:value={form.pengarang} />
          </div>
          <div class="col-md-3">
            <label class="form-label small fw-semibold mb-1">ISBN</label>
            <input class="form-control form-control-sm" bind:value={form.isbn} />
          </div>
          <div class="col-md-3">
            <label class="form-label small fw-semibold mb-1">Kategori</label>
            <input class="form-control form-control-sm" bind:value={form.kategori} placeholder="Fiksi / Pelajaran" />
          </div>
          <div class="col-md-3">
            <label class="form-label small fw-semibold mb-1">Penerbit</label>
            <input class="form-control form-control-sm" bind:value={form.penerbit} />
          </div>
          <div class="col-md-3">
            <label class="form-label small fw-semibold mb-1">Tahun Terbit</label>
            <input type="number" class="form-control form-control-sm" bind:value={form.tahun_terbit} />
          </div>
          <div class="col-md-6">
            <label class="form-label small fw-semibold mb-1">Jumlah Eksemplar</label>
            <input type="number" min="1" class="form-control form-control-sm" bind:value={form.total_eksemplar} />
          </div>
          <div class="col-md-6">
            <label class="form-label small fw-semibold mb-1">Lokasi Rak</label>
            <input class="form-control form-control-sm" bind:value={form.lokasi_rak} placeholder="A-1" />
          </div>
        </div>
        <div class="mt-3 d-flex gap-2">
          <button class="btn btn-primary btn-sm" onclick={saveBook} disabled={saving}>
            {#if saving}
              <span class="spinner-border spinner-border-sm me-1"></span>Menyimpan...
            {:else}
              <i class="bi bi-check-circle me-1"></i>Simpan
            {/if}
          </button>
          <button class="btn btn-outline-secondary btn-sm" onclick={() => { showForm = false; error = ''; }}>Batal</button>
        </div>
      </div>
    </div>
  {/if}

  <!-- Search -->
  <div class="mb-3">
    <div class="input-group input-group-sm" style="max-width:320px;">
      <span class="input-group-text"><i class="bi bi-search"></i></span>
      <input type="text" class="form-control" placeholder="Cari judul, kode, atau pengarang..." bind:value={search} />
    </div>
  </div>

  <!-- Books List -->
  {#if loading}
    <div class="text-center py-5">
      <div class="spinner-border text-success" role="status"></div>
      <p class="mt-2 text-secondary small">Memuat...</p>
    </div>
  {:else}
    <div class="card border shadow-sm">
      <div class="card-header bg-white py-2">
        <span class="fw-bold small">{filtered.length} buku</span>
      </div>
      <div class="card-body p-0">
        {#if filtered.length === 0}
          <div class="text-center py-4 text-secondary small">
            {#if search}
              Tidak ada buku dengan kata kunci "{search}"
            {:else}
              Belum ada data buku. Klik "Tambah Buku" untuk mulai.
            {/if}
          </div>
        {:else}
          <div class="table-responsive">
            <table class="table table-sm mb-0 small align-middle">
              <thead class="table-light">
                <tr>
                  <th>#</th>
                  <th>Kode</th>
                  <th>Judul</th>
                  <th>Pengarang</th>
                  <th>Kategori</th>
                  <th class="text-center">Eksemplar</th>
                  <th class="text-center">Tersedia</th>
                  <th>Rak</th>
                  <th></th>
                </tr>
              </thead>
              <tbody>
                {#each filtered as book, i (book.id)}
                  <tr>
                    <td class="text-muted">{i + 1}</td>
                    <td class="text-muted" style="font-family:monospace;font-size:0.75rem;">{book.kode || '—'}</td>
                    <td class="fw-semibold">{book.judul || '—'}</td>
                    <td>{book.pengarang || '—'}</td>
                    <td>{book.kategori || '—'}</td>
                    <td class="text-center">{book.total_eksemplar || 0}</td>
                    <td class="text-center">
                      <span class="badge {(book.tersedia || 0) > 0 ? 'bg-success' : 'bg-secondary'}">
                        {book.tersedia || 0}
                      </span>
                    </td>
                    <td class="text-muted">{book.lokasi_rak || '—'}</td>
                    <td>
                      <div class="btn-group btn-group-sm">
                        <button class="btn btn-outline-primary py-0 px-2" onclick={() => openEdit(book)} aria-label="Edit" title="Edit">
                          <i class="bi bi-pencil"></i>
                        </button>
                        <button class="btn btn-outline-danger py-0 px-2" onclick={() => deleteBook(book)} disabled={saving} aria-label="Hapus" title="Hapus">
                          <i class="bi bi-trash"></i>
                        </button>
                      </div>
                    </td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        {/if}
      </div>
    </div>
  {/if}
</div>
