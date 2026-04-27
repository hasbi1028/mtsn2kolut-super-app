import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { apiPost } from '$lib/server/api';

export const POST: RequestHandler = async () => {
	await apiPost('/api/jobs/cancel-all');
	return json({ ok: true });
};
