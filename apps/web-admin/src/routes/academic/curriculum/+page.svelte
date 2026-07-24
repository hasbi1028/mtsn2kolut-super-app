<script lang="ts">
  import { Badge } from '$lib/components/ui/badge';
  import { Button } from '$lib/components/ui/button';
  import * as Card from '$lib/components/ui/card';
  import * as Tabs from '$lib/components/ui/tabs';
  import * as Dialog from '$lib/components/ui/dialog';
  import { Input } from '$lib/components/ui/input';
  import { toast } from '$lib/components/ui/sonner';
  import LoadingButton from '$lib/components/LoadingButton.svelte';

  let { data } = $props();

  type Profile = {
    id: string; code: string; name: string; regulation_reference: string;
    education_level: string; effective_academic_year_id: string;
    status: string; notes: string; created_at: string; updated_at: string;
  };
  type Allocation = {
    id: string; curriculum_profile_id: string; curriculum_code: string; curriculum_name: string;
    subject_id: string; subject_code: string; subject_name: string;
    level: string; subject_group: string;
    intra_weekly_hours: number; koku_weekly_hours: number; total_weekly_hours: number;
    counts_for_schedule: boolean; counts_for_report: boolean;
    counts_for_assessment: boolean; counts_for_ranking: boolean;
    is_required: boolean; notes: string;
  };
  type Rombel = {
    id: string; code: string; name: string; level: string; is_active: boolean;
    academic_year_id: string; academic_year_name: string;
  };
  type Assignment = {
    id: string; class_id: string; class_code: string; class_name: string; class_level: string;
    curriculum_profile_id: string; curriculum_code: string; curriculum_name: string;
    is_active: boolean; notes: string;
  };

  let profiles = $state<Profile[]>(data.profiles ?? []);
  let rombels = $state<Rombel[]>(data.rombels ?? []);
  let selectedProfile = $state<Profile | null>(null);
  let allocations = $state<Allocation[]>([]);
  let assignments = $state<Assignment[]>([]);
  let selectedLevel = $state('VII');
  let loadingAlloc = $state(false);
  let loadingAssign = $state(false);

  // Tambah profile
  let showTambah = $state(false);
  let tambahForm = $state({ code: '', name: '', regulation_reference: '', education_level: 'MTs', status: 'draft', notes: '' });
  let tambahLoading = $state(false);
  let actLoading = $state<string | null>(null);
  let delLoading = $state<string | null>(null);

  // Tambah allocation
  let showAlloc = $state(false);
  let allocForm = $state({ subject_name: '', level: 'VII', group: 'wajib', intra: 0, koku: 0, total: 0, notes: '' });
  let allocLoading = $state(false);

  const levels = ['VII', 'VIII', 'IX'];
  const groups = ['wajib', 'pilihan', 'muatan_lokal', 'layanan', 'kokurikuler', 'kegiatan'];

  async function refreshProfiles() {
    const res = await fetch('/api/academic/curriculum/profiles');
    if (res.ok) { const p = await res.json(); profiles = p.items ?? []; }
  }

  async function selectProfile(p: Profile) {
    selectedProfile = p;
    selectedLevel = 'VII';
    loadAllocations();
    loadAssignments();
  }

  async function loadAllocations() {
    if (!selectedProfile) return;
    loadingAlloc = true;
    const res = await fetch(`/api/academic/curriculum/profiles/${selectedProfile.id}/allocations?level=${selectedLevel}`);
    if (res.ok) { const p = await res.json(); allocations = p.items ?? []; }
    else allocations = [];
    loadingAlloc = false;
  }

  async function loadAssignments() {
    if (!selectedProfile) return;
    loadingAssign = true;
    const res = await fetch(`/api/academic/curriculum/assignments?class_id=`);
    if (res.ok) {
      const p = await res.json();
      const all: Assignment[] = p.items ?? [];
      assignments = all.filter(a => a.curriculum_profile_id === selectedProfile!.id);
    } else assignments = [];
    loadingAssign = false;
  }

  $effect(() => { if (selectedLevel && selectedProfile) loadAllocations(); });

  async function submitTambah() {
    if (!tambahForm.code || !tambahForm.name) { toast.error('Kode dan nama wajib diisi'); return; }
    tambahLoading = true;
    try {
      const res = await fetch('/api/academic/curriculum/profiles', {
        method: 'POST', headers: { 'content-type': 'application/json' },
        body: JSON.stringify(tambahForm),
      });
      if (res.ok) {
        toast.success(`Kurikulum "${tambahForm.code}" dibuat`);
        showTambah = false;
        tambahForm = { code: '', name: '', regulation_reference: '', education_level: 'MTs', status: 'draft', notes: '' };
        await refreshProfiles();
      } else {
        const e = await res.json().catch(() => ({}));
        toast.error(e?.error || 'Gagal membuat kurikulum');
      }
    } catch { toast.error('Gagal membuat kurikulum'); }
    finally { tambahLoading = false; }
  }

  async function activateProfile(id: string) {
    actLoading = id;
    try {
      const res = await fetch(`/api/academic/curriculum/profiles/${id}/activate`, { method: 'POST' });
      if (res.ok) { toast.success('Kurikulum diaktifkan'); await refreshProfiles(); }
      else toast.error('Gagal mengaktifkan');
    } catch { toast.error('Gagal mengaktifkan'); }
    finally { actLoading = null; }
  }

  async function deleteProfile(id: string, name: string) {
    if (!confirm(`Hapus kurikulum "${name}"?`)) return;
    delLoading = id;
    try {
      const res = await fetch(`/api/academic/curriculum/profiles/${id}/activate`, { method: 'DELETE' });
      if (res.ok || res.status === 204) {
        toast.success(`Kurikulum "${name}" dihapus`);
        if (selectedProfile?.id === id) selectedProfile = null;
        await refreshProfiles();
      } else toast.error('Gagal menghapus');
    } catch { toast.error('Gagal menghapus'); }
    finally { delLoading = null; }
  }

  async function submitAlloc() {
    if (!allocForm.subject_name || !selectedProfile) { toast.error('Nama mapel wajib diisi'); return; }
    allocLoading = true;
    try {
      const res = await fetch(`/api/academic/curriculum/profiles/${selectedProfile.id}/allocations`, {
        method: 'POST', headers: { 'content-type': 'application/json' },
        body: JSON.stringify({
          subject_id: '00000000-0000-0000-0000-000000000000',
          level: allocForm.level,
          subject_group: allocForm.group,
          intra_weekly_hours: allocForm.intra,
          koku_weekly_hours: allocForm.koku,
          total_weekly_hours: allocForm.total || allocForm.intra + allocForm.koku,
          notes: allocForm.notes,
        }),
      });
      if (res.ok) {
        toast.success(`Mapel ditambahkan ke ${selectedProfile.code}`);
        showAlloc = false;
        allocForm = { subject_name: '', level: selectedLevel, group: 'wajib', intra: 0, koku: 0, total: 0, notes: '' };
        await loadAllocations();
      } else {
        const e = await res.json().catch(() => ({}));
        toast.error(e?.error || 'Gagal menambah mapel');
      }
    } catch { toast.error('Gagal menambah mapel'); }
    finally { allocLoading = false; }
  }

  async function deleteAlloc(id: string) {
    if (!confirm('Hapus alokasi mapel ini?')) return;
    const res = await fetch(`/api/academic/curriculum/allocations/${id}`, { method: 'DELETE' });
    if (res.ok || res.status === 204) { toast.success('Alokasi dihapus'); await loadAllocations(); }
    else toast.error('Gagal menghapus');
  }

  async function assignRombel(classId: string) {
    if (!selectedProfile) return;
    const res = await fetch('/api/academic/curriculum/assignments', {
      method: 'POST', headers: { 'content-type': 'application/json' },
      body: JSON.stringify({ class_id: classId, curriculum_profile_id: selectedProfile.id, notes: '' }),
    });
    if (res.ok) { toast.success('Rombel di-assign'); await loadAssignments(); }
    else { const e = await res.json().catch(() => ({})); toast.error(e?.error || 'Gagal assign'); }
  }

  async function unassignRombel(id: string) {
    const res = await fetch(`/api/academic/curriculum/assignments/${id}`, { method: 'DELETE' });
    if (res.ok || res.status === 204) { toast.success('Assignment dihapus'); await loadAssignments(); }
    else toast.error('Gagal hapus assignment');
  }

  function statusBadge(s: string) {
    if (s === 'active') return 'default';
    if (s === 'archived') return 'secondary';
    return 'outline';
  }
</script>

<svelte:head><title>Kurikulum — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6">
  <div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
    <div>
      <h1 class="text-2xl font-black text-foreground">Kurikulum</h1>
      <p class="mt-1 text-sm text-muted-foreground">Kelola profil kurikulum, alokasi mata pelajaran, dan assign ke rombel.</p>
    </div>
    <Button onclick={() => (showTambah = true)}>+ Tambah Kurikulum</Button>
  </div>

  {#if profiles.length === 0}
    <Card.Root>
      <Card.Content class="p-8 text-center">
        <p class="text-sm text-muted-foreground">Belum ada kurikulum. Tambah kurikulum baru untuk memulai.</p>
      </Card.Content>
    </Card.Root>
  {:else}
    <div class="grid gap-6 lg:grid-cols-3">
      <div class="space-y-3 lg:col-span-1">
        <h2 class="text-sm font-semibold text-muted-foreground">Daftar Kurikulum</h2>
        {#each profiles as p (p.id)}
          <Card.Root
            class="cursor-pointer transition-all hover:border-primary/30 {selectedProfile?.id === p.id ? 'border-primary ring-1 ring-primary/20' : ''} {p.status === 'archived' ? 'opacity-60' : ''}"
            onclick={() => selectProfile(p)}
          >
            <Card.Content class="p-4">
              <div class="flex items-start justify-between gap-2">
                <div class="min-w-0">
                  <div class="flex items-center gap-2">
                    <h3 class="text-sm font-semibold text-foreground truncate">{p.code}</h3>
                    <Badge variant={statusBadge(p.status)} class="text-[9px] px-1.5">{p.status}</Badge>
                  </div>
                  <p class="mt-0.5 text-xs text-muted-foreground truncate">{p.name}</p>
                </div>
              </div>
            </Card.Content>
          </Card.Root>
        {/each}
      </div>

      <div class="lg:col-span-2">
        {#if selectedProfile}
          <Card.Root>
            <Card.Header class="px-5 pt-4 pb-3">
              <div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
                <div>
                  <Card.Title class="text-base">{selectedProfile.name}</Card.Title>
                  <Card.Description>{selectedProfile.regulation_reference || selectedProfile.code}</Card.Description>
                </div>
                <div class="flex items-center gap-2">
                  {#if selectedProfile.status !== 'active'}
                    <Button size="sm" onclick={() => void activateProfile(selectedProfile.id)} disabled={actLoading === selectedProfile.id}>Aktifkan</Button>
                  {/if}
                  <Button size="sm" variant="outline" class="text-destructive" onclick={() => void deleteProfile(selectedProfile.id, selectedProfile.name)} disabled={delLoading === selectedProfile.id}>Hapus</Button>
                </div>
              </div>
            </Card.Header>
            <Card.Content class="space-y-6 p-5 pt-0">
              <Tabs.Root value="allocations" class="w-full">
                <Tabs.List>
                  <Tabs.Trigger value="allocations">Alokasi Mapel</Tabs.Trigger>
                  <Tabs.Trigger value="assignments">Assign Rombel</Tabs.Trigger>
                </Tabs.List>

                <Tabs.Content value="allocations" class="space-y-4 pt-4">
                  <div class="flex items-center justify-between gap-3">
                    <div class="flex gap-1">
                      {#each levels as lv}
                        <Button size="sm" variant={selectedLevel === lv ? 'default' : 'outline'} onclick={() => (selectedLevel = lv)}>{lv}</Button>
                      {/each}
                    </div>
                    <Button size="sm" onclick={() => (showAlloc = true)}>+ Tambah Mapel</Button>
                  </div>

                  {#if loadingAlloc}
                    <p class="text-sm text-muted-foreground">Memuat...</p>
                  {:else if allocations.length === 0}
                    <p class="text-sm text-muted-foreground py-4 text-center">Belum ada mapel untuk tingkat {selectedLevel}.</p>
                  {:else}
                    <div class="overflow-x-auto rounded-lg border border-border">
                      <table class="w-full text-sm">
                        <thead class="bg-muted/30 text-muted-foreground text-xs uppercase">
                          <tr>
                            <th class="px-3 py-2 text-left">Mapel</th>
                            <th class="px-3 py-2 text-center">Kelompok</th>
                            <th class="px-3 py-2 text-center">Jam Intrakur</th>
                            <th class="px-3 py-2 text-center">Jam Kokur</th>
                            <th class="px-3 py-2 text-center">Total</th>
                            <th class="px-3 py-2 text-right">Aksi</th>
                          </tr>
                        </thead>
                        <tbody class="divide-y divide-border">
                          {#each allocations as a (a.id)}
                            <tr class="hover:bg-muted/20">
                              <td class="px-3 py-2 font-medium text-foreground">{a.subject_name}</td>
                              <td class="px-3 py-2 text-center text-muted-foreground text-xs">{a.subject_group}</td>
                              <td class="px-3 py-2 text-center">{a.intra_weekly_hours}</td>
                              <td class="px-3 py-2 text-center">{a.koku_weekly_hours}</td>
                              <td class="px-3 py-2 text-center font-semibold">{a.total_weekly_hours}</td>
                              <td class="px-3 py-2 text-right">
                                <Button size="sm" variant="outline" class="text-destructive text-xs px-2" onclick={() => void deleteAlloc(a.id)}>Hapus</Button>
                              </td>
                            </tr>
                          {/each}
                        </tbody>
                      </table>
                    </div>
                  {/if}
                </Tabs.Content>

                <Tabs.Content value="assignments" class="space-y-4 pt-4">
                  {#if loadingAssign}
                    <p class="text-sm text-muted-foreground">Memuat...</p>
                  {:else}
                    <div class="overflow-x-auto rounded-lg border border-border">
                      <table class="w-full text-sm">
                        <thead class="bg-muted/30 text-muted-foreground text-xs uppercase">
                          <tr>
                            <th class="px-3 py-2 text-left">Rombel</th>
                            <th class="px-3 py-2 text-center">Tingkat</th>
                            <th class="px-3 py-2 text-center">Kurikulum</th>
                            <th class="px-3 py-2 text-right">Aksi</th>
                          </tr>
                        </thead>
                        <tbody class="divide-y divide-border">
                          {#each rombels as r (r.id)}
                            {@const assigned = assignments.find(a => a.class_id === r.id)}
                            <tr class="hover:bg-muted/20">
                              <td class="px-3 py-2 font-medium text-foreground">{r.name}</td>
                              <td class="px-3 py-2 text-center text-muted-foreground">{r.level}</td>
                              <td class="px-3 py-2 text-center">
                                {#if assigned}
                                  <Badge variant="default" class="text-[9px] bg-primary/10 text-primary">{assigned.curriculum_code}</Badge>
                                {:else}
                                  <span class="text-muted-foreground italic">—</span>
                                {/if}
                              </td>
                              <td class="px-3 py-2 text-right">
                                {#if assigned}
                                  <Button size="sm" variant="outline" class="text-destructive text-xs px-2" onclick={() => void unassignRombel(assigned.id)}>Lepas</Button>
                                {:else}
                                  <Button size="sm" variant="outline" class="text-xs px-2" onclick={() => void assignRombel(r.id)}>Assign</Button>
                                {/if}
                              </td>
                            </tr>
                          {/each}
                        </tbody>
                      </table>
                    </div>
                  {/if}
                </Tabs.Content>
              </Tabs.Root>
            </Card.Content>
          </Card.Root>
        {:else}
          <Card.Root>
            <Card.Content class="p-12 text-center">
              <p class="text-sm text-muted-foreground">Pilih kurikulum dari daftar di samping untuk melihat detail.</p>
            </Card.Content>
          </Card.Root>
        {/if}
      </div>
    </div>
  {/if}
</div>

<Dialog.Root bind:open={showTambah}>
  <Dialog.Content>
    <div class="space-y-4">
      <div><h2 class="text-base font-semibold text-foreground">Tambah Kurikulum</h2></div>
      <div class="grid gap-4 sm:grid-cols-2">
        <div class="space-y-1.5"><label class="text-xs font-medium text-muted-foreground">Kode Kurikulum</label><Input bind:value={tambahForm.code} placeholder="Contoh: KMA-1503-2025" /></div>
        <div class="space-y-1.5"><label class="text-xs font-medium text-muted-foreground">Nama Kurikulum</label><Input bind:value={tambahForm.name} placeholder="Contoh: Kurikulum Merdeka" /></div>
        <div class="space-y-1.5"><label class="text-xs font-medium text-muted-foreground">Regulasi</label><Input bind:value={tambahForm.regulation_reference} placeholder="Contoh: KMA 1503 Tahun 2025" /></div>
        <div class="space-y-1.5">
          <label class="text-xs font-medium text-muted-foreground">Tingkat Pendidikan</label>
          <select bind:value={tambahForm.education_level} class="flex h-10 w-full rounded-lg border border-input bg-background px-3 text-sm">
            <option value="MTs">MTs</option>
            <option value="MA">MA</option>
            <option value="MI">MI</option>
          </select>
        </div>
        <div class="space-y-1.5">
          <label class="text-xs font-medium text-muted-foreground">Status</label>
          <select bind:value={tambahForm.status} class="flex h-10 w-full rounded-lg border border-input bg-background px-3 text-sm">
            <option value="draft">Draft</option>
            <option value="active">Aktif</option>
          </select>
        </div>
      </div>
      <div class="flex justify-end gap-2">
        <Button variant="outline" onclick={() => (showTambah = false)}>Batal</Button>
        <LoadingButton onclick={() => void submitTambah()} loading={tambahLoading}>Simpan</LoadingButton>
      </div>
    </div>
  </Dialog.Content>
</Dialog.Root>

<Dialog.Root bind:open={showAlloc}>
  <Dialog.Content>
    <div class="space-y-4">
      <div><h2 class="text-base font-semibold text-foreground">Tambah Mata Pelajaran</h2>
        <p class="text-sm text-muted-foreground">{selectedProfile?.code} — Tingkat {allocForm.level}</p>
      </div>
      <div class="grid gap-4 sm:grid-cols-2">
        <div class="space-y-1.5 sm:col-span-2"><label class="text-xs font-medium text-muted-foreground">Nama Mata Pelajaran</label><Input bind:value={allocForm.subject_name} placeholder="Contoh: Matematika" /></div>
        <div class="space-y-1.5">
          <label class="text-xs font-medium text-muted-foreground">Tingkat</label>
          <select bind:value={allocForm.level} class="flex h-10 w-full rounded-lg border border-input bg-background px-3 text-sm">
            {#each levels as lv}<option value={lv}>Kelas {lv}</option>{/each}
          </select>
        </div>
        <div class="space-y-1.5">
          <label class="text-xs font-medium text-muted-foreground">Kelompok</label>
          <select bind:value={allocForm.group} class="flex h-10 w-full rounded-lg border border-input bg-background px-3 text-sm">
            {#each groups as g}<option value={g}>{g}</option>{/each}
          </select>
        </div>
        <div class="space-y-1.5"><label class="text-xs font-medium text-muted-foreground">Jam Intrakurikuler</label><Input type="number" bind:value={allocForm.intra} min="0" step="0.5" /></div>
        <div class="space-y-1.5"><label class="text-xs font-medium text-muted-foreground">Jam Kokurikuler</label><Input type="number" bind:value={allocForm.koku} min="0" step="0.5" /></div>
        <div class="space-y-1.5"><label class="text-xs font-medium text-muted-foreground">Total Jam</label><Input type="number" bind:value={allocForm.total} min="0" step="0.5" /></div>
      </div>
      <div class="flex justify-end gap-2">
        <Button variant="outline" onclick={() => (showAlloc = false)}>Batal</Button>
        <LoadingButton onclick={() => void submitAlloc()} loading={allocLoading}>Simpan</LoadingButton>
      </div>
    </div>
  </Dialog.Content>
</Dialog.Root>
