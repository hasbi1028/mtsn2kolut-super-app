import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { proxy, ApiError, handleRouteError, readRequestJson } from '$lib/server/api';

const CONTACT_FIELDS = new Set(['phone', 'email', 'address']);

export const GET: RequestHandler = async (event) => {
	if (!event.locals.user) {
		return json({ error: 'Unauthorized' }, { status: 401 });
	}

	try {
		const data = await proxy(event).get('/api/auth/account');
		return json(data);
	} catch (e) {
		if (e instanceof ApiError && e.status === 401) {
			return json({ error: 'Unauthorized' }, { status: 401 });
		}
		return handleRouteError(e, 'auth/account GET');
	}
};

export const PATCH: RequestHandler = async (event) => {
	if (!event.locals.user) {
		return json({ error: 'Unauthorized' }, { status: 401 });
	}

	try {
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		if (typeof body !== 'object' || body === null || Array.isArray(body)) {
			return json({ error: 'Kontak pribadi tidak valid' }, { status: 400 });
		}
		const payload: Record<string, string> = {};
		for (const [key, value] of Object.entries(body)) {
			if (!CONTACT_FIELDS.has(key)) {
				return json({ error: `Field ${key} tidak dapat diubah dari akun saya` }, { status: 400 });
			}
			if (value !== null && typeof value !== 'string') {
				return json({ error: 'Kontak pribadi tidak valid' }, { status: 400 });
			}
			payload[key] = value ?? '';
		}
		const data = await proxy(event).patch('/api/auth/account/contact', payload);
		return json(data);
	} catch (e) {
		if (e instanceof ApiError && e.status === 401) {
			return json({ error: 'Unauthorized' }, { status: 401 });
		}
		return handleRouteError(e, 'auth/account PATCH');
	}
};
