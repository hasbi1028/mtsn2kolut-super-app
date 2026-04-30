<script lang="ts">
  import * as Card from '$lib/components/ui/card';
  import * as Table from '$lib/components/ui/table';
  import { Input } from '$lib/components/ui/input';
  import { Button } from '$lib/components/ui/button';
  import { Badge } from '$lib/components/ui/badge';
  import { toast } from '$lib/components/ui/sonner';
  import LoadingButton from '$lib/components/LoadingButton.svelte';
  import EmptyStatePanel from '$lib/components/EmptyStatePanel.svelte';
  import SuccessPanel from '$lib/components/SuccessPanel.svelte';

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
  let success  = $state('');

  function showToast(msg: string) {
    toast.success(msg);
  }

  function showError(msg: string) {
    toast.error(msg);
  }

  async function addSchedule() {
    if (!newTime) return;
    adding = true;
    success = '';
    try {
      const res = await fetch('/api/pusaka/schedules', {
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
        showError(d.error ?? 'Gagal menambah jadwal');
        return;
      }
      newTime  = '';
      newLabel = '';
      success = 'Jadwal rekap baru berhasil ditambahkan. Jangan lupa simpan perubahan utama bila masih ada penyesuaian label atau status.';
      onsave();
    } catch { showError('Gagal menambah jadwal'); }
    finally { adding = false; }
  }

  async function deleteSchedule(id: string) {
    try {
      success = '';
      await fetch(`/api/pusaka/schedules/${id}`, { method: 'DELETE' });
      success = 'Jadwal rekap berhasil dihapus.';
      onsave();
    } catch { showError('Gagal menghapus jadwal'); }
  }

  async function saveChanges() {
    saving = true;
    success = '';
    try {
      const res = await fetch('/api/pusaka/schedules', {
        method: 'PUT',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify({ schedules: rekapSchedules }),
      });
      if (res.ok) { showToast('Jadwal disimpan'); success = 'Perubahan jadwal rekap otomatis berhasil disimpan.'; onsave(); }
      else showError('Gagal menyimpan');
    } catch { showError('Gagal menyimpan'); }
    finally { saving = false; }
  }
</script>

<Card.Root>
  <Card.Header class="pb-3">
    <div class="flex flex-col gap-4">
      <div>
        <Card.Title class="text-base">Jadwal Rekap Otomatis</Card.Title>
        <Card.Description>Scrape kehadiran PUSAKA Kemenag — bisa tambah beberapa waktu per hari</Card.Description>
      </div>
      <div class="grid gap-3 md:grid-cols-[1.2fr_0.8fr_auto]">
        <div>
          <p class="mb-1 text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">Nama Jadwal Baru</p>
          <Input class="h-10 text-sm" bind:value={newLabel} placeholder="Nama jadwal (opsional)" />
        </div>
        <div>
          <p class="mb-1 text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">Jam Rekap</p>
          <Input class="h-10 font-mono text-sm" bind:value={newTime} placeholder="HH:MM" type="time" />
        </div>
        <div class="flex items-end gap-2">
          <LoadingButton size="sm" variant="outline" onclick={addSchedule}
            loading={adding}
            loadingLabel="Menambah..."
            disabled={adding || !newTime} class="h-10">
            + Tambah
          </LoadingButton>
          <LoadingButton onclick={saveChanges} size="sm" loading={saving} loadingLabel="Menyimpan..." disabled={saving} class="h-10">
            Simpan
          </LoadingButton>
        </div>
      </div>
    </div>
  </Card.Header>

  <Card.Content class="space-y-4 p-0">
    {#if success}
      <div class="px-6 pt-1">
        <SuccessPanel title="Jadwal Rekap Diperbarui" message={success} compact />
      </div>
    {/if}
    <div class="overflow-x-auto">
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
            <Table.Cell colspan={4} class="p-4">
              <EmptyStatePanel
                compact
                title="Belum ada jadwal rekap"
                description="Tambahkan jam rekap pertama agar sinkronisasi kehadiran otomatis mulai berjalan dari panel ini."
              />
            </Table.Cell>
          </Table.Row>
        {/each}
      </Table.Body>
    </Table.Root>
    </div>
  </Card.Content>
</Card.Root>
