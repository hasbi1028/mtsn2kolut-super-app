import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { proxy, ApiError, handleRouteError, readRequestJson } from '$lib/server/api';

const SELF_CHANGE_REQUEST_FIELDS = new Set(['profile_type', 'field_key', 'requested_value', 'reason']);

function sanitizeCreatePayload(body: Record<string, unknown>) {
	const payload: Record<string, string> = {};
	for (const [key, value] of Object.entries(body)) {
		if (!SELF_CHANGE_REQUEST_FIELDS.has(key)) {
			return { error: `Field ${key} tidak dapat dikirim dari akun saya` };
		}
		if (value !== null && typeof value !== 'string') {
			return { error: 'Permintaan perubahan data resmi tidak valid' };
		}
		payload[key] = value ?? '';
	}
	return { payload };
}

export const GET: RequestHandler = async (event) => {
	if (!event.locals.user) {
		return json({ error: 'Unauthorized' }, { status: 401 });
	}

	try {
		const data = await proxy(event).get('/api/auth/account/change-requests');
		return json(data);
	} catch (e) {
		if (e instanceof ApiError && e.status === 401) {
			return json({ error: 'Unauthorized' }, { status: 401 });
		}
		return handleRouteError(e, 'auth/account/change-requests GET');
	}
};

export const POST: RequestHandler = async (event) => {
	if (!event.locals.user) {
		return json({ error: 'Unauthorized' }, { status: 401 });
	}

	try {
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		if (typeof body !== 'object' || body === null || Array.isArray(body)) {
			return json({ error: 'Permintaan perubahan data resmi tidak valid' }, { status: 400 });
		}
		const sanitized = sanitizeCreatePayload(body);
		if ('error' in sanitized) {
			return json({ error: sanitized.error }, { status: 400 });
		}
		const data = await proxy(event).post('/api/auth/account/change-requests', sanitized.payload);
		return json(data, { status: 201 });
	} catch (e) {
		if (e instanceof ApiError && e.status === 401) {
			return json({ error: 'Unauthorized' }, { status: 401 });
		}
		return handleRouteError(e, 'auth/account/change-requests POST');
	}
};
