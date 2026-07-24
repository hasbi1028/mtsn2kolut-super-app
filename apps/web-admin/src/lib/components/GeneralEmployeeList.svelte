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
  import { readClientJson } from '$lib/client/api';
  import { displayName } from '$lib/utils/display-name';

  interface Employee {
    id: string;
    pegawai_uid: string;
    nip: string;
    nama: string;
    unit_kerja: string;
    employment_type: string;
    tanggal_lahir: string;
    jenis_kelamin: string;
    tempat_lahir: string;
    pusaka_eligible: boolean;
    has_pusaka_account: boolean;
    pusaka_is_enabled: boolean;
    is_active: boolean;
  }

  let { employees, onreload }: {
    employees: Employee[];
    onreload?: () => void | Promise<void>;
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
    tanggal_lahir: '',
    jenis_kelamin: '',
    tempat_lahir: '',
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


  function formatBirthDate(value: string) {
    if (!value) return '—';
    const [year, month, day] = value.split('-');
    if (!year || !month || !day) return value;
    return `${day}/${month}/${year}`;
  }

  function employmentLabel(value: string) {
    return { pns: 'PNS', pppk: 'PPPK', honorer: 'Honorer', lainnya: 'Lainnya' }[value] ?? value;
  }

  function genderLabel(value: string) {
    return { L: 'Laki-laki', P: 'Perempuan' }[value] ?? '—';
  }

  function employeeName(employee: Employee) {
    return displayName({ nama: employee.nama, name: employee.nip }, 'Pegawai tanpa nama');
  }

  function showError(message: string) {
    toast.error(message);
  }

  function mutationErrorMessage(error: unknown, fallbackMessage: string) {
    if (error instanceof Error && error.message.trim()) return error.message;
    return fallbackMessage;
  }

  function openEditDialog(employee: Employee) {
    editingEmployee = employee;
    editForm = {
      nip: employee.nip,
      nama: employee.nama,
      unit_kerja: employee.unit_kerja,
      employment_type: employee.employment_type,
      tanggal_lahir: employee.tanggal_lahir || '',
      jenis_kelamin: employee.jenis_kelamin || '',
      tempat_lahir: employee.tempat_lahir || '',
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
      await readClientJson<unknown>(res);
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
      await readClientJson<unknown>(res);
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
      await readClientJson<unknown>(res);
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
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground mb-1">Filter Unit Kerja</p>
          <select bind:value={filterUnitKerja} class="h-10 w-full rounded-lg border border-input bg-background px-3 text-sm text-foreground">
            <option value="">Semua unit</option>
            {#each unitKerjaOptions as unit (unit)}
              <option value={unit}>{unit}</option>
            {/each}
          </select>
        </div>
        <div>
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground mb-1">Status Kepegawaian</p>
          <select bind:value={filterEmploymentType} class="h-10 w-full rounded-lg border border-input bg-background px-3 text-sm text-foreground">
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
          <Table.Head class="hidden md:table-cell">Identitas</Table.Head>
          <Table.Head>Status Kepegawaian</Table.Head>
          <Table.Head class="hidden md:table-cell">PUSAKA</Table.Head>
          <Table.Head class="text-right">Aksi</Table.Head>
        </Table.Row>
      </Table.Header>
      <Table.Body>
        {#each filteredEmployees as e (e.id)}
          <Table.Row>
            <Table.Cell>
              <div class="font-medium">{employeeName(e)}</div>
              <div class="text-xs text-muted-foreground">NIP {e.nip || 'belum diisi'}</div>
              {#if e.unit_kerja}
                <div class="text-xs text-muted-foreground">{e.unit_kerja}</div>
              {/if}
              <div class="mt-1">
                {#if e.is_active}
                  <Badge variant="outline" class="text-[11px] border-primary/20 text-primary">Aktif</Badge>
                {:else}
                  <Badge variant="secondary" class="text-[11px]">Nonaktif</Badge>
                {/if}
              </div>
            </Table.Cell>
            <Table.Cell class="hidden md:table-cell text-sm text-muted-foreground">{e.unit_kerja || '—'}</Table.Cell>
            <Table.Cell class="hidden md:table-cell text-sm text-muted-foreground">
              <div>{e.tempat_lahir || '—'}, {formatBirthDate(e.tanggal_lahir)}</div>
              <div class="text-xs">{genderLabel(e.jenis_kelamin)}</div>
            </Table.Cell>
            <Table.Cell>
              <Badge variant="outline">{employmentLabel(e.employment_type)}</Badge>
            </Table.Cell>
            <Table.Cell class="hidden md:table-cell">
              {#if e.pusaka_eligible}
                {#if e.has_pusaka_account}
                  {#if e.pusaka_is_enabled}
                    <Badge variant="outline" class="border-primary/20 text-primary">Terkonfigurasi & aktif</Badge>
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
                <div class="flex flex-wrap items-center justify-end gap-1 md:gap-2">
                  <span class="text-xs text-amber-600">Hapus pegawai ini?</span>
                  <form method="POST" action="?/hapus" class="inline">
                    <input type="hidden" name="id" value={e.id} />
                    <button type="submit" class="inline-flex items-center justify-center rounded-md px-2 md:px-3 py-1 md:py-1.5 text-xs md:text-sm font-medium text-destructive-foreground bg-destructive hover:bg-destructive/90">
                      Ya, Hapus
                    </button>
                  </form>
                  <button class="inline-flex items-center justify-center rounded-md px-2 md:px-3 py-1 md:py-1.5 text-xs md:text-sm font-medium hover:bg-accent" onclick={() => (confirmId = null)}>Batal</button>
                </div>
              {:else}
                <div class="flex flex-wrap items-center justify-end gap-1 md:gap-2">
                  {#if e.pusaka_eligible}
                    <a href={resolve('/pusaka/employees')}>
                      <Button size="sm" variant="outline" class="text-xs md:text-sm px-2 md:px-3">Kelola PUSAKA</Button>
                    </a>
                  {/if}
                  <Button size="sm" variant="outline" onclick={() => openEditDialog(e)} class="text-xs md:text-sm px-2 md:px-3">
                    Edit
                  </Button>
                  <form method="POST" action="?/nonaktifkan" class="inline">
                    <input type="hidden" name="id" value={e.id} />
                    <input type="hidden" name="is_active" value={String(e.is_active)} />
                    <button type="submit" class="inline-flex items-center justify-center rounded-md px-2 md:px-3 py-1 md:py-1.5 text-xs md:text-sm font-medium border border-input bg-background hover:bg-accent hover:text-accent-foreground">
                      {e.is_active ? 'Nonaktifkan' : 'Aktifkan'}
                    </button>
                  </form>
                  <form method="POST" action="?/hapus" class="inline">
                    <input type="hidden" name="id" value={e.id} />
                    <button type="submit" class="inline-flex items-center justify-center rounded-md px-2 md:px-3 py-1 md:py-1.5 text-xs md:text-sm font-medium text-destructive hover:bg-destructive/10">
                      Hapus
                    </button>
                  </form>
                </div>
              {/if}
            </Table.Cell>
          </Table.Row>
        {:else}
          <Table.Row>
            <Table.Cell colspan={6} class="p-4">
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
        <h2 class="text-base font-semibold text-foreground">Edit Pegawai</h2>
        <p class="mt-1 text-sm text-muted-foreground">Perbarui data umum pegawai tanpa masuk ke area operasional PUSAKA.</p>
      </div>

      <div class="grid gap-3 sm:grid-cols-2">
        <div>
          <label for="edit-pegawai-uid" class="mb-1 block text-xs font-medium text-muted-foreground">ID internal pegawai</label>
          <input id="edit-pegawai-uid" class="w-full rounded-md border border-input bg-muted px-3 py-2 font-mono text-sm text-muted-foreground" value={editingEmployee?.pegawai_uid ?? ''} readonly />
        </div>
        <div>
          <label for="edit-nip" class="mb-1 block text-xs font-medium text-muted-foreground">NIP</label>
          <input id="edit-nip" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" placeholder="Opsional untuk honorer" bind:value={editForm.nip} />
        </div>
        <div>
          <label for="edit-nama" class="mb-1 block text-xs font-medium text-muted-foreground">Nama</label>
          <input id="edit-nama" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={editForm.nama} />
        </div>
        <div>
          <label for="edit-unit" class="mb-1 block text-xs font-medium text-muted-foreground">Unit Kerja</label>
          <input id="edit-unit" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={editForm.unit_kerja} />
        </div>
        <div>
          <label for="edit-tempat-lahir" class="mb-1 block text-xs font-medium text-muted-foreground">Tempat Lahir</label>
          <input id="edit-tempat-lahir" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={editForm.tempat_lahir} />
        </div>
        <div>
          <label for="edit-tanggal-lahir" class="mb-1 block text-xs font-medium text-muted-foreground">Tanggal Lahir</label>
          <input id="edit-tanggal-lahir" type="date" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={editForm.tanggal_lahir} />
          <p class="mt-1 text-[11px] text-muted-foreground">Nomor internal pegawai baru memakai tahun lahir saat dibuat.</p>
        </div>
        <div>
          <label for="edit-jenis-kelamin" class="mb-1 block text-xs font-medium text-muted-foreground">Jenis Kelamin</label>
          <select id="edit-jenis-kelamin" class="h-10 w-full rounded-lg border border-input bg-background px-3 text-sm text-foreground" bind:value={editForm.jenis_kelamin}>
            <option value="">Belum diisi</option>
            <option value="L">Laki-laki</option>
            <option value="P">Perempuan</option>
          </select>
        </div>
        <div>
          <label for="edit-type" class="mb-1 block text-xs font-medium text-muted-foreground">Status Kepegawaian</label>
          <select id="edit-type" class="h-10 w-full rounded-lg border border-input bg-background px-3 text-sm text-foreground" bind:value={editForm.employment_type}>
            <option value="pns">PNS</option>
            <option value="pppk">PPPK</option>
            <option value="honorer">Honorer</option>
            <option value="lainnya">Lainnya</option>
          </select>
        </div>
      </div>

      <label class="flex items-center gap-2 rounded-md border border-border px-3 py-2 text-sm text-foreground">
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
