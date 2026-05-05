import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, proxy, readOptionalRequestJson } from '$lib/server/api';

interface GoSetting { key: string; value: string }

const BLOCKED = new Set(['admin_password', 'admin_password_hash', 'admin_username']);
const NUMBER_KEYS = new Set([
	'max_concurrent',
	'pusaka_geo_base_lat',
	'pusaka_geo_base_lng',
	'pusaka_geo_default_radius_m',
	'pusaka_geo_checkin_radius_m',
	'pusaka_geo_checkout_radius_m'
]);

export const GET = async (event: RequestEvent) => {
	try {
		const rows = await proxy(event).get<GoSetting[]>('/api/pusaka/settings');
		const flat: Record<string, unknown> = {};
		for (const { key, value } of rows) {
			if (BLOCKED.has(key)) continue;
			if (NUMBER_KEYS.has(key)) flat[key] = Number(value);
			else if (key === 'headless') flat[key] = value === 'true';
			else flat[key] = value;
		}
		return json(flat);
	} catch (e) {
		return handleRouteError(e, 'pusaka/settings GET');
	}
};

export const PUT = async (event: RequestEvent) => {
	try {
		const payload = await readOptionalRequestJson<Record<string, unknown>>(event.request, {});
		const p = proxy(event);
		await Promise.all(
			Object.entries(payload).map(([key, val]) =>
				p.put(`/api/pusaka/settings/${key}`, { value: String(val) })
			)
		);
		return json({ ok: true, ...payload });
	} catch (e) {
		return handleRouteError(e, 'pusaka/settings PUT');
	}
};
