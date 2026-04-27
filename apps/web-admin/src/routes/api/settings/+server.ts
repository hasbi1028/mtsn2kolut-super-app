import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { apiGet, apiPut, handleRouteError } from '$lib/server/api';

interface GoSetting { key: string; value: string }

const BLOCKED = new Set(['admin_password', 'admin_password_hash', 'admin_username']);

export const GET: RequestHandler = async () => {
	try {
		const rows = await apiGet<GoSetting[]>('/api/settings');
		const flat: Record<string, unknown> = {};
		for (const { key, value } of rows) {
			if (BLOCKED.has(key)) continue;
			if (key === 'max_concurrent') flat[key] = Number(value) || 5;
			else if (key === 'headless')  flat[key] = value === 'true';
			else                          flat[key] = value;
		}
		return json(flat);
	} catch (e) {
		return handleRouteError(e, 'settings GET');
	}
};

export const PUT: RequestHandler = async ({ request }) => {
	try {
		const payload = await request.json().catch(() => ({})) as Record<string, unknown>;
		await Promise.all(
			Object.entries(payload).map(([key, val]) =>
				apiPut(`/api/settings/${key}`, { value: String(val) })
			)
		);
		return json({ ok: true, ...payload });
	} catch (e) {
		return handleRouteError(e, 'settings PUT');
	}
};
