<script lang="ts">
  import { Badge } from '$lib/components/ui/badge';
  import { Button } from '$lib/components/ui/button';
  import * as Card from '$lib/components/ui/card';
  import * as Dialog from '$lib/components/ui/dialog';
  import { Input } from '$lib/components/ui/input';
  import { toast } from '$lib/components/ui/sonner';
  import LoadingButton from '$lib/components/LoadingButton.svelte';
  import { page } from '$app/state';
  import { enhance } from '$app/forms';
  import { goto } from '$app/navigation';

  let { data } = $props();

  type Murid = {
    id: string; nis: string; nisn: string; nama: string; gender: string;
    class_name: string; class_code: string; is_active: boolean; status: string;
    nik: string; tempat_lahir: string; tanggal_lahir: string | null; alamat: string;
    agama: string; phone: string; parent_name: string; parent_phone: string;
    total_violation_points: number;
  };

  type Rombel = { id: string; code: string; name: string; level: string };

  let murid = $state<Murid[]>(data.murid ?? []);
  let rombels = $state<Rombel[]>(data.rombels ?? []);
  let total = $state(data.total ?? 0);
  let pages = $state(data.pages ?? 0);
  let currentPage = $state(data.page ?? 1);
  let perPage = $state(data.perPage ?? 25);
  let loading = $state(false);

  // — Search / Filter
  let search = $state('');
  let filterStatus = $state('');

  // — Tambah dialog
  let showTambah = $state(false);
  let tambahLoading = $state(false);
  let tambahForm = $state({
    nis: '', nisn: '', nama: '', gender: 'L', class_id: '',
    status: 'active', nik: '', tempat_lahir: '', tanggal_lahir: '',
    alamat: '', agama: '', phone: '', parent_name: '', parent_phone: '',
  });

  // — Edit dialog
  let showEdit = $state(false);
  let editId = $state('');
  let editForm = $state({ nik: '', tempat_lahir: '', tanggal_lahir: '', alamat: '', agama: '', anak_ke: '', phone: '', parent_name: '', parent_phone: '' });
  let editLoading = $state(false);

  // — Delete confirmation
  let showDelete = $state(false);
  let deleteId = $state('');
  let deleteName = $state('');
  let deleteLoading = $state(false);

  function filtered(): Murid[] {
    let f = murid;
    if (search) { const s = search.toLowerCase(); f = f.filter(m => m.nama.toLowerCase().includes(s) || m.nis.toLowerCase().includes(s) || m.nisn.toLowerCase().includes(s) || m.nik.toLowerCase().includes(s)); }
    if (filterStatus) { f = f.filter(m => m.status === filterStatus); }
    return f;
  }

  async function submitTambah() {
    if (!tambahForm.nis || !tambahForm.nama) { toast.error('NIS dan Nama wajib diisi'); return; }
    tambahLoading = true;
    try {
      const r = await fetch('/api/kesiswaan/murid', {
        method: 'POST', headers: { 'content-type': 'application/json' },
        body: JSON.stringify(tambahForm),
      });
      if (r.ok) {
        const p = await r.json();
        toast.success(`Murid ${tambahForm.nama} tersimpan`);
        showTambah = false;
        resetTambahForm();
        await refresh();
      } else { const e = await r.json().catch(() => ({})); toast.error(e?.error || 'Gagal'); }
    } catch { toast.error('Gagal'); }
    finally { tambahLoading = false; }
  }

  function resetTambahForm() {
    tambahForm = { nis: '', nisn: '', nama: '', gender: 'L', class_id: '', status: 'active', nik: '', tempat_lahir: '', tanggal_lahir: '', alamat: '', agama: '', phone: '', parent_name: '', parent_phone: '' };
  }

  function openEdit(m: Murid) {
    editId = m.id;
    editForm = {
      nik: m.nik, tempat_lahir: m.tempat_lahir,
      tanggal_lahir: m.tanggal_lahir ?? '',
      alamat: m.alamat, agama: m.agama,
      anak_ke: '', phone: m.phone,
      parent_name: m.parent_name, parent_phone: m.parent_phone,
    };
    showEdit = true;
  }

  async function submitEdit() {
    if (!editId) return;
    editLoading = true;
    try {
      const r = await fetch(`/api/kesiswaan/murid/${editId}/profile`, {
        method: 'PUT', headers: { 'content-type': 'application/json' },
        body: JSON.stringify({
          nik: editForm.nik, tempat_lahir: editForm.tempat_lahir,
          tanggal_lahir: editForm.tanggal_lahir || null,
          alamat: editForm.alamat, agama: editForm.agama,
          anak_ke: parseInt(editForm.anak_ke) || 0,
          phone: editForm.phone, parent_name: editForm.parent_name,
          parent_phone: editForm.parent_phone,
        }),
      });
      if (r.ok) { toast.success('Profil murid diupdate'); showEdit = false; await refresh(); }
      else { const e = await r.json().catch(() => ({})); toast.error(e?.error || 'Gagal'); }
    } catch { toast.error('Gagal'); }
    finally { editLoading = false; }
  }

  function confirmDelete(m: Murid) {
    deleteId = m.id;
    deleteName = m.nama;
    showDelete = true;
  }

  async function submitDelete() {
    if (!deleteId) return;
    deleteLoading = true;
    try {
      const r = await fetch(`/api/kesiswaan/murid/${deleteId}`, { method: 'DELETE' });
      if (r.ok) { toast.success(`Murid ${deleteName} dihapus`); showDelete = false; await refresh(); }
      else { const e = await r.json().catch(() => ({})); toast.error(e?.error || 'Gagal'); }
    } catch { toast.error('Gagal'); }
    finally { deleteLoading = false; }
  }

  async function refresh() {
    loading = true;
    try {
      const r = await fetch('/api/kesiswaan/murid');
      if (r.ok) { const p = await r.json(); murid = p.data ?? []; }
    } catch {}
    finally { loading = false; }
  }

  function statusColor(s: string): string {
    switch (s) {
      case 'active': return 'bg-primary/10 text-primary';
      case 'alumni': return 'bg-muted text-muted-foreground';
      case 'mutasi': return 'bg-orange-50 text-orange-700';
      case 'dropout': return 'bg-red-50 text-red-700';
      default: return 'bg-muted/30 text-muted-foreground';
    }
  }

  function genderLabel(g: string): string {
    return g === 'L' ? 'Lk' : g === 'P' ? 'Pr' : g;
  }

  function pageRange(): number[] {
    const range: number[] = [];
    const start = Math.max(1, currentPage - 2);
    const end = Math.min(pages, currentPage + 2);
    for (let i = start; i <= end; i++) range.push(i);
    return range;
  }

  function goToPage(p: number) {
    if (p < 1 || p > pages) return;
    currentPage = p;
    goto(`?page=${p}&per_page=${perPage}`, { keepFocus: true });
  }
</script>

<svelte:head><title>Data Murid — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-4 max-w-screen-xl mx-auto">
  <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2">
    <div>
      <h1 class="text-xl font-black">Data Murid</h1>
      <p class="text-sm text-muted-foreground">Data pokok peserta didik madrasah.</p>
    </div>
    <Button onclick={() => showTambah = true} size="sm">+ Tambah Murid</Button>
  </div>

  <!-- Search & Filter -->
  <div class="flex flex-wrap items-center gap-2">
    <div class="relative flex-1 min-w-[200px] max-w-sm">
      <svg class="absolute left-2.5 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-muted-foreground" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
      <input type="text" placeholder="Cari nama, NIS, NISN, NIK..." bind:value={search} class="w-full h-9 pl-8 pr-3 rounded-lg border border-input bg-background text-sm" />
    </div>
    <select bind:value={filterStatus} class="h-9 rounded-lg border border-input bg-background px-3 text-sm">
      <option value="">Semua Status</option>
      <option value="active">Aktif</option>
      <option value="alumni">Alumni</option>
      <option value="mutasi">Mutasi</option>
      <option value="dropout">Dropout</option>
    </select>
    <span class="text-xs text-muted-foreground">{total} murid</span>
  </div>

  {#if filtered().length === 0}
    <Card.Root><Card.Content class="p-8 text-center"><p class="text-sm text-muted-foreground">{#if murid.length === 0}Belum ada data murid. Klik "+ Tambah Murid" untuk menambahkan.{:else}Tidak ditemukan.{/if}</p></Card.Content></Card.Root>
  {:else}
    <!-- Desktop: Table -->
    <div class="hidden md:block overflow-x-auto rounded-lg border border-border">
      <table class="w-full text-sm">
        <thead><tr class="bg-muted/30 text-muted-foreground text-xs uppercase"><th class="px-3 py-2 text-left">Nama</th><th class="px-3 py-2 text-left">NIS</th><th class="px-3 py-2 text-left hidden sm:table-cell">Kelas</th><th class="px-3 py-2 text-center hidden lg:table-cell">L/P</th><th class="px-3 py-2 text-left hidden lg:table-cell">NIK</th><th class="px-3 py-2 text-center">Status</th><th class="px-3 py-2 text-right">Aksi</th></tr></thead>
        <tbody class="divide-y divide-border">
          {#each filtered() as m (m.id)}<tr class="hover:bg-muted/10">
            <td class="px-3 py-2 font-medium text-xs md:text-sm whitespace-nowrap">{m.nama}</td>
            <td class="px-3 py-2 text-xs text-muted-foreground">{m.nis}</td>
            <td class="px-3 py-2 text-xs text-muted-foreground hidden sm:table-cell">{m.class_code || '—'}</td>
            <td class="px-3 py-2 text-center text-xs hidden lg:table-cell">{genderLabel(m.gender)}</td>
            <td class="px-3 py-2 text-xs text-muted-foreground hidden lg:table-cell">{m.nik || '—'}</td>
            <td class="px-3 py-2 text-center"><span class="inline-flex items-center rounded-full px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wider {statusColor(m.status)}">{m.status}</span></td>
            <td class="px-3 py-2 text-right whitespace-nowrap space-x-1">
              <button class="text-xs text-primary hover:underline" onclick={() => openEdit(m)}>Edit</button>
              <button class="text-xs text-red-500 hover:underline" onclick={() => confirmDelete(m)}>Hapus</button>
            </td>
          </tr>{/each}
        </tbody>
      </table>
    </div>

    <!-- Mobile: Card list -->
    <div class="md:hidden space-y-2">
      {#each filtered() as m (m.id)}
        <div class="rounded-xl border border-border bg-base-100 shadow-sm p-3 space-y-1">
          <div class="flex items-center justify-between gap-2">
            <span class="text-sm font-semibold">{m.nama}</span>
            <span class="inline-flex items-center rounded-full px-2 py-0.5 text-[9px] font-semibold uppercase tracking-wider {statusColor(m.status)}">{m.status}</span>
          </div>
          <div class="flex flex-wrap gap-x-3 gap-y-0.5 text-xs text-muted-foreground">
            <span>NIS: {m.nis}</span>
            <span>Kelas: {m.class_code || '—'}</span>
            <span>{genderLabel(m.gender)}</span>
          </div>
          <div class="pt-1 flex gap-2">
            <button class="text-xs text-primary hover:underline" onclick={() => openEdit(m)}>Edit Profil</button>
            <button class="text-xs text-red-500 hover:underline" onclick={() => confirmDelete(m)}>Hapus</button>
          </div>
        </div>
      {/each}
    </div>
  {/if}

  <!-- Pagination -->
  {#if pages > 1}
    <div class="flex items-center justify-center gap-1 pt-4 pb-2">
      <button class="px-2.5 py-1 text-xs font-medium rounded-md transition-colors {currentPage <= 1 ? 'opacity-40 cursor-not-allowed' : 'hover:bg-muted/20 border border-border'}" disabled={currentPage <= 1} onclick={() => goToPage(currentPage - 1)}>
        ‹ Prev
      </button>
      {#each pageRange() as p}
        <button class="min-w-[28px] px-2 py-1 text-xs font-medium rounded-md transition-colors {p === currentPage ? 'bg-primary text-primary-foreground' : 'hover:bg-muted/20 border border-border'}" onclick={() => goToPage(p)}>
          {p}
        </button>
      {/each}
      <button class="px-2.5 py-1 text-xs font-medium rounded-md transition-colors {currentPage >= pages ? 'opacity-40 cursor-not-allowed' : 'hover:bg-muted/20 border border-border'}" disabled={currentPage >= pages} onclick={() => goToPage(currentPage + 1)}>
        Next ›
      </button>
    </div>
    <p class="text-center text-[10px] text-muted-foreground">Halaman {currentPage} dari {pages} ({total} murid)</p>
  {/if}
</div>

<!-- Dialog Tambah Murid -->
<Dialog.Root bind:open={showTambah}>
  <Dialog.Content class="max-w-xl">
    <div class="space-y-4">
      <h2 class="text-base font-semibold">Tambah Murid Baru</h2>
      <div class="grid gap-3 sm:grid-cols-2">
        <div class="space-y-1"><label class="text-xs font-medium text-muted-foreground">NIS *</label><Input bind:value={tambahForm.nis} placeholder="Nomor Induk Siswa" /></div>
        <div class="space-y-1"><label class="text-xs font-medium text-muted-foreground">NISN</label><Input bind:value={tambahForm.nisn} placeholder="Nomor Induk Nasional" /></div>
        <div class="space-y-1 sm:col-span-2"><label class="text-xs font-medium text-muted-foreground">Nama Lengkap *</label><Input bind:value={tambahForm.nama} placeholder="Nama sesuai ijazah" /></div>
        <div class="space-y-1"><label class="text-xs font-medium text-muted-foreground">Jenis Kelamin</label>
          <select bind:value={tambahForm.gender} class="flex h-9 w-full rounded-lg border border-input bg-background px-3 text-sm">
            <option value="L">Laki-laki</option>
            <option value="P">Perempuan</option>
          </select>
        </div>
        <div class="space-y-1"><label class="text-xs font-medium text-muted-foreground">Rombel</label>
          <select bind:value={tambahForm.class_id} class="flex h-9 w-full rounded-lg border border-input bg-background px-3 text-sm">
            <option value="">— Pilih Rombel —</option>
            {#each rombels as r}
              <option value={r.id}>{r.code} — {r.name}</option>
            {/each}
          </select>
        </div>
        <div class="space-y-1"><label class="text-xs font-medium text-muted-foreground">Status</label>
          <select bind:value={tambahForm.status} class="flex h-9 w-full rounded-lg border border-input bg-background px-3 text-sm">
            <option value="active">Aktif</option>
            <option value="alumni">Alumni</option>
            <option value="mutasi">Mutasi</option>
            <option value="dropout">Dropout</option>
          </select>
        </div>
        <div class="space-y-1"><label class="text-xs font-medium text-muted-foreground">NIK</label><Input bind:value={tambahForm.nik} placeholder="NIK KTP" /></div>
        <div class="space-y-1"><label class="text-xs font-medium text-muted-foreground">Tempat Lahir</label><Input bind:value={tambahForm.tempat_lahir} placeholder="Kota lahir" /></div>
        <div class="space-y-1"><label class="text-xs font-medium text-muted-foreground">Tanggal Lahir</label><Input type="date" bind:value={tambahForm.tanggal_lahir} /></div>
        <div class="space-y-1"><label class="text-xs font-medium text-muted-foreground">Agama</label>
          <select bind:value={tambahForm.agama} class="flex h-9 w-full rounded-lg border border-input bg-background px-3 text-sm">
            <option value="">— Pilih Agama —</option>
            <option value="Islam">Islam</option>
            <option value="Kristen">Kristen</option>
            <option value="Katolik">Katolik</option>
            <option value="Hindu">Hindu</option>
            <option value="Buddha">Buddha</option>
            <option value="Khonghucu">Khonghucu</option>
          </select>
        </div>
        <div class="space-y-1"><label class="text-xs font-medium text-muted-foreground">No. HP</label><Input bind:value={tambahForm.phone} placeholder="08xxxx" /></div>
        <div class="space-y-1 sm:col-span-2"><label class="text-xs font-medium text-muted-foreground">Alamat</label>
          <textarea bind:value={tambahForm.alamat} placeholder="Alamat lengkap" class="flex w-full rounded-lg border border-input bg-background px-3 py-2 text-sm min-h-[50px]"></textarea>
        </div>
        <div class="space-y-1"><label class="text-xs font-medium text-muted-foreground">Nama Orang Tua</label><Input bind:value={tambahForm.parent_name} placeholder="Nama wali" /></div>
        <div class="space-y-1"><label class="text-xs font-medium text-muted-foreground">No. HP Orang Tua</label><Input bind:value={tambahForm.parent_phone} placeholder="08xxxx" /></div>
      </div>
      <div class="flex justify-end gap-2 pt-2">
        <Button variant="outline" onclick={() => (showTambah = false)}>Batal</Button>
        <LoadingButton onclick={() => void submitTambah()} loading={tambahLoading}>Simpan</LoadingButton>
      </div>
    </div>
  </Dialog.Content>
</Dialog.Root>

<!-- Dialog Edit Murid -->
<Dialog.Root bind:open={showEdit}>
  <Dialog.Content>
    <div class="space-y-4">
      <h2 class="text-base font-semibold">Edit Profil Murid</h2>
      <div class="grid gap-3 sm:grid-cols-2">
        <div class="space-y-1"><label class="text-xs font-medium text-muted-foreground">NIK</label><Input bind:value={editForm.nik} placeholder="NIK" /></div>
        <div class="space-y-1"><label class="text-xs font-medium text-muted-foreground">Tempat Lahir</label><Input bind:value={editForm.tempat_lahir} placeholder="Kolaka" /></div>
        <div class="space-y-1"><label class="text-xs font-medium text-muted-foreground">Tanggal Lahir</label><Input type="date" bind:value={editForm.tanggal_lahir} /></div>
        <div class="space-y-1"><label class="text-xs font-medium text-muted-foreground">Agama</label><Input bind:value={editForm.agama} placeholder="Islam" /></div>
        <div class="space-y-1"><label class="text-xs font-medium text-muted-foreground">Anak Ke-</label><Input type="number" bind:value={editForm.anak_ke} min="0" /></div>
        <div class="space-y-1"><label class="text-xs font-medium text-muted-foreground">No. HP</label><Input bind:value={editForm.phone} placeholder="08xxxx" /></div>
        <div class="space-y-1 sm:col-span-2"><label class="text-xs font-medium text-muted-foreground">Alamat</label><Input bind:value={editForm.alamat} placeholder="Alamat lengkap" /></div>
        <div class="space-y-1"><label class="text-xs font-medium text-muted-foreground">Nama Orang Tua</label><Input bind:value={editForm.parent_name} placeholder="Nama wali" /></div>
        <div class="space-y-1"><label class="text-xs font-medium text-muted-foreground">No. HP Orang Tua</label><Input bind:value={editForm.parent_phone} placeholder="08xxxx" /></div>
      </div>
      <div class="flex justify-end gap-2 pt-2">
        <Button variant="outline" onclick={() => (showEdit = false)}>Batal</Button>
        <LoadingButton onclick={() => void submitEdit()} loading={editLoading}>Simpan</LoadingButton>
      </div>
    </div>
  </Dialog.Content>
</Dialog.Root>

<!-- Dialog Hapus Murid -->
<Dialog.Root bind:open={showDelete}>
  <Dialog.Content>
    <div class="space-y-4">
      <h2 class="text-base font-semibold text-red-600">Hapus Murid</h2>
      <p class="text-sm text-muted-foreground">Yakin ingin menghapus <strong>{deleteName}</strong>? Data yang dihapus tidak dapat dikembalikan.</p>
      <div class="flex justify-end gap-2 pt-2">
        <Button variant="outline" onclick={() => (showDelete = false)}>Batal</Button>
        <LoadingButton variant="destructive" onclick={() => void submitDelete()} loading={deleteLoading}>Ya, Hapus</LoadingButton>
      </div>
    </div>
  </Dialog.Content>
</Dialog.Root>
