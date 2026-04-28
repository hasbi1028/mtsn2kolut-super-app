<script lang="ts">
  import * as Table from '$lib/components/ui/table';

  interface AttendanceRecord {
    tanggal: string;
    nama: string;
    nip: string;
    jam_masuk: string;
    jam_pulang: string;
  }

  let { records }: { records: AttendanceRecord[] } = $props();
</script>

<div class="overflow-x-auto rounded-md border">
  <Table.Root>
    <Table.Header>
      <Table.Row>
        <Table.Head>Tanggal</Table.Head>
        <Table.Head>Pegawai</Table.Head>
        <Table.Head>Jam Masuk</Table.Head>
        <Table.Head>Jam Pulang</Table.Head>
      </Table.Row>
    </Table.Header>
    <Table.Body>
      {#each records as a}
        <Table.Row>
          <Table.Cell>{a.tanggal}</Table.Cell>
          <Table.Cell>
            {a.nama}
            <span class="text-muted-foreground">({a.nip})</span>
          </Table.Cell>
          <Table.Cell>{a.jam_masuk || '-'}</Table.Cell>
          <Table.Cell>{a.jam_pulang || '-'}</Table.Cell>
        </Table.Row>
      {/each}
    </Table.Body>
  </Table.Root>
</div>
