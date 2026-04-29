<script lang="ts">
  import * as Card from '$lib/components/ui/card';
  import * as Table from '$lib/components/ui/table';
  import { Input } from '$lib/components/ui/input';
  import { Button } from '$lib/components/ui/button';
  import { Badge } from '$lib/components/ui/badge';

  interface Schedule {
    id: string;
    label: string;
    run_time: string;
    run_type: string;
    is_enabled: boolean;
  }

  let { schedules = $bindable(), onsave }: { schedules: Schedule[]; onsave: () => void } = $props();

  // Only show morning (rekap) schedules in this component
  const rekapSchedules = $derived(schedules.filter(s => s.run_type === 'morning'));

  let newTime  = $state('');
  let newLabel = $state('');
  let adding   = $state(false);
  let saving   = $state(false);
  let toast    = $state('');

  function showToast(msg: string) {
    toast = msg;
    setTimeout(() => (toast = ''), 3000);
  }

  async function addSchedule() {
    if (!newTime) return;
    adding = true;
    try {
      const res = await fetch('/api/schedules', {
        method: 'POST',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify({
          label:      newLabel.trim() || `Rekap ${newTime}`,
          run_time:   newTime,
          run_type:   'morning',
          is_enabled: true,
        }),
      });
      if (!res.ok) {
        const d = await res.json().catch(() => ({})) as { error?: string };
        showToast(d.error ?? 'Gagal menambah jadwal');
        return;
      }
      newTime  = '';
      newLabel = '';
      onsave();
    } catch { showToast('Gagal menambah jadwal'); }
    finally { adding = false; }
  }

  async function deleteSchedule(id: string) {
    try {
      await fetch(`/api/schedules/${id}`, { method: 'DELETE' });
      onsave();
    } catch { showToast('Gagal menghapus jadwal'); }
  }

  async function saveChanges() {
    saving = true;
    try {
      const res = await fetch('/api/schedules', {
        method: 'PUT',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify({ schedules: rekapSchedules }),
      });
      if (res.ok) { showToast('Jadwal disimpan'); onsave(); }
      else showToast('Gagal menyimpan');
    } catch { showToast('Gagal menyimpan'); }
    finally { saving = false; }
  }
</script>

<Card.Root>
  <Card.Header class="pb-3">
    <div class="flex items-center justify-between">
      <div>
        <Card.Title class="text-base">Jadwal Rekap Otomatis</Card.Title>
        <Card.Description>Scrape kehadiran PUSAKA Kemenag — bisa tambah beberapa waktu per hari</Card.Description>
      </div>
      <Button onclick={saveChanges} size="sm" disabled={saving}>
        {saving ? 'Menyimpan...' : 'Simpan'}
      </Button>
    </div>
  </Card.Header>

  {#if toast}
    <div class="mx-6 mb-3 rounded-md border border-green-200 bg-green-50 px-3 py-2 text-sm text-green-800">
      {toast}
    </div>
  {/if}

  <Card.Content class="p-0 overflow-x-auto">
    <Table.Root>
      <Table.Header>
        <Table.Row>
          <Table.Head>Nama Jadwal</Table.Head>
          <Table.Head class="w-28">Jam</Table.Head>
          <Table.Head class="text-center w-20">Aktif</Table.Head>
          <Table.Head class="w-20"></Table.Head>
        </Table.Row>
      </Table.Header>
      <Table.Body>
        {#each rekapSchedules as s (s.id)}
          {@const idx = schedules.findIndex(x => x.id === s.id)}
          <Table.Row>
            <Table.Cell>
              <Input class="h-8 text-sm" bind:value={schedules[idx].label} />
            </Table.Cell>
            <Table.Cell>
              <Input class="w-24 font-mono h-8" bind:value={schedules[idx].run_time} placeholder="07:00" />
            </Table.Cell>
            <Table.Cell class="text-center">
              <input type="checkbox" bind:checked={schedules[idx].is_enabled}
                class="h-4 w-4 rounded border-input accent-green-700" />
            </Table.Cell>
            <Table.Cell>
              <Button size="sm" variant="ghost" class="text-destructive hover:text-destructive h-8"
                onclick={() => deleteSchedule(s.id)}>
                Hapus
              </Button>
            </Table.Cell>
          </Table.Row>
        {:else}
          <Table.Row>
            <Table.Cell colspan={4} class="py-6 text-center text-muted-foreground text-sm">
              Belum ada jadwal rekap.
            </Table.Cell>
          </Table.Row>
        {/each}

        <!-- Add new schedule row -->
        <Table.Row class="bg-slate-50/60">
          <Table.Cell>
            <Input class="h-8 text-sm" bind:value={newLabel} placeholder="Nama jadwal (opsional)" />
          </Table.Cell>
          <Table.Cell>
            <Input class="w-24 font-mono h-8" bind:value={newTime} placeholder="HH:MM" type="time" />
          </Table.Cell>
          <Table.Cell class="text-center">
            <Badge variant="outline" class="text-xs">Aktif</Badge>
          </Table.Cell>
          <Table.Cell>
            <Button size="sm" variant="outline" onclick={addSchedule}
              disabled={adding || !newTime} class="h-8">
              {adding ? '...' : '+ Tambah'}
            </Button>
          </Table.Cell>
        </Table.Row>
      </Table.Body>
    </Table.Root>
  </Card.Content>
</Card.Root>
