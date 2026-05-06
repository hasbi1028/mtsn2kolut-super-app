import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { proxy, apiPath, handleRouteError, readRequestJson, requiredRouteParam } from '$lib/server/api';

const REVIEW_FIELDS = new Set(['status', 'review_note']);

export const PATCH: RequestHandler = async (event) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		if (typeof body !== 'object' || body === null || Array.isArray(body)) {
			return json({ error: 'Review permintaan perubahan data tidak valid' }, { status: 400 });
		}
		const payload: Record<string, string> = {};
		for (const [key, value] of Object.entries(body)) {
			if (!REVIEW_FIELDS.has(key)) {
				return json({ error: `Field ${key} tidak dapat dikirim untuk review` }, { status: 400 });
			}
			if (value !== null && typeof value !== 'string') {
				return json({ error: 'Review permintaan perubahan data tidak valid' }, { status: 400 });
			}
			payload[key] = value ?? '';
		}
		if (payload.status !== 'approved' && payload.status !== 'rejected') {
			return json({ error: 'Status review harus approved atau rejected' }, { status: 400 });
		}
		const data = await proxy(event).patch(apiPath`/api/users/change-requests/${id}`, payload);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'users/change-requests/:id PATCH');
	}
};
