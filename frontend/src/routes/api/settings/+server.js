import { json } from '@sveltejs/kit';
import { getAppSettings, updateAppSettings } from '$lib/server/settings';
import { logInfo } from '$lib/server/logger';

export function GET() {
  return json(getAppSettings());
}

export async function PUT({ request }) {
  const payload = await request.json().catch(() => ({}));
  const settings = updateAppSettings({
    max_concurrent: payload.max_concurrent,
    headless: payload.headless
  });

  logInfo('app settings updated', settings);
  return json(settings);
}
