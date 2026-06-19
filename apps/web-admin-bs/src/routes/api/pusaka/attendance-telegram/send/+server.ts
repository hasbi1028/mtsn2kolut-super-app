import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, jsonProxyResponse, proxy, readOptionalRequestJson } from '$lib/server/api';

export const POST = async (event: RequestEvent) => {
	try {
		const payload = await readOptionalRequestJson<Record<string, unknown>>(event.request, {});
		const res = await proxy(event).fetch('/api/pusaka/attendance-telegram/send', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(payload)
		});
		return await jsonProxyResponse<Record<string, unknown>>(res);
	} catch (e) {
		return handleRouteError(e, 'pusaka/attendance-telegram/send POST');
	}
};
