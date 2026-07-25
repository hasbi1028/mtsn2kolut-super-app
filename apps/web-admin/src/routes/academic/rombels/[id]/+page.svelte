<script lang="ts">
  import { Button } from '$lib/components/ui/button';
  import * as Card from '$lib/components/ui/card';
  import * as Dialog from '$lib/components/ui/dialog';
  import { Input } from '$lib/components/ui/input';
  import { toast } from '$lib/components/ui/sonner';
  import LoadingButton from '$lib/components/LoadingButton.svelte';

  let { data } = $props();

  type Student = {
    student_id: string;
    nis: string;
    nisn: string;
    student_name: string;
    gender: string;
    is_active: boolean;
    status: string;
  };

  type RombelDetail = {
    id: string;
    code: string;
    name: string;
    level: string;
    is_active: boolean;
    academic_year_name: string;
  };

  type HomeroomAssignment = {
    id: string;
    employee_id: string;
    employee_name: string;
    is_active: boolean;
  };

  type RombelOption = {
    id: string;
    code: string;
    name: string;
    level: string;
  };

  let detail = $state<RombelDetail | null>(data.detail);
  let students = $state<Student[]>(data.students ?? []);
  let loading = $state(false);
  let homeroom = $state<HomeroomAssignment | null>(null);
  let homeroomLoading = $state(false);

  // — Assign student dialog
  let showAssign = $state(false);
  let unassignedList = $state<{ id: string; nis: string; nisn: string; nama: string; gender: string }[]>([]);
  let selectedStudents = $state<Set<string>>(new Set());
  let assignLoading = $state(false);
  let searchUnassigned = $state('');

  // — Pindah Kelas dialog
  let showPindah = $state(false);
  let pindahTarget = $state<string>('');
  let pindahStudentName = $state('');
  let pindahStudentId = $state('');
  let rombelOptions = $state<RombelOption[]>([]);
  let pindahLoading = $state(false);

  // — Wali Kelas dialog
  let showWaliKelas = $state(false);
  let waliKelasSearch = $state('');
  let employees = $state<{ id: string; nama: string; nip: string }[]>([]);
  let waliLoading = $state(false);
  let waliSelected = $state<string>('');

  async function loadHomeroom() {
    if (!detail) return;
    homeroomLoading = true;
    try {
      const res = await fetch(`/api/academic/rombels/${detail.id}/homeroom`);
      if (res.ok) {
        const p = await res.json();
        homeroom = p ?? null;
      }
    } catch { /* ignore */ }
    finally { homeroomLoading = false; }
  }

  async function loadStudents() {
    if (!detail) return;
    loading = true;
    try {
      const res = await fetch(`/api/academic/rombels/${detail.id}/students`);
      if (res.ok) {
        const p = await res.json();
        students = p.items ?? [];
      }
    } catch { toast.error('Gagal memuat siswa'); }
    finally { loading = false; }
  }

  async function openAssignDialog() {
    showAssign = true;
    selectedStudents = new Set();
    try {
      const res = await fetch('/api/academic/rombels/unassigned-students');
      if (res.ok) {
        const p = await res.json();
        unassignedList = p.items ?? [];
      } else { unassignedList = []; }
    } catch { unassignedList = []; toast.error('Gagal memuat daftar siswa'); }
  }

  function toggleSelectStudent(id: string) {
    const next = new Set(selectedStudents);
    if (next.has(id)) next.delete(id); else next.add(id);
    selectedStudents = next;
  }

  async function submitAssign() {
    const ids = Array.from(selectedStudents);
    if (ids.length === 0) { toast.error('Pilih siswa terlebih dahulu'); return; }
    if (!detail) return;
    assignLoading = true;
    try {
      const res = await fetch(`/api/academic/rombels/${detail.id}/students`, {
        method: 'POST',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify({ student_ids: ids }),
      });
      if (res.ok) {
        toast.success(`${ids.length} siswa ditambahkan`);
        showAssign = false;
        await loadStudents();
      } else {
        const body = await res.json().catch(() => ({}));
        toast.error(body?.error || 'Gagal');
      }
    } catch { toast.error('Gagal'); }
    finally { assignLoading = false; }
  }

  async function removeStudent(studentId: string, studentName: string) {
    if (!confirm(`Hapus "${studentName}" dari rombel?`)) return;
    try {
      const res = await fetch(`/api/academic/rombels/${detail!.id}/students/${studentId}`, { method: 'DELETE' });
      if (res.ok || res.status === 204) {
        toast.success(`${studentName} dihapus dari rombel`);
        await loadStudents();
      } else { toast.error('Gagal menghapus siswa'); }
    } catch { toast.error('Gagal'); }
  }

  // — Pindah Kelas
  async function openPindahDialog(studentId: string, studentName: string) {
    pindahStudentId = studentId;
    pindahStudentName = studentName;
    pindahTarget = '';
    showPindah = true;

    // Load all rombels except current
    if (rombelOptions.length === 0) {
      try {
        const res = await fetch('/api/academic/rombels');
        if (res.ok) {
          const p = await res.json();
          rombelOptions = (p.items ?? []).filter((r: RombelOption) => r.id !== detail?.id);
        }
      } catch { toast.error('Gagal memuat daftar rombel'); }
    }
  }

  async function submitPindah() {
    if (!pindahTarget) { toast.error('Pilih rombel tujuan'); return; }
    pindahLoading = true;
    try {
      // Assign student to new class
      const res = await fetch(`/api/academic/rombels/${pindahTarget}/students`, {
        method: 'POST',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify({ student_ids: [pindahStudentId] }),
      });
      if (res.ok) {
        toast.success(`${pindahStudentName} pindah ke ${rombelOptions.find(r => r.id === pindahTarget)?.code ?? 'rombel lain'}`);
        showPindah = false;
        await loadStudents();
      } else {
        const body = await res.json().catch(() => ({}));
        toast.error(body?.error || 'Gagal memindahkan siswa');
      }
    } catch { toast.error('Gagal'); }
    finally { pindahLoading = false; }
  }

  // — Wali Kelas
  async function openWaliDialog() {
    showWaliKelas = true;
    waliKelasSearch = '';
    waliSelected = homeroom?.employee_id ?? '';
    try {
      const res = await fetch('/api/employees');
      if (res.ok) {
        const p = await res.json();
        // Handle both {items: [...]} and {data: [...]} and plain [...]
        const list = Array.isArray(p) ? p : (p?.items ?? p?.data ?? []);
        employees = list.map((e: any) => ({
          id: e.id, nama: e.nama, nip: e.nip ?? ''
        }));
      }
    } catch { toast.error('Gagal memuat pegawai'); }
  }

  async function submitWaliKelas() {
    if (!waliSelected) { toast.error('Pilih wali kelas'); return; }
    if (!detail) return;
    waliLoading = true;
    try {
      const res = await fetch(`/api/academic/rombels/${detail.id}/homeroom`, {
        method: 'POST',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify({ employee_id: waliSelected }),
      });
      if (res.ok) {
        toast.success('Wali kelas berhasil diperbarui');
        showWaliKelas = false;
        await loadHomeroom();
      } else {
        const body = await res.json().catch(() => ({}));
        toast.error(body?.error || 'Gagal');
      }
    } catch { toast.error('Gagal'); }
    finally { waliLoading = false; }
  }

  let filteredEmployees = $derived(
    waliKelasSearch
      ? employees.filter(e =>
          e.nama.toLowerCase().includes(waliKelasSearch.toLowerCase()) ||
          e.nip.includes(waliKelasSearch)
        )
      : employees
  );

  let filteredUnassigned = $derived(
    searchUnassigned
      ? unassignedList.filter(s =>
          s.nama.toLowerCase().includes(searchUnassigned.toLowerCase()) ||
          s.nis.includes(searchUnassigned)
        )
      : unassignedList
  );

  // Load homeroom on init
  $effect(() => {
    if (detail) { loadHomeroom(); }
  });
</script>

<svelte:head><title>{detail?.code ?? 'Rombel'} — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6">
  <!-- Header -->
  <div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
    <div>
      <a href="/academic/rombels" class="text-sm text-primary hover:underline">&larr; Kembali ke Rombel</a>
      {#if detail}
        <h1 class="text-2xl font-black text-foreground mt-1">{detail.code} — {detail.name}</h1>
        <p class="text-sm text-muted-foreground">{detail.level} &middot; {detail.academic_year_name}</p>
      {:else}
        <h1 class="text-2xl font-black text-foreground mt-1">Rombel tidak ditemukan</h1>
      {/if}
    </div>
    <Button onclick={openAssignDialog}>+ Tambah Siswa</Button>
  </div>

  <!-- Wali Kelas Card -->
  <Card.Root class="border-primary/20">
    <Card.Content class="p-4 flex items-center justify-between gap-3">
      <div class="flex items-center gap-3 min-w-0">
        <span class="text-lg shrink-0">👨‍🏫</span>
        <div class="min-w-0">
          <p class="text-xs font-medium text-muted-foreground uppercase tracking-wider">Wali Kelas</p>
          {#if homeroomLoading}
            <p class="text-sm text-muted-foreground italic">Memuat...</p>
          {:else if homeroom}
            <p class="text-sm font-semibold truncate">{homeroom.employee_name}</p>
          {:else}
            <p class="text-sm text-muted-foreground italic">Belum ada wali kelas</p>
          {/if}
        </div>
      </div>
      <Button size="sm" variant="outline" onclick={openWaliDialog}>
        {homeroom ? 'Ganti' : 'Pilih'}
      </Button>
    </Card.Content>
  </Card.Root>

  <!-- Student list -->
  {#if loading}
    <p class="text-sm text-muted-foreground">Memuat siswa...</p>
  {:else if students.length === 0}
    <Card.Root>
      <Card.Content class="p-8 text-center">
        <p class="text-sm text-muted-foreground">Belum ada siswa di rombel ini. Klik "+ Tambah Siswa" untuk menambahkan.</p>
      </Card.Content>
    </Card.Root>
  {:else}
    <div class="flex items-center justify-between">
      <p class="text-sm text-muted-foreground">{students.length} siswa</p>
    </div>
    <Card.Root>
      <Card.Content class="p-0">
        <!-- Desktop table -->
        <div class="hidden md:block overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="bg-muted/30 text-muted-foreground text-xs uppercase">
                <th class="px-4 py-2.5 text-left">No</th>
                <th class="px-4 py-2.5 text-left">Nama</th>
                <th class="px-4 py-2.5 text-left">NIS</th>
                <th class="px-4 py-2.5 text-left">NISN</th>
                <th class="px-4 py-2.5 text-center">JK</th>
                <th class="px-4 py-2.5 text-center">Status</th>
                <th class="px-4 py-2.5 text-right">Aksi</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-border">
              {#each students as s, idx (s.student_id)}
                <tr class="hover:bg-muted/10">
                  <td class="px-4 py-2.5 text-xs text-muted-foreground">{idx + 1}</td>
                  <td class="px-4 py-2.5 text-sm font-medium">{s.student_name}</td>
                  <td class="px-4 py-2.5 text-xs text-muted-foreground">{s.nis}</td>
                  <td class="px-4 py-2.5 text-xs text-muted-foreground">{s.nisn || '—'}</td>
                  <td class="px-4 py-2.5 text-xs text-center">{s.gender === 'L' ? 'L' : 'P'}</td>
                  <td class="px-4 py-2.5 text-center">
                    {#if s.is_active}
                      <span class="text-[10px] font-semibold text-green-600">Aktif</span>
                    {:else}
                      <span class="text-[10px] font-semibold text-muted-foreground">Nonaktif</span>
                    {/if}
                  </td>
                  <td class="px-4 py-2.5 text-right">
                    <div class="inline-flex items-center gap-1">
                      <button
                        class="text-xs text-primary hover:underline"
                        onclick={() => openPindahDialog(s.student_id, s.student_name)}
                      >Pindah</button>
                      <button
                        class="text-xs text-destructive hover:underline"
                        onclick={() => removeStudent(s.student_id, s.student_name)}
                      >Kelupakan</button>
                    </div>
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
        <!-- Mobile cards -->
        <div class="md:hidden space-y-1 p-3">
          {#each students as s, idx (s.student_id)}
            <div class="flex items-center justify-between rounded-lg border border-border px-3 py-2">
              <div class="min-w-0 flex-1">
                <p class="text-sm font-medium truncate">{s.student_name}</p>
                <p class="text-xs text-muted-foreground">{s.nis}</p>
              </div>
              <div class="flex items-center gap-1 shrink-0 ml-2">
                <button
                  class="text-xs text-primary hover:underline"
                  onclick={() => openPindahDialog(s.student_id, s.student_name)}
                >Pindah</button>
                <button
                  class="text-xs text-destructive hover:underline"
                  onclick={() => removeStudent(s.student_id, s.student_name)}
                >Kelupakan</button>
              </div>
            </div>
          {/each}
        </div>
      </Card.Content>
    </Card.Root>
  {/if}
</div>

<!-- Dialog Assign Siswa -->
<Dialog.Root bind:open={showAssign}>
  <Dialog.Content class="max-w-lg">
    <div class="space-y-4">
      <h2 class="text-base font-semibold">Tambah Siswa ke {detail?.code ?? 'Rombel'}</h2>
      <div><Input placeholder="Cari nama atau NIS..." bind:value={searchUnassigned} /></div>
      {#if unassignedList.length === 0}
        <p class="text-sm text-muted-foreground text-center py-4">Semua siswa sudah memiliki rombel.</p>
      {:else}
        <div class="max-h-64 overflow-y-auto space-y-1">
          {#each filteredUnassigned as s (s.id)}
            <label class="flex items-center gap-3 rounded-lg border border-border px-3 py-2 hover:bg-muted/10 cursor-pointer">
              <input type="checkbox" checked={selectedStudents.has(s.id)} onchange={() => toggleSelectStudent(s.id)} class="size-4 accent-primary" />
              <div class="min-w-0 flex-1">
                <span class="text-sm font-medium">{s.nama}</span>
                <span class="text-xs text-muted-foreground ml-2">NIS: {s.nis}</span>
              </div>
            </label>
          {/each}
        </div>
      {/if}
      <div class="flex justify-between items-center pt-2">
        <p class="text-xs text-muted-foreground">{selectedStudents.size} siswa dipilih</p>
        <div class="flex gap-2">
          <Button variant="outline" onclick={() => (showAssign = false)}>Batal</Button>
          <LoadingButton onclick={() => void submitAssign()} loading={assignLoading} loadingLabel="Menambahkan...">
            Tambahkan ({selectedStudents.size})
          </LoadingButton>
        </div>
      </div>
    </div>
  </Dialog.Content>
</Dialog.Root>

<!-- Dialog Pindah Kelas -->
<Dialog.Root bind:open={showPindah}>
  <Dialog.Content class="max-w-sm">
    <div class="space-y-4">
      <h2 class="text-base font-semibold">Pindah Kelas</h2>
      <p class="text-sm text-muted-foreground">Pindahkan <strong>{pindahStudentName}</strong> ke rombel:</p>
      <select bind:value={pindahTarget} class="flex h-10 w-full rounded-lg border border-input bg-background px-3 text-sm text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring">
        <option value="">— Pilih Rombel —</option>
        {#each rombelOptions as r}
          <option value={r.id}>{r.code} — {r.level}</option>
        {/each}
      </select>
      <div class="flex justify-end gap-2 pt-2">
        <Button variant="outline" onclick={() => (showPindah = false)}>Batal</Button>
        <LoadingButton onclick={() => void submitPindah()} loading={pindahLoading} loadingLabel="Memindahkan...">
          Pindahkan
        </LoadingButton>
      </div>
    </div>
  </Dialog.Content>
</Dialog.Root>

<!-- Dialog Wali Kelas -->
<Dialog.Root bind:open={showWaliKelas}>
  <Dialog.Content class="max-w-md">
    <div class="space-y-4">
      <h2 class="text-base font-semibold">Pilih Wali Kelas</h2>
      <p class="text-sm text-muted-foreground">Untuk {detail?.code ?? 'rombel'}</p>
      <div><Input placeholder="Cari guru..." bind:value={waliKelasSearch} /></div>
      <div class="max-h-64 overflow-y-auto space-y-1">
        {#if filteredEmployees.length === 0}
          <p class="text-sm text-muted-foreground text-center py-4">Tidak ada pegawai ditemukan.</p>
        {:else}
          {#each filteredEmployees as e (e.id)}
            <label class="flex items-center gap-3 rounded-lg border border-border px-3 py-2 hover:bg-muted/10 cursor-pointer">
              <input type="radio" name="wali" value={e.id} checked={waliSelected === e.id} onchange={() => (waliSelected = e.id)} class="size-4 accent-primary" />
              <div class="min-w-0 flex-1">
                <span class="text-sm font-medium">{e.nama}</span>
                {#if e.nip}
                  <span class="text-xs text-muted-foreground ml-2">NIP: {e.nip}</span>
                {/if}
              </div>
            </label>
          {/each}
        {/if}
      </div>
      <div class="flex justify-end gap-2 pt-2">
        <Button variant="outline" onclick={() => (showWaliKelas = false)}>Batal</Button>
        <LoadingButton onclick={() => void submitWaliKelas()} loading={waliLoading} loadingLabel="Menyimpan...">
          Simpan
        </LoadingButton>
      </div>
    </div>
  </Dialog.Content>
</Dialog.Root>
