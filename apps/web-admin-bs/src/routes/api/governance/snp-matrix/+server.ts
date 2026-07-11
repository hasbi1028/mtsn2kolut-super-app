import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

export const GET: RequestHandler = async (event) => {
	try {
		const data = await proxy(event).get('/api/governance/snp-matrix');
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'governance/snp-matrix GET');
	}
};
