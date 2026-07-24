<script lang="ts">
  import { Badge } from '$lib/components/ui/badge';
  import { Button } from '$lib/components/ui/button';
  import * as Card from '$lib/components/ui/card';
  import * as Dialog from '$lib/components/ui/dialog';
  import { Input } from '$lib/components/ui/input';
  import { toast } from '$lib/components/ui/sonner';
  import LoadingButton from '$lib/components/LoadingButton.svelte';

  let { data } = $props();

  type Profile = { id: string; code: string; name: string; regulation_reference: string; education_level: string; status: string; notes: string; };
  type Rombel = { id: string; code: string; name: string; level: string; };
  type Allocation = { id: string; subject_id: string; subject_name: string; level: string; subject_group: string; intra_weekly_hours: number; koku_weekly_hours: number; total_weekly_hours: number; };

  let profiles = $state<Profile[]>(data.profiles ?? []);
  let rombels = $state<Rombel[]>(data.rombels ?? []);
  let selected = $state<Profile | null>(null);
  let allocations = $state<Allocation[]>([]);
  let assignments = $state<{ id: string; class_id: string; class_code: string; class_name: string; class_level: string; curriculum_code: string; }[]>([]);
  let level = $state('VII');
  let activeTab = $state('alloc');
  let loading = $state(false);

  // - Edit profile
  let showEditProfile = $state(false);
  let editForm = $state({ code: '', name: '', regulation_reference: '', status: 'draft', notes: '' });

  // - Tambah profile
  let showTambah = $state(false);
  let tambahForm = $state({ code: '', name: '', regulation_reference: '', status: 'draft' });
  let tambahLoading = $state(false);

  // - Tambah alokasi
  let showAddAlloc = $state(false);
  let addForm = $state({ subject_name: '', lv: 'VII', grp: 'wajib', intra: '2', koku: '0' });
  let addLoading = $state(false);

  function loadAllocs() {
    if (!selected) return;
    loading = true;
    fetch(`/api/academic/curriculum/profiles/${selected.id}/allocations?level=${level}`)
      .then(r => r.json()).then(p => { allocations = p.items ?? []; }).catch(() => {}).finally(() => loading = false);
  }
  function loadAssigns() {
    if (!selected) return;
    fetch('/api/academic/curriculum/assignments?class_id=')
      .then(r => r.json()).then(p => { const all = p.items ?? []; assignments = all.filter((a: any) => a.curriculum_profile_id === selected!.id); }).catch(() => {});
  }

  function doSelect(p: Profile) {
    selected = p; level = 'VII'; activeTab = 'alloc';
    loadAllocs(); loadAssigns();
  }

  async function refreshProfiles() {
    const r = await fetch('/api/academic/curriculum/profiles');
    if (r.ok) { const p = await r.json(); profiles = p.items ?? []; }
  }

  async function submitTambah() {
    if (!tambahForm.code || !tambahForm.name) return;
    tambahLoading = true;
    const r = await fetch('/api/academic/curriculum/profiles', { method: 'POST', headers: { 'content-type': 'application/json' }, body: JSON.stringify(tambahForm) });
    if (r.ok) { toast.success('Kurikulum dibuat'); showTambah = false; await refreshProfiles(); }
    else { const e = await r.json().catch(() => ({})); toast.error(e?.error || 'Gagal'); }
    tambahLoading = false;
  }

  async function activate(id: string) {
    await fetch(`/api/academic/curriculum/profiles/${id}/activate`, { method: 'POST' });
    toast.success('Diaktifkan'); await refreshProfiles();
  }

  async function delProfile(id: string, name: string) {
    if (!confirm(`Hapus "${name}"?`)) return;
    const r = await fetch(`/api/academic/curriculum/profiles/${id}`, { method: 'DELETE' });
    if (r.ok || r.status === 204) { toast.success(`"${name}" dihapus`); if (selected?.id === id) { selected = null; } await refreshProfiles(); }
    else toast.error('Gagal');
  }

  async function openEdit() {
    if (!selected) return;
    editForm = { code: selected.code, name: selected.name, regulation_reference: selected.regulation_reference, status: selected.status, notes: selected.notes };
    showEditProfile = true;
  }

  async function submitEdit() {
    if (!selected) return;
    const r = await fetch(`/api/academic/curriculum/profiles/${selected.id}`, { method: 'PUT', headers: { 'content-type': 'application/json' }, body: JSON.stringify(editForm) });
    if (r.ok) { toast.success('Diupdate'); showEditProfile = false; await refreshProfiles(); const u = profiles.find(p => p.id === selected!.id); if (u) selected = u; }
    else { const e = await r.json().catch(() => ({})); toast.error(e?.error || 'Gagal'); }
  }

  async function submitAddAlloc() {
    if (!addForm.subject_name || !selected) return;
    addLoading = true;
    const r = await fetch(`/api/academic/curriculum/profiles/${selected.id}/allocations`, { method: 'POST', headers: { 'content-type': 'application/json' }, body: JSON.stringify({ subject_id: '00000000-0000-0000-0000-000000000000', level: addForm.lv, subject_group: addForm.grp, intra_weekly_hours: parseFloat(addForm.intra) || 0, koku_weekly_hours: parseFloat(addForm.koku) || 0, total_weekly_hours: (parseFloat(addForm.intra) || 0) + (parseFloat(addForm.koku) || 0) }) });
    if (r.ok) { toast.success('Mapel ditambahkan'); showAddAlloc = false; addForm = { subject_name: '', lv: level, grp: 'wajib', intra: '2', koku: '0' }; await loadAllocs(); }
    else { const e = await r.json().catch(() => ({})); toast.error(e?.error || 'Gagal'); }
    addLoading = false;
  }

  async function delAlloc(id: string) {
    if (!confirm('Hapus alokasi?')) return;
    const r = await fetch(`/api/academic/curriculum/allocations/${id}`, { method: 'DELETE' });
    if (r.ok || r.status === 204) { toast.success('Dihapus'); await loadAllocs(); } else toast.error('Gagal');
  }

  async function assign(cid: string) {
    if (!selected) return;
    const r = await fetch('/api/academic/curriculum/assignments', { method: 'POST', headers: { 'content-type': 'application/json' }, body: JSON.stringify({ class_id: cid, curriculum_profile_id: selected.id }) });
    if (r.ok) { toast.success('Assign OK'); await loadAssigns(); } else { const e = await r.json().catch(() => ({})); toast.error(e?.error || 'Gagal'); }
  }

  async function unassign(id: string) {
    const r = await fetch(`/api/academic/curriculum/assignments/${id}`, { method: 'DELETE' });
    if (r.ok || r.status === 204) { toast.success('Lepas'); await loadAssigns(); } else toast.error('Gagal');
  }
</script>

<svelte:head><title>Kurikulum — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-4 max-w-screen-xl mx-auto">
  <div class="flex items-center justify-between gap-3">
    <div><h1 class="text-xl font-black">Kurikulum</h1><p class="text-sm text-muted-foreground">Profil, alokasi mapel, dan assign rombel.</p></div>
    <Button onclick={() => (showTambah = true)}>+ Tambah</Button>
  </div>

  {#if profiles.length === 0}
    <Card.Root><Card.Content class="p-8 text-center"><p class="text-sm text-muted-foreground">Belum ada kurikulum.</p></Card.Content></Card.Root>
  {:else}
    <!-- Desktop: side by side -->
    <div class="hidden lg:grid lg:grid-cols-3 gap-6">
      <div class="space-y-2">
        <h2 class="text-xs font-semibold text-muted-foreground uppercase tracking-wider px-0.5">Daftar</h2>
        {#each profiles as p (p.id)}
          <button type="button" class="w-full text-left rounded-xl border border-border bg-base-100 shadow-sm p-3.5 transition-all hover:border-primary/30 {selected?.id === p.id ? 'border-primary ring-1 ring-primary/20' : ''} {p.status === 'archived' ? 'opacity-60' : ''}" onclick={() => doSelect(p)}>
            <div class="flex items-center gap-2"><div class="min-w-0"><div class="flex items-center gap-1.5"><span class="text-sm font-semibold truncate">{p.code}</span><Badge variant={p.status === 'active' ? 'default' : p.status === 'archived' ? 'secondary' : 'outline'} class="text-[9px] px-1.5">{p.status}</Badge></div><p class="mt-0.5 text-xs text-muted-foreground truncate">{p.name}</p></div></div>
          </button>
        {/each}
      </div>
      <div class="lg:col-span-2">
        {#if selected}
          <div class="card bg-base-100 border border-base-300 shadow-sm">
            <div class="card-header px-5 pt-4 pb-2">
              <div class="flex flex-wrap items-center justify-between gap-2">
                <div><h3 class="text-base font-semibold">{selected.name}</h3><p class="text-xs text-muted-foreground">{selected.regulation_reference || selected.code}</p></div>
                <div class="flex gap-1.5">
                  <Button size="sm" variant="outline" onclick={openEdit}>Edit</Button>
                  {#if selected.status !== 'active'}<Button size="sm" onclick={() => activate(selected.id)}>Aktifkan</Button>{/if}
                  <Button size="sm" variant="outline" class="text-destructive" onclick={() => delProfile(selected.id, selected.name)}>Hapus</Button>
                </div>
              </div>
            </div>
            <div class="card-content px-5 pb-5 space-y-4">
              <div class="flex gap-1 border-b border-border pb-1">
                <button class="px-3 py-1.5 text-xs font-medium rounded-t transition-colors {activeTab === 'alloc' ? 'bg-primary/10 text-primary' : 'text-muted-foreground'}" onclick={() => activeTab = 'alloc'}>Alokasi Mapel</button>
                <button class="px-3 py-1.5 text-xs font-medium rounded-t transition-colors {activeTab === 'assign' ? 'bg-primary/10 text-primary' : 'text-muted-foreground'}" onclick={() => activeTab = 'assign'}>Assign Rombel</button>
              </div>

              {#if activeTab === 'alloc'}
                <div class="flex flex-wrap items-center justify-between gap-2">
                  <div class="flex gap-1">{#each ['VII','VIII','IX'] as lv}<button class="px-2.5 py-1 text-xs font-medium rounded-md transition-colors {level === lv ? 'bg-primary text-primary-foreground' : 'border border-border text-muted-foreground'}" onclick={() => level = lv}>{lv}</button>{/each}</div>
                  <Button size="sm" onclick={() => (showAddAlloc = true)}>+ Mapel</Button>
                </div>
                {#if loading}<p class="text-sm text-muted-foreground py-4 text-center">Memuat...</p>
                {:else if allocations.length === 0}<p class="text-sm text-muted-foreground py-6 text-center">Belum ada mapel untuk {level}.</p>
                {:else}
                  <div class="overflow-x-auto -mx-5"><div class="inline-block min-w-full align-middle"><table class="w-full text-sm"><thead><tr class="bg-muted/30 text-muted-foreground text-xs uppercase"><th class="px-3 py-2 text-left">Mapel</th><th class="px-3 py-2 text-center hidden sm:table-cell">Kelompok</th><th class="px-3 py-2 text-center">Intra</th><th class="px-3 py-2 text-center hidden sm:table-cell">Koku</th><th class="px-3 py-2 text-center">Total</th><th class="px-3 py-2 text-right">Aksi</th></tr></thead><tbody class="divide-y divide-border">
                    {#each allocations as a (a.id)}<tr class="hover:bg-muted/20"><td class="px-3 py-2 font-medium text-xs md:text-sm">{a.subject_name}</td><td class="px-3 py-2 text-center text-muted-foreground text-xs hidden sm:table-cell">{a.subject_group}</td><td class="px-3 py-2 text-center text-xs">{a.intra_weekly_hours}</td><td class="px-3 py-2 text-center text-xs hidden sm:table-cell">{a.koku_weekly_hours}</td><td class="px-3 py-2 text-center font-semibold text-xs">{a.total_weekly_hours}</td><td class="px-3 py-2 text-right whitespace-nowrap"><button class="text-xs text-destructive hover:underline" onclick={() => delAlloc(a.id)}>Hapus</button></td></tr>{/each}
                  </tbody></table></div></div>
                {/if}
              {:else}
                <div class="overflow-x-auto -mx-5"><div class="inline-block min-w-full align-middle"><table class="w-full text-sm"><thead><tr class="bg-muted/30 text-muted-foreground text-xs uppercase"><th class="px-3 py-2 text-left">Rombel</th><th class="px-3 py-2 text-center hidden sm:table-cell">Tingkat</th><th class="px-3 py-2 text-center">Kurikulum</th><th class="px-3 py-2 text-right">Aksi</th></tr></thead><tbody class="divide-y divide-border">
                  {#each rombels as r (r.id)}<tr class="hover:bg-muted/20">
                    <td class="px-3 py-2 font-medium text-xs">{r.name}</td>
                    <td class="px-3 py-2 text-center text-xs text-muted-foreground hidden sm:table-cell">{r.level}</td>
                    <td class="px-3 py-2 text-center">{#if assignments.find(a => a.class_id === r.id)}<Badge variant="default" class="text-[9px] bg-primary/10 text-primary">{(assignments.find(a => a.class_id === r.id))?.curriculum_code}</Badge>{:else}<span class="text-muted-foreground italic text-xs">—</span>{/if}</td>
                    <td class="px-3 py-2 text-right">{#if assignments.find(a => a.class_id === r.id)}<button class="text-xs text-destructive hover:underline" onclick={() => unassign(assignments.find(a => a.class_id === r.id)!.id)}>Lepas</button>{:else}<button class="text-xs text-primary hover:underline" onclick={() => assign(r.id)}>Assign</button>{/if}</td>
                  </tr>{/each}
                </tbody></table></div></div>
              {/if}
            </div>
          </div>
        {:else}
          <Card.Root><Card.Content class="p-12 text-center"><p class="text-sm text-muted-foreground">Pilih kurikulum dari daftar.</p></Card.Content></Card.Root>
        {/if}
      </div>
    </div>

    <!-- Mobile: full screen toggle -->
    <div class="lg:hidden space-y-3">
      {#if !selected}
        {#each profiles as p (p.id)}
          <button type="button" class="w-full text-left rounded-xl border border-border bg-base-100 shadow-sm p-4 transition-transform active:scale-[0.98]" onclick={() => doSelect(p)}>
            <div class="flex items-center justify-between gap-2">
              <div class="min-w-0 flex-1"><div class="flex items-center gap-1.5"><span class="text-sm font-semibold">{p.code}</span><Badge variant={p.status === 'active' ? 'default' : p.status === 'archived' ? 'secondary' : 'outline'} class="text-[9px] px-1.5">{p.status}</Badge></div><p class="mt-0.5 text-xs text-muted-foreground">{p.name}</p></div>
              <svg class="w-4 h-4 text-muted-foreground shrink-0" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="9 18 15 12 9 6"/></svg>
            </div>
          </button>
        {/each}
      {:else}
        <div class="space-y-3">
          <button class="flex items-center gap-1 text-sm font-medium text-muted-foreground" onclick={() => selected = null}><svg class="w-4 h-4" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="15 18 9 12 15 6"/></svg>Kembali</button>
          <div class="card bg-base-100 border border-base-300 shadow-sm">
            <div class="card-header px-4 pt-3 pb-2">
              <div class="flex flex-wrap items-center justify-between gap-2">
                <div><h3 class="text-sm font-semibold">{selected.name}</h3><p class="text-xs text-muted-foreground">{selected.regulation_reference || selected.code}</p></div>
                <div class="flex gap-1"><Button size="sm" variant="outline" class="text-xs px-2" onclick={openEdit}>Edit</Button><Button size="sm" variant="outline" class="text-destructive text-xs px-2" onclick={() => delProfile(selected.id, selected.name)}>Hapus</Button></div>
              </div>
            </div>
            <div class="card-content px-4 pb-4 space-y-3">
              <div class="flex gap-1 border-b border-border pb-1 overflow-x-auto">
                <button class="whitespace-nowrap px-2 py-1.5 text-xs font-medium rounded-t {activeTab === 'alloc' ? 'bg-primary/10 text-primary' : 'text-muted-foreground'}" onclick={() => activeTab = 'alloc'}>Alokasi Mapel</button>
                <button class="whitespace-nowrap px-2 py-1.5 text-xs font-medium rounded-t {activeTab === 'assign' ? 'bg-primary/10 text-primary' : 'text-muted-foreground'}" onclick={() => activeTab = 'assign'}>Assign Rombel</button>
              </div>
              {#if activeTab === 'alloc'}
                <div class="flex flex-wrap items-center justify-between gap-2">
                  <div class="flex gap-1">{#each ['VII','VIII','IX'] as lv}<button class="px-2 py-1 text-xs font-medium rounded-md {level === lv ? 'bg-primary text-primary-foreground' : 'border border-border text-muted-foreground'}" onclick={() => level = lv}>{lv}</button>{/each}</div>
                  <Button size="sm" class="text-xs px-2.5" onclick={() => (showAddAlloc = true)}>+ Mapel</Button>
                </div>
                {#if loading}<p class="text-sm text-muted-foreground py-4 text-center">Memuat...</p>
                {:else if allocations.length === 0}<p class="text-sm text-muted-foreground py-6 text-center">Belum ada mapel {level}.</p>
                {:else}<div class="overflow-x-auto -mx-4"><div class="inline-block min-w-full align-middle"><table class="w-full text-sm"><thead><tr class="bg-muted/30 text-muted-foreground text-xs uppercase"><th class="px-3 py-2 text-left">Mapel</th><th class="px-3 py-2 text-center">Jam</th><th class="px-3 py-2 text-right">Aksi</th></tr></thead><tbody class="divide-y divide-border">{#each allocations as a (a.id)}<tr class="hover:bg-muted/20"><td class="px-3 py-2 font-medium text-xs">{a.subject_name}</td><td class="px-3 py-2 text-center text-xs">{a.total_weekly_hours}</td><td class="px-3 py-2 text-right"><button class="text-xs text-destructive hover:underline" onclick={() => delAlloc(a.id)}>Hapus</button></td></tr>{/each}</tbody></table></div></div>{/if}
              {:else}
                <div class="overflow-x-auto -mx-4"><div class="inline-block min-w-full align-middle"><table class="w-full text-sm"><thead><tr class="bg-muted/30 text-muted-foreground text-xs uppercase"><th class="px-3 py-2 text-left">Rombel</th><th class="px-3 py-2 text-center">Status</th><th class="px-3 py-2 text-right">Aksi</th></tr></thead><tbody class="divide-y divide-border">{#each rombels as r (r.id)}<tr class="hover:bg-muted/20"><td class="px-3 py-2 font-medium text-xs">{r.name}</td><td class="px-3 py-2 text-center">{#if assignments.find(a => a.class_id === r.id)}<Badge variant="default" class="text-[9px] bg-primary/10 text-primary">{(assignments.find(a => a.class_id === r.id))?.curriculum_code}</Badge>{:else}<span class="text-muted-foreground italic text-xs">—</span>{/if}</td><td class="px-3 py-2 text-right">{#if assignments.find(a => a.class_id === r.id)}<button class="text-xs text-destructive hover:underline" onclick={() => unassign(assignments.find(a => a.class_id === r.id)!.id)}>Lepas</button>{:else}<button class="text-xs text-primary hover:underline" onclick={() => assign(r.id)}>Assign</button>{/if}</td></tr>{/each}</tbody></table></div></div>
              {/if}
            </div>
          </div>
        </div>
      {/if}
    </div>
  {/if}
</div>

<Dialog.Root bind:open={showTambah}>
  <Dialog.Content><div class="space-y-4"><h2 class="text-base font-semibold">Tambah Kurikulum</h2>
    <div class="grid gap-3 sm:grid-cols-2"><div class="space-y-1"><label class="text-xs font-medium text-muted-foreground">Kode</label><Input bind:value={tambahForm.code} placeholder="KMA-1503-2025" /></div><div class="space-y-1"><label class="text-xs font-medium text-muted-foreground">Nama</label><Input bind:value={tambahForm.name} placeholder="Kurikulum Merdeka" /></div><div class="space-y-1 sm:col-span-2"><label class="text-xs font-medium text-muted-foreground">Regulasi</label><Input bind:value={tambahForm.regulation_reference} placeholder="KMA 1503 Th 2025" /></div></div>
    <div class="flex justify-end gap-2"><Button variant="outline" onclick={() => (showTambah = false)}>Batal</Button><LoadingButton onclick={() => submitTambah()} loading={tambahLoading}>Simpan</LoadingButton></div>
  </div></Dialog.Content>
</Dialog.Root>

<Dialog.Root bind:open={showEditProfile}>
  <Dialog.Content><div class="space-y-4"><h2 class="text-base font-semibold">Edit Kurikulum</h2>
    <div class="grid gap-3 sm:grid-cols-2"><div class="space-y-1"><label class="text-xs font-medium text-muted-foreground">Kode</label><Input bind:value={editForm.code} /></div><div class="space-y-1"><label class="text-xs font-medium text-muted-foreground">Nama</label><Input bind:value={editForm.name} /></div><div class="space-y-1 sm:col-span-2"><label class="text-xs font-medium text-muted-foreground">Regulasi</label><Input bind:value={editForm.regulation_reference} /></div></div>
    <div class="flex justify-end gap-2"><Button variant="outline" onclick={() => (showEditProfile = false)}>Batal</Button><LoadingButton onclick={() => submitEdit()} loading={false}>Simpan</LoadingButton></div>
  </div></Dialog.Content>
</Dialog.Root>

<Dialog.Root bind:open={showAddAlloc}>
  <Dialog.Content><div class="space-y-4"><h2 class="text-base font-semibold">Tambah Mapel</h2><p class="text-sm text-muted-foreground">{selected?.code} — {addForm.lv}</p>
    <div class="grid gap-3 sm:grid-cols-2"><div class="space-y-1 sm:col-span-2"><label class="text-xs font-medium text-muted-foreground">Nama Mapel</label><Input bind:value={addForm.subject_name} placeholder="Matematika" /></div><div class="space-y-1"><label class="text-xs font-medium text-muted-foreground">Tingkat</label><select bind:value={addForm.lv} class="flex h-9 w-full rounded-lg border border-input bg-background px-3 text-sm"><option value="VII">VII</option><option value="VIII">VIII</option><option value="IX">IX</option></select></div><div class="space-y-1"><label class="text-xs font-medium text-muted-foreground">Kelompok</label><select bind:value={addForm.grp} class="flex h-9 w-full rounded-lg border border-input bg-background px-3 text-sm"><option value="wajib">wajib</option><option value="pilihan">pilihan</option><option value="muatan_lokal">muatan_lokal</option><option value="kokurikuler">kokurikuler</option></select></div><div class="space-y-1"><label class="text-xs font-medium text-muted-foreground">Jam</label><Input type="number" bind:value={addForm.intra} min="0" step="0.5" /></div></div>
    <div class="flex justify-end gap-2"><Button variant="outline" onclick={() => (showAddAlloc = false)}>Batal</Button><LoadingButton onclick={() => submitAddAlloc()} loading={addLoading}>Simpan</LoadingButton></div>
  </div></Dialog.Content>
</Dialog.Root>
