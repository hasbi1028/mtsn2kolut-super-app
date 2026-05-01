import type { RequestEvent } from '@sveltejs/kit';
import { env } from '$env/dynamic/private';
import { handleRouteError, jsonProxyResponse, readRequestJson } from '$lib/server/api';

const BASE = (env.API_BASE_URL ?? 'http://localhost:8080').replace(/\/$/, '');

export const POST = async (event: RequestEvent) => {
	try {
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const res = await event.fetch(`${BASE}/api/public/register-student`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(body),
		});
		return await jsonProxyResponse<Record<string, unknown>>(res, { status: 201 });
	} catch (e) {
		return handleRouteError(e, 'public/register-student POST');
	}
};
