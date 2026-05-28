import { bankSoalBackendPathWithQuery } from '$lib/server/bank-soal-backend-paths';
import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, proxy } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const data = await proxy(event).get(bankSoalBackendPathWithQuery('/questions/reports', event.url.searchParams));
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'bank-soal/reports GET');
	}
};
