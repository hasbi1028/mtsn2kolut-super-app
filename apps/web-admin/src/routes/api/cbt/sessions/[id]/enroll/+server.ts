import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { handleRouteError } from '$lib/server/api';
import { env } from '$env/dynamic/private';

const BASE = (env.API_BASE_URL ?? 'http://localhost:8080').replace(/\/$/, '');
const INTERNAL_KEY = env.INTERNAL_API_KEY ?? '';

export const POST: RequestHandler = async ({ params, request }) => {
	try {
		const body = await request.json();
		const res = await fetch(`${BASE}/api/cbt/sessions/${params.id}/enroll`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json', 'X-Internal-Key': INTERNAL_KEY },
			body: JSON.stringify(body),
		});
		const data = await res.json();
		return json(data, { status: res.status });
	} catch (e) {
		return handleRouteError(e, 'cbt/sessions enroll POST');
	}
};
