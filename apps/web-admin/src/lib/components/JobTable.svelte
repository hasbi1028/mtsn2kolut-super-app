<script lang="ts">
  import * as Table from '$lib/components/ui/table';
  import { Badge } from '$lib/components/ui/badge';

  interface Job {
    created_at: string;
    created_at_wita?: string;
    nama: string;
    run_type: string;
    status: string;
    attempts: number;
    max_attempts: number;
    claimed_by?: string;
    error_message?: string;
  }

  let { jobs }: { jobs: Job[] } = $props();

  function statusVariant(s: string): 'default' | 'destructive' | 'outline' | 'secondary' {
    if (s === 'success') return 'default';
    if (s === 'failed')  return 'destructive';
    if (s === 'running') return 'outline';
    return 'secondary';
  }
</script>

<div class="overflow-x-auto rounded-md border">
  <Table.Root>
    <Table.Header>
      <Table.Row>
        <Table.Head>Waktu</Table.Head>
        <Table.Head>Nama</Table.Head>
        <Table.Head>Tipe</Table.Head>
        <Table.Head>Status</Table.Head>
        <Table.Head>Attempt</Table.Head>
        <Table.Head>Worker</Table.Head>
        <Table.Head>Error</Table.Head>
      </Table.Row>
    </Table.Header>
    <Table.Body>
      {#each jobs as j}
        <Table.Row>
          <Table.Cell class="text-muted-foreground text-xs">{j.created_at_wita || j.created_at}</Table.Cell>
          <Table.Cell class="font-medium">{j.nama}</Table.Cell>
          <Table.Cell>{j.run_type}</Table.Cell>
          <Table.Cell>
            <Badge variant={statusVariant(j.status)}>{j.status}</Badge>
          </Table.Cell>
          <Table.Cell class="text-center">{j.attempts}/{j.max_attempts}</Table.Cell>
          <Table.Cell class="text-muted-foreground">{j.claimed_by || '-'}</Table.Cell>
          <Table.Cell class="text-muted-foreground">{j.error_message || '-'}</Table.Cell>
        </Table.Row>
      {/each}
    </Table.Body>
  </Table.Root>
</div>
