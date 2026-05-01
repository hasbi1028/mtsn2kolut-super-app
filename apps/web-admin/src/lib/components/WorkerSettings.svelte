<script lang="ts">
  import * as Card from '$lib/components/ui/card';
  import { Input } from '$lib/components/ui/input';
  import LoadingButton from '$lib/components/LoadingButton.svelte';

  let { settings = $bindable(), onsave }: {
    settings: { max_concurrent: number; headless: boolean };
    onsave: () => void | Promise<void>;
  } = $props();

  let saving = $state(false);

  async function handleSave() {
    saving = true;
    try {
      await onsave();
    } finally {
      saving = false;
    }
  }
</script>

<Card.Root>
  <Card.Header class="pb-3">
    <Card.Title class="text-base">Worker Settings</Card.Title>
    <Card.Description>Konfigurasi Playwright worker untuk job Pusaka</Card.Description>
  </Card.Header>
  <Card.Content>
    <div class="grid grid-cols-1 gap-4 sm:grid-cols-3 items-end">
      <div>
        <label for="max-concurrent" class="mb-1.5 block text-sm font-medium">Max Concurrent</label>
        <Input id="max-concurrent" type="number" min="1" max="20" bind:value={settings.max_concurrent} class="w-full" />
      </div>
      <div class="flex items-center gap-3 pb-1">
        <input id="headless-mode" type="checkbox" bind:checked={settings.headless}
          class="h-4 w-4 rounded border-input accent-green-700" />
        <label for="headless-mode" class="text-sm font-medium cursor-pointer">Headless Mode</label>
      </div>
      <div>
        <LoadingButton onclick={() => void handleSave()} loading={saving} loadingLabel="Menyimpan..." class="w-full sm:w-auto" label="Simpan Setting" />
        <p class="mt-1.5 text-xs text-muted-foreground">Worker sinkron otomatis ~30 detik.</p>
      </div>
    </div>
  </Card.Content>
</Card.Root>
