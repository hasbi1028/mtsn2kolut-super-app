<script lang="ts">
  import * as Card from '$lib/components/ui/card';
  import { Input } from '$lib/components/ui/input';
  import LoadingButton from '$lib/components/LoadingButton.svelte';

  type WorkerSettingsState = {
    max_concurrent: number;
    headless: boolean;
    pusaka_geo_base_lat: number;
    pusaka_geo_base_lng: number;
    pusaka_geo_default_radius_m: number;
    pusaka_geo_checkin_radius_m: number;
    pusaka_geo_checkout_radius_m: number;
  };

  let { settings = $bindable(), onsave }: {
    settings: WorkerSettingsState;
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
    <div class="mt-5 border-t border-border pt-5">
      <h3 class="text-sm font-semibold text-foreground">Geolocation PUSAKA</h3>
      <div class="mt-3 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-5">
        <div>
          <label for="pusaka-geo-base-lat" class="mb-1.5 block text-sm font-medium">Latitude Pusat</label>
          <Input id="pusaka-geo-base-lat" type="number" min="-90" max="90" step="0.0000001" bind:value={settings.pusaka_geo_base_lat} class="w-full" />
        </div>
        <div>
          <label for="pusaka-geo-base-lng" class="mb-1.5 block text-sm font-medium">Longitude Pusat</label>
          <Input id="pusaka-geo-base-lng" type="number" min="-180" max="180" step="0.0000001" bind:value={settings.pusaka_geo_base_lng} class="w-full" />
        </div>
        <div>
          <label for="pusaka-geo-default-radius" class="mb-1.5 block text-sm font-medium">Radius Default</label>
          <Input id="pusaka-geo-default-radius" type="number" min="1" max="200" bind:value={settings.pusaka_geo_default_radius_m} class="w-full" />
        </div>
        <div>
          <label for="pusaka-geo-checkin-radius" class="mb-1.5 block text-sm font-medium">Radius Masuk</label>
          <Input id="pusaka-geo-checkin-radius" type="number" min="1" max="200" bind:value={settings.pusaka_geo_checkin_radius_m} class="w-full" />
        </div>
        <div>
          <label for="pusaka-geo-checkout-radius" class="mb-1.5 block text-sm font-medium">Radius Pulang</label>
          <Input id="pusaka-geo-checkout-radius" type="number" min="1" max="200" bind:value={settings.pusaka_geo_checkout_radius_m} class="w-full" />
        </div>
      </div>
    </div>
  </Card.Content>
</Card.Root>
