<script lang="ts">
  import { Badge } from '$lib/components/ui/badge';
  import { Button } from '$lib/components/ui/button';
  import * as Card from '$lib/components/ui/card';
  import * as Dialog from '$lib/components/ui/dialog';
  import { Input } from '$lib/components/ui/input';
  import { toast } from '$lib/components/ui/sonner';
  import LoadingButton from '$lib/components/LoadingButton.svelte';
  import { page } from '$app/stores';

  let { data } = $props();

  type Profile = { id: string; code: string; name: string; regulation_reference: string; education_level: string; effective_academic_year_id: string; status: string; notes: string; };
  type Rombel = { id: string; code: string; name: string; level: string; academic_year_id: string; academic_year_name: string; };
  type Allocation = { id: string; subject_id: string; subject_code: string; subject_name: string; level: string; subject_group: string; intra_weekly_hours: number; koku_weekly_hours: number; total_weekly_hours: number; is_required: boolean; };

  let profiles = $state<Profile[]>(data.profiles ?? []);
  let rombels = $state<Rombel[]>(data.rombels ?? []);
  let selectedProfile = $state<Profile | null>(null);
  let allocations = $state<Allocation[]>([]);
  let assignments = $state<{ id: string; class_id: string; class_code: string; class_name: string; class_level: string; curriculum_profile_id: string; curriculum_code: string; }[]>([]);
  let selectedLevel = $state('VII');
  let activeTab = $state('alloc');
  let loading = $state(false);

  // — Edit profile
  let showEditProfile = $state(false);
  let editProfileForm = $state<Profile>({ id: '', code: '', name: '', regulation_reference: '', education_level: 'MTs', effective_academic_year_id: '', status: 'draft', notes: '' });
  let editProfileLoading = $state(false);

  // — Tambah profile
  let showTambah = $state(false);
  let tambahForm = $state({ code: '', name: '', regulation_reference: '', education_level: 'MTs', status: 'draft', notes: '' });
  let tambahLoading = $state(false);
  let actLoading = $state<string | null>(null);

  // — Tambah alokasi
  let showAddAlloc = $state(false);
  let addAllocForm = $state({ subject_name: '', level: 'VII', group: 'wajib', intra: '2', koku: '0', notes: '' });
  let addAllocLoading = $state(false);

  // — Edit alokasi
  let showEditAlloc = $state(false);
  let editAllocId = $state('');
  let editAllocForm = $state({ level: 'VII', group: 'wajib', intra: '2', koku: '0', notes: '' });
  let editAllocLoading = $state(false);

  const levels = ['VII', 'VIII', 'IX'];
  const groups = ['wajib', 'pilihan', 'muatan_lokal', 'layanan', 'kokurikuler', 'kegiatan'];

  // — Responsive: di mobile detail ganti list
  let showDetail = $state(false);

  async function refreshProfiles() {
    const r = await fetch('/api/academic/curriculum/profiles');
    if (r.ok) { const p = await r.json(); profiles = p.items ?? []; }
  }

  async function selectProfile(p: Profile) {
    selectedProfile = p;
    showDetail = true;
    selectedLevel = 'VII';
    activeTab = 'alloc';
    loadAllocations();
    loadAssignments();
  }

  function backToList() {
    showDetail = false;
    selectedProfile = null;
  }

  async function loadAllocations() {
    if (!selectedProfile) return;
    loading = true;
    const r = await fetch(`/api/academic/curriculum/profiles/${selectedProfile.id}/allocations?level=${selectedLevel}`);
    if (r.ok) { const p = await r.json(); allocations = p.items ?? []; } else allocations = [];
    loading = false;
  }

  async function loadAssignments() {
    if (!selectedProfile) return;
    const r = await fetch(`/api/academic/curriculum/assignments?class_id=`);
    if (r.ok) { const p = await r.json(); const all = p.items ?? []; assignments = all.filter((a: any) => a.curriculum_profile_id === selectedProfile!.id); }
    else assignments = [];
  }

  $effect(() => { if (selectedLevel && selectedProfile) loadAllocations(); });

  // — Profile CRUD

  async function submitTambah() {
    if (!tambahForm.code || !tambahForm.name) { toast.error('Kode dan nama wajib'); return; }
    tambahLoading = true;
    try {
      const r = await fetch('/api/academic/curriculum/profiles', {
        method: 'POST', headers: { 'content-type': 'application/json' }, body: JSON.stringify(tambahForm),
      });
      if (r.ok) {
        toast.success('Kurikulum dibuat'); showTambah = false;
        tambahForm = { code: '', name: '', regulation_reference: '', education_level: 'MTs', status: 'draft', notes: '' };
        await refreshProfiles();
      } else { const e = await r.json().catch(() => ({})); toast.error(e?.error || 'Gagal'); }
    } catch { toast.error('Gagal'); }
    finally { tambahLoading = false; }
  }

  function openEditProfile() {
    if (!selectedProfile) return;
    editProfileForm = { ...selectedProfile };
    showEditProfile = true;
  }

  async function submitEditProfile() {
    if (!editProfileForm.code || !editProfileForm.name || !selectedProfile) { toast.error('Kode dan nama wajib'); return; }
    editProfileLoading = true;
    try {
      const r = await fetch(`/api/academic/curriculum/profiles/${selectedProfile.id}`, { method: 'PUT', headers: { 'content-type': 'application/json' }, body: JSON.stringify(editProfileForm) });
      if (r.ok) {
        toast.success('Kurikulum diupdate'); showEditProfile = false;
        await refreshProfiles();
        const updated = profiles.find(p => p.id === selectedProfile!.id);
        if (updated) selectedProfile = updated;
      } else { const e = await r.json().catch(() => ({})); toast.error(e?.error || 'Gagal'); }
    } catch { toast.error('Gagal'); }
    finally { editProfileLoading = false; }
  }

  async function activateProfile(id: string) {
    actLoading = id;
    try {
      const r = await fetch(`/api/academic/curriculum/profiles/${id}/activate`, { method: 'POST' });
      if (r.ok) { toast.success('Diaktifkan'); await refreshProfiles(); } else toast.error('Gagal');
    } catch { toast.error('Gagal'); }
    finally { actLoading = null; }
  }

  async function deleteProfile(id: string, name: string) {
    if (!confirm(`Hapus "${name}"?`)) return;
    const r = await fetch(`/api/academic/curriculum/profiles/${id}`, { method: 'DELETE' });
    if (r.ok || r.status === 204) {
      toast.success(`"${name}" dihapus`);
      if (selectedProfile?.id === id) { selectedProfile = null; showDetail = false; }
      await refreshProfiles();
    } else toast.error('Gagal');
  }

  // — Allocation CRUD

  async function submitAddAlloc() {
    if (!addAllocForm.subject_name || !selectedProfile) { toast.error('Nama mapel wajib'); return; }
    addAllocLoading = true;
    try {
      const r = await fetch(`/api/academic/curriculum/profiles/${selectedProfile.id}/allocations`, {
        method: 'POST', headers: { 'content-type': 'application/json' }, body: JSON.stringify({
          subject_id: '00000000-0000-0000-0000-000000000000', level: addAllocForm.level,
          subject_group: addAllocForm.group, intra_weekly_hours: parseFloat(addAllocForm.intra) || 0,
          koku_weekly_hours: parseFloat(addAllocForm.koku) || 0,
          total_weekly_hours: (parseFloat(addAllocForm.intra) || 0) + (parseFloat(addAllocForm.koku) || 0), notes: addAllocForm.notes,
        }),
      });
      if (r.ok) { toast.success('Mapel ditambahkan'); showAddAlloc = false; addAllocForm = { subject_name: '', level: selectedLevel, group: 'wajib', intra: '2', koku: '0', notes: '' }; await loadAllocations(); }
      else { const e = await r.json().catch(() => ({})); toast.error(e?.error || 'Gagal'); }
    } catch { toast.error('Gagal'); }
    finally { addAllocLoading = false; }
  }

  function openEditAlloc(a: Allocation) {
    editAllocId = a.id;
    editAllocForm = { level: a.level, group: a.subject_group, intra: String(a.intra_weekly_hours), koku: String(a.koku_weekly_hours), notes: '' };
    showEditAlloc = true;
  }

  async function submitEditAlloc() {
    if (!editAllocId) return;
    editAllocLoading = true;
    try {
      const r = await fetch(`/api/academic/curriculum/allocations/${editAllocId}`, {
        method: 'PUT', headers: { 'content-type': 'application/json' }, body: JSON.stringify({
          subject_id: '00000000-0000-0000-0000-000000000000', level: editAllocForm.level, subject_group: editAllocForm.group,
          intra_weekly_hours: parseFloat(editAllocForm.intra) || 0, koku_weekly_hours: parseFloat(editAllocForm.koku) || 0,
          total_weekly_hours: (parseFloat(editAllocForm.intra) || 0) + (parseFloat(editAllocForm.koku) || 0), notes: editAllocForm.notes,
        }),
      });
      if (r.ok) { toast.success('Alokasi diupdate'); showEditAlloc = false; await loadAllocations(); }
      else { const e = await r.json().catch(() => ({})); toast.error(e?.error || 'Gagal'); }
    } catch { toast.error('Gagal'); }
    finally { editAllocLoading = false; }
  }

  async function deleteAlloc(id: string) {
    if (!confirm('Hapus alokasi ini?')) return;
    const r = await fetch(`/api/academic/curriculum/allocations/${id}`, { method: 'DELETE' });
    if (r.ok || r.status === 204) { toast.success('Dihapus'); await loadAllocations(); } else toast.error('Gagal');
  }

  // — Assign

  async function assignRombel(classId: string) {
    if (!selectedProfile) return;
    const r = await fetch('/api/academic/curriculum/assignments', {
      method: 'POST', headers: { 'content-type': 'application/json' }, body: JSON.stringify({ class_id: classId, curriculum_profile_id: selectedProfile.id, notes: '' }),
    });
    if (r.ok) { toast.success('Assign OK'); await loadAssignments(); } else { const e = await r.json().catch(() => ({})); toast.error(e?.error || 'Gagal'); }
  }

  async function unassignRombel(id: string) {
    const r = await fetch(`/api/academic/curriculum/assignments/${id}`, { method: 'DELETE' });
    if (r.ok || r.status === 204) { toast.success('Lepas OK'); await loadAssignments(); } else toast.error('Gagal');
  }

  function statusBadge(s: string) {
    if (s === 'active') return 'default';
    if (s === 'archived') return 'secondary';
    return 'outline';
  }
</script>

<svelte:head><title>Kurikulum — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-4 mx-auto max-w-screen-xl">
  <!-- HEADER -->
  <div class="flex items-center justify-between gap-3">
    <div>
      <h1 class="text-xl font-black text-foreground">Kurikulum</h1>
      <p class="text-sm text-muted-foreground">Profil, alokasi mapel, dan assign rombel.</p>
    </div>
    <Button onclick={() => (showTambah = true)} class="shrink-0">+ Tambah</Button>
  </div>

  {#if profiles.length === 0}
    <Card.Root><Card.Content class="p-8 text-center"><p class="text-sm text-muted-foreground">Belum ada kurikulum.</p></Card.Content></Card.Root>
  {:else}
    <!-- =============== DESKTOP: side-by-side =============== -->
    <div class="hidden lg:grid lg:grid-cols-3 gap-6">
      <!-- Sidebar list -->
      <div class="space-y-2 lg:col-span-1">
        <h2 class="text-xs font-semibold text-muted-foreground uppercase tracking-wider px-0.5">Daftar</h2>
        {#each profiles as p (p.id)}
          <Card.Root class="cursor-pointer transition-all hover:border-primary/30 {selectedProfile?.id === p.id ? 'border-primary ring-1 ring-primary/20' : ''} {p.status === 'archived' ? 'opacity-60' : ''}" onclick={() => selectProfile(p)}>
            <Card.Content class="p-3.5">
              <div class="flex items-center justify-between gap-2">
                <div class="min-w-0">
                  <div class="flex items-center gap-1.5">
                    <span class="text-sm font-semibold text-foreground truncate">{p.code}</span>
                    <Badge variant={statusBadge(p.status)} class="text-[9px] px-1.5">{p.status}</Badge>
                  </div>
                  <p class="mt-0.5 text-xs text-muted-foreground truncate">{p.name}</p>
                </div>
              </div>
            </Card.Content>
          </Card.Root>
        {/each}
      </div>

      <!-- Detail panel -->
      <div class="lg:col-span-2">
        {#if selectedProfile}
          <DetailPanel {selectedProfile} {allocations} {assignments} {rombels} {loading} {selectedLevel} {activeTab} {levels} {groups}
            onswitchlevel={(lv) => (selectedLevel = lv)}
            ontabswitch={(tab) => (activeTab = tab)}
            oneditprofile={openEditProfile}
            onactivate={(id) => void activateProfile(id)}
            ondelete={(id, name) => void deleteProfile(id, name)}
            onaddalloc={() => (showAddAlloc = true)}
            oneditaloc={openEditAlloc}
            ondeletalloc={deleteAlloc}
            onassign={assignRombel}
            onunassign={unassignRombel}
            actLoading={actLoading}
          />
        {:else}
          <Card.Root><Card.Content class="p-12 text-center"><p class="text-sm text-muted-foreground">Pilih kurikulum dari daftar di samping.</p></Card.Content></Card.Root>
        {/if}
      </div>
    </div>

    <!-- =============== MOBILE: list → detail pattern =============== -->
    <div class="lg:hidden space-y-3">
      {#if !showDetail || !selectedProfile}
        <!-- List -->
        {#each profiles as p (p.id)}
          <Card.Root class="cursor-pointer active:scale-[0.98] transition-transform {p.status === 'archived' ? 'opacity-60' : ''}" onclick={() => selectProfile(p)}>
            <Card.Content class="p-4">
              <div class="flex items-center justify-between gap-2">
                <div class="min-w-0 flex-1">
                  <div class="flex items-center gap-1.5">
                    <span class="text-sm font-semibold text-foreground">{p.code}</span>
                    <Badge variant={statusBadge(p.status)} class="text-[9px] px-1.5">{p.status}</Badge>
                  </div>
                  <p class="mt-0.5 text-xs text-muted-foreground">{p.name}</p>
                </div>
                <svg class="w-4 h-4 text-muted-foreground shrink-0" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="9 18 15 12 9 6"/></svg>
              </div>
            </Card.Content>
          </Card.Root>
        {/each}
      {:else}
        <!-- Detail full screen mobile -->
        <div class="space-y-3">
          <!-- Back header -->
          <div class="flex items-center gap-2">
            <button onclick={backToList} class="flex items-center gap-1 text-sm font-medium text-muted-foreground hover:text-foreground transition-colors">
              <svg class="w-4 h-4" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="15 18 9 12 15 6"/></svg>
              Kembali
            </button>
          </div>

          <DetailPanel {selectedProfile} {allocations} {assignments} {rombels} {loading} {selectedLevel} {activeTab} {levels} {groups}
            onswitchlevel={(lv) => (selectedLevel = lv)}
            ontabswitch={(tab) => (activeTab = tab)}
            oneditprofile={openEditProfile}
            onactivate={(id) => void activateProfile(id)}
            ondelete={(id, name) => void deleteProfile(id, name)}
            onaddalloc={() => (showAddAlloc = true)}
            oneditaloc={openEditAlloc}
            ondeletalloc={deleteAlloc}
            onassign={assignRombel}
            onunassign={unassignRombel}
            actLoading={actLoading}
          />
        </div>
      {/if}
    </div>
  {/if}
</div>

<!-- ─── Detail Panel Component ─── -->
{#snippet DetailPanel({selectedProfile: profile, allocations, assignments, rombels, loading, selectedLevel, activeTab, levels, groups, onswitchlevel, ontabswitch, oneditprofile, onactivate, ondelete, onaddalloc, oneditaloc, ondeletalloc, onassign, onunassign, actLoading}: any)}
  <Card.Root>
    <Card.Header class="px-4 pt-3 pb-2 md:px-5 md:pt-4 md:pb-2">
      <div class="flex flex-wrap items-center justify-between gap-2">
        <div class="min-w-0">
          <Card.Title class="text-sm md:text-base break-words">{profile.name}</Card.Title>
          <Card.Description class="text-xs break-words">{profile.regulation_reference || profile.code}</Card.Description>
        </div>
        <div class="flex items-center gap-1.5 shrink-0">
          <Button size="sm" variant="outline" class="text-xs px-2 md:px-3" onclick={oneditprofile}>Edit</Button>
          {#if profile.status !== 'active'}
            <Button size="sm" class="text-xs px-2 md:px-3" onclick={() => onactivate(profile.id)} disabled={actLoading === profile.id}>Aktifkan</Button>
          {/if}
          <Button size="sm" variant="outline" class="text-destructive text-xs px-2 md:px-3" onclick={() => ondelete(profile.id, profile.name)}>Hapus</Button>
        </div>
      </div>
    </Card.Header>

    <Card.Content class="px-4 pb-4 md:px-5 md:pb-5 space-y-4">
      <!-- Tab buttons -->
      <div class="flex gap-1 border-b border-border overflow-x-auto pb-1 -mx-4 px-4 md:mx-0 md:px-0">
        <button class="whitespace-nowrap px-2 md:px-3 py-1.5 text-xs font-medium rounded-t transition-colors {activeTab === 'alloc' ? 'bg-primary/10 text-primary' : 'text-muted-foreground hover:text-foreground'}" onclick={() => ontabswitch('alloc')}>Alokasi Mapel</button>
        <button class="whitespace-nowrap px-2 md:px-3 py-1.5 text-xs font-medium rounded-t transition-colors {activeTab === 'assign' ? 'bg-primary/10 text-primary' : 'text-muted-foreground hover:text-foreground'}" onclick={() => ontabswitch('assign')}>Assign Rombel</button>
      </div>

      {#if activeTab === 'alloc'}
        <!-- Level tabs + add -->
        <div class="flex flex-wrap items-center justify-between gap-2">
          <div class="flex gap-1">
            {#each levels as lv}
              <button class="px-2.5 py-1 text-xs font-medium rounded-md transition-colors {selectedLevel === lv ? 'bg-primary text-primary-foreground' : 'border border-border text-muted-foreground hover:border-primary/40'}" onclick={() => onswitchlevel(lv)}>{lv}</button>
            {/each}
          </div>
          <Button size="sm" class="text-xs px-2.5" onclick={onaddalloc}>+ Mapel</Button>
        </div>

        {#if loading}
          <p class="text-sm text-muted-foreground py-4 text-center">Memuat...</p>
        {:else if allocations.length === 0}
          <p class="text-sm text-muted-foreground py-6 text-center">Belum ada mapel untuk tingkat {selectedLevel}.</p>
        {:else}
          <!-- Table responsive -->
          <div class="overflow-x-auto -mx-4 md:mx-0">
            <table class="w-full text-sm min-w-[360px] md:min-w-0">
              <thead><tr class="bg-muted/30 text-muted-foreground text-xs uppercase">
                <th class="px-2 md:px-3 py-2 text-left">Mapel</th>
                <th class="px-2 md:px-3 py-2 text-center hidden sm:table-cell">Kelompok</th>
                <th class="px-2 md:px-3 py-2 text-center">Intra</th>
                <th class="px-2 md:px-3 py-2 text-center hidden sm:table-cell">Koku</th>
                <th class="px-2 md:px-3 py-2 text-center">Total</th>
                <th class="px-2 md:px-3 py-2 text-right">Aksi</th>
              </tr></thead>
              <tbody class="divide-y divide-border">
                {#each allocations as a (a.id)}
                  <tr class="hover:bg-muted/20">
                    <td class="px-2 md:px-3 py-2 font-medium text-foreground text-xs md:text-sm">{a.subject_name}</td>
                    <td class="px-2 md:px-3 py-2 text-center text-muted-foreground text-xs hidden sm:table-cell">{a.subject_group}</td>
                    <td class="px-2 md:px-3 py-2 text-center text-xs md:text-sm">{a.intra_weekly_hours}</td>
                    <td class="px-2 md:px-3 py-2 text-center text-xs md:text-sm hidden sm:table-cell">{a.koku_weekly_hours}</td>
                    <td class="px-2 md:px-3 py-2 text-center font-semibold text-xs md:text-sm">{a.total_weekly_hours}</td>
                    <td class="px-2 md:px-3 py-2 text-right whitespace-nowrap">
                      <button class="text-xs text-primary hover:underline mr-2" onclick={() => oneditaloc(a)}>Edit</button>
                      <button class="text-xs text-destructive hover:underline" onclick={() => ondeletalloc(a.id)}>Hapus</button>
                    </td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        {/if}
      {:else}
        <!-- Assign Rombel -->
        <div class="overflow-x-auto -mx-4 md:mx-0">
          <table class="w-full text-sm min-w-[320px] md:min-w-0">
            <thead><tr class="bg-muted/30 text-muted-foreground text-xs uppercase">
              <th class="px-2 md:px-3 py-2 text-left">Rombel</th>
              <th class="px-2 md:px-3 py-2 text-center hidden sm:table-cell">Tingkat</th>
              <th class="px-2 md:px-3 py-2 text-center">Kurikulum</th>
              <th class="px-2 md:px-3 py-2 text-right">Aksi</th>
            </tr></thead>
            <tbody class="divide-y divide-border">
              {#each rombels as r (r.id)}
                {@const assigned = assignments.find(a => a.class_id === r.id)}
                <tr class="hover:bg-muted/20">
                  <td class="px-2 md:px-3 py-2 font-medium text-foreground text-xs md:text-sm">{r.name}</td>
                  <td class="px-2 md:px-3 py-2 text-center text-muted-foreground text-xs hidden sm:table-cell">{r.level}</td>
                  <td class="px-2 md:px-3 py-2 text-center">
                    {#if assigned}
                      <Badge variant="default" class="text-[9px] bg-primary/10 text-primary">{assigned.curriculum_code}</Badge>
                    {:else}
                      <span class="text-muted-foreground italic text-xs">—</span>
                    {/if}
                  </td>
                  <td class="px-2 md:px-3 py-2 text-right">
                    {#if assigned}
                      <button class="text-xs text-destructive hover:underline" onclick={() => onunassign(assigned.id)}>Lepas</button>
                    {:else}
                      <button class="text-xs text-primary hover:underline" onclick={() => onassign(r.id)}>Assign</button>
                    {/if}
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      {/if}
    </Card.Content>
  </Card.Root>
{/snippet}

<!-- ─── Modal Tambah Kurikulum ─── -->
<Dialog.Root bind:open={showTambah}>
  <Dialog.Content>
    <div class="space-y-4">
      <h2 class="text-base font-semibold text-foreground">Tambah Kurikulum</h2>
      <div class="grid gap-3 sm:grid-cols-2">
        <div class="space-y-1"><label class="text-xs font-medium text-muted-foreground">Kode *</label><Input bind:value={tambahForm.code} placeholder="KMA-1503-2025" /></div>
        <div class="space-y-1"><label class="text-xs font-medium text-muted-foreground">Nama *</label><Input bind:value={tambahForm.name} placeholder="Kurikulum Merdeka" /></div>
        <div class="space-y-1 sm:col-span-2"><label class="text-xs font-medium text-muted-foreground">Regulasi</label><Input bind:value={tambahForm.regulation_reference} placeholder="KMA 1503 Th 2025" /></div>
        <div class="space-y-1">
          <label class="text-xs font-medium text-muted-foreground">Tingkat</label>
          <select bind:value={tambahForm.education_level} class="flex h-9 w-full rounded-lg border border-input bg-background px-3 text-sm"><option value="MTs">MTs</option><option value="MA">MA</option><option value="MI">MI</option></select>
        </div>
        <div class="space-y-1">
          <label class="text-xs font-medium text-muted-foreground">Status</label>
          <select bind:value={tambahForm.status} class="flex h-9 w-full rounded-lg border border-input bg-background px-3 text-sm"><option value="draft">Draft</option><option value="active">Aktif</option></select>
        </div>
      </div>
      <div class="flex justify-end gap-2 pt-1">
        <Button variant="outline" onclick={() => (showTambah = false)}>Batal</Button>
        <LoadingButton onclick={() => void submitTambah()} loading={tambahLoading}>Simpan</LoadingButton>
      </div>
    </div>
  </Dialog.Content>
</Dialog.Root>

<!-- ─── Modal Edit Kurikulum ─── -->
<Dialog.Root bind:open={showEditProfile}>
  <Dialog.Content>
    <div class="space-y-4">
      <h2 class="text-base font-semibold text-foreground">Edit Kurikulum</h2>
      <div class="grid gap-3 sm:grid-cols-2">
        <div class="space-y-1"><label class="text-xs font-medium text-muted-foreground">Kode *</label><Input bind:value={editProfileForm.code} /></div>
        <div class="space-y-1"><label class="text-xs font-medium text-muted-foreground">Nama *</label><Input bind:value={editProfileForm.name} /></div>
        <div class="space-y-1 sm:col-span-2"><label class="text-xs font-medium text-muted-foreground">Regulasi</label><Input bind:value={editProfileForm.regulation_reference} /></div>
        <div class="space-y-1">
          <label class="text-xs font-medium text-muted-foreground">Tingkat</label>
          <select bind:value={editProfileForm.education_level} class="flex h-9 w-full rounded-lg border border-input bg-background px-3 text-sm"><option value="MTs">MTs</option><option value="MA">MA</option><option value="MI">MI</option></select>
        </div>
        <div class="space-y-1">
          <label class="text-xs font-medium text-muted-foreground">Status</label>
          <select bind:value={editProfileForm.status} class="flex h-9 w-full rounded-lg border border-input bg-background px-3 text-sm"><option value="draft">Draft</option><option value="active">Aktif</option><option value="archived">Arsip</option></select>
        </div>
        <div class="space-y-1 sm:col-span-2"><label class="text-xs font-medium text-muted-foreground">Catatan</label><Input bind:value={editProfileForm.notes} /></div>
      </div>
      <div class="flex justify-end gap-2 pt-1">
        <Button variant="outline" onclick={() => (showEditProfile = false)}>Batal</Button>
        <LoadingButton onclick={() => void submitEditProfile()} loading={editProfileLoading}>Simpan</LoadingButton>
      </div>
    </div>
  </Dialog.Content>
</Dialog.Root>

<!-- ─── Modal Tambah Alokasi ─── -->
<Dialog.Root bind:open={showAddAlloc}>
  <Dialog.Content>
    <div class="space-y-4">
      <div><h2 class="text-base font-semibold text-foreground">Tambah Mapel</h2>
      <p class="text-sm text-muted-foreground">{selectedProfile?.code} — Tingkat {addAllocForm.level}</p></div>
      <div class="grid gap-3 sm:grid-cols-2">
        <div class="space-y-1 sm:col-span-2"><label class="text-xs font-medium text-muted-foreground">Nama Mapel</label><Input bind:value={addAllocForm.subject_name} placeholder="Matematika" /></div>
        <div class="space-y-1">
          <label class="text-xs font-medium text-muted-foreground">Tingkat</label>
          <select bind:value={addAllocForm.level} class="flex h-9 w-full rounded-lg border border-input bg-background px-3 text-sm">{#each levels as lv}<option value={lv}>Kelas {lv}</option>{/each}</select>
        </div>
        <div class="space-y-1">
          <label class="text-xs font-medium text-muted-foreground">Kelompok</label>
          <select bind:value={addAllocForm.group} class="flex h-9 w-full rounded-lg border border-input bg-background px-3 text-sm">{#each groups as g}<option value={g}>{g}</option>{/each}</select>
        </div>
        <div class="space-y-1"><label class="text-xs font-medium text-muted-foreground">Jam Intra</label><Input type="number" bind:value={addAllocForm.intra} min="0" step="0.5" /></div>
        <div class="space-y-1"><label class="text-xs font-medium text-muted-foreground">Jam Koku</label><Input type="number" bind:value={addAllocForm.koku} min="0" step="0.5" /></div>
      </div>
      <div class="flex justify-end gap-2 pt-1">
        <Button variant="outline" onclick={() => (showAddAlloc = false)}>Batal</Button>
        <LoadingButton onclick={() => void submitAddAlloc()} loading={addAllocLoading}>Simpan</LoadingButton>
      </div>
    </div>
  </Dialog.Content>
</Dialog.Root>

<!-- ─── Modal Edit Alokasi ─── -->
<Dialog.Root bind:open={showEditAlloc}>
  <Dialog.Content>
    <div class="space-y-4">
      <h2 class="text-base font-semibold text-foreground">Edit Alokasi</h2>
      <div class="grid gap-3 sm:grid-cols-2">
        <div class="space-y-1">
          <label class="text-xs font-medium text-muted-foreground">Tingkat</label>
          <select bind:value={editAllocForm.level} class="flex h-9 w-full rounded-lg border border-input bg-background px-3 text-sm">{#each levels as lv}<option value={lv}>Kelas {lv}</option>{/each}</select>
        </div>
        <div class="space-y-1">
          <label class="text-xs font-medium text-muted-foreground">Kelompok</label>
          <select bind:value={editAllocForm.group} class="flex h-9 w-full rounded-lg border border-input bg-background px-3 text-sm">{#each groups as g}<option value={g}>{g}</option>{/each}</select>
        </div>
        <div class="space-y-1"><label class="text-xs font-medium text-muted-foreground">Jam Intra</label><Input type="number" bind:value={editAllocForm.intra} min="0" step="0.5" /></div>
        <div class="space-y-1"><label class="text-xs font-medium text-muted-foreground">Jam Koku</label><Input type="number" bind:value={editAllocForm.koku} min="0" step="0.5" /></div>
      </div>
      <div class="flex justify-end gap-2 pt-1">
        <Button variant="outline" onclick={() => (showEditAlloc = false)}>Batal</Button>
        <LoadingButton onclick={() => void submitEditAlloc()} loading={editAllocLoading}>Simpan</LoadingButton>
      </div>
    </div>
  </Dialog.Content>
</Dialog.Root>
