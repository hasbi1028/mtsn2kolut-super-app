import type { RequestEvent } from '@sveltejs/kit';
import { apiPathWithQuery, handleRouteError, jsonProxyResponse, proxy, readOptionalRequestJson } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const res = await proxy(event).fetch(apiPathWithQuery('/api/pusaka/attendance-telegram/settings', event.url.searchParams));
		return await jsonProxyResponse<Record<string, unknown>>(res);
	} catch (e) {
		return handleRouteError(e, 'pusaka/attendance-telegram/settings GET');
	}
};

export const PUT = async (event: RequestEvent) => {
	try {
		const payload = await readOptionalRequestJson<Record<string, unknown>>(event.request, {});
		const res = await proxy(event).fetch('/api/pusaka/attendance-telegram/settings', {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(payload)
		});
		return await jsonProxyResponse<Record<string, unknown>>(res);
	} catch (e) {
		return handleRouteError(e, 'pusaka/attendance-telegram/settings PUT');
	}
};
