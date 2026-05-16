import { cbtBackendPath, cbtBackendPathWithQuery, cbtApiPath } from '$lib/server/cbt-backend-paths';
import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, proxy, readRequestJson } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const data = await proxy(event).get(cbtBackendPathWithQuery('/approvals', event.url.searchParams));
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/approvals GET');
	}
};

export const POST = async (event: RequestEvent) => {
	try {
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const data = await proxy(event).post(cbtBackendPath('/approvals'), body);
		return json(data, { status: 201 });
	} catch (e) {
		return handleRouteError(e, 'cbt/approvals POST');
	}
};

export const DELETE = async (event: RequestEvent) => {
	try {
		const id = event.url.searchParams.get('id');
		if (!id) return json({ error: 'id pengesahan wajib diisi' }, { status: 400 });
		const body = await readRequestJson<Record<string, unknown>>(event.request).catch(() => ({}));
		const data = await proxy(event).post(cbtApiPath`/approvals/${id}/revoke`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/approvals DELETE');
	}
};
