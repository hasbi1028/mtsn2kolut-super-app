import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { getAppSettings, updateAppSettings } from '$lib/server/settings';
import { logInfo } from '$lib/server/logger';

export const GET: RequestHandler = () => json(getAppSettings());

export const PUT: RequestHandler = async ({ request }) => {
	const payload = await request.json().catch(() => ({})) as Record<string, unknown>;
	const settings = updateAppSettings({
		max_concurrent: payload.max_concurrent as number | undefined,
		headless:       payload.headless as boolean | undefined,
	});
	logInfo('app settings updated', settings as Record<string, unknown>);
	return json(settings);
};
