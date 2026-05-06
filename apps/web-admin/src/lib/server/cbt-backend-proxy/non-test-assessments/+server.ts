import { cbtApiPath, cbtBackendPath } from '$lib/server/cbt-backend-paths';
import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import {
	apiPathWithQuery,
	handleRouteError,
	proxy,
	readRequestJson
} from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const data = await proxy(event).get(apiPathWithQuery(cbtBackendPath('/non-test-assessments'), event.url.searchParams));
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/non-test-assessments GET');
	}
};

export const POST = async (event: RequestEvent) => {
	try {
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const data = await proxy(event).post(cbtBackendPath('/non-test-assessments'), body);
		return json(data, { status: 201 });
	} catch (e) {
		return handleRouteError(e, 'cbt/non-test-assessments POST');
	}
};
