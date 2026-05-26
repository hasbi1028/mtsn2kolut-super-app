import type { RequestHandler } from './$types';
import { json } from '@sveltejs/kit';
import { listCbtMobileArtifacts } from '$lib/server/cbt-mobile-artifacts';

export const GET: RequestHandler = async (event) => {
	return json(await listCbtMobileArtifacts(event), {
		headers: {
			'Cache-Control': 'no-store, max-age=0'
		}
	});
};
