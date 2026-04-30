<script lang="ts">
  import * as Card from '$lib/components/ui/card';
  import * as Table from '$lib/components/ui/table';
  import { Badge } from '$lib/components/ui/badge';
  import { Button } from '$lib/components/ui/button';
  import { resolve } from '$app/paths';

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
    onreload: () => void;
  } = $props();

  let confirmId = $state<string | null>(null);
  let busyId = $state<string | null>(null);
  let filterEmploymentType = $state('');

  let filteredEmployees = $derived(
    filterEmploymentType
      ? employees.filter((employee) => employee.employment_type === filterEmploymentType)
      : employees
  );

  function employmentLabel(value: string) {
    return { pns: 'PNS', pppk: 'PPPK', honorer: 'Honorer', lainnya: 'Lainnya' }[value] ?? value;
  }

  async function toggleEmployeeStatus(emp: Employee) {
    const next = !emp.is_active;
    if (!confirm(`${next ? 'Aktifkan' : 'Nonaktifkan'} pegawai ${emp.nama}?`)) return;
    busyId = emp.id;
    try {
      const res = await fetch(`/api/employees/${emp.id}/status`, {
        method: 'PATCH',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify({ is_active: next }),
      });
      if (!res.ok) {
        alert('Gagal memperbarui status pegawai');
        return;
      }
      onreload();
    } finally {
      busyId = null;
    }
  }

  async function doDelete(id: string) {
    busyId = id;
    await fetch('/api/employees', {
      method: 'DELETE',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify({ id }),
    });
    busyId = null;
    confirmId = null;
    onreload();
  }
</script>

<Card.Root>
  <Card.Header class="pb-3">
    <div class="flex items-center justify-between gap-3">
      <div>
        <Card.Title class="text-base">Master Pegawai Sekolah</Card.Title>
        <Card.Description>Menampilkan seluruh pegawai sekolah. Operasional akun, jadwal, dan job PUSAKA dikelola dari menu PUSAKA.</Card.Description>
      </div>
      <div class="flex items-center gap-2">
        <select bind:value={filterEmploymentType} class="rounded-md border border-input bg-background px-3 py-2 text-sm">
          <option value="">Semua status</option>
          <option value="pns">PNS</option>
          <option value="pppk">PPPK</option>
          <option value="honorer">Honorer</option>
          <option value="lainnya">Lainnya</option>
        </select>
        <Badge variant="secondary">{filteredEmployees.length} pegawai</Badge>
      </div>
    </div>
  </Card.Header>
  <Card.Content class="overflow-x-auto p-0">
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
                  <Button size="sm" variant="destructive" onclick={() => doDelete(e.id)} disabled={busyId === e.id}>
                    {busyId === e.id ? '...' : 'Ya, Hapus'}
                  </Button>
                  <Button size="sm" variant="ghost" onclick={() => (confirmId = null)}>Batal</Button>
                </div>
              {:else}
                <div class="flex flex-wrap items-center justify-end gap-2">
                  {#if e.pusaka_eligible}
                    <a href={resolve('/pusaka/employees')}>
                      <Button size="sm" variant="outline">Kelola PUSAKA</Button>
                    </a>
                  {/if}
                  <Button size="sm" variant="outline" onclick={() => toggleEmployeeStatus(e)} disabled={busyId === e.id}>
                    {e.is_active ? 'Nonaktifkan' : 'Aktifkan'}
                  </Button>
                  <Button size="sm" variant="ghost" class="text-destructive hover:text-destructive" onclick={() => (confirmId = e.id)}>
                    Hapus
                  </Button>
                </div>
              {/if}
            </Table.Cell>
          </Table.Row>
        {:else}
          <Table.Row>
            <Table.Cell colspan={5} class="py-12 text-center text-muted-foreground">Belum ada data pegawai.</Table.Cell>
          </Table.Row>
        {/each}
      </Table.Body>
    </Table.Root>
  </Card.Content>
</Card.Root>
