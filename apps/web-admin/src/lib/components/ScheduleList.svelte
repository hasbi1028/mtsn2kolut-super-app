<script lang="ts">
  import * as Card from '$lib/components/ui/card';
  import * as Table from '$lib/components/ui/table';
  import { Input } from '$lib/components/ui/input';
  import { Badge } from '$lib/components/ui/badge';
  import { toast } from '$lib/components/ui/sonner';
  import LoadingButton from '$lib/components/LoadingButton.svelte';
  import EmptyStatePanel from '$lib/components/EmptyStatePanel.svelte';
  import SuccessPanel from '$lib/components/SuccessPanel.svelte';
  import { readClientJson } from '$lib/client/api';

  interface Schedule {
    id: string;
    label: string;
    run_time: string;
    run_type: string;
    is_enabled: boolean;
  }

  let { schedules = $bindable(), onsave }: { schedules: Schedule[]; onsave: () => void | Promise<void> } = $props();

  // Only show morning (rekap) schedules in this component
  const rekapSchedules = $derived(schedules.filter(s => s.run_type === 'morning'));

  let newTime  = $state('');
  let newLabel = $state('');
  let adding   = $state(false);
  let saving   = $state(false);
  let deleteBusyId = $state<string | null>(null);
  let success  = $state('');

  function showToast(msg: string) {
    toast.success(msg);
  }

  function showError(msg: string) {
    toast.error(msg);
  }

  function mutationErrorMessage(error: unknown, fallbackMessage: string) {
    if (error instanceof Error && error.message.trim()) return error.message;
    return fallbackMessage;
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
      await readClientJson<unknown>(res);
      newTime  = '';
      newLabel = '';
      success = 'Jadwal rekap baru berhasil ditambahkan. Jangan lupa simpan perubahan utama bila masih ada penyesuaian label atau status.';
      await onsave();
    } catch (error) { showError(mutationErrorMessage(error, 'Gagal menambah jadwal')); }
    finally { adding = false; }
  }

  async function deleteSchedule(id: string) {
    deleteBusyId = id;
    try {
      success = '';
      const res = await fetch(`/api/pusaka/schedules/${id}`, { method: 'DELETE' });
      await readClientJson<unknown>(res);
      success = 'Jadwal rekap berhasil dihapus.';
      await onsave();
    } catch (error) { showError(mutationErrorMessage(error, 'Gagal menghapus jadwal')); }
    finally { deleteBusyId = null; }
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
      await readClientJson<unknown>(res);
      showToast('Jadwal disimpan');
      success = 'Perubahan jadwal rekap otomatis berhasil disimpan.';
      await onsave();
    } catch (error) { showError(mutationErrorMessage(error, 'Gagal menyimpan')); }
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
          <p class="mb-1 text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">Nama Jadwal Baru</p>
          <Input class="h-10 text-sm" bind:value={newLabel} placeholder="Nama jadwal (opsional)" />
        </div>
        <div>
          <p class="mb-1 text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">Jam Rekap</p>
          <Input class="h-10 font-mono text-sm" bind:value={newTime} placeholder="HH:MM" type="time" />
        </div>
        <div class="flex items-end gap-2">
          <LoadingButton size="sm" variant="outline" onclick={addSchedule}
            loading={adding}
            loadingLabel="Menambah..."
            disabled={adding || !newTime} class="h-10">
            + Tambah
          </LoadingButton>
          <LoadingButton onclick={() => void saveChanges()} size="sm" loading={saving} loadingLabel="Menyimpan..." disabled={saving} class="h-10">
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
              <LoadingButton
                size="sm"
                variant="ghost"
                class="text-destructive hover:text-destructive h-8"
                onclick={() => deleteSchedule(s.id)}
                loading={deleteBusyId === s.id}
                loadingLabel="Menghapus..."
                disabled={deleteBusyId !== null && deleteBusyId !== s.id}
              >
                Hapus
              </LoadingButton>
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
