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

  const typeLabels: Record<string, string> = {
    morning: 'Rekap', afternoon: 'Rekap', checkin: 'Masuk', checkout: 'Pulang',
  };
</script>

<Card.Root>
  <Card.Header class="pb-3">
    <div class="flex items-center justify-between">
      <div>
        <Card.Title class="text-base">Jadwal Otomatis</Card.Title>
        <Card.Description>Jadwal pengambilan absensi terjadwal harian</Card.Description>
      </div>
      <Button onclick={onsave} size="sm">Simpan Jadwal</Button>
    </div>
  </Card.Header>
  <Card.Content class="p-0 overflow-x-auto">
    <Table.Root>
      <Table.Header>
        <Table.Row>
          <Table.Head>Nama Jadwal</Table.Head>
          <Table.Head class="w-28">Jam</Table.Head>
          <Table.Head class="w-36">Tipe</Table.Head>
          <Table.Head class="text-center w-20">Aktif</Table.Head>
        </Table.Row>
      </Table.Header>
      <Table.Body>
        {#each schedules as s, i (s.id)}
          <Table.Row>
            <Table.Cell class="font-medium">{s.label}</Table.Cell>
            <Table.Cell>
              <Input class="w-24 font-mono" bind:value={schedules[i].run_time} placeholder="07:00" />
            </Table.Cell>
            <Table.Cell>
              <select
                bind:value={schedules[i].run_type}
                class="h-9 w-full rounded-md border border-input bg-background px-3 text-sm"
              >
                {#each Object.entries(typeLabels) as [val, lbl]}
                  <option value={val}>{lbl}</option>
                {/each}
              </select>
            </Table.Cell>
            <Table.Cell class="text-center">
              <input type="checkbox" bind:checked={schedules[i].is_enabled}
                class="h-4 w-4 rounded border-input accent-green-700" />
            </Table.Cell>
          </Table.Row>
        {:else}
          <Table.Row>
            <Table.Cell colspan={4} class="py-8 text-center text-muted-foreground">
              Belum ada jadwal.
            </Table.Cell>
          </Table.Row>
        {/each}
      </Table.Body>
    </Table.Root>
  </Card.Content>
</Card.Root>
