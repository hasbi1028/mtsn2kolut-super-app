import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { env } from '$env/dynamic/private';
import { requireAuthHeaders, handleRouteError } from '$lib/server/api';

const BASE = (env.API_BASE_URL ?? 'http://localhost:8080').replace(/\/$/, '');

export const GET = async (event: RequestEvent) => {
	try {
		const accessToken = event.locals.accessToken ?? event.cookies.get('access_token');
		const qs = event.url.searchParams.toString();
		const res = await fetch(`${BASE}/api/pusaka/attendance${qs ? `?${qs}` : ''}`, {
			headers: requireAuthHeaders(accessToken),
		});
		const data = await res.json().catch(() => ({}));
		return json(data, { status: res.status });
	} catch (e) {
		return handleRouteError(e, 'pusaka/attendance GET');
	}
};
