import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

interface GoSetting { key: string; value: string }

const BLOCKED = new Set(['admin_password', 'admin_password_hash', 'admin_username']);

export const GET = async (event: RequestEvent) => {
	try {
		const rows = await proxy(event).get<GoSetting[]>('/api/pusaka/settings');
		const flat: Record<string, unknown> = {};
		for (const { key, value } of rows) {
			if (BLOCKED.has(key)) continue;
			if (key === 'max_concurrent') flat[key] = Number(value) || 5;
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
		const payload = await event.request.json().catch(() => ({})) as Record<string, unknown>;
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
