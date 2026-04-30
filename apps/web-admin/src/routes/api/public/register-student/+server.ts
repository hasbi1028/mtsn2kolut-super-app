import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { env } from '$env/dynamic/private';
import { handleRouteError } from '$lib/server/api';

const BASE = (env.API_BASE_URL ?? 'http://localhost:8080').replace(/\/$/, '');

export const POST = async (event: RequestEvent) => {
	try {
		const body = await event.request.json() as Record<string, unknown>;
		const res = await fetch(`${BASE}/api/public/register-student`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(body),
		});
		const data = await res.json().catch(() => ({}));
		if (!res.ok) return json({ error: data.error ?? `HTTP ${res.status}` }, { status: res.status });
		return json(data, { status: 201 });
	} catch (e) {
		return handleRouteError(e, 'public/register-student POST');
	}
};
