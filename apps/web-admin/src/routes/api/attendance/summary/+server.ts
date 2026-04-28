import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { env } from '$env/dynamic/private';
import { authHeaders, handleRouteError } from '$lib/server/api';

const BASE = (env.API_BASE_URL ?? 'http://localhost:8080').replace(/\/$/, '');

export const GET = async (event: RequestEvent) => {
	try {
		const accessToken = event.cookies.get('access_token');
		const startDate = event.url.searchParams.get('start_date') ?? '';
		const endDate = event.url.searchParams.get('end_date') ?? '';

		const params = new URLSearchParams();
		if (startDate) params.set('start_date', startDate);
		if (endDate) params.set('end_date', endDate);

		const qs = params.toString();
		const res = await fetch(`${BASE}/api/attendance/summary${qs ? `?${qs}` : ''}`, {
			headers: authHeaders(accessToken),
		});
		const data = await res.json().catch(() => ({}));
		return json(data, { status: res.status });
	} catch (e) {
		return handleRouteError(e, 'attendance/summary GET');
	}
};
