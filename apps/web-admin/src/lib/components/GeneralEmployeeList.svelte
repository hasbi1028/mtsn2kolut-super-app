<script lang="ts">
  import * as Card from '$lib/components/ui/card';
  import * as Table from '$lib/components/ui/table';
  import { Badge } from '$lib/components/ui/badge';
  import { Button } from '$lib/components/ui/button';
  import { resolve } from '$app/paths';
  import * as Dialog from '$lib/components/ui/dialog';
  import { toast } from '$lib/components/ui/sonner';
  import LoadingButton from '$lib/components/LoadingButton.svelte';
  import EmptyStatePanel from '$lib/components/EmptyStatePanel.svelte';
  import SuccessPanel from '$lib/components/SuccessPanel.svelte';
  import { confirmAction } from '$lib/confirm-dialog';

  interface Employee {
    id: string;
    nip: string;
    nama: string;
    unit_kerja: string;
    employment_type: string;
    pusaka_eligible: boolean;
    has_pusaka_account: boolean;
    pusaka_is_enabled: boolean;
    is_active: boolean;
  }

  let { employees, onreload }: {
    employees: Employee[];
    onreload: () => void | Promise<void>;
  } = $props();

  let confirmId = $state<string | null>(null);
  let busyId = $state<string | null>(null);
  let filterEmploymentType = $state('');
  let filterUnitKerja = $state('');
  let showEditDialog = $state(false);
  let editBusy = $state(false);
  let success = $state('');
  let editingEmployee = $state<Employee | null>(null);
  let editForm = $state({
    nip: '',
    nama: '',
    unit_kerja: '',
    employment_type: 'lainnya',
    is_active: true,
  });

  let unitKerjaOptions = $derived(
    [...new Set(
      employees
        .map((employee) => employee.unit_kerja.trim())
        .filter((unit) => unit.length > 0)
    )].sort((a, b) => a.localeCompare(b, 'id-ID'))
  );

  let filteredEmployees = $derived.by(() => {
    let scoped = filterEmploymentType
      ? employees.filter((employee) => employee.employment_type === filterEmploymentType)
      : employees;
    if (filterUnitKerja) {
      scoped = scoped.filter((employee) => employee.unit_kerja === filterUnitKerja);
    }
    return scoped;
  });

  function employmentLabel(value: string) {
    return { pns: 'PNS', pppk: 'PPPK', honorer: 'Honorer', lainnya: 'Lainnya' }[value] ?? value;
  }

  function apiErrorMessage(payload: unknown) {
    if (typeof payload !== 'object' || payload === null) return '';
    if ('error' in payload) {
      const error = payload.error;
      if (typeof error === 'string' && error.trim()) return error;
    }
    if ('message' in payload) {
      const message = payload.message;
      if (typeof message === 'string' && message.trim()) return message;
    }
    return '';
  }

  function showError(message: string) {
    toast.error(message);
  }

  function mutationErrorMessage(error: unknown, fallbackMessage: string) {
    if (error instanceof Error && error.message.trim()) return error.message;
    return fallbackMessage;
  }

  function throwApiError(payload: unknown, fallbackMessage: string): never {
    throw new Error(apiErrorMessage(payload) || fallbackMessage);
  }

  function openEditDialog(employee: Employee) {
    editingEmployee = employee;
    editForm = {
      nip: employee.nip,
      nama: employee.nama,
      unit_kerja: employee.unit_kerja,
      employment_type: employee.employment_type,
      is_active: employee.is_active,
    };
    showEditDialog = true;
  }

  async function saveEdit() {
    if (!editingEmployee) return;
    editBusy = true;
    success = '';
    try {
      const res = await fetch(`/api/employees/${editingEmployee.id}`, {
        method: 'PUT',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify(editForm),
      });
      const data = await res.json().catch(() => null);
      if (!res.ok) {
        throwApiError(data, 'Gagal memperbarui pegawai');
      }
      showEditDialog = false;
      success = `Data pegawai ${editForm.nama} berhasil diperbarui.`;
      await onreload();
    } catch (error) {
      showError(mutationErrorMessage(error, 'Gagal memperbarui pegawai'));
    } finally {
      editBusy = false;
    }
  }

  async function toggleEmployeeStatus(emp: Employee) {
    const next = !emp.is_active;
    if (!(await confirmAction({
      title: next ? 'Aktifkan Pegawai' : 'Nonaktifkan Pegawai',
      message: `${next ? 'Aktifkan' : 'Nonaktifkan'} pegawai ${emp.nama}?`,
      confirmLabel: next ? 'Aktifkan' : 'Nonaktifkan',
      tone: 'warning'
    }))) return;
    busyId = emp.id;
    success = '';
    try {
      const res = await fetch(`/api/employees/${emp.id}/status`, {
        method: 'PATCH',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify({ is_active: next }),
      });
      const data = await res.json().catch(() => null);
      if (!res.ok) {
        throwApiError(data, 'Gagal memperbarui status pegawai');
      }
      success = `Status pegawai ${emp.nama} berhasil diubah menjadi ${next ? 'aktif' : 'nonaktif'}.`;
      await onreload();
    } catch (error) {
      showError(mutationErrorMessage(error, 'Gagal memperbarui status pegawai'));
    } finally {
      busyId = null;
    }
  }

  async function doDelete(id: string) {
    busyId = id;
    success = '';
    try {
      const res = await fetch('/api/employees', {
        method: 'DELETE',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify({ id }),
      });
      const data = await res.json().catch(() => null);
      if (!res.ok) {
        throwApiError(data, 'Gagal menghapus pegawai');
      }
      confirmId = null;
      success = 'Data pegawai berhasil dihapus dari master.';
      await onreload();
    } catch (error) {
      showError(mutationErrorMessage(error, 'Gagal menghapus pegawai'));
    } finally {
      busyId = null;
    }
  }
</script>

<Card.Root>
  <Card.Header class="pb-3">
    <div class="flex flex-col gap-4">
      <div>
        <Card.Title class="text-base">Master Pegawai Sekolah</Card.Title>
        <Card.Description>Menampilkan seluruh pegawai sekolah. Operasional akun, jadwal, dan job PUSAKA dikelola dari menu PUSAKA.</Card.Description>
      </div>
      <div class="grid gap-3 md:grid-cols-[1fr_1fr_auto]">
        <div>
          <p class="mb-1 text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">Filter Unit Kerja</p>
          <select bind:value={filterUnitKerja} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
            <option value="">Semua unit</option>
            {#each unitKerjaOptions as unit (unit)}
              <option value={unit}>{unit}</option>
            {/each}
          </select>
        </div>
        <div>
          <p class="mb-1 text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">Status Kepegawaian</p>
          <select bind:value={filterEmploymentType} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
            <option value="">Semua status</option>
            <option value="pns">PNS</option>
            <option value="pppk">PPPK</option>
            <option value="honorer">Honorer</option>
            <option value="lainnya">Lainnya</option>
          </select>
        </div>
        <div class="flex items-end">
          <Badge variant="secondary" class="h-10 px-3">{filteredEmployees.length} pegawai</Badge>
        </div>
      </div>
    </div>
  </Card.Header>
  <Card.Content class="space-y-4 p-0">
    {#if success}
      <div class="px-6 pt-1">
        <SuccessPanel title="Master Pegawai Diperbarui" message={success} compact />
      </div>
    {/if}
    <div class="overflow-x-auto">
    <Table.Root>
      <Table.Header>
        <Table.Row>
          <Table.Head>Pegawai</Table.Head>
          <Table.Head class="hidden md:table-cell">Unit Kerja</Table.Head>
          <Table.Head>Status Kepegawaian</Table.Head>
          <Table.Head>PUSAKA</Table.Head>
          <Table.Head class="text-right">Aksi</Table.Head>
        </Table.Row>
      </Table.Header>
      <Table.Body>
        {#each filteredEmployees as e (e.id)}
          <Table.Row>
            <Table.Cell>
              <div class="font-medium">{e.nama}</div>
              <div class="font-mono text-xs text-muted-foreground">{e.nip}</div>
              <div class="mt-1">
                {#if e.is_active}
                  <Badge variant="outline" class="text-[11px] border-emerald-300 text-emerald-700">Aktif</Badge>
                {:else}
                  <Badge variant="secondary" class="text-[11px]">Nonaktif</Badge>
                {/if}
              </div>
            </Table.Cell>
            <Table.Cell class="hidden md:table-cell text-sm text-muted-foreground">{e.unit_kerja || '—'}</Table.Cell>
            <Table.Cell>
              <Badge variant="outline">{employmentLabel(e.employment_type)}</Badge>
            </Table.Cell>
            <Table.Cell>
              {#if e.pusaka_eligible}
                {#if e.has_pusaka_account}
                  {#if e.pusaka_is_enabled}
                    <Badge variant="outline" class="border-emerald-300 text-emerald-700">Terkonfigurasi & aktif</Badge>
                  {:else}
                    <Badge variant="secondary">Terkonfigurasi, dinonaktifkan</Badge>
                  {/if}
                {:else}
                  <Badge variant="secondary">Eligible, belum setup</Badge>
                {/if}
              {:else}
                <Badge variant="secondary">Tidak eligible</Badge>
              {/if}
            </Table.Cell>
            <Table.Cell class="text-right">
              {#if confirmId === e.id}
                <div class="flex flex-wrap items-center justify-end gap-2">
                  <span class="text-xs text-amber-700">Hapus pegawai ini?</span>
                  <LoadingButton size="sm" variant="destructive" onclick={() => doDelete(e.id)} loading={busyId === e.id} loadingLabel="Menghapus..." disabled={busyId === e.id}>
                    Ya, Hapus
                  </LoadingButton>
                  <Button size="sm" variant="ghost" onclick={() => (confirmId = null)}>Batal</Button>
                </div>
              {:else}
                <div class="flex flex-wrap items-center justify-end gap-2">
                  {#if e.pusaka_eligible}
                    <a href={resolve('/pusaka/employees')}>
                      <Button size="sm" variant="outline">Kelola PUSAKA</Button>
                    </a>
                  {/if}
                  <Button size="sm" variant="outline" onclick={() => openEditDialog(e)}>
                    Edit
                  </Button>
                  <LoadingButton size="sm" variant="outline" onclick={() => toggleEmployeeStatus(e)} loading={busyId === e.id} loadingLabel="Memproses..." disabled={busyId === e.id}>
                    {e.is_active ? 'Nonaktifkan' : 'Aktifkan'}
                  </LoadingButton>
                  <Button size="sm" variant="ghost" class="text-destructive hover:text-destructive" onclick={() => (confirmId = e.id)}>
                    Hapus
                  </Button>
                </div>
              {/if}
            </Table.Cell>
          </Table.Row>
        {:else}
          <Table.Row>
            <Table.Cell colspan={5} class="p-4">
              <EmptyStatePanel
                compact
                title="Belum ada data pegawai"
                description="Tambahkan pegawai pertama dari form di atas agar master pegawai dan alur operasional sekolah mulai terbangun."
              />
            </Table.Cell>
          </Table.Row>
        {/each}
      </Table.Body>
    </Table.Root>
    </div>
  </Card.Content>
</Card.Root>

<Dialog.Root bind:open={showEditDialog}>
  <Dialog.Content>
    <div class="space-y-4">
      <div>
        <h2 class="text-base font-semibold text-slate-900">Edit Pegawai</h2>
        <p class="mt-1 text-sm text-slate-500">Perbarui data umum pegawai tanpa masuk ke area operasional PUSAKA.</p>
      </div>

      <div class="grid gap-3 sm:grid-cols-2">
        <div>
          <label for="edit-nip" class="mb-1 block text-xs font-medium text-slate-600">NIP</label>
          <input id="edit-nip" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={editForm.nip} />
        </div>
        <div>
          <label for="edit-nama" class="mb-1 block text-xs font-medium text-slate-600">Nama</label>
          <input id="edit-nama" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={editForm.nama} />
        </div>
        <div>
          <label for="edit-unit" class="mb-1 block text-xs font-medium text-slate-600">Unit Kerja</label>
          <input id="edit-unit" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={editForm.unit_kerja} />
        </div>
        <div>
          <label for="edit-type" class="mb-1 block text-xs font-medium text-slate-600">Status Kepegawaian</label>
          <select id="edit-type" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={editForm.employment_type}>
            <option value="pns">PNS</option>
            <option value="pppk">PPPK</option>
            <option value="honorer">Honorer</option>
            <option value="lainnya">Lainnya</option>
          </select>
        </div>
      </div>

      <label class="flex items-center gap-2 rounded-md border border-slate-200 px-3 py-2 text-sm text-slate-700">
        <input type="checkbox" bind:checked={editForm.is_active} />
        Pegawai aktif
      </label>

      <div class="flex justify-end gap-2">
        <Button variant="outline" onclick={() => (showEditDialog = false)}>Batal</Button>
        <LoadingButton onclick={() => void saveEdit()} loading={editBusy} loadingLabel="Menyimpan..." disabled={editBusy}>Simpan Perubahan</LoadingButton>
      </div>
    </div>
  </Dialog.Content>
</Dialog.Root>
