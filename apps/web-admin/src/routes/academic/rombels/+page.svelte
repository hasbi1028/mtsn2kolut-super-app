<script lang="ts">
  import { Badge } from '$lib/components/ui/badge';
  import { Button } from '$lib/components/ui/button';
  import * as Card from '$lib/components/ui/card';
  import * as Dialog from '$lib/components/ui/dialog';
  import { Input } from '$lib/components/ui/input';
  import { toast } from '$lib/components/ui/sonner';
  import LoadingButton from '$lib/components/LoadingButton.svelte';
  import { page } from '$app/state';

  let { data } = $props();

  let roles = $derived(page.data.user?.roles ?? (page.data.user?.role ? [page.data.user.role] : []));
  let isAdmin = $derived(roles.includes('admin'));

  type Rombel = {
    id: string; code: string; name: string; level: string;
    is_active: boolean; academic_year_id: string; academic_year_name: string;
  };

  let rombels = $state<Rombel[]>(data.items ?? []);
  let activeSemester = $state(data.activeSemester);
  let showTambah = $state(false);
  let tambahForm = $state({ code: '', name: '', level: 'VII' });
  let tambahLoading = $state(false);
  let deleteLoading = $state<string | null>(null);

  // ── Edit Rombel ──
  let showEdit = $state(false);
  let editRombel = $state<Rombel | null>(null);
  let editForm = $state({ code: '', name: '', level: 'VII', is_active: true });
  let editLoading = $state(false);

  const levels = ['VII', 'VIII', 'IX'];

  async function refresh() {
    const res = await fetch('/api/academic/rombels');
    if (res.ok) {
      const payload = await res.json();
      rombels = payload.items ?? [];
    }
  }

  async function submitTambah() {
    if (!tambahForm.code || !tambahForm.name) {
      toast.error('Kode dan nama rombel wajib diisi');
      return;
    }
    tambahLoading = true;
    try {
      const res = await fetch('/api/academic/rombels', {
        method: 'POST',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify({
          academic_year_id: activeSemester?.academic_year_id ?? '',
          code: tambahForm.code,
          name: tambahForm.name,
          level: tambahForm.level,
          is_active: true,
        }),
      });
      if (res.ok) {
        toast.success(`Rombel ${tambahForm.code} berhasil dibuat`);
        showTambah = false;
        tambahForm = { code: '', name: '', level: 'VII' };
        await refresh();
      } else {
        const body = await res.json().catch(() => ({}));
        toast.error(body?.error || 'Gagal membuat rombel');
      }
    } catch { toast.error('Gagal membuat rombel'); }
    finally { tambahLoading = false; }
  }

  function openEdit(rombel: Rombel) {
    editRombel = rombel;
    editForm = { code: rombel.code, name: rombel.name, level: rombel.level, is_active: rombel.is_active };
    showEdit = true;
  }

  async function submitEdit() {
    if (!editRombel || !editForm.code || !editForm.name) {
      toast.error('Kode dan nama rombel wajib diisi');
      return;
    }
    editLoading = true;
    try {
      const res = await fetch(`/api/academic/rombels/${editRombel.id}`, {
        method: 'PUT',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify({
          code: editForm.code,
          name: editForm.name,
          level: editForm.level,
          is_active: editForm.is_active,
        }),
      });
      if (res.ok) {
        toast.success(`Rombel ${editForm.code} berhasil diperbarui`);
        showEdit = false;
        editRombel = null;
        await refresh();
      } else {
        const body = await res.json().catch(() => ({}));
        toast.error(body?.error || 'Gagal memperbarui rombel');
      }
    } catch { toast.error('Gagal memperbarui rombel'); }
    finally { editLoading = false; }
  }

  async function hapusRombel(id: string, label: string) {
    if (!confirm(`Hapus rombel "${label}"?`)) return;
    deleteLoading = id;
    try {
      const res = await fetch(`/api/academic/rombels/${id}`, { method: 'DELETE' });
      if (res.ok || res.status === 204) {
        toast.success(`Rombel "${label}" dihapus`);
        await refresh();
      } else { toast.error('Gagal menghapus rombel'); }
    } catch { toast.error('Gagal menghapus rombel'); }
    finally { deleteLoading = null; }
  }

  const grouped = $derived.by(() => {
    const map: Record<string, Rombel[]> = {};
    for (const r of rombels) {
      if (!map[r.level]) map[r.level] = [];
      map[r.level].push(r);
    }
    for (const key of Object.keys(map)) {
      map[key].sort((a, b) => a.code.localeCompare(b.code));
    }
    const order = ['VII', 'VIII', 'IX'];
    const sorted: Record<string, Rombel[]> = {};
    for (const lv of order) { if (map[lv]) sorted[lv] = map[lv]; }
    for (const lv of Object.keys(map)) { if (!order.includes(lv)) sorted[lv] = map[lv]; }
    return sorted;
  });
</script>

<svelte:head><title>Rombel — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6">
  <!-- Header -->
  <div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
    <div>
      <h1 class="text-2xl font-black text-foreground">Rombongan Belajar</h1>
      <p class="mt-1 text-sm text-muted-foreground">
        Kelola kelas untuk
        {#if activeSemester}
          <span class="font-semibold text-foreground">{activeSemester.label}</span>
        {:else}
          <span class="italic">— belum ada semester aktif</span>
        {/if}
      </p>
    </div>
    {#if isAdmin}
      <Button onclick={() => (showTambah = true)} disabled={!activeSemester}>+ Tambah Rombel</Button>
    {/if}
  </div>

  <!-- Empty state -->
  {#if Object.keys(grouped).length === 0}
    <Card.Root>
      <Card.Content class="p-8 text-center">
        <p class="text-sm text-muted-foreground">Belum ada rombel untuk semester ini.</p>
      </Card.Content>
    </Card.Root>
  {:else}
    {#each Object.entries(grouped) as [level, kelas]}
      <div class="space-y-3">
        <h2 class="text-lg font-bold text-foreground">Kelas {level}</h2>
        <div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {#each kelas as rombel (rombel.id)}
            <Card.Root class="relative overflow-hidden hover:shadow-md transition-shadow {rombel.is_active ? '' : 'opacity-60'}">
              <Card.Content class="p-5">
                <div class="flex items-start justify-between gap-3">
                  <div class="flex-1 min-w-0">
                    <a href={`/academic/rombels/${rombel.id}`} class="hover:text-primary transition-colors">
                      <h3 class="text-base font-semibold text-foreground">{rombel.code}</h3>
                    </a>
                    <p class="mt-0.5 text-xs text-muted-foreground truncate">{rombel.name}</p>
                    <div class="flex gap-1.5 mt-2">
                      <Badge variant="secondary" class="text-[10px]">{rombel.level}</Badge>
                      {#if !rombel.is_active}
                        <Badge variant="outline" class="text-[10px]">nonaktif</Badge>
                      {/if}
                    </div>
                  </div>
                  {#if isAdmin}
                    <div class="flex flex-col gap-1 shrink-0">
                      <Button size="sm" variant="outline" class="text-xs px-2.5" onclick={() => openEdit(rombel)}>Edit</Button>
                      <Button size="sm" variant="outline" class="text-xs px-2.5 text-destructive" onclick={() => hapusRombel(rombel.id, rombel.name)} disabled={deleteLoading === rombel.id}>Hapus</Button>
                    </div>
                  {/if}
                </div>
              </Card.Content>
            </Card.Root>
          {/each}
        </div>
      </div>
    {/each}
  {/if}
</div>

<!-- Dialog Tambah -->
<Dialog.Root bind:open={showTambah}>
  <Dialog.Content>
    <div class="space-y-4">
      <div>
        <h2 class="text-base font-semibold text-foreground">Tambah Rombel</h2>
        <p class="mt-1 text-sm text-muted-foreground">
          Semester: <span class="font-medium text-foreground">{activeSemester?.label ?? '—'}</span>
        </p>
      </div>
      <div class="grid gap-4 sm:grid-cols-2">
        <div class="space-y-1.5">
          <label class="block text-xs font-medium text-muted-foreground">Kode Rombel</label>
          <Input bind:value={tambahForm.code} placeholder="Contoh: VII.A" />
        </div>
        <div class="space-y-1.5">
          <label class="block text-xs font-medium text-muted-foreground">Nama Rombel</label>
          <Input bind:value={tambahForm.name} placeholder="Contoh: VII.A" />
        </div>
        <div class="space-y-1.5">
          <label class="block text-xs font-medium text-muted-foreground">Tingkat</label>
          <select bind:value={tambahForm.level} class="flex h-10 w-full rounded-lg border border-input bg-background px-3 text-sm text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring">
            {#each levels as lv}<option value={lv}>Kelas {lv}</option>{/each}
          </select>
        </div>
        <div class="flex items-end">
          <p class="text-xs text-muted-foreground">Rombel akan dibuat di tahun {activeSemester?.label ?? '—'}</p>
        </div>
      </div>
      <div class="flex justify-end gap-2">
        <Button variant="outline" onclick={() => (showTambah = false)}>Batal</Button>
        <LoadingButton onclick={() => void submitTambah()} loading={tambahLoading} loadingLabel="Menyimpan...">Simpan</LoadingButton>
      </div>
    </div>
  </Dialog.Content>
</Dialog.Root>

<!-- Dialog Edit -->
<Dialog.Root bind:open={showEdit}>
  <Dialog.Content>
    <div class="space-y-4">
      <h2 class="text-base font-semibold text-foreground">Edit Rombel</h2>
      <p class="text-sm text-muted-foreground">Perbarui data rombel {editRombel?.code}</p>
      <div class="grid gap-4 sm:grid-cols-2">
        <div class="space-y-1.5">
          <label class="block text-xs font-medium text-muted-foreground">Kode Rombel</label>
          <Input bind:value={editForm.code} placeholder="Contoh: VII.A" />
        </div>
        <div class="space-y-1.5">
          <label class="block text-xs font-medium text-muted-foreground">Nama Rombel</label>
          <Input bind:value={editForm.name} placeholder="Contoh: VII.A" />
        </div>
        <div class="space-y-1.5">
          <label class="block text-xs font-medium text-muted-foreground">Tingkat</label>
          <select bind:value={editForm.level} class="flex h-10 w-full rounded-lg border border-input bg-background px-3 text-sm text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring">
            {#each levels as lv}<option value={lv}>Kelas {lv}</option>{/each}
          </select>
        </div>
        <div class="space-y-1.5">
          <label class="block text-xs font-medium text-muted-foreground">Status</label>
          <div class="flex items-center gap-2 pt-2">
            <input type="checkbox" id="edit-active" bind:checked={editForm.is_active} class="size-4 accent-primary" />
            <label for="edit-active" class="text-xs text-foreground">Aktif</label>
          </div>
        </div>
      </div>
      <div class="flex justify-end gap-2">
        <Button variant="outline" onclick={() => (showEdit = false)}>Batal</Button>
        <LoadingButton onclick={() => void submitEdit()} loading={editLoading} loadingLabel="Menyimpan...">Simpan Perubahan</LoadingButton>
      </div>
    </div>
  </Dialog.Content>
</Dialog.Root>
