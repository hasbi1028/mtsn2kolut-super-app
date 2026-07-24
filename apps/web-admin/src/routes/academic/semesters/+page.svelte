<script lang="ts">
  import { Badge } from '$lib/components/ui/badge';
  import { Button } from '$lib/components/ui/button';
  import * as Card from '$lib/components/ui/card';
  import * as Dialog from '$lib/components/ui/dialog';
  import { Input } from '$lib/components/ui/input';
  import { toast } from '$lib/components/ui/sonner';
  import LoadingButton from '$lib/components/LoadingButton.svelte';

  let { data } = $props();

  type Semester = {
    id: string;
    academic_year_id: string;
    academic_year_name: string;
    name: string;
    label: string;
    start_date: string;
    end_date: string;
    is_active: boolean;
  };

  let semesters = $state<Semester[]>(data.items ?? []);
  let loadingId = $state<string | null>(null);
  let showTambah = $state(false);
  let tambahForm = $state({ label: '', name: 'Ganjil', start_date: '', end_date: '', is_active: false });
  let tambahLoading = $state(false);

  async function refresh() {
    const res = await fetch('/api/academic/semesters');
    if (res.ok) {
      const payload = await res.json();
      semesters = payload.items ?? [];
    }
  }

  async function activateSemester(id: string) {
    loadingId = id;
    try {
      const res = await fetch(`/api/academic/semesters/${id}/activate`, { method: 'POST' });
      if (res.ok) {
        toast.success('Semester berhasil diaktifkan');
        await refresh();
      } else {
        toast.error('Gagal mengaktifkan semester');
      }
    } catch {
      toast.error('Gagal mengaktifkan semester');
    } finally {
      loadingId = null;
    }
  }

  async function deleteSemester(id: string, label: string) {
    if (!confirm(`Hapus "${label}"?`)) return;
    loadingId = id;
    try {
      const res = await fetch(`/api/academic/semesters/${id}`, { method: 'DELETE' });
      if (res.ok || res.status === 204) {
        toast.success(`Semester "${label}" dihapus`);
        await refresh();
      } else {
        toast.error('Gagal menghapus semester');
      }
    } catch {
      toast.error('Gagal menghapus semester');
    } finally {
      loadingId = null;
    }
  }

  async function submitTambah() {
    if (!tambahForm.label || !tambahForm.start_date || !tambahForm.end_date) {
      toast.error('Label, tanggal mulai, dan tanggal selesai wajib diisi');
      return;
    }
    tambahLoading = true;
    try {
      const res = await fetch('/api/academic/semesters', {
        method: 'POST',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify({
          academic_year_id: '',
          name: tambahForm.name,
          label: tambahForm.label,
          start_date: tambahForm.start_date,
          end_date: tambahForm.end_date,
          is_active: false,
        }),
      });
      if (res.ok) {
        toast.success(`Semester "${tambahForm.label}" berhasil dibuat`);
        showTambah = false;
        tambahForm = { label: '', name: 'Ganjil', start_date: '', end_date: '', is_active: false };
        await refresh();
      } else {
        const body = await res.json().catch(() => ({}));
        toast.error(body?.error || 'Gagal membuat semester');
      }
    } catch {
      toast.error('Gagal membuat semester');
    } finally {
      tambahLoading = false;
    }
  }

  function formatDate(dateStr: string) {
    if (!dateStr) return '—';
    const [y, m, d] = dateStr.split('-');
    return `${d}/${m}/${y}`;
  }
</script>

<svelte:head><title>Semester — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6">
  <div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
    <div>
      <h1 class="text-2xl font-black text-foreground">Semester</h1>
      <p class="mt-1 text-sm text-muted-foreground">Kelola tahun pelajaran dan semester aktif madrasah.</p>
    </div>
    <Button onclick={() => (showTambah = true)}>+ Tambah Semester</Button>
  </div>

  {#if semesters.length === 0}
    <Card.Root>
      <Card.Content class="p-8 text-center">
        <p class="text-sm text-muted-foreground">Belum ada semester. Tambah semester baru untuk memulai.</p>
      </Card.Content>
    </Card.Root>
  {:else}
    <div class="space-y-4">
      {#each semesters as sem (sem.id)}
        <Card.Root class={sem.is_active ? 'border-primary/30' : ''}>
          <Card.Content class="p-5">
            <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
              <div class="min-w-0 flex-1">
                <div class="flex items-center gap-2">
                  <h3 class="text-base font-semibold text-foreground">{sem.label}</h3>
                  {#if sem.is_active}
                    <Badge variant="default" class="bg-primary text-primary-foreground text-[10px]">Aktif</Badge>
                  {/if}
                </div>
                <p class="mt-1 text-xs text-muted-foreground">
                  {formatDate(sem.start_date)} — {formatDate(sem.end_date)}
                </p>
              </div>
              <div class="flex flex-wrap items-center gap-2">
                {#if !sem.is_active}
                  <Button
                    size="sm"
                    variant="default"
                    onclick={() => void activateSemester(sem.id)}
                    disabled={loadingId === sem.id}
                  >
                    Aktifkan
                  </Button>
                {/if}
                <Button
                  size="sm"
                  variant="outline"
                  class="text-destructive"
                  onclick={() => void deleteSemester(sem.id, sem.label)}
                  disabled={loadingId === sem.id}
                >
                  Hapus
                </Button>
              </div>
            </div>
          </Card.Content>
        </Card.Root>
      {/each}
    </div>
  {/if}
</div>

<Dialog.Root bind:open={showTambah}>
  <Dialog.Content>
    <div class="space-y-4">
      <div>
        <h2 class="text-base font-semibold text-foreground">Tambah Semester</h2>
        <p class="mt-1 text-sm text-muted-foreground">Buat semester baru untuk tahun pelajaran.</p>
      </div>
      <div class="grid gap-4 sm:grid-cols-2">
        <div class="space-y-1.5 sm:col-span-2">
          <label class="block text-xs font-medium text-muted-foreground">Label Semester</label>
          <Input bind:value={tambahForm.label} placeholder="Contoh: Semester Ganjil 2026/2027" />
        </div>
        <div class="space-y-1.5">
          <label class="block text-xs font-medium text-muted-foreground">Nama Semester</label>
          <select bind:value={tambahForm.name} class="flex h-10 w-full rounded-lg border border-input bg-background px-3 text-sm text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring">
            <option value="Ganjil">Ganjil</option>
            <option value="Genap">Genap</option>
          </select>
        </div>
        <div class="space-y-1.5">
          <label class="block text-xs font-medium text-muted-foreground">Status</label>
          <label class="flex h-10 items-center gap-2 rounded-lg border border-input px-3 text-sm">
            <input type="checkbox" bind:checked={tambahForm.is_active} />
            Aktifkan sekarang
          </label>
        </div>
        <div class="space-y-1.5">
          <label class="block text-xs font-medium text-muted-foreground">Tanggal Mulai</label>
          <Input type="date" bind:value={tambahForm.start_date} />
        </div>
        <div class="space-y-1.5">
          <label class="block text-xs font-medium text-muted-foreground">Tanggal Selesai</label>
          <Input type="date" bind:value={tambahForm.end_date} />
        </div>
      </div>
      <div class="flex justify-end gap-2">
        <Button variant="outline" onclick={() => (showTambah = false)}>Batal</Button>
        <LoadingButton onclick={() => void submitTambah()} loading={tambahLoading} loadingLabel="Menyimpan...">Simpan</LoadingButton>
      </div>
    </div>
  </Dialog.Content>
</Dialog.Root>
