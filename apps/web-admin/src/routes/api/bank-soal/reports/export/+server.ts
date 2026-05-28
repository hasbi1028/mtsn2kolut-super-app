import { bankSoalBackendPathWithQuery } from '$lib/server/bank-soal-backend-paths';
import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, proxy, readRequestJson, streamProxyResponse } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const response = await proxy(event).fetch(bankSoalBackendPathWithQuery('/questions/reports/export', event.url.searchParams));
		return streamProxyResponse(response, { fallbackMessage: 'Export laporan Bank Soal gagal' });
	} catch (e) {
		return handleRouteError(e, 'bank-soal/reports/export GET');
	}
};

export const POST = async (event: RequestEvent) => {
	try {
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const response = await proxy(event).fetch(bankSoalBackendPathWithQuery('/questions/reports/export', event.url.searchParams), {
			method: 'POST',
			body: JSON.stringify(body),
			headers: { 'content-type': 'application/json' }
		});
		return streamProxyResponse(response, { fallbackMessage: 'Export laporan Bank Soal gagal' });
	} catch (e) {
		return handleRouteError(e, 'bank-soal/reports/export POST');
	}
};
