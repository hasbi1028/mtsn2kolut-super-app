import {
	apiPath,
	apiPathWithQuery,
	handleRouteError,
	proxy,
	requiredRouteParam,
	streamProxyResponse
} from '$lib/server/api';
import type { RequestEvent } from '@sveltejs/kit';

export const GET = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const upstream = await proxy(event).fetch(
			apiPathWithQuery(apiPath`/api/cbt/assets/${id}/file`, event.url.searchParams)
		);
		return await streamProxyResponse(upstream, {
			fallbackMessage: 'Gagal mengambil aset CBT.',
			defaultCacheControl: 'private, max-age=300'
		});
	} catch (e) {
		return handleRouteError(e, 'cbt/assets/[id]/file GET');
	}
};
